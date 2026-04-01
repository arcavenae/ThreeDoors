# ADR-0035: Three-Tier Decision System for Party Mode

- **Status:** Accepted
- **Date:** 2026-03-08
- **Decision Makers:** Project founder
- **Related PRs:** #265
- **Related ADRs:** ADR-0029 (Governance Phase Renaming)

## Context

The project has accumulated 46+ party mode artifacts, some exceeding 600 lines. Full party mode sessions (18 agents) are expensive in time and tokens for tactical decisions. Cross-project analysis (Sprint 336 retro, orc-penny sidecars) confirms diminishing returns per round of multi-agent deliberation.

Not all decisions warrant the same level of deliberation. A config change and an architectural redesign should not go through the same process.

## Decision

Adopt a three-tier decision system:

### Tier 1: Quick Decision
- **Process:** Single agent recommends, human approves
- **When to use:** Config/tooling changes, naming decisions, minor refactors, story task ordering
- **Artifact:** Brief note in PR description or Decision Board entry
- **Examples:** "Should we use `snake_case` for this field?", "Which linter rule to enable?"

### Tier 2: Standard Decision
- **Process:** 3-agent party mode (relevant domain experts)
- **When to use:** Feature design, API contracts, integration approaches, multi-story coordination
- **Artifact:** Party mode artifact saved to `_bmad-output/planning-artifacts/`
- **Examples:** "How should the Todoist adapter handle field mapping?", "What's the UX for snooze?"

### Tier 3: Full Decision
- **Process:** Full party mode (6+ agents), extensive deliberation
- **When to use:** Architecture changes, philosophy/SOUL.md changes, epic scoping, methodology changes
- **Artifact:** Full party mode artifact + Decision Board entry + ADR if applicable
- **Examples:** "Should we adopt a new persistence layer?", "How should agent governance work?"

## Rationale

- Right-sizes deliberation cost to decision importance
- Tier 1 eliminates the overhead of spinning up party mode for trivial choices
- Tier 2 captures the middle ground — decisions that benefit from multiple perspectives but don't need 18 agents
- Tier 3 preserves the full deliberation process for decisions that genuinely warrant it
- Clear triggers prevent bikeshedding about which tier to use

## Rejected Alternatives

- **Two-tier system (Quick vs. Full):** Misses the middle ground. Many feature design decisions need more than one agent's perspective but far less than full party mode.
- **Guidelines only, no tiers:** Agents default to full party mode (the safest choice) when given discretion, negating the cost savings.
