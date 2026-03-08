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

## Phase 3+ - Party Mode Recommendations (Accepted)

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

**Docker E2E & Headless TUI Testing:**

**FR52:** The system shall provide a headless TUI test harness using Bubbletea's `teatest` package for automated interaction testing with programmatic key input and model assertions

**FR53:** The system shall include golden file snapshot tests for TUI visual regression detection, comparing rendered output against stored reference files

**FR54:** The system shall provide a Docker-based reproducible test environment (`Dockerfile.test` + `docker-compose.test.yml`) for E2E test execution in CI

**Door Theme System:**

**FR55:** The system shall provide a theme registry containing named door themes, each defining a render function that produces a visually distinct ASCII/ANSI art frame around task content

**FR56:** The system shall ship with at least three door themes (Modern/Minimalist, Sci-Fi/Spaceship, Japanese Shoji) plus a Classic theme that preserves the current Lipgloss border rendering

**FR57:** The system shall display a horizontal theme preview during first-run onboarding, enabling theme browsing and selection before the first session begins

**FR58:** The system shall provide a theme selection view accessible via `:theme` command in the TUI, allowing users to change their active theme with immediate visual effect (no restart required)

**FR59:** The system shall persist the selected theme in `~/.threedoors/config.yaml` as a string key (e.g., `theme: modern`), falling back to the default theme (Modern/Minimalist) when the configured theme name is invalid

**FR60:** The system shall apply the same theme to all three doors in the trio (single theme selection, not per-door assignment)

**FR61:** The system shall gracefully fall back to Classic theme rendering when the terminal width is below a theme's declared minimum width

**FR62:** The system shall overlay door number labels separately from the theme frame, maintaining consistent door identification across all themes

**Seasonal Door Theme Variants:**

**FR132:** The system shall provide seasonal door theme variants — self-contained `DoorTheme` implementations with time-appropriate visual patterns (e.g., crystalline patterns for winter, flowing lines for spring, radiating shapes for summer, angular textures for autumn) — using only Unicode characters from the box-drawing (`U+2500–U+257F`), block elements (`U+2580–U+259F`), and geometric shapes (`U+25A0–U+25FF`) ranges per NFR17

**FR133:** The system shall support automatic seasonal theme switching based on the current date, with configurable season date ranges stored in `~/.threedoors/config.yaml` — defaulting to meteorological seasons (spring: March 1, summer: June 1, autumn: September 1, winter: December 1) — checked on application startup and on each planning session start

**FR134:** The system shall allow users to disable automatic seasonal switching via `seasonal_themes: false` in `~/.threedoors/config.yaml` (default: `true`), reverting to the user's manually selected base theme when disabled

**FR135:** The system shall provide a `:seasonal` command in the TUI that displays all seasonal theme variants in a horizontal preview grid (consistent with the `:theme` command from FR58), allowing manual season override for testing or preference

**FR136:** All seasonal theme variants shall maintain WCAG AA contrast ratios (minimum 4.5:1 for text content) in both light and dark terminal color schemes, validated by automated contrast-ratio checks in theme test suites

**FR137:** Seasonal themes shall fall back to the user's configured base theme when the terminal width is below the seasonal variant's declared minimum width, consistent with FR61 fallback behavior

**Door Visual Appearance — Door-Like Proportions:**

**FR138:** The door rendering system shall use portrait-oriented aspect ratios (taller than wide) for all door themes, with a minimum door height of 12 rows, to achieve visual recognition as doors rather than cards or panels

**FR139:** All door themes shall include a panel divider — at least one horizontal line within the door frame creating distinct upper and lower panels — as this is the strongest "door" signifier after proportion

**FR140:** All door themes shall render an asymmetric handle/knob on the right side of the door at approximately 60% of the door height, using theme-appropriate handle characters (e.g., `●` for Classic, `○` for Modern, `◈──┤` for Sci-Fi, `○` recessed for Shoji)

**FR141:** All door themes shall render a threshold or floor line at the bottom edge of the door frame, using theme-appropriate treatment (e.g., shadow characters `▔`, floor grating `▓`, or distinct bottom border) to suggest a ground plane

**FR142:** Door numbers shall be rendered in the lintel/header area of each door frame (top border), styled as room numbers rather than content labels, consistent across all themes per FR62

**FR143:** The `DoorTheme.Render()` function signature shall accept a height parameter in addition to width: `Render(content string, width int, height int, selected bool) string` — enabling height-aware door rendering

**FR144:** A `DoorAnatomy` helper type shall calculate structural row positions from door height, including lintel row, content start row, panel divider row (~45% height), handle row (~60% height), and threshold row

**FR145:** The `DoorTheme` struct shall include a `MinHeight` field alongside the existing `MinWidth`, declaring the minimum terminal height required for door-like rendering

**FR146:** When terminal height is below a theme's `MinHeight`, the system shall fall back to compact rendering mode (current landscape card layout), providing graceful degradation rather than broken proportions

**FR147:** All door appearance signifiers (proportion, panel divider, handle, threshold) shall work in monochrome mode using structural elements rather than color — ensuring accessibility for users with color vision deficiencies or monochrome terminals

**Door Selection Interaction Feedback (Issue #219):**

**FR148:** When a door is selected, the system shall render the selected door with strong visual emphasis (bold text, bright foreground, enhanced border) and render unselected doors with diminished styling (faint/dimmed text, subdued borders), creating an unmistakable focus funnel effect

**FR149:** The visual contrast between selected and unselected door states shall be distinguishable in monochrome mode using structural emphasis (bold, faint) rather than relying solely on color differentiation

**FR150:** Pressing the same door selection key that selected the currently active door shall toggle the selection off, returning to the neutral state (no door selected) — enabling reversible exploration without re-rolling

**FR151:** The 'q' key shall function as a universal quit command from all non-text-input views, providing a consistent escape route regardless of current screen

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

**TD-NFR8:** The system shall never panic due to nil provider initialization — all code paths that obtain a `TaskProvider` from factory functions must check for nil and return a descriptive error before use (ref: Issue #218, Story 23.11)

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

**NFR28:** The seasonal theme date-range resolver shall be a pure function with no I/O dependencies, completing date-to-season resolution in under 1 microsecond as measured by Go benchmark tests

**NFR29:** Each seasonal theme variant shall include golden file tests at three widths (minimum, 80-column standard, 120-column wide) in both selected and unselected states — totaling 24 golden files minimum (4 seasons x 3 widths x 2 states)

**NFR30:** Seasonal theme contrast ratios shall be validated programmatically in test suites by extracting Lipgloss color values and computing WCAG luminance ratios, not by manual visual inspection alone

---

## Phase 4+ - Task Source Sync Integration (Accepted)

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

**Todoist Integration:**

**FR89:** The system shall integrate with Todoist as a task source using the REST API v1, reading tasks with structured field mapping (content to Text, description to Context, labels to Tags, priority to Effort with scale inversion), filtering out deleted tasks (is_deleted == true)

**FR90:** The system shall support Todoist authentication via personal API token configured in `~/.threedoors/config.yaml`, with optional project ID filtering (`project_ids`) or Todoist filter expressions (`filter`) for scoping which tasks to import — these two options are mutually exclusive

**FR91:** The system shall map Todoist priority values (1=normal, 2=high, 3=urgent, 4=critical) to ThreeDoors Effort values with appropriate scale inversion (Todoist 4 maps to highest effort), with priority 0 (no priority) mapping to lowest effort

**FR92:** The system shall support bidirectional Todoist sync by completing tasks via the REST API when tasks are marked complete in ThreeDoors, with offline queuing via WALProvider

**GitHub Issues Integration:**

**FR93:** The system shall integrate with GitHub Issues as a task source using the official go-github SDK, reading issues with structured field mapping (title to Text, body to Context, labels to Tags, state to Status), filtering by assignee and configurable repository scope

**FR94:** The system shall support GitHub authentication via Personal Access Token (PAT) configured in `~/.threedoors/config.yaml` or `GITHUB_TOKEN` environment variable, with a configurable repository list (`repos`) and assignee filter (default: `@me`) for scoping which issues to import

**FR95:** The system shall map GitHub Issues fields to ThreeDoors task model: `open` state maps to `todo`, `closed` state maps to `complete`, labels matching `priority:*` convention map to Effort values, `milestone.due_on` maps to due date, and `in-progress` label maps to `in-progress` status

**FR96:** The system shall support bidirectional GitHub sync by closing issues via the GitHub API when tasks are marked complete in ThreeDoors, with offline queuing via WALProvider

**Linear Integration:**

**FR116:** The system shall integrate with Linear as a task source using the Linear GraphQL API, reading issues with structured field mapping (title to Text, description to Context, labels to Tags, state.type to Status with full workflow state mapping, priority to Effort with scale inversion, estimate to Effort as secondary signal, dueDate to due date), filtered by team and assignee

**FR117:** The system shall support Linear authentication via personal API key configured in `~/.threedoors/config.yaml` or `LINEAR_API_KEY` environment variable, with a configurable team ID list (`team_ids`) for scoping which issues to import — supporting multiple teams

**FR118:** The system shall map Linear workflow states to ThreeDoors statuses: `triage`/`backlog`/`unstarted` map to `todo`, `started` maps to `in-progress`, `completed` maps to `complete`, `cancelled` maps to `archived`; and map Linear priority values (0=no priority, 1=urgent, 2=high, 3=medium, 4=low) to ThreeDoors Effort with appropriate inversion, with `estimate` (story points) as a secondary effort signal when priority is absent or zero

**FR119:** The system shall support bidirectional Linear sync by transitioning issues to the team's "Done" workflow state via the Linear GraphQL API when tasks are marked complete in ThreeDoors, with offline queuing via WALProvider

**Sync Protocol Hardening:**

**FR70:** The system shall provide a sync scheduler with per-provider independent sync loops, supporting hybrid push (Watch channel) and polling with adaptive intervals (backoff on failure, reset on success)

**FR71:** The system shall implement a per-provider circuit breaker (Closed → Open → Half-Open states) that prevents cascading failures and displays health state in the TUI sync status

**FR72:** The system shall maintain canonical ID mapping via `SourceRef` entries linking internal task UUIDs to provider-native IDs, enabling permanent cross-provider deduplication

---

## Phase 5+ - Self-Driving Development Pipeline (Accepted)

*The following requirements enable ThreeDoors to dispatch its own development tasks to multiclaude worker agents, creating a closed loop where the app manages its own development.*

**Self-Driving Dev Dispatch:**

**FR73:** The system shall support task dispatch to multiclaude worker agents via a `DevDispatch` metadata struct on the `Task` type, maintaining dispatch state (queued, dispatched, completed, failed) orthogonal to the task lifecycle status

**FR74:** The system shall provide a file-based dev queue at `~/.threedoors/dev-queue.yaml` that persists dispatch items with task reference, acceptance criteria, scope constraints, worker name, PR number, and status — using the existing atomic write pattern for safe persistence

**FR75:** The system shall provide a dispatch engine (`internal/dispatch/`) that wraps the `multiclaude` CLI, supporting `CreateWorker`, `ListWorkers`, `GetHistory`, and `RemoveWorker` operations via a `Dispatcher` interface for testability

**FR76:** The system shall support task dispatch from the TUI via an 'x' key binding in the detail view and a `:dispatch` command in the command palette, with a confirmation dialog before enqueueing

**FR77:** The system shall provide a dev queue view accessible via `:devqueue` command, displaying pending/dispatched/completed queue items with approve ('y'), reject ('n'), and kill ('K') actions

**FR78:** The system shall poll for worker status updates via `tea.Tick` at 30-second intervals using `multiclaude repo history`, automatically updating queue items and task `DevDispatch` fields with PR number, URL, and status

**FR79:** The system shall auto-generate follow-up tasks when workers produce results: "Review PR #N" on PR creation, "Fix CI on PR #N" on CI failure, and "Address review comments on PR #N" when review comments are posted

**FR80:** The system shall enforce safety guardrails: maximum concurrent workers (default 2), manual approval gate by default (`auto_dispatch: false`), minimum 5-minute cooldown between dispatches, daily dispatch limit (default 10), and JSONL audit log at `~/.threedoors/dev-dispatch.log`

**Self-Driving Dev Dispatch NFRs:**

**NFR24:** The dispatch engine shall validate that `multiclaude` is available on PATH during initialization; if not found, dispatch-related key bindings and commands shall be hidden/disabled with a user-facing message

**NFR25:** The dispatch feature shall be gated behind `dev_dispatch_enabled: true` in `~/.threedoors/config.yaml`, disabled by default, to prevent accidental exposure for users who do not use multiclaude

**NFR26:** All dispatched worker task descriptions shall include fork workflow instructions, commit signing requirements, and scope constraints — ensuring workers follow the project's contribution standards

**NFR27:** The dev queue file shall survive TUI process restarts; on TUI launch, the dispatch engine shall read the queue and resume polling for any items in `dispatched` status

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

## Phase 6+ - Daily Planning Mode (Accepted)

*The following requirements add a guided daily planning ritual to ThreeDoors, creating a proactive morning engagement loop that drives long-term retention.*

**Daily Planning Mode:**

**FR97:** The system shall provide a daily planning mode accessible via `threedoors plan` CLI command or `:plan` TUI command that guides users through a structured morning planning ritual with three sequential steps: review, select, and confirm

**FR98:** The planning mode shall present yesterday's incomplete tasks with options to continue (leave in pool with focus priority), skip (leave in pool without priority), or snooze (open SnoozeView to set a defer-until date via the same mechanism as the Z-key snooze action) each task — this is a quick triage step that integrates with the first-class snooze/defer system

**FR99:** The planning mode shall allow users to select up to 5 tasks as "today's focus" from the full task pool, defaulting to 3 focus tasks (consistent with the Three Doors brand), using the session-scoped `+focus` tag convention so focus state resets on the next planning session

**FR100:** The planning mode shall infer current energy level from time of day as a default (morning = high, afternoon = medium, evening = low) and display it with an option to override, using the energy level to filter and sort focus task suggestions by matching effort level

**FR101:** The planning session shall display a soft progress indicator (step counter and elapsed time) without enforcing a hard time limit — showing a gentle nudge message after 10 minutes but never forcibly ending the session

**FR102:** Today's focus tasks (tagged `+focus`) shall receive an elevated scoring boost in door selection, appearing more frequently as doors until completed or until the focus session expires (planning session timestamp + 16 hours, or next planning session, whichever comes first)

**FR103:** The system shall track planning session completion, duration, number of tasks reviewed, and number of focus tasks selected as a `planning_session` event type in the JSONL session metrics log

---

## Phase 6+ - Snooze/Defer as First-Class Action (Accepted)

*The following requirements surface the existing deferred status as a first-class user action with date-based snooze, making the door pool trustworthy by ensuring only actionable tasks appear.*

**Snooze/Defer:**

**FR104:** The system shall provide a snooze action accessible via the `Z` key when a door is selected in the doors view or when viewing a task in the detail view, presenting quick defer options: Tomorrow (next day 9:00 AM local), Next Week (next Monday 9:00 AM local), Pick Date (text input for custom date), and Someday (indefinite deferral with no return date)

**FR105:** When a task is snoozed, the system shall set a `defer_until` timestamp on the task (nil for Someday), transition its status to `deferred`, and immediately remove it from door selection eligibility — the existing `GetAvailableForDoors()` filter already excludes deferred-status tasks, so no filter change is needed

**FR106:** Deferred tasks with a `defer_until` date in the past shall automatically return to `todo` status, checked both on application startup and periodically during sessions via a 1-minute `tea.Tick` interval — tasks deferred as "Someday" (nil `defer_until`) remain deferred until manually un-snoozed

**FR107:** The system shall provide a `:deferred` command in the command palette that displays all currently snoozed tasks sorted by return date (soonest first, Someday tasks last), with actions to un-snooze (`u` key, returns to todo immediately) or edit snooze date (`e` key, reopens SnoozeView)

**FR108:** The system shall log snooze events (task ID, defer-until date, snooze option selected) as a `snooze` event type in the JSONL session metrics log, and log auto-return events as `snooze_return` when deferred tasks are automatically restored to todo

**Snooze/Defer Status Transitions:**

**FR109:** The system shall support status transitions from `in-progress` to `deferred` and from `blocked` to `deferred`, in addition to the existing `todo` to `deferred` transition — enabling users to snooze tasks regardless of their current active state

---

## Phase 6+ - Expand/Fork Key Implementations (Accepted)

*The following requirements complete the Expand and Fork key actions in the TUI detail view, based on Design Decision H9.*

**Expand (Manual Sub-Task Creation):**

**FR120:** The system shall provide an Expand action via the `E` key in the detail view that enters a sequential subtask creation mode — after submitting one subtask (Enter), the input stays open for the next subtask until the user presses Esc to exit expand mode

**FR121:** Subtasks created via Expand shall have their `parent_id` field set to the parent task's ID, establishing a native parent-child relationship in the core task model — the `parent_id` field is an optional string pointer (`*string`) stored as `parent_id` in YAML, backward-compatible with existing tasks

**FR122:** The detail view shall display a subtask list below the task text when the viewed task has children, showing each subtask's status icon and text in an indented tree format, with a completion ratio summary (e.g., "Subtasks: 2/5 complete")

**FR123:** Parent tasks that have one or more subtasks shall be excluded from door selection by `GetAvailableForDoors()` — the user has decomposed the task, so only the subtasks should appear as doors

**FR124:** Subtasks shall NOT inherit effort, tags, or context from their parent — each subtask is an independent work item with its own metadata

**Fork (Variant Creation):**

**FR125:** The system shall provide a Fork action via the `F` key in the detail view that creates a variant of the current task by copying Text, Context, Effort, and Tags while resetting Status to `todo`, clearing Blocker and Notes, setting fresh timestamps, and adding a note "Forked from: [truncated original text]"

**FR126:** Fork variants shall be cross-referenced to the original task via the enrichment DB using a `forked-from` relationship type — the core `ForkTask` factory returns a concrete `*Task` and the main model establishes the cross-reference

---

## Phase 6+ - Task Dependencies & Blocked-Task Filtering (Accepted)

*The following requirements add a native dependency graph for tasks, ensuring the Three Doors only present genuinely actionable tasks by automatically filtering those whose prerequisites are incomplete.*

**Task Dependencies:**

**FR110:** The system shall support a `depends_on` field on tasks containing a list of task IDs that must be completed before the task becomes actionable — stored as `depends_on: [task-id-1, task-id-2]` in YAML and persisted through the enrichment DB for cross-provider dependencies

**FR111:** Tasks whose dependencies include any task not in `complete` status shall be automatically excluded from door selection by `GetAvailableForDoors()` — the filter checks dependency completion state on every door refresh, requiring no manual status management; if a dependency references a task ID that no longer exists in the pool, the dependency is treated as unmet (pessimistic) and the task remains blocked

**FR112:** The system shall display a "Blocked by: [task text]" indicator on tasks in the doors view and detail view when they have incomplete dependencies, showing the first incomplete dependency's text (truncated to 40 characters) with a count of additional blockers if more than one exists (e.g., "Blocked by: Review PR from alex... (+2 more)")

**FR113:** When a task is marked complete and other tasks depend on it, the system shall check all dependents — if a dependent's dependencies are now all complete, emit a `dependency_unblocked` notification event and refresh the doors view to potentially include the newly unblocked task

**FR114:** The system shall provide dependency management in the task detail view: `+` key to add a dependency (opens task search/picker), `-` key to remove a selected dependency, with the dependency list displayed in the detail view below the notes section

**FR115:** The system shall detect and reject circular dependencies when adding a new dependency — if adding dependency A->B would create a cycle (B already depends on A directly or transitively), the operation fails with a user-visible error message "Cannot add dependency: would create a circular chain"

---

## Phase 6+ - Undo Task Completion (Accepted)

*The following requirements allow users to reverse accidental task completion, addressing a validated pain point from the Phase 1 Validation Gate review.*

**Undo Task Completion:**

**FR127:** The system shall support undoing task completion by allowing the `complete → todo` status transition — when a user reverses a completed task, the `CompletedAt` timestamp is cleared, the task status is set to `todo`, and the task immediately becomes eligible for door selection again

**FR128:** The system shall log an `undo_complete` event in the JSONL session metrics when a task completion is reversed, capturing the task ID, original completion timestamp, and time elapsed since completion — enabling behavioral analysis of accidental completions

**FR129:** The system shall NOT modify the append-only completed task log (`completed.txt`) when a task completion is undone — the completed log remains an immutable audit trail; the undo is tracked separately via session metrics

**FR130:** The undo operation shall have no time limit — users can reverse a task completion regardless of how much time has elapsed since the task was originally completed

**FR131:** When a completed task is undone and that task was a dependency for other tasks (per FR113), the system shall re-evaluate dependent tasks — any dependents that were unblocked by the original completion shall have their dependency status rechecked, potentially returning to blocked state if the undone task was their only completed prerequisite

---

## Developer Experience & AI Agent Tooling (Accepted)

*The following requirements establish project-level AI agent alignment and developer workflow automation, based on findings from docs/research/ai-tooling-findings.md and party mode consensus (2026-03-08).*

**SOUL.md — Project Philosophy Document:**

**NFR-DX1:** The project shall maintain a SOUL.md document at the project root defining the project's philosophy, design principles, and behavioral guidelines for AI agents — ensuring consistent decision-making aligned with ThreeDoors values (progress over perfection, work with human nature, three doors not three hundred, local-first privacy-always, meet users where they are)

**Custom Claude Code Slash Commands:**

**NFR-DX2:** The project shall provide a `/pre-pr` Claude Code slash command that automates an 8-step pre-PR validation checklist (branch freshness, formatting via `gofumpt`, linting via `golangci-lint`, tests via `go test`, race detection via `go test -race`, dead code via `go vet`, scope review via `git diff`, commit cleanliness check) — reducing CI failures and enforcing NFR-CQ1 through NFR-CQ5

**NFR-DX3:** The project shall provide a `/validate-adapter` Claude Code slash command that checks TaskProvider implementations for interface compliance, error wrapping patterns, factory registration, test coverage, and atomic write usage

**NFR-DX4:** The project shall provide a `/check-patterns` Claude Code slash command that scans the codebase for design pattern violations (direct status mutation without `IsValidTransition()`, direct file writes bypassing atomic pattern, `fmt.Println` in TUI code, panics in user code, provider instantiation outside factory, missing error wrapping with `%w`)

**NFR-DX5:** The project shall provide a `/new-story` Claude Code slash command that generates story files from a standard template, referencing CLAUDE.md for coding standards and pre-PR checklists instead of embedding them

**NFR-DX6:** The project shall treat story files and specs as living documentation — completed stories MUST be updated retroactively when code improvements, architectural changes, or lessons learned diverge from what the specs describe. Specs are the authoritative system description; learning and improvements captured only in code (not reflected back in specs) is an anti-pattern. Deleting all code and rebuilding from specs alone should produce a better program, not a regression.

---

## Task Source Integration NFRs

> Requirements specific to API-based and IPC-based task source adapters.

**NFR20:** API-based adapters must handle HTTP 429 (Too Many Requests) responses by respecting the `Retry-After` header and applying exponential backoff with jitter before retrying

**NFR21:** API-based adapters must cache loaded tasks locally with a configurable TTL to avoid hitting rate limits on every TUI refresh; stale cache must be flagged in the UI

**NFR22:** Credential storage for external adapters must use environment variables or `~/.threedoors/config.yaml` settings — credentials must never be stored in task YAML files or sync state

**NFR23:** The per-provider circuit breaker must trip to Open state after 5 consecutive failures within a 2-minute window; probe requests must occur at configurable intervals (default 30 seconds, max 30 minutes with exponential backoff)

---
