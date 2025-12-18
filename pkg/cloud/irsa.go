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

package cloud

import (
	"context"
	"fmt"

	core "k8s.io/api/core/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AWSIRSARoleAnnotationKey              = "eks.amazonaws.com/role-arn"
	PodIdentityAssociationIDAnnotationKey = "klusters.dev/pod-identity-association-id"
	GCPWorkloadIdentityAnnotationKey      = "klusters.dev/iam.gke.io/workloadIdentity"
)

func GetCloudAnnotationsFromOperator(ctx context.Context, kc client.Client) (map[string]string, error) {
	annotations, err := GetCloudAnnotationsFromServiceAccount(ctx, kc)
	if err != nil {
		return nil, err
	}
	return annotations, nil
}

func GetCloudAnnotationsFromServiceAccount(ctx context.Context, kc client.Client) (map[string]string, error) {
	sa, err := getServiceAccount(ctx, kc, kmapi.ObjectReference{
		Name:      meta.PodServiceAccount(),
		Namespace: meta.PodNamespace(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve service account: %w", err)
	}
	annotations := map[string]string{}
	annotations[AWSIRSARoleAnnotationKey] = sa.Annotations[AWSIRSARoleAnnotationKey]
	annotations[PodIdentityAssociationIDAnnotationKey] = sa.Annotations[PodIdentityAssociationIDAnnotationKey]
	annotations[GCPWorkloadIdentityAnnotationKey] = sa.Annotations[GCPWorkloadIdentityAnnotationKey]
	return annotations, nil
}

func getServiceAccount(ctx context.Context, c client.Client, ref kmapi.ObjectReference) (*core.ServiceAccount, error) {
	sa := &core.ServiceAccount{}
	if err := c.Get(ctx, ref.ObjectKey(), sa); err != nil {
		return nil, err
	}
	return sa, nil
}
