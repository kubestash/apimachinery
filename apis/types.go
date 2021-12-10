package apis

import core "k8s.io/api/core/v1"

type Driver string

const (
	DriverRestic Driver = "Restic"
	DriverWalG   Driver = "WalG"
)

type VolumeSource struct {
	core.VolumeSource
	VolumeClaimTemplate *core.PersistentVolumeClaimTemplate `json:"volumeClaimTemplate,omitempty"`
}
