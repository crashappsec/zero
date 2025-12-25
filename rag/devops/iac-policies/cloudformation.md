# CloudFormation Security Patterns

**Category**: devops/iac-policies
**Description**: AWS CloudFormation security and organizational policy patterns
**CWE**: CWE-732 (Incorrect Permission Assignment), CWE-311 (Missing Encryption)

---

## S3 Bucket Patterns

### Public S3 Bucket ACL
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)AccessControl:\s*(?:PublicRead|PublicReadWrite)`
- S3 buckets should not have public ACLs
- Example: `AccessControl: PublicRead`
- Remediation: Use `AccessControl: Private` and bucket policies

### S3 Bucket Without Encryption
**Type**: regex
**Severity**: high
**Pattern**: `(?i)Type:\s*AWS::S3::Bucket(?:(?!BucketEncryption).)*$`
- S3 buckets should have encryption enabled
- Example: S3 bucket without BucketEncryption property
- Remediation: Add BucketEncryption with AES256 or aws:kms

### S3 Public Access Not Blocked
**Type**: regex
**Severity**: high
**Pattern**: `(?i)BlockPublicAcls:\s*false`
- S3 buckets should block public access
- Example: `BlockPublicAcls: false`
- Remediation: Set all PublicAccessBlockConfiguration options to true

---

## Security Group Patterns

### Security Group Open to Internet
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)CidrIp:\s*0\.0\.0\.0/0`
- Security groups should not allow unrestricted access
- Example: `CidrIp: 0.0.0.0/0`
- Remediation: Restrict to specific CIDR ranges

### SSH Open to Internet
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)FromPort:\s*22[\s\S]*?CidrIp:\s*0\.0\.0\.0/0`
- SSH should not be open to the internet
- Example: Port 22 with unrestricted CIDR
- Remediation: Use bastion hosts or VPN

### RDP Open to Internet
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)FromPort:\s*3389[\s\S]*?CidrIp:\s*0\.0\.0\.0/0`
- RDP should not be open to the internet
- Example: Port 3389 with unrestricted CIDR
- Remediation: Use bastion hosts or VPN

---

## RDS Patterns

### RDS Publicly Accessible
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)PubliclyAccessible:\s*(?:true|'true'|"true")`
- RDS instances should not be publicly accessible
- Example: `PubliclyAccessible: true`
- Remediation: Set `PubliclyAccessible: false`

### RDS Without Encryption
**Type**: regex
**Severity**: high
**Pattern**: `(?i)Type:\s*AWS::RDS::DBInstance(?:(?!StorageEncrypted).)*$`
- RDS instances should have storage encryption
- Example: RDS instance without StorageEncrypted
- Remediation: Add `StorageEncrypted: true`

### RDS Without Multi-AZ
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)MultiAZ:\s*(?:false|'false'|"false")`
- Production RDS instances should use Multi-AZ
- Example: `MultiAZ: false`
- Remediation: Set `MultiAZ: true` for high availability

---

## IAM Patterns

### Wildcard IAM Action
**Type**: regex
**Severity**: high
**Pattern**: `(?i)Action:\s*\*`
- IAM policies should not use wildcard actions
- Example: `Action: "*"`
- Remediation: Specify explicit actions

### Wildcard IAM Resource
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)Resource:\s*\*`
- IAM policies should not grant access to all resources
- Example: `Resource: "*"`
- Remediation: Specify resource ARNs

### IAM User With Console Access
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)Type:\s*AWS::IAM::User[\s\S]*?LoginProfile:`
- IAM users with console access should use SSO instead
- Example: IAM User with LoginProfile
- Remediation: Use AWS SSO or federated access

---

## EC2 Patterns

### EC2 Without IMDSv2
**Type**: regex
**Severity**: high
**Pattern**: `(?i)HttpTokens:\s*optional`
- EC2 instances should require IMDSv2
- Example: `HttpTokens: optional`
- Remediation: Set `HttpTokens: required`

### Unencrypted EBS Volume
**Type**: regex
**Severity**: high
**Pattern**: `(?i)Encrypted:\s*(?:false|'false'|"false")`
- EBS volumes should be encrypted
- Example: `Encrypted: false`
- Remediation: Set `Encrypted: true`

---

## Lambda Patterns

### Lambda Without VPC
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)Type:\s*AWS::Lambda::Function(?:(?!VpcConfig).)*$`
- Lambda functions accessing private resources should be in VPC
- Example: Lambda without VpcConfig
- Remediation: Add VpcConfig with appropriate subnets

### Lambda Timeout Too High
**Type**: regex
**Severity**: low
**Pattern**: `(?i)Timeout:\s*(?:[6-9][0-9]{2}|[1-9][0-9]{3,})`
- Lambda timeout should be appropriate for the function
- Example: `Timeout: 900`
- Remediation: Set reasonable timeout based on function needs

---

## Logging and Monitoring

### CloudTrail Not Enabled
**Type**: regex
**Severity**: high
**Pattern**: `(?i)IsLogging:\s*(?:false|'false'|"false")`
- CloudTrail should be enabled
- Example: `IsLogging: false`
- Remediation: Set `IsLogging: true`

### CloudWatch Logs Not Encrypted
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)Type:\s*AWS::Logs::LogGroup(?:(?!KmsKeyId).)*$`
- CloudWatch Logs should be encrypted with KMS
- Example: LogGroup without KmsKeyId
- Remediation: Add KmsKeyId property

---

## Organizational Policies

### Missing DeletionPolicy
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)Type:\s*AWS::(?:RDS::DBInstance|S3::Bucket)(?:(?!DeletionPolicy).)*$`
- Critical resources should have DeletionPolicy
- Example: RDS or S3 without DeletionPolicy
- Remediation: Add `DeletionPolicy: Retain` or `Snapshot`

### Missing Tags
**Type**: regex
**Severity**: low
**Pattern**: `(?i)Type:\s*AWS::(?:(?!Tags).)*$`
- Resources should be tagged for management
- Example: Resource without Tags property
- Remediation: Add Tags including Environment, Owner, Project

### Using Default KMS Key
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)KmsKeyId:\s*(?:alias/aws/|aws/)`
- Use customer-managed KMS keys for better control
- Example: `KmsKeyId: alias/aws/s3`
- Remediation: Create and use customer-managed KMS key

---

## Detection Confidence

**Regex Detection**: 85%
**Policy Compliance**: 90%

---

## References

- CIS AWS Foundations Benchmark
- AWS Well-Architected Framework
- CloudFormation Security Best Practices
