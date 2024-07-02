package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type FirmwareBuilderImage struct {
	Registry   string          `json:"registry,omitempty"`
	Repository string          `json:"repository,omitempty"`
	Tag        string          `json:"tag,omitempty"`
	PullPolicy core.PullPolicy `json:"pullPolicy,omitempty"`
}

type FirmwareBuilder struct {
	Image FirmwareBuilderImage `json:"image,omitempty"`
}

// DeviceClassSpec defines the desired state of DeviceClass
type DeviceClassSpec struct {
	// +kubebuilder:validation:Optional
	Builder FirmwareBuilder `json:"builder,omitempty"`
}

// DeviceClassStatus defines the observed state of DeviceClass
type DeviceClassStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DeviceClass is the Schema for the deviceClasses API
type DeviceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceClassSpec   `json:"spec,omitempty"`
	Status DeviceClassStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DeviceClassList contains a list of DeviceClass
type DeviceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceClass{}, &DeviceClassList{})
}
