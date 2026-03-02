package tui

import (
	"fmt"
	"math/rand/v2"
	"os"

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
)

// MainModel is the root Bubbletea model that orchestrates view transitions.
type MainModel struct {
	viewMode            ViewMode
	previousView        ViewMode
	doorsView           *DoorsView
	detailView          *DetailView
	moodView            *MoodView
	searchView          *SearchView
	pool                *tasks.TaskPool
	tracker             *tasks.SessionTracker
	provider            tasks.TaskProvider
	flash               string
	width               int
	height              int
	searchQuery         string
	searchSelectedIndex int
}

// NewMainModel creates the root application model.
func NewMainModel(pool *tasks.TaskPool, tracker *tasks.SessionTracker, provider tasks.TaskProvider) *MainModel {
	return &MainModel{
		viewMode:  ViewDoors,
		doorsView: NewDoorsView(pool, tracker),
		pool:      pool,
		tracker:   tracker,
		provider:  provider,
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
		return m, nil

	case ClearFlashMsg:
		m.flash = ""
		return m, nil

	case ReturnToDoorsMsg:
		// If we came from search, return to search instead
		if m.previousView == ViewSearch {
			m.searchView = NewSearchView(m.pool, m.tracker)
			m.searchView.SetWidth(m.width)
			m.searchView.RestoreState(m.searchQuery, m.searchSelectedIndex)
			m.viewMode = ViewSearch
			m.detailView = nil
			m.previousView = ViewDoors
			return m, nil
		}
		m.viewMode = ViewDoors
		m.detailView = nil
		m.moodView = nil
		m.doorsView.RefreshDoors()
		return m, nil

	case ReturnToSearchMsg:
		m.searchView = NewSearchView(m.pool, m.tracker)
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

	case TaskAddedMsg:
		m.pool.AddTask(msg.Task)
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.flash = "Task added"
		return m, ClearFlashCmd()

	case TaskCompletedMsg:
		if err := m.provider.MarkComplete(msg.Task.ID); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to mark complete: %v\n", err)
			m.flash = "Error completing task"
			return m, ClearFlashCmd()
		}
		m.pool.RemoveTask(msg.Task.ID)
		m.doorsView.IncrementCompleted()
		m.flash = celebrationMessages[rand.IntN(len(celebrationMessages))]
		m.viewMode = ViewDoors
		m.detailView = nil
		m.doorsView.RefreshDoors()
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
	}

	return m, nil
}

func (m *MainModel) updateDoors(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
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
		case "m", "M":
			return m, func() tea.Msg { return ShowMoodMsg{} }
		case "/":
			m.searchView = NewSearchView(m.pool, m.tracker)
			m.searchView.SetWidth(m.width)
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

func (m *MainModel) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.searchView == nil {
		return m, nil
	}
	cmd := m.searchView.Update(msg)
	return m, cmd
}

func (m *MainModel) saveTasks() error {
	allTasks := m.pool.GetAllTasks()
	return m.provider.SaveTasks(allTasks)
}

// View implements tea.Model.
func (m *MainModel) View() string {
	var view string
	switch m.viewMode {
	case ViewDetail:
		if m.detailView != nil {
			view = m.detailView.View()
		}
	case ViewMood:
		if m.moodView != nil {
			view = m.moodView.View()
		}
	case ViewSearch:
		if m.searchView != nil {
			view = m.searchView.View()
		}
	default:
		view = m.doorsView.View()
	}

	if m.flash != "" {
		view += "\n" + flashStyle.Render(m.flash)
	}

	return view
}
