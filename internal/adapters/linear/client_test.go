package linear

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// isolatedHTTPClient returns an http.Client with its own transport,
// preventing CloseIdleConnections races between parallel tests.
func isolatedHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{},
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("failed to encode response: %v", err)
	}
}

func TestLinearClientQueryViewer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header format: no "Bearer" prefix
		auth := r.Header.Get("Authorization")
		if auth != "lin_api_test" {
			t.Errorf("Authorization header = %q, want %q", auth, "lin_api_test")
		}

		resp := graphQLResponse{
			Data: json.RawMessage(`{"viewer":{"id":"u1","name":"Test User","email":"test@test.com"}}`),
		}
		writeJSON(t, w, resp)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL

	viewer, err := client.QueryViewer(context.Background())
	if err != nil {
		t.Fatalf("QueryViewer() error = %v", err)
	}
	if viewer.ID != "u1" {
		t.Errorf("Viewer.ID = %q, want %q", viewer.ID, "u1")
	}
	if viewer.Name != "Test User" {
		t.Errorf("Viewer.Name = %q, want %q", viewer.Name, "Test User")
	}
}

func TestLinearClientQueryTeamIssues(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			return
		}

		teamID, _ := req.Variables["teamId"].(string)
		if teamID != "team-1" {
			t.Errorf("teamId = %q, want %q", teamID, "team-1")
		}

		resp := graphQLResponse{
			Data: json.RawMessage(`{
				"team": {
					"issues": {
						"nodes": [{
							"id": "issue-1",
							"identifier": "ENG-123",
							"title": "Test Issue",
							"description": "A test issue",
							"priority": 2,
							"estimate": 3,
							"dueDate": "2026-04-01",
							"createdAt": "2026-03-01T00:00:00Z",
							"updatedAt": "2026-03-05T00:00:00Z",
							"state": {"id": "state-1", "name": "In Progress", "type": "started"},
							"team": {"id": "team-1", "key": "ENG"},
							"labels": {"nodes": [{"name": "bug"}, {"name": "urgent"}]},
							"assignee": {"id": "user-1", "name": "Dev", "email": "dev@test.com", "isMe": true}
						}],
						"pageInfo": {"hasNextPage": false, "endCursor": ""}
					}
				}
			}`),
		}
		writeJSON(t, w, resp)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL

	conn, err := client.QueryTeamIssues(context.Background(), "team-1", "")
	if err != nil {
		t.Fatalf("QueryTeamIssues() error = %v", err)
	}
	if len(conn.Nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(conn.Nodes))
	}

	issue := conn.Nodes[0]
	if issue.Identifier != "ENG-123" {
		t.Errorf("Identifier = %q, want %q", issue.Identifier, "ENG-123")
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Title = %q, want %q", issue.Title, "Test Issue")
	}
	if issue.Priority != 2 {
		t.Errorf("Priority = %d, want 2", issue.Priority)
	}
	if issue.State.Type != "started" {
		t.Errorf("State.Type = %q, want %q", issue.State.Type, "started")
	}
	if len(issue.Labels.Nodes) != 2 {
		t.Errorf("Labels count = %d, want 2", len(issue.Labels.Nodes))
	}
	if issue.Assignee == nil || !issue.Assignee.IsMe {
		t.Error("expected assignee with IsMe=true")
	}
}

func TestLinearClientPagination(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		var hasNext bool
		var cursor string
		var issueID string

		switch count {
		case 1:
			hasNext = true
			cursor = "cursor-page2"
			issueID = "issue-page1"
		case 2:
			hasNext = false
			cursor = ""
			issueID = "issue-page2"
		default:
			t.Error("unexpected third request")
			return
		}

		resp := graphQLResponse{
			Data: json.RawMessage(fmt.Sprintf(`{
				"team": {
					"issues": {
						"nodes": [{"id": %q, "identifier": "ENG-%d", "title": "Issue %d", "priority": 0, "createdAt": "2026-03-01T00:00:00Z", "updatedAt": "2026-03-01T00:00:00Z", "state": {"id": "s1", "name": "Todo", "type": "unstarted"}, "team": {"id": "t1", "key": "ENG"}, "labels": {"nodes": []}}],
						"pageInfo": {"hasNextPage": %v, "endCursor": %q}
					}
				}
			}`, issueID, count, count, hasNext, cursor)),
		}
		writeJSON(t, w, resp)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL

	issues, err := client.FetchAllIssues(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("FetchAllIssues() error = %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("got %d issues, want 2", len(issues))
	}
	if callCount.Load() != 2 {
		t.Errorf("made %d API calls, want 2", callCount.Load())
	}
}

func TestLinearClientRateLimit(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		if count <= 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		resp := graphQLResponse{
			Data: json.RawMessage(`{"viewer":{"id":"u1","name":"Test","email":"t@t.com"}}`),
		}
		writeJSON(t, w, resp)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL
	client.httpClient = isolatedHTTPClient()
	client.sleepFn = func(_ time.Duration) {} // no-op sleep for tests

	viewer, err := client.QueryViewer(context.Background())
	if err != nil {
		t.Fatalf("QueryViewer() error = %v after retries", err)
	}
	if viewer.ID != "u1" {
		t.Errorf("Viewer.ID = %q, want %q", viewer.ID, "u1")
	}
	if callCount.Load() != 3 {
		t.Errorf("made %d calls, want 3 (2 rate-limited + 1 success)", callCount.Load())
	}
}

func TestLinearClientRateLimitExhausted(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL
	client.httpClient = isolatedHTTPClient()
	client.sleepFn = func(_ time.Duration) {}

	_, err := client.QueryViewer(context.Background())
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !IsRateLimitError(err) {
		t.Errorf("expected RateLimitError, got %T: %v", err, err)
	}
}

func TestLinearClientAuthError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("bad_key")
	client.baseURL = server.URL

	_, err := client.QueryViewer(context.Background())
	if err == nil {
		t.Fatal("expected auth error")
	}
}

func TestLinearClientGraphQLError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := graphQLResponse{
			Errors: []graphQLError{{Message: "Team not found"}},
		}
		writeJSON(t, w, resp)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL

	_, err := client.QueryTeamIssues(context.Background(), "bad-team", "")
	if err == nil {
		t.Fatal("expected GraphQL error")
	}
}

func TestLinearClientNetworkTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL
	client.httpClient.Timeout = 100 * time.Millisecond

	_, err := client.QueryViewer(context.Background())
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"empty", "", 60 * time.Second},
		{"valid seconds", "30", 30 * time.Second},
		{"invalid", "abc", 60 * time.Second},
		{"zero", "0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseRetryAfter(tt.value)
			if got != tt.want {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestLinearClientQueryWorkflowStates(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := graphQLResponse{
			Data: json.RawMessage(`{
				"team": {
					"states": {
						"nodes": [
							{"id": "s1", "name": "Triage", "type": "triage"},
							{"id": "s2", "name": "Backlog", "type": "backlog"},
							{"id": "s3", "name": "Todo", "type": "unstarted"},
							{"id": "s4", "name": "In Progress", "type": "started"},
							{"id": "s5", "name": "Done", "type": "completed"},
							{"id": "s6", "name": "Cancelled", "type": "cancelled"}
						]
					}
				}
			}`),
		}
		writeJSON(t, w, resp)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL

	states, err := client.QueryWorkflowStates(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("QueryWorkflowStates() error = %v", err)
	}
	if len(states) != 6 {
		t.Errorf("got %d states, want 6", len(states))
	}

	// Verify state types
	typeMap := make(map[string]string)
	for _, s := range states {
		typeMap[s.Type] = s.Name
	}
	expectedTypes := []string{"triage", "backlog", "unstarted", "started", "completed", "cancelled"}
	for _, et := range expectedTypes {
		if _, ok := typeMap[et]; !ok {
			t.Errorf("missing workflow state type %q", et)
		}
	}
}

func TestIsRateLimitError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"rate limit error", &RateLimitError{RetryAfter: time.Second}, true},
		{"wrapped rate limit", fmt.Errorf("wrapped: %w", &RateLimitError{RetryAfter: time.Second}), true},
		{"other error", fmt.Errorf("some error"), false},
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

func TestLinearClientMutateIssueState(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			return
		}

		issueID, _ := req.Variables["id"].(string)
		stateID, _ := req.Variables["stateId"].(string)

		if issueID != "issue-uuid-1" {
			t.Errorf("issueID = %q, want %q", issueID, "issue-uuid-1")
		}
		if stateID != "state-completed" {
			t.Errorf("stateID = %q, want %q", stateID, "state-completed")
		}

		resp := graphQLResponse{
			Data: json.RawMessage(`{
				"issueUpdate": {
					"success": true,
					"issue": {
						"id": "issue-uuid-1",
						"state": {"id": "state-completed", "name": "Done", "type": "completed"}
					}
				}
			}`),
		}
		writeJSON(t, w, resp)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL

	result, err := client.MutateIssueState(context.Background(), "issue-uuid-1", "state-completed")
	if err != nil {
		t.Fatalf("MutateIssueState() error = %v", err)
	}
	if !result.Success {
		t.Error("expected Success=true")
	}
	if result.Issue == nil {
		t.Fatal("expected issue in result")
	}
	if result.Issue.State.Type != "completed" {
		t.Errorf("State.Type = %q, want %q", result.Issue.State.Type, "completed")
	}
}

func TestLinearClientMutateIssueUpdate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			return
		}

		title, _ := req.Variables["title"].(string)
		desc, _ := req.Variables["description"].(string)

		if title != "Updated Title" {
			t.Errorf("title = %q, want %q", title, "Updated Title")
		}
		if desc != "New description" {
			t.Errorf("description = %q, want %q", desc, "New description")
		}

		resp := graphQLResponse{
			Data: json.RawMessage(`{
				"issueUpdate": {
					"success": true,
					"issue": {
						"id": "issue-uuid-1",
						"state": {"id": "s4", "name": "In Progress", "type": "started"}
					}
				}
			}`),
		}
		writeJSON(t, w, resp)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL

	result, err := client.MutateIssueUpdate(context.Background(), "issue-uuid-1", "Updated Title", "New description")
	if err != nil {
		t.Fatalf("MutateIssueUpdate() error = %v", err)
	}
	if !result.Success {
		t.Error("expected Success=true")
	}
}

func TestLinearClientMutateIssueState_GraphQLError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := graphQLResponse{
			Errors: []graphQLError{{Message: "Issue not found"}},
		}
		writeJSON(t, w, resp)
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL

	_, err := client.MutateIssueState(context.Background(), "bad-id", "state-1")
	if err == nil {
		t.Fatal("expected error for GraphQL error response")
	}
}

func TestLinearClientMalformedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`not valid json`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewLinearClient("lin_api_test")
	client.baseURL = server.URL

	_, err := client.QueryViewer(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed response")
	}
}
