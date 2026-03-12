package oauth

import (
	"encoding/json"
	"fmt"
	"time"
)

// RefreshThreshold is the duration before token expiry at which a
// pre-emptive refresh should be triggered.
const RefreshThreshold = 5 * time.Minute

// TokenMetadata stores OAuth token information alongside the access token.
// It is serialized as JSON and stored in the credential store, replacing
// the plain access token string for OAuth connections.
type TokenMetadata struct {
	AccessToken   string    `json:"access_token"`
	RefreshToken  string    `json:"refresh_token,omitempty"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	TokenEndpoint string    `json:"token_endpoint,omitempty"`
	ClientID      string    `json:"client_id,omitempty"`
}

// NeedsRefresh returns true if the access token expires within the
// refresh threshold window (5 minutes by default). Returns false if
// no expiry is set (e.g., non-expiring tokens).
func (m *TokenMetadata) NeedsRefresh() bool {
	if m.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().UTC().Add(RefreshThreshold).After(m.ExpiresAt)
}

// IsExpired returns true if the access token has already expired.
func (m *TokenMetadata) IsExpired() bool {
	if m.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().UTC().After(m.ExpiresAt)
}

// CanRefresh returns true if the metadata contains enough information
// to attempt a token refresh (refresh token + token endpoint + client ID).
func (m *TokenMetadata) CanRefresh() bool {
	return m.RefreshToken != "" && m.TokenEndpoint != "" && m.ClientID != ""
}

// MarshalCredential serializes TokenMetadata to a JSON string for
// storage in the credential store.
func MarshalCredential(meta *TokenMetadata) (string, error) {
	data, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("marshal token metadata: %w", err)
	}
	return string(data), nil
}

// UnmarshalCredential deserializes a credential string into TokenMetadata.
// If the string is not valid JSON (i.e., a plain API key or token), it
// returns a TokenMetadata with just the AccessToken field populated.
func UnmarshalCredential(credential string) *TokenMetadata {
	if credential == "" {
		return &TokenMetadata{}
	}

	var meta TokenMetadata
	if err := json.Unmarshal([]byte(credential), &meta); err != nil {
		// Not JSON — treat as a plain access token (API key).
		return &TokenMetadata{AccessToken: credential}
	}

	// Guard against a JSON string that parsed but has no access token
	// (e.g., "{}"). If there's no access_token field but the raw string
	// looks like a token, use it as-is.
	if meta.AccessToken == "" && credential != "{}" {
		return &TokenMetadata{AccessToken: credential}
	}

	return &meta
}

// NewTokenMetadataFromResponse creates TokenMetadata from a device code
// flow TokenResponse, recording the token endpoint and client ID for
// future refresh operations.
func NewTokenMetadataFromResponse(resp *TokenResponse, tokenEndpoint, clientID string) *TokenMetadata {
	meta := &TokenMetadata{
		AccessToken:   resp.AccessToken,
		RefreshToken:  resp.RefreshToken,
		TokenEndpoint: tokenEndpoint,
		ClientID:      clientID,
	}

	if resp.ExpiresIn > 0 {
		meta.ExpiresAt = time.Now().UTC().Add(time.Duration(resp.ExpiresIn) * time.Second)
	}

	return meta
}
