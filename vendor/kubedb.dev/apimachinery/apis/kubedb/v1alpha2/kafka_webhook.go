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
	"context"
	"errors"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"

	errors2 "github.com/pkg/errors"
	"gomodules.xyz/pointer"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var kafkalog = logf.Log.WithName("kafka-resource")

func (k *Kafka) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(k).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kafka-kubedb-com-v1alpha1-kafka,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=kafkas,verbs=create;update,versions=v1alpha1,name=mkafka.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &Kafka{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *Kafka) Default(ctx context.Context, obj runtime.Object) error {
	if k == nil {
		return nil
	}
	kafkalog.Info("default", "name", k.Name)
	k.SetDefaults()
	return nil
}

//+kubebuilder:webhook:path=/validate-kafka-kubedb-com-v1alpha1-kafka,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=kafkas,verbs=create;update,versions=v1alpha1,name=vkafka.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &Kafka{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *Kafka) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	kafkalog.Info("validate create", "name", k.Name)
	return nil, k.ValidateCreateOrUpdate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *Kafka) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	kafkalog.Info("validate update", "name", k.Name)
	return nil, k.ValidateCreateOrUpdate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *Kafka) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	kafkalog.Info("validate delete", "name", k.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	var allErr field.ErrorList
	if k.Spec.DeletionPolicy == DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("teminationPolicy"),
			k.Name,
			"Can not delete as terminationPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "Kafka"}, k.Name, allErr)
	}
	return nil, nil
}

func (k *Kafka) ValidateCreateOrUpdate() error {
	var allErr field.ErrorList
	// TODO(user): fill in your validation logic upon object creation.
	if k.Spec.EnableSSL {
		if k.Spec.TLS == nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("enableSSL"),
				k.Name,
				".spec.tls can't be nil, if .spec.enableSSL is true"))
		}
	} else {
		if k.Spec.TLS != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("enableSSL"),
				k.Name,
				".spec.tls must be nil, if .spec.enableSSL is disabled"))
		}
	}
	if k.Spec.Topology != nil {
		if k.Spec.Topology.Controller == nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology").Child("controller"),
				k.Name,
				".spec.topology.controller can't be empty in topology cluster"))
		}
		if k.Spec.Topology.Broker == nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology").Child("broker"),
				k.Name,
				".spec.topology.broker can't be empty in topology cluster"))
		}

		if k.Spec.Replicas != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
				k.Name,
				"doesn't support spec.replicas when spec.topology is set"))
		}
		if k.Spec.Storage != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology").Child("broker"),
				k.Name,
				"doesn't support spec.storage when spec.topology is set"))
		}
		if k.Spec.PodTemplate.Spec.Containers[0].Resources.Size() != 0 {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("resources"),
				k.Name,
				"doesn't support spec.podTemplate.spec.resources when spec.topology is set"))
		}

		if *k.Spec.Topology.Controller.Replicas <= 0 {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology").Child("controller").Child("replicas"),
				k.Name,
				"number of replicas can not be less be 0 or less"))
		}

		if *k.Spec.Topology.Broker.Replicas <= 0 {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology").Child("broker").Child("replicas"),
				k.Name,
				"number of replicas can not be 0 or less"))
		}

		// validate that multiple nodes don't have same suffixes
		err := k.validateNodeSuffix(k.Spec.Topology)
		if err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology"),
				k.Name,
				err.Error()))
		}

		// validate that node replicas are not 0 or negative
		err = k.validateNodeReplicas(k.Spec.Topology)
		if err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology"),
				k.Name,
				err.Error()))
		}
	} else {
		// number of replicas can not be 0 or less
		if k.Spec.Replicas != nil && *k.Spec.Replicas <= 0 {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
				k.Name,
				"number of replicas can not be 0 or less"))
		}
	}

	if k.Spec.Halted && k.Spec.DeletionPolicy == DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("halted"),
			k.Name,
			`can't halt if deletionPolicy is set to "DoNotTerminate"`))
	}

	err := k.validateVersion(k)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
			k.Name,
			err.Error()))
	}

	err = k.validateVolumes(k)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumes"),
			k.Name,
			err.Error()))
	}

	err = k.validateVolumesMountPaths(&k.Spec.PodTemplate)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumeMounts"),
			k.Name,
			err.Error()))
	}

	if k.Spec.StorageType == "" {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storageType"),
			k.Name,
			"StorageType can not be empty"))
	} else {
		if k.Spec.StorageType != StorageTypeDurable && k.Spec.StorageType != StorageTypeEphemeral {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("storageType"),
				k.Name,
				"StorageType should be either durable or ephemeral"))
		}
		if k.Spec.StorageType == StorageTypeEphemeral && k.Spec.DeletionPolicy == DeletionPolicyHalt {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("deletionPolicy"),
				k.Name,
				`'spec.deletionPolicy: Halt' can not be used for 'Ephemeral' storage`))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "Kafka"}, k.Name, allErr)
}

func (k *Kafka) validateVersion(db *Kafka) error {
	kfVersion := &catalog.KafkaVersion{}
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, kfVersion)
	if err != nil {
		return errors.New("version not supported")
	}
	return nil
}

func (k *Kafka) validateNodeSuffix(topology *KafkaClusterTopology) error {
	tMap := topology.ToMap()
	names := make(map[string]bool)
	for _, value := range tMap {
		names[value.Suffix] = true
	}
	if len(tMap) != len(names) {
		return errors.New("two or more node cannot have same suffix")
	}
	return nil
}

func (k *Kafka) validateNodeReplicas(topology *KafkaClusterTopology) error {
	tMap := topology.ToMap()
	for key, node := range tMap {
		if pointer.Int32(node.Replicas) <= 0 {
			return errors2.Errorf("replicas for node role %s must be alteast 1", string(key))
		}
	}
	return nil
}

var kafkaReservedVolumes = []string{
	kubedb.KafkaVolumeData,
	kubedb.KafkaVolumeConfig,
	kubedb.KafkaVolumeTempConfig,
}

func (k *Kafka) validateVolumes(db *Kafka) error {
	if db.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := make([]string, len(kafkaReservedVolumes))
	copy(rsv, kafkaReservedVolumes)
	if db.Spec.TLS != nil && db.Spec.TLS.Certificates != nil {
		for _, c := range db.Spec.TLS.Certificates {
			rsv = append(rsv, db.CertSecretVolumeName(KafkaCertificateAlias(c.Alias)))
		}
	}
	volumes := db.Spec.PodTemplate.Spec.Volumes
	for _, rv := range rsv {
		for _, ugv := range volumes {
			if ugv.Name == rv {
				return errors.New("Cannot use a reserve volume name: " + rv)
			}
		}
	}
	return nil
}

var kafkaReservedVolumeMountPaths = []string{
	kubedb.KafkaConfigDir,
	kubedb.KafkaTempConfigDir,
	kubedb.KafkaDataDir,
	kubedb.KafkaMetaDataDir,
	kubedb.KafkaCertDir,
}

func (k *Kafka) validateVolumesMountPaths(podTemplate *ofst.PodTemplateSpec) error {
	if podTemplate == nil {
		return nil
	}
	if podTemplate.Spec.Containers == nil {
		return nil
	}

	for _, rvmp := range kafkaReservedVolumeMountPaths {
		containerList := podTemplate.Spec.Containers
		for i := range containerList {
			mountPathList := containerList[i].VolumeMounts
			for j := range mountPathList {
				if mountPathList[j].MountPath == rvmp {
					return errors.New("Can't use a reserve volume mount path name: " + rvmp)
				}
			}
		}
	}

	if podTemplate.Spec.InitContainers == nil {
		return nil
	}

	for _, rvmp := range kafkaReservedVolumeMountPaths {
		containerList := podTemplate.Spec.InitContainers
		for i := range containerList {
			mountPathList := containerList[i].VolumeMounts
			for j := range mountPathList {
				if mountPathList[j].MountPath == rvmp {
					return errors.New("Can't use a reserve volume mount path name: " + rvmp)
				}
			}
		}
	}

	return nil
}
