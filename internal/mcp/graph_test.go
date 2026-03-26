package mcp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestEdgeTypeConstants(t *testing.T) {
	t.Parallel()

	types := []EdgeType{EdgeBlocks, EdgeRelatedTo, EdgeSubtaskOf, EdgeDuplicate, EdgeSequence, EdgeCrossRef}
	expected := []string{"blocks", "related-to", "subtask-of", "duplicate-of", "sequential", "cross-ref"}
	for i, et := range types {
		if string(et) != expected[i] {
			t.Errorf("EdgeType %d = %q, want %q", i, et, expected[i])
		}
	}
}

func TestRelationshipInferencerTextSimilarity(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "deploy kubernetes cluster to staging env now", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "2", Text: "deploy kubernetes cluster to production env now", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "3", Text: "buy groceries from store", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	ri := NewRelationshipInferencer(pool)
	edges := ri.inferTextSimilarity(pool.GetAllTasks())

	found := false
	for _, e := range edges {
		if (e.FromID == "1" && e.ToID == "2") || (e.FromID == "2" && e.ToID == "1") {
			found = true
			if e.Type != EdgeRelatedTo {
				t.Errorf("edge type = %q, want %q", e.Type, EdgeRelatedTo)
			}
			if e.Source != "inferred:text" {
				t.Errorf("source = %q, want inferred:text", e.Source)
			}
			if e.Weight <= 0.7 {
				t.Errorf("weight = %f, want > 0.7", e.Weight)
			}
		}
	}
	if !found {
		t.Error("expected related-to edge between tasks 1 and 2")
	}
}

func TestRelationshipInferencerDuplicates(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "deploy the kubernetes cluster", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "2", Text: "deploy the kubernetes cluster", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	ri := NewRelationshipInferencer(pool)
	edges := ri.inferDuplicates(pool.GetAllTasks())

	if len(edges) != 1 {
		t.Fatalf("expected 1 duplicate edge, got %d", len(edges))
	}
	if edges[0].Type != EdgeDuplicate {
		t.Errorf("type = %q, want duplicate-of", edges[0].Type)
	}
	if edges[0].Weight < 0.95 {
		t.Errorf("weight = %f, want >= 0.95", edges[0].Weight)
	}
}

func TestRelationshipInferencerTemporalSequence(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	t1 := now.Add(-20 * time.Minute)
	t2 := now.Add(-5 * time.Minute)
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "setup deploy pipeline", Status: core.StatusComplete, CreatedAt: now, UpdatedAt: now, CompletedAt: &t1})
	pool.AddTask(&core.Task{ID: "2", Text: "run deploy pipeline test", Status: core.StatusComplete, CreatedAt: now, UpdatedAt: now, CompletedAt: &t2})

	ri := NewRelationshipInferencer(pool)
	edges := ri.inferTemporalSequence(pool.GetAllTasks())

	if len(edges) == 0 {
		t.Fatal("expected temporal sequence edges")
	}
	if edges[0].Type != EdgeSequence {
		t.Errorf("type = %q, want sequential", edges[0].Type)
	}
	if edges[0].Source != "inferred:temporal" {
		t.Errorf("source = %q, want inferred:temporal", edges[0].Source)
	}
}

func TestRelationshipInferencerBlockerChains(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "auth-fix", Text: "fix authentication bug", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "deploy", Text: "deploy to production", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, Blocker: "waiting on auth-fix"})

	ri := NewRelationshipInferencer(pool)
	edges := ri.inferBlockerChains(pool.GetAllTasks())

	if len(edges) == 0 {
		t.Fatal("expected blocker edges")
	}
	found := false
	for _, e := range edges {
		if e.FromID == "auth-fix" && e.ToID == "deploy" && e.Type == EdgeBlocks {
			found = true
		}
	}
	if !found {
		t.Error("expected auth-fix blocks deploy edge")
	}
}

func TestRelationshipInferencerSubtaskPatterns(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "write tests", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "2", Text: "write tests for authentication module and validate coverage", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	ri := NewRelationshipInferencer(pool)
	edges := ri.inferSubtaskPatterns(pool.GetAllTasks())

	found := false
	for _, e := range edges {
		if e.FromID == "1" && e.ToID == "2" && e.Type == EdgeSubtaskOf {
			found = true
		}
	}
	if !found {
		t.Error("expected subtask-of edge from task 1 to task 2")
	}
}

func TestRelationshipInferencerCrossProviderRefs(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{
		ID: "1", Text: "task one", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now,
		SourceRefs: []core.SourceRef{{Provider: "jira", NativeID: "PROJ-123"}},
	})
	pool.AddTask(&core.Task{
		ID: "2", Text: "task two", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now,
		SourceRefs: []core.SourceRef{{Provider: "jira", NativeID: "PROJ-123"}},
	})

	ri := NewRelationshipInferencer(pool)
	edges := ri.inferCrossProviderRefs(pool.GetAllTasks())

	if len(edges) != 1 {
		t.Fatalf("expected 1 cross-ref edge, got %d", len(edges))
	}
	if edges[0].Type != EdgeCrossRef {
		t.Errorf("type = %q, want cross-ref", edges[0].Type)
	}
	if edges[0].Weight != 1.0 {
		t.Errorf("weight = %f, want 1.0", edges[0].Weight)
	}
}

func TestInferAllCombinesStrategies(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "deploy the kubernetes cluster", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "2", Text: "deploy the kubernetes cluster", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	ri := NewRelationshipInferencer(pool)
	edges := ri.InferAll()

	if len(edges) == 0 {
		t.Fatal("expected edges from InferAll")
	}
}

func TestInferAllSingleTask(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "solo task", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	ri := NewRelationshipInferencer(pool)
	edges := ri.InferAll()
	if len(edges) != 0 {
		t.Errorf("expected 0 edges for single task, got %d", len(edges))
	}
}

func TestWalkGraph(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "a", Text: "task a", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "b", Text: "task b", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "c", Text: "task c", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "a", ToID: "b", Type: EdgeBlocks, Weight: 0.9, Source: "explicit"},
		{FromID: "b", ToID: "c", Type: EdgeBlocks, Weight: 0.8, Source: "explicit"},
	}

	graph, err := WalkGraph(pool, edges, WalkGraphOptions{TaskID: "a", Depth: 2})
	if err != nil {
		t.Fatalf("WalkGraph: %v", err)
		return
	}

	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph.Nodes))
	}
	if graph.Nodes["a"].Depth != 0 {
		t.Errorf("node a depth = %d, want 0", graph.Nodes["a"].Depth)
	}
	if graph.Nodes["b"].Depth != 1 {
		t.Errorf("node b depth = %d, want 1", graph.Nodes["b"].Depth)
	}
}

func TestWalkGraphDepthLimit(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "a", Text: "task a", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "b", Text: "task b", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "c", Text: "task c", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "a", ToID: "b", Type: EdgeBlocks, Weight: 0.9, Source: "explicit"},
		{FromID: "b", ToID: "c", Type: EdgeBlocks, Weight: 0.8, Source: "explicit"},
	}

	graph, err := WalkGraph(pool, edges, WalkGraphOptions{TaskID: "a", Depth: 1})
	if err != nil {
		t.Fatalf("WalkGraph: %v", err)
		return
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes at depth 1, got %d", len(graph.Nodes))
	}
	if _, ok := graph.Nodes["c"]; ok {
		t.Error("node c should not be included at depth 1")
	}
}

func TestWalkGraphEdgeTypeFilter(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "a", Text: "task a", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "b", Text: "task b", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "c", Text: "task c", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "a", ToID: "b", Type: EdgeBlocks, Weight: 0.9, Source: "explicit"},
		{FromID: "a", ToID: "c", Type: EdgeRelatedTo, Weight: 0.5, Source: "inferred:text"},
	}

	graph, err := WalkGraph(pool, edges, WalkGraphOptions{
		TaskID:    "a",
		Depth:     2,
		EdgeTypes: []EdgeType{EdgeBlocks},
	})
	if err != nil {
		t.Fatalf("WalkGraph: %v", err)
		return
	}

	if _, ok := graph.Nodes["c"]; ok {
		t.Error("node c should not be included when filtering for blocks only")
	}
	if _, ok := graph.Nodes["b"]; !ok {
		t.Error("node b should be included")
	}
}

func TestWalkGraphNotFound(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	_, err := WalkGraph(pool, nil, WalkGraphOptions{TaskID: "missing"})
	if err == nil {
		t.Fatal("expected error for missing task")
		return
	}
}

func TestFindPaths(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "a", Text: "a", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "b", Text: "b", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "c", Text: "c", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "a", ToID: "b", Type: EdgeBlocks},
		{FromID: "b", ToID: "c", Type: EdgeBlocks},
		{FromID: "a", ToID: "c", Type: EdgeRelatedTo},
	}

	paths, err := FindPaths(pool, edges, "a", "c", 5)
	if err != nil {
		t.Fatalf("FindPaths: %v", err)
		return
	}

	if len(paths) < 2 {
		t.Errorf("expected at least 2 paths, got %d", len(paths))
	}
}

func TestFindPathsNotFound(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	_, err := FindPaths(pool, nil, "missing", "also-missing", 5)
	if err == nil {
		t.Fatal("expected error for missing task")
		return
	}
}

func TestGetCriticalPath(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "a", Text: "a", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "b", Text: "b", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "c", Text: "c", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "a", ToID: "b", Type: EdgeBlocks},
		{FromID: "b", ToID: "c", Type: EdgeBlocks},
		{FromID: "a", ToID: "c", Type: EdgeRelatedTo}, // not blocks, should be ignored
	}

	path, err := GetCriticalPath(pool, edges, "a")
	if err != nil {
		t.Fatalf("GetCriticalPath: %v", err)
		return
	}

	if len(path) != 3 {
		t.Errorf("expected path length 3, got %d: %v", len(path), path)
	}
	if path[0] != "a" || path[1] != "b" || path[2] != "c" {
		t.Errorf("expected path [a b c], got %v", path)
	}
}

func TestGetCriticalPathNotFound(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	_, err := GetCriticalPath(pool, nil, "missing")
	if err == nil {
		t.Fatal("expected error for missing task")
		return
	}
}

func TestGetOrphans(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "a", Text: "connected", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "b", Text: "connected", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "c", Text: "orphan", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "a", ToID: "b", Type: EdgeBlocks},
	}

	orphans := GetOrphans(pool, edges)
	if len(orphans) != 1 {
		t.Fatalf("expected 1 orphan, got %d", len(orphans))
	}
	if orphans[0].ID != "c" {
		t.Errorf("orphan id = %q, want c", orphans[0].ID)
	}
}

func TestGetClusters(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "a", Text: "a", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "b", Text: "b", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "c", Text: "c", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "d", Text: "d", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "a", ToID: "b", Type: EdgeBlocks},
		{FromID: "c", ToID: "d", Type: EdgeRelatedTo},
	}

	clusters := GetClusters(pool, edges)
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
	for _, c := range clusters {
		if len(c.Tasks) != 2 {
			t.Errorf("expected cluster size 2, got %d", len(c.Tasks))
		}
	}
}

func TestGetClustersNoEdges(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	clusters := GetClusters(pool, nil)
	if len(clusters) != 0 {
		t.Errorf("expected 0 clusters, got %d", len(clusters))
	}
}

func TestCrossProviderLinkerOverlap(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "deploy kubernetes cluster", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "local"})
	pool.AddTask(&core.Task{ID: "2", Text: "deploy kubernetes cluster staging", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "jira"})
	pool.AddTask(&core.Task{ID: "3", Text: "buy groceries", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "jira"})

	linker := NewCrossProviderLinker(pool)
	overlap := linker.GetProviderOverlap("local", "jira", nil)

	if overlap.ProviderA != "local" || overlap.ProviderB != "jira" {
		t.Errorf("providers = %q/%q, want local/jira", overlap.ProviderA, overlap.ProviderB)
	}
	if len(overlap.SimilarPairs) == 0 {
		t.Error("expected similar pairs between local and jira tasks")
	}
}

func TestCrossProviderLinkerUnifiedView(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "deploy kubernetes to staging", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "local"})
	pool.AddTask(&core.Task{ID: "2", Text: "kubernetes monitoring setup", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "jira"})
	pool.AddTask(&core.Task{ID: "3", Text: "buy groceries", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "local"})

	linker := NewCrossProviderLinker(pool)
	view := linker.GetUnifiedView("kubernetes", nil)

	if view.Topic != "kubernetes" {
		t.Errorf("topic = %q, want kubernetes", view.Topic)
	}
	if len(view.Tasks) < 2 {
		t.Errorf("expected at least 2 matching tasks, got %d", len(view.Tasks))
	}
}

func TestCrossProviderLinkerSuggestCrossLinks(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "deploy the kubernetes cluster to production", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "local"})
	pool.AddTask(&core.Task{ID: "2", Text: "deploy the kubernetes cluster to production", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "jira"})

	linker := NewCrossProviderLinker(pool)
	suggestions := linker.SuggestCrossLinks()

	if len(suggestions) == 0 {
		t.Fatal("expected cross-link suggestions")
	}
	found := false
	for _, s := range suggestions {
		if s.EdgeType == EdgeDuplicate {
			found = true
		}
	}
	if !found {
		t.Error("expected duplicate suggestion for identical tasks across providers")
	}
}

func TestCrossProviderLinkerSuggestNoProviders(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "1", Text: "task one", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "local"})

	linker := NewCrossProviderLinker(pool)
	suggestions := linker.SuggestCrossLinks()
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for single provider, got %d", len(suggestions))
	}
}

func TestDeduplicateEdges(t *testing.T) {
	t.Parallel()

	edges := []TaskEdge{
		{FromID: "a", ToID: "b", Type: EdgeBlocks, Weight: 0.9},
		{FromID: "a", ToID: "b", Type: EdgeBlocks, Weight: 0.8},
		{FromID: "a", ToID: "b", Type: EdgeRelatedTo, Weight: 0.5},
	}

	deduped := deduplicateEdges(edges)
	if len(deduped) != 2 {
		t.Errorf("expected 2 unique edges, got %d", len(deduped))
	}
}

func TestMatchesEdgeTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		et     EdgeType
		filter []EdgeType
		want   bool
	}{
		{"empty filter matches all", EdgeBlocks, nil, true},
		{"matching type", EdgeBlocks, []EdgeType{EdgeBlocks, EdgeRelatedTo}, true},
		{"non-matching type", EdgeSequence, []EdgeType{EdgeBlocks}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchesEdgeTypes(tt.et, tt.filter)
			if got != tt.want {
				t.Errorf("matchesEdgeTypes(%q, %v) = %v, want %v", tt.et, tt.filter, got, tt.want)
			}
		})
	}
}

// Tool integration tests

func TestToolWalkGraph(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "a", Text: "deploy the kubernetes cluster", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
		&core.Task{ID: "b", Text: "deploy the kubernetes cluster", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "walk_graph", map[string]any{"task_id": "a", "depth": 2})
	text := parseToolText(t, resp)

	var data struct {
		Graph struct {
			Nodes map[string]json.RawMessage `json:"nodes"`
			Edges []json.RawMessage          `json:"edges"`
		} `json:"graph"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.Graph.Nodes) == 0 {
		t.Error("expected nodes in graph")
	}
}

func TestToolWalkGraphMissingID(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "walk_graph", map[string]any{})
	if resp.Error == nil {
		t.Fatal("expected error for missing task_id")
		return
	}
}

func TestToolFindPaths(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "a", Text: "task a", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
		&core.Task{ID: "b", Text: "task b", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "find_paths", map[string]any{"from_id": "a", "to_id": "b"})
	text := parseToolText(t, resp)

	var data struct {
		Paths []json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func TestToolFindPathsMissingParams(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "find_paths", map[string]any{"from_id": "a"})
	if resp.Error == nil {
		t.Fatal("expected error for missing to_id")
		return
	}
}

func TestToolGetCriticalPath(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "root", Text: "root task", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "get_critical_path", map[string]any{"root_id": "root"})
	text := parseToolText(t, resp)

	var data struct {
		Path   []string `json:"path"`
		Length int      `json:"length"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Length == 0 {
		t.Error("expected at least root in path")
	}
}

func TestToolGetCriticalPathMissingID(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "get_critical_path", map[string]any{})
	if resp.Error == nil {
		t.Fatal("expected error for missing root_id")
		return
	}
}

func TestToolGetOrphans(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "standalone task", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "get_orphans", nil)
	text := parseToolText(t, resp)

	var data struct {
		Tasks []json.RawMessage `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func TestToolGetClusters(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "get_clusters", nil)
	text := parseToolText(t, resp)

	var data struct {
		Clusters []json.RawMessage `json:"clusters"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func TestToolGetProviderOverlap(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "task", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "local"},
		&core.Task{ID: "2", Text: "task", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "jira"},
	)

	resp := dispatchToolCall(t, s, "get_provider_overlap", map[string]any{"provider_a": "local", "provider_b": "jira"})
	text := parseToolText(t, resp)

	var data struct {
		Overlap struct {
			ProviderA string `json:"provider_a"`
			ProviderB string `json:"provider_b"`
		} `json:"overlap"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Overlap.ProviderA != "local" {
		t.Errorf("provider_a = %q, want local", data.Overlap.ProviderA)
	}
}

func TestToolGetProviderOverlapMissingParams(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "get_provider_overlap", map[string]any{"provider_a": "local"})
	if resp.Error == nil {
		t.Fatal("expected error for missing provider_b")
		return
	}
}

func TestToolGetUnifiedView(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "kubernetes deployment", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "get_unified_view", map[string]any{"topic": "kubernetes"})
	text := parseToolText(t, resp)

	var data struct {
		View struct {
			Topic string `json:"topic"`
		} `json:"view"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.View.Topic != "kubernetes" {
		t.Errorf("topic = %q, want kubernetes", data.View.Topic)
	}
}

func TestToolGetUnifiedViewMissingTopic(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "get_unified_view", map[string]any{})
	if resp.Error == nil {
		t.Fatal("expected error for missing topic")
		return
	}
}

func TestToolSuggestCrossLinks(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "suggest_cross_links", nil)
	text := parseToolText(t, resp)

	var data struct {
		Suggestions []json.RawMessage `json:"suggestions"`
	}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

// Resource integration tests

func TestResourceGraphDependencies(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "task one", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchRead(t, s, "threedoors://graph/dependencies")
	if resp.Error != nil {
		t.Fatalf("unexpected error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
		return
	}
}

func TestResourceGraphCrossProvider(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "1", Text: "task one", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now, SourceProvider: "local"},
	)

	resp := dispatchRead(t, s, "threedoors://graph/cross-provider")
	if resp.Error != nil {
		t.Fatalf("unexpected error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
		return
	}
}

func TestToolsListIncludesGraphTools(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/list",
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ToolsListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	names := make(map[string]bool)
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}

	expected := []string{
		"walk_graph", "find_paths", "get_critical_path",
		"get_orphans", "get_clusters",
		"get_provider_overlap", "get_unified_view", "suggest_cross_links",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing graph tool: %s", name)
		}
	}
}

func TestResourcesListIncludesGraphResources(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "resources/list",
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourcesListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	uris := make(map[string]bool)
	for _, r := range result.Resources {
		uris[r.URI] = true
	}

	expected := []string{"threedoors://graph/dependencies", "threedoors://graph/cross-provider"}
	for _, uri := range expected {
		if !uris[uri] {
			t.Errorf("missing graph resource: %s", uri)
		}
	}
}
