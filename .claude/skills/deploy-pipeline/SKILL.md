---
name: deploy-pipeline
description: Use when setting up CI/CD pipelines, deployment workflows, or containerized deployments. Orchestrates cicd, docker, and cloud provider agents to create complete deployment pipelines. MUST BE USED when user mentions deploy, pipeline, CI/CD setup, GitHub Actions workflow creation, or containerized deployment to AWS/Azure.
---

# Deploy Pipeline — Complete Deployment Workflow

Orchestrate a full deployment pipeline from Dockerfile to cloud deployment, combining CI/CD, container, and cloud infrastructure agents.

## When to Use

- Setting up a new CI/CD pipeline for a project
- Containerizing an application for cloud deployment
- Creating GitHub Actions / GitLab CI workflows with cloud deployment
- Migrating from manual deploys to automated pipelines

## Agents Activated

| Agent | Role |
|-------|------|
| `cicd` | Pipeline definition (GitHub Actions, GitLab CI, Jenkins) |
| `docker` | Dockerfile creation, multi-stage builds, image optimization |
| `aws` / `azure` | Cloud deployment target (ECS, App Service, EKS, AKS) |
| `networking` | Load balancers, DNS, security groups |
| `security` | Image scanning, secrets management in pipeline |

## Workflow

### Step 1: Detect and Analyze

1. Scan the project to detect the stack (language, framework, package manager)
2. Identify existing Dockerfile, CI configs, or cloud infrastructure
3. Ask the user for:
   - CI/CD platform preference (GitHub Actions / GitLab CI / Jenkins)
   - Cloud provider (AWS / Azure)
   - Deployment target (ECS, EKS, App Service, AKS, Lambda)
   - Environments needed (dev, staging, prod)

### Step 2: Containerize

Using the `docker` agent:

1. Generate an optimized Dockerfile with multi-stage build
2. Create `.dockerignore` for the detected stack
3. Add health check endpoint configuration
4. Generate `docker-compose.yml` for local development

### Step 3: Create Pipeline

Using the `cicd` agent:

1. Generate the CI/CD workflow file:
   - Build and test stage
   - Docker image build and push to registry (ECR/ACR)
   - Deploy stage per environment
   - Rollback mechanism
2. Configure environment-specific variables and secrets
3. Add branch protection rules documentation

### Step 4: Cloud Infrastructure

Using the `aws` or `azure` agent:

1. Generate IaC for the deployment target:
   - **AWS ECS**: Task definition, service, ALB, target group
   - **AWS EKS**: Deployment, service, ingress manifests
   - **Azure App Service**: App service plan, web app, deployment slots
   - **Azure AKS**: Deployment, service, ingress manifests
2. Configure networking (security groups, NSGs, load balancer rules)
3. Set up container registry (ECR/ACR)

### Step 5: Security Hardening

Using the `security` agent:

1. Add container image scanning step to pipeline (Trivy/Snyk)
2. Configure secrets management (GitHub Secrets / Azure Key Vault / AWS Secrets Manager)
3. Add SAST/DAST steps if applicable

### Step 6: Documentation

1. Generate deployment README with:
   - Architecture diagram (Mermaid)
   - Environment setup instructions
   - Manual rollback procedure
   - Monitoring endpoints

## Output Files

```
project/
├── Dockerfile
├── .dockerignore
├── docker-compose.yml              # Local dev
├── .github/workflows/deploy.yml    # Or .gitlab-ci.yml
├── infra/
│   ├── main.tf                     # Or ARM/Bicep templates
│   ├── variables.tf
│   └── environments/
│       ├── dev.tfvars
│       ├── staging.tfvars
│       └── prod.tfvars
└── docs/
    └── deployment.md
```

## Key Principles

- **Immutable deployments**: Always deploy new containers, never modify running ones
- **Environment parity**: Dev/staging/prod should be as similar as possible
- **Rollback-first**: Every deployment must have a clear rollback path
- **Secrets never in code**: All secrets through environment variables or secret managers
- **Scan before deploy**: No image reaches production without vulnerability scanning
