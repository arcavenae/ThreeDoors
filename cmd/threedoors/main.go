package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/arcaven/ThreeDoors/internal/adapters/applenotes"
	"github.com/arcaven/ThreeDoors/internal/adapters/github"
	"github.com/arcaven/ThreeDoors/internal/adapters/jira"
	"github.com/arcaven/ThreeDoors/internal/adapters/obsidian"
	"github.com/arcaven/ThreeDoors/internal/adapters/reminders"
	"github.com/arcaven/ThreeDoors/internal/adapters/textfile"
	"github.com/arcaven/ThreeDoors/internal/adapters/todoist"
	"github.com/arcaven/ThreeDoors/internal/cli"
	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/core/connection"
	"github.com/arcaven/ThreeDoors/internal/dist"
	"github.com/arcaven/ThreeDoors/internal/enrichment"
	"github.com/arcaven/ThreeDoors/internal/intelligence"
	"github.com/arcaven/ThreeDoors/internal/mcp"
	"github.com/arcaven/ThreeDoors/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// version is set at build time via -ldflags "-X main.version=<semver>"
var version = "dev"

// channel is set at build time via -ldflags "-X main.channel=alpha" for alpha builds.
// Empty string means stable/default channel.
var channel = ""

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(dist.FormatVersionWithChannel(version, channel))
		os.Exit(0)
	}

	// Route to CLI if the first arg is a known subcommand (except "plan" which uses TUI)
	isPlanMode := len(os.Args) > 1 && os.Args[1] == "plan"
	if len(os.Args) > 1 && isSubcommand(os.Args[1]) && !isPlanMode {
		os.Exit(cli.Execute())
	}

	// Register built-in adapters with the global registry
	registerBuiltinAdapters(core.DefaultRegistry())

	configDir, configErr := core.GetConfigDirPath()
	var cfg *core.ProviderConfig
	if configErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: config dir not found: %v, using defaults\n", configErr)
		cfg = &core.ProviderConfig{Provider: "textfile", NoteTitle: "ThreeDoors Tasks"}
	} else {
		configPath := filepath.Join(configDir, "config.yaml")

		// Generate sample config on first run if none exists
		if err := core.GenerateSampleConfig(configPath, core.DefaultRegistry()); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to generate sample config: %v\n", err)
		}

		var loadErr error
		cfg, loadErr = core.LoadProviderConfig(configPath)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: config load failed: %v, using defaults\n", loadErr)
			cfg = &core.ProviderConfig{Provider: "textfile", NoteTitle: "ThreeDoors Tasks"}
		} else {
			core.WarnCredentialExposure(os.Stderr, configPath, cfg)
		}
	}

	var provider core.TaskProvider
	var resolved *connection.ResolvedConnections
	var eventLog *connection.SyncEventLog

	if len(cfg.Connections) > 0 && configErr == nil {
		// Connection-managed mode: create providers via ConnectionManager
		configPath := filepath.Join(configDir, "config.yaml")
		eventLog = connection.NewSyncEventLog(configDir)
		var resolveErr error
		resolved, resolveErr = connection.ResolveFromConfig(cfg, core.DefaultRegistry(), configPath, eventLog)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: connection init failed: %v, falling back to legacy\n", resolveErr)
		}
	}

	if resolved != nil {
		// Use connection-managed providers via aggregator
		providerMap := make(map[string]core.TaskProvider, len(resolved.Providers))
		var firstProvider string
		for connID, p := range resolved.Providers {
			conn, _ := resolved.Manager.Get(connID)
			if conn != nil {
				// Transition to Connected state: Disconnected → Connecting → Connected
				_ = resolved.Manager.Transition(connID, connection.StateConnecting)
				_ = resolved.Manager.Transition(connID, connection.StateConnected)
			}
			providerMap[p.Name()] = p
			if firstProvider == "" {
				firstProvider = p.Name()
			}
		}
		agg := core.NewMultiSourceAggregatorWithDefault(providerMap, firstProvider)
		provider = agg
	} else if len(cfg.Providers) > 1 {
		// Multi-provider mode: aggregate tasks from all configured providers
		agg, aggErr := core.ResolveAllProviders(cfg, core.DefaultRegistry())
		if aggErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize providers: %v\n", aggErr)
			os.Exit(1)
		}
		provider = agg
	} else {
		// Single-provider mode: backward-compatible path
		baseProvider := core.NewProviderFromConfig(cfg)
		if configErr == nil {
			provider = core.NewWALProvider(baseProvider, configDir)
		} else {
			provider = baseProvider
		}
	}
	loadedTasks, err := provider.LoadTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tasks: %v\n", err)
		os.Exit(1)
	}

	pool := core.NewTaskPool()
	for _, t := range loadedTasks {
		pool.AddTask(t)
	}

	tracker := core.NewSessionTracker()
	hc := core.NewHealthChecker(provider)

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
			analyzer := core.NewPatternAnalyzer()
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

	// Initialize proposal store (optional — non-fatal if it fails)
	var proposalStore *mcp.ProposalStore
	if configErr == nil {
		proposalsPath := filepath.Join(configDir, "proposals.jsonl")
		ps, psErr := mcp.NewProposalStore(proposalsPath, pool)
		if psErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: proposal store init failed: %v\n", psErr)
		} else {
			proposalStore = ps
		}
	}

	isFirstRun := configErr == nil && core.IsFirstRun(configDir)
	model := tui.NewMainModel(pool, tracker, provider, hc, isFirstRun, enrichDB)
	if cfg.ShowKeyHints != nil {
		model.SetShowKeyHints(*cfg.ShowKeyHints)
	}
	if configErr == nil {
		model.SetConfigPath(filepath.Join(configDir, "config.yaml"))
	}
	// Set base theme from config and resolve seasonal override.
	if cfg.Theme != "" {
		model.SetBaseThemeName(cfg.Theme)
	}
	model.SetSeasonalEnabled(cfg.SeasonalThemesEnabled())
	if proposalStore != nil {
		model.SetProposalStore(proposalStore)
	}
	if agentSvc != nil {
		model.SetAgentService(agentSvc)
	}
	if resolved != nil {
		model.SetConnectionManager(resolved.Manager)
	}
	if eventLog != nil {
		model.SetSyncEventLog(eventLog)
	}
	if isPlanMode {
		model.SetPlanningMode(true)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
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
		writer := core.NewMetricsWriter(configDir)
		if writeErr := writer.AppendSession(tracker.Finalize()); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session metrics: %v\n", writeErr)
		}
	}
}

// registerBuiltinAdapters registers the built-in task provider adapters
// with the given registry. This is called during application startup.
func registerBuiltinAdapters(reg *core.Registry) {
	// Text file provider: YAML-based local file storage
	_ = reg.Register("textfile", func(config *core.ProviderConfig) (core.TaskProvider, error) {
		return textfile.NewTextFileProvider(), nil
	})

	// Apple Notes provider: wrapped in FallbackProvider for graceful degradation
	_ = reg.Register("applenotes", func(config *core.ProviderConfig) (core.TaskProvider, error) {
		primary := applenotes.NewAppleNotesProvider(config.NoteTitle)
		fallback := textfile.NewTextFileProvider()
		return core.NewFallbackProvider(primary, fallback), nil
	})

	// Jira provider: reads tasks from Jira via REST API with JQL filtering.
	_ = reg.Register("jira", jira.Factory)

	// GitHub Issues provider: reads issues from GitHub repos via go-github SDK.
	_ = reg.Register("github", github.Factory)

	// Obsidian vault provider: reads/writes Markdown checkbox tasks.
	// Validates vault path on startup; falls back to textfile on failure.
	_ = reg.Register("obsidian", func(config *core.ProviderConfig) (core.TaskProvider, error) {
		vaultPath := ""
		tasksFolder := ""
		filePattern := ""
		dailyNotesEnabled := ""
		dailyNotesFolder := ""
		dailyNotesHeading := ""
		dailyNotesFormat := ""
		for _, p := range config.Providers {
			if p.Name == "obsidian" {
				vaultPath = p.GetSetting("vault_path", "")
				tasksFolder = p.GetSetting("tasks_folder", "")
				filePattern = p.GetSetting("file_pattern", "")
				dailyNotesEnabled = p.GetSetting("daily_notes", "")
				dailyNotesFolder = p.GetSetting("daily_notes_folder", "")
				dailyNotesHeading = p.GetSetting("daily_notes_heading", "")
				dailyNotesFormat = p.GetSetting("daily_notes_format", "")
				break
			}
		}
		if vaultPath == "" {
			return nil, fmt.Errorf("obsidian adapter requires vault_path setting")
		}

		adapter := obsidian.NewObsidianAdapter(vaultPath, tasksFolder, filePattern)

		// Configure daily notes if enabled
		if dailyNotesEnabled == "true" {
			adapter.SetDailyNotes(&obsidian.DailyNotesConfig{
				Enabled:    true,
				Folder:     dailyNotesFolder,
				Heading:    dailyNotesHeading,
				DateFormat: dailyNotesFormat,
			})
		}

		if err := obsidian.ValidateVaultPath(vaultPath); err != nil {
			fallback := textfile.NewTextFileProvider()
			fmt.Fprintf(os.Stderr, "Warning: %v. Falling back to text file provider.\n", err)
			return core.NewFallbackProvider(adapter, fallback), nil
		}

		return adapter, nil
	})

	// Apple Reminders provider: macOS-only via JXA/osascript.
	// On non-macOS platforms the factory returns a descriptive error.
	_ = reg.Register("reminders", reminders.NewFactory())

	// Todoist provider: reads tasks from Todoist via REST API (read-only).
	_ = reg.Register("todoist", todoist.Factory)
}

// isSubcommand checks whether arg is a known CLI subcommand name.
func isSubcommand(arg string) bool {
	for _, name := range cli.KnownSubcommands() {
		if arg == name {
			return true
		}
	}
	return false
}
