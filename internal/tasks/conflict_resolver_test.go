package tasks

import (
	"testing"
	"time"
)

func TestConflictSet_NewConflictSet(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	conflicts := []Conflict{
		{
			LocalTask:  &Task{ID: "1", Text: "local version", UpdatedAt: now},
			RemoteTask: &Task{ID: "1", Text: "remote version", UpdatedAt: now},
		},
	}

	cs := NewConflictSet("TestProvider", conflicts)
	if cs.Provider != "TestProvider" {
		t.Errorf("Provider = %q, want %q", cs.Provider, "TestProvider")
	}
	if len(cs.Conflicts) != 1 {
		t.Fatalf("len(Conflicts) = %d, want 1", len(cs.Conflicts))
	}
	if cs.Current != 0 {
		t.Errorf("Current = %d, want 0", cs.Current)
	}
}

func TestConflictSet_CurrentConflict(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	cs := NewConflictSet("Test", []Conflict{
		{
			LocalTask:  &Task{ID: "1", Text: "local", UpdatedAt: now},
			RemoteTask: &Task{ID: "1", Text: "remote", UpdatedAt: now},
		},
	})

	ic := cs.CurrentConflict()
	if ic == nil {
		t.Fatal("CurrentConflict returned nil")
	}
	if ic.Conflict.LocalTask.Text != "local" {
		t.Errorf("LocalTask.Text = %q, want %q", ic.Conflict.LocalTask.Text, "local")
	}
}

func TestConflictSet_CurrentConflict_Empty(t *testing.T) {
	t.Parallel()
	cs := NewConflictSet("Test", nil)
	if ic := cs.CurrentConflict(); ic != nil {
		t.Errorf("expected nil for empty set, got %+v", ic)
	}
}

func TestConflictSet_ResolveKeepLocal(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	local := &Task{ID: "1", Text: "local text", Status: StatusInProgress, UpdatedAt: now}
	remote := &Task{ID: "1", Text: "remote text", Status: StatusTodo, UpdatedAt: now}

	cs := NewConflictSet("Test", []Conflict{{LocalTask: local, RemoteTask: remote}})

	if err := cs.Resolve(ChoiceKeepLocal); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if !cs.AllResolved() {
		t.Error("expected AllResolved to be true")
	}

	resolutions := cs.Resolutions()
	if len(resolutions) != 1 {
		t.Fatalf("len(Resolutions) = %d, want 1", len(resolutions))
	}
	if resolutions[0].Winner != "local" {
		t.Errorf("Winner = %q, want %q", resolutions[0].Winner, "local")
	}
	if resolutions[0].WinningTask != local {
		t.Error("WinningTask should be the local task")
	}
	if resolutions[0].LocalOverridden {
		t.Error("LocalOverridden should be false for keep local")
	}
}

func TestConflictSet_ResolveKeepRemote(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	local := &Task{ID: "1", Text: "local text", UpdatedAt: now}
	remote := &Task{ID: "1", Text: "remote text", UpdatedAt: now}

	cs := NewConflictSet("Test", []Conflict{{LocalTask: local, RemoteTask: remote}})

	if err := cs.Resolve(ChoiceKeepRemote); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	resolutions := cs.Resolutions()
	if resolutions[0].Winner != "remote" {
		t.Errorf("Winner = %q, want %q", resolutions[0].Winner, "remote")
	}
	if resolutions[0].WinningTask != remote {
		t.Error("WinningTask should be the remote task")
	}
	if !resolutions[0].LocalOverridden {
		t.Error("LocalOverridden should be true for keep remote")
	}
}

func TestConflictSet_ResolveKeepBoth(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	local := &Task{ID: "1", Text: "local text", UpdatedAt: now}
	remote := &Task{ID: "1", Text: "remote text", UpdatedAt: now}

	cs := NewConflictSet("Test", []Conflict{{LocalTask: local, RemoteTask: remote}})

	if err := cs.Resolve(ChoiceKeepBoth); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	resolutions := cs.Resolutions()
	if resolutions[0].Winner != "both" {
		t.Errorf("Winner = %q, want %q", resolutions[0].Winner, "both")
	}
	if resolutions[0].LocalOverridden {
		t.Error("LocalOverridden should be false for keep both")
	}
}

func TestConflictSet_MultipleConflicts(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	conflicts := []Conflict{
		{
			LocalTask:  &Task{ID: "1", Text: "first local", UpdatedAt: now},
			RemoteTask: &Task{ID: "1", Text: "first remote", UpdatedAt: now},
		},
		{
			LocalTask:  &Task{ID: "2", Text: "second local", UpdatedAt: now},
			RemoteTask: &Task{ID: "2", Text: "second remote", UpdatedAt: now},
		},
	}

	cs := NewConflictSet("Test", conflicts)

	if cs.UnresolvedCount() != 2 {
		t.Errorf("UnresolvedCount = %d, want 2", cs.UnresolvedCount())
	}

	if err := cs.Resolve(ChoiceKeepLocal); err != nil {
		t.Fatalf("Resolve first: %v", err)
	}

	if cs.AllResolved() {
		t.Error("should not be AllResolved after first resolution")
	}
	if cs.UnresolvedCount() != 1 {
		t.Errorf("UnresolvedCount = %d, want 1", cs.UnresolvedCount())
	}

	ic := cs.CurrentConflict()
	if ic == nil {
		t.Fatal("CurrentConflict returned nil for second conflict")
	}
	if ic.Conflict.LocalTask.Text != "second local" {
		t.Errorf("second conflict LocalTask.Text = %q, want %q", ic.Conflict.LocalTask.Text, "second local")
	}

	if err := cs.Resolve(ChoiceKeepRemote); err != nil {
		t.Fatalf("Resolve second: %v", err)
	}

	if !cs.AllResolved() {
		t.Error("expected AllResolved after both resolved")
	}

	resolutions := cs.Resolutions()
	if len(resolutions) != 2 {
		t.Fatalf("len(Resolutions) = %d, want 2", len(resolutions))
	}
}

func TestConflictSet_ResolveNoConflict(t *testing.T) {
	t.Parallel()
	cs := NewConflictSet("Test", nil)

	err := cs.Resolve(ChoiceKeepLocal)
	if err == nil {
		t.Error("expected error when resolving empty set")
	}
}

func TestConflictChoice_Constants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		choice ConflictChoice
		want   string
	}{
		{"local", ChoiceKeepLocal, "local"},
		{"remote", ChoiceKeepRemote, "remote"},
		{"both", ChoiceKeepBoth, "both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.choice) != tt.want {
				t.Errorf("ConflictChoice = %q, want %q", tt.choice, tt.want)
			}
		})
	}
}
