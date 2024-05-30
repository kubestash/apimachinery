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
	corev1 "k8s.io/api/core/v1"
	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (_ BackupVerification) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralBackupVerification))
}

func (b *BackupVerification) UsageAllowed(srcNamespace *corev1.Namespace) bool {
	allowedNamespace := b.Spec.UsagePolicy.AllowedNamespaces
	if *allowedNamespace.From == apis.NamespacesFromAll {
		return true
	}

	if *allowedNamespace.From == apis.NamespacesFromSame {
		return b.Namespace == srcNamespace.Name
	}

	return selectorMatches(allowedNamespace.Selector, srcNamespace.Labels)
}
