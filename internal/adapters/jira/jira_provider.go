package jira

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
	providerName        = "jira"
	searchPageSize      = 50
	maxConflictRetries  = 3
	conflictBaseBackoff = 500 * time.Millisecond
	cacheFileName       = "jira-cache.yaml"
)

// Searcher abstracts the Jira API operations needed by the provider.
// This enables testing without hitting a real Jira instance.
type Searcher interface {
	SearchJQL(ctx context.Context, jql string, fields []string, maxResults int, pageToken string) (*SearchResult, error)
	GetTransitions(ctx context.Context, issueKey string) ([]Transition, error)
	DoTransition(ctx context.Context, issueKey, transitionID string) error
}

// FieldMapper maps Jira issue fields to ThreeDoors task fields.
type FieldMapper struct {
	statusMap     map[string]core.TaskStatus
	effortMap     map[string]core.TaskEffort
	defaultEffort core.TaskEffort
}

// DefaultFieldMapper returns a FieldMapper with standard Jira-to-ThreeDoors mappings.
func DefaultFieldMapper() *FieldMapper {
	return &FieldMapper{
		statusMap: map[string]core.TaskStatus{
			"new":           core.StatusTodo,
			"undefined":     core.StatusTodo,
			"indeterminate": core.StatusInProgress,
			"done":          core.StatusComplete,
		},
		effortMap: map[string]core.TaskEffort{
			"Highest": core.EffortDeepWork,
			"High":    core.EffortDeepWork,
			"Medium":  core.EffortMedium,
			"Low":     core.EffortQuickWin,
			"Lowest":  core.EffortQuickWin,
		},
		defaultEffort: core.EffortMedium,
	}
}

// MapStatus converts a Jira statusCategory key to a ThreeDoors TaskStatus.
func (fm *FieldMapper) MapStatus(categoryKey string) core.TaskStatus {
	if status, ok := fm.statusMap[categoryKey]; ok {
		return status
	}
	return core.StatusTodo
}

// MapEffort converts a Jira priority name to a ThreeDoors TaskEffort.
func (fm *FieldMapper) MapEffort(priorityName string) core.TaskEffort {
	if effort, ok := fm.effortMap[priorityName]; ok {
		return effort
	}
	return fm.defaultEffort
}

// MapContext builds a ThreeDoors context string from project key and labels.
func (fm *FieldMapper) MapContext(projectKey string, labels []string) string {
	if len(labels) > 0 {
		return fmt.Sprintf("[%s] %s", projectKey, strings.Join(labels, ", "))
	}
	return fmt.Sprintf("[%s]", projectKey)
}

// MapIssueToTask converts a Jira Issue to a ThreeDoors Task.
func (fm *FieldMapper) MapIssueToTask(issue Issue) *core.Task {
	now := time.Now().UTC()

	priorityName := ""
	if issue.Fields.Priority != nil {
		priorityName = issue.Fields.Priority.Name
	}

	return &core.Task{
		ID:             issue.Key,
		Text:           issue.Fields.Summary,
		Context:        fm.MapContext(issue.Fields.Project.Key, issue.Fields.Labels),
		Status:         fm.MapStatus(issue.Fields.Status.StatusCategory.Key),
		Effort:         fm.MapEffort(priorityName),
		SourceProvider: providerName,
		SourceRefs: []core.SourceRef{
			{Provider: providerName, NativeID: issue.Key},
		},
		Notes:     []core.TaskNote{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// JiraProvider implements core.TaskProvider for Jira with bidirectional sync.
// MarkComplete transitions issues to Done via the Jira REST API.
// LoadTasks caches results locally for offline fallback.
type JiraProvider struct {
	searcher  Searcher
	jql       string
	mapper    *FieldMapper
	cachePath string
	sleepFn   func(time.Duration) // injectable for testing
}

// NewJiraProvider creates a JiraProvider with the given searcher, JQL query, and field mapper.
// Use SetCachePath to enable local cache for offline fallback.
func NewJiraProvider(searcher Searcher, jql string, mapper *FieldMapper) *JiraProvider {
	return &JiraProvider{
		searcher: searcher,
		jql:      jql,
		mapper:   mapper,
		sleepFn:  time.Sleep,
	}
}

// SetCachePath sets the directory for the local task cache file.
// The cache file is written on successful LoadTasks and read as fallback when offline.
func (p *JiraProvider) SetCachePath(configDir string) {
	p.cachePath = filepath.Join(configDir, cacheFileName)
}

// Name returns the provider identifier.
func (p *JiraProvider) Name() string {
	return providerName
}

// LoadTasks executes the configured JQL query, paginates all results,
// and maps them to ThreeDoors tasks. On success, updates the local cache.
// On failure, falls back to the local cache if available.
func (p *JiraProvider) LoadTasks() ([]*core.Task, error) {
	tasks, err := p.loadFromAPI()
	if err == nil {
		p.writeCache(tasks)
		return tasks, nil
	}

	// API failed — try cache fallback
	if p.cachePath != "" {
		cached, cacheErr := p.readCache()
		if cacheErr == nil {
			fmt.Fprintf(os.Stderr, "Warning: Jira API unavailable (%v), using cached tasks\n", err)
			return cached, nil
		}
	}

	return nil, fmt.Errorf("jira load tasks: %w", err)
}

// loadFromAPI fetches tasks from the Jira API with pagination.
func (p *JiraProvider) loadFromAPI() ([]*core.Task, error) {
	ctx := context.Background()
	fields := []string{"summary", "status", "priority", "project", "labels", "issuetype", "created", "updated"}

	var allTasks []*core.Task
	pageToken := ""

	for {
		result, err := p.searcher.SearchJQL(ctx, p.jql, fields, searchPageSize, pageToken)
		if err != nil {
			return nil, fmt.Errorf("search: %w", err)
		}

		for _, issue := range result.Issues {
			allTasks = append(allTasks, p.mapper.MapIssueToTask(issue))
		}

		if result.IsLast || result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	if allTasks == nil {
		allTasks = []*core.Task{}
	}

	return allTasks, nil
}

// SaveTask returns ErrReadOnly; Jira provider is read-only.
func (p *JiraProvider) SaveTask(_ *core.Task) error {
	return core.ErrReadOnly
}

// SaveTasks returns ErrReadOnly; Jira provider is read-only.
func (p *JiraProvider) SaveTasks(_ []*core.Task) error {
	return core.ErrReadOnly
}

// DeleteTask returns ErrReadOnly; Jira provider is read-only.
func (p *JiraProvider) DeleteTask(_ string) error {
	return core.ErrReadOnly
}

// MarkComplete transitions a Jira issue to Done by discovering the available
// transitions, finding one with statusCategory.key == "done", and executing it.
// If the issue is already Done, this is a no-op (idempotent for WAL replay).
// Retries on 409 Conflict with exponential backoff.
func (p *JiraProvider) MarkComplete(taskID string) error {
	ctx := context.Background()

	// Get available transitions to check current status and find Done transition
	transitions, err := p.searcher.GetTransitions(ctx, taskID)
	if err != nil {
		return fmt.Errorf("jira mark complete %s: get transitions: %w", taskID, err)
	}

	// Find a transition whose target statusCategory is "done"
	doneTransitionID := ""
	for _, t := range transitions {
		if t.To.StatusCategory.Key == "done" {
			doneTransitionID = t.ID
			break
		}
	}

	if doneTransitionID == "" {
		// No "done" transition available — issue may already be done
		return nil
	}

	// Execute the transition with conflict retry
	for attempt := range maxConflictRetries {
		err = p.searcher.DoTransition(ctx, taskID, doneTransitionID)
		if err == nil {
			return nil
		}

		if !IsConflictError(err) {
			return fmt.Errorf("jira mark complete %s: %w", taskID, err)
		}

		// 409 Conflict — retry with exponential backoff
		if attempt < maxConflictRetries-1 {
			backoff := conflictBaseBackoff * (1 << uint(attempt))
			p.sleepFn(backoff)
		}
	}

	return fmt.Errorf("jira mark complete %s: %w", taskID, err)
}

// Watch returns nil; Jira provider does not support file watching.
func (p *JiraProvider) Watch() <-chan core.ChangeEvent {
	return nil
}

// HealthCheck tests API connectivity by executing a minimal search.
func (p *JiraProvider) HealthCheck() core.HealthCheckResult {
	start := time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := p.searcher.SearchJQL(ctx, p.jql, []string{"summary"}, 1, "")

	duration := time.Since(start)

	if err != nil {
		return core.HealthCheckResult{
			Items: []core.HealthCheckItem{
				{
					Name:       "jira_connectivity",
					Status:     core.HealthFail,
					Message:    fmt.Sprintf("Jira API unreachable: %v", err),
					Suggestion: "Check Jira URL, credentials, and network connectivity",
				},
			},
			Overall:  core.HealthFail,
			Duration: duration,
		}
	}

	return core.HealthCheckResult{
		Items: []core.HealthCheckItem{
			{
				Name:    "jira_connectivity",
				Status:  core.HealthOK,
				Message: "Jira API reachable",
			},
		},
		Overall:  core.HealthOK,
		Duration: duration,
	}
}

// Factory creates a JiraProvider from a ProviderConfig.
// Uses ParseConfig for validation, defaults, and env var fallback.
func Factory(config *core.ProviderConfig) (core.TaskProvider, error) {
	settings := findJiraSettings(config)
	if settings == nil {
		return nil, fmt.Errorf("jira factory: no jira provider settings found")
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		return nil, fmt.Errorf("jira factory: %w", err)
	}

	authConfig := AuthConfig{
		Type:     cfg.AuthType,
		URL:      cfg.URL,
		Email:    cfg.Email,
		APIToken: cfg.APIToken,
	}

	client := NewClient(authConfig)
	return NewJiraProvider(client, cfg.JQL, DefaultFieldMapper()), nil
}

// findJiraSettings locates the jira provider entry in the config.
func findJiraSettings(config *core.ProviderConfig) map[string]string {
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
func (p *JiraProvider) writeCache(tasks []*core.Task) {
	if p.cachePath == "" {
		return
	}

	entry := cacheEntry{
		LastUpdated: time.Now().UTC(),
		Tasks:       tasks,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to marshal jira cache: %v\n", err)
		return
	}

	dir := filepath.Dir(p.cachePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create jira cache dir: %v\n", err)
		return
	}

	tmpPath := p.cachePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create jira cache temp file: %v\n", err)
		return
	}

	w := bufio.NewWriter(f)
	if _, err := w.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to write jira cache: %v\n", err)
		return
	}

	if err := w.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to flush jira cache: %v\n", err)
		return
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to sync jira cache: %v\n", err)
		return
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to close jira cache: %v\n", err)
		return
	}

	if err := os.Rename(tmpPath, p.cachePath); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to rename jira cache: %v\n", err)
	}
}

// readCache reads tasks from the local cache file.
func (p *JiraProvider) readCache() ([]*core.Task, error) {
	data, err := os.ReadFile(p.cachePath)
	if err != nil {
		return nil, fmt.Errorf("read jira cache: %w", err)
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal jira cache: %w", err)
	}

	return entry.Tasks, nil
}
