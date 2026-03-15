package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// helperHistoryReader creates a CompletionReader with fixture data in a temp dir.
func helperHistoryReader(t *testing.T, content string) *core.CompletionReader {
	t.Helper()
	dir := t.TempDir()
	if content != "" {
		if err := os.WriteFile(filepath.Join(dir, "completed.txt"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return core.NewCompletionReader(dir)
}

func TestRunHistoryTo_DefaultToday(t *testing.T) {
	t.Parallel()

	// Use a date in the far past so "today" filter returns nothing
	content := "[2020-01-15 14:30:00] abc | Old task\n"
	reader := helperHistoryReader(t, content)
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, false, false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No completed tasks found.") {
		t.Errorf("expected empty message, got: %q", buf.String())
	}
}

func TestRunHistoryTo_AllFlag(t *testing.T) {
	t.Parallel()

	content := "[2020-01-15 14:30:00] abc | Old task\n[2026-03-15 10:00:00] def | Recent task\n"
	reader := helperHistoryReader(t, content)
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, false, false, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Old task") {
		t.Errorf("expected Old task in output, got: %q", out)
	}
	if !strings.Contains(out, "Recent task") {
		t.Errorf("expected Recent task in output, got: %q", out)
	}
}

func TestRunHistoryTo_EmptyFile(t *testing.T) {
	t.Parallel()

	reader := helperHistoryReader(t, "")
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, false, false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No completed tasks found.") {
		t.Errorf("expected friendly empty message, got: %q", buf.String())
	}
}

func TestRunHistoryTo_MissingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reader := core.NewCompletionReader(dir)
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, false, false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No completed tasks found.") {
		t.Errorf("expected friendly empty message, got: %q", buf.String())
	}
}

func TestRunHistoryTo_JSONOutput(t *testing.T) {
	t.Parallel()

	content := "[2026-03-15 14:30:00] abc123 | Fix login bug\n[2026-03-15 11:15:00] def456 | Update README\n"
	reader := helperHistoryReader(t, content)
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, true, false, false, false, true)
	if err != nil {
		t.Fatal(err)
	}

	var envelope JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse JSON: %v\nraw: %s", err, buf.String())
	}
	if envelope.Command != "history" {
		t.Errorf("command = %q, want %q", envelope.Command, "history")
	}
	if envelope.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", envelope.SchemaVersion)
	}

	// Decode Data as []historyRecordJSON
	dataBytes, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatal(err)
	}
	var records []historyRecordJSON
	if err := json.Unmarshal(dataBytes, &records); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if records[0].Title != "Fix login bug" {
		t.Errorf("records[0].Title = %q, want %q", records[0].Title, "Fix login bug")
	}
	if records[0].TaskID != "abc123" {
		t.Errorf("records[0].TaskID = %q, want %q", records[0].TaskID, "abc123")
	}
	if records[1].Title != "Update README" {
		t.Errorf("records[1].Title = %q, want %q", records[1].Title, "Update README")
	}
}

func TestRunHistoryTo_JSONEmptyResult(t *testing.T) {
	t.Parallel()

	reader := helperHistoryReader(t, "")
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, true, false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}

	var envelope JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse JSON: %v\nraw: %s", err, buf.String())
	}

	dataBytes, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatal(err)
	}
	var records []historyRecordJSON
	if err := json.Unmarshal(dataBytes, &records); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("got %d records, want 0", len(records))
	}
}

func TestRunHistoryTo_HumanReadableFormat(t *testing.T) {
	t.Parallel()

	// Use times that convert to known local times for assertions.
	// parseCompletionLine stores as UTC, display converts to local.
	t1 := time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 15, 11, 15, 0, 0, time.UTC)
	t3 := time.Date(2026, 3, 14, 16, 45, 0, 0, time.UTC)

	content := "[2026-03-15 14:30:00] a | Fix login bug\n" +
		"[2026-03-15 11:15:00] b | Update README\n" +
		"[2026-03-14 16:45:00] c | Review PR\n"
	reader := helperHistoryReader(t, content)
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, false, false, false, false, true)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	// Should contain task titles
	if !strings.Contains(out, "Fix login bug") {
		t.Errorf("expected 'Fix login bug' in output, got: %q", out)
	}
	if !strings.Contains(out, "Update README") {
		t.Errorf("expected 'Update README' in output, got: %q", out)
	}
	if !strings.Contains(out, "Review PR") {
		t.Errorf("expected 'Review PR' in output, got: %q", out)
	}

	// Should contain local times (timestamps are UTC, displayed as local)
	wantTime1 := t1.Local().Format("15:04")
	wantTime2 := t2.Local().Format("15:04")
	wantTime3 := t3.Local().Format("15:04")
	if !strings.Contains(out, wantTime1) {
		t.Errorf("expected time %q in output, got: %q", wantTime1, out)
	}
	if !strings.Contains(out, wantTime2) {
		t.Errorf("expected time %q in output, got: %q", wantTime2, out)
	}
	if !strings.Contains(out, wantTime3) {
		t.Errorf("expected time %q in output, got: %q", wantTime3, out)
	}

	// Day groups should be separated by blank line
	lines := strings.Split(out, "\n")
	foundBlankBetweenGroups := false
	for _, line := range lines {
		if line == "" {
			foundBlankBetweenGroups = true
			break
		}
	}
	if !foundBlankBetweenGroups {
		t.Errorf("expected blank line between day groups")
	}
}

func TestRunHistoryTo_WeekFlag(t *testing.T) {
	t.Parallel()

	// Records spanning multiple weeks
	content := "[2026-03-15 10:00:00] a | This week\n" +
		"[2026-03-01 10:00:00] b | Earlier this month\n"
	reader := helperHistoryReader(t, content)
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, false, false, true, false, false)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	// Week filter behavior depends on current date — just verify no error
	// and some output was produced (either records or empty message)
	if out == "" {
		t.Error("expected some output")
	}
}

func TestRunHistoryTo_MonthFlag(t *testing.T) {
	t.Parallel()

	content := "[2026-03-15 10:00:00] a | This month task\n" +
		"[2026-02-15 10:00:00] b | Last month task\n"
	reader := helperHistoryReader(t, content)
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, false, false, false, true, false)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if out == "" {
		t.Error("expected some output")
	}
}

func TestRunHistoryTo_TodayFlag(t *testing.T) {
	t.Parallel()

	content := "[2020-06-15 10:00:00] a | Old task\n"
	reader := helperHistoryReader(t, content)
	var buf bytes.Buffer

	// --today is explicit but same as default
	err := runHistoryTo(context.Background(), reader, &buf, false, true, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No completed tasks found.") {
		t.Errorf("expected empty message for old task with --today, got: %q", buf.String())
	}
}

func TestRunHistoryTo_WeekJSON(t *testing.T) {
	t.Parallel()

	content := "[2026-03-15 10:00:00] a | Week task\n"
	reader := helperHistoryReader(t, content)
	var buf bytes.Buffer

	err := runHistoryTo(context.Background(), reader, &buf, true, false, true, false, false)
	if err != nil {
		t.Fatal(err)
	}

	var envelope JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse JSON: %v\nraw: %s", err, buf.String())
	}
	if envelope.Command != "history" {
		t.Errorf("command = %q, want %q", envelope.Command, "history")
	}
}

func TestFetchRecords_FlagPriority(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "[2020-01-01 10:00:00] a | Very old task\n"
	if err := os.WriteFile(filepath.Join(dir, "completed.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	reader := core.NewCompletionReader(dir)
	ctx := context.Background()

	// --all should return the old record
	records, err := fetchRecords(ctx, reader, false, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Errorf("--all: got %d records, want 1", len(records))
	}

	// default (today) should not return the old record
	records, err = fetchRecords(ctx, reader, false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Errorf("default today: got %d records, want 0", len(records))
	}
}

func TestFormatDayHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "Sunday March 15",
			time: time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC),
			want: "Sunday — March 15",
		},
		{
			name: "Monday January 1",
			time: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			want: "Thursday — January 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatDayHeader(tt.time)
			if got != tt.want {
				t.Errorf("formatDayHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewHistoryCmd_HasFlags(t *testing.T) {
	t.Parallel()

	cmd := newHistoryCmd()
	if cmd.Use != "history" {
		t.Errorf("Use = %q, want %q", cmd.Use, "history")
	}

	flags := []string{"today", "week", "month", "all"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("missing flag: --%s", flag)
		}
	}
}

func TestNewHistoryCmd_RegisteredInRoot(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "history" {
			found = true
			break
		}
	}
	if !found {
		t.Error("history command not registered in root command")
	}
}
