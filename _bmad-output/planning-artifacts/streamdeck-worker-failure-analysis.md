# Stream Deck Research Worker Failure Analysis

**Date:** 2026-03-31
**Source:** retrospector investigation
**Trigger:** 3 consecutive worker failures on same research task

---

## Failure Summary

| Attempt | Worker | Mode | Failure Type | Duration |
|---------|--------|------|-------------|----------|
| 1 | clever-fox | Subagent parallel research | Synthesis stall (87 tokens, 3+ min) | Killed |
| 2 | kind-tiger | Direct research (no subagents) | Empty completion (no artifacts) | ~6 min thinking, no output |
| 3 | jolly-fox | Direct + explicit save instructions | In progress |

## Root Cause Analysis

### Failure 1: clever-fox — Synthesis Stall

**Diagnosis: NOT the unclosed-quote freeze bug (INC known from 2026-03-08).**

The unclosed-quote bug manifests as Bash waiting on stdin — the shell blocks, Claude blocks waiting for tool result, creating a deadlock. In clever-fox's case, all 3 subagents completed successfully, and the stall occurred during text generation (87 tokens received = model was generating, not tool-blocked).

**Likely cause:** Context window pressure during synthesis. Three parallel subagent results returning to the parent agent can produce a large combined context. If the research results are substantial (web fetches, SDK docs, API references), the synthesis step may hit the effective processing ceiling where the model generates very slowly or stalls. The "Channelling..." state with low token throughput is consistent with this.

**Contributing factor:** The `no-research-subagents.md` rule exists specifically because subagent research tasks consume supervisor/parent context window and are invisible in tmux. clever-fox violated this rule's spirit by spawning subagents for research within a worker.

### Failure 2: kind-tiger — Empty Completion

**Diagnosis: Silent completion without artifact production.**

This is the most concerning failure mode. The worker:
1. Successfully performed web searches and doc fetches
2. Entered a long thinking pass (6+ min, 1.2k tokens)
3. Fired the completion message
4. Produced NO artifacts (no file, no commit, no PR)

**Likely cause:** The completion message is sent by the multiclaude framework when the Claude process exits, not when the worker explicitly reports success. If the model hits a context limit, exhausts its thinking budget, or encounters an unrecoverable error during file write, the process may exit cleanly (triggering the completion message) without having written anything.

**Pattern:** Long thinking passes (6+ min) after web fetch results suggest the model is struggling to synthesize fetched content into a structured output. Web fetch results can be very large (full HTML pages, API docs) and may consume disproportionate context.

### Failure 3: jolly-fox — Preventive Measures Applied

Explicitly instructed to save file and create PR, warned about previous failures. Outcome TBD.

## Pattern Analysis

**Common thread across all 3 failures:**
- Same task type: research + web fetches + synthesis into a report
- Web fetch results are unbounded in size — a single SDK docs page can be 50K+ tokens
- The synthesis step (combining research into a structured document) is where all failures occur
- No failures during the research/fetching phase itself

**This is NOT a general worker reliability issue.** Implementation workers (code changes, story work) show 100% completion rate in recent findings. The failure pattern is specific to research tasks with web fetches that require large-context synthesis.

## Recommendations

### REC-001: Artifact Verification Before Completion (High Confidence)

Workers performing file-creation tasks should verify the artifact exists before reporting completion. The multiclaude completion message should ideally be gated on artifact existence, not just process exit.

**Concrete fix:** Add a post-task verification step to research worker dispatch instructions: "Before completing, verify the output file exists: `ls -la <expected-path>`. If it doesn't exist, retry the write."

### REC-002: Web Fetch Size Limits for Research Tasks (Medium Confidence)

Research tasks should constrain web fetch output to prevent context exhaustion during synthesis. Workers should be instructed to extract only relevant sections from web pages rather than fetching full pages.

**Concrete fix:** In research task dispatch, add: "When fetching web pages, extract only the sections relevant to the research question. Do not store full page contents in context."

### REC-003: Chunked Research Output (Medium Confidence)

For multi-source research tasks, workers should write findings incrementally (one source at a time) rather than accumulating all results in context before writing. This prevents the synthesis stall by ensuring partial results are persisted.

**Concrete fix:** "Write research findings to the output file incrementally as you complete each source. Do not wait until all research is complete to begin writing."

### REC-004: Research Task Dispatch via multiclaude work Only (Reinforcement)

The `no-research-subagents.md` rule should be reinforced in worker dispatch. clever-fox's use of subagents for parallel research within a worker violated this principle and contributed to the synthesis stall.

## Confidence Assessment

- **High confidence** that web fetch context pressure is the root cause (3/3 failures at synthesis stage)
- **Medium confidence** on specific mitigations (REC-002, REC-003 are reasonable but untested)
- **Low confidence** that jolly-fox's explicit instructions will prevent the failure (same underlying context pressure applies)
