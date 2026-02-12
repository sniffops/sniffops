package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sniffops/sniffops/internal/k8s"
	"github.com/sniffops/sniffops/internal/risk"
	"github.com/sniffops/sniffops/internal/trace"
)

// DeleteInput은 sniff_delete Tool의 입력입니다
type DeleteInput struct {
	Namespace string `json:"namespace" jsonschema:"Kubernetes namespace (required for namespaced resources)"`
	Kind      string `json:"kind" jsonschema:"Resource kind (e.g., Pod, Deployment, Service)"`
	Name      string `json:"name" jsonschema:"Resource name"`
}

// DeleteOutput은 sniff_delete Tool의 출력입니다
type DeleteOutput struct {
	Deleted  string `json:"deleted" jsonschema:"Deleted resource identifier (kind/name)"`
	Warning  string `json:"warning,omitempty" jsonschema:"Warning message for critical operations"`
	RiskInfo string `json:"risk_info,omitempty" jsonschema:"Risk level and reason"`
}

// DeleteHandler는 sniff_delete Tool의 핸들러입니다
//
// 이 Tool은 Kubernetes 리소스를 삭제합니다:
// - 위험한 작업이므로 기본 위험도 critical
// - 경고 메시지 포함
// - Trace 기록 및 위험도 평가 수행
func DeleteHandler(
	k8sClient *k8s.Client,
	traceStore *trace.Store,
	riskEvaluator *risk.Evaluator,
	sessionID string,
) mcp.ToolHandlerFor[DeleteInput, DeleteOutput] {
	return func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input DeleteInput,
	) (*mcp.CallToolResult, DeleteOutput, error) {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			return nil, DeleteOutput{}, ctx.Err()
		default:
		}

		// Trace 시작
		startTime := time.Now()
		traceID := uuid.New().String()

		// Build command string
		command := fmt.Sprintf("kubectl delete %s -n %s %s", input.Kind, input.Namespace, input.Name)

		// 초기 trace 레코드 생성
		tr := &trace.Trace{
			ID:             traceID,
			SessionID:      sessionID,
			Timestamp:      startTime.UnixMilli(),
			ToolName:       "sniff_delete",
			Command:        command,
			Namespace:      input.Namespace,
			ResourceKind:   input.Kind,
			TargetResource: input.Name,
		}

		// 위험도 평가 (삭제 전 평가)
		riskLevel, riskReason := riskEvaluator.Evaluate(risk.EvalContext{
			ToolName:     "sniff_delete",
			Namespace:    input.Namespace,
			ResourceKind: input.Kind,
			Action:       "delete",
		})

		// K8s API 호출 (Delete)
		execErr := k8sClient.Delete(ctx, input.Namespace, input.Kind, input.Name)

		// Trace 완료 처리
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		var output DeleteOutput
		output.Deleted = fmt.Sprintf("%s/%s", input.Kind, input.Name)
		output.RiskInfo = fmt.Sprintf("Risk Level: %s - %s", riskLevel, riskReason)

		// 위험도가 critical이면 경고 메시지 추가
		if riskLevel == risk.RiskCritical {
			output.Warning = "⚠️  CRITICAL OPERATION: This is a destructive action that cannot be undone. The resource has been permanently deleted."
		}

		// Trace 레코드 완성
		tr.LatencyMs = int(duration.Milliseconds())
		tr.RiskLevel = string(riskLevel)
		tr.RiskReason = riskReason

		if execErr != nil {
			tr.Result = "failure"
			tr.ErrorMessage = execErr.Error()
		} else {
			tr.Result = "success"
			// Output을 JSON으로 저장
			outputJSON, _ := json.Marshal(output)
			tr.Output = string(outputJSON)
		}

		// Trace 저장
		if err := traceStore.Insert(tr); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save trace: %v\n", err)
		}

		// 에러 발생 시 반환
		if execErr != nil {
			return nil, DeleteOutput{}, fmt.Errorf("failed to delete K8s resource: %w", execErr)
		}

		return &mcp.CallToolResult{}, output, nil
	}
}

// GetDeleteToolDefinition은 sniff_delete Tool의 MCP Tool 정의를 반환합니다
func GetDeleteToolDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "sniff_delete",
		Description: "⚠️  Delete a Kubernetes resource. This is a CRITICAL operation that permanently removes the resource. Use with extreme caution.",
	}
}
