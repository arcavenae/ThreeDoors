package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// model represents the application's state.
type model struct {
	allTasks          []tasks.Task
	displayedDoors    []tasks.Task
	selectedDoorIndex int // -1 for no selection, 0 for left, 1 for center, 2 for right
	width             int
}

// getThreeRandomDoors selects 3 random tasks from the provided allTasks pool.
func getThreeRandomDoors(allTasks []tasks.Task) []tasks.Task {
	if len(allTasks) < 3 {
		panic("Not enough tasks loaded to display three doors.")
	}
	displayedDoors := make([]tasks.Task, 3)
	perm := rand.Perm(len(allTasks))
	for i := 0; i < 3; i++ {
		displayedDoors[i] = allTasks[perm[i]]
	}
	return displayedDoors
}

// initialModel initializes the application's model.
func initialModel() model {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	allTasks, err := tasks.LoadTasks()
	if err != nil {
		// In a real application, you might want to handle this error more gracefully,
		// perhaps by showing an error screen or logging it. For now, we'll panic.
		panic(fmt.Sprintf("Failed to load tasks: %v", err))
	}

	displayedDoors := getThreeRandomDoors(allTasks)

	return model{
		allTasks:          allTasks,
		displayedDoors:    displayedDoors,
		selectedDoorIndex: -1, // Initialize to no door selected
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "a", "left": // Select left door
			m.selectedDoorIndex = 0
		case "w", "up": // Select center door
			m.selectedDoorIndex = 1
		case "d", "right": // Select right door
			m.selectedDoorIndex = 2
		case "s", "down": // Re-roll tasks
			m.displayedDoors = getThreeRandomDoors(m.allTasks)
			m.selectedDoorIndex = -1 // Reset selection after re-rolling
		case "c": // Mark selected task as complete
			// TODO: Implement task completion logic
		case "b": // Mark selected task as blocked
			// TODO: Implement task blocking logic
		case "i": // Mark selected task as in progress
			// TODO: Implement task in progress logic
		case "e": // Expand selected task (into more tasks)
			// TODO: Implement task expansion logic
		case "f": // Fork selected task (clone/split)
			// TODO: Implement task forking logic
		case "p": // Procrastinate/avoid selected task
			// TODO: Implement task procrastination logic
		}
	}
	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}
	s.WriteString("ThreeDoors - Technical Demo\n\n")

	// Calculate door width based on terminal width
	doorWidth := (m.width - 6) / 3 // m.width - 6 for padding/borders, then divide by 3 doors

	unselectedDoorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(doorWidth)

	selectedDoorStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("86")). // Green color for selected
		Padding(1, 2).
		Width(doorWidth)

	var renderedDoors []string
	for i, task := range m.displayedDoors {
		doorContent := task.Text // Removed "Door X:" label
		if i == m.selectedDoorIndex {
			renderedDoors = append(renderedDoors, selectedDoorStyle.Render(doorContent))
		} else {
			renderedDoors = append(renderedDoors, unselectedDoorStyle.Render(doorContent))
		}
	}

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedDoors...))
	s.WriteString("\n\nPress 'q' or 'ctrl+c' to quit.\n")
	s.WriteString("Use 'a'/'left', 'w'/'up', 'd'/'right' to select doors, 's'/'down' to re-roll.\n")
	s.WriteString("Use 'c' (complete), 'b' (blocked), 'i' (in progress), 'e' (expand), 'f' (fork), 'p' (procrastinate) for task management.\n")
	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
