package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PingInput은 sniff_ping Tool의 입력입니다 (빈 구조체)
type PingInput struct{}

// PingOutput은 sniff_ping Tool의 출력입니다
type PingOutput struct {
	Message string `json:"message" jsonschema:"status message from SniffOps server"`
}

// PingHandler는 sniff_ping Tool의 핸들러입니다
//
// 이 Tool은 SniffOps MCP 서버가 정상적으로 작동하는지 확인하는
// 간단한 헬스 체크 기능을 제공합니다.
//
// 입력: 없음
// 출력: "SniffOps MCP Server is running"
func PingHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input PingInput,
) (*mcp.CallToolResult, PingOutput, error) {
	// Context 취소 확인
	select {
	case <-ctx.Done():
		return nil, PingOutput{}, ctx.Err()
	default:
	}

	// 간단한 응답 반환
	output := PingOutput{
		Message: "SniffOps MCP Server is running",
	}

	return &mcp.CallToolResult{}, output, nil
}

// GetToolDefinition은 sniff_ping Tool의 MCP Tool 정의를 반환합니다
func GetPingToolDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "sniff_ping",
		Description: "Check if SniffOps MCP server is running (health check)",
	}
}
