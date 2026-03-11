package cli

import (
	"fmt"
	"path/filepath"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/core/connection"
)

// cliContext holds the initialized provider and task pool for CLI commands.
type cliContext struct {
	provider core.TaskProvider
	pool     *core.TaskPool
	resolved *connection.ResolvedConnections
}

// bootstrap loads configuration and initializes the provider and task pool.
func bootstrap() (*cliContext, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return nil, fmt.Errorf("config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	var provider core.TaskProvider
	if len(cfg.Providers) > 1 {
		agg, aggErr := core.ResolveAllProviders(cfg, core.DefaultRegistry())
		if aggErr != nil {
			return nil, fmt.Errorf("init providers: %w", aggErr)
		}
		provider = agg
	} else {
		baseProvider := core.NewProviderFromConfig(cfg)
		if baseProvider == nil {
			return nil, fmt.Errorf("failed to initialize provider")
		}
		provider = core.NewWALProvider(baseProvider, configDir)
	}

	tasks, err := provider.LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("load tasks: %w", err)
	}

	pool := core.NewTaskPool()
	for _, t := range tasks {
		pool.AddTask(t)
	}

	// Resolve named connections (if any exist in config).
	resolved, _ := connection.ResolveFromConfig(cfg, core.DefaultRegistry(), configPath, nil)

	return &cliContext{provider: provider, pool: pool, resolved: resolved}, nil
}
