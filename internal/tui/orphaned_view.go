package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OrphanedView displays orphaned tasks and lets the user keep or delete them.
type OrphanedView struct {
	pool          *core.TaskPool
	width         int
	height        int
	selectedIndex int
	tasks         []*core.Task // cached snapshot, refreshed on actions
}

// NewOrphanedView creates a new OrphanedView.
func NewOrphanedView(pool *core.TaskPool) *OrphanedView {
	ov := &OrphanedView{pool: pool}
	ov.refreshTasks()
	return ov
}

// refreshTasks reloads the orphaned task list from the pool, sorted by OrphanedAt descending.
func (ov *OrphanedView) refreshTasks() {
	ov.tasks = ov.pool.GetOrphanedTasks()
	sort.Slice(ov.tasks, func(i, j int) bool {
		ti := ov.tasks[i].OrphanedAt
		tj := ov.tasks[j].OrphanedAt
		if ti == nil && tj == nil {
			return false
		}
		if ti == nil {
			return false
		}
		if tj == nil {
			return true
		}
		return ti.After(*tj)
	})
	if ov.selectedIndex >= len(ov.tasks) {
		ov.selectedIndex = max(0, len(ov.tasks)-1)
	}
}

// SetWidth sets the terminal width for rendering.
func (ov *OrphanedView) SetWidth(w int) {
	ov.width = w
}

// SetHeight sets the terminal height for rendering.
func (ov *OrphanedView) SetHeight(h int) {
	ov.height = h
}

// Update handles key events for the orphaned view.
func (ov *OrphanedView) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.Type {
	case tea.KeyEsc:
		return func() tea.Msg { return ReturnToDoorsMsg{} }

	case tea.KeyEnter:
		// Enter = keep locally
		if len(ov.tasks) == 0 {
			return nil
		}
		task := ov.tasks[ov.selectedIndex]
		return func() tea.Msg {
			return OrphanedTaskActionMsg{TaskID: task.ID, Action: "keep"}
		}

	case tea.KeyRunes:
		if len(keyMsg.Runes) == 0 {
			return nil
		}
		switch keyMsg.Runes[0] {
		case 'q':
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		case 'j':
			if ov.selectedIndex < len(ov.tasks)-1 {
				ov.selectedIndex++
			}
		case 'k':
			if ov.selectedIndex > 0 {
				ov.selectedIndex--
			}
		case 'K', 'e': // keep locally
			if len(ov.tasks) == 0 {
				return nil
			}
			task := ov.tasks[ov.selectedIndex]
			return func() tea.Msg {
				return OrphanedTaskActionMsg{TaskID: task.ID, Action: "keep"}
			}
		case 'd', 'x': // delete permanently
			if len(ov.tasks) == 0 {
				return nil
			}
			task := ov.tasks[ov.selectedIndex]
			return func() tea.Msg {
				return OrphanedTaskActionMsg{TaskID: task.ID, Action: "delete"}
			}
		}

	case tea.KeyUp:
		if ov.selectedIndex > 0 {
			ov.selectedIndex--
		}
	case tea.KeyDown:
		if ov.selectedIndex < len(ov.tasks)-1 {
			ov.selectedIndex++
		}
	}

	return nil
}

// View renders the orphaned tasks list.
func (ov *OrphanedView) View() string {
	width := ov.width
	if width < 20 {
		width = 20
	}

	var s strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Orphaned Tasks"))

	if len(ov.tasks) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Padding(1, 2)
		fmt.Fprintf(&s, "%s\n", emptyStyle.Render("No orphaned tasks.\nTasks deleted remotely will appear here for review."))
		fmt.Fprintf(&s, "\n")
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
		fmt.Fprintf(&s, "%s", hintStyle.Render(" esc:back"))
		return s.String()
	}

	// Clamp selectedIndex
	if ov.selectedIndex >= len(ov.tasks) {
		ov.selectedIndex = len(ov.tasks) - 1
	}

	for i, task := range ov.tasks {
		source := task.EffectiveSourceProvider()
		if source == "" {
			source = "unknown"
		}

		orphanedTime := "unknown"
		if task.OrphanedAt != nil {
			orphanedTime = formatTimeAgo(*task.OrphanedAt)
		}

		// Build the row
		indicator := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("!")
		sourceStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(source)
		timeStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(orphanedTime)

		row := fmt.Sprintf(" %s %-30s %s  %s", indicator, truncate(task.Text, 30), sourceStyled, timeStyled)

		if i == ov.selectedIndex {
			row = lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Bold(true).
				Width(width - 2).
				Render(row)
		}
		fmt.Fprintf(&s, "%s\n", row)
	}

	// Footer
	fmt.Fprintf(&s, "\n")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	fmt.Fprintf(&s, "%s", hintStyle.Render(" K/enter:keep locally  d/x:delete  j/k:navigate  esc:back"))

	return s.String()
}

// formatTimeAgo returns a human-readable "time ago" string.
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
}
