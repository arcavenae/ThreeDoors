package tui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

const flashDuration = 3 * time.Second

var (
	// Status colors
	colorTodo       = lipgloss.Color("252")
	colorInProgress = lipgloss.Color("214")
	colorBlocked    = lipgloss.Color("196")
	colorInReview   = lipgloss.Color("39")
	colorComplete   = lipgloss.Color("82")
	colorAccent     = lipgloss.Color("63")
	colorSelected   = lipgloss.Color("86")
	colorGreeting   = lipgloss.Color("229")
	colorDoorBright = lipgloss.Color("255")

	// Per-door accent colors (left, center, right)
	doorColors = []lipgloss.Color{
		lipgloss.Color("86"),  // Door 0 (left) — cyan
		lipgloss.Color("212"), // Door 1 (center) — magenta
		lipgloss.Color("220"), // Door 2 (right) — yellow
	}

	// Door styles
	doorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	selectedDoorStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(colorDoorBright).
				Padding(1, 2)

	// Detail view styles
	detailBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	flashStyle = lipgloss.NewStyle().
			Foreground(colorComplete).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	moodHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	// Search styles
	searchResultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	searchSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSelected).
				Background(lipgloss.Color("236"))

	commandModeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))

	// Greeting style
	greetingStyle = lipgloss.NewStyle().
			Foreground(colorGreeting)

	// Separator style
	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	// Greeting messages pool — "progress over perfection" theme
	greetingMessages = []string{
		"Pick one. Start small. That's progress.",
		"Perfection is a trap. Progress is a practice.",
		"Three doors. One choice. Zero wrong answers.",
		"The best task to do is the one you actually start.",
		"You don't need to do it all. Just do one.",
		"Small steps count. Open a door.",
		"Done is better than perfect. Let's go.",
		"Every completed task is a win.",
	}

	// Celebration messages pool — varied completion messages
	celebrationMessages = []string{
		"Progress over perfection. Just pick one and start.",
		"Another one done! You're on a roll.",
		"Task complete! Small wins add up.",
		"Nice work. What's behind the next door?",
		"Done! That's one less thing on your plate.",
		"Crushed it! Keep the momentum going.",
		"One down. Progress feels good, doesn't it?",
		"Completed! You showed up and shipped it.",
		"That's progress! Every task matters.",
		"Well done! The best task is a done task.",
	}

	// Task-added messages pool — encouraging messages after adding a task
	taskAddedMessages = []string{
		"Task captured! Every task written down is a weight lifted.",
		"Added! Getting it out of your head is step one.",
		"New task logged. You're staying on top of things.",
		"Got it! One more thing you won't forget.",
		"Task added. Naming it is half the battle.",
		"Captured! Your future self will thank you.",
		"Added to the mix. Progress starts with awareness.",
		"Logged! You're building momentum just by tracking.",
	}

	// Door-refresh messages pool — encouraging messages when re-rolling doors
	doorRefreshMessages = []string{
		"Fresh options! Sometimes a new perspective helps.",
		"New doors, new possibilities.",
		"Shuffled! Trust your gut on the next pick.",
		"Re-rolled. Every choice is a good one.",
		"New set! The right task will catch your eye.",
		"Fresh draw. No wrong answers here.",
	}

	// Next-steps view styles
	nextStepsHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	nextStepsOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	// Health check styles
	healthOKStyle = lipgloss.NewStyle().
			Foreground(colorComplete).
			Bold(true)

	healthFailStyle = lipgloss.NewStyle().
			Foreground(colorBlocked).
			Bold(true)

	healthWarnStyle = lipgloss.NewStyle().
			Foreground(colorInProgress).
			Bold(true)

	healthSuggestionStyle = lipgloss.NewStyle().
				Foreground(colorInProgress)

	// Values/goals styles
	valuesFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)

	valuesHeaderStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	feedbackHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))

	improvementHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	valuesFooterSeparator = "  ·  "

	valuesSelectedPrefix = "▸ "
)

// StatusColor returns the lipgloss color for a given status string.
func StatusColor(status string) lipgloss.Color {
	switch status {
	case "todo":
		return colorTodo
	case "in-progress":
		return colorInProgress
	case "blocked":
		return colorBlocked
	case "in-review":
		return colorInReview
	case "complete":
		return colorComplete
	default:
		return colorTodo
	}
}
