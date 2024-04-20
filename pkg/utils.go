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

package pkg

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"gomodules.xyz/envsubst"
	core "k8s.io/api/core/v1"
	"kubestash.dev/apimachinery/apis"
	"sort"

	"encoding/json"
	vsapi "github.com/kubernetes-csi/external-snapshotter/client/v7/apis/volumesnapshot/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	cu "kmodules.xyz/client-go/client"
	addonapi "kubestash.dev/apimachinery/apis/addons/v1alpha1"
	coreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewUncachedClient() (client.Client, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config. Reason: %w", err)
	}

	return cu.NewUncachedClient(
		cfg,
		clientsetscheme.AddToScheme,
		storageapi.AddToScheme,
		coreapi.AddToScheme,
		addonapi.AddToScheme,
		vsapi.AddToScheme,
	)
}

func GetTmpVolumeAndMount() (core.Volume, core.VolumeMount) {
	vol := core.Volume{
		Name: apis.TempDirVolumeName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	mnt := core.VolumeMount{
		Name:      apis.TempDirVolumeName,
		MountPath: apis.TempDirMountPath,
	}

	return vol, mnt
}

func ResolveWithInputs(obj interface{}, inputs map[string]string) error {
	// convert to JSON, apply replacements and convert back to struct
	jsonObj, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	resolved, err := envsubst.EvalMap(string(jsonObj), inputs)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(resolved), obj)
}

func FindAppropriateAddonVersion(addonVersions []string, dbVersion string) (string, error) {
	if addonVersions == nil {
		return "", fmt.Errorf("available list of addon-versions can't be empty")
	}
	semverDBVersion, err := semver.NewVersion(dbVersion)
	if err != nil {
		return "", err
	}

	type distance struct {
		major, minor, patch int64
		addon               string
	}
	abs := func(x, y uint64) int64 {
		tmp := int64(x) - int64(y)
		if tmp < 0 {
			tmp = -tmp
		}
		return tmp
	}
	distances := make([]distance, 0)
	for _, av := range addonVersions {
		sav, err := semver.NewVersion(av)
		if err != nil {
			return "", err
		}
		distances = append(distances, distance{
			major: abs(sav.Major(), semverDBVersion.Major()),
			minor: abs(sav.Minor(), semverDBVersion.Minor()),
			patch: abs(sav.Patch(), semverDBVersion.Patch()),
			addon: av,
		})
	}
	sort.Slice(distances, func(i, j int) bool {
		if distances[i].major == distances[j].major {
			if distances[i].minor == distances[j].minor {
				return distances[i].patch < distances[j].patch
			}
			return distances[i].minor < distances[j].minor
		}
		return distances[i].major < distances[j].major
	})
	return distances[0].addon, nil
}
