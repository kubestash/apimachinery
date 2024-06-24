package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	kmc "kmodules.xyz/client-go/client"
	"kubestash.dev/apimachinery/apis"
	"kubestash.dev/apimachinery/apis/storage/v1alpha1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	staleSnapshotsFilePath = apis.TempDirMountPath + "/stale_snapshots_by_repo.json"
	pruneErrorsFilePath    = apis.TempDirMountPath + "/prune_errors_by_repo.json"
)

func ReadRepoStaleSnapshotListFromFile() (map[string][]kmapi.ObjectReference, error) {
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

func ReadRepoSnapshotPruneErrorsFromFile() (map[string][]string, error) {
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

func GetSnapshotsFromRefs(client client.Client, refs []kmapi.ObjectReference) ([]v1alpha1.Snapshot, error) {
	var snapshots []v1alpha1.Snapshot
	for _, ref := range refs {
		snapshot, err := getSnapshot(client, ref)
		if err != nil {
			return nil, fmt.Errorf("failed to get snapshot %s/%s: %w", ref.Namespace, ref.Name, err)
		}
		snapshots = append(snapshots, *snapshot)
	}
	return snapshots, nil
}

func SetSnapshotDeletionPolicyToWipeout(ctx context.Context, c client.Client, snap *v1alpha1.Snapshot) error {
	_, err := kmc.CreateOrPatch(
		ctx,
		c,
		snap,
		func(obj client.Object, createOp bool) client.Object {
			in := obj.(*v1alpha1.Snapshot)
			in.Spec.DeletionPolicy = v1alpha1.DeletionPolicyWipeOut
			return in
		},
	)
	return err
}

func WritePruneErrorToFile(repoName string, pruneErr error) error {
	if err := createFileIfNotExists(pruneErrorsFilePath); err != nil {
		return fmt.Errorf("failed to create %s: %w", pruneErrorsFilePath, err)
	}

	pruneErrors, err := ReadRepoSnapshotPruneErrorsFromFile()
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
	staleSnapshots, err := ReadRepoStaleSnapshotListFromFile()
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

func InitStaleSnapshotsFile(repoName string) error {
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

func getSnapshot(client client.Client, ref kmapi.ObjectReference) (*v1alpha1.Snapshot, error) {
	snapshot := &v1alpha1.Snapshot{}
	if err := client.Get(context.Background(), ref.ObjectKey(), snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}
