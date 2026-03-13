package cli

import (
	"testing"
)

func TestNewDocAuditCmd_Structure(t *testing.T) {
	t.Parallel()

	cmd := newDocAuditCmd()

	if cmd.Use != "doc-audit" {
		t.Errorf("Use = %q, want %q", cmd.Use, "doc-audit")
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	for _, name := range []string{"root", "jsonl"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag %q", name)
		}
	}
}

func TestDocAuditCmd_Registered(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "doc-audit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("doc-audit command should be registered in root command")
	}
}
