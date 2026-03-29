# Dark Factory Control Pattern (DFCP) — Permissions & Authority Gates Research

**Date:** 2026-03-29
**Companion to:** [Dark Factory Research](dark-factory-research.md) (R-003)
**Scope:** GitHub permissions, token scopes, branch protection, CODEOWNERS, authority gates, and the formal DFCP pattern definition

---

## Table of Contents

1. [Dark Factory Repo Permissions (Full Autonomy)](#1-dark-factory-repo-permissions-full-autonomy)
2. [Human-Gated Golden Repo Permissions (Controlled)](#2-human-gated-golden-repo-permissions-controlled)
3. [Design Decision Gates](#3-design-decision-gates)
4. [DFCP Pattern Definition](#4-dfcp-pattern-definition)
5. [Implementation for arcavenae Org](#5-implementation-for-arcavenae-org)
6. [Decisions](#decisions)
7. [Open Questions](#open-questions)

---

## 1. Dark Factory Repo Permissions (Full Autonomy)

### Token Strategy: GitHub App vs Fine-Grained PAT

| Credential Type | Pros | Cons | Recommendation |
|----------------|------|------|----------------|
| **GitHub App** | Repo-scoped install, renewable tokens, org-wide management, webhook support | Must generate installation tokens (short-lived), more setup | **Recommended for production** |
| **Fine-Grained PAT** | Simpler setup, repo-scoped, granular permissions | Tied to user account, expires, no webhook support | **Acceptable for Phase 1 PoC** |
| **Classic PAT** | Simple, well-understood | Overly broad, all-or-nothing repo access | **Rejected** — too broad for dark factory isolation |

### Required Token Scopes (Dark Factory)

For AI agents to operate autonomously in a dark factory repo, the following permissions are needed:

| Permission | Access Level | Why |
|-----------|-------------|-----|
| **Contents** | Read & Write | Push code, create branches, read files |
| **Pull requests** | Read & Write | Create PRs, comment, manage labels |
| **Workflows** | Read & Write | **Critical:** Modify `.github/workflows/` files. Without this, pushes containing workflow changes are rejected with "refusing to allow...to create or update workflow" |
| **Actions** | Read & Write | Trigger workflow runs, read CI results |
| **Issues** | Read & Write | Create/manage issues for tracking |
| **Metadata** | Read | Required baseline for all API operations |
| **Checks** | Read & Write | Create/read status checks |
| **Commit statuses** | Read & Write | Report CI status to PRs |

**Critical finding:** The `workflows` permission is the #1 blocker for AI autonomy. GitHub enforces this even for tag pushes if the commit touches `.github/workflows/`. This is the exact issue that blocks our current merge-queue from merging workflow PRs (see MEMORY.md OAuth limitation).

### Branch Protection for Dark Factory Repos

Dark factory repos need branch protection that enforces quality without requiring human approval:

```yaml
# Recommended ruleset for dark-factory repos
name: "factory-main-protection"
target: branch
conditions:
  ref_name:
    include: ["refs/heads/main"]
enforcement: active
rules:
  - type: pull_request
    parameters:
      required_approving_review_count: 0     # No human review required
      dismiss_stale_reviews_on_push: false
      require_code_owner_review: false        # No CODEOWNERS in dark factory
      require_last_push_approval: false
  - type: required_status_checks
    parameters:
      strict_required_status_checks_policy: true
      required_status_checks:
        - context: "CI"                       # Factory CI must pass
  - type: non_fast_forward                    # Prevent force-push
  - type: deletion                            # Prevent branch deletion
```

**Key design choice:** 0 required approvals + required CI status checks = AI can merge autonomously as long as CI passes. This is the "full autonomy" configuration.

### Auto-Merge Configuration

Dark factory repos should enable auto-merge so agents don't block on manual merge clicks:

```bash
# Enable auto-merge on the repo
gh api repos/arcavenae/{factory-repo}/settings --method PATCH \
  --field allow_auto_merge=true \
  --field delete_branch_on_merge=true
```

With 0 required approvals, auto-merge fires as soon as CI passes. The merge-queue agent (or equivalent) enables auto-merge on each PR via:

```bash
gh pr merge --auto --squash <PR_NUMBER>
```

### Secret Management

| Secret | Purpose | Provisioning |
|--------|---------|-------------|
| GitHub App private key / PAT | Push, PR, merge operations | Org-level secret, scoped to `*-df-*` repos |
| `GITHUB_TOKEN` (auto) | CI workflow operations | Automatic, no setup needed |
| No deployment secrets | Dark factories NEVER deploy | Intentional omission |
| No package registry creds | Dark factories NEVER publish | Intentional omission |

**Principle:** Dark factory repos get the minimum secrets needed for development and CI. Zero deployment credentials, zero registry credentials. This is a contamination prevention measure.

### CI/CD Pipeline Permissions

```yaml
# .github/workflows/ci.yml — dark factory template
permissions:
  contents: read
  checks: write
  pull-requests: write    # For CI status comments
  # NO deployments, NO packages, NO pages
```

Dark factory CI runs tests and reports results. It never deploys, publishes, or triggers external systems.

---

## 2. Human-Gated Golden Repo Permissions (Controlled)

### Files Requiring Human Approval

Based on ThreeDoors' governance model, these files/paths control the project and must be human-gated:

| Path Pattern | Gate Type | Rationale |
|-------------|-----------|-----------|
| `SOUL.md` | **Hard gate** — human-only | Project philosophy; defines AI behavior boundaries |
| `CLAUDE.md` | **Hard gate** — human-only | Agent instructions; modifying changes all AI behavior |
| `.claude/rules/*.md` | **Hard gate** — human-only | Agent rules; same risk as CLAUDE.md |
| `ROADMAP.md` | **Scope gate** — human review | Controls what gets worked on |
| `docs/prd/epic-list.md` | **Scope gate** — human review | Epic definitions and priorities |
| `docs/prd/epics-and-stories.md` | **Scope gate** — human review | Story breakdown and sequencing |
| `docs/decisions/BOARD.md` | **Design gate** — human review | Architectural decisions |
| `.github/workflows/*.yml` | **Infrastructure gate** — human review | CI/CD pipeline; security-sensitive |
| `.github/CODEOWNERS` | **Authority gate** — human-only | Controls the gate system itself |
| `agents/*.md` | **Authority gate** — human review | Agent behavior definitions |

### CODEOWNERS Template for Golden Repo

```
# CODEOWNERS — ThreeDoors Golden Repo
# Gate files that control project direction and AI behavior

# Hard gates (human-only, no AI modification)
SOUL.md                         @skippy
CLAUDE.md                       @skippy
.claude/                        @skippy

# Scope gates (human review required)
ROADMAP.md                      @skippy
docs/prd/epic-list.md          @skippy
docs/prd/epics-and-stories.md  @skippy

# Design gates (human review required)
docs/decisions/BOARD.md         @skippy

# Infrastructure gates (human review required)
.github/                        @skippy

# Authority gates (human review required)
agents/                         @skippy
.github/CODEOWNERS              @skippy

# Everything else — AI agents can self-merge via merge-queue
# (no CODEOWNERS entry = no required review)
```

### Branch Protection for Golden Repo

The golden repo needs a layered ruleset that treats different paths differently:

**Ruleset 1: Base protection (all PRs)**
```yaml
name: "main-protection"
rules:
  - type: pull_request
    parameters:
      required_approving_review_count: 0     # AI can merge non-gated files
      require_code_owner_review: true        # BUT files with CODEOWNERS need review
  - type: required_status_checks
    parameters:
      required_status_checks:
        - context: "Quality Gate"
        - context: "Docker E2E Tests"
        - context: "Performance Benchmarks"
  - type: required_signatures
  - type: non_fast_forward
  - type: deletion
```

**Key insight:** With `require_code_owner_review: true` and 0 required approvals, PRs touching CODEOWNERS-covered files need the owner's approval, but PRs touching only uncovered files can merge with just CI passing. This creates a two-tier system within a single ruleset.

**Ruleset 2: Path-based team review (GitHub rulesets feature)**

GitHub's newer ruleset feature allows requiring review from specific teams for specific file paths. This augments CODEOWNERS:

```yaml
name: "governance-file-protection"
rules:
  - type: file_path_restriction
    parameters:
      restricted_file_paths:
        - "SOUL.md"
        - "CLAUDE.md"
        - ".claude/**"
```

This prevents pushes that modify these files entirely — they MUST go through a PR.

### Human Override Mechanisms

| Scenario | Override Method | Audit |
|----------|---------------|-------|
| Emergency hotfix bypassing CODEOWNERS | Ruleset bypass list (org admin) | GitHub audit log |
| Unlocking a blocked PR | Owner approves CODEOWNERS review | PR review history |
| Temporarily relaxing a gate | Disable ruleset, make change, re-enable | Git blame + ruleset history |
| Overriding AI decision in BOARD.md | Direct commit on feature branch, human-reviewed PR | Standard PR review |

---

## 3. Design Decision Gates

### Gate Taxonomy

Gates are files that control project behavior through their content, enforced by a combination of branch protection, CODEOWNERS, and agent instructions.

```
┌──────────────────────────────────────────────────────┐
│                    AUTHORITY GATES                     │
│   WHO can do what (enforceable via GitHub)            │
│   CODEOWNERS, branch protection, rulesets             │
├──────────────────────────────────────────────────────┤
│                    SCOPE GATES                        │
│   WHAT gets worked on (enforced by agent instructions)│
│   ROADMAP.md, epic-list.md, epics-and-stories.md     │
├──────────────────────────────────────────────────────┤
│                    DESIGN GATES                       │
│   HOW things are built (enforced by agent instructions│
│   + decision record)                                  │
│   BOARD.md, architecture docs, SOUL.md, CLAUDE.md    │
├──────────────────────────────────────────────────────┤
│                    QUALITY GATES                      │
│   Whether it MEETS STANDARDS (enforced by CI)         │
│   Required status checks, linting, tests, coverage    │
└──────────────────────────────────────────────────────┘
```

### Enforcement Mechanisms Per Gate Type

| Gate Type | Mechanism | Enforcement Level |
|-----------|-----------|-------------------|
| **Authority** | CODEOWNERS + rulesets | **Technical** — GitHub blocks merges |
| **Scope** | CLAUDE.md instructions + merge-queue agent checks | **Behavioral** — agents self-enforce; merge-queue rejects out-of-scope PRs |
| **Design** | BOARD.md consultation + party mode requirement | **Behavioral** — agents check BOARD.md before implementing |
| **Quality** | Required CI status checks | **Technical** — GitHub blocks merges |

### Making Behavioral Gates More Enforceable

Currently, scope and design gates rely on agent instructions (CLAUDE.md). To strengthen enforcement:

1. **CI-based scope check:** A GitHub Action that verifies PRs reference a valid story file and the story's epic is listed in ROADMAP.md. Reject PRs that fail this check.

2. **BOARD.md conflict detection:** A CI check that verifies no PR contradicts an existing decided (D-*) entry in BOARD.md without explicitly referencing the override.

3. **Story file status validation:** CI check that verifies story file exists and has correct format before allowing merge.

These transform behavioral gates into technical gates, closing the enforcement gap.

---

## 4. DFCP Pattern Definition

### Formal Definition

The **Dark Factory Control Pattern (DFCP)** is a permission and authority model for managing repositories in an AI-driven software development organization. It defines two complementary profiles:

1. **Dark Factory Profile** — Maximum autonomy for AI agents in disposable repos
2. **Golden Repo Profile** — Human-gated governance for the canonical product repo

The pattern separates concerns: dark factories optimize for exploration speed, golden repos optimize for human control.

### Permission Matrix

| Dimension | Dark Factory | Golden Repo | Hybrid (Future) |
|-----------|-------------|-------------|-----------------|
| **Human review required** | Never | CODEOWNERS-gated paths | Configurable per path |
| **CI required** | Yes (basic) | Yes (full suite) | Yes (full suite) |
| **Auto-merge** | Enabled, 0 approvals | Enabled, CODEOWNERS-gated | Enabled, path-dependent |
| **Workflow modification** | Unrestricted (workflow scope) | Human-gated (.github/) | Human-gated |
| **CODEOWNERS** | None | Comprehensive | Selective |
| **Deployment secrets** | None | Available | None for AI paths |
| **Package publishing** | Forbidden | Human-gated | Forbidden for AI paths |
| **Scope enforcement** | Spec-only (no ROADMAP) | ROADMAP + epic-list + stories | ROADMAP (human paths only) |
| **Decision recording** | Optional | Mandatory (BOARD.md) | Mandatory |
| **Provenance tagging** | Automatic (L3/L4) | Required (all levels) | Required |
| **Branch protection** | CI-only gates | CI + CODEOWNERS + signatures | Full |
| **Token scopes** | contents, PRs, workflows, actions, checks | contents, PRs (no workflows for AI) | Scoped per path |

### Lifecycle: Dark Factory → Golden Repo Promotion

```
Dark Factory Output
        │
        ▼
┌─────────────────┐
│ Spec-Only Entry  │──── Default: insights feed spec refinements, no code crosses
│ (No code moves)  │
└─────────────────┘
        │
        ▼ (Explicit human decision)
┌─────────────────┐
│ Cherry-Pick PR   │──── Selected files via tagged PR: `provenance: dark-factory`
│ (Minimal code)   │     CODEOWNERS review applies. Elevated scrutiny.
└─────────────────┘
        │
        ▼ (Rare, high ceremony)
┌─────────────────┐
│ Wholesale Adopt  │──── Full module replacement. Architecture review.
│ (Module replace) │     Human sign-off + full test suite + provenance audit.
└─────────────────┘
```

**Audit trail at each stage:**
- PR label: `provenance:dark-factory`
- Commit trailer: `Provenance: L3/dark-factory/{factory-id}`
- PR description: link to factory repo, judge reports, spec version
- BOARD.md entry: decision to adopt with rationale

### DFCP Configuration File

Each repo declares its DFCP profile in a `.dfcp.yaml` at the repo root:

```yaml
# .dfcp.yaml — declares this repo's DFCP profile
dfcp:
  version: "1.0"
  profile: "dark-factory"  # or "golden", "hybrid"

  # Dark factory specific
  factory:
    golden_repo: "arcavenae/ThreeDoors"
    autonomy_level: "L3"
    max_duration: "4h"
    max_cost_usd: 50
    allow_workflow_modification: true
    allow_deployment: false
    allow_package_publish: false

  # Golden repo specific (only in golden profile)
  gates:
    authority:
      codeowners: true
      require_code_owner_review: true
    scope:
      ci_scope_check: true
      roadmap_enforcement: true
    design:
      board_conflict_check: true
    quality:
      required_checks:
        - "Quality Gate"
        - "Docker E2E Tests"
        - "Performance Benchmarks"

  # Provenance
  provenance:
    mandatory: true
    default_level: "L2"  # golden: L2 (AI-supervised), dark-factory: L3/L4
```

---

## 5. Implementation for arcavenae Org

### Current State Assessment

| Setting | Current Value | DFCP Recommendation |
|---------|--------------|---------------------|
| Repo visibility | Public | Golden: Public. Dark factories: Private (per OQ-1/P-008) |
| Branch protection | Rulesets: 0 approvals, CI required, signed commits | Good baseline. Add CODEOWNERS |
| Auto-merge | Enabled | Keep |
| Delete branch on merge | Enabled | Keep |
| CODEOWNERS | None | **Add immediately** |
| Workflow token scope | Limited (merge-queue can't merge workflow PRs) | **Upgrade for dark factories** |
| Rulesets | 1 active (main protection) | Add CODEOWNERS requirement |

### Concrete Steps

#### Phase A: Golden Repo Hardening (Do Now)

1. **Create `.github/CODEOWNERS`** with the template from Section 2
2. **Update existing ruleset** to add `require_code_owner_review: true`:
   ```bash
   gh api repos/arcavenae/ThreeDoors/rulesets/13640284 --method PUT \
     --input <(cat <<'EOF'
   {
     "name": "main protection",
     "target": "branch",
     "enforcement": "active",
     "conditions": {"ref_name": {"include": ["refs/heads/main"], "exclude": []}},
     "rules": [
       {
         "type": "pull_request",
         "parameters": {
           "required_approving_review_count": 0,
           "dismiss_stale_reviews_on_push": false,
           "require_code_owner_review": true,
           "require_last_push_approval": false,
           "required_review_thread_resolution": false
         }
       },
       {
         "type": "required_status_checks",
         "parameters": {
           "strict_required_status_checks_policy": true,
           "do_not_enforce_on_create": false,
           "required_status_checks": [
             {"context": "Docker E2E Tests", "integration_id": 15368},
             {"context": "Performance Benchmarks", "integration_id": 15368},
             {"context": "Quality Gate", "integration_id": 15368}
           ]
         }
       },
       {"type": "required_signatures"},
       {"type": "non_fast_forward"},
       {"type": "deletion"}
     ]
   }
   EOF
   )
   ```
3. **Create scope-check CI workflow** (validates story file reference in PRs)

#### Phase B: Dark Factory Token Setup (Phase 1 PoC)

1. **Create a Fine-Grained PAT** for the PoC:
   - Scoped to `arcavenae/*-df-*` repos (create repos first)
   - Permissions: contents (RW), pull-requests (RW), workflows (RW), actions (RW), checks (RW), metadata (R)
   - Expiration: 30 days (PoC duration)

2. **Create a GitHub App** for production (Phase 2+):
   - App name: `multiclaude-dark-factory`
   - Permissions: same as PAT above
   - Install on org, scoped to dark factory repos only
   - Benefit: installation tokens are short-lived (1 hour), auto-renewed

#### Phase C: Repo Template (Phase 1 PoC)

Create `arcavenae/.github-dark-factory-template` as a GitHub template repo:
- Pre-configured CI workflow
- Relaxed CLAUDE.md (no story requirements, no scope gates)
- `.dfcp.yaml` with dark-factory profile
- `.factory-manifest.yaml` template
- No CODEOWNERS
- No deployment workflows

### Token/App Comparison for arcavenae

| Factor | Fine-Grained PAT (Phase 1) | GitHub App (Phase 2+) |
|--------|---------------------------|----------------------|
| Setup effort | 5 minutes | 30 minutes |
| Token lifetime | 30-90 day expiry | Auto-renewed (1h tokens) |
| Scope | Per-repo selection | Org-wide install, repo-scoped |
| Audit trail | User-attributed | App-attributed (cleaner) |
| `workflows` scope | ✅ Supported | ✅ Supported (since 2020) |
| Multi-factory | New token per factory or shared | Single app, multiple installs |
| Recommendation | **PoC only** | **Production** |

---

## Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| D-DFCP-1: Credential type for dark factories | GitHub App for production; Fine-Grained PAT for Phase 1 PoC | Classic PAT (too broad); OAuth App (user-scoped, not right model) | GitHub App provides repo-scoped, auto-renewing, org-managed credentials. PAT acceptable for short-lived PoC |
| D-DFCP-2: Dark factory branch protection | 0 approvals + required CI + no CODEOWNERS | No protection (too risky even for disposable); human review (defeats purpose) | CI gates catch broken code; 0 approvals enables autonomy; non-fast-forward prevents history loss |
| D-DFCP-3: Golden repo authority model | CODEOWNERS + `require_code_owner_review` + existing CI gates | Path-based rulesets only (newer, less tested); CODEOWNERS without enforcement (leaky) | CODEOWNERS is mature, well-understood, and combined with the ruleset flag creates a two-tier system: gated files need human review, ungated files can auto-merge |
| D-DFCP-4: Gate taxonomy | Four types: Authority (technical), Scope (behavioral→CI), Design (behavioral), Quality (technical) | Single-tier (too coarse); Six-tier (over-engineered) | Four types map cleanly to enforcement mechanisms and cover all control surfaces |
| D-DFCP-5: Behavioral gate enforcement | CI checks that validate scope (story file) and design (BOARD.md) conformance | Agent instructions only (leaky); Full CI enforcement of all gates (too rigid for design gates) | CI scope checks close the biggest enforcement gap; design gates remain advisory with BOARD.md conflict detection as a warning |
| D-DFCP-6: Dark factory CI scope | Tests + lint only; no deploy, no publish, no external triggers | Full CI including deploy stages; No CI at all | Deploy/publish are contamination vectors; no CI produces unusable output; test+lint is the quality floor |
| D-DFCP-7: DFCP configuration file | `.dfcp.yaml` at repo root declaring profile | README convention; No declaration (implicit) | Machine-readable; enables CI automation; explicit > implicit for security-sensitive configuration |

### Rejected Alternatives (Cross-Section)

| Alternative | Why Rejected |
|------------|-------------|
| Classic PAT for dark factories | All-or-nothing repo access; can't scope to specific repos; no granular permissions |
| No branch protection in dark factories | Even disposable repos benefit from CI gates — broken code wastes factory runtime |
| Required human review in dark factories | Defeats the purpose of autonomous operation; human judgment applies at gallery/promotion stage |
| Single permission profile for all repos | Dark factories need autonomy; golden repos need gates. One size doesn't fit |
| CODEOWNERS without ruleset enforcement | Without `require_code_owner_review`, CODEOWNERS is informational only — agents would ignore it |
| Rulesets-only (no CODEOWNERS) | Rulesets can restrict paths but don't provide the ownership model. CODEOWNERS defines WHO, rulesets enforce it |
| Webhook-based gate enforcement | Over-engineered for current scale; CI checks are simpler and sufficient |

---

## Open Questions

| # | Question | Options | Recommendation |
|---|----------|---------|----------------|
| OQ-DFCP-1 | Should golden repo CODEOWNERS apply to story files (`docs/stories/`)? | Yes (human gates all stories) / No (workers can update status) | **No** — workers must update story status to `Done (PR #NNN)`. Gate only planning docs, not individual stories |
| OQ-DFCP-2 | Should the scope-check CI workflow block merge or just warn? | Block / Warn / Off | **Warn initially**, upgrade to block after validation period |
| OQ-DFCP-3 | Should dark factory repos inherit org-level rulesets or define their own? | Org-level / Per-repo / Template | **Template** — org-level rulesets would affect golden repos too; templates provide defaults that can be customized |
| OQ-DFCP-4 | When should the CODEOWNERS + ruleset update be applied to ThreeDoors? | Immediately / After Phase 0 / After Phase 1 | **Immediately** — golden repo hardening is independent of dark factory work and immediately valuable |
| OQ-DFCP-5 | Should `agents/*.md` changes require human review? | Yes / No | **Yes** — agent definitions control AI behavior, similar risk profile to CLAUDE.md |

---

## Sources

- [GitHub Apps: Workflow permission (2020)](https://github.blog/changelog/2020-04-07-github-apps-workflow-permission/)
- [About code owners - GitHub Docs](https://docs.github.com/articles/about-code-owners)
- [Available rules for rulesets - GitHub Docs](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets)
- [Automatically merging a pull request - GitHub Docs](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/automatically-merging-a-pull-request)
- [Managing auto-merge for pull requests - GitHub Docs](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-auto-merge-for-pull-requests-in-your-repository)
- [Permissions required for fine-grained personal access tokens - GitHub Docs](https://docs.github.com/en/rest/authentication/permissions-required-for-fine-grained-personal-access-tokens)
- [Differences between GitHub Apps and OAuth apps - GitHub Docs](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/differences-between-github-apps-and-oauth-apps)
- [Managing a branch protection rule - GitHub Docs](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/managing-a-branch-protection-rule)
- [Creating rulesets for a repository - GitHub Docs](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository)
- [Controlling permissions for GITHUB_TOKEN - GitHub Docs](https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/controlling-permissions-for-github_token)
