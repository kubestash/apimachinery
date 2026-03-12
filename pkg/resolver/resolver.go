/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resolver

import (
	"context"
	"fmt"
	"strings"

	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"
	"kubestash.dev/apimachinery/pkg/blob"
	"kubestash.dev/apimachinery/pkg/retry"

	aws2 "github.com/aws/aws-sdk-go-v2/aws"
	"gomodules.xyz/restic"
	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewBackupStorageResolver creates a StorageConfigResolver that resolves storage configuration
// from a BackupStorage custom resource. This is the default resolver for the kubestash project.
func NewBackupStorageResolver(kbClient client.Client, bs *storageapi.BackupStorage) restic.StorageConfigResolver {
	return func(backend *restic.Backend) error {
		return ResolveResticBackend(kbClient, bs, backend)
	}
}

func ResolveResticBackend(kbClient client.Client, bs *storageapi.BackupStorage, backend *restic.Backend) error {
	var storageSecretName string
	switch {
	case bs.Spec.Storage.S3 != nil:
		s3 := bs.Spec.Storage.S3
		storageSecretName = s3.SecretName

		backend.StorageConfig = &restic.StorageConfig{
			Provider:       string(storageapi.ProviderS3),
			Bucket:         s3.Bucket,
			Endpoint:       s3.Endpoint,
			Region:         s3.Region,
			Prefix:         s3.Prefix,
			InsecureTLS:    s3.InsecureTLS,
			MaxConnections: s3.MaxConnections,
		}

	case bs.Spec.Storage.GCS != nil:
		gcs := bs.Spec.Storage.GCS
		storageSecretName = gcs.SecretName

		backend.StorageConfig = &restic.StorageConfig{
			Provider:       string(storageapi.ProviderGCS),
			Bucket:         gcs.Bucket,
			Prefix:         gcs.Prefix,
			MaxConnections: gcs.MaxConnections,
		}

	case bs.Spec.Storage.Azure != nil:
		azure := bs.Spec.Storage.Azure
		storageSecretName = azure.SecretName

		backend.StorageConfig = &restic.StorageConfig{
			Provider:            string(storageapi.ProviderAzure),
			Bucket:              azure.Container,
			Prefix:              azure.Prefix,
			AzureStorageAccount: azure.StorageAccount,
			MaxConnections:      azure.MaxConnections,
		}

	case bs.Spec.Storage.Local != nil:
		local := bs.Spec.Storage.Local

		backend.StorageConfig = &restic.StorageConfig{
			Provider:       string(storageapi.ProviderLocal),
			Bucket:         local.MountPath,
			Prefix:         local.SubPath,
			MaxConnections: local.MaxConnections,
		}

		if backend.MountPath != "" {
			backend.Bucket = backend.MountPath
		}

	default:
		return fmt.Errorf("no storage backend configured in BackupStorage %s/%s", bs.Namespace, bs.Name)
	}

	if storageSecretName != "" {
		secret := &core.Secret{}
		if err := kbClient.Get(context.Background(),
			client.ObjectKey{Name: storageSecretName, Namespace: bs.Namespace},
			secret); err != nil {
			return fmt.Errorf("failed to get storage Secret %s/%s: %w", bs.Namespace, storageSecretName, err)
		}

		backend.StorageSecret = secret
	} else if backend.Provider == string(storageapi.ProviderS3) {
		creds, err := getS3Credentials(context.Background(), kbClient, bs)
		if err != nil {
			return fmt.Errorf("failed to get S3 credentials for BackupStorage %s/%s: %w", bs.Namespace, bs.Name, err)
		}
		backend.Envs = convertS3CredentialsToEnvMap(creds)
	}

	return nil
}

func getS3Credentials(ctx context.Context, kbClient client.Client, bs *storageapi.BackupStorage) (*aws2.Credentials, error) {
	b, err := blob.NewBlob(ctx, kbClient, bs)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob client for BackupStorage %s/%s: %w", bs.Namespace, bs.Name, err)
	}
	customRetry := func(err error, output string) bool {
		if err == nil {
			return false
		}
		return strings.Contains(err.Error(), "failed to refresh cached credentials") ||
			strings.Contains(err.Error(), "failed to retrieve credentials") ||
			strings.Contains(err.Error(), "AssumeRoleWithWebIdentity")
	}

	rc := retry.NewRetryConfig(func(config *retry.RetryConfig) {
		config.ShouldRetry = customRetry
	})
	cred, err := rc.RunWithRetry(ctx, func() (any, error) {
		return b.GetS3Credentials(ctx, false)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 credentials for BackupStorage %s/%s: %w", bs.Namespace, bs.Name, err)
	}

	klog.Infof("Successfully got S3 credentials for BackupStorage %s/%s", bs.Namespace, bs.Name)
	return cred.(*aws2.Credentials), nil
}

func convertS3CredentialsToEnvMap(creds *aws2.Credentials) map[string]string {
	values := make(map[string]string)
	values[blob.AWSAccessKeyId] = creds.AccessKeyID
	values[blob.AWSSecretAccessKey] = creds.SecretAccessKey
	values[blob.AWSSessionToken] = creds.SessionToken
	return values
}
