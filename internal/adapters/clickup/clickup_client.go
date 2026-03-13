package clickup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const defaultBaseURL = "https://api.clickup.com/api/v2"

// RateLimitError is returned when the ClickUp API responds with 429 Too Many Requests.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("clickup rate limit exceeded, retry after %s", e.RetryAfter)
}

// IsRateLimitError returns true if err is a RateLimitError.
func IsRateLimitError(err error) bool {
	var rle *RateLimitError
	return errors.As(err, &rle)
}

// ClickUpTask represents a task from the ClickUp API.
type ClickUpTask struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Status ClickUpStatus  `json:"status"`
	URL    string         `json:"url"`
	List   ClickUpListRef `json:"list"`
}

// ClickUpStatus represents a task's status in ClickUp.
type ClickUpStatus struct {
	Status string `json:"status"`
	Type   string `json:"type"`
}

// ClickUpListRef is a lightweight reference to a list within a task response.
type ClickUpListRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ClickUpSpace represents a space in the ClickUp hierarchy.
type ClickUpSpace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ClickUpList represents a list in the ClickUp hierarchy.
type ClickUpList struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ClickUpUser represents an authenticated ClickUp user.
type ClickUpUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// tasksResponse wraps the ClickUp GET tasks response.
type tasksResponse struct {
	Tasks []ClickUpTask `json:"tasks"`
}

// spacesResponse wraps the ClickUp GET spaces response.
type spacesResponse struct {
	Spaces []ClickUpSpace `json:"spaces"`
}

// listsResponse wraps the ClickUp GET lists response.
type listsResponse struct {
	Lists []ClickUpList `json:"lists"`
}

// userResponse wraps the ClickUp GET user response.
type userResponse struct {
	User ClickUpUser `json:"user"`
}

// AuthConfig holds authentication settings for the ClickUp client.
type AuthConfig struct {
	APIToken string `yaml:"-"`
	BaseURL  string // Override for testing; defaults to https://api.clickup.com/api/v2
}

// Client is a thin HTTP client for the ClickUp REST API v2.
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new ClickUp API client with the given auth configuration.
func NewClient(config AuthConfig) *Client {
	base := config.BaseURL
	if base == "" {
		base = defaultBaseURL
	}

	return &Client{
		baseURL:    base,
		apiToken:   config.APIToken,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetTasks retrieves tasks from a ClickUp list with page-based pagination (0-indexed).
func (c *Client) GetTasks(ctx context.Context, listID string, page int) ([]ClickUpTask, error) {
	path := fmt.Sprintf("/list/%s/task?page=%d", listID, page)

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("clickup get tasks list %s: %w", listID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clickup get tasks list %s: unexpected status %d", listID, resp.StatusCode)
	}

	var result tasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("clickup get tasks list %s decode: %w", listID, err)
	}
	return result.Tasks, nil
}

// UpdateTaskStatus updates the status of a task in ClickUp.
func (c *Client) UpdateTaskStatus(ctx context.Context, taskID, status string) error {
	path := fmt.Sprintf("/task/%s", taskID)

	body := struct {
		Status string `json:"status"`
	}{Status: status}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("clickup update task %s marshal: %w", taskID, err)
	}

	resp, err := c.do(ctx, http.MethodPut, path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("clickup update task %s: %w", taskID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("clickup update task %s: unexpected status %d", taskID, resp.StatusCode)
	}
	return nil
}

// GetSpaces retrieves all spaces in a ClickUp workspace/team.
func (c *Client) GetSpaces(ctx context.Context, teamID string) ([]ClickUpSpace, error) {
	path := fmt.Sprintf("/team/%s/space", teamID)

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("clickup get spaces team %s: %w", teamID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clickup get spaces team %s: unexpected status %d", teamID, resp.StatusCode)
	}

	var result spacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("clickup get spaces team %s decode: %w", teamID, err)
	}
	return result.Spaces, nil
}

// GetLists retrieves all lists in a ClickUp folder.
func (c *Client) GetLists(ctx context.Context, folderID string) ([]ClickUpList, error) {
	path := fmt.Sprintf("/folder/%s/list", folderID)

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("clickup get lists folder %s: %w", folderID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clickup get lists folder %s: unexpected status %d", folderID, resp.StatusCode)
	}

	var result listsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("clickup get lists folder %s decode: %w", folderID, err)
	}
	return result.Lists, nil
}

// GetAuthorizedUser retrieves the currently authenticated user for health checks.
func (c *Client) GetAuthorizedUser(ctx context.Context) (*ClickUpUser, error) {
	resp, err := c.do(ctx, http.MethodGet, "/user", nil)
	if err != nil {
		return nil, fmt.Errorf("clickup get user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clickup get user: unexpected status %d", resp.StatusCode)
	}

	var result userResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("clickup get user decode: %w", err)
	}
	return &result.User, nil
}

// do executes an HTTP request with authentication and rate limit handling.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("clickup create request: %w", err)
	}

	req.Header.Set("Authorization", c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("clickup http: %w", err)
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
		return 60 * time.Second
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 60 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
