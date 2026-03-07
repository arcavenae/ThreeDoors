# Party Mode Consensus: GitHub Issues Integration (Epic 26)

**Date:** 2026-03-07
**Participants:** Winston (Architect), John (PM), Amelia (Dev), Quinn (QA), Bob (SM), Murat (Test Architect)

## Unanimous Decisions

1. **Epic 26** — GitHub Issues Integration, 4 stories following established adapter pattern
2. **Use `go-github` SDK** (google/go-github) — official, well-maintained, handles pagination/auth/rate-limiting natively
3. **Label-based conventions** for priority mapping (`priority:critical/high/medium/low` labels to Effort) and status enrichment (`in-progress` label maps to in-progress status)
4. **Explicit repo list** in config.yaml (`repos: ["owner/repo1", "owner/repo2"]`) — org-level queries deferred
5. **Read-only first** (Story 26.2 delivers 80% value), then bidirectional sync (Story 26.3)
6. **Schedule after Epic 25** (Todoist) to leverage learnings from first raw-HTTP adapter
7. **Source badge:** `[GH]`
8. **Package:** `internal/adapters/github/`
9. **Assignee filter:** Default `@me` — only show issues assigned to the authenticated user
10. **Mock strategy:** Interface-based mocking of go-github client for unit tests, httptest.NewServer for integration tests

## Key Technical Decisions

- `go-github` SDK adds a dependency — acceptable (Google-maintained, Apache-2.0, no transitive concerns)
- Reuse `RateLimitError` pattern from Jira adapter, wrapping go-github's native rate limit errors
- Contract test coverage target: 80%+ for github package
- Config: PAT token via `GITHUB_TOKEN` env var or `config.yaml` settings

## No Dissenting Opinions

All agents agree this is a well-understood, low-risk integration following proven patterns.
