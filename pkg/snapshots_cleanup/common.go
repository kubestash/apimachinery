package snapshotscleanup

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	kmc "kmodules.xyz/client-go/client"
	"kubestash.dev/apimachinery/apis"
	coreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
	"kubestash.dev/apimachinery/apis/storage/v1alpha1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	staleSnapshotsFilePath        = apis.TempDirMountPath + "/stale_snapshots_by_repo.json"
	retentionPolicyStatusFilePath = apis.TempDirMountPath + "/retention_policy_status_by_repo.json"
)

func AppendError(currentError string, newError error) string {
	if currentError != "" {
		return fmt.Sprintf("%s,%s", currentError, newError.Error())
	}
	return newError.Error()
}

func FetchStaleSnapshotsByRepo() (map[string][]kmapi.ObjectReference, error) {
	byteData, err := os.ReadFile(staleSnapshotsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", staleSnapshotsFilePath, err)
	}

	data := map[string][]kmapi.ObjectReference{}
	if len(byteData) > 0 {
		if err := json.Unmarshal(byteData, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %w", staleSnapshotsFilePath, err)
		}
	}

	return data, nil
}

func FetchPruneStatusByRepo() (map[string]coreapi.RetentionPolicyApplyStatus, error) {
	byteData, err := os.ReadFile(retentionPolicyStatusFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %w", retentionPolicyStatusFilePath, err)
	}

	pruneStatus := map[string]coreapi.RetentionPolicyApplyStatus{}
	if len(byteData) > 0 {
		if err := json.Unmarshal(byteData, &pruneStatus); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s file: %w", retentionPolicyStatusFilePath, err)
		}
	}

	return pruneStatus, nil
}

func GetSnapshotsFromRefs(client client.Client, refs []kmapi.ObjectReference) ([]v1alpha1.Snapshot, error) {
	var snapshots []v1alpha1.Snapshot
	for _, ref := range refs {
		snapshot, err := getSnapshot(client, ref)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch snapshot %s/%s: %w", ref.Namespace, ref.Name, err)
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

func UpdateRetentionPolicyStatus(repoName string, statusErr error) error {
	if err := createFileIfNotExists(retentionPolicyStatusFilePath); err != nil {
		return fmt.Errorf("failed to create %s: %w", retentionPolicyStatusFilePath, err)
	}

	pruneStatus, err := FetchPruneStatusByRepo()
	if err != nil {
		return err
	}

	status, exists := pruneStatus[repoName]
	if !exists {
		status = coreapi.RetentionPolicyApplyStatus{}
	}

	if statusErr != nil {
		status.Error = AppendError(status.Error, statusErr)
	}

	pruneStatus[repoName] = status

	jsonData, err := json.Marshal(pruneStatus)
	if err != nil {
		return fmt.Errorf("failed to marshal status data: %w", err)
	}

	if err = os.WriteFile(retentionPolicyStatusFilePath, jsonData, fs.ModePerm); err != nil {
		return fmt.Errorf("failed to write to JSON file: %w", err)
	}

	return nil
}

func WriteSnapshotsToStSaleSnapshotsFile(snapshots []v1alpha1.Snapshot) error {
	data, err := FetchStaleSnapshotsByRepo()
	if err != nil {
		return err
	}

	for _, snap := range snapshots {
		data[snap.Spec.Repository] = append(data[snap.Spec.Repository], kmapi.ObjectReference{
			Name:      snap.Name,
			Namespace: snap.Namespace,
		})
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	if err = os.WriteFile(staleSnapshotsFilePath, jsonData, fs.ModePerm); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}
	return nil
}

func InitStaleSnapshotsFile(repoName string) error {
	if err := createFileIfNotExists(staleSnapshotsFilePath); err != nil {
		return fmt.Errorf("failed to create %s: %w", staleSnapshotsFilePath, err)
	}

	byteData, err := os.ReadFile(staleSnapshotsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read %s file: %w", staleSnapshotsFilePath, err)
	}

	data := map[string][]kmapi.ObjectReference{}
	if len(byteData) > 0 {
		if err := json.Unmarshal(byteData, &data); err != nil {
			return fmt.Errorf("failed to unmarshal %s file: %w", staleSnapshotsFilePath, err)
		}
	}
	data[repoName] = []kmapi.ObjectReference{}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	if err = os.WriteFile(staleSnapshotsFilePath, jsonData, fs.ModePerm); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
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
