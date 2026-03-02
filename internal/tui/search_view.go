package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchView handles search and command palette functionality.
type SearchView struct {
	textInput     textinput.Model
	results       []*tasks.Task
	selectedIndex int
	pool          *tasks.TaskPool
	tracker       *tasks.SessionTracker
	healthChecker *tasks.HealthChecker
	width         int
	isCommandMode bool
}

// NewSearchView creates a new SearchView.
func NewSearchView(pool *tasks.TaskPool, tracker *tasks.SessionTracker, hc *tasks.HealthChecker) *SearchView {
	ti := textinput.New()
	ti.Placeholder = "Search tasks... (or :command for commands)"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	return &SearchView{
		textInput:     ti,
		selectedIndex: -1,
		pool:          pool,
		tracker:       tracker,
		healthChecker: hc,
	}
}

// SetWidth sets the terminal width for rendering.
func (sv *SearchView) SetWidth(w int) {
	sv.width = w
	if w > 4 {
		sv.textInput.Width = w - 4
	}
}

// RestoreState restores search state after returning from detail view.
func (sv *SearchView) RestoreState(query string, selectedIndex int) {
	sv.textInput.SetValue(query)
	sv.results = sv.filterTasks(query)
	sv.selectedIndex = selectedIndex
	if sv.selectedIndex >= len(sv.results) {
		sv.selectedIndex = len(sv.results) - 1
	}
}

// filterTasks returns tasks matching query by case-insensitive substring match.
func (sv *SearchView) filterTasks(query string) []*tasks.Task {
	if query == "" {
		return nil
	}
	lowerQuery := strings.ToLower(query)
	allTasks := sv.pool.GetAllTasks()
	var matched []*tasks.Task
	for _, t := range allTasks {
		if strings.Contains(strings.ToLower(t.Text), lowerQuery) {
			matched = append(matched, t)
		}
	}
	return matched
}

// checkCommandMode updates isCommandMode based on input.
func (sv *SearchView) checkCommandMode() {
	sv.isCommandMode = strings.HasPrefix(sv.textInput.Value(), ":")
}

// parseCommand splits a command string into command name and arguments.
func parseCommand(input string) (string, string) {
	input = strings.TrimPrefix(input, ":")
	parts := strings.SplitN(input, " ", 2)
	cmd := strings.ToLower(strings.TrimSpace(parts[0]))
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}
	return cmd, args
}

// executeCommand processes a command from the input.
func (sv *SearchView) executeCommand() tea.Cmd {
	cmd, args := parseCommand(sv.textInput.Value())

	switch cmd {
	case "add":
		if args == "" {
			return func() tea.Msg {
				return FlashMsg{Text: "Usage: :add <task text>"}
			}
		}
		newTask := tasks.NewTask(args)
		return func() tea.Msg {
			return TaskAddedMsg{Task: newTask}
		}

	case "mood":
		if args != "" {
			return func() tea.Msg {
				return MoodCapturedMsg{Mood: args}
			}
		}
		return func() tea.Msg {
			return ShowMoodMsg{}
		}

	case "stats":
		return sv.showStats()

	case "health":
		return sv.runHealthCheck()

	case "help":
		return func() tea.Msg {
			return FlashMsg{Text: "Commands: :add <text>, :mood [mood], :stats, :health, :help, :quit | Keys: / search, a/w/d select, s re-roll, Enter open, m mood, q quit"}
		}

	case "quit", "exit":
		return tea.Quit

	case "":
		return nil

	default:
		return func() tea.Msg {
			return FlashMsg{Text: fmt.Sprintf("Unknown command: '%s'. Type :help for available commands.", cmd)}
		}
	}
}

func (sv *SearchView) runHealthCheck() tea.Cmd {
	if sv.healthChecker == nil {
		return func() tea.Msg {
			return FlashMsg{Text: "Health check not available"}
		}
	}
	return func() tea.Msg {
		result := sv.healthChecker.RunAll()
		return HealthCheckMsg{Result: result}
	}
}

func (sv *SearchView) showStats() tea.Cmd {
	if sv.tracker == nil {
		return func() tea.Msg {
			return FlashMsg{Text: "Session stats: No tracker available"}
		}
	}
	metrics := sv.tracker.Finalize()
	text := fmt.Sprintf("Session Stats | Completed: %d | Doors viewed: %d | Refreshes: %d",
		metrics.TasksCompleted, metrics.DetailViews, metrics.RefreshesUsed)
	return func() tea.Msg {
		return FlashMsg{Text: text}
	}
}

// Update handles messages for the search view.
func (sv *SearchView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return func() tea.Msg { return SearchClosedMsg{} }

		case tea.KeyEnter:
			if sv.isCommandMode {
				cmd := sv.executeCommand()
				sv.textInput.SetValue("")
				sv.isCommandMode = false
				return cmd
			}
			if sv.selectedIndex >= 0 && sv.selectedIndex < len(sv.results) {
				task := sv.results[sv.selectedIndex]
				return func() tea.Msg {
					return SearchResultSelectedMsg{Task: task}
				}
			}
			return nil

		case tea.KeyUp:
			if len(sv.results) > 0 && sv.selectedIndex > 0 {
				sv.selectedIndex--
			}
			return nil

		case tea.KeyDown:
			if len(sv.results) > 0 {
				if sv.selectedIndex < len(sv.results)-1 {
					sv.selectedIndex++
				}
			}
			return nil

		default:
			// Check for j/k vi-style navigation
			if msg.Type == tea.KeyRunes {
				r := string(msg.Runes)
				if r == "j" && !sv.isCommandMode && sv.textInput.Value() == "" {
					if len(sv.results) > 0 && sv.selectedIndex < len(sv.results)-1 {
						sv.selectedIndex++
					}
					return nil
				}
				if r == "k" && !sv.isCommandMode && sv.textInput.Value() == "" {
					if len(sv.results) > 0 && sv.selectedIndex > 0 {
						sv.selectedIndex--
					}
					return nil
				}
			}
		}
	}

	// Delegate to textinput for typing, cursor, etc.
	var cmd tea.Cmd
	sv.textInput, cmd = sv.textInput.Update(msg)

	// Update search results based on current input
	oldQuery := sv.textInput.Value()
	sv.checkCommandMode()
	if !sv.isCommandMode {
		sv.results = sv.filterTasks(oldQuery)
		// Reset selection when results change
		if len(sv.results) > 0 {
			if sv.selectedIndex < 0 {
				sv.selectedIndex = 0
			}
			if sv.selectedIndex >= len(sv.results) {
				sv.selectedIndex = len(sv.results) - 1
			}
		} else {
			sv.selectedIndex = -1
		}
	}

	return cmd
}

// View renders the search view.
func (sv *SearchView) View() string {
	s := strings.Builder{}

	s.WriteString(headerStyle.Render("ThreeDoors - Search"))
	s.WriteString("\n\n")

	query := sv.textInput.Value()

	// Render results (bottom-up: best match closest to input)
	if sv.isCommandMode {
		s.WriteString(commandModeStyle.Render("Command mode"))
		s.WriteString("\n\n")
	} else if query != "" && len(sv.results) == 0 {
		s.WriteString(helpStyle.Render(fmt.Sprintf("No tasks match '%s'", query)))
		s.WriteString("\n\n")
	} else if len(sv.results) > 0 {
		// Render results top to bottom (bottom-up display: last result closest to input)
		for i, task := range sv.results {
			statusColor := StatusColor(string(task.Status))
			statusIndicator := lipgloss.NewStyle().
				Foreground(statusColor).
				Render(fmt.Sprintf("[%s]", task.Status))

			line := fmt.Sprintf("  %s %s", statusIndicator, task.Text)
			if i == sv.selectedIndex {
				line = searchSelectedStyle.Render(line)
			} else {
				line = searchResultStyle.Render(line)
			}
			s.WriteString(line)
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}

	// Render input at bottom
	s.WriteString(sv.textInput.View())
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("↑/↓ navigate | Enter select | Esc close | : commands"))

	return s.String()
}
