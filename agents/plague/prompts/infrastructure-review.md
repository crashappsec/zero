# Infrastructure Review Prompt

## Context
You are reviewing infrastructure code (Terraform, Kubernetes, etc.) for security, reliability, and best practices.

## Review Areas

### Security
- [ ] No secrets in code
- [ ] Least privilege IAM policies
- [ ] Encryption at rest and in transit
- [ ] Network isolation configured
- [ ] No public access to internal resources

### Reliability
- [ ] High availability configured
- [ ] Resource limits defined
- [ ] Health checks in place
- [ ] Backup and recovery planned
- [ ] Disaster recovery considered

### Observability
- [ ] Logging enabled
- [ ] Metrics collected
- [ ] Alerting configured
- [ ] Tracing available

### Cost
- [ ] Right-sized resources
- [ ] Auto-scaling configured
- [ ] Unused resources identified
- [ ] Reserved capacity evaluated

### Maintainability
- [ ] Modules used for reusability
- [ ] Variables properly documented
- [ ] Version constraints defined
- [ ] Consistent naming conventions

## Output Format

```markdown
## Infrastructure Review: [Component Name]

### Security Assessment

| Check | Status | Notes |
|-------|--------|-------|
| No hardcoded secrets | Pass/Fail | |
| Least privilege | Pass/Fail | |
| Encryption | Pass/Fail | |
| Network isolation | Pass/Fail | |

### Critical Issues

1. **[Issue Name]** (severity: critical|high|medium|low)
   - Location: `path/to/file.tf:line`
   - Issue: Description
   - Risk: What could happen
   - Fix: How to remediate

### Warnings

1. **[Warning Name]**
   - Location: `path/to/file.tf:line`
   - Issue: Description
   - Recommendation: Suggested improvement

### Best Practice Recommendations

1. **[Recommendation]**
   - Current: What's there now
   - Suggested: What it should be
   - Why: Benefit

### Resource Summary

| Resource Type | Count | Notes |
|---------------|-------|-------|
| EC2 Instances | X | |
| RDS Databases | X | |
| S3 Buckets | X | |

### Cost Estimate

If calculable, provide rough cost estimate or savings opportunities.
```

## Example Output

```markdown
## Infrastructure Review: Production EKS Cluster

### Security Assessment

| Check | Status | Notes |
|-------|--------|-------|
| No hardcoded secrets | PASS | Using AWS Secrets Manager |
| Least privilege | FAIL | IAM role has wildcard permissions |
| Encryption | PASS | EBS and S3 encrypted |
| Network isolation | PASS | Private subnets, NACLs configured |

### Critical Issues

1. **Wildcard IAM Permissions** (severity: critical)
   - Location: `modules/eks/iam.tf:45`
   - Issue: Node IAM role has `"Resource": "*"` for S3 access
   - Risk: Compromised pods could access any S3 bucket in account
   - Fix: Scope to specific bucket ARNs:
     ```hcl
     resources = ["arn:aws:s3:::${var.bucket_name}/*"]
     ```

2. **Public Subnet for Nodes** (severity: high)
   - Location: `modules/eks/vpc.tf:23`
   - Issue: Worker nodes in public subnet with public IPs
   - Risk: Direct attack surface from internet
   - Fix: Use private subnets with NAT gateway

### Warnings

1. **No Pod Security Policy/Standards**
   - Location: `modules/eks/main.tf`
   - Issue: No pod security policy configured
   - Recommendation: Enable Pod Security Standards (baseline or restricted)

2. **Single NAT Gateway**
   - Location: `modules/vpc/nat.tf:12`
   - Issue: One NAT gateway for all AZs
   - Recommendation: Deploy NAT gateway per AZ for HA

### Best Practice Recommendations

1. **Enable Container Insights**
   - Current: CloudWatch agent not configured
   - Suggested: Enable Container Insights for metrics/logs
   - Why: Better visibility into pod-level resource usage

2. **Add Cluster Autoscaler**
   - Current: Fixed node count
   - Suggested: Deploy Cluster Autoscaler or Karpenter
   - Why: Cost optimization and automatic scaling

### Resource Summary

| Resource Type | Count | Notes |
|---------------|-------|-------|
| EKS Cluster | 1 | v1.28 |
| Node Groups | 2 | m5.large |
| Nodes | 6 | 3 per group |
| NAT Gateways | 1 | Single point of failure |
| Load Balancers | 3 | ALB per ingress |

### Cost Estimate

Current: ~$800/month
- 6x m5.large: $420
- NAT Gateway: $32 + data
- ALBs: $48 + data
- EBS: ~$50

Potential savings:
- Right-size nodes to m5.medium: -$210/month
- Use Spot for non-critical workloads: -$150/month
```
