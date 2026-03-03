package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClaudeBackend_Name(t *testing.T) {
	t.Parallel()
	b := NewClaudeBackend(ClaudeConfig{APIKey: "test"})
	if b.Name() != "claude" {
		t.Errorf("Name() = %q, want %q", b.Name(), "claude")
	}
}

func TestClaudeBackend_Complete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		prompt     string
		apiKey     string
		serverResp string
		statusCode int
		wantErr    bool
		wantText   string
	}{
		{
			name:       "successful completion",
			prompt:     "hello",
			apiKey:     "sk-test",
			serverResp: `{"content":[{"type":"text","text":"world"}]}`,
			statusCode: http.StatusOK,
			wantText:   "world",
		},
		{
			name:    "empty prompt",
			prompt:  "",
			apiKey:  "sk-test",
			wantErr: true,
		},
		{
			name:    "missing api key",
			prompt:  "hello",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:       "api error response",
			prompt:     "hello",
			apiKey:     "sk-test",
			serverResp: `{"content":[],"error":{"type":"invalid_request_error","message":"bad request"}}`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "server error",
			prompt:     "hello",
			apiKey:     "sk-test",
			serverResp: `{"error":{"type":"overloaded_error","message":"overloaded"}}`,
			statusCode: http.StatusServiceUnavailable,
			wantErr:    true,
		},
		{
			name:       "empty content",
			prompt:     "hello",
			apiKey:     "sk-test",
			serverResp: `{"content":[]}`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "multiple content blocks",
			prompt:     "hello",
			apiKey:     "sk-test",
			serverResp: `{"content":[{"type":"text","text":"first"},{"type":"text","text":"second"}]}`,
			statusCode: http.StatusOK,
			wantText:   "first\nsecond",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.Header.Get("x-api-key") != tt.apiKey {
					t.Errorf("expected x-api-key %q, got %q", tt.apiKey, r.Header.Get("x-api-key"))
				}
				if r.Header.Get("anthropic-version") != "2023-06-01" {
					t.Errorf("expected anthropic-version 2023-06-01, got %q", r.Header.Get("anthropic-version"))
				}
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResp))
			}))
			t.Cleanup(server.Close)

			backend := NewClaudeBackendWithClient(
				ClaudeConfig{APIKey: tt.apiKey, Model: "test-model"},
				server.Client(),
				server.URL,
			)

			result, err := backend.Complete(context.Background(), tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result != tt.wantText {
				t.Errorf("Complete() = %q, want %q", result, tt.wantText)
			}
		})
	}
}

func TestClaudeBackend_Available(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		apiKey string
		want   bool
	}{
		{"available with key", "sk-test", true},
		{"unavailable without key", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := NewClaudeBackend(ClaudeConfig{APIKey: tt.apiKey})
			if got := b.Available(context.Background()); got != tt.want {
				t.Errorf("Available() = %v, want %v", got, tt.want)
			}
		})
	}
}
