package tui

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestSmoke_AppLaunch(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Buy groceries", "Read book", "Exercise"),
	)

	// Wait for the initial render to contain the app header.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("ThreeDoors"))
	}, teatest.WithDuration(3*time.Second))
}

func TestSmoke_DoorDisplay(t *testing.T) {
	taskTexts := []string{"Alpha task", "Beta task", "Gamma task"}
	tm := NewTestApp(t, WithTasks(taskTexts...))

	// Wait for doors to render — at least one task text should appear.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		for _, text := range taskTexts {
			if bytes.Contains(bts, []byte(text)) {
				return true
			}
		}
		return false
	}, teatest.WithDuration(3*time.Second))
}

func TestSmoke_Quit(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task 1", "Task 2", "Task 3"),
	)

	// Give the app a moment to initialize and render.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("ThreeDoors"))
	}, teatest.WithDuration(3*time.Second))

	// Send quit key.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// The program should exit. Retrieve final model to confirm.
	fm := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second))
	if fm == nil {
		t.Fatal("expected non-nil final model after quit")
		return
	}

	// Verify the final model is our MainModel.
	if _, ok := fm.(*MainModel); !ok {
		t.Errorf("expected *MainModel, got %T", fm)
	}
}

func TestSmoke_DoorSelection(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"select left door with A", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}},
		{"select center door with W", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}},
		{"select right door with D", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTestApp(t,
				WithTasks("Task X", "Task Y", "Task Z"),
			)

			// Wait for initial render.
			teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
				return bytes.Contains(bts, []byte("ThreeDoors"))
			}, teatest.WithDuration(3*time.Second))

			// Select a door.
			tm.Send(tt.key)

			// Give the app time to process the key.
			time.Sleep(200 * time.Millisecond)

			// Quit cleanly and verify the app didn't crash.
			tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
			fm := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second))
			if fm == nil {
				t.Fatal("expected non-nil final model")
				return
			}
		})
	}
}

func TestSmoke_Reroll(t *testing.T) {
	tm := NewTestApp(t,
		WithTasks("Task 1", "Task 2", "Task 3", "Task 4", "Task 5", "Task 6"),
	)

	// Wait for initial render.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("ThreeDoors"))
	}, teatest.WithDuration(3*time.Second))

	// Press S to reroll doors.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	// Give the app time to process.
	time.Sleep(200 * time.Millisecond)

	// Quit cleanly.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	fm := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second))
	if fm == nil {
		t.Fatal("expected non-nil final model after reroll + quit")
		return
	}
}
