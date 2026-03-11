package linear

import "time"

// GraphQL query and response types for the Linear API.
// These provide compile-time safety for query construction and response parsing (AC6).

// Viewer represents the authenticated Linear user returned by the viewer query.
type Viewer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ViewerResponse wraps the viewer query result.
type ViewerResponse struct {
	Viewer Viewer `json:"viewer"`
}

// PageInfo contains cursor-based pagination metadata.
type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

// IssueNode represents a single Linear issue in the paginated response.
type IssueNode struct {
	ID          string      `json:"id"`
	Identifier  string      `json:"identifier"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Priority    int         `json:"priority"`
	Estimate    *float64    `json:"estimate"`
	DueDate     *string     `json:"dueDate"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	State       IssueState  `json:"state"`
	Team        IssueTeam   `json:"team"`
	Labels      IssueLabels `json:"labels"`
	Assignee    *IssueUser  `json:"assignee"`
}

// IssueState holds the workflow state of an issue.
type IssueState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// IssueTeam identifies which team owns the issue.
type IssueTeam struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

// IssueLabels wraps the paginated labels connection.
type IssueLabels struct {
	Nodes []IssueLabel `json:"nodes"`
}

// IssueLabel represents a single label on an issue.
type IssueLabel struct {
	Name string `json:"name"`
}

// IssueUser represents a user assigned to an issue.
type IssueUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	IsMe  bool   `json:"isMe"`
}

// IssueConnection wraps the paginated issues response.
type IssueConnection struct {
	Nodes    []IssueNode `json:"nodes"`
	PageInfo PageInfo    `json:"pageInfo"`
}

// IssuesResponse wraps the team issues query result.
type IssuesResponse struct {
	Team struct {
		Issues IssueConnection `json:"issues"`
	} `json:"team"`
}

// WorkflowState represents a team workflow state.
type WorkflowState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// WorkflowStatesConnection wraps the paginated workflow states response.
type WorkflowStatesConnection struct {
	Nodes []WorkflowState `json:"nodes"`
}

// WorkflowStatesResponse wraps the team workflowStates query result.
type WorkflowStatesResponse struct {
	Team struct {
		States WorkflowStatesConnection `json:"states"`
	} `json:"team"`
}
