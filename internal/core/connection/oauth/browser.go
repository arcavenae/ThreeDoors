package oauth

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser attempts to open the given URL in the user's default browser.
// Returns an error if the platform is unsupported or the command fails.
func OpenBrowser(ctx context.Context, url string) error {
	if url == "" {
		return fmt.Errorf("open browser: URL must not be empty")
	}

	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("open browser: unsupported platform %s", runtime.GOOS)
	}

	if err := exec.CommandContext(ctx, cmd, args...).Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
