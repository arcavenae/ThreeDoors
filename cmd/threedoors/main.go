package main

import (
	"fmt"
	"os"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/arcaven/ThreeDoors/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	loadedTasks, err := tasks.LoadTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tasks: %v\n", err)
		os.Exit(1)
	}

	pool := tasks.NewTaskPool()
	for _, t := range loadedTasks {
		pool.AddTask(t)
	}

	tracker := tasks.NewSessionTracker()
	model := tui.NewMainModel(pool, tracker)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Persist session metrics on exit
	configDir, err := tasks.GetConfigDirPath()
	if err == nil {
		writer := tasks.NewMetricsWriter(configDir)
		if writeErr := writer.AppendSession(tracker.Finalize()); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session metrics: %v\n", writeErr)
		}
	}
}
