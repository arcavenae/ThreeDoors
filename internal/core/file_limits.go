package core

import (
	"bufio"
	"fmt"
	"os"
)

// MaxTaskFileSize is the maximum allowed size for task YAML files (10MB).
const MaxTaskFileSize = 10 * 1024 * 1024

// MaxConfigFileSize is the maximum allowed size for config YAML files (1MB).
const MaxConfigFileSize = 1 * 1024 * 1024

// MaxJSONLLineSize is the maximum allowed line size for JSONL scanners (1MB).
// Matches the buffer used by the MCP transport.
const MaxJSONLLineSize = 1 * 1024 * 1024

// ReadFileWithLimit reads a file after verifying it does not exceed maxBytes.
// It returns an error before reading if the file is too large, preventing
// excessive memory allocation from corrupted or malicious files.
func ReadFileWithLimit(path string, maxBytes int64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("file %s exceeds size limit (%d bytes > %d bytes)", path, info.Size(), maxBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// NewLimitedScanner creates a bufio.Scanner with an explicit buffer size.
// The initial buffer is 64KB and the maximum line size is maxLineBytes.
// This prevents silent data loss from lines exceeding the default 64KB limit.
func NewLimitedScanner(f *os.File, maxLineBytes int) *bufio.Scanner {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineBytes)
	return scanner
}
