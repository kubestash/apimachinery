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

	"kmodules.xyz/client-go/apiextensions"
	cutil "kmodules.xyz/client-go/conditions"
	meta_util "kmodules.xyz/client-go/meta"
)

func (_ RestoreSession) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralRestoreSession))
}

func (rs *RestoreSession) CalculatePhase() RestorePhase {
	if cutil.IsConditionFalse(rs.Status.Conditions, TypeValidationPassed) {
		return RestoreInvalid
	}

	if cutil.IsConditionTrue(rs.Status.Conditions, TypeDeadlineExceeded) ||
		cutil.IsConditionFalse(rs.Status.Conditions, TypePreRestoreHooksExecutionSucceeded) ||
		cutil.IsConditionFalse(rs.Status.Conditions, TypePostRestoreHooksExecutionSucceeded) ||
		cutil.IsConditionFalse(rs.Status.Conditions, TypeRestoreExecutorEnsured) {
		return RestoreFailed
	}

	componentsPhase := rs.getComponentsPhase()
	if componentsPhase == RestorePending || componentsPhase == RestoreRunning {
		return componentsPhase
	}

	if rs.postHooksExecutionCompleted() {
		return componentsPhase
	}

	return RestoreRunning
}

func (rs *RestoreSession) AllComponentsCompleted() bool {
	phase := rs.getComponentsPhase()
	return phase == RestoreSucceeded || phase == RestoreFailed
}

func (rs *RestoreSession) postHooksExecutionCompleted() bool {
	hooks := rs.Spec.Hooks
	if hooks != nil && hooks.PostRestore != nil {
		return cutil.HasCondition(rs.Status.Conditions, TypePostRestoreHooksExecutionSucceeded)
	}
	return true
}

func (rs *RestoreSession) getComponentsPhase() RestorePhase {
	if len(rs.Status.Components) == 0 {
		return RestorePending
	}

	failedComponent := 0
	successfulComponent := 0

	for _, c := range rs.Status.Components {
		if c.Phase == RestoreSucceeded {
			successfulComponent++
		}
		if c.Phase == RestoreFailed {
			failedComponent++
		}
	}

	totalComponents := int(rs.Status.TotalComponents)

	if successfulComponent == totalComponents {
		return RestoreSucceeded
	}

	if successfulComponent+failedComponent == totalComponents {
		return RestoreFailed
	}

	return RestoreRunning
}

func (rs *RestoreSession) OffshootLabels() map[string]string {
	newLabels := make(map[string]string)
	newLabels[meta_util.ComponentLabelKey] = apis.KubeStashRestoreComponent
	newLabels[meta_util.ManagedByLabelKey] = apis.KubeStashKey
	newLabels[apis.KubeStashInvokerName] = rs.Name
	newLabels[apis.KubeStashInvokerNamespace] = rs.Namespace

	return apis.UpsertLabels(rs.Labels, newLabels)
}
