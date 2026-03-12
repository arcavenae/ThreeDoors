package llm

import (
	"testing"
)

func TestPlainTextParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []byte
		want    string
		wantErr bool
	}{
		{"trims whitespace", []byte("  hello world  \n"), "hello world", false},
		{"returns content as-is when clean", []byte("hello"), "hello", false},
		{"empty input returns error", []byte(""), "", true},
		{"whitespace-only returns error", []byte("   \n\t  "), "", true},
		{"nil input returns error", nil, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := PlainTextParser{}
			got, err := p.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parse() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArgTemplate(t *testing.T) {
	t.Parallel()

	at := ArgTemplate{
		Flag:    "--system-prompt",
		Value:   "You are helpful.",
		Enabled: true,
	}
	if at.Flag != "--system-prompt" {
		t.Errorf("Flag = %q, want %q", at.Flag, "--system-prompt")
	}
	if at.Value != "You are helpful." {
		t.Errorf("Value = %q, want %q", at.Value, "You are helpful.")
	}
	if !at.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestInputMethodConstants(t *testing.T) {
	t.Parallel()

	if InputStdin != 0 {
		t.Errorf("InputStdin = %d, want 0", InputStdin)
	}
	if InputArg != 1 {
		t.Errorf("InputArg = %d, want 1", InputArg)
	}
}
