# ThreeDoors Project Structure

## Classification

| Attribute | Value |
|---|---|
| **Project Name** | ThreeDoors |
| **Repository Type** | Monolith |
| **Project Type ID** | cli |
| **Language** | Go 1.25.4 |
| **TUI Framework** | Bubbletea (Charm ecosystem) + Lipgloss |
| **Architecture Pattern** | Model-View-Update (MVU) |
| **Build System** | Make |
| **Data Storage** | Local flat files (~/.threedoors/) |
| **Current Phase** | Technical Demo & Validation (Epic 1) |

## Project Parts

| Part ID | Type | Root Path | Description |
|---|---|---|---|
| threedoors | cli | / | Single monolith TUI application |

## Legacy BMAD v4 Artifacts

| Location | Contents |
|---|---|
| `.ai/workflow-state.yaml` | v4 workflow tracker (greenfield-service, analyst stage) |
| `.bmad-core/` | Full v4 BMAD installation (agents, tasks, templates, workflows) |
| `docs/bmm-workflow-status.yaml` | v4 BMM workflow status tracking |
| `docs/brief.md` | Product brief (completed) |
| `docs/prd/` | Sharded PRD — 10 files (completed, validated) |
| `docs/architecture/` | Sharded architecture — 19 files (completed) |
| `docs/stories/` | Stories 1.1 and 1.2 (completed) |
| `docs/qa/gates/` | QA gate for story 1.1 |
| `docs/.archive/` | Monolithic PRD and architecture from Nov 7, 2025 |
| `docs/brainstorming-session-results.md` | Initial brainstorming output |
| `docs/CHANGELOG-2025-11-07-to-11.md` | Changelog |
| `docs/DELIVERABLES-SUMMARY.md` | Deliverables summary |
| `docs/validation-decision-rubric.md` | Validation decision rubric |

## Source Code Status

The README describes an intended project structure:
- `cmd/threedoors/` — Application entry point
- `internal/tasks/` — Task domain logic

No Go source files were detected at the project root. Code may not have been committed yet or may exist in a different branch.
