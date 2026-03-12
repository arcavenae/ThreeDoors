package mcpbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/mcp"
)

// mockRunner records commands and returns configured output.
type mockRunner struct {
	output map[string]string
	err    map[string]error
	calls  [][]string
}

func (m *mockRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	m.calls = append(m.calls, append([]string{name}, args...))

	if m.err != nil {
		if err, ok := m.err[key]; ok {
			return nil, err
		}
	}
	if m.output != nil {
		if out, ok := m.output[key]; ok {
			return []byte(out), nil
		}
	}
	return []byte(""), nil
}

func callTool(t *testing.T, server *BridgeServer, toolName string, args any) mcp.Response {
	t.Helper()

	params := map[string]any{"name": toolName}
	if args != nil {
		argsJSON, err := json.Marshal(args)
		if err != nil {
			t.Fatalf("marshal args: %v", err)
		}
		params["arguments"] = json.RawMessage(argsJSON)
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}

	req := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":%s}`, paramsJSON)
	resp, err := server.HandleRequest([]byte(req))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	var response mcp.Response
	if err := json.Unmarshal(resp, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return response
}

func extractToolText(t *testing.T, response mcp.Response) string {
	t.Helper()

	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal tool result: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("tool result has no content")
	}
	return result.Content[0].Text
}

func isToolError(t *testing.T, response mcp.Response) bool {
	t.Helper()

	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}

	var result struct {
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal tool result: %v", err)
	}
	return result.IsError
}

func TestToolStatus(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude status": "Daemon: running\nAgents: 3 active\nWorkers: 2 active",
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_status", nil)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "Daemon: running") {
		t.Errorf("status output missing expected content, got: %s", text)
	}
}

func TestToolWorkerList(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude worker list": "worker-1  implement story 53.1  2026-03-12T10:00:00Z\nworker-2  fix CI  2026-03-12T11:00:00Z",
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_worker_list", nil)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "worker-1") {
		t.Errorf("worker list missing expected content, got: %s", text)
	}
}

func TestToolMessageList(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude message list": "msg-001  supervisor  Need help with task\nmsg-002  merge-queue  PR ready to merge",
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_message_list", nil)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "msg-001") {
		t.Errorf("message list missing expected content, got: %s", text)
	}
}

func TestToolMessageRead(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude message read msg-001": "From: supervisor\nTo: worker-1\nBody: Please check the CI failure on PR #42",
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_message_read", map[string]string{
		"message_id": "msg-001",
	})
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "Please check the CI failure") {
		t.Errorf("message read missing expected content, got: %s", text)
	}
}

func TestToolMessageRead_MissingID(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	response := callTool(t, server, "multiclaude_message_read", map[string]string{})
	if !isToolError(t, response) {
		t.Error("expected tool error for missing message_id")
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "message_id is required") {
		t.Errorf("error message should mention message_id, got: %s", text)
	}
}

func TestToolRepoHistory(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude repo history": "task-001  completed  implement story 53.1\ntask-002  failed  fix CI pipeline",
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_repo_history", nil)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "task-001") {
		t.Errorf("repo history missing expected content, got: %s", text)
	}
}

func TestToolRepoHistory_WithParams(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude repo history -n 5 --status completed": "task-001  completed  implement story 53.1",
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_repo_history", map[string]any{
		"count":  5,
		"status": "completed",
	})
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	// Verify the correct command was called.
	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(runner.calls))
	}
	args := runner.calls[0]
	expected := []string{"multiclaude", "repo", "history", "-n", "5", "--status", "completed"}
	if len(args) != len(expected) {
		t.Fatalf("args = %v, want %v", args, expected)
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("arg[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestToolCommandFailure(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		err: map[string]error{
			"multiclaude status": fmt.Errorf("command multiclaude failed: exit status 1: daemon not running"),
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_status", nil)
	if response.Error != nil {
		t.Fatalf("tool errors should be in result, not JSON-RPC error: %v", response.Error)
	}

	if !isToolError(t, response) {
		t.Error("expected isError=true for failed command")
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "daemon not running") {
		t.Errorf("error text should contain stderr, got: %s", text)
	}
}
