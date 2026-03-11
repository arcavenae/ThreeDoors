package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestWriteDoctorJSON_WithCategories(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Configuration",
				Status: core.CheckOK,
				Checks: []core.CheckResult{
					{
						Name:    "config file",
						Status:  core.CheckOK,
						Message: "Config file found",
					},
					{
						Name:       "provider",
						Status:     core.CheckWarn,
						Message:    "Provider not configured",
						Suggestion: "Run threedoors config set provider textfile",
					},
				},
			},
			{
				Name:   "Environment",
				Status: core.CheckOK,
				Checks: []core.CheckResult{
					{
						Name:    "terminal",
						Status:  core.CheckOK,
						Message: "Terminal detected",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	if err := writeDoctorJSON(formatter, result); err != nil {
		t.Fatalf("writeDoctorJSON: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if env.Command != "doctor" {
		t.Errorf("command = %q, want %q", env.Command, "doctor")
	}

	dataBytes, err := json.Marshal(env.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	var dr doctorResultJSON
	if err := json.Unmarshal(dataBytes, &dr); err != nil {
		t.Fatalf("unmarshal doctor result: %v", err)
	}

	if len(dr.Categories) != 2 {
		t.Fatalf("got %d categories, want 2", len(dr.Categories))
	}

	if dr.Categories[0].Name != "Configuration" {
		t.Errorf("cat[0].Name = %q, want %q", dr.Categories[0].Name, "Configuration")
	}
	if len(dr.Categories[0].Checks) != 2 {
		t.Fatalf("cat[0] has %d checks, want 2", len(dr.Categories[0].Checks))
	}
	if dr.Categories[0].Checks[1].Suggestion != "Run threedoors config set provider textfile" {
		t.Errorf("check suggestion = %q", dr.Categories[0].Checks[1].Suggestion)
	}
}

func TestWriteDoctorJSON_EmptyResult(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	if err := writeDoctorJSON(formatter, result); err != nil {
		t.Fatalf("writeDoctorJSON: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if env.Command != "doctor" {
		t.Errorf("command = %q, want %q", env.Command, "doctor")
	}
}

func TestWriteDoctorHuman_NoIssues(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Configuration",
				Status: core.CheckOK,
				Checks: []core.CheckResult{
					{
						Name:    "config file",
						Status:  core.CheckOK,
						Message: "Config file exists",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := writeDoctorHuman(formatter, result, false); err != nil {
		t.Fatalf("writeDoctorHuman: %v", err)
	}

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("ThreeDoors Doctor")) {
		t.Errorf("missing header, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Configuration")) {
		t.Errorf("missing category name, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Config file exists")) {
		t.Errorf("missing check message, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("No issues found")) {
		t.Errorf("missing summary, got: %s", output)
	}
}

func TestWriteDoctorHuman_WithSuggestions(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Configuration",
				Status: core.CheckWarn,
				Checks: []core.CheckResult{
					{
						Name:       "provider",
						Status:     core.CheckWarn,
						Message:    "No provider configured",
						Suggestion: "Set a provider",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := writeDoctorHuman(formatter, result, false); err != nil {
		t.Fatalf("writeDoctorHuman: %v", err)
	}

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Set a provider")) {
		t.Errorf("missing suggestion, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("issue")) {
		t.Errorf("missing issue count, got: %s", output)
	}
}

func TestNewDoctorCmd_Structure(t *testing.T) {
	t.Parallel()

	cmd := newDoctorCmd()
	if cmd.Use != "doctor" {
		t.Errorf("Use = %q, want %q", cmd.Use, "doctor")
	}

	hasHealthAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "health" {
			hasHealthAlias = true
			break
		}
	}
	if !hasHealthAlias {
		t.Error("missing 'health' alias")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestDoctorCategoryJSON_Serialization(t *testing.T) {
	t.Parallel()

	cat := doctorCategoryJSON{
		Name:   "Test Category",
		Status: "ok",
		Checks: []doctorCheckJSON{
			{
				Name:    "check1",
				Status:  "ok",
				Message: "All good",
			},
			{
				Name:       "check2",
				Status:     "warn",
				Message:    "Minor issue",
				Suggestion: "Fix it",
			},
		},
	}

	data, err := json.Marshal(cat)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded doctorCategoryJSON
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Name != "Test Category" {
		t.Errorf("Name = %q, want %q", decoded.Name, "Test Category")
	}
	if len(decoded.Checks) != 2 {
		t.Fatalf("got %d checks, want 2", len(decoded.Checks))
	}
	if decoded.Checks[1].Suggestion != "Fix it" {
		t.Errorf("Suggestion = %q, want %q", decoded.Checks[1].Suggestion, "Fix it")
	}
}

func TestDoctorCheckJSON_OmitsEmptySuggestion(t *testing.T) {
	t.Parallel()

	check := doctorCheckJSON{
		Name:    "test",
		Status:  "ok",
		Message: "good",
	}

	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := decoded["suggestion"]; ok {
		t.Error("empty suggestion should be omitted from JSON")
	}
}

func TestStatusIcon_DefaultCase(t *testing.T) {
	t.Parallel()

	icon := statusIcon(core.CheckStatus(99))
	if icon == "" {
		t.Error("statusIcon for unknown status should return raw icon, not empty")
	}
}

func TestPluralize_AdditionalWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		word  string
		count int
		want  string
	}{
		{"task", 0, "tasks"},
		{"task", 1, "task"},
		{"task", 5, "tasks"},
		{"category", 0, "categories"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			t.Parallel()
			got := pluralize(tt.word, tt.count)
			if got != tt.want {
				t.Errorf("pluralize(%q, %d) = %q, want %q", tt.word, tt.count, got, tt.want)
			}
		})
	}
}

func TestNewDoctorCmd_Flags(t *testing.T) {
	t.Parallel()

	cmd := newDoctorCmd()

	// --verbose / -v
	vFlag := cmd.Flags().Lookup("verbose")
	if vFlag == nil {
		t.Fatal("missing --verbose flag")
	}
	if vFlag.Shorthand != "v" {
		t.Errorf("verbose shorthand = %q, want %q", vFlag.Shorthand, "v")
	}

	// --category
	catFlag := cmd.Flags().Lookup("category")
	if catFlag == nil {
		t.Fatal("missing --category flag")
	}

	// --skip-version
	svFlag := cmd.Flags().Lookup("skip-version")
	if svFlag == nil {
		t.Fatal("missing --skip-version flag")
	}
}

func TestWriteDoctorHuman_VerboseShowsCheckNames(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Environment",
				Status: core.CheckOK,
				Checks: []core.CheckResult{
					{
						Name:    "Config directory",
						Status:  core.CheckOK,
						Message: "Config directory exists (/tmp/test)",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := writeDoctorHuman(formatter, result, true); err != nil {
		t.Fatalf("writeDoctorHuman: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Name: Config directory") {
		t.Errorf("verbose output should show check name, got: %s", output)
	}
}

func TestWriteDoctorHuman_NonVerboseHidesCheckNames(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Environment",
				Status: core.CheckOK,
				Checks: []core.CheckResult{
					{
						Name:    "Config directory",
						Status:  core.CheckOK,
						Message: "Config directory exists (/tmp/test)",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := writeDoctorHuman(formatter, result, false); err != nil {
		t.Fatalf("writeDoctorHuman: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Name: Config directory") {
		t.Errorf("non-verbose output should not show check name, got: %s", output)
	}
}

func TestWriteDoctorHuman_SkippedCategory(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Environment",
				Status: core.CheckOK,
				Checks: []core.CheckResult{
					{Status: core.CheckOK, Message: "ok"},
				},
			},
			{
				Name:   "Version",
				Status: core.CheckSkip,
				Checks: []core.CheckResult{
					{Status: core.CheckSkip, Message: "Skipped (not selected)"},
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := writeDoctorHuman(formatter, result, false); err != nil {
		t.Fatalf("writeDoctorHuman: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Skipped (not selected)") {
		t.Errorf("should show skipped message, got: %s", output)
	}
}

func TestWriteDoctorHuman_SummaryWithFixable(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Sync",
				Status: core.CheckWarn,
				Checks: []core.CheckResult{
					{
						Status:     core.CheckWarn,
						Message:    "Orphaned temp files",
						Suggestion: "Run --fix to clean up",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := writeDoctorHuman(formatter, result, false); err != nil {
		t.Fatalf("writeDoctorHuman: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Doctor found issues in 1 category.") {
		t.Errorf("missing summary line, got: %s", output)
	}
	if !strings.Contains(output, "threedoors doctor --fix") {
		t.Errorf("missing fix suggestion, got: %s", output)
	}
}

func TestWriteDoctorHuman_SummaryNoFixable(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Sync",
				Status: core.CheckWarn,
				Checks: []core.CheckResult{
					{
						Status:  core.CheckWarn,
						Message: "Something off",
						// No suggestion — not fixable
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := writeDoctorHuman(formatter, result, false); err != nil {
		t.Fatalf("writeDoctorHuman: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Doctor found issues in 1 category.") {
		t.Errorf("missing summary, got: %s", output)
	}
	if strings.Contains(output, "--fix") {
		t.Errorf("should not show fix suggestion when no fixable issues, got: %s", output)
	}
}

func TestDoctorExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result core.DoctorResult
		want   int
	}{
		{
			name: "no issues",
			result: core.DoctorResult{Categories: []core.CategoryResult{
				{Status: core.CheckOK, Checks: []core.CheckResult{{Status: core.CheckOK}}},
			}},
			want: ExitSuccess,
		},
		{
			name: "warnings only",
			result: core.DoctorResult{Categories: []core.CategoryResult{
				{Status: core.CheckWarn, Checks: []core.CheckResult{{Status: core.CheckWarn}}},
			}},
			want: ExitDoctorWarning,
		},
		{
			name: "errors",
			result: core.DoctorResult{Categories: []core.CategoryResult{
				{Status: core.CheckFail, Checks: []core.CheckResult{{Status: core.CheckFail}}},
			}},
			want: ExitDoctorError,
		},
		{
			name: "mixed warns and errors",
			result: core.DoctorResult{Categories: []core.CategoryResult{
				{Status: core.CheckWarn, Checks: []core.CheckResult{{Status: core.CheckWarn}}},
				{Status: core.CheckFail, Checks: []core.CheckResult{{Status: core.CheckFail}}},
			}},
			want: ExitDoctorError,
		},
		{
			name: "skipped categories ignored",
			result: core.DoctorResult{Categories: []core.CategoryResult{
				{Status: core.CheckSkip, Checks: []core.CheckResult{{Status: core.CheckFail}}},
				{Status: core.CheckOK, Checks: []core.CheckResult{{Status: core.CheckOK}}},
			}},
			want: ExitSuccess,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := doctorExitCode(tt.result)
			if got != tt.want {
				t.Errorf("doctorExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValidCategoryKeyList(t *testing.T) {
	t.Parallel()

	keys := validCategoryKeyList()
	if len(keys) != len(core.ValidCategoryKeys) {
		t.Errorf("got %d keys, want %d", len(keys), len(core.ValidCategoryKeys))
	}
	// Verify sorted
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			t.Errorf("keys not sorted: %v", keys)
			break
		}
	}
}
