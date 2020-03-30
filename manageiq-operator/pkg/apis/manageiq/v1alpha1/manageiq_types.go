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

	DatabaseName string `json:"databaseName"`
	DatabasePort string `json:"databasePort"`
	DatabaseUser string `json:"databaseUser"`
	DatabasePassword string `json:"databasePassword"`
	// Application region number
	DatabaseRegion string `json:"databaseRegion"`
	// Containerized database volume size
	DatabaseVolumeCapacity string `json:"databaseVolumeCapacity"`

	HttpdCPUReq string `json:"httpdCPUReq"`
	HttpdImageName string `json:"httpdImageName"`
	HttpdImageTag string `json:"httpdImageTag"`
	HttpdMemLimit string `json:"httpdMemLimit"`
	HttpdMemReq string `json:"httpdMemReq"`


	// memcachedpod deployment information
	MemcachedCPUReq string `json:"memcachedCPUReq"`
	MemcachedImageName string `json:"memcachedImageName"`
	MemcachedImageTag string `json:"memcachedImageTag"` 
	MemcachedMaxConnection string `json:"memcachedMaxConnection"`
	MemcachedMaxMemory string `json:"memcachedMaxMemory"`
	MemcachedMemLimit string `json:"memcachedMemLimit"`
	MemcachedMemReq string `json:"memcachedMemReq"`
	MemcachedSlabPageSize string `json:"memcachedSlabPageSize"`

	//  orchestrator deployment information
	OrchestratorCPUReq string `json:"orchestratorCPUReq"`
	OrchestratorImageName string `json:"orchestratorImageName"`
	OrchestratorImageNamespace string `json:"orchestratorImageNamespace"`
	OrchestratorImageTag string `json:"orchestratorImageTag"`
	OrchestratorMemLimit string `json:"orchestratorMemLimit"`
	OrchestratorMemReq string `json:"orchestratorMemReq"`

	// postgres database pod deployment information
	PostgresqlCPUReq string `json:"postgresqlCPUReq"`
	PostgresqlImgName string `json:"postgresqlImgName"`
	PostgresqlImgTag string `json:"postgresqlImgTag"`
	PostgresqlMaxConnections string `json:"postgresqlMaxConnections"` 
	PostgresqlMemLimit string `json:"postgresqlMemLimit"` 
	PostgresqlMemReq string `json:"postgresqlMemReq"`
	PostgresqlSharedBuffers string `json:"postgresqlSharedBuffers"`

	EncryptionKey string `json:"encryptionKey"`



}

// ManageiqStatus defines the observed state of Manageiq
// +k8s:openapi-gen=true
type ManageiqStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	HttpdIsSetup bool `json:"httpdIssetup"`
	MemcachedIsSetup bool `json:"memcachedIssetup"`
	PostgresqlIsSetup bool`json:"postgresqlIssetup"`
	OrchestratorIsSetup bool `json:"orchestratorISSetup"`
	NetworkIsSetup bool `json:"networkIssetup"`
	
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
