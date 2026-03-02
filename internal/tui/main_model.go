package tui

import (
	"fmt"
	"os"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultWidth  = 80
	defaultHeight = 24
	doorCount     = 3
)

// ViewMode tracks which view is currently active.
type ViewMode int

const (
	ViewDoors ViewMode = iota
	ViewDetail
	ViewMood
)

// MainModel is the root Bubbletea model that orchestrates view transitions.
type MainModel struct {
	viewMode   ViewMode
	doorsView  *DoorsView
	detailView *DetailView
	moodView   *MoodView
	pool       *tasks.TaskPool
	tracker    *tasks.SessionTracker
	flash      string
	width      int
	height     int
}

// NewMainModel creates the root application model.
func NewMainModel(pool *tasks.TaskPool, tracker *tasks.SessionTracker) *MainModel {
	return &MainModel{
		viewMode:  ViewDoors,
		doorsView: NewDoorsView(pool, tracker),
		pool:      pool,
		tracker:   tracker,
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
		return m, nil

	case ClearFlashMsg:
		m.flash = ""
		return m, nil

	case ReturnToDoorsMsg:
		m.viewMode = ViewDoors
		m.detailView = nil
		m.moodView = nil
		m.doorsView.RefreshDoors()
		return m, nil

	case TaskCompletedMsg:
		m.doorsView.IncrementCompleted()
		if err := tasks.AppendCompleted(msg.Task); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to log completed task: %v\n", err)
		}
		m.pool.RemoveTask(msg.Task.ID)
		if err := m.saveTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks: %v\n", err)
		}
		m.flash = "Progress over perfection. Just pick one and start."
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

func (m *MainModel) saveTasks() error {
	allTasks := m.pool.GetAllTasks()
	return tasks.SaveTasks(allTasks)
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
	default:
		view = m.doorsView.View()
	}

	if m.flash != "" {
		view += "\n" + flashStyle.Render(m.flash)
	}

	return view
}
