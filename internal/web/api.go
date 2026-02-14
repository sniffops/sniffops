package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sniffops/sniffops/internal/trace"
)

// handleTraces handles GET /api/traces with filtering and pagination
func (s *Server) handleTraces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	filter := &trace.ListFilter{
		Tool:      query.Get("tool"),
		Namespace: query.Get("namespace"),
		RiskLevel: query.Get("risk"),
		Limit:     parseIntParam(query.Get("limit"), 50),
		Offset:    parseIntParam(query.Get("offset"), 0),
	}

	// Parse time range
	if startStr := query.Get("start"); startStr != "" {
		if startMs, err := strconv.ParseInt(startStr, 10, 64); err == nil {
			t := time.UnixMilli(startMs)
			filter.StartTime = &t
		}
	}

	if endStr := query.Get("end"); endStr != "" {
		if endMs, err := strconv.ParseInt(endStr, 10, 64); err == nil {
			t := time.UnixMilli(endMs)
			filter.EndTime = &t
		}
	}

	// Get traces
	traces, err := s.store.List(filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get total count
	total, err := s.store.Count(filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response
	response := map[string]interface{}{
		"traces": traces,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}

	respondJSON(w, http.StatusOK, response)
}

// handleTraceByID handles GET /api/traces/:id
func (s *Server) handleTraceByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract ID from path /api/traces/:id
	path := strings.TrimPrefix(r.URL.Path, "/api/traces/")
	if path == "" {
		respondError(w, http.StatusBadRequest, "trace ID required")
		return
	}

	trace, err := s.store.GetByID(path)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, trace)
}

// handleStats handles GET /api/stats
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse period parameter (optional, empty = all-time)
	period := r.URL.Query().Get("period")

	// Get statistics
	stats, err := trace.GetStats(s.store.DB(), period)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// handleNamespaces handles GET /api/namespaces
func (s *Server) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	namespaces, err := trace.GetDistinctValues(s.store.DB(), "namespace")
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, namespaces)
}

// handleTools handles GET /api/tools
func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tools, err := trace.GetDistinctValues(s.store.DB(), "tool_name")
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tools)
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func parseIntParam(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	if val < 0 {
		return defaultValue
	}
	return val
}
