package tui

import (
	"fmt"
	"strings"

	"github.com/arcavenae/ThreeDoors/internal/intelligence/services"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// extractStep tracks the current phase of the extraction flow.
type extractStep int

const (
	extractStepSourceSelect extractStep = iota
	extractStepFileInput
	extractStepPasteInput
	extractStepLoading
	extractStepReview
	extractStepEditing
)

// ExtractView handles the full extraction flow: source selection, loading,
// review, and task editing before import.
type ExtractView struct {
	step        extractStep
	tasks       []services.ExtractedTask
	selected    []bool
	cursorIndex int
	width       int
	errorMsg    string

	// Source selection
	source string // "file", "clipboard", "paste"

	// File input
	fileInput textinput.Model

	// Paste input
	pasteArea textarea.Model

	// Editing state
	editIndex   int
	editField   int // 0=text, 1=effort, 2=tags
	editInput   textinput.Model
	backendName string
	sourceLabel string
}

// NewExtractView creates an ExtractView starting at source selection.
func NewExtractView() *ExtractView {
	fi := textinput.New()
	fi.Placeholder = "path/to/file.txt"
	fi.CharLimit = 512

	pa := textarea.New()
	pa.Placeholder = "Paste your text here..."
	pa.CharLimit = 32768

	ei := textinput.New()
	ei.CharLimit = 512

	return &ExtractView{
		step:      extractStepSourceSelect,
		fileInput: fi,
		pasteArea: pa,
		editInput: ei,
	}
}

// SetWidth sets the terminal width for rendering.
func (ev *ExtractView) SetWidth(w int) {
	ev.width = w
	ev.pasteArea.SetWidth(w - 8)
}

// SetResult populates the view with extraction results, transitioning to review.
func (ev *ExtractView) SetResult(tasks []services.ExtractedTask, backendName string) {
	ev.step = extractStepReview
	ev.tasks = tasks
	ev.backendName = backendName
	ev.selected = make([]bool, len(tasks))
	for i := range ev.selected {
		ev.selected[i] = !tasks[i].Duplicate
	}
	ev.cursorIndex = 0
	ev.errorMsg = ""
}

// SetError sets the error state.
func (ev *ExtractView) SetError(errMsg string) {
	ev.step = extractStepReview
	ev.errorMsg = errMsg
	ev.tasks = nil
}

// SelectedCount returns the number of currently selected tasks.
func (ev *ExtractView) SelectedCount() int {
	count := 0
	for _, s := range ev.selected {
		if s {
			count++
		}
	}
	return count
}

// SelectedTasks returns the tasks that are currently selected.
func (ev *ExtractView) SelectedTasks() []services.ExtractedTask {
	var result []services.ExtractedTask
	for i, s := range ev.selected {
		if s {
			result = append(result, ev.tasks[i])
		}
	}
	return result
}

// Update handles input in the extract view.
func (ev *ExtractView) Update(msg tea.Msg) tea.Cmd {
	switch ev.step {
	case extractStepSourceSelect:
		return ev.updateSourceSelect(msg)
	case extractStepFileInput:
		return ev.updateFileInput(msg)
	case extractStepPasteInput:
		return ev.updatePasteInput(msg)
	case extractStepLoading:
		return ev.updateLoading(msg)
	case extractStepReview:
		return ev.updateReview(msg)
	case extractStepEditing:
		return ev.updateEditing(msg)
	}
	return nil
}

func (ev *ExtractView) updateSourceSelect(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f":
			ev.step = extractStepFileInput
			ev.source = "file"
			ev.fileInput.Focus()
			return textinput.Blink
		case "c":
			ev.source = "clipboard"
			ev.sourceLabel = "clipboard"
			ev.step = extractStepLoading
			return func() tea.Msg {
				return ExtractStartMsg{Source: "clipboard"}
			}
		case "p":
			ev.step = extractStepPasteInput
			ev.source = "paste"
			ev.pasteArea.Focus()
			return textarea.Blink
		case "esc", "q":
			return func() tea.Msg { return ExtractCancelMsg{} }
		}
	}
	return nil
}

func (ev *ExtractView) updateFileInput(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			path := strings.TrimSpace(ev.fileInput.Value())
			if path == "" {
				return nil
			}
			ev.sourceLabel = path
			ev.step = extractStepLoading
			return func() tea.Msg {
				return ExtractStartMsg{Source: "file", Input: path}
			}
		case "esc":
			ev.step = extractStepSourceSelect
			ev.fileInput.Reset()
			return nil
		}
	}
	var cmd tea.Cmd
	ev.fileInput, cmd = ev.fileInput.Update(msg)
	return cmd
}

func (ev *ExtractView) updatePasteInput(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			text := strings.TrimSpace(ev.pasteArea.Value())
			if text == "" {
				return nil
			}
			ev.sourceLabel = "pasted text"
			ev.step = extractStepLoading
			return func() tea.Msg {
				return ExtractStartMsg{Source: "text", Input: text}
			}
		case "esc":
			ev.step = extractStepSourceSelect
			ev.pasteArea.Reset()
			return nil
		}
	}
	var cmd tea.Cmd
	ev.pasteArea, cmd = ev.pasteArea.Update(msg)
	return cmd
}

func (ev *ExtractView) updateLoading(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return func() tea.Msg { return ExtractCancelMsg{} }
		}
	}
	return nil
}

func (ev *ExtractView) updateReview(msg tea.Msg) tea.Cmd {
	if ev.errorMsg != "" || len(ev.tasks) == 0 {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			_ = msg
			return func() tea.Msg { return ExtractCancelMsg{} }
		}
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return ev.handleReviewKey(msg)
	}
	return nil
}

func (ev *ExtractView) handleReviewKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		return func() tea.Msg { return ExtractCancelMsg{} }
	case "up", "k":
		if ev.cursorIndex > 0 {
			ev.cursorIndex--
		}
	case "down", "j":
		if ev.cursorIndex < len(ev.tasks)-1 {
			ev.cursorIndex++
		}
	case " ":
		if ev.cursorIndex >= 0 && ev.cursorIndex < len(ev.selected) {
			ev.selected[ev.cursorIndex] = !ev.selected[ev.cursorIndex]
		}
	case "a":
		allSelected := ev.SelectedCount() == len(ev.tasks)
		for i := range ev.selected {
			ev.selected[i] = !allSelected
		}
	case "n":
		for i := range ev.selected {
			ev.selected[i] = false
		}
	case "e":
		if ev.cursorIndex >= 0 && ev.cursorIndex < len(ev.tasks) {
			ev.editIndex = ev.cursorIndex
			ev.editField = 0
			ev.editInput.SetValue(ev.tasks[ev.cursorIndex].Text)
			ev.editInput.Focus()
			ev.step = extractStepEditing
			return textinput.Blink
		}
	case "enter":
		selected := ev.SelectedTasks()
		if len(selected) == 0 {
			return nil
		}
		source := ev.sourceLabel
		return func() tea.Msg {
			return ExtractImportMsg{
				Tasks:  selected,
				Source: source,
			}
		}
	}
	return nil
}

func (ev *ExtractView) updateEditing(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			ev.applyEdit()
			ev.editField++
			if ev.editField > 2 {
				ev.step = extractStepReview
				return nil
			}
			ev.loadEditField()
			return nil
		case "esc":
			ev.step = extractStepReview
			return nil
		case "tab":
			ev.applyEdit()
			ev.editField = (ev.editField + 1) % 3
			ev.loadEditField()
			return nil
		}
	}
	var cmd tea.Cmd
	ev.editInput, cmd = ev.editInput.Update(msg)
	return cmd
}

func (ev *ExtractView) loadEditField() {
	task := ev.tasks[ev.editIndex]
	switch ev.editField {
	case 0:
		ev.editInput.SetValue(task.Text)
		ev.editInput.Placeholder = "Task text"
	case 1:
		ev.editInput.SetValue(fmt.Sprintf("%d", task.Effort))
		ev.editInput.Placeholder = "Effort (1-5)"
	case 2:
		ev.editInput.SetValue(strings.Join(task.Tags, ", "))
		ev.editInput.Placeholder = "Tags (comma-separated)"
	}
	ev.editInput.CursorEnd()
}

func (ev *ExtractView) applyEdit() {
	val := strings.TrimSpace(ev.editInput.Value())
	switch ev.editField {
	case 0:
		if val != "" {
			ev.tasks[ev.editIndex].Text = val
		}
	case 1:
		var effort int
		if _, err := fmt.Sscanf(val, "%d", &effort); err == nil && effort >= 1 && effort <= 5 {
			ev.tasks[ev.editIndex].Effort = effort
		}
	case 2:
		if val == "" {
			ev.tasks[ev.editIndex].Tags = nil
		} else {
			tags := strings.Split(val, ",")
			cleaned := make([]string, 0, len(tags))
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					cleaned = append(cleaned, tag)
				}
			}
			ev.tasks[ev.editIndex].Tags = cleaned
		}
	}
}

// View renders the extract view.
func (ev *ExtractView) View() string {
	var s strings.Builder

	w := ev.width - 6
	if w < 40 {
		w = 40
	}

	s.WriteString(extractHeaderStyle.Render("EXTRACT TASKS"))
	s.WriteString("\n\n")

	switch ev.step {
	case extractStepSourceSelect:
		ev.renderSourceSelect(&s)
	case extractStepFileInput:
		ev.renderFileInput(&s)
	case extractStepPasteInput:
		ev.renderPasteInput(&s)
	case extractStepLoading:
		ev.renderLoading(&s)
	case extractStepReview:
		ev.renderReview(&s)
	case extractStepEditing:
		ev.renderEditing(&s)
	}

	return detailBorder.Width(w).Render(s.String())
}

func (ev *ExtractView) renderSourceSelect(s *strings.Builder) {
	s.WriteString("Choose a source:\n\n")
	fmt.Fprintf(s, "  %s  Read from a file\n", extractKeyStyle.Render("[f]"))
	fmt.Fprintf(s, "  %s  Read from clipboard\n", extractKeyStyle.Render("[c]"))
	fmt.Fprintf(s, "  %s  Paste text directly\n", extractKeyStyle.Render("[p]"))
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("[Esc] Cancel"))
}

func (ev *ExtractView) renderFileInput(s *strings.Builder) {
	s.WriteString("Enter file path:\n\n")
	s.WriteString(ev.fileInput.View())
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("[Enter] Submit  [Esc] Back"))
}

func (ev *ExtractView) renderPasteInput(s *strings.Builder) {
	s.WriteString("Paste your text (Ctrl+D to submit):\n\n")
	s.WriteString(ev.pasteArea.View())
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("[Ctrl+D] Submit  [Esc] Back"))
}

func (ev *ExtractView) renderLoading(s *strings.Builder) {
	s.WriteString(extractLoadingStyle.Render("Extracting tasks..."))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("[Esc] Cancel"))
}

func (ev *ExtractView) renderReview(s *strings.Builder) {
	if ev.errorMsg != "" {
		fmt.Fprintf(s, "%s\n\n", extractErrorStyle.Render(ev.errorMsg))
		s.WriteString(helpStyle.Render("Press any key to return"))
		return
	}

	if len(ev.tasks) == 0 {
		s.WriteString(extractLoadingStyle.Render("No actionable tasks found in this text"))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press any key to return"))
		return
	}

	fmt.Fprintf(s, "Tasks (%d/%d selected):\n\n", ev.SelectedCount(), len(ev.tasks))

	for i, task := range ev.tasks {
		cursor := "  "
		if i == ev.cursorIndex {
			cursor = "> "
		}
		checkbox := "[ ]"
		if ev.selected[i] {
			checkbox = "[x]"
		}

		text := task.Text
		if len(text) > 70 {
			text = text[:67] + "..."
		}

		effortBadge := ""
		if task.Effort > 0 {
			effortBadge = fmt.Sprintf(" %s", extractEffortStyle.Render(fmt.Sprintf("E%d", task.Effort)))
		}

		dupBadge := ""
		if task.Duplicate {
			dupBadge = " " + extractDupStyle.Render("[dup?]")
		}

		fmt.Fprintf(s, "%s%s %s%s%s\n", cursor, checkbox, text, effortBadge, dupBadge)
	}

	s.WriteString("\n")
	if ev.sourceLabel != "" || ev.backendName != "" {
		parts := make([]string, 0, 2)
		if ev.sourceLabel != "" {
			parts = append(parts, fmt.Sprintf("Source: %s", ev.sourceLabel))
		}
		if ev.backendName != "" {
			parts = append(parts, fmt.Sprintf("Backend: %s", ev.backendName))
		}
		s.WriteString(extractMetaStyle.Render(strings.Join(parts, "  •  ")))
		s.WriteString("\n")
	}
	s.WriteString(separatorStyle.Render("─────────────────────────────────"))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("[Space] Toggle  [A]ll  [N]one  [E]dit  [Enter] Import  [Esc] Cancel"))
}

func (ev *ExtractView) renderEditing(s *strings.Builder) {
	task := ev.tasks[ev.editIndex]
	fmt.Fprintf(s, "Editing task %d: %s\n\n", ev.editIndex+1,
		extractParentStyle.Render(truncate(task.Text, 60)))

	fieldNames := []string{"Text", "Effort", "Tags"}
	for i, name := range fieldNames {
		marker := "  "
		if i == ev.editField {
			marker = "> "
		}
		fmt.Fprintf(s, "%s%s: ", marker, name)
		if i == ev.editField {
			s.WriteString(ev.editInput.View())
		} else {
			switch i {
			case 0:
				s.WriteString(task.Text)
			case 1:
				fmt.Fprintf(s, "%d", task.Effort)
			case 2:
				s.WriteString(strings.Join(task.Tags, ", "))
			}
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("[Enter/Tab] Next field  [Esc] Done editing"))
}

// Styles for the extract view.
var (
	extractHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	extractKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	extractParentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Bold(true)

	extractLoadingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Italic(true)

	extractErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))

	extractEffortStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)

	extractDupStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Italic(true)

	extractMetaStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))
)
