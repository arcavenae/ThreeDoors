package mcp

import (
	"encoding/json"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestHandleInitialize(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test-v1")

	params := InitializeParams{
		ProtocolVersion: MCPVersion,
		Capabilities:    ClientCaps{},
		ClientInfo:      EntityInfo{Name: "test-client", Version: "1.0"},
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
		return
	}

	reqID, _ := json.Marshal(1)
	req := Request{
		JSONRPC: jsonRPCVersion,
		ID:      reqID,
		Method:  "initialize",
		Params:  paramsJSON,
	}
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
		return
	}

	respBytes, err := server.HandleRequest(raw)
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
		return
	}

	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
		return
	}

	var result InitializeResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if result.ProtocolVersion != MCPVersion {
		t.Errorf("protocol version = %q, want %q", result.ProtocolVersion, MCPVersion)
	}
	if result.ServerInfo.Name != "threedoors-mcp" {
		t.Errorf("server name = %q, want %q", result.ServerInfo.Name, "threedoors-mcp")
	}
	if result.ServerInfo.Version != "test-v1" {
		t.Errorf("server version = %q, want %q", result.ServerInfo.Version, "test-v1")
	}
}

func TestCapabilityAdvertisement(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	resp := server.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "initialize",
	})

	resultBytes, _ := json.Marshal(resp.Result)
	var result InitializeResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	caps := result.Capabilities
	if caps.Resources == nil {
		t.Fatal("resources capability is nil")
		return
	}
	if !caps.Resources.Subscribe {
		t.Error("resources.subscribe should be true")
	}
	if !caps.Resources.ListChanged {
		t.Error("resources.listChanged should be true")
	}
	if caps.Tools == nil {
		t.Fatal("tools capability is nil")
		return
	}
	if caps.Prompts == nil {
		t.Fatal("prompts capability is nil")
		return
	}
	if !caps.Prompts.ListChanged {
		t.Error("prompts.listChanged should be true")
	}
}

func TestResourcesListReturnsDefinitions(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	resp := server.dispatch(&Request{
		ID:     json.RawMessage(`2`),
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
	if len(result.Resources) == 0 {
		t.Error("expected resources to be populated")
	}
}

func TestToolsListReturnsDefinitions(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	resp := server.dispatch(&Request{
		ID:     json.RawMessage(`3`),
		Method: "tools/list",
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ToolsListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result.Tools) == 0 {
		t.Error("expected tools to be populated")
	}
}

func TestPromptsListEmpty(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	resp := server.dispatch(&Request{
		ID:     json.RawMessage(`4`),
		Method: "prompts/list",
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}
}

func TestMethodNotFound(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	resp := server.dispatch(&Request{
		ID:     json.RawMessage(`5`),
		Method: "nonexistent/method",
	})

	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
		return
	}
	if resp.Error.Code != CodeMethodNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeMethodNotFound)
	}
}

func TestInvalidJSON(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	respBytes, err := server.HandleRequest([]byte(`{not valid json`))
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
		return
	}

	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected parse error")
		return
	}
	if resp.Error.Code != CodeParseError {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeParseError)
	}
}

func TestInvalidJSONRPCVersion(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	req := Request{
		JSONRPC: "1.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	}
	raw, _ := json.Marshal(req)

	respBytes, err := server.HandleRequest(raw)
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
		return
	}

	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected invalid request error")
		return
	}
	if resp.Error.Code != CodeInvalidRequest {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidRequest)
	}
}

func TestNotificationNoResponse(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	req := Request{
		JSONRPC: jsonRPCVersion,
		Method:  "notifications/initialized",
	}
	raw, _ := json.Marshal(req)

	respBytes, err := server.HandleRequest(raw)
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
		return
	}
	if respBytes != nil {
		t.Errorf("expected nil response for notification, got %s", respBytes)
	}
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	var called bool
	server.Use(func(next Handler) Handler {
		return func(req *Request) *Response {
			called = true
			return next(req)
		}
	})

	req := Request{
		JSONRPC: jsonRPCVersion,
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	}
	raw, _ := json.Marshal(req)
	if _, err := server.HandleRequest(raw); err != nil {
		t.Fatalf("HandleRequest: %v", err)
	}

	if !called {
		t.Error("middleware was not called")
	}
}
