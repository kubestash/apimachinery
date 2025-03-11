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
	"kubestash.dev/apimachinery/apis/storage/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var repositorylog = logf.Log.WithName("repository-resource")

type RepositoryCustomWebhook struct{}

type Repository struct {
	*v1alpha1.Repository
}

// SetupRepositoryWebhookWithManager registers the webhook for Repository in the manager.
func SetupRepositoryWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha1.Repository{}).
		WithValidator(&RepositoryCustomWebhook{}).
		WithDefaulter(&RepositoryCustomWebhook{}).
		Complete()
}

//+kubebuilder:webhook:path=/validate-storage-kubestash-com-v1alpha1-repository,mutating=true,failurePolicy=fail,sideEffects=None,groups=storage.kubestash.com,resources=repositories,verbs=create;update,versions=v1alpha1,name=vrepository.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &RepositoryCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (_ *RepositoryCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	var ok bool
	var r Repository
	r.Repository, ok = obj.(*v1alpha1.Repository)
	if !ok {
		return fmt.Errorf("expected Repository but got %T", obj)
	}
	repositorylog.Info("default", "name", r.Name)

	// TODO(user): fill in your default logic.
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-storage-kubestash-com-v1alpha1-repository,mutating=false,failurePolicy=fail,sideEffects=None,groups=storage.kubestash.com,resources=repositories,verbs=create;update,versions=v1alpha1,name=vrepository.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &RepositoryCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (_ *RepositoryCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var r Repository
	r.Repository, ok = obj.(*v1alpha1.Repository)
	if !ok {
		return nil, fmt.Errorf("expected Repository but got %T", obj)
	}
	repositorylog.Info("Validation for Repository upon creation", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (_ *RepositoryCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var rNew, rOld Repository
	rNew.Repository, ok = newObj.(*v1alpha1.Repository)
	if !ok {
		return nil, fmt.Errorf("expected Repository but got %T", newObj)
	}
	repositorylog.Info("Validation for Repository upon creation", "name", rNew.Name)

	rOld.Repository, ok = oldObj.(*v1alpha1.Repository)
	if !ok {
		return nil, fmt.Errorf("expected Repository but got %T", oldObj)
	}

	if rOld.Spec.Path != rNew.Spec.Path {
		return nil, fmt.Errorf("repository path can not be updated")
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (_ *RepositoryCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var r Repository
	r.Repository, ok = obj.(*v1alpha1.Repository)
	if !ok {
		return nil, fmt.Errorf("expected Repository but got %T", obj)
	}
	repositorylog.Info("Validation for Repository upon creation", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
