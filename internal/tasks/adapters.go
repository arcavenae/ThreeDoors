package tasks

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
}
