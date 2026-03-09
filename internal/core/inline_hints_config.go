package core

// InlineHintsConfig holds the inline key hints configuration.
// Simplified in Story 39.13: removed auto-fade (SessionCount, FadeThreshold).
type InlineHintsConfig struct {
	ShowInlineHints bool `yaml:"show_inline_hints"`
}

// DefaultInlineHintsConfig returns the default inline hints configuration.
func DefaultInlineHintsConfig() *InlineHintsConfig {
	return &InlineHintsConfig{
		ShowInlineHints: true,
	}
}

// ResolveInlineHintState determines whether inline hints should be shown.
// Returns (enabled, fade). Fade is always false after Story 39.13 removed auto-fade.
func ResolveInlineHintState(cfg *InlineHintsConfig) (enabled bool, fade bool) {
	return cfg.ShowInlineHints, false
}
