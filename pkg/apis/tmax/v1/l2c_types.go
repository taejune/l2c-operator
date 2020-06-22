package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// L2CSpec defines the desired state of L2C
type L2CSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	ProjectName string `json:"projectName"`

	AccessCode string `json:"accessCode"`

	GitUrl string `json:"gitUrl"`

	GitRevision string `json:"gitRevision"`

	ImageUrl string `json:"imageUrl"`

	ImageRegSecret string `json:"imageRegistrySecretName,omitempty"`

	// +kubebuilder:validation:Enum=wildfly
	WasSourceType string `json:"wasSourceType"`

	// +kubebuilder:validation:Enum=jeus
	WasTargetType string `json:"wasTargetType"`

	WasPort int32 `json:"wasPort"`

	// +kubebuilder:validation:Enum=ClusterIP;LoadBalancer;NodePort
	WasServiceType string `json:"wasServiceType,omitempty"`

	WasPackageServer string `json:"wasPackageServerUrl,omitempty"`

	DbMigrate bool `json:"dbMigrate,omitempty"`

	// +kubebuilder:validation:Enum=TIBERO
	DbTargetType string `json:"dbTargetType,omitempty"`

	DbTargetStorageSize string `json:"dbTargetStorageSize,omitempty"`

	// +kubebuilder:validation:Enum=ClusterIP;LoadBalancer;NodePort
	DbTargetServieceType string `json:"dbTargetServiceType,omitempty"`

	DbTargetUser string `json:"dbTargetUser,omitempty"`

	DbTargetPassword string `json:"dbTargetPassword,omitempty"`

	// +kubebuilder:validation:Enum=ORACLE
	DbSourceType string `json:"dbSourceType,omitempty"`

	// +kubebuilder:validation:Pattern=[0-9]*\.[0-9]*\.[0-9]*\.[0-9]*
	DbSourceIp string `json:"dbSourceIp,omitempty"`

	DbSourcePort int32 `json:"dbSourcePort,omitempty"`

	DbSourceUser string `json:"dbSourceUser,omitempty"`

	DbSourcePassword string `json:"dbSourcePassword,omitempty"`

	DbSourceSid string `json:"dbSourceSid,omitempty"`
}

// L2CStatus defines the observed state of L2C
type L2CStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	//
	Status Status `json:"status,omitempty"`

	//
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// L2C is the Schema for the l2cs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=l2cs,scope=Namespaced
type L2C struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   L2CSpec   `json:"spec,omitempty"`
	Status L2CStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// L2CList contains a list of L2C
type L2CList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []L2C `json:"items"`
}

func init() {
	SchemeBuilder.Register(&L2C{}, &L2CList{})
}
