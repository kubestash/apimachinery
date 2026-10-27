/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package blob

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync/atomic"
	"time"

	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"
	"kubestash.dev/apimachinery/pkg/workerpool"

	aws2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// ErrObjectLockNotSupported is returned when the provider of the BackupStorage does not support Object Lock.
var ErrObjectLockNotSupported = errors.New("object lock is supported only for the s3 provider")

// lockDir is the directory where restic keeps its lock files. The lock files are deleted
// by restic as soon as an operation completes, so there is no point to keep them locked.
const lockDir = "locks"

type ObjectLockConfig struct {
	Enabled          bool
	Mode             storageapi.ObjectLockMode
	DefaultRetention time.Duration
}

type ExtendRetentionStats struct {
	// Scanned is the number of objects checked
	Scanned int64
	// Extended is the number of objects whose retention has been extended
	Extended int64
}

func (b *Blob) s3Client(ctx context.Context) (*s3.Client, error) {
	if b.backupStorage.Spec.Storage.Provider != storageapi.ProviderS3 {
		return nil, ErrObjectLockNotSupported
	}
	cfg, err := b.getS3Config(ctx, false)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = true
	}), nil
}

// GetObjectLockConfig returns the Object Lock configuration of the bucket.
// It returns ErrObjectLockNotSupported if the provider does not support Object Lock.
func (b *Blob) GetObjectLockConfig(ctx context.Context) (*ObjectLockConfig, error) {
	c, err := b.s3Client(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{
		Bucket: aws2.String(b.backupStorage.Spec.Storage.S3.Bucket),
	})
	if err != nil {
		var apiErr smithy.APIError
		// Some S3 compatible storages do not implement Object Lock at all.
		if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "ObjectLockConfigurationNotFoundError" || apiErr.ErrorCode() == "NotImplemented") {
			return &ObjectLockConfig{}, nil
		}
		return nil, err
	}

	cfg := &ObjectLockConfig{}
	if out.ObjectLockConfiguration == nil || out.ObjectLockConfiguration.ObjectLockEnabled != types.ObjectLockEnabledEnabled {
		return cfg, nil
	}
	cfg.Enabled = true
	if rule := out.ObjectLockConfiguration.Rule; rule != nil && rule.DefaultRetention != nil {
		cfg.Mode = toObjectLockMode(rule.DefaultRetention.Mode)
		if rule.DefaultRetention.Days != nil {
			cfg.DefaultRetention = time.Duration(*rule.DefaultRetention.Days) * 24 * time.Hour
		} else if rule.DefaultRetention.Years != nil {
			cfg.DefaultRetention = time.Duration(*rule.DefaultRetention.Years) * 365 * 24 * time.Hour
		}
	}
	return cfg, nil
}

// ExtendObjectRetention extends the retain-until date of every current object inside dir to now + period.
// Objects that are already locked for more than half of the period are skipped, so an object is
// updated about once in every half period. Deleted objects (i.e. objects hidden by a delete marker)
// are not listed, so their retention is never extended and they become deletable after it expires.
// The retention is never shortened.
func (b *Blob) ExtendObjectRetention(ctx context.Context, dir string, mode storageapi.ObjectLockMode, period time.Duration) (*ExtendRetentionStats, error) {
	if period <= 0 {
		return nil, fmt.Errorf("invalid retention period %v", period)
	}
	c, err := b.s3Client(ctx)
	if err != nil {
		return nil, err
	}

	bucket := b.backupStorage.Spec.Storage.S3.Bucket
	prefix := strings.Trim(path.Join(b.prefix, dir), "/") + "/"
	if prefix == "/" {
		prefix = ""
	}
	retainUntil := time.Now().Add(period).UTC()
	threshold := time.Now().Add(period / 2)

	stats := &ExtendRetentionStats{}
	wp := workerpool.NewWorkerPool(ctx, max(b.maxConnections, 10))
	pager := s3.NewListObjectsV2Paginator(c, &s3.ListObjectsV2Input{
		Bucket: aws2.String(bucket),
		Prefix: aws2.String(prefix),
	})
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			_ = wp.Wait()
			return stats, fmt.Errorf("failed to list objects under %q: %w", prefix, err)
		}
		for _, obj := range page.Contents {
			key := aws2.ToString(obj.Key)
			if strings.HasSuffix(key, "/") || isLockFile(strings.TrimPrefix(key, prefix)) {
				continue
			}
			wp.Run(func() error {
				atomic.AddInt64(&stats.Scanned, 1)
				extended, err := extendRetention(ctx, c, bucket, key, mode, retainUntil, threshold)
				if err != nil {
					return fmt.Errorf("failed to extend retention of %q: %w", key, err)
				}
				if extended {
					atomic.AddInt64(&stats.Extended, 1)
				}
				return nil
			})
		}
	}
	return stats, wp.Wait()
}

func extendRetention(ctx context.Context, c *s3.Client, bucket, key string, mode storageapi.ObjectLockMode, retainUntil, threshold time.Time) (bool, error) {
	cur, err := c.GetObjectRetention(ctx, &s3.GetObjectRetentionInput{
		Bucket: aws2.String(bucket),
		Key:    aws2.String(key),
	})
	if err != nil && !isNoRetentionError(err) {
		return false, err
	}

	retentionMode := toS3RetentionMode(mode)
	if err == nil && cur.Retention != nil {
		if cur.Retention.RetainUntilDate != nil && !cur.Retention.RetainUntilDate.Before(threshold) {
			return false, nil
		}
		// Keep the mode of the object. Changing it from COMPLIANCE to GOVERNANCE is not allowed.
		if cur.Retention.Mode != "" {
			retentionMode = types.ObjectLockRetentionMode(cur.Retention.Mode)
		}
	}

	_, err = c.PutObjectRetention(ctx, &s3.PutObjectRetentionInput{
		Bucket: aws2.String(bucket),
		Key:    aws2.String(key),
		Retention: &types.ObjectLockRetention{
			Mode:            retentionMode,
			RetainUntilDate: aws2.Time(retainUntil),
		},
	})
	return err == nil, err
}

func isLockFile(relKey string) bool {
	return relKey == lockDir || strings.HasPrefix(relKey, lockDir+"/") || strings.Contains(relKey, "/"+lockDir+"/")
}

func isNoRetentionError(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchObjectLockConfiguration"
}

func toObjectLockMode(mode types.ObjectLockRetentionMode) storageapi.ObjectLockMode {
	switch mode {
	case types.ObjectLockRetentionModeCompliance:
		return storageapi.ObjectLockModeCompliance
	case types.ObjectLockRetentionModeGovernance:
		return storageapi.ObjectLockModeGovernance
	default:
		return ""
	}
}

// toS3RetentionMode uses GOVERNANCE unless COMPLIANCE is asked explicitly, as COMPLIANCE retention can never be removed.
func toS3RetentionMode(mode storageapi.ObjectLockMode) types.ObjectLockRetentionMode {
	if mode == storageapi.ObjectLockModeCompliance {
		return types.ObjectLockRetentionModeCompliance
	}
	return types.ObjectLockRetentionModeGovernance
}
