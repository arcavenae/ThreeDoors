package core

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the health state of a circuit breaker.
type CircuitState int

const (
	// CircuitClosed means the provider is healthy; all requests are forwarded.
	CircuitClosed CircuitState = iota
	// CircuitOpen means the provider has tripped; requests are fast-failed.
	CircuitOpen
	// CircuitHalfOpen means the provider is being probed with a single request.
	CircuitHalfOpen
)

// String returns the human-readable name of the circuit state.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// ErrCircuitOpen is returned when Execute is called on an open circuit breaker.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreakerConfig holds tunable parameters for a circuit breaker.
type CircuitBreakerConfig struct {
	FailureThreshold int
	FailureWindow    time.Duration
	ProbeInterval    time.Duration
	MaxProbeInterval time.Duration
}

// DefaultCircuitBreakerConfig returns sensible defaults.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		FailureWindow:    2 * time.Minute,
		ProbeInterval:    30 * time.Second,
		MaxProbeInterval: 30 * time.Minute,
	}
}

// CircuitBreaker isolates a provider from cascading failures by tracking
// consecutive failures and transitioning through Closed → Open → Half-Open states.
type CircuitBreaker struct {
	mu            sync.Mutex
	state         CircuitState
	config        CircuitBreakerConfig
	failures      []time.Time // timestamps of recent failures
	probeInterval time.Duration
	lastTripped   time.Time
	now           func() time.Time // injectable clock for testing
}

// NewCircuitBreaker creates a circuit breaker with the given configuration.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state:         CircuitClosed,
		config:        config,
		failures:      make([]time.Time, 0, config.FailureThreshold),
		probeInterval: config.ProbeInterval,
		now:           func() time.Time { return time.Now().UTC() },
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.checkTransitions()
	return cb.state
}

// Execute runs the given function through the circuit breaker.
// In Closed state, the function is called and success/failure is tracked.
// In Open state, ErrCircuitOpen is returned without calling fn.
// In Half-Open state, the function is called as a probe.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	cb.checkTransitions()

	switch cb.state {
	case CircuitClosed:
		cb.mu.Unlock()
		err := fn()
		cb.mu.Lock()
		defer cb.mu.Unlock()
		if err != nil {
			cb.recordFailure()
		} else {
			cb.recordSuccess()
		}
		return err

	case CircuitOpen:
		cb.mu.Unlock()
		return ErrCircuitOpen

	case CircuitHalfOpen:
		cb.mu.Unlock()
		err := fn()
		cb.mu.Lock()
		defer cb.mu.Unlock()
		if err != nil {
			cb.tripOpen()
			cb.probeInterval *= 2
			if cb.probeInterval > cb.config.MaxProbeInterval {
				cb.probeInterval = cb.config.MaxProbeInterval
			}
		} else {
			cb.reset()
		}
		return err

	default:
		cb.mu.Unlock()
		return fmt.Errorf("circuit breaker: unknown state %d", cb.state)
	}
}

// Reset forces the circuit breaker back to Closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.reset()
}

// checkTransitions evaluates whether the circuit should transition states.
// Must be called with cb.mu held.
func (cb *CircuitBreaker) checkTransitions() {
	if cb.state == CircuitOpen {
		now := cb.now()
		if now.Sub(cb.lastTripped) >= cb.probeInterval {
			cb.state = CircuitHalfOpen
		}
	}
}

// recordFailure adds a failure timestamp and checks for threshold breach.
// Must be called with cb.mu held.
func (cb *CircuitBreaker) recordFailure() {
	now := cb.now()
	cb.failures = append(cb.failures, now)
	cb.pruneOldFailures(now)

	if len(cb.failures) >= cb.config.FailureThreshold {
		cb.tripOpen()
	}
}

// recordSuccess clears the failure history.
// Must be called with cb.mu held.
func (cb *CircuitBreaker) recordSuccess() {
	cb.failures = cb.failures[:0]
}

// tripOpen transitions to Open state.
// Must be called with cb.mu held.
func (cb *CircuitBreaker) tripOpen() {
	cb.state = CircuitOpen
	cb.lastTripped = cb.now()
	cb.failures = cb.failures[:0]
}

// reset transitions to Closed state with clean history.
// Must be called with cb.mu held.
func (cb *CircuitBreaker) reset() {
	cb.state = CircuitClosed
	cb.failures = cb.failures[:0]
	cb.probeInterval = cb.config.ProbeInterval
}

// pruneOldFailures removes failures outside the failure window.
// Must be called with cb.mu held.
func (cb *CircuitBreaker) pruneOldFailures(now time.Time) {
	cutoff := now.Add(-cb.config.FailureWindow)
	i := 0
	for i < len(cb.failures) && cb.failures[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		copy(cb.failures, cb.failures[i:])
		cb.failures = cb.failures[:len(cb.failures)-i]
	}
}
