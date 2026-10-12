---
name: git-flow
description: Use when the user asks to commit, open a PR, push, merge or close out a change ("commitealo", "abrí el PR", "subilo", "generá dos commits en un branch aparte", "mergeá", "cerrá el PR"). Applies the user's and the project's conventions (branch from the right base, conventional commits, no Co-Authored-By trailer, PR language of the repo, confirmation before push/merge, post-merge verification). Replaces the old /commit and /pr commands.
---

# Git Flow

From working-tree changes to a merged PR, with the conventions the user keeps
correcting by hand. Confirm before anything that leaves the machine.

## 1. Read the project's conventions first

Check, in this order, and apply what you find:

- Repo `CLAUDE.md` / `.claude/skills/repo-conventions` (e.g. usagi-shorts: base branch
  `develop`, PRs in Spanish, one feature at a time).
- `CONTRIBUTING.md`, `.github/PULL_REQUEST_TEMPLATE.md`, commitlint/commitizen config.
- `git log -20 --format=%s` to match the existing message style and language.
- Default branch: `gh repo view --json defaultBranchRef -q .defaultBranchRef.name`
  (fallback `main`). Ruleset/protection: assume the default branch is protected and
  always go through a PR unless the user says otherwise.

User-level rules (apply unless the project says otherwise):

- **No `Co-Authored-By` trailer** in commits.
- Commit messages: Conventional Commits (`type(scope): subject`), subject in the
  language of the repo's recent commits, body explains *why*.
- PR title/body in the repo's language. Spanish projects → Spanish PR.
- Never force-push, never rewrite pushed history, never `git add -A` blindly over
  untracked files the user did not mention.

## 2. Plan the commits

1. `git status --short` and `git diff --stat`. List untracked files explicitly and ask
   before adding anything that looks generated, local or secret (`.env`, `*.local.*`,
   build output, large binaries).
2. Group changes into **logical commits** (one concern each). If the user asked for a
   specific number of commits ("dos commits"), respect it.
3. If currently on the default branch: create a branch `type/short-slug` from the right
   base (`develop` if the project uses it, else default). If already on a feature
   branch, stay.
4. Show the plan: branch, commits (message + files). **Confirm** if anything is
   ambiguous; otherwise proceed to commit (commits are local and reversible).

## 3. Commit

For each group: `git add <files>` then `git commit -m "<message>"`. Verify with
`git log --oneline -n <k>`. No trailers unless the project requires them.

## 4. Push and PR — confirm first

Ask once: "¿Pusheo `<branch>` y abro el PR contra `<base>`?" Then:

```bash
git push -u origin <branch>
gh pr create --base <base> --title "<title>" --body "<body>"
```

PR body: summary (what/why, 3-6 bullets), test plan (checkboxes, what was actually
run), notes for the reviewer (migration steps, follow-ups). Language per step 1.
Reply with the PR URL only.

If this repo needs a credential helper (e.g. `git -c credential.helper='!gh auth
git-credential' push`), use it; do not ask the user to log in again if `gh auth
status` is OK.

## 5. Merge (only when asked) and verify

- Stacked PRs: merge in dependency order; re-check each one after the previous merge.
- Never merge with `--admin` or bypass rules on your own; if the ruleset blocks it, say
  so and let the user merge from the web UI.
- After the merge: `git checkout <base> && git pull`, confirm the merge commit contains
  the expected changes (`git log --oneline -3`, `git diff <before>..<after> --stat`),
  delete the local branch, and report. If the project tracks features/flags that flip
  only after merge (e.g. `passes: true`), do that now and commit it separately.

## Guardrails

- Confirmation is mandatory before: push, PR creation, merge, branch deletion on the
  remote, anything with `--force`.
- Never include secrets, tokens or personal data in commit messages or PR bodies.
- If `gh` is unavailable, give the user the exact commands instead of improvising with
  the web UI.
