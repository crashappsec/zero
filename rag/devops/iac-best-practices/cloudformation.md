# CloudFormation Best Practices Patterns

**Category**: devops/iac-best-practices
**Description**: AWS CloudFormation template organizational and operational best practices
**Type**: best-practice

---

## Template Metadata

### Missing Description
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `AWSTemplateFormatVersion[^\n]+\n(?:(?!Description:).)*Resources:`
- Templates should have a Description for documentation
- Remediation: Add `Description: "Purpose of this template"`

### Outdated Template Version
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `AWSTemplateFormatVersion:\s*["']?2010-09-09["']?`
- Consider using latest CloudFormation features
- Note: 2010-09-09 is still the only valid version

---

## Parameter Best Practices

### Parameter Without Description
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `Parameters:\s*\n\s+\w+:\s*\n\s+Type:[^\n]+\n(?:(?!Description:).)*\n\s+\w+:`
- Parameters should have descriptions
- Remediation: Add `Description: "What this parameter controls"`

### Parameter Without Default
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `Parameters:\s*\n\s+\w+:\s*\n\s+Type:[^\n]+\n(?:(?!Default:).)*\n\s+\w+:`
- Consider providing sensible defaults for optional parameters
- Remediation: Add `Default: <value>` where appropriate

### Parameter Without Constraints
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `Type:\s*String\s*\n(?:(?!AllowedPattern:|AllowedValues:).)*\n\s+\w+:`
- String parameters should have validation constraints
- Remediation: Add `AllowedPattern` or `AllowedValues`

---

## Tagging Best Practices

### Resource Without Tags
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `Type:\s*AWS::[^\n]+\n\s+Properties:\s*\n(?:(?!Tags:).)*\n\s+\w+:`
- AWS resources should be tagged for organization
- Remediation: Add `Tags:` with Environment, Owner, CostCenter

---

## Output Best Practices

### Missing Outputs Section
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `Resources:\s*\n(?:(?!Outputs:).)*$`
- Templates should export useful values via Outputs
- Remediation: Add `Outputs:` section with key resource references

### Output Without Description
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `Outputs:\s*\n\s+\w+:\s*\n\s+Value:[^\n]+\n(?:(?!Description:).)*\n\s+\w+:`
- Outputs should have descriptions
- Remediation: Add `Description: "What this output provides"`

### Output Without Export
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `Outputs:\s*\n\s+\w+:\s*\n\s+Value:[^\n]+\n(?:(?!Export:).)*\n\s+\w+:`
- Consider exporting outputs for cross-stack references
- Remediation: Add `Export: { Name: !Sub "${AWS::StackName}-<name>" }`

---

## Naming Conventions

### Hardcoded Resource Names
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `(?:BucketName|TableName|FunctionName):\s*["'][^$!][^"']+["']`
- Avoid hardcoded names; use dynamic naming with stack name
- Example: `BucketName: "my-bucket"` (bad)
- Remediation: Use `!Sub "${AWS::StackName}-bucket"` (good)
