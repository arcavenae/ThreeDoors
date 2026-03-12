package services

import (
	"context"
	"fmt"
	"testing"
)

// enrichMockBackend implements llm.LLMBackend for testing.
type enrichMockBackend struct {
	name      string
	response  string
	err       error
	available bool
}

func (m *enrichMockBackend) Name() string { return m.name }
func (m *enrichMockBackend) Complete(_ context.Context, _ string) (string, error) {
	return m.response, m.err
}
func (m *enrichMockBackend) Available(_ context.Context) bool { return m.available }

func TestNewTaskEnricher(t *testing.T) {
	t.Parallel()
	backend := &enrichMockBackend{name: "test"}
	enricher := NewTaskEnricher(backend)
	if enricher == nil {
		t.Fatal("NewTaskEnricher returned nil")
	}
	if enricher.backend != backend {
		t.Error("backend not set correctly")
	}
}

func TestEnrich(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		taskText   string
		response   string
		backendErr error
		wantErr    bool
		wantTags   int
		wantEffort int
	}{
		{
			name:       "successful enrichment",
			taskText:   "taxes",
			response:   `{"enriched_text": "Gather 2025 tax documents (W-2, 1099s, receipts) and schedule appointment with accountant", "tags": ["finance", "personal", "deadline"], "effort": 3, "context": "Tax filing deadline is April 15. Gather all income and deduction documents."}`,
			wantTags:   3,
			wantEffort: 3,
		},
		{
			name:       "enrichment with code fence",
			taskText:   "fix login",
			response:   "```json\n{\"enriched_text\": \"Fix login session timeout bug\", \"tags\": [\"bug\", \"auth\"], \"effort\": 2, \"context\": \"Users report being logged out unexpectedly.\"}\n```",
			wantTags:   2,
			wantEffort: 2,
		},
		{
			name:       "enrichment with surrounding text",
			taskText:   "clean garage",
			response:   "Here's the enrichment:\n{\"enriched_text\": \"Organize and clean garage — sort tools, donate unused items\", \"tags\": [\"home\", \"organization\"], \"effort\": 4, \"context\": \"Spring cleaning opportunity. Consider renting a dumpster for bulk items.\"}\nHope that helps!",
			wantTags:   2,
			wantEffort: 4,
		},
		{
			name:     "empty task text",
			taskText: "",
			wantErr:  true,
		},
		{
			name:     "whitespace-only task text",
			taskText: "   ",
			wantErr:  true,
		},
		{
			name:       "backend error",
			taskText:   "something",
			backendErr: fmt.Errorf("connection refused"),
			wantErr:    true,
		},
		{
			name:     "invalid JSON response",
			taskText: "something",
			response: "I don't understand the task",
			wantErr:  true,
		},
		{
			name:     "empty enriched text in response",
			taskText: "something",
			response: `{"enriched_text": "", "tags": [], "effort": 1, "context": "test"}`,
			wantErr:  true,
		},
		{
			name:     "effort out of range high",
			taskText: "something",
			response: `{"enriched_text": "do something", "tags": [], "effort": 10, "context": "test"}`,
			wantErr:  true,
		},
		{
			name:     "effort out of range negative",
			taskText: "something",
			response: `{"enriched_text": "do something", "tags": [], "effort": -1, "context": "test"}`,
			wantErr:  true,
		},
		{
			name:       "minimal valid response",
			taskText:   "buy milk",
			response:   `{"enriched_text": "Buy milk from the grocery store", "tags": ["errands"], "effort": 1, "context": "Running low on dairy."}`,
			wantTags:   1,
			wantEffort: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			backend := &enrichMockBackend{
				name:      "test-backend",
				response:  tt.response,
				err:       tt.backendErr,
				available: true,
			}
			enricher := NewTaskEnricher(backend)
			result, err := enricher.Enrich(context.Background(), tt.taskText)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.OriginalText != tt.taskText {
				t.Errorf("OriginalText = %q, want %q", result.OriginalText, tt.taskText)
			}
			if result.EnrichedText == "" {
				t.Error("EnrichedText should not be empty")
			}
			if len(result.Tags) != tt.wantTags {
				t.Errorf("got %d tags, want %d", len(result.Tags), tt.wantTags)
			}
			if result.Effort != tt.wantEffort {
				t.Errorf("Effort = %d, want %d", result.Effort, tt.wantEffort)
			}
		})
	}
}

func TestEnrichedTaskValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		task    EnrichedTask
		wantErr bool
	}{
		{
			name:    "valid",
			task:    EnrichedTask{EnrichedText: "do something", Effort: 3},
			wantErr: false,
		},
		{
			name:    "empty enriched text",
			task:    EnrichedTask{EnrichedText: "", Effort: 1},
			wantErr: true,
		},
		{
			name:    "effort too high",
			task:    EnrichedTask{EnrichedText: "test", Effort: 6},
			wantErr: true,
		},
		{
			name:    "effort negative",
			task:    EnrichedTask{EnrichedText: "test", Effort: -1},
			wantErr: true,
		},
		{
			name:    "effort zero is valid",
			task:    EnrichedTask{EnrichedText: "test", Effort: 0},
			wantErr: false,
		},
		{
			name:    "effort five is valid",
			task:    EnrichedTask{EnrichedText: "test", Effort: 5},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildEnrichmentPrompt(t *testing.T) {
	t.Parallel()
	prompt := buildEnrichmentPrompt("buy groceries")
	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}
	if !containsAll(prompt, "buy groceries", "enriched_text", "tags", "effort", "context", "JSON") {
		t.Error("prompt missing expected content")
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
