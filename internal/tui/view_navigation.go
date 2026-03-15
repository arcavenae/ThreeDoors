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
	ViewHistory
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
	case ViewHistory:
		return "History"
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
	if v, svf, handled := m.sourceViewContent(); handled {
		return v, svf
	}
	if v, svf, handled := m.auxiliaryViewContent(); handled {
		return v, svf
	}
	switch m.viewMode {
	case ViewDetail:
		if m.detailView != nil {
			view = m.detailView.View()
		}
		showValuesFooter = true
	case ViewSearch:
		if m.searchView != nil {
			view = m.searchView.View()
		}
		showValuesFooter = true
	default:
		view = m.doorsView.View()
		showValuesFooter = true
	}
	return view, showValuesFooter
}
