package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	gogithub "github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// RateLimitError is returned when the GitHub API responds with 403 rate limit exceeded.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("github rate limit exceeded, retry after %s", e.RetryAfter)
}

// IsRateLimitError returns true if err is a RateLimitError.
func IsRateLimitError(err error) bool {
	var rle *RateLimitError
	return errors.As(err, &rle)
}

// GitHubIssue maps relevant go-github Issue fields for internal use.
type GitHubIssue struct {
	Number    int
	Title     string
	Body      string
	State     string
	Labels    []string
	Assignee  string
	CreatedAt time.Time
	UpdatedAt time.Time
	HTMLURL   string
	Repo      string // "owner/repo" for tracking origin
}

// GitHubClient wraps the go-github SDK for issue operations.
type GitHubClient struct {
	client *gogithub.Client
}

// NewGitHubClient creates a new GitHub API client with PAT authentication.
// If config.Token is empty, an unauthenticated client is created.
func NewGitHubClient(config *GitHubConfig) *GitHubClient {
	var httpClient *http.Client
	if config.Token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})
		httpClient = oauth2.NewClient(context.Background(), ts)
	}

	return &GitHubClient{
		client: gogithub.NewClient(httpClient),
	}
}

// ListIssues returns open issues for the given repo, optionally filtered by assignee.
// Pass an empty assignee to list all open issues.
func (c *GitHubClient) ListIssues(ctx context.Context, owner, repo, assignee string) ([]*GitHubIssue, error) {
	opts := &gogithub.IssueListByRepoOptions{
		State: "open",
		ListOptions: gogithub.ListOptions{
			PerPage: 100,
		},
	}
	if assignee != "" {
		opts.Assignee = assignee
	}

	var allIssues []*GitHubIssue
	for {
		issues, resp, err := c.client.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, c.wrapError("list issues", err)
		}

		for _, issue := range issues {
			if issue.PullRequestLinks != nil {
				continue // skip PRs
			}
			allIssues = append(allIssues, mapIssue(issue, owner+"/"+repo))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

// CloseIssue sets an issue's state to "closed".
func (c *GitHubClient) CloseIssue(ctx context.Context, owner, repo string, issueNumber int) error {
	state := "closed"
	_, _, err := c.client.Issues.Edit(ctx, owner, repo, issueNumber, &gogithub.IssueRequest{
		State: &state,
	})
	if err != nil {
		return c.wrapError(fmt.Sprintf("close issue %d", issueNumber), err)
	}
	return nil
}

// GetAuthenticatedUser returns the login name of the authenticated user.
func (c *GitHubClient) GetAuthenticatedUser(ctx context.Context) (string, error) {
	user, _, err := c.client.Users.Get(ctx, "")
	if err != nil {
		return "", c.wrapError("get authenticated user", err)
	}
	return user.GetLogin(), nil
}

// wrapError converts go-github errors into adapter error types.
func (c *GitHubClient) wrapError(operation string, err error) error {
	var rateLimitErr *gogithub.RateLimitError
	if errors.As(err, &rateLimitErr) {
		retryAfter := time.Until(rateLimitErr.Rate.Reset.Time)
		if retryAfter < 0 {
			retryAfter = 60 * time.Second
		}
		return &RateLimitError{RetryAfter: retryAfter}
	}

	var abuseErr *gogithub.AbuseRateLimitError
	if errors.As(err, &abuseErr) {
		retryAfter := 60 * time.Second
		if abuseErr.RetryAfter != nil {
			retryAfter = *abuseErr.RetryAfter
		}
		return &RateLimitError{RetryAfter: retryAfter}
	}

	return fmt.Errorf("github %s: %w", operation, err)
}

// mapIssue converts a go-github Issue to a GitHubIssue.
func mapIssue(issue *gogithub.Issue, repo string) *GitHubIssue {
	gi := &GitHubIssue{
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		Body:      issue.GetBody(),
		State:     issue.GetState(),
		HTMLURL:   issue.GetHTMLURL(),
		Repo:      repo,
		CreatedAt: issue.GetCreatedAt().UTC(),
		UpdatedAt: issue.GetUpdatedAt().UTC(),
	}

	for _, label := range issue.Labels {
		gi.Labels = append(gi.Labels, label.GetName())
	}

	if issue.Assignee != nil {
		gi.Assignee = issue.Assignee.GetLogin()
	}

	return gi
}
