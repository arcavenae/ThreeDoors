package jira

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

const (
	providerName   = "jira"
	searchPageSize = 50
)

// Searcher abstracts the Jira API operations needed by the provider.
// This enables testing without hitting a real Jira instance.
type Searcher interface {
	SearchJQL(ctx context.Context, jql string, fields []string, maxResults int, pageToken string) (*SearchResult, error)
	GetTransitions(ctx context.Context, issueKey string) ([]Transition, error)
	DoTransition(ctx context.Context, issueKey, transitionID string) error
}

// FieldMapper maps Jira issue fields to ThreeDoors task fields.
type FieldMapper struct {
	statusMap     map[string]core.TaskStatus
	effortMap     map[string]core.TaskEffort
	defaultEffort core.TaskEffort
}

// DefaultFieldMapper returns a FieldMapper with standard Jira-to-ThreeDoors mappings.
func DefaultFieldMapper() *FieldMapper {
	return &FieldMapper{
		statusMap: map[string]core.TaskStatus{
			"new":           core.StatusTodo,
			"undefined":     core.StatusTodo,
			"indeterminate": core.StatusInProgress,
			"done":          core.StatusComplete,
		},
		effortMap: map[string]core.TaskEffort{
			"Highest": core.EffortDeepWork,
			"High":    core.EffortDeepWork,
			"Medium":  core.EffortMedium,
			"Low":     core.EffortQuickWin,
			"Lowest":  core.EffortQuickWin,
		},
		defaultEffort: core.EffortMedium,
	}
}

// MapStatus converts a Jira statusCategory key to a ThreeDoors TaskStatus.
func (fm *FieldMapper) MapStatus(categoryKey string) core.TaskStatus {
	if status, ok := fm.statusMap[categoryKey]; ok {
		return status
	}
	return core.StatusTodo
}

// MapEffort converts a Jira priority name to a ThreeDoors TaskEffort.
func (fm *FieldMapper) MapEffort(priorityName string) core.TaskEffort {
	if effort, ok := fm.effortMap[priorityName]; ok {
		return effort
	}
	return fm.defaultEffort
}

// MapContext builds a ThreeDoors context string from project key and labels.
func (fm *FieldMapper) MapContext(projectKey string, labels []string) string {
	if len(labels) > 0 {
		return fmt.Sprintf("[%s] %s", projectKey, strings.Join(labels, ", "))
	}
	return fmt.Sprintf("[%s]", projectKey)
}

// MapIssueToTask converts a Jira Issue to a ThreeDoors Task.
func (fm *FieldMapper) MapIssueToTask(issue Issue) *core.Task {
	now := time.Now().UTC()

	priorityName := ""
	if issue.Fields.Priority != nil {
		priorityName = issue.Fields.Priority.Name
	}

	return &core.Task{
		ID:             issue.Key,
		Text:           issue.Fields.Summary,
		Context:        fm.MapContext(issue.Fields.Project.Key, issue.Fields.Labels),
		Status:         fm.MapStatus(issue.Fields.Status.StatusCategory.Key),
		Effort:         fm.MapEffort(priorityName),
		SourceProvider: providerName,
		SourceRefs: []core.SourceRef{
			{Provider: providerName, NativeID: issue.Key},
		},
		Notes:     []core.TaskNote{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// JiraProvider implements core.TaskProvider as a read-only adapter for Jira.
type JiraProvider struct {
	searcher Searcher
	jql      string
	mapper   *FieldMapper
}

// NewJiraProvider creates a JiraProvider with the given searcher, JQL query, and field mapper.
func NewJiraProvider(searcher Searcher, jql string, mapper *FieldMapper) *JiraProvider {
	return &JiraProvider{
		searcher: searcher,
		jql:      jql,
		mapper:   mapper,
	}
}

// Name returns the provider identifier.
func (p *JiraProvider) Name() string {
	return providerName
}

// LoadTasks executes the configured JQL query, paginates all results,
// and maps them to ThreeDoors tasks.
func (p *JiraProvider) LoadTasks() ([]*core.Task, error) {
	ctx := context.Background()
	fields := []string{"summary", "status", "priority", "project", "labels", "issuetype", "created", "updated"}

	var allTasks []*core.Task
	pageToken := ""

	for {
		result, err := p.searcher.SearchJQL(ctx, p.jql, fields, searchPageSize, pageToken)
		if err != nil {
			return nil, fmt.Errorf("jira load tasks: %w", err)
		}

		for _, issue := range result.Issues {
			allTasks = append(allTasks, p.mapper.MapIssueToTask(issue))
		}

		if result.IsLast || result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	if allTasks == nil {
		allTasks = []*core.Task{}
	}

	return allTasks, nil
}

// SaveTask returns ErrReadOnly; Jira provider is read-only.
func (p *JiraProvider) SaveTask(_ *core.Task) error {
	return core.ErrReadOnly
}

// SaveTasks returns ErrReadOnly; Jira provider is read-only.
func (p *JiraProvider) SaveTasks(_ []*core.Task) error {
	return core.ErrReadOnly
}

// DeleteTask returns ErrReadOnly; Jira provider is read-only.
func (p *JiraProvider) DeleteTask(_ string) error {
	return core.ErrReadOnly
}

// MarkComplete returns ErrReadOnly; Jira provider is read-only.
func (p *JiraProvider) MarkComplete(_ string) error {
	return core.ErrReadOnly
}

// Watch returns nil; Jira provider does not support file watching.
func (p *JiraProvider) Watch() <-chan core.ChangeEvent {
	return nil
}

// HealthCheck tests API connectivity by executing a minimal search.
func (p *JiraProvider) HealthCheck() core.HealthCheckResult {
	start := time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := p.searcher.SearchJQL(ctx, p.jql, []string{"summary"}, 1, "")

	duration := time.Since(start)

	if err != nil {
		return core.HealthCheckResult{
			Items: []core.HealthCheckItem{
				{
					Name:       "jira_connectivity",
					Status:     core.HealthFail,
					Message:    fmt.Sprintf("Jira API unreachable: %v", err),
					Suggestion: "Check Jira URL, credentials, and network connectivity",
				},
			},
			Overall:  core.HealthFail,
			Duration: duration,
		}
	}

	return core.HealthCheckResult{
		Items: []core.HealthCheckItem{
			{
				Name:    "jira_connectivity",
				Status:  core.HealthOK,
				Message: "Jira API reachable",
			},
		},
		Overall:  core.HealthOK,
		Duration: duration,
	}
}

// Factory creates a JiraProvider from a ProviderConfig.
// Required settings: url, api_token, jql.
// Optional settings: email (required for basic auth), auth_type (default: "basic").
func Factory(config *core.ProviderConfig) (core.TaskProvider, error) {
	settings := findJiraSettings(config)
	if settings == nil {
		return nil, fmt.Errorf("jira factory: no jira provider settings found")
	}

	url := settings["url"]
	if url == "" {
		return nil, fmt.Errorf("jira factory: missing required setting 'url'")
	}

	apiToken := settings["api_token"]
	if apiToken == "" {
		return nil, fmt.Errorf("jira factory: missing required setting 'api_token'")
	}

	jql := settings["jql"]
	if jql == "" {
		return nil, fmt.Errorf("jira factory: missing required setting 'jql'")
	}

	authType := AuthBasic
	if settings["auth_type"] == "pat" {
		authType = AuthPAT
	}

	authConfig := AuthConfig{
		Type:     authType,
		URL:      url,
		Email:    settings["email"],
		APIToken: apiToken,
	}

	client := NewClient(authConfig)
	return NewJiraProvider(client, jql, DefaultFieldMapper()), nil
}

// findJiraSettings locates the jira provider entry in the config.
func findJiraSettings(config *core.ProviderConfig) map[string]string {
	if config == nil {
		return nil
	}
	for _, entry := range config.Providers {
		if entry.Name == providerName {
			return entry.Settings
		}
	}
	return nil
}
