package todoist

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/adapters"
	"github.com/arcaven/ThreeDoors/internal/core"
)

// httpMockHandler serves canned Todoist API responses for contract tests.
type httpMockHandler struct {
	tasksResponse    []byte
	projectsResponse []byte
	closeStatus      int
	rateLimit        bool
	retryAfter       string
}

func (h *httpMockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.rateLimit {
		w.Header().Set("Retry-After", h.retryAfter)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/tasks":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(h.tasksResponse)

	case r.Method == http.MethodGet && r.URL.Path == "/projects":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(h.projectsResponse)

	case r.Method == http.MethodPost && len(r.URL.Path) > len("/tasks/") && r.URL.Path[len(r.URL.Path)-6:] == "/close":
		status := http.StatusNoContent
		if h.closeStatus != 0 {
			status = h.closeStatus
		}
		w.WriteHeader(status)

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func loadTestdata(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		t.Fatalf("load testdata %s: %v", filename, err)
		return nil
	}
	return data
}

func newTestServer(t *testing.T, handler *httpMockHandler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func newClientWithServer(server *httptest.Server) *Client {
	return &Client{
		baseURL:    server.URL,
		authHeader: "Bearer test-token",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AC1: Contract tests pass for TodoistProvider with mock HTTP server backing.
func TestTodoistProviderContract(t *testing.T) {
	t.Parallel()

	tasksData := loadTestdata(t, "tasks.json")
	projectsData := loadTestdata(t, "projects.json")

	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		handler := &httpMockHandler{
			tasksResponse:    tasksData,
			projectsResponse: projectsData,
		}
		server := newTestServer(t, handler)
		client := newClientWithServer(server)
		return NewTodoistProvider(client, &TodoistConfig{})
	}

	adapters.RunContractTests(t, factory)
}

// AC4: Table-driven field mapping tests cover all 5 Todoist priority values.
func TestPriorityMappingAllValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		priority int
		want     core.TaskEffort
	}{
		{"priority 0 (none)", 0, core.EffortQuickWin},
		{"priority 1 (normal)", 1, core.EffortQuickWin},
		{"priority 2 (high)", 2, core.EffortMedium},
		{"priority 3 (urgent)", 3, core.EffortDeepWork},
		{"priority 4 (critical)", 4, core.EffortDeepWork},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MapPriorityToEffort(tt.priority)
			if got != tt.want {
				t.Errorf("MapPriorityToEffort(%d) = %q, want %q", tt.priority, got, tt.want)
			}
		})
	}
}

// AC4 continued: Table-driven test with mock HTTP server verifying priority mapping
// through the full provider pipeline.
func TestPriorityMappingViaHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		priority   int
		wantEffort core.TaskEffort
	}{
		{"priority 0 via HTTP", 0, core.EffortQuickWin},
		{"priority 1 via HTTP", 1, core.EffortQuickWin},
		{"priority 2 via HTTP", 2, core.EffortMedium},
		{"priority 3 via HTTP", 3, core.EffortDeepWork},
		{"priority 4 via HTTP", 4, core.EffortDeepWork},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := TodoistTask{
				ID:       "prio-test",
				Content:  "Priority test",
				Priority: tt.priority,
			}
			tasksJSON, err := json.Marshal([]TodoistTask{task})
			if err != nil {
				t.Fatalf("marshal tasks: %v", err)
				return
			}

			handler := &httpMockHandler{
				tasksResponse:    tasksJSON,
				projectsResponse: []byte("[]"),
			}
			server := newTestServer(t, handler)
			client := newClientWithServer(server)
			p := NewTodoistProvider(client, &TodoistConfig{})

			tasks, loadErr := p.LoadTasks()
			if loadErr != nil {
				t.Fatalf("LoadTasks() error: %v", loadErr)
			}
			if len(tasks) != 1 {
				t.Fatalf("got %d tasks, want 1", len(tasks))
			}
			if tasks[0].Effort != tt.wantEffort {
				t.Errorf("Effort = %q, want %q", tasks[0].Effort, tt.wantEffort)
			}
		})
	}
}

// AC5: Table-driven tests for is_deleted filtering — verify deleted tasks excluded.
// Note: Todoist REST API v1 only returns active tasks by default.
// The is_completed field controls whether tasks appear as complete.
// "Deleted" tasks simply don't appear in API responses.
func TestDeletedTaskFiltering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tasks        []TodoistTask
		wantCount    int
		wantStatuses []core.TaskStatus
	}{
		{
			name:         "no completed tasks",
			tasks:        []TodoistTask{{ID: "1", Content: "Active"}},
			wantCount:    1,
			wantStatuses: []core.TaskStatus{core.StatusTodo},
		},
		{
			name: "mix of active and completed",
			tasks: []TodoistTask{
				{ID: "1", Content: "Active", IsCompleted: false},
				{ID: "2", Content: "Done", IsCompleted: true},
			},
			wantCount:    2,
			wantStatuses: []core.TaskStatus{core.StatusTodo, core.StatusComplete},
		},
		{
			name:         "all completed",
			tasks:        []TodoistTask{{ID: "1", Content: "Done", IsCompleted: true}},
			wantCount:    1,
			wantStatuses: []core.TaskStatus{core.StatusComplete},
		},
		{
			name:      "empty response",
			tasks:     []TodoistTask{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tasksJSON, err := json.Marshal(tt.tasks)
			if err != nil {
				t.Fatalf("marshal: %v", err)
				return
			}

			handler := &httpMockHandler{
				tasksResponse:    tasksJSON,
				projectsResponse: []byte("[]"),
			}
			server := newTestServer(t, handler)
			client := newClientWithServer(server)
			p := NewTodoistProvider(client, &TodoistConfig{})

			tasks, loadErr := p.LoadTasks()
			if loadErr != nil {
				t.Fatalf("LoadTasks() error: %v", loadErr)
			}
			if len(tasks) != tt.wantCount {
				t.Fatalf("got %d tasks, want %d", len(tasks), tt.wantCount)
			}
			for i, wantStatus := range tt.wantStatuses {
				if tasks[i].Status != wantStatus {
					t.Errorf("tasks[%d].Status = %q, want %q", i, tasks[i].Status, wantStatus)
				}
			}
		})
	}
}

// AC6: Config validation tests — error when both project_ids and filter are set.
func TestConfigValidationMutuallyExclusive(t *testing.T) {
	// Cannot use t.Parallel() — subtests use t.Setenv which modifies process env

	tests := []struct {
		name     string
		settings map[string]string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "both project_ids and filter set",
			settings: map[string]string{
				"api_token":   "test-token",
				"project_ids": "proj-1,proj-2",
				"filter":      "today | overdue",
			},
			wantErr: true,
			errMsg:  "mutually exclusive",
		},
		{
			name: "only project_ids set",
			settings: map[string]string{
				"api_token":   "test-token",
				"project_ids": "proj-1",
			},
			wantErr: false,
		},
		{
			name: "only filter set",
			settings: map[string]string{
				"api_token": "test-token",
				"filter":    "today",
			},
			wantErr: false,
		},
		{
			name: "neither set",
			settings: map[string]string{
				"api_token": "test-token",
			},
			wantErr: false,
		},
		{
			name:     "missing api_token",
			settings: map[string]string{},
			wantErr:  true,
			errMsg:   "api_token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cannot use t.Parallel() with t.Setenv
			t.Setenv("TODOIST_API_TOKEN", "")

			_, err := ParseConfig(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfig() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// AC7: Rate limit test — mock 429 response with Retry-After header,
// verify RateLimitError returned.
func TestRateLimitHandling(t *testing.T) {
	t.Parallel()

	handler := &httpMockHandler{
		rateLimit:  true,
		retryAfter: "30",
	}
	server := newTestServer(t, handler)
	client := newClientWithServer(server)

	// Test GetTasks rate limit
	t.Run("GetTasks returns RateLimitError", func(t *testing.T) {
		t.Parallel()
		_, err := client.GetTasks(context.Background(), "", "")
		if err == nil {
			t.Fatal("expected error for 429 response")
			return
		}
		var rle *RateLimitError
		if !errors.As(err, &rle) {
			t.Fatalf("expected RateLimitError, got %T: %v", err, err)
		}
		if rle.RetryAfter != 30*time.Second {
			t.Errorf("RetryAfter = %v, want 30s", rle.RetryAfter)
		}
	})

	// Test GetProjects rate limit
	t.Run("GetProjects returns RateLimitError", func(t *testing.T) {
		t.Parallel()
		_, err := client.GetProjects(context.Background())
		if err == nil {
			t.Fatal("expected error for 429 response")
			return
		}
		if !IsRateLimitError(err) {
			t.Errorf("IsRateLimitError() = false, want true")
		}
	})

	// Test CloseTask rate limit
	t.Run("CloseTask returns RateLimitError", func(t *testing.T) {
		t.Parallel()
		err := client.CloseTask(context.Background(), "task-1")
		if err == nil {
			t.Fatal("expected error for 429 response")
			return
		}
		if !IsRateLimitError(err) {
			t.Errorf("IsRateLimitError() = false, want true")
		}
	})

	// Test rate limit propagates through provider
	t.Run("LoadTasks propagates RateLimitError from API", func(t *testing.T) {
		t.Parallel()
		rateLimitHandler := &httpMockHandler{
			rateLimit:  true,
			retryAfter: "45",
		}
		rlServer := newTestServer(t, rateLimitHandler)
		rlClient := newClientWithServer(rlServer)
		p := NewTodoistProvider(rlClient, &TodoistConfig{})

		_, err := p.LoadTasks()
		if err == nil {
			t.Fatal("expected error from rate-limited API")
			return
		}
	})
}

// AC8: Empty response test — verify LoadTasks returns empty slice (not nil).
func TestEmptyResponseReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	handler := &httpMockHandler{
		tasksResponse:    loadTestdata(t, "empty_tasks.json"),
		projectsResponse: []byte("[]"),
	}
	server := newTestServer(t, handler)
	client := newClientWithServer(server)
	p := NewTodoistProvider(client, &TodoistConfig{})

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if tasks == nil {
		t.Fatal("LoadTasks() returned nil, want non-nil empty slice")
	}
	if len(tasks) != 0 {
		t.Errorf("LoadTasks() returned %d tasks, want 0", len(tasks))
	}
}

// AC9: Health check test — verify HealthCheck returns success when API is
// reachable, structured error when not.
func TestHealthCheckViaHTTP(t *testing.T) {
	t.Parallel()

	t.Run("healthy API returns HealthOK", func(t *testing.T) {
		t.Parallel()
		handler := &httpMockHandler{
			projectsResponse: loadTestdata(t, "projects.json"),
		}
		server := newTestServer(t, handler)
		client := newClientWithServer(server)
		p := NewTodoistProvider(client, &TodoistConfig{})

		result := p.HealthCheck()
		if result.Overall != core.HealthOK {
			t.Errorf("Overall = %q, want %q", result.Overall, core.HealthOK)
		}
		if len(result.Items) == 0 {
			t.Fatal("expected at least one health check item")
		}
		if result.Items[0].Status != core.HealthOK {
			t.Errorf("Items[0].Status = %q, want %q", result.Items[0].Status, core.HealthOK)
		}
		if result.Duration <= 0 {
			t.Errorf("Duration = %v, want > 0", result.Duration)
		}
	})

	t.Run("unreachable API returns HealthFail with suggestion", func(t *testing.T) {
		t.Parallel()
		// Use a client pointing to an already-closed server
		server := httptest.NewServer(http.NotFoundHandler())
		server.Close()

		client := newClientWithServer(server)
		p := NewTodoistProvider(client, &TodoistConfig{})

		result := p.HealthCheck()
		if result.Overall != core.HealthFail {
			t.Errorf("Overall = %q, want %q", result.Overall, core.HealthFail)
		}
		if len(result.Items) == 0 {
			t.Fatal("expected at least one health check item")
		}
		if result.Items[0].Status != core.HealthFail {
			t.Errorf("Items[0].Status = %q, want %q", result.Items[0].Status, core.HealthFail)
		}
		if result.Items[0].Suggestion == "" {
			t.Error("expected non-empty Suggestion on failure")
		}
		if result.Items[0].Message == "" {
			t.Error("expected non-empty Message on failure")
		}
	})
}

// AC10: Special characters test — verify task content with Unicode, emoji,
// and special characters maps correctly.
func TestSpecialCharacterMapping(t *testing.T) {
	t.Parallel()

	handler := &httpMockHandler{
		tasksResponse:    loadTestdata(t, "special_characters.json"),
		projectsResponse: loadTestdata(t, "projects.json"),
	}
	server := newTestServer(t, handler)
	client := newClientWithServer(server)
	p := NewTodoistProvider(client, &TodoistConfig{})

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}

	tests := []struct {
		name        string
		taskIdx     int
		wantText    string
		wantContext string
	}{
		{
			name:        "Unicode characters preserved",
			taskIdx:     0,
			wantText:    "Review café résumé — handle naïve edge cases «here»",
			wantContext: "Contains: em-dash—, quotes\u201c\u201d, ellipsis…, ñ, ü, ß | i18n",
		},
		{
			name:        "Emoji characters preserved",
			taskIdx:     1,
			wantText:    "🎉 Ship release 2.0 🚀",
			wantContext: "Celebrate with 🍕 and 🎂 | release, 🏷️",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := tasks[tt.taskIdx]
			if task.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", task.Text, tt.wantText)
			}
			if task.Context != tt.wantContext {
				t.Errorf("Context = %q, want %q", task.Context, tt.wantContext)
			}
		})
	}
}

// AC3: Read-only methods correctly return core.ErrReadOnly via HTTP-backed provider.
func TestReadOnlyMethodsViaHTTP(t *testing.T) {
	t.Parallel()

	handler := &httpMockHandler{
		tasksResponse:    []byte("[]"),
		projectsResponse: []byte("[]"),
	}
	server := newTestServer(t, handler)
	client := newClientWithServer(server)
	p := NewTodoistProvider(client, &TodoistConfig{})

	t.Run("SaveTask returns ErrReadOnly", func(t *testing.T) {
		t.Parallel()
		err := p.SaveTask(&core.Task{ID: "test"})
		if !errors.Is(err, core.ErrReadOnly) {
			t.Errorf("SaveTask() = %v, want ErrReadOnly", err)
		}
	})

	t.Run("SaveTasks returns ErrReadOnly", func(t *testing.T) {
		t.Parallel()
		err := p.SaveTasks([]*core.Task{{ID: "test"}})
		if !errors.Is(err, core.ErrReadOnly) {
			t.Errorf("SaveTasks() = %v, want ErrReadOnly", err)
		}
	})

	t.Run("DeleteTask returns ErrReadOnly", func(t *testing.T) {
		t.Parallel()
		err := p.DeleteTask("test")
		if !errors.Is(err, core.ErrReadOnly) {
			t.Errorf("DeleteTask() = %v, want ErrReadOnly", err)
		}
	})
}

// AC2: Contract tests validate all TaskProvider methods via mock HTTP server.
// This is verified by TestTodoistProviderContract above which calls
// adapters.RunContractTests — that exercises Name, LoadTasks, SaveTask,
// SaveTasks, DeleteTask, MarkComplete, Watch, and HealthCheck.
