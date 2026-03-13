package clickup

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

const (
	providerName  = "clickup"
	cacheFileName = "clickup-cache.yaml"
)

// TaskFetcher abstracts the ClickUp API operations needed by the provider.
type TaskFetcher interface {
	GetTasks(ctx context.Context, listID string, page int) ([]ClickUpTask, error)
	GetAuthorizedUser(ctx context.Context) (*ClickUpUser, error)
	UpdateTaskStatus(ctx context.Context, taskID, status string) error
}

// ClickUpProvider implements core.TaskProvider for ClickUp with bidirectional sync.
// Tasks are fetched from the ClickUp REST API and mapped to ThreeDoors tasks.
// MarkComplete and BlockTask write back to ClickUp via the API.
// Save/Delete return core.ErrReadOnly.
type ClickUpProvider struct {
	client               TaskFetcher
	listIDs              []string
	statusMapping        map[string]core.TaskStatus
	reverseStatusMapping map[core.TaskStatus]string
	pollInterval         time.Duration
	cachePath            string
	cb                   *core.CircuitBreaker
}

// NewClickUpProvider creates a ClickUpProvider with the given API client and config.
func NewClickUpProvider(client TaskFetcher, cfg *ClickUpConfig) *ClickUpProvider {
	statusMapping := DefaultStatusMapping
	if len(cfg.StatusMapping) > 0 {
		statusMapping = cfg.StatusMapping
	}

	reverseMapping := DefaultReverseStatusMapping
	if cfg.DoneStatus != "" || cfg.BlockedStatus != "" {
		reverseMapping = make(map[core.TaskStatus]string)
		for k, v := range DefaultReverseStatusMapping {
			reverseMapping[k] = v
		}
		if cfg.DoneStatus != "" {
			reverseMapping[core.StatusComplete] = cfg.DoneStatus
		}
		if cfg.BlockedStatus != "" {
			reverseMapping[core.StatusBlocked] = cfg.BlockedStatus
		}
	}

	cbConfig := core.DefaultCircuitBreakerConfig()
	cbConfig.FailureThreshold = 3 // AC7: trips after 3 consecutive failures

	return &ClickUpProvider{
		client:               client,
		listIDs:              cfg.ListIDs,
		statusMapping:        statusMapping,
		reverseStatusMapping: reverseMapping,
		pollInterval:         cfg.PollInterval,
		cb:                   core.NewCircuitBreaker(cbConfig),
	}
}

// SetCachePath sets the directory for the local task cache file.
func (p *ClickUpProvider) SetCachePath(configDir string) {
	p.cachePath = filepath.Join(configDir, cacheFileName)
}

// Name returns the provider identifier.
func (p *ClickUpProvider) Name() string {
	return providerName
}

// LoadTasks fetches tasks from configured ClickUp lists and maps them to ThreeDoors tasks.
// Tasks without a name are skipped with a warning log.
// On API failure, falls back to local cache if available.
func (p *ClickUpProvider) LoadTasks() ([]*core.Task, error) {
	tasks, err := p.loadFromAPI()
	if err == nil {
		p.writeCache(tasks)
		return tasks, nil
	}

	if p.cachePath != "" {
		cached, cacheErr := p.readCache()
		if cacheErr == nil {
			fmt.Fprintf(os.Stderr, "Warning: ClickUp API unavailable (%v), using cached tasks\n", err)
			return cached, nil
		}
	}

	return nil, fmt.Errorf("clickup load tasks: %w", err)
}

// loadFromAPI fetches tasks from the ClickUp API.
func (p *ClickUpProvider) loadFromAPI() ([]*core.Task, error) {
	ctx := context.Background()

	var allTasks []*core.Task

	for _, listID := range p.listIDs {
		apiTasks, err := p.client.GetTasks(ctx, listID, 0)
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", listID, err)
		}
		for i := range apiTasks {
			task, ok := p.mapClickUpTask(&apiTasks[i])
			if !ok {
				continue
			}
			allTasks = append(allTasks, task)
		}
	}

	if allTasks == nil {
		allTasks = []*core.Task{}
	}

	return allTasks, nil
}

// mapClickUpTask converts a ClickUp API task to a ThreeDoors Task.
// Returns false if the task should be skipped (e.g., missing required fields).
func (p *ClickUpProvider) mapClickUpTask(ct *ClickUpTask) (*core.Task, bool) {
	if ct.Name == "" {
		log.Printf("clickup: skipping task %s: missing name", ct.ID)
		return nil, false
	}

	now := time.Now().UTC()

	task := &core.Task{
		ID:             ct.ID,
		Text:           ct.Name,
		Status:         p.mapStatus(ct.Status.Status),
		Effort:         MapPriority(ct.Priority),
		SourceProvider: providerName,
		SourceRefs: []core.SourceRef{
			{
				Provider: providerName,
				NativeID: ct.ID,
			},
		},
		Notes:     []core.TaskNote{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// AC4: description → Context
	if ct.Description != "" {
		task.Context = ct.Description
	}

	// AC8: tags → Context (appended)
	if len(ct.Tags) > 0 {
		tagNames := make([]string, len(ct.Tags))
		for i, tag := range ct.Tags {
			tagNames[i] = tag.Name
		}
		tagStr := strings.Join(tagNames, ", ")
		if task.Context != "" {
			task.Context = task.Context + " | " + tagStr
		} else {
			task.Context = tagStr
		}
	}

	// AC7: due_date (Unix ms) → DueDate (UTC)
	if ct.DueDate != "" {
		if dueDate, err := parseUnixMillis(ct.DueDate); err == nil {
			task.DeferUntil = &dueDate
		}
	}

	// AC6: SourceRef with list ID and URL
	if ct.List.ID != "" {
		task.SourceRefs[0].NativeID = fmt.Sprintf("%s:%s", ct.ID, ct.List.ID)
	}
	if ct.URL != "" {
		task.SourceProvider = fmt.Sprintf("clickup:%s", ct.URL)
	}

	return task, true
}

// mapStatus converts a ClickUp status string to a ThreeDoors TaskStatus.
// Falls back to StatusTodo for unknown statuses.
func (p *ClickUpProvider) mapStatus(clickUpStatus string) core.TaskStatus {
	normalized := strings.ToLower(strings.TrimSpace(clickUpStatus))
	if status, ok := p.statusMapping[normalized]; ok {
		return status
	}
	return core.StatusTodo
}

// MapPriority converts a ClickUp priority to a ThreeDoors TaskEffort.
// ClickUp priorities: 1=Urgent, 2=High, 3=Normal, 4=Low.
// nil priority means no priority set.
func MapPriority(priority *ClickUpPriority) core.TaskEffort {
	if priority == nil {
		return core.EffortQuickWin
	}
	switch priority.ID {
	case "1": // Urgent
		return core.EffortDeepWork
	case "2": // High
		return core.EffortDeepWork
	case "3": // Normal
		return core.EffortMedium
	case "4": // Low
		return core.EffortQuickWin
	default:
		return core.EffortQuickWin
	}
}

// parseUnixMillis parses a Unix millisecond timestamp string to time.Time in UTC.
func parseUnixMillis(ms string) (time.Time, error) {
	millis, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse unix millis %q: %w", ms, err)
	}
	return time.Unix(millis/1000, (millis%1000)*int64(time.Millisecond)).UTC(), nil
}

// SaveTask returns ErrReadOnly; ClickUp provider does not support arbitrary saves.
func (p *ClickUpProvider) SaveTask(_ *core.Task) error {
	return core.ErrReadOnly
}

// SaveTasks returns ErrReadOnly; ClickUp provider does not support arbitrary saves.
func (p *ClickUpProvider) SaveTasks(_ []*core.Task) error {
	return core.ErrReadOnly
}

// DeleteTask returns ErrReadOnly; ClickUp provider does not support deletes.
func (p *ClickUpProvider) DeleteTask(_ string) error {
	return core.ErrReadOnly
}

// MarkComplete updates the task status to the configured "done" status in ClickUp,
// protected by a circuit breaker. When the circuit is open, returns ErrCircuitOpen
// so the WAL layer can queue the operation for retry.
func (p *ClickUpProvider) MarkComplete(taskID string) error {
	doneStatus, ok := p.reverseStatusMapping[core.StatusComplete]
	if !ok {
		doneStatus = "complete"
	}

	return p.cb.Execute(func() error {
		ctx := context.Background()
		return p.client.UpdateTaskStatus(ctx, taskID, doneStatus)
	})
}

// UpdateStatus updates a task's status in ClickUp via the reverse status mapping.
// Protected by the circuit breaker for resilience.
func (p *ClickUpProvider) UpdateStatus(taskID string, status core.TaskStatus) error {
	clickUpStatus, ok := p.reverseStatusMapping[status]
	if !ok {
		return fmt.Errorf("clickup: no reverse mapping for status %q", status)
	}

	return p.cb.Execute(func() error {
		ctx := context.Background()
		return p.client.UpdateTaskStatus(ctx, taskID, clickUpStatus)
	})
}

// CircuitState returns the current state of the circuit breaker.
func (p *ClickUpProvider) CircuitState() core.CircuitState {
	return p.cb.State()
}

// Watch returns nil; ClickUp provider uses poll-based sync.
func (p *ClickUpProvider) Watch() <-chan core.ChangeEvent {
	return nil
}

// HealthCheck verifies API connectivity via GetAuthorizedUser.
func (p *ClickUpProvider) HealthCheck() core.HealthCheckResult {
	start := time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := p.client.GetAuthorizedUser(ctx)
	duration := time.Since(start)

	if err != nil {
		return core.HealthCheckResult{
			Items: []core.HealthCheckItem{
				{
					Name:       "clickup_connectivity",
					Status:     core.HealthFail,
					Message:    fmt.Sprintf("ClickUp API unreachable: %v", err),
					Suggestion: "Check API token and network connectivity",
				},
			},
			Overall:  core.HealthFail,
			Duration: duration,
		}
	}

	return core.HealthCheckResult{
		Items: []core.HealthCheckItem{
			{
				Name:    "clickup_connectivity",
				Status:  core.HealthOK,
				Message: "ClickUp API reachable",
			},
		},
		Overall:  core.HealthOK,
		Duration: duration,
	}
}

// PollInterval returns the configured poll interval for sync.
func (p *ClickUpProvider) PollInterval() time.Duration {
	return p.pollInterval
}

// Factory creates a ClickUpProvider from a ProviderConfig, wrapped with WALProvider
// for offline-first bidirectional sync.
func Factory(config *core.ProviderConfig) (core.TaskProvider, error) {
	settings := findClickUpSettings(config)
	if settings == nil {
		return nil, fmt.Errorf("clickup factory: no clickup provider settings found")
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		return nil, fmt.Errorf("clickup factory: %w", err)
	}

	client := NewClient(AuthConfig{
		APIToken: cfg.APIToken,
	})
	provider := NewClickUpProvider(client, cfg)

	configDir, dirErr := core.GetConfigDirPath()
	if dirErr == nil {
		provider.SetCachePath(configDir)
		return core.NewWALProvider(provider, configDir), nil
	}

	return provider, nil
}

// findClickUpSettings locates the clickup provider entry in the config.
func findClickUpSettings(config *core.ProviderConfig) map[string]string {
	if config == nil {
		return nil
	}
	for _, entry := range config.Providers {
		if entry.Name == providerName {
			return entry.Settings
		}
	}
	return nil
}

// cacheEntry wraps tasks with a timestamp for cache staleness detection.
type cacheEntry struct {
	LastUpdated time.Time    `json:"last_updated"`
	Tasks       []*core.Task `json:"tasks"`
}

// writeCache writes tasks to the local cache file using atomic write.
func (p *ClickUpProvider) writeCache(tasks []*core.Task) {
	if p.cachePath == "" {
		return
	}

	entry := cacheEntry{
		LastUpdated: time.Now().UTC(),
		Tasks:       tasks,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to marshal clickup cache: %v\n", err)
		return
	}

	dir := filepath.Dir(p.cachePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create clickup cache dir: %v\n", err)
		return
	}

	tmpPath := p.cachePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create clickup cache temp file: %v\n", err)
		return
	}

	w := bufio.NewWriter(f)
	if _, err := w.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to write clickup cache: %v\n", err)
		return
	}

	if err := w.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to flush clickup cache: %v\n", err)
		return
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to sync clickup cache: %v\n", err)
		return
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to close clickup cache: %v\n", err)
		return
	}

	if err := os.Rename(tmpPath, p.cachePath); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to rename clickup cache: %v\n", err)
	}
}

// readCache reads tasks from the local cache file.
func (p *ClickUpProvider) readCache() ([]*core.Task, error) {
	data, err := os.ReadFile(p.cachePath)
	if err != nil {
		return nil, fmt.Errorf("read clickup cache: %w", err)
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal clickup cache: %w", err)
	}

	return entry.Tasks, nil
}
