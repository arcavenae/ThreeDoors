# ADR-0022: CLI Interface with Cobra

- **Status:** Accepted
- **Date:** 2026-03-01
- **Decision Makers:** Research (PR #154), architecture review
- **Related PRs:** #154, #168, #170-#173, #182, #188-#190, #192, #194-#195

## Context

ThreeDoors originally launched directly into TUI mode. Power users and automation scripts need non-interactive access to task management. A CLI interface complements the TUI.

## Decision

Add a **Cobra-based CLI** alongside the existing TUI:

- `threedoors` (no args) — launches TUI (backward compatible)
- `threedoors task list|show|add|complete|edit|delete|search` — task management
- `threedoors task block|unblock` — blocking operations
- `threedoors doors` — CLI three-doors experience
- `threedoors mood|stats` — analytics
- `threedoors config` — configuration management
- `threedoors health|version` — diagnostics
- Supports `--json` flag for machine-readable output
- Supports stdin/pipe input for scripting

## Rationale

- Cobra is the standard Go CLI framework (used by kubectl, Hugo, GitHub CLI)
- Subcommand structure scales to many operations
- `--json` output enables integration with `jq`, scripts, and other tools
- Stdin support enables piped workflows (`cat tasks.txt | threedoors task add`)
- Backward compatible — bare `threedoors` command still launches TUI

## Consequences

### Positive
- Scriptable task management for power users
- Machine-readable output for automation
- Consistent subcommand structure
- Exit codes enforce Unix conventions (0=success, 1=error, 2=usage)

### Negative
- CLI and TUI share code but have different output paths
- JSON output formatting adds maintenance per command
- Race condition risk from global state (resolved in PR #192)
- 10 stories (Epic 23) — significant implementation effort
