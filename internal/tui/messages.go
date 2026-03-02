package tui

import (
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

// ReturnToDoorsMsg is sent when the user wants to go back to the doors view.
type ReturnToDoorsMsg struct{}

// TaskUpdatedMsg is sent when a task has been modified.
type TaskUpdatedMsg struct {
	Task *tasks.Task
}

// ShowMoodMsg is sent to open the mood capture dialog.
type ShowMoodMsg struct{}

// MoodCapturedMsg is sent when mood has been recorded.
type MoodCapturedMsg struct {
	Mood       string
	CustomText string
}

// TaskCompletedMsg is sent when a task is marked complete.
type TaskCompletedMsg struct {
	Task *tasks.Task
}

// FlashMsg triggers a temporary message display.
type FlashMsg struct {
	Text string
}

// ClearFlashMsg clears the flash message.
type ClearFlashMsg struct{}

// SearchResultSelectedMsg is sent when a user selects a task from search results.
type SearchResultSelectedMsg struct {
	Task *tasks.Task
}

// TaskAddedMsg is sent when the :add command creates a new task.
type TaskAddedMsg struct {
	Task *tasks.Task
}

// SearchClosedMsg is sent when the user exits search mode.
type SearchClosedMsg struct{}

// ReturnToSearchMsg is sent to restore search view from detail view.
type ReturnToSearchMsg struct {
	Query         string
	SelectedIndex int
}

// ClearFlashCmd returns a command that clears the flash after a delay.
func ClearFlashCmd() tea.Cmd {
	return tea.Tick(flashDuration, func(_ time.Time) tea.Msg {
		return ClearFlashMsg{}
	})
}
