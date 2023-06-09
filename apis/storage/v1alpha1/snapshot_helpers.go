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
	"fmt"
	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/crds"
	"regexp"
	"strconv"
	"strings"

	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	"kmodules.xyz/client-go/meta"
)

func (_ Snapshot) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralSnapshot))
}

func (s *Snapshot) CalculatePhase() SnapshotPhase {
	if kmapi.IsConditionFalse(s.Status.Conditions, TypeSnapshotMetadataUploaded) ||
		kmapi.IsConditionFalse(s.Status.Conditions, TypeRecentSnapshotListUpdated) {
		return SnapshotFailed
	}

	return s.GetComponentsPhase()
}

func (s *Snapshot) GetComponentsPhase() SnapshotPhase {
	if len(s.Status.Components) == 0 {
		return SnapshotPending
	}

	failedComponent := 0
	successfulComponent := 0

	for _, c := range s.Status.Components {
		if c.Phase == ComponentPhaseSucceeded {
			successfulComponent++
		}
		if c.Phase == ComponentPhaseFailed {
			failedComponent++
		}
	}

	totalComponents := int(s.Status.TotalComponents)

	if successfulComponent == totalComponents {
		return SnapshotSucceeded
	}

	if successfulComponent+failedComponent == totalComponents {
		return SnapshotFailed
	}

	return SnapshotRunning
}

func (s *Snapshot) IsCompleted() bool {
	return s.Status.Phase == SnapshotSucceeded || s.Status.Phase == SnapshotFailed
}

func (s *Snapshot) GetIntegrity() *bool {
	if s.Status.Components == nil {
		return nil
	}

	result := true
	for _, component := range s.Status.Components {
		if component.Integrity == nil {
			return nil
		}
		result = result && *component.Integrity
	}
	return &result
}

func (s *Snapshot) GetSize() string {
	if s.Status.Components == nil {
		return ""
	}

	var totalSizeInByte uint64
	for _, component := range s.Status.Components {
		if component.Size == "" {
			return ""
		}

		sizeWithUnit := strings.Split(component.Size, " ")
		if len(sizeWithUnit) < 2 {
			return ""
		}

		sizeInByte, err := convertSizeToByte(sizeWithUnit)
		if err != nil {
			return ""
		}
		totalSizeInByte += sizeInByte
	}
	return formatBytes(totalSizeInByte)
}

func convertSizeToByte(sizeWithUnit []string) (uint64, error) {
	numeral, err := strconv.ParseFloat(sizeWithUnit[0], 64)
	if err != nil {
		return 0, err
	}

	switch sizeWithUnit[1] {
	case "TiB":
		return uint64(numeral * (1 << 40)), nil
	case "GiB":
		return uint64(numeral * (1 << 30)), nil
	case "MiB":
		return uint64(numeral * (1 << 20)), nil
	case "KiB":
		return uint64(numeral * (1 << 10)), nil
	case "B":
		return uint64(numeral), nil
	default:
		return 0, fmt.Errorf("no valid unit matched")
	}
}

func formatBytes(c uint64) string {
	b := float64(c)
	switch {
	case c > 1<<40:
		return fmt.Sprintf("%.3f TiB", b/(1<<40))
	case c > 1<<30:
		return fmt.Sprintf("%.3f GiB", b/(1<<30))
	case c > 1<<20:
		return fmt.Sprintf("%.3f MiB", b/(1<<20))
	case c > 1<<10:
		return fmt.Sprintf("%.3f KiB", b/(1<<10))
	default:
		return fmt.Sprintf("%d B", c)
	}
}

func GenerateSnapshotName(repoName, backupSession string) string {
	backupSessionRegex := regexp.MustCompile("(.*)-([0-9]+)$")
	subMatches := backupSessionRegex.FindStringSubmatch(backupSession)
	return meta.ValidNameWithPrefixNSuffix(repoName, subMatches[1], subMatches[2])
}

func (s *Snapshot) OffshootLabels() map[string]string {
	newLabels := make(map[string]string)
	newLabels[meta.ManagedByLabelKey] = apis.KubeStashKey
	newLabels[apis.KubeStashInvokerName] = s.Name
	newLabels[apis.KubeStashInvokerNamespace] = s.Namespace
	return apis.UpsertLabels(s.Labels, newLabels)
}
