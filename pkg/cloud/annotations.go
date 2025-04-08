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
	AWSIRSARoleAnnotation        = "eks.amazonaws.com/role-arn"
	AKSManagedIdentityAnnotation = "azure.workload.identity/client-id"
)

func GetCloudOIDCAnnotationFromOperator(ctx context.Context, kc client.Client) (map[string]string, error) {
	sa, err := getServiceAccount(ctx, kc, kmapi.ObjectReference{
		Name:      meta.PodServiceAccount(),
		Namespace: meta.PodNamespace(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get service account: %w", err)
	}

	annotations := make(map[string]string)
	setValueIfPresent(sa, AWSIRSARoleAnnotation, annotations)
	setValueIfPresent(sa, AKSManagedIdentityAnnotation, annotations)

	return annotations, nil
}

func setValueIfPresent(sa *core.ServiceAccount, key string, out map[string]string) {
	if val, ok := sa.Annotations[key]; ok && val != "" {
		out[key] = val
	}
}

func getServiceAccount(ctx context.Context, c client.Client, ref kmapi.ObjectReference) (*core.ServiceAccount, error) {
	sa := &core.ServiceAccount{}
	if err := c.Get(ctx, ref.ObjectKey(), sa); err != nil {
		return nil, err
	}
	return sa, nil
}
