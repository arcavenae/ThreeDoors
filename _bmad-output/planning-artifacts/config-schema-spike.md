# Spike Report: Config.yaml Schema & Migration (Story 3.5.3)

## Executive Summary

This spike audits the current configuration landscape of ThreeDoors and proposes a unified config.yaml schema design for Epic 7's config-driven provider selection. The key finding is that **the config infrastructure already exists** — `ProviderConfig`, `Registry`, multi-provider support, and sample config generation are all implemented. This spike documents the current state, identifies gaps, proposes the canonical schema, and validates the zero-friction migration path.

**Recommendation:** The existing config.yaml schema is sound and ready for Epic 7. The primary work remaining is documentation (adapter developer guide — Story 7.3) and minor schema formalization (adding `schema_version` for future migrations). No breaking changes are required.

## Current Configuration Audit

### Config File Location

All configuration lives in `~/.threedoors/config.yaml`. Path management is centralized in `internal/core/config_paths.go`:

```go
const configDir = ".threedoors"   // ~/.threedoors/
```

Key functions: `GetConfigDirPath()`, `EnsureConfigDir()`, `GetTasksFilePath()`

### Data Files in ~/.threedoors/

| File | Format | Purpose |
|------|--------|---------|
| `config.yaml` | YAML | All application configuration |
| `tasks.yaml` | YAML | Local task storage (textfile provider) |
| `completed.txt` | Line-delimited text | Completion history log |
| `sessions.jsonl` | JSON Lines | Session metrics |
| `patterns.json` | JSON | Cached pattern analysis results |
| `enrichment.db` | SQLite | Task enrichment database |
| `sync-queue.jsonl` | JSON Lines | Write-ahead log for offline-first sync |

### Config Sections Currently in config.yaml

The config.yaml file is read by multiple independent loaders. Each loader unmarshals only its own section, ignoring unknown fields (standard YAML behavior):

| Section | Go Type | Package | Loader |
|---------|---------|---------|--------|
| `provider` | `string` | `core` | `LoadProviderConfig()` |
| `note_title` | `string` | `core` | `LoadProviderConfig()` |
| `providers` | `[]ProviderEntry` | `core` | `LoadProviderConfig()` |
| `llm` | `llm.Config` | `intelligence/llm` | `LoadProviderConfig()` (embedded) |
| `values` | `[]string` | `core` | `LoadValuesConfig()` |
| `onboarding_complete` | `bool` | `core` | `IsFirstRun()` / `MarkOnboardingComplete()` |
| `calendar` | `calendar.Config` | `calendar` | `calendar.LoadConfig()` |

### Current Provider Infrastructure

**Registry** (`internal/core/registry.go`):
- Central `Registry` with factory pattern: `Register(name, factory)` / `InitProvider(name, config)`
- Global singleton via `DefaultRegistry()`
- Thread-safe (sync.RWMutex)
- Three built-in adapters registered at startup: `textfile`, `applenotes`, `obsidian`

**Provider Resolution** (`internal/core/provider_config.go`):
- `ResolveActiveProvider(cfg, reg)` — single-provider mode (backward compatible)
- `ResolveAllProviders(cfg, reg)` — multi-provider mode via `MultiSourceAggregator`
- Falls back to `textfile` when no config present

**Advanced Providers**:
- `FallbackProvider` — wraps primary with graceful degradation to fallback
- `WALProvider` — write-ahead log for offline-first reliability
- `MultiSourceAggregator` — merges tasks from multiple providers

## Proposed Canonical Schema

### Complete config.yaml

```yaml
# ThreeDoors Configuration
# Schema version for future migration support
schema_version: 1

# --- Provider Configuration ---
# Simple mode: single provider (backward compatible)
provider: textfile
note_title: ThreeDoors Tasks

# Advanced mode: multiple providers (overrides simple mode when present)
# providers:
#   - name: textfile
#     settings:
#       task_file: ~/.threedoors/tasks.yaml
#
#   - name: applenotes
#     settings:
#       note_title: ThreeDoors Tasks
#
#   - name: obsidian
#     settings:
#       vault_path: /path/to/your/vault
#       tasks_folder: tasks
#       file_pattern: "*.md"
#       daily_notes: true
#       daily_notes_folder: Daily
#       daily_notes_heading: "## Tasks"
#       daily_notes_format: "2006-01-02.md"

# --- User Values & Goals ---
# values:
#   - "Focus on health"
#   - "Build meaningful relationships"

# --- Calendar Awareness ---
# calendar:
#   enabled: true
#   sources:
#     - type: applescript
#     - type: ics
#       path: /path/to/calendar.ics
#     - type: caldav_cache

# --- LLM Task Decomposition ---
# llm:
#   backend: claude           # or "ollama"
#   claude:
#     model: claude-sonnet-4-20250514
#   ollama:
#     endpoint: http://localhost:11434
#     model: llama3.2
#   decomposition:
#     output_repo: /path/to/repo
#     output_branch_prefix: story/

# --- Internal State (managed by app) ---
# onboarding_complete: true
```

### Schema Design Decisions

**1. Dual-mode provider config (simple + advanced)**

The current design supports both a legacy flat `provider:` field and a modern `providers:` list. This is the correct approach:

- **Simple mode:** `provider: textfile` — for users who need one backend
- **Advanced mode:** `providers: [{name: textfile, settings: {...}}, ...]` — for multi-source aggregation
- **Resolution rule:** If `providers:` list is present and non-empty, it takes precedence over `provider:`

This is already implemented and tested.

**2. Settings as map[string]string**

Provider settings use `map[string]string` rather than typed structs. This is intentional:

- Adapters define their own settings keys (no core schema changes needed for new adapters)
- Settings are passed to factory functions which validate and parse as needed
- Trade-off: no compile-time type safety for settings values

**3. No credential storage in config.yaml**

Secrets are loaded from environment variables:
- `ANTHROPIC_API_KEY` for Claude API (loaded at runtime, never serialized via `yaml:"-"`)
- Provider-specific credentials should follow the same pattern

**4. Multiple independent loaders**

Different packages load their own sections from the same file. This works because YAML ignores unknown fields during unmarshal. The trade-off is that there's no single "parse all config" function — each subsystem reads what it needs.

### Schema Version Strategy

Adding `schema_version: 1` enables future migrations:

```go
type SchemaHeader struct {
    SchemaVersion int `yaml:"schema_version"`
}
```

Migration logic:
1. Read `schema_version` from config.yaml
2. If absent, treat as version 0 (pre-versioning — current behavior)
3. If version < current, run migration functions in order
4. Version 0 → 1: No changes needed (current schema is version 1)

This is low-cost to add now and prevents painful migrations later.

## Migration Path Analysis

### Scenario 1: New User (No config.yaml)

**Current behavior:** `GenerateSampleConfig()` creates a sample config with `provider: textfile` and commented examples. `LoadProviderConfig()` returns defaults if file missing.

**With schema_version:** Same behavior. Sample config gets `schema_version: 1` header. No user action needed.

**Verdict: Zero friction. No changes required.**

### Scenario 2: Existing User with Legacy Flat Config

```yaml
provider: applenotes
note_title: My Tasks
```

**Current behavior:** `LoadProviderConfig()` reads `provider` and `note_title` fields. Since `providers` list is empty, falls back to flat mode.

**With schema_version:** Same behavior. Missing `schema_version` treated as version 0, which maps to the same schema. No migration needed.

**Verdict: Zero friction. Fully backward compatible.**

### Scenario 3: Existing User with Multi-Provider Config

```yaml
providers:
  - name: textfile
  - name: applenotes
    settings:
      note_title: Work Tasks
```

**Current behavior:** `LoadProviderConfig()` reads `providers` list. `ResolveAllProviders()` initializes all configured providers.

**With schema_version:** Same behavior. Add `schema_version: 1` on next write.

**Verdict: Zero friction. No changes required.**

### Scenario 4: Config with Calendar, LLM, Values, Onboarding

```yaml
provider: textfile
onboarding_complete: true
values:
  - "Stay focused"
calendar:
  enabled: true
  sources:
    - type: applescript
llm:
  backend: claude
  decomposition:
    output_repo: /code/myproject
```

**Current behavior:** Each subsystem reads its own section independently.

**With schema_version:** No impact. Each loader continues to read its own section.

**Verdict: Zero friction. All sections coexist.**

## Breaking Changes Analysis

### No Breaking Changes Identified

The current config.yaml schema is **fully backward compatible**:

1. Missing fields default to safe values (textfile provider, empty values, disabled calendar)
2. Unknown fields are silently ignored by YAML unmarshal
3. The `providers` list and flat `provider` field coexist without conflict
4. Adding `schema_version` is additive — does not break existing configs

### Potential Future Breaking Changes

| Change | Risk | Mitigation |
|--------|------|------------|
| Removing `provider` flat field | Medium | Keep indefinitely; `providers` list takes precedence |
| Changing settings from `map[string]string` to typed structs | Low | Would require adapter-specific YAML tags; avoid this |
| Renaming `note_title` to provider-specific setting | Low | Keep as alias; applenotes factory reads both |
| Moving `onboarding_complete` to separate state file | Low | Read from both locations during migration window |

### Recommendation: No Breaking Changes Needed for Epic 7

Epic 7's three stories are already supported by the current schema:
- **Story 7.1 (Registry):** Already implemented via `Registry` in `internal/core/registry.go`
- **Story 7.2 (Config-driven selection):** Already implemented via `ProviderConfig.Providers` list
- **Story 7.3 (Developer guide):** Documentation only — no schema changes needed

## Gaps Identified

### Gap 1: No schema_version field

**Impact:** Low now, high if schema changes are needed later.
**Recommendation:** Add `schema_version` to `ProviderConfig` struct and sample config. Default to 1. Cost: ~10 lines of code.

### Gap 2: No unified config struct

Multiple packages each unmarshal their own section from config.yaml independently. There is no single struct that represents the complete config.yaml.

**Impact:** Low. Each subsystem is self-contained. The trade-off is correct for a small app.
**Recommendation:** Defer. A unified struct would add coupling between packages. Revisit if the number of config sections grows beyond 8-10.

### Gap 3: No config validation command

Users have no way to validate their config.yaml without running the app.

**Impact:** Low. The app handles invalid config gracefully (falls back to defaults with warnings).
**Recommendation:** Consider a `threedoors config validate` subcommand in a future story. Not needed for Epic 7.

### Gap 4: Sample config doesn't include all sections

`GenerateSampleConfig()` only generates provider-related config. It doesn't include calendar, LLM, values, or onboarding sections.

**Impact:** Low. Users discover these features through documentation.
**Recommendation:** Expand `GenerateSampleConfig()` to include commented examples for all sections. Nice-to-have for Epic 7.

### Gap 5: Provider settings are stringly-typed

All provider settings are `map[string]string`. There's no validation at the schema level — adapters validate at initialization time.

**Impact:** Low. Errors are caught at startup with clear error messages.
**Recommendation:** Keep current design. Typed settings would require schema changes for every new adapter, defeating the plugin architecture's purpose.

## Sample config.yaml (Complete)

See the "Proposed Canonical Schema" section above for the complete sample config.yaml with all sections documented and commented.

## Validation Test Results

The existing test suite in `internal/core/provider_config_test.go` already validates:

| Test | Status |
|------|--------|
| Load valid config | Pass |
| Missing file returns defaults (textfile) | Pass |
| Invalid YAML returns error | Pass |
| Empty file returns defaults | Pass |
| Round-trip read/write | Pass |
| Providers list parsing | Pass |
| Empty providers list | Pass |
| Backward-compatible flat provider | Pass |
| Provider entry with no settings | Pass |
| GetSetting with fallback | Pass |
| Sample config generation | Pass |
| Sample config doesn't overwrite existing | Pass |
| Resolve with providers list | Pass |
| Resolve fallback to flat provider | Pass |
| Resolve no config defaults to textfile | Pass |
| Resolve unknown provider returns error | Pass |
| Settings passed to factory | Pass |
| First provider used as primary | Pass |

**All 18 migration scenarios are covered by existing tests.** No new tests needed for the migration path validation.

## Conclusion

The ThreeDoors config.yaml infrastructure is well-designed and ready for Epic 7:

1. **Schema is sound** — dual-mode provider config (simple + advanced) with typed subsystem sections
2. **Migration path is zero-friction** — all existing configs continue to work unchanged
3. **No breaking changes** — adding `schema_version` is the only recommended change
4. **Epic 7 is unblocked** — Stories 7.1 and 7.2 are already implemented; Story 7.3 is documentation
5. **Test coverage is comprehensive** — 18 tests cover all config loading and resolution scenarios

### Action Items for Epic 7

1. Add `schema_version: 1` to `ProviderConfig` struct and sample config (minor code change)
2. Write adapter developer guide (Story 7.3) documenting the provider settings pattern
3. Expand sample config to include all sections (nice-to-have)
4. Consider `threedoors config validate` subcommand (future story)
