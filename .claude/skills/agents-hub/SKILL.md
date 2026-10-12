---
name: agents-hub
description: Use at the START of ANY DevOps, Cloud, IaC, backend, frontend, mobile, database, security, or general software task to select and orchestrate the right specialized agent. The canonical routing guide for the Claude DevOps Agents Hub. MUST BE USED to decide which of the 15 specialized agents (terraform, aws, azure, kubernetes, docker, cicd, databases, monitoring, networking, architecture, programming, web, mobile, finops, security) to invoke, how to fan work out to them in parallel with the Agent tool, and when to escalate to an end-to-end workflow skill.
---

# Agents Hub — Routing & Orchestration Guide

This is the **entry point** for working with the Claude DevOps Agents Hub. Use it to
route any task to the correct specialized agent(s), to run independent work **in
parallel**, and to do all of that **proactively**, without waiting for the user to name
an agent.

## When to Use

- At the start of **any** technical task (DevOps, Cloud, IaC, web, mobile, DB, security, code).
- When a task spans multiple domains and you need to decide which agents to combine.
- When a task has independent parts that could run at the same time.
- When unsure whether to use a single agent or escalate to a full workflow skill.

## Core Rule: Detect, then Delegate

**Detect the context first, then act.** Scan the request for the trigger keywords below
(the Context Router hook also prints a `[context-router] Suggested agents: ...` line on
every prompt) and delegate to the matching agent(s) with the `Agent` tool. Do **not**
wait for the user to say "use the X agent".

```text
Agent(subagent_type="terraform", prompt="...", description="IaC task")
Agent(subagent_type="aws",       prompt="...", description="Cloud task")
Agent(subagent_type="web",       prompt="...", description="Frontend/backend task")
```

If multiple agents apply, pick **one primary** (highest relevance) and add **co-activated**
agents as support (see the co-activation map).

## Parallel Work: the Main Instance Dispatches

The instance talking to the user is the orchestrator. It splits the work, launches
agents, keeps working while they run, and relays what matters when they report back.

| Situation | Do this |
|-----------|---------|
| Two or more tasks that **don't depend on each other** | Launch all the `Agent` calls **in the same message**. They run concurrently in the background; their reports arrive as notifications. |
| Task B needs the result of task A | Run A, wait for its report, then run B. |
| The agent needs **everything you already know** (long conversation, decisions made) | `subagent_type="fork"`: a clone with your full context, same model. |
| The agent needs **expert instructions and a clean context** | A specialized agent (`terraform`, `web`, `security`...): it gets only your prompt plus its own expert prompt. |
| Two agents will **edit the same repo** at the same time | `isolation="worktree"`: each one works in its own git worktree, nothing gets overwritten. |
| Broad research / audit / comparison across many files | Ask for a **report** (read-only): the agent reads, you decide. Keeps file dumps out of your context. |
| Implementation | Give a **closed, verifiable** task: what to change, how to check it (tests, `terraform plan`, lint, build). |
| Large orchestration (N agents, review → verify pipelines) | The `Workflow` tool, **only** when the user asked for it ("use a workflow", "ultracode"). |

Rules of thumb:

- Once you delegate something, **don't also do it yourself**; wait for the result.
- Give each agent a 3-5 word `description` and a self-contained `prompt` (it does not see
  the conversation unless it is a fork).
- Budget: one agent per independent concern. Three focused agents beat one that does everything.
- The agent's report is not shown to the user: **relay the conclusion**, not the transcript.

Example — "audit this repo":

```text
Agent(subagent_type="programming", description="Audit content layer",
      prompt="READ-ONLY. Audit .claude/agents, skills, commands: frontmatter validity, stale refs, overlaps. Report under 600 words.")
Agent(subagent_type="programming", description="Audit installer scripts",
      prompt="READ-ONLY. Audit scripts/*.sh and the MCP configs: bash -n, what install writes, portability. Report under 600 words.")
# both in ONE message → run in parallel while you test the router by hand
```

## Agent Catalog (15)

| Agent | Use For | Trigger Keywords |
|-------|---------|------------------|
| `architecture` | System design, C4 diagrams, ADRs, trade-offs | architecture, arquitectura, microservices, monolith, ddd, c4, adr, scalability |
| `terraform` | Terraform/Terragrunt/Terraspace, HCL, IaC modules | terraform, hcl, tfvars, terragrunt, terraspace, iac |
| `aws` | AWS services (EC2, ECS, EKS, Lambda, S3, RDS, DynamoDB, IAM) | aws, ec2, ecs, eks, lambda, s3, rds, dynamodb, cloudwatch, api gateway |
| `azure` | Azure (App Service, AKS, Functions, Cosmos DB, Bicep) | azure, app service, aks, azure functions, cosmos db, bicep, key vault |
| `kubernetes` | K8s, Helm, Kustomize, operators, troubleshooting | kubernetes, k8s, kubectl, helm, pod, ingress, hpa, argocd, cluster |
| `docker` | Dockerfiles, multi-stage builds, Compose, image security | docker, dockerfile, container, contenedor, compose, distroless |
| `cicd` | GitHub Actions, GitLab CI, Jenkins, pipelines, releases | cicd, ci/cd, pipeline, github actions, gitlab ci, deploy, despliegue, release, rollback |
| `databases` | PostgreSQL, MySQL, DynamoDB, Redis, MongoDB, migrations, ORMs | database, base de datos, postgres, mysql, redis, mongodb, migration, schema, orm |
| `monitoring` | Prometheus, Grafana, CloudWatch, SLOs, alerting, tracing | monitoring, monitoreo, prometheus, grafana, alert, metric, dashboard, slo, logs |
| `networking` | VPCs, subnets, CIDR, security groups, load balancers, DNS | network, vpc, subnet, cidr, security group, alb, nlb, dns, route53, vpn |
| `programming` | Clean code, design patterns, testing, code review, refactor | code review, refactor, clean code, solid, design pattern, unit test, tdd |
| `web` | React, Next.js, Vue, Node.js, FastAPI, Django, Nginx | react, next.js, vue, typescript, node.js, express, fastapi, frontend, backend |
| `mobile` | React Native, Expo, Flutter, iOS, Android, MVPs | mobile, movil, react native, expo, flutter, ios, android, swift, kotlin |
| `finops` | Cost optimization, rightsizing, Savings Plans, budgets | cost, costos, finops, billing, budget, rightsizing, savings plan |
| `security` | CVE scanning, OWASP, secrets detection, IAM, compliance | security, seguridad, vulnerability, cve, owasp, secrets, iam, trivy, gitleaks |

## Co-Activation Map

When the **primary** agent activates, also consider these support agents (mirrors
`~/.claude/keywords.json`):

| Primary | Co-activated support agents |
|---------|-----------------------------|
| `terraform` | `aws`, `azure`, `networking` |
| `aws` / `azure` | `terraform`, `networking` |
| `kubernetes` | `docker`, `monitoring` |
| `docker` | `cicd` |
| `cicd` | `docker` |
| `databases` | `programming` |
| `monitoring` | `kubernetes` |
| `security` | `cicd`, `docker` |
| `mobile` | `web` |

Bring `architecture` in for any non-trivial design decision, even if no keyword fired.

## Routing Workflow

### Step 1: Classify the request
Identify the dominant domain from the trigger keywords. Examples:
- "deploy a pod with Terraform" → primary `kubernetes` + `terraform`, support `docker`, `networking`.
- "MVP app in React Native" → primary `mobile`, support `web`, `programming`, `databases`.
- "optimize my AWS bill" → primary `finops`, support `aws`.

### Step 2: Split
List the independent parts. Each one is a candidate for its own agent, launched together.

### Step 3: Decide agents vs workflow skill
- **One or a few focused tasks** → `Agent` calls directly.
- **End-to-end, multi-stage goal** → escalate to a workflow skill (table below).

### Step 4: Delegate proactively
Launch the agents (in parallel when independent). State briefly which agent(s) you are
using and why.

### Step 5: Verify
Give the result a way to be checked (tests, linters, `terraform plan`, browser).
Optionally chain with `superpowers:verification-before-completion`.

## Escalation: Workflow Skills

For complete, multi-agent flows, escalate to one of the 5 workflow skills instead of
wiring agents by hand. The Context Router suggests them (`skill: ...`) when the anchor
agent plus another of its agents are active.

| Goal | Skill | Agents orchestrated |
|------|-------|---------------------|
| CI/CD + Docker + cloud deploy | `deploy-pipeline` | cicd, docker, aws/azure, networking, security |
| Metrics + logging + alerting + SLOs | `observability-setup` | monitoring, aws/azure, kubernetes |
| VPC + Terraform modules + multi-env | `infrastructure-as-code` | terraform, networking, aws/azure, security, architecture |
| Security audit + CVE fix + hardening | `security-hardening` | security, docker, cicd, aws/azure, kubernetes |
| New project: API + DB + frontend + Docker | `full-stack-scaffold` | architecture, web/mobile, databases, programming, docker |

Recommended chain with the `superpowers` plugin:

```text
superpowers:brainstorming → refine requirements
    ↓
superpowers:writing-plans → implementation plan
    ↓
agents-hub → route / split / delegate
    ↓
Workflow Skill or direct agents → execute (in parallel when possible)
    ↓
superpowers:verification-before-completion → validate
```

## Global Install Notes

This skill lives in a **global** install (`~/.claude/skills/agents-hub/`):

- Install everything (agents, commands, skills, scripts) with:
  ```bash
  ./scripts/claude-agents-cli.sh install            # symlink (auto-updates on git pull)
  ./scripts/claude-agents-cli.sh install --copy     # independent copy
  ```
- The **Context Router** hook (`~/.claude/scripts/context-router.py`) prints one
  `[context-router]` line per prompt based on `~/.claude/keywords.json`. This skill
  gives you the routing and orchestration logic when you act on it.
- The **Pool Coordinator** is for several Claude *terminals* open at once (opt-in via
  `CLAUDE_INSTANCE`); it is not how you run agents in parallel.
- Verify with `./scripts/claude-agents-cli.sh status`.

## Key Principles

- **Detect, then delegate**: never wait for the user to name an agent.
- **Independent work runs in parallel**: same message, several `Agent` calls.
- **One primary, few supports**: avoid activating everything at once.
- **Design-aware**: keep `architecture` in the loop for non-trivial work.
- **Escalate deliberately**: use a workflow skill when the goal is end-to-end.
- **Verify before done**: give every result a way to be checked.
- **Security first**: never expose secrets or credentials in any agent output.
