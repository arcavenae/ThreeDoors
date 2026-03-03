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

// HealthCheckMsg is sent when a health check completes.
type HealthCheckMsg struct {
	Result tasks.HealthCheckResult
}

// SearchClosedMsg is sent when the user exits search mode.
type SearchClosedMsg struct{}

// AddTaskPromptMsg is sent when :add is typed without text to open inline add mode.
type AddTaskPromptMsg struct{}

// AddTaskWithContextPromptMsg is sent when :add-ctx or :add --why is used
// to open the multi-step context capture flow.
type AddTaskWithContextPromptMsg struct {
	PrefilledText string
}

// ValuesSavedMsg is sent when values/goals have been saved.
type ValuesSavedMsg struct {
	Config *tasks.ValuesConfig
}

// ShowValuesSetupMsg is sent to open the values setup flow.
type ShowValuesSetupMsg struct{}

// ShowValuesEditMsg is sent to open the values edit flow.
type ShowValuesEditMsg struct{}

// ShowFeedbackMsg is sent to open the door feedback dialog.
type ShowFeedbackMsg struct {
	Task *tasks.Task
}

// DoorFeedbackMsg is sent when door feedback has been submitted.
type DoorFeedbackMsg struct {
	Task         *tasks.Task
	FeedbackType string
	Comment      string
}

// RequestQuitMsg is sent when the user requests to quit.
// MainModel intercepts this to show the improvement prompt if criteria are met.
type RequestQuitMsg struct{}

// ImprovementSubmittedMsg is sent when the user submits an improvement suggestion.
type ImprovementSubmittedMsg struct {
	Text string
}

// ImprovementSkippedMsg is sent when the user skips the improvement prompt.
type ImprovementSkippedMsg struct{}

// ShowNextStepsMsg is sent to open the contextual next-steps view.
type ShowNextStepsMsg struct {
	Context string // what triggered it: "completed", "added"
}

// NextStepSelectedMsg is sent when the user picks a next-step option.
type NextStepSelectedMsg struct {
	Action string // action identifier: "doors", "add", "mood", "search", "stats"
}

// NextStepDismissedMsg is sent when the user dismisses the next-steps view.
type NextStepDismissedMsg struct{}

// TagUpdatedMsg is sent when the :tag command finishes editing a task's categories.
type TagUpdatedMsg struct {
	Task *tasks.Task
}

// TagCancelledMsg is sent when the user cancels the :tag editor.
type TagCancelledMsg struct{}

// ShowTagViewMsg is sent when :tag is selected from command palette.
type ShowTagViewMsg struct{}

// ShowAvoidancePromptMsg is sent to display the avoidance action prompt.
type ShowAvoidancePromptMsg struct {
	Task *tasks.Task
}

// AvoidanceActionMsg is sent when the user picks an avoidance prompt action.
type AvoidanceActionMsg struct {
	Task   *tasks.Task
	Action string // "reconsider", "breakdown", "defer", "archive"
}

// ReturnToSearchMsg is sent to restore search view from detail view.
type ReturnToSearchMsg struct {
	Query         string
	SelectedIndex int
}

// ShowInsightsMsg is sent to open the insights dashboard view.
type ShowInsightsMsg struct{}

// NavigateToLinkedMsg is sent when user selects a linked task to navigate to.
type NavigateToLinkedMsg struct {
	Task *tasks.Task
}

// SyncStatusUpdateMsg is sent when a provider's sync status changes.
type SyncStatusUpdateMsg struct {
	ProviderName string
	Phase        tasks.SyncPhase
	PendingCount int
	ErrorMsg     string
}

// ClearFlashCmd returns a command that clears the flash after a delay.
func ClearFlashCmd() tea.Cmd {
	return tea.Tick(flashDuration, func(_ time.Time) tea.Msg {
		return ClearFlashMsg{}
	})
}
