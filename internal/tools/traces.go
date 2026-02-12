package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sniffops/sniffops/internal/trace"
)

// TracesInput은 sniff_traces Tool의 입력입니다
type TracesInput struct {
	Tool      string `json:"tool,omitempty" jsonschema:"Filter by tool name (e.g., sniff_get, sniff_delete)"`
	Namespace string `json:"namespace,omitempty" jsonschema:"Filter by namespace"`
	RiskLevel string `json:"risk_level,omitempty" jsonschema:"Filter by risk level (low, medium, high, critical)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of traces to return (default: 20, max: 100)"`
	Offset    int    `json:"offset,omitempty" jsonschema:"Offset for pagination (default: 0)"`
}

// TracesOutput은 sniff_traces Tool의 출력입니다
type TracesOutput struct {
	Traces []*trace.Trace `json:"traces" jsonschema:"List of trace records"`
	Count  int            `json:"count" jsonschema:"Number of traces returned"`
	Total  int            `json:"total" jsonschema:"Total number of traces matching filters"`
}

// TracesHandler는 sniff_traces Tool의 핸들러입니다
//
// 이 Tool은 SQLite에서 trace를 조회합니다:
// - 필터링: tool, namespace, risk_level
// - 페이지네이션: limit, offset
// - 기본 limit: 20
func TracesHandler(
	traceStore *trace.Store,
) mcp.ToolHandlerFor[TracesInput, TracesOutput] {
	return func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input TracesInput,
	) (*mcp.CallToolResult, TracesOutput, error) {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			return nil, TracesOutput{}, ctx.Err()
		default:
		}

		// Default limit to 20
		if input.Limit <= 0 {
			input.Limit = 20
		}

		// Cap limit at 100
		if input.Limit > 100 {
			input.Limit = 100
		}

		// Build filter
		filter := &trace.ListFilter{
			Tool:      input.Tool,
			Namespace: input.Namespace,
			RiskLevel: input.RiskLevel,
			Limit:     input.Limit,
			Offset:    input.Offset,
		}

		// Query traces
		traces, err := traceStore.List(filter)
		if err != nil {
			return nil, TracesOutput{}, fmt.Errorf("failed to query traces: %w", err)
		}

		// Count total matching traces
		total, err := traceStore.Count(filter)
		if err != nil {
			return nil, TracesOutput{}, fmt.Errorf("failed to count traces: %w", err)
		}

		output := TracesOutput{
			Traces: traces,
			Count:  len(traces),
			Total:  total,
		}

		return &mcp.CallToolResult{}, output, nil
	}
}

// GetTracesToolDefinition은 sniff_traces Tool의 MCP Tool 정의를 반환합니다
func GetTracesToolDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "sniff_traces",
		Description: "Query trace records from the audit log. Filter by tool, namespace, risk level. Supports pagination with limit and offset.",
	}
}
