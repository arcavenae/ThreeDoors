---
name: 'reconcile-docs'
description: 'Scan story files and compare against planning docs (epic-list.md, epics-and-stories.md, ROADMAP.md). Report staleness and optionally auto-fix.'
---

# Reconcile Planning Documents

**Goal:** Detect and fix drift between story file statuses and the three planning documents.

## Step 1: Gather Story File Statuses

Scan all `docs/stories/*.story.md` files. For each file, extract:
- Story number (from filename)
- Status line (Done, Not Started, Draft, In Progress, etc.)
- PR number (if status includes one)

Group stories by epic number.

## Step 2: Compare Against Planning Docs

### epic-list.md (`docs/prd/epic-list.md`)
For each epic mentioned in epic-list.md:
- Compare the listed story count against actual story files
- Compare the listed status (Complete, Partial, Not Started) against actual story statuses
- Flag any epic where the listed status doesn't match reality

### epics-and-stories.md (`docs/prd/epics-and-stories.md`)
- Check that all epics with story files are represented
- Compare listed statuses against story file statuses

### ROADMAP.md
- Check that epic progress counts are accurate
- Flag any epics listed as "Not Started" that have completed stories

## Step 3: Report

Output a table of discrepancies:

```
| Document | Epic | Listed Status | Actual Status | Action Needed |
```

## Step 4: Auto-Fix (if user confirms)

If the user says "fix" or "auto-fix":
1. Update epic-list.md with correct statuses and story counts
2. Update ROADMAP.md with correct epic progress
3. Update the total count summary line in epic-list.md
4. Stage changes and report what was updated

If the user does NOT confirm, just report the discrepancies — do not modify files.

## Important

- Story files are the source of truth — never modify story file statuses
- Only update planning docs to match story file reality
- If an epic exists in story files but not in planning docs, flag it but do not auto-create entries (that requires PM review)
