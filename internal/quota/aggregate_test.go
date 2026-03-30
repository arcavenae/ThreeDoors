package quota

import (
	"testing"
	"time"
)

func makeInteractions() []Interaction {
	base := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
	return []Interaction{
		{SessionID: "s1", Timestamp: base.Add(-6 * time.Hour), Tokens: TokenCount{InputTokens: 1000, OutputTokens: 500}},  // outside 5h window
		{SessionID: "s1", Timestamp: base.Add(-4 * time.Hour), Tokens: TokenCount{InputTokens: 2000, OutputTokens: 1000}}, // inside
		{SessionID: "s2", Timestamp: base.Add(-3 * time.Hour), Tokens: TokenCount{InputTokens: 3000, OutputTokens: 1500}}, // inside
		{SessionID: "s1", Timestamp: base.Add(-1 * time.Hour), Tokens: TokenCount{InputTokens: 500, OutputTokens: 250}},   // inside
		{SessionID: "s2", Timestamp: base, Tokens: TokenCount{InputTokens: 1000, OutputTokens: 500}},                      // inside (at boundary)
	}
}

func TestAggregateWindow(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
	interactions := makeInteractions()

	wu := AggregateWindow(interactions, base, DefaultWindow)

	// 4 interactions inside the 5h window (the -6h one is excluded)
	wantInput := int64(2000 + 3000 + 500 + 1000)
	wantOutput := int64(1000 + 1500 + 250 + 500)
	if wu.Tokens.InputTokens != wantInput {
		t.Errorf("input = %d, want %d", wu.Tokens.InputTokens, wantInput)
	}
	if wu.Tokens.OutputTokens != wantOutput {
		t.Errorf("output = %d, want %d", wu.Tokens.OutputTokens, wantOutput)
	}

	if len(wu.Sessions) != 2 {
		t.Fatalf("sessions = %d, want 2", len(wu.Sessions))
	}

	if wu.WindowStart != base.Add(-DefaultWindow) {
		t.Errorf("window_start = %v, want %v", wu.WindowStart, base.Add(-DefaultWindow))
	}
	if wu.WindowEnd != base {
		t.Errorf("window_end = %v, want %v", wu.WindowEnd, base)
	}
}

func TestAggregateWindowEmpty(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	wu := AggregateWindow(nil, now, DefaultWindow)
	if wu.Tokens.Total() != 0 {
		t.Errorf("expected 0 tokens for nil interactions, got %d", wu.Tokens.Total())
	}
	if len(wu.Sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(wu.Sessions))
	}
}

func TestAggregateWindowBoundary(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
	// Interaction exactly at window start boundary (windowEnd - window)
	interactions := []Interaction{
		{SessionID: "s1", Timestamp: base.Add(-5 * time.Hour), Tokens: TokenCount{InputTokens: 100}},
	}

	wu := AggregateWindow(interactions, base, DefaultWindow)
	// Exactly at boundary: windowStart = base - 5h. The interaction timestamp
	// equals windowStart, which is Before(windowStart)==false, so it's included.
	if wu.Tokens.InputTokens != 100 {
		t.Errorf("boundary interaction: input = %d, want 100", wu.Tokens.InputTokens)
	}
}

func TestSnapshot(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
	interactions := makeInteractions()

	snap := Snapshot(interactions, base, DefaultWindow, PlanMax5x)

	wantConsumed := int64(2000 + 3000 + 500 + 1000 + 1000 + 1500 + 250 + 500)
	if snap.TokensConsumed != wantConsumed {
		t.Errorf("consumed = %d, want %d", snap.TokensConsumed, wantConsumed)
	}
	if snap.TokenBudget != 88_000 {
		t.Errorf("budget = %d, want 88000", snap.TokenBudget)
	}

	wantPct := float64(wantConsumed) / 88_000 * 100
	if snap.UsagePercent != wantPct {
		t.Errorf("pct = %.2f, want %.2f", snap.UsagePercent, wantPct)
	}
	if snap.Budget.Name != "Max 5x" {
		t.Errorf("budget name = %q, want %q", snap.Budget.Name, "Max 5x")
	}
}

func TestSnapshotZeroBudget(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	snap := Snapshot(nil, now, DefaultWindow, PlanBudget{Name: "Zero", TokenBudget: 0})
	if snap.UsagePercent != 0 {
		t.Errorf("expected 0%% for zero budget, got %.2f", snap.UsagePercent)
	}
}
