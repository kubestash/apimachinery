//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kmodules.xyz/client-go/api/v1"
	"kubestash.dev/apimachinery/apis"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureSpec) DeepCopyInto(out *AzureSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureSpec.
func (in *AzureSpec) DeepCopy() *AzureSpec {
	if in == nil {
		return nil
	}
	out := new(AzureSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *B2Spec) DeepCopyInto(out *B2Spec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new B2Spec.
func (in *B2Spec) DeepCopy() *B2Spec {
	if in == nil {
		return nil
	}
	out := new(B2Spec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Backend) DeepCopyInto(out *Backend) {
	*out = *in
	if in.Local != nil {
		in, out := &in.Local, &out.Local
		*out = new(LocalSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.S3 != nil {
		in, out := &in.S3, &out.S3
		*out = new(S3Spec)
		**out = **in
	}
	if in.GCS != nil {
		in, out := &in.GCS, &out.GCS
		*out = new(GCSSpec)
		**out = **in
	}
	if in.Azure != nil {
		in, out := &in.Azure, &out.Azure
		*out = new(AzureSpec)
		**out = **in
	}
	if in.Swift != nil {
		in, out := &in.Swift, &out.Swift
		*out = new(SwiftSpec)
		**out = **in
	}
	if in.B2 != nil {
		in, out := &in.B2, &out.B2
		*out = new(B2Spec)
		**out = **in
	}
	if in.Rest != nil {
		in, out := &in.Rest, &out.Rest
		*out = new(RestServerSpec)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Backend.
func (in *Backend) DeepCopy() *Backend {
	if in == nil {
		return nil
	}
	out := new(Backend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupStorage) DeepCopyInto(out *BackupStorage) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupStorage.
func (in *BackupStorage) DeepCopy() *BackupStorage {
	if in == nil {
		return nil
	}
	out := new(BackupStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BackupStorage) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupStorageList) DeepCopyInto(out *BackupStorageList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]BackupStorage, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupStorageList.
func (in *BackupStorageList) DeepCopy() *BackupStorageList {
	if in == nil {
		return nil
	}
	out := new(BackupStorageList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BackupStorageList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupStorageSpec) DeepCopyInto(out *BackupStorageSpec) {
	*out = *in
	in.Storage.DeepCopyInto(&out.Storage)
	if in.UsagePolicy != nil {
		in, out := &in.UsagePolicy, &out.UsagePolicy
		*out = new(apis.UsagePolicy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupStorageSpec.
func (in *BackupStorageSpec) DeepCopy() *BackupStorageSpec {
	if in == nil {
		return nil
	}
	out := new(BackupStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupStorageStatus) DeepCopyInto(out *BackupStorageStatus) {
	*out = *in
	if in.Repositories != nil {
		in, out := &in.Repositories, &out.Repositories
		*out = make([]RepositoryInfo, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupStorageStatus.
func (in *BackupStorageStatus) DeepCopy() *BackupStorageStatus {
	if in == nil {
		return nil
	}
	out := new(BackupStorageStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Component) DeepCopyInto(out *Component) {
	*out = *in
	if in.Integrity != nil {
		in, out := &in.Integrity, &out.Integrity
		*out = new(bool)
		**out = **in
	}
	if in.ResticStats != nil {
		in, out := &in.ResticStats, &out.ResticStats
		*out = make([]ResticStats, len(*in))
		copy(*out, *in)
	}
	if in.VolumeSnapshotterStats != nil {
		in, out := &in.VolumeSnapshotterStats, &out.VolumeSnapshotterStats
		*out = new(VolumeSnapshotterStats)
		**out = **in
	}
	if in.WalSegments != nil {
		in, out := &in.WalSegments, &out.WalSegments
		*out = make([]WalSegment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Component.
func (in *Component) DeepCopy() *Component {
	if in == nil {
		return nil
	}
	out := new(Component)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Duration) DeepCopyInto(out *Duration) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Duration.
func (in *Duration) DeepCopy() *Duration {
	if in == nil {
		return nil
	}
	out := new(Duration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FailedSnapshotsKeepPolicy) DeepCopyInto(out *FailedSnapshotsKeepPolicy) {
	*out = *in
	if in.Last != nil {
		in, out := &in.Last, &out.Last
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FailedSnapshotsKeepPolicy.
func (in *FailedSnapshotsKeepPolicy) DeepCopy() *FailedSnapshotsKeepPolicy {
	if in == nil {
		return nil
	}
	out := new(FailedSnapshotsKeepPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCSSpec) DeepCopyInto(out *GCSSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCSSpec.
func (in *GCSSpec) DeepCopy() *GCSSpec {
	if in == nil {
		return nil
	}
	out := new(GCSSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalSpec) DeepCopyInto(out *LocalSpec) {
	*out = *in
	in.VolumeSource.DeepCopyInto(&out.VolumeSource)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalSpec.
func (in *LocalSpec) DeepCopy() *LocalSpec {
	if in == nil {
		return nil
	}
	out := new(LocalSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Repository) DeepCopyInto(out *Repository) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Repository.
func (in *Repository) DeepCopy() *Repository {
	if in == nil {
		return nil
	}
	out := new(Repository)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Repository) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositoryInfo) DeepCopyInto(out *RepositoryInfo) {
	*out = *in
	if in.Synced != nil {
		in, out := &in.Synced, &out.Synced
		*out = new(bool)
		**out = **in
	}
	if in.Error != nil {
		in, out := &in.Error, &out.Error
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositoryInfo.
func (in *RepositoryInfo) DeepCopy() *RepositoryInfo {
	if in == nil {
		return nil
	}
	out := new(RepositoryInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositoryList) DeepCopyInto(out *RepositoryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Repository, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositoryList.
func (in *RepositoryList) DeepCopy() *RepositoryList {
	if in == nil {
		return nil
	}
	out := new(RepositoryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RepositoryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositorySpec) DeepCopyInto(out *RepositorySpec) {
	*out = *in
	out.AppRef = in.AppRef
	out.StorageRef = in.StorageRef
	if in.EncryptionSecret != nil {
		in, out := &in.EncryptionSecret, &out.EncryptionSecret
		*out = new(v1.ObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositorySpec.
func (in *RepositorySpec) DeepCopy() *RepositorySpec {
	if in == nil {
		return nil
	}
	out := new(RepositorySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositoryStatus) DeepCopyInto(out *RepositoryStatus) {
	*out = *in
	if in.Integrity != nil {
		in, out := &in.Integrity, &out.Integrity
		*out = new(bool)
		**out = **in
	}
	if in.SnapshotCount != nil {
		in, out := &in.SnapshotCount, &out.SnapshotCount
		*out = new(int32)
		**out = **in
	}
	if in.RecentSnapshots != nil {
		in, out := &in.RecentSnapshots, &out.RecentSnapshots
		*out = make([]SnapshotInfo, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositoryStatus.
func (in *RepositoryStatus) DeepCopy() *RepositoryStatus {
	if in == nil {
		return nil
	}
	out := new(RepositoryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestServerSpec) DeepCopyInto(out *RestServerSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestServerSpec.
func (in *RestServerSpec) DeepCopy() *RestServerSpec {
	if in == nil {
		return nil
	}
	out := new(RestServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResticStats) DeepCopyInto(out *ResticStats) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResticStats.
func (in *ResticStats) DeepCopy() *ResticStats {
	if in == nil {
		return nil
	}
	out := new(ResticStats)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RetentionPolicy) DeepCopyInto(out *RetentionPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RetentionPolicy.
func (in *RetentionPolicy) DeepCopy() *RetentionPolicy {
	if in == nil {
		return nil
	}
	out := new(RetentionPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RetentionPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RetentionPolicyList) DeepCopyInto(out *RetentionPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RetentionPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RetentionPolicyList.
func (in *RetentionPolicyList) DeepCopy() *RetentionPolicyList {
	if in == nil {
		return nil
	}
	out := new(RetentionPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RetentionPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RetentionPolicySpec) DeepCopyInto(out *RetentionPolicySpec) {
	*out = *in
	if in.UsagePolicy != nil {
		in, out := &in.UsagePolicy, &out.UsagePolicy
		*out = new(apis.UsagePolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.SuccessfulSnapshots != nil {
		in, out := &in.SuccessfulSnapshots, &out.SuccessfulSnapshots
		*out = new(SuccessfulSnapshotsKeepPolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.FailedSnapshots != nil {
		in, out := &in.FailedSnapshots, &out.FailedSnapshots
		*out = new(FailedSnapshotsKeepPolicy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RetentionPolicySpec.
func (in *RetentionPolicySpec) DeepCopy() *RetentionPolicySpec {
	if in == nil {
		return nil
	}
	out := new(RetentionPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *S3Spec) DeepCopyInto(out *S3Spec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new S3Spec.
func (in *S3Spec) DeepCopy() *S3Spec {
	if in == nil {
		return nil
	}
	out := new(S3Spec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Snapshot) DeepCopyInto(out *Snapshot) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Snapshot.
func (in *Snapshot) DeepCopy() *Snapshot {
	if in == nil {
		return nil
	}
	out := new(Snapshot)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Snapshot) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SnapshotInfo) DeepCopyInto(out *SnapshotInfo) {
	*out = *in
	if in.SnapshotTime != nil {
		in, out := &in.SnapshotTime, &out.SnapshotTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SnapshotInfo.
func (in *SnapshotInfo) DeepCopy() *SnapshotInfo {
	if in == nil {
		return nil
	}
	out := new(SnapshotInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SnapshotList) DeepCopyInto(out *SnapshotList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Snapshot, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SnapshotList.
func (in *SnapshotList) DeepCopy() *SnapshotList {
	if in == nil {
		return nil
	}
	out := new(SnapshotList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SnapshotList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SnapshotSpec) DeepCopyInto(out *SnapshotSpec) {
	*out = *in
	out.AppRef = in.AppRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SnapshotSpec.
func (in *SnapshotSpec) DeepCopy() *SnapshotSpec {
	if in == nil {
		return nil
	}
	out := new(SnapshotSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SnapshotStatus) DeepCopyInto(out *SnapshotStatus) {
	*out = *in
	if in.SnapshotTime != nil {
		in, out := &in.SnapshotTime, &out.SnapshotTime
		*out = (*in).DeepCopy()
	}
	if in.LastUpdateTime != nil {
		in, out := &in.LastUpdateTime, &out.LastUpdateTime
		*out = (*in).DeepCopy()
	}
	if in.Integrity != nil {
		in, out := &in.Integrity, &out.Integrity
		*out = new(bool)
		**out = **in
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Tasks != nil {
		in, out := &in.Tasks, &out.Tasks
		*out = make(map[string]Task, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SnapshotStatus.
func (in *SnapshotStatus) DeepCopy() *SnapshotStatus {
	if in == nil {
		return nil
	}
	out := new(SnapshotStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SuccessfulSnapshotsKeepPolicy) DeepCopyInto(out *SuccessfulSnapshotsKeepPolicy) {
	*out = *in
	if in.Last != nil {
		in, out := &in.Last, &out.Last
		*out = new(int32)
		**out = **in
	}
	if in.Hourly != nil {
		in, out := &in.Hourly, &out.Hourly
		*out = new(int32)
		**out = **in
	}
	if in.Daily != nil {
		in, out := &in.Daily, &out.Daily
		*out = new(int32)
		**out = **in
	}
	if in.Weekly != nil {
		in, out := &in.Weekly, &out.Weekly
		*out = new(int32)
		**out = **in
	}
	if in.Monthly != nil {
		in, out := &in.Monthly, &out.Monthly
		*out = new(int32)
		**out = **in
	}
	if in.Yearly != nil {
		in, out := &in.Yearly, &out.Yearly
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SuccessfulSnapshotsKeepPolicy.
func (in *SuccessfulSnapshotsKeepPolicy) DeepCopy() *SuccessfulSnapshotsKeepPolicy {
	if in == nil {
		return nil
	}
	out := new(SuccessfulSnapshotsKeepPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SwiftSpec) DeepCopyInto(out *SwiftSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SwiftSpec.
func (in *SwiftSpec) DeepCopy() *SwiftSpec {
	if in == nil {
		return nil
	}
	out := new(SwiftSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Task) DeepCopyInto(out *Task) {
	*out = *in
	if in.Components != nil {
		in, out := &in.Components, &out.Components
		*out = make(map[string]Component, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Task.
func (in *Task) DeepCopy() *Task {
	if in == nil {
		return nil
	}
	out := new(Task)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotterStats) DeepCopyInto(out *VolumeSnapshotterStats) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotterStats.
func (in *VolumeSnapshotterStats) DeepCopy() *VolumeSnapshotterStats {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotterStats)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WalSegment) DeepCopyInto(out *WalSegment) {
	*out = *in
	if in.Start != nil {
		in, out := &in.Start, &out.Start
		*out = (*in).DeepCopy()
	}
	if in.End != nil {
		in, out := &in.End, &out.End
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WalSegment.
func (in *WalSegment) DeepCopy() *WalSegment {
	if in == nil {
		return nil
	}
	out := new(WalSegment)
	in.DeepCopyInto(out)
	return out
}
