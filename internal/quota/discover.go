package quota

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultConfigDir returns the Claude Code config directory, respecting
// the CLAUDE_CONFIG_DIR environment variable (AC7).
func DefaultConfigDir() string {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

// DiscoverSessionFiles finds all JSONL session files under the config
// directory's projects/ subtree.
func DiscoverSessionFiles(configDir string) ([]string, error) {
	projectsDir := filepath.Join(configDir, "projects")

	info, err := os.Stat(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	var files []string
	err = filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
