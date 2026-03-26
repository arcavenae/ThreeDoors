package connection

import (
	"context"
	"sync"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// ConnSchedulerResult extends SchedulerResult with the connection ID.
type ConnSchedulerResult struct {
	ConnectionID string
	Provider     string
	Tasks        []*core.Task
	Err          error
	Duration     time.Duration
}

// ConnAwareScheduler wraps per-connection polling loops that respect
// connection state. When a connection is Paused, its sync cycle is skipped.
// Results are enriched with the connection ID for downstream routing.
type ConnAwareScheduler struct {
	mu      sync.Mutex
	manager *ConnectionManager
	bridge  *ProviderBridge
	loops   []connLoop
	results chan ConnSchedulerResult
	cancel  context.CancelFunc
	stopped bool
	wg      sync.WaitGroup
}

type connLoop struct {
	connID   string
	provider core.TaskProvider
	config   core.ProviderLoopConfig
}

// NewConnAwareScheduler creates a scheduler that checks connection state
// before each poll cycle.
func NewConnAwareScheduler(manager *ConnectionManager, bridge *ProviderBridge) *ConnAwareScheduler {
	return &ConnAwareScheduler{
		manager: manager,
		bridge:  bridge,
	}
}

// AddConnection registers a connection for scheduled polling.
func (s *ConnAwareScheduler) AddConnection(connID string, provider core.TaskProvider, config core.ProviderLoopConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loops = append(s.loops, connLoop{
		connID:   connID,
		provider: provider,
		config:   config,
	})
}

// Start launches per-connection goroutines and returns the results channel.
func (s *ConnAwareScheduler) Start(ctx context.Context) <-chan ConnSchedulerResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results = make(chan ConnSchedulerResult, len(s.loops)*2)
	s.stopped = false

	loopCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	for i := range s.loops {
		s.wg.Add(1)
		go s.runLoop(loopCtx, s.loops[i])
	}

	go func() {
		s.wg.Wait()
		close(s.results)
	}()

	return s.results
}

// Stop cancels all loops and waits for them to finish.
func (s *ConnAwareScheduler) Stop() {
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

func (s *ConnAwareScheduler) runLoop(ctx context.Context, loop connLoop) {
	defer s.wg.Done()

	interval := core.NewAdaptiveInterval(loop.config.MinInterval, loop.config.MaxInterval, loop.config.Jitter)
	watchCh := loop.provider.Watch()

	// Initial poll
	s.poll(ctx, loop, interval)

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
					watchCh = nil
					continue
				}
				s.poll(ctx, loop, interval)
			case <-timer.C:
				s.poll(ctx, loop, interval)
			}
		} else {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				s.poll(ctx, loop, interval)
			}
		}
	}
}

// poll checks connection state before executing LoadTasks.
// Paused connections are silently skipped.
func (s *ConnAwareScheduler) poll(ctx context.Context, loop connLoop, interval *core.AdaptiveInterval) {
	conn, err := s.manager.Get(loop.connID)
	if err != nil {
		return // connection removed, stop polling
	}

	// Skip paused or disconnected connections
	if conn.State == StatePaused || conn.State == StateDisconnected {
		return
	}

	start := time.Now().UTC()
	tasks, loadErr := loop.provider.LoadTasks()
	duration := time.Since(start)

	if loadErr != nil {
		interval.OnFailure()
	} else {
		interval.OnSuccess()
	}

	result := ConnSchedulerResult{
		ConnectionID: loop.connID,
		Provider:     loop.provider.Name(),
		Tasks:        tasks,
		Err:          loadErr,
		Duration:     duration,
	}

	select {
	case s.results <- result:
	case <-ctx.Done():
	}
}
