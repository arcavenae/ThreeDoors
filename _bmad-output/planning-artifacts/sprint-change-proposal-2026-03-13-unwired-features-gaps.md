# Sprint Change Proposal: Fix Unwired Features — CLI Adapter Registration, ClickUp Wiring, Provider Spec Parity

**Date:** 2026-03-13
**Triggered by:** Unwired features audit (`_bmad-output/planning-artifacts/unwired-features-audit.md`)
**Severity:** High (Gap 2 is Critical; Gaps 3-4 are Low but easy to bundle)

## Problem Statement

The unwired features audit revealed three gaps where implemented code is not properly connected to its user-facing entry points:

1. **Gap 2 (CRITICAL):** `registerBuiltinAdapters()` in `cmd/threedoors/main.go` only runs in the TUI code path (line 61). CLI commands exit at line 57 via `os.Exit(cli.Execute())` before adapter registration. Any CLI command that uses `core.DefaultRegistry()` (e.g., `threedoors task list`, `threedoors doors`) operates against an **empty registry** and silently fails for non-textfile providers. This is a functional bug affecting all users with non-textfile providers who use CLI commands.

2. **Gap 3:** The ClickUp adapter (`internal/adapters/clickup/`) is fully implemented with tests (3 source files, 3 test files, passes contract tests) and is even registered in `registerBuiltinAdapters()` (main.go line 344). However, it is **not listed** in the CLI connect command's `ValidArgs` (line 90) or `knownProviderSpecs` (lines 39-69), and **not listed** in the TUI connect wizard's `DefaultProviderSpecs()`. Users cannot connect to ClickUp through either the CLI or TUI.

3. **Gap 4:** The CLI connect command's `knownProviderSpecs` map covers only 4 providers (todoist, github, jira, textfile), while the CLI's own `ValidArgs` lists 8 providers and the TUI wizard + adapter registry registers 9 (including clickup). Running `threedoors connect obsidian` accepts the provider name but doesn't enforce required flags (e.g., `--path` for `vault_path`). All flag values pass through as raw settings without validation.

## Impact Analysis

**Gap 2:** Every CLI command that uses the adapter registry fails silently for non-textfile users. The textfile provider fallback also fails because it's not registered. This affects `threedoors task list`, `threedoors doors`, and potentially other commands. Stats and connect commands happen to work because they use different code paths. The bug hasn't been caught because TUI users (the majority) are unaffected, and textfile is the default provider.

**Gap 3:** ClickUp users cannot use ThreeDoors with ClickUp despite the adapter being fully functional. Epic 63 (ClickUp Integration) is marked COMPLETE but the integration is inaccessible.

**Gap 4:** Users attempting to connect providers like obsidian, applenotes, reminders, or linear via CLI get no flag validation. This can lead to incomplete configurations that fail at runtime with unhelpful errors.

## Proposed Approach

**Single epic with 3 stories**, addressing all gaps in dependency order:

1. **Story 1: CLI Adapter Registration Fix (Gap 2)** — Move `registerBuiltinAdapters(core.DefaultRegistry())` to before the CLI/TUI routing branch in `main()`. Add regression tests that verify CLI commands work with a non-textfile provider configured. This is the critical fix.

2. **Story 2: ClickUp Connect Wiring (Gap 3)** — Add ClickUp to CLI `knownProviderSpecs` with proper flag spec, add to CLI `ValidArgs`, and add to TUI `DefaultProviderSpecs()`. Verify the full connect flow works for ClickUp via both CLI and TUI.

3. **Story 3: Provider Spec Parity & Validation (Gap 4)** — Add `knownProviderSpecs` entries for all 9 registered providers (applenotes, obsidian, reminders, linear, clickup). Ensure required flags are enforced for each. Add a test that verifies `knownProviderSpecs` keys match the registered adapter names.

## Rejected Alternatives

1. **Call registration inside `cli.Execute()`** — This would duplicate the registration call and create a maintenance risk where TUI and CLI registration diverge. Moving the single call before the routing branch is simpler and DRY.

2. **Defer ClickUp wiring to a future sprint** — The adapter is complete, tested, and Epic 63 is marked COMPLETE. Leaving it unwired is misleading and only requires adding a few map entries. No reason to defer.

3. **Fix only Gap 2 and leave Gaps 3-4** — Gaps 3-4 are small, related, and can be fixed in the same sprint. Bundling them into one epic keeps the work focused and avoids orphaned partial fixes.

4. **Auto-generate knownProviderSpecs from registry** — While elegant, this would require refactoring the flag spec to be part of the adapter interface (adding `FlagSpec()` to each adapter). This is over-engineering for the current problem. A static map with a parity test is sufficient.

## Stories Required

| Story | Title | Scope | Effort |
|-------|-------|-------|--------|
| X.1 | CLI Adapter Registration Fix | Move `registerBuiltinAdapters()` before CLI routing; regression tests | S |
| X.2 | ClickUp Connect Wiring | Add ClickUp to `knownProviderSpecs`, `ValidArgs`, `DefaultProviderSpecs()` | S |
| X.3 | Provider Spec Parity & Validation | Add all 9 providers to `knownProviderSpecs`; parity test | M |

## Effort Estimate

**Small** — Total estimated effort is 1-2 hours across all three stories. Gap 2 is a one-line fix plus tests. Gaps 3-4 are map entry additions plus validation logic.
