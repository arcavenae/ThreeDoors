package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/arcavenae/ThreeDoors/internal/adapters/applenotes"
	"github.com/arcavenae/ThreeDoors/internal/adapters/clickup"
	"github.com/arcavenae/ThreeDoors/internal/adapters/jira"
	"github.com/arcavenae/ThreeDoors/internal/adapters/linear"
	"github.com/arcavenae/ThreeDoors/internal/adapters/obsidian"
	"github.com/arcavenae/ThreeDoors/internal/adapters/reminders"
	"github.com/arcavenae/ThreeDoors/internal/adapters/textfile"
	"github.com/arcavenae/ThreeDoors/internal/adapters/todoist"
	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/enrichment"
	"github.com/arcavenae/ThreeDoors/internal/mcp"
)

var version = "dev"

func main() {
	transportFlag := flag.String("transport", "stdio", "transport type: stdio or sse")
	portFlag := flag.Int("port", 8080, "port for SSE transport")
	flag.Parse()

	if err := run(*transportFlag, *portFlag); err != nil {
		fmt.Fprintf(os.Stderr, "threedoors-mcp: %v\n", err)
		os.Exit(1)
	}
}

func run(transportType string, port int) error {
	registry := core.NewRegistry()
	registerBuiltinAdapters(registry)

	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return fmt.Errorf("get config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	var provider core.TaskProvider
	var aggregator *core.MultiSourceAggregator

	if len(cfg.Providers) > 1 {
		agg, aggErr := core.ResolveAllProviders(cfg, registry)
		if aggErr != nil {
			return fmt.Errorf("resolve providers: %w", aggErr)
		}
		aggregator = agg
		provider = agg
	} else {
		baseProvider := core.NewProviderFromConfig(cfg)
		if baseProvider == nil {
			return fmt.Errorf("no task provider available: check your configuration in %s", configPath)
		}
		provider = core.NewWALProvider(baseProvider, configDir)
	}

	loadedTasks, err := provider.LoadTasks()
	if err != nil {
		return fmt.Errorf("load tasks: %w", err)
	}

	pool := core.NewTaskPool()
	for _, t := range loadedTasks {
		pool.AddTask(t)
	}

	session := core.NewSessionTracker()

	dbPath := filepath.Join(configDir, "enrichment.db")
	enrichDB, err := enrichment.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: enrichment db: %v\n", err)
	}

	server := mcp.NewMCPServer(registry, aggregator, pool, session, enrichDB, version)

	addr := fmt.Sprintf(":%d", port)
	transport := mcp.TransportFromFlags(transportType, addr, os.Stdin, os.Stdout)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = transport.Serve(ctx, server)

	if enrichDB != nil {
		if closeErr := enrichDB.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close enrichment db: %v\n", closeErr)
		}
	}

	return err
}

func registerBuiltinAdapters(reg *core.Registry) {
	_ = reg.Register("textfile", func(config *core.ProviderConfig) (core.TaskProvider, error) {
		return textfile.NewTextFileProvider(), nil
	})

	_ = reg.Register("applenotes", func(config *core.ProviderConfig) (core.TaskProvider, error) {
		primary := applenotes.NewAppleNotesProvider(config.NoteTitle)
		fallback := textfile.NewTextFileProvider()
		return core.NewFallbackProvider(primary, fallback), nil
	})

	_ = reg.Register("jira", jira.Factory)

	_ = reg.Register("obsidian", func(config *core.ProviderConfig) (core.TaskProvider, error) {
		vaultPath := ""
		for _, p := range config.Providers {
			if p.Name == "obsidian" {
				vaultPath = p.GetSetting("vault_path", "")
				break
			}
		}
		if vaultPath == "" {
			return nil, fmt.Errorf("obsidian adapter requires vault_path setting")
		}
		adapter := obsidian.NewObsidianAdapter(vaultPath, "", "")
		return adapter, nil
	})

	_ = reg.Register("reminders", reminders.NewFactory())

	_ = reg.Register("todoist", todoist.Factory)

	_ = reg.Register("linear", linear.Factory)

	_ = reg.Register("clickup", clickup.Factory)
}
