package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// importStep tracks the current step in the import flow.
type importStep int

const (
	importStepPath    importStep = iota // User enters file path
	importStepPreview                   // Show preview of parsed tasks
)

// ImportView handles the standalone :import command flow.
type ImportView struct {
	step         importStep
	textInput    textinput.Model
	importResult *core.ImportResult
	importError  string
	width        int
}

// NewImportView creates a new import view, optionally pre-filling a file path.
func NewImportView(prefilledPath string) *ImportView {
	ti := textinput.New()
	ti.Placeholder = "Path to task file (e.g. ~/tasks.txt)..."
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 50

	iv := &ImportView{
		step:      importStepPath,
		textInput: ti,
	}

	if prefilledPath != "" {
		ti.SetValue(prefilledPath)
		iv.textInput = ti
		// Try to parse immediately
		iv.tryParse(prefilledPath)
	}

	return iv
}

// SetWidth sets the terminal width for rendering.
func (iv *ImportView) SetWidth(w int) {
	iv.width = w
	if w > 4 {
		iv.textInput.Width = w - 4
	}
}

// tryParse attempts to parse the file at the given path.
// On success, advances to the preview step. On failure, sets importError.
func (iv *ImportView) tryParse(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}

	result, err := core.ImportTasksFromFile(path)
	if err != nil {
		iv.importError = err.Error()
		return
	}
	if len(result.Tasks) == 0 {
		iv.importError = "No tasks found in file"
		return
	}

	iv.importResult = result
	iv.importError = ""
	iv.step = importStepPreview
	iv.textInput.Blur()
}

// Update handles messages for the import view.
func (iv *ImportView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch iv.step {
		case importStepPath:
			return iv.updatePath(msg)
		case importStepPreview:
			return iv.updatePreview(msg)
		}
	}

	// Let textinput handle non-key messages when in path step
	if iv.step == importStepPath {
		var cmd tea.Cmd
		iv.textInput, cmd = iv.textInput.Update(msg)
		return cmd
	}

	return nil
}

func (iv *ImportView) updatePath(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEscape:
		return func() tea.Msg { return ReturnToDoorsMsg{} }

	case tea.KeyEnter:
		path := strings.TrimSpace(iv.textInput.Value())
		if path == "" {
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
		iv.tryParse(path)
		return nil
	}

	var cmd tea.Cmd
	iv.textInput, cmd = iv.textInput.Update(msg)
	return cmd
}

func (iv *ImportView) updatePreview(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n":
		return func() tea.Msg { return ReturnToDoorsMsg{} }

	case "enter", "y":
		if iv.importResult == nil {
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
		// Filter to only TODO tasks
		var todoTasks []*core.Task
		for _, t := range iv.importResult.Tasks {
			if t.Status == core.StatusTodo {
				todoTasks = append(todoTasks, t)
			}
		}
		source := filepath.Base(iv.importResult.SourcePath)
		return func() tea.Msg {
			return ImportConfirmedMsg{
				Tasks:  todoTasks,
				Source: source,
			}
		}
	}

	return nil
}

// View renders the import view.
func (iv *ImportView) View() string {
	switch iv.step {
	case importStepPath:
		return iv.viewPath()
	case importStepPreview:
		return iv.viewPreview()
	default:
		return ""
	}
}

func (iv *ImportView) viewPath() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Import Tasks"))
	fmt.Fprintf(&s, "Import tasks from a file.\n")
	fmt.Fprintf(&s, "Supports plain text (one per line) and Markdown checkboxes.\n\n")

	fmt.Fprintf(&s, "%s\n", iv.textInput.View())

	if iv.importError != "" {
		fmt.Fprintf(&s, "\n%s\n", healthFailStyle.Render(iv.importError))
	}

	fmt.Fprintf(&s, "\n%s", helpStyle.Render("Enter to parse | Esc to cancel"))

	return s.String()
}

func (iv *ImportView) viewPreview() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Import Preview"))

	if iv.importResult == nil {
		fmt.Fprintf(&s, "No tasks to preview.\n")
		return s.String()
	}

	total := len(iv.importResult.Tasks)
	todoCount := 0
	for _, t := range iv.importResult.Tasks {
		if t.Status == core.StatusTodo {
			todoCount++
		}
	}

	fmt.Fprintf(&s, "Found %s in %s format.\n",
		headerStyle.Render(fmt.Sprintf("%d tasks", total)),
		iv.importResult.Format)
	if todoCount < total {
		fmt.Fprintf(&s, "%d incomplete, %d already done.\n", todoCount, total-todoCount)
	}
	fmt.Fprintf(&s, "\n")

	// Show first few tasks as preview
	previewMax := 5
	if previewMax > total {
		previewMax = total
	}
	for i := 0; i < previewMax; i++ {
		t := iv.importResult.Tasks[i]
		status := "[ ]"
		if t.Status == core.StatusComplete {
			status = "[x]"
		}
		fmt.Fprintf(&s, "  %s %s\n", status, t.Text)
	}
	if total > previewMax {
		fmt.Fprintf(&s, "  %s\n", helpStyle.Render(fmt.Sprintf("... and %d more", total-previewMax)))
	}

	fmt.Fprintf(&s, "\n%s", helpStyle.Render("Enter/y to import | Esc/n to cancel"))

	return s.String()
}
