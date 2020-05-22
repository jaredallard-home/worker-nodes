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
	_ DeviceInterface = &deviceClient{}
)

type DeviceInterface interface {
	List(context.Context, metav1.ListOptions) (*v1alpha1.DeviceList, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Device, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(context.Context, metav1.DeleteOptions, metav1.ListOptions) error
	Update(context.Context, *v1alpha1.Device) (*v1alpha1.Device, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Device, err error)
	Create(context.Context, *v1alpha1.Device, metav1.CreateOptions) (*v1alpha1.Device, error)
	Watch(context.Context, metav1.ListOptions) (watch.Interface, error)
}

type deviceClient struct {
	client rest.Interface
	ns     string
}

// List returns all devices in a namespace
func (c *deviceClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.DeviceList, error) {
	result := v1alpha1.DeviceList{}
	err := c.client.
		Get().
		Namespace(c.ns).
		Resource("devices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Get returns a given device by it's name
func (c *deviceClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Device, error) {
	result := v1alpha1.Device{}
	err := c.client.
		Get().
		Namespace(c.ns).
		Resource("devices").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Delete takes name of the device and deletes it. Returns an error if one occurs.
func (c *deviceClient) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("devices").
		Name(name).
		Body(options).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *deviceClient) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("devices").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do(ctx).
		Error()
}

// Update takes the representation of a device and updates it. Returns the server's representation of the device, and an error, if there is any.
func (c *deviceClient) Update(ctx context.Context, d *v1alpha1.Device) (result *v1alpha1.Device, err error) {
	result = &v1alpha1.Device{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("devices").
		Name(d.Name).
		Body(d).
		Do(ctx).
		Into(result)
	return
}

// Patch applies the patch and returns the patched device.
func (c *deviceClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Device, err error) {
	result = &v1alpha1.Device{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("devices").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Create creates a device
func (c *deviceClient) Create(ctx context.Context, device *v1alpha1.Device, opts metav1.CreateOptions) (*v1alpha1.Device, error) {
	result := v1alpha1.Device{}
	err := c.client.
		Post().
		Namespace(c.ns).
		Resource("devices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(device).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Watch creates a watch that will return devices when they are modified
func (c *deviceClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.
		Get().
		Namespace(c.ns).
		Resource("devices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}
