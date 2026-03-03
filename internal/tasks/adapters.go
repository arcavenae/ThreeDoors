package tasks

import "fmt"

// RegisterBuiltinAdapters registers the built-in task provider adapters
// with the given registry. This should be called during application startup.
func RegisterBuiltinAdapters(reg *Registry) {
	// Text file provider: YAML-based local file storage
	_ = reg.Register("textfile", func(config *ProviderConfig) (TaskProvider, error) {
		return NewTextFileProvider(), nil
	})

	// Apple Notes provider: wrapped in FallbackProvider for graceful degradation
	_ = reg.Register("applenotes", func(config *ProviderConfig) (TaskProvider, error) {
		primary := NewAppleNotesProvider(config.NoteTitle)
		fallback := NewTextFileProvider()
		return NewFallbackProvider(primary, fallback), nil
	})

	// Obsidian vault provider: reads/writes Markdown checkbox tasks
	_ = reg.Register("obsidian", func(config *ProviderConfig) (TaskProvider, error) {
		vaultPath := ""
		tasksFolder := ""
		for _, p := range config.Providers {
			if p.Name == "obsidian" {
				vaultPath = p.GetSetting("vault_path", "")
				tasksFolder = p.GetSetting("tasks_folder", "")
				break
			}
		}
		if vaultPath == "" {
			return nil, fmt.Errorf("obsidian adapter requires vault_path setting")
		}
		return NewObsidianAdapter(vaultPath, tasksFolder), nil
	})
}
