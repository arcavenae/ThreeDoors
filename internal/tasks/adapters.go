package tasks

import (
	"fmt"
	"os"
)

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

	// Obsidian vault provider: reads/writes Markdown checkbox tasks.
	// Validates vault path on startup; falls back to textfile on failure.
	_ = reg.Register("obsidian", func(config *ProviderConfig) (TaskProvider, error) {
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

		adapter := NewObsidianAdapter(vaultPath, tasksFolder, filePattern)

		// Configure daily notes if enabled
		if dailyNotesEnabled == "true" {
			adapter.SetDailyNotes(&DailyNotesConfig{
				Enabled:    true,
				Folder:     dailyNotesFolder,
				Heading:    dailyNotesHeading,
				DateFormat: dailyNotesFormat,
			})
		}

		if err := ValidateVaultPath(vaultPath); err != nil {
			fallback := NewTextFileProvider()
			fmt.Fprintf(os.Stderr, "Warning: %v. Falling back to text file provider.\n", err)
			return NewFallbackProvider(adapter, fallback), nil
		}

		return adapter, nil
	})
}
