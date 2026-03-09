package tui

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/muesli/termenv"
)

// makeInsightsTestSession creates a SessionMetrics for insights view tests.
func makeInsightsTestSession(startTime time.Time, completed int, moods []string, doorPositions []int) core.SessionMetrics {
	entries := make([]core.MoodEntry, len(moods))
	for i, m := range moods {
		entries[i] = core.MoodEntry{Mood: m, Timestamp: startTime}
	}
	selections := make([]core.DoorSelectionRecord, len(doorPositions))
	for i, p := range doorPositions {
		selections[i] = core.DoorSelectionRecord{DoorPosition: p, TaskText: "task", Timestamp: startTime}
	}
	return core.SessionMetrics{
		SessionID:       uuid.New().String(),
		StartTime:       startTime,
		EndTime:         startTime.Add(30 * time.Minute),
		DurationSeconds: 1800,
		TasksCompleted:  completed,
		MoodEntries:     entries,
		MoodEntryCount:  len(moods),
		DoorSelections:  selections,
	}
}

// writeInsightsSessionsFile creates a sessions.jsonl for insights tests.
func writeInsightsSessionsFile(t *testing.T, dir string, sessions []core.SessionMetrics) string {
	t.Helper()
	path := filepath.Join(dir, "sessions.jsonl")
	var buf bytes.Buffer
	for _, s := range sessions {
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	return path
}

func setupInsightsView(t *testing.T) *InsightsView {
	t.Helper()
	dir := t.TempDir()
	now := time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	sessions := []core.SessionMetrics{
		makeInsightsTestSession(time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC), 3, []string{"Focused"}, []int{0, 1}),
		makeInsightsTestSession(time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC), 5, []string{"Tired"}, []int{1, 1, 2}),
		makeInsightsTestSession(time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC), 4, []string{"Focused", "Energized"}, []int{0, 2}),
	}
	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := core.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	cc := core.NewCompletionCounterWithNow(frozen)

	iv := NewInsightsView(pa, cc)
	iv.SetWidth(80)
	return iv
}

func TestNewInsightsView(t *testing.T) {
	iv := setupInsightsView(t)
	if iv == nil {
		t.Fatal("NewInsightsView() returned nil")
	}
}

func TestInsightsView_View_ContainsSections(t *testing.T) {
	iv := setupInsightsView(t)
	output := iv.View()

	expectedSections := []string{
		"YOUR INSIGHTS DASHBOARD",
		"COMPLETION TRENDS",
		"STREAKS",
		"MOOD & PRODUCTIVITY",
		"DOOR PICKS",
		"Press Esc to return",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("View() output missing section %q", section)
		}
	}
}

func TestInsightsView_View_HeroNumber(t *testing.T) {
	iv := setupInsightsView(t)
	output := iv.View()

	// Total tasks: 3 + 5 + 4 = 12
	if !strings.Contains(output, "12 tasks completed") {
		t.Errorf("View() should contain hero number with total 12, got:\n%s", output)
	}
	if !strings.Contains(output, "★") {
		t.Errorf("View() hero number should contain star decoration")
	}
}

func TestInsightsView_View_ColdStart(t *testing.T) {
	// Only 1 session — below threshold
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	sessions := []core.SessionMetrics{
		makeInsightsTestSession(now, 2, []string{"Focused"}, []int{1}),
	}
	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := core.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	cc := core.NewCompletionCounterWithNow(frozen)
	iv := NewInsightsView(pa, cc)
	iv.SetWidth(80)

	output := iv.View()
	if !strings.Contains(output, "Keep using ThreeDoors to unlock insights!") {
		t.Errorf("cold start view should contain unlock message, got: %q", output)
	}
	// Cold start should NOT show hero number or empty panels
	if strings.Contains(output, "tasks completed") {
		t.Errorf("cold start should not show hero number")
	}
	if strings.Contains(output, "COMPLETION TRENDS") {
		t.Errorf("cold start should not show empty panels")
	}
}

func TestInsightsView_View_ColdStartStyled(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	sessions := []core.SessionMetrics{
		makeInsightsTestSession(now, 2, []string{"Focused"}, []int{1}),
	}
	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := core.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	cc := core.NewCompletionCounterWithNow(frozen)
	iv := NewInsightsView(pa, cc)
	iv.SetWidth(80)

	output := iv.View()
	// Should still have the dashboard header
	if !strings.Contains(output, "YOUR INSIGHTS DASHBOARD") {
		t.Errorf("cold start should still show dashboard header")
	}
	// Should have the Esc help
	if !strings.Contains(output, "Press Esc to return") {
		t.Errorf("cold start should show Esc help")
	}
}

func TestInsightsView_Update_EscReturns(t *testing.T) {
	iv := setupInsightsView(t)

	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc should produce a command")
	}

	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("Esc should produce ReturnToDoorsMsg, got %T", msg)
	}
}

func TestInsightsView_SetWidth(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetWidth(120)

	output := iv.View()
	if output == "" {
		t.Error("View() should not return empty string after SetWidth")
	}
}

func TestInsightsView_SetHeight(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetHeight(24)

	if iv.height != 24 {
		t.Errorf("SetHeight(24) should set height to 24, got %d", iv.height)
	}
}

func TestInsightsView_LayoutMode(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		wantMode string
	}{
		{"compact at 40", 40, layoutCompact},
		{"compact at 59", 59, layoutCompact},
		{"narrow at 60", 60, layoutNarrow},
		{"narrow at 79", 79, layoutNarrow},
		{"standard at 80", 80, layoutStandard},
		{"standard at 119", 119, layoutStandard},
		{"wide at 120", 120, layoutWide},
		{"wide at 200", 200, layoutWide},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iv := setupInsightsView(t)
			iv.SetWidth(tt.width)
			got := iv.layoutMode()
			if got != tt.wantMode {
				t.Errorf("layoutMode() at width %d = %q, want %q", tt.width, got, tt.wantMode)
			}
		})
	}
}

func TestInsightsView_RenderCache(t *testing.T) {
	iv := setupInsightsView(t)

	// First render populates cache
	first := iv.View()
	if !iv.cacheValid {
		t.Error("cache should be valid after first render")
	}

	// Second render should return cached view
	second := iv.View()
	if first != second {
		t.Error("cached view should be identical to first render")
	}

	// Changing width invalidates cache
	iv.SetWidth(120)
	if iv.cacheValid {
		t.Error("cache should be invalid after SetWidth")
	}

	// Re-render produces new output
	third := iv.View()
	if !iv.cacheValid {
		t.Error("cache should be valid after re-render")
	}
	// At 120 width, layout is different from 80 width
	if third == first {
		t.Error("view at 120 width should differ from view at 80 width")
	}
}

func TestInsightsView_OutputWidthNeverExceedsTerminal(t *testing.T) {
	widths := []int{40, 60, 80, 100, 120, 160}

	for _, width := range widths {
		t.Run(strings.ReplaceAll(strings.TrimSpace(lipgloss.NewStyle().Render("")), "", "")+
			"width_"+strings.ReplaceAll(strings.TrimSpace(lipgloss.NewStyle().Render("")), "", "")+
			string(rune('0'+width/100))+string(rune('0'+(width%100)/10))+string(rune('0'+width%10)),
			func(t *testing.T) {
				iv := setupInsightsView(t)
				iv.SetWidth(width)
				output := iv.View()

				for i, line := range strings.Split(output, "\n") {
					// Use lipgloss.Width for ANSI-aware measurement
					lineWidth := lipgloss.Width(line)
					if lineWidth > width {
						t.Errorf("line %d exceeds terminal width %d (got %d): %q",
							i+1, width, lineWidth, line)
					}
				}
			})
	}
}

func TestInsightsView_CompactLayoutSingleColumn(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetWidth(50) // compact mode
	output := iv.View()

	// All sections should be present
	sections := []string{"COMPLETION TRENDS", "STREAKS", "MOOD & PRODUCTIVITY", "DOOR PICKS"}
	for _, section := range sections {
		if !strings.Contains(output, section) {
			t.Errorf("compact layout missing section %q", section)
		}
	}
}

func TestInsightsView_WideLayoutThreeColumns(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetWidth(140) // wide mode

	output := iv.View()

	// All sections present
	sections := []string{"COMPLETION TRENDS", "STREAKS", "MOOD & PRODUCTIVITY", "DOOR PICKS"}
	for _, section := range sections {
		if !strings.Contains(output, section) {
			t.Errorf("wide layout missing section %q", section)
		}
	}
}

func TestInsightsView_PanelsHaveBorders(t *testing.T) {
	iv := setupInsightsView(t)
	output := iv.View()

	// Rounded border uses ╭ ╮ ╰ ╯ characters
	borderChars := []string{"╭", "╮", "╰", "╯"}
	for _, ch := range borderChars {
		if !strings.Contains(output, ch) {
			t.Errorf("View() output should contain rounded border character %q", ch)
		}
	}
}

func TestInsightsView_DashboardHeaderPresent(t *testing.T) {
	iv := setupInsightsView(t)
	output := iv.View()

	if !strings.Contains(output, "YOUR INSIGHTS DASHBOARD") {
		t.Error("View() should contain 'YOUR INSIGHTS DASHBOARD' header")
	}
}

func TestInsightsView_InvalidateCache(t *testing.T) {
	iv := setupInsightsView(t)
	_ = iv.View()
	if !iv.cacheValid {
		t.Fatal("cache should be valid after render")
	}

	iv.invalidateCache()
	if iv.cacheValid {
		t.Error("invalidateCache() should mark cache as invalid")
	}
}

func TestInsightsView_BuildStreaks_NoStreak(t *testing.T) {
	iv := setupInsightsView(t)
	content := iv.buildStreaks()
	if !strings.Contains(content, "No active streak") {
		t.Errorf("buildStreaks() with no streak should contain 'No active streak', got %q", content)
	}
}

func TestInsightsView_BuildDoorPreferences(t *testing.T) {
	iv := setupInsightsView(t)
	content := iv.buildDoorPreferences()

	if !strings.Contains(content, "Left:") || !strings.Contains(content, "Center:") || !strings.Contains(content, "Right:") {
		t.Errorf("buildDoorPreferences() should contain Left/Center/Right, got %q", content)
	}
}

func TestInsightsView_BuildCompletionTrends(t *testing.T) {
	iv := setupInsightsView(t)
	content := iv.buildCompletionTrends()

	if !strings.Contains(content, "This week:") {
		t.Errorf("buildCompletionTrends() should contain week-over-week, got %q", content)
	}
}

func TestInsightsView_ContentWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  int
	}{
		{"zero width defaults to 78", 0, 78},
		{"standard width", 80, 78},
		{"very small", 2, 2},
		{"one", 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iv := &InsightsView{width: tt.width}
			got := iv.contentWidth()
			if got != tt.want {
				t.Errorf("contentWidth() = %d, want %d", got, tt.want)
			}
		})
	}
}

// Golden file tests for consistent rendering at standard widths.
// Use Ascii profile for deterministic output regardless of test execution order.
func TestInsightsView_GoldenFile_80col(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsView(t)
	iv.SetWidth(80)
	output := iv.View()

	goldenPath := filepath.Join("testdata", "insights_80col.golden")
	updateGoldenFile(t, goldenPath, output)
	compareGoldenFile(t, goldenPath, output)
}

func TestInsightsView_GoldenFile_120col(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsView(t)
	iv.SetWidth(120)
	output := iv.View()

	goldenPath := filepath.Join("testdata", "insights_120col.golden")
	updateGoldenFile(t, goldenPath, output)
	compareGoldenFile(t, goldenPath, output)
}

// updateGoldenFile writes a golden file if UPDATE_GOLDEN env var is set or file doesn't exist.
func updateGoldenFile(t *testing.T, path, content string) {
	t.Helper()
	if os.Getenv("UPDATE_GOLDEN") != "" {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create testdata dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
		return
	}
	// Create if doesn't exist (first run)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create testdata dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
	}
}

// compareGoldenFile compares output against the golden file.
func compareGoldenFile(t *testing.T, path, actual string) {
	t.Helper()
	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file %s: %v", path, err)
	}
	if string(expected) != actual {
		t.Errorf("output does not match golden file %s.\nRun with UPDATE_GOLDEN=1 to update.\n\nExpected:\n%s\n\nActual:\n%s",
			path, string(expected), actual)
	}
}
