package connection

import "fmt"

// ConnectionState represents the lifecycle state of a data source connection.
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateSyncing
	StateError
	StateAuthExpired
	StatePaused
)

var stateNames = [...]string{
	StateDisconnected: "disconnected",
	StateConnecting:   "connecting",
	StateConnected:    "connected",
	StateSyncing:      "syncing",
	StateError:        "error",
	StateAuthExpired:  "auth_expired",
	StatePaused:       "paused",
}

// String returns the human-readable name of the connection state.
func (s ConnectionState) String() string {
	if int(s) >= 0 && int(s) < len(stateNames) {
		return stateNames[s]
	}
	return fmt.Sprintf("unknown(%d)", int(s))
}

// validTransitions defines which state transitions are allowed.
// Key: current state, Value: set of valid target states.
var validTransitions = map[ConnectionState][]ConnectionState{
	StateDisconnected: {StateConnecting},
	StateConnecting:   {StateConnected, StateError, StateAuthExpired},
	StateConnected:    {StateSyncing, StateError, StateAuthExpired, StatePaused, StateDisconnected},
	StateSyncing:      {StateConnected, StateError},
	StateError:        {StateConnecting, StateDisconnected},
	StateAuthExpired:  {StateConnecting, StateDisconnected},
	StatePaused:       {StateConnected, StateDisconnected},
}

// ErrInvalidTransition is returned when an invalid state transition is attempted.
var ErrInvalidTransition = fmt.Errorf("invalid state transition")

// ValidateTransition checks whether transitioning from one state to another is allowed.
func ValidateTransition(from, to ConnectionState) error {
	targets, ok := validTransitions[from]
	if !ok {
		return fmt.Errorf("%w: no transitions defined for state %s", ErrInvalidTransition, from)
	}
	for _, t := range targets {
		if t == to {
			return nil
		}
	}
	return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
}
