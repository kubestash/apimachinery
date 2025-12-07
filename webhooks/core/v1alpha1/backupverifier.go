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
var backupverifierlog = logf.Log.WithName("backupverifier-resource")

type BackupVerifierCustomWebhook struct{}

type BackupVerifier struct {
	*v1alpha1.BackupVerifier
}

// SetupBackupVerifierWebhookWithManager registers the webhook for BackupVerifier in the manager.
func SetupBackupVerifierWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha1.BackupVerifier{}).
		WithValidator(&BackupVerifierCustomWebhook{}).
		// WithDefaulter(&BackupVerifierCustomWebhook{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-kubestash-com-v1alpha1-backupverifier,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=backupverifiers,verbs=create;update,versions=v1alpha1,name=vbackupverifier.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &BackupVerifierCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (*BackupVerifierCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var b BackupVerifier
	b.BackupVerifier, ok = obj.(*v1alpha1.BackupVerifier)
	if !ok {
		return nil, fmt.Errorf("expected BackupVerifier but got %T", obj)
	}
	backupverifierlog.Info("Validation for BackupVerifier upon creation", "name", b.Name)

	if err := b.validateVerifier(); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (*BackupVerifierCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var bNew, bOld BackupVerifier
	bNew.BackupVerifier, ok = newObj.(*v1alpha1.BackupVerifier)
	if !ok {
		return nil, fmt.Errorf("expected BackupVerifier but got %T", newObj)
	}
	backupverifierlog.Info("Validation for BackupVerifier upon update", "name", bNew.Name)

	bOld.BackupVerifier, ok = oldObj.(*v1alpha1.BackupVerifier)
	if !ok {
		return nil, fmt.Errorf("expected BackupVerifier but got %T", oldObj)
	}

	if err := bNew.validateVerifier(); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (*BackupVerifierCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var b BackupVerifier
	b.BackupVerifier, ok = obj.(*v1alpha1.BackupVerifier)
	if !ok {
		return nil, fmt.Errorf("expected BackupVerifier but got %T", obj)
	}
	backupverifierlog.Info("Validation for BackupVerifier upon delete", "name", b.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (b *BackupVerifier) validateVerifier() error {
	if b.Spec.RestoreOption == nil {
		return fmt.Errorf("restoreOption for backupVerifier %s/%s cannot be empty", b.Namespace, b.Name)
	}

	if b.Spec.RestoreOption.AddonInfo == nil {
		return fmt.Errorf("addonInfo in restoreOption for backupVerifier %s/%s cannot be empty", b.Namespace, b.Name)
	}

	if b.Spec.Scheduler != nil {
		return fmt.Errorf("scheduler for backupVerifier %s/%s cannot be empty", b.Namespace, b.Name)
	}

	if b.Spec.Type == "" {
		return fmt.Errorf("type of backupVerifier %s/%s cannot be empty", b.Namespace, b.Name)
	}

	if b.Spec.Type == v1alpha1.QueryVerificationType {
		if b.Spec.Query == nil {
			return fmt.Errorf("query in backupVerifier %s/%s cannot be empty", b.Namespace, b.Name)
		}
		if b.Spec.Function == "" {
			return fmt.Errorf("function in backupVerifier %s/%s cannot be empty", b.Namespace, b.Name)
		}
	}

	if b.Spec.Type == v1alpha1.ScriptVerificationType {
		if b.Spec.Script == nil {
			return fmt.Errorf("script in backupVerifier %s/%s cannot be empty", b.Namespace, b.Name)
		}

		if b.Spec.Script.Location == "" {
			return fmt.Errorf("script location in backupVerifier %s/%s cannot be empty", b.Namespace, b.Name)
		}

		if b.Spec.Function == "" {
			return fmt.Errorf("function in backupVerifier %s/%s cannot be empty", b.Namespace, b.Name)
		}
	}

	return nil
}
