package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRefreshToken_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("expected form content type, got %s", ct)
		}

		gt := r.URL.Query().Get("grant_type")
		if gt != "refresh_token" {
			t.Errorf("grant_type = %q, want %q", gt, "refresh_token")
		}
		rt := r.URL.Query().Get("refresh_token")
		if rt != "rt_old_123" {
			t.Errorf("refresh_token = %q, want %q", rt, "rt_old_123")
		}
		cid := r.URL.Query().Get("client_id")
		if cid != "test-client" {
			t.Errorf("client_id = %q, want %q", cid, "test-client")
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, TokenResponse{
			AccessToken:  "at_new_456",
			RefreshToken: "rt_new_789",
			ExpiresIn:    7200,
			TokenType:    "bearer",
		})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	token, err := client.RefreshToken(context.Background(), srv.URL, "test-client", "rt_old_123")
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}

	if token.AccessToken != "at_new_456" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "at_new_456")
	}
	if token.RefreshToken != "rt_new_789" {
		t.Errorf("RefreshToken = %q, want %q", token.RefreshToken, "rt_new_789")
	}
	if token.ExpiresIn != 7200 {
		t.Errorf("ExpiresIn = %d, want 7200", token.ExpiresIn)
	}
}

func TestRefreshToken_ExpiredRefreshToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		errorCode string
	}{
		{name: "invalid_grant", errorCode: "invalid_grant"},
		{name: "expired_token", errorCode: "expired_token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				data, _ := json.Marshal(map[string]string{"error": tt.errorCode})
				if _, err := w.Write(data); err != nil {
					t.Errorf("failed to write: %v", err)
				}
			}))
			t.Cleanup(srv.Close)

			client := NewClient(srv.Client())
			_, err := client.RefreshToken(context.Background(), srv.URL, "test-client", "rt_expired")
			if !errors.Is(err, ErrRefreshTokenExpired) {
				t.Errorf("expected ErrRefreshTokenExpired, got %v", err)
			}
		})
	}
}

func TestRefreshToken_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("internal error")); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	_, err := client.RefreshToken(context.Background(), srv.URL, "test-client", "rt_123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "HTTP 500") {
		t.Errorf("error %q should contain 'HTTP 500'", err.Error())
	}
}

func TestRefreshToken_MissingAccessTokenInResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, map[string]string{"token_type": "bearer"})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	_, err := client.RefreshToken(context.Background(), srv.URL, "test-client", "rt_123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "missing access_token") {
		t.Errorf("error %q should contain 'missing access_token'", err.Error())
	}
}

func TestRefreshToken_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tokenEndpoint string
		clientID      string
		refreshToken  string
		errContain    string
	}{
		{
			name:         "empty token endpoint",
			clientID:     "c",
			refreshToken: "r",
			errContain:   "token endpoint",
		},
		{
			name:          "empty client ID",
			tokenEndpoint: "https://example.com/token",
			refreshToken:  "r",
			errContain:    "client ID",
		},
		{
			name:          "empty refresh token",
			tokenEndpoint: "https://example.com/token",
			clientID:      "c",
			errContain:    "refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := NewClient(nil)
			_, err := client.RefreshToken(context.Background(), tt.tokenEndpoint, tt.clientID, tt.refreshToken)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !containsStr(err.Error(), tt.errContain) {
				t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
			}
		})
	}
}

func TestRefreshToken_HTTPClientError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close()

	client := NewClient(srv.Client())
	_, err := client.RefreshToken(context.Background(), srv.URL, "test-client", "rt_123")
	if err == nil {
		t.Fatal("expected error from closed server, got nil")
	}
	if !containsStr(err.Error(), "token refresh request") {
		t.Errorf("error %q should contain 'token refresh request'", err.Error())
	}
}

func TestRefreshToken_InvalidJSONResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("not json")); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	_, err := client.RefreshToken(context.Background(), srv.URL, "test-client", "rt_123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "response parse") {
		t.Errorf("error %q should contain 'response parse'", err.Error())
	}
}

func TestRefreshToken_401WithNonGrantError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		data, _ := json.Marshal(map[string]string{"error": "invalid_client"})
		if _, err := w.Write(data); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client())
	_, err := client.RefreshToken(context.Background(), srv.URL, "test-client", "rt_123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should NOT be ErrRefreshTokenExpired — invalid_client is a different error.
	if errors.Is(err, ErrRefreshTokenExpired) {
		t.Error("should not be ErrRefreshTokenExpired for invalid_client")
	}
}
