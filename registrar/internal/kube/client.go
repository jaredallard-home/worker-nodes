package kube

import (
	"errors"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// New returns the correct kube config to use.
func New() (*rest.Config, error) {
	c, err := rest.InClusterConfig()
	if !errors.Is(err, rest.ErrNotInCluster) || err == nil {
		return c, err
	}

	// if we reached here, we're not running in a kubernetes cluster
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{CurrentContext: "dev-environment"}
	cli := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
	return cli.ClientConfig()
}
