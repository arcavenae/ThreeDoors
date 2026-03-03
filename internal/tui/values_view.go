package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ValuesViewMode tracks whether we are in setup or edit mode.
type ValuesViewMode int

const (
	ValuesSetupMode ValuesViewMode = iota
	ValuesEditMode
)

// ValuesView handles the values/goals setup and editing flow.
type ValuesView struct {
	mode          ValuesViewMode
	config        *core.ValuesConfig
	textInput     textinput.Model
	selectedIndex int
	width         int
}

// NewValuesSetupView creates a values view in setup mode.
func NewValuesSetupView(cfg *core.ValuesConfig) *ValuesView {
	ti := textinput.New()
	ti.Placeholder = "Enter a value or goal (1-5 total)..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 40

	return &ValuesView{
		mode:      ValuesSetupMode,
		config:    cfg,
		textInput: ti,
	}
}

// NewValuesEditView creates a values view in edit mode.
func NewValuesEditView(cfg *core.ValuesConfig) *ValuesView {
	ti := textinput.New()
	ti.Placeholder = "Enter new value..."
	ti.CharLimit = 200
	ti.Width = 40

	return &ValuesView{
		mode:      ValuesEditMode,
		config:    cfg,
		textInput: ti,
	}
}

// SetWidth sets the terminal width for rendering.
func (vv *ValuesView) SetWidth(w int) {
	vv.width = w
	if w > 4 {
		vv.textInput.Width = w - 4
	}
}

// Update handles messages for the values view.
func (vv *ValuesView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if vv.mode == ValuesSetupMode {
			return vv.updateSetup(msg)
		}
		return vv.updateEdit(msg)
	}

	var cmd tea.Cmd
	vv.textInput, cmd = vv.textInput.Update(msg)
	return cmd
}

func (vv *ValuesView) updateSetup(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEscape:
		if vv.config.HasValues() {
			return vv.saveAndReturn()
		}
		return func() tea.Msg { return ReturnToDoorsMsg{} }

	case tea.KeyEnter:
		text := strings.TrimSpace(vv.textInput.Value())
		if text == "" {
			if vv.config.HasValues() {
				return vv.saveAndReturn()
			}
			return nil
		}
		if err := vv.config.AddValue(text); err != nil {
			return func() tea.Msg { return FlashMsg{Text: err.Error()} }
		}
		vv.textInput.SetValue("")
		if len(vv.config.Values) >= 5 {
			return vv.saveAndReturn()
		}
		return nil
	}

	var cmd tea.Cmd
	vv.textInput, cmd = vv.textInput.Update(msg)
	return cmd
}

func (vv *ValuesView) updateEdit(msg tea.KeyMsg) tea.Cmd {
	if vv.textInput.Focused() {
		return vv.updateEditInput(msg)
	}

	switch msg.String() {
	case "esc", "q":
		return vv.saveAndReturn()
	case "up", "k":
		if vv.selectedIndex > 0 {
			vv.selectedIndex--
		}
	case "down", "j":
		if vv.selectedIndex < len(vv.config.Values)-1 {
			vv.selectedIndex++
		}
	case "d", "backspace":
		if len(vv.config.Values) > 0 {
			_ = vv.config.RemoveValue(vv.selectedIndex)
			if vv.selectedIndex >= len(vv.config.Values) && vv.selectedIndex > 0 {
				vv.selectedIndex--
			}
		}
	case "K":
		if len(vv.config.Values) > 0 {
			_ = vv.config.MoveUp(vv.selectedIndex)
			if vv.selectedIndex > 0 {
				vv.selectedIndex--
			}
		}
	case "J":
		if len(vv.config.Values) > 0 {
			_ = vv.config.MoveDown(vv.selectedIndex)
			if vv.selectedIndex < len(vv.config.Values)-1 {
				vv.selectedIndex++
			}
		}
	case "a":
		if len(vv.config.Values) < 5 {
			vv.textInput.Focus()
			vv.textInput.SetValue("")
		}
	}
	return nil
}

func (vv *ValuesView) updateEditInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEscape:
		vv.textInput.Blur()
		return nil
	case tea.KeyEnter:
		text := strings.TrimSpace(vv.textInput.Value())
		if text != "" {
			if err := vv.config.AddValue(text); err != nil {
				return func() tea.Msg { return FlashMsg{Text: err.Error()} }
			}
		}
		vv.textInput.Blur()
		vv.textInput.SetValue("")
		return nil
	}

	var cmd tea.Cmd
	vv.textInput, cmd = vv.textInput.Update(msg)
	return cmd
}

func (vv *ValuesView) saveAndReturn() tea.Cmd {
	cfg := vv.config
	return func() tea.Msg {
		return ValuesSavedMsg{Config: cfg}
	}
}

// View renders the values view.
func (vv *ValuesView) View() string {
	s := strings.Builder{}

	if vv.mode == ValuesSetupMode {
		s.WriteString(headerStyle.Render("ThreeDoors - Values & Goals Setup"))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Define what matters most to you (1-5 values/goals)."))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("These will appear during your task sessions as a reminder."))
		s.WriteString("\n\n")

		if len(vv.config.Values) > 0 {
			s.WriteString(valuesHeaderStyle.Render("Your values:"))
			s.WriteString("\n")
			for i, v := range vv.config.Values {
				fmt.Fprintf(&s, "  %d. %s\n", i+1, v)
			}
			s.WriteString("\n")
		}

		s.WriteString(vv.textInput.View())
		s.WriteString("\n\n")

		count := len(vv.config.Values)
		remaining := 5 - count
		if count == 0 {
			s.WriteString(helpStyle.Render("Enter your first value/goal and press Enter"))
		} else if remaining > 0 {
			s.WriteString(helpStyle.Render(fmt.Sprintf("Enter to add (%d/%d) | Enter empty to finish | Esc to finish", count, 5)))
		}
	} else {
		s.WriteString(headerStyle.Render("ThreeDoors - Edit Values & Goals"))
		s.WriteString("\n\n")

		if len(vv.config.Values) == 0 {
			s.WriteString(helpStyle.Render("No values defined. Press 'a' to add one."))
			s.WriteString("\n\n")
		} else {
			for i, v := range vv.config.Values {
				prefix := "  "
				if i == vv.selectedIndex {
					prefix = valuesSelectedPrefix
				}
				line := fmt.Sprintf("%s%d. %s", prefix, i+1, v)
				if i == vv.selectedIndex {
					s.WriteString(searchSelectedStyle.Render(line))
				} else {
					s.WriteString(line)
				}
				s.WriteString("\n")
			}
			s.WriteString("\n")
		}

		if vv.textInput.Focused() {
			s.WriteString(vv.textInput.View())
			s.WriteString("\n\n")
			s.WriteString(helpStyle.Render("Enter to add | Esc to cancel"))
		} else {
			s.WriteString(helpStyle.Render("↑/↓ navigate | d delete | K/J reorder | a add | Esc/q done"))
		}
	}

	return s.String()
}

// RenderValuesFooter renders a subtle footer showing the user's values/goals.
func RenderValuesFooter(cfg *core.ValuesConfig) string {
	if cfg == nil || !cfg.HasValues() {
		return ""
	}

	joined := strings.Join(cfg.Values, valuesFooterSeparator)
	return "\n" + valuesFooterStyle.Render(joined)
}
