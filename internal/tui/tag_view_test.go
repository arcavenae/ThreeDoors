package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestTagView(t *testing.T) *TagView {
	t.Helper()
	task := core.NewTask("Test task for tags")
	task.Type = core.TypeCreative
	task.Effort = core.EffortMedium
	task.Location = core.LocationHome
	tv := NewTagView(task)
	tv.SetWidth(80)
	return tv
}

func TestNewTagView_NilTask(t *testing.T) {
	t.Parallel()
	tv := NewTagView(nil)
	if tv != nil {
		t.Error("expected nil for nil task")
	}
}

func TestNewTagView_SnapshotsOriginals(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Snapshot test")
	task.Type = core.TypeTechnical
	task.Effort = core.EffortDeepWork
	task.Location = core.LocationWork

	tv := NewTagView(task)
	if tv.origType != core.TypeTechnical {
		t.Errorf("expected origType %q, got %q", core.TypeTechnical, tv.origType)
	}
	if tv.origEffort != core.EffortDeepWork {
		t.Errorf("expected origEffort %q, got %q", core.EffortDeepWork, tv.origEffort)
	}
	if tv.origLocation != core.LocationWork {
		t.Errorf("expected origLocation %q, got %q", core.LocationWork, tv.origLocation)
	}
}

func TestTagView_SetWidth(t *testing.T) {
	t.Parallel()
	tv := newTestTagView(t)
	tv.SetWidth(120)
	if tv.width != 120 {
		t.Errorf("expected width 120, got %d", tv.width)
	}
}

func TestTagView_EscapeRestoresOriginals(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Escape test")
	task.Type = core.TypeCreative
	tv := NewTagView(task)

	// Modify the task type
	task.Type = core.TypeTechnical

	// Press Escape — should restore
	cmd := tv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected non-nil cmd from Escape")
		return
	}

	msg := cmd()
	if _, ok := msg.(TagCancelledMsg); !ok {
		t.Errorf("expected TagCancelledMsg, got %T", msg)
	}

	if task.Type != core.TypeCreative {
		t.Errorf("expected type restored to %q, got %q", core.TypeCreative, task.Type)
	}
}

func TestTagView_NavigateFields(t *testing.T) {
	t.Parallel()
	tv := newTestTagView(t)

	// Initial field index is 0
	if tv.fieldIndex != 0 {
		t.Errorf("expected initial fieldIndex 0, got %d", tv.fieldIndex)
	}

	// Navigate down
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if tv.fieldIndex != 1 {
		t.Errorf("expected fieldIndex 1 after down, got %d", tv.fieldIndex)
	}

	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if tv.fieldIndex != 2 {
		t.Errorf("expected fieldIndex 2 after down, got %d", tv.fieldIndex)
	}

	// Navigate up
	tv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if tv.fieldIndex != 1 {
		t.Errorf("expected fieldIndex 1 after up, got %d", tv.fieldIndex)
	}

	// Can't go below 0
	tv.Update(tea.KeyMsg{Type: tea.KeyUp})
	tv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if tv.fieldIndex != 0 {
		t.Errorf("expected fieldIndex 0 (clamped), got %d", tv.fieldIndex)
	}
}

func TestTagView_SelectDone(t *testing.T) {
	t.Parallel()
	tv := newTestTagView(t)

	// Navigate to "Done" field (index 3)
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})

	cmd := tv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd from Done")
		return
	}

	msg := cmd()
	if tagMsg, ok := msg.(TagUpdatedMsg); !ok {
		t.Errorf("expected TagUpdatedMsg, got %T", msg)
	} else if tagMsg.Task == nil {
		t.Error("expected non-nil task in TagUpdatedMsg")
	}
}

func TestTagView_EnterFieldThenSelectValue(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Value selection")
	tv := NewTagView(task)

	// Enter Type field (index 0)
	tv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if tv.state != tagSelectingValue {
		t.Fatal("expected tagSelectingValue state after enter")
	}

	// Navigate to "Creative" (index 1)
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Select it
	tv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if task.Type != core.TypeCreative {
		t.Errorf("expected type %q, got %q", core.TypeCreative, task.Type)
	}
	if tv.state != tagSelectingField {
		t.Error("expected return to tagSelectingField after value selection")
	}
}

func TestTagView_EffortValueSelection(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Effort test")
	tv := NewTagView(task)

	// Navigate to Effort field (index 1) and enter
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Navigate to "Deep Work" (index 3)
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if task.Effort != core.EffortDeepWork {
		t.Errorf("expected effort %q, got %q", core.EffortDeepWork, task.Effort)
	}
}

func TestTagView_LocationValueSelection(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Location test")
	tv := NewTagView(task)

	// Navigate to Location field (index 2) and enter
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Navigate to "Errands" (index 3)
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if task.Location != core.LocationErrands {
		t.Errorf("expected location %q, got %q", core.LocationErrands, task.Location)
	}
}

func TestTagView_ValueNavigationClamped(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Clamp test")
	tv := NewTagView(task)

	// Enter Type field
	tv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Try navigating past the end
	for i := 0; i < 20; i++ {
		tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if tv.valueIndex != len(typeOptions)-1 {
		t.Errorf("expected valueIndex clamped to %d, got %d", len(typeOptions)-1, tv.valueIndex)
	}

	// Try navigating past the start
	for i := 0; i < 20; i++ {
		tv.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
	if tv.valueIndex != 0 {
		t.Errorf("expected valueIndex clamped to 0, got %d", tv.valueIndex)
	}
}

func TestTagView_CurrentValueIndex(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Current value test")
	task.Effort = core.EffortMedium
	tv := NewTagView(task)

	// Navigate to Effort field and enter
	tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Value index should be pre-set to the current value
	expectedIdx := -1
	for i, opt := range effortOptions {
		if opt.value == core.EffortMedium {
			expectedIdx = i
			break
		}
	}
	if tv.valueIndex != expectedIdx {
		t.Errorf("expected valueIndex %d for EffortMedium, got %d", expectedIdx, tv.valueIndex)
	}
}

func TestTagView_View_FieldSelecting(t *testing.T) {
	t.Parallel()
	tv := newTestTagView(t)

	output := tv.View()
	if !strings.Contains(output, "Select field to edit") {
		t.Error("expected 'Select field to edit' in view")
	}
	if !strings.Contains(output, "Type") {
		t.Error("expected 'Type' field label")
	}
	if !strings.Contains(output, "Effort") {
		t.Error("expected 'Effort' field label")
	}
	if !strings.Contains(output, "Location") {
		t.Error("expected 'Location' field label")
	}
	if !strings.Contains(output, "Done") {
		t.Error("expected 'Done' field label")
	}
	if !strings.Contains(output, ">") {
		t.Error("expected cursor indicator")
	}
}

func TestTagView_View_ValueSelecting(t *testing.T) {
	t.Parallel()
	tv := newTestTagView(t)

	// Enter Type field
	tv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	output := tv.View()
	if !strings.Contains(output, "Select Type") {
		t.Error("expected 'Select Type' in view")
	}
	if !strings.Contains(output, "Creative") {
		t.Error("expected 'Creative' option")
	}
	if !strings.Contains(output, "Technical") {
		t.Error("expected 'Technical' option")
	}
}

func TestTagView_View_ShowsCurrentValues(t *testing.T) {
	t.Parallel()
	task := core.NewTask("View current values")
	task.Type = core.TypeTechnical
	task.Effort = core.EffortQuickWin
	tv := NewTagView(task)
	tv.SetWidth(80)

	output := tv.View()
	if !strings.Contains(output, string(core.TypeTechnical)) {
		t.Error("expected current type value in view")
	}
	if !strings.Contains(output, string(core.EffortQuickWin)) {
		t.Error("expected current effort value in view")
	}
}
