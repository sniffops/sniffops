package tools

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sniffops/sniffops/internal/trace"
)

// StatsInput은 sniff_stats Tool의 입력입니다
type StatsInput struct {
	// Optional: time range filters (future enhancement)
	// StartTime string `json:"start_time,omitempty" jsonschema:"Start time for stats (ISO 8601 format)"`
	// EndTime   string `json:"end_time,omitempty" jsonschema:"End time for stats (ISO 8601 format)"`
}

// StatsOutput은 sniff_stats Tool의 출력입니다
type StatsOutput struct {
	TotalTraces      int                 `json:"total_traces" jsonschema:"Total number of trace records"`
	ToolCounts       map[string]int      `json:"tool_counts" jsonschema:"Count of traces per tool"`
	RiskDistribution map[string]int      `json:"risk_distribution" jsonschema:"Count of traces per risk level"`
	NamespaceCounts  map[string]int      `json:"namespace_counts" jsonschema:"Count of traces per namespace"`
	SuccessRate      float64             `json:"success_rate" jsonschema:"Success rate (percentage)"`
	ResultCounts     map[string]int      `json:"result_counts" jsonschema:"Count of traces per result (success/failure)"`
	AvgLatencyMs     float64             `json:"avg_latency_ms" jsonschema:"Average latency in milliseconds"`
}

// StatsHandler는 sniff_stats Tool의 핸들러입니다
//
// 이 Tool은 trace 통계를 반환합니다:
// - 총 trace 수
// - tool별 카운트
// - risk_level 분포
// - namespace별 카운트
// - 성공률
// - 평균 latency
func StatsHandler(
	traceStore *trace.Store,
) mcp.ToolHandlerFor[StatsInput, StatsOutput] {
	return func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input StatsInput,
	) (*mcp.CallToolResult, StatsOutput, error) {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			return nil, StatsOutput{}, ctx.Err()
		default:
		}

		// Get total count
		totalTraces, err := traceStore.Count(&trace.ListFilter{})
		if err != nil {
			return nil, StatsOutput{}, fmt.Errorf("failed to count total traces: %w", err)
		}

		// Initialize output
		output := StatsOutput{
			TotalTraces:      totalTraces,
			ToolCounts:       make(map[string]int),
			RiskDistribution: make(map[string]int),
			NamespaceCounts:  make(map[string]int),
			ResultCounts:     make(map[string]int),
		}

		// If no traces, return empty stats
		if totalTraces == 0 {
			return &mcp.CallToolResult{}, output, nil
		}

		// Get store's underlying DB for custom queries
		// We need direct SQL access for aggregations
		db := traceStore.DB()
		if db == nil {
			return nil, StatsOutput{}, fmt.Errorf("trace store db is nil")
		}

		// Query tool counts
		rows, err := db.QueryContext(ctx, "SELECT tool_name, COUNT(*) FROM traces GROUP BY tool_name")
		if err != nil {
			return nil, StatsOutput{}, fmt.Errorf("failed to query tool counts: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var toolName string
			var count int
			if err := rows.Scan(&toolName, &count); err != nil {
				return nil, StatsOutput{}, fmt.Errorf("failed to scan tool count: %w", err)
			}
			output.ToolCounts[toolName] = count
		}
		rows.Close()

		// Query risk distribution
		rows, err = db.QueryContext(ctx, "SELECT risk_level, COUNT(*) FROM traces GROUP BY risk_level")
		if err != nil {
			return nil, StatsOutput{}, fmt.Errorf("failed to query risk distribution: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var riskLevel string
			var count int
			if err := rows.Scan(&riskLevel, &count); err != nil {
				return nil, StatsOutput{}, fmt.Errorf("failed to scan risk distribution: %w", err)
			}
			output.RiskDistribution[riskLevel] = count
		}
		rows.Close()

		// Query namespace counts (top 10)
		rows, err = db.QueryContext(ctx, "SELECT namespace, COUNT(*) FROM traces WHERE namespace != '' GROUP BY namespace ORDER BY COUNT(*) DESC LIMIT 10")
		if err != nil {
			return nil, StatsOutput{}, fmt.Errorf("failed to query namespace counts: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var namespace string
			var count int
			if err := rows.Scan(&namespace, &count); err != nil {
				return nil, StatsOutput{}, fmt.Errorf("failed to scan namespace count: %w", err)
			}
			output.NamespaceCounts[namespace] = count
		}
		rows.Close()

		// Query result counts
		rows, err = db.QueryContext(ctx, "SELECT result, COUNT(*) FROM traces GROUP BY result")
		if err != nil {
			return nil, StatsOutput{}, fmt.Errorf("failed to query result counts: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var result string
			var count int
			if err := rows.Scan(&result, &count); err != nil {
				return nil, StatsOutput{}, fmt.Errorf("failed to scan result count: %w", err)
			}
			output.ResultCounts[result] = count
		}
		rows.Close()

		// Calculate success rate
		successCount := output.ResultCounts["success"]
		if totalTraces > 0 {
			output.SuccessRate = float64(successCount) / float64(totalTraces) * 100
		}

		// Query average latency
		var avgLatency sql.NullFloat64
		err = db.QueryRowContext(ctx, "SELECT AVG(latency_ms) FROM traces WHERE latency_ms > 0").Scan(&avgLatency)
		if err != nil {
			return nil, StatsOutput{}, fmt.Errorf("failed to query average latency: %w", err)
		}
		if avgLatency.Valid {
			output.AvgLatencyMs = avgLatency.Float64
		}

		return &mcp.CallToolResult{}, output, nil
	}
}

// GetStatsToolDefinition은 sniff_stats Tool의 MCP Tool 정의를 반환합니다
func GetStatsToolDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "sniff_stats",
		Description: "Get statistics about trace records: total count, tool usage, risk distribution, namespace activity, success rate, and average latency.",
	}
}
