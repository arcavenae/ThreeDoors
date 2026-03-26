package testkit

import (
	"errors"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestMockProvider_Name(t *testing.T) {
	t.Parallel()
	m := NewMockProvider()
	if m.Name() != "mock" {
		t.Errorf("Name() = %q, want %q", m.Name(), "mock")
	}
}

func TestMockProvider_SaveAndLoad(t *testing.T) {
	t.Parallel()
	m := NewMockProvider()

	tasks := []*core.Task{NewTask("a", "Alpha"), NewTask("b", "Beta")}
	if err := m.SaveTasks(tasks); err != nil {
		t.Fatalf("SaveTasks() error: %v", err)
	}

	loaded, err := m.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	RequireTaskCount(t, loaded, 2)
}

func TestMockProvider_SaveTask(t *testing.T) {
	t.Parallel()
	m := NewMockProvider()

	task := NewTask("x", "First")
	if err := m.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}
	if m.SaveCount != 1 {
		t.Errorf("SaveCount = %d, want 1", m.SaveCount)
	}

	// Update existing
	task.Text = "Updated"
	if err := m.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() update error: %v", err)
	}

	loaded, _ := m.LoadTasks()
	RequireTaskCount(t, loaded, 1)
	AssertTaskText(t, loaded, "x", "Updated")
}

func TestMockProvider_DeleteTask(t *testing.T) {
	t.Parallel()
	m := NewMockProviderWithTasks([]*core.Task{
		NewTask("keep", "Keep"),
		NewTask("drop", "Drop"),
	})

	if err := m.DeleteTask("drop"); err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}

	loaded, _ := m.LoadTasks()
	RequireTaskCount(t, loaded, 1)
	AssertTaskExists(t, loaded, "keep")
	AssertTaskAbsent(t, loaded, "drop")
}

func TestMockProvider_DeleteTask_NotFound(t *testing.T) {
	t.Parallel()
	m := NewMockProvider()
	err := m.DeleteTask("ghost")
	if err == nil {
		t.Error("DeleteTask() on non-existent should error")
	}
}

func TestMockProvider_MarkComplete(t *testing.T) {
	t.Parallel()
	m := NewMockProviderWithTasks([]*core.Task{NewTask("c", "Complete me")})

	if err := m.MarkComplete("c"); err != nil {
		t.Fatalf("MarkComplete() error: %v", err)
	}

	loaded, _ := m.LoadTasks()
	AssertTaskStatus(t, loaded, "c", core.StatusComplete)
}

func TestMockProvider_ErrorInjection(t *testing.T) {
	t.Parallel()
	injected := errors.New("injected")

	tests := []struct {
		name    string
		setup   func(*MockProvider)
		trigger func(*MockProvider) error
	}{
		{
			name:    "load error",
			setup:   func(m *MockProvider) { m.ErrLoad = injected },
			trigger: func(m *MockProvider) error { _, err := m.LoadTasks(); return err },
		},
		{
			name:    "save error",
			setup:   func(m *MockProvider) { m.ErrSave = injected },
			trigger: func(m *MockProvider) error { return m.SaveTask(NewTask("x", "X")) },
		},
		{
			name:    "save batch error",
			setup:   func(m *MockProvider) { m.ErrSaveBatch = injected },
			trigger: func(m *MockProvider) error { return m.SaveTasks([]*core.Task{}) },
		},
		{
			name:    "delete error",
			setup:   func(m *MockProvider) { m.ErrDelete = injected },
			trigger: func(m *MockProvider) error { return m.DeleteTask("x") },
		},
		{
			name:    "mark complete error",
			setup:   func(m *MockProvider) { m.ErrMarkComplete = injected },
			trigger: func(m *MockProvider) error { return m.MarkComplete("x") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := NewMockProvider()
			tt.setup(m)
			err := tt.trigger(m)
			if !errors.Is(err, injected) {
				t.Errorf("got err=%v, want injected error", err)
			}
		})
	}
}

func TestMockProvider_Reset(t *testing.T) {
	t.Parallel()
	m := NewMockProviderWithTasks([]*core.Task{NewTask("a", "A")})
	m.ErrLoad = errors.New("boom")
	m.LoadCount = 5

	m.Reset()

	loaded, err := m.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() after reset: %v", err)
	}
	RequireTaskCount(t, loaded, 0)
	if m.LoadCount != 1 {
		t.Errorf("LoadCount after reset = %d, want 1 (from the call above)", m.LoadCount)
	}
}

func TestMockProvider_Watch(t *testing.T) {
	t.Parallel()
	m := NewMockProvider()
	if ch := m.Watch(); ch != nil {
		t.Error("Watch() should return nil")
	}
}

func TestMockProvider_HealthCheck(t *testing.T) {
	t.Parallel()
	m := NewMockProvider()
	_ = m.HealthCheck()
}
