package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ArtefactsSpec struct {
	Image   ImageRef    `json:"image,omitempty"`
	Ingress IngressSpec `json:"ingress,omitempty"`
}

type IngressSpec struct {
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Required
	Hostname string `json:"hostname,omitempty"`
}

type MQTTSpec struct {
	// +kubebuilder:validation:Required
	SensorsTopic string `json:"sensorsTopic,omitempty"`
	// +kubebuilder:validation:Optional
	Hostname string `json:"hostname,omitempty"`
	// +kubebuilder:validation:Optional
	Port  int32      `json:"port,omitempty"`
	Users []MQTTUser `json:"users,omitempty"`
}

type SecretInjectionMapping struct {
	// +kubebuilder:validation:Required
	Username string `json:"username,omitempty"`
	// +kubebuilder:validation:Required
	Password string `json:"password,omitempty"`
	// +kubebuilder:validation:Optional
	SensorsTopic string `json:"sensorsTopic,omitempty"`
	// +kubebuilder:validation:Optional
	BrokerUrl string `json:"brokerUrl,omitempty"`
}

type SecretInjection struct {
	// name is unique within a namespace to reference a secret resource.
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// namespace defines the space within which the secret name must be unique.
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	// +kubebuilder:validation:Required
	Mapping SecretInjectionMapping `json:"mapping,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default []
	Annotations map[string]string `json:"annotations,omitempty"`
}

type MQTTUser struct {
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Required
	Role string `json:"role,omitempty"`
	// +kubebuilder:validation:Optional
	InjectSecrets []SecretInjection `json:"injectSecrets,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DeviceClusterSpec defines the desired state of DeviceCluster
type DeviceClusterSpec struct {
	// Configuration for storing and accessing artefacts
	Artefacts ArtefactsSpec `json:"artefacts,omitempty"`
	// +kubebuilder:validation:Required
	// Configuration of the MQTT broker
	MQTT MQTTSpec `json:"mqtt"`
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
