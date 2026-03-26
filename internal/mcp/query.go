package mcp

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// SearchResult pairs a task with its relevance score and match metadata.
type SearchResult struct {
	Task      *core.Task `json:"task"`
	Score     float64    `json:"score"`
	MatchedOn []string   `json:"matched_on"`
}

// SearchOptions controls text search behavior.
type SearchOptions struct {
	MaxResults  int
	RecencyDays int // tasks updated within this window get a boost
}

// DefaultSearchOptions returns sensible defaults for text search.
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		MaxResults:  50,
		RecencyDays: 7,
	}
}

// TaskQueryEngine provides text-based search over a TaskPool with
// field weighting and recency boost.
type TaskQueryEngine struct {
	pool *core.TaskPool
}

// NewTaskQueryEngine creates a query engine backed by the given pool.
func NewTaskQueryEngine(pool *core.TaskPool) *TaskQueryEngine {
	return &TaskQueryEngine{pool: pool}
}

// Search performs a text query across all tasks, scoring by token overlap
// with field weighting (text 3x, context 2x, notes 1x) and recency boost.
func (qe *TaskQueryEngine) Search(query string, opts SearchOptions) []SearchResult {
	if opts.MaxResults <= 0 {
		opts.MaxResults = 50
	}

	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	tasks := qe.pool.GetAllTasks()
	var results []SearchResult

	for _, task := range tasks {
		score, matched := qe.scoreTask(task, queryTokens, opts.RecencyDays)
		if score > 0 {
			results = append(results, SearchResult{
				Task:      task,
				Score:     score,
				MatchedOn: matched,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}

	return results
}

// scoreTask computes a relevance score for a task against the query tokens.
func (qe *TaskQueryEngine) scoreTask(task *core.Task, queryTokens []string, recencyDays int) (float64, []string) {
	var totalScore float64
	var matched []string

	// Field weights: text 3x, context 2x, notes 1x.
	fields := []struct {
		name   string
		text   string
		weight float64
	}{
		{"text", task.Text, 3.0},
		{"context", task.Context, 2.0},
	}

	// Concatenate all notes into a single field.
	var noteTexts []string
	for _, n := range task.Notes {
		noteTexts = append(noteTexts, n.Text)
	}
	if len(noteTexts) > 0 {
		fields = append(fields, struct {
			name   string
			text   string
			weight float64
		}{"notes", strings.Join(noteTexts, " "), 1.0})
	}

	for _, f := range fields {
		if f.text == "" {
			continue
		}
		fieldTokens := tokenize(f.text)
		sim := jaccardSimilarity(queryTokens, fieldTokens)
		if sim > 0 {
			totalScore += sim * f.weight
			matched = append(matched, f.name)
		}
	}

	if totalScore == 0 {
		return 0, nil
	}

	// Recency boost: tasks updated recently get up to 20% bonus.
	if recencyDays > 0 {
		daysSinceUpdate := time.Since(task.UpdatedAt).Hours() / 24
		if daysSinceUpdate < float64(recencyDays) {
			boost := 1.0 + 0.2*(1.0-daysSinceUpdate/float64(recencyDays))
			totalScore *= boost
		}
	}

	// Normalize to 0.0–1.0 range. Max possible raw score is 6.0 (3+2+1) * 1.2 boost = 7.2.
	normalized := math.Min(totalScore/7.2, 1.0)
	return normalized, matched
}

// jaccardSimilarity computes |A ∩ B| / |A ∪ B| over two token sets.
func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	setB := make(map[string]struct{}, len(b))
	for _, tok := range b {
		setB[tok] = struct{}{}
	}

	var intersection int
	setA := make(map[string]struct{}, len(a))
	for _, tok := range a {
		setA[tok] = struct{}{}
		if _, ok := setB[tok]; ok {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// tokenize lowercases and splits text into word tokens.
func tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	// Deduplicate.
	seen := make(map[string]struct{}, len(words))
	var tokens []string
	for _, w := range words {
		// Strip common punctuation.
		w = strings.Trim(w, ".,;:!?\"'()[]{}")
		if w == "" {
			continue
		}
		if _, ok := seen[w]; !ok {
			seen[w] = struct{}{}
			tokens = append(tokens, w)
		}
	}
	return tokens
}

// FilterOptions defines criteria for filtering tasks.
type FilterOptions struct {
	Status        string `json:"status,omitempty"`
	Type          string `json:"type,omitempty"`
	Effort        string `json:"effort,omitempty"`
	Provider      string `json:"provider,omitempty"`
	TextContains  string `json:"text_contains,omitempty"`
	CreatedAfter  string `json:"created_after,omitempty"`
	CreatedBefore string `json:"created_before,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	SortBy        string `json:"sort_by,omitempty"`
	SortOrder     string `json:"sort_order,omitempty"`
}

// FilterTasks applies filter options to a task slice and returns matching tasks.
func FilterTasks(tasks []*core.Task, opts FilterOptions) []*core.Task {
	var result []*core.Task

	var createdAfter, createdBefore time.Time
	if opts.CreatedAfter != "" {
		createdAfter, _ = time.Parse(time.RFC3339, opts.CreatedAfter)
	}
	if opts.CreatedBefore != "" {
		createdBefore, _ = time.Parse(time.RFC3339, opts.CreatedBefore)
	}

	for _, t := range tasks {
		if opts.Status != "" && string(t.Status) != opts.Status {
			continue
		}
		if opts.Type != "" && string(t.Type) != opts.Type {
			continue
		}
		if opts.Effort != "" && string(t.Effort) != opts.Effort {
			continue
		}
		if opts.Provider != "" && t.EffectiveSourceProvider() != opts.Provider {
			continue
		}
		if opts.TextContains != "" && !strings.Contains(strings.ToLower(t.Text), strings.ToLower(opts.TextContains)) {
			continue
		}
		if !createdAfter.IsZero() && t.CreatedAt.Before(createdAfter) {
			continue
		}
		if !createdBefore.IsZero() && t.CreatedAt.After(createdBefore) {
			continue
		}
		result = append(result, t)
	}

	// Sort.
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	desc := strings.ToLower(opts.SortOrder) == "desc"

	sort.Slice(result, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "updated_at":
			less = result[i].UpdatedAt.Before(result[j].UpdatedAt)
		case "text":
			less = result[i].Text < result[j].Text
		default: // created_at
			less = result[i].CreatedAt.Before(result[j].CreatedAt)
		}
		if desc {
			return !less
		}
		return less
	})

	if opts.Limit > 0 && len(result) > opts.Limit {
		result = result[:opts.Limit]
	}

	return result
}
