package quota

import (
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		fixture          string
		wantInteractions int
		wantErr          bool
	}{
		{
			name:             "normal multi-entry file",
			fixture:          "normal.jsonl",
			wantInteractions: 2,
		},
		{
			name:             "empty file",
			fixture:          "empty.jsonl",
			wantInteractions: 0,
		},
		{
			name:             "malformed entries skipped",
			fixture:          "malformed.jsonl",
			wantInteractions: 2,
		},
		{
			name:             "multi-session file",
			fixture:          "multi_session.jsonl",
			wantInteractions: 4,
		},
		{
			name:    "nonexistent file",
			fixture: "does_not_exist.jsonl",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join("testdata", tt.fixture)
			interactions, err := ParseFile(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(interactions) != tt.wantInteractions {
				t.Errorf("got %d interactions, want %d", len(interactions), tt.wantInteractions)
			}
		})
	}
}

func TestParseFileTokenValues(t *testing.T) {
	t.Parallel()

	interactions, err := ParseFile(filepath.Join("testdata", "normal.jsonl"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(interactions) != 2 {
		t.Fatalf("got %d interactions, want 2", len(interactions))
	}

	first := interactions[0]
	if first.Tokens.InputTokens != 100 {
		t.Errorf("input_tokens = %d, want 100", first.Tokens.InputTokens)
	}
	if first.Tokens.OutputTokens != 50 {
		t.Errorf("output_tokens = %d, want 50", first.Tokens.OutputTokens)
	}
	if first.Tokens.CacheCreationInputTokens != 200 {
		t.Errorf("cache_creation = %d, want 200", first.Tokens.CacheCreationInputTokens)
	}
	if first.Tokens.CacheReadInputTokens != 300 {
		t.Errorf("cache_read = %d, want 300", first.Tokens.CacheReadInputTokens)
	}
	if got := first.Tokens.Total(); got != 650 {
		t.Errorf("total = %d, want 650", got)
	}
	if first.SessionID != "sess-001" {
		t.Errorf("session_id = %q, want %q", first.SessionID, "sess-001")
	}
	if first.Model != "claude-opus-4-6" {
		t.Errorf("model = %q, want %q", first.Model, "claude-opus-4-6")
	}
}

func TestParseFiles(t *testing.T) {
	t.Parallel()

	paths := []string{
		filepath.Join("testdata", "normal.jsonl"),
		filepath.Join("testdata", "multi_session.jsonl"),
		filepath.Join("testdata", "does_not_exist.jsonl"),
	}
	interactions := ParseFiles(paths)
	// normal has 2, multi_session has 4, nonexistent is skipped
	if len(interactions) != 6 {
		t.Errorf("got %d interactions, want 6", len(interactions))
	}
}
