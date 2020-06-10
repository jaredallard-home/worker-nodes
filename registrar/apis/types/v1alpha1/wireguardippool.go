package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type WireguardIPPoolSpec struct {
	// CIDR for this block.
	CIDR string `json:"cidr"`
}

type WireguardIPPoolStatus struct {
	// Created denotes if this pool has been created and is active
	Created bool `json:"created"`

	// SecretRef is the secret that contains the wireguard private key
	// that powers that IP pool
	SecretRef string `json:"secretRef"`

	// PublicKey is the key of the server powering this IP Pool
	PublicKey string `json:"publicKey"`

	// UsedAddresses is the number of IP address currently being used
	// in this pool
	UsedAddresses int `json:"usedAddresses"`
}

// +kubebuilder:object:root=true
type WireguardIPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WireguardIPPoolSpec   `json:"spec"`
	Status WireguardIPPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type WireguardIPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WireguardIPPool `json:"items"`
}
