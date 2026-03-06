package core

import (
	"context"
	"errors"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockSyncProvider implements TaskProvider for scheduler tests.
type mockSyncProvider struct {
	name      string
	tasks     []*Task
	err       error
	loadCount atomic.Int64
	watchCh   chan ChangeEvent
}

func newMockSyncProvider(name string) *mockSyncProvider {
	return &mockSyncProvider{name: name}
}

func (m *mockSyncProvider) Name() string                   { return m.name }
func (m *mockSyncProvider) SaveTask(_ *Task) error         { return nil }
func (m *mockSyncProvider) SaveTasks(_ []*Task) error      { return nil }
func (m *mockSyncProvider) DeleteTask(_ string) error      { return nil }
func (m *mockSyncProvider) MarkComplete(_ string) error    { return nil }
func (m *mockSyncProvider) HealthCheck() HealthCheckResult { return HealthCheckResult{} }

func (m *mockSyncProvider) LoadTasks() ([]*Task, error) {
	m.loadCount.Add(1)
	return m.tasks, m.err
}

func (m *mockSyncProvider) Watch() <-chan ChangeEvent {
	return m.watchCh
}

// --- AdaptiveInterval Tests ---

func TestAdaptiveIntervalDefaults(t *testing.T) {
	t.Parallel()
	ai := NewAdaptiveInterval(30*time.Second, 5*time.Minute, 0.2)

	if ai.Current() != 30*time.Second {
		t.Errorf("initial interval = %v, want %v", ai.Current(), 30*time.Second)
	}
}

func TestAdaptiveIntervalOnSuccess(t *testing.T) {
	t.Parallel()
	ai := NewAdaptiveInterval(30*time.Second, 5*time.Minute, 0.0) // no jitter for determinism

	// Bump interval up via failures
	ai.OnFailure()
	ai.OnFailure()
	if ai.Current() != 2*time.Minute {
		t.Errorf("after 2 failures = %v, want %v", ai.Current(), 2*time.Minute)
	}

	// Success resets to min
	ai.OnSuccess()
	if ai.Current() != 30*time.Second {
		t.Errorf("after success = %v, want %v", ai.Current(), 30*time.Second)
	}
}

func TestAdaptiveIntervalOnFailure(t *testing.T) {
	t.Parallel()
	ai := NewAdaptiveInterval(30*time.Second, 5*time.Minute, 0.0)

	tests := []struct {
		name     string
		expected time.Duration
	}{
		{"first failure", 1 * time.Minute},
		{"second failure", 2 * time.Minute},
		{"third failure (capped)", 4 * time.Minute},
		{"fourth failure (capped at max)", 5 * time.Minute},
		{"fifth failure (stays at max)", 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai.OnFailure()
			if ai.Current() != tt.expected {
				t.Errorf("got %v, want %v", ai.Current(), tt.expected)
			}
		})
	}
}

func TestAdaptiveIntervalJitter(t *testing.T) {
	t.Parallel()
	ai := NewAdaptiveInterval(100*time.Second, 10*time.Minute, 0.2)

	// Collect multiple jittered values
	seen := make(map[time.Duration]bool)
	for i := 0; i < 50; i++ {
		d := ai.Next()
		seen[d] = true

		// Must be within ±20% of 100s → [80s, 120s]
		lo := time.Duration(float64(100*time.Second) * 0.8)
		hi := time.Duration(float64(100*time.Second) * 1.2)
		if d < lo || d > hi {
			t.Errorf("jittered interval %v outside [%v, %v]", d, lo, hi)
		}
	}

	// Should see some variation (not all same value)
	if len(seen) < 2 {
		t.Error("jitter produced no variation across 50 samples")
	}
}

// --- SchedulerResult Tests ---

func TestSchedulerResultFields(t *testing.T) {
	t.Parallel()
	task := NewTask("test task")
	result := SchedulerResult{
		Provider: "textfile",
		Tasks:    []*Task{task},
		Err:      nil,
		Duration: 50 * time.Millisecond,
	}
	if result.Provider != "textfile" {
		t.Errorf("provider = %q, want %q", result.Provider, "textfile")
	}
	if len(result.Tasks) != 1 {
		t.Errorf("tasks count = %d, want 1", len(result.Tasks))
	}
}

// --- SyncScheduler Lifecycle Tests ---

func TestSyncSchedulerStartStop(t *testing.T) {
	t.Parallel()
	p := newMockSyncProvider("test")

	cfg := ProviderLoopConfig{
		MinInterval: 50 * time.Millisecond,
		MaxInterval: 200 * time.Millisecond,
		Jitter:      0.0,
	}
	sched := NewSyncScheduler()
	sched.AddProvider(p, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := sched.Start(ctx)
	if results == nil {
		t.Fatal("Start() returned nil channel")
	}

	// Let it run a couple polls
	time.Sleep(120 * time.Millisecond)

	sched.Stop()

	// Should be able to drain results
	drained := 0
	for range results {
		drained++
	}
	if drained == 0 {
		t.Error("expected at least one result before stop")
	}
}

func TestSyncSchedulerStopIdempotent(t *testing.T) {
	t.Parallel()
	sched := NewSyncScheduler()
	p := newMockSyncProvider("test")
	sched.AddProvider(p, ProviderLoopConfig{
		MinInterval: 100 * time.Millisecond,
		MaxInterval: 1 * time.Second,
		Jitter:      0.0,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched.Start(ctx)

	// Multiple stops should not panic
	sched.Stop()
	sched.Stop()
	sched.Stop()
}

func TestSyncSchedulerNoProviders(t *testing.T) {
	t.Parallel()
	sched := NewSyncScheduler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := sched.Start(ctx)
	if results == nil {
		t.Fatal("Start() returned nil channel even with no providers")
	}

	sched.Stop()

	// Channel should close
	_, ok := <-results
	if ok {
		t.Error("expected results channel to be closed")
	}
}

// --- Fan-in Tests ---

func TestSyncSchedulerFanIn(t *testing.T) {
	t.Parallel()
	p1 := newMockSyncProvider("provider-a")
	p1.tasks = []*Task{NewTask("task a")}
	p2 := newMockSyncProvider("provider-b")
	p2.tasks = []*Task{NewTask("task b")}

	cfg := ProviderLoopConfig{
		MinInterval: 50 * time.Millisecond,
		MaxInterval: 200 * time.Millisecond,
		Jitter:      0.0,
	}

	sched := NewSyncScheduler()
	sched.AddProvider(p1, cfg)
	sched.AddProvider(p2, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := sched.Start(ctx)

	// Collect results from both providers
	seen := make(map[string]bool)
	timeout := time.After(2 * time.Second)
	for len(seen) < 2 {
		select {
		case r := <-results:
			seen[r.Provider] = true
		case <-timeout:
			t.Fatalf("timed out waiting for results from both providers, got: %v", seen)
		}
	}

	if !seen["provider-a"] || !seen["provider-b"] {
		t.Errorf("missing providers in results: %v", seen)
	}

	sched.Stop()
}

// --- Error Handling / Adaptive Backoff Tests ---

func TestSyncSchedulerErrorBackoff(t *testing.T) {
	t.Parallel()
	p := newMockSyncProvider("failing")
	p.err = errors.New("connection refused")

	cfg := ProviderLoopConfig{
		MinInterval: 50 * time.Millisecond,
		MaxInterval: 400 * time.Millisecond,
		Jitter:      0.0,
	}

	sched := NewSyncScheduler()
	sched.AddProvider(p, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := sched.Start(ctx)

	// Collect a few error results
	var errResults []SchedulerResult
	timeout := time.After(2 * time.Second)
	for len(errResults) < 3 {
		select {
		case r := <-results:
			if r.Err != nil {
				errResults = append(errResults, r)
			}
		case <-timeout:
			t.Fatalf("timed out, got only %d error results", len(errResults))
		}
	}

	sched.Stop()

	// All results should have errors
	for i, r := range errResults {
		if r.Err == nil {
			t.Errorf("result %d: expected error, got nil", i)
		}
		if r.Provider != "failing" {
			t.Errorf("result %d: provider = %q, want %q", i, r.Provider, "failing")
		}
	}
}

// --- Watch Channel Tests ---

func TestSyncSchedulerWatchTrigger(t *testing.T) {
	t.Parallel()
	watchCh := make(chan ChangeEvent, 1)
	p := newMockSyncProvider("watcher")
	p.watchCh = watchCh
	p.tasks = []*Task{NewTask("watched task")}

	cfg := ProviderLoopConfig{
		MinInterval: 10 * time.Second, // long poll interval
		MaxInterval: 30 * time.Second,
		Jitter:      0.0,
	}

	sched := NewSyncScheduler()
	sched.AddProvider(p, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := sched.Start(ctx)

	// Wait for initial poll result
	select {
	case <-results:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial poll")
	}

	initialLoads := p.loadCount.Load()

	// Trigger watch event — should cause immediate re-poll
	watchCh <- ChangeEvent{Type: ChangeUpdated, TaskID: "test", Source: "watcher"}

	// Should get result quickly (not waiting 10s for poll)
	select {
	case r := <-results:
		if r.Provider != "watcher" {
			t.Errorf("provider = %q, want %q", r.Provider, "watcher")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for watch-triggered result")
	}

	if p.loadCount.Load() <= initialLoads {
		t.Error("watch event did not trigger a LoadTasks call")
	}

	sched.Stop()
}

// --- Goroutine Leak Detection ---

func TestSyncSchedulerNoGoroutineLeaks(t *testing.T) {
	t.Parallel()

	// Stabilize goroutine count before measuring baseline.
	// Other parallel tests may be starting/stopping goroutines,
	// so we take the baseline as close to our work as possible.
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	p1 := newMockSyncProvider("leak-test-1")
	p2 := newMockSyncProvider("leak-test-2")
	p3 := newMockSyncProvider("leak-test-3")

	cfg := ProviderLoopConfig{
		MinInterval: 50 * time.Millisecond,
		MaxInterval: 200 * time.Millisecond,
		Jitter:      0.0,
	}

	sched := NewSyncScheduler()
	sched.AddProvider(p1, cfg)
	sched.AddProvider(p2, cfg)
	sched.AddProvider(p3, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := sched.Start(ctx)

	// Let it run
	time.Sleep(200 * time.Millisecond)

	sched.Stop()

	// Drain results channel
	for range results {
	}

	// Poll for goroutines to wind down — other parallel tests can cause
	// transient goroutine count spikes, so retry a few times.
	leaked := true
	for attempt := range 10 {
		runtime.GC()
		time.Sleep(100 * time.Duration(attempt+1) * time.Millisecond)

		after := runtime.NumGoroutine()
		// Allow slack of 5 for runtime/GC/test-framework goroutines
		// that may come and go independently of our scheduler.
		if after <= baseline+5 {
			leaked = false
			break
		}
	}

	if leaked {
		after := runtime.NumGoroutine()
		t.Errorf("goroutine leak: baseline=%d, after=%d (delta=%d)", baseline, after, after-baseline)
	}
}

// --- Context Cancellation Tests ---

func TestSyncSchedulerContextCancellation(t *testing.T) {
	t.Parallel()
	p := newMockSyncProvider("ctx-cancel")

	cfg := ProviderLoopConfig{
		MinInterval: 50 * time.Millisecond,
		MaxInterval: 200 * time.Millisecond,
		Jitter:      0.0,
	}

	sched := NewSyncScheduler()
	sched.AddProvider(p, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	results := sched.Start(ctx)

	// Cancel context — should stop scheduler
	cancel()

	// Results channel should eventually close
	timeout := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-results:
			if !ok {
				return // channel closed, test passes
			}
		case <-timeout:
			t.Fatal("results channel not closed after context cancellation")
		}
	}
}

// --- Concurrent Access Tests ---

func TestSyncSchedulerConcurrentResults(t *testing.T) {
	t.Parallel()

	const numProviders = 5
	providers := make([]*mockSyncProvider, numProviders)
	for i := range providers {
		providers[i] = newMockSyncProvider(
			"concurrent-" + string(rune('a'+i)),
		)
		providers[i].tasks = []*Task{NewTask("task")}
	}

	cfg := ProviderLoopConfig{
		MinInterval: 30 * time.Millisecond,
		MaxInterval: 100 * time.Millisecond,
		Jitter:      0.0,
	}

	sched := NewSyncScheduler()
	for _, p := range providers {
		sched.AddProvider(p, cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := sched.Start(ctx)

	// Concurrently read results from multiple goroutines
	var wg sync.WaitGroup
	var totalResults atomic.Int64
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range results {
				totalResults.Add(1)
			}
		}()
	}

	time.Sleep(300 * time.Millisecond)
	sched.Stop()
	wg.Wait()

	if totalResults.Load() == 0 {
		t.Error("no results received from concurrent readers")
	}
}

// --- AdaptiveInterval Math Tests ---

func TestAdaptiveIntervalBackoffMultiplier(t *testing.T) {
	t.Parallel()
	ai := NewAdaptiveInterval(1*time.Second, 1*time.Minute, 0.0)

	// Each failure doubles: 1s → 2s → 4s → 8s → 16s → 32s → 60s (capped)
	expected := []time.Duration{
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		32 * time.Second,
		60 * time.Second,
		60 * time.Second,
	}

	for i, want := range expected {
		ai.OnFailure()
		got := ai.Current()
		if got != want {
			t.Errorf("failure %d: got %v, want %v", i+1, got, want)
		}
	}
}

func TestAdaptiveIntervalJitterBounds(t *testing.T) {
	t.Parallel()
	jitter := 0.2
	base := 100 * time.Millisecond
	ai := NewAdaptiveInterval(base, 10*time.Second, jitter)

	lo := time.Duration(float64(base) * (1 - jitter))
	hi := time.Duration(float64(base) * (1 + jitter))

	for i := 0; i < 100; i++ {
		d := ai.Next()
		if d < lo || d > hi {
			t.Errorf("sample %d: %v outside [%v, %v]", i, d, lo, hi)
		}
	}
}

func TestAdaptiveIntervalZeroJitter(t *testing.T) {
	t.Parallel()
	ai := NewAdaptiveInterval(5*time.Second, 1*time.Minute, 0.0)

	// With zero jitter, Next() should always return Current() exactly
	for i := 0; i < 10; i++ {
		if ai.Next() != ai.Current() {
			t.Errorf("zero jitter: Next() != Current()")
		}
	}
}

// --- ProviderLoopConfig Defaults ---

func TestDefaultProviderLoopConfigs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config ProviderLoopConfig
	}{
		{"filesystem", DefaultFileProviderLoopConfig()},
		{"api", DefaultAPIProviderLoopConfig()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.config.MinInterval <= 0 {
				t.Error("MinInterval must be positive")
			}
			if tt.config.MaxInterval < tt.config.MinInterval {
				t.Error("MaxInterval must be >= MinInterval")
			}
			if tt.config.Jitter < 0 || tt.config.Jitter > 1 {
				t.Errorf("Jitter %v outside [0, 1]", tt.config.Jitter)
			}
		})
	}
}

// --- AdaptiveInterval with fake clock ---

func TestAdaptiveIntervalReset(t *testing.T) {
	t.Parallel()
	ai := NewAdaptiveInterval(1*time.Second, 1*time.Minute, 0.0)

	// Build up some backoff
	ai.OnFailure()
	ai.OnFailure()
	ai.OnFailure()
	if ai.Current() != 8*time.Second {
		t.Fatalf("expected 8s after 3 failures, got %v", ai.Current())
	}

	// Reset brings it back to min
	ai.Reset()
	if ai.Current() != 1*time.Second {
		t.Errorf("after reset: got %v, want %v", ai.Current(), 1*time.Second)
	}
}

// --- SyncScheduler with mixed watch/poll providers ---

func TestSyncSchedulerMixedWatchAndPoll(t *testing.T) {
	t.Parallel()

	// Provider with watch support
	watchCh := make(chan ChangeEvent, 1)
	watcher := newMockSyncProvider("with-watch")
	watcher.watchCh = watchCh
	watcher.tasks = []*Task{NewTask("watched")}

	// Provider without watch support (poll only)
	poller := newMockSyncProvider("poll-only")
	poller.tasks = []*Task{NewTask("polled")}

	cfg := ProviderLoopConfig{
		MinInterval: 50 * time.Millisecond,
		MaxInterval: 200 * time.Millisecond,
		Jitter:      0.0,
	}

	sched := NewSyncScheduler()
	sched.AddProvider(watcher, cfg)
	sched.AddProvider(poller, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := sched.Start(ctx)

	// Both should produce results
	seen := make(map[string]bool)
	timeout := time.After(2 * time.Second)
	for len(seen) < 2 {
		select {
		case r := <-results:
			seen[r.Provider] = true
		case <-timeout:
			t.Fatalf("timed out, seen: %v", seen)
		}
	}

	sched.Stop()
}

// Ensure AdaptiveInterval cap uses proper math
func TestAdaptiveIntervalCapMath(t *testing.T) {
	t.Parallel()
	ai := NewAdaptiveInterval(1*time.Second, 10*time.Second, 0.0)

	// 1 → 2 → 4 → 8 → 10(capped) → 10
	for i := 0; i < 10; i++ {
		ai.OnFailure()
	}

	got := ai.Current()
	if got != 10*time.Second {
		t.Errorf("capped interval = %v, want 10s", got)
	}
}

// Verify duration tracking in results
func TestSyncSchedulerResultDuration(t *testing.T) {
	t.Parallel()

	// Slow provider
	p := &slowMockProvider{
		mockSyncProvider: mockSyncProvider{
			name:  "slow",
			tasks: []*Task{NewTask("slow task")},
		},
		delay: 50 * time.Millisecond,
	}

	cfg := ProviderLoopConfig{
		MinInterval: 100 * time.Millisecond,
		MaxInterval: 1 * time.Second,
		Jitter:      0.0,
	}

	sched := NewSyncScheduler()
	sched.AddProvider(p, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := sched.Start(ctx)

	select {
	case r := <-results:
		if r.Duration < 50*time.Millisecond {
			t.Errorf("duration %v too short, expected >= 50ms", r.Duration)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for result")
	}

	sched.Stop()
}

// slowMockProvider adds artificial delay to LoadTasks.
type slowMockProvider struct {
	mockSyncProvider
	delay time.Duration
}

func (s *slowMockProvider) LoadTasks() ([]*Task, error) {
	time.Sleep(s.delay)
	return s.mockSyncProvider.LoadTasks()
}

// Verify jitter formula: interval * (1 + jitter * (rand.Float64()*2 - 1))
func TestAdaptiveIntervalJitterFormula(t *testing.T) {
	t.Parallel()
	base := 100 * time.Millisecond
	jitter := 0.2
	ai := NewAdaptiveInterval(base, 10*time.Second, jitter)

	var minSeen, maxSeen time.Duration
	minSeen = math.MaxInt64
	for i := 0; i < 1000; i++ {
		d := ai.Next()
		if d < minSeen {
			minSeen = d
		}
		if d > maxSeen {
			maxSeen = d
		}
	}

	expectedLo := time.Duration(float64(base) * (1 - jitter))
	expectedHi := time.Duration(float64(base) * (1 + jitter))

	if minSeen < expectedLo {
		t.Errorf("min seen %v < expected lower bound %v", minSeen, expectedLo)
	}
	if maxSeen > expectedHi {
		t.Errorf("max seen %v > expected upper bound %v", maxSeen, expectedHi)
	}
}
