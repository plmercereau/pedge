package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core "k8s.io/api/core/v1"
)

// TODO  
// storage:
//     # TODO
//     endpoint: http://minio-service.esp-cluster.svc.cluster.local:9000
//     bucket: firmware
//     # TODO add a secret for the access key and secret
//     accessKeyId: ""
//     secretAccessKey: ""

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type FirmwareBuilderImage struct {
	Registry string `json:"registry,omitempty"`
	Repository string `json:"repository,omitempty"`
	Tag string `json:"tag,omitempty"`
	PullPolicy core.PullPolicy `json:"pullPolicy,omitempty"`
}

type FirmwareBuilder struct {
	Image FirmwareBuilderImage `json:"image,omitempty"`
}

// FirmwareSpec defines the desired state of Firmware
type FirmwareSpec struct {
	Builder FirmwareBuilder `json:"builder,omitempty"`
}

// FirmwareStatus defines the observed state of Firmware
type FirmwareStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Firmware is the Schema for the firmwares API
type Firmware struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirmwareSpec   `json:"spec,omitempty"`
	Status FirmwareStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FirmwareList contains a list of Firmware
type FirmwareList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Firmware `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Firmware{}, &FirmwareList{})
}
