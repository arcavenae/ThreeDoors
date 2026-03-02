package tui

import (
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

// SelectDoorMsg is sent when a user selects a door to view its details.
type SelectDoorMsg struct {
	DoorIndex int
	Task      *tasks.Task
}

// ReturnToDoorsMsg is sent when the user wants to go back to the doors view.
type ReturnToDoorsMsg struct{}

// TaskUpdatedMsg is sent when a task has been modified.
type TaskUpdatedMsg struct {
	Task *tasks.Task
}

// RefreshDoorsMsg is sent to trigger a new random door selection.
type RefreshDoorsMsg struct{}

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

// ClearFlashCmd returns a command that clears the flash after a delay.
func ClearFlashCmd() tea.Cmd {
	return tea.Tick(flashDuration, func(_ time.Time) tea.Msg {
		return ClearFlashMsg{}
	})
}
