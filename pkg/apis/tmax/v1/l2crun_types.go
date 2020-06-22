package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// L2CRunSpec defines the desired state of L2CRun
type L2CRunSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// L2cName is the object name of L2c to be referred
	L2cName string `json:"l2cName"`
}

// L2CRunStatus defines the observed state of L2CRun
type L2CRunStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	//
	StartTime *metav1.Time `json:"startTime,omitempty"`

	//
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	//
	Conditions []L2cRunSHCondition `json:"conditions,omitempty"`

	//
	Phase Phase `json:"phase,omitempty"`

	//
	Status Status `json:"status,omitempty"`

	//
	Message string `json:"message,omitempty"`
}

type L2cRunSHCondition struct {
	//
	Type Phase `json:"type,omitempty"`

	//
	Status Status `json:"status,omitempty"`

	//
	Message string `json:"message,omitempty"`

	//
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// L2CRun is the Schema for the l2cruns API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=l2cruns,scope=Namespaced
type L2CRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   L2CRunSpec   `json:"spec,omitempty"`
	Status L2CRunStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// L2CRunList contains a list of L2CRun
type L2CRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []L2CRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&L2CRun{}, &L2CRunList{})
}
