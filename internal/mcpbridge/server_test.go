package mcpbridge

import (
	"encoding/json"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/mcp"
)

func TestHandleRequest_Initialize(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0-test")

	req := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0"}}}`
	resp, err := server.HandleRequest([]byte(req))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	var response mcp.Response
	if err := json.Unmarshal(resp, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}

	var result mcp.InitializeResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if result.ServerInfo.Name != "multiclaude-mcp-bridge" {
		t.Errorf("server name = %q, want %q", result.ServerInfo.Name, "multiclaude-mcp-bridge")
	}
	if result.ServerInfo.Version != "1.0.0-test" {
		t.Errorf("version = %q, want %q", result.ServerInfo.Version, "1.0.0-test")
	}
	if result.ProtocolVersion != mcp.MCPVersion {
		t.Errorf("protocol version = %q, want %q", result.ProtocolVersion, mcp.MCPVersion)
	}
	if result.Capabilities.Tools == nil {
		t.Error("tools capability should be advertised")
	}
}

func TestHandleRequest_ToolsList(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	req := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	resp, err := server.HandleRequest([]byte(req))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	var response mcp.Response
	if err := json.Unmarshal(resp, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	resultBytes, _ := json.Marshal(response.Result)
	var result mcp.ToolsListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	expectedTools := map[string]bool{
		"multiclaude_status":       false,
		"multiclaude_worker_list":  false,
		"multiclaude_message_list": false,
		"multiclaude_message_read": false,
		"multiclaude_repo_history": false,
	}

	for _, tool := range result.Tools {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestHandleRequest_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	resp, err := server.HandleRequest([]byte("not json"))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	var response mcp.Response
	if err := json.Unmarshal(resp, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.Error == nil {
		t.Fatal("expected error response for invalid JSON")
	}
	if response.Error.Code != mcp.CodeParseError {
		t.Errorf("error code = %d, want %d", response.Error.Code, mcp.CodeParseError)
	}
}

func TestHandleRequest_InvalidVersion(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	req := `{"jsonrpc":"1.0","id":1,"method":"initialize"}`
	resp, err := server.HandleRequest([]byte(req))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	var response mcp.Response
	if err := json.Unmarshal(resp, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.Error == nil {
		t.Fatal("expected error for invalid jsonrpc version")
	}
	if response.Error.Code != mcp.CodeInvalidRequest {
		t.Errorf("error code = %d, want %d", response.Error.Code, mcp.CodeInvalidRequest)
	}
}

func TestHandleRequest_MethodNotFound(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	req := `{"jsonrpc":"2.0","id":1,"method":"nonexistent/method"}`
	resp, err := server.HandleRequest([]byte(req))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	var response mcp.Response
	if err := json.Unmarshal(resp, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if response.Error.Code != mcp.CodeMethodNotFound {
		t.Errorf("error code = %d, want %d", response.Error.Code, mcp.CodeMethodNotFound)
	}
}

func TestHandleRequest_Notification(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	// Notification has no ID field.
	req := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	resp, err := server.HandleRequest([]byte(req))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	if resp != nil {
		t.Errorf("notification should return nil response, got %s", resp)
	}
}

func TestHandleRequest_UnknownTool(t *testing.T) {
	t.Parallel()

	server := NewBridgeServer(&mockRunner{}, "1.0.0")

	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"unknown_tool"}}`
	resp, err := server.HandleRequest([]byte(req))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	var response mcp.Response
	if err := json.Unmarshal(resp, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
}
