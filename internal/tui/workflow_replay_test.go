package tui

import (
	"bytes"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

// waitForApp waits for the ThreeDoors header to appear, indicating the app is ready.
func waitForApp(t *testing.T, tm *teatest.TestModel) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("ThreeDoors"))
	}, teatest.WithDuration(3*time.Second))
}

// execCmd executes a tea.Cmd if non-nil and feeds the resulting message back
// into the model. Returns any secondary command from the update.
func execCmd(m *MainModel, cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	_, nextCmd := m.Update(msg)
	return nextCmd
}

// --- Workflow 1: Door Selection → Detail View ---

func TestWorkflow_DoorSelection(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		expectedIndex int
	}{
		{"select left door with A", "a", 0},
		{"select middle door with W", "w", 1},
		{"select right door with D", "d", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModel("Task Alpha", "Task Beta", "Task Gamma")

			// Select a door.
			m.Update(keyMsg(tt.key))

			if m.doorsView.selectedDoorIndex != tt.expectedIndex {
				t.Errorf("expected selectedDoorIndex=%d, got %d", tt.expectedIndex, m.doorsView.selectedDoorIndex)
			}

			// Open the selected door.
			m.Update(keyMsg("enter"))

			if m.viewMode != ViewDetail {
				t.Errorf("expected ViewDetail, got %d", m.viewMode)
			}
			if m.detailView == nil {
				t.Fatal("expected detailView to be set")
			}
		})
	}
}

// --- Workflow 2: Re-roll Doors ---

func TestWorkflow_RerollDoors(t *testing.T) {
	// Use enough tasks so re-roll has variety.
	m := makeModel("T1", "T2", "T3", "T4", "T5", "T6")

	// Record initial doors.
	initialDoors := make([]string, len(m.doorsView.currentDoors))
	for i, d := range m.doorsView.currentDoors {
		initialDoors[i] = d.ID
	}

	// Press S to re-roll.
	_, cmd := m.Update(keyMsg("s"))
	execCmd(m, cmd)

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after re-roll, got %d", m.viewMode)
	}
	if len(m.doorsView.currentDoors) != 3 {
		t.Errorf("expected 3 doors after re-roll, got %d", len(m.doorsView.currentDoors))
	}
}

// --- Workflow 3: Complete Task ---

func TestWorkflow_CompleteTask(t *testing.T) {
	provider := &testProvider{}
	pool := makePool("Complete me", "Keep me A", "Keep me B")
	tracker := core.NewSessionTracker()
	m := NewMainModel(pool, tracker, provider, nil, false, nil)

	// Select left door and open it.
	m.Update(keyMsg("a"))
	m.Update(keyMsg("enter"))

	if m.viewMode != ViewDetail {
		t.Fatalf("expected ViewDetail, got %d", m.viewMode)
	}

	detailTaskID := m.detailView.task.ID

	// Press C to complete.
	_, cmd := m.Update(keyMsg("c"))
	execCmd(m, cmd)

	// Verify task was removed from pool.
	if m.pool.GetTask(detailTaskID) != nil {
		t.Error("completed task should be removed from pool")
	}

	// Verify provider was notified.
	if len(provider.completedIDs) == 0 {
		t.Fatal("expected provider.MarkComplete to be called")
	}
	if provider.completedIDs[0] != detailTaskID {
		t.Errorf("expected completed ID %s, got %s", detailTaskID, provider.completedIDs[0])
	}

	// Verify transition to next-steps view.
	if m.viewMode != ViewNextSteps {
		t.Errorf("expected ViewNextSteps after completion, got %d", m.viewMode)
	}
}

// --- Workflow 4: Mark Blocked with Blocker Text ---

func TestWorkflow_MarkBlocked(t *testing.T) {
	m := makeModel("Block me", "Other task A", "Other task B")

	// Select door and open detail.
	m.Update(keyMsg("a"))
	m.Update(keyMsg("enter"))

	if m.viewMode != ViewDetail {
		t.Fatalf("expected ViewDetail, got %d", m.viewMode)
	}

	detailTask := m.detailView.task

	// Press B to enter blocker input mode.
	m.detailView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if m.detailView.mode != DetailModeBlockerInput {
		t.Fatalf("expected DetailModeBlockerInput, got %d", m.detailView.mode)
	}

	// Type blocker text character by character.
	blockerText := "Waiting for API"
	for _, ch := range blockerText {
		m.detailView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	if m.detailView.blockerInput != blockerText {
		t.Errorf("expected blockerInput=%q, got %q", blockerText, m.detailView.blockerInput)
	}

	// Press Enter to submit.
	cmd := m.detailView.Update(tea.KeyMsg{Type: tea.KeyEnter})
	execCmd(m, cmd)

	// Verify task status.
	if detailTask.Status != core.StatusBlocked {
		t.Errorf("expected task status %s, got %s", core.StatusBlocked, detailTask.Status)
	}
	if detailTask.Blocker != blockerText {
		t.Errorf("expected blocker=%q, got %q", blockerText, detailTask.Blocker)
	}
}

// --- Workflow 5: Add Task via Search :add Command ---

func TestWorkflow_AddTaskViaSearch(t *testing.T) {
	provider := &testProvider{}
	pool := makePool("Existing A", "Existing B", "Existing C")
	tracker := core.NewSessionTracker()
	m := NewMainModel(pool, tracker, provider, nil, false, nil)

	initialCount := pool.Count()

	// Open search with / key.
	m.Update(keyMsg("/"))

	if m.viewMode != ViewSearch {
		t.Fatalf("expected ViewSearch, got %d", m.viewMode)
	}

	// Type :add command — set value directly since textinput handles character input.
	m.searchView.textInput.SetValue(":add Buy groceries")
	m.searchView.checkCommandMode()

	if !m.searchView.isCommandMode {
		t.Fatal("expected command mode to be active")
	}

	// Press Enter to execute the command.
	cmd := m.searchView.Update(tea.KeyMsg{Type: tea.KeyEnter})
	execCmd(m, cmd)

	// Verify task was added to pool.
	if pool.Count() != initialCount+1 {
		t.Errorf("expected pool count %d, got %d", initialCount+1, pool.Count())
	}

	// Find the new task.
	found := false
	for _, task := range pool.GetAllTasks() {
		if task.Text == "Buy groceries" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find 'Buy groceries' in pool")
	}
}

// --- Workflow 6: Door Feedback ---

func TestWorkflow_DoorFeedback(t *testing.T) {
	tests := []struct {
		name         string
		feedbackKey  string
		expectedType string
	}{
		{"blocked feedback", "1", "blocked"},
		{"not-now feedback", "2", "not-now"},
		{"needs-breakdown feedback", "3", "needs-breakdown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModel("Feedback task", "Other A", "Other B")

			// Select left door.
			m.Update(keyMsg("a"))

			// Press N for feedback.
			_, cmd := m.Update(keyMsg("n"))
			execCmd(m, cmd)

			if m.feedbackView == nil {
				t.Fatal("expected feedbackView to be set")
			}

			// Select feedback option.
			feedbackCmd := m.feedbackView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.feedbackKey)})

			if feedbackCmd == nil {
				t.Fatal("expected feedback command to be returned")
			}

			msg := feedbackCmd()
			dfm, ok := msg.(DoorFeedbackMsg)
			if !ok {
				t.Fatalf("expected DoorFeedbackMsg, got %T", msg)
			}
			if dfm.FeedbackType != tt.expectedType {
				t.Errorf("expected feedback type %q, got %q", tt.expectedType, dfm.FeedbackType)
			}
		})
	}
}

// --- Workflow 7: Detail View Esc Returns to Doors ---

func TestWorkflow_DetailViewEscReturns(t *testing.T) {
	m := makeModel("Task A", "Task B", "Task C")

	// Select door and enter detail.
	m.Update(keyMsg("a"))
	m.Update(keyMsg("enter"))

	if m.viewMode != ViewDetail {
		t.Fatalf("expected ViewDetail, got %d", m.viewMode)
	}

	// Press Esc to return.
	_, cmd := m.Update(keyMsg("esc"))
	execCmd(m, cmd)

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after Esc, got %d", m.viewMode)
	}
}

// --- Workflow 8: Search → Find Task → Open Detail ---

func TestWorkflow_SearchAndOpenTask(t *testing.T) {
	m := makeModel("Alpha task", "Beta task", "Gamma task")

	// Open search.
	m.Update(keyMsg("/"))

	if m.viewMode != ViewSearch {
		t.Fatalf("expected ViewSearch, got %d", m.viewMode)
	}

	// Set search query and populate results.
	m.searchView.textInput.SetValue("Alpha")
	m.searchView.results = m.searchView.filterTasks("Alpha")
	if len(m.searchView.results) > 0 {
		m.searchView.selectedIndex = 0
	}

	if len(m.searchView.results) == 0 {
		t.Fatal("expected search results for 'Alpha'")
	}

	// Press Enter to select the search result.
	cmd := m.searchView.Update(tea.KeyMsg{Type: tea.KeyEnter})
	execCmd(m, cmd)

	if m.viewMode != ViewDetail {
		t.Errorf("expected ViewDetail after search selection, got %d", m.viewMode)
	}
	if m.detailView == nil {
		t.Fatal("expected detailView to be set")
	}
	if m.detailView.task.Text != "Alpha task" {
		t.Errorf("expected task 'Alpha task', got %q", m.detailView.task.Text)
	}
}

// --- Workflow 9: Mark In-Progress ---

func TestWorkflow_MarkInProgress(t *testing.T) {
	m := makeModel("Do this", "Other A", "Other B")

	// Select and open door.
	m.Update(keyMsg("a"))
	m.Update(keyMsg("enter"))

	detailTask := m.detailView.task

	// Press I to mark in-progress.
	cmd := m.detailView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	execCmd(m, cmd)

	if detailTask.Status != core.StatusInProgress {
		t.Errorf("expected status %s, got %s", core.StatusInProgress, detailTask.Status)
	}

	// After TaskUpdatedMsg, should return to doors.
	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after status change, got %d", m.viewMode)
	}
}

// --- Workflow 10: Full Journey (Select → Complete → Verify Pool) ---

func TestWorkflow_FullJourney_SelectCompleteReturn(t *testing.T) {
	provider := &testProvider{}
	pool := makePool("Journey task", "Filler A", "Filler B", "Filler C")
	tracker := core.NewSessionTracker()
	m := NewMainModel(pool, tracker, provider, nil, false, nil)

	// 1. Start at doors.
	if m.viewMode != ViewDoors {
		t.Fatalf("expected ViewDoors at start, got %d", m.viewMode)
	}

	// 2. Select left door.
	m.Update(keyMsg("a"))

	// 3. Open detail.
	m.Update(keyMsg("enter"))
	if m.viewMode != ViewDetail {
		t.Fatalf("expected ViewDetail, got %d", m.viewMode)
	}

	completedTaskID := m.detailView.task.ID

	// 4. Complete task.
	_, cmd := m.Update(keyMsg("c"))
	execCmd(m, cmd)

	// 5. Should be in NextSteps view.
	if m.viewMode != ViewNextSteps {
		t.Errorf("expected ViewNextSteps, got %d", m.viewMode)
	}

	// 6. Verify task removed.
	if m.pool.GetTask(completedTaskID) != nil {
		t.Error("completed task should be removed from pool")
	}

	// 7. Verify provider notified.
	if len(provider.completedIDs) != 1 || provider.completedIDs[0] != completedTaskID {
		t.Error("provider.MarkComplete not called correctly")
	}
}

// --- Workflow 11: Blocker Cancel (Esc during blocker input) ---

func TestWorkflow_BlockerInputCancel(t *testing.T) {
	m := makeModel("Cancel blocker", "Other A", "Other B")

	// Navigate to detail.
	m.Update(keyMsg("a"))
	m.Update(keyMsg("enter"))

	detailTask := m.detailView.task
	originalStatus := detailTask.Status

	// Enter blocker input mode.
	m.detailView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if m.detailView.mode != DetailModeBlockerInput {
		t.Fatalf("expected blocker input mode")
	}

	// Type some text.
	for _, ch := range "partial" {
		m.detailView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Cancel with Esc.
	m.detailView.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if m.detailView.mode != DetailModeView {
		t.Errorf("expected DetailModeView after cancel, got %d", m.detailView.mode)
	}
	if m.detailView.blockerInput != "" {
		t.Errorf("expected empty blockerInput after cancel, got %q", m.detailView.blockerInput)
	}
	if detailTask.Status != originalStatus {
		t.Errorf("expected status unchanged at %s, got %s", originalStatus, detailTask.Status)
	}
}

// --- Workflow 12: Multiple Re-rolls ---

func TestWorkflow_MultipleRerolls(t *testing.T) {
	m := makeModel("T1", "T2", "T3", "T4", "T5", "T6", "T7", "T8", "T9", "T10")

	// Reroll 3 times in succession.
	for i := range 3 {
		_, cmd := m.Update(keyMsg("s"))
		execCmd(m, cmd)

		if m.viewMode != ViewDoors {
			t.Errorf("reroll %d: expected ViewDoors, got %d", i+1, m.viewMode)
		}
		if len(m.doorsView.currentDoors) != 3 {
			t.Errorf("reroll %d: expected 3 doors, got %d", i+1, len(m.doorsView.currentDoors))
		}
	}
}

// --- Workflow 13: Search Close with Esc ---

func TestWorkflow_SearchCloseEsc(t *testing.T) {
	m := makeModel("Task A", "Task B", "Task C")

	// Open search.
	m.Update(keyMsg("/"))

	if m.viewMode != ViewSearch {
		t.Fatalf("expected ViewSearch, got %d", m.viewMode)
	}

	// Close with Esc.
	cmd := m.searchView.Update(tea.KeyMsg{Type: tea.KeyEscape})
	execCmd(m, cmd)

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after Esc from search, got %d", m.viewMode)
	}
}

// --- Workflow 14: Feedback Custom Comment ---

func TestWorkflow_FeedbackCustomComment(t *testing.T) {
	m := makeModel("Custom feedback task", "Other A", "Other B")

	// Select door and request feedback.
	m.Update(keyMsg("a"))
	_, cmd := m.Update(keyMsg("n"))
	execCmd(m, cmd)

	if m.feedbackView == nil {
		t.Fatal("expected feedbackView to be set")
	}

	// Press 4 for custom comment.
	m.feedbackView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})

	if !m.feedbackView.isCustom {
		t.Fatal("expected custom input mode")
	}

	// Type custom comment.
	for _, ch := range "Needs rethinking" {
		m.feedbackView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Submit with Enter.
	feedbackCmd := m.feedbackView.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if feedbackCmd == nil {
		t.Fatal("expected feedback command")
	}

	msg := feedbackCmd()
	dfm, ok := msg.(DoorFeedbackMsg)
	if !ok {
		t.Fatalf("expected DoorFeedbackMsg, got %T", msg)
	}
	if dfm.FeedbackType != "other" {
		t.Errorf("expected feedback type 'other', got %q", dfm.FeedbackType)
	}
	if dfm.Comment != "Needs rethinking" {
		t.Errorf("expected comment 'Needs rethinking', got %q", dfm.Comment)
	}
}

// --- Workflow 15: E2E teatest — Door Selection Replay ---

func TestWorkflow_Teatest_DoorSelectAndQuit(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task One", "Task Two", "Task Three"),
	)
	waitForApp(t, tm)

	// Select left door.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	time.Sleep(200 * time.Millisecond)

	// Open detail.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for detail view.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("TASK DETAILS"))
	}, teatest.WithDuration(2*time.Second))

	// Return to doors.
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})
	time.Sleep(200 * time.Millisecond)

	// Quit.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	fm := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second))
	if fm == nil {
		t.Fatal("expected non-nil final model")
	}
}

// --- Workflow 16: E2E teatest — Reroll Replay ---

func TestWorkflow_Teatest_RerollReplay(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("A1", "A2", "A3", "A4", "A5", "A6"),
	)
	waitForApp(t, tm)

	// Reroll twice.
	for range 2 {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		time.Sleep(300 * time.Millisecond)
	}

	// Should still be showing doors — look for door border characters
	// which appear in every door render cycle.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		// Door borders use box-drawing characters in the output.
		return bytes.Contains(bts, []byte("│"))
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	fm := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second))
	if _, ok := fm.(*MainModel); !ok {
		t.Errorf("expected *MainModel, got %T", fm)
	}
}
