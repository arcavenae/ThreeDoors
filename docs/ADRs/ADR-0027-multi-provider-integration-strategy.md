# ADR-0027: Multi-Provider Integration Strategy

- **Status:** Accepted
- **Date:** 2026-02-01
- **Decision Makers:** Design decisions C4, H8, M3, M4, M11
- **Related PRs:** #73-#75, #132, #137-#139, #148-#153, #155, #158, #201-#205
- **Related ADRs:** ADR-0006 (TaskProvider), ADR-0007 (Registration), ADR-0015 (Dedup)

## Context

ThreeDoors integrates with 6 external task sources. Each integration has unique API characteristics, field mappings, and phasing decisions. A consistent strategy was needed.

## Decision

Apply a **consistent three-phase integration pattern** for each provider:

1. **Phase 1: Read-only** — Pull tasks into ThreeDoors, map fields, display in doors
2. **Phase 2: Bidirectional sync** — Push status changes and edits back to source
3. **Phase 3: Advanced features** — Provider-specific capabilities (webhooks, rich content)

### Provider-Specific Decisions

| Provider | Phase 1 Scope | Field Mapping | Special Handling |
|----------|--------------|---------------|-----------------|
| **Apple Notes** | Read tasks from Notes | Text → Task name, body → Context | AppleScript bridge (JXA) |
| **Obsidian** | Read from vault markdown | Checkbox items → Tasks | fsnotify for Watch() |
| **Jira** | Cloud only (Decision M11) | Priority → Effort, ADF → plain text (Decision M3) | Story points deferred (Decision M4) |
| **Apple Reminders** | Read reminders | DueDate → native field (Decision C5), priority 0/null → no effort (Decision M6) | 15min polling (Decision M5) |
| **GitHub Issues** | Read issues + labels | Labels → Categories, milestone → Context | REST API v3 |

### Integration Priority Order (Decision H8)

Apple Notes → Obsidian → Apple Reminders → Jira → GitHub Issues → Todoist (planned) → Linear (planned)

## Consequences

### Positive
- Consistent pattern reduces per-integration learning curve
- Read-only first de-risks each integration
- Contract tests validate all adapters uniformly
- Field mapping decisions documented per provider

### Negative
- Three-phase approach means full bidirectional sync takes longer per provider
- Provider-specific quirks (JXA, ADF, GraphQL) resist full standardization
- Six adapters increase maintenance burden and test matrix
