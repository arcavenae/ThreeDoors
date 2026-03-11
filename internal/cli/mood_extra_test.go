package cli

import (
	"testing"
)

func TestNewMoodSetCmd_ArgsValidation(t *testing.T) {
	t.Parallel()

	cmd := newMoodSetCmd()

	// Requires 1-2 args (RangeArgs(1, 2))
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error for 0 args")
	}
	if err := cmd.Args(cmd, []string{"focused"}); err != nil {
		t.Errorf("expected 1 arg to be valid, got: %v", err)
	}
	if err := cmd.Args(cmd, []string{"custom", "feeling great"}); err != nil {
		t.Errorf("expected 2 args to be valid, got: %v", err)
	}
	if err := cmd.Args(cmd, []string{"a", "b", "c"}); err == nil {
		t.Error("expected error for 3 args")
	}
}

func TestNewMoodHistoryCmd_ArgsValidation(t *testing.T) {
	t.Parallel()

	cmd := newMoodHistoryCmd()
	if cmd.Use != "history" {
		t.Errorf("Use = %q, want %q", cmd.Use, "history")
	}
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("expected error for extra args")
	}
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("expected 0 args to be valid, got: %v", err)
	}
}

func TestIsValidMood_AllCases(t *testing.T) {
	t.Parallel()

	// Test uppercase versions
	for _, mood := range validMoods {
		upper := string(mood[0]-32) + mood[1:]
		if !isValidMood(upper) {
			t.Errorf("isValidMood(%q) should be true (case insensitive)", upper)
		}
	}
}
