# Evidence Integration Plan

**Branch:** `phuket`
**Version:** 1.0
**Status:** Planning

## Overview

Consolidate all reporting to use [Evidence](https://evidence.dev) as the single reporting mechanism. This replaces the previous markdown report generation (report.go files in each scanner) with beautiful, interactive HTML reports.

### Key Changes

1. **Remove old report.go files** - Delete markdown generators from all 9 scanners
2. **Keep terminal output** - Consistent progress + summary in terminal during scans
3. **Auto-generate Evidence reports** - Run Evidence build after each scan
4. **Clickable URL** - Print report URL in terminal for easy access

### Terminal Output Pattern (Consistent Across All Scanners)

```
$ zero hydrate expressjs/express

  Cloning expressjs/express...
  ✓ Cloned (1,234 files)

  Running scanners...
  ├─ sbom           ✓ 234 packages                    2.1s
  ├─ package-analysis ✓ 12 vulns (3 critical, 5 high)   8.4s
  ├─ code-security  ✓ 5 secrets, 23 issues             15.2s
  ├─ crypto         ✓ 2 weak ciphers                   1.8s
  ├─ devops         ✓ 45 IaC issues, 8 actions issues  12.3s
  ├─ code-quality   ✓ Score: 72/100                    4.5s
  ├─ tech-id        ✓ Node.js, React, MongoDB          3.2s
  ├─ code-ownership ✓ Bus factor: 3                    2.1s
  └─ devx           ✓ Onboarding: 65%                  1.9s

  ✓ Scan complete (51.5s)

  Generating report...
  ✓ Report ready

  ╭──────────────────────────────────────────────────────╮
  │  📊 View Report: file:///~/.zero/repos/expressjs/   │
  │     express/report/index.html                        │
  ╰──────────────────────────────────────────────────────╯
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        zero report <repo>                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  .zero/repos/<project>/analysis/     reports/template/          │
│  ├── sbom.json ──────────────────►  ├── sources/                │
│  ├── package-analysis.json           │   └── zero/              │
│  ├── code-security.json              │       ├── connection.yaml│
│  ├── devops.json                     │       └── load-data.js   │
│  ├── crypto.json                     ├── pages/                 │
│  ├── code-quality.json               │   ├── index.md           │
│  ├── tech-id.json                    │   ├── security.md        │
│  ├── code-ownership.json             │   ├── dependencies.md    │
│  └── devx.json                       │   ├── devops.md          │
│                                      │   └── quality.md         │
│                                      └── evidence.plugins.yaml  │
│                                                                  │
│  ──────────────────────────────────────────────────────────────  │
│                                                                  │
│  evidence build ──► .zero/repos/<project>/report/index.html     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow

1. **Scanner JSON** → Located in `.zero/repos/<project>/analysis/*.json`
2. **Evidence Template** → Bundled in `reports/template/`
3. **Data Loading** → JavaScript source reads JSON files
4. **Report Generation** → Evidence builds static HTML
5. **Output** → Single `report/` directory with all assets

## Migration: Remove Old Report System

### Files to Delete

```
pkg/scanners/sbom/report.go
pkg/scanners/package-analysis/report.go
pkg/scanners/code-security/report.go
pkg/scanners/crypto/report.go
pkg/scanners/devops/report.go
pkg/scanners/code-quality/report.go
pkg/scanners/code-quality/report_test.go
pkg/scanners/code-quality/README_REPORTS.md
pkg/scanners/tech-id/report.go
pkg/scanners/code-ownership/report.go
pkg/scanners/devx/report.go
pkg/report/builder.go
pkg/report/builder_test.go
pkg/report/helpers.go
pkg/report/helpers_test.go
pkg/report/types.go
```

### Files to Keep/Modify

| File | Action | Purpose |
|------|--------|---------|
| `pkg/scanner/scanner.go` | Modify | Add report generation hook after scan |
| `pkg/status/status.go` | Keep | Terminal status display |
| `pkg/terminal/terminal.go` | Keep | Terminal output formatting |

### Terminal Summary Functions

Each scanner needs a `Summary() string` method that returns a one-line summary for terminal output:

```go
// In each scanner
func (s *Scanner) Summary(result *Result) string {
    switch s.name {
    case "sbom":
        return fmt.Sprintf("%d packages", result.Summary.TotalComponents)
    case "package-analysis":
        return fmt.Sprintf("%d vulns (%d critical, %d high)",
            result.Summary.Total, result.Summary.Critical, result.Summary.High)
    case "code-security":
        return fmt.Sprintf("%d secrets, %d issues",
            result.Summary.Secrets, result.Summary.Vulns)
    // ... etc
    }
}
```

## Implementation Phases

### Phase 0: Cleanup Old Report System

1. Delete all `report.go` files from scanners
2. Delete `pkg/report/` directory
3. Add `Summary() string` method to each scanner
4. Update scanner runner to use new summary method

### Phase 1: Evidence Template Project

Create the base Evidence project structure bundled with Zero.

**Directory Structure:**
```
reports/
└── template/
    ├── evidence.plugins.yaml      # Plugin configuration
    ├── package.json               # Evidence dependencies
    ├── sources/
    │   └── zero/
    │       ├── connection.yaml    # Source configuration
    │       └── scanner-data.js    # Loads all scanner JSON
    └── pages/
        ├── index.md               # Executive summary
        ├── security.md            # Security findings
        ├── dependencies.md        # SBOM & packages
        ├── devops.md              # DevOps analysis
        ├── quality.md             # Code quality
        ├── ownership.md           # Code ownership
        └── tech-stack.md          # Technology detection
```

**Data Source (scanner-data.js):**
```javascript
import fs from 'fs';
import path from 'path';

const dataDir = process.env.ZERO_DATA_DIR || './data';

// Load all scanner outputs
const scanners = [
  'sbom', 'package-analysis', 'code-security', 'devops',
  'crypto', 'code-quality', 'technology', 'code-ownership', 'devx'
];

const data = {};
for (const scanner of scanners) {
  const filePath = path.join(dataDir, `${scanner}.json`);
  if (fs.existsSync(filePath)) {
    data[scanner] = JSON.parse(fs.readFileSync(filePath, 'utf8'));
  }
}

// Flatten findings for SQL queries
export const findings = Object.entries(data).flatMap(([scanner, result]) => {
  const items = result?.findings?.findings || result?.findings || [];
  return Array.isArray(items) ? items.map(f => ({...f, scanner})) : [];
});

export const summaries = Object.entries(data).map(([scanner, result]) => ({
  scanner,
  ...result?.summary,
  timestamp: result?.timestamp,
  duration: result?.duration_seconds
}));

export const metadata = {
  repository: data.sbom?.repository || data.devops?.repository,
  timestamp: data.sbom?.timestamp || data.devops?.timestamp,
  scanners: Object.keys(data)
};
```

### Phase 2: Report Pages

#### Executive Summary (index.md)
```markdown
---
title: Security Report
---

# {metadata.repository}

<LastRefreshed prefix="Last scan:"/>

## Security Posture

<BigValue
  data={severity_counts}
  value=critical
  title="Critical"
  comparison=high
/>

<BigValue
  data={severity_counts}
  value=high
  title="High"
/>

<BigValue
  data={severity_counts}
  value=medium
  title="Medium"
/>

## Findings by Scanner

<BarChart
  data={findings_by_scanner}
  x=scanner
  y=count
  series=severity
  type=stacked
/>

## Scanner Results

<DataTable
  data={scanner_summary}
  rows=10
/>
```

#### Security Page (security.md)
```markdown
# Security Findings

## Code Security

<DataTable
  data={code_security_findings}
  search=true
  rows=20
>
  <Column id=severity/>
  <Column id=rule_id/>
  <Column id=title/>
  <Column id=file/>
  <Column id=line/>
</DataTable>

## Cryptography Issues

<DataTable data={crypto_findings}/>

## Secrets Detected

<Alert status="warning" title="Secrets Found">
  {secrets_count} potential secrets detected in the codebase.
</Alert>
```

#### Dependencies Page (dependencies.md)
```markdown
# Dependencies

## SBOM Summary

<BigValue data={sbom_summary} value=total_components title="Total Packages"/>

## Vulnerabilities

<BarChart
  data={vuln_by_severity}
  x=severity
  y=count
  colorPalette={['#dc2626','#ea580c','#ca8a04','#65a30d']}
/>

<DataTable
  data={vulnerabilities}
  search=true
>
  <Column id=package/>
  <Column id=version/>
  <Column id=severity/>
  <Column id=cve/>
  <Column id=fix_version title="Fix Available"/>
</DataTable>

## License Distribution

<DonutChart
  data={licenses}
  name=license
  value=count
/>
```

#### DevOps Page (devops.md)
```markdown
# DevOps & Infrastructure

## DORA Metrics

<BigValue data={dora} value=deployment_frequency_class title="Deploy Frequency"/>
<BigValue data={dora} value=lead_time_class title="Lead Time"/>
<BigValue data={dora} value=change_failure_class title="Change Failure Rate"/>
<BigValue data={dora} value=mttr_class title="MTTR"/>

## IaC Security

<Heatmap
  data={iac_by_type_severity}
  x=type
  y=severity
  value=count
/>

## GitHub Actions

<DataTable data={github_actions_findings}/>

## Container Security

<DataTable data={container_findings}/>
```

### Phase 3: Auto-Generation After Scans

Modify the hydrate command to auto-generate reports after scanning:

**File: cmd/zero/cmd/hydrate.go (modifications)**
```go
func runHydrate(cmd *cobra.Command, args []string) error {
    // ... existing clone and scan logic ...

    // After all scanners complete:
    term.Info("Generating report...")

    gen := evidence.NewGenerator(zeroHome)
    reportPath, err := gen.Generate(evidence.Options{
        Repository:  repo,
        OpenBrowser: false, // Don't auto-open during batch
    })

    if err != nil {
        term.Warn("Report generation failed: %v", err)
    } else {
        term.Success("Report ready")
        term.Box(fmt.Sprintf("📊 View Report: %s", reportPath))
    }

    return nil
}
```

**File: cmd/zero/cmd/report.go**
```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/crashappsec/zero/pkg/evidence"
)

var reportCmd = &cobra.Command{
    Use:   "report <owner/repo>",
    Short: "Generate or view interactive HTML report",
    Long:  "Generate a beautiful interactive report using Evidence",
    Args:  cobra.ExactArgs(1),
    RunE:  runReport,
}

var (
    reportOutput string
    reportOpen   bool
    reportServe  bool
    reportRegen  bool
)

func init() {
    reportCmd.Flags().StringVarP(&reportOutput, "output", "o", "", "Output directory")
    reportCmd.Flags().BoolVar(&reportOpen, "open", true, "Open report in browser")
    reportCmd.Flags().BoolVar(&reportServe, "serve", false, "Start dev server")
    reportCmd.Flags().BoolVar(&reportRegen, "regenerate", false, "Force regenerate report")
    rootCmd.AddCommand(reportCmd)
}

func runReport(cmd *cobra.Command, args []string) error {
    repo := args[0]
    gen := evidence.NewGenerator(zeroHome)

    // Check if report already exists
    reportPath := gen.ReportPath(repo)
    if !reportRegen && fileExists(reportPath) {
        // Just open existing report
        if reportOpen {
            return openBrowser(reportPath)
        }
        term.Info("Report: %s", reportPath)
        return nil
    }

    opts := evidence.Options{
        Repository:  repo,
        OutputDir:   reportOutput,
        OpenBrowser: reportOpen,
        DevServer:   reportServe,
    }

    return gen.Generate(opts)
}
```

**File: pkg/evidence/generator.go**
```go
package evidence

import (
    "os"
    "os/exec"
    "path/filepath"
)

type Generator struct {
    zeroHome string
    template string
}

func NewGenerator(zeroHome string) *Generator {
    return &Generator{
        zeroHome: zeroHome,
        template: filepath.Join(zeroHome, "reports", "template"),
    }
}

func (g *Generator) Generate(opts Options) error {
    // 1. Create working directory
    workDir := filepath.Join(g.zeroHome, "repos", opts.Repository, "report-build")
    os.MkdirAll(workDir, 0755)

    // 2. Copy template
    copyDir(g.template, workDir)

    // 3. Set data directory environment
    dataDir := filepath.Join(g.zeroHome, "repos", opts.Repository, "analysis")
    os.Setenv("ZERO_DATA_DIR", dataDir)

    // 4. Install dependencies (if needed)
    if !fileExists(filepath.Join(workDir, "node_modules")) {
        exec.Command("npm", "install").Run()
    }

    // 5. Build or serve
    if opts.DevServer {
        return g.serve(workDir)
    }
    return g.build(workDir, opts)
}

func (g *Generator) build(workDir string, opts Options) error {
    // Run evidence build
    cmd := exec.Command("npx", "evidence", "build")
    cmd.Dir = workDir
    cmd.Run()

    // Copy output
    outDir := opts.OutputDir
    if outDir == "" {
        outDir = filepath.Join(g.zeroHome, "repos", opts.Repository, "report")
    }
    copyDir(filepath.Join(workDir, "build"), outDir)

    // Open browser
    if opts.OpenBrowser {
        exec.Command("open", filepath.Join(outDir, "index.html")).Run()
    }

    return nil
}
```

### Phase 4: npm Bundle Strategy

**Option A: Embedded Node.js**
- Bundle Evidence as npm package in Zero
- Use embedded Node.js runtime (like pkg or nexe)
- Pros: Single binary, no user deps
- Cons: Larger binary size (~50MB)

**Option B: npx on demand**
- Require Node.js installed on system
- Use `npx evidence` to run
- Pros: Smaller Zero binary, always latest Evidence
- Cons: Requires Node.js

**Option C: Pre-built static template**
- Generate Evidence project at build time
- Ship pre-compiled static assets
- User only needs to copy JSON data
- Pros: No runtime deps
- Cons: Less flexible

**Recommendation:** Option B (npx) for initial implementation, with clear Node.js requirement in docs. Can add Option A later for "zero-dep" experience.

### Phase 5: Styling & Branding

**Custom Theme (evidence.plugins.yaml):**
```yaml
appearance:
  colors:
    primary: '#6366f1'  # Indigo
    accent: '#8b5cf6'   # Purple
    background: '#0f172a'  # Dark slate
  fonts:
    body: 'Inter, sans-serif'
    mono: 'JetBrains Mono, monospace'
```

**Custom Components:**
- `<SeverityBadge>` - Color-coded severity indicator
- `<FindingCard>` - Expandable finding details
- `<ScannerStatus>` - Scanner health indicator
- `<TrendSparkline>` - Historical trend mini-chart

## File Structure

```
zero/
├── reports/
│   └── template/
│       ├── evidence.plugins.yaml
│       ├── package.json
│       ├── sources/
│       │   └── zero/
│       │       ├── connection.yaml
│       │       └── scanner-data.js
│       ├── pages/
│       │   ├── index.md
│       │   ├── security.md
│       │   ├── dependencies.md
│       │   ├── devops.md
│       │   ├── quality.md
│       │   ├── ownership.md
│       │   └── tech-stack.md
│       └── components/
│           ├── SeverityBadge.svelte
│           └── FindingCard.svelte
├── pkg/
│   └── evidence/
│       ├── generator.go
│       ├── template.go
│       └── browser.go
└── cmd/zero/cmd/
    └── report.go
```

## CLI Usage

```bash
# Generate and open report
zero report expressjs/express

# Generate to custom directory
zero report expressjs/express -o ./my-report

# Start dev server for live editing
zero report expressjs/express --serve

# Generate without opening browser
zero report expressjs/express --no-open
```

## Dependencies

**Node.js (runtime):**
- Node.js 18+ required
- Evidence 34.x
- DuckDB (bundled with Evidence)

**Go (build):**
- No new Go dependencies
- Uses os/exec for npm/npx

## Testing Plan

1. **Unit Tests**
   - Template file copying
   - Environment variable setting
   - Path resolution

2. **Integration Tests**
   - Full report generation with sample data
   - Dev server startup/shutdown
   - Browser opening

3. **Visual Tests**
   - Screenshot comparison of reports
   - Chart rendering validation

## Rollout

1. **Week 1:** Evidence template project, data sources
2. **Week 2:** Report pages (index, security, dependencies)
3. **Week 3:** Report pages (devops, quality, ownership)
4. **Week 4:** CLI command, browser integration
5. **Week 5:** Styling, custom components, polish

## Success Metrics

- Report generation time < 30 seconds
- Single HTML file works offline
- All 9 scanner outputs visualized
- Mobile-responsive layout
- Accessibility (WCAG 2.1 AA)

## Future Enhancements

- PDF export
- Scheduled report generation
- Email delivery
- Comparison reports (diff between scans)
- Custom report templates
- White-labeling support
