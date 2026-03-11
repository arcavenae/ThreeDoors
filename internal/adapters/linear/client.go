package linear

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

const (
	// BaseURL is the Linear GraphQL API endpoint.
	BaseURL = "https://api.linear.app/graphql"

	// maxRateLimitRetries is the maximum retry count for rate-limited requests.
	maxRateLimitRetries = 3

	// defaultPageSize is the number of issues per GraphQL page.
	defaultPageSize = 50
)

// RateLimitError is returned when the Linear API responds with 429 Too Many Requests.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("linear rate limit exceeded, retry after %s", e.RetryAfter)
}

// IsRateLimitError returns true if err is a RateLimitError.
func IsRateLimitError(err error) bool {
	var rle *RateLimitError
	return errors.As(err, &rle)
}

// graphQLRequest is the JSON body sent to the Linear GraphQL endpoint.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// graphQLResponse wraps a raw GraphQL response with possible errors.
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors,omitempty"`
}

// graphQLError represents a single error in a GraphQL response.
type graphQLError struct {
	Message string `json:"message"`
}

// GraphQLClient abstracts the Linear GraphQL API for testability.
type GraphQLClient interface {
	// QueryViewer returns the authenticated user info.
	QueryViewer(ctx context.Context) (*Viewer, error)
	// QueryTeamIssues returns paginated issues for a team.
	// Pass an empty cursor for the first page.
	QueryTeamIssues(ctx context.Context, teamID, cursor string) (*IssueConnection, error)
	// QueryWorkflowStates returns all workflow states for a team.
	QueryWorkflowStates(ctx context.Context, teamID string) ([]WorkflowState, error)
}

// LinearClient implements GraphQLClient using HTTP requests to the Linear GraphQL API.
type LinearClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	sleepFn    func(time.Duration) // injectable for testing
}

// NewLinearClient creates a new Linear GraphQL API client.
func NewLinearClient(apiKey string) *LinearClient {
	return &LinearClient{
		baseURL:    BaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		sleepFn:    time.Sleep,
	}
}

// QueryViewer queries the authenticated user info (AC5).
func (c *LinearClient) QueryViewer(ctx context.Context) (*Viewer, error) {
	const query = `query { viewer { id name email } }`

	data, err := c.executeWithRetry(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("query viewer: %w", err)
	}

	var resp ViewerResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("query viewer decode: %w", err)
	}
	return &resp.Viewer, nil
}

// QueryTeamIssues fetches a page of issues for a team with cursor-based pagination (AC2).
func (c *LinearClient) QueryTeamIssues(ctx context.Context, teamID, cursor string) (*IssueConnection, error) {
	const query = `query TeamIssues($teamId: String!, $first: Int!, $after: String) {
  team(id: $teamId) {
    issues(first: $first, after: $after, filter: { state: { type: { nin: ["completed", "cancelled"] } } }) {
      nodes {
        id
        identifier
        title
        description
        priority
        estimate
        dueDate
        createdAt
        updatedAt
        state { id name type }
        team { id key }
        labels { nodes { name } }
        assignee { id name email isMe }
      }
      pageInfo { hasNextPage endCursor }
    }
  }
}`

	vars := map[string]any{
		"teamId": teamID,
		"first":  defaultPageSize,
	}
	if cursor != "" {
		vars["after"] = cursor
	}

	data, err := c.executeWithRetry(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("query team issues %s: %w", teamID, err)
	}

	var resp IssuesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("query team issues decode: %w", err)
	}
	return &resp.Team.Issues, nil
}

// QueryWorkflowStates fetches all workflow states for a team.
func (c *LinearClient) QueryWorkflowStates(ctx context.Context, teamID string) ([]WorkflowState, error) {
	const query = `query WorkflowStates($teamId: String!) {
  team(id: $teamId) {
    states { nodes { id name type } }
  }
}`

	vars := map[string]any{"teamId": teamID}

	data, err := c.executeWithRetry(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("query workflow states %s: %w", teamID, err)
	}

	var resp WorkflowStatesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("query workflow states decode: %w", err)
	}
	return resp.Team.States.Nodes, nil
}

// FetchAllIssues fetches all issues for a team using cursor-based pagination (AC2).
func (c *LinearClient) FetchAllIssues(ctx context.Context, teamID string) ([]IssueNode, error) {
	var all []IssueNode
	cursor := ""

	for {
		conn, err := c.QueryTeamIssues(ctx, teamID, cursor)
		if err != nil {
			return nil, err
		}

		all = append(all, conn.Nodes...)

		if !conn.PageInfo.HasNextPage {
			break
		}
		cursor = conn.PageInfo.EndCursor
	}

	return all, nil
}

// executeWithRetry executes a GraphQL query with rate limit retry and exponential backoff (AC7).
func (c *LinearClient) executeWithRetry(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error) {
	var lastErr error

	for attempt := range maxRateLimitRetries {
		data, err := c.execute(ctx, query, variables)
		if err == nil {
			return data, nil
		}

		var rle *RateLimitError
		if !errors.As(err, &rle) {
			return nil, err
		}

		lastErr = err
		if attempt < maxRateLimitRetries-1 {
			backoff := rle.RetryAfter
			if backoff <= 0 {
				backoff = time.Duration(1<<uint(attempt)) * time.Second
			}
			c.sleepFn(backoff)
		}
	}

	return nil, lastErr
}

// execute sends a single GraphQL request and returns the data field.
func (c *LinearClient) execute(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error) {
	reqBody := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("linear marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("linear create request: %w", err)
	}

	// Linear auth: no "Bearer" prefix (AC1)
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("linear http: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle rate limiting (AC7)
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, &RateLimitError{RetryAfter: retryAfter}
	}

	// Handle auth errors
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("linear auth failed: invalid API key (HTTP 401)")
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("linear unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("linear decode response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("linear graphql error: %s", gqlResp.Errors[0].Message)
	}

	return gqlResp.Data, nil
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
