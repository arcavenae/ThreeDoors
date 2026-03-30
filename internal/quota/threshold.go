package quota

import (
	"fmt"
	"time"
)

// Tier represents a warning threshold tier with its label and suggested action.
type Tier struct {
	Percent    float64 // Threshold percentage (e.g., 70.0)
	Label      string  // Human-readable label (e.g., "yellow")
	Suggestion string  // Advisory action suggestion
}

// ThresholdConfig holds the warning engine configuration.
type ThresholdConfig struct {
	// Tiers are the warning thresholds, ordered from lowest to highest.
	Tiers []Tier

	// PeakStartHour is the start of peak hours in PT (default: 5 = 05:00).
	PeakStartHour int
	// PeakEndHour is the end of peak hours in PT (default: 11 = 11:00).
	PeakEndHour int

	// PeakShiftFactor is the multiplier for threshold reduction during peak.
	// Default: 0.8 (thresholds shift ~20% lower during peak hours).
	PeakShiftFactor float64

	// NotifyViaCLI controls notification delivery method.
	// If true, warnings go to stdout. If false, via multiclaude message send.
	NotifyViaCLI bool
}

// DefaultThresholdConfig returns the default 4-tier warning configuration.
func DefaultThresholdConfig() ThresholdConfig {
	return ThresholdConfig{
		Tiers: []Tier{
			{Percent: 70, Label: "green", Suggestion: "Consider monitoring closely"},
			{Percent: 80, Label: "yellow", Suggestion: "Consider reducing heartbeat frequency"},
			{Percent: 90, Label: "orange", Suggestion: "Consider pausing non-critical agents"},
			{Percent: 95, Label: "red", Suggestion: "Critical — consider pausing all but P0 work"},
		},
		PeakStartHour:   5,
		PeakEndHour:     11,
		PeakShiftFactor: 0.8,
		NotifyViaCLI:    false,
	}
}

// ptLocation is the America/Los_Angeles timezone for peak hour detection.
var ptLocation *time.Location

func init() {
	var err error
	ptLocation, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		// Fallback: use fixed offset UTC-8 (PST). This is imprecise during
		// PDT but acceptable as a last resort on systems without tzdata.
		ptLocation = time.FixedZone("PST", -8*60*60)
	}
}

// IsPeakHour returns true if the given time falls within Anthropic peak hours
// (default 05:00-11:00 PT).
func (c ThresholdConfig) IsPeakHour(t time.Time) bool {
	pt := t.In(ptLocation)
	hour := pt.Hour()
	return hour >= c.PeakStartHour && hour < c.PeakEndHour
}

// EffectiveThreshold returns the adjusted threshold percentage, accounting for
// peak-hour shifts. During peak hours, thresholds are reduced by PeakShiftFactor
// (e.g., 70% becomes 56% with a 0.8 factor).
func (c ThresholdConfig) EffectiveThreshold(tier Tier, now time.Time) float64 {
	if c.IsPeakHour(now) {
		return tier.Percent * c.PeakShiftFactor
	}
	return tier.Percent
}

// EvaluationResult holds the outcome of a threshold evaluation.
type EvaluationResult struct {
	// Triggered is true if any threshold was exceeded.
	Triggered bool
	// ActiveTier is the highest tier that was exceeded. Nil if none triggered.
	ActiveTier *Tier
	// EffectivePercent is the adjusted threshold that was exceeded.
	EffectivePercent float64
	// UsagePercent is the actual usage percentage.
	UsagePercent float64
	// IsPeak indicates whether the evaluation occurred during peak hours.
	IsPeak bool
	// RemainingTokens is the estimated tokens remaining.
	RemainingTokens int64
	// TimeUntilReset is the duration until the usage window resets.
	TimeUntilReset time.Duration
}

// Evaluate checks the current usage against configured thresholds and returns
// the highest triggered tier. This function is ADVISORY ONLY — it never blocks,
// throttles, or kills anything.
func (c ThresholdConfig) Evaluate(usage WindowUsage, now time.Time) EvaluationResult {
	usagePct := usage.UsagePercent()
	isPeak := c.IsPeakHour(now)

	result := EvaluationResult{
		UsagePercent:    usagePct,
		IsPeak:          isPeak,
		RemainingTokens: usage.RemainingTokens(),
		TimeUntilReset:  usage.TimeUntilReset(now),
	}

	// Check tiers from highest to lowest, return the highest triggered.
	for i := len(c.Tiers) - 1; i >= 0; i-- {
		tier := c.Tiers[i]
		effective := c.EffectiveThreshold(tier, now)
		if usagePct >= effective {
			result.Triggered = true
			result.ActiveTier = &c.Tiers[i]
			result.EffectivePercent = effective
			return result
		}
	}

	return result
}

// FormatWarning generates a human-readable warning message from an evaluation result.
// Returns empty string if no threshold was triggered.
func FormatWarning(result EvaluationResult) string {
	if !result.Triggered || result.ActiveTier == nil {
		return ""
	}

	peakIndicator := ""
	if result.IsPeak {
		peakIndicator = " [PEAK HOURS]"
	}

	resetStr := formatDuration(result.TimeUntilReset)

	return fmt.Sprintf(
		"QUOTA_WARNING [%s]%s: Usage at %.1f%% (threshold: %.1f%%). "+
			"~%s tokens remaining. Window resets in %s. "+
			"Suggestion: %s",
		result.ActiveTier.Label,
		peakIndicator,
		result.UsagePercent,
		result.EffectivePercent,
		formatTokenCount(result.RemainingTokens),
		resetStr,
		result.ActiveTier.Suggestion,
	)
}

// formatDuration formats a duration as a human-readable string like "2h 15m".
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// formatTokenCount formats a token count with K/M suffixes for readability.
func formatTokenCount(tokens int64) string {
	switch {
	case tokens >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	case tokens >= 1_000:
		return fmt.Sprintf("%.1fK", float64(tokens)/1_000)
	default:
		return fmt.Sprintf("%d", tokens)
	}
}
