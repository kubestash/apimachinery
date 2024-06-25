package snapshot

import (
	"context"
	"fmt"
	kmapi "kmodules.xyz/client-go/api/v1"
	kmc "kmodules.xyz/client-go/client"
	"kubestash.dev/apimachinery/apis/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

func getSnapshot(client client.Client, ref kmapi.ObjectReference) (*v1alpha1.Snapshot, error) {
	snapshot := &v1alpha1.Snapshot{}
	if err := client.Get(context.Background(), ref.ObjectKey(), snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}
