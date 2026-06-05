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

package progress

import (
	"context"
	"fmt"

	"kubestash.dev/apimachinery/apis"
	coreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	"github.com/docker/go-units"
	"gomodules.xyz/restic"
	kmc "kmodules.xyz/client-go/client"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (pg *Progress) setBackupProgress(repoName string, status *restic.ResticStatus) error {
	progress := &storageapi.BackupProgress{
		SecondsElapsed: status.SecondsElapsed,
		TotalFiles:     int64(status.TotalFiles),
		BackupDone:     units.HumanSize(float64(status.BytesDone)),
	}
	if status.TotalBytes > 0 {
		progress.Total = units.HumanSize(float64(status.TotalBytes))
	}
	if status.PercentDone*100 > 0 {
		progress.PercentDone = fmt.Sprintf("%.2f%%", status.PercentDone*100)
	}
	if status.SecondsElapsed > 0 {
		progress.Speed = units.HumanSize(float64(status.BytesDone)/float64(status.SecondsElapsed)) + "/s"
	}

	for idx := range pg.snapshots {
		snap := &pg.snapshots[idx]
		if snap.Spec.Repository != repoName {
			continue
		}
		if snap.Status.Components == nil {
			snap.Status.Components = make(map[string]storageapi.Component)
		}
		newComp := snap.Status.Components[pg.component]

		newComp.Driver = apis.DriverRestic
		if len(newComp.ResticStats) == 0 {
			newComp.ResticStats = []storageapi.ResticStats{{Progress: progress}}
		} else {
			newComp.ResticStats[0].Progress = progress
		}

		if err := pg.updateSnapshotComponentStatus(snap, newComp); err != nil {
			return fmt.Errorf("failed to update snapshot progress: %w", err)
		}
	}
	return nil
}

func (pg *Progress) setRestoreProgress(status *restic.ResticStatus) error {
	progress := &coreapi.RestoreProgress{
		SecondsElapsed: status.SecondsElapsed,
		TotalFiles:     int64(status.TotalFiles),
		RestoreDone:    units.HumanSize(float64(status.BytesRestored)),
	}
	if status.TotalBytes > 0 {
		progress.Total = units.HumanSize(float64(status.TotalBytes))
	}
	if status.PercentDone*100 > 0 {
		progress.PercentDone = fmt.Sprintf("%.2f%%", status.PercentDone*100)
	}
	if status.SecondsElapsed > 0 {
		progress.Speed = units.HumanSize(float64(status.BytesRestored)/float64(status.SecondsElapsed)) + "/s"
	}

	var comp coreapi.ComponentRestoreStatus
	if val, exist := pg.restoreSession.Status.Components[pg.component]; exist {
		comp = val
	} else {
		comp = coreapi.ComponentRestoreStatus{}
	}

	comp.Progress = progress
	if err := pg.updateRestoreSessionComponentStatus(comp); err != nil {
		return fmt.Errorf("failed to update restore progress: %w", err)
	}
	return nil
}

func (pg *Progress) updateSnapshotComponentStatus(snap *storageapi.Snapshot, comp storageapi.Component) error {
	_, err := kmc.PatchStatus(
		context.Background(),
		pg.kbClient,
		snap,
		func(obj client.Object) client.Object {
			in := obj.(*storageapi.Snapshot)
			if in.Status.Components == nil {
				in.Status.Components = make(map[string]storageapi.Component)
			}
			in.Status.Components[pg.component] = comp
			return in
		})
	return err
}

func (pg *Progress) updateRestoreSessionComponentStatus(comp coreapi.ComponentRestoreStatus) error {
	_, err := kmc.PatchStatus(
		context.Background(),
		pg.kbClient,
		pg.restoreSession,
		func(obj client.Object) client.Object {
			in := obj.(*coreapi.RestoreSession)
			if in.Status.Components == nil {
				in.Status.Components = make(map[string]coreapi.ComponentRestoreStatus)
			}
			in.Status.Components[pg.component] = comp
			return in
		},
	)
	return err
}
