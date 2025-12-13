# Infrastructure Domain Knowledge

This document consolidates RAG knowledge for the **infra** super scanner.

## Features Covered
- **iac**: Infrastructure as Code security
- **containers**: Container image security
- **github_actions**: GitHub Actions security
- **dora**: DORA metrics calculation
- **git**: Git repository insights

## Related RAG Directories

### DORA Metrics
- `rag/dora/` - DORA metrics knowledge
- `rag/dora-metrics/` - Detailed metrics guidance
  - Deployment frequency
  - Lead time for changes
  - Change failure rate
  - Mean time to recovery

### Architecture
- `rag/architecture/` - System design knowledge
  - Infrastructure patterns
  - Cloud architecture
  - Container orchestration

### Technology Identification
- `rag/technology-identification/` - Technology detection
  - `cloud-providers/` - AWS, GCP, Azure patterns
  - CI/CD tool detection

## Key Concepts

### Infrastructure as Code (IaC) Security

#### Terraform Security Checks
- Hardcoded secrets in configuration
- Public S3 buckets
- Unrestricted security groups
- Missing encryption at rest
- IAM over-permissive policies

#### Kubernetes Security Checks
- Privileged containers
- Missing resource limits
- Root user containers
- Exposed sensitive data in ConfigMaps
- Missing network policies

#### CloudFormation Security Checks
- Public access configurations
- Missing encryption
- Overly permissive IAM roles
- Security group misconfigurations

### Container Security

#### Dockerfile Best Practices
- Use specific base image tags (not `latest`)
- Run as non-root user
- Minimize layers
- Don't expose unnecessary ports
- Use multi-stage builds
- Scan for vulnerabilities

#### Image Vulnerabilities
- OS package vulnerabilities
- Application dependencies
- Base image vulnerabilities
- Configuration issues

### GitHub Actions Security

#### Common Vulnerabilities
1. **Unpinned Actions** - Use SHA pins, not version tags
2. **Script Injection** - Don't use `${{ github.event.* }}` in run commands
3. **Secret Exposure** - Don't echo secrets or use in URLs
4. **Excessive Permissions** - Use minimum required permissions
5. **Self-hosted Runner Risks** - Protect from privilege escalation

#### Severity Mapping
| Issue | Severity |
|-------|----------|
| Script injection vulnerability | Critical |
| Secrets in workflow logs | High |
| Unpinned to SHA | Medium |
| Excessive permissions | Medium |
| Missing CODEOWNERS protection | Low |

### DORA Metrics

#### Four Key Metrics
| Metric | Elite | High | Medium | Low |
|--------|-------|------|--------|-----|
| Deployment Frequency | On-demand (multiple/day) | Weekly-Monthly | Monthly-Quarterly | > Quarterly |
| Lead Time for Changes | < 1 day | 1 day - 1 week | 1 week - 1 month | > 1 month |
| Change Failure Rate | 0-15% | 16-30% | 31-45% | > 45% |
| MTTR | < 1 hour | < 1 day | < 1 week | > 1 week |

#### Calculation Methods
- **Deployment Frequency**: Count releases/tags in period
- **Lead Time**: Time from first commit to release
- **Change Failure Rate**: Fix releases / total releases
- **MTTR**: Time between failure and fix release

### Git Insights

#### Health Indicators
- **Bus Factor**: Number of key contributors (higher is better)
- **Activity Level**: Commit frequency trends
- **Contributor Distribution**: Code ownership balance
- **Code Churn**: Frequently modified files (potential hotspots)
- **Code Age**: Distribution of last-modified times

## Agent Expertise

### Plague Agent
The **Plague** agent (DevOps specialist) should be consulted for:
- IaC security findings
- Container security issues
- Kubernetes configuration review
- Infrastructure hardening

### Joey Agent
The **Joey** agent (build engineer) should be consulted for:
- GitHub Actions security
- CI/CD pipeline optimization
- Build configuration review

### Gibson Agent
The **Gibson** agent (engineering leader) should be consulted for:
- DORA metrics interpretation
- Team health assessment
- Engineering KPIs

## Output Schema

The infra scanner produces a single `infra.json` file with:
```json
{
  "features_run": ["iac", "containers", "github_actions", "dora", "git"],
  "summary": {
    "iac": { "total_findings": N, "critical": N, ... },
    "containers": { "total_findings": N, "dockerfiles_scanned": N, ... },
    "github_actions": { "total_findings": N, "workflows_scanned": N, ... },
    "dora": { "deployment_frequency": N, "overall_class": "elite|high|medium|low", ... },
    "git": { "total_commits": N, "bus_factor": N, ... }
  },
  "findings": {
    "iac": [...],
    "containers": [...],
    "github_actions": [...],
    "dora": { "deployments": [...], ... },
    "git": { "contributors": [...], ... }
  }
}
```

## Severity Classification

| Finding Type | Critical | High | Medium | Low |
|--------------|----------|------|--------|-----|
| IaC | Public data exposure | Hardcoded secrets | Missing encryption | Best practice |
| Containers | CVE Critical | CVE High, root user | CVE Medium, unpinned | CVE Low |
| GitHub Actions | Script injection | Secret exposure | Unpinned actions | Permissions |
| DORA | N/A | Low performer | Medium performer | - |
| Git | N/A | Bus factor = 1 | Low activity | - |

## Detection Tools

### IaC Security
- **Checkov**: Primary tool for Terraform, K8s, CloudFormation
- **Trivy**: Fallback for IaC and container scanning

### Container Security
- **Trivy**: Image vulnerability scanning
- **Dockerfile linting**: Best practice checks

### GitHub Actions
- **Custom Analysis**: Pattern-based security checks
- **Semgrep**: YAML security rules
