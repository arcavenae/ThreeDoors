package cli

// healthCheckJSON is the JSON representation of a single health check item.
// Retained for backward compatibility with existing test infrastructure.
type healthCheckJSON struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// healthResultJSON is the JSON envelope data for the legacy health command.
// Retained for backward compatibility with existing test infrastructure.
type healthResultJSON struct {
	Overall    string            `json:"overall"`
	DurationMs int64             `json:"duration_ms"`
	Checks     []healthCheckJSON `json:"checks"`
}
