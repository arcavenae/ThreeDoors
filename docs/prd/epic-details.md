# Epic Details

## Epic 1: Three Doors Technical Demo

**Epic Goal:** Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists.

**Scope:** CLI/TUI application that reads tasks from a text file, presents three random tasks as "doors," allows refresh and selection, marks tasks complete, and tracks progress.

**Story Sequence Optimization:** Stories reordered to validate refresh UX before completion (moved 1.6→1.4). Non-essential features simplified or made optional to focus on core validation.

---

### Story 1.1: Project Setup & Basic Bubbletea App

**As a** developer,
**I want** a working Go project with Bubbletea framework,
**so that** I have a foundation for building the Three Doors TUI.

**Acceptance Criteria:**
1. Go module initialized with `go mod init github.com/arcaven/ThreeDoors`
2. Bubbletea and Lipgloss dependencies added
3. Basic TUI application renders "ThreeDoors - Technical Demo" header
4. Application responds to 'q' keypress to quit
5. `Makefile` with `build`, `run`, and `clean` targets works
6. Application compiles and runs without errors

**Estimated Time:** 30-45 minutes

---

### Story 1.2: Display Three Doors from a Task File

**As a** developer,
**I want** the application to read tasks from a text file and display three of them as "doors",
**so that** I can see the core interface of the application.

**Acceptance Criteria:**
1. On startup, the application reads tasks from `~/.threedoors/tasks.txt`
2. If `tasks.txt` does not exist, it is created with a default set of 3-5 sample tasks
3. A `FileManager` component is created in `internal/tasks/file_manager.go` to handle file reading and creation
4. A `Task` model is defined in `internal/tasks/task.go` to represent a task (text content)
5. The main application view displays three randomly selected tasks from the file
6. The display shows three randomly selected tasks, distributed horizontally from left to right, without "Door X" labels
7. The application responds to 'a' or 'left arrow' for the left door, 'w' or 'up arrow' for the middle/center door, and 'd' or 'right arrow' for the right door
8. Pressing 's' or 'down arrow' re-rolls the doors, presenting a new set of three tasks
9. Initially, or after re-rolling, no door is selected/active
10. The application responds to the following keystrokes for task management (functionality to be implemented in future stories):
    - 'c': Mark selected task as complete
    - 'b': Mark selected task as blocked
    - 'i': Mark selected task as in progress
    - 'e': Expand selected task (into more tasks)
    - 'f': Fork selected task (clone/split)
    - 'p': Procrastinate/avoid selected task
11. The application still quits on 'q' or 'ctrl+c'

**Key Design Decisions:**
- Three doors are rendered horizontally, each occupying approximately 1/3rd of the terminal width (dynamic adjustment)
- No "Door X" labels displayed to reduce visual clutter
- No door is selected initially or after re-rolling to avoid bias
- Arrow keys provided as alternative to WASD for accessibility

**Estimated Time:** 60-90 minutes

---

### Story 1.3: Door Selection & Task Status Management

**As a** user,
**I want** to select a door and update the task's status,
**so that** I can take action on tasks and track my progress.

**Acceptance Criteria:**
1. Pressing A/Left Arrow, W/Up Arrow, or D/Right Arrow selects the corresponding door (left, center, right)
2. Selected task is highlighted/indicated visually
3. **Door Opening Animation & Expanded Detail View:**
   - When door is selected (or Enter is pressed), door presents optional animation as if opening
   - Selected door shifts to left position and expands to fill the screen
   - Task detail view displays:
     - Task text (full, not truncated)
     - Any existing task metadata/details (status, notes, timestamps, etc.)
     - Status action menu with all available options
   - **Esc** key closes the door and returns to previous screen (context-aware):
     - If opened from three doors view → returns to three doors view
     - If opened from search (Story 1.3a) → returns to search view with text preserved and previous selection highlighted
4. Status action menu shows available options:
   - **C**: Mark as Complete
   - **B**: Mark as Blocked
   - **I**: Mark as In Progress
   - **E**: Expand (break into more tasks)
   - **F**: Fork (clone/split task)
   - **P**: Procrastinate (defer task)
   - **R**: Flag for Rework
   - **M**: Log Mood/Context (also available from door view without selection)
   - **Esc**: Close door and return to previous screen
5. Pressing any status key (within expanded detail view) applies that status to the selected task
6. Pressing **M** from door view (no task selection needed) opens mood capture dialog:
   - Multiple choice options: "Focused", "Tired", "Stressed", "Energized", "Distracted", "Calm", "Other"
   - If "Other" selected, prompt for custom text input (word or phrase)
   - Mood entries are timestamped and recorded in session metrics
   - Returns to door view immediately after capture
7. Completed tasks are removed from available task pool (in-memory) and appended to `~/.threedoors/completed.txt` with timestamp
8. Blocked/deferred/rework tasks remain in the pool but are tagged with status
9. New set of three doors is displayed automatically after any status change (door closes and returns to three doors view)
10. Session completion count increments for completed tasks (e.g., "✓ Completed this session: 3")
11. "Progress over perfection" message shown after completing a task
12. Door selection is tracked: which door position (left/center/right) was selected
13. Task bypass is tracked: doors shown but not selected before refresh
14. Mood entries are tracked with timestamps for correlation with task selection patterns

**Key Design Decisions:**
- **Door opening animation is optional** but provides delightful visual feedback
- **Expanded detail view** shifts door left and fills screen for focused interaction
- **Context-aware return** (Esc) maintains navigation state (critical for search integration in 1.3a)
- All status changes should be tracked for future pattern analysis
- Door position preferences (left vs center vs right) captured for learning
- Tasks that are expanded or forked create new task entries in tasks.txt
- Blocked tasks should prompt for optional blocker note
- **Mood tracking is casual and low-friction** - accessible anytime via 'M' key without needing task selection
- Mood data provides crucial context for correlating emotional state with task selection behavior
- Multiple choice moods keep capture quick; custom text option allows nuanced expression

**Estimated Time:** 90-120 minutes (including door animation, expanded detail view, mood capture UI)

---

### Story 1.3a: Quick Search & Command Palette

**As a** user,
**I want** to quickly search for specific tasks and execute commands via a text input interface,
**so that** I can efficiently find and act on tasks without scrolling through the three doors.

**Acceptance Criteria:**

**Search Mode (Default):**
1. Pressing **/** key from three doors view opens text input field at bottom of screen
2. Text input appears with placeholder text: "Search tasks... (or :command for commands)"
3. As user types, live substring matching filters task list:
   - Matching tasks display from **bottom-up** extending up the screen
   - List updates continuously as each character is entered
   - If no matches found, show message: "No tasks match '<search text>'"
   - Empty input shows no results (doesn't show all tasks)
4. **Navigation within search results:**
   - **Arrow keys** (up/down/left/right): Navigate through search results
   - **A/S/D/W keys**: Navigate through search results (s=down, w=up, a/d reserved for future horizontal navigation)
   - **H/J/K/L keys** (vi-style): Navigate through search results (j=down, k=up, h/l reserved)
   - Selected result is highlighted
   - Enter key: Opens selected task in expanded detail view (same as Story 1.3 door selection)
5. **Task detail from search:**
   - When task is opened via Enter, shows same expanded detail view as Story 1.3
   - **Esc** from detail view returns to search with:
     - Search text preserved in input field
     - Previously selected task still highlighted
     - User can continue searching or refine search
6. **Exit search mode:**
   - **Esc** key (when in search input, not in task detail) clears search and returns to three doors view
   - **Ctrl+C** also exits search mode

**Command Mode (vi-style):**
7. Typing **:** as first character in empty text input switches to command mode
8. Command mode indicated by visual cue (e.g., prompt changes to ":")
9. **Available commands:**
   - **:add <task text>** - Add new task to tasks.txt
   - **:edit** - Edit current task list file directly (opens in $EDITOR or vim)
   - **:mood [mood]** - Quick mood log (optional mood parameter, otherwise prompts)
   - **:stats** - Show session statistics (tasks completed, doors viewed, time in session, etc.)
   - **:chat** - Open AI chat interface for task-related questions/help (deferred implementation)
   - **:help** - Display available commands and key bindings
   - **:quit** or **:exit** - Exit application
10. Commands execute on Enter key
11. Invalid commands show error: "Unknown command: '<command>'. Type :help for available commands."
12. **Esc** exits command mode and returns to three doors view

**Key Design Decisions:**
- **Bottom-up list display** reduces eye travel distance from input field
- **Multiple navigation schemes** (arrows, WASD, HJKL) accommodate different user preferences
- **Live substring matching** provides instant feedback
- **Context preservation** (search text + selection) when returning from task detail is critical for UX
- **Command palette** (`:`) provides power-user features without cluttering main UI
- **:chat command deferred** to post-validation (placeholder for future AI integration)
- Search only matches task text (no metadata search in Tech Demo phase)

**Estimated Time:** 90-120 minutes (search mode + navigation + command parsing)

---

### Story 1.5: Session Metrics Tracking

**As a** developer validating the Three Doors concept,
**I want** objective session metrics collected automatically,
**so that** I can make a data-informed decision at the validation gate instead of relying solely on subjective impressions.

**Priority:** HIGH (enables objective validation)

**Context:** The validation decision gate asks: "Does Three Doors reduce friction vs. traditional task lists?" Without metrics, this is purely subjective. This story adds lightweight, non-intrusive tracking to provide objective evidence. The metrics should be invisible to the user during normal operation - no prompts, no UI changes, just silent background collection.

**Future Pattern Analysis Foundation:** Capturing door selection patterns (left vs center vs right), bypass behaviors (which tasks are skipped), status change patterns (what gets blocked, procrastinated, or completed), and **mood/emotional context** creates the data foundation for Epic 4 (Learning & Intelligent Door Selection). Over time, the application will use this data to:
- Predict which types of tasks the user tends to skip or avoid
- Identify patterns that indicate blockers or discouragement
- **Correlate emotional state with task selection behavior** (e.g., "stressed" → avoids complex tasks)
- Surface insights to help users understand their work patterns
- Adapt door selection to encourage balanced progress across different task types
- Help users re-evaluate goals when persistent avoidance patterns emerge

**Acceptance Criteria:**

1. **SessionTracker component created** in `internal/tasks/session_tracker.go`
   - Tracks session_id, start/end times, duration
   - Tracks behavioral counters: tasks_completed, doors_viewed, refreshes_used, detail_views, notes_added, status_changes, mood_entries
   - Tracks time_to_first_door_seconds (key friction metric)
   - **NEW: Door selection patterns** - tracks which door position selected (left=0, center=1, right=2) per selection
   - **NEW: Task bypass tracking** - records tasks shown in doors but not selected before refresh
   - **NEW: Status change details** - records what status was applied (complete, blocked, in-progress, expand, fork, procrastinate, rework)
   - **NEW: Task content capture** - stores task text with each interaction for future pattern analysis
   - **NEW: Mood tracking** - captures timestamped mood entries (predefined or custom text) for correlation with task behavior
   - Constructor `NewSessionTracker()` initializes with UUID and current timestamp
   - Methods: `RecordDoorViewed(doorPosition int)`, `RecordRefresh(doorTasks []string)`, `RecordDetailView()`, `RecordTaskCompleted(taskText string)`, `RecordNoteAdded()`, `RecordStatusChange(status string, taskText string)`, `RecordDoorSelection(doorPosition int, taskText string)`, `RecordMood(mood string, customText string)`
   - Method `Finalize()` calculates duration and returns metrics for persistence
   - Mood entries stored as array: `[{timestamp, mood, custom_text}]`

2. **MetricsWriter component created** in `internal/tasks/metrics_writer.go`
   - Constructor `NewMetricsWriter(baseDir string)` sets path to sessions.jsonl
   - Method `AppendSession(metrics *SessionMetrics)` writes JSON line to file
   - Creates file if doesn't exist, appends to existing file
   - Returns error on failure (caller logs warning, doesn't crash)

3. **SessionTracker integrated into MainModel**
   - MainModel includes sessionTracker field
   - SessionTracker passed to DoorsView and TaskDetailView constructors
   - No UI changes (completely invisible to user)

4. **Recording calls integrated into DoorsView**
   - Door selection (1/2/3) calls `RecordDoorViewed()`
   - Refresh (R) calls `RecordRefresh()`
   - Mood capture (M) calls `RecordMood()` with selected mood and optional custom text
   - Recording happens before transitioning to detail view

5. **Recording calls integrated into TaskDetailView**
   - Constructor calls `RecordDetailView()` on initialization
   - Note addition calls `RecordNoteAdded()`
   - Status change calls `RecordStatusChange()`
   - Completion calls both `RecordStatusChange()` and `RecordTaskCompleted()`

6. **Session persistence on app exit**
   - `cmd/threedoors/main.go` calls `Finalize()` and `AppendSession()` on clean exit
   - Write failures logged as warning to stderr, don't block exit
   - File created: `~/.threedoors/sessions.jsonl` (JSON Lines format)

7. **Metrics file format validated**
   - Each line is valid JSON (parseable by `jq`)
   - File is append-only, human-readable
   - Manual verification: run app 2-3 times, verify lines in sessions.jsonl

8. **Performance requirements met**
   - Recording adds <1ms overhead per event
   - No UI lag or stuttering observed
   - Memory overhead negligible (<1KB per session)

9. **Error handling implemented**
   - File write failures don't crash app (warning logged)
   - JSON marshal failures don't crash app
   - No error dialogs shown to user

**Analysis Scripts Created:**
- `scripts/analyze_sessions.sh` - Session summary and averages
- `scripts/daily_completions.sh` - Daily completion counts from completed.txt
- `scripts/validation_decision.sh` - Automated validation criteria evaluation

**Deferred to Epic 4 (Learning & Intelligent Door Selection):**
- Pattern analysis algorithms (which task types are consistently avoided)
- ML-based task selection optimization
- User insight reports ("You tend to skip tasks containing X")
- Adaptive door selection based on historical patterns
- Goal re-evaluation prompts when persistent avoidance detected
- Task categorization and tagging for pattern recognition

**Deferred to Future:**
- Daily aggregation reports (manual analysis via scripts)
- In-app metrics visualization
- Friction score prompts (manual logging recommended)
- Metrics export to other formats
- Historical cleanup/rotation

**Rationale:** Provides objective data to answer "Does Three Doors reduce friction?" Metrics enable data-informed decision at validation gate rather than relying solely on subjective feel. Enhanced tracking of door selection patterns, task bypass behaviors, and **mood/emotional context** creates the data foundation needed for future learning and adaptation features (Epic 4). Mood correlation will help identify which emotional states lead to productive task selection vs avoidance.

**Estimated Time:** 85-100 minutes (enhanced tracking + mood capture)

---

### Story 1.6: Essential Polish (SIMPLIFIED)

**As a** user,
**I want** the app to feel polished enough to use daily,
**so that** I enjoy the validation experience.

**Acceptance Criteria:**
1. Lipgloss styling applied: distinct colors for doors, success messages (green), prompts (yellow/blue)
2. "Progress over perfection" message embedded in interface (startup greeting or post-completion)
3. Application feels responsive and smooth (no noticeable lag)

**Deferred to MVP:**
- README.md (you're the only user for validation)
- Extensive edge case handling (all tasks completed celebration, 1-2 tasks remaining display logic)
- Advanced error messaging

**Rationale for Simplification:** Focus on making the core experience pleasant. Edge cases are unlikely to be hit during 1-week validation. README isn't needed when you built it.

**Estimated Time:** 20-30 minutes

---

## Epic 2-5: Post-Validation Epics (Placeholder)

**Note:** These epics are placeholders for post-validation planning. Detailed stories will be created only if Epic 1 successfully validates the Three Doors concept.

**Epic 2: Foundation & Apple Notes Integration**
*Stories to be defined after Epic 1 validation and Apple Notes integration spike*

**Epic 3: Enhanced Interaction & Task Context**
*Stories to be defined based on learnings from Epic 2 usage patterns*

**Epic 4: Learning & Intelligent Door Selection**

**Goal:** Use historical session metrics (captured in Story 1.4) to analyze user patterns and adapt door selection to improve task engagement and completion rates.

**Key Capabilities to Build:**
- **Pattern Recognition:** Analyze which types of tasks are consistently selected vs bypassed
- **Mood Correlation Analysis:** Identify which emotional states (focused, stressed, tired, etc.) correlate with task selection, avoidance, or completion patterns
- **Avoidance Detection:** Flag tasks or task patterns that are repeatedly shown but never selected
- **Status Pattern Analysis:** Track which task types tend to get blocked, procrastinated, or reworked (correlated with mood state)
- **Adaptive Selection:** Adjust door algorithm based on current mood state and historical patterns (e.g., show simpler tasks when user logs "tired")
- **User Insights:** Surface reports like "When stressed, you tend to avoid complex technical tasks" or "Your highest completion rate is when feeling 'focused'"
- **Goal Re-evaluation Prompts:** When persistent avoidance detected (especially with specific mood patterns), suggest user review if tasks align with goals
- **Encouragement System:** Gently encourage work on task types that haven't been touched in a while, with mood-aware messaging
- **Position Preference Analysis (Minor):** Track if user has bias toward certain door positions (left/center/right)

**Data Foundation:** Epic 1 Story 1.4 creates the metrics infrastructure capturing door position selections, task bypasses, status changes, and **mood/emotional context**—all essential for pattern analysis. Mood tracking enables correlation between emotional state and work behavior, allowing adaptive task selection based on current user state.

*Detailed stories to be defined based on sufficient usage data from Epic 3*

**Epic 5: Data Layer & Enrichment (Optional)**
*Stories to be defined only if clear need emerges from Epic 4*

---
