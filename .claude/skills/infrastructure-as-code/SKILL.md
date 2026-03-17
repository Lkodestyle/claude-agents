---
name: infrastructure-as-code
description: Use when creating cloud infrastructure with Terraform, designing networking, or setting up multi-environment IaC. Orchestrates terraform, networking, cloud, security, and architecture agents for complete infrastructure workflows. MUST BE USED when user mentions Terraform modules, VPC design, multi-environment infrastructure, or IaC from scratch.
---

# Infrastructure as Code — Complete IaC Workflow

Design and implement cloud infrastructure using Terraform with proper networking, security, and multi-environment support.

## When to Use

- Creating cloud infrastructure from scratch
- Designing VPC/networking architecture
- Setting up multi-environment Terraform (dev/staging/prod)
- Creating reusable Terraform modules
- Migrating existing infrastructure to IaC

## Agents Activated

| Agent | Role |
|-------|------|
| `terraform` | HCL code, modules, state management, best practices |
| `networking` | VPC/VNet design, subnets, CIDR planning, security groups |
| `aws` / `azure` | Cloud-specific resources and services |
| `security` | IAM policies, encryption, compliance |
| `architecture` | Overall design, trade-offs, component placement |

## Workflow

### Step 1: Architecture Design

Using the `architecture` agent:

1. Understand the target architecture:
   - What services/applications will run?
   - Expected traffic and scale requirements
   - Compliance requirements (PCI, HIPAA, SOC2)
   - Disaster recovery needs
2. Ask the user for:
   - Cloud provider (AWS / Azure / multi-cloud)
   - Environments (dev, staging, prod)
   - Region strategy (single, multi-region, DR)
   - State backend preference (S3+DynamoDB / Azure Storage / Terraform Cloud)

### Step 2: Network Design

Using the `networking` agent:

1. Design VPC/VNet architecture:
   - CIDR block planning (non-overlapping across environments)
   - Subnet strategy (public, private, data, management)
   - Availability zone distribution
2. Configure network security:
   - Security groups / NSGs with least-privilege rules
   - NACLs for subnet-level control
   - NAT gateways for private subnet internet access
3. Plan connectivity:
   - VPN / peering requirements
   - Load balancer placement (ALB/NLB/App Gateway)
   - DNS configuration (Route53 / Azure DNS)

### Step 3: Terraform Structure

Using the `terraform` agent:

1. Set up project structure:
   ```
   infra/
   ├── modules/          # Reusable modules
   │   ├── networking/
   │   ├── compute/
   │   ├── database/
   │   └── security/
   ├── environments/
   │   ├── dev/
   │   ├── staging/
   │   └── prod/
   ├── backend.tf
   ├── providers.tf
   └── versions.tf
   ```
2. Configure remote state backend
3. Set up provider configuration with version pinning
4. Create variable definitions and tfvars per environment

### Step 4: Module Development

Using the `terraform` agent:

1. **Networking module**: VPC, subnets, route tables, NAT, IGW
2. **Compute module**: ECS/EKS/App Service/VM configurations
3. **Database module**: RDS/Aurora/Cosmos DB with encryption and backups
4. **Security module**: IAM roles, policies, KMS keys, secrets
5. Each module follows:
   - `main.tf` — resources
   - `variables.tf` — inputs with descriptions and validation
   - `outputs.tf` — exported values for composition
   - `README.md` — module documentation

### Step 5: Security Hardening

Using the `security` agent:

1. IAM: Least-privilege policies, no wildcard permissions
2. Encryption: At-rest (KMS/CMK) and in-transit (TLS)
3. Secrets: No hardcoded values, use secret manager references
4. Tagging: Mandatory tags strategy (environment, owner, cost-center)
5. Compliance checks: tfsec / checkov rules

### Step 6: Validation and Documentation

1. Generate `terraform plan` commands per environment
2. Create CI pipeline for Terraform (plan on PR, apply on merge)
3. Generate documentation:
   - Architecture diagram (Mermaid)
   - Network diagram with CIDR blocks
   - Module dependency graph
   - Runbook for common operations

## Output Files

```
infra/
├── modules/
│   ├── networking/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── compute/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── database/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── security/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
├── environments/
│   ├── dev/
│   │   ├── main.tf
│   │   ├── terraform.tfvars
│   │   └── backend.hcl
│   ├── staging/
│   │   ├── main.tf
│   │   ├── terraform.tfvars
│   │   └── backend.hcl
│   └── prod/
│       ├── main.tf
│       ├── terraform.tfvars
│       └── backend.hcl
├── providers.tf
├── versions.tf
└── docs/
    ├── architecture.md
    └── network-diagram.md
```

## Key Principles

- **DRY**: Reusable modules, environment-specific only in tfvars
- **State isolation**: Separate state file per environment
- **Version pinning**: Pin provider and module versions
- **No hardcoded values**: Everything through variables
- **Least privilege**: Minimal IAM permissions per resource
- **Tagging strategy**: Every resource tagged for cost tracking and ownership
- **Plan before apply**: Always review terraform plan output
