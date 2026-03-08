# ADR-0004: Monolithic CLI Architecture

- **Status:** Accepted
- **Date:** 2025-11-07
- **Decision Makers:** Project founder
- **Related PRs:** #6, #8
- **Related ADRs:** ADR-0005 (Layered Architecture Evolution)

## Context

ThreeDoors is a personal task management tool. The architecture must support:
- Fast startup (sub-second)
- No external service dependencies
- Simple distribution (single binary)
- Local-first operation
- Future extensibility for multiple task sources

## Considered Options

1. **Monolithic CLI** — Single binary, all code in one process
2. **Client-Server** — Separate daemon for data management, CLI for interaction
3. **Microservices** — Separate services for tasks, sync, intelligence
4. **Plugin-based** — Core binary with dynamically loaded plugins

## Decision

Build a **monolithic CLI application** as a single Go binary with no external service dependencies.

## Rationale

- Single binary is the simplest distribution model
- No daemon management complexity
- No network overhead for local operations
- No port conflicts or service discovery
- Fastest startup time
- Aligns with "local-first" principle — all core features work offline

## Consequences

### Positive
- Zero infrastructure requirements — works on any macOS machine with a terminal
- No background processes consuming resources when not in use
- Simple installation: download binary, run it
- All state in `~/.threedoors/` directory — easy to backup, move, or delete

### Negative
- No real-time sync between multiple running instances
- Background sync requires the app to be running (solved via sync scheduler during sessions)
- MCP server mode (Epic 24) runs as a separate process when needed
- Self-driving pipeline (Epic 22) dispatch requires external orchestration (multiclaude)
