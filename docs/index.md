# ThreeDoors Project Documentation Index

## Project Overview

- **Type:** Monolith CLI/TUI Application
- **Primary Language:** Go 1.25.4
- **Architecture:** Model-View-Update (MVU) via Bubbletea
- **Status:** Technical Demo & Validation (Epic 1) — planning complete, source code not yet on main

### Quick Reference

- **Tech Stack:** Go 1.25.4 + Bubbletea 1.2.4 + Lipgloss 1.0.0 + YAML storage
- **Entry Point:** `cmd/threedoors/main.go` (planned)
- **Architecture Pattern:** Two-layer monolith (TUI + Domain), MVU pattern
- **Data:** Local YAML files at `~/.threedoors/`

## Generated Documentation

- [Project Overview](./project-overview.md) — Project summary, status, roadmap
- [Architecture Summary](./architecture.md) — Consolidated architecture overview
- [Source Tree Analysis](./source-tree-analysis.md) — Directory structure and critical folders
- [Technology Stack](./technology-stack.md) — Full technology table
- [Development Guide](./development-guide.md) — Prerequisites, build, test, coding standards
- [Project Structure](./project-structure.md) — Classification and legacy artifact inventory
- [Comprehensive Analysis](./comprehensive-analysis.md) — Component and data model analysis
- [Existing Documentation Inventory](./existing-documentation-inventory.md) — Catalog of all docs found

## Existing Documentation (BMAD v4 Artifacts)

### Product & Requirements

- [Product Brief](./brief.md) — Executive product brief with vision and scope
- [PRD (index)](./prd/index.md) — Product Requirements Document (10 sharded files)
  - [Goals & Background](./prd/goals-and-background-context.md)
  - [Requirements](./prd/requirements.md)
  - [UI Design Goals](./prd/user-interface-design-goals.md)
  - [Technical Assumptions](./prd/technical-assumptions.md)
  - [Epic List](./prd/epic-list.md)
  - [Epic Details](./prd/epic-details.md)
  - [Checklist Results](./prd/checklist-results-report.md)
  - [Next Steps](./prd/next-steps.md)
  - [Appendix: Story Optimization](./prd/appendix-story-optimization-summary.md)

### Architecture (Detailed)

- [Architecture Index](./architecture/index.md) — Full sharded architecture (19 files)
  - [Introduction](./architecture/introduction.md)
  - [High-Level Architecture](./architecture/high-level-architecture.md)
  - [Tech Stack](./architecture/tech-stack.md)
  - [Components](./architecture/components.md)
  - [Core Workflows](./architecture/core-workflows.md)
  - [Data Models](./architecture/data-models.md)
  - [Data Storage Schema](./architecture/data-storage-schema.md)
  - [Source Tree](./architecture/source-tree.md)
  - [Coding Standards](./architecture/coding-standards.md)
  - [Test Strategy](./architecture/test-strategy-and-standards.md)
  - [Error Handling](./architecture/error-handling-strategy.md)
  - [Security](./architecture/security.md)
  - [External APIs](./architecture/external-apis.md)
  - [REST API Spec](./architecture/rest-api-spec.md)
  - [Infrastructure & Deployment](./architecture/infrastructure-and-deployment.md)
  - [Checklist Results](./architecture/checklist-results-report.md)
  - [Next Steps](./architecture/next-steps.md)

### Stories & Implementation

- [Story 1.1: Project Setup & Basic Bubbletea App](./stories/1.1.story.md) — Completed
- [Story 1.2: Display Three Doors from Task File](./stories/1.2.story.md) — Completed
- [QA Gate 1.1](./qa/gates/1.1-project-setup-basic-bubbletea-app.yml) — Story 1.1 QA gate

### Process & Tracking

- [BMM Workflow Status](./bmm-workflow-status.yaml) — BMAD v4 workflow tracking
- [Brainstorming Results](./brainstorming-session-results.md) — Initial brainstorming session
- [Deliverables Summary](./DELIVERABLES-SUMMARY.md) — Summary of all deliverables
- [Changelog](./CHANGELOG-2025-11-07-to-11.md) — Nov 7-11 2025 changes
- [Validation Decision Rubric](./validation-decision-rubric.md) — Decision rubric

### Archive

- [Archive README](./.archive/README.md) — Archive description
- [Monolithic PRD (Nov 7, 2025)](./.archive/prd-monolithic-2025-11-07.md) — Pre-sharding version
- [Monolithic Architecture (Nov 7, 2025)](./.archive/architecture-monolithic-2025-11-07.md) — Pre-sharding version

## Getting Started

1. Ensure Go 1.25.4+ is installed (`go version`)
2. Clone: `git clone https://github.com/arcaven/ThreeDoors.git`
3. Build: `make build` (once source code is committed)
4. Run: `make run` or `./bin/threedoors`
5. Read the [Development Guide](./development-guide.md) for coding standards and testing

## For AI-Assisted Development

When creating a brownfield PRD or planning new features, point your workflow to this `index.md` as the primary context source. Key files for AI context:

1. **This index** — Navigation map
2. **[Architecture Summary](./architecture.md)** — Quick architectural understanding
3. **[PRD](./prd/index.md)** — Full requirements and epic details
4. **[Development Guide](./development-guide.md)** — Coding standards and patterns
5. **[Components](./architecture/components.md)** — Detailed component specifications
