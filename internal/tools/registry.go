package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sniffops/sniffops/internal/k8s"
	"github.com/sniffops/sniffops/internal/risk"
	"github.com/sniffops/sniffops/internal/trace"
)

// RegisterAllTools registers all MCP Tools to the server
//
// This function is the central place for Tool registration. It keeps Tool registration
// logic organized and maintainable.
//
// Parameters:
//   - server: MCP server instance
//   - k8sClient: Kubernetes client for API calls (can be nil for tools that don't need it)
//   - traceStore: SQLite store for trace recording (can be nil to disable tracing)
//   - riskEvaluator: Risk evaluator for security assessment (can be nil to skip risk eval)
//   - sessionID: Session ID for trace records
func RegisterAllTools(
	server *mcp.Server,
	k8sClient *k8s.Client,
	traceStore *trace.Store,
	riskEvaluator *risk.Evaluator,
	sessionID string,
) {
	// 1. sniff_ping - Health check (no dependencies)
	mcp.AddTool(
		server,
		GetPingToolDefinition(),
		PingHandler,
	)

	// 2. sniff_get - K8s resource retrieval
	if k8sClient != nil && traceStore != nil && riskEvaluator != nil {
		mcp.AddTool(
			server,
			GetGetToolDefinition(),
			GetHandler(k8sClient, traceStore, riskEvaluator, sessionID),
		)
	}

	// 3. sniff_logs - Pod logs retrieval
	if k8sClient != nil && traceStore != nil && riskEvaluator != nil {
		mcp.AddTool(
			server,
			GetLogsToolDefinition(),
			LogsHandler(k8sClient, traceStore, riskEvaluator, sessionID),
		)
	}

	// 4. sniff_apply - Apply K8s resources (TASK-010)
	if k8sClient != nil && traceStore != nil && riskEvaluator != nil {
		mcp.AddTool(
			server,
			GetApplyToolDefinition(),
			ApplyHandler(k8sClient, traceStore, riskEvaluator, sessionID),
		)
	}

	// 5. sniff_delete - Delete K8s resources (TASK-010) ⚠️  CRITICAL
	if k8sClient != nil && traceStore != nil && riskEvaluator != nil {
		mcp.AddTool(
			server,
			GetDeleteToolDefinition(),
			DeleteHandler(k8sClient, traceStore, riskEvaluator, sessionID),
		)
	}

	// 6. sniff_scale - Scale Deployments/StatefulSets (TASK-010)
	if k8sClient != nil && traceStore != nil && riskEvaluator != nil {
		mcp.AddTool(
			server,
			GetScaleToolDefinition(),
			ScaleHandler(k8sClient, traceStore, riskEvaluator, sessionID),
		)
	}

	// 7. sniff_exec - Execute commands in pods (TASK-010) ⚠️  CRITICAL
	if k8sClient != nil && traceStore != nil && riskEvaluator != nil {
		mcp.AddTool(
			server,
			GetExecToolDefinition(),
			ExecHandler(k8sClient, traceStore, riskEvaluator, sessionID),
		)
	}

	// 8. sniff_traces - Query trace records (TASK-010)
	if traceStore != nil {
		mcp.AddTool(
			server,
			GetTracesToolDefinition(),
			TracesHandler(traceStore),
		)
	}

	// 9. sniff_stats - Get trace statistics (TASK-010)
	if traceStore != nil {
		mcp.AddTool(
			server,
			GetStatsToolDefinition(),
			StatsHandler(traceStore),
		)
	}
}
