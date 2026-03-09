# Spike Report: Apple Notes Integration (Story 2.2)

## Executive Summary

This spike evaluates three approaches for integrating Apple Notes with the ThreeDoors TUI application: **AppleScript/osascript**, **direct SQLite access**, and **MCP server**. The evaluation covers read/write capability, reliability, permissions model, testability, and CI/CD compatibility.

**Recommendation: AppleScript/osascript** is the recommended approach for both reads and writes, with a hybrid fallback to text file when Apple Notes is unavailable. This approach is already partially implemented (read-only) in Story 2.3.

## Comparison Matrix

| Criterion | AppleScript/osascript | Direct SQLite | MCP Server |
|-----------|----------------------|---------------|------------|
| Read capability | Full (plaintext text property) | Partial (snippet only, full content in protobuf) | Non-Viable (no maintained server found) |
| Write capability | Full (set body of note) | Dangerous (CloudKit corruption risk) | N/A |
| Reliability | High (Apple-supported API) | Low (undocumented schema, breaks between versions) | N/A |
| Complexity | Low (shell out to osascript) | High (protobuf parsing, schema reverse-engineering) | N/A |
| Permissions | Automation consent (low friction) | Full Disk Access (high friction, invasive) | N/A |
| macOS version support | 10.10+ (broad) | Schema varies by version (fragile) | N/A |
| External dependencies | None (osascript is built-in) | modernc.org/sqlite + protobuf parser | npm + Node.js runtime |
| Latency | 100-400ms per call (acceptable) | 5-50ms per query (fast) | N/A |
| Sync safety | Safe (goes through Notes.app) | Unsafe (bypasses CloudKit sync) | N/A |
| Testability (mockability) | High (CommandExecutor interface) | High (in-memory SQLite) | N/A |
| CI/CD compatibility (Linux) | Unit tests portable (mock), integration macOS-only | Unit tests portable (mock), integration macOS-only | N/A |
| Concurrency safety | Sequential (one osascript at a time recommended) | WAL mode supports concurrent reads | N/A |

## Recommendation with Rationale

**Chosen approach: AppleScript/osascript**

Rationale:
1. **Already proven** — Story 2.3 shipped a read-only AppleNotesProvider using osascript
2. **Write support confirmed** — `set body of note` works for modifying note content
3. **Lowest permissions barrier** — Only requires Automation consent (not Full Disk Access)
4. **Sync-safe** — All operations go through Notes.app, which handles CloudKit sync properly
5. **No external dependencies** — osascript is built into every macOS installation
6. **Testable** — CommandExecutor interface pattern already enables full mocking in unit tests

For Stories 2.3–2.6, extend the existing `AppleNotesProvider` to add write support via `set body of note`.

### Hybrid Approach Note

The system should continue using the FallbackProvider pattern: AppleNotesProvider as primary, TextFileProvider as fallback. This is already implemented.

## Error Taxonomy

### AppleScript Errors
| Error | Cause | Handling |
|-------|-------|----------|
| `context.DeadlineExceeded` | osascript > 2s timeout | Fall back to TextFileProvider |
| `Can't get note "..."` | Note doesn't exist | Show error, fall back |
| `Not authorized` / `not allowed` | Automation permission denied | Show setup instructions, fall back |
| `exec.ErrNotFound` | Not on macOS | Fall back to TextFileProvider |
| `exec.ExitError` | osascript process failure | Log to stderr, fall back |
| Concurrent call conflict | Two osascript calls overlap | Serialize calls (mutex) |

### SQLite Errors (for reference)
| Error | Cause | Handling |
|-------|-------|----------|
| Permission denied | No Full Disk Access | Not usable without admin action |
| Database locked | Notes.app has write lock | Retry with backoff |
| Table not found | Schema changed between macOS versions | Approach is fragile |

## CI/CD Compatibility

- **Linux GitHub Actions (default):** Run unit tests with mocked CommandExecutor. All parsing, task management, and sync logic is testable without macOS.
- **macOS-only (`//go:build darwin`):** Integration tests that actually call osascript. These require a macOS runner or self-hosted runner.
- **Recommended CI strategy:** Default workflow runs portable tests on Linux. Optional macOS matrix job for integration tests on release branches.

## Performance Benchmarks

See `scripts/spike/benchmarks/` for raw JSON data.

Expected ranges (from PoC testing):
- **osascript read:** p50 ~150ms, p95 ~300ms, max ~500ms
- **osascript write:** p50 ~200ms, p95 ~400ms, max ~700ms
- **Notes.app cold start:** First call may take 1-2s if Notes.app is not running (it auto-launches)

All operations within the <500ms NFR6 budget for typical use.

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| osascript write strips formatting | Medium | Low | Use HTML body (`set body of note`) to preserve formatting |
| Notes.app auto-launches on osascript call | Certain | Low | Document behavior; users expect Notes to be available |
| Automation permission dialog confuses users | Low | Medium | First-run setup guide in README, `:health` command (Story 2.6) |
| macOS update breaks osascript API | Very Low | High | Monitor Apple release notes; `plaintext text` has been stable since 10.10 |
| Checkbox format not standardized | Medium | Medium | Support multiple formats: `- [ ]`, `- [x]`, plain text |

## Effort Estimate for Stories 2.3–2.6

| Story | Effort | Notes |
|-------|--------|-------|
| 2.3 Read Tasks | **Done** (PR #17 merged) | AppleNotesProvider already implemented |
| 2.4 Write Tasks | 4-6 hours | Add `set body of note` to AppleNotesProvider, handle formatting |
| 2.5 Bidirectional Sync | **Done** (PR #15 merged) | SyncEngine already implemented |
| 2.6 Health Check | 2-3 hours | Test connectivity, permissions, note existence |

## Approach Details

### AppleScript/osascript (Recommended)

**Read:** `tell application "Notes" to get plaintext text of note "ThreeDoors Tasks"`
**Write:** `tell application "Notes" to set body of note "ThreeDoors Tasks" to "<new content>"`
**PoC:** `scripts/spike/poc_applescript_read.sh`, `scripts/spike/poc_applescript_write.sh`, `scripts/spike/poc_applescript.go`

### Direct SQLite (Not Recommended)

**Path:** `~/Library/Group Containers/group.com.apple.notes/NoteStore.sqlite`
**Read:** Possible via ZSNIPPET column, but full content requires protobuf parsing of ZMERGEABLEDATA1
**Write:** Not recommended — bypasses CloudKit sync, risk of corruption
**PoC:** `scripts/spike/poc_sqlite_read.sh`

### MCP Server (Non-Viable)

No maintained MCP server for Apple Notes was found in npm registry or GitHub. Any future MCP server would likely wrap AppleScript anyway, adding unnecessary complexity.
**PoC:** `scripts/spike/poc_mcp.sh` (search results only)
