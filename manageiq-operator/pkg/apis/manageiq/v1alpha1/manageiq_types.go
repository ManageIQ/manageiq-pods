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

	// Application name used for deployed objects (default: manageiq)
	// +optional
	AppName string `json:"appName"`

	// Domain name for the external route. Used for external authentication configuration
	ApplicationDomain string `json:"applicationDomain"`

	// Database region number (default: 0)
	// +optional
	DatabaseRegion string `json:"databaseRegion"`

	// Group name to create with the super admin role.
	// This can be used to seed a group when using external authentication
	// +optional
	InitialAdminGroupName string `json:"initialAdminGroupName"`

	// Flag to trigger worker resource constraint enforcement (default: false)
	// +optional
	EnforceWorkerResourceConstraints *bool `json:"enforceWorkerResourceConstraints"`

	// Database volume size (default: 15Gi)
	// +optional
	DatabaseVolumeCapacity string `json:"databaseVolumeCapacity"`

	// Secret containing the database access information, content generated if not provided (default: postgresql-secrets)
	// +optional
	DatabaseSecret string `json:"databaseSecret"`

	// Secret containing the tls cert and key for the ingress, content generated if not provided (default: tls-secret)
	// +optional
	TLSSecret string `json:"tlsSecret"`

	// Secret containing the image registry authentication information needed for the manageiq images
	// +optional
	ImagePullSecret string `json:"imagePullSecret"`

	// StorageClass name that will be used by manageiq data stores
	// +optional
	StorageClassName string `json:"storageClassName"`

	// Image namespace used for the httpd deployment (default: manageiq)
	// Note: the exact image will be determined by the authentication method selected
	// +optional
	HttpdImageNamespace string `json:"httpdImageNamespace"`
	// Image tag used for the httpd deployment (default: latest)
	// +optional
	HttpdImageTag string `json:"httpdImageTag"`

	// Type of httpd authentication (default: internal)
	// Options: internal, external, active-directory, saml, openid-connect
	// Note: external, active-directory, and saml require an httpd container with elevated privileges
	// +optional
	HttpdAuthenticationType string `json:"httpdAuthenticationType"`
	// URL for the OIDC provider
	// Only used with the openid-connect authentication type
	// +optional
	OIDCProviderURL string `json:"oidcProviderURL"`
	// Secret containing the trusted CA certificate file(s) for the OIDC server.
	// Only used with the openid-connect authentication type
	// +optional
	OIDCCACertSecret string `json:"oidcCaCertSecret"`
	// URL for OIDC authentication introspection
	// Only used with the openid-connect authentication type.
	// If not specified, the operator will attempt to fetch its value from the
	// "introspection_endpoint" field in the Provider metadata at the
	// OIDCProviderURL provided.
	// +optional
	OIDCOAuthIntrospectionURL string `json:"oidcAuthIntrospectionURL"`
	// Secret name containing the OIDC client id and secret
	// Only used with the openid-connect authentication type
	// +optional
	OIDCClientSecret string `json:"oidcClientSecret"`
	// Secret containing the httpd configuration files
	// Mutually exclusive with the OIDCClientSecret and OIDCProviderURL if using openid-connect
	// +optional
	HttpdAuthConfig string `json:"httpdAuthConfig"`
	// Flag to enable SSO in the application (default: false)
	// +optional
	EnableSSO *bool `json:"enableSSO"`
	// Flag to allow logging into the application without SSO (default: true)
	// +optional
	EnableApplicationLocalLogin *bool `json:"enableApplicationLocalLogin"`

	// Httpd deployment CPU limit (default: no limit)
	// +optional
	HttpdCpuLimit string `json:"httpdCpuLimit"`
	// Httpd deployment CPU request (default: no request)
	// +optional
	HttpdCpuRequest string `json:"httpdCpuRequest"`
	// Httpd deployment memory limit (default: no limit)
	// +optional
	HttpdMemoryLimit string `json:"httpdMemoryLimit"`
	// Httpd deployment memory request (default: no limit)
	// +optional
	HttpdMemoryRequest string `json:"httpdMemoryRequest"`

	// Image used for the memcached deployment (default: manageiq/memcached)
	// +optional
	MemcachedImageName string `json:"memcachedImageName"`
	// Image tag used for the memcached deployment (default: latest)
	// +optional
	MemcachedImageTag string `json:"memcachedImageTag"`

	// Memcached deployment CPU limit (default: no limit)
	// +optional
	MemcachedCpuLimit string `json:"memcachedCpuLimit"`
	// Memcached deployment CPU request (default: no request)
	// +optional
	MemcachedCpuRequest string `json:"memcachedCpuRequest"`
	// Memcached deployment memory limit (default: no limit)
	// +optional
	MemcachedMemoryLimit string `json:"memcachedMemoryLimit"`
	// Memcached deployment memory request (default: no limit)
	// +optional
	MemcachedMemoryRequest string `json:"memcachedMemoryRequest"`

	// Memcached max simultaneous connections (default: 1024)
	// +optional
	MemcachedMaxConnection string `json:"memcachedMaxConnection"`
	// Memcached item memory in megabytes (default: 64)
	// +optional
	MemcachedMaxMemory string `json:"memcachedMaxMemory"`
	// Memcached max item size (default: 1m, min: 1k, max: 1024m)
	// +optional
	MemcachedSlabPageSize string `json:"memcachedSlabPageSize"`

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

	// Orchestrator deployment CPU limit (default: no limit)
	// +optional
	OrchestratorCpuLimit string `json:"orchestratorCpuLimit"`
	// Orchestrator deployment CPU request (default: no request)
	// +optional
	OrchestratorCpuRequest string `json:"orchestratorCpuRequest"`
	// Orchestrator deployment memory limit (default: no limit)
	// +optional
	OrchestratorMemoryLimit string `json:"orchestratorMemoryLimit"`
	// Orchestrator deployment memory request (default: no limit)
	// +optional
	OrchestratorMemoryRequest string `json:"orchestratorMemoryRequest"`

	// Image string used for the base worker deployments
	// By default this is determined by the orchestrator pod
	// +optional
	BaseWorkerImage string `json:"baseWorkerImage"`
	// Image string used for the webserver worker deployments
	// By default this is determined by the orchestrator pod
	// +optional
	WebserverWorkerImage string `json:"webserverWorkerImage"`
	// Image string used for the UI worker deployments
	// By default this is determined by the orchestrator pod
	// +optional
	UIWorkerImage string `json:"uiWorkerImage"`

	// Image used for the postgresql deployment (Default: docker.io/manageiq/postgresql)
	// +optional
	PostgresqlImageName string `json:"postgresqlImageName"`
	// Image tag used for the postgresql deployment (Default: 10)
	// +optional
	PostgresqlImageTag string `json:"postgresqlImageTag"`

	// PostgreSQL deployment CPU limit (default: no limit)
	// +optional
	PostgresqlCpuLimit string `json:"postgresqlCpuLimit"`
	// PostgreSQL deployment CPU request (default: no request)
	// +optional
	PostgresqlCpuRequest string `json:"postgresqlCpuRequest"`
	// PostgreSQL deployment memory limit (default: no limit)
	// +optional
	PostgresqlMemoryLimit string `json:"postgresqlMemoryLimit"`
	// PostgreSQL deployment memory request (default: no limit)
	// +optional
	PostgresqlMemoryRequest string `json:"postgresqlMemoryRequest"`

	// PostgreSQL maximum connection setting (default: 1000)
	// +optional
	PostgresqlMaxConnections string `json:"postgresqlMaxConnections"`
	// PostgreSQL shared buffers setting (default: 1GB)
	// +optional
	PostgresqlSharedBuffers string `json:"postgresqlSharedBuffers"`

	// Flag to indicate if Kafka and Zookeeper should be deployed (default: false)
	// +optional
	DeployMessagingService *bool `json:"deployMessagingService"`

	// Image used for the kafka deployment (default: docker.io/bitnami/kafka)
	// +optional
	KafkaImageName string `json:"kafkaImageName"`
	// Image tag used for the kafka deployment (default: latest)
	// +optional
	KafkaImageTag string `json:"kafkaImageTag"`
	// Kafka volume size (default: 1Gi)
	// +optional
	KafkaVolumeCapacity string `json:"kafkaVolumeCapacity"`
	// Kafka deployment CPU limit (default: no limit)
	// +optional
	KafkaCpuLimit string `json:"kafkaCpulimit"`
	// Kafka deployment CPU request (default: no request)
	// +optional
	KafkaCpuRequest string `json:"kafkaCpuRequest"`
	// Kafka deployment memory limit (default: no limit)
	// +optional
	KafkaMemoryLimit string `json:"kafkaMemoryLimit"`
	// Kafka deployment memory request (default: no limit)
	// +optional
	KafkaMemoryRequest string `json:"kafkaMemoryRequest"`

	// Image used for the zookeeper deployment (default: docker.io/bitnami/zookeeper)
	// +optional
	ZookeeperImageName string `json:"zookeeperImageName"`
	// Image tag used for the zookeeper deployment (default: latest)
	// +optional
	ZookeeperImageTag string `json:"zookeeperImageTag"`
	// Zookeeper volume size (default: 1Gi)
	// +optional
	ZookeeperVolumeCapacity string `json:"zookeeperVolumeCapacity"`
	// Zookeeper deployment CPU limit (default: no limit)
	// +optional
	ZookeeperCpuLimit string `json:"zookeeperCpulimit"`
	// Zookeeper deployment CPU request (default: no request)
	// +optional
	ZookeeperCpuRequest string `json:"zookeeperCpuRequest"`
	// Zookeeper deployment memory limit (default: no limit)
	// +optional
	ZookeeperMemoryLimit string `json:"zookeeperMemoryLimit"`
	// Zookeeper deployment memory request (default: no limit)
	// +optional
	ZookeeperMemoryRequest string `json:"zookeeperMemoryRequest"`

	// Secret containing the kafka access information, content generated if not provided (default: kafka-secrets)
	// +optional
	KafkaSecret string `json:"kafkaSecret"`
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

func (m *ManageIQ) Initialize() {
	spec := &m.Spec

	if spec.AppName == "" {
		spec.AppName = "manageiq"
	}

	if spec.DatabaseRegion == "" {
		spec.DatabaseRegion = "0"
	}

	if spec.EnforceWorkerResourceConstraints == nil {
		spec.EnforceWorkerResourceConstraints = new(bool)
	}

	if spec.DatabaseVolumeCapacity == "" {
		spec.DatabaseVolumeCapacity = "15Gi"
	}

	if spec.HttpdImageNamespace == "" {
		spec.HttpdImageNamespace = "manageiq"
	}

	if spec.HttpdImageTag == "" {
		spec.HttpdImageTag = "latest"
	}

	if spec.HttpdAuthenticationType == "" {
		spec.HttpdAuthenticationType = "internal"
	}

	if spec.EnableSSO == nil {
		spec.EnableSSO = new(bool)
	}

	if spec.EnableApplicationLocalLogin == nil {
		t := true
		spec.EnableApplicationLocalLogin = &t
	}

	if spec.MemcachedImageName == "" {
		spec.MemcachedImageName = "manageiq/memcached"
	}

	if spec.MemcachedImageTag == "" {
		spec.MemcachedImageTag = "latest"
	}

	if spec.MemcachedMaxConnection == "" {
		spec.MemcachedMaxConnection = "1024"
	}

	if spec.MemcachedMaxMemory == "" {
		spec.MemcachedMaxMemory = "64"
	}

	if spec.MemcachedSlabPageSize == "" {
		spec.MemcachedSlabPageSize = "1m"
	}

	if spec.OrchestratorImageName == "" {
		spec.OrchestratorImageName = "manageiq-orchestrator"
	}

	if spec.OrchestratorImageNamespace == "" {
		spec.OrchestratorImageNamespace = "manageiq"
	}

	if spec.OrchestratorImageTag == "" {
		spec.OrchestratorImageTag = "latest-jansa"
	}

	if spec.OrchestratorInitialDelay == "" {
		spec.OrchestratorInitialDelay = "480"
	}

	if spec.PostgresqlImageName == "" {
		spec.PostgresqlImageName = "docker.io/manageiq/postgresql"
	}

	if spec.PostgresqlImageTag == "" {
		spec.PostgresqlImageTag = "10"
	}

	if spec.PostgresqlMaxConnections == "" {
		spec.PostgresqlMaxConnections = "1000"
	}

	if spec.PostgresqlSharedBuffers == "" {
		spec.PostgresqlSharedBuffers = "1GB"
	}

	if spec.DeployMessagingService == nil {
		spec.DeployMessagingService = new(bool)
	}

	if spec.KafkaImageName == "" {
		spec.KafkaImageName = "docker.io/bitnami/kafka"
	}

	if spec.KafkaImageTag == "" {
		spec.KafkaImageTag = "latest"
	}

	if spec.KafkaVolumeCapacity == "" {
		spec.KafkaVolumeCapacity = "1Gi"
	}

	if spec.ZookeeperImageName == "" {
		spec.ZookeeperImageName = "docker.io/bitnami/zookeeper"
	}

	if spec.ZookeeperImageTag == "" {
		spec.ZookeeperImageTag = "latest"
	}

	if spec.ZookeeperVolumeCapacity == "" {
		spec.ZookeeperVolumeCapacity = "1Gi"
	}
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
