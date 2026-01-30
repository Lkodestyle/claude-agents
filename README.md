# Claude DevOps Agents Hub

A comprehensive collection of specialized AI agents for Claude Code, focused on DevOps, Cloud Infrastructure, IaC, and software development. Includes cognitive features for intelligent agent activation and multi-instance coordination.

## Features

- **15 Specialized Agents** - DevOps, Cloud, IaC, and development expertise
- **8 Slash Commands** - `/commit`, `/pr`, `/review`, `/test`, `/explain`, `/refactor`, `/debug`, `/doc`
- **Cognitive Context Router** - Intelligent agent activation based on conversation keywords
- **Pool Coordinator** - Multi-instance coordination for teams
- **Memory Manager** - MCP memory management utilities
- **MCP Servers** - Pre-configured servers for memory, docs, and integrations

## Quick Start

```bash
# Clone the repository
git clone https://github.com/Lkodestyle/claude-agents.git ~/claude-agents
cd ~/claude-agents

# Install with symlinks (recommended - auto-updates with git pull)
./scripts/claude-agents-cli.sh install

# Or install with copy (independent of repo)
./scripts/claude-agents-cli.sh install --copy

# Verify installation
./scripts/claude-agents-cli.sh status
```

## Agents

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

| Command | Description | Example |
|---------|-------------|---------|
| `/commit` | Generate conventional commit message | `git add . && /commit` |
| `/pr` | Create PR with auto-generated description | `/pr` |
| `/review` | Code review staged changes or files | `/review` or `/review src/file.ts` |
| `/test` | Generate unit tests for code | `/test src/utils.ts` |
| `/explain` | Explain how code works | `/explain src/auth/` |
| `/refactor` | Suggest and apply refactoring | `/refactor src/legacy.ts` |
| `/debug` | Help debug errors | `/debug "TypeError: cannot read..."` |
| `/doc` | Generate documentation | `/doc src/api/` |

## Cognitive Features

Inspired by [claude-cognitive](https://github.com/GMaN1911/claude-cognitive), this repo includes intelligent context management.

### Context Router

The Context Router automatically activates relevant agents based on conversation keywords:

```
User: "Help me deploy a Kubernetes pod with Terraform"
      ↓ Keywords detected: kubernetes, terraform, deploy
      ↓ Primary agents: kubernetes, terraform
      ↓ Co-activated: docker, cicd, networking
```

**How it works:**

1. **Keyword Matching** - Scans prompts for keywords defined in `keywords.json`
2. **Attention Scores** - Each agent has a score (0.0 - 1.0) based on relevance
3. **Decay Mechanism** - Unused agents gradually lose attention over time
4. **Co-activation** - Related agents get a boost when primary agents activate

**Activation Tiers:**

| Tier | Score | Injection |
|------|-------|-----------|
| HOT | > 0.8 | Full agent content |
| WARM | 0.25 - 0.8 | Headers only (first 25 lines) |
| COLD | < 0.25 | Not injected |

**Configuration:**

Edit `.claude/keywords.json` to customize:

```json
{
  "keywords": {
    "agents/terraform.md": ["terraform", "hcl", "tfvars", "module"],
    "agents/aws.md": ["aws", "ec2", "lambda", "s3"]
  },
  "co_activation": {
    "agents/terraform.md": ["agents/aws.md", "agents/networking.md"]
  },
  "pinned": ["agents/architecture.md"],
  "thresholds": {
    "hot": 0.8,
    "warm": 0.25,
    "max_hot_files": 4,
    "max_chars": 25000
  }
}
```

### Pool Coordinator

Coordinate work across multiple Claude Code instances:

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

## MCP Servers

Pre-configured MCP servers in `.mcp.json`:

| Server | Description | Setup Required |
|--------|-------------|----------------|
| `memory` | Persistent memory across sessions | None |
| `context7` | Up-to-date library documentation | Add "use context7" to prompts |
| `supabase` | Supabase project interaction | OAuth (automatic) |
| `notion` | Notion workspace access | `NOTION_TOKEN` env var |

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
  install     Install agents and scripts
              --symlink   Use symlinks (default, auto-updates)
              --copy      Copy files (independent)
              --global    Install to ~/.claude (default)
              --local     Install to current project

  sync        Sync config after git pull (for --copy installs)

  status      Show installation status and agent list

  test        Test context router and pool coordinator

  proxy       Start mcp-proxy server
              --background  Run in background

  uninstall   Remove installed agents and scripts

  help        Show help message
```

## Hooks

Claude Code hooks are configured in `.claude/settings.json`:

| Hook | Script | Function |
|------|--------|----------|
| `UserPromptSubmit` | `context-router.py` | Activate agents based on keywords |
| `SessionStart` | `pool-loader.py` | Load state from other instances |
| `Stop` | `pool-extractor.py` | Save completion signals |

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
| `MCP_PROXY_TOKEN` | Token for mcp-proxy auth | For mcp-proxy |
| `MAX_MCP_OUTPUT_TOKENS` | Token limit for MCP (default: 25000) | No |

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
│   ├── commands/             # Slash commands
│   │   ├── commit.md
│   │   ├── debug.md
│   │   ├── doc.md
│   │   ├── explain.md
│   │   ├── pr.md
│   │   ├── refactor.md
│   │   ├── review.md
│   │   └── test.md
│   ├── scripts/              # Cognitive scripts
│   │   ├── context-router.py
│   │   ├── memory-manager.py
│   │   ├── pool-extractor.py
│   │   ├── pool-loader.py
│   │   └── pool-query.py
│   ├── keywords.json         # Keyword activation config
│   └── settings.json         # Hooks configuration
├── scripts/
│   └── claude-agents-cli.sh  # CLI installer
└── templates/                # Reusable templates
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
