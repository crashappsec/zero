# Project Health Domain Knowledge

This document consolidates RAG knowledge for the **health** super scanner.

## Features Covered
- **technology**: Technology stack discovery
- **documentation**: Documentation completeness
- **tests**: Test coverage analysis
- **ownership**: Code ownership patterns

## Related RAG Directories

### Technology Identification
- `rag/technology-identification/` - Technology detection
  - `cloud-providers/` - Cloud service detection
  - `databases/` - Database technology patterns
  - Framework and language detection

### Code Ownership
- `rag/code-ownership/` - Ownership knowledge
  - CODEOWNERS format
  - Contributor analysis
  - Maintainer identification

### Frontend/Backend Engineering
- `rag/frontend-engineering/` - Frontend knowledge
- `rag/backend-engineering/` - Backend knowledge
  - Technology best practices
  - Framework conventions

### Brand
- `rag/brand/` - Project branding
  - README standards
  - Documentation templates

## Key Concepts

### Technology Discovery

#### Detection Sources
1. **Config Files**: package.json, go.mod, Cargo.toml, etc.
2. **File Extensions**: .ts, .go, .rs, .py, etc.
3. **SBOM Analysis**: Dependencies reveal frameworks
4. **Directory Structure**: Common patterns (src/, tests/, etc.)

#### Technology Categories
- **Languages**: Go, Python, JavaScript, TypeScript, Rust, Java, etc.
- **Frameworks**: React, Vue, Angular, Django, Express, etc.
- **Databases**: PostgreSQL, MySQL, MongoDB, Redis, etc.
- **Cloud Services**: AWS, GCP, Azure service detection
- **CI/CD**: GitHub Actions, GitLab CI, Jenkins, etc.
- **Build Tools**: Webpack, Vite, Maven, Gradle, etc.
- **Package Managers**: npm, yarn, pip, cargo, etc.

### Documentation Quality

#### Project Documentation
| Document | Purpose | Score Impact |
|----------|---------|--------------|
| README.md | Project overview, installation, usage | +15 points base |
| CHANGELOG.md | Version history | +5 points |
| CONTRIBUTING.md | Contribution guidelines | +5 points |
| LICENSE | Legal terms | +3 points |
| API docs | Swagger/OpenAPI | +10 points |
| Architecture docs | System design | +10 points |

#### README Quality Factors
- **Installation section**: +3 points
- **Usage examples**: +3 points
- **Code examples**: +3 points
- **Word count > 500**: +3 points
- **Badges**: Recognition of CI/quality

#### Code Documentation
- File-level documentation (package comments)
- Public function documentation
- Type/struct documentation
- Documentation ratio (files with docs / total files)

### Test Coverage

#### Coverage Metrics
- **Line Coverage**: Percentage of code lines executed
- **Branch Coverage**: Percentage of branches taken
- **Function Coverage**: Percentage of functions called

#### Coverage Sources
- **Coverage Reports**: lcov, istanbul, Go coverage profiles
- **Test File Analysis**: Test file count vs source files
- **CI Configuration**: Coverage in CI pipelines

#### Framework Detection
| Language | Test Framework | Coverage Tool |
|----------|----------------|---------------|
| Go | go test | go test -cover |
| JavaScript | Jest, Mocha | nyc, istanbul |
| Python | pytest, unittest | coverage.py |
| Java | JUnit | JaCoCo |
| Rust | cargo test | cargo-tarpaulin |

#### Coverage Thresholds
- **Critical**: No tests found
- **High concern**: < 50% coverage
- **Medium concern**: 50-80% coverage
- **Acceptable**: > 80% coverage

### Code Ownership

#### CODEOWNERS Format
```
# File pattern    Owner(s)
*.js              @frontend-team
*.go              @backend-team
/docs/            @docs-team
*                 @core-maintainers
```

#### Ownership Metrics
- **Contributor count**: Total unique contributors
- **Active contributors**: Contributors in last 30/90 days
- **Files analyzed**: Files with recent changes
- **Orphaned files**: Files with no recent commits

#### Bus Factor
- Number of contributors who account for 80% of commits
- Low bus factor (1-2) indicates risk
- High bus factor (5+) indicates healthy distribution

## Agent Expertise

### Nikon Agent
The **Nikon** agent (software architect) should be consulted for:
- Technology stack analysis
- Architecture recommendations
- Framework selection guidance

### Gibson Agent
The **Gibson** agent (engineering leader) should be consulted for:
- Documentation requirements
- Test coverage goals
- Team health metrics

### Dade/Acid Agents
Backend (**Dade**) and frontend (**Acid**) agents may assist with:
- Technology-specific guidance
- Framework best practices

## Output Schema

The health scanner produces a single `health.json` file with:
```json
{
  "features_run": ["technology", "documentation", "tests", "ownership"],
  "summary": {
    "technology": { "total_technologies": N, "primary_languages": [...], ... },
    "documentation": { "overall_score": N, "has_readme": bool, ... },
    "tests": { "overall_coverage": N, "test_framework": "...", ... },
    "ownership": { "total_contributors": N, "has_codeowners": bool, ... }
  },
  "findings": {
    "technology": { "technologies": [...] },
    "documentation": { "project_docs": {...}, "code_docs": {...}, "issues": [...] },
    "tests": { "coverage": {...}, "infrastructure": {...}, "issues": [...] },
    "ownership": { "contributors": [...], "codeowners": [...], "orphaned_files": [...] }
  }
}
```

## Scoring System

### Documentation Score (0-100)
| Component | Max Points |
|-----------|------------|
| README present + quality | 27 points |
| CHANGELOG | 5 points |
| CONTRIBUTING | 5 points |
| LICENSE | 3 points |
| Code documentation ratio | 40 points |
| API documentation | 10 points |
| Architecture documentation | 10 points |

### Health Indicators

#### Technology Health
- Modern framework versions
- Active ecosystems
- Security-focused tools present

#### Documentation Health
- Score > 70: Good
- Score 40-70: Needs improvement
- Score < 40: Critical gaps

#### Test Health
- Coverage > 80%: Good
- Coverage 50-80%: Acceptable
- Coverage < 50%: Concerning
- No tests: Critical

#### Ownership Health
- Bus factor > 3: Good
- Bus factor 2-3: Moderate risk
- Bus factor 1: High risk
- No CODEOWNERS: Missing governance
