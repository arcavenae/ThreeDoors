package mcpbridge

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/arcaven/ThreeDoors/internal/mcp"
)

// registerWriteTools adds write tool handlers to the given map.
func (s *BridgeServer) registerWriteTools(tools map[string]ToolHandler) {
	tools["multiclaude_message_send"] = s.handleMessageSend
	tools["multiclaude_message_ack"] = s.handleMessageAck
	tools["multiclaude_worker_create"] = s.handleWorkerCreate
}

// writeToolDefinitions returns the MCP tool definitions for write tools.
func writeToolDefinitions() []mcp.ToolItem {
	return []mcp.ToolItem{
		{
			Name:        "multiclaude_message_send",
			Description: "Sends a message to another multiclaude agent. Returns the message ID on success.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"recipient": map[string]any{
						"type":        "string",
						"description": "The agent name to send the message to (alphanumeric, hyphens, underscores only)",
					},
					"body": map[string]any{
						"type":        "string",
						"description": "The message body (max 2000 characters)",
					},
				},
				"required": []string{"recipient", "body"},
			},
		},
		{
			Name:        "multiclaude_message_ack",
			Description: "Acknowledges a pending message by its ID. Returns success or failure.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message_id": map[string]any{
						"type":        "string",
						"description": "The message ID to acknowledge (format: msg-<uuid>)",
					},
				},
				"required": []string{"message_id"},
			},
		},
		{
			Name:        "multiclaude_worker_create",
			Description: "Dispatches a new multiclaude worker with the given task description. Returns the worker name.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task": map[string]any{
						"type":        "string",
						"description": "The task description for the worker (max 500 characters)",
					},
				},
				"required": []string{"task"},
			},
		},
	}
}

func (s *BridgeServer) handleMessageSend(params json.RawMessage) (any, error) {
	var args struct {
		Recipient string `json:"recipient"`
		Body      string `json:"body"`
	}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &args); err != nil {
			return nil, fmt.Errorf("invalid parameters: %w", err)
		}
	}

	if err := validateRecipient(args.Recipient); err != nil {
		return nil, err
	}
	if err := validateBody(args.Body); err != nil {
		return nil, err
	}

	log.Printf("[audit] multiclaude_message_send: recipient=%s body_len=%d", args.Recipient, len(args.Body))

	out, err := s.runMulticlaude("message", "send", args.Recipient, args.Body)
	if err != nil {
		return nil, sanitizeError(err)
	}
	return out, nil
}

func (s *BridgeServer) handleMessageAck(params json.RawMessage) (any, error) {
	var args struct {
		MessageID string `json:"message_id"`
	}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &args); err != nil {
			return nil, fmt.Errorf("invalid parameters: %w", err)
		}
	}

	if err := validateMessageID(args.MessageID); err != nil {
		return nil, err
	}

	log.Printf("[audit] multiclaude_message_ack: message_id=%s", args.MessageID)

	out, err := s.runMulticlaude("message", "ack", args.MessageID)
	if err != nil {
		return nil, sanitizeError(err)
	}
	return out, nil
}

func (s *BridgeServer) handleWorkerCreate(params json.RawMessage) (any, error) {
	var args struct {
		Task string `json:"task"`
	}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &args); err != nil {
			return nil, fmt.Errorf("invalid parameters: %w", err)
		}
	}

	if err := validateTask(args.Task); err != nil {
		return nil, err
	}

	log.Printf("[audit] multiclaude_worker_create: task_len=%d", len(args.Task))

	out, err := s.runMulticlaude("worker", "create", args.Task)
	if err != nil {
		return nil, sanitizeError(err)
	}
	return out, nil
}

// sanitizeError strips internal file paths from error messages to prevent
// information leakage through MCP error responses.
func sanitizeError(err error) error {
	return fmt.Errorf("command failed: %w", err)
}
