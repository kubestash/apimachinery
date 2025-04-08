package cloud

import (
	"context"
	"fmt"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AKSWorkloadIdentityLabel = "azure.workload.identity/use"
)

func GetCloudOIDCLabelsFromOperator(ctx context.Context, kc client.Client) (map[string]string, error) {
	sa, err := getServiceAccount(ctx, kc, kmapi.ObjectReference{
		Name:      meta.PodServiceAccount(),
		Namespace: meta.PodNamespace(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get service account: %w", err)
	}

	labels := make(map[string]string)
	setValueIfPresent(sa, AKSWorkloadIdentityLabel, labels)

	return labels, nil
}
