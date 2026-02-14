package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/sniffops/sniffops/internal/server"
	"github.com/sniffops/sniffops/internal/web"
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
	var webPort int
	webCmd := &cobra.Command{
		Use:   "web",
		Short: "Start web UI server",
		Long:  "Start HTTP server to serve web-based trace viewer UI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWeb(webPort)
		},
	}

	webCmd.Flags().IntVarP(&webPort, "port", "p", 3000, "HTTP server port")

	rootCmd.AddCommand(serveCmd, webCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runServe starts the MCP server (stdio transport)
func runServe() error {
	// Context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nSniffOps MCP server shutting down...")
		cancel()
	}()

	// 서버 초기화
	cfg := &server.Config{
		TraceDBPath: "", // 빈 문자열 = 기본 경로 (~/.sniffops/traces.db)
	}

	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	defer srv.Close()

	fmt.Fprintf(os.Stderr, "SniffOps MCP server started (session: %s)\n", srv.GetSessionID())
	fmt.Fprintln(os.Stderr, "Registered tools: sniff_ping, sniff_get, sniff_logs")
	fmt.Fprintln(os.Stderr, "Trace database: ~/.sniffops/traces.db")
	fmt.Fprintln(os.Stderr, "Listening on stdio...")

	// MCP 서버 실행 (blocking)
	if err := srv.Run(ctx); err != nil {
		return fmt.Errorf("server run failed: %w", err)
	}

	return nil
}

// runWeb starts the web UI HTTP server
func runWeb(port int) error {
	// Context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nSniffOps web server shutting down...")
		cancel()
	}()

	// Initialize web server
	cfg := &web.Config{
		Port:        port,
		TraceDBPath: "", // Default path (~/.sniffops/traces.db)
	}

	srv, err := web.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}
	defer srv.Close()

	// Start server (blocking)
	if err := srv.Run(ctx); err != nil {
		return fmt.Errorf("web server failed: %w", err)
	}

	return nil
}
