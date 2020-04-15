package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManageIQSpec defines the desired state of ManageIQ
type ManageIQSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

	// Application name used for deployed objects (default: manageiq)
	// +optional
	AppName string `json:"appName"`

	// Initial password for "admin" user (default: smartvm)
	// +optional
	ApplicationAdminPassword string `json:"applicationAdminPassword"`

	// Domain name for the external route. Used for external authentication configuration
	ApplicationDomain string `json:"applicationDomain"`

	// Database region number (default: 0)
	// +optional
	DatabaseRegion string `json:"databaseRegion"`

	// Database volume size (default: 15Gi)
	// +optional
	DatabaseVolumeCapacity string `json:"databaseVolumeCapacity"`

	// Secret containing the database access information, content generated if not provided (default: postgresql-secrets)
	// +optional
	DatabaseSecret string `json:"databaseSecret"`

	// Secret containing the tls cert and key for the ingress, content generated if not provided (default: tls-secret)
	// +optional
	TLSSecret string `json:"tlsSecret"`

	// Image used for the httpd deployment (default: manageiq/httpd)
	// +optional
	HttpdImageName string `json:"httpdImageName"`
	// Image tag used for the httpd deployment (default: latest)
	// +optional
	HttpdImageTag string `json:"httpdImageTag"`

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
	// Memcached max item size (default: 1mb, min: 1k, max: 1024m)
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

	// Orchestrator deployment CPU request (default: no request)
	// +optional
	OrchestratorCpuRequest string `json:"orchestratorCpuRequest"`
	// Orchestrator deployment memory limit (default: no limit)
	// +optional
	OrchestratorMemoryLimit string `json:"orchestratorMemoryLimit"`
	// Orchestrator deployment memory request (default: no limit)
	// +optional
	OrchestratorMemoryRequest string `json:"orchestratorMemoryRequest"`

	// Image used for the postgresql deployment (Default: docker.io/manageiq/postgresql)
	// +optional
	PostgresqlImageName string `json:"postgresqlImageName"`
	// Image tag used for the postgresql deployment (Default: 10)
	// +optional
	PostgresqlImageTag string `json:"postgresqlImageTag"`

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

	// Kafka volume size (default: 1Gi)
	// +optional
	KafkaVolumeCapacity string `json:"kafkaVolumeCapacity"`
	// Zookeeper volume size (default: 1Gi)
	// +optional
	ZookeeperVolumeCapacity string `json:"zookeeperVolumeCapacity"`
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

	if spec.ApplicationAdminPassword == "" {
		spec.ApplicationAdminPassword = "smartvm"
	}

	if spec.DatabaseRegion == "" {
		spec.DatabaseRegion = "0"
	}

	if spec.DatabaseVolumeCapacity == "" {
		spec.DatabaseVolumeCapacity = "15Gi"
	}

	if spec.HttpdImageName == "" {
		spec.HttpdImageName = "manageiq/httpd"
	}

	if spec.HttpdImageTag == "" {
		spec.HttpdImageTag = "latest"
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
		spec.MemcachedSlabPageSize = "1mb"
	}

	if spec.OrchestratorImageName == "" {
		spec.OrchestratorImageName = "manageiq-orchestrator"
	}

	if spec.OrchestratorImageNamespace == "" {
		spec.OrchestratorImageNamespace = "manageiq"
	}

	if spec.OrchestratorImageTag == "" {
		spec.OrchestratorImageTag = "latest"
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

	if spec.KafkaVolumeCapacity == "" {
		spec.KafkaVolumeCapacity = "1Gi"
	}

	if spec.ZookeeperVolumeCapacity == "" {
		spec.ZookeeperVolumeCapacity = "1Gi"
	}
}
