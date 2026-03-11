package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckSessionsFile_NoFile(t *testing.T) {
	t.Parallel()
	dc := NewDoctorChecker(t.TempDir())
	result := dc.Run()

	sessionCat := requireCategory(t, result, "Session Data")
	sessionCheck := findCheckInCategory(t, sessionCat, "Session history")

	if sessionCheck.Status != CheckInfo {
		t.Errorf("status = %v, want %v (message: %s)", sessionCheck.Status, CheckInfo, sessionCheck.Message)
	}
	if sessionCheck.Message != "No sessions recorded yet" {
		t.Errorf("message = %q, want %q", sessionCheck.Message, "No sessions recorded yet")
	}
}

func TestCheckSessionsFile_EmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sessions.jsonl"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	sessionCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Session history")
	if sessionCheck.Status != CheckInfo {
		t.Errorf("status = %v, want %v", sessionCheck.Status, CheckInfo)
	}
}

func TestCheckSessionsFile_ValidSessions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	sessions := []SessionMetrics{
		{
			SessionID: "abc-123",
			StartTime: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			SessionID: "def-456",
			StartTime: time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 3, 1, 14, 15, 0, 0, time.UTC),
		},
	}
	writeSessionsFile(t, dir, sessions)

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	sessionCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Session history")
	if sessionCheck.Status != CheckOK {
		t.Errorf("status = %v, want %v (message: %s)", sessionCheck.Status, CheckOK, sessionCheck.Message)
	}
	if sessionCheck.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestCheckSessionsFile_CorruptLines(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	content := `{"session_id":"a","start_time":"2026-01-01T00:00:00Z"}
{not valid json}
{"session_id":"b","start_time":"2026-01-02T00:00:00Z"}
{also bad
{"session_id":"c","start_time":"2026-01-03T00:00:00Z"}
`
	if err := os.WriteFile(filepath.Join(dir, "sessions.jsonl"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	sessionCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Session history")
	if sessionCheck.Status != CheckWarn {
		t.Errorf("status = %v, want %v (message: %s)", sessionCheck.Status, CheckWarn, sessionCheck.Message)
	}
	assertContains(t, sessionCheck.Message, "2 corrupt lines")
	assertContains(t, sessionCheck.Message, "lines 2, 4")
	assertContains(t, sessionCheck.Suggestion, "doctor --fix")
}

func TestCheckSessionsFile_IncompleteSessions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Line 1: valid, Line 2: missing session_id, Line 3: missing start_time (zero)
	content := `{"session_id":"a","start_time":"2026-01-01T00:00:00Z"}
{"start_time":"2026-01-02T00:00:00Z"}
{"session_id":"b"}
`
	if err := os.WriteFile(filepath.Join(dir, "sessions.jsonl"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	sessionCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Session history")
	if sessionCheck.Status != CheckWarn {
		t.Errorf("status = %v, want %v (message: %s)", sessionCheck.Status, CheckWarn, sessionCheck.Message)
	}
	assertContains(t, sessionCheck.Message, "2 incomplete entries")
}

func TestCheckSessionsFile_CorruptTakesPrecedence(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Both corrupt and incomplete present — corrupt lines are reported first
	content := `{not json}
{"start_time":"2026-01-02T00:00:00Z"}
{"session_id":"a","start_time":"2026-01-01T00:00:00Z"}
`
	if err := os.WriteFile(filepath.Join(dir, "sessions.jsonl"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	sessionCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Session history")
	if sessionCheck.Status != CheckWarn {
		t.Errorf("status = %v, want %v", sessionCheck.Status, CheckWarn)
	}
	assertContains(t, sessionCheck.Message, "corrupt lines")
}

func TestCheckPatternsFile_NoFile(t *testing.T) {
	t.Parallel()
	dc := NewDoctorChecker(t.TempDir())
	result := dc.Run()

	patternCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Pattern cache")
	if patternCheck.Status != CheckOK {
		t.Errorf("status = %v, want %v (message: %s)", patternCheck.Status, CheckOK, patternCheck.Message)
	}
	assertContains(t, patternCheck.Message, "not yet generated")
}

func TestCheckPatternsFile_ValidPatterns(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	report := PatternReport{
		GeneratedAt:  time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		SessionCount: 10,
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "patterns.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	patternCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Pattern cache")
	if patternCheck.Status != CheckOK {
		t.Errorf("status = %v, want %v (message: %s)", patternCheck.Status, CheckOK, patternCheck.Message)
	}
	assertContains(t, patternCheck.Message, "10 sessions")
}

func TestCheckPatternsFile_CorruptJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "patterns.json"), []byte("{not json}"), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	patternCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Pattern cache")
	if patternCheck.Status != CheckWarn {
		t.Errorf("status = %v, want %v (message: %s)", patternCheck.Status, CheckWarn, patternCheck.Message)
	}
	assertContains(t, patternCheck.Message, "Pattern cache corrupt")
	assertContains(t, patternCheck.Suggestion, "doctor --fix")
}

func TestCheckPatternsFile_EmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "patterns.json"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	patternCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Pattern cache")
	if patternCheck.Status != CheckWarn {
		t.Errorf("status = %v, want %v", patternCheck.Status, CheckWarn)
	}
}

func TestSessionDataCategory_RegisteredAutomatically(t *testing.T) {
	t.Parallel()
	dc := NewDoctorChecker(t.TempDir())
	result := dc.Run()

	// Verify "Session Data" exists as a category without assuming a specific index.
	found := false
	for _, cat := range result.Categories {
		if cat.Name == "Session Data" {
			found = true
			break
		}
	}
	if !found {
		names := make([]string, len(result.Categories))
		for i, cat := range result.Categories {
			names[i] = cat.Name
		}
		t.Fatalf("category %q not found in %v", "Session Data", names)
	}
}

func TestFormatLineNumbers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		lines []int
		want  string
	}{
		{"single", []int{42}, "42"},
		{"three", []int{1, 5, 10}, "1, 5, 10"},
		{"five", []int{1, 2, 3, 4, 5}, "1, 2, 3, 4, 5"},
		{"six truncates", []int{1, 2, 3, 4, 5, 6}, "1, 2, 3, 4, 5 and 1 more"},
		{"ten truncates", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "1, 2, 3, 4, 5 and 5 more"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatLineNumbers(tt.lines)
			if got != tt.want {
				t.Errorf("formatLineNumbers(%v) = %q, want %q", tt.lines, got, tt.want)
			}
		})
	}
}

func TestCheckSessionsFile_SessionDates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	sessions := []SessionMetrics{
		{
			SessionID: "first",
			StartTime: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			SessionID: "last",
			StartTime: time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC),
		},
	}
	writeSessionsFile(t, dir, sessions)

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	sessionCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Session history")
	assertContains(t, sessionCheck.Message, "first: 2025-06-15")
	assertContains(t, sessionCheck.Message, "last: 2026-03-10")
}

func TestCheckSessionsFile_ManyCorruptLines(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create file with 8 corrupt lines
	content := ""
	for i := 0; i < 8; i++ {
		content += "{bad json}\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "sessions.jsonl"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(dir)
	result := dc.Run()

	sessionCheck := findCheckInCategory(t, requireCategory(t, result, "Session Data"), "Session history")
	assertContains(t, sessionCheck.Message, "8 corrupt lines")
	assertContains(t, sessionCheck.Message, "and 3 more")
}

// --- test helpers ---

// requireCategory wraps findCategory (from doctor_database_test.go) with a fatal on nil.
func requireCategory(t *testing.T, result DoctorResult, name string) CategoryResult {
	t.Helper()
	cat := findCategory(result, name)
	if cat == nil {
		t.Fatalf("category %q not found", name)
	}
	return *cat
}

func findCheckInCategory(t *testing.T, cat CategoryResult, name string) CheckResult {
	t.Helper()
	for _, check := range cat.Checks {
		if check.Name == name {
			return check
		}
	}
	t.Fatalf("check %q not found in category %q", name, cat.Name)
	return CheckResult{}
}

func assertContains(t *testing.T, got, substr string) {
	t.Helper()
	if got == "" {
		t.Errorf("got empty string, want to contain %q", substr)
		return
	}
	for i := 0; i <= len(got)-len(substr); i++ {
		if got[i:i+len(substr)] == substr {
			return
		}
	}
	t.Errorf("got %q, want to contain %q", got, substr)
}
