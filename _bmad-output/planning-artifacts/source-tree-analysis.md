# ThreeDoors Source Tree Analysis

## Current Repository Structure

```
ThreeDoors/
├── README.md                          # Project overview, features, key bindings, getting started
├── AGENTS.md                          # AI agent configuration (204KB)
├── .tool-versions                     # Go 1.25.4 (asdf version manager)
├── .gitignore                         # Comprehensive ignore rules
│
├── docs/                              # 📚 Project documentation (primary knowledge base)
│   ├── brief.md                       # Product brief — project vision and scope
│   ├── brainstorming-session-results.md # Initial brainstorming output
│   ├── bmm-workflow-status.yaml       # BMAD v4 workflow tracking
│   ├── DELIVERABLES-SUMMARY.md        # Summary of all deliverables
│   ├── CHANGELOG-2025-11-07-to-11.md  # Development changelog
│   ├── validation-decision-rubric.md  # Decision rubric for validation
│   │
│   ├── prd/                           # 📋 Product Requirements Document (sharded, 10 files)
│   │   ├── index.md                   # PRD table of contents
│   │   ├── goals-and-background-context.md
│   │   ├── requirements.md            # Technical demo + full MVP requirements
│   │   ├── user-interface-design-goals.md
│   │   ├── technical-assumptions.md
│   │   ├── epic-list.md               # Epic listing by phase
│   │   ├── epic-details.md            # Detailed epic/story breakdowns
│   │   ├── checklist-results-report.md
│   │   ├── next-steps.md
│   │   └── appendix-story-optimization-summary.md
│   │
│   ├── architecture/                  # 🏗️ Architecture Documentation (sharded, 19 files)
│   │   ├── index.md                   # Architecture table of contents
│   │   ├── introduction.md
│   │   ├── table-of-contents.md
│   │   ├── high-level-architecture.md # MVU pattern, layer diagram
│   │   ├── tech-stack.md              # Full technology table
│   │   ├── components.md              # TUI + Domain layer components
│   │   ├── core-workflows.md          # 6 mermaid sequence diagrams
│   │   ├── data-models.md             # Task, TaskPool, DoorSelection models
│   │   ├── data-storage-schema.md     # YAML file format specs
│   │   ├── source-tree.md             # Planned source tree
│   │   ├── coding-standards.md        # Go coding conventions
│   │   ├── test-strategy-and-standards.md
│   │   ├── error-handling-strategy.md
│   │   ├── security.md
│   │   ├── external-apis.md           # N/A (local-only)
│   │   ├── rest-api-spec.md           # N/A (no REST API)
│   │   ├── infrastructure-and-deployment.md
│   │   ├── checklist-results-report.md
│   │   └── next-steps.md
│   │
│   ├── stories/                       # 📖 User Stories
│   │   ├── 1.1.story.md               # Project Setup & Basic Bubbletea App ✅
│   │   └── 1.2.story.md               # Display Three Doors from Task File ✅
│   │
│   ├── qa/                            # 🧪 QA Gates
│   │   └── gates/
│   │       └── 1.1-project-setup-basic-bubbletea-app.yml
│   │
│   └── .archive/                      # 📦 Archived Documentation
│       ├── README.md
│       ├── prd-monolithic-2025-11-07.md
│       └── architecture-monolithic-2025-11-07.md
│
├── _bmad/                             # 🔧 BMAD v6 Installation (current)
├── _bmad-output/                      # 📤 BMAD v6 Output (empty)
│   ├── planning-artifacts/
│   ├── implementation-artifacts/
│   └── test-artifacts/
│
├── .ai/                               # 🤖 Legacy AI workflow state
│   └── workflow-state.yaml            # v4 greenfield-service workflow tracker
│
├── .bmad-core/                        # 📦 Legacy BMAD v4 Installation
│   ├── agents/                        # v4 agent definitions
│   ├── agent-teams/                   # v4 agent team configs
│   ├── checklists/                    # v4 validation checklists
│   ├── data/                          # v4 data files
│   ├── tasks/                         # v4 task definitions
│   ├── templates/                     # v4 document templates
│   ├── workflows/                     # v4 workflow definitions
│   ├── utils/                         # v4 utilities
│   ├── core-config.yaml               # v4 core configuration
│   ├── install-manifest.yaml          # v4 installation manifest
│   ├── enhanced-ide-development-workflow.md
│   ├── user-guide.md
│   └── working-in-the-brownfield.md
│
├── .windsurf/workflows/               # Windsurf IDE agent workflows
├── .gemini/commands/                   # Gemini IDE commands
├── .claude/                            # Claude Code configuration
└── .ignore/                            # Additional ignore rules

## Planned Source Tree (Not Yet Implemented)

Per the architecture documentation, the following source structure is planned:

```
cmd/threedoors/                        # 🚀 Application entry point
├── main.go                            # Program entry, Bubbletea program init
└── main_test.go                       # Entry point tests

internal/
├── tui/                               # 🖥️ TUI Layer (Bubbletea components)
│   ├── model.go                       # MainModel — root view router
│   ├── doors_view.go                  # DoorsView — three doors display
│   ├── detail_view.go                 # TaskDetailView — task details
│   ├── status_menu.go                 # StatusUpdateMenu — status selector
│   └── notes_input.go                 # NotesInputView — text input
│
└── tasks/                             # 📦 Domain Layer
    ├── task.go                        # Task model + validation
    ├── task_pool.go                   # TaskPool — in-memory collection
    ├── file_manager.go                # FileManager — YAML I/O
    ├── status_manager.go              # StatusManager — state machine
    ├── door_selector.go               # DoorSelector — selection algorithm
    └── session_tracker.go             # Session metrics tracking
```

## Critical Folders

| Folder | Purpose | Status |
|---|---|---|
| `docs/prd/` | Product requirements — drives all development | Complete |
| `docs/architecture/` | Technical design — component specs, data models, workflows | Complete |
| `docs/stories/` | Implementation stories with acceptance criteria | 2 of 6 complete |
| `cmd/threedoors/` | Application entry point | Not yet created |
| `internal/tui/` | TUI components (Bubbletea) | Not yet created |
| `internal/tasks/` | Domain logic and persistence | Not yet created |
