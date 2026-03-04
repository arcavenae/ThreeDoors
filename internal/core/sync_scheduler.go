package core

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"
)

// SchedulerResult carries the outcome of a single provider's sync poll.
// The scheduler fans results from all provider loops into a single channel.
type SchedulerResult struct {
	Provider string
	Tasks    []*Task
	Err      error
	Duration time.Duration
}

// ProviderLoopConfig holds per-provider polling configuration.
type ProviderLoopConfig struct {
	MinInterval time.Duration
	MaxInterval time.Duration
	Jitter      float64 // 0.0–1.0, applied as ±Jitter fraction
}

// DefaultFileProviderLoopConfig returns defaults for filesystem-backed providers
// (TextFile, Obsidian) where fsnotify is the primary trigger and polling is fallback.
func DefaultFileProviderLoopConfig() ProviderLoopConfig {
	return ProviderLoopConfig{
		MinInterval: 30 * time.Second,
		MaxInterval: 5 * time.Minute,
		Jitter:      0.2,
	}
}

// DefaultAPIProviderLoopConfig returns defaults for API-backed providers
// where polling is the only mechanism.
func DefaultAPIProviderLoopConfig() ProviderLoopConfig {
	return ProviderLoopConfig{
		MinInterval: 30 * time.Second,
		MaxInterval: 30 * time.Minute,
		Jitter:      0.2,
	}
}

// AdaptiveInterval manages exponential backoff with jitter for polling intervals.
// On success the interval resets to min; on failure it doubles (capped at max).
// Jitter prevents thundering herd: interval * (1 + jitter * (rand*2 - 1)).
type AdaptiveInterval struct {
	min     time.Duration
	max     time.Duration
	jitter  float64
	current time.Duration
	rng     *rand.Rand
}

// NewAdaptiveInterval creates an adaptive interval starting at min.
func NewAdaptiveInterval(min, max time.Duration, jitter float64) *AdaptiveInterval {
	return &AdaptiveInterval{
		min:     min,
		max:     max,
		jitter:  jitter,
		current: min,
		rng:     rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

// Current returns the current base interval (without jitter applied).
func (ai *AdaptiveInterval) Current() time.Duration {
	return ai.current
}

// Next returns the current interval with jitter applied.
func (ai *AdaptiveInterval) Next() time.Duration {
	if ai.jitter == 0 {
		return ai.current
	}
	// interval * (1 + jitter * (rand.Float64()*2 - 1))
	factor := 1.0 + ai.jitter*(ai.rng.Float64()*2-1)
	return time.Duration(float64(ai.current) * factor)
}

// OnSuccess resets the interval to the minimum.
func (ai *AdaptiveInterval) OnSuccess() {
	ai.current = ai.min
}

// OnFailure doubles the interval, capped at max.
func (ai *AdaptiveInterval) OnFailure() {
	ai.current *= 2
	if ai.current > ai.max {
		ai.current = ai.max
	}
}

// Reset restores the interval to its initial minimum value.
func (ai *AdaptiveInterval) Reset() {
	ai.current = ai.min
}

// providerLoop holds the state for a single provider's sync goroutine.
type providerLoop struct {
	provider TaskProvider
	config   ProviderLoopConfig
}

// SyncScheduler manages per-provider goroutines that independently poll for
// task changes. Results from all providers fan into a single channel.
type SyncScheduler struct {
	mu      sync.Mutex
	loops   []providerLoop
	results chan SchedulerResult
	cancel  context.CancelFunc
	stopped bool
	wg      sync.WaitGroup
}

// NewSyncScheduler creates a scheduler with no providers configured.
// Use AddProvider to register providers before calling Start.
func NewSyncScheduler() *SyncScheduler {
	return &SyncScheduler{}
}

// AddProvider registers a provider with its polling configuration.
// Must be called before Start.
func (s *SyncScheduler) AddProvider(provider TaskProvider, config ProviderLoopConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loops = append(s.loops, providerLoop{
		provider: provider,
		config:   config,
	})
}

// Start launches per-provider goroutines and returns the fan-in results channel.
// The channel is closed when Stop is called or the context is cancelled.
func (s *SyncScheduler) Start(ctx context.Context) <-chan SchedulerResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results = make(chan SchedulerResult, len(s.loops)*2)
	s.stopped = false

	loopCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	for i := range s.loops {
		s.wg.Add(1)
		go s.runLoop(loopCtx, s.loops[i])
	}

	// Close results channel when all loops finish
	go func() {
		s.wg.Wait()
		close(s.results)
	}()

	return s.results
}

// Stop cancels all provider loops and waits for them to finish.
// Idempotent — safe to call multiple times.
func (s *SyncScheduler) Stop() {
	s.mu.Lock()
	if s.stopped || s.cancel == nil {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	cancel := s.cancel
	s.mu.Unlock()

	cancel()
	s.wg.Wait()
}

// runLoop is the goroutine body for a single provider.
// It runs a hybrid watch+poll loop: watch events trigger immediate re-polls,
// while the adaptive timer provides fallback polling.
func (s *SyncScheduler) runLoop(ctx context.Context, loop providerLoop) {
	defer s.wg.Done()

	interval := NewAdaptiveInterval(loop.config.MinInterval, loop.config.MaxInterval, loop.config.Jitter)
	watchCh := loop.provider.Watch()

	// Do an initial poll immediately
	s.poll(ctx, loop.provider, interval)

	for {
		timer := time.NewTimer(interval.Next())

		if watchCh != nil {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case _, ok := <-watchCh:
				timer.Stop()
				if !ok {
					// Watch channel closed, fall back to poll-only
					watchCh = nil
					continue
				}
				s.poll(ctx, loop.provider, interval)
			case <-timer.C:
				s.poll(ctx, loop.provider, interval)
			}
		} else {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				s.poll(ctx, loop.provider, interval)
			}
		}
	}
}

// poll executes a single LoadTasks call and sends the result to the fan-in channel.
func (s *SyncScheduler) poll(ctx context.Context, provider TaskProvider, interval *AdaptiveInterval) {
	start := time.Now().UTC()
	tasks, err := provider.LoadTasks()
	duration := time.Since(start)

	if err != nil {
		interval.OnFailure()
	} else {
		interval.OnSuccess()
	}

	result := SchedulerResult{
		Provider: provider.Name(),
		Tasks:    tasks,
		Err:      err,
		Duration: duration,
	}

	// Send result, respecting cancellation
	select {
	case s.results <- result:
	case <-ctx.Done():
	}
}
