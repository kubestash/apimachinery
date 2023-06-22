package v1alpha1

import core "k8s.io/api/core/v1"

// ToVolumeAndMount returns volumes and mounts for local backend
func (l LocalSpec) ToVolumeAndMount(volName string) (core.Volume, core.VolumeMount) {
	vol := core.Volume{
		Name:         volName,
		VolumeSource: *l.VolumeSource.ToAPIObject(),
	}
	mnt := core.VolumeMount{
		Name:      volName,
		MountPath: l.MountPath,
		SubPath:   l.SubPath,
	}
	return vol, mnt
}
