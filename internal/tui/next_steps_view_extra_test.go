package tui

import (
	"testing"
)

func TestNextStepsView_DefaultContext(t *testing.T) {
	t.Parallel()
	pool := makePool("t1", "t2")
	nv := NewNextStepsView("unknown-context", pool, nil)
	if nv.header != "What would you like to do next?" {
		t.Errorf("expected default header, got %q", nv.header)
	}
	// defaultOptions returns 3 options: doors, add, mood
	if len(nv.options) != 3 {
		t.Errorf("expected 3 default options, got %d", len(nv.options))
	}
	actions := map[string]bool{}
	for _, opt := range nv.options {
		actions[opt.Action] = true
	}
	for _, expected := range []string{"doors", "add", "mood"} {
		if !actions[expected] {
			t.Errorf("expected default option %q", expected)
		}
	}
}
