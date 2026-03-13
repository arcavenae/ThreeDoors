# Connect Wizard Gap Research

**Date:** 2026-03-13
**Trigger:** `threedoors connect <provider>` (no flags) reports "interactive wizard not yet available"

## Summary

The interactive connect wizard **exists and is fully implemented** in the TUI (`internal/tui/connect_wizard.go`). The CLI connect command (`internal/cli/connect.go`) was built separately and **never wired to the wizard** — it returns a stub error instead.

The TUI wizard is accessible via `:connect` inside the TUI and works end-to-end. The gap is exclusively in the CLI entry point.

## What Exists

### TUI Wizard (Story 44.1, PR #574) — COMPLETE
- **File:** `internal/tui/connect_wizard.go` (~830 lines)
- **Tests:** `internal/tui/connect_wizard_test.go` (~1380 lines, 35+ tests)
- Full 4-step flow implemented:
  1. Provider selection (all 8 providers: todoist, github, jira, linear, obsidian, textfile, reminders, applenotes)
  2. Provider-specific config (API token, OAuth device code flow, file path)
  3. Sync configuration (mode + poll interval)
  4. Test & confirm
- OAuth device code flow with browser auto-open
- Environment token detection (GH_TOKEN, GITHUB_TOKEN, LINEAR_API_KEY)
- Provider auto-detection with `ApplyDetection()` (detected providers promoted to top)
- Path pre-fill from detection results
- Esc cancellation at any step with cleanup
- Wired into main_model.go via `ShowConnectWizardMsg` → `ViewConnectWizard`
- `:connect` TUI command registered in `commandRegistry`
- E2E test covers: entry via `:connect` → cancel with Esc → return to doors

### CLI Connect Command (Story 45.1, PR #573) — Non-Interactive Only
- **File:** `internal/cli/connect.go` (~304 lines)
- **Tests:** `internal/cli/connect_test.go`
- Works with flags: `threedoors connect todoist --label "Work" --token $TOKEN`
- Supports: todoist, github, jira, textfile (via `knownProviderSpecs`)
- JSON output mode (`--json`)
- Connection health test on creation
- **Line 101:** When `label == ""` (no flags), returns:
  ```go
  return fmt.Errorf("interactive wizard not yet available (Story 44.1), use flags: --label, --token, etc")
  ```

## What's Missing

The **only** gap is the bridge between the CLI `connect` command and the TUI wizard:

1. **No TUI launch from CLI:** When `threedoors connect <provider>` is run without `--label`, the CLI should detect it's in a terminal and launch the TUI wizard (similar to how other CLI tools drop into interactive mode). Instead, it returns a hard-coded error string.

2. **Provider mismatch:** The CLI's `knownProviderSpecs` covers 4 providers (todoist, github, jira, textfile). The TUI wizard's `DefaultProviderSpecs()` covers 8 providers (adds linear, obsidian, reminders, applenotes). The wizard has broader coverage.

3. **No `ConnectWizardCompleteMsg` handler in CLI context:** The wizard emits `ConnectWizardCompleteMsg` which is handled by the TUI's `main_model.go`. If the wizard were launched from the CLI, it would need its own handler to call `ConnectionService.Add()` and persist credentials — or just launch the full TUI in wizard mode.

## What Would Need to Be Done

### Option A: Launch Full TUI in Wizard Mode (Recommended)
When `threedoors connect <provider>` is called without `--label` and stdin is a terminal:
1. Create a Bubbletea program with the `ConnectWizard` model
2. Pre-select the provider (since it's passed as `args[0]`)
3. Run the TUI program, collect the `ConnectWizardCompleteMsg`
4. Use `ConnectionService.Add()` to persist, same as the non-interactive path

**Effort:** Small — the wizard model already implements `tea.Model`. Need a thin wrapper that creates a `tea.Program` with just the wizard (not the full ThreeDoors TUI).

### Option B: Standalone Interactive Prompts (Not Recommended)
Build a separate interactive flow using `huh` forms directly in the CLI, without Bubbletea. This would duplicate logic already in `connect_wizard.go`.

### Implementation Notes
- The `ConnectWizard` already accepts provider specs and a `ConnectionManager`
- It could be pre-advanced to skip Step 1 (provider select) since the provider is already known from the CLI arg
- Need to handle the case where `args[0]` is a provider not in `knownProviderSpecs` but IS in the wizard's `DefaultProviderSpecs()` (e.g., "linear", "obsidian")
- The `ValidArgs` on the cobra command already includes the extended set: `todoist, github, jira, textfile, applenotes, obsidian, reminders`

## Stories Referenced
- **Story 44.1** (PR #574): Setup Wizard with huh Forms — DONE, fully implemented in TUI
- **Story 45.1** (PR #573): `threedoors connect` Command (Non-Interactive) — DONE, but interactive fallback stubbed

## Recommendation

This is a small integration story — likely 1-2 hours of implementation. The wizard code is battle-tested with 35+ tests. The work is purely wiring: detect terminal, create a minimal Bubbletea program wrapping `ConnectWizard`, handle the completion message, and remove the stub error on line 101 of `connect.go`.
