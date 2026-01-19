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

package restic

import (
	"context"
	"fmt"

	storage "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewBackupStorageResolver creates a StorageConfigResolver that resolves storage configuration
// from a BackupStorage custom resource. This is the default resolver for the kubestash project.
func NewBackupStorageResolver(kbClient client.Client, bsRef metav1.ObjectMeta) StorageConfigResolver {
	return func(backend *Backend) error {
		bs := &storage.BackupStorage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bsRef.Name,
				Namespace: bsRef.Namespace,
			},
		}

		if err := kbClient.Get(context.Background(), client.ObjectKeyFromObject(bs), bs); err != nil {
			return fmt.Errorf("failed to get BackupStorage %s/%s: %w", bsRef.Namespace, bsRef.Name, err)
		}
		var storageSecretName string
		switch {
		case bs.Spec.Storage.S3 != nil:
			s3 := bs.Spec.Storage.S3
			storageSecretName = s3.SecretName
			backend.StorageConfig = &StorageConfig{
				Provider:       string(storage.ProviderS3),
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
			backend.StorageConfig = &StorageConfig{
				Provider:       string(storage.ProviderGCS),
				Bucket:         gcs.Bucket,
				Prefix:         gcs.Prefix,
				MaxConnections: gcs.MaxConnections,
			}
		case bs.Spec.Storage.Azure != nil:
			azure := bs.Spec.Storage.Azure
			storageSecretName = azure.SecretName
			backend.StorageConfig = &StorageConfig{
				Provider:       string(storage.ProviderAzure),
				Bucket:         azure.Container,
				Prefix:         azure.Prefix,
				StorageAccount: azure.StorageAccount,
				MaxConnections: azure.MaxConnections,
			}
		case bs.Spec.Storage.Local != nil:
			local := bs.Spec.Storage.Local
			backend.StorageConfig = &StorageConfig{
				Provider:       string(storage.ProviderLocal),
				Bucket:         local.MountPath,
				Prefix:         local.SubPath,
				MaxConnections: local.MaxConnections,
			}
			if backend.MountPath != "" {
				backend.Bucket = backend.MountPath
			}
		default:
			return fmt.Errorf("no storage backend configured in BackupStorage %s/%s", bsRef.Namespace, bsRef.Name)
		}

		if storageSecretName != "" {
			secret := &core.Secret{}
			if err := kbClient.Get(context.Background(), client.ObjectKey{Name: storageSecretName, Namespace: bsRef.Namespace}, secret); err != nil {
				return fmt.Errorf("failed to get storage Secret %s/%s: %w", bsRef.Namespace, storageSecretName, err)
			}
			backend.StorageSecret = secret
		}
		return nil
	}
}
