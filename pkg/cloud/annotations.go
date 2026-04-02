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
	AWSSeedRoleAnnotationName        = "go.klusters.dev/seed-role-name"
	AWSIRSARoleAnnotation            = "eks.amazonaws.com/role-arn"
	GCPWorkloadIdentityAnnotationKey = "go.klusters.dev/iam-gke-io-workloadIdentity"
	BucketAnnotationKey              = "go.klusters.dev/bucket-names"
	GCPClusterNameAnnotationKey      = "go.klusters.dev/iam-gke-cluster-name"
	GCPClusterRegionAnnotationKey    = "go.klusters.dev/iam-gke-cluster-region"
	GCPProjectIDAnnotationKey        = "go.klusters.dev/iam-gke-project-id"
	GCPProjectNumberAnnotationKey    = "go.klusters.dev/iam-gke-project-number"
	GCPRolesAnnotationKey            = "go.klusters.dev/iam-gke-roles"

	AzureMINameAnnotation         = "klusters.dev/azure-mi-name"
	AzureResourceGroupAnnotation  = "klusters.dev/azure-rg-name"
	AzureSubscriptionIDAnnotation = "klusters.dev/azure-subscription-id"
	AzureMIClientIDAnnotation     = "azure.workload.identity/client-id"
	AzureMITenantIDAnnotation     = "azure.workload.identity/tenant-id"
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
	sa, err := getServiceAccount(ctx, kc, &kmapi.ObjectReference{
		Name:      meta.PodServiceAccount(),
		Namespace: meta.PodNamespace(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve service account: %w", err)
	}
	annotations := map[string]string{}
	if val, ok := sa.Annotations[AWSIRSARoleAnnotation]; ok {
		annotations[AWSSeedRoleAnnotationName] = val
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
	if val, ok := sa.Annotations[AzureMINameAnnotation]; ok {
		annotations[AzureMINameAnnotation] = val
	}
	if val, ok := sa.Annotations[AzureResourceGroupAnnotation]; ok {
		annotations[AzureResourceGroupAnnotation] = val
	}
	if val, ok := sa.Annotations[AzureSubscriptionIDAnnotation]; ok {
		annotations[AzureSubscriptionIDAnnotation] = val
	}
	return annotations, nil
}

func getServiceAccount(ctx context.Context, c client.Client, ref *kmapi.ObjectReference) (*core.ServiceAccount, error) {
	sa := &core.ServiceAccount{}
	if err := c.Get(ctx, ref.ObjectKey(), sa); err != nil {
		return nil, err
	}
	return sa, nil
}

func setBucketAnnotations(annotations map[string]string, storages ...storageapi.BackupStorage) {
	if !meta.HasKey(annotations, AWSSeedRoleAnnotationName) && !meta.HasKey(annotations, GCPWorkloadIdentityAnnotationKey) &&
		!meta.HasKey(annotations, AzureSubscriptionIDAnnotation) {
		return
	}

	bucketNames := ""
	for _, backupStorage := range storages {
		switch backupStorage.Spec.Storage.Provider {
		case storageapi.ProviderS3:
			bucketNames = fmt.Sprintf("%s,%s", bucketNames, backupStorage.Spec.Storage.S3.Bucket)
		case storageapi.ProviderGCS:
			bucketNames = fmt.Sprintf("%s,%s", bucketNames, backupStorage.Spec.Storage.GCS.Bucket)
		case storageapi.ProviderAzure:
			bucketNames = fmt.Sprintf("%s#%s,%s", bucketNames,
				backupStorage.Spec.Storage.Azure.Container, backupStorage.Spec.Storage.Azure.StorageAccount) // format "container1,storageaccount1#container2,storageaccount2#..."
		}
	}
	if bucketNames != "" {
		annotations[BucketAnnotationKey] = bucketNames[1:]
	}
}

func AddCloudAnnotationsToSAIfNeeded(ctx context.Context, kbClient client.Client, bs *storageapi.BackupStorage, saRef *kmapi.ObjectReference, bcRef *kmapi.ObjectReference) error {
	sa, err := getServiceAccount(ctx, kbClient, saRef)
	if err != nil {
		return fmt.Errorf("failed to get service account: %v", err)
	}
	if !isCloudAnnotationNeeded(bs, sa) { // Return if not needed
		return nil
	}
	if err := addAnnotationsToServiceAccount(kbClient, bs, sa, bcRef); err != nil {
		return fmt.Errorf("failed to add cloud annotations to database service account: %w", err)
	}
	if !hasCredLessManagerProvidedAnnotation(bs, sa) {
		return fmt.Errorf("credential-less cloud annotation for the DB service account does not exist yet")
	}
	return nil
}

func hasCredLessManagerProvidedAnnotation(bs *storageapi.BackupStorage, sa *core.ServiceAccount) bool {
	switch bs.Spec.Storage.Provider {
	case storageapi.ProviderS3:
		_, exists := sa.Annotations[AWSIRSARoleAnnotation]
		return exists
	default:
		return false
	}
}

func isCloudAnnotationNeeded(bs *storageapi.BackupStorage, sa *core.ServiceAccount) bool {
	if bs.IsCredentialLessModeEnabled() {
		switch bs.Spec.Storage.Provider {
		case storageapi.ProviderS3:
			_, ok := sa.Annotations[AWSSeedRoleAnnotationName]
			return !ok
			// case storageapi.ProviderAzure:
		}
	}
	return false
}

func addAnnotationsToServiceAccount(kbClient client.Client, bs *storageapi.BackupStorage, sa *core.ServiceAccount, bcRef *kmapi.ObjectReference) error {
	if hasRequiredCloudAnnotations(bs, sa) {
		return nil
	}

	annotations, err := getLatestSuccessfulBackupAnnotations(context.Background(), kbClient, bcRef)
	if err != nil {
		return fmt.Errorf("failed to fetch annotations from latest successful backup: %w", err)
	}
	if err := annotateRequiredAnnotations(bs, sa, annotations); err != nil {
		return fmt.Errorf("failed to annotate required cloud annotations: %w", err)
	}
	if err := kbClient.Update(context.Background(), sa); err != nil {
		return fmt.Errorf("failed to update service account %s/%s with cloud annotations: %w",
			sa.Namespace, sa.Name, err)
	}
	return nil
}

func hasRequiredCloudAnnotations(bs *storageapi.BackupStorage, sa *core.ServiceAccount) bool {
	if bs.Spec.Storage.Provider == storageapi.ProviderS3 {
		return sa.Annotations[AWSSeedRoleAnnotationName] != "" &&
			sa.Annotations[BucketAnnotationKey] != ""
	}
	return false
}

func getLatestSuccessfulBackupAnnotations(ctx context.Context, kbClient client.Client, bcRef *kmapi.ObjectReference) (map[string]string, error) {
	session, err := findLatestSuccessfulBackupSession(ctx, kbClient, bcRef)
	if err != nil {
		return nil, err
	}
	return getBackupJobAnnotations(ctx, kbClient, session)
}

func findLatestSuccessfulBackupSession(ctx context.Context, kbClient client.Client, bcRef *kmapi.ObjectReference) (*v1alpha1.BackupSession, error) {
	list := &v1alpha1.BackupSessionList{}
	opts := []client.ListOption{
		client.MatchingLabels{apis.KubeStashInvokerName: bcRef.Name},
		client.InNamespace(bcRef.Namespace),
	}
	if err := kbClient.List(ctx, list, opts...); err != nil {
		return nil, fmt.Errorf("failed to list backup sessions for %s/%s: %w",
			bcRef.Namespace, bcRef.Name, err)
	}

	// Sort by creation timestamp descending (newest first)
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].CreationTimestamp.After(list.Items[j].CreationTimestamp.Time)
	})

	for i := range list.Items {
		session := &list.Items[i]
		if session.Status.Phase == v1alpha1.BackupSessionSucceeded {
			return session, nil
		}
	}
	return nil, fmt.Errorf("no successful backup session found for configuration %s/%s",
		bcRef.Namespace, bcRef.Name)
}

func getBackupJobAnnotations(ctx context.Context, kbClient client.Client, session *v1alpha1.BackupSession) (map[string]string, error) {
	list := &batchv1.JobList{}
	opts := []client.ListOption{
		client.MatchingLabels{
			apis.KubeStashSessionName:      sessionFullBackup,
			meta.ComponentLabelKey:         apis.KubeStashBackupComponent,
			apis.KubeStashInvokerName:      session.Name,
			apis.KubeStashInvokerNamespace: session.Namespace,
		},
		client.InNamespace(session.Namespace),
	}
	if err := kbClient.List(ctx, list, opts...); err != nil {
		return nil, fmt.Errorf("failed to list backup jobs for session %s/%s: %w",
			session.Namespace, session.Name, err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no backup jobs found for session %s/%s",
			session.Namespace, session.Name)
	}

	return list.Items[0].Annotations, nil
}

func applyAWSAnnotations(sa *core.ServiceAccount, source map[string]string) error {
	required := map[string]string{
		AWSSeedRoleAnnotationName: source[AWSSeedRoleAnnotationName],
		BucketAnnotationKey:       source[BucketAnnotationKey],
	}

	for key, val := range required {
		if val == "" {
			return fmt.Errorf("required annotation %q missing from backup job annotations", key)
		}
		if sa.Annotations == nil {
			sa.Annotations = make(map[string]string)
		}
		sa.Annotations[key] = val
	}
	return nil
}

func annotateRequiredAnnotations(bs *storageapi.BackupStorage, sa *core.ServiceAccount, annotations map[string]string) error {
	switch bs.Spec.Storage.Provider {
	case storageapi.ProviderS3:
		return applyAWSAnnotations(sa, annotations)
	// case storageapi.ProviderGCS:
	// 	return applyGCPAnnotations(sa, annotations)
	default:
		return fmt.Errorf("unsupported storage provider: %s", bs.Spec.Storage.Provider)

	}
}
