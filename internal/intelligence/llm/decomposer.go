package llm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TaskDecomposer uses an LLM backend to break tasks into BMAD-style story specs.
type TaskDecomposer struct {
	backend LLMBackend
}

// NewTaskDecomposer creates a TaskDecomposer with the given LLM backend.
func NewTaskDecomposer(backend LLMBackend) *TaskDecomposer {
	return &TaskDecomposer{backend: backend}
}

// Decompose breaks a task description into structured story specs.
func (d *TaskDecomposer) Decompose(ctx context.Context, taskDescription string) (*DecompositionResult, error) {
	if taskDescription == "" {
		return nil, fmt.Errorf("decompose: task description must not be empty")
	}

	prompt := buildDecompositionPrompt(taskDescription)

	response, err := d.backend.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("decompose via %s: %w", d.backend.Name(), err)
	}

	// Extract JSON from response — LLMs often wrap JSON in markdown code blocks.
	jsonStr := extractJSON(response)

	specs, err := ParseStorySpecs(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("decompose parse output from %s: %w", d.backend.Name(), err)
	}

	for i := range specs {
		if err := specs[i].Validate(); err != nil {
			return nil, fmt.Errorf("decompose validate story %d: %w", i, err)
		}
	}

	return &DecompositionResult{
		SourceTask:  taskDescription,
		Stories:     specs,
		Backend:     d.backend.Name(),
		GeneratedAt: time.Now().UTC(),
	}, nil
}

func buildDecompositionPrompt(taskDescription string) string {
	return fmt.Sprintf(`You are a software project planner. Break the following task into implementable user stories following the BMAD methodology.

TASK:
%s

OUTPUT FORMAT:
Return a JSON array of story objects. Each story must have:
- "epic": the epic name/number this belongs to
- "story_id": a unique story identifier (e.g., "14.1", "14.2")
- "title": concise story title
- "user_story": "As a [role], I want [goal], So that [benefit]" format
- "acceptance_criteria": array of testable acceptance criteria strings
- "tasks": array of implementation task strings
- "dev_notes": technical notes for the developer

Return ONLY valid JSON, no markdown formatting, no explanation outside the JSON.`, taskDescription)
}

// extractJSON attempts to extract a JSON block from LLM output that may contain
// markdown code fences or surrounding text.
func extractJSON(raw string) string {
	// Try to find JSON within code fences.
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

	// Find the first [ or { and match to the last ] or }.
	trimmed := strings.TrimSpace(raw)
	arrayStart := strings.Index(trimmed, "[")
	objectStart := strings.Index(trimmed, "{")

	start := -1
	var closer byte
	if arrayStart >= 0 && (objectStart < 0 || arrayStart < objectStart) {
		start = arrayStart
		closer = ']'
	} else if objectStart >= 0 {
		start = objectStart
		closer = '}'
	}

	if start >= 0 {
		end := strings.LastIndexByte(trimmed, closer)
		if end > start {
			return trimmed[start : end+1]
		}
	}

	return trimmed
}
