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
