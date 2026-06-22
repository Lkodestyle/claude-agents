---
name: agents-hub
description: Use at the START of ANY DevOps, Cloud, IaC, backend, frontend, mobile, database, security, or general software task to select and orchestrate the right specialized agent. The canonical routing guide for the Claude DevOps Agents Hub. MUST BE USED to decide which of the 15 specialized agents (terraform, aws, azure, kubernetes, docker, cicd, databases, monitoring, networking, architecture, programming, web, mobile, finops, security) to invoke, how to invoke them proactively via the Task tool, and when to escalate to an end-to-end workflow skill.
---

# Agents Hub — Specialized Agent Routing Guide

This is the **entry point** for working with the Claude DevOps Agents Hub. Use it to
route any task to the correct specialized agent(s) and to invoke them **proactively**,
without waiting for the user to name an agent.

## When to Use

- At the start of **any** technical task (DevOps, Cloud, IaC, web, mobile, DB, security, code).
- When a task spans multiple domains and you need to decide which agents to combine.
- When unsure whether to use a single agent or escalate to a full workflow skill.
- To confirm the canonical agent catalog after a global install (`~/.claude/`).

## Core Rule: Proactive Activation

**Detect the context first, then act.** Before answering or editing, scan the request
for the keywords below and activate the matching agent(s) automatically via the Task tool.
Do **not** wait for the user to say "use the X agent".

```text
Task(subagent_type="terraform", prompt="...", description="IaC task")
Task(subagent_type="aws",       prompt="...", description="Cloud task")
Task(subagent_type="web",       prompt="...", description="Frontend/backend task")
```

If multiple agents apply, pick **one primary** (highest relevance) and add **co-activated**
agents as support (see the co-activation map). Keep architecture in mind for any non-trivial design.

## Agent Catalog (15)

| Agent | Use For | Trigger Keywords |
|-------|---------|------------------|
| `architecture` | System design, C4 diagrams, ADRs, trade-offs (pinned/always relevant) | architecture, design, microservices, monolith, ddd, c4, adr, scalability |
| `terraform` | Terraform/Terragrunt/Terraspace, HCL, IaC modules | terraform, hcl, tfvars, state, plan, apply, module, provider, iac |
| `aws` | AWS services (EC2, ECS, EKS, Lambda, S3, RDS, DynamoDB, IAM) | aws, ec2, ecs, eks, lambda, s3, rds, dynamodb, cloudwatch, api gateway |
| `azure` | Azure (App Service, AKS, Functions, Cosmos DB, Bicep) | azure, app service, aks, azure functions, cosmos db, bicep, key vault |
| `kubernetes` | K8s, Helm, Kustomize, operators, troubleshooting | kubernetes, k8s, kubectl, helm, pod, deployment, ingress, hpa, argocd |
| `docker` | Dockerfiles, multi-stage builds, Compose, image security | docker, dockerfile, container, image, compose, registry, distroless |
| `cicd` | GitHub Actions, GitLab CI, Jenkins, pipelines, releases | cicd, ci/cd, pipeline, github actions, gitlab ci, deploy, release, rollback |
| `databases` | PostgreSQL, MySQL, DynamoDB, Redis, MongoDB, migrations, ORMs | database, postgres, mysql, dynamodb, redis, mongodb, migration, schema, orm |
| `monitoring` | Prometheus, Grafana, CloudWatch, SLOs, alerting, tracing | monitoring, prometheus, grafana, alert, metric, dashboard, slo, observability |
| `networking` | VPCs, subnets, CIDR, security groups, load balancers, DNS | network, vpc, subnet, cidr, security group, alb, nlb, dns, route53, vpn |
| `programming` | Clean code, design patterns, testing, code review, refactor | code review, refactor, clean code, solid, design pattern, tdd, rest, graphql |
| `web` | React, Next.js, Vue, Node.js, FastAPI, Django, Nginx | react, next.js, vue, typescript, node.js, express, fastapi, frontend, backend |
| `mobile` | React Native, Expo, Flutter, iOS, Android, MVPs | mobile, app, react native, expo, flutter, ios, android, swift, kotlin, mvp |
| `finops` | Cost optimization, rightsizing, Savings Plans, budgets | cost, finops, billing, budget, rightsizing, reserved instance, savings plan |
| `security` | CVE scanning, OWASP, secrets detection, IAM, compliance | security, vulnerability, cve, owasp, secret, iam, audit, trivy, snyk, gitleaks |

## Co-Activation Map

When the **primary** agent activates, also bring in these support agents (mirrors `~/.claude/keywords.json`):

| Primary | Co-activated support agents |
|---------|-----------------------------|
| `terraform` | `aws`, `azure`, `networking` |
| `aws` | `terraform`, `networking`, `finops` |
| `azure` | `terraform`, `networking`, `finops` |
| `kubernetes` | `docker`, `monitoring`, `networking` |
| `docker` | `kubernetes`, `cicd` |
| `cicd` | `docker`, `terraform` |
| `databases` | `programming`, `monitoring` |
| `monitoring` | `kubernetes`, `aws` |
| `architecture` | `programming`, `databases` |
| `security` | `cicd`, `docker`, `aws` |
| `mobile` | `web`, `programming`, `databases` |

`architecture` is **pinned**: keep it available for any non-trivial design decision.

## Routing Workflow

### Step 1: Classify the request
Identify the dominant domain from the trigger keywords. Examples:
- "deploy a pod with Terraform" → primary `kubernetes` + `terraform`, support `docker`, `networking`.
- "MVP app in React Native" → primary `mobile`, support `web`, `programming`, `databases`.
- "optimize my AWS bill" → primary `finops`, support `aws`.

### Step 2: Decide single agent vs workflow skill
- **Single/few agents** → invoke directly with the Task tool for focused tasks.
- **End-to-end, multi-stage goal** → escalate to a workflow skill (see table below).

### Step 3: Invoke proactively
Call the primary agent first; add co-activated agents as the task expands. State briefly
which agent(s) you are using and why.

### Step 4: Verify
Prefer giving the result a way to be verified (tests, linters, `terraform plan`, browser).
Optionally chain with `superpowers:verification-before-completion`.

## Escalation: Workflow Skills

For complete, multi-agent flows, escalate to one of the 5 workflow skills instead of
wiring agents by hand:

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
agents-hub → route to the right agent(s) / workflow skill
    ↓
Workflow Skill or direct agent → execute
    ↓
superpowers:verification-before-completion → validate
```

## Global Install Notes

This skill is designed to live in a **global** install (`~/.claude/skills/agents-hub/`):

- Install everything (agents, commands, skills, scripts) with:
  ```bash
  ./scripts/claude-agents-cli.sh install            # symlink (auto-updates on git pull)
  ./scripts/claude-agents-cli.sh install --copy     # independent copy
  ```
- The **Context Router** hook (`~/.claude/scripts/context-router.py`) injects agent
  context automatically based on `~/.claude/keywords.json`. This skill complements it by
  giving you the explicit routing/invocation logic when you act.
- Verify with `./scripts/claude-agents-cli.sh status` (lists agents, commands, skills).

## Key Principles

- **Detect, then act**: never wait for the user to name an agent.
- **One primary, several supports**: avoid activating everything at once.
- **Design-aware**: keep `architecture` in the loop for non-trivial work.
- **Escalate deliberately**: use a workflow skill when the goal is end-to-end.
- **Verify before done**: give every result a way to be checked.
- **Security first**: never expose secrets or credentials in any agent output.
