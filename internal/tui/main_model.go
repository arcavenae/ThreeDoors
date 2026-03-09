package tui

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/dispatch"
	"github.com/arcaven/ThreeDoors/internal/enrichment"
	"github.com/arcaven/ThreeDoors/internal/intelligence"
	"github.com/arcaven/ThreeDoors/internal/mcp"
	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	"github.com/charmbracelet/bubbles/spinner"
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
	ViewConflict
	ViewSyncLog
	ViewThemePicker
	ViewDevQueue
	ViewProposals
	ViewHelp
	ViewDeferred
	ViewSnooze
)

// MainModel is the root Bubbletea model that orchestrates view transitions.
type MainModel struct {
	viewMode              ViewMode
	previousView          ViewMode
	doorsView             *DoorsView
	detailView            *DetailView
	moodView              *MoodView
	searchView            *SearchView
	healthView            *HealthView
	addTaskView           *AddTaskView
	valuesView            *ValuesView
	feedbackView          *FeedbackView
	improvementView       *ImprovementView
	nextStepsView         *NextStepsView
	avoidancePromptView   *AvoidancePromptView
	insightsView          *InsightsView
	onboardingView        *OnboardingView
	conflictView          *ConflictView
	syncLogView           *SyncLogView
	themePickerView       *ThemePicker
	devQueueView          *DevQueueView
	proposalsView         *ProposalsView
	helpView              *HelpView
	deferredListView      *DeferredListView
	snoozeView            *SnoozeView
	configPath            string
	pool                  *core.TaskPool
	tracker               *core.SessionTracker
	provider              core.TaskProvider
	healthChecker         *core.HealthChecker
	completionCounter     *core.CompletionCounter
	patternReport         *core.PatternReport
	patternAnalyzer       *core.PatternAnalyzer
	enrichDB              *enrichment.DB
	valuesConfig          *core.ValuesConfig
	syncTracker           *core.SyncStatusTracker
	agentService          *intelligence.AgentService
	decomposing           bool
	syncLog               *core.SyncLog
	dedupStore            *core.DedupStore
	duplicateTaskIDs      map[string]bool
	duplicatePairs        []core.DuplicatePair
	devDispatchEnabled    bool
	dispatcher            dispatch.Dispatcher
	devQueue              *dispatch.DevQueue
	proposalStore         *mcp.ProposalStore
	pollingActive         bool
	syncSpinner           *SyncSpinner
	flash                 string
	width                 int
	height                int
	showKeybindingBar     bool
	showKeybindingOverlay bool
	overlayState          OverlayState
	searchQuery           string
	searchSelectedIndex   int
	promptedTasks         map[string]bool
}

// NewMainModel creates the root application model.
// If isFirstRun is true, the onboarding wizard is shown before the doors view.
func NewMainModel(pool *core.TaskPool, tracker *core.SessionTracker, provider core.TaskProvider, hc *core.HealthChecker, isFirstRun bool, edb *enrichment.DB) *MainModel {
	// Load values config
	var valuesConfig *core.ValuesConfig
	if path, err := core.GetValuesConfigPath(); err == nil {
		if cfg, err := core.LoadValuesConfig(path); err == nil {
			valuesConfig = cfg
		}
	}
	if valuesConfig == nil {
		valuesConfig = &core.ValuesConfig{}
	}

	// Initialize completion counter for daily tracking
	cc := core.NewCompletionCounter()
	if configPath, err := core.GetConfigDirPath(); err == nil {
		if loadErr := cc.LoadFromFile(filepath.Join(configPath, "completed.txt")); loadErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load completion history: %v\n", loadErr)
		}
	}

	// Initialize pattern analyzer: load both cached report and session history
	pa := core.NewPatternAnalyzer()
	var patternReport *core.PatternReport
	if configPath, err := core.GetConfigDirPath(); err == nil {
		patternReport, _ = pa.LoadPatterns(filepath.Join(configPath, "patterns.json"))
		if loadErr := pa.LoadSessions(filepath.Join(configPath, "sessions.jsonl")); loadErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load session history: %v\n", loadErr)
		}
	}

	// Initialize sync log and status tracker
	var syncLog *core.SyncLog
	if configPath, err := core.GetConfigDirPath(); err == nil {
		syncLog = core.NewSyncLog(configPath)
	}

	syncTracker := core.NewSyncStatusTracker()
	syncTracker.Register("Local")
	// Check if provider is WAL-wrapped and show pending count
	if walP, ok := provider.(*core.WALProvider); ok {
		syncTracker.Register("WAL")
		if pending := walP.PendingCount(); pending > 0 {
			syncTracker.SetPending("WAL", pending)
		}
	}

	// Initialize dedup store for duplicate detection decisions
	var dedupStore *core.DedupStore
	duplicateTaskIDs := make(map[string]bool)
	var duplicatePairs []core.DuplicatePair
	if configPath, err := core.GetConfigDirPath(); err == nil {
		ds, dsErr := core.NewDedupStore(filepath.Join(configPath, "dedup_decisions.yaml"))
		if dsErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load dedup store: %v\n", dsErr)
		} else {
			dedupStore = ds
			allTasks := pool.GetAllTasks()
			rawPairs := core.DetectDuplicates(allTasks, 0.8)
			duplicatePairs = dedupStore.FilterUndecided(rawPairs)
			for _, p := range duplicatePairs {
				duplicateTaskIDs[p.TaskA.ID] = true
				duplicateTaskIDs[p.TaskB.ID] = true
			}
		}
	}

	doorsView := NewDoorsView(pool, tracker)
	doorsView.SetAvoidanceData(patternReport)
	doorsView.SetInsightsData(pa, cc)
	doorsView.SetSyncTracker(syncTracker)
	doorsView.SetDuplicateTaskIDs(duplicateTaskIDs)

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
		syncLog:           syncLog,
		dedupStore:        dedupStore,
		duplicateTaskIDs:  duplicateTaskIDs,
		duplicatePairs:    duplicatePairs,
		syncSpinner:       NewSyncSpinner(),
		promptedTasks:     make(map[string]bool),
		showKeybindingBar: true, // default: bar visible
	}

	doorsView.SetSyncSpinner(m.syncSpinner)

	if isFirstRun {
		m.onboardingView = NewOnboardingView()
		m.viewMode = ViewOnboarding
	}

	return m
}

// SetConfigPath sets the path to config.yaml for theme persistence.
func (m *MainModel) SetConfigPath(path string) {
	m.configPath = path
}

// SetShowKeybindingBar sets the initial keybinding bar visibility from config.
func (m *MainModel) SetShowKeybindingBar(show bool) {
	m.showKeybindingBar = show
}

// SetDevDispatch configures dev dispatch with the given dispatcher, queue, and enabled flag.
func (m *MainModel) SetDevDispatch(enabled bool, d dispatch.Dispatcher, q *dispatch.DevQueue) {
	m.devDispatchEnabled = enabled
	m.dispatcher = d
	m.devQueue = q
}

// SetAgentService sets the agent service for LLM task decomposition.
func (m *MainModel) SetAgentService(svc *intelligence.AgentService) {
	m.agentService = svc
}

// SetProposalStore sets the proposal store for the proposal review view.
func (m *MainModel) SetProposalStore(store *mcp.ProposalStore) {
	m.proposalStore = store
	if store != nil {
		count := PendingProposalCount(store)
		m.doorsView.SetPendingProposals(count)
	}
}

// SetDispatcher sets the Dispatcher used for worker status polling.
func (m *MainModel) SetDispatcher(d dispatch.Dispatcher) {
	m.dispatcher = d
}

// SetDevQueue sets the DevQueue used for tracking dispatched items.
func (m *MainModel) SetDevQueue(q *dispatch.DevQueue) {
	m.devQueue = q
}

// Init implements tea.Model.
func (m *MainModel) Init() tea.Cmd {
	// Check for expired deferred tasks on startup and start periodic tick.
	returned := core.CheckDeferredReturnsWithTracker(m.pool, m.tracker)
	if returned > 0 {
		m.doorsView.RefreshDoors()
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after defer return: %v\n", err)
		}
	}
	return deferReturnTickCmd()
}

// Update implements tea.Model.
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.doorsView.SetWidth(msg.Width)
		m.doorsView.SetHeight(m.contentHeight())
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
		if m.conflictView != nil {
			m.conflictView.SetWidth(msg.Width)
		}
		if m.syncLogView != nil {
			m.syncLogView.SetWidth(msg.Width)
		}
		if m.themePickerView != nil {
			m.themePickerView.SetWidth(msg.Width)
		}
		if m.devQueueView != nil {
			m.devQueueView.SetWidth(msg.Width)
		}
		if m.proposalsView != nil {
			m.proposalsView.SetWidth(msg.Width)
			m.proposalsView.SetHeight(m.contentHeight())
		}
		if m.helpView != nil {
			m.helpView.SetWidth(msg.Width)
			m.helpView.SetHeight(msg.Height)
		}
		if m.deferredListView != nil {
			m.deferredListView.SetWidth(msg.Width)
		}
		if m.snoozeView != nil {
			m.snoozeView.SetWidth(msg.Width)
		}
		return m, nil

	case ClearFlashMsg:
		m.flash = ""
		return m, nil

	case ReturnToDoorsMsg:
		// If we came from search, return to search instead
		if m.previousView == ViewSearch {
			m.searchView = m.newSearchView()
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
		m.deferredListView = nil
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
		m.searchView = m.newSearchView()
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

	case ExpandTaskMsg:
		newTask := core.NewTask(msg.NewTaskText)
		m.pool.AddTask(newTask)
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.flash = "Subtask added"
		m.detailView = nil
		m.doorsView.RefreshDoors()
		m.viewMode = ViewDoors
		return m, ClearFlashCmd()

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
		// Check for newly unblocked tasks before removing the completed task
		unblockedTasks := core.GetNewlyUnblockedTasks(msg.Task.ID, m.pool)
		core.ClearCompletedDependency(msg.Task.ID, m.pool)
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
		cmds := []tea.Cmd{ClearFlashCmd()}
		if len(unblockedTasks) > 0 {
			cmds = append(cmds, func() tea.Msg {
				return DependencyUnblockedMsg{
					UnblockedTasks: unblockedTasks,
					CompletedDepID: msg.Task.ID,
				}
			})
		}
		return m, tea.Batch(cmds...)

	case TaskUndoneMsg:
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.viewMode = ViewDoors
		m.detailView = nil
		m.doorsView.RefreshDoors()
		m.flash = "Task uncompleted — returned to todo"
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
		if path, err := core.GetValuesConfigPath(); err == nil {
			if saveErr := core.SaveValuesConfig(path, msg.Config); saveErr != nil {
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
			configDir, err := core.GetConfigDirPath()
			if err == nil {
				if writeErr := core.WriteImprovement(configDir, m.tracker.GetSessionID(), msg.Text); writeErr != nil {
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
			m.searchView = m.newSearchView()
			m.searchView.SetWidth(m.width)
			m.viewMode = ViewSearch
			m.previousView = ViewDoors
		case "stats":
			m.searchView = m.newSearchView()
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
			if err := msg.Task.UpdateStatus(core.StatusInProgress); err == nil {
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
			if err := msg.Task.UpdateStatus(core.StatusDeferred); err == nil {
				if m.tracker != nil {
					m.tracker.RecordSnooze(msg.Task.ID, msg.Task.DeferUntil, "someday")
				}
				if err := m.saveTasks(); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
				}
			}
			m.viewMode = ViewDoors
			m.doorsView.RefreshDoors()
			m.flash = "Task set aside for later"
			return m, ClearFlashCmd()
		case "archive":
			if err := msg.Task.UpdateStatus(core.StatusArchived); err == nil {
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

	case ShowSnoozeMsg:
		m.snoozeView = NewSnoozeView(msg.Task)
		m.snoozeView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewSnooze
		return m, nil

	case TaskSnoozedMsg:
		m.snoozeView = nil
		msg.Task.DeferUntil = msg.DeferDate
		if err := msg.Task.UpdateStatus(core.StatusDeferred); err != nil {
			m.viewMode = m.previousView
			m.flash = "Cannot snooze: " + err.Error()
			return m, ClearFlashCmd()
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after snooze: %v\n", err)
		}
		if m.previousView == ViewDeferred {
			if m.deferredListView != nil {
				m.deferredListView.Refresh()
			}
			m.viewMode = ViewDeferred
			m.flash = "Snooze date updated"
		} else {
			m.viewMode = ViewDoors
			m.doorsView.RefreshDoors()
			m.flash = "Task snoozed"
		}
		return m, ClearFlashCmd()

	case SnoozeCancelledMsg:
		m.snoozeView = nil
		m.viewMode = m.previousView
		return m, nil

	case OnboardingCompletedMsg:
		m.onboardingView = nil
		m.viewMode = ViewDoors
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

	case SyncConflictMsg:
		cv := NewConflictView(msg.ConflictSet, m.syncLog)
		cv.SetWidth(m.width)
		m.conflictView = cv
		m.previousView = m.viewMode
		m.viewMode = ViewConflict
		return m, nil

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
		m.viewMode = ViewDoors
		m.doorsView.RefreshDoors()
		m.flash = fmt.Sprintf("%d conflict(s) resolved", len(resolutions))
		return m, ClearFlashCmd()

	case ShowSyncLogMsg:
		sv := NewSyncLogView(msg.Entries)
		sv.SetWidth(m.width)
		m.syncLogView = sv
		m.previousView = m.viewMode
		m.viewMode = ViewSyncLog
		return m, nil
	case DuplicateDismissedMsg:
		m.refreshDuplicates()
		m.flash = "Duplicate flag dismissed"
		m.detailView = nil
		m.viewMode = ViewDoors
		m.doorsView.RefreshDoors()
		return m, ClearFlashCmd()

	case DuplicateMergedMsg:
		// Remove the duplicate task
		if err := m.provider.DeleteTask(msg.RemovedTask.ID); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to delete merged duplicate: %v\n", err)
		}
		m.pool.RemoveTask(msg.RemovedTask.ID)
		m.refreshDuplicates()
		m.flash = "Duplicate merged"
		m.detailView = nil
		m.viewMode = ViewDoors
		m.doorsView.RefreshDoors()
		return m, ClearFlashCmd()

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
		m.viewMode = ViewThemePicker
		return m, nil

	case ThemeSelectedMsg:
		m.doorsView.SetThemeByName(msg.Name)
		m.themePickerView = nil
		m.viewMode = ViewDoors
		m.doorsView.RefreshDoors()
		m.flash = fmt.Sprintf("Theme changed to %s", msg.Name)
		return m, tea.Batch(ClearFlashCmd(), m.saveThemeCmd(msg.Name))

	case ThemeCancelledMsg:
		m.themePickerView = nil
		m.viewMode = ViewDoors
		return m, nil

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
		return m, nil

	case DependencyAddedMsg:
		m.doorsView.RefreshDoors()
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after dependency added: %v\n", err)
		}
		m.flash = "Dependency added"
		return m, ClearFlashCmd()

	case DependencyRemovedMsg:
		m.doorsView.RefreshDoors()
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after dependency removed: %v\n", err)
		}
		m.flash = "Dependency removed"
		return m, ClearFlashCmd()

	case DeferReturnTickMsg:
		returned := core.CheckDeferredReturnsWithTracker(m.pool, m.tracker)
		if returned > 0 {
			m.doorsView.RefreshDoors()
			if err := m.saveTasks(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save tasks after defer return: %v\n", err)
			}
		}
		return m, deferReturnTickCmd()

	case workerPollTickMsg:
		if m.dispatcher == nil || !m.hasDispatchedItems() {
			m.pollingActive = false
			return m, nil
		}
		return m, m.pollWorkerStatusCmd()

	case WorkerStatusMsg:
		cmd := m.handleWorkerStatus(msg)
		return m, cmd

	case ShowProposalsMsg:
		if m.proposalStore == nil {
			m.flash = "No proposal store available"
			return m, ClearFlashCmd()
		}
		m.proposalsView = NewProposalsView(m.proposalStore, m.pool, m.provider)
		m.proposalsView.SetWidth(m.width)
		m.proposalsView.SetHeight(m.contentHeight())
		m.previousView = m.viewMode
		m.viewMode = ViewProposals
		return m, nil

	case ProposalApprovedMsg:
		m.flash = fmt.Sprintf("Approved proposal for task %s", msg.TaskID)
		m.doorsView.SetPendingProposals(PendingProposalCount(m.proposalStore))
		return m, ClearFlashCmd()

	case ProposalRejectedMsg:
		m.flash = "Proposal rejected"
		m.doorsView.SetPendingProposals(PendingProposalCount(m.proposalStore))
		return m, ClearFlashCmd()

	case ProposalBatchApprovedMsg:
		m.flash = fmt.Sprintf("Approved %d proposals", msg.Count)
		m.doorsView.SetPendingProposals(PendingProposalCount(m.proposalStore))
		return m, ClearFlashCmd()

	case ShowHelpMsg:
		hv := NewHelpView()
		hv.SetWidth(m.width)
		hv.SetHeight(m.height)
		m.helpView = hv
		m.previousView = m.viewMode
		m.viewMode = ViewHelp
		return m, nil

	case ShowDeferredListMsg:
		m.deferredListView = NewDeferredListView(m.pool)
		m.deferredListView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewDeferred
		return m, nil

	case UnsnoozeTaskMsg:
		if err := msg.Task.UpdateStatus(core.StatusTodo); err != nil {
			m.flash = fmt.Sprintf("Cannot un-snooze: %v", err)
			return m, ClearFlashCmd()
		}
		msg.Task.DeferUntil = nil
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.flash = "Task un-snoozed — returned to todo"
		if m.deferredListView != nil {
			m.deferredListView.Refresh()
		}
		return m, ClearFlashCmd()

	case EditDeferDateMsg:
		m.snoozeView = NewSnoozeView(msg.Task)
		m.snoozeView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewSnooze
		return m, nil

	case ShowDevQueueMsg:
		if m.devQueue == nil || m.dispatcher == nil {
			m.flash = "Dev queue not available"
			return m, ClearFlashCmd()
		}
		m.devQueueView = NewDevQueueView(m.devQueue, m.dispatcher)
		m.devQueueView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewDevQueue
		return m, nil

	case DevDispatchRequestMsg:
		return m.handleDevDispatch(msg.Task)

	case DevDispatchResultMsg:
		if msg.Err != nil {
			m.flash = fmt.Sprintf("Dispatch failed: %s", msg.Err.Error())
		} else {
			m.flash = "Task queued for dev dispatch"
		}
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after dispatch: %v\n", err)
		}
		return m, ClearFlashCmd()

	case spinner.TickMsg:
		if m.syncSpinner != nil {
			cmd := m.syncSpinner.Update(msg)
			return m, cmd
		}
		return m, nil

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
					return m, m.syncSpinner.Tick()
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
		return m, nil
	}

	// When overlay is visible, intercept all keys.
	if m.showKeybindingOverlay {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.updateOverlay(keyMsg)
		}
		return m, nil
	}

	// Global '?' opens help from any non-text-input view
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "?" && !m.isTextInputActive() {
		m.showKeybindingOverlay = true
		m.overlayState = OverlayState{
			ScrollOffset: 0,
			ViewMode:     m.viewMode,
		}
		return m, nil
	}

	// Global 'h' toggles keybinding bar from any non-text-input view
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "h" && !m.isTextInputActive() {
		m.showKeybindingBar = !m.showKeybindingBar
		m.doorsView.SetHeight(m.contentHeight())
		if m.proposalsView != nil {
			m.proposalsView.SetHeight(m.contentHeight())
		}
		return m, m.saveKeybindingBarCmd(m.showKeybindingBar)
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
	case ViewConflict:
		return m.updateConflict(msg)
	case ViewSyncLog:
		return m.updateSyncLog(msg)
	case ViewThemePicker:
		return m.updateThemePicker(msg)
	case ViewDevQueue:
		return m.updateDevQueue(msg)
	case ViewProposals:
		return m.updateProposals(msg)
	case ViewHelp:
		return m.updateHelp(msg)
	case ViewDeferred:
		return m.updateDeferred(msg)
	case ViewSnooze:
		return m.updateSnooze(msg)
	}

	return m, nil
}

// updateOverlay handles key events when the keybinding overlay is visible.
// It intercepts all keys — only ?, esc, and scroll keys have behavior; the rest are consumed.
func (m *MainModel) updateOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc":
		m.showKeybindingOverlay = false
		m.overlayState.ScrollOffset = 0
		return m, nil
	case "up", "k":
		if m.overlayState.ScrollOffset > 0 {
			m.overlayState.ScrollOffset--
		}
		return m, nil
	case "down", "j":
		m.overlayState.ScrollOffset = ClampScrollOffset(m.overlayState.ScrollOffset+1, m.height)
		return m, nil
	}
	// All other keys consumed — no pass-through.
	return m, nil
}

// saveKeybindingBarCmd returns a tea.Cmd that persists the bar toggle to config.yaml.
func (m *MainModel) saveKeybindingBarCmd(show bool) tea.Cmd {
	configPath := m.configPath
	return func() tea.Msg {
		if configPath == "" {
			return nil
		}
		cfg, err := core.LoadProviderConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load config for bar toggle save: %v\n", err)
			return nil
		}
		cfg.ShowKeybindingBar = &show
		if err := core.SaveProviderConfig(configPath, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save bar toggle to config: %v\n", err)
		}
		return nil
	}
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
			if m.doorsView.selectedDoorIndex == 0 {
				m.doorsView.selectedDoorIndex = -1
			} else {
				m.doorsView.selectedDoorIndex = 0
			}
		case "w", "up":
			if m.doorsView.selectedDoorIndex == 1 {
				m.doorsView.selectedDoorIndex = -1
			} else {
				m.doorsView.selectedDoorIndex = 1
			}
		case "d", "right":
			if m.doorsView.selectedDoorIndex == 2 {
				m.doorsView.selectedDoorIndex = -1
			} else {
				m.doorsView.selectedDoorIndex = 2
			}
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
		case "enter", " ":
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
		case "S":
			return m, func() tea.Msg { return ShowProposalsMsg{} }
		case "m", "M":
			return m, func() tea.Msg { return ShowMoodMsg{} }
		case "/":
			m.searchView = m.newSearchView()
			m.searchView.SetWidth(m.width)
			m.viewMode = ViewSearch
			m.previousView = ViewDoors
			return m, nil
		case "z", "Z":
			if m.doorsView.selectedDoorIndex >= 0 && m.doorsView.selectedDoorIndex < len(m.doorsView.currentDoors) {
				task := m.doorsView.currentDoors[m.doorsView.selectedDoorIndex]
				return m, func() tea.Msg { return ShowSnoozeMsg{Task: task} }
			}
		case ":":
			m.searchView = m.newSearchView()
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

func (m *MainModel) updateConflict(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.conflictView == nil {
		return m, nil
	}
	cmd := m.conflictView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateThemePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.themePickerView == nil {
		return m, nil
	}
	cmd := m.themePickerView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateSyncLog(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.syncLogView == nil {
		return m, nil
	}
	cmd := m.syncLogView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.helpView == nil {
		return m, nil
	}
	cmd := m.helpView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateDevQueue(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.devQueueView == nil {
		return m, nil
	}
	cmd := m.devQueueView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateProposals(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.proposalsView == nil {
		return m, nil
	}
	cmd := m.proposalsView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateValues(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.valuesView == nil {
		return m, nil
	}
	cmd := m.valuesView.Update(msg)
	return m, cmd
}

// contentHeight returns the available height for view content, accounting for
// the keybinding bar when it is visible and the terminal is tall enough to show it.
func (m *MainModel) contentHeight() int {
	if m.showKeybindingBar && m.height >= barHeightHidden {
		return m.height - 2 // separator + bar
	}
	return m.height
}

// isTextInputActive returns true when the current view has an active text input
// field where 'q' should be treated as text, not as a quit command.
func (m *MainModel) isTextInputActive() bool {
	switch m.viewMode {
	case ViewSearch:
		return true
	case ViewAddTask:
		return true
	case ViewImprovement:
		return true
	case ViewOnboarding:
		// Onboarding has text input during the values step
		return true
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
		// When text input is focused, treat 'q' as text
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

func (m *MainModel) newDetailView(task *core.Task) *DetailView {
	dv := NewDetailView(task, m.tracker, m.enrichDB, m.pool)
	dv.SetWidth(m.width)
	dv.SetAgentService(m.agentService)
	if m.duplicateTaskIDs[task.ID] && m.dedupStore != nil {
		pair := m.findDuplicatePair(task.ID)
		dv.SetDuplicateInfo(true, m.dedupStore, pair)
	}
	if m.devDispatchEnabled && m.dispatcher != nil {
		available := m.dispatcher.CheckAvailable(context.Background()) == nil
		dv.SetDevDispatchInfo(true, available)
	}
	return dv
}

func (m *MainModel) newSearchView() *SearchView {
	sv := NewSearchView(m.pool, m.tracker, m.healthChecker, m.completionCounter, m.patternReport)
	sv.SetSyncLog(m.syncLog)
	sv.SetDuplicateTaskIDs(m.duplicateTaskIDs)
	if m.devDispatchEnabled && m.dispatcher != nil {
		if m.dispatcher.CheckAvailable(context.Background()) == nil {
			sv.SetDevDispatchEnabled(true)
		}
	}
	return sv
}

func (m *MainModel) saveTasks() error {
	allTasks := m.pool.GetAllTasks()
	return m.provider.SaveTasks(allTasks)
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

// workerPollTickCmd returns a tea.Cmd that fires a workerPollTickMsg after the poll interval.
func workerPollTickCmd() tea.Cmd {
	return tea.Tick(workerPollInterval, func(_ time.Time) tea.Msg {
		return workerPollTickMsg{}
	})
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

// View implements tea.Model.
func (m *MainModel) View() string {
	// Overlay takes over the entire screen when visible.
	if m.showKeybindingOverlay {
		return RenderKeybindingOverlay(m.overlayState, m.width, m.height)
	}

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
	case ViewDeferred:
		if m.deferredListView != nil {
			view = m.deferredListView.View()
		}
	case ViewSnooze:
		if m.snoozeView != nil {
			view = m.snoozeView.View()
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

	// Append keybinding bar when enabled.
	doorSelected := m.viewMode == ViewDoors && m.doorsView != nil && m.doorsView.selectedDoorIndex >= 0
	barOutput := RenderKeybindingBar(m.viewMode, m.width, m.height, m.showKeybindingBar, doorSelected)
	if barOutput != "" {
		view += "\n" + barOutput
	}

	return view
}
