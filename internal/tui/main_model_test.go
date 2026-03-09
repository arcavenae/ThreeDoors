package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/adapters/textfile"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	tea "github.com/charmbracelet/bubbletea"
)

// --- helpers ---

// testProvider is a no-op TaskProvider for TUI testing.
type testProvider struct {
	completedIDs []string
	completeErr  error
	savedTasks   []*core.Task
	saveErr      error
}

func (p *testProvider) LoadTasks() ([]*core.Task, error) { return nil, nil }
func (p *testProvider) SaveTask(t *core.Task) error      { return p.saveErr }

func (p *testProvider) SaveTasks(ts []*core.Task) error {
	p.savedTasks = append(p.savedTasks, ts...)
	return p.saveErr
}

func (p *testProvider) DeleteTask(_ string) error { return nil }

func (p *testProvider) MarkComplete(id string) error {
	if p.completeErr != nil {
		return p.completeErr
	}
	p.completedIDs = append(p.completedIDs, id)
	return nil
}

func (p *testProvider) Name() string                        { return "test" }
func (p *testProvider) Watch() <-chan core.ChangeEvent      { return nil }
func (p *testProvider) HealthCheck() core.HealthCheckResult { return core.HealthCheckResult{} }

func makePool(texts ...string) *core.TaskPool {
	pool := core.NewTaskPool()
	for _, t := range texts {
		pool.AddTask(core.NewTask(t))
	}
	return pool
}

func makeModel(texts ...string) *MainModel {
	return makeModelWithProvider(&testProvider{}, texts...)
}

func makeModelWithProvider(provider core.TaskProvider, texts ...string) *MainModel {
	pool := makePool(texts...)
	tracker := core.NewSessionTracker()
	return NewMainModel(pool, tracker, provider, nil, false, nil)
}

func keyMsg(s string) tea.Msg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// --- MainModel Creation ---

func TestNewMainModel_DefaultsToDoorsView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors, got %d", m.viewMode)
	}
}

func TestNewMainModel_DoorsViewCreated(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	if m.doorsView == nil {
		t.Fatal("doorsView should not be nil")
	}
}

func TestNewMainModel_DetailViewNil(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	if m.detailView != nil {
		t.Error("detailView should be nil initially")
	}
}

func TestInit_ReturnsDeferReturnTickCmd(t *testing.T) {
	m := makeModel("task1")
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a defer-return tick command")
	}
}

// --- Door Selection Keys ---

func TestDoorsView_SelectLeftDoor(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("a"))
	if m.doorsView.selectedDoorIndex != 0 {
		t.Errorf("expected selectedDoorIndex 0, got %d", m.doorsView.selectedDoorIndex)
	}
}

func TestDoorsView_SelectCenterDoor(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("w"))
	if m.doorsView.selectedDoorIndex != 1 {
		t.Errorf("expected selectedDoorIndex 1, got %d", m.doorsView.selectedDoorIndex)
	}
}

func TestDoorsView_SelectRightDoor(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("d"))
	if m.doorsView.selectedDoorIndex != 2 {
		t.Errorf("expected selectedDoorIndex 2, got %d", m.doorsView.selectedDoorIndex)
	}
}

func TestDoorsView_ArrowLeftSelectsDoor(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("left"))
	if m.doorsView.selectedDoorIndex != 0 {
		t.Errorf("expected selectedDoorIndex 0, got %d", m.doorsView.selectedDoorIndex)
	}
}

func TestDoorsView_ArrowUpSelectsDoor(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("up"))
	if m.doorsView.selectedDoorIndex != 1 {
		t.Errorf("expected selectedDoorIndex 1, got %d", m.doorsView.selectedDoorIndex)
	}
}

func TestDoorsView_ArrowRightSelectsDoor(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("right"))
	if m.doorsView.selectedDoorIndex != 2 {
		t.Errorf("expected selectedDoorIndex 2, got %d", m.doorsView.selectedDoorIndex)
	}
}

// --- Deselect Toggle (Story 36.2) ---

func TestDoorsView_DeselectToggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		selectKey   string
		doorIndex   int
		deselectKey string
		wantIndex   int
	}{
		// Same key deselects
		{"select door 0 with a then deselect with a", "a", 0, "a", -1},
		{"select door 0 with left then deselect with left", "left", 0, "left", -1},
		{"select door 1 with w then deselect with w", "w", 1, "w", -1},
		{"select door 1 with up then deselect with up", "up", 1, "up", -1},
		{"select door 2 with d then deselect with d", "d", 2, "d", -1},
		{"select door 2 with right then deselect with right", "right", 2, "right", -1},
		// Different key switches (does NOT deselect)
		{"select door 0 then switch to door 1", "a", 0, "w", 1},
		{"select door 0 then switch to door 2", "a", 0, "d", 2},
		{"select door 1 then switch to door 0", "w", 1, "a", 0},
		{"select door 1 then switch to door 2", "w", 1, "d", 2},
		{"select door 2 then switch to door 0", "d", 2, "a", 0},
		{"select door 2 then switch to door 1", "d", 2, "w", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := makeModel("task1", "task2", "task3")

			// First key: select the door
			m.Update(keyMsg(tt.selectKey))
			if m.doorsView.selectedDoorIndex != tt.doorIndex {
				t.Fatalf("after select: expected index %d, got %d", tt.doorIndex, m.doorsView.selectedDoorIndex)
			}

			// Second key: toggle or switch
			m.Update(keyMsg(tt.deselectKey))
			if m.doorsView.selectedDoorIndex != tt.wantIndex {
				t.Errorf("after second key: expected index %d, got %d", tt.wantIndex, m.doorsView.selectedDoorIndex)
			}
		})
	}
}

func TestDoorsView_DeselectThenReselect(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")

	// Select door 0
	m.Update(keyMsg("a"))
	if m.doorsView.selectedDoorIndex != 0 {
		t.Fatalf("expected 0, got %d", m.doorsView.selectedDoorIndex)
	}

	// Deselect door 0
	m.Update(keyMsg("a"))
	if m.doorsView.selectedDoorIndex != -1 {
		t.Fatalf("expected -1 after deselect, got %d", m.doorsView.selectedDoorIndex)
	}

	// Re-select door 0
	m.Update(keyMsg("a"))
	if m.doorsView.selectedDoorIndex != 0 {
		t.Errorf("expected 0 after re-select, got %d", m.doorsView.selectedDoorIndex)
	}
}

// --- Enter Key ---

func TestEnterKey_WithSelection_OpensDetail(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	// Select center door first
	m.Update(keyMsg("w"))
	// Press Enter
	m.Update(keyMsg("enter"))

	if m.viewMode != ViewDetail {
		t.Errorf("expected ViewDetail, got %d", m.viewMode)
	}
	if m.detailView == nil {
		t.Fatal("detailView should not be nil after Enter")
	}
}

func TestEnterKey_NoSelection_Noop(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	// No door selected (selectedDoorIndex == -1)
	m.Update(keyMsg("enter"))

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors, got %d (Enter with no selection should be no-op)", m.viewMode)
	}
	if m.detailView != nil {
		t.Error("detailView should remain nil when no door selected")
	}
}

// --- Spacebar as Enter Alias (Story 39.6) ---

func TestSpacebar_OpensSelectedDoor(t *testing.T) {
	tests := []struct {
		name      string
		selectKey string
		doorIndex int
	}{
		{"left door via a", "a", 0},
		{"center door via w", "w", 1},
		{"right door via d", "d", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModel("task1", "task2", "task3")
			m.Update(keyMsg(tt.selectKey))
			if m.doorsView.selectedDoorIndex != tt.doorIndex {
				t.Fatalf("expected selectedDoorIndex %d, got %d", tt.doorIndex, m.doorsView.selectedDoorIndex)
			}
			m.Update(keyMsg(" "))
			if m.viewMode != ViewDetail {
				t.Errorf("expected ViewDetail after spacebar, got %d", m.viewMode)
			}
			if m.detailView == nil {
				t.Fatal("detailView should not be nil after spacebar")
			}
		})
	}
}

func TestSpacebar_NoSelection_Noop(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg(" "))

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors, got %d (spacebar with no selection should be no-op)", m.viewMode)
	}
	if m.detailView != nil {
		t.Error("detailView should remain nil when no door selected")
	}
}

// --- Esc from Detail Returns to Doors ---

func TestEscFromDetail_ReturnsToDoors(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	// Enter detail view
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))
	if m.viewMode != ViewDetail {
		t.Fatal("should be in ViewDetail")
	}

	// Press Esc — the detail view returns a ReturnToDoorsMsg
	_, cmd := m.Update(keyMsg("esc"))
	if cmd != nil {
		// Execute the command to get the message
		msg := cmd()
		m.Update(msg)
	}

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after Esc, got %d", m.viewMode)
	}
	if m.detailView != nil {
		t.Error("detailView should be nil after returning to doors")
	}
}

// --- M Key Opens Mood ---

func TestMKey_OpensMoodFromDoors(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	_, cmd := m.Update(keyMsg("m"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}
	if m.viewMode != ViewMood {
		t.Errorf("expected ViewMood, got %d", m.viewMode)
	}
	if m.moodView == nil {
		t.Error("moodView should not be nil")
	}
}

// --- Quit Keys ---

func TestQKey_QuitsFromDoors(t *testing.T) {
	m := makeModel("task1")
	_, cmd := m.Update(keyMsg("q"))
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestCtrlC_QuitsFromDoors(t *testing.T) {
	m := makeModel("task1")
	_, cmd := m.Update(keyMsg("ctrl+c"))
	if cmd == nil {
		t.Fatal("expected quit command from Ctrl+C")
	}
}

// --- Refresh / Reroll ---

func TestDownKey_RerollsDoors(t *testing.T) {
	m := makeModel("task1", "task2", "task3", "task4", "task5")
	// Select a door first
	m.Update(keyMsg("w"))
	if m.doorsView.selectedDoorIndex != 1 {
		t.Fatal("door should be selected")
	}
	// Reroll
	m.Update(keyMsg("s"))
	if m.doorsView.selectedDoorIndex != -1 {
		t.Errorf("selectedDoorIndex should be -1 after reroll, got %d", m.doorsView.selectedDoorIndex)
	}
}

// --- View Rendering ---

func TestDoorsView_RendersHeader(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	view := m.View()
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("View should contain 'ThreeDoors' header")
	}
}

func TestDoorsView_RendersTaskTexts(t *testing.T) {
	m := makeModel("Alpha Task", "Beta Task", "Gamma Task")
	view := m.View()
	// At least some task texts should appear (3 doors from 3 tasks)
	found := 0
	for _, text := range []string{"Alpha Task", "Beta Task", "Gamma Task"} {
		if strings.Contains(view, text) {
			found++
		}
	}
	if found == 0 {
		t.Error("View should contain at least one task text")
	}
}

func TestDoorsView_RendersHelpText(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	view := m.View()
	if !strings.Contains(view, "quit") {
		t.Error("View should contain help text with quit instruction")
	}
}

func TestDoorsView_AllTasksDone_ShowsMessage(t *testing.T) {
	pool := core.NewTaskPool()
	tracker := core.NewSessionTracker()
	m := NewMainModel(pool, tracker, textfile.NewTextFileProvider(), nil, false, nil)
	view := m.View()
	if !strings.Contains(view, "All tasks done") {
		t.Errorf("View should show 'All tasks done' when pool is empty, got: %s", view)
	}
}

// --- Completion Counter ---

func TestDoorsView_CompletionCounter_ShowsAfterCompletion(t *testing.T) {
	m := makeModel("task1", "task2", "task3", "task4", "task5")

	// Enter detail view
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))

	// Complete the task — now goes to ViewNextSteps
	_, cmd := m.Update(keyMsg("c"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	if m.viewMode != ViewNextSteps {
		t.Errorf("expected ViewNextSteps after completion, got %d", m.viewMode)
	}

	// Dismiss next steps to return to doors
	_, cmd = m.Update(keyMsg("esc"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	view := m.View()
	if !strings.Contains(view, "Completed this session: 1") {
		t.Errorf("View should contain 'Completed this session: 1', got: %s", view)
	}
}

// --- Flash Messages ---

func TestFlashMessage_ShowsAfterCompletion(t *testing.T) {
	m := makeModel("task1", "task2", "task3", "task4", "task5")

	// Enter detail view and complete
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))
	_, cmd := m.Update(keyMsg("c"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	// Flash should be one of the celebration messages from the pool
	if m.flash == "" {
		t.Fatal("flash should be set after task completion")
	}

	found := false
	for _, celebration := range celebrationMessages {
		if strings.HasPrefix(m.flash, celebration) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("flash message %q does not start with any celebrationMessages", m.flash)
	}
}

func TestFlashMessage_ClearedByClearFlashMsg(t *testing.T) {
	m := makeModel("task1", "task2", "task3", "task4", "task5")

	// Set flash via FlashMsg
	m.Update(FlashMsg{Text: "Test flash message"})
	view := m.View()
	if !strings.Contains(view, "Test flash message") {
		t.Fatal("Flash should be visible before clearing")
	}

	// Clear flash
	m.Update(ClearFlashMsg{})

	view = m.View()
	if strings.Contains(view, "Test flash message") {
		t.Error("Flash message should be cleared after ClearFlashMsg")
	}
}

// --- View Mode Routing ---

func TestViewRouting_DoorsView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	view := m.View()
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("DoorsView should be rendered when viewMode is ViewDoors")
	}
}

func TestViewRouting_DetailView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))
	view := m.View()
	if !strings.Contains(view, "TASK DETAILS") {
		t.Error("DetailView should be rendered when viewMode is ViewDetail")
	}
}

func TestViewRouting_MoodView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	_, cmd := m.Update(keyMsg("m"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}
	view := m.View()
	if !strings.Contains(view, "How are you feeling") {
		t.Error("MoodView should be rendered when viewMode is ViewMood")
	}
}

// --- Window Resize ---

func TestWindowSizeMsg_UpdatesDimensions(t *testing.T) {
	m := makeModel("task1")
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.width != 120 || m.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", m.width, m.height)
	}
}

// --- Status Change Auto-Refreshes ---

func TestStatusChange_ReturnsToDoorsAndRefreshes(t *testing.T) {
	m := makeModel("task1", "task2", "task3", "task4", "task5")

	// Enter detail view
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))
	if m.viewMode != ViewDetail {
		t.Fatal("should be in ViewDetail")
	}

	// Change status to in-progress
	_, cmd := m.Update(keyMsg("i"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after status change, got %d", m.viewMode)
	}
}

// --- Mood Captured Returns to Doors ---

func TestMoodCaptured_ReturnsToDoors(t *testing.T) {
	m := makeModel("task1", "task2", "task3")

	// Open mood
	_, cmd := m.Update(keyMsg("m"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}
	if m.viewMode != ViewMood {
		t.Fatal("should be in ViewMood")
	}

	// Select a mood (press "1" for Focused)
	_, cmd = m.Update(keyMsg("1"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after mood capture, got %d", m.viewMode)
	}
}

func TestMoodCaptured_FlashMessage(t *testing.T) {
	m := makeModel("task1", "task2", "task3")

	// Open mood
	_, cmd := m.Update(keyMsg("m"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	// Select Focused
	_, cmd = m.Update(keyMsg("1"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	view := m.View()
	if !strings.Contains(view, "Mood logged: Focused") {
		t.Errorf("expected mood flash message, got: %s", view)
	}
}

// --- Session Tracker Integration ---

func TestSessionTracker_DoorSelectionRecorded(t *testing.T) {
	m := makeModel("task1", "task2", "task3")

	// Select and open a door
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))

	metrics := m.tracker.Finalize()
	if len(metrics.DoorSelections) == 0 {
		t.Error("expected door selection to be recorded in tracker")
	}
}

func TestSessionTracker_RefreshRecorded(t *testing.T) {
	m := makeModel("task1", "task2", "task3", "task4", "task5")
	m.Update(keyMsg("s"))

	metrics := m.tracker.Finalize()
	if metrics.RefreshesUsed != 1 {
		t.Errorf("expected 1 refresh, got %d", metrics.RefreshesUsed)
	}
}

// --- MarkComplete Provider Integration ---

func TestTaskCompleted_CallsProviderMarkComplete(t *testing.T) {
	provider := &testProvider{}
	m := makeModelWithProvider(provider, "task1", "task2", "task3", "task4", "task5")

	// Enter detail and complete
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))
	if m.detailView == nil {
		t.Fatal("should be in detail view")
	}
	taskID := m.detailView.task.ID

	_, cmd := m.Update(keyMsg("c"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	if len(provider.completedIDs) != 1 {
		t.Fatalf("expected 1 MarkComplete call, got %d", len(provider.completedIDs))
	}
	if provider.completedIDs[0] != taskID {
		t.Errorf("MarkComplete called with %q, want %q", provider.completedIDs[0], taskID)
	}
}

func TestTaskCompleted_PoolRemovedOnSuccess(t *testing.T) {
	provider := &testProvider{}
	m := makeModelWithProvider(provider, "task1", "task2", "task3", "task4", "task5")

	// Enter detail and complete
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))
	taskID := m.detailView.task.ID
	initialCount := len(m.pool.GetAllTasks())

	_, cmd := m.Update(keyMsg("c"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	finalCount := len(m.pool.GetAllTasks())
	if finalCount != initialCount-1 {
		t.Errorf("pool should have %d tasks, got %d", initialCount-1, finalCount)
	}
	for _, task := range m.pool.GetAllTasks() {
		if task.ID == taskID {
			t.Error("completed task should not be in pool")
		}
	}
}

func TestTaskCompleted_PoolRemainsOnProviderFailure(t *testing.T) {
	provider := &testProvider{completeErr: fmt.Errorf("disk full")}
	m := makeModelWithProvider(provider, "task1", "task2", "task3", "task4", "task5")

	initialCount := len(m.pool.GetAllTasks())

	// Enter detail and attempt complete
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))
	_, cmd := m.Update(keyMsg("c"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	finalCount := len(m.pool.GetAllTasks())
	if finalCount != initialCount {
		t.Errorf("pool should still have %d tasks on failure, got %d", initialCount, finalCount)
	}
	if m.flash != "Error completing task" {
		t.Errorf("flash should show error, got %q", m.flash)
	}
}

// --- Story 3.1: Quick Add Mode Tests ---

// T7: AddTaskPromptMsg transitions to ViewAddTask
func TestAddTaskPromptMsg_TransitionsToAddTaskView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(AddTaskPromptMsg{})
	if m.viewMode != ViewAddTask {
		t.Errorf("expected ViewAddTask, got %d", m.viewMode)
	}
}

// T8: TaskAddedMsg adds to pool, sets flash from message pool
func TestTaskAddedMsg_AddsToPool(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	initialCount := m.pool.Count()

	newTask := core.NewTask("brand new task")
	m.Update(TaskAddedMsg{Task: newTask})

	if m.pool.Count() != initialCount+1 {
		t.Errorf("expected pool count %d, got %d", initialCount+1, m.pool.Count())
	}
	// Flash should be one of the task-added messages from the pool
	found := false
	for _, msg := range taskAddedMessages {
		if m.flash == msg {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("flash message %q not found in taskAddedMessages pool", m.flash)
	}
}

// T9: Colon key from doors view opens search in command mode
func TestColonKey_OpensSearchInCommandMode(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg(":"))
	if m.viewMode != ViewSearch {
		t.Errorf("expected ViewSearch after ':', got %d", m.viewMode)
	}
	if m.searchView == nil {
		t.Fatal("searchView should be created")
	}
	val := m.searchView.textInput.Value()
	if val != ":" {
		t.Errorf("expected search input pre-populated with ':', got %q", val)
	}
}

// T10: TaskAddedMsg with empty pool - pool gets the task
func TestTaskAddedMsg_EmptyPool_AddsTask(t *testing.T) {
	pool := core.NewTaskPool()
	tracker := core.NewSessionTracker()
	m := NewMainModel(pool, tracker, textfile.NewTextFileProvider(), nil, false, nil)

	if m.pool.Count() != 0 {
		t.Fatal("pool should start empty")
	}

	newTask := core.NewTask("first task ever")
	m.Update(TaskAddedMsg{Task: newTask})

	if m.pool.Count() != 1 {
		t.Errorf("expected pool count 1, got %d", m.pool.Count())
	}
}

// TaskAddedMsg from search context returns to search
func TestTaskAddedMsg_FromSearch_ReturnsToSearch(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	// Simulate coming from search
	m.previousView = ViewSearch
	m.viewMode = ViewAddTask

	newTask := core.NewTask("new task from search")
	m.Update(TaskAddedMsg{Task: newTask})

	if m.viewMode != ViewSearch {
		t.Errorf("expected ViewSearch after TaskAddedMsg from search context, got %d", m.viewMode)
	}
}

// TaskAddedMsg from doors context shows next-steps view
func TestTaskAddedMsg_FromDoors_ShowsNextSteps(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.previousView = ViewDoors
	m.viewMode = ViewAddTask

	newTask := core.NewTask("new task from doors")
	m.Update(TaskAddedMsg{Task: newTask})

	if m.viewMode != ViewNextSteps {
		t.Errorf("expected ViewNextSteps after TaskAddedMsg from doors context, got %d", m.viewMode)
	}
	if m.nextStepsView == nil {
		t.Error("nextStepsView should not be nil")
	}
}

// AddTaskPromptMsg preserves previousView
func TestAddTaskPromptMsg_PreservesPreviousView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	// Start from doors view
	m.viewMode = ViewDoors
	m.Update(AddTaskPromptMsg{})
	if m.viewMode != ViewAddTask {
		t.Errorf("expected ViewAddTask, got %d", m.viewMode)
	}
	if m.previousView != ViewDoors {
		t.Errorf("expected previousView ViewDoors, got %d", m.previousView)
	}
}

// --- Story 0.34: Q-Key Behavior Tests (scoped quit to doors only) ---

// isQuitCmd checks if a tea.Cmd produces a RequestQuitMsg when executed.
func isQuitCmd(t *testing.T, cmd tea.Cmd) bool {
	t.Helper()
	if cmd == nil {
		return false
	}
	msg := cmd()
	_, ok := msg.(RequestQuitMsg)
	return ok
}

// isGoBackCmd checks if a tea.Cmd produces a go-back message when executed.
// Go-back messages include ReturnToDoorsMsg, NextStepDismissedMsg, and ThemeCancelledMsg.
func isGoBackCmd(t *testing.T, cmd tea.Cmd) bool {
	t.Helper()
	if cmd == nil {
		return false
	}
	msg := cmd()
	switch msg.(type) {
	case ReturnToDoorsMsg, NextStepDismissedMsg, ThemeCancelledMsg:
		return true
	}
	return false
}

func TestQKey_DoorsView_Quits(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	if m.viewMode != ViewDoors {
		t.Fatal("should be in ViewDoors")
	}

	_, cmd := m.Update(keyMsg("q"))
	if !isQuitCmd(t, cmd) {
		t.Error("'q' from doors view should produce RequestQuitMsg")
	}
}

func TestQKey_DetailView_GoesBack(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("w"))
	m.Update(keyMsg("enter"))
	if m.viewMode != ViewDetail {
		t.Fatal("should be in ViewDetail")
	}

	_, cmd := m.Update(keyMsg("q"))
	if isQuitCmd(t, cmd) {
		t.Error("'q' from detail view should NOT quit")
	}
	if !isGoBackCmd(t, cmd) {
		t.Error("'q' from detail view should go back")
	}
}

func TestQKey_InsightsView_GoesBack(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.insightsView = NewInsightsView(nil, nil)
	m.viewMode = ViewInsights

	_, cmd := m.Update(keyMsg("q"))
	if isQuitCmd(t, cmd) {
		t.Error("'q' from insights view should NOT quit (regression #330)")
	}
	if !isGoBackCmd(t, cmd) {
		t.Error("'q' from insights view should go back")
	}
}

func TestQKey_MoodView_CustomInput_NoAction(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	_, cmd := m.Update(keyMsg("m"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}
	if m.viewMode != ViewMood {
		t.Fatal("should be in ViewMood")
	}

	// Enter custom input mode by pressing "7"
	m.Update(keyMsg("7"))
	if !m.moodView.isCustom {
		t.Fatal("should be in custom input mode")
	}

	_, cmd = m.Update(keyMsg("q"))
	if isQuitCmd(t, cmd) || isGoBackCmd(t, cmd) {
		t.Error("'q' in mood custom input mode should be treated as text")
	}
}

// Table-driven test covering q-key behavior across all views (Story 0.34)
func TestQKey_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*MainModel)
		wantQuit   bool
		wantGoBack bool
	}{
		{
			name:     "doors view — q quits",
			setup:    func(_ *MainModel) {},
			wantQuit: true,
		},
		{
			name: "detail view (normal mode) — q goes back",
			setup: func(m *MainModel) {
				m.Update(keyMsg("w"))
				m.Update(keyMsg("enter"))
			},
			wantGoBack: true,
		},
		{
			name: "mood view (selection mode) — q goes back",
			setup: func(m *MainModel) {
				_, cmd := m.Update(keyMsg("m"))
				if cmd != nil {
					msg := cmd()
					m.Update(msg)
				}
			},
			wantGoBack: true,
		},
		{
			name: "health view — q goes back",
			setup: func(m *MainModel) {
				m.healthView = NewHealthView(core.HealthCheckResult{})
				m.viewMode = ViewHealth
			},
			wantGoBack: true,
		},
		{
			name: "insights view — q goes back",
			setup: func(m *MainModel) {
				m.insightsView = NewInsightsView(nil, nil)
				m.viewMode = ViewInsights
			},
			wantGoBack: true,
		},
		{
			name: "next steps view — q goes back",
			setup: func(m *MainModel) {
				m.nextStepsView = NewNextStepsView("completed", m.pool, nil)
				m.viewMode = ViewNextSteps
			},
			wantGoBack: true,
		},
		{
			name: "avoidance prompt — q goes back",
			setup: func(m *MainModel) {
				task := m.pool.GetAllTasks()[0]
				m.avoidancePromptView = NewAvoidancePromptView(task, 15)
				m.viewMode = ViewAvoidancePrompt
			},
			wantGoBack: true,
		},
		{
			name: "theme picker — q goes back",
			setup: func(m *MainModel) {
				m.themePickerView = NewThemePicker(themes.NewDefaultRegistry(), "")
				m.viewMode = ViewThemePicker
			},
			wantGoBack: true,
		},
		{
			name: "dev queue — q goes back",
			setup: func(m *MainModel) {
				m.devQueueView = &DevQueueView{}
				m.viewMode = ViewDevQueue
			},
			wantGoBack: true,
		},
		{
			name: "feedback (selection mode) — q goes back",
			setup: func(m *MainModel) {
				task := m.pool.GetAllTasks()[0]
				m.feedbackView = NewFeedbackView(task)
				m.viewMode = ViewFeedback
			},
			wantGoBack: true,
		},
		{
			name: "search view (text input) — q is text",
			setup: func(m *MainModel) {
				m.Update(keyMsg("/"))
			},
		},
		{
			name: "add task view (text input) — q is text",
			setup: func(m *MainModel) {
				m.addTaskView = NewAddTaskView()
				m.viewMode = ViewAddTask
			},
		},
		{
			name: "improvement view (text input) — q is text",
			setup: func(m *MainModel) {
				m.improvementView = NewImprovementView()
				m.viewMode = ViewImprovement
			},
		},
		{
			name: "feedback (custom input) — q is text",
			setup: func(m *MainModel) {
				task := m.pool.GetAllTasks()[0]
				m.feedbackView = NewFeedbackView(task)
				m.feedbackView.isCustom = true
				m.viewMode = ViewFeedback
			},
		},
		{
			name: "mood (custom input) — q is text",
			setup: func(m *MainModel) {
				_, cmd := m.Update(keyMsg("m"))
				if cmd != nil {
					msg := cmd()
					m.Update(msg)
				}
				m.moodView.isCustom = true
			},
		},
		{
			name: "detail (blocker input) — q is text",
			setup: func(m *MainModel) {
				m.Update(keyMsg("w"))
				m.Update(keyMsg("enter"))
				m.detailView.mode = DetailModeBlockerInput
			},
		},
		{
			name: "detail (expand input) — q is text",
			setup: func(m *MainModel) {
				m.Update(keyMsg("w"))
				m.Update(keyMsg("enter"))
				m.detailView.mode = DetailModeExpandInput
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModel("task1", "task2", "task3", "task4", "task5")
			tt.setup(m)

			_, cmd := m.Update(keyMsg("q"))
			gotQuit := isQuitCmd(t, cmd)
			gotGoBack := isGoBackCmd(t, cmd)

			if gotQuit != tt.wantQuit {
				t.Errorf("'q' in %s: got quit=%v, want quit=%v", tt.name, gotQuit, tt.wantQuit)
			}
			if gotGoBack != tt.wantGoBack {
				t.Errorf("'q' in %s: got goBack=%v, want goBack=%v", tt.name, gotGoBack, tt.wantGoBack)
			}
		})
	}
}
