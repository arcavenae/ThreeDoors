package tui

import (
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/dispatch"
	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
	tea "github.com/charmbracelet/bubbletea"
)

// ReturnToDoorsMsg is sent when the user wants to go back to the doors view.
type ReturnToDoorsMsg struct{}

// TaskUpdatedMsg is sent when a task has been modified.
type TaskUpdatedMsg struct {
	Task *core.Task
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
	Task *core.Task
}

// ExpandTaskMsg is sent when the user creates a subtask via the expand action.
type ExpandTaskMsg struct {
	ParentTask  *core.Task
	NewTaskText string
}

// FlashMsg triggers a temporary message display.
type FlashMsg struct {
	Text string
}

// ClearFlashMsg clears the flash message.
type ClearFlashMsg struct{}

// SearchResultSelectedMsg is sent when a user selects a task from search results.
type SearchResultSelectedMsg struct {
	Task *core.Task
}

// TaskAddedMsg is sent when the :add command creates a new task.
type TaskAddedMsg struct {
	Task *core.Task
}

// HealthCheckMsg is sent when a health check completes.
type HealthCheckMsg struct {
	Result core.HealthCheckResult
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
	Config *core.ValuesConfig
}

// ShowValuesSetupMsg is sent to open the values setup flow.
type ShowValuesSetupMsg struct{}

// ShowValuesEditMsg is sent to open the values edit flow.
type ShowValuesEditMsg struct{}

// ShowFeedbackMsg is sent to open the door feedback dialog.
type ShowFeedbackMsg struct {
	Task *core.Task
}

// DoorFeedbackMsg is sent when door feedback has been submitted.
type DoorFeedbackMsg struct {
	Task         *core.Task
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
	Task *core.Task
}

// TagCancelledMsg is sent when the user cancels the :tag editor.
type TagCancelledMsg struct{}

// ShowTagViewMsg is sent when :tag is selected from command palette.
type ShowTagViewMsg struct{}

// ShowAvoidancePromptMsg is sent to display the avoidance action prompt.
type ShowAvoidancePromptMsg struct {
	Task *core.Task
}

// AvoidanceActionMsg is sent when the user picks an avoidance prompt action.
type AvoidanceActionMsg struct {
	Task   *core.Task
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
	Task *core.Task
}

// SyncStatusUpdateMsg is sent when a provider's sync status changes.
type SyncStatusUpdateMsg struct {
	ProviderName string
	Phase        core.SyncPhase
	PendingCount int
	ErrorMsg     string
}

// DecomposeStartMsg is sent when the user triggers task decomposition.
type DecomposeStartMsg struct {
	TaskID          string
	TaskDescription string
}

// DecomposeResultMsg is sent when task decomposition completes (success or failure).
type DecomposeResultMsg struct {
	TaskID string
	Result *llm.DecompositionResult
	Err    error
}

// SyncConflictMsg is sent when sync detects conflicts requiring user resolution.
type SyncConflictMsg struct {
	ConflictSet *core.ConflictSet
}

// ConflictResolvedMsg is sent when a user resolves a conflict.
type ConflictResolvedMsg struct {
	ConflictSet *core.ConflictSet
}

// ShowSyncLogMsg is sent to open the sync log view.
type ShowSyncLogMsg struct {
	Entries []core.SyncLogEntry
}

// ShowThemePickerMsg is sent to open the theme picker view.
type ShowThemePickerMsg struct{}

// ThemeSelectedMsg is sent when the user selects a theme from the picker.
type ThemeSelectedMsg struct {
	Name string
}

// ThemeCancelledMsg is sent when the user cancels the theme picker.
type ThemeCancelledMsg struct{}

// WorkerStatusMsg carries polling results from Dispatcher.GetHistory through the Update chain.
type WorkerStatusMsg struct {
	History []dispatch.HistoryEntry
	Err     error
}

// workerPollTickMsg is an internal tick message that triggers worker status polling.
type workerPollTickMsg struct{}

// DeferReturnTickMsg fires periodically to check for expired deferred tasks.
type DeferReturnTickMsg time.Time

// DuplicateDismissedMsg is sent when the user dismisses a duplicate flag (marks as distinct).
type DuplicateDismissedMsg struct {
	Task *core.Task
}

// DuplicateMergedMsg is sent when the user merges a duplicate pair (removes the other copy).
type DuplicateMergedMsg struct {
	Task        *core.Task
	RemovedTask *core.Task
}

// ShowProposalsMsg is sent to open the proposal review view.
type ShowProposalsMsg struct{}

// ProposalApprovedMsg is sent when a proposal is approved and applied.
type ProposalApprovedMsg struct {
	ProposalID string
	TaskID     string
}

// ProposalRejectedMsg is sent when a proposal is rejected.
type ProposalRejectedMsg struct {
	ProposalID string
}

// ProposalBatchApprovedMsg is sent when multiple proposals are approved at once.
type ProposalBatchApprovedMsg struct {
	Count int
}

// ShowHelpMsg is sent to open the dedicated help view.
type ShowHelpMsg struct{}

// ShowDevQueueMsg is sent to open the dev queue view.
type ShowDevQueueMsg struct{}

// DevQueueWorkerCreatedMsg is sent when a worker creation completes (success or failure).
type DevQueueWorkerCreatedMsg struct {
	WorkerName string
	Err        error
}

// DevQueueWorkerRemovedMsg is sent when a worker removal completes (success or failure).
type DevQueueWorkerRemovedMsg struct {
	Err error
}

// DevDispatchRequestMsg is sent when the user confirms dispatching a task.
type DevDispatchRequestMsg struct {
	Task *core.Task
}

// DevDispatchResultMsg is sent when dev dispatch queue write completes.
type DevDispatchResultMsg struct {
	TaskID string
	Err    error
}

// ClearFlashCmd returns a command that clears the flash after a delay.
func ClearFlashCmd() tea.Cmd {
	return tea.Tick(flashDuration, func(_ time.Time) tea.Msg {
		return ClearFlashMsg{}
	})
}
