# Epic Number Registry

> **Purpose:** Prevents epic number collisions when multiple agents work in parallel. Check this table before assigning a new epic number. Project-watchdog is the MUTEX for epic/story number allocation. See D-112.
>
> Extracted from [BOARD.md](BOARD.md) — see [ARCHIVE.md](ARCHIVE.md) for historical decisions.

| Epic | Feature | Allocated | Status |
|------|---------|-----------|--------|
| 39 | Keybinding Display System | 2026-03-08 | Complete (12/13, 1 cancelled) |
| 40 | Beautiful Stats Display | 2026-03-08 | Complete (10/10) |
| 41 | Charm Ecosystem Adoption & TUI Polish | 2026-03-09 | Complete (6/6) |
| 42 | Application Security Hardening | 2026-03-09 | In Progress (1/5) |
| 43 | Connection Manager Infrastructure | 2026-03-09 | In Progress (2/6) |
| 44 | Sources TUI | 2026-03-09 | Not Started (0/7) |
| 45 | Sources CLI | 2026-03-09 | Not Started (0/5) |
| 46 | OAuth Device Code Flow | 2026-03-09 | Not Started (0/4) |
| 47 | Sync Lifecycle & Advanced Features | 2026-03-09 | Not Started (0/4) |
| 48 | Door-Like Doors — Visual Door Metaphor Enhancement | 2026-03-09 | Not Started (0/4) |
| 49 | ThreeDoors Doctor — Self-Diagnosis Command | 2026-03-10 | In Progress (1/10) |
| 50 | In-App Bug Reporting | 2026-03-10 | Not Started (0/3) |
| 51 | SLAES — Self-Learning Agentic Engineering System | 2026-03-10 | In Progress (0/10) |
| 52 | Envoy Three-Layer Firewall | 2026-03-10 | Not Started (0/4) |
| 53 | Remote Collaboration — multiclaude Cross-Machine Access | 2026-03-10 | Not Started (0/5) |
| 54 | Gemini Research Supervisor — Deep Research Agent Infrastructure | 2026-03-11 | Not Started (0/5) |
| 55 | CI Optimization Phase 1 | 2026-03-11 | Complete (3/3) |
| 56 | Door Visual Redesign — Three-Layer Depth System | 2026-03-11 | Not Started (0/5) |
| 57 | LLM CLI Services | 2026-03-11 | Not Started (0/8) |
| 58 | Supervisor Shift Handover — Context-Aware Supervisor Rotation | 2026-03-11 | Not Started (0/7) |
| 59 | Full-Terminal Vertical Layout | 2026-03-11 | Not Started (0/2) |
| 60 | README Overhaul | 2026-03-11 | In Progress (1/5) |
| 61 | GitHub Pages User Guide | 2026-03-11 | Not Started (0/4) |
| 62 | Retrospector Agent Reliability | 2026-03-12 | Not Started (0/3) |
| 63 | ClickUp Integration | 2026-03-13 | Not Started (0/4) |
| 64 | Cross-Computer Sync | 2026-03-13 | Not Started (0/6) |
| 65 | *(next available)* | — | — |

**Rules:**
1. Before creating a new epic, check this table for the next available number
2. Reserve the number here FIRST, before creating story files or updating ROADMAP.md
3. **Project-watchdog is the MUTEX** for epic number allocation — request numbers through supervisor/project-watchdog
4. Workers and `/plan-work` agents NEVER self-assign epic numbers — always request from supervisor
5. Completed epics (0-38) are not listed here — see ROADMAP.md Completed Epics table
6. When an epic completes, update its status here to "Complete" (do not remove it)

