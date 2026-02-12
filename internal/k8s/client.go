// Package k8s provides Kubernetes client wrapper using client-go dynamic client.
// It supports both in-cluster and out-of-cluster kubeconfig loading.
package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes dynamic client for flexible resource operations.
type Client struct {
	dynamicClient dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	restMapper    meta.RESTMapper
	config        *rest.Config
}

// NewClient creates a new Kubernetes client.
// It automatically detects in-cluster config or loads from kubeconfig.
func NewClient() (*Client, error) {
	config, err := loadKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Create REST mapper for GVK -> GVR conversion
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get API group resources: %w", err)
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	return &Client{
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		restMapper:      restMapper,
		config:          config,
	}, nil
}

// loadKubeConfig loads kubeconfig with the following precedence:
// 1. In-cluster config (if running inside a pod)
// 2. KUBECONFIG environment variable
// 3. ~/.kube/config (default location)
func loadKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfigPath, err)
	}

	return config, nil
}

// GetResource retrieves a single Kubernetes resource by namespace, kind, and name.
// Returns an unstructured object for flexible handling of any resource type.
func (c *Client) GetResource(ctx context.Context, namespace, kind, name string) (*unstructured.Unstructured, error) {
	if kind == "" {
		return nil, fmt.Errorf("kind is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required for resource kind=%s", kind)
	}

	gvr, namespaced, err := c.kindToGVR(kind)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve kind=%s: %w", kind, err)
	}

	// Validate namespace requirement
	if namespaced && namespace == "" {
		return nil, fmt.Errorf("namespace is required for namespaced resource kind=%s", kind)
	}
	if !namespaced && namespace != "" {
		return nil, fmt.Errorf("namespace should not be specified for cluster-scoped resource kind=%s", kind)
	}

	var resource dynamic.ResourceInterface
	if namespaced {
		resource = c.dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resource = c.dynamicClient.Resource(gvr)
	}

	obj, err := resource.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource namespace=%s kind=%s name=%s: %w",
			namespace, kind, name, err)
	}

	return obj, nil
}

// ListResources lists Kubernetes resources by namespace, kind, and optional label selector.
// Returns a list of unstructured objects.
func (c *Client) ListResources(ctx context.Context, namespace, kind, labelSelector string) (*unstructured.UnstructuredList, error) {
	if kind == "" {
		return nil, fmt.Errorf("kind is required")
	}

	gvr, namespaced, err := c.kindToGVR(kind)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve kind=%s: %w", kind, err)
	}

	// Validate namespace requirement
	if namespaced && namespace == "" {
		return nil, fmt.Errorf("namespace is required for namespaced resource kind=%s", kind)
	}
	if !namespaced && namespace != "" {
		return nil, fmt.Errorf("namespace should not be specified for cluster-scoped resource kind=%s", kind)
	}

	var resource dynamic.ResourceInterface
	if namespaced {
		resource = c.dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resource = c.dynamicClient.Resource(gvr)
	}

	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	list, err := resource.List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources namespace=%s kind=%s labelSelector=%s: %w",
			namespace, kind, labelSelector, err)
	}

	return list, nil
}

// kindToGVR converts a Kubernetes kind (e.g., "Pod", "Deployment") to GroupVersionResource.
// It uses the REST mapper to resolve the GVR and determine if the resource is namespaced.
func (c *Client) kindToGVR(kind string) (schema.GroupVersionResource, bool, error) {
	// Create a partially specified GVK (only Kind is known)
	gvk := schema.GroupVersionKind{
		Kind: kind,
	}

	// Use REST mapper to find the full GVR
	mapping, err := c.restMapper.RESTMapping(gvk.GroupKind())
	if err != nil {
		return schema.GroupVersionResource{}, false, fmt.Errorf("failed to map kind=%s to GVR: %w", kind, err)
	}

	return mapping.Resource, mapping.Scope.Name() == meta.RESTScopeNameNamespace, nil
}

// Config returns the underlying REST config (useful for debugging or extensions).
func (c *Client) Config() *rest.Config {
	return c.config
}
