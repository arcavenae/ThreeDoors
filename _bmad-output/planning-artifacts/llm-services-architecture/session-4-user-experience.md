# Party Mode Session 4: User Experience — LLM Service Interactions

**Date:** 2026-03-11
**Participants:** UX Designer, PM, Dev, Architect, TEA/QA
**Topic:** How does the user interact with LLM services? What feels magical vs intrusive?

---

## Round 1: The Interaction Spectrum

### UX Designer (Opening)

SOUL.md gives us clear guidance: "Opening ThreeDoors should feel like a friend saying: 'Hey, here are three things you could do right now. Pick one. Any one. Let's go.'"

LLM services must extend this feeling, not disrupt it. The friend doesn't say "I've analyzed your productivity patterns and recommend task #47 based on your circadian rhythm." The friend says "This one's quick — want to knock it out?"

I see three interaction modes, from most to least user-initiated:

1. **Explicit commands** — User types `:extract`, `:enrich`, `:breakdown`
2. **Contextual suggestions** — ThreeDoors offers when appropriate (e.g., viewing a large task: "Want me to break this down?")
3. **Ambient enrichment** — Background processing that enriches data without interruption

### PM

SOUL.md says user-initiated. So:
- Mode 1 (explicit commands): **Always OK**
- Mode 2 (contextual suggestions): **OK if subtle and dismissible**
- Mode 3 (ambient enrichment): **Risky — must be opt-in and invisible**

Let's design each.

---

## Round 2: Explicit Commands

### Dev

Here's the command palette integration. ThreeDoors already has `:command` mode. We add:

| Command | Action | Context |
|---------|--------|---------|
| `:extract` | Extract tasks from text/file/clipboard | Global |
| `:enrich` | Enrich current task with LLM | Task detail view |
| `:breakdown` | Break current task into subtasks | Task detail view |
| `:llm-status` | Show LLM backend status | Global |

And CLI equivalents:

```bash
threedoors extract --file notes.txt
threedoors extract --clipboard
threedoors enrich <task-id>
threedoors breakdown <task-id>
threedoors llm status
```

### UX Designer

For TUI commands, the flow should feel snappy and predictable:

**`:extract` flow:**
```
User types :extract
→ Prompt: "Source? [f]ile [c]lipboard [p]aste"
→ User picks source
→ Spinner: "Extracting tasks..."  (1-10 seconds)
→ Review screen with extracted tasks
→ User selects and imports
→ Flash message: "Imported 3 tasks"
→ Return to doors (new tasks may appear)
```

**`:enrich` flow (in task detail view):**
```
User types :enrich while viewing a task
→ Spinner: "Enriching task..."  (2-5 seconds)
→ Show enriched version alongside original:

  Original: "Fix the login bug"
  Enriched:
    Text: "Fix the login bug — session timeout not renewing"
    Tags: bug, auth, backend
    Effort: 2
    Context: "Users report being logged out after 30 min despite
              'remember me' checkbox. Likely session cookie expiry."

→ Prompt: "[a]ccept [e]dit [d]iscard"
→ If accept: update task in place
→ If edit: open editor with enriched version
→ If discard: return to original
```

**`:breakdown` flow (in task detail view):**
```
User types :breakdown while viewing a task
→ Spinner: "Breaking down task..."  (3-10 seconds)
→ Show proposed subtasks:

  Breaking down: "Redesign the settings page"

  Proposed subtasks:
  1. ☐ Audit current settings page for unused options  [~1]
  2. ☐ Design new settings layout in Figma             [~2]
  3. ☐ Implement settings form with validation          [~3]
  4. ☐ Add settings persistence to config.yaml          [~2]
  5. ☐ Write tests for new settings page                [~2]

→ Prompt: "[Space] toggle [E] edit [Enter] create selected [Esc] cancel"
→ Selected subtasks become new tasks linked to parent
```

### PM

I love the enrichment UX. The side-by-side comparison lets the user see exactly what the LLM added. No black box magic — full transparency.

One thing: the enrichment should **never overwrite** without showing the diff. The user must see what changed.

---

## Round 3: Contextual Suggestions

### UX Designer

Contextual suggestions are the tricky part. They should feel like gentle nudges, not nagging. Here's my proposal:

**When viewing a large/vague task in detail view:**
```
┌─────────────────────────────────────────────────────┐
│ TASK DETAILS                                        │
├─────────────────────────────────────────────────────┤
│                                                     │
│ Redesign the entire frontend                        │
│                                                     │
│ Status: TODO                                        │
│ Effort: 5                                           │
│                                                     │
│ ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄ │
│ This task seems large. Break it down? (B)           │
│ ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄ │
└─────────────────────────────────────────────────────┘
```

**Rules for suggestions:**
1. **Only in detail view** — never interrupt the doors view
2. **One suggestion at a time** — never stack multiple suggestions
3. **Dismissible and non-blocking** — press any key to dismiss
4. **No repeat** — once dismissed for a task, don't suggest again (persisted per task)
5. **Heuristic trigger** — only suggest breakdown when effort ≥ 4 or text is vague (< 10 words, no verbs)
6. **LLM not called yet** — the suggestion is heuristic. LLM is only called if user accepts.

### Architect

The heuristic trigger is smart. We're NOT calling the LLM to decide whether to suggest — we're using simple rules:

```go
func shouldSuggestBreakdown(task *Task) bool {
    if task.Effort >= 4 {
        return true
    }
    words := strings.Fields(task.Text)
    if len(words) < 5 {
        return true  // very vague
    }
    return false
}
```

The LLM is only invoked when the user acts on the suggestion. This keeps the doors view fast and free of latency.

### TEA/QA

What about suggesting enrichment? A task like "bug" or "thing to do" is too vague to be useful. We could suggest:

```
This task could use more detail. Enrich it? (E)
```

Same rules: heuristic trigger (text < 3 words), non-blocking, dismissible, no repeat.

---

## Round 4: Ambient Enrichment (Background)

### Architect

Ambient enrichment is the most controversial. The idea: when ThreeDoors is idle (user hasn't pressed a key for 30+ seconds), background-enrich tasks that lack tags or effort estimates.

**I'm against this for MVP.** Here's why:

1. **Privacy concern** — silently sending task text to cloud LLMs violates SOUL.md
2. **Cost concern** — background processing burns API credits without user awareness
3. **Latency concern** — CLI process startup for each task in background
4. **Battery concern** — constant subprocess spawning drains laptop battery
5. **Correctness concern** — LLM enrichments without review may be wrong

### PM

Agreed. Ambient enrichment is a P2 feature at most, and only for local backends (Ollama). For cloud backends, it should require explicit opt-in with a consent dialog:

```
"ThreeDoors can automatically enrich your tasks with tags and effort
estimates using claude. This sends task text to Anthropic's servers.

Enable background enrichment? [y/N]"
```

### UX Designer

If we ever build ambient enrichment, it should be **completely invisible** in the UI. No spinners, no notifications, no status bar changes. The user opens a task and sees tags are already there. It feels like ThreeDoors "just knows."

But that's a future polish item. For now, explicit commands only.

---

## Round 5: The Magical Moments

### UX Designer

Let me paint the picture of what "magical" looks like with LLM services:

**Moment 1: The Meeting Brain Dump**
User just got out of a meeting. Opens ThreeDoors, types `:extract`, pastes their messy notes. 10 seconds later, 6 clean actionable tasks appear. User taps Space-Space-Space-Enter, and they're in the pool. Next time doors appear, one of them is "Email Sarah about Q2 budget." The user smiles.

**Moment 2: The Overwhelming Task**
User stares at "Redesign the entire frontend." They've been re-rolling past it for days. Today they press B for breakdown. 8 seconds later, 5 bite-sized subtasks appear. User picks 3, imports them. Next doors refresh — "Audit current settings page for unused options" appears behind a door. Suddenly it feels doable.

**Moment 3: The Quick Enrichment**
User has a task that just says "taxes." They press E for enrich. Claude fills in: "Gather 2025 tax documents (W-2, 1099s, receipts) and schedule appointment with accountant. Due: April 15." The task goes from dread to actionable.

**Moment 4: The Clipboard Import**
User copies a list from a Slack message. Types `:extract`, selects clipboard. 5 tasks appear. Some are already in their pool (duplicate detection catches them). User imports the 2 new ones.

### PM

These moments all share the same structure:
1. User has a moment of friction
2. User takes ONE action (command or keybind)
3. LLM does work in seconds
4. User reviews and confirms
5. Friction is resolved

That's the ThreeDoors philosophy: **reduce friction to starting**. The LLM doesn't manage tasks — it removes barriers to creating and understanding them.

### Dev

One technical note on the "magical" feel: **latency matters enormously**. If the LLM takes 30 seconds, the magic dies. The spinner becomes a wall.

Latency targets:
- Extraction: < 5 seconds for short text, < 15 seconds for long documents
- Enrichment: < 3 seconds per task
- Breakdown: < 8 seconds per task

For Ollama (local), these are achievable with fast models (llama3.2, phi-3). For Claude CLI, network latency adds 1-3 seconds. For Gemini CLI, similar.

If latency exceeds targets, we should:
1. Show estimated time in spinner: "Extracting tasks... (~10s)"
2. Allow cancel during processing (Esc)
3. Consider pre-warming: when user enters detail view, silently start loading enrichment data (speculative prefetch — but ONLY for local backends)

---

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| S4-D1 | Three interaction modes: explicit, contextual, ambient | Explicit commands for MVP; contextual suggestions for P1; ambient for P2 (local-only) | All three at once | Incremental rollout reduces risk; explicit-first builds trust |
| S4-D2 | TUI commands: `:extract`, `:enrich`, `:breakdown` | Command palette + keybinds in detail view | Menu-based navigation | Consistent with existing ThreeDoors command model |
| S4-D3 | Always show before/after for enrichment | Side-by-side diff with accept/edit/discard | Silent in-place update | Transparency builds trust; prevents unwanted changes |
| S4-D4 | Contextual suggestions: heuristic-triggered, no LLM call | Simple rules (effort ≥ 4, short text) trigger suggestion line | LLM-powered suggestion engine | Fast, private, no API cost for suggestions |
| S4-D5 | Ambient enrichment deferred to P2, local-only | Opt-in, invisible, Ollama only | Available at launch; cloud backends | Privacy, cost, battery, correctness concerns |
| S4-D6 | Cancel support during all LLM operations | Esc cancels in-progress LLM calls | No cancel / background processing | Latency can be unpredictable; user must be able to bail |
| S4-D7 | Latency targets: extract <5s, enrich <3s, breakdown <8s | Targets inform backend selection and UX feedback | No latency requirements | Slow LLM responses kill the "magical" feeling |
| S4-D8 | Review-then-import for all LLM outputs | User explicitly confirms before data changes | Auto-apply with undo | SOUL.md: user control; LLMs hallucinate |
