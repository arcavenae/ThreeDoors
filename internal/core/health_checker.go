package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// textFileIdentifier is a marker interface implemented by text-file-based providers.
type textFileIdentifier interface {
	IsTextFileBackend() bool
}

// isTextFileBackend returns true if the provider is a text-file-based backend.
func isTextFileBackend(p TaskProvider) bool {
	if tf, ok := p.(textFileIdentifier); ok {
		return tf.IsTextFileBackend()
	}
	return false
}

// HealthStatus represents the status of a health check item.
type HealthStatus string

const (
	HealthOK   HealthStatus = "OK"
	HealthFail HealthStatus = "FAIL"
	HealthWarn HealthStatus = "WARN"
)

// HealthCheckItem represents a single health check result.
type HealthCheckItem struct {
	Name       string
	Status     HealthStatus
	Message    string
	Suggestion string // actionable fix, empty if OK
}

// HealthCheckResult contains the results of all health checks.
type HealthCheckResult struct {
	Items    []HealthCheckItem
	Overall  HealthStatus
	Duration time.Duration
}

// HealthChecker performs system health checks.
type HealthChecker struct {
	provider TaskProvider
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker(provider TaskProvider) *HealthChecker {
	return &HealthChecker{provider: provider}
}

// CheckTaskFile verifies that tasks.yaml exists and is readable/writable.
func (hc *HealthChecker) CheckTaskFile() HealthCheckItem {
	item := HealthCheckItem{Name: "Task File"}

	tasksPath, err := GetTasksFilePath()
	if err != nil {
		item.Status = HealthFail
		item.Message = fmt.Sprintf("Cannot determine task file path: %v", err)
		item.Suggestion = "Check ~/.threedoors/ directory permissions"
		return item
	}

	// Check if file exists
	if _, err := os.Stat(tasksPath); err != nil {
		item.Status = HealthFail
		item.Message = "Task file not found"
		item.Suggestion = "Check ~/.threedoors/ directory permissions"
		return item
	}

	// Test writability by creating a temp file in the same directory (mirrors atomic write pattern)
	tmpPath := filepath.Join(filepath.Dir(tasksPath), ".healthcheck.tmp")
	f, err := os.Create(tmpPath)
	if err != nil {
		item.Status = HealthFail
		item.Message = "Task file directory is not writable"
		item.Suggestion = "Check file permissions on ~/.threedoors/tasks.yaml"
		return item
	}
	_ = f.Close()
	_ = os.Remove(tmpPath)

	item.Status = HealthOK
	item.Message = "Task file exists and is writable"
	return item
}

// CheckDatabaseReadWrite verifies task file can be loaded and parsed.
func (hc *HealthChecker) CheckDatabaseReadWrite() HealthCheckItem {
	item := HealthCheckItem{Name: "Database"}

	if hc.provider == nil {
		item.Status = HealthFail
		item.Message = "No provider configured"
		item.Suggestion = "Configure a task provider"
		return item
	}

	tasks, err := hc.provider.LoadTasks()
	if err != nil {
		item.Status = HealthFail
		item.Message = fmt.Sprintf("Failed to load tasks: %v", err)
		item.Suggestion = "Task file may be corrupt. Try backing up and recreating ~/.threedoors/tasks.yaml"
		return item
	}

	item.Status = HealthOK
	item.Message = fmt.Sprintf("%d tasks loaded successfully", len(tasks))
	return item
}

// CheckSyncStatus checks sync state health.
func (hc *HealthChecker) CheckSyncStatus() HealthCheckItem {
	item := HealthCheckItem{Name: "Sync Status"}

	state, err := LoadSyncState()
	if err != nil {
		item.Status = HealthWarn
		item.Message = "Sync state file may be corrupt"
		item.Suggestion = "Sync state file may be corrupt. It will be rebuilt on next sync."
		return item
	}

	if state.LastSyncTime.IsZero() {
		item.Status = HealthWarn
		item.Message = "No sync history found"
		item.Suggestion = "No sync history found. Sync will initialize on next provider connection"
		return item
	}

	elapsed := time.Since(state.LastSyncTime)
	if elapsed >= 24*time.Hour {
		item.Status = HealthWarn
		item.Message = fmt.Sprintf("Last sync: %s ago", formatDuration(elapsed))
		item.Suggestion = "Press S in doors view to trigger a sync"
		return item
	}

	item.Status = HealthOK
	item.Message = fmt.Sprintf("Last sync: %s ago", formatDuration(elapsed))
	return item
}

// CheckAppleNotesAccess checks Apple Notes provider availability.
func (hc *HealthChecker) CheckAppleNotesAccess() HealthCheckItem {
	item := HealthCheckItem{Name: "Apple Notes"}

	if hc.provider == nil {
		item.Status = HealthFail
		item.Message = "No provider configured"
		item.Suggestion = "Configure a task provider"
		return item
	}

	if isTextFileBackend(hc.provider) {
		item.Status = HealthWarn
		item.Message = "Apple Notes not configured - using text file backend"
		return item
	}

	// Try loading to verify connectivity
	_, err := hc.provider.LoadTasks()
	if err != nil {
		item.Status = HealthFail
		item.Message = fmt.Sprintf("Cannot access Apple Notes: %v", err)
		item.Suggestion = "Grant Full Disk Access in System Settings > Privacy & Security"
		return item
	}
	item.Status = HealthOK
	item.Message = "Apple Notes accessible"
	return item
}

// RunAll runs all health checks and returns the combined result.
func (hc *HealthChecker) RunAll() HealthCheckResult {
	start := time.Now()

	// Pre-load tasks once to avoid double provider calls
	var cachedTasks []*Task
	var cachedErr error
	if hc.provider != nil {
		cachedTasks, cachedErr = hc.provider.LoadTasks()
	}

	var items []HealthCheckItem
	items = append(items, hc.CheckTaskFile())
	items = append(items, hc.checkDatabaseWithCache(cachedTasks, cachedErr))
	items = append(items, hc.CheckSyncStatus())
	items = append(items, hc.checkAppleNotesWithCache(cachedErr))

	return HealthCheckResult{
		Items:    items,
		Overall:  computeOverallStatus(items),
		Duration: time.Since(start),
	}
}

// checkDatabaseWithCache checks database using pre-loaded task data.
func (hc *HealthChecker) checkDatabaseWithCache(tasks []*Task, err error) HealthCheckItem {
	item := HealthCheckItem{Name: "Database"}
	if hc.provider == nil {
		item.Status = HealthFail
		item.Message = "No provider configured"
		item.Suggestion = "Configure a task provider"
		return item
	}
	if err != nil {
		item.Status = HealthFail
		item.Message = fmt.Sprintf("Failed to load tasks: %v", err)
		item.Suggestion = "Task file may be corrupt. Try backing up and recreating ~/.threedoors/tasks.yaml"
		return item
	}
	item.Status = HealthOK
	item.Message = fmt.Sprintf("%d tasks loaded successfully", len(tasks))
	return item
}

// checkAppleNotesWithCache checks Apple Notes using cached load result.
func (hc *HealthChecker) checkAppleNotesWithCache(loadErr error) HealthCheckItem {
	item := HealthCheckItem{Name: "Apple Notes"}
	if hc.provider == nil {
		item.Status = HealthFail
		item.Message = "No provider configured"
		item.Suggestion = "Configure a task provider"
		return item
	}
	if isTextFileBackend(hc.provider) {
		item.Status = HealthWarn
		item.Message = "Apple Notes not configured - using text file backend"
		return item
	}
	if loadErr != nil {
		item.Status = HealthFail
		item.Message = fmt.Sprintf("Cannot access Apple Notes: %v", loadErr)
		item.Suggestion = "Grant Full Disk Access in System Settings > Privacy & Security"
		return item
	}
	item.Status = HealthOK
	item.Message = "Apple Notes accessible"
	return item
}

// computeOverallStatus determines the worst status across all items.
func computeOverallStatus(items []HealthCheckItem) HealthStatus {
	overall := HealthOK
	for _, item := range items {
		if item.Status == HealthFail {
			return HealthFail
		}
		if item.Status == HealthWarn && overall == HealthOK {
			overall = HealthWarn
		}
	}
	return overall
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
