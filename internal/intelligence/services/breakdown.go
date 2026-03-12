package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
)

// Subtask represents a proposed subtask from LLM breakdown.
type Subtask struct {
	Text           string `json:"text"`
	EffortEstimate string `json:"effort_estimate"`
}

// BreakdownResult holds the output of a task breakdown operation.
type BreakdownResult struct {
	ParentTaskID string
	ParentText   string
	Subtasks     []Subtask
	Backend      string
	GeneratedAt  time.Time
}

// BreakdownService wraps an LLM backend to break tasks into simple subtasks.
type BreakdownService struct {
	backend llm.LLMBackend
}

// NewBreakdownService creates a BreakdownService with the given LLM backend.
func NewBreakdownService(backend llm.LLMBackend) *BreakdownService {
	return &BreakdownService{backend: backend}
}

// Breakdown decomposes a task into subtasks via the LLM backend.
func (s *BreakdownService) Breakdown(ctx context.Context, taskID, taskDescription string) (*BreakdownResult, error) {
	if taskDescription == "" {
		return nil, fmt.Errorf("breakdown: task description must not be empty")
	}

	prompt := buildBreakdownPrompt(taskDescription)

	response, err := s.backend.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("breakdown via %s: %w", s.backend.Name(), err)
	}

	subtasks, err := parseSubtasks(response)
	if err != nil {
		return nil, fmt.Errorf("breakdown parse output from %s: %w", s.backend.Name(), err)
	}

	if len(subtasks) == 0 {
		return nil, fmt.Errorf("breakdown: LLM returned no subtasks")
	}

	return &BreakdownResult{
		ParentTaskID: taskID,
		ParentText:   taskDescription,
		Subtasks:     subtasks,
		Backend:      s.backend.Name(),
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func buildBreakdownPrompt(taskDescription string) string {
	return fmt.Sprintf(`You are a task decomposition assistant. Break the following task into smaller, actionable subtasks.

TASK:
%s

OUTPUT FORMAT:
Return a JSON array of subtask objects. Each subtask must have:
- "text": a clear, actionable subtask description
- "effort_estimate": one of "small", "medium", "large"

Aim for 3-7 subtasks. Each should be independently completable.
Return ONLY valid JSON, no markdown formatting, no explanation outside the JSON.`, taskDescription)
}

// parseSubtasks extracts subtasks from LLM output.
func parseSubtasks(raw string) ([]Subtask, error) {
	jsonStr := extractJSON(raw)

	var subtasks []Subtask
	if err := json.Unmarshal([]byte(jsonStr), &subtasks); err != nil {
		return nil, fmt.Errorf("parse subtasks: %w", err)
	}

	// Validate each subtask has text
	for i, st := range subtasks {
		if strings.TrimSpace(st.Text) == "" {
			return nil, fmt.Errorf("parse subtasks: subtask %d has empty text", i)
		}
	}

	return subtasks, nil
}

// extractJSON attempts to extract a JSON block from LLM output that may contain
// markdown code fences or surrounding text.
func extractJSON(raw string) string {
	if start := strings.Index(raw, "```json"); start != -1 {
		content := raw[start+7:]
		if end := strings.Index(content, "```"); end != -1 {
			return strings.TrimSpace(content[:end])
		}
	}
	if start := strings.Index(raw, "```"); start != -1 {
		content := raw[start+3:]
		if end := strings.Index(content, "```"); end != -1 {
			return strings.TrimSpace(content[:end])
		}
	}

	trimmed := strings.TrimSpace(raw)

	objStart := strings.Index(trimmed, "{")
	arrayStart := strings.Index(trimmed, "[")

	// Pick whichever JSON structure appears first.
	if objStart >= 0 && (arrayStart < 0 || objStart < arrayStart) {
		end := strings.LastIndexByte(trimmed, '}')
		if end > objStart {
			return trimmed[objStart : end+1]
		}
	}
	if arrayStart >= 0 {
		end := strings.LastIndexByte(trimmed, ']')
		if end > arrayStart {
			return trimmed[arrayStart : end+1]
		}
	}

	return trimmed
}
