package core

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestThreeWayMergeTaskLists_Stress(t *testing.T) {
	t.Parallel()

	// Fixed seed for reproducibility
	rng := rand.New(rand.NewSource(42))

	deviceIDs := []string{"device-alpha", "device-beta", "device-gamma"}
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Create base tasks
	numTasks := 20
	baseTasks := make([]*Task, numTasks)
	for i := range numTasks {
		baseTasks[i] = &Task{
			ID:        fmt.Sprintf("task-%03d", i),
			Text:      fmt.Sprintf("Base task %d", i),
			Status:    StatusTodo,
			Effort:    EffortMedium,
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		}
	}

	// Generate local modifications (device-alpha)
	localTasks := make([]*Task, len(baseTasks))
	for i, bt := range baseTasks {
		cp := *bt
		localTasks[i] = &cp
	}

	// Generate remote modifications (device-beta and device-gamma)
	remoteTasks := make([]*Task, len(baseTasks))
	for i, bt := range baseTasks {
		cp := *bt
		remoteTasks[i] = &cp
	}

	conflictCount := 0
	efforts := []TaskEffort{EffortQuickWin, EffortMedium, EffortDeepWork}

	// Apply 17+ changes per device to ensure 50+ total conflicts
	for _, deviceID := range deviceIDs {
		changesPerDevice := 17 + rng.Intn(3) // 17-19 changes per device
		for j := range changesPerDevice {
			taskIdx := rng.Intn(numTasks)
			fieldChoice := rng.Intn(3) // 0=text, 1=effort, 2=context

			ts := baseTime.Add(time.Duration(j+1) * time.Minute)
			clock := VectorClock{deviceID: uint64(j + 1)}
			version := FieldVersion{
				DeviceID:  deviceID,
				UpdatedAt: ts,
				Version:   uint64(j + 1),
			}

			var target *Task
			if deviceID == "device-alpha" {
				target = localTasks[taskIdx]
			} else {
				target = remoteTasks[taskIdx]
			}

			if target.FieldVersions == nil {
				target.FieldVersions = make(map[string]FieldVersion)
			}
			if target.VectorClock == nil {
				target.VectorClock = make(map[string]uint64)
			}

			switch fieldChoice {
			case 0:
				target.Text = fmt.Sprintf("Modified by %s change %d", deviceID, j)
				target.FieldVersions["text"] = version
			case 1:
				target.Effort = efforts[rng.Intn(len(efforts))]
				target.FieldVersions["effort"] = version
			case 2:
				target.Context = fmt.Sprintf("Context from %s", deviceID)
				target.FieldVersions["context"] = version
			}

			target.UpdatedAt = ts
			VectorClock(target.VectorClock).Merge(clock)
			conflictCount++
		}
	}

	// Also add concurrent creates (new tasks from both sides)
	for i := 0; i < 3; i++ {
		localTasks = append(localTasks, &Task{
			ID:        fmt.Sprintf("local-new-%d", i),
			Text:      fmt.Sprintf("New local task %d", i),
			Status:    StatusTodo,
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		})
		remoteTasks = append(remoteTasks, &Task{
			ID:        fmt.Sprintf("remote-new-%d", i),
			Text:      fmt.Sprintf("New remote task %d", i),
			Status:    StatusTodo,
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		})
	}

	// Run the merge
	outcome := ThreeWayMergeTaskLists(baseTasks, localTasks, remoteTasks)

	// Verify: no panics got us this far

	// Verify: all base tasks + new tasks are in merged result
	// (modify-wins means nothing gets deleted)
	expectedMinTasks := numTasks + 6 // 20 base + 3 local new + 3 remote new
	if len(outcome.MergedTasks) < expectedMinTasks {
		t.Errorf("expected at least %d merged tasks, got %d", expectedMinTasks, len(outcome.MergedTasks))
	}

	// Verify: no data loss — every base task ID exists in merged result
	mergedByID := make(map[string]*Task)
	for _, task := range outcome.MergedTasks {
		if _, exists := mergedByID[task.ID]; exists {
			t.Errorf("duplicate task ID in merged result: %s", task.ID)
		}
		mergedByID[task.ID] = task
	}
	for _, bt := range baseTasks {
		if _, exists := mergedByID[bt.ID]; !exists {
			t.Errorf("base task %s missing from merged result (data loss)", bt.ID)
		}
	}

	// Verify: new tasks from both sides are present
	for i := 0; i < 3; i++ {
		if _, exists := mergedByID[fmt.Sprintf("local-new-%d", i)]; !exists {
			t.Errorf("local new task %d missing", i)
		}
		if _, exists := mergedByID[fmt.Sprintf("remote-new-%d", i)]; !exists {
			t.Errorf("remote new task %d missing", i)
		}
	}

	// Verify: total changes applied exceed 50
	if conflictCount < 50 {
		t.Errorf("expected 50+ conflicting changes, generated only %d", conflictCount)
	}

	t.Logf("Stress test: %d total changes, %d merged tasks, %d conflict records",
		conflictCount, len(outcome.MergedTasks), len(outcome.Conflicts))
}
