package v1alpha1

import (
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

	// VolumeClaimTemplates is a list of `PersistentVolumeClaim` templates for Minecraft server container.
	// A claim named "minecraft-data" must be included in the list.
	// +kubebuilder:validation:MinItems=1
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates"`
}

func (s MinecraftSpec) validateCreate() field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}

func (s MinecraftSpec) validateUpdate(old MinecraftSpec) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
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

// PodTemplateSpec describes the data a pod should have when created from a template.
// This is slightly modified from corev1.PodTemplateSpec.
type PodTemplateSpec struct {
	// Standard object's metadata.  The name in this metadata is ignored.
	// +optional
	ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the pod.
	Spec PodSpec `json:"spec"`
}

// PodSpec is a description of a pod.
// This is slightly modified from corev1.PodSpec.
type PodSpec struct {
	// Standard object's metadata.  The name in this metadata is ignored.
	// +optional
	ObjectMeta `json:"metadata,omitempty"`

	// List of containers belonging to the pod.
	// The name of the Minecraft server container in this Containers must be `minecraft`.
	Containers []corev1.Container `json:"containers"`

	// List of volumes that can be mounted by containers belonging to the pod.
	// +optional
	Volumes []corev1.Volume `json:"volumes"`

	// Specification of the desired behavior of the service.
	// +optional
	ServiceSpec corev1.ServiceSpec `json:"serviceSpec,omitempty"`

	// MinecraftConfigMapName is a `ConfigMap` name of Minecraft server config.
	// +nullable
	// +optional
	MinecraftConfigMapName *string `json:"minecraftConfigMapName,omitempty"`
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
