package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// historyHeaderHeight is the number of lines consumed by the title header above the viewport.
const historyHeaderHeight = 2

// historyFooterHeight is the number of lines consumed by the footer below the viewport.
const historyFooterHeight = 3

// HistoryView displays a scrollable list of completed tasks grouped by date.
type HistoryView struct {
	records  []core.CompletionRecord
	viewport viewport.Model
	width    int
	height   int
	nowFunc  func() time.Time
}

// NewHistoryView creates a new HistoryView with the given completion records (newest-first).
func NewHistoryView(records []core.CompletionRecord, nowFunc func() time.Time) *HistoryView {
	if nowFunc == nil {
		nowFunc = time.Now
	}
	hv := &HistoryView{
		records: records,
		width:   80,
		height:  24,
		nowFunc: nowFunc,
	}
	hv.initViewport()
	return hv
}

// SetWidth sets the terminal width and re-renders content.
func (hv *HistoryView) SetWidth(w int) {
	hv.width = w
	hv.viewport.Width = w
	hv.viewport.SetContent(hv.renderContent())
}

// SetHeight sets the terminal height and adjusts viewport.
func (hv *HistoryView) SetHeight(h int) {
	hv.height = h
	vpHeight := h - historyHeaderHeight - historyFooterHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	hv.viewport.Height = vpHeight
}

// initViewport creates and configures the viewport.
func (hv *HistoryView) initViewport() {
	vpHeight := hv.height - historyHeaderHeight - historyFooterHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	hv.viewport = NewScrollableView(hv.width, vpHeight)
	hv.viewport.SetContent(hv.renderContent())
}

// renderContent pre-computes all history content as a single string.
func (hv *HistoryView) renderContent() string {
	if len(hv.records) == 0 {
		return ""
	}

	now := hv.nowFunc()
	local := now.Local()
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())

	var s strings.Builder
	var currentDate string

	for _, rec := range hv.records {
		recLocal := rec.CompletedAt.In(local.Location())
		dateLabel := formatDateHeader(recLocal, today)

		if dateLabel != currentDate {
			if currentDate != "" {
				s.WriteString("\n")
			}
			fmt.Fprintf(&s, "  %s\n", syncLogHeaderStyle.Render(dateLabel))
			currentDate = dateLabel
		}

		timeStr := syncLogTimestampStyle.Render(recLocal.Format("15:04"))
		title := syncLogEntryStyle.Render(rec.Title)

		if rec.Source != "" {
			badge := syncLogTimestampStyle.Render(fmt.Sprintf("[%s]", rec.Source))
			fmt.Fprintf(&s, "    %s  %s %s\n", timeStr, title, badge)
		} else {
			fmt.Fprintf(&s, "    %s  %s\n", timeStr, title)
		}
	}

	return s.String()
}

// formatDateHeader returns a human-friendly date label.
// Uses local timezone for determining "Today", "Yesterday", and day-of-week labels.
func formatDateHeader(t time.Time, today time.Time) string {
	tDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	diff := int(today.Sub(tDate).Hours() / 24) //nolint:mnd

	switch {
	case diff == 0:
		return fmt.Sprintf("Today \u2014 %s", t.Format("January 2"))
	case diff == 1:
		return fmt.Sprintf("Yesterday \u2014 %s", t.Format("January 2"))
	case diff < 7: //nolint:mnd
		return fmt.Sprintf("%s \u2014 %s", t.Weekday().String(), t.Format("January 2"))
	default:
		return t.Format("January 2")
	}
}

// Update handles key presses for scrolling and closing.
func (hv *HistoryView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}

	var cmd tea.Cmd
	hv.viewport, cmd = hv.viewport.Update(msg)
	return cmd
}

// View renders the history view.
func (hv *HistoryView) View() string {
	var s strings.Builder

	fmt.Fprintf(&s, "  %s\n\n", syncLogHeaderStyle.Render("Completion History"))

	if len(hv.records) == 0 {
		s.WriteString(helpStyle.Render("  No completed tasks yet. Pick a door and get started!"))
		return s.String()
	}

	s.WriteString(hv.viewport.View())

	fmt.Fprintf(&s, "\n\n  %3.f%%", hv.viewport.ScrollPercent()*100) //nolint:mnd
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("  j/k scroll | PgUp/PgDn page | q/Esc return"))

	return s.String()
}
