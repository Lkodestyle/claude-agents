---
name: security-hardening
description: Use when auditing security, scanning vulnerabilities, hardening containers or infrastructure, or implementing compliance controls. Orchestrates security, docker, cicd, and cloud agents for comprehensive security workflows. MUST BE USED when user mentions security audit, vulnerability scanning, CVE remediation, secrets detection, or compliance hardening.
---

# Security Hardening — Comprehensive Security Audit and Fix

Run a complete security audit across code, containers, infrastructure, and CI/CD pipelines, then apply remediations.

## When to Use

- Security audit of an application or infrastructure
- Vulnerability scanning and CVE remediation
- Hardening Docker images and Kubernetes deployments
- Detecting and rotating exposed secrets
- Adding security steps to CI/CD pipelines
- Compliance preparation (SOC2, PCI-DSS, HIPAA)

## Agents Activated

| Agent | Role |
|-------|------|
| `security` | Vulnerability scanning, OWASP, compliance, secrets detection |
| `docker` | Container image hardening, distroless, non-root |
| `cicd` | Pipeline security steps, SAST/DAST integration |
| `aws` / `azure` | IAM policies, security groups, encryption, cloud security |
| `kubernetes` | Pod security, network policies, RBAC (if applicable) |

## Workflow

### Step 1: Reconnaissance

1. Identify the scope:
   - Application code (languages, frameworks, dependencies)
   - Container images (Dockerfiles, base images)
   - Infrastructure (cloud provider, IaC files)
   - CI/CD pipelines (workflows, deployment scripts)
2. Ask the user for:
   - Compliance requirements (SOC2, PCI, HIPAA, or none)
   - Known security concerns
   - Priority: fix critical only vs comprehensive hardening

### Step 2: Dependency Audit

Using the `security` agent:

1. Scan dependencies for known CVEs:
   - `npm audit` / `pip audit` / `bundle audit` / `cargo audit`
   - Trivy filesystem scan: `trivy fs --severity HIGH,CRITICAL .`
   - Snyk test (if available): `snyk test`
2. Generate vulnerability report with:
   - CVE ID, severity, affected package, fixed version
   - Prioritized by severity and exploitability
3. Apply fixes:
   - Update dependencies to patched versions
   - Document any accepted risks for unfixable vulnerabilities

### Step 3: Secrets Detection

Using the `security` agent:

1. Scan for exposed secrets:
   - `gitleaks detect --source . --verbose`
   - Check for hardcoded credentials, API keys, tokens
   - Scan environment files, configs, and CI/CD variables
2. Remediate findings:
   - Remove secrets from code
   - Rotate compromised credentials
   - Move secrets to proper secret manager
   - Add `.gitignore` rules and pre-commit hooks

### Step 4: Container Hardening

Using the `docker` agent:

1. Audit Dockerfiles:
   - Use specific base image tags (no `latest`)
   - Multi-stage builds to minimize attack surface
   - Run as non-root user
   - No unnecessary packages or tools
   - Use distroless or alpine base images
2. Scan container images:
   - `trivy image <image>` for CVEs
   - Check for running as root
   - Verify no secrets baked into layers
3. Apply fixes and regenerate optimized Dockerfile

### Step 5: Infrastructure Security

Using the `aws` / `azure` and `security` agents:

1. **IAM audit**:
   - No wildcard (`*`) permissions
   - Principle of least privilege
   - No inline policies (use managed policies)
   - MFA enforcement for console access
2. **Network security**:
   - No public access to databases
   - Security groups: minimal ingress, deny by default
   - Encryption in transit (TLS everywhere)
3. **Data security**:
   - Encryption at rest for all storage (S3, RDS, EBS)
   - Backup configuration and retention
   - Access logging enabled

### Step 6: Pipeline Security

Using the `cicd` agent:

1. Add security steps to CI/CD pipeline:
   - SAST (Static Application Security Testing)
   - Dependency scanning
   - Container image scanning
   - Secrets scanning (prevent commits with secrets)
2. Configure security gates:
   - Block deploys with CRITICAL vulnerabilities
   - Require security review for HIGH findings
   - Allow WARNING with documented acceptance
3. Add pre-commit hooks:
   - `gitleaks` for secret detection
   - `tfsec` for Terraform security (if applicable)

### Step 7: Kubernetes Security (if applicable)

Using the `kubernetes` agent:

1. Pod Security Standards:
   - No privileged containers
   - Read-only root filesystem
   - Drop all capabilities, add only needed
   - Resource limits set
2. Network Policies: Restrict pod-to-pod communication
3. RBAC: Minimal ClusterRole bindings

### Step 8: Report and Documentation

1. Generate security report:
   - Findings summary (critical/high/medium/low counts)
   - Remediations applied
   - Accepted risks with justification
   - Recommendations for ongoing security
2. Create security runbook:
   - Incident response procedures
   - Secret rotation procedures
   - Vulnerability triage process

## Output Files

```
security/
├── reports/
│   ├── vulnerability-report.md    # CVE findings and status
│   ├── secrets-audit.md           # Secrets scan results
│   └── compliance-checklist.md    # Compliance status
├── configs/
│   ├── .gitleaks.toml             # Gitleaks config
│   ├── trivy.yaml                 # Trivy config
│   └── .pre-commit-config.yaml    # Pre-commit hooks
└── docs/
    └── security-runbook.md        # Security procedures
```

## Key Principles

- **Shift left**: Security checks as early as possible in the pipeline
- **Defense in depth**: Multiple layers of security controls
- **Least privilege**: Minimum permissions required for each component
- **Secrets never in code**: Use secret managers, environment variables, or sealed secrets
- **Scan continuously**: Security is not a one-time activity
- **Accept risk explicitly**: Document any accepted vulnerabilities with business justification
