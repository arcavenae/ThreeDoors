package connection

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestConnectionState_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state ConnectionState
		want  string
	}{
		{"disconnected", StateDisconnected, "disconnected"},
		{"connecting", StateConnecting, "connecting"},
		{"connected", StateConnected, "connected"},
		{"syncing", StateSyncing, "syncing"},
		{"error", StateError, "error"},
		{"auth expired", StateAuthExpired, "auth_expired"},
		{"paused", StatePaused, "paused"},
		{"unknown", ConnectionState(99), "unknown(99)"},
		{"negative", ConnectionState(-1), "unknown(-1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("ConnectionState(%d).String() = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestValidateTransition_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from ConnectionState
		to   ConnectionState
	}{
		// Disconnected transitions
		{"disconnected to connecting", StateDisconnected, StateConnecting},

		// Connecting transitions
		{"connecting to connected", StateConnecting, StateConnected},
		{"connecting to error", StateConnecting, StateError},
		{"connecting to auth expired", StateConnecting, StateAuthExpired},

		// Connected transitions
		{"connected to syncing", StateConnected, StateSyncing},
		{"connected to error", StateConnected, StateError},
		{"connected to auth expired", StateConnected, StateAuthExpired},
		{"connected to paused", StateConnected, StatePaused},
		{"connected to disconnected", StateConnected, StateDisconnected},

		// Syncing transitions
		{"syncing to connected", StateSyncing, StateConnected},
		{"syncing to error", StateSyncing, StateError},

		// Error transitions
		{"error to connecting", StateError, StateConnecting},
		{"error to disconnected", StateError, StateDisconnected},

		// AuthExpired transitions
		{"auth expired to connecting", StateAuthExpired, StateConnecting},
		{"auth expired to disconnected", StateAuthExpired, StateDisconnected},

		// Paused transitions
		{"paused to connected", StatePaused, StateConnected},
		{"paused to disconnected", StatePaused, StateDisconnected},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateTransition(tt.from, tt.to); err != nil {
				t.Errorf("ValidateTransition(%s, %s) = %v, want nil", tt.from, tt.to, err)
			}
		})
	}
}

func TestValidateTransition_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from ConnectionState
		to   ConnectionState
	}{
		{"disconnected to syncing", StateDisconnected, StateSyncing},
		{"disconnected to connected", StateDisconnected, StateConnected},
		{"disconnected to error", StateDisconnected, StateError},
		{"disconnected to paused", StateDisconnected, StatePaused},
		{"syncing to disconnected", StateSyncing, StateDisconnected},
		{"syncing to paused", StateSyncing, StatePaused},
		{"syncing to connecting", StateSyncing, StateConnecting},
		{"error to connected", StateError, StateConnected},
		{"error to syncing", StateError, StateSyncing},
		{"paused to syncing", StatePaused, StateSyncing},
		{"paused to connecting", StatePaused, StateConnecting},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateTransition(tt.from, tt.to)
			if err == nil {
				t.Errorf("ValidateTransition(%s, %s) = nil, want error", tt.from, tt.to)
			}
			if !errors.Is(err, ErrInvalidTransition) {
				t.Errorf("ValidateTransition(%s, %s) error = %v, want ErrInvalidTransition", tt.from, tt.to, err)
			}
		})
	}
}

func TestValidateTransition_SameState(t *testing.T) {
	t.Parallel()

	states := []ConnectionState{
		StateDisconnected, StateConnecting, StateConnected,
		StateSyncing, StateError, StateAuthExpired, StatePaused,
	}
	for _, s := range states {
		t.Run(s.String()+" to self", func(t *testing.T) {
			t.Parallel()
			err := ValidateTransition(s, s)
			if err == nil {
				t.Errorf("ValidateTransition(%s, %s) = nil, want error (same-state transition)", s, s)
			}
		})
	}
}

func TestConnectionState_MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state ConnectionState
		want  string
	}{
		{"disconnected", StateDisconnected, `"disconnected"`},
		{"connected", StateConnected, `"connected"`},
		{"syncing", StateSyncing, `"syncing"`},
		{"error", StateError, `"error"`},
		{"auth_expired", StateAuthExpired, `"auth_expired"`},
		{"paused", StatePaused, `"paused"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.state)
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}
			if string(data) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", data, tt.want)
			}
		})
	}
}

func TestConnectionState_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    ConnectionState
		wantErr bool
	}{
		{"disconnected", `"disconnected"`, StateDisconnected, false},
		{"connected", `"connected"`, StateConnected, false},
		{"error", `"error"`, StateError, false},
		{"unknown", `"bogus"`, 0, true},
		{"invalid json", `123`, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got ConnectionState
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidTransitions_AllStatesHaveEntry(t *testing.T) {
	t.Parallel()

	allStates := []ConnectionState{
		StateDisconnected, StateConnecting, StateConnected,
		StateSyncing, StateError, StateAuthExpired, StatePaused,
	}
	for _, s := range allStates {
		if _, ok := validTransitions[s]; !ok {
			t.Errorf("state %s has no entry in validTransitions table", s)
		}
	}
}
