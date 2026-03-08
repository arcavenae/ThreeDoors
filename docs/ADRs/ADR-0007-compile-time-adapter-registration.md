# ADR-0007: Compile-Time Adapter Registration

- **Status:** Accepted
- **Date:** 2026-01-15 (Epic 7)
- **Decision Makers:** Design decision H2
- **Related PRs:** #68 (Story 7.1), #70 (Story 7.2)
- **Related ADRs:** ADR-0006 (TaskProvider Interface)

## Context

With multiple task providers, the system needs a mechanism to discover and instantiate adapters. Three approaches were considered.

## Considered Options

1. **Compile-time registration** — Adapters are Go packages imported in `main.go`
2. **Config-driven factory** — YAML config specifies adapter name + settings; factory creates instances
3. **Plugin system** — Go plugins or subprocess-based plugins loaded at runtime

## Decision

Use **compile-time registration** (Option A). Adapters are Go packages registered via import in `main.go`. Configuration in `config.yaml` controls which registered adapters are active and their settings.

Evolve to config-driven factory (Option B) when 3+ adapters exist (this threshold has been reached, but the compile-time approach has proven sufficient).

## Rationale

- Simplest implementation — no reflection, no dynamic loading
- Go's plugin system is fragile on macOS (CGO issues, version coupling)
- All current adapters are first-party — no third-party adapter ecosystem yet
- Config-driven activation (which adapters are enabled) is separate from code-level registration

## Consequences

### Positive
- Type safety at compile time — no runtime adapter loading failures
- Easy to understand — follow the imports to see available adapters
- No plugin ABI compatibility concerns
- Adapter developer guide (PR #72) documents the pattern clearly

### Negative
- Adding a new adapter requires recompilation
- Cannot distribute adapter-only updates independently
- All adapters compiled into binary even if unused (minimal binary size impact)
