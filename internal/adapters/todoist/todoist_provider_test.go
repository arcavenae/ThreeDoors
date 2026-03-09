package todoist

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// mockTaskFetcher implements TaskFetcher for testing.
type mockTaskFetcher struct {
	tasks      []TodoistTask
	tasksErr   error
	projects   []TodoistProject
	projectErr error
	// Track calls for verification
	getTasksCalls   int
	getProjectCalls int
	lastProjectID   string
	lastFilter      string
}

func (m *mockTaskFetcher) GetTasks(_ context.Context, projectID, filter string) ([]TodoistTask, error) {
	m.getTasksCalls++
	m.lastProjectID = projectID
	m.lastFilter = filter
	if m.tasksErr != nil {
		return nil, m.tasksErr
	}
	return m.tasks, nil
}

func (m *mockTaskFetcher) GetProjects(_ context.Context) ([]TodoistProject, error) {
	m.getProjectCalls++
	if m.projectErr != nil {
		return nil, m.projectErr
	}
	return m.projects, nil
}

func TestTodoistProviderName(t *testing.T) {
	t.Parallel()
	p := NewTodoistProvider(&mockTaskFetcher{}, &TodoistConfig{})
	if got := p.Name(); got != "todoist" {
		t.Errorf("Name() = %q, want %q", got, "todoist")
	}
}

func TestTodoistProviderLoadTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mock       *mockTaskFetcher
		cfg        *TodoistConfig
		wantCount  int
		wantErr    bool
		wantText   string
		wantEffort core.TaskEffort
		wantStatus core.TaskStatus
	}{
		{
			name: "basic task mapping",
			mock: &mockTaskFetcher{
				tasks: []TodoistTask{
					{
						ID:          "123",
						Content:     "Buy groceries",
						Description: "Milk, eggs, bread",
						Priority:    1,
						Labels:      []string{"shopping"},
						IsCompleted: false,
						ProjectID:   "proj1",
					},
				},
				projects: []TodoistProject{
					{ID: "proj1", Name: "Personal"},
				},
			},
			cfg:        &TodoistConfig{},
			wantCount:  1,
			wantText:   "Buy groceries",
			wantEffort: core.EffortQuickWin,
			wantStatus: core.StatusTodo,
		},
		{
			name: "completed task maps to complete status",
			mock: &mockTaskFetcher{
				tasks: []TodoistTask{
					{
						ID:          "456",
						Content:     "Done task",
						Priority:    3,
						IsCompleted: true,
						ProjectID:   "proj1",
					},
				},
				projects: []TodoistProject{
					{ID: "proj1", Name: "Work"},
				},
			},
			cfg:        &TodoistConfig{},
			wantCount:  1,
			wantStatus: core.StatusComplete,
			wantEffort: core.EffortDeepWork,
		},
		{
			name: "empty task list returns empty slice",
			mock: &mockTaskFetcher{
				tasks:    []TodoistTask{},
				projects: []TodoistProject{},
			},
			cfg:       &TodoistConfig{},
			wantCount: 0,
		},
		{
			name: "API error propagates",
			mock: &mockTaskFetcher{
				tasksErr: errors.New("network error"),
				projects: []TodoistProject{},
			},
			cfg:     &TodoistConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := NewTodoistProvider(tt.mock, tt.cfg)
			tasks, err := p.LoadTasks()

			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadTasks() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if len(tasks) != tt.wantCount {
				t.Fatalf("LoadTasks() returned %d tasks, want %d", len(tasks), tt.wantCount)
			}
			if tt.wantCount == 0 {
				return
			}

			task := tasks[0]
			if tt.wantText != "" && task.Text != tt.wantText {
				t.Errorf("task.Text = %q, want %q", task.Text, tt.wantText)
			}
			if tt.wantEffort != "" && task.Effort != tt.wantEffort {
				t.Errorf("task.Effort = %q, want %q", task.Effort, tt.wantEffort)
			}
			if tt.wantStatus != "" && task.Status != tt.wantStatus {
				t.Errorf("task.Status = %q, want %q", task.Status, tt.wantStatus)
			}
		})
	}
}

func TestTodoistProviderFieldMapping(t *testing.T) {
	t.Parallel()

	mock := &mockTaskFetcher{
		tasks: []TodoistTask{
			{
				ID:          "task-1",
				Content:     "Task content",
				Description: "Task description",
				Priority:    2,
				Labels:      []string{"urgent", "work"},
				IsCompleted: false,
				ProjectID:   "proj-1",
			},
		},
		projects: []TodoistProject{
			{ID: "proj-1", Name: "MyProject"},
		},
	}

	p := NewTodoistProvider(mock, &TodoistConfig{})
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]

	// AC1: ID maps directly
	if task.ID != "task-1" {
		t.Errorf("ID = %q, want %q", task.ID, "task-1")
	}

	// AC4: content -> Text
	if task.Text != "Task content" {
		t.Errorf("Text = %q, want %q", task.Text, "Task content")
	}

	// AC4: description -> Context (with labels appended)
	if task.Context != "Task description | urgent, work" {
		t.Errorf("Context = %q, want %q", task.Context, "Task description | urgent, work")
	}

	// AC5: priority 2 -> medium
	if task.Effort != core.EffortMedium {
		t.Errorf("Effort = %q, want %q", task.Effort, core.EffortMedium)
	}

	// AC4: is_completed false -> todo
	if task.Status != core.StatusTodo {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusTodo)
	}

	// AC7: SourceProvider includes project name
	if task.SourceProvider != "todoist:MyProject" {
		t.Errorf("SourceProvider = %q, want %q", task.SourceProvider, "todoist:MyProject")
	}

	// SourceRefs set correctly
	if len(task.SourceRefs) != 1 {
		t.Fatalf("SourceRefs length = %d, want 1", len(task.SourceRefs))
	}
	if task.SourceRefs[0].Provider != "todoist" {
		t.Errorf("SourceRefs[0].Provider = %q, want %q", task.SourceRefs[0].Provider, "todoist")
	}
	if task.SourceRefs[0].NativeID != "task-1" {
		t.Errorf("SourceRefs[0].NativeID = %q, want %q", task.SourceRefs[0].NativeID, "task-1")
	}
}

func TestTodoistProviderLabelsAsContext(t *testing.T) {
	t.Parallel()

	mock := &mockTaskFetcher{
		tasks: []TodoistTask{
			{
				ID:        "t1",
				Content:   "No description task",
				Labels:    []string{"home", "weekend"},
				ProjectID: "p1",
			},
		},
		projects: []TodoistProject{
			{ID: "p1", Name: "Tasks"},
		},
	}

	p := NewTodoistProvider(mock, &TodoistConfig{})
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	// AC6: labels mapped to context when no description
	if tasks[0].Context != "home, weekend" {
		t.Errorf("Context = %q, want %q", tasks[0].Context, "home, weekend")
	}
}

func TestTodoistProviderProjectFilter(t *testing.T) {
	t.Parallel()

	mock := &mockTaskFetcher{
		tasks: []TodoistTask{
			{ID: "t1", Content: "Task 1", ProjectID: "proj-a"},
		},
		projects: []TodoistProject{
			{ID: "proj-a", Name: "Alpha"},
		},
	}

	cfg := &TodoistConfig{
		ProjectIDs: []string{"proj-a"},
	}

	p := NewTodoistProvider(mock, cfg)
	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if mock.lastProjectID != "proj-a" {
		t.Errorf("GetTasks called with projectID = %q, want %q", mock.lastProjectID, "proj-a")
	}
}

func TestTodoistProviderFilterExpression(t *testing.T) {
	t.Parallel()

	mock := &mockTaskFetcher{
		tasks:    []TodoistTask{},
		projects: []TodoistProject{},
	}

	cfg := &TodoistConfig{
		Filter: "today | overdue",
	}

	p := NewTodoistProvider(mock, cfg)
	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if mock.lastFilter != "today | overdue" {
		t.Errorf("GetTasks called with filter = %q, want %q", mock.lastFilter, "today | overdue")
	}
}

func TestTodoistProviderReadOnlyMethods(t *testing.T) {
	t.Parallel()

	p := NewTodoistProvider(&mockTaskFetcher{}, &TodoistConfig{})

	// AC8: all write methods return ErrReadOnly
	if err := p.SaveTask(&core.Task{}); !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("SaveTask() = %v, want ErrReadOnly", err)
	}
	if err := p.SaveTasks([]*core.Task{}); !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("SaveTasks() = %v, want ErrReadOnly", err)
	}
	if err := p.DeleteTask("id"); !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("DeleteTask() = %v, want ErrReadOnly", err)
	}
	if err := p.MarkComplete("id"); !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("MarkComplete() = %v, want ErrReadOnly", err)
	}
}

func TestTodoistProviderWatch(t *testing.T) {
	t.Parallel()

	p := NewTodoistProvider(&mockTaskFetcher{}, &TodoistConfig{})

	// AC9: Watch returns nil
	if ch := p.Watch(); ch != nil {
		t.Errorf("Watch() = %v, want nil", ch)
	}
}

func TestTodoistProviderHealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("healthy", func(t *testing.T) {
		t.Parallel()
		mock := &mockTaskFetcher{
			projects: []TodoistProject{{ID: "1", Name: "Test"}},
		}
		p := NewTodoistProvider(mock, &TodoistConfig{})

		result := p.HealthCheck()
		if result.Overall != core.HealthOK {
			t.Errorf("Overall = %q, want %q", result.Overall, core.HealthOK)
		}
		if len(result.Items) != 1 {
			t.Fatalf("Items length = %d, want 1", len(result.Items))
		}
		if result.Items[0].Name != "todoist_connectivity" {
			t.Errorf("Item name = %q, want %q", result.Items[0].Name, "todoist_connectivity")
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		t.Parallel()
		mock := &mockTaskFetcher{
			projectErr: errors.New("auth failed"),
		}
		p := NewTodoistProvider(mock, &TodoistConfig{})

		result := p.HealthCheck()
		if result.Overall != core.HealthFail {
			t.Errorf("Overall = %q, want %q", result.Overall, core.HealthFail)
		}
		if result.Items[0].Suggestion == "" {
			t.Error("expected non-empty suggestion on failure")
		}
	})
}

func TestTodoistProviderCacheFallback(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, cacheFileName)

	// Pre-populate cache
	cached := cacheEntry{
		Tasks: []*core.Task{
			{ID: "cached-1", Text: "Cached task", Status: core.StatusTodo},
		},
	}
	data, err := json.Marshal(cached)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cachePath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	mock := &mockTaskFetcher{
		tasksErr: errors.New("API down"),
		projects: []TodoistProject{},
	}

	p := NewTodoistProvider(mock, &TodoistConfig{})
	p.cachePath = cachePath

	tasks, loadErr := p.LoadTasks()
	if loadErr != nil {
		t.Fatalf("LoadTasks() with cache fallback should not error, got: %v", loadErr)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 cached task, got %d", len(tasks))
	}
	if tasks[0].Text != "Cached task" {
		t.Errorf("cached task text = %q, want %q", tasks[0].Text, "Cached task")
	}
}

func TestTodoistProviderCacheWrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	mock := &mockTaskFetcher{
		tasks: []TodoistTask{
			{ID: "t1", Content: "Fresh task", ProjectID: "p1"},
		},
		projects: []TodoistProject{
			{ID: "p1", Name: "Work"},
		},
	}

	p := NewTodoistProvider(mock, &TodoistConfig{})
	p.SetCachePath(tmpDir)

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	// Verify cache was written
	cachePath := filepath.Join(tmpDir, cacheFileName)
	data, readErr := os.ReadFile(cachePath)
	if readErr != nil {
		t.Fatalf("cache file not created: %v", readErr)
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("cache unmarshal error: %v", err)
	}
	if len(entry.Tasks) != 1 {
		t.Errorf("cached %d tasks, want 1", len(entry.Tasks))
	}
}

func TestTodoistProviderSourceProviderWithoutProject(t *testing.T) {
	t.Parallel()

	mock := &mockTaskFetcher{
		tasks: []TodoistTask{
			{ID: "t1", Content: "Task", ProjectID: "unknown-proj"},
		},
		projects: []TodoistProject{}, // no projects returned
	}

	p := NewTodoistProvider(mock, &TodoistConfig{})
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	// When project name is unknown, SourceProvider should be just "todoist"
	if tasks[0].SourceProvider != "todoist" {
		t.Errorf("SourceProvider = %q, want %q", tasks[0].SourceProvider, "todoist")
	}
}

func TestTodoistProviderImplementsInterface(t *testing.T) {
	t.Parallel()

	// Compile-time check that TodoistProvider implements TaskProvider
	var _ core.TaskProvider = (*TodoistProvider)(nil)
}

func TestFactory(t *testing.T) {
	t.Parallel()

	t.Run("no settings returns error", func(t *testing.T) {
		t.Parallel()
		_, err := Factory(&core.ProviderConfig{})
		if err == nil {
			t.Error("Factory() with no settings should return error")
		}
	})

	t.Run("missing api token returns error", func(t *testing.T) {
		t.Parallel()
		config := &core.ProviderConfig{
			Providers: []core.ProviderEntry{
				{Name: "todoist", Settings: map[string]string{}},
			},
		}
		_, err := Factory(config)
		if err == nil {
			t.Error("Factory() with no API token should return error")
		}
	})
}
