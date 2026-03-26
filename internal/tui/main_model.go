package tui

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/core/connection"
	"github.com/arcavenae/ThreeDoors/internal/dispatch"
	"github.com/arcavenae/ThreeDoors/internal/enrichment"
	"github.com/arcavenae/ThreeDoors/internal/intelligence"
	"github.com/arcavenae/ThreeDoors/internal/intelligence/services"
	"github.com/arcavenae/ThreeDoors/internal/mcp"
	tea "github.com/charmbracelet/bubbletea"
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
	nextStepsView       *NextStepsView
	avoidancePromptView *AvoidancePromptView
	insightsView        *InsightsView
	onboardingView      *OnboardingView
	conflictView        *ConflictView
	syncLogView         *SyncLogView
	themePickerView     *ThemePicker
	devQueueView        *DevQueueView
	proposalsView       *ProposalsView
	helpView            *HelpView
	deferredListView    *DeferredListView
	snoozeView          *SnoozeView
	planningView        *PlanningView
	sourcesView         *SourcesView
	sourceDetailView    *SourceDetailView
	syncLogDetailView   *SyncLogDetailView
	connectWizard       *ConnectWizard
	disconnectDialog    *DisconnectDialog
	reauthDialog        *ReauthDialog
	importView          *ImportView
	bugReportView       *BugReportView
	breakdownView       *BreakdownView
	extractView         *ExtractView
	orphanedView        *OrphanedView
	historyView         *HistoryView
	completionReader    *core.CompletionReader
	breakdownService    *services.BreakdownService
	extractor           *services.TaskExtractor
	planningMode        bool // CLI --plan: exit after planning instead of showing doors
	planningTimestamp   *time.Time
	configPath          string
	pool                *core.TaskPool
	tracker             *core.SessionTracker
	provider            core.TaskProvider
	healthChecker       *core.HealthChecker
	completionCounter   *core.CompletionCounter
	patternReport       *core.PatternReport
	patternAnalyzer     *core.PatternAnalyzer
	enrichDB            *enrichment.DB
	valuesConfig        *core.ValuesConfig
	syncTracker         *core.SyncStatusTracker
	agentService        *intelligence.AgentService
	decomposing         bool
	enricher            *services.TaskEnricher
	enriching           bool

	syncLog               *core.SyncLog
	dedupStore            *core.DedupStore
	duplicateTaskIDs      map[string]bool
	duplicatePairs        []core.DuplicatePair
	devDispatchEnabled    bool
	dispatcher            dispatch.Dispatcher
	devQueue              *dispatch.DevQueue
	proposalStore         *mcp.ProposalStore
	connMgr               *connection.ConnectionManager
	connSvc               *connection.ConnectionService
	syncEventLog          *connection.SyncEventLog
	milestoneChecker      *core.MilestoneChecker
	pollingActive         bool
	syncSpinner           *SyncSpinner
	flash                 string
	width                 int
	height                int
	showKeyHints          bool
	showKeybindingOverlay bool
	keybindingOverlay     *KeybindingOverlay
	searchQuery           string
	searchSelectedIndex   int
	promptedTasks         map[string]bool
	breadcrumbs           BreadcrumbTrail
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

	// showKeyHints defaults to true (D-092: discoverable by default)

	// Initialize completion counter for daily tracking
	cc := core.NewCompletionCounter()
	var cr *core.CompletionReader
	if configPath, err := core.GetConfigDirPath(); err == nil {
		if loadErr := cc.LoadFromFile(filepath.Join(configPath, "completed.txt")); loadErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load completion history: %v\n", loadErr)
		}
		cr = core.NewCompletionReader(configPath)
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

	// Initialize milestone checker
	var mc *core.MilestoneChecker
	if configPath, err := core.GetConfigDirPath(); err == nil {
		mc = core.NewMilestoneChecker(configPath)
		if loadErr := mc.Load(); loadErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load milestones: %v\n", loadErr)
			mc = nil
		}
	}

	// Load planning timestamp for focus boost
	var planningTs *time.Time
	if configPath, err := core.GetConfigDirPath(); err == nil {
		planningTs = LoadPlanningTimestamp(configPath)
	}

	doorsView := NewDoorsView(pool, tracker)
	doorsView.SetAvoidanceData(patternReport)
	doorsView.SetInsightsData(pa, cc)
	doorsView.SetSyncTracker(syncTracker)
	doorsView.SetDuplicateTaskIDs(duplicateTaskIDs)
	if planningTs != nil {
		doorsView.SetPlanningTimestamp(planningTs)
	}
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
		milestoneChecker:  mc,
		completionReader:  cr,
		planningTimestamp: planningTs,
		promptedTasks:     make(map[string]bool),
		showKeyHints:      true, // default: hints visible (D-092)
	}

	doorsView.SetSyncSpinner(m.syncSpinner)
	doorsView.SetShowKeyHints(m.showKeyHints)

	if isFirstRun {
		m.onboardingView = NewOnboardingView()
		m.setViewMode(ViewOnboarding)
	}

	return m
}

// SetConfigPath sets the path to config.yaml for theme persistence.
func (m *MainModel) SetConfigPath(path string) {
	m.configPath = path
}

// SetShowKeyHints sets the initial key hints visibility from config.
func (m *MainModel) SetShowKeyHints(show bool) {
	m.showKeyHints = show
	if m.doorsView != nil {
		m.doorsView.SetShowKeyHints(show)
	}
}

// SetBaseThemeName sets the user's configured theme as the active theme and
// stores the name for seasonal fallback. Call before SetSeasonalEnabled.
func (m *MainModel) SetBaseThemeName(name string) {
	m.doorsView.SetBaseThemeName(name)
}

// SetSeasonalEnabled enables or disables automatic seasonal theme switching.
// When enabled, immediately resolves the seasonal theme for the current time.
func (m *MainModel) SetSeasonalEnabled(enabled bool) {
	m.doorsView.SetSeasonalEnabled(enabled)
	if enabled {
		m.doorsView.ResolveSeasonalTheme(time.Now().UTC())
	}
}

// SetDevDispatch configures dev dispatch with the given dispatcher, queue, and enabled flag.
func (m *MainModel) SetDevDispatch(enabled bool, d dispatch.Dispatcher, q *dispatch.DevQueue) {
	m.devDispatchEnabled = enabled
	m.dispatcher = d
	m.devQueue = q
}

// SetPlanningMode enables planning mode — launches directly into PlanningView
// and exits after planning completes (used by CLI `plan` subcommand).
func (m *MainModel) SetPlanningMode(enabled bool) {
	m.planningMode = enabled
	if enabled {
		pv := NewPlanningView(m.pool, m.provider)
		pv.SetWidth(m.width)
		pv.SetHeight(m.height)
		m.planningView = pv
		m.setViewMode(ViewPlanning)
	}
}

// SetAgentService sets the agent service for LLM task decomposition.
func (m *MainModel) SetAgentService(svc *intelligence.AgentService) {
	m.agentService = svc
}

// SetEnricher sets the task enricher for LLM task enrichment.
func (m *MainModel) SetEnricher(enricher *services.TaskEnricher) {
	m.enricher = enricher
}

// SetProposalStore sets the proposal store for the proposal review view.
func (m *MainModel) SetProposalStore(store *mcp.ProposalStore) {
	m.proposalStore = store
	if store != nil {
		count := PendingProposalCount(store)
		m.doorsView.SetPendingProposals(count)
	}
}

// SetConnectionManager sets the connection manager for the sources dashboard.
func (m *MainModel) SetConnectionManager(mgr *connection.ConnectionManager) {
	m.connMgr = mgr
}

// SetConnectionService sets the connection service for re-authentication.
func (m *MainModel) SetConnectionService(svc *connection.ConnectionService) {
	m.connSvc = svc
}

// SetSyncEventLog sets the sync event log for viewing connection-specific sync events.
func (m *MainModel) SetSyncEventLog(log *connection.SyncEventLog) {
	m.syncEventLog = log
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
	cmds := []tea.Cmd{deferReturnTickCmd()}
	if m.planningView != nil && m.viewMode == ViewPlanning {
		cmds = append(cmds, m.planningView.Init())
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model.
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle task management view messages (Planning, AddTask, Breakdown,
	// Extract, Import, Snooze, Deferred).
	if model, cmd, handled := m.handleTaskViewMessage(msg); handled {
		return model, cmd
	}

	// Handle source/sync view messages (Sources, SourceDetail, SyncLog,
	// SyncLogDetail, ConnectWizard, Disconnect, Reauth).
	if model, cmd, handled := m.handleSourceViewMessage(msg); handled {
		return model, cmd
	}

	// Handle auxiliary view messages (Help, BugReport, ThemePicker, Health,
	// Insights, Mood, Feedback, ValuesGoals, Orphaned, Conflict, etc.).
	if model, cmd, handled := m.handleAuxiliaryViewMessage(msg); handled {
		return model, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.breadcrumbs.Record(m.viewMode.String(), fmt.Sprintf("resize:%dx%d", msg.Width, msg.Height))
		m.width = msg.Width
		m.height = msg.Height
		m.doorsView.SetWidth(msg.Width)
		m.doorsView.SetHeight(m.contentHeight())
		contentH := m.contentHeight()
		m.resizeTaskViews(msg.Width, msg.Height, contentH)
		m.resizeSourceViews(msg.Width, msg.Height)
		m.resizeAuxiliaryViews(msg.Width, msg.Height, contentH)
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
			m.setViewMode(ViewSearch)
			m.detailView = nil
			m.addTaskView = nil
			m.previousView = ViewDoors
			return m, nil
		}
		m.setViewMode(ViewDoors)
		m.detailView = nil
		m.moodView = nil
		m.healthView = nil
		m.insightsView = nil
		m.sourcesView = nil
		m.disconnectDialog = nil
		m.reauthDialog = nil
		m.addTaskView = nil
		m.deferredListView = nil
		m.importView = nil
		m.bugReportView = nil
		m.doorsView.RefreshDoors()
		return m, nil

	case NavigateToLinkedMsg:
		m.detailView = m.newDetailView(msg.Task)
		m.setViewMode(ViewDetail)
		return m, nil

	case ReturnToSearchMsg:
		m.searchView = m.newSearchView()
		m.searchView.SetWidth(m.width)
		m.searchView.RestoreState(msg.Query, msg.SelectedIndex)
		m.setViewMode(ViewSearch)
		m.detailView = nil
		m.previousView = ViewDoors
		return m, nil

	case SearchClosedMsg:
		m.setViewMode(ViewDoors)
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
		m.setViewMode(ViewDetail)
		return m, nil

	case ExpandTaskMsg:
		newTask := core.NewTask(msg.NewTaskText)
		parentID := msg.ParentTask.ID
		newTask.ParentID = &parentID
		m.pool.AddTask(newTask)
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.flash = "Subtask added"
		return m, ClearFlashCmd()

	case TaskForkedMsg:
		m.pool.AddTask(msg.Variant)
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		if m.enrichDB != nil {
			ref := &enrichment.CrossReference{
				SourceTaskID: msg.Original.ID,
				TargetTaskID: msg.Variant.ID,
				SourceSystem: "local",
				Relationship: "forked-from",
			}
			if err := m.enrichDB.AddCrossReference(ref); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to create fork cross-reference: %v\n", err)
			}
		}
		m.flash = "Forked!"
		m.detailView = nil
		m.doorsView.RefreshDoors()
		m.setViewMode(ViewDoors)
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
		m.setViewMode(ViewNextSteps)
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
		m.setViewMode(ViewDoors)
		m.detailView = nil
		m.doorsView.RefreshDoors()
		m.flash = "Task uncompleted — returned to todo"
		return m, ClearFlashCmd()

	case TaskUpdatedMsg:
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.setViewMode(ViewDoors)
		m.detailView = nil
		m.doorsView.RefreshDoors()
		return m, nil

	case RequestQuitMsg:
		return m, tea.Quit
	}

	// Record non-text key events as breadcrumbs (privacy: never record tea.KeyRunes).
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type != tea.KeyRunes {
		m.breadcrumbs.Record(m.viewMode.String(), "key:"+keyMsg.String())
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
		m.keybindingOverlay = NewKeybindingOverlay(
			OverlayState{ViewMode: m.viewMode},
			m.width, m.height,
		)
		return m, nil
	}

	// Global 'h' toggles key hints from any non-text-input view
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "h" && !m.isTextInputActive() {
		m.showKeyHints = !m.showKeyHints
		m.doorsView.SetShowKeyHints(m.showKeyHints)
		m.doorsView.SetHeight(m.contentHeight())
		if m.proposalsView != nil {
			m.proposalsView.SetHeight(m.contentHeight())
		}
		return m, m.saveKeyHintsCmd(m.showKeyHints)
	}

	// Global ':' opens command mode from any non-text-input view
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == ":" && !m.isTextInputActive() {
		m.searchView = m.newSearchView()
		m.searchView.SetWidth(m.width)
		m.searchView.textInput.SetValue(":")
		m.searchView.checkCommandMode()
		m.previousView = m.viewMode
		m.setViewMode(ViewSearch)
		return m, nil
	}

	// Delegate to current view — auxiliary views first, then core views.
	if model, cmd, handled := m.updateAuxiliaryView(msg); handled {
		return model, cmd
	}
	switch m.viewMode {
	case ViewDoors:
		return m.updateDoors(msg)
	case ViewDetail:
		return m.updateDetail(msg)
	case ViewSearch:
		return m.updateSearch(msg)
	case ViewAddTask:
		return m.updateAddTask(msg)
	case ViewDeferred:
		return m.updateDeferred(msg)
	case ViewSnooze:
		return m.updateSnooze(msg)
	case ViewPlanning:
		return m.updatePlanning(msg)
	case ViewBreakdown:
		return m.updateBreakdown(msg)
	case ViewExtract:
		return m.updateExtract(msg)
	case ViewImport:
		return m.updateImport(msg)
	case ViewSyncLog:
		return m.updateSyncLog(msg)
	case ViewSources:
		return m.updateSources(msg)
	case ViewSourceDetail:
		return m.updateSourceDetail(msg)
	case ViewSyncLogDetail:
		return m.updateSyncLogDetail(msg)
	case ViewConnectWizard:
		return m.updateConnectWizard(msg)
	case ViewDisconnect:
		return m.updateDisconnect(msg)
	case ViewReauth:
		return m.updateReauth(msg)
	}

	return m, nil
}

func (m *MainModel) updateDoors(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case animationFrameMsg:
		if m.doorsView.doorAnimation != nil && m.doorsView.doorAnimation.Update() {
			return m, AnimationTickCmd()
		}
		return m, nil
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
			m.doorsView.doorAnimation.SetSelection(m.doorsView.selectedDoorIndex)
			return m, AnimationTickCmd()
		case "w", "up":
			if m.doorsView.selectedDoorIndex == 1 {
				m.doorsView.selectedDoorIndex = -1
			} else {
				m.doorsView.selectedDoorIndex = 1
			}
			m.doorsView.doorAnimation.SetSelection(m.doorsView.selectedDoorIndex)
			return m, AnimationTickCmd()
		case "d", "right":
			if m.doorsView.selectedDoorIndex == 2 {
				m.doorsView.selectedDoorIndex = -1
			} else {
				m.doorsView.selectedDoorIndex = 2
			}
			m.doorsView.doorAnimation.SetSelection(m.doorsView.selectedDoorIndex)
			return m, AnimationTickCmd()
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
				m.setViewMode(ViewDetail)
			}
		case "n", "N":
			if m.doorsView.selectedDoorIndex >= 0 && m.doorsView.selectedDoorIndex < len(m.doorsView.currentDoors) {
				task := m.doorsView.currentDoors[m.doorsView.selectedDoorIndex]
				return m, func() tea.Msg { return ShowFeedbackMsg{Task: task} }
			}
		case "H":
			return m, func() tea.Msg { return ShowHistoryMsg{} }
		case "S":
			return m, func() tea.Msg { return ShowProposalsMsg{} }
		case "m", "M":
			return m, func() tea.Msg { return ShowMoodMsg{} }
		case "/":
			m.searchView = m.newSearchView()
			m.searchView.SetWidth(m.width)
			m.setViewMode(ViewSearch)
			m.previousView = ViewDoors
			return m, nil
		case "z", "Z":
			if m.doorsView.selectedDoorIndex >= 0 && m.doorsView.selectedDoorIndex < len(m.doorsView.currentDoors) {
				task := m.doorsView.currentDoors[m.doorsView.selectedDoorIndex]
				return m, func() tea.Msg { return ShowSnoozeMsg{Task: task} }
			}
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

func (m *MainModel) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.searchView == nil {
		return m, nil
	}
	cmd := m.searchView.Update(msg)
	return m, cmd
}

func (m *MainModel) newDetailView(task *core.Task) *DetailView {
	dv := NewDetailView(task, m.tracker, m.enrichDB, m.pool)
	dv.SetWidth(m.width)
	dv.SetAgentService(m.agentService)
	dv.SetEnricher(m.enricher)
	dv.SetInlineHints(m.resolveHints())
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
	sv.SetInlineHints(m.resolveHints())
	sv.SetHeight(m.height)
	sv.breadcrumbs = &m.breadcrumbs
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

// View implements tea.Model.
func (m *MainModel) View() string {
	// Overlay takes over the entire screen when visible.
	if m.showKeybindingOverlay && m.keybindingOverlay != nil {
		return m.keybindingOverlay.View()
	}

	view, showValuesFooter := m.currentViewContent()

	if m.flash != "" {
		view += "\n" + flashStyle.Render(m.flash)
	}

	if showValuesFooter {
		view += RenderValuesFooter(m.valuesConfig)
	}

	// Build footer: keybinding bar for non-doors views.
	var footer string
	if m.viewMode != ViewDoors {
		barCtx := m.buildBarContext()
		barOutput := RenderKeybindingBarWithContext(barCtx, m.width, m.height, m.showKeyHints)
		if barOutput != "" {
			footer = barOutput
		}
	}

	// Pad output to exactly m.height lines so the TUI fills the terminal.
	return layoutFull("", view, footer, m.height)
}
