package risk

import (
	"fmt"
	"strings"
)

// RiskLevel represents the risk level of a Kubernetes operation
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// EvalContext contains the context for risk evaluation
type EvalContext struct {
	ToolName       string // e.g., "sniff_get", "sniff_delete"
	Namespace      string // e.g., "production", "kube-system"
	ResourceKind   string // e.g., "pod", "secret", "configmap"
	Action         string // e.g., "get", "delete", "scale"
	ResourceCount  int    // Number of resources affected (0 means scale to 0)
	TargetResource string // e.g., "pod/nginx-abc123"
}

// Evaluator evaluates the risk level of Kubernetes operations
type Evaluator struct{}

// NewEvaluator creates a new risk evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate calculates the risk level and reason for the given context
func (e *Evaluator) Evaluate(ctx EvalContext) (level RiskLevel, reason string) {
	// Rule 1: Get base risk from tool/command type
	baseRisk := e.getCommandRisk(ctx.ToolName, ctx.Action)

	// Rule 2: Apply namespace weight (critical namespaces escalate risk)
	if e.isCriticalNamespace(ctx.Namespace) {
		baseRisk = e.escalate(baseRisk)
	}

	// Rule 3: Apply resource kind weight (sensitive resources escalate risk)
	if e.isSensitiveResource(ctx.ResourceKind) {
		baseRisk = e.escalate(baseRisk)
	}

	// Rule 4: Special cases
	// Scale to 0 is always critical
	if ctx.ResourceCount == 0 && (ctx.ToolName == "sniff_scale" || ctx.Action == "scale") {
		baseRisk = RiskCritical
	}

	// Generate reason
	reason = e.generateReason(ctx, baseRisk)

	return baseRisk, reason
}

// getCommandRisk returns the base risk level for a given tool/action
func (e *Evaluator) getCommandRisk(tool, action string) RiskLevel {
	// Priority 1: Check tool name (MCP tool convention)
	switch tool {
	case "sniff_get", "sniff_logs", "sniff_traces", "sniff_stats":
		return RiskLow
	case "sniff_apply":
		return RiskMedium
	case "sniff_delete":
		return RiskCritical
	case "sniff_exec":
		return RiskCritical
	case "sniff_scale":
		return RiskHigh
	}

	// Priority 2: Check action (for custom/direct calls)
	actionLower := strings.ToLower(action)
	switch {
	case strings.Contains(actionLower, "get") || strings.Contains(actionLower, "list"):
		return RiskLow
	case strings.Contains(actionLower, "logs") || strings.Contains(actionLower, "describe"):
		return RiskLow
	case strings.Contains(actionLower, "delete"):
		return RiskCritical
	case strings.Contains(actionLower, "exec"):
		return RiskCritical
	case strings.Contains(actionLower, "apply") || strings.Contains(actionLower, "create"):
		return RiskMedium
	case strings.Contains(actionLower, "patch") || strings.Contains(actionLower, "update"):
		return RiskMedium
	case strings.Contains(actionLower, "scale"):
		return RiskHigh
	}

	// Default: medium (unknown operation)
	return RiskMedium
}

// isCriticalNamespace checks if the namespace is critical
func (e *Evaluator) isCriticalNamespace(ns string) bool {
	if ns == "" {
		return false
	}

	criticalNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"production",
		"prod",
		"default",
	}

	nsLower := strings.ToLower(ns)
	for _, critical := range criticalNamespaces {
		if nsLower == critical {
			return true
		}
	}

	return false
}

// isSensitiveResource checks if the resource kind contains sensitive data
func (e *Evaluator) isSensitiveResource(kind string) bool {
	if kind == "" {
		return false
	}

	sensitiveResources := []string{
		"secret",
		"configmap",
		"serviceaccount",
		"clusterrole",
		"clusterrolebinding",
		"persistentvolume",
		"storageclass",
	}

	kindLower := strings.ToLower(kind)
	for _, sensitive := range sensitiveResources {
		if kindLower == sensitive {
			return true
		}
	}

	return false
}

// escalate increases the risk level by one step
func (e *Evaluator) escalate(level RiskLevel) RiskLevel {
	switch level {
	case RiskLow:
		return RiskMedium
	case RiskMedium:
		return RiskHigh
	case RiskHigh:
		return RiskCritical
	case RiskCritical:
		return RiskCritical // Already at max
	}
	return level
}

// generateReason creates a human-readable reason for the risk level
func (e *Evaluator) generateReason(ctx EvalContext, level RiskLevel) string {
	reasons := []string{}

	// Describe the operation (check tool/action first, independent of level)
	if ctx.ToolName == "sniff_delete" || strings.Contains(strings.ToLower(ctx.Action), "delete") {
		reasons = append(reasons, "Destructive operation: delete")
	} else if ctx.ToolName == "sniff_exec" || strings.Contains(strings.ToLower(ctx.Action), "exec") {
		reasons = append(reasons, fmt.Sprintf("Command execution: %s", ctx.ToolName))
	} else if ctx.ToolName == "sniff_scale" || strings.Contains(strings.ToLower(ctx.Action), "scale") {
		reasons = append(reasons, fmt.Sprintf("Scale operation: %s", ctx.ToolName))
	}

	// Check namespace criticality
	if e.isCriticalNamespace(ctx.Namespace) {
		reasons = append(reasons, fmt.Sprintf("Critical namespace: %s", ctx.Namespace))
	}

	// Check resource sensitivity
	if e.isSensitiveResource(ctx.ResourceKind) {
		reasons = append(reasons, fmt.Sprintf("Sensitive resource: %s", ctx.ResourceKind))
	}

	// Special cases
	if ctx.ResourceCount == 0 && (ctx.ToolName == "sniff_scale" || strings.Contains(strings.ToLower(ctx.Action), "scale")) {
		reasons = append(reasons, "Scaling to 0 replicas")
	}

	// Default reason for low-risk operations
	if len(reasons) == 0 {
		if level == RiskLow {
			return "Read-only operation"
		}
		if level == RiskMedium {
			return fmt.Sprintf("Resource modification: %s", ctx.ToolName)
		}
	}

	return strings.Join(reasons, "; ")
}
