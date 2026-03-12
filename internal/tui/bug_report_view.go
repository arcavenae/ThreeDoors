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
	bugReportSubmitting
	bugReportSuccess
	bugReportError
)

// ShowBugReportMsg is sent to open the bug report view.
type ShowBugReportMsg struct{}

// BugReportView manages the bug report input and preview workflow.
type BugReportView struct {
	state        bugReportState
	textArea     textarea.Model
	viewport     viewport.Model
	report       *BugReport
	envInfo      EnvironmentInfo
	breadcrumbs  string
	width        int
	height       int
	hasToken     bool
	successMsg   string
	errorMsg     string
	failedMethod string
	submitMethod string
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
		hasToken:    hasGitHubToken(),
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
	// Handle submission result messages in any state.
	switch msg := msg.(type) {
	case BugReportSubmittedMsg:
		v.state = bugReportSuccess
		v.submitMethod = msg.Method
		switch msg.Method {
		case "browser":
			v.successMsg = "Opening GitHub in your browser..."
		case "api":
			v.successMsg = fmt.Sprintf("Issue created: %s", msg.Details)
		case "file":
			v.successMsg = fmt.Sprintf("Report saved to %s", msg.Details)
		case "clipboard":
			v.successMsg = "URL copied to clipboard"
		}
		return nil
	case BugReportErrorMsg:
		v.state = bugReportError
		v.failedMethod = msg.Method
		v.errorMsg = fmt.Sprintf("Could not submit via %s: %s", msg.Method, msg.Err)
		return nil
	}

	switch v.state {
	case bugReportInput:
		return v.updateInput(msg)
	case bugReportPreview:
		return v.updatePreview(msg)
	case bugReportSuccess:
		return v.updateSuccess(msg)
	case bugReportError:
		return v.updateError(msg)
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
		case "b":
			v.state = bugReportSubmitting
			return openBrowserCmd(v.report)
		case "s":
			if v.hasToken {
				v.state = bugReportSubmitting
				client := newGitHubClientForBugReport()
				return submitViaAPICmd(v.report, client)
			}
		case "f":
			v.state = bugReportSubmitting
			return saveBugReportCmd(v.report)
		}
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

func (v *BugReportView) updateSuccess(msg tea.Msg) tea.Cmd {
	if _, ok := msg.(tea.KeyMsg); ok {
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	}
	return nil
}

func (v *BugReportView) updateError(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "b":
			// Offer browser as fallback from API failure.
			if v.failedMethod == "api" {
				v.state = bugReportSubmitting
				return openBrowserCmd(v.report)
			}
			// Offer clipboard as fallback from browser failure.
			if v.failedMethod == "browser" {
				issueURL := BuildIssueURL(v.report)
				return copyToClipboardCmd(issueURL)
			}
		case "f":
			// File save is always available as a fallback.
			v.state = bugReportSubmitting
			return saveBugReportCmd(v.report)
		case "esc", "q":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}
	return nil
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
	case bugReportSubmitting:
		return v.viewSubmitting()
	case bugReportSuccess:
		return v.viewSuccess()
	case bugReportError:
		return v.viewError()
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

	hints := "[b] Open in browser  [f] Save to file"
	if v.hasToken {
		hints = "[b] Open in browser  [s] Submit via GitHub  [f] Save to file"
	}
	hints += "  |  j/k scroll  [Esc] Back to edit  [q] Cancel"
	s.WriteString(helpStyle.Render(hints))

	return s.String()
}

func (v *BugReportView) viewSubmitting() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", syncLogHeaderStyle.Render("Bug Report"))
	s.WriteString(helpStyle.Render("Submitting your report..."))

	return s.String()
}

func (v *BugReportView) viewSuccess() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", syncLogHeaderStyle.Render("Bug Report"))
	fmt.Fprintf(&s, "%s\n\n", v.successMsg)
	s.WriteString(helpStyle.Render("Press any key to return"))

	return s.String()
}

func (v *BugReportView) viewError() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", syncLogHeaderStyle.Render("Bug Report"))
	fmt.Fprintf(&s, "%s\n\n", v.errorMsg)

	switch v.failedMethod {
	case "api":
		s.WriteString(helpStyle.Render("[b] Try browser instead  [f] Save to file  [Esc] Cancel"))
	case "browser":
		s.WriteString(helpStyle.Render("[b] Copy URL to clipboard  [f] Save to file  [Esc] Cancel"))
	default:
		s.WriteString(helpStyle.Render("[f] Save to file  [Esc] Cancel"))
	}

	return s.String()
}
