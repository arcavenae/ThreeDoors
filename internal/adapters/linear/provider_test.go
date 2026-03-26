package linear

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
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
	// pages supports multi-page pagination: teamID → []IssueConnection (one per page).
	// When set, QueryTeamIssues returns pages sequentially using the cursor.
	pages map[string][]IssueConnection

	// Mutation tracking
	mutateStateIssueID string
	mutateStateStateID string
	mutateStateResult  *MutationResult
	mutateStateErr     error

	mutateUpdateIssueID string
	mutateUpdateTitle   string
	mutateUpdateDesc    string
	mutateUpdateResult  *MutationResult
	mutateUpdateErr     error
}

func (m *mockGraphQLClient) QueryViewer(_ context.Context) (*Viewer, error) {
	m.callCount++
	return m.viewer, m.viewerErr
}

func (m *mockGraphQLClient) QueryTeamIssues(_ context.Context, teamID, cursor string) (*IssueConnection, error) {
	m.callCount++
	if m.issuesErr != nil {
		return nil, m.issuesErr
	}

	// Multi-page mode
	if m.pages != nil {
		teamPages, ok := m.pages[teamID]
		if !ok {
			return &IssueConnection{}, nil
		}
		// Determine page index from cursor
		pageIdx := 0
		if cursor != "" {
			for i, p := range teamPages {
				if p.PageInfo.EndCursor == cursor && i+1 < len(teamPages) {
					pageIdx = i + 1
					break
				}
			}
		}
		if pageIdx < len(teamPages) {
			return &teamPages[pageIdx], nil
		}
		return &IssueConnection{}, nil
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

func (m *mockGraphQLClient) MutateIssueState(_ context.Context, issueID, stateID string) (*MutationResult, error) {
	m.callCount++
	m.mutateStateIssueID = issueID
	m.mutateStateStateID = stateID
	return m.mutateStateResult, m.mutateStateErr
}

func (m *mockGraphQLClient) MutateIssueUpdate(_ context.Context, issueID, title, description string) (*MutationResult, error) {
	m.callCount++
	m.mutateUpdateIssueID = issueID
	m.mutateUpdateTitle = title
	m.mutateUpdateDesc = description
	return m.mutateUpdateResult, m.mutateUpdateErr
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

	// SaveTasks and DeleteTask remain read-only (AC6)
	if err := provider.SaveTasks([]*core.Task{}); err != core.ErrReadOnly {
		t.Errorf("SaveTasks() = %v, want ErrReadOnly", err)
	}
	if err := provider.DeleteTask("x"); err != core.ErrReadOnly {
		t.Errorf("DeleteTask() = %v, want ErrReadOnly", err)
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
		name        string
		viewerErr   error
		wantOverall core.HealthStatus
	}{
		// Connectivity OK but no sync yet → WARN (sync item is WARN)
		{"healthy no sync", nil, core.HealthWarn},
		{"unhealthy", fmt.Errorf("auth failed"), core.HealthFail},
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

			if result.Overall != tt.wantOverall {
				t.Errorf("Overall = %q, want %q", result.Overall, tt.wantOverall)
			}

			// Verify connectivity item exists
			foundConn := false
			for _, item := range result.Items {
				if item.Name == "linear_connectivity" {
					foundConn = true
				}
			}
			if !foundConn {
				t.Error("HealthCheck should include linear_connectivity item")
			}

			// Verify sync item exists
			foundSync := false
			for _, item := range result.Items {
				if item.Name == "linear_sync" {
					foundSync = true
				}
			}
			if !foundSync {
				t.Error("HealthCheck should include linear_sync item")
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

// TestLinearProvider_Pagination verifies that LoadTasks fetches all pages
// via cursor-based pagination through the provider (AC6 — Story 30.4).
func TestLinearProvider_Pagination(t *testing.T) {
	t.Parallel()

	client := &mockGraphQLClient{
		pages: map[string][]IssueConnection{
			"team-1": {
				{
					Nodes:    []IssueNode{newTestIssue("1", "TEAM-1", "Page 1 Task", "started", 2)},
					PageInfo: PageInfo{HasNextPage: true, EndCursor: "cursor-page1"},
				},
				{
					Nodes:    []IssueNode{newTestIssue("2", "TEAM-2", "Page 2 Task", "unstarted", 3)},
					PageInfo: PageInfo{HasNextPage: true, EndCursor: "cursor-page2"},
				},
				{
					Nodes:    []IssueNode{newTestIssue("3", "TEAM-3", "Page 3 Task", "backlog", 4)},
					PageInfo: PageInfo{HasNextPage: false},
				},
			},
		},
	}

	provider := NewLinearProvider(client, newTestConfig())
	tasks, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3 (one from each page)", len(tasks))
	}

	// Verify all pages were fetched (3 API calls for 3 pages)
	if client.callCount != 3 {
		t.Errorf("made %d API calls, want 3 (one per page)", client.callCount)
	}
}

// newMockWithIssuesAndStates creates a mock client with issues loaded and workflow states configured.
func newMockWithIssuesAndStates() *mockGraphQLClient {
	return &mockGraphQLClient{
		viewer: &Viewer{ID: "u1", Name: "Test", Email: "test@example.com"},
		issues: map[string]*IssueConnection{
			"team-1": {
				Nodes: []IssueNode{
					{
						ID:         "uuid-issue-1",
						Identifier: "TEAM-1",
						Title:      "Task One",
						Priority:   2,
						State:      IssueState{ID: "s4", Name: "In Progress", Type: "started"},
						Team:       IssueTeam{ID: "team-1", Key: "TEAM"},
						Labels:     IssueLabels{},
						CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				},
				PageInfo: PageInfo{HasNextPage: false},
			},
		},
		states: map[string][]WorkflowState{
			"team-1": {
				{ID: "s1", Name: "Triage", Type: "triage"},
				{ID: "s3", Name: "Todo", Type: "unstarted"},
				{ID: "s4", Name: "In Progress", Type: "started"},
				{ID: "s5", Name: "Done", Type: "completed"},
				{ID: "s6", Name: "Cancelled", Type: "cancelled"},
			},
		},
		mutateStateResult:  &MutationResult{Success: true, Issue: &MutedIssue{ID: "uuid-issue-1", State: IssueState{ID: "s5", Name: "Done", Type: "completed"}}},
		mutateUpdateResult: &MutationResult{Success: true, Issue: &MutedIssue{ID: "uuid-issue-1", State: IssueState{ID: "s4", Name: "In Progress", Type: "started"}}},
	}
}

func TestLinearProvider_MarkComplete(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())

	// Load tasks first to populate issue index
	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	// AC1: MarkComplete sends mutation to transition to completed state
	err = provider.MarkComplete("linear:TEAM-1")
	if err != nil {
		t.Fatalf("MarkComplete() error = %v", err)
	}

	// Verify correct issue UUID was used
	if client.mutateStateIssueID != "uuid-issue-1" {
		t.Errorf("mutation issueID = %q, want %q", client.mutateStateIssueID, "uuid-issue-1")
	}

	// AC2: Verify the completed state ID was discovered dynamically
	if client.mutateStateStateID != "s5" {
		t.Errorf("mutation stateID = %q, want %q (completed state)", client.mutateStateStateID, "s5")
	}
}

func TestLinearProvider_MarkComplete_CachesStateID(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	// Add a second issue
	conn := client.issues["team-1"]
	conn.Nodes = append(conn.Nodes, IssueNode{
		ID:         "uuid-issue-2",
		Identifier: "TEAM-2",
		Title:      "Task Two",
		Priority:   3,
		State:      IssueState{ID: "s4", Name: "In Progress", Type: "started"},
		Team:       IssueTeam{ID: "team-1", Key: "TEAM"},
		Labels:     IssueLabels{},
		CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	})

	provider := NewLinearProvider(client, newTestConfig())
	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	callsBefore := client.callCount

	// First MarkComplete — queries workflow states
	if err := provider.MarkComplete("linear:TEAM-1"); err != nil {
		t.Fatalf("first MarkComplete() error = %v", err)
	}
	callsAfterFirst := client.callCount

	// Second MarkComplete — should use cached state
	if err := provider.MarkComplete("linear:TEAM-2"); err != nil {
		t.Fatalf("second MarkComplete() error = %v", err)
	}
	callsAfterSecond := client.callCount

	// First call: 1 QueryWorkflowStates + 1 MutateIssueState = 2 calls
	firstCalls := callsAfterFirst - callsBefore
	if firstCalls != 2 {
		t.Errorf("first MarkComplete made %d API calls, want 2 (state query + mutation)", firstCalls)
	}

	// Second call: only 1 MutateIssueState (state cached)
	secondCalls := callsAfterSecond - callsAfterFirst
	if secondCalls != 1 {
		t.Errorf("second MarkComplete made %d API calls, want 1 (cached state, only mutation)", secondCalls)
	}
}

func TestLinearProvider_MarkComplete_UnknownTask(t *testing.T) {
	t.Parallel()

	provider := NewLinearProvider(&mockGraphQLClient{}, newTestConfig())

	err := provider.MarkComplete("linear:UNKNOWN-999")
	if err == nil {
		t.Fatal("MarkComplete() should fail for unknown task ID")
	}
}

func TestLinearProvider_MarkComplete_MutationFailure(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	client.mutateStateErr = fmt.Errorf("network error")

	provider := NewLinearProvider(client, newTestConfig())
	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	// AC5: Failed mutations are logged with task ID and error
	err = provider.MarkComplete("linear:TEAM-1")
	if err == nil {
		t.Fatal("MarkComplete() should fail when mutation fails")
	}
	if !contains(err.Error(), "linear:TEAM-1") {
		t.Errorf("error should contain task ID, got %q", err.Error())
	}
}

func TestLinearProvider_MarkComplete_NoCompletedState(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	// Remove the completed state
	client.states["team-1"] = []WorkflowState{
		{ID: "s1", Name: "Triage", Type: "triage"},
		{ID: "s4", Name: "In Progress", Type: "started"},
	}

	provider := NewLinearProvider(client, newTestConfig())
	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	err = provider.MarkComplete("linear:TEAM-1")
	if err == nil {
		t.Fatal("MarkComplete() should fail when no completed state exists")
	}
	if !contains(err.Error(), "no completed workflow state") {
		t.Errorf("error should mention missing state, got %q", err.Error())
	}
}

func TestLinearProvider_MarkComplete_SuccessFalse(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	client.mutateStateResult = &MutationResult{Success: false}

	provider := NewLinearProvider(client, newTestConfig())
	_, _ = provider.LoadTasks()

	err := provider.MarkComplete("linear:TEAM-1")
	if err == nil {
		t.Fatal("MarkComplete() should fail when mutation returns success=false")
	}
	if !contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got %q", err.Error())
	}
}

func TestLinearProvider_SaveTask(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())

	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	// AC4: SaveTask updates title and description
	task := &core.Task{
		ID:      "linear:TEAM-1",
		Text:    "Updated Title",
		Context: "Updated description",
	}
	err = provider.SaveTask(task)
	if err != nil {
		t.Fatalf("SaveTask() error = %v", err)
	}

	if client.mutateUpdateIssueID != "uuid-issue-1" {
		t.Errorf("mutation issueID = %q, want %q", client.mutateUpdateIssueID, "uuid-issue-1")
	}
	if client.mutateUpdateTitle != "Updated Title" {
		t.Errorf("mutation title = %q, want %q", client.mutateUpdateTitle, "Updated Title")
	}
	if client.mutateUpdateDesc != "Updated description" {
		t.Errorf("mutation description = %q, want %q", client.mutateUpdateDesc, "Updated description")
	}
}

func TestLinearProvider_SaveTask_UnknownTask(t *testing.T) {
	t.Parallel()

	provider := NewLinearProvider(&mockGraphQLClient{}, newTestConfig())

	err := provider.SaveTask(&core.Task{ID: "linear:UNKNOWN-999"})
	if err == nil {
		t.Fatal("SaveTask() should fail for unknown task ID")
	}
}

func TestLinearProvider_SaveTask_MutationFailure(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	client.mutateUpdateErr = fmt.Errorf("network error")

	provider := NewLinearProvider(client, newTestConfig())
	_, _ = provider.LoadTasks()

	err := provider.SaveTask(&core.Task{ID: "linear:TEAM-1", Text: "Updated"})
	if err == nil {
		t.Fatal("SaveTask() should fail when mutation fails")
	}
}

func TestLinearProvider_HealthCheck_SyncStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		doSync          bool
		viewerErr       error
		wantOverall     core.HealthStatus
		wantSyncMessage string
	}{
		{
			name:            "no sync yet",
			doSync:          false,
			wantOverall:     core.HealthWarn,
			wantSyncMessage: "No successful sync recorded",
		},
		{
			name:            "after successful sync",
			doSync:          true,
			wantOverall:     core.HealthOK,
			wantSyncMessage: "Last successful sync:",
		},
		{
			name:            "api unreachable no sync",
			doSync:          false,
			viewerErr:       fmt.Errorf("connection refused"),
			wantOverall:     core.HealthFail,
			wantSyncMessage: "No successful sync recorded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newMockWithIssuesAndStates()
			client.viewerErr = tt.viewerErr
			provider := NewLinearProvider(client, newTestConfig())

			if tt.doSync {
				// Load tasks then complete one to set lastSyncSuccess
				client.viewerErr = nil
				_, _ = provider.LoadTasks()
				_ = provider.MarkComplete("linear:TEAM-1")
				// Restore viewerErr for health check
				client.viewerErr = tt.viewerErr
			}

			result := provider.HealthCheck()

			if result.Overall != tt.wantOverall {
				t.Errorf("Overall = %q, want %q", result.Overall, tt.wantOverall)
			}

			// Check sync item exists
			found := false
			for _, item := range result.Items {
				if item.Name == "linear_sync" {
					found = true
					if !contains(item.Message, tt.wantSyncMessage) {
						t.Errorf("sync message = %q, want to contain %q", item.Message, tt.wantSyncMessage)
					}
				}
			}
			if !found {
				t.Error("HealthCheck should include linear_sync item")
			}
		})
	}
}

func TestLinearProvider_IssueIndexPopulatedOnLoad(t *testing.T) {
	t.Parallel()

	client := newMockWithIssuesAndStates()
	provider := NewLinearProvider(client, newTestConfig())

	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	provider.mu.RLock()
	info, ok := provider.issueIndex["linear:TEAM-1"]
	provider.mu.RUnlock()

	if !ok {
		t.Fatal("issueIndex should contain linear:TEAM-1 after LoadTasks")
	}
	if info.issueID != "uuid-issue-1" {
		t.Errorf("issueID = %q, want %q", info.issueID, "uuid-issue-1")
	}
	if info.teamID != "team-1" {
		t.Errorf("teamID = %q, want %q", info.teamID, "team-1")
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
