package dist

import (
	"errors"
	"testing"
)

func TestNotarizer_Submit_CorrectArguments(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["xcrun"] = StubResponse{Output: []byte("Successfully submitted"), Err: nil}

	n := NewNotarizer(stub, "user@example.com", "app-specific-pwd", "TEAMID123")
	err := n.Submit("binary.zip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if len(stub.Calls) != 1 {
		t.Fatalf("len(Calls) = %d, want 1", len(stub.Calls))
	}

	call := stub.Calls[0]
	if call.Name != "xcrun" {
		t.Errorf("command = %q, want xcrun", call.Name)
	}

	expectedArgs := []string{
		"notarytool", "submit",
		"binary.zip",
		"--apple-id", "user@example.com",
		"--password", "app-specific-pwd",
		"--team-id", "TEAMID123",
		"--wait",
		"--timeout", "900",
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

func TestNotarizer_Submit_CustomTimeout(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["xcrun"] = StubResponse{Output: nil, Err: nil}

	n := NewNotarizer(stub, "user@example.com", "pwd", "TEAM")
	n.Timeout = 600
	_ = n.Submit("file.zip")

	call := stub.Calls[0]
	// Find --timeout arg
	for i, arg := range call.Args {
		if arg == "--timeout" && i+1 < len(call.Args) {
			if call.Args[i+1] != "600" {
				t.Errorf("timeout = %q, want %q", call.Args[i+1], "600")
			}
			return
		}
	}
	t.Error("--timeout flag not found in arguments")
}

func TestNotarizer_Submit_ErrorWrapsOutput(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["xcrun"] = StubResponse{
		Output: []byte("Error: invalid credentials"),
		Err:    errors.New("exit status 1"),
	}

	n := NewNotarizer(stub, "bad@email.com", "wrong-pwd", "TEAM")
	err := n.Submit("binary.zip")

	if err == nil {
		t.Fatal("expected error, got nil")
		return
	}
}

func TestNotarizer_Staple_CorrectArguments(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["xcrun"] = StubResponse{Output: nil, Err: nil}

	n := NewNotarizer(stub, "", "", "")
	err := n.Staple("threedoors.pkg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	call := stub.Calls[0]
	if call.Name != "xcrun" {
		t.Errorf("command = %q, want xcrun", call.Name)
	}

	expectedArgs := []string{"stapler", "staple", "threedoors.pkg"}
	if len(call.Args) != len(expectedArgs) {
		t.Fatalf("args count = %d, want %d\ngot: %v", len(call.Args), len(expectedArgs), call.Args)
	}

	for i, want := range expectedArgs {
		if call.Args[i] != want {
			t.Errorf("args[%d] = %q, want %q", i, call.Args[i], want)
		}
	}
}

func TestNotarizer_Assess_CorrectArguments(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["spctl"] = StubResponse{Output: nil, Err: nil}

	n := NewNotarizer(stub, "", "", "")
	err := n.Assess("threedoors-darwin-arm64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	call := stub.Calls[0]
	if call.Name != "spctl" {
		t.Errorf("command = %q, want spctl", call.Name)
	}

	expectedArgs := []string{"--assess", "--type", "execute", "threedoors-darwin-arm64"}
	for i, want := range expectedArgs {
		if call.Args[i] != want {
			t.Errorf("args[%d] = %q, want %q", i, call.Args[i], want)
		}
	}
}

func TestNotarizer_Assess_FailureReturnsError(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["spctl"] = StubResponse{
		Output: []byte("rejected"),
		Err:    errors.New("exit status 3"),
	}

	n := NewNotarizer(stub, "", "", "")
	err := n.Assess("unsigned-binary")

	if err == nil {
		t.Fatal("expected error for rejected binary, got nil")
		return
	}
}

func TestNotarizer_DefaultTimeout(t *testing.T) {
	n := NewNotarizer(NewStubRunner(), "", "", "")
	if n.Timeout != 900 {
		t.Errorf("default Timeout = %d, want 900", n.Timeout)
	}
}
