package tui

import tea "github.com/charmbracelet/bubbletea"

// ViewMode tracks which view is currently active.
type ViewMode int

const (
	ViewDoors ViewMode = iota
	ViewDetail
	ViewMood
	ViewSearch
	ViewHealth
	ViewAddTask
	ViewValuesGoals
	ViewFeedback
	ViewNextSteps
	ViewAvoidancePrompt
	ViewInsights
	ViewOnboarding
	ViewConflict
	ViewSyncLog
	ViewThemePicker
	ViewDevQueue
	ViewProposals
	ViewHelp
	ViewDeferred
	ViewSnooze
	ViewPlanning
	ViewSources
	ViewSourceDetail
	ViewSyncLogDetail
	ViewConnectWizard
	ViewDisconnect
	ViewImport
	ViewBugReport
	ViewBreakdown
	ViewExtract
	ViewOrphaned
	ViewReauth
)

// String returns the human-readable name of the view mode.
func (v ViewMode) String() string {
	switch v {
	case ViewDoors:
		return "Doors"
	case ViewDetail:
		return "Detail"
	case ViewMood:
		return "Mood"
	case ViewSearch:
		return "Search"
	case ViewHealth:
		return "Health"
	case ViewAddTask:
		return "AddTask"
	case ViewValuesGoals:
		return "ValuesGoals"
	case ViewFeedback:
		return "Feedback"
	case ViewNextSteps:
		return "NextSteps"
	case ViewAvoidancePrompt:
		return "AvoidancePrompt"
	case ViewInsights:
		return "Insights"
	case ViewOnboarding:
		return "Onboarding"
	case ViewConflict:
		return "Conflict"
	case ViewSyncLog:
		return "SyncLog"
	case ViewThemePicker:
		return "ThemePicker"
	case ViewDevQueue:
		return "DevQueue"
	case ViewProposals:
		return "Proposals"
	case ViewHelp:
		return "Help"
	case ViewDeferred:
		return "Deferred"
	case ViewSnooze:
		return "Snooze"
	case ViewPlanning:
		return "Planning"
	case ViewSources:
		return "Sources"
	case ViewSourceDetail:
		return "SourceDetail"
	case ViewSyncLogDetail:
		return "SyncLogDetail"
	case ViewConnectWizard:
		return "ConnectWizard"
	case ViewDisconnect:
		return "Disconnect"
	case ViewImport:
		return "Import"
	case ViewBugReport:
		return "BugReport"
	case ViewBreakdown:
		return "Breakdown"
	case ViewExtract:
		return "Extract"
	case ViewOrphaned:
		return "Orphaned"
	case ViewReauth:
		return "Reauth"
	default:
		return "Unknown"
	}
}

// setViewMode transitions to a new view and records a breadcrumb if the view changed.
func (m *MainModel) setViewMode(v ViewMode) {
	if v != m.viewMode {
		m.breadcrumbs.Record(v.String(), "view:"+v.String())
	}
	m.viewMode = v
}

// goBack returns to the previous view.
func (m *MainModel) goBack() {
	m.setViewMode(m.previousView)
}

// updateOverlay handles key events when the keybinding overlay is visible.
// It intercepts all keys — only ?, esc, and scroll keys have behavior; the rest are consumed.
func (m *MainModel) updateOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc":
		m.showKeybindingOverlay = false
		m.keybindingOverlay = nil
		return m, nil
	}
	if m.keybindingOverlay != nil {
		cmd := m.keybindingOverlay.Update(msg)
		return m, cmd
	}
	return m, nil
}

// contentHeight returns the available height for view content, accounting for
// the keybinding bar when it is visible and the terminal is tall enough to show it.
func (m *MainModel) contentHeight() int {
	// Keybinding bar only shows in non-door views when key hints are on
	if m.showKeyHints && m.viewMode != ViewDoors && m.height >= barHeightHidden {
		return m.height - 2 // separator + bar
	}
	return m.height
}

// currentViewContent returns the rendered content for the currently active view.
func (m *MainModel) currentViewContent() (view string, showValuesFooter bool) {
	if v, svf, handled := m.taskViewContent(); handled {
		return v, svf
	}
	switch m.viewMode {
	case ViewDetail:
		if m.detailView != nil {
			view = m.detailView.View()
		}
		showValuesFooter = true
	case ViewMood:
		if m.moodView != nil {
			view = m.moodView.View()
		}
	case ViewSearch:
		if m.searchView != nil {
			view = m.searchView.View()
		}
		showValuesFooter = true
	case ViewHealth:
		if m.healthView != nil {
			view = m.healthView.View()
		}
	case ViewInsights:
		if m.insightsView != nil {
			view = m.insightsView.View()
		}
	case ViewOrphaned:
		if m.orphanedView != nil {
			view = m.orphanedView.View()
		}
	case ViewSources:
		if m.sourcesView != nil {
			view = m.sourcesView.View()
		}
	case ViewSourceDetail:
		if m.sourceDetailView != nil {
			view = m.sourceDetailView.View()
		}
	case ViewSyncLogDetail:
		if m.syncLogDetailView != nil {
			view = m.syncLogDetailView.View()
		}
	case ViewConnectWizard:
		if m.connectWizard != nil {
			view = m.connectWizard.View()
		}
	case ViewDisconnect:
		if m.disconnectDialog != nil {
			view = m.disconnectDialog.View()
		}
	case ViewReauth:
		if m.reauthDialog != nil {
			view = m.reauthDialog.View()
		}
	case ViewBugReport:
		if m.bugReportView != nil {
			view = m.bugReportView.View()
		}
	case ViewValuesGoals:
		if m.valuesView != nil {
			view = m.valuesView.View()
		}
	case ViewFeedback:
		if m.feedbackView != nil {
			view = m.feedbackView.View()
		}
	case ViewNextSteps:
		if m.nextStepsView != nil {
			view = m.nextStepsView.View()
		}
		showValuesFooter = true
	case ViewAvoidancePrompt:
		if m.avoidancePromptView != nil {
			view = m.avoidancePromptView.View()
		}
	case ViewOnboarding:
		if m.onboardingView != nil {
			view = m.onboardingView.View()
		}
	case ViewConflict:
		if m.conflictView != nil {
			view = m.conflictView.View()
		}
	case ViewSyncLog:
		if m.syncLogView != nil {
			view = m.syncLogView.View()
		}
	case ViewThemePicker:
		if m.themePickerView != nil {
			view = m.themePickerView.View()
		}
	case ViewDevQueue:
		if m.devQueueView != nil {
			view = m.devQueueView.View()
		}
	case ViewProposals:
		if m.proposalsView != nil {
			view = m.proposalsView.View()
		}
	case ViewHelp:
		if m.helpView != nil {
			view = m.helpView.View()
		}
	default:
		view = m.doorsView.View()
		showValuesFooter = true
	}
	return view, showValuesFooter
}
