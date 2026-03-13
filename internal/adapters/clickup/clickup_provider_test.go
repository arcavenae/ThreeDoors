package clickup

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/adapters"
	"github.com/arcaven/ThreeDoors/internal/core"
)

// mockTaskFetcher implements TaskFetcher for unit testing.
type mockTaskFetcher struct {
	tasksByList   map[string][]ClickUpTask // listID -> tasks
	tasksErr      error
	user          *ClickUpUser
	userErr       error
	getTasksCalls int
	getUserCalls  int
}

func (m *mockTaskFetcher) GetTasks(_ context.Context, listID string, _ int) ([]ClickUpTask, error) {
	m.getTasksCalls++
	if m.tasksErr != nil {
		return nil, m.tasksErr
	}
	return m.tasksByList[listID], nil
}

func (m *mockTaskFetcher) GetAuthorizedUser(_ context.Context) (*ClickUpUser, error) {
	m.getUserCalls++
	if m.userErr != nil {
		return nil, m.userErr
	}
	return m.user, nil
}

func TestClickUpProviderName(t *testing.T) {
	t.Parallel()
	p := NewClickUpProvider(&mockTaskFetcher{}, &ClickUpConfig{})
	if got := p.Name(); got != "clickup" {
		t.Errorf("Name() = %q, want %q", got, "clickup")
	}
}

func TestClickUpProviderImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ core.TaskProvider = (*ClickUpProvider)(nil)
}

func TestClickUpProviderLoadTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mock       *mockTaskFetcher
		cfg        *ClickUpConfig
		wantCount  int
		wantErr    bool
		wantText   string
		wantStatus core.TaskStatus
		wantEffort core.TaskEffort
	}{
		{
			name: "basic task mapping",
			mock: &mockTaskFetcher{
				tasksByList: map[string][]ClickUpTask{
					"list1": {
						{
							ID:   "task-1",
							Name: "Fix login bug",
							Status: ClickUpStatus{
								Status: "in progress",
								Type:   "custom",
							},
							Priority: &ClickUpPriority{ID: "3"},
							URL:      "https://app.clickup.com/t/task-1",
							List:     ClickUpListRef{ID: "list1", Name: "Sprint 1"},
						},
					},
				},
				user: &ClickUpUser{ID: 1},
			},
			cfg:        &ClickUpConfig{ListIDs: []string{"list1"}},
			wantCount:  1,
			wantText:   "Fix login bug",
			wantStatus: core.StatusInProgress,
			wantEffort: core.EffortMedium,
		},
		{
			name: "empty task list returns empty slice",
			mock: &mockTaskFetcher{
				tasksByList: map[string][]ClickUpTask{
					"list1": {},
				},
			},
			cfg:       &ClickUpConfig{ListIDs: []string{"list1"}},
			wantCount: 0,
		},
		{
			name: "API error propagates",
			mock: &mockTaskFetcher{
				tasksErr: errors.New("network error"),
			},
			cfg:     &ClickUpConfig{ListIDs: []string{"list1"}},
			wantErr: true,
		},
		{
			name: "completed task maps to StatusComplete",
			mock: &mockTaskFetcher{
				tasksByList: map[string][]ClickUpTask{
					"list1": {
						{
							ID:     "task-2",
							Name:   "Done task",
							Status: ClickUpStatus{Status: "complete"},
						},
					},
				},
			},
			cfg:        &ClickUpConfig{ListIDs: []string{"list1"}},
			wantCount:  1,
			wantStatus: core.StatusComplete,
		},
		{
			name: "closed task maps to StatusComplete",
			mock: &mockTaskFetcher{
				tasksByList: map[string][]ClickUpTask{
					"list1": {
						{
							ID:     "task-3",
							Name:   "Closed task",
							Status: ClickUpStatus{Status: "closed"},
						},
					},
				},
			},
			cfg:        &ClickUpConfig{ListIDs: []string{"list1"}},
			wantCount:  1,
			wantStatus: core.StatusComplete,
		},
		{
			name: "unknown status maps to StatusTodo",
			mock: &mockTaskFetcher{
				tasksByList: map[string][]ClickUpTask{
					"list1": {
						{
							ID:     "task-4",
							Name:   "Custom status task",
							Status: ClickUpStatus{Status: "awaiting review"},
						},
					},
				},
			},
			cfg:        &ClickUpConfig{ListIDs: []string{"list1"}},
			wantCount:  1,
			wantStatus: core.StatusTodo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := NewClickUpProvider(tt.mock, tt.cfg)
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
			if tt.wantStatus != "" && task.Status != tt.wantStatus {
				t.Errorf("task.Status = %q, want %q", task.Status, tt.wantStatus)
			}
			if tt.wantEffort != "" && task.Effort != tt.wantEffort {
				t.Errorf("task.Effort = %q, want %q", task.Effort, tt.wantEffort)
			}
		})
	}
}

func TestClickUpProviderFieldMapping(t *testing.T) {
	t.Parallel()

	mock := &mockTaskFetcher{
		tasksByList: map[string][]ClickUpTask{
			"list1": {
				{
					ID:          "task-100",
					Name:        "Implement feature X",
					Description: "Detailed description of feature X",
					Status:      ClickUpStatus{Status: "to do"},
					Priority:    &ClickUpPriority{ID: "2"},
					DueDate:     "1735689600000", // 2025-01-01T00:00:00Z
					Tags: []ClickUpTag{
						{Name: "frontend"},
						{Name: "urgent"},
					},
					URL:  "https://app.clickup.com/t/task-100",
					List: ClickUpListRef{ID: "list1", Name: "Backlog"},
				},
			},
		},
	}

	p := NewClickUpProvider(mock, &ClickUpConfig{ListIDs: []string{"list1"}})
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]

	// AC4: name → Text
	if task.Text != "Implement feature X" {
		t.Errorf("Text = %q, want %q", task.Text, "Implement feature X")
	}

	// AC4: description + tags → Context
	if task.Context != "Detailed description of feature X | frontend, urgent" {
		t.Errorf("Context = %q, want %q", task.Context, "Detailed description of feature X | frontend, urgent")
	}

	// AC4: status mapping
	if task.Status != core.StatusTodo {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusTodo)
	}

	// AC5: priority mapping (2=High → deep-work)
	if task.Effort != core.EffortDeepWork {
		t.Errorf("Effort = %q, want %q", task.Effort, core.EffortDeepWork)
	}

	// AC7: due date mapping
	if task.DeferUntil == nil {
		t.Fatal("DeferUntil is nil, expected non-nil")
	}
	expectedDue := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !task.DeferUntil.Equal(expectedDue) {
		t.Errorf("DeferUntil = %v, want %v", task.DeferUntil, expectedDue)
	}

	// AC6: SourceRef populated
	if len(task.SourceRefs) != 1 {
		t.Fatalf("SourceRefs length = %d, want 1", len(task.SourceRefs))
	}
	if task.SourceRefs[0].Provider != "clickup" {
		t.Errorf("SourceRefs[0].Provider = %q, want %q", task.SourceRefs[0].Provider, "clickup")
	}
	// NativeID includes task ID and list ID
	if task.SourceRefs[0].NativeID != "task-100:list1" {
		t.Errorf("SourceRefs[0].NativeID = %q, want %q", task.SourceRefs[0].NativeID, "task-100:list1")
	}

	// AC6: SourceProvider includes URL for back-linking
	if task.SourceProvider != "clickup:https://app.clickup.com/t/task-100" {
		t.Errorf("SourceProvider = %q, want %q", task.SourceProvider, "clickup:https://app.clickup.com/t/task-100")
	}
}

func TestClickUpProviderTagsAsContext(t *testing.T) {
	t.Parallel()

	mock := &mockTaskFetcher{
		tasksByList: map[string][]ClickUpTask{
			"list1": {
				{
					ID:     "t1",
					Name:   "No description task",
					Status: ClickUpStatus{Status: "to do"},
					Tags: []ClickUpTag{
						{Name: "backend"},
						{Name: "api"},
					},
					List: ClickUpListRef{ID: "list1"},
				},
			},
		},
	}

	p := NewClickUpProvider(mock, &ClickUpConfig{ListIDs: []string{"list1"}})
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	// Tags mapped to context when no description
	if tasks[0].Context != "backend, api" {
		t.Errorf("Context = %q, want %q", tasks[0].Context, "backend, api")
	}
}

func TestClickUpProviderPriorityMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		priority   *ClickUpPriority
		wantEffort core.TaskEffort
	}{
		{"nil priority (none)", nil, core.EffortQuickWin},
		{"priority 1 (Urgent)", &ClickUpPriority{ID: "1"}, core.EffortDeepWork},
		{"priority 2 (High)", &ClickUpPriority{ID: "2"}, core.EffortDeepWork},
		{"priority 3 (Normal)", &ClickUpPriority{ID: "3"}, core.EffortMedium},
		{"priority 4 (Low)", &ClickUpPriority{ID: "4"}, core.EffortQuickWin},
		{"unknown priority", &ClickUpPriority{ID: "99"}, core.EffortQuickWin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MapPriority(tt.priority)
			if got != tt.wantEffort {
				t.Errorf("MapPriority(%v) = %q, want %q", tt.priority, got, tt.wantEffort)
			}
		})
	}
}

func TestClickUpProviderStatusMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     string
		wantStatus core.TaskStatus
	}{
		{"to do", "to do", core.StatusTodo},
		{"open", "open", core.StatusTodo},
		{"in progress", "in progress", core.StatusInProgress},
		{"complete", "complete", core.StatusComplete},
		{"closed", "closed", core.StatusComplete},
		{"done", "done", core.StatusComplete},
		{"unknown status", "custom_status", core.StatusTodo},
		{"case insensitive", "To Do", core.StatusTodo},
		{"with whitespace", "  in progress  ", core.StatusInProgress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := NewClickUpProvider(&mockTaskFetcher{}, &ClickUpConfig{})
			got := p.mapStatus(tt.status)
			if got != tt.wantStatus {
				t.Errorf("mapStatus(%q) = %q, want %q", tt.status, got, tt.wantStatus)
			}
		})
	}
}

func TestClickUpProviderCustomStatusMapping(t *testing.T) {
	t.Parallel()

	customMapping := map[string]core.TaskStatus{
		"awaiting review": core.StatusInReview,
		"blocked":         core.StatusBlocked,
		"backlog":         core.StatusTodo,
	}

	mock := &mockTaskFetcher{
		tasksByList: map[string][]ClickUpTask{
			"list1": {
				{
					ID:     "t1",
					Name:   "Review task",
					Status: ClickUpStatus{Status: "awaiting review"},
					List:   ClickUpListRef{ID: "list1"},
				},
			},
		},
	}

	p := NewClickUpProvider(mock, &ClickUpConfig{
		ListIDs:       []string{"list1"},
		StatusMapping: customMapping,
	})
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if tasks[0].Status != core.StatusInReview {
		t.Errorf("Status = %q, want %q", tasks[0].Status, core.StatusInReview)
	}
}

func TestClickUpProviderDueDateConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dueDate string
		wantNil bool
		wantUTC time.Time
	}{
		{
			name:    "valid timestamp",
			dueDate: "1735689600000", // 2025-01-01T00:00:00Z
			wantUTC: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "timestamp with milliseconds",
			dueDate: "1735689600500", // 2025-01-01T00:00:00.500Z
			wantUTC: time.Date(2025, 1, 1, 0, 0, 0, 500000000, time.UTC),
		},
		{
			name:    "empty due date",
			dueDate: "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockTaskFetcher{
				tasksByList: map[string][]ClickUpTask{
					"list1": {
						{
							ID:      "t1",
							Name:    "Due date task",
							Status:  ClickUpStatus{Status: "to do"},
							DueDate: tt.dueDate,
							List:    ClickUpListRef{ID: "list1"},
						},
					},
				},
			}

			p := NewClickUpProvider(mock, &ClickUpConfig{ListIDs: []string{"list1"}})
			tasks, err := p.LoadTasks()
			if err != nil {
				t.Fatalf("LoadTasks() error = %v", err)
			}

			if tt.wantNil {
				if tasks[0].DeferUntil != nil {
					t.Errorf("DeferUntil = %v, want nil", tasks[0].DeferUntil)
				}
				return
			}

			if tasks[0].DeferUntil == nil {
				t.Fatal("DeferUntil is nil, want non-nil")
			}
			if !tasks[0].DeferUntil.Equal(tt.wantUTC) {
				t.Errorf("DeferUntil = %v, want %v", tasks[0].DeferUntil, tt.wantUTC)
			}
		})
	}
}

func TestClickUpProviderGracefulSkip(t *testing.T) {
	t.Parallel()

	mock := &mockTaskFetcher{
		tasksByList: map[string][]ClickUpTask{
			"list1": {
				{
					ID:     "t1",
					Name:   "", // Missing name — should be skipped
					Status: ClickUpStatus{Status: "to do"},
					List:   ClickUpListRef{ID: "list1"},
				},
				{
					ID:     "t2",
					Name:   "Valid task",
					Status: ClickUpStatus{Status: "to do"},
					List:   ClickUpListRef{ID: "list1"},
				},
			},
		},
	}

	p := NewClickUpProvider(mock, &ClickUpConfig{ListIDs: []string{"list1"}})
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	// AC11: malformed task skipped, valid task included
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task (skipped malformed), got %d", len(tasks))
	}
	if tasks[0].ID != "t2" {
		t.Errorf("task.ID = %q, want %q", tasks[0].ID, "t2")
	}
}

func TestClickUpProviderMultipleLists(t *testing.T) {
	t.Parallel()

	mock := &mockTaskFetcher{
		tasksByList: map[string][]ClickUpTask{
			"list1": {
				{ID: "t1", Name: "Task from list 1", Status: ClickUpStatus{Status: "to do"}, List: ClickUpListRef{ID: "list1"}},
			},
			"list2": {
				{ID: "t2", Name: "Task from list 2", Status: ClickUpStatus{Status: "open"}, List: ClickUpListRef{ID: "list2"}},
			},
		},
	}

	p := NewClickUpProvider(mock, &ClickUpConfig{ListIDs: []string{"list1", "list2"}})
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks from 2 lists, got %d", len(tasks))
	}
}

func TestClickUpProviderReadOnlyMethods(t *testing.T) {
	t.Parallel()

	p := NewClickUpProvider(&mockTaskFetcher{}, &ClickUpConfig{})

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

func TestClickUpProviderWatch(t *testing.T) {
	t.Parallel()

	p := NewClickUpProvider(&mockTaskFetcher{}, &ClickUpConfig{})
	if ch := p.Watch(); ch != nil {
		t.Errorf("Watch() = %v, want nil", ch)
	}
}

func TestClickUpProviderHealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("healthy", func(t *testing.T) {
		t.Parallel()
		mock := &mockTaskFetcher{
			user: &ClickUpUser{ID: 1, Username: "test"},
		}
		p := NewClickUpProvider(mock, &ClickUpConfig{})

		result := p.HealthCheck()
		if result.Overall != core.HealthOK {
			t.Errorf("Overall = %q, want %q", result.Overall, core.HealthOK)
		}
		if len(result.Items) != 1 {
			t.Fatalf("Items length = %d, want 1", len(result.Items))
		}
		if result.Items[0].Name != "clickup_connectivity" {
			t.Errorf("Item name = %q, want %q", result.Items[0].Name, "clickup_connectivity")
		}
		if mock.getUserCalls != 1 {
			t.Errorf("GetAuthorizedUser called %d times, want 1", mock.getUserCalls)
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		t.Parallel()
		mock := &mockTaskFetcher{
			userErr: errors.New("auth failed"),
		}
		p := NewClickUpProvider(mock, &ClickUpConfig{})

		result := p.HealthCheck()
		if result.Overall != core.HealthFail {
			t.Errorf("Overall = %q, want %q", result.Overall, core.HealthFail)
		}
		if result.Items[0].Suggestion == "" {
			t.Error("expected non-empty suggestion on failure")
		}
	})
}

func TestClickUpProviderPollInterval(t *testing.T) {
	t.Parallel()

	p := NewClickUpProvider(&mockTaskFetcher{}, &ClickUpConfig{
		PollInterval: 5 * time.Minute,
	})
	if got := p.PollInterval(); got != 5*time.Minute {
		t.Errorf("PollInterval() = %v, want %v", got, 5*time.Minute)
	}
}

func TestParseUnixMillis(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"valid timestamp", "1735689600000", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), false},
		{"zero", "0", time.Unix(0, 0).UTC(), false},
		{"with milliseconds", "1735689600500", time.Date(2025, 1, 1, 0, 0, 0, 500000000, time.UTC), false},
		{"invalid string", "not-a-number", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseUnixMillis(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseUnixMillis(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("parseUnixMillis(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- Integration tests with httptest ---

// clickUpMockHandler serves canned ClickUp API responses.
type clickUpMockHandler struct {
	tasksResponse map[string][]byte // listID -> JSON response
	userResponse  []byte
	rateLimited   bool
	retryAfter    string
}

func (h *clickUpMockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.rateLimited {
		w.Header().Set("Retry-After", h.retryAfter)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	path := r.URL.Path

	// Match /list/{id}/task
	if r.Method == http.MethodGet && len(path) > 6 && path[:6] == "/list/" {
		parts := splitPath(path)
		if len(parts) >= 3 && parts[2] == "task" {
			listID := parts[1]
			w.Header().Set("Content-Type", "application/json")
			if data, ok := h.tasksResponse[listID]; ok {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(data)
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"tasks":[]}`))
			}
			return
		}
	}

	// Match /user
	if r.Method == http.MethodGet && path == "/user" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if h.userResponse != nil {
			_, _ = w.Write(h.userResponse)
		} else {
			_, _ = w.Write([]byte(`{"user":{"id":1,"username":"test","email":"test@example.com"}}`))
		}
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func splitPath(path string) []string {
	var parts []string
	for _, p := range splitOnSlash(path) {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitOnSlash(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func newTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func newClientForServer(server *httptest.Server) *Client {
	return NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
}

// AC12: Integration test with httptest serving canned ClickUp responses through full provider flow.
func TestClickUpProviderIntegration(t *testing.T) {
	t.Parallel()

	tasks := tasksResponse{
		Tasks: []ClickUpTask{
			{
				ID:          "cu-task-1",
				Name:        "Build API endpoint",
				Description: "Create REST endpoint for user profiles",
				Status:      ClickUpStatus{Status: "in progress", Type: "custom"},
				Priority:    &ClickUpPriority{ID: "2"},
				DueDate:     "1735689600000",
				Tags:        []ClickUpTag{{Name: "backend"}, {Name: "api"}},
				URL:         "https://app.clickup.com/t/cu-task-1",
				List:        ClickUpListRef{ID: "list-100", Name: "Sprint 3"},
			},
			{
				ID:     "cu-task-2",
				Name:   "Write tests",
				Status: ClickUpStatus{Status: "to do", Type: "custom"},
				Tags:   []ClickUpTag{{Name: "testing"}},
				URL:    "https://app.clickup.com/t/cu-task-2",
				List:   ClickUpListRef{ID: "list-100", Name: "Sprint 3"},
			},
		},
	}

	tasksJSON, err := json.Marshal(tasks)
	if err != nil {
		t.Fatalf("marshal tasks: %v", err)
	}

	handler := &clickUpMockHandler{
		tasksResponse: map[string][]byte{
			"list-100": tasksJSON,
		},
	}
	server := newTestServer(t, handler)
	client := newClientForServer(server)

	p := NewClickUpProvider(client, &ClickUpConfig{
		ListIDs:      []string{"list-100"},
		PollInterval: 5 * time.Minute,
	})

	loadedTasks, loadErr := p.LoadTasks()
	if loadErr != nil {
		t.Fatalf("LoadTasks() error: %v", loadErr)
	}

	if len(loadedTasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(loadedTasks))
	}

	// Verify first task full mapping
	task1 := loadedTasks[0]
	if task1.Text != "Build API endpoint" {
		t.Errorf("task1.Text = %q, want %q", task1.Text, "Build API endpoint")
	}
	if task1.Context != "Create REST endpoint for user profiles | backend, api" {
		t.Errorf("task1.Context = %q, want %q", task1.Context, "Create REST endpoint for user profiles | backend, api")
	}
	if task1.Status != core.StatusInProgress {
		t.Errorf("task1.Status = %q, want %q", task1.Status, core.StatusInProgress)
	}
	if task1.Effort != core.EffortDeepWork {
		t.Errorf("task1.Effort = %q, want %q", task1.Effort, core.EffortDeepWork)
	}
	if task1.DeferUntil == nil {
		t.Error("task1.DeferUntil is nil, want non-nil")
	}

	// Verify second task
	task2 := loadedTasks[1]
	if task2.Text != "Write tests" {
		t.Errorf("task2.Text = %q, want %q", task2.Text, "Write tests")
	}
	if task2.Context != "testing" {
		t.Errorf("task2.Context = %q, want %q", task2.Context, "testing")
	}
	if task2.Status != core.StatusTodo {
		t.Errorf("task2.Status = %q, want %q", task2.Status, core.StatusTodo)
	}
}

func TestClickUpProviderHealthCheckViaHTTP(t *testing.T) {
	t.Parallel()

	t.Run("healthy API returns HealthOK", func(t *testing.T) {
		t.Parallel()
		handler := &clickUpMockHandler{}
		server := newTestServer(t, handler)
		client := newClientForServer(server)
		p := NewClickUpProvider(client, &ClickUpConfig{})

		result := p.HealthCheck()
		if result.Overall != core.HealthOK {
			t.Errorf("Overall = %q, want %q", result.Overall, core.HealthOK)
		}
		if result.Duration <= 0 {
			t.Errorf("Duration = %v, want > 0", result.Duration)
		}
	})

	t.Run("unreachable API returns HealthFail with suggestion", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.NotFoundHandler())
		server.Close()

		client := newClientForServer(server)
		p := NewClickUpProvider(client, &ClickUpConfig{})

		result := p.HealthCheck()
		if result.Overall != core.HealthFail {
			t.Errorf("Overall = %q, want %q", result.Overall, core.HealthFail)
		}
		if result.Items[0].Suggestion == "" {
			t.Error("expected non-empty Suggestion on failure")
		}
	})
}

// AC12: Contract test with httptest server.
func TestClickUpProviderContract(t *testing.T) {
	t.Parallel()

	tasks := tasksResponse{
		Tasks: []ClickUpTask{
			{
				ID:     "contract-task-1",
				Name:   "Contract test task",
				Status: ClickUpStatus{Status: "to do"},
				List:   ClickUpListRef{ID: "contract-list"},
			},
		},
	}
	tasksJSON, err := json.Marshal(tasks)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		handler := &clickUpMockHandler{
			tasksResponse: map[string][]byte{
				"contract-list": tasksJSON,
			},
		}
		server := newTestServer(t, handler)
		client := newClientForServer(server)
		return NewClickUpProvider(client, &ClickUpConfig{
			ListIDs: []string{"contract-list"},
		})
	}

	adapters.RunContractTests(t, factory)
}

func TestClickUpProviderReadOnlyMethodsViaHTTP(t *testing.T) {
	t.Parallel()

	handler := &clickUpMockHandler{}
	server := newTestServer(t, handler)
	client := newClientForServer(server)
	p := NewClickUpProvider(client, &ClickUpConfig{})

	t.Run("SaveTask returns ErrReadOnly", func(t *testing.T) {
		t.Parallel()
		if err := p.SaveTask(&core.Task{ID: "test"}); !errors.Is(err, core.ErrReadOnly) {
			t.Errorf("SaveTask() = %v, want ErrReadOnly", err)
		}
	})

	t.Run("SaveTasks returns ErrReadOnly", func(t *testing.T) {
		t.Parallel()
		if err := p.SaveTasks([]*core.Task{{ID: "test"}}); !errors.Is(err, core.ErrReadOnly) {
			t.Errorf("SaveTasks() = %v, want ErrReadOnly", err)
		}
	})

	t.Run("DeleteTask returns ErrReadOnly", func(t *testing.T) {
		t.Parallel()
		if err := p.DeleteTask("test"); !errors.Is(err, core.ErrReadOnly) {
			t.Errorf("DeleteTask() = %v, want ErrReadOnly", err)
		}
	})

	t.Run("MarkComplete returns ErrReadOnly", func(t *testing.T) {
		t.Parallel()
		if err := p.MarkComplete("test"); !errors.Is(err, core.ErrReadOnly) {
			t.Errorf("MarkComplete() = %v, want ErrReadOnly", err)
		}
	})
}
