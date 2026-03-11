package core

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileWithLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		maxBytes int64
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "normal file within limit",
			content:  "key: value\n",
			maxBytes: 1024,
			wantErr:  false,
		},
		{
			name:     "empty file",
			content:  "",
			maxBytes: 1024,
			wantErr:  false,
		},
		{
			name:     "file exactly at limit",
			content:  strings.Repeat("x", 100),
			maxBytes: 100,
			wantErr:  false,
		},
		{
			name:     "file exceeds limit",
			content:  strings.Repeat("x", 101),
			maxBytes: 100,
			wantErr:  true,
			errMsg:   "exceeds size limit",
		},
		{
			name:     "large file exceeds limit",
			content:  strings.Repeat("x", 2048),
			maxBytes: 1024,
			wantErr:  true,
			errMsg:   "exceeds size limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "test.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}

			data, err := ReadFileWithLimit(path, tt.maxBytes)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if string(data) != tt.content {
					t.Errorf("content mismatch: got %q, want %q", string(data), tt.content)
				}
			}
		})
	}
}

func TestReadFileWithLimit_NonexistentFile(t *testing.T) {
	t.Parallel()
	_, err := ReadFileWithLimit("/nonexistent/path/file.yaml", 1024)
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestNewLimitedScanner(t *testing.T) {
	t.Parallel()

	t.Run("handles normal lines", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "test.jsonl")
		content := `{"id":"1","name":"task1"}
{"id":"2","name":"task2"}
`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		scanner := NewLimitedScanner(f, MaxJSONLLineSize)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("unexpected scanner error: %v", err)
		}
		if len(lines) != 2 {
			t.Errorf("expected 2 lines, got %d", len(lines))
		}
	})

	t.Run("handles lines exceeding small buffer", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "test.jsonl")

		shortLine := `{"id":"1","name":"short"}`
		longLine := `{"id":"2","name":"` + strings.Repeat("x", 200) + `"}`
		afterLine := `{"id":"3","name":"after"}`
		content := shortLine + "\n" + longLine + "\n" + afterLine + "\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		// Use a raw scanner with small buffer to test overflow behavior
		// (NewLimitedScanner uses 64KB initial buffer which would mask this)
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64), 100)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		// Scanner stops on buffer overflow error — this is expected behavior
		// The scanner will read the first line, then hit the long line and error
		if scanner.Err() == nil {
			t.Error("expected scanner error for oversized line, got nil")
		}
		if len(lines) != 1 {
			t.Errorf("expected 1 line before overflow, got %d", len(lines))
		}
	})
}
