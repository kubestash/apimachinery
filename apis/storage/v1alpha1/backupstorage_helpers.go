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
	"time"

	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/crds"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/apiextensions"
	cutil "kmodules.xyz/client-go/conditions"
	"kmodules.xyz/client-go/meta"
)

func (*BackupStorage) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralBackupStorage))
}

func (b *BackupStorage) CalculatePhase() BackupStoragePhase {
	if b.ObjectLockEnforced() && !cutil.IsConditionTrue(b.Status.Conditions, TypeObjectLockVerified) {
		return BackupStorageNotReady
	}
	if cutil.IsConditionTrue(b.Status.Conditions, TypeBackendInitialized) {
		if !cutil.HasCondition(b.Status.Conditions, TypeBackendSecretFound) {
			return BackupStorageReady
		}
		if cutil.IsConditionTrue(b.Status.Conditions, TypeBackendSecretFound) {
			return BackupStorageReady
		}
	}
	return BackupStorageNotReady
}

// ObjectLockEnforced returns true if the user requires the bucket to have Object Lock enabled.
func (b *BackupStorage) ObjectLockEnforced() bool {
	return b.Spec.ObjectLock != nil && (b.Spec.ObjectLock.Required || b.Spec.ObjectLock.Mode != "")
}

// ObjectLockEnabled returns true if Object Lock was found enabled on the bucket.
func (b *BackupStorage) ObjectLockEnabled() bool {
	return b.Status.ObjectLock != nil && b.Status.ObjectLock.Enabled
}

// ObjectLockRetentionPeriod returns how long the objects still used by the backups should stay locked.
// It returns 0 if Object Lock is not enabled on the bucket or the retention extension is disabled.
func (b *BackupStorage) ObjectLockRetentionPeriod() time.Duration {
	if !b.ObjectLockEnabled() {
		return 0
	}
	if b.Spec.ObjectLock != nil {
		if b.Spec.ObjectLock.DisableRetentionExtension {
			return 0
		}
		if b.Spec.ObjectLock.RetentionPeriod != nil {
			return b.Spec.ObjectLock.RetentionPeriod.Duration
		}
	}
	if b.Status.ObjectLock.DefaultRetention != nil {
		return b.Status.ObjectLock.DefaultRetention.Duration
	}
	return 0
}

// ObjectLockMode returns the retention mode to use when extending the retention of an object.
func (b *BackupStorage) ObjectLockMode() ObjectLockMode {
	if b.Spec.ObjectLock != nil && b.Spec.ObjectLock.Mode != "" {
		return b.Spec.ObjectLock.Mode
	}
	if b.Status.ObjectLock != nil {
		return b.Status.ObjectLock.Mode
	}
	return ""
}

func (b *BackupStorage) UsageAllowed(srcNamespace *core.Namespace) bool {
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
	newLabels[meta.ManagedByLabelKey] = apis.KubeStashKey
	newLabels[apis.KubeStashInvokerKind] = ResourceKindBackupStorage
	newLabels[apis.KubeStashInvokerName] = b.Name
	newLabels[apis.KubeStashInvokerNamespace] = b.Namespace
	return apis.UpsertLabels(b.Labels, newLabels)
}

func (b *BackupStorage) LocalProvider() bool {
	return b.Spec.Storage.Provider == ProviderLocal
}

func (b *BackupStorage) LocalNetworkVolume() bool {
	if b.Spec.Storage.Provider == ProviderLocal &&
		b.Spec.Storage.Local.NFS != nil {
		return true
	}
	return false
}

// LocalBackendPVC returns true if the local backend is backed by a PersistentVolumeClaim.
// Like network volumes, a PVC can be mounted by the backend accessor deployment,
// so snapshot download and repository maintenance commands can be served through it.
func (b *BackupStorage) LocalBackendPVC() bool {
	if b.Spec.Storage.Provider == ProviderLocal &&
		b.Spec.Storage.Local.PersistentVolumeClaim != nil {
		return true
	}
	return false
}

func (b *BackupStorage) IsCredentialLessModeEnabled() bool {
	switch b.Spec.Storage.Provider {
	case ProviderS3:
		return b.Spec.Storage.S3.SecretName == ""
	case ProviderAzure:
		return b.Spec.Storage.Azure.SecretName == ""
	default:
		return false
	}
}
