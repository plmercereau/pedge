package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MQTTSpec struct {
	// +kubebuilder:validation:Required
	SensorsTopic string `json:"sensorsTopic,omitempty"`
}

type InfluxDB struct {
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace,omitempty"` // TODO add default value through a webhook
	// +kubebuilder:validation:Required
	SecretReference corev1.LocalObjectReference `json:"secretReference,omitempty"` // TODO default to influxdb-auth
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DevicesClusterSpec defines the desired state of DevicesCluster
type DevicesClusterSpec struct {
	// +kubebuilder:validation:Required
	PersistentVolumeClaimName string `json:"persistentVolumeClaimName,omitempty"`
	// +kubebuilder:validation:Required
	MQTT MQTTSpec `json:"mqtt"`
	// +kubebuilder:validation:Optional
	InfluxDB InfluxDB `json:"influxdb"`
}

// DevicesClusterStatus defines the observed state of DevicesCluster
type DevicesClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DevicesCluster is the Schema for the devicesclusters API
type DevicesCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DevicesClusterSpec   `json:"spec,omitempty"`
	Status DevicesClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DevicesClusterList contains a list of DevicesCluster
type DevicesClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DevicesCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DevicesCluster{}, &DevicesClusterList{})
}
