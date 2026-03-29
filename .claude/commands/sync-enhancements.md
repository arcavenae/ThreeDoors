# Sync Enhancements — Push reusable multiclaude improvements to the enhancements repo

Sync multiclaude customizations from ThreeDoors to the `arcavenae/multiclaude-enhancements` private repo.

**Input:** $ARGUMENTS — optional: specific files or "all" to sync everything. Default: sync all changed files.

## What Gets Synced

These files contain reusable multiclaude improvements extracted from ThreeDoors:

- `agents/*.md` — Agent definition templates (pr-shepherd, merge-queue, envoy, project-watchdog, arch-watchdog)
- `.claude/commands/plan-work.md` — Research-to-stories pipeline skill
- `.claude/commands/implement-story.md` — Story implementation skill
- `.claude/commands/course-correct.md` — Course correction skill
- `.claude/commands/reconcile-docs.md` — Doc reconciliation skill
- `docs/operations/*.md` — Operational runbooks and incident reports

## Steps

1. Check the enhancements repo exists:
   ```bash
   ls multiclaude-enhancements/ || git clone git@github.com:arcavenae/multiclaude-enhancements.git multiclaude-enhancements/
   ```

2. Sync files from ThreeDoors to the enhancements repo:
   ```bash
   # Copy agent definitions
   cp agents/*.md multiclaude-enhancements/agents/

   # Copy relevant skills
   cp .claude/commands/plan-work.md multiclaude-enhancements/skills/
   cp .claude/commands/implement-story.md multiclaude-enhancements/skills/
   cp .claude/commands/course-correct.md multiclaude-enhancements/skills/
   cp .claude/commands/reconcile-docs.md multiclaude-enhancements/skills/
   ```

3. Commit and push:
   ```bash
   cd multiclaude-enhancements
   git add -A
   git commit -S -m "sync: update from ThreeDoors $(date +%Y-%m-%d)"
   git push origin main
   ```

## Notes

- The enhancements repo is PRIVATE — never make it public
- Only sync files that are generic/reusable — strip ThreeDoors-specific references
- Agent definitions may reference ThreeDoors-specific paths — consider parameterizing
