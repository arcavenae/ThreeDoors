# User Journeys

## Journey 1: Daily Task Selection (Tech Demo)

**User State:** Developer opens ThreeDoors to decide what to work on next.

**Flow:**
1. User launches `threedoors` in terminal
2. Three doors appear showing three diverse tasks from `tasks.txt`
3. User scans the three options (< 5 seconds decision time target)
4. User selects a door (A/W/D or arrow keys) or refreshes for new options (S/down arrow)
5. Selected door expands to show task detail view with status actions
6. User marks task status (complete, blocked, in progress, expand, fork, or procrastinate)
7. Three new doors appear automatically

**Supported By:** TD1, TD2, TD3, TD4, TD5, TD6, TD7, TD8, TD9

**Success Criteria:** User completes at least one task per session; Three Doors selection takes less time than scanning a full list.

---

## Journey 2: Quick Task Search (Tech Demo)

**User State:** Developer knows which task they want to find but it is not currently displayed in the three doors.

**Flow:**
1. User presses `/` to open search mode
2. Types search text; matching tasks appear bottom-up as live results
3. User navigates results with arrow keys or HJKL
4. Presses Enter to open task in detail view
5. Takes action on task or presses Esc to return to search with text preserved

**Supported By:** TD1, Story 1.3a requirements

**Success Criteria:** User finds target task within 3 keystrokes of typing.

---

## Journey 3: Quick Task Capture (Tech Demo)

**User State:** Developer thinks of a new task while working and wants to capture it without leaving the terminal.

**Flow:**
1. User presses `/` then types `:add Buy groceries`
2. Task is appended to `tasks.txt`
3. User returns to three doors view
4. New task is available in the task pool for future door selections

**Supported By:** Story 1.3a `:add` command

**Success Criteria:** Task captured in under 5 seconds without leaving the TUI.

---

## Journey 4: Mood-Aware Session (Tech Demo)

**User State:** Developer wants to log current emotional state to build data for future adaptive selection.

**Flow:**
1. User presses `M` from door view at any time
2. Mood dialog shows options: Focused, Tired, Stressed, Energized, Distracted, Calm, Other
3. User selects mood (or types custom text for Other)
4. Mood is timestamped and recorded in session metrics
5. Returns to door view immediately

**Supported By:** Story 1.3 mood tracking, Story 1.5 session metrics

**Success Criteria:** Mood captured in under 3 seconds; mood data appears in `sessions.jsonl`.

---

## Journey 5: Session Review (Tech Demo)

**User State:** Developer wants to see how productive the current session has been.

**Flow:**
1. User presses `/` then types `:stats`
2. Session statistics display: tasks completed, doors viewed, time in session, refreshes used
3. User reviews progress and returns to door view

**Supported By:** Story 1.3a `:stats` command, Story 1.5 session metrics

**Success Criteria:** Stats display within 100ms; completion count matches actual completions.

---

## Journey 6: Apple Notes Task Management (Post-Validation)

**User State:** Developer captures tasks on iPhone via Apple Notes and wants them available in ThreeDoors on Mac.

**Flow:**
1. User adds tasks to Apple Notes on iPhone
2. User launches ThreeDoors on Mac
3. ThreeDoors syncs with Apple Notes and loads new tasks into pool
4. User selects and completes tasks via Three Doors interface
5. Completions sync back to Apple Notes
6. Health check command verifies connectivity

**Supported By:** FR2, FR4, FR5, FR12, FR15

**Success Criteria:** Sync completes within 2 seconds; bidirectional changes reflected on next app launch.

---

## Journey 7: Extended Task Capture with Context (Post-Validation)

**User State:** Developer wants to capture not just a task but why it matters.

**Flow:**
1. User enters extended capture mode
2. Provides task description and optional context (why this matters, effort level, type)
3. Task is stored with full metadata
4. Context is available in detail view and feeds into learning algorithms

**Supported By:** FR3, FR16, FR21

**Success Criteria:** Extended capture completes in under 30 seconds; context retrievable in detail view.

---

## Journey 8: Adaptive Door Selection (Post-Validation)

**User State:** Developer has used ThreeDoors for several weeks; system has learned patterns.

**Flow:**
1. User opens ThreeDoors and logs mood as "Tired"
2. System uses historical mood-task correlation data to select doors
3. Doors show lower-effort, quick-win tasks appropriate for tired state
4. User completes a task; system reinforces the pattern
5. System surfaces insight: "When tired, you complete 2x more quick-win tasks"

**Supported By:** FR20, FR21, Epic 4 (Learning & Intelligent Door Selection)

**Success Criteria:** Door selection patterns differ measurably based on mood state; user reports doors feel more relevant.

---

## Journey 9: Door Theme Customization

**User State:** User wants to personalize the Three Doors appearance to match their aesthetic preference.

**Flow:**
1. During first-run onboarding, user is presented with a horizontal preview of available door themes
2. User browses themes with arrow keys, seeing doors rendered in each theme style
3. User selects preferred theme (e.g., Modern/Minimalist, Sci-Fi/Spaceship, Japanese Shoji, or Classic)
4. Theme is applied immediately — all three doors render with the chosen theme frame
5. Later, user types `:theme` to open the theme selection view
6. User previews and switches to a different theme; change takes effect immediately without restart
7. Selected theme persists in `~/.threedoors/config.yaml`
8. When terminal is too narrow for the active theme, doors gracefully fall back to Classic rendering

**Supported By:** FR55, FR56, FR57, FR58, FR59, FR60, FR61, FR62

**Success Criteria:** Theme change applies instantly to all three doors; config.yaml reflects selection after restart; narrow terminal triggers Classic fallback without error.

---

## Journey 10: First-Run Onboarding

**User State:** User has just installed ThreeDoors and launches it for the first time with no existing configuration or tasks.

**Flow:**
1. User runs `threedoors` in terminal for the first time
2. ThreeDoors detects no config file and enters the onboarding flow
3. Empty state view explains the Three Doors concept and prompts the user to add their first tasks
4. User adds 3-5 tasks via guided inline prompts
5. Adapter registry presents available task providers (local text file, Apple Reminders, Jira, etc.)
6. User selects a provider and configures it; config is written to `~/.threedoors/config.yaml`
7. Horizontal theme preview appears — user browses and selects a door theme
8. First set of three doors appears with the user's tasks rendered in the chosen theme
9. User selects a door and completes their first task

**Supported By:** FR31, FR32, FR33, FR55, FR56, FR57, FR58, FR59

**Success Criteria:** New user goes from install to first task completion in under 3 minutes; config.yaml is created with valid provider and theme settings; no errors or blank screens during the flow.

---

## Journey 11: Task Source Connection & Multi-Source Usage

**User State:** User has been using ThreeDoors with local tasks and wants to connect an external task source (e.g., Jira, Todoist, GitHub Issues) to pull in work tasks alongside personal ones.

**Flow:**
1. User decides to connect their Jira board and runs the setup wizard via `:connect` command
2. Wizard prompts for provider type, server URL, and authentication credentials
3. User enters API token; ThreeDoors validates the connection and confirms sync access
4. Initial sync pulls issues from the configured JQL filter into the local task pool
5. Three doors now show a mix of local tasks and Jira issues, each with source attribution
6. User selects a door showing a Jira issue and marks it complete
7. Completion syncs back to Jira — the issue transitions to "Done" in the remote system
8. User checks sync status indicator in the TUI to confirm bidirectional sync is healthy

**Supported By:** FR63, FR64, FR65, FR66, FR46, FR48, FR70, FR71, FR72

**Success Criteria:** External tasks appear in doors within one sync cycle; source attribution is visible on each task; completing a task in ThreeDoors reflects in the remote system; sync status indicator shows healthy state.

---

## Journey 12: Daily Planning Mode

**User State:** User starts their workday and wants to review yesterday's progress and set today's focus before diving into tasks.

**Flow:**
1. User runs `threedoors plan` or types `:plan` in the TUI to enter planning mode
2. Step 1 (Review): Yesterday's incomplete tasks are presented one by one — user triages each as continue, skip, or snooze
3. Snoozed tasks open the SnoozeView for date selection and are deferred out of the active pool
4. Step 2 (Select): Full task pool is displayed with energy-aware sorting — system infers energy level from time of day (morning = high) and suggests matching tasks
5. User overrides energy level to "medium" and selects 3 tasks as today's focus
6. Step 3 (Confirm): Summary shows focus tasks tagged with `+focus`; user confirms the plan
7. Planning session metrics are logged (duration, tasks reviewed, focus count)
8. User returns to three doors view — focus-tagged tasks appear more frequently as doors throughout the day

**Supported By:** FR97, FR98, FR99, FR100, FR101, FR102, FR103

**Success Criteria:** Planning session completes in under 5 minutes; focus tasks appear at elevated frequency in door selection; planning metrics are recorded in session log; energy override persists for the session.

---

## Journey 13: Snooze & Defer Workflow

**User State:** User encounters a task in the doors that isn't actionable right now but will be relevant next week.

**Flow:**
1. User sees a task in the three doors view and selects it
2. User presses `Z` to open the snooze action
3. Snooze options appear: Tomorrow, Next Week, Pick Date, Someday
4. User selects "Next Week" — task is deferred until next Monday 9:00 AM
5. Task immediately disappears from door selection; doors refresh with a new task in its place
6. Snooze event is logged in the JSONL session metrics
7. On Monday morning, the deferred task automatically returns to `todo` status
8. Task reappears in door selection; user completes it
9. User types `:deferred` to review all currently snoozed tasks, un-snoozes one early

**Supported By:** FR104, FR105, FR106, FR107, FR108, FR109

**Success Criteria:** Snoozed task is removed from doors within the same session; task auto-returns on the scheduled date; `:deferred` view shows all snoozed tasks sorted by return date; snooze events appear in session log.

---

## Journey 14: CLI Task Management

**User State:** Power user wants to manage tasks non-interactively from the command line or integrate ThreeDoors into shell scripts.

**Flow:**
1. User runs `threedoors task list` to see all current tasks in a formatted table
2. User pipes `threedoors task list --json` into `jq` to filter tasks by tag
3. User adds a new task: `threedoors task add "Review quarterly OKRs"`
4. User checks current doors: `threedoors doors` shows the three active doors
5. User picks a door without entering the TUI: `threedoors doors pick 2`
6. User completes the task: `threedoors task complete <id>`
7. User reviews session stats: `threedoors stats` shows completions and session metrics
8. User generates shell completions: `threedoors completion zsh >> ~/.zshrc`

**Supported By:** FR81, FR82, FR83, FR84, FR85

**Success Criteria:** All commands return valid output in both human-readable and `--json` modes; shell completions work without errors; commands are composable in shell pipelines.

---
