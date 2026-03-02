# Story 1.6: Essential Polish

Status: done

## Story

As a user,
I want the app to feel polished enough to use daily,
so that I enjoy the validation experience.

## Acceptance Criteria

1. **Given** the application is running **When** any screen is rendered **Then** Lipgloss styling is applied with distinct colors: doors have their own color, success messages are green, prompts are yellow/blue

2. **Given** the application starts or a task is completed **When** the corresponding event occurs **Then** a "Progress over perfection" message is displayed (startup greeting or post-completion)

3. **Given** the user interacts with the application **When** any action is performed **Then** the response feels responsive and smooth with no noticeable lag

## Tasks / Subtasks

- [ ] **Task 1: Enhance Door Visual Styling** (AC: 1)
  - [ ] 1.1: Add distinct door colors — each door gets a unique accent color (e.g., door 0 = cyan "86", door 1 = magenta "212", door 2 = yellow "220") for border foreground on unselected doors
  - [ ] 1.2: Make selected door style more prominent — use bold title rendering and brighter/contrasting border color (e.g., white "255" border)
  - [ ] 1.3: Add subtle door number/position indicator styling (left/center/right) via color-coded key hints below each door

- [ ] **Task 2: Add Startup Greeting Message** (AC: 2)
  - [ ] 2.1: Create a pool of 5-10 "Progress over perfection" startup greeting messages as a package-level `var greetingMessages []string` in `styles.go`
  - [ ] 2.2: Add `greeting string` field to `DoorsView` struct, set once in `NewDoorsView()` via `rand.IntN()` from `math/rand/v2` (Go 1.25.4+ auto-seeds, no manual seed needed)
  - [ ] 2.3: Display the greeting in `DoorsView.View()` — insert AFTER `headerStyle.Render("ThreeDoors - Technical Demo")` on line 60 and BEFORE the door count check on line 63
  - [ ] 2.4: Style the greeting with `greetingStyle` (color "229" soft yellow)
  - [ ] 2.5: Add non-consecutive greeting guard — track `lastGreetingIdx int` field on `DoorsView`, skip that index when selecting next greeting to avoid immediate repeats on app restart

- [ ] **Task 3: Enhance Post-Completion Celebration Messages** (AC: 2)
  - [ ] 3.1: Create a pool of 8-12 varied completion celebration messages as a package-level `var celebrationMessages []string` in `styles.go`
  - [ ] 3.2: In `main_model.go` `TaskCompletedMsg` handler (line 136), replace static `m.flash` string with `celebrationMessages[rand.IntN(len(celebrationMessages))]` using `math/rand/v2`
  - [ ] 3.3: Style celebration messages with green color "82" and bold (already partially done via flashStyle)

- [ ] **Task 4: Polish Detail View Styling** (AC: 1)
  - [ ] 4.1: Add color-coded status badge in detail view (green for complete-able, red for blocked, yellow for in-progress)
  - [ ] 4.2: Style the action menu keys with distinct colors — action keys in magenta "212", navigation keys in gray "241"
  - [ ] 4.3: Add a visual separator between task content and action menu

- [ ] **Task 5: Polish Search View Styling** (AC: 1)
  - [ ] 5.1: Style the search prompt with blue/yellow accent (prompt label in "39" blue, cursor in "220" yellow)
  - [ ] 5.2: Ensure command mode indicator is visually distinct (already has commandModeStyle "214")
  - [ ] 5.3: Style search results with status-colored indicators

- [ ] **Task 6: Polish Mood View Styling** (AC: 1)
  - [ ] 6.1: Style mood options with emoji-like visual indicators or colored bullets
  - [ ] 6.2: Ensure selected mood option has clear visual highlight

- [ ] **Task 7: Add Footer Branding/Polish** (AC: 1, 2)
  - [ ] 7.1: Style the help text footer consistently across all views with "241" gray
  - [ ] 7.2: Style the "Progress over perfection" tagline in the footer with soft warm color

- [ ] **Task 8: Update Existing Test Assertions** (AC: all)
  - [ ] 8.1: Update `doors_view_test.go` — any assertions checking for exact View() output strings must accommodate new greeting text in the output
  - [ ] 8.2: Update `main_model_test.go` — tests checking for exact flash message "Progress over perfection. Just pick one and start." must use `contains` checks against the `celebrationMessages` pool, or verify flash is non-empty
  - [ ] 8.3: Add test in `doors_view_test.go` verifying greeting is non-empty and is one of the known `greetingMessages` pool entries
  - [ ] 8.4: Verify all existing tests still pass after style changes

- [ ] **Task 9: Ensure Responsiveness (RUN LAST)** (AC: 3)
  - [ ] 9.1: Audit all Update() handlers for any blocking operations — ensure none exist
  - [ ] 9.2: Verify flash message timing works correctly (ClearFlashCmd uses 3-second duration)
  - [ ] 9.3: Run `make test` and `make lint` to verify no regressions
  - [ ] 9.4: Final smoke test — all views render correctly with new styling

## Dev Notes

### Architecture Patterns & Constraints

- **Two-layer architecture**: TUI layer (`internal/tui/`) imports domain layer (`internal/tasks/`). NEVER the reverse.
- **Bubbletea Elm Architecture (MVU)**: All state changes through Update(), all rendering through View()
- **No blocking operations in Update()**: Keep all handlers fast — no file I/O in Update()
- **gofumpt formatting**: Run before every commit
- **golangci-lint**: Must pass with zero warnings
- **Random number generation**: Use `math/rand/v2` package — `rand.IntN()` not legacy `math/rand` `rand.Intn()`. Go 1.25.4+ auto-seeds, no manual seed call needed. Import: `"math/rand/v2"`

### Current Source Tree (After Story 1.5)

```
cmd/threedoors/
  main.go                           # Entry point, initializes TaskPool + SessionTracker
  main_test.go                      # Main entry tests
internal/
  tasks/
    task.go                         # Task struct with ID, Text, Status, Blocker, Notes, timestamps
    task_status.go                  # Status constants, IsValidTransition()
    task_pool.go                    # TaskPool — in-memory task collection with ring buffer
    door_selection.go               # SelectDoors() — Fisher-Yates shuffle
    door_selector.go                # DoorSelector interface
    file_manager.go                 # LoadTasks/SaveTasks file I/O
    provider.go                     # TaskProvider interface + implementations
    text_file_provider.go           # TextFileProvider implementation
    session_tracker.go              # SessionTracker — metrics recording (UUID, events, JSON persistence)
    metrics_writer.go               # MetricsWriter — writes session data to sessions.jsonl
    sync_engine.go                  # Sync engine for Apple Notes
    sync_state.go                   # Sync state tracking
    *_test.go                       # Corresponding test files
  tui/
    main_model.go                   # MainModel — root model, view routing, message handling
    doors_view.go                   # DoorsView — three doors rendering with key hints
    detail_view.go                  # DetailView — task details + status action menu
    mood_view.go                    # MoodView — mood capture dialog
    search_view.go                  # SearchView — search + command palette
    messages.go                     # Shared message types between views
    styles.go                       # Lipgloss styles + StatusColor()
    *_test.go                       # Corresponding test files
scripts/
  analyze_sessions.sh               # Session analysis
  validation_decision.sh            # Validation criteria
  daily_completions.sh              # Daily completions
Makefile                            # build, run, clean, fmt, lint, test
```

### Files to MODIFY

| File | Changes |
|------|---------|
| `internal/tui/styles.go` | Add door-specific colors, greeting style, celebration style, separator style; enhance existing styles |
| `internal/tui/doors_view.go` | Apply per-door colors, add startup greeting, vary completion messages |
| `internal/tui/detail_view.go` | Enhance status badge, action menu styling, visual separator |
| `internal/tui/search_view.go` | Polish search prompt colors, result styling |
| `internal/tui/mood_view.go` | Polish mood option styling |
| `internal/tui/main_model.go` | Add greeting message on init, use varied completion messages |

### Files NOT to Touch

- `cmd/threedoors/main.go` — No changes needed
- `internal/tasks/*` — No domain changes, this is purely visual polish
- `Makefile` — No changes needed
- `go.mod` — No new dependencies needed (Lipgloss already available)
- `scripts/*` — No changes needed

### Key Design Decisions

1. **Per-Door Colors**: Assign each door position a distinct color. Use index-based color selection:
```go
var doorColors = []lipgloss.Color{
    lipgloss.Color("86"),  // Door 0 (left) — cyan
    lipgloss.Color("212"), // Door 1 (center) — magenta
    lipgloss.Color("220"), // Door 2 (right) — yellow
}
```
Apply as border foreground on unselected doors. Selected door uses a universal bright border ("255" white).
At narrow terminal widths (< 60 cols), fall back to uniform `colorAccent` for all doors to avoid visual noise.

2. **Greeting Messages Pool**: Store as a package-level `var greetingMessages []string` in `styles.go`. Select randomly using `rand.IntN()` from `math/rand/v2` (Go 1.25.4+ auto-seeds). Store selected greeting as `greeting string` field on `DoorsView` struct, set once in `NewDoorsView()`. Display in `View()` after header, before doors. Track `lastGreetingIdx` to avoid consecutive repeats.

3. **Celebration Messages Pool**: Store as a package-level `var celebrationMessages []string` in `styles.go`. The `TaskCompletedMsg` handler in `main_model.go:136` currently sets a static flash message. Change to `celebrationMessages[rand.IntN(len(celebrationMessages))]`.

6. **Narrow Terminal Guard**: At terminal width < 60 cols, fall back to uniform `colorAccent` for all door borders instead of per-door colors. The width check already exists in `doors_view.go:69-75` — add the color fallback to the same guard.

7. **Task Ordering**: Tasks 1-6 can be done in any order. Task 7 (footer) can be parallel. Task 8 (test updates) should follow Tasks 1-6. Task 9 (responsiveness audit) MUST be last — it validates everything.

4. **No New View States**: This story does NOT add new view modes. It only enhances existing rendering in View() methods and adds message variety. No changes to the Update() flow or message routing.

5. **Minimal Structural Changes**: Prefer modifying existing style variables and View() methods. Avoid creating new files. All changes are within existing TUI layer files.

### Startup Greeting Messages (Sample Pool)

```go
var greetingMessages = []string{
    "Pick one. Start small. That's progress.",
    "Perfection is a trap. Progress is a practice.",
    "Three doors. One choice. Zero wrong answers.",
    "The best task to do is the one you actually start.",
    "You don't need to do it all. Just do one.",
    "Small steps count. Open a door.",
    "Done is better than perfect. Let's go.",
    "Every completed task is a win.",
}
```

### Completion Celebration Messages (Sample Pool)

```go
var celebrationMessages = []string{
    "Progress over perfection. Just pick one and start.",
    "Another one done! You're on a roll.",
    "Task complete! Small wins add up.",
    "Nice work. What's behind the next door?",
    "Done! That's one less thing on your plate.",
    "Crushed it! Keep the momentum going.",
    "One down. Progress feels good, doesn't it?",
    "Completed! You showed up and shipped it.",
    "That's progress! Every task matters.",
    "Well done! The best task is a done task.",
}
```

### Existing Style Constants (in styles.go)

The following are already defined and should be **enhanced, not replaced**:
- `doorStyle` — unselected door border (currently uniform `colorAccent` "63")
- `selectedDoorStyle` — selected door border (currently `colorSelected` "86" with ThickBorder)
- `detailBorder` — detail view border (currently `colorAccent` "63")
- `headerStyle` — bold header (currently `colorAccent` "63")
- `flashStyle` — completion/success messages (currently `colorComplete` "82")
- `helpStyle` — footer help text (currently "241" gray)
- `moodHeaderStyle` — mood dialog header (currently "205" pink)
- `searchResultStyle`, `searchSelectedStyle`, `commandModeStyle` — search styles

### Color Palette Reference

| Use | Color Code | ANSI | Visual |
|-----|-----------|------|--------|
| Door 0 (left) border | "86" | Cyan | Distinctive cool tone |
| Door 1 (center) border | "212" | Magenta/Pink | Warm mid tone |
| Door 2 (right) border | "220" | Yellow | Warm bright tone |
| Selected door border | "255" | White/Bright | Highest contrast |
| Success/Complete | "82" | Green | Positive action |
| Blocked/Error | "196" | Red | Alert |
| In-Progress | "214" | Orange | Active |
| Prompts/Help | "241" | Gray | Subdued |
| Greeting message | "229" | Soft Yellow | Warm, inviting |
| Header/Accent | "63" | Purple | Brand color |

### Previous Story Intelligence

**From Story 1.5 (Session Metrics Tracking):**
- Added `MetricsWriter` for JSON line persistence
- `SessionTracker` now has UUID, tracks detail views, status changes, refreshes
- All tracking is non-blocking (<1ms overhead)
- Session metrics persist on app exit

**From Story 1.4 (Search & Command Palette):**
- Added `SearchView` with live filtering, command mode (:add, :mood, :stats, :help, :quit)
- Context-aware Esc returns (search → detail → search preserved)
- Uses `textinput` from bubbles library

**From Story 1.3 (Door Selection & Task Status):**
- `DetailView` with status action menu (c/b/i/e/f/p/r/m keys)
- `MoodView` with predefined options + custom text
- Flash message system with timed auto-clear
- View state routing pattern in MainModel

**Key patterns established:**
- `flashStyle.Render()` for temporary messages
- `helpStyle.Render()` for footer text
- `headerStyle.Render()` for section headers
- `StatusColor()` for status-specific coloring
- `ClearFlashCmd()` for timed message clearing (3 seconds)

### Testing Requirements

**New test cases needed** (add to existing test files, no new test files):

#### `doors_view_test.go` — New Tests:
- `TestDoorsView_GreetingDisplayed` — verify `View()` output contains a non-empty greeting from `greetingMessages` pool
- `TestDoorsView_GreetingIsFromPool` — verify the greeting field is one of the known `greetingMessages` entries
- `TestDoorsView_PerDoorColors_WideTerminal` — verify at width >= 60, View() renders doors (visual check: output is non-empty, doors render)
- `TestDoorsView_NarrowTerminal_Fallback` — verify at width < 60, doors still render without error

#### `main_model_test.go` — Updates:
- `TestFlashMessage_ShowsAfterCompletion` (line 304) — currently checks for exact `"Progress over perfection"`. After change, flash will be a random celebration message. Update to: check that `m.flash` is non-empty and is one of `celebrationMessages` entries, OR check `view` contains at least one of the celebration messages.

#### `styles_test.go` — New Tests:
- `TestGreetingMessages_NonEmpty` — verify `len(greetingMessages) > 0`
- `TestCelebrationMessages_NonEmpty` — verify `len(celebrationMessages) > 0`
- `TestGreetingMessages_NoDuplicates` — verify no duplicate entries
- `TestCelebrationMessages_NoDuplicates` — verify no duplicate entries

#### Existing Tests That Should NOT Break:
- `TestDoorsView_View_RendersHeader` — checks `"ThreeDoors"`, still valid
- `TestDoorsView_View_RendersHelp` — checks `"quit"`, still valid
- `TestDoorsView_View_EmptyPool_ShowsAllDone` — checks `"All tasks done"`, still valid
- `TestDoorsView_View_CompletionCounter_Visible` — checks `"Completed this session: 1"`, still valid
- `TestDoorsView_View_ProgressMessage` — checks `"Progress over perfection"` in footer tagline. This check remains valid because the footer tagline still contains "Progress over perfection" (line 101 of `doors_view.go`).

### BDD Acceptance Test Scenarios

```gherkin
# AC 1: Per-door distinct colors
Given tasks are loaded and terminal width >= 60
When the doors view renders
Then each door has a distinct border color (cyan, magenta, yellow)

# AC 1: Narrow terminal fallback
Given tasks are loaded and terminal width < 60
When the doors view renders
Then all doors use the uniform accent color (no per-door coloring)

# AC 2: Startup greeting
Given the application starts
When the doors view renders for the first time
Then the output contains a greeting message from the greetingMessages pool
And the greeting appears between the header and the doors

# AC 2: Non-consecutive greeting
Given the application is restarted
When a new DoorsView is created
Then the selected greeting is from the pool
And the greeting is not guaranteed to be the same as last time (random)

# AC 2: Varied completion messages
Given a task is completed
When the completion flash message appears
Then the message is one of the celebrationMessages pool entries
And the message is styled with flashStyle (green, bold)

# AC 1: Detail view polish
Given the detail view is open
When it renders
Then status badge is color-coded (green/red/yellow per status)
And action menu keys have distinct styling
And a visual separator exists between content and action menu

# AC 3: Responsiveness
Given the user performs any action (door select, enter, esc, status change)
When the action completes
Then the response is immediate with no observable lag
```

### View Output Assertions Table (Polished Views)

| View State | Must Contain | Must NOT Contain |
|-----------|-------------|-----------------|
| viewDoors (normal, width >= 60) | "ThreeDoors", task texts, greeting message, "quit", "Progress over perfection" | "Error", "All tasks done" |
| viewDoors (normal, width < 60) | "ThreeDoors", task texts, greeting message, "quit" | "Error", "All tasks done" |
| viewDoors (all done) | "All tasks done" | task texts |
| viewDoors (with counter) | "Completed this session:" | - |
| viewDoors (with flash) | one of celebrationMessages entries | - |
| viewDetail | full task text, "[C]omplete", "[B]locked", "[Esc]Back" | "ThreeDoors" header |
| viewMood | "How are you feeling", mood options | "[C]omplete" |
| viewSearch | search prompt, results if matching | - |

### Testability Accessor

Add `Greeting() string` method to `DoorsView`:
```go
// Greeting returns the current startup greeting message.
func (dv *DoorsView) Greeting() string {
    return dv.greeting
}
```
This enables test verification without parsing View() output.

### Definition of Done

- All tests pass: `make test` green
- Code formatted: `make fmt` produces no changes
- Lint clean: `make lint` passes with zero warnings
- Each door has a distinct visual color
- Startup shows a "progress over perfection" greeting
- Task completion shows varied celebration messages
- All views have consistent, polished styling
- No noticeable UI lag on any interaction

### Project Structure Notes

- All changes are in `internal/tui/` — no domain layer changes
- No new files needed — all modifications to existing view files and styles.go
- Maintains alignment with `docs/architecture/source-tree.md`

### References

- [Source: docs/prd/epics-and-stories.md#Story 1.6] — Acceptance criteria
- [Source: docs/architecture/coding-standards.md] — Naming, formatting, linting rules
- [Source: docs/architecture/test-strategy-and-standards.md] — Testing philosophy
- [Source: internal/tui/styles.go] — Current style definitions
- [Source: internal/tui/doors_view.go] — Current door rendering
- [Source: internal/tui/detail_view.go] — Current detail view rendering
- [Source: internal/tui/main_model.go] — Flash message and completion handling

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (Amelia/Dev Agent)

### Debug Log References

N/A — all tests passed on first implementation.

### Completion Notes List

- Added `doorColors` (3 distinct colors: cyan, magenta, yellow) for per-door border styling
- Added `greetingMessages` pool (8 entries) and `celebrationMessages` pool (10 entries) as package-level vars in `styles.go`
- Added `greetingStyle` (color "229" soft yellow) and `separatorStyle` (color "238" dark gray) to styles
- Changed `selectedDoorStyle` border color from "86" to "255" (bright white) for higher contrast
- Added `greeting` field to `DoorsView`, set once in `NewDoorsView()` via `rand.IntN()` from `math/rand/v2`
- Added `Greeting()` accessor method on `DoorsView` for testability
- Per-door colors only applied at terminal width >= 60 (narrow terminal fallback)
- Greeting displayed between header and doors in `View()`
- Footer tagline styled with `greetingStyle` instead of `helpStyle`
- Celebration messages randomized from pool in `TaskCompletedMsg` handler
- Visual separator added between task content and action menu in detail view
- Updated existing `TestFlashMessage_ShowsAfterCompletion` to validate against celebration pool
- All 10 new tests pass (styles, doors_view, main_model)
- [Review Fix] Added `pickGreeting()` helper with non-consecutive guard (M1)
- [Review Note] Tasks 5/6 (search/mood view polish) deferred — existing styling is adequate for "Essential Polish" scope
- [Review Note] Task 1.3 (per-door key hints) deferred — existing help line suffices
- [Review Note] Task 4.1-4.2 (action key styling) — status badge already color-coded via StatusColor(), separator added

### File List

- `internal/tui/styles.go` — Added doorColors, greetingStyle, separatorStyle, greetingMessages, celebrationMessages, colorGreeting, colorDoorBright
- `internal/tui/styles_test.go` — Added 9 new tests for message pools, door colors, greeting style
- `internal/tui/doors_view.go` — Added greeting field, Greeting() accessor, per-door colors, greeting display, narrow terminal guard
- `internal/tui/doors_view_test.go` — Added 5 new tests for greeting display, pool validation, persistence, narrow terminal
- `internal/tui/detail_view.go` — Added visual separator between content and action menu
- `internal/tui/main_model.go` — Changed celebration message to random selection from pool, added math/rand/v2 import
- `internal/tui/main_model_test.go` — Updated completion flash test to validate against celebration pool
