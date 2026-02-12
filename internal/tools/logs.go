package tools

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sniffops/sniffops/internal/k8s"
	"github.com/sniffops/sniffops/internal/risk"
	"github.com/sniffops/sniffops/internal/trace"
)

// LogsInput은 sniff_logs Tool의 입력입니다
type LogsInput struct {
	Namespace string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Pod       string `json:"pod" jsonschema:"Pod name"`
	Container string `json:"container,omitempty" jsonschema:"Container name (optional; uses first container if omitted)"`
	Lines     int64  `json:"lines,omitempty" jsonschema:"Number of log lines to retrieve (default: 100)"`
}

// LogsOutput은 sniff_logs Tool의 출력입니다
type LogsOutput struct {
	Logs  string `json:"logs" jsonschema:"Pod logs as text"`
	Lines int    `json:"lines" jsonschema:"Number of lines returned"`
}

// LogsHandler는 sniff_logs Tool의 핸들러입니다
//
// 이 Tool은 Kubernetes Pod의 로그를 조회합니다.
// - Trace 기록 및 위험도 평가 수행
func LogsHandler(
	k8sClient *k8s.Client,
	traceStore *trace.Store,
	riskEvaluator *risk.Evaluator,
	sessionID string,
) mcp.ToolHandlerFor[LogsInput, LogsOutput] {
	return func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input LogsInput,
	) (*mcp.CallToolResult, LogsOutput, error) {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			return nil, LogsOutput{}, ctx.Err()
		default:
		}

		// Trace 시작
		startTime := time.Now()
		traceID := uuid.New().String()

		// Default lines to 100
		if input.Lines <= 0 {
			input.Lines = 100
		}

		// Build command string
		command := fmt.Sprintf("kubectl logs -n %s %s", input.Namespace, input.Pod)
		if input.Container != "" {
			command += fmt.Sprintf(" -c %s", input.Container)
		}
		command += fmt.Sprintf(" --tail=%d", input.Lines)

		// User intent 생성
		userIntent := fmt.Sprintf("Get logs from pod %s in namespace %s", input.Pod, input.Namespace)
		if input.Container != "" {
			userIntent = fmt.Sprintf("Get logs from container %s in pod %s (namespace %s)", input.Container, input.Pod, input.Namespace)
		}

		// 초기 trace 레코드 생성
		tr := &trace.Trace{
			ID:             traceID,
			SessionID:      sessionID,
			Timestamp:      startTime.UnixMilli(),
			UserIntent:     userIntent,
			ToolName:       "sniff_logs",
			Command:        command,
			Namespace:      input.Namespace,
			ResourceKind:   "Pod",
			TargetResource: input.Pod,
		}

		// K8s API 호출 (Pod 로그 조회)
		logs, execErr := k8sClient.Logs(ctx, k8s.LogsRequest{
			Namespace: input.Namespace,
			Pod:       input.Pod,
			Container: input.Container,
			Lines:     input.Lines,
		})

		// Trace 완료 처리
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// 위험도 평가
		riskLevel, riskReason := riskEvaluator.Evaluate(risk.EvalContext{
			ToolName:     "sniff_logs",
			Namespace:    input.Namespace,
			ResourceKind: "Pod",
		})

		// Trace 레코드 완성
		tr.LatencyMs = int(duration.Milliseconds())
		tr.RiskLevel = string(riskLevel)
		tr.RiskReason = riskReason

		var output LogsOutput
		if execErr != nil {
			tr.Result = "failure"
			tr.ErrorMessage = execErr.Error()
		} else {
			tr.Result = "success"
			tr.Output = trace.SanitizeOutput(logs) // 로그에서 민감 정보 sanitize
			output.Logs = logs
			// Count lines (rough estimate)
			output.Lines = len(logs)
			if len(logs) > 0 {
				lineCount := 0
				for _, c := range logs {
					if c == '\n' {
						lineCount++
					}
				}
				output.Lines = lineCount
			}
		}

		// Trace 저장
		if err := traceStore.Insert(tr); err != nil {
			// Trace 저장 실패는 로깅만 하고 Tool 실행은 계속
			fmt.Fprintf(os.Stderr, "Warning: failed to save trace: %v\n", err)
		}

		// 에러 발생 시 반환
		if execErr != nil {
			return nil, LogsOutput{}, fmt.Errorf("failed to get pod logs: %w", execErr)
		}

		return &mcp.CallToolResult{}, output, nil
	}
}

// GetLogsToolDefinition은 sniff_logs Tool의 MCP Tool 정의를 반환합니다
func GetLogsToolDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "sniff_logs",
		Description: "Get Kubernetes pod logs. Retrieves recent log lines from a pod. Optionally specify container name and number of lines.",
	}
}
