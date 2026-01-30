---
allowed-tools: Bash(git:*), Bash(gh:*)
description: Create a pull request with auto-generated description
---

# Pull Request Command

Create a GitHub Pull Request with automatically generated description.

## Steps

1. Get current branch: `git branch --show-current`
2. Get base branch (usually main/master): `git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@'`
3. Get commits in this branch: `git log origin/<base>..HEAD --oneline`
4. Get diff stats: `git diff origin/<base>..HEAD --stat`
5. Generate PR description

## PR Template

```markdown
## Summary

[Brief description of changes - 2-3 sentences]

## Changes

- [List main changes as bullet points]

## Type of Change

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)
- [ ] Documentation update

## Testing

- [ ] Tests pass locally
- [ ] New tests added for new functionality

## Screenshots (if applicable)

[Add screenshots here]
```

## Create PR

Use GitHub CLI:

```bash
gh pr create --title "<title>" --body "<body>"
```

## Rules

- Title should be concise and descriptive
- Body should explain WHY not just WHAT
- Link related issues with "Closes #XX" or "Fixes #XX"
- Add reviewers if known

Ask user to confirm title and description before creating PR.
