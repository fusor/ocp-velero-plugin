package clients

import (
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"k8s.io/client-go/discovery"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

// NewCoreClient returns a kubernetes CoreV1Client
func NewCoreClient() (*corev1.CoreV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := corev1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewImageClient returns an openshift ImageV1Client
func NewImageClient() (*imagev1.ImageV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := imagev1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewDiscoveryClient returns a client-go DiscoveryClient
func NewDiscoveryClient() (*discovery.DiscoveryClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewRouteClient returns an openshift RouteV1Client
func NewRouteClient() (*routev1.RouteV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := routev1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewBuildClient returns an openshift BuildV1Client
func NewBuildClient() (*buildv1.BuildV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := buildv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
