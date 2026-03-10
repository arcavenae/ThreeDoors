package dist

import (
	"errors"
	"testing"
)

func TestStubRunner_RecordsCalls(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["echo"] = StubResponse{Output: []byte("hello"), Err: nil}

	out, err := stub.Run("echo", "hello", "world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if string(out) != "hello" {
		t.Errorf("output = %q, want %q", out, "hello")
	}
	if len(stub.Calls) != 1 {
		t.Fatalf("len(Calls) = %d, want 1", len(stub.Calls))
	}
	if stub.Calls[0].Name != "echo" {
		t.Errorf("Calls[0].Name = %q, want %q", stub.Calls[0].Name, "echo")
	}
	if len(stub.Calls[0].Args) != 2 || stub.Calls[0].Args[0] != "hello" {
		t.Errorf("Calls[0].Args = %v, want [hello world]", stub.Calls[0].Args)
	}
}

func TestStubRunner_ReturnsError(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["fail"] = StubResponse{Output: []byte("oops"), Err: errors.New("exit 1")}

	out, err := stub.Run("fail")

	if err == nil {
		t.Fatal("expected error, got nil")
		return
	}
	if string(out) != "oops" {
		t.Errorf("output = %q, want %q", out, "oops")
	}
}

func TestStubRunner_UnconfiguredCommandErrors(t *testing.T) {
	stub := NewStubRunner()

	_, err := stub.Run("unknown-cmd")

	if err == nil {
		t.Fatal("expected error for unconfigured command, got nil")
		return
	}
}

func TestStubRunner_MultipleCallsTracked(t *testing.T) {
	stub := NewStubRunner()
	stub.Responses["cmd1"] = StubResponse{Output: []byte("a"), Err: nil}
	stub.Responses["cmd2"] = StubResponse{Output: []byte("b"), Err: nil}

	_, _ = stub.Run("cmd1", "--flag")
	_, _ = stub.Run("cmd2", "-v")

	if len(stub.Calls) != 2 {
		t.Fatalf("len(Calls) = %d, want 2", len(stub.Calls))
	}
	if stub.Calls[0].Name != "cmd1" {
		t.Errorf("Calls[0].Name = %q, want cmd1", stub.Calls[0].Name)
	}
	if stub.Calls[1].Name != "cmd2" {
		t.Errorf("Calls[1].Name = %q, want cmd2", stub.Calls[1].Name)
	}
}
