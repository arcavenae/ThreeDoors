package clickup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("uses default base URL", func(t *testing.T) {
		t.Parallel()
		c := NewClient(AuthConfig{APIToken: "tok"})
		if c.baseURL != defaultBaseURL {
			t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
		}
	})

	t.Run("uses custom base URL", func(t *testing.T) {
		t.Parallel()
		c := NewClient(AuthConfig{APIToken: "tok", BaseURL: "http://localhost:8080"})
		if c.baseURL != "http://localhost:8080" {
			t.Errorf("baseURL = %q, want %q", c.baseURL, "http://localhost:8080")
		}
	})
}

func TestClientAuthHeader(t *testing.T) {
	t.Parallel()

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userResponse{User: ClickUpUser{ID: 1}}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "pk_my_api_token", BaseURL: server.URL})
	_, err := client.GetAuthorizedUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ClickUp uses plain token, not Bearer
	if gotAuth != "pk_my_api_token" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "pk_my_api_token")
	}
}

func TestClientGetTasks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/list/12345/task" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "0" {
			t.Errorf("unexpected page: %s", r.URL.Query().Get("page"))
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}

		resp := tasksResponse{
			Tasks: []ClickUpTask{
				{
					ID:   "abc123",
					Name: "Fix login bug",
					Status: ClickUpStatus{
						Status: "in progress",
						Type:   "custom",
					},
					URL:  "https://app.clickup.com/t/abc123",
					List: ClickUpListRef{ID: "12345", Name: "Sprint 1"},
				},
				{
					ID:   "def456",
					Name: "Add tests",
					Status: ClickUpStatus{
						Status: "to do",
						Type:   "custom",
					},
					URL:  "https://app.clickup.com/t/def456",
					List: ClickUpListRef{ID: "12345", Name: "Sprint 1"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	tasks, err := client.GetTasks(context.Background(), "12345", 0)
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != "abc123" {
		t.Errorf("task[0].ID = %q, want %q", tasks[0].ID, "abc123")
	}
	if tasks[0].Name != "Fix login bug" {
		t.Errorf("task[0].Name = %q, want %q", tasks[0].Name, "Fix login bug")
	}
	if tasks[0].Status.Status != "in progress" {
		t.Errorf("task[0].Status.Status = %q, want %q", tasks[0].Status.Status, "in progress")
	}
}

func TestClientGetTasksPagination(t *testing.T) {
	t.Parallel()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		page := r.URL.Query().Get("page")

		var resp tasksResponse
		switch page {
		case "0":
			resp = tasksResponse{
				Tasks: []ClickUpTask{{ID: "task1", Name: "Task 1"}},
			}
		case "1":
			resp = tasksResponse{
				Tasks: []ClickUpTask{{ID: "task2", Name: "Task 2"}},
			}
		default:
			resp = tasksResponse{Tasks: []ClickUpTask{}}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "t", BaseURL: server.URL})

	tasks0, err := client.GetTasks(context.Background(), "list1", 0)
	if err != nil {
		t.Fatalf("page 0: %v", err)
	}
	if len(tasks0) != 1 || tasks0[0].ID != "task1" {
		t.Errorf("page 0: unexpected tasks: %v", tasks0)
	}

	tasks1, err := client.GetTasks(context.Background(), "list1", 1)
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if len(tasks1) != 1 || tasks1[0].ID != "task2" {
		t.Errorf("page 1: unexpected tasks: %v", tasks1)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestClientGetTasksServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "t", BaseURL: server.URL})
	_, err := client.GetTasks(context.Background(), "list1", 0)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !containsStr(err.Error(), "unexpected status 500") {
		t.Errorf("error = %q, want to contain 'unexpected status 500'", err)
	}
}

func TestClientUpdateTaskStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/task/abc123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.Status != "complete" {
			t.Errorf("status = %q, want %q", body.Status, "complete")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// ClickUp returns the updated task; we just need 200
		if _, err := w.Write([]byte(`{"id":"abc123","status":{"status":"complete"}}`)); err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	err := client.UpdateTaskStatus(context.Background(), "abc123", "complete")
	if err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}
}

func TestClientUpdateTaskStatusError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "t", BaseURL: server.URL})
	err := client.UpdateTaskStatus(context.Background(), "nonexistent", "done")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !containsStr(err.Error(), "unexpected status 404") {
		t.Errorf("error = %q, want to contain 'unexpected status 404'", err)
	}
}

func TestClientGetSpaces(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/team/team123/space" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := spacesResponse{
			Spaces: []ClickUpSpace{
				{ID: "space1", Name: "Engineering"},
				{ID: "space2", Name: "Marketing"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	spaces, err := client.GetSpaces(context.Background(), "team123")
	if err != nil {
		t.Fatalf("GetSpaces: %v", err)
	}

	if len(spaces) != 2 {
		t.Fatalf("expected 2 spaces, got %d", len(spaces))
	}
	if spaces[0].ID != "space1" {
		t.Errorf("space[0].ID = %q, want %q", spaces[0].ID, "space1")
	}
	if spaces[1].Name != "Marketing" {
		t.Errorf("space[1].Name = %q, want %q", spaces[1].Name, "Marketing")
	}
}

func TestClientGetLists(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/folder/folder456/list" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := listsResponse{
			Lists: []ClickUpList{
				{ID: "list1", Name: "Backlog"},
				{ID: "list2", Name: "Sprint 1"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	lists, err := client.GetLists(context.Background(), "folder456")
	if err != nil {
		t.Fatalf("GetLists: %v", err)
	}

	if len(lists) != 2 {
		t.Fatalf("expected 2 lists, got %d", len(lists))
	}
	if lists[0].Name != "Backlog" {
		t.Errorf("list[0].Name = %q, want %q", lists[0].Name, "Backlog")
	}
}

func TestClientGetAuthorizedUser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/user" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := userResponse{
			User: ClickUpUser{
				ID:       12345,
				Username: "testuser",
				Email:    "test@example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	user, err := client.GetAuthorizedUser(context.Background())
	if err != nil {
		t.Fatalf("GetAuthorizedUser: %v", err)
	}

	if user.ID != 12345 {
		t.Errorf("user.ID = %d, want %d", user.ID, 12345)
	}
	if user.Username != "testuser" {
		t.Errorf("user.Username = %q, want %q", user.Username, "testuser")
	}
	if user.Email != "test@example.com" {
		t.Errorf("user.Email = %q, want %q", user.Email, "test@example.com")
	}
}

func TestClientGetAuthorizedUserError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "bad-token", BaseURL: server.URL})
	_, err := client.GetAuthorizedUser(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !containsStr(err.Error(), "unexpected status 401") {
		t.Errorf("error = %q, want to contain 'unexpected status 401'", err)
	}
}

func TestClientRateLimitHandling(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "42")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	_, err := client.GetTasks(context.Background(), "list1", 0)

	if err == nil {
		t.Fatal("expected rate limit error")
	}

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}

	if rle.RetryAfter != 42*time.Second {
		t.Errorf("RetryAfter = %s, want 42s", rle.RetryAfter)
	}

	if !IsRateLimitError(err) {
		t.Error("IsRateLimitError should return true")
	}
}

func TestClientRateLimitDefaultRetryAfter(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	_, err := client.GetTasks(context.Background(), "list1", 0)

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}

	if rle.RetryAfter != 60*time.Second {
		t.Errorf("default RetryAfter = %s, want 60s", rle.RetryAfter)
	}
}

func TestClientRateLimitOnUpdate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "10")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	err := client.UpdateTaskStatus(context.Background(), "task1", "done")

	if !IsRateLimitError(err) {
		t.Errorf("expected rate limit error, got %v", err)
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"valid seconds", "30", 30 * time.Second},
		{"empty", "", 60 * time.Second},
		{"invalid", "not-a-number", 60 * time.Second},
		{"zero", "0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseRetryAfter(tt.value)
			if got != tt.want {
				t.Errorf("parseRetryAfter(%q) = %s, want %s", tt.value, got, tt.want)
			}
		})
	}
}

func TestClientContentTypeHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want %q", r.Header.Get("Content-Type"), "application/json")
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept = %q, want %q", r.Header.Get("Accept"), "application/json")
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userResponse{User: ClickUpUser{ID: 1}}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	_, err := client.GetAuthorizedUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsRateLimitError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"rate limit error", &RateLimitError{RetryAfter: 10 * time.Second}, true},
		{"wrapped rate limit error", fmt.Errorf("wrapped: %w", &RateLimitError{RetryAfter: 5 * time.Second}), true},
		{"other error", errors.New("something else"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsRateLimitError(tt.err); got != tt.want {
				t.Errorf("IsRateLimitError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientContextCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{APIToken: "test-token", BaseURL: server.URL})
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.GetTasks(ctx, "list1", 0)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
