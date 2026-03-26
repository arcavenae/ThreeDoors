package connection

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core/connection/oauth"
)

func newTestTokenRefresher(t *testing.T) (*TokenRefresher, *stubCredentialStore, *ConnectionManager, *SyncEventLog) {
	t.Helper()
	tmpDir := t.TempDir()
	creds := newStubCredentialStore()
	mgr := NewConnectionManager(nil)
	eventLog := NewSyncEventLog(filepath.Join(tmpDir, "events"))
	refresher := NewTokenRefresher(creds, mgr, eventLog, nil)
	return refresher, creds, mgr, eventLog
}

func makeOAuthCredential(t *testing.T, tokenEndpoint string, expiresAt time.Time) string {
	t.Helper()
	meta := &oauth.TokenMetadata{
		AccessToken:   "at_old",
		RefreshToken:  "rt_valid",
		ExpiresAt:     expiresAt,
		TokenEndpoint: tokenEndpoint,
		ClientID:      "test-client",
	}
	serialized, err := oauth.MarshalCredential(meta)
	if err != nil {
		t.Fatalf("MarshalCredential: %v", err)
	}
	return serialized
}

func addConnectedConnection(t *testing.T, mgr *ConnectionManager) *Connection {
	t.Helper()
	conn, err := mgr.Add("github", "test-conn", nil)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := mgr.Transition(conn.ID, StateConnecting); err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if err := mgr.Transition(conn.ID, StateConnected); err != nil {
		t.Fatalf("Transition: %v", err)
	}
	return conn
}

func TestEnsureValidToken_PlainAPIKey(t *testing.T) {
	t.Parallel()

	refresher, creds, mgr, _ := newTestTokenRefresher(t)
	conn := addConnectedConnection(t, mgr)

	credKey := ConnCredentialKey(conn)
	if err := creds.Set(credKey, "sk_plain_api_key"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	result, err := refresher.EnsureValidToken(context.Background(), conn)
	if err != nil {
		t.Fatalf("EnsureValidToken: %v", err)
	}
	if result.Refreshed {
		t.Error("plain API key should not trigger refresh")
	}
	if result.AccessToken != "sk_plain_api_key" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "sk_plain_api_key")
	}
}

func TestEnsureValidToken_NotExpiring(t *testing.T) {
	t.Parallel()

	refresher, creds, mgr, _ := newTestTokenRefresher(t)
	conn := addConnectedConnection(t, mgr)

	credKey := ConnCredentialKey(conn)
	cred := makeOAuthCredential(t, "https://example.com/token", time.Now().UTC().Add(1*time.Hour))
	if err := creds.Set(credKey, cred); err != nil {
		t.Fatalf("Set: %v", err)
	}

	result, err := refresher.EnsureValidToken(context.Background(), conn)
	if err != nil {
		t.Fatalf("EnsureValidToken: %v", err)
	}
	if result.Refreshed {
		t.Error("token with 1h remaining should not refresh")
	}
	if result.AccessToken != "at_old" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "at_old")
	}
}

func TestEnsureValidToken_RefreshesExpiringToken(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := oauth.TokenResponse{
			AccessToken:  "at_refreshed",
			RefreshToken: "rt_new",
			ExpiresIn:    3600,
			TokenType:    "bearer",
		}
		data, _ := json.Marshal(resp)
		if _, err := w.Write(data); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	refresher, creds, mgr, eventLog := newTestTokenRefresher(t)
	refresher.oauth = oauth.NewClient(srv.Client())
	conn := addConnectedConnection(t, mgr)

	credKey := ConnCredentialKey(conn)
	cred := makeOAuthCredential(t, srv.URL, time.Now().UTC().Add(2*time.Minute))
	if err := creds.Set(credKey, cred); err != nil {
		t.Fatalf("Set: %v", err)
	}

	result, err := refresher.EnsureValidToken(context.Background(), conn)
	if err != nil {
		t.Fatalf("EnsureValidToken: %v", err)
	}
	if !result.Refreshed {
		t.Error("expected token to be refreshed")
	}
	if result.AccessToken != "at_refreshed" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "at_refreshed")
	}

	// Verify new credential was stored.
	stored, err := creds.Get(credKey)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	newMeta := oauth.UnmarshalCredential(stored)
	if newMeta.AccessToken != "at_refreshed" {
		t.Errorf("stored AccessToken = %q, want %q", newMeta.AccessToken, "at_refreshed")
	}
	if newMeta.RefreshToken != "rt_new" {
		t.Errorf("stored RefreshToken = %q, want %q", newMeta.RefreshToken, "rt_new")
	}

	// Verify token_refreshed event was logged.
	events, err := eventLog.EventsByType(conn.ID, EventTokenRefreshed, 10)
	if err != nil {
		t.Fatalf("EventsByType: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 token_refreshed event, got %d", len(events))
	}
}

func TestEnsureValidToken_PreservesRefreshTokenWhenNotRotated(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := oauth.TokenResponse{
			AccessToken: "at_refreshed",
			ExpiresIn:   3600,
			TokenType:   "bearer",
			// No refresh token in response — server didn't rotate it.
		}
		data, _ := json.Marshal(resp)
		if _, err := w.Write(data); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	refresher, creds, mgr, _ := newTestTokenRefresher(t)
	refresher.oauth = oauth.NewClient(srv.Client())
	conn := addConnectedConnection(t, mgr)

	credKey := ConnCredentialKey(conn)
	cred := makeOAuthCredential(t, srv.URL, time.Now().UTC().Add(2*time.Minute))
	if err := creds.Set(credKey, cred); err != nil {
		t.Fatalf("Set: %v", err)
	}

	result, err := refresher.EnsureValidToken(context.Background(), conn)
	if err != nil {
		t.Fatalf("EnsureValidToken: %v", err)
	}
	if !result.Refreshed {
		t.Error("expected token to be refreshed")
	}

	// Verify old refresh token was preserved.
	stored, err := creds.Get(credKey)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	newMeta := oauth.UnmarshalCredential(stored)
	if newMeta.RefreshToken != "rt_valid" {
		t.Errorf("stored RefreshToken = %q, want %q (preserved)", newMeta.RefreshToken, "rt_valid")
	}
}

func TestEnsureValidToken_RefreshTokenExpired_TransitionsToAuthExpired(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(map[string]string{"error": "invalid_grant"})
		if _, err := w.Write(data); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	refresher, creds, mgr, eventLog := newTestTokenRefresher(t)
	refresher.oauth = oauth.NewClient(srv.Client())
	conn := addConnectedConnection(t, mgr)

	credKey := ConnCredentialKey(conn)
	cred := makeOAuthCredential(t, srv.URL, time.Now().UTC().Add(2*time.Minute))
	if err := creds.Set(credKey, cred); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_, err := refresher.EnsureValidToken(context.Background(), conn)
	if !errors.Is(err, oauth.ErrRefreshTokenExpired) {
		t.Errorf("expected ErrRefreshTokenExpired, got %v", err)
	}

	// Verify connection transitioned to AuthExpired.
	got, _ := mgr.Get(conn.ID)
	if got.State != StateAuthExpired {
		t.Errorf("State = %s, want %s", got.State, StateAuthExpired)
	}

	// Verify reauth_required event was logged.
	events, err := eventLog.EventsByType(conn.ID, EventReauthRequired, 10)
	if err != nil {
		t.Fatalf("EventsByType: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 reauth_required event, got %d", len(events))
	}
}

func TestEnsureValidToken_MissingCredential(t *testing.T) {
	t.Parallel()

	refresher, _, mgr, _ := newTestTokenRefresher(t)
	conn := addConnectedConnection(t, mgr)

	_, err := refresher.EnsureValidToken(context.Background(), conn)
	if err == nil {
		t.Fatal("expected error for missing credential")
	}
}

func TestHandleUnauthorized(t *testing.T) {
	t.Parallel()

	refresher, _, mgr, eventLog := newTestTokenRefresher(t)
	conn := addConnectedConnection(t, mgr)

	refresher.HandleUnauthorized(conn)

	got, _ := mgr.Get(conn.ID)
	if got.State != StateAuthExpired {
		t.Errorf("State = %s, want %s", got.State, StateAuthExpired)
	}

	events, err := eventLog.EventsByType(conn.ID, EventReauthRequired, 10)
	if err != nil {
		t.Fatalf("EventsByType: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 reauth_required event, got %d", len(events))
	}
	if events[0].Error != "API returned 401 Unauthorized" {
		t.Errorf("event error = %q, want %q", events[0].Error, "API returned 401 Unauthorized")
	}
}
