package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/arcaven/ThreeDoors/internal/tasks"
)

const (
	defaultWidth  = 80
	defaultHeight = 24
	doorCount     = 3
)

// Model is the root Bubbletea model for the application.
type Model struct {
	quitting    bool
	tasks       []tasks.Task
	doors       []tasks.Task
	selectedIdx int
	width       int
	height      int
	err         error
	taskLoader  tasks.TaskLoader
}

// tasksLoadedMsg is sent when tasks are successfully loaded from file.
type tasksLoadedMsg struct {
	tasks []tasks.Task
}

// tasksLoadErrorMsg is sent when task loading fails.
type tasksLoadErrorMsg struct {
	err error
}

// NewModel creates a Model with default FileManager (~/.threedoors).
func NewModel() Model {
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".threedoors")
	return Model{
		selectedIdx: -1,
		width:       defaultWidth,
		height:      defaultHeight,
		taskLoader:  tasks.NewFileManager(baseDir),
	}
}

// NewModelWithTasks creates a Model pre-loaded with tasks (for testing).
func NewModelWithTasks(taskList []tasks.Task) Model {
	m := Model{
		selectedIdx: -1,
		width:       defaultWidth,
		height:      defaultHeight,
		tasks:       taskList,
	}
	m.doors = tasks.SelectRandomDoors(m.tasks, doorCount, nil)
	return m
}

// Init returns a command to load tasks from the file system.
func (m Model) Init() tea.Cmd {
	if m.taskLoader == nil {
		return nil
	}
	loader := m.taskLoader
	return func() tea.Msg {
		loadedTasks, err := loader.LoadTasks()
		if err != nil {
			return tasksLoadErrorMsg{err: err}
		}
		return tasksLoadedMsg{tasks: loadedTasks}
	}
}

// Update handles all messages including key presses, window resizing, and task loading.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		m.tasks = msg.tasks
		m.doors = tasks.SelectRandomDoors(m.tasks, doorCount, nil)
		m.selectedIdx = -1
		return m, nil

	case tasksLoadErrorMsg:
		m.err = msg.err
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyLeft:
		m.selectedIdx = 0
		return m, nil

	case tea.KeyUp:
		m.selectedIdx = 1
		return m, nil

	case tea.KeyRight:
		m.selectedIdx = 2
		return m, nil

	case tea.KeyDown:
		m.rerollDoors()
		return m, nil

	case tea.KeyRunes:
		return m.handleRuneKey(string(msg.Runes))
	}

	return m, nil
}

func (m Model) handleRuneKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "a":
		m.selectedIdx = 0
	case "w":
		m.selectedIdx = 1
	case "d":
		m.selectedIdx = 2
	case "s":
		m.rerollDoors()
	case "c", "b", "i", "e", "f", "p":
		// Future task management keys — no-op for Story 1.2
	}
	return m, nil
}

func (m *Model) rerollDoors() {
	m.doors = tasks.SelectRandomDoors(m.tasks, doorCount, m.doors)
	m.selectedIdx = -1
}

// View renders the three doors display.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error loading tasks: %s\n\nPress q to quit\n", m.err)
	}

	if len(m.tasks) == 0 && m.taskLoader != nil {
		// Still loading or no tasks
		return "Loading tasks...\n"
	}

	if len(m.tasks) == 0 {
		return "No tasks found. Add tasks to ~/.threedoors/tasks.txt\n\nPress q to quit\n"
	}

	if len(m.doors) == 0 {
		return "No tasks found. Add tasks to ~/.threedoors/tasks.txt\n\nPress q to quit\n"
	}

	return m.renderDoors()
}

func (m Model) renderDoors() string {
	doorWidth := m.width / len(m.doors)
	if doorWidth < 10 {
		doorWidth = 10
	}
	// Account for border (2 chars) and padding (2 chars each side)
	contentWidth := doorWidth - 6
	if contentWidth < 4 {
		contentWidth = 4
	}

	unselectedStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(doorWidth)

	selectedStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Bold(true).
		Padding(1, 2).
		Width(doorWidth)

	var renderedDoors []string
	for i, door := range m.doors {
		text := door.Text
		if len(text) > contentWidth {
			text = text[:contentWidth-3] + "..."
		}

		style := unselectedStyle
		if i == m.selectedIdx {
			style = selectedStyle
		}
		renderedDoors = append(renderedDoors, style.Render(text))
	}

	header := lipgloss.NewStyle().Bold(true).Render("ThreeDoors - Technical Demo")
	doorsRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedDoors...)

	hint := "\n[a/←] left  [w/↑] center  [d/→] right  [s/↓] re-roll  [q] quit"

	return header + "\n\n" + doorsRow + hint + "\n"
}
