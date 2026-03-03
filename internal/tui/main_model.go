package tui

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"

	"github.com/arcaven/ThreeDoors/internal/enrichment"
	"github.com/arcaven/ThreeDoors/internal/intelligence"
	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

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
	ViewImprovement
	ViewNextSteps
	ViewAvoidancePrompt
	ViewInsights
	ViewOnboarding
)

// MainModel is the root Bubbletea model that orchestrates view transitions.
type MainModel struct {
	viewMode            ViewMode
	previousView        ViewMode
	doorsView           *DoorsView
	detailView          *DetailView
	moodView            *MoodView
	searchView          *SearchView
	healthView          *HealthView
	addTaskView         *AddTaskView
	valuesView          *ValuesView
	feedbackView        *FeedbackView
	improvementView     *ImprovementView
	nextStepsView       *NextStepsView
	avoidancePromptView *AvoidancePromptView
	insightsView        *InsightsView
	onboardingView      *OnboardingView
	pool                *tasks.TaskPool
	tracker             *tasks.SessionTracker
	provider            tasks.TaskProvider
	healthChecker       *tasks.HealthChecker
	completionCounter   *tasks.CompletionCounter
	patternReport       *tasks.PatternReport
	patternAnalyzer     *tasks.PatternAnalyzer
	enrichDB            *enrichment.DB
	valuesConfig        *tasks.ValuesConfig
	syncTracker         *tasks.SyncStatusTracker
	agentService        *intelligence.AgentService
	decomposing         bool
	flash               string
	width               int
	height              int
	searchQuery         string
	searchSelectedIndex int
	promptedTasks       map[string]bool
}

// NewMainModel creates the root application model.
// If isFirstRun is true, the onboarding wizard is shown before the doors view.
func NewMainModel(pool *tasks.TaskPool, tracker *tasks.SessionTracker, provider tasks.TaskProvider, hc *tasks.HealthChecker, isFirstRun bool, edb *enrichment.DB) *MainModel {
	// Load values config
	var valuesConfig *tasks.ValuesConfig
	if path, err := tasks.GetValuesConfigPath(); err == nil {
		if cfg, err := tasks.LoadValuesConfig(path); err == nil {
			valuesConfig = cfg
		}
	}
	if valuesConfig == nil {
		valuesConfig = &tasks.ValuesConfig{}
	}

	// Initialize completion counter for daily tracking
	cc := tasks.NewCompletionCounter()
	if configPath, err := tasks.GetConfigDirPath(); err == nil {
		if loadErr := cc.LoadFromFile(filepath.Join(configPath, "completed.txt")); loadErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load completion history: %v\n", loadErr)
		}
	}

	// Initialize pattern analyzer: load both cached report and session history
	pa := tasks.NewPatternAnalyzer()
	var patternReport *tasks.PatternReport
	if configPath, err := tasks.GetConfigDirPath(); err == nil {
		patternReport, _ = pa.LoadPatterns(filepath.Join(configPath, "patterns.json"))
		if loadErr := pa.LoadSessions(filepath.Join(configPath, "sessions.jsonl")); loadErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load session history: %v\n", loadErr)
		}
	}

	// Initialize sync status tracker
	syncTracker := tasks.NewSyncStatusTracker()
	syncTracker.Register("Local")
	// Check if provider is WAL-wrapped and show pending count
	if walP, ok := provider.(*tasks.WALProvider); ok {
		syncTracker.Register("WAL")
		if pending := walP.PendingCount(); pending > 0 {
			syncTracker.SetPending("WAL", pending)
		}
	}

	doorsView := NewDoorsView(pool, tracker)
	doorsView.SetAvoidanceData(patternReport)
	doorsView.SetInsightsData(pa, cc)
	doorsView.SetSyncTracker(syncTracker)

	m := &MainModel{
		viewMode:          ViewDoors,
		doorsView:         doorsView,
		pool:              pool,
		tracker:           tracker,
		provider:          provider,
		healthChecker:     hc,
		completionCounter: cc,
		patternReport:     patternReport,
		patternAnalyzer:   pa,
		enrichDB:          edb,
		valuesConfig:      valuesConfig,
		syncTracker:       syncTracker,
		promptedTasks:     make(map[string]bool),
	}

	if isFirstRun {
		m.onboardingView = NewOnboardingView()
		m.viewMode = ViewOnboarding
	}

	return m
}

// SetAgentService sets the agent service for LLM task decomposition.
func (m *MainModel) SetAgentService(svc *intelligence.AgentService) {
	m.agentService = svc
}

// Init implements tea.Model.
func (m *MainModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.doorsView.SetWidth(msg.Width)
		if m.detailView != nil {
			m.detailView.SetWidth(msg.Width)
		}
		if m.moodView != nil {
			m.moodView.SetWidth(msg.Width)
		}
		if m.searchView != nil {
			m.searchView.SetWidth(msg.Width)
		}
		if m.healthView != nil {
			m.healthView.SetWidth(msg.Width)
		}
		if m.insightsView != nil {
			m.insightsView.SetWidth(msg.Width)
		}
		if m.addTaskView != nil {
			m.addTaskView.SetWidth(msg.Width)
		}
		if m.valuesView != nil {
			m.valuesView.SetWidth(msg.Width)
		}
		if m.feedbackView != nil {
			m.feedbackView.SetWidth(msg.Width)
		}
		if m.improvementView != nil {
			m.improvementView.SetWidth(msg.Width)
		}
		if m.nextStepsView != nil {
			m.nextStepsView.SetWidth(msg.Width)
		}
		if m.avoidancePromptView != nil {
			m.avoidancePromptView.SetWidth(msg.Width)
		}
		if m.onboardingView != nil {
			m.onboardingView.SetWidth(msg.Width)
		}
		return m, nil

	case ClearFlashMsg:
		m.flash = ""
		return m, nil

	case ReturnToDoorsMsg:
		// If we came from search, return to search instead
		if m.previousView == ViewSearch {
			m.searchView = NewSearchView(m.pool, m.tracker, m.healthChecker, m.completionCounter, m.patternReport)
			m.searchView.SetWidth(m.width)
			m.searchView.RestoreState(m.searchQuery, m.searchSelectedIndex)
			m.viewMode = ViewSearch
			m.detailView = nil
			m.addTaskView = nil
			m.previousView = ViewDoors
			return m, nil
		}
		m.viewMode = ViewDoors
		m.detailView = nil
		m.moodView = nil
		m.healthView = nil
		m.insightsView = nil
		m.addTaskView = nil
		m.doorsView.RefreshDoors()
		return m, nil

	case HealthCheckMsg:
		m.healthView = NewHealthView(msg.Result)
		m.healthView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewHealth
		return m, nil

	case NavigateToLinkedMsg:
		m.detailView = m.newDetailView(msg.Task)
		m.viewMode = ViewDetail
		return m, nil

	case ShowInsightsMsg:
		m.insightsView = NewInsightsView(m.patternAnalyzer, m.completionCounter)
		m.insightsView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewInsights
		return m, nil

	case ReturnToSearchMsg:
		m.searchView = NewSearchView(m.pool, m.tracker, m.healthChecker, m.completionCounter, m.patternReport)
		m.searchView.SetWidth(m.width)
		m.searchView.RestoreState(msg.Query, msg.SelectedIndex)
		m.viewMode = ViewSearch
		m.detailView = nil
		m.previousView = ViewDoors
		return m, nil

	case SearchClosedMsg:
		m.viewMode = ViewDoors
		m.searchView = nil
		m.previousView = ViewDoors
		return m, nil

	case SearchResultSelectedMsg:
		// Save search state for context-aware return
		if m.searchView != nil {
			m.searchQuery = m.searchView.textInput.Value()
			m.searchSelectedIndex = m.searchView.selectedIndex
		}
		m.previousView = ViewSearch
		m.detailView = m.newDetailView(msg.Task)
		m.viewMode = ViewDetail
		return m, nil

	case AddTaskPromptMsg:
		m.addTaskView = NewAddTaskView()
		m.addTaskView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewAddTask
		return m, nil

	case AddTaskWithContextPromptMsg:
		m.addTaskView = NewAddTaskWithContextView()
		m.addTaskView.SetWidth(m.width)
		if msg.PrefilledText != "" {
			m.addTaskView.capturedText = msg.PrefilledText
			m.addTaskView.step = stepContext
			m.addTaskView.textInput.Placeholder = "Why does this matter? (Enter to skip)"
		}
		m.previousView = m.viewMode
		m.viewMode = ViewAddTask
		return m, nil

	case TaskAddedMsg:
		m.pool.AddTask(msg.Task)
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.flash = taskAddedMessages[rand.IntN(len(taskAddedMessages))]
		m.addTaskView = nil
		// Return to previous view if it was search, otherwise show next steps
		if m.previousView == ViewSearch {
			m.searchView = NewSearchView(m.pool, m.tracker, m.healthChecker, m.completionCounter, m.patternReport)
			m.searchView.SetWidth(m.width)
			m.viewMode = ViewSearch
			m.previousView = ViewDoors
		} else {
			m.doorsView.RefreshDoors()
			m.nextStepsView = NewNextStepsView("added", m.pool, m.completionCounter)
			m.nextStepsView.SetWidth(m.width)
			m.viewMode = ViewNextSteps
		}
		return m, ClearFlashCmd()

	case TaskCompletedMsg:
		if err := m.provider.MarkComplete(msg.Task.ID); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to mark complete: %v\n", err)
			m.flash = "Error completing task"
			return m, ClearFlashCmd()
		}
		m.pool.RemoveTask(msg.Task.ID)
		m.doorsView.IncrementCompleted()
		m.completionCounter.IncrementToday()
		celebration := celebrationMessages[rand.IntN(len(celebrationMessages))]
		if dailyMsg := m.completionCounter.FormatCompletionMessage(); dailyMsg != "" {
			celebration += " | " + dailyMsg
		}
		m.flash = celebration
		m.detailView = nil
		m.doorsView.RefreshDoors()
		// Show next-steps view instead of returning directly to doors
		m.nextStepsView = NewNextStepsView("completed", m.pool, m.completionCounter)
		m.nextStepsView.SetWidth(m.width)
		m.viewMode = ViewNextSteps
		return m, ClearFlashCmd()

	case TaskUpdatedMsg:
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.viewMode = ViewDoors
		m.detailView = nil
		m.doorsView.RefreshDoors()
		return m, nil

	case MoodCapturedMsg:
		if m.tracker != nil {
			m.tracker.RecordMood(msg.Mood, msg.CustomText)
		}
		m.viewMode = ViewDoors
		m.moodView = nil
		m.flash = fmt.Sprintf("Mood logged: %s", msg.Mood)
		return m, ClearFlashCmd()

	case ShowMoodMsg:
		m.moodView = NewMoodView()
		m.moodView.SetWidth(m.width)
		m.viewMode = ViewMood
		return m, nil

	case ShowFeedbackMsg:
		m.feedbackView = NewFeedbackView(msg.Task)
		m.feedbackView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewFeedback
		return m, nil

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
		m.viewMode = ViewDoors
		m.doorsView.RefreshDoors()
		m.flash = "Feedback recorded"
		return m, ClearFlashCmd()

	case ShowValuesSetupMsg:
		m.valuesView = NewValuesSetupView(m.valuesConfig)
		m.valuesView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewValuesGoals
		return m, nil

	case ShowValuesEditMsg:
		m.valuesView = NewValuesEditView(m.valuesConfig)
		m.valuesView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewValuesGoals
		return m, nil

	case ValuesSavedMsg:
		m.valuesConfig = msg.Config
		if path, err := tasks.GetValuesConfigPath(); err == nil {
			if saveErr := tasks.SaveValuesConfig(path, msg.Config); saveErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save values config: %v\n", saveErr)
			}
		}
		m.valuesView = nil
		m.flash = "Values saved"
		m.viewMode = ViewDoors
		m.doorsView.RefreshDoors()
		return m, ClearFlashCmd()

	case RequestQuitMsg:
		// Show improvement prompt if session qualifies (5+ min OR 1+ completions)
		if m.tracker != nil {
			metrics := m.tracker.GetMetricsSnapshot()
			if metrics.TasksCompleted >= 1 || metrics.DurationSeconds() >= 300 {
				m.improvementView = NewImprovementView()
				m.improvementView.SetWidth(m.width)
				m.viewMode = ViewImprovement
				return m, nil
			}
		}
		return m, tea.Quit

	case ImprovementSubmittedMsg:
		if m.tracker != nil && msg.Text != "" {
			configDir, err := tasks.GetConfigDirPath()
			if err == nil {
				if writeErr := tasks.WriteImprovement(configDir, m.tracker.GetSessionID(), msg.Text); writeErr != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to save improvement: %v\n", writeErr)
				}
			}
		}
		return m, tea.Quit

	case ImprovementSkippedMsg:
		return m, tea.Quit

	case ShowNextStepsMsg:
		m.nextStepsView = NewNextStepsView(msg.Context, m.pool, m.completionCounter)
		m.nextStepsView.SetWidth(m.width)
		m.viewMode = ViewNextSteps
		return m, nil

	case NextStepSelectedMsg:
		m.nextStepsView = nil
		switch msg.Action {
		case "doors":
			m.viewMode = ViewDoors
			m.doorsView.RefreshDoors()
			m.doorsView.RotateFooterMessage()
		case "add":
			return m, func() tea.Msg { return AddTaskPromptMsg{} }
		case "mood":
			return m, func() tea.Msg { return ShowMoodMsg{} }
		case "search":
			m.searchView = NewSearchView(m.pool, m.tracker, m.healthChecker, m.completionCounter, m.patternReport)
			m.searchView.SetWidth(m.width)
			m.viewMode = ViewSearch
			m.previousView = ViewDoors
		case "stats":
			m.searchView = NewSearchView(m.pool, m.tracker, m.healthChecker, m.completionCounter, m.patternReport)
			m.searchView.SetWidth(m.width)
			m.searchView.textInput.SetValue(":stats")
			m.searchView.checkCommandMode()
			m.viewMode = ViewSearch
			m.previousView = ViewDoors
		default:
			m.viewMode = ViewDoors
			m.doorsView.RefreshDoors()
		}
		return m, nil

	case NextStepDismissedMsg:
		m.nextStepsView = nil
		m.viewMode = ViewDoors
		m.doorsView.RefreshDoors()
		m.doorsView.RotateFooterMessage()
		return m, nil

	case ShowAvoidancePromptMsg:
		m.avoidancePromptView = NewAvoidancePromptView(msg.Task, m.doorsView.avoidanceMap[msg.Task.Text])
		m.avoidancePromptView.SetWidth(m.width)
		m.promptedTasks[msg.Task.Text] = true
		m.previousView = m.viewMode
		m.viewMode = ViewAvoidancePrompt
		return m, nil

	case AvoidanceActionMsg:
		m.avoidancePromptView = nil
		switch msg.Action {
		case "reconsider":
			if err := msg.Task.UpdateStatus(tasks.StatusInProgress); err == nil {
				if err := m.saveTasks(); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
				}
			}
			m.detailView = m.newDetailView(msg.Task)
			m.viewMode = ViewDetail
			m.flash = "Taking it on!"
			return m, ClearFlashCmd()
		case "breakdown":
			m.detailView = m.newDetailView(msg.Task)
			m.viewMode = ViewDetail
			return m, nil
		case "defer":
			if err := msg.Task.UpdateStatus(tasks.StatusDeferred); err == nil {
				if err := m.saveTasks(); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
				}
			}
			m.viewMode = ViewDoors
			m.doorsView.RefreshDoors()
			m.flash = "Task set aside for later"
			return m, ClearFlashCmd()
		case "archive":
			if err := msg.Task.UpdateStatus(tasks.StatusArchived); err == nil {
				if err := m.provider.MarkComplete(msg.Task.ID); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to archive task: %v\n", err)
				}
				m.pool.RemoveTask(msg.Task.ID)
			}
			m.viewMode = ViewDoors
			m.doorsView.RefreshDoors()
			m.flash = "Task archived"
			return m, ClearFlashCmd()
		default:
			m.viewMode = ViewDoors
			m.doorsView.RefreshDoors()
		}
		return m, nil

	case OnboardingCompletedMsg:
		m.onboardingView = nil
		m.viewMode = ViewDoors
		// Save values if provided
		if len(msg.Values) > 0 {
			m.valuesConfig = &tasks.ValuesConfig{Values: msg.Values}
			if path, err := tasks.GetValuesConfigPath(); err == nil {
				if saveErr := tasks.SaveValuesConfig(path, m.valuesConfig); saveErr != nil {
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
		if configDir, err := tasks.GetConfigDirPath(); err == nil {
			if markErr := tasks.MarkOnboardingComplete(configDir); markErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save onboarding state: %v\n", markErr)
			}
		}
		var cmd tea.Cmd
		if m.flash != "" {
			cmd = ClearFlashCmd()
		}
		return m, cmd

	case FlashMsg:
		m.flash = msg.Text
		return m, ClearFlashCmd()

	case DecomposeStartMsg:
		if m.decomposing {
			m.flash = "Decomposition already in progress"
			return m, ClearFlashCmd()
		}
		m.decomposing = true
		m.flash = "Decomposing task..."
		return m, m.runDecompose(msg.TaskID, msg.TaskDescription)

	case DecomposeResultMsg:
		m.decomposing = false
		if msg.Err != nil {
			m.flash = fmt.Sprintf("Decompose failed: %s", msg.Err.Error())
			return m, ClearFlashCmd()
		}
		m.flash = fmt.Sprintf("Decomposed into %d stories", len(msg.Result.Stories))
		return m, ClearFlashCmd()

	case SyncStatusUpdateMsg:
		if m.syncTracker != nil {
			switch msg.Phase {
			case tasks.SyncPhaseSynced:
				m.syncTracker.SetSynced(msg.ProviderName)
			case tasks.SyncPhaseSyncing:
				m.syncTracker.SetSyncing(msg.ProviderName)
			case tasks.SyncPhasePending:
				m.syncTracker.SetPending(msg.ProviderName, msg.PendingCount)
			case tasks.SyncPhaseError:
				m.syncTracker.SetError(msg.ProviderName, msg.ErrorMsg)
			}
		}
		return m, nil
	}

	// Delegate to current view
	switch m.viewMode {
	case ViewDoors:
		return m.updateDoors(msg)
	case ViewDetail:
		return m.updateDetail(msg)
	case ViewMood:
		return m.updateMood(msg)
	case ViewSearch:
		return m.updateSearch(msg)
	case ViewHealth:
		return m.updateHealth(msg)
	case ViewInsights:
		return m.updateInsights(msg)
	case ViewAddTask:
		return m.updateAddTask(msg)
	case ViewValuesGoals:
		return m.updateValues(msg)
	case ViewFeedback:
		return m.updateFeedback(msg)
	case ViewImprovement:
		return m.updateImprovement(msg)
	case ViewNextSteps:
		return m.updateNextSteps(msg)
	case ViewAvoidancePrompt:
		return m.updateAvoidancePrompt(msg)
	case ViewOnboarding:
		return m.updateOnboarding(msg)
	}

	return m, nil
}

func (m *MainModel) updateImprovement(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.improvementView == nil {
		return m, nil
	}
	cmd := m.improvementView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateDoors(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, func() tea.Msg { return RequestQuitMsg{} }
		case "a", "left":
			m.doorsView.selectedDoorIndex = 0
		case "w", "up":
			m.doorsView.selectedDoorIndex = 1
		case "d", "right":
			m.doorsView.selectedDoorIndex = 2
		case "s", "down":
			if m.tracker != nil {
				m.tracker.RecordRefresh(m.doorsView.GetCurrentDoorTexts())
			}
			m.doorsView.RefreshDoors()
			m.doorsView.RotateFooterMessage()
			// Check for 10+ bypassed tasks and show avoidance prompt
			if task := m.findAvoidancePromptTask(); task != nil {
				return m, func() tea.Msg { return ShowAvoidancePromptMsg{Task: task} }
			}
			m.flash = doorRefreshMessages[rand.IntN(len(doorRefreshMessages))]
			return m, ClearFlashCmd()
		case "enter":
			if m.doorsView.selectedDoorIndex >= 0 && m.doorsView.selectedDoorIndex < len(m.doorsView.currentDoors) {
				task := m.doorsView.currentDoors[m.doorsView.selectedDoorIndex]
				if m.tracker != nil {
					m.tracker.RecordDoorSelection(m.doorsView.selectedDoorIndex, task.Text)
				}
				m.detailView = m.newDetailView(task)
				m.viewMode = ViewDetail
			}
		case "n", "N":
			if m.doorsView.selectedDoorIndex >= 0 && m.doorsView.selectedDoorIndex < len(m.doorsView.currentDoors) {
				task := m.doorsView.currentDoors[m.doorsView.selectedDoorIndex]
				return m, func() tea.Msg { return ShowFeedbackMsg{Task: task} }
			}
		case "m", "M":
			return m, func() tea.Msg { return ShowMoodMsg{} }
		case "/":
			m.searchView = NewSearchView(m.pool, m.tracker, m.healthChecker, m.completionCounter, m.patternReport)
			m.searchView.SetWidth(m.width)
			m.viewMode = ViewSearch
			m.previousView = ViewDoors
			return m, nil
		case ":":
			m.searchView = NewSearchView(m.pool, m.tracker, m.healthChecker, m.completionCounter, m.patternReport)
			m.searchView.SetWidth(m.width)
			m.searchView.textInput.SetValue(":")
			m.searchView.checkCommandMode()
			m.viewMode = ViewSearch
			m.previousView = ViewDoors
			return m, nil
		}
	}
	return m, nil
}

func (m *MainModel) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.detailView == nil {
		return m, nil
	}
	cmd := m.detailView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateMood(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.moodView == nil {
		return m, nil
	}
	cmd := m.moodView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateHealth(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.healthView == nil {
		return m, nil
	}
	cmd := m.healthView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateInsights(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.insightsView == nil {
		return m, nil
	}
	cmd := m.insightsView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.searchView == nil {
		return m, nil
	}
	cmd := m.searchView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateAddTask(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.addTaskView == nil {
		return m, nil
	}
	cmd := m.addTaskView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateFeedback(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.feedbackView == nil {
		return m, nil
	}
	cmd := m.feedbackView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateNextSteps(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.nextStepsView == nil {
		return m, nil
	}
	cmd := m.nextStepsView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateAvoidancePrompt(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.avoidancePromptView == nil {
		return m, nil
	}
	cmd := m.avoidancePromptView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateOnboarding(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.onboardingView == nil {
		return m, nil
	}
	cmd := m.onboardingView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateValues(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.valuesView == nil {
		return m, nil
	}
	cmd := m.valuesView.Update(msg)
	return m, cmd
}

// findAvoidancePromptTask checks current doors for a task with 10+ bypasses
// that hasn't already been prompted this session. Returns the first match or nil.
func (m *MainModel) findAvoidancePromptTask() *tasks.Task {
	for _, task := range m.doorsView.currentDoors {
		count, ok := m.doorsView.avoidanceMap[task.Text]
		if ok && count >= 10 && !m.promptedTasks[task.Text] {
			return task
		}
	}
	return nil
}

func (m *MainModel) newDetailView(task *tasks.Task) *DetailView {
	dv := NewDetailView(task, m.tracker, m.enrichDB, m.pool)
	dv.SetWidth(m.width)
	dv.SetAgentService(m.agentService)
	return dv
}

func (m *MainModel) saveTasks() error {
	allTasks := m.pool.GetAllTasks()
	return m.provider.SaveTasks(allTasks)
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

// View implements tea.Model.
func (m *MainModel) View() string {
	var view string
	showValuesFooter := false

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
	case ViewAddTask:
		if m.addTaskView != nil {
			view = m.addTaskView.View()
		}
	case ViewValuesGoals:
		if m.valuesView != nil {
			view = m.valuesView.View()
		}
	case ViewFeedback:
		if m.feedbackView != nil {
			view = m.feedbackView.View()
		}
	case ViewImprovement:
		if m.improvementView != nil {
			view = m.improvementView.View()
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
	default:
		view = m.doorsView.View()
		showValuesFooter = true
	}

	if m.flash != "" {
		view += "\n" + flashStyle.Render(m.flash)
	}

	if showValuesFooter {
		view += RenderValuesFooter(m.valuesConfig)
	}

	return view
}
