package dist

import "fmt"

// CodeSigner handles macOS code signing operations.
type CodeSigner struct {
	Runner   CommandRunner
	Identity string
}

// NewCodeSigner creates a CodeSigner with the given signing identity.
func NewCodeSigner(runner CommandRunner, identity string) *CodeSigner {
	return &CodeSigner{Runner: runner, Identity: identity}
}

// Sign signs a binary with the Developer ID Application certificate.
// Uses --force --options runtime for hardened runtime (required for notarization).
func (cs *CodeSigner) Sign(binaryPath string) error {
	output, err := cs.Runner.Run("codesign",
		"--force",
		"--options", "runtime",
		"--sign", cs.Identity,
		"--timestamp",
		binaryPath,
	)
	if err != nil {
		return fmt.Errorf("codesign failed for %s: %w\nOutput: %s", binaryPath, err, output)
	}
	return nil
}

// Verify verifies a signed binary passes strict codesign checks.
func (cs *CodeSigner) Verify(binaryPath string) error {
	output, err := cs.Runner.Run("codesign",
		"--verify",
		"--deep",
		"--strict",
		binaryPath,
	)
	if err != nil {
		return fmt.Errorf("codesign verification failed for %s: %w\nOutput: %s", binaryPath, err, output)
	}
	return nil
}
