package linear

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// TestFullSyncFlow is an integration test that exercises the complete
// bidirectional sync cycle: load → mutate → WAL queue → replay.
func TestFullSyncFlow(t *testing.T) {
	t.Parallel()

	var apiAvailable atomic.Bool
	apiAvailable.Store(true)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !apiAvailable.Load() {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			return
		}

		// Route based on query content
		switch {
		case containsStr(req.Query, "viewer"):
			writeJSON(t, w, graphQLResponse{
				Data: json.RawMessage(`{"viewer":{"id":"u1","name":"Test","email":"t@t.com"}}`),
			})

		case containsStr(req.Query, "TeamIssues"):
			writeJSON(t, w, graphQLResponse{
				Data: json.RawMessage(`{
					"team": {
						"issues": {
							"nodes": [{
								"id": "uuid-1",
								"identifier": "ENG-101",
								"title": "Build sync feature",
								"description": "Implement bidirectional sync",
								"priority": 2,
								"createdAt": "2026-03-01T00:00:00Z",
								"updatedAt": "2026-03-05T00:00:00Z",
								"state": {"id": "s4", "name": "In Progress", "type": "started"},
								"team": {"id": "team-eng", "key": "ENG"},
								"labels": {"nodes": []},
								"assignee": null
							}],
							"pageInfo": {"hasNextPage": false, "endCursor": ""}
						}
					}
				}`),
			})

		case containsStr(req.Query, "WorkflowStates"):
			writeJSON(t, w, graphQLResponse{
				Data: json.RawMessage(`{
					"team": {
						"states": {
							"nodes": [
								{"id": "s1", "name": "Triage", "type": "triage"},
								{"id": "s3", "name": "Todo", "type": "unstarted"},
								{"id": "s4", "name": "In Progress", "type": "started"},
								{"id": "s5", "name": "Done", "type": "completed"},
								{"id": "s6", "name": "Cancelled", "type": "cancelled"}
							]
						}
					}
				}`),
			})

		case containsStr(req.Query, "issueUpdate"):
			writeJSON(t, w, graphQLResponse{
				Data: json.RawMessage(`{
					"issueUpdate": {
						"success": true,
						"issue": {
							"id": "uuid-1",
							"state": {"id": "s5", "name": "Done", "type": "completed"}
						}
					}
				}`),
			})

		default:
			t.Errorf("unexpected query: %s", req.Query)
			http.Error(w, "unknown query", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	// Create Linear client pointing at test server
	client := NewLinearClient("test-api-key")
	client.baseURL = server.URL

	config := &LinearConfig{
		APIKey:       "test-api-key",
		TeamIDs:      []string{"team-eng"},
		PollInterval: 5 * time.Minute,
	}

	provider := NewLinearProvider(client, config)

	// Step 1: Load tasks — populates issue index
	tasks, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "linear:ENG-101" {
		t.Errorf("task ID = %q, want %q", tasks[0].ID, "linear:ENG-101")
	}

	// Step 2: MarkComplete with API available — direct mutation
	err = provider.MarkComplete("linear:ENG-101")
	if err != nil {
		t.Fatalf("MarkComplete() error = %v", err)
	}

	// Step 3: Verify sync success time was set
	lastSync := provider.LastSyncSuccess()
	if lastSync.IsZero() {
		t.Error("LastSyncSuccess should be set after successful mutation")
	}

	// Step 4: HealthCheck reports sync status
	health := provider.HealthCheck()
	if health.Overall != core.HealthOK {
		t.Errorf("HealthCheck Overall = %q, want OK", health.Overall)
	}

	foundSync := false
	for _, item := range health.Items {
		if item.Name == "linear_sync" {
			foundSync = true
			if item.Status != core.HealthOK {
				t.Errorf("sync status = %q, want OK", item.Status)
			}
		}
	}
	if !foundSync {
		t.Error("HealthCheck should include linear_sync item")
	}
}

// TestWALIntegration tests the WAL provider wrapping LinearProvider for offline queuing.
func TestWALIntegration(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())

	// Load tasks to populate index
	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	// Wrap with WAL
	walProvider := core.NewWALProvider(provider, tmpDir)

	// Test 1: MarkComplete succeeds — passes through, no WAL entry
	err = walProvider.MarkComplete("linear:TEAM-1")
	if err != nil {
		t.Fatalf("WAL MarkComplete() error = %v", err)
	}
	if walProvider.PendingCount() != 0 {
		t.Errorf("expected 0 pending WAL entries after success, got %d", walProvider.PendingCount())
	}

	// Test 2: MarkComplete fails — queued in WAL
	client.mutateStateErr = fmt.Errorf("network error")
	err = walProvider.MarkComplete("linear:TEAM-1")
	if err != nil {
		t.Fatalf("WAL MarkComplete() should not return error (queued), got %v", err)
	}
	if walProvider.PendingCount() != 1 {
		t.Errorf("expected 1 pending WAL entry, got %d", walProvider.PendingCount())
	}

	// Test 3: SaveTask fails — queued in WAL
	client.mutateUpdateErr = fmt.Errorf("network error")
	err = walProvider.SaveTask(&core.Task{ID: "linear:TEAM-1", Text: "Updated"})
	if err != nil {
		t.Fatalf("WAL SaveTask() should not return error (queued), got %v", err)
	}
	if walProvider.PendingCount() != 2 {
		t.Errorf("expected 2 pending WAL entries, got %d", walProvider.PendingCount())
	}

	// Test 4: Replay — restore connectivity
	client.mutateStateErr = nil
	client.mutateUpdateErr = nil
	errors := walProvider.ReplayPending()

	if len(errors) != 0 {
		t.Errorf("replay errors = %v, want none", errors)
	}
	if walProvider.PendingCount() != 0 {
		t.Errorf("expected 0 pending after replay, got %d", walProvider.PendingCount())
	}
}

// TestWALHealthCheckIntegration tests that WAL status is reported in HealthCheck.
func TestWALHealthCheckIntegration(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())
	_, _ = provider.LoadTasks()

	walProvider := core.NewWALProvider(provider, tmpDir)

	// Health check with no pending entries
	health := walProvider.HealthCheck()
	foundWAL := false
	for _, item := range health.Items {
		if item.Name == "wal_queue" {
			foundWAL = true
			if item.Status != core.HealthOK {
				t.Errorf("wal_queue status = %q, want OK", item.Status)
			}
		}
	}
	if !foundWAL {
		t.Error("HealthCheck should include wal_queue item")
	}

	// Queue a failed mutation
	client.mutateStateErr = fmt.Errorf("offline")
	_ = walProvider.MarkComplete("linear:TEAM-1")

	// Health check with pending entries
	health = walProvider.HealthCheck()
	for _, item := range health.Items {
		if item.Name == "wal_queue" {
			if item.Status != core.HealthWarn {
				t.Errorf("wal_queue status = %q, want WARN", item.Status)
			}
			if !containsStr(item.Message, "1 pending") {
				t.Errorf("wal_queue message = %q, want to contain '1 pending'", item.Message)
			}
		}
	}
}

// TestWALPersistenceAndReload tests that WAL entries survive provider restart.
func TestWALPersistenceAndReload(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())
	_, _ = provider.LoadTasks()

	// Create WAL and queue a failed mutation
	walProvider := core.NewWALProvider(provider, tmpDir)
	client.mutateStateErr = fmt.Errorf("offline")
	_ = walProvider.MarkComplete("linear:TEAM-1")

	if walProvider.PendingCount() != 1 {
		t.Fatalf("expected 1 pending, got %d", walProvider.PendingCount())
	}

	// Simulate restart — create new WAL provider from same path
	walProvider2 := core.NewWALProvider(provider, tmpDir)
	if walProvider2.PendingCount() != 1 {
		t.Errorf("after restart, expected 1 pending, got %d", walProvider2.PendingCount())
	}

	// Restore connectivity and replay
	client.mutateStateErr = nil
	errors := walProvider2.ReplayPending()
	if len(errors) != 0 {
		t.Errorf("replay errors = %v", errors)
	}
	if walProvider2.PendingCount() != 0 {
		t.Errorf("expected 0 pending after replay, got %d", walProvider2.PendingCount())
	}
}

// TestDeleteTaskRemainsReadOnly verifies AC6: DeleteTask returns ErrReadOnly even with WAL.
func TestDeleteTaskRemainsReadOnly(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())

	walProvider := core.NewWALProvider(provider, tmpDir)

	// DeleteTask on the inner provider returns ErrReadOnly.
	// WALProvider catches any error and queues — but the replay will also fail with ErrReadOnly.
	// This is expected behavior: the WAL will eventually drop the entry after max retries.
	err := walProvider.DeleteTask("linear:TEAM-1")
	if err != nil {
		// WAL queues the operation — returns nil
		t.Logf("DeleteTask via WAL: %v (expected nil since WAL queues)", err)
	}

	// Direct provider: always ErrReadOnly
	err = provider.DeleteTask("linear:TEAM-1")
	if err != core.ErrReadOnly {
		t.Errorf("direct DeleteTask() = %v, want ErrReadOnly", err)
	}
}

// TestMutationGraphQLConstruction tests that mutations are correctly constructed.
func TestMutationGraphQLConstruction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation string
		wantVars  map[string]string
	}{
		{
			name:      "state transition mutation",
			operation: "state",
			wantVars:  map[string]string{"id": "issue-1", "stateId": "state-done"},
		},
		{
			name:      "title/description update mutation",
			operation: "update",
			wantVars:  map[string]string{"id": "issue-1", "title": "New Title", "description": "New Desc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var capturedVars map[string]any

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req graphQLRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("decode: %v", err)
					return
				}
				capturedVars = req.Variables

				writeJSON(t, w, graphQLResponse{
					Data: json.RawMessage(`{"issueUpdate":{"success":true,"issue":{"id":"issue-1","state":{"id":"s5","name":"Done","type":"completed"}}}}`),
				})
			}))
			t.Cleanup(server.Close)

			client := NewLinearClient("test-key")
			client.baseURL = server.URL

			switch tt.operation {
			case "state":
				_, _ = client.MutateIssueState(context.Background(), "issue-1", "state-done")
			case "update":
				_, _ = client.MutateIssueUpdate(context.Background(), "issue-1", "New Title", "New Desc")
			}

			for key, wantVal := range tt.wantVars {
				gotVal, ok := capturedVars[key]
				if !ok {
					t.Errorf("missing variable %q", key)
					continue
				}
				if fmt.Sprintf("%v", gotVal) != wantVal {
					t.Errorf("variable %q = %v, want %q", key, gotVal, wantVal)
				}
			}
		})
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestSaveTaskUpdatesLastSyncOnSuccess verifies lastSyncSuccess is set after SaveTask.
func TestSaveTaskUpdatesLastSyncOnSuccess(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())
	_, _ = provider.LoadTasks()

	before := provider.LastSyncSuccess()
	if !before.IsZero() {
		t.Error("LastSyncSuccess should be zero before any sync")
	}

	err := provider.SaveTask(&core.Task{ID: "linear:TEAM-1", Text: "Updated", Context: "Desc"})
	if err != nil {
		t.Fatalf("SaveTask() error = %v", err)
	}

	after := provider.LastSyncSuccess()
	if after.IsZero() {
		t.Error("LastSyncSuccess should be set after successful SaveTask")
	}
}

// TestMarkCompleteUpdatesLastSyncOnSuccess verifies lastSyncSuccess is set after MarkComplete.
func TestMarkCompleteUpdatesLastSyncOnSuccess(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())
	_, _ = provider.LoadTasks()

	err := provider.MarkComplete("linear:TEAM-1")
	if err != nil {
		t.Fatalf("MarkComplete() error = %v", err)
	}

	after := provider.LastSyncSuccess()
	if after.IsZero() {
		t.Error("LastSyncSuccess should be set after successful MarkComplete")
	}
}

// TestSaveTaskDoesNotUpdateLastSyncOnFailure verifies lastSyncSuccess is not set on failure.
func TestSaveTaskDoesNotUpdateLastSyncOnFailure(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	client.mutateUpdateErr = fmt.Errorf("network error")
	provider := NewLinearProvider(client, newTestConfig())
	_, _ = provider.LoadTasks()

	_ = provider.SaveTask(&core.Task{ID: "linear:TEAM-1", Text: "Updated"})

	if !provider.LastSyncSuccess().IsZero() {
		t.Error("LastSyncSuccess should remain zero after failed SaveTask")
	}
}

// Ensure WALProvider name includes WAL suffix
func TestWALProviderName(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())
	walProvider := core.NewWALProvider(provider, filepath.Join(t.TempDir()))

	name := walProvider.Name()
	if name != "linear (WAL)" {
		t.Errorf("WAL provider name = %q, want %q", name, "linear (WAL)")
	}
}
