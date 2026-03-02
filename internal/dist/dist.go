package dist

import (
	"fmt"
	"os/exec"
)

// CommandRunner abstracts command execution for testability.
// Production code uses RealRunner; tests use StubRunner.
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

// RealRunner executes actual OS commands.
type RealRunner struct{}

// Run executes a command and returns combined stdout+stderr output.
func (r RealRunner) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// StubCall records a single command invocation for test assertions.
type StubCall struct {
	Name string
	Args []string
}

// StubResponse defines the output and error for a stubbed command.
type StubResponse struct {
	Output []byte
	Err    error
}

// StubRunner records calls and returns pre-configured responses.
// Keys in Responses are the command name (first matching response is used).
type StubRunner struct {
	Calls     []StubCall
	Responses map[string]StubResponse
}

// NewStubRunner creates a StubRunner with an empty response map.
func NewStubRunner() *StubRunner {
	return &StubRunner{
		Responses: make(map[string]StubResponse),
	}
}

// Run records the call and returns the configured response (or error if not found).
func (s *StubRunner) Run(name string, args ...string) ([]byte, error) {
	s.Calls = append(s.Calls, StubCall{Name: name, Args: args})

	if resp, ok := s.Responses[name]; ok {
		return resp.Output, resp.Err
	}
	return nil, fmt.Errorf("stub: no response configured for %q", name)
}
