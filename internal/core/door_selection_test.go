package core

import (
	"testing"
)

func TestSelectRandomDoors_ReturnsRequestedCount(t *testing.T) {
	allTasks := []Task{
		{Text: "A"},
		{Text: "B"},
		{Text: "C"},
		{Text: "D"},
		{Text: "E"},
	}

	result := SelectRandomDoors(allTasks, 3, nil)
	if len(result) != 3 {
		t.Errorf("SelectRandomDoors() returned %d tasks, want 3", len(result))
	}
}

func TestSelectRandomDoors_ReturnsAllWhenExactCount(t *testing.T) {
	allTasks := []Task{
		{Text: "A"}, {Text: "B"}, {Text: "C"},
	}

	result := SelectRandomDoors(allTasks, 3, nil)
	if len(result) != 3 {
		t.Errorf("SelectRandomDoors() returned %d tasks, want 3", len(result))
	}
}

func TestSelectRandomDoors_ReturnsFewerWhenNotEnough(t *testing.T) {
	allTasks := []Task{
		{Text: "A"}, {Text: "B"},
	}

	result := SelectRandomDoors(allTasks, 3, nil)
	if len(result) != 2 {
		t.Errorf("SelectRandomDoors() returned %d tasks, want 2", len(result))
	}
}

func TestSelectRandomDoors_ReturnsNilForEmptyInput(t *testing.T) {
	result := SelectRandomDoors(nil, 3, nil)
	if result != nil {
		t.Errorf("SelectRandomDoors() returned %v, want nil", result)
	}
}

func TestSelectRandomDoors_ReturnsNilForZeroCount(t *testing.T) {
	allTasks := []Task{{Text: "A"}}
	result := SelectRandomDoors(allTasks, 0, nil)
	if result != nil {
		t.Errorf("SelectRandomDoors() returned %v, want nil", result)
	}
}

func TestSelectRandomDoors_AvoidsExcludedWhenPossible(t *testing.T) {
	allTasks := []Task{
		{Text: "A"},
		{Text: "B"},
		{Text: "C"},
		{Text: "D"},
		{Text: "E"},
	}
	exclude := []Task{{Text: "A"}, {Text: "B"}}

	// Run multiple times to check exclusion behavior
	for range 20 {
		result := SelectRandomDoors(allTasks, 3, exclude)
		if len(result) != 3 {
			t.Fatalf("SelectRandomDoors() returned %d tasks, want 3", len(result))
		}
		for _, r := range result {
			if r.Text == "A" || r.Text == "B" {
				// With 3 non-excluded tasks available, excluded tasks should not appear
				t.Errorf("SelectRandomDoors() included excluded task %q", r.Text)
			}
		}
	}
}

func TestSelectRandomDoors_FallsBackWhenTooFewNonExcluded(t *testing.T) {
	allTasks := []Task{
		{Text: "A"}, {Text: "B"}, {Text: "C"},
	}
	exclude := []Task{{Text: "A"}, {Text: "B"}}

	// Only 1 non-excluded task, but we want 3 — should fall back to all tasks
	result := SelectRandomDoors(allTasks, 3, exclude)
	if len(result) != 3 {
		t.Errorf("SelectRandomDoors() returned %d tasks, want 3", len(result))
	}
}
