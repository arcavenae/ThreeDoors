package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
)

// EnrichedTask holds the result of LLM task enrichment.
// It contains both the original and enriched versions so the user
// can see exactly what changed (transparency per S4-D3).
type EnrichedTask struct {
	OriginalText string   `json:"original_text"`
	EnrichedText string   `json:"enriched_text"`
	Tags         []string `json:"tags"`
	Effort       int      `json:"effort"`
	Context      string   `json:"context"`
}

// Validate checks that the enriched task has minimum required fields.
func (e *EnrichedTask) Validate() error {
	if e.EnrichedText == "" {
		return fmt.Errorf("enriched text must not be empty")
	}
	if e.Effort < 0 || e.Effort > 5 {
		return fmt.Errorf("effort must be 0-5, got %d", e.Effort)
	}
	return nil
}

// TaskEnricher uses an LLM backend to add context, tags, and effort
// estimates to sparse tasks. User-initiated only (S1-D3).
type TaskEnricher struct {
	backend llm.LLMBackend
}

// NewTaskEnricher creates a TaskEnricher with the given LLM backend.
func NewTaskEnricher(backend llm.LLMBackend) *TaskEnricher {
	return &TaskEnricher{backend: backend}
}

// Enrich sends a task's text to the LLM for enrichment and returns
// the enriched result with tags, effort estimate, and context.
func (e *TaskEnricher) Enrich(ctx context.Context, taskText string) (*EnrichedTask, error) {
	taskText = strings.TrimSpace(taskText)
	if taskText == "" {
		return nil, fmt.Errorf("enrich: task text must not be empty")
	}

	prompt := buildEnrichmentPrompt(taskText)

	response, err := e.backend.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("enrich via %s: %w", e.backend.Name(), err)
	}

	result, err := parseEnrichmentResponse(response)
	if err != nil {
		return nil, fmt.Errorf("enrich parse response from %s: %w", e.backend.Name(), err)
	}

	result.OriginalText = taskText

	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("enrich validate result: %w", err)
	}

	return result, nil
}

// parseEnrichmentResponse extracts an EnrichedTask from the LLM's JSON response.
func parseEnrichmentResponse(raw string) (*EnrichedTask, error) {
	jsonStr := extractJSON(raw)

	var result EnrichedTask
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parse enrichment JSON: %w", err)
	}

	return &result, nil
}

// extractJSON is defined in breakdown.go — shared across services in this package.
