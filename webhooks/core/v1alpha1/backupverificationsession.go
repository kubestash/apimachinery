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
	"kubestash.dev/apimachinery/apis/core/v1alpha1"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var backupverificationsessionlog = logf.Log.WithName("backupverificationsession-resource")

type BackupVerificationSessionCustomWebhook struct{}

type BackupVerificationSession struct {
	*v1alpha1.BackupVerificationSession
}

// SetupBackupVerificationSessionWebhookWithManager registers the webhook for BackupVerificationSession in the manager.
func SetupBackupVerificationSessionWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha1.BackupVerificationSession{}).
		WithValidator(&BackupVerificationSessionCustomWebhook{}).
		//WithDefaulter(&BackupVerificationSessionCustomWebhook{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-kubestash-com-v1alpha1-backupverificationsession,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=backupverificationsessions,verbs=create;update,versions=v1alpha1,name=vbackupverificationsession.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &BackupVerificationSessionCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (_ *BackupVerificationSessionCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var b BackupVerificationSession
	b.BackupVerificationSession, ok = obj.(*v1alpha1.BackupVerificationSession)
	if !ok {
		return nil, fmt.Errorf("expected BackupVerificationSession but got %T", obj)
	}
	backupverificationsessionlog.Info("Validation for BackupVerificationSession upon creation", "name", b.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (_ *BackupVerificationSessionCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var bNew, bOld BackupVerificationSession
	bNew.BackupVerificationSession, ok = newObj.(*v1alpha1.BackupVerificationSession)
	if !ok {
		return nil, fmt.Errorf("expected BackupVerificationSession but got %T", newObj)
	}
	backupverificationsessionlog.Info("Validation for BackupVerificationSession upon update", "name", bNew.Name)

	bOld.BackupVerificationSession, ok = oldObj.(*v1alpha1.BackupVerificationSession)
	if !ok {
		return nil, fmt.Errorf("expected BackupVerificationSession but got %T", oldObj)
	}

	if !reflect.DeepEqual(bOld.Spec, bNew.Spec) {
		return nil, fmt.Errorf("spec can not be updated")
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (_ *BackupVerificationSessionCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var b BackupVerificationSession
	b.BackupVerificationSession, ok = obj.(*v1alpha1.BackupVerificationSession)
	if !ok {
		return nil, fmt.Errorf("expected BackupVerificationSession but got %T", obj)
	}
	backupverificationsessionlog.Info("Validation for BackupVerificationSession upon delete", "name", b.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
