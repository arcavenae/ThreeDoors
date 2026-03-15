package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

func fixedNow() time.Time {
	return time.Date(2026, time.March, 15, 14, 30, 0, 0, time.Local)
}

func sampleRecords() []core.CompletionRecord {
	loc := time.Local
	return []core.CompletionRecord{
		{Title: "Write unit tests", CompletedAt: time.Date(2026, 3, 15, 10, 30, 0, 0, loc)},
		{Title: "Review PR #42", CompletedAt: time.Date(2026, 3, 15, 9, 15, 0, 0, loc)},
		{Title: "Deploy staging", CompletedAt: time.Date(2026, 3, 14, 17, 0, 0, 0, loc), Source: "Jira"},
		{Title: "Fix login bug", CompletedAt: time.Date(2026, 3, 14, 14, 45, 0, 0, loc)},
		{Title: "Update docs", CompletedAt: time.Date(2026, 3, 12, 11, 0, 0, 0, loc), Source: "Linear"},
		{Title: "Sprint planning", CompletedAt: time.Date(2026, 3, 10, 9, 0, 0, 0, loc)},
		{Title: "Database migration", CompletedAt: time.Date(2026, 3, 5, 16, 30, 0, 0, loc)},
	}
}

func TestGolden_HistoryView(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	hv := NewHistoryView(sampleRecords(), fixedNow)
	hv.SetWidth(80)
	hv.SetHeight(30)
	out := hv.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_HistoryViewEmpty(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	hv := NewHistoryView(nil, fixedNow)
	hv.SetWidth(80)
	hv.SetHeight(30)
	out := hv.View()
	golden.RequireEqual(t, []byte(out))
}

func TestHistoryView_Update_Quit(t *testing.T) {
	t.Parallel()
	hv := NewHistoryView(sampleRecords(), fixedNow)

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"q key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"esc key", tea.KeyMsg{Type: tea.KeyEsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := hv.Update(tt.key)
			if cmd == nil {
				t.Fatal("expected non-nil cmd for quit key")
			}
			msg := cmd()
			if _, ok := msg.(ReturnToDoorsMsg); !ok {
				t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
			}
		})
	}
}

func TestFormatDateHeader(t *testing.T) {
	t.Parallel()
	loc := time.Local
	today := time.Date(2026, 3, 15, 0, 0, 0, 0, loc) // Sunday

	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{"today", time.Date(2026, 3, 15, 10, 0, 0, 0, loc), "Today \u2014 March 15"},
		{"yesterday", time.Date(2026, 3, 14, 10, 0, 0, 0, loc), "Yesterday \u2014 March 14"},
		{"within week", time.Date(2026, 3, 12, 10, 0, 0, 0, loc), "Thursday \u2014 March 12"},
		{"older", time.Date(2026, 3, 5, 10, 0, 0, 0, loc), "March 5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatDateHeader(tt.date, today)
			if got != tt.expected {
				t.Errorf("formatDateHeader() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHistoryView_SetDimensions(t *testing.T) {
	t.Parallel()
	hv := NewHistoryView(sampleRecords(), fixedNow)

	hv.SetWidth(120)
	if hv.width != 120 {
		t.Errorf("width = %d, want 120", hv.width)
	}
	if hv.viewport.Width != 120 {
		t.Errorf("viewport.Width = %d, want 120", hv.viewport.Width)
	}

	hv.SetHeight(40)
	if hv.height != 40 {
		t.Errorf("height = %d, want 40", hv.height)
	}
	expectedVPHeight := 40 - historyHeaderHeight - historyFooterHeight
	if hv.viewport.Height != expectedVPHeight {
		t.Errorf("viewport.Height = %d, want %d", hv.viewport.Height, expectedVPHeight)
	}
}

func TestHistoryView_SetHeight_MinimumOne(t *testing.T) {
	t.Parallel()
	hv := NewHistoryView(nil, fixedNow)
	hv.SetHeight(1)
	if hv.viewport.Height != 1 {
		t.Errorf("viewport.Height = %d, want 1 (minimum)", hv.viewport.Height)
	}
}

func TestHistoryView_SourceBadge(t *testing.T) {
	t.Parallel()
	lipgloss.SetColorProfile(termenv.Ascii)
	defer lipgloss.SetColorProfile(termenv.TrueColor)

	records := []core.CompletionRecord{
		{Title: "Task with source", CompletedAt: time.Date(2026, 3, 15, 10, 0, 0, 0, time.Local), Source: "Jira"},
	}
	hv := NewHistoryView(records, fixedNow)
	content := hv.renderContent()
	if content == "" {
		t.Fatal("expected non-empty content")
	}
	if !strings.Contains(content, "Jira") {
		t.Error("expected source badge [Jira] in rendered content")
	}
}

func TestHistoryView_EmptyRenderContent(t *testing.T) {
	t.Parallel()
	hv := NewHistoryView(nil, fixedNow)
	content := hv.renderContent()
	if content != "" {
		t.Errorf("expected empty content for no records, got %q", content)
	}
}

func TestHistoryView_NilNowFunc(t *testing.T) {
	t.Parallel()
	hv := NewHistoryView(sampleRecords(), nil)
	if hv.nowFunc == nil {
		t.Fatal("expected nowFunc to be set to default")
	}
}
