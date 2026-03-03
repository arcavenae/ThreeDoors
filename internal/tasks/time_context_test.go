package tasks

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"
)

func TestTimeContextScore_NilContext(t *testing.T) {
	t.Parallel()
	tasks := []*Task{{Effort: EffortQuickWin}}
	got := TimeContextScore(tasks, nil)
	if got != 0 {
		t.Errorf("TimeContextScore with nil context = %d, want 0", got)
	}
}

func TestTimeContextScore_NoCalendar(t *testing.T) {
	t.Parallel()
	tasks := []*Task{{Effort: EffortQuickWin}}
	ctx := &TimeContext{HasCalendar: false, AvailableTime: 10 * time.Minute}
	got := TimeContextScore(tasks, ctx)
	if got != 0 {
		t.Errorf("TimeContextScore with HasCalendar=false = %d, want 0", got)
	}
}

func TestTimeContextScore_EmptyTasks(t *testing.T) {
	t.Parallel()
	ctx := &TimeContext{HasCalendar: true, AvailableTime: 10 * time.Minute}
	got := TimeContextScore(nil, ctx)
	if got != 0 {
		t.Errorf("TimeContextScore with nil tasks = %d, want 0", got)
	}
}

func TestTimeContextScore_ShortBlock(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		effort TaskEffort
		want   int
	}{
		{"quick-win preferred", EffortQuickWin, 2},
		{"medium moderate", EffortMedium, 1},
		{"deep-work deprioritized", EffortDeepWork, 0},
		{"uncategorized neutral", "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &TimeContext{HasCalendar: true, AvailableTime: 15 * time.Minute}
			tasks := []*Task{{Effort: tt.effort}}
			got := TimeContextScore(tasks, ctx)
			if got != tt.want {
				t.Errorf("TimeContextScore(short block, %q) = %d, want %d", tt.effort, got, tt.want)
			}
		})
	}
}

func TestTimeContextScore_MediumBlock(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		effort TaskEffort
		want   int
	}{
		{"quick-win moderate", EffortQuickWin, 1},
		{"medium preferred", EffortMedium, 2},
		{"deep-work deprioritized", EffortDeepWork, 0},
		{"uncategorized neutral", "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &TimeContext{HasCalendar: true, AvailableTime: 60 * time.Minute}
			tasks := []*Task{{Effort: tt.effort}}
			got := TimeContextScore(tasks, ctx)
			if got != tt.want {
				t.Errorf("TimeContextScore(medium block, %q) = %d, want %d", tt.effort, got, tt.want)
			}
		})
	}
}

func TestTimeContextScore_LongBlock(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		effort TaskEffort
		want   int
	}{
		{"quick-win deprioritized", EffortQuickWin, 0},
		{"medium moderate", EffortMedium, 1},
		{"deep-work preferred", EffortDeepWork, 2},
		{"uncategorized neutral", "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &TimeContext{HasCalendar: true, AvailableTime: 120 * time.Minute}
			tasks := []*Task{{Effort: tt.effort}}
			got := TimeContextScore(tasks, ctx)
			if got != tt.want {
				t.Errorf("TimeContextScore(long block, %q) = %d, want %d", tt.effort, got, tt.want)
			}
		})
	}
}

func TestTimeContextScore_MultipleTasks(t *testing.T) {
	t.Parallel()
	ctx := &TimeContext{HasCalendar: true, AvailableTime: 20 * time.Minute}
	tasks := []*Task{
		{Effort: EffortQuickWin}, // +2
		{Effort: EffortMedium},   // +1
		{Effort: EffortDeepWork}, // +0
	}
	got := TimeContextScore(tasks, ctx)
	if got != 3 {
		t.Errorf("TimeContextScore(short block, mixed) = %d, want 3", got)
	}
}

func TestTimeContextScore_BoundaryValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		available time.Duration
		effort    TaskEffort
		want      int
	}{
		{"exactly 30min quick-win", 30 * time.Minute, EffortQuickWin, 2},
		{"exactly 30min medium", 30 * time.Minute, EffortMedium, 1},
		{"31min quick-win", 31 * time.Minute, EffortQuickWin, 1},
		{"31min medium", 31 * time.Minute, EffortMedium, 2},
		{"exactly 90min medium", 90 * time.Minute, EffortMedium, 2},
		{"exactly 90min deep-work", 90 * time.Minute, EffortDeepWork, 0},
		{"91min deep-work", 91 * time.Minute, EffortDeepWork, 2},
		{"91min medium", 91 * time.Minute, EffortMedium, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &TimeContext{HasCalendar: true, AvailableTime: tt.available}
			tasks := []*Task{{Effort: tt.effort}}
			got := TimeContextScore(tasks, ctx)
			if got != tt.want {
				t.Errorf("TimeContextScore(%v, %q) = %d, want %d", tt.available, tt.effort, got, tt.want)
			}
		})
	}
}

// --- SelectDoorsWithTimeContext Tests ---

func TestSelectDoorsWithTimeContext_NilContext_FallsBack(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	pool.AddTask(&Task{ID: "1", Text: "task1", Status: StatusTodo, Effort: EffortQuickWin})
	pool.AddTask(&Task{ID: "2", Text: "task2", Status: StatusTodo, Effort: EffortMedium})
	pool.AddTask(&Task{ID: "3", Text: "task3", Status: StatusTodo, Effort: EffortDeepWork})

	result := SelectDoorsWithTimeContext(pool, 3, nil)
	if len(result) != 3 {
		t.Fatalf("SelectDoorsWithTimeContext(nil ctx) returned %d tasks, want 3", len(result))
	}
}

func TestSelectDoorsWithTimeContext_NoCalendar_FallsBack(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	pool.AddTask(&Task{ID: "1", Text: "task1", Status: StatusTodo, Effort: EffortQuickWin})
	pool.AddTask(&Task{ID: "2", Text: "task2", Status: StatusTodo, Effort: EffortMedium})
	pool.AddTask(&Task{ID: "3", Text: "task3", Status: StatusTodo, Effort: EffortDeepWork})

	ctx := &TimeContext{HasCalendar: false}
	result := SelectDoorsWithTimeContext(pool, 3, ctx)
	if len(result) != 3 {
		t.Fatalf("SelectDoorsWithTimeContext(no calendar) returned %d tasks, want 3", len(result))
	}
}

func TestSelectDoorsWithTimeContext_EmptyPool(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	ctx := &TimeContext{HasCalendar: true, AvailableTime: 15 * time.Minute}
	result := SelectDoorsWithTimeContext(pool, 3, ctx)
	if result != nil {
		t.Errorf("SelectDoorsWithTimeContext(empty pool) = %v, want nil", result)
	}
}

func TestSelectDoorsWithTimeContext_FewerTasksThanCount(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	pool.AddTask(&Task{ID: "1", Text: "task1", Status: StatusTodo, Effort: EffortQuickWin})
	ctx := &TimeContext{HasCalendar: true, AvailableTime: 15 * time.Minute}
	result := SelectDoorsWithTimeContext(pool, 3, ctx)
	if len(result) != 1 {
		t.Fatalf("SelectDoorsWithTimeContext(1 task) returned %d tasks, want 1", len(result))
	}
}

func TestSelectDoorsWithTimeContext_ShortBlock_PrefersQuickWin(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	// Add many quick-win and deep-work tasks to give the algorithm plenty of choices
	for i := range 5 {
		pool.AddTask(&Task{
			ID: fmt.Sprintf("qw-%d", i), Text: fmt.Sprintf("quick task %d", i),
			Status: StatusTodo, Effort: EffortQuickWin,
			Type: TaskType([]string{"creative", "administrative", "technical", "physical", ""}[i%5]),
		})
	}
	for i := range 5 {
		pool.AddTask(&Task{
			ID: fmt.Sprintf("dw-%d", i), Text: fmt.Sprintf("deep task %d", i),
			Status: StatusTodo, Effort: EffortDeepWork,
			Type: TaskType([]string{"creative", "administrative", "technical", "physical", ""}[i%5]),
		})
	}

	ctx := &TimeContext{HasCalendar: true, AvailableTime: 15 * time.Minute}

	// Run multiple times and check that quick-win tasks appear more often
	quickWinCount := 0
	deepWorkCount := 0
	const iterations = 50
	for range iterations {
		// Reset recently-shown buffer
		p := NewTaskPool()
		for _, t := range pool.GetAllTasks() {
			p.AddTask(t)
		}
		result := SelectDoorsWithTimeContext(p, 3, ctx)
		for _, t := range result {
			switch t.Effort {
			case EffortQuickWin:
				quickWinCount++
			case EffortDeepWork:
				deepWorkCount++
			}
		}
	}

	if quickWinCount <= deepWorkCount {
		t.Errorf("Short block: quick-win count (%d) should exceed deep-work count (%d)", quickWinCount, deepWorkCount)
	}
}

func TestSelectDoorsWithTimeContext_LongBlock_PrefersDeepWork(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	for i := range 5 {
		pool.AddTask(&Task{
			ID: fmt.Sprintf("qw-%d", i), Text: fmt.Sprintf("quick task %d", i),
			Status: StatusTodo, Effort: EffortQuickWin,
			Type: TaskType([]string{"creative", "administrative", "technical", "physical", ""}[i%5]),
		})
	}
	for i := range 5 {
		pool.AddTask(&Task{
			ID: fmt.Sprintf("dw-%d", i), Text: fmt.Sprintf("deep task %d", i),
			Status: StatusTodo, Effort: EffortDeepWork,
			Type: TaskType([]string{"creative", "administrative", "technical", "physical", ""}[i%5]),
		})
	}

	ctx := &TimeContext{HasCalendar: true, AvailableTime: 120 * time.Minute}

	quickWinCount := 0
	deepWorkCount := 0
	const iterations = 50
	for range iterations {
		p := NewTaskPool()
		for _, t := range pool.GetAllTasks() {
			p.AddTask(t)
		}
		result := SelectDoorsWithTimeContext(p, 3, ctx)
		for _, t := range result {
			switch t.Effort {
			case EffortQuickWin:
				quickWinCount++
			case EffortDeepWork:
				deepWorkCount++
			}
		}
	}

	if deepWorkCount <= quickWinCount {
		t.Errorf("Long block: deep-work count (%d) should exceed quick-win count (%d)", deepWorkCount, quickWinCount)
	}
}

func TestSelectDoorsWithTimeContextAndRand_ShortBlock_HighTimeScore(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	pool.AddTask(&Task{ID: "1", Text: "t1", Status: StatusTodo, Effort: EffortQuickWin, Type: TypeCreative})
	pool.AddTask(&Task{ID: "2", Text: "t2", Status: StatusTodo, Effort: EffortMedium, Type: TypeAdministrative})
	pool.AddTask(&Task{ID: "3", Text: "t3", Status: StatusTodo, Effort: EffortDeepWork, Type: TypeTechnical})
	pool.AddTask(&Task{ID: "4", Text: "t4", Status: StatusTodo, Effort: EffortQuickWin, Type: TypePhysical})

	ctx := &TimeContext{HasCalendar: true, AvailableTime: 15 * time.Minute}
	rng := rand.New(rand.NewPCG(42, 0))
	result := selectDoorsWithTimeContextAndRand(pool, 3, ctx, rng)

	if len(result) != 3 {
		t.Fatalf("expected 3 doors, got %d", len(result))
	}

	// With a short block, quick-win tasks should score higher.
	// Combined score should be at least diversity (>=5) + time context (>=3).
	diversityScore := DiversityScore(result)
	timeScore := TimeContextScore(result, ctx)
	totalScore := diversityScore + timeScore

	if totalScore < 8 {
		t.Errorf("expected high combined score (>= 8), got diversity=%d + time=%d = %d",
			diversityScore, timeScore, totalScore)
	}
}

// --- FormatTimeContext Tests ---

func TestFormatTimeContext_NilContext(t *testing.T) {
	t.Parallel()
	got := FormatTimeContext(nil)
	if got != "" {
		t.Errorf("FormatTimeContext(nil) = %q, want empty", got)
	}
}

func TestFormatTimeContext_NoCalendar(t *testing.T) {
	t.Parallel()
	ctx := &TimeContext{HasCalendar: false}
	got := FormatTimeContext(ctx)
	if got != "" {
		t.Errorf("FormatTimeContext(no calendar) = %q, want empty", got)
	}
}

func TestFormatTimeContext_FarEvent(t *testing.T) {
	t.Parallel()
	ctx := &TimeContext{
		HasCalendar: true,
		NextEventIn: 5 * time.Hour,
	}
	got := FormatTimeContext(ctx)
	if got != "" {
		t.Errorf("FormatTimeContext(far event) = %q, want empty (>4hr hidden)", got)
	}
}

func TestFormatTimeContext_Formatting(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		nextIn    time.Duration
		eventName string
		want      string
	}{
		{"45 min with name", 45 * time.Minute, "Team Standup", "Next event in 45 min — Team Standup"},
		{"90 min with name", 90 * time.Minute, "Design Review", "Next event in 1h 30min — Design Review"},
		{"10 min no name", 10 * time.Minute, "", "Next event in 10 min"},
		{"2 hours with name", 2 * time.Hour, "All Hands", "Next event in 2h 0min — All Hands"},
		{"exactly 4 hours", 4 * time.Hour, "Meeting", "Next event in 4h 0min — Meeting"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &TimeContext{
				HasCalendar:   true,
				NextEventIn:   tt.nextIn,
				NextEventName: tt.eventName,
			}
			got := FormatTimeContext(ctx)
			if got != tt.want {
				t.Errorf("FormatTimeContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTimeContext_ZeroDuration(t *testing.T) {
	t.Parallel()
	ctx := &TimeContext{
		HasCalendar: true,
		NextEventIn: 0,
	}
	got := FormatTimeContext(ctx)
	if got != "" {
		t.Errorf("FormatTimeContext(zero duration) = %q, want empty", got)
	}
}
