package v2alpha1

import (
	"github.com/RHsyseng/operator-utils/pkg/olm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ActiveMQArtemisSpec defines the desired state of ActiveMQArtemis
// +k8s:openapi-gen=true
type ActiveMQArtemisSpec struct {
	// User name for standard broker user. It is required for connecting to the broker. If left empty, it will be generated.
	AdminUser string `json:"adminUser,omitempty"`
	// Password for standard broker user. It is required for connecting to the broker. If left empty, it will be generated.
	AdminPassword  string             `json:"adminPassword,omitempty"`
	DeploymentPlan DeploymentPlanType `json:"deploymentPlan,omitempty"`
	// Configuration of all acceptors
	Acceptors []AcceptorType `json:"acceptors,omitempty"`
	// Configuration of all connectors
	Connectors []ConnectorType `json:"connectors,omitempty"`
	// Configuration for the embedded web console
	Console ConsoleType `json:"console,omitempty"`
	// Whether or not to enable the continuity plugin. See required fields for continuity.
	EnableContinuity bool `json:"enableContinuity,omitempty"`
	// Name the continuity site. Must be unique in the set of peers (is same across the artemis cluster). Required for continuity.
	SiteId string `json:"siteId,omitempty"`
	// Username to connect to the local broker for continuity connections. Required for continuity.
	LocalContinuityUser string `json:"localContinuityUser,omitempty"`
	// Password to connect to the local broker for continuity connections. Required for continuity.
	LocalContinuityPass string `json:"localContinuityPass,omitempty"`
	// Password to connect to the remote broker for continuity connections. Required for continuity.
	RemoteContinuityPass string `json:"remoteContinuityPass,omitempty"`
	// Username to connect to the remote broker for continuity connections. Required for continuity.
	RemoteContinuityUser string `json:"remoteContinuityUser,omitempty"`
	// Username to connect to the peer site broker/cluster for continuity connections. Required for continuity.
	PeerSiteUrl string `json:"peerSiteUrl,omitempty"`
	// Username to connect to the peer site broker/cluster for continuity connections. Required for continuity.
	PeerContinuityUser string `json:"peerContinuityUser,omitempty"`
	// Password to connect to the peer site broker/cluster for continuity connections. Required for continuity.
	PeerContinuityPass string `json:"peerContinuityPass,omitempty"`
	// Identifies this is the site that should be active when first started. If another active site is connected to, this site will defer to the other. You can also start both sites inactive and explictly activate the desired start. Required for continuity.
	ActiveOnStart bool `json:"activeOnStart,omitempty"`
	// Size of the broker id cache size, used by the broker to remove duplicate messages across sites. Make sure the id cache is sufficiently sized for your volume of messages. The default is 3000.
	BrokerIdCacheSize int `json:"brokerIdCacheSize,omitempty"`
	// Amount of time in millseconds to delay messages in the inflow staging queues before delivering them to the target queues. Useful for active:active site topologies. The default is 60000 ms or 1 minute.
	InflowStagingDelay int `json:"inflowStagingDelay,omitempty"`
	// Bridge reconnection interval for all the bridges created by the continuity plugin. The default is 1000 ms or 1 second.
	BridgeInterval int `json:"bridgeInterval,omitempty"`
	// Bridge reconnection interval backoff multiplier for all the bridges created by the continuity plugin. The default is 0.5.
	BridgeIntervalMultiplier float32 `json:"bridgeIntervalMultiplier,omitempty"`
	// Time in milliseconds between polls to check for a site to be exhausted during deactivation. The default is 100 ms.
	OutflowExhaustedPollDuration int `json:"outflowExhaustedPollDuration,omitempty"`
	// Time in milliseconds between polls to check for a site to be consume all the acks during activation. The default is 100 ms.
	OutflowAcksConsumedPollDuration int `json:"outflowAcksConsumedPollDuration,omitempty"`
	// Time in milliseconds between polls to all the inflow acks have been consumed during activation. The default is 100 ms.
	InflowAcksConsumedPollDuration int `json:"inflowAcksConsumedPollDuration,omitempty"`
	// Time in milliseconds to activate a site and start serving clients, overriding the wait for the peer site to be exhausted, and acks to be consumed. The default is 300000 ms or 5 minutes.
	ActivationTimeout int `json:"activationTimeout,omitempty"`
	// Whether or not to reorganized all the address, queue, divert, and bridge primitives under the continuity hierarchy in JMX/Jolokia. The default is true.
	ReorgManagement bool `json:"reorgManagement,omitempty"`
	// Logging level for the continuity plugin. The default is INFO.
	// +kubebuilder:validation:Enum=TRACE,DEBUG,INFO,ERROR,FATAL
	ContinuityLogLevel string `json:"continuityLogLevel,omitempty"`
}

type DeploymentPlanType struct {
	// The image used for the broker deployment
	Image string `json:"image,omitempty"`
	// The number of broker pods to deploy
	// +kubebuilder:validation:Maximum=16
	// +kubebuilder:validation:Minimum=1
	Size int32 `json:"size,omitempty"`
	// If true require user password login credentials for broker protocol ports
	RequireLogin bool `json:"requireLogin,omitempty"`
	// If true use persistent volume via persistent volume claim for journal storage
	PersistenceEnabled bool `json:"persistenceEnabled,omitempty"`
	// If aio use ASYNCIO, if nio use NIO for journal IO
	JournalType string `json:"journalType,omitempty"`
	// If true migrate messages on scaledown
	MessageMigration *bool `json:"messageMigration,omitempty"`
}

type AcceptorType struct {
	// The name of the acceptor
	Name string `json:"name"`
	// Port number
	Port int32 `json:"port,omitempty"`
	// The protocols to enable for this acceptor
	Protocols string `json:"protocols,omitempty"`
	// Whether or not to enable SSL on this port
	SSLEnabled bool `json:"sslEnabled,omitempty"`
	// Name of the secret to use for ssl information
	SSLSecret string `json:"sslSecret,omitempty"`
	// Comma separated list of cipher suites used for SSL communication.
	EnabledCipherSuites string `json:"enabledCipherSuites,omitempty"`
	// Comma separated list of protocols used for SSL communication.
	EnabledProtocols string `json:"enabledProtocols,omitempty"`
	// Tells a client connecting to this acceptor that 2-way SSL is required. This property takes precedence over wantClientAuth.
	NeedClientAuth bool `json:"needClientAuth,omitempty"`
	// Tells a client connecting to this acceptor that 2-way SSL is requested but not required. Overridden by needClientAuth.
	WantClientAuth bool `json:"wantClientAuth,omitempty"`
	// The CN of the connecting client's SSL certificate will be compared to its hostname to verify they match. This is useful only for 2-way SSL.
	VerifyHost bool `json:"verifyHost,omitempty"`
	// Used to change the SSL Provider between JDK and OPENSSL. The default is JDK.
	SSLProvider string `json:"sslProvider,omitempty"`
	// A regular expression used to match the server_name extension on incoming SSL connections. If the name doesn't match then the connection to the acceptor will be rejected.
	SNIHost string `json:"sniHost,omitempty"`
	// Whether or not to expose this acceptor
	Expose bool `json:"expose,omitempty"`
	// To indicate which kind of routing type to use.
	AnycastPrefix string `json:"anycastPrefix,omitempty"`
	// To indicate which kind of routing type to use.
	MulticastPrefix string `json:"multicastPrefix,omitempty"`
	// Limits the number of connections which the acceptor will allow. When this limit is reached a DEBUG level message is issued to the log, and the connection is refused.
	ConnectionsAllowed int `json:"connectionsAllowed,omitempty"`
}

type ConnectorType struct {
	// The name of the acceptor
	Name string `json:"name"`
	// The type either tcp or vm
	Type string `json:"type,omitempty"`
	// Hostname or IP to connect to
	Host string `json:"host"`
	// Port number
	Port int32 `json:"port"`
	// Whether or not to enable SSL on this port
	SSLEnabled bool `json:"sslEnabled,omitempty"`
	// Name of the secret to use for ssl information
	SSLSecret string `json:"sslSecret,omitempty"`
	// Comma separated list of cipher suites used for SSL communication.
	EnabledCipherSuites string `json:"enabledCipherSuites,omitempty"`
	// Comma separated list of protocols used for SSL communication.
	EnabledProtocols string `json:"enabledProtocols,omitempty"`
	// Tells a client connecting to this acceptor that 2-way SSL is required. This property takes precedence over wantClientAuth.
	NeedClientAuth bool `json:"needClientAuth,omitempty"`
	//Tells a client connecting to this acceptor that 2-way SSL is requested but not required. Overridden by needClientAuth.
	WantClientAuth bool `json:"wantClientAuth,omitempty"`
	// The CN of the connecting client's SSL certificate will be compared to its hostname to verify they match. This is useful only for 2-way SSL.
	VerifyHost bool `json:"verifyHost,omitempty"`
	// Used to change the SSL Provider between JDK and OPENSSL. The default is JDK.
	SSLProvider string `json:"sslProvider,omitempty"`
	// A regular expression used to match the server_name extension on incoming SSL connections. If the name doesn't match then the connection to the acceptor will be rejected.
	SNIHost string `json:"sniHost,omitempty"`
	// Whether or not to expose this connector
	Expose bool `json:"expose,omitempty"`
}

type ConsoleType struct {
	// Whether or not to expose this port
	Expose bool `json:"expose,omitempty"`
	// Whether or not to enable SSL on this port
	SSLEnabled bool `json:"sslEnabled,omitempty"`
	// Name of the secret to use for ssl information
	SSLSecret string `json:"sslSecret,omitempty"`
	// If the embedded server requires client authentication
	UseClientAuth bool `json:"useClientAuth,omitempty"`
}

// ActiveMQArtemisStatus defines the observed state of ActiveMQArtemis
// +k8s:openapi-gen=true
type ActiveMQArtemisStatus struct {
	PodStatus olm.DeploymentStatus `json:"podStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemis is the Schema for the activemqartemis API
// +k8s:openapi-gen=true
type ActiveMQArtemis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ActiveMQArtemisSpec   `json:"spec,omitempty"`
	Status ActiveMQArtemisStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemisList contains a list of ActiveMQArtemis
type ActiveMQArtemisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ActiveMQArtemis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ActiveMQArtemis{}, &ActiveMQArtemisList{})
}
