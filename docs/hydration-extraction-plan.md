# Hydration Data Extraction Plan

## Current State

The hydration process currently extracts the following data:

| Phase | Output File | Data Extracted |
|-------|-------------|----------------|
| dependencies | dependencies.json | SBOM (package list, versions, ecosystems) |
| technology | technology.json | Detected technologies, frameworks, tools |
| vulnerabilities | vulnerabilities.json | CVEs, severity, affected packages |
| licenses | licenses.json | License types, compliance status |
| ownership | ownership.json | Code owners, contributor stats |
| dora | dora.json | DORA metrics (deployment frequency, lead time) |
| package-health | package-health.json | Abandonment risk, typosquatting |
| provenance | provenance.json | Build attestations, SLSA levels |
| security-findings | security-findings.json | SAST findings, secrets detection |

---

## Proposed New Extractors

### Category 1: Developer Productivity Insights (DPI)

#### 1.1 Code Complexity Metrics
**Output**: `complexity.json`
**Purpose**: Understand code maintainability and technical debt hotspots

**Data to Extract**:
- Cyclomatic complexity per file/function
- Cognitive complexity (SonarQube-style)
- Lines of code (LOC, SLOC, comment ratio)
- Function/method length distribution
- Nesting depth
- File size distribution
- Largest files/functions (hotspots)

**Implementation**:
- Use `lizard` (Python) for multi-language complexity analysis
- Use `cloc` for LOC statistics
- Parse AST for language-specific metrics

**Example Output**:
```json
{
  "summary": {
    "total_files": 245,
    "total_loc": 45000,
    "avg_complexity": 4.2,
    "high_complexity_functions": 12
  },
  "hotspots": [
    {"file": "src/api/handler.ts", "complexity": 45, "loc": 890},
    {"file": "src/utils/parser.ts", "complexity": 32, "loc": 450}
  ],
  "by_language": {
    "TypeScript": {"files": 180, "loc": 35000, "avg_complexity": 4.5},
    "Python": {"files": 65, "loc": 10000, "avg_complexity": 3.8}
  }
}
```

---

#### 1.2 Test Coverage Analysis
**Output**: `test-coverage.json`
**Purpose**: Understand testing practices and coverage gaps

**Data to Extract**:
- Test file locations and count
- Test-to-code ratio
- Test frameworks detected (Jest, Pytest, etc.)
- Presence of different test types (unit, integration, e2e)
- Uncovered directories/modules
- Mock/stub usage patterns

**Implementation**:
- Scan for test file patterns (`*.test.ts`, `*_test.py`, `*_spec.rb`)
- Parse test configuration files (jest.config.js, pytest.ini)
- Detect test frameworks from imports
- Identify test utilities (mocking libraries)

**Example Output**:
```json
{
  "summary": {
    "test_files": 89,
    "source_files": 245,
    "test_to_code_ratio": 0.36,
    "test_frameworks": ["jest", "testing-library"]
  },
  "test_types": {
    "unit": 67,
    "integration": 15,
    "e2e": 7
  },
  "uncovered_directories": [
    "src/legacy/",
    "src/utils/deprecated/"
  ],
  "mocking_libraries": ["msw", "jest-mock"]
}
```

---

#### 1.3 Documentation Analysis
**Output**: `documentation.json`
**Purpose**: Assess documentation quality and completeness

**Data to Extract**:
- README presence and quality indicators
- API documentation (OpenAPI specs, JSDoc, docstrings)
- Code comment density
- Documentation file inventory (ADRs, guides, runbooks)
- Changelog presence
- Contributing guidelines

**Implementation**:
- Check for standard documentation files
- Parse OpenAPI/Swagger specs
- Count JSDoc/docstring coverage
- Analyze comment-to-code ratio

**Example Output**:
```json
{
  "summary": {
    "documentation_score": 72,
    "readme_quality": "good",
    "api_docs_present": true
  },
  "files": {
    "readme": "README.md",
    "changelog": "CHANGELOG.md",
    "contributing": "CONTRIBUTING.md",
    "api_spec": "openapi.yaml"
  },
  "code_documentation": {
    "jsdoc_coverage": 0.45,
    "comment_ratio": 0.12
  },
  "adrs": ["docs/adr/001-database-choice.md", "docs/adr/002-auth-strategy.md"]
}
```

---

#### 1.4 Dependency Graph & Architecture
**Output**: `architecture.json`
**Purpose**: Understand codebase structure and dependencies

**Data to Extract**:
- Internal module dependencies
- Circular dependency detection
- Layer violations (e.g., UI importing from data layer directly)
- Import graph statistics
- Entry points
- Dead code detection (unreferenced exports)

**Implementation**:
- Use `madge` for JavaScript/TypeScript
- Use `pydeps` for Python
- Parse import statements across languages
- Build dependency graph

**Example Output**:
```json
{
  "summary": {
    "modules": 45,
    "circular_dependencies": 3,
    "orphan_files": 8
  },
  "circular_dependencies": [
    ["src/api/auth.ts", "src/utils/user.ts", "src/api/auth.ts"]
  ],
  "layers": {
    "ui": ["src/components/", "src/pages/"],
    "api": ["src/api/"],
    "data": ["src/models/", "src/repositories/"]
  },
  "layer_violations": [
    {"from": "src/components/UserList.tsx", "to": "src/repositories/user.ts"}
  ],
  "entry_points": ["src/index.ts", "src/server.ts"]
}
```

---

#### 1.5 Git History Analysis
**Output**: `git-insights.json`
**Purpose**: Understand development patterns and team dynamics

**Data to Extract**:
- Commit frequency patterns (by day/hour)
- Active contributors (last 30/90/365 days)
- File churn (frequently modified files)
- Code age distribution
- Branch patterns
- Merge vs rebase usage
- PR/commit message quality
- Long-lived branches

**Implementation**:
- Parse git log output
- Analyze commit patterns
- Calculate file modification frequency
- Use `git-fame` or custom scripts

**Example Output**:
```json
{
  "summary": {
    "total_commits": 2450,
    "active_contributors_90d": 8,
    "avg_commits_per_week": 45
  },
  "contributors": [
    {"name": "Alice", "commits_90d": 156, "files_touched": 89},
    {"name": "Bob", "commits_90d": 98, "files_touched": 45}
  ],
  "high_churn_files": [
    {"file": "src/api/routes.ts", "changes_90d": 45},
    {"file": "src/config/settings.ts", "changes_90d": 32}
  ],
  "code_age": {
    "0-30d": 0.15,
    "31-90d": 0.25,
    "91-365d": 0.35,
    "365d+": 0.25
  },
  "patterns": {
    "most_active_day": "Tuesday",
    "most_active_hour": 14,
    "avg_pr_size": 125
  }
}
```

---

### Category 2: Security Analysis

#### 2.1 Secrets & Credentials Detection
**Output**: `secrets-scan.json`
**Purpose**: Identify exposed secrets and credential patterns

**Data to Extract**:
- Detected secrets (API keys, tokens, passwords)
- Secret patterns in different file types
- .env file analysis
- Hardcoded credentials in code
- Secret management tool detection
- Git history secret exposure

**Implementation**:
- Use `trufflehog` or `gitleaks` for scanning
- Parse .env files (without values)
- Detect secret management patterns (Vault, AWS Secrets Manager)

**Example Output**:
```json
{
  "summary": {
    "secrets_found": 3,
    "severity": "high",
    "env_files": 2
  },
  "findings": [
    {"type": "aws_access_key", "file": "config/legacy.py", "line": 45, "committed": true},
    {"type": "api_key", "file": ".env.example", "line": 12, "committed": false}
  ],
  "secret_management": {
    "detected": ["aws-secrets-manager", "dotenv"],
    "env_vars_referenced": ["DATABASE_URL", "API_KEY", "JWT_SECRET"]
  },
  "recommendations": [
    "Rotate AWS access key found in config/legacy.py",
    "Add .env to .gitignore"
  ]
}
```

---

#### 2.2 Infrastructure Security Analysis
**Output**: `infrastructure-security.json`
**Purpose**: Analyze IaC and configuration security

**Data to Extract**:
- Terraform/CloudFormation misconfigurations
- Docker security issues (running as root, exposed ports)
- Kubernetes security (RBAC, network policies, pod security)
- CI/CD pipeline security (secret handling, permissions)
- Cloud resource exposure (public S3, open security groups)

**Implementation**:
- Use `checkov` or `tfsec` for IaC scanning
- Use `hadolint` for Dockerfile linting
- Use `kubesec` for Kubernetes manifests
- Parse CI/CD configs for security patterns

**Example Output**:
```json
{
  "summary": {
    "findings": 12,
    "critical": 2,
    "high": 4,
    "medium": 6
  },
  "terraform": {
    "findings": [
      {"resource": "aws_s3_bucket.data", "issue": "Public access enabled", "severity": "critical"},
      {"resource": "aws_security_group.web", "issue": "0.0.0.0/0 ingress on port 22", "severity": "high"}
    ]
  },
  "docker": {
    "findings": [
      {"file": "Dockerfile", "issue": "Running as root", "severity": "medium"},
      {"file": "Dockerfile", "issue": "Using latest tag", "severity": "low"}
    ]
  },
  "kubernetes": {
    "findings": [
      {"file": "k8s/deployment.yaml", "issue": "No resource limits", "severity": "medium"}
    ]
  }
}
```

---

#### 2.3 Authentication & Authorization Patterns
**Output**: `auth-analysis.json`
**Purpose**: Understand authentication implementation

**Data to Extract**:
- Auth providers detected (Auth0, Cognito, custom)
- Session management patterns
- JWT usage and configuration
- OAuth/OIDC implementation
- RBAC/ABAC patterns
- Password handling (hashing algorithms)
- MFA presence

**Implementation**:
- Detect auth libraries in dependencies
- Parse auth configuration files
- Scan for JWT patterns
- Identify session middleware

**Example Output**:
```json
{
  "summary": {
    "auth_providers": ["auth0"],
    "mfa_detected": true,
    "session_type": "jwt"
  },
  "jwt": {
    "library": "jsonwebtoken",
    "algorithm_detected": "RS256",
    "expiry_pattern": "1h access, 7d refresh"
  },
  "oauth": {
    "providers": ["google", "github"],
    "pkce_enabled": true
  },
  "password": {
    "hashing": "bcrypt",
    "min_rounds": 10
  },
  "concerns": [
    "JWT secret appears hardcoded in tests"
  ]
}
```

---

#### 2.4 API Security Analysis
**Output**: `api-security.json`
**Purpose**: Analyze API security posture

**Data to Extract**:
- Endpoint inventory
- Authentication requirements per endpoint
- Rate limiting presence
- Input validation patterns
- CORS configuration
- API versioning
- Deprecated endpoints
- GraphQL security (if applicable)

**Implementation**:
- Parse OpenAPI specs
- Scan route definitions
- Detect middleware patterns
- Analyze GraphQL schemas

**Example Output**:
```json
{
  "summary": {
    "total_endpoints": 45,
    "authenticated": 42,
    "public": 3,
    "rate_limited": 38
  },
  "endpoints": [
    {"path": "/api/users", "method": "GET", "auth": "jwt", "rate_limit": "100/min"},
    {"path": "/api/public/health", "method": "GET", "auth": "none", "rate_limit": "none"}
  ],
  "cors": {
    "origins": ["https://app.example.com"],
    "credentials": true
  },
  "concerns": [
    "3 endpoints missing rate limiting",
    "DELETE /api/users/{id} allows any authenticated user"
  ]
}
```

---

#### 2.5 Data Flow & Privacy Analysis
**Output**: `data-privacy.json`
**Purpose**: Understand data handling and privacy compliance

**Data to Extract**:
- PII field detection in schemas
- Data storage locations
- Encryption at rest patterns
- Data transmission security
- Logging of sensitive data
- Data retention patterns
- GDPR/CCPA compliance indicators

**Implementation**:
- Parse database schemas for PII patterns
- Detect encryption libraries
- Scan logging statements
- Identify data models

**Example Output**:
```json
{
  "summary": {
    "pii_fields_detected": 12,
    "encryption_at_rest": true,
    "sensitive_logging_risk": 2
  },
  "pii_locations": [
    {"model": "User", "fields": ["email", "phone", "ssn"]},
    {"model": "Payment", "fields": ["card_number", "cvv"]}
  ],
  "encryption": {
    "at_rest": ["AES-256"],
    "in_transit": ["TLS 1.3"]
  },
  "logging_concerns": [
    {"file": "src/api/auth.ts", "line": 45, "issue": "Logging user password field"}
  ],
  "compliance_indicators": {
    "consent_management": true,
    "data_deletion_endpoint": true,
    "data_export_endpoint": false
  }
}
```

---

### Category 3: Repository Health & Maintenance

#### 3.1 Technical Debt Indicators
**Output**: `tech-debt.json`
**Purpose**: Quantify and locate technical debt

**Data to Extract**:
- TODO/FIXME/HACK comments
- Deprecated API usage
- Outdated dependencies (major versions behind)
- Dead code indicators
- Duplicated code
- Long parameter lists
- God classes/files

**Implementation**:
- Grep for debt markers (TODO, FIXME, etc.)
- Compare dependency versions
- Use `jscpd` for duplication detection
- Parse AST for code smells

**Example Output**:
```json
{
  "summary": {
    "debt_score": 65,
    "todo_count": 45,
    "fixme_count": 12,
    "hack_count": 3,
    "outdated_major": 8
  },
  "markers": [
    {"type": "TODO", "file": "src/api/auth.ts", "line": 23, "text": "TODO: Add rate limiting"},
    {"type": "FIXME", "file": "src/utils/date.ts", "line": 89, "text": "FIXME: Timezone bug"}
  ],
  "outdated_dependencies": [
    {"name": "react", "current": "16.8.0", "latest": "18.2.0", "versions_behind": 2},
    {"name": "webpack", "current": "4.46.0", "latest": "5.88.0", "versions_behind": 1}
  ],
  "duplication": {
    "percentage": 4.5,
    "blocks": 23,
    "largest_block": {"files": ["src/api/users.ts", "src/api/teams.ts"], "lines": 45}
  }
}
```

---

#### 3.2 Build & CI Analysis
**Output**: `build-analysis.json`
**Purpose**: Understand build system and CI health

**Data to Extract**:
- Build tool detection (webpack, vite, gradle, etc.)
- Build script complexity
- CI/CD pipeline steps
- Average build time indicators
- Caching effectiveness
- Parallelization opportunities
- Flaky test patterns

**Implementation**:
- Parse build configuration files
- Analyze CI workflow files
- Detect build optimizations

**Example Output**:
```json
{
  "summary": {
    "build_tools": ["vite", "typescript"],
    "ci_platform": "github-actions",
    "estimated_build_complexity": "medium"
  },
  "build_config": {
    "entry_points": 3,
    "output_formats": ["esm", "cjs"],
    "code_splitting": true,
    "tree_shaking": true
  },
  "ci_pipeline": {
    "stages": ["lint", "test", "build", "deploy"],
    "parallelization": true,
    "caching": ["node_modules", ".next/cache"],
    "estimated_duration": "8-12 min"
  },
  "optimization_opportunities": [
    "Consider parallel test execution",
    "Add dependency caching for faster installs"
  ]
}
```

---

## Implementation Architecture

### Extractor Interface

Each extractor should implement a standard interface:

```bash
#!/bin/bash
# Example extractor: complexity-analyser-data.sh

# Standard interface
# Input: --local-path <path> OR --repo <owner/repo>
# Output: JSON to stdout
# Exit codes: 0 = success, 1 = error, 2 = skipped (not applicable)

extract_complexity() {
    local repo_path="$1"

    # Run analysis tools
    lizard "$repo_path" --json > /tmp/lizard.json
    cloc "$repo_path" --json > /tmp/cloc.json

    # Combine and format output
    jq -s '{
        complexity: .[0],
        loc: .[1]
    }' /tmp/lizard.json /tmp/cloc.json
}
```

### Output Directory Structure

```
~/.gibson/projects/{project-id}/analysis/
├── manifest.json           # Metadata about all analyses
├── dependencies.json       # SBOM
├── vulnerabilities.json    # CVE data
├── technology.json         # Detected tech stack
├── licenses.json          # License compliance
├── complexity.json        # Code metrics        [NEW]
├── test-coverage.json     # Testing analysis    [NEW]
├── documentation.json     # Docs analysis       [NEW]
├── architecture.json      # Dependency graph    [NEW]
├── git-insights.json      # Git history         [NEW]
├── secrets-scan.json      # Secrets detection   [NEW]
├── infra-security.json    # IaC security        [NEW]
├── auth-analysis.json     # Auth patterns       [NEW]
├── api-security.json      # API analysis        [NEW]
├── data-privacy.json      # Privacy analysis    [NEW]
├── tech-debt.json         # Debt indicators     [NEW]
└── build-analysis.json    # Build system        [NEW]
```

---

## Priority Implementation Order

### Phase 1: High Value, Low Complexity (Week 1-2)
1. **tech-debt.json** - Simple grep + version comparison
2. **documentation.json** - File existence + basic parsing
3. **git-insights.json** - Git log parsing (already have dora)
4. **test-coverage.json** - File pattern matching

### Phase 2: Security Essentials (Week 3-4)
5. **secrets-scan.json** - Integrate gitleaks/trufflehog
6. **infra-security.json** - Integrate checkov
7. **auth-analysis.json** - Pattern detection

### Phase 3: Advanced Analysis (Week 5-6)
8. **complexity.json** - Integrate lizard + cloc
9. **architecture.json** - Integrate madge
10. **api-security.json** - OpenAPI parsing + route analysis

### Phase 4: Privacy & Compliance (Week 7-8)
11. **data-privacy.json** - Schema analysis + PII detection
12. **build-analysis.json** - Build config parsing

---

## Tool Dependencies

| Extractor | Tools Required | Installation |
|-----------|---------------|--------------|
| complexity | lizard, cloc | `pip install lizard`, `brew install cloc` |
| architecture | madge | `npm install -g madge` |
| secrets-scan | gitleaks | `brew install gitleaks` |
| infra-security | checkov | `pip install checkov` |
| test-coverage | (built-in) | - |
| git-insights | git | (built-in) |
| tech-debt | jscpd | `npm install -g jscpd` |

---

## Integration with Claude Analysis

Each JSON output should be structured to facilitate Claude's analysis:

1. **Summary section** - High-level metrics for quick assessment
2. **Findings section** - Detailed issues with file/line references
3. **Recommendations section** - Actionable suggestions
4. **Metadata section** - Tool versions, timestamps, confidence levels

This allows Claude to:
- Quickly assess repository health from summaries
- Drill into specific issues when needed
- Provide actionable recommendations
- Cross-reference findings across extractors
