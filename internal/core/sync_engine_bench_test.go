package core

import (
	"fmt"
	"testing"
	"time"
)

// benchSyncData creates matched local, remote, and sync state data for sync benchmarks.
// overlap controls how many tasks exist in both local and remote.
// extraRemote controls how many tasks exist only in remote (new tasks).
func benchSyncData(overlap, extraRemote int) (SyncState, []*Task, []*Task) {
	var local, remote []*Task
	snapshots := make(map[string]TaskSnapshot)

	// Overlapping tasks (exist in sync state, local, and remote)
	for i := range overlap {
		id := fmt.Sprintf("sync-%d", i)
		t := &Task{
			ID:        id,
			Text:      fmt.Sprintf("Sync task %d", i),
			Status:    StatusTodo,
			Notes:     []TaskNote{},
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		}
		local = append(local, t)
		remote = append(remote, t)
		snapshots[id] = TaskSnapshot{
			ID:        id,
			Text:      t.Text,
			Status:    t.Status,
			UpdatedAt: baseTime,
			Dirty:     false,
		}
	}

	// Extra remote tasks (new tasks)
	for i := range extraRemote {
		id := fmt.Sprintf("remote-new-%d", i)
		remote = append(remote, &Task{
			ID:        id,
			Text:      fmt.Sprintf("New remote task %d", i),
			Status:    StatusTodo,
			Notes:     []TaskNote{},
			CreatedAt: laterTime,
			UpdatedAt: laterTime,
		})
	}

	syncState := SyncState{
		LastSyncTime:  baseTime,
		TaskSnapshots: snapshots,
	}
	return syncState, local, remote
}

func BenchmarkDetectChanges(b *testing.B) {
	tests := []struct {
		name        string
		overlap     int
		extraRemote int
	}{
		{"small/no-changes", 10, 0},
		{"medium/no-changes", 100, 0},
		{"large/no-changes", 500, 0},
		{"medium/10-new", 100, 10},
		{"large/50-new", 500, 50},
	}

	engine := NewSyncEngine()

	for _, tt := range tests {
		syncState, local, remote := benchSyncData(tt.overlap, tt.extraRemote)
		b.Run(tt.name, func(b *testing.B) {
			for range b.N {
				engine.DetectChanges(syncState, local, remote)
			}
		})
	}
}

func BenchmarkResolveConflicts(b *testing.B) {
	sizes := []int{1, 10, 50}
	engine := NewSyncEngine()

	for _, n := range sizes {
		conflicts := make([]Conflict, n)
		for i := range n {
			conflicts[i] = Conflict{
				LocalTask: &Task{
					ID:        fmt.Sprintf("conflict-%d", i),
					Text:      fmt.Sprintf("Local version %d", i),
					Status:    StatusTodo,
					UpdatedAt: laterTime,
				},
				RemoteTask: &Task{
					ID:        fmt.Sprintf("conflict-%d", i),
					Text:      fmt.Sprintf("Remote version %d", i),
					Status:    StatusInProgress,
					UpdatedAt: latestTime,
				},
			}
		}
		b.Run(fmt.Sprintf("conflicts=%d", n), func(b *testing.B) {
			for range b.N {
				engine.ResolveConflicts(conflicts)
			}
		})
	}
}

func BenchmarkSyncFull(b *testing.B) {
	sizes := []int{10, 100, 500}
	engine := NewSyncEngine()

	for _, n := range sizes {
		tasks := benchTasks(n)
		provider := newInMemoryProvider()
		provider.tasks = tasks

		b.Run(fmt.Sprintf("tasks=%d", n), func(b *testing.B) {
			for range b.N {
				syncState := newTestSyncState(tasks...)
				pool := poolFromTasks(tasks...)
				_, _ = engine.Sync(provider, syncState, pool)
			}
		})
	}
}

// TestSyncNFR13 validates the <100ms NFR for sync operations.
func TestSyncNFR13(t *testing.T) {
	tests := []struct {
		name   string
		nTasks int
	}{
		{"small sync (10)", 10},
		{"medium sync (100)", 100},
		{"large sync (500)", 500},
	}

	engine := NewSyncEngine()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := benchTasks(tt.nTasks)
			provider := newInMemoryProvider()
			provider.tasks = tasks

			start := time.Now()
			for range 100 {
				syncState := newTestSyncState(tasks...)
				pool := poolFromTasks(tasks...)
				_, err := engine.Sync(provider, syncState, pool)
				if err != nil {
					t.Fatalf("Sync failed: %v", err)
					return
				}
			}
			elapsed := time.Since(start)
			avgPerOp := elapsed / 100
			if avgPerOp > 100*time.Millisecond {
				t.Errorf("Sync with %d tasks took %v avg (NFR13: <100ms)", tt.nTasks, avgPerOp)
			}
		})
	}
}
