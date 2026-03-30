package quota

import (
	"path/filepath"
	"sort"
	"strings"
)

// PriorityTier classifies agents by their operational importance.
type PriorityTier int

const (
	// TierP0 is for active workers — highest priority, doing story work.
	TierP0 PriorityTier = iota
	// TierP1 is for merge-queue and pr-shepherd — critical infrastructure.
	TierP1
	// TierP2 is for envoy, watchdogs — monitoring and coordination.
	TierP2
	// TierP3 is for retrospector — lowest priority, background analysis.
	TierP3
	// TierUnknown is for agents that cannot be classified.
	TierUnknown
)

// String returns the human-readable tier label.
func (t PriorityTier) String() string {
	switch t {
	case TierP0:
		return "P0"
	case TierP1:
		return "P1"
	case TierP2:
		return "P2"
	case TierP3:
		return "P3"
	default:
		return "P?"
	}
}

// defaultTierMap maps known agent names to their priority tiers.
var defaultTierMap = map[string]PriorityTier{
	"merge-queue":      TierP1,
	"pr-shepherd":      TierP1,
	"envoy":            TierP2,
	"arch-watchdog":    TierP2,
	"project-watchdog": TierP2,
	"retrospector":     TierP3,
}

// AgentTier returns the priority tier for a given agent name.
// Workers (names starting with known worker prefixes) are P0.
// Known persistent agents use the default tier map.
// Unknown agents return TierUnknown.
func AgentTier(agentName string) PriorityTier {
	if tier, ok := defaultTierMap[agentName]; ok {
		return tier
	}
	if agentName == "unknown" {
		return TierUnknown
	}
	// Workers and other active agents default to P0.
	return TierP0
}

// AgentUsage holds aggregated usage data for a single agent.
type AgentUsage struct {
	AgentName     string       `json:"agent_name"`
	Tier          PriorityTier `json:"tier"`
	TierLabel     string       `json:"tier_label"`
	Tokens        TokenCount   `json:"tokens"`
	TotalTokens   int64        `json:"total_tokens"`
	UsagePercent  float64      `json:"usage_percent"`
	SessionCount  int          `json:"session_count"`
	Interactions  int          `json:"interactions"`
	Productive    int          `json:"productive"`
	Overhead      int          `json:"overhead"`
	ProductivePct float64      `json:"productive_pct"`
}

// Attribution holds the full per-agent usage breakdown.
type Attribution struct {
	Agents      []AgentUsage `json:"agents"`
	TotalTokens int64        `json:"total_tokens"`
}

// AgentMapper resolves JSONL session file paths to agent names.
type AgentMapper struct {
	// WorktreeBase is the multiclaude worktree root, e.g. ~/.multiclaude/wts/ThreeDoors/
	WorktreeBase string
	// ManualMappings maps session file paths (or prefixes) to agent names.
	ManualMappings map[string]string
}

// MapAgent determines the agent name for a given JSONL source file path.
// It tries worktree path inference first, then manual mappings, then falls
// back to "unknown".
func (m *AgentMapper) MapAgent(sourcePath string) string {
	if name := m.inferFromWorktree(sourcePath); name != "" {
		return name
	}
	if name := m.lookupManual(sourcePath); name != "" {
		return name
	}
	return "unknown"
}

// inferFromWorktree extracts the agent name from a worktree path pattern.
// multiclaude worktrees live at <base>/<agent-name>/, so the path segment
// after the worktree base is the agent name.
func (m *AgentMapper) inferFromWorktree(sourcePath string) string {
	if m.WorktreeBase == "" {
		return ""
	}
	base := filepath.Clean(m.WorktreeBase)
	clean := filepath.Clean(sourcePath)

	if !strings.HasPrefix(clean, base) {
		return ""
	}

	// Path after base: /<agent-name>/...
	rest := strings.TrimPrefix(clean, base)
	rest = strings.TrimPrefix(rest, string(filepath.Separator))
	parts := strings.SplitN(rest, string(filepath.Separator), 2)
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return parts[0]
}

func (m *AgentMapper) lookupManual(sourcePath string) string {
	for prefix, name := range m.ManualMappings {
		if strings.Contains(sourcePath, prefix) {
			return name
		}
	}
	return ""
}

// Attribute groups interactions by agent and produces a full Attribution report.
func Attribute(interactions []Interaction, mapper *AgentMapper) Attribution {
	agentMap := make(map[string]*agentAccum)

	for _, ix := range interactions {
		name := mapper.MapAgent(ix.SourcePath)
		acc, ok := agentMap[name]
		if !ok {
			acc = &agentAccum{sessions: make(map[string]struct{})}
			agentMap[name] = acc
		}
		acc.tokens.InputTokens += ix.Tokens.InputTokens
		acc.tokens.OutputTokens += ix.Tokens.OutputTokens
		acc.tokens.CacheCreationInputTokens += ix.Tokens.CacheCreationInputTokens
		acc.tokens.CacheReadInputTokens += ix.Tokens.CacheReadInputTokens
		acc.interactions++
		acc.sessions[ix.SessionID] = struct{}{}
		if ix.HasToolUse {
			acc.productive++
		} else {
			acc.overhead++
		}
	}

	var totalTokens int64
	for _, acc := range agentMap {
		totalTokens += acc.tokens.Total()
	}

	agents := make([]AgentUsage, 0, len(agentMap))
	for name, acc := range agentMap {
		total := acc.tokens.Total()
		var usagePct float64
		if totalTokens > 0 {
			usagePct = float64(total) / float64(totalTokens) * 100
		}
		var productivePct float64
		if acc.interactions > 0 {
			productivePct = float64(acc.productive) / float64(acc.interactions) * 100
		}
		tier := AgentTier(name)
		agents = append(agents, AgentUsage{
			AgentName:     name,
			Tier:          tier,
			TierLabel:     tier.String(),
			Tokens:        acc.tokens,
			TotalTokens:   total,
			UsagePercent:  usagePct,
			SessionCount:  len(acc.sessions),
			Interactions:  acc.interactions,
			Productive:    acc.productive,
			Overhead:      acc.overhead,
			ProductivePct: productivePct,
		})
	}

	// Sort by total tokens descending (AC5).
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].TotalTokens > agents[j].TotalTokens
	})

	return Attribution{
		Agents:      agents,
		TotalTokens: totalTokens,
	}
}

type agentAccum struct {
	tokens       TokenCount
	interactions int
	productive   int
	overhead     int
	sessions     map[string]struct{}
}
