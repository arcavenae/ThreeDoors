package core

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreakerStateTransitions(t *testing.T) {
	t.Parallel()

	errTest := errors.New("test error")

	tests := []struct {
		name       string
		setup      func(cb *CircuitBreaker, clock *fakeClock)
		wantState  CircuitState
		wantString string
	}{
		{
			name:       "initial state is closed",
			setup:      func(cb *CircuitBreaker, clock *fakeClock) {},
			wantState:  CircuitClosed,
			wantString: "closed",
		},
		{
			name: "stays closed under threshold",
			setup: func(cb *CircuitBreaker, clock *fakeClock) {
				for i := 0; i < 4; i++ {
					_ = cb.Execute(func() error { return errTest })
				}
			},
			wantState:  CircuitClosed,
			wantString: "closed",
		},
		{
			name: "opens after threshold failures",
			setup: func(cb *CircuitBreaker, clock *fakeClock) {
				for i := 0; i < 5; i++ {
					_ = cb.Execute(func() error { return errTest })
				}
			},
			wantState:  CircuitOpen,
			wantString: "open",
		},
		{
			name: "transitions to half-open after probe interval",
			setup: func(cb *CircuitBreaker, clock *fakeClock) {
				for i := 0; i < 5; i++ {
					_ = cb.Execute(func() error { return errTest })
				}
				clock.advance(31 * time.Second)
			},
			wantState:  CircuitHalfOpen,
			wantString: "half-open",
		},
		{
			name: "closes after successful probe",
			setup: func(cb *CircuitBreaker, clock *fakeClock) {
				for i := 0; i < 5; i++ {
					_ = cb.Execute(func() error { return errTest })
				}
				clock.advance(31 * time.Second)
				_ = cb.Execute(func() error { return nil })
			},
			wantState:  CircuitClosed,
			wantString: "closed",
		},
		{
			name: "re-opens after failed probe",
			setup: func(cb *CircuitBreaker, clock *fakeClock) {
				for i := 0; i < 5; i++ {
					_ = cb.Execute(func() error { return errTest })
				}
				clock.advance(31 * time.Second)
				_ = cb.Execute(func() error { return errTest })
			},
			wantState:  CircuitOpen,
			wantString: "open",
		},
		{
			name: "success resets failure count",
			setup: func(cb *CircuitBreaker, clock *fakeClock) {
				for i := 0; i < 4; i++ {
					_ = cb.Execute(func() error { return errTest })
				}
				_ = cb.Execute(func() error { return nil })
				for i := 0; i < 4; i++ {
					_ = cb.Execute(func() error { return errTest })
				}
			},
			wantState:  CircuitClosed,
			wantString: "closed",
		},
		{
			name: "old failures outside window are pruned",
			setup: func(cb *CircuitBreaker, clock *fakeClock) {
				for i := 0; i < 3; i++ {
					_ = cb.Execute(func() error { return errTest })
				}
				clock.advance(3 * time.Minute) // beyond 2m window
				for i := 0; i < 4; i++ {
					_ = cb.Execute(func() error { return errTest })
				}
			},
			wantState:  CircuitClosed,
			wantString: "closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clock := newFakeClock()
			cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
			cb.now = clock.now

			tt.setup(cb, clock)

			got := cb.State()
			if got != tt.wantState {
				t.Errorf("State() = %v, want %v", got, tt.wantState)
			}
			if got.String() != tt.wantString {
				t.Errorf("State().String() = %q, want %q", got.String(), tt.wantString)
			}
		})
	}
}

func TestCircuitBreakerExecuteOpenReturnsError(t *testing.T) {
	t.Parallel()

	clock := newFakeClock()
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	cb.now = clock.now

	errTest := errors.New("test error")
	for i := 0; i < 5; i++ {
		_ = cb.Execute(func() error { return errTest })
	}

	called := false
	err := cb.Execute(func() error {
		called = true
		return nil
	})

	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Execute() error = %v, want ErrCircuitOpen", err)
	}
	if called {
		t.Error("Execute() called fn when circuit is open")
	}
}

func TestCircuitBreakerProbeIntervalDoubles(t *testing.T) {
	t.Parallel()

	clock := newFakeClock()
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	cb.now = clock.now

	errTest := errors.New("test error")

	// Trip to open
	for i := 0; i < 5; i++ {
		_ = cb.Execute(func() error { return errTest })
	}

	// First probe after 30s — fail it
	clock.advance(31 * time.Second)
	_ = cb.Execute(func() error { return errTest })

	// Should need 60s now (doubled)
	clock.advance(31 * time.Second) // only 31s, not enough
	if cb.State() != CircuitOpen {
		t.Error("expected Open after 31s (probe interval should be 60s)")
	}

	clock.advance(30 * time.Second) // now 61s total
	if cb.State() != CircuitHalfOpen {
		t.Error("expected HalfOpen after 61s")
	}

	// Fail again — should need 120s
	_ = cb.Execute(func() error { return errTest })
	clock.advance(61 * time.Second)
	if cb.State() != CircuitOpen {
		t.Error("expected Open after 61s (probe interval should be 120s)")
	}

	clock.advance(60 * time.Second) // now 121s total
	if cb.State() != CircuitHalfOpen {
		t.Error("expected HalfOpen after 121s")
	}
}

func TestCircuitBreakerProbeIntervalCapsAtMax(t *testing.T) {
	t.Parallel()

	clock := newFakeClock()
	config := CircuitBreakerConfig{
		FailureThreshold: 1,
		FailureWindow:    time.Minute,
		ProbeInterval:    10 * time.Minute,
		MaxProbeInterval: 30 * time.Minute,
	}
	cb := NewCircuitBreaker(config)
	cb.now = clock.now

	errTest := errors.New("test error")

	// Trip
	_ = cb.Execute(func() error { return errTest })

	// Probe at 10m, fail → interval becomes 20m
	clock.advance(11 * time.Minute)
	_ = cb.Execute(func() error { return errTest })

	// Probe at 20m, fail → interval becomes 30m (capped at max)
	clock.advance(21 * time.Minute)
	_ = cb.Execute(func() error { return errTest })

	// Probe at 30m, fail → interval stays 30m (capped)
	clock.advance(31 * time.Minute)
	_ = cb.Execute(func() error { return errTest })

	// Verify still capped at 30m
	clock.advance(29 * time.Minute)
	if cb.State() != CircuitOpen {
		t.Error("expected Open — probe interval should be capped at 30m")
	}
	clock.advance(2 * time.Minute)
	if cb.State() != CircuitHalfOpen {
		t.Error("expected HalfOpen after 31m (capped at 30m)")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	t.Parallel()

	clock := newFakeClock()
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	cb.now = clock.now

	errTest := errors.New("test error")
	for i := 0; i < 5; i++ {
		_ = cb.Execute(func() error { return errTest })
	}

	if cb.State() != CircuitOpen {
		t.Fatal("expected Open before Reset")
	}

	cb.Reset()

	if cb.State() != CircuitClosed {
		t.Error("expected Closed after Reset")
	}

	// Verify probe interval was also reset
	for i := 0; i < 5; i++ {
		_ = cb.Execute(func() error { return errTest })
	}
	clock.advance(31 * time.Second) // original 30s interval should work
	if cb.State() != CircuitHalfOpen {
		t.Error("expected HalfOpen — probe interval should have been reset to default")
	}
}

func TestCircuitBreakerConcurrentAccess(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	errTest := errors.New("test error")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cb.Execute(func() error {
				if time.Now().UnixNano()%2 == 0 {
					return errTest
				}
				return nil
			})
			_ = cb.State()
		}()
	}
	wg.Wait()

	// Just verify no panic or deadlock
	_ = cb.State()
}

func TestCircuitBreakerExecutePassesThroughFunctionError(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	errSpecific := errors.New("specific error")

	err := cb.Execute(func() error { return errSpecific })
	if !errors.Is(err, errSpecific) {
		t.Errorf("Execute() error = %v, want %v", err, errSpecific)
	}
}

func TestCircuitBreakerExecuteReturnsNilOnSuccess(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

// fakeClock provides a deterministic clock for testing.
type fakeClock struct {
	mu      sync.Mutex
	current time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{
		current: time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC),
	}
}

func (fc *fakeClock) now() time.Time {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.current
}

func (fc *fakeClock) advance(d time.Duration) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.current = fc.current.Add(d)
}
