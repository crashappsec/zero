# SBOM Scanner Reports

The SBOM scanner generates two types of reports in Markdown format:

## Report Types

### 1. Technical Report (`sbom-technical-report.md`)

Detailed technical report for engineers containing:

- **SBOM Generation**: Complete SBOM statistics
  - Component counts and breakdowns by type/ecosystem
  - Complete component listing with versions, licenses, ecosystems
  - Dependency graph statistics
  - SBOM metadata (format, version, timestamp)

- **SBOM Integrity**: Verification results
  - Lockfile comparison details
  - Missing packages (in lockfiles but not in SBOM)
  - Extra packages (in SBOM but not in lockfiles)
  - SBOM drift analysis (added, removed, changed components)

### 2. Executive Report (`sbom-executive-report.md`)

High-level summary for engineering leaders containing:

- **Executive Summary**: Overall SBOM quality score (A-F grade)
- **Key Metrics**: Coverage and integrity metrics
- **Key Findings**: Critical issues, areas for improvement, strengths
- **Recommendations**: Immediate actions and short-term improvements
- **Business Impact**: Supply chain risk and compliance readiness assessment

## Report Generation

Reports are automatically generated when the SBOM scanner runs:

```bash
./zero scan --scanner sbom --output-dir .zero/analysis
```

This creates:
- `.zero/analysis/sbom.json` - Machine-readable results
- `.zero/analysis/sbom.cdx.json` - CycloneDX SBOM
- `.zero/analysis/sbom-technical-report.md` - Technical report
- `.zero/analysis/sbom-executive-report.md` - Executive report

## Programmatic Generation

You can also generate reports programmatically:

```go
import "github.com/crashappsec/zero/pkg/scanners/sbom"

// Generate both reports
err := sbom.WriteReports("/path/to/analysis/dir")

// Or generate individually
data, err := sbom.LoadReportData("/path/to/analysis/dir")
if err != nil {
    return err
}

techReport := sbom.GenerateTechnicalReport(data)
execReport := sbom.GenerateExecutiveReport(data)
```

## Scoring Methodology

### Generation Score (0-100)
- Base score: 50 points
- Has components: +20 points
- Has dependency graph: +15 points
- Multiple ecosystems: +10 points
- 10+ components: +5 points

### Integrity Score (0-100)
- Starts at: 100 points
- Missing packages deduction:
  - 50+ missing: -40 points
  - 10-50 missing: -30 points
  - 1-10 missing: -20 points
- Extra packages deduction:
  - 50+ extra: -20 points
  - 10-50 extra: -15 points
  - 1-10 extra: -10 points
- Drift detected: -10 points
- Lockfiles found: +5 points

### Overall Score
Average of generation and integrity scores.

### Grade Scale
- A: 90-100 (Excellent)
- B: 80-89 (Good)
- C: 70-79 (Fair)
- D: 60-69 (Needs Improvement)
- F: 0-59 (Critical)

## Example Reports

### Technical Report Sections
```markdown
# SBOM Technical Report

**Repository:** `myorg/myproject`
**Generated:** 2025-12-19 10:30:00 UTC

## 1. SBOM Generation

### Summary
| Metric | Value |
|--------|-------|
| SBOM Tool | cdxgen |
| Spec Version | 1.5 |
| Total Components | 150 |
| Has Dependencies | Yes |

### Components by Ecosystem
| Ecosystem | Count |
|-----------|-------|
| npm | 100 |
| pypi | 50 |
```

### Executive Report Sections
```markdown
# SBOM Executive Report

**Repository:** `myorg/myproject`
**Date:** December 19, 2025

## Executive Summary

### Overall SBOM Quality: B (85/100)

| Area | Score | Status |
|------|-------|--------|
| **SBOM Generation** | 90/100 | Excellent |
| **SBOM Integrity** | 80/100 | Excellent |

## Recommendations

### Immediate Actions
1. Fix SBOM generation errors
2. Regenerate SBOM to include missing packages

### Short-term Improvements
1. Enable dependency graph generation
2. Update SBOM after dependency changes
```

## Integration with Other Scanners

The SBOM scanner serves as the source of truth for package data. Other scanners depend on it:

- **package-analysis**: Reads `sbom.cdx.json` for vulnerability scanning
- **tech-id**: Uses SBOM components for technology detection

The reports help assess SBOM quality before other scanners consume the data.
