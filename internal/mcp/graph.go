package mcp

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// EdgeType classifies the relationship between two tasks.
type EdgeType string

const (
	EdgeBlocks    EdgeType = "blocks"
	EdgeRelatedTo EdgeType = "related-to"
	EdgeSubtaskOf EdgeType = "subtask-of"
	EdgeDuplicate EdgeType = "duplicate-of"
	EdgeSequence  EdgeType = "sequential"
	EdgeCrossRef  EdgeType = "cross-ref"
)

// TaskNode wraps a task with graph metadata.
type TaskNode struct {
	Task     *core.Task `json:"task"`
	Provider string     `json:"provider"`
	Depth    int        `json:"depth"`
}

// TaskEdge represents a directed relationship between two tasks.
type TaskEdge struct {
	FromID string   `json:"from_id"`
	ToID   string   `json:"to_id"`
	Type   EdgeType `json:"type"`
	Weight float64  `json:"weight"`
	Source string   `json:"source"`
}

// TaskGraph holds computed nodes and edges.
type TaskGraph struct {
	Nodes map[string]*TaskNode `json:"nodes"`
	Edges []TaskEdge           `json:"edges"`
}

// RelationshipInferencer discovers edges between tasks using multiple strategies.
type RelationshipInferencer struct {
	pool *core.TaskPool
}

// NewRelationshipInferencer creates an inferencer backed by the given pool.
func NewRelationshipInferencer(pool *core.TaskPool) *RelationshipInferencer {
	return &RelationshipInferencer{pool: pool}
}

// InferAll runs all inference strategies and returns the combined edge set.
func (ri *RelationshipInferencer) InferAll() []TaskEdge {
	tasks := ri.pool.GetAllTasks()
	if len(tasks) < 2 {
		return nil
	}

	var edges []TaskEdge
	edges = append(edges, ri.inferTextSimilarity(tasks)...)
	edges = append(edges, ri.inferTemporalSequence(tasks)...)
	edges = append(edges, ri.inferCrossProviderRefs(tasks)...)
	edges = append(edges, ri.inferBlockerChains(tasks)...)
	edges = append(edges, ri.inferSubtaskPatterns(tasks)...)
	edges = append(edges, ri.inferDuplicates(tasks)...)
	return deduplicateEdges(edges)
}

// inferTextSimilarity finds tasks with >70% token overlap.
func (ri *RelationshipInferencer) inferTextSimilarity(tasks []*core.Task) []TaskEdge {
	var edges []TaskEdge
	for i := 0; i < len(tasks); i++ {
		tokensI := tokenize(tasks[i].Text)
		if len(tokensI) == 0 {
			continue
		}
		for j := i + 1; j < len(tasks); j++ {
			tokensJ := tokenize(tasks[j].Text)
			if len(tokensJ) == 0 {
				continue
			}
			sim := jaccardSimilarity(tokensI, tokensJ)
			if sim > 0.7 && sim < 0.95 {
				edges = append(edges, TaskEdge{
					FromID: tasks[i].ID,
					ToID:   tasks[j].ID,
					Type:   EdgeRelatedTo,
					Weight: sim,
					Source: "inferred:text",
				})
			}
		}
	}
	return edges
}

// inferTemporalSequence finds tasks completed within 30min with similar text.
func (ri *RelationshipInferencer) inferTemporalSequence(tasks []*core.Task) []TaskEdge {
	var completed []*core.Task
	for _, t := range tasks {
		if t.CompletedAt != nil {
			completed = append(completed, t)
		}
	}
	sort.Slice(completed, func(i, j int) bool {
		return completed[i].CompletedAt.Before(*completed[j].CompletedAt)
	})

	var edges []TaskEdge
	for i := 0; i < len(completed)-1; i++ {
		a := completed[i]
		b := completed[i+1]
		diff := b.CompletedAt.Sub(*a.CompletedAt)
		if diff < 0 {
			diff = -diff
		}
		if diff <= 30*time.Minute {
			tokA := tokenize(a.Text)
			tokB := tokenize(b.Text)
			sim := jaccardSimilarity(tokA, tokB)
			if sim > 0.2 {
				edges = append(edges, TaskEdge{
					FromID: a.ID,
					ToID:   b.ID,
					Type:   EdgeSequence,
					Weight: 0.6 + sim*0.4,
					Source: "inferred:temporal",
				})
			}
		}
	}
	return edges
}

// inferCrossProviderRefs finds tasks sharing SourceRef entries across providers.
func (ri *RelationshipInferencer) inferCrossProviderRefs(tasks []*core.Task) []TaskEdge {
	type refKey struct {
		Provider string
		NativeID string
	}
	refIndex := make(map[refKey][]string) // refKey → task IDs
	for _, t := range tasks {
		for _, ref := range t.SourceRefs {
			key := refKey{ref.Provider, ref.NativeID}
			refIndex[key] = append(refIndex[key], t.ID)
		}
	}

	var edges []TaskEdge
	for _, ids := range refIndex {
		if len(ids) < 2 {
			continue
		}
		for i := 0; i < len(ids); i++ {
			for j := i + 1; j < len(ids); j++ {
				edges = append(edges, TaskEdge{
					FromID: ids[i],
					ToID:   ids[j],
					Type:   EdgeCrossRef,
					Weight: 1.0,
					Source: "explicit",
				})
			}
		}
	}
	return edges
}

// inferBlockerChains finds tasks whose blocker text mentions another task's text or ID.
func (ri *RelationshipInferencer) inferBlockerChains(tasks []*core.Task) []TaskEdge {
	var edges []TaskEdge
	for _, t := range tasks {
		if t.Blocker == "" {
			continue
		}
		blockerLower := strings.ToLower(t.Blocker)
		for _, other := range tasks {
			if other.ID == t.ID {
				continue
			}
			if strings.Contains(blockerLower, strings.ToLower(other.ID)) ||
				(len(other.Text) > 3 && strings.Contains(blockerLower, strings.ToLower(other.Text))) {
				edges = append(edges, TaskEdge{
					FromID: other.ID,
					ToID:   t.ID,
					Type:   EdgeBlocks,
					Weight: 0.8,
					Source: "inferred:text",
				})
			}
		}
	}
	return edges
}

// inferSubtaskPatterns finds tasks whose text is a subset of another task's text.
func (ri *RelationshipInferencer) inferSubtaskPatterns(tasks []*core.Task) []TaskEdge {
	var edges []TaskEdge
	for i := 0; i < len(tasks); i++ {
		tokI := tokenize(tasks[i].Text)
		if len(tokI) < 2 {
			continue
		}
		setI := tokenSet(tokI)
		for j := 0; j < len(tasks); j++ {
			if i == j {
				continue
			}
			tokJ := tokenize(tasks[j].Text)
			if len(tokJ) < 2 || len(tokI) >= len(tokJ) {
				continue
			}
			setJ := tokenSet(tokJ)
			// Check if i's tokens are a subset of j's tokens.
			contained := 0
			for tok := range setI {
				if _, ok := setJ[tok]; ok {
					contained++
				}
			}
			ratio := float64(contained) / float64(len(setI))
			if ratio >= 0.8 && float64(len(setI))/float64(len(setJ)) < 0.8 {
				edges = append(edges, TaskEdge{
					FromID: tasks[i].ID,
					ToID:   tasks[j].ID,
					Type:   EdgeSubtaskOf,
					Weight: ratio,
					Source: "inferred:text",
				})
			}
		}
	}
	return edges
}

// inferDuplicates finds tasks with >=95% token overlap.
func (ri *RelationshipInferencer) inferDuplicates(tasks []*core.Task) []TaskEdge {
	var edges []TaskEdge
	for i := 0; i < len(tasks); i++ {
		tokI := tokenize(tasks[i].Text)
		if len(tokI) == 0 {
			continue
		}
		for j := i + 1; j < len(tasks); j++ {
			tokJ := tokenize(tasks[j].Text)
			if len(tokJ) == 0 {
				continue
			}
			sim := jaccardSimilarity(tokI, tokJ)
			if sim >= 0.95 {
				edges = append(edges, TaskEdge{
					FromID: tasks[i].ID,
					ToID:   tasks[j].ID,
					Type:   EdgeDuplicate,
					Weight: sim,
					Source: "inferred:text",
				})
			}
		}
	}
	return edges
}

// WalkGraphOptions controls the walk_graph traversal.
type WalkGraphOptions struct {
	TaskID    string     `json:"task_id"`
	Depth     int        `json:"depth"`
	Direction string     `json:"direction"`
	EdgeTypes []EdgeType `json:"edge_types,omitempty"`
}

// WalkGraph performs BFS from a root task and returns the subgraph within depth.
func WalkGraph(pool *core.TaskPool, edges []TaskEdge, opts WalkGraphOptions) (*TaskGraph, error) {
	if opts.Depth <= 0 {
		opts.Depth = 2
	}
	if opts.Direction == "" {
		opts.Direction = "both"
	}

	root := pool.GetTask(opts.TaskID)
	if root == nil {
		return nil, fmt.Errorf("task not found: %s", opts.TaskID)
	}

	// Build adjacency lists.
	outgoing := make(map[string][]int)
	incoming := make(map[string][]int)
	for i, e := range edges {
		if !matchesEdgeTypes(e.Type, opts.EdgeTypes) {
			continue
		}
		outgoing[e.FromID] = append(outgoing[e.FromID], i)
		incoming[e.ToID] = append(incoming[e.ToID], i)
	}

	graph := &TaskGraph{
		Nodes: make(map[string]*TaskNode),
	}
	graph.Nodes[opts.TaskID] = &TaskNode{
		Task:     root,
		Provider: root.EffectiveSourceProvider(),
		Depth:    0,
	}

	type bfsItem struct {
		id    string
		depth int
	}
	queue := []bfsItem{{id: opts.TaskID, depth: 0}}
	visited := map[string]bool{opts.TaskID: true}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.depth >= opts.Depth {
			continue
		}

		var neighborEdgeIdxs []int
		switch opts.Direction {
		case "outgoing":
			neighborEdgeIdxs = outgoing[item.id]
		case "incoming":
			neighborEdgeIdxs = incoming[item.id]
		default: // "both"
			neighborEdgeIdxs = append(outgoing[item.id], incoming[item.id]...)
		}

		for _, idx := range neighborEdgeIdxs {
			e := edges[idx]
			graph.Edges = append(graph.Edges, e)

			neighborID := e.ToID
			if neighborID == item.id {
				neighborID = e.FromID
			}

			if visited[neighborID] {
				continue
			}
			visited[neighborID] = true

			t := pool.GetTask(neighborID)
			if t == nil {
				continue
			}
			graph.Nodes[neighborID] = &TaskNode{
				Task:     t,
				Provider: t.EffectiveSourceProvider(),
				Depth:    item.depth + 1,
			}
			queue = append(queue, bfsItem{id: neighborID, depth: item.depth + 1})
		}
	}

	if graph.Edges == nil {
		graph.Edges = []TaskEdge{}
	}
	return graph, nil
}

// FindPaths returns all simple paths between from and to up to maxDepth.
func FindPaths(pool *core.TaskPool, edges []TaskEdge, fromID, toID string, maxDepth int) ([][]string, error) {
	if maxDepth <= 0 {
		maxDepth = 5
	}

	if pool.GetTask(fromID) == nil {
		return nil, fmt.Errorf("task not found: %s", fromID)
	}
	if pool.GetTask(toID) == nil {
		return nil, fmt.Errorf("task not found: %s", toID)
	}

	adj := make(map[string][]string)
	for _, e := range edges {
		adj[e.FromID] = append(adj[e.FromID], e.ToID)
		adj[e.ToID] = append(adj[e.ToID], e.FromID)
	}

	var paths [][]string
	var dfs func(current string, path []string, visited map[string]bool)
	dfs = func(current string, path []string, visited map[string]bool) {
		if current == toID {
			p := make([]string, len(path))
			copy(p, path)
			paths = append(paths, p)
			return
		}
		if len(path) > maxDepth {
			return
		}
		for _, neighbor := range adj[current] {
			if visited[neighbor] {
				continue
			}
			visited[neighbor] = true
			dfs(neighbor, append(path, neighbor), visited)
			delete(visited, neighbor)
		}
	}

	visited := map[string]bool{fromID: true}
	dfs(fromID, []string{fromID}, visited)
	return paths, nil
}

// GetCriticalPath returns the longest dependency chain from the root task,
// following only "blocks" edges.
func GetCriticalPath(pool *core.TaskPool, edges []TaskEdge, rootID string) ([]string, error) {
	if pool.GetTask(rootID) == nil {
		return nil, fmt.Errorf("task not found: %s", rootID)
	}

	// Build adjacency for blocks edges only (from → to means "from blocks to").
	adj := make(map[string][]string)
	for _, e := range edges {
		if e.Type == EdgeBlocks {
			adj[e.FromID] = append(adj[e.FromID], e.ToID)
		}
	}

	var longest []string
	var dfs func(id string, path []string, visited map[string]bool)
	dfs = func(id string, path []string, visited map[string]bool) {
		if len(path) > len(longest) {
			longest = make([]string, len(path))
			copy(longest, path)
		}
		for _, next := range adj[id] {
			if visited[next] {
				continue
			}
			visited[next] = true
			dfs(next, append(path, next), visited)
			delete(visited, next)
		}
	}

	visited := map[string]bool{rootID: true}
	dfs(rootID, []string{rootID}, visited)
	return longest, nil
}

// GetOrphans returns tasks with no relationships in the edge set.
func GetOrphans(pool *core.TaskPool, edges []TaskEdge) []*core.Task {
	connected := make(map[string]bool)
	for _, e := range edges {
		connected[e.FromID] = true
		connected[e.ToID] = true
	}

	var orphans []*core.Task
	for _, t := range pool.GetAllTasks() {
		if !connected[t.ID] {
			orphans = append(orphans, t)
		}
	}
	return orphans
}

// Cluster is a group of related tasks found by connected-component detection.
type Cluster struct {
	Tasks []string `json:"tasks"`
}

// GetClusters returns groups of related tasks using connected-component detection.
func GetClusters(pool *core.TaskPool, edges []TaskEdge) []Cluster {
	adj := make(map[string][]string)
	nodeSet := make(map[string]bool)
	for _, e := range edges {
		adj[e.FromID] = append(adj[e.FromID], e.ToID)
		adj[e.ToID] = append(adj[e.ToID], e.FromID)
		nodeSet[e.FromID] = true
		nodeSet[e.ToID] = true
	}

	visited := make(map[string]bool)
	var clusters []Cluster

	for node := range nodeSet {
		if visited[node] {
			continue
		}
		var component []string
		queue := []string{node}
		visited[node] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			component = append(component, curr)

			for _, neighbor := range adj[curr] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}
		sort.Strings(component)
		clusters = append(clusters, Cluster{Tasks: component})
	}

	sort.Slice(clusters, func(i, j int) bool {
		return len(clusters[i].Tasks) > len(clusters[j].Tasks)
	})
	return clusters
}

// CrossProviderLinker discovers relationships between tasks from different providers.
type CrossProviderLinker struct {
	pool *core.TaskPool
}

// NewCrossProviderLinker creates a linker backed by the given pool.
func NewCrossProviderLinker(pool *core.TaskPool) *CrossProviderLinker {
	return &CrossProviderLinker{pool: pool}
}

// ProviderOverlap describes shared tasks between two providers.
type ProviderOverlap struct {
	ProviderA    string        `json:"provider_a"`
	ProviderB    string        `json:"provider_b"`
	SharedCount  int           `json:"shared_count"`
	SharedEdges  []TaskEdge    `json:"shared_edges"`
	SimilarPairs []SimilarPair `json:"similar_pairs"`
}

// SimilarPair is a pair of tasks from different providers with a similarity score.
type SimilarPair struct {
	TaskA      string  `json:"task_a"`
	ProviderA  string  `json:"provider_a"`
	TaskB      string  `json:"task_b"`
	ProviderB  string  `json:"provider_b"`
	Similarity float64 `json:"similarity"`
	Reason     string  `json:"reason"`
}

// GetProviderOverlap finds shared or similar tasks between two providers.
func (cl *CrossProviderLinker) GetProviderOverlap(providerA, providerB string, edges []TaskEdge) ProviderOverlap {
	tasks := cl.pool.GetAllTasks()

	var tasksA, tasksB []*core.Task
	for _, t := range tasks {
		p := t.EffectiveSourceProvider()
		if p == providerA {
			tasksA = append(tasksA, t)
		}
		if p == providerB {
			tasksB = append(tasksB, t)
		}
	}

	result := ProviderOverlap{
		ProviderA: providerA,
		ProviderB: providerB,
	}

	// Find edges that cross providers.
	providerOf := make(map[string]string)
	for _, t := range tasks {
		providerOf[t.ID] = t.EffectiveSourceProvider()
	}
	for _, e := range edges {
		pFrom := providerOf[e.FromID]
		pTo := providerOf[e.ToID]
		if (pFrom == providerA && pTo == providerB) || (pFrom == providerB && pTo == providerA) {
			result.SharedEdges = append(result.SharedEdges, e)
			result.SharedCount++
		}
	}

	// Find similar pairs by text.
	for _, a := range tasksA {
		tokA := tokenize(a.Text)
		if len(tokA) == 0 {
			continue
		}
		for _, b := range tasksB {
			tokB := tokenize(b.Text)
			sim := jaccardSimilarity(tokA, tokB)
			if sim > 0.3 {
				result.SimilarPairs = append(result.SimilarPairs, SimilarPair{
					TaskA:      a.ID,
					ProviderA:  providerA,
					TaskB:      b.ID,
					ProviderB:  providerB,
					Similarity: sim,
					Reason:     "text_similarity",
				})
			}
		}
	}

	if result.SharedEdges == nil {
		result.SharedEdges = []TaskEdge{}
	}
	if result.SimilarPairs == nil {
		result.SimilarPairs = []SimilarPair{}
	}
	return result
}

// UnifiedView groups tasks by topic/keyword across all providers.
type UnifiedView struct {
	Topic string      `json:"topic"`
	Tasks []*TaskNode `json:"tasks"`
	Edges []TaskEdge  `json:"edges"`
}

// GetUnifiedView returns tasks matching a topic across all providers.
func (cl *CrossProviderLinker) GetUnifiedView(topic string, edges []TaskEdge) UnifiedView {
	topicTokens := tokenize(topic)
	tasks := cl.pool.GetAllTasks()

	view := UnifiedView{
		Topic: topic,
	}

	matchedIDs := make(map[string]bool)
	for _, t := range tasks {
		tokT := tokenize(t.Text)
		sim := jaccardSimilarity(topicTokens, tokT)
		if sim > 0.2 {
			view.Tasks = append(view.Tasks, &TaskNode{
				Task:     t,
				Provider: t.EffectiveSourceProvider(),
				Depth:    0,
			})
			matchedIDs[t.ID] = true
		}
	}

	for _, e := range edges {
		if matchedIDs[e.FromID] && matchedIDs[e.ToID] {
			view.Edges = append(view.Edges, e)
		}
	}

	if view.Tasks == nil {
		view.Tasks = []*TaskNode{}
	}
	if view.Edges == nil {
		view.Edges = []TaskEdge{}
	}
	return view
}

// CrossLinkSuggestion is a proposed cross-provider relationship.
type CrossLinkSuggestion struct {
	TaskA      string   `json:"task_a"`
	ProviderA  string   `json:"provider_a"`
	TaskB      string   `json:"task_b"`
	ProviderB  string   `json:"provider_b"`
	Confidence float64  `json:"confidence"`
	EdgeType   EdgeType `json:"edge_type"`
	Reason     string   `json:"reason"`
}

// SuggestCrossLinks proposes cross-provider relationships based on similarity heuristics.
func (cl *CrossProviderLinker) SuggestCrossLinks() []CrossLinkSuggestion {
	tasks := cl.pool.GetAllTasks()

	// Group tasks by provider.
	byProvider := make(map[string][]*core.Task)
	for _, t := range tasks {
		p := t.EffectiveSourceProvider()
		if p != "" {
			byProvider[p] = append(byProvider[p], t)
		}
	}

	providers := make([]string, 0, len(byProvider))
	for p := range byProvider {
		providers = append(providers, p)
	}
	sort.Strings(providers)

	var suggestions []CrossLinkSuggestion
	for i := 0; i < len(providers); i++ {
		for j := i + 1; j < len(providers); j++ {
			pA := providers[i]
			pB := providers[j]
			for _, a := range byProvider[pA] {
				tokA := tokenize(a.Text)
				if len(tokA) == 0 {
					continue
				}
				for _, b := range byProvider[pB] {
					tokB := tokenize(b.Text)
					sim := jaccardSimilarity(tokA, tokB)
					if sim >= 0.95 {
						suggestions = append(suggestions, CrossLinkSuggestion{
							TaskA: a.ID, ProviderA: pA,
							TaskB: b.ID, ProviderB: pB,
							Confidence: sim, EdgeType: EdgeDuplicate,
							Reason: "near-identical text across providers",
						})
					} else if sim > 0.5 {
						suggestions = append(suggestions, CrossLinkSuggestion{
							TaskA: a.ID, ProviderA: pA,
							TaskB: b.ID, ProviderB: pB,
							Confidence: sim, EdgeType: EdgeRelatedTo,
							Reason: "significant text overlap across providers",
						})
					}

					// Check shared SourceRef.
					for _, refA := range a.SourceRefs {
						if b.HasSourceRef(refA.Provider, refA.NativeID) {
							suggestions = append(suggestions, CrossLinkSuggestion{
								TaskA: a.ID, ProviderA: pA,
								TaskB: b.ID, ProviderB: pB,
								Confidence: 1.0, EdgeType: EdgeCrossRef,
								Reason: "shared source reference",
							})
						}
					}
				}
			}
		}
	}

	return suggestions
}

// tokenSet converts a token slice to a set.
func tokenSet(tokens []string) map[string]struct{} {
	s := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		s[t] = struct{}{}
	}
	return s
}

// matchesEdgeTypes checks if a type is in the filter list (empty = match all).
func matchesEdgeTypes(t EdgeType, filter []EdgeType) bool {
	if len(filter) == 0 {
		return true
	}
	for _, f := range filter {
		if f == t {
			return true
		}
	}
	return false
}

// deduplicateEdges removes duplicate edges (same from, to, type).
func deduplicateEdges(edges []TaskEdge) []TaskEdge {
	type edgeKey struct {
		from, to string
		typ      EdgeType
	}
	seen := make(map[edgeKey]bool)
	var result []TaskEdge
	for _, e := range edges {
		key := edgeKey{e.FromID, e.ToID, e.Type}
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, e)
	}
	return result
}
