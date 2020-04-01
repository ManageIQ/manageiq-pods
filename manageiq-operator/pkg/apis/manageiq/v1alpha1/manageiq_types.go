package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// ManageiqSpec defines the desired state of Manageiq
// +k8s:openapi-gen=true
type ManageiqSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	// Application name used for deployed objects
	AppName string `json:"appName"`

	// admin user initial password
	ApplicationAdminPassword string `json:"applicationAdminPassword"`

	// Domain name for the external route
	// Used for external authentication configuration
	ApplicationDomain string `json:"applicationDomain"`

	DatabaseName     string `json:"databaseName"`
	DatabasePort     string `json:"databasePort"`
	DatabaseUser     string `json:"databaseUser"`
	DatabasePassword string `json:"databasePassword"`
	// Application region number
	DatabaseRegion string `json:"databaseRegion"`
	// Containerized database volume size
	DatabaseVolumeCapacity string `json:"databaseVolumeCapacity"`

	HttpdCpuRequest    string `json:"httpdCpuRequest"`
	HttpdImageName     string `json:"httpdImageName"`
	HttpdImageTag      string `json:"httpdImageTag"`
	HttpdMemoryLimit   string `json:"httpdMemoryLimit"`
	HttpdMemoryRequest string `json:"httpdMemoryRequest"`

	// memcachedpod deployment information
	MemcachedCpuRequest    string `json:"memcachedCpuRequest"`
	MemcachedImageName     string `json:"memcachedImageName"`
	MemcachedImageTag      string `json:"memcachedImageTag"`
	MemcachedMaxConnection string `json:"memcachedMaxConnection"`
	MemcachedMaxMemory     string `json:"memcachedMaxMemory"`
	MemcachedMemoryLimit   string `json:"memcachedMemoryLimit"`
	MemcachedMemoryRequest string `json:"memcachedMemoryRequest"`
	MemcachedSlabPageSize  string `json:"memcachedSlabPageSize"`

	//  orchestrator deployment information
	OrchestratorCpuRequest     string `json:"orchestratorCpuRequest"`
	OrchestratorImageName      string `json:"orchestratorImageName"`
	OrchestratorImageNamespace string `json:"orchestratorImageNamespace"`
	OrchestratorImageTag       string `json:"orchestratorImageTag"`
	OrchestratorMemoryLimit    string `json:"orchestratorMemoryLimit"`
	OrchestratorMemoryRequest  string `json:"orchestratorMemoryRequest"`

	// postgres database pod deployment information
	PostgresqlCpuRequest     string `json:"postgresqlCpuRequest"`
	PostgresqlImageName      string `json:"postgresqlImageName"`
	PostgresqlImageTag       string `json:"postgresqlImageTag"`
	PostgresqlMaxConnections string `json:"postgresqlMaxConnections"`
	PostgresqlMemoryLimit    string `json:"postgresqlMemoryLimit"`
	PostgresqlMemoryRequest  string `json:"postgresqlMemoryRequest"`
	PostgresqlSharedBuffers  string `json:"postgresqlSharedBuffers"`

	EncryptionKey string `json:"encryptionKey"`
}

// ManageiqStatus defines the observed state of Manageiq
// +k8s:openapi-gen=true
type ManageiqStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	HttpdIsSetup        bool `json:"httpdIssetup"`
	MemcachedIsSetup    bool `json:"memcachedIssetup"`
	PostgresqlIsSetup   bool `json:"postgresqlIssetup"`
	OrchestratorIsSetup bool `json:"orchestratorISSetup"`
	NetworkIsSetup      bool `json:"networkIssetup"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Manageiq is the Schema for the manageiqs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
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
