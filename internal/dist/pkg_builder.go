package dist

import (
	"fmt"
	"os"
	"path/filepath"
)

// PkgBuilder handles macOS pkg installer creation.
type PkgBuilder struct {
	Runner          CommandRunner
	SigningIdentity string
	Identifier      string
}

// NewPkgBuilder creates a PkgBuilder with the given installer signing identity.
func NewPkgBuilder(runner CommandRunner, signingIdentity string) *PkgBuilder {
	return &PkgBuilder{
		Runner:          runner,
		SigningIdentity: signingIdentity,
		Identifier:      "com.arcaven.threedoors",
	}
}

// Build creates a signed pkg installer from a binary.
func (pb *PkgBuilder) Build(binaryPath, version, outputPath string) error {
	stagingDir, err := os.MkdirTemp("", "threedoors-pkg-*")
	if err != nil {
		return fmt.Errorf("failed to create staging dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(stagingDir) }()

	binDir := filepath.Join(stagingDir, "usr", "local", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("failed to create bin dir: %w", err)
	}

	srcData, err := os.ReadFile(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %w", err)
	}
	destPath := filepath.Join(binDir, "threedoors")
	if err := os.WriteFile(destPath, srcData, 0o755); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	output, err := pb.Runner.Run("pkgbuild",
		"--root", stagingDir,
		"--identifier", pb.Identifier,
		"--version", version,
		"--install-location", "/",
		"--sign", pb.SigningIdentity,
		outputPath,
	)
	if err != nil {
		return fmt.Errorf("pkgbuild failed: %w\nOutput: %s", err, output)
	}
	return nil
}
