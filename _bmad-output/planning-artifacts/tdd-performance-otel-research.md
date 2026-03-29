# TDD, Performance Management & OTEL for Agentic Dark Factories

**Date:** 2026-03-29
**Type:** Research Artifact
**Researcher:** Worker (gentle-otter)

---

## Table of Contents

1. [Prior Art Survey](#1-prior-art-survey)
2. [TDD for Agent Changes](#2-tdd-for-agent-changes)
3. [Performance Measurement & A/B Testing](#3-performance-measurement--ab-testing)
4. [OTEL Instrumentation for Dark Factories](#4-otel-instrumentation-for-dark-factories)
5. [Unified Data Model](#5-unified-data-model)
6. [Integration Architecture](#6-how-tdd--performance--otel-compose)
7. [Feasibility: Marvel vs ThreeDoors NOW](#7-feasibility-marvel-vs-threedoors-now)
8. [Open Questions for Human Decision](#8-open-questions-for-human-decision)

---

## 1. Prior Art Survey

### 1.1 TDD / Agent Testing Frameworks

| Tool / Paper | What It Does | Key Insight | Link |
|---|---|---|---|
| **Scenario (LangWatch)** | Domain-driven TDD for agents. User Simulator + System Under Test + Judge Agent run scenarios as tests. | Red-green-refactor for agents: write scenario criteria first (red), implement agent behavior (green), refactor prompts/models while scenarios hold. Non-deterministic outputs evaluated by judge LLM against business criteria, not string matching. | [langwatch.ai](https://langwatch.ai/blog/from-scenario-to-finished-how-to-test-ai-agents-with-domain-driven-tdd) |
| **TDAD (arxiv, March 2026)** | Test-Driven Agentic Development. Graph-based impact analysis maps code→test dependencies so agents know which tests to verify after changes. | **Context > procedure.** Prescriptive TDD instructions (write test first) actually *worsened* agent performance (+63% regressions). Giving agents a dependency map of *which tests matter* reduced regressions 72% (6.08% → 1.82%). | [arxiv.org/html/2603.17973](https://arxiv.org/html/2603.17973) |
| **Promptfoo** | CLI + YAML-based prompt/agent evaluation. Red-teaming, regression testing, CI integration. Acquired by OpenAI March 2026, still MIT. | Declarative test configs: define inputs, expected behaviors, scoring functions. Runs in CI. Best for prompt regression testing, not full agent workflow testing. | [github.com/promptfoo/promptfoo](https://github.com/promptfoo/promptfoo) |
| **Braintrust** | Platform for dataset management, evaluation scoring, experiment tracking, CI-based release gates. $80M raise Feb 2026. | Unified eval + monitoring. Connects experiment results to production telemetry. Strongest TypeScript support. | [braintrust.dev](https://www.braintrust.dev) |
| **LangSmith** | LangChain's eval platform. Captures full agent trajectory (steps, tool calls, reasoning). Evaluators score intermediate decisions. | Deep integration with LangChain ecosystem. Best for Python teams using LangChain agents. | [langchain.com/evaluation](https://www.langchain.com/evaluation) |
| **SWE-bench** | Benchmark: resolve real GitHub issues. Variants: Lite (300), Verified (human-validated), Multimodal. | Industry standard for coding agent evaluation. Correlated with production performance more than MMLU. | [llm-stats.com/benchmarks/swe-bench-verified](https://llm-stats.com/benchmarks/swe-bench-verified) |
| **AgentBench** | Multi-domain agent eval: OS, database, web, games, puzzles. 8 environments. | Tests general agent capabilities, not just coding. Good for measuring reasoning and tool use. | [arxiv.org/abs/2308.03688](https://arxiv.org/abs/2308.03688) |
| **Holdout Scenarios (Dark Factory pattern)** | Plain-English acceptance tests hidden from the coding agent. Isolated evaluator deploys, tests, LLM-judges. 2-of-3 pass, 90% gate. | Train/test separation for agents: agent never sees the validation criteria. Prevents overfitting to tests. | [hackernoon.com](https://hackernoon.com/the-dark-factory-pattern-moving-from-ai-assisted-to-fully-autonomous-coding) |

### 1.2 Performance / Experiment Tracking

| Tool | What It Does | Relevance |
|---|---|---|
| **MLflow** | Open-source ML lifecycle: experiment tracking, model registry, deployment. 30M+ monthly downloads. | Agent configs are analogous to ML models — version them, compare runs, promote winners. Native tracing for LLM agents. |
| **Weights & Biases** | Experiment visualization, hyperparameter sweeps, artifact tracking. | Best interactive visualization. Many teams use alongside MLflow (W&B for viz, MLflow for registry). |
| **LMSYS Chatbot Arena** | Crowdsourced blind comparison of LLM responses. Elo ratings. | Proves the arena concept works for comparing AI system outputs at scale. |
| **Neptune.ai** | Was third major player. Acquired by OpenAI, shut down SaaS March 2026. | Market consolidation — MLflow and W&B are the survivors. |

### 1.3 OTEL for AI/LLM

| Resource | What It Covers | Status |
|---|---|---|
| **OTEL GenAI Semantic Conventions** | Standard schema for LLM spans, metrics, events. Agent spans (create_agent, invoke_agent, execute_tool). | **Experimental** (Development status). Active SIG. |
| **OpenLLMetry** | Community OTEL instrumentation for GenAI providers. | Bridges gap while official semconv stabilizes. |
| **Datadog OTEL GenAI** | Native support for OTEL GenAI semconv v1.37+. | First major vendor adoption of the standard. |
| **Claude Code Metrics Pipeline** | OTEL → Collector → Prometheus → Grafana for Claude Code telemetry. | Directly applicable — already works with Claude Code. |

---

## 2. TDD for Agent Changes

### 2.1 The Problem

Agent definitions (CLAUDE.md, agent prompts in `agents/*.md`, hooks, settings) are the "source code" of agent behavior. But unlike Go code, there's no `go test` for agent definitions. When you change a prompt, you don't know if the agent got better or worse until something breaks in production.

### 2.2 Red-Green-Refactor for Agent Definitions

Adapting the Scenario framework's approach to multiclaude:

**Red Phase — Write the behavioral spec first:**
```yaml
# agent-tests/merge-queue/handles-ci-failure.scenario.yaml
name: "merge-queue handles CI failure"
description: "When a PR's CI fails, merge-queue should NOT merge and should notify supervisor"
setup:
  - create_pr: {title: "test PR", ci_status: "failure"}
judge_criteria:
  - "Agent does NOT merge the PR"
  - "Agent sends a message to supervisor about the CI failure"
  - "Agent does NOT retry more than 2 times"
max_turns: 20
required_pass_rate: 0.67  # 2 of 3 runs
```

**Green Phase — Implement/modify the agent definition to pass.**

**Refactor Phase — Optimize prompts, restructure instructions, while scenarios hold.**

### 2.3 Agent Arena Architecture

```
┌─────────────────────────────────────────────┐
│              Agent Arena Runner              │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ Config A  │  │ Config B  │  │ Config C  │  │
│  │ (baseline)│  │ (variant) │  │ (variant) │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
│       │              │              │        │
│  ┌────▼──────────────▼──────────────▼────┐  │
│  │        Benchmark Task Suite           │  │
│  │  (known tasks with expected outcomes) │  │
│  └────┬──────────────┬──────────────┬────┘  │
│       │              │              │        │
│  ┌────▼─────┐  ┌─────▼────┐  ┌─────▼────┐  │
│  │ Run A.1  │  │ Run B.1  │  │ Run C.1  │  │
│  │ Run A.2  │  │ Run B.2  │  │ Run C.2  │  │
│  │ Run A.3  │  │ Run B.3  │  │ Run C.3  │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
│       │              │              │        │
│  ┌────▼──────────────▼──────────────▼────┐  │
│  │          Judge Panel                  │  │
│  │  (LLM judges + deterministic checks) │  │
│  └────┬──────────────────────────────────┘  │
│       │                                     │
│  ┌────▼─────────────────────────────────┐   │
│  │      Results + Statistical Analysis  │   │
│  │  (pass rates, token costs, quality)  │   │
│  └──────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

### 2.4 Key Design Decisions

**TDAD's crucial finding applies here:** Don't write prescriptive agent workflow instructions. Instead, give agents *information* (what to check, what matters) and let them figure out the execution path. This means:

1. **Scenario-based, not assertion-based** — Judge agents evaluate outcomes against criteria, not exact output matching
2. **Holdout validation** — Some scenarios are hidden from the agent during development (prevents teaching-to-the-test)
3. **Statistical, not binary** — Run each scenario N times (rec: 3-5), require M/N passes (rec: 2/3 or 3/5)
4. **Multi-dimensional scoring** — Don't just measure "did it work?" Measure: task completion, token efficiency, code quality, time-to-completion, error recovery

### 2.5 Proposed Scenario Categories for multiclaude

| Category | What It Tests | Example |
|---|---|---|
| **Core competence** | Does the agent do its primary job? | Worker implements story, passes tests, creates PR |
| **Boundary behavior** | Does the agent stay in scope? | Worker encounters out-of-scope issue, notes it in PR instead of fixing |
| **Error recovery** | Does the agent handle failures? | Worker's `just lint` fails, agent fixes and retries |
| **Interaction protocol** | Does the agent communicate correctly? | Worker escalates ambiguous AC to supervisor instead of guessing |
| **Safety guardrails** | Does the agent respect forbidden actions? | Worker doesn't push to main, doesn't modify CODEOWNERS |
| **Resource efficiency** | Does the agent use resources wisely? | Worker completes task within token budget, doesn't retry infinitely |

---

## 3. Performance Measurement & A/B Testing

### 3.1 Metrics Framework

Drawing from MLOps experiment tracking and dark factory measurement patterns:

**Tier 1: Outcome Metrics (did it work?)**
| Metric | Definition | How to Measure |
|---|---|---|
| Task completion rate | % of assigned tasks that result in a merged PR | `merged_prs / assigned_tasks` |
| Acceptance criteria pass rate | % of ACs met per story | Judge evaluation of PR against story ACs |
| PR acceptance rate | % of PRs that merge without revision | `first_attempt_merges / total_prs` |
| Code quality score | Lint warnings, test coverage, complexity | `golangci-lint`, `go test -cover`, cyclomatic complexity |

**Tier 2: Efficiency Metrics (how well?)**
| Metric | Definition | How to Measure |
|---|---|---|
| Token consumption | Total input + output tokens per task | OTEL traces (see Section 4) |
| Time-to-merge | Wall clock from task assignment to PR merge | Git timestamps |
| Token efficiency | Useful output tokens / total tokens | Productive code + tests vs retries + errors + overhead |
| Cost per task | Dollar cost including all retries and overhead | Token count × model pricing |

**Tier 3: Process Metrics (how healthy?)**
| Metric | Definition | How to Measure |
|---|---|---|
| Churn rate | % of code changed then changed again before merge | Git diff analysis |
| Retry rate | Tool calls that failed and were retried | OTEL spans (error → retry pattern) |
| Hook violation rate | % of tool calls blocked by safety hooks | Hook execution logs |
| Scope creep incidents | Changes outside assigned story scope | PR diff vs story file scope |

### 3.2 A/B Testing Framework

Adapting MLOps experiment tracking for agent configurations:

```
┌─────────────────────────────────────────┐
│          Experiment Registry            │
│  (versioned agent configs + results)    │
├─────────────────────────────────────────┤
│                                         │
│  Experiment: worker-prompt-v2           │
│  ├─ Baseline: agents/worker.md@v1.0    │
│  ├─ Variant:  agents/worker.md@v2.0    │
│  ├─ Tasks: [story-1, story-2, ..., N]  │
│  ├─ Runs per config: 3                 │
│  └─ Metrics: completion, tokens, time  │
│                                         │
│  Results:                               │
│  ├─ Baseline: 87% complete, 45K tokens │
│  ├─ Variant:  92% complete, 38K tokens │
│  └─ Δ: +5% completion, -16% tokens    │
│         p-value: 0.03 (significant)    │
└─────────────────────────────────────────┘
```

**Statistical rigor requirements:**
- Minimum **10-20 runs per config** for meaningful comparison (per-task variance is high with LLMs)
- Use paired comparisons (same task, different configs) to control for task difficulty
- Report confidence intervals, not just point estimates
- For subtle improvements, may need 50+ runs — budget accordingly

### 3.3 Dark Factory of Factories (Meta-Orchestrator)

The meta-level: evaluating which *factory configuration* produces better results.

```
Factory Config A:              Factory Config B:
- 3 workers                    - 5 workers
- Opus for planning            - Sonnet for planning
- TDD-first workflow           - Implement-first workflow
- Strict scope hooks           - Relaxed scope hooks
        │                              │
        ▼                              ▼
  ┌──────────┐                  ┌──────────┐
  │ Factory A │                  │ Factory B │
  │ Run 1..N  │                  │ Run 1..N  │
  └─────┬────┘                  └─────┬────┘
        │                              │
        ▼                              ▼
  ┌──────────────────────────────────────┐
  │     Gallery Comparison Panel         │
  │  (outcome metrics + cost analysis)   │
  └──────────────────────────────────────┘
```

This aligns with the existing R-003 gallery model (3-5 variants × 3-5 generations) but adds:
- **Configuration versioning** — treat the factory config as the experiment variable
- **Cross-factory metrics** — compare not just code output but total cost, time, and quality
- **Promotion workflow** — winning configs get promoted (like ML model registry staging → production)

---

## 4. OTEL Instrumentation for Dark Factories

### 4.1 OTEL Signal Mapping

Each measurement need maps to the right OTEL signal type:

| What to Measure | OTEL Signal | Why This Signal |
|---|---|---|
| Agent lifecycle (task start → PR merge) | **Traces** (spans) | Hierarchical timing, parent-child relationships |
| Token consumption per call | **Metrics** (histogram) | Aggregatable, dashboardable, alertable |
| Tool usage frequency | **Metrics** (counter) | Simple counts, group by tool name |
| Error details, hook violations | **Events** (span events) | Rich context attached to the relevant span |
| Git activity (commits, PRs) | **Metrics** (counter) | Aggregatable over time |
| Cost tracking | **Metrics** (histogram) | Derivable from token metrics + pricing |
| Agent definition overhead | **Metrics** (gauge) | System prompt token count, changes infrequently |

### 4.2 Span Hierarchy for Dark Factory

Using OTEL GenAI semantic conventions (experimental, but the standard):

```
factory_run (root span)
├── agent.create_agent "worker-alpha" (gen_ai.agent.name)
│   └── gen_ai.usage.input_tokens, gen_ai.usage.output_tokens
├── agent.invoke_agent "worker-alpha"
│   ├── phase.research (custom span)
│   │   ├── tool.execute "Read" (file reads)
│   │   ├── tool.execute "Grep" (searches)
│   │   └── tool.execute "WebSearch" (external)
│   ├── phase.implementation (custom span)
│   │   ├── tool.execute "Write" (file creation)
│   │   ├── tool.execute "Edit" (file modification)
│   │   └── tool.execute "Bash" (compilation, tests)
│   ├── phase.testing (custom span)
│   │   ├── tool.execute "Bash" {"command": "just test"}
│   │   └── tool.execute "Bash" {"command": "just lint"}
│   ├── phase.pr_creation (custom span)
│   │   └── tool.execute "Bash" {"command": "gh pr create"}
│   └── phase.overhead (custom span)
│       ├── message.send (inter-agent communication)
│       ├── hook.violation (blocked tool calls)
│       └── retry (error recovery loops)
├── judge.evaluate (evaluation phase)
│   ├── judge.scenario_1 (per-scenario evaluation)
│   └── judge.scenario_2
└── factory_run.result (outcome metrics)
```

### 4.3 Custom Metrics for Dark Factories

Beyond the standard OTEL GenAI metrics (`gen_ai.client.token.usage`, `gen_ai.client.operation.duration`):

```
# Token breakdown by phase
dark_factory.phase.token_usage{phase="research|implementation|testing|overhead", token_type="input|output"}

# Tool usage
dark_factory.tool.invocations{tool="Read|Edit|Write|Bash|Grep|Glob|WebSearch", status="success|error|blocked"}

# Git activity
dark_factory.git.operations{operation="commit|push|pr_create|pr_merge", agent="worker-name"}

# Cost tracking (derived metric)
dark_factory.cost.usd{agent="worker-name", phase="research|implementation|testing|overhead"}

# Churn
dark_factory.code.churn{type="lines_added|lines_removed|lines_changed_then_changed"}

# Hook violations
dark_factory.hook.violations{hook="git-safety|scope-check", agent="worker-name"}

# Agent definition overhead
dark_factory.agent.definition_tokens{agent="worker|merge-queue|supervisor"}

# Message queue overhead
dark_factory.messaging.tokens{direction="send|receive", type="heartbeat|task|escalation"}

# Research vs productive ratio
dark_factory.work.ratio{type="research|planning|implementation|testing|overhead"}
```

### 4.4 Collector Architecture

```
┌─────────────────────────────────────────────────┐
│                 Dark Factory Host                │
│                                                  │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────────────┐   │
│  │Worker│ │Worker│ │Worker│ │  Supervisor   │   │
│  │  A   │ │  B   │ │  C   │ │              │   │
│  └──┬───┘ └──┬───┘ └──┬───┘ └──────┬───────┘   │
│     │        │        │            │            │
│     └────────┴────────┴────────────┘            │
│                    │                             │
│              OTLP (gRPC/HTTP)                   │
│                    │                             │
│  ┌─────────────────▼───────────────────────┐    │
│  │        Local OTEL Collector             │    │
│  │  ┌─────────┐ ┌──────────┐ ┌─────────┐  │    │
│  │  │Receivers│→│Processors│→│Exporters │  │    │
│  │  │  OTLP   │ │ Batch    │ │Prometheus│  │    │
│  │  │         │ │ Filter   │ │ OTLP     │  │    │
│  │  │         │ │ Resource │ │ File     │  │    │
│  │  └─────────┘ └──────────┘ └─────────┘  │    │
│  └─────────────────────────────────────────┘    │
│                    │                             │
└────────────────────┼─────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
  ┌─────▼──────┐          ┌──────▼──────┐
  │ Prometheus │          │  Grafana    │
  │ (metrics)  │◄─────────│ (dashboards)│
  └────────────┘          └─────────────┘
        │
  ┌─────▼──────┐
  │Grafana Tempo│  (optional, for traces)
  │ or Jaeger   │
  └────────────┘
```

**For multi-factory (factory-of-factories):**
Each factory host runs its own local collector. All collectors export to a central backend (Grafana Cloud, self-hosted Grafana stack, or Honeycomb). Factory-level resource attributes (`factory.id`, `factory.config.version`) enable cross-factory comparison.

### 4.5 Backend Recommendations

| Backend | Strengths | Weaknesses | Recommendation |
|---|---|---|---|
| **Prometheus + Grafana** (self-hosted) | Free, proven, full control, Claude Code already exports OTLP | Operational burden, local only by default | **PoC/ThreeDoors NOW** |
| **Grafana Cloud** (free tier) | 10K metrics, 50GB traces free. No ops. | Vendor dependency, data leaves machine | **Multi-factory production** |
| **Honeycomb** | Best trace exploration UX, high-cardinality queries | Expensive at scale, traces-focused (weaker metrics) | **If trace debugging is priority** |
| **Datadog** | Native OTEL GenAI semconv support (v1.37+), rich AI monitoring | Most expensive, heavy agent | **Enterprise with existing Datadog** |
| **Jaeger** (self-hosted) | Free, great trace visualization | Traces only, no metrics | **Supplement to Prometheus** |

### 4.6 Operator Dashboard Design

What an operator needs to see at a glance:

**Panel 1: Factory Health**
- Active factories and their current phase
- Overall pass rate (last 24h)
- Cost burn rate ($/hour)

**Panel 2: Agent Performance**
- Per-agent task completion rate
- Per-agent average token consumption
- Per-agent error/retry rate
- Agent definition token overhead (how much context each agent burns just existing)

**Panel 3: Cost Breakdown**
- Cost by phase (research vs implementation vs testing vs overhead)
- Cost by agent type
- Cost trend over time
- Cache hit rate (cacheRead efficiency — a "healthy chart has a fat cacheRead slice")

**Panel 4: Quality Gate**
- Scenario pass rates by category
- PR acceptance rate
- Hook violation trend
- Scope creep incidents

**Panel 5: Efficiency**
- Research:productive work ratio
- Message queue token overhead
- Housekeeping (git maintenance, sync) token cost
- Time-to-merge distribution

---

## 5. Unified Data Model

### 5.1 The Connection: OTEL → Performance → TDD

```
                    ┌─────────────────────┐
                    │   OTEL Telemetry    │
                    │  (traces, metrics,  │
                    │   events, logs)     │
                    └────────┬────────────┘
                             │
                    ┌────────▼────────────┐
                    │  Performance Store  │
                    │  (Prometheus/Grafana │
                    │   or experiment DB) │
                    └────────┬────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼──────┐ ┌────▼───────┐ ┌───▼──────────┐
     │  A/B Testing  │ │  Scenario  │ │  Dashboards  │
     │  Framework    │ │  Evaluator │ │  & Alerts    │
     │ (compare      │ │ (TDD judge │ │ (operator    │
     │  configs)     │ │  panel)    │ │  visibility) │
     └───────────────┘ └────────────┘ └──────────────┘
```

### 5.2 Core Entities

```yaml
# Factory Run
factory_run:
  id: "run-2026-03-29-001"
  config_version: "worker-v2.1"
  factory_config:
    agent_definitions: {worker: "sha256:abc...", supervisor: "sha256:def..."}
    model: "claude-opus-4-6"
    hooks: ["git-safety.sh@v1.2"]
    settings: {max_workers: 3, tdd_mode: true}
  tasks: ["story-40.10", "story-39.12"]
  start_time: "2026-03-29T10:00:00Z"
  end_time: "2026-03-29T11:45:00Z"

# Agent Run (child of factory run)
agent_run:
  id: "agent-run-001"
  factory_run_id: "run-2026-03-29-001"
  agent_type: "worker"
  agent_definition_hash: "sha256:abc..."
  task: "story-40.10"
  outcome: "success"
  pr_number: 393
  metrics:
    total_tokens: 45000
    input_tokens: 32000
    output_tokens: 13000
    cache_read_tokens: 18000
    cost_usd: 0.67
    duration_seconds: 1800
    tool_calls: 47
    tool_errors: 3
    hook_violations: 1
    retries: 2

# Experiment
experiment:
  id: "exp-worker-tdd-mode"
  baseline: "worker-v1.0"
  variant: "worker-v2.0"
  tasks: ["story-1", "story-2", ..., "story-20"]
  runs_per_config: 3
  results:
    baseline: {completion: 0.87, avg_tokens: 45000, avg_cost: 0.67}
    variant: {completion: 0.92, avg_tokens: 38000, avg_cost: 0.57}
    p_value: 0.03
    recommendation: "promote variant"

# Scenario Result
scenario_result:
  id: "scenario-result-001"
  agent_run_id: "agent-run-001"
  scenario: "handles-ci-failure"
  runs: 3
  passes: 2
  pass_rate: 0.67
  judge_scores:
    - {criteria: "does not merge", score: 1.0}
    - {criteria: "notifies supervisor", score: 0.67}
    - {criteria: "limits retries", score: 1.0}
```

### 5.3 How Data Flows

1. **Agent executes** → OTEL spans and metrics emitted automatically (token usage, tool calls, timing)
2. **OTEL Collector** → processes, batches, exports to Prometheus/Grafana
3. **Scenario evaluator** → runs after agent completes, queries OTEL data + judges output
4. **Experiment tracker** → aggregates scenario results across runs, computes statistics
5. **Dashboard** → visualizes all of the above for human operators
6. **A/B framework** → compares experiment results, recommends promotions

---

## 6. How TDD + Performance + OTEL Compose

### 6.1 Integration Diagram

```
 ┌──────────────────────────────────────────────────────────────┐
 │                    Agent Definition Change                    │
 │              (e.g., modify agents/worker.md)                 │
 └────────────────────────┬─────────────────────────────────────┘
                          │
                          ▼
 ┌──────────────────────────────────────────────────────────────┐
 │                   1. TDD Scenario Suite                       │
 │  Run scenarios against new definition. Judge panel evaluates. │
 │  Gate: scenarios must pass at ≥ baseline rate.               │
 │  Output: scenario pass/fail results + judge scores           │
 └────────────────────────┬─────────────────────────────────────┘
                          │ (if scenarios pass)
                          ▼
 ┌──────────────────────────────────────────────────────────────┐
 │                2. A/B Experiment Run                          │
 │  Run N tasks through both old and new config.                │
 │  OTEL collects traces + metrics from both.                   │
 │  Output: paired comparison data (tokens, time, quality)      │
 └────────────────────────┬─────────────────────────────────────┘
                          │ (if metrics improve or hold)
                          ▼
 ┌──────────────────────────────────────────────────────────────┐
 │              3. Statistical Analysis                          │
 │  Compare outcome metrics. Test for significance.             │
 │  Output: recommendation (promote / reject / need more data)  │
 └────────────────────────┬─────────────────────────────────────┘
                          │ (if promote)
                          ▼
 ┌──────────────────────────────────────────────────────────────┐
 │              4. Promotion to Production                       │
 │  New agent definition becomes the baseline.                  │
 │  Old definition archived in experiment registry.             │
 │  OTEL continues monitoring for regression.                   │
 └──────────────────────────────────────────────────────────────┘
```

### 6.2 The Feedback Loop

- **OTEL → Performance**: Telemetry data feeds into performance metrics automatically
- **Performance → A/B**: Performance metrics power statistical comparison of configs
- **A/B → TDD**: A/B results inform which scenarios need to be added/updated
- **TDD → Agent Defs**: Validated changes get promoted; failed ones get rejected with data
- **Agent Defs → OTEL**: New definitions generate new telemetry, closing the loop

---

## 7. Feasibility: Marvel vs ThreeDoors NOW

### 7.1 What We Can Prototype in ThreeDoors NOW

| Component | Effort | What It Looks Like | Prerequisite |
|---|---|---|---|
| **OTEL metrics collection** | **Low** (days) | Enable Claude Code's built-in OTLP export (`CLAUDE_CODE_ENABLE_TELEMETRY=1`). Local Docker: Collector + Prometheus + Grafana. | Docker on dev machine |
| **Custom dark factory metrics** | **Medium** (1-2 weeks) | Shell script wrapper around `multiclaude work` that parses JSONL transcripts post-hoc and pushes metrics to Prometheus via pushgateway. Not real-time but captures token usage, tool calls, timing. | OTEL stack running |
| **Scenario test suite** | **Medium** (1-2 weeks) | YAML scenario files in `agent-tests/`. Runner script spawns agent in isolated worktree, feeds task, evaluates output. Judge is a separate Claude call scoring against criteria. | Agent arena runner script |
| **Holdout validation** | **Low** (days) | Separate `agent-tests/holdout/` directory. Gitignored from agent worktrees (agents can't see them). Evaluator runs them post-completion. | Scenario framework |
| **Experiment tracking** | **Low** (days) | JSONL files tracking config version → results. Simple `jq` analysis. Upgrade to MLflow later. | Scenario framework |
| **Basic A/B comparison** | **Medium** (1-2 weeks) | Script: run same task through config A and config B, collect metrics, compute paired differences. | OTEL + scenarios |
| **Cost dashboard** | **Low** (days) | Grafana dashboard from Claude Code's native metrics. Panels: token usage, cost, cache efficiency. | OTEL stack |

**Total for ThreeDoors PoC: ~4-6 weeks of focused effort.**

### 7.2 What Needs Marvel

| Component | Why It Needs Marvel | What Marvel Adds |
|---|---|---|
| **Multi-factory comparison** | Running multiple dark factories in parallel across repos | Factory-of-factories orchestration, central OTEL backend, cross-factory dashboards |
| **Agent config registry** | Versioning and promoting agent definitions across projects | Centralized experiment tracking, promotion workflow, rollback capability |
| **Statistical significance engine** | Rigorous A/B testing with proper power analysis | Sample size calculation, sequential testing, automated promotion decisions |
| **Cross-factory OTEL** | Comparing telemetry from factories running different configs | Central collector, factory-level resource attributes, cross-cutting queries |
| **Automated arena** | Scheduled benchmark runs, regression detection, alerting | CI integration, scheduled runs, Slack notifications, trend analysis |
| **Judge panel calibration** | Ensuring LLM judges are consistent and accurate | Inter-rater reliability metrics, judge prompt versioning, calibration datasets |

### 7.3 Migration Path

```
Phase 0 (NOW): Claude Code OTLP → local Grafana stack
Phase 1 (ThreeDoors): Scenario framework + basic experiment tracking
Phase 2 (ThreeDoors): A/B comparison scripts + holdout validation
Phase 3 (Marvel): Multi-factory OTEL + central experiment registry
Phase 4 (Marvel): Automated arena + statistical engine + promotion workflow
```

---

## 8. Open Questions for Human Decision

| ID | Question | Options | Recommendation |
|---|---|---|---|
| Q-TDD-1 | Should scenario tests be visible to agents (transparent TDD) or hidden (holdout validation)? | A) All visible — agents know what they're tested on. B) Mix — core scenarios visible, holdout scenarios hidden. C) All hidden — pure black-box evaluation. | **B) Mix.** Core scenarios guide agent behavior (like unit tests guide code). Holdout scenarios prevent overfitting (like held-out test sets in ML). |
| Q-TDD-2 | What judge model for scenario evaluation? Same model as agent, different model, or multi-model panel? | A) Same model. B) Different model (e.g., Sonnet judging Opus). C) Multi-model panel. | **B) Different model** for PoC (cheaper, avoids self-evaluation bias). **C) Multi-model panel** for production (R-003 already proposed two-tier judges). |
| Q-PERF-1 | Minimum sample size for A/B tests — how many runs per config before declaring a winner? | A) 5 runs. B) 10 runs. C) 20 runs. D) Adaptive (start with 10, add more if borderline). | **D) Adaptive.** Start with 10, use sequential testing to stop early if effect is large or continue if ambiguous. Balances cost vs confidence. |
| Q-PERF-2 | Should experiment tracking start with JSONL files or go straight to MLflow? | A) JSONL + jq scripts (simpler, local). B) MLflow (richer, more overhead). | **A) JSONL for PoC.** MLflow is excellent but premature for 5-10 experiments. Move to MLflow when we have 50+ tracked experiments or need team collaboration. |
| Q-OTEL-1 | Should we instrument multiclaude itself or post-process JSONL transcripts? | A) Real-time: modify multiclaude Go code to emit OTEL spans. B) Post-hoc: parse JSONL transcripts after completion and push metrics. C) Hybrid: Claude Code native OTLP + post-hoc transcript parsing. | **C) Hybrid.** Claude Code already exports OTLP for token/cost metrics. Post-hoc parsing fills in custom metrics (phase breakdown, tool classification). Real-time multiclaude instrumentation is ideal but requires Go development — save for Marvel. |
| Q-OTEL-2 | Backend choice for ThreeDoors PoC? | A) Prometheus + Grafana (Docker Compose). B) Grafana Cloud free tier. C) Honeycomb free tier. | **A) Local Prometheus + Grafana.** Zero cost, full control, proven with Claude Code. Matches the Claude Code metrics dashboard pattern already documented. |

---

## Appendix A: Key Sources

- [Scenario — Domain-Driven TDD for Agents (LangWatch)](https://langwatch.ai/blog/from-scenario-to-finished-how-to-test-ai-agents-with-domain-driven-tdd)
- [TDAD: Test-Driven Agentic Development (arxiv, March 2026)](https://arxiv.org/html/2603.17973)
- [OTEL GenAI Agent Spans Specification](https://opentelemetry.io/docs/specs/semconv/gen-ai/gen-ai-agent-spans/)
- [OTEL GenAI Metrics Specification](https://opentelemetry.io/docs/specs/semconv/gen-ai/gen-ai-metrics/)
- [AI Agent Observability (OTEL Blog)](https://opentelemetry.io/blog/2025/ai-agent-observability/)
- [OTEL GenAI Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/)
- [Dark Factory Pattern — L1-L4 Architecture (HackerNoon)](https://hackernoon.com/the-dark-factory-pattern-moving-from-ai-assisted-to-fully-autonomous-coding)
- [Claude Code Metrics Dashboard (Sealos)](https://sealos.io/blog/claude-code-metrics/)
- [Token Usage + Cost Tracking with OTEL (OneUptime)](https://oneuptime.com/blog/post/2026-02-06-track-token-usage-prompt-costs-model-latency-opentelemetry/view)
- [Promptfoo — Agent Eval Framework](https://github.com/promptfoo/promptfoo)
- [Braintrust — LLM Eval Platform](https://www.braintrust.dev)
- [LangSmith — Agent Evaluation](https://www.langchain.com/evaluation)
- [SWE-bench Verified Leaderboard](https://llm-stats.com/benchmarks/swe-bench-verified)
- [MLflow — Experiment Tracking + Model Registry](https://mlflow.org)
- [Dark Factory Architecture — L4 (Signals/aktagon)](https://signals.aktagon.com/articles/2026/03/dark-factory-architecture-how-level-4-actually-works/)
- [AI Agent Performance Measurement (Microsoft)](https://www.microsoft.com/en-us/dynamics-365/blog/it-professional/2026/02/04/ai-agent-performance-measurement/)
- [Best AI Agent Evaluation Benchmarks 2025 (o-mega)](https://o-mega.ai/articles/the-best-ai-agent-evals-and-benchmarks-full-2025-guide)
- [Top 5 Agent Evaluation Tools 2026 (Maxim)](https://www.getmaxim.ai/articles/top-5-tools-for-agent-evaluation-in-2026/)
- [OTEL Standardizes LLM Tracing (Dev|Journal, March 2026)](https://earezki.com/ai-news/2026-03-21-opentelemetry-just-standardized-llm-tracing-heres-what-it-actually-looks-like-in-code/)
