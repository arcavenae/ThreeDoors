package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func testConfig(authURL, tokenURL string) DeviceCodeConfig {
	if authURL == "" {
		authURL = "https://placeholder.example.com/auth"
	}
	if tokenURL == "" {
		tokenURL = "https://placeholder.example.com/token"
	}
	return DeviceCodeConfig{
		AuthEndpoint:  authURL,
		TokenEndpoint: tokenURL,
		ClientID:      "test-client-id",
		Scopes:        []string{"repo", "read:user"},
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("failed to write JSON response: %v", err)
	}
}

func TestStartDeviceCodeFlow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		handler    http.HandlerFunc
		config     func(url string) DeviceCodeConfig
		wantErr    bool
		wantCode   string
		wantUser   string
		wantURI    string
		errContain string
	}{
		{
			name: "successful flow",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
					t.Errorf("expected form content type, got %s", ct)
				}
				if err := r.ParseForm(); err != nil {
					t.Fatal(err)
				}
				if cid := r.FormValue("client_id"); cid != "test-client-id" {
					t.Errorf("expected client_id test-client-id, got %s", cid)
				}
				if scope := r.FormValue("scope"); scope != "repo read:user" {
					t.Errorf("expected scope 'repo read:user', got %s", scope)
				}
				w.Header().Set("Content-Type", "application/json")
				writeJSON(t, w, DeviceCodeResponse{
					DeviceCode:      "dev-code-123",
					UserCode:        "ABCD-1234",
					VerificationURI: "https://example.com/device",
					ExpiresIn:       900,
					Interval:        5,
				})
			},
			config: func(url string) DeviceCodeConfig {
				return testConfig(url, "")
			},
			wantCode: "dev-code-123",
			wantUser: "ABCD-1234",
			wantURI:  "https://example.com/device",
		},
		{
			name: "no scopes sent when empty",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Fatal(err)
				}
				if scope := r.FormValue("scope"); scope != "" {
					t.Errorf("expected no scope, got %s", scope)
				}
				w.Header().Set("Content-Type", "application/json")
				writeJSON(t, w, DeviceCodeResponse{
					DeviceCode:      "dev-code-456",
					UserCode:        "EFGH-5678",
					VerificationURI: "https://example.com/device",
					ExpiresIn:       600,
					Interval:        5,
				})
			},
			config: func(url string) DeviceCodeConfig {
				return DeviceCodeConfig{
					AuthEndpoint:  url,
					TokenEndpoint: "https://placeholder.example.com/token",
					ClientID:      "test-client-id",
				}
			},
			wantCode: "dev-code-456",
			wantUser: "EFGH-5678",
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := w.Write([]byte("internal server error")); err != nil {
					t.Errorf("failed to write error response: %v", err)
				}
			},
			config: func(url string) DeviceCodeConfig {
				return testConfig(url, "")
			},
			wantErr:    true,
			errContain: "HTTP 500",
		},
		{
			name: "missing required fields in response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				writeJSON(t, w, map[string]string{"device_code": "abc"})
			},
			config: func(url string) DeviceCodeConfig {
				return testConfig(url, "")
			},
			wantErr:    true,
			errContain: "missing required fields",
		},
		{
			name: "default interval when server omits it",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				writeJSON(t, w, DeviceCodeResponse{
					DeviceCode:      "dev-code-789",
					UserCode:        "IJKL-9012",
					VerificationURI: "https://example.com/device",
					ExpiresIn:       300,
					Interval:        0, // omitted
				})
			},
			config: func(url string) DeviceCodeConfig {
				return testConfig(url, "")
			},
			wantCode: "dev-code-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(tt.handler)
			t.Cleanup(srv.Close)

			client := NewClient(srv.Client())
			config := tt.config(srv.URL)
			resp, err := client.StartDeviceCodeFlow(context.Background(), config)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !containsStr(err.Error(), tt.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantCode != "" && resp.DeviceCode != tt.wantCode {
				t.Errorf("DeviceCode = %q, want %q", resp.DeviceCode, tt.wantCode)
			}
			if tt.wantUser != "" && resp.UserCode != tt.wantUser {
				t.Errorf("UserCode = %q, want %q", resp.UserCode, tt.wantUser)
			}
			if tt.wantURI != "" && resp.VerificationURI != tt.wantURI {
				t.Errorf("VerificationURI = %q, want %q", resp.VerificationURI, tt.wantURI)
			}
			if resp.Interval <= 0 {
				t.Error("Interval should be > 0 (defaulted to 5)")
			}
		})
	}
}

func TestStartDeviceCodeFlow_ConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		config     DeviceCodeConfig
		errContain string
	}{
		{
			name:       "empty auth endpoint",
			config:     DeviceCodeConfig{TokenEndpoint: "https://example.com/token", ClientID: "c"},
			errContain: "auth endpoint",
		},
		{
			name:       "empty token endpoint",
			config:     DeviceCodeConfig{AuthEndpoint: "https://example.com/auth", ClientID: "c"},
			errContain: "token endpoint",
		},
		{
			name:       "empty client ID",
			config:     DeviceCodeConfig{AuthEndpoint: "https://example.com/auth", TokenEndpoint: "https://example.com/token"},
			errContain: "client ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := NewClient(nil)
			_, err := client.StartDeviceCodeFlow(context.Background(), tt.config)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !containsStr(err.Error(), tt.errContain) {
				t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
			}
		})
	}
}

func TestPollForToken_Success(t *testing.T) {
	t.Parallel()

	var pollCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		if gt := r.FormValue("grant_type"); gt != "urn:ietf:params:oauth:grant-type:device_code" {
			t.Errorf("expected device_code grant_type, got %s", gt)
		}
		if dc := r.FormValue("device_code"); dc != "dev-code-123" {
			t.Errorf("expected device_code dev-code-123, got %s", dc)
		}

		n := pollCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if n < 3 {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(t, w, tokenErrorResponse{Error: "authorization_pending"})
			return
		}
		writeJSON(t, w, TokenResponse{
			AccessToken:  "gho_abc123",
			RefreshToken: "ghr_refresh456",
			ExpiresIn:    3600,
			TokenType:    "bearer",
		})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-123",
		ExpiresIn:  30,
		Interval:   1, // 1s for fast tests
	}

	token, err := client.PollForToken(context.Background(), config, dcResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "gho_abc123" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "gho_abc123")
	}
	if token.RefreshToken != "ghr_refresh456" {
		t.Errorf("RefreshToken = %q, want %q", token.RefreshToken, "ghr_refresh456")
	}
	if token.TokenType != "bearer" {
		t.Errorf("TokenType = %q, want %q", token.TokenType, "bearer")
	}
	if got := pollCount.Load(); got != 3 {
		t.Errorf("expected 3 polls, got %d", got)
	}
}

func TestPollForToken_SlowDown(t *testing.T) {
	t.Parallel()

	var pollCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := pollCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		if n == 1 {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(t, w, tokenErrorResponse{Error: "slow_down"})
			return
		}
		writeJSON(t, w, TokenResponse{
			AccessToken: "token-after-slowdown",
			TokenType:   "bearer",
		})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-slow",
		ExpiresIn:  30,
		Interval:   1, // 1s base
	}

	start := time.Now()
	token, err := client.PollForToken(context.Background(), config, dcResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "token-after-slowdown" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "token-after-slowdown")
	}

	// After slow_down, interval increases by 5s. First wait is 1s, second wait is 6s.
	// Total should be at least 6s.
	elapsed := time.Since(start)
	if elapsed < 6*time.Second {
		t.Errorf("expected at least 6s elapsed (slow_down backoff), got %v", elapsed)
	}
}

func TestPollForToken_Timeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, tokenErrorResponse{Error: "authorization_pending"})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-timeout",
		ExpiresIn:  2, // expires fast
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if err != ErrTimeout {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

func TestPollForToken_ContextCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, tokenErrorResponse{Error: "authorization_pending"})
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-cancel",
		ExpiresIn:  60,
		Interval:   1,
	}

	_, err := client.PollForToken(ctx, config, dcResp)
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestPollForToken_AccessDenied(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, tokenErrorResponse{Error: "access_denied"})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-denied",
		ExpiresIn:  60,
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestPollForToken_ExpiredToken(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, tokenErrorResponse{Error: "expired_token"})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-expired",
		ExpiresIn:  60,
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestPollForToken_NilDeviceCodeResponse(t *testing.T) {
	t.Parallel()
	client := NewClient(nil)
	_, err := client.PollForToken(context.Background(), DeviceCodeConfig{}, nil)
	if err == nil {
		t.Fatal("expected error for nil device code response, got nil")
	}
}

func TestPollForToken_GitHubStyleOKWithError(t *testing.T) {
	t.Parallel()

	var pollCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := pollCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if n < 2 {
			writeJSON(t, w, map[string]string{
				"error": "authorization_pending",
			})
			return
		}
		writeJSON(t, w, TokenResponse{
			AccessToken: "gho_github_style",
			TokenType:   "bearer",
		})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-github",
		ExpiresIn:  30,
		Interval:   1,
	}

	token, err := client.PollForToken(context.Background(), config, dcResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "gho_github_style" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "gho_github_style")
	}
}

func TestPollForToken_UnexpectedError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, tokenErrorResponse{Error: "some_unknown_error"})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-unexpected",
		ExpiresIn:  30,
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "unexpected error") {
		t.Errorf("error %q should contain 'unexpected error'", err.Error())
	}
}

func TestPollOnce_MissingAccessToken(t *testing.T) {
	t.Parallel()

	// Server returns 200 with no access_token and no error field.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, map[string]string{"token_type": "bearer"})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-notoken",
		ExpiresIn:  30,
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "missing access_token") {
		t.Errorf("error %q should contain 'missing access_token'", err.Error())
	}
}

func TestPollOnce_UnparseableErrorResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("not json at all")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-badjson",
		ExpiresIn:  30,
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "token error response parse") {
		t.Errorf("error %q should contain 'token error response parse'", err.Error())
	}
}

func TestPollOnce_EmptyErrorCode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, map[string]string{"error": ""})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-emptyerr",
		ExpiresIn:  30,
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "no error code") {
		t.Errorf("error %q should contain 'no error code'", err.Error())
	}
}

func TestPollOnce_UnparseableSuccessResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("not valid json")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-badjson-success",
		ExpiresIn:  30,
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "token response parse") {
		t.Errorf("error %q should contain 'token response parse'", err.Error())
	}
}

func TestStartDeviceCodeFlow_InvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("{invalid json")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	config := testConfig(srv.URL, "")

	_, err := client.StartDeviceCodeFlow(context.Background(), config)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "response parse") {
		t.Errorf("error %q should contain 'response parse'", err.Error())
	}
}

func TestStartDeviceCodeFlow_HTTPClientError(t *testing.T) {
	t.Parallel()

	// Use a server that's already closed so the HTTP request fails.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close()

	client := NewClient(srv.Client())
	config := testConfig(srv.URL, "")

	_, err := client.StartDeviceCodeFlow(context.Background(), config)
	if err == nil {
		t.Fatal("expected error from closed server, got nil")
	}
	if !containsStr(err.Error(), "device code request") {
		t.Errorf("error %q should contain 'device code request'", err.Error())
	}
}

func TestStartDeviceCodeFlow_InvalidAuthEndpoint(t *testing.T) {
	t.Parallel()

	client := NewClient(nil)
	config := DeviceCodeConfig{
		AuthEndpoint:  "://bad-url",
		TokenEndpoint: "https://example.com/token",
		ClientID:      "test",
	}

	_, err := client.StartDeviceCodeFlow(context.Background(), config)
	if err == nil {
		t.Fatal("expected error from invalid URL, got nil")
	}
	if !containsStr(err.Error(), "device code request") {
		t.Errorf("error %q should contain 'device code request'", err.Error())
	}
}

func TestPollOnce_HTTPClientError(t *testing.T) {
	t.Parallel()

	// Use a server that's already closed so the HTTP request fails.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close()

	client := NewClient(srv.Client())
	config := testConfig("", srv.URL)
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-httperr",
		ExpiresIn:  30,
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err == nil {
		t.Fatal("expected error from closed server, got nil")
	}
	if !containsStr(err.Error(), "token poll request") {
		t.Errorf("error %q should contain 'token poll request'", err.Error())
	}
}

func TestPollOnce_InvalidTokenEndpoint(t *testing.T) {
	t.Parallel()

	client := NewClient(nil)
	config := DeviceCodeConfig{
		AuthEndpoint:  "https://example.com/auth",
		TokenEndpoint: "://bad-url",
		ClientID:      "test",
	}
	dcResp := &DeviceCodeResponse{
		DeviceCode: "dev-code-badurl",
		ExpiresIn:  30,
		Interval:   1,
	}

	_, err := client.PollForToken(context.Background(), config, dcResp)
	if err == nil {
		t.Fatal("expected error from invalid URL, got nil")
	}
	if !containsStr(err.Error(), "token poll request") {
		t.Errorf("error %q should contain 'token poll request'", err.Error())
	}
}

func TestNewClient_DefaultHTTPClient(t *testing.T) {
	t.Parallel()
	client := NewClient(nil)
	if client.httpClient == nil {
		t.Fatal("expected default HTTP client, got nil")
	}
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", client.httpClient.Timeout)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
