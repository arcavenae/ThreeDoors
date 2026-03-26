package jira

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/adapters"
	"github.com/arcavenae/ThreeDoors/internal/core"
)

// mockSearcher implements the Searcher interface for testing.
type mockSearcher struct {
	results      []*SearchResult
	callIdx      int
	err          error
	transitions  []Transition
	transErr     error
	doTransErr   error
	doTransCalls int32 // atomic counter
}

func (m *mockSearcher) SearchJQL(_ context.Context, _ string, _ []string, _ int, _ string) (*SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.callIdx >= len(m.results) {
		return &SearchResult{IsLast: true}, nil
	}
	result := m.results[m.callIdx]
	m.callIdx++
	return result, nil
}

func (m *mockSearcher) GetTransitions(_ context.Context, _ string) ([]Transition, error) {
	if m.transErr != nil {
		return nil, m.transErr
	}
	return m.transitions, nil
}

func (m *mockSearcher) DoTransition(_ context.Context, _, _ string) error {
	atomic.AddInt32(&m.doTransCalls, 1)
	return m.doTransErr
}

// healthySearcher returns an empty result for any search (used for health check tests).
type healthySearcher struct{}

func (h *healthySearcher) SearchJQL(_ context.Context, _ string, _ []string, _ int, _ string) (*SearchResult, error) {
	return &SearchResult{IsLast: true}, nil
}

func (h *healthySearcher) GetTransitions(_ context.Context, _ string) ([]Transition, error) {
	return nil, nil
}

func (h *healthySearcher) DoTransition(_ context.Context, _, _ string) error {
	return nil
}

func TestJiraProviderContract(t *testing.T) {
	t.Parallel()
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		return NewJiraProvider(&healthySearcher{}, "project = TEST", DefaultFieldMapper())
	}
	adapters.RunContractTests(t, factory)
}

func TestName(t *testing.T) {
	t.Parallel()
	p := NewJiraProvider(&healthySearcher{}, "project = TEST", DefaultFieldMapper())
	if p.Name() != "jira" {
		t.Errorf("Name() = %q, want %q", p.Name(), "jira")
	}
}

func TestWatch(t *testing.T) {
	t.Parallel()
	p := NewJiraProvider(&healthySearcher{}, "project = TEST", DefaultFieldMapper())
	if ch := p.Watch(); ch != nil {
		t.Errorf("Watch() = %v, want nil", ch)
	}
}

func TestReadOnlyMethods(t *testing.T) {
	t.Parallel()
	p := NewJiraProvider(&healthySearcher{}, "project = TEST", DefaultFieldMapper())

	tests := []struct {
		name string
		fn   func() error
	}{
		{"SaveTask", func() error { return p.SaveTask(core.NewTask("test")) }},
		{"SaveTasks", func() error { return p.SaveTasks([]*core.Task{core.NewTask("test")}) }},
		{"DeleteTask", func() error { return p.DeleteTask("test-id") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if !errors.Is(err, core.ErrReadOnly) {
				t.Errorf("%s() error = %v, want ErrReadOnly", tt.name, err)
			}
		})
	}
}

func TestLoadTasks_SinglePage(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{
		results: []*SearchResult{
			{
				Issues: []Issue{
					makeIssue("PROJ-1", "First task", "new", "High", "PROJ", []string{"backend"}),
					makeIssue("PROJ-2", "Second task", "indeterminate", "Low", "PROJ", []string{"frontend"}),
				},
				IsLast: true,
			},
		},
	}

	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("LoadTasks() returned %d tasks, want 2", len(tasks))
	}

	task1 := findTask(tasks, "PROJ-1")
	if task1 == nil {
		t.Fatal("task PROJ-1 not found")
		return
	}
	if task1.Text != "First task" {
		t.Errorf("task1.Text = %q, want %q", task1.Text, "First task")
	}
	if task1.Status != core.StatusTodo {
		t.Errorf("task1.Status = %q, want %q", task1.Status, core.StatusTodo)
	}
	if task1.Effort != core.EffortDeepWork {
		t.Errorf("task1.Effort = %q, want %q", task1.Effort, core.EffortDeepWork)
	}
	if task1.Context != "[PROJ] backend" {
		t.Errorf("task1.Context = %q, want %q", task1.Context, "[PROJ] backend")
	}
	if task1.SourceProvider != "jira" {
		t.Errorf("task1.SourceProvider = %q, want %q", task1.SourceProvider, "jira")
	}

	task2 := findTask(tasks, "PROJ-2")
	if task2 == nil {
		t.Fatal("task PROJ-2 not found")
		return
	}
	if task2.Status != core.StatusInProgress {
		t.Errorf("task2.Status = %q, want %q", task2.Status, core.StatusInProgress)
	}
	if task2.Effort != core.EffortQuickWin {
		t.Errorf("task2.Effort = %q, want %q", task2.Effort, core.EffortQuickWin)
	}
}

func TestLoadTasks_Pagination(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{
		results: []*SearchResult{
			{
				Issues: []Issue{
					makeIssue("PROJ-1", "Page 1", "new", "Medium", "PROJ", nil),
				},
				NextPageToken: "token-2",
				IsLast:        false,
			},
			{
				Issues: []Issue{
					makeIssue("PROJ-2", "Page 2", "done", "Low", "PROJ", nil),
				},
				IsLast: true,
			},
		},
	}

	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("LoadTasks() returned %d tasks, want 2", len(tasks))
	}
}

func TestLoadTasks_SearchError(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{err: errors.New("connection refused")}
	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())

	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error, got nil")
	}
}

func TestLoadTasks_EmptyResults(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{
		results: []*SearchResult{
			{Issues: []Issue{}, IsLast: true},
		},
	}

	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("LoadTasks() returned %d tasks, want 0", len(tasks))
	}
}

func TestHealthCheck_Healthy(t *testing.T) {
	t.Parallel()
	p := NewJiraProvider(&healthySearcher{}, "project = TEST", DefaultFieldMapper())
	result := p.HealthCheck()

	if result.Overall != core.HealthOK {
		t.Errorf("HealthCheck().Overall = %q, want %q", result.Overall, core.HealthOK)
	}
	if len(result.Items) == 0 {
		t.Error("HealthCheck().Items is empty, want at least 1 item")
	}
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{err: errors.New("connection refused")}
	p := NewJiraProvider(searcher, "project = TEST", DefaultFieldMapper())
	result := p.HealthCheck()

	if result.Overall != core.HealthFail {
		t.Errorf("HealthCheck().Overall = %q, want %q", result.Overall, core.HealthFail)
	}
}

func TestMapStatus(t *testing.T) {
	t.Parallel()
	fm := DefaultFieldMapper()

	tests := []struct {
		name       string
		category   string
		wantStatus core.TaskStatus
	}{
		{"new maps to todo", "new", core.StatusTodo},
		{"indeterminate maps to in-progress", "indeterminate", core.StatusInProgress},
		{"done maps to complete", "done", core.StatusComplete},
		{"undefined maps to todo", "undefined", core.StatusTodo},
		{"unknown maps to todo", "something-random", core.StatusTodo},
		{"empty maps to todo", "", core.StatusTodo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fm.MapStatus(tt.category)
			if got != tt.wantStatus {
				t.Errorf("MapStatus(%q) = %q, want %q", tt.category, got, tt.wantStatus)
			}
		})
	}
}

func TestMapEffort(t *testing.T) {
	t.Parallel()
	fm := DefaultFieldMapper()

	tests := []struct {
		name       string
		priority   string
		wantEffort core.TaskEffort
	}{
		{"Highest maps to deep-work", "Highest", core.EffortDeepWork},
		{"High maps to deep-work", "High", core.EffortDeepWork},
		{"Medium maps to medium", "Medium", core.EffortMedium},
		{"Low maps to quick-win", "Low", core.EffortQuickWin},
		{"Lowest maps to quick-win", "Lowest", core.EffortQuickWin},
		{"unknown maps to medium", "Custom", core.EffortMedium},
		{"empty maps to medium", "", core.EffortMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fm.MapEffort(tt.priority)
			if got != tt.wantEffort {
				t.Errorf("MapEffort(%q) = %q, want %q", tt.priority, got, tt.wantEffort)
			}
		})
	}
}

func TestMapContext(t *testing.T) {
	t.Parallel()
	fm := DefaultFieldMapper()

	tests := []struct {
		name       string
		projectKey string
		labels     []string
		want       string
	}{
		{"project with labels", "PROJ", []string{"backend", "api"}, "[PROJ] backend, api"},
		{"project without labels", "PROJ", nil, "[PROJ]"},
		{"project with empty labels", "PROJ", []string{}, "[PROJ]"},
		{"project with single label", "PROJ", []string{"frontend"}, "[PROJ] frontend"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fm.MapContext(tt.projectKey, tt.labels)
			if got != tt.want {
				t.Errorf("MapContext(%q, %v) = %q, want %q", tt.projectKey, tt.labels, got, tt.want)
			}
		})
	}
}

func TestMapIssueToTask(t *testing.T) {
	t.Parallel()
	fm := DefaultFieldMapper()

	issue := makeIssue("TEST-42", "Implement feature", "indeterminate", "Highest", "TEST", []string{"epic", "v2"})
	task := fm.MapIssueToTask(issue)

	if task.ID != "TEST-42" {
		t.Errorf("task.ID = %q, want %q", task.ID, "TEST-42")
	}
	if task.Text != "Implement feature" {
		t.Errorf("task.Text = %q, want %q", task.Text, "Implement feature")
	}
	if task.Status != core.StatusInProgress {
		t.Errorf("task.Status = %q, want %q", task.Status, core.StatusInProgress)
	}
	if task.Effort != core.EffortDeepWork {
		t.Errorf("task.Effort = %q, want %q", task.Effort, core.EffortDeepWork)
	}
	if task.Context != "[TEST] epic, v2" {
		t.Errorf("task.Context = %q, want %q", task.Context, "[TEST] epic, v2")
	}
	if task.SourceProvider != "jira" {
		t.Errorf("task.SourceProvider = %q, want %q", task.SourceProvider, "jira")
	}
	if len(task.SourceRefs) != 1 || task.SourceRefs[0].NativeID != "TEST-42" || task.SourceRefs[0].Provider != "jira" {
		t.Errorf("task.SourceRefs = %v, want [{jira TEST-42}]", task.SourceRefs)
	}
}

func TestMapIssueToTask_NilPriority(t *testing.T) {
	t.Parallel()
	fm := DefaultFieldMapper()

	issue := Issue{
		Key: "TEST-1",
		Fields: IssueFields{
			Summary:  "No priority",
			Status:   IssueStatus{StatusCategory: StatusCategory{Key: "new"}},
			Priority: nil,
			Project:  IssueProject{Key: "TEST"},
		},
	}

	task := fm.MapIssueToTask(issue)
	if task.Effort != core.EffortMedium {
		t.Errorf("task.Effort = %q, want %q (default for nil priority)", task.Effort, core.EffortMedium)
	}
}

func TestFactory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings map[string]string
		wantErr  bool
	}{
		{
			name: "valid basic auth",
			settings: map[string]string{
				"url":       "https://test.atlassian.net",
				"auth_type": "basic",
				"email":     "user@test.com",
				"api_token": "token123",
				"jql":       "project = TEST",
			},
			wantErr: false,
		},
		{
			name: "valid PAT auth",
			settings: map[string]string{
				"url":       "https://jira.corp.com",
				"auth_type": "pat",
				"api_token": "pat-token",
				"jql":       "project = TEST",
			},
			wantErr: false,
		},
		{
			name: "valid with defaults",
			settings: map[string]string{
				"url":       "https://test.atlassian.net",
				"auth_type": "basic",
			},
			wantErr: false,
		},
		{
			name:     "missing url",
			settings: map[string]string{"auth_type": "basic", "api_token": "token"},
			wantErr:  true,
		},
		{
			name:     "missing auth_type",
			settings: map[string]string{"url": "https://test.atlassian.net", "api_token": "token"},
			wantErr:  true,
		},
		{
			name:     "invalid auth_type",
			settings: map[string]string{"url": "https://test.atlassian.net", "auth_type": "oauth"},
			wantErr:  true,
		},
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config := &core.ProviderConfig{
				Providers: []core.ProviderEntry{
					{Name: "jira", Settings: tt.settings},
				},
			}
			provider, err := Factory(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Factory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && provider == nil {
				t.Error("Factory() returned nil provider")
			}
		})
	}
}

// --- MarkComplete tests (Story 19.3) ---

func TestMarkComplete_TransitionDiscovery(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{
		transitions: []Transition{
			{ID: "21", Name: "In Progress", To: IssueStatus{StatusCategory: StatusCategory{Key: "indeterminate"}}},
			{ID: "31", Name: "Done", To: IssueStatus{StatusCategory: StatusCategory{Key: "done"}}},
		},
	}

	p := NewJiraProvider(searcher, "project = TEST", DefaultFieldMapper())
	p.sleepFn = func(_ time.Duration) {} // no-op sleep

	err := p.MarkComplete("TEST-1")
	if err != nil {
		t.Fatalf("MarkComplete() error: %v", err)
	}

	calls := atomic.LoadInt32(&searcher.doTransCalls)
	if calls != 1 {
		t.Errorf("DoTransition called %d times, want 1", calls)
	}
}

func TestMarkComplete_AlreadyDone(t *testing.T) {
	t.Parallel()
	// No "done" transition available means issue is already done
	searcher := &mockSearcher{
		transitions: []Transition{
			{ID: "11", Name: "Reopen", To: IssueStatus{StatusCategory: StatusCategory{Key: "new"}}},
		},
	}

	p := NewJiraProvider(searcher, "project = TEST", DefaultFieldMapper())

	err := p.MarkComplete("TEST-1")
	if err != nil {
		t.Fatalf("MarkComplete() should be no-op when already done, got error: %v", err)
	}

	calls := atomic.LoadInt32(&searcher.doTransCalls)
	if calls != 0 {
		t.Errorf("DoTransition called %d times, want 0 (issue already done)", calls)
	}
}

func TestMarkComplete_NoTransitions(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{
		transitions: []Transition{},
	}

	p := NewJiraProvider(searcher, "project = TEST", DefaultFieldMapper())

	err := p.MarkComplete("TEST-1")
	if err != nil {
		t.Fatalf("MarkComplete() should be no-op with no transitions, got error: %v", err)
	}
}

func TestMarkComplete_GetTransitionsError(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{
		transErr: errors.New("network error"),
	}

	p := NewJiraProvider(searcher, "project = TEST", DefaultFieldMapper())

	err := p.MarkComplete("TEST-1")
	if err == nil {
		t.Fatal("MarkComplete() expected error, got nil")
	}
}

func TestMarkComplete_DoTransitionError(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{
		transitions: []Transition{
			{ID: "31", Name: "Done", To: IssueStatus{StatusCategory: StatusCategory{Key: "done"}}},
		},
		doTransErr: errors.New("permission denied"),
	}

	p := NewJiraProvider(searcher, "project = TEST", DefaultFieldMapper())
	p.sleepFn = func(_ time.Duration) {}

	err := p.MarkComplete("TEST-1")
	if err == nil {
		t.Fatal("MarkComplete() expected error, got nil")
	}

	// Non-conflict error should not retry
	calls := atomic.LoadInt32(&searcher.doTransCalls)
	if calls != 1 {
		t.Errorf("DoTransition called %d times, want 1 (no retry for non-conflict)", calls)
	}
}

func TestMarkComplete_ConflictRetry(t *testing.T) {
	t.Parallel()

	// conflictThenSuccessSearcher fails with ConflictError N times, then succeeds
	callCount := int32(0)
	searcher := &mockSearcher{
		transitions: []Transition{
			{ID: "31", Name: "Done", To: IssueStatus{StatusCategory: StatusCategory{Key: "done"}}},
		},
	}

	// Override DoTransition to fail twice then succeed
	conflictSearcher := &conflictThenSuccessSearcher{
		mockSearcher:   searcher,
		failCount:      2,
		currentAttempt: &callCount,
	}

	p := NewJiraProvider(conflictSearcher, "project = TEST", DefaultFieldMapper())
	p.sleepFn = func(_ time.Duration) {} // no-op sleep

	err := p.MarkComplete("TEST-1")
	if err != nil {
		t.Fatalf("MarkComplete() error after retries: %v", err)
	}

	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("DoTransition called %d times, want 3 (2 conflicts + 1 success)", atomic.LoadInt32(&callCount))
	}
}

func TestMarkComplete_ConflictExhaustsRetries(t *testing.T) {
	t.Parallel()
	searcher := &mockSearcher{
		transitions: []Transition{
			{ID: "31", Name: "Done", To: IssueStatus{StatusCategory: StatusCategory{Key: "done"}}},
		},
		doTransErr: &ConflictError{IssueKey: "TEST-1"},
	}

	p := NewJiraProvider(searcher, "project = TEST", DefaultFieldMapper())
	p.sleepFn = func(_ time.Duration) {}

	err := p.MarkComplete("TEST-1")
	if err == nil {
		t.Fatal("MarkComplete() expected error after exhausting retries")
	}
	if !IsConflictError(err) {
		t.Errorf("expected ConflictError in chain, got: %v", err)
	}

	calls := atomic.LoadInt32(&searcher.doTransCalls)
	if calls != int32(maxConflictRetries) {
		t.Errorf("DoTransition called %d times, want %d", calls, maxConflictRetries)
	}
}

// --- Cache tests (Story 19.3) ---

func TestLoadTasks_WritesCacheOnSuccess(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	searcher := &mockSearcher{
		results: []*SearchResult{
			{
				Issues: []Issue{
					makeIssue("PROJ-1", "Cached task", "new", "Medium", "PROJ", nil),
				},
				IsLast: true,
			},
		},
	}

	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())
	p.SetCachePath(tmpDir)

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("LoadTasks() returned %d tasks, want 1", len(tasks))
	}

	// Verify cache file was written
	cachePath := filepath.Join(tmpDir, cacheFileName)
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file was not created")
	}
}

func TestLoadTasks_FallsBackToCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// First: successful load to populate cache
	searcher := &mockSearcher{
		results: []*SearchResult{
			{
				Issues: []Issue{
					makeIssue("PROJ-1", "Cached task", "new", "High", "PROJ", nil),
				},
				IsLast: true,
			},
		},
	}

	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())
	p.SetCachePath(tmpDir)

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("initial LoadTasks() error: %v", err)
	}

	// Second: API failure should use cache
	failSearcher := &mockSearcher{err: errors.New("connection refused")}
	p2 := NewJiraProvider(failSearcher, "project = PROJ", DefaultFieldMapper())
	p2.SetCachePath(tmpDir)

	tasks, err := p2.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() with cache fallback error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("LoadTasks() returned %d cached tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "PROJ-1" {
		t.Errorf("cached task ID = %q, want %q", tasks[0].ID, "PROJ-1")
	}
}

func TestLoadTasks_FailsWithoutCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	searcher := &mockSearcher{err: errors.New("connection refused")}
	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())
	p.SetCachePath(tmpDir)

	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error when API fails and no cache, got nil")
	}
}

func TestLoadTasks_FailsWithoutCachePath(t *testing.T) {
	t.Parallel()

	searcher := &mockSearcher{err: errors.New("connection refused")}
	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())
	// No SetCachePath call

	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error when API fails and no cache path, got nil")
	}
}

func TestLoadTasks_CacheNotWrittenWithoutPath(t *testing.T) {
	t.Parallel()

	searcher := &mockSearcher{
		results: []*SearchResult{
			{Issues: []Issue{makeIssue("PROJ-1", "Task", "new", "Medium", "PROJ", nil)}, IsLast: true},
		},
	}

	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())
	// No SetCachePath — cache should not be written (no panic either)

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
}

func TestCacheAtomicWrite(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	searcher := &mockSearcher{
		results: []*SearchResult{
			{Issues: []Issue{makeIssue("PROJ-1", "Task", "new", "Medium", "PROJ", nil)}, IsLast: true},
		},
	}

	p := NewJiraProvider(searcher, "project = PROJ", DefaultFieldMapper())
	p.SetCachePath(tmpDir)

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	// Verify no .tmp file remains
	tmpPath := filepath.Join(tmpDir, cacheFileName+".tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful write")
	}
}

// --- WAL wrapping test (Story 19.3 AC3) ---

func TestWALProviderWrapping(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	searcher := &mockSearcher{
		transitions: []Transition{
			{ID: "31", Name: "Done", To: IssueStatus{StatusCategory: StatusCategory{Key: "done"}}},
		},
	}

	jiraProvider := NewJiraProvider(searcher, "project = TEST", DefaultFieldMapper())
	jiraProvider.sleepFn = func(_ time.Duration) {}

	walProvider := core.NewWALProvider(jiraProvider, tmpDir)

	// MarkComplete should succeed through WAL
	err := walProvider.MarkComplete("TEST-1")
	if err != nil {
		t.Fatalf("WAL MarkComplete() error: %v", err)
	}
}

// --- FallbackProvider wrapping test (Story 19.3 AC4) ---

func TestFallbackProviderWrapping(t *testing.T) {
	t.Parallel()

	// Primary: failing Jira provider
	failSearcher := &mockSearcher{err: errors.New("connection refused")}
	primary := NewJiraProvider(failSearcher, "project = PROJ", DefaultFieldMapper())

	// Fallback: working Jira provider
	fallbackSearcher := &mockSearcher{
		results: []*SearchResult{
			{Issues: []Issue{makeIssue("PROJ-1", "Fallback task", "new", "Medium", "PROJ", nil)}, IsLast: true},
		},
	}
	fallback := NewJiraProvider(fallbackSearcher, "project = PROJ", DefaultFieldMapper())

	fp := core.NewFallbackProvider(primary, fallback)

	// Primary fails, fallback succeeds
	tasks, err := fp.LoadTasks()
	if err != nil {
		t.Fatalf("FallbackProvider LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("FallbackProvider returned %d tasks, want 1", len(tasks))
	}
}

// --- Helper types ---

// conflictThenSuccessSearcher wraps a mockSearcher but fails DoTransition N times
// with ConflictError before succeeding.
type conflictThenSuccessSearcher struct {
	*mockSearcher
	failCount      int
	currentAttempt *int32
}

func (c *conflictThenSuccessSearcher) DoTransition(ctx context.Context, issueKey, transID string) error {
	attempt := atomic.AddInt32(c.currentAttempt, 1)
	if int(attempt) <= c.failCount {
		return &ConflictError{IssueKey: issueKey}
	}
	return nil
}

// makeIssue is a test helper to build a Jira Issue with common fields.
func makeIssue(key, summary, statusCategoryKey, priorityName, projectKey string, labels []string) Issue {
	issue := Issue{
		Key: key,
		Fields: IssueFields{
			Summary: summary,
			Status: IssueStatus{
				Name:           "Some Status",
				StatusCategory: StatusCategory{Key: statusCategoryKey},
			},
			Project: IssueProject{Key: projectKey},
			Labels:  labels,
			Created: time.Now().UTC().Format(time.RFC3339),
			Updated: time.Now().UTC().Format(time.RFC3339),
		},
	}
	if priorityName != "" {
		issue.Fields.Priority = &IssuePriority{Name: priorityName}
	}
	return issue
}

// findTask searches a task slice by ID.
func findTask(tasks []*core.Task, id string) *core.Task {
	for _, t := range tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}
