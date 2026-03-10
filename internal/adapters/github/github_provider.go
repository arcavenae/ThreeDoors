package github

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

const (
	providerName        = "github"
	cacheFileName       = "github-cache.yaml"
	defaultTTL          = 5 * time.Minute
	maxRateLimitRetries = 3
)

// default priority label to effort mappings per AC4
var defaultPriorityEffort = map[string]core.TaskEffort{
	"priority:critical": core.EffortDeepWork,
	"priority:high":     core.EffortDeepWork,
	"priority:medium":   core.EffortMedium,
	"priority:low":      core.EffortQuickWin,
}

// IssueLister abstracts the GitHub API operations needed by the provider.
// This enables testing without hitting a real GitHub instance.
type IssueLister interface {
	ListIssues(ctx context.Context, owner, repo, assignee string) ([]*GitHubIssue, error)
	GetAuthenticatedUser(ctx context.Context) (string, error)
	CloseIssue(ctx context.Context, owner, repo string, issueNumber int) error
}

// GitHubProvider implements core.TaskProvider for GitHub Issues.
// Save and Delete return core.ErrReadOnly. MarkComplete closes issues via the GitHub API.
type GitHubProvider struct {
	client          IssueLister
	config          *GitHubConfig
	cachePath       string
	cacheTTL        time.Duration
	lastCacheUpdate time.Time
	cachedTasks     []*core.Task
	effortMap       map[string]core.TaskEffort
	watchCh         chan core.ChangeEvent
	stopCh          chan struct{}
	cb              *core.CircuitBreaker
	sleepFn         func(time.Duration) // injectable for testing
}

// NewGitHubProvider creates a GitHubProvider with the given client and config.
func NewGitHubProvider(client IssueLister, config *GitHubConfig) *GitHubProvider {
	effortMap := make(map[string]core.TaskEffort)
	for k, v := range defaultPriorityEffort {
		effortMap[k] = v
	}
	// Override with configured priority labels
	if config.PriorityLabels != nil {
		for label, effort := range config.PriorityLabels {
			effortMap[label] = core.TaskEffort(effort)
		}
	}

	return &GitHubProvider{
		client:    client,
		config:    config,
		cacheTTL:  config.PollInterval,
		effortMap: effortMap,
		stopCh:    make(chan struct{}),
		cb:        core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		sleepFn:   time.Sleep,
	}
}

// SetCachePath sets the directory for the local task cache file.
func (p *GitHubProvider) SetCachePath(configDir string) {
	p.cachePath = filepath.Join(configDir, cacheFileName)
}

// Name returns the provider identifier.
func (p *GitHubProvider) Name() string {
	return providerName
}

// LoadTasks fetches issues from all configured repos, maps them to tasks,
// and returns a merged result. Uses cache when within TTL.
func (p *GitHubProvider) LoadTasks() ([]*core.Task, error) {
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
			fmt.Fprintf(os.Stderr, "Warning: GitHub API unavailable (%v), using cached tasks\n", err)
			p.cachedTasks = cached
			return cached, nil
		}
	}

	return nil, fmt.Errorf("github load tasks: %w", err)
}

// loadFromAPI fetches issues from all configured repos via the GitHub API.
func (p *GitHubProvider) loadFromAPI() ([]*core.Task, error) {
	ctx := context.Background()

	// Resolve @me assignee
	assignee := p.config.Assignee
	if assignee == "@me" {
		user, err := p.client.GetAuthenticatedUser(ctx)
		if err != nil {
			return nil, fmt.Errorf("resolve @me assignee: %w", err)
		}
		assignee = user
	}

	var allTasks []*core.Task
	for _, repoSpec := range p.config.Repos {
		owner, repo, err := splitOwnerRepo(repoSpec)
		if err != nil {
			return nil, fmt.Errorf("invalid repo %q: %w", repoSpec, err)
		}

		issues, err := p.client.ListIssues(ctx, owner, repo, assignee)
		if err != nil {
			return nil, fmt.Errorf("list issues %s: %w", repoSpec, err)
		}

		for _, issue := range issues {
			allTasks = append(allTasks, p.mapIssueToTask(issue))
		}
	}

	if allTasks == nil {
		allTasks = []*core.Task{}
	}
	return allTasks, nil
}

// mapIssueToTask converts a GitHubIssue to a core.Task with field mappings.
func (p *GitHubProvider) mapIssueToTask(issue *GitHubIssue) *core.Task {
	taskID := fmt.Sprintf("github:%s#%d", issue.Repo, issue.Number)
	location := fmt.Sprintf("%s#%d", issue.Repo, issue.Number)

	task := &core.Task{
		ID:             taskID,
		Text:           issue.Title,
		Context:        issue.Body,
		Status:         p.mapStatus(issue),
		Effort:         p.mapEffort(issue.Labels),
		SourceProvider: providerName,
		SourceRefs: []core.SourceRef{
			{Provider: providerName, NativeID: taskID},
		},
		Notes:     []core.TaskNote{},
		CreatedAt: issue.CreatedAt,
		UpdatedAt: issue.UpdatedAt,
	}

	// AC11: Location set to owner/repo#number
	task.Location = core.TaskLocation(location)

	// AC5: milestone.due_on mapped to due date context
	if issue.MilestoneDueOn != nil {
		if task.Context != "" {
			task.Context += "\n\n"
		}
		task.Context += fmt.Sprintf("Due: %s", issue.MilestoneDueOn.Format("2006-01-02"))
	}

	return task
}

// mapStatus converts GitHub issue state to ThreeDoors TaskStatus (AC2, AC3).
func (p *GitHubProvider) mapStatus(issue *GitHubIssue) core.TaskStatus {
	if issue.State == "closed" {
		return core.StatusComplete
	}

	// Check for in-progress label
	for _, label := range issue.Labels {
		if label == p.config.InProgressLabel {
			return core.StatusInProgress
		}
	}

	return core.StatusTodo
}

// mapEffort converts GitHub labels to effort level (AC4).
func (p *GitHubProvider) mapEffort(labels []string) core.TaskEffort {
	for _, label := range labels {
		if effort, ok := p.effortMap[label]; ok {
			return effort
		}
	}
	return core.EffortQuickWin // default per AC4
}

// SaveTask returns ErrReadOnly; GitHub provider is read-only (AC6).
func (p *GitHubProvider) SaveTask(_ *core.Task) error {
	return core.ErrReadOnly
}

// SaveTasks returns ErrReadOnly; GitHub provider is read-only (AC6).
func (p *GitHubProvider) SaveTasks(_ []*core.Task) error {
	return core.ErrReadOnly
}

// DeleteTask returns ErrReadOnly; GitHub provider is read-only (AC6).
func (p *GitHubProvider) DeleteTask(_ string) error {
	return core.ErrReadOnly
}

// MarkComplete closes a GitHub issue by calling the API via the circuit breaker.
// The taskID must be in the format "github:<owner>/<repo>#<number>".
// Rate limit errors are retried with exponential backoff respecting Retry-After.
// Other API failures are returned as-is for WALProvider to queue.
func (p *GitHubProvider) MarkComplete(taskID string) error {
	owner, repo, number, err := parseTaskID(taskID)
	if err != nil {
		return fmt.Errorf("github mark complete: %w", err)
	}

	return p.cb.Execute(func() error {
		return p.closeIssueWithRetry(owner, repo, number)
	})
}

// closeIssueWithRetry calls CloseIssue with rate limit retry and exponential backoff.
func (p *GitHubProvider) closeIssueWithRetry(owner, repo string, number int) error {
	ctx := context.Background()
	var lastErr error

	for attempt := range maxRateLimitRetries {
		err := p.client.CloseIssue(ctx, owner, repo, number)
		if err == nil {
			return nil
		}

		var rle *RateLimitError
		if !errors.As(err, &rle) {
			return fmt.Errorf("close issue %s/%s#%d: %w", owner, repo, number, err)
		}

		lastErr = err
		if attempt < maxRateLimitRetries-1 {
			backoff := rle.RetryAfter
			if backoff <= 0 {
				backoff = time.Duration(1<<uint(attempt)) * time.Second
			}
			p.sleepFn(backoff)
		}
	}

	return fmt.Errorf("close issue %s/%s#%d: %w", owner, repo, number, lastErr)
}

// parseTaskID extracts owner, repo, and issue number from "github:<owner>/<repo>#<number>".
func parseTaskID(taskID string) (string, string, int, error) {
	rest, found := strings.CutPrefix(taskID, "github:")
	if !found {
		return "", "", 0, fmt.Errorf("invalid github task ID %q: missing github: prefix", taskID)
	}

	hashIdx := strings.LastIndex(rest, "#")
	if hashIdx < 0 {
		return "", "", 0, fmt.Errorf("invalid github task ID %q: missing # separator", taskID)
	}

	repoSpec := rest[:hashIdx]
	numberStr := rest[hashIdx+1:]

	owner, repo, err := splitOwnerRepo(repoSpec)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid github task ID %q: %w", taskID, err)
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid github task ID %q: bad issue number: %w", taskID, err)
	}

	return owner, repo, number, nil
}

// Watch returns a channel that emits ChangeEvents on polling interval (AC13).
func (p *GitHubProvider) Watch() <-chan core.ChangeEvent {
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

// HealthCheck verifies API connectivity via GetAuthenticatedUser (AC7).
func (p *GitHubProvider) HealthCheck() core.HealthCheckResult {
	start := time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := p.client.GetAuthenticatedUser(ctx)
	duration := time.Since(start)

	if err != nil {
		return core.HealthCheckResult{
			Items: []core.HealthCheckItem{
				{
					Name:       "github_connectivity",
					Status:     core.HealthFail,
					Message:    fmt.Sprintf("GitHub API unreachable: %v", err),
					Suggestion: "Check GITHUB_TOKEN and network connectivity",
				},
			},
			Overall:  core.HealthFail,
			Duration: duration,
		}
	}

	return core.HealthCheckResult{
		Items: []core.HealthCheckItem{
			{
				Name:    "github_connectivity",
				Status:  core.HealthOK,
				Message: "GitHub API reachable",
			},
		},
		Overall:  core.HealthOK,
		Duration: duration,
	}
}

// Factory creates a GitHubProvider from a ProviderConfig (AC10).
func Factory(config *core.ProviderConfig) (core.TaskProvider, error) {
	settings := findGitHubSettings(config)
	if settings == nil {
		return nil, fmt.Errorf("github factory: no github provider settings found")
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		return nil, fmt.Errorf("github factory: %w", err)
	}

	client := NewGitHubClient(cfg)
	return NewGitHubProvider(client, cfg), nil
}

// findGitHubSettings locates the github provider entry in the config.
func findGitHubSettings(config *core.ProviderConfig) map[string]string {
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

// isCacheValid checks if the in-memory cache is still within TTL.
func (p *GitHubProvider) isCacheValid() bool {
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

// writeCache writes tasks to the local cache file using atomic write.
func (p *GitHubProvider) writeCache(tasks []*core.Task) {
	if p.cachePath == "" {
		return
	}

	entry := cacheEntry{
		LastUpdated: time.Now().UTC(),
		Tasks:       tasks,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to marshal github cache: %v\n", err)
		return
	}

	dir := filepath.Dir(p.cachePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create github cache dir: %v\n", err)
		return
	}

	tmpPath := p.cachePath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create github cache temp file: %v\n", err)
		return
	}

	w := bufio.NewWriter(f)
	if _, err := w.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to write github cache: %v\n", err)
		return
	}

	if err := w.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to flush github cache: %v\n", err)
		return
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to sync github cache: %v\n", err)
		return
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to close github cache: %v\n", err)
		return
	}

	if err := os.Rename(tmpPath, p.cachePath); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Warning: failed to rename github cache: %v\n", err)
	}
}

// readCache reads tasks from the local cache file.
func (p *GitHubProvider) readCache() ([]*core.Task, error) {
	data, err := os.ReadFile(p.cachePath)
	if err != nil {
		return nil, fmt.Errorf("read github cache: %w", err)
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal github cache: %w", err)
	}

	return entry.Tasks, nil
}

// splitOwnerRepo splits "owner/repo" into its parts.
func splitOwnerRepo(repoSpec string) (string, string, error) {
	parts := strings.SplitN(repoSpec, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("must be in owner/repo format")
	}
	return parts[0], parts[1], nil
}
