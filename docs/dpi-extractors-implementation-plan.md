# DPI Data Extractors Implementation Plan

## Overview

This document outlines the implementation plan for new Developer Productivity Insights (DPI) data extractors, following the current Phantom/Gibson architecture pattern of:
1. **Data extractors** (`*-data.sh`) - Static analysis, no AI, outputs JSON
2. **Claude-enhanced analyzers** (`*-analyser.sh`) - AI-powered deep analysis

---

## Current Architecture Summary

### Storage Structure
```
~/.phantom/                          # or ~/.gibson/
├── config.json                      # Global configuration
├── index.json                       # Project index and active project
├── cache/                           # Shared cache
└── projects/
    └── {org}/{repo}/
        ├── project.json             # Project metadata
        ├── repo/                    # Cloned repository
        └── analysis/
            ├── manifest.json        # Analysis metadata
            ├── sbom.cdx.json        # CycloneDX SBOM
            ├── dependencies.json    # Dependency analysis
            ├── technology.json      # Technology detection
            ├── vulnerabilities.json # CVE findings
            ├── licenses.json        # License analysis
            ├── ownership.json       # Code ownership
            ├── dora.json            # DORA metrics
            ├── package-health.json  # Package health
            ├── provenance.json      # Build provenance
            ├── security-findings.json # Code security
            └── iac-security.json    # IaC security (NEW)
```

### Analyzer Modes
- **quick** (~30s): dependencies, technology, vulnerabilities, licenses
- **standard** (~2min): + ownership, dora
- **advanced** (~5min): + package-health, security-findings, iac-security, provenance
- **deep** (~10min): + Claude AI analysis
- **security** (~3min): dependencies, vulnerabilities, package-health, security-findings, iac-security, provenance

---

## New DPI Extractors

### Phase 1: High Value, Low Complexity (Immediate)

#### 1.1 Technical Debt Scanner
**File**: `utils/tech-debt/tech-debt-data.sh`
**Output**: `tech-debt.json`
**Mode**: quick, standard, advanced, deep

**Implementation**:
- Grep for TODO/FIXME/HACK/XXX comments with context
- Parse `@deprecated` annotations
- Count outdated dependencies (major versions behind)
- Detect duplicated code using `jscpd` (if available)
- Identify long files/functions exceeding thresholds

**Data Structure**:
```json
{
  "analyzer": "tech-debt",
  "version": "1.0.0",
  "summary": {
    "debt_score": 65,
    "todo_count": 45,
    "fixme_count": 12,
    "hack_count": 3,
    "deprecated_apis": 8,
    "outdated_major": 5
  },
  "markers": [
    {"type": "TODO", "file": "src/api/auth.ts", "line": 23, "text": "Add rate limiting", "age_days": 180}
  ],
  "outdated_dependencies": [
    {"name": "react", "current": "16.8.0", "latest": "18.2.0", "versions_behind": 2}
  ],
  "duplication": {
    "percentage": 4.5,
    "blocks": 23
  }
}
```

**Tools**:
- grep (built-in)
- jscpd (optional, npm)
- cloc (optional, for LOC)

---

#### 1.2 Documentation Analyzer
**File**: `utils/documentation/documentation-data.sh`
**Output**: `documentation.json`
**Mode**: standard, advanced, deep

**Implementation**:
- Check for standard documentation files (README, CONTRIBUTING, CHANGELOG, LICENSE)
- Detect API documentation (OpenAPI/Swagger specs, JSDoc, docstrings)
- Calculate comment-to-code ratio
- Inventory documentation files (docs/, ADRs, runbooks)
- Check for code of conduct, security policy

**Data Structure**:
```json
{
  "analyzer": "documentation",
  "version": "1.0.0",
  "summary": {
    "documentation_score": 72,
    "readme_exists": true,
    "api_docs_present": true,
    "comment_ratio": 0.12
  },
  "files": {
    "readme": {"path": "README.md", "size_bytes": 8500, "sections": ["Installation", "Usage", "API"]},
    "changelog": {"path": "CHANGELOG.md", "last_updated": "2024-01-15"},
    "contributing": "CONTRIBUTING.md",
    "security_policy": "SECURITY.md"
  },
  "api_documentation": {
    "openapi_spec": "openapi.yaml",
    "jsdoc_coverage": 0.45,
    "docstring_coverage": 0.32
  },
  "adrs": ["docs/adr/001-database-choice.md"],
  "missing": ["CODE_OF_CONDUCT.md", "CODEOWNERS"]
}
```

---

#### 1.3 Git Insights Analyzer
**File**: `utils/git-insights/git-insights-data.sh`
**Output**: `git-insights.json`
**Mode**: standard, advanced, deep

**Implementation**:
- Parse git log for commit patterns (frequency, timing)
- Calculate contributor activity (30/90/365 day windows)
- Identify high-churn files (frequently modified)
- Analyze code age distribution
- Detect branch patterns and PR sizes

**Data Structure**:
```json
{
  "analyzer": "git-insights",
  "version": "1.0.0",
  "summary": {
    "total_commits": 2450,
    "active_contributors_90d": 8,
    "avg_commits_per_week": 45,
    "bus_factor": 3
  },
  "contributors": [
    {"name": "alice", "email": "alice@example.com", "commits_90d": 156, "lines_added_90d": 12500}
  ],
  "high_churn_files": [
    {"file": "src/api/routes.ts", "changes_90d": 45, "contributors": 4}
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
    "avg_commit_size_lines": 45
  }
}
```

---

#### 1.4 Test Coverage Analyzer
**File**: `utils/test-coverage/test-coverage-data.sh`
**Output**: `test-coverage.json`
**Mode**: standard, advanced, deep

**Implementation**:
- Scan for test file patterns by language
- Detect test frameworks from imports/dependencies
- Calculate test-to-code ratio
- Identify directories without tests
- Detect test utilities (mocking libraries)

**Data Structure**:
```json
{
  "analyzer": "test-coverage",
  "version": "1.0.0",
  "summary": {
    "test_files": 89,
    "source_files": 245,
    "test_to_code_ratio": 0.36,
    "estimated_coverage": "medium"
  },
  "test_frameworks": ["jest", "@testing-library/react"],
  "test_types": {
    "unit": {"count": 67, "pattern": "*.test.ts"},
    "integration": {"count": 15, "pattern": "*.integration.test.ts"},
    "e2e": {"count": 7, "pattern": "e2e/*.spec.ts"}
  },
  "mocking_libraries": ["msw", "jest-mock"],
  "uncovered_directories": ["src/legacy/", "src/utils/deprecated/"]
}
```

---

### Phase 2: Security Essentials (Next Sprint)

#### 2.1 Secrets Scanner (TruffleHog Integration)
**File**: `utils/secrets-scan/secrets-scan-data.sh`
**Output**: `secrets-scan.json`
**Mode**: security, advanced, deep

**Implementation**:
- Use TruffleHog for comprehensive secret detection
- Support 800+ secret types from RAG database
- Scan git history for exposed secrets
- Detect .env file patterns (without exposing values)
- Identify secret management tools in use

**Claude Enhancement**:
- Active secret validation (attempt read-only API calls)
- Risk assessment and prioritization
- Remediation guidance

**Data Structure**:
```json
{
  "analyzer": "secrets-scan",
  "version": "1.0.0",
  "summary": {
    "secrets_found": 3,
    "severity": "high",
    "in_history": 2,
    "in_current": 1
  },
  "findings": [
    {
      "type": "aws_access_key",
      "detector": "AWS",
      "file": "config/legacy.py",
      "line": 45,
      "in_history": true,
      "commit": "abc123",
      "verified": false
    }
  ],
  "secret_management": {
    "tools_detected": ["aws-secrets-manager", "dotenv"],
    "env_vars_referenced": ["DATABASE_URL", "API_KEY"]
  }
}
```

**RAG Patterns to Create**: `rag/secrets-detection/` with patterns for all 800+ TruffleHog detector types.

---

#### 2.2 Authentication Analysis
**File**: `utils/auth-analysis/auth-analysis-data.sh`
**Output**: `auth-analysis.json`
**Mode**: security, advanced, deep

**Implementation**:
- Detect auth providers from dependencies (Auth0, Cognito, etc.)
- Identify JWT usage patterns
- Check for session management
- Detect OAuth/OIDC implementation
- Identify password hashing algorithms

**Data Structure**:
```json
{
  "analyzer": "auth-analysis",
  "version": "1.0.0",
  "summary": {
    "auth_providers": ["auth0"],
    "session_type": "jwt",
    "mfa_detected": true
  },
  "jwt": {
    "library": "jsonwebtoken",
    "patterns_found": ["sign", "verify"],
    "algorithm_hints": ["RS256"]
  },
  "oauth": {
    "providers": ["google", "github"],
    "pkce_patterns": true
  },
  "password": {
    "hashing_library": "bcrypt",
    "salt_rounds_hint": 10
  }
}
```

---

### Phase 3: Advanced Analysis (Following Sprint)

#### 3.1 Code Complexity Analyzer
**File**: `utils/complexity/complexity-data.sh`
**Output**: `complexity.json`
**Mode**: advanced, deep

**Implementation**:
- Use `lizard` for cyclomatic complexity
- Use `cloc` for lines of code
- Calculate cognitive complexity
- Identify complexity hotspots
- Generate maintainability index

**Tools Required**: lizard (pip), cloc (brew)

---

#### 3.2 Architecture Analyzer
**File**: `utils/architecture/architecture-data.sh`
**Output**: `architecture.json`
**Mode**: advanced, deep

**Implementation**:
- Use `madge` for JavaScript/TypeScript dependency graphs
- Detect circular dependencies
- Identify layer violations
- Find orphan files (not imported anywhere)
- Map entry points

**Tools Required**: madge (npm)

---

#### 3.3 API Security Analyzer
**File**: `utils/api-security/api-security-data.sh`
**Output**: `api-security.json`
**Mode**: security, advanced, deep

**Implementation**:
- Parse OpenAPI specs for endpoint inventory
- Detect authentication requirements
- Identify rate limiting patterns
- Analyze CORS configuration
- Check for GraphQL security

---

### Phase 4: Privacy & Compliance (Future)

#### 4.1 Data Privacy Analyzer
**File**: `utils/data-privacy/data-privacy-data.sh`
**Output**: `data-privacy.json`
**Mode**: advanced, deep

**Implementation**:
- Parse SQL/ORM schemas for PII patterns
- Detect sensitive field names (email, ssn, phone, etc.)
- Identify encryption patterns
- Check for data retention indicators
- Map compliance indicators (GDPR consent, data export)

**Claude Enhancement**:
- Privacy Expert Agent integration
- Regulatory compliance assessment
- Data flow analysis

**RAG Patterns**: `rag/data-privacy/pii-patterns.json` with field name patterns for detecting PII.

---

#### 4.2 Build Analysis
**File**: `utils/build-analysis/build-analysis-data.sh`
**Output**: `build-analysis.json`
**Mode**: advanced, deep

**Implementation**:
- Detect build tools (webpack, vite, gradle, etc.)
- Parse CI/CD configuration
- Analyze build scripts complexity
- Identify caching opportunities
- Estimate build duration indicators

---

## Integration Points

### 1. Adding to bootstrap.sh

For each new analyzer:

```bash
# Add to get_analyzers_for_mode()
advanced)
    echo "... tech-debt documentation git-insights test-coverage ..."
    ;;

# Add case in run_analyzer()
tech-debt)
    analyzer_script="tech-debt-data.sh"
    ;;

# Add run function
run_tech_debt_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local tech_debt_script="$UTILS_ROOT/tech-debt/tech-debt-data.sh"

    if [[ -x "$tech_debt_script" ]]; then
        "$tech_debt_script" --local-path "$repo_path" -o "$output_path/tech-debt.json" 2>/dev/null
    else
        # Fallback JSON
        cat > "$output_path/tech-debt.json" << EOF
{
  "analyzer": "tech-debt",
  "status": "analyzer_not_found",
  ...
}
EOF
    fi
}
```

### 2. Adding to hydrate.sh

```bash
# Add display name
"tech-debt") echo "Technical debt" ;;

# Add time estimate
"tech-debt") echo "~10s" ;;

# Add result parsing
"tech-debt")
    if [[ -f "$analysis_path/tech-debt.json" ]]; then
        local score=$(jq -r '.summary.debt_score // 0' "$analysis_path/tech-debt.json")
        local todos=$(jq -r '.summary.todo_count // 0' "$analysis_path/tech-debt.json")
        printf "\033[2mscore: %s, %s TODOs\033[0m" "$score" "$todos"
    fi
    ;;
```

### 3. Adding to phantom.sh

Add to recommended tools check if external tools are required.

---

## Claude-Enhanced Analyzers

For each data extractor, a corresponding Claude-enabled analyzer can be created:

```bash
# Pattern: utils/{category}/{category}-analyser.sh

# Implementation approach:
# 1. Run data extractor first
# 2. Load data extractor output
# 3. Send to Claude with specialized prompts
# 4. Generate enhanced analysis with recommendations

# Example: tech-debt-analyser.sh
if [[ "${USE_CLAUDE:-}" == "true" ]]; then
    # Load RAG context
    # Call Claude API with findings
    # Generate prioritized remediation plan
fi
```

---

## Tool Dependencies Summary

| Analyzer | Required Tools | Installation |
|----------|---------------|--------------|
| tech-debt | grep (built-in), jscpd (opt) | `npm i -g jscpd` |
| documentation | grep (built-in) | - |
| git-insights | git (built-in) | - |
| test-coverage | grep (built-in) | - |
| secrets-scan | trufflehog | `brew install trufflehog` |
| auth-analysis | grep (built-in) | - |
| complexity | lizard, cloc | `pip install lizard`, `brew install cloc` |
| architecture | madge | `npm i -g madge` |
| api-security | grep (built-in) | - |
| data-privacy | grep (built-in) | - |
| build-analysis | grep (built-in) | - |
| iac-security | checkov | `pip install checkov` (DONE) |

---

## Implementation Schedule

### Sprint 1 (Current)
- [x] IaC Security (Checkov) - COMPLETE
- [ ] Technical Debt Scanner
- [ ] Documentation Analyzer

### Sprint 2
- [ ] Git Insights Analyzer
- [ ] Test Coverage Analyzer
- [ ] Secrets Scanner (TruffleHog)

### Sprint 3
- [ ] Authentication Analysis
- [ ] Code Complexity Analyzer
- [ ] API Security Analyzer

### Sprint 4
- [ ] Architecture Analyzer
- [ ] Data Privacy Analyzer
- [ ] Build Analysis

### Sprint 5
- [ ] Claude-enhanced versions for all analyzers
- [ ] Privacy Expert Agent persona

---

## Testing Strategy

Each analyzer should include:

1. **Unit tests**: Test individual functions
2. **Integration tests**: Test against sample repositories
3. **Snapshot tests**: Compare output against known baselines

Test repositories in `phantom-tests` org:
- `phantom-tests/platform` - TypeScript/Node.js
- `phantom-tests/whisper` - Python
- `phantom-tests/mitmproxy` - Python (large)
- `phantom-tests/material-ui` - TypeScript/React

---

## Success Metrics

- All analyzers complete in < 30s for typical repos
- JSON output validates against schema
- Graceful degradation when tools unavailable
- Claude-enhanced versions provide actionable insights

---

## Database Storage Options

### Current State: Filesystem Storage

The current architecture stores all analysis results as JSON files in `~/.phantom/projects/{org}/{repo}/analysis/`. This approach has served well for development but has limitations for production use.

**Current Limitations**:
- No cross-project querying (e.g., "find all repos with critical vulnerabilities")
- No historical trend tracking without manual file management
- No concurrent access protection
- Limited scalability for large organizations (1000+ repos)
- No relationship mapping between projects
- Backup/restore requires file system operations

### Recommended Options

#### Option 1: SQLite (Embedded) - **Recommended for Single-User/Small Teams**

**Pros**:
- Zero infrastructure - single file database
- Full SQL querying capability
- Excellent for aggregation and trend analysis
- ACID compliance for data integrity
- Works offline/air-gapped
- Easy backup (single file)
- JSON1 extension for storing/querying JSON fields

**Cons**:
- Single writer limitation (fine for CLI use)
- Not suitable for concurrent multi-user access
- File locking can be problematic on network drives

**Implementation**:
```bash
# Storage location
~/.phantom/phantom.db

# Schema design
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    org TEXT NOT NULL,
    repo TEXT NOT NULL,
    url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(org, repo)
);

CREATE TABLE analyses (
    id INTEGER PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id),
    analyzer TEXT NOT NULL,
    version TEXT,
    status TEXT,
    data JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX(project_id, analyzer)
);

CREATE TABLE vulnerabilities (
    id INTEGER PRIMARY KEY,
    analysis_id INTEGER REFERENCES analyses(id),
    cve_id TEXT,
    severity TEXT,
    package TEXT,
    version TEXT,
    fixed_version TEXT,
    INDEX(severity),
    INDEX(cve_id)
);
```

**Migration Path**:
- Keep JSON files as source of truth initially
- SQLite as index/query layer
- Hybrid mode: write to both, read from SQLite for queries

---

#### Option 2: DuckDB (Analytical) - **Recommended for Analytics-Heavy Use**

**Pros**:
- Columnar storage optimized for analytics
- Excellent for time-series trend analysis
- Fast aggregations across large datasets
- Direct Parquet/JSON file reading
- Can query existing JSON files without import
- Great for "dashboard" style queries

**Cons**:
- Analytical focus, not ideal for transactional updates
- Larger memory footprint than SQLite
- Less mature ecosystem

**Implementation**:
```sql
-- Can directly query existing JSON files!
SELECT
    json_extract(data, '$.summary.critical') as critical_vulns,
    json_extract(data, '$.summary.high') as high_vulns
FROM read_json_auto('~/.phantom/projects/*/analysis/vulnerabilities.json');

-- Or use proper tables for performance
CREATE TABLE vulnerability_trends (
    project_id VARCHAR,
    scan_date DATE,
    critical INTEGER,
    high INTEGER,
    medium INTEGER,
    low INTEGER
);
```

**Best For**: Organizations wanting analytics/dashboards over historical data.

---

#### Option 3: PostgreSQL - **Recommended for Multi-User/Enterprise**

**Pros**:
- Production-grade, battle-tested
- Full concurrent access
- Advanced querying (JSONB, full-text search)
- Excellent tooling ecosystem
- Can be hosted (Supabase, Neon, RDS)
- Supports complex relationships

**Cons**:
- Requires infrastructure
- More complex setup
- Overkill for single-user CLI
- Not suitable for offline/air-gapped without local install

**Implementation**:
```sql
-- Same schema as SQLite but with JSONB
CREATE TABLE analyses (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id),
    analyzer TEXT NOT NULL,
    version TEXT,
    status TEXT,
    data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    -- JSONB indexing for fast queries
    CONSTRAINT valid_data CHECK (jsonb_typeof(data) = 'object')
);

-- GIN index for JSONB queries
CREATE INDEX idx_analyses_data ON analyses USING GIN (data);

-- Example query: Find all repos with critical vulns
SELECT p.org, p.repo, a.data->'summary'->>'critical' as critical
FROM analyses a
JOIN projects p ON p.id = a.project_id
WHERE a.analyzer = 'vulnerabilities'
  AND (a.data->'summary'->>'critical')::int > 0;
```

---

#### Option 4: MongoDB - **Alternative for Document Store Preference**

**Pros**:
- Native JSON/BSON storage
- Flexible schema (good for evolving analyzers)
- Aggregation pipeline for analytics
- Atlas for managed hosting

**Cons**:
- Requires infrastructure
- Different query paradigm
- Less suitable for relational queries across analyzers

**Best For**: Teams already using MongoDB or preferring document-oriented approach.

---

### Recommended Architecture: Hybrid Approach

For maximum flexibility, implement a **layered storage architecture**:

```
┌─────────────────────────────────────────────────────────────┐
│                      Query Layer                            │
│  (SQLite/DuckDB for local, PostgreSQL for enterprise)       │
└─────────────────────────────────────────────────────────────┘
                              ▲
                              │ Sync
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    JSON File Storage                        │
│     ~/.phantom/projects/{org}/{repo}/analysis/*.json        │
│                  (Source of Truth)                          │
└─────────────────────────────────────────────────────────────┘
```

**Implementation Phases**:

1. **Phase 1: SQLite Index Layer**
   - Add `~/.phantom/phantom.db` as query index
   - Auto-sync on analysis completion
   - Add `phantom query` command for SQL queries
   - Keep JSON files as source of truth

2. **Phase 2: Historical Tracking**
   - Store timestamped snapshots in SQLite
   - Enable trend queries ("vulns over last 30 days")
   - Add `--history` flag to analyzers

3. **Phase 3: Enterprise PostgreSQL Option**
   - Add `PHANTOM_DATABASE_URL` environment variable
   - Support PostgreSQL for multi-user deployments
   - Implement sync between local and remote

4. **Phase 4: Dashboard Integration**
   - DuckDB for analytics queries
   - Parquet export for BI tools
   - Grafana/Metabase integration

---

### Configuration

```json
// ~/.phantom/config.json
{
  "storage": {
    "backend": "sqlite",           // sqlite | postgres | filesystem
    "database_url": null,          // For postgres: postgres://...
    "keep_json_files": true,       // Keep JSON files as backup
    "history": {
      "enabled": true,
      "retention_days": 90
    }
  }
}
```

---

### Query Examples

With database storage, enable powerful queries:

```bash
# Find all repos with critical vulnerabilities
phantom query "SELECT org, repo FROM vulnerabilities WHERE severity='critical'"

# Vulnerability trend for a repo
phantom query "SELECT date, critical, high FROM vuln_history WHERE repo='platform' ORDER BY date DESC LIMIT 30"

# Technology inventory across org
phantom query "SELECT technology, COUNT(*) FROM technologies GROUP BY technology ORDER BY COUNT(*) DESC"

# Find repos using deprecated packages
phantom query "SELECT repo, package FROM dependencies WHERE deprecated=true"

# DORA metrics summary
phantom query "SELECT repo, AVG(lead_time_days), AVG(deploy_frequency) FROM dora GROUP BY repo"
```

---

### Migration Strategy

1. **Backward Compatible**: JSON files remain the default
2. **Opt-in Database**: Enable via config or CLI flag
3. **Auto-migration**: On first database use, import existing JSON
4. **Dual-write**: Write to both JSON and database during transition
5. **Gradual Adoption**: Teams can adopt database features incrementally
