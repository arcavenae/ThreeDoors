# Branch Protection Recommendations for `main`

## Status

Branch protection is **not currently configured** on `main`. The API returned 404 when queried (no admin access from fork). These are recommended settings for the repo admin to apply.

## Recommended Settings

### Required Status Checks

| Check | Required | Notes |
|-------|----------|-------|
| `Quality Gate` | Yes | Must pass before merge — runs go vet, golangci-lint, tests, build |

### Pull Request Reviews

| Setting | Value | Rationale |
|---------|-------|-----------|
| Require approvals | 1 | At least one reviewer for code quality |
| Dismiss stale reviews | Yes | Re-review after force-push |
| Require review from code owners | No | No CODEOWNERS file configured |

### Branch Restrictions

| Setting | Value | Rationale |
|---------|-------|-----------|
| Restrict direct pushes to `main` | Yes | All changes via PR |
| Allow force pushes | No | Protect commit history |
| Allow deletions | No | Prevent accidental branch deletion |
| Require linear history | Recommended | Cleaner git log, enforces rebase workflow |
| Require signed commits | Recommended | Verify commit authorship |

### Additional Settings

| Setting | Value | Rationale |
|---------|-------|-----------|
| Include administrators | Yes | Admins follow same rules |
| Require conversation resolution | Recommended | Ensure review comments are addressed |

## How to Apply (Admin)

Using GitHub CLI:

```bash
gh api repos/arcaven/ThreeDoors/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["Quality Gate"]}' \
  --field enforce_admins=true \
  --field required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true}' \
  --field restrictions=null \
  --field allow_force_pushes=false \
  --field allow_deletions=false \
  --field required_linear_history=true
```

## How to Verify (Admin)

```bash
gh api repos/arcaven/ThreeDoors/branches/main/protection --jq '{
  status_checks: .required_status_checks.contexts,
  reviews_required: .required_pull_request_reviews.required_approving_review_count,
  enforce_admins: .enforce_admins.enabled,
  allow_force_pushes: .allow_force_pushes.enabled,
  allow_deletions: .allow_deletions.enabled,
  linear_history: .required_linear_history.enabled
}'
```

Expected output:

```json
{
  "status_checks": ["Quality Gate"],
  "reviews_required": 1,
  "enforce_admins": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "linear_history": true
}
```
