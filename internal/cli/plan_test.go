package cli

import (
	"bytes"
	"testing"
)

func TestPlanCmd_Help(t *testing.T) {
	t.Parallel()

	cmd := newPlanCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("plan --help should not error: %v", err)
		return
	}

	output := buf.String()
	if output == "" {
		t.Error("plan --help should produce output")
	}
}

func TestPlanCmd_Registered(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	var found bool
	for _, cmd := range root.Commands() {
		if cmd.Name() == "plan" {
			found = true
			break
		}
	}
	if !found {
		t.Error("plan command should be registered in root")
	}
}

func TestPlanCmd_InKnownSubcommands(t *testing.T) {
	t.Parallel()

	known := KnownSubcommands()
	var found bool
	for _, name := range known {
		if name == "plan" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'plan' should be in KnownSubcommands")
	}
}

func TestIsPlanCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"plan command", []string{"threedoors", "plan"}, true},
		{"other command", []string{"threedoors", "doors"}, false},
		{"no args", []string{"threedoors"}, false},
		{"empty", nil, false},
		{"empty slice", []string{}, false},
		{"single element", []string{"threedoors"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetPlanCommandArgs(tt.args)
			got := IsPlanCommand()
			if got != tt.want {
				t.Errorf("IsPlanCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetPlanCommandArgs(t *testing.T) {
	original := planCommandArgs
	t.Cleanup(func() { planCommandArgs = original })

	args := []string{"a", "b", "c"}
	SetPlanCommandArgs(args)

	if len(planCommandArgs) != 3 {
		t.Fatalf("planCommandArgs has %d elements, want 3", len(planCommandArgs))
	}
	if planCommandArgs[0] != "a" {
		t.Errorf("planCommandArgs[0] = %q, want %q", planCommandArgs[0], "a")
	}
}

func TestNewPlanCmd_Structure(t *testing.T) {
	t.Parallel()

	cmd := newPlanCmd()
	if cmd.Use != "plan" {
		t.Errorf("Use = %q, want %q", cmd.Use, "plan")
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestNewPlanCmd_RunE_NoError(t *testing.T) {
	t.Parallel()

	cmd := newPlanCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("RunE() returned error: %v", err)
	}
}
