package linear

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

const (
	providerName  = "linear"
	cacheFileName = "linear-cache.yaml"
	defaultTTL    = 5 * time.Minute
)

// statusTypeMap maps Linear state.type values to ThreeDoors TaskStatus (AC3).
var statusTypeMap = map[string]core.TaskStatus{
	"triage":    core.StatusTodo,
	"backlog":   core.StatusTodo,
	"unstarted": core.StatusTodo,
	"started":   core.StatusInProgress,
	"completed": core.StatusComplete,
	"cancelled": core.StatusArchived,
}

// LinearProvider implements core.TaskProvider for Linear issues.
// SaveTask, SaveTasks, and DeleteTask return core.ErrReadOnly.
// MarkComplete returns core.ErrReadOnly (read-only in this story).
type LinearProvider struct {
	client          GraphQLClient
	config          *LinearConfig
	cachePath       string
	cacheTTL        time.Duration
	lastCacheUpdate time.Time
	cachedTasks     []*core.Task
	watchCh         chan core.ChangeEvent
	stopCh          chan struct{}
}

// NewLinearProvider creates a LinearProvider with the given client and config.
func NewLinearProvider(client GraphQLClient, config *LinearConfig) *LinearProvider {
	return &LinearProvider{
		client:   client,
		config:   config,
		cacheTTL: defaultTTL,
		stopCh:   make(chan struct{}),
	}
}

// SetCachePath sets the directory for the local task cache file.
func (p *LinearProvider) SetCachePath(configDir string) {
	p.cachePath = filepath.Join(configDir, cacheFileName)
}

// Name returns the provider identifier.
func (p *LinearProvider) Name() string {
	return providerName
}

// LoadTasks fetches issues from all configured teams, maps them to tasks,
// and returns a merged result. Uses cache when within TTL.
func (p *LinearProvider) LoadTasks() ([]*core.Task, error) {
	if p.isCacheValid() {
		return p.cachedTasks, nil
	}

	tasks, err := p.loadFromAPI()
	if err == nil {
		p.cachedTasks = tasks
		p.lastCacheUpdate = time.Now().UTC()
		p.writeCache(tasks)
		return tasks, nil
	}

	// API failed — try disk cache fallback
	if p.cachePath != "" {
		cached, cacheErr := p.readCache()
		if cacheErr == nil {
			fmt.Fprintf(os.Stderr, "Warning: Linear API unavailable (%v), using cached tasks\n", err)
			p.cachedTasks = cached
			return cached, nil
		}
	}

	return nil, fmt.Errorf("linear load tasks: %w", err)
}

// loadFromAPI fetches issues from all configured teams via the Linear API.
func (p *LinearProvider) loadFromAPI() ([]*core.Task, error) {
	ctx := context.Background()
	var allTasks []*core.Task

	for _, teamID := range p.config.TeamIDs {
		issues, err := p.fetchTeamIssues(ctx, teamID)
		if err != nil {
			return nil, fmt.Errorf("fetch team %s: %w", teamID, err)
		}

		for i := range issues {
			task := mapIssueToTask(&issues[i])
			allTasks = append(allTasks, task)
		}
	}

	if allTasks == nil {
		allTasks = []*core.Task{}
	}
	return allTasks, nil
}

// fetchTeamIssues fetches all issues for a team, filtering by assignee if configured (AC9).
func (p *LinearProvider) fetchTeamIssues(ctx context.Context, teamID string) ([]IssueNode, error) {
	issues, err := fetchAllIssues(ctx, p.client, teamID)
	if err != nil {
		return nil, err
	}

	if p.config.Assignee == "" {
		return issues, nil
	}

	var filtered []IssueNode
	for _, issue := range issues {
		if issue.Assignee == nil {
			continue
		}
		if issue.Assignee.Email == p.config.Assignee ||
			issue.Assignee.Name == p.config.Assignee ||
			issue.Assignee.ID == p.config.Assignee {
			filtered = append(filtered, issue)
		}
	}
	return filtered, nil
}

// fetchAllIssues fetches all issues for a team using cursor-based pagination.
// This is a package-level function to allow testing with mock clients.
func fetchAllIssues(ctx context.Context, client GraphQLClient, teamID string) ([]IssueNode, error) {
	var all []IssueNode
	cursor := ""

	for {
		conn, err := client.QueryTeamIssues(ctx, teamID, cursor)
		if err != nil {
			return nil, err
		}

		all = append(all, conn.Nodes...)

		if !conn.PageInfo.HasNextPage {
			break
		}
		cursor = conn.PageInfo.EndCursor
	}

	return all, nil
}

// mapIssueToTask converts a Linear IssueNode to a core.Task with field mappings (AC2).
func mapIssueToTask(issue *IssueNode) *core.Task {
	taskID := fmt.Sprintf("linear:%s", issue.Identifier)

	task := &core.Task{
		ID:             taskID,
		Text:           issue.Title,
		Status:         MapStatus(issue.State.Type),
		Effort:         MapEffort(issue.Priority, issue.Estimate),
		SourceProvider: providerName,
		SourceRefs: []core.SourceRef{
			{Provider: providerName, NativeID: taskID},
		},
		Notes:     []core.TaskNote{},
		CreatedAt: issue.CreatedAt,
		UpdatedAt: issue.UpdatedAt,
	}

	// AC2: description → Context (Markdown preserved)
	var contextParts []string
	if issue.Description != "" {
		contextParts = append(contextParts, issue.Description)
	}

	// AC2: labels → appended to Context
	if len(issue.Labels.Nodes) > 0 {
		var labels []string
		for _, l := range issue.Labels.Nodes {
			labels = append(labels, l.Name)
		}
		contextParts = append(contextParts, fmt.Sprintf("Labels: %s", strings.Join(labels, ", ")))
	}

	// AC2: dueDate → appended to Context
	if issue.DueDate != nil && *issue.DueDate != "" {
		contextParts = append(contextParts, fmt.Sprintf("Due: %s", *issue.DueDate))
	}

	if len(contextParts) > 0 {
		task.Context = strings.Join(contextParts, "\n\n")
	}

	// Set location to the Linear identifier for reference
	task.Location = core.TaskLocation(issue.Identifier)

	return task
}

// MapStatus maps a Linear state.type to a ThreeDoors TaskStatus (AC3).
func MapStatus(stateType string) core.TaskStatus {
	if status, ok := statusTypeMap[stateType]; ok {
		return status
	}
	return core.StatusTodo
}

// MapEffort maps Linear priority (and optional estimate) to ThreeDoors TaskEffort (AC4).
// Linear priority: 1=urgent, 2=high, 3=medium, 4=low, 0=no priority.
// When priority is 0 and estimate is present, estimate is used as a secondary signal.
func MapEffort(priority int, estimate *float64) core.TaskEffort {
	switch priority {
	case 1: // urgent → deep-work
		return core.EffortDeepWork
	case 2: // high → deep-work
		return core.EffortDeepWork
	case 3: // medium → medium
		return core.EffortMedium
	case 4: // low → quick-win
		return core.EffortQuickWin
	case 0: // no priority — use estimate as fallback
		return mapEstimateToEffort(estimate)
	default:
		return core.EffortMedium
	}
}

// mapEstimateToEffort maps a Linear estimate to effort when priority is 0 (AC4).
func mapEstimateToEffort(estimate *float64) core.TaskEffort {
	if estimate == nil {
		return core.EffortMedium
	}
	e := *estimate
	switch {
	case e <= 1:
		return core.EffortQuickWin
	case e <= 3:
		return core.EffortMedium
	default:
		return core.EffortDeepWork
	}
}

// SaveTask returns ErrReadOnly; Linear provider is read-only (AC5).
func (p *LinearProvider) SaveTask(_ *core.Task) error {
	return core.ErrReadOnly
}

// SaveTasks returns ErrReadOnly; Linear provider is read-only (AC5).
func (p *LinearProvider) SaveTasks(_ []*core.Task) error {
	return core.ErrReadOnly
}

// DeleteTask returns ErrReadOnly; Linear provider is read-only (AC5).
func (p *LinearProvider) DeleteTask(_ string) error {
	return core.ErrReadOnly
}

// MarkComplete returns ErrReadOnly; Linear provider is read-only (AC5).
func (p *LinearProvider) MarkComplete(_ string) error {
	return core.ErrReadOnly
}

// Watch returns a channel that emits ChangeEvents at poll_interval (AC6).
func (p *LinearProvider) Watch() <-chan core.ChangeEvent {
	if p.watchCh != nil {
		return p.watchCh
	}

	p.watchCh = make(chan core.ChangeEvent, 1)

	go func() {
		ticker := time.NewTicker(p.config.PollInterval)
		defer ticker.Stop()
		defer close(p.watchCh)

		for {
			select {
			case <-ticker.C:
				tasks, err := p.loadFromAPI()
				if err != nil {
					continue
				}
				p.cachedTasks = tasks
				p.lastCacheUpdate = time.Now().UTC()
				p.writeCache(tasks)

				select {
				case p.watchCh <- core.ChangeEvent{
					Type:   core.ChangeUpdated,
					Source: providerName,
				}:
				default:
				}
			case <-p.stopCh:
				return
			}
		}
	}()

	return p.watchCh
}

// Stop signals the Watch goroutine to exit.
func (p *LinearProvider) Stop() {
	select {
	case <-p.stopCh:
		// already stopped
	default:
		close(p.stopCh)
	}
}

// HealthCheck verifies API connectivity via QueryViewer.
func (p *LinearProvider) HealthCheck() core.HealthCheckResult {
	start := time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := p.client.QueryViewer(ctx)
	duration := time.Since(start)

	if err != nil {
		return core.HealthCheckResult{
			Items: []core.HealthCheckItem{
				{
					Name:       "linear_connectivity",
					Status:     core.HealthFail,
					Message:    fmt.Sprintf("Linear API unreachable: %v", err),
					Suggestion: "Check LINEAR_API_KEY and network connectivity",
				},
			},
			Overall:  core.HealthFail,
			Duration: duration,
		}
	}

	return core.HealthCheckResult{
		Items: []core.HealthCheckItem{
			{
				Name:    "linear_connectivity",
				Status:  core.HealthOK,
				Message: "Linear API reachable",
			},
		},
		Overall:  core.HealthOK,
		Duration: duration,
	}
}

// Factory creates a LinearProvider from a ProviderConfig.
func Factory(config *core.ProviderConfig) (core.TaskProvider, error) {
	settings := findLinearSettings(config)
	if settings == nil {
		return nil, fmt.Errorf("linear factory: no linear provider settings found")
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		return nil, fmt.Errorf("linear factory: %w", err)
	}

	client := NewLinearClient(cfg.APIKey)
	return NewLinearProvider(client, cfg), nil
}

// findLinearSettings locates the linear provider entry in the config.
func findLinearSettings(config *core.ProviderConfig) map[string]string {
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

// isCacheValid checks if the in-memory cache is still within TTL (AC7).
func (p *LinearProvider) isCacheValid() bool {
	if p.cachedTasks == nil {
		return false
	}
	return time.Since(p.lastCacheUpdate) < p.cacheTTL
}

// cacheEntry wraps tasks with a timestamp for cache staleness detection.
type cacheEntry struct {
	LastUpdated time.Time    `json:"last_updated"`
	Tasks       []*core.Task `json:"tasks"`
}

// writeCache writes tasks to the local cache file using atomic write (AC7).
func (p *LinearProvider) writeCache(tasks []*core.Task) {
	if p.cachePath == "" {
		return
	}

	entry := cacheEntry{
		LastUpdated: time.Now().UTC(),
		Tasks:       tasks,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to marshal linear cache: %v\n", err)
		return
	}

	dir := filepath.Dir(p.cachePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create linear cache dir: %v\n", err)
		return
	}

	tmpPath := p.cachePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create linear cache temp file: %v\n", err)
		return
	}

	w := bufio.NewWriter(f)
	if _, err := w.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to write linear cache: %v\n", err)
		return
	}

	if err := w.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to flush linear cache: %v\n", err)
		return
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to sync linear cache: %v\n", err)
		return
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to close linear cache: %v\n", err)
		return
	}

	if err := os.Rename(tmpPath, p.cachePath); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to rename linear cache: %v\n", err)
	}
}

// readCache reads tasks from the local cache file.
func (p *LinearProvider) readCache() ([]*core.Task, error) {
	data, err := os.ReadFile(p.cachePath)
	if err != nil {
		return nil, fmt.Errorf("read linear cache: %w", err)
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal linear cache: %w", err)
	}

	return entry.Tasks, nil
}
