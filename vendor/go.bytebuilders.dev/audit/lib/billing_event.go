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

package lib

import (
	"context"
	"strings"

	api "go.bytebuilders.dev/audit/api/v1"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	clustermeta "kmodules.xyz/client-go/cluster"
	"kmodules.xyz/client-go/discovery"
	corev1alpha1 "kmodules.xyz/resource-metadata/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BillingEventCreator struct {
	Mapper          discovery.ResourceMapper
	ClusterMetadata *kmapi.ClusterMetadata
	ClientBilling   bool
	NamespaceLister client.Reader
	PodLister       client.Reader
	PVCLister       client.Reader
}

func (p *BillingEventCreator) CreateEvent(obj client.Object) (*api.Event, error) {
	rid, err := p.Mapper.ResourceIDForGVK(obj.GetObjectKind().GroupVersionKind())
	if err != nil {
		return nil, err
	}

	res, err := corev1alpha1.ToGenericResource(obj, rid, p.ClusterMetadata)
	if err != nil {
		return nil, err
	}

	if p.NamespaceLister != nil {
		var ns core.Namespace
		err = p.NamespaceLister.Get(context.TODO(), client.ObjectKey{Name: obj.GetNamespace()}, &ns)
		if err != nil {
			return nil, err
		}
		res.Spec.Namespace = &corev1alpha1.NamespaceInfo{
			UID:               ns.UID,
			Name:              ns.Name,
			CreationTimestamp: ns.CreationTimestamp,
		}
		res.Spec.Namespace.EnableResourceTrial = ns.Annotations[kmapi.AceEnableResourceTrialKey] == "true"
		if ns.Labels[kmapi.ClientOrgKey] == "true" {
			res.Spec.Namespace.AceOrgID = ns.Annotations[kmapi.AceOrgIDKey]

			orgMetadata := map[string]string{}
			for k, v := range ns.Annotations {
				if after, found := strings.CutPrefix(k, kmapi.ClientKeyPrefix); found {
					orgMetadata[after] = v
				}
			}
			res.Spec.Namespace.AceOrgMetadata = orgMetadata
		}

		// ensure cluster mode is always up-to-date
		var ks core.Namespace
		err = p.NamespaceLister.Get(context.TODO(), client.ObjectKey{Name: metav1.NamespaceSystem}, &ks)
		if err != nil {
			return nil, err
		}
		res.Spec.Cluster.Mode = clustermeta.DetectClusterMode(&ks)
	}

	if p.ClientBilling {
		if r, ok := obj.(Resource); ok {
			var podList core.PodList
			err = p.PodLister.List(context.TODO(), &podList, client.InNamespace(obj.GetNamespace()), client.MatchingLabels(r.OffshootSelectors()))
			if err != nil {
				return nil, err
			}
			podresources := make([]corev1alpha1.ComputeResource, 0, len(podList.Items))
			for _, pod := range podList.Items {
				pr := corev1alpha1.ComputeResource{
					UID:               pod.UID,
					Name:              pod.Name,
					CreationTimestamp: pod.CreationTimestamp,
					Containers:        make([]corev1alpha1.ContainerResource, 0, len(pod.Spec.Containers)),
					InitContainers:    make([]corev1alpha1.ContainerResource, 0, len(pod.Spec.Containers)),
				}
				for _, c := range pod.Spec.Containers {
					pr.Containers = append(pr.Containers, corev1alpha1.ContainerResource{
						Name:          c.Name,
						Resource:      c.Resources,
						RestartPolicy: c.RestartPolicy,
					})
				}
				for _, c := range pod.Spec.InitContainers {
					pr.InitContainers = append(pr.InitContainers, corev1alpha1.ContainerResource{
						Name:          c.Name,
						Resource:      c.Resources,
						RestartPolicy: c.RestartPolicy,
					})
				}
				podresources = append(podresources, pr)
			}
			res.Spec.Pods = podresources

			var pvcList core.PersistentVolumeClaimList
			err = p.PVCLister.List(context.TODO(), &pvcList, client.InNamespace(obj.GetNamespace()), client.MatchingLabels(r.OffshootSelectors()))
			if err != nil {
				return nil, err
			}
			pvcresources := make([]corev1alpha1.StorageResource, 0, len(podList.Items))
			for _, pvc := range pvcList.Items {
				if pvc.Status.Phase == core.ClaimBound {
					pvcresources = append(pvcresources, corev1alpha1.StorageResource{
						UID:               pvc.UID,
						Name:              pvc.Name,
						CreationTimestamp: pvc.CreationTimestamp,
						Resources:         pvc.Spec.Resources,
					})
				}
			}
			res.Spec.Storage = pvcresources
		}
	}

	return &api.Event{
		Resource:   res,
		ResourceID: *rid,
	}, nil
}

type Resource interface {
	OffshootSelectors() map[string]string
}
