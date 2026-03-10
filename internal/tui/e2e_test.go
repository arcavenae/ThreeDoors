package tui

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

// sendKey is a test helper that sends a rune key to the test model.
func sendKey(t *testing.T, tm *teatest.TestModel, r rune) {
	t.Helper()
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
}

// sendSpecialKey is a test helper that sends a special key (enter, esc, etc.) to the test model.
func sendSpecialKey(t *testing.T, tm *teatest.TestModel, keyType tea.KeyType) {
	t.Helper()
	tm.Send(tea.KeyMsg{Type: keyType})
}

// waitForContent is a test helper that waits for the output to contain the given text.
func waitForContent(t *testing.T, tm *teatest.TestModel, text string) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte(text))
	}, teatest.WithDuration(5*time.Second))
}

// finalMainModel is a test helper that quits and returns the *MainModel from the final state.
func finalMainModel(t *testing.T, tm *teatest.TestModel) *MainModel {
	t.Helper()
	sendKey(t, tm, 'q')
	// The quit may trigger an improvement prompt if conditions met;
	// send esc to skip it in case it appears.
	time.Sleep(100 * time.Millisecond)
	sendSpecialKey(t, tm, tea.KeyEsc)
	fm := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second))
	if fm == nil {
		t.Fatal("expected non-nil final model")
		return nil
	}
	mm, ok := fm.(*MainModel)
	if !ok {
		t.Fatalf("expected *MainModel, got %T", fm)
	}
	return mm
}

// --- AC1: Full User Workflow E2E Tests ---

func TestE2E_FullWorkflow_LaunchViewSelectManageExit(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Buy groceries", "Read a book", "Go for a run"),
	)

	// 1. Launch → verify doors render with "ThreeDoors" header.
	waitForContent(t, tm, "ThreeDoors")

	// 2. Select the left door (key 'a').
	sendKey(t, tm, 'a')
	time.Sleep(200 * time.Millisecond)

	// 3. Press Enter to open detail view.
	sendSpecialKey(t, tm, tea.KeyEnter)
	time.Sleep(200 * time.Millisecond)

	// 4. Complete the task (key 'c').
	sendKey(t, tm, 'c')
	time.Sleep(200 * time.Millisecond)

	// 5. Dismiss next-steps view (Esc back to doors).
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(200 * time.Millisecond)

	// 6. Quit cleanly.
	mm := finalMainModel(t, tm)

	// Verify session tracker recorded the workflow.
	if mm.tracker == nil {
		t.Fatal("expected tracker to be initialized")
		return
	}
	metrics := mm.tracker.Finalize()
	if metrics.TasksCompleted < 1 {
		t.Errorf("expected at least 1 task completed, got %d", metrics.TasksCompleted)
	}
}

func TestE2E_DoorSelection_AllThreeDoors(t *testing.T) {
	tests := []struct {
		name     string
		key      rune
		doorIdx  int
		doorName string
	}{
		{"left door via A", 'a', 0, "left"},
		{"center door via W", 'w', 1, "center"},
		{"right door via D", 'd', 2, "right"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTestApp(t,
				WithTasks("Task Alpha", "Task Beta", "Task Gamma"),
			)

			waitForContent(t, tm, "ThreeDoors")

			// Select door.
			sendKey(t, tm, tt.key)
			time.Sleep(100 * time.Millisecond)

			// Confirm selection with Enter.
			sendSpecialKey(t, tm, tea.KeyEnter)
			time.Sleep(200 * time.Millisecond)

			// We should now be in detail view — press Esc to return.
			sendSpecialKey(t, tm, tea.KeyEsc)
			time.Sleep(100 * time.Millisecond)

			// Quit and verify.
			mm := finalMainModel(t, tm)
			metrics := mm.tracker.Finalize()

			if len(metrics.DoorSelections) < 1 {
				t.Errorf("expected at least 1 door selection, got %d", len(metrics.DoorSelections))
			}
			if len(metrics.DoorSelections) > 0 && metrics.DoorSelections[0].DoorPosition != tt.doorIdx {
				t.Errorf("expected door position %d, got %d", tt.doorIdx, metrics.DoorSelections[0].DoorPosition)
			}
		})
	}
}

func TestE2E_RerollDoors(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("T1", "T2", "T3", "T4", "T5", "T6"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Reroll doors twice with 's' key.
	sendKey(t, tm, 's')
	time.Sleep(200 * time.Millisecond)
	sendKey(t, tm, 's')
	time.Sleep(200 * time.Millisecond)

	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	if metrics.RefreshesUsed != 2 {
		t.Errorf("expected 2 refreshes, got %d", metrics.RefreshesUsed)
	}
}

// --- AC2: Session Metrics Verification Tests ---

func TestE2E_SessionMetrics_DoorsViewedAndDetailViews(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task 1", "Task 2", "Task 3"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Select left door and open detail.
	sendKey(t, tm, 'a')
	time.Sleep(100 * time.Millisecond)
	sendSpecialKey(t, tm, tea.KeyEnter)
	time.Sleep(200 * time.Millisecond)

	// Return to doors.
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(100 * time.Millisecond)

	// Select center door and open detail.
	sendKey(t, tm, 'w')
	time.Sleep(100 * time.Millisecond)
	sendSpecialKey(t, tm, tea.KeyEnter)
	time.Sleep(200 * time.Millisecond)

	// Return to doors.
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(100 * time.Millisecond)

	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	if metrics.DoorsViewed < 2 {
		t.Errorf("expected at least 2 doors viewed, got %d", metrics.DoorsViewed)
	}
	if metrics.DetailViews < 2 {
		t.Errorf("expected at least 2 detail views, got %d", metrics.DetailViews)
	}
	if len(metrics.DoorSelections) < 2 {
		t.Errorf("expected at least 2 door selections, got %d", len(metrics.DoorSelections))
	}
}

func TestE2E_SessionMetrics_TaskCompletion(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Complete me", "Keep me", "Also keep"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Select and complete a task.
	sendKey(t, tm, 'a')
	time.Sleep(100 * time.Millisecond)
	sendSpecialKey(t, tm, tea.KeyEnter)
	time.Sleep(200 * time.Millisecond)
	sendKey(t, tm, 'c')
	time.Sleep(200 * time.Millisecond)

	// Dismiss next-steps.
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(100 * time.Millisecond)

	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	if metrics.TasksCompleted != 1 {
		t.Errorf("expected 1 task completed, got %d", metrics.TasksCompleted)
	}
	if metrics.StatusChanges < 1 {
		t.Errorf("expected at least 1 status change, got %d", metrics.StatusChanges)
	}
}

func TestE2E_SessionMetrics_RefreshTracking(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("A", "B", "C", "D", "E", "F", "G"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Refresh 3 times.
	for range 3 {
		sendKey(t, tm, 's')
		time.Sleep(150 * time.Millisecond)
	}

	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	if metrics.RefreshesUsed != 3 {
		t.Errorf("expected 3 refreshes, got %d", metrics.RefreshesUsed)
	}
	// Each refresh should have recorded bypassed tasks.
	if len(metrics.TaskBypasses) != 3 {
		t.Errorf("expected 3 bypass records, got %d", len(metrics.TaskBypasses))
	}
}

func TestE2E_SessionMetrics_MoodEntries(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task 1", "Task 2", "Task 3"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Open mood view.
	sendKey(t, tm, 'm')
	time.Sleep(200 * time.Millisecond)

	// Select "Focused" (key '1').
	sendKey(t, tm, '1')
	time.Sleep(200 * time.Millisecond)

	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	if metrics.MoodEntryCount != 1 {
		t.Errorf("expected 1 mood entry, got %d", metrics.MoodEntryCount)
	}
	if len(metrics.MoodEntries) != 1 {
		t.Errorf("expected 1 mood detail entry, got %d", len(metrics.MoodEntries))
	}
	if len(metrics.MoodEntries) > 0 && metrics.MoodEntries[0].Mood != "Focused" {
		t.Errorf("expected mood 'Focused', got %q", metrics.MoodEntries[0].Mood)
	}
}

func TestE2E_SessionMetrics_SessionIDAndTiming(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task 1", "Task 2", "Task 3"),
	)

	waitForContent(t, tm, "ThreeDoors")

	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	if metrics.SessionID == "" {
		t.Error("expected non-empty session ID")
	}
	if metrics.StartTime.IsZero() {
		t.Error("expected non-zero start time")
	}
	if metrics.EndTime.IsZero() {
		t.Error("expected non-zero end time")
	}
	if metrics.DurationSeconds <= 0 {
		t.Error("expected positive duration")
	}
}

// --- AC3: Search and Command Palette Workflow Tests ---

func TestE2E_SearchWorkflow_OpenSearchAndType(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Buy groceries", "Read a book", "Exercise daily"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Open search view with '/'.
	sendKey(t, tm, '/')
	time.Sleep(200 * time.Millisecond)

	// Type a search query character by character.
	for _, r := range "Buy" {
		sendKey(t, tm, r)
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for search results to include our match.
	waitForContent(t, tm, "Buy groceries")

	// Close search with Esc.
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(100 * time.Millisecond)

	mm := finalMainModel(t, tm)
	if mm.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after closing search, got %d", mm.viewMode)
	}
}

func TestE2E_SearchWorkflow_SelectSearchResult(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Buy groceries", "Read a book", "Exercise daily"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Open search.
	sendKey(t, tm, '/')
	time.Sleep(200 * time.Millisecond)

	// Search for "Read".
	for _, r := range "Read" {
		sendKey(t, tm, r)
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)

	// Navigate down to select first result.
	sendSpecialKey(t, tm, tea.KeyDown)
	time.Sleep(100 * time.Millisecond)

	// Open the result with Enter.
	sendSpecialKey(t, tm, tea.KeyEnter)
	time.Sleep(200 * time.Millisecond)

	// We should now be in detail view — Esc returns to search (previousView=Search).
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(100 * time.Millisecond)

	// Now we're back in search view — Esc again to close search and return to doors.
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(100 * time.Millisecond)

	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	// Opening a task from search should increment detail views.
	if metrics.DetailViews < 1 {
		t.Errorf("expected at least 1 detail view from search, got %d", metrics.DetailViews)
	}
}

func TestE2E_CommandPalette_MoodCommand(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task 1", "Task 2", "Task 3"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Open command palette with ':'.
	sendKey(t, tm, ':')
	time.Sleep(200 * time.Millisecond)

	// Type "mood" command.
	for _, r := range "mood" {
		sendKey(t, tm, r)
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)

	// Select the :mood command.
	sendSpecialKey(t, tm, tea.KeyDown)
	time.Sleep(100 * time.Millisecond)
	sendSpecialKey(t, tm, tea.KeyEnter)
	time.Sleep(200 * time.Millisecond)

	// Now we should be in mood view. Select "Energized" (key '4').
	sendKey(t, tm, '4')
	time.Sleep(200 * time.Millisecond)

	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	if metrics.MoodEntryCount < 1 {
		t.Errorf("expected at least 1 mood entry from command palette, got %d", metrics.MoodEntryCount)
	}
}

func TestE2E_CommandPalette_HelpCommand(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task 1", "Task 2", "Task 3"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Open command palette.
	sendKey(t, tm, ':')
	time.Sleep(200 * time.Millisecond)

	// Type "help".
	for _, r := range "help" {
		sendKey(t, tm, r)
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)

	// Verify :help appears in command palette results.
	waitForContent(t, tm, ":help")

	// Close with Esc.
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(100 * time.Millisecond)

	_ = finalMainModel(t, tm)
}

// --- AC3 continued: Mood Tracking Workflow Tests ---

func TestE2E_MoodTracking_AllOptions(t *testing.T) {
	tests := []struct {
		name     string
		key      rune
		expected string
	}{
		{"Focused", '1', "Focused"},
		{"Tired", '2', "Tired"},
		{"Stressed", '3', "Stressed"},
		{"Energized", '4', "Energized"},
		{"Distracted", '5', "Distracted"},
		{"Calm", '6', "Calm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTestApp(t,
				WithTasks("Task 1", "Task 2", "Task 3"),
			)

			waitForContent(t, tm, "ThreeDoors")

			// Open mood via 'm' key.
			sendKey(t, tm, 'm')
			time.Sleep(200 * time.Millisecond)

			// Select mood.
			sendKey(t, tm, tt.key)
			time.Sleep(200 * time.Millisecond)

			mm := finalMainModel(t, tm)
			mood := mm.tracker.LatestMood()

			if mood != tt.expected {
				t.Errorf("expected mood %q, got %q", tt.expected, mood)
			}
		})
	}
}

func TestE2E_MoodTracking_EscCancels(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task 1", "Task 2", "Task 3"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// Open mood view.
	sendKey(t, tm, 'm')
	time.Sleep(200 * time.Millisecond)

	// Cancel with Esc.
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(200 * time.Millisecond)

	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	// No mood should have been recorded.
	if metrics.MoodEntryCount != 0 {
		t.Errorf("expected 0 mood entries after cancel, got %d", metrics.MoodEntryCount)
	}
}

// --- Combined Workflow Tests ---

func TestE2E_FullSession_MultipleActions(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task A", "Task B", "Task C", "Task D", "Task E", "Task F"),
	)

	waitForContent(t, tm, "ThreeDoors")

	// 1. Log a mood.
	sendKey(t, tm, 'm')
	time.Sleep(200 * time.Millisecond)
	sendKey(t, tm, '1') // Focused
	time.Sleep(200 * time.Millisecond)

	// 2. Reroll doors.
	sendKey(t, tm, 's')
	time.Sleep(200 * time.Millisecond)

	// 3. Select and complete a task.
	sendKey(t, tm, 'a')
	time.Sleep(100 * time.Millisecond)
	sendSpecialKey(t, tm, tea.KeyEnter)
	time.Sleep(200 * time.Millisecond)
	sendKey(t, tm, 'c')
	time.Sleep(200 * time.Millisecond)

	// Dismiss next-steps.
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(100 * time.Millisecond)

	// 4. Search for a task.
	sendKey(t, tm, '/')
	time.Sleep(200 * time.Millisecond)
	for _, r := range "Task" {
		sendKey(t, tm, r)
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)

	// Close search.
	sendSpecialKey(t, tm, tea.KeyEsc)
	time.Sleep(100 * time.Millisecond)

	// 5. Quit and verify comprehensive metrics.
	mm := finalMainModel(t, tm)
	metrics := mm.tracker.Finalize()

	if metrics.MoodEntryCount != 1 {
		t.Errorf("expected 1 mood entry, got %d", metrics.MoodEntryCount)
	}
	if metrics.RefreshesUsed != 1 {
		t.Errorf("expected 1 refresh, got %d", metrics.RefreshesUsed)
	}
	if metrics.TasksCompleted != 1 {
		t.Errorf("expected 1 task completed, got %d", metrics.TasksCompleted)
	}
	if metrics.DetailViews < 1 {
		t.Errorf("expected at least 1 detail view, got %d", metrics.DetailViews)
	}
	if len(metrics.DoorSelections) < 1 {
		t.Errorf("expected at least 1 door selection, got %d", len(metrics.DoorSelections))
	}
}
