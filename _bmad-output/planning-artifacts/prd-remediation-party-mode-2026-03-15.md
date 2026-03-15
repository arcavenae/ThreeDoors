---
name: PRD Remediation Party Mode
date: 2026-03-15
trigger: PRD Validation Report (R-002, P-007)
participants: John (PM), Winston (Architect), Mary (Analyst), Sally (UX)
rounds: 3 per topic (6 total)
topics:
  - Phase numbering consolidation in product-scope.md
  - User journey expansion in user-journeys.md
---

# PRD Remediation Party Mode — Phase Numbering & User Journeys

## Topic 1: Phase Numbering Consolidation

### Problem

`docs/prd/product-scope.md` has 16 "phases" with chaotic numbering: 1, 2, 3, 3.5, 3.5+ (×3), 4 (×2), 4+, 4.5, 5 (×2), 5.5, 5.5+, 6, 6+, 7, 8. Two Phase 4 entries and two Phase 5 entries contain overlapping content. The decimal proliferation makes it impossible to reason about project scope or navigate the document.

### Adopted Approach: 7-Phase Integer Schema

Consolidate into exactly 7 phases with integer-only numbering. Use `###` sub-headings within phases for thematic grouping. No decimals.

| Phase | Title | Epics |
|-------|-------|-------|
| 1 | Foundation | 1-2 |
| 2 | Growth | 3-7, 17 |
| 3 | Platform Expansion | 8-15, 18, 28-29, 31 |
| 4 | Task Source Integration | 19-21, 25-26, 30, 43-47, 63 |
| 5 | User Experience & Polish | 27, 32-33, 35-36, 39-41, 48, 56-57, 59, 69-70 |
| 6 | Developer Experience & Governance | 22-24, 34, 37-38, 42, 49-53, 55, 58, 60-62, 65, 67 |
| 7 | Future Expansion | 16, 54, 64 |

**Key design rules:**
- Exactly one `##` heading per phase, integer numbering only
- `###` sub-sections within phases for thematic grouping (e.g., "Visual Polish" within Phase 5)
- Each phase heading includes epic list: `## Phase N: Title (Epics X-Y, Z)`
- No phases collapsed for completion — scope doc is the *what*, not the *when*
- Completion status belongs in `epic-list.md`, not here

### Rejected Alternatives

1. **Keep decimal phases but normalize** (e.g., 3.5 → 3.1): Still produces confusing numbering and doesn't solve the thematic incoherence. Phase 3.5+ (×3) would become 3.1, 3.2, 3.3 — misleading sub-phase relationship.

2. **Collapse completed phases into summary**: Would lose the scope specification. The document serves as the complete product scope reference — removing completed scope means losing traceability from requirements to scope.

3. **8+ phases with finer granularity**: More phases = more confusion. The whole problem was too many phases. 7 is already at the upper limit of the recommendation.

4. **Named phases without numbers**: Phases like "Foundation", "Growth" etc. without numbers lose ordering semantics. Numbering conveys progression.

### Consensus

PM, Architect, and Analyst unanimously endorsed the 7-phase integer schema. The mapping was validated against all 67+ completed epics to ensure every epic has exactly one home.

---

## Topic 2: User Journey Expansion

### Problem

Only 9 user journeys exist, covering Phase 1-2 features. With 67+ completed epics spanning 6+ phases, major user workflows have no journey representation. This creates a traceability gap from user experience back to requirements.

### Adopted Approach: Add 5 New Journeys (10-14)

| # | Journey | Key Epics | FR Coverage |
|---|---------|-----------|-------------|
| 10 | First-Run Onboarding | 8, 17 | FR31-FR34, FR55-FR62 |
| 11 | Task Source Connection & Multi-Source Usage | 43-47 | FR63-FR80 |
| 12 | Daily Planning Mode | 27 | FR81-FR88 |
| 13 | Snooze & Defer Workflow | 28 | FR40-FR46 |
| 14 | CLI Task Management | 23 | FR47-FR51 |

**Journey 10: First-Run Onboarding**
- User launches ThreeDoors for the first time → empty state → guided task addition → first door selection → theme picker → ready to use

**Journey 11: Task Source Connection & Multi-Source Usage**
- User decides to connect an external source → runs setup wizard → configures credentials → verifies sync → doors now show mixed tasks from multiple sources → completes a task → completion syncs back to source

**Journey 12: Daily Planning Mode**
- User starts morning ritual → reviews yesterday's progress → energy-aware task selection → confirms today's focus tasks → starts working with curated door set

**Journey 13: Snooze & Defer Workflow**
- User encounters a task they can't do now → presses Z to snooze → selects "Next Week" → task disappears from door pool → a week later, task auto-returns to doors → user completes it

**Journey 14: CLI Task Management**
- Power user manages tasks from CLI → `threedoors task list` → `threedoors task add "Fix login bug"` → `threedoors door` for quick selection → `threedoors task complete <id>` → `threedoors stats` for session summary

### Rejected Alternatives

1. **Add Doctor journey**: Too niche — `threedoors doctor` is a diagnostic tool, not a recurring user workflow. Referenced by Epic 49 but doesn't represent a distinct mental model.

2. **Add separate Multi-Source Aggregation journey**: Aggregation is implicitly covered by Journey 11 — once you connect 2+ sources, multi-source aggregation happens automatically. A separate journey would duplicate content.

3. **Add MCP/LLM Integration journey**: Too implementation-focused for a user journey. MCP consumers are developer tools, not end users. Could be added later if LLM-driven task management becomes a primary user flow.

4. **Add Bug Reporting journey**: Epic 50 (`:bug` command) is a secondary feature, not a primary workflow.

### Consensus

PM, UX Designer, and Architect unanimously endorsed 5 new journeys (10-14). UX Designer prioritized First-Run Onboarding as the most important missing journey (every user's entry point). Architect requested Journey 11 explicitly mention mixed-source doors as the architectural proof point.

---

## Summary

Two decisions made:
1. **Phase numbering**: Consolidate 16 phases → 7 integer phases. No decimals.
2. **User journeys**: Add 5 new journeys (10-14) covering onboarding, source connection, daily planning, snooze, and CLI.

Both decisions are docs-only PRD changes. No code impact.
