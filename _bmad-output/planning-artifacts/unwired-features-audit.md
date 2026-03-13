# Unwired Features Audit

**Date:** 2026-03-13
**Trigger:** Connect wizard gap (Story 44.1) raised the question: are there other features implemented but not wired to their CLI/user entry points?

## Executive Summary

Found **5 distinct gaps** across three categories: CLI stubs returning errors, adapter registration asymmetry between TUI and CLI paths, and a fully implemented adapter that is never registered anywhere.

## Gap 1: Connect Wizard — TUI Built, CLI Stubbed (KNOWN)

**Severity:** Medium
**Files:** `internal/cli/connect.go:101`, `internal/tui/connect_wizard.go`

The TUI connect wizard (830 lines, 35+ tests) is fully functional via `:connect` in the TUI. The CLI `threedoors connect <provider>` returns a hard-coded error when called without `--label`:

```go
return fmt.Errorf("interactive wizard not yet available (Story 44.1), use flags: --label, --token, etc")
```

Already documented in `connect-wizard-gap-research.md`. Fix: launch a minimal Bubbletea program wrapping the existing `ConnectWizard` model.

## Gap 2: CLI Commands Skip Adapter Registration (CRITICAL)

**Severity:** High
**File:** `cmd/threedoors/main.go:43-48`

The `registerBuiltinAdapters()` function registers all 8 provider adapters (textfile, applenotes, jira, github, obsidian, reminders, todoist, linear) with the global `DefaultRegistry()`. However, this registration **only happens in the TUI code path**:

```go
// Line 43-44: CLI path — exits BEFORE registration
if len(os.Args) > 1 && isSubcommand(os.Args[1]) && !isPlanMode {
    os.Exit(cli.Execute())
}

// Line 48: TUI path — registration happens here
registerBuiltinAdapters(core.DefaultRegistry())
```

**Impact:** Every CLI command that calls `bootstrap()` or `core.DefaultRegistry()` works against an **empty registry**. This means:
- `threedoors task list` — fails for any user with a non-textfile provider (the textfile fallback in `NewProviderFromConfig` also fails since "textfile" isn't registered)
- `threedoors doors` — same issue
- `threedoors stats` — works (reads JSONL directly, doesn't use registry)
- `threedoors sources` — partially works (uses `connection.ResolveFromConfig` which has its own adapter resolution)
- `threedoors connect todoist --label X --token Y` — works (uses `ConnectionService.Add`, doesn't need registry)

**Why it hasn't been caught:** The CLI commands that exercise the registry (`task`, `doors`) are less commonly used than the TUI. Users who run `threedoors` without arguments (TUI mode) never hit this bug. The `textfile` provider happens to be the fallback, but it too fails because it's not registered in the empty registry.

**Fix:** Move `registerBuiltinAdapters(core.DefaultRegistry())` to before the CLI/TUI routing branch, or call it inside `cli.Execute()`.

## Gap 3: ClickUp Adapter — Implemented, Never Registered

**Severity:** Low
**Files:** `internal/adapters/clickup/` (3 source files + 3 test files)

The ClickUp adapter (`clickup_provider.go`, `clickup_client.go`, `config.go`) is fully implemented with tests but is:
1. **Not registered** in `registerBuiltinAdapters()` in `main.go`
2. **Not listed** in the CLI connect command's `ValidArgs` or `knownProviderSpecs`
3. **Not listed** in the TUI connect wizard's `DefaultProviderSpecs()`

The adapter was built as part of Epic 63 (task source expansion) but the registration wiring was never completed. Stories 63.1-63.4 track ClickUp integration, but the adapter exists in an orphaned state — it compiles and passes contract tests but is inaccessible to users.

## Gap 4: CLI Connect Provider Mismatch

**Severity:** Low
**File:** `internal/cli/connect.go:38-68` vs `internal/tui/connect_wizard.go`

The CLI connect command's `knownProviderSpecs` map covers **4 providers** (todoist, github, jira, textfile), while:
- The CLI's own `ValidArgs` lists **7 providers** (adds applenotes, obsidian, reminders)
- The TUI wizard's `DefaultProviderSpecs()` covers **8 providers** (adds linear)
- The adapter registry registers **8 providers** (same 8 as TUI)

If a user runs `threedoors connect obsidian --label "My Vault" --path /path`, the command accepts "obsidian" as a valid arg but then treats it as an unknown provider — all flag values are passed through as raw settings without validation. It technically works via the catch-all in `buildConnectSettings`, but required flags (like `--path` for obsidian's `vault_path`) are not enforced.

## Gap 5: TUI-Only Features Without CLI Equivalents

**Severity:** Informational (by design, not a bug)

These TUI features have no CLI command equivalents. This is expected since many are interactive-only, but documenting for completeness:

| TUI Feature | TUI Access | CLI Command | Status |
|---|---|---|---|
| Task breakdown | `:breakdown` | — | No CLI equivalent |
| Bug report | `:bug` | — | No CLI equivalent |
| Dev queue | `:devqueue` | — | No CLI equivalent |
| Dispatch view | `:dispatch` | — | No CLI equivalent |
| Task enrichment | `:enrich` | — | No CLI equivalent |
| Import tasks | `:import` | — | No CLI equivalent (but `extract` exists) |
| Insights dashboard | `:insights`/`:dashboard` | — | No CLI equivalent |
| Orphaned tasks | `:orphaned` | — | No CLI equivalent |
| Sync log | `:synclog` | — | No CLI equivalent |
| Theme picker | `:theme` | — | No CLI equivalent |
| Deferred tasks | `:deferred` | — | No CLI equivalent |
| AI suggestions | `:suggestions` | — | No CLI equivalent |
| Values/goals | `:goals` | — | No CLI equivalent |
| Seasonal themes | `:seasonal` | — | No CLI equivalent |
| Mood recording | `:mood` | `threedoors mood set` | **Wired** |
| Stats | `:stats` | `threedoors stats` | **Wired** |
| LLM status | `:llm-status` | `threedoors llm status` | **Wired** |
| Sources list | `:sources` | `threedoors sources` | **Wired** |
| Connect wizard | `:connect` | `threedoors connect` | **Stubbed** (Gap 1) |
| Extract tasks | `:extract` | `threedoors extract` | **Wired** |
| Health check | `:health` | `threedoors doctor` | **Wired** |
| Tag editing | `:tag` | — | No CLI equivalent |

Most TUI-only features are appropriately interactive-only. The ones most likely to benefit from CLI equivalents would be `:import` (for scripting), `:deferred` (for listing snoozed tasks), and `:orphaned` (for task hygiene scripts).

## Prioritized Recommendations

1. **Gap 2 (CLI adapter registration)** — **Fix immediately.** This is a functional bug: CLI commands fail silently for non-textfile users. One-line fix: move `registerBuiltinAdapters()` call above the CLI routing.

2. **Gap 1 (Connect wizard)** — **Small integration story.** Wire the existing TUI wizard to the CLI path. ~1-2 hours, well-documented in `connect-wizard-gap-research.md`.

3. **Gap 4 (Provider spec mismatch)** — **Fix alongside Gap 1.** When wiring the wizard, also update `knownProviderSpecs` to cover all 8 registered providers.

4. **Gap 3 (ClickUp adapter)** — **Depends on Epic 63 status.** Either register it in `registerBuiltinAdapters()` and add to connect command, or mark Epic 63 stories as blocked/deferred.

5. **Gap 5 (TUI-only features)** — **No action needed.** These are interactive features that don't have natural CLI equivalents. Future CLI additions should be driven by user requests, not completeness.

## Methodology

- Searched `internal/cli/*.go` for `not yet available`, `TODO`, `stub`, `placeholder` patterns
- Cross-referenced all CLI commands in `root.go` against TUI commands in `commandRegistry`
- Traced adapter registration from `main.go` through both CLI and TUI code paths
- Listed all `internal/adapters/` directories and verified each appears in `registerBuiltinAdapters()`
- Compared `knownProviderSpecs` (CLI) vs `ValidArgs` (CLI) vs `DefaultProviderSpecs` (TUI) vs registered adapters
