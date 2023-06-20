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

package apis

import "time"

const (
	RequeueTimeInterval = 10 * time.Second

	KubeStashCleanupFinalizer = "kubestash.com/cleanup"
	KubeStashKey              = "kubestash.com"
	KubeDBGroupName           = "kubedb.com"
)

const (
	OwnerKey = ".metadata.controller"
)

const (
	KubeStashBackupComponent  = "kubestash-backup"
	KubeStashRestoreComponent = "kubestash-restore"
)

// Keys for offshoot labels
const (
	KubeStashInvokerNamespace = "kubestash.com/invoker-ns"
)

// Keys for structure logging
const (
	KeyTargetKind      = "target_kind"
	KeyTargetName      = "target_name"
	KeyTargetNamespace = "target_namespace"
	KeyReason          = "reason"
	KeyName            = "name"
)

// Keys for BackupBlueprint
const (
	VariablesKey       = "variables.kubestash.com"
	BackupBlueprintKey = "blueprint.kubestash.com"

	KeyBlueprintName      = BackupBlueprintKey + "/name"
	KeyBlueprintNamespace = BackupBlueprintKey + "/namespace"
	KeyBlueprintSessions  = BackupBlueprintKey + "/sessions"
)

// RBAC related constants
const (
	KindClusterRole = "ClusterRole"

	KubeStashBackupJobClusterRole  = "kubestash-backup-job"
	KubeStashRestoreJobClusterRole = "kubestash-restore-job"
	KubeStashCronJobClusterRole    = "kubestash-cron-job"
)

// Reconciliation related constants
const (
	Requeue      = true
	DoNotRequeue = false
)

// Workload related constants
const (
	EnvComponentName    = "COMPONENT_NAME"
	ComponentPod        = "pod"
	ComponentDeployment = "deployment"

	KindStatefulSet = "StatefulSet"
	KindDaemonSet   = "DaemonSet"
	KindDeployment  = "Deployment"
)
