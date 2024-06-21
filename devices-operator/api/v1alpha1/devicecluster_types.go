package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type QueueSpec struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DeviceClusterSpec defines the desired state of DeviceCluster
type DeviceClusterSpec struct {
	// +kubebuilder:validation:Required
	Queue QueueSpec `json:"queue"`
}

// DeviceClusterStatus defines the observed state of DeviceCluster
type DeviceClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DeviceCluster is the Schema for the deviceclusters API
type DeviceCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceClusterSpec   `json:"spec,omitempty"`
	Status DeviceClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DeviceClusterList contains a list of DeviceCluster
type DeviceClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceCluster{}, &DeviceClusterList{})
}
