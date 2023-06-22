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
	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/crds"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	"kmodules.xyz/client-go/meta"
)

func (_ BackupStorage) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralBackupStorage))
}

func (b *BackupStorage) CalculatePhase() BackupStoragePhase {
	if kmapi.IsConditionTrue(b.Status.Conditions, TypeBackendInitialized) {
		if !kmapi.HasCondition(b.Status.Conditions, TypeBackendSecretFound) {
			return BackupStorageReady
		}
		if kmapi.IsConditionTrue(b.Status.Conditions, TypeBackendSecretFound) {
			return BackupStorageReady
		}
	}
	return BackupStorageNotReady
}

func (b *BackupStorage) UsageAllowed(srcNamespace *core.Namespace) bool {
	if b.Spec.UsagePolicy == nil {
		return b.Namespace == srcNamespace.Name
	}
	return b.isNamespaceAllowed(srcNamespace)
}

func (b *BackupStorage) isNamespaceAllowed(srcNamespace *core.Namespace) bool {
	allowedNamespaces := b.Spec.UsagePolicy.AllowedNamespaces

	if allowedNamespaces.From == nil {
		return false
	}

	if *allowedNamespaces.From == apis.NamespacesFromAll {
		return true
	}

	if *allowedNamespaces.From == apis.NamespacesFromSame {
		return b.Namespace == srcNamespace.Name
	}

	return selectorMatches(allowedNamespaces.Selector, srcNamespace.Labels)
}

func selectorMatches(ls *metav1.LabelSelector, srcLabels map[string]string) bool {
	selector, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		klog.Infoln("invalid label selector: ", ls)
		return false
	}
	return selector.Matches(labels.Set(srcLabels))
}

func (b *BackupStorage) OffshootLabels() map[string]string {
	newLabels := make(map[string]string)
	newLabels[meta.ComponentLabelKey] = apis.KubeStashStorageComponent
	newLabels[meta.ManagedByLabelKey] = apis.KubeStashKey
	newLabels[apis.KubeStashInvokerName] = b.Name
	newLabels[apis.KubeStashInvokerNamespace] = b.Namespace
	return upsertLabels(b.Labels, newLabels)
}

func upsertLabels(oldLabels, newLabels map[string]string) map[string]string {
	if oldLabels == nil {
		oldLabels = make(map[string]string, len(newLabels))
	}
	for k, v := range newLabels {
		oldLabels[k] = v
	}
	return oldLabels
}
