package tui

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// handleTaskViewMessage handles Update() messages for task management views
// (Planning, AddTask, Breakdown, Extract, Import, Snooze, Deferred).
// Returns (model, cmd, handled). If handled is false, the caller should
// continue processing in the main Update switch.
func (m *MainModel) handleTaskViewMessage(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case AddTaskPromptMsg:
		m.addTaskView = NewAddTaskView()
		m.addTaskView.SetWidth(m.width)
		m.addTaskView.SetInlineHints(m.resolveHints())
		m.previousView = m.viewMode
		m.setViewMode(ViewAddTask)
		return m, nil, true

	case AddTaskWithContextPromptMsg:
		m.addTaskView = NewAddTaskWithContextView()
		m.addTaskView.SetWidth(m.width)
		m.addTaskView.SetInlineHints(m.resolveHints())
		if msg.PrefilledText != "" {
			m.addTaskView.capturedText = msg.PrefilledText
			m.addTaskView.step = stepContext
			m.addTaskView.textInput.Placeholder = "Why does this matter? (Enter to skip)"
		}
		m.previousView = m.viewMode
		m.setViewMode(ViewAddTask)
		return m, nil, true

	case TaskAddedMsg:
		m.pool.AddTask(msg.Task)
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.flash = taskAddedMessages[rand.IntN(len(taskAddedMessages))]
		m.addTaskView = nil
		// Return to previous view if it was search, otherwise show next steps
		if m.previousView == ViewSearch {
			m.searchView = m.newSearchView()
			m.searchView.SetWidth(m.width)
			m.setViewMode(ViewSearch)
			m.previousView = ViewDoors
		} else {
			m.doorsView.RefreshDoors()
			m.nextStepsView = NewNextStepsView("added", m.pool, m.completionCounter)
			m.nextStepsView.SetWidth(m.width)
			m.setViewMode(ViewNextSteps)
		}
		return m, ClearFlashCmd(), true

	case ShowSnoozeMsg:
		m.snoozeView = NewSnoozeView(msg.Task)
		m.snoozeView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.setViewMode(ViewSnooze)
		return m, nil, true

	case TaskSnoozedMsg:
		m.snoozeView = nil
		msg.Task.DeferUntil = msg.DeferDate
		if err := msg.Task.UpdateStatus(core.StatusDeferred); err != nil {
			m.goBack()
			m.flash = "Cannot snooze: " + err.Error()
			return m, ClearFlashCmd(), true
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after snooze: %v\n", err)
		}
		if m.previousView == ViewDeferred {
			if m.deferredListView != nil {
				m.deferredListView.Refresh()
			}
			m.setViewMode(ViewDeferred)
			m.flash = "Snooze date updated"
		} else {
			m.setViewMode(ViewDoors)
			m.doorsView.RefreshDoors()
			m.flash = "Task snoozed"
		}
		return m, ClearFlashCmd(), true

	case SnoozeCancelledMsg:
		m.snoozeView = nil
		m.goBack()
		return m, nil, true

	case EditDeferDateMsg:
		m.snoozeView = NewSnoozeView(msg.Task)
		m.snoozeView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.setViewMode(ViewSnooze)
		return m, nil, true

	case ShowDeferredListMsg:
		m.deferredListView = NewDeferredListView(m.pool)
		m.deferredListView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.setViewMode(ViewDeferred)
		return m, nil, true

	case UnsnoozeTaskMsg:
		if err := msg.Task.UpdateStatus(core.StatusTodo); err != nil {
			m.flash = fmt.Sprintf("Cannot un-snooze: %v", err)
			return m, ClearFlashCmd(), true
		}
		msg.Task.DeferUntil = nil
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.flash = "Task un-snoozed — returned to todo"
		if m.deferredListView != nil {
			m.deferredListView.Refresh()
		}
		return m, ClearFlashCmd(), true

	case BreakdownStartMsg:
		if m.breakdownService == nil {
			m.flash = "LLM not configured for breakdown"
			return m, ClearFlashCmd(), true
		}
		bv := NewBreakdownViewLoading(msg.Task)
		bv.SetWidth(m.width)
		m.breakdownView = bv
		m.previousView = m.viewMode
		m.setViewMode(ViewBreakdown)
		return m, m.runBreakdown(msg.Task), true

	case BreakdownResultMsg:
		if m.breakdownView == nil {
			return m, nil, true
		}
		if msg.Err != nil {
			m.breakdownView.SetError(msg.Err.Error())
			return m, nil, true
		}
		m.breakdownView.SetResult(msg.Result)
		return m, nil, true

	case BreakdownImportMsg:
		var tasks []*core.Task
		for _, st := range msg.Subtasks {
			t := core.NewTask(st.Text)
			tasks = append(tasks, t)
		}
		for _, t := range tasks {
			m.pool.AddTask(t)
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.breakdownView = nil
		m.setViewMode(ViewDoors)
		m.doorsView.RefreshDoors()
		m.flash = fmt.Sprintf("Imported %d subtasks", len(tasks))
		return m, ClearFlashCmd(), true

	case BreakdownCancelMsg:
		m.breakdownView = nil
		m.goBack()
		return m, nil, true

	case ShowExtractMsg:
		ev := NewExtractView()
		ev.SetWidth(m.width)
		m.extractView = ev
		m.previousView = m.viewMode
		m.setViewMode(ViewExtract)
		return m, nil, true

	case ExtractStartMsg:
		if m.extractor == nil {
			if m.extractView != nil {
				m.extractView.SetError("LLM service unavailable — configure an LLM backend to use :extract")
			}
			return m, nil, true
		}
		return m, m.runExtraction(msg.Source, msg.Input), true

	case ExtractResultMsg:
		if m.extractView == nil {
			return m, nil, true
		}
		if msg.Err != nil {
			m.extractView.SetError(msg.Err.Error())
			return m, nil, true
		}
		if len(msg.Tasks) == 0 {
			m.extractView.SetResult(nil, msg.BackendName)
			return m, nil, true
		}
		m.extractView.SetResult(msg.Tasks, msg.BackendName)
		return m, nil, true

	case ExtractImportMsg:
		for _, et := range msg.Tasks {
			t := core.NewTask(et.Text)
			m.pool.AddTask(t)
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after extract import: %v\n", err)
		}
		m.extractView = nil
		m.flash = fmt.Sprintf("Imported %d tasks from %s", len(msg.Tasks), msg.Source)
		m.doorsView.RefreshDoors()
		m.setViewMode(ViewDoors)
		return m, ClearFlashCmd(), true

	case ExtractCancelMsg:
		m.extractView = nil
		m.goBack()
		return m, nil, true

	case ShowPlanningMsg:
		pv := NewPlanningView(m.pool, m.provider)
		pv.SetWidth(m.width)
		pv.SetHeight(m.height)
		m.planningView = pv
		m.previousView = m.viewMode
		m.setViewMode(ViewPlanning)
		return m, pv.Init(), true

	case PlanningCompleteMsg:
		m.planningTimestamp = &msg.Timestamp
		m.doorsView.SetPlanningTimestamp(&msg.Timestamp)
		// Re-check seasonal theme on planning session start (handles overnight
		// sessions crossing season boundaries — AC7 of Story 33.3).
		m.doorsView.ResolveSeasonalTheme(time.Now().UTC())
		m.doorsView.RefreshDoors()
		m.doorsView.RotateFooterMessage()
		if m.planningMode {
			// CLI plan mode: exit after planning
			return m, tea.Quit, true
		}
		m.setViewMode(ViewDoors)
		m.planningView = nil
		focusCount := len(msg.FocusTasks)
		if focusCount > 0 {
			m.flash = fmt.Sprintf("Planning complete! %d focus task(s) set.", focusCount)
		} else {
			m.flash = "Planning complete!"
		}
		return m, ClearFlashCmd(), true

	case PlanningCancelledMsg:
		if m.planningMode {
			return m, tea.Quit, true
		}
		m.setViewMode(ViewDoors)
		m.planningView = nil
		return m, nil, true

	case ShowImportMsg:
		iv := NewImportView(msg.PrefilledPath)
		iv.SetWidth(m.width)
		m.importView = iv
		m.previousView = m.viewMode
		m.setViewMode(ViewImport)
		return m, nil, true

	case ImportConfirmedMsg:
		for _, t := range msg.Tasks {
			m.pool.AddTask(t)
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after import: %v\n", err)
		}
		m.importView = nil
		m.flash = fmt.Sprintf("Imported %d tasks from %s", len(msg.Tasks), msg.Source)
		m.doorsView.RefreshDoors()
		m.setViewMode(ViewDoors)
		return m, ClearFlashCmd(), true
	}

	return m, nil, false
}

// resizeTaskViews handles WindowSizeMsg for task management views.
func (m *MainModel) resizeTaskViews(width, height, contentH int) {
	if m.addTaskView != nil {
		m.addTaskView.SetWidth(width)
		m.addTaskView.SetHeight(contentH)
	}
	if m.deferredListView != nil {
		m.deferredListView.SetWidth(width)
		m.deferredListView.SetHeight(contentH)
	}
	if m.snoozeView != nil {
		m.snoozeView.SetWidth(width)
		m.snoozeView.SetHeight(contentH)
	}
	if m.planningView != nil {
		m.planningView.SetWidth(width)
		m.planningView.SetHeight(height)
	}
	if m.extractView != nil {
		m.extractView.SetWidth(width)
	}
	if m.breakdownView != nil {
		m.breakdownView.SetWidth(width)
		m.breakdownView.SetHeight(contentH)
	}
	if m.importView != nil {
		m.importView.SetWidth(width)
		m.importView.SetHeight(contentH)
	}
}

// taskViewContent handles View() for task management views.
// Returns (view, showValuesFooter, handled).
func (m *MainModel) taskViewContent() (string, bool, bool) {
	switch m.viewMode {
	case ViewAddTask:
		if m.addTaskView != nil {
			return m.addTaskView.View(), false, true
		}
		return "", false, true
	case ViewDeferred:
		if m.deferredListView != nil {
			return m.deferredListView.View(), false, true
		}
		return "", false, true
	case ViewSnooze:
		if m.snoozeView != nil {
			return m.snoozeView.View(), false, true
		}
		return "", false, true
	case ViewPlanning:
		if m.planningView != nil {
			return m.planningView.View(), false, true
		}
		return "", false, true
	case ViewBreakdown:
		if m.breakdownView != nil {
			return m.breakdownView.View(), false, true
		}
		return "", false, true
	case ViewExtract:
		if m.extractView != nil {
			return m.extractView.View(), false, true
		}
		return "", false, true
	case ViewImport:
		if m.importView != nil {
			return m.importView.View(), false, true
		}
		return "", false, true
	}
	return "", false, false
}

// isTaskTextInputActive returns whether a task management view has active text input.
func (m *MainModel) isTaskTextInputActive() bool {
	switch m.viewMode {
	case ViewAddTask:
		return true
	case ViewImport:
		return m.importView != nil && m.importView.step == importStepPath
	case ViewExtract:
		return m.extractView != nil &&
			(m.extractView.step == extractStepFileInput ||
				m.extractView.step == extractStepPasteInput ||
				m.extractView.step == extractStepEditing)
	}
	return false
}

// Update delegate methods for task management views.

func (m *MainModel) updateAddTask(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.addTaskView == nil {
		return m, nil
	}
	cmd := m.addTaskView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateDeferred(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.deferredListView == nil {
		return m, nil
	}
	cmd := m.deferredListView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateSnooze(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.snoozeView == nil {
		return m, nil
	}
	cmd := m.snoozeView.Update(msg)
	return m, cmd
}

func (m *MainModel) updatePlanning(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.planningView == nil {
		return m, nil
	}
	cmd := m.planningView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateBreakdown(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.breakdownView == nil {
		return m, nil
	}
	cmd := m.breakdownView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateExtract(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.extractView == nil {
		return m, nil
	}
	cmd := m.extractView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateImport(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.importView == nil {
		return m, nil
	}
	cmd := m.importView.Update(msg)
	return m, cmd
}
