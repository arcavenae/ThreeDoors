package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// ErrRefreshTokenExpired is returned when the refresh token itself has expired
// or been revoked by the authorization server.
var ErrRefreshTokenExpired = fmt.Errorf("token refresh: refresh token expired or revoked")

// RefreshToken exchanges a refresh token for a new access token using the
// OAuth 2.0 token endpoint (RFC 6749 Section 6).
func (c *Client) RefreshToken(ctx context.Context, tokenEndpoint, clientID, refreshToken string) (*TokenResponse, error) {
	if tokenEndpoint == "" {
		return nil, fmt.Errorf("token refresh: token endpoint must not be empty")
	}
	if clientID == "" {
		return nil, fmt.Errorf("token refresh: client ID must not be empty")
	}
	if refreshToken == "" {
		return nil, fmt.Errorf("token refresh: refresh token must not be empty")
	}

	form := url.Values{
		"client_id":     {clientID},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("token refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.URL.RawQuery = form.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token refresh request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("token refresh response read: %w", err)
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		var errResp tokenErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			if errResp.Error == "invalid_grant" || errResp.Error == "expired_token" {
				return nil, ErrRefreshTokenExpired
			}
		}
		return nil, fmt.Errorf("token refresh: HTTP %d: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("token refresh response parse: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("token refresh: response missing access_token")
	}

	return &tokenResp, nil
}
