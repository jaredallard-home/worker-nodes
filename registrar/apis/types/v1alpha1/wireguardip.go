package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type WireguardIPSpec struct {
	// DeviceRef is the device this IP has been assigned to.
	DeviceRef string `json:"deviceRef"`

	// IPAddress of this Wireguard IP Address. Equal to the name.
	IPAdress string `json:"ipAddress"`
}

type WireguardIPStatus struct {
	// Active denotes if this IP is being used.
	Active bool `json:"active"`
}

// +kubebuilder:object:root=true
type WireguardIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WireguardIPSpec   `json:"spec"`
	Status WireguardIPStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type WireguardIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WireguardIP `json:"items"`
}
