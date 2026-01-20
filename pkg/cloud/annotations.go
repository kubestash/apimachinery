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

package cloud

import (
	"context"
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/meta"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AWSIRSARoleAnnotationKey         = "eks.amazonaws.com/role-arn"
	GCPWorkloadIdentityAnnotationKey = "go.klusters.dev/iam-gke-io-workloadIdentity"
	S3orGCSBucketAnnotationKey       = "go.klusters.dev/bucket-names"
)

func GetCloudAnnotations(ctx context.Context, kc client.Client, storages ...storageapi.BackupStorage) (map[string]string, error) {
	annotations, err := GetCloudAnnotationsFromServiceAccount(ctx, kc)
	if err != nil {
		return nil, err
	}
	if storages != nil && (meta.HasKey(annotations, AWSIRSARoleAnnotationKey) || meta.HasKey(annotations, GCPWorkloadIdentityAnnotationKey)) {
		annotations[S3orGCSBucketAnnotationKey] = getBucketAnnotationValueFromS3orGCSBackupStorage(storages...)
	}
	return annotations, nil
}

func GetCloudAnnotationsFromServiceAccount(ctx context.Context, kc client.Client) (map[string]string, error) {
	sa, err := getServiceAccount(ctx, kc, kmapi.ObjectReference{
		Name:      meta.PodServiceAccount(),
		Namespace: meta.PodNamespace(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve service account: %w", err)
	}
	annotations := map[string]string{}
	if val, ok := sa.Annotations[AWSIRSARoleAnnotationKey]; ok {
		annotations[AWSIRSARoleAnnotationKey] = val
	}
	if val, ok := sa.Annotations[GCPWorkloadIdentityAnnotationKey]; ok {
		annotations[GCPWorkloadIdentityAnnotationKey] = val
	}
	return annotations, nil
}

func getServiceAccount(ctx context.Context, c client.Client, ref kmapi.ObjectReference) (*core.ServiceAccount, error) {
	sa := &core.ServiceAccount{}
	if err := c.Get(ctx, ref.ObjectKey(), sa); err != nil {
		return nil, err
	}
	return sa, nil
}

func getBucketAnnotationValueFromS3orGCSBackupStorage(storages ...storageapi.BackupStorage) string {
	var bucketNames []string
	for _, backupStorage := range storages {
		switch backupStorage.Spec.Storage.Provider {
		case storageapi.ProviderS3:
			bucketNames = append(bucketNames, backupStorage.Spec.Storage.S3.Bucket)
		case storageapi.ProviderGCS:
			bucketNames = append(bucketNames, backupStorage.Spec.Storage.GCS.Bucket)
		}
	}
	bucketAnnotationValue := ""
	if bucketNames != nil {
		for _, bucketName := range bucketNames {
			bucketAnnotationValue += bucketName + ","
		}
		bucketAnnotationValue = strings.TrimSuffix(bucketAnnotationValue, ",")
	}
	return bucketAnnotationValue
}
