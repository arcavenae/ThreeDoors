package tui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

const flashDuration = 3 * time.Second

var (
	// Status colors — use CompleteColor for graceful degradation on 16-color terminals.
	colorTodo       lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#d0d0d0", ANSI256: "252", ANSI: "7"}
	colorInProgress lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#ffaf00", ANSI256: "214", ANSI: "3"}
	colorBlocked    lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#ff0000", ANSI256: "196", ANSI: "1"}
	colorInReview   lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#00afff", ANSI256: "39", ANSI: "4"}
	colorComplete   lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#5fff00", ANSI256: "82", ANSI: "2"}
	colorAccent     lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#5f5fff", ANSI256: "63", ANSI: "5"}
	colorSelected   lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#5fffd7", ANSI256: "86", ANSI: "6"}
	colorGreeting   lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#ffffaf", ANSI256: "229", ANSI: "11"}
	colorDoorBright lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#eeeeee", ANSI256: "255", ANSI: "15"}

	// Per-door accent colors (left, center, right)
	doorColors = []lipgloss.TerminalColor{
		lipgloss.CompleteColor{TrueColor: "#5fffd7", ANSI256: "86", ANSI: "6"},   // Door 0 (left) — cyan
		lipgloss.CompleteColor{TrueColor: "#ff87d7", ANSI256: "212", ANSI: "13"}, // Door 1 (center) — magenta
		lipgloss.CompleteColor{TrueColor: "#ffd700", ANSI256: "220", ANSI: "11"}, // Door 2 (right) — yellow
	}

	// Door styles
	doorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	// unselectedDoorStyle dims unselected doors when a selection is active,
	// creating a focus funnel toward the selected door.
	unselectedDoorStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.CompleteColor{TrueColor: "#585858", ANSI256: "240", ANSI: "8"}).
				Padding(1, 2).
				Faint(true)

	selectedDoorStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(colorDoorBright).
				Padding(1, 2).
				Bold(true).
				Foreground(colorDoorBright)

	// selectedContentStyle applies bold + bright foreground to door content text.
	selectedContentStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorDoorBright)

	// unselectedContentStyle dims door content when another door is selected.
	unselectedContentStyle = lipgloss.NewStyle().
				Faint(true)

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
			Foreground(lipgloss.CompleteColor{TrueColor: "#626262", ANSI256: "241", ANSI: "8"})

	moodHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.CompleteColor{TrueColor: "#ff5faf", ANSI256: "205", ANSI: "13"})

	// Search styles
	searchResultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#d0d0d0", ANSI256: "252", ANSI: "7"})

	searchSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSelected).
				Background(lipgloss.CompleteColor{TrueColor: "#303030", ANSI256: "236", ANSI: "0"})

	commandModeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.CompleteColor{TrueColor: "#ffaf00", ANSI256: "214", ANSI: "3"})

	// Greeting style
	greetingStyle = lipgloss.NewStyle().
			Foreground(colorGreeting)

	// Separator style
	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.CompleteColor{TrueColor: "#444444", ANSI256: "238", ANSI: "8"})

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
				Foreground(lipgloss.CompleteColor{TrueColor: "#5fffd7", ANSI256: "86", ANSI: "6"})

	nextStepsOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#d0d0d0", ANSI256: "252", ANSI: "7"})

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
				Foreground(lipgloss.CompleteColor{TrueColor: "#767676", ANSI256: "243", ANSI: "8"}).
				Italic(true)

	valuesHeaderStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	feedbackHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.CompleteColor{TrueColor: "#ffaf00", ANSI256: "214", ANSI: "3"})

	valuesFooterSeparator = "  ·  "

	valuesSelectedPrefix = "▸ "

	// Badge style for category tags on door cards
	badgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.CompleteColor{TrueColor: "#767676", ANSI256: "243", ANSI: "8"})

	// Conflict view styles
	conflictHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.CompleteColor{TrueColor: "#ff0000", ANSI256: "196", ANSI: "1"})

	conflictLocalStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.CompleteColor{TrueColor: "#5fffd7", ANSI256: "86", ANSI: "6"}).
				Padding(1, 2)

	conflictRemoteStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.CompleteColor{TrueColor: "#ff87d7", ANSI256: "212", ANSI: "13"}).
				Padding(1, 2)

	conflictDiffStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#ffaf00", ANSI256: "214", ANSI: "3"}).
				Bold(true)

	// Sync log styles
	syncLogHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.CompleteColor{TrueColor: "#00afff", ANSI256: "39", ANSI: "4"})

	syncLogEntryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#d0d0d0", ANSI256: "252", ANSI: "7"})

	syncLogTimestampStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#767676", ANSI256: "243", ANSI: "8"})

	syncLogErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#ff0000", ANSI256: "196", ANSI: "1"})

	// Proposal view styles
	proposalHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.CompleteColor{TrueColor: "#5fffd7", ANSI256: "86", ANSI: "6"})

	proposalSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSelected).
				Background(lipgloss.CompleteColor{TrueColor: "#303030", ANSI256: "236", ANSI: "0"})

	proposalStaleStyle = lipgloss.NewStyle().
				Faint(true).
				Strikethrough(true)

	proposalBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#ffaf00", ANSI256: "214", ANSI: "3"}).
				Bold(true)

	focusBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	proposalTypeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#00afff", ANSI256: "39", ANSI: "4"})

	proposalDiffAddStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#5fff00", ANSI256: "82", ANSI: "2"})

	proposalDiffRemoveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#ff0000", ANSI256: "196", ANSI: "1"})

	proposalPaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.CompleteColor{TrueColor: "#5f5fff", ANSI256: "63", ANSI: "5"}).
				Padding(1, 2)

	// Stats dashboard styles (Epic 40)
	statsPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#555555", Dark: "#555555"})

	statsDashboardHeaderStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#EEEEEE"}).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.AdaptiveColor{Light: "#555555", Dark: "#555555"}).
					Align(lipgloss.Center)

	statsHeroStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FCD34D")).
			Bold(true).
			Align(lipgloss.Center)

	statsSectionHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#EEEEEE"}).
				Bold(true)

	// Mood bar chart colors (Story 40.4)
	moodColors = map[string]lipgloss.AdaptiveColor{
		"Focused":   {Light: "#2563EB", Dark: "#60A5FA"},
		"Energized": {Light: "#D97706", Dark: "#FBBF24"},
		"Calm":      {Light: "#059669", Dark: "#34D399"},
		"Tired":     {Light: "#6B7280", Dark: "#9CA3AF"},
	}

	defaultMoodColor = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}
)

// StatusColor returns the lipgloss color for a given status string.
func StatusColor(status string) lipgloss.TerminalColor {
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
	case "deferred":
		return lipgloss.CompleteColor{TrueColor: "#767676", ANSI256: "243", ANSI: "8"}
	case "archived":
		return lipgloss.CompleteColor{TrueColor: "#585858", ANSI256: "240", ANSI: "8"}
	default:
		return colorTodo
	}
}
