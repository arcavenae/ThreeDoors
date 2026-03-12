package services

import (
	"context"
	"errors"
	"testing"
)

func TestBreakdownServiceEmptyDescription(t *testing.T) {
	t.Parallel()
	svc := NewBreakdownService(&mockBackend{name: "test"})

	_, err := svc.Breakdown(context.Background(), "task-1", "")
	if err == nil {
		t.Fatal("expected error for empty description")
	}
}

func TestBreakdownServiceBackendError(t *testing.T) {
	t.Parallel()
	svc := NewBreakdownService(&mockBackend{
		name: "test",
		err:  errors.New("network timeout"),
	})

	_, err := svc.Breakdown(context.Background(), "task-1", "Build a thing")
	if err == nil {
		t.Fatal("expected error when backend fails")
	}
	if !containsStr(err.Error(), "network timeout") {
		t.Errorf("error should wrap backend error, got: %v", err)
	}
}

func TestBreakdownServiceSuccess(t *testing.T) {
	t.Parallel()
	svc := NewBreakdownService(&mockBackend{
		name: "test-llm",
		responses: []string{`[
			{"text": "Set up project structure", "effort_estimate": "small"},
			{"text": "Implement core logic", "effort_estimate": "medium"},
			{"text": "Write tests", "effort_estimate": "small"}
		]`},
	})

	result, err := svc.Breakdown(context.Background(), "task-42", "Build the feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ParentTaskID != "task-42" {
		t.Errorf("ParentTaskID = %q, want %q", result.ParentTaskID, "task-42")
	}
	if result.ParentText != "Build the feature" {
		t.Errorf("ParentText = %q, want %q", result.ParentText, "Build the feature")
	}
	if result.Backend != "test-llm" {
		t.Errorf("Backend = %q, want %q", result.Backend, "test-llm")
	}
	if len(result.Subtasks) != 3 {
		t.Fatalf("got %d subtasks, want 3", len(result.Subtasks))
	}
	if result.Subtasks[0].Text != "Set up project structure" {
		t.Errorf("Subtasks[0].Text = %q", result.Subtasks[0].Text)
	}
	if result.Subtasks[1].EffortEstimate != "medium" {
		t.Errorf("Subtasks[1].EffortEstimate = %q, want %q", result.Subtasks[1].EffortEstimate, "medium")
	}
}

func TestBreakdownServiceMarkdownWrapped(t *testing.T) {
	t.Parallel()
	svc := NewBreakdownService(&mockBackend{
		name: "test",
		responses: []string{"Here are the subtasks:\n```json\n" +
			`[{"text": "Step 1", "effort_estimate": "small"}]` +
			"\n```\n"},
	})

	result, err := svc.Breakdown(context.Background(), "t1", "Do something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Subtasks) != 1 {
		t.Fatalf("got %d subtasks, want 1", len(result.Subtasks))
	}
	if result.Subtasks[0].Text != "Step 1" {
		t.Errorf("Subtasks[0].Text = %q, want %q", result.Subtasks[0].Text, "Step 1")
	}
}

func TestBreakdownServiceNoSubtasks(t *testing.T) {
	t.Parallel()
	svc := NewBreakdownService(&mockBackend{
		name:      "test",
		responses: []string{"[]"},
	})

	_, err := svc.Breakdown(context.Background(), "t1", "Do something")
	if err == nil {
		t.Fatal("expected error for empty subtask list")
	}
}

func TestBreakdownServiceInvalidJSON(t *testing.T) {
	t.Parallel()
	svc := NewBreakdownService(&mockBackend{
		name:      "test",
		responses: []string{"not json at all"},
	})

	_, err := svc.Breakdown(context.Background(), "t1", "Do something")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestBreakdownServiceEmptySubtaskText(t *testing.T) {
	t.Parallel()
	svc := NewBreakdownService(&mockBackend{
		name:      "test",
		responses: []string{`[{"text": "", "effort_estimate": "small"}]`},
	})

	_, err := svc.Breakdown(context.Background(), "t1", "Do something")
	if err == nil {
		t.Fatal("expected error for empty subtask text")
	}
}

func TestBreakdownServiceContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := NewBreakdownService(&mockBackend{
		name: "test",
		err:  ctx.Err(),
	})

	_, err := svc.Breakdown(ctx, "t1", "Do something")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestParseSubtasks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "plain JSON array",
			input: `[{"text": "A", "effort_estimate": "small"}, {"text": "B", "effort_estimate": "large"}]`,
			want:  2,
		},
		{
			name:  "code fence wrapped",
			input: "```json\n[{\"text\": \"A\", \"effort_estimate\": \"small\"}]\n```",
			want:  1,
		},
		{
			name:  "generic code fence",
			input: "```\n[{\"text\": \"A\", \"effort_estimate\": \"small\"}]\n```",
			want:  1,
		},
		{
			name:    "invalid JSON",
			input:   "not json",
			wantErr: true,
		},
		{
			name:    "empty text field",
			input:   `[{"text": " ", "effort_estimate": "small"}]`,
			wantErr: true,
		},
		{
			name:  "missing effort estimate is ok",
			input: `[{"text": "Do thing"}]`,
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseSubtasks(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseSubtasks() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("parseSubtasks() returned %d subtasks, want %d", len(got), tt.want)
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
			name:  "json code fence",
			input: "text\n```json\n[1,2,3]\n```\nmore",
			want:  "[1,2,3]",
		},
		{
			name:  "generic code fence",
			input: "text\n```\n[1,2,3]\n```\nmore",
			want:  "[1,2,3]",
		},
		{
			name:  "bare array",
			input: "here: [1,2,3] done",
			want:  "[1,2,3]",
		},
		{
			name:  "plain text",
			input: "no json here",
			want:  "no json here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractJSON(tt.input)
			if got != tt.want {
				t.Errorf("extractJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildBreakdownPrompt(t *testing.T) {
	t.Parallel()
	prompt := buildBreakdownPrompt("Build a website")
	if !containsStr(prompt, "Build a website") {
		t.Error("prompt should contain the task description")
	}
	if !containsStr(prompt, "JSON") {
		t.Error("prompt should mention JSON output format")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstring(s, sub))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
