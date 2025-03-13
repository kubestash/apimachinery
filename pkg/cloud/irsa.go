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
	AWSIRSARoleAnnotation = "eks.amazonaws.com/role-arn"
)

func GetIRSARoleAnnotationFromOperator(ctx context.Context, kc client.Client) (map[string]string, error) {
	val, err := GetIRSARoleAnnotationValue(ctx, kc)
	if err != nil {
		return nil, err
	}
	if val != "" {
		return map[string]string{AWSIRSARoleAnnotation: val}, nil
	}
	return nil, nil
}

func GetIRSARoleAnnotationValue(ctx context.Context, kc client.Client) (string, error) {
	sa, err := getServiceAccount(ctx, kc, kmapi.ObjectReference{
		Name:      meta.PodServiceAccount(),
		Namespace: meta.PodNamespace(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve service account: %w", err)
	}

	val, _ := sa.Annotations[AWSIRSARoleAnnotation]
	return val, nil
}

func getServiceAccount(ctx context.Context, c client.Client, ref kmapi.ObjectReference) (*core.ServiceAccount, error) {
	sa := &core.ServiceAccount{}
	if err := c.Get(ctx, ref.ObjectKey(), sa); err != nil {
		return nil, err
	}
	return sa, nil
}
