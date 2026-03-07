package reminders

import (
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// RemindersConfig holds parsed configuration for the Reminders adapter.
type RemindersConfig struct {
	Lists            []string
	IncludeCompleted bool
}

// ParseRemindersConfig extracts Reminders-specific settings from the
// provider configuration. An empty "lists" setting means all lists.
func ParseRemindersConfig(cfg *core.ProviderConfig) RemindersConfig {
	var rc RemindersConfig

	for _, p := range cfg.Providers {
		if p.Name == "reminders" {
			rc.Lists = parseLists(p.GetSetting("lists", ""))
			rc.IncludeCompleted = parseBool(p.GetSetting("include_completed", "false"))
			return rc
		}
	}

	return rc
}

// parseLists splits a comma-separated list string into trimmed, non-empty names.
func parseLists(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var lists []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			lists = append(lists, p)
		}
	}
	return lists
}

// parseBool returns true if the string is "true" (case-insensitive).
func parseBool(s string) bool {
	return strings.EqualFold(s, "true")
}
