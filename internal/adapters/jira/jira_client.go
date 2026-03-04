package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AuthType identifies the authentication method for a Jira instance.
type AuthType string

const (
	// AuthBasic uses email + API token for Jira Cloud.
	AuthBasic AuthType = "basic"
	// AuthPAT uses a Personal Access Token for Jira Server/DC.
	AuthPAT AuthType = "pat"
)

// AuthConfig holds authentication settings for the Jira client.
type AuthConfig struct {
	Type     AuthType
	URL      string // Base URL (e.g., "https://company.atlassian.net")
	Email    string // Cloud only
	APIToken string // Cloud: API token, Server: PAT
}

// RateLimitError is returned when the Jira API responds with 429 Too Many Requests.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("jira rate limit exceeded, retry after %s", e.RetryAfter)
}

// Issue represents a Jira issue from a search result.
type Issue struct {
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

// IssueFields contains the fields of a Jira issue.
type IssueFields struct {
	Summary   string         `json:"summary"`
	Status    IssueStatus    `json:"status"`
	Priority  *IssuePriority `json:"priority,omitempty"`
	Project   IssueProject   `json:"project"`
	Labels    []string       `json:"labels"`
	IssueType *IssueType     `json:"issuetype,omitempty"`
	Created   string         `json:"created"`
	Updated   string         `json:"updated"`
}

// IssueStatus represents a Jira issue status.
type IssueStatus struct {
	Name           string         `json:"name"`
	StatusCategory StatusCategory `json:"statusCategory"`
}

// StatusCategory is the broad category of a Jira status.
type StatusCategory struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// IssuePriority represents a Jira priority level.
type IssuePriority struct {
	Name string `json:"name"`
}

// IssueProject represents a Jira project.
type IssueProject struct {
	Key string `json:"key"`
}

// IssueType represents a Jira issue type.
type IssueType struct {
	Name string `json:"name"`
}

// Transition represents a possible status transition for a Jira issue.
type Transition struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	To   IssueStatus `json:"to"`
}

// SearchResult holds the response from a Jira JQL search.
type SearchResult struct {
	Issues        []Issue `json:"issues"`
	NextPageToken string  `json:"nextPageToken,omitempty"`
	IsLast        bool    `json:"isLast"`
}

type transitionsResponse struct {
	Transitions []Transition `json:"transitions"`
}

type searchRequest struct {
	JQL           string   `json:"jql"`
	Fields        []string `json:"fields"`
	MaxResults    int      `json:"maxResults"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}

type transitionRequest struct {
	Transition transitionID `json:"transition"`
}

type transitionID struct {
	ID string `json:"id"`
}

// Client is a thin HTTP client for the Jira REST API v3.
type Client struct {
	baseURL    string
	authHeader string
	httpClient *http.Client
}

// NewClient creates a new Jira API client with the given auth configuration.
func NewClient(config AuthConfig) *Client {
	var authHeader string
	switch config.Type {
	case AuthBasic:
		encoded := base64.StdEncoding.EncodeToString([]byte(config.Email + ":" + config.APIToken))
		authHeader = "Basic " + encoded
	case AuthPAT:
		authHeader = "Bearer " + config.APIToken
	}

	return &Client{
		baseURL:    strings.TrimRight(config.URL, "/"),
		authHeader: authHeader,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SearchJQL searches for issues using a JQL query.
// Pass an empty pageToken for the first page.
func (c *Client) SearchJQL(ctx context.Context, jql string, fields []string, maxResults int, pageToken string) (*SearchResult, error) {
	reqBody := searchRequest{
		JQL:        jql,
		Fields:     fields,
		MaxResults: maxResults,
	}
	if pageToken != "" {
		reqBody.NextPageToken = pageToken
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("jira search marshal: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/rest/api/3/search/jql", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("jira search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira search: unexpected status %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("jira search decode: %w", err)
	}
	return &result, nil
}

// GetTransitions returns the available status transitions for an issue.
func (c *Client) GetTransitions(ctx context.Context, issueKey string) ([]Transition, error) {
	path := fmt.Sprintf("/rest/api/3/issue/%s/transitions", issueKey)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("jira get transitions %s: %w", issueKey, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira get transitions %s: unexpected status %d", issueKey, resp.StatusCode)
	}

	var result transitionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("jira get transitions %s decode: %w", issueKey, err)
	}
	return result.Transitions, nil
}

// DoTransition executes a status transition on an issue.
// Returns nil on success (HTTP 204). Returns an error on failure,
// including a specific message for 409 Conflict (concurrent transition).
func (c *Client) DoTransition(ctx context.Context, issueKey, transitionID string) error {
	path := fmt.Sprintf("/rest/api/3/issue/%s/transitions", issueKey)
	reqBody := transitionRequest{
		Transition: transitionIDStruct{ID: transitionID},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("jira transition %s marshal: %w", issueKey, err)
	}

	resp, err := c.do(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("jira transition %s: %w", issueKey, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusConflict:
		return fmt.Errorf("jira transition %s: conflict (concurrent transition in progress)", issueKey)
	default:
		return fmt.Errorf("jira transition %s: unexpected status %d", issueKey, resp.StatusCode)
	}
}

type transitionIDStruct = transitionID

// do executes an HTTP request with authentication and rate limit handling.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("jira create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jira http: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		_ = resp.Body.Close()
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, &RateLimitError{RetryAfter: retryAfter}
	}

	return resp, nil
}

// parseRetryAfter parses the Retry-After header value as seconds.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 60 * time.Second // default fallback
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 60 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

// IsRateLimitError returns true if err is a RateLimitError.
func IsRateLimitError(err error) bool {
	var rle *RateLimitError
	return errors.As(err, &rle)
}
