package textfile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"gopkg.in/yaml.v3"
)

// setupBenchDir creates a temp directory and configures core.SetHomeDir for benchmark use.
// Returns a cleanup function to restore state.
func setupBenchDir(b *testing.B) string {
	b.Helper()
	tempDir := b.TempDir()
	core.SetHomeDir(tempDir)
	b.Cleanup(func() { core.SetHomeDir("") })
	return tempDir
}

// writeBenchTasks writes n tasks to the YAML file for benchmarks.
func writeBenchTasks(b *testing.B, tempDir string, n int) {
	b.Helper()
	configPath := filepath.Join(tempDir, ".threedoors")
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		b.Fatalf("failed to create config dir: %v", err)
	}

	now := time.Now().UTC()
	tasks := make([]*core.Task, n)
	for i := range n {
		tasks[i] = &core.Task{
			ID:        fmt.Sprintf("bench-%d", i),
			Text:      fmt.Sprintf("Benchmark task number %d with some text", i),
			Status:    core.StatusTodo,
			Notes:     []core.TaskNote{},
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	tf := TasksFile{Tasks: tasks}
	data, err := yaml.Marshal(&tf)
	if err != nil {
		b.Fatalf("failed to marshal tasks: %v", err)
	}

	yamlPath := filepath.Join(configPath, "tasks.yaml")
	if err := os.WriteFile(yamlPath, data, 0o644); err != nil {
		b.Fatalf("failed to write tasks file: %v", err)
	}
}

func BenchmarkLoadTasks(b *testing.B) {
	sizes := []int{10, 100, 500}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("tasks=%d", n), func(b *testing.B) {
			tempDir := setupBenchDir(b)
			writeBenchTasks(b, tempDir, n)
			b.ResetTimer()
			for range b.N {
				_, err := LoadTasks()
				if err != nil {
					b.Fatalf("LoadTasks failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkSaveTasks(b *testing.B) {
	sizes := []int{10, 100, 500}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("tasks=%d", n), func(b *testing.B) {
			tempDir := setupBenchDir(b)
			// Ensure config dir exists
			configPath := filepath.Join(tempDir, ".threedoors")
			if err := os.MkdirAll(configPath, 0o755); err != nil {
				b.Fatalf("failed to create config dir: %v", err)
			}

			now := time.Now().UTC()
			tasks := make([]*core.Task, n)
			for i := range n {
				tasks[i] = &core.Task{
					ID:        fmt.Sprintf("bench-%d", i),
					Text:      fmt.Sprintf("Benchmark task number %d with some text", i),
					Status:    core.StatusTodo,
					Notes:     []core.TaskNote{},
					CreatedAt: now,
					UpdatedAt: now,
				}
			}

			b.ResetTimer()
			for range b.N {
				if err := SaveTasks(tasks); err != nil {
					b.Fatalf("SaveTasks failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkProviderSaveTask(b *testing.B) {
	sizes := []int{10, 100, 500}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("pool=%d", n), func(b *testing.B) {
			tempDir := setupBenchDir(b)
			writeBenchTasks(b, tempDir, n)

			provider := NewTextFileProvider()
			updateTask := &core.Task{
				ID:        "bench-0",
				Text:      "Updated benchmark task 0",
				Status:    core.StatusInProgress,
				Notes:     []core.TaskNote{},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			b.ResetTimer()
			for range b.N {
				if err := provider.SaveTask(updateTask); err != nil {
					b.Fatalf("SaveTask failed: %v", err)
				}
			}
		})
	}
}

// TestAdapterReadWriteNFR13 validates the <100ms NFR for adapter operations.
func TestAdapterReadWriteNFR13(t *testing.T) {
	tests := []struct {
		name   string
		nTasks int
	}{
		{"small (10 tasks)", 10},
		{"medium (100 tasks)", 100},
		{"large (500 tasks)", 500},
	}

	for _, tt := range tests {
		t.Run("read/"+tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			core.SetHomeDir(tempDir)
			t.Cleanup(func() { core.SetHomeDir("") })

			// Set up data
			configPath := filepath.Join(tempDir, ".threedoors")
			if err := os.MkdirAll(configPath, 0o755); err != nil {
				t.Fatalf("failed to create config dir: %v", err)
			}

			now := time.Now().UTC()
			tasks := make([]*core.Task, tt.nTasks)
			for i := range tt.nTasks {
				tasks[i] = &core.Task{
					ID:        fmt.Sprintf("nfr-%d", i),
					Text:      fmt.Sprintf("NFR task %d", i),
					Status:    core.StatusTodo,
					Notes:     []core.TaskNote{},
					CreatedAt: now,
					UpdatedAt: now,
				}
			}
			tf := TasksFile{Tasks: tasks}
			data, err := yaml.Marshal(&tf)
			if err != nil {
				t.Fatalf("marshal: %v", err)
				return
			}
			if err := os.WriteFile(filepath.Join(configPath, "tasks.yaml"), data, 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}

			start := time.Now()
			for range 100 {
				if _, err := LoadTasks(); err != nil {
					t.Fatalf("LoadTasks: %v", err)
				}
			}
			avgPerOp := time.Since(start) / 100
			if avgPerOp > 100*time.Millisecond {
				t.Errorf("LoadTasks with %d tasks took %v avg (NFR13: <100ms)", tt.nTasks, avgPerOp)
			}
		})

		t.Run("write/"+tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			core.SetHomeDir(tempDir)
			t.Cleanup(func() { core.SetHomeDir("") })

			configPath := filepath.Join(tempDir, ".threedoors")
			if err := os.MkdirAll(configPath, 0o755); err != nil {
				t.Fatalf("failed to create config dir: %v", err)
			}

			now := time.Now().UTC()
			tasks := make([]*core.Task, tt.nTasks)
			for i := range tt.nTasks {
				tasks[i] = &core.Task{
					ID:        fmt.Sprintf("nfr-%d", i),
					Text:      fmt.Sprintf("NFR task %d", i),
					Status:    core.StatusTodo,
					Notes:     []core.TaskNote{},
					CreatedAt: now,
					UpdatedAt: now,
				}
			}

			start := time.Now()
			for range 100 {
				if err := SaveTasks(tasks); err != nil {
					t.Fatalf("SaveTasks: %v", err)
				}
			}
			avgPerOp := time.Since(start) / 100
			if avgPerOp > 100*time.Millisecond {
				t.Errorf("SaveTasks with %d tasks took %v avg (NFR13: <100ms)", tt.nTasks, avgPerOp)
			}
		})
	}
}
