# Terraform Versions

## Terraform CLI

### Current Versions (2025)

#### Terraform v1.x (Current Stable)
- **Latest**: 1.10.x (as of 2025)
- **Status**: Active development
- **Release Cycle**: ~3-4 months between minor versions
- **Repository**: https://github.com/hashicorp/terraform

#### Recent Version History
- **1.10.x** (2025) - Current
- **1.9.x** (2024) - Stable
- **1.8.x** (2024) - Stable
- **1.7.x** (2024) - Stable
- **1.6.x** (2023) - Stable (LTS candidate)
- **1.5.x** (2023) - Stable (import blocks, checks)
- **1.4.x** (2023) - Stable
- **1.3.x** (2022) - Stable (optional attributes)
- **1.2.x** (2022) - Stable
- **1.1.x** (2021) - Stable
- **1.0.x** (2021) - Stable (first v1 release)

### Terraform v0.x (Legacy)
- **0.15.x** - Last 0.x series
- **0.14.x** - Lock file introduced
- **0.13.x** - Required providers
- **0.12.x** - HCL2, for expressions
- **0.11.x and earlier** - EOL

### Version Detection

#### CLI Version Check
```bash
terraform version
# Output: Terraform v1.10.0
# on linux_amd64

terraform --version
terraform -v
```

#### Version in Configuration
```hcl
terraform {
  required_version = ">= 1.5.0"
}
```

#### Version Constraints
```hcl
# Exact version
required_version = "= 1.5.0"

# Minimum version
required_version = ">= 1.5.0"

# Pessimistic constraint
required_version = "~> 1.5.0"  # 1.5.x, but not 1.6.0

# Range
required_version = ">= 1.5.0, < 2.0.0"

# Multiple constraints
required_version = ">= 1.5.0, != 1.5.1, < 2.0.0"
```

## Provider Versions

### AWS Provider
```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
```

#### Version History
- **5.x** (2023-present) - Current
- **4.x** (2022-2023) - Previous stable
- **3.x** (2020-2022) - EOL
- **2.x** (2018-2020) - EOL
- **1.x** (2017-2018) - EOL

### Azure Provider
```hcl
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}
```

#### Version History
- **3.x** (2022-present) - Current
- **2.x** (2020-2022) - Previous stable
- **1.x** (2018-2020) - EOL

### Google Cloud Provider
```hcl
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}
```

#### Version History
- **5.x** (2023-present) - Current
- **4.x** (2022-2023) - Previous stable
- **3.x** (2020-2022) - EOL
- **2.x** (2019-2020) - EOL

### Kubernetes Provider
```hcl
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}
```

#### Version History
- **2.x** (2021-present) - Current
- **1.x** (2019-2021) - EOL

## Version Constraint Operators

### Exact Version
```hcl
version = "= 1.5.0"
# Only 1.5.0 is allowed
```

### Pessimistic Constraint (~>)
```hcl
version = "~> 1.5.0"
# Allows 1.5.x (1.5.0, 1.5.1, etc.)
# Does NOT allow 1.6.0

version = "~> 1.5"
# Allows 1.x where x >= 5 (1.5.0, 1.6.0, etc.)
# Does NOT allow 2.0.0
```

### Comparison Operators
```hcl
version = ">= 1.5.0"   # 1.5.0 or higher
version = "> 1.5.0"    # Higher than 1.5.0
version = "<= 1.5.0"   # 1.5.0 or lower
version = "< 1.5.0"    # Lower than 1.5.0
version = "!= 1.5.0"   # Any version except 1.5.0
```

### Combined Constraints
```hcl
version = ">= 1.5.0, < 2.0.0"
# Between 1.5.0 (inclusive) and 2.0.0 (exclusive)

version = "~> 1.5.0, != 1.5.3"
# 1.5.x except 1.5.3
```

## Dependency Lock File

### .terraform.lock.hcl
```hcl
# This file is maintained automatically by "terraform init".
provider "registry.terraform.io/hashicorp/aws" {
  version     = "5.31.0"
  constraints = "~> 5.0"
  hashes = [
    "h1:abc123...",
    "zh:def456...",
  ]
}

provider "registry.terraform.io/hashicorp/random" {
  version     = "3.6.0"
  constraints = "~> 3.0"
  hashes = [
    "h1:xyz789...",
  ]
}
```

### Lock File Commands
```bash
# Initialize and create/update lock file
terraform init

# Upgrade providers within constraints
terraform init -upgrade

# Update specific provider
terraform init -upgrade=hashicorp/aws
```

## Terraform State Version

### State File Format
```json
{
  "version": 4,
  "terraform_version": "1.5.0",
  "serial": 1,
  "lineage": "abc-def-123",
  "outputs": {},
  "resources": []
}
```

### State Version History
- **Version 4** - Terraform 0.12+
- **Version 3** - Terraform 0.11
- **Version 2** - Terraform 0.10
- **Version 1** - Terraform 0.9 and earlier

### State Compatibility
- Terraform can read state files from older versions
- State is automatically upgraded on write
- Cannot downgrade Terraform version after state upgrade

## Version Management Tools

### tfenv (Terraform version manager)
```bash
# Install tfenv
brew install tfenv

# List available versions
tfenv list-remote

# Install specific version
tfenv install 1.5.0

# Use specific version
tfenv use 1.5.0

# Use version from .terraform-version file
tfenv install
tfenv use
```

### .terraform-version file
```
1.5.0
```

### asdf (Universal version manager)
```bash
# Install asdf
brew install asdf

# Add terraform plugin
asdf plugin add terraform

# Install specific version
asdf install terraform 1.5.0

# Set global version
asdf global terraform 1.5.0

# Set local version (per directory)
asdf local terraform 1.5.0
```

### .tool-versions file
```
terraform 1.5.0
```

## CI/CD Version Specifications

### GitHub Actions
```yaml
- name: Setup Terraform
  uses: hashicorp/setup-terraform@v3
  with:
    terraform_version: 1.5.0
    # or
    terraform_version: ~1.5
```

### GitLab CI
```yaml
image:
  name: hashicorp/terraform:1.5.0
  entrypoint: [""]
```

### Docker
```dockerfile
FROM hashicorp/terraform:1.5.0
```

### CircleCI
```yaml
docker:
  - image: hashicorp/terraform:1.5.0
```

## Version Detection Patterns

### From terraform.tf or versions.tf
```hcl
terraform {
  required_version = ">= 1.5.0"
}
```

### From .terraform.lock.hcl
```hcl
provider "registry.terraform.io/hashicorp/aws" {
  version     = "5.31.0"
}
```

### From .terraform-version
```
1.5.0
```

### From CI/CD configuration
```yaml
terraform_version: 1.5.0
```

### From Dockerfile
```dockerfile
FROM hashicorp/terraform:1.5.0
```

## Breaking Changes Between Major Versions

### 0.11 → 0.12
- HCL syntax overhaul
- First-class expressions
- Type system improvements
- For expressions
- Dynamic blocks

### 0.12 → 0.13
- Required providers block
- Provider source specification
- depends_on for modules
- count and for_each in modules

### 0.13 → 0.14
- Lock file introduction
- Sensitive values in outputs
- Provider requirements in modules

### 0.14 → 0.15
- Sensitive values propagation
- Module expansion improvements
- Provider configuration improvements

### 0.15 → 1.0
- No major breaking changes
- Stability commitment

### 1.0 → 1.5
- Import blocks (1.5)
- Check blocks (1.5)
- Terraform test framework (1.6)
- Removed blocks (1.7)
- Provider functions (1.8)

## Detection Confidence

- **HIGH**: terraform version command output
- **HIGH**: required_version in terraform block
- **HIGH**: .terraform.lock.hcl version declarations
- **HIGH**: .terraform-version file
- **MEDIUM**: Provider version constraints
- **MEDIUM**: CI/CD configuration terraform version
- **LOW**: Inferred from feature usage (language features)
