# Party Mode Session 1: LLM Service Use Cases Deep Dive

**Date:** 2026-03-11
**Participants:** PM, Architect, UX Designer, Dev, TEA/QA, SM
**Topic:** Which LLM services deliver the most value for ThreeDoors users? Ruthless prioritization.

---

## Round 1: Use Case Inventory

### PM (Opening)

Let me frame the question through SOUL.md's lens. ThreeDoors exists to reduce decision friction and get humans moving. Every LLM service we consider must pass this test: **does it reduce friction to starting a task, or does it add complexity?**

The six proposed services:

1. **Task Extraction** — pull tasks from Apple Notes, Obsidian, transcripts
2. **Task Enrichment** — add context, tags, estimates, dependencies
3. **Task Breakdown** — decompose large tasks into subtasks
4. **Task Organization** — suggest groupings, priorities, sequences
5. **Pattern Analysis** — observe user behavior, identify trends
6. **Task Recommendation** — suggest which tasks are most suitable

Let me prioritize by user impact.

### Architect

Before we prioritize, let me clarify what "LLM service via CLI" actually means architecturally. The user's clarification is critical: ThreeDoors calls out to `claude`, `gemini`, or `ollama` CLIs as subprocesses. This is fundamentally different from our existing `LLMBackend` interface in `internal/intelligence/llm/backend.go`, which uses HTTP APIs.

CLI invocation means:
- **Subprocess execution** via `os/exec` — not HTTP clients
- **Stdin/stdout piping** — send prompt on stdin, read response on stdout
- **No persistent connection** — each invocation is a fresh process
- **CLI-specific flags** — each tool has its own argument format
- **Authentication** — handled by the CLI tool itself (OAuth, API keys, etc.)

This changes the cost model. CLI calls have higher latency (process startup) but zero auth management overhead from our side.

### UX Designer

From a user experience perspective, the services fall into two categories:

**Active services** (user explicitly requests):
- Task Extraction ("import my notes into ThreeDoors")
- Task Breakdown ("break this task into smaller pieces")

**Passive services** (happen automatically or are suggested):
- Task Enrichment ("your task is missing context — here's what I inferred")
- Task Organization ("these tasks seem related — group them?")
- Pattern Analysis ("you tend to avoid writing tasks on Mondays")
- Task Recommendation ("based on your patterns, start with this one")

SOUL.md says "show less, not more." Passive services risk feeling intrusive. Active services feel like tools the user reaches for.

### Dev

Looking at what we already have:
- `LLMTaskDecomposer` in `internal/intelligence/llm/decomposer.go` — already does task breakdown via HTTP API
- `AgentService` in `internal/intelligence/agent_service.go` — orchestrates decomposition + git output
- Epic 14 (LLM Task Decomposition) is COMPLETE — 2 stories shipped

We're not starting from zero. The question is: do we extend the existing `LLMBackend` interface to support CLI backends, or create a new abstraction?

### TEA/QA

From a testing perspective, CLI-based services are actually easier to test than HTTP APIs. We can:
- Mock the CLI binary with a shell script that returns canned responses
- Test the prompt construction without any network
- Verify error handling for CLI not found, timeout, non-zero exit codes
- Use `testdata/` fixtures for expected CLI outputs

The testing story is clean. My concern is quality consistency — different LLMs will produce different quality outputs. We need to define what "good enough" looks like for each service.

---

## Round 2: Ruthless Prioritization

### PM

Let me apply the "friction reduction" test to each:

| Service | Friction Reduced | Complexity Added | Verdict |
|---------|-----------------|------------------|---------|
| **Task Extraction** | HIGH — users have tasks scattered everywhere | MEDIUM — need source-specific parsing | **P0** |
| **Task Breakdown** | HIGH — large tasks cause paralysis | LOW — already have decomposer | **P0** |
| **Task Enrichment** | MEDIUM — nice but not blocking action | LOW — simple prompt + response | **P1** |
| **Task Recommendation** | MEDIUM — helps when all 3 doors feel equal | MEDIUM — needs pattern data | **P1** |
| **Task Organization** | LOW — ThreeDoors deliberately avoids organization | HIGH — fights core philosophy | **P2/Icebox** |
| **Pattern Analysis** | LOW — interesting but not actionable in-the-moment | HIGH — needs data pipeline | **P2/Icebox** |

**My P0 recommendation:** Task Extraction and Task Breakdown. These directly reduce the two biggest friction points: "I have tasks but they're not in ThreeDoors" and "this task feels too big to start."

### Architect

I agree with PM's prioritization, and I want to add architectural reasoning:

**Task Extraction** is the highest-leverage service because it feeds the entire ThreeDoors pipeline. Without tasks, nothing else matters. And it's the perfect CLI-LLM use case: pipe a blob of text to an LLM, get structured task YAML back.

**Task Breakdown** is already proven (Epic 14). Extending it to use CLI backends is low-risk, high-value.

**Task Enrichment** is a natural extension — same pipe-text-get-structured-output pattern.

**Task Organization and Pattern Analysis** fight SOUL.md. ThreeDoors explicitly says "Not a second brain (no knowledge graph, no linking, no tagging taxonomy)." Organization suggestions push toward that anti-pattern.

### UX Designer

I want to champion **Task Recommendation** as P1, not P2. Here's why:

When a user sees 3 doors and all feel equally overwhelming, they re-roll. And re-roll. And re-roll. This is the core failure mode of ThreeDoors. An LLM that can whisper "this one is the quickest win" or "you've been avoiding this — 15 minutes and it's done" directly addresses the philosophy.

But it has to be subtle. Not a popup. Maybe a faint annotation on one door: `~15 min · good starting point`. The LLM enriches the door selection, not the task list.

### SM

From a sprint planning perspective, I'd sequence it as:

1. **CLI Backend Abstraction** (infrastructure — enables everything else)
2. **Task Extraction** (highest user value, feeds the pipeline)
3. **Task Breakdown via CLI** (extend existing decomposer)
4. **Task Enrichment** (natural next step, same pattern)
5. **Task Recommendation** (needs enrichment data + pattern data first)

Each builds on the previous. No skipping.

### Dev

I want to flag a technical consideration for Task Extraction specifically. The user mentioned "Apple Notes, Obsidian docs, transcripts." We already have adapters for Apple Notes and Obsidian — they extract tasks using pattern matching (checkbox syntax, etc.).

The LLM-powered extraction is different: it understands **intent**. "I should probably email Sarah about the budget" in a meeting transcript is a task, but no regex will catch it. This is where LLM extraction shines — converting unstructured natural language into structured tasks.

The pipeline would be:
1. User provides raw text (paste, file path, or source reference)
2. ThreeDoors pipes it to `claude`/`gemini`/`ollama` with an extraction prompt
3. LLM returns structured tasks (JSON or YAML)
4. ThreeDoors parses and imports into task pool
5. User reviews and confirms

---

## Round 3: Anti-Patterns and Risks

### TEA/QA

Risks I see:

1. **LLM hallucination in extraction** — LLM invents tasks that weren't in the source text. Mitigation: always show extracted tasks for user confirmation before import.

2. **Inconsistent output format** — Different LLMs format JSON/YAML differently. Mitigation: strict schema validation + retry with format correction prompt.

3. **CLI availability** — User may not have `claude` or `gemini` installed. Mitigation: graceful degradation with clear error messages. `ollama` as offline fallback.

4. **Cost/latency** — Cloud LLM CLIs cost money per call. Mitigation: batch operations where possible, cache results, show cost estimates.

5. **Privacy** — SOUL.md says "Local-First, Privacy-Always." Sending task text to cloud LLMs violates this for some users. Mitigation: Ollama as default, cloud as opt-in with explicit consent.

### Architect

The privacy point is critical. Our architecture must support a **tiered privacy model**:

- **Tier 1 (Default): Local-only** — Ollama, llama.cpp. No data leaves the machine.
- **Tier 2 (Opt-in): Cloud with consent** — Claude Code CLI, Gemini CLI. User explicitly enables. Data goes to API provider.
- **No Tier 3** — We never proxy through our own servers. We never store user data remotely.

This aligns perfectly with SOUL.md's "No telemetry, no analytics, no phone-home."

### PM

One more anti-pattern to call out: **over-automating**.

ThreeDoors is not an AI task manager. It's a human achievement partner. The LLM services should feel like having a helpful friend, not an AI overlord. Every LLM interaction should be:
- **User-initiated** (not automatic)
- **Transparent** (show what the LLM did)
- **Reversible** (user can undo/reject)
- **Optional** (app works perfectly without any LLM)

---

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| S1-D1 | Task Extraction is P0 | LLM-powered natural language extraction from any text source | Regex/pattern-only extraction | LLMs understand intent; regex only catches format |
| S1-D2 | Task Breakdown is P0 | Extend existing Epic 14 decomposer to CLI backends | Build new decomposer from scratch | Already have working decomposer; just need new backend |
| S1-D3 | Task Enrichment is P1 | LLM adds context, tags, estimates on demand | Automatic enrichment without user action | SOUL.md: user-initiated, not automatic |
| S1-D4 | Task Recommendation is P1 | Subtle door annotations from LLM analysis | Prominent recommendation UI / auto-reordering | Must not fight "3 doors" constraint |
| S1-D5 | Task Organization is P2/Icebox | Defer indefinitely | Build organization features | Fights "Not a second brain" (SOUL.md) |
| S1-D6 | Pattern Analysis is P2/Icebox | Defer indefinitely | Build analytics dashboard | Not actionable in-the-moment; adds complexity |
| S1-D7 | Privacy-tiered LLM model | Local-first (Ollama default), cloud opt-in | Cloud-first with local fallback | SOUL.md: "Local-First, Privacy-Always" |
| S1-D8 | All LLM services user-initiated | Explicit user action triggers LLM | Background/automatic LLM calls | SOUL.md: reduce friction, not add automation |
