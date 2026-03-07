# Next Phase Prioritization: CLI, MCP/LLM, iPhone App

**Date:** 2026-03-07
**Status:** Research Complete
**Method:** BMAD Party Mode multi-agent prioritization (7 agents)
**Participants:** John (PM), Winston (Architect), Victor (Innovation Strategist), Amelia (Dev), Murat (TEA), Quinn (QA), Bob (SM)

---

## Executive Summary

With all existing epics (0-22) complete across 164+ merged PRs, ThreeDoors is ready for its next major phase. Three candidate epics were evaluated: CLI interface (Epic 23), MCP/LLM integration (Epic 24), and iPhone app (Epic 16).

**Unanimous recommendation: CLI (23) -> MCP (24) -> iPhone (16)**

The CLI and MCP epics share a dependency chain, build on the existing Go codebase, and together create a strategic moat as the first task manager with MCP-based LLM integration. The iPhone app is an independent parallel effort requiring a new language, toolchain, and deployment pipeline.

---

## Prioritized Roadmap

### Phase 1: Epic 23 — CLI Interface (Immediate, Next)

**Priority:** P0 — Ship first
**Estimated effort:** 5-8 stories (Medium)
**Technical risk:** Low

#### Rationale

- `internal/core` is ~90% CLI-ready with zero Bubbletea dependencies
- All domain types already carry JSON struct tags
- Cobra is a proven, well-adopted framework (used by gh, kubectl, docker, hugo)
- `--json` output becomes the foundation for MCP integration (Epic 24)
- Additive to existing binary — no breaking changes, no regression risk
- Team skill match: Go (existing expertise)

#### Scope (MVP — Phase 1 of CLI)

8 core commands that enable both human power users and LLM agents:

| Command | Purpose |
|---------|---------|
| `threedoors doors` | Get 3 random tasks (signature command) |
| `threedoors task list` | List tasks with `--status`, `--type`, `--effort` filters |
| `threedoors task show <id>` | Show task detail (supports prefix match) |
| `threedoors task add "text"` | Add task with optional `--context` |
| `threedoors task complete <id>` | Mark task complete (batch via variadic IDs) |
| `threedoors task block <id>` | Mark blocked with `--reason` |
| `threedoors health` | Full system health check |
| `threedoors version` | Version info |

All commands support `--json` for structured output with schema versioning.

#### New Files

- `internal/cli/root.go` — Cobra root command, `--json` persistent flag
- `internal/cli/task.go` — task subcommands
- `internal/cli/doors.go` — doors command
- `internal/cli/health.go` — health command
- `internal/cli/output.go` — JSON/table output formatter
- Tests for each (`*_test.go`)

#### Modified Files

- `cmd/threedoors/main.go` — subcommand detection and CLI routing
- `internal/core/task_pool.go` — add `FindByPrefix()` method
- `go.mod` — add Cobra dependency

#### Key Design Decisions

- **Backward compatible**: `threedoors` with no args still launches the TUI
- **Non-interactive by default**: CLI commands print and exit; `--interactive` opt-in for humans
- **ID prefix matching**: Short prefixes like git SHAs for human usability
- **Exit codes 0-5**: Machine-parseable error handling
- **JSON envelope**: `schema_version`, `command`, `data`, `metadata` in all responses

#### Dependencies

- `github.com/spf13/cobra`
- No changes to `internal/tui/` (zero coupling)

#### Research Reference

- Full design: `docs/research/cli-interface-design.md`

---

### Phase 2: Epic 24 — MCP/LLM Integration (After CLI Ships)

**Priority:** P1 — Ship second
**Estimated effort:** 5-7 stories (Medium)
**Technical risk:** Medium (MCP SDK maturity)

#### Rationale

- Reuses CLI's JSON schemas directly — minimal new contract design
- `TaskProvider` interface maps 1:1 to MCP resources
- Proposal/approval pattern for writes is the main new work
- Competitive advantage: no existing task manager offers MCP integration with multi-provider aggregation
- Creates a platform play — ThreeDoors becomes infrastructure for AI coding agents
- Same Go codebase, same team skills

#### Scope

| Component | Purpose |
|-----------|---------|
| MCP server (`threedoors mcp serve`) | stdio transport for LLM clients |
| Read-only resources | Tasks, providers, session history, analytics |
| Controlled tools | Query tasks, propose enrichments, analyze patterns |
| Guardrails | LLMs propose, users approve — never direct writes |
| Prompt templates | Common task analysis queries |

#### Architecture

```
MCP Server (internal/mcp/)
    |
    +---> Security Middleware (proposal/approval)
    |
    +---> internal/core/ (shared domain, same as CLI and TUI)
```

#### Key Design Decisions

- **LLMs never directly edit task data** — all modifications flow through proposal/approval
- **Read-heavy by design** — most MCP resources are read-only
- **JSON schemas shared with CLI** — no duplication
- **Cross-provider queries** — unique capability ("What's blocking me across all task sources?")

#### Dependencies

- Epic 23 (CLI) must ship first — JSON output schemas are the foundation
- MCP Go SDK (or lightweight custom implementation)

#### Research Reference

- Full design: `docs/research/llm-integration-mcp.md`

---

### Phase 3: Epic 16 — iPhone App (After MCP Ships)

**Priority:** P2 — Ship third
**Estimated effort:** 10-15 stories (Large)
**Technical risk:** High

#### Rationale

- Entirely new codebase (Swift/SwiftUI), new toolchain (Xcode), new deployment (App Store)
- Protocol-level sharing, not code sharing — ground-up build of ~2,500 lines of domain logic
- Apple Notes sync already works via iCloud — mobile app benefits from existing infrastructure
- Independent of CLI/MCP — no blocking dependency, but also no synergy
- Requires Apple Developer Program enrollment ($99/year)
- Different skill set (Swift vs Go)

#### Scope (MVP)

| Feature | Purpose |
|---------|---------|
| Three Doors display | Swipeable card carousel |
| Open a door | Tap to see task detail |
| Mark complete | One tap or swipe right |
| Refresh doors | Pull-to-refresh |
| Apple Notes sync | Read tasks from shared Apple Notes |
| Status changes | Blocked, in-progress, complete |
| Session metrics | Basic local tracking |

#### Architecture

- Native SwiftUI targeting iOS 17+
- Apple Notes as single source of truth (same note the TUI uses)
- iCloud Drive for config/metrics sync
- No Go code sharing — port interfaces and algorithms to Swift

#### Key Risks

| Risk | Mitigation |
|------|------------|
| Swift skill gap | Research spike during Phase 2 |
| App Store review | Minimal privacy footprint, no IAP |
| Apple Notes API access from Swift | Validate during research spike |
| Two codebases to maintain | Shared specifications, not shared code |

#### Recommended: Research Spike During Phase 2

While Epic 24 (MCP) is in development, run a parallel research spike:
- Scaffold Xcode project
- Validate Apple Notes access from Swift
- Prototype Three Doors card carousel in SwiftUI
- Estimate true effort for MVP

This eliminates cold-start delay when Phase 3 begins.

#### Research Reference

- Full design: `docs/research/mobile-app-research.md`

---

## Comparative Analysis

### Effort and Risk Matrix

| Factor | CLI (Epic 23) | MCP (Epic 24) | iPhone (Epic 16) |
|--------|---------------|---------------|------------------|
| New files | ~8 in `internal/cli/` | ~6 in `internal/mcp/` | Entire Xcode project |
| Modified files | 3 (main.go, task_pool.go, go.mod) | 2 (main.go, go.mod) | 0 Go files |
| Language | Go (existing) | Go (existing) | Swift (new) |
| Technical uncertainty | Low | Medium | High |
| Testability | High (golden tests, exit codes) | High (MCP test harness) | Medium (XCTest, new stack) |
| Regression risk | Minimal (additive) | Minimal (additive) | Zero (separate codebase) |
| Deployment complexity | None (same binary) | None (same binary) | App Store review, code signing |
| Team skill match | Go | Go | Swift |

### Dependency Graph

```
Epic 23 (CLI)
    |
    |  JSON schemas, output contracts
    v
Epic 24 (MCP Server)
    |
    |  (no dependency)
    v
Epic 16 (iPhone App)  <-- independent, can start after either
```

### Strategic Value

| Dimension | CLI (23) | MCP (24) | iPhone (16) |
|-----------|----------|----------|-------------|
| User reach | Power users, scripters | LLM agents, AI tools | Mobile users |
| Competitive moat | Low (many CLIs exist) | High (no MCP task managers) | Low (many task apps) |
| Platform potential | Foundation for MCP | Enables ecosystem | Standalone app |
| Revenue potential | None directly | Integration partnerships | App Store |

---

## Decision Summary

The prioritization is driven by three factors:

1. **Dependency ordering**: CLI output schemas are the foundation for MCP. Building MCP without CLI would require designing the same JSON contracts twice.

2. **Effort efficiency**: CLI + MCP together (~10-15 stories in Go) are roughly equivalent to the iPhone app alone (~10-15 stories in Swift). The Go epics leverage the existing codebase; the iPhone app starts from scratch.

3. **Strategic positioning**: The CLI+MCP combination creates a unique market position — the first task manager with native LLM integration via MCP. This is a category-defining capability that compounds in value as AI coding agents proliferate.

**Final order: Epic 23 (CLI) -> Epic 24 (MCP) -> Epic 16 (iPhone)**

---

## Sources

- `docs/research/cli-interface-design.md` — CLI architecture and command taxonomy
- `docs/research/llm-integration-mcp.md` — MCP server design and LLM integration patterns
- `docs/research/mobile-app-research.md` — iPhone app technology and UX decisions
- `docs/prd/epics-and-stories.md` — Current epic status and requirements inventory
