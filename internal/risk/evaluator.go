// Package risk provides risk evaluation logic for K8s operations.
package risk

import (
	"fmt"
	"strings"
)

// Level represents risk severity
type Level string

const (
	LevelLow      Level = "low"
	LevelMedium   Level = "medium"
	LevelHigh     Level = "high"
	LevelCritical Level = "critical"
)

// EvalContext contains context for risk evaluation
type EvalContext struct {
	ToolName      string
	Namespace     string
	ResourceKind  string
	ResourceCount int
}

// Evaluator evaluates risk level for K8s operations
type Evaluator struct {
	criticalNamespaces []string
}

// NewEvaluator creates a new risk evaluator with default critical namespaces
func NewEvaluator() *Evaluator {
	return &Evaluator{
		criticalNamespaces: []string{"production", "prod", "default", "kube-system"},
	}
}

// Evaluate calculates risk level and reason for a given operation
func (e *Evaluator) Evaluate(ctx EvalContext) (level Level, reason string) {
	// Rule 1: Get base risk from tool name
	baseRisk := e.getCommandRisk(ctx.ToolName)

	// Rule 2: Escalate if critical namespace
	if e.isCriticalNamespace(ctx.Namespace) {
		baseRisk = e.escalate(baseRisk)
	}

	// Rule 3: Scale to 0 is critical
	if ctx.ResourceCount == 0 && ctx.ToolName == "sniff_scale" {
		baseRisk = LevelCritical
	}

	// Generate reason
	reason = e.getReason(ctx, baseRisk)

	return baseRisk, reason
}

// getCommandRisk returns base risk level for a tool
func (e *Evaluator) getCommandRisk(tool string) Level {
	switch tool {
	case "sniff_get", "sniff_logs", "sniff_traces", "sniff_stats", "sniff_ping":
		return LevelLow
	case "sniff_apply":
		return LevelMedium
	case "sniff_delete", "sniff_scale", "sniff_exec":
		return LevelHigh
	default:
		return LevelMedium
	}
}

// isCriticalNamespace checks if namespace is considered critical
func (e *Evaluator) isCriticalNamespace(ns string) bool {
	for _, critical := range e.criticalNamespaces {
		if strings.EqualFold(ns, critical) {
			return true
		}
	}
	return false
}

// escalate increases risk level by one step
func (e *Evaluator) escalate(level Level) Level {
	switch level {
	case LevelLow:
		return LevelMedium
	case LevelMedium:
		return LevelHigh
	case LevelHigh:
		return LevelCritical
	default:
		return level
	}
}

// getReason generates human-readable reason for risk level
func (e *Evaluator) getReason(ctx EvalContext, level Level) string {
	var reasons []string

	// Add reason for destructive operations
	if level >= LevelHigh {
		reasons = append(reasons, fmt.Sprintf("Destructive operation: %s", ctx.ToolName))
	}

	// Add reason for critical namespace
	if e.isCriticalNamespace(ctx.Namespace) {
		reasons = append(reasons, fmt.Sprintf("Critical namespace: %s", ctx.Namespace))
	}

	// Add reason for scale to 0
	if ctx.ResourceCount == 0 && ctx.ToolName == "sniff_scale" {
		reasons = append(reasons, "Scaling to 0 replicas")
	}

	// Default reason for low-risk operations
	if len(reasons) == 0 {
		return "Read-only operation"
	}

	return strings.Join(reasons, "; ")
}
