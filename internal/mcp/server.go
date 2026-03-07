package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/core/metrics"
	"github.com/arcaven/ThreeDoors/internal/enrichment"
)

// Handler processes an MCP JSON-RPC request and returns a response.
type Handler func(req *Request) *Response

// Middleware wraps a Handler to add cross-cutting behavior (logging, auth, etc.).
type Middleware func(Handler) Handler

// MCPServer implements the Model Context Protocol server.
// It wraps existing ThreeDoors core components and exposes them via MCP.
type MCPServer struct {
	registry       *core.Registry
	aggregator     *core.MultiSourceAggregator
	pool           *core.TaskPool
	session        *core.SessionTracker
	enrichDB       *enrichment.DB
	sessionsReader *metrics.Reader
	middleware     []Middleware
	handler        Handler
	version        string

	initialized bool
}

// NewMCPServer creates an MCPServer wired to the given core components.
func NewMCPServer(
	registry *core.Registry,
	aggregator *core.MultiSourceAggregator,
	pool *core.TaskPool,
	session *core.SessionTracker,
	enrichDB *enrichment.DB,
	version string,
) *MCPServer {
	s := &MCPServer{
		registry:   registry,
		aggregator: aggregator,
		pool:       pool,
		session:    session,
		enrichDB:   enrichDB,
		version:    version,
	}
	s.handler = s.buildHandler()
	return s
}

// SetSessionsReader configures the reader for historical session data.
func (s *MCPServer) SetSessionsReader(r *metrics.Reader) {
	s.sessionsReader = r
}

// Use appends middleware to the server's middleware chain.
func (s *MCPServer) Use(mw Middleware) {
	s.middleware = append(s.middleware, mw)
	s.handler = s.buildHandler()
}

// HandleRequest processes a single JSON-RPC request and returns a response.
func (s *MCPServer) HandleRequest(raw []byte) ([]byte, error) {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		resp := NewErrorResponse(nil, CodeParseError, "parse error")
		return json.Marshal(resp)
	}

	if req.JSONRPC != jsonRPCVersion {
		resp := NewErrorResponse(req.ID, CodeInvalidRequest, "invalid jsonrpc version")
		return json.Marshal(resp)
	}

	// Notifications (no ID) are fire-and-forget.
	if req.ID == nil {
		s.handleNotification(&req)
		return nil, nil
	}

	resp := s.handler(&req)
	return json.Marshal(resp)
}

func (s *MCPServer) handleNotification(req *Request) {
	switch req.Method {
	case "notifications/initialized":
		s.initialized = true
	}
}

func (s *MCPServer) buildHandler() Handler {
	h := s.dispatch
	// Apply middleware in reverse order so the first-added runs outermost.
	for i := len(s.middleware) - 1; i >= 0; i-- {
		h = s.middleware[i](h)
	}
	return h
}

func (s *MCPServer) dispatch(req *Request) *Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourceRead(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolCall(req)
	case "prompts/list":
		return s.handlePromptsList(req)
	case "prompts/get":
		return s.handlePromptsGet(req)
	default:
		return NewErrorResponse(req.ID, CodeMethodNotFound,
			fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *MCPServer) handleInitialize(req *Request) *Response {
	var params InitializeParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid params: %v", err))
		}
	}

	result := InitializeResult{
		ProtocolVersion: MCPVersion,
		Capabilities: ServerCaps{
			Resources: &ResourcesCap{
				Subscribe:   true,
				ListChanged: true,
			},
			Tools:   &ToolsCap{},
			Prompts: &PromptsCap{ListChanged: true},
		},
		ServerInfo: EntityInfo{
			Name:    "threedoors-mcp",
			Version: s.version,
		},
	}
	return NewResponse(req.ID, result)
}

func (s *MCPServer) handleResourcesList(req *Request) *Response {
	result := ResourcesListResult{Resources: resourceDefinitions()}
	return NewResponse(req.ID, result)
}

func (s *MCPServer) handleToolsList(req *Request) *Response {
	result := ToolsListResult{Tools: toolDefinitions()}
	return NewResponse(req.ID, result)
}

func (s *MCPServer) handlePromptsList(req *Request) *Response {
	result := PromptsListResult{Prompts: promptDefinitions()}
	return NewResponse(req.ID, result)
}

func (s *MCPServer) handlePromptsGet(req *Request) *Response {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments,omitempty"`
	}
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid params: %v", err))
		}
	}

	template, ok := promptTemplates[params.Name]
	if !ok {
		return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("unknown prompt: %s", params.Name))
	}

	result := PromptGetResult{
		Description: template.Description,
		Messages: []PromptMessage{
			{
				Role: "user",
				Content: PromptContent{
					Type: "text",
					Text: template.Template,
				},
			},
		},
	}
	return NewResponse(req.ID, result)
}
