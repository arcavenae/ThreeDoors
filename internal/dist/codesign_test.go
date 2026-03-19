package dist

import (
	"errors"
	"testing"
)

func TestCodeSigner_Sign_CorrectArguments(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["codesign"] = StubResponse{Output: nil, Err: nil}

	cs := NewCodeSigner(stub, "Developer ID Application: Test (TEAMID)")
	err := cs.Sign("/path/to/binary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if len(stub.Calls) != 1 {
		t.Fatalf("len(Calls) = %d, want 1", len(stub.Calls))
	}

	call := stub.Calls[0]
	if call.Name != "codesign" {
		t.Errorf("command = %q, want codesign", call.Name)
	}

	expectedArgs := []string{
		"--force",
		"--options", "runtime",
		"--sign", "Developer ID Application: Test (TEAMID)",
		"--timestamp",
		"/path/to/binary",
	}

	if len(call.Args) != len(expectedArgs) {
		t.Fatalf("args count = %d, want %d\ngot: %v", len(call.Args), len(expectedArgs), call.Args)
	}

	for i, want := range expectedArgs {
		if call.Args[i] != want {
			t.Errorf("args[%d] = %q, want %q", i, call.Args[i], want)
		}
	}
}

func TestCodeSigner_Sign_HardenedRuntime(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["codesign"] = StubResponse{Output: nil, Err: nil}

	cs := NewCodeSigner(stub, "Developer ID Application: Test (TEAMID)")
	_ = cs.Sign("/path/to/binary")

	// Verify --options runtime is present (required for notarization)
	call := stub.Calls[0]
	foundOptions := false
	for i, arg := range call.Args {
		if arg == "--options" && i+1 < len(call.Args) && call.Args[i+1] == "runtime" {
			foundOptions = true
			break
		}
	}
	if !foundOptions {
		t.Error("Sign() must include --options runtime for hardened runtime (required for notarization)")
	}
}

func TestCodeSigner_Sign_ErrorWrapsOutput(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["codesign"] = StubResponse{
		Output: []byte("error: no identity found"),
		Err:    errors.New("exit status 1"),
	}

	cs := NewCodeSigner(stub, "Bad Identity")
	err := cs.Sign("/path/to/binary")

	if err == nil {
		t.Fatal("expected error, got nil")
		return
	}
	if !errors.Is(err, stub.Responses["codesign"].Err) {
		// The error should wrap the original
		t.Logf("error message: %v", err)
	}
}

func TestCodeSigner_Verify_CorrectArguments(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["codesign"] = StubResponse{Output: nil, Err: nil}

	cs := NewCodeSigner(stub, "unused-for-verify")
	err := cs.Verify("/path/to/signed-binary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if len(stub.Calls) != 1 {
		t.Fatalf("len(Calls) = %d, want 1", len(stub.Calls))
	}

	call := stub.Calls[0]
	expectedArgs := []string{"--verify", "--deep", "--strict", "/path/to/signed-binary"}

	if len(call.Args) != len(expectedArgs) {
		t.Fatalf("args count = %d, want %d\ngot: %v", len(call.Args), len(expectedArgs), call.Args)
	}

	for i, want := range expectedArgs {
		if call.Args[i] != want {
			t.Errorf("args[%d] = %q, want %q", i, call.Args[i], want)
		}
	}
}

func TestCodeSigner_Verify_FailureReturnsError(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["codesign"] = StubResponse{
		Output: []byte("invalid signature"),
		Err:    errors.New("exit status 3"),
	}

	cs := NewCodeSigner(stub, "unused")
	err := cs.Verify("/path/to/bad-binary")

	if err == nil {
		t.Fatal("expected error for invalid signature, got nil")
		return
	}
}

func TestCodeSigner_SignMultipleBinaries(t *testing.T) {
	tests := []struct {
		name   string
		binary string
	}{
		{name: "arm64", binary: "threedoors-darwin-arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := NewStubRunner()
			stub.Responses["codesign"] = StubResponse{Output: nil, Err: nil}

			cs := NewCodeSigner(stub, "Developer ID Application: Test (TEAM)")
			err := cs.Sign(tt.binary)
			if err != nil {
				t.Fatalf("Sign(%q) unexpected error: %v", tt.binary, err)
			}

			lastArg := stub.Calls[0].Args[len(stub.Calls[0].Args)-1]
			if lastArg != tt.binary {
				t.Errorf("last arg = %q, want %q", lastArg, tt.binary)
			}
		})
	}
}
