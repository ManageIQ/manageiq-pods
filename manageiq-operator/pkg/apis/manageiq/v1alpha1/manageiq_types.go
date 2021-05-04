package v1alpha1

import (
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManageIQSpec defines the desired state of ManageIQ
type ManageIQSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

	// Optional Annotations to apply to the Httpd, Kafka, Memcached, Orchestrator and PostgresQL Pods
	// +optional
	AppAnnotations map[string]string `json:"appAnnotations,omitempty"`

	// Domain name for the external route. Used for external authentication configuration
	ApplicationDomain string `json:"applicationDomain"`

	// Application name used for deployed objects (default: manageiq)
	// +optional
	AppName string `json:"appName"`

	// This label will be applied to essential resources that need to be backed up (default: manageiq.org/backup)
	// +optional
	BackupLabelName string `json:"backupLabelName"`

	// Image string used for the base worker deployments
	// By default this is determined by the orchestrator pod
	// +optional
	BaseWorkerImage string `json:"baseWorkerImage"`

	// Database region number (default: 0)
	// +optional
	DatabaseRegion string `json:"databaseRegion"`

	// Secret containing the database access information, content generated if not provided (default: postgresql-secrets)
	// +optional
	DatabaseSecret string `json:"databaseSecret"`

	// Database volume size (default: 15Gi)
	// +optional
	DatabaseVolumeCapacity string `json:"databaseVolumeCapacity"`

	// Flag to indicate if Kafka and Zookeeper should be deployed (default: false)
	// +optional
	DeployMessagingService *bool `json:"deployMessagingService"`

	// Flag to allow logging into the application without SSO (default: true)
	// +optional
	EnableApplicationLocalLogin *bool `json:"enableApplicationLocalLogin"`

	// Flag to enable SSO in the application (default: false)
	// +optional
	EnableSSO *bool `json:"enableSSO"`

	// Flag to trigger worker resource constraint enforcement (default: false)
	// +optional
	EnforceWorkerResourceConstraints *bool `json:"enforceWorkerResourceConstraints"`

	// Secret containing the httpd configuration files
	// Mutually exclusive with the OIDCClientSecret and OIDCProviderURL if using openid-connect
	// +optional
	HttpdAuthConfig string `json:"httpdAuthConfig"`

	// Type of httpd authentication (default: internal)
	// Options: internal, external, active-directory, saml, openid-connect
	// Note: external, active-directory, and saml require an httpd container with elevated privileges
	// +optional
	// +kubebuilder:validation:Pattern=\A(active-directory|external|internal|openid-connect|saml)\z
	HttpdAuthenticationType string `json:"httpdAuthenticationType"`

	// Httpd deployment CPU limit (default: no limit)
	// +optional
	HttpdCpuLimit string `json:"httpdCpuLimit"`

	// Httpd deployment CPU request (default: no request)
	// +optional
	HttpdCpuRequest string `json:"httpdCpuRequest"`

	// Image string used for the httpd deployment
	// (default: <HttpdImageNamespace>/httpd[-init]:<HttpdImageTag>)
	// +optional
	HttpdImage string `json:"httpdImage"`

	// Image namespace used for the httpd deployment (default: manageiq)
	// Note: the exact image will be determined by the authentication method selected
	// +optional
	HttpdImageNamespace string `json:"httpdImageNamespace"`

	// Image tag used for the httpd deployment (default: latest)
	// +optional
	HttpdImageTag string `json:"httpdImageTag"`

	// Httpd deployment memory limit (default: no limit)
	// +optional
	HttpdMemoryLimit string `json:"httpdMemoryLimit"`

	// Httpd deployment memory request (default: no limit)
	// +optional
	HttpdMemoryRequest string `json:"httpdMemoryRequest"`

	// Secret containing the image registry authentication information needed for the manageiq images
	// +optional
	ImagePullSecret string `json:"imagePullSecret"`

	// Group name to create with the super admin role.
	// This can be used to seed a group when using external authentication
	// +optional
	InitialAdminGroupName string `json:"initialAdminGroupName"`

	// Kafka deployment CPU limit (default: no limit)
	// +optional
	KafkaCpuLimit string `json:"kafkaCpulimit"`

	// Kafka deployment CPU request (default: no request)
	// +optional
	KafkaCpuRequest string `json:"kafkaCpuRequest"`

	// Image string used for the kafka deployment
	// (default: <KafkaImageName>:<KafkaImageTag>)
	// +optional
	KafkaImage string `json:"kafkaImage"`

	// Image used for the kafka deployment (default: docker.io/bitnami/kafka)
	// +optional
	KafkaImageName string `json:"kafkaImageName"`

	// Image tag used for the kafka deployment (default: latest)
	// +optional
	KafkaImageTag string `json:"kafkaImageTag"`

	// Kafka deployment memory limit (default: no limit)
	// +optional
	KafkaMemoryLimit string `json:"kafkaMemoryLimit"`

	// Kafka deployment memory request (default: no limit)
	// +optional
	KafkaMemoryRequest string `json:"kafkaMemoryRequest"`

	// Secret containing the kafka access information, content generated if not provided (default: kafka-secrets)
	// +optional
	KafkaSecret string `json:"kafkaSecret"`

	// Kafka volume size (default: 1Gi)
	// +optional
	KafkaVolumeCapacity string `json:"kafkaVolumeCapacity"`

	// Memcached deployment CPU limit (default: no limit)
	// +optional
	MemcachedCpuLimit string `json:"memcachedCpuLimit"`

	// Memcached deployment CPU request (default: no request)
	// +optional
	MemcachedCpuRequest string `json:"memcachedCpuRequest"`

	// Image string used for the memcached deployment
	// (default: <MemcachedImageName>:<MemcachedImageTag>)
	// +optional
	MemcachedImage string `json:"memcachedImage"`

	// Image used for the memcached deployment (default: manageiq/memcached)
	// +optional
	MemcachedImageName string `json:"memcachedImageName"`

	// Image tag used for the memcached deployment (default: latest)
	// +optional
	MemcachedImageTag string `json:"memcachedImageTag"`

	// Memcached max simultaneous connections (default: 1024)
	// +optional
	MemcachedMaxConnection string `json:"memcachedMaxConnection"`

	// Memcached item memory in megabytes (default: 64)
	// +optional
	MemcachedMaxMemory string `json:"memcachedMaxMemory"`

	// Memcached deployment memory limit (default: no limit)
	// +optional
	MemcachedMemoryLimit string `json:"memcachedMemoryLimit"`

	// Memcached deployment memory request (default: no limit)
	// +optional
	MemcachedMemoryRequest string `json:"memcachedMemoryRequest"`

	// Memcached max item size (default: 1m, min: 1k, max: 1024m)
	// +optional
	MemcachedSlabPageSize string `json:"memcachedSlabPageSize"`

	// Secret containing the trusted CA certificate file(s) for the OIDC server.
	// Only used with the openid-connect authentication type
	// +optional
	OIDCCACertSecret string `json:"oidcCaCertSecret"`

	// Secret name containing the OIDC client id and secret
	// Only used with the openid-connect authentication type
	// +optional
	OIDCClientSecret string `json:"oidcClientSecret"`

	// URL for OIDC authentication introspection
	// Only used with the openid-connect authentication type.
	// If not specified, the operator will attempt to fetch its value from the
	// "introspection_endpoint" field in the Provider metadata at the
	// OIDCProviderURL provided.
	// +optional
	OIDCOAuthIntrospectionURL string `json:"oidcAuthIntrospectionURL"`

	// URL for the OIDC provider
	// Only used with the openid-connect authentication type
	// +optional
	OIDCProviderURL string `json:"oidcProviderURL"`

	// Orchestrator deployment CPU limit (default: no limit)
	// +optional
	OrchestratorCpuLimit string `json:"orchestratorCpuLimit"`

	// Orchestrator deployment CPU request (default: no request)
	// +optional
	OrchestratorCpuRequest string `json:"orchestratorCpuRequest"`

	// Image string used for the orchestrator deployment
	// (default: <OrchestratorImageNamespace>/<OrchestratorImageName>:<OrchestratorImageTag>)
	// +optional
	OrchestratorImage string `json:"orchestratorImage"`

	// Image name used for the orchestrator deployment (default: manageiq-orchestrator)
	// +optional
	OrchestratorImageName string `json:"orchestratorImageName"`

	// Image namespace used for the orchestrator and worker deployments (default: manageiq)
	// +optional
	OrchestratorImageNamespace string `json:"orchestratorImageNamespace"`

	// Image tag used for the orchestrator and worker deployments (default: latest)
	// +optional
	OrchestratorImageTag string `json:"orchestratorImageTag"`

	// Number of seconds to wait before starting the orchestrator liveness check (default: 480)
	// +optional
	OrchestratorInitialDelay string `json:"orchestratorInitialDelay"`

	// Orchestrator deployment memory limit (default: no limit)
	// +optional
	OrchestratorMemoryLimit string `json:"orchestratorMemoryLimit"`

	// Orchestrator deployment memory request (default: no limit)
	// +optional
	OrchestratorMemoryRequest string `json:"orchestratorMemoryRequest"`

	// PostgreSQL deployment CPU limit (default: no limit)
	// +optional
	PostgresqlCpuLimit string `json:"postgresqlCpuLimit"`

	// PostgreSQL deployment CPU request (default: no request)
	// +optional
	PostgresqlCpuRequest string `json:"postgresqlCpuRequest"`

	// Image string used for the postgresql deployment
	// (default: <PostgresqlImageName>:<PostgresqlImageTag>)
	// +optional
	PostgresqlImage string `json:"postgresqlImage"`

	// Image used for the postgresql deployment (Default: docker.io/manageiq/postgresql)
	// +optional
	PostgresqlImageName string `json:"postgresqlImageName"`

	// Image tag used for the postgresql deployment (Default: 10)
	// +optional
	PostgresqlImageTag string `json:"postgresqlImageTag"`

	// PostgreSQL maximum connection setting (default: 1000)
	// +optional
	PostgresqlMaxConnections string `json:"postgresqlMaxConnections"`

	// PostgreSQL deployment memory limit (default: no limit)
	// +optional
	PostgresqlMemoryLimit string `json:"postgresqlMemoryLimit"`

	// PostgreSQL deployment memory request (default: no limit)
	// +optional
	PostgresqlMemoryRequest string `json:"postgresqlMemoryRequest"`

	// PostgreSQL shared buffers setting (default: 1GB)
	// +optional
	PostgresqlSharedBuffers string `json:"postgresqlSharedBuffers"`

	// Server GUID (default: auto-generated)
	// +optional
	ServerGuid string `json:"serverGuid"`

	// StorageClass name that will be used by manageiq data stores
	// +optional
	StorageClassName string `json:"storageClassName"`

	// Secret containing the tls cert and key for the ingress, content generated if not provided (default: tls-secret)
	// +optional
	TLSSecret string `json:"tlsSecret"`

	// Image string used for the UI worker deployments
	// By default this is determined by the orchestrator pod
	// +optional
	UIWorkerImage string `json:"uiWorkerImage"`

	// Image string used for the webserver worker deployments
	// By default this is determined by the orchestrator pod
	// +optional
	WebserverWorkerImage string `json:"webserverWorkerImage"`

	// Zookeeper deployment CPU limit (default: no limit)
	// +optional
	ZookeeperCpuLimit string `json:"zookeeperCpulimit"`

	// Zookeeper deployment CPU request (default: no request)
	// +optional
	ZookeeperCpuRequest string `json:"zookeeperCpuRequest"`

	// Image string used for the zookeeper deployment
	// (default: <ZookeeperImageName>:<ZookeeperImageTag>)
	// +optional
	ZookeeperImage string `json:"zookeeperImage"`

	// Image used for the zookeeper deployment (default: docker.io/bitnami/zookeeper)
	// +optional
	ZookeeperImageName string `json:"zookeeperImageName"`

	// Image tag used for the zookeeper deployment (default: latest)
	// +optional
	ZookeeperImageTag string `json:"zookeeperImageTag"`

	// Zookeeper deployment memory limit (default: no limit)
	// +optional
	ZookeeperMemoryLimit string `json:"zookeeperMemoryLimit"`

	// Zookeeper deployment memory request (default: no limit)
	// +optional
	ZookeeperMemoryRequest string `json:"zookeeperMemoryRequest"`

	// Zookeeper volume size (default: 1Gi)
	// +optional
	ZookeeperVolumeCapacity string `json:"zookeeperVolumeCapacity"`
}

// ManageIQStatus defines the observed state of ManageIQ
type ManageIQStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManageIQ is the Schema for the manageiqs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=manageiqs,scope=Namespaced
type ManageIQ struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManageIQSpec   `json:"spec,omitempty"`
	Status ManageIQStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManageIQList contains a list of ManageIQ
type ManageIQList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManageIQ `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManageIQ{}, &ManageIQList{})
}

func (m *ManageIQ) Validate() error {
	spec := m.Spec
	errs := []string{}

	if spec.HttpdAuthenticationType == "openid-connect" {
		if spec.HttpdAuthConfig != "" && (spec.OIDCProviderURL != "" || spec.OIDCOAuthIntrospectionURL != "" || spec.OIDCClientSecret != "") {
			// Invalid if config and any other info is also provided
			errs = append(errs, "OIDCProviderURL, OIDCOAuthIntrospectionURL, and OIDCClientSecret are invalid when HttpdAuthConfig is specified")
		} else if spec.HttpdAuthConfig == "" && (spec.OIDCProviderURL == "" || spec.OIDCClientSecret == "") {
			// Need to provide either the entire config or a secret and provider url
			errs = append(errs, "HttpdAuthConfig or both OIDCProviderURL and OIDCClientSecret must be provided for openid-connect authentication")
		}
	} else {
		if spec.OIDCProviderURL != "" {
			errs = append(errs, fmt.Sprintf("OIDCProviderURL is not allowed for authentication type %s", spec.HttpdAuthenticationType))
		}

		if spec.OIDCOAuthIntrospectionURL != "" {
			errs = append(errs, fmt.Sprintf("OIDCOAuthIntrospectionURL is not allowed for authentication type %s", spec.HttpdAuthenticationType))
		}

		if spec.OIDCClientSecret != "" {
			errs = append(errs, fmt.Sprintf("OIDCClientSecret is not allowed for authentication type %s", spec.HttpdAuthenticationType))
		}

		if spec.OIDCCACertSecret != "" {
			errs = append(errs, fmt.Sprintf("OIDCCACertSecret is not allowed for authentication type %s", spec.HttpdAuthenticationType))
		}
	}

	if len(errs) > 0 {
		err := fmt.Sprintf("validation failed for ManageIQ object: %s", strings.Join(errs, ", "))
		return errors.New(err)
	} else {
		return nil
	}
}
