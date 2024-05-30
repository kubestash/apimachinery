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
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"kubestash.dev/apimachinery/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var backupverificationlog = logf.Log.WithName("backupverification-resource")

func (r *BackupVerification) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-core-kubestash-com-v1alpha1-backupverification,mutating=true,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=backupverifications,verbs=create;update,versions=v1alpha1,name=mbackupverification.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &BackupVerification{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *BackupVerification) Default() {
	backupverificationlog.Info("default", "name", r.Name)

	if r.Spec.UsagePolicy == nil {
		r.setDefaultUsagePolicy()
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-kubestash-com-v1alpha1-backupverification,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=backupverifications,verbs=create;update,versions=v1alpha1,name=vbackupverification.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &BackupVerification{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *BackupVerification) ValidateCreate() (admission.Warnings, error) {
	backupverificationlog.Info("validate create", "name", r.Name)

	if r.Spec.Function == "" {
		return nil, fmt.Errorf("function is required")
	}

	if r.Spec.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	if err := r.validateVerificationType(); err != nil {
		return nil, err
	}

	return nil, r.validateUsagePolicy()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *BackupVerification) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	backupverificationlog.Info("validate update", "name", r.Name)

	if r.Spec.Function == "" {
		return nil, fmt.Errorf("function is required")
	}

	if r.Spec.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	if err := r.validateVerificationType(); err != nil {
		return nil, err
	}

	return nil, r.validateUsagePolicy()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *BackupVerification) ValidateDelete() (admission.Warnings, error) {
	backupverificationlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *BackupVerification) setDefaultUsagePolicy() {
	fromSameNamespace := apis.NamespacesFromSame
	r.Spec.UsagePolicy = &apis.UsagePolicy{
		AllowedNamespaces: apis.AllowedNamespaces{
			From: &fromSameNamespace,
		},
	}
}

func (r *BackupVerification) validateVerificationType() error {
	switch r.Spec.Type {
	case FileVerificationType:
		if r.Spec.File == nil {
			return fmt.Errorf("file field can not be empty for file verification type")
		}
	case QueryVerificationType:
		if r.Spec.Query == nil {
			return fmt.Errorf("query field can not be empty for query verification type")
		}
	case ScriptVerificationType:
		if r.Spec.Script == nil {
			return fmt.Errorf("script field can not be empty for script verification type")
		}
	}
	return nil
}

func (r *BackupVerification) validateUsagePolicy() error {
	if *r.Spec.UsagePolicy.AllowedNamespaces.From == apis.NamespacesFromSelector &&
		r.Spec.UsagePolicy.AllowedNamespaces.Selector == nil {
		return fmt.Errorf("selector cannot be empty for usage policy of type %q", apis.NamespacesFromSelector)
	}
	return nil
}
