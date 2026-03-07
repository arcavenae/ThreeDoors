package mcp

import "encoding/json"

// JSON-RPC 2.0 protocol types for the Model Context Protocol.

const jsonRPCVersion = "2.0"

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// Notification represents a JSON-RPC 2.0 notification (no ID).
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Standard JSON-RPC error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// Application-level error codes.
const (
	CodeRateLimited = -32000
	CodeReadOnly    = -32001
)

// MCP protocol version and capability types.
const (
	MCPVersion = "2024-11-05"
)

// InitializeParams is sent by the client during the initialize handshake.
type InitializeParams struct {
	ProtocolVersion string     `json:"protocolVersion"`
	Capabilities    ClientCaps `json:"capabilities"`
	ClientInfo      EntityInfo `json:"clientInfo"`
}

// ClientCaps represents client capabilities declared during initialization.
type ClientCaps struct {
	Roots    *RootsCap    `json:"roots,omitempty"`
	Sampling *SamplingCap `json:"sampling,omitempty"`
}

// RootsCap indicates the client supports filesystem roots.
type RootsCap struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCap indicates the client supports LLM sampling.
type SamplingCap struct{}

// EntityInfo identifies a client or server.
type EntityInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult is the server's response to the initialize request.
type InitializeResult struct {
	ProtocolVersion string     `json:"protocolVersion"`
	Capabilities    ServerCaps `json:"capabilities"`
	ServerInfo      EntityInfo `json:"serverInfo"`
}

// ServerCaps represents the server's advertised capabilities.
type ServerCaps struct {
	Resources *ResourcesCap `json:"resources,omitempty"`
	Tools     *ToolsCap     `json:"tools,omitempty"`
	Prompts   *PromptsCap   `json:"prompts,omitempty"`
}

// ResourcesCap advertises resource capabilities.
type ResourcesCap struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCap advertises tool capabilities.
type ToolsCap struct{}

// PromptsCap advertises prompt capabilities.
type PromptsCap struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourceItem describes a single resource in the resources/list response.
type ResourceItem struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourcesListResult is the response to resources/list.
type ResourcesListResult struct {
	Resources []ResourceItem `json:"resources"`
}

// ToolItem describes a single tool in the tools/list response.
type ToolItem struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	InputSchema any    `json:"inputSchema"`
}

// ToolsListResult is the response to tools/list.
type ToolsListResult struct {
	Tools []ToolItem `json:"tools"`
}

// PromptItem describes a single prompt in the prompts/list response.
type PromptItem struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument describes a prompt argument.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptsListResult is the response to prompts/list.
type PromptsListResult struct {
	Prompts []PromptItem `json:"prompts"`
}

// PromptContent holds the content of a prompt message.
type PromptContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// PromptMessage is a single message in a prompt response.
type PromptMessage struct {
	Role    string        `json:"role"`
	Content PromptContent `json:"content"`
}

// PromptGetResult is the response to prompts/get.
type PromptGetResult struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// NewResponse creates a successful JSON-RPC response.
func NewResponse(id json.RawMessage, result any) *Response {
	return &Response{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse creates an error JSON-RPC response.
func NewErrorResponse(id json.RawMessage, code int, message string) *Response {
	return &Response{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
}
