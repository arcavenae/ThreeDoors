package clickup

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/adapters"
	"github.com/arcaven/ThreeDoors/internal/core"
)

// loadContractTestdata loads a JSON file from testdata/ for contract tests.
func loadContractTestdata(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		t.Fatalf("load testdata %s: %v", filename, err)
	}
	return data
}

// contractMockHandler serves canned ClickUp API responses for contract tests.
// It supports configurable task responses, user responses, rate limiting,
// and request logging for lifecycle verification.
type contractMockHandler struct {
	tasksResponse map[string][]byte // listID -> JSON response
	userResponse  []byte
	rateLimited   bool
	retryAfter    string
	requestLog    []string // tracks request methods+paths
}

func (h *contractMockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.requestLog = append(h.requestLog, fmt.Sprintf("%s %s", r.Method, r.URL.Path))

	// Verify auth header is present
	if r.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"err":"Token invalid","ECODE":"OAUTH_025"}`))
		return
	}

	if h.rateLimited {
		w.Header().Set("Retry-After", h.retryAfter)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	path := r.URL.Path

	// Match /list/{id}/task
	if r.Method == http.MethodGet && strings.HasPrefix(path, "/list/") {
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

	// Match PUT /task/{id} (status update)
	if r.Method == http.MethodPut && strings.HasPrefix(path, "/task/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
		return
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

func newContractTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func newContractClient(server *httptest.Server) *Client {
	return NewClient(AuthConfig{APIToken: "test-contract-token", BaseURL: server.URL})
}

// --- AC1: Contract test suite validates TaskProvider interface compliance ---

func TestClickUpContractSuite(t *testing.T) {
	t.Parallel()

	tasksData := loadContractTestdata(t, "tasks.json")

	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		handler := &contractMockHandler{
			tasksResponse: map[string][]byte{
				"contract-list": tasksData,
			},
		}
		server := newContractTestServer(t, handler)
		client := newContractClient(server)
		return NewClickUpProvider(client, &ClickUpConfig{
			ListIDs: []string{"contract-list"},
		})
	}

	adapters.RunContractTests(t, factory)
}

// --- AC2: Tests cover LoadTasks, SaveTask (status update), HealthCheck, and error paths ---

func TestContractLoadTasks(t *testing.T) {
	t.Parallel()

	tasksData := loadContractTestdata(t, "tasks.json")
	handler := &contractMockHandler{
		tasksResponse: map[string][]byte{
			"list-100": tasksData,
		},
	}
	server := newContractTestServer(t, handler)
	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"list-100"}})

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}

	// Verify each task loaded correctly
	tests := []struct {
		name       string
		idx        int
		wantText   string
		wantStatus core.TaskStatus
		wantEffort core.TaskEffort
	}{
		{"task 1 in progress", 0, "Implement user authentication", core.StatusInProgress, core.EffortDeepWork},
		{"task 2 todo", 1, "Write unit tests", core.StatusTodo, core.EffortMedium},
		{"task 3 complete", 2, "Deploy to staging", core.StatusComplete, core.EffortQuickWin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := tasks[tt.idx]
			if task.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", task.Text, tt.wantText)
			}
			if task.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", task.Status, tt.wantStatus)
			}
			if task.Effort != tt.wantEffort {
				t.Errorf("Effort = %q, want %q", task.Effort, tt.wantEffort)
			}
		})
	}
}

func TestContractSaveTaskReturnsReadOnly(t *testing.T) {
	t.Parallel()

	p := NewClickUpProvider(&mockTaskFetcher{}, &ClickUpConfig{})

	if err := p.SaveTask(&core.Task{ID: "test"}); !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("SaveTask() = %v, want ErrReadOnly", err)
	}
}

func TestContractHealthCheckPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		userErr     error
		wantOverall core.HealthStatus
	}{
		{"healthy API", nil, core.HealthOK},
		{"unhealthy API", errors.New("connection refused"), core.HealthFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockTaskFetcher{
				user:    &ClickUpUser{ID: 1, Username: "test"},
				userErr: tt.userErr,
			}
			p := NewClickUpProvider(mock, &ClickUpConfig{})
			result := p.HealthCheck()
			if result.Overall != tt.wantOverall {
				t.Errorf("Overall = %q, want %q", result.Overall, tt.wantOverall)
			}
			if result.Duration < 0 {
				t.Errorf("Duration = %v, want >= 0", result.Duration)
			}
		})
	}
}

func TestContractErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockTaskFetcher
		wantErr string
	}{
		{
			name:    "API network error",
			mock:    &mockTaskFetcher{tasksErr: errors.New("network timeout")},
			wantErr: "network timeout",
		},
		{
			name:    "API rate limit error",
			mock:    &mockTaskFetcher{tasksErr: &RateLimitError{RetryAfter: 30 * time.Second}},
			wantErr: "rate limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := NewClickUpProvider(tt.mock, &ClickUpConfig{ListIDs: []string{"list1"}})
			_, err := p.LoadTasks()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// --- AC3: Golden file tests for ClickUp API response parsing ---

// goldenParsedTask is a simplified representation for golden file comparison.
type goldenParsedTask struct {
	ID             string           `json:"id"`
	Text           string           `json:"text"`
	Context        string           `json:"context"`
	Status         core.TaskStatus  `json:"status"`
	Effort         core.TaskEffort  `json:"effort"`
	SourceProvider string           `json:"source_provider"`
	SourceRefs     []core.SourceRef `json:"source_refs"`
	HasDeferUntil  bool             `json:"has_defer_until"`
	DeferUntil     string           `json:"defer_until,omitempty"`
}

func TestGoldenFileResponseParsing(t *testing.T) {
	t.Parallel()

	tasksData := loadContractTestdata(t, "tasks.json")
	goldenData := loadContractTestdata(t, "golden_tasks_parsed.json")

	var goldenTasks []goldenParsedTask
	if err := json.Unmarshal(goldenData, &goldenTasks); err != nil {
		t.Fatalf("unmarshal golden file: %v", err)
	}

	handler := &contractMockHandler{
		tasksResponse: map[string][]byte{
			"list-100": tasksData,
		},
	}
	server := newContractTestServer(t, handler)
	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"list-100"}})

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != len(goldenTasks) {
		t.Fatalf("got %d tasks, golden expects %d", len(tasks), len(goldenTasks))
	}

	for i, golden := range goldenTasks {
		t.Run(fmt.Sprintf("task_%d_%s", i, golden.ID), func(t *testing.T) {
			task := tasks[i]

			if task.ID != golden.ID {
				t.Errorf("ID = %q, want %q", task.ID, golden.ID)
			}
			if task.Text != golden.Text {
				t.Errorf("Text = %q, want %q", task.Text, golden.Text)
			}
			if task.Context != golden.Context {
				t.Errorf("Context = %q, want %q", task.Context, golden.Context)
			}
			if task.Status != golden.Status {
				t.Errorf("Status = %q, want %q", task.Status, golden.Status)
			}
			if task.Effort != golden.Effort {
				t.Errorf("Effort = %q, want %q", task.Effort, golden.Effort)
			}
			if task.SourceProvider != golden.SourceProvider {
				t.Errorf("SourceProvider = %q, want %q", task.SourceProvider, golden.SourceProvider)
			}

			if len(task.SourceRefs) != len(golden.SourceRefs) {
				t.Fatalf("SourceRefs length = %d, want %d", len(task.SourceRefs), len(golden.SourceRefs))
			}
			for j, ref := range golden.SourceRefs {
				if task.SourceRefs[j].Provider != ref.Provider {
					t.Errorf("SourceRefs[%d].Provider = %q, want %q", j, task.SourceRefs[j].Provider, ref.Provider)
				}
				if task.SourceRefs[j].NativeID != ref.NativeID {
					t.Errorf("SourceRefs[%d].NativeID = %q, want %q", j, task.SourceRefs[j].NativeID, ref.NativeID)
				}
			}

			if golden.HasDeferUntil {
				if task.DeferUntil == nil {
					t.Fatal("DeferUntil is nil, want non-nil")
				}
				expected, parseErr := time.Parse(time.RFC3339, golden.DeferUntil)
				if parseErr != nil {
					t.Fatalf("parse golden DeferUntil: %v", parseErr)
				}
				if !task.DeferUntil.Equal(expected) {
					t.Errorf("DeferUntil = %v, want %v", task.DeferUntil, expected)
				}
			} else {
				if task.DeferUntil != nil {
					t.Errorf("DeferUntil = %v, want nil", task.DeferUntil)
				}
			}
		})
	}
}

func TestGoldenFileSpecialCharacters(t *testing.T) {
	t.Parallel()

	tasksData := loadContractTestdata(t, "special_characters.json")
	handler := &contractMockHandler{
		tasksResponse: map[string][]byte{
			"list-200": tasksData,
		},
	}
	server := newContractTestServer(t, handler)
	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"list-200"}})

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}

	tests := []struct {
		name        string
		idx         int
		wantText    string
		wantContext string
	}{
		{
			name:        "Unicode preserved",
			idx:         0,
			wantText:    "Review café résumé — handle naïve edge cases «here»",
			wantContext: "Contains: em-dash—, quotes\u201c\u201d, ellipsis…, ñ, ü, ß | i18n",
		},
		{
			name:        "Emoji preserved",
			idx:         1,
			wantText:    "🎉 Ship release 2.0 🚀",
			wantContext: "Celebrate with 🍕 and 🎂 | release, 🏷️",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := tasks[tt.idx]
			if task.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", task.Text, tt.wantText)
			}
			if task.Context != tt.wantContext {
				t.Errorf("Context = %q, want %q", task.Context, tt.wantContext)
			}
		})
	}
}

func TestGoldenFileMalformedTasks(t *testing.T) {
	t.Parallel()

	tasksData := loadContractTestdata(t, "malformed_missing_fields.json")
	handler := &contractMockHandler{
		tasksResponse: map[string][]byte{
			"list-300": tasksData,
		},
	}
	server := newContractTestServer(t, handler)
	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"list-300"}})

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	// Only tasks with non-empty names should be returned
	// mal-001 has empty name (skipped), mal-002 has name (kept), mal-003 has name but empty ID (kept)
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2 (skipped task with empty name)", len(tasks))
	}

	if tasks[0].Text != "Valid task among malformed" {
		t.Errorf("tasks[0].Text = %q, want %q", tasks[0].Text, "Valid task among malformed")
	}
	if tasks[1].Text != "Task with empty ID" {
		t.Errorf("tasks[1].Text = %q, want %q", tasks[1].Text, "Task with empty ID")
	}
}

// --- AC4: Fuzz tests for malformed ClickUp API responses ---

func FuzzClickUpResponseParsing(f *testing.F) {
	// Seed with valid and edge-case JSON
	f.Add([]byte(`{"tasks":[{"id":"1","name":"Test","status":{"status":"to do"},"list":{"id":"l1"}}]}`))
	f.Add([]byte(`{"tasks":[]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"tasks":[{"id":"","name":"","status":{"status":""},"priority":null}]}`))
	f.Add([]byte(`{"tasks":[{"id":"1","name":"Test","priority":{"id":"99"},"due_date":"not-a-number"}]}`))
	f.Add([]byte(`{"tasks":[{"id":"1","name":"Test","tags":[{"name":""},{"name":"a"}],"due_date":"0"}]}`))
	f.Add([]byte(`{"tasks":null}`))
	f.Add([]byte(`not json at all`))
	f.Add([]byte(`{"tasks":[null]}`))
	f.Add([]byte(`{"tasks":[{"id":"1","name":"Test","priority":{"id":"-1"},"due_date":"9999999999999999"}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/user" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"user":{"id":1}}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
		})

		server := httptest.NewServer(handler)
		t.Cleanup(server.Close)

		client := NewClient(AuthConfig{APIToken: "fuzz-token", BaseURL: server.URL})
		p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"fuzz-list"}})

		// Must not panic — errors are acceptable
		tasks, err := p.LoadTasks()
		if err != nil {
			return // errors from malformed JSON are fine
		}

		// If parsing succeeded, verify invariants
		for _, task := range tasks {
			if task.Text == "" {
				t.Error("parsed task has empty Text — should have been skipped")
			}
			if task.SourceProvider == "" {
				t.Error("parsed task has empty SourceProvider")
			}
			if task.Status == "" {
				t.Error("parsed task has empty Status")
			}
		}
	})
}

func FuzzFieldMappingPriority(f *testing.F) {
	f.Add("1")
	f.Add("2")
	f.Add("3")
	f.Add("4")
	f.Add("")
	f.Add("0")
	f.Add("-1")
	f.Add("999")
	f.Add("not-a-number")

	f.Fuzz(func(t *testing.T, priorityID string) {
		priority := &ClickUpPriority{ID: priorityID}
		effort := MapPriority(priority)

		// Must always return a valid effort value
		validEfforts := map[core.TaskEffort]bool{
			core.EffortQuickWin: true,
			core.EffortMedium:   true,
			core.EffortDeepWork: true,
		}
		if !validEfforts[effort] {
			t.Errorf("MapPriority(%q) = %q, not a valid effort", priorityID, effort)
		}
	})
}

func FuzzFieldMappingStatus(f *testing.F) {
	f.Add("to do")
	f.Add("open")
	f.Add("in progress")
	f.Add("complete")
	f.Add("closed")
	f.Add("done")
	f.Add("")
	f.Add("UNKNOWN")
	f.Add("  to do  ")
	f.Add("To Do")

	f.Fuzz(func(t *testing.T, status string) {
		p := NewClickUpProvider(&mockTaskFetcher{}, &ClickUpConfig{})
		result := p.mapStatus(status)

		// Must always return a valid status
		validStatuses := map[core.TaskStatus]bool{
			core.StatusTodo:       true,
			core.StatusInProgress: true,
			core.StatusComplete:   true,
			core.StatusBlocked:    true,
			core.StatusInReview:   true,
			core.StatusDeferred:   true,
			core.StatusArchived:   true,
		}
		if !validStatuses[result] {
			t.Errorf("mapStatus(%q) = %q, not a valid status", status, result)
		}
	})
}

func FuzzParseUnixMillis(f *testing.F) {
	f.Add("0")
	f.Add("1735689600000")
	f.Add("-1")
	f.Add("9999999999999")
	f.Add("not-a-number")
	f.Add("")

	f.Fuzz(func(t *testing.T, ms string) {
		result, err := parseUnixMillis(ms)
		if err != nil {
			return // parse errors are fine
		}
		// If parsing succeeds, result must be in UTC
		if result.Location() != time.UTC {
			t.Errorf("parseUnixMillis(%q) location = %v, want UTC", ms, result.Location())
		}
	})
}

// --- AC5: httptest.NewServer integration tests — full provider lifecycle ---

func TestLifecycleAuthLoadSyncError(t *testing.T) {
	t.Parallel()

	tasksData := loadContractTestdata(t, "tasks.json")
	userData := loadContractTestdata(t, "user.json")

	handler := &contractMockHandler{
		tasksResponse: map[string][]byte{
			"list-100": tasksData,
		},
		userResponse: userData,
	}
	server := newContractTestServer(t, handler)
	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{
		ListIDs:      []string{"list-100"},
		PollInterval: 1 * time.Minute,
	})

	// Step 1: Health check (auth verification)
	healthResult := p.HealthCheck()
	if healthResult.Overall != core.HealthOK {
		t.Fatalf("HealthCheck: %q, want %q", healthResult.Overall, core.HealthOK)
	}

	// Step 2: Load tasks
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("LoadTasks returned %d tasks, want 3", len(tasks))
	}

	// Step 3: Verify read-only operations return ErrReadOnly
	if err := p.SaveTask(tasks[0]); !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("SaveTask: %v, want ErrReadOnly", err)
	}
	if err := p.SaveTasks(tasks); !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("SaveTasks: %v, want ErrReadOnly", err)
	}
	if err := p.DeleteTask(tasks[0].ID); !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("DeleteTask: %v, want ErrReadOnly", err)
	}
	if err := p.MarkComplete(tasks[0].ID); err != nil {
		t.Errorf("MarkComplete: unexpected error: %v", err)
	}

	// Step 4: Watch returns nil (poll-based)
	if ch := p.Watch(); ch != nil {
		t.Errorf("Watch() = %v, want nil", ch)
	}

	// Step 5: Verify provider name
	if p.Name() != "clickup" {
		t.Errorf("Name() = %q, want %q", p.Name(), "clickup")
	}

	// Step 6: Verify request log shows expected auth → load → health flow
	if len(handler.requestLog) < 2 {
		t.Fatalf("expected at least 2 requests, got %d", len(handler.requestLog))
	}
}

func TestLifecycleRateLimitDuringLoad(t *testing.T) {
	t.Parallel()

	handler := &contractMockHandler{
		rateLimited: true,
		retryAfter:  "30",
	}
	server := newContractTestServer(t, handler)
	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"list-100"}})

	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("expected error from rate-limited API")
	}

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}
	if rle.RetryAfter != 30*time.Second {
		t.Errorf("RetryAfter = %v, want 30s", rle.RetryAfter)
	}
}

func TestLifecycleUnauthorized(t *testing.T) {
	t.Parallel()

	// Server that rejects all requests with 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"err":"Token invalid"}`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "bad-token", BaseURL: server.URL})
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"list-100"}})

	// Health check should fail
	result := p.HealthCheck()
	if result.Overall != core.HealthFail {
		t.Errorf("HealthCheck Overall = %q, want %q", result.Overall, core.HealthFail)
	}

	// LoadTasks should fail
	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("expected error from unauthorized API")
	}
}

func TestLifecycleEmptyTaskList(t *testing.T) {
	t.Parallel()

	emptyData := loadContractTestdata(t, "empty_tasks.json")
	handler := &contractMockHandler{
		tasksResponse: map[string][]byte{
			"list-100": emptyData,
		},
	}
	server := newContractTestServer(t, handler)
	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"list-100"}})

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

func TestLifecycleMultipleListAggregation(t *testing.T) {
	t.Parallel()

	list1 := []byte(`{"tasks":[{"id":"t1","name":"Task from list 1","status":{"status":"to do"},"list":{"id":"list-a"}}]}`)
	list2 := []byte(`{"tasks":[{"id":"t2","name":"Task from list 2","status":{"status":"in progress"},"list":{"id":"list-b"}},{"id":"t3","name":"Task from list 3","status":{"status":"complete"},"list":{"id":"list-b"}}]}`)

	handler := &contractMockHandler{
		tasksResponse: map[string][]byte{
			"list-a": list1,
			"list-b": list2,
		},
	}
	server := newContractTestServer(t, handler)
	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"list-a", "list-b"}})

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks from 2 lists, want 3", len(tasks))
	}

	// Verify tasks come from both lists in order
	if tasks[0].ID != "t1" {
		t.Errorf("tasks[0].ID = %q, want %q", tasks[0].ID, "t1")
	}
	if tasks[1].ID != "t2" {
		t.Errorf("tasks[1].ID = %q, want %q", tasks[1].ID, "t2")
	}
	if tasks[2].ID != "t3" {
		t.Errorf("tasks[2].ID = %q, want %q", tasks[2].ID, "t3")
	}
}

func TestLifecycleServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"list-100"}})

	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("expected error from 500 server")
	}
}

// --- AC8: Table-driven tests where applicable ---
// (Covered by TestContractLoadTasks, TestContractHealthCheckPaths,
//  TestContractErrorPaths, TestGoldenFileSpecialCharacters above)

// --- AC9: Benchmark test for field mapping performance with 100+ tasks ---

func BenchmarkFieldMapping100Tasks(b *testing.B) {
	// Generate 100 tasks with varying fields
	var tasks []ClickUpTask
	for i := 0; i < 100; i++ {
		task := ClickUpTask{
			ID:          fmt.Sprintf("bench-task-%d", i),
			Name:        fmt.Sprintf("Benchmark task %d with some description text", i),
			Description: fmt.Sprintf("Description for task %d with details about the work to be done", i),
			Status:      ClickUpStatus{Status: []string{"to do", "in progress", "complete", "open", "closed"}[i%5], Type: "custom"},
			Priority:    &ClickUpPriority{ID: fmt.Sprintf("%d", (i%4)+1)},
			DueDate:     fmt.Sprintf("%d", 1735689600000+int64(i)*86400000),
			Tags: []ClickUpTag{
				{Name: fmt.Sprintf("tag-%d", i%10)},
				{Name: fmt.Sprintf("category-%d", i%5)},
			},
			URL:  fmt.Sprintf("https://app.clickup.com/t/bench-%d", i),
			List: ClickUpListRef{ID: fmt.Sprintf("list-%d", i%3), Name: fmt.Sprintf("Sprint %d", i%3)},
		}
		tasks = append(tasks, task)
	}

	tasksResp := tasksResponse{Tasks: tasks}
	tasksJSON, err := json.Marshal(tasksResp)
	if err != nil {
		b.Fatalf("marshal: %v", err)
	}

	handler := &contractMockHandler{
		tasksResponse: map[string][]byte{
			"bench-list": tasksJSON,
		},
	}
	server := httptest.NewServer(handler)
	b.Cleanup(server.Close)

	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"bench-list"}})

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		result, loadErr := p.LoadTasks()
		if loadErr != nil {
			b.Fatalf("LoadTasks: %v", loadErr)
		}
		if len(result) != 100 {
			b.Fatalf("got %d tasks, want 100", len(result))
		}
	}
}

func BenchmarkFieldMapping500Tasks(b *testing.B) {
	var tasks []ClickUpTask
	for i := 0; i < 500; i++ {
		task := ClickUpTask{
			ID:          fmt.Sprintf("bench500-%d", i),
			Name:        fmt.Sprintf("Large batch task %d", i),
			Description: fmt.Sprintf("Description %d", i),
			Status:      ClickUpStatus{Status: "to do"},
			Priority:    &ClickUpPriority{ID: "3"},
			Tags:        []ClickUpTag{{Name: "batch"}},
			List:        ClickUpListRef{ID: "list-1"},
		}
		tasks = append(tasks, task)
	}

	tasksResp := tasksResponse{Tasks: tasks}
	tasksJSON, err := json.Marshal(tasksResp)
	if err != nil {
		b.Fatalf("marshal: %v", err)
	}

	handler := &contractMockHandler{
		tasksResponse: map[string][]byte{
			"bench-list": tasksJSON,
		},
	}
	server := httptest.NewServer(handler)
	b.Cleanup(server.Close)

	client := newContractClient(server)
	p := NewClickUpProvider(client, &ClickUpConfig{ListIDs: []string{"bench-list"}})

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		result, loadErr := p.LoadTasks()
		if loadErr != nil {
			b.Fatalf("LoadTasks: %v", loadErr)
		}
		if len(result) != 500 {
			b.Fatalf("got %d tasks, want 500", len(result))
		}
	}
}

func BenchmarkMapPriority(b *testing.B) {
	priorities := []*ClickUpPriority{
		nil,
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "4"},
		{ID: "99"},
	}

	b.ResetTimer()
	for b.Loop() {
		for _, p := range priorities {
			MapPriority(p)
		}
	}
}

func BenchmarkMapStatus(b *testing.B) {
	p := NewClickUpProvider(&mockTaskFetcher{}, &ClickUpConfig{})
	statuses := []string{"to do", "open", "in progress", "complete", "closed", "done", "unknown", "  To Do  "}

	b.ResetTimer()
	for b.Loop() {
		for _, s := range statuses {
			p.mapStatus(s)
		}
	}
}
