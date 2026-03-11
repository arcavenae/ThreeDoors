package cli

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestIsValidMood(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		mood  string
		valid bool
	}{
		{"focused lowercase", "focused", true},
		{"focused mixed case", "Focused", true},
		{"energized", "energized", true},
		{"tired", "tired", true},
		{"stressed", "stressed", true},
		{"neutral", "neutral", true},
		{"calm", "calm", true},
		{"distracted", "distracted", true},
		{"invalid mood", "happy", false},
		{"empty string", "", false},
		{"custom is not valid mood", "custom", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isValidMood(tt.mood)
			if got != tt.valid {
				t.Errorf("isValidMood(%q) = %v, want %v", tt.mood, got, tt.valid)
			}
		})
	}
}

func TestNewMoodCmd_Structure(t *testing.T) {
	t.Parallel()

	cmd := newMoodCmd()
	if cmd.Use != "mood" {
		t.Errorf("Use = %q, want %q", cmd.Use, "mood")
	}

	subCmds := cmd.Commands()
	names := make(map[string]bool)
	for _, sub := range subCmds {
		names[sub.Name()] = true
	}

	for _, want := range []string{"set", "history"} {
		if !names[want] {
			t.Errorf("missing %q subcommand", want)
		}
	}
}

func TestMoodSetCmd_Args(t *testing.T) {
	t.Parallel()

	cmd := newMoodSetCmd()
	if cmd.Use != "set <mood> [custom-text]" {
		t.Errorf("Use = %q, want expected", cmd.Use)
	}
}

func TestMoodHistoryCmd_EmptyJSON(t *testing.T) {
	t.Parallel()

	// Verify the JSON envelope structure by testing the formatter directly
	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	var entries []interface{}
	err := formatter.WriteJSON("mood history", entries, map[string]int{"total": 0})
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
		return
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Command != "mood history" {
		t.Errorf("command = %q, want %q", env.Command, "mood history")
	}
	if env.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", env.SchemaVersion)
	}
}

func TestValidMoods_ContainsExpected(t *testing.T) {
	t.Parallel()

	expected := []string{"focused", "energized", "tired", "stressed", "neutral", "calm", "distracted"}
	for _, mood := range expected {
		if !isValidMood(mood) {
			t.Errorf("expected %q to be a valid mood", mood)
		}
	}
}
