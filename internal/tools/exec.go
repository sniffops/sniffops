package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sniffops/sniffops/internal/k8s"
	"github.com/sniffops/sniffops/internal/risk"
	"github.com/sniffops/sniffops/internal/trace"
)

// ExecInput은 sniff_exec Tool의 입력입니다
type ExecInput struct {
	Namespace string   `json:"namespace" jsonschema:"Kubernetes namespace"`
	Pod       string   `json:"pod" jsonschema:"Pod name"`
	Container string   `json:"container,omitempty" jsonschema:"Container name (optional; uses first container if omitted)"`
	Command   []string `json:"command" jsonschema:"Command to execute (array of strings, e.g., ['ls', '-la'])"`
}

// ExecOutput은 sniff_exec Tool의 출력입니다
type ExecOutput struct {
	Output   string `json:"output" jsonschema:"Command output (stdout + stderr)"`
	Pod      string `json:"pod" jsonschema:"Pod name"`
	Command  string `json:"command" jsonschema:"Executed command"`
	Warning  string `json:"warning,omitempty" jsonschema:"Warning message for critical operations"`
	RiskInfo string `json:"risk_info,omitempty" jsonschema:"Risk level and reason"`
}

// ExecHandler는 sniff_exec Tool의 핸들러입니다
//
// 이 Tool은 Kubernetes Pod에서 명령을 실행합니다:
// - 위험한 작업이므로 기본 위험도 critical
// - 경고 메시지 포함
// - Trace 기록 및 위험도 평가 수행
func ExecHandler(
	k8sClient *k8s.Client,
	traceStore *trace.Store,
	riskEvaluator *risk.Evaluator,
	sessionID string,
) mcp.ToolHandlerFor[ExecInput, ExecOutput] {
	return func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input ExecInput,
	) (*mcp.CallToolResult, ExecOutput, error) {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			return nil, ExecOutput{}, ctx.Err()
		default:
		}

		// Trace 시작
		startTime := time.Now()
		traceID := uuid.New().String()

		// Build command string
		commandStr := strings.Join(input.Command, " ")
		kubectlCmd := fmt.Sprintf("kubectl exec -n %s %s", input.Namespace, input.Pod)
		if input.Container != "" {
			kubectlCmd += fmt.Sprintf(" -c %s", input.Container)
		}
		kubectlCmd += fmt.Sprintf(" -- %s", commandStr)

		// User intent 생성
		userIntent := fmt.Sprintf("Execute command '%s' in pod %s (namespace %s)", commandStr, input.Pod, input.Namespace)

		// 초기 trace 레코드 생성
		tr := &trace.Trace{
			ID:             traceID,
			SessionID:      sessionID,
			Timestamp:      startTime.UnixMilli(),
			UserIntent:     userIntent,
			ToolName:       "sniff_exec",
			Command:        kubectlCmd,
			Namespace:      input.Namespace,
			ResourceKind:   "Pod",
			TargetResource: input.Pod,
		}

		// 위험도 평가 (exec 전 평가)
		riskLevel, riskReason := riskEvaluator.Evaluate(risk.EvalContext{
			ToolName:     "sniff_exec",
			Namespace:    input.Namespace,
			ResourceKind: "Pod",
			Action:       "exec",
		})

		// K8s API 호출 (Exec)
		execOutput, execErr := k8sClient.Exec(ctx, k8s.ExecRequest{
			Namespace: input.Namespace,
			Pod:       input.Pod,
			Container: input.Container,
			Command:   input.Command,
		})

		// Trace 완료 처리
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		var output ExecOutput
		output.Pod = input.Pod
		output.Command = commandStr
		output.RiskInfo = fmt.Sprintf("Risk Level: %s - %s", riskLevel, riskReason)

		if execErr == nil {
			output.Output = execOutput

			// 위험도가 critical이면 경고 메시지 추가
			if riskLevel == risk.RiskCritical {
				output.Warning = "⚠️  CRITICAL: Command execution in pods can modify container state, access sensitive data, or affect running processes. Review the output carefully."
			}
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
			// Output을 저장 (민감 정보 sanitize 적용)
			tr.Output = trace.SanitizeOutput(execOutput)
		}

		// Trace 저장
		if err := traceStore.Insert(tr); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save trace: %v\n", err)
		}

		// 에러 발생 시 반환
		if execErr != nil {
			return nil, ExecOutput{}, fmt.Errorf("failed to exec command in pod: %w", execErr)
		}

		return &mcp.CallToolResult{}, output, nil
	}
}

// GetExecToolDefinition은 sniff_exec Tool의 MCP Tool 정의를 반환합니다
func GetExecToolDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "sniff_exec",
		Description: "⚠️  Execute a command in a Kubernetes pod. This is a CRITICAL operation that can modify container state or access sensitive data. Use with extreme caution.",
	}
}
