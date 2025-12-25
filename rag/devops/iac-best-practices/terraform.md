# Terraform Best Practices Patterns

**Category**: devops/iac-best-practices
**Description**: Terraform organizational and operational best practices
**Type**: best-practice

---

## Tagging Patterns

### Missing Required Tags
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `resource\s+"aws_[^"]+"\s+"[^"]+"\s*\{(?:(?!tags\s*=).)*\}`
- AWS resources should have tags for organization and cost tracking
- Example: Resource without `tags` block
- Remediation: Add `tags = { Environment = var.environment, Owner = var.owner }`

### Missing Environment Tag
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `tags\s*=\s*\{(?:(?!Environment).)*\}`
- All resources should be tagged with Environment
- Remediation: Add `Environment = var.environment` to tags block

---

## Module Best Practices

### Unpinned Module Version
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `source\s*=\s*"git::[^"]+(?<!\.git\?ref=[a-f0-9]+)"`
- Git-sourced modules should be pinned to a specific ref
- Example: `source = "git::https://example.com/module.git"`
- Remediation: Pin to tag or commit: `source = "git::...?ref=v1.0.0"`

### Registry Module Without Version
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `module\s+"[^"]+"\s*\{[^}]*source\s*=\s*"[^/]+/[^/]+/[^"]+"\s*(?!version)`
- Registry modules should specify a version constraint
- Remediation: Add `version = "~> 1.0"` to module block

---

## State Management

### Local Backend (Not Remote)
**Type**: regex
**Severity**: high
**Category**: best-practice
**Pattern**: `terraform\s*\{[^}]*(?<!backend\s*")[^}]*\}`
- Production Terraform should use remote state backend
- Example: Missing `backend "s3"` or `backend "gcs"` block
- Remediation: Configure remote backend for state locking and team collaboration

---

## Variable Best Practices

### Variable Without Description
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `variable\s+"[^"]+"\s*\{(?:(?!description).)*\}`
- Variables should have descriptions for documentation
- Remediation: Add `description = "Purpose of this variable"`

### Variable Without Type
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `variable\s+"[^"]+"\s*\{(?:(?!type).)*\}`
- Variables should have explicit type constraints
- Remediation: Add `type = string` or appropriate type

---

## Output Best Practices

### Output Without Description
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `output\s+"[^"]+"\s*\{(?:(?!description).)*\}`
- Outputs should have descriptions
- Remediation: Add `description = "What this output provides"`

---

## Naming Conventions

### Non-Lowercase Resource Name
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `resource\s+"[^"]+"\s+"[^"]*[A-Z][^"]*"`
- Resource names should be lowercase with underscores
- Example: `resource "aws_instance" "MyServer"` (bad)
- Remediation: Use `resource "aws_instance" "my_server"` (good)
