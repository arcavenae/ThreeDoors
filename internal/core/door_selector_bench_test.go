package core

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"
)

// benchPool creates a TaskPool with n categorized tasks for benchmark use.
func benchPool(n int) *TaskPool {
	pool := NewTaskPool()
	types := []TaskType{TypeCreative, TypeAdministrative, TypeTechnical, TypePhysical, ""}
	efforts := []TaskEffort{EffortQuickWin, EffortMedium, EffortDeepWork, ""}
	locations := []TaskLocation{LocationHome, LocationWork, LocationAnywhere, ""}

	for i := range n {
		t := &Task{
			ID:        fmt.Sprintf("bench-%d", i),
			Text:      fmt.Sprintf("Benchmark task %d", i),
			Status:    StatusTodo,
			Type:      types[i%len(types)],
			Effort:    efforts[i%len(efforts)],
			Location:  locations[i%len(locations)],
			Notes:     []TaskNote{},
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		}
		pool.AddTask(t)
	}
	return pool
}

// benchTasks creates a slice of n categorized tasks for benchmark use.
func benchTasks(n int) []*Task {
	types := []TaskType{TypeCreative, TypeAdministrative, TypeTechnical, TypePhysical, ""}
	efforts := []TaskEffort{EffortQuickWin, EffortMedium, EffortDeepWork, ""}
	locations := []TaskLocation{LocationHome, LocationWork, LocationAnywhere, ""}

	tasks := make([]*Task, n)
	for i := range n {
		tasks[i] = &Task{
			ID:        fmt.Sprintf("bench-%d", i),
			Text:      fmt.Sprintf("Benchmark task %d", i),
			Status:    StatusTodo,
			Type:      types[i%len(types)],
			Effort:    efforts[i%len(efforts)],
			Location:  locations[i%len(locations)],
			Notes:     []TaskNote{},
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		}
	}
	return tasks
}

func BenchmarkDiversityScore(b *testing.B) {
	sizes := []int{3, 10, 50, 100}
	for _, n := range sizes {
		tasks := benchTasks(n)
		b.Run(fmt.Sprintf("tasks=%d", n), func(b *testing.B) {
			for range b.N {
				DiversityScore(tasks)
			}
		})
	}
}

func BenchmarkSelectDoors(b *testing.B) {
	sizes := []int{10, 50, 100, 500}
	for _, n := range sizes {
		pool := benchPool(n)
		b.Run(fmt.Sprintf("pool=%d", n), func(b *testing.B) {
			for range b.N {
				// Reset recently shown to keep selection consistent
				pool.recentlyShown = make([]string, 10)
				pool.recentlyShownIdx = 0
				SelectDoors(pool, 3)
			}
		})
	}
}

func BenchmarkSelectDoorsWithRand(b *testing.B) {
	sizes := []int{10, 50, 100, 500}
	for _, n := range sizes {
		pool := benchPool(n)
		rng := rand.New(rand.NewPCG(42, 0))
		b.Run(fmt.Sprintf("pool=%d", n), func(b *testing.B) {
			for range b.N {
				pool.recentlyShown = make([]string, 10)
				pool.recentlyShownIdx = 0
				selectDoorsWithRand(pool, 3, rng, nil)
			}
		})
	}
}

// TestSelectDoorsNFR13 validates the <100ms NFR for door selection.
func TestSelectDoorsNFR13(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
	}{
		{"small pool (10)", 10},
		{"medium pool (100)", 100},
		{"large pool (500)", 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := benchPool(tt.poolSize)
			start := time.Now()
			for range 100 {
				pool.recentlyShown = make([]string, 10)
				pool.recentlyShownIdx = 0
				SelectDoors(pool, 3)
			}
			elapsed := time.Since(start)
			avgPerOp := elapsed / 100
			if avgPerOp > 100*time.Millisecond {
				t.Errorf("SelectDoors with pool=%d took %v avg (NFR13: <100ms)", tt.poolSize, avgPerOp)
			}
		})
	}
}
