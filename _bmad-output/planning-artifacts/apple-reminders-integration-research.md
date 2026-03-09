# Apple Reminders Integration Research

## Executive Summary

This document evaluates approaches for integrating Apple Reminders with ThreeDoors as a new `TaskProvider` adapter. Unlike Apple Notes (which stores free-form text requiring line-by-line parsing), Reminders is a structured task manager with native fields for title, notes, due date, priority, and completion status — making it a natural fit for ThreeDoors' task model.

**Recommended approach:** JXA (JavaScript for Automation) via `osascript` for the initial implementation, with an optional future migration path to EventKit via cgo for performance. This mirrors the Apple Notes adapter's architecture (osascript + `CommandExecutor` interface) while leveraging JXA's native JSON output to avoid brittle string parsing.

## Comparison with Apple Notes Adapter

| Aspect | Apple Notes | Apple Reminders |
|--------|------------|-----------------|
| Data model | Free-form text (parsed line-by-line) | Structured fields (title, notes, due, priority) |
| Task identity | Position-based SHA-1 (fragile) | Native persistent ID from Reminders.app |
| Completion tracking | Markdown checkbox `- [x]` | Native `completed` boolean field |
| Priority | Not supported | Native 0-9 scale |
| Due dates | Not supported | Native `due date` field |
| Multiple sources | Single note | Multiple reminder lists |
| Access method | `osascript` (AppleScript) | `osascript` (JXA recommended) |
| TCC permission | Automation (Notes.app) | Reminders (`kTCCServiceReminders`) |
| iCloud sync | Via Notes.app | Via `remindd` daemon |
| Write support | Read-modify-write of note body | Direct per-reminder CRUD |

**Key lesson from Apple Notes:** The `CommandExecutor` interface pattern enables full mocking in unit tests and CI portability. The Reminders adapter should reuse this pattern.

**Key improvement over Apple Notes:** Reminders provides stable, persistent IDs per reminder — no more position-based ID fragility. Each reminder has a system-assigned `id` property that survives reordering, edits, and sync.

## Access Methods Evaluated

### 1. JXA via osascript (Recommended)

JavaScript for Automation called through `osascript -l JavaScript`.

**Advantages:**
- Native `JSON.stringify()` output — no brittle text parsing
- ISO 8601 date formatting via `toISOString()`
- Same `CommandExecutor` pattern as Apple Notes adapter
- No build dependencies (osascript is built-in)
- Cross-platform unit tests via mocking

**Disadvantages:**
- ~500ms latency per invocation (subprocess overhead)
- Apple has not invested in JXA since ~2016 (semi-maintained)
- Some edge cases with `whose` clause filtering

**Example — Read incomplete reminders as JSON:**

```javascript
const app = Application("Reminders");
const list = app.lists.byName("Work");
const reminders = list.reminders.whose({completed: false})();
JSON.stringify(reminders.map(r => ({
    id: r.id(),
    title: r.name(),
    notes: r.body(),
    dueDate: r.dueDate() ? r.dueDate().toISOString() : null,
    priority: r.priority(),
    completed: r.completed(),
    flagged: r.flagged(),
    creationDate: r.creationDate().toISOString(),
    modificationDate: r.modificationDate().toISOString()
})));
```

**Example — Complete a reminder by ID:**

```javascript
const app = Application("Reminders");
const r = app.reminders.byId("x-apple-reminder://...");
r.completed = true;
```

**Example — Create a reminder:**

```javascript
const app = Application("Reminders");
const list = app.lists.byName("Work");
list.reminders.push(app.Reminder({
    name: "Task title",
    body: "Notes here",
    priority: 5
}));
```

### 2. AppleScript via osascript

Traditional AppleScript called through `osascript -e`.

**Advantages:**
- Widely documented, many examples available
- Same approach as the existing Apple Notes adapter

**Disadvantages:**
- Output is locale-dependent (dates formatted per system locale)
- Requires fragile text parsing for multi-field output
- Multi-line body text requires careful escaping
- No native JSON output

**Verdict:** Not recommended. JXA provides the same access with structured JSON output, eliminating the parsing issues that complicated the Apple Notes adapter.

### 3. EventKit via cgo

Native macOS framework accessed through Objective-C bridge code compiled with cgo.

**Advantages:**
- Sub-millisecond reads (vs ~500ms for osascript)
- Complete API access (subtasks, recurrence, URLs, attachments)
- Real-time change notifications (`EKEventStoreChangedNotification`)
- Proven in Go: [go-eventkit](https://github.com/BRO3886/go-eventkit) library

**Disadvantages:**
- Requires CGO_ENABLED=1 and Xcode Command Line Tools
- Cannot cross-compile (darwin-only Objective-C compilation)
- Significantly more complex build pipeline
- Binary tied to macOS SDK version
- Harder to test (cgo in test builds)

**Verdict:** Good future optimization if osascript latency becomes a bottleneck. Not recommended for initial implementation due to build complexity. Could be offered as an optional build tag (`//go:build cgo && darwin`).

### 4. Shortcuts CLI

macOS `shortcuts run` command to invoke pre-built Shortcuts.

**Verdict:** Not viable as a primary integration. Requires users to manually create shortcuts in the GUI. No structured input/output. Only useful as a supplementary user-customization hook.

### 5. Third-Party Go Library (go-eventkit)

[github.com/BRO3886/go-eventkit](https://github.com/BRO3886/go-eventkit) wraps EventKit in a Go-friendly API.

**Verdict:** Promising for a future cgo-based adapter. Same trade-offs as EventKit above, but reduces implementation effort. Worth monitoring for maturity.

## Proposed Adapter Design

### Package Structure

```
internal/adapters/reminders/
├── reminders_provider.go       # TaskProvider implementation
├── reminders_provider_test.go  # Unit tests with mock executor
├── jxa_scripts.go              # JXA script templates
├── field_mapping.go            # Reminder ↔ Task field conversion
├── errors.go                   # Error categorization (reuse pattern from applenotes)
└── retry.go                    # Retry config (reuse pattern from applenotes)
```

### Field Mapping: Reminder → Task

| Reminder Field | Task Field | Mapping |
|---------------|-----------|---------|
| `id` | `ID` | Use directly (stable, persistent) |
| `name` | `Text` | Direct mapping |
| `body` | `Notes[0].Text` | Map to first TaskNote |
| `completed` | `Status` | `true` → `StatusComplete`, `false` → `StatusTodo` |
| `priority` | Custom field or Context | 1-4=high, 5=medium, 6-9=low, 0=none |
| `dueDate` | Could extend Task model or use Context | See design decision below |
| `flagged` | Could map to effort or type | Optional |
| `creationDate` | `CreatedAt` | Direct mapping |
| `modificationDate` | `UpdatedAt` | Direct mapping |
| `completionDate` | `CompletedAt` | Direct mapping |
| list name | `SourceProvider` | `"reminders:<list-name>"` |

### Design Decisions

**1. Due dates and priority — extend Task or encode in Context?**

Option A: Encode in `Context` field (e.g., `"due:2026-03-10 priority:high"`)
- Pro: No model changes, works with existing TUI
- Con: Lossy, not machine-parseable for round-trip

Option B: Add `DueDate` and `Priority` fields to `core.Task`
- Pro: First-class support, sortable, filterable
- Con: Requires model changes, affects all adapters

**Recommendation:** Option A for initial implementation (no model changes needed). Option B as a future enhancement when multiple adapters benefit from due dates.

**2. ID strategy**

Unlike Apple Notes (position-based IDs), Reminders provides stable, persistent IDs (`x-apple-reminder://...` URIs). The adapter should use these directly as `Task.ID`, avoiding the fragility of the Notes adapter.

**3. List selection**

The adapter should accept a list of Reminders list names in its config. Default behavior: read from all lists. Users can filter to specific lists (e.g., "Work", "ThreeDoors").

```go
type Config struct {
    Lists          []string      // empty = all lists
    Timeout        time.Duration
    Retry          RetryConfig
    Logger         LogFunc
    IncludeCompleted bool        // default false
}
```

**4. Write-back support**

The adapter should support full CRUD:
- `SaveTask` → create or update reminder (by ID)
- `DeleteTask` → delete reminder
- `MarkComplete` → set `completed = true`

This is a significant improvement over the Apple Notes adapter, which returns `ErrReadOnly` for `MarkComplete`.

**5. Watch support**

Initial implementation: `Watch()` returns `nil` (no watch support), same as Apple Notes.

Future enhancement: Poll-based watcher that periodically re-reads reminders and emits `ChangeEvent` for detected changes. EventKit's `EKEventStoreChangedNotification` would enable true push-based watching but requires cgo.

### Constructor

```go
func NewRemindersProvider(listNames []string, opts ...Option) *RemindersProvider
```

Following the project convention: accept interfaces (CommandExecutor), return concrete type.

### Contract Tests

The Reminders adapter should support the full `RunContractTests` suite, unlike the Apple Notes adapter which uses a subset due to position-based ID limitations. Stable reminder IDs enable full ID round-trip testing.

Integration tests (requiring actual Reminders.app access) should be guarded with a build tag:

```go
//go:build integration && darwin
```

## Privacy & Permissions

### TCC Behavior

- First `osascript` call targeting Reminders triggers a macOS consent dialog
- The dialog attributes the request to the parent app (Terminal.app, iTerm2, etc.)
- Once granted, all scripts from that terminal have access
- Permission managed in System Settings > Privacy & Security > Reminders

### For Distributed Binary

- Unsigned binary: TCC tracks by absolute file path; moving/rebuilding may re-prompt
- Signed binary: TCC tracks by code signature; stable across updates
- Recommendation: document the TCC prompt in user-facing docs; code-sign for Homebrew distribution

### Checking Permission

The adapter's `HealthCheck()` should attempt a lightweight Reminders read and report the result:
- `HealthOK` if reminders are accessible
- `HealthFail` with suggestion to grant Reminders access in System Settings

## iCloud Sync Considerations

- Changes made via EventKit or AppleScript are immediately visible locally
- iCloud propagation to other devices: typically 1-10 seconds, occasionally minutes
- Conflict resolution: last-writer-wins at the field level (iCloud/CloudKit semantics)
- No API to force sync or check sync status
- ThreeDoors does not need to manage sync — the OS handles it transparently
- Completing a task in ThreeDoors will eventually sync to iPhone/iPad Reminders

## Performance Expectations

Based on Apple Notes adapter benchmarks and Reminders-specific characteristics:

| Operation | Expected Latency | Notes |
|-----------|-----------------|-------|
| Read all reminders (1 list, <50 items) | ~300-500ms | Single osascript invocation |
| Read all reminders (5 lists, <200 items) | ~500-800ms | May need multiple invocations |
| Complete a reminder | ~200-400ms | Single osascript invocation |
| Create a reminder | ~200-400ms | Single osascript invocation |
| Delete a reminder | ~200-400ms | Single osascript invocation |

All within the project's NFR6 budget (<500ms for individual operations). Batch reads of large lists may exceed this; pagination or background loading may be needed.

## Implementation Roadmap

### Phase 1: Read-Only Adapter
- JXA scripts to list reminder lists and read reminders
- Field mapping to `core.Task`
- `CommandExecutor` interface for testability
- Unit tests with mocked executor
- `HealthCheck()` implementation
- Config for list filtering

### Phase 2: Write Support
- `SaveTask` (create/update reminders)
- `MarkComplete` (set completed flag)
- `DeleteTask` (remove reminder)
- Error categorization and retry logic
- Contract test compliance

### Phase 3: Watch Support (Optional)
- Poll-based change detection
- Emit `ChangeEvent` on detected changes

### Phase 4: EventKit Migration (Optional)
- cgo-based adapter behind build tag
- Sub-millisecond reads
- Real-time change notifications

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| JXA deprecated by Apple | Medium | JXA runs on current macOS; if removed, migrate to EventKit or Shortcuts |
| TCC permission denied | Low | Clear error message + HealthCheck guidance |
| Large reminder lists (>500 items) | Medium | Paginate reads, filter by list, exclude completed |
| iCloud sync conflicts | Low | Last-writer-wins is acceptable; document behavior |
| osascript latency spikes | Low | Retry with backoff (same as Apple Notes adapter) |
| Reminders.app not running | Low | osascript launches it automatically |

## References

- [Apple EventKit Documentation](https://developer.apple.com/documentation/eventkit)
- [go-eventkit — Go EventKit bindings](https://github.com/BRO3886/go-eventkit)
- [rem — Go CLI for macOS Reminders](https://github.com/BRO3886/rem)
- [remindctl — Swift CLI for Reminders](https://github.com/steipete/remindctl)
- ThreeDoors Apple Notes spike report: `docs/spike-reports/2.2-apple-notes-integration.md`
- ThreeDoors adapter developer guide: `docs/adapter-developer-guide.md`
