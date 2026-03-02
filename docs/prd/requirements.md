# Requirements

## Technical Demo & Validation Phase Requirements

**Core Requirements (Week 1 - Build & Validate):**

**TD1:** The system shall provide a CLI/TUI interface optimized for terminal emulators (iTerm2 and similar)

**TD2:** The system shall read tasks from a simple local text file (e.g., `~/.threedoors/tasks.txt`)

**TD3:** The system shall display the Three Doors interface showing three tasks selected from the text file

**TD4:** The system shall allow users to select a door using 'a' or 'left arrow' for the left door, 'w' or 'up arrow' for the center door, and 'd' or 'right arrow' for the right door. Initially, or after re-rolling, no door shall be selected.

**TD5:** The system shall provide a refresh mechanism using 's' or 'down arrow' to generate a new set of three doors.

**TD6:** The system shall display the three doors with dynamic width adjustment based on the terminal size.

**TD7:** The system shall respond to the following keystrokes for task management, with functionality to be implemented in future stories:
    *   'c': Mark selected task as complete.
    *   'b': Mark selected task as blocked.
    *   'i': Mark selected task as in progress.
    *   'e': Expand selected task (into more tasks).
    *   'f': Fork selected task (clone/split).
    *   'p': Procrastinate/avoid selected task.

**TD8:** The system shall embed "progress over perfection" messaging in the interface

**TD9:** The system shall write completed tasks to a separate file (e.g., `~/.threedoors/completed.txt`) with timestamp

**Success Criteria for Phase:**
- Built and running within 4-8 hours of development time
- Developer uses it daily for 1 week to validate UX concept
- Three Doors selection feels meaningfully different from a simple list
- Decisions made on whether to proceed to Full MVP based on real usage

---

## Full MVP Requirements (Post-Validation - Deferred)

**Phase 2 - Apple Notes Integration:**

**FR2:** The system shall integrate with Apple Notes as the primary task storage backend, enabling bidirectional sync

**FR4:** The system shall retrieve and display tasks from Apple Notes within the application interface

**FR5:** The system shall allow users to mark tasks as complete, updating both the application state and Apple Notes

**FR12:** The system shall allow updates to tasks from either the application or directly in Apple Notes on iPhone, with changes reflected bidirectionally

**FR15:** The system shall provide a health check command to verify Apple Notes connectivity and database integrity

**Phase 3 - Enhanced Interaction & Learning:**

**FR3:** The system shall allow users to capture new tasks with optional context (what and why) through the CLI/TUI

**FR6:** The system shall display user-defined values and goals persistently throughout task work sessions

**FR7:** The system shall provide a "choose-your-own-adventure" interactive navigation flow that presents options rather than demands

**FR8:** The system shall track daily task completion count and display comparison to previous day's count

**FR9:** The system shall prompt the user once per session with: "What's one thing you could improve about this list/task/goal right now?"

**FR10:** The system shall embed "progress over perfection" messaging throughout interaction patterns and interface copy (enhanced beyond Tech Demo)

**FR16:** The system shall support a "quick add" mode for capturing tasks with minimal interaction

**FR18:** The system shall allow users to provide feedback on why a specific door isn't suitable with options: Blocked, Not now, Needs breakdown, or Other comment

**FR19:** The system shall capture and store blocker information when a task is marked as blocked

**FR20:** The system shall use door selection and feedback patterns to inform future door selection (learning which task types suit which contexts)

**FR21:** The system shall categorize tasks by type, effort level, and context to enable diverse door selection

**Phase 4 - Distribution & Packaging (macOS):**

**FR22:** The system shall provide macOS binaries that are code-signed with a valid Apple Developer certificate so that Gatekeeper does not quarantine the binary on download

**FR23:** The system shall be notarized with Apple's notarization service so that Gatekeeper allows execution without requiring users to bypass security warnings

**FR24:** The system shall be installable via Homebrew using a custom tap (`brew install arcaven/tap/threedoors`), with the formula downloading the appropriate signed binary for the user's architecture

**FR25:** The system shall provide a DMG or pkg installer as an alternative installation method for users who prefer graphical installation over Homebrew

**FR26:** The CI/CD pipeline shall automate code signing, notarization, Homebrew formula updates, and installer generation as part of the release process

**Phase 5 - Data Layer & Enrichment:**

**FR11:** The system shall maintain a local enrichment layer (SQLite and/or vector database) for metadata, cross-references, and relationships that cannot be stored in source systems

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

**NFR3:** The system shall operate on macOS as primary platform, with binaries that are code-signed and notarized for seamless Gatekeeper approval

**NFR4:** The system shall store all user data locally or in user's iCloud (via Apple Notes), with no external telemetry or tracking

**NFR5:** The system shall store application state and enrichment data locally (cross-computer sync deferred to post-MVP)

**NFR6:** The system shall respond to user interactions within the CLI/TUI with minimal latency (target: <500ms for typical operations)

**NFR7:** The system shall provide graceful degradation when Apple Notes integration is unavailable, maintaining core functionality

**NFR8:** The system shall implement secure credential storage using OS keychain for any API keys or authentication tokens

**NFR9:** The system shall never log sensitive user data or credentials

**NFR10:** The system shall use Make as the build system

**NFR11:** The system shall maintain clear architectural separation between core engine, TUI layer, integration adapters, and enrichment storage

**NFR12:** The system shall maintain data integrity even when Apple Notes is modified externally while app is running

---

## Code Quality & Submission Standards

These non-functional requirements establish code quality gates that all contributions must pass before submission. They are derived from recurring patterns identified across 31+ pull requests (see `docs/architecture/pr-submission-standards.md` for full rationale and evidence).

**NFR-CQ1:** All code must pass `gofumpt` formatting before submission — running `gofumpt -l .` must produce no output (Evidence: PRs #9, #10, #23, #24 required formatting fix-up commits)

**NFR-CQ2:** All code must pass `golangci-lint run ./...` with zero issues before submission, including `errcheck` and `staticcheck` analyzers (Evidence: PR #16 required 2 extra commits to resolve lint findings)

**NFR-CQ3:** All branches must be rebased onto `upstream/main` before PR creation — no merge commits, no stale branches (Evidence: PRs #3, #5, #19, #23 were delayed by merge conflicts from stale branches)

**NFR-CQ4:** All PRs must have a clean `git diff --stat` showing only in-scope changes — no unrelated files, no accidental directory additions (Evidence: PR #5 included an unrelated `agents/` directory)

**NFR-CQ5:** All fix-up commits must be squashed before PR submission — PRs should contain a single clean commit or logically separated commits, not iterative fix-up trails (Evidence: ~15 PRs contained avoidable "fix: address code review findings" commits)

---
