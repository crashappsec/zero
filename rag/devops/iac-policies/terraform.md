# Terraform Security Patterns

**Category**: devops/iac-policies
**Description**: Terraform security and organizational policy patterns
**CWE**: CWE-732 (Incorrect Permission Assignment), CWE-311 (Missing Encryption)

---

## Access Control Patterns

### Public S3 Bucket ACL
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)acl\s*=\s*["'](?:public-read|public-read-write)["']`
- S3 buckets should not have public ACLs
- Example: `acl = "public-read"`
- Remediation: Use `acl = "private"` and configure bucket policies for specific access

### S3 Bucket Without Encryption
**Type**: regex
**Severity**: high
**Pattern**: `(?i)resource\s*"aws_s3_bucket"\s*"[^"]+"\s*\{(?:(?!server_side_encryption_configuration).)*\}`
- S3 buckets should have server-side encryption enabled
- Example: Missing `server_side_encryption_configuration` block
- Remediation: Add encryption configuration with AES-256 or AWS KMS

### Public Security Group Ingress
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)cidr_blocks\s*=\s*\[\s*["']0\.0\.0\.0/0["']`
- Security groups should not allow unrestricted ingress from the internet
- Example: `cidr_blocks = ["0.0.0.0/0"]`
- Remediation: Restrict CIDR blocks to specific IP ranges or VPCs

### Open SSH Port to World
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)from_port\s*=\s*22[^0-9].*?cidr_blocks\s*=\s*\[\s*["']0\.0\.0\.0/0["']`
- SSH access should not be open to the entire internet
- Example: Port 22 with `cidr_blocks = ["0.0.0.0/0"]`
- Remediation: Restrict SSH access to known IP ranges or use bastion hosts

### Open RDP Port to World
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)from_port\s*=\s*3389[^0-9].*?cidr_blocks\s*=\s*\[\s*["']0\.0\.0\.0/0["']`
- RDP access should not be open to the entire internet
- Example: Port 3389 with `cidr_blocks = ["0.0.0.0/0"]`
- Remediation: Restrict RDP access to known IP ranges or use VPN

---

## Encryption Patterns

### Unencrypted EBS Volume
**Type**: regex
**Severity**: high
**Pattern**: `(?i)resource\s*"aws_ebs_volume".*?encrypted\s*=\s*false`
- EBS volumes should be encrypted at rest
- Example: `encrypted = false`
- Remediation: Set `encrypted = true` and configure KMS key

### RDS Without Encryption
**Type**: regex
**Severity**: high
**Pattern**: `(?i)resource\s*"aws_db_instance".*?storage_encrypted\s*=\s*false`
- RDS instances should have storage encryption enabled
- Example: `storage_encrypted = false`
- Remediation: Set `storage_encrypted = true`

### RDS Public Access
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)publicly_accessible\s*=\s*true`
- Database instances should not be publicly accessible
- Example: `publicly_accessible = true`
- Remediation: Set `publicly_accessible = false` and use VPC endpoints

### ELB Without SSL
**Type**: regex
**Severity**: high
**Pattern**: `(?i)resource\s*"aws_elb".*?listener\s*\{[^}]*lb_protocol\s*=\s*["']HTTP["']`
- Load balancers should use HTTPS for secure communication
- Example: `lb_protocol = "HTTP"`
- Remediation: Configure `lb_protocol = "HTTPS"` with SSL certificate

---

## Logging and Monitoring Patterns

### CloudTrail Not Enabled
**Type**: regex
**Severity**: high
**Pattern**: `(?i)enable_logging\s*=\s*false`
- CloudTrail logging should be enabled for audit purposes
- Example: `enable_logging = false`
- Remediation: Set `enable_logging = true`

### VPC Flow Logs Disabled
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)resource\s*"aws_vpc"(?:(?!aws_flow_log).)*$`
- VPCs should have flow logs enabled for network monitoring
- Example: VPC without associated flow log resource
- Remediation: Create `aws_flow_log` resource for the VPC

### S3 Access Logging Disabled
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)resource\s*"aws_s3_bucket"\s*"[^"]+"\s*\{(?:(?!logging).)*\}`
- S3 buckets should have access logging enabled
- Example: Missing `logging` block
- Remediation: Add `logging` block with target bucket

---

## IAM Patterns

### Wildcard IAM Action
**Type**: regex
**Severity**: high
**Pattern**: `(?i)"Action"\s*:\s*\[\s*["']\*["']`
- IAM policies should not grant wildcard actions
- Example: `"Action": ["*"]`
- Remediation: Specify required actions explicitly

### Wildcard IAM Resource
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)"Resource"\s*:\s*\[\s*["']\*["']`
- IAM policies should not grant access to all resources
- Example: `"Resource": ["*"]`
- Remediation: Specify specific resource ARNs

### Assume Role Without Condition
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)"Effect"\s*:\s*["']Allow["'].*?"Action"\s*:\s*\[\s*["']sts:AssumeRole["'](?:(?!"Condition").)*$`
- AssumeRole policies should have conditions for security
- Example: Allow AssumeRole without conditions
- Remediation: Add conditions like external ID or MFA requirement

---

## Network Patterns

### Missing VPC
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)resource\s*"aws_instance"(?:(?!vpc_security_group_ids).)*\}`
- EC2 instances should be deployed in a VPC with proper security groups
- Example: Instance without VPC security group configuration
- Remediation: Deploy in VPC with appropriate security groups

### Default Security Group Used
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)security_groups\s*=\s*\[\s*["']default["']`
- Default security groups should not be used
- Example: `security_groups = ["default"]`
- Remediation: Create and use custom security groups

---

## Organizational Policies

### Missing Tags
**Type**: regex
**Severity**: low
**Pattern**: `(?i)resource\s*"aws_[^"]+"\s*"[^"]+"\s*\{(?:(?!tags).)*\}`
- Resources should be tagged for cost allocation and management
- Example: Resource without `tags` block
- Remediation: Add tags including Owner, Environment, Project

### Resource Without Provider Region
**Type**: regex
**Severity**: low
**Pattern**: `(?i)provider\s*"aws"\s*\{(?:(?!region).)*\}`
- AWS provider should have explicit region configuration
- Example: Provider without region
- Remediation: Specify `region` in provider block

---

## Detection Confidence

**Regex Detection**: 85%
**Policy Compliance**: 90%

---

## References

- CIS AWS Foundations Benchmark
- AWS Well-Architected Framework
- Terraform Security Best Practices
