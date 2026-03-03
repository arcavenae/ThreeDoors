package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

// --- helpers ---

// testProvider is a no-op TaskProvider for TUI testing.
type testProvider struct {
	completedIDs []string
	completeErr  error
	savedTasks   []*tasks.Task
	saveErr      error
}

func (p *testProvider) LoadTasks() ([]*tasks.Task, error) { return nil, nil }
func (p *testProvider) SaveTask(t *tasks.Task) error      { return p.saveErr }

func (p *testProvider) SaveTasks(ts []*tasks.Task) error {
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

func makePool(texts ...string) *tasks.TaskPool {
	pool := tasks.NewTaskPool()
	for _, t := range texts {
		pool.AddTask(tasks.NewTask(t))
	}
	return pool
}

func makeModel(texts ...string) *MainModel {
	return makeModelWithProvider(&testProvider{}, texts...)
}

func makeModelWithProvider(provider tasks.TaskProvider, texts ...string) *MainModel {
	pool := makePool(texts...)
	tracker := tasks.NewSessionTracker()
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

func TestInit_ReturnsNil(t *testing.T) {
	m := makeModel("task1")
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
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
	pool := tasks.NewTaskPool()
	tracker := tasks.NewSessionTracker()
	m := NewMainModel(pool, tracker, tasks.NewTextFileProvider(), nil, false, nil)
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

	newTask := tasks.NewTask("brand new task")
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
	pool := tasks.NewTaskPool()
	tracker := tasks.NewSessionTracker()
	m := NewMainModel(pool, tracker, tasks.NewTextFileProvider(), nil, false, nil)

	if m.pool.Count() != 0 {
		t.Fatal("pool should start empty")
	}

	newTask := tasks.NewTask("first task ever")
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

	newTask := tasks.NewTask("new task from search")
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

	newTask := tasks.NewTask("new task from doors")
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
