package quota

import (
	"path/filepath"
	"testing"
	"time"
)

func TestAgentTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		agent    string
		wantTier PriorityTier
	}{
		{"worker is P0", "clever-rabbit", TierP0},
		{"merge-queue is P1", "merge-queue", TierP1},
		{"pr-shepherd is P1", "pr-shepherd", TierP1},
		{"envoy is P2", "envoy", TierP2},
		{"arch-watchdog is P2", "arch-watchdog", TierP2},
		{"project-watchdog is P2", "project-watchdog", TierP2},
		{"retrospector is P3", "retrospector", TierP3},
		{"unknown is TierUnknown", "unknown", TierUnknown},
		{"arbitrary worker is P0", "happy-wolf", TierP0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := AgentTier(tt.agent)
			if got != tt.wantTier {
				t.Errorf("AgentTier(%q) = %v, want %v", tt.agent, got, tt.wantTier)
			}
		})
	}
}

func TestPriorityTierString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tier PriorityTier
		want string
	}{
		{TierP0, "P0"},
		{TierP1, "P1"},
		{TierP2, "P2"},
		{TierP3, "P3"},
		{TierUnknown, "P?"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.tier.String(); got != tt.want {
				t.Errorf("PriorityTier(%d).String() = %q, want %q", tt.tier, got, tt.want)
			}
		})
	}
}

func TestAgentMapperWorktree(t *testing.T) {
	t.Parallel()

	mapper := &AgentMapper{
		WorktreeBase: "/home/user/.multiclaude/wts/ThreeDoors",
	}

	tests := []struct {
		name       string
		sourcePath string
		want       string
	}{
		{
			name:       "worker worktree path",
			sourcePath: "/home/user/.multiclaude/wts/ThreeDoors/clever-rabbit/.claude/projects/hash/sess.jsonl",
			want:       "clever-rabbit",
		},
		{
			name:       "merge-queue worktree path",
			sourcePath: "/home/user/.multiclaude/wts/ThreeDoors/merge-queue/.claude/projects/hash/sess.jsonl",
			want:       "merge-queue",
		},
		{
			name:       "unrelated path falls back to unknown",
			sourcePath: "/home/user/.claude/projects/hash/sess.jsonl",
			want:       "unknown",
		},
		{
			name:       "empty source path",
			sourcePath: "",
			want:       "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapper.MapAgent(tt.sourcePath)
			if got != tt.want {
				t.Errorf("MapAgent(%q) = %q, want %q", tt.sourcePath, got, tt.want)
			}
		})
	}
}

func TestAgentMapperManual(t *testing.T) {
	t.Parallel()

	mapper := &AgentMapper{
		ManualMappings: map[string]string{
			"special-project": "supervisor",
		},
	}

	got := mapper.MapAgent("/home/user/.claude/projects/special-project/sess.jsonl")
	if got != "supervisor" {
		t.Errorf("MapAgent with manual mapping = %q, want %q", got, "supervisor")
	}
}

func TestAgentMapperEmptyBase(t *testing.T) {
	t.Parallel()

	mapper := &AgentMapper{}
	got := mapper.MapAgent("/some/path/sess.jsonl")
	if got != "unknown" {
		t.Errorf("MapAgent with empty mapper = %q, want %q", got, "unknown")
	}
}

func TestAttribute(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
	mapper := &AgentMapper{
		WorktreeBase: "/wts/repo",
	}

	interactions := []Interaction{
		// Worker: 2 productive, 1 overhead
		{
			SessionID: "s1", Timestamp: base, SourcePath: "/wts/repo/clever-rabbit/proj/s1.jsonl",
			Tokens: TokenCount{InputTokens: 5000, OutputTokens: 2000}, HasToolUse: true,
		},
		{
			SessionID: "s1", Timestamp: base, SourcePath: "/wts/repo/clever-rabbit/proj/s1.jsonl",
			Tokens: TokenCount{InputTokens: 3000, OutputTokens: 1000}, HasToolUse: true,
		},
		{
			SessionID: "s1", Timestamp: base, SourcePath: "/wts/repo/clever-rabbit/proj/s1.jsonl",
			Tokens: TokenCount{InputTokens: 1000, OutputTokens: 500}, HasToolUse: false,
		},
		// Merge-queue: 1 productive
		{
			SessionID: "s2", Timestamp: base, SourcePath: "/wts/repo/merge-queue/proj/s2.jsonl",
			Tokens: TokenCount{InputTokens: 2000, OutputTokens: 800}, HasToolUse: true,
		},
		// Retrospector: 1 overhead
		{
			SessionID: "s3", Timestamp: base, SourcePath: "/wts/repo/retrospector/proj/s3.jsonl",
			Tokens: TokenCount{InputTokens: 500, OutputTokens: 100}, HasToolUse: false,
		},
	}

	attr := Attribute(interactions, mapper)

	if len(attr.Agents) != 3 {
		t.Fatalf("agents = %d, want 3", len(attr.Agents))
	}

	// Should be sorted by total tokens descending.
	if attr.Agents[0].AgentName != "clever-rabbit" {
		t.Errorf("top agent = %q, want clever-rabbit", attr.Agents[0].AgentName)
	}

	// Verify worker breakdown.
	worker := attr.Agents[0]
	if worker.Interactions != 3 {
		t.Errorf("worker interactions = %d, want 3", worker.Interactions)
	}
	if worker.Productive != 2 {
		t.Errorf("worker productive = %d, want 2", worker.Productive)
	}
	if worker.Overhead != 1 {
		t.Errorf("worker overhead = %d, want 1", worker.Overhead)
	}
	if worker.SessionCount != 1 {
		t.Errorf("worker sessions = %d, want 1", worker.SessionCount)
	}
	if worker.Tier != TierP0 {
		t.Errorf("worker tier = %v, want P0", worker.Tier)
	}
	wantWorkerTokens := int64(5000 + 2000 + 3000 + 1000 + 1000 + 500)
	if worker.TotalTokens != wantWorkerTokens {
		t.Errorf("worker total = %d, want %d", worker.TotalTokens, wantWorkerTokens)
	}

	// Verify merge-queue.
	mq := attr.Agents[1]
	if mq.AgentName != "merge-queue" {
		t.Errorf("second agent = %q, want merge-queue", mq.AgentName)
	}
	if mq.Tier != TierP1 {
		t.Errorf("merge-queue tier = %v, want P1", mq.Tier)
	}

	// Verify retrospector.
	retro := attr.Agents[2]
	if retro.AgentName != "retrospector" {
		t.Errorf("third agent = %q, want retrospector", retro.AgentName)
	}
	if retro.Tier != TierP3 {
		t.Errorf("retrospector tier = %v, want P3", retro.Tier)
	}
	if retro.Productive != 0 {
		t.Errorf("retrospector productive = %d, want 0", retro.Productive)
	}
	if retro.Overhead != 1 {
		t.Errorf("retrospector overhead = %d, want 1", retro.Overhead)
	}

	// Verify percentages sum to ~100.
	var totalPct float64
	for _, a := range attr.Agents {
		totalPct += a.UsagePercent
	}
	if totalPct < 99.9 || totalPct > 100.1 {
		t.Errorf("usage percentages sum to %.2f, want ~100", totalPct)
	}
}

func TestAttributeUnmapped(t *testing.T) {
	t.Parallel()

	mapper := &AgentMapper{WorktreeBase: "/wts/repo"}
	interactions := []Interaction{
		{
			SessionID: "s1", SourcePath: "/other/path/sess.jsonl",
			Tokens: TokenCount{InputTokens: 1000, OutputTokens: 500},
		},
	}

	attr := Attribute(interactions, mapper)

	if len(attr.Agents) != 1 {
		t.Fatalf("agents = %d, want 1", len(attr.Agents))
	}
	if attr.Agents[0].AgentName != "unknown" {
		t.Errorf("unmapped agent = %q, want unknown", attr.Agents[0].AgentName)
	}
	if attr.Agents[0].Tier != TierUnknown {
		t.Errorf("unknown tier = %v, want TierUnknown", attr.Agents[0].Tier)
	}
}

func TestAttributeEmpty(t *testing.T) {
	t.Parallel()

	mapper := &AgentMapper{WorktreeBase: "/wts/repo"}
	attr := Attribute(nil, mapper)

	if len(attr.Agents) != 0 {
		t.Errorf("agents = %d, want 0", len(attr.Agents))
	}
	if attr.TotalTokens != 0 {
		t.Errorf("total = %d, want 0", attr.TotalTokens)
	}
}

func TestAttributeMultipleSessions(t *testing.T) {
	t.Parallel()

	mapper := &AgentMapper{WorktreeBase: "/wts/repo"}
	base := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
	interactions := []Interaction{
		{
			SessionID: "s1", Timestamp: base, SourcePath: "/wts/repo/envoy/proj/s1.jsonl",
			Tokens: TokenCount{InputTokens: 1000},
		},
		{
			SessionID: "s2", Timestamp: base, SourcePath: "/wts/repo/envoy/proj/s2.jsonl",
			Tokens: TokenCount{InputTokens: 2000},
		},
		{
			SessionID: "s3", Timestamp: base, SourcePath: "/wts/repo/envoy/proj/s3.jsonl",
			Tokens: TokenCount{InputTokens: 500},
		},
	}

	attr := Attribute(interactions, mapper)

	if len(attr.Agents) != 1 {
		t.Fatalf("agents = %d, want 1", len(attr.Agents))
	}
	if attr.Agents[0].SessionCount != 3 {
		t.Errorf("envoy sessions = %d, want 3", attr.Agents[0].SessionCount)
	}
}

func TestAttributeZeroUsageAgent(t *testing.T) {
	t.Parallel()

	mapper := &AgentMapper{WorktreeBase: "/wts/repo"}
	interactions := []Interaction{
		{
			SessionID: "s1", SourcePath: "/wts/repo/envoy/proj/s1.jsonl",
			Tokens: TokenCount{},
		},
	}

	attr := Attribute(interactions, mapper)

	if len(attr.Agents) != 1 {
		t.Fatalf("agents = %d, want 1", len(attr.Agents))
	}
	if attr.Agents[0].TotalTokens != 0 {
		t.Errorf("zero-usage total = %d, want 0", attr.Agents[0].TotalTokens)
	}
	if attr.Agents[0].UsagePercent != 0 {
		t.Errorf("zero-usage pct = %.2f, want 0", attr.Agents[0].UsagePercent)
	}
}

func TestParseFileDetectsToolUse(t *testing.T) {
	t.Parallel()

	interactions, err := ParseFile(filepath.Join("testdata", "agent_worker.jsonl"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(interactions) != 3 {
		t.Fatalf("got %d interactions, want 3", len(interactions))
	}

	// First two have tool_use, third does not.
	if !interactions[0].HasToolUse {
		t.Error("interaction 0: expected HasToolUse=true")
	}
	if !interactions[1].HasToolUse {
		t.Error("interaction 1: expected HasToolUse=true")
	}
	if interactions[2].HasToolUse {
		t.Error("interaction 2: expected HasToolUse=false")
	}
}

func TestParseFileSetsSourcePath(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "normal.jsonl")
	interactions, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, ix := range interactions {
		if ix.SourcePath != path {
			t.Errorf("interaction %d: SourcePath = %q, want %q", i, ix.SourcePath, path)
		}
	}
}

func TestAttributeProductivePct(t *testing.T) {
	t.Parallel()

	mapper := &AgentMapper{WorktreeBase: "/wts/repo"}
	interactions := []Interaction{
		{
			SessionID: "s1", SourcePath: "/wts/repo/worker/p/s.jsonl",
			Tokens: TokenCount{InputTokens: 100}, HasToolUse: true,
		},
		{
			SessionID: "s1", SourcePath: "/wts/repo/worker/p/s.jsonl",
			Tokens: TokenCount{InputTokens: 100}, HasToolUse: true,
		},
		{
			SessionID: "s1", SourcePath: "/wts/repo/worker/p/s.jsonl",
			Tokens: TokenCount{InputTokens: 100}, HasToolUse: false,
		},
		{
			SessionID: "s1", SourcePath: "/wts/repo/worker/p/s.jsonl",
			Tokens: TokenCount{InputTokens: 100}, HasToolUse: false,
		},
	}

	attr := Attribute(interactions, mapper)
	// 2/4 = 50%
	wantPct := 50.0
	if attr.Agents[0].ProductivePct != wantPct {
		t.Errorf("ProductivePct = %.2f, want %.2f", attr.Agents[0].ProductivePct, wantPct)
	}
}
