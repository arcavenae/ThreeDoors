package mcpbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/mcp"
)

// commandTimeout is the maximum time allowed for a multiclaude CLI command.
const commandTimeout = 30 * time.Second

// registerTools wires up all MCP tool handlers for the bridge server.
func (s *BridgeServer) registerTools() map[string]ToolHandler {
	tools := map[string]ToolHandler{
		"multiclaude_status":       s.handleStatus,
		"multiclaude_worker_list":  s.handleWorkerList,
		"multiclaude_message_list": s.handleMessageList,
		"multiclaude_message_read": s.handleMessageRead,
		"multiclaude_repo_history": s.handleRepoHistory,
	}
	s.registerWriteTools(tools)
	return tools
}

// toolDefinitions returns the MCP tool definitions for tools/list.
func toolDefinitions() []mcp.ToolItem {
	defs := []mcp.ToolItem{
		{
			Name:        "multiclaude_status",
			Description: "Returns the current multiclaude system status including agents, their states, and worker count.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "multiclaude_worker_list",
			Description: "Returns the list of active workers with their names, tasks, and creation times.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "multiclaude_message_list",
			Description: "Returns pending inter-agent messages.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "multiclaude_message_read",
			Description: "Returns the full content of a specific message by its ID.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message_id": map[string]any{
						"type":        "string",
						"description": "The message ID to read",
					},
				},
				"required": []string{"message_id"},
			},
		},
		{
			Name:        "multiclaude_repo_history",
			Description: "Returns recent task history entries for the current repository.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"count": map[string]any{
						"type":        "integer",
						"description": "Number of history entries to return (default: 10)",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "Filter by task status (e.g., completed, failed)",
					},
				},
			},
		},
	}
	return append(defs, writeToolDefinitions()...)
}

func (s *BridgeServer) runMulticlaude(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	out, err := s.runner.Run(ctx, "multiclaude", args...)
	if err != nil {
		return "", fmt.Errorf("multiclaude %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *BridgeServer) handleStatus(_ json.RawMessage) (any, error) {
	out, err := s.runMulticlaude("status")
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *BridgeServer) handleWorkerList(_ json.RawMessage) (any, error) {
	out, err := s.runMulticlaude("worker", "list")
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *BridgeServer) handleMessageList(_ json.RawMessage) (any, error) {
	out, err := s.runMulticlaude("message", "list")
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *BridgeServer) handleMessageRead(params json.RawMessage) (any, error) {
	var args struct {
		MessageID string `json:"message_id"`
	}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &args); err != nil {
			return nil, fmt.Errorf("invalid parameters: %w", err)
		}
	}
	if args.MessageID == "" {
		return nil, fmt.Errorf("message_id is required")
	}

	out, err := s.runMulticlaude("message", "read", args.MessageID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *BridgeServer) handleRepoHistory(params json.RawMessage) (any, error) {
	var args struct {
		Count  int    `json:"count,omitempty"`
		Status string `json:"status,omitempty"`
	}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &args); err != nil {
			return nil, fmt.Errorf("invalid parameters: %w", err)
		}
	}

	cmdArgs := []string{"repo", "history"}
	if args.Count > 0 {
		cmdArgs = append(cmdArgs, "-n", fmt.Sprintf("%d", args.Count))
	}
	if args.Status != "" {
		cmdArgs = append(cmdArgs, "--status", args.Status)
	}

	out, err := s.runMulticlaude(cmdArgs...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
