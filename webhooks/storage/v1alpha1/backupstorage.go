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

package v1alpha1

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/apis/storage/v1alpha1"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
)

// log is for logging in this package.
var backupstoragelog = logf.Log.WithName("backupstorage-resource")

type BackupStorageCustomWebhook struct{}

type BackupStorage struct {
	*v1alpha1.BackupStorage
}

// SetupBackupStorageWebhookWithManager registers the webhook for BackupStorage in the manager.
func SetupBackupStorageWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha1.BackupStorage{}).
		WithValidator(&BackupStorageCustomWebhook{}).
		WithDefaulter(&BackupStorageCustomWebhook{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-storage-kubestash-com-v1alpha1-backupstorage,mutating=true,failurePolicy=fail,sideEffects=None,groups=storage.kubestash.com,resources=backupstorages,verbs=create;update,versions=v1alpha1,name=mbackupstorage.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &BackupStorageCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (_ *BackupStorageCustomWebhook) Default(_ context.Context, obj runtime.Object) error {
	var ok bool
	var s BackupStorage
	s.BackupStorage, ok = obj.(*v1alpha1.BackupStorage)
	if !ok {
		return fmt.Errorf("expected BackupStorage but got %T", obj)
	}
	backupstoragelog.Info("default", "name", s.Name)

	if s.Spec.UsagePolicy == nil {
		s.setDefaultUsagePolicy()
	}
	s.removeTrailingSlash()
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-storage-kubestash-com-v1alpha1-backupstorage,mutating=false,failurePolicy=fail,sideEffects=None,groups=storage.kubestash.com,resources=backupstorages,verbs=create;update,versions=v1alpha1,name=vbackupstorage.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &BackupStorageCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (_ *BackupStorageCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var b BackupStorage
	b.BackupStorage, ok = obj.(*v1alpha1.BackupStorage)
	if !ok {
		return nil, fmt.Errorf("expected BackupStorage but got %T", obj)
	}
	backupstoragelog.Info("Validation for BackupStorage upon creation", "name", b.Name)

	c := apis.GetRuntimeClient()

	if b.Spec.Default {
		if err := b.validateSingleDefaultBackupStorageInSameNamespace(ctx, c); err != nil {
			return nil, err
		}
	}

	if err := b.validateUsagePolicy(); err != nil {
		return nil, err
	}

	return nil, b.validateUniqueDirectory(ctx, c)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (_ *BackupStorageCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var bNew, bOld BackupStorage
	bNew.BackupStorage, ok = newObj.(*v1alpha1.BackupStorage)
	if !ok {
		return nil, fmt.Errorf("expected BackupStorage but got %T", newObj)
	}
	backupstoragelog.Info("Validation for BackupStorage upon update", "name", bNew.Name)

	bOld.BackupStorage, ok = oldObj.(*v1alpha1.BackupStorage)
	if !ok {
		return nil, fmt.Errorf("expected BackupStorage but got %T", oldObj)
	}

	c := apis.GetRuntimeClient()
	if bNew.Spec.Default {
		if err := bNew.validateSingleDefaultBackupStorageInSameNamespace(ctx, c); err != nil {
			return nil, err
		}
	}

	if err := bNew.validateUsagePolicy(); err != nil {
		return nil, err
	}

	if err := bNew.validateUpdateStorage(bOld.BackupStorage); err != nil {
		return nil, err
	}

	return nil, bNew.validateUniqueDirectory(ctx, c)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (_ *BackupStorageCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var s BackupStorage
	s.BackupStorage, ok = obj.(*v1alpha1.BackupStorage)
	if !ok {
		return nil, fmt.Errorf("expected BackupStorage but got %T", obj)
	}
	backupstoragelog.Info("Validation for BackupStorage upon delete", "name", s.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (b *BackupStorage) removeTrailingSlash() {
	if b.Spec.Storage.S3 != nil {
		b.Spec.Storage.S3.Bucket = strings.TrimSuffix(b.Spec.Storage.S3.Bucket, "/")
		b.Spec.Storage.S3.Endpoint = strings.TrimSuffix(b.Spec.Storage.S3.Endpoint, "/")
		b.Spec.Storage.S3.Prefix = strings.TrimSuffix(b.Spec.Storage.S3.Prefix, "/")
	}
	if b.Spec.Storage.GCS != nil {
		b.Spec.Storage.GCS.Bucket = strings.TrimSuffix(b.Spec.Storage.GCS.Bucket, "/")
		b.Spec.Storage.GCS.Prefix = strings.TrimSuffix(b.Spec.Storage.GCS.Prefix, "/")
	}
	if b.Spec.Storage.Azure != nil {
		b.Spec.Storage.Azure.Prefix = strings.TrimSuffix(b.Spec.Storage.Azure.Prefix, "/")
	}
}

func (b *BackupStorage) setDefaultUsagePolicy() {
	fromSameNamespace := apis.NamespacesFromSame
	b.Spec.UsagePolicy = &apis.UsagePolicy{
		AllowedNamespaces: apis.AllowedNamespaces{
			From: &fromSameNamespace,
		},
	}
}

func (b *BackupStorage) validateSingleDefaultBackupStorageInSameNamespace(ctx context.Context, c client.Client) error {
	bsList := v1alpha1.BackupStorageList{}
	if err := c.List(ctx, &bsList, client.InNamespace(b.Namespace)); err != nil {
		return err
	}

	for _, bs := range bsList.Items {
		if !b.isSameBackupStorage(bs) &&
			bs.Spec.Default {
			return fmt.Errorf("multiple default BackupStorages are not allowed within the same namespace")
		}
	}

	return nil
}

func (b *BackupStorage) validateUsagePolicy() error {
	if *b.Spec.UsagePolicy.AllowedNamespaces.From == apis.NamespacesFromSelector &&
		b.Spec.UsagePolicy.AllowedNamespaces.Selector == nil {
		return fmt.Errorf("selector cannot be empty for usage policy of type %q", apis.NamespacesFromSelector)
	}
	return nil
}

func (b *BackupStorage) isSameBackupStorage(bs v1alpha1.BackupStorage) bool {
	if b.Namespace == bs.Namespace &&
		b.Name == bs.Name {
		return true
	}
	return false
}

func (b *BackupStorage) validateUpdateStorage(old *v1alpha1.BackupStorage) error {
	if !reflect.DeepEqual(old.Spec.Storage, b.Spec.Storage) &&
		len(b.Status.Repositories) != 0 {
		return fmt.Errorf("BackupStorage is currently in use and cannot be modified")
	}
	return nil
}

func (b *BackupStorage) validateUniqueDirectory(ctx context.Context, c client.Client) error {
	bsList := v1alpha1.BackupStorageList{}
	if err := c.List(ctx, &bsList); err != nil {
		return err
	}

	for _, bs := range bsList.Items {
		if !b.isSameBackupStorage(bs) &&
			b.isPointToSameDir(bs) {
			return fmt.Errorf("no two BackupStorage should point to the same directory of the same bucket")
		}
	}

	return nil
}

func (b *BackupStorage) isPointToSameDir(bs v1alpha1.BackupStorage) bool {
	if b.Spec.Storage.Provider != bs.Spec.Storage.Provider {
		return false
	}

	switch b.Spec.Storage.Provider {
	case v1alpha1.ProviderS3:
		if b.Spec.Storage.S3.Bucket == bs.Spec.Storage.S3.Bucket &&
			b.Spec.Storage.S3.Region == bs.Spec.Storage.S3.Region &&
			b.Spec.Storage.S3.Prefix == bs.Spec.Storage.S3.Prefix {
			return true
		}
		return false
	case v1alpha1.ProviderGCS:
		if b.Spec.Storage.GCS.Bucket == bs.Spec.Storage.GCS.Bucket &&
			b.Spec.Storage.GCS.Prefix == bs.Spec.Storage.GCS.Prefix {
			return true
		}
		return false
	case v1alpha1.ProviderAzure:
		if b.Spec.Storage.Azure.StorageAccount == bs.Spec.Storage.Azure.StorageAccount &&
			b.Spec.Storage.Azure.Container == bs.Spec.Storage.Azure.Container &&
			b.Spec.Storage.Azure.Prefix == bs.Spec.Storage.Azure.Prefix {
			return true
		}
		return false
	default:
		return false
	}
}
