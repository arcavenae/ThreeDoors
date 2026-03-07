package github

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

func TestListIssues(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/repos/testowner/testrepo/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		if r.URL.Query().Get("state") != "open" {
			t.Errorf("expected state=open, got %s", r.URL.Query().Get("state"))
		}
		if r.URL.Query().Get("assignee") != "testuser" {
			t.Errorf("expected assignee=testuser, got %s", r.URL.Query().Get("assignee"))
		}

		issues := []map[string]any{
			{
				"number":     42,
				"title":      "Fix login bug",
				"body":       "Users can't login",
				"state":      "open",
				"html_url":   "https://github.com/testowner/testrepo/issues/42",
				"created_at": "2026-03-01T10:00:00Z",
				"updated_at": "2026-03-02T14:30:00Z",
				"labels": []map[string]any{
					{"name": "bug"},
					{"name": "high-priority"},
				},
				"assignee": map[string]any{
					"login": "testuser",
				},
			},
			{
				"number":     43,
				"title":      "Add feature",
				"state":      "open",
				"html_url":   "https://github.com/testowner/testrepo/issues/43",
				"created_at": "2026-03-02T10:00:00Z",
				"updated_at": "2026-03-03T14:30:00Z",
				"labels":     []map[string]any{},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(issues); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server.URL)
	issues, err := client.ListIssues(context.Background(), "testowner", "testrepo", "testuser")
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}

	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues))
	}

	issue := issues[0]
	if issue.Number != 42 {
		t.Errorf("number = %d, want 42", issue.Number)
	}
	if issue.Title != "Fix login bug" {
		t.Errorf("title = %q, want %q", issue.Title, "Fix login bug")
	}
	if issue.Body != "Users can't login" {
		t.Errorf("body = %q, want %q", issue.Body, "Users can't login")
	}
	if issue.State != "open" {
		t.Errorf("state = %q, want %q", issue.State, "open")
	}
	if issue.Repo != "testowner/testrepo" {
		t.Errorf("repo = %q, want %q", issue.Repo, "testowner/testrepo")
	}
	if len(issue.Labels) != 2 {
		t.Fatalf("labels count = %d, want 2", len(issue.Labels))
	}
	if issue.Labels[0] != "bug" {
		t.Errorf("labels[0] = %q, want %q", issue.Labels[0], "bug")
	}
	if issue.Assignee != "testuser" {
		t.Errorf("assignee = %q, want %q", issue.Assignee, "testuser")
	}
	if issue.HTMLURL != "https://github.com/testowner/testrepo/issues/42" {
		t.Errorf("html_url = %q", issue.HTMLURL)
	}

	// Second issue has no assignee
	if issues[1].Assignee != "" {
		t.Errorf("issue 43 assignee = %q, want empty", issues[1].Assignee)
	}
}

func TestListIssuesSkipsPRs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issues := []map[string]any{
			{
				"number":     1,
				"title":      "Real issue",
				"state":      "open",
				"created_at": "2026-03-01T10:00:00Z",
				"updated_at": "2026-03-01T10:00:00Z",
			},
			{
				"number":     2,
				"title":      "A pull request",
				"state":      "open",
				"created_at": "2026-03-01T10:00:00Z",
				"updated_at": "2026-03-01T10:00:00Z",
				"pull_request": map[string]any{
					"url": "https://api.github.com/repos/o/r/pulls/2",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(issues); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server.URL)
	issues, err := client.ListIssues(context.Background(), "o", "r", "")
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue (PR filtered), got %d", len(issues))
	}
	if issues[0].Title != "Real issue" {
		t.Errorf("expected 'Real issue', got %q", issues[0].Title)
	}
}

func TestListIssuesPagination(t *testing.T) {
	t.Parallel()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		page := r.URL.Query().Get("page")

		var issues []map[string]any
		switch page {
		case "", "0", "1":
			issues = []map[string]any{
				{
					"number":     1,
					"title":      "Issue 1",
					"state":      "open",
					"created_at": "2026-03-01T10:00:00Z",
					"updated_at": "2026-03-01T10:00:00Z",
				},
			}
			w.Header().Set("Link", `<`+r.URL.Path+`?page=2>; rel="next"`)
		case "2":
			issues = []map[string]any{
				{
					"number":     2,
					"title":      "Issue 2",
					"state":      "open",
					"created_at": "2026-03-01T10:00:00Z",
					"updated_at": "2026-03-01T10:00:00Z",
				},
			}
			// No Link header = last page
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(issues); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server.URL)
	issues, err := client.ListIssues(context.Background(), "o", "r", "")
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}

	if len(issues) != 2 {
		t.Fatalf("expected 2 issues across pages, got %d", len(issues))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestCloseIssue(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/repos/testowner/testrepo/issues/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req["state"] != "closed" {
			t.Errorf("expected state=closed, got %v", req["state"])
		}

		resp := map[string]any{
			"number":     42,
			"state":      "closed",
			"title":      "Fixed bug",
			"created_at": "2026-03-01T10:00:00Z",
			"updated_at": "2026-03-03T10:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server.URL)
	err := client.CloseIssue(context.Background(), "testowner", "testrepo", 42)
	if err != nil {
		t.Fatalf("CloseIssue: %v", err)
	}
}

func TestCloseIssueNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]any{
			"message": "Not Found",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server.URL)
	err := client.CloseIssue(context.Background(), "o", "r", 999)
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestGetAuthenticatedUser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/user" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := map[string]any{
			"login": "testuser",
			"id":    12345,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server.URL)
	login, err := client.GetAuthenticatedUser(context.Background())
	if err != nil {
		t.Fatalf("GetAuthenticatedUser: %v", err)
	}
	if login != "testuser" {
		t.Errorf("login = %q, want %q", login, "testuser")
	}
}

func TestRateLimitHandling(t *testing.T) {
	t.Parallel()

	resetTime := time.Now().UTC().Add(30 * time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"message":           "API rate limit exceeded",
			"documentation_url": "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting",
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server.URL)
	_, err := client.ListIssues(context.Background(), "o", "r", "")
	if err == nil {
		t.Fatal("expected rate limit error")
	}

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}

	if rle.RetryAfter <= 0 {
		t.Errorf("expected positive RetryAfter, got %s", rle.RetryAfter)
	}

	if !IsRateLimitError(err) {
		t.Error("IsRateLimitError should return true")
	}
}

func TestRateLimitErrorString(t *testing.T) {
	t.Parallel()

	err := &RateLimitError{RetryAfter: 42 * time.Second}
	want := "github rate limit exceeded, retry after 42s"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestIsRateLimitErrorFalse(t *testing.T) {
	t.Parallel()

	if IsRateLimitError(errors.New("some other error")) {
		t.Error("IsRateLimitError should return false for non-rate-limit errors")
	}
	if IsRateLimitError(nil) {
		t.Error("IsRateLimitError should return false for nil")
	}
}

func TestNewGitHubClientNoToken(t *testing.T) {
	t.Parallel()

	cfg := &GitHubConfig{
		Repos: []string{"owner/repo"},
	}
	client := NewGitHubClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.client == nil {
		t.Fatal("expected non-nil underlying go-github client")
	}
}

func TestNewGitHubClientWithToken(t *testing.T) {
	t.Parallel()

	cfg := &GitHubConfig{
		Token: "test-token",
		Repos: []string{"owner/repo"},
	}
	client := NewGitHubClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestMapIssueTimestampsUTC(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issues := []map[string]any{
			{
				"number":     1,
				"title":      "Test",
				"state":      "open",
				"created_at": "2026-03-01T10:00:00-05:00",
				"updated_at": "2026-03-02T08:00:00+02:00",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(issues); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server.URL)
	issues, err := client.ListIssues(context.Background(), "o", "r", "")
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].CreatedAt.Location().String() != "UTC" {
		t.Errorf("CreatedAt not UTC: %v", issues[0].CreatedAt.Location())
	}
	if issues[0].UpdatedAt.Location().String() != "UTC" {
		t.Errorf("UpdatedAt not UTC: %v", issues[0].UpdatedAt.Location())
	}
}

// newTestClient creates a GitHubClient pointing at a test server.
func newTestClient(t *testing.T, baseURL string) *GitHubClient {
	t.Helper()
	cfg := &GitHubConfig{
		Token: "test-token",
		Repos: []string{"testowner/testrepo"},
	}
	client := NewGitHubClient(cfg)
	client.client.BaseURL, _ = client.client.BaseURL.Parse(baseURL + "/")
	return client
}
