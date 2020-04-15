package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManageiqSpec defines the desired state of Manageiq
type ManageiqSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Application name used for deployed objects
	AppName string `json:"appName"`

	// admin user initial password
	ApplicationAdminPassword string `json:"applicationAdminPassword"`

	// Domain name for the external route
	// Used for external authentication configuration
	ApplicationDomain string `json:"applicationDomain"`

	// Application region number
	DatabaseRegion string `json:"databaseRegion"`
	// Containerized database volume size
	DatabaseVolumeCapacity string `json:"databaseVolumeCapacity"`
	// +optional
	DatabaseSecret string `json:"databaseSecret"`

	// Secret containing the tls cert and key for the ingress
	TLSSecret string `json:"tlsSecret"`

	HttpdCpuRequest    string `json:"httpdCpuRequest"`
	HttpdImageName     string `json:"httpdImageName"`
	HttpdImageTag      string `json:"httpdImageTag"`
	HttpdMemoryLimit   string `json:"httpdMemoryLimit"`
	HttpdMemoryRequest string `json:"httpdMemoryRequest"`

	MemcachedCpuRequest    string `json:"memcachedCpuRequest"`
	MemcachedImageName     string `json:"memcachedImageName"`
	MemcachedImageTag      string `json:"memcachedImageTag"`
	MemcachedMaxConnection string `json:"memcachedMaxConnection"`
	MemcachedMaxMemory     string `json:"memcachedMaxMemory"`
	MemcachedMemoryLimit   string `json:"memcachedMemoryLimit"`
	MemcachedMemoryRequest string `json:"memcachedMemoryRequest"`
	MemcachedSlabPageSize  string `json:"memcachedSlabPageSize"`

	OrchestratorCpuRequest     string `json:"orchestratorCpuRequest"`
	OrchestratorImageName      string `json:"orchestratorImageName"`
	OrchestratorImageNamespace string `json:"orchestratorImageNamespace"`
	OrchestratorImageTag       string `json:"orchestratorImageTag"`
	OrchestratorMemoryLimit    string `json:"orchestratorMemoryLimit"`
	OrchestratorMemoryRequest  string `json:"orchestratorMemoryRequest"`

	PostgresqlCpuRequest     string `json:"postgresqlCpuRequest"`
	PostgresqlImageName      string `json:"postgresqlImageName"`
	PostgresqlImageTag       string `json:"postgresqlImageTag"`
	PostgresqlMaxConnections string `json:"postgresqlMaxConnections"`
	PostgresqlMemoryLimit    string `json:"postgresqlMemoryLimit"`
	PostgresqlMemoryRequest  string `json:"postgresqlMemoryRequest"`
	PostgresqlSharedBuffers  string `json:"postgresqlSharedBuffers"`

	KafkaVolumeCapacity     string `json:"kafkaVolumeCapacity"`
	ZookeeperVolumeCapacity string `json:"zookeeperVolumeCapacity"`
	// +optional
	KafkaSecret string `json:"kafkaSecret"`

	EncryptionKey string `json:"encryptionKey"`
}

// ManageiqStatus defines the observed state of Manageiq
type ManageiqStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Manageiq is the Schema for the manageiqs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=manageiqs,scope=Namespaced
type Manageiq struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManageiqSpec   `json:"spec,omitempty"`
	Status ManageiqStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManageiqList contains a list of Manageiq
type ManageiqList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Manageiq `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Manageiq{}, &ManageiqList{})
}
