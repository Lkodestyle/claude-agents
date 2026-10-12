---
name: session-save
description: Use when the user wants to record what was done in this session ("anotá todo en el vault", "guardá lo que hicimos", "cerremos la sesión", "actualizá las notas", "documentá esto en obsidian"). Writes a dated session note and updates the project note and index in the Obsidian vault configured on THIS machine (~/.claude/vault.json), following that vault's rules, and saves durable facts to memory. Shows the draft before writing.
---

# Session Save

Persist the session's outcome where the user will look for it next time: a session note
in the vault, the project's status note, the vault index, and (only for durable facts)
memory. Always on the vault of **this machine**, never on another one's.

## 0. Preconditions

Same as `session-resume`: read `~/.claude/vault.json` (ask once and create it if
missing), check `path` exists, pick the MCP (`obsidian-vault` / `obsidian` /
`filesystem`), resolve the project folder. If the vault is unreachable, do NOT write
anywhere else as a substitute: tell the user and offer to leave the note text in the
reply so they can paste it later.

## 1. Learn the vault's conventions before writing

- Read `rules_note` if configured. Its rules win over this skill.
- Read the **latest existing session note** of this project and use it as the template
  (frontmatter keys, headings, tags, link style). If there is none, use the default
  template below.
- Read the project note to find its status/last-update section.

## 2. Collect what happened (from evidence, not memory)

- `git log --oneline <start>..HEAD` where `<start>` is the commit at session start (or
  since the last session note's date), plus `gh pr list --author @me --state all
  --limit 10` filtered to today.
- Decisions taken in the conversation, blockers found, and what was explicitly left for
  next time.
- Files/areas touched (`git diff --stat` of the session's commits).

## 3. Draft, show, then write

Draft the session note. Default template (used only if the vault has no example):

```markdown
---
date: YYYY-MM-DD
project: <repo-name>
tags: [sesion, <repo-name>]
---
# Sesión YYYY-MM-DD — <título corto>

## Qué se hizo
- ...

## Decisiones
- ...

## Pendiente / siguiente paso
- ...

## Referencias
- PRs: #.. ; commits: abc123..def456
```

File name: `YYYY-MM-DD-<slug>.md` under `<project folder>/<sessions_subdir>/` (create
the folder if needed). Then:

1. Show the user the draft and the list of files that will change (session note, project
   note status section, index entry). **Wait for confirmation.**
2. Write the session note.
3. Update the project note: its status/"última actualización" section with today's date
   and 1-3 lines. Edit the section in place; do not rewrite the note.
4. Update `index_note` if configured: add/refresh the project's line with today's date
   (keep the existing format of that note).
5. Memory: save ONLY durable facts (decisions, constraints, conventions) to Claude's
   auto-memory for this project, one fact per file, following the memory rules. Session
   chronology does not go to memory; it is in the vault.

## 4. Confirm

Reply with the paths written (vault-relative) and one line per update. If something
failed, say which step and leave the content in the reply.

## Guardrails

- Never write to a vault other than the one in this machine's `vault.json`; never copy
  work-vault content into memory shared across machines.
- Never delete or rewrite existing notes; append or edit sections in place.
- Dates come from the system clock, not from the last note.
- Respect the vault's language (Spanish by default for Luke's vaults).
