package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ActiveMQArtemisAddressSpec defines the desired state of ActiveMQArtemisAddress
// +k8s:openapi-gen=true
type ActiveMQArtemisAddressSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	AddressName string `json:"addressName"`
	QueueName   string `json:"queueName"`
	RoutingType string `json:"routingType"`
}

// ActiveMQArtemisAddressStatus defines the observed state of ActiveMQArtemisAddress
// +k8s:openapi-gen=true
type ActiveMQArtemisAddressStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemisAddress is the Schema for the activemqartemisaddresses API
// +k8s:openapi-gen=true
type ActiveMQArtemisAddress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ActiveMQArtemisAddressSpec   `json:"spec,omitempty"`
	Status ActiveMQArtemisAddressStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemisAddressList contains a list of ActiveMQArtemisAddress
type ActiveMQArtemisAddressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ActiveMQArtemisAddress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ActiveMQArtemisAddress{}, &ActiveMQArtemisAddressList{})
}
