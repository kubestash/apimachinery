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

package v1alpha2

import (
	"fmt"
	"reflect"

	"kmodules.xyz/resource-metrics/api"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	api.Register(schema.GroupVersionKind{
		Group:   "kubedb.com",
		Version: "v1alpha2",
		Kind:    "Cassandra",
	}, Cassandra{}.ResourceCalculator())
}

type Cassandra struct{}

func (r Cassandra) ResourceCalculator() api.ResourceCalculator {
	return &api.ResourceCalculatorFuncs{
		AppRoles:               []api.PodRole{api.PodRoleDefault},
		RuntimeRoles:           []api.PodRole{api.PodRoleDefault, api.PodRoleExporter},
		RoleReplicasFn:         r.roleReplicasFn,
		ModeFn:                 r.modeFn,
		UsesTLSFn:              r.usesTLSFn,
		RoleResourceLimitsFn:   r.roleResourceFn(api.ResourceLimits),
		RoleResourceRequestsFn: r.roleResourceFn(api.ResourceRequests),
	}
}

func (r Cassandra) roleReplicasFn(obj map[string]interface{}) (api.ReplicaList, error) {
	result := api.ReplicaList{}
	topology, found, err := unstructured.NestedMap(obj, "spec", "topology")
	if err != nil {
		return nil, err
	}
	if found && topology != nil {
		var replicas int64 = 0

		racks, _, err := unstructured.NestedSlice(topology, "rack")
		if err != nil {
			return nil, err
		}

		for _, rack := range racks {

			replica, _, err := unstructured.NestedInt64(rack.(map[string]interface{}), "replicas")
			if err != nil {
				return nil, err
			}
			replicas += replica
		}
		result[api.PodRoleDefault] = replicas

	} else {
		// standalone
		replicas, found, err := unstructured.NestedInt64(obj, "spec", "replicas")
		if err != nil {
			return nil, fmt.Errorf("failed to read spec.replicas %v: %w", obj, err)
		}
		if !found {
			result[api.PodRoleDefault] = 1
		} else {
			result[api.PodRoleDefault] = replicas
		}
	}
	return result, nil
}

func (r Cassandra) modeFn(obj map[string]interface{}) (string, error) {
	topology, found, err := unstructured.NestedFieldNoCopy(obj, "spec", "topology")
	if err != nil {
		return "", err
	}
	if found && !reflect.ValueOf(topology).IsNil() {
		return DBModeCluster, nil
	}
	return DBModeStandalone, nil
}

func (r Cassandra) usesTLSFn(obj map[string]interface{}) (bool, error) {
	_, found, err := unstructured.NestedFieldNoCopy(obj, "spec", "tls")
	return found, err
}

func (r Cassandra) roleResourceFn(fn func(rr core.ResourceRequirements) core.ResourceList) func(obj map[string]interface{}) (map[api.PodRole]api.PodInfo, error) {
	return func(obj map[string]interface{}) (map[api.PodRole]api.PodInfo, error) {
		exporter, err := api.ContainerResources(obj, fn, "spec", "monitor", "prometheus", "exporter")
		if err != nil {
			return nil, err
		}

		topology, found, err := unstructured.NestedMap(obj, "spec", "topology")
		if err != nil {
			return nil, err
		}
		if found && topology != nil {
			racks, _, err := unstructured.NestedSlice(topology, "rack")
			if err != nil {
				return nil, err
			}
			var totalReplicas int64 = 0
			var totalRes core.ResourceList
			for i, c := range racks {
				rack := c.(map[string]interface{})

				cRes, cReplicas, err := api.AppNodeResourcesV2(rack, fn, CassandraContainerName)
				if err != nil {
					return nil, err
				}
				totalReplicas += cReplicas
				if i == 0 {
					totalRes = cRes
				}
			}

			return map[api.PodRole]api.PodInfo{
				api.PodRoleDefault:  {Resource: totalRes, Replicas: totalReplicas},
				api.PodRoleExporter: {Resource: exporter, Replicas: totalReplicas},
			}, nil
		} else {
			container, replicas, err := api.AppNodeResourcesV2(obj, fn, CassandraContainerName, "spec")
			if err != nil {
				return nil, err
			}
			return map[api.PodRole]api.PodInfo{
				api.PodRoleDefault:  {Resource: container, Replicas: replicas},
				api.PodRoleExporter: {Resource: exporter, Replicas: replicas},
			}, nil
		}
	}
}
