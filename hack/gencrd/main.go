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
	"io/ioutil"
	"os"
	"path/filepath"

	addonsinstall "stash.appscode.dev/kubestash/apis/addons/install"
	addonsv1alpha1 "stash.appscode.dev/kubestash/apis/addons/v1alpha1"
	coreinstall "stash.appscode.dev/kubestash/apis/core/install"
	corev1alpha1 "stash.appscode.dev/kubestash/apis/core/v1alpha1"
	storageinstall "stash.appscode.dev/kubestash/apis/storage/install"
	storagev1alpha1 "stash.appscode.dev/kubestash/apis/storage/v1alpha1"

	"github.com/go-openapi/spec"
	gort "gomodules.xyz/runtime"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
	"k8s.io/kube-openapi/pkg/common"
	"kmodules.xyz/client-go/openapi"
)

func generateSwaggerJson(version string) {
	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)

	addonsinstall.Install(Scheme)
	coreinstall.Install(Scheme)
	storageinstall.Install(Scheme)

	apispec, err := openapi.RenderOpenAPISpec(openapi.Config{
		Scheme: Scheme,
		Codecs: Codecs,
		Info: spec.InfoProps{
			Title:   "KubeStash",
			Version: version,
			Contact: &spec.ContactInfo{
				Name:  "AppsCode Inc.",
				URL:   "https://appscode.com",
				Email: "hello@appscode.com",
			},
			License: &spec.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			addonsv1alpha1.GetOpenAPIDefinitions,
			corev1alpha1.GetOpenAPIDefinitions,
			storagev1alpha1.GetOpenAPIDefinitions,
		},
		//nolint:govet
		Resources: []openapi.TypeInfo{
			// addons/v1alpha1 resources
			{addonsv1alpha1.GroupVersion, addonsv1alpha1.ResourcePluralAddon, addonsv1alpha1.ResourceKindAddon, false},
			{addonsv1alpha1.GroupVersion, addonsv1alpha1.ResourcePluralFunction, addonsv1alpha1.ResourceKindFunction, false},

			// core/v1alpha1 resources
			{corev1alpha1.GroupVersion, corev1alpha1.ResourcePluralBackupBatch, corev1alpha1.ResourceKindBackupBatch, true},
			{corev1alpha1.GroupVersion, corev1alpha1.ResourcePluralBackupBlueprint, corev1alpha1.ResourceKindBackupBlueprint, true},
			{corev1alpha1.GroupVersion, corev1alpha1.ResourcePluralBackupConfiguration, corev1alpha1.ResourceKindBackupConfiguration, true},
			{corev1alpha1.GroupVersion, corev1alpha1.ResourcePluralBackupSession, corev1alpha1.ResourceKindBackupSession, true},
			{corev1alpha1.GroupVersion, corev1alpha1.ResourcePluralHookTemplate, corev1alpha1.ResourceKindHookTemplate, true},
			{corev1alpha1.GroupVersion, corev1alpha1.ResourcePluralRestoreSession, corev1alpha1.ResourceKindRestoreSession, true},

			// storage/v1alpha1 resources
			{storagev1alpha1.GroupVersion, storagev1alpha1.ResourcePluralBackupStorage, storagev1alpha1.ResourceKindBackupStorage, false},
			{storagev1alpha1.GroupVersion, storagev1alpha1.ResourcePluralRepository, storagev1alpha1.ResourceKindRepository, false},
			{storagev1alpha1.GroupVersion, storagev1alpha1.ResourcePluralRetentionPolicy, storagev1alpha1.ResourceKindRetentionPolicy, false},
			{storagev1alpha1.GroupVersion, storagev1alpha1.ResourcePluralSnapshot, storagev1alpha1.ResourceKindSnapshot, false},
		},
	})
	if err != nil {
		klog.Fatal(err)
	}

	filename := gort.GOPath() + "/src/stash.appscode.dev/kubestash/openapi/swagger.json"
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		klog.Fatal(err)
	}
	err = ioutil.WriteFile(filename, []byte(apispec), 0644)
	if err != nil {
		klog.Fatal(err)
	}
}

func main() {
	var version string
	flag.StringVar(&version, "version", "v0.1.0", "KubeStash version")
	flag.Parse()
	generateSwaggerJson(version)
}
