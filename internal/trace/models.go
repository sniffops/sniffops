package trace

import "time"

// Trace represents a single K8s operation trace record
type Trace struct {
	// Identity
	ID        string `json:"id" db:"id"`
	SessionID string `json:"session_id" db:"session_id"`
	Timestamp int64  `json:"timestamp" db:"timestamp"` // Unix timestamp (ms)

	// Request Context
	UserIntent string `json:"user_intent,omitempty" db:"user_intent"`
	ToolName   string `json:"tool_name" db:"tool_name"`

	// K8s Command Details
	Command        string `json:"command" db:"command"`
	TargetResource string `json:"target_resource,omitempty" db:"target_resource"`
	Namespace      string `json:"namespace,omitempty" db:"namespace"`
	ResourceKind   string `json:"resource_kind,omitempty" db:"resource_kind"`

	// Risk & Security
	RiskLevel  string `json:"risk_level" db:"risk_level"`
	RiskReason string `json:"risk_reason,omitempty" db:"risk_reason"`

	// Execution Result
	Result       string `json:"result" db:"result"`
	Output       string `json:"output,omitempty" db:"output"`
	ErrorMessage string `json:"error_message,omitempty" db:"error_message"`

	// Metrics
	LatencyMs    int     `json:"latency_ms,omitempty" db:"latency_ms"`
	TokensInput  int     `json:"tokens_input,omitempty" db:"tokens_input"`
	TokensOutput int     `json:"tokens_output,omitempty" db:"tokens_output"`
	CostEstimate float64 `json:"cost_estimate,omitempty" db:"cost_estimate"`

	// Metadata
	Kubeconfig  string `json:"kubeconfig,omitempty" db:"kubeconfig"`
	ClusterName string `json:"cluster_name,omitempty" db:"cluster_name"`
}

// ListFilter defines filtering options for trace queries
type ListFilter struct {
	// Filtering
	Tool      string
	Namespace string
	RiskLevel string
	StartTime *time.Time
	EndTime   *time.Time

	// Pagination
	Limit  int
	Offset int
}
