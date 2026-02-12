package trace

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // SQLite driver (CGO-free)
)

// Store handles SQLite operations for trace storage
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store with the given database path
// If dbPath is empty, uses default: ~/.sniffops/traces.db
func NewStore(dbPath string) (*Store, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(home, ".sniffops", "traces.db")
	}

	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &Store{db: db}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// initSchema creates the traces and metadata tables if they don't exist
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS traces (
		-- Identity
		id              TEXT PRIMARY KEY,
		session_id      TEXT NOT NULL,
		timestamp       INTEGER NOT NULL,
		
		-- Request Context
		user_intent     TEXT,
		tool_name       TEXT NOT NULL,
		
		-- K8s Command Details
		command         TEXT NOT NULL,
		target_resource TEXT,
		namespace       TEXT,
		resource_kind   TEXT,
		
		-- Risk & Security
		risk_level      TEXT NOT NULL,
		risk_reason     TEXT,
		
		-- Execution Result
		result          TEXT NOT NULL,
		output          TEXT,
		error_message   TEXT,
		
		-- Metrics
		latency_ms      INTEGER,
		tokens_input    INTEGER,
		tokens_output   INTEGER,
		cost_estimate   REAL,
		
		-- Metadata
		kubeconfig      TEXT,
		cluster_name    TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_session_id ON traces(session_id);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON traces(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_namespace ON traces(namespace);
	CREATE INDEX IF NOT EXISTS idx_risk_level ON traces(risk_level);
	CREATE INDEX IF NOT EXISTS idx_tool_name ON traces(tool_name);

	CREATE TABLE IF NOT EXISTS metadata (
		key   TEXT PRIMARY KEY,
		value TEXT
	);

	INSERT OR IGNORE INTO metadata (key, value) VALUES ('schema_version', '1');
	INSERT OR IGNORE INTO metadata (key, value) VALUES ('created_at', datetime('now'));
	`

	_, err := s.db.Exec(schema)
	return err
}

// Insert saves a new trace to the database
func (s *Store) Insert(trace *Trace) error {
	if trace == nil {
		return fmt.Errorf("trace cannot be nil")
	}

	query := `
	INSERT INTO traces (
		id, session_id, timestamp,
		user_intent, tool_name,
		command, target_resource, namespace, resource_kind,
		risk_level, risk_reason,
		result, output, error_message,
		latency_ms, tokens_input, tokens_output, cost_estimate,
		kubeconfig, cluster_name
	) VALUES (
		?, ?, ?,
		?, ?,
		?, ?, ?, ?,
		?, ?,
		?, ?, ?,
		?, ?, ?, ?,
		?, ?
	)
	`

	_, err := s.db.Exec(query,
		trace.ID, trace.SessionID, trace.Timestamp,
		trace.UserIntent, trace.ToolName,
		trace.Command, trace.TargetResource, trace.Namespace, trace.ResourceKind,
		trace.RiskLevel, trace.RiskReason,
		trace.Result, trace.Output, trace.ErrorMessage,
		trace.LatencyMs, trace.TokensInput, trace.TokensOutput, trace.CostEstimate,
		trace.Kubeconfig, trace.ClusterName,
	)

	if err != nil {
		return fmt.Errorf("failed to insert trace: %w", err)
	}

	return nil
}

// GetByID retrieves a single trace by ID
func (s *Store) GetByID(id string) (*Trace, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	query := `
	SELECT 
		id, session_id, timestamp,
		user_intent, tool_name,
		command, target_resource, namespace, resource_kind,
		risk_level, risk_reason,
		result, output, error_message,
		latency_ms, tokens_input, tokens_output, cost_estimate,
		kubeconfig, cluster_name
	FROM traces
	WHERE id = ?
	`

	trace := &Trace{}
	err := s.db.QueryRow(query, id).Scan(
		&trace.ID, &trace.SessionID, &trace.Timestamp,
		&trace.UserIntent, &trace.ToolName,
		&trace.Command, &trace.TargetResource, &trace.Namespace, &trace.ResourceKind,
		&trace.RiskLevel, &trace.RiskReason,
		&trace.Result, &trace.Output, &trace.ErrorMessage,
		&trace.LatencyMs, &trace.TokensInput, &trace.TokensOutput, &trace.CostEstimate,
		&trace.Kubeconfig, &trace.ClusterName,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("trace not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	return trace, nil
}

// List retrieves traces with optional filters
func (s *Store) List(filter *ListFilter) ([]*Trace, error) {
	if filter == nil {
		filter = &ListFilter{}
	}

	// Default limit
	if filter.Limit <= 0 {
		filter.Limit = 100
	}

	// Build query with filters
	query := `
	SELECT 
		id, session_id, timestamp,
		user_intent, tool_name,
		command, target_resource, namespace, resource_kind,
		risk_level, risk_reason,
		result, output, error_message,
		latency_ms, tokens_input, tokens_output, cost_estimate,
		kubeconfig, cluster_name
	FROM traces
	WHERE 1=1
	`

	var conditions []string
	var args []interface{}

	// Apply filters
	if filter.Tool != "" {
		conditions = append(conditions, "tool_name = ?")
		args = append(args, filter.Tool)
	}

	if filter.Namespace != "" {
		conditions = append(conditions, "namespace = ?")
		args = append(args, filter.Namespace)
	}

	if filter.RiskLevel != "" {
		conditions = append(conditions, "risk_level = ?")
		args = append(args, filter.RiskLevel)
	}

	if filter.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, filter.StartTime.UnixMilli())
	}

	if filter.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, filter.EndTime.UnixMilli())
	}

	// Add conditions to query
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Order by timestamp (newest first)
	query += " ORDER BY timestamp DESC"

	// Add limit and offset
	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces: %w", err)
	}
	defer rows.Close()

	// Scan results
	var traces []*Trace
	for rows.Next() {
		trace := &Trace{}
		err := rows.Scan(
			&trace.ID, &trace.SessionID, &trace.Timestamp,
			&trace.UserIntent, &trace.ToolName,
			&trace.Command, &trace.TargetResource, &trace.Namespace, &trace.ResourceKind,
			&trace.RiskLevel, &trace.RiskReason,
			&trace.Result, &trace.Output, &trace.ErrorMessage,
			&trace.LatencyMs, &trace.TokensInput, &trace.TokensOutput, &trace.CostEstimate,
			&trace.Kubeconfig, &trace.ClusterName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trace: %w", err)
		}
		traces = append(traces, trace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating traces: %w", err)
	}

	return traces, nil
}

// Count returns the total number of traces matching the filter
func (s *Store) Count(filter *ListFilter) (int, error) {
	if filter == nil {
		filter = &ListFilter{}
	}

	query := "SELECT COUNT(*) FROM traces WHERE 1=1"
	var conditions []string
	var args []interface{}

	// Apply same filters as List
	if filter.Tool != "" {
		conditions = append(conditions, "tool_name = ?")
		args = append(args, filter.Tool)
	}

	if filter.Namespace != "" {
		conditions = append(conditions, "namespace = ?")
		args = append(args, filter.Namespace)
	}

	if filter.RiskLevel != "" {
		conditions = append(conditions, "risk_level = ?")
		args = append(args, filter.RiskLevel)
	}

	if filter.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, filter.StartTime.UnixMilli())
	}

	if filter.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, filter.EndTime.UnixMilli())
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count traces: %w", err)
	}

	return count, nil
}
