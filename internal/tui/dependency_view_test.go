package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// --- FormatBlockedBy ---

func TestFormatBlockedBy_Empty(t *testing.T) {
	t.Parallel()
	result := FormatBlockedBy(nil, 40)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestFormatBlockedBy_SingleBlocker(t *testing.T) {
	t.Parallel()
	blockers := []*core.Task{{Text: "Write deployment script"}}
	result := FormatBlockedBy(blockers, 40)
	expected := "Blocked by: Write deployment script"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatBlockedBy_MultipleBlockers(t *testing.T) {
	t.Parallel()
	blockers := []*core.Task{
		{Text: "Task B"},
		{Text: "Task C"},
		{Text: "Task D"},
	}
	result := FormatBlockedBy(blockers, 40)
	expected := "Blocked by: Task B (+2 more)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatBlockedBy_Truncation(t *testing.T) {
	t.Parallel()
	longText := "This is a very long task description that exceeds forty characters"
	blockers := []*core.Task{{Text: longText}}
	result := FormatBlockedBy(blockers, 40)
	if !strings.HasPrefix(result, "Blocked by: ") {
		t.Errorf("expected prefix 'Blocked by: ', got %q", result)
	}
	// Text portion should be 40 chars max (including "...")
	textPart := strings.TrimPrefix(result, "Blocked by: ")
	if len(textPart) > 40 {
		t.Errorf("text part should be at most 40 chars, got %d: %q", len(textPart), textPart)
	}
	if !strings.HasSuffix(textPart, "...") {
		t.Errorf("truncated text should end with '...', got %q", textPart)
	}
}

func TestFormatBlockedBy_TruncationWithMultiple(t *testing.T) {
	t.Parallel()
	longText := "This is a very long task description that exceeds forty characters easily"
	blockers := []*core.Task{
		{Text: longText},
		{Text: "Another task"},
	}
	result := FormatBlockedBy(blockers, 40)
	if !strings.Contains(result, "(+1 more)") {
		t.Errorf("expected '+1 more' suffix, got %q", result)
	}
}

// --- Detail View Dependency Rendering ---

func newTestDetailViewWithPool(text string) (*DetailView, *core.TaskPool) {
	pool := core.NewTaskPool()
	task := core.NewTask(text)
	pool.AddTask(task)
	dv := NewDetailView(task, nil, nil, pool)
	return dv, pool
}

func TestDetailView_NoDeps_NoDependencySection(t *testing.T) {
	t.Parallel()
	dv, _ := newTestDetailViewWithPool("task with no deps")
	dv.SetWidth(80)
	view := dv.View()
	if strings.Contains(view, "Dependencies") {
		t.Error("should not show Dependencies section when task has no dependencies")
	}
}

func TestDetailView_WithDeps_ShowsDependencySection(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	depTask := core.NewTask("blocking task")
	pool.AddTask(depTask)

	task := core.NewTask("dependent task")
	task.DependsOn = []string{depTask.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.SetWidth(80)
	view := dv.View()

	if !strings.Contains(view, "Dependencies") {
		t.Error("should show Dependencies section")
	}
	if !strings.Contains(view, "blocking task") {
		t.Error("should show dependency task text")
	}
	if !strings.Contains(view, "[ ]") {
		t.Error("should show unchecked checkbox for incomplete dependency")
	}
}

func TestDetailView_CompleteDep_ShowsChecked(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	depTask := core.NewTask("completed dep")
	_ = depTask.UpdateStatus(core.StatusComplete)
	pool.AddTask(depTask)

	task := core.NewTask("my task")
	task.DependsOn = []string{depTask.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.SetWidth(80)
	view := dv.View()

	if !strings.Contains(view, "[x]") {
		t.Error("should show checked checkbox for complete dependency")
	}
}

func TestDetailView_MixedDeps_ShowsBlockedBy(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()

	completeDep := core.NewTask("done dep")
	_ = completeDep.UpdateStatus(core.StatusComplete)
	pool.AddTask(completeDep)

	blockingDep := core.NewTask("blocking dep")
	pool.AddTask(blockingDep)

	task := core.NewTask("my task")
	task.DependsOn = []string{completeDep.ID, blockingDep.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.SetWidth(80)
	view := dv.View()

	if !strings.Contains(view, "Blocked by:") {
		t.Error("should show 'Blocked by:' when there are blocking deps")
	}
	if !strings.Contains(view, "blocking dep") {
		t.Error("should show the blocking dep text")
	}
}

func TestDetailView_OrphanedDep_ShowsDeletedTask(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("orphan dep task")
	task.DependsOn = []string{"nonexistent-id"}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.SetWidth(80)
	view := dv.View()

	if !strings.Contains(view, "[deleted task]") {
		t.Error("should show '[deleted task]' for orphaned dependency")
	}
}

// --- Dependency Help Hints ---

func TestDetailView_WithPool_ShowsDepHints(t *testing.T) {
	t.Parallel()
	dv, _ := newTestDetailViewWithPool("test task")
	dv.SetWidth(80)
	view := dv.View()

	if !strings.Contains(view, "[+]dep") {
		t.Error("should show [+]dep hint when pool is available")
	}
	if !strings.Contains(view, "[-]dep") {
		t.Error("should show [-]dep hint when pool is available")
	}
}

func TestDetailView_NoPool_NoDepHints(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("test task")
	dv.SetWidth(80)
	view := dv.View()

	if strings.Contains(view, "[+]dep") {
		t.Error("should not show [+]dep hint when pool is nil")
	}
}

// --- Add Dependency (+) Key ---

func TestDetailView_PlusKey_NoPool_ShowsFlash(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if cmd == nil {
		t.Fatal("'+' without pool should return a flash command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "not available") {
		t.Errorf("expected 'not available' message, got %q", fm.Text)
	}
}

func TestDetailView_PlusKey_NoCandidates_ShowsFlash(t *testing.T) {
	t.Parallel()
	// Pool with only the task itself
	dv, _ := newTestDetailViewWithPool("only task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if cmd == nil {
		t.Fatal("'+' with no candidates should return a flash command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "No tasks available") {
		t.Errorf("expected 'No tasks available' message, got %q", fm.Text)
	}
}

func TestDetailView_PlusKey_EntersDepAddMode(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("my task")
	pool.AddTask(task)
	other := core.NewTask("other task")
	pool.AddTask(other)

	dv := NewDetailView(task, nil, nil, pool)
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if cmd != nil {
		t.Error("'+' with candidates should not return a command (enters mode)")
	}
	if dv.mode != DetailModeDepAdd {
		t.Errorf("expected DetailModeDepAdd, got %d", dv.mode)
	}
}

func TestDetailView_DepAdd_EnterAddsDependency(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("my task")
	pool.AddTask(task)
	other := core.NewTask("dependency task")
	pool.AddTask(other)

	dv := NewDetailView(task, nil, nil, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return a command")
	}
	msg := cmd()
	dam, ok := msg.(DependencyAddedMsg)
	if !ok {
		t.Fatalf("expected DependencyAddedMsg, got %T", msg)
	}
	if dam.Task.ID != task.ID {
		t.Errorf("expected task ID %q, got %q", task.ID, dam.Task.ID)
	}
	if len(task.DependsOn) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(task.DependsOn))
	}
	if task.DependsOn[0] != other.ID {
		t.Errorf("expected dependency ID %q, got %q", other.ID, task.DependsOn[0])
	}
	if dv.mode != DetailModeView {
		t.Errorf("expected DetailModeView after adding, got %d", dv.mode)
	}
}

func TestDetailView_DepAdd_CycleRejected(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	taskA := core.NewTask("task A")
	taskB := core.NewTask("task B")
	taskB.DependsOn = []string{taskA.ID}
	pool.AddTask(taskA)
	pool.AddTask(taskB)

	// Try to add B as dependency of A — this would create A->B->A cycle
	dv := NewDetailView(taskA, nil, nil, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})

	// Navigate to taskB in the candidates
	// Find which index taskB is at
	for i, c := range dv.depAddCandidates {
		if c.ID == taskB.ID {
			dv.depAddSelectedIndex = i
			break
		}
	}

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("cycle rejection should return a command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "circular chain") {
		t.Errorf("expected circular chain error, got %q", fm.Text)
	}
	if len(taskA.DependsOn) != 0 {
		t.Error("dependency should not have been added")
	}
}

func TestDetailView_DepAdd_EscCancels(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("my task")
	pool.AddTask(task)
	other := core.NewTask("other")
	pool.AddTask(other)

	dv := NewDetailView(task, nil, nil, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	dv.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if dv.mode != DetailModeView {
		t.Errorf("expected DetailModeView after cancel, got %d", dv.mode)
	}
}

// --- Remove Dependency (-) Key ---

func TestDetailView_MinusKey_NoDeps_ShowsFlash(t *testing.T) {
	t.Parallel()
	dv, _ := newTestDetailViewWithPool("no deps task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	if cmd == nil {
		t.Fatal("'-' with no deps should return a flash command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "No dependencies") {
		t.Errorf("expected 'No dependencies' message, got %q", fm.Text)
	}
}

func TestDetailView_MinusKey_EntersDepBrowse(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	dep := core.NewTask("dep task")
	pool.AddTask(dep)
	task := core.NewTask("my task")
	task.DependsOn = []string{dep.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	if cmd != nil {
		t.Error("'-' should not return a command (enters dep browse mode)")
	}
	if dv.mode != DetailModeDepBrowse {
		t.Errorf("expected DetailModeDepBrowse, got %d", dv.mode)
	}
}

func TestDetailView_DepBrowse_DeleteRemovesDep(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	dep := core.NewTask("dep to remove")
	pool.AddTask(dep)
	task := core.NewTask("my task")
	task.DependsOn = []string{dep.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	if cmd == nil {
		t.Fatal("remove should return a command")
	}
	msg := cmd()
	drm, ok := msg.(DependencyRemovedMsg)
	if !ok {
		t.Fatalf("expected DependencyRemovedMsg, got %T", msg)
	}
	if drm.DependencyID != dep.ID {
		t.Errorf("expected removed dep ID %q, got %q", dep.ID, drm.DependencyID)
	}
	if len(task.DependsOn) != 0 {
		t.Errorf("expected 0 dependencies after removal, got %d", len(task.DependsOn))
	}
	if dv.mode != DetailModeView {
		t.Errorf("should return to view mode after removing last dep, got %d", dv.mode)
	}
}

func TestDetailView_DepBrowse_NavigationKeys(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	dep1 := core.NewTask("dep 1")
	dep2 := core.NewTask("dep 2")
	pool.AddTask(dep1)
	pool.AddTask(dep2)
	task := core.NewTask("my task")
	task.DependsOn = []string{dep1.ID, dep2.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})

	if dv.depBrowseIndex != 0 {
		t.Errorf("initial index should be 0, got %d", dv.depBrowseIndex)
	}

	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if dv.depBrowseIndex != 1 {
		t.Errorf("after j, index should be 1, got %d", dv.depBrowseIndex)
	}

	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if dv.depBrowseIndex != 0 {
		t.Errorf("after k, index should be 0, got %d", dv.depBrowseIndex)
	}
}

func TestDetailView_DepBrowse_EnterNavigates(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	dep := core.NewTask("navigable dep")
	pool.AddTask(dep)
	task := core.NewTask("my task")
	task.DependsOn = []string{dep.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter in dep browse should navigate")
	}
	msg := cmd()
	nav, ok := msg.(NavigateToLinkedMsg)
	if !ok {
		t.Fatalf("expected NavigateToLinkedMsg, got %T", msg)
	}
	if nav.Task.ID != dep.ID {
		t.Errorf("expected navigation to dep %q, got %q", dep.ID, nav.Task.ID)
	}
}

func TestDetailView_DepBrowse_EscReturns(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	dep := core.NewTask("dep")
	pool.AddTask(dep)
	task := core.NewTask("my task")
	task.DependsOn = []string{dep.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	dv.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if dv.mode != DetailModeView {
		t.Errorf("expected DetailModeView after esc, got %d", dv.mode)
	}
}

// --- DepAdd Mode View Rendering ---

func TestDetailView_DepAddMode_ShowsPickerPrompt(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("my task")
	pool.AddTask(task)
	other := core.NewTask("candidate task")
	pool.AddTask(other)

	dv := NewDetailView(task, nil, nil, pool)
	dv.SetWidth(80)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})

	view := dv.View()
	if !strings.Contains(view, "Select task to add as dependency") {
		t.Error("should show dependency picker prompt")
	}
	if !strings.Contains(view, "candidate task") {
		t.Error("should show candidate task in picker")
	}
}

// --- DepBrowse Mode View Rendering ---

func TestDetailView_DepBrowseMode_ShowsHelpText(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	dep := core.NewTask("dep")
	pool.AddTask(dep)
	task := core.NewTask("my task")
	task.DependsOn = []string{dep.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.SetWidth(80)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})

	view := dv.View()
	if !strings.Contains(view, "Navigate") {
		t.Error("should show navigation help in dep browse mode")
	}
	if !strings.Contains(view, "Remove") {
		t.Error("should show remove help in dep browse mode")
	}
}

// --- DepAdd excludes existing deps and self ---

func TestDetailView_DepAddCandidates_ExcludesExistingDeps(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	existingDep := core.NewTask("existing dep")
	pool.AddTask(existingDep)
	newCandidate := core.NewTask("new candidate")
	pool.AddTask(newCandidate)
	task := core.NewTask("my task")
	task.DependsOn = []string{existingDep.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	candidates := dv.buildDepAddCandidates()

	for _, c := range candidates {
		if c.ID == existingDep.ID {
			t.Error("should not include existing dependency in candidates")
		}
		if c.ID == task.ID {
			t.Error("should not include self in candidates")
		}
	}
	found := false
	for _, c := range candidates {
		if c.ID == newCandidate.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("should include non-dep tasks as candidates")
	}
}

// --- Multiple dep removal adjusts index ---

func TestDetailView_DepBrowse_RemoveAdjustsIndex(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	dep1 := core.NewTask("dep 1")
	dep2 := core.NewTask("dep 2")
	pool.AddTask(dep1)
	pool.AddTask(dep2)
	task := core.NewTask("my task")
	task.DependsOn = []string{dep1.ID, dep2.ID}
	pool.AddTask(task)

	dv := NewDetailView(task, nil, nil, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})

	// Navigate to second dep
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if dv.depBrowseIndex != 1 {
		t.Fatalf("expected index 1, got %d", dv.depBrowseIndex)
	}

	// Remove second dep — index should adjust
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})

	if dv.depBrowseIndex != 0 {
		t.Errorf("after removing last item, index should adjust to 0, got %d", dv.depBrowseIndex)
	}
	if len(task.DependsOn) != 1 {
		t.Errorf("expected 1 dep remaining, got %d", len(task.DependsOn))
	}
}
