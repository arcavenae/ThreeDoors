package tui

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/dispatch"
	"github.com/arcavenae/ThreeDoors/internal/intelligence/services"
	"github.com/arcavenae/ThreeDoors/internal/tui/themes"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// handleAuxiliaryViewMessage handles Update() messages for auxiliary views
// (Help, BugReport, ThemePicker, Health, Insights, Mood, Feedback, ValuesGoals,
// Orphaned, Conflict, Proposals, DevQueue, NextSteps, Avoidance, Onboarding,
// History, and related domain messages).
// Returns (model, cmd, handled). If handled is false, the caller should
// continue processing in the main Update switch.
func (m *MainModel) handleAuxiliaryViewMessage(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case HealthCheckMsg:
		m.healthView = NewHealthView(msg.Result)
		m.healthView.SetWidth(m.width)
		m.healthView.SetInlineHints(m.resolveHints())
		m.previousView = m.viewMode
		m.setViewMode(ViewHealth)
		return m, nil, true

	case ShowInsightsMsg:
		var activeTheme *themes.DoorTheme
		if dv := m.doorsView; dv != nil {
			activeTheme = dv.Theme()
		}
		m.insightsView = NewInsightsView(m.patternAnalyzer, m.completionCounter, activeTheme, m.milestoneChecker)
		m.insightsView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.setViewMode(ViewInsights)
		animCmd := m.insightsView.StartAnimation()

		// Check for milestone celebrations on view entry
		var totalTasks, currentStreak, sessionCount int
		if m.patternAnalyzer != nil {
			totalTasks = m.patternAnalyzer.GetTotalCompleted()
			sessionCount = m.patternAnalyzer.GetSessionCount()
		}
		if m.completionCounter != nil {
			currentStreak = m.completionCounter.GetStreak()
		}
		milestoneCmd := m.insightsView.CheckAndShowMilestone(totalTasks, currentStreak, sessionCount)
		cmd := tea.Batch(animCmd, milestoneCmd)
		return m, cmd, true

	case ShowHistoryMsg:
		if m.completionReader != nil {
			records, err := m.completionReader.Read(context.Background())
			if err != nil {
				m.flash = "Failed to load history"
				return m, ClearFlashCmd(), true
			}
			m.historyView = NewHistoryView(records, nil)
			m.historyView.SetWidth(m.width)
			m.historyView.SetHeight(m.contentHeight())
			m.previousView = m.viewMode
			m.setViewMode(ViewHistory)
		}
		return m, nil, true

	case ShowOrphanedMsg:
		m.orphanedView = NewOrphanedView(m.pool)
		m.orphanedView.SetWidth(m.width)
		m.orphanedView.SetHeight(m.height)
		m.previousView = m.viewMode
		m.setViewMode(ViewOrphaned)
		return m, nil, true

	case OrphanedTaskActionMsg:
		return m.handleOrphanedAction(msg)

	case MoodCapturedMsg:
		if m.tracker != nil {
			m.tracker.RecordMood(msg.Mood, msg.CustomText)
		}
		m.setViewMode(ViewDoors)
		m.moodView = nil
		m.flash = fmt.Sprintf("Mood logged: %s", msg.Mood)
		return m, ClearFlashCmd(), true

	case ShowMoodMsg:
		m.moodView = NewMoodView()
		m.moodView.SetWidth(m.width)
		m.moodView.SetInlineHints(m.resolveHints())
		m.setViewMode(ViewMood)
		return m, nil, true

	case ShowFeedbackMsg:
		m.feedbackView = NewFeedbackView(msg.Task)
		m.feedbackView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.setViewMode(ViewFeedback)
		return m, nil, true

	case DoorFeedbackMsg:
		if m.tracker != nil {
			m.tracker.RecordDoorFeedback(msg.Task.ID, msg.FeedbackType, msg.Comment)
		}
		if msg.FeedbackType == "needs-breakdown" {
			msg.Task.AddNote("Flagged: needs breakdown")
			if err := m.saveTasks(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
			}
		}
		m.feedbackView = nil
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		m.flash = "Feedback recorded"
		return m, ClearFlashCmd(), true

	case ShowValuesSetupMsg:
		m.valuesView = NewValuesSetupView(m.valuesConfig)
		m.valuesView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.setViewMode(ViewValuesGoals)
		return m, nil, true

	case ShowValuesEditMsg:
		m.valuesView = NewValuesEditView(m.valuesConfig)
		m.valuesView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.setViewMode(ViewValuesGoals)
		return m, nil, true

	case ValuesSavedMsg:
		m.valuesConfig = msg.Config
		if path, err := core.GetValuesConfigPath(); err == nil {
			if saveErr := core.SaveValuesConfig(path, msg.Config); saveErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save values config: %v\n", saveErr)
			}
		}
		m.valuesView = nil
		m.flash = "Values saved"
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		return m, ClearFlashCmd(), true

	case ShowNextStepsMsg:
		m.nextStepsView = NewNextStepsView(msg.Context, m.pool, m.completionCounter)
		m.nextStepsView.SetWidth(m.width)
		m.setViewMode(ViewNextSteps)
		return m, nil, true

	case NextStepSelectedMsg:
		m.nextStepsView = nil
		switch msg.Action {
		case "doors":
			m.setViewMode(ViewDoors)
			m.doorsView.RefreshDoors()
			m.doorsView.RotateFooterMessage()
		case "add":
			return m, func() tea.Msg { return AddTaskPromptMsg{} }, true
		case "mood":
			return m, func() tea.Msg { return ShowMoodMsg{} }, true
		case "search":
			m.searchView = m.newSearchView()
			m.searchView.SetWidth(m.width)
			m.setViewMode(ViewSearch)
			m.previousView = ViewDoors
		case "stats":
			m.searchView = m.newSearchView()
			m.searchView.SetWidth(m.width)
			m.searchView.textInput.SetValue(":stats")
			m.searchView.checkCommandMode()
			m.setViewMode(ViewSearch)
			m.previousView = ViewDoors
		default:
			m.setViewMode(ViewDoors)
			m.doorsView.RefreshDoors()
		}
		return m, nil, true

	case NextStepDismissedMsg:
		m.nextStepsView = nil
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		m.doorsView.RotateFooterMessage()
		return m, nil, true

	case ShowAvoidancePromptMsg:
		m.avoidancePromptView = NewAvoidancePromptView(msg.Task, m.doorsView.avoidanceMap[msg.Task.Text])
		m.avoidancePromptView.SetWidth(m.width)
		m.promptedTasks[msg.Task.Text] = true
		m.previousView = m.viewMode
		m.setViewMode(ViewAvoidancePrompt)
		return m, nil, true

	case AvoidanceActionMsg:
		return m.handleAvoidanceAction(msg)

	case OnboardingCompletedMsg:
		return m.handleOnboardingCompleted(msg)

	case ShowThemePickerMsg:
		currentTheme := ""
		if dv := m.doorsView; dv != nil && dv.theme != nil {
			currentTheme = dv.theme.Name
		}
		reg := m.doorsView.themeRegistry
		if reg == nil {
			reg = themes.NewDefaultRegistry()
		}
		m.themePickerView = NewThemePicker(reg, currentTheme)
		m.themePickerView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.setViewMode(ViewThemePicker)
		return m, nil, true

	case ShowSeasonalPickerMsg:
		currentTheme := ""
		if dv := m.doorsView; dv != nil && dv.theme != nil {
			currentTheme = dv.theme.Name
		}
		reg := m.doorsView.themeRegistry
		if reg == nil {
			reg = themes.NewDefaultRegistry()
		}
		m.themePickerView = NewSeasonalThemePicker(reg, currentTheme)
		m.themePickerView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewThemePicker
		return m, nil, true

	case ThemeSelectedMsg:
		seasonal := m.themePickerView != nil && m.themePickerView.IsSeasonal()
		m.doorsView.SetThemeByName(msg.Name)
		m.themePickerView = nil
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		m.flash = fmt.Sprintf("Theme changed to %s", msg.Name)
		if seasonal {
			return m, ClearFlashCmd(), true
		}
		return m, tea.Batch(ClearFlashCmd(), m.saveThemeCmd(msg.Name)), true

	case ThemeCancelledMsg:
		m.themePickerView = nil
		m.setViewMode(ViewDoors)
		return m, nil, true

	case SyncConflictMsg:
		cv := NewConflictView(msg.ConflictSet, m.syncLog)
		cv.SetWidth(m.width)
		m.conflictView = cv
		m.previousView = m.viewMode
		m.setViewMode(ViewConflict)
		return m, nil, true

	case ConflictResolvedMsg:
		// Apply resolutions to the pool
		resolutions := msg.ConflictSet.Resolutions()
		for _, r := range resolutions {
			if r.Winner == "both" {
				// "Keep both" — keep local as-is, no update needed
				continue
			}
			m.pool.UpdateTask(r.WinningTask)
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save after conflict resolution: %v\n", err)
		}
		m.conflictView = nil
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		m.flash = fmt.Sprintf("%d conflict(s) resolved", len(resolutions))
		return m, ClearFlashCmd(), true

	case DuplicateDismissedMsg:
		m.refreshDuplicates()
		m.flash = "Duplicate flag dismissed"
		m.detailView = nil
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		return m, ClearFlashCmd(), true

	case DuplicateMergedMsg:
		// Remove the duplicate task
		if err := m.provider.DeleteTask(msg.RemovedTask.ID); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to delete merged duplicate: %v\n", err)
		}
		m.pool.RemoveTask(msg.RemovedTask.ID)
		m.refreshDuplicates()
		m.flash = "Duplicate merged"
		m.detailView = nil
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		return m, ClearFlashCmd(), true

	case ShowProposalsMsg:
		if m.proposalStore == nil {
			m.flash = "No proposal store available"
			return m, ClearFlashCmd(), true
		}
		m.proposalsView = NewProposalsView(m.proposalStore, m.pool, m.provider)
		m.proposalsView.SetWidth(m.width)
		m.proposalsView.SetHeight(m.contentHeight())
		m.previousView = m.viewMode
		m.setViewMode(ViewProposals)
		return m, nil, true

	case ProposalApprovedMsg:
		m.flash = fmt.Sprintf("Approved proposal for task %s", msg.TaskID)
		m.doorsView.SetPendingProposals(PendingProposalCount(m.proposalStore))
		return m, ClearFlashCmd(), true

	case ProposalRejectedMsg:
		m.flash = "Proposal rejected"
		m.doorsView.SetPendingProposals(PendingProposalCount(m.proposalStore))
		return m, ClearFlashCmd(), true

	case ProposalBatchApprovedMsg:
		m.flash = fmt.Sprintf("Approved %d proposals", msg.Count)
		m.doorsView.SetPendingProposals(PendingProposalCount(m.proposalStore))
		return m, ClearFlashCmd(), true

	case ShowHelpMsg:
		hv := NewHelpView()
		hv.SetWidth(m.width)
		hv.SetHeight(m.height)
		m.helpView = hv
		m.previousView = m.viewMode
		m.setViewMode(ViewHelp)
		return m, nil, true

	case ShowBugReportMsg:
		themeName := ""
		if m.doorsView != nil {
			themeName = m.doorsView.baseThemeName
		}
		taskCount := m.pool.Count()
		providerCount := 0
		if m.connMgr != nil {
			providerCount = m.connMgr.Count()
		}
		sessionDur := time.Duration(m.tracker.GetMetricsSnapshot().DurationSeconds() * float64(time.Second))
		env := CollectEnvironment(m.width, m.height, m.viewMode.String(), themeName, taskCount, providerCount, sessionDur)
		breadcrumbs := m.breadcrumbs.Format()
		bv := NewBugReportView(env, breadcrumbs)
		bv.SetWidth(m.width)
		bv.SetHeight(m.height)
		m.bugReportView = bv
		m.previousView = m.viewMode
		m.setViewMode(ViewBugReport)
		return m, nil, true

	case ShowDevQueueMsg:
		if m.devQueue == nil || m.dispatcher == nil {
			m.flash = "Dev queue not available"
			return m, ClearFlashCmd(), true
		}
		m.devQueueView = NewDevQueueView(m.devQueue, m.dispatcher)
		m.devQueueView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.setViewMode(ViewDevQueue)
		return m, nil, true

	case DevDispatchRequestMsg:
		model, cmd := m.handleDevDispatch(msg.Task)
		return model, cmd, true

	case DevDispatchResultMsg:
		if msg.Err != nil {
			m.flash = fmt.Sprintf("Dispatch failed: %s", msg.Err.Error())
		} else {
			m.flash = "Task queued for dev dispatch"
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after dispatch: %v\n", err)
		}
		return m, ClearFlashCmd(), true

	case ShowLLMStatusMsg:
		return m, runLLMStatusCmd(), true

	case LLMStatusResultMsg:
		m.flash = msg.Text
		return m, ClearFlashCmd(), true

	case FlashMsg:
		m.flash = msg.Text
		return m, ClearFlashCmd(), true

	case InlineHintsToggleMsg:
		switch msg.Arg {
		case "on":
			m.showKeyHints = true
			m.flash = "Key hints enabled"
		case "off":
			m.showKeyHints = false
			m.flash = "Key hints disabled"
		default:
			m.showKeyHints = !m.showKeyHints
			if m.showKeyHints {
				m.flash = "Key hints enabled"
			} else {
				m.flash = "Key hints disabled"
			}
		}
		m.doorsView.SetShowKeyHints(m.showKeyHints)
		m.doorsView.SetHeight(m.contentHeight())
		m.setViewMode(ViewDoors)
		return m, tea.Batch(ClearFlashCmd(), m.saveKeyHintsCmd(m.showKeyHints)), true

	case DecomposeStartMsg:
		if m.decomposing {
			m.flash = "Decomposition already in progress"
			return m, ClearFlashCmd(), true
		}
		m.decomposing = true
		m.flash = "Decomposing task..."
		return m, m.runDecompose(msg.TaskID, msg.TaskDescription), true

	case DecomposeResultMsg:
		m.decomposing = false
		if msg.Err != nil {
			m.flash = fmt.Sprintf("Decompose failed: %s", msg.Err.Error())
			return m, ClearFlashCmd(), true
		}
		m.flash = fmt.Sprintf("Decomposed into %d stories", len(msg.Result.Stories))
		return m, ClearFlashCmd(), true

	case EnrichStartMsg:
		if m.enriching {
			m.flash = "Enrichment already in progress"
			return m, ClearFlashCmd(), true
		}
		m.enriching = true
		m.flash = "Enriching task..."
		return m, m.runEnrich(msg.TaskID, msg.TaskText), true

	case EnrichResultMsg:
		m.enriching = false
		// Forward to the detail view for display
		if m.detailView != nil {
			cmd := m.detailView.Update(msg)
			return m, cmd, true
		}
		if msg.Err != nil {
			m.flash = fmt.Sprintf("Enrich failed: %s", msg.Err.Error())
		}
		return m, ClearFlashCmd(), true

	case EnrichAcceptMsg:
		task := m.pool.GetTask(msg.TaskID)
		if task != nil {
			task.Text = msg.EnrichedText
			if msg.Context != "" {
				task.Context = msg.Context
			}
			switch {
			case msg.Effort <= 1:
				task.Effort = core.EffortQuickWin
			case msg.Effort <= 3:
				task.Effort = core.EffortMedium
			case msg.Effort <= 5:
				task.Effort = core.EffortDeepWork
			}
			task.UpdatedAt = time.Now().UTC()
			m.flash = "Task enriched!"
			_ = m.saveTasks()
		} else {
			m.flash = "Task not found"
		}
		m.setViewMode(ViewDoors)
		return m, ClearFlashCmd(), true

	case EnrichCommandMsg:
		// :enrich command from search view — route to current detail view
		if m.detailView != nil && m.detailView.task != nil {
			if m.enricher == nil {
				m.flash = "LLM not configured — enrichment unavailable"
				return m, ClearFlashCmd(), true
			}
			desc := m.detailView.task.Text
			taskID := m.detailView.task.ID
			m.detailView.mode = DetailModeEnrichLoading
			m.enriching = true
			m.flash = "Enriching task..."
			m.setViewMode(ViewDetail)
			return m, m.runEnrich(taskID, desc), true
		}
		m.flash = "Open a task detail view first, then use :enrich"
		return m, ClearFlashCmd(), true

	case DependencyUnblockedMsg:
		m.doorsView.RefreshDoors()
		if m.tracker != nil {
			for _, task := range msg.UnblockedTasks {
				m.tracker.RecordDependencyUnblocked(task.ID, msg.CompletedDepID)
			}
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after dependency unblock: %v\n", err)
		}
		return m, nil, true

	case DependencyAddedMsg:
		m.doorsView.RefreshDoors()
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after dependency added: %v\n", err)
		}
		m.flash = "Dependency added"
		return m, ClearFlashCmd(), true

	case DependencyRemovedMsg:
		m.doorsView.RefreshDoors()
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after dependency removed: %v\n", err)
		}
		m.flash = "Dependency removed"
		return m, ClearFlashCmd(), true

	case DeferReturnTickMsg:
		returned := core.CheckDeferredReturnsWithTracker(m.pool, m.tracker)
		if returned > 0 {
			m.doorsView.RefreshDoors()
			if err := m.saveTasks(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save tasks after defer return: %v\n", err)
			}
		}
		return m, deferReturnTickCmd(), true

	case workerPollTickMsg:
		if m.dispatcher == nil || !m.hasDispatchedItems() {
			m.pollingActive = false
			return m, nil, true
		}
		return m, m.pollWorkerStatusCmd(), true

	case WorkerStatusMsg:
		cmd := m.handleWorkerStatus(msg)
		return m, cmd, true

	case spinner.TickMsg:
		if m.syncSpinner != nil {
			cmd := m.syncSpinner.Update(msg)
			return m, cmd, true
		}
		return m, nil, true

	case SyncStatusUpdateMsg:
		if m.syncTracker != nil {
			switch msg.Phase {
			case core.SyncPhaseSynced:
				m.syncTracker.SetSynced(msg.ProviderName)
				if m.syncSpinner != nil {
					m.syncSpinner.Stop()
				}
			case core.SyncPhaseSyncing:
				m.syncTracker.SetSyncing(msg.ProviderName)
				if m.syncSpinner != nil {
					m.syncSpinner.Start(msg.ProviderName)
					return m, m.syncSpinner.Tick(), true
				}
			case core.SyncPhasePending:
				m.syncTracker.SetPending(msg.ProviderName, msg.PendingCount)
			case core.SyncPhaseError:
				m.syncTracker.SetError(msg.ProviderName, msg.ErrorMsg)
				if m.syncSpinner != nil {
					m.syncSpinner.Stop()
				}
			}
		}
		return m, nil, true
	}

	return m, nil, false
}

// handleOrphanedAction processes orphaned task keep/delete actions.
func (m *MainModel) handleOrphanedAction(msg OrphanedTaskActionMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.Action {
	case "keep":
		kept := m.pool.KeepOrphanedTask(msg.TaskID)
		if kept != nil {
			if err := m.saveTasks(); err != nil {
				log.Printf("save after orphan keep: %v", err)
			}
			if m.orphanedView != nil {
				m.orphanedView.refreshTasks()
			}
			if len(m.pool.GetOrphanedTasks()) == 0 {
				m.setViewMode(ViewDoors)
				return m, func() tea.Msg {
					return FlashMsg{Text: fmt.Sprintf("Kept '%s' as local task. No more orphaned tasks.", kept.Text)}
				}, true
			}
			return m, func() tea.Msg {
				return FlashMsg{Text: fmt.Sprintf("Kept '%s' as local task.", kept.Text)}
			}, true
		}
	case "delete":
		task := m.pool.GetTask(msg.TaskID)
		if task != nil {
			text := task.Text
			m.pool.RemoveTask(msg.TaskID)
			if err := m.saveTasks(); err != nil {
				log.Printf("save after orphan delete: %v", err)
			}
			if m.orphanedView != nil {
				m.orphanedView.refreshTasks()
			}
			if len(m.pool.GetOrphanedTasks()) == 0 {
				m.setViewMode(ViewDoors)
				return m, func() tea.Msg {
					return FlashMsg{Text: fmt.Sprintf("Deleted '%s'. No more orphaned tasks.", text)}
				}, true
			}
			return m, func() tea.Msg {
				return FlashMsg{Text: fmt.Sprintf("Deleted '%s'.", text)}
			}, true
		}
	}
	return m, nil, true
}

// handleAvoidanceAction processes avoidance prompt responses.
func (m *MainModel) handleAvoidanceAction(msg AvoidanceActionMsg) (tea.Model, tea.Cmd, bool) {
	m.avoidancePromptView = nil
	switch msg.Action {
	case "reconsider":
		if err := msg.Task.UpdateStatus(core.StatusInProgress); err == nil {
			if err := m.saveTasks(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
			}
		}
		m.detailView = m.newDetailView(msg.Task)
		m.setViewMode(ViewDetail)
		m.flash = "Taking it on!"
		return m, ClearFlashCmd(), true
	case "breakdown":
		m.detailView = m.newDetailView(msg.Task)
		m.setViewMode(ViewDetail)
		return m, nil, true
	case "defer":
		if err := msg.Task.UpdateStatus(core.StatusDeferred); err == nil {
			if m.tracker != nil {
				m.tracker.RecordSnooze(msg.Task.ID, msg.Task.DeferUntil, "someday")
			}
			if err := m.saveTasks(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
			}
		}
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		m.flash = "Task set aside for later"
		return m, ClearFlashCmd(), true
	case "archive":
		if err := msg.Task.UpdateStatus(core.StatusArchived); err == nil {
			if err := m.provider.MarkComplete(msg.Task.ID); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to archive task: %v\n", err)
			}
			m.pool.RemoveTask(msg.Task.ID)
		}
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		m.flash = "Task archived"
		return m, ClearFlashCmd(), true
	default:
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
	}
	return m, nil, true
}

// handleOnboardingCompleted processes onboarding completion.
func (m *MainModel) handleOnboardingCompleted(msg OnboardingCompletedMsg) (tea.Model, tea.Cmd, bool) {
	m.onboardingView = nil
	m.setViewMode(ViewDoors)
	// Save values if provided
	if len(msg.Values) > 0 {
		m.valuesConfig = &core.ValuesConfig{Values: msg.Values}
		if path, err := core.GetValuesConfigPath(); err == nil {
			if saveErr := core.SaveValuesConfig(path, m.valuesConfig); saveErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save values config: %v\n", saveErr)
			}
		}
	}
	// Import tasks if provided
	if len(msg.ImportedTasks) > 0 {
		for _, t := range msg.ImportedTasks {
			m.pool.AddTask(t)
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save imported tasks: %v\n", err)
		}
		m.flash = fmt.Sprintf("%d tasks imported!", len(msg.ImportedTasks))
	}
	m.doorsView.RefreshDoors()
	// Persist onboarding state
	if configDir, err := core.GetConfigDirPath(); err == nil {
		if markErr := core.MarkOnboardingComplete(configDir); markErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save onboarding state: %v\n", markErr)
		}
	}
	var cmd tea.Cmd
	if m.flash != "" {
		cmd = ClearFlashCmd()
	}
	return m, cmd, true
}

// resizeAuxiliaryViews updates dimensions for auxiliary views on window resize.
func (m *MainModel) resizeAuxiliaryViews(width, height, contentH int) {
	if m.detailView != nil {
		m.detailView.SetWidth(width)
		m.detailView.SetHeight(contentH)
	}
	if m.moodView != nil {
		m.moodView.SetWidth(width)
		m.moodView.SetHeight(contentH)
	}
	if m.searchView != nil {
		m.searchView.SetWidth(width)
		m.searchView.SetHeight(height)
	}
	if m.healthView != nil {
		m.healthView.SetWidth(width)
		m.healthView.SetHeight(contentH)
	}
	if m.insightsView != nil {
		m.insightsView.SetWidth(width)
		m.insightsView.SetHeight(contentH)
	}
	if m.valuesView != nil {
		m.valuesView.SetWidth(width)
		m.valuesView.SetHeight(contentH)
	}
	if m.feedbackView != nil {
		m.feedbackView.SetWidth(width)
		m.feedbackView.SetHeight(contentH)
	}
	if m.nextStepsView != nil {
		m.nextStepsView.SetWidth(width)
		m.nextStepsView.SetHeight(contentH)
	}
	if m.avoidancePromptView != nil {
		m.avoidancePromptView.SetWidth(width)
		m.avoidancePromptView.SetHeight(contentH)
	}
	if m.onboardingView != nil {
		m.onboardingView.SetWidth(width)
		m.onboardingView.SetHeight(contentH)
	}
	if m.conflictView != nil {
		m.conflictView.SetWidth(width)
		m.conflictView.SetHeight(contentH)
	}
	if m.syncLogView != nil {
		m.syncLogView.SetWidth(width)
		m.syncLogView.SetHeight(height)
	}
	if m.themePickerView != nil {
		m.themePickerView.SetWidth(width)
		m.themePickerView.SetHeight(contentH)
	}
	if m.devQueueView != nil {
		m.devQueueView.SetWidth(width)
		m.devQueueView.SetHeight(contentH)
	}
	if m.proposalsView != nil {
		m.proposalsView.SetWidth(width)
		m.proposalsView.SetHeight(contentH)
	}
	if m.helpView != nil {
		m.helpView.SetWidth(width)
		m.helpView.SetHeight(height)
	}
	if m.orphanedView != nil {
		m.orphanedView.SetWidth(width)
		m.orphanedView.SetHeight(height)
	}
	if m.bugReportView != nil {
		m.bugReportView.SetWidth(width)
		m.bugReportView.SetHeight(height)
	}
	if m.historyView != nil {
		m.historyView.SetWidth(width)
		m.historyView.SetHeight(contentH)
	}
}

// auxiliaryViewContent returns the rendered content for auxiliary views.
// Returns (view, showValuesFooter, handled).
func (m *MainModel) auxiliaryViewContent() (string, bool, bool) {
	switch m.viewMode {
	case ViewMood:
		if m.moodView != nil {
			return m.moodView.View(), false, true
		}
		return "", false, true
	case ViewHealth:
		if m.healthView != nil {
			return m.healthView.View(), false, true
		}
		return "", false, true
	case ViewInsights:
		if m.insightsView != nil {
			return m.insightsView.View(), false, true
		}
		return "", false, true
	case ViewOrphaned:
		if m.orphanedView != nil {
			return m.orphanedView.View(), false, true
		}
		return "", false, true
	case ViewHistory:
		if m.historyView != nil {
			return m.historyView.View(), false, true
		}
		return "", false, true
	case ViewBugReport:
		if m.bugReportView != nil {
			return m.bugReportView.View(), false, true
		}
		return "", false, true
	case ViewValuesGoals:
		if m.valuesView != nil {
			return m.valuesView.View(), false, true
		}
		return "", false, true
	case ViewFeedback:
		if m.feedbackView != nil {
			return m.feedbackView.View(), false, true
		}
		return "", false, true
	case ViewNextSteps:
		if m.nextStepsView != nil {
			return m.nextStepsView.View(), true, true
		}
		return "", true, true
	case ViewAvoidancePrompt:
		if m.avoidancePromptView != nil {
			return m.avoidancePromptView.View(), false, true
		}
		return "", false, true
	case ViewOnboarding:
		if m.onboardingView != nil {
			return m.onboardingView.View(), false, true
		}
		return "", false, true
	case ViewConflict:
		if m.conflictView != nil {
			return m.conflictView.View(), false, true
		}
		return "", false, true
	case ViewThemePicker:
		if m.themePickerView != nil {
			return m.themePickerView.View(), false, true
		}
		return "", false, true
	case ViewDevQueue:
		if m.devQueueView != nil {
			return m.devQueueView.View(), false, true
		}
		return "", false, true
	case ViewProposals:
		if m.proposalsView != nil {
			return m.proposalsView.View(), false, true
		}
		return "", false, true
	case ViewHelp:
		if m.helpView != nil {
			return m.helpView.View(), false, true
		}
		return "", false, true
	}
	return "", false, false
}

// updateAuxiliaryView delegates Update() to the current auxiliary view.
// Returns (model, cmd, handled).
func (m *MainModel) updateAuxiliaryView(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch m.viewMode {
	case ViewMood:
		if m.moodView == nil {
			return m, nil, true
		}
		cmd := m.moodView.Update(msg)
		return m, cmd, true
	case ViewHealth:
		if m.healthView == nil {
			return m, nil, true
		}
		cmd := m.healthView.Update(msg)
		return m, cmd, true
	case ViewInsights:
		if m.insightsView == nil {
			return m, nil, true
		}
		cmd := m.insightsView.Update(msg)
		return m, cmd, true
	case ViewOrphaned:
		if m.orphanedView == nil {
			return m, nil, true
		}
		cmd := m.orphanedView.Update(msg)
		return m, cmd, true
	case ViewHistory:
		if m.historyView == nil {
			return m, nil, true
		}
		cmd := m.historyView.Update(msg)
		return m, cmd, true
	case ViewBugReport:
		if m.bugReportView == nil {
			return m, nil, true
		}
		cmd := m.bugReportView.Update(msg)
		return m, cmd, true
	case ViewFeedback:
		if m.feedbackView == nil {
			return m, nil, true
		}
		cmd := m.feedbackView.Update(msg)
		return m, cmd, true
	case ViewNextSteps:
		if m.nextStepsView == nil {
			return m, nil, true
		}
		cmd := m.nextStepsView.Update(msg)
		return m, cmd, true
	case ViewAvoidancePrompt:
		if m.avoidancePromptView == nil {
			return m, nil, true
		}
		cmd := m.avoidancePromptView.Update(msg)
		return m, cmd, true
	case ViewOnboarding:
		if m.onboardingView == nil {
			return m, nil, true
		}
		cmd := m.onboardingView.Update(msg)
		return m, cmd, true
	case ViewConflict:
		if m.conflictView == nil {
			return m, nil, true
		}
		cmd := m.conflictView.Update(msg)
		return m, cmd, true
	case ViewThemePicker:
		if m.themePickerView == nil {
			return m, nil, true
		}
		cmd := m.themePickerView.Update(msg)
		return m, cmd, true
	case ViewHelp:
		if m.helpView == nil {
			return m, nil, true
		}
		cmd := m.helpView.Update(msg)
		return m, cmd, true
	case ViewDevQueue:
		if m.devQueueView == nil {
			return m, nil, true
		}
		cmd := m.devQueueView.Update(msg)
		return m, cmd, true
	case ViewProposals:
		if m.proposalsView == nil {
			return m, nil, true
		}
		cmd := m.proposalsView.Update(msg)
		return m, cmd, true
	case ViewValuesGoals:
		if m.valuesView == nil {
			return m, nil, true
		}
		cmd := m.valuesView.Update(msg)
		return m, cmd, true
	}
	return m, nil, false
}

// isAuxiliaryTextInputActive returns true when an auxiliary view has active text input.
func (m *MainModel) isAuxiliaryTextInputActive() bool {
	switch m.viewMode {
	case ViewSearch:
		return true
	case ViewOnboarding:
		return true
	case ViewBugReport:
		return m.bugReportView != nil && m.bugReportView.state == bugReportInput
	case ViewFeedback:
		return m.feedbackView != nil && m.feedbackView.isCustom
	case ViewMood:
		return m.moodView != nil && m.moodView.isCustom
	case ViewDetail:
		if m.detailView == nil {
			return false
		}
		return m.detailView.mode == DetailModeBlockerInput ||
			m.detailView.mode == DetailModeExpandInput
	case ViewValuesGoals:
		return m.valuesView != nil && m.valuesView.textInput.Focused()
	}
	return false
}

// findAvoidancePromptTask checks current doors for a task with 10+ bypasses
// that hasn't already been prompted this session. Returns the first match or nil.
func (m *MainModel) findAvoidancePromptTask() *core.Task {
	for _, task := range m.doorsView.currentDoors {
		count, ok := m.doorsView.avoidanceMap[task.Text]
		if ok && count >= 10 && !m.promptedTasks[task.Text] {
			return task
		}
	}
	return nil
}

// resolveHints returns the current inline hint enabled/fade state.
func (m *MainModel) resolveHints() bool {
	return m.showKeyHints
}

// findDuplicatePair finds the DuplicatePair involving the given task ID.
func (m *MainModel) findDuplicatePair(taskID string) *core.DuplicatePair {
	for i := range m.duplicatePairs {
		if m.duplicatePairs[i].TaskA.ID == taskID || m.duplicatePairs[i].TaskB.ID == taskID {
			return &m.duplicatePairs[i]
		}
	}
	return nil
}

// refreshDuplicates re-runs duplicate detection (after merge/dismiss).
func (m *MainModel) refreshDuplicates() {
	m.duplicateTaskIDs = make(map[string]bool)
	m.duplicatePairs = nil
	if m.dedupStore == nil {
		return
	}
	allTasks := m.pool.GetAllTasks()
	rawPairs := core.DetectDuplicates(allTasks, 0.8)
	m.duplicatePairs = m.dedupStore.FilterUndecided(rawPairs)
	for _, p := range m.duplicatePairs {
		m.duplicateTaskIDs[p.TaskA.ID] = true
		m.duplicateTaskIDs[p.TaskB.ID] = true
	}
	m.doorsView.SetDuplicateTaskIDs(m.duplicateTaskIDs)
}

// saveKeyHintsCmd returns a tea.Cmd that persists the key hints toggle to config.yaml.
func (m *MainModel) saveKeyHintsCmd(show bool) tea.Cmd {
	configPath := m.configPath
	return func() tea.Msg {
		if configPath == "" {
			return nil
		}
		cfg, err := core.LoadProviderConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load config for key hints save: %v\n", err)
			return nil
		}
		cfg.ShowKeyHints = &show
		cfg.ShowKeybindingBar = nil // clean up legacy field
		if err := core.SaveProviderConfig(configPath, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save key hints to config: %v\n", err)
		}
		return nil
	}
}

// saveThemeCmd returns a tea.Cmd that persists the theme to config.yaml.
func (m *MainModel) saveThemeCmd(themeName string) tea.Cmd {
	configPath := m.configPath
	return func() tea.Msg {
		if configPath == "" {
			return nil
		}
		cfg, err := core.LoadProviderConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load config for theme save: %v\n", err)
			return nil
		}
		cfg.Theme = themeName
		if err := core.SaveProviderConfig(configPath, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save theme to config: %v\n", err)
		}
		return nil
	}
}

// runDecompose returns a tea.Cmd that runs LLM decomposition asynchronously.
func (m *MainModel) runDecompose(taskID, taskDescription string) tea.Cmd {
	svc := m.agentService
	return func() tea.Msg {
		if svc == nil {
			return DecomposeResultMsg{
				TaskID: taskID,
				Err:    fmt.Errorf("LLM not configured"),
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		result, err := svc.DecomposeAndWrite(ctx, taskDescription)
		return DecomposeResultMsg{
			TaskID: taskID,
			Result: result,
			Err:    err,
		}
	}
}

// runBreakdown returns a tea.Cmd that runs LLM task breakdown asynchronously.
func (m *MainModel) runBreakdown(task *core.Task) tea.Cmd {
	svc := m.breakdownService
	taskID := task.ID
	taskDesc := task.Text
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		result, err := svc.Breakdown(ctx, taskID, taskDesc)
		return BreakdownResultMsg{
			TaskID: taskID,
			Result: result,
			Err:    err,
		}
	}
}

// runEnrich returns a tea.Cmd that runs LLM enrichment asynchronously.
func (m *MainModel) runEnrich(taskID, taskText string) tea.Cmd {
	enricher := m.enricher
	return func() tea.Msg {
		if enricher == nil {
			return EnrichResultMsg{
				TaskID: taskID,
				Err:    fmt.Errorf("LLM not configured"),
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := enricher.Enrich(ctx, taskText)
		return EnrichResultMsg{
			TaskID: taskID,
			Result: result,
			Err:    err,
		}
	}
}

// runExtraction returns a tea.Cmd that runs LLM extraction asynchronously.
func (m *MainModel) runExtraction(source, input string) tea.Cmd {
	ext := m.extractor
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var tasks []services.ExtractedTask
		var err error

		switch source {
		case "text":
			tasks, err = ext.ExtractFromText(ctx, input)
		case "file":
			tasks, err = ext.ExtractFromFile(ctx, input)
		case "clipboard":
			tasks, err = ext.ExtractFromClipboard(ctx)
		default:
			err = fmt.Errorf("unknown extraction source: %s", source)
		}

		return ExtractResultMsg{
			Tasks: tasks,
			Err:   err,
		}
	}
}

// handleDevDispatch processes a confirmed dispatch request by queuing the task.
func (m *MainModel) handleDevDispatch(task *core.Task) (tea.Model, tea.Cmd) {
	if m.devQueue == nil {
		m.flash = "Dev queue not configured"
		return m, ClearFlashCmd()
	}

	now := time.Now().UTC()
	task.DevDispatch = &dispatch.DevDispatch{
		Queued:   true,
		QueuedAt: &now,
	}

	q := m.devQueue
	item := dispatch.QueueItem{
		TaskID:   task.ID,
		TaskText: task.Text,
		Context:  task.Context,
		Status:   dispatch.QueueItemPending,
		QueuedAt: &now,
	}

	cmd := func() tea.Msg {
		if err := q.Add(item); err != nil {
			return DevDispatchResultMsg{TaskID: task.ID, Err: err}
		}
		return DevDispatchResultMsg{TaskID: task.ID}
	}

	return m, cmd
}

// handleWorkerStatus matches history entries to dispatched queue items and updates statuses.
func (m *MainModel) handleWorkerStatus(msg WorkerStatusMsg) tea.Cmd {
	if msg.Err != nil {
		log.Printf("worker status poll error: %v", msg.Err)
		if m.hasDispatchedItems() {
			return workerPollTickCmd()
		}
		m.pollingActive = false
		return nil
	}

	// Build lookup from worker name to history entry
	historyByWorker := make(map[string]dispatch.HistoryEntry, len(msg.History))
	for _, entry := range msg.History {
		historyByWorker[entry.WorkerName] = entry
	}

	// Match dispatched queue items to history entries
	items := m.devQueue.List()
	for _, item := range items {
		if item.Status != dispatch.QueueItemDispatched || item.WorkerName == "" {
			continue
		}

		entry, found := historyByWorker[item.WorkerName]
		if !found {
			continue
		}

		m.updateQueueItemFromHistory(item.ID, entry)
		m.updateTaskFromHistory(item.TaskID, entry)

		// Generate follow-up tasks for completed/failed items
		updatedItem, err := m.devQueue.Get(item.ID)
		if err == nil {
			m.generateFollowUpTasks(updatedItem)
		}
	}

	// Continue or stop polling
	if m.hasDispatchedItems() {
		return workerPollTickCmd()
	}
	m.pollingActive = false
	return nil
}

// updateQueueItemFromHistory updates a queue item based on a history entry.
func (m *MainModel) updateQueueItemFromHistory(itemID string, entry dispatch.HistoryEntry) {
	newStatus := mapHistoryStatus(entry.Status)
	if err := m.devQueue.Update(itemID, func(qi *dispatch.QueueItem) {
		qi.Status = newStatus
		qi.PRNumber = entry.PRNumber
		qi.PRURL = entry.PRURL
		if newStatus == dispatch.QueueItemCompleted || newStatus == dispatch.QueueItemFailed {
			now := time.Now().UTC()
			qi.CompletedAt = &now
		}
	}); err != nil {
		log.Printf("update queue item %s: %v", itemID, err)
	}
}

// updateTaskFromHistory updates a task's DevDispatch fields from a history entry.
func (m *MainModel) updateTaskFromHistory(taskID string, entry dispatch.HistoryEntry) {
	if taskID == "" {
		return
	}
	task := m.pool.GetTask(taskID)
	if task == nil {
		return
	}
	if task.DevDispatch == nil {
		task.DevDispatch = &dispatch.DevDispatch{}
	}
	task.DevDispatch.PRNumber = entry.PRNumber
	task.DevDispatch.PRStatus = mapPRStatus(entry.Status)
	m.pool.UpdateTask(task)
	if err := m.saveTasks(); err != nil {
		log.Printf("save tasks after worker status update: %v", err)
	}
}

// generateFollowUpTasks creates review and CI-fix tasks for a completed queue item.
func (m *MainModel) generateFollowUpTasks(item dispatch.QueueItem) {
	if item.Status != dispatch.QueueItemCompleted && item.Status != dispatch.QueueItemFailed {
		return
	}

	// Build existing task text set for deduplication
	existingTexts := make(map[string]bool)
	for _, t := range m.pool.GetAllTasks() {
		existingTexts[t.Text] = true
	}

	followUps := dispatch.GenerateFollowUpTasks(item, existingTexts)
	for _, fu := range followUps {
		task := core.NewTaskWithContext(fu.Text, fu.Context)
		task.DevDispatch = fu.DevDispatch
		m.pool.AddTask(task)
	}

	if len(followUps) > 0 {
		if err := m.saveTasks(); err != nil {
			log.Printf("save follow-up tasks: %v", err)
		}
	}
}

// hasDispatchedItems returns true if any queue items are in dispatched status.
func (m *MainModel) hasDispatchedItems() bool {
	if m.devQueue == nil {
		return false
	}
	for _, item := range m.devQueue.List() {
		if item.Status == dispatch.QueueItemDispatched {
			return true
		}
	}
	return false
}

// startPollingIfNeeded starts the polling tick if there are dispatched items and polling is not active.
func (m *MainModel) startPollingIfNeeded() tea.Cmd {
	if m.pollingActive || !m.hasDispatchedItems() {
		return nil
	}
	m.pollingActive = true
	return workerPollTickCmd()
}

// pollWorkerStatusCmd returns a tea.Cmd that calls GetHistory and returns a WorkerStatusMsg.
func (m *MainModel) pollWorkerStatusCmd() tea.Cmd {
	d := m.dispatcher
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		history, err := d.GetHistory(ctx, 10)
		return WorkerStatusMsg{History: history, Err: err}
	}
}
