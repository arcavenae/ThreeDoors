package mcp

import (
	"fmt"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func newTestPool(tasks ...*core.Task) *core.TaskPool {
	pool := core.NewTaskPool()
	for _, t := range tasks {
		pool.AddTask(t)
	}
	return pool
}

func newTestTask(id, text string) *core.Task {
	now := time.Now().UTC()
	return &core.Task{
		ID:        id,
		Text:      text,
		Status:    core.StatusTodo,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestTokenize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"single word", "hello", 1},
		{"multiple words", "hello world foo", 3},
		{"duplicates removed", "hello hello hello", 1},
		{"punctuation stripped", "hello, world! foo.", 3},
		{"case normalized", "Hello WORLD Foo", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tokens := tokenize(tt.input)
			if len(tokens) != tt.want {
				t.Errorf("tokenize(%q) = %d tokens, want %d", tt.input, len(tokens), tt.want)
			}
		})
	}
}

func TestJaccardSimilarity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b []string
		want float64
	}{
		{"empty both", nil, nil, 0},
		{"empty a", nil, []string{"x"}, 0},
		{"identical", []string{"a", "b"}, []string{"a", "b"}, 1.0},
		{"no overlap", []string{"a"}, []string{"b"}, 0},
		{"partial", []string{"a", "b"}, []string{"b", "c"}, 1.0 / 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := jaccardSimilarity(tt.a, tt.b)
			if abs(got-tt.want) > 0.001 {
				t.Errorf("jaccardSimilarity(%v, %v) = %f, want %f", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func TestSearchEmptyQuery(t *testing.T) {
	t.Parallel()

	pool := newTestPool(newTestTask("1", "buy groceries"))
	engine := NewTaskQueryEngine(pool)

	results := engine.Search("", DefaultSearchOptions())
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(results))
	}
}

func TestSearchFindsMatchingTasks(t *testing.T) {
	t.Parallel()

	pool := newTestPool(
		newTestTask("1", "buy groceries at the store"),
		newTestTask("2", "write documentation for API"),
		newTestTask("3", "buy milk and eggs"),
	)
	engine := NewTaskQueryEngine(pool)

	results := engine.Search("buy groceries", DefaultSearchOptions())
	if len(results) == 0 {
		t.Fatal("expected results for 'buy groceries'")
	}

	// Task 1 should score highest (most token overlap with "buy groceries").
	if results[0].Task.ID != "1" {
		t.Errorf("expected task 1 first, got task %s", results[0].Task.ID)
	}
	if results[0].Score <= 0 {
		t.Error("expected positive score")
	}
	if len(results[0].MatchedOn) == 0 {
		t.Error("expected matched_on to be populated")
	}
}

func TestSearchFieldWeighting(t *testing.T) {
	t.Parallel()

	// Task with match in text (3x weight) should score higher than context (2x).
	textMatch := newTestTask("1", "deploy kubernetes cluster")
	contextMatch := newTestTask("2", "some other text")
	contextMatch.Context = "deploy kubernetes cluster"

	pool := newTestPool(textMatch, contextMatch)
	engine := NewTaskQueryEngine(pool)

	results := engine.Search("deploy kubernetes", DefaultSearchOptions())
	if len(results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Task.ID != "1" {
		t.Errorf("text match should rank higher, got task %s first", results[0].Task.ID)
	}
}

func TestSearchNotesField(t *testing.T) {
	t.Parallel()

	task := newTestTask("1", "general task")
	task.Notes = []core.TaskNote{
		{Timestamp: time.Now().UTC(), Text: "needs kubernetes deployment"},
	}

	pool := newTestPool(task)
	engine := NewTaskQueryEngine(pool)

	results := engine.Search("kubernetes", DefaultSearchOptions())
	if len(results) == 0 {
		t.Fatal("expected results from notes field match")
	}

	found := false
	for _, m := range results[0].MatchedOn {
		if m == "notes" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'notes' in matched_on")
	}
}

func TestSearchMaxResults(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	for i := 0; i < 20; i++ {
		pool.AddTask(newTestTask(fmt.Sprintf("%d", i), "common search term here"))
	}

	engine := NewTaskQueryEngine(pool)
	opts := DefaultSearchOptions()
	opts.MaxResults = 5

	results := engine.Search("common search", opts)
	if len(results) > 5 {
		t.Errorf("expected max 5 results, got %d", len(results))
	}
}

func TestFilterTasks(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	tasks := []*core.Task{
		{ID: "1", Text: "todo task", Status: core.StatusTodo, Type: core.TypeCreative, Effort: core.EffortQuickWin, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now},
		{ID: "2", Text: "done task", Status: core.StatusComplete, Type: core.TypeTechnical, Effort: core.EffortDeepWork, CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now},
		{ID: "3", Text: "blocked task", Status: core.StatusBlocked, CreatedAt: now, UpdatedAt: now},
	}

	tests := []struct {
		name string
		opts FilterOptions
		want int
	}{
		{"no filter", FilterOptions{}, 3},
		{"by status", FilterOptions{Status: "todo"}, 1},
		{"by type", FilterOptions{Type: "creative"}, 1},
		{"by effort", FilterOptions{Effort: "quick-win"}, 1},
		{"text contains", FilterOptions{TextContains: "blocked"}, 1},
		{"limit", FilterOptions{Limit: 2}, 2},
		{"no match", FilterOptions{Status: "deferred"}, 0},
		{"sort desc", FilterOptions{SortBy: "created_at", SortOrder: "desc"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := FilterTasks(tasks, tt.opts)
			if len(result) != tt.want {
				t.Errorf("FilterTasks(%s) returned %d, want %d", tt.name, len(result), tt.want)
			}
		})
	}
}

func TestFilterTasksByCreatedAfter(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	tasks := []*core.Task{
		{ID: "1", Text: "old", Status: core.StatusTodo, CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now},
		{ID: "2", Text: "new", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	}

	result := FilterTasks(tasks, FilterOptions{
		CreatedAfter: now.Add(-1 * time.Hour).Format(time.RFC3339),
	})
	if len(result) != 1 || result[0].ID != "2" {
		t.Errorf("expected only the new task, got %d tasks", len(result))
	}
}

func TestFilterTasksSortByText(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	tasks := []*core.Task{
		{ID: "1", Text: "cherry", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
		{ID: "2", Text: "apple", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
		{ID: "3", Text: "banana", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now},
	}

	result := FilterTasks(tasks, FilterOptions{SortBy: "text"})
	if result[0].Text != "apple" || result[1].Text != "banana" || result[2].Text != "cherry" {
		t.Errorf("expected alphabetical sort, got %s, %s, %s", result[0].Text, result[1].Text, result[2].Text)
	}
}
