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

	"kubestash.dev/apimachinery/apis/core/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var backupbatchlog = logf.Log.WithName("backupbatch-resource")

type (
	BackupBatchCustomWebhook struct{}
	BackupBatch              struct {
		*v1alpha1.BackupBatch
	}
)

// SetupBackupBatchWebhookWithManager registers the webhook for BackupBatch in the manager.
func SetupBackupBatchWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha1.BackupBatch{}).
		WithValidator(&BackupBatchCustomWebhook{}).
		// WithDefaulter(&BackupBatchCustomWebhook{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-kubestash-com-v1alpha1-backupbatch,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=backupbatches,verbs=create;update,versions=v1alpha1,name=vbackupbatch.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &BackupBatchCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (*BackupBatchCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var b BackupBatch
	b.BackupBatch, ok = obj.(*v1alpha1.BackupBatch)
	if !ok {
		return nil, fmt.Errorf("expected BackupBatch but got %T", obj)
	}
	backupbatchlog.Info("Validation for BackupBatch upon creation", "name", b.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (*BackupBatchCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var bNew, bOld BackupBatch
	bNew.BackupBatch, ok = newObj.(*v1alpha1.BackupBatch)
	if !ok {
		return nil, fmt.Errorf("expected BackupBatch but got %T", newObj)
	}
	backupbatchlog.Info("Validation for BackupBatch upon update", "name", bNew.Name)

	bOld.BackupBatch, ok = oldObj.(*v1alpha1.BackupBatch)
	if !ok {
		return nil, fmt.Errorf("expected BackupBatch but got %T", oldObj)
	}

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (*BackupBatchCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var b BackupBatch
	b.BackupBatch, ok = obj.(*v1alpha1.BackupBatch)
	if !ok {
		return nil, fmt.Errorf("expected BackupBatch but got %T", obj)
	}
	backupbatchlog.Info("Validation for BackupBatch upon delete", "name", b.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
