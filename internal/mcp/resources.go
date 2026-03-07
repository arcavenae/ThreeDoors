package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// ResourceContent represents a single content item in a resource read response.
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
}

// ResourceReadResult is the response to resources/read.
type ResourceReadResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceReadParams is the client request for resources/read.
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// ResponseMetadata provides query context for LLM clients.
type ResponseMetadata struct {
	TotalCount       int      `json:"total_count"`
	ReturnedCount    int      `json:"returned_count"`
	QueryTimeMs      float64  `json:"query_time_ms"`
	ProvidersQueried []string `json:"providers_queried"`
	DataFreshness    string   `json:"data_freshness"`
}

// resourceDefinitions returns the static list of MCP resources this server exposes.
func resourceDefinitions() []ResourceItem {
	return []ResourceItem{
		{
			URI:         "threedoors://tasks",
			Name:        "All Tasks",
			Description: "List of all tasks across all providers",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://tasks/{id}",
			Name:        "Task by ID",
			Description: "Single task detail by ID",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://tasks/status/{status}",
			Name:        "Tasks by Status",
			Description: "Tasks filtered by status (todo, in-progress, complete, etc.)",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://tasks/provider/{name}",
			Name:        "Tasks by Provider",
			Description: "Tasks filtered by source provider name",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://providers",
			Name:        "Provider Registry",
			Description: "All configured providers with health status",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://providers/{name}/status",
			Name:        "Provider Status",
			Description: "Health status of a single provider",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://session/current",
			Name:        "Current Session",
			Description: "Current session metrics",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://session/history",
			Name:        "Session History",
			Description: "Historical session metrics",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://analytics/mood-correlation",
			Name:        "Mood Correlation",
			Description: "Mood rating vs productivity metrics correlation",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://analytics/time-of-day",
			Name:        "Time of Day Productivity",
			Description: "Hourly completion rates with peak and slump hours",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://analytics/streaks",
			Name:        "Streak Analysis",
			Description: "Current and historical completion streak data",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://analytics/burnout-risk",
			Name:        "Burnout Risk",
			Description: "Composite burnout risk score from multiple behavioral signals",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://analytics/task-preferences",
			Name:        "Task Preferences",
			Description: "Task type preferences by mood and time",
			MimeType:    "application/json",
		},
		{
			URI:         "threedoors://analytics/weekly-summary",
			Name:        "Weekly Summary",
			Description: "Weekly productivity summary with patterns and recommendations",
			MimeType:    "application/json",
		},
	}
}

// handleResourceRead dispatches a resources/read request to the appropriate handler.
func (s *MCPServer) handleResourceRead(req *Request) *Response {
	var params ResourceReadParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid params: %v", err))
		}
	}

	uri := params.URI
	if uri == "" {
		return NewErrorResponse(req.ID, CodeInvalidParams, "uri is required")
	}

	start := time.Now().UTC()

	switch {
	case uri == "threedoors://tasks":
		return s.readAllTasks(req, start)
	case uri == "threedoors://providers":
		return s.readProviders(req, start)
	case uri == "threedoors://session/current":
		return s.readCurrentSession(req, start)
	case uri == "threedoors://session/history":
		return s.readSessionHistory(req, start)
	case strings.HasPrefix(uri, "threedoors://tasks/status/"):
		status := strings.TrimPrefix(uri, "threedoors://tasks/status/")
		return s.readTasksByStatus(req, status, start)
	case strings.HasPrefix(uri, "threedoors://tasks/provider/"):
		provider := strings.TrimPrefix(uri, "threedoors://tasks/provider/")
		return s.readTasksByProvider(req, provider, start)
	case strings.HasPrefix(uri, "threedoors://providers/") && strings.HasSuffix(uri, "/status"):
		name := strings.TrimPrefix(uri, "threedoors://providers/")
		name = strings.TrimSuffix(name, "/status")
		return s.readProviderStatus(req, name, start)
	case uri == "threedoors://analytics/mood-correlation":
		return s.readAnalyticsMoodCorrelation(req, start)
	case uri == "threedoors://analytics/time-of-day":
		return s.readAnalyticsTimeOfDay(req, start)
	case uri == "threedoors://analytics/streaks":
		return s.readAnalyticsStreaks(req, start)
	case uri == "threedoors://analytics/burnout-risk":
		return s.readAnalyticsBurnoutRisk(req, start)
	case uri == "threedoors://analytics/task-preferences":
		return s.readAnalyticsTaskPreferences(req, start)
	case uri == "threedoors://analytics/weekly-summary":
		return s.readAnalyticsWeeklySummary(req, start)
	case strings.HasPrefix(uri, "threedoors://tasks/"):
		id := strings.TrimPrefix(uri, "threedoors://tasks/")
		return s.readTaskByID(req, id, start)
	default:
		return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("unknown resource URI: %s", uri))
	}
}

func (s *MCPServer) readAllTasks(req *Request, start time.Time) *Response {
	tasks := s.pool.GetAllTasks()

	type tasksResponse struct {
		Tasks    []*core.Task     `json:"tasks"`
		Metadata ResponseMetadata `json:"_metadata"`
	}

	providers := s.providerNames()
	resp := tasksResponse{
		Tasks: tasks,
		Metadata: ResponseMetadata{
			TotalCount:       len(tasks),
			ReturnedCount:    len(tasks),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: providers,
			DataFreshness:    "live",
		},
	}
	if resp.Tasks == nil {
		resp.Tasks = []*core.Task{}
	}

	return s.resourceJSON(req, "threedoors://tasks", resp)
}

func (s *MCPServer) readTaskByID(req *Request, id string, start time.Time) *Response {
	task := s.pool.GetTask(id)
	if task == nil {
		return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("task not found: %s", id))
	}

	// Attempt to attach enrichment data.
	var enrichment any
	if s.enrichDB != nil {
		if meta, err := s.enrichDB.GetTaskMetadata(task.ID); err == nil {
			enrichment = meta
		}
	}

	type taskWithEnrichment struct {
		*core.Task
		Enrichment any `json:"enrichment,omitempty"`
	}

	resp := struct {
		Task     taskWithEnrichment `json:"task"`
		Metadata ResponseMetadata   `json:"_metadata"`
	}{
		Task: taskWithEnrichment{Task: task, Enrichment: enrichment},
		Metadata: ResponseMetadata{
			TotalCount:       1,
			ReturnedCount:    1,
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}

	return s.resourceJSON(req, "threedoors://tasks/"+id, resp)
}

func (s *MCPServer) readTasksByStatus(req *Request, status string, start time.Time) *Response {
	if err := core.ValidateStatus(status); err != nil {
		return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid status: %s", status))
	}

	tasks := s.pool.GetTasksByStatus(core.TaskStatus(status))

	type tasksResponse struct {
		Tasks    []*core.Task     `json:"tasks"`
		Metadata ResponseMetadata `json:"_metadata"`
	}

	allTasks := s.pool.GetAllTasks()
	resp := tasksResponse{
		Tasks: tasks,
		Metadata: ResponseMetadata{
			TotalCount:       len(allTasks),
			ReturnedCount:    len(tasks),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	if resp.Tasks == nil {
		resp.Tasks = []*core.Task{}
	}

	return s.resourceJSON(req, "threedoors://tasks/status/"+status, resp)
}

func (s *MCPServer) readTasksByProvider(req *Request, provider string, start time.Time) *Response {
	allTasks := s.pool.GetAllTasks()
	var tasks []*core.Task
	for _, t := range allTasks {
		if t.EffectiveSourceProvider() == provider {
			tasks = append(tasks, t)
		}
	}

	type tasksResponse struct {
		Tasks    []*core.Task     `json:"tasks"`
		Metadata ResponseMetadata `json:"_metadata"`
	}

	resp := tasksResponse{
		Tasks: tasks,
		Metadata: ResponseMetadata{
			TotalCount:       len(allTasks),
			ReturnedCount:    len(tasks),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: []string{provider},
			DataFreshness:    "live",
		},
	}
	if resp.Tasks == nil {
		resp.Tasks = []*core.Task{}
	}

	return s.resourceJSON(req, "threedoors://tasks/provider/"+provider, resp)
}

func (s *MCPServer) readProviders(req *Request, start time.Time) *Response {
	names := s.registry.ListProviders()

	type providerInfo struct {
		Name   string                  `json:"name"`
		Active bool                    `json:"active"`
		Health *core.HealthCheckResult `json:"health,omitempty"`
	}

	var providers []providerInfo
	for _, name := range names {
		info := providerInfo{Name: name}
		if p, err := s.registry.GetProvider(name); err == nil {
			info.Active = true
			h := p.HealthCheck()
			info.Health = &h
		}
		providers = append(providers, info)
	}
	if providers == nil {
		providers = []providerInfo{}
	}

	type providersResponse struct {
		Providers []providerInfo   `json:"providers"`
		Metadata  ResponseMetadata `json:"_metadata"`
	}

	resp := providersResponse{
		Providers: providers,
		Metadata: ResponseMetadata{
			TotalCount:       len(providers),
			ReturnedCount:    len(providers),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: names,
			DataFreshness:    "live",
		},
	}

	return s.resourceJSON(req, "threedoors://providers", resp)
}

func (s *MCPServer) readProviderStatus(req *Request, name string, start time.Time) *Response {
	p, err := s.registry.GetProvider(name)
	if err != nil {
		return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("provider not found: %s", name))
	}

	health := p.HealthCheck()

	type statusResponse struct {
		Name     string                 `json:"name"`
		Health   core.HealthCheckResult `json:"health"`
		Metadata ResponseMetadata       `json:"_metadata"`
	}

	resp := statusResponse{
		Name:   name,
		Health: health,
		Metadata: ResponseMetadata{
			TotalCount:       1,
			ReturnedCount:    1,
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: []string{name},
			DataFreshness:    "live",
		},
	}

	return s.resourceJSON(req, fmt.Sprintf("threedoors://providers/%s/status", name), resp)
}

func (s *MCPServer) readCurrentSession(req *Request, start time.Time) *Response {
	snapshot := s.session.GetMetricsSnapshot()

	type sessionResponse struct {
		SessionID      string           `json:"session_id"`
		TasksCompleted int              `json:"tasks_completed"`
		DurationSecs   float64          `json:"duration_seconds"`
		Metadata       ResponseMetadata `json:"_metadata"`
	}

	resp := sessionResponse{
		SessionID:      s.session.GetSessionID(),
		TasksCompleted: snapshot.TasksCompleted,
		DurationSecs:   snapshot.DurationSeconds(),
		Metadata: ResponseMetadata{
			TotalCount:       1,
			ReturnedCount:    1,
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}

	return s.resourceJSON(req, "threedoors://session/current", resp)
}

func (s *MCPServer) readSessionHistory(req *Request, start time.Time) *Response {
	type historyResponse struct {
		Sessions []core.SessionMetrics `json:"sessions"`
		Metadata ResponseMetadata      `json:"_metadata"`
	}

	var sessions []core.SessionMetrics
	if s.sessionsReader != nil {
		var err error
		sessions, err = s.sessionsReader.ReadAll()
		if err != nil {
			sessions = []core.SessionMetrics{}
		}
	}
	if sessions == nil {
		sessions = []core.SessionMetrics{}
	}

	resp := historyResponse{
		Sessions: sessions,
		Metadata: ResponseMetadata{
			TotalCount:       len(sessions),
			ReturnedCount:    len(sessions),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "snapshot",
		},
	}

	return s.resourceJSON(req, "threedoors://session/history", resp)
}

// resourceJSON marshals data as JSON text and wraps in a ResourceReadResult.
func (s *MCPServer) resourceJSON(req *Request, uri string, data any) *Response {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return NewErrorResponse(req.ID, CodeInternalError, fmt.Sprintf("marshal resource: %v", err))
	}

	result := ResourceReadResult{
		Contents: []ResourceContent{{
			URI:      uri,
			MimeType: "application/json",
			Text:     string(jsonBytes),
		}},
	}
	return NewResponse(req.ID, result)
}

// providerNames returns the names of all registered providers.
func (s *MCPServer) providerNames() []string {
	names := s.registry.ListProviders()
	if names == nil {
		return []string{}
	}
	return names
}

func (s *MCPServer) readAnalyticsMoodCorrelation(req *Request, start time.Time) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.resourceJSON(req, "threedoors://analytics/mood-correlation", map[string]any{"error": "analytics not available"})
	}
	thirtyDaysAgo := time.Now().UTC().AddDate(0, 0, -30)
	result, err := pm.MoodCorrelationAnalysis(thirtyDaysAgo, time.Now().UTC())
	if err != nil {
		return NewErrorResponse(req.ID, CodeInternalError, fmt.Sprintf("mood correlation: %v", err))
	}
	return s.resourceJSON(req, "threedoors://analytics/mood-correlation", result)
}

func (s *MCPServer) readAnalyticsTimeOfDay(req *Request, start time.Time) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.resourceJSON(req, "threedoors://analytics/time-of-day", map[string]any{"error": "analytics not available"})
	}
	thirtyDaysAgo := time.Now().UTC().AddDate(0, 0, -30)
	result, err := pm.ProductivityProfileAnalysis(thirtyDaysAgo, time.Now().UTC())
	if err != nil {
		return NewErrorResponse(req.ID, CodeInternalError, fmt.Sprintf("productivity profile: %v", err))
	}
	return s.resourceJSON(req, "threedoors://analytics/time-of-day", result)
}

func (s *MCPServer) readAnalyticsStreaks(req *Request, start time.Time) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.resourceJSON(req, "threedoors://analytics/streaks", map[string]any{"error": "analytics not available"})
	}
	result, err := pm.StreakAnalysis()
	if err != nil {
		return NewErrorResponse(req.ID, CodeInternalError, fmt.Sprintf("streak analysis: %v", err))
	}
	return s.resourceJSON(req, "threedoors://analytics/streaks", result)
}

func (s *MCPServer) readAnalyticsBurnoutRisk(req *Request, start time.Time) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.resourceJSON(req, "threedoors://analytics/burnout-risk", map[string]any{"error": "analytics not available"})
	}
	result, err := pm.BurnoutRisk()
	if err != nil {
		return NewErrorResponse(req.ID, CodeInternalError, fmt.Sprintf("burnout risk: %v", err))
	}
	return s.resourceJSON(req, "threedoors://analytics/burnout-risk", result)
}

func (s *MCPServer) readAnalyticsTaskPreferences(req *Request, start time.Time) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.resourceJSON(req, "threedoors://analytics/task-preferences", map[string]any{"error": "analytics not available"})
	}
	// Task preferences reuses mood correlation as it contains task type breakdown per mood.
	thirtyDaysAgo := time.Now().UTC().AddDate(0, 0, -30)
	result, err := pm.MoodCorrelationAnalysis(thirtyDaysAgo, time.Now().UTC())
	if err != nil {
		return NewErrorResponse(req.ID, CodeInternalError, fmt.Sprintf("task preferences: %v", err))
	}
	return s.resourceJSON(req, "threedoors://analytics/task-preferences", result)
}

func (s *MCPServer) readAnalyticsWeeklySummary(req *Request, start time.Time) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.resourceJSON(req, "threedoors://analytics/weekly-summary", map[string]any{"error": "analytics not available"})
	}
	result, err := pm.WeeklySummaryAnalysis(time.Now().UTC())
	if err != nil {
		return NewErrorResponse(req.ID, CodeInternalError, fmt.Sprintf("weekly summary: %v", err))
	}
	return s.resourceJSON(req, "threedoors://analytics/weekly-summary", result)
}

// patternMiner creates a PatternMiner from the server's session reader and pool.
func (s *MCPServer) patternMiner() *PatternMiner {
	if s.sessionsReader == nil {
		return nil
	}
	return NewPatternMiner(s.sessionsReader, s.pool)
}

// millisSince returns milliseconds elapsed since t.
func millisSince(t time.Time) float64 {
	return float64(time.Since(t).Microseconds()) / 1000.0
}
