package connection

import (
	"fmt"
	"sort"
	"time"
)

// WarningKind identifies the type of proactive health warning.
type WarningKind int

const (
	WarningTokenExpiry WarningKind = iota
	WarningRateLimit
	WarningErrorStreak
)

// tokenExpiryWindow is the look-ahead for token expiry warnings.
const tokenExpiryWindow = 7 * 24 * time.Hour

// rateLimitThreshold is the usage percentage that triggers a rate limit warning.
const rateLimitThreshold = 80

// errorStreakThreshold is the number of consecutive sync errors that triggers a warning.
const errorStreakThreshold = 3

// HealthWarning represents a proactive notification about a connection issue.
type HealthWarning struct {
	ConnectionID    string
	ConnectionLabel string
	Kind            WarningKind
	Message         string
}

// String returns the warning message.
func (w HealthWarning) String() string {
	return w.Message
}

// UpdateTokenExpiry sets the OAuth token expiry time for a connection.
func (m *ConnectionManager) UpdateTokenExpiry(id string, expiry time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.connections[id]
	if !ok {
		return fmt.Errorf("update token expiry %s: %w", id, ErrConnectionNotFound)
	}
	conn.TokenExpiry = expiry
	return nil
}

// UpdateRateLimit records the current rate limit state from API response headers.
func (m *ConnectionManager) UpdateRateLimit(id string, remaining, total int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.connections[id]
	if !ok {
		return fmt.Errorf("update rate limit %s: %w", id, ErrConnectionNotFound)
	}
	conn.RateLimitRemaining = remaining
	conn.RateLimitTotal = total
	return nil
}

// RecordSyncError increments the consecutive error counter for a connection.
func (m *ConnectionManager) RecordSyncError(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.connections[id]
	if !ok {
		return fmt.Errorf("record sync error %s: %w", id, ErrConnectionNotFound)
	}
	conn.ConsecutiveErrors++
	return nil
}

// RecordSyncSuccess resets the consecutive error counter for a connection.
func (m *ConnectionManager) RecordSyncSuccess(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.connections[id]
	if !ok {
		return fmt.Errorf("record sync success %s: %w", id, ErrConnectionNotFound)
	}
	conn.ConsecutiveErrors = 0
	return nil
}

// HealthWarnings returns proactive warnings for all connections.
// Warnings are sorted by priority: token expiry > rate limit > error streak,
// then by connection label within each kind.
func (m *ConnectionManager) HealthWarnings() []HealthWarning {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now().UTC()
	var warnings []HealthWarning

	for _, conn := range m.connections {
		warnings = append(warnings, checkConnectionHealth(conn, now)...)
	}

	sort.Slice(warnings, func(i, j int) bool {
		if warnings[i].Kind != warnings[j].Kind {
			return warnings[i].Kind < warnings[j].Kind
		}
		return warnings[i].ConnectionLabel < warnings[j].ConnectionLabel
	})

	return warnings
}

func checkConnectionHealth(conn *Connection, now time.Time) []HealthWarning {
	var warnings []HealthWarning

	// Token expiry check
	if !conn.TokenExpiry.IsZero() && conn.TokenExpiry.Before(now.Add(tokenExpiryWindow)) {
		daysLeft := int(time.Until(conn.TokenExpiry).Hours() / 24)
		var msg string
		if daysLeft <= 0 {
			msg = fmt.Sprintf("⚠ %s token expired — :sources to renew", conn.Label)
		} else if daysLeft == 1 {
			msg = fmt.Sprintf("⚠ %s token expires in 1 day — :sources to renew", conn.Label)
		} else {
			msg = fmt.Sprintf("⚠ %s token expires in %d days — :sources to renew", conn.Label, daysLeft)
		}
		warnings = append(warnings, HealthWarning{
			ConnectionID:    conn.ID,
			ConnectionLabel: conn.Label,
			Kind:            WarningTokenExpiry,
			Message:         msg,
		})
	}

	// Rate limit check
	if conn.RateLimitTotal > 0 {
		usedPercent := 100 - (conn.RateLimitRemaining * 100 / conn.RateLimitTotal)
		if usedPercent >= rateLimitThreshold {
			warnings = append(warnings, HealthWarning{
				ConnectionID:    conn.ID,
				ConnectionLabel: conn.Label,
				Kind:            WarningRateLimit,
				Message:         fmt.Sprintf("⚠ %s rate limit %d%% — sync may slow down", conn.Label, usedPercent),
			})
		}
	}

	// Error streak check
	if conn.ConsecutiveErrors >= errorStreakThreshold {
		warnings = append(warnings, HealthWarning{
			ConnectionID:    conn.ID,
			ConnectionLabel: conn.Label,
			Kind:            WarningErrorStreak,
			Message:         fmt.Sprintf("⚠ %s sync failing — :sources to investigate", conn.Label),
		})
	}

	return warnings
}
