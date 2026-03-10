package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Option Generation ---

func TestNextStepsView_CompletionContext_HasDoorOption(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	found := false
	for _, opt := range nv.options {
		if opt.Action == "doors" {
			found = true
			break
		}
	}
	if !found {
		t.Error("completion context should have 'doors' option")
	}
}

func TestNextStepsView_CompletionContext_HasAddOption(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	found := false
	for _, opt := range nv.options {
		if opt.Action == "add" {
			found = true
			break
		}
	}
	if !found {
		t.Error("completion context should have 'add' option")
	}
}

func TestNextStepsView_CompletionContext_HasMoodOption(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	found := false
	for _, opt := range nv.options {
		if opt.Action == "mood" {
			found = true
			break
		}
	}
	if !found {
		t.Error("completion context should have 'mood' option")
	}
}

func TestNextStepsView_CompletionContext_HasStatsOption(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	found := false
	for _, opt := range nv.options {
		if opt.Action == "stats" {
			found = true
			break
		}
	}
	if !found {
		t.Error("completion context should have 'stats' option")
	}
}

func TestNextStepsView_CompletionContext_BlockedTasks_ShowsReviewOption(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	// Set one task as blocked
	for _, task := range pool.GetAllTasks() {
		_ = task.UpdateStatus(core.StatusBlocked)
		break
	}
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	found := false
	for _, opt := range nv.options {
		if opt.Action == "search" && strings.Contains(opt.Label, "blocked") {
			found = true
			break
		}
	}
	if !found {
		t.Error("completion context with blocked tasks should have 'review blocked' option")
	}
}

func TestNextStepsView_CompletionContext_NoBlockedTasks_NoReviewOption(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	for _, opt := range nv.options {
		if strings.Contains(opt.Label, "blocked") {
			t.Error("completion context without blocked tasks should not have 'review blocked' option")
		}
	}
}

func TestNextStepsView_AddedContext_HasBackToDoors(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("added", pool, core.NewCompletionCounter())
	found := false
	for _, opt := range nv.options {
		if opt.Action == "doors" {
			found = true
			break
		}
	}
	if !found {
		t.Error("added context should have 'doors' option")
	}
}

func TestNextStepsView_AddedContext_HasAddAnother(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("added", pool, core.NewCompletionCounter())
	found := false
	for _, opt := range nv.options {
		if opt.Action == "add" {
			found = true
			break
		}
	}
	if !found {
		t.Error("added context should have 'add' option")
	}
}

func TestNextStepsView_AddedContext_HasSearchOption(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("added", pool, core.NewCompletionCounter())
	found := false
	for _, opt := range nv.options {
		if opt.Action == "search" {
			found = true
			break
		}
	}
	if !found {
		t.Error("added context should have 'search' option when pool has tasks")
	}
}

func TestNextStepsView_MinimumOptions(t *testing.T) {
	pool := makePool("task1")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	if len(nv.options) < 3 {
		t.Errorf("next steps should have at least 3 options, got %d", len(nv.options))
	}
}

// --- Key Handling ---

func TestNextStepsView_NumberKeySelects(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	nv.SetWidth(80)

	cmd := nv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if cmd == nil {
		t.Fatal("pressing '1' should return a command")
		return
	}
	msg := cmd()
	nsm, ok := msg.(NextStepSelectedMsg)
	if !ok {
		t.Fatalf("expected NextStepSelectedMsg, got %T", msg)
	}
	if nsm.Action != nv.options[0].Action {
		t.Errorf("expected action %q, got %q", nv.options[0].Action, nsm.Action)
	}
}

func TestNextStepsView_InvalidNumberKey_NoOp(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())

	// Press a number beyond the options count
	key := fmt.Sprintf("%d", len(nv.options)+1)
	cmd := nv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	if cmd != nil {
		t.Error("pressing an out-of-range number should be a no-op")
	}
}

func TestNextStepsView_EscDismisses(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())

	cmd := nv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("Esc should return a command")
		return
	}
	msg := cmd()
	if _, ok := msg.(NextStepDismissedMsg); !ok {
		t.Errorf("expected NextStepDismissedMsg, got %T", msg)
	}
}

// --- View Rendering ---

func TestNextStepsView_RendersHeader(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	nv.SetWidth(80)
	view := nv.View()
	if !strings.Contains(view, "next") && !strings.Contains(view, "Nice") && !strings.Contains(view, "fire") && !strings.Contains(view, "momentum") {
		t.Errorf("View should contain a header message, got: %s", view)
	}
}

func TestNextStepsView_RendersOptions(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	nv.SetWidth(80)
	view := nv.View()
	for i, opt := range nv.options {
		expected := fmt.Sprintf("%d. %s", i+1, opt.Label)
		if !strings.Contains(view, expected) {
			t.Errorf("View should contain option %q", expected)
		}
	}
}

func TestNextStepsView_RendersHelpText(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	nv := NewNextStepsView("completed", pool, core.NewCompletionCounter())
	nv.SetWidth(80)
	view := nv.View()
	if !strings.Contains(view, "Esc") {
		t.Error("View should contain Esc in help text")
	}
}

// --- Completion Header Variants ---

func TestPickCompletionHeader_HighCount(t *testing.T) {
	cc := core.NewCompletionCounter()
	for i := 0; i < 5; i++ {
		cc.IncrementToday()
	}
	header := pickCompletionHeader(cc)
	if !strings.Contains(header, "fire") {
		t.Errorf("expected 'fire' header for 5+ completions, got %q", header)
	}
}

func TestPickCompletionHeader_MediumCount(t *testing.T) {
	cc := core.NewCompletionCounter()
	for i := 0; i < 3; i++ {
		cc.IncrementToday()
	}
	header := pickCompletionHeader(cc)
	if !strings.Contains(header, "momentum") {
		t.Errorf("expected 'momentum' header for 3+ completions, got %q", header)
	}
}

func TestPickCompletionHeader_LowCount(t *testing.T) {
	cc := core.NewCompletionCounter()
	header := pickCompletionHeader(cc)
	if !strings.Contains(header, "Nice") {
		t.Errorf("expected 'Nice' header for low count, got %q", header)
	}
}

func TestPickCompletionHeader_NilCounter(t *testing.T) {
	header := pickCompletionHeader(nil)
	if !strings.Contains(header, "Nice") {
		t.Errorf("expected 'Nice' header for nil counter, got %q", header)
	}
}
