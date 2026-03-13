package adapters_test

import (
	"context"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/adapters"
	"github.com/arcaven/ThreeDoors/internal/adapters/linear"
	"github.com/arcaven/ThreeDoors/internal/core"
)

// linearMockClient implements linear.GraphQLClient for contract tests.
// It returns pre-loaded issues so that LoadTasks returns seeded tasks.
type linearMockClient struct {
	issues []linear.IssueNode
}

func (m *linearMockClient) QueryViewer(_ context.Context) (*linear.Viewer, error) {
	return &linear.Viewer{ID: "u1", Name: "Contract Test", Email: "test@example.com"}, nil
}

func (m *linearMockClient) QueryTeamIssues(_ context.Context, _, _ string) (*linear.IssueConnection, error) {
	return &linear.IssueConnection{
		Nodes:    m.issues,
		PageInfo: linear.PageInfo{HasNextPage: false},
	}, nil
}

func (m *linearMockClient) QueryWorkflowStates(_ context.Context, _ string) ([]linear.WorkflowState, error) {
	return []linear.WorkflowState{
		{ID: "s1", Name: "Todo", Type: "unstarted"},
		{ID: "s2", Name: "In Progress", Type: "started"},
	}, nil
}

func (m *linearMockClient) MutateIssueState(_ context.Context, _, _ string) (*linear.MutationResult, error) {
	return &linear.MutationResult{Success: true}, nil
}

func (m *linearMockClient) MutateIssueUpdate(_ context.Context, _, _, _ string) (*linear.MutationResult, error) {
	return &linear.MutationResult{Success: true}, nil
}

// TestLinearProviderContract runs the full adapters.RunContractTests suite
// against LinearProvider using a mocked GraphQLClient (AC1, AC2 — Story 30.4).
func TestLinearProviderContract(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()

		client := &linearMockClient{
			issues: []linear.IssueNode{
				{
					ID:         "issue-1",
					Identifier: "CONTRACT-1",
					Title:      "Contract test task A",
					Priority:   2,
					State:      linear.IssueState{ID: "s1", Name: "In Progress", Type: "started"},
					Team:       linear.IssueTeam{ID: "team-1", Key: "CT"},
					Labels:     linear.IssueLabels{},
					CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:         "issue-2",
					Identifier: "CONTRACT-2",
					Title:      "Contract test task B",
					Priority:   4,
					State:      linear.IssueState{ID: "s2", Name: "Todo", Type: "unstarted"},
					Team:       linear.IssueTeam{ID: "team-1", Key: "CT"},
					Labels:     linear.IssueLabels{},
					CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
		}

		config := &linear.LinearConfig{
			APIKey:       "test-contract-key",
			TeamIDs:      []string{"team-1"},
			PollInterval: 5 * time.Minute,
		}

		provider := linear.NewLinearProvider(client, config)
		t.Cleanup(provider.Stop)
		return provider
	}

	adapters.RunContractTests(t, factory)
}

// TestLinearProviderContract_EmptyTeam runs contract tests with a provider
// that has no issues, exercising the empty-result path.
func TestLinearProviderContract_EmptyTeam(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()

		client := &linearMockClient{issues: []linear.IssueNode{}}
		config := &linear.LinearConfig{
			APIKey:       "test-contract-key",
			TeamIDs:      []string{"team-1"},
			PollInterval: 5 * time.Minute,
		}

		provider := linear.NewLinearProvider(client, config)
		t.Cleanup(provider.Stop)
		return provider
	}

	adapters.RunContractTests(t, factory)
}
