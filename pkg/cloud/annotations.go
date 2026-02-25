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

	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	core "k8s.io/api/core/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AWSIRSARoleAnnotationKey         = "eks.amazonaws.com/role-arn"
	GCPWorkloadIdentityAnnotationKey = "go.klusters.dev/iam-gke-io-workloadIdentity"
	BucketAnnotationKey              = "go.klusters.dev/bucket-names"
	GCPClusterNameAnnotationKey      = "go.klusters.dev/iam-gke-cluster-name"
	GCPClusterRegionAnnotationKey    = "go.klusters.dev/iam-gke-cluster-region"
	GCPProjectIDAnnotationKey        = "go.klusters.dev/iam-gke-project-id"
	GCPProjectNumberAnnotationKey    = "go.klusters.dev/iam-gke-project-number"
	GCPRolesAnnotationKey            = "go.klusters.dev/iam-gke-roles"
)

func GetCloudAnnotations(ctx context.Context, kc client.Client, storages ...storageapi.BackupStorage) (map[string]string, error) {
	annotations, err := GetCloudAnnotationsFromServiceAccount(ctx, kc)
	if err != nil {
		return nil, err
	}
	if storages != nil {
		setBucketAnnotations(annotations, storages...)
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
	if val, ok := sa.Annotations[GCPClusterNameAnnotationKey]; ok {
		annotations[GCPClusterNameAnnotationKey] = val
	}
	if val, ok := sa.Annotations[GCPClusterRegionAnnotationKey]; ok {
		annotations[GCPClusterRegionAnnotationKey] = val
	}
	if val, ok := sa.Annotations[GCPProjectIDAnnotationKey]; ok {
		annotations[GCPProjectIDAnnotationKey] = val
	}
	if val, ok := sa.Annotations[GCPProjectNumberAnnotationKey]; ok {
		annotations[GCPProjectNumberAnnotationKey] = val
	}
	if val, ok := sa.Annotations[GCPRolesAnnotationKey]; ok {
		annotations[GCPRolesAnnotationKey] = val
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

func setBucketAnnotations(annotations map[string]string, storages ...storageapi.BackupStorage) {
	if !meta.HasKey(annotations, AWSIRSARoleAnnotationKey) || meta.HasKey(annotations, GCPWorkloadIdentityAnnotationKey) {
		return
	}

	bucketNames := ""
	for _, backupStorage := range storages {
		switch backupStorage.Spec.Storage.Provider {
		case storageapi.ProviderS3:
			bucketNames = fmt.Sprintf("%s,%s", bucketNames, backupStorage.Spec.Storage.S3.Bucket)
		case storageapi.ProviderGCS:
			bucketNames = fmt.Sprintf("%s,%s", bucketNames, backupStorage.Spec.Storage.GCS.Bucket)
		}
	}
	if bucketNames != "" {
		annotations[BucketAnnotationKey] = bucketNames[1:]
	}
}
