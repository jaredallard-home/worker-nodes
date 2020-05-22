package v1alpha1

import (
	"context"
	"time"

	"github.com/jaredallard-home/worker-nodes/registrar/apis/types/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// verify we satisfy the interface on compile time
var (
	_ WireguardIPPoolInterface = &wireguardIPPoolClient{}
)

type WireguardIPPoolInterface interface {
	List(context.Context, metav1.ListOptions) (*v1alpha1.WireguardIPPoolList, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.WireguardIPPool, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(context.Context, metav1.DeleteOptions, metav1.ListOptions) error
	Update(context.Context, *v1alpha1.WireguardIPPool) (*v1alpha1.WireguardIPPool, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.WireguardIPPool, err error)
	Create(context.Context, *v1alpha1.WireguardIPPool, metav1.CreateOptions) (*v1alpha1.WireguardIPPool, error)
	Watch(context.Context, metav1.ListOptions) (watch.Interface, error)
}

type wireguardIPPoolClient struct {
	client rest.Interface
	ns     string
}

// List returns all wireguards in a namespace
func (c *wireguardIPPoolClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.WireguardIPPoolList, error) {
	result := v1alpha1.WireguardIPPoolList{}
	err := c.client.
		Get().
		Namespace(c.ns).
		Resource("wireguardippools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Get returns a given wireguard by it's name
func (c *wireguardIPPoolClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.WireguardIPPool, error) {
	result := v1alpha1.WireguardIPPool{}
	err := c.client.
		Get().
		Namespace(c.ns).
		Resource("wireguardippools").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Delete takes name of the wireguard and deletes it. Returns an error if one occurs.
func (c *wireguardIPPoolClient) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("wireguardippools").
		Name(name).
		Body(options).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *wireguardIPPoolClient) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("wireguardippools").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do(ctx).
		Error()
}

// Update takes the representation of a wireguard and updates it. Returns the server's representation of the wireguard, and an error, if there is any.
func (c *wireguardIPPoolClient) Update(ctx context.Context, w *v1alpha1.WireguardIPPool) (result *v1alpha1.WireguardIPPool, err error) {
	result = &v1alpha1.WireguardIPPool{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("wireguardippools").
		Name(w.Name).
		Body(w).
		Do(ctx).
		Into(result)
	return
}

// Patch applies the patch and returns the patched wireguard.
func (c *wireguardIPPoolClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.WireguardIPPool, err error) {
	result = &v1alpha1.WireguardIPPool{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("wireguardippools").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Create creates a wireguard
func (c *wireguardIPPoolClient) Create(ctx context.Context, w *v1alpha1.WireguardIPPool, opts metav1.CreateOptions) (*v1alpha1.WireguardIPPool, error) {
	result := v1alpha1.WireguardIPPool{}
	err := c.client.
		Post().
		Namespace(c.ns).
		Resource("wireguardippools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(w).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Watch creates a watch that will return wireguards when they are modified
func (c *wireguardIPPoolClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.
		Get().
		Namespace(c.ns).
		Resource("wireguardippools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}
