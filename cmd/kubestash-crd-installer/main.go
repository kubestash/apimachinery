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

package main

import (
	"flag"
	"os"

	addonapi "kubestash.dev/apimachinery/apis/addons/v1alpha1"
	coreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/apiextensions"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	metrics "kmodules.xyz/custom-resources/apis/metrics/v1alpha1"
)

var (
	masterURL  string
	kubeConfig string
)

func init() {
	flag.StringVar(&kubeConfig, "kubeconfig", kubeConfig, "Path to a kube config file. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server. Overrides any value in kube config file. Only required if out-of-cluster.")
}

func main() {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		klog.Errorf("Error building kubeconfig: %s", err.Error())
		os.Exit(1)
	}

	crdClient, err := crd_cs.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("Error building CRD client: %s", err.Error())
		os.Exit(1)
	}

	if err := registerCRDs(crdClient); err != nil {
		klog.Errorf("Error building CRD client: %s", err.Error())
		os.Exit(1)
	}
	klog.Infoln("Successfully installed all CRDs.")
}

func registerCRDs(crdClient crd_cs.Interface) error {
	var resources []*apiextensions.CustomResourceDefinition

	resources = append(resources, getKubeStashCRDs()...)
	resources = append(resources, getAppCatalogCRDs()...)
	resources = append(resources, getMetricCRDs()...)

	return apiextensions.RegisterCRDs(crdClient, resources)
}

func getKubeStashCRDs() []*apiextensions.CustomResourceDefinition {
	return []*apiextensions.CustomResourceDefinition{
		addonapi.Addon{}.CustomResourceDefinition(),
		addonapi.Function{}.CustomResourceDefinition(),
		coreapi.BackupConfiguration{}.CustomResourceDefinition(),
		coreapi.BackupSession{}.CustomResourceDefinition(),
		coreapi.RestoreSession{}.CustomResourceDefinition(),
		coreapi.BackupBlueprint{}.CustomResourceDefinition(),
		coreapi.HookTemplate{}.CustomResourceDefinition(),
		storageapi.BackupStorage{}.CustomResourceDefinition(),
		storageapi.Repository{}.CustomResourceDefinition(),
		storageapi.Snapshot{}.CustomResourceDefinition(),
		storageapi.RetentionPolicy{}.CustomResourceDefinition(),
	}
}

func getAppCatalogCRDs() []*apiextensions.CustomResourceDefinition {
	return []*apiextensions.CustomResourceDefinition{
		appcatalog.AppBinding{}.CustomResourceDefinition(),
	}
}

func getMetricCRDs() []*apiextensions.CustomResourceDefinition {
	return []*apiextensions.CustomResourceDefinition{
		metrics.MetricsConfiguration{}.CustomResourceDefinition(),
	}
}
