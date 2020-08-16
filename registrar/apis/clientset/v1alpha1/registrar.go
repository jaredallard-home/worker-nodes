package v1alpha1

import (
	"github.com/jaredallard-home/worker-nodes/registrar/apis/types/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type RegistrarClientset struct {
	*kubernetes.Clientset
	registrarV1Alpha1Client RegistrarV1Alpha1Interface
}

func (rc *RegistrarClientset) RegistrarV1Alpha1Client() RegistrarV1Alpha1Interface {
	return rc.registrarV1Alpha1Client
}

type RegistrarV1Alpha1Interface interface {
	Devices(namespace string) DeviceInterface
}

type RegistrarV1Alpha1Client struct {
	client rest.Interface
}

func NewForConfig(c *rest.Config) (*RegistrarClientset, error) {
	k, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	v1alpha1.SchemeBuilder.Register(
		&v1alpha1.DeviceList{},
		&v1alpha1.Device{},
	)

	if err := v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &RegistrarClientset{k, &RegistrarV1Alpha1Client{client: client}}, nil
}

func (c *RegistrarV1Alpha1Client) Devices(namespace string) DeviceInterface {
	return &deviceClient{
		client: c.client,
		ns:     namespace,
	}
}
