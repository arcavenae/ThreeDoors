package clickup

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

const providerName = "clickup"

// TaskFetcher abstracts the ClickUp API operations needed by the provider.
type TaskFetcher interface {
	GetTasks(ctx context.Context, listID string, page int) ([]ClickUpTask, error)
	GetAuthorizedUser(ctx context.Context) (*ClickUpUser, error)
}

// ClickUpProvider implements core.TaskProvider for ClickUp.
// Tasks are fetched from the ClickUp REST API and mapped to ThreeDoors tasks.
// Write operations return core.ErrReadOnly.
type ClickUpProvider struct {
	client        TaskFetcher
	listIDs       []string
	statusMapping map[string]core.TaskStatus
	pollInterval  time.Duration
}

// NewClickUpProvider creates a ClickUpProvider with the given API client and config.
func NewClickUpProvider(client TaskFetcher, cfg *ClickUpConfig) *ClickUpProvider {
	statusMapping := DefaultStatusMapping
	if len(cfg.StatusMapping) > 0 {
		statusMapping = cfg.StatusMapping
	}

	return &ClickUpProvider{
		client:        client,
		listIDs:       cfg.ListIDs,
		statusMapping: statusMapping,
		pollInterval:  cfg.PollInterval,
	}
}

// Name returns the provider identifier.
func (p *ClickUpProvider) Name() string {
	return providerName
}

// LoadTasks fetches tasks from configured ClickUp lists and maps them to ThreeDoors tasks.
// Tasks without a name are skipped with a warning log.
func (p *ClickUpProvider) LoadTasks() ([]*core.Task, error) {
	ctx := context.Background()

	var allTasks []*core.Task

	for _, listID := range p.listIDs {
		apiTasks, err := p.client.GetTasks(ctx, listID, 0)
		if err != nil {
			return nil, fmt.Errorf("clickup load tasks list %s: %w", listID, err)
		}
		for i := range apiTasks {
			task, ok := p.mapClickUpTask(&apiTasks[i])
			if !ok {
				continue
			}
			allTasks = append(allTasks, task)
		}
	}

	if allTasks == nil {
		allTasks = []*core.Task{}
	}

	return allTasks, nil
}

// mapClickUpTask converts a ClickUp API task to a ThreeDoors Task.
// Returns false if the task should be skipped (e.g., missing required fields).
func (p *ClickUpProvider) mapClickUpTask(ct *ClickUpTask) (*core.Task, bool) {
	if ct.Name == "" {
		log.Printf("clickup: skipping task %s: missing name", ct.ID)
		return nil, false
	}

	now := time.Now().UTC()

	task := &core.Task{
		ID:             ct.ID,
		Text:           ct.Name,
		Status:         p.mapStatus(ct.Status.Status),
		Effort:         MapPriority(ct.Priority),
		SourceProvider: providerName,
		SourceRefs: []core.SourceRef{
			{
				Provider: providerName,
				NativeID: ct.ID,
			},
		},
		Notes:     []core.TaskNote{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// AC4: description → Context
	if ct.Description != "" {
		task.Context = ct.Description
	}

	// AC8: tags → Context (appended)
	if len(ct.Tags) > 0 {
		tagNames := make([]string, len(ct.Tags))
		for i, tag := range ct.Tags {
			tagNames[i] = tag.Name
		}
		tagStr := strings.Join(tagNames, ", ")
		if task.Context != "" {
			task.Context = task.Context + " | " + tagStr
		} else {
			task.Context = tagStr
		}
	}

	// AC7: due_date (Unix ms) → DueDate (UTC)
	if ct.DueDate != "" {
		if dueDate, err := parseUnixMillis(ct.DueDate); err == nil {
			task.DeferUntil = &dueDate
		}
	}

	// AC6: SourceRef with list ID and URL
	if ct.List.ID != "" {
		task.SourceRefs[0].NativeID = fmt.Sprintf("%s:%s", ct.ID, ct.List.ID)
	}
	if ct.URL != "" {
		task.SourceProvider = fmt.Sprintf("clickup:%s", ct.URL)
	}

	return task, true
}

// mapStatus converts a ClickUp status string to a ThreeDoors TaskStatus.
// Falls back to StatusTodo for unknown statuses.
func (p *ClickUpProvider) mapStatus(clickUpStatus string) core.TaskStatus {
	normalized := strings.ToLower(strings.TrimSpace(clickUpStatus))
	if status, ok := p.statusMapping[normalized]; ok {
		return status
	}
	return core.StatusTodo
}

// MapPriority converts a ClickUp priority to a ThreeDoors TaskEffort.
// ClickUp priorities: 1=Urgent, 2=High, 3=Normal, 4=Low.
// nil priority means no priority set.
func MapPriority(priority *ClickUpPriority) core.TaskEffort {
	if priority == nil {
		return core.EffortQuickWin
	}
	switch priority.ID {
	case "1": // Urgent
		return core.EffortDeepWork
	case "2": // High
		return core.EffortDeepWork
	case "3": // Normal
		return core.EffortMedium
	case "4": // Low
		return core.EffortQuickWin
	default:
		return core.EffortQuickWin
	}
}

// parseUnixMillis parses a Unix millisecond timestamp string to time.Time in UTC.
func parseUnixMillis(ms string) (time.Time, error) {
	millis, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse unix millis %q: %w", ms, err)
	}
	return time.Unix(millis/1000, (millis%1000)*int64(time.Millisecond)).UTC(), nil
}

// SaveTask returns ErrReadOnly; ClickUp provider is read-only.
func (p *ClickUpProvider) SaveTask(_ *core.Task) error {
	return core.ErrReadOnly
}

// SaveTasks returns ErrReadOnly; ClickUp provider is read-only.
func (p *ClickUpProvider) SaveTasks(_ []*core.Task) error {
	return core.ErrReadOnly
}

// DeleteTask returns ErrReadOnly; ClickUp provider is read-only.
func (p *ClickUpProvider) DeleteTask(_ string) error {
	return core.ErrReadOnly
}

// MarkComplete returns ErrReadOnly; ClickUp provider is read-only.
func (p *ClickUpProvider) MarkComplete(_ string) error {
	return core.ErrReadOnly
}

// Watch returns nil; ClickUp provider uses poll-based sync.
func (p *ClickUpProvider) Watch() <-chan core.ChangeEvent {
	return nil
}

// HealthCheck verifies API connectivity via GetAuthorizedUser.
func (p *ClickUpProvider) HealthCheck() core.HealthCheckResult {
	start := time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := p.client.GetAuthorizedUser(ctx)
	duration := time.Since(start)

	if err != nil {
		return core.HealthCheckResult{
			Items: []core.HealthCheckItem{
				{
					Name:       "clickup_connectivity",
					Status:     core.HealthFail,
					Message:    fmt.Sprintf("ClickUp API unreachable: %v", err),
					Suggestion: "Check API token and network connectivity",
				},
			},
			Overall:  core.HealthFail,
			Duration: duration,
		}
	}

	return core.HealthCheckResult{
		Items: []core.HealthCheckItem{
			{
				Name:    "clickup_connectivity",
				Status:  core.HealthOK,
				Message: "ClickUp API reachable",
			},
		},
		Overall:  core.HealthOK,
		Duration: duration,
	}
}

// PollInterval returns the configured poll interval for sync.
func (p *ClickUpProvider) PollInterval() time.Duration {
	return p.pollInterval
}
