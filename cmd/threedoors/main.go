package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/arcaven/ThreeDoors/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	configDir, _ := tasks.GetConfigDirPath()
	cfg, err := tasks.LoadProviderConfig(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: config load failed: %v, using defaults\n", err)
		cfg = &tasks.ProviderConfig{Provider: "textfile"}
	}

	provider := tasks.NewProviderFromConfig(cfg)
	loadedTasks, err := provider.LoadTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tasks: %v\n", err)
		os.Exit(1)
	}

	pool := tasks.NewTaskPool()
	for _, t := range loadedTasks {
		pool.AddTask(t)
	}

	tracker := tasks.NewSessionTracker()
	model := tui.NewMainModel(pool, tracker, provider)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Persist session metrics on exit
	configDir, err = tasks.GetConfigDirPath()
	if err == nil {
		writer := tasks.NewMetricsWriter(configDir)
		if writeErr := writer.AppendSession(tracker.Finalize()); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session metrics: %v\n", writeErr)
		}
	}
}
