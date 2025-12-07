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

	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/apis/core/v1alpha1"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kmapi "kmodules.xyz/client-go/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var backupblueprintlog = logf.Log.WithName("backupblueprint-resource")

type BackupBlueprintCustomWebhook struct{}

type BackupBlueprint struct {
	*v1alpha1.BackupBlueprint
}

// SetupBackupBlueprintWebhookWithManager registers the webhook for BackupBlueprint in the manager.
func SetupBackupBlueprintWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha1.BackupBlueprint{}).
		WithValidator(&BackupBlueprintCustomWebhook{}).
		WithDefaulter(&BackupBlueprintCustomWebhook{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-core-kubestash-com-v1alpha1-backupblueprint,mutating=true,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=backupblueprints,verbs=create;update,versions=v1alpha1,name=mbackupblueprint.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &BackupBlueprintCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (*BackupBlueprintCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	var ok bool
	var r BackupBlueprint
	r.BackupBlueprint, ok = obj.(*v1alpha1.BackupBlueprint)
	if !ok {
		return fmt.Errorf("expected BackupBlueprint but got %T", obj)
	}

	backupblueprintlog.Info("default", "name", r.Name)

	if r.Spec.UsagePolicy == nil {
		r.setDefaultUsagePolicy()
	}
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-kubestash-com-v1alpha1-backupblueprint,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=backupblueprints,verbs=create;update,versions=v1alpha1,name=vbackupblueprint.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &BackupBlueprintCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (*BackupBlueprintCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var b BackupBlueprint
	b.BackupBlueprint, ok = obj.(*v1alpha1.BackupBlueprint)
	if !ok {
		return nil, fmt.Errorf("expected BackupBlueprint but got %T", obj)
	}
	backupblueprintlog.Info("Validation for BackupBlueprint upon creation", "name", b.Name)

	if err := b.validateUsagePolicy(); err != nil {
		return nil, err
	}

	return nil, b.validateBackendsAgainstUsagePolicy(ctx, apis.GetRuntimeClient())
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (*BackupBlueprintCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var bNew, bold BackupBlueprint
	bNew.BackupBlueprint, ok = newObj.(*v1alpha1.BackupBlueprint)
	if !ok {
		return nil, fmt.Errorf("expected BackupBlueprint but got %T", newObj)
	}
	backupblueprintlog.Info("Validation for BackupBlueprint upon update", "name", bNew.Name)

	bold.BackupBlueprint, ok = oldObj.(*v1alpha1.BackupBlueprint)
	if !ok {
		return nil, fmt.Errorf("expected BackupBlueprint but got %T", oldObj)
	}

	if err := bNew.validateUsagePolicy(); err != nil {
		return nil, err
	}

	return nil, bNew.validateBackendsAgainstUsagePolicy(ctx, apis.GetRuntimeClient())
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (*BackupBlueprintCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	var ok bool
	var b BackupBlueprint
	b.BackupBlueprint, ok = obj.(*v1alpha1.BackupBlueprint)
	if !ok {
		return nil, fmt.Errorf("expected BackupBlueprint but got %T", obj)
	}
	backupblueprintlog.Info("Validation for BackupBlueprint upon delete", "name", b.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (b *BackupBlueprint) setDefaultUsagePolicy() {
	fromSameNamespace := apis.NamespacesFromSame
	b.Spec.UsagePolicy = &apis.UsagePolicy{
		AllowedNamespaces: apis.AllowedNamespaces{
			From: &fromSameNamespace,
		},
	}
}

func (b *BackupBlueprint) validateBackendsAgainstUsagePolicy(ctx context.Context, c client.Client) error {
	if b.Spec.BackupConfigurationTemplate == nil {
		return fmt.Errorf("backupConfigurationTemplate can not be empty")
	}

	for _, backend := range b.Spec.BackupConfigurationTemplate.Backends {
		bs, err := b.getBackupStorage(ctx, c, backend.StorageRef)
		if err != nil {
			if kerr.IsNotFound(err) {
				continue
			}
			return err
		}

		ns := &core.Namespace{ObjectMeta: v1.ObjectMeta{Name: b.Namespace}}
		if err := c.Get(ctx, client.ObjectKeyFromObject(ns), ns); err != nil {
			return err
		}

		if !bs.UsageAllowed(ns) {
			return fmt.Errorf("namespace %q is not allowed to refer BackupStorage %s/%s. Please, check the `usagePolicy` of the BackupStorage", b.Namespace, bs.Name, bs.Namespace)
		}
	}
	return nil
}

func (b *BackupBlueprint) getBackupStorage(ctx context.Context, c client.Client, ref *kmapi.ObjectReference) (*storageapi.BackupStorage, error) {
	bs := &storageapi.BackupStorage{
		ObjectMeta: v1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
		},
	}

	if bs.Namespace == "" {
		bs.Namespace = b.Namespace
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(bs), bs); err != nil {
		return nil, err
	}
	return bs, nil
}

func (b *BackupBlueprint) validateUsagePolicy() error {
	if *b.Spec.UsagePolicy.AllowedNamespaces.From == apis.NamespacesFromSelector &&
		b.Spec.UsagePolicy.AllowedNamespaces.Selector == nil {
		return fmt.Errorf("selector cannot be empty for usage policy of type %q", apis.NamespacesFromSelector)
	}
	return nil
}
