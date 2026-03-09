package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func makeSelectTasks(n int) []*core.Task {
	tasks := make([]*core.Task, n)
	for i := range n {
		t := core.NewTask("Task " + string(rune('A'+i)))
		t.Status = core.StatusTodo
		t.Type = core.TypeTechnical
		tasks[i] = t
	}
	return tasks
}

func makeEnergyTasks() []*core.Task {
	high := core.NewTask("Deep work task")
	high.Effort = core.EffortDeepWork
	high.Status = core.StatusTodo

	medium := core.NewTask("Medium task")
	medium.Effort = core.EffortMedium
	medium.Status = core.StatusTodo

	low := core.NewTask("Quick win")
	low.Effort = core.EffortQuickWin
	low.Status = core.StatusTodo

	noEffort := core.NewTask("Untagged task")
	noEffort.Status = core.StatusTodo

	return []*core.Task{high, medium, low, noEffort}
}

func selectKeyMsg(k string) tea.KeyMsg {
	if k == " " {
		return tea.KeyMsg{Type: tea.KeySpace}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func selectSpecialKeyMsg(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

// --- AC 27.3.1: SelectView Bubbletea Model ---

func TestNewSelectView(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	if sv == nil {
		t.Fatal("NewSelectView returned nil")
	}
}

func TestSelectViewDisplaysTaskList(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	// No effort tags = match any energy
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)
	output := sv.View()
	if !strings.Contains(output, "Task A") {
		t.Errorf("expected Task A in view, got:\n%s", output)
	}
}

func TestSelectViewShowsSelectionCount(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)
	output := sv.View()
	if !strings.Contains(output, "Focus: 0/3 selected") {
		t.Errorf("expected selection counter in view, got:\n%s", output)
	}
}

func TestSelectViewSelectionCountUpdates(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	sv.Update(selectSpecialKeyMsg(tea.KeySpace))

	output := sv.View()
	if !strings.Contains(output, "Focus: 1/3 selected") {
		t.Errorf("expected Focus: 1/3 selected, got:\n%s", output)
	}
}

// --- AC 27.3.2: Energy Level Display and Override ---

func TestSelectViewShowsEnergyLevel(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)
	output := sv.View()
	if !strings.Contains(output, "High") {
		t.Errorf("expected energy level in view, got:\n%s", output)
	}
}

func TestSelectViewEnergyKeyCycles(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// E cycles: High -> Medium
	sv.Update(selectKeyMsg("e"))
	if sv.energy != string(core.EnergyMedium) {
		t.Errorf("expected energy Medium after first E, got %s", sv.energy)
	}

	// Medium -> Low
	sv.Update(selectKeyMsg("e"))
	if sv.energy != string(core.EnergyLow) {
		t.Errorf("expected energy Low after second E, got %s", sv.energy)
	}

	// Low -> All
	sv.Update(selectKeyMsg("e"))
	if sv.energy != EnergyAll {
		t.Errorf("expected energy All after third E, got %s", sv.energy)
	}

	// All -> High
	sv.Update(selectKeyMsg("e"))
	if sv.energy != string(core.EnergyHigh) {
		t.Errorf("expected energy High after fourth E, got %s", sv.energy)
	}
}

func TestSelectViewEnergyOverrideTracked(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)

	if sv.energyOverride {
		t.Error("expected no override initially")
	}

	sv.Update(selectKeyMsg("e"))

	if !sv.energyOverride {
		t.Error("expected override flag after E key")
	}
}

func TestSelectViewEnergyFiltersTasks(t *testing.T) {
	t.Parallel()
	tasks := makeEnergyTasks()
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// High energy: deep-work + untagged
	output := sv.View()
	if !strings.Contains(output, "Deep work task") {
		t.Error("expected deep-work task in high energy filter")
	}
	if strings.Contains(output, "Quick win") {
		t.Error("quick-win should not appear in high energy filter")
	}

	// Switch to Low
	sv.Update(selectKeyMsg("e")) // Medium
	sv.Update(selectKeyMsg("e")) // Low
	output = sv.View()
	if !strings.Contains(output, "Quick win") {
		t.Error("expected quick-win in low energy filter")
	}
}

func TestSelectViewEnergyAllShowsEverything(t *testing.T) {
	t.Parallel()
	tasks := makeEnergyTasks()
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Cycle to All
	sv.Update(selectKeyMsg("e")) // Medium
	sv.Update(selectKeyMsg("e")) // Low
	sv.Update(selectKeyMsg("e")) // All

	output := sv.View()
	if !strings.Contains(output, "Deep work task") || !strings.Contains(output, "Quick win") {
		t.Error("expected all tasks visible in All energy mode")
	}
}

// --- AC 27.3.3: Focus Task Selection ---

func TestSelectViewSpaceTogglesSelection(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Select first task
	sv.Update(selectSpecialKeyMsg(tea.KeySpace))
	if sv.selectedCount() != 1 {
		t.Errorf("expected 1 selected, got %d", sv.selectedCount())
	}

	// Deselect first task
	sv.Update(selectSpecialKeyMsg(tea.KeySpace))
	if sv.selectedCount() != 0 {
		t.Errorf("expected 0 selected after toggle off, got %d", sv.selectedCount())
	}
}

func TestSelectViewFocusTagAdded(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	sv.Update(selectSpecialKeyMsg(tea.KeySpace))

	if !core.HasFocusTag(tasks[0]) {
		t.Error("expected +focus tag on selected task")
	}
}

func TestSelectViewFocusTagRemoved(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Select then deselect
	sv.Update(selectSpecialKeyMsg(tea.KeySpace))
	sv.Update(selectSpecialKeyMsg(tea.KeySpace))

	if core.HasFocusTag(tasks[0]) {
		t.Error("expected +focus tag removed after deselection")
	}
}

func TestSelectViewTargetReachedMessage(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(5)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Select 3 tasks
	for i := 0; i < 3; i++ {
		sv.Update(selectSpecialKeyMsg(tea.KeySpace))
		sv.Update(selectSpecialKeyMsg(tea.KeyDown))
	}

	output := sv.View()
	if !strings.Contains(output, "Target reached") {
		t.Errorf("expected 'Target reached!' message at 3 selected, got:\n%s", output)
	}
}

func TestSelectViewMaxFiveSelection(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(7)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Select 5 tasks
	for i := 0; i < 5; i++ {
		sv.Update(selectSpecialKeyMsg(tea.KeySpace))
		sv.Update(selectSpecialKeyMsg(tea.KeyDown))
	}

	// Try to select a 6th
	sv.Update(selectSpecialKeyMsg(tea.KeySpace))

	if sv.selectedCount() != 5 {
		t.Errorf("expected max 5 selected, got %d", sv.selectedCount())
	}

	output := sv.View()
	if !strings.Contains(output, "Maximum 5 focus tasks") {
		t.Errorf("expected max message, got:\n%s", output)
	}
}

func TestSelectViewShowsCheckbox(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	output := sv.View()
	if !strings.Contains(output, "[ ]") {
		t.Error("expected unchecked checkbox in view")
	}

	sv.Update(selectSpecialKeyMsg(tea.KeySpace))
	output = sv.View()
	if !strings.Contains(output, "[x]") {
		t.Error("expected checked checkbox after selection")
	}
}

func TestSelectViewEnterConfirmsSelection(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Select one task
	sv.Update(selectSpecialKeyMsg(tea.KeySpace))

	cmd := sv.Update(selectSpecialKeyMsg(tea.KeyEnter))
	if cmd == nil {
		t.Fatal("expected command on Enter with selections")
	}
	msg := cmd()
	completeMsg, ok := msg.(SelectCompleteMsg)
	if !ok {
		t.Fatalf("expected SelectCompleteMsg, got %T", msg)
	}
	if len(completeMsg.FocusTasks) != 1 {
		t.Errorf("expected 1 focus task, got %d", len(completeMsg.FocusTasks))
	}
}

// --- AC 27.3.4: Task List Presentation ---

func TestSelectViewShowsEffortTag(t *testing.T) {
	t.Parallel()
	tasks := makeEnergyTasks()
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)
	output := sv.View()
	if !strings.Contains(output, "deep-work") {
		t.Errorf("expected effort tag in view, got:\n%s", output)
	}
}

func TestSelectViewShowsStatus(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	tasks[0].Status = core.StatusInProgress
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)
	output := sv.View()
	if !strings.Contains(output, "in-progress") {
		t.Errorf("expected status in view, got:\n%s", output)
	}
}

func TestSelectViewArrowNavigation(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	if sv.cursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", sv.cursor)
	}

	sv.Update(selectSpecialKeyMsg(tea.KeyDown))
	if sv.cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", sv.cursor)
	}

	sv.Update(selectSpecialKeyMsg(tea.KeyUp))
	if sv.cursor != 0 {
		t.Errorf("expected cursor at 0 after up, got %d", sv.cursor)
	}
}

func TestSelectViewVimNavigation(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)

	sv.Update(selectKeyMsg("j"))
	if sv.cursor != 1 {
		t.Errorf("expected cursor at 1 after j, got %d", sv.cursor)
	}

	sv.Update(selectKeyMsg("k"))
	if sv.cursor != 0 {
		t.Errorf("expected cursor at 0 after k, got %d", sv.cursor)
	}
}

func TestSelectViewCursorBounds(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)

	// Go up from 0
	sv.Update(selectSpecialKeyMsg(tea.KeyUp))
	if sv.cursor != 0 {
		t.Errorf("expected cursor clamped at 0, got %d", sv.cursor)
	}

	// Go past end
	sv.Update(selectSpecialKeyMsg(tea.KeyDown))
	sv.Update(selectSpecialKeyMsg(tea.KeyDown))
	sv.Update(selectSpecialKeyMsg(tea.KeyDown))
	if sv.cursor != 1 {
		t.Errorf("expected cursor clamped at 1, got %d", sv.cursor)
	}
}

// --- AC 27.3.5: Edge Cases ---

func TestSelectViewFewerThanThreeTasks(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Select both tasks
	sv.Update(selectSpecialKeyMsg(tea.KeySpace))
	sv.Update(selectSpecialKeyMsg(tea.KeyDown))
	sv.Update(selectSpecialKeyMsg(tea.KeySpace))

	if sv.selectedCount() != 2 {
		t.Errorf("expected 2 selected with fewer tasks, got %d", sv.selectedCount())
	}
}

func TestSelectViewNoMatchingTasks(t *testing.T) {
	t.Parallel()
	tasks := makeEnergyTasks()
	// All tasks have specific effort; filter for energy that only matches some
	// Create tasks that only match low energy
	lowOnly := []*core.Task{tasks[2]} // quick-win only
	sv := NewSelectView(lowOnly, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	output := sv.View()
	if !strings.Contains(output, "No matching tasks") {
		t.Errorf("expected no matching tasks message, got:\n%s", output)
	}
}

func TestSelectViewZeroSelectionWarning(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Press Enter with zero selections
	cmd := sv.Update(selectSpecialKeyMsg(tea.KeyEnter))
	if cmd != nil {
		t.Error("expected no command on Enter with zero selections (should prompt)")
	}

	output := sv.View()
	if !strings.Contains(output, "No focus tasks selected") {
		t.Errorf("expected zero-selection warning, got:\n%s", output)
	}
}

func TestSelectViewZeroSelectionConfirmYes(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Trigger zero-selection prompt
	sv.Update(selectSpecialKeyMsg(tea.KeyEnter))

	// Confirm with Y
	cmd := sv.Update(selectKeyMsg("y"))
	if cmd == nil {
		t.Fatal("expected command on Y confirmation")
	}
	msg := cmd()
	if _, ok := msg.(SelectCompleteMsg); !ok {
		t.Errorf("expected SelectCompleteMsg, got %T", msg)
	}
}

func TestSelectViewZeroSelectionConfirmNo(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Trigger zero-selection prompt
	sv.Update(selectSpecialKeyMsg(tea.KeyEnter))

	// Decline with N
	sv.Update(selectKeyMsg("n"))

	if sv.confirmEmpty {
		t.Error("expected confirmEmpty to be dismissed after N")
	}
}

// --- AC 27.3.6: Navigation ---

func TestSelectViewEscReturnsToReview(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)

	cmd := sv.Update(selectSpecialKeyMsg(tea.KeyEscape))
	if cmd == nil {
		t.Fatal("expected command on Esc")
	}
	msg := cmd()
	if _, ok := msg.(SelectCancelMsg); !ok {
		t.Errorf("expected SelectCancelMsg on Esc, got %T", msg)
	}
}

func TestSelectViewHelpOverlay(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	sv.Update(selectKeyMsg("?"))
	output := sv.View()

	if !strings.Contains(output, "Select Help") {
		t.Errorf("expected help overlay, got:\n%s", output)
	}
	if !strings.Contains(output, "Toggle focus selection") {
		t.Errorf("expected help content, got:\n%s", output)
	}
}

func TestSelectViewHelpOverlayDismiss(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)

	sv.Update(selectKeyMsg("?"))
	sv.Update(selectKeyMsg("?"))

	if sv.showHelp {
		t.Error("expected help dismissed after second ?")
	}
}

func TestSelectViewShowsStepIndicator(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)
	output := sv.View()

	if !strings.Contains(output, "Step 2/3") {
		t.Errorf("expected step indicator 'Step 2/3' in view, got:\n%s", output)
	}
}

func TestSelectViewShowsActionKeys(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(2)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)
	output := sv.View()

	if !strings.Contains(output, "[Space]") || !strings.Contains(output, "[E]nergy") {
		t.Errorf("expected action key hints, got:\n%s", output)
	}
}

// --- SelectCompleteMsg carries correct data ---

func TestSelectCompleteMsgCarriesEnergyOverride(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(3)
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Override energy
	sv.Update(selectKeyMsg("e"))

	// Select a task
	sv.Update(selectSpecialKeyMsg(tea.KeySpace))

	cmd := sv.Update(selectSpecialKeyMsg(tea.KeyEnter))
	msg := cmd().(SelectCompleteMsg)

	if !msg.EnergyOverride {
		t.Error("expected EnergyOverride true in complete msg")
	}
	if msg.EnergyLevel != core.EnergyMedium {
		t.Errorf("expected energy Medium, got %s", msg.EnergyLevel)
	}
}

func TestSelectViewSetWidth(t *testing.T) {
	t.Parallel()
	sv := NewSelectView(makeSelectTasks(1), core.EnergyHigh)
	sv.SetWidth(120)
	if sv.width != 120 {
		t.Errorf("expected width 120, got %d", sv.width)
	}
}

func TestSelectViewEmptyPool(t *testing.T) {
	t.Parallel()
	sv := NewSelectView(nil, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)
	output := sv.View()
	if !strings.Contains(output, "No matching tasks") {
		t.Errorf("expected empty state, got:\n%s", output)
	}
}

func TestSelectViewFocusTagNotDuplicatedOnReselect(t *testing.T) {
	t.Parallel()
	tasks := makeSelectTasks(1)
	// Pre-tag the task with +focus
	tasks[0].Text = tasks[0].Text + " +focus"
	sv := NewSelectView(tasks, core.EnergyHigh)
	sv.SetWidth(80)
	sv.SetHeight(40)

	// Toggle on — should not add duplicate +focus
	sv.selected[tasks[0].ID] = true
	sv.Update(selectSpecialKeyMsg(tea.KeySpace)) // toggle off
	sv.Update(selectSpecialKeyMsg(tea.KeySpace)) // toggle on

	count := strings.Count(strings.ToLower(tasks[0].Text), "+focus")
	if count > 1 {
		t.Errorf("expected at most 1 +focus tag, got %d in: %q", count, tasks[0].Text)
	}
}
