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

	// Image for minecraft server
	// +optional
	Image *string `json:"image,omitempty"`

	// PersistentVolumeClaimSpec is a specification of `PersistentVolumeClaim` for persisting data in minecraft.
	VolumeClaimSpec corev1.PersistentVolumeClaimSpec `json:"volumeClaimSpec"`
}

func (s MinecraftSpec) validateCreate() field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}

func (s MinecraftSpec) validateUpdate(old MinecraftSpec) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
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
