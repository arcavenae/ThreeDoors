package linear

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetectEnvToken_LinearAPIKey(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "lin_api_abc123")

	token, envVar := DetectEnvToken()
	if token != "lin_api_abc123" {
		t.Errorf("token = %q, want %q", token, "lin_api_abc123")
	}
	if envVar != "LINEAR_API_KEY" {
		t.Errorf("envVar = %q, want %q", envVar, "LINEAR_API_KEY")
	}
}

func TestDetectEnvToken_NotSet(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "")

	token, envVar := DetectEnvToken()
	if token != "" {
		t.Errorf("token = %q, want empty", token)
	}
	if envVar != "" {
		t.Errorf("envVar = %q, want empty", envVar)
	}
}

func TestVerifyToken_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "lin_api_test" {
			t.Errorf("Authorization header = %q, want %q", auth, "lin_api_test")
		}

		resp := graphQLResponse{
			Data: json.RawMessage(`{"viewer":{"id":"u1","name":"Test User","email":"test@linear.app"}}`),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	name, err := VerifyToken(context.Background(), "lin_api_test", server.URL)
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}
	if name != "Test User" {
		t.Errorf("name = %q, want %q", name, "Test User")
	}
}

func TestVerifyToken_EmptyToken(t *testing.T) {
	t.Parallel()

	_, err := VerifyToken(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
	if !strings.Contains(err.Error(), "token must not be empty") {
		t.Errorf("error %q should contain 'token must not be empty'", err.Error())
	}
}

func TestVerifyToken_AuthFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	_, err := VerifyToken(context.Background(), "bad_key", server.URL)
	if err == nil {
		t.Fatal("expected auth error, got nil")
	}
}

func TestVerifyToken_CancelledContext(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := graphQLResponse{
			Data: json.RawMessage(`{"viewer":{"id":"u1","name":"Test","email":"t@t.com"}}`),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := VerifyToken(ctx, "some-token", server.URL)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
