package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sniffops/sniffops/internal/k8s"
	"github.com/sniffops/sniffops/internal/risk"
	"github.com/sniffops/sniffops/internal/trace"
)

// ScaleInput은 sniff_scale Tool의 입력입니다
type ScaleInput struct {
	Namespace string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Name      string `json:"name" jsonschema:"Resource name (Deployment or StatefulSet)"`
	Replicas  int32  `json:"replicas" jsonschema:"Target replica count"`
}

// ScaleOutput은 sniff_scale Tool의 출력입니다
type ScaleOutput struct {
	Scaled   string `json:"scaled" jsonschema:"Scaled resource identifier"`
	Replicas int32  `json:"replicas" jsonschema:"New replica count"`
	Warning  string `json:"warning,omitempty" jsonschema:"Warning message for risky operations"`
	RiskInfo string `json:"risk_info,omitempty" jsonschema:"Risk level and reason"`
}

// ScaleHandler는 sniff_scale Tool의 핸들러입니다
//
// 이 Tool은 Kubernetes Deployment/StatefulSet을 스케일합니다:
// - Scale to 0은 critical 위험도
// - Trace 기록 및 위험도 평가 수행
func ScaleHandler(
	k8sClient *k8s.Client,
	traceStore *trace.Store,
	riskEvaluator *risk.Evaluator,
	sessionID string,
) mcp.ToolHandlerFor[ScaleInput, ScaleOutput] {
	return func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input ScaleInput,
	) (*mcp.CallToolResult, ScaleOutput, error) {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			return nil, ScaleOutput{}, ctx.Err()
		default:
		}

		// Trace 시작
		startTime := time.Now()
		traceID := uuid.New().String()

		// Build command string (assume Deployment by default)
		command := fmt.Sprintf("kubectl scale deployment -n %s %s --replicas=%d", input.Namespace, input.Name, input.Replicas)

		// 초기 trace 레코드 생성
		tr := &trace.Trace{
			ID:             traceID,
			SessionID:      sessionID,
			Timestamp:      startTime.UnixMilli(),
			ToolName:       "sniff_scale",
			Command:        command,
			Namespace:      input.Namespace,
			ResourceKind:   "Deployment", // 일반적으로 Deployment
			TargetResource: input.Name,
		}

		// 위험도 평가 (scale 전 평가)
		riskLevel, riskReason := riskEvaluator.Evaluate(risk.EvalContext{
			ToolName:      "sniff_scale",
			Namespace:     input.Namespace,
			ResourceKind:  "Deployment",
			Action:        "scale",
			ResourceCount: int(input.Replicas),
		})

		// K8s API 호출 (Scale) - try Deployment first, then StatefulSet
		var result interface{}
		var execErr error
		var kind string

		// Try Deployment first
		deploymentResult, err := k8sClient.Scale(ctx, input.Namespace, "Deployment", input.Name, input.Replicas)
		if err == nil {
			result = deploymentResult.Object
			kind = "Deployment"
		} else {
			// Try StatefulSet if Deployment fails
			statefulsetResult, err2 := k8sClient.Scale(ctx, input.Namespace, "StatefulSet", input.Name, input.Replicas)
			if err2 == nil {
				result = statefulsetResult.Object
				kind = "StatefulSet"
			} else {
				// Both failed
				execErr = fmt.Errorf("failed to scale as Deployment or StatefulSet: %w", err)
			}
		}

		// Trace 완료 처리
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		var output ScaleOutput
		if execErr == nil {
			output.Scaled = fmt.Sprintf("%s/%s", kind, input.Name)
			output.Replicas = input.Replicas
			output.RiskInfo = fmt.Sprintf("Risk Level: %s - %s", riskLevel, riskReason)

			// Scale to 0은 critical
			if input.Replicas == 0 {
				output.Warning = "⚠️  CRITICAL: Scaled to 0 replicas. Service will be unavailable until scaled back up."
			} else if riskLevel == risk.RiskCritical || riskLevel == risk.RiskHigh {
				output.Warning = fmt.Sprintf("⚠️  %s: This scaling operation may affect service availability.", riskLevel)
			}

			tr.ResourceKind = kind
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
			fmt.Fprintf(req.Session.LoggingChannel(), "Warning: failed to save trace: %v\n", err)
		}

		// 에러 발생 시 반환
		if execErr != nil {
			return nil, ScaleOutput{}, fmt.Errorf("failed to scale K8s resource: %w", execErr)
		}

		return &mcp.CallToolResult{}, output, nil
	}
}

// GetScaleToolDefinition은 sniff_scale Tool의 MCP Tool 정의를 반환합니다
func GetScaleToolDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "sniff_scale",
		Description: "Scale a Kubernetes Deployment or StatefulSet to a specified number of replicas. ⚠️  Scaling to 0 will make the service unavailable.",
	}
}
