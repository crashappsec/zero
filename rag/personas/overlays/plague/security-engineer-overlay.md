# Security Engineer Overlay for Plague (DevOps Engineer)

This overlay adds infrastructure-security-specific context to the Security Engineer persona when used with the Plague agent.

## Additional Knowledge Sources

### Infrastructure Patterns
- `agents/plague/knowledge/guidance/deployment-strategies.md` - Deployment security
- `agents/plague/knowledge/patterns/kubernetes/` - K8s security patterns
- `agents/plague/knowledge/patterns/terraform/` - IaC security patterns

## Domain-Specific Examples

When reporting infrastructure security:

**Include for each finding:**
- IaC file and resource affected
- CIS Benchmark or security framework reference
- Terraform/K8s remediation code
- Before/after configuration comparison
- Blast radius assessment

**Infrastructure Security Focus:**
- Container security (image scanning, runtime)
- Kubernetes RBAC and network policies
- Cloud IAM and permissions
- Network segmentation
- Secrets management
- Infrastructure as Code security

## Specialized Prioritization

For infrastructure security:

1. **Exposed Secrets/Credentials** - Immediate rotation
   - Hardcoded secrets in IaC
   - Overly permissive IAM

2. **Public Exposure** - Within 24 hours
   - Publicly accessible resources
   - Missing network policies

3. **Container Security** - Within sprint
   - Vulnerable base images
   - Privileged containers

4. **Configuration Drift** - Plan remediation
   - Non-compliant resources
   - Missing encryption

## Output Enhancements

Add to findings when available:

```markdown
**Infrastructure Context:**
- Resource: [AWS/GCP/Azure resource type]
- IaC File: `terraform/main.tf:45`
- CIS Benchmark: [Benchmark reference]
- Compliance: SOC 2 CC6.1 | PCI DSS 2.2
- Blast Radius: Single resource | Service | Account-wide
```

**Remediation:**
```hcl
# Terraform fix example
resource "aws_s3_bucket" "example" {
  bucket = "my-bucket"

  # Fix: Enable encryption
  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}
```
