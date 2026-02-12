package k8s

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadKubeConfig tests kubeconfig loading logic.
func TestLoadKubeConfig(t *testing.T) {
	// Save original env
	originalKubeconfig := os.Getenv("KUBECONFIG")
	defer func() {
		if originalKubeconfig != "" {
			os.Setenv("KUBECONFIG", originalKubeconfig)
		} else {
			os.Unsetenv("KUBECONFIG")
		}
	}()

	t.Run("LoadFromDefaultLocation", func(t *testing.T) {
		// Unset KUBECONFIG to force default path
		os.Unsetenv("KUBECONFIG")

		config, err := loadKubeConfig()
		
		// If ~/.kube/config doesn't exist, we expect an error
		homeDir, _ := os.UserHomeDir()
		defaultPath := filepath.Join(homeDir, ".kube", "config")
		
		if _, statErr := os.Stat(defaultPath); os.IsNotExist(statErr) {
			if err == nil {
				t.Error("Expected error when kubeconfig doesn't exist, got nil")
			}
			t.Skipf("Skipping: kubeconfig not found at %s", defaultPath)
			return
		}

		if err != nil {
			t.Skipf("Skipping: failed to load kubeconfig from %s: %v", defaultPath, err)
			return
		}

		if config == nil {
			t.Error("Expected non-nil config")
		}
		if config.Host == "" {
			t.Error("Expected non-empty Host in config")
		}
	})

	t.Run("LoadFromEnvVariable", func(t *testing.T) {
		// Create a temporary kubeconfig path (doesn't need to be valid for this test)
		tmpPath := "/tmp/test-kubeconfig"
		os.Setenv("KUBECONFIG", tmpPath)

		_, err := loadKubeConfig()
		
		// We expect an error since the file doesn't exist
		if err == nil {
			t.Error("Expected error when loading non-existent kubeconfig")
		}
	})
}

// TestNewClient tests client initialization.
func TestNewClient(t *testing.T) {
	client, err := NewClient()
	
	// If we're not in a cluster and don't have a valid kubeconfig, expect error
	if err != nil {
		t.Skipf("Skipping: cannot create client (no kubeconfig or not in cluster): %v", err)
		return
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.dynamicClient == nil {
		t.Error("Expected non-nil dynamicClient")
	}
	if client.discoveryClient == nil {
		t.Error("Expected non-nil discoveryClient")
	}
	if client.restMapper == nil {
		t.Error("Expected non-nil restMapper")
	}
	if client.config == nil {
		t.Error("Expected non-nil config")
	}
}

// TestGetResource tests GetResource with validation (no actual K8s cluster needed).
func TestGetResource(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Skipping: cannot create client: %v", err)
		return
	}

	ctx := context.Background()

	t.Run("MissingKind", func(t *testing.T) {
		_, err := client.GetResource(ctx, "default", "", "test-pod")
		if err == nil {
			t.Error("Expected error when kind is empty")
		}
	})

	t.Run("MissingName", func(t *testing.T) {
		_, err := client.GetResource(ctx, "default", "Pod", "")
		if err == nil {
			t.Error("Expected error when name is empty")
		}
	})

	// Real K8s call would require a running cluster - skip for unit tests
	t.Run("RealGetResource", func(t *testing.T) {
		t.Skip("Skipping real K8s API call - requires running cluster")
		
		// Example of what a real test would look like:
		// obj, err := client.GetResource(ctx, "default", "Pod", "test-pod")
		// if err != nil {
		//     t.Fatalf("Failed to get pod: %v", err)
		// }
		// if obj.GetName() != "test-pod" {
		//     t.Errorf("Expected pod name 'test-pod', got %s", obj.GetName())
		// }
	})
}

// TestListResources tests ListResources with validation (no actual K8s cluster needed).
func TestListResources(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Skipping: cannot create client: %v", err)
		return
	}

	ctx := context.Background()

	t.Run("MissingKind", func(t *testing.T) {
		_, err := client.ListResources(ctx, "default", "", "")
		if err == nil {
			t.Error("Expected error when kind is empty")
		}
	})

	// Real K8s call would require a running cluster - skip for unit tests
	t.Run("RealListResources", func(t *testing.T) {
		t.Skip("Skipping real K8s API call - requires running cluster")
		
		// Example of what a real test would look like:
		// list, err := client.ListResources(ctx, "default", "Pod", "app=nginx")
		// if err != nil {
		//     t.Fatalf("Failed to list pods: %v", err)
		// }
		// if len(list.Items) == 0 {
		//     t.Log("No pods found with label app=nginx")
		// }
	})
}

// TestKindToGVR tests kind resolution (requires discovery client).
func TestKindToGVR(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Skipping: cannot create client: %v", err)
		return
	}

	t.Run("ResolvePodKind", func(t *testing.T) {
		gvr, namespaced, err := client.kindToGVR("Pod")
		if err != nil {
			t.Skipf("Skipping: failed to resolve Pod kind: %v", err)
			return
		}

		if gvr.Resource != "pods" {
			t.Errorf("Expected resource 'pods', got %s", gvr.Resource)
		}
		if !namespaced {
			t.Error("Expected Pod to be namespaced")
		}
	})

	t.Run("ResolveNodeKind", func(t *testing.T) {
		gvr, namespaced, err := client.kindToGVR("Node")
		if err != nil {
			t.Skipf("Skipping: failed to resolve Node kind: %v", err)
			return
		}

		if gvr.Resource != "nodes" {
			t.Errorf("Expected resource 'nodes', got %s", gvr.Resource)
		}
		if namespaced {
			t.Error("Expected Node to be cluster-scoped (not namespaced)")
		}
	})

	t.Run("InvalidKind", func(t *testing.T) {
		_, _, err := client.kindToGVR("InvalidKindDoesNotExist")
		if err == nil {
			t.Error("Expected error when resolving invalid kind")
		}
	})
}
