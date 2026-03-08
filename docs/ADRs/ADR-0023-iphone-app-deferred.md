# ADR-0023: iPhone App Deferred to Icebox

- **Status:** Accepted
- **Date:** 2026-03-07
- **Decision Makers:** Project founder
- **Related PRs:** #47 (Epic 16 creation), #165 (story files)

## Context

Epic 16 proposed a SwiftUI iPhone app for mobile task management. Seven stories were planned covering shared data layer, SwiftUI views, sync, notifications, and App Store distribution.

## Decision

**Defer indefinitely** (icebox). Epic 16 is not scheduled and will not be implemented unless re-entry conditions are met.

## Rationale

- No validated user demand for mobile access
- Core user persona is CLI/TUI power user on macOS
- MCP server (Epic 24) may serve mobile-adjacent use cases via LLM agents
- Adds significant platform, build, and distribution complexity
- Apple Developer Program for iOS has additional review and compliance burden
- SwiftUI development is outside the project's Go expertise

## Re-Entry Conditions

Revisit if:
1. 5+ distinct user requests for mobile access, OR
2. MCP proves insufficient for on-the-go task management

## Consequences

### Positive
- Avoids premature platform expansion
- Engineering effort focused on core value proposition
- No iOS build infrastructure to maintain
- No App Store review compliance burden

### Negative
- No mobile access to tasks
- Users must use terminal (or MCP-enabled AI app) for task management
- Could lose potential mobile-first users (not the target persona)
