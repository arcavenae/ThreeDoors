package mcpbridge

import (
	"encoding/json"
	"fmt"

	"github.com/arcavenae/ThreeDoors/internal/mcp"
)

// BridgeServer implements an MCP server that wraps multiclaude CLI commands
// as MCP tools, enabling remote Claude Code sessions to query the multiclaude
// system programmatically via MCP-over-SSH.
type BridgeServer struct {
	runner  CommandRunner
	version string
	tools   map[string]ToolHandler
}

// ToolHandler processes an MCP tool call and returns the result content.
type ToolHandler func(params json.RawMessage) (any, error)

// NewBridgeServer creates a bridge MCP server that delegates to multiclaude
// CLI commands via the given runner.
func NewBridgeServer(runner CommandRunner, version string) *BridgeServer {
	s := &BridgeServer{
		runner:  runner,
		version: version,
	}
	s.tools = s.registerTools()
	return s
}

// HandleRequest processes a single JSON-RPC request and returns a response.
func (s *BridgeServer) HandleRequest(raw []byte) ([]byte, error) {
	var req mcp.Request
	if err := json.Unmarshal(raw, &req); err != nil {
		resp := mcp.NewErrorResponse(nil, mcp.CodeParseError, "parse error")
		return json.Marshal(resp)
	}

	if req.JSONRPC != "2.0" {
		resp := mcp.NewErrorResponse(req.ID, mcp.CodeInvalidRequest, "invalid jsonrpc version")
		return json.Marshal(resp)
	}

	// Notifications (no ID) are fire-and-forget.
	if req.ID == nil {
		return nil, nil
	}

	resp := s.dispatch(&req)
	return json.Marshal(resp)
}

func (s *BridgeServer) dispatch(req *mcp.Request) *mcp.Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolCall(req)
	default:
		return mcp.NewErrorResponse(req.ID, mcp.CodeMethodNotFound,
			fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *BridgeServer) handleInitialize(req *mcp.Request) *mcp.Response {
	result := mcp.InitializeResult{
		ProtocolVersion: mcp.MCPVersion,
		Capabilities: mcp.ServerCaps{
			Tools: &mcp.ToolsCap{},
		},
		ServerInfo: mcp.EntityInfo{
			Name:    "multiclaude-mcp-bridge",
			Version: s.version,
		},
	}
	return mcp.NewResponse(req.ID, result)
}

func (s *BridgeServer) handleToolsList(req *mcp.Request) *mcp.Response {
	result := mcp.ToolsListResult{Tools: toolDefinitions()}
	return mcp.NewResponse(req.ID, result)
}

func (s *BridgeServer) handleToolCall(req *mcp.Request) *mcp.Response {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments,omitempty"`
	}
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return mcp.NewErrorResponse(req.ID, mcp.CodeInvalidParams,
				fmt.Sprintf("invalid params: %v", err))
		}
	}

	handler, ok := s.tools[params.Name]
	if !ok {
		return mcp.NewErrorResponse(req.ID, mcp.CodeInvalidParams,
			fmt.Sprintf("unknown tool: %s", params.Name))
	}

	result, err := handler(params.Arguments)
	if err != nil {
		return mcp.NewResponse(req.ID, toolErrorResult(err.Error()))
	}

	return mcp.NewResponse(req.ID, toolSuccessResult(result))
}

// toolSuccessResult wraps a result value in the MCP tool call response format.
func toolSuccessResult(data any) map[string]any {
	text, ok := data.(string)
	if !ok {
		b, _ := json.Marshal(data)
		text = string(b)
	}
	return map[string]any{
		"content": []map[string]string{
			{"type": "text", "text": text},
		},
	}
}

// toolErrorResult wraps an error message in the MCP tool call error format.
func toolErrorResult(msg string) map[string]any {
	return map[string]any{
		"content": []map[string]string{
			{"type": "text", "text": msg},
		},
		"isError": true,
	}
}
