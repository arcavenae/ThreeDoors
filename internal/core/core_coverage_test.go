package core

import (
	"testing"
)

// --- Aggregator Name/Watch/HealthCheck ---

func TestMultiSourceAggregator_Name(t *testing.T) {
	t.Parallel()
	a := &MultiSourceAggregator{}
	if a.Name() != "multi-source" {
		t.Errorf("Name() = %q, want %q", a.Name(), "multi-source")
	}
}

func TestMultiSourceAggregator_Watch(t *testing.T) {
	t.Parallel()
	a := &MultiSourceAggregator{}
	if a.Watch() != nil {
		t.Error("expected Watch to return nil")
	}
}

func TestMultiSourceAggregator_HealthCheck(t *testing.T) {
	t.Parallel()
	a := &MultiSourceAggregator{
		providers: map[string]TaskProvider{},
	}
	result := a.HealthCheck()
	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
}

// --- FallbackProvider Name/Watch/HealthCheck ---

type stubProvider struct {
	name   string
	tasks  []*Task
	err    error
	health HealthCheckResult
}

func (s *stubProvider) LoadTasks() ([]*Task, error)    { return s.tasks, s.err }
func (s *stubProvider) SaveTask(_ *Task) error         { return s.err }
func (s *stubProvider) SaveTasks(_ []*Task) error      { return s.err }
func (s *stubProvider) DeleteTask(_ string) error      { return s.err }
func (s *stubProvider) MarkComplete(_ string) error    { return s.err }
func (s *stubProvider) Name() string                   { return s.name }
func (s *stubProvider) Watch() <-chan ChangeEvent      { return nil }
func (s *stubProvider) HealthCheck() HealthCheckResult { return s.health }

func TestFallbackProvider_Name_Primary(t *testing.T) {
	t.Parallel()
	fp := NewFallbackProvider(
		&stubProvider{name: "primary"},
		&stubProvider{name: "fallback"},
	)
	if fp.Name() != "primary" {
		t.Errorf("Name() = %q, want %q", fp.Name(), "primary")
	}
}

func TestFallbackProvider_Name_Fallback(t *testing.T) {
	t.Parallel()
	fp := NewFallbackProvider(
		&stubProvider{name: "primary"},
		&stubProvider{name: "fallback"},
	)
	fp.usedFallback = true
	if fp.Name() != "fallback (fallback)" {
		t.Errorf("Name() = %q, want %q", fp.Name(), "fallback (fallback)")
	}
}

func TestFallbackProvider_Watch(t *testing.T) {
	t.Parallel()
	fp := NewFallbackProvider(
		&stubProvider{name: "primary"},
		&stubProvider{name: "fallback"},
	)
	if fp.Watch() != nil {
		t.Error("expected Watch to return nil")
	}
}

func TestFallbackProvider_Watch_Fallback(t *testing.T) {
	t.Parallel()
	fp := NewFallbackProvider(
		&stubProvider{name: "primary"},
		&stubProvider{name: "fallback"},
	)
	fp.usedFallback = true
	if fp.Watch() != nil {
		t.Error("expected Watch to return nil from fallback")
	}
}

func TestFallbackProvider_HealthCheck_Primary(t *testing.T) {
	t.Parallel()
	fp := NewFallbackProvider(
		&stubProvider{name: "primary", health: HealthCheckResult{Items: []HealthCheckItem{{Name: "test", Status: HealthOK}}}},
		&stubProvider{name: "fallback"},
	)
	result := fp.HealthCheck()
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Status != HealthOK {
		t.Error("expected HealthOK")
	}
}

func TestFallbackProvider_HealthCheck_Fallback(t *testing.T) {
	t.Parallel()
	fp := NewFallbackProvider(
		&stubProvider{name: "primary"},
		&stubProvider{name: "fallback", health: HealthCheckResult{Items: []HealthCheckItem{{Name: "fb", Status: HealthWarn}}}},
	)
	fp.usedFallback = true
	result := fp.HealthCheck()
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Status != HealthWarn {
		t.Error("expected HealthWarn from fallback")
	}
}

// --- SelectDoors ---

func TestSelectDoors(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	for i := 0; i < 10; i++ {
		pool.AddTask(NewTask("task"))
	}

	doors := SelectDoors(pool, 3)
	if len(doors) != 3 {
		t.Errorf("expected 3 doors, got %d", len(doors))
	}
}

func TestSelectDoors_Empty(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	doors := SelectDoors(pool, 3)
	if doors != nil {
		t.Errorf("expected nil for empty pool, got %v", doors)
	}
}

// --- SelectDoorsWithMood ---

func TestSelectDoorsWithMood_NoMood(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	for i := 0; i < 5; i++ {
		pool.AddTask(NewTask("task"))
	}

	doors := SelectDoorsWithMood(pool, 3, "", nil)
	if len(doors) != 3 {
		t.Errorf("expected 3 doors, got %d", len(doors))
	}
}

func TestSelectDoorsWithMood_WithMood(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	for i := 0; i < 10; i++ {
		task := NewTask("task")
		task.Type = TypeCreative
		pool.AddTask(task)
	}

	patterns := &PatternReport{
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", PreferredType: string(TypeCreative), PreferredEffort: string(EffortDeepWork)},
		},
	}

	doors := SelectDoorsWithMood(pool, 3, "Focused", patterns)
	if len(doors) != 3 {
		t.Errorf("expected 3 doors, got %d", len(doors))
	}
}

// --- GetBypassRate ---

func TestPatternAnalyzer_GetBypassRate(t *testing.T) {
	t.Parallel()
	pa := NewPatternAnalyzer()
	pa.sessions = []SessionMetrics{
		{DoorsViewed: 10, RefreshesUsed: 3},
		{DoorsViewed: 10, RefreshesUsed: 2},
	}

	rate := pa.GetBypassRate()
	if rate <= 0 {
		t.Errorf("expected positive bypass rate, got %f", rate)
	}
}

func TestPatternAnalyzer_GetBypassRate_NoViews(t *testing.T) {
	t.Parallel()
	pa := NewPatternAnalyzer()
	rate := pa.GetBypassRate()
	if rate != 0 {
		t.Errorf("expected 0 bypass rate for no sessions, got %f", rate)
	}
}
