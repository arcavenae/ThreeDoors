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

const claudeAPIURL = "https://api.anthropic.com/v1/messages"

// ClaudeBackend communicates with the Anthropic Claude API.
type ClaudeBackend struct {
	apiKey string
	model  string
	client *http.Client
	apiURL string
}

// NewClaudeBackend creates a Claude backend with the given config.
func NewClaudeBackend(cfg ClaudeConfig) *ClaudeBackend {
	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &ClaudeBackend{
		apiKey: cfg.APIKey,
		model:  model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		apiURL: claudeAPIURL,
	}
}

// NewClaudeBackendWithClient creates a Claude backend with a custom HTTP client and URL (for testing).
func NewClaudeBackendWithClient(cfg ClaudeConfig, client *http.Client, apiURL string) *ClaudeBackend {
	b := NewClaudeBackend(cfg)
	b.client = client
	b.apiURL = apiURL
	return b
}

func (c *ClaudeBackend) Name() string {
	return "claude"
}

// claudeRequest is the JSON body for the Anthropic messages API.
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse is the JSON response from the Anthropic messages API.
type claudeResponse struct {
	Content []claudeContent `json:"content"`
	Error   *claudeError    `json:"error,omitempty"`
}

type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *ClaudeBackend) Complete(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", ErrEmptyPrompt
	}

	if c.apiKey == "" {
		return "", fmt.Errorf("claude API key not configured: set ANTHROPIC_API_KEY env var")
	}

	reqBody := claudeRequest{
		Model:     c.model,
		MaxTokens: 4096,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("claude marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("claude create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("claude returned status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", fmt.Errorf("claude decode response: %w", err)
	}

	if claudeResp.Error != nil {
		return "", fmt.Errorf("claude API error: %s: %s", claudeResp.Error.Type, claudeResp.Error.Message)
	}

	var texts []string
	for _, c := range claudeResp.Content {
		if c.Type == "text" {
			texts = append(texts, c.Text)
		}
	}

	text := strings.TrimSpace(strings.Join(texts, "\n"))
	if text == "" {
		return "", ErrEmptyResponse
	}

	return text, nil
}

func (c *ClaudeBackend) Available(ctx context.Context) bool {
	return c.apiKey != ""
}
