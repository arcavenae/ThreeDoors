# Requirements

## Technical Demo & Validation Phase Requirements

**Core Requirements (Week 1 - Build & Validate):**

**TD1:** The system shall provide a CLI/TUI interface optimized for terminal emulators (iTerm2 and similar)

**TD2:** The system shall read tasks from a simple local text file (e.g., `~/.threedoors/tasks.txt`)

**TD3:** The system shall display the Three Doors interface showing three tasks selected from the text file

**TD4:** The system shall support door selection via 'a' or 'left arrow' (left door), 'w' or 'up arrow' (center door), and 'd' or 'right arrow' (right door). Initially, or after re-rolling, no door shall be selected.

**TD5:** The system shall provide a refresh mechanism using 's' or 'down arrow' to generate a new set of three doors.

**TD6:** The system shall display the three doors with dynamic width adjustment based on the terminal size.

**TD7:** The system shall respond to the following keystrokes for task management, with functionality to be implemented in future stories:
    *   'c': Mark selected task as complete.
    *   'b': Mark selected task as blocked.
    *   'i': Mark selected task as in progress.
    *   'e': Expand selected task (into more tasks).
    *   'f': Fork selected task (clone/split).
    *   'p': Procrastinate/avoid selected task.

**TD8:** The system shall display at least one "progress over perfection" encouragement message per session (e.g., on startup greeting or after completing a task)

**TD9:** The system shall write completed tasks to a separate file (e.g., `~/.threedoors/completed.txt`) with timestamp

**Success Criteria for Phase:**
- Built and running within 4-8 hours of development time
- Developer uses it daily for 1 week to validate UX concept
- Three Doors selection results in faster time-to-first-action compared to scanning a full task list (measured via session metrics)
- Decisions made on whether to proceed to Full MVP based on real usage

---

## Full MVP Requirements (Post-Validation - Deferred)

**Phase 2 - Apple Notes Integration:**

**FR2:** The system shall integrate with Apple Notes as the primary task storage backend, enabling bidirectional sync

**FR4:** The system shall retrieve and display tasks from Apple Notes within the application interface

**FR5:** The system shall mark tasks as complete upon user action, updating both the application state and Apple Notes

**FR12:** The system shall allow updates to tasks from either the application or directly in Apple Notes on iPhone, with changes reflected bidirectionally

**FR15:** The system shall provide a health check command to verify Apple Notes connectivity and database integrity

**Phase 3 - Enhanced Interaction & Learning:**

**FR3:** The system shall support new task capture with optional context (what and why) through the CLI/TUI

**FR6:** The system shall display user-defined values and goals persistently throughout task work sessions

**FR7:** The system shall provide a "choose-your-own-adventure" interactive navigation flow that presents options rather than demands

**FR8:** The system shall track daily task completion count and display comparison to previous day's count

**FR9:** The system shall prompt the user once per session with: "What's one thing you could improve about this list/task/goal right now?"

**FR10:** The system shall embed "progress over perfection" messaging throughout interaction patterns and interface copy (enhanced beyond Tech Demo)

**FR16:** The system shall support a "quick add" mode for capturing tasks in 3 or fewer keystrokes beyond typing the task text (e.g., `:add <text>` + Enter)

**FR18:** The system shall present door feedback options (Blocked, Not now, Needs breakdown, Other comment) when a door is dismissed, capturing the reason for future selection tuning

**FR19:** The system shall capture and store blocker information when a task is marked as blocked

**FR20:** The system shall use door selection and feedback patterns to inform future door selection (learning which task types suit which contexts)

**FR21:** The system shall categorize tasks by type, effort level, and context to enable diverse door selection

**Phase 4 - Distribution & Packaging (macOS):**

**FR22:** The system shall provide macOS binaries that are code-signed with a valid Apple Developer certificate so that Gatekeeper does not quarantine the binary on download

**FR23:** The system shall be notarized with Apple's notarization service so that Gatekeeper allows execution without requiring users to bypass security warnings

**FR24:** The system shall be installable via Homebrew using a custom tap (`brew install arcaven/tap/threedoors`), with the formula downloading the appropriate signed binary for the user's architecture

**FR25:** The system shall provide a DMG or pkg installer as an alternative installation method for users who prefer graphical installation over Homebrew

**FR26:** The release process shall automate code signing, notarization, distribution formula updates, and installer generation without manual intervention

**Phase 5 - Data Layer & Enrichment:**

**FR11:** The system shall maintain a local enrichment layer for metadata, cross-references, and relationships that cannot be stored in source systems

---

## Phase 6+ - Party Mode Recommendations (Accepted)

*The following requirements were accepted through party mode consensus review and extend the product roadmap.*

**Obsidian Integration (P0 - #2 Integration After Apple Notes):**

**FR27:** The system shall integrate with Obsidian vaults as a task storage backend, reading and writing Markdown files in a configurable vault directory

**FR28:** The system shall support bidirectional sync with Obsidian, detecting external changes to vault files and reflecting them in the application

**FR29:** The system shall support Obsidian vault configuration via `~/.threedoors/config.yaml`, including vault path, target folder, and file naming conventions

**FR30:** The system shall integrate with Obsidian daily notes, optionally reading/writing tasks from/to daily note files

**Plugin/Adapter SDK:**

**FR31:** The system shall provide an adapter registry that discovers and loads task provider adapters at runtime

**FR32:** The system shall support config-driven provider selection via `~/.threedoors/config.yaml`, allowing users to specify active backends and their configuration

**FR33:** The system shall provide an adapter developer guide and interface specification enabling third-party integration development

**Psychology Research & Validation:**

**FR34:** The system shall document the evidence base for the Three Doors choice architecture (why three options), mood-task correlation models, procrastination intervention mechanisms, and the "progress over perfection" motivational framework

**LLM Task Decomposition & Agent Action Queue:**

**FR35:** The system shall support LLM-powered task decomposition, where a user-selected task is broken into implementable stories/specs by a language model

**FR36:** The system shall output LLM-generated stories/specs to a git repository structure compatible with coding agents (Claude Code, multiclaude)

**FR37:** The system shall support configurable LLM backends (local and cloud) for task decomposition

**First-Run Onboarding Experience:**

**FR38:** The system shall provide a first-run welcome flow that guides users through setting up values/goals, explains the Three Doors concept, and walks through key bindings

**FR39:** The system shall support optional import from existing task sources (text files, other tools) during onboarding

**Sync Observability & Offline-First:**

**FR40:** The system shall operate offline-first, queuing changes locally when sync targets are unavailable and replaying them when connectivity is restored

**FR41:** The system shall display a sync status indicator in the TUI showing current sync state per provider

**FR42:** The system shall provide conflict visualization when sync conflicts are detected between local and remote state

**FR43:** The system shall maintain a sync log for debugging synchronization issues

**Calendar Awareness (Local-First, No OAuth):**

**FR44:** The system shall read local calendar sources only (no OAuth or cloud API dependencies) to determine upcoming events and available time blocks

**FR45:** The system shall use calendar context to inform time-contextual door selection (e.g., suggesting tasks appropriate for available time blocks)

**Multi-Source Task Aggregation:**

**FR46:** The system shall aggregate tasks from all configured providers into a unified cross-provider task pool

**FR47:** The system shall detect and flag potential duplicate tasks across providers

**FR48:** The system shall display source attribution in the TUI, indicating which provider each task originates from

**Testing Strategy:**

**FR49:** The system shall include integration tests for Apple Notes E2E workflows

**FR50:** The system shall include contract tests validating adapter compliance with the TaskProvider interface

**FR51:** The system shall include functional E2E tests covering full user workflows

**Door Theme System:**

**FR55:** The system shall provide a theme registry containing named door themes, each defining a render function that produces a visually distinct ASCII/ANSI art frame around task content

**FR56:** The system shall ship with at least three door themes (Modern/Minimalist, Sci-Fi/Spaceship, Japanese Shoji) plus a Classic theme that preserves the current Lipgloss border rendering

**FR57:** The system shall display a horizontal theme preview during first-run onboarding, enabling theme browsing and selection before the first session begins

**FR58:** The system shall provide a theme selection view accessible via `:theme` command in the TUI, allowing users to change their active theme with immediate visual effect (no restart required)

**FR59:** The system shall persist the selected theme in `~/.threedoors/config.yaml` as a string key (e.g., `theme: modern`), falling back to the default theme (Modern/Minimalist) when the configured theme name is invalid

**FR60:** The system shall apply the same theme to all three doors in the trio (single theme selection, not per-door assignment)

**FR61:** The system shall gracefully fall back to Classic theme rendering when the terminal width is below a theme's declared minimum width

**FR62:** The system shall overlay door number labels separately from the theme frame, maintaining consistent door identification across all themes

---

## Non-Functional Requirements

**Technical Demo Phase:**

**TD-NFR1:** The system shall be built in Go 1.25.4+ using idiomatic patterns and gofumpt formatting standards

**TD-NFR2:** The system shall use the Bubbletea/Charm Bracelet ecosystem for TUI implementation

**TD-NFR3:** The system shall operate on macOS as the primary target platform

**TD-NFR4:** The system shall store all data in local text files (`~/.threedoors/` directory) with no external services or telemetry

**TD-NFR5:** The system shall respond to user interactions within the CLI/TUI with minimal latency (target: <100ms for typical operations given simple file I/O)

**TD-NFR6:** The system shall use Make as the build system with simple targets: `build`, `run`, `clean`

**TD-NFR7:** The system shall gracefully handle missing or corrupted task files by creating defaults

---

**Full MVP Phase (Post-Validation - Deferred):**

**NFR1:** The system shall maintain idiomatic Go patterns and gofumpt formatting standards

**NFR2:** The system shall continue using Bubbletea/Charm Bracelet ecosystem

**NFR3:** The system shall operate on macOS 13+ (Ventura and later) as primary platform, with binaries code-signed and notarized such that `spctl --assess --type execute` returns "accepted" and Gatekeeper permits first launch without user override

**NFR4:** The system shall store all user data locally or in user's iCloud (via Apple Notes), with no external telemetry or tracking

**NFR5:** The system shall store application state and enrichment data locally (cross-computer sync deferred to post-MVP)

**NFR6:** The system shall respond to user interactions within the CLI/TUI with minimal latency (target: <500ms for typical operations)

**NFR7:** The system shall fall back to local text file storage when Apple Notes integration is unavailable, maintaining door selection, task status changes, and session metrics without error — verified by running the full test suite with Apple Notes disconnected

**NFR8:** The system shall implement secure credential storage using OS keychain for any API keys or authentication tokens

**NFR9:** The system shall never write API keys, authentication tokens, or keychain data to log files, session metrics, or stdout — verified by `grep -ri` scan of all output files after an integration test run returning zero matches

**NFR10:** The system shall use Make as the build system

**NFR11:** The system shall enforce architectural separation such that `internal/core` has zero import dependencies on `internal/tui`, adapter packages, or enrichment storage — verified by `go vet` dependency analysis and CI import-cycle checks

**NFR12:** The system shall maintain data integrity even when Apple Notes is modified externally while app is running

**NFR13:** The system shall enforce <100ms response time for adapter operations (read/write/sync) as a performance benchmark, validated by automated performance tests

**NFR14:** The system shall operate offline-first, with all core functionality available without network connectivity; sync operations shall be queued and replayed when targets become available

**NFR15:** The system shall never require OAuth or cloud API credentials for calendar integration; only local calendar sources (AppleScript, .ics files, CalDAV cache) are permitted

**NFR16:** The system shall maintain CI coverage gates ensuring test coverage does not regress below established thresholds

**NFR17:** Door theme rendering shall use only Unicode characters from the box-drawing (`U+2500–U+257F`), block elements (`U+2580–U+259F`), and geometric shapes (`U+25A0–U+25FF`) ranges to ensure consistent rendering across modern terminal emulators (iTerm2, Terminal.app, Alacritty, kitty, Windows Terminal)

**NFR18:** Door theme render functions shall complete within 1ms for standard terminal widths (40-200 columns), as theme rendering is pure string manipulation with no I/O

**NFR19:** Each door theme shall include golden file tests verifying rendered output at multiple widths (minimum width, standard width, wide terminal) in both selected and unselected states

---

## Phase 7+ - Task Source Sync Integration (Accepted)

*The following requirements extend the product with external task source integrations and sync protocol improvements.*

**Jira Integration:**

**FR63:** The system shall integrate with Jira as a read-only task source, querying issues via configurable JQL and mapping them to the ThreeDoors task model

**FR64:** The system shall provide configurable status mapping from Jira `statusCategory` and `status.name` to ThreeDoors `TaskStatus` values, with `statusCategory` as the default fallback

**FR65:** The system shall support Jira Cloud authentication via API Token + Basic Auth, and Jira Server/DC authentication via Personal Access Token + Bearer, configurable in `~/.threedoors/config.yaml`

**FR66:** The system shall support bidirectional Jira sync by transitioning issues to "Done" via the Jira transitions API when tasks are marked complete in ThreeDoors, with offline queuing via WALProvider

**Apple Reminders Integration:**

**FR67:** The system shall integrate with Apple Reminders as a task source using JXA (JavaScript for Automation) via `osascript`, reading reminders with structured field mapping (title, notes, due date, priority, completion status)

**FR68:** The system shall support full CRUD operations on Apple Reminders: creating, updating, completing, and deleting reminders from within the ThreeDoors TUI

**FR69:** The system shall allow users to configure which Apple Reminders lists to include as task sources via `~/.threedoors/config.yaml`, defaulting to all lists

**Sync Protocol Hardening:**

**FR70:** The system shall provide a sync scheduler with per-provider independent sync loops, supporting hybrid push (Watch channel) and polling with adaptive intervals (backoff on failure, reset on success)

**FR71:** The system shall implement a per-provider circuit breaker (Closed → Open → Half-Open states) that prevents cascading failures and displays health state in the TUI sync status

**FR72:** The system shall maintain canonical ID mapping via `SourceRef` entries linking internal task UUIDs to provider-native IDs, enabling permanent cross-provider deduplication

---

## Code Quality & Submission Standards

These non-functional requirements establish code quality gates that all contributions must pass before submission. They are derived from recurring patterns identified across 31+ pull requests (see `docs/architecture/pr-submission-standards.md` for full rationale and evidence).

**NFR-CQ1:** All code must pass `gofumpt` formatting before submission — running `gofumpt -l .` must produce no output (Evidence: PRs #9, #10, #23, #24 required formatting fix-up commits)

**NFR-CQ2:** All code must pass `golangci-lint run ./...` with zero issues before submission, including `errcheck` and `staticcheck` analyzers (Evidence: PR #16 required 2 extra commits to resolve lint findings)

**NFR-CQ3:** All branches must be rebased onto `upstream/main` before PR creation — no merge commits, no stale branches (Evidence: PRs #3, #5, #19, #23 were delayed by merge conflicts from stale branches)

**NFR-CQ4:** All PRs must have a clean `git diff --stat` showing only in-scope changes — no unrelated files, no accidental directory additions (Evidence: PR #5 included an unrelated `agents/` directory)

**NFR-CQ5:** All fix-up commits must be squashed before PR submission — PRs should contain a single clean commit or logically separated commits, not iterative fix-up trails (Evidence: ~15 PRs contained avoidable "fix: address code review findings" commits)

---

## Systemic NFRs Derived from PR Analysis (PRs #1–#49)

> Analysis of all 49 PRs found 18 (37%) required post-submission changes. These NFRs prevent recurring defect classes. For detailed code examples and patterns, see `docs/architecture/coding-standards.md` Rules 9–13.

| NFR ID | Requirement | Coding Standard | Evidence |
|--------|------------|-----------------|----------|
| **NFR-SB1** | Use `fmt.Fprintf()` not `WriteString(Sprintf())` for all string building | Rule 9 | PRs #42, #44, #45 (11+ instances, 5 fix-ups) |
| **NFR-SB2** | Sweep entire codebase when fixing a lint category, not just reported lines | Rule 13 | PR #42 (3 incremental fix commits) |
| **NFR-EH1** | Check ALL error return values including `f.Close()`, `os.Remove()`, `os.WriteFile()` | Rule 10 | PRs #16, #42, #43 (18+ violations) |
| **NFR-EH2** | Makefile targets must not silently swallow errors | Rule 10 | PR #16 |
| **NFR-EH3** | Configuration/setup errors must be handled or explicitly documented as ignored | Rule 10 | PR #17 |
| **NFR-IS1** | Escape all user strings interpolated into AppleScript/shell/interpreted languages | Rule 11 | PR #17 (injection vulnerability) |
| **NFR-IS2** | Include test cases with special characters for dynamic command construction | Rule 11 | PR #17 |
| **NFR-TQ1** | Deleting test cases requires equivalent replacement coverage in the same PR | — | PRs #5, #7 (324 deleted lines, retroactive fix) |
| **NFR-TQ2** | Test assertions must verify actual outcomes, not just absence of errors | — | PR #20 |
| **NFR-TQ3** | Collections must be tested for ordering; non-ordered results must be sorted | — | PR #14 (non-deterministic search) |
| **NFR-TR1** | `time.Now()` called once per operation, reused — never inside loops | Rule 12 | PR #17 |
| **NFR-TR2** | Random selection must include anti-repeat guard | — | PR #18 |
| **NFR-BH1** | Re-run `gofumpt` after every rebase (rebase can introduce formatting drift) | — | PR #23 |
| **NFR-BH2** | Implement stories in dependency order to avoid merge conflicts | — | PRs #3, #5 |
| **NFR-BH3** | Coordinate parallel agent story assignments to prevent duplicate work | — | PRs #14/#13, #49/#45 (1,157+ lines wasted) |

---

## Task Source Integration NFRs

> Requirements specific to API-based and IPC-based task source adapters.

**NFR20:** API-based adapters must handle HTTP 429 (Too Many Requests) responses by respecting the `Retry-After` header and applying exponential backoff with jitter before retrying

**NFR21:** API-based adapters must cache loaded tasks locally with a configurable TTL to avoid hitting rate limits on every TUI refresh; stale cache must be flagged in the UI

**NFR22:** Credential storage for external adapters must use environment variables or `~/.threedoors/config.yaml` settings — credentials must never be stored in task YAML files or sync state

**NFR23:** The per-provider circuit breaker must trip to Open state after 5 consecutive failures within a 2-minute window; probe requests must occur at configurable intervals (default 30 seconds, max 30 minutes with exponential backoff)

---
