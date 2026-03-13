package mcpbridge

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateRecipient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid simple", "supervisor", false, ""},
		{"valid with hyphen", "merge-queue", false, ""},
		{"valid with underscore", "my_agent", false, ""},
		{"valid alphanumeric", "worker42", false, ""},
		{"empty", "", true, "is required"},
		{"shell semicolon", "agent;rm -rf /", true, "prohibited character"},
		{"shell pipe", "agent|cat", true, "prohibited character"},
		{"shell subshell", "agent$(whoami)", true, "prohibited character"},
		{"shell backtick", "agent`id`", true, "prohibited character"},
		{"shell and", "a&&b", true, "prohibited character"},
		{"shell or", "a||b", true, "prohibited character"},
		{"shell redirect out", "a>b", true, "prohibited character"},
		{"shell redirect in", "a<b", true, "prohibited character"},
		{"spaces", "my agent", true, "must contain only"},
		{"special chars", "agent@host", true, "must contain only"},
		{"dots", "agent.name", true, "must contain only"},
		{"slash", "agent/name", true, "must contain only"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateRecipient(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRecipient(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("expected ValidationError, got %T", err)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateMessageID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid uuid", "msg-abc123-def4-5678", false, ""},
		{"valid short", "msg-abc", false, ""},
		{"valid with hyphens", "msg-a1b2-c3d4-e5f6-7890", false, ""},
		{"empty", "", true, "is required"},
		{"missing prefix", "abc-123", true, "must match format"},
		{"wrong prefix", "message-abc", true, "must match format"},
		{"uppercase", "msg-ABC123", true, "must match format"},
		{"shell injection", "msg-abc;rm -rf /", true, "prohibited character"},
		{"null byte", "msg-abc\x00def", true, "prohibited character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateMessageID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMessageID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid short", "Hello, world!", false, ""},
		{"valid long", strings.Repeat("a", MaxBodyLength), false, ""},
		{"empty", "", true, "is required"},
		{"too long", strings.Repeat("a", MaxBodyLength+1), true, "exceeds maximum length"},
		{"shell semicolon", "hello; rm -rf /", true, "prohibited character"},
		{"shell pipe", "hello | cat", true, "prohibited character"},
		{"null byte", "hello\x00world", true, "prohibited character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateBody(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBody(%q...) error = %v, wantErr %v", truncate(tt.input, 20), err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid short", "implement story 53.4", false, ""},
		{"valid at limit", strings.Repeat("x", MaxTaskLength), false, ""},
		{"empty", "", true, "is required"},
		{"too long", strings.Repeat("x", MaxTaskLength+1), true, "exceeds maximum length"},
		{"shell injection", "task $(rm -rf /)", true, "prohibited character"},
		{"backtick injection", "task `id`", true, "prohibited character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTask(%q...) error = %v, wantErr %v", truncate(tt.input, 20), err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateNoShellMetachars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"clean string", "hello world", false},
		{"semicolon", "a;b", true},
		{"pipe", "a|b", true},
		{"subshell", "$(cmd)", true},
		{"backtick", "`cmd`", true},
		{"double and", "a&&b", true},
		{"double or", "a||b", true},
		{"redirect out", "a>b", true},
		{"redirect in", "a<b", true},
		{"null byte", "a\x00b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateNoShellMetachars("test_field", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNoShellMetachars(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	t.Parallel()

	ve := &ValidationError{Field: "recipient", Message: "is required"}
	expected := "validation error for recipient: is required"
	if ve.Error() != expected {
		t.Errorf("Error() = %q, want %q", ve.Error(), expected)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
