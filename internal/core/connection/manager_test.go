package connection

import (
	"errors"
	"sync"
	"testing"
)

func TestConnectionManager_Add(t *testing.T) {
	t.Parallel()

	t.Run("successful add", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, err := m.Add("jira", "Work Jira", map[string]string{"url": "https://jira.example.com"})
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
		if conn.ID == "" {
			t.Error("Add() returned connection with empty ID")
		}
		if m.Count() != 1 {
			t.Errorf("Count() = %d, want 1", m.Count())
		}
	})

	t.Run("add with empty provider", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		_, err := m.Add("", "label", nil)
		if err == nil {
			t.Error("Add() with empty provider should return error")
		}
	})

	t.Run("add multiple connections", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		_, err := m.Add("jira", "Work", nil)
		if err != nil {
			t.Fatalf("first Add() error = %v", err)
		}
		_, err = m.Add("github", "OSS", nil)
		if err != nil {
			t.Fatalf("second Add() error = %v", err)
		}
		if m.Count() != 2 {
			t.Errorf("Count() = %d, want 2", m.Count())
		}
	})
}

func TestConnectionManager_Remove(t *testing.T) {
	t.Parallel()

	t.Run("remove existing", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)
		if err := m.Remove(conn.ID); err != nil {
			t.Fatalf("Remove() error = %v", err)
		}
		if m.Count() != 0 {
			t.Errorf("Count() = %d, want 0", m.Count())
		}
	})

	t.Run("remove nonexistent", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		err := m.Remove("nonexistent")
		if err == nil {
			t.Error("Remove() nonexistent should return error")
		}
		if !errors.Is(err, ErrConnectionNotFound) {
			t.Errorf("Remove() error = %v, want ErrConnectionNotFound", err)
		}
	})
}

func TestConnectionManager_Get(t *testing.T) {
	t.Parallel()

	t.Run("get existing", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		added, _ := m.Add("jira", "Work", nil)
		got, err := m.Get(added.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.ID != added.ID {
			t.Errorf("Get() ID = %s, want %s", got.ID, added.ID)
		}
	})

	t.Run("get nonexistent", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		_, err := m.Get("nonexistent")
		if err == nil {
			t.Error("Get() nonexistent should return error")
		}
		if !errors.Is(err, ErrConnectionNotFound) {
			t.Errorf("Get() error = %v, want ErrConnectionNotFound", err)
		}
	})
}

func TestConnectionManager_List(t *testing.T) {
	t.Parallel()

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		list := m.List()
		if len(list) != 0 {
			t.Errorf("List() len = %d, want 0", len(list))
		}
	})

	t.Run("sorted by label", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		if _, err := m.Add("github", "Zulu OSS", nil); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
		if _, err := m.Add("jira", "Alpha Work", nil); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
		if _, err := m.Add("todoist", "Bravo Personal", nil); err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		list := m.List()
		if len(list) != 3 {
			t.Fatalf("List() len = %d, want 3", len(list))
		}
		if list[0].Label != "Alpha Work" {
			t.Errorf("List()[0].Label = %q, want %q", list[0].Label, "Alpha Work")
		}
		if list[1].Label != "Bravo Personal" {
			t.Errorf("List()[1].Label = %q, want %q", list[1].Label, "Bravo Personal")
		}
		if list[2].Label != "Zulu OSS" {
			t.Errorf("List()[2].Label = %q, want %q", list[2].Label, "Zulu OSS")
		}
	})

	t.Run("list includes state and metadata", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", map[string]string{"url": "https://jira.example.com"})

		list := m.List()
		if len(list) != 1 {
			t.Fatalf("List() len = %d, want 1", len(list))
		}
		if list[0].State != StateDisconnected {
			t.Errorf("State = %s, want %s", list[0].State, StateDisconnected)
		}
		if list[0].ID != conn.ID {
			t.Errorf("ID = %s, want %s", list[0].ID, conn.ID)
		}
	})
}

func TestConnectionManager_Transition(t *testing.T) {
	t.Parallel()

	t.Run("valid transition", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)

		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition to Connecting error = %v", err)
		}
		got, _ := m.Get(conn.ID)
		if got.State != StateConnecting {
			t.Errorf("State = %s, want %s", got.State, StateConnecting)
		}
	})

	t.Run("invalid transition", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)

		err := m.Transition(conn.ID, StateSyncing)
		if err == nil {
			t.Fatal("Transition Disconnected→Syncing should return error")
			return
		}
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("error = %v, want ErrInvalidTransition", err)
		}
		got, _ := m.Get(conn.ID)
		if got.State != StateDisconnected {
			t.Errorf("State changed to %s after invalid transition, want %s", got.State, StateDisconnected)
		}
	})

	t.Run("transition nonexistent", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		err := m.Transition("nonexistent", StateConnecting)
		if err == nil {
			t.Error("Transition on nonexistent should return error")
		}
		if !errors.Is(err, ErrConnectionNotFound) {
			t.Errorf("error = %v, want ErrConnectionNotFound", err)
		}
	})

	t.Run("full lifecycle", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)

		transitions := []ConnectionState{
			StateConnecting, StateConnected, StateSyncing, StateConnected,
			StatePaused, StateConnected, StateSyncing, StateError,
			StateConnecting, StateConnected, StateDisconnected,
		}
		for i, to := range transitions {
			if err := m.Transition(conn.ID, to); err != nil {
				t.Fatalf("step %d: Transition to %s error = %v", i, to, err)
			}
		}
	})

	t.Run("sync completion updates LastSync", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)

		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition to Connecting: %v", err)
		}
		if err := m.Transition(conn.ID, StateConnected); err != nil {
			t.Fatalf("Transition to Connected: %v", err)
		}
		if err := m.Transition(conn.ID, StateSyncing); err != nil {
			t.Fatalf("Transition to Syncing: %v", err)
		}

		got, _ := m.Get(conn.ID)
		if !got.LastSync.IsZero() {
			t.Error("LastSync should be zero before first sync completion")
		}

		if err := m.Transition(conn.ID, StateConnected); err != nil {
			t.Fatalf("Transition back to Connected: %v", err)
		}

		got, _ = m.Get(conn.ID)
		if got.LastSync.IsZero() {
			t.Error("LastSync should be set after Syncing→Connected transition")
		}
	})
}

func TestConnectionManager_TransitionWithError(t *testing.T) {
	t.Parallel()

	t.Run("records error message", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)
		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition to Connecting: %v", err)
		}

		err := m.TransitionWithError(conn.ID, StateError, "connection refused")
		if err != nil {
			t.Fatalf("TransitionWithError() error = %v", err)
		}
		got, _ := m.Get(conn.ID)
		if got.LastError != "connection refused" {
			t.Errorf("LastError = %q, want %q", got.LastError, "connection refused")
		}
	})
}

func TestConnectionManager_EventCallback(t *testing.T) {
	t.Parallel()

	t.Run("callback receives events", func(t *testing.T) {
		t.Parallel()
		var events []StateChangeEvent
		var mu sync.Mutex
		m := NewConnectionManager(func(e StateChangeEvent) {
			mu.Lock()
			events = append(events, e)
			mu.Unlock()
		})
		conn, _ := m.Add("jira", "Work", nil)

		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition to Connecting: %v", err)
		}
		if err := m.Transition(conn.ID, StateConnected); err != nil {
			t.Fatalf("Transition to Connected: %v", err)
		}

		mu.Lock()
		defer mu.Unlock()
		if len(events) != 2 {
			t.Fatalf("received %d events, want 2", len(events))
		}
		if events[0].From != StateDisconnected || events[0].To != StateConnecting {
			t.Errorf("event[0] = %s→%s, want %s→%s", events[0].From, events[0].To, StateDisconnected, StateConnecting)
		}
		if events[1].From != StateConnecting || events[1].To != StateConnected {
			t.Errorf("event[1] = %s→%s, want %s→%s", events[1].From, events[1].To, StateConnecting, StateConnected)
		}
		if events[0].ConnectionID != conn.ID {
			t.Errorf("event[0].ConnectionID = %s, want %s", events[0].ConnectionID, conn.ID)
		}
	})

	t.Run("no callback on invalid transition", func(t *testing.T) {
		t.Parallel()
		callCount := 0
		m := NewConnectionManager(func(_ StateChangeEvent) {
			callCount++
		})
		conn, _ := m.Add("jira", "Work", nil)

		_ = m.Transition(conn.ID, StateSyncing) // invalid, error expected
		if callCount != 0 {
			t.Errorf("callback called %d times on invalid transition, want 0", callCount)
		}
	})

	t.Run("nil callback is safe", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)
		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition with nil callback error = %v", err)
		}
	})

	t.Run("error event includes error message", func(t *testing.T) {
		t.Parallel()
		var captured StateChangeEvent
		m := NewConnectionManager(func(e StateChangeEvent) {
			captured = e
		})
		conn, _ := m.Add("jira", "Work", nil)
		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition to Connecting: %v", err)
		}

		if err := m.TransitionWithError(conn.ID, StateError, "timeout"); err != nil {
			t.Fatalf("TransitionWithError: %v", err)
		}
		if captured.Error != "timeout" {
			t.Errorf("event.Error = %q, want %q", captured.Error, "timeout")
		}
	})
}

func TestConnectionManager_NeedsAttention(t *testing.T) {
	t.Parallel()

	t.Run("empty when no connections", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		result := m.NeedsAttention()
		if len(result) != 0 {
			t.Errorf("NeedsAttention() len = %d, want 0", len(result))
		}
	})

	t.Run("empty when all healthy", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)
		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := m.Transition(conn.ID, StateConnected); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		result := m.NeedsAttention()
		if len(result) != 0 {
			t.Errorf("NeedsAttention() len = %d, want 0", len(result))
		}
	})

	t.Run("includes error state", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)
		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := m.TransitionWithError(conn.ID, StateError, "timeout"); err != nil {
			t.Fatalf("TransitionWithError: %v", err)
		}
		result := m.NeedsAttention()
		if len(result) != 1 {
			t.Fatalf("NeedsAttention() len = %d, want 1", len(result))
		}
		if result[0].State != StateError {
			t.Errorf("State = %s, want %s", result[0].State, StateError)
		}
	})

	t.Run("includes auth expired state", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)
		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := m.Transition(conn.ID, StateAuthExpired); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		result := m.NeedsAttention()
		if len(result) != 1 {
			t.Fatalf("NeedsAttention() len = %d, want 1", len(result))
		}
		if result[0].State != StateAuthExpired {
			t.Errorf("State = %s, want %s", result[0].State, StateAuthExpired)
		}
	})

	t.Run("excludes paused state", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)
		if err := m.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := m.Transition(conn.ID, StateConnected); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := m.Transition(conn.ID, StatePaused); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		result := m.NeedsAttention()
		if len(result) != 0 {
			t.Errorf("NeedsAttention() len = %d, want 0 (paused should be excluded)", len(result))
		}
	})

	t.Run("auth expired sorted before error", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)

		// Add error connection first
		errConn, _ := m.Add("github", "OSS", nil)
		if err := m.Transition(errConn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := m.TransitionWithError(errConn.ID, StateError, "timeout"); err != nil {
			t.Fatalf("TransitionWithError: %v", err)
		}

		// Add auth expired connection second
		authConn, _ := m.Add("jira", "Work", nil)
		if err := m.Transition(authConn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := m.Transition(authConn.ID, StateAuthExpired); err != nil {
			t.Fatalf("Transition: %v", err)
		}

		result := m.NeedsAttention()
		if len(result) != 2 {
			t.Fatalf("NeedsAttention() len = %d, want 2", len(result))
		}
		if result[0].State != StateAuthExpired {
			t.Errorf("result[0].State = %s, want AuthExpired (higher priority)", result[0].State)
		}
		if result[1].State != StateError {
			t.Errorf("result[1].State = %s, want Error", result[1].State)
		}
	})

	t.Run("same state sorted by label", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)

		connZ, _ := m.Add("github", "Zulu", nil)
		if err := m.Transition(connZ.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := m.TransitionWithError(connZ.ID, StateError, "fail"); err != nil {
			t.Fatalf("TransitionWithError: %v", err)
		}

		connA, _ := m.Add("jira", "Alpha", nil)
		if err := m.Transition(connA.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := m.TransitionWithError(connA.ID, StateError, "fail"); err != nil {
			t.Fatalf("TransitionWithError: %v", err)
		}

		result := m.NeedsAttention()
		if len(result) != 2 {
			t.Fatalf("NeedsAttention() len = %d, want 2", len(result))
		}
		if result[0].Label != "Alpha" {
			t.Errorf("result[0].Label = %q, want %q", result[0].Label, "Alpha")
		}
		if result[1].Label != "Zulu" {
			t.Errorf("result[1].Label = %q, want %q", result[1].Label, "Zulu")
		}
	})
}

func TestConnectionManager_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	m := NewConnectionManager(nil)
	var wg sync.WaitGroup

	// Concurrent adds
	for i := range 50 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _ = m.Add("provider", "Label"+string(rune('A'+n%26)), nil)
		}(i)
	}
	wg.Wait()

	if m.Count() != 50 {
		t.Errorf("Count() = %d, want 50 after concurrent adds", m.Count())
	}

	// Concurrent reads and transitions
	conns := m.List()
	for _, conn := range conns[:10] {
		wg.Add(3)
		go func(id string) {
			defer wg.Done()
			_, _ = m.Get(id)
		}(conn.ID)
		go func(id string) {
			defer wg.Done()
			m.List()
		}(conn.ID)
		go func(id string) {
			defer wg.Done()
			_ = m.Transition(id, StateConnecting)
		}(conn.ID)
	}
	wg.Wait()
}
