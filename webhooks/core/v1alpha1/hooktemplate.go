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
	"kubestash.dev/apimachinery/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
)

// log is for logging in this package.
var hooktemplatelog = logf.Log.WithName("hooktemplate-resource")

type HookTemplateCustomDefaulter struct{}
type HookTemplateCustomValidator struct{}

type HookTemplate struct {
	*v1alpha1.HookTemplate
}

// SetupHookTemplateWebhookWithManager registers the webhook for HookTemplate in the manager.
func SetupHookTemplateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha1.HookTemplate{}).
		WithValidator(&HookTemplateCustomValidator{}).
		WithDefaulter(&HookTemplateCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-core-kubestash-com-v1alpha1-hooktemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=hooktemplates,verbs=create;update,versions=v1alpha1,name=mhooktemplate.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &HookTemplateCustomDefaulter{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (_ *HookTemplateCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	var ok bool
	var h HookTemplate
	h.HookTemplate, ok = obj.(*v1alpha1.HookTemplate)
	if !ok {
		return fmt.Errorf("expected HookTemplate but got %T", obj)
	}
	hooktemplatelog.Info("default", "name", h.Name)

	if h.Spec.UsagePolicy == nil {
		h.setDefaultUsagePolicy()
	}
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-kubestash-com-v1alpha1-hooktemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=hooktemplates,verbs=create;update,versions=v1alpha1,name=vhooktemplate.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &HookTemplateCustomValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (_ *HookTemplateCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var h HookTemplate
	h.HookTemplate, ok = obj.(*v1alpha1.HookTemplate)
	if !ok {
		return nil, fmt.Errorf("expected HookTemplate but got %T", obj)
	}
	hooktemplatelog.Info("Validation for HookTemplate upon creation", "name", h.Name)

	if h.Spec.Executor == nil {
		return nil, fmt.Errorf("executor can not be empty")
	}

	if err := h.validateActionForNonFunctionExecutor(); err != nil {
		return nil, err
	}

	if err := h.validateUsagePolicy(); err != nil {
		return nil, err
	}

	return nil, h.validateExecutorInfo()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (_ *HookTemplateCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var hNew, hOld HookTemplate
	hNew.HookTemplate, ok = newObj.(*v1alpha1.HookTemplate)
	if !ok {
		return nil, fmt.Errorf("expected HookTemplate but got %T", newObj)
	}
	hooktemplatelog.Info("Validation for HookTemplate upon update", "name", hNew.Name)

	hOld.HookTemplate, ok = oldObj.(*v1alpha1.HookTemplate)
	if !ok {
		return nil, fmt.Errorf("expected HookTemplate but got %T", oldObj)
	}

	if hNew.Spec.Executor == nil {
		return nil, fmt.Errorf("executor field can not be empty")
	}

	if err := hNew.validateActionForNonFunctionExecutor(); err != nil {
		return nil, err
	}

	if err := hNew.validateUsagePolicy(); err != nil {
		return nil, err
	}

	return nil, hNew.validateExecutorInfo()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HookTemplateCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var s HookTemplate
	s.HookTemplate, ok = obj.(*v1alpha1.HookTemplate)
	if !ok {
		return nil, fmt.Errorf("expected HookTemplate but got %T", obj)
	}
	hooktemplatelog.Info("Validation for HookTemplate upon creation", "name", s.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *HookTemplate) setDefaultUsagePolicy() {
	fromSameNamespace := apis.NamespacesFromSame
	r.Spec.UsagePolicy = &apis.UsagePolicy{
		AllowedNamespaces: apis.AllowedNamespaces{
			From: &fromSameNamespace,
		},
	}
}

func (r *HookTemplate) validateExecutorInfo() error {
	if r.Spec.Executor.Type == v1alpha1.HookExecutorFunction {
		if r.Spec.Executor.Function == nil {
			return fmt.Errorf("function field can not be empty for function type executor")
		}
	}

	if r.Spec.Executor.Type == v1alpha1.HookExecutorPod {
		if r.Spec.Executor.Pod == nil {
			return fmt.Errorf("pod field can not be empty for pod type executor")
		}
		if r.Spec.Executor.Pod.Selector == "" {
			return fmt.Errorf("selector field can not be empty for pod type executor")
		}

		selectors := strings.Split(r.Spec.Executor.Pod.Selector, ",")
		for _, sel := range selectors {
			if len(strings.Split(strings.Trim(sel, " "), "=")) < 2 {
				return fmt.Errorf("invalid selector is provided for pod type executor")
			}
		}
	}
	return nil
}

func (r *HookTemplate) validateActionForNonFunctionExecutor() error {
	if r.Spec.Executor.Type != v1alpha1.HookExecutorFunction &&
		r.Spec.Action == nil {
		return fmt.Errorf("action can not be empty for pod or operator type executor")
	}
	return nil
}

func (r *HookTemplate) validateUsagePolicy() error {
	if *r.Spec.UsagePolicy.AllowedNamespaces.From == apis.NamespacesFromSelector &&
		r.Spec.UsagePolicy.AllowedNamespaces.Selector == nil {
		return fmt.Errorf("selector cannot be empty for usage policy of type %q", apis.NamespacesFromSelector)
	}
	return nil
}
