package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

// --- helpers ---

func makePool(texts ...string) *tasks.TaskPool {
	pool := tasks.NewTaskPool()
	for _, t := range texts {
		pool.AddTask(tasks.NewTask(t))
	}
	return pool
}

func makeModel(texts ...string) *MainModel {
	pool := makePool(texts...)
	tracker := tasks.NewSessionTracker()
	return NewMainModel(pool, tracker)
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
	m := NewMainModel(pool, tracker)
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

	// Complete the task
	_, cmd := m.Update(keyMsg("c"))
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
		if m.flash == celebration {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("flash message %q is not in the celebrationMessages pool", m.flash)
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
