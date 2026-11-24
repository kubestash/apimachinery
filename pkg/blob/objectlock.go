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
	"time"

	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	aws2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// ErrObjectLockNotSupported is returned when the provider of the BackupStorage does not support Object Lock.
var ErrObjectLockNotSupported = errors.New("object lock is supported only for the s3 provider")

type ObjectLockConfig struct {
	Enabled          bool
	Mode             storageapi.ObjectLockMode
	DefaultRetention time.Duration
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
