package mcp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func dispatchToolCall(t *testing.T, s *MCPServer, name string, args any) *Response {
	t.Helper()
	var argsJSON json.RawMessage
	if args != nil {
		var err error
		argsJSON, err = json.Marshal(args)
		if err != nil {
			t.Fatalf("marshal args: %v", err)
		}
	}
	params, _ := json.Marshal(ToolCallParams{Name: name, Arguments: argsJSON})
	return s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/call",
		Params: params,
	})
}

func parseToolText(t *testing.T, resp *Response) string {
	t.Helper()
	if resp.Error != nil {
		t.Fatalf("unexpected error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
	}
	resultBytes, _ := json.Marshal(resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal tool result: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("empty tool content")
	}
	return result.Content[0].Text
}

func TestToolsListPopulated(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/list",
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ToolsListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.Tools) != 22 {
		t.Errorf("expected 22 tools, got %d", len(result.Tools))
	}

	names := make(map[string]bool)
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	expected := []string{
		"query_tasks", "get_task", "list_providers", "get_session", "search_tasks",
		"get_mood_correlation", "get_productivity_profile", "burnout_risk", "get_completions",
		"walk_graph", "find_paths", "get_critical_path", "get_orphans", "get_clusters",
		"get_provider_overlap", "get_unified_view", "suggest_cross_links",
		"prioritize_tasks", "analyze_workload", "focus_recommendation", "what_if", "context_switch_analysis",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestToolQueryTasks(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "buy milk", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
		&core.Task{ID: "2", Text: "write code", Status: core.StatusComplete, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "query_tasks", map[string]any{"status": "todo"})
	text := parseToolText(t, resp)

	var data struct {
		Tasks    []json.RawMessage `json:"tasks"`
		Metadata ResponseMetadata  `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.Tasks) != 1 {
		t.Errorf("expected 1 todo task, got %d", len(data.Tasks))
	}
	if data.Metadata.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", data.Metadata.TotalCount)
	}
}

func TestToolQueryTasksNoFilter(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "a", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
		&core.Task{ID: "2", Text: "b", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "query_tasks", nil)
	text := parseToolText(t, resp)

	var data struct {
		Tasks []json.RawMessage `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(data.Tasks))
	}
}

func TestToolGetTask(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task := &core.Task{
		ID: "abc", Text: "important task", Status: core.StatusTodo,
		Context: "some context", CreatedAt: now, UpdatedAt: now,
		Notes: []core.TaskNote{{Timestamp: now, Text: "a note"}},
	}
	s := newTestServerWithTasks(task)

	resp := dispatchToolCall(t, s, "get_task", map[string]string{"task_id": "abc"})
	text := parseToolText(t, resp)

	var data struct {
		Task struct {
			ID      string `json:"id"`
			Context string `json:"context"`
			Notes   []struct {
				Text string `json:"text"`
			} `json:"notes"`
		} `json:"task"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Task.ID != "abc" {
		t.Errorf("task id = %q, want abc", data.Task.ID)
	}
	if data.Task.Context != "some context" {
		t.Errorf("task context = %q, want 'some context'", data.Task.Context)
	}
	if len(data.Task.Notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(data.Task.Notes))
	}
}

func TestToolGetTaskNotFound(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "get_task", map[string]string{"task_id": "nonexistent"})

	resultBytes, _ := json.Marshal(resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.IsError {
		t.Error("expected isError=true for missing task")
	}
}

func TestToolGetTaskMissingID(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "get_task", map[string]string{})
	if resp.Error == nil {
		t.Fatal("expected error for missing task_id")
	}
}

func TestToolListProviders(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "list_providers", nil)
	text := parseToolText(t, resp)

	var data struct {
		Providers []json.RawMessage `json:"providers"`
		Metadata  ResponseMetadata  `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.Providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(data.Providers))
	}
}

func TestToolGetSessionCurrent(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "get_session", map[string]string{"type": "current"})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

func TestToolGetSessionHistory(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "get_session", map[string]string{"type": "history"})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

func TestToolSearchTasks(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "deploy kubernetes cluster", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
		&core.Task{ID: "2", Text: "buy groceries", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "search_tasks", map[string]any{"query": "kubernetes deploy"})
	text := parseToolText(t, resp)

	var data struct {
		Results  []json.RawMessage `json:"results"`
		Metadata ResponseMetadata  `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.Results) == 0 {
		t.Error("expected search results")
	}
}

func TestToolSearchTasksEmptyQuery(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "search_tasks", map[string]string{"query": ""})
	if resp.Error == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestToolUnknown(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "nonexistent_tool", nil)
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != CodeMethodNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeMethodNotFound)
	}
}
