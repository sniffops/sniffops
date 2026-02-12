package risk

import (
	"testing"
)

func TestGetCommandRisk(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name     string
		toolName string
		action   string
		want     RiskLevel
	}{
		// Low risk tools
		{
			name:     "sniff_get is low risk",
			toolName: "sniff_get",
			action:   "get",
			want:     RiskLow,
		},
		{
			name:     "sniff_logs is low risk",
			toolName: "sniff_logs",
			action:   "logs",
			want:     RiskLow,
		},
		{
			name:     "sniff_traces is low risk",
			toolName: "sniff_traces",
			action:   "query",
			want:     RiskLow,
		},
		{
			name:     "sniff_stats is low risk",
			toolName: "sniff_stats",
			action:   "stats",
			want:     RiskLow,
		},
		// Medium risk tools
		{
			name:     "sniff_apply is medium risk",
			toolName: "sniff_apply",
			action:   "apply",
			want:     RiskMedium,
		},
		// High risk tools
		{
			name:     "sniff_scale is high risk",
			toolName: "sniff_scale",
			action:   "scale",
			want:     RiskHigh,
		},
		// Critical risk tools
		{
			name:     "sniff_delete is critical risk",
			toolName: "sniff_delete",
			action:   "delete",
			want:     RiskCritical,
		},
		{
			name:     "sniff_exec is critical risk",
			toolName: "sniff_exec",
			action:   "exec",
			want:     RiskCritical,
		},
		// Action-based detection (when tool name is unknown)
		{
			name:     "list action is low risk",
			toolName: "",
			action:   "list",
			want:     RiskLow,
		},
		{
			name:     "describe action is low risk",
			toolName: "",
			action:   "describe",
			want:     RiskLow,
		},
		{
			name:     "create action is medium risk",
			toolName: "",
			action:   "create",
			want:     RiskMedium,
		},
		{
			name:     "patch action is medium risk",
			toolName: "",
			action:   "patch",
			want:     RiskMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.getCommandRisk(tt.toolName, tt.action)
			if got != tt.want {
				t.Errorf("getCommandRisk(%q, %q) = %v, want %v", tt.toolName, tt.action, got, tt.want)
			}
		})
	}
}

func TestIsCriticalNamespace(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name      string
		namespace string
		want      bool
	}{
		{
			name:      "kube-system is critical",
			namespace: "kube-system",
			want:      true,
		},
		{
			name:      "kube-public is critical",
			namespace: "kube-public",
			want:      true,
		},
		{
			name:      "production is critical",
			namespace: "production",
			want:      true,
		},
		{
			name:      "prod is critical",
			namespace: "prod",
			want:      true,
		},
		{
			name:      "default is critical",
			namespace: "default",
			want:      true,
		},
		{
			name:      "kube-node-lease is critical",
			namespace: "kube-node-lease",
			want:      true,
		},
		{
			name:      "development is not critical",
			namespace: "development",
			want:      false,
		},
		{
			name:      "test is not critical",
			namespace: "test",
			want:      false,
		},
		{
			name:      "empty namespace is not critical",
			namespace: "",
			want:      false,
		},
		{
			name:      "case insensitive: PRODUCTION is critical",
			namespace: "PRODUCTION",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.isCriticalNamespace(tt.namespace)
			if got != tt.want {
				t.Errorf("isCriticalNamespace(%q) = %v, want %v", tt.namespace, got, tt.want)
			}
		})
	}
}

func TestIsSensitiveResource(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name         string
		resourceKind string
		want         bool
	}{
		{
			name:         "Secret is sensitive",
			resourceKind: "secret",
			want:         true,
		},
		{
			name:         "ConfigMap is sensitive",
			resourceKind: "configmap",
			want:         true,
		},
		{
			name:         "ServiceAccount is sensitive",
			resourceKind: "serviceaccount",
			want:         true,
		},
		{
			name:         "ClusterRole is sensitive",
			resourceKind: "clusterrole",
			want:         true,
		},
		{
			name:         "ClusterRoleBinding is sensitive",
			resourceKind: "clusterrolebinding",
			want:         true,
		},
		{
			name:         "PersistentVolume is sensitive",
			resourceKind: "persistentvolume",
			want:         true,
		},
		{
			name:         "StorageClass is sensitive",
			resourceKind: "storageclass",
			want:         true,
		},
		{
			name:         "Pod is not sensitive",
			resourceKind: "pod",
			want:         false,
		},
		{
			name:         "Deployment is not sensitive",
			resourceKind: "deployment",
			want:         false,
		},
		{
			name:         "Service is not sensitive",
			resourceKind: "service",
			want:         false,
		},
		{
			name:         "empty resource is not sensitive",
			resourceKind: "",
			want:         false,
		},
		{
			name:         "case insensitive: SECRET is sensitive",
			resourceKind: "SECRET",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.isSensitiveResource(tt.resourceKind)
			if got != tt.want {
				t.Errorf("isSensitiveResource(%q) = %v, want %v", tt.resourceKind, got, tt.want)
			}
		})
	}
}

func TestEscalate(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name  string
		level RiskLevel
		want  RiskLevel
	}{
		{
			name:  "low escalates to medium",
			level: RiskLow,
			want:  RiskMedium,
		},
		{
			name:  "medium escalates to high",
			level: RiskMedium,
			want:  RiskHigh,
		},
		{
			name:  "high escalates to critical",
			level: RiskHigh,
			want:  RiskCritical,
		},
		{
			name:  "critical stays critical",
			level: RiskCritical,
			want:  RiskCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.escalate(tt.level)
			if got != tt.want {
				t.Errorf("escalate(%v) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name       string
		ctx        EvalContext
		wantLevel  RiskLevel
		wantReason string // Partial match (contains)
	}{
		// Basic tool risk
		{
			name: "get pod in dev namespace is low risk",
			ctx: EvalContext{
				ToolName:     "sniff_get",
				Namespace:    "development",
				ResourceKind: "pod",
			},
			wantLevel:  RiskLow,
			wantReason: "Read-only",
		},
		{
			name: "apply deployment in dev namespace is medium risk",
			ctx: EvalContext{
				ToolName:     "sniff_apply",
				Namespace:    "development",
				ResourceKind: "deployment",
			},
			wantLevel:  RiskMedium,
			wantReason: "modification",
		},
		// Namespace weight
		{
			name: "get pod in kube-system is medium risk (escalated from low)",
			ctx: EvalContext{
				ToolName:     "sniff_get",
				Namespace:    "kube-system",
				ResourceKind: "pod",
			},
			wantLevel:  RiskMedium,
			wantReason: "kube-system",
		},
		{
			name: "apply in production is high risk (escalated from medium)",
			ctx: EvalContext{
				ToolName:     "sniff_apply",
				Namespace:    "production",
				ResourceKind: "deployment",
			},
			wantLevel:  RiskHigh,
			wantReason: "production",
		},
		// Resource weight
		{
			name: "get secret in dev is medium risk (escalated from low)",
			ctx: EvalContext{
				ToolName:     "sniff_get",
				Namespace:    "development",
				ResourceKind: "secret",
			},
			wantLevel:  RiskMedium,
			wantReason: "Sensitive resource",
		},
		{
			name: "apply configmap in dev is high risk (escalated from medium)",
			ctx: EvalContext{
				ToolName:     "sniff_apply",
				Namespace:    "development",
				ResourceKind: "configmap",
			},
			wantLevel:  RiskHigh,
			wantReason: "configmap",
		},
		// Complex case: delete + kube-system + secret = critical
		{
			name: "delete secret in kube-system is critical",
			ctx: EvalContext{
				ToolName:     "sniff_delete",
				Namespace:    "kube-system",
				ResourceKind: "secret",
			},
			wantLevel:  RiskCritical,
			wantReason: "delete",
		},
		// Scale to 0
		{
			name: "scale to 0 is always critical",
			ctx: EvalContext{
				ToolName:      "sniff_scale",
				Namespace:     "development",
				ResourceKind:  "deployment",
				ResourceCount: 0,
			},
			wantLevel:  RiskCritical,
			wantReason: "Scaling to 0",
		},
		{
			name: "scale to 0 in production is critical",
			ctx: EvalContext{
				ToolName:      "sniff_scale",
				Namespace:     "production",
				ResourceKind:  "deployment",
				ResourceCount: 0,
			},
			wantLevel:  RiskCritical,
			wantReason: "Scaling to 0",
		},
		// Exec is always critical
		{
			name: "exec in dev namespace is critical",
			ctx: EvalContext{
				ToolName:     "sniff_exec",
				Namespace:    "development",
				ResourceKind: "pod",
			},
			wantLevel:  RiskCritical,
			wantReason: "execution",
		},
		// Multiple escalations
		{
			name: "get secret in kube-system is high risk (low→medium→high)",
			ctx: EvalContext{
				ToolName:     "sniff_get",
				Namespace:    "kube-system",
				ResourceKind: "secret",
			},
			wantLevel:  RiskHigh,
			wantReason: "", // Don't check reason for this one
		},
		{
			name: "delete deployment in production is critical",
			ctx: EvalContext{
				ToolName:     "sniff_delete",
				Namespace:    "production",
				ResourceKind: "deployment",
			},
			wantLevel:  RiskCritical,
			wantReason: "delete",
		},
		{
			name: "normal scale is high risk",
			ctx: EvalContext{
				ToolName:      "sniff_scale",
				Namespace:     "development",
				ResourceKind:  "deployment",
				ResourceCount: 3,
			},
			wantLevel:  RiskHigh,
			wantReason: "Scale operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLevel, gotReason := e.Evaluate(tt.ctx)
			if gotLevel != tt.wantLevel {
				t.Errorf("Evaluate() level = %v, want %v", gotLevel, tt.wantLevel)
			}
			if tt.wantReason != "" && !contains(gotReason, tt.wantReason) {
				t.Errorf("Evaluate() reason = %q, want to contain %q", gotReason, tt.wantReason)
			}
		})
	}
}

func TestGenerateReason(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name       string
		ctx        EvalContext
		level      RiskLevel
		wantReason string // Partial match
	}{
		{
			name: "delete operation reason",
			ctx: EvalContext{
				ToolName:     "sniff_delete",
				Namespace:    "production",
				ResourceKind: "pod",
			},
			level:      RiskCritical,
			wantReason: "Destructive operation: delete",
		},
		{
			name: "exec operation reason",
			ctx: EvalContext{
				ToolName:     "sniff_exec",
				Namespace:    "development",
				ResourceKind: "pod",
			},
			level:      RiskCritical,
			wantReason: "Command execution",
		},
		{
			name: "scale to 0 reason",
			ctx: EvalContext{
				ToolName:      "sniff_scale",
				Namespace:     "production",
				ResourceKind:  "deployment",
				ResourceCount: 0,
			},
			level:      RiskCritical,
			wantReason: "Scaling to 0 replicas",
		},
		{
			name: "critical namespace reason",
			ctx: EvalContext{
				ToolName:     "sniff_get",
				Namespace:    "kube-system",
				ResourceKind: "pod",
			},
			level:      RiskMedium,
			wantReason: "Critical namespace: kube-system",
		},
		{
			name: "sensitive resource reason",
			ctx: EvalContext{
				ToolName:     "sniff_get",
				Namespace:    "development",
				ResourceKind: "secret",
			},
			level:      RiskMedium,
			wantReason: "Sensitive resource: secret",
		},
		{
			name: "low risk read-only reason",
			ctx: EvalContext{
				ToolName:     "sniff_get",
				Namespace:    "development",
				ResourceKind: "pod",
			},
			level:      RiskLow,
			wantReason: "Read-only operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReason := e.generateReason(tt.ctx, tt.level)
			if !contains(gotReason, tt.wantReason) {
				t.Errorf("generateReason() = %q, want to contain %q", gotReason, tt.wantReason)
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsCI(s, substr))
}

func containsCI(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
