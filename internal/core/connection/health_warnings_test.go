package connection

import (
	"testing"
	"time"
)

func TestHealthWarning_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		w    HealthWarning
		want string
	}{
		{
			name: "token expiry warning",
			w: HealthWarning{
				ConnectionLabel: "Todoist",
				Kind:            WarningTokenExpiry,
				Message:         "⚠ Todoist token expires in 5 days — :sources to renew",
			},
			want: "⚠ Todoist token expires in 5 days — :sources to renew",
		},
		{
			name: "rate limit warning",
			w: HealthWarning{
				ConnectionLabel: "GitHub",
				Kind:            WarningRateLimit,
				Message:         "⚠ GitHub rate limit 82% — sync may slow down",
			},
			want: "⚠ GitHub rate limit 82% — sync may slow down",
		},
		{
			name: "error streak warning",
			w: HealthWarning{
				ConnectionLabel: "Jira",
				Kind:            WarningErrorStreak,
				Message:         "⚠ Jira sync failing — :sources to investigate",
			},
			want: "⚠ Jira sync failing — :sources to investigate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.w.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConnectionManager_UpdateTokenExpiry(t *testing.T) {
	t.Parallel()

	t.Run("sets token expiry on connection", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("todoist", "Personal", nil)

		expiry := time.Now().UTC().Add(5 * 24 * time.Hour)
		if err := m.UpdateTokenExpiry(conn.ID, expiry); err != nil {
			t.Fatalf("UpdateTokenExpiry() error = %v", err)
		}

		got, _ := m.Get(conn.ID)
		if !got.TokenExpiry.Equal(expiry) {
			t.Errorf("TokenExpiry = %v, want %v", got.TokenExpiry, expiry)
		}
	})

	t.Run("nonexistent connection returns error", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		err := m.UpdateTokenExpiry("nonexistent", time.Now().UTC())
		if err == nil {
			t.Error("UpdateTokenExpiry() on nonexistent should return error")
		}
	})
}

func TestConnectionManager_UpdateRateLimit(t *testing.T) {
	t.Parallel()

	t.Run("sets rate limit fields", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("github", "OSS", nil)

		if err := m.UpdateRateLimit(conn.ID, 180, 1000); err != nil {
			t.Fatalf("UpdateRateLimit() error = %v", err)
		}

		got, _ := m.Get(conn.ID)
		if got.RateLimitRemaining != 180 {
			t.Errorf("RateLimitRemaining = %d, want 180", got.RateLimitRemaining)
		}
		if got.RateLimitTotal != 1000 {
			t.Errorf("RateLimitTotal = %d, want 1000", got.RateLimitTotal)
		}
	})

	t.Run("nonexistent connection returns error", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		err := m.UpdateRateLimit("nonexistent", 100, 1000)
		if err == nil {
			t.Error("UpdateRateLimit() on nonexistent should return error")
		}
	})
}

func TestConnectionManager_RecordSyncError(t *testing.T) {
	t.Parallel()

	t.Run("increments consecutive error count", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)

		if err := m.RecordSyncError(conn.ID); err != nil {
			t.Fatalf("RecordSyncError() error = %v", err)
		}
		got, _ := m.Get(conn.ID)
		if got.ConsecutiveErrors != 1 {
			t.Errorf("ConsecutiveErrors = %d, want 1", got.ConsecutiveErrors)
		}

		if err := m.RecordSyncError(conn.ID); err != nil {
			t.Fatalf("RecordSyncError() error = %v", err)
		}
		got, _ = m.Get(conn.ID)
		if got.ConsecutiveErrors != 2 {
			t.Errorf("ConsecutiveErrors = %d, want 2", got.ConsecutiveErrors)
		}
	})

	t.Run("nonexistent connection returns error", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		err := m.RecordSyncError("nonexistent")
		if err == nil {
			t.Error("RecordSyncError() on nonexistent should return error")
		}
	})
}

func TestConnectionManager_RecordSyncSuccess(t *testing.T) {
	t.Parallel()

	t.Run("resets consecutive error count", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)

		// Build up errors
		_ = m.RecordSyncError(conn.ID)
		_ = m.RecordSyncError(conn.ID)
		_ = m.RecordSyncError(conn.ID)

		if err := m.RecordSyncSuccess(conn.ID); err != nil {
			t.Fatalf("RecordSyncSuccess() error = %v", err)
		}
		got, _ := m.Get(conn.ID)
		if got.ConsecutiveErrors != 0 {
			t.Errorf("ConsecutiveErrors = %d, want 0 after success", got.ConsecutiveErrors)
		}
	})

	t.Run("nonexistent connection returns error", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		err := m.RecordSyncSuccess("nonexistent")
		if err == nil {
			t.Error("RecordSyncSuccess() on nonexistent should return error")
		}
	})
}

func TestConnectionManager_HealthWarnings(t *testing.T) {
	t.Parallel()

	t.Run("no warnings when all healthy", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		_, _ = m.Add("jira", "Work", nil)

		warnings := m.HealthWarnings()
		if len(warnings) != 0 {
			t.Errorf("HealthWarnings() len = %d, want 0", len(warnings))
		}
	})

	t.Run("token expiry within 7 days", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("todoist", "Personal", nil)

		expiry := time.Now().UTC().Add(5 * 24 * time.Hour)
		_ = m.UpdateTokenExpiry(conn.ID, expiry)

		warnings := m.HealthWarnings()
		if len(warnings) != 1 {
			t.Fatalf("HealthWarnings() len = %d, want 1", len(warnings))
		}
		if warnings[0].Kind != WarningTokenExpiry {
			t.Errorf("Kind = %v, want WarningTokenExpiry", warnings[0].Kind)
		}
		if warnings[0].ConnectionLabel != "Personal" {
			t.Errorf("ConnectionLabel = %q, want %q", warnings[0].ConnectionLabel, "Personal")
		}
	})

	t.Run("token expiry beyond 7 days is no warning", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("todoist", "Personal", nil)

		expiry := time.Now().UTC().Add(10 * 24 * time.Hour)
		_ = m.UpdateTokenExpiry(conn.ID, expiry)

		warnings := m.HealthWarnings()
		if len(warnings) != 0 {
			t.Errorf("HealthWarnings() len = %d, want 0", len(warnings))
		}
	})

	t.Run("expired token warns", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("todoist", "Personal", nil)

		expiry := time.Now().UTC().Add(-1 * time.Hour)
		_ = m.UpdateTokenExpiry(conn.ID, expiry)

		warnings := m.HealthWarnings()
		if len(warnings) != 1 {
			t.Fatalf("HealthWarnings() len = %d, want 1", len(warnings))
		}
		if warnings[0].Kind != WarningTokenExpiry {
			t.Errorf("Kind = %v, want WarningTokenExpiry", warnings[0].Kind)
		}
	})

	t.Run("rate limit above 80 percent", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("github", "OSS", nil)

		_ = m.UpdateRateLimit(conn.ID, 180, 1000) // 82% used

		warnings := m.HealthWarnings()
		if len(warnings) != 1 {
			t.Fatalf("HealthWarnings() len = %d, want 1", len(warnings))
		}
		if warnings[0].Kind != WarningRateLimit {
			t.Errorf("Kind = %v, want WarningRateLimit", warnings[0].Kind)
		}
	})

	t.Run("rate limit at 80 percent triggers warning", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("github", "OSS", nil)

		_ = m.UpdateRateLimit(conn.ID, 200, 1000) // exactly 80% used

		warnings := m.HealthWarnings()
		if len(warnings) != 1 {
			t.Fatalf("HealthWarnings() len = %d, want 1 (80%% should trigger)", len(warnings))
		}
	})

	t.Run("rate limit below 80 percent is no warning", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("github", "OSS", nil)

		_ = m.UpdateRateLimit(conn.ID, 500, 1000) // 50% used

		warnings := m.HealthWarnings()
		if len(warnings) != 0 {
			t.Errorf("HealthWarnings() len = %d, want 0", len(warnings))
		}
	})

	t.Run("rate limit with zero total is no warning", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("github", "OSS", nil)

		_ = m.UpdateRateLimit(conn.ID, 0, 0) // not tracked

		warnings := m.HealthWarnings()
		if len(warnings) != 0 {
			t.Errorf("HealthWarnings() len = %d, want 0", len(warnings))
		}
	})

	t.Run("3 consecutive errors warns", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)

		_ = m.RecordSyncError(conn.ID)
		_ = m.RecordSyncError(conn.ID)
		_ = m.RecordSyncError(conn.ID)

		warnings := m.HealthWarnings()
		if len(warnings) != 1 {
			t.Fatalf("HealthWarnings() len = %d, want 1", len(warnings))
		}
		if warnings[0].Kind != WarningErrorStreak {
			t.Errorf("Kind = %v, want WarningErrorStreak", warnings[0].Kind)
		}
	})

	t.Run("2 consecutive errors is no warning", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("jira", "Work", nil)

		_ = m.RecordSyncError(conn.ID)
		_ = m.RecordSyncError(conn.ID)

		warnings := m.HealthWarnings()
		if len(warnings) != 0 {
			t.Errorf("HealthWarnings() len = %d, want 0", len(warnings))
		}
	})

	t.Run("multiple warnings from different connections", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)

		// Token expiring
		c1, _ := m.Add("todoist", "Alpha", nil)
		_ = m.UpdateTokenExpiry(c1.ID, time.Now().UTC().Add(3*24*time.Hour))

		// Rate limit high
		c2, _ := m.Add("github", "Bravo", nil)
		_ = m.UpdateRateLimit(c2.ID, 100, 1000) // 90% used

		// Error streak
		c3, _ := m.Add("jira", "Charlie", nil)
		_ = m.RecordSyncError(c3.ID)
		_ = m.RecordSyncError(c3.ID)
		_ = m.RecordSyncError(c3.ID)

		warnings := m.HealthWarnings()
		if len(warnings) != 3 {
			t.Fatalf("HealthWarnings() len = %d, want 3", len(warnings))
		}
	})

	t.Run("multiple warnings from same connection", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		conn, _ := m.Add("github", "OSS", nil)

		_ = m.UpdateTokenExpiry(conn.ID, time.Now().UTC().Add(2*24*time.Hour))
		_ = m.UpdateRateLimit(conn.ID, 100, 1000) // 90% used

		warnings := m.HealthWarnings()
		if len(warnings) != 2 {
			t.Fatalf("HealthWarnings() len = %d, want 2", len(warnings))
		}
	})

	t.Run("warnings sorted by priority: token > rate limit > error streak", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)

		// Error streak (lowest priority)
		c1, _ := m.Add("jira", "Work", nil)
		_ = m.RecordSyncError(c1.ID)
		_ = m.RecordSyncError(c1.ID)
		_ = m.RecordSyncError(c1.ID)

		// Rate limit (medium priority)
		c2, _ := m.Add("github", "OSS", nil)
		_ = m.UpdateRateLimit(c2.ID, 100, 1000)

		// Token expiry (highest priority)
		c3, _ := m.Add("todoist", "Personal", nil)
		_ = m.UpdateTokenExpiry(c3.ID, time.Now().UTC().Add(2*24*time.Hour))

		warnings := m.HealthWarnings()
		if len(warnings) != 3 {
			t.Fatalf("HealthWarnings() len = %d, want 3", len(warnings))
		}
		if warnings[0].Kind != WarningTokenExpiry {
			t.Errorf("warnings[0].Kind = %v, want WarningTokenExpiry", warnings[0].Kind)
		}
		if warnings[1].Kind != WarningRateLimit {
			t.Errorf("warnings[1].Kind = %v, want WarningRateLimit", warnings[1].Kind)
		}
		if warnings[2].Kind != WarningErrorStreak {
			t.Errorf("warnings[2].Kind = %v, want WarningErrorStreak", warnings[2].Kind)
		}
	})

	t.Run("zero token expiry is ignored", func(t *testing.T) {
		t.Parallel()
		m := NewConnectionManager(nil)
		_, _ = m.Add("github", "OSS", nil)

		// TokenExpiry is zero value — should not trigger warning
		warnings := m.HealthWarnings()
		if len(warnings) != 0 {
			t.Errorf("HealthWarnings() len = %d, want 0 for zero expiry", len(warnings))
		}
	})
}
