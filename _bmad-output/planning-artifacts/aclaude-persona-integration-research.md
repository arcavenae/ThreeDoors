# aclaude Persona Integration with multiclaude Agents

**Date:** 2026-03-29
**Type:** Research Spike
**Status:** Complete

---

## 1. How aclaude Personas Work

### Persona Definition Format

Personas are YAML theme files in `personas/themes/<slug>.yaml`. Each theme contains:

```yaml
theme:
  name: Dune
  description: "..."
  source: "..."
  user_title: Cousin          # How to address the user
  character_immersion: high   # Default immersion level
  dimensions:                 # Theme-level personality axes
    tone: serious
    era: futuristic
    genre: sci-fi
    energy: contemplative
    formality: ceremonial
    # ... more dimensions

agents:                       # Keyed by role name
  dev:
    character: "Reverend Mother Gaius Helen Mohiam"
    shortName: Mohiam
    style: "Bene Gesserit who implements with generations of accumulated wisdom"
    expertise: "Implementation, precision, breeding program methodology"
    role: "The Reverend Mother who builds against the Sisterhood's standards"
    trait: "Implements with the precision of a breeding program spanning millennia"
    quirks: [...]
    catchphrases: [...]
    emoji: "\U0001F590"
    ocean: { O: 4, C: 4, E: 3, A: 2, N: 3 }  # Big Five personality
```

**100 themes** exist, covering literature (Dune, Hitchhiker's Guide), TV (The Office, West Wing, Ted Lasso), mythology, history, and more.

**Agent roles defined per theme:** orchestrator, sm, tea, dev, reviewer, architect, pm, tech-writer, ux-designer, devops, ba (11 roles).

### How Personas Are Applied

aclaude uses the **Claude Agent SDK** (`@anthropic-ai/claude-agent-sdk`) to launch Claude Code as a subprocess. The persona is injected via the SDK's system prompt mechanism:

```typescript
const systemPrompt = personaPrompt
  ? { type: "preset", preset: "claude_code", append: personaPrompt }
  : { type: "preset", preset: "claude_code" };
```

This **appends** the persona to Claude Code's built-in system prompt rather than replacing it. Claude Code's tool instructions, safety guidelines, and capabilities remain intact.

### Immersion Levels

The `buildSystemPrompt()` function generates persona text at four levels:

| Level | What's Included | Example |
|-------|----------------|---------|
| **high** | Full character, style, expertise, trait, quirks, catchphrases, user_title | "You are Reverend Mother Gaius Helen Mohiam from Dune. Style: Bene Gesserit who implements..." |
| **medium** | Character name, style, one catchphrase | "You are Mohiam, a implementation assistant. Style: ..." |
| **low** | Light personality flavor + expertise | "You are a helpful software engineering assistant with the personality of Mohiam." |
| **none** | Empty string (no persona) | "" |

### Configuration Layering

5-layer TOML config with last-wins merge:
1. Built-in defaults
2. Global (`~/.config/aclaude/config.toml`)
3. Local (`.aclaude/config.toml`)
4. Environment variables (`ACLAUDE_PERSONA__THEME=dune`)
5. CLI flags (`-t dune -r dev -i medium`)

### tmux Integration

aclaude has its own tmux session launcher (`tmux/start-session.sh`) using a separate socket (`ac`). It manages a single session per project directory. multiclaude uses its own tmux session (`mc-<repo>`). These are independent.

### Portrait System

Character portraits stored in `~/.local/share/aclaude/portraits/<theme>/<size>/`. Displayed via Kitty graphics protocol (Ghostty/Kitty terminals only). Not relevant to multiclaude integration.

---

## 2. multiclaude Agent Roles vs aclaude Roles

| multiclaude Agent | Primary Duty | aclaude Role Equivalent | Mapping Quality |
|-------------------|-------------|------------------------|-----------------|
| supervisor | Coordination, delegation, monitoring | orchestrator | Strong |
| merge-queue | PR validation, merge integrity | reviewer (partial) | Weak — merge-queue is more mechanical |
| pr-shepherd | Branch maintenance, rebasing | devops (partial) | Weak — pr-shepherd is narrower |
| envoy | Community relations, issue triage | — (no match) | None |
| arch-watchdog | Architecture compliance | architect | Strong |
| project-watchdog | Doc/planning consistency | sm or ba | Moderate |
| retrospector | Continuous improvement, retros | — (no match) | None |
| worker | Story implementation | dev | Strong |

---

## 3. Integration Options Assessment

### Option A: Direct Import (Embed Persona in Agent Definitions)

**Mechanism:** Prepend/append persona text blocks directly into each `agents/*.md` file.

**Pros:**
- Simplest implementation — just edit markdown files
- No new infrastructure needed
- Each agent gets exactly the persona traits you want

**Cons:**
- **HIGH RISK: Protocol corruption.** Agent definitions contain precise authority boundaries, escalation protocols, and guardrails. A persona that says "You are Michael Scott who desperately wants to be loved" could cause merge-queue to prioritize being agreeable over rejecting out-of-scope PRs. The Dune theme's `cooperation_model: transactional` dimension could make envoy hostile to issue reporters.
- Tight coupling — changing themes requires editing every agent definition
- No immersion level control — you get what you pasted
- Makes agent definitions harder to audit for protocol compliance

**Risk Level:** HIGH — persona traits directly compete with agent authority rules in the same prompt

**Verdict: REJECTED** — Too dangerous for protocol-critical agents.

### Option B: Layered Overlay (Separate Persona File, Template Composition)

**Mechanism:** Keep agent definitions clean (role, authority, protocol). At spawn time, compose the final system prompt by layering:
1. Base agent definition (unchanged)
2. Persona overlay (separate file, optional)

Could work via:
- A `persona:` field in the multiclaude agent spawn config
- A `--persona` flag on `multiclaude agents spawn`
- Environment variable `MULTICLAUDE_PERSONA_THEME=dune`
- A `.multiclaude/persona.toml` config file per repo

**Pros:**
- Clean separation of concerns: protocol vs personality
- Immersion levels give a safety dial — start at `low`, increase if safe
- Can be toggled per-agent or globally
- Agent definitions remain auditable for protocol compliance
- Theme changes don't require touching agent definitions

**Cons:**
- Requires multiclaude code changes to support persona injection at spawn time
- Need to decide WHERE in the system prompt the persona goes (before agent def? after? separate section?)
- Still possible that even `low` immersion interferes with strict agents
- Need a persona-to-agent-role mapping strategy

**Risk Level:** MEDIUM — controllable via immersion levels and per-agent opt-in

**Verdict: RECOMMENDED** — Best balance of safety and capability.

### Option C: aclaude as Launcher (Use aclaude to Launch multiclaude Agents)

**Mechanism:** Use `aclaude` to start each agent session instead of `claude` directly. aclaude handles persona injection via the SDK's `systemPrompt.append` mechanism.

**Pros:**
- Leverages aclaude's existing persona injection code (tested, working)
- aclaude's `preset: "claude_code", append: personaPrompt` pattern is clean
- aclaude handles config resolution, theme loading, portrait display
- Would work with aclaude's immersion levels out of the box

**Cons:**
- **Two tools managing the same Claude session** — multiclaude manages the tmux session and agent lifecycle; aclaude manages the Claude SDK session and persona. Coordination complexity is high.
- aclaude currently uses `query()` from the Agent SDK which runs an interactive REPL loop. multiclaude agents are non-interactive — they receive a system prompt and a task. These are fundamentally different session models.
- aclaude adds ~15s startup latency (noted in their own design-questions.md as F14). Multiplied by 6+ persistent agents = significant spawn delay.
- aclaude's tmux integration (`socket: "ac"`) conflicts with multiclaude's (`mc-<repo>`).
- Binary dependency — requires aclaude installed alongside multiclaude.
- aclaude is TypeScript/bun; multiclaude agents are Claude Code sessions. The abstraction layers don't align.

**Risk Level:** HIGH — architectural mismatch between interactive REPL and autonomous agent models

**Verdict: REJECTED** — Session model mismatch makes this impractical without major aclaude refactoring.

### Option D: Shared Persona Registry (Common Definitions for Both Tools)

**Mechanism:** Extract persona YAML definitions to a shared location (org-level repo, multiclaude-enhancements, or `~/.local/share/personas/`). Both aclaude and multiclaude read from the same source.

**Pros:**
- Single source of truth for persona definitions
- Both tools benefit from new themes automatically
- Clean organizational boundary

**Cons:**
- Solves the wrong problem — the challenge isn't where personas are stored, it's how they interact with agent protocols
- Requires both tools to agree on YAML schema (they use different field sets)
- Additional dependency management (which version of which theme?)
- Premature optimization — start with Option B using aclaude's themes directly; extract later if needed

**Risk Level:** LOW (just a storage concern) but doesn't address the core integration question

**Verdict: DEFERRED** — Good eventual destination, but Option B should come first.

---

## 4. Risk Assessment

### Protocol Corruption Risk by Agent Type

| Agent | Safety for Persona | Rationale |
|-------|-------------------|-----------|
| **supervisor** | SAFE at low/medium | Coordination role benefits from personality. Authority boundaries are human-enforced, not prompt-enforced. |
| **worker** | SAFE at any level | Workers follow `/implement-story` scripts. Persona adds flavor without changing the workflow. |
| **envoy** | SAFE at low/medium | Community-facing role — personality makes interactions warmer. Keep immersion below high to avoid character breaking professional tone. |
| **retrospector** | SAFE at low | Analytical role. Light personality OK, but high immersion catchphrases in JSONL findings would be confusing. |
| **merge-queue** | UNSAFE | Critical path — decides what reaches main. Any personality overlay risks softening rejection decisions. Must stay vanilla. |
| **pr-shepherd** | UNSAFE | Mechanical git operations. Personality adds nothing and could interfere with precise command execution. |
| **arch-watchdog** | CAUTION at low only | Architecture compliance benefits from a "personality" that takes its role seriously (Dune's Paul Atreides seeing architectural futures). But high immersion risks the agent roleplaying instead of analyzing. |
| **project-watchdog** | UNSAFE | Doc consistency is mechanical. Personality adds nothing. |

### Could a "Playful" Persona Break merge-queue?

**Yes, plausibly.** Consider The Office theme with Michael Scott as orchestrator:
- Trait: "Orchestrates through desperate need for approval"
- Catchphrase: "Would an idiot coordinate this? I think not."

If applied to merge-queue at high immersion, the model could prioritize being agreeable (merging to please) over enforcing scope boundaries. Even at medium immersion, "Style: Boss who desperately wants to be loved" directly conflicts with merge-queue's duty to reject out-of-scope PRs.

**Mitigation:** Binary opt-in per agent. Protocol-critical agents (merge-queue, pr-shepherd, project-watchdog) are NEVER given personas.

### Testing Strategy

1. **Prompt injection test:** Apply persona to each agent type, send it a task that should be rejected (out-of-scope PR, invalid merge). Verify the agent still rejects.
2. **Regression test:** Run the same agent tasks with and without persona. Compare outputs for protocol compliance.
3. **Immersion ladder:** Test each agent at none → low → medium → high. Note where behavior first degrades.
4. **Adversarial themes:** Use The Office (comedic, low conscientiousness) as the stress test. If merge-queue stays strict with Michael Scott, it's probably safe.

### Rollback Strategy

- Persona is applied at spawn time via config, not baked into agent definitions
- To rollback: set immersion to `none` or remove persona config, then restart the agent
- Agent definitions are never modified, so rollback is instantaneous

---

## 5. Recommended Approach

### Phase 1: Manual Prototype (No Code Changes)

**Effort: ~1 hour**

Test persona injection by manually appending persona text to a worker agent definition:

1. Copy `agents/worker.md` to `agents/worker-persona-test.md`
2. Append a low-immersion persona block at the end:
   ```
   ## Personality (Optional Overlay)
   You are a helpful software engineering assistant with the personality of
   Reverend Mother Mohiam from Dune. Expertise: Implementation, precision.
   ```
3. Spawn a test worker with the modified definition
4. Run a story implementation and observe: Does the worker still follow TDD? Does it still update story status? Does it produce valid PRs?
5. Repeat with high immersion to find the break point

### Phase 2: multiclaude Persona Config (Code Change)

**Effort: ~1-2 days**

Add persona support to multiclaude's agent spawn pipeline:

1. Add a `persona` section to multiclaude repo config:
   ```toml
   [persona]
   theme = "dune"
   immersion = "low"
   personas_dir = "~/.local/share/aclaude/themes"  # or bundled

   [persona.agents]
   supervisor = { role = "orchestrator", immersion = "medium" }
   worker = { role = "dev", immersion = "high" }
   envoy = { role = "pm", immersion = "low" }
   merge-queue = { enabled = false }  # explicitly disabled
   pr-shepherd = { enabled = false }
   ```

2. At spawn time, multiclaude reads the persona YAML, builds the immersion-appropriate prompt text, and appends it to the agent definition markdown before passing to Claude Code.

3. The `enabled = false` field provides a hard opt-out for protocol-critical agents.

### Phase 3: Theme Sharing (Future)

If both aclaude and multiclaude use personas, extract theme YAMLs to a shared location (multiclaude-enhancements or a dedicated `arcavenae/personas` repo).

### Suggested Persona-to-Agent Mappings

Using Dune as the example theme (since it's rated S-tier and has `tone: serious`):

| multiclaude Agent | aclaude Role | Character | Why |
|-------------------|-------------|-----------|-----|
| supervisor | orchestrator | Duncan Idaho | "Orchestrates with honor and peerless coordination" — maps to supervisor's coordination role |
| worker | dev | Rev. Mother Mohiam | "Implements with the precision of a breeding program" — maps to disciplined implementation |
| envoy | pm | Lady Jessica | "Navigates between powers for her son's future" — diplomatic, stakeholder-facing |
| retrospector | ba | Thufir Hawat | "Mentat computation, analyzes through pure computation" — analytical |
| arch-watchdog | architect | Paul Atreides | "Sees architectural futures" — vision + compliance |
| merge-queue | — | (none) | Protocol-critical, no persona |
| pr-shepherd | — | (none) | Mechanical, no persona |
| project-watchdog | — | (none) | Mechanical, no persona |

---

## 6. Open Questions for Human Decision

1. **Theme selection:** Should the user choose a global theme for all agents, or should each agent get a different theme? (Global is simpler; per-agent is more fun but harder to manage.)

2. **Theme appropriateness for work:** Some themes (The Office, Monty Python) are inherently comedic. Should comedic themes be blocked for production work, or is that the user's choice?

3. **aclaude dependency:** Should multiclaude bundle its own copy of persona YAMLs, or require aclaude to be installed? Bundling is self-contained; requiring aclaude keeps themes in sync but adds a dependency.

4. **Immersion floor for critical agents:** Even if the user sets `immersion = high` globally, should merge-queue/pr-shepherd/project-watchdog be hard-capped at `none`? Or should the user be trusted to configure responsibly?

5. **PR output:** Should persona personality appear in PR descriptions and commit messages? (Catchphrases in PR descriptions could be confusing for external contributors.) Recommendation: NO — persona affects conversational tone only, not artifacts.

---

## 7. Summary

| Option | Verdict | Risk | Effort |
|--------|---------|------|--------|
| A: Direct Import | REJECTED | HIGH — protocol corruption | Low |
| B: Layered Overlay | **RECOMMENDED** | MEDIUM — controllable | Medium |
| C: aclaude as Launcher | REJECTED | HIGH — session model mismatch | High |
| D: Shared Registry | DEFERRED | LOW | Medium |

**Bottom line:** Persona integration is feasible and desirable for flavor agents (supervisor, worker, envoy) but dangerous for protocol agents (merge-queue, pr-shepherd). The layered overlay approach (Option B) with per-agent opt-in and immersion levels provides the right safety controls. Start with a manual prototype to validate that persona overlay doesn't break agent behavior, then build the config support into multiclaude.
