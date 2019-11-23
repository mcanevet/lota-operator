package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type terraformProviderAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// LotaProviderSpec defines the desired state of LotaProvider
// +k8s:openapi-gen=true
type LotaProviderSpec struct {
	Name    string                       `json:"name"`
	Version string                       `json:"version"`
	Schema  []terraformProviderAttribute `json:"schema"`
}

// LotaProviderStatus defines the observed state of LotaProvider
// +k8s:openapi-gen=true
type LotaProviderStatus struct {
	Resources []string `json:"resources"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LotaProvider is the Schema for the lotaproviders API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=lotaproviders,scope=Namespaced
type LotaProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LotaProviderSpec   `json:"spec,omitempty"`
	Status LotaProviderStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LotaProviderList contains a list of LotaProvider
type LotaProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LotaProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LotaProvider{}, &LotaProviderList{})
}
