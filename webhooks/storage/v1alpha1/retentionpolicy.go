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
	"k8s.io/utils/pointer"
	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/apis/storage/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var retentionpolicylog = logf.Log.WithName("retentionpolicy-resource")

type RetentionPolicyCustomDefaulter struct{}
type RetentionPolicyCustomValidator struct{}

type RetentionPolicy struct {
	*v1alpha1.RetentionPolicy
}

// SetupRetentionPolicyWebhookWithManager registers the webhook for RetentionPolicy in the manager.
func SetupRetentionPolicyWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha1.RetentionPolicy{}).
		WithValidator(&RetentionPolicyCustomValidator{}).
		WithDefaulter(&RetentionPolicyCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-storage-kubestash-com-v1alpha1-retentionpolicy,mutating=true,failurePolicy=fail,sideEffects=None,groups=storage.kubestash.com,resources=retentionpolicies,verbs=create;update,versions=v1alpha1,name=mretentionpolicy.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &RetentionPolicyCustomDefaulter{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (_ *RetentionPolicyCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	var ok bool
	var r RetentionPolicy
	r.RetentionPolicy, ok = obj.(*v1alpha1.RetentionPolicy)
	if !ok {
		return fmt.Errorf("expected RetentionPolicy but got %T", obj)
	}
	retentionpolicylog.Info("default", "name", r.Name)

	if r.Spec.UsagePolicy == nil {
		r.setDefaultUsagePolicy()
	}

	if r.Spec.FailedSnapshots == nil {
		r.setDefaultFailedSnapshots()
	}
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-storage-kubestash-com-v1alpha1-retentionpolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=storage.kubestash.com,resources=retentionpolicies,verbs=create;update,versions=v1alpha1,name=vretentionpolicy.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &RetentionPolicyCustomValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (_ *RetentionPolicyCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var r RetentionPolicy
	r.RetentionPolicy, ok = obj.(*v1alpha1.RetentionPolicy)
	if !ok {
		return nil, fmt.Errorf("expected RetentionPolicy but got %T", obj)
	}
	retentionpolicylog.Info("Validation for RetentionPolicy upon creation", "name", r.Name)

	c := apis.GetRuntimeClient()

	if err := r.validateMaxRetentionPeriodFormat(); err != nil {
		return nil, err
	}

	if err := r.validateProvidedPolicy(); err != nil {
		return nil, err
	}

	if err := r.validateUsagePolicy(); err != nil {
		return nil, err
	}

	return nil, r.validateSingleDefaultRetentionPolicyInSameNamespace(ctx, c)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (_ *RetentionPolicyCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var rNew, rOld RetentionPolicy
	rNew.RetentionPolicy, ok = newObj.(*v1alpha1.RetentionPolicy)
	if !ok {
		return nil, fmt.Errorf("expected RetentionPolicy but got %T", newObj)
	}
	retentionpolicylog.Info("Validation for RetentionPolicy upon update", "name", rNew.Name)

	rOld.RetentionPolicy, ok = oldObj.(*v1alpha1.RetentionPolicy)
	if !ok {
		return nil, fmt.Errorf("expected RetentionPolicy but got %T", oldObj)
	}

	c := apis.GetRuntimeClient()

	if err := rNew.validateMaxRetentionPeriodFormat(); err != nil {
		return nil, err
	}

	if err := rNew.validateProvidedPolicy(); err != nil {
		return nil, err
	}

	if err := rNew.validateUsagePolicy(); err != nil {
		return nil, err
	}

	return nil, rNew.validateSingleDefaultRetentionPolicyInSameNamespace(ctx, c)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (_ *RetentionPolicyCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var r RetentionPolicy
	r.RetentionPolicy, ok = obj.(*v1alpha1.RetentionPolicy)
	if !ok {
		return nil, fmt.Errorf("expected RetentionPolicy but got %T", obj)
	}
	retentionpolicylog.Info("Validation for RetentionPolicy upon delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *RetentionPolicy) setDefaultUsagePolicy() {
	fromSameNamespace := apis.NamespacesFromSame
	r.Spec.UsagePolicy = &apis.UsagePolicy{
		AllowedNamespaces: apis.AllowedNamespaces{
			From: &fromSameNamespace,
		},
	}
}

func (r *RetentionPolicy) setDefaultFailedSnapshots() {
	r.Spec.FailedSnapshots = &v1alpha1.FailedSnapshotsKeepPolicy{
		Last: pointer.Int32(1),
	}
}

func (r *RetentionPolicy) validateMaxRetentionPeriodFormat() error {
	if r.Spec.MaxRetentionPeriod != "" {
		_, err := v1alpha1.ParseDuration(string(r.Spec.MaxRetentionPeriod))
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RetentionPolicy) validateProvidedPolicy() error {
	if r.Spec.MaxRetentionPeriod == "" &&
		r.Spec.SuccessfulSnapshots == nil {
		return fmt.Errorf("one of maxRetentionPeriod and successfulSnapshots policy must be provided")
	}
	return nil
}

func (r *RetentionPolicy) validateSingleDefaultRetentionPolicyInSameNamespace(ctx context.Context, c client.Client) error {
	if !r.Spec.Default {
		return nil
	}

	rpList := v1alpha1.RetentionPolicyList{}
	if err := c.List(ctx, &rpList, client.InNamespace(r.Namespace)); err != nil {
		return err
	}

	for _, rp := range rpList.Items {
		if !r.isSameRetentionPolicy(rp) &&
			rp.Spec.Default {
			return fmt.Errorf("multiple default RetentionPolicies are not allowed within the same namespace")
		}
	}

	return nil
}

func (r *RetentionPolicy) isSameRetentionPolicy(rp v1alpha1.RetentionPolicy) bool {
	if r.Namespace == rp.Namespace &&
		r.Name == rp.Name {
		return true
	}
	return false
}

func (r *RetentionPolicy) validateUsagePolicy() error {
	if *r.Spec.UsagePolicy.AllowedNamespaces.From == apis.NamespacesFromSelector &&
		r.Spec.UsagePolicy.AllowedNamespaces.Selector == nil {
		return fmt.Errorf("selector cannot be empty for usage policy of type %q", apis.NamespacesFromSelector)
	}
	return nil
}
