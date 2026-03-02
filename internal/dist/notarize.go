package dist

import "fmt"

// Notarizer handles Apple notarization operations.
type Notarizer struct {
	Runner   CommandRunner
	AppleID  string
	Password string
	TeamID   string
	Timeout  int // seconds, default 900
}

// NewNotarizer creates a Notarizer with the given Apple credentials.
func NewNotarizer(runner CommandRunner, appleID, password, teamID string) *Notarizer {
	return &Notarizer{
		Runner:   runner,
		AppleID:  appleID,
		Password: password,
		TeamID:   teamID,
		Timeout:  900,
	}
}

// Submit submits a file for notarization and waits for completion.
// The file must be a zip, pkg, or dmg (standalone binaries must be zipped first).
func (n *Notarizer) Submit(filePath string) error {
	timeout := fmt.Sprintf("%d", n.Timeout)
	output, err := n.Runner.Run("xcrun", "notarytool", "submit",
		filePath,
		"--apple-id", n.AppleID,
		"--password", n.Password,
		"--team-id", n.TeamID,
		"--wait",
		"--timeout", timeout,
	)
	if err != nil {
		return fmt.Errorf("notarization failed for %s: %w\nOutput: %s", filePath, err, output)
	}
	return nil
}

// Staple staples the notarization ticket to a pkg or dmg.
// Note: standalone Mach-O binaries cannot be stapled.
func (n *Notarizer) Staple(filePath string) error {
	output, err := n.Runner.Run("xcrun", "stapler", "staple", filePath)
	if err != nil {
		return fmt.Errorf("stapling failed for %s: %w\nOutput: %s", filePath, err, output)
	}
	return nil
}

// Assess verifies a binary passes Gatekeeper assessment.
func (n *Notarizer) Assess(binaryPath string) error {
	output, err := n.Runner.Run("spctl",
		"--assess",
		"--type", "execute",
		binaryPath,
	)
	if err != nil {
		return fmt.Errorf("gatekeeper assessment failed for %s: %w\nOutput: %s", binaryPath, err, output)
	}
	return nil
}
