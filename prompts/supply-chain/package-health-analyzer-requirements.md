<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Package Health Analyzer - Requirements and Implementation Prompt

## Overview

Build a comprehensive package health analysis system that evaluates the operational health, maintenance status, and best practices compliance of software dependencies. This complements vulnerability scanning by focusing on long-term package health, maintenance risks, and dependency management best practices.

**Status**: 🚧 Experimental - New capability for supply chain utilities

## Objectives

### Primary Goals
1. **Operational Health Assessment**: Identify deprecated, unmaintained, or risky packages
2. **deps.dev Integration**: Leverage OpenSSF Scorecard, package metadata, and health signals
3. **Best Practices Enforcement**: Check for version pinning, standardization, and update policies
4. **Maintenance Risk Detection**: Flag packages with irregular updates, low activity, or known issues
5. **AI-Enhanced Recommendations**: Provide actionable guidance for package management improvements

### Key Capabilities
- Package deprecation detection
- OpenSSF Scorecard integration
- Package health scoring
- Dependency freshness analysis
- Version pinning validation
- Library standardization checks
- Update risk assessment
- Alternative package suggestions

## Architecture

### Module Structure
```
utils/supply-chain/package-health-analysis/
├── package-health-analyzer.sh              # Base analyzer
├── package-health-analyzer-claude.sh       # AI-enhanced analyzer
└── compare-analyzers.sh                    # Comparison tool
```

### Integration Points
- Integrates with existing supply-chain-scanner.sh via `--health` flag
- Uses same configuration system (hierarchical config)
- Follows same multi-repo/org scanning patterns
- Leverages existing SBOM generation infrastructure

## Data Sources

### deps.dev API (v3alpha)
**Primary health data source**

**Key Endpoints:**
```
GET /v3alpha/systems/{system}/packages/{package}
GET /v3alpha/systems/{system}/packages/{package}/versions/{version}
GET /v3alpha/advisories/{advisoryId}
GET /v3alpha/projects/{projectKey}
```

**Health Signals:**
- OpenSSF Scorecard metrics (10+ security/quality checks)
- Package metadata (description, licenses, links)
- Version history and release cadence
- Dependency relationships
- Security advisories
- Deprecation status
- Project health indicators

**Scorecard Metrics** (0-10 scale):
- Binary-Artifacts: No binary artifacts in source
- Branch-Protection: Branch protection rules
- CI-Tests: Continuous integration testing
- CII-Best-Practices: OpenSSF badge level
- Code-Review: Code review requirements
- Contributors: Multiple contributors
- Dangerous-Workflow: No dangerous workflow patterns
- Dependency-Update-Tool: Automated dependency updates
- Fuzzing: Fuzz testing integration
- License: Valid license file
- Maintained: Active maintenance
- Pinned-Dependencies: Dependencies pinned
- SAST: Static analysis security testing
- Security-Policy: Security policy file
- Signed-Releases: Release signing
- Token-Permissions: Minimal token permissions
- Vulnerabilities: No known vulnerabilities

### npm Registry API
**For npm ecosystem specifics**
- Package deprecation warnings
- Latest versions and tags
- Download statistics
- Maintainer information

### GitHub API (via gh CLI)
**Repository health signals**
- Last commit date
- Issue/PR activity
- Contributor activity
- Archived status
- Fork status

## Analysis Capabilities

### 1. Deprecation Detection

**Checks:**
- Package marked as deprecated in registry
- Maintainer deprecation notices
- README deprecation warnings
- Package archived/abandoned
- Successor packages identified

**Output:**
```
Package: lodash@4.17.20
Status: DEPRECATED
Deprecation Notice: "This package is no longer maintained. Use lodash-es instead."
Successor: lodash-es@4.17.21
Migration Guide: https://github.com/lodash/lodash/wiki/Migrating
Risk Level: HIGH
```

### 2. OpenSSF Scorecard Analysis

**Checks:**
- Overall scorecard score
- Individual metric scores
- Critical failures (score < 3)
- Missing security practices

**Scoring:**
- 9-10: Excellent
- 7-8: Good
- 5-6: Fair
- 3-4: Poor
- 0-2: Critical

**Output:**
```
Package: express@4.18.2
OpenSSF Scorecard: 7.2/10 (Good)

Metrics:
  ✓ Code-Review: 10/10 (All changes reviewed)
  ✓ CI-Tests: 10/10 (Tests run on all commits)
  ✓ Maintained: 9/10 (Active maintenance)
  ⚠ Branch-Protection: 5/10 (Weak branch protection)
  ⚠ Pinned-Dependencies: 4/10 (Dependencies not pinned)
  ✗ Fuzzing: 0/10 (No fuzzing)

Recommendations:
  - Enable branch protection rules
  - Pin all dependencies to specific versions
  - Integrate fuzz testing
```

### 3. Package Health Scoring

**Custom Health Score** (0-100):
```
Health Score =
  (OpenSSF Score * 0.30) +
  (Maintenance Score * 0.25) +
  (Security Score * 0.25) +
  (Freshness Score * 0.10) +
  (Popularity Score * 0.10)
```

**Maintenance Score:**
- Recent commits (last 90 days)
- Release frequency
- Issue response time
- Active maintainers

**Security Score:**
- No critical vulnerabilities
- Security policy present
- Signed releases
- Vulnerability response time

**Freshness Score:**
- Dependency updates
- Version currency
- Time since last release

**Popularity Score:**
- Download counts
- GitHub stars
- Dependent packages

### 4. Version Pinning Analysis

**Checks:**
```json
{
  "package.json": {
    "express": "^4.18.0",  // ⚠ Caret range
    "lodash": "~4.17.0",   // ⚠ Tilde range
    "axios": "1.5.0"       // ✓ Pinned
  }
}
```

**Validation:**
- Detect unpinned dependencies (^, ~, *, >, <, >=, <=)
- Check for wildcard versions
- Verify lock file exists (package-lock.json, yarn.lock)
- Validate lock file freshness

**Output:**
```
Version Pinning Analysis:
  Total Dependencies: 150
  Pinned: 45 (30%)
  Caret Range (^): 80 (53%)
  Tilde Range (~): 20 (13%)
  Wildcards (*): 5 (3%)

Risk Assessment: MEDIUM
- 105 dependencies allow automatic updates
- Lock file exists: ✓
- Lock file age: 15 days (Fresh)

Recommendation: Pin critical dependencies to exact versions
```

### 5. Library Standardization

**Checks:**
- Multiple versions of same library
- Conflicting dependencies
- Duplicate functionality (e.g., multiple HTTP clients)
- Deprecated libraries still in use

**Example:**
```
Standardization Issues Found:

Multiple Versions:
  lodash:
    - 4.17.20 (used by 15 packages)
    - 4.17.21 (used by 8 packages)
    - 3.10.1 (used by 2 packages)  ⚠ OUTDATED

  Recommendation: Standardize on lodash@4.17.21

Duplicate Functionality:
  HTTP Clients:
    - axios (10 occurrences)
    - node-fetch (5 occurrences)
    - request (3 occurrences)  ⚠ DEPRECATED

  Recommendation: Standardize on axios, remove request

Conflicting Dependencies:
  react:
    - Direct: 18.2.0
    - Peer (via react-router): ^18.0.0
    - Transitive (via old-lib): 17.0.2  ⚠ CONFLICT

  Risk: Runtime errors possible
  Action: Update old-lib or find alternative
```

### 6. Maintenance Risk Assessment

**Risk Factors:**
- Last commit > 365 days ago
- No releases in last 180 days
- Unresolved critical issues
- Few active contributors
- Single maintainer (bus factor = 1)

**Risk Levels:**
- **CRITICAL**: Abandoned (no activity 2+ years)
- **HIGH**: Inactive (no activity 1+ year)
- **MEDIUM**: Slow maintenance (no activity 6+ months)
- **LOW**: Active maintenance

**Output:**
```
Maintenance Risk Analysis:

HIGH RISK (5 packages):
  ├─ old-parser@2.1.0
  │  Last commit: 18 months ago
  │  Last release: 24 months ago
  │  Open critical issues: 12
  │  Maintainers: 1 (inactive)
  │  Alternative: new-parser@3.0.0

  └─ deprecated-util@1.5.0
     Status: DEPRECATED
     Successor: modern-util@2.0.0
     Migration: Breaking changes, ~4 hours

MEDIUM RISK (12 packages):
  └─ slow-lib@4.2.0
     Last release: 8 months ago
     Issues unaddressed: 45
     Recommendation: Monitor, consider alternatives

Summary:
  Critical: 0
  High: 5
  Medium: 12
  Low: 133

Action: Replace 5 high-risk dependencies
```

## Implementation Requirements

### 1. Repository Cloning and SBOM Generation

**Follow existing patterns:**
```bash
# Clone repository if needed
clone_repository() {
    local repo="$1"
    local target_dir="$2"

    if [[ ! -d "$target_dir" ]]; then
        gh repo clone "$repo" "$target_dir"
    fi
}

# Generate SBOM if missing
generate_sbom() {
    local repo_dir="$1"
    local sbom_path="$repo_dir/bom.json"

    if [[ ! -f "$sbom_path" ]]; then
        syft "$repo_dir" -o cyclonedx-json > "$sbom_path"
    fi

    echo "$sbom_path"
}
```

### 2. deps.dev API Integration

**API Client:**
```bash
# Query package information
query_depsdev_package() {
    local system="$1"  # npm, pypi, go, maven, etc.
    local package="$2"
    local version="$3"

    local url="https://api.deps.dev/v3alpha/systems/${system}/packages/${package}/versions/${version}"

    curl -s "$url" | jq .
}

# Get OpenSSF Scorecard
get_scorecard() {
    local system="$1"
    local package="$2"
    local version="$3"

    local data=$(query_depsdev_package "$system" "$package" "$version")
    echo "$data" | jq '.projects[0].scorecard'
}
```

### 3. Health Score Calculation

**Algorithm:**
```bash
calculate_health_score() {
    local package="$1"
    local version="$2"

    # Get data
    local scorecard=$(get_scorecard "npm" "$package" "$version")
    local maintenance=$(calculate_maintenance_score "$package")
    local security=$(calculate_security_score "$package")
    local freshness=$(calculate_freshness_score "$package" "$version")
    local popularity=$(get_popularity_score "$package")

    # Weighted calculation
    local health=$((
        (scorecard * 30 / 10) +
        (maintenance * 25 / 100) +
        (security * 25 / 100) +
        (freshness * 10 / 100) +
        (popularity * 10 / 100)
    ))

    echo "$health"
}
```

### 4. Configuration Integration

**Add to supply-chain config:**
```json
{
  "modules": {
    "supply_chain": {
      "package_health": {
        "enabled": true,
        "min_health_score": 50,
        "check_deprecation": true,
        "check_pinning": true,
        "check_standardization": true,
        "risk_levels": ["high", "critical"],
        "openssf": {
          "min_overall_score": 5,
          "required_checks": [
            "Code-Review",
            "CI-Tests",
            "License"
          ]
        }
      }
    }
  }
}
```

### 5. Output Formats

**Table Format** (default):
```
╭──────────────────┬─────────┬────────────┬─────────────────╮
│ Package          │ Health  │ OpenSSF    │ Risk Level      │
├──────────────────┼─────────┼────────────┼─────────────────┤
│ express@4.18.2   │ 85/100  │ 7.2/10     │ LOW             │
│ lodash@4.17.20   │ 45/100  │ 5.8/10     │ MEDIUM (old)    │
│ old-lib@1.0.0    │ 20/100  │ 3.2/10     │ HIGH (inactive) │
╰──────────────────┴─────────┴────────────┴─────────────────╯
```

**JSON Format**:
```json
{
  "packages": [
    {
      "name": "express",
      "version": "4.18.2",
      "health_score": 85,
      "openssf_score": 7.2,
      "risk_level": "low",
      "deprecated": false,
      "maintenance": {
        "last_commit": "2024-11-15",
        "release_frequency": "monthly",
        "status": "active"
      },
      "issues": [],
      "recommendations": []
    }
  ]
}
```

## Claude-Enhanced Analysis

### AI Capabilities

**Context-Aware Analysis:**
- Pattern recognition across multiple packages
- Risk narrative generation
- Migration planning and effort estimation
- Alternative package suggestions
- Best practices recommendations

**Example Prompt Structure:**
```markdown
Analyze this package health scan for operational risks and improvement opportunities.

Package Health Data:
[JSON data with health scores, deprecations, scorecard results]

Provide:

1. **Overall Health Assessment**
   - Current state summary
   - Key risk areas
   - Health trends

2. **Critical Issues**
   - Deprecated packages requiring immediate action
   - High-risk unmaintained dependencies
   - Security/quality concerns from OpenSSF scores

3. **Standardization Opportunities**
   - Version conflicts to resolve
   - Duplicate libraries to consolidate
   - Pinning recommendations

4. **Migration Planning**
   - Packages requiring replacement
   - Suggested alternatives with rationale
   - Estimated migration effort
   - Breaking change analysis

5. **Best Practices Recommendations**
   - Dependency management improvements
   - Update policies to implement
   - Quality gates to enforce

Focus on ACTIONABLE insights that development teams can execute.
Do NOT provide generic advice - be specific to this scan's results.
```

### AI-Enhanced Features

**Smart Alternative Suggestions:**
- Analyzes package purpose and API
- Suggests modern replacements
- Considers ecosystem compatibility
- Estimates migration complexity

**Migration Effort Estimation:**
```
Migration: request → axios

Complexity: MEDIUM (4-6 hours)
Breaking Changes:
  - Different promise API
  - Request config format changed
  - Response structure differs

Code Locations: 15 files
Typical Patterns:
  request.get() → axios.get()
  request.post() → axios.post()

Testing Required:
  - Integration tests: 25 tests
  - Unit tests: 40 tests

Risks:
  - Error handling differences
  - Timeout behavior changed
  - Proxy configuration format
```

**Health Trend Analysis:**
- Compares current vs historical health
- Identifies degrading packages
- Predicts future maintenance issues

## Command-Line Interface

### Base Analyzer

```bash
# Basic usage
./package-health-analyzer.sh owner/repo

# With options
./package-health-analyzer.sh \
  --format json \
  --min-health 50 \
  --check-pinning \
  --check-deprecation \
  owner/repo

# Multi-repo
./package-health-analyzer.sh --org myorg

# Specific checks
./package-health-analyzer.sh \
  --check-standardization \
  --check-scorecard \
  owner/repo
```

### Claude-Enhanced Analyzer

```bash
# AI-enhanced analysis
export ANTHROPIC_API_KEY="your-key"
./package-health-analyzer-claude.sh owner/repo

# With migration planning
./package-health-analyzer-claude.sh \
  --suggest-alternatives \
  --estimate-effort \
  owner/repo
```

### Central Orchestrator Integration

```bash
# Via supply-chain-scanner.sh
./supply-chain-scanner.sh --health owner/repo

# Combined analysis
./supply-chain-scanner.sh --all owner/repo
# Runs: vulnerability, provenance, AND health
```

## RAG Knowledge Base Requirements

### Create Documentation Files

**1. rag/supply-chain/package-health/deps-dev-api.md**
- Complete deps.dev API v3alpha reference
- All endpoints with parameters
- Response schemas
- Authentication and rate limits
- Example queries
- OpenSSF Scorecard integration

**2. rag/supply-chain/package-health/openssf-scorecard.md**
- All 15+ scorecard checks explained
- Scoring methodology (0-10 scale)
- Interpretation guidelines
- Best practices for each check
- Industry benchmarks

**3. rag/supply-chain/package-health/package-management-best-practices.md**
- Version pinning strategies
- Lock file management
- Dependency update policies
- Security update workflows
- Library standardization
- Deprecation handling

**4. rag/supply-chain/package-health/package-ecosystems.md**
- npm registry API
- PyPI API
- Go modules
- Maven Central
- Cargo registry
- Ecosystem-specific health signals

**5. rag/supply-chain/package-health/maintenance-risk-assessment.md**
- Risk factor definitions
- Activity metrics and thresholds
- Bus factor calculation
- Succession planning
- Alternative selection criteria

## Testing Requirements

### Test Cases

1. **Deprecation Detection**
   - Test with known deprecated packages
   - Verify successor identification
   - Check migration guidance extraction

2. **Health Scoring**
   - Test with high/medium/low health packages
   - Verify score calculation accuracy
   - Validate threshold enforcement

3. **Version Pinning**
   - Test various version specifiers
   - Verify lock file detection
   - Check conflict identification

4. **Standardization**
   - Test multi-version scenarios
   - Verify duplicate detection
   - Check consolidation suggestions

5. **Multi-Repo Scanning**
   - Test org-wide scanning
   - Verify batch processing
   - Check aggregated reporting

### Known Good/Bad Packages for Testing

**High Health:**
- express@latest
- react@latest
- lodash@latest (despite being large)

**Deprecated:**
- request (successor: axios)
- node-uuid (successor: uuid)

**Unmaintained:**
- Find packages with >2 years no activity

## Documentation Requirements

### README.md

Include:
- Clear experimental status
- Purpose and capabilities
- Quick start guide
- API integration details
- Configuration options
- Output format examples
- Limitations and roadmap

### CHANGELOG.md

Track:
- Version history
- Features added
- Known limitations
- Breaking changes
- Migration guides

## Success Criteria

### Must Have (MVP)
- [x] Deprecation detection working
- [x] deps.dev API integration
- [x] OpenSSF Scorecard display
- [x] Basic health scoring
- [x] Multi-repo support
- [x] JSON output format

### Should Have
- [x] Version pinning analysis
- [x] Library standardization checks
- [x] Claude-enhanced recommendations
- [x] Migration effort estimation
- [x] RAG documentation complete

### Nice to Have
- [ ] Historical trend tracking
- [ ] Dashboard integration
- [ ] Automated PR creation for updates
- [ ] Policy-as-code enforcement
- [ ] Slack/email notifications

## Integration with Existing System

### Config System
- Use hierarchical config (global + module)
- Follow existing config patterns
- Add package-health section to config

### Multi-Repo Scanning
- Use same patterns as vulnerability/provenance
- Support --org and --repo flags
- Handle GitHub authentication same way

### Output and Reporting
- Follow same format patterns (table/JSON/markdown)
- Use same color coding
- Integrate with central orchestrator

### Error Handling
- Consistent error messages
- Graceful API failures
- Retry logic for transient errors

## Implementation Priority

### Phase 1: Core Health Analysis
1. Create package-health-analyzer.sh base script
2. Implement deps.dev API client
3. Add OpenSSF Scorecard fetching
4. Implement basic health scoring
5. Add deprecation detection

### Phase 2: Best Practices Checks
1. Version pinning analysis
2. Library standardization
3. Maintenance risk assessment
4. Configuration integration

### Phase 3: AI Enhancement
1. Create package-health-analyzer-claude.sh
2. Implement analysis prompts
3. Add alternative suggestions
4. Migration effort estimation

### Phase 4: Documentation & RAG
1. Create RAG documentation files
2. Write README and CHANGELOG
3. Add examples and test cases
4. Integration documentation

## Notes

- Mark as experimental initially
- Gather feedback from early users
- Iterate on health scoring algorithm
- Consider community contributions to RAG
- Plan for v2.0 with policy enforcement

## References

- deps.dev API: https://docs.deps.dev/api/v3alpha/
- OpenSSF Scorecard: https://github.com/ossf/scorecard
- SLSA: https://slsa.dev/
- OpenSSF Best Practices: https://bestpractices.coreinfrastructure.org/
