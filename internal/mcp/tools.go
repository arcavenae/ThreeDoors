package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// ToolCallParams is the client request for tools/call.
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ToolCallResult is the response to tools/call.
type ToolCallResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent is a single content item in a tool result.
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// toolDefinitions returns the static list of MCP tools this server exposes.
func toolDefinitions() []ToolItem {
	return []ToolItem{
		{
			Name:        "query_tasks",
			Description: "Query tasks with filters. Returns matching tasks with metadata.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"status":         map[string]any{"type": "string", "description": "Filter by status (todo, in-progress, complete, etc.)"},
					"type":           map[string]any{"type": "string", "description": "Filter by task type (creative, administrative, technical, physical)"},
					"effort":         map[string]any{"type": "string", "description": "Filter by effort level (quick-win, medium, deep-work)"},
					"provider":       map[string]any{"type": "string", "description": "Filter by source provider name"},
					"text_contains":  map[string]any{"type": "string", "description": "Filter tasks containing this text (case-insensitive)"},
					"created_after":  map[string]any{"type": "string", "description": "ISO 8601 datetime — only tasks created after this"},
					"created_before": map[string]any{"type": "string", "description": "ISO 8601 datetime — only tasks created before this"},
					"limit":          map[string]any{"type": "integer", "description": "Max results to return (default 50)"},
					"sort_by":        map[string]any{"type": "string", "description": "Sort field: created_at, updated_at, text (default created_at)"},
					"sort_order":     map[string]any{"type": "string", "description": "Sort direction: asc or desc (default asc)"},
				},
			},
		},
		{
			Name:        "get_task",
			Description: "Get full task detail including enrichment data for a given task ID.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{"type": "string", "description": "The task ID to retrieve"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "list_providers",
			Description: "List configured task providers with health status and sync freshness.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "get_session",
			Description: "Get current or historical session metrics.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"type": map[string]any{"type": "string", "description": "Session type: 'current' or 'history' (default current)"},
				},
			},
		},
		{
			Name:        "search_tasks",
			Description: "Full-text search across tasks with relevance scoring. Uses field-weighted Jaccard similarity (text 3x, context 2x, notes 1x) with recency boost.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query":       map[string]any{"type": "string", "description": "Search query text"},
					"max_results": map[string]any{"type": "integer", "description": "Max results to return (default 50)"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "get_mood_correlation",
			Description: "Get mood vs productivity correlation data for a time range.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"from": map[string]any{"type": "string", "description": "ISO 8601 start datetime (default 30 days ago)"},
					"to":   map[string]any{"type": "string", "description": "ISO 8601 end datetime (default now)"},
				},
			},
		},
		{
			Name:        "get_productivity_profile",
			Description: "Get time-of-day productivity analysis with peak and slump hours.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"from": map[string]any{"type": "string", "description": "ISO 8601 start datetime (default 30 days ago)"},
					"to":   map[string]any{"type": "string", "description": "ISO 8601 end datetime (default now)"},
				},
			},
		},
		{
			Name:        "burnout_risk",
			Description: "Assess burnout risk from behavioral signals. Returns composite score 0-1 with level (low/moderate/warning).",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "walk_graph",
			Description: "Traverse the task relationship graph from a root task using BFS. Returns a subgraph of nodes and edges within the specified depth.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id":    map[string]any{"type": "string", "description": "Root task ID to start traversal from"},
					"depth":      map[string]any{"type": "integer", "description": "Max traversal depth (default 2)"},
					"direction":  map[string]any{"type": "string", "description": "Traversal direction: outgoing, incoming, or both (default both)"},
					"edge_types": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Filter by edge types: blocks, related-to, subtask-of, duplicate-of, sequential, cross-ref"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "find_paths",
			Description: "Find all simple paths between two tasks up to a maximum depth.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"from_id":   map[string]any{"type": "string", "description": "Starting task ID"},
					"to_id":     map[string]any{"type": "string", "description": "Destination task ID"},
					"max_depth": map[string]any{"type": "integer", "description": "Max path length (default 5)"},
				},
				"required": []string{"from_id", "to_id"},
			},
		},
		{
			Name:        "get_critical_path",
			Description: "Get the longest dependency chain from a root task, following only 'blocks' edges.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"root_id": map[string]any{"type": "string", "description": "Root task ID"},
				},
				"required": []string{"root_id"},
			},
		},
		{
			Name:        "get_orphans",
			Description: "Get tasks with no relationships to any other tasks.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "get_completions",
			Description: "Get completion data with optional grouping and enrichment.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"from":             map[string]any{"type": "string", "description": "ISO 8601 start datetime (default 30 days ago)"},
					"to":               map[string]any{"type": "string", "description": "ISO 8601 end datetime (default now)"},
					"group_by":         map[string]any{"type": "string", "description": "Group results by: day, hour, week (default day)"},
					"include_mood":     map[string]any{"type": "boolean", "description": "Include mood data in results"},
					"include_patterns": map[string]any{"type": "boolean", "description": "Include pattern analysis"},
				},
			},
		},
		{
			Name:        "get_clusters",
			Description: "Get groups of related tasks using connected-component detection.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "get_provider_overlap",
			Description: "Find shared or similar tasks between two providers.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"provider_a": map[string]any{"type": "string", "description": "First provider name"},
					"provider_b": map[string]any{"type": "string", "description": "Second provider name"},
				},
				"required": []string{"provider_a", "provider_b"},
			},
		},
		{
			Name:        "get_unified_view",
			Description: "Get tasks matching a topic across all providers with their relationships.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"topic": map[string]any{"type": "string", "description": "Topic or keyword to search for"},
				},
				"required": []string{"topic"},
			},
		},
		{
			Name:        "suggest_cross_links",
			Description: "Suggest cross-provider relationships based on text similarity, shared references, and temporal proximity.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
}

// handleToolCall dispatches a tools/call request to the appropriate handler.
func (s *MCPServer) handleToolCall(req *Request) *Response {
	var params ToolCallParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid params: %v", err))
		}
	}

	switch params.Name {
	case "query_tasks":
		return s.toolQueryTasks(req, params.Arguments)
	case "get_task":
		return s.toolGetTask(req, params.Arguments)
	case "list_providers":
		return s.toolListProviders(req)
	case "get_session":
		return s.toolGetSession(req, params.Arguments)
	case "search_tasks":
		return s.toolSearchTasks(req, params.Arguments)
	case "get_mood_correlation":
		return s.toolGetMoodCorrelation(req, params.Arguments)
	case "get_productivity_profile":
		return s.toolGetProductivityProfile(req, params.Arguments)
	case "burnout_risk":
		return s.toolBurnoutRisk(req)
	case "get_completions":
		return s.toolGetCompletions(req, params.Arguments)
	case "propose_enrichment":
		if s.proposalStore == nil {
			return s.toolError(req, "proposal store not configured")
		}
		return s.toolProposeEnrichment(req, params.Arguments)
	case "suggest_task":
		if s.proposalStore == nil {
			return s.toolError(req, "proposal store not configured")
		}
		return s.toolSuggestTask(req, params.Arguments)
	case "suggest_relationship":
		if s.proposalStore == nil {
			return s.toolError(req, "proposal store not configured")
		}
		return s.toolSuggestRelationship(req, params.Arguments)
	case "walk_graph":
		return s.toolWalkGraph(req, params.Arguments)
	case "find_paths":
		return s.toolFindPaths(req, params.Arguments)
	case "get_critical_path":
		return s.toolGetCriticalPath(req, params.Arguments)
	case "get_orphans":
		return s.toolGetOrphans(req)
	case "get_clusters":
		return s.toolGetClusters(req)
	case "get_provider_overlap":
		return s.toolGetProviderOverlap(req, params.Arguments)
	case "get_unified_view":
		return s.toolGetUnifiedView(req, params.Arguments)
	case "suggest_cross_links":
		return s.toolSuggestCrossLinks(req)
	case "prioritize_tasks":
		return s.toolPrioritizeTasks(req, params.Arguments)
	case "analyze_workload":
		return s.toolAnalyzeWorkload(req)
	case "focus_recommendation":
		return s.toolFocusRecommendation(req, params.Arguments)
	case "what_if":
		return s.toolWhatIf(req, params.Arguments)
	case "context_switch_analysis":
		return s.toolContextSwitchAnalysis(req, params.Arguments)
	default:
		return NewErrorResponse(req.ID, CodeMethodNotFound, fmt.Sprintf("unknown tool: %s", params.Name))
	}
}

func (s *MCPServer) toolQueryTasks(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var opts FilterOptions
	if args != nil {
		if err := json.Unmarshal(args, &opts); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}

	allTasks := s.pool.GetAllTasks()
	filtered := FilterTasks(allTasks, opts)

	type queryResult struct {
		Tasks    []*taskSummary   `json:"tasks"`
		Metadata ResponseMetadata `json:"_metadata"`
	}

	summaries := make([]*taskSummary, len(filtered))
	for i, t := range filtered {
		summaries[i] = newTaskSummary(t)
	}

	result := queryResult{
		Tasks: summaries,
		Metadata: ResponseMetadata{
			TotalCount:       len(allTasks),
			ReturnedCount:    len(filtered),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}

	return s.toolJSON(req, result)
}

func (s *MCPServer) toolGetTask(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		TaskID string `json:"task_id"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}
	if params.TaskID == "" {
		return NewErrorResponse(req.ID, CodeInvalidParams, "task_id is required")
	}

	task := s.pool.GetTask(params.TaskID)
	if task == nil {
		return s.toolError(req, fmt.Sprintf("task not found: %s", params.TaskID))
	}

	// Attach enrichment data if available.
	var enrichment any
	if s.enrichDB != nil {
		if meta, err := s.enrichDB.GetTaskMetadata(task.ID); err == nil {
			enrichment = meta
		}
	}

	type taskResult struct {
		Task       *taskDetail      `json:"task"`
		Enrichment any              `json:"enrichment,omitempty"`
		Metadata   ResponseMetadata `json:"_metadata"`
	}

	result := taskResult{
		Task:       newTaskDetail(task),
		Enrichment: enrichment,
		Metadata: ResponseMetadata{
			TotalCount:       1,
			ReturnedCount:    1,
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}

	return s.toolJSON(req, result)
}

func (s *MCPServer) toolListProviders(req *Request) *Response {
	start := time.Now().UTC()

	names := s.registry.ListProviders()

	type providerInfo struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
		Health string `json:"health"`
	}

	var providers []providerInfo
	for _, name := range names {
		info := providerInfo{Name: name}
		if p, err := s.registry.GetProvider(name); err == nil {
			info.Active = true
			h := p.HealthCheck()
			info.Health = string(h.Overall)
		} else {
			info.Health = "UNKNOWN"
		}
		providers = append(providers, info)
	}
	if providers == nil {
		providers = []providerInfo{}
	}

	type listResult struct {
		Providers []providerInfo   `json:"providers"`
		Metadata  ResponseMetadata `json:"_metadata"`
	}

	result := listResult{
		Providers: providers,
		Metadata: ResponseMetadata{
			TotalCount:       len(providers),
			ReturnedCount:    len(providers),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: names,
			DataFreshness:    "live",
		},
	}

	return s.toolJSON(req, result)
}

func (s *MCPServer) toolGetSession(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		Type string `json:"type"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}

	if params.Type == "history" {
		return s.readSessionHistory(req, start)
	}

	// Default: current session.
	return s.readCurrentSession(req, start)
}

func (s *MCPServer) toolSearchTasks(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}
	if params.Query == "" {
		return NewErrorResponse(req.ID, CodeInvalidParams, "query is required")
	}

	opts := DefaultSearchOptions()
	if params.MaxResults > 0 {
		opts.MaxResults = params.MaxResults
	}

	engine := NewTaskQueryEngine(s.pool)
	results := engine.Search(params.Query, opts)

	type searchResponse struct {
		Results  []SearchResult   `json:"results"`
		Metadata ResponseMetadata `json:"_metadata"`
	}

	allTasks := s.pool.GetAllTasks()
	if results == nil {
		results = []SearchResult{}
	}

	resp := searchResponse{
		Results: results,
		Metadata: ResponseMetadata{
			TotalCount:       len(allTasks),
			ReturnedCount:    len(results),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}

	return s.toolJSON(req, resp)
}

func (s *MCPServer) toolWalkGraph(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		TaskID    string   `json:"task_id"`
		Depth     int      `json:"depth"`
		Direction string   `json:"direction"`
		EdgeTypes []string `json:"edge_types"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}
	if params.TaskID == "" {
		return NewErrorResponse(req.ID, CodeInvalidParams, "task_id is required")
	}

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()

	var edgeTypes []EdgeType
	for _, et := range params.EdgeTypes {
		edgeTypes = append(edgeTypes, EdgeType(et))
	}

	graph, err := WalkGraph(s.pool, edges, WalkGraphOptions{
		TaskID:    params.TaskID,
		Depth:     params.Depth,
		Direction: params.Direction,
		EdgeTypes: edgeTypes,
	})
	if err != nil {
		return s.toolError(req, err.Error())
	}

	type walkResult struct {
		Graph    *TaskGraph       `json:"graph"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	result := walkResult{
		Graph: graph,
		Metadata: ResponseMetadata{
			TotalCount:       len(graph.Nodes),
			ReturnedCount:    len(graph.Nodes),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolFindPaths(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		FromID   string `json:"from_id"`
		ToID     string `json:"to_id"`
		MaxDepth int    `json:"max_depth"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}
	if params.FromID == "" || params.ToID == "" {
		return NewErrorResponse(req.ID, CodeInvalidParams, "from_id and to_id are required")
	}

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()

	paths, err := FindPaths(s.pool, edges, params.FromID, params.ToID, params.MaxDepth)
	if err != nil {
		return s.toolError(req, err.Error())
	}
	if paths == nil {
		paths = [][]string{}
	}

	type pathsResult struct {
		Paths    [][]string       `json:"paths"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	result := pathsResult{
		Paths: paths,
		Metadata: ResponseMetadata{
			TotalCount:       len(paths),
			ReturnedCount:    len(paths),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolGetCriticalPath(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		RootID string `json:"root_id"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}
	if params.RootID == "" {
		return NewErrorResponse(req.ID, CodeInvalidParams, "root_id is required")
	}

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()

	path, err := GetCriticalPath(s.pool, edges, params.RootID)
	if err != nil {
		return s.toolError(req, err.Error())
	}
	if path == nil {
		path = []string{}
	}

	type criticalPathResult struct {
		Path     []string         `json:"path"`
		Length   int              `json:"length"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	result := criticalPathResult{
		Path:   path,
		Length: len(path),
		Metadata: ResponseMetadata{
			TotalCount:       1,
			ReturnedCount:    1,
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolGetOrphans(req *Request) *Response {
	start := time.Now().UTC()

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()
	orphans := GetOrphans(s.pool, edges)
	if orphans == nil {
		orphans = []*core.Task{}
	}

	type orphansResult struct {
		Tasks    []*core.Task     `json:"tasks"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	allTasks := s.pool.GetAllTasks()
	result := orphansResult{
		Tasks: orphans,
		Metadata: ResponseMetadata{
			TotalCount:       len(allTasks),
			ReturnedCount:    len(orphans),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolGetClusters(req *Request) *Response {
	start := time.Now().UTC()

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()
	clusters := GetClusters(s.pool, edges)
	if clusters == nil {
		clusters = []Cluster{}
	}

	type clustersResult struct {
		Clusters []Cluster        `json:"clusters"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	result := clustersResult{
		Clusters: clusters,
		Metadata: ResponseMetadata{
			TotalCount:       len(clusters),
			ReturnedCount:    len(clusters),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolGetProviderOverlap(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		ProviderA string `json:"provider_a"`
		ProviderB string `json:"provider_b"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}
	if params.ProviderA == "" || params.ProviderB == "" {
		return NewErrorResponse(req.ID, CodeInvalidParams, "provider_a and provider_b are required")
	}

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()
	linker := NewCrossProviderLinker(s.pool)
	overlap := linker.GetProviderOverlap(params.ProviderA, params.ProviderB, edges)

	type overlapResult struct {
		Overlap  ProviderOverlap  `json:"overlap"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	result := overlapResult{
		Overlap: overlap,
		Metadata: ResponseMetadata{
			TotalCount:       1,
			ReturnedCount:    1,
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolGetUnifiedView(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		Topic string `json:"topic"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}
	if params.Topic == "" {
		return NewErrorResponse(req.ID, CodeInvalidParams, "topic is required")
	}

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()
	linker := NewCrossProviderLinker(s.pool)
	view := linker.GetUnifiedView(params.Topic, edges)

	type viewResult struct {
		View     UnifiedView      `json:"view"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	result := viewResult{
		View: view,
		Metadata: ResponseMetadata{
			TotalCount:       len(view.Tasks),
			ReturnedCount:    len(view.Tasks),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolSuggestCrossLinks(req *Request) *Response {
	start := time.Now().UTC()

	linker := NewCrossProviderLinker(s.pool)
	suggestions := linker.SuggestCrossLinks()
	if suggestions == nil {
		suggestions = []CrossLinkSuggestion{}
	}

	type suggestResult struct {
		Suggestions []CrossLinkSuggestion `json:"suggestions"`
		Metadata    ResponseMetadata      `json:"_metadata"`
	}
	result := suggestResult{
		Suggestions: suggestions,
		Metadata: ResponseMetadata{
			TotalCount:       len(suggestions),
			ReturnedCount:    len(suggestions),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

// toolJSON marshals data as JSON and wraps in a ToolCallResult.
func (s *MCPServer) toolJSON(req *Request, data any) *Response {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return NewErrorResponse(req.ID, CodeInternalError, fmt.Sprintf("marshal tool result: %v", err))
	}

	result := ToolCallResult{
		Content: []ToolContent{{
			Type: "text",
			Text: string(jsonBytes),
		}},
	}
	return NewResponse(req.ID, result)
}

// toolError returns a tool-level error (not JSON-RPC error).
func (s *MCPServer) toolError(req *Request, msg string) *Response {
	result := ToolCallResult{
		Content: []ToolContent{{
			Type: "text",
			Text: msg,
		}},
		IsError: true,
	}
	return NewResponse(req.ID, result)
}

// taskSummary is a lightweight view for list results.
type taskSummary struct {
	ID             string     `json:"id"`
	Text           string     `json:"text"`
	Status         string     `json:"status"`
	Type           string     `json:"type,omitempty"`
	Effort         string     `json:"effort,omitempty"`
	SourceProvider string     `json:"source_provider,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

func newTaskSummary(t *core.Task) *taskSummary {
	return &taskSummary{
		ID:             t.ID,
		Text:           t.Text,
		Status:         string(t.Status),
		Type:           string(t.Type),
		Effort:         string(t.Effort),
		SourceProvider: t.EffectiveSourceProvider(),
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		CompletedAt:    t.CompletedAt,
	}
}

// taskDetail is a full view for single-task results.
type taskDetail struct {
	*taskSummary
	Context string          `json:"context,omitempty"`
	Notes   []core.TaskNote `json:"notes,omitempty"`
	Blocker string          `json:"blocker,omitempty"`
}

func (s *MCPServer) parseTimeRange(args json.RawMessage) (time.Time, time.Time) {
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -30)
	to := now

	if args != nil {
		var params struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		if err := json.Unmarshal(args, &params); err == nil {
			if params.From != "" {
				if t, err := time.Parse(time.RFC3339, params.From); err == nil {
					from = t
				}
			}
			if params.To != "" {
				if t, err := time.Parse(time.RFC3339, params.To); err == nil {
					to = t
				}
			}
		}
	}
	return from, to
}

func (s *MCPServer) toolGetMoodCorrelation(req *Request, args json.RawMessage) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.toolError(req, "analytics not available: no session reader configured")
	}
	from, to := s.parseTimeRange(args)
	result, err := pm.MoodCorrelationAnalysis(from, to)
	if err != nil {
		return s.toolError(req, fmt.Sprintf("mood correlation: %v", err))
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolGetProductivityProfile(req *Request, args json.RawMessage) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.toolError(req, "analytics not available: no session reader configured")
	}
	from, to := s.parseTimeRange(args)
	result, err := pm.ProductivityProfileAnalysis(from, to)
	if err != nil {
		return s.toolError(req, fmt.Sprintf("productivity profile: %v", err))
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolBurnoutRisk(req *Request) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.toolError(req, "analytics not available: no session reader configured")
	}
	result, err := pm.BurnoutRisk()
	if err != nil {
		return s.toolError(req, fmt.Sprintf("burnout risk: %v", err))
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolGetCompletions(req *Request, args json.RawMessage) *Response {
	pm := s.patternMiner()
	if pm == nil {
		return s.toolError(req, "analytics not available: no session reader configured")
	}
	from, to := s.parseTimeRange(args)

	var params struct {
		GroupBy         string `json:"group_by"`
		IncludeMood     bool   `json:"include_mood"`
		IncludePatterns bool   `json:"include_patterns"`
	}
	if args != nil {
		_ = json.Unmarshal(args, &params)
	}
	if params.GroupBy == "" {
		params.GroupBy = "day"
	}

	sessions, err := pm.sessionsInRange(from, to)
	if err != nil {
		return s.toolError(req, fmt.Sprintf("get completions: %v", err))
	}

	type completionGroup struct {
		Key         string   `json:"key"`
		Completions int      `json:"completions"`
		Sessions    int      `json:"sessions"`
		AvgMood     string   `json:"avg_mood,omitempty"`
		Patterns    []string `json:"patterns,omitempty"`
	}

	groups := make(map[string]*completionGroup)
	for _, s := range sessions {
		var key string
		switch params.GroupBy {
		case "hour":
			key = fmt.Sprintf("%02d:00", s.StartTime.Hour())
		case "week":
			year, week := s.StartTime.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		default:
			key = s.StartTime.Format("2006-01-02")
		}

		g, ok := groups[key]
		if !ok {
			g = &completionGroup{Key: key}
			groups[key] = g
		}
		g.Completions += s.TasksCompleted
		g.Sessions++
	}

	var result []completionGroup
	for _, g := range groups {
		result = append(result, *g)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	return s.toolJSON(req, map[string]any{
		"from":        from,
		"to":          to,
		"group_by":    params.GroupBy,
		"completions": result,
	})
}

func newTaskDetail(t *core.Task) *taskDetail {
	return &taskDetail{
		taskSummary: newTaskSummary(t),
		Context:     t.Context,
		Notes:       t.Notes,
		Blocker:     t.Blocker,
	}
}
