package quota

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

// jsonlEntry mirrors the relevant fields of a Claude Code JSONL line.
// Only assistant-type entries carry usage data.
type jsonlEntry struct {
	Type      string    `json:"type"`
	SessionID string    `json:"sessionId"`
	Timestamp time.Time `json:"timestamp"`
	Message   *struct {
		Model   string          `json:"model"`
		Content json.RawMessage `json:"content"`
		Usage   *struct {
			InputTokens              int64 `json:"input_tokens"`
			OutputTokens             int64 `json:"output_tokens"`
			CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// contentBlock is used to inspect the type field of content array elements.
type contentBlock struct {
	Type string `json:"type"`
}

// ParseFile reads a single JSONL file and extracts interactions with token
// usage data. Malformed or non-assistant entries are skipped (AC3).
func ParseFile(path string) (interactions []Interaction, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open session file %s: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close session file %s: %w", path, cerr)
		}
	}()

	return parseReader(f, path)
}

func parseReader(r io.Reader, source string) ([]Interaction, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var interactions []Interaction
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry jsonlEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			slog.Warn("skipping malformed JSONL line",
				"source", source,
				"line", lineNum,
				"error", err,
			)
			continue
		}

		if entry.Type != "assistant" {
			continue
		}
		if entry.Message == nil || entry.Message.Usage == nil {
			continue
		}

		u := entry.Message.Usage
		hasToolUse := containsToolUse(entry.Message.Content)
		interactions = append(interactions, Interaction{
			SessionID:  entry.SessionID,
			Timestamp:  entry.Timestamp,
			Model:      entry.Message.Model,
			HasToolUse: hasToolUse,
			SourcePath: source,
			Tokens: TokenCount{
				InputTokens:              u.InputTokens,
				OutputTokens:             u.OutputTokens,
				CacheCreationInputTokens: u.CacheCreationInputTokens,
				CacheReadInputTokens:     u.CacheReadInputTokens,
			},
		})
	}

	if err := scanner.Err(); err != nil {
		return interactions, fmt.Errorf("scan session file %s: %w", source, err)
	}
	return interactions, nil
}

// containsToolUse checks if a raw JSON content field contains any tool_use blocks.
// Content may be a string (user messages) or an array of objects (assistant messages).
func containsToolUse(raw json.RawMessage) bool {
	if len(raw) == 0 || raw[0] != '[' {
		return false
	}
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return false
	}
	for _, b := range blocks {
		if b.Type == "tool_use" {
			return true
		}
	}
	return false
}

// ParseFiles reads multiple JSONL files and returns all interactions.
// Errors on individual files are logged and skipped (AC3).
func ParseFiles(paths []string) []Interaction {
	var all []Interaction
	for _, p := range paths {
		interactions, err := ParseFile(p)
		if err != nil {
			slog.Warn("skipping session file",
				"path", p,
				"error", err,
			)
			continue
		}
		all = append(all, interactions...)
	}
	return all
}
