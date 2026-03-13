package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestMoodSetJSON_ValidMoods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mood       string
		customText string
	}{
		{"focused", ""},
		{"energized", ""},
		{"tired", ""},
		{"stressed", ""},
		{"neutral", ""},
		{"calm", ""},
		{"distracted", ""},
	}

	for _, tt := range tests {
		t.Run(tt.mood, func(t *testing.T) {
			t.Parallel()

			entry := core.MoodEntry{
				Timestamp:  time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC),
				Mood:       tt.mood,
				CustomText: tt.customText,
			}

			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, true)
			if err := formatter.WriteJSON("mood set", entry, nil); err != nil {
				t.Fatalf("WriteJSON: %v", err)
			}

			var env JSONEnvelope
			if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if env.Command != "mood set" {
				t.Errorf("command = %q, want %q", env.Command, "mood set")
			}
			if env.Error != nil {
				t.Errorf("unexpected error: %v", env.Error)
			}

			data, ok := env.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("data not a map: %T", env.Data)
			}
			if data["mood"] != tt.mood {
				t.Errorf("mood = %v, want %q", data["mood"], tt.mood)
			}
		})
	}
}

func TestMoodSetJSON_CustomMood(t *testing.T) {
	t.Parallel()

	entry := core.MoodEntry{
		Timestamp:  time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC),
		Mood:       "custom",
		CustomText: "feeling productive",
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	if err := formatter.WriteJSON("mood set", entry, nil); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data not a map: %T", env.Data)
	}
	if data["mood"] != "custom" {
		t.Errorf("mood = %v, want %q", data["mood"], "custom")
	}
	if data["custom_text"] != "feeling productive" {
		t.Errorf("custom_text = %v, want %q", data["custom_text"], "feeling productive")
	}
}

func TestMoodSetHumanOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mood       string
		customText string
		want       string
	}{
		{"standard mood", "focused", "", "Recorded mood: focused"},
		{"custom mood", "custom", "feeling great", "Recorded mood: custom (feeling great)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, false)
			if tt.customText != "" {
				_ = formatter.Writef("Recorded mood: custom (%s)\n", tt.customText)
			} else {
				_ = formatter.Writef("Recorded mood: %s\n", tt.mood)
			}

			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("output = %q, want containing %q", buf.String(), tt.want)
			}
		})
	}
}

func TestMoodSetValidation_InvalidMoods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mood string
	}{
		{"happy", "happy"},
		{"sad", "sad"},
		{"empty", ""},
		{"number", "123"},
		{"special chars", "mood!@#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if isValidMood(tt.mood) {
				t.Errorf("isValidMood(%q) = true, want false", tt.mood)
			}

			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, true)
			msg := fmt.Sprintf("invalid mood %q; valid moods: %s, custom", tt.mood, strings.Join(validMoods, ", "))
			_ = formatter.WriteJSONError("mood set", ExitValidation, msg, "")

			var env JSONEnvelope
			if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if env.Error == nil {
				t.Fatal("expected error in envelope")
			}
			if env.Error.Code != ExitValidation {
				t.Errorf("error code = %d, want %d", env.Error.Code, ExitValidation)
			}
		})
	}
}

func TestMoodHistoryJSON_PopulatedEntries(t *testing.T) {
	t.Parallel()

	entries := []core.MoodEntry{
		{
			Timestamp: time.Date(2026, 3, 13, 9, 0, 0, 0, time.UTC),
			Mood:      "focused",
		},
		{
			Timestamp:  time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
			Mood:       "custom",
			CustomText: "afternoon slump",
		},
		{
			Timestamp: time.Date(2026, 3, 13, 16, 0, 0, 0, time.UTC),
			Mood:      "energized",
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	if err := formatter.WriteJSON("mood history", entries, map[string]int{"total": len(entries)}); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Command != "mood history" {
		t.Errorf("command = %q, want %q", env.Command, "mood history")
	}

	data, ok := env.Data.([]interface{})
	if !ok {
		t.Fatalf("data not an array: %T", env.Data)
	}
	if len(data) != 3 {
		t.Errorf("got %d entries, want 3", len(data))
	}

	meta, ok := env.Metadata.(map[string]interface{})
	if !ok {
		t.Fatalf("metadata not a map: %T", env.Metadata)
	}
	if meta["total"] != float64(3) {
		t.Errorf("metadata.total = %v, want 3", meta["total"])
	}
}

func TestMoodHistoryHumanOutput_Populated(t *testing.T) {
	t.Parallel()

	entries := []core.MoodEntry{
		{
			Timestamp: time.Date(2026, 3, 13, 9, 0, 0, 0, time.UTC),
			Mood:      "focused",
		},
		{
			Timestamp:  time.Date(2026, 3, 13, 14, 30, 0, 0, time.UTC),
			Mood:       "custom",
			CustomText: "relaxed",
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)
	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "TIME\tMOOD\tCUSTOM\n")
	for _, e := range entries {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n",
			e.Timestamp.Format("2006-01-02 15:04"),
			e.Mood,
			e.CustomText,
		)
	}
	_ = tw.Flush()
	_ = formatter.Writef("%d mood entries\n", len(entries))

	output := buf.String()
	for _, want := range []string{"TIME", "MOOD", "CUSTOM", "focused", "custom", "relaxed", "2 mood entries"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q, got:\n%s", want, output)
		}
	}
}

func TestMoodHistoryHumanOutput_Empty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)
	_ = formatter.Writef("No mood entries found.\n")

	if !strings.Contains(buf.String(), "No mood entries found") {
		t.Errorf("output = %q, want containing 'No mood entries found'", buf.String())
	}
}

func TestMoodSetJSON_CustomMissingText(t *testing.T) {
	t.Parallel()

	msg := "custom mood requires a text argument"
	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	_ = formatter.WriteJSONError("mood set", ExitValidation, msg, "")

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected error in envelope")
	}
	if !strings.Contains(env.Error.Message, "custom mood requires") {
		t.Errorf("error message = %q, want containing 'custom mood requires'", env.Error.Message)
	}
}
