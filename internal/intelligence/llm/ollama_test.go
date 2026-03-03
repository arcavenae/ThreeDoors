package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaBackend_Name(t *testing.T) {
	t.Parallel()
	b := NewOllamaBackend(OllamaConfig{})
	if b.Name() != "ollama" {
		t.Errorf("Name() = %q, want %q", b.Name(), "ollama")
	}
}

func TestOllamaBackend_Complete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		prompt     string
		serverResp string
		statusCode int
		wantErr    bool
		wantText   string
	}{
		{
			name:       "successful completion",
			prompt:     "hello",
			serverResp: `{"response":"world","done":true}`,
			statusCode: http.StatusOK,
			wantText:   "world",
		},
		{
			name:    "empty prompt",
			prompt:  "",
			wantErr: true,
		},
		{
			name:       "empty response",
			prompt:     "hello",
			serverResp: `{"response":"","done":true}`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "server error",
			prompt:     "hello",
			serverResp: `{"error":"model not found"}`,
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "trims whitespace",
			prompt:     "hello",
			serverResp: `{"response":"  trimmed  ","done":true}`,
			statusCode: http.StatusOK,
			wantText:   "trimmed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/generate" {
					t.Errorf("expected /api/generate, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResp))
			}))
			t.Cleanup(server.Close)

			backend := NewOllamaBackendWithClient(
				OllamaConfig{Endpoint: server.URL, Model: "test-model"},
				server.Client(),
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

func TestOllamaBackend_Available(t *testing.T) {
	t.Parallel()

	t.Run("available when server responds OK", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/tags" {
				t.Errorf("expected /api/tags, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models":[]}`))
		}))
		t.Cleanup(server.Close)

		backend := NewOllamaBackendWithClient(
			OllamaConfig{Endpoint: server.URL},
			server.Client(),
		)

		if !backend.Available(context.Background()) {
			t.Error("Available() = false, want true")
		}
	})

	t.Run("unavailable when server down", func(t *testing.T) {
		t.Parallel()
		backend := NewOllamaBackend(OllamaConfig{Endpoint: "http://localhost:1"})
		if backend.Available(context.Background()) {
			t.Error("Available() = true, want false")
		}
	})
}
