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
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var restoresessionlog = logf.Log.WithName("restoresession-resource")

func (r *RestoreSession) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-kubestash-com-v1alpha1-restoresession,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=restoresessions,verbs=create;update,versions=v1alpha1,name=vrestoresession.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &RestoreSession{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *RestoreSession) ValidateCreate() error {
	restoresessionlog.Info("validate create", "name", r.Name)

	if err := r.checkEmptySnapshot(); err != nil {
		return err
	}

	if err := r.checkEmptyRepositoryForLatestSnapshot(); err != nil {
		return err
	}

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *RestoreSession) ValidateUpdate(old runtime.Object) error {
	restoresessionlog.Info("validate update", "name", r.Name)

	if err := r.checkEmptySnapshot(); err != nil {
		return err
	}

	if err := r.checkEmptyRepositoryForLatestSnapshot(); err != nil {
		return err
	}

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *RestoreSession) ValidateDelete() error {
	restoresessionlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *RestoreSession) checkEmptySnapshot() error {
	if r.Spec.DataSource.Snapshot == "" {
		return fmt.Errorf("snapshot can not be empty")
	}
	return nil
}

func (r *RestoreSession) checkEmptyRepositoryForLatestSnapshot() error {
	if r.Spec.DataSource.Snapshot == "latest" &&
		r.Spec.DataSource.Repository == "" {
		return fmt.Errorf("reposity can not be empty for latest snapshot")
	}
	return nil
}
