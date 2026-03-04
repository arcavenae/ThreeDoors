# Design Decisions Requiring Maintainer Input

> **Request:** A `/party-mode` session is requested to workshop these decisions collaboratively. Many of these are interconnected — resolving them in a facilitated group discussion will surface trade-offs and dependencies that sequential decision-making would miss.

**Total decisions identified:** 42
**Priority breakdown:** 7 Critical | 10 High | 14 Medium | 11 Low

---

## How to Use This Document

Each decision is framed as a clear question with options, trade-offs, and a recommendation where the research supports one. Decisions are grouped by priority:

- **Critical** — Blocks active or imminent implementation work
- **High** — Blocks the next wave of epics or has architectural ripple effects
- **Medium** — Needs resolution before specific stories can be written
- **Low** — Can be deferred but should be tracked

---

## Critical Decisions (Blocking Current/Imminent Work)

### C1. Task Model Extension Strategy

**Question:** How should the Task struct evolve to support DueDate, Priority, and multi-source tracking (SourceRefs)?

**Context:** Epics 19 (Jira), 20 (Apple Reminders), 12 (Calendar), 13 (Multi-Source Aggregation), and 21 (Sync Hardening) all need fields that don't exist yet. The approach chosen here affects migration burden and backward compatibility for every downstream epic.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Encode in Context field | Stuff DueDate/Priority into the existing `Context` string | No breaking change, but parsing is fragile and limits querying |
| B. Incremental struct extension | Add fields one at a time as each epic needs them | Controlled migration, but multiple schema bumps |
| C. All-at-once model v2 | Single migration adding DueDate, Priority, SourceRefs[], FieldVersions | One migration to handle, but larger blast radius |

**Recommendation:** Option B (incremental) — aligns with YAGNI principle and lets each epic own its migration.

**Blocked:** Epics 12, 13, 19, 20, 21

**Source:** `docs/research/sync-architecture-scaling-research.md`, `docs/research/task-source-expansion-research.md`, `docs/architecture/task-sync-architecture.md`

---

### C2. Sync Scheduler Architecture

**Question:** Should the sync scheduler run as a background goroutine or integrate into the Bubbletea event loop via `tea.Cmd`?

**Context:** The task-sync-architecture doc recommends `tea.Cmd` but this hasn't been finalized. This decision affects all sync-related work (Epics 11, 19, 20, 21) and determines how sync status is displayed in the TUI.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Background goroutine + channel | Standard Go concurrency; sync results sent via channel to TUI | Familiar pattern, but harder to test; must bridge to Bubbletea via subscription |
| B. `tea.Cmd` integration | Sync operations dispatched as Bubbletea commands returning `tea.Msg` | Native to TUI framework, testable, but async operations need careful `tea.Batch` management |
| C. Hybrid | Goroutine manages scheduling; emits `tea.Cmd` for each sync cycle | Combines benefits but adds complexity |

**Recommendation:** Option B (`tea.Cmd`) — consistent with Bubbletea patterns already used in the project, better testability.

**Blocked:** Epics 11, 19, 20, 21

**Source:** `docs/architecture/task-sync-architecture.md`

---

### C3. Conflict Resolution Strategy

**Question:** Should sync conflicts use task-level last-write-wins (LWW) or property-level merging?

**Context:** Task-level LWW is simpler but risks data loss (e.g., editing title on one device while editing notes on another — one edit is lost). Property-level merging preserves both changes but requires per-field timestamp tracking, increasing storage and complexity.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Task-level LWW | Entire task is replaced by most recent write | Simple, but loses concurrent edits to different fields |
| B. Property-level LWW | Each field tracks its own `(value, updatedAt, actor)` | Preserves concurrent edits, but ~3x storage per task, complex migration |
| C. Task-level LWW now, property-level later | Ship with A, migrate to B when multi-source is real | Lower upfront cost, but migration is harder once data exists |

**Recommendation:** Option C — start simple, but design the SourceRef schema to accommodate future property-level tracking.

**Blocked:** Epic 21 (Sync Protocol Hardening)

**Source:** `docs/research/sync-architecture-scaling-research.md`, `docs/architecture/task-sync-architecture.md`

---

### C4. Jira Integration Phase 1 Scope

**Question:** Is Jira Phase 1 read-only (pull tasks into ThreeDoors) or bidirectional (also push status changes back)?

**Context:** Read-only is ~50% less implementation effort. Research recommends read-only Phase 1 with bidirectional in Phase 2, but this hasn't been confirmed.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Read-only Phase 1 | Pull Jira issues into doors; status changes stay local | Faster to ship, but users must update Jira separately |
| B. Bidirectional Phase 1 | Full sync including pushing completions back to Jira | Higher value, but doubles implementation effort and adds write-conflict risk |
| C. Read-only + status-only writeback | Pull tasks, push only status transitions (no field edits) | Middle ground; status is the most important field to sync |

**Recommendation:** Option A (read-only) for Phase 1, confirmed by research.

**Blocked:** Epic 19 story breakdown and estimation

**Source:** `docs/research/jira-integration-research.md`

---

### C5. Apple Reminders DueDate Handling

**Question:** Should due dates from Apple Reminders be encoded in the Context field or added as a native Task struct field?

**Context:** Apple Reminders has native due dates and priorities. The current Task struct has neither. This decision is tightly coupled with C1 (Task Model Extension).

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Encode in Context | `Context: "due:2026-03-15 priority:high"` | No struct change, but fragile parsing |
| B. Add DueDate field now | `DueDate *time.Time` on Task struct | Clean, but triggers schema migration |
| C. Adapter-local metadata | Store in adapter-specific sidecar, don't surface in Task | No migration, but due dates invisible to door selection |

**Recommendation:** Option B — due dates are too important to hack into Context. Align with C1 decision.

**Blocked:** Epic 20 (Apple Reminders Integration)

**Source:** `docs/research/apple-reminders-integration-research.md`

---

### C6. Dedup Strategy for Multi-Source Tasks

**Question:** When the same logical task appears in multiple providers (e.g., a Jira ticket and an Obsidian note), how should dedup work?

**Context:** Multi-source aggregation (Epic 13) depends on this. False-positive dedup is worse than no dedup (merging unrelated tasks is destructive).

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Manual linking only | User explicitly links tasks across providers | Safe but tedious; no automatic matching |
| B. Auto-detect with confirmation | Fuzzy match on title/description, prompt user to confirm | Best UX, but needs confirmation UI design |
| C. Eager auto-merge | Automatically merge tasks above a similarity threshold | Convenient but risks false positives |
| D. SourceRef-based linking | Tasks linked via explicit cross-references in provider metadata | Reliable, but requires adapter support for cross-refs |

**Recommendation:** Option B (auto-detect + confirm) for user-facing, Option D (SourceRef) for programmatic.

**Blocked:** Epic 13 (Multi-Source Aggregation)

**Source:** `docs/research/task-source-expansion-research.md`, `docs/architecture/task-sync-architecture.md`

---

### C7. SourceRef Migration Path

**Question:** How should the existing single `SourceProvider` field migrate to the multi-source `[]SourceRef` model?

**Context:** Epic 21 requires schema version 1→2. Breaking change that affects all existing task files.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Auto-migrate on load | Detect v1, convert `SourceProvider` → `SourceRefs[0]`, write v2 | Seamless for users, but adds migration code permanently |
| B. Migration CLI command | `threedoors migrate` one-time conversion | Explicit, but users must run it manually |
| C. Dual-read support | Read both v1 and v2 formats, write v2 only | No forced migration, but dual-format code debt |

**Recommendation:** Option A (auto-migrate on load) — least user friction.

**Blocked:** Epic 21 (Sync Protocol Hardening)

**Source:** `docs/architecture/task-sync-architecture.md`

---

## High Priority Decisions (Blocking Next Phase)

### H1. Epic 6 Go/No-Go (Data Layer & Enrichment)

**Question:** Should Epic 6 (SQLite enrichment layer) be pursued, deferred, or cancelled?

**Context:** Marked as "optional — may be YAGNI" in epic details. No evaluation criteria exist for making this decision.

**Recommendation:** Define a concrete trigger (e.g., "if >3 adapters need cross-reference queries, build it; otherwise cancel"). Currently leaning toward cancel — file-based storage with in-memory indexing seems sufficient.

**Blocked:** Architecture planning for Epics 13+

**Source:** `docs/prd/epic-details.md`

---

### H2. Plugin/Adapter SDK Scope

**Question:** What is the MVP adapter SDK? How are adapters discovered and configured?

**Context:** Epic 7 acknowledges over-engineering risk ("keep minimal until 3+ adapters"). But "minimal" isn't defined. Config validation, registry discovery, and settings schemas are all unspecified.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Compile-time registration | Adapters are Go packages imported in `main.go` | Simplest, but requires recompilation to add adapters |
| B. Config-driven factory | YAML config specifies adapter name + settings; factory creates instances | Flexible, but needs schema validation |
| C. Plugin system (Go plugins or subprocess) | True plugins loaded at runtime | Most flexible, but Go plugin support is fragile on macOS |

**Recommendation:** Option A for now, evolve to B when 3+ adapters exist.

**Blocked:** Epics 8, 19, 20 (adapter implementation approach)

**Source:** `docs/prd/epic-details.md`, `docs/research/task-source-expansion-research.md`

---

### H3. Circuit Breaker Configuration

**Question:** Should circuit breaker parameters be hardcoded defaults, per-provider configurable, or globally configurable?

**Context:** The sync architecture doc specifies defaults (5 failures / 2 min window / 30s–30min probe interval) but doesn't say whether users can tune these.

**Recommendation:** Hardcoded defaults with per-provider override in config.yaml. Most users won't need to change them.

**Blocked:** Epic 21 stories

**Source:** `docs/architecture/task-sync-architecture.md`

---

### H4. Circuit Breaker TUI Display

**Question:** Should circuit breaker state (open/closed/half-open) be visible in the TUI? If so, where?

**Context:** When a provider is in circuit-breaker-open state, users need to know their tasks may be stale.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Status bar indicator | Small icon/text in footer showing provider health | Minimal UI disruption, always visible |
| B. Sync status view | Dedicated view (e.g., `[S]` key) showing all provider statuses | More detail, but adds a new view to maintain |
| C. Door badge | Badge on each door showing source health | Per-task visibility, but clutters the doors view |

**Recommendation:** Option A (status bar) — least invasive, most informative.

**Blocked:** Epic 21 TUI integration stories

**Source:** `docs/architecture/task-sync-architecture.md`

---

### H5. Adapter Contract Test Strategy

**Question:** Should all adapters implement the full `TaskProvider` interface, or should contract tests be adapter-specific?

**Context:** Some adapters are read-only for Phase 1 (Jira). The contract test suite needs to handle this.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Full interface, return `ErrReadOnly` | All adapters implement all methods; read-only ones return sentinel error | Uniform interface, but boilerplate for read-only adapters |
| B. Separate ReadProvider/WriteProvider interfaces | Split interface; adapters implement what they support | Cleaner, but breaks existing `TaskProvider` interface |
| C. Capability flags | Single interface with `Capabilities()` method returning supported operations | Flexible, but runtime capability checking adds complexity |

**Recommendation:** Option A — simplest, and `ErrReadOnly` is already the established pattern.

**Blocked:** Epic 9 (Testing Strategy), Epics 19, 20

**Source:** `docs/prd/epic-details.md`, `docs/research/jira-integration-research.md`

---

### H6. Onboarding Import Sources

**Question:** Which task sources should the onboarding flow (Epic 10) support for initial import?

**Context:** Onboarding currently targets text files only. But if Obsidian and Apple Notes adapters ship before/alongside onboarding, should import support them too?

**Recommendation:** Text files only for initial onboarding. Add adapter-specific import flows as each adapter ships.

**Blocked:** Epic 10 story completion

**Source:** `docs/prd/epic-details.md`

---

### H7. Multiclaude Auto-Execution Approach

**Question:** Which approach should be used for automated story dispatch to multiclaude workers?

**Context:** Research identifies four options ranging from shell script MVP to full multiclaude pipeline command. No timeline or phase decision has been made.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Shell script MVP | Bash script that parses story files and dispatches `multiclaude worker create` | Works today, minimal effort, but fragile |
| B. GitHub Actions | Trigger workers from CI on story file changes | Natural for PR-based flow, but complex for local dev |
| C. Supervisor enhancement | Extend existing multiclaude supervisor with story awareness | Uses existing infra, but risks context window bloat |
| D. `multiclaude pipeline` command | First-class pipeline support in multiclaude | Most robust, but requires multiclaude changes |

**Recommendation:** Option A (shell script) as immediate MVP, then evaluate C vs D.

**Blocked:** Self-driving development pipeline

**Source:** `docs/research/multiclaude-auto-execution-research.md`

---

### H8. Integration Adapter Priority Order

**Question:** After Apple Notes (done) and the current Epic 17 work, which adapter should be built next?

**Context:** Research recommends Todoist first (simple REST API, large user base), then GitHub Issues, then Linear. But Jira (Epic 19) and Apple Reminders (Epic 20) are already specced.

**Options:**

| Option | Description | Trade-off |
|--------|-------------|-----------|
| A. Jira → Apple Reminders → Todoist | Follow existing epic numbering | Jira is complex; slower to ship a second adapter |
| B. Apple Reminders → Jira → Todoist | Reminders is simpler and stays in Apple ecosystem | Good ecosystem story, but smaller user base than Jira |
| C. Todoist → Apple Reminders → Jira | Research-recommended order | Fastest path to a shipped adapter, but no existing epic/stories |

**Recommendation:** Option B — Apple Reminders builds on existing Apple Notes patterns, Jira can follow.

**Blocked:** Roadmap prioritization for post-Epic 17 work

**Source:** `docs/research/task-source-expansion-research.md`, `docs/research/apple-reminders-integration-research.md`, `docs/research/jira-integration-research.md`

---

### H9. Expand and Fork Feature Definitions

**Question:** What should `[E]xpand` and `[F]ork` do in the detail view?

**Context:** Both are stubbed in `internal/tui/detail_view.go` as "not yet implemented." These are user-facing features with no specification.

**Options for Expand:**

| Option | Description |
|--------|-------------|
| A. Manual sub-task creation | Opens a form to add sub-tasks manually |
| B. LLM-powered decomposition | Sends task to LLM for automatic breakdown (overlaps with `[G]enerate`) |
| C. Template expansion | Applies a template to create standard sub-tasks |

**Options for Fork:**

| Option | Description |
|--------|-------------|
| A. Duplicate task | Creates a copy with new ID, preserving all fields |
| B. Variant creation | Creates a copy with some fields stripped (e.g., no assignee, reset status) |
| C. Branch task | Creates a linked alternative approach to the same goal |

**Recommendation:** Expand = Option A (manual sub-tasks), Fork = Option B (variant creation). Keep LLM decomposition in `[G]enerate`.

**Blocked:** Detail view stories (no epic assigned yet)

**Source:** `internal/tui/detail_view.go`

---

### H10. Performance Benchmark Scope

**Question:** Which operations should be benchmarked against the <100ms NFR, and for which adapters?

**Context:** NFR specifies "<100ms" but doesn't say for what. Local file operations are trivially fast; network adapters obviously can't meet 100ms.

**Recommendation:** Benchmark `LoadTasks()`, `SaveTask()`, `MarkComplete()` for local adapters only. Network adapters should have separate SLAs (e.g., <2s for API calls).

**Blocked:** Epic 9 (Testing Strategy & Quality Gates)

**Source:** `docs/prd/requirements.md`

---

## Medium Priority Decisions

### M1. Door Theme Terminal Width Fallback

**Question:** At what terminal width should the theme system fall back to Classic (ASCII-only) rendering?

**Recommendation:** 60 columns — standard minimum for TUI apps.

**Blocked:** Epic 17 golden file tests

**Source:** `docs/prd/epics-and-stories.md` (Epic 17 stories)

---

### M2. Theme Preview Interaction Design

**Question:** How should theme preview work during onboarding? How many themes visible at once? Scrolling behavior?

**Recommendation:** Show 3 themes side-by-side (matching the "three doors" metaphor), arrow keys to scroll through remaining themes.

**Blocked:** Epic 17, Story 17.5 or 17.6

---

### M3. Jira ADF Description Handling

**Question:** Should Jira's Atlassian Document Format descriptions be ignored, converted to plain text, or rendered as markdown?

**Recommendation:** Strip to plain text for Phase 1 (extract text nodes only). Full ADF rendering is Phase 3+.

**Blocked:** Epic 19 implementation

**Source:** `docs/research/jira-integration-research.md`

---

### M4. Jira Story Points vs Priority Mapping

**Question:** Should ThreeDoors' Effort field map to Jira story points or Jira priority?

**Recommendation:** Priority (simpler, always present). Story points can be added as an option in Phase 2.

**Blocked:** Epic 19 implementation

**Source:** `docs/research/jira-integration-research.md`

---

### M5. Apple Reminders Watch() Polling Interval

**Question:** How frequently should the Apple Reminders adapter poll for changes?

**Recommendation:** 30-second default, configurable in provider settings.

**Blocked:** Epic 20 implementation

**Source:** `docs/research/apple-reminders-integration-research.md`

---

### M6. Apple Reminders Priority Edge Cases

**Question:** How should priority values of 0 or null from Apple Reminders be mapped?

**Recommendation:** Map 0/null to "no effort assigned" (default Effort value).

**Blocked:** Epic 20 implementation

**Source:** `docs/research/apple-reminders-integration-research.md`

---

### M7. Obsidian Multi-Vault Support

**Question:** Should the Obsidian adapter support multiple vaults simultaneously, or one vault at a time?

**Recommendation:** Single vault for Phase 1. Multi-vault is a Phase 2 feature.

**Blocked:** Epic 8 implementation

**Source:** `docs/prd/epic-details.md`

---

### M8. Link Relationship Types

**Question:** Should task cross-references support only "related" or additional types (blocks, duplicates, parent-of)?

**Recommendation:** Start with "related" only. Add "blocks" when dependency tracking is needed (Epic 13+).

**Blocked:** Detail view stories

**Source:** `internal/tui/detail_view.go`

---

### M9. Config Validation Strategy

**Question:** Who validates adapter-specific configuration? Provider constructor, registry factory, or dedicated validator?

**Recommendation:** Provider constructor validates on creation. Return clear error messages. No separate validator needed.

**Blocked:** Epic 7 (Plugin/Adapter SDK)

**Source:** `docs/research/task-source-expansion-research.md`

---

### M10. Env Var vs Config File Precedence

**Question:** Should environment variables override `config.yaml` settings, or only serve as fallback?

**Recommendation:** Env vars override config file (standard 12-factor convention).

**Blocked:** Epic 7 configuration design

---

### M11. Jira Cloud vs Server/Data Center

**Question:** Should Jira Cloud, Server, and Data Center all be supported in Phase 1?

**Recommendation:** Cloud only for Phase 1. Server/DC support in Phase 2 if demand exists.

**Blocked:** Epic 19 implementation scope

**Source:** `docs/research/jira-integration-research.md`

---

### M12. Story Automation Readiness Field

**Question:** Should existing story files be updated to include an `Automation: yes|no` field, or only apply to new stories?

**Recommendation:** Add to new stories only. Backfilling completed stories adds no value.

**Blocked:** Self-driving pipeline implementation

**Source:** `docs/research/multiclaude-auto-execution-research.md`

---

### M13. Sync Log Retention Policy

**Question:** How long should sync logs be retained? What rotation policy?

**Recommendation:** 7 days of sync logs, rotated daily. Configurable via config.yaml.

**Blocked:** Epic 11 (Sync Observability)

---

### M14. CI Coverage Thresholds

**Question:** What are the target code coverage thresholds? Global? Per-package? Regression tolerance?

**Recommendation:** 70% global minimum, no per-package targets, 0% regression tolerance (coverage must not decrease).

**Blocked:** Epic 9 (Testing Strategy)

**Source:** `docs/prd/requirements.md`

---

## Low Priority Decisions (Future, Can Be Deferred)

### L1. Blocker Auto-Unblock Behavior

**Question:** Should blockers auto-clear when their condition is resolved, or require manual removal?

**Recommendation:** Manual removal — auto-detection is complex and error-prone.

**Source:** `internal/tui/detail_view.go`

---

### L2. LLM Decomposition Output Format

**Question:** What format should `[G]enerate stories` produce? YAML tasks, BMAD story files, or git-committed stories?

**Recommendation:** YAML tasks first (native format). BMAD story generation is a separate feature.

**Source:** `internal/tui/detail_view.go`

---

### L3. Cross-Computer Sync Timeline

**Question:** When should cross-computer sync be reconsidered?

**Recommendation:** After CloudKit integration (Epic 20) proves stable. Not before Phase 4.

**Source:** `docs/prd/product-scope.md`

---

### L4. Web Interface Reconsideration

**Question:** Under what conditions would a web interface become worth building?

**Recommendation:** Only if mobile app (Epic 16) proves that multi-platform demand exists. Not before Phase 5.

**Source:** `docs/prd/product-scope.md`

---

### L5. OAuth Support Timeline

**Question:** When should OAuth 2.0 support be added for API-based adapters?

**Recommendation:** When the first adapter that requires it (likely MS To Do or Google Tasks) is prioritized.

**Source:** `docs/research/task-source-expansion-research.md`

---

### L6. iPhone App Architecture

**Question:** Should the iPhone app share the Apple Notes backend, or use a dedicated mobile persistence layer?

**Recommendation:** Shared Apple Notes backend via CloudKit. Dedicated persistence only if CloudKit proves insufficient.

**Blocked:** Epic 16 (far future)

**Source:** `docs/prd/epic-details.md`

---

### L7. Webhook Support Trigger

**Question:** When should webhook support (push-based sync) be added?

**Recommendation:** When an HTTP-based provider (Todoist, Linear) with webhook support is implemented.

**Source:** `docs/research/sync-architecture-scaling-research.md`

---

### L8. Emergency Quality Gate Override

**Question:** Can critical fixes bypass lint/test quality gates? Under what conditions?

**Recommendation:** No overrides. If a fix is critical, it can still pass lint. This prevents bad habits.

**Source:** `docs/prd/requirements.md`

---

### L9. Per-Door Theming

**Question:** Should users be able to theme individual doors differently (rejected in Epic 17, but may resurface)?

**Recommendation:** Keep rejected. Single theme is simpler and more cohesive.

**Source:** `docs/prd/epics-and-stories.md` (Epic 17)

---

### L10. Adapter Cache TTL Standardization

**Question:** Should cache TTL values be standardized across adapters (currently inconsistent: 5 min file-based, 15 min network)?

**Recommendation:** Standardize: 5 min for local, 60s for network (with configurable override).

**Source:** `docs/research/task-source-expansion-research.md`, `docs/research/sync-architecture-scaling-research.md`

---

### L11. Auto-Execution Cost Control

**Question:** What is the maximum number of stories per automated multiclaude run?

**Recommendation:** 3 stories per run. Allows review between batches.

**Source:** `docs/research/multiclaude-auto-execution-research.md`

---

## Decision Dependency Map

Some decisions are interconnected. Resolving them in order avoids rework:

```
C1 (Task Model) ──→ C5 (DueDate) ──→ Epic 20
                ──→ C7 (SourceRef Migration) ──→ Epic 21
                ──→ C6 (Dedup) ──→ Epic 13

C2 (Sync Scheduler) ──→ C3 (Conflict Resolution) ──→ Epic 21
                     ──→ H3 (Circuit Breaker Config) ──→ H4 (CB TUI)

C4 (Jira Scope) ──→ H5 (Contract Tests) ──→ Epic 19
H8 (Adapter Priority) ──→ H2 (SDK Scope) ──→ Epics 7, 8, 19, 20
H7 (Auto-Execution) ──→ M12 (Automation Field) ──→ Pipeline
```

**Recommended workshop order for /party-mode session:**
1. C1 → C5 → C7 (Task model chain)
2. C2 → C3 → H3 → H4 (Sync architecture chain)
3. C4 → C6 → H5 (Integration scope chain)
4. H8 → H2 (Adapter priority chain)
5. H7 → M12 (Automation chain)
6. Remaining decisions by priority
