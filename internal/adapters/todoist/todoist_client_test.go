package todoist

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientGetTasks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/tasks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		tasks := []TodoistTask{
			{
				ID:          "12345",
				Content:     "Buy groceries",
				Description: "Milk, eggs, bread",
				ProjectID:   "2233",
				Priority:    3,
				Labels:      []string{"errands"},
				CreatedAt:   "2026-03-01T10:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tasks); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)

	tasks, err := client.GetTasks(context.Background(), "", "")
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
		return
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != "12345" {
		t.Errorf("expected task ID 12345, got %s", tasks[0].ID)
	}
	if tasks[0].Content != "Buy groceries" {
		t.Errorf("unexpected content: %s", tasks[0].Content)
	}
	if tasks[0].Description != "Milk, eggs, bread" {
		t.Errorf("unexpected description: %s", tasks[0].Description)
	}
}

func TestClientGetTasksWithProjectID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("project_id") != "2233" {
			t.Errorf("expected project_id=2233, got %s", r.URL.Query().Get("project_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]TodoistTask{}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.GetTasks(context.Background(), "2233", "")
	if err != nil {
		t.Fatalf("GetTasks with project_id: %v", err)
		return
	}
}

func TestClientGetTasksWithFilter(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("filter") != "today" {
			t.Errorf("expected filter=today, got %s", r.URL.Query().Get("filter"))
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]TodoistTask{}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.GetTasks(context.Background(), "", "today")
	if err != nil {
		t.Fatalf("GetTasks with filter: %v", err)
		return
	}
}

func TestClientGetTasksWithDueDate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tasks := []TodoistTask{
			{
				ID:      "99",
				Content: "Deadline task",
				Due: &DueDate{
					Date:      "2026-03-15",
					Datetime:  "2026-03-15T09:00:00Z",
					Recurring: false,
					String:    "Mar 15",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tasks); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	tasks, err := client.GetTasks(context.Background(), "", "")
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
		return
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Due == nil {
		t.Fatal("expected due date, got nil")
	}
	if tasks[0].Due.Date != "2026-03-15" {
		t.Errorf("unexpected due date: %s", tasks[0].Due.Date)
	}
}

func TestClientCloseTask(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/tasks/12345/close" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	err := client.CloseTask(context.Background(), "12345")
	if err != nil {
		t.Fatalf("CloseTask: %v", err)
		return
	}
}

func TestClientCloseTaskError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	err := client.CloseTask(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
		return
	}
}

func TestClientGetProjects(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		projects := []TodoistProject{
			{ID: "100", Name: "Inbox", Color: "grey", Order: 0},
			{ID: "200", Name: "Work", Color: "blue", Order: 1},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(projects); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	projects, err := client.GetProjects(context.Background())
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
		return
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Name != "Inbox" {
		t.Errorf("unexpected project name: %s", projects[0].Name)
	}
	if projects[1].ID != "200" {
		t.Errorf("unexpected project ID: %s", projects[1].ID)
	}
}

func TestClientRateLimitHandling(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "42")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.GetTasks(context.Background(), "", "")

	if err == nil {
		t.Fatal("expected rate limit error")
		return
	}

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}

	if rle.RetryAfter != 42*time.Second {
		t.Errorf("expected RetryAfter 42s, got %s", rle.RetryAfter)
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

	client := newTestClient(t, server)
	_, err := client.GetTasks(context.Background(), "", "")

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}

	if rle.RetryAfter != 60*time.Second {
		t.Errorf("expected default RetryAfter 60s, got %s", rle.RetryAfter)
	}
}

func TestClientBearerAuth(t *testing.T) {
	t.Parallel()

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]TodoistProject{}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.GetProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if gotAuth != "Bearer test-token" {
		t.Errorf("expected Bearer auth header, got %q", gotAuth)
	}
}

func TestClientGetTasksServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.GetTasks(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for 500 response")
		return
	}
}

func TestClientGetProjectsServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.GetProjects(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
		return
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

// newTestClient creates a Client pointing at the given test server.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client := NewClient(AuthConfig{APIToken: "test-token"})
	client.baseURL = server.URL
	return client
}
