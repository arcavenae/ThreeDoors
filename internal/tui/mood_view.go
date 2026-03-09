package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var moodOptions = []string{
	"Focused",
	"Tired",
	"Stressed",
	"Energized",
	"Distracted",
	"Calm",
	"Other",
}

// MoodView displays the mood capture dialog.
type MoodView struct {
	customInput string
	isCustom    bool
	width       int
	hintEnabled bool
}

// NewMoodView creates a new mood capture view.
func NewMoodView() *MoodView {
	return &MoodView{}
}

// SetInlineHints sets the inline hint display state.
func (mv *MoodView) SetInlineHints(enabled bool) {
	mv.hintEnabled = enabled
}

// SetWidth sets the terminal width.
func (mv *MoodView) SetWidth(w int) {
	mv.width = w
}

// Update handles key input for mood selection.
func (mv *MoodView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if mv.isCustom {
			return mv.handleCustomInput(msg)
		}
		return mv.handleMoodSelection(msg)
	}
	return nil
}

func (mv *MoodView) handleMoodSelection(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	case "1":
		return moodCmd("Focused", "")
	case "2":
		return moodCmd("Tired", "")
	case "3":
		return moodCmd("Stressed", "")
	case "4":
		return moodCmd("Energized", "")
	case "5":
		return moodCmd("Distracted", "")
	case "6":
		return moodCmd("Calm", "")
	case "7":
		mv.isCustom = true
		mv.customInput = ""
	}
	return nil
}

func (mv *MoodView) handleCustomInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		if mv.customInput != "" {
			return moodCmd("Other", mv.customInput)
		}
	case "esc":
		mv.isCustom = false
		mv.customInput = ""
	case "backspace":
		if len(mv.customInput) > 0 {
			mv.customInput = mv.customInput[:len(mv.customInput)-1]
		}
	default:
		if len(msg.String()) == 1 && len(mv.customInput) < 100 {
			mv.customInput += msg.String()
		}
	}
	return nil
}

func moodCmd(mood, custom string) tea.Cmd {
	return func() tea.Msg {
		return MoodCapturedMsg{Mood: mood, CustomText: custom}
	}
}

// View renders the mood capture dialog.
func (mv *MoodView) View() string {
	s := strings.Builder{}
	s.WriteString(moodHeaderStyle.Render("How are you feeling?"))
	s.WriteString("\n\n")

	for i, mood := range moodOptions {
		if mv.hintEnabled {
			hint := renderInlineHint(fmt.Sprintf("%d", i+1), mv.hintEnabled)
			fmt.Fprintf(&s, "  %s %s\n", hint, mood)
		} else {
			fmt.Fprintf(&s, "  %d. %s\n", i+1, mood)
		}
	}

	s.WriteString("\n")

	if mv.isCustom {
		s.WriteString("Enter your mood: " + mv.customInput + "_\n")
	} else if mv.hintEnabled {
		h := func(key string) string { return renderInlineHint(key, mv.hintEnabled) }
		s.WriteString(helpStyle.Render(h("esc") + " Cancel"))
	} else {
		s.WriteString(helpStyle.Render("Press 1-7 to select | Esc to cancel"))
	}

	w := mv.width - 6
	if w < 40 {
		w = 40
	}
	return detailBorder.Width(w).Render(s.String())
}
