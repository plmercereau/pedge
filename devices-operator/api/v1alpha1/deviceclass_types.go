package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type FirmwareBuilder struct {
	Image ImageRef `json:"image,omitempty"`
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=name
	Env []core.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

type ConfigBuilder struct {
	Image ImageRef `json:"image,omitempty"`
}

type DeviceClusterReference struct {
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
}

// DeviceClassSpec defines the desired state of DeviceClass
type DeviceClassSpec struct {
	// +kubebuilder:validation:Required
	DeviceClusterReference DeviceClusterReference `json:"deviceClusterReference"`
	// +kubebuilder:validation:Required
	Builder FirmwareBuilder `json:"builder,omitempty"`
	// +kubebuilder:validation:Required
	Config ConfigBuilder `json:"config,omitempty"`
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
