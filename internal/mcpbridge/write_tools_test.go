package mcpbridge

import (
	"strings"
	"testing"
)

func TestToolMessageSend(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude message send supervisor Hello from remote": "msg-abc123-def4",
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_message_send", map[string]string{
		"recipient": "supervisor",
		"body":      "Hello from remote",
	})
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "msg-abc123") {
		t.Errorf("expected message ID in output, got: %s", text)
	}

	// Verify correct CLI args were passed.
	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(runner.calls))
	}
	args := runner.calls[0]
	expected := []string{"multiclaude", "message", "send", "supervisor", "Hello from remote"}
	if len(args) != len(expected) {
		t.Fatalf("args = %v, want %v", args, expected)
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("arg[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestToolMessageSend_MissingRecipient(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	response := callTool(t, server, "multiclaude_message_send", map[string]string{
		"body": "some message",
	})
	if !isToolError(t, response) {
		t.Error("expected tool error for missing recipient")
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "recipient") {
		t.Errorf("error should mention recipient, got: %s", text)
	}
}

func TestToolMessageSend_MissingBody(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	response := callTool(t, server, "multiclaude_message_send", map[string]string{
		"recipient": "supervisor",
	})
	if !isToolError(t, response) {
		t.Error("expected tool error for missing body")
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "body") {
		t.Errorf("error should mention body, got: %s", text)
	}
}

func TestToolMessageSend_ShellInjection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		recipient string
		body      string
	}{
		{"recipient semicolon", "agent;rm -rf /", "hello"},
		{"recipient pipe", "agent|cat /etc/passwd", "hello"},
		{"body semicolon", "supervisor", "hello; rm -rf /"},
		{"body subshell", "supervisor", "hello $(whoami)"},
		{"body backtick", "supervisor", "hello `id`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := NewBridgeServer(&mockRunner{}, "1.0.0")
			response := callTool(t, server, "multiclaude_message_send", map[string]string{
				"recipient": tt.recipient,
				"body":      tt.body,
			})
			if !isToolError(t, response) {
				t.Error("expected tool error for shell injection attempt")
			}

			text := extractToolText(t, response)
			if !strings.Contains(text, "prohibited character") && !strings.Contains(text, "must contain only") {
				t.Errorf("error should mention validation failure, got: %s", text)
			}
		})
	}
}

func TestToolMessageAck(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude message ack msg-abc123-def4": "acknowledged",
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_message_ack", map[string]string{
		"message_id": "msg-abc123-def4",
	})
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "acknowledged") {
		t.Errorf("expected acknowledgment in output, got: %s", text)
	}
}

func TestToolMessageAck_MissingID(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	response := callTool(t, server, "multiclaude_message_ack", map[string]string{})
	if !isToolError(t, response) {
		t.Error("expected tool error for missing message_id")
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "message_id") {
		t.Errorf("error should mention message_id, got: %s", text)
	}
}

func TestToolMessageAck_InvalidID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
	}{
		{"wrong prefix", "abc-123"},
		{"uppercase", "msg-ABC123"},
		{"shell injection", "msg-abc;rm -rf /"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := NewBridgeServer(&mockRunner{}, "1.0.0")
			response := callTool(t, server, "multiclaude_message_ack", map[string]string{
				"message_id": tt.id,
			})
			if !isToolError(t, response) {
				t.Errorf("expected tool error for invalid ID %q", tt.id)
			}
		})
	}
}

func TestToolWorkerCreate(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		output: map[string]string{
			"multiclaude worker create implement story 53.4": "jolly-dolphin",
		},
	}
	server := NewBridgeServer(runner, "1.0.0")

	response := callTool(t, server, "multiclaude_worker_create", map[string]string{
		"task": "implement story 53.4",
	})
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "jolly-dolphin") {
		t.Errorf("expected worker name in output, got: %s", text)
	}
}

func TestToolWorkerCreate_MissingTask(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	response := callTool(t, server, "multiclaude_worker_create", map[string]string{})
	if !isToolError(t, response) {
		t.Error("expected tool error for missing task")
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "task") {
		t.Errorf("error should mention task, got: %s", text)
	}
}

func TestToolWorkerCreate_TaskTooLong(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	response := callTool(t, server, "multiclaude_worker_create", map[string]string{
		"task": strings.Repeat("x", MaxTaskLength+1),
	})
	if !isToolError(t, response) {
		t.Error("expected tool error for task exceeding length limit")
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "exceeds maximum length") {
		t.Errorf("error should mention length limit, got: %s", text)
	}
}

func TestToolWorkerCreate_ShellInjection(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	response := callTool(t, server, "multiclaude_worker_create", map[string]string{
		"task": "implement $(rm -rf /)",
	})
	if !isToolError(t, response) {
		t.Error("expected tool error for shell injection attempt")
	}

	text := extractToolText(t, response)
	if !strings.Contains(text, "prohibited character") {
		t.Errorf("error should mention prohibited character, got: %s", text)
	}
}

func TestWriteToolsRegistered(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	writeTools := []string{
		"multiclaude_message_send",
		"multiclaude_message_ack",
		"multiclaude_worker_create",
	}
	for _, name := range writeTools {
		if _, ok := server.tools[name]; !ok {
			t.Errorf("write tool %q not registered", name)
		}
	}
}

func TestWriteToolDefinitions(t *testing.T) {
	t.Parallel()

	defs := toolDefinitions()
	expectedTools := map[string]bool{
		"multiclaude_message_send":  false,
		"multiclaude_message_ack":   false,
		"multiclaude_worker_create": false,
	}

	for _, tool := range defs {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("missing write tool definition: %s", name)
		}
	}
}
