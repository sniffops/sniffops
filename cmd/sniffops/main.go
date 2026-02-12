package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "v0.1.0"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "sniffops",
		Short:   "SniffOps - AI-driven K8s observability MCP server",
		Long:    "SniffOps tracks and analyzes all K8s operations performed by AI agents through MCP protocol.",
		Version: version,
	}

	// serve 명령어 - MCP 서버 시작 (stdio)
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start MCP server (stdio mode)",
		Long:  "Start SniffOps MCP server. This command is called by Claude Code automatically.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe()
		},
	}

	// web 명령어 - 웹 UI HTTP 서버 시작
	webCmd := &cobra.Command{
		Use:   "web",
		Short: "Start web UI server",
		Long:  "Start HTTP server to serve web-based trace viewer UI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWeb()
		},
	}

	webCmd.Flags().IntP("port", "p", 3000, "HTTP server port")

	rootCmd.AddCommand(serveCmd, webCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runServe starts the MCP server (stdio transport)
func runServe() error {
	// TODO: MCP 서버 초기화
	// 1. SQLite DB 초기화
	// 2. K8s client 초기화
	// 3. MCP 서버 생성 및 Tool 등록
	// 4. stdio 통신 시작
	fmt.Fprintln(os.Stderr, "SniffOps MCP server starting...")
	fmt.Fprintln(os.Stderr, "MCP server implementation: TODO")
	return fmt.Errorf("MCP server not implemented yet")
}

// runWeb starts the web UI HTTP server
func runWeb() error {
	// TODO: 웹 UI 서버 시작
	// 1. SQLite DB 연결
	// 2. HTTP 핸들러 등록 (/api/traces, /api/stats)
	// 3. 임베디드 React UI 서빙
	fmt.Fprintln(os.Stderr, "SniffOps web UI server starting...")
	fmt.Fprintln(os.Stderr, "Web server implementation: TODO")
	return fmt.Errorf("web server not implemented yet")
}
