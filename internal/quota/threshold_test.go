package quota

import (
	"strings"
	"testing"
	"time"
)

func TestWindowUsage_UsagePercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		total      int64
		budget     int64
		wantApprox float64
	}{
		{"zero budget", 50000, 0, 0},
		{"zero usage", 0, 100000, 0},
		{"50 percent", 50000, 100000, 50.0},
		{"100 percent", 100000, 100000, 100.0},
		{"over budget", 120000, 100000, 120.0},
		{"typical max 5x", 61600, 88000, 70.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u := WindowUsage{TotalTokens: tt.total, PlanBudget: tt.budget}
			got := u.UsagePercent()
			if diff := got - tt.wantApprox; diff > 0.1 || diff < -0.1 {
				t.Errorf("UsagePercent() = %.2f, want ~%.2f", got, tt.wantApprox)
			}
		})
	}
}

func TestWindowUsage_RemainingTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		total  int64
		budget int64
		want   int64
	}{
		{"under budget", 50000, 100000, 50000},
		{"at budget", 100000, 100000, 0},
		{"over budget", 120000, 100000, 0},
		{"zero budget", 50000, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u := WindowUsage{TotalTokens: tt.total, PlanBudget: tt.budget}
			got := u.RemainingTokens()
			if got != tt.want {
				t.Errorf("RemainingTokens() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestWindowUsage_TimeUntilReset(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		windowEnd time.Time
		want      time.Duration
	}{
		{"2 hours remaining", now.Add(2 * time.Hour), 2 * time.Hour},
		{"window expired", now.Add(-1 * time.Hour), 0},
		{"exactly now", now, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u := WindowUsage{WindowEnd: tt.windowEnd}
			got := u.TimeUntilReset(now)
			if got != tt.want {
				t.Errorf("TimeUntilReset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPeakHour(t *testing.T) {
	t.Parallel()

	cfg := DefaultThresholdConfig()
	pt, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("load PT timezone: %v", err)
	}

	tests := []struct {
		name     string
		ptHour   int
		wantPeak bool
	}{
		{"4am PT - before peak", 4, false},
		{"5am PT - start of peak", 5, true},
		{"8am PT - mid peak", 8, true},
		{"10am PT - late peak", 10, true},
		{"11am PT - end of peak", 11, false},
		{"15pm PT - afternoon", 15, false},
		{"0am PT - midnight", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create time in PT timezone
			ptTime := time.Date(2026, 3, 30, tt.ptHour, 30, 0, 0, pt)
			got := cfg.IsPeakHour(ptTime)
			if got != tt.wantPeak {
				t.Errorf("IsPeakHour(%d:30 PT) = %v, want %v", tt.ptHour, got, tt.wantPeak)
			}
		})
	}
}

func TestEffectiveThreshold(t *testing.T) {
	t.Parallel()

	cfg := DefaultThresholdConfig()
	pt, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("load PT timezone: %v", err)
	}

	tier70 := cfg.Tiers[0] // 70%

	tests := []struct {
		name string
		hour int
		want float64
	}{
		{"off-peak no shift", 15, 70.0},
		{"peak shifted", 8, 56.0}, // 70 * 0.8 = 56
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ptTime := time.Date(2026, 3, 30, tt.hour, 0, 0, 0, pt)
			got := cfg.EffectiveThreshold(tier70, ptTime)
			if diff := got - tt.want; diff > 0.01 || diff < -0.01 {
				t.Errorf("EffectiveThreshold() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}

func TestEvaluate_AllTiers(t *testing.T) {
	t.Parallel()

	cfg := DefaultThresholdConfig()
	// Use off-peak time so thresholds are not shifted
	offPeak := time.Date(2026, 3, 30, 20, 0, 0, 0, time.UTC) // 12pm PT

	tests := []struct {
		name          string
		usagePercent  float64
		wantTriggered bool
		wantLabel     string
	}{
		{"below all thresholds", 50, false, ""},
		{"just below 70", 69.9, false, ""},
		{"exactly at 70 - green", 70.0, true, "green"},
		{"just above 70", 70.1, true, "green"},
		{"at 80 - yellow", 80.0, true, "yellow"},
		{"between 80 and 90", 85.0, true, "yellow"},
		{"at 90 - orange", 90.0, true, "orange"},
		{"between 90 and 95", 92.0, true, "orange"},
		{"at 95 - red", 95.0, true, "red"},
		{"above 95", 99.0, true, "red"},
		{"at 100 - over budget", 100.0, true, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			usage := WindowUsage{
				TotalTokens: int64(tt.usagePercent * 1000),
				PlanBudget:  100000,
				WindowEnd:   offPeak.Add(2 * time.Hour),
			}
			result := cfg.Evaluate(usage, offPeak)
			if result.Triggered != tt.wantTriggered {
				t.Errorf("Triggered = %v, want %v", result.Triggered, tt.wantTriggered)
			}
			if tt.wantTriggered {
				if result.ActiveTier == nil {
					t.Fatal("ActiveTier is nil but Triggered is true")
				}
				if result.ActiveTier.Label != tt.wantLabel {
					t.Errorf("ActiveTier.Label = %q, want %q", result.ActiveTier.Label, tt.wantLabel)
				}
			}
		})
	}
}

func TestEvaluate_PeakHoursShiftThresholds(t *testing.T) {
	t.Parallel()

	cfg := DefaultThresholdConfig()
	pt, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("load PT timezone: %v", err)
	}

	// During peak, 70% threshold shifts to 56% (70 * 0.8)
	peakTime := time.Date(2026, 3, 30, 8, 0, 0, 0, pt) // 8am PT

	tests := []struct {
		name          string
		usagePercent  float64
		wantTriggered bool
		wantLabel     string
	}{
		{"55% - below shifted 70 threshold", 55, false, ""},
		{"56% - at shifted 70 threshold (green)", 56.0, true, "green"},
		{"60% - above shifted green, below shifted yellow", 60.0, true, "green"},
		{"64% - at shifted 80 threshold (yellow)", 64.0, true, "yellow"},
		{"72% - at shifted 90 threshold (orange)", 72.0, true, "orange"},
		{"76% - at shifted 95 threshold (red)", 76.0, true, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			usage := WindowUsage{
				TotalTokens: int64(tt.usagePercent * 1000),
				PlanBudget:  100000,
				WindowEnd:   peakTime.Add(3 * time.Hour),
			}
			result := cfg.Evaluate(usage, peakTime)
			if result.Triggered != tt.wantTriggered {
				t.Errorf("Triggered = %v, want %v (usage=%.1f%%)", result.Triggered, tt.wantTriggered, tt.usagePercent)
			}
			if tt.wantTriggered {
				if result.ActiveTier == nil {
					t.Fatal("ActiveTier is nil but Triggered is true")
				}
				if result.ActiveTier.Label != tt.wantLabel {
					t.Errorf("ActiveTier.Label = %q, want %q", result.ActiveTier.Label, tt.wantLabel)
				}
			}
			if !result.IsPeak {
				t.Error("IsPeak should be true during peak hours")
			}
		})
	}
}

func TestEvaluate_NeverBlocks(t *testing.T) {
	t.Parallel()

	cfg := DefaultThresholdConfig()
	now := time.Date(2026, 3, 30, 20, 0, 0, 0, time.UTC)

	// Even at 200% usage, the engine only returns advisory data
	usage := WindowUsage{
		TotalTokens: 200000,
		PlanBudget:  100000,
		WindowEnd:   now.Add(1 * time.Hour),
	}
	result := cfg.Evaluate(usage, now)
	if !result.Triggered {
		t.Error("Expected triggered at 200% usage")
	}
	if result.ActiveTier == nil {
		t.Fatal("ActiveTier should not be nil")
	}
	if result.ActiveTier.Label != "red" {
		t.Errorf("Expected red tier at 200%%, got %q", result.ActiveTier.Label)
	}
	// Verify the result is purely informational — no blocking fields
	// The struct has no blocking/throttling fields by design (AC4)
}

func TestFormatWarning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		result    EvaluationResult
		contains  []string
		wantEmpty bool
	}{
		{
			name:      "no trigger",
			result:    EvaluationResult{Triggered: false},
			wantEmpty: true,
		},
		{
			name: "green tier off-peak",
			result: EvaluationResult{
				Triggered:        true,
				ActiveTier:       &Tier{Label: "green", Suggestion: "Consider monitoring closely"},
				EffectivePercent: 70.0,
				UsagePercent:     72.5,
				IsPeak:           false,
				RemainingTokens:  27500,
				TimeUntilReset:   2*time.Hour + 15*time.Minute,
			},
			contains: []string{
				"QUOTA_WARNING",
				"[green]",
				"72.5%",
				"70.0%",
				"27.5K",
				"2h 15m",
				"Consider monitoring closely",
			},
		},
		{
			name: "red tier peak hours",
			result: EvaluationResult{
				Triggered:        true,
				ActiveTier:       &Tier{Label: "red", Suggestion: "Critical — consider pausing all but P0 work"},
				EffectivePercent: 76.0,
				UsagePercent:     96.0,
				IsPeak:           true,
				RemainingTokens:  4000,
				TimeUntilReset:   30 * time.Minute,
			},
			contains: []string{
				"QUOTA_WARNING",
				"[red]",
				"[PEAK HOURS]",
				"96.0%",
				"76.0%",
				"4.0K",
				"30m",
				"Critical",
			},
		},
		{
			name: "large remaining tokens formatted as M",
			result: EvaluationResult{
				Triggered:        true,
				ActiveTier:       &Tier{Label: "yellow", Suggestion: "Consider reducing heartbeat frequency"},
				EffectivePercent: 80.0,
				UsagePercent:     81.0,
				RemainingTokens:  1_500_000,
				TimeUntilReset:   4 * time.Hour,
			},
			contains: []string{"1.5M"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatWarning(tt.result)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("expected empty string, got %q", got)
				}
				return
			}
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("warning missing %q:\n  got: %s", s, got)
				}
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "0m"},
		{"negative", -5 * time.Minute, "0m"},
		{"minutes only", 45 * time.Minute, "45m"},
		{"hours and minutes", 2*time.Hour + 30*time.Minute, "2h 30m"},
		{"hours only", 3 * time.Hour, "3h 0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFormatTokenCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tokens int64
		want   string
	}{
		{"small", 500, "500"},
		{"thousands", 27500, "27.5K"},
		{"millions", 1500000, "1.5M"},
		{"exact thousand", 1000, "1.0K"},
		{"exact million", 1000000, "1.0M"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatTokenCount(tt.tokens)
			if got != tt.want {
				t.Errorf("formatTokenCount(%d) = %q, want %q", tt.tokens, got, tt.want)
			}
		})
	}
}

func TestDefaultThresholdConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultThresholdConfig()

	if len(cfg.Tiers) != 4 {
		t.Fatalf("expected 4 tiers, got %d", len(cfg.Tiers))
	}

	// Verify tiers are ordered ascending
	for i := 1; i < len(cfg.Tiers); i++ {
		if cfg.Tiers[i].Percent <= cfg.Tiers[i-1].Percent {
			t.Errorf("tiers not ascending: tier[%d]=%.1f <= tier[%d]=%.1f",
				i, cfg.Tiers[i].Percent, i-1, cfg.Tiers[i-1].Percent)
		}
	}

	// Verify defaults
	expectedLabels := []string{"green", "yellow", "orange", "red"}
	for i, label := range expectedLabels {
		if cfg.Tiers[i].Label != label {
			t.Errorf("tier[%d].Label = %q, want %q", i, cfg.Tiers[i].Label, label)
		}
	}

	if cfg.PeakShiftFactor != 0.8 {
		t.Errorf("PeakShiftFactor = %f, want 0.8", cfg.PeakShiftFactor)
	}
}
