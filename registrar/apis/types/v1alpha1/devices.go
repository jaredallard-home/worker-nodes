package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type DeviceSpec struct {
	// SecretRef is the secret that contains all secret information for
	// this device
	// +kubebuilder:validation:Required
	SecretRef string `json:"secretRef"`
	// +kubebuilder:validation:Required
	WireguardIPRef string `json:"wireguardIPRef"`
}

type DeviceStatus struct {
	// Registered denotes wether or not this device is considered as
	// being registered or not.
	Registered bool `json:"registered"`

	// PublicKey is the wireguard public key for this device
	PublicKey string `json:"publicKey"`
}

// +kubebuilder:object:root=true
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceSpec   `json:"spec"`
	Status DeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Device `json:"items"`
}
