package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arcaven/ThreeDoors/internal/dist"
	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/arcaven/ThreeDoors/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// version is set at build time via -ldflags "-X main.version=<semver>"
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(dist.FormatVersion(version))
		os.Exit(0)
	}

	configDir, configErr := tasks.GetConfigDirPath()
	var cfg *tasks.ProviderConfig
	if configErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: config dir not found: %v, using defaults\n", configErr)
		cfg = &tasks.ProviderConfig{Provider: "textfile", NoteTitle: "ThreeDoors Tasks"}
	} else {
		var loadErr error
		cfg, loadErr = tasks.LoadProviderConfig(filepath.Join(configDir, "config.yaml"))
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: config load failed: %v, using defaults\n", loadErr)
			cfg = &tasks.ProviderConfig{Provider: "textfile", NoteTitle: "ThreeDoors Tasks"}
		}
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
	hc := tasks.NewHealthChecker(provider)

	// Run pattern analysis in background (non-blocking)
	if configErr == nil {
		go func() {
			analyzer := tasks.NewPatternAnalyzer()
			sessionsPath := filepath.Join(configDir, "sessions.jsonl")
			patternsPath := filepath.Join(configDir, "patterns.json")

			cached, _ := analyzer.LoadPatterns(patternsPath)
			sessions, err := analyzer.ReadSessions(sessionsPath)
			if err != nil {
				return
			}
			if !analyzer.NeedsReanalysis(cached, sessions) {
				return
			}
			report, err := analyzer.Analyze(sessions)
			if err != nil || report == nil {
				return
			}
			_ = analyzer.SavePatterns(report, patternsPath)
		}()
	}

	model := tui.NewMainModel(pool, tracker, provider, hc)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Persist session metrics on exit
	if configErr == nil {
		writer := tasks.NewMetricsWriter(configDir)
		if writeErr := writer.AppendSession(tracker.Finalize()); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session metrics: %v\n", writeErr)
		}
	}
}
