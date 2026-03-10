package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/tui/themes"
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

	iv := NewInsightsView(pa, cc, nil, nil)
	iv.SetWidth(80)
	return iv
}

func TestNewInsightsView(t *testing.T) {
	iv := setupInsightsView(t)
	if iv == nil {
		t.Fatal("NewInsightsView() returned nil")
		return
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
	iv := NewInsightsView(pa, cc, nil, nil)
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
	iv := NewInsightsView(pa, cc, nil, nil)
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
		return
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

// --- Story 40.4: Bar chart tests ---

func TestBarChart(t *testing.T) {
	// Use Ascii profile so bar output has no ANSI codes for easy comparison.
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	tests := []struct {
		name  string
		ratio float64
		width int
		want  string
	}{
		{"full bar", 1.0, 10, "██████████"},
		{"empty bar", 0.0, 10, "░░░░░░░░░░"},
		{"half bar", 0.5, 10, "█████░░░░░"},
		{"zero width", 0.5, 0, ""},
		{"ratio clamped above 1", 1.5, 5, "█████"},
		{"ratio clamped below 0", -0.5, 5, "░░░░░"},
		{"one char full", 1.0, 1, "█"},
		{"rounding up at 0.75", 0.75, 4, "███░"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := barChart(tt.ratio, tt.width, defaultMoodColor)
			if got != tt.want {
				t.Errorf("barChart(%v, %d) = %q, want %q", tt.ratio, tt.width, got, tt.want)
			}
		})
	}
}

func TestBuildMoodCorrelations_BarChars(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	// Create a view with multiple moods at different averages so both █ and ░ appear.
	dir := t.TempDir()
	now := time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	sessions := []core.SessionMetrics{
		makeInsightsTestSession(time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC), 5, []string{"Focused"}, []int{0}),
		makeInsightsTestSession(time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC), 2, []string{"Tired"}, []int{1}),
		makeInsightsTestSession(time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC), 5, []string{"Focused"}, []int{0}),
		makeInsightsTestSession(time.Date(2026, 3, 7, 11, 0, 0, 0, time.UTC), 2, []string{"Tired"}, []int{1}),
	}
	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := core.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}
	cc := core.NewCompletionCounterWithNow(frozen)
	iv := NewInsightsView(pa, cc, nil, nil)
	iv.SetWidth(80)

	content := iv.buildMoodCorrelations()

	// Should contain both bar characters (Focused=full, Tired=partial)
	if !strings.Contains(content, "█") {
		t.Errorf("mood correlations should contain filled bar chars (█), got:\n%s", content)
	}
	if !strings.Contains(content, "░") {
		t.Errorf("mood correlations should contain empty bar chars (░), got:\n%s", content)
	}

	// Should still contain summary info
	if !strings.Contains(content, "Most productive mood:") {
		t.Errorf("mood correlations should still contain 'Most productive mood:', got:\n%s", content)
	}
}

func TestBuildMoodCorrelations_HighestMoodFullBar(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsView(t)
	iv.SetWidth(80) // standard layout → barWidth = 20
	content := iv.buildMoodCorrelations()

	// The highest mood (first in sorted list) should have a full bar (20 █ chars).
	fullBar := strings.Repeat("█", 20)
	if !strings.Contains(content, fullBar) {
		t.Errorf("highest mood should have full bar (%d █ chars), got:\n%s", 20, content)
	}
}

func TestBuildMoodCorrelations_EmptyData(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	// Create sessions without mood data
	sessions := []core.SessionMetrics{
		makeInsightsTestSession(now.Add(-48*time.Hour), 3, nil, []int{0}),
		makeInsightsTestSession(now.Add(-24*time.Hour), 2, nil, []int{1}),
		makeInsightsTestSession(now, 4, nil, []int{2}),
	}
	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := core.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}
	cc := core.NewCompletionCounterWithNow(frozen)
	iv := NewInsightsView(pa, cc, nil, nil)
	iv.SetWidth(80)

	content := iv.buildMoodCorrelations()
	if !strings.Contains(content, "Not enough mood data yet") {
		t.Errorf("empty mood data should show message, got: %q", content)
	}
	if strings.Contains(content, "█") || strings.Contains(content, "░") {
		t.Errorf("empty mood data should NOT show bar chars, got: %q", content)
	}
}

func TestBuildMoodCorrelations_CompactBarWidth(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsView(t)
	iv.SetWidth(50) // compact mode → barWidth = 10
	content := iv.buildMoodCorrelations()

	// The highest mood should have a bar of width 10.
	fullBar := strings.Repeat("█", 10)
	if !strings.Contains(content, fullBar) {
		t.Errorf("compact mode highest mood should have 10-char bar, got:\n%s", content)
	}
	// Should NOT have a 20-char bar.
	longBar := strings.Repeat("█", 20)
	if strings.Contains(content, longBar) {
		t.Errorf("compact mode should not have 20-char bar")
	}
}

func TestMoodBarWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  int
	}{
		{"compact", 40, 10},
		{"narrow", 70, 15},
		{"standard", 80, 20},
		{"wide", 140, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iv := setupInsightsView(t)
			iv.SetWidth(tt.width)
			got := iv.moodBarWidth()
			if got != tt.want {
				t.Errorf("moodBarWidth() at width %d = %d, want %d", tt.width, got, tt.want)
			}
		})
	}
}

// ============================================================
// Story 40.6 — Session Highlights Tests
// ============================================================

func setupInsightsViewWithHighlights(t *testing.T) *InsightsView {
	t.Helper()
	dir := t.TempDir()
	now := time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	sessions := []core.SessionMetrics{
		{
			SessionID:           uuid.New().String(),
			StartTime:           time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC),
			EndTime:             time.Date(2026, 3, 5, 10, 30, 0, 0, time.UTC),
			DurationSeconds:     1800,
			TasksCompleted:      3,
			DetailViews:         2,
			NotesAdded:          1,
			TimeToFirstDoorSecs: 8.0,
			DoorSelections:      []core.DoorSelectionRecord{{DoorPosition: 0, TaskText: "t1", Timestamp: now}},
			MoodEntries:         []core.MoodEntry{{Mood: "Focused", Timestamp: now}},
			MoodEntryCount:      1,
		},
		{
			SessionID:           uuid.New().String(),
			StartTime:           time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC),
			EndTime:             time.Date(2026, 3, 6, 11, 0, 0, 0, time.UTC),
			DurationSeconds:     3600,
			TasksCompleted:      5,
			DetailViews:         3,
			NotesAdded:          2,
			TimeToFirstDoorSecs: 4.5,
			DoorSelections:      []core.DoorSelectionRecord{{DoorPosition: 1, TaskText: "t2", Timestamp: now}, {DoorPosition: 2, TaskText: "t3", Timestamp: now}},
			MoodEntries:         []core.MoodEntry{{Mood: "Tired", Timestamp: now}},
			MoodEntryCount:      1,
		},
		{
			SessionID:           uuid.New().String(),
			StartTime:           time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
			EndTime:             time.Date(2026, 3, 7, 10, 15, 0, 0, time.UTC),
			DurationSeconds:     900,
			TasksCompleted:      4,
			DetailViews:         0,
			NotesAdded:          0,
			TimeToFirstDoorSecs: 2.0,
			DoorSelections:      []core.DoorSelectionRecord{{DoorPosition: 0, TaskText: "t4", Timestamp: now}},
			MoodEntries:         []core.MoodEntry{{Mood: "Focused", Timestamp: now}},
			MoodEntryCount:      1,
		},
	}

	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := core.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	cc := core.NewCompletionCounterWithNow(frozen)
	iv := NewInsightsView(pa, cc, nil, nil)
	iv.SetWidth(80)
	return iv
}

func TestInsightsView_SessionHighlights_PanelPresent(t *testing.T) {
	iv := setupInsightsViewWithHighlights(t)
	output := iv.View()

	if !strings.Contains(output, "SESSION HIGHLIGHTS") {
		t.Error("View() should contain 'SESSION HIGHLIGHTS' panel")
	}
}

func TestInsightsView_SessionHighlights_MetricsShown(t *testing.T) {
	iv := setupInsightsViewWithHighlights(t)
	output := iv.View()

	expectedMetrics := []string{
		"Doors opened",
		"Tasks completed",
		"Avg session",
		"Fastest first pick",
		"Detail views",
		"Notes added",
		"Longest streak",
		"Peak hour",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(output, metric) {
			t.Errorf("View() missing metric %q", metric)
		}
	}
}

func TestInsightsView_SessionHighlights_Values(t *testing.T) {
	iv := setupInsightsViewWithHighlights(t)
	content := iv.buildSessionHighlights()

	// Total doors: 1 + 2 + 1 = 4
	if !strings.Contains(content, "4") {
		t.Errorf("highlights should contain total doors 4, got:\n%s", content)
	}
	// Total tasks: 3 + 5 + 4 = 12
	if !strings.Contains(content, "12") {
		t.Errorf("highlights should contain total tasks 12, got:\n%s", content)
	}
	// Peak hour: 10 appears 3 times
	if !strings.Contains(content, "10am") {
		t.Errorf("highlights should contain peak hour '10am', got:\n%s", content)
	}
	// Longest streak: Mar 5, 6, 7 = 3 days
	if !strings.Contains(content, "3 days") {
		t.Errorf("highlights should contain '3 days' streak, got:\n%s", content)
	}
}

func TestInsightsView_SessionHighlights_OmitsZeroData(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	// Session with no detail views, no notes, no first door time
	sessions := []core.SessionMetrics{
		{
			SessionID:           uuid.New().String(),
			StartTime:           time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC),
			DurationSeconds:     600,
			TasksCompleted:      1,
			DetailViews:         0,
			NotesAdded:          0,
			TimeToFirstDoorSecs: 0,
			DoorSelections:      []core.DoorSelectionRecord{{DoorPosition: 0, TaskText: "t1", Timestamp: now}},
			MoodEntries:         []core.MoodEntry{{Mood: "Focused", Timestamp: now}},
			MoodEntryCount:      1,
		},
		{
			SessionID:           uuid.New().String(),
			StartTime:           time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC),
			DurationSeconds:     600,
			TasksCompleted:      2,
			DetailViews:         0,
			NotesAdded:          0,
			TimeToFirstDoorSecs: 0,
			DoorSelections:      []core.DoorSelectionRecord{{DoorPosition: 1, TaskText: "t2", Timestamp: now}},
			MoodEntries:         []core.MoodEntry{{Mood: "Tired", Timestamp: now}},
			MoodEntryCount:      1,
		},
		{
			SessionID:       uuid.New().String(),
			StartTime:       time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
			DurationSeconds: 600,
			TasksCompleted:  1,
			DoorSelections:  []core.DoorSelectionRecord{{DoorPosition: 2, TaskText: "t3", Timestamp: now}},
			MoodEntries:     []core.MoodEntry{{Mood: "Focused", Timestamp: now}},
			MoodEntryCount:  1,
		},
	}

	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := core.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	cc := core.NewCompletionCounterWithNow(frozen)
	iv := NewInsightsView(pa, cc, nil, nil)
	iv.SetWidth(80)

	content := iv.buildSessionHighlights()

	if strings.Contains(content, "Detail views") {
		t.Error("should omit 'Detail views' when count is 0")
	}
	if strings.Contains(content, "Notes added") {
		t.Error("should omit 'Notes added' when count is 0")
	}
	if strings.Contains(content, "Fastest first pick") {
		t.Error("should omit 'Fastest first pick' when no valid data")
	}

	// Should still show metrics that have data
	if !strings.Contains(content, "Doors opened") {
		t.Error("should show 'Doors opened' when doors > 0")
	}
	if !strings.Contains(content, "Tasks completed") {
		t.Error("should show 'Tasks completed' when tasks > 0")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"milliseconds", 500 * time.Millisecond, "500ms"},
		{"seconds integer", 5 * time.Second, "5s"},
		{"seconds fractional", 5500 * time.Millisecond, "5.5s"},
		{"minutes only", 15 * time.Minute, "15m"},
		{"minutes and seconds", 15*time.Minute + 30*time.Second, "15m 30s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFormatHour(t *testing.T) {
	tests := []struct {
		hour int
		want string
	}{
		{0, "12am"},
		{1, "1am"},
		{9, "9am"},
		{11, "11am"},
		{12, "12pm"},
		{13, "1pm"},
		{23, "11pm"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatHour(tt.hour)
			if got != tt.want {
				t.Errorf("formatHour(%d) = %q, want %q", tt.hour, got, tt.want)
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

func TestSparkline(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		want   string
	}{
		{"empty", nil, ""},
		{"all zeros", []int{0, 0, 0}, "▁▁▁"},
		{"single max", []int{5}, "█"},
		{"single zero", []int{0}, "▁"},
		{"ascending", []int{0, 1, 2, 3, 4, 5, 6, 7}, "▁▂▃▄▅▆▇█"},
		{"mixed", []int{3, 0, 5, 1}, "▅▁█▂"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sparkline(tt.values)
			if got != tt.want {
				t.Errorf("sparkline(%v) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}

func TestBlendHexColors(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		t    float64
		want string
	}{
		{"start color at 0.0", "#3B82F6", "#EAB308", 0.0, "#3B82F6"},
		{"end color at 1.0", "#3B82F6", "#EAB308", 1.0, "#EAB308"},
		{"midpoint", "#000000", "#FFFFFF", 0.5, "#7F7F7F"},
		{"clamp below 0", "#000000", "#FFFFFF", -1.0, "#000000"},
		{"clamp above 1", "#000000", "#FFFFFF", 2.0, "#FFFFFF"},
		{"quarter", "#000000", "#FF0000", 0.25, "#3F0000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blendHexColors(tt.from, tt.to, tt.t)
			if !strings.EqualFold(got, tt.want) {
				t.Errorf("blendHexColors(%q, %q, %.1f) = %q, want %q", tt.from, tt.to, tt.t, got, tt.want)
			}
		})
	}
}

// --- Tab Navigation Tests (Story 40.8) ---

func TestInsightsView_TabSwitchesToDetail(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetHeight(24)

	// Initially on Overview tab (activeTab == 0)
	if iv.activeTab != 0 {
		t.Fatalf("initial activeTab should be 0, got %d", iv.activeTab)
	}

	// Press Tab → switches to Detail tab
	iv.Update(tea.KeyMsg{Type: tea.KeyTab})
	if iv.activeTab != 1 {
		t.Errorf("after Tab, activeTab should be 1, got %d", iv.activeTab)
	}
}

func TestInsightsView_ShiftTabSwitchesBack(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetHeight(24)

	// Switch to Detail tab first
	iv.Update(tea.KeyMsg{Type: tea.KeyTab})
	if iv.activeTab != 1 {
		t.Fatalf("after Tab, activeTab should be 1, got %d", iv.activeTab)
	}

	// Press Shift-Tab → switches back to Overview
	iv.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if iv.activeTab != 0 {
		t.Errorf("after Shift-Tab, activeTab should be 0, got %d", iv.activeTab)
	}
}

func TestInsightsView_TabWraps(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetHeight(24)

	// Tab from Overview wraps to Detail
	iv.Update(tea.KeyMsg{Type: tea.KeyTab})
	if iv.activeTab != 1 {
		t.Fatalf("after first Tab, activeTab should be 1, got %d", iv.activeTab)
	}

	// Tab from Detail wraps back to Overview
	iv.Update(tea.KeyMsg{Type: tea.KeyTab})
	if iv.activeTab != 0 {
		t.Errorf("after second Tab (wrap), activeTab should be 0, got %d", iv.activeTab)
	}
}

func TestInsightsView_ShiftTabWraps(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetHeight(24)

	// Shift-Tab from Overview wraps to Detail
	iv.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if iv.activeTab != 1 {
		t.Errorf("Shift-Tab from Overview should wrap to Detail (1), got %d", iv.activeTab)
	}
}

func TestInsightsView_EscReturnsFromEitherTab(t *testing.T) {
	tests := []struct {
		name string
		tab  int
	}{
		{"from Overview", 0},
		{"from Detail", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iv := setupInsightsView(t)
			iv.SetHeight(24)
			iv.activeTab = tt.tab

			cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEsc})
			if cmd == nil {
				t.Fatal("Esc should produce a command")
				return
			}
			msg := cmd()
			if _, ok := msg.(ReturnToDoorsMsg); !ok {
				t.Errorf("Esc should produce ReturnToDoorsMsg, got %T", msg)
			}
		})
	}
}

func TestStyledSparkline(t *testing.T) {
	t.Run("empty input returns empty string", func(t *testing.T) {
		got := styledSparkline(nil)
		if got != "" {
			t.Errorf("styledSparkline(nil) = %q, want empty", got)
		}
	})

	t.Run("all zeros produces styled output", func(t *testing.T) {
		got := styledSparkline([]int{0, 0, 0})
		if got == "" {
			t.Error("styledSparkline([]int{0,0,0}) should not be empty")
		}
		// Each character should be ▁ (styled)
		if !strings.Contains(got, "▁") {
			t.Errorf("all-zero sparkline should contain ▁, got %q", got)
		}
	})

	t.Run("single value produces output", func(t *testing.T) {
		got := styledSparkline([]int{5})
		if got == "" {
			t.Error("styledSparkline single value should not be empty")
		}
		if !strings.Contains(got, "█") {
			t.Errorf("single nonzero sparkline should contain █, got %q", got)
		}
	})

	t.Run("output contains all sparkline characters", func(t *testing.T) {
		values := []int{0, 3, 7}
		got := styledSparkline(values)
		// Should contain ▁ for 0, some mid char for 3, █ for 7
		if !strings.Contains(got, "▁") {
			t.Errorf("should contain ▁ for zero value, got %q", got)
		}
		if !strings.Contains(got, "█") {
			t.Errorf("should contain █ for max value, got %q", got)
		}
	})

	t.Run("uses ANSI color codes when color enabled", func(t *testing.T) {
		lipgloss.SetColorProfile(termenv.TrueColor)
		t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

		got := styledSparkline([]int{0, 5, 10})
		if !strings.Contains(got, "\x1b[") {
			t.Errorf("styled sparkline should contain ANSI codes with TrueColor, got %q", got)
		}
	})
}

func TestStyledSparklineChars(t *testing.T) {
	t.Run("returns correct number of chars", func(t *testing.T) {
		values := []int{1, 2, 3, 4, 5}
		chars := styledSparklineChars(values)
		if len(chars) != len(values) {
			t.Errorf("styledSparklineChars returned %d chars, want %d", len(chars), len(values))
		}
	})

	t.Run("nil returns nil", func(t *testing.T) {
		chars := styledSparklineChars(nil)
		if chars != nil {
			t.Errorf("styledSparklineChars(nil) should return nil, got %v", chars)
		}
	})

	t.Run("all zeros returns minimum height chars", func(t *testing.T) {
		chars := styledSparklineChars([]int{0, 0, 0})
		for i, ch := range chars {
			if !strings.Contains(ch, "▁") {
				t.Errorf("char %d should contain ▁ for zero value, got %q", i, ch)
			}
		}
	})
}

func TestBuildCompletionTrends_ContainsSparklineChars(t *testing.T) {
	iv := setupInsightsView(t)
	content := iv.buildCompletionTrends()
	// Should contain sparkline block characters
	hasBlock := false
	for _, ch := range sparkChars {
		if strings.ContainsRune(content, ch) {
			hasBlock = true
			break
		}
	}
	if !hasBlock {
		t.Errorf("buildCompletionTrends() should contain sparkline block characters, got:\n%s", content)
	}
}

func TestInsightsView_View_ContainsFunFact(t *testing.T) {
	iv := setupInsightsView(t)
	output := iv.View()

	// Fun fact should contain the gold star prefix
	if !strings.Contains(output, "★") {
		t.Error("View() should contain gold star for fun fact")
	}
}

func TestInsightsView_View_ColdStart_NoFunFact(t *testing.T) {
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
	iv := NewInsightsView(pa, cc, nil, nil)
	iv.SetWidth(80)

	output := iv.View()
	// Cold start should NOT show fun fact panels (only the cold start message)
	if strings.Contains(output, "COMPLETION TRENDS") {
		t.Error("cold start should not show data panels including fun facts")
	}
}

func TestInsightsView_TabIndicatorOverview(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsView(t)
	iv.SetHeight(24)
	iv.activeTab = 0

	output := iv.View()
	if !strings.Contains(output, "[Overview]") {
		t.Errorf("Overview tab indicator should show [Overview] when active, output:\n%s", output)
	}
	if strings.Contains(output, "[Detail]") {
		t.Errorf("Detail should NOT be bracketed when Overview is active")
	}
}

func TestInsightsView_TabIndicatorDetail(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsView(t)
	iv.SetHeight(24)
	iv.activeTab = 1

	output := iv.View()
	if !strings.Contains(output, "[Detail]") {
		t.Errorf("Detail tab indicator should show [Detail] when active, output:\n%s", output)
	}
	if strings.Contains(output, "[Overview]") {
		t.Errorf("Overview should NOT be bracketed when Detail is active")
	}
}

func TestInsightsView_TabInvalidatesCache(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetHeight(24)

	_ = iv.View() // populate cache
	if !iv.cacheValid {
		t.Fatal("cache should be valid after render")
	}

	iv.Update(tea.KeyMsg{Type: tea.KeyTab})
	if iv.cacheValid {
		t.Error("cache should be invalidated after tab switch")
	}
}

func TestInsightsView_DetailTabShowsHeatmap(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsView(t)
	iv.SetHeight(30)
	iv.activeTab = 1

	output := iv.View()
	if !strings.Contains(output, "ACTIVITY HEATMAP") {
		t.Errorf("Detail tab should show heatmap, got:\n%s", output)
	}
}

func TestInsightsView_DetailTabViewportScrollKeys(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetHeight(24)
	iv.activeTab = 1

	// These should not panic or return errors — viewport handles scroll keys
	for _, key := range []string{"j", "k"} {
		iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	}
	iv.Update(tea.KeyMsg{Type: tea.KeyUp})
	iv.Update(tea.KeyMsg{Type: tea.KeyDown})
}

func TestInsightsView_OverviewTabStillShowsPanels(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetHeight(24)
	iv.activeTab = 0

	output := iv.View()
	expectedSections := []string{
		"YOUR INSIGHTS DASHBOARD",
		"COMPLETION TRENDS",
		"STREAKS",
		"MOOD & PRODUCTIVITY",
		"DOOR PICKS",
	}
	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Overview tab missing section %q", section)
		}
	}
}

// ============================================================
// Story 40.5 — Activity Heatmap Tests
// ============================================================

func TestHeatmapLevel(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  int
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"two", 2, 1},
		{"three", 3, 2},
		{"four", 4, 2},
		{"five", 5, 3},
		{"six", 6, 3},
		{"seven", 7, 4},
		{"ten", 10, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := heatmapLevel(tt.count)
			if got != tt.want {
				t.Errorf("heatmapLevel(%d) = %d, want %d", tt.count, got, tt.want)
			}
		})
	}
}

func TestHeatmapGrid_EmptyData(t *testing.T) {
	grid, weeks := heatmapGrid(map[string]int{})
	if weeks != 0 {
		t.Errorf("empty data should return 0 weeks, got %d", weeks)
	}
	for d := 0; d < 7; d++ {
		if len(grid[d]) != 0 {
			t.Errorf("empty data grid[%d] should be empty", d)
		}
	}
}

func TestHeatmapGrid_SingleWeek(t *testing.T) {
	daily := map[string]int{
		"2026-03-02": 3, // Monday
		"2026-03-04": 5, // Wednesday
		"2026-03-06": 1, // Friday
	}
	grid, weeks := heatmapGrid(daily)
	if weeks != 1 {
		t.Errorf("single week data should return 1 week, got %d", weeks)
	}
	// Monday (index 0) = 3
	if grid[0][0] != 3 {
		t.Errorf("Monday should be 3, got %d", grid[0][0])
	}
	// Wednesday (index 2) = 5
	if grid[2][0] != 5 {
		t.Errorf("Wednesday should be 5, got %d", grid[2][0])
	}
	// Friday (index 4) = 1
	if grid[4][0] != 1 {
		t.Errorf("Friday should be 1, got %d", grid[4][0])
	}
	// Tuesday (index 1) = 0 (no data)
	if grid[1][0] != 0 {
		t.Errorf("Tuesday should be 0, got %d", grid[1][0])
	}
}

func TestHeatmapGrid_MultipleWeeks(t *testing.T) {
	daily := map[string]int{
		"2026-02-23": 2, // Monday week 1
		"2026-03-02": 4, // Monday week 2
		"2026-03-09": 6, // Monday week 3
	}
	grid, weeks := heatmapGrid(daily)
	if weeks != 3 {
		t.Errorf("three weeks should return 3, got %d", weeks)
	}
	// Monday values across weeks
	if grid[0][0] != 2 {
		t.Errorf("week 1 Monday should be 2, got %d", grid[0][0])
	}
	if grid[0][1] != 4 {
		t.Errorf("week 2 Monday should be 4, got %d", grid[0][1])
	}
	if grid[0][2] != 6 {
		t.Errorf("week 3 Monday should be 6, got %d", grid[0][2])
	}
}

func TestHeatmapGrid_CapsAt8Weeks(t *testing.T) {
	daily := make(map[string]int)
	// Create 12 weeks of data
	start := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC) // Monday
	for w := 0; w < 12; w++ {
		day := start.AddDate(0, 0, w*7)
		daily[day.Format("2006-01-02")] = w + 1
	}
	_, weeks := heatmapGrid(daily)
	if weeks != 8 {
		t.Errorf("should cap at 8 weeks, got %d", weeks)
	}
}

func TestMondayOf(t *testing.T) {
	tests := []struct {
		name string
		date string
		want string
	}{
		{"monday stays", "2026-03-02", "2026-03-02"},
		{"tuesday", "2026-03-03", "2026-03-02"},
		{"wednesday", "2026-03-04", "2026-03-02"},
		{"sunday", "2026-03-08", "2026-03-02"},
		{"next monday", "2026-03-09", "2026-03-09"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := time.Parse("2006-01-02", tt.date)
			got := mondayOf(d).Format("2006-01-02")
			if got != tt.want {
				t.Errorf("mondayOf(%s) = %s, want %s", tt.date, got, tt.want)
			}
		})
	}
}

func TestRenderHeatmap_NarrowTerminalHidesHeatmap(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsView(t)
	iv.SetWidth(50) // compact mode
	var s strings.Builder
	iv.renderHeatmap(&s)
	output := s.String()

	if !strings.Contains(output, "Widen terminal") {
		t.Errorf("compact mode should show widen message, got: %q", output)
	}
	if strings.Contains(output, "ACTIVITY HEATMAP") {
		t.Errorf("compact mode should NOT show heatmap grid")
	}
}

func TestRenderHeatmap_ShowsGridAndLegend(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsView(t)
	iv.SetWidth(80)
	var s strings.Builder
	iv.renderHeatmap(&s)
	output := s.String()

	// Should have title
	if !strings.Contains(output, "ACTIVITY HEATMAP") {
		t.Errorf("heatmap should contain title, got:\n%s", output)
	}

	// Should have day labels
	for _, day := range []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"} {
		if !strings.Contains(output, day) {
			t.Errorf("heatmap should contain day label %q", day)
		}
	}

	// Should have week column headers
	if !strings.Contains(output, "W1") {
		t.Errorf("heatmap should contain week header W1")
	}

	// Should have legend
	if !strings.Contains(output, "Less") || !strings.Contains(output, "More") {
		t.Errorf("heatmap should contain legend with Less/More")
	}
}

func setupInsightsViewWithHeatmapData(t *testing.T) *InsightsView {
	t.Helper()
	dir := t.TempDir()
	now := time.Date(2026, 3, 9, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	// Create sessions spanning 3 weeks
	var sessions []core.SessionMetrics
	for i := 0; i < 21; i++ {
		day := now.AddDate(0, 0, -i)
		completions := (i % 5) + 1
		sessions = append(sessions, makeInsightsTestSession(day, completions, []string{"Focused"}, []int{0}))
	}

	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := core.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	cc := core.NewCompletionCounterWithNow(frozen)
	iv := NewInsightsView(pa, cc, nil, nil)
	iv.SetWidth(80)
	iv.SetHeight(30)
	return iv
}

func TestRenderHeatmap_WithMultiWeekData(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsViewWithHeatmapData(t)
	var s strings.Builder
	iv.renderHeatmap(&s)
	output := s.String()

	// Should show heatmap with shade characters or dots
	hasShade := false
	for _, ch := range heatmapChars {
		if ch != ' ' && strings.ContainsRune(output, ch) {
			hasShade = true
			break
		}
	}
	if !hasShade && !strings.Contains(output, "·") {
		t.Errorf("heatmap should contain shade characters or dots, got:\n%s", output)
	}
}

func TestInsightsView_DetailTabHeatmapIntegration(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	iv := setupInsightsViewWithHeatmapData(t)
	iv.activeTab = 1

	output := iv.View()

	// Should contain heatmap content
	if !strings.Contains(output, "ACTIVITY HEATMAP") {
		t.Errorf("Detail tab should show heatmap, got:\n%s", output)
	}
	// Should still have tab indicator
	if !strings.Contains(output, "[Detail]") {
		t.Errorf("Detail tab should show tab indicator")
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

// --- Story 40.9: Theme-matched stats color palettes ---

func TestAllThemesProvideStatsColors(t *testing.T) {
	allThemes := []*themes.DoorTheme{
		themes.NewClassicTheme(),
		themes.NewModernTheme(),
		themes.NewSciFiTheme(),
		themes.NewShojiTheme(),
	}

	for _, th := range allThemes {
		t.Run(th.Name, func(t *testing.T) {
			if th.Colors.StatsAccent == "" {
				t.Errorf("theme %q has empty StatsAccent", th.Name)
			}
			if th.Colors.StatsGradientStart == "" {
				t.Errorf("theme %q has empty StatsGradientStart", th.Name)
			}
			if th.Colors.StatsGradientEnd == "" {
				t.Errorf("theme %q has empty StatsGradientEnd", th.Name)
			}
		})
	}
}

func TestThemeStatsColorsAreDistinct(t *testing.T) {
	allThemes := []*themes.DoorTheme{
		themes.NewClassicTheme(),
		themes.NewModernTheme(),
		themes.NewSciFiTheme(),
		themes.NewShojiTheme(),
	}

	seen := make(map[string]string) // accent → theme name
	for _, th := range allThemes {
		if prev, ok := seen[th.Colors.StatsAccent]; ok {
			t.Errorf("themes %q and %q share StatsAccent %q", prev, th.Name, th.Colors.StatsAccent)
		}
		seen[th.Colors.StatsAccent] = th.Name
	}
}

func setupInsightsViewWithTheme(t *testing.T, theme *themes.DoorTheme) *InsightsView {
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

	iv := NewInsightsView(pa, cc, theme, nil)
	iv.SetWidth(80)
	return iv
}

func TestInsightsView_NilThemeFallsBackToIndependentPalette(t *testing.T) {
	iv := setupInsightsViewWithTheme(t, nil)

	// With nil theme, statsAccentColor should return fallback gold.
	if got := iv.statsAccentColor(); got != "#FCD34D" {
		t.Errorf("nil theme statsAccentColor() = %q, want #FCD34D", got)
	}

	start, end := iv.statsGradientColors()
	if start != sparkColorStart.Dark {
		t.Errorf("nil theme gradient start = %q, want %q", start, sparkColorStart.Dark)
	}
	if end != sparkColorEnd.Dark {
		t.Errorf("nil theme gradient end = %q, want %q", end, sparkColorEnd.Dark)
	}
}

func TestInsightsView_ThemeColorsApplied(t *testing.T) {
	allThemes := []*themes.DoorTheme{
		themes.NewClassicTheme(),
		themes.NewModernTheme(),
		themes.NewSciFiTheme(),
		themes.NewShojiTheme(),
	}

	for _, th := range allThemes {
		t.Run(th.Name, func(t *testing.T) {
			iv := setupInsightsViewWithTheme(t, th)

			if got := iv.statsAccentColor(); got != th.Colors.StatsAccent {
				t.Errorf("theme %q statsAccentColor() = %q, want %q", th.Name, got, th.Colors.StatsAccent)
			}

			start, end := iv.statsGradientColors()
			if start != th.Colors.StatsGradientStart {
				t.Errorf("theme %q gradient start = %q, want %q", th.Name, start, th.Colors.StatsGradientStart)
			}
			if end != th.Colors.StatsGradientEnd {
				t.Errorf("theme %q gradient end = %q, want %q", th.Name, end, th.Colors.StatsGradientEnd)
			}
		})
	}
}

func TestInsightsView_ThemedViewRendersDifferently(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	nilIV := setupInsightsViewWithTheme(t, nil)
	nilOutput := nilIV.View()

	classicIV := setupInsightsViewWithTheme(t, themes.NewClassicTheme())
	classicOutput := classicIV.View()

	scifiIV := setupInsightsViewWithTheme(t, themes.NewSciFiTheme())
	scifiOutput := scifiIV.View()

	// With TrueColor, themed views should differ from nil-theme (different ANSI codes).
	if nilOutput == classicOutput {
		t.Error("classic-themed output should differ from nil-themed output")
	}
	if classicOutput == scifiOutput {
		t.Error("classic-themed output should differ from scifi-themed output")
	}
}

func TestStyledSparklineCharsWithGradient(t *testing.T) {
	t.Run("custom gradient produces output", func(t *testing.T) {
		chars := styledSparklineCharsWithGradient([]int{1, 5, 10}, "#22C55E", "#06B6D4")
		if len(chars) != 3 {
			t.Fatalf("expected 3 chars, got %d", len(chars))
		}
		for i, ch := range chars {
			if ch == "" {
				t.Errorf("char %d is empty", i)
			}
		}
	})

	t.Run("nil input returns nil", func(t *testing.T) {
		chars := styledSparklineCharsWithGradient(nil, "#000000", "#FFFFFF")
		if chars != nil {
			t.Errorf("expected nil, got %v", chars)
		}
	})
}

// --- Story 40.7: Animated Counter Reveals ---

func TestInsightsView_AnimationProgressIncrements(t *testing.T) {
	iv := setupInsightsView(t)

	cmd := iv.StartAnimation()
	if cmd == nil {
		t.Fatal("StartAnimation() should return a tick command")
		return
	}
	if !iv.animating {
		t.Fatal("expected animating=true after StartAnimation()")
	}
	if iv.animationProgress != 0.0 {
		t.Fatalf("expected progress=0.0, got %f", iv.animationProgress)
	}

	// Send one tick — progress should advance by animationStep
	cmd = iv.Update(StatsAnimationTickMsg{})
	if cmd == nil {
		t.Fatal("expected tick command while animating")
		return
	}
	expectedProgress := animationStep
	if iv.animationProgress < expectedProgress-0.001 || iv.animationProgress > expectedProgress+0.001 {
		t.Fatalf("expected progress ~%f, got %f", expectedProgress, iv.animationProgress)
	}
}

func TestInsightsView_AnimationStopsAfterCompletion(t *testing.T) {
	iv := setupInsightsView(t)
	iv.StartAnimation()

	// Fast-forward to near completion
	iv.animationProgress = 1.0 - animationStep/2

	cmd := iv.Update(StatsAnimationTickMsg{})
	if cmd != nil {
		t.Fatal("expected nil command after animation completes")
	}
	if iv.animating {
		t.Fatal("expected animating=false after completion")
	}
	if iv.animationProgress != 1.0 {
		t.Fatalf("expected progress=1.0 at completion, got %f", iv.animationProgress)
	}
	if !iv.hasAnimatedThisSession {
		t.Fatal("expected hasAnimatedThisSession=true after completion")
	}
}

func TestInsightsView_ReentrySkipsAnimation(t *testing.T) {
	iv := setupInsightsView(t)

	// Simulate completed animation
	iv.hasAnimatedThisSession = true

	cmd := iv.StartAnimation()
	if cmd != nil {
		t.Fatal("StartAnimation() should return nil for re-entry")
	}
	if iv.animating {
		t.Fatal("expected animating=false for re-entry")
	}
}

func TestInsightsView_EscDuringAnimationStopsCleanly(t *testing.T) {
	iv := setupInsightsView(t)
	iv.StartAnimation()

	// Advance a few ticks
	iv.Update(StatsAnimationTickMsg{})
	iv.Update(StatsAnimationTickMsg{})
	if !iv.animating {
		t.Fatal("expected still animating")
	}

	// Press Esc
	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected ReturnToDoorsMsg command")
		return
	}
	if iv.animating {
		t.Fatal("expected animating=false after Esc")
	}

	// Verify the command produces ReturnToDoorsMsg
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Fatalf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestInsightsView_AnimatedValueInterpolation(t *testing.T) {
	iv := setupInsightsView(t)

	tests := []struct {
		name       string
		animating  bool
		progress   float64
		finalValue int
		want       int
	}{
		{"not animating returns final", false, 0.5, 100, 100},
		{"progress 0 returns 0", true, 0.0, 100, 0},
		{"progress 0.5 returns half", true, 0.5, 100, 50},
		{"progress 1.0 returns final", true, 1.0, 100, 100},
		{"progress 0.3 rounds down", true, 0.3, 10, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iv.animating = tt.animating
			iv.animationProgress = tt.progress
			got := iv.animatedValue(tt.finalValue)
			if got != tt.want {
				t.Errorf("animatedValue(%d) = %d, want %d", tt.finalValue, got, tt.want)
			}
		})
	}
}

func TestInsightsView_AnimatedHeroNumber(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	iv := setupInsightsView(t)

	// Start animation at progress 0
	iv.StartAnimation()
	output := iv.View()
	if !strings.Contains(output, "0 tasks completed") {
		t.Errorf("expected '0 tasks completed' at progress 0, got:\n%s", output)
	}

	// Complete animation
	iv.animating = false
	iv.animationProgress = 1.0
	iv.hasAnimatedThisSession = true
	iv.invalidateCache()
	output = iv.View()
	total := iv.analyzer.GetTotalCompleted()
	expected := fmt.Sprintf("%d tasks completed", total)
	if !strings.Contains(output, expected) {
		t.Errorf("expected '%s' after animation, got:\n%s", expected, output)
	}
}

func TestInsightsView_CacheDisabledDuringAnimation(t *testing.T) {
	iv := setupInsightsView(t)
	iv.StartAnimation()

	// Render once
	first := iv.View()

	// Advance progress — should re-render even if cacheValid
	iv.animationProgress = 0.5
	iv.cacheValid = true // artificially set cache valid
	second := iv.View()

	// During animation, cache should be bypassed so we get different output
	// (hero number changes from 0 to ~half)
	if first == second {
		t.Error("expected different output after progress change during animation")
	}
}

func TestInsightsView_TickAfterAnimationComplete(t *testing.T) {
	iv := setupInsightsView(t)
	// Not animating — lingering tick should be ignored
	iv.hasAnimatedThisSession = true
	iv.animating = false

	cmd := iv.Update(StatsAnimationTickMsg{})
	if cmd != nil {
		t.Fatal("expected nil command for tick when not animating")
	}
}

// --- Story 40.10: Milestone Celebrations ---

func TestInsightsView_MilestoneBannerShows(t *testing.T) {
	dir := t.TempDir()
	mc := core.NewMilestoneChecker(dir)
	if err := mc.Load(); err != nil {
		t.Fatal(err)
	}

	pa := core.NewPatternAnalyzer()
	cc := core.NewCompletionCounter()
	iv := NewInsightsView(pa, cc, nil, mc)
	iv.SetWidth(80)

	// Trigger milestone check with sessionCount=1 (first session)
	cmd := iv.CheckAndShowMilestone(0, 0, 1)
	if cmd == nil {
		t.Fatal("expected auto-dismiss command for milestone banner")
		return
	}
	if !iv.bannerActive {
		t.Error("banner should be active after milestone check")
	}

	output := iv.View()
	if !strings.Contains(output, "Welcome! Your journey starts here.") {
		t.Error("view should contain milestone message")
	}
}

func TestInsightsView_MilestoneBannerDismissOnKey(t *testing.T) {
	dir := t.TempDir()
	mc := core.NewMilestoneChecker(dir)
	if err := mc.Load(); err != nil {
		t.Fatal(err)
	}

	pa := core.NewPatternAnalyzer()
	cc := core.NewCompletionCounter()
	iv := NewInsightsView(pa, cc, nil, mc)
	iv.SetWidth(80)
	iv.CheckAndShowMilestone(0, 0, 1)

	// Any key should dismiss
	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if iv.bannerActive {
		t.Error("banner should be dismissed after keypress")
	}
	// Key should NOT be passed through (no ReturnToDoorsMsg etc.)
	if cmd != nil {
		t.Error("keypress during banner should not produce a command")
	}
}

func TestInsightsView_MilestoneBannerDismissOnTimer(t *testing.T) {
	dir := t.TempDir()
	mc := core.NewMilestoneChecker(dir)
	if err := mc.Load(); err != nil {
		t.Fatal(err)
	}

	pa := core.NewPatternAnalyzer()
	cc := core.NewCompletionCounter()
	iv := NewInsightsView(pa, cc, nil, mc)
	iv.SetWidth(80)
	iv.CheckAndShowMilestone(0, 0, 1)

	// Simulate timer firing
	iv.Update(MilestoneDismissMsg{})
	if iv.bannerActive {
		t.Error("banner should be dismissed after MilestoneDismissMsg")
	}
}

func TestInsightsView_MilestoneShownOnlyOnce(t *testing.T) {
	dir := t.TempDir()
	mc := core.NewMilestoneChecker(dir)
	if err := mc.Load(); err != nil {
		t.Fatal(err)
	}

	pa := core.NewPatternAnalyzer()
	cc := core.NewCompletionCounter()

	// First entry: should trigger
	iv1 := NewInsightsView(pa, cc, nil, mc)
	iv1.SetWidth(80)
	cmd1 := iv1.CheckAndShowMilestone(0, 0, 1)
	if cmd1 == nil {
		t.Fatal("first check should trigger milestone")
		return
	}

	// Second entry with same checker: should NOT trigger (already marked shown)
	iv2 := NewInsightsView(pa, cc, nil, mc)
	iv2.SetWidth(80)
	cmd2 := iv2.CheckAndShowMilestone(0, 0, 1)
	if cmd2 != nil {
		t.Error("second check should not trigger same milestone again")
	}
}

func TestInsightsView_NilMilestoneChecker(t *testing.T) {
	pa := core.NewPatternAnalyzer()
	cc := core.NewCompletionCounter()
	iv := NewInsightsView(pa, cc, nil, nil)
	iv.SetWidth(80)

	cmd := iv.CheckAndShowMilestone(100, 10, 5)
	if cmd != nil {
		t.Error("nil milestone checker should return nil command")
	}
	if iv.bannerActive {
		t.Error("banner should not be active with nil checker")
	}
}
