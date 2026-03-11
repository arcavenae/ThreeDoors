package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WriteImprovement appends an improvement suggestion to improvements.txt.
// Format: [YYYY-MM-DD HH:MM:SS] (session-id) text
// Creates the file if it doesn't exist.
func WriteImprovement(baseDir, sessionID, text string) error {
	path := filepath.Join(baseDir, "improvements.txt")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("opening improvements file: %w", err)
	}
	defer f.Close() //nolint:errcheck // best-effort close on append file

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] (%s) %s\n", timestamp, sessionID, text)

	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("writing improvement: %w", err)
	}

	return nil
}
