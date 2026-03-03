package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaBackend communicates with a local Ollama instance via its HTTP API.
type OllamaBackend struct {
	endpoint string
	model    string
	client   *http.Client
}

// NewOllamaBackend creates an Ollama backend with the given config.
func NewOllamaBackend(cfg OllamaConfig) *OllamaBackend {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	model := cfg.Model
	if model == "" {
		model = "llama3.2"
	}
	return &OllamaBackend{
		endpoint: strings.TrimRight(endpoint, "/"),
		model:    model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// NewOllamaBackendWithClient creates an Ollama backend with a custom HTTP client (for testing).
func NewOllamaBackendWithClient(cfg OllamaConfig, client *http.Client) *OllamaBackend {
	b := NewOllamaBackend(cfg)
	b.client = client
	return b
}

func (o *OllamaBackend) Name() string {
	return "ollama"
}

// ollamaRequest is the JSON body for the Ollama generate API.
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaResponse is the JSON response from the Ollama generate API.
type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (o *OllamaBackend) Complete(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", ErrEmptyPrompt
	}

	reqBody := ollamaRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ollama marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.endpoint+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("ollama create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("ollama decode response: %w", err)
	}

	text := strings.TrimSpace(ollamaResp.Response)
	if text == "" {
		return "", ErrEmptyResponse
	}

	return text, nil
}

func (o *OllamaBackend) Available(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.endpoint+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}
