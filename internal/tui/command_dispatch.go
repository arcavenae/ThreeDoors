package tui

import (
	"time"

	"github.com/arcaven/ThreeDoors/internal/dispatch"
	tea "github.com/charmbracelet/bubbletea"
)

// workerPollInterval is the interval between worker status polling ticks.
const workerPollInterval = 30 * time.Second

// deferReturnInterval is the interval between deferred task auto-return checks.
const deferReturnInterval = 1 * time.Minute

// deferReturnTickCmd returns a tea.Cmd that fires a DeferReturnTickMsg after the interval.
func deferReturnTickCmd() tea.Cmd {
	return tea.Tick(deferReturnInterval, func(t time.Time) tea.Msg {
		return DeferReturnTickMsg(t)
	})
}

// workerPollTickCmd returns a tea.Cmd that fires a workerPollTickMsg after the poll interval.
func workerPollTickCmd() tea.Cmd {
	return tea.Tick(workerPollInterval, func(_ time.Time) tea.Msg {
		return workerPollTickMsg{}
	})
}

// mapHistoryStatus maps a multiclaude history status to a QueueItemStatus.
func mapHistoryStatus(status string) dispatch.QueueItemStatus {
	switch status {
	case "completed", "open", "merged":
		return dispatch.QueueItemCompleted
	case "failed", "no-pr":
		return dispatch.QueueItemFailed
	default:
		return dispatch.QueueItemDispatched
	}
}

// mapPRStatus maps a multiclaude history status to a PR status string for display.
func mapPRStatus(status string) string {
	switch status {
	case "open":
		return "open"
	case "merged":
		return "merged"
	case "completed":
		return "open"
	default:
		return status
	}
}

// buildBarContext constructs a BarContext with sub-mode awareness for the
// keybinding bar, allowing it to show context-appropriate keys.
func (m *MainModel) buildBarContext() BarContext {
	ctx := BarContext{Mode: m.viewMode}
	switch m.viewMode {
	case ViewDoors:
		ctx.DoorSelected = m.doorsView != nil && m.doorsView.selectedDoorIndex >= 0
	case ViewDetail:
		if m.detailView != nil {
			ctx.DetailMode = m.detailView.mode
		}
	case ViewSearch:
		ctx.CommandMode = m.searchView != nil && m.searchView.isCommandMode
	}
	return ctx
}

// isTextInputActive returns true when the current view has an active text input
// field where 'q' should be treated as text, not as a quit command.
func (m *MainModel) isTextInputActive() bool {
	if m.isTaskTextInputActive() {
		return true
	}
	return m.isAuxiliaryTextInputActive()
}
