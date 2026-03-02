package tui

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"

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
	pool                *tasks.TaskPool
	tracker             *tasks.SessionTracker
	provider            tasks.TaskProvider
	healthChecker       *tasks.HealthChecker
	completionCounter   *tasks.CompletionCounter
	patternReport       *tasks.PatternReport
	valuesConfig        *tasks.ValuesConfig
	flash               string
	width               int
	height              int
	searchQuery         string
	searchSelectedIndex int
}

// NewMainModel creates the root application model.
func NewMainModel(pool *tasks.TaskPool, tracker *tasks.SessionTracker, provider tasks.TaskProvider, hc *tasks.HealthChecker) *MainModel {
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

	// Load cached pattern report (non-blocking — analysis runs in main.go goroutine)
	var patternReport *tasks.PatternReport
	if configPath, err := tasks.GetConfigDirPath(); err == nil {
		analyzer := tasks.NewPatternAnalyzer()
		patternReport, _ = analyzer.LoadPatterns(filepath.Join(configPath, "patterns.json"))
	}

	doorsView := NewDoorsView(pool, tracker)
	doorsView.SetAvoidanceData(patternReport)

	return &MainModel{
		viewMode:          ViewDoors,
		doorsView:         doorsView,
		pool:              pool,
		tracker:           tracker,
		provider:          provider,
		healthChecker:     hc,
		completionCounter: cc,
		patternReport:     patternReport,
		valuesConfig:      valuesConfig,
	}
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
		m.addTaskView = nil
		m.doorsView.RefreshDoors()
		return m, nil

	case HealthCheckMsg:
		m.healthView = NewHealthView(msg.Result)
		m.healthView.SetWidth(m.width)
		m.previousView = m.viewMode
		m.viewMode = ViewHealth
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
		m.detailView = NewDetailView(msg.Task, m.tracker)
		m.detailView.SetWidth(m.width)
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

	case FlashMsg:
		m.flash = msg.Text
		return m, ClearFlashCmd()
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
			m.flash = doorRefreshMessages[rand.IntN(len(doorRefreshMessages))]
			return m, ClearFlashCmd()
		case "enter":
			if m.doorsView.selectedDoorIndex >= 0 && m.doorsView.selectedDoorIndex < len(m.doorsView.currentDoors) {
				task := m.doorsView.currentDoors[m.doorsView.selectedDoorIndex]
				if m.tracker != nil {
					m.tracker.RecordDoorSelection(m.doorsView.selectedDoorIndex, task.Text)
				}
				m.detailView = NewDetailView(task, m.tracker)
				m.detailView.SetWidth(m.width)
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

func (m *MainModel) updateValues(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.valuesView == nil {
		return m, nil
	}
	cmd := m.valuesView.Update(msg)
	return m, cmd
}

func (m *MainModel) saveTasks() error {
	allTasks := m.pool.GetAllTasks()
	return m.provider.SaveTasks(allTasks)
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
