package v1alpha1

import (
	"fmt"

	"github.com/kmdkuk/mcing/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MinecraftSpec defines the desired state of Minecraft
type MinecraftSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PodTemplate is a `Pod` template for Minecraft server container.
	PodTemplate PodTemplateSpec `json:"podTemplate"`

	// PersistentVolumeClaimSpec is a specification of `PersistentVolumeClaim` for persisting data in minecraft.
	// A claim named "minecraft-data" must be included in the list.
	// +kubebuilder:validation:MinItems=1
	VolumeClaimTemplates []PersistentVolumeClaim `json:"volumeClaimTemplates"`

	// ServiceTemplate is a `Service` template.
	// +optional
	ServiceTemplate *ServiceTemplate `json:"serviceTemplate,omitempty"`

	// operators on server. exec /op or /deop
	// +optional
	Ops Ops `json:"ops,omitempty"`

	// whitelist
	// +optional
	Whitelist Whitelist `json:"whitelist,omitempty"`

	// ServerPropertiesConfigMapName is a `ConfigMap` name of `server.properties`.
	// +nullable
	// +optional
	ServerPropertiesConfigMapName *string `json:"serverPropertiesConfigMapName,omitempty"`

	// OtherConfigMapName is a `ConfigMap` name of other configurations file(eg. banned-ips.json, ops.json etc)
	// +nullable
	// +optional
	OtherConfigMapName *string `json:"otherConfigMapName,omitempty"`
}

type Ops struct {
	// user name exec /op or /deop
	// +optional
	Users []string `json:"users,omitempty"`
}

type Whitelist struct {
	// exec /whitelist on
	// +optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// user name exec /whitelist add or /whitelist remove
	// +optional
	Users []string `json:"users,omitempty"`
}

// PodTemplateSpec describes the data a pod should have when created from a template.
// This is slightly modified from corev1.PodTemplateSpec.
type PodTemplateSpec struct {
	// Standard object's metadata.  The name in this metadata is ignored.
	// +optional
	ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the pod.
	// The name of the MySQL server container in this spec must be `minecraft`.
	Spec corev1.PodSpec `json:"spec"`
}

func (p *PodTemplateSpec) ToCoreV1() corev1.PodTemplateSpec {
	podTemplateSpec := corev1.PodTemplateSpec{}
	podTemplateSpec.Name = p.Name
	if len(p.Labels) > 0 {
		podTemplateSpec.Labels = make(map[string]string)
		for k, v := range p.Labels {
			podTemplateSpec.Labels[k] = v
		}
	}
	if len(p.Annotations) > 0 {
		podTemplateSpec.Annotations = make(map[string]string)
		for k, v := range p.Annotations {
			podTemplateSpec.Annotations[k] = v
		}
	}
	podTemplateSpec.Spec = *p.Spec.DeepCopy()
	return podTemplateSpec
}

// PersistentVolumeClaim is a user's request for and claim to a persistent volume.
// This is slightly modified from corev1.PersistentVolumeClaim.
type PersistentVolumeClaim struct {
	// Standard object's metadata.
	ObjectMeta `json:"metadata"`

	// Spec defines the desired characteristics of a volume requested by a pod author.
	Spec corev1.PersistentVolumeClaimSpec `json:"spec"`
}

func (p PersistentVolumeClaim) ToCoreV1() corev1.PersistentVolumeClaim {
	claim := corev1.PersistentVolumeClaim{}
	claim.Name = p.Name
	if len(p.Labels) > 0 {
		claim.Labels = make(map[string]string)
		for k, v := range p.Labels {
			claim.Labels[k] = v
		}
	}
	if len(p.Annotations) > 0 {
		claim.Annotations = make(map[string]string)
		for k, v := range p.Annotations {
			claim.Annotations[k] = v
		}
	}
	claim.Spec = *p.Spec.DeepCopy()
	if claim.Spec.VolumeMode == nil {
		modeFilesystem := corev1.PersistentVolumeFilesystem
		claim.Spec.VolumeMode = &modeFilesystem
	}
	claim.Status.Phase = corev1.ClaimPending
	return claim
}

// ServiceTemplate define the desired spec and annotations of Service
type ServiceTemplate struct {
	// Standard object's metadata. Only `annotations` and `labels` are valid.
	// +optional
	ObjectMeta `json:"metadata,omitempty"`

	// Spec is the ServiceSpec
	// +optional
	Spec *corev1.ServiceSpec `json:"spec,omitempty"`
}

// ObjectMeta is metadata of objects.
// This is partially copied from metav1.ObjectMeta.
type ObjectMeta struct {
	// Name is the name of the object.
	// +optional
	Name string `json:"name,omitempty"`

	// Labels is a map of string keys and values.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is a map of string keys and values.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (s MinecraftSpec) validateCreate() field.ErrorList {
	var allErrs field.ErrorList

	p := field.NewPath("spec")

	pp := p.Child("volumeClaimTemplates")
	ok := false
	for i := range s.VolumeClaimTemplates {
		vc := &s.VolumeClaimTemplates[i]
		if vc.Name == constants.DataVolumeName {
			ok = true
		}
	}
	if !ok {
		allErrs = append(allErrs, field.Required(pp, fmt.Sprintf("required volume claim template %s is missing", constants.DataVolumeName)))
	}

	p = p.Child("podTemplate", "spec")

	pp = p.Child("containers")
	minecraftIndex := -1
	for i := range s.PodTemplate.Spec.Containers {
		c := &s.PodTemplate.Spec.Containers[i]
		if c.Name == constants.MinecraftContainerName {
			minecraftIndex = i
		}
	}

	if minecraftIndex == -1 {
		allErrs = append(allErrs, field.Required(pp, fmt.Sprintf("required container %s is missing", constants.MinecraftContainerName)))
	} else {
		pp := p.Child("containers").Index(minecraftIndex).Child("ports")
		for i := range s.PodTemplate.Spec.Containers[minecraftIndex].Ports {
			port := &s.PodTemplate.Spec.Containers[minecraftIndex].Ports[i]
			switch port.ContainerPort {
			case constants.ServerPort, constants.RconPort:
				allErrs = append(allErrs, field.Invalid(pp.Index(i), port.ContainerPort, "reserved port"))
			}
			switch port.Name {
			case constants.ServerPortName, constants.RconPortName:
				allErrs = append(allErrs, field.Invalid(pp.Index(i), port.Name, "reserved port name"))
			}
		}

		pp = p.Child("containers").Index(minecraftIndex).Child("env")
		hasEula := false
		for i := range s.PodTemplate.Spec.Containers[minecraftIndex].Env {
			env := &s.PodTemplate.Spec.Containers[minecraftIndex].Env[i]
			switch env.Name {
			case constants.EulaEnvName:
				hasEula = true
			}
		}
		if !hasEula {
			allErrs = append(allErrs, field.Required(pp, "EULA is required. The server will not start unless EULA=true."))
		}
	}
	return allErrs
}

func (s MinecraftSpec) validateUpdate(old MinecraftSpec) field.ErrorList {
	var allErrs field.ErrorList

	return append(allErrs, s.validateCreate()...)
}

// MinecraftStatus defines the observed state of Minecraft
type MinecraftStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Minecraft is the Schema for the minecrafts API
type Minecraft struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinecraftSpec   `json:"spec,omitempty"`
	Status MinecraftStatus `json:"status,omitempty"`
}

func (m *Minecraft) PrefixedName() string {
	return "mcing-" + m.Name
}

func (m *Minecraft) HeadlessServiceName() string {
	return m.PrefixedName() + "-headless"
}

//+kubebuilder:object:root=true

// MinecraftList contains a list of Minecraft
type MinecraftList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Minecraft `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Minecraft{}, &MinecraftList{})
}
