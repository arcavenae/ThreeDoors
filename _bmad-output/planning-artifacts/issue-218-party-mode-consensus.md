# Party Mode Consensus: Issue #218 — Nil Pointer Panic on Missing Provider

**Date:** 2026-03-08
**Participants:** John (PM), Winston (Architect), Amelia (Dev), Quinn (QA), Bob (SM), Murat (Test Architect)
**GitHub Issue:** #218

## Verdict

**Valid P0 bug.** Crash on default first-run path. Not previously reported or rejected.

## Root Cause

`loadTaskPool()` in `internal/cli/doors.go:195` calls `NewProviderFromConfig(cfg)` which returns nil when both provider resolution and textfile fallback fail. Line 198 dereferences nil → SIGSEGV. Same gap exists in MCP server at `cmd/threedoors-mcp/main.go:61`.

Note: `internal/cli/bootstrap.go:37-39` already has the correct nil check pattern.

## Agreed Fix Approach

### Immediate (Story 23.11)

1. Add nil check in `loadTaskPool()` after `NewProviderFromConfig()` — return descriptive error
2. Add nil check in MCP server initialization — return descriptive error
3. Add test coverage for nil provider scenarios in both paths
4. Error message: clear, actionable (e.g., "no task provider available; check provider configuration")

### Follow-up (Optional, Separate Story)

Refactor `NewProviderFromConfig()` to return `(TaskProvider, error)` instead of just `TaskProvider`. Eliminates nil-return anti-pattern at the source. All callers forced to handle error. Medium scope.

## Scope Boundary

**In Scope:** Nil checks, error messages, tests for doors.go and MCP server
**Out of Scope:** Signature refactor, auto-provider creation, first-run UX improvements

## Decisions Summary

| Decision | Status | Rationale | Alternatives Rejected |
|----------|--------|-----------|----------------------|
| Nil check guard in loadTaskPool() and MCP server | Adopted | Minimal fix for P0 crash; matches existing bootstrap.go pattern | Signature refactor (too large for bug fix scope) |
| Descriptive error message on nil provider | Adopted | Users need actionable guidance; silent failure is worse | Auto-creating default provider (feature scope, not bug fix) |
| Fix both CLI and MCP server paths | Adopted | Same gap, same risk — inconsistent to fix only one | CLI-only fix (leaves MCP vulnerable) |
| Table-driven tests for nil scenarios | Adopted | Consistent with project testing standards | Ad-hoc test cases (inconsistent) |
| Defer NewProviderFromConfig() signature refactor | Rejected | Too large for bug fix scope; separate story | — |

## Risk Assessment

| Provider Path | Nil Risk | Defense After Fix |
|---|---|---|
| `loadTaskPool()` single-provider | HIGH → FIXED | Nil check + error |
| `loadTaskPool()` multi-provider | LOW | Returns error (already safe) |
| `bootstrap.go` CLI | LOW | Nil check exists |
| MCP server | HIGH → FIXED | Nil check + error |

## Requirements Alignment

- **TD-NFR7:** "Gracefully handle missing or corrupted task files" — nil panic violates this
- **TD-NFR8 (new):** "Never panic due to nil provider initialization" — added to PRD
- **NFR7:** "Fall back to local text file storage" — fallback exists but fails silently
- **CLAUDE.md:** "No panics in user-facing code" — explicit rule being violated
