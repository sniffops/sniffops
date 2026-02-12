package server

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sniffops/sniffops/internal/k8s"
	"github.com/sniffops/sniffops/internal/risk"
	"github.com/sniffops/sniffops/internal/tools"
	"github.com/sniffops/sniffops/internal/trace"
)

var (
	// sessionID는 프로세스 시작 시 한 번 생성되어 모든 trace에 사용됨
	sessionID = uuid.New().String()
)

// Server는 SniffOps MCP 서버를 나타냅니다
type Server struct {
	mcpServer     *mcp.Server
	sessionID     string
	k8sClient     *k8s.Client
	traceStore    *trace.Store
	riskEvaluator *risk.Evaluator
}

// Config는 서버 초기화 설정입니다
type Config struct {
	TraceDBPath string // SQLite 데이터베이스 경로 (비어있으면 기본 경로)
}

// New는 새로운 SniffOps MCP 서버를 생성합니다
func New(cfg *Config) (*Server, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	// 1. K8s client 초기화
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s client: %w", err)
	}

	// 2. Trace store 초기화
	traceStore, err := trace.NewStore(cfg.TraceDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace store: %w", err)
	}

	// 3. Risk evaluator 초기화
	riskEvaluator := risk.NewEvaluator()

	// 4. MCP 서버 생성
	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "sniffops",
			Version: "v0.1.0",
		},
		nil, // ServerOptions (로깅 등 추가 가능)
	)

	s := &Server{
		mcpServer:     mcpServer,
		sessionID:     sessionID,
		k8sClient:     k8sClient,
		traceStore:    traceStore,
		riskEvaluator: riskEvaluator,
	}

	// 5. Tool 등록
	s.registerTools()

	return s, nil
}

// registerTools는 모든 MCP Tool을 등록합니다
func (s *Server) registerTools() {
	tools.RegisterAllTools(
		s.mcpServer,
		s.k8sClient,
		s.traceStore,
		s.riskEvaluator,
		s.sessionID,
	)
}

// Run은 MCP 서버를 stdio transport로 시작합니다
func (s *Server) Run(ctx context.Context) error {
	// StdioTransport 생성 및 실행
	transport := &mcp.StdioTransport{}

	if err := s.mcpServer.Run(ctx, transport); err != nil {
		return fmt.Errorf("MCP server run failed: %w", err)
	}

	return nil
}

// Close는 서버 리소스를 정리합니다
func (s *Server) Close() error {
	if s.traceStore != nil {
		return s.traceStore.Close()
	}
	return nil
}

// GetSessionID는 현재 세션 ID를 반환합니다
func (s *Server) GetSessionID() string {
	return s.sessionID
}
