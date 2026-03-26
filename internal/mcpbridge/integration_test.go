package mcpbridge

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/mcp"
)

// TestIntegration_FullSession simulates a complete MCP session: initialize,
// list tools, call each tool, and verify structured responses.
func TestIntegration_FullSession(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude status":                                  "Daemon: running\nAgents: supervisor, merge-queue, pr-shepherd\nWorkers: 2 active",
			"multiclaude worker list":                             "zealous-squirrel  implement story 53.3  2026-03-12T10:00:00Z",
			"multiclaude message list":                            "msg-abc  supervisor  Check PR status",
			"multiclaude message read msg-abc":                    "From: supervisor\nTo: merge-queue\nBody: Check PR #42 status and merge if green",
			"multiclaude repo history -n 3":                       "task-001  completed  story 53.1\ntask-002  completed  story 53.2\ntask-003  in_progress  story 53.3",
			"multiclaude message send merge-queue PR 42 is green": "msg-def456",
			"multiclaude message ack msg-abc":                     "acknowledged",
			"multiclaude worker create implement story 53.4":      "jolly-dolphin",
		},
	}
	server := NewBridgeServer(runner, "1.0.0-integration")

	// Step 1: Initialize.
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"claude-code","version":"1.0.0"}}}`
	initResp := mustRequest(t, server, initReq)
	if initResp.Error != nil {
		t.Fatalf("initialize failed: %v", initResp.Error)
	}

	// Step 2: Send initialized notification.
	notifReq := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	resp, err := server.HandleRequest([]byte(notifReq))
	if err != nil {
		t.Fatalf("notification: %v", err)
	}
	if resp != nil {
		t.Error("notification should return nil")
	}

	// Step 3: List tools.
	toolsReq := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	toolsResp := mustRequest(t, server, toolsReq)
	if toolsResp.Error != nil {
		t.Fatalf("tools/list failed: %v", toolsResp.Error)
	}

	resultBytes, _ := json.Marshal(toolsResp.Result)
	var toolsList mcp.ToolsListResult
	if err := json.Unmarshal(resultBytes, &toolsList); err != nil {
		t.Fatalf("unmarshal tools list: %v", err)
	}
	if len(toolsList.Tools) != 8 {
		t.Errorf("expected 8 tools, got %d", len(toolsList.Tools))
	}

	// Step 4: Call each tool and verify output.
	tests := []struct {
		name     string
		tool     string
		args     string
		contains string
	}{
		{
			name:     "status",
			tool:     "multiclaude_status",
			args:     `{}`,
			contains: "Daemon: running",
		},
		{
			name:     "worker_list",
			tool:     "multiclaude_worker_list",
			args:     `{}`,
			contains: "zealous-squirrel",
		},
		{
			name:     "message_list",
			tool:     "multiclaude_message_list",
			args:     `{}`,
			contains: "msg-abc",
		},
		{
			name:     "message_read",
			tool:     "multiclaude_message_read",
			args:     `{"message_id":"msg-abc"}`,
			contains: "Check PR #42",
		},
		{
			name:     "repo_history",
			tool:     "multiclaude_repo_history",
			args:     `{"count":3}`,
			contains: "task-001",
		},
		{
			name:     "message_send",
			tool:     "multiclaude_message_send",
			args:     `{"recipient":"merge-queue","body":"PR 42 is green"}`,
			contains: "msg-def456",
		},
		{
			name:     "message_ack",
			tool:     "multiclaude_message_ack",
			args:     `{"message_id":"msg-abc"}`,
			contains: "acknowledged",
		},
		{
			name:     "worker_create",
			tool:     "multiclaude_worker_create",
			args:     `{"task":"implement story 53.4"}`,
			contains: "jolly-dolphin",
		},
	}

	for i, tt := range tests {
		reqJSON := buildToolCallRequest(t, i+10, tt.tool, tt.args)
		toolResp := mustRequest(t, server, reqJSON)
		if toolResp.Error != nil {
			t.Errorf("[%s] unexpected JSON-RPC error: %v", tt.name, toolResp.Error)
			continue
		}

		text := extractToolText(t, toolResp)
		if !strings.Contains(text, tt.contains) {
			t.Errorf("[%s] output missing %q, got: %s", tt.name, tt.contains, text)
		}
	}
}

func mustRequest(t *testing.T, server *BridgeServer, reqJSON string) mcp.Response {
	t.Helper()

	resp, err := server.HandleRequest([]byte(reqJSON))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	var response mcp.Response
	if err := json.Unmarshal(resp, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return response
}

func buildToolCallRequest(t *testing.T, id int, toolName, argsJSON string) string {
	t.Helper()

	params := map[string]any{
		"name":      toolName,
		"arguments": json.RawMessage(argsJSON),
	}
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  "tools/call",
		"params":  json.RawMessage(paramsBytes),
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return string(reqBytes)
}
