# Introduction

This document outlines the overall project architecture for **ThreeDoors**, including backend systems, shared services, and non-UI specific concerns. Its primary goal is to serve as the guiding architectural blueprint for AI-driven development, ensuring consistency and adherence to chosen patterns and technologies.

**Relationship to Frontend Architecture:**
Since ThreeDoors is a CLI/TUI application with no separate web/mobile frontend, this document serves as the complete architectural specification. The TUI layer (using Bubbletea) is treated as part of the presentation layer within this architecture.

## Starter Template or Existing Project

**Decision: No Starter Template**

After reviewing the PRD, this is a **greenfield project** with the following characteristics:

- **Language & Framework:** Go 1.25.4+ with Bubbletea/Charm Bracelet ecosystem
- **Project Type:** CLI/TUI application (not a web framework starter)
- **Approach:** Built from scratch using `go mod init`

**Rationale:**
- Go CLI applications don't typically use starter templates beyond `go mod init`
- Bubbletea provides the TUI framework but doesn't impose project structure
- The PRD specifies a simple, custom structure appropriate for the Technical Demo phase
- Building from scratch aligns with the "no abstractions yet" principle for rapid validation

**Project Structure (from PRD):**
```
ThreeDoors/
├── cmd/threedoors/        # Main application entry point
├── internal/
│   ├── tui/              # Bubbletea Three Doors interface
│   └── tasks/            # Simple file I/O (read tasks.txt, write completed.txt)
├── docs/                  # Documentation (including PRD)
├── .bmad-core/           # BMAD methodology artifacts
├── Makefile              # Simple build: build, run, clean
└── README.md             # Quick start guide
```

## Change Log

| Date | Version | Description | Author |
|------|---------|-------------|--------|
| 2025-11-07 | 1.0 | Initial architecture document for Technical Demo with full task management (status, notes, blocker tracking) | Winston (Architect) |
| 2025-11-08 | 1.2 | Incorporated user feedback: enhanced UX with new key bindings (arrow keys), dynamic door sizing, no initial selection, removal of "Door X" labels, and introduction of new task management key bindings (c, b, i, e, f, p) for future implementation. Noted architectural implications for future task management features. | Bob (SM Agent) |
| 2026-03-02 | 2.0 | Major architecture update for 9 PRD recommendations (party mode consensus): added post-MVP five-layer architecture, adapter registry/plugin SDK, Obsidian adapter, Apple Notes adapter, sync engine with offline-first queue and conflict resolution, calendar awareness (local-first, no OAuth), multi-source task aggregation with dedup, LLM task decomposition agent queue, onboarding flow, new data models (ProviderConfig, ChangeEvent, SyncState, CalendarEvent, StorySpec, EnrichmentRecord), expanded tech stack (SQLite, fsnotify, Markdown parser, .ics parser), new workflows (multi-provider startup, sync with conflict detection, LLM decomposition, calendar-aware door selection), expanded source tree, and updated storage schemas (config.yaml, enrichment.db, sync-state/). | Architect Agent |

---
