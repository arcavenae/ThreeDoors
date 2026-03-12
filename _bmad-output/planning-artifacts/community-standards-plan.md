# Community Standards Plan — ThreeDoors

> **Status: Planned — see Stories 0.42 (Done, PR #512) and 0.43 (Not Started)**
> PR #498 created stories. Story 0.42 (Issue & PR Templates) is Done. Story 0.43 (Community Standards Documentation) is Not Started.

> Planning artifact for adding GitHub community standards files.
> Generated: 2026-03-10

## Current State

**GitHub Community Health Score: 42%**

Source: `gh api repos/arcaven/ThreeDoors/community/profile`

### What Exists

| File | Status |
|------|--------|
| README.md | Present — comprehensive with badges, install instructions, screenshots |
| LICENSE | Present — MIT License (Copyright 2025 arcaven) |
| CODE_OF_CONDUCT.md | **Missing** |
| CONTRIBUTING.md | **Missing** |
| SECURITY.md | **Missing** |
| Issue templates | **Missing** — no `.github/ISSUE_TEMPLATE/` directory |
| PR template | **Missing** — no `.github/PULL_REQUEST_TEMPLATE.md` |
| SUPPORT.md | **Missing** |

### What Exists in `.github/`

- `dependabot.yml` — GitHub Actions version pinning (monthly)
- `workflows/ci.yml` — Full CI pipeline (lint, test, coverage, benchmarks, Docker E2E)
- `workflows/release.yml` — Release pipeline (build, sign, notarize, Homebrew)
- `workflows/release-verify.yml` — Release verification

---

## Files to Create

### 1. CODE_OF_CONDUCT.md (root)

**Recommendation:** Contributor Covenant v2.1 (industry standard, GitHub recognizes it automatically).

**Location:** `/CODE_OF_CONDUCT.md`

**Content:** Standard Contributor Covenant 2.1 text with:
- Contact method: GitHub Issues (label `conduct`) or email TBD from project owner
- Enforcement: Project maintainers (currently solo — arcaven)
- No modifications to the core text needed

**Project-specific notes:**
- ThreeDoors is a solo-human-directed AI-agent project. The CoC should apply to human contributors and issue reporters. Agent-generated content is reviewed by the human maintainer.
- Keep it standard — don't over-customize. The Contributor Covenant is recognized by GitHub's community profile checker.

### 2. CONTRIBUTING.md (root)

**Recommendation:** Comprehensive guide covering the unique development model.

**Location:** `/CONTRIBUTING.md`

**Content outline:**

```
# Contributing to ThreeDoors

## Development Model
- Solo maintainer + AI agent team (BMAD methodology)
- Story-driven development — all work requires a story file
- PRs are reviewed by the maintainer before merge

## Getting Started
- Prerequisites: Go 1.25.4+, make, gofumpt, golangci-lint
- Clone & build: `git clone ... && make build`
- Run tests: `make test`
- Run linter: `make lint`
- Format code: `make fmt`

## How to Contribute

### Reporting Bugs
- Use the Bug Report issue template
- Include: Go version, OS, steps to reproduce, expected vs actual

### Suggesting Features
- Use the Feature Request issue template
- Note: ThreeDoors is opinionated — read SOUL.md before proposing features
- "Three doors, not three hundred" — features that add complexity need strong justification

### Submitting Code
1. Fork the repo and create a feature branch
2. Ensure a story file exists in `docs/stories/` (or create one)
3. Write tests (table-driven, stdlib `testing` only — no testify)
4. Run the full quality gate: `make fmt && make lint && make test`
5. Run race detector: `go test -race ./...`
6. Create a PR using the PR template

### Code Standards
- See CLAUDE.md "Go Quality Rules" section for comprehensive standards
- gofumpt formatting (not gofmt)
- golangci-lint with zero warnings
- Table-driven tests, t.Helper(), t.Parallel()
- Error wrapping with %w, context.Context first parameter
- Atomic file writes for persistence

## What NOT to Contribute
- Features that conflict with SOUL.md philosophy
- Dependencies that add significant weight (prefer stdlib)
- Telemetry, analytics, or phone-home features (local-first, privacy-always)
- Testify or other test framework dependencies

## Architecture Overview
- `cmd/threedoors/` — Entry point
- `internal/tasks/` — Task domain (models, providers, persistence)
- `internal/tui/` — Bubbletea views and UI components
- TaskProvider interface for storage backends

## License
MIT — contributions are made under the same license.
```

### 3. SECURITY.md (root)

**Recommendation:** Standard security policy with responsible disclosure.

**Location:** `/SECURITY.md`

**Content outline:**

```
# Security Policy

## Supported Versions
- Only the latest release on `main` is supported
- Alpha releases (threedoors-a) are not covered by security support

## Reporting a Vulnerability
- DO NOT open a public GitHub issue for security vulnerabilities
- Use GitHub's private vulnerability reporting:
  Security tab → "Report a vulnerability"
- Or email: [TBD — project owner to provide]
- Expected response time: within 7 days
- We will coordinate disclosure timing with the reporter

## Scope
- ThreeDoors is a local-first application — no network services by default
- Task data is stored locally (YAML files, JSONL session logs)
- Integration adapters (Jira, GitHub, Apple Notes) use user-provided tokens
- No telemetry or data collection

## Security Design Principles
- Local-first: data never leaves the user's machine unless explicitly configured
- No cloud sync by default
- Integration tokens stored locally, never transmitted to third parties
- Atomic file writes to prevent data corruption
```

**Note:** GitHub has built-in private vulnerability reporting. Recommend enabling it in repo Settings → Security → "Private vulnerability reporting".

### 4. Issue Templates (.github/ISSUE_TEMPLATE/)

**Location:** `.github/ISSUE_TEMPLATE/`

**Format:** YAML form format (not markdown) — provides structured input fields.

#### 4a. Bug Report (`bug-report.yml`)

```yaml
name: Bug Report
description: Report a bug in ThreeDoors
title: "[Bug]: "
labels: ["bug", "triage"]
body:
  - type: markdown
    attributes:
      value: |
        Thanks for reporting a bug! Please fill out the details below.
  - type: textarea
    id: description
    attributes:
      label: What happened?
      description: A clear description of the bug.
      placeholder: Describe the bug...
    validations:
      required: true
  - type: textarea
    id: expected
    attributes:
      label: Expected behavior
      description: What did you expect to happen?
    validations:
      required: true
  - type: textarea
    id: reproduce
    attributes:
      label: Steps to reproduce
      description: How can we reproduce this bug?
      placeholder: |
        1. Open ThreeDoors with `threedoors`
        2. Press '...'
        3. See error
    validations:
      required: true
  - type: dropdown
    id: interface
    attributes:
      label: Interface
      description: Which interface were you using?
      options:
        - TUI (interactive terminal)
        - CLI (headless commands)
        - MCP Server
    validations:
      required: true
  - type: input
    id: version
    attributes:
      label: ThreeDoors version
      description: Run `threedoors --version` to find this.
      placeholder: "0.1.0-alpha.20260310.abc1234"
    validations:
      required: true
  - type: dropdown
    id: os
    attributes:
      label: Operating System
      options:
        - macOS (Apple Silicon)
        - macOS (Intel)
        - Linux
        - Other
    validations:
      required: true
  - type: input
    id: go-version
    attributes:
      label: Go version (if building from source)
      placeholder: "1.25.4"
  - type: textarea
    id: logs
    attributes:
      label: Relevant log output
      description: Paste any error messages or logs.
      render: shell
  - type: textarea
    id: task-file
    attributes:
      label: Task file (if relevant)
      description: Paste a minimal task file that reproduces the issue (remove personal data).
      render: yaml
```

#### 4b. Feature Request (`feature-request.yml`)

```yaml
name: Feature Request
description: Suggest a new feature or enhancement
title: "[Feature]: "
labels: ["enhancement", "triage"]
body:
  - type: markdown
    attributes:
      value: |
        Before requesting a feature, please read [SOUL.md](https://github.com/arcaven/ThreeDoors/blob/main/SOUL.md) to understand ThreeDoors' design philosophy. Features that increase complexity or conflict with "three doors, not three hundred" will likely be declined.
  - type: textarea
    id: problem
    attributes:
      label: Problem or motivation
      description: What problem does this feature solve? Why do you need it?
      placeholder: "I'm always frustrated when..."
    validations:
      required: true
  - type: textarea
    id: solution
    attributes:
      label: Proposed solution
      description: How would you like this to work?
    validations:
      required: true
  - type: textarea
    id: alternatives
    attributes:
      label: Alternatives considered
      description: What other approaches did you consider?
  - type: dropdown
    id: area
    attributes:
      label: Area
      description: Which part of ThreeDoors does this affect?
      options:
        - TUI (interactive terminal)
        - CLI (headless commands)
        - MCP Server
        - Task providers / integrations
        - Data / analytics
        - Documentation
        - Other
    validations:
      required: true
  - type: checkboxes
    id: soul-check
    attributes:
      label: Philosophy check
      options:
        - label: This feature reduces friction (doesn't add steps for the user)
          required: false
        - label: This feature respects local-first, privacy-always principles
          required: false
        - label: I've read SOUL.md and believe this aligns with the project's philosophy
          required: true
```

#### 4c. Question (`question.yml`)

```yaml
name: Question
description: Ask a question about ThreeDoors
title: "[Question]: "
labels: ["question"]
body:
  - type: markdown
    attributes:
      value: |
        Have a question? Check the [README](https://github.com/arcaven/ThreeDoors/blob/main/README.md) and [docs/](https://github.com/arcaven/ThreeDoors/tree/main/docs) first.
  - type: textarea
    id: question
    attributes:
      label: Your question
      description: What would you like to know?
    validations:
      required: true
  - type: dropdown
    id: topic
    attributes:
      label: Topic
      options:
        - Installation / setup
        - Configuration
        - Task providers / integrations
        - TUI usage
        - CLI usage
        - MCP server
        - Contributing / development
        - Other
```

#### 4d. Config file (`config.yml`)

Disables blank issue creation, directing users to templates:

```yaml
blank_issues_enabled: false
contact_links:
  - name: Documentation
    url: https://github.com/arcaven/ThreeDoors/tree/main/docs
    about: Check the docs before opening an issue
```

### 5. PR Template (.github/PULL_REQUEST_TEMPLATE.md)

**Location:** `.github/PULL_REQUEST_TEMPLATE.md`

**Content — project-specific, references story-driven development:**

```markdown
## Summary

<!-- What does this PR do? Link the story file. -->

**Story:** `docs/stories/X.Y.story.md`

## Changes

<!-- Bullet list of what changed and why. -->

-

## Checklist

- [ ] Story file exists and acceptance criteria are met
- [ ] Story file updated with `Status: Done (PR #NNN)`
- [ ] Tests added/updated (table-driven, stdlib `testing`)
- [ ] `make fmt` — code is formatted with gofumpt
- [ ] `make lint` — zero warnings from golangci-lint
- [ ] `make test` — all tests pass
- [ ] `go test -race ./...` — no race conditions (required for `internal/tui/` and `internal/cli/` changes)
- [ ] No `//nolint` directives without justifying comments
- [ ] ROADMAP.md updated (if completing an epic)
- [ ] Commit messages follow format: `feat|fix|docs: description (Story X.Y)`

## Testing

<!-- How was this tested? What test cases were added? -->

## Opportunities

<!-- Anything noticed that's out of scope but worth noting for future work. -->
```

### 6. SUPPORT.md (root)

**Recommendation:** Brief file pointing to existing resources.

**Location:** `/SUPPORT.md`

**Content:**

```markdown
# Getting Help with ThreeDoors

## Documentation

- [README](README.md) — Overview, installation, and quick start
- [docs/](docs/) — Architecture, stories, and design decisions
- [SOUL.md](SOUL.md) — Project philosophy and design principles

## Asking Questions

- Open a [Question issue](https://github.com/arcaven/ThreeDoors/issues/new?template=question.yml) on GitHub

## Reporting Bugs

- Open a [Bug Report](https://github.com/arcaven/ThreeDoors/issues/new?template=bug-report.yml) on GitHub
- See [SECURITY.md](SECURITY.md) for security vulnerabilities (do NOT use public issues)

## Feature Requests

- Open a [Feature Request](https://github.com/arcaven/ThreeDoors/issues/new?template=feature-request.yml) on GitHub
- Please read [SOUL.md](SOUL.md) first — ThreeDoors is opinionated by design
```

---

## Implementation Plan

### Recommended Order

1. **Issue templates** (`.github/ISSUE_TEMPLATE/*.yml` + `config.yml`) — Immediate value, structures incoming issues
2. **PR template** (`.github/PULL_REQUEST_TEMPLATE.md`) — Enforces story-driven checklist on every PR
3. **CONTRIBUTING.md** — Guides potential contributors
4. **CODE_OF_CONDUCT.md** — Standard Contributor Covenant 2.1
5. **SECURITY.md** — Responsible disclosure policy
6. **SUPPORT.md** — Points to existing resources

### Story Recommendation

This should be a single story (e.g., Story 0.38 or similar infrastructure story) since all files are documentation-only and can ship in one PR. No code changes, no test changes — just new markdown and YAML files.

### Labels Needed

The issue templates reference these labels (verify they exist in the repo):
- `bug` — likely exists
- `enhancement` — likely exists
- `triage` — may need creation
- `question` — may need creation

### Expected Health Score After Implementation

Current: 42% → Expected: **100%** (all community standards files present)

### Owner Decision Points

1. **Contact email for CoC and SECURITY.md** — arcaven needs to provide a contact email, or decide to use GitHub-only channels (Issues for CoC, private vulnerability reporting for security)
2. **Enable private vulnerability reporting** — Settings → Security → check "Private vulnerability reporting"
3. **Label audit** — Verify `triage` and `question` labels exist or create them
4. **CoC enforcement scope** — Standard Contributor Covenant is fine for a solo-maintained project; no committee needed

---

## Project-Specific Customizations

These are the key ways ThreeDoors' community standards differ from a generic open-source project:

1. **Story-driven development** — PR template requires story file reference and status update
2. **SOUL.md philosophy gate** — Feature request template requires acknowledgment of design philosophy
3. **AI agent development model** — CONTRIBUTING.md should explain the solo-human + agent-team model without requiring contributors to use it
4. **No testify** — Contributing guide explicitly calls out stdlib-only testing
5. **gofumpt not gofmt** — Formatting standard is stricter than Go default
6. **Race detector mandatory** — For TUI/CLI changes, called out in PR checklist
7. **Local-first privacy** — Security policy scope is simpler since there are no network services by default
8. **BMAD pipeline** — Not mentioned in contributor-facing docs (internal process), but story-driven development is the public-facing manifestation
