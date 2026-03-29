# Dark Factory & Human/AI Work Stream Differentiation — Research Artifact

**Date:** 2026-03-29
**Rounds:** 10 party mode sessions
**Participants:** PM (John), Architect (Winston), Innovation Strategist (Victor), QA (Murat), Analyst (Mary), UX Designer (Sally), Creative Problem Solver (Dr. Quinn), Brainstorming Coach (Carson), Storyteller (Sophia), Scrum Master (Bob), Quick Flow Dev (Barry), TEA (Murat), BMad Master

---

## Table of Contents

1. [Terminology & Taxonomy](#round-1-terminology--taxonomy)
2. [Dark Factory Repo Architecture](#round-2-dark-factory-repo-architecture)
3. [Multi-Variant Gallery Model](#round-3-multi-variant-gallery-model)
4. [Iterative Dispose-and-Rebuild Cycle](#round-4-iterative-dispose-and-rebuild-cycle)
5. [AI Judges Panel](#round-5-ai-judges-panel)
6. [Autonomy & Governance](#round-6-autonomy--governance)
7. [Implementation Roadmap](#round-7-implementation-roadmap)
8. [Integration with multiclaude](#round-8-integration-with-multiclaude)
9. [Edge Cases & Failure Modes](#round-9-edge-cases--failure-modes)
10. [Final Synthesis & Recommendations](#round-10-final-synthesis--recommendations)

---

## Round 1: Terminology & Taxonomy

### Key Terms Established

| Term | Definition |
|------|-----------|
| **Dark Factory** | Autonomous AI development environment that produces software without human-in-the-loop during execution |
| **Golden Repo** | The authoritative, human-governed repository representing the actual product |
| **Autonomy Spectrum (L0-L4)** | Classification from fully human to fully AI-evaluated work |
| **Provenance** | Metadata tracking who/what created an artifact and under what conditions |
| **Factory Run** | A single execution cycle of a dark factory, producing one variant |
| **Gallery** | Collection of variants from parallel factory runs, presented to human for selection |
| **Dispose-and-Rebuild** | Pattern where code is discarded after each cycle; learnings feed back into specs |

### Autonomy Spectrum

| Level | Label | Description | Example |
|-------|-------|-------------|---------|
| L0 | Human-crafted | 100% human authored | Developer writes code directly |
| L1 | AI-assisted | Human directs, AI generates, human reviews | Copilot, Claude Code in interactive mode |
| L2 | AI-supervised | AI plans and executes, human approves gates | Current multiclaude workers |
| L3 | AI-autonomous | AI runs end-to-end, human reviews output only | Dark factory with human judge |
| L4 | AI-evaluated | AI runs and AI judges evaluate; human sees gallery | Full dark factory with AI judges panel |

### Provenance Metadata Schema (Proposed)

Three technical primitives needed:
1. **Origin tags** — metadata on every artifact declaring provenance level
2. **Factory ID** — unique identifier linking to spec version, agent config, and evaluation criteria
3. **Lineage chain** — traceability from final artifact back through decision points

### Consumers of Provenance Data

1. **Human developer/owner** — trust and review calibration
2. **Reviewer (human or AI)** — scrutiny level
3. **Compliance/audit** — responsibility tracking
4. **System itself** — preventing dark factory contamination of golden repo
5. **Next factory iteration** — distinguishing prior factory output from human curation

### Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Work classification | L0-L4 Autonomy Spectrum | Binary human/AI; Simple provenance tags | Spectrum captures nuance of AI-assisted vs autonomous; binary is too coarse |
| Dark factory definition | Autonomous AI dev without human-in-the-loop during execution | Any automated CI/CD; Any AI-assisted development | Distinguishes from CI/CD (automated pipeline) and AI-assisted (human in loop) |

---

## Round 2: Dark Factory Repo Architecture

### Architecture

```
golden-repo/           # The real product (multiclaude-governed)
├── .threedoors/       # Product code, tests, docs
├── CLAUDE.md          # Full governance rules
└── ROADMAP.md         # Scope gates active

dark-factory-alpha/    # Disposable clone (ephemeral)
├── .threedoors/       # Generated code (disposable)
├── CLAUDE.md          # Relaxed rules — no story requirement, no scope gates
├── .factory-manifest.yaml  # Links back to golden repo + spec version
└── specs/             # COPIED from golden repo — read-only reference
```

### Key Architectural Principles

- **Dark repos are GitHub repos** — not local directories. Own CI, branches, PRs.
- **Specs are copied in, not linked** — snapshot at creation for reproducibility.
- **No git relationship** — NOT forks. Fresh repos from template.
- **Naming convention**: `{project}-df-{variant}-{run-number}`
- **Lifecycle**: Create → Run → Evaluate → Archive or Delete. Never merge back by default.

### Code Re-Entry Protocol (Three Tiers)

1. **Spec-only re-entry** (default): Extract insights, update golden repo specs. No code crosses.
2. **Cherry-pick re-entry** (explicit approval): Specific files via normal PR, tagged `provenance: dark-factory`, elevated scrutiny.
3. **Wholesale adoption** (rare, high ceremony): Entire module replacement. Human sign-off, full test suite, architecture review.

**Governance rule:** Code never flows from dark factory to golden repo without explicit human decision and a tagged PR.

### Contamination Prevention

| Vector | Mitigation |
|--------|-----------|
| Git push accident | No remotes pointing to golden repo in factory repos |
| CI cross-trigger | Dark factory GitHub App has zero permissions on golden repo |
| Dependency publishing | Factory CLAUDE.md forbids publishing packages |
| Copy-paste | Re-entry protocol governance; future: code similarity scanning |

### Factory Manifest Schema

```yaml
factory:
  golden_repo: "arcavenae/ThreeDoors"
  spec_snapshot: "sha256:abc123"
  created: "2026-03-29T10:00:00Z"
  variant: "alpha"
  run: 1
  autonomy_level: L3
  agents:
    supervisor: claude-opus-4-6
    workers: 3
  constraints:
    max_duration: "4h"
    max_cost: "$50"
    no_external_deps: true
```

### Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Dark factory isolation | Separate GitHub repos (not forks, not branches) | Fork of golden repo; Branch in golden repo | Forks create implicit upstream links; branches share CI/secrets |
| Code re-entry | Three-tiered protocol with human gate | Never merge back; Auto-merge | Never merge is too wasteful; auto-merge too risky |
| Governance | Relaxed CLAUDE.md — no story requirements, no scope gates | Full golden repo governance; No governance | Full governance kills exploration; no governance = unusable output |
| Naming | `{project}-df-{variant}-{run}` | Random names; Same name as golden | Clear, parseable convention prevents confusion |
| Secrets/CI | Dedicated minimal-scope credentials, no deployment pipeline | Shared credentials; No CI | Shared creds = contamination risk; No CI = no quality signal |

---

## Round 3: Multi-Variant Gallery Model

### Gallery Architecture

Three layers:
1. **The Spec** — what to build (fixed per generation, refined between generations)
2. **The Divergence Prompts** — how to approach it (varies per variant, AI-generated)
3. **The Factory Output** — the actual working app (disposable, evaluated by human)

### Controlled Divergence

Fix the **what** (features, acceptance criteria), vary the **how** (architecture, UX approach, technology choices). Each variant gets a divergence prompt pushing it in a specific direction.

Examples: "prioritize keyboard navigation", "optimize for mobile", "minimize dependencies", "maximize visual feedback", "prioritize accessibility"

### Gallery Manifest

```yaml
gallery:
  id: "gallery-2026-03-29-001"
  golden_repo: "arcavenae/ThreeDoors"
  spec_version: "v2.1.0"
  generation: 1
  variants:
    - id: "alpha"
      repo: "arcavenae/threedoors-df-alpha-001"
      status: "complete"
      approach_summary: "Minimal UI, keyboard-first"
      deploy_url: "https://alpha-001.darkfactory.dev"
    - id: "beta"
      repo: "arcavenae/threedoors-df-beta-001"
      status: "complete"
      approach_summary: "Rich visual, mouse-friendly"
    - id: "gamma"
      repo: "arcavenae/threedoors-df-gamma-001"
      status: "complete"
      approach_summary: "Hybrid with context-aware defaults"
  feedback:
    - generation: 1
      preferences: []
```

### Human Interaction with Gallery

The human is a **curator, not a builder**:
1. **Gallery Landing**: All variants side-by-side with one-sentence descriptions
2. **Try Mode**: Click into any variant to use it as a running instance
3. **Annotate**: Tag moments while using: "love this", "hate this", "interesting idea"
4. **Compare**: Side-by-side comparison of specific features across variants
5. **Synthesize**: Write preference statements (not code review comments)
6. **Next Generation**: Preferences feed into spec refinements

### Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Gallery size | 3-5 variants per generation | Single variant; 10+ | Single = normal dev; 10+ overwhelms human |
| Generation limit | 3-5 generations before convergence or rethink | Unlimited | Prevents infinite loops; forces spec clarity |
| Divergence mechanism | Controlled divergence prompts (fix what, vary how) | Random; Identical specs | Random = incomparable; identical = near-duplicates |
| Human interaction | Try running apps, annotate, preference statements | Code review; Spec editing | Curator role, not builder role |
| Feedback structure | Categorized preference statements | Freeform; Numeric scoring | Structured for factories; rich for nuance |

---

## Round 4: Iterative Dispose-and-Rebuild Cycle

### Four Phases

1. **Run**: Dark factories produce variants
2. **Evaluate**: Human tries variants, records preferences
3. **Extract**: Learnings distilled into spec refinements
4. **Dispose**: Code deleted; specs are the only surviving artifact

### Spec Lifecycle

```
specs/
├── v1.0.0/                    # Immutable — used by generation 1
│   ├── prd.md
│   ├── architecture.md
│   ├── ux-guidelines.md
│   └── acceptance-criteria.md
├── v1.1.0/                    # Immutable — used by generation 2
│   ├── prd.md                 # Updated with gen-1 learnings
│   ├── CHANGELOG.md           # What changed and why
├── feedback/
│   ├── gen-1-preferences.md
│   ├── gen-1-extraction.md    # AI-distilled spec refinements
└── divergence/
    ├── gen-1-prompts.md
    └── gen-2-prompts.md
```

### Spec Refinement Protocol

1. Human writes preference statements (experiential, emotional)
2. Extractor agent translates preferences into spec requirements
3. Human reviews and approves proposed spec changes
4. Approved changes create new immutable spec version
5. CHANGELOG records: what changed, which feedback drove it, which variant inspired it

### What Survives Disposal

| Preserved | Discarded |
|-----------|-----------|
| Spec refinements | Source code |
| Human feedback | Git history |
| Test results summary | Full test suites |
| Behavioral fingerprints | Build artifacts |
| Performance profiles | Dependencies |
| Judge reports | CI logs (beyond summary) |

### Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Disposal policy | Full code deletion after learning extraction | Archive indefinitely; Partial deletion | Archives accumulate baggage; partial defeats purpose |
| What survives | Spec refinements, feedback, test summaries, fingerprints | Code; Full test suites; Git history | Code is deliberately discarded; metadata captures learnings |
| Spec lifecycle | Immutable versioned snapshots in golden repo | Mutable document; Specs in dark factory | Immutability = reproducibility; golden repo = non-disposable |
| Feedback translation | AI-assisted extraction with human approval | Fully manual; Fully automated | Manual too slow; automated loses judgment |

---

## Round 5: AI Judges Panel

### Two-Tier Evaluation

**Tier 1 (Automated — CI-based):**
- Build success
- Test pass rate
- Lint/vet compliance
- Spec conformance (AC coverage)
- Performance benchmarks
- Binary: pass/fail + quantitative scores

**Tier 2 (AI Judges — qualitative):**
- UX coherence and intuitiveness
- Innovation quality and elegance
- Architectural patterns
- Single agent with multiple evaluation prompts
- Produces written reports with reasoning

### Judge Panel Composition

| Judge | Lens | Key Criteria |
|-------|------|-------------|
| Spec Conformance | Does it do what the spec says? | AC coverage, feature completeness |
| Code Quality | Is the code well-crafted? | Patterns, coverage, error handling |
| UX | Is it pleasant to use? | Intuitiveness, accessibility, coherence |
| Performance | Is it fast? | Startup, memory, latency, binary size |
| Innovation | Does it introduce clever solutions? | Novel approaches, elegant patterns |

### Bias Mitigation

1. **Rotate judge prompts** across generations
2. **Human override** always trumps judge evaluation
3. **Judge calibration** — compare to human preferences periodically
4. **Transparent scoring** — all reports human-readable

### Flow

1. Factories produce 5 variants
2. AI judges evaluate all 5
3. Judges eliminate obviously failing variants
4. Judges rank remaining with scores and reasoning
5. Human sees top 3+ with judge reports attached
6. Human makes final call based on experiential evaluation

### Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Judge role | Pre-filter and advise | Final decision-maker; No judges | Can't judge subjective fit; no judges wastes time |
| Structure | Two-tier: automated CI + single AI agent | Five separate agents; All automated | Multiple agents = overhead; all automated can't assess quality |
| Bias mitigation | Rotation, human override, transparency, calibration | Fixed judges; Trust AI fully | Fixed = feedback loops; full trust removes agency |
| Synthesis | Single judge panel report per generation | Per-judge only; No synthesis | Single doc is human-friendly |

---

## Round 6: Autonomy & Governance

### Three Governance Zones

| Zone | Authority | Examples | Enforcement |
|------|-----------|----------|-------------|
| **Green** (fully autonomous) | Create repos/branches/PRs in dark factory. Write code, tests, docs. Run CI. | Standard factory operations | GitHub App permissions scoped to `{project}-df-*` |
| **Yellow** (autonomous + audit trail) | Add dependencies (logged). Deviate from spec (documented). Exceed budget by up to 20%. | Architectural decisions, dependency additions | `deps-change.yaml` log, `deviation-log.md`, budget alerts |
| **Red** (requires human) | Access external APIs. Publish packages. Create repos outside namespace. Modify golden repo. Exceed budget >20%. | External integrations, golden repo changes | No API keys provisioned, no publish creds, branch protection |

### Inter-Factory Governance

1. **No cross-factory contamination** — factory A cannot read factory B's code
2. **No factory communication** — factories don't coordinate
3. **Resource fairness** — no factory monopolizes compute
4. **Independent failure** — one crash doesn't affect others

### Generation Timing

**Time-bounded with early completion**: Set max duration (e.g., 4 hours). Early finishers enter polishing mode. Late factories evaluated as-is.

### Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Governance model | Three zones with technical enforcement | Policy-only; Full lockdown | Policy leaks; lockdown prevents exploration |
| Inter-factory isolation | Complete isolation | Shared learnings; No isolation | Shared reduces divergence; isolation = independence |
| Generation timing | Time-bounded with early completion | Completion-bounded; Quality-bounded | Prevents blocking; incomplete is informative |
| Budget enforcement | Soft at 80%, hard at 100% + 20% buffer | No limits; Hard with no buffer | No limits = runaway; no buffer = abrupt stop |

---

## Round 7: Implementation Roadmap

### Phased Approach

| Phase | Name | Scope | Value | Prerequisites |
|-------|------|-------|-------|---------------|
| 0 | Provenance Tagging | Add L0-L4 labels to stories, commits, PRs in golden repo | Immediate traceability | None |
| 1 | Single Dark Factory | Manual: create template repo, copy specs, run one factory | Concept validation | Phase 0 |
| 2 | Gallery Coordinator | `multiclaude dark-factory create` command, 3 variants, Tier 1 judges | One-command multi-variant | Phase 1 validated |
| 3 | Feedback Loop | Spec versioning, extraction agent, divergence planner, dispose-rebuild | Full iterative cycle | Phase 2 |
| 4 | AI Judges Panel | Tier 2 judges, synthesis reports, calibration | Pre-filtered gallery | Phase 3 |
| 5 | Full Autonomy | Scheduled runs, auto-refinement proposals, budget monitoring, notifications | Hands-off operation | Phase 4 |

### Phase 0 Specification

```yaml
# Story file addition
provenance:
  autonomy_level: L2
  created_by: worker/brave-otter
  factory_id: null
```

```bash
# Commit trailer
git commit -m "feat: add task sorting (Story 42.1)" \
  --trailer "Provenance: L2/ai-supervised/worker-brave-otter"

# PR labels
gh pr create --label "provenance:ai-supervised" ...
```

### Validation Experiment (Phase 1)

1. Take an existing unimplemented story
2. Create one dark factory repo manually
3. Let it run autonomously (L3) for 4 hours
4. Compare output to normal multiclaude worker
5. Measure: time-to-completion, spec conformance, code quality, test coverage

### Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Approach | 6-phase incremental | Big-bang; Phase 0 only | Incremental validates each step; big-bang risks wrong thing |
| Validation gate | Each phase proves value before next | Fixed roadmap | Experimental — should stop if hypothesis fails |
| Phase 0 scope | Provenance fields + trailers + PR labels | Full schema; Nothing | Full is premature; nothing = no tracking |
| PoC method | Dark factory vs normal worker comparison | Theory only; Full gallery first | Empirical data beats speculation |

---

## Round 8: Integration with multiclaude

### Architecture Extension

```
multiclaude
├── repo management      # existing
├── agent management     # existing
├── worktree management  # existing
├── dark-factory         # NEW
│   ├── create
│   ├── status
│   ├── gallery
│   ├── judge
│   ├── dispose
│   └── feedback
└── messaging            # existing
```

### Integration Principles

- Shares multiclaude agent spawning with relaxed config
- Separate `factory-state.json` (not merged into `state.json`)
- Repo-prefixed agent names: `df-alpha/brave-otter`
- Factory worktrees don't need daemon refresh (autonomous, not rebasing)

### Resource Management

- Per-factory token budget with graceful shutdown
- 3-5 repos x 3-5 agents = 9-25 simultaneous sessions (manage API limits)
- On-demand refresh only (no 5-minute daemon cycle for factories)

### Human UX

All CLI-native:
- `multiclaude dark-factory try alpha-001` — launch variant binary
- `multiclaude dark-factory feedback alpha-001 "message"` — record preferences
- `multiclaude dark-factory compare alpha-001 beta-001` — side-by-side diff

### Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Integration | Extend multiclaude with module | Separate tool; Fork | Separate fragments UX; forking duplicates infra |
| State | Separate `factory-state.json` | Merge into `state.json` | Modular; doesn't bloat core |
| Agent management | Reuse spawning with relaxed config | Independent system | Reuse infra; relaxation via CLAUDE.md |
| Human UX | CLI-native commands | Web UI; IDE plugin | Matches existing multiclaude + ThreeDoors philosophy |

---

## Round 9: Edge Cases & Failure Modes

### Failure Mode Analysis

| ID | Failure Mode | Cause | Mitigation | Recovery |
|----|-------------|-------|------------|----------|
| F1 | Factory produces nothing usable | Vague spec, extreme divergence | Tier 1 early detection, time-bound | Analyze why — spec or factory? |
| F2 | All variants converge | Insufficient divergence | Convergence detection | Accept (valid finding) or redesign prompts |
| F3 | Inconsistent human feedback | Changing preferences | Preference drift tracking + alerts | Reconciliation session |
| F4 | Code leaks to golden repo | Copy-paste, misconfigured remote | Branch protection, separate identities | Revert, create proper re-entry PR |
| F5 | Budget exceeded | Complex spec, inefficient agents | Hard cap + 20% buffer | Auto-stop, evaluate partial results |
| F6 | Spec regression | Misinterpreted feedback | Immutable versions with rollback | Roll back to previous spec version |
| F7 | AI judge blind spots | Incomplete rubric | Calibration against human preferences | Update rubrics from override patterns |
| F8 | Factory repo compromised | Supply chain attack | Minimal permissions, dep pinning, SBOM | Delete factory, alert human, security review |

### Key Safety Insight

**Most failure modes are recoverable because the dark factory is disposable by design.** Worst case: wasted compute. No data loss (specs in golden repo), no production impact (no production access), no golden repo corruption (branch protection).

### Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Safety guarantee | Disposability | Transaction rollback; Undo | Simpler, more reliable |
| Spec regression | Immutable versions + rollback | Mutable with tracking | Clean rollback vs unrecoverable drift |
| Budget overrun | Auto-stop + partial evaluation | Manual; Unlimited | Prevents runaway; partial still valuable |
| Contamination detection | Branch protection + identities + similarity scanning | Trust only; Block all re-entry | Trust leaks; blocking too restrictive |

---

## Round 10: Final Synthesis & Recommendations

### Core Recommendations

1. **The Dark Factory Model is viable and worth pursuing incrementally.** Validate the core hypothesis: does spec refinement through disposal produce better software than code iteration?

2. **Start with Phase 0 (Provenance Tagging) immediately.** Zero-risk, immediately valuable, prerequisite for everything else.

3. **Validate with single dark factory PoC before gallery infrastructure.** One story, one factory, one comparison.

4. **Golden repo / dark factory separation is non-negotiable.** Separate repos, credentials, governance. No code flows without human approval + provenance tagging.

5. **Gallery model: 3-5 variants, 3-5 generations, controlled divergence.** Right scale for exploration without evaluation overload.

6. **Code is the most disposable artifact.** Paradigm shift: specs and learnings are primary assets. Code is a regenerable side effect.

7. **AI judges are accelerators, not requirements.** Start without them (Phases 0-2). Add when gallery has enough variants for pre-filtering.

8. **Extend multiclaude, don't replace it.** `multiclaude dark-factory` commands, separate state, relaxed CLAUDE.md templates.

9. **Immutable spec versioning is the backbone.** Every generation against frozen snapshot. Feedback produces new versions.

10. **Green/Yellow/Red governance with technical enforcement.** GitHub App permissions, CI scoping, budget hard-caps.

11. **Two-tier evaluation: automated CI + qualitative AI judges.** CI gives 80% of value.

12. **Dark factory testing is acceptance testing, not TDD.** Evaluate at boundary, not during execution.

13. **Terminal-native, friction-minimal human UX.** Gallery via CLI. Try variants by launching binaries. Annotate with text.

### Open Questions for Human Decision

| # | Question | Options | Recommendation |
|---|----------|---------|----------------|
| OQ-1 | Should dark factory repos be public or private? | Public / Private | **Private** — disposable code shouldn't be public |
| OQ-2 | Who can trigger a dark factory run? | Anyone / Owner / Configurable | **Configurable** with owner default |
| OQ-3 | Should output be preserved after disposal? | Full / Metadata / Nothing | **Metadata only** |
| OQ-4 | Maximum budget per factory run? | Fixed / Configurable / Unlimited | **Configurable, $50/run default** |
| OQ-5 | Phase 0 provenance mandatory or opt-in? | Mandatory all / Opt-in | **Mandatory for AI, opt-in for human** |

### Rejected Alternatives (Cross-Round)

| Alternative | Why Rejected |
|------------|-------------|
| Dark factory as golden repo branch | Shares CI/secrets, invites contamination |
| Dark factory as fork | Implicit upstream links invite contamination |
| Binary human/AI classification | Too coarse — misses L1-L4 nuance |
| No disposal (archive everything) | Accumulates baggage, defeats fresh-start principle |
| Five separate AI judge agents | Coordination overhead, diminishing returns vs single multi-lens agent |
| Full golden repo governance in factories | Kills the exploration that dark factories exist for |
| Web UI for gallery | Fragments UX, doesn't match terminal-native project philosophy |
| Auto-merge factory output to golden | Removes human judgment from the only irreversible action |
| Unlimited factory budget | Runaway cost risk with no safety net |
| Mutable specs | Loses reproducibility, risks unrecoverable drift |
