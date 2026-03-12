package connection

// HealthCheckResult contains the outcome of a connection health check.
type HealthCheckResult struct {
	APIReachable bool              `json:"api_reachable"`
	TokenValid   bool              `json:"token_valid"`
	RateLimitOK  bool              `json:"rate_limit_ok"`
	TaskCount    int               `json:"task_count"`
	Details      map[string]string `json:"details,omitempty"`
}

// Healthy returns true when the API is reachable, the token is valid,
// and rate limiting is not a problem.
func (r HealthCheckResult) Healthy() bool {
	return r.APIReachable && r.TokenValid && r.RateLimitOK
}

// HealthChecker performs a lightweight health check for a connection.
// Implementations are provider-specific (e.g., checking whether the API
// endpoint responds, the token is accepted, and rate limits are OK).
type HealthChecker interface {
	CheckHealth(conn *Connection, credential string) (HealthCheckResult, error)
}

// Syncer triggers an immediate sync cycle for a connection.
// Implementations are provider-specific.
type Syncer interface {
	Sync(conn *Connection, credential string) error
}
