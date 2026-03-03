package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// onboardingStep tracks the current step in the onboarding wizard.
type onboardingStep int

const (
	stepWelcome onboardingStep = iota
	stepKeybindings
	stepDone
)

// OnboardingView guides first-time users through the Three Doors concept.
type OnboardingView struct {
	step       onboardingStep
	width      int
	triedKeys  map[string]bool
	lastAction string
}

// OnboardingCompletedMsg is sent when onboarding finishes.
type OnboardingCompletedMsg struct{}

// NewOnboardingView creates a new onboarding wizard.
func NewOnboardingView() *OnboardingView {
	return &OnboardingView{
		triedKeys: make(map[string]bool),
	}
}

// SetWidth sets the terminal width for rendering.
func (ov *OnboardingView) SetWidth(w int) {
	ov.width = w
}

// Update handles key input for the onboarding wizard.
func (ov *OnboardingView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Skip remaining onboarding at any step
		if key == "esc" || key == "ctrl+c" {
			return func() tea.Msg { return OnboardingCompletedMsg{} }
		}

		switch ov.step {
		case stepWelcome:
			if key == "enter" || key == " " {
				ov.step = stepKeybindings
			}
			return nil

		case stepKeybindings:
			switch key {
			case "enter":
				ov.step = stepDone
				return nil
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

		case stepDone:
			if key == "enter" || key == " " {
				return func() tea.Msg { return OnboardingCompletedMsg{} }
			}
			return nil
		}
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

	stepIndicator := helpStyle.Render("Step 1 of 3")
	fmt.Fprintf(&s, "%s\n", stepIndicator)
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

	stepIndicator := helpStyle.Render("Step 2 of 3")
	fmt.Fprintf(&s, "%s\n", stepIndicator)
	fmt.Fprintf(&s, "%s", helpStyle.Render("Enter to continue | Esc to skip"))

	return s.String()
}

func (ov *OnboardingView) viewDone() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("You're All Set!"))

	fmt.Fprintf(&s, "Here are a few more keys you'll find useful:\n\n")

	fmt.Fprintf(&s, "  %-16s  %s\n", "Enter", "Open selected task")
	fmt.Fprintf(&s, "  %-16s  %s\n", "C", "Complete a task")
	fmt.Fprintf(&s, "  %-16s  %s\n", "M", "Log your mood")
	fmt.Fprintf(&s, "  %-16s  %s\n", "/", "Search tasks")
	fmt.Fprintf(&s, "  %-16s  %s\n", ":", "Command palette")
	fmt.Fprintf(&s, "  %-16s  %s\n", "Q", "Quit")

	fmt.Fprintf(&s, "\n%s\n\n", headerStyle.Render("Remember: progress over perfection."))

	stepIndicator := helpStyle.Render("Step 3 of 3")
	fmt.Fprintf(&s, "%s\n", stepIndicator)
	fmt.Fprintf(&s, "%s", helpStyle.Render("Enter to start | Esc to skip"))

	return s.String()
}
