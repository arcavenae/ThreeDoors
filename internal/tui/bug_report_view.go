package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// bugReportState tracks the sub-state within the bug report view.
type bugReportState int

const (
	bugReportInput bugReportState = iota
	bugReportPreview
)

// ShowBugReportMsg is sent to open the bug report view.
type ShowBugReportMsg struct{}

// BugReportView manages the bug report input and preview workflow.
type BugReportView struct {
	state       bugReportState
	textArea    textarea.Model
	viewport    viewport.Model
	report      *BugReport
	envInfo     EnvironmentInfo
	breadcrumbs string
	width       int
	height      int
}

// NewBugReportView creates a new bug report view with the given environment context.
func NewBugReportView(env EnvironmentInfo, breadcrumbs string) *BugReportView {
	ta := textarea.New()
	ta.Placeholder = "Describe the bug you encountered..."
	ta.CharLimit = 2000
	ta.SetWidth(60)
	ta.SetHeight(6)
	ta.Focus()

	return &BugReportView{
		state:       bugReportInput,
		textArea:    ta,
		envInfo:     env,
		breadcrumbs: breadcrumbs,
		width:       80,
		height:      24,
	}
}

// SetWidth sets the terminal width.
func (v *BugReportView) SetWidth(w int) {
	v.width = w
	taWidth := w - 4
	if taWidth < 20 {
		taWidth = 20
	}
	v.textArea.SetWidth(taWidth)
}

// SetHeight sets the terminal height.
func (v *BugReportView) SetHeight(h int) {
	v.height = h
}

// Update handles key presses for the bug report view.
func (v *BugReportView) Update(msg tea.Msg) tea.Cmd {
	switch v.state {
	case bugReportInput:
		return v.updateInput(msg)
	case bugReportPreview:
		return v.updatePreview(msg)
	}
	return nil
}

func (v *BugReportView) updateInput(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		case tea.KeyEnter:
			desc := strings.TrimSpace(v.textArea.Value())
			if desc == "" {
				return nil
			}
			v.report = &BugReport{
				Description: desc,
				Environment: v.envInfo,
				Breadcrumbs: v.breadcrumbs,
				Timestamp:   time.Now().UTC(),
			}
			v.state = bugReportPreview
			v.initPreviewViewport()
			return nil
		}
	}

	var cmd tea.Cmd
	v.textArea, cmd = v.textArea.Update(msg)
	return cmd
}

func (v *BugReportView) updatePreview(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			v.state = bugReportInput
			v.textArea.Focus()
			return nil
		case "q":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

func (v *BugReportView) initPreviewViewport() {
	vpHeight := v.height - 6
	if vpHeight < 4 {
		vpHeight = 4
	}
	v.viewport = NewScrollableView(v.width, vpHeight)
	if v.report != nil {
		v.viewport.SetContent(v.report.FormatMarkdown())
	}
}

// View renders the bug report view.
func (v *BugReportView) View() string {
	switch v.state {
	case bugReportPreview:
		return v.viewPreview()
	default:
		return v.viewInput()
	}
}

func (v *BugReportView) viewInput() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", syncLogHeaderStyle.Render("Bug Report"))

	s.WriteString(v.textArea.View())
	s.WriteString("\n\n")

	// Environment summary
	fmt.Fprintf(&s, "%s\n", helpStyle.Render("Environment data that will be included:"))
	fmt.Fprintf(&s, "%s\n", helpStyle.Render(fmt.Sprintf("  Version: %s (%s)", v.envInfo.Version, v.envInfo.Commit)))
	fmt.Fprintf(&s, "%s\n", helpStyle.Render(fmt.Sprintf("  OS: %s/%s  Terminal: %dx%d", v.envInfo.OS, v.envInfo.Arch, v.envInfo.TerminalWidth, v.envInfo.TerminalHeight)))
	fmt.Fprintf(&s, "%s\n", helpStyle.Render(fmt.Sprintf("  Theme: %s  Tasks: %d  Providers: %d", v.envInfo.ThemeName, v.envInfo.TaskCount, v.envInfo.ProviderCount)))
	s.WriteString("\n")

	// Privacy disclaimer
	fmt.Fprintf(&s, "%s\n\n", helpStyle.Render("No task names, content, or personal data will be included."))

	// Keybinding hints
	s.WriteString(helpStyle.Render("[Enter] Preview  [Esc] Cancel"))

	return s.String()
}

func (v *BugReportView) viewPreview() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", syncLogHeaderStyle.Render("Bug Report Preview"))

	s.WriteString(v.viewport.View())

	fmt.Fprintf(&s, "\n\n  %3.f%%\n", v.viewport.ScrollPercent()*100) //nolint:mnd
	s.WriteString(helpStyle.Render("j/k scroll | [Esc] Back to edit | [q] Cancel"))

	return s.String()
}
