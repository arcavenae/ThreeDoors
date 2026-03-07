package core

import (
	"errors"
	"testing"
)

// aggMockProvider is a configurable test double for TaskProvider used in aggregator tests.
type aggMockProvider struct {
	name       string
	tasks      []*Task
	loadErr    error
	saveErr    error
	deleteErr  error
	savedTasks []*Task
	deletedIDs []string
	completed  []string
}

func newAggMockProvider(name string, taskTexts ...string) *aggMockProvider {
	mp := &aggMockProvider{name: name}
	for _, text := range taskTexts {
		t := NewTask(text)
		t.SourceProvider = name
		mp.tasks = append(mp.tasks, t)
	}
	return mp
}

func (m *aggMockProvider) LoadTasks() ([]*Task, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.tasks, nil
}

func (m *aggMockProvider) SaveTask(task *Task) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedTasks = append(m.savedTasks, task)
	return nil
}

func (m *aggMockProvider) SaveTasks(tasks []*Task) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedTasks = append(m.savedTasks, tasks...)
	return nil
}

func (m *aggMockProvider) DeleteTask(taskID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, taskID)
	return nil
}

func (m *aggMockProvider) MarkComplete(taskID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.completed = append(m.completed, taskID)
	return nil
}

func (m *aggMockProvider) Name() string              { return m.name }
func (m *aggMockProvider) Watch() <-chan ChangeEvent { return nil }
func (m *aggMockProvider) HealthCheck() HealthCheckResult {
	return HealthCheckResult{}
}

func TestMultiSourceAggregator_LoadTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		providers     map[string]TaskProvider
		wantTaskCount int
		wantErr       bool
	}{
		{
			name: "loads from multiple providers",
			providers: map[string]TaskProvider{
				"textfile": newAggMockProvider("textfile", "task1", "task2"),
				"obsidian": newAggMockProvider("obsidian", "task3"),
			},
			wantTaskCount: 3,
			wantErr:       false,
		},
		{
			name: "single provider works",
			providers: map[string]TaskProvider{
				"textfile": newAggMockProvider("textfile", "task1"),
			},
			wantTaskCount: 1,
			wantErr:       false,
		},
		{
			name:          "no providers returns error",
			providers:     map[string]TaskProvider{},
			wantTaskCount: 0,
			wantErr:       true,
		},
		{
			name: "all providers failing returns error",
			providers: map[string]TaskProvider{
				"textfile": &aggMockProvider{name: "textfile", loadErr: errors.New("disk full")},
				"obsidian": &aggMockProvider{name: "obsidian", loadErr: errors.New("vault missing")},
			},
			wantTaskCount: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agg := NewMultiSourceAggregator(tt.providers)
			tasks, err := agg.LoadTasks()

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(tasks) != tt.wantTaskCount {
				t.Errorf("LoadTasks() returned %d tasks, want %d", len(tasks), tt.wantTaskCount)
			}
		})
	}
}

func TestMultiSourceAggregator_ProviderFailureIsolation(t *testing.T) {
	t.Parallel()

	failingProvider := &aggMockProvider{
		name:    "obsidian",
		loadErr: errors.New("vault not found"),
	}
	workingProvider := newAggMockProvider("textfile", "task1", "task2")

	providers := map[string]TaskProvider{
		"textfile": workingProvider,
		"obsidian": failingProvider,
	}

	agg := NewMultiSourceAggregator(providers)
	tasks, err := agg.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() should succeed when at least one provider works: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("LoadTasks() returned %d tasks, want 2 from working provider", len(tasks))
	}
}

func TestMultiSourceAggregator_SourceProviderMetadata(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile", "text task")
	obsidianProvider := newAggMockProvider("obsidian", "obsidian task")

	providers := map[string]TaskProvider{
		"textfile": textProvider,
		"obsidian": obsidianProvider,
	}

	agg := NewMultiSourceAggregator(providers)
	tasks, err := agg.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}

	sourceProviders := make(map[string]bool)
	for _, task := range tasks {
		if task.SourceProvider == "" {
			t.Errorf("task %q has empty SourceProvider", task.Text)
		}
		sourceProviders[task.SourceProvider] = true
	}

	if !sourceProviders["textfile"] {
		t.Error("expected tasks with SourceProvider 'textfile'")
	}
	if !sourceProviders["obsidian"] {
		t.Error("expected tasks with SourceProvider 'obsidian'")
	}
}

func TestMultiSourceAggregator_SaveTaskRouting(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	obsidianProvider := newAggMockProvider("obsidian")

	providers := map[string]TaskProvider{
		"textfile": textProvider,
		"obsidian": obsidianProvider,
	}

	agg := NewMultiSourceAggregator(providers)

	task := NewTask("test task")
	task.SourceProvider = "obsidian"

	if err := agg.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	if len(obsidianProvider.savedTasks) != 1 {
		t.Errorf("expected obsidian provider to receive save, got %d saves", len(obsidianProvider.savedTasks))
	}
	if len(textProvider.savedTasks) != 0 {
		t.Errorf("expected textfile provider to receive 0 saves, got %d", len(textProvider.savedTasks))
	}
}

func TestMultiSourceAggregator_SaveTaskUnknownSourceFallback(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	obsidianProvider := newAggMockProvider("obsidian")

	providers := map[string]TaskProvider{
		"textfile": textProvider,
		"obsidian": obsidianProvider,
	}

	agg := NewMultiSourceAggregatorWithDefault(providers, "textfile")

	task := NewTask("orphan task")
	task.SourceProvider = "" // no source

	if err := agg.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	if len(textProvider.savedTasks) != 1 {
		t.Errorf("expected fallback to textfile, got %d saves", len(textProvider.savedTasks))
	}
}

func TestMultiSourceAggregator_DeleteTaskRouting(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	obsidianProvider := newAggMockProvider("obsidian")

	providers := map[string]TaskProvider{
		"textfile": textProvider,
		"obsidian": obsidianProvider,
	}

	agg := NewMultiSourceAggregator(providers)
	agg.trackTaskOrigin("task-123", "obsidian")

	if err := agg.DeleteTask("task-123"); err != nil {
		t.Fatalf("DeleteTask() unexpected error: %v", err)
	}

	if len(obsidianProvider.deletedIDs) != 1 || obsidianProvider.deletedIDs[0] != "task-123" {
		t.Errorf("expected obsidian provider to receive delete for task-123")
	}
	if len(textProvider.deletedIDs) != 0 {
		t.Errorf("expected textfile provider to receive 0 deletes")
	}
}

func TestMultiSourceAggregator_MarkCompleteRouting(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	obsidianProvider := newAggMockProvider("obsidian")

	providers := map[string]TaskProvider{
		"textfile": textProvider,
		"obsidian": obsidianProvider,
	}

	agg := NewMultiSourceAggregator(providers)
	agg.trackTaskOrigin("task-456", "textfile")

	if err := agg.MarkComplete("task-456"); err != nil {
		t.Fatalf("MarkComplete() unexpected error: %v", err)
	}

	if len(textProvider.completed) != 1 || textProvider.completed[0] != "task-456" {
		t.Errorf("expected textfile provider to receive MarkComplete for task-456")
	}
}

func TestMultiSourceAggregator_SaveTasksGroupsByProvider(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	obsidianProvider := newAggMockProvider("obsidian")

	providers := map[string]TaskProvider{
		"textfile": textProvider,
		"obsidian": obsidianProvider,
	}

	agg := NewMultiSourceAggregatorWithDefault(providers, "textfile")

	tasks := []*Task{
		{ID: "1", Text: "text task", SourceProvider: "textfile"},
		{ID: "2", Text: "obs task 1", SourceProvider: "obsidian"},
		{ID: "3", Text: "obs task 2", SourceProvider: "obsidian"},
		{ID: "4", Text: "orphan", SourceProvider: ""},
	}

	if err := agg.SaveTasks(tasks); err != nil {
		t.Fatalf("SaveTasks() unexpected error: %v", err)
	}

	// textfile should get task 1 + orphan task 4
	if len(textProvider.savedTasks) != 2 {
		t.Errorf("textfile provider got %d tasks, want 2", len(textProvider.savedTasks))
	}
	// obsidian should get tasks 2 and 3
	if len(obsidianProvider.savedTasks) != 2 {
		t.Errorf("obsidian provider got %d tasks, want 2", len(obsidianProvider.savedTasks))
	}
}

func TestMultiSourceAggregator_ImplementsTaskProvider(t *testing.T) {
	t.Parallel()

	providers := map[string]TaskProvider{
		"textfile": newAggMockProvider("textfile"),
	}

	var _ TaskProvider = NewMultiSourceAggregator(providers)
}

func TestMultiSourceAggregator_GetProviderForTask(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	obsidianProvider := newAggMockProvider("obsidian")

	providers := map[string]TaskProvider{
		"textfile": textProvider,
		"obsidian": obsidianProvider,
	}

	agg := NewMultiSourceAggregator(providers)
	agg.trackTaskOrigin("task-abc", "obsidian")

	provider, err := agg.GetProviderForTask("task-abc")
	if err != nil {
		t.Fatalf("GetProviderForTask() unexpected error: %v", err)
	}
	if provider != obsidianProvider {
		t.Error("GetProviderForTask() returned wrong provider")
	}

	_, err = agg.GetProviderForTask("nonexistent")
	if err == nil {
		t.Error("GetProviderForTask() should error for unknown task")
	}
}

func TestMultiSourceAggregator_LoadErrors(t *testing.T) {
	t.Parallel()

	agg := NewMultiSourceAggregator(map[string]TaskProvider{
		"broken": &aggMockProvider{name: "broken", loadErr: errors.New("fail")},
	})

	_, err := agg.LoadTasks()
	if err == nil {
		t.Error("LoadTasks() should return error when all providers fail")
	}

	if !errors.Is(err, ErrAllProvidersFailed) {
		t.Errorf("expected ErrAllProvidersFailed, got: %v", err)
	}
}

func TestMultiSourceAggregator_LoadTasks_MigratesSourceRefs(t *testing.T) {
	t.Parallel()

	// Task with legacy SourceProvider, no SourceRefs
	task := NewTask("legacy task")
	task.SourceProvider = "textfile"

	provider := &aggMockProvider{name: "textfile", tasks: []*Task{task}}

	agg := NewMultiSourceAggregator(map[string]TaskProvider{"textfile": provider})
	tasks, err := agg.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	loaded := tasks[0]

	if len(loaded.SourceRefs) == 0 {
		t.Fatal("expected SourceRefs to be populated after load")
	}
	if !loaded.HasSourceRef("textfile", loaded.ID) {
		t.Errorf("expected SourceRef with provider %q and native ID %q", "textfile", loaded.ID)
	}
}

func TestMultiSourceAggregator_LoadTasks_AddsSourceRefForProvider(t *testing.T) {
	t.Parallel()

	// Task with no SourceProvider or SourceRefs at all
	task := NewTask("new task")

	provider := &aggMockProvider{name: "obsidian", tasks: []*Task{task}}

	agg := NewMultiSourceAggregator(map[string]TaskProvider{"obsidian": provider})
	tasks, err := agg.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}

	loaded := tasks[0]
	if !loaded.HasSourceRef("obsidian", loaded.ID) {
		t.Error("expected SourceRef added for the loading provider")
	}
}

func TestMultiSourceAggregator_SaveTask_MultiProviderWriteRouting(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	obsidianProvider := newAggMockProvider("obsidian")

	providers := map[string]TaskProvider{
		"textfile": textProvider,
		"obsidian": obsidianProvider,
	}

	agg := NewMultiSourceAggregatorWithDefault(providers, "textfile")

	task := NewTask("multi-provider task")
	task.AddSourceRef("textfile", task.ID)
	task.AddSourceRef("obsidian", "obs-note-123")

	if err := agg.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	if len(textProvider.savedTasks) != 1 {
		t.Errorf("textfile should get 1 save, got %d", len(textProvider.savedTasks))
	}
	if len(obsidianProvider.savedTasks) != 1 {
		t.Errorf("obsidian should get 1 save, got %d", len(obsidianProvider.savedTasks))
	}
}

func TestMultiSourceAggregator_SaveTask_FallsBackWithNoRefs(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	providers := map[string]TaskProvider{"textfile": textProvider}

	agg := NewMultiSourceAggregatorWithDefault(providers, "textfile")

	task := NewTask("no refs task")
	// No SourceRefs, no SourceProvider

	if err := agg.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	if len(textProvider.savedTasks) != 1 {
		t.Errorf("default provider should get save, got %d saves", len(textProvider.savedTasks))
	}
}

func TestMultiSourceAggregator_SaveTask_SkipsUnknownProviderInRefs(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	providers := map[string]TaskProvider{"textfile": textProvider}

	agg := NewMultiSourceAggregatorWithDefault(providers, "textfile")

	task := NewTask("task with unknown provider ref")
	task.AddSourceRef("textfile", task.ID)
	task.AddSourceRef("nonexistent", "xyz")

	if err := agg.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	if len(textProvider.savedTasks) != 1 {
		t.Errorf("known provider should get save, got %d", len(textProvider.savedTasks))
	}
}

func TestMultiSourceAggregator_SaveTask_DeduplicatesSameProvider(t *testing.T) {
	t.Parallel()

	textProvider := newAggMockProvider("textfile")
	providers := map[string]TaskProvider{"textfile": textProvider}

	agg := NewMultiSourceAggregatorWithDefault(providers, "textfile")

	task := NewTask("task")
	task.AddSourceRef("textfile", "id-1")
	task.AddSourceRef("textfile", "id-2") // same provider, different native ID

	if err := agg.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	// Should only save once to the same provider
	if len(textProvider.savedTasks) != 1 {
		t.Errorf("same provider should only get 1 save, got %d", len(textProvider.savedTasks))
	}
}
