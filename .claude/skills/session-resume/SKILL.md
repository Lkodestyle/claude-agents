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
     "path": "/home/kryze/brain-vault",
     "mcp": "obsidian-vault",
     "projects_dir": "02-projects",
     "sessions_subdir": "sesiones",
     "rules_note": "CLAUDE.md",
     "session_template": "_templates/template-sesion.md",
     "validate": "python3 .claude/validate-vault.py",
     "git": { "branch": "master", "push": true },
     "projects": { "claude-agents": "02-projects/claude-agents" }
   }
   ```

   - `mcp`: `obsidian-vault` (file-based, works with Obsidian closed), `obsidian` (Local
     REST API, needs Obsidian open) or `filesystem` (plain Read/Write on `path`).
   - `rules_note`: the vault's own contract (frontmatter vocabulary, naming, what an
     agent may touch). **Its rules win over this skill.**
   - `projects`: optional map repo-name → vault folder; default `<projects_dir>/<repo-name>`.
   - `git`: present when the vault is a git repo that must be pulled/pushed.

   If the file does not exist: ask ONE question ("¿Qué vault uso en esta máquina? ruta y
   carpeta de proyectos") and create it. Do not guess from other machines' memories:
   the work PC has a different, local-only vault.

2. **Vault reachable and fresh.** `path` must exist (the SessionStart health hook already
   warns when it is not). If `git` is configured, `git -C <path> pull --ff-only -q` first
   so you read the latest notes from the other machines (on a credentials error over
   https, retry with `git -c credential.helper='!gh auth git-credential' pull`; if it
   still fails, say so and read the local copy). If the MCP fails, fall back to
   `filesystem` on `path`; if that fails too, say so once and continue git-only.

3. **Project identity.** Repo name = basename of `git rev-parse --show-toplevel` (or
   `CLAUDE_PROJECT_DIR`). Resolve the vault folder via `projects`, else
   `<projects_dir>/<repo-name>`. If none exists, say it and continue git-only.

## 1. Gather (parallel where possible)

Run these together; none depends on another:

- **Vault**: read `rules_note` (skim: you need the section on how an agent behaves), the
  project's **root note** (named like the project, e.g. `claude-agents.md`; never a
  `README.md`), its status note if one exists (`estado*.md`, `estado-y-deuda.md`,
  `estado-features.md`), and the **most recent** note under `sessions_subdir` (sort by
  the `YYYY-MM-DD-` prefix). Do not read the whole folder.
- **Git**: `git fetch -q`, then `git status -sb`, `git log --oneline -15`, stashes.
- **PRs**: `gh pr list --state open --limit 10`; if on a feature branch,
  `gh pr view --json state,title,reviewDecision`. Skip silently if `gh` fails.
- **Local plans/progress**: `claude-progress.txt`, `plans/`, `feature_list.json`,
  `TODO.md`, `docs/roadmap*` — whichever exist in the repo.

For a broad or unfamiliar repo, delegate the vault read and the git/PR read to two
`Agent` calls in the same message and keep only their conclusions.

## 2. Report

Keep it under ~15 lines:

```
## <project> — dónde quedamos
Última sesión en el vault: <fecha> — <1 línea>   (o "sin notas de sesión")
Estado según la nota del proyecto (updated <fecha>): <1-2 líneas>
Git: rama <x>, <n> commits sin push, <m> cambios sin commitear; PRs abiertos: #..
Pendiente / siguiente paso según las notas: ...
Desfases: lo que git muestra y el vault no refleja (p. ej. "la nota dice 9 comandos, el repo tiene 0")
```

Then **propose one next action** and ask whether to go with it. Do not start working on
it before the user answers.

## 3. Guardrails

- Never write to the vault here; that is `session-save`.
- Never mix vaults: every path comes from this machine's `vault.json`.
- Flag staleness explicitly with real dates ("la última nota es del 10, hoy es 12").
- If the vault is unreachable, say it once and move on. No retry loops.
