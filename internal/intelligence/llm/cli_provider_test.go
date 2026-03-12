package llm

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockCLIRunner implements CLIRunner for testing.
type mockCLIRunner struct {
	stdout string
	stderr string
	err    error

	// captured invocation for assertions
	capturedStdin   string
	capturedName    string
	capturedArgs    []string
	capturedContext context.Context
}

func (m *mockCLIRunner) RunWithStdin(ctx context.Context, stdin string, name string, args ...string) (string, string, error) {
	m.capturedContext = ctx
	m.capturedStdin = stdin
	m.capturedName = name
	m.capturedArgs = args
	return m.stdout, m.stderr, m.err
}

func TestCLIProviderImplementsLLMBackend(t *testing.T) {
	t.Parallel()
	var _ LLMBackend = (*CLIProvider)(nil)
}

func TestCLIProviderName(t *testing.T) {
	t.Parallel()

	p := NewCLIProvider(CLISpec{Name: "test-cli"}, &mockCLIRunner{})
	if got := p.Name(); got != "test-cli" {
		t.Errorf("Name() = %q, want %q", got, "test-cli")
	}
}

func TestCLIProviderComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		spec      CLISpec
		prompt    string
		stdout    string
		stderr    string
		runErr    error
		want      string
		wantErr   bool
		wantErrIs error
		wantStdin string
		wantArgs  []string
	}{
		{
			name: "stdin input method",
			spec: CLISpec{
				Name:        "test",
				Command:     "test-cli",
				BaseArgs:    []string{"--print"},
				InputMethod: InputStdin,
			},
			prompt:    "hello",
			stdout:    "response text\n",
			want:      "response text",
			wantStdin: "hello",
			wantArgs:  []string{"--print"},
		},
		{
			name: "arg input method",
			spec: CLISpec{
				Name:        "test",
				Command:     "test-cli",
				BaseArgs:    []string{"run", "model"},
				InputMethod: InputArg,
			},
			prompt:    "hello",
			stdout:    "response text\n",
			want:      "response text",
			wantStdin: "",
			wantArgs:  []string{"run", "model", "hello"},
		},
		{
			name: "with system prompt and output format",
			spec: CLISpec{
				Name:     "test",
				Command:  "test-cli",
				BaseArgs: []string{"--print"},
				SystemPrompt: ArgTemplate{
					Flag:    "--system-prompt",
					Value:   "Be helpful.",
					Enabled: true,
				},
				OutputFormat: ArgTemplate{
					Flag:    "--output-format",
					Value:   "json",
					Enabled: true,
				},
				InputMethod: InputStdin,
			},
			prompt:    "hello",
			stdout:    "response\n",
			want:      "response",
			wantStdin: "hello",
			wantArgs:  []string{"--print", "--system-prompt", "Be helpful.", "--output-format", "json"},
		},
		{
			name:      "empty prompt returns ErrEmptyPrompt",
			spec:      CLISpec{Name: "test", Command: "test-cli"},
			prompt:    "",
			wantErr:   true,
			wantErrIs: ErrEmptyPrompt,
		},
		{
			name:      "empty response returns ErrEmptyResponse",
			spec:      CLISpec{Name: "test", Command: "test-cli"},
			prompt:    "hello",
			stdout:    "   \n  ",
			wantErr:   true,
			wantErrIs: ErrEmptyResponse,
		},
		{
			name:    "non-zero exit wraps stderr",
			spec:    CLISpec{Name: "test", Command: "test-cli"},
			prompt:  "hello",
			stderr:  "command not found",
			runErr:  errors.New("exit status 1"),
			wantErr: true,
		},
		{
			name: "disabled arg templates are excluded",
			spec: CLISpec{
				Name:    "test",
				Command: "test-cli",
				SystemPrompt: ArgTemplate{
					Flag:    "--system-prompt",
					Value:   "ignored",
					Enabled: false,
				},
				OutputFormat: ArgTemplate{
					Flag:    "--output-format",
					Value:   "ignored",
					Enabled: false,
				},
				InputMethod: InputStdin,
			},
			prompt:    "hello",
			stdout:    "ok\n",
			want:      "ok",
			wantStdin: "hello",
			wantArgs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockCLIRunner{
				stdout: tt.stdout,
				stderr: tt.stderr,
				err:    tt.runErr,
			}
			p := NewCLIProvider(tt.spec, runner)

			got, err := p.Complete(context.Background(), tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("Complete() error = %v, want errors.Is %v", err, tt.wantErrIs)
			}
			if got != tt.want {
				t.Errorf("Complete() = %q, want %q", got, tt.want)
			}
			if tt.wantStdin != "" && runner.capturedStdin != tt.wantStdin {
				t.Errorf("stdin = %q, want %q", runner.capturedStdin, tt.wantStdin)
			}
			if tt.wantArgs != nil {
				if len(runner.capturedArgs) != len(tt.wantArgs) {
					t.Fatalf("args = %v, want %v", runner.capturedArgs, tt.wantArgs)
				}
				for i, arg := range tt.wantArgs {
					if runner.capturedArgs[i] != arg {
						t.Errorf("args[%d] = %q, want %q", i, runner.capturedArgs[i], arg)
					}
				}
			}
		})
	}
}

func TestCLIProviderCompleteTimeout(t *testing.T) {
	t.Parallel()

	runner := &mockCLIRunner{}
	spec := CLISpec{
		Name:    "test",
		Command: "test-cli",
		Timeout: 5 * time.Second,
	}
	p := NewCLIProvider(spec, runner)

	_, _ = p.Complete(context.Background(), "hello")

	// Verify the context passed to the runner has a deadline
	if runner.capturedContext == nil {
		t.Fatal("runner was not called")
	}
	deadline, ok := runner.capturedContext.Deadline()
	if !ok {
		t.Fatal("context should have a deadline")
	}
	// Deadline should be roughly 5 seconds from now (with some tolerance)
	remaining := time.Until(deadline)
	if remaining > 6*time.Second || remaining < 0 {
		t.Errorf("deadline remaining = %v, want ~5s", remaining)
	}
}

func TestCLIProviderDefaultTimeout(t *testing.T) {
	t.Parallel()

	p := NewCLIProvider(CLISpec{Name: "test", Command: "test-cli"}, &mockCLIRunner{})
	if p.spec.Timeout != DefaultCLITimeout {
		t.Errorf("timeout = %v, want %v", p.spec.Timeout, DefaultCLITimeout)
	}
}

func TestCLIProviderDefaultParser(t *testing.T) {
	t.Parallel()

	p := NewCLIProvider(CLISpec{Name: "test", Command: "test-cli"}, &mockCLIRunner{})
	if _, ok := p.spec.Parser.(PlainTextParser); !ok {
		t.Errorf("parser = %T, want PlainTextParser", p.spec.Parser)
	}
}

func TestCLIProviderNonZeroExitIncludesStderr(t *testing.T) {
	t.Parallel()

	runner := &mockCLIRunner{
		stderr: "error: model not loaded",
		err:    errors.New("exit status 1"),
	}
	p := NewCLIProvider(CLISpec{Name: "myctl", Command: "myctl"}, runner)

	_, err := p.Complete(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error")
	}

	errMsg := err.Error()
	if !containsAll(errMsg, "myctl", "error: model not loaded") {
		t.Errorf("error = %q, want it to contain cli name and stderr", errMsg)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		found := false
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestBuildArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		spec   CLISpec
		prompt string
		want   []string
	}{
		{
			name: "base args only with stdin",
			spec: CLISpec{
				BaseArgs:    []string{"--print"},
				InputMethod: InputStdin,
			},
			prompt: "hello",
			want:   []string{"--print"},
		},
		{
			name: "base args with arg input",
			spec: CLISpec{
				BaseArgs:    []string{"run", "model"},
				InputMethod: InputArg,
			},
			prompt: "hello",
			want:   []string{"run", "model", "hello"},
		},
		{
			name: "all enabled templates",
			spec: CLISpec{
				BaseArgs: []string{"--print"},
				SystemPrompt: ArgTemplate{
					Flag:    "--system",
					Value:   "Be concise.",
					Enabled: true,
				},
				OutputFormat: ArgTemplate{
					Flag:    "--output-format",
					Value:   "json",
					Enabled: true,
				},
				InputMethod: InputStdin,
			},
			prompt: "hello",
			want:   []string{"--print", "--system", "Be concise.", "--output-format", "json"},
		},
		{
			name: "no base args, stdin",
			spec: CLISpec{
				InputMethod: InputStdin,
			},
			prompt: "hello",
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &CLIProvider{spec: tt.spec}
			got := p.buildArgs(tt.prompt)
			if len(got) != len(tt.want) {
				t.Fatalf("buildArgs() = %v (len %d), want %v (len %d)", got, len(got), tt.want, len(tt.want))
			}
			for i, arg := range tt.want {
				if got[i] != arg {
					t.Errorf("buildArgs()[%d] = %q, want %q", i, got[i], arg)
				}
			}
		})
	}
}
