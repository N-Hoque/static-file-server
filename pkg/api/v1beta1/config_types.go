package v1beta1

import (
	"github.com/N-Hoque/static-file-server/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Host",type="string",JSONPath=".spec.host"
// +kubebuilder:printcolumn:name="Port",type="string",JSONPath=".spec.port"
// +kubebuilder:printcolumn:name="Cors",type="boolean",JSONPath=".spec.cors"
// +kubebuilder:printcolumn:name="TLS Version",type="string",JSONPath=".spec.tls-min-vers"

// SfsConfig is the Schema for the Config API
type SfsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   config.Config   `json:"spec,omitempty"`
	Status SfsConfigStatus `json:"status,omitempty"`
}

// SfsConfigStatus defines the observed state of SfsConfig
type SfsConfigStatus struct{}

func init() {
	SchemeBuilder.Register(&SfsConfig{})
}
