package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// commandDef defines a command with its name and description for autocomplete.
type commandDef struct {
	Name string
	Desc string
}

// commandRegistry is the single source of truth for all available commands.
var commandRegistry = []commandDef{
	{"add", "Create a new task"},
	{"add-ctx", "Add task with context (why it matters)"},
	{"breakdown", "Break down a task into subtasks"},
	{"bug", "Report a bug"},
	{"dashboard", "Open insights dashboard"},
	{"deferred", "View deferred/snoozed tasks"},
	{"devqueue", "View dev dispatch queue"},
	{"dispatch", "Dev dispatch info"},
	{"enrich", "Enrich current task with LLM"},
	{"extract", "Extract tasks from text/file/clipboard"},
	{"goals", "View or edit values/goals"},
	{"health", "Run health check"},
	{"help", "Show help screen"},
	{"history", "View completed tasks"},
	{"import", "Import tasks from a file"},
	{"llm-status", "Show LLM backend status"},
	{"insights", "Show pattern insights"},
	{"mood", "Record current mood"},
	{"orphaned", "View orphaned tasks"},
	{"quit", "Quit application"},
	{"seasonal", "Pick seasonal theme (session only)"},
	{"stats", "Show session statistics"},
	{"connect", "Connect a new data source"},
	{"sources", "View connected data sources"},
	{"suggestions", "View AI task suggestions"},
	{"synclog", "View sync operation log"},
	{"tag", "Edit task categories/tags"},
	{"theme", "Open theme picker"},
}

// filterCommands returns commands whose names match the given prefix (case-insensitive).
func filterCommands(prefix string) []commandDef {
	prefix = strings.ToLower(prefix)
	var matched []commandDef
	for _, cmd := range commandRegistry {
		if strings.HasPrefix(cmd.Name, prefix) {
			matched = append(matched, cmd)
		}
	}
	return matched
}

// SearchView handles search and command palette functionality.
type SearchView struct {
	textInput            textinput.Model
	results              []*core.Task
	selectedIndex        int
	pool                 *core.TaskPool
	tracker              *core.SessionTracker
	healthChecker        *core.HealthChecker
	completionCounter    *core.CompletionCounter
	patternReport        *core.PatternReport
	syncLog              *core.SyncLog
	width                int
	isCommandMode        bool
	duplicateTaskIDs     map[string]bool
	devDispatchEnabled   bool
	commandSuggestions   []commandDef
	commandSelectedIndex int
	height               int
	hintEnabled          bool
	breadcrumbs          *BreadcrumbTrail
}

// NewSearchView creates a new SearchView.
func NewSearchView(pool *core.TaskPool, tracker *core.SessionTracker, hc *core.HealthChecker, cc *core.CompletionCounter, pr *core.PatternReport) *SearchView {
	ti := textinput.New()
	ti.Placeholder = "Search core... (or :command for commands)"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	return &SearchView{
		textInput:         ti,
		selectedIndex:     -1,
		pool:              pool,
		tracker:           tracker,
		healthChecker:     hc,
		completionCounter: cc,
		patternReport:     pr,
		duplicateTaskIDs:  make(map[string]bool),
	}
}

// SetInlineHints sets the inline hint display state.
func (sv *SearchView) SetInlineHints(enabled bool) {
	sv.hintEnabled = enabled
}

// SetHeight sets the terminal height for bottom-anchored rendering.
func (sv *SearchView) SetHeight(h int) {
	sv.height = h
}

// SetWidth sets the terminal width for rendering.
func (sv *SearchView) SetWidth(w int) {
	sv.width = w
	if w > 4 {
		sv.textInput.Width = w - 4
	}
}

// SetSyncLog sets the sync log for the :synclog command.
func (sv *SearchView) SetSyncLog(sl *core.SyncLog) {
	sv.syncLog = sl
}

// SetDuplicateTaskIDs sets the set of task IDs flagged as potential duplicates.
func (sv *SearchView) SetDuplicateTaskIDs(ids map[string]bool) {
	sv.duplicateTaskIDs = ids
}

// SetDevDispatchEnabled sets whether the :dispatch command is available.
func (sv *SearchView) SetDevDispatchEnabled(enabled bool) {
	sv.devDispatchEnabled = enabled
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
func (sv *SearchView) filterTasks(query string) []*core.Task {
	if query == "" {
		return nil
	}
	lowerQuery := strings.ToLower(query)
	allTasks := sv.pool.GetAllTasks()
	var matched []*core.Task
	for _, t := range allTasks {
		if strings.Contains(strings.ToLower(t.Text), lowerQuery) {
			matched = append(matched, t)
		}
	}
	slices.SortFunc(matched, func(a, b *core.Task) int {
		return strings.Compare(a.Text, b.Text)
	})
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

	// Record command name (without arguments) as breadcrumb.
	if cmd != "" && sv.breadcrumbs != nil {
		sv.breadcrumbs.Record("Search", "cmd:"+cmd)
	}

	switch cmd {
	case "add":
		if args == "" {
			return func() tea.Msg {
				return AddTaskPromptMsg{}
			}
		}
		if args == "--why" {
			return func() tea.Msg {
				return AddTaskWithContextPromptMsg{}
			}
		}
		if strings.HasPrefix(args, "--why ") {
			taskText := strings.TrimPrefix(args, "--why ")
			if taskText = strings.TrimSpace(taskText); taskText != "" {
				return func() tea.Msg {
					return AddTaskWithContextPromptMsg{PrefilledText: taskText}
				}
			}
			return func() tea.Msg {
				return AddTaskWithContextPromptMsg{}
			}
		}
		newTask := core.NewTask(args)
		return func() tea.Msg {
			return TaskAddedMsg{Task: newTask}
		}

	case "add-ctx":
		if args == "" {
			return func() tea.Msg {
				return AddTaskWithContextPromptMsg{}
			}
		}
		return func() tea.Msg {
			return AddTaskWithContextPromptMsg{PrefilledText: args}
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

	case "dashboard":
		return func() tea.Msg { return ShowInsightsMsg{} }

	case "connect":
		return func() tea.Msg { return ShowConnectWizardMsg{} }

	case "sources":
		return func() tea.Msg { return ShowSourcesMsg{} }

	case "insights":
		report := sv.patternReport
		switch args {
		case "mood":
			text := core.FormatMoodInsights(report)
			return func() tea.Msg { return FlashMsg{Text: text} }
		case "avoidance":
			text := core.FormatAvoidanceInsights(report)
			return func() tea.Msg { return FlashMsg{Text: text} }
		case "":
			// No args — open the full insights dashboard
			return func() tea.Msg { return ShowInsightsMsg{} }
		default:
			text := core.FormatInsights(report)
			return func() tea.Msg { return FlashMsg{Text: text} }
		}

	case "goals":
		if args == "edit" {
			return func() tea.Msg { return ShowValuesEditMsg{} }
		}
		return func() tea.Msg { return ShowValuesSetupMsg{} }

	case "orphaned":
		return func() tea.Msg { return ShowOrphanedMsg{} }

	case "synclog":
		return sv.showSyncLog()

	case "tag":
		return func() tea.Msg { return ShowTagViewMsg{} }

	case "seasonal":
		return func() tea.Msg { return ShowSeasonalPickerMsg{} }

	case "theme":
		return func() tea.Msg { return ShowThemePickerMsg{} }

	case "deferred", "snoozed":
		return func() tea.Msg { return ShowDeferredListMsg{} }

	case "devqueue":
		return func() tea.Msg { return ShowDevQueueMsg{} }

	case "suggestions":
		return func() tea.Msg { return ShowProposalsMsg{} }

	case "dispatch":
		if !sv.devDispatchEnabled {
			return func() tea.Msg {
				return FlashMsg{Text: "Dev dispatch is not enabled. Set dev_dispatch_enabled: true in config."}
			}
		}
		return func() tea.Msg {
			return FlashMsg{Text: "Use 'x' in task detail view to dispatch a specific task."}
		}

	case "import":
		return func() tea.Msg {
			return ShowImportMsg{PrefilledPath: args}
		}

	case "hints":
		return func() tea.Msg {
			return InlineHintsToggleMsg{Arg: args}
		}

	case "help":
		return func() tea.Msg {
			return ShowHelpMsg{}
		}

	case "history":
		return func() tea.Msg {
			return ShowHistoryMsg{}
		}

	case "plan":
		return func() tea.Msg {
			return ShowPlanningMsg{}
		}

	case "breakdown":
		return func() tea.Msg {
			return FlashMsg{Text: "Use 'g' in task detail view to break down a task."}
		}
	case "enrich":
		return func() tea.Msg { return EnrichCommandMsg{} }

	case "llm-status":
		return func() tea.Msg { return ShowLLMStatusMsg{} }

	case "extract":
		return func() tea.Msg { return ShowExtractMsg{} }

	case "bug":
		return func() tea.Msg { return ShowBugReportMsg{} }

	case "quit", "exit":
		return func() tea.Msg { return RequestQuitMsg{} }

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

	todayCount := 0
	yesterdayCount := 0
	streak := 0
	if sv.completionCounter != nil {
		todayCount = sv.completionCounter.GetTodayCount()
		yesterdayCount = sv.completionCounter.GetYesterdayCount()
		streak = sv.completionCounter.GetStreak()
	}

	text := fmt.Sprintf("Stats | Today: %d | Yesterday: %d | Doors: %d | Streak: %d days",
		todayCount, yesterdayCount, metrics.DetailViews, streak)
	return func() tea.Msg {
		return FlashMsg{Text: text}
	}
}

func (sv *SearchView) showSyncLog() tea.Cmd {
	if sv.syncLog == nil {
		return func() tea.Msg {
			return FlashMsg{Text: "Sync log not available"}
		}
	}
	entries, err := sv.syncLog.ReadRecentEntries(100)
	if err != nil {
		return func() tea.Msg {
			return FlashMsg{Text: fmt.Sprintf("Error reading sync log: %v", err)}
		}
	}
	return func() tea.Msg {
		return ShowSyncLogMsg{Entries: entries}
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
				// If a suggestion is highlighted and user hasn't typed a full command,
				// execute the highlighted suggestion
				if len(sv.commandSuggestions) > 0 && sv.commandSelectedIndex >= 0 &&
					sv.commandSelectedIndex < len(sv.commandSuggestions) {
					// Check if the current input matches the selected suggestion
					cmd, _ := parseCommand(sv.textInput.Value())
					selectedCmd := sv.commandSuggestions[sv.commandSelectedIndex].Name
					if cmd != selectedCmd {
						// Fill in the command and execute
						sv.textInput.SetValue(":" + selectedCmd)
					}
				}
				cmd := sv.executeCommand()
				sv.textInput.SetValue("")
				sv.isCommandMode = false
				sv.commandSuggestions = nil
				sv.commandSelectedIndex = 0
				return cmd
			}
			if sv.selectedIndex >= 0 && sv.selectedIndex < len(sv.results) {
				task := sv.results[sv.selectedIndex]
				return func() tea.Msg {
					return SearchResultSelectedMsg{Task: task}
				}
			}
			return nil

		case tea.KeyTab:
			if sv.isCommandMode && len(sv.commandSuggestions) > 0 &&
				sv.commandSelectedIndex >= 0 && sv.commandSelectedIndex < len(sv.commandSuggestions) {
				selected := sv.commandSuggestions[sv.commandSelectedIndex]
				sv.textInput.SetValue(":" + selected.Name)
				sv.textInput.SetCursor(len(sv.textInput.Value()))
				sv.updateCommandSuggestions()
				return nil
			}

		case tea.KeyUp:
			if sv.isCommandMode && len(sv.commandSuggestions) > 0 {
				if sv.commandSelectedIndex > 0 {
					sv.commandSelectedIndex--
				} else {
					sv.commandSelectedIndex = len(sv.commandSuggestions) - 1
				}
				return nil
			}
			if len(sv.results) > 0 && sv.selectedIndex > 0 {
				sv.selectedIndex--
			}
			return nil

		case tea.KeyDown:
			if sv.isCommandMode && len(sv.commandSuggestions) > 0 {
				if sv.commandSelectedIndex < len(sv.commandSuggestions)-1 {
					sv.commandSelectedIndex++
				} else {
					sv.commandSelectedIndex = 0
				}
				return nil
			}
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
	query := sv.textInput.Value()
	sv.checkCommandMode()
	if sv.isCommandMode {
		sv.updateCommandSuggestions()
	} else {
		sv.commandSuggestions = nil
		sv.commandSelectedIndex = 0
		sv.results = sv.filterTasks(query)
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

// updateCommandSuggestions refreshes the command suggestion list based on current input.
func (sv *SearchView) updateCommandSuggestions() {
	cmd, _ := parseCommand(sv.textInput.Value())
	sv.commandSuggestions = filterCommands(cmd)
	// Clamp selection index
	if sv.commandSelectedIndex >= len(sv.commandSuggestions) {
		if len(sv.commandSuggestions) > 0 {
			sv.commandSelectedIndex = len(sv.commandSuggestions) - 1
		} else {
			sv.commandSelectedIndex = 0
		}
	}
	// Default to first suggestion
	if sv.commandSelectedIndex < 0 && len(sv.commandSuggestions) > 0 {
		sv.commandSelectedIndex = 0
	}
}

// View renders the search view.
func (sv *SearchView) View() string {
	// Build the content section (middle area between header and input).
	var content strings.Builder

	query := sv.textInput.Value()

	// Content lines tracks how many lines the content area occupies.
	contentLines := 0

	if sv.isCommandMode {
		fmt.Fprintf(&content, "%s\n\n", commandModeStyle.Render("Command mode"))
		contentLines += 2

		// Render suggestions with fixed-height padding so input stays stable.
		maxSuggestions := len(commandRegistry)
		rendered := 0
		if len(sv.commandSuggestions) > 0 {
			for i, cmd := range sv.commandSuggestions {
				name := fmt.Sprintf("  :%-14s", cmd.Name)
				line := fmt.Sprintf("%s%s", name, cmd.Desc)
				if i == sv.commandSelectedIndex {
					fmt.Fprintf(&content, "%s\n", searchSelectedStyle.Render(line))
				} else {
					fmt.Fprintf(&content, "%s\n", searchResultStyle.Render(line))
				}
				rendered++
			}
		} else {
			fmt.Fprintf(&content, "%s\n", helpStyle.Render("No matching commands"))
			rendered++
		}
		// Pad remaining lines to keep total suggestion area at maxSuggestions + 1 (trailing blank).
		for i := rendered; i < maxSuggestions; i++ {
			content.WriteString("\n")
		}
		content.WriteString("\n")
		contentLines += maxSuggestions + 1
	} else if query != "" && len(sv.results) == 0 {
		fmt.Fprintf(&content, "%s\n\n", helpStyle.Render(fmt.Sprintf("No tasks match '%s'", query)))
		contentLines += 2
	} else if len(sv.results) > 0 {
		for i, task := range sv.results {
			statusColor := StatusColor(string(task.Status))
			statusIndicator := lipgloss.NewStyle().
				Foreground(statusColor).
				Render(fmt.Sprintf("[%s]", task.Status))

			srcBadge := SourceBadge(task.SourceProvider)
			dupBadge := ""
			if sv.duplicateTaskIDs[task.ID] {
				dupBadge = " " + DuplicateIndicator()
			}
			line := fmt.Sprintf("  %s %s %s%s", statusIndicator, task.Text, srcBadge, dupBadge)
			if i == sv.selectedIndex {
				line = searchSelectedStyle.Render(line)
			} else {
				line = searchResultStyle.Render(line)
			}
			fmt.Fprintf(&content, "%s\n", line)
			contentLines++
		}
		content.WriteString("\n")
		contentLines++
	}

	// Build the footer: input line + blank + help line = 3 lines.
	var footer strings.Builder
	fmt.Fprintf(&footer, "%s\n\n", sv.textInput.View())
	if sv.hintEnabled {
		h := func(key string) string { return renderInlineHint(key, sv.hintEnabled) }
		fmt.Fprintf(&footer, "%s", helpStyle.Render(h("↑/↓")+" navigate | "+h("enter")+" select | "+h("esc")+" close | : commands"))
	} else {
		fmt.Fprintf(&footer, "%s", helpStyle.Render("↑/↓ navigate | Enter select | Esc close | : commands"))
	}
	footerLines := 3

	// Assemble: header (2 lines) + padding + content + footer.
	var s strings.Builder
	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("ThreeDoors - Search"))
	headerLines := 2

	// Push content to the bottom by inserting blank lines.
	if sv.height > 0 {
		usedLines := headerLines + contentLines + footerLines
		padding := sv.height - usedLines
		if padding > 0 {
			for i := 0; i < padding; i++ {
				s.WriteString("\n")
			}
		}
	}

	fmt.Fprintf(&s, "%s", content.String())
	fmt.Fprintf(&s, "%s", footer.String())

	return s.String()
}
