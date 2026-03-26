# Next Steps

## Immediate Next Steps: Begin Technical Demo Implementation

**Objective:** Implement Epic 1 (Three Doors Technical Demo) with expanded scope to validate full task management workflow.

**Starting Point:** Story 1.1 - Project Setup & Basic Bubbletea App

## Updated Epic 1 Story Breakdown

### Story 1.1: Project Setup & Basic Bubbletea App
**Estimated Time:** 30-45 minutes

**Acceptance Criteria:**
1. Go module initialized
2. Dependencies added (Bubbletea, Lipgloss, Bubbles, yaml.v3, uuid)
3. Basic TUI renders header
4. 'q' quits application
5. Directory structure created
6. Makefile works
7. Compiles without errors

### Story 1.2: Data Models & YAML Schema
**Estimated Time:** 45-60 minutes

**Acceptance Criteria:**
1. Task model with all fields
2. TaskStatus enum defined
3. TaskNote model
4. Config model
5. Validation methods
6. Status transition methods
7. YAML tags
8. Unit tests (70%+ coverage)

### Story 1.3: File I/O & YAML Parsing
**Estimated Time:** 60-75 minutes

**Acceptance Criteria:**
1. FileManager component
2. Creates `~/.threedoors/` directory
3. LoadTasks() reads YAML
4. SaveTasks() uses atomic writes
5. InitializeFiles() creates samples
6. Handles corrupted files
7. Integration tests

### Story 1.4: Task Pool & Door Selection
**Estimated Time:** 45-60 minutes

**Acceptance Criteria:**
1. TaskPool implemented
2. GetAvailableForDoors() filters correctly
3. Recently-shown ring buffer
4. DoorSelector random selection
5. Handles edge cases (0-100 tasks)
6. Unit tests

### Story 1.5: Three Doors Display (DoorsView)
**Estimated Time:** 60-75 minutes

**Acceptance Criteria:**
1. DoorsView component
2. Three boxes with Lipgloss
3. Status indicators (emoji)
4. Truncation with "..."
5. Handles 0-2 tasks
6. Keyboard handlers
7. "Progress over perfection" message

### Story 1.6: Door Refresh Mechanism
**Estimated Time:** 15-20 minutes

**Acceptance Criteria:**
1. R key generates new doors
2. Excludes recently shown
3. Random variety
4. Edge case handling

### Story 1.7: Task Detail View
**Estimated Time:** 90-120 minutes

**Acceptance Criteria:**
1. TaskDetailView component
2. Full task display
3. Status with emoji
4. Timestamps (local timezone)
5. Notes list
6. Blocker display
7. Options menu
8. View modes
9. ESC returns to doors

### Story 1.8: Status Update Menu
**Estimated Time:** 45-60 minutes

**Acceptance Criteria:**
1. StatusUpdateMenu component
2. All 5 status options
3. Current status checkmark
4. Invalid transitions grayed
5. Validation with StatusManager
6. Error handling
7. ESC cancels

### Story 1.9: Notes Input
**Estimated Time:** 45-60 minutes

**Acceptance Criteria:**
1. NotesInputView with Bubbles textarea
2. Multi-line input
3. Character counter
4. Ctrl+S saves
5. ESC cancels
6. Timestamp saved
7. YAML updated

### Story 1.10: Blocker Input
**Estimated Time:** 30-45 minutes

**Acceptance Criteria:**
1. BlockerInputView component
2. Single-line input
3. Enter saves
4. Sets status=blocked
5. YAML updated

### Story 1.11: Task Completion Flow
**Estimated Time:** 45-60 minutes

**Acceptance Criteria:**
1. Complete from status menu
2. CompletedAt timestamp
3. Append to completed.txt
4. Remove from pool
5. Auto-refresh doors
6. Session count increments

### Story 1.12: Integration & Polish
**Estimated Time:** 60-90 minutes

**Acceptance Criteria:**
1. View transitions work
2. Message routing correct
3. Consistent styling
4. Color scheme applied
5. Terminal responsive
6. Error handling tested
7. Edge cases tested
8. README.md
9. Linting passes
10. Formatting applied

## Updated Time Estimate

**Total Epic 1 Time:** 8-12 hours (was 3-6 hours)

**Validation Period:** 1 week of daily use after implementation

## Development Approach

**Sequential Execution:**
1. Stories 1.1-1.4: Foundation (4-5 hours)
2. Stories 1.5-1.6: Basic UI (1.5-2 hours)
3. **Checkpoint:** Working simplified version
4. Stories 1.7-1.11: Full features (4-5 hours)
5. Story 1.12: Polish (1-1.5 hours)

## Post-Validation Decision Gate

**Success Criteria:**
- ✅ Three Doors reduces friction
- ✅ Status management adds value
- ✅ Notes/progress tracking useful
- ✅ Daily use without annoyance

**If Successful → Epic 2:**
- Apple Notes integration spike
- TaskProvider interface refactor
- Bidirectional sync

**If Unsuccessful → Pivot:**
- Document learnings
- Reassess problem

## Quick Start Commands

```bash
cd /Users/michael.pursifull/work/simple-todo

# Initialize module
go mod init github.com/arcavenae/ThreeDoors

# Add dependencies
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get gopkg.in/yaml.v3@latest
go get github.com/google/uuid@latest
go get golang.org/x/term@latest

# Create directories
mkdir -p cmd/threedoors internal/tui internal/tasks

# Install tools
go install mvdan.cc/gofumpt@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Reference Resources:**
- Bubbletea: https://github.com/charmbracelet/bubbletea/tree/master/tutorials
- Lipgloss: https://github.com/charmbracelet/lipgloss/tree/master/examples
- Bubbles: https://github.com/charmbracelet/bubbles
- YAML v3: https://pkg.go.dev/gopkg.in/yaml.v3

---

**Document Complete**

Generated using BMAD-METHOD™ framework
Architect: Winston
Session Date: 2025-11-07
