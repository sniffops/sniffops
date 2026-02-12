package trace

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) (*Store, func()) {
	t.Helper()

	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func createTestTrace(sessionID string, toolName string) *Trace {
	return &Trace{
		ID:             uuid.New().String(),
		SessionID:      sessionID,
		Timestamp:      time.Now().UnixMilli(),
		UserIntent:     "Test user intent",
		ToolName:       toolName,
		Command:        "kubectl get pods -n default",
		TargetResource: "pod/*",
		Namespace:      "default",
		ResourceKind:   "pod",
		RiskLevel:      "low",
		RiskReason:     "Read-only operation",
		Result:         "success",
		Output:         "NAME    READY   STATUS\npod-1   1/1     Running",
		LatencyMs:      100,
		TokensInput:    50,
		TokensOutput:   30,
		CostEstimate:   0.001,
		Kubeconfig:     "~/.kube/config",
		ClusterName:    "test-cluster",
	}
}

func TestNewStore(t *testing.T) {
	t.Run("creates store with custom path", func(t *testing.T) {
		store, cleanup := setupTestDB(t)
		defer cleanup()

		if store == nil {
			t.Fatal("expected store to be created")
		}
	})

	t.Run("creates store with default path", func(t *testing.T) {
		// Use empty path to trigger default
		tmpDir := t.TempDir()
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)

		store, err := NewStore("")
		if err != nil {
			t.Fatalf("failed to create store with default path: %v", err)
		}
		defer store.Close()

		expectedPath := filepath.Join(tmpDir, ".sniffops", "traces.db")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected database at %s, but it doesn't exist", expectedPath)
		}
	})
}

func TestInsertAndGetByID(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	trace := createTestTrace("session-123", "sniff_get")

	// Insert trace
	err := store.Insert(trace)
	if err != nil {
		t.Fatalf("failed to insert trace: %v", err)
	}

	// Get by ID
	retrieved, err := store.GetByID(trace.ID)
	if err != nil {
		t.Fatalf("failed to get trace: %v", err)
	}

	// Verify fields
	if retrieved.ID != trace.ID {
		t.Errorf("expected ID %s, got %s", trace.ID, retrieved.ID)
	}
	if retrieved.SessionID != trace.SessionID {
		t.Errorf("expected SessionID %s, got %s", trace.SessionID, retrieved.SessionID)
	}
	if retrieved.ToolName != trace.ToolName {
		t.Errorf("expected ToolName %s, got %s", trace.ToolName, retrieved.ToolName)
	}
	if retrieved.Namespace != trace.Namespace {
		t.Errorf("expected Namespace %s, got %s", trace.Namespace, retrieved.Namespace)
	}
	if retrieved.RiskLevel != trace.RiskLevel {
		t.Errorf("expected RiskLevel %s, got %s", trace.RiskLevel, retrieved.RiskLevel)
	}
}

func TestInsertNilTrace(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	err := store.Insert(nil)
	if err == nil {
		t.Error("expected error when inserting nil trace")
	}
}

func TestGetByIDNotFound(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := store.GetByID("non-existent-id")
	if err == nil {
		t.Error("expected error when getting non-existent trace")
	}
}

func TestGetByIDEmptyID(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := store.GetByID("")
	if err == nil {
		t.Error("expected error when getting trace with empty ID")
	}
}

func TestList(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert multiple traces
	sessionID := "session-456"
	traces := []*Trace{
		createTestTrace(sessionID, "sniff_get"),
		createTestTrace(sessionID, "sniff_logs"),
		createTestTrace(sessionID, "sniff_apply"),
	}

	// Modify some fields for filtering
	traces[0].Namespace = "production"
	traces[0].RiskLevel = "low"
	traces[1].Namespace = "default"
	traces[1].RiskLevel = "low"
	traces[2].Namespace = "production"
	traces[2].RiskLevel = "medium"

	for _, trace := range traces {
		if err := store.Insert(trace); err != nil {
			t.Fatalf("failed to insert trace: %v", err)
		}
	}

	t.Run("list all traces", func(t *testing.T) {
		results, err := store.List(nil)
		if err != nil {
			t.Fatalf("failed to list traces: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 traces, got %d", len(results))
		}
	})

	t.Run("filter by tool", func(t *testing.T) {
		filter := &ListFilter{Tool: "sniff_get"}
		results, err := store.List(filter)
		if err != nil {
			t.Fatalf("failed to list traces: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 trace, got %d", len(results))
		}
		if results[0].ToolName != "sniff_get" {
			t.Errorf("expected tool sniff_get, got %s", results[0].ToolName)
		}
	})

	t.Run("filter by namespace", func(t *testing.T) {
		filter := &ListFilter{Namespace: "production"}
		results, err := store.List(filter)
		if err != nil {
			t.Fatalf("failed to list traces: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 traces, got %d", len(results))
		}
		for _, trace := range results {
			if trace.Namespace != "production" {
				t.Errorf("expected namespace production, got %s", trace.Namespace)
			}
		}
	})

	t.Run("filter by risk level", func(t *testing.T) {
		filter := &ListFilter{RiskLevel: "low"}
		results, err := store.List(filter)
		if err != nil {
			t.Fatalf("failed to list traces: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 traces, got %d", len(results))
		}
		for _, trace := range results {
			if trace.RiskLevel != "low" {
				t.Errorf("expected risk level low, got %s", trace.RiskLevel)
			}
		}
	})

	t.Run("filter by time range", func(t *testing.T) {
		now := time.Now()
		start := now.Add(-1 * time.Hour)
		end := now.Add(1 * time.Hour)

		filter := &ListFilter{
			StartTime: &start,
			EndTime:   &end,
		}
		results, err := store.List(filter)
		if err != nil {
			t.Fatalf("failed to list traces: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 traces within time range, got %d", len(results))
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		filter := &ListFilter{
			Namespace: "production",
			RiskLevel: "medium",
		}
		results, err := store.List(filter)
		if err != nil {
			t.Fatalf("failed to list traces: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 trace, got %d", len(results))
		}
		if results[0].ToolName != "sniff_apply" {
			t.Errorf("expected tool sniff_apply, got %s", results[0].ToolName)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		filter := &ListFilter{Limit: 2, Offset: 0}
		results, err := store.List(filter)
		if err != nil {
			t.Fatalf("failed to list traces: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 traces (page 1), got %d", len(results))
		}

		// Get next page
		filter.Offset = 2
		results, err = store.List(filter)
		if err != nil {
			t.Fatalf("failed to list traces: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 trace (page 2), got %d", len(results))
		}
	})
}

func TestCount(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert traces
	sessionID := "session-789"
	traces := []*Trace{
		createTestTrace(sessionID, "sniff_get"),
		createTestTrace(sessionID, "sniff_get"),
		createTestTrace(sessionID, "sniff_logs"),
	}

	traces[0].Namespace = "production"
	traces[1].Namespace = "default"
	traces[2].Namespace = "production"

	for _, trace := range traces {
		if err := store.Insert(trace); err != nil {
			t.Fatalf("failed to insert trace: %v", err)
		}
	}

	t.Run("count all traces", func(t *testing.T) {
		count, err := store.Count(nil)
		if err != nil {
			t.Fatalf("failed to count traces: %v", err)
		}

		if count != 3 {
			t.Errorf("expected count 3, got %d", count)
		}
	})

	t.Run("count with filter", func(t *testing.T) {
		filter := &ListFilter{Namespace: "production"}
		count, err := store.Count(filter)
		if err != nil {
			t.Fatalf("failed to count traces: %v", err)
		}

		if count != 2 {
			t.Errorf("expected count 2, got %d", count)
		}
	})
}

func TestOrderByTimestamp(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert traces with different timestamps
	sessionID := "session-order"
	now := time.Now()

	trace1 := createTestTrace(sessionID, "sniff_get")
	trace1.Timestamp = now.Add(-2 * time.Hour).UnixMilli()

	trace2 := createTestTrace(sessionID, "sniff_logs")
	trace2.Timestamp = now.Add(-1 * time.Hour).UnixMilli()

	trace3 := createTestTrace(sessionID, "sniff_apply")
	trace3.Timestamp = now.UnixMilli()

	for _, trace := range []*Trace{trace1, trace2, trace3} {
		if err := store.Insert(trace); err != nil {
			t.Fatalf("failed to insert trace: %v", err)
		}
	}

	// List should return newest first
	results, err := store.List(nil)
	if err != nil {
		t.Fatalf("failed to list traces: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 traces, got %d", len(results))
	}

	// Verify order (newest first)
	if results[0].ID != trace3.ID {
		t.Errorf("expected newest trace first, got %s", results[0].ID)
	}
	if results[1].ID != trace2.ID {
		t.Errorf("expected second newest trace, got %s", results[1].ID)
	}
	if results[2].ID != trace1.ID {
		t.Errorf("expected oldest trace last, got %s", results[2].ID)
	}
}
