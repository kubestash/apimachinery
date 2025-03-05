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
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var restoresessionlog = logf.Log.WithName("restoresession-resource")

type RestoreSessionCustomDefaulter struct{}
type RestoreSessionCustomValidator struct{}

type RestoreSession struct {
	*v1alpha1.RestoreSession
}

// SetupRestoreSessionWebhookWithManager registers the webhook for RestoreSession in the manager.
func SetupRestoreSessionWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha1.RestoreSession{}).
		WithValidator(&RestoreSessionCustomValidator{}).
		//WithDefaulter(&RestoreSessionCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-kubestash-com-v1alpha1-restoresession,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=restoresessions,verbs=create;update,versions=v1alpha1,name=vrestoresession.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &RestoreSessionCustomValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (_ *RestoreSessionCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var r RestoreSession
	r.RestoreSession, ok = obj.(*v1alpha1.RestoreSession)
	if !ok {
		return nil, fmt.Errorf("expected RestoreSession but got %T", obj)
	}
	restoresessionlog.Info("Validation for RestoreSession upon creation", "name", r.Name)

	if err := r.ValidateDataSource(); err != nil {
		return nil, err
	}

	return nil, r.validateHookTemplatesAgainstUsagePolicy(context.Background(), apis.GetRuntimeClient())
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (_ *RestoreSessionCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var rNew, rOld RestoreSession
	rNew.RestoreSession, ok = newObj.(*v1alpha1.RestoreSession)
	if !ok {
		return nil, fmt.Errorf("expected RestoreSession but got %T", newObj)
	}
	restoresessionlog.Info("Validation for RestoreSession upon update", "name", rNew.Name)

	rOld.RestoreSession, ok = oldObj.(*v1alpha1.RestoreSession)
	if !ok {
		return nil, fmt.Errorf("expected RestoreSession but got %T", oldObj)
	}

	if err := rNew.ValidateDataSource(); err != nil {
		return nil, err
	}

	return nil, rNew.validateHookTemplatesAgainstUsagePolicy(context.Background(), apis.GetRuntimeClient())
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (_ *RestoreSessionCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var s RestoreSession
	s.RestoreSession, ok = obj.(*v1alpha1.RestoreSession)
	if !ok {
		return nil, fmt.Errorf("expected RestoreSession but got %T", obj)
	}
	restoresessionlog.Info("Validation for RestoreSession upon delete", "name", s.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *RestoreSession) ValidateDataSource() error {
	if r.Spec.DataSource.PITR != nil {
		if err := r.checkIfTargetTimeIsNil(); err != nil {
			return err
		}
		if err := r.checkIfRepoIsEmptyForTargetTime(); err != nil {
			return err
		}
	} else {
		if err := r.checkIfSnapshotIsEmpty(); err != nil {
			return err
		}
		if err := r.checkIfRepoIsEmptyForLatestSnapshot(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RestoreSession) checkIfTargetTimeIsNil() error {
	if r.Spec.DataSource.PITR.TargetTime == nil {
		return fmt.Errorf("targetTime can not be empty for the Point-In-Time Recovery (PITR) feature")
	}
	return nil
}

func (r *RestoreSession) checkIfRepoIsEmptyForTargetTime() error {
	if r.Spec.DataSource.Repository == "" {
		return fmt.Errorf("repository can not be empty for the Point-In-Time Recovery (PITR) feature")
	}
	return nil
}

func (r *RestoreSession) checkIfSnapshotIsEmpty() error {
	if r.Spec.DataSource.Snapshot == "" {
		return fmt.Errorf("snapshot can not be empty")
	}
	return nil
}

func (r *RestoreSession) checkIfRepoIsEmptyForLatestSnapshot() error {
	if r.Spec.DataSource.Snapshot == "latest" &&
		r.Spec.DataSource.Repository == "" {
		return fmt.Errorf("repository can not be empty for latest snapshot")
	}
	return nil
}

func (r *RestoreSession) validateHookTemplatesAgainstUsagePolicy(ctx context.Context, c client.Client) error {
	hookTemplates := r.getHookTemplates()
	for _, ht := range hookTemplates {
		err := c.Get(ctx, client.ObjectKeyFromObject(&ht), &ht)
		if err != nil {
			if kerr.IsNotFound(err) {
				continue
			}
			return err
		}

		ns := &core.Namespace{ObjectMeta: metav1.ObjectMeta{Name: r.Namespace}}
		if err := c.Get(ctx, client.ObjectKeyFromObject(ns), ns); err != nil {
			return err
		}

		if !ht.UsageAllowed(ns) {
			return fmt.Errorf("namespace %q is not allowed to refer HookTemplate %s/%s. Please, check the `usagePolicy` of the HookTemplate", r.Namespace, ht.Name, ht.Namespace)
		}
	}
	return nil
}

func (r *RestoreSession) getHookTemplates() []v1alpha1.HookTemplate {
	var hookTemplates []v1alpha1.HookTemplate
	if r.Spec.Hooks != nil {
		hookTemplates = append(hookTemplates, r.getHookTemplatesFromHookInfo(r.Spec.Hooks.PreRestore)...)
		hookTemplates = append(hookTemplates, r.getHookTemplatesFromHookInfo(r.Spec.Hooks.PostRestore)...)
	}
	return hookTemplates
}

func (r *RestoreSession) getHookTemplatesFromHookInfo(hooks []v1alpha1.HookInfo) []v1alpha1.HookTemplate {
	var hookTemplates []v1alpha1.HookTemplate
	for _, hook := range hooks {
		if hook.HookTemplate != nil {
			hookTemplates = append(hookTemplates, v1alpha1.HookTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      hook.HookTemplate.Name,
					Namespace: hook.HookTemplate.Namespace,
				},
			})
		}
	}
	return hookTemplates
}
