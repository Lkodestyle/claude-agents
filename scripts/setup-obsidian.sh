#!/bin/bash
# scripts/setup-obsidian.sh
# Script para configurar el MCP server de Obsidian Vault en Claude Code
#
# Uso:
#   ./scripts/setup-obsidian.sh                    # Setup interactivo
#   ./scripts/setup-obsidian.sh /path/to/vault     # Setup con path directo

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

log_info() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "${BLUE}[→]${NC} $1"; }

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║   Obsidian Vault MCP Setup               ║${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""

# Check prerequisites
if ! command -v npx &> /dev/null; then
    log_error "npx not found. Install Node.js first: https://nodejs.org"
    exit 1
fi

# Get vault path
VAULT_PATH="${1:-}"

if [[ -z "$VAULT_PATH" ]]; then
    echo -e "${BOLD}Enter the path to your Obsidian vault:${NC}"
    echo -e "  Example: ~/brain-vault"
    echo -e "  Example: /mnt/c/Users/you/Documents/my-vault"
    echo ""
    read -r -p "> " VAULT_PATH
fi

# Expand tilde
VAULT_PATH="${VAULT_PATH/#\~/$HOME}"

# Validate vault path
if [[ ! -d "$VAULT_PATH" ]]; then
    log_error "Directory not found: $VAULT_PATH"
    exit 1
fi

if [[ ! -d "$VAULT_PATH/.obsidian" ]]; then
    log_warn "No .obsidian directory found. This may not be an Obsidian vault."
    read -r -p "Continue anyway? (y/N) " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        log_info "Aborted."
        exit 0
    fi
fi

log_info "Vault found: $VAULT_PATH"

# Pre-install the MCP package
log_step "Installing @bitbonsai/mcpvault..."
npx -y @bitbonsai/mcpvault@latest --help > /dev/null 2>&1 || {
    log_warn "Could not verify mcpvault package. It will be downloaded on first use."
}
log_info "mcpvault package ready"

# Determine MCP config location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
MCP_CONFIG="$REPO_DIR/.mcp.json"

if [[ ! -f "$MCP_CONFIG" ]]; then
    log_error "MCP config not found: $MCP_CONFIG"
    exit 1
fi

# Update .mcp.json
log_step "Updating MCP configuration..."

python3 -c "
import json
import sys

config_path = '$MCP_CONFIG'
vault_path = '$VAULT_PATH'

with open(config_path, 'r') as f:
    config = json.load(f)

servers = config.get('mcpServers', {})

servers['obsidian-vault'] = {
    'type': 'stdio',
    'command': 'npx',
    'args': ['-y', '@bitbonsai/mcpvault@latest', vault_path]
}

config['mcpServers'] = servers

with open(config_path, 'w') as f:
    json.dump(config, f, indent=2)
    f.write('\n')

print('OK')
" || {
    log_error "Failed to update MCP config"
    exit 1
}

log_info "MCP config updated: $MCP_CONFIG"

# Also update global config if requested
echo ""
echo -e "${BOLD}Also install to global config (~/.mcp.json)?${NC}"
read -r -p "(y/N) " install_global

if [[ "$install_global" == "y" || "$install_global" == "Y" ]]; then
    GLOBAL_MCP="$HOME/.mcp.json"

    if [[ -f "$GLOBAL_MCP" ]]; then
        python3 -c "
import json

config_path = '$GLOBAL_MCP'
vault_path = '$VAULT_PATH'

with open(config_path, 'r') as f:
    config = json.load(f)

servers = config.get('mcpServers', {})

servers['obsidian-vault'] = {
    'type': 'stdio',
    'command': 'npx',
    'args': ['-y', '@bitbonsai/mcpvault@latest', vault_path]
}

config['mcpServers'] = servers

with open(config_path, 'w') as f:
    json.dump(config, f, indent=2)
    f.write('\n')

print('OK')
"
        log_info "Global MCP config updated: $GLOBAL_MCP"
    else
        python3 -c "
import json

vault_path = '$VAULT_PATH'
config = {
    'mcpServers': {
        'obsidian-vault': {
            'type': 'stdio',
            'command': 'npx',
            'args': ['-y', '@bitbonsai/mcpvault@latest', vault_path]
        }
    }
}

with open('$GLOBAL_MCP', 'w') as f:
    json.dump(config, f, indent=2)
    f.write('\n')

print('OK')
"
        log_info "Global MCP config created: $GLOBAL_MCP"
    fi
fi

# Optional: create base vault structure
echo ""
echo -e "${BOLD}Create complete vault structure (folders, templates, guides)?${NC}"
echo -e "  Creates a full brain-vault ready for use with Claude Code."
echo -e "  Only creates files that don't already exist (safe to re-run)."
read -r -p "(y/N) " create_structure

if [[ "$create_structure" == "y" || "$create_structure" == "Y" ]]; then
    log_step "Creating vault structure..."

    # -- Directories --
    mkdir -p "$VAULT_PATH"/{00-inbox,01-daily,02-projects,03-areas,04-resources,05-archive,06-meetings,07-prompts,08-snippets,_templates,_assets}
    log_info "Directories created"

    # -- CLAUDE.md --
    if [[ ! -f "$VAULT_PATH/CLAUDE.md" ]]; then
        cat > "$VAULT_PATH/CLAUDE.md" << 'VAULTEOF'
# CLAUDE.md — Brain Vault

> This file is read automatically by Claude Code at session start.
> It tells Claude who you are, how the vault is organized, and how to behave.

## About Me

- **Name:** _your name_
- **Role:** _your role_
- **Goal:** Maintain a local second brain with actionable knowledge, project context, and technical resources.

## Vault Structure

```
brain-vault/
├── 00-inbox/        → Quick capture. Everything enters here first.
├── 01-daily/        → Daily notes, journaling, standup notes.
├── 02-projects/     → One subfolder per active project, each with its own README.md.
├── 03-areas/        → Ongoing responsibilities (career, health, finances, etc).
├── 04-resources/    → Processed knowledge: tutorials, references, how-tos.
├── 05-archive/      → Completed projects and inactive notes.
├── 06-meetings/     → Meeting notes and action items.
├── 07-prompts/      → Personal prompt library.
├── 08-snippets/     → Reusable code fragments.
├── _templates/      → Obsidian templates for creating notes fast.
├── _assets/         → Images, PDFs, attachments.
└── CLAUDE.md        → This file (context for Claude Code).
```

## Note Conventions

- All markdown files follow Obsidian conventions.
- Use `[[wikilinks]]` for internal links between notes.
- Use `#tags` for quick categorization.
- YAML frontmatter required on every note: at least `date`, `tags`.
- File names in kebab-case: `my-important-note.md`.

## How Claude Should Behave in This Vault

1. **Always read this file first** when starting a session.
2. **Respect the folder structure.** Don't create files outside the defined structure.
3. **Use wikilinks** when referencing other notes in the vault.
4. **Include frontmatter** in every new note it creates.
5. **Capture in 00-inbox/** if it's unclear where something goes.
6. **Don't modify existing notes** without confirmation, unless explicitly asked.
7. **Check 02-projects/** for context before answering questions about a project.
8. **Suggest connections** between notes when it detects thematic relationships.

## Active Projects

<!-- Update this table when you start or finish a project -->

| Project | Folder | Stack | Status |
|---------|--------|-------|--------|
| _example_ | `02-projects/example/` | _React, Node_ | `active` |

## Main Tags

- `#inbox` — Unprocessed
- `#project` — Related to a project
- `#resource` — Reference material
- `#snippet` — Reusable code
- `#meeting` — Meeting note
- `#prompt` — Saved prompt
- `#idea` — Undeveloped idea
- `#til` — Today I Learned
- `#bug` — Documented bug and its solution
- `#decision` — Technical decision taken (informal ADR)
VAULTEOF
        log_info "Created CLAUDE.md"
    else
        log_warn "CLAUDE.md already exists, skipping"
    fi

    # -- START-HERE.md --
    if [[ ! -f "$VAULT_PATH/START-HERE.md" ]]; then
        cat > "$VAULT_PATH/START-HERE.md" << 'VAULTEOF'
---
date: {{date}}
tags:
  - meta
---

# Welcome to your Brain Vault

This is your personal knowledge base. Everything you learn, think, and work on goes here.

## First Steps

### 1. Open this vault in Obsidian

Download [Obsidian](https://obsidian.md) if you don't have it, and open this folder as a vault.

### 2. Install recommended plugins

Go to **Settings > Community Plugins > Browse** and install:

- **Templater** — Templates with dynamic variables (set template folder to `_templates/`)
- **Dataview** — SQL-like queries on your notes (for dashboards)
- **Calendar** — Calendar view for daily notes

### 3. Start capturing

Don't overthink it. Drop everything in `00-inbox/` and organize later. The system grows with you.

## Using with Claude Code

Open a terminal, navigate to the vault, and run:

```bash
cd ~/path-to/brain-vault
claude
```

Claude reads `CLAUDE.md` automatically and understands your structure. You can ask things like:

- "Create a project note for my new app"
- "Search my resources for everything about Docker"
- "Summarize my daily notes from this week"

## Vault Structure

Read `CLAUDE.md` for the complete structure, or check the `README.md` inside each folder.
VAULTEOF
        log_info "Created START-HERE.md"
    else
        log_warn "START-HERE.md already exists, skipping"
    fi

    # -- Folder READMEs --
    create_readme() {
        local folder="$1"
        local content="$2"
        local filepath="$VAULT_PATH/$folder/README.md"
        if [[ ! -f "$filepath" ]]; then
            echo "$content" > "$filepath"
        fi
    }

    create_readme "00-inbox" "---
date: {{date}}
tags:
  - meta
---

# Inbox

Quick capture. Everything enters here first.
Process regularly: move notes to the right folder, add tags, link to projects.

> Tip: Don't let this pile up. Review weekly."

    create_readme "01-daily" "---
date: {{date}}
tags:
  - meta
---

# Daily Notes

One note per day. Use the \`template-daily\` template.
Name format: \`YYYY-MM-DD.md\`"

    create_readme "02-projects" "---
date: {{date}}
tags:
  - meta
---

# Projects

One subfolder per active project. When a project finishes, move its folder to \`05-archive/\`.

## Structure per project

\`\`\`
02-projects/
└── my-project/
    ├── README.md          → Description, stack, objectives (use template-proyecto)
    ├── decisions/         → ADRs and technical decisions
    ├── notes/             → Loose project notes
    └── tasks.md           → Pending tasks (optional)
\`\`\`

## Create a new project

1. Create a subfolder with the project name in kebab-case.
2. Use the \`template-proyecto\` template for README.md.
3. Update the Active Projects table below.

## Active Projects

<!-- Keep this list updated -->

| Project | Stack | Status | Started |
|---------|-------|--------|---------|"

    create_readme "03-areas" "---
date: {{date}}
tags:
  - meta
---

# Areas

Ongoing responsibilities — things without an end date that you want to maintain and improve.

## Examples

- **Career** — Professional goals, skills to develop, CV.
- **Finances** — Budget, investments, expenses.
- **Health** — Exercise routines, habits, medical notes.
- **Learning** — Courses, books, certifications in progress.

## Difference with Projects

A **project** has a goal and an end date. An **area** is something you maintain indefinitely."

    create_readme "04-resources" "---
date: {{date}}
tags:
  - meta
---

# Resources

Processed knowledge: tutorials, references, how-tos, and anything you want to remember.
Use the \`template-recurso\` template for new resources."

    create_readme "05-archive" "---
date: {{date}}
tags:
  - meta
---

# Archive

Completed projects and inactive notes. Move folders here when done — don't delete."

    create_readme "06-meetings" "---
date: {{date}}
tags:
  - meta
---

# Meetings

Meeting notes and action items. Use the \`template-meeting\` template.
Name format: \`YYYY-MM-DD-meeting-topic.md\`"

    create_readme "07-prompts" "---
date: {{date}}
tags:
  - meta
---

# Prompts

Personal library of prompts that work well. Use the \`template-prompt\` template.
Rate effectiveness and note which model works best."

    create_readme "08-snippets" "---
date: {{date}}
tags:
  - meta
---

# Snippets

Reusable code fragments. Use the \`template-snippet\` template.
Organize by language or category."

    log_info "Folder READMEs created"

    # -- Templates --
    create_template() {
        local name="$1"
        local content="$2"
        local filepath="$VAULT_PATH/_templates/$name"
        if [[ ! -f "$filepath" ]]; then
            echo "$content" > "$filepath"
        fi
    }

    create_template "template-daily.md" '---
date: {{date}}
tags:
  - daily
---

# Daily — {{date}}

## Yesterday I did

-

## Today I plan

- [ ]

## Blockers

-

## Notes / Ideas

_Whatever crosses your mind today._

## Links

- [[]]'

    create_template "template-proyecto.md" '---
date: {{date}}
tags:
  - project
status: active
stack: []
repo: ""
---

# {{title}}

## Description

_What is this project and what problem does it solve?_

## Objectives

- [ ] Objective 1
- [ ] Objective 2

## Tech Stack

| Technology | Use |
|-----------|-----|
| _e.g. React_ | _Frontend_ |

## Key Decisions

### {{date}} — _Decision title_

**Context:** ...
**Decision:** ...
**Reason:** ...

## Related Links

- Repo: [link]()
- Docs: [[]]

## Notes

_Free space for project notes._

## Progress Log

### {{date}}
- Project started.'

    create_template "template-meeting.md" '---
date: {{date}}
tags:
  - meeting
participants: []
project: ""
---

# Meeting — {{title}}

**Date:** {{date}}
**Participants:**
**Project:** [[]]

## Agenda

1.

## Notes

_Summary of what was discussed._

## Decisions Made

-

## Action Items

- [ ] @_who_ — _what_ — _by when_

## Follow-up

_Next steps and date of next meeting._'

    create_template "template-recurso.md" '---
date: {{date}}
tags:
  - resource
category: ""
source: ""
---

# {{title}}

## Summary

_What is this resource about and why did you save it?_

## Key Points

1.
2.
3.

## Personal Notes

_Your interpretation, how it applies to your work._

## Connections

- Related to: [[]]
- Useful for project: [[]]

## Source

- URL:
- Author:
- Date accessed: {{date}}'

    create_template "template-snippet.md" '---
date: {{date}}
tags:
  - snippet
language: ""
category: ""
---

# {{title}}

## Description

_What does this snippet do and when to use it?_

## Code

```
// Your code here
```

## Usage

_Example of how to use it in context._

## Notes

- Works with:
- Dependencies:
- Related: [[]]'

    create_template "template-prompt.md" '---
date: {{date}}
tags:
  - prompt
model: ""
use_case: ""
---

# {{title}}

## Prompt

```
Your prompt here
```

## Use Case

_When and why do you use this prompt?_

## Example Output

_Paste an example response you liked._

## Variants

- **Short version:** ...
- **Detailed version:** ...'

    create_template "template-til.md" '---
date: {{date}}
tags:
  - til
topic: ""
---

# TIL — {{title}}

## What I learned

_Clear and concise explanation._

## Example

```
// Code or practical example
```

## Why it matters

_How does it impact your work or understanding?_

## Source

- [[]]'

    log_info "Templates created (7 templates)"
    log_info "Vault structure complete!"
fi

# Summary
echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║   Setup Complete!                        ║${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""
log_info "Vault path: $VAULT_PATH"
log_info "MCP server: @bitbonsai/mcpvault@latest"
echo ""
echo -e "${BOLD}Available tools in Claude Code:${NC}"
echo "  - read_note / write_note — Read and write notes"
echo "  - search_notes — Search vault by content"
echo "  - list_directory — Browse vault structure"
echo "  - get_vault_stats — Vault statistics"
echo "  - manage_tags — Tag management"
echo ""
log_info "Restart Claude Code to activate the MCP server"
echo ""
