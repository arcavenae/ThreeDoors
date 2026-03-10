package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestOutputFormatter_WriteJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		command  string
		data     interface{}
		metadata interface{}
		wantCmd  string
		wantVer  int
	}{
		{
			name:     "single object data",
			command:  "task.show",
			data:     map[string]string{"id": "abc123", "text": "Buy milk"},
			metadata: nil,
			wantCmd:  "task.show",
			wantVer:  1,
		},
		{
			name:    "list data with metadata",
			command: "task.list",
			data:    []string{"a", "b", "c"},
			metadata: map[string]interface{}{
				"total":    3,
				"filtered": 3,
			},
			wantCmd: "task.list",
			wantVer: 1,
		},
		{
			name:     "nil data",
			command:  "task.complete",
			data:     nil,
			metadata: nil,
			wantCmd:  "task.complete",
			wantVer:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			f := NewOutputFormatter(&buf, true)

			if err := f.WriteJSON(tt.command, tt.data, tt.metadata); err != nil {
				t.Fatalf("WriteJSON() error: %v", err)
			}

			var env JSONEnvelope
			if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}

			if env.SchemaVersion != tt.wantVer {
				t.Errorf("schema_version = %d, want %d", env.SchemaVersion, tt.wantVer)
			}
			if env.Command != tt.wantCmd {
				t.Errorf("command = %q, want %q", env.Command, tt.wantCmd)
			}
			if env.Error != nil {
				t.Errorf("error should be nil, got %+v", env.Error)
			}
		})
	}
}

func TestOutputFormatter_WriteJSONError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		command    string
		code       int
		message    string
		detail     string
		wantCode   int
		wantMsg    string
		wantDetail string
	}{
		{
			name:     "not found error",
			command:  "task.show",
			code:     ExitNotFound,
			message:  "task not found",
			detail:   "no task with prefix \"xyz\"",
			wantCode: ExitNotFound,
			wantMsg:  "task not found",
		},
		{
			name:     "validation error without detail",
			command:  "task.status",
			code:     ExitValidation,
			message:  "invalid status transition",
			detail:   "",
			wantCode: ExitValidation,
			wantMsg:  "invalid status transition",
		},
		{
			name:     "ambiguous input",
			command:  "task.show",
			code:     ExitAmbiguousInput,
			message:  "ambiguous prefix",
			detail:   "prefix \"a\" matches 5 tasks",
			wantCode: ExitAmbiguousInput,
			wantMsg:  "ambiguous prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			f := NewOutputFormatter(&buf, true)

			if err := f.WriteJSONError(tt.command, tt.code, tt.message, tt.detail); err != nil {
				t.Fatalf("WriteJSONError() error: %v", err)
			}

			var env JSONEnvelope
			if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}

			if env.SchemaVersion != 1 {
				t.Errorf("schema_version = %d, want 1", env.SchemaVersion)
			}
			if env.Command != tt.command {
				t.Errorf("command = %q, want %q", env.Command, tt.command)
			}
			if env.Data != nil {
				t.Errorf("data should be nil for error response, got %v", env.Data)
			}
			if env.Error == nil {
				t.Fatal("error should not be nil")
				return
			}
			if env.Error.Code != tt.wantCode {
				t.Errorf("error.code = %d, want %d", env.Error.Code, tt.wantCode)
			}
			if env.Error.Message != tt.wantMsg {
				t.Errorf("error.message = %q, want %q", env.Error.Message, tt.wantMsg)
			}
			if tt.detail != "" && env.Error.Detail != tt.detail {
				t.Errorf("error.detail = %q, want %q", env.Error.Detail, tt.detail)
			}
		})
	}
}

func TestOutputFormatter_TableWriter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	f := NewOutputFormatter(&buf, false)

	tw := f.TableWriter()
	if _, err := fmt.Fprintln(tw, "ID\tSTATUS\tTEXT"); err != nil {
		t.Fatalf("Fprintln() error: %v", err)
	}
	if _, err := fmt.Fprintln(tw, "abc123\ttodo\tBuy milk"); err != nil {
		t.Fatalf("Fprintln() error: %v", err)
	}
	if _, err := fmt.Fprintln(tw, "def456\tactive\tWrite tests"); err != nil {
		t.Fatalf("Fprintln() error: %v", err)
	}
	if err := tw.Flush(); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatal("table output should not be empty")
	}
	// tabwriter should align columns with spaces
	if !bytes.Contains(buf.Bytes(), []byte("abc123")) {
		t.Error("output should contain task ID")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Buy milk")) {
		t.Error("output should contain task text")
	}
}

func TestOutputFormatter_Writef(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	f := NewOutputFormatter(&buf, false)

	if err := f.Writef("Found %d tasks\n", 42); err != nil {
		t.Fatalf("Writef() error: %v", err)
	}

	want := "Found 42 tasks\n"
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

func TestOutputFormatter_IsJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		jsonMode bool
		want     bool
	}{
		{"json mode", true, true},
		{"human mode", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			f := NewOutputFormatter(&buf, tt.jsonMode)
			if got := f.IsJSON(); got != tt.want {
				t.Errorf("IsJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONEnvelope_Structure(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	f := NewOutputFormatter(&buf, true)

	data := map[string]string{"id": "test-123"}
	metadata := map[string]int{"total": 1}
	if err := f.WriteJSON("task.show", data, metadata); err != nil {
		t.Fatalf("WriteJSON() error: %v", err)
	}

	// Parse as raw JSON to verify exact structure
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}

	requiredKeys := []string{"schema_version", "command", "data", "metadata"}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing required key %q in JSON envelope", key)
		}
	}

	// error should be omitted (omitempty) when nil
	if _, ok := raw["error"]; ok {
		t.Error("error key should be omitted in success response")
	}
}

func TestJSONErrorEnvelope_Structure(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	f := NewOutputFormatter(&buf, true)

	if err := f.WriteJSONError("task.show", ExitNotFound, "not found", "detail"); err != nil {
		t.Fatalf("WriteJSONError() error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}

	requiredKeys := []string{"schema_version", "command", "error"}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing required key %q in JSON error envelope", key)
		}
	}

	// data should be omitted when nil
	if _, ok := raw["data"]; ok {
		t.Error("data key should be omitted in error response")
	}
}
