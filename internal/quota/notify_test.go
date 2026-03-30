package quota

import "testing"

func TestNotifier_SendEmpty(t *testing.T) {
	t.Parallel()

	n := NewNotifier(true)
	if err := n.Send(""); err != nil {
		t.Errorf("Send empty message should return nil, got: %v", err)
	}
}

func TestNewNotifier(t *testing.T) {
	t.Parallel()

	t.Run("cli mode", func(t *testing.T) {
		t.Parallel()
		n := NewNotifier(true)
		if !n.CLIMode {
			t.Error("expected CLIMode true")
		}
	})

	t.Run("multiclaude mode", func(t *testing.T) {
		t.Parallel()
		n := NewNotifier(false)
		if n.CLIMode {
			t.Error("expected CLIMode false")
		}
	})
}
