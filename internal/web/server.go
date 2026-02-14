package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sniffops/sniffops/internal/trace"
)

// Server represents the HTTP API server for SniffOps Web UI
type Server struct {
	addr   string
	store  *trace.Store
	server *http.Server
}

// Config holds configuration for the web server
type Config struct {
	Port        int
	TraceDBPath string
}

// New creates a new web server instance
func New(cfg *Config) (*Server, error) {
	if cfg == nil {
		cfg = &Config{Port: 3000}
	}

	// Initialize trace store
	store, err := trace.NewStore(cfg.TraceDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize trace store: %w", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	s := &Server{
		addr:  addr,
		store: store,
	}

	// Setup routes
	mux := http.NewServeMux()
	s.setupRoutes(mux)

	// Create HTTP server
	s.server = &http.Server{
		Addr:         addr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

// setupRoutes registers all HTTP routes
func (s *Server) setupRoutes(mux *http.ServeMux) {
	// API endpoints
	mux.HandleFunc("/api/traces", s.handleTraces)
	mux.HandleFunc("/api/traces/", s.handleTraceByID)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/namespaces", s.handleNamespaces)
	mux.HandleFunc("/api/tools", s.handleTools)

	// Serve embedded frontend (fallback to static files)
	mux.Handle("/", http.FileServer(http.FS(DistFS)))
}

// Run starts the HTTP server
func (s *Server) Run(ctx context.Context) error {
	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Printf("Starting SniffOps Web UI on http://localhost%s", s.addr)
		log.Printf("API endpoints available at http://localhost%s/api/*", s.addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		log.Println("Shutting down web server...")
		// Graceful shutdown with 5 second timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

// Close closes the server and underlying resources
func (s *Server) Close() error {
	if s.store != nil {
		return s.store.Close()
	}
	return nil
}

// corsMiddleware adds CORS headers for development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow localhost for development (React dev server on 5173)
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
