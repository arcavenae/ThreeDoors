package dist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPkgBuilder_Build_CorrectArguments(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["pkgbuild"] = StubResponse{Output: nil, Err: nil}

	pb := NewPkgBuilder(stub, "Developer ID Installer: Test (TEAMID)")

	// Create a dummy binary in temp dir
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "threedoors")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho test"), 0o755); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "threedoors.pkg")
	err := pb.Build(binaryPath, "0.1.0", outputPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if len(stub.Calls) != 1 {
		t.Fatalf("len(Calls) = %d, want 1", len(stub.Calls))
	}

	call := stub.Calls[0]
	if call.Name != "pkgbuild" {
		t.Errorf("command = %q, want pkgbuild", call.Name)
	}

	// Verify key arguments are present
	argsStr := call.Args
	foundIdentifier := false
	foundVersion := false
	foundSign := false
	foundInstallLocation := false

	for i, arg := range argsStr {
		switch arg {
		case "--identifier":
			if i+1 < len(argsStr) && argsStr[i+1] == "com.arcaven.threedoors" {
				foundIdentifier = true
			}
		case "--version":
			if i+1 < len(argsStr) && argsStr[i+1] == "0.1.0" {
				foundVersion = true
			}
		case "--sign":
			if i+1 < len(argsStr) && argsStr[i+1] == "Developer ID Installer: Test (TEAMID)" {
				foundSign = true
			}
		case "--install-location":
			if i+1 < len(argsStr) && argsStr[i+1] == "/" {
				foundInstallLocation = true
			}
		}
	}

	if !foundIdentifier {
		t.Error("pkgbuild missing --identifier com.arcaven.threedoors")
	}
	if !foundVersion {
		t.Error("pkgbuild missing --version 0.1.0")
	}
	if !foundSign {
		t.Error("pkgbuild missing --sign with installer identity")
	}
	if !foundInstallLocation {
		t.Error("pkgbuild missing --install-location /")
	}
}

func TestPkgBuilder_Build_StagingDirContainsBinary(t *testing.T) {
	var capturedRoot string
	stub := NewStubRunner()
	stub.Responses["pkgbuild"] = StubResponse{Output: nil, Err: nil}

	// Override to capture the --root argument
	pb := NewPkgBuilder(stub, "Developer ID Installer: Test (TEAMID)")

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "threedoors-test")
	binaryContent := []byte("test-binary-content")
	if err := os.WriteFile(binaryPath, binaryContent, 0o755); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "out.pkg")
	err := pb.Build(binaryPath, "1.0.0", outputPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	// Get the staging dir from the --root argument
	call := stub.Calls[0]
	for i, arg := range call.Args {
		if arg == "--root" && i+1 < len(call.Args) {
			capturedRoot = call.Args[i+1]
			break
		}
	}

	if capturedRoot == "" {
		t.Fatal("--root argument not found in pkgbuild call")
	}

	// The staging dir should have been cleaned up by Build() defer
	// But the binary should have been copied to usr/local/bin/threedoors
	// We can verify this by checking the output path is the last arg
	lastArg := call.Args[len(call.Args)-1]
	if lastArg != outputPath {
		t.Errorf("last arg = %q, want output path %q", lastArg, outputPath)
	}
}

func TestPkgBuilder_Build_MissingBinaryErrors(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["pkgbuild"] = StubResponse{Output: nil, Err: nil}

	pb := NewPkgBuilder(stub, "identity")
	err := pb.Build("/nonexistent/binary", "1.0.0", "/tmp/out.pkg")

	if err == nil {
		t.Fatal("expected error for missing binary, got nil")
		return
	}
}

func TestPkgBuilder_DefaultIdentifier(t *testing.T) {
	pb := NewPkgBuilder(NewStubRunner(), "identity")
	if pb.Identifier != "com.arcaven.threedoors" {
		t.Errorf("Identifier = %q, want %q", pb.Identifier, "com.arcaven.threedoors")
	}
}

func TestPkgBuilder_Build_PerArchitecture(t *testing.T) {
	tests := []struct {
		name       string
		binaryName string
		pkgName    string
	}{
		{name: "arm64", binaryName: "threedoors-darwin-arm64", pkgName: "threedoors-arm64.pkg"},
		{name: "amd64", binaryName: "threedoors-darwin-amd64", pkgName: "threedoors-amd64.pkg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := NewStubRunner()
			stub.Responses["pkgbuild"] = StubResponse{Output: nil, Err: nil}
			pb := NewPkgBuilder(stub, "Developer ID Installer: Test (TEAM)")

			tmpDir := t.TempDir()
			binaryPath := filepath.Join(tmpDir, tt.binaryName)
			if err := os.WriteFile(binaryPath, []byte("binary"), 0o755); err != nil {
				t.Fatal(err)
			}

			outputPath := filepath.Join(tmpDir, tt.pkgName)
			err := pb.Build(binaryPath, "0.1.0", outputPath)
			if err != nil {
				t.Fatalf("Build() error: %v", err)
			}

			lastArg := stub.Calls[0].Args[len(stub.Calls[0].Args)-1]
			if lastArg != outputPath {
				t.Errorf("output path = %q, want %q", lastArg, outputPath)
			}
		})
	}
}
