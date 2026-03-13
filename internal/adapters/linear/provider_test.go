package linear

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// mockGraphQLClient is a test double for GraphQLClient.
type mockGraphQLClient struct {
	viewer    *Viewer
	viewerErr error
	issues    map[string]*IssueConnection // teamID → connection
	issuesErr error
	states    map[string][]WorkflowState
	statesErr error
	callCount int
}

func (m *mockGraphQLClient) QueryViewer(_ context.Context) (*Viewer, error) {
	m.callCount++
	return m.viewer, m.viewerErr
}

func (m *mockGraphQLClient) QueryTeamIssues(_ context.Context, teamID, _ string) (*IssueConnection, error) {
	m.callCount++
	if m.issuesErr != nil {
		return nil, m.issuesErr
	}
	conn, ok := m.issues[teamID]
	if !ok {
		return &IssueConnection{}, nil
	}
	return conn, nil
}

func (m *mockGraphQLClient) QueryWorkflowStates(_ context.Context, teamID string) ([]WorkflowState, error) {
	m.callCount++
	if m.statesErr != nil {
		return nil, m.statesErr
	}
	return m.states[teamID], nil
}

func newTestConfig() *LinearConfig {
	return &LinearConfig{
		APIKey:       "test-key",
		TeamIDs:      []string{"team-1"},
		PollInterval: 5 * time.Minute,
	}
}

func newTestIssue(id, identifier, title, stateType string, priority int) IssueNode {
	return IssueNode{
		ID:         id,
		Identifier: identifier,
		Title:      title,
		Priority:   priority,
		State:      IssueState{ID: "s1", Name: stateType, Type: stateType},
		Team:       IssueTeam{ID: "team-1", Key: "TEAM"},
		Labels:     IssueLabels{},
		CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
}

func TestMapStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stateType string
		want      core.TaskStatus
	}{
		{"triage maps to todo", "triage", core.StatusTodo},
		{"backlog maps to todo", "backlog", core.StatusTodo},
		{"unstarted maps to todo", "unstarted", core.StatusTodo},
		{"started maps to in-progress", "started", core.StatusInProgress},
		{"completed maps to complete", "completed", core.StatusComplete},
		{"cancelled maps to archived", "cancelled", core.StatusArchived},
		{"unknown maps to todo", "custom-state", core.StatusTodo},
		{"empty maps to todo", "", core.StatusTodo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MapStatus(tt.stateType)
			if got != tt.want {
				t.Errorf("MapStatus(%q) = %q, want %q", tt.stateType, got, tt.want)
			}
		})
	}
}

func TestMapEffort(t *testing.T) {
	t.Parallel()

	est1 := 1.0
	est2 := 2.0
	est5 := 5.0

	tests := []struct {
		name     string
		priority int
		estimate *float64
		want     core.TaskEffort
	}{
		{"urgent priority", 1, nil, core.EffortDeepWork},
		{"high priority", 2, nil, core.EffortDeepWork},
		{"medium priority", 3, nil, core.EffortMedium},
		{"low priority", 4, nil, core.EffortQuickWin},
		{"no priority no estimate", 0, nil, core.EffortMedium},
		{"no priority small estimate", 0, &est1, core.EffortQuickWin},
		{"no priority medium estimate", 0, &est2, core.EffortMedium},
		{"no priority large estimate", 0, &est5, core.EffortDeepWork},
		{"unknown priority", 99, nil, core.EffortMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MapEffort(tt.priority, tt.estimate)
			if got != tt.want {
				t.Errorf("MapEffort(%d, %v) = %q, want %q", tt.priority, tt.estimate, got, tt.want)
			}
		})
	}
}

func TestMapIssueToTask_FieldMapping(t *testing.T) {
	t.Parallel()

	dueDate := "2026-03-15"
	issue := IssueNode{
		ID:          "issue-1",
		Identifier:  "TEAM-123",
		Title:       "Fix the bug",
		Description: "This is a **markdown** description",
		Priority:    2,
		DueDate:     &dueDate,
		State:       IssueState{ID: "s1", Name: "In Progress", Type: "started"},
		Team:        IssueTeam{ID: "team-1", Key: "TEAM"},
		Labels: IssueLabels{
			Nodes: []IssueLabel{
				{Name: "bug"},
				{Name: "frontend"},
			},
		},
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	task := mapIssueToTask(&issue)

	// AC2: title → Text
	if task.Text != "Fix the bug" {
		t.Errorf("Text = %q, want %q", task.Text, "Fix the bug")
	}

	// AC2: identifier → ID with linear: prefix
	if task.ID != "linear:TEAM-123" {
		t.Errorf("ID = %q, want %q", task.ID, "linear:TEAM-123")
	}

	// AC2: description → Context (Markdown preserved)
	if task.Context == "" {
		t.Fatal("Context should not be empty")
	}
	if !contains(task.Context, "**markdown**") {
		t.Errorf("Context should preserve markdown, got %q", task.Context)
	}

	// AC2: labels → in Context
	if !contains(task.Context, "Labels: bug, frontend") {
		t.Errorf("Context should contain labels, got %q", task.Context)
	}

	// AC2: dueDate → in Context
	if !contains(task.Context, "Due: 2026-03-15") {
		t.Errorf("Context should contain due date, got %q", task.Context)
	}

	// AC3: started → in-progress
	if task.Status != core.StatusInProgress {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusInProgress)
	}

	// AC4: priority 2 (high) → deep-work
	if task.Effort != core.EffortDeepWork {
		t.Errorf("Effort = %q, want %q", task.Effort, core.EffortDeepWork)
	}

	// SourceProvider
	if task.SourceProvider != "linear" {
		t.Errorf("SourceProvider = %q, want %q", task.SourceProvider, "linear")
	}

	// SourceRefs
	if len(task.SourceRefs) != 1 || task.SourceRefs[0].Provider != "linear" {
		t.Errorf("SourceRefs = %v, want [{linear linear:TEAM-123}]", task.SourceRefs)
	}

	// Location
	if task.Location != core.TaskLocation("TEAM-123") {
		t.Errorf("Location = %q, want %q", task.Location, "TEAM-123")
	}
}

func TestMapIssueToTask_MinimalFields(t *testing.T) {
	t.Parallel()

	issue := IssueNode{
		ID:         "issue-2",
		Identifier: "TEAM-456",
		Title:      "Simple task",
		Priority:   0,
		State:      IssueState{Type: "unstarted"},
		Team:       IssueTeam{ID: "team-1", Key: "TEAM"},
		Labels:     IssueLabels{},
		CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	task := mapIssueToTask(&issue)

	if task.Context != "" {
		t.Errorf("Context should be empty for issue with no description/labels/due date, got %q", task.Context)
	}
	if task.Status != core.StatusTodo {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusTodo)
	}
	if task.Effort != core.EffortMedium {
		t.Errorf("Effort = %q, want %q", task.Effort, core.EffortMedium)
	}
}

func TestLinearProvider_LoadTasks(t *testing.T) {
	t.Parallel()

	client := &mockGraphQLClient{
		issues: map[string]*IssueConnection{
			"team-1": {
				Nodes: []IssueNode{
					newTestIssue("1", "TEAM-1", "Task One", "started", 2),
					newTestIssue("2", "TEAM-2", "Task Two", "unstarted", 4),
				},
				PageInfo: PageInfo{HasNextPage: false},
			},
		},
	}

	provider := NewLinearProvider(client, newTestConfig())
	tasks, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}

	if tasks[0].Text != "Task One" {
		t.Errorf("tasks[0].Text = %q, want %q", tasks[0].Text, "Task One")
	}
	if tasks[0].Status != core.StatusInProgress {
		t.Errorf("tasks[0].Status = %q, want %q", tasks[0].Status, core.StatusInProgress)
	}
	if tasks[1].Effort != core.EffortQuickWin {
		t.Errorf("tasks[1].Effort = %q, want %q", tasks[1].Effort, core.EffortQuickWin)
	}
}

func TestLinearProvider_LoadTasks_MultipleTeams(t *testing.T) {
	t.Parallel()

	client := &mockGraphQLClient{
		issues: map[string]*IssueConnection{
			"team-1": {
				Nodes:    []IssueNode{newTestIssue("1", "A-1", "Task A", "started", 1)},
				PageInfo: PageInfo{HasNextPage: false},
			},
			"team-2": {
				Nodes:    []IssueNode{newTestIssue("2", "B-1", "Task B", "backlog", 3)},
				PageInfo: PageInfo{HasNextPage: false},
			},
		},
	}

	config := &LinearConfig{
		APIKey:       "test-key",
		TeamIDs:      []string{"team-1", "team-2"},
		PollInterval: 5 * time.Minute,
	}

	provider := NewLinearProvider(client, config)
	tasks, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
}

func TestLinearProvider_AssigneeFiltering(t *testing.T) {
	t.Parallel()

	assignee := &IssueUser{ID: "u1", Name: "Alice", Email: "alice@example.com"}
	other := &IssueUser{ID: "u2", Name: "Bob", Email: "bob@example.com"}

	issue1 := newTestIssue("1", "TEAM-1", "Alice's task", "started", 2)
	issue1.Assignee = assignee

	issue2 := newTestIssue("2", "TEAM-2", "Bob's task", "started", 3)
	issue2.Assignee = other

	issue3 := newTestIssue("3", "TEAM-3", "Unassigned", "backlog", 4)

	client := &mockGraphQLClient{
		issues: map[string]*IssueConnection{
			"team-1": {
				Nodes:    []IssueNode{issue1, issue2, issue3},
				PageInfo: PageInfo{HasNextPage: false},
			},
		},
	}

	config := &LinearConfig{
		APIKey:       "test-key",
		TeamIDs:      []string{"team-1"},
		Assignee:     "alice@example.com",
		PollInterval: 5 * time.Minute,
	}

	provider := NewLinearProvider(client, config)
	tasks, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1 (only Alice's)", len(tasks))
	}
	if tasks[0].Text != "Alice's task" {
		t.Errorf("got %q, want Alice's task", tasks[0].Text)
	}
}

func TestLinearProvider_ReadOnly(t *testing.T) {
	t.Parallel()

	provider := NewLinearProvider(&mockGraphQLClient{}, newTestConfig())

	if err := provider.SaveTask(&core.Task{}); err != core.ErrReadOnly {
		t.Errorf("SaveTask() = %v, want ErrReadOnly", err)
	}
	if err := provider.SaveTasks([]*core.Task{}); err != core.ErrReadOnly {
		t.Errorf("SaveTasks() = %v, want ErrReadOnly", err)
	}
	if err := provider.DeleteTask("x"); err != core.ErrReadOnly {
		t.Errorf("DeleteTask() = %v, want ErrReadOnly", err)
	}
	if err := provider.MarkComplete("x"); err != core.ErrReadOnly {
		t.Errorf("MarkComplete() = %v, want ErrReadOnly", err)
	}
}

func TestLinearProvider_Name(t *testing.T) {
	t.Parallel()

	provider := NewLinearProvider(&mockGraphQLClient{}, newTestConfig())
	if got := provider.Name(); got != "linear" {
		t.Errorf("Name() = %q, want %q", got, "linear")
	}
}

func TestLinearProvider_CacheTTL(t *testing.T) {
	t.Parallel()

	client := &mockGraphQLClient{
		issues: map[string]*IssueConnection{
			"team-1": {
				Nodes:    []IssueNode{newTestIssue("1", "TEAM-1", "Task", "started", 2)},
				PageInfo: PageInfo{HasNextPage: false},
			},
		},
	}

	provider := NewLinearProvider(client, newTestConfig())

	// First call hits API
	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("first LoadTasks() error = %v", err)
	}
	firstCallCount := client.callCount

	// Second call should use cache
	_, err = provider.LoadTasks()
	if err != nil {
		t.Fatalf("second LoadTasks() error = %v", err)
	}

	if client.callCount != firstCallCount {
		t.Errorf("expected cache hit, but API was called again (callCount went from %d to %d)",
			firstCallCount, client.callCount)
	}
}

func TestLinearProvider_DiskCacheFallback(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Write a cache file with known tasks
	entry := cacheEntry{
		LastUpdated: time.Now().UTC(),
		Tasks: []*core.Task{
			{
				ID: "cached-1", Text: "Cached task", Status: core.StatusTodo,
				CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
			},
		},
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(tmpDir, cacheFileName)
	if err := os.WriteFile(cachePath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Provider with failing API
	client := &mockGraphQLClient{
		issuesErr: fmt.Errorf("network error"),
	}
	provider := NewLinearProvider(client, newTestConfig())
	provider.cachePath = cachePath

	tasks, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() should use cache fallback, got error: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "cached-1" {
		t.Errorf("expected cached task, got %v", tasks)
	}
}

func TestLinearProvider_HealthCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		viewerErr error
		wantOK    bool
	}{
		{"healthy", nil, true},
		{"unhealthy", fmt.Errorf("auth failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := &mockGraphQLClient{
				viewer:    &Viewer{ID: "u1", Name: "Test", Email: "test@example.com"},
				viewerErr: tt.viewerErr,
			}
			provider := NewLinearProvider(client, newTestConfig())
			result := provider.HealthCheck()

			if tt.wantOK && result.Overall != core.HealthOK {
				t.Errorf("Overall = %q, want OK", result.Overall)
			}
			if !tt.wantOK && result.Overall != core.HealthFail {
				t.Errorf("Overall = %q, want FAIL", result.Overall)
			}
		})
	}
}

func TestFactory(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "test-key-from-env")

	config := &core.ProviderConfig{
		Providers: []core.ProviderEntry{
			{
				Name: "linear",
				Settings: map[string]string{
					"team_ids": "team-1,team-2",
				},
			},
		},
	}

	provider, err := Factory(config)
	if err != nil {
		t.Fatalf("Factory() error = %v", err)
	}
	if provider.Name() != "linear" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "linear")
	}
}

func TestFactory_NoSettings(t *testing.T) {
	t.Parallel()

	config := &core.ProviderConfig{}
	_, err := Factory(config)
	if err == nil {
		t.Fatal("Factory() should fail with no settings")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
