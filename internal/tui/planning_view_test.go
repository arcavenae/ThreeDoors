package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// mockProvider implements core.TaskProvider for testing.
type mockPlanningProvider struct {
	tasks []*core.Task
	saved bool
}

func (p *mockPlanningProvider) LoadTasks() ([]*core.Task, error) { return p.tasks, nil }
func (p *mockPlanningProvider) SaveTask(t *core.Task) error      { return nil }
func (p *mockPlanningProvider) SaveTasks(tasks []*core.Task) error {
	p.saved = true
	return nil
}
func (p *mockPlanningProvider) DeleteTask(id string) error     { return nil }
func (p *mockPlanningProvider) MarkComplete(id string) error   { return nil }
func (p *mockPlanningProvider) Name() string                   { return "mock" }
func (p *mockPlanningProvider) ProviderType() string           { return "mock" }
func (p *mockPlanningProvider) SupportsWrite() bool            { return true }
func (p *mockPlanningProvider) SupportsStatusTransition() bool { return true }
func (p *mockPlanningProvider) SupportedStatusTransitions() map[core.TaskStatus][]core.TaskStatus {
	return nil
}
func (p *mockPlanningProvider) Watch() <-chan core.ChangeEvent { return nil }
func (p *mockPlanningProvider) HealthCheck() core.HealthCheckResult {
	return core.HealthCheckResult{}
}

func newTestPool(tasks ...*core.Task) *core.TaskPool {
	pool := core.NewTaskPool()
	for _, t := range tasks {
		pool.AddTask(t)
	}
	return pool
}

func TestPlanningView_InitialStep(t *testing.T) {
	t.Parallel()
	pool := newTestPool(core.NewTask("task1"))

	pv := NewPlanningView(pool, &mockPlanningProvider{})
	pv.SetWidth(80)
	pv.SetHeight(40)

	// Without guidance file, step should be guidance (first-time)
	// but since configPath may not exist, it defaults to review
	if pv.step != planningStepReview && pv.step != planningStepGuidance {
		t.Errorf("expected review or guidance step, got %d", pv.step)
	}
}

func TestPlanningView_ReviewToSelect(t *testing.T) {
	t.Parallel()
	pool := newTestPool(core.NewTask("task1"))

	pv := NewPlanningView(pool, &mockPlanningProvider{})
	pv.step = planningStepReview
	pv.initReviewStep()
	pv.SetWidth(80)

	// Simulate review complete
	cmd := pv.Update(ReviewCompleteMsg{Reviewed: 1, Continued: 1})
	if cmd != nil {
		t.Error("review-to-select transition should not return a command")
	}
	if pv.step != planningStepSelect {
		t.Errorf("expected select step, got %d", pv.step)
	}
	if pv.selectView == nil {
		t.Error("selectView should be initialized")
	}
}

func TestPlanningView_SelectToConfirm(t *testing.T) {
	t.Parallel()
	pool := newTestPool(core.NewTask("task1"))

	pv := NewPlanningView(pool, &mockPlanningProvider{})
	pv.step = planningStepSelect
	pv.initSelectStep()
	pv.SetWidth(80)
	pv.SetHeight(40)

	// Simulate select complete
	tasks := []*core.Task{core.NewTask("focused")}
	cmd := pv.Update(SelectCompleteMsg{
		FocusTasks:     tasks,
		EnergyLevel:    core.EnergyHigh,
		EnergyOverride: false,
	})
	if cmd == nil {
		t.Error("select-to-confirm transition should return init command")
	}
	if pv.step != planningStepConfirm {
		t.Errorf("expected confirm step, got %d", pv.step)
	}
	if pv.confirmView == nil {
		t.Error("confirmView should be initialized")
	}
}

func TestPlanningView_SelectCancelReturnsToReview(t *testing.T) {
	t.Parallel()
	pool := newTestPool(core.NewTask("task1"))

	pv := NewPlanningView(pool, &mockPlanningProvider{})
	pv.step = planningStepSelect
	pv.initSelectStep()
	pv.SetWidth(80)

	pv.Update(SelectCancelMsg{})
	if pv.step != planningStepReview {
		t.Errorf("expected review step after cancel, got %d", pv.step)
	}
}

func TestPlanningView_ConfirmCancelReturnsToSelect(t *testing.T) {
	t.Parallel()
	pool := newTestPool(core.NewTask("task1"))

	pv := NewPlanningView(pool, &mockPlanningProvider{})
	pv.step = planningStepConfirm
	pv.SetWidth(80)

	pv.Update(ConfirmCancelMsg{})
	if pv.step != planningStepSelect {
		t.Errorf("expected select step after cancel, got %d", pv.step)
	}
}

func TestPlanningView_ConfirmCompleteFinishes(t *testing.T) {
	t.Parallel()
	task := core.NewTask("task1")
	pool := newTestPool(task)
	provider := &mockPlanningProvider{}

	pv := NewPlanningView(pool, provider)
	pv.step = planningStepConfirm
	pv.SetWidth(80)

	cmd := pv.Update(ConfirmCompleteMsg{
		FocusTasks:     []*core.Task{task},
		EnergyLevel:    core.EnergyMedium,
		EnergyOverride: true,
	})
	if cmd == nil {
		t.Fatal("confirm complete should return a command")
	}

	msg := cmd()
	pcm, ok := msg.(PlanningCompleteMsg)
	if !ok {
		t.Fatalf("expected PlanningCompleteMsg, got %T", msg)
	}
	if len(pcm.FocusTasks) != 1 {
		t.Errorf("expected 1 focus task, got %d", len(pcm.FocusTasks))
	}
	if pcm.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestPlanningView_GuidanceDismiss(t *testing.T) {
	t.Parallel()
	pool := newTestPool(core.NewTask("task1"))

	pv := NewPlanningView(pool, &mockPlanningProvider{})
	pv.step = planningStepGuidance
	pv.showGuidance = true
	pv.SetWidth(80)

	cmd := pv.Update(planningGuidanceDismissMsg{})
	if cmd == nil {
		t.Error("guidance dismiss should return review init command")
	}
	if pv.step != planningStepReview {
		t.Errorf("expected review step after guidance, got %d", pv.step)
	}
}

func TestPlanningView_GuidanceAnyKeyDismisses(t *testing.T) {
	t.Parallel()
	pool := newTestPool(core.NewTask("task1"))

	pv := NewPlanningView(pool, &mockPlanningProvider{})
	pv.step = planningStepGuidance
	pv.SetWidth(80)

	cmd := pv.updateGuidance(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd == nil {
		t.Fatal("any key should return dismiss command")
	}
	msg := cmd()
	if _, ok := msg.(planningGuidanceDismissMsg); !ok {
		t.Fatalf("expected planningGuidanceDismissMsg, got %T", msg)
	}
}

func TestPlanningView_GuidanceEscCancels(t *testing.T) {
	t.Parallel()
	pool := newTestPool(core.NewTask("task1"))

	pv := NewPlanningView(pool, &mockPlanningProvider{})
	pv.step = planningStepGuidance
	pv.SetWidth(80)

	cmd := pv.updateGuidance(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should return cancel command")
	}
	msg := cmd()
	if _, ok := msg.(PlanningCancelledMsg); !ok {
		t.Fatalf("expected PlanningCancelledMsg, got %T", msg)
	}
}

func TestPlanningView_ViewRendersCorrectStep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		step     planningStep
		contains string
	}{
		{"guidance", planningStepGuidance, "Welcome to Daily Planning"},
		{"review", planningStepReview, "Review"},
		{"select", planningStepSelect, "Select"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := newTestPool(core.NewTask("task1"))
			pv := NewPlanningView(pool, &mockPlanningProvider{})
			pv.SetWidth(80)
			pv.SetHeight(40)
			pv.step = tt.step

			switch tt.step {
			case planningStepReview:
				pv.initReviewStep()
			case planningStepSelect:
				pv.initSelectStep()
			}

			view := pv.View()
			if !strings.Contains(view, tt.contains) {
				t.Errorf("view should contain %q for step %d", tt.contains, tt.step)
			}
		})
	}
}

func TestPlanningView_FinishClearsFocusTags(t *testing.T) {
	t.Parallel()

	oldTask := core.NewTask("old task +focus")
	newTask := core.NewTask("new task")
	pool := newTestPool(oldTask, newTask)
	provider := &mockPlanningProvider{}

	pv := NewPlanningView(pool, provider)
	pv.step = planningStepConfirm
	pv.SetWidth(80)

	cmd := pv.finishSession(ConfirmCompleteMsg{
		FocusTasks:  []*core.Task{newTask},
		EnergyLevel: core.EnergyHigh,
	})
	if cmd == nil {
		t.Fatal("finishSession should return a command")
	}

	// Old task should have focus cleared
	if core.HasFocusTag(oldTask) {
		t.Error("old task should have focus tag cleared")
	}

	// New task should have focus tag applied
	if !core.HasFocusTag(newTask) {
		t.Error("new task should have focus tag applied")
	}

	// Provider should have saved
	if !provider.saved {
		t.Error("provider should have saved tasks")
	}
}

func TestPlanningTimestamp_SaveAndLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)

	savePlanningTimestamp(dir, now)

	loaded := LoadPlanningTimestamp(dir)
	if loaded == nil {
		t.Fatal("should load timestamp")
		return
	}
	if !loaded.Equal(now) {
		t.Errorf("loaded %v != saved %v", loaded, now)
	}
}

func TestPlanningTimestamp_LoadMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	loaded := LoadPlanningTimestamp(dir)
	if loaded != nil {
		t.Error("should return nil for missing timestamp")
	}
}

func TestPlanningTimestamp_LoadInvalid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "planning_timestamp")
	if err := os.WriteFile(path, []byte("not-a-time"), 0o644); err != nil {
		t.Fatal(err)
	}

	loaded := LoadPlanningTimestamp(dir)
	if loaded != nil {
		t.Error("should return nil for invalid timestamp")
	}
}
