package todoist

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
	providerName  = "todoist"
	cacheFileName = "todoist-cache.yaml"
)

// TaskFetcher abstracts the Todoist API operations needed by the provider.
// This enables testing without hitting a real Todoist instance.
type TaskFetcher interface {
	GetTasks(ctx context.Context, projectID, filter string) ([]TodoistTask, error)
	GetProjects(ctx context.Context) ([]TodoistProject, error)
}

// TodoistProvider implements core.TaskProvider for Todoist (read-only).
// Tasks are fetched from the Todoist REST API and mapped to ThreeDoors tasks.
// Write operations return core.ErrReadOnly.
type TodoistProvider struct {
	client     TaskFetcher
	projectIDs []string
	filter     string
	cachePath  string
}

// NewTodoistProvider creates a TodoistProvider with the given API client and config.
func NewTodoistProvider(client TaskFetcher, cfg *TodoistConfig) *TodoistProvider {
	return &TodoistProvider{
		client:     client,
		projectIDs: cfg.ProjectIDs,
		filter:     cfg.Filter,
	}
}

// SetCachePath sets the directory for the local task cache file.
func (p *TodoistProvider) SetCachePath(configDir string) {
	p.cachePath = filepath.Join(configDir, cacheFileName)
}

// Name returns the provider identifier.
func (p *TodoistProvider) Name() string {
	return providerName
}

// LoadTasks fetches tasks from the Todoist API, maps them to ThreeDoors tasks,
// and updates the local cache. Falls back to cache on API failure.
func (p *TodoistProvider) LoadTasks() ([]*core.Task, error) {
	tasks, err := p.loadFromAPI()
	if err == nil {
		p.writeCache(tasks)
		return tasks, nil
	}

	if p.cachePath != "" {
		cached, cacheErr := p.readCache()
		if cacheErr == nil {
			fmt.Fprintf(os.Stderr, "Warning: Todoist API unavailable (%v), using cached tasks\n", err)
			return cached, nil
		}
	}

	return nil, fmt.Errorf("todoist load tasks: %w", err)
}

// loadFromAPI fetches tasks from the Todoist API.
// If projectIDs are configured, fetches tasks for each project.
// Otherwise, fetches with the optional filter expression.
func (p *TodoistProvider) loadFromAPI() ([]*core.Task, error) {
	ctx := context.Background()

	// Resolve project names for source provider labels
	projectNames, _ := p.resolveProjectNames(ctx)

	var allTasks []*core.Task

	if len(p.projectIDs) > 0 {
		for _, pid := range p.projectIDs {
			apiTasks, err := p.client.GetTasks(ctx, pid, "")
			if err != nil {
				return nil, fmt.Errorf("project %s: %w", pid, err)
			}
			projectName := projectNames[pid]
			for i := range apiTasks {
				allTasks = append(allTasks, mapTodoistTask(&apiTasks[i], projectName))
			}
		}
	} else {
		apiTasks, err := p.client.GetTasks(ctx, "", p.filter)
		if err != nil {
			return nil, fmt.Errorf("fetch: %w", err)
		}
		for i := range apiTasks {
			projectName := projectNames[apiTasks[i].ProjectID]
			allTasks = append(allTasks, mapTodoistTask(&apiTasks[i], projectName))
		}
	}

	if allTasks == nil {
		allTasks = []*core.Task{}
	}

	return allTasks, nil
}

// resolveProjectNames fetches all projects and returns a map of ID->Name.
func (p *TodoistProvider) resolveProjectNames(ctx context.Context) (map[string]string, error) {
	projects, err := p.client.GetProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve project names: %w", err)
	}

	names := make(map[string]string, len(projects))
	for _, proj := range projects {
		names[proj.ID] = proj.Name
	}
	return names, nil
}

// mapTodoistTask converts a Todoist API task to a ThreeDoors Task.
func mapTodoistTask(t *TodoistTask, projectName string) *core.Task {
	now := time.Now().UTC()

	sourceProvider := providerName
	if projectName != "" {
		sourceProvider = fmt.Sprintf("todoist:%s", projectName)
	}

	// Build context from labels
	var taskContext string
	if len(t.Labels) > 0 {
		taskContext = strings.Join(t.Labels, ", ")
	}

	task := &core.Task{
		ID:             t.ID,
		Text:           t.Content,
		Context:        taskContext,
		Status:         MapStatus(t.IsCompleted),
		Effort:         MapPriorityToEffort(t.Priority),
		SourceProvider: sourceProvider,
		SourceRefs: []core.SourceRef{
			{Provider: providerName, NativeID: t.ID},
		},
		Notes:     []core.TaskNote{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if t.Description != "" {
		task.Context = t.Description
		if len(t.Labels) > 0 {
			task.Context = t.Description + " | " + strings.Join(t.Labels, ", ")
		}
	}

	return task
}

// SaveTask returns ErrReadOnly; Todoist provider is read-only.
func (p *TodoistProvider) SaveTask(_ *core.Task) error {
	return core.ErrReadOnly
}

// SaveTasks returns ErrReadOnly; Todoist provider is read-only.
func (p *TodoistProvider) SaveTasks(_ []*core.Task) error {
	return core.ErrReadOnly
}

// DeleteTask returns ErrReadOnly; Todoist provider is read-only.
func (p *TodoistProvider) DeleteTask(_ string) error {
	return core.ErrReadOnly
}

// MarkComplete returns ErrReadOnly; Todoist provider is read-only.
func (p *TodoistProvider) MarkComplete(_ string) error {
	return core.ErrReadOnly
}

// Watch returns nil; Todoist provider does not support webhook-based watching
// with personal API tokens.
func (p *TodoistProvider) Watch() <-chan core.ChangeEvent {
	return nil
}

// HealthCheck tests API connectivity by calling GetProjects.
func (p *TodoistProvider) HealthCheck() core.HealthCheckResult {
	start := time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := p.client.GetProjects(ctx)
	duration := time.Since(start)

	if err != nil {
		return core.HealthCheckResult{
			Items: []core.HealthCheckItem{
				{
					Name:       "todoist_connectivity",
					Status:     core.HealthFail,
					Message:    fmt.Sprintf("Todoist API unreachable: %v", err),
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
				Name:    "todoist_connectivity",
				Status:  core.HealthOK,
				Message: "Todoist API reachable",
			},
		},
		Overall:  core.HealthOK,
		Duration: duration,
	}
}

// Factory creates a TodoistProvider from a ProviderConfig.
func Factory(config *core.ProviderConfig) (core.TaskProvider, error) {
	settings := findTodoistSettings(config)
	if settings == nil {
		return nil, fmt.Errorf("todoist factory: no todoist provider settings found")
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		return nil, fmt.Errorf("todoist factory: %w", err)
	}

	client := NewClient(AuthConfigFrom(cfg))
	provider := NewTodoistProvider(client, cfg)

	configDir, dirErr := core.GetConfigDirPath()
	if dirErr == nil {
		provider.SetCachePath(configDir)
	}

	return provider, nil
}

// findTodoistSettings locates the todoist provider entry in the config.
func findTodoistSettings(config *core.ProviderConfig) map[string]string {
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
func (p *TodoistProvider) writeCache(tasks []*core.Task) {
	if p.cachePath == "" {
		return
	}

	entry := cacheEntry{
		LastUpdated: time.Now().UTC(),
		Tasks:       tasks,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to marshal todoist cache: %v\n", err)
		return
	}

	dir := filepath.Dir(p.cachePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create todoist cache dir: %v\n", err)
		return
	}

	tmpPath := p.cachePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create todoist cache temp file: %v\n", err)
		return
	}

	w := bufio.NewWriter(f)
	if _, err := w.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to write todoist cache: %v\n", err)
		return
	}

	if err := w.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to flush todoist cache: %v\n", err)
		return
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to sync todoist cache: %v\n", err)
		return
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to close todoist cache: %v\n", err)
		return
	}

	if err := os.Rename(tmpPath, p.cachePath); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to rename todoist cache: %v\n", err)
	}
}

// readCache reads tasks from the local cache file.
func (p *TodoistProvider) readCache() ([]*core.Task, error) {
	data, err := os.ReadFile(p.cachePath)
	if err != nil {
		return nil, fmt.Errorf("read todoist cache: %w", err)
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal todoist cache: %w", err)
	}

	return entry.Tasks, nil
}
