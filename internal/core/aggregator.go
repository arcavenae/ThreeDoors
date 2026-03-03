package core

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

// ErrAllProvidersFailed is returned when every configured provider fails to load tasks.
var ErrAllProvidersFailed = errors.New("all providers failed to load tasks")

// ErrNoProviders is returned when no providers are configured.
var ErrNoProviders = errors.New("no providers configured")

// MultiSourceAggregator merges tasks from multiple TaskProvider instances into
// a unified pool. It implements TaskProvider so it can be used as a drop-in
// replacement for single-provider usage. Writes are routed back to the
// originating provider based on each task's SourceProvider field.
type MultiSourceAggregator struct {
	mu              sync.RWMutex
	providers       map[string]TaskProvider
	taskOrigins     map[string]string // taskID → provider name
	defaultProvider string
}

// NewMultiSourceAggregator creates an aggregator over the given named providers.
// The first provider in iteration order is used as the default for tasks with
// unknown origin. For deterministic default selection, use NewMultiSourceAggregatorWithDefault.
func NewMultiSourceAggregator(providers map[string]TaskProvider) *MultiSourceAggregator {
	defaultName := ""
	for name := range providers {
		defaultName = name
		break
	}
	return &MultiSourceAggregator{
		providers:       providers,
		taskOrigins:     make(map[string]string),
		defaultProvider: defaultName,
	}
}

// NewMultiSourceAggregatorWithDefault creates an aggregator with an explicit default
// provider name for tasks whose SourceProvider is empty or unknown.
func NewMultiSourceAggregatorWithDefault(providers map[string]TaskProvider, defaultProvider string) *MultiSourceAggregator {
	return &MultiSourceAggregator{
		providers:       providers,
		taskOrigins:     make(map[string]string),
		defaultProvider: defaultProvider,
	}
}

// LoadTasks collects tasks from all providers. Individual provider failures are
// isolated — if at least one provider succeeds, its tasks are returned and
// failures are logged as warnings. Returns ErrAllProvidersFailed only if every
// provider fails, and ErrNoProviders if none are configured.
func (a *MultiSourceAggregator) LoadTasks() ([]*Task, error) {
	if len(a.providers) == 0 {
		return nil, ErrNoProviders
	}

	var allTasks []*Task
	var errs []error
	succeeded := 0

	for name, provider := range a.providers {
		tasks, err := provider.LoadTasks()
		if err != nil {
			log.Printf("Warning: provider %q failed to load tasks: %v", name, err)
			errs = append(errs, fmt.Errorf("provider %q: %w", name, err))
			continue
		}

		for _, t := range tasks {
			t.SourceProvider = name
			a.trackTaskOrigin(t.ID, name)
		}
		allTasks = append(allTasks, tasks...)
		succeeded++
	}

	if succeeded == 0 {
		return nil, fmt.Errorf("%w: %v", ErrAllProvidersFailed, errors.Join(errs...))
	}

	return allTasks, nil
}

// SaveTask routes the save to the task's originating provider based on
// SourceProvider. If SourceProvider is empty or unknown, the default provider is used.
func (a *MultiSourceAggregator) SaveTask(task *Task) error {
	provider, _ := a.resolveProvider(task.SourceProvider)
	return provider.SaveTask(task)
}

// SaveTasks groups tasks by their SourceProvider and saves each group to
// its originating provider. Tasks with empty SourceProvider go to the default.
func (a *MultiSourceAggregator) SaveTasks(tasks []*Task) error {
	grouped := make(map[string][]*Task)
	for _, t := range tasks {
		name := t.SourceProvider
		if name == "" {
			name = a.defaultProvider
		}
		grouped[name] = append(grouped[name], t)
	}

	var errs []error
	for name, group := range grouped {
		provider, ok := a.providers[name]
		if !ok {
			provider = a.providers[a.defaultProvider]
		}
		if err := provider.SaveTasks(group); err != nil {
			errs = append(errs, fmt.Errorf("save to provider %q: %w", name, err))
		}
	}

	return errors.Join(errs...)
}

// DeleteTask routes the delete to the task's originating provider.
func (a *MultiSourceAggregator) DeleteTask(taskID string) error {
	providerName := a.getTaskOrigin(taskID)
	provider, _ := a.resolveProvider(providerName)
	return provider.DeleteTask(taskID)
}

// MarkComplete routes the completion to the task's originating provider.
func (a *MultiSourceAggregator) MarkComplete(taskID string) error {
	providerName := a.getTaskOrigin(taskID)
	provider, _ := a.resolveProvider(providerName)
	return provider.MarkComplete(taskID)
}

// Name returns a composite name listing all aggregated providers.
func (a *MultiSourceAggregator) Name() string {
	return "multi-source"
}

// Watch returns nil because the aggregator does not merge watch channels.
// Callers should watch individual providers via GetProviderForTask.
func (a *MultiSourceAggregator) Watch() <-chan ChangeEvent {
	return nil
}

// HealthCheck collects health results from all aggregated providers.
func (a *MultiSourceAggregator) HealthCheck() HealthCheckResult {
	result := HealthCheckResult{}
	for name, provider := range a.providers {
		sub := provider.HealthCheck()
		for _, item := range sub.Items {
			item.Name = name + ": " + item.Name
			result.Items = append(result.Items, item)
		}
	}
	return result
}

// GetProviderForTask returns the TaskProvider instance that owns the given task.
func (a *MultiSourceAggregator) GetProviderForTask(taskID string) (TaskProvider, error) {
	providerName := a.getTaskOrigin(taskID)
	if providerName == "" {
		return nil, fmt.Errorf("get provider for task %q: origin unknown", taskID)
	}
	provider, ok := a.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("get provider for task %q: provider %q not found", taskID, providerName)
	}
	return provider, nil
}

// trackTaskOrigin records which provider a task came from.
func (a *MultiSourceAggregator) trackTaskOrigin(taskID, providerName string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.taskOrigins[taskID] = providerName
}

// getTaskOrigin returns the provider name for a task, or empty string if unknown.
func (a *MultiSourceAggregator) getTaskOrigin(taskID string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.taskOrigins[taskID]
}

// resolveProvider returns the provider for the given name, falling back to
// the default provider. Always returns a valid provider as long as the
// aggregator has at least one provider.
func (a *MultiSourceAggregator) resolveProvider(name string) (TaskProvider, bool) {
	if name != "" {
		if p, ok := a.providers[name]; ok {
			return p, true
		}
	}
	return a.providers[a.defaultProvider], false
}
