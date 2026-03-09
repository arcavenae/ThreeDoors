package core

import (
	"testing"
	"time"
)

func TestInferEnergyFromTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		hour int
		want EnergyLevel
	}{
		{"midnight", 0, EnergyLow},
		{"3am", 3, EnergyLow},
		{"5am", 5, EnergyLow},
		{"6am", 6, EnergyHigh},
		{"9am", 9, EnergyHigh},
		{"11am", 11, EnergyHigh},
		{"noon", 12, EnergyMedium},
		{"2pm", 14, EnergyMedium},
		{"4pm", 16, EnergyMedium},
		{"5pm", 17, EnergyLow},
		{"8pm", 20, EnergyLow},
		{"11pm", 23, EnergyLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := time.Date(2026, 3, 8, tt.hour, 30, 0, 0, time.UTC)
			got := InferEnergyFromTime(ts)
			if got != tt.want {
				t.Errorf("InferEnergyFromTime(hour=%d) = %q, want %q", tt.hour, got, tt.want)
			}
		})
	}
}

func TestInferEnergyFromTime_boundaries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		hour int
		min  int
		want EnergyLevel
	}{
		{"5:59am", 5, 59, EnergyLow},
		{"6:00am", 6, 0, EnergyHigh},
		{"11:59am", 11, 59, EnergyHigh},
		{"12:00pm", 12, 0, EnergyMedium},
		{"4:59pm", 16, 59, EnergyMedium},
		{"5:00pm", 17, 0, EnergyLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := time.Date(2026, 3, 8, tt.hour, tt.min, 0, 0, time.UTC)
			got := InferEnergyFromTime(ts)
			if got != tt.want {
				t.Errorf("InferEnergyFromTime(%02d:%02d) = %q, want %q", tt.hour, tt.min, got, tt.want)
			}
		})
	}
}

func TestEnergyMatchScore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		effort TaskEffort
		energy EnergyLevel
		want   float64
	}{
		// Perfect matches (1.0)
		{"high+deep-work", EffortDeepWork, EnergyHigh, 1.0},
		{"medium+medium", EffortMedium, EnergyMedium, 1.0},
		{"low+quick-win", EffortQuickWin, EnergyLow, 1.0},
		// Adjacent matches (0.5)
		{"high+medium", EffortMedium, EnergyHigh, 0.5},
		{"medium+deep-work", EffortDeepWork, EnergyMedium, 0.5},
		{"medium+quick-win", EffortQuickWin, EnergyMedium, 0.5},
		{"low+medium", EffortMedium, EnergyLow, 0.5},
		// Mismatches (0.25)
		{"high+quick-win", EffortQuickWin, EnergyHigh, 0.25},
		{"low+deep-work", EffortDeepWork, EnergyLow, 0.25},
		// No effort tag (0.75)
		{"high+empty", "", EnergyHigh, 0.75},
		{"medium+empty", "", EnergyMedium, 0.75},
		{"low+empty", "", EnergyLow, 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := &Task{Effort: tt.effort}
			got := EnergyMatchScore(task, tt.energy)
			if got != tt.want {
				t.Errorf("EnergyMatchScore(effort=%q, energy=%q) = %v, want %v",
					tt.effort, tt.energy, got, tt.want)
			}
		})
	}
}

func TestSortByEnergyMatch(t *testing.T) {
	t.Parallel()

	t.Run("sorts best match first", func(t *testing.T) {
		t.Parallel()
		tasks := []*Task{
			{ID: "mismatch", Effort: EffortQuickWin},
			{ID: "perfect", Effort: EffortDeepWork},
			{ID: "adjacent", Effort: EffortMedium},
			{ID: "noeffort", Effort: ""},
		}
		SortByEnergyMatch(tasks, EnergyHigh)

		wantOrder := []string{"perfect", "noeffort", "adjacent", "mismatch"}
		for i, id := range wantOrder {
			if tasks[i].ID != id {
				t.Errorf("position %d: got %q, want %q", i, tasks[i].ID, id)
			}
		}
	})

	t.Run("stable sort preserves order for equal scores", func(t *testing.T) {
		t.Parallel()
		tasks := []*Task{
			{ID: "a", Effort: EffortDeepWork},
			{ID: "b", Effort: EffortDeepWork},
			{ID: "c", Effort: EffortDeepWork},
		}
		SortByEnergyMatch(tasks, EnergyHigh)

		for i, id := range []string{"a", "b", "c"} {
			if tasks[i].ID != id {
				t.Errorf("stable sort broken: position %d got %q, want %q", i, tasks[i].ID, id)
			}
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		t.Parallel()
		SortByEnergyMatch(nil, EnergyHigh) // should not panic
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()
		SortByEnergyMatch([]*Task{}, EnergyHigh) // should not panic
	})
}

func TestFilterByEnergy(t *testing.T) {
	t.Parallel()

	tasks := []*Task{
		{ID: "deep", Effort: EffortDeepWork},
		{ID: "medium", Effort: EffortMedium},
		{ID: "quick", Effort: EffortQuickWin},
		{ID: "none", Effort: ""},
	}

	tests := []struct {
		name    string
		energy  EnergyLevel
		wantIDs []string
	}{
		{"high energy", EnergyHigh, []string{"deep", "none"}},
		{"medium energy", EnergyMedium, []string{"medium", "none"}},
		{"low energy", EnergyLow, []string{"quick", "none"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FilterByEnergy(tasks, tt.energy)
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("FilterByEnergy(%q) returned %d tasks, want %d", tt.energy, len(got), len(tt.wantIDs))
			}
			for i, id := range tt.wantIDs {
				if got[i].ID != id {
					t.Errorf("position %d: got %q, want %q", i, got[i].ID, id)
				}
			}
		})
	}

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()
		got := FilterByEnergy(nil, EnergyHigh)
		if got != nil {
			t.Errorf("FilterByEnergy(nil) = %v, want nil", got)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		got := FilterByEnergy([]*Task{}, EnergyHigh)
		if got != nil {
			t.Errorf("FilterByEnergy([]) = %v, want nil", got)
		}
	})
}

func TestEnergySetScore(t *testing.T) {
	t.Parallel()
	tasks := []*Task{
		{Effort: EffortDeepWork},
		{Effort: EffortMedium},
		{Effort: EffortQuickWin},
	}
	// High energy: deep=1.0, medium=0.5, quick=0.25 → 1.75
	got := EnergySetScore(tasks, EnergyHigh)
	if got != 1.75 {
		t.Errorf("EnergySetScore(high) = %v, want 1.75", got)
	}

	// Empty set
	got = EnergySetScore(nil, EnergyHigh)
	if got != 0 {
		t.Errorf("EnergySetScore(nil) = %v, want 0", got)
	}
}

func TestNextEnergyLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		current EnergyLevel
		want    EnergyLevel
	}{
		{"high to medium", EnergyHigh, EnergyMedium},
		{"medium to low", EnergyMedium, EnergyLow},
		{"low to high", EnergyLow, EnergyHigh},
		{"unknown to high", EnergyLevel(""), EnergyHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NextEnergyLevel(tt.current)
			if got != tt.want {
				t.Errorf("NextEnergyLevel(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}

func TestEnergyLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		energy EnergyLevel
		want   string
	}{
		{EnergyHigh, "High"},
		{EnergyMedium, "Medium"},
		{EnergyLow, "Low"},
		{EnergyLevel(""), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := EnergyLabel(tt.energy)
			if got != tt.want {
				t.Errorf("EnergyLabel(%q) = %q, want %q", tt.energy, got, tt.want)
			}
		})
	}
}

func TestTimeOfDayLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		hour int
		want string
	}{
		{"midnight", 0, "evening"},
		{"5am", 5, "evening"},
		{"6am", 6, "morning"},
		{"11am", 11, "morning"},
		{"noon", 12, "afternoon"},
		{"4pm", 16, "afternoon"},
		{"5pm", 17, "evening"},
		{"11pm", 23, "evening"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := time.Date(2026, 3, 9, tt.hour, 0, 0, 0, time.UTC)
			got := TimeOfDayLabel(ts)
			if got != tt.want {
				t.Errorf("TimeOfDayLabel(hour=%d) = %q, want %q", tt.hour, got, tt.want)
			}
		})
	}
}

func TestEnergyDisplayString(t *testing.T) {
	t.Parallel()
	morning := time.Date(2026, 3, 9, 9, 0, 0, 0, time.UTC)
	evening := time.Date(2026, 3, 9, 20, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		energy     EnergyLevel
		t          time.Time
		overridden bool
		want       string
	}{
		{"auto morning high", EnergyHigh, morning, false, "Energy: High (morning)"},
		{"auto evening low", EnergyLow, evening, false, "Energy: Low (evening)"},
		{"override medium", EnergyMedium, morning, true, "Energy: Medium (override)"},
		{"override high", EnergyHigh, evening, true, "Energy: High (override)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := EnergyDisplayString(tt.energy, tt.t, tt.overridden)
			if got != tt.want {
				t.Errorf("EnergyDisplayString(%q, overridden=%v) = %q, want %q",
					tt.energy, tt.overridden, got, tt.want)
			}
		})
	}
}

func TestMatchesEnergy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		effort TaskEffort
		energy EnergyLevel
		want   bool
	}{
		{"high+deep-work", EffortDeepWork, EnergyHigh, true},
		{"high+medium", EffortMedium, EnergyHigh, false},
		{"high+quick-win", EffortQuickWin, EnergyHigh, false},
		{"medium+medium", EffortMedium, EnergyMedium, true},
		{"medium+deep-work", EffortDeepWork, EnergyMedium, false},
		{"medium+quick-win", EffortQuickWin, EnergyMedium, false},
		{"low+quick-win", EffortQuickWin, EnergyLow, true},
		{"low+medium", EffortMedium, EnergyLow, false},
		{"low+deep-work", EffortDeepWork, EnergyLow, false},
		{"high+empty", "", EnergyHigh, true},
		{"medium+empty", "", EnergyMedium, true},
		{"low+empty", "", EnergyLow, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := &Task{Effort: tt.effort}
			got := MatchesEnergy(task, tt.energy)
			if got != tt.want {
				t.Errorf("MatchesEnergy(effort=%q, energy=%q) = %v, want %v",
					tt.effort, tt.energy, got, tt.want)
			}
		})
	}
}
