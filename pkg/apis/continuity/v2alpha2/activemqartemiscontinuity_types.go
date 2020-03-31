package v2alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ActiveMQArtemisContinuitySpec defines the desired state of ActiveMQArtemisContinuity
// +k8s:openapi-gen=true
type ActiveMQArtemisContinuitySpec struct {
	SiteId                   string  `json:"siteId,omitempty"`
	LocalContinuityUser      string  `json:"localContinuityUser,omitempty"`
	LocalContinuityPass      string  `json:"localContinuityPass,omitempty"`
	RemoteContinuityPass     string  `json:"remoteContinuityPass,omitempty"`
	RemoteContinuityUser     string  `json:"remoteContinuityUser,omitempty"`
	ActiveOnStart            bool    `json:"activeOnStart,omitempty"`
	InflowStagingDelay       int     `json:"inflowStagingDelay,omitempty"`
	BridgeInterval           int     `json:"bridgeInterval,omitempty"`
	BridgeIntervalMultiplier float32 `json:"bridgeIntervalMultiplier,omitempty"`
	PollDuration             int     `json:"pollDuration,omitempty"`
	ActivationTimeout        int     `json:"activationTimeout,omitempty"`
	ReorgManagement          bool    `json:"reorgManagement,omitempty"`
}

// ActiveMQArtemisContinuityStatus defines the observed state of ActiveMQArtemisContinuity
// +k8s:openapi-gen=true
type ActiveMQArtemisContinuityStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemisContinuity is the Schema for the ActiveMQArtemisContinuity API
// +k8s:openapi-gen=true
type ActiveMQArtemisContinuity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ActiveMQArtemisContinuitySpec   `json:"spec,omitempty"`
	Status ActiveMQArtemisContinuityStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemisContinuityList contains a list of ActiveMQArtemisContinuity
type ActiveMQArtemisContinuityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ActiveMQArtemisContinuity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ActiveMQArtemisContinuity{}, &ActiveMQArtemisContinuityList{})
}
