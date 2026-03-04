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
