package tasks

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
	// Provider is the legacy flat provider name (backward compatible).
	Provider string `yaml:"provider,omitempty"`
	// NoteTitle is the legacy Apple Notes title (backward compatible).
	NoteTitle string `yaml:"note_title,omitempty"`
	// Providers is the new config-driven list of active providers.
	Providers []ProviderEntry `yaml:"providers,omitempty"`
}

// defaultProviderConfig returns the default configuration.
func defaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		Provider:  "textfile",
		NoteTitle: "ThreeDoors Tasks",
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
		case "obsidian":
			fmt.Fprintf(&b, "#       vault_path: /path/to/your/vault\n")
			fmt.Fprintf(&b, "#       tasks_folder: tasks  # Optional: subfolder within vault\n")
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
