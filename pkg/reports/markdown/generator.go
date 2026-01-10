// Package markdown generates markdown reports from analysis data
package markdown

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/crashappsec/zero/pkg/core/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Category represents one of the 6 dimensions of engineering intelligence
type Category string

const (
	CategorySecurity    Category = "security"
	CategorySupplyChain Category = "supply-chain"
	CategoryQuality     Category = "quality"
	CategoryDevOps      Category = "devops"
	CategoryTechnology  Category = "technology"
	CategoryTeam        Category = "team"
)

// AllCategories returns all available categories
func AllCategories() []Category {
	return []Category{
		CategorySecurity,
		CategorySupplyChain,
		CategoryQuality,
		CategoryDevOps,
		CategoryTechnology,
		CategoryTeam,
	}
}

// CategoryAnalyzers maps categories to their analyzers
var CategoryAnalyzers = map[Category][]string{
	CategorySecurity:    {"code-security"},
	CategorySupplyChain: {"code-packages"},
	CategoryQuality:     {"code-quality"},
	CategoryDevOps:      {"devops"},
	CategoryTechnology:  {"technology-identification"},
	CategoryTeam:        {"code-ownership", "developer-experience"},
}

// Options configures the report generator
type Options struct {
	Project   string   // Project ID (owner/repo)
	Category  Category // Generate report for specific category
	Analyzer  string   // Generate report for specific analyzer
	Output    string   // Output file path (empty = stdout)
}

// Generator creates markdown reports from analysis data
type Generator struct {
	cfg          *config.Config
	opts         *Options
	analysisPath string
}

// New creates a new report generator
func New(opts *Options) (*Generator, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Resolve analysis path
	parts := strings.Split(opts.Project, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid project format: expected owner/repo, got %s", opts.Project)
	}

	analysisPath := filepath.Join(cfg.ZeroHome(), "repos", parts[0], parts[1], "analysis")
	if _, err := os.Stat(analysisPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no analysis found for %s (run 'zero hydrate %s' first)", opts.Project, opts.Project)
	}

	return &Generator{
		cfg:          cfg,
		opts:         opts,
		analysisPath: analysisPath,
	}, nil
}

// Generate creates the markdown report
func (g *Generator) Generate() error {
	var output io.Writer = os.Stdout
	if g.opts.Output != "" {
		f, err := os.Create(g.opts.Output)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		output = f
	}

	// Determine what to generate
	if g.opts.Analyzer != "" {
		return g.generateAnalyzerReport(output, g.opts.Analyzer)
	}

	if g.opts.Category != "" {
		return g.generateCategoryReport(output, g.opts.Category)
	}

	// Generate aggregated report across all categories
	return g.generateAggregatedReport(output)
}

// generateAnalyzerReport generates a report for a specific analyzer
func (g *Generator) generateAnalyzerReport(w io.Writer, analyzer string) error {
	data, err := g.loadAnalyzerData(analyzer)
	if err != nil {
		return err
	}

	g.writeHeader(w, fmt.Sprintf("%s Analysis Report", analyzerDisplayName(analyzer)))
	g.writeAnalyzerSection(w, analyzer, data)
	return nil
}

// generateCategoryReport generates a report for a specific category
func (g *Generator) generateCategoryReport(w io.Writer, category Category) error {
	g.writeHeader(w, fmt.Sprintf("%s Report", categoryDisplayName(category)))

	analyzers, ok := CategoryAnalyzers[category]
	if !ok {
		return fmt.Errorf("unknown category: %s", category)
	}

	for _, analyzer := range analyzers {
		data, err := g.loadAnalyzerData(analyzer)
		if err != nil {
			// Skip if analyzer data doesn't exist
			fmt.Fprintf(w, "\n## %s\n\n*No data available*\n", analyzerDisplayName(analyzer))
			continue
		}
		g.writeAnalyzerSection(w, analyzer, data)
	}

	return nil
}

// generateAggregatedReport generates a comprehensive report across all categories
func (g *Generator) generateAggregatedReport(w io.Writer) error {
	g.writeHeader(w, "Engineering Intelligence Report")

	// Write executive summary
	g.writeExecutiveSummary(w)

	// Write each category section
	for _, category := range AllCategories() {
		fmt.Fprintf(w, "\n---\n\n# %s\n\n", categoryDisplayName(category))

		analyzers := CategoryAnalyzers[category]
		for _, analyzer := range analyzers {
			data, err := g.loadAnalyzerData(analyzer)
			if err != nil {
				continue
			}
			g.writeAnalyzerSection(w, analyzer, data)
		}
	}

	return nil
}

// writeHeader writes the report header
func (g *Generator) writeHeader(w io.Writer, title string) {
	fmt.Fprintf(w, "# %s\n\n", title)
	fmt.Fprintf(w, "**Project:** %s  \n", g.opts.Project)
	fmt.Fprintf(w, "**Generated:** %s  \n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "\n")
}

// writeExecutiveSummary writes the executive summary section
func (g *Generator) writeExecutiveSummary(w io.Writer) {
	fmt.Fprintf(w, "## Executive Summary\n\n")

	// Collect summary stats from each analyzer
	stats := make(map[string]interface{})

	// Security stats
	if data, err := g.loadAnalyzerData("code-security"); err == nil {
		if summary, ok := data["summary"].(map[string]interface{}); ok {
			stats["security"] = summary
		}
	}

	// Supply chain stats
	if data, err := g.loadAnalyzerData("code-packages"); err == nil {
		if summary, ok := data["summary"].(map[string]interface{}); ok {
			stats["supply_chain"] = summary
		}
	}

	// Write summary table
	fmt.Fprintf(w, "| Category | Status | Key Findings |\n")
	fmt.Fprintf(w, "|----------|--------|---------------|\n")

	for _, category := range AllCategories() {
		status, findings := g.getCategorySummary(category)
		fmt.Fprintf(w, "| %s | %s | %s |\n", categoryDisplayName(category), status, findings)
	}

	fmt.Fprintf(w, "\n")
}

// getCategorySummary returns a summary for a category
func (g *Generator) getCategorySummary(category Category) (status string, findings string) {
	analyzers := CategoryAnalyzers[category]

	totalCritical := 0
	totalHigh := 0
	totalFindings := 0

	for _, analyzer := range analyzers {
		data, err := g.loadAnalyzerData(analyzer)
		if err != nil {
			continue
		}

		// Extract severity counts from summary
		if summary, ok := data["summary"].(map[string]interface{}); ok {
			if counts, ok := summary["severity_counts"].(map[string]interface{}); ok {
				if c, ok := counts["critical"].(float64); ok {
					totalCritical += int(c)
				}
				if h, ok := counts["high"].(float64); ok {
					totalHigh += int(h)
				}
			}
			if total, ok := summary["total_findings"].(float64); ok {
				totalFindings += int(total)
			}
		}
	}

	// Determine status based on findings
	if totalCritical > 0 {
		status = "Critical"
	} else if totalHigh > 0 {
		status = "Warning"
	} else if totalFindings > 0 {
		status = "Info"
	} else {
		status = "Good"
	}

	if totalFindings == 0 {
		findings = "No issues found"
	} else {
		parts := []string{}
		if totalCritical > 0 {
			parts = append(parts, fmt.Sprintf("%d critical", totalCritical))
		}
		if totalHigh > 0 {
			parts = append(parts, fmt.Sprintf("%d high", totalHigh))
		}
		if len(parts) == 0 {
			findings = fmt.Sprintf("%d findings", totalFindings)
		} else {
			findings = strings.Join(parts, ", ")
		}
	}

	return status, findings
}

// writeAnalyzerSection writes a section for an analyzer
func (g *Generator) writeAnalyzerSection(w io.Writer, analyzer string, data map[string]interface{}) {
	fmt.Fprintf(w, "\n## %s\n\n", analyzerDisplayName(analyzer))

	// Write summary if available
	if summary, ok := data["summary"].(map[string]interface{}); ok {
		g.writeSummarySection(w, summary)
	}

	// Write findings if available
	if findings, ok := data["findings"].([]interface{}); ok && len(findings) > 0 {
		g.writeFindingsSection(w, findings)
	}

	// Write analyzer-specific sections
	switch analyzer {
	case "code-packages":
		g.writePackagesSection(w, data)
	case "code-security":
		g.writeSecuritySection(w, data)
	case "devops":
		g.writeDevOpsSection(w, data)
	case "technology-identification":
		g.writeTechnologySection(w, data)
	case "code-ownership":
		g.writeOwnershipSection(w, data)
	case "code-quality":
		g.writeQualitySection(w, data)
	}
}

// writeSummarySection writes a summary section
func (g *Generator) writeSummarySection(w io.Writer, summary map[string]interface{}) {
	if counts, ok := summary["severity_counts"].(map[string]interface{}); ok {
		fmt.Fprintf(w, "### Severity Distribution\n\n")
		fmt.Fprintf(w, "| Severity | Count |\n")
		fmt.Fprintf(w, "|----------|-------|\n")
		titleCaser := cases.Title(language.English)
		for _, sev := range []string{"critical", "high", "medium", "low", "info"} {
			if count, ok := counts[sev].(float64); ok && count > 0 {
				fmt.Fprintf(w, "| %s | %.0f |\n", titleCaser.String(sev), count)
			}
		}
		fmt.Fprintf(w, "\n")
	}
}

// writeFindingsSection writes a findings section
func (g *Generator) writeFindingsSection(w io.Writer, findings []interface{}) {
	// Group findings by severity
	critical := []map[string]interface{}{}
	high := []map[string]interface{}{}
	other := []map[string]interface{}{}

	for _, f := range findings {
		if finding, ok := f.(map[string]interface{}); ok {
			severity := ""
			if s, ok := finding["severity"].(string); ok {
				severity = strings.ToLower(s)
			}
			switch severity {
			case "critical":
				critical = append(critical, finding)
			case "high":
				high = append(high, finding)
			default:
				other = append(other, finding)
			}
		}
	}

	// Write critical findings
	if len(critical) > 0 {
		fmt.Fprintf(w, "### Critical Findings\n\n")
		for _, f := range critical {
			g.writeFinding(w, f)
		}
	}

	// Write high findings
	if len(high) > 0 {
		fmt.Fprintf(w, "### High Severity Findings\n\n")
		for _, f := range high {
			g.writeFinding(w, f)
		}
	}

	// Summarize other findings
	if len(other) > 0 {
		fmt.Fprintf(w, "### Other Findings\n\n")
		fmt.Fprintf(w, "*%d additional findings with lower severity*\n\n", len(other))
	}
}

// writeFinding writes a single finding
func (g *Generator) writeFinding(w io.Writer, finding map[string]interface{}) {
	title := getStringField(finding, "title", "Finding")
	severity := getStringField(finding, "severity", "unknown")
	file := getStringField(finding, "file", "")
	line := getStringField(finding, "line", "")
	message := getStringField(finding, "message", "")

	fmt.Fprintf(w, "- **%s** [%s]\n", title, severity)
	if file != "" {
		location := file
		if line != "" {
			location = fmt.Sprintf("%s:%s", file, line)
		}
		fmt.Fprintf(w, "  - Location: `%s`\n", location)
	}
	if message != "" {
		fmt.Fprintf(w, "  - %s\n", message)
	}
	fmt.Fprintf(w, "\n")
}

// writePackagesSection writes package-specific sections
func (g *Generator) writePackagesSection(w io.Writer, data map[string]interface{}) {
	// Write dependency stats
	if deps, ok := data["dependencies"].([]interface{}); ok {
		fmt.Fprintf(w, "### Dependencies\n\n")
		fmt.Fprintf(w, "Total packages: %d\n\n", len(deps))
	}

	// Write license distribution
	if licenses, ok := data["licenses"].(map[string]interface{}); ok {
		fmt.Fprintf(w, "### License Distribution\n\n")
		fmt.Fprintf(w, "| License | Count |\n")
		fmt.Fprintf(w, "|---------|-------|\n")
		for license, count := range licenses {
			if c, ok := count.(float64); ok {
				fmt.Fprintf(w, "| %s | %.0f |\n", license, c)
			}
		}
		fmt.Fprintf(w, "\n")
	}

	// Write vulnerability summary
	if vulns, ok := data["vulnerabilities"].([]interface{}); ok && len(vulns) > 0 {
		fmt.Fprintf(w, "### Vulnerabilities\n\n")
		fmt.Fprintf(w, "%d vulnerabilities found in dependencies.\n\n", len(vulns))
	}
}

// writeSecuritySection writes security-specific sections
func (g *Generator) writeSecuritySection(w io.Writer, data map[string]interface{}) {
	// Write secrets summary
	if secrets, ok := data["secrets"].([]interface{}); ok && len(secrets) > 0 {
		fmt.Fprintf(w, "### Secrets Detected\n\n")
		fmt.Fprintf(w, "%d potential secrets found.\n\n", len(secrets))
	}

	// Write crypto findings
	if crypto, ok := data["crypto"].(map[string]interface{}); ok {
		fmt.Fprintf(w, "### Cryptographic Issues\n\n")
		if ciphers, ok := crypto["weak_ciphers"].([]interface{}); ok && len(ciphers) > 0 {
			fmt.Fprintf(w, "- Weak ciphers: %d\n", len(ciphers))
		}
		if keys, ok := crypto["hardcoded_keys"].([]interface{}); ok && len(keys) > 0 {
			fmt.Fprintf(w, "- Hardcoded keys: %d\n", len(keys))
		}
		fmt.Fprintf(w, "\n")
	}
}

// writeDevOpsSection writes DevOps-specific sections
func (g *Generator) writeDevOpsSection(w io.Writer, data map[string]interface{}) {
	// Write DORA metrics if available
	if dora, ok := data["dora"].(map[string]interface{}); ok {
		fmt.Fprintf(w, "### DORA Metrics\n\n")
		fmt.Fprintf(w, "| Metric | Value |\n")
		fmt.Fprintf(w, "|--------|-------|\n")
		if df, ok := dora["deployment_frequency"].(string); ok {
			fmt.Fprintf(w, "| Deployment Frequency | %s |\n", df)
		}
		if lt, ok := dora["lead_time"].(string); ok {
			fmt.Fprintf(w, "| Lead Time | %s |\n", lt)
		}
		if cfr, ok := dora["change_failure_rate"].(float64); ok {
			fmt.Fprintf(w, "| Change Failure Rate | %.1f%% |\n", cfr*100)
		}
		if mttr, ok := dora["mttr"].(string); ok {
			fmt.Fprintf(w, "| MTTR | %s |\n", mttr)
		}
		fmt.Fprintf(w, "\n")
	}

	// Write IaC findings
	if iac, ok := data["iac"].([]interface{}); ok && len(iac) > 0 {
		fmt.Fprintf(w, "### Infrastructure as Code\n\n")
		fmt.Fprintf(w, "%d IaC findings.\n\n", len(iac))
	}

	// Write container findings
	if containers, ok := data["containers"].([]interface{}); ok && len(containers) > 0 {
		fmt.Fprintf(w, "### Container Security\n\n")
		fmt.Fprintf(w, "%d container findings.\n\n", len(containers))
	}
}

// writeTechnologySection writes technology-specific sections
func (g *Generator) writeTechnologySection(w io.Writer, data map[string]interface{}) {
	// Write detected technologies
	if techs, ok := data["technologies"].([]interface{}); ok && len(techs) > 0 {
		fmt.Fprintf(w, "### Detected Technologies\n\n")
		fmt.Fprintf(w, "| Technology | Category |\n")
		fmt.Fprintf(w, "|------------|----------|\n")
		for i, t := range techs {
			if i >= 15 { // Limit to top 15
				fmt.Fprintf(w, "| ... | *%d more* |\n", len(techs)-15)
				break
			}
			if tech, ok := t.(map[string]interface{}); ok {
				name := getStringField(tech, "name", "Unknown")
				category := getStringField(tech, "category", "")
				fmt.Fprintf(w, "| %s | %s |\n", name, category)
			}
		}
		fmt.Fprintf(w, "\n")
	}

	// Write AI/ML section
	if models, ok := data["models"].([]interface{}); ok && len(models) > 0 {
		fmt.Fprintf(w, "### AI/ML Models\n\n")
		fmt.Fprintf(w, "%d ML models detected.\n\n", len(models))
	}
}

// writeOwnershipSection writes ownership-specific sections
func (g *Generator) writeOwnershipSection(w io.Writer, data map[string]interface{}) {
	// Write bus factor
	if busFactor, ok := data["bus_factor"].(map[string]interface{}); ok {
		fmt.Fprintf(w, "### Bus Factor Analysis\n\n")
		if factor, ok := busFactor["factor"].(float64); ok {
			fmt.Fprintf(w, "**Bus Factor:** %.0f\n\n", factor)
		}
		if risk, ok := busFactor["risk"].(string); ok {
			fmt.Fprintf(w, "**Risk Level:** %s\n\n", risk)
		}
	}

	// Write top contributors
	if contributors, ok := data["contributors"].([]interface{}); ok && len(contributors) > 0 {
		fmt.Fprintf(w, "### Top Contributors\n\n")
		fmt.Fprintf(w, "| Contributor | Commits | Lines Changed |\n")
		fmt.Fprintf(w, "|-------------|---------|---------------|\n")
		for i, c := range contributors {
			if i >= 10 { // Top 10
				break
			}
			if contrib, ok := c.(map[string]interface{}); ok {
				name := getStringField(contrib, "name", "Unknown")
				commits := getStringField(contrib, "commits", "0")
				lines := getStringField(contrib, "lines_changed", "0")
				fmt.Fprintf(w, "| %s | %s | %s |\n", name, commits, lines)
			}
		}
		fmt.Fprintf(w, "\n")
	}
}

// writeQualitySection writes quality-specific sections
func (g *Generator) writeQualitySection(w io.Writer, data map[string]interface{}) {
	// Write tech debt
	if debt, ok := data["tech_debt"].(map[string]interface{}); ok {
		fmt.Fprintf(w, "### Technical Debt\n\n")
		if todos, ok := debt["todos"].(float64); ok {
			fmt.Fprintf(w, "- TODOs: %.0f\n", todos)
		}
		if fixmes, ok := debt["fixmes"].(float64); ok {
			fmt.Fprintf(w, "- FIXMEs: %.0f\n", fixmes)
		}
		if hacks, ok := debt["hacks"].(float64); ok {
			fmt.Fprintf(w, "- HACKs: %.0f\n", hacks)
		}
		fmt.Fprintf(w, "\n")
	}

	// Write complexity
	if complexity, ok := data["complexity"].(map[string]interface{}); ok {
		fmt.Fprintf(w, "### Code Complexity\n\n")
		if avg, ok := complexity["average_cyclomatic"].(float64); ok {
			fmt.Fprintf(w, "Average cyclomatic complexity: %.1f\n\n", avg)
		}
	}
}

// loadAnalyzerData loads JSON data for an analyzer
func (g *Generator) loadAnalyzerData(analyzer string) (map[string]interface{}, error) {
	filename := analyzer + ".json"
	path := filepath.Join(g.analysisPath, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filename, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}

	return result, nil
}

// Helper functions

func analyzerDisplayName(analyzer string) string {
	names := map[string]string{
		"code-packages":             "Supply Chain Analysis",
		"code-security":             "Code Security",
		"code-quality":              "Code Quality",
		"devops":                    "DevOps",
		"technology-identification": "Technology Stack",
		"code-ownership":            "Code Ownership",
		"developer-experience":      "Developer Experience",
	}
	if name, ok := names[analyzer]; ok {
		return name
	}
	return analyzer
}

func categoryDisplayName(category Category) string {
	names := map[Category]string{
		CategorySecurity:    "Security",
		CategorySupplyChain: "Supply Chain",
		CategoryQuality:     "Quality",
		CategoryDevOps:      "DevOps",
		CategoryTechnology:  "Technology",
		CategoryTeam:        "Team",
	}
	if name, ok := names[category]; ok {
		return name
	}
	return string(category)
}

func getStringField(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return fmt.Sprintf("%.0f", val)
		case int:
			return fmt.Sprintf("%d", val)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return defaultVal
}
