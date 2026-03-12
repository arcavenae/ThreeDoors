package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/core/connection"
	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// benchTaskPool builds a task pool with n tasks using varied categories.
func benchTaskPool(n int) *core.TaskPool {
	pool := core.NewTaskPool()
	types := []core.TaskType{core.TypeCreative, core.TypeAdministrative, core.TypeTechnical, core.TypePhysical, ""}
	efforts := []core.TaskEffort{core.EffortQuickWin, core.EffortMedium, core.EffortDeepWork, ""}
	locations := []core.TaskLocation{core.LocationHome, core.LocationWork, core.LocationAnywhere, ""}

	baseTime := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	for i := range n {
		t := &core.Task{
			ID:        fmt.Sprintf("bench-%d", i),
			Text:      fmt.Sprintf("Benchmark task number %d with some detail", i),
			Status:    core.StatusTodo,
			Type:      types[i%len(types)],
			Effort:    efforts[i%len(efforts)],
			Location:  locations[i%len(locations)],
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		}
		if i%5 == 0 {
			t.SourceProvider = "github"
		}
		if i%7 == 0 {
			t.Context = "Some context about why this task matters"
		}
		pool.AddTask(t)
	}
	return pool
}

// benchSetASCII sets ASCII color profile for deterministic benchmark output.
func benchSetASCII(b *testing.B) {
	b.Helper()
	lipgloss.SetColorProfile(termenv.Ascii)
	b.Cleanup(func() {
		lipgloss.SetColorProfile(termenv.TrueColor)
	})
}

func BenchmarkDoorsView(b *testing.B) {
	benchSetASCII(b)

	pool := benchTaskPool(30)
	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	dv.SetWidth(120)
	dv.SetHeight(40)
	dv.SetThemeByName("classic")
	dv.SetShowKeyHints(true)
	dv.completedCount = 5

	// Populate avoidance data
	dv.avoidanceMap = map[string]int{
		"Benchmark task number 0 with some detail": 8,
		"Benchmark task number 5 with some detail": 6,
	}
	dv.avoidanceShown = map[string]int{
		"Benchmark task number 0 with some detail": 12,
		"Benchmark task number 5 with some detail": 9,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = dv.View()
	}
}

func BenchmarkDoorsViewSelected(b *testing.B) {
	benchSetASCII(b)

	pool := benchTaskPool(30)
	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	dv.SetWidth(120)
	dv.SetHeight(40)
	dv.SetThemeByName("classic")
	dv.SetShowKeyHints(true)
	dv.selectedDoorIndex = 1

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = dv.View()
	}
}

func BenchmarkInsightsView(b *testing.B) {
	benchSetASCII(b)

	// Create analyzer with sufficient session data via temp JSONL file
	analyzer := core.NewPatternAnalyzer()
	sessionsFile := writeBenchSessions(b, 10)
	if err := analyzer.LoadSessions(sessionsFile); err != nil {
		b.Fatalf("load sessions: %v", err)
	}

	counter := core.NewCompletionCounter()
	for range 25 {
		counter.IncrementToday()
	}

	theme, _ := themes.NewDefaultRegistry().Get("classic")
	iv := NewInsightsView(analyzer, counter, theme, nil)
	iv.SetWidth(120)
	iv.SetHeight(40)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		// Invalidate cache each iteration to measure actual render cost
		iv.invalidateCache()
		_ = iv.View()
	}
}

func BenchmarkInsightsViewCompact(b *testing.B) {
	benchSetASCII(b)

	analyzer := core.NewPatternAnalyzer()
	sessionsFile := writeBenchSessions(b, 10)
	if err := analyzer.LoadSessions(sessionsFile); err != nil {
		b.Fatalf("load sessions: %v", err)
	}

	counter := core.NewCompletionCounter()
	for range 15 {
		counter.IncrementToday()
	}

	theme, _ := themes.NewDefaultRegistry().Get("classic")
	iv := NewInsightsView(analyzer, counter, theme, nil)
	iv.SetWidth(50) // compact layout
	iv.SetHeight(24)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		iv.invalidateCache()
		_ = iv.View()
	}
}

func BenchmarkSourcesView(b *testing.B) {
	benchSetASCII(b)

	connMgr := connection.NewConnectionManager(nil)
	providers := []string{"github", "todoist", "linear", "jira", "notion"}
	for i, p := range providers {
		conn, err := connMgr.Add(p, fmt.Sprintf("My %s", p), map[string]string{"url": fmt.Sprintf("https://%s.example.com", p)})
		if err != nil {
			b.Fatalf("add connection %d: %v", i, err)
		}
		conn.TaskCount = (i + 1) * 10
		conn.LastSync = time.Now().UTC().Add(-time.Duration(i*15) * time.Minute)
	}

	sv := NewSourcesView(connMgr)
	sv.SetWidth(120)
	sv.SetHeight(40)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = sv.View()
	}
}

func BenchmarkDetailView(b *testing.B) {
	benchSetASCII(b)

	baseTime := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	task := &core.Task{
		ID:             "detail-bench-1",
		Text:           "A detailed benchmark task with comprehensive content",
		Status:         core.StatusInProgress,
		Type:           core.TypeTechnical,
		Effort:         core.EffortDeepWork,
		Location:       core.LocationWork,
		SourceProvider: "github",
		Context:        "This task has context explaining why it's important for benchmarking",
		Blocker:        "Waiting on upstream dependency",
		CreatedAt:      baseTime,
		UpdatedAt:      baseTime.Add(2 * time.Hour),
		Notes: []core.TaskNote{
			{Text: "First note about progress", Timestamp: baseTime.Add(30 * time.Minute)},
			{Text: "Second note with more detail", Timestamp: baseTime.Add(1 * time.Hour)},
			{Text: "Third note wrapping up", Timestamp: baseTime.Add(90 * time.Minute)},
		},
	}

	pool := benchTaskPool(20)
	dv := NewDetailView(task, nil, nil, pool)
	dv.width = 120

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = dv.View()
	}
}

// writeBenchSessions creates a temporary JSONL file with n session entries
// and returns the file path.
func writeBenchSessions(b *testing.B, n int) string {
	b.Helper()
	dir := b.TempDir()
	path := filepath.Join(dir, "sessions.jsonl")

	f, err := os.Create(path)
	if err != nil {
		b.Fatalf("create sessions file: %v", err)
	}
	b.Cleanup(func() {
		if cerr := f.Close(); cerr != nil {
			b.Logf("close sessions file: %v", cerr)
		}
	})

	moods := []string{"energized", "focused", "calm", "tired", "stressed"}
	baseTime := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)

	for i := range n {
		start := baseTime.Add(time.Duration(i*24) * time.Hour)
		sm := core.SessionMetrics{
			SessionID:           fmt.Sprintf("bench-session-%d", i),
			StartTime:           start,
			EndTime:             start.Add(25 * time.Minute),
			DurationSeconds:     1500,
			TasksCompleted:      3 + (i % 5),
			DoorsViewed:         6 + (i % 4),
			RefreshesUsed:       i % 3,
			DetailViews:         1 + (i % 3),
			NotesAdded:          i % 2,
			TimeToFirstDoorSecs: 5.0 + float64(i%10),
			DoorSelections: []core.DoorSelectionRecord{
				{DoorPosition: i % 3, TaskText: fmt.Sprintf("Task %d", i)},
				{DoorPosition: (i + 1) % 3, TaskText: fmt.Sprintf("Task %d", i+10)},
			},
			MoodEntries: []core.MoodEntry{
				{Mood: moods[i%len(moods)], Timestamp: start.Add(5 * time.Minute)},
			},
		}

		data, err := json.Marshal(sm)
		if err != nil {
			b.Fatalf("marshal session %d: %v", i, err)
		}
		if _, werr := fmt.Fprintf(f, "%s\n", data); werr != nil {
			b.Fatalf("write session %d: %v", i, werr)
		}
	}

	return path
}
