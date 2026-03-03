package tui

import (
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

// newGoldenDoorsView creates a DoorsView with deterministic state for golden file tests.
// It sets fixed greeting/footer messages and assigns tasks directly to currentDoors
// (bypassing random selection) to ensure reproducible golden file output.
func newGoldenDoorsView(t *testing.T, taskTexts ...string) *DoorsView {
	t.Helper()

	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	pool := core.NewTaskPool()
	var doorTasks []*core.Task
	for _, text := range taskTexts {
		task := core.NewTask(text)
		pool.AddTask(task)
		doorTasks = append(doorTasks, task)
	}

	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	dv.width = 80
	dv.greeting = "Pick one. Start small. That's progress."
	dv.footerMessage = "Three doors. One choice. Zero wrong answers."
	// Set doors directly to avoid random selection order.
	dv.currentDoors = doorTasks
	return dv
}

// newGoldenMainModel creates a MainModel with deterministic state for golden file tests.
func newGoldenMainModel(t *testing.T, taskTexts ...string) *MainModel {
	t.Helper()

	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	core.SetHomeDir(t.TempDir())
	t.Cleanup(func() { core.SetHomeDir("") })

	pool := core.NewTaskPool()
	var doorTasks []*core.Task
	for _, text := range taskTexts {
		task := core.NewTask(text)
		pool.AddTask(task)
		doorTasks = append(doorTasks, task)
	}

	tracker := core.NewSessionTracker()
	provider := &testProvider{}
	model := NewMainModel(pool, tracker, provider, nil, false, nil)

	// Fix random state for deterministic output.
	model.doorsView.greeting = "Pick one. Start small. That's progress."
	model.doorsView.footerMessage = "Three doors. One choice. Zero wrong answers."
	model.doorsView.currentDoors = doorTasks
	model.width = 80
	model.doorsView.width = 80
	return model
}

func TestGolden_MainDoorsView(t *testing.T) {
	dv := newGoldenDoorsView(t, "Buy groceries", "Read Go book", "Exercise for 30 min")
	out := dv.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_EmptyState(t *testing.T) {
	dv := newGoldenDoorsView(t)
	out := dv.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_TooFewTasks_One(t *testing.T) {
	dv := newGoldenDoorsView(t, "Solo task")
	out := dv.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_TooFewTasks_Two(t *testing.T) {
	dv := newGoldenDoorsView(t, "First task", "Second task")
	out := dv.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_DoorSelectionHighlight(t *testing.T) {
	tests := []struct {
		name    string
		doorIdx int
	}{
		{"LeftDoor", 0},
		{"CenterDoor", 1},
		{"RightDoor", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dv := newGoldenDoorsView(t, "Task Alpha", "Task Beta", "Task Gamma")
			dv.selectedDoorIndex = tt.doorIdx
			out := dv.View()
			golden.RequireEqual(t, []byte(out))
		})
	}
}

func TestGolden_StatusBarWithValues(t *testing.T) {
	model := newGoldenMainModel(t, "Write tests", "Review PRs", "Update docs")
	model.valuesConfig = &core.ValuesConfig{
		Values: []string{"Focus", "Quality", "Balance"},
	}
	out := model.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_CompletedSessionCounter(t *testing.T) {
	dv := newGoldenDoorsView(t, "Remaining task A", "Remaining task B", "Remaining task C")
	dv.completedCount = 3
	out := dv.View()
	golden.RequireEqual(t, []byte(out))
}
