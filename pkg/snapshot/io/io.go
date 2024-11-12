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

package io

import (
	"fmt"
	"io/fs"
	"os"

	"encoding/json"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/apis/storage/v1alpha1"
)

const (
	staleSnapshotsFilePath = apis.TempDirMountPath + "/stale_snapshots_by_repo.json"
	pruneErrorsFilePath    = apis.TempDirMountPath + "/prune_errors_by_repo.json"
)

func InitStaleSnapshotListFile(repoName string) error {
	if err := createFileIfNotExists(staleSnapshotsFilePath); err != nil {
		return fmt.Errorf("failed to create %s file: %w", staleSnapshotsFilePath, err)
	}

	bytes, err := os.ReadFile(staleSnapshotsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read %s file: %w", staleSnapshotsFilePath, err)
	}

	staleSnapshots := map[string][]kmapi.ObjectReference{}
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, &staleSnapshots); err != nil {
			return fmt.Errorf("failed to unmarshal %s file: %w", staleSnapshotsFilePath, err)
		}
	}
	staleSnapshots[repoName] = []kmapi.ObjectReference{}

	jsonData, err := json.Marshal(staleSnapshots)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot list data: %w", err)
	}

	if err = os.WriteFile(staleSnapshotsFilePath, jsonData, fs.ModePerm); err != nil {
		return fmt.Errorf("failed to write to %s file: %w", staleSnapshotsFilePath, err)
	}
	return nil
}

func WritePruneErrorToFile(repoName string, pruneErr error) error {
	if err := createFileIfNotExists(pruneErrorsFilePath); err != nil {
		return fmt.Errorf("failed to create %s: %w", pruneErrorsFilePath, err)
	}

	pruneErrors, err := ReadPruneErrorsFromFile()
	if err != nil {
		return err
	}

	if pruneErr != nil {
		pruneErrors[repoName] = append(pruneErrors[repoName], pruneErr.Error())
	}

	jsonData, err := json.Marshal(pruneErrors)
	if err != nil {
		return fmt.Errorf("failed to marshal prune errors data: %w", err)
	}

	if err = os.WriteFile(pruneErrorsFilePath, jsonData, fs.ModePerm); err != nil {
		return fmt.Errorf("failed to write to %s file: %w", pruneErrorsFilePath, err)
	}

	return nil
}

func AddSnapshotListToFile(snapshots []v1alpha1.Snapshot) error {
	staleSnapshots, err := ReadStaleSnapshotListFromFile()
	if err != nil {
		return err
	}

	for _, snap := range snapshots {
		staleSnapshots[snap.Spec.Repository] = append(staleSnapshots[snap.Spec.Repository], kmapi.ObjectReference{
			Name:      snap.Name,
			Namespace: snap.Namespace,
		})
	}

	jsonData, err := json.Marshal(staleSnapshots)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot list data: %w", err)
	}

	if err = os.WriteFile(staleSnapshotsFilePath, jsonData, fs.ModePerm); err != nil {
		return fmt.Errorf("failed to write to %s file: %w", staleSnapshotsFilePath, err)
	}
	return nil
}

func ReadStaleSnapshotListFromFile() (map[string][]kmapi.ObjectReference, error) {
	bytes, err := os.ReadFile(staleSnapshotsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %w", staleSnapshotsFilePath, err)
	}

	staleSnapshots := map[string][]kmapi.ObjectReference{}
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, &staleSnapshots); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s file: %w", staleSnapshotsFilePath, err)
		}
	}

	return staleSnapshots, nil
}

func ReadPruneErrorsFromFile() (map[string][]string, error) {
	bytes, err := os.ReadFile(pruneErrorsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %w", pruneErrorsFilePath, err)
	}

	pruneErrors := map[string][]string{}
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, &pruneErrors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s file: %w", pruneErrorsFilePath, err)
		}
	}

	return pruneErrors, nil
}

func createFileIfNotExists(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				klog.Infoln("failed to close file:", err)
			}
		}(file)
		return nil
	}
	return err
}
