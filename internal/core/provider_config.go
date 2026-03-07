package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
	"gopkg.in/yaml.v3"
)

// ProviderEntry represents a configured provider with its per-provider settings.
type ProviderEntry struct {
	Name     string            `yaml:"name"`
	Settings map[string]string `yaml:"settings,omitempty"`
}

// GetSetting returns the value for a key, or the fallback if not found.
func (pe ProviderEntry) GetSetting(key, fallback string) string {
	if pe.Settings == nil {
		return fallback
	}
	v, ok := pe.Settings[key]
	if !ok {
		return fallback
	}
	return v
}

// ProviderConfig holds configuration for which task provider to use.
type ProviderConfig struct {
	// SchemaVersion enables future config migrations. Default: 1.
	SchemaVersion int `yaml:"schema_version,omitempty"`
	// Provider is the legacy flat provider name (backward compatible).
	Provider string `yaml:"provider,omitempty"`
	// NoteTitle is the legacy Apple Notes title (backward compatible).
	NoteTitle string `yaml:"note_title,omitempty"`
	// Providers is the new config-driven list of active providers.
	Providers []ProviderEntry `yaml:"providers,omitempty"`
	// LLM holds LLM backend configuration for task decomposition.
	LLM llm.Config `yaml:"llm,omitempty"`
	// Theme is the door theme name (e.g. "classic", "modern", "scifi", "shoji").
	Theme string `yaml:"theme,omitempty"`
}

// CurrentSchemaVersion is the current config.yaml schema version.
const CurrentSchemaVersion = 1

// defaultProviderConfig returns the default configuration.
func defaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		SchemaVersion: CurrentSchemaVersion,
		Provider:      "textfile",
		NoteTitle:     "ThreeDoors Tasks",
	}
}

// LoadProviderConfig reads provider configuration from a YAML file.
// Returns defaults if the file does not exist.
func LoadProviderConfig(path string) (*ProviderConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultProviderConfig(), nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return defaultProviderConfig(), nil
	}

	cfg := defaultProviderConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Ensure defaults for empty fields (backward compatibility)
	if cfg.Provider == "" && len(cfg.Providers) == 0 {
		cfg.Provider = "textfile"
	}
	if cfg.NoteTitle == "" {
		cfg.NoteTitle = "ThreeDoors Tasks"
	}

	return cfg, nil
}

// ResolveActiveProvider creates a TaskProvider based on the configuration and registry.
// If the config has a providers list, the first provider is used as the primary.
// Otherwise, the legacy flat provider field is used for backward compatibility.
func ResolveActiveProvider(cfg *ProviderConfig, reg *Registry) (TaskProvider, error) {
	if len(cfg.Providers) > 0 {
		return resolveFromProvidersList(cfg, reg)
	}
	return resolveFromFlatConfig(cfg, reg)
}

// resolveFromProvidersList initializes the first configured provider as primary.
func resolveFromProvidersList(cfg *ProviderConfig, reg *Registry) (TaskProvider, error) {
	primary := cfg.Providers[0]

	if !reg.IsRegistered(primary.Name) {
		return nil, fmt.Errorf("resolve provider %q: not registered", primary.Name)
	}

	provider, err := reg.InitProvider(primary.Name, cfg)
	if err != nil {
		return nil, fmt.Errorf("resolve provider %q: %w", primary.Name, err)
	}

	return provider, nil
}

// resolveFromFlatConfig handles the legacy flat provider field.
func resolveFromFlatConfig(cfg *ProviderConfig, reg *Registry) (TaskProvider, error) {
	name := cfg.Provider
	if name == "" {
		name = "textfile"
	}

	if !reg.IsRegistered(name) {
		return nil, fmt.Errorf("resolve provider %q: not registered", name)
	}

	provider, err := reg.InitProvider(name, cfg)
	if err != nil {
		return nil, fmt.Errorf("resolve provider %q: %w", name, err)
	}

	return provider, nil
}

// ResolveAllProviders initializes all configured providers and returns a
// MultiSourceAggregator that merges their tasks. If only one provider is
// configured, it still wraps it in the aggregator for consistent behavior.
// The first configured provider is used as the default for tasks with unknown origin.
func ResolveAllProviders(cfg *ProviderConfig, reg *Registry) (*MultiSourceAggregator, error) {
	entries := cfg.Providers
	if len(entries) == 0 {
		name := cfg.Provider
		if name == "" {
			name = "textfile"
		}
		entries = []ProviderEntry{{Name: name}}
	}

	providers := make(map[string]TaskProvider)
	var firstProvider string
	for _, entry := range entries {
		if !reg.IsRegistered(entry.Name) {
			log.Printf("Warning: provider %q not registered, skipping", entry.Name)
			continue
		}
		provider, err := reg.InitProvider(entry.Name, cfg)
		if err != nil {
			log.Printf("Warning: provider %q failed to initialize: %v", entry.Name, err)
			continue
		}
		providers[entry.Name] = provider
		if firstProvider == "" {
			firstProvider = entry.Name
		}
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers could be initialized")
	}

	return NewMultiSourceAggregatorWithDefault(providers, firstProvider), nil
}

// SaveProviderConfig persists provider configuration to a YAML file using atomic write.
func SaveProviderConfig(path string, cfg *ProviderConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal provider config: %w", err)
	}

	tmpPath := path + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp config file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write temp config file: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync temp config file: %w", err)
	}
	_ = f.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp config file: %w", err)
	}
	return nil
}

// GenerateSampleConfig writes a sample config.yaml with commented-out provider examples.
// If the file already exists, it does nothing (preserves user config).
func GenerateSampleConfig(path string, reg *Registry) error {
	// Do not overwrite existing config
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	providers := reg.ListProviders()

	var b strings.Builder
	fmt.Fprintf(&b, "# ThreeDoors Configuration\n")
	fmt.Fprintf(&b, "# See docs for available providers and settings.\n\n")
	fmt.Fprintf(&b, "schema_version: %d\n\n", CurrentSchemaVersion)
	fmt.Fprintf(&b, "# Active provider (simple mode — use 'providers:' list for advanced config)\n")
	fmt.Fprintf(&b, "provider: textfile\n")
	fmt.Fprintf(&b, "note_title: ThreeDoors Tasks\n\n")
	fmt.Fprintf(&b, "# Advanced: configure multiple providers with per-provider settings\n")
	fmt.Fprintf(&b, "# Uncomment and customize the providers list below:\n")
	fmt.Fprintf(&b, "#\n")
	fmt.Fprintf(&b, "# providers:\n")

	for _, name := range providers {
		fmt.Fprintf(&b, "#   - name: %s\n", name)
		fmt.Fprintf(&b, "#     settings:\n")
		switch name {
		case "textfile":
			fmt.Fprintf(&b, "#       task_file: ~/.threedoors/tasks.yaml\n")
		case "applenotes":
			fmt.Fprintf(&b, "#       note_title: ThreeDoors Tasks\n")
		case "jira":
			fmt.Fprintf(&b, "#       url: https://company.atlassian.net\n")
			fmt.Fprintf(&b, "#       auth_type: basic  # basic or pat\n")
			fmt.Fprintf(&b, "#       email: user@example.com  # Cloud only (basic auth)\n")
			fmt.Fprintf(&b, "#       api_token: \"\"  # Or set JIRA_API_TOKEN env var\n")
			fmt.Fprintf(&b, "#       jql: \"assignee = currentUser() AND statusCategory != Done\"\n")
			fmt.Fprintf(&b, "#       max_results: \"50\"  # Optional\n")
			fmt.Fprintf(&b, "#       poll_interval: 30s  # Optional\n")
		case "obsidian":
			fmt.Fprintf(&b, "#       vault_path: /path/to/your/vault\n")
			fmt.Fprintf(&b, "#       tasks_folder: tasks  # Optional: subfolder within vault\n")
			fmt.Fprintf(&b, "#       file_pattern: \"*.md\"  # Optional: glob pattern for task files\n")
			fmt.Fprintf(&b, "#       daily_notes: true  # Optional: enable daily note integration\n")
			fmt.Fprintf(&b, "#       daily_notes_folder: Daily  # Optional: daily notes folder\n")
			fmt.Fprintf(&b, "#       daily_notes_heading: \"## Tasks\"  # Optional: heading for tasks\n")
			fmt.Fprintf(&b, "#       daily_notes_format: \"2006-01-02.md\"  # Optional: Go date format\n")
		default:
			fmt.Fprintf(&b, "#       # Add provider-specific settings here\n")
		}
	}

	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create sample config: %w", err)
	}

	if _, err := f.WriteString(b.String()); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write sample config: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync sample config: %w", err)
	}
	_ = f.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename sample config: %w", err)
	}

	return nil
}
