package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WildflySpec defines the desired state of Wildfly
// +k8s:openapi-gen=true
type WildflySpec struct {
	Size    int32              `json:"size"`
	Image   string             `json:"image"`
	Version string             `json:"version"`
	Cmd     []string           `json:"cmd"`
	Ports   []WildflyPortProto `json:"ports"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// WildflyPorts defines sets of port/protocol
type WildflyPortProto struct {
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

// WildflyStatus defines the observed state of Wildfly
// +k8s:openapi-gen=true
type WildflyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Wildfly is the Schema for the wildflies API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Wildfly struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WildflySpec   `json:"spec,omitempty"`
	Status WildflyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WildflyList contains a list of Wildfly
type WildflyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Wildfly `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Wildfly{}, &WildflyList{})
}
