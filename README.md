# Claude DevOps Agents Hub

A comprehensive collection of specialized AI agents for Claude Code, focused on DevOps, Cloud Infrastructure, IaC, and software development. Includes cognitive features for intelligent agent activation and multi-instance coordination.

## Features

- **15 Specialized Agents** - DevOps, Cloud, IaC, and development expertise
- **Cognitive Context Router** - Intelligent agent activation based on conversation keywords
- **Pool Coordinator** - Multi-instance coordination for teams
- **Memory Manager** - MCP memory management utilities
- **MCP Servers** - Pre-configured servers for memory, docs, and integrations

## Quick Start

```bash
# Clone the repository
git clone https://github.com/Lkodestyle/claude-agents.git ~/claude-agents
cd ~/claude-agents

# Full setup (agents + MCP + Obsidian + plugins — interactive)
./scripts/setup.sh

# Or install only agents/skills
./scripts/claude-agents-cli.sh install

# Verify installation
./scripts/claude-agents-cli.sh status
```

The `setup.sh` script runs a complete onboarding in 4 steps:
1. Install agents, skills, and scripts (symlink or copy)
2. Configure MCP servers (memory, context7, notion, supabase)
3. Setup Obsidian vault with full structure (folders, templates, guides)
4. Enable recommended plugins (superpowers, skill-creator)

## Agents

`architecture` and `security` run on `opus` (reasoning-heavy, low volume); the other 13 on `sonnet`.

| Agent | Specialty | Keywords |
|-------|-----------|----------|
| `architecture` | System design, C4 diagrams, ADRs, trade-offs | architecture, design, microservices, scalability |
| `terraform` | Terraform, Terraspace, Terragrunt, IaC | terraform, hcl, tfvars, module, provider |
| `aws` | Amazon Web Services (EC2, ECS, Lambda, RDS, etc.) | aws, ec2, lambda, s3, dynamodb, cloudwatch |
| `azure` | Microsoft Azure (App Services, AKS, Functions, etc.) | azure, app service, aks, cosmos db, vnet |
| `kubernetes` | K8s, Helm, Kustomize, operators, troubleshooting | kubernetes, k8s, helm, pod, deployment, ingress |
| `docker` | Dockerfiles, multi-stage builds, Compose, security | docker, dockerfile, container, compose, image |
| `cicd` | GitHub Actions, GitLab CI, Jenkins, pipelines | cicd, pipeline, github actions, deploy, workflow |
| `databases` | PostgreSQL, DynamoDB, Redis, MongoDB, migrations | database, postgresql, mysql, redis, sql, migration |
| `monitoring` | Prometheus, Grafana, CloudWatch, SLOs, alerting | monitoring, prometheus, grafana, alert, metric |
| `networking` | VPCs, subnets, CIDR, security groups, load balancers | network, vpc, subnet, security group, load balancer |
| `programming` | Clean code, design patterns, testing, code review | code review, refactor, solid, design pattern, test |
| `web` | React, Next.js, Node.js, FastAPI, Docker, Nginx | react, next.js, typescript, node.js, frontend |
| `mobile` | React Native, Expo, Flutter, iOS, Android, MVPs | mobile, react native, expo, flutter, ios, android |
| `finops` | Cost optimization, rightsizing, Savings Plans | cost, finops, billing, budget, optimization |
| `security` | CVE scanning, OWASP, secrets detection, IAM | security, vulnerability, cve, owasp, secret |

## Slash Commands

None. The repo used to ship 8 (`/commit`, `/pr`, `/review`, `/test`, `/explain`,
`/refactor`, `/debug`, `/doc`); ten months of prompt history showed zero uses, since
the same requests work in natural language and the agents/skills handle them. Repeated
workflows live as skills instead.

## Skills

Installed automatically with `claude-agents-cli.sh install` into `~/.claude/skills/`.

### Routing Guide

| Skill | Purpose |
|-------|---------|
| `agents-hub` | Canonical entry point: routes any task to the right agent(s), maps co-activation, and explains how to fan work out to agents with the Agent tool. Use it at the start of any technical task. |

### Day-to-day Skills

Transversal to any project. They read `~/.claude/vault.json` (per machine, never in
git) to know which Obsidian vault to use, so the same skill works on a personal PC and
on a work PC with a different, local-only vault.

| Skill | Triggers on | What it does |
|-------|-------------|--------------|
| `session-resume` | "where were we", "por dónde nos quedamos", "retomemos" | Rebuilds context from the vault's project + last session note, git, open PRs and plan files; proposes the next step |
| `session-save` | "write it to the vault", "anotá todo", "cerremos la sesión" | Dated session note + project status + index in the vault, following its rules; durable facts to memory; shows the draft first |
| `git-flow` | "commit it", "open the PR", "commitealo y abrí el PR" | Branch from the right base, conventional commits without Co-Authored-By, PR in the repo's language, confirmation before push/merge, post-merge check |

### Workflow Skills

Complete end-to-end workflows that orchestrate multiple agents together.

| Skill | Agents Used | Description |
|-------|-------------|-------------|
| `deploy-pipeline` | cicd, docker, aws/azure, networking, security | Full CI/CD pipeline: Dockerfile + workflow + cloud deploy + IaC |
| `observability-setup` | monitoring, aws/azure, kubernetes | Metrics + logging + alerting + dashboards + SLOs |
| `infrastructure-as-code` | terraform, networking, aws/azure, security, architecture | VPC design + Terraform modules + multi-environment IaC |
| `security-hardening` | security, docker, cicd, aws/azure, kubernetes | Dependency audit + secrets scan + container hardening + pipeline security |
| `full-stack-scaffold` | architecture, web/mobile, databases, programming, docker | Project scaffolding: API + DB + frontend + Docker Compose + tests |

## Recommended Plugins

### Essential

| Plugin | Description | Install |
|--------|-------------|---------|
| `superpowers` | 14 process skills: brainstorming, writing-plans, TDD, systematic-debugging, verification-before-completion, git worktrees, parallel agents | `claude plugin install superpowers@claude-plugins-official` then `claude plugin enable superpowers@claude-plugins-official` (check with `claude plugin list`) |
| `skill-creator` | Create and test custom skills with eval framework | Via `/plugins` in Claude Code |
| `context7` | Up-to-date library/framework documentation | Included in `.mcp.json` |

### Optional (by use case)

| Plugin | Use Case |
|--------|----------|
| `github` / `gitlab` | Repos, PRs, issues, code review automation |
| `playwright` | E2E browser testing |
| `supabase` | Backend-as-a-service |
| `firebase` | Firebase apps and services |
| `slack` | Messaging, bug reports, notifications |
| `linear` | Issue tracking |
| `stripe` | Payments integration |

### Superpowers + Workflow Skills Integration

The workflow skills complement superpowers for complete development flows:

```
superpowers:brainstorming → Refine requirements
    ↓
superpowers:writing-plans → Create implementation plan
    ↓
Workflow Skill (deploy-pipeline, etc.) → Execute the work
    ↓
superpowers:verification-before-completion → Validate results
    ↓
superpowers:requesting-code-review → Review before merge
```

List all recommended plugins: `./scripts/claude-agents-cli.sh plugins`

## Cognitive Features

Inspired by [claude-cognitive](https://github.com/GMaN1911/claude-cognitive), this repo includes intelligent context management.

### Context Router

The Context Router nudges Claude towards the right agents, in one line per prompt:

```
User: "armá un pipeline de github actions que buildee el dockerfile y deploye a ecs"
      ↓
[context-router] Suggested agents: cicd, docker, aws · related: networking, terraform · skill: deploy-pipeline
```

It does **not** inject the agent files themselves: each subagent already gets
its own prompt when invoked, so repeating it in the main context only costs
tokens (the v2 router injected up to 25k chars per prompt).

**How it works:**

1. **Keyword Matching** - whole-word, case- and accent-insensitive, English and Spanish
2. **Attention Scores** - each agent has a score (0.0 - 1.0); a match sets it to 1.0
3. **Decay** - scores fade per turn, so a topic from a few prompts ago is still "related"
4. **Co-activation** - related agents get a boost when a primary agent fires
5. **Skill hint** - when 2+ agents of a workflow skill are HOT, the skill is suggested

**Tiers:**

| Tier | Score | Output |
|------|-------|--------|
| HOT | >= 0.8 | `Suggested agents:` |
| WARM | 0.25 - 0.8 | `related:` |
| COLD | < 0.25 | nothing |

**Where things live:**

- Config: `~/.claude/keywords.json` (a project can override it with its own `.claude/keywords.json`)
- State: `~/.claude/router/<project>.json` — never inside the project tree
- History: off by default (`"history": true` writes `<project>.history.jsonl` next to the state)

**Configuration** (`keywords.json`):

```json
{
  "output": "compact",            // "full" restores the v2 content injection
  "history": false,
  "keywords": {
    "agents/terraform.md": ["terraform", "hcl", "tfvars", "terragrunt"],
    "agents/aws.md": ["aws", "ec2", "lambda", "s3"]
  },
  "co_activation": {
    "agents/terraform.md": ["agents/aws.md", "agents/networking.md"]
  },
  "skills": {
    "infrastructure-as-code": ["terraform", "networking", "aws", "azure"]
  },
  "pinned": [],
  "thresholds": { "hot": 0.8, "warm": 0.25, "max_hot": 4, "max_warm": 4 }
}
```

Keep keywords specific. Generic words (`plan`, `app`, `test`, `state`, `image`)
fire on unrelated prompts and were removed in v3.

### Pool Coordinator

Coordinate work across multiple Claude Code **terminals** open at the same time
(you with two sessions, or a team). It is **opt-in**: the hooks exit immediately
unless `CLAUDE_INSTANCE` is set. Note this is not what runs agents in parallel —
that is the native `Agent` tool (see the `agents-hub` skill).

```bash
# Terminal 1
export CLAUDE_INSTANCE=A
claude

# Terminal 2
export CLAUDE_INSTANCE=B
claude
```

**Features:**

- Detects completed tasks and blockers from other instances
- Shares state via `.claude/pool/instance_state.jsonl`
- Prevents duplicate work in team settings
- Loads recent activity at session start

**Pool CLI:**

```bash
# Query pool state
python3 .claude/scripts/pool-query.py --count
python3 .claude/scripts/pool-query.py --recent 5
```

## Memory Manager

Manage MCP memory to prevent token overflow:

```bash
# Show memory statistics
python3 .claude/scripts/memory-manager.py stats

# List all entities
python3 .claude/scripts/memory-manager.py list

# Search for entities
python3 .claude/scripts/memory-manager.py search "keyword"

# Export memory (backup)
python3 .claude/scripts/memory-manager.py export backup.json

# Clear memory (creates automatic backup)
python3 .claude/scripts/memory-manager.py clear

# Import memory
python3 .claude/scripts/memory-manager.py import backup.json
```

**When to use:**

- Getting token limit errors
- Memory stats show > 20,000 estimated tokens
- Starting a fresh project context

## Jarvis

Jarvis (voice assistant + semantic memory) vive ahora en su propio repo:
https://github.com/Lkodestyle/jarvis

## MCP Servers

Pre-configured MCP servers in `.mcp.json`:

| Server | Description | Setup Required |
|--------|-------------|----------------|
| `memory` | Persistent memory across sessions | None |
| `context7` | Up-to-date library documentation | Add "use context7" to prompts |
| `supabase` | Supabase project interaction | OAuth (automatic) |
| `notion` | Notion workspace access | `NOTION_TOKEN` env var |

### Obsidian (optional, per machine)

The session skills read the vault through whatever `~/.claude/vault.json` names as `mcp`:

- **`obsidian-vault`** (recommended): [`@bitbonsai/mcpvault`](https://www.npmjs.com/package/@bitbonsai/mcpvault),
  file-based, works with Obsidian closed. Register it once per machine with that machine's vault path:
  `claude mcp add obsidian-vault -s user -- npx @bitbonsai/mcpvault@latest /path/to/vault`
- **`obsidian`** (Local REST API plugin, live edits while Obsidian is open): needs Obsidian running, the
  plugin's HTTP server enabled (port 27123) and its API key. It is no longer in `.mcp.json`: it failed on
  every machine where Obsidian was closed or the key unset. To use it on a machine:
  `claude mcp add obsidian -s user -t http http://127.0.0.1:27123/mcp/ -H "Authorization: Bearer $OBSIDIAN_API_KEY"`
- **`filesystem`**: plain reads/writes on `path`, no MCP at all.

`scripts/setup-obsidian.sh` walks through the REST setup if you want it.

### Setup

**Notion:**
1. Create integration at https://www.notion.so/profile/integrations
2. Set environment variable:
   ```bash
   export NOTION_TOKEN="ntn_your_token_here"
   ```

**Context7:**
Just add "use context7" to your prompts to get up-to-date documentation.

### MCP Proxy (Optional)

Aggregate multiple MCP servers into a single endpoint:

```bash
# Install mcp-proxy (requires Go)
go install github.com/tbxark/mcp-proxy@latest

# Start proxy with included config
./scripts/claude-agents-cli.sh proxy

# Or run in background
./scripts/claude-agents-cli.sh proxy --background
```

## CLI Reference

```bash
./scripts/claude-agents-cli.sh <command> [options]

Commands:
  install     Install agents, skills, and scripts
              --symlink   Use symlinks (default, auto-updates)
              --copy      Copy files (independent)
              --global    Install to ~/.claude (default)
              --local     Install to current project

  sync        Sync config after git pull (for --copy installs)

  status      Show installation status (agents, skills, hooks)

  test        Test context router and pool coordinator

  proxy       Start mcp-proxy server
              --background  Run in background

  obsidian    Setup Obsidian Vault MCP server

  plugins     Show recommended plugins and installation guide

  uninstall   Remove installed agents, skills, and scripts

  help        Show help message
```

## Mods (interface plugins)

Two hooks-module plugins live in `mods/` and are installed by `install`/`sync` through the
repo's own marketplace (`.claude-plugin/marketplace.json`), at user scope. Claude Code reads
them **from the repo folder**, so after a `git pull` a `/reload-plugins` is all it takes.

| Mod | What you see |
|-----|--------------|
| `session-band` | At session start, a band above the prompt: last session note of this project in the vault (date, title, next steps), git branch/dirty/unpushed, open PRs. Hides on the first prompt. Adds `/vault` (pane with the last note's closing state and next steps) and `/health` (re-run the health check). After 3h of session with nothing written to the vault's sessions folder, a toast + native notification reminds you to save. |
| `guardrails` | Invisible until needed. Denies: `git push --force` (unless `--force-with-lease`), deleting `main`/`master`, `git reset --hard` on main, `rm -rf` of `~` or `/`, and edits to `.env*` (not `.env.example`), `secrets/`, `credentials.json`, `*token*.json`, SSH keys. |

Manual install on a machine: `claude plugin marketplace add <repo path>` then
`claude plugin install session-band@claude-agents` (same for `guardrails`). Remove with
`claude plugin uninstall <mod>@claude-agents`. Both need Claude Code with mods support (v2.1.29x+).

## Status Line

`install`/`sync` also configure a status line (only if you have none; a custom one is left
alone) from `.claude/scripts/statusline.py`:

```
Fable · claude-agents · main* ↑2 · PR #11
ctx ████░░░░░░ 38% (76k/200k) · cache 91% warm · $0.42 · 12m · 5h 24% · 7d 41%
```

Line 1: model, project, git branch (`*` dirty, `↑n` unpushed), `vault Nd` (days since the last
session note of this project in the machine's vault; yellow after a day, red after a week), open PR,
`fast`/effort when non-default. Line 2: context window usage (green/yellow/red), prompt-cache hit ratio and
warm/cold, session cost and duration, 5-hour and 7-day rate-limit usage. Fields that Claude
Code does not provide (older versions, no git) are skipped. To use your own, set
`statusLine` in `~/.claude/settings.json`; the installer will not touch it.

## Hooks

The hooks are defined once in `scripts/hooks.json` and installed by
`claude-agents-cli.sh install` into your **global** `~/.claude/settings.json`:

| Hook | Script | Function |
|------|--------|----------|
| `UserPromptSubmit` | `context-router.py` | Activate agents based on keywords |
| `SessionStart` | `session-health.py` | Warn about Obsidian down, vault not mounted, duplicated hooks, broken agent links |
| `SessionStart` | `pool-loader.py` | Load state from other instances (only with `CLAUDE_INSTANCE`) |
| `Stop` | `pool-extractor.py` | Save completion signals |

The installer **merges** into your existing `settings.json`: it only replaces
hooks that reference these three scripts and leaves everything else (model,
theme, other tools' hooks, your own permissions) untouched. A copy of the
previous file is kept as `settings.json.backup.<timestamp>`.

The repo's own `.claude/settings.json` deliberately contains **no hooks**: if a
project declares the same hooks as the global settings, Claude Code runs them
twice per event. `claude-agents-cli.sh status` reports how many times each hook
would run.

**Protect sensitive files:**

Add to `.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 -c \"import json,sys; d=json.load(sys.stdin); p=d.get('tool_input',{}).get('file_path',''); sys.exit(2 if '.env' in p else 0)\""
          }
        ]
      }
    ]
  }
}
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `CLAUDE_INSTANCE` | Instance ID for pool coordination (A, B, C...) | No |
| `NOTION_TOKEN` | Notion integration token | For Notion MCP |
| `GITHUB_TOKEN` | GitHub personal access token | For GitHub MCP |
| `OBSIDIAN_API_KEY` | API key from Local REST API plugin | For Obsidian MCP |
| `MCP_PROXY_TOKEN` | Token for mcp-proxy auth | For mcp-proxy |
| `MAX_MCP_OUTPUT_TOKENS` | Token limit per MCP tool result (default 25000). Raising it globally raises the cost of every MCP call; prefer per-tool `_meta["anthropic/maxResultSizeChars"]` | No |

## Project Structure

```
claude-agents/
├── CLAUDE.md                 # Project instructions for Claude
├── README.md                 # This file
├── .mcp.json                 # MCP servers configuration
├── mcp-proxy-config.json     # MCP proxy configuration
├── .gitignore
├── .claude/
│   ├── agents/               # 15 specialized agents
│   │   ├── architecture.md
│   │   ├── aws.md
│   │   ├── azure.md
│   │   ├── cicd.md
│   │   ├── databases.md
│   │   ├── docker.md
│   │   ├── finops.md
│   │   ├── kubernetes.md
│   │   ├── mobile.md
│   │   ├── monitoring.md
│   │   ├── networking.md
│   │   ├── programming.md
│   │   ├── security.md
│   │   ├── terraform.md
│   │   └── web.md
│   │   ├── commit.md
│   │   ├── diagnose.md
│   │   ├── doc.md
│   │   ├── explain.md
│   │   ├── pr.md
│   │   ├── refactor.md
│   │   ├── review-changes.md
│   │   └── test.md
│   ├── skills/               # Skills (routing guide + workflows)
│   │   ├── agents-hub/           # Agent routing guide (entry point)
│   │   ├── deploy-pipeline/
│   │   ├── observability-setup/
│   │   ├── infrastructure-as-code/
│   │   ├── security-hardening/
│   │   └── full-stack-scaffold/
│   ├── scripts/              # Cognitive scripts
│   │   ├── context-router.py
│   │   ├── memory-manager.py
│   │   ├── pool-extractor.py
│   │   ├── pool-loader.py
│   │   ├── session-health.py # SessionStart health check
│   │   ├── statusline.py     # Status line (context, cache, cost, git)
│   │   └── pool-query.py
│   ├── keywords.json         # Keyword activation config
│   └── settings.json         # Hooks configuration
├── scripts/
│   ├── setup.sh              # Full onboarding (runs all setup steps)
│   └── claude-agents-cli.sh  # CLI installer
└── templates/                # Reference guides (one README per type: ecs, gh-workflow, gitlab-ci, tf-module)
    ├── ecs-service/
    ├── github-workflow/
    ├── gitlab-ci/
    └── terraform-module/
```

## Sync Across Machines

```bash
# On any machine: clone and install
git clone https://github.com/Lkodestyle/claude-agents.git ~/claude-agents
cd ~/claude-agents
./scripts/claude-agents-cli.sh install

# To update (if using symlinks)
cd ~/claude-agents
git pull
# Done! Symlinks auto-update

# To update (if using --copy)
cd ~/claude-agents
git pull
./scripts/claude-agents-cli.sh sync
```

## Customization

Create `CLAUDE.local.md` in any project for local context (not committed):

```markdown
# Local Context

- Client: MyCompany
- Environment: Production US-East-1
- Special requirements: PCI compliance
```

## Best Practices

Tips from the Claude Code team for getting the most out of this setup:

**1. Verification is the #1 tip** — Give Claude a way to verify its work (tests, linters, browser). This 2-3x the quality of the output. Use `superpowers:verification-before-completion`.

**2. Git worktrees for parallel work** — Run 3-5 Claude instances, each in its own worktree:
```bash
git worktree add .claude/worktrees/feature-a origin/main
cd .claude/worktrees/feature-a && claude
```

**3. Start in Plan Mode** — Invest in the plan first (`shift+tab` twice), then switch to auto-accept. Claude can often 1-shot the implementation from a good plan.

**4. Keep CLAUDE.md alive** — After every correction, tell Claude: *"Update CLAUDE.md so you don't make that mistake again."* The error rate drops measurably over time.

**5. Skills for repetitive work** — If you do something more than once a day, make it a skill in `.claude/skills/<name>/SKILL.md` so it triggers from natural language.

**6. PostToolUse hook for formatting** — Auto-format after every Write/Edit to prevent CI failures:
```json
"PostToolUse": [{
  "matcher": "Write|Edit",
  "hooks": [{ "type": "command", "command": "npx prettier --write \"$(jq -r '.tool_input.file_path')\" || true" }]
}]
```

**7. Pre-approve safe permissions** — Use `/permissions` to allow common safe commands instead of `--dangerously-skip-permissions`. Commit them in `settings.json`.

**8. Challenge Claude** — "Implement the elegant solution" beats accepting the first fix. "Prove to me this works" forces verification.

**9. Use subagents for more compute** — Add "use subagents" to requests when you want Claude to throw more compute at the problem.

**10. PostCompact hook** — Reinject critical instructions when context gets compacted:
```json
"PostCompact": [{
  "matcher": "",
  "hooks": [{ "type": "command", "command": "cat CLAUDE.md | head -50" }]
}]
```

> Source: [How Boris Uses Claude Code](https://howborisusesclaudecode.com) — Boris Cherny, creator of Claude Code at Anthropic.

## References

This project incorporates patterns from:

- [claude-cognitive](https://github.com/GMaN1911/claude-cognitive) - Context Router and Pool Coordinator patterns
- [mcp-proxy](https://github.com/tbxark/mcp-proxy) - MCP server aggregation
- [gen-ai-cve-patching](https://github.com/aws-samples/gen-ai-cve-patching) - Security agent patterns

## Contributing

1. Fork the repo
2. Create your branch: `git checkout -b feature/new-agent`
3. Commit: `git commit -m 'Add Kubernetes agent'`
4. Push: `git push origin feature/new-agent`
5. Open a PR

## License

MIT
