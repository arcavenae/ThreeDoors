# Party Mode Session 3: Source Integration — Task Extraction Pipeline

**Date:** 2026-03-11
**Participants:** Dev, Architect, UX Designer, PM, TEA/QA
**Topic:** How do we extract tasks from Apple Notes, Obsidian, transcripts, and arbitrary text? What's the pipeline from raw text to ThreeDoors task?

---

## Round 1: The Extraction Problem

### Dev (Opening)

We already have adapters that read from Apple Notes and Obsidian. But they use **pattern matching** — looking for checkbox syntax (`- [ ]`), specific note structures, etc. The LLM extraction service solves a different problem: understanding **natural language intent**.

Consider this Apple Note:

```
Meeting with Sarah — March 10

Need to follow up on the Q2 budget proposal.
Ask Mike about the server migration timeline.
Remember to update the design docs before Friday.
The client demo is next Tuesday — need to prep slides.
```

Our Apple Notes adapter sees zero tasks here (no checkboxes). An LLM sees four.

### Architect

So the extraction pipeline has two modes:

1. **Structured extraction** (existing adapters) — pattern-based, fast, deterministic
2. **Intelligent extraction** (new LLM service) — intent-based, slower, non-deterministic

These complement each other. The adapter gives you tasks that are already formatted as tasks. The LLM gives you tasks hiding in prose.

### UX Designer

The user journey for extraction should be:

1. User types `:extract` or triggers via CLI (`threedoors extract`)
2. User provides source: file path, clipboard, or source reference
3. ThreeDoors shows extracted tasks for review
4. User confirms/edits/discards each task
5. Confirmed tasks are added to the task pool

The review step is CRITICAL. Never auto-import LLM-extracted tasks. The user must see and approve each one.

---

## Round 2: Source-Specific Pipelines

### Dev

Let me map out how we get raw text from each source:

**Apple Notes:**
```go
// Already have JXA/AppleScript bridge in internal/adapters/applenotes/
// Can fetch note body text via existing infrastructure
// New: pipe note body to LLM for intent-based extraction
func (e *TaskExtractor) ExtractFromAppleNote(ctx context.Context, noteID string) ([]Task, error) {
    // 1. Fetch note content via AppleScript
    body, err := e.notesClient.GetNoteBody(noteID)
    // 2. Send to LLM for extraction
    tasks, err := e.extractFromText(ctx, body)
    // 3. Return for user review
    return tasks, nil
}
```

**Obsidian:**
```go
// Already have vault reader in internal/adapters/obsidian/
// Can read any .md file
// New: pipe markdown content to LLM for extraction beyond checkbox syntax
func (e *TaskExtractor) ExtractFromObsidianNote(ctx context.Context, filePath string) ([]Task, error) {
    content, err := os.ReadFile(filePath)
    tasks, err := e.extractFromText(ctx, string(content))
    return tasks, nil
}
```

**Transcripts / arbitrary text:**
```go
// Most flexible: user provides text directly
func (e *TaskExtractor) ExtractFromText(ctx context.Context, text string) ([]Task, error) {
    return e.extractFromText(ctx, text)
}

// Or from a file path
func (e *TaskExtractor) ExtractFromFile(ctx context.Context, path string) ([]Task, error) {
    content, err := os.ReadFile(path)
    return e.extractFromText(ctx, string(content))
}
```

**Clipboard:**
```go
// Read from system clipboard — macOS specific via pbpaste
func (e *TaskExtractor) ExtractFromClipboard(ctx context.Context) ([]Task, error) {
    stdout, _, err := e.runner.Run(ctx, "pbpaste")
    return e.extractFromText(ctx, string(stdout))
}
```

### Architect

The common denominator is `extractFromText(ctx, text) → []Task`. All sources eventually produce a text blob. The source-specific methods just handle getting that blob.

The core extraction function:

```go
func (e *TaskExtractor) extractFromText(ctx context.Context, text string) ([]ExtractedTask, error) {
    if len(text) == 0 {
        return nil, fmt.Errorf("extract: empty input text")
    }
    if len(text) > e.maxInputSize {
        return nil, fmt.Errorf("extract: input too large (%d bytes, max %d)", len(text), e.maxInputSize)
    }

    prompt := e.buildExtractionPrompt(text)
    response, err := e.backend.Complete(ctx, prompt)
    if err != nil {
        return nil, fmt.Errorf("extract tasks: %w", err)
    }

    tasks, err := e.parseExtractionResponse(response)
    if err != nil {
        return nil, fmt.Errorf("parse extraction response: %w", err)
    }

    return tasks, nil
}
```

### PM

What's the maximum input size? Meeting transcripts can be enormous.

### Dev

CLI tools have different limits:
- **Claude Code CLI**: Handles large context well, but process startup + response time scales with input
- **Gemini CLI**: Large context window, good for long documents
- **Ollama**: Depends on model, typically 4K-128K tokens

I'd suggest a **chunking strategy** for large inputs:
1. If text < 4K tokens: send as-is
2. If text > 4K tokens: split into overlapping chunks, extract from each, deduplicate results

But for MVP, just set a reasonable limit (e.g., 32KB) and tell the user to trim if it's too long. Chunking is a P2 optimization.

---

## Round 3: The Extraction Prompt

### Dev

The prompt is the most important piece. It determines quality. Here's my draft:

```go
const extractionPromptTemplate = `You are a task extraction assistant. Given the following text,
identify all actionable tasks, to-dos, and commitments.

For each task, extract:
- text: A clear, actionable description (imperative form)
- effort: Estimated effort 1-5 (1=tiny, 5=major project)
- tags: Relevant categories (comma-separated)

Return ONLY a JSON array. No explanation, no markdown fencing.

Example output:
[
  {"text": "Email Sarah about Q2 budget proposal", "effort": 1, "tags": "communication,finance"},
  {"text": "Prep slides for client demo", "effort": 3, "tags": "presentation,client"}
]

If no tasks are found, return an empty array: []

Text to analyze:
---
%s
---`
```

### TEA/QA

Prompt engineering concerns:

1. **JSON-only output** — essential for parsing. Different LLMs handle this differently. Claude is reliable. Ollama models vary. We need a **response validator** that checks for valid JSON and retries once if parsing fails.

2. **Imperative form** — "Email Sarah" not "I should email Sarah." The prompt should enforce this.

3. **Effort estimation** — LLMs are inconsistent at effort estimation. Consider making this optional and letting the user set effort manually.

4. **Tag inference** — Useful but noisy. The LLM might invent tag names that don't match the user's taxonomy. Consider: extract tags but let user edit before import.

### Architect

The prompt should be **configurable**. Power users might want to customize the extraction prompt for their domain. Store a default prompt template, allow override in config:

```yaml
llm:
  prompts:
    extraction: |
      Custom extraction prompt here...
      {{.Text}}
```

But for MVP, hardcode the prompt and make it configurable later.

### UX Designer

The extracted task should carry metadata about its origin:

```go
type ExtractedTask struct {
    Text       string   `json:"text"`
    Effort     int      `json:"effort"`
    Tags       []string `json:"tags"`
    Source     string   // "apple-notes:note-id" or "file:/path/to/file" or "clipboard"
    SourceLine int      // approximate line in source text
    Confidence float64  // LLM's confidence (if available)
}
```

This lets the user trace back to where a task came from.

---

## Round 4: The Review Flow

### UX Designer

The TUI review flow is critical for trust. Here's my proposal:

```
┌─────────────────────────────────────────────────────┐
│ EXTRACTED TASKS (4 found)                           │
├─────────────────────────────────────────────────────┤
│                                                     │
│ ✓ 1. Email Sarah about Q2 budget proposal     [~1]  │
│   2. Ask Mike about server migration timeline [~1]  │
│ ✓ 3. Update design docs before Friday         [~2]  │
│   4. Prep slides for client demo              [~3]  │
│                                                     │
│ Source: Meeting with Sarah — March 10               │
│ Backend: claude-cli                                 │
├─────────────────────────────────────────────────────┤
│ [Space] Toggle  [E] Edit  [Enter] Import Selected   │
│ [A] Select All  [N] Select None  [Esc] Cancel       │
└─────────────────────────────────────────────────────┘
```

Key interactions:
- **Space** toggles selection on current task
- **E** opens inline editor for the selected task (edit text, effort, tags)
- **Enter** imports all selected tasks to the task pool
- **A/N** select all / select none
- **Esc** discards everything

### PM

This is good. I want to add one thing: after import, briefly show a confirmation:

```
Imported 3 tasks from "Meeting with Sarah — March 10"
```

Then return to the doors view. The imported tasks will appear in future door selections.

### Dev

For the CLI interface (`threedoors extract`), the flow is simpler:

```bash
# From file
threedoors extract --file meeting-notes.txt

# From clipboard
threedoors extract --clipboard

# From stdin (pipe)
cat transcript.txt | threedoors extract

# Output: list extracted tasks, ask for confirmation
Found 4 tasks:
  1. Email Sarah about Q2 budget proposal [effort: 1]
  2. Ask Mike about server migration timeline [effort: 1]
  3. Update design docs before Friday [effort: 2]
  4. Prep slides for client demo [effort: 3]

Import all? [y/N/select]:
```

With `--json` flag, output structured JSON for scripting:
```bash
threedoors extract --file notes.txt --json | jq '.[] | .text'
```

---

## Round 5: Edge Cases and Error Handling

### TEA/QA

Edge cases to handle:

1. **Empty source** — No text to extract from. Return clear message: "No text found in source."
2. **No tasks found** — LLM returns empty array. Message: "No actionable tasks found in this text."
3. **LLM unavailable** — No CLI tool found, or tool fails. Message: "LLM service unavailable. Install `claude` or `ollama` to enable task extraction." with instructions.
4. **Malformed LLM response** — JSON parse error. Retry once with stricter prompt. If still fails, show raw response and ask user to report.
5. **Duplicate tasks** — Extracted task matches existing task. Flag with "(possible duplicate)" annotation.
6. **Very long input** — Exceeds limit. Message: "Input too large (X KB). Maximum is 32 KB. Consider splitting into smaller sections."
7. **Binary/non-text input** — User provides a binary file. Detect and reject: "Cannot extract tasks from binary files."
8. **Rate limiting** — Claude CLI has rate limits. Catch and inform user: "Rate limited. Try again in X seconds, or switch to local backend (ollama)."

### Architect

For duplicate detection, we should leverage the existing `DuplicateDetector` component from Epic 13. When extracted tasks are ready for import, run them through dedup against the current task pool.

### Dev

One more consideration: **incremental extraction**. If a user extracts from the same Apple Note twice (maybe they added new content), we should detect which tasks are new vs. already imported. This prevents duplicate imports.

Track extraction history:
```yaml
# ~/.threedoors/extraction-history.jsonl
{"timestamp": "2026-03-11T10:00:00Z", "source": "apple-notes:note-123", "tasks_imported": ["task-abc", "task-def"]}
```

On re-extraction, compare new results against previously imported tasks from the same source.

---

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| S3-D1 | All sources reduce to `extractFromText(text)` | Common text extraction pipeline | Source-specific LLM prompts | Simpler architecture; source metadata carried separately |
| S3-D2 | User review required before import | Interactive selection + edit UI | Auto-import with undo | Trust building; LLMs hallucinate; SOUL.md user control |
| S3-D3 | JSON-only LLM output format | Strict JSON array response | YAML, markdown, or free-text parsing | Reliable parsing across all LLM backends |
| S3-D4 | Clipboard support via pbpaste (macOS) | Platform-specific clipboard access | Cross-platform clipboard library | Mac-first user base; can add Linux later |
| S3-D5 | 32KB input size limit for MVP | Hard limit with clear error message | Chunking strategy for large inputs | YAGNI; chunking adds complexity; users can trim |
| S3-D6 | Extraction history in JSONL | Track source + imported task IDs | No history tracking | Prevents duplicate imports from same source |
| S3-D7 | Hardcoded prompt template for MVP | Default prompt; configurable later | Configurable from day one | Ship fast; prompt customization is P2 |
| S3-D8 | Retry once on malformed LLM response | Single retry with stricter prompt | No retry / unlimited retries | Balances reliability vs. latency; one retry is usually sufficient |
