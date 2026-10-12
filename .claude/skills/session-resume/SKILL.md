---
name: session-resume
description: Use at the start of a session when the user wants to pick up where they left off on a project ("por dónde nos quedamos", "repasá el estado", "retomemos", "qué sigue", "where were we"). Rebuilds context from the Obsidian vault configured on THIS machine (~/.claude/vault.json), git history, open PRs and plan files, then proposes the next step. Never assumes which vault or folder layout exists.
---

# Session Resume

Rebuild working context for the current project from four sources, in this order:
the machine's Obsidian vault, git, open PRs, and local plan/progress files. End with a
short "where we are / what's next" and a concrete proposal.

## 0. Preconditions (do these silently, report only problems)

1. **Vault config.** Read `~/.claude/vault.json`. It is per machine and never in git:

   ```json
   {
     "name": "brain-vault",
     "path": "/mnt/e/brain-vault",
     "mcp": "obsidian-vault",
     "projects_dir": "02-projects",
     "sessions_subdir": "sesiones",
     "index_note": "00-home.md",
     "rules_note": "00-meta/rules.md",
     "projects": { "claude-agents": "02-projects/claude-agents" }
   }
   ```

   - `mcp`: which MCP to use for reads/writes: `obsidian-vault` (file-based, works with
     Obsidian closed), `obsidian` (Local REST API, needs Obsidian open), or `filesystem`
     (plain Read/Write on `path`).
   - `projects`: optional explicit map repo-name → vault folder; otherwise look for
     `<projects_dir>/<repo-name>/`.
   - `rules_note` / `index_note` / `sessions_subdir` are optional.

   If the file does not exist: ask the user ONE question ("¿Qué vault uso en esta
   máquina? ruta y carpeta de proyectos") and create it. Do not guess from other
   machines' memories: the work PC has a different, local-only vault.

2. **Vault reachable.** `path` must exist (the SessionStart health hook already warns
   when a drive is not mounted). If the chosen MCP fails, fall back to `filesystem` on
   `path`; if that also fails, say so and continue with git-only context.

3. **Project identity.** Repo name = basename of `git rev-parse --show-toplevel` (or
   `CLAUDE_PROJECT_DIR`). Resolve the vault folder via `projects` map, else
   `<projects_dir>/<repo-name>`. If none exists, say it and continue git-only.

## 1. Gather (parallel where possible)

Run these together; none depends on another:

- **Vault**: read the project's main note (README.md / `<repo-name>.md` / index note of
  the folder), the **most recent** note under `sessions_subdir` (sort by filename date
  or mtime), and `rules_note` if set. Do not read the whole folder.
- **Git**: `git status --short`, `git log --oneline -15`, current branch, ahead/behind
  (`git fetch -q` first, then `git status -sb`), stashes.
- **PRs**: `gh pr list --state open --limit 10` and, if on a feature branch,
  `gh pr view --json state,title,reviewDecision` for it. Skip silently if `gh` fails.
- **Local plans/progress**: `claude-progress.txt`, `plans/`, `feature_list.json`,
  `TODO.md`, `docs/roadmap*` — whichever exist in the repo.

For a broad or unfamiliar repo, delegate the vault read and the git/PR read to two
`Agent` calls in the same message and keep only their conclusions.

## 2. Report

Keep it under ~15 lines. Structure:

```
## <project> — dónde quedamos
Última sesión: <fecha> — <1 línea de la nota de sesión>
Estado: <1-2 líneas desde la nota del proyecto / progress file>
Git: rama <x>, <n> commits sin push, <m> cambios sin commitear; PRs abiertos: #..
Pendiente / siguiente paso (según notas): ...
Desfases: <lo que git muestra y la nota no refleja, o viceversa>
```

Then **propose one next action** and ask whether to go with it. Do not start working
on it before the user answers.

## 3. Guardrails

- Never write to the vault here; that is `session-save`.
- Never mix vaults: all paths come from this machine's `vault.json`.
- If the vault is unreachable, say it once and move on with git context. Do not retry
  the MCP in a loop.
- Dates: compare the last session note date with today and flag staleness explicitly
  ("la última nota es del 10, hoy es 12").
