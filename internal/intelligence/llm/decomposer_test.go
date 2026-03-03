package llm

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// mockBackend implements LLMBackend for testing.
type mockBackend struct {
	name      string
	response  string
	err       error
	available bool
}

func (m *mockBackend) Name() string { return m.name }

func (m *mockBackend) Complete(_ context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", ErrEmptyPrompt
	}
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *mockBackend) Available(_ context.Context) bool { return m.available }

func TestTaskDecomposer_Decompose(t *testing.T) {
	t.Parallel()

	validJSON := `[{"story_id":"14.1","title":"Test Story","user_story":"As a dev, I want tests, So that I ship quality","acceptance_criteria":["AC1"],"tasks":["Task1"],"dev_notes":"notes"}]`

	tests := []struct {
		name        string
		task        string
		backend     *mockBackend
		wantErr     bool
		wantStories int
	}{
		{
			name: "successful decomposition",
			task: "Build authentication system",
			backend: &mockBackend{
				name:     "test",
				response: validJSON,
			},
			wantStories: 1,
		},
		{
			name: "handles code fences",
			task: "Build auth",
			backend: &mockBackend{
				name:     "test",
				response: "```json\n" + validJSON + "\n```",
			},
			wantStories: 1,
		},
		{
			name:    "empty task description",
			task:    "",
			backend: &mockBackend{name: "test"},
			wantErr: true,
		},
		{
			name: "backend error",
			task: "Build auth",
			backend: &mockBackend{
				name: "test",
				err:  fmt.Errorf("connection refused"),
			},
			wantErr: true,
		},
		{
			name: "invalid json response",
			task: "Build auth",
			backend: &mockBackend{
				name:     "test",
				response: "I cannot help with that",
			},
			wantErr: true,
		},
		{
			name: "missing required fields",
			task: "Build auth",
			backend: &mockBackend{
				name:     "test",
				response: `[{"story_id":"","title":"","user_story":"","acceptance_criteria":[],"tasks":[]}]`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := NewTaskDecomposer(tt.backend)
			result, err := d.Decompose(context.Background(), tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decompose() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(result.Stories) != tt.wantStories {
				t.Errorf("Decompose() got %d stories, want %d", len(result.Stories), tt.wantStories)
			}
			if !tt.wantErr && result.Backend != tt.backend.name {
				t.Errorf("Decompose() backend = %q, want %q", result.Backend, tt.backend.name)
			}
		})
	}
}

func TestExtractJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain json array",
			input: `[{"key":"value"}]`,
			want:  `[{"key":"value"}]`,
		},
		{
			name:  "json in code fence",
			input: "```json\n[{\"key\":\"value\"}]\n```",
			want:  `[{"key":"value"}]`,
		},
		{
			name:  "json in generic fence",
			input: "```\n[{\"key\":\"value\"}]\n```",
			want:  `[{"key":"value"}]`,
		},
		{
			name:  "json with surrounding text",
			input: "Here is the output:\n[{\"key\":\"value\"}]\nDone.",
			want:  `[{"key":"value"}]`,
		},
		{
			name:  "json object",
			input: `{"stories":[]}`,
			want:  `{"stories":[]}`,
		},
		{
			name:  "no json",
			input: "no json here",
			want:  "no json here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractJSON(tt.input)
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("extractJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}
