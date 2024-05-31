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
	staleSnapshotsFile        = apis.TempDirMountPath + "/" + "stale_snapshots_by_repo.json"
	retentionPolicyStatusFile = apis.TempDirMountPath + "/" + "retention_policy_status_by_repo.json"
)

func ensureFileExists(filename string) error {
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
	}
	return nil
}

func appendError(currentError string, newError error) string {
	if currentError != "" {
		return fmt.Sprintf("%s,%s", currentError, newError.Error())
	}
	return newError.Error()
}

func getSnapshot(client client.Client, ref kmapi.ObjectReference) (*v1alpha1.Snapshot, error) {
	snapshot := &v1alpha1.Snapshot{}
	if err := client.Get(context.Background(), ref.ObjectKey(), snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
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

func LoadStaleSnapshotsByRepo() (map[string][]kmapi.ObjectReference, error) {
	byteData, err := os.ReadFile(staleSnapshotsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", staleSnapshotsFile, err)
	}

	data := map[string][]kmapi.ObjectReference{}
	if len(byteData) > 0 {
		if err := json.Unmarshal(byteData, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %w", staleSnapshotsFile, err)
		}
	}

	return data, nil
}

func UpdateRetentionPolicyStatus(repoName string, statusErr error) error {
	if err := ensureFileExists(retentionPolicyStatusFile); err != nil {
		return fmt.Errorf("failed to create %s: %w", retentionPolicyStatusFile, err)
	}

	byteData, err := os.ReadFile(retentionPolicyStatusFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", retentionPolicyStatusFile, err)
	}

	statusMap := map[string]coreapi.RetentionPolicyApplyStatus{}
	status := coreapi.RetentionPolicyApplyStatus{}
	if len(byteData) > 0 {
		if err := json.Unmarshal(byteData, &statusMap); err != nil {
			return fmt.Errorf("failed to unmarshal %s: %w", retentionPolicyStatusFile, err)
		}
		status = statusMap[repoName]
	}

	if statusErr != nil {
		status.Error = appendError(status.Error, statusErr)
	}

	statusMap[repoName] = status
	jsonData, err := json.Marshal(statusMap)
	if err != nil {
		return fmt.Errorf("failed to marshal status data: %w", err)
	}

	if err = os.WriteFile(retentionPolicyStatusFile, jsonData, fs.ModePerm); err != nil {
		return fmt.Errorf("failed to write to JSON file: %w", err)
	}

	return nil
}
