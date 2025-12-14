# Quality Scanner

The Quality scanner provides comprehensive code quality analysis, including technical debt detection, complexity analysis, test coverage, and documentation quality assessment.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `quality` |
| **Version** | 3.2.0 |
| **Output File** | `quality.json` |
| **Dependencies** | None |
| **Estimated Time** | 30-90 seconds |

## Features

### 1. Technical Debt (`tech_debt`)

Finds technical debt markers and code issues.

**Configuration:**
```json
{
  "tech_debt": {
    "enabled": true,
    "include_markers": true,
    "include_issues": true,
    "marker_types": ["TODO", "FIXME", "HACK", "XXX", "BUG", "WORKAROUND"]
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable tech debt scanning |
| `include_markers` | bool | `true` | Scan for TODO/FIXME/etc markers |
| `include_issues` | bool | `true` | Scan for code issues |
| `marker_types` | []string | (see above) | Marker types to detect |

**Detected Markers:**

| Marker | Priority | Description |
|--------|----------|-------------|
| `FIXME` | high | Urgent fixes needed |
| `XXX` | high | Critical attention required |
| `BUG` | high | Known bug markers |
| `HACK` | high | Workaround/hack markers |
| `WORKAROUND` | high | Temporary workarounds |
| `TODO` | medium | Planned work items |
| `REFACTOR` | medium | Code needing refactoring |
| `OPTIMIZE` | medium | Performance improvements needed |
| `CLEANUP` | medium | Code cleanup needed |
| `TECH_DEBT` | medium | Explicit tech debt markers |
| `TEMP` | medium | Temporary code |
| `NOTE` | low | Informational notes |
| `IDEA` | low | Ideas for improvements |
| `REVIEW` | low | Review reminders |

**Detected Code Issues:**

| Issue Type | Severity | Description | Suggestion |
|------------|----------|-------------|------------|
| `deprecated-usage` | medium | `@deprecated` annotation found | Replace with current alternative |
| `suppressed-warning` | low | Linter suppression (`// eslint-disable`, `# noqa`, `// nosec`) | Address underlying issue |
| `debug-statement` | low | `console.log`, `console.debug`, etc. | Remove or use proper logging |
| `hardcoded-delay` | medium | `sleep(1000)`, `delay(5000)` | Use async/event-driven patterns |
| `empty-catch` | high | `catch(e) {}` with empty body | Handle or log errors |
| `magic-value` | low | Magic numbers/hardcoded values mentioned | Extract to named constant |
| `disabled-test` | medium | `DISABLED`, `SKIP`, `PENDING` test markers | Fix or remove disabled tests |
| `hard-exit` | medium | `process.exit()`, `os.exit()`, `sys.exit()` | Use graceful shutdown |

**Hotspot Detection:**
Files with the most debt markers are identified as "hotspots" - files that may need priority attention.

### 2. Complexity (`complexity`)

Detects code complexity issues using Semgrep's maintainability rules.

**Configuration:**
```json
{
  "complexity": {
    "enabled": true,
    "check_cyclomatic": true,
    "check_cognitive": true,
    "check_nesting": true,
    "max_function_lines": 50,
    "max_cyclomatic": 10
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable complexity analysis |
| `check_cyclomatic` | bool | `true` | Check cyclomatic complexity |
| `check_cognitive` | bool | `true` | Check cognitive complexity |
| `check_nesting` | bool | `true` | Check deep nesting |
| `max_function_lines` | int | `50` | Max lines per function |
| `max_cyclomatic` | int | `10` | Max cyclomatic complexity |

**Detected Issues:**

| Type | Description | Suggestion |
|------|-------------|------------|
| `complexity-cyclomatic` | High cyclomatic complexity | Break into smaller functions |
| `complexity-long-function` | Function too long | Extract logic into helpers |
| `complexity-deep-nesting` | Deep nesting levels | Use early returns, guard clauses |
| `complexity-too-many-params` | Too many parameters | Group into objects/structs |
| `complexity-cognitive` | High cognitive complexity | Simplify control flow |
| `complexity-general` | General complexity issue | Consider refactoring |

**Supported Languages:**
- Go (`.go`)
- Python (`.py`)
- JavaScript/TypeScript (`.js`, `.ts`)
- Java (`.java`)

### 3. Test Coverage (`test_coverage`)

Analyzes test infrastructure and coverage reports.

**Configuration:**
```json
{
  "test_coverage": {
    "enabled": true,
    "parse_reports": true,
    "analyze_infrastructure": true,
    "minimum_threshold": 80,
    "check_test_patterns": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable coverage analysis |
| `parse_reports` | bool | `true` | Parse coverage report files |
| `analyze_infrastructure` | bool | `true` | Analyze test setup |
| `minimum_threshold` | int | `80` | Minimum coverage percentage |
| `check_test_patterns` | bool | `true` | Check for test file patterns |

**Test Framework Detection:**

| Pattern | Framework |
|---------|-----------|
| `*_test.go` | go-test |
| `*.test.js`, `*.spec.js` | jest |
| `*.test.ts`, `*.spec.ts` | jest |
| `*.test.tsx`, `*.spec.tsx` | jest |
| `test_*.py`, `*_test.py` | pytest |
| `*Test.java` | junit |
| `*_spec.rb` | rspec |

**Coverage Report Parsing:**

| Format | File Patterns |
|--------|---------------|
| LCOV | `lcov.info`, `*.lcov` |
| Istanbul | `coverage.json`, `coverage/coverage-final.json` |
| Go Coverage | `coverage.out`, `coverage.txt` |
| Cobertura | `coverage.xml`, `cobertura.xml` |
| JaCoCo | `jacoco.xml` |

**Test Infrastructure Analysis:**
- Test file count and distribution
- Test-to-source ratio
- Framework detection
- Test configuration files (jest.config.js, pytest.ini, etc.)

**Output:**
```json
{
  "test_coverage": {
    "has_tests": true,
    "test_file_count": 145,
    "test_frameworks": ["jest", "pytest"],
    "coverage_reports": ["coverage/lcov.info"],
    "line_coverage": 78.5,
    "branch_coverage": 65.2,
    "meets_threshold": false,
    "test_to_source_ratio": 0.35,
    "coverage_by_directory": {
      "src/api/": 85.2,
      "src/core/": 72.1,
      "src/utils/": 90.5
    }
  }
}
```

### 4. Documentation (`documentation`)

Analyzes project documentation quality and completeness.

**Configuration:**
```json
{
  "documentation": {
    "enabled": true,
    "check_readme": true,
    "check_readme_quality": true,
    "check_changelog": true,
    "check_contributing": true,
    "check_api_docs": true,
    "check_code_comments": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable documentation analysis |
| `check_readme` | bool | `true` | Check for README files |
| `check_readme_quality` | bool | `true` | Analyze README quality |
| `check_changelog` | bool | `true` | Check for changelog |
| `check_contributing` | bool | `true` | Check for contributing guide |
| `check_api_docs` | bool | `true` | Check for API documentation |
| `check_code_comments` | bool | `true` | Analyze code comment density |

**Documentation Files Checked:**

| Category | Files |
|----------|-------|
| README | `README.md`, `README.rst`, `README.txt`, `readme.md` |
| Changelog | `CHANGELOG.md`, `HISTORY.md`, `CHANGES.md`, `CHANGELOG` |
| License | `LICENSE`, `LICENSE.md`, `LICENSE.txt` |
| Contributing | `CONTRIBUTING.md`, `CONTRIBUTE.md` |
| Security | `SECURITY.md` |
| Code of Conduct | `CODE_OF_CONDUCT.md` |
| API Docs | `docs/`, `api/`, `openapi.yaml`, `swagger.yaml` |

**README Quality Metrics:**
- Word count (minimum 100 for good quality)
- Has sections (markdown headers)
- Has installation instructions
- Has usage examples
- Has badges/shields
- Has table of contents (for longer docs)

**Documentation Score (0-100):**

| Component | Points |
|-----------|--------|
| README exists | 30 |
| README quality (length, structure) | 20 |
| Changelog | 10 |
| License | 15 |
| Contributing guide | 10 |
| Security policy | 10 |
| API documentation | 5 |

**Code Comment Analysis:**
- Comment density (comments per line of code)
- Public API documentation (for Go, TypeScript, Java)
- Missing function/method documentation

**Output:**
```json
{
  "documentation": {
    "score": 85,
    "has_readme": true,
    "readme_file": "README.md",
    "readme_quality": {
      "word_count": 1500,
      "has_sections": true,
      "has_installation": true,
      "has_usage": true,
      "has_badges": true,
      "quality_level": "good"
    },
    "has_changelog": true,
    "has_license": true,
    "license_type": "MIT",
    "has_contributing": true,
    "has_security_policy": true,
    "has_api_docs": true,
    "api_doc_types": ["openapi"],
    "code_comment_density": 0.15,
    "missing": []
  }
}
```

## How It Works

### Technical Flow

1. **Parallel Execution**: All features run concurrently
2. **File Scanning**: Walks repository files, skipping excluded directories
3. **Pattern Matching**: Uses regex patterns to detect markers and issues
4. **Semgrep Analysis**: Uses `p/maintainability` rules for complexity (if available)
5. **Coverage Parsing**: Parses coverage report files if found
6. **Documentation Analysis**: Checks for documentation files and quality
7. **Aggregation**: Combines results with quality scoring

### Excluded Directories

- `.git`
- `node_modules`
- `vendor`
- `dist`
- `build`
- `.venv`
- `__pycache__`
- `target`

### Scanned File Extensions

```
.go, .py, .js, .ts, .tsx, .jsx, .java, .rb, .php, .cs, .cpp, .c, .h, .hpp, .rs, .swift, .kt, .scala, .vue, .svelte
```

## Usage

### Command Line

```bash
# Run quality scanner only
./zero scan --scanner quality /path/to/repo

# Run quality profile
./zero hydrate owner/repo --profile quality-only
```

### Programmatic Usage

```go
import "github.com/crashappsec/zero/pkg/scanners/quality"

opts := &scanner.ScanOptions{
    RepoPath:  "/path/to/repo",
    OutputDir: "/path/to/output",
    FeatureConfig: map[string]interface{}{
        "tech_debt": map[string]interface{}{
            "enabled": true,
            "include_markers": true,
            "include_issues": true,
        },
        "complexity": map[string]interface{}{
            "enabled": true,
        },
        "test_coverage": map[string]interface{}{
            "enabled": true,
            "minimum_threshold": 80,
        },
        "documentation": map[string]interface{}{
            "enabled": true,
            "check_readme_quality": true,
        },
    },
}

scanner := &quality.QualityScanner{}
result, err := scanner.Run(ctx, opts)
```

## Output Format

```json
{
  "scanner": "quality",
  "version": "3.2.0",
  "metadata": {
    "features_run": ["tech_debt", "complexity", "test_coverage", "documentation"]
  },
  "summary": {
    "overall_score": 72,
    "tech_debt": {
      "total_markers": 45,
      "total_issues": 12,
      "by_type": {"TODO": 25, "FIXME": 10, "HACK": 5, "XXX": 5},
      "by_priority": {"high": 20, "medium": 20, "low": 5},
      "files_affected": 15
    },
    "complexity": {
      "total_issues": 8,
      "high": 2,
      "medium": 4,
      "low": 2,
      "files_affected": 5,
      "by_type": {
        "complexity-long-function": 3,
        "complexity-deep-nesting": 3,
        "complexity-cyclomatic": 2
      }
    },
    "test_coverage": {
      "has_tests": true,
      "test_file_count": 145,
      "test_frameworks": ["jest", "pytest"],
      "line_coverage": 78.5,
      "meets_threshold": false
    },
    "documentation": {
      "score": 85,
      "has_readme": true,
      "has_changelog": true,
      "has_license": true,
      "has_contributing": true
    },
    "errors": []
  },
  "findings": {
    "tech_debt": {
      "markers": [
        {
          "type": "TODO",
          "priority": "medium",
          "file": "src/utils/helpers.js",
          "line": 42,
          "text": "TODO: Refactor this to use async/await"
        }
      ],
      "issues": [
        {
          "type": "empty-catch",
          "severity": "high",
          "file": "src/api/client.js",
          "line": 78,
          "description": "Empty catch block swallows errors",
          "suggestion": "Handle or log errors appropriately"
        }
      ],
      "hotspots": [
        {
          "file": "src/legacy/processor.js",
          "total_markers": 12,
          "by_type": {"TODO": 8, "FIXME": 4}
        }
      ]
    },
    "complexity": {...},
    "test_coverage": {...},
    "documentation": {...}
  }
}
```

## Prerequisites

| Tool | Required For | Install Command |
|------|--------------|-----------------|
| semgrep | complexity feature | `pip install semgrep` or `brew install semgrep` |

**Note:** tech_debt, test_coverage, and documentation features work without any external tools.

## Profiles

| Profile | tech_debt | complexity | test_coverage | documentation |
|---------|-----------|------------|---------------|---------------|
| `quick` | - | - | - | - |
| `standard` | Yes | - | Yes | Yes |
| `full` | Yes | Yes | Yes | Yes |
| `quality-only` | Yes | Yes | Yes | Yes |

## Related Scanners

- **code-security**: Complements with security analysis
- **ownership**: Code ownership often correlates with quality
- **health**: Overall project health metrics

## See Also

- [Code Security Scanner](code-security.md) - Security-focused code analysis
- [Ownership Scanner](ownership.md) - Code ownership analysis
- [Semgrep Maintainability Rules](https://semgrep.dev/p/maintainability) - Complexity detection rules
