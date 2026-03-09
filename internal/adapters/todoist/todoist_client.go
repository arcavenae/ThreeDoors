package todoist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	// BaseURL is the Todoist REST API v1 endpoint.
	BaseURL = "https://api.todoist.com/rest/v1"
)

// RateLimitError is returned when the Todoist API responds with 429 Too Many Requests.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("todoist rate limit exceeded, retry after %s", e.RetryAfter)
}

// IsRateLimitError returns true if err is a RateLimitError.
func IsRateLimitError(err error) bool {
	var rle *RateLimitError
	return errors.As(err, &rle)
}

// TodoistTask represents a task from the Todoist REST API.
type TodoistTask struct {
	ID          string   `json:"id"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	ProjectID   string   `json:"project_id"`
	SectionID   string   `json:"section_id"`
	Priority    int      `json:"priority"`
	Labels      []string `json:"labels"`
	Due         *DueDate `json:"due"`
	IsCompleted bool     `json:"is_completed"`
	CreatedAt   string   `json:"created_at"`
	URL         string   `json:"url"`
}

// DueDate represents a Todoist task due date.
type DueDate struct {
	Date      string `json:"date"`
	Datetime  string `json:"datetime"`
	Recurring bool   `json:"recurring"`
	String    string `json:"string"`
}

// TodoistProject represents a project from the Todoist REST API.
type TodoistProject struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Order int    `json:"order"`
}

// Client is a thin HTTP client for the Todoist REST API v1.
type Client struct {
	baseURL    string
	authHeader string
	httpClient *http.Client
}

// NewClient creates a new Todoist API client with the given auth configuration.
func NewClient(config AuthConfig) *Client {
	return &Client{
		baseURL:    BaseURL,
		authHeader: "Bearer " + config.APIToken,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetTasks retrieves active tasks, optionally filtered by project ID or filter expression.
// Only one of projectID or filter should be non-empty (enforced by config validation).
func (c *Client) GetTasks(ctx context.Context, projectID, filter string) ([]TodoistTask, error) {
	path := "/tasks"
	sep := '?'

	if projectID != "" {
		path += fmt.Sprintf("%cproject_id=%s", sep, projectID)
		sep = '&'
	}
	if filter != "" {
		path += fmt.Sprintf("%cfilter=%s", sep, filter)
	}

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("todoist get tasks: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("todoist get tasks: unexpected status %d", resp.StatusCode)
	}

	var tasks []TodoistTask
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("todoist get tasks decode: %w", err)
	}
	return tasks, nil
}

// CloseTask marks a task as completed.
func (c *Client) CloseTask(ctx context.Context, taskID string) error {
	path := fmt.Sprintf("/tasks/%s/close", taskID)

	resp, err := c.do(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("todoist close task %s: %w", taskID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("todoist close task %s: unexpected status %d", taskID, resp.StatusCode)
	}
	return nil
}

// GetProjects retrieves all projects. Useful for health checks and config validation.
func (c *Client) GetProjects(ctx context.Context) ([]TodoistProject, error) {
	resp, err := c.do(ctx, http.MethodGet, "/projects", nil)
	if err != nil {
		return nil, fmt.Errorf("todoist get projects: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("todoist get projects: unexpected status %d", resp.StatusCode)
	}

	var projects []TodoistProject
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("todoist get projects decode: %w", err)
	}
	return projects, nil
}

// do executes an HTTP request with authentication and rate limit handling.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("todoist create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("todoist http: %w", err)
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
