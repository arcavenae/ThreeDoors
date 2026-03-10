package mcp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func newTestServer() *MCPServer {
	return NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")
}

func newTestServerWithTasks(tasks ...*core.Task) *MCPServer {
	pool := core.NewTaskPool()
	for _, t := range tasks {
		pool.AddTask(t)
	}
	return NewMCPServer(core.NewRegistry(), nil, pool, core.NewSessionTracker(), nil, "test")
}

func dispatchRead(t *testing.T, s *MCPServer, uri string) *Response {
	t.Helper()
	params, _ := json.Marshal(ResourceReadParams{URI: uri})
	return s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "resources/read",
		Params: params,
	})
}

func TestResourcesListPopulated(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "resources/list",
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourcesListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.Resources) != 16 {
		t.Errorf("expected 16 resources, got %d", len(result.Resources))
	}

	// Check that expected URIs are present.
	uris := make(map[string]bool)
	for _, r := range result.Resources {
		uris[r.URI] = true
	}
	expectedURIs := []string{
		"threedoors://tasks",
		"threedoors://tasks/{id}",
		"threedoors://tasks/status/{status}",
		"threedoors://tasks/provider/{name}",
		"threedoors://providers",
		"threedoors://providers/{name}/status",
		"threedoors://session/current",
		"threedoors://session/history",
	}
	for _, uri := range expectedURIs {
		if !uris[uri] {
			t.Errorf("missing resource URI: %s", uri)
		}
	}
}

func TestReadAllTasks(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task := &core.Task{ID: "t1", Text: "test task", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now}
	s := newTestServerWithTasks(task)

	resp := dispatchRead(t, s, "threedoors://tasks")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(result.Contents))
	}
	if result.Contents[0].MimeType != "application/json" {
		t.Errorf("expected application/json, got %s", result.Contents[0].MimeType)
	}

	// Parse the inner JSON to check metadata.
	var data struct {
		Tasks    []json.RawMessage `json:"tasks"`
		Metadata ResponseMetadata  `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &data); err != nil {
		t.Fatalf("unmarshal inner: %v", err)
	}
	if len(data.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(data.Tasks))
	}
	if data.Metadata.TotalCount != 1 {
		t.Errorf("expected total_count=1, got %d", data.Metadata.TotalCount)
	}
	if data.Metadata.ReturnedCount != 1 {
		t.Errorf("expected returned_count=1, got %d", data.Metadata.ReturnedCount)
	}
}

func TestReadTaskByID(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task := &core.Task{ID: "abc-123", Text: "find me", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now}
	s := newTestServerWithTasks(task)

	resp := dispatchRead(t, s, "threedoors://tasks/abc-123")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var data struct {
		Task struct {
			ID string `json:"id"`
		} `json:"task"`
	}
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &data); err != nil {
		t.Fatalf("unmarshal inner: %v", err)
	}
	if data.Task.ID != "abc-123" {
		t.Errorf("expected task abc-123, got %s", data.Task.ID)
	}
}

func TestReadTaskByIDNotFound(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchRead(t, s, "threedoors://tasks/nonexistent")
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent task")
		return
	}
}

func TestReadTasksByStatus(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "todo", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
		&core.Task{ID: "2", Text: "done", Status: core.StatusComplete, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchRead(t, s, "threedoors://tasks/status/todo")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var data struct {
		Tasks    []json.RawMessage `json:"tasks"`
		Metadata ResponseMetadata  `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &data); err != nil {
		t.Fatalf("unmarshal inner: %v", err)
	}
	if len(data.Tasks) != 1 {
		t.Errorf("expected 1 todo task, got %d", len(data.Tasks))
	}
	if data.Metadata.TotalCount != 2 {
		t.Errorf("expected total_count=2, got %d", data.Metadata.TotalCount)
	}
}

func TestReadTasksByStatusInvalid(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchRead(t, s, "threedoors://tasks/status/bogus")
	if resp.Error == nil {
		t.Fatal("expected error for invalid status")
		return
	}
}

func TestReadTasksByProvider(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task := &core.Task{ID: "1", Text: "task", Status: core.StatusTodo, SourceProvider: "textfile", CreatedAt: now, UpdatedAt: now}
	task.MigrateSourceProvider()
	s := newTestServerWithTasks(task)

	resp := dispatchRead(t, s, "threedoors://tasks/provider/textfile")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var data struct {
		Tasks []json.RawMessage `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &data); err != nil {
		t.Fatalf("unmarshal inner: %v", err)
	}
	if len(data.Tasks) != 1 {
		t.Errorf("expected 1 task from textfile provider, got %d", len(data.Tasks))
	}
}

func TestReadProviders(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchRead(t, s, "threedoors://providers")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var data struct {
		Providers []json.RawMessage `json:"providers"`
		Metadata  ResponseMetadata  `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &data); err != nil {
		t.Fatalf("unmarshal inner: %v", err)
	}
	// Empty registry should return empty list.
	if len(data.Providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(data.Providers))
	}
}

func TestReadCurrentSession(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchRead(t, s, "threedoors://session/current")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var data struct {
		SessionID string           `json:"session_id"`
		Metadata  ResponseMetadata `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &data); err != nil {
		t.Fatalf("unmarshal inner: %v", err)
	}
	if data.SessionID == "" {
		t.Error("expected non-empty session_id")
	}
	if data.Metadata.QueryTimeMs < 0 {
		t.Error("expected non-negative query_time_ms")
	}
}

func TestReadSessionHistory(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchRead(t, s, "threedoors://session/history")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var data struct {
		Sessions []json.RawMessage `json:"sessions"`
		Metadata ResponseMetadata  `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &data); err != nil {
		t.Fatalf("unmarshal inner: %v", err)
	}
	// No reader set, should be empty.
	if len(data.Sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(data.Sessions))
	}
}

func TestReadUnknownURI(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchRead(t, s, "threedoors://unknown")
	if resp.Error == nil {
		t.Fatal("expected error for unknown URI")
		return
	}
}

func TestReadMissingURI(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "resources/read",
		Params: json.RawMessage(`{}`),
	})
	if resp.Error == nil {
		t.Fatal("expected error for missing URI")
		return
	}
}

func TestMetadataFieldsPresent(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "task", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchRead(t, s, "threedoors://tasks")
	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var data struct {
		Metadata ResponseMetadata `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &data); err != nil {
		t.Fatalf("unmarshal inner: %v", err)
	}

	m := data.Metadata
	if m.TotalCount != 1 {
		t.Errorf("total_count = %d, want 1", m.TotalCount)
	}
	if m.ReturnedCount != 1 {
		t.Errorf("returned_count = %d, want 1", m.ReturnedCount)
	}
	if m.QueryTimeMs < 0 {
		t.Error("query_time_ms should be non-negative")
	}
	if m.DataFreshness != "live" {
		t.Errorf("data_freshness = %q, want live", m.DataFreshness)
	}
	if m.ProvidersQueried == nil {
		t.Error("providers_queried should not be nil")
	}
}
