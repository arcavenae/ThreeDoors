package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
)

// DefaultMaxInputSize is the maximum input size for extraction (32KB).
const DefaultMaxInputSize = 32 * 1024

// DefaultDuplicateThreshold is the minimum similarity score to flag a duplicate.
const DefaultDuplicateThreshold = 0.8

// ErrEmptyInput is returned when the input text is empty.
var ErrEmptyInput = fmt.Errorf("extraction input is empty")

// ErrInputTooLarge is returned when the input exceeds the size limit.
var ErrInputTooLarge = fmt.Errorf("extraction input too large")

// ErrParseFailed is returned when the LLM response cannot be parsed as JSON.
var ErrParseFailed = fmt.Errorf("extraction response parse failed")

// ExtractedTask represents a task extracted from unstructured text by an LLM.
type ExtractedTask struct {
	Text       string   `json:"text"`
	Effort     int      `json:"effort"`
	Tags       []string `json:"tags"`
	Source     string   `json:"-"`
	Confidence float64  `json:"confidence,omitempty"`
	Duplicate  bool     `json:"-"`
}

// ExtractorOption configures a TaskExtractor.
type ExtractorOption func(*TaskExtractor)

// WithMaxInputSize sets the maximum input size in bytes.
func WithMaxInputSize(size int) ExtractorOption {
	return func(e *TaskExtractor) {
		if size > 0 {
			e.maxInputSize = size
		}
	}
}

// WithRunner sets the CLIRunner for clipboard access.
func WithRunner(runner llm.CLIRunner) ExtractorOption {
	return func(e *TaskExtractor) {
		e.runner = runner
	}
}

// WithDuplicateChecker sets a function to check for duplicates against existing tasks.
func WithDuplicateChecker(checker DuplicateChecker) ExtractorOption {
	return func(e *TaskExtractor) {
		e.dedupChecker = checker
	}
}

// DuplicateChecker checks if an extracted task text is a potential duplicate
// of an existing task. Returns true and the similarity score if duplicate.
type DuplicateChecker func(text string) (isDuplicate bool, similarity float64)

// TaskExtractor extracts actionable tasks from unstructured text using an LLM.
type TaskExtractor struct {
	backend      llm.LLMBackend
	runner       llm.CLIRunner
	maxInputSize int
	dedupChecker DuplicateChecker
}

// NewTaskExtractor creates a TaskExtractor with the given backend and options.
func NewTaskExtractor(backend llm.LLMBackend, opts ...ExtractorOption) *TaskExtractor {
	e := &TaskExtractor{
		backend:      backend,
		maxInputSize: DefaultMaxInputSize,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// ExtractFromText extracts tasks from the given text blob.
func (e *TaskExtractor) ExtractFromText(ctx context.Context, text string) ([]ExtractedTask, error) {
	if err := e.validateInput(text); err != nil {
		return nil, err
	}

	prompt := buildExtractionPrompt(text)
	response, err := e.backend.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("extract tasks: %w", err)
	}

	tasks, err := parseExtractionResponse(response)
	if err != nil {
		// Retry once with stricter prompt.
		retryPrompt := buildRetryPrompt(text)
		response, retryErr := e.backend.Complete(ctx, retryPrompt)
		if retryErr != nil {
			return nil, fmt.Errorf("extract tasks retry: %w", retryErr)
		}
		tasks, err = parseExtractionResponse(response)
		if err != nil {
			return nil, fmt.Errorf("parse extraction response after retry: %w", ErrParseFailed)
		}
	}

	// Set source and check duplicates.
	for i := range tasks {
		tasks[i].Source = "text"
		if e.dedupChecker != nil {
			isDup, _ := e.dedupChecker(tasks[i].Text)
			tasks[i].Duplicate = isDup
		}
	}

	return tasks, nil
}

// ExtractFromFile reads a file and extracts tasks from its content.
func (e *TaskExtractor) ExtractFromFile(ctx context.Context, path string) ([]ExtractedTask, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}

	tasks, err := e.ExtractFromText(ctx, string(content))
	if err != nil {
		return nil, err
	}

	for i := range tasks {
		tasks[i].Source = fmt.Sprintf("file:%s", path)
	}

	return tasks, nil
}

// ExtractFromClipboard reads the system clipboard and extracts tasks.
func (e *TaskExtractor) ExtractFromClipboard(ctx context.Context) ([]ExtractedTask, error) {
	if e.runner == nil {
		return nil, fmt.Errorf("extract from clipboard: no command runner configured")
	}

	stdout, stderr, err := e.runner.RunWithStdin(ctx, "", "pbpaste")
	if err != nil {
		return nil, fmt.Errorf("read clipboard: %s: %w", stderr, err)
	}

	tasks, err := e.ExtractFromText(ctx, stdout)
	if err != nil {
		return nil, err
	}

	for i := range tasks {
		tasks[i].Source = "clipboard"
	}

	return tasks, nil
}

// validateInput checks that the input text is non-empty and within size limits.
func (e *TaskExtractor) validateInput(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("validate input: %w", ErrEmptyInput)
	}
	if len(text) > e.maxInputSize {
		return fmt.Errorf("validate input: %d bytes exceeds %d byte limit: %w",
			len(text), e.maxInputSize, ErrInputTooLarge)
	}
	return nil
}

// parseExtractionResponse parses the LLM response as a JSON array of ExtractedTask.
func parseExtractionResponse(response string) ([]ExtractedTask, error) {
	response = strings.TrimSpace(response)

	// Strip markdown code fences if present.
	response = stripCodeFences(response)

	var tasks []ExtractedTask
	if err := json.Unmarshal([]byte(response), &tasks); err != nil {
		return nil, fmt.Errorf("unmarshal tasks: %w", err)
	}

	return tasks, nil
}

// stripCodeFences removes ```json ... ``` wrapping if present.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Remove opening fence line.
		if idx := strings.Index(s, "\n"); idx >= 0 {
			s = s[idx+1:]
		}
		// Remove closing fence.
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}
