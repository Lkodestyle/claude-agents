---
name: session-save
description: Use when the user wants to record what was done in this session ("anotá todo en el vault", "guardá lo que hicimos", "cerremos la sesión", "actualizá las notas", "documentá esto en obsidian"). Writes a dated session note from the vault's own template, refreshes the project note, validates, and commits/pushes the vault, all on the Obsidian vault configured on THIS machine (~/.claude/vault.json) and under that vault's rules. Shows the draft before writing.
---

# Session Save

Persist the session's outcome where the user will look for it next time: a session note
and an updated project note in the vault, plus (only for durable facts) memory. Always
on the vault of **this machine**, never on another one's.

## 0. Preconditions

Same as `session-resume`: read `~/.claude/vault.json` (ask once and create it if
missing), check `path` exists, choose the MCP (`obsidian-vault` / `obsidian` /
`filesystem`), resolve the project folder. If `git` is configured: `git -C <path> pull
--ff-only -q` before touching anything, so you never write on top of a stale clone.
Run the pull as its OWN command and check its result: if it fails for credentials over
https, retry with `git -c credential.helper='!gh auth git-credential' pull --ff-only`.
Never chain the pull with the writes using `&&`: a failed pull would silently skip them.

If the vault is unreachable, do NOT write anywhere else as a substitute: say so and
leave the note text in the reply for the user to paste later.

## 1. Learn the vault's rules before writing

- Read `rules_note`. It defines the frontmatter contract (e.g. `category` from a closed
  vocabulary, `date`, `project`, `status`, `tags`), naming rules, what is append-only, and
  what needs confirmation. **Those rules win over anything below.**
- Read `session_template` (e.g. `_templates/template-sesion.md`). Every new note comes
  from the template; never write frontmatter from scratch.
- Read the **latest existing session note** of this project to match its style (title
  format, link style, how `Contexto` chains to the previous session). If there is none,
  the template alone is the reference.
- Read the project's root note (named like the project) to find its `updated:` field
  and status/“estado” section.

## 2. Collect what happened (from evidence, not memory)

- `git log --oneline <start>..HEAD` where `<start>` is the commit at session start (or
  since the last session note's date); PRs: `gh pr list --author @me --state all
  --limit 10` filtered to today, with merge state.
- Decisions taken in the conversation, blockers, what was explicitly left for next time.
- How things were validated (tests run, commands, real runs). Vault session notes
  usually require this section: without it the session is not closed.

## 3. Draft, show, then write

1. **Session note** from the template: file `<project folder>/<sessions_subdir>/
   YYYY-MM-DD-<slug>.md` (create the folder if missing). Fill every section the template
   has; frontmatter per the rules (category wikilink, `date` today, `project`
   wikilink, `status`, project tag + `sesion` tag, no per-feature tags, no new tags).
2. **Project root note**: refresh `updated:` and the status/"estado" lines (1-3 lines),
   editing in place. Many vaults forbid modifying existing notes without explicit
   confirmation, so this edit is part of what you show for approval.
3. **No hand-maintained index.** If the vault generates its tables from frontmatter
   (Bases/Dataview), there is nothing to update; a correct frontmatter is what makes the
   note appear. Only touch an index note if `vault.json` names one (`index_note`).
4. Show the user: the full session note draft and the exact diff for the project note.
   **Wait for confirmation.** Then write both.
5. **Validate**: run `validate` from `vault.json` (e.g. `python3 .claude/validate-vault.py`)
   from the vault root. Fix what it reports before continuing.
6. **Commit and push** when `git` is configured: from the vault root,
   `git add -A && git commit -m "docs(<project>): <qué cambió>" && git push origin
   <branch>`. Use the git identity the vault's rules specify, if any. This is what keeps
   the other machines' clones current; skipping it is how the vault "queda desactualizado".
7. **Memory**: save ONLY durable facts (decisions, constraints, conventions) to Claude's
   auto-memory for this project, one fact per file, following the memory rules. Session
   chronology stays in the vault.

## 4. Confirm

Reply with the vault-relative paths written, the validator result, and the commit hash
(or "sin push: <motivo>"). If a step failed, say which and leave the content in the reply.

## Guardrails

- Never write to a vault other than this machine's `vault.json`; never copy work-vault
  content into memory shared across machines.
- Session notes are append-only: never rewrite a previous one; a changed decision is a
  new note that supersedes it.
- Never create duplicate generic names (`README.md`) inside a project folder.
- Dates come from the system clock, not from the last note.
- Language of the vault (Spanish for Luke's vaults).
