/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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

	addonsv1alpha1 "stash.appscode.dev/kubestash/apis/addons/v1alpha1"
	corev1alpha1 "stash.appscode.dev/kubestash/apis/core/v1alpha1"
	storagev1alpha1 "stash.appscode.dev/kubestash/apis/storage/v1alpha1"
	"stash.appscode.dev/kubestash/crds"

	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/apiextensions"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	metrics "kmodules.xyz/custom-resources/apis/metrics/v1alpha1"
	kmodules_crds "kmodules.xyz/custom-resources/crds"
)

var (
	masterURL                string
	kubeConfig               string
	enableEnterpriseFeatures bool
)

func init() {
	flag.StringVar(&kubeConfig, "kube-config", "", "Path to a kube config file. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kube config file. Only required if out-of-cluster.")
	flag.BoolVar(&enableEnterpriseFeatures, "enterprise", false, "Specify whether enterprise features enabled or not.")
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

	stashCRDs, err := getKubeStashCRDs()
	if err != nil {
		return err
	}
	resources = append(resources, stashCRDs...)

	appCatalogCRDs, err := getAppCatalogCRDs()
	if err != nil {
		return err
	}
	resources = append(resources, appCatalogCRDs...)

	if enableEnterpriseFeatures {
		metricCRDs, err := getMetricCRDs()
		if err != nil {
			return err
		}
		resources = append(resources, metricCRDs...)
	}
	return apiextensions.RegisterCRDs(crdClient, resources)
}

func getKubeStashCRDs() ([]*apiextensions.CustomResourceDefinition, error) {
	gvrs := []schema.GroupVersionResource{
		// addons/v1alpha1 resources
		{Group: addonsv1alpha1.GroupVersion.Group, Version: addonsv1alpha1.GroupVersion.Version, Resource: addonsv1alpha1.ResourcePluralAddon},
		{Group: addonsv1alpha1.GroupVersion.Group, Version: addonsv1alpha1.GroupVersion.Version, Resource: addonsv1alpha1.ResourcePluralFunction},

		// core/v1alpha1 resources
		{Group: corev1alpha1.GroupVersion.Group, Version: corev1alpha1.GroupVersion.Version, Resource: corev1alpha1.ResourcePluralBackupBatch},
		{Group: corev1alpha1.GroupVersion.Group, Version: corev1alpha1.GroupVersion.Version, Resource: corev1alpha1.ResourcePluralBackupBlueprint},
		{Group: corev1alpha1.GroupVersion.Group, Version: corev1alpha1.GroupVersion.Version, Resource: corev1alpha1.ResourcePluralBackupConfiguration},
		{Group: corev1alpha1.GroupVersion.Group, Version: corev1alpha1.GroupVersion.Version, Resource: corev1alpha1.ResourcePluralBackupSession},
		{Group: corev1alpha1.GroupVersion.Group, Version: corev1alpha1.GroupVersion.Version, Resource: corev1alpha1.ResourcePluralHookTemplate},
		{Group: corev1alpha1.GroupVersion.Group, Version: corev1alpha1.GroupVersion.Version, Resource: corev1alpha1.ResourcePluralRestoreSession},

		// storage/v1alpha1 resources
		{Group: storagev1alpha1.GroupVersion.Group, Version: storagev1alpha1.GroupVersion.Version, Resource: storagev1alpha1.ResourcePluralBackupStorage},
		{Group: storagev1alpha1.GroupVersion.Group, Version: storagev1alpha1.GroupVersion.Version, Resource: storagev1alpha1.ResourcePluralRepository},
		{Group: storagev1alpha1.GroupVersion.Group, Version: storagev1alpha1.GroupVersion.Version, Resource: storagev1alpha1.ResourcePluralRetentionPolicy},
		{Group: storagev1alpha1.GroupVersion.Group, Version: storagev1alpha1.GroupVersion.Version, Resource: storagev1alpha1.ResourcePluralSnapshot},
	}

	var kubeStashCRDs []*apiextensions.CustomResourceDefinition
	for i := range gvrs {
		crd, err := crds.CustomResourceDefinition(gvrs[i])
		if err != nil {
			return nil, err
		}
		kubeStashCRDs = append(kubeStashCRDs, crd)
	}
	return kubeStashCRDs, nil
}

func getAppCatalogCRDs() ([]*apiextensions.CustomResourceDefinition, error) {
	gvrs := []schema.GroupVersionResource{
		// v1alpha1 resources
		{Group: appcatalog.SchemeGroupVersion.Group, Version: appcatalog.SchemeGroupVersion.Version, Resource: appcatalog.ResourceApps},
	}
	var appCatalogCRDs []*apiextensions.CustomResourceDefinition
	for i := range gvrs {
		crd, err := kmodules_crds.CustomResourceDefinition(gvrs[i])
		if err != nil {
			return nil, err
		}
		appCatalogCRDs = append(appCatalogCRDs, crd)
	}
	return appCatalogCRDs, nil
}

func getMetricCRDs() ([]*apiextensions.CustomResourceDefinition, error) {
	gvrs := []schema.GroupVersionResource{
		// v1alpha1 resources
		{Group: metrics.SchemeGroupVersion.Group, Version: metrics.SchemeGroupVersion.Version, Resource: metrics.ResourceMetricsConfigurations},
	}

	var metricCRDs []*apiextensions.CustomResourceDefinition
	for i := range gvrs {
		crd, err := kmodules_crds.CustomResourceDefinition(gvrs[i])
		if err != nil {
			return nil, err
		}
		metricCRDs = append(metricCRDs, crd)
	}
	return metricCRDs, nil
}
