package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// onboardingStep tracks the current step in the onboarding wizard.
type onboardingStep int

const (
	stepWelcome onboardingStep = iota
	stepKeybindings
	stepValues
	stepTheme
	stepImport
	stepImportPreview
	stepDone
)

const onboardingTotalSteps = 6

// OnboardingView guides first-time users through the Three Doors concept,
// values/goals setup, theme selection, and task import.
type OnboardingView struct {
	step       onboardingStep
	width      int
	triedKeys  map[string]bool
	lastAction string

	// Values/goals state
	values    []string
	textInput textinput.Model

	// Theme picker state
	themePicker   *ThemePicker
	selectedTheme string

	// Import state
	importResult *core.ImportResult
	importError  string
}

// OnboardingCompletedMsg is sent when onboarding finishes.
type OnboardingCompletedMsg struct {
	Values        []string
	ImportedTasks []*core.Task
	SelectedTheme string
}

// NewOnboardingView creates a new onboarding wizard.
func NewOnboardingView() *OnboardingView {
	ti := textinput.New()
	ti.CharLimit = 200
	ti.Width = 40

	reg := themes.NewDefaultRegistry()

	return &OnboardingView{
		triedKeys:     make(map[string]bool),
		textInput:     ti,
		themePicker:   NewThemePicker(reg, ""),
		selectedTheme: themes.DefaultThemeName,
	}
}

// SetWidth sets the terminal width for rendering.
func (ov *OnboardingView) SetWidth(w int) {
	ov.width = w
	if w > 6 {
		ov.textInput.Width = w - 6
	}
	if ov.themePicker != nil {
		ov.themePicker.SetWidth(w)
	}
}

func (ov *OnboardingView) completeMsg() tea.Cmd {
	values := ov.values
	var imported []*core.Task
	if ov.importResult != nil {
		imported = ov.importResult.Tasks
	}
	selectedTheme := ov.selectedTheme
	return func() tea.Msg {
		return OnboardingCompletedMsg{
			Values:        values,
			ImportedTasks: imported,
			SelectedTheme: selectedTheme,
		}
	}
}

func (ov *OnboardingView) stepNumber() int {
	switch ov.step {
	case stepWelcome:
		return 1
	case stepKeybindings:
		return 2
	case stepValues:
		return 3
	case stepTheme:
		return 4
	case stepImport, stepImportPreview:
		return 5
	case stepDone:
		return 6
	}
	return 1
}

// Update handles key input for the onboarding wizard.
func (ov *OnboardingView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Skip remaining onboarding at any step (except when typing)
		if key == "ctrl+c" {
			return ov.completeMsg()
		}

		switch ov.step {
		case stepWelcome:
			return ov.updateWelcome(key)
		case stepKeybindings:
			return ov.updateKeybindings(key)
		case stepValues:
			return ov.updateValues(msg)
		case stepTheme:
			return ov.updateTheme(msg)
		case stepImport:
			return ov.updateImport(msg)
		case stepImportPreview:
			return ov.updateImportPreview(key)
		case stepDone:
			return ov.updateDone(key)
		}
	}

	// Let textinput handle non-key messages when active
	if ov.step == stepValues || ov.step == stepImport {
		var cmd tea.Cmd
		ov.textInput, cmd = ov.textInput.Update(msg)
		return cmd
	}

	return nil
}

func (ov *OnboardingView) updateWelcome(key string) tea.Cmd {
	if key == "esc" {
		return ov.completeMsg()
	}
	if key == "enter" || key == " " {
		ov.step = stepKeybindings
	}
	return nil
}

func (ov *OnboardingView) updateKeybindings(key string) tea.Cmd {
	if key == "esc" {
		return ov.completeMsg()
	}
	switch key {
	case "enter":
		ov.step = stepValues
		ov.textInput.Placeholder = "Enter a value or goal..."
		ov.textInput.SetValue("")
		ov.textInput.Focus()
	case "a", "left":
		ov.triedKeys["left"] = true
		ov.lastAction = "Left door selected!"
	case "w", "up":
		ov.triedKeys["up"] = true
		ov.lastAction = "Center door selected!"
	case "d", "right":
		ov.triedKeys["right"] = true
		ov.lastAction = "Right door selected!"
	case "s", "down":
		ov.triedKeys["reroll"] = true
		ov.lastAction = "Doors re-rolled!"
	default:
		ov.lastAction = ""
	}
	return nil
}

func (ov *OnboardingView) updateValues(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch msg.Type {
	case tea.KeyEscape:
		ov.step = stepTheme
		ov.textInput.Blur()
		return nil

	case tea.KeyEnter:
		text := strings.TrimSpace(ov.textInput.Value())
		if text == "" {
			// Empty enter = done with values, move to theme picker
			ov.step = stepTheme
			ov.textInput.Blur()
			return nil
		}
		if len(ov.values) >= 5 {
			return nil
		}
		if len(text) > 200 {
			text = text[:200]
		}
		ov.values = append(ov.values, text)
		ov.textInput.SetValue("")
		if len(ov.values) >= 5 {
			ov.step = stepTheme
			ov.textInput.Blur()
		}
		return nil
	}

	_ = key
	var cmd tea.Cmd
	ov.textInput, cmd = ov.textInput.Update(msg)
	return cmd
}

func (ov *OnboardingView) viewTheme() string {
	var s strings.Builder
	s.WriteString(ov.themePicker.View())
	fmt.Fprintf(&s, "\n%s", helpStyle.Render(fmt.Sprintf("Step %d of %d", ov.stepNumber(), onboardingTotalSteps)))
	return s.String()
}

func (ov *OnboardingView) updateTheme(msg tea.KeyMsg) tea.Cmd {
	cmd := ov.themePicker.Update(msg)
	if cmd != nil {
		// Intercept ThemeSelectedMsg to capture selection and advance step
		result := cmd()
		if sel, ok := result.(ThemeSelectedMsg); ok {
			ov.selectedTheme = sel.Name
			ov.step = stepImport
			ov.textInput.Placeholder = "Path to task file (e.g. ~/tasks.txt)..."
			ov.textInput.SetValue("")
			ov.textInput.Focus()
			ov.importError = ""
			return nil
		}
	}
	return nil
}

func (ov *OnboardingView) updateImport(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEscape:
		ov.step = stepDone
		ov.textInput.Blur()
		return nil

	case tea.KeyEnter:
		path := strings.TrimSpace(ov.textInput.Value())
		if path == "" {
			// Skip import
			ov.step = stepDone
			ov.textInput.Blur()
			return nil
		}

		result, err := core.ImportTasksFromFile(path)
		if err != nil {
			ov.importError = err.Error()
			return nil
		}
		if len(result.Tasks) == 0 {
			ov.importError = "No tasks found in file"
			return nil
		}

		ov.importResult = result
		ov.importError = ""
		ov.step = stepImportPreview
		ov.textInput.Blur()
		return nil
	}

	var cmd tea.Cmd
	ov.textInput, cmd = ov.textInput.Update(msg)
	return cmd
}

func (ov *OnboardingView) updateImportPreview(key string) tea.Cmd {
	switch key {
	case "esc", "n":
		// Reject import
		ov.importResult = nil
		ov.step = stepDone
	case "enter", "y":
		// Accept import
		ov.step = stepDone
	}
	return nil
}

func (ov *OnboardingView) updateDone(key string) tea.Cmd {
	if key == "esc" || key == "enter" || key == " " {
		return ov.completeMsg()
	}
	return nil
}

// View renders the current onboarding step.
func (ov *OnboardingView) View() string {
	w := ov.width - 6
	if w < 40 {
		w = 40
	}

	var content string
	switch ov.step {
	case stepWelcome:
		content = ov.viewWelcome()
	case stepKeybindings:
		content = ov.viewKeybindings()
	case stepValues:
		content = ov.viewValues()
	case stepTheme:
		content = ov.viewTheme()
	case stepImport:
		content = ov.viewImport()
	case stepImportPreview:
		content = ov.viewImportPreview()
	case stepDone:
		content = ov.viewDone()
	}

	return detailBorder.Width(w).Render(content)
}

func (ov *OnboardingView) viewWelcome() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Welcome to ThreeDoors"))

	fmt.Fprintf(&s, "ThreeDoors helps you overcome task paralysis by\n")
	fmt.Fprintf(&s, "showing you just %s at a time.\n\n", headerStyle.Render("three tasks"))

	fmt.Fprintf(&s, "Instead of staring at a long to-do list,\n")
	fmt.Fprintf(&s, "you pick from three doors — like a game show.\n\n")

	fmt.Fprintf(&s, "The trick? %s.\n", headerStyle.Render("There are no wrong answers"))
	fmt.Fprintf(&s, "Every door leads to progress.\n\n")

	fmt.Fprintf(&s, "%s\n", helpStyle.Render(fmt.Sprintf("Step %d of %d", ov.stepNumber(), onboardingTotalSteps)))
	fmt.Fprintf(&s, "%s", helpStyle.Render("Enter to continue | Esc to skip"))

	return s.String()
}

func (ov *OnboardingView) viewKeybindings() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Key Bindings"))
	fmt.Fprintf(&s, "Try the keys below to see how navigation works:\n\n")

	keys := []struct {
		keys   string
		action string
		tried  string
	}{
		{"A / Left Arrow", "Select left door", "left"},
		{"W / Up Arrow", "Select center door", "up"},
		{"D / Right Arrow", "Select right door", "right"},
		{"S / Down Arrow", "Re-roll doors", "reroll"},
	}

	for _, k := range keys {
		check := "  "
		if ov.triedKeys[k.tried] {
			check = flashStyle.Render("* ")
		}
		fmt.Fprintf(&s, "  %s%-16s  %s\n", check, k.keys, k.action)
	}

	if ov.lastAction != "" {
		fmt.Fprintf(&s, "\n  %s\n", flashStyle.Render(ov.lastAction))
	}

	triedCount := len(ov.triedKeys)
	fmt.Fprintf(&s, "\n%s\n", helpStyle.Render(fmt.Sprintf("Tried %d of 4 keys", triedCount)))

	fmt.Fprintf(&s, "%s\n", helpStyle.Render(fmt.Sprintf("Step %d of %d", ov.stepNumber(), onboardingTotalSteps)))
	fmt.Fprintf(&s, "%s", helpStyle.Render("Enter to continue | Esc to skip"))

	return s.String()
}

func (ov *OnboardingView) viewValues() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Values & Goals (not tasks)"))
	fmt.Fprintf(&s, "What are your motivational anchors? (e.g. \"family\",\n")
	fmt.Fprintf(&s, "\"health\", \"financial freedom\") These appear as a\n")
	fmt.Fprintf(&s, "small footer reminder — they are NOT your tasks.\n\n")
	fmt.Fprintf(&s, "%s\n\n", helpStyle.Render("Tasks are added later with ':add' or by editing tasks.yaml"))

	if len(ov.values) > 0 {
		fmt.Fprintf(&s, "%s\n", valuesHeaderStyle.Render("Your values:"))
		for i, v := range ov.values {
			fmt.Fprintf(&s, "  %d. %s\n", i+1, v)
		}
		fmt.Fprintf(&s, "\n")
	}

	fmt.Fprintf(&s, "%s\n\n", ov.textInput.View())

	count := len(ov.values)
	remaining := 5 - count
	if count == 0 {
		fmt.Fprintf(&s, "%s\n", helpStyle.Render("Type a value/goal and press Enter"))
	} else if remaining > 0 {
		fmt.Fprintf(&s, "%s\n", helpStyle.Render(fmt.Sprintf("Added %d/5 | Enter to add more | Empty Enter to continue", count)))
	} else {
		fmt.Fprintf(&s, "%s\n", helpStyle.Render("Maximum 5 values reached"))
	}

	fmt.Fprintf(&s, "%s\n", helpStyle.Render(fmt.Sprintf("Step %d of %d", ov.stepNumber(), onboardingTotalSteps)))
	fmt.Fprintf(&s, "%s", helpStyle.Render("Enter empty to continue | Esc to skip"))

	return s.String()
}

func (ov *OnboardingView) viewImport() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Import Tasks"))
	fmt.Fprintf(&s, "Already have tasks in a file? Import them now.\n")
	fmt.Fprintf(&s, "Supports plain text (one per line) and Markdown checkboxes.\n\n")

	fmt.Fprintf(&s, "%s\n", ov.textInput.View())

	if ov.importError != "" {
		fmt.Fprintf(&s, "\n%s\n", healthFailStyle.Render(ov.importError))
	}

	fmt.Fprintf(&s, "\n%s\n", helpStyle.Render("Enter a file path or leave empty to skip"))
	fmt.Fprintf(&s, "%s\n", helpStyle.Render("You can also import later with :import"))
	fmt.Fprintf(&s, "%s\n", helpStyle.Render(fmt.Sprintf("Step %d of %d", ov.stepNumber(), onboardingTotalSteps)))
	fmt.Fprintf(&s, "%s", helpStyle.Render("Enter to import | Esc to skip"))

	return s.String()
}

func (ov *OnboardingView) viewImportPreview() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Import Preview"))

	if ov.importResult == nil {
		fmt.Fprintf(&s, "No tasks to preview.\n")
		return s.String()
	}

	total := len(ov.importResult.Tasks)
	todoCount := 0
	for _, t := range ov.importResult.Tasks {
		if t.Status == core.StatusTodo {
			todoCount++
		}
	}

	fmt.Fprintf(&s, "Found %s in %s format.\n", headerStyle.Render(fmt.Sprintf("%d tasks", total)), ov.importResult.Format)
	if todoCount < total {
		fmt.Fprintf(&s, "%d incomplete, %d already done.\n", todoCount, total-todoCount)
	}
	fmt.Fprintf(&s, "\n")

	// Show first few tasks as preview
	previewMax := 5
	if previewMax > total {
		previewMax = total
	}
	fmt.Fprintf(&s, "%s\n", valuesHeaderStyle.Render("Preview:"))
	for i := 0; i < previewMax; i++ {
		t := ov.importResult.Tasks[i]
		status := "[ ]"
		if t.Status == core.StatusComplete {
			status = "[x]"
		}
		fmt.Fprintf(&s, "  %s %s\n", status, t.Text)
	}
	if total > previewMax {
		fmt.Fprintf(&s, "  %s\n", helpStyle.Render(fmt.Sprintf("... and %d more", total-previewMax)))
	}

	fmt.Fprintf(&s, "\n%s\n", helpStyle.Render(fmt.Sprintf("Step %d of %d", ov.stepNumber(), onboardingTotalSteps)))
	fmt.Fprintf(&s, "%s", helpStyle.Render("Enter/y to import | Esc/n to skip"))

	return s.String()
}

func (ov *OnboardingView) viewDone() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("You're All Set!"))

	// Summary of what was set up
	if len(ov.values) > 0 {
		fmt.Fprintf(&s, "%s %d values/goals saved\n", flashStyle.Render("*"), len(ov.values))
	}
	if ov.selectedTheme != "" {
		fmt.Fprintf(&s, "%s Theme: %s\n", flashStyle.Render("*"), ov.selectedTheme)
	}
	if ov.importResult != nil && len(ov.importResult.Tasks) > 0 {
		fmt.Fprintf(&s, "%s %d tasks imported\n", flashStyle.Render("*"), len(ov.importResult.Tasks))
	}
	if len(ov.values) > 0 || ov.selectedTheme != "" || (ov.importResult != nil && len(ov.importResult.Tasks) > 0) {
		fmt.Fprintf(&s, "\n")
	}

	fmt.Fprintf(&s, "Here are a few more keys you'll find useful:\n\n")

	fmt.Fprintf(&s, "  %-16s  %s\n", "Enter", "Open selected task")
	fmt.Fprintf(&s, "  %-16s  %s\n", "C", "Complete a task")
	fmt.Fprintf(&s, "  %-16s  %s\n", "M", "Log your mood")
	fmt.Fprintf(&s, "  %-16s  %s\n", "/", "Search tasks")
	fmt.Fprintf(&s, "  %-16s  %s\n", ":", "Command palette")
	fmt.Fprintf(&s, "  %-16s  %s\n", "Q", "Quit")

	fmt.Fprintf(&s, "\n%s\n\n", headerStyle.Render("Remember: progress over perfection."))

	fmt.Fprintf(&s, "%s\n", helpStyle.Render(fmt.Sprintf("Step %d of %d", ov.stepNumber(), onboardingTotalSteps)))
	fmt.Fprintf(&s, "%s", helpStyle.Render("Enter to start | Esc to skip"))

	return s.String()
}
