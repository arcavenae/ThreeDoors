package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPromptDoorSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		doorCount int
		want      int
		wantErr   bool
	}{
		{"valid door 1", "1\n", 3, 1, false},
		{"valid door 2", "2\n", 3, 2, false},
		{"valid door 3", "3\n", 3, 3, false},
		{"with whitespace", "  2  \n", 3, 2, false},
		{"too high", "4\n", 3, 0, true},
		{"too low", "0\n", 3, 0, true},
		{"negative", "-1\n", 3, 0, true},
		{"not a number", "abc\n", 3, 0, true},
		{"empty input", "\n", 3, 0, true},
		{"single door valid", "1\n", 1, 1, false},
		{"single door invalid", "2\n", 1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reader := strings.NewReader(tt.input)
			var writer bytes.Buffer

			got, err := promptDoorSelection(reader, &writer, tt.doorCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("promptDoorSelection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("promptDoorSelection() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPromptDoorSelection_EOF(t *testing.T) {
	t.Parallel()

	reader := strings.NewReader("")
	var writer bytes.Buffer

	_, err := promptDoorSelection(reader, &writer, 3)
	if err == nil {
		t.Fatal("expected error on EOF")
	}
}

func TestPromptDoorSelection_WritesPrompt(t *testing.T) {
	t.Parallel()

	reader := strings.NewReader("1\n")
	var writer bytes.Buffer

	_, err := promptDoorSelection(reader, &writer, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prompt := writer.String()
	if !strings.Contains(prompt, "Pick a door") {
		t.Errorf("expected prompt text, got: %q", prompt)
	}
}

func TestStdoutIsTerminal_NonTTY(t *testing.T) {
	t.Parallel()

	// In test environment, stdout is typically not a terminal
	// Save and restore the original function
	orig := stdoutIsTerminal
	t.Cleanup(func() { stdoutIsTerminal = orig })

	// Test with forced non-TTY
	stdoutIsTerminal = func() bool { return false }
	if stdoutIsTerminal() {
		t.Error("expected non-TTY when forced false")
	}

	// Test with forced TTY
	stdoutIsTerminal = func() bool { return true }
	if !stdoutIsTerminal() {
		t.Error("expected TTY when forced true")
	}
}
