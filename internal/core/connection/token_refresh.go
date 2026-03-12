package connection

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/arcaven/ThreeDoors/internal/core/connection/oauth"
)

// ErrUnauthorized is returned when an API responds with 401, indicating
// the credential (API key or access token) is no longer valid.
var ErrUnauthorized = fmt.Errorf("unauthorized: credential rejected by remote API")

// TokenRefresher handles pre-emptive OAuth token refresh before sync cycles.
// For OAuth connections, it checks token expiry and silently refreshes when
// the access token is within 5 minutes of expiry. For API key connections,
// it detects 401 responses and transitions to AuthExpired.
type TokenRefresher struct {
	creds    CredentialStore
	manager  *ConnectionManager
	eventLog *SyncEventLog
	oauth    *oauth.Client
}

// NewTokenRefresher creates a TokenRefresher with the given dependencies.
func NewTokenRefresher(creds CredentialStore, manager *ConnectionManager, eventLog *SyncEventLog, httpClient *http.Client) *TokenRefresher {
	return &TokenRefresher{
		creds:    creds,
		manager:  manager,
		eventLog: eventLog,
		oauth:    oauth.NewClient(httpClient),
	}
}

// RefreshResult describes the outcome of a refresh attempt.
type RefreshResult struct {
	Refreshed   bool   // true if token was refreshed
	AccessToken string // the current (possibly refreshed) access token
}

// EnsureValidToken checks whether the credential for a connection needs
// refreshing and performs the refresh if possible. Returns the current
// access token (refreshed or existing).
//
// For OAuth connections (TokenMetadata with refresh_token), it:
//  1. Checks if the access token is within 5 minutes of expiry
//  2. If so, attempts silent refresh using the refresh token
//  3. On success, stores new tokens in the credential store and logs the event
//  4. On failure (refresh token expired/revoked), transitions to AuthExpired
//
// For API key connections (plain string credential), it returns the token as-is.
func (r *TokenRefresher) EnsureValidToken(ctx context.Context, conn *Connection) (RefreshResult, error) {
	credKey := ConnCredentialKey(conn)
	rawCred, err := r.creds.Get(credKey)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("ensure valid token %s: %w", conn.ID, err)
	}

	meta := oauth.UnmarshalCredential(rawCred)

	// Plain API key or non-expiring token — return as-is.
	if !meta.CanRefresh() {
		return RefreshResult{AccessToken: meta.AccessToken}, nil
	}

	// Token still valid — no refresh needed.
	if !meta.NeedsRefresh() {
		return RefreshResult{AccessToken: meta.AccessToken}, nil
	}

	// Attempt refresh.
	tokenResp, refreshErr := r.oauth.RefreshToken(
		ctx,
		meta.TokenEndpoint,
		meta.ClientID,
		meta.RefreshToken,
	)
	if refreshErr != nil {
		if errors.Is(refreshErr, oauth.ErrRefreshTokenExpired) {
			r.transitionToAuthExpired(conn, "refresh token expired or revoked")
			return RefreshResult{}, refreshErr
		}
		return RefreshResult{}, fmt.Errorf("ensure valid token %s: %w", conn.ID, refreshErr)
	}

	// Update metadata with new tokens.
	newMeta := oauth.NewTokenMetadataFromResponse(tokenResp, meta.TokenEndpoint, meta.ClientID)
	// Preserve the refresh token if the server didn't issue a new one.
	if newMeta.RefreshToken == "" {
		newMeta.RefreshToken = meta.RefreshToken
	}

	serialized, err := oauth.MarshalCredential(newMeta)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("ensure valid token %s: %w", conn.ID, err)
	}

	if err := r.creds.Set(credKey, serialized); err != nil {
		return RefreshResult{}, fmt.Errorf("ensure valid token store %s: %w", conn.ID, err)
	}

	if r.eventLog != nil {
		_ = r.eventLog.LogTokenRefreshed(conn.ID)
	}

	return RefreshResult{Refreshed: true, AccessToken: newMeta.AccessToken}, nil
}

// HandleUnauthorized should be called when a sync or API call returns 401.
// It transitions the connection to AuthExpired and logs the event.
func (r *TokenRefresher) HandleUnauthorized(conn *Connection) {
	r.transitionToAuthExpired(conn, "API returned 401 Unauthorized")
}

func (r *TokenRefresher) transitionToAuthExpired(conn *Connection, reason string) {
	_ = r.manager.TransitionWithError(conn.ID, StateAuthExpired, reason)
	if r.eventLog != nil {
		_ = r.eventLog.Append(SyncEvent{
			ConnectionID: conn.ID,
			Type:         EventReauthRequired,
			Error:        reason,
			Summary:      fmt.Sprintf("Re-authentication required: %s", reason),
		})
	}
}
