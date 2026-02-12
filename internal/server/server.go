package server

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sniffops/sniffops/internal/tools"
)

var (
	// sessionID는 프로세스 시작 시 한 번 생성되어 모든 trace에 사용됨
	sessionID = uuid.New().String()
)

// Server는 SniffOps MCP 서버를 나타냅니다
type Server struct {
	mcpServer *mcp.Server
	sessionID string
}

// Config는 서버 초기화 설정입니다
type Config struct {
	// 향후 DB, K8s client 등 설정 추가 예정
}

// New는 새로운 SniffOps MCP 서버를 생성합니다
func New(cfg *Config) (*Server, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	// MCP 서버 생성
	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "sniffops",
			Version: "v0.1.0",
		},
		nil, // ServerOptions (로깅 등 추가 가능)
	)

	s := &Server{
		mcpServer: mcpServer,
		sessionID: sessionID,
	}

	// Tool 등록
	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return s, nil
}

// registerTools는 모든 MCP Tool을 등록합니다
func (s *Server) registerTools() error {
	// sniff_ping Tool 등록 (Hello World)
	mcp.AddTool(
		s.mcpServer,
		tools.GetPingToolDefinition(),
		tools.PingHandler,
	)

	return nil
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

// GetSessionID는 현재 세션 ID를 반환합니다
func (s *Server) GetSessionID() string {
	return s.sessionID
}
