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

// ApplyInput은 sniff_apply Tool의 입력입니다
type ApplyInput struct {
	Manifest string `json:"manifest" jsonschema:"Kubernetes resource manifest (YAML or JSON string)"`
}

// ApplyOutput은 sniff_apply Tool의 출력입니다
type ApplyOutput struct {
	Applied  interface{} `json:"applied" jsonschema:"Applied resource in JSON format"`
	Resource string      `json:"resource" jsonschema:"Resource identifier (kind/name)"`
}

// ApplyHandler는 sniff_apply Tool의 핸들러입니다
//
// 이 Tool은 Kubernetes 리소스를 apply합니다:
// - Server-side apply 사용
// - Trace 기록 및 위험도 평가 수행 (기본 high)
func ApplyHandler(
	k8sClient *k8s.Client,
	traceStore *trace.Store,
	riskEvaluator *risk.Evaluator,
	sessionID string,
) mcp.ToolHandlerFor[ApplyInput, ApplyOutput] {
	return func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input ApplyInput,
	) (*mcp.CallToolResult, ApplyOutput, error) {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			return nil, ApplyOutput{}, ctx.Err()
		default:
		}

		// Trace 시작
		startTime := time.Now()
		traceID := uuid.New().String()

		// Build command string
		command := "kubectl apply -f -"

		// User intent 생성
		userIntent := "Apply Kubernetes resource from manifest"

		// 초기 trace 레코드 생성
		tr := &trace.Trace{
			ID:         traceID,
			SessionID:  sessionID,
			Timestamp:  startTime.UnixMilli(),
			UserIntent: userIntent,
			ToolName:   "sniff_apply",
			Command:    command,
		}

		// K8s API 호출 (Apply)
		result, execErr := k8sClient.Apply(ctx, input.Manifest)

		// Trace 완료 처리
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		var output ApplyOutput
		var namespace, kind, name string

		if execErr == nil && result != nil {
			// Extract metadata for trace
			namespace = result.GetNamespace()
			kind = result.GetKind()
			name = result.GetName()

			output.Applied = result.Object
			output.Resource = fmt.Sprintf("%s/%s", kind, name)

			tr.Namespace = namespace
			tr.ResourceKind = kind
			tr.TargetResource = name
		}

		// 위험도 평가
		riskLevel, riskReason := riskEvaluator.Evaluate(risk.EvalContext{
			ToolName:     "sniff_apply",
			Namespace:    namespace,
			ResourceKind: kind,
			Action:       "apply",
		})

		// Trace 레코드 완성
		tr.LatencyMs = int(duration.Milliseconds())
		tr.RiskLevel = string(riskLevel)
		tr.RiskReason = riskReason

		if execErr != nil {
			tr.Result = "failure"
			tr.ErrorMessage = execErr.Error()
		} else {
			tr.Result = "success"
			// Output을 JSON으로 저장 (민감 정보 sanitize 적용)
			outputJSON, _ := json.Marshal(output)
			tr.Output = trace.SanitizeOutput(string(outputJSON))
		}

		// Trace 저장
		if err := traceStore.Insert(tr); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save trace: %v\n", err)
		}

		// 에러 발생 시 반환
		if execErr != nil {
			return nil, ApplyOutput{}, fmt.Errorf("failed to apply K8s resource: %w", execErr)
		}

		return &mcp.CallToolResult{}, output, nil
	}
}

// GetApplyToolDefinition은 sniff_apply Tool의 MCP Tool 정의를 반환합니다
func GetApplyToolDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "sniff_apply",
		Description: "Apply a Kubernetes resource using server-side apply. Accepts YAML or JSON manifest. Creates or updates the resource.",
	}
}
