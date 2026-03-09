package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Test Helpers ---

func newTestSearchView(texts ...string) *SearchView {
	pool := makePool(texts...)
	return NewSearchView(pool, nil, nil, nil, nil)
}

func newTestSearchViewWithTracker(texts ...string) (*SearchView, *core.SessionTracker) {
	pool := makePool(texts...)
	tracker := core.NewSessionTracker()
	return NewSearchView(pool, tracker, nil, nil, nil), tracker
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
	pool := core.NewTaskPool()
	t1 := core.NewTask("todo task")
	t2 := core.NewTask("blocked task")
	_ = t2.UpdateStatus(core.StatusBlocked)
	t3 := core.NewTask("in-progress task")
	_ = t3.UpdateStatus(core.StatusInProgress)
	pool.AddTask(t1)
	pool.AddTask(t2)
	pool.AddTask(t3)

	sv := NewSearchView(pool, nil, nil, nil, nil)
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
	sv.results = []*core.Task{
		core.NewTask("Task A"),
		core.NewTask("Task B"),
		core.NewTask("Task C"),
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

func TestSearchView_AddCommand_NoText_EmitsAddTaskPromptMsg(t *testing.T) {
	sv := newTestSearchView("existing task")
	sv.textInput.SetValue(":add")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":add with no text should return a command")
	}
	msg := cmd()
	_, ok := msg.(AddTaskPromptMsg)
	if !ok {
		t.Errorf("expected AddTaskPromptMsg, got %T", msg)
	}
}

// --- :add-ctx Command Tests ---

func TestSearchView_AddCtxCommand_NoArgs_EmitsPromptMsg(t *testing.T) {
	sv := newTestSearchView("existing task")
	sv.textInput.SetValue(":add-ctx")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":add-ctx should return a command")
	}
	msg := cmd()
	ctxMsg, ok := msg.(AddTaskWithContextPromptMsg)
	if !ok {
		t.Errorf("expected AddTaskWithContextPromptMsg, got %T", msg)
	}
	if ctxMsg.PrefilledText != "" {
		t.Errorf("expected empty prefilled text, got %q", ctxMsg.PrefilledText)
	}
}

func TestSearchView_AddCtxCommand_WithArgs_EmitsPromptMsgWithPrefill(t *testing.T) {
	sv := newTestSearchView("existing task")
	sv.textInput.SetValue(":add-ctx Buy groceries")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":add-ctx with args should return a command")
	}
	msg := cmd()
	ctxMsg, ok := msg.(AddTaskWithContextPromptMsg)
	if !ok {
		t.Errorf("expected AddTaskWithContextPromptMsg, got %T", msg)
	}
	if ctxMsg.PrefilledText != "Buy groceries" {
		t.Errorf("expected prefilled text 'Buy groceries', got %q", ctxMsg.PrefilledText)
	}
}

func TestSearchView_AddWhyFlag_EmitsPromptMsg(t *testing.T) {
	sv := newTestSearchView("existing task")
	sv.textInput.SetValue(":add --why")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":add --why should return a command")
	}
	msg := cmd()
	_, ok := msg.(AddTaskWithContextPromptMsg)
	if !ok {
		t.Errorf("expected AddTaskWithContextPromptMsg, got %T", msg)
	}
}

func TestSearchView_AddWhyFlag_WithText_EmitsPromptMsgWithPrefill(t *testing.T) {
	sv := newTestSearchView("existing task")
	sv.textInput.SetValue(":add --why Buy groceries")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":add --why with text should return a command")
	}
	msg := cmd()
	ctxMsg, ok := msg.(AddTaskWithContextPromptMsg)
	if !ok {
		t.Errorf("expected AddTaskWithContextPromptMsg, got %T", msg)
	}
	if ctxMsg.PrefilledText != "Buy groceries" {
		t.Errorf("expected prefilled text 'Buy groceries', got %q", ctxMsg.PrefilledText)
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
	// :quit sends RequestQuitMsg which triggers quit flow
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
	if _, ok := msg.(ShowHelpMsg); !ok {
		t.Errorf("expected ShowHelpMsg, got %T", msg)
	}
}

// --- :theme Command ---

func TestSearchView_ThemeCommand_EmitsShowThemePickerMsg(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":theme")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":theme should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ShowThemePickerMsg); !ok {
		t.Errorf("expected ShowThemePickerMsg, got %T", msg)
	}
}

func TestSearchView_HelpCommand_ContainsTheme(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":help")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":help should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ShowHelpMsg); !ok {
		t.Fatalf("expected ShowHelpMsg, got %T", msg)
	}
	// Verify :theme is listed in the help content data
	found := false
	for _, section := range helpContent {
		for _, entry := range section.Entries {
			if entry.Key == ":theme" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected helpContent to contain ':theme' entry")
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

// --- Stable Ordering Tests ---

func TestSearchView_FilterTasks_StableOrdering(t *testing.T) {
	// Create tasks whose map iteration order will be randomized.
	// With enough tasks, Go's map randomization makes unstable ordering
	// detectable across repeated calls.
	texts := []string{
		"Zebra task",
		"Alpha task",
		"Mango task",
		"Beta task",
		"Omega task",
		"Delta task",
		"Gamma task",
	}
	sv := newTestSearchView(texts...)

	// Call filterTasks multiple times — results must be identical each time
	first := sv.filterTasks("task")
	if len(first) != len(texts) {
		t.Fatalf("expected %d matches, got %d", len(texts), len(first))
	}

	for i := 0; i < 20; i++ {
		got := sv.filterTasks("task")
		if len(got) != len(first) {
			t.Fatalf("iteration %d: length changed from %d to %d", i, len(first), len(got))
		}
		for j := range first {
			if got[j].Text != first[j].Text {
				t.Errorf("iteration %d, index %d: got %q, want %q — order is not stable",
					i, j, got[j].Text, first[j].Text)
			}
		}
	}

	// Verify alphabetical sort order
	for i := 1; i < len(first); i++ {
		if first[i-1].Text > first[i].Text {
			t.Errorf("results not sorted: %q should come before %q",
				first[i].Text, first[i-1].Text)
		}
	}
}

func TestSearchView_FilterTasks_StableOrdering_SubsetPreserved(t *testing.T) {
	// When adding/removing a character, unaffected results should stay
	// in the same relative order.
	texts := []string{
		"Buy apples",
		"Buy bananas",
		"Sell oranges",
		"Buy cherries",
	}
	sv := newTestSearchView(texts...)

	broad := sv.filterTasks("Buy")
	if len(broad) != 3 {
		t.Fatalf("expected 3 matches for 'Buy', got %d", len(broad))
	}

	// Narrow the query — the subset that still matches must keep the same order
	narrow := sv.filterTasks("Buy b")
	if len(narrow) != 1 {
		t.Fatalf("expected 1 match for 'Buy b', got %d", len(narrow))
	}
	if narrow[0].Text != "Buy bananas" {
		t.Errorf("expected 'Buy bananas', got %q", narrow[0].Text)
	}

	// Verify broad results are alphabetically sorted
	for i := 1; i < len(broad); i++ {
		if broad[i-1].Text > broad[i].Text {
			t.Errorf("broad results not sorted: %q before %q", broad[i-1].Text, broad[i].Text)
		}
	}
}

// --- Edge Cases ---

func TestSearchView_EmptyPool_NoResults(t *testing.T) {
	pool := core.NewTaskPool()
	sv := NewSearchView(pool, nil, nil, nil, nil)
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

// --- SetHeight and Bottom-Anchored Layout Tests ---

func TestSearchView_SetHeight(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.SetHeight(40)
	if sv.height != 40 {
		t.Errorf("expected height 40, got %d", sv.height)
	}
}

func TestSearchView_CommandMode_FixedHeightSuggestions(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Full suggestions (all commands shown)
	sv.textInput.SetValue(":")
	sv.isCommandMode = true
	sv.commandSuggestions = filterCommands("")
	viewAll := sv.View()
	allLines := strings.Count(viewAll, "\n")

	// Filtered suggestions (fewer commands shown)
	sv.textInput.SetValue(":he")
	sv.commandSuggestions = filterCommands("he")
	viewFiltered := sv.View()
	filteredLines := strings.Count(viewFiltered, "\n")

	// Both views should have the same total line count — padding compensates
	if allLines != filteredLines {
		t.Errorf("expected same line count for all vs filtered suggestions, got %d vs %d", allLines, filteredLines)
	}
}

func TestSearchView_CommandMode_NoSuggestions_SameHeight(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.SetWidth(80)
	sv.SetHeight(40)

	// All suggestions
	sv.textInput.SetValue(":")
	sv.isCommandMode = true
	sv.commandSuggestions = filterCommands("")
	viewAll := sv.View()
	allLines := strings.Count(viewAll, "\n")

	// No matching suggestions
	sv.textInput.SetValue(":zzzzz")
	sv.commandSuggestions = filterCommands("zzzzz")
	viewNone := sv.View()
	noneLines := strings.Count(viewNone, "\n")

	if allLines != noneLines {
		t.Errorf("expected same line count for all vs no suggestions, got %d vs %d", allLines, noneLines)
	}
}

func TestSearchView_BottomAnchor_PaddingPresent(t *testing.T) {
	sv := newTestSearchView()
	sv.SetWidth(80)
	sv.SetHeight(50)

	// With no content, the view should have padding to fill 50 lines
	view := sv.View()
	lines := strings.Count(view, "\n")
	// Header (2) + footer (3) + padding should approach height
	if lines < 40 {
		t.Errorf("expected padding to push content toward bottom, got only %d lines for height 50", lines)
	}
}

func TestSearchView_NoHeight_NoPadding(t *testing.T) {
	sv := newTestSearchView("task1")
	sv.SetWidth(80)
	// height = 0 (default) — no bottom-anchoring
	view := sv.View()
	lines := strings.Count(view, "\n")
	// Without height set, view should be compact
	if lines > 20 {
		t.Errorf("expected compact view without height set, got %d lines", lines)
	}
}
