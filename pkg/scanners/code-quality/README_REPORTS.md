# Code Quality Scanner - Report Generation

The code-quality scanner generates two types of reports to help teams understand and improve code maintainability.

## Report Types

### 1. Technical Report (`code-quality-technical-report.md`)

A detailed technical report for engineers containing:

- **Technical Debt Analysis**
  - Total debt markers (TODO, FIXME, HACK, etc.)
  - Code quality issues (empty catches, debug statements, etc.)
  - Debt hotspots - files with the most technical debt
  - High-priority markers requiring immediate attention

- **Complexity Analysis**
  - Cyclomatic complexity issues
  - Long functions and deep nesting
  - Maintainability assessment
  - High-complexity functions requiring refactoring

- **Test Coverage Analysis**
  - Test frameworks detected
  - Line coverage percentages
  - Coverage report locations
  - Coverage gaps and recommendations

- **Documentation Analysis**
  - README quality assessment
  - CHANGELOG presence
  - API documentation availability
  - Documentation score and recommendations

### 2. Executive Report (`code-quality-executive-report.md`)

A high-level summary for engineering leaders containing:

- **Overall Code Health Score** (A-F grade)
- **Health Metrics Dashboard** - scores for each area
- **Key Metrics Summary** - condensed statistics
- **Critical Findings** - issues requiring immediate attention
- **Recommendations** - prioritized action items
- **Business Impact Assessment** - maintainability risk and velocity impact

## Usage

### Generate Reports Programmatically

```go
import "github.com/crashappsec/zero/pkg/scanners/code-quality"

// After running the scanner, generate reports
err := codequality.WriteReports("/path/to/analysis/directory")
if err != nil {
    log.Fatal(err)
}
```

### Generate Reports from Existing Data

```go
// Load existing scan results
data, err := codequality.LoadReportData("/path/to/analysis/directory")
if err != nil {
    log.Fatal(err)
}

// Generate individual reports
techReport := codequality.GenerateTechnicalReport(data)
execReport := codequality.GenerateExecutiveReport(data)

// Write to files or display
fmt.Println(techReport)
```

## Report Outputs

Reports are written to the analysis directory:

```
.zero/repos/<project>/analysis/
├── code-quality.json                      # Raw scan data
├── code-quality-technical-report.md       # Technical report
└── code-quality-executive-report.md       # Executive report
```

## Scoring System

### Overall Code Health Score (0-100)

Calculated as the average of:
- **Technical Debt Score** - Based on number and priority of debt markers
- **Complexity Score** - Based on complexity issues by severity
- **Test Coverage Score** - Based on line coverage percentage
- **Documentation Score** - Based on presence of README, CHANGELOG, API docs

### Grade Scale

- **A (90-100)**: Excellent code health
- **B (80-89)**: Good code health
- **C (70-79)**: Acceptable with room for improvement
- **D (60-69)**: Needs improvement
- **F (0-59)**: Critical issues requiring attention

## Key Metrics Explained

### Technical Debt

- **Debt Markers**: TODO, FIXME, HACK, BUG, WORKAROUND annotations
- **Priority Levels**: High (FIXME, HACK) vs Medium (TODO) vs Low (NOTE, IDEA)
- **Code Issues**: Empty catch blocks, debug statements, hardcoded delays
- **Hotspots**: Files with the most accumulated debt

### Complexity

- **Cyclomatic Complexity**: Number of independent paths through code
- **Long Functions**: Functions exceeding recommended length
- **Deep Nesting**: Excessive nesting levels
- **Too Many Parameters**: Functions with excessive parameters

### Test Coverage

- **Line Coverage**: Percentage of code lines executed by tests
- **Test Frameworks**: Jest, pytest, go-test, JUnit, etc.
- **Coverage Reports**: lcov.info, coverage.json, coverage.out, etc.

### Documentation

- **README**: Project overview, installation, usage
- **CHANGELOG**: Version history and changes
- **API Docs**: OpenAPI/Swagger specs or docs directory

## Business Impact Assessment

Reports include business impact analysis:

- **Maintainability Risk**: Low/Moderate/High/Critical
- **Velocity Impact**: Impact on development speed
- **Key Risks**: Primary concerns for the codebase

Example:
- **Critical** (0-39): Severe quality issues pose substantial risk
- **High** (40-59): Quality issues may slow feature development
- **Moderate** (60-79): Minor impact on development velocity
- **Low** (80-100): Minimal risk, good code health

## Integration

### CI/CD Pipeline

```bash
# Run scanner
./zero scan --scanner code-quality --output /tmp/analysis

# Generate reports
./zero report --scanner code-quality --dir /tmp/analysis

# Upload to artifact storage
aws s3 cp /tmp/analysis/*.md s3://bucket/reports/
```

### Automated Alerts

Set thresholds for automated alerts:

```go
data, _ := codequality.LoadReportData(analysisDir)
score := calculateOverallHealthScore(data.Summary)

if score < 60 {
    // Send alert to Slack/email
    alert.Send("Code health score below threshold: " + score)
}
```

## Best Practices

1. **Run regularly** - Include in CI/CD pipeline for continuous monitoring
2. **Track trends** - Monitor score changes over time
3. **Set thresholds** - Define minimum acceptable scores for each metric
4. **Prioritize fixes** - Address high-priority debt and critical complexity first
5. **Review reports** - Discuss in team retrospectives and planning sessions

## Example Report Sections

### Technical Report - Complexity Section

```markdown
## 2. Complexity Analysis

### Summary

| Metric | Value |
|--------|-------|
| Total Complexity Issues | 18 |
| High Severity | 5 |
| Medium Severity | 8 |
| Low Severity | 5 |
| Files Affected | 12 |

### Maintainability Assessment

Status: **Needs Improvement**

Moderate complexity issues present. Consider refactoring high-complexity areas to improve maintainability.

### Critical Complexity Issues

| Type | Location | Issue | Recommendation |
|------|----------|-------|----------------|
| Cyclomatic | src/processor.go:150 | High cyclomatic complexity detected | Break down into smaller functions with single responsibilities |
```

### Executive Report - Business Impact

```markdown
## Business Impact

**Maintainability Risk:** Moderate

**Developer Velocity Impact:** Minor slowdown

> **Key Risk:** Code quality issues may slow feature development and increase bug rates
```

## Customization

The report builder supports customization:

```go
// Custom report sections
b := report.NewBuilder()
b.Title("Custom Code Quality Report")
b.Section(2, "Custom Analysis")
b.Paragraph("Your custom content here")
b.Table(headers, rows)
```

## Support

For issues or questions about report generation:
- Check the source: `pkg/scanners/code-quality/report.go`
- Review tests: `pkg/scanners/code-quality/report_test.go`
- See shared utilities: `pkg/report/builder.go`
