package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Test Helpers ---

func newTestSearchView(texts ...string) *SearchView {
	pool := makePool(texts...)
	return NewSearchView(pool, nil)
}

func newTestSearchViewWithTracker(texts ...string) (*SearchView, *tasks.SessionTracker) {
	pool := makePool(texts...)
	tracker := tasks.NewSessionTracker()
	return NewSearchView(pool, tracker), tracker
}

// --- filterTasks Tests ---

func TestSearchView_FilterTasks_ExactMatch(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	results := sv.filterTasks("Write unit tests")
	if len(results) != 1 {
		t.Errorf("expected 1 match, got %d", len(results))
	}
	if results[0].Text != "Write unit tests" {
		t.Errorf("expected 'Write unit tests', got %q", results[0].Text)
	}
}

func TestSearchView_FilterTasks_PartialMatch(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	results := sv.filterTasks("unit")
	if len(results) != 1 {
		t.Errorf("expected 1 match for 'unit', got %d", len(results))
	}
}

func TestSearchView_FilterTasks_CaseInsensitive(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	results := sv.filterTasks("fix")
	if len(results) != 1 {
		t.Errorf("expected 1 match for 'fix' (case-insensitive), got %d", len(results))
	}
}

func TestSearchView_FilterTasks_MultipleMatches(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Write docs", "Fix bug")
	results := sv.filterTasks("Write")
	if len(results) != 2 {
		t.Errorf("expected 2 matches for 'Write', got %d", len(results))
	}
}

func TestSearchView_FilterTasks_NoMatch(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	results := sv.filterTasks("nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 matches, got %d", len(results))
	}
}

func TestSearchView_FilterTasks_EmptyQuery(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	results := sv.filterTasks("")
	if len(results) != 0 {
		t.Errorf("expected 0 matches for empty query, got %d", len(results))
	}
}

func TestSearchView_FilterTasks_SpecialCharacters(t *testing.T) {
	sv := newTestSearchView("Fix [bug] in (parser)", "Test task")
	results := sv.filterTasks("[bug]")
	if len(results) != 1 {
		t.Errorf("expected 1 match for '[bug]' (literal match, not regex), got %d", len(results))
	}
}

func TestSearchView_FilterTasks_AllStatuses(t *testing.T) {
	pool := tasks.NewTaskPool()
	t1 := tasks.NewTask("todo task")
	t2 := tasks.NewTask("blocked task")
	_ = t2.UpdateStatus(tasks.StatusBlocked)
	t3 := tasks.NewTask("in-progress task")
	_ = t3.UpdateStatus(tasks.StatusInProgress)
	pool.AddTask(t1)
	pool.AddTask(t2)
	pool.AddTask(t3)

	sv := NewSearchView(pool, nil)
	results := sv.filterTasks("task")
	if len(results) != 3 {
		t.Errorf("expected 3 matches (all statuses searched), got %d", len(results))
	}
}

// --- Command Parsing Tests ---

func TestSearchView_ParseCommand_AddWithText(t *testing.T) {
	cmd, args := parseCommand(":add Buy groceries")
	if cmd != "add" {
		t.Errorf("expected cmd 'add', got %q", cmd)
	}
	if args != "Buy groceries" {
		t.Errorf("expected args 'Buy groceries', got %q", args)
	}
}

func TestSearchView_ParseCommand_MoodNoArgs(t *testing.T) {
	cmd, args := parseCommand(":mood")
	if cmd != "mood" {
		t.Errorf("expected cmd 'mood', got %q", cmd)
	}
	if args != "" {
		t.Errorf("expected empty args, got %q", args)
	}
}

func TestSearchView_ParseCommand_MoodWithArg(t *testing.T) {
	cmd, args := parseCommand(":mood Focused")
	if cmd != "mood" {
		t.Errorf("expected cmd 'mood', got %q", cmd)
	}
	if args != "Focused" {
		t.Errorf("expected args 'Focused', got %q", args)
	}
}

func TestSearchView_ParseCommand_Stats(t *testing.T) {
	cmd, _ := parseCommand(":stats")
	if cmd != "stats" {
		t.Errorf("expected cmd 'stats', got %q", cmd)
	}
}

func TestSearchView_ParseCommand_Help(t *testing.T) {
	cmd, _ := parseCommand(":help")
	if cmd != "help" {
		t.Errorf("expected cmd 'help', got %q", cmd)
	}
}

func TestSearchView_ParseCommand_Quit(t *testing.T) {
	cmd, _ := parseCommand(":quit")
	if cmd != "quit" {
		t.Errorf("expected cmd 'quit', got %q", cmd)
	}
}

func TestSearchView_ParseCommand_Exit(t *testing.T) {
	cmd, _ := parseCommand(":exit")
	if cmd != "exit" {
		t.Errorf("expected cmd 'exit', got %q", cmd)
	}
}

func TestSearchView_ParseCommand_CaseInsensitive(t *testing.T) {
	cmd, _ := parseCommand(":HELP")
	if cmd != "help" {
		t.Errorf("expected cmd 'help' (lowered), got %q", cmd)
	}
}

func TestSearchView_ParseCommand_EmptyCommand(t *testing.T) {
	cmd, _ := parseCommand(":")
	if cmd != "" {
		t.Errorf("expected empty cmd for ':', got %q", cmd)
	}
}

func TestSearchView_ParseCommand_UnknownCommand(t *testing.T) {
	cmd, _ := parseCommand(":foo")
	if cmd != "foo" {
		t.Errorf("expected cmd 'foo', got %q", cmd)
	}
}

// --- Navigation Tests ---

func TestSearchView_NavigationDown_MovesSelection(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	sv.textInput.SetValue("t") // Matches all three
	sv.results = sv.filterTasks("t")
	sv.selectedIndex = -1

	sv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sv.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0 after down, got %d", sv.selectedIndex)
	}
}

func TestSearchView_NavigationUp_MovesSelection(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	sv.textInput.SetValue("t")
	sv.results = sv.filterTasks("t")
	sv.selectedIndex = 1

	sv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sv.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0 after up, got %d", sv.selectedIndex)
	}
}

func TestSearchView_NavigationJK_ViStyle(t *testing.T) {
	sv := newTestSearchView("Task A", "Task B", "Task C")
	// j/k navigation only works when textInput is empty (to avoid conflicting with typing)
	sv.textInput.SetValue("")
	sv.results = []*tasks.Task{
		tasks.NewTask("Task A"),
		tasks.NewTask("Task B"),
		tasks.NewTask("Task C"),
	}
	sv.selectedIndex = 0

	// j moves down
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if sv.selectedIndex != 1 {
		t.Errorf("expected selectedIndex 1 after j, got %d", sv.selectedIndex)
	}

	// k moves up
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if sv.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0 after k, got %d", sv.selectedIndex)
	}
}

func TestSearchView_NavigationBoundsCheck_UpperBound(t *testing.T) {
	sv := newTestSearchView("Task A", "Task B")
	sv.textInput.SetValue("Task")
	sv.results = sv.filterTasks("Task")
	sv.selectedIndex = 1 // At last item

	sv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sv.selectedIndex > 1 {
		t.Errorf("selectedIndex should not exceed results length, got %d", sv.selectedIndex)
	}
}

func TestSearchView_NavigationBoundsCheck_LowerBound(t *testing.T) {
	sv := newTestSearchView("Task A", "Task B")
	sv.textInput.SetValue("Task")
	sv.results = sv.filterTasks("Task")
	sv.selectedIndex = 0

	sv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sv.selectedIndex < 0 {
		t.Errorf("selectedIndex should not go below 0, got %d", sv.selectedIndex)
	}
}

// --- Esc Key Tests ---

func TestSearchView_Esc_SendsSearchClosedMsg(t *testing.T) {
	sv := newTestSearchView("task1", "task2", "task3")
	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("Esc should return a command")
	}
	msg := cmd()
	if _, ok := msg.(SearchClosedMsg); !ok {
		t.Errorf("expected SearchClosedMsg, got %T", msg)
	}
}

// --- Enter Key on Selected Result ---

func TestSearchView_Enter_WithSelection_SendsSearchResultSelectedMsg(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	sv.textInput.SetValue("unit")
	sv.results = sv.filterTasks("unit")
	sv.selectedIndex = 0

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter on selected result should return a command")
	}
	msg := cmd()
	srm, ok := msg.(SearchResultSelectedMsg)
	if !ok {
		t.Errorf("expected SearchResultSelectedMsg, got %T", msg)
	}
	if srm.Task.Text != "Write unit tests" {
		t.Errorf("expected task 'Write unit tests', got %q", srm.Task.Text)
	}
}

func TestSearchView_Enter_NoSelection_Noop(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug")
	sv.textInput.SetValue("unit")
	sv.results = sv.filterTasks("unit")
	sv.selectedIndex = -1

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// No selection, no command (or it could be nil)
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(SearchResultSelectedMsg); ok {
			t.Error("should NOT send SearchResultSelectedMsg with no selection")
		}
	}
}

// --- Command Mode Tests ---

func TestSearchView_CommandMode_DetectedByColon(t *testing.T) {
	sv := newTestSearchView("task1", "task2")
	sv.textInput.SetValue(":")
	sv.checkCommandMode()
	if !sv.isCommandMode {
		t.Error("isCommandMode should be true when input starts with ':'")
	}
}

func TestSearchView_CommandMode_NotDetectedWithoutColon(t *testing.T) {
	sv := newTestSearchView("task1", "task2")
	sv.textInput.SetValue("search text")
	sv.checkCommandMode()
	if sv.isCommandMode {
		t.Error("isCommandMode should be false when input doesn't start with ':'")
	}
}

// --- :add Command Tests ---

func TestSearchView_AddCommand_CreatesTask(t *testing.T) {
	sv := newTestSearchView("existing task")
	sv.textInput.SetValue(":add New task from search")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":add should return a command")
	}
	msg := cmd()
	tam, ok := msg.(TaskAddedMsg)
	if !ok {
		t.Errorf("expected TaskAddedMsg, got %T", msg)
	}
	if tam.Task.Text != "New task from search" {
		t.Errorf("expected task text 'New task from search', got %q", tam.Task.Text)
	}
}

func TestSearchView_AddCommand_NoText_ShowsUsageError(t *testing.T) {
	sv := newTestSearchView("existing task")
	sv.textInput.SetValue(":add")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":add with no text should return a command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "Usage") {
		t.Errorf("expected usage message, got %q", fm.Text)
	}
}

// --- :quit / :exit Commands ---

func TestSearchView_QuitCommand_SendsQuit(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":quit")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":quit should return a command")
	}
	// tea.Quit returns a special batch message
}

func TestSearchView_ExitCommand_SendsQuit(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":exit")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":exit should return a command")
	}
}

// --- :help Command ---

func TestSearchView_HelpCommand_ShowsHelp(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":help")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":help should return a command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "help") && !strings.Contains(fm.Text, "Help") && !strings.Contains(fm.Text, "Commands") {
		t.Errorf("expected help text, got %q", fm.Text)
	}
}

// --- :stats Command ---

func TestSearchView_StatsCommand_ShowsStats(t *testing.T) {
	sv, _ := newTestSearchViewWithTracker("task1")
	sv.textInput.SetValue(":stats")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":stats should return a command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "Session") && !strings.Contains(fm.Text, "Stats") && !strings.Contains(fm.Text, "stats") {
		t.Errorf("expected stats text, got %q", fm.Text)
	}
}

// --- Unknown Command ---

func TestSearchView_UnknownCommand_ShowsError(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":foo")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal("unknown command should return a command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "Unknown command") {
		t.Errorf("expected 'Unknown command', got %q", fm.Text)
	}
	if !strings.Contains(fm.Text, "foo") {
		t.Errorf("expected command name 'foo' in error message, got %q", fm.Text)
	}
}

// --- RestoreState ---

func TestSearchView_RestoreState(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	sv.RestoreState("unit", 0)
	if sv.textInput.Value() != "unit" {
		t.Errorf("expected textInput value 'unit', got %q", sv.textInput.Value())
	}
	if sv.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0, got %d", sv.selectedIndex)
	}
	if len(sv.results) != 1 {
		t.Errorf("expected 1 result after restore, got %d", len(sv.results))
	}
}

// --- SetWidth ---

func TestSearchView_SetWidth(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.SetWidth(120)
	if sv.width != 120 {
		t.Errorf("expected width 120, got %d", sv.width)
	}
}

// --- View Rendering ---

func TestSearchView_View_ShowsSearchInput(t *testing.T) {
	sv := newTestSearchView("task1", "task2")
	sv.SetWidth(80)
	view := sv.View()
	if !strings.Contains(view, "Search") {
		t.Error("View should contain search-related text")
	}
}

func TestSearchView_View_ShowsNoMatchesMessage(t *testing.T) {
	sv := newTestSearchView("task1", "task2")
	sv.SetWidth(80)
	sv.textInput.SetValue("nonexistent")
	sv.results = sv.filterTasks("nonexistent")
	view := sv.View()
	if !strings.Contains(view, "No tasks match") {
		t.Error("View should show 'No tasks match' message when no results")
	}
}

func TestSearchView_View_ShowsResults(t *testing.T) {
	sv := newTestSearchView("Write unit tests", "Fix login bug", "Review PR")
	sv.SetWidth(80)
	sv.textInput.SetValue("unit")
	sv.results = sv.filterTasks("unit")
	view := sv.View()
	if !strings.Contains(view, "Write unit tests") {
		t.Error("View should show matching task text")
	}
}

func TestSearchView_View_CommandModeIndicator(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.SetWidth(80)
	sv.textInput.SetValue(":")
	sv.isCommandMode = true
	view := sv.View()
	// Command mode should have some visual indicator
	if !strings.Contains(view, "command") && !strings.Contains(view, "Command") && !strings.Contains(view, ":") {
		t.Error("View should indicate command mode")
	}
}

// --- Edge Cases ---

func TestSearchView_EmptyPool_NoResults(t *testing.T) {
	pool := tasks.NewTaskPool()
	sv := NewSearchView(pool, nil)
	sv.textInput.SetValue("anything")
	results := sv.filterTasks("anything")
	if len(results) != 0 {
		t.Errorf("expected 0 results from empty pool, got %d", len(results))
	}
}

func TestSearchView_TerminalResize_UpdatesTextInputWidth(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.SetWidth(120)
	// textInput width should have been updated
	if sv.textInput.Width == 0 {
		t.Error("textInput width should be updated on SetWidth")
	}
}
