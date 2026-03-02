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

	// Door styles
	doorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	selectedDoorStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(colorSelected).
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
