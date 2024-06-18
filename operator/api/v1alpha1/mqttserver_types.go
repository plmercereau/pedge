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

// MQTTServerSpec defines the desired state of MQTTServer
type MQTTServerSpec struct {
	// +kubebuilder:validation:Required
	Queue QueueSpec `json:"queue"`
}

// MQTTServerStatus defines the observed state of MQTTServer
type MQTTServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MQTTServer is the Schema for the mqttservers API
type MQTTServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MQTTServerSpec   `json:"spec,omitempty"`
	Status MQTTServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MQTTServerList contains a list of MQTTServer
type MQTTServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MQTTServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MQTTServer{}, &MQTTServerList{})
}
