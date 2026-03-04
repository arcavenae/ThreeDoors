package jira

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientSearchJQL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/api/3/search/jql" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}

		var req searchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.JQL != "assignee = currentUser()" {
			t.Errorf("unexpected JQL: %s", req.JQL)
		}

		resp := SearchResult{
			Issues: []Issue{
				{
					Key: "PROJ-42",
					Fields: IssueFields{
						Summary: "Fix login bug",
						Status: IssueStatus{
							Name:           "In Progress",
							StatusCategory: StatusCategory{Key: "indeterminate", Name: "In Progress"},
						},
						Priority: &IssuePriority{Name: "High"},
						Project:  IssueProject{Key: "PROJ"},
						Labels:   []string{"backend", "auth"},
						Created:  "2026-03-01T10:00:00.000+0000",
						Updated:  "2026-03-02T14:30:00.000+0000",
					},
				},
			},
			IsLast: true,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{
		Type:     AuthBasic,
		URL:      server.URL,
		Email:    "test@example.com",
		APIToken: "test-token",
	})

	result, err := client.SearchJQL(context.Background(), "assignee = currentUser()", []string{"summary", "status"}, 50, "")
	if err != nil {
		t.Fatalf("SearchJQL: %v", err)
	}

	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(result.Issues))
	}
	if result.Issues[0].Key != "PROJ-42" {
		t.Errorf("expected key PROJ-42, got %s", result.Issues[0].Key)
	}
	if result.Issues[0].Fields.Summary != "Fix login bug" {
		t.Errorf("unexpected summary: %s", result.Issues[0].Fields.Summary)
	}
	if !result.IsLast {
		t.Error("expected IsLast to be true")
	}
}

func TestClientSearchJQLPagination(t *testing.T) {
	t.Parallel()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var req searchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		var resp SearchResult
		switch req.NextPageToken {
		case "":
			resp = SearchResult{
				Issues:        []Issue{{Key: "PROJ-1"}},
				NextPageToken: "page2token",
				IsLast:        false,
			}
		case "page2token":
			resp = SearchResult{
				Issues: []Issue{{Key: "PROJ-2"}},
				IsLast: true,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{Type: AuthBasic, URL: server.URL, Email: "e", APIToken: "t"})

	// Page 1
	result1, err := client.SearchJQL(context.Background(), "test", nil, 1, "")
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if result1.IsLast {
		t.Error("page 1 should not be last")
	}
	if result1.NextPageToken != "page2token" {
		t.Errorf("unexpected next token: %s", result1.NextPageToken)
	}

	// Page 2
	result2, err := client.SearchJQL(context.Background(), "test", nil, 1, result1.NextPageToken)
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if !result2.IsLast {
		t.Error("page 2 should be last")
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestClientGetTransitions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/api/3/issue/PROJ-42/transitions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := transitionsResponse{
			Transitions: []Transition{
				{
					ID:   "21",
					Name: "In Progress",
					To: IssueStatus{
						Name:           "In Progress",
						StatusCategory: StatusCategory{Key: "indeterminate"},
					},
				},
				{
					ID:   "31",
					Name: "Done",
					To: IssueStatus{
						Name:           "Done",
						StatusCategory: StatusCategory{Key: "done"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{Type: AuthPAT, URL: server.URL, APIToken: "test-pat"})

	transitions, err := client.GetTransitions(context.Background(), "PROJ-42")
	if err != nil {
		t.Fatalf("GetTransitions: %v", err)
	}

	if len(transitions) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(transitions))
	}
	if transitions[1].ID != "31" {
		t.Errorf("expected Done transition ID 31, got %s", transitions[1].ID)
	}
	if transitions[1].To.StatusCategory.Key != "done" {
		t.Errorf("expected done category, got %s", transitions[1].To.StatusCategory.Key)
	}
}

func TestClientDoTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		statusCode  int
		wantErr     bool
		errContains string
	}{
		{
			name:       "success",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:        "conflict",
			statusCode:  http.StatusConflict,
			wantErr:     true,
			errContains: "conflict",
		},
		{
			name:        "server error",
			statusCode:  http.StatusInternalServerError,
			wantErr:     true,
			errContains: "unexpected status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
			}))
			t.Cleanup(server.Close)

			client := NewClient(AuthConfig{Type: AuthBasic, URL: server.URL, Email: "e", APIToken: "t"})
			err := client.DoTransition(context.Background(), "PROJ-42", "31")

			if (err != nil) != tt.wantErr {
				t.Errorf("DoTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !containsStr(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestClientRateLimitHandling(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "42")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{Type: AuthBasic, URL: server.URL, Email: "e", APIToken: "t"})
	_, err := client.SearchJQL(context.Background(), "test", nil, 50, "")

	if err == nil {
		t.Fatal("expected rate limit error")
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		// No Retry-After header
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{Type: AuthBasic, URL: server.URL, Email: "e", APIToken: "t"})
	_, err := client.SearchJQL(context.Background(), "test", nil, 50, "")

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}

	if rle.RetryAfter != 60*time.Second {
		t.Errorf("expected default RetryAfter 60s, got %s", rle.RetryAfter)
	}
}

func TestClientAuthBasic(t *testing.T) {
	t.Parallel()

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(SearchResult{IsLast: true}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{Type: AuthBasic, URL: server.URL, Email: "user@test.com", APIToken: "mytoken"})
	_, err := client.SearchJQL(context.Background(), "test", nil, 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsStr(gotAuth, "Basic ") {
		t.Errorf("expected Basic auth header, got %q", gotAuth)
	}
}

func TestClientAuthPAT(t *testing.T) {
	t.Parallel()

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(SearchResult{IsLast: true}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(AuthConfig{Type: AuthPAT, URL: server.URL, APIToken: "my-pat-token"})
	_, err := client.SearchJQL(context.Background(), "test", nil, 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth != "Bearer my-pat-token" {
		t.Errorf("expected Bearer auth header, got %q", gotAuth)
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

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
