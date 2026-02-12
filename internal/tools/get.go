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

// GetInput은 sniff_get Tool의 입력입니다
type GetInput struct {
	Namespace string `json:"namespace" jsonschema:"Kubernetes namespace (required for namespaced resources)"`
	Kind      string `json:"kind" jsonschema:"Resource kind (e.g., Pod, Deployment, Service)"`
	Name      string `json:"name,omitempty" jsonschema:"Resource name (optional; if omitted, lists all resources)"`
}

// GetOutput은 sniff_get Tool의 출력입니다
type GetOutput struct {
	Resources interface{} `json:"resources" jsonschema:"K8s resource(s) in JSON format"`
	Count     int         `json:"count" jsonschema:"Number of resources returned"`
}

// GetHandler는 sniff_get Tool의 핸들러입니다
//
// 이 Tool은 Kubernetes 리소스를 조회합니다:
// - name이 주어지면 GetResource로 단일 리소스 조회
// - name이 없으면 ListResources로 목록 조회
// - Trace 기록 및 위험도 평가 수행
func GetHandler(
	k8sClient *k8s.Client,
	traceStore *trace.Store,
	riskEvaluator *risk.Evaluator,
	sessionID string,
) mcp.ToolHandlerFor[GetInput, GetOutput] {
	return func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input GetInput,
	) (*mcp.CallToolResult, GetOutput, error) {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			return nil, GetOutput{}, ctx.Err()
		default:
		}

		// Trace 시작
		startTime := time.Now()
		traceID := uuid.New().String()

		// Build command string
		command := fmt.Sprintf("kubectl get %s -n %s", input.Kind, input.Namespace)
		if input.Name != "" {
			command += fmt.Sprintf(" %s", input.Name)
		}

		// User intent 생성
		userIntent := fmt.Sprintf("Get %s resources in namespace %s", input.Kind, input.Namespace)
		if input.Name != "" {
			userIntent = fmt.Sprintf("Get %s %s in namespace %s", input.Kind, input.Name, input.Namespace)
		}

		// 초기 trace 레코드 생성
		tr := &trace.Trace{
			ID:             traceID,
			SessionID:      sessionID,
			Timestamp:      startTime.UnixMilli(),
			UserIntent:     userIntent,
			ToolName:       "sniff_get",
			Command:        command,
			Namespace:      input.Namespace,
			ResourceKind:   input.Kind,
			TargetResource: input.Name,
		}

		// K8s API 호출
		var output GetOutput
		var execErr error

		if input.Name != "" {
			// GetResource (단일 리소스 조회)
			resource, err := k8sClient.GetResource(ctx, input.Namespace, input.Kind, input.Name)
			if err != nil {
				execErr = err
			} else {
				output.Resources = resource.Object
				output.Count = 1
			}
		} else {
			// ListResources (목록 조회)
			resourceList, err := k8sClient.ListResources(ctx, input.Namespace, input.Kind, "")
			if err != nil {
				execErr = err
			} else {
				output.Resources = resourceList.Items
				output.Count = len(resourceList.Items)
			}
		}

		// Trace 완료 처리
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// 위험도 평가
		riskLevel, riskReason := riskEvaluator.Evaluate(risk.EvalContext{
			ToolName:     "sniff_get",
			Namespace:    input.Namespace,
			ResourceKind: input.Kind,
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
			// Trace 저장 실패는 로깅만 하고 Tool 실행은 계속
			fmt.Fprintf(os.Stderr, "Warning: failed to save trace: %v\n", err)
		}

		// 에러 발생 시 반환
		if execErr != nil {
			return nil, GetOutput{}, fmt.Errorf("failed to get K8s resource: %w", execErr)
		}

		return &mcp.CallToolResult{}, output, nil
	}
}

// GetGetToolDefinition은 sniff_get Tool의 MCP Tool 정의를 반환합니다
func GetGetToolDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "sniff_get",
		Description: "Get Kubernetes resources (pod, deployment, service, etc). If name is provided, retrieves a single resource; otherwise lists all resources of that kind in the namespace.",
	}
}
