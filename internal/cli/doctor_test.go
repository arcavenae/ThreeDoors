package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestDoctorCommandRegistered(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "doctor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("doctor command not registered on root")
	}
}

func TestHealthAliasResolvesToDoctor(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	for _, cmd := range root.Commands() {
		if cmd.Name() == "doctor" {
			for _, alias := range cmd.Aliases {
				if alias == "health" {
					return // pass
				}
			}
			t.Error("doctor command does not have 'health' alias")
			return
		}
	}
	t.Error("doctor command not found")
}

func TestDoctorJSONOutput(t *testing.T) {
	t.Parallel()

	dc := core.NewDoctorChecker(t.TempDir())
	result := dc.Run()

	cats := make([]doctorCategoryJSON, 0, len(result.Categories))
	for _, cat := range result.Categories {
		checks := make([]doctorCheckJSON, 0, len(cat.Checks))
		for _, check := range cat.Checks {
			checks = append(checks, doctorCheckJSON{
				Name:       check.Name,
				Status:     check.Status.String(),
				Message:    check.Message,
				Suggestion: check.Suggestion,
			})
		}
		cats = append(cats, doctorCategoryJSON{
			Name:   cat.Name,
			Status: cat.Status.String(),
			Checks: checks,
		})
	}
	data := doctorResultJSON{Categories: cats}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	if err := formatter.WriteJSON("doctor", data, nil); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if env.Command != "doctor" {
		t.Errorf("command = %q, want %q", env.Command, "doctor")
	}
	if env.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", env.SchemaVersion)
	}

	envData, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not a map: %T", env.Data)
	}

	catsRaw, ok := envData["categories"]
	if !ok {
		t.Fatal("missing 'categories' field in data")
	}
	catsArr, ok := catsRaw.([]interface{})
	if !ok {
		t.Fatalf("categories is not an array: %T", catsRaw)
	}
	if len(catsArr) == 0 {
		t.Error("expected at least one category")
	}

	// Verify first category structure
	cat0, ok := catsArr[0].(map[string]interface{})
	if !ok {
		t.Fatalf("category[0] is not a map: %T", catsArr[0])
	}
	if _, ok := cat0["name"]; !ok {
		t.Error("category missing 'name' field")
	}
	if _, ok := cat0["status"]; !ok {
		t.Error("category missing 'status' field")
	}
	if _, ok := cat0["checks"]; !ok {
		t.Error("category missing 'checks' field")
	}
}

func TestDoctorHumanOutput(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Environment",
				Status: core.CheckOK,
				Checks: []core.CheckResult{
					{Name: "Config directory", Status: core.CheckOK, Message: "Config directory exists (/tmp/test)"},
					{Name: "Config file", Status: core.CheckOK, Message: "Config file valid (schema v2)"},
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)
	_ = formatter.Writef("ThreeDoors Doctor (dev • test/amd64)\n\n")
	for _, cat := range result.Categories {
		_ = formatter.Writef("%s %s\n", cat.Status.Icon(), cat.Name)
		for _, check := range cat.Checks {
			_ = formatter.Writef("    %s %s\n", check.Status.Icon(), check.Message)
		}
		_ = formatter.Writef("\n")
	}
	_ = formatter.Writef("No issues found\n")

	output := buf.String()
	if !strings.Contains(output, "ThreeDoors Doctor") {
		t.Error("missing header line")
	}
	if !strings.Contains(output, "Environment") {
		t.Error("missing Environment category")
	}
	if !strings.Contains(output, "No issues found") {
		t.Error("missing summary line")
	}
}

func TestDoctorHumanOutput_WithIssues(t *testing.T) {
	t.Parallel()

	result := core.DoctorResult{
		Categories: []core.CategoryResult{
			{
				Name:   "Environment",
				Status: core.CheckWarn,
				Checks: []core.CheckResult{
					{Name: "Config directory", Status: core.CheckOK, Message: "ok"},
					{Name: "Config file", Status: core.CheckWarn, Message: "missing", Suggestion: "create it"},
				},
			},
		},
	}

	warnings, errors := result.IssueCount()
	total := warnings + errors
	catCount := result.CategoryIssueCount()

	summary := strings.Builder{}
	fmt.Fprintf(&summary, "%d %s in %d %s",
		total, pluralize("issue", total),
		catCount, pluralize("category", catCount))

	if !strings.Contains(summary.String(), "1 issue in 1 category") {
		t.Errorf("summary = %q, want containing '1 issue in 1 category'", summary.String())
	}
}

func TestPluralize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		word  string
		count int
		want  string
	}{
		{"issue", 0, "issues"},
		{"issue", 1, "issue"},
		{"issue", 2, "issues"},
		{"category", 1, "category"},
		{"category", 2, "categories"},
	}
	for _, tt := range tests {
		t.Run(strings.Join([]string{tt.word, strings.Repeat("s", tt.count)}, "_"), func(t *testing.T) {
			t.Parallel()
			if got := pluralize(tt.word, tt.count); got != tt.want {
				t.Errorf("pluralize(%q, %d) = %q, want %q", tt.word, tt.count, got, tt.want)
			}
		})
	}
}

func TestStatusIcon(t *testing.T) {
	t.Parallel()

	// Just verify it doesn't panic and returns non-empty for each status
	statuses := []core.CheckStatus{core.CheckOK, core.CheckInfo, core.CheckSkip, core.CheckWarn, core.CheckFail}
	for _, s := range statuses {
		icon := statusIcon(s)
		if icon == "" {
			t.Errorf("statusIcon(%v) returned empty string", s)
		}
	}
}
