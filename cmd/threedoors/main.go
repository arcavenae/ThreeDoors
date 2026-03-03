package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/arcaven/ThreeDoors/internal/dist"
	"github.com/arcaven/ThreeDoors/internal/enrichment"
	"github.com/arcaven/ThreeDoors/internal/intelligence"
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

	// Register built-in adapters with the global registry
	tasks.RegisterBuiltinAdapters(tasks.DefaultRegistry())

	configDir, configErr := tasks.GetConfigDirPath()
	var cfg *tasks.ProviderConfig
	if configErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: config dir not found: %v, using defaults\n", configErr)
		cfg = &tasks.ProviderConfig{Provider: "textfile", NoteTitle: "ThreeDoors Tasks"}
	} else {
		configPath := filepath.Join(configDir, "config.yaml")

		// Generate sample config on first run if none exists
		if err := tasks.GenerateSampleConfig(configPath, tasks.DefaultRegistry()); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to generate sample config: %v\n", err)
		}

		var loadErr error
		cfg, loadErr = tasks.LoadProviderConfig(configPath)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: config load failed: %v, using defaults\n", loadErr)
			cfg = &tasks.ProviderConfig{Provider: "textfile", NoteTitle: "ThreeDoors Tasks"}
		}
	}

	var provider tasks.TaskProvider
	if len(cfg.Providers) > 1 {
		// Multi-provider mode: aggregate tasks from all configured providers
		agg, aggErr := tasks.ResolveAllProviders(cfg, tasks.DefaultRegistry())
		if aggErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize providers: %v\n", aggErr)
			os.Exit(1)
		}
		provider = agg
	} else {
		// Single-provider mode: backward-compatible path
		baseProvider := tasks.NewProviderFromConfig(cfg)
		if configErr == nil {
			provider = tasks.NewWALProvider(baseProvider, configDir)
		} else {
			provider = baseProvider
		}
	}
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

	// Load enrichment database and run pattern analysis in parallel (non-blocking)
	var enrichDB *enrichment.DB
	var enrichWg sync.WaitGroup

	if configErr == nil {
		enrichWg.Add(1)
		go func() {
			defer enrichWg.Done()
			dbPath := filepath.Join(configDir, "enrichment.db")
			edb, err := enrichment.Open(dbPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: enrichment db failed to open: %v\n", err)
				return
			}
			enrichDB = edb
		}()

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

	// Wait for enrichment DB to be ready before creating the model
	enrichWg.Wait()

	// Initialize agent service for LLM task decomposition (optional — non-fatal if config is missing)
	var agentSvc *intelligence.AgentService
	if cfg.LLM.Output.OutputRepo != "" {
		svc, agentErr := intelligence.NewAgentService(cfg.LLM)
		if agentErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: LLM agent service init failed: %v\n", agentErr)
		} else {
			agentSvc = svc
		}
	}

	isFirstRun := configErr == nil && tasks.IsFirstRun(configDir)
	model := tui.NewMainModel(pool, tracker, provider, hc, isFirstRun, enrichDB)
	if agentSvc != nil {
		model.SetAgentService(agentSvc)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Close enrichment database
	if enrichDB != nil {
		if closeErr := enrichDB.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close enrichment db: %v\n", closeErr)
		}
	}

	// Persist session metrics on exit
	if configErr == nil {
		writer := tasks.NewMetricsWriter(configDir)
		if writeErr := writer.AppendSession(tracker.Finalize()); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session metrics: %v\n", writeErr)
		}
	}
}
