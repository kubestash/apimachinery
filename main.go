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
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	v1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	storagecontrollers "stash.appscode.dev/kubestash/controllers/storage"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	addonsv1alpha1 "stash.appscode.dev/kubestash/apis/addons/v1alpha1"
	configv1alpha1 "stash.appscode.dev/kubestash/apis/config/v1alpha1"
	corev1alpha1 "stash.appscode.dev/kubestash/apis/core/v1alpha1"
	storagev1alpha1 "stash.appscode.dev/kubestash/apis/storage/v1alpha1"
	corecontrollers "stash.appscode.dev/kubestash/controllers/core"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(addonsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(corev1alpha1.AddToScheme(scheme))
	utilruntime.Must(storagev1alpha1.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	var err error
	ctrlConfig := configv1alpha1.KubeStashConfig{}
	options := ctrl.Options{Scheme: scheme}
	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&corecontrollers.BackupBatchReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BackupBatch")
		os.Exit(1)
	}
	if err = (&corecontrollers.BackupBlueprintReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BackupBlueprint")
		os.Exit(1)
	}
	if err = (&corecontrollers.BackupConfigurationReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BackupConfiguration")
		os.Exit(1)
	}
	if err = (&corecontrollers.BackupSessionReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BackupSession")
		os.Exit(1)
	}
	if err = (&corecontrollers.RestoreSessionReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RestoreSession")
		os.Exit(1)
	}
	if err = (&storagecontrollers.BackupStorageReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BackupStorage")
		os.Exit(1)
	}
	if err = (&storagecontrollers.RepositoryReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Repository")
		os.Exit(1)
	}
	if err = (&storagecontrollers.SnapshotReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Snapshot")
		os.Exit(1)
	}
	if err = (&corev1alpha1.BackupBatch{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "BackupBatch")
		os.Exit(1)
	}
	if err = (&corev1alpha1.BackupBlueprint{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "BackupBlueprint")
		os.Exit(1)
	}
	if err = (&corev1alpha1.BackupConfiguration{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "BackupConfiguration")
		os.Exit(1)
	}
	if err = (&corev1alpha1.BackupSession{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "BackupSession")
		os.Exit(1)
	}
	if err = (&corev1alpha1.HookTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "HookTemplate")
		os.Exit(1)
	}
	if err = (&corev1alpha1.RestoreSession{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "RestoreSession")
		os.Exit(1)
	}
	if err = (&storagev1alpha1.BackupStorage{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "BackupStorage")
		os.Exit(1)
	}
	if err = (&storagev1alpha1.Repository{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Repository")
		os.Exit(1)
	}
	if err = (&storagev1alpha1.RetentionPolicy{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "RetentionPolicy")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	go updateWebhookCaBundle(mgr, &ctrlConfig)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func updateWebhookCaBundle(mgr manager.Manager, ctrlConfig *configv1alpha1.KubeStashConfig) {
	if mgr.GetCache().WaitForCacheSync(context.TODO()) {
		if err := updateMutatingWebhookCABundle(mgr, ctrlConfig); err != nil {
			setupLog.Error(err, "unable to update caBundle for MutatingWebhookConfiguration")
			os.Exit(1)
		}
		if err := updateValidatingWebhookCABundle(mgr, ctrlConfig); err != nil {
			setupLog.Error(err, "unable to update caBundle for ValidatingWebhookConfiguration")
			os.Exit(1)
		}
	}
}

func updateMutatingWebhookCABundle(mgr manager.Manager, ctrlConfig *configv1alpha1.KubeStashConfig) error {
	webhook := &v1.MutatingWebhookConfiguration{}
	err := mgr.GetClient().Get(context.TODO(), types.NamespacedName{
		Name: ctrlConfig.WebhookInfo.Mutating.Name,
	}, webhook)
	if err != nil {
		return err
	}

	caBundle, err := ioutil.ReadFile(filepath.Join(ctrlConfig.Webhook.CertDir, "ca.crt"))
	if err != nil {
		return err
	}
	for i := range webhook.Webhooks {
		webhook.Webhooks[i].ClientConfig.CABundle = caBundle
	}
	return mgr.GetClient().Update(context.TODO(), webhook, &client.UpdateOptions{})
}

func updateValidatingWebhookCABundle(mgr manager.Manager, ctrlConfig *configv1alpha1.KubeStashConfig) error {
	webhook := &v1.ValidatingWebhookConfiguration{}
	err := mgr.GetClient().Get(context.TODO(), types.NamespacedName{
		Name: ctrlConfig.WebhookInfo.Validating.Name,
	}, webhook)
	if err != nil {
		return err
	}

	caBundle, err := ioutil.ReadFile(filepath.Join(ctrlConfig.Webhook.CertDir, "ca.crt"))
	if err != nil {
		return err
	}
	for i := range webhook.Webhooks {
		webhook.Webhooks[i].ClientConfig.CABundle = caBundle
	}
	return mgr.GetClient().Update(context.TODO(), webhook, &client.UpdateOptions{})
}
