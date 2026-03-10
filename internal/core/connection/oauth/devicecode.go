package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DeviceCodeConfig holds the OAuth endpoints and client credentials
// needed to initiate a device code flow (RFC 8628).
type DeviceCodeConfig struct {
	AuthEndpoint  string // e.g. "https://github.com/login/device/code"
	TokenEndpoint string // e.g. "https://github.com/login/oauth/access_token"
	ClientID      string
	Scopes        []string
}

// DeviceCodeResponse is returned by the authorization server when a device
// code is successfully requested.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"` // seconds until device_code expires
	Interval        int    `json:"interval"`   // minimum polling interval in seconds
}

// TokenResponse holds the tokens returned after successful authorization.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// ErrTimeout is returned when the device code expires before the user authorizes.
var ErrTimeout = fmt.Errorf("device code flow: authorization timed out")

// ErrAccessDenied is returned when the user explicitly denies authorization.
var ErrAccessDenied = fmt.Errorf("device code flow: access denied by user")

// ErrExpiredToken is returned when the device code has expired on the server.
var ErrExpiredToken = fmt.Errorf("device code flow: device code expired")

// Client performs OAuth device code flow operations.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a Client with the given HTTP client.
// If httpClient is nil, a default client with a 30-second timeout is used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{httpClient: httpClient}
}

const (
	maxResponseBytes = 1 << 20 // 1 MiB
	slowDownBackoff  = 5       // seconds per RFC 8628 Section 3.5
)

// StartDeviceCodeFlow requests a device code from the authorization server.
func (c *Client) StartDeviceCodeFlow(ctx context.Context, config DeviceCodeConfig) (*DeviceCodeResponse, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	form := url.Values{
		"client_id": {config.ClientID},
	}
	if len(config.Scopes) > 0 {
		form.Set("scope", strings.Join(config.Scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.AuthEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("device code request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device code request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("device code response read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var dcResp DeviceCodeResponse
	if err := json.Unmarshal(body, &dcResp); err != nil {
		return nil, fmt.Errorf("device code response parse: %w", err)
	}

	if dcResp.DeviceCode == "" || dcResp.UserCode == "" || dcResp.VerificationURI == "" {
		return nil, fmt.Errorf("device code response: missing required fields")
	}

	// Default interval to 5 seconds if server doesn't specify.
	if dcResp.Interval <= 0 {
		dcResp.Interval = 5
	}

	return &dcResp, nil
}

// PollForToken polls the token endpoint until the user authorizes,
// the device code expires, or the context is cancelled.
func (c *Client) PollForToken(ctx context.Context, config DeviceCodeConfig, dcResp *DeviceCodeResponse) (*TokenResponse, error) {
	if dcResp == nil {
		return nil, fmt.Errorf("poll for token: device code response must not be nil")
	}

	interval := time.Duration(dcResp.Interval) * time.Second
	deadline := time.Now().UTC().Add(time.Duration(dcResp.ExpiresIn) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		if time.Now().UTC().After(deadline) {
			return nil, ErrTimeout
		}

		tokenResp, pollErr, err := c.pollOnce(ctx, config, dcResp.DeviceCode)
		if err != nil {
			return nil, err
		}
		if tokenResp != nil {
			return tokenResp, nil
		}

		switch pollErr {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += slowDownBackoff * time.Second
			continue
		case "access_denied":
			return nil, ErrAccessDenied
		case "expired_token":
			return nil, ErrExpiredToken
		default:
			return nil, fmt.Errorf("device code poll: unexpected error: %s", pollErr)
		}
	}
}

// tokenErrorResponse is the error shape from the token endpoint.
type tokenErrorResponse struct {
	Error string `json:"error"`
}

// pollOnce makes a single token poll request.
// Returns (token, "", nil) on success, (nil, errorCode, nil) on expected OAuth errors,
// or (nil, "", err) on unexpected failures.
func (c *Client) pollOnce(ctx context.Context, config DeviceCodeConfig, deviceCode string) (*TokenResponse, string, error) {
	form := url.Values{
		"client_id":   {config.ClientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, "", fmt.Errorf("token poll request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("token poll request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, "", fmt.Errorf("token poll response read: %w", err)
	}

	// Success case.
	if resp.StatusCode == http.StatusOK {
		var tokenResp TokenResponse
		if err := json.Unmarshal(body, &tokenResp); err != nil {
			return nil, "", fmt.Errorf("token response parse: %w", err)
		}
		if tokenResp.AccessToken == "" {
			// Some providers return 200 with an error in the body (e.g. GitHub).
			var errResp tokenErrorResponse
			if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
				return nil, errResp.Error, nil
			}
			return nil, "", fmt.Errorf("token response: missing access_token")
		}
		return &tokenResp, "", nil
	}

	// Error case — parse the error code.
	var errResp tokenErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil, "", fmt.Errorf("token error response parse (HTTP %d): %w", resp.StatusCode, err)
	}
	if errResp.Error == "" {
		return nil, "", fmt.Errorf("token poll: HTTP %d with no error code", resp.StatusCode)
	}

	return nil, errResp.Error, nil
}

func validateConfig(config DeviceCodeConfig) error {
	if config.AuthEndpoint == "" {
		return fmt.Errorf("device code config: auth endpoint must not be empty")
	}
	if config.TokenEndpoint == "" {
		return fmt.Errorf("device code config: token endpoint must not be empty")
	}
	if config.ClientID == "" {
		return fmt.Errorf("device code config: client ID must not be empty")
	}
	return nil
}
