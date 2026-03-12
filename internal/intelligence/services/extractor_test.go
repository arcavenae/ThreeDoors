package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockRunner implements llm.CLIRunner for testing clipboard.
type mockRunner struct {
	stdout string
	stderr string
	err    error
}

func (m *mockRunner) RunWithStdin(_ context.Context, _ string, _ string, _ ...string) (string, string, error) {
	return m.stdout, m.stderr, m.err
}

func TestNewTaskExtractor(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{name: "test"}
	e := NewTaskExtractor(backend)

	if e.backend != backend {
		t.Error("expected backend to be set")
	}
	if e.maxInputSize != DefaultMaxInputSize {
		t.Errorf("expected maxInputSize=%d, got %d", DefaultMaxInputSize, e.maxInputSize)
	}
}

func TestNewTaskExtractorWithOptions(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{name: "test"}
	runner := &mockRunner{}
	e := NewTaskExtractor(backend,
		WithMaxInputSize(1024),
		WithRunner(runner),
	)

	if e.maxInputSize != 1024 {
		t.Errorf("expected maxInputSize=1024, got %d", e.maxInputSize)
	}
	if e.runner != runner {
		t.Error("expected runner to be set")
	}
}

func TestExtractFromText(t *testing.T) {
	t.Parallel()

	validJSON := `[
		{"text": "Email Sarah about budget", "effort": 1, "tags": ["communication"]},
		{"text": "Prep slides for demo", "effort": 3, "tags": ["presentation"]}
	]`

	tests := []struct {
		name      string
		input     string
		response  string
		wantCount int
		wantErr   bool
		errTarget error
	}{
		{
			name:      "valid extraction",
			input:     "Need to email Sarah about budget. Also prep slides for demo.",
			response:  validJSON,
			wantCount: 2,
		},
		{
			name:      "empty array response",
			input:     "Nothing actionable here.",
			response:  "[]",
			wantCount: 0,
		},
		{
			name:      "empty input",
			input:     "",
			wantErr:   true,
			errTarget: ErrEmptyInput,
		},
		{
			name:      "whitespace-only input",
			input:     "   \n\t  ",
			wantErr:   true,
			errTarget: ErrEmptyInput,
		},
		{
			name:      "response with code fences",
			input:     "Email Sarah about budget.",
			response:  "```json\n" + validJSON + "\n```",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			backend := &mockBackend{
				name:      "test",
				responses: []string{tt.response},
			}
			e := NewTaskExtractor(backend)

			tasks, err := e.ExtractFromText(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errTarget != nil && !errors.Is(err, tt.errTarget) {
					t.Errorf("expected error %v, got %v", tt.errTarget, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tasks) != tt.wantCount {
				t.Errorf("expected %d tasks, got %d", tt.wantCount, len(tasks))
			}
		})
	}
}

func TestExtractFromTextInputTooLarge(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{name: "test", responses: []string{"[]"}}
	e := NewTaskExtractor(backend, WithMaxInputSize(100))

	largeInput := strings.Repeat("a", 101)
	_, err := e.ExtractFromText(context.Background(), largeInput)

	if err == nil {
		t.Fatal("expected error for oversized input")
	}
	if !errors.Is(err, ErrInputTooLarge) {
		t.Errorf("expected ErrInputTooLarge, got %v", err)
	}
}

func TestExtractFromTextRetryOnMalformedJSON(t *testing.T) {
	t.Parallel()

	validJSON := `[{"text": "Do the thing", "effort": 2, "tags": ["work"]}]`

	backend := &mockBackend{
		name:      "test",
		responses: []string{"not valid json", validJSON},
	}
	e := NewTaskExtractor(backend)

	tasks, err := e.ExtractFromText(context.Background(), "Some meeting notes here.")
	if err != nil {
		t.Fatalf("expected retry to succeed, got: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task after retry, got %d", len(tasks))
	}
	if backend.callCount != 2 {
		t.Errorf("expected 2 backend calls (initial + retry), got %d", backend.callCount)
	}
}

func TestExtractFromTextRetryAlsoFails(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		name:      "test",
		responses: []string{"bad json", "still bad json"},
	}
	e := NewTaskExtractor(backend)

	_, err := e.ExtractFromText(context.Background(), "Some text.")
	if err == nil {
		t.Fatal("expected error when retry also fails")
	}
	if !errors.Is(err, ErrParseFailed) {
		t.Errorf("expected ErrParseFailed, got %v", err)
	}
}

func TestExtractFromTextBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		name: "test",
		err:  errors.New("connection refused"),
	}
	e := NewTaskExtractor(backend)

	_, err := e.ExtractFromText(context.Background(), "Some text.")
	if err == nil {
		t.Fatal("expected error from backend failure")
	}
	if !strings.Contains(err.Error(), "extract tasks") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestExtractedTaskFields(t *testing.T) {
	t.Parallel()

	response := `[{
		"text": "Email Sarah about budget",
		"effort": 2,
		"tags": ["communication", "finance"],
		"confidence": 0.95
	}]`

	backend := &mockBackend{name: "test", responses: []string{response}}
	e := NewTaskExtractor(backend)

	tasks, err := e.ExtractFromText(context.Background(), "Need to email Sarah.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Text != "Email Sarah about budget" {
		t.Errorf("text = %q, want %q", task.Text, "Email Sarah about budget")
	}
	if task.Effort != 2 {
		t.Errorf("effort = %d, want 2", task.Effort)
	}
	if len(task.Tags) != 2 || task.Tags[0] != "communication" || task.Tags[1] != "finance" {
		t.Errorf("tags = %v, want [communication finance]", task.Tags)
	}
	if task.Confidence != 0.95 {
		t.Errorf("confidence = %f, want 0.95", task.Confidence)
	}
	if task.Source != "text" {
		t.Errorf("source = %q, want %q", task.Source, "text")
	}
}

func TestExtractFromFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(filePath, []byte("Need to email Sarah about budget."), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	response := `[{"text": "Email Sarah about budget", "effort": 1, "tags": ["communication"]}]`
	backend := &mockBackend{name: "test", responses: []string{response}}
	e := NewTaskExtractor(backend)

	tasks, err := e.ExtractFromFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if !strings.HasPrefix(tasks[0].Source, "file:") {
		t.Errorf("source = %q, want prefix 'file:'", tasks[0].Source)
	}
}

func TestExtractFromFileNotFound(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{name: "test", responses: []string{"[]"}}
	e := NewTaskExtractor(backend)

	_, err := e.ExtractFromFile(context.Background(), "/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestExtractFromClipboard(t *testing.T) {
	t.Parallel()

	response := `[{"text": "Review PR", "effort": 2, "tags": ["dev"]}]`
	backend := &mockBackend{name: "test", responses: []string{response}}
	runner := &mockRunner{stdout: "Need to review PR before end of day."}
	e := NewTaskExtractor(backend, WithRunner(runner))

	tasks, err := e.ExtractFromClipboard(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Source != "clipboard" {
		t.Errorf("source = %q, want %q", tasks[0].Source, "clipboard")
	}
}

func TestExtractFromClipboardNoRunner(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{name: "test", responses: []string{"[]"}}
	e := NewTaskExtractor(backend) // no runner configured

	_, err := e.ExtractFromClipboard(context.Background())
	if err == nil {
		t.Fatal("expected error when no runner configured")
	}
}

func TestExtractFromClipboardRunnerError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{name: "test", responses: []string{"[]"}}
	runner := &mockRunner{err: errors.New("pbpaste not found"), stderr: "command not found"}
	e := NewTaskExtractor(backend, WithRunner(runner))

	_, err := e.ExtractFromClipboard(context.Background())
	if err == nil {
		t.Fatal("expected error from runner failure")
	}
}

func TestDuplicateChecking(t *testing.T) {
	t.Parallel()

	response := `[
		{"text": "Email Sarah about budget", "effort": 1, "tags": ["communication"]},
		{"text": "Something totally new", "effort": 2, "tags": ["misc"]}
	]`

	checker := func(text string) (bool, float64) {
		if strings.Contains(strings.ToLower(text), "email sarah") {
			return true, 0.92
		}
		return false, 0.0
	}

	backend := &mockBackend{name: "test", responses: []string{response}}
	e := NewTaskExtractor(backend, WithDuplicateChecker(checker))

	tasks, err := e.ExtractFromText(context.Background(), "Email Sarah about budget. Also do something new.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if !tasks[0].Duplicate {
		t.Error("expected first task to be flagged as duplicate")
	}
	if tasks[1].Duplicate {
		t.Error("expected second task to NOT be flagged as duplicate")
	}
}

func TestParseExtractionResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		response  string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "valid JSON array",
			response:  `[{"text": "Do thing", "effort": 1, "tags": ["work"]}]`,
			wantCount: 1,
		},
		{
			name:      "empty array",
			response:  "[]",
			wantCount: 0,
		},
		{
			name:     "invalid JSON",
			response: "not json at all",
			wantErr:  true,
		},
		{
			name:     "JSON object instead of array",
			response: `{"text": "single task"}`,
			wantErr:  true,
		},
		{
			name:      "JSON with code fences",
			response:  "```json\n[{\"text\": \"Do thing\", \"effort\": 1, \"tags\": [\"work\"]}]\n```",
			wantCount: 1,
		},
		{
			name:      "whitespace around JSON",
			response:  "  \n []\n  ",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tasks, err := parseExtractionResponse(tt.response)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tasks) != tt.wantCount {
				t.Errorf("expected %d tasks, got %d", tt.wantCount, len(tasks))
			}
		})
	}
}

func TestBuildExtractionPrompt(t *testing.T) {
	t.Parallel()

	prompt := buildExtractionPrompt("test input text")
	if !strings.Contains(prompt, "test input text") {
		t.Error("prompt should contain the input text")
	}
	if !strings.Contains(prompt, "JSON array") {
		t.Error("prompt should request JSON output")
	}
	if !strings.Contains(prompt, "imperative") {
		t.Error("prompt should request imperative form")
	}
}

func TestBuildRetryPrompt(t *testing.T) {
	t.Parallel()

	prompt := buildRetryPrompt("test input text")
	if !strings.Contains(prompt, "test input text") {
		t.Error("retry prompt should contain the input text")
	}
	if !strings.Contains(prompt, "could not be parsed") {
		t.Error("retry prompt should mention parse failure")
	}
}

func TestStripCodeFences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no fences",
			input: `[{"text": "hello"}]`,
			want:  `[{"text": "hello"}]`,
		},
		{
			name:  "json fences",
			input: "```json\n[{\"text\": \"hello\"}]\n```",
			want:  `[{"text": "hello"}]`,
		},
		{
			name:  "plain fences",
			input: "```\n[{\"text\": \"hello\"}]\n```",
			want:  `[{"text": "hello"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stripCodeFences(tt.input)
			if got != tt.want {
				t.Errorf("stripCodeFences(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
