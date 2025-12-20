package devops

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/crashappsec/zero/pkg/report"
)

// ReportData holds the data needed to generate reports
type ReportData struct {
	Repository string
	Timestamp  time.Time
	Summary    Summary
	Findings   Findings
}

// LoadReportData loads devops.json from the analysis directory
func LoadReportData(analysisDir string) (*ReportData, error) {
	devopsPath := filepath.Join(analysisDir, "devops.json")
	data, err := os.ReadFile(devopsPath)
	if err != nil {
		return nil, fmt.Errorf("reading devops.json: %w", err)
	}

	var result struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing devops.json: %w", err)
	}

	return &ReportData{
		Repository: result.Repository,
		Timestamp:  result.Timestamp,
		Summary:    result.Summary,
		Findings:   result.Findings,
	}, nil
}

// GenerateTechnicalReport creates a detailed technical report for security engineers
func GenerateTechnicalReport(data *ReportData) string {
	b := report.NewBuilder()

	b.Title("DevOps Security Report - Technical Analysis")

	meta := report.ReportMeta{
		Repository:  data.Repository,
		Timestamp:   data.Timestamp,
		ScannerDesc: "DevOps Security Scanner - Infrastructure, Containers, CI/CD, and DORA Metrics",
	}
	b.Meta(meta)

	// Executive Summary
	b.Section(2, "Executive Summary")
	b.Paragraph(generateExecutiveSummary(data))

	// IaC Security
	if data.Summary.IaC != nil {
		generateIaCSection(b, data)
	}

	// Container Security
	if data.Summary.Containers != nil {
		generateContainersSection(b, data)
	}

	// GitHub Actions Security
	if data.Summary.GitHubActions != nil {
		generateGitHubActionsSection(b, data)
	}

	// DORA Metrics
	if data.Summary.DORA != nil {
		generateDORASection(b, data)
	}

	// Git Analysis
	if data.Summary.Git != nil {
		generateGitSection(b, data)
	}

	// Errors
	if len(data.Summary.Errors) > 0 {
		b.Section(2, "Scan Errors")
		b.List(data.Summary.Errors)
	}

	b.Footer("DevOps")

	return b.String()
}

func generateExecutiveSummary(data *ReportData) string {
	var findings []string

	if data.Summary.IaC != nil {
		total := data.Summary.IaC.Critical + data.Summary.IaC.High + data.Summary.IaC.Medium + data.Summary.IaC.Low
		findings = append(findings, fmt.Sprintf("**IaC Security:** %d findings (%d critical, %d high)",
			total, data.Summary.IaC.Critical, data.Summary.IaC.High))
	}

	if data.Summary.Containers != nil {
		total := data.Summary.Containers.Critical + data.Summary.Containers.High + data.Summary.Containers.Medium + data.Summary.Containers.Low
		findings = append(findings, fmt.Sprintf("**Container Security:** %d findings (%d critical, %d high)",
			total, data.Summary.Containers.Critical, data.Summary.Containers.High))
	}

	if data.Summary.GitHubActions != nil {
		total := data.Summary.GitHubActions.Critical + data.Summary.GitHubActions.High + data.Summary.GitHubActions.Medium + data.Summary.GitHubActions.Low
		findings = append(findings, fmt.Sprintf("**CI/CD Security:** %d findings (%d critical, %d high)",
			total, data.Summary.GitHubActions.Critical, data.Summary.GitHubActions.High))
	}

	if data.Summary.DORA != nil {
		findings = append(findings, fmt.Sprintf("**DORA Performance:** %s class",
			strings.ToUpper(data.Summary.DORA.OverallClass)))
	}

	if data.Summary.Git != nil {
		findings = append(findings, fmt.Sprintf("**Team Health:** %d contributors, bus factor %d",
			data.Summary.Git.TotalContributors, data.Summary.Git.BusFactor))
	}

	return strings.Join(findings, "\n\n")
}

func generateIaCSection(b *report.ReportBuilder, data *ReportData) {
	iac := data.Summary.IaC

	b.Section(2, "Infrastructure as Code Security")

	if iac.Error != "" {
		b.Paragraph(fmt.Sprintf("Error: %s", iac.Error))
		return
	}

	// Summary stats
	b.Section(3, "Summary")

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Total Findings", fmt.Sprintf("%d", iac.TotalFindings)},
		{"Critical", fmt.Sprintf("%d", iac.Critical)},
		{"High", fmt.Sprintf("%d", iac.High)},
		{"Medium", fmt.Sprintf("%d", iac.Medium)},
		{"Low", fmt.Sprintf("%d", iac.Low)},
		{"Files Scanned", fmt.Sprintf("%d", iac.FilesScanned)},
		{"Tool", iac.Tool},
	}
	b.Table(headers, rows)

	// By type breakdown
	if len(iac.ByType) > 0 {
		b.Section(3, "Findings by Infrastructure Type")

		typeHeaders := []string{"Type", "Count"}
		var typeRows [][]string

		// Sort by count descending
		type typeCount struct {
			typ   string
			count int
		}
		var typeCounts []typeCount
		for t, c := range iac.ByType {
			typeCounts = append(typeCounts, typeCount{t, c})
		}
		sort.Slice(typeCounts, func(i, j int) bool {
			return typeCounts[i].count > typeCounts[j].count
		})

		for _, tc := range typeCounts {
			typeRows = append(typeRows, []string{tc.typ, fmt.Sprintf("%d", tc.count)})
		}
		b.Table(typeHeaders, typeRows)
	}

	// Detailed findings
	if len(data.Findings.IaC) > 0 {
		b.Section(3, "Critical and High Severity Findings")

		// Filter critical and high
		var criticalHigh []IaCFinding
		for _, f := range data.Findings.IaC {
			if strings.ToLower(f.Severity) == "critical" || strings.ToLower(f.Severity) == "high" {
				criticalHigh = append(criticalHigh, f)
			}
		}

		// Sort by severity (critical first)
		sort.Slice(criticalHigh, func(i, j int) bool {
			si := severityToInt(criticalHigh[i].Severity)
			sj := severityToInt(criticalHigh[j].Severity)
			return si < sj
		})

		if len(criticalHigh) > 0 {
			headers := []string{"Severity", "Rule ID", "Title", "File", "Resource"}
			var rows [][]string

			for _, f := range criticalHigh {
				line := ""
				if f.Line > 0 {
					line = fmt.Sprintf(":%d", f.Line)
				}
				resource := f.Resource
				if resource == "" {
					resource = "-"
				}
				rows = append(rows, []string{
					strings.ToUpper(f.Severity),
					f.RuleID,
					f.Title,
					f.File + line,
					resource,
				})
			}
			b.Table(headers, rows)

			// Show details for critical findings
			var criticalOnly []IaCFinding
			for _, f := range criticalHigh {
				if strings.ToLower(f.Severity) == "critical" {
					criticalOnly = append(criticalOnly, f)
				}
			}

			if len(criticalOnly) > 0 {
				b.Section(4, "Critical Findings Details")
				for _, f := range criticalOnly {
					b.KeyValue("Rule", f.RuleID)
					b.KeyValue("Title", f.Title)
					b.KeyValue("File", f.File)
					if f.Line > 0 {
						b.KeyValue("Line", fmt.Sprintf("%d", f.Line))
					}
					if f.Resource != "" {
						b.KeyValue("Resource", f.Resource)
					}
					if f.Description != "" {
						b.Paragraph(fmt.Sprintf("**Description:** %s", f.Description))
					}
					if f.Resolution != "" {
						b.Paragraph(fmt.Sprintf("**Resolution:** %s", f.Resolution))
					}
					b.Divider()
				}
			}
		} else {
			b.Paragraph("No critical or high severity findings.")
		}

		// Show all findings count by type
		if iac.TotalFindings > len(criticalHigh) {
			b.Section(4, "Additional Findings")
			b.Paragraph(fmt.Sprintf("%d medium and low severity findings were identified. Review the full devops.json for details.",
				iac.TotalFindings-len(criticalHigh)))
		}
	}
}

func generateContainersSection(b *report.ReportBuilder, data *ReportData) {
	containers := data.Summary.Containers

	b.Section(2, "Container Security")

	if containers.Error != "" {
		b.Paragraph(fmt.Sprintf("Error: %s", containers.Error))
		return
	}

	// Summary stats
	b.Section(3, "Summary")

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Total Findings", fmt.Sprintf("%d", containers.TotalFindings)},
		{"Critical", fmt.Sprintf("%d", containers.Critical)},
		{"High", fmt.Sprintf("%d", containers.High)},
		{"Medium", fmt.Sprintf("%d", containers.Medium)},
		{"Low", fmt.Sprintf("%d", containers.Low)},
		{"Dockerfiles Scanned", fmt.Sprintf("%d", containers.DockerfilesScanned)},
		{"Images Scanned", fmt.Sprintf("%d", containers.ImagesScanned)},
	}
	b.Table(headers, rows)

	// By image breakdown
	if len(containers.ByImage) > 0 {
		b.Section(3, "Findings by Container Image")

		imageHeaders := []string{"Image", "Vulnerabilities"}
		var imageRows [][]string

		// Sort by count descending
		type imageCount struct {
			image string
			count int
		}
		var imageCounts []imageCount
		for img, c := range containers.ByImage {
			imageCounts = append(imageCounts, imageCount{img, c})
		}
		sort.Slice(imageCounts, func(i, j int) bool {
			return imageCounts[i].count > imageCounts[j].count
		})

		for _, ic := range imageCounts {
			imageRows = append(imageRows, []string{ic.image, fmt.Sprintf("%d", ic.count)})
		}
		b.Table(imageHeaders, imageRows)
	}

	// Detailed findings
	if len(data.Findings.Containers) > 0 {
		b.Section(3, "Critical and High Severity Vulnerabilities")

		// Filter critical and high
		var criticalHigh []ContainerFinding
		for _, f := range data.Findings.Containers {
			if strings.ToLower(f.Severity) == "critical" || strings.ToLower(f.Severity) == "high" {
				criticalHigh = append(criticalHigh, f)
			}
		}

		// Sort by CVSS score descending
		sort.Slice(criticalHigh, func(i, j int) bool {
			return criticalHigh[i].CVSS > criticalHigh[j].CVSS
		})

		if len(criticalHigh) > 0 {
			headers := []string{"Severity", "CVE", "Package", "Version", "Fixed In", "CVSS"}
			var rows [][]string

			for _, f := range criticalHigh {
				fixedVersion := f.FixedVersion
				if fixedVersion == "" {
					fixedVersion = "N/A"
				}
				cvss := "-"
				if f.CVSS > 0 {
					cvss = fmt.Sprintf("%.1f", f.CVSS)
				}
				rows = append(rows, []string{
					strings.ToUpper(f.Severity),
					f.VulnID,
					f.Package,
					f.Version,
					fixedVersion,
					cvss,
				})
			}
			b.Table(headers, rows)

			// Show details for top critical vulnerabilities
			var topCritical []ContainerFinding
			count := 0
			for _, f := range criticalHigh {
				if strings.ToLower(f.Severity) == "critical" && count < 10 {
					topCritical = append(topCritical, f)
					count++
				}
			}

			if len(topCritical) > 0 {
				b.Section(4, "Top Critical Vulnerabilities")
				for _, f := range topCritical {
					b.KeyValue("Vulnerability ID", f.VulnID)
					b.KeyValue("Title", f.Title)
					b.KeyValue("Package", fmt.Sprintf("%s@%s", f.Package, f.Version))
					b.KeyValue("Image", f.Image)
					if f.Dockerfile != "" {
						b.KeyValue("Dockerfile", f.Dockerfile)
					}
					if f.FixedVersion != "" {
						b.KeyValue("Fixed In", f.FixedVersion)
					}
					if f.CVSS > 0 {
						b.KeyValue("CVSS Score", fmt.Sprintf("%.1f", f.CVSS))
					}
					if f.Description != "" {
						b.Paragraph(fmt.Sprintf("**Description:** %s", f.Description))
					}
					if len(f.References) > 0 {
						b.Paragraph("**References:**")
						b.List(f.References)
					}
					b.Divider()
				}
			}
		} else {
			b.Paragraph("No critical or high severity vulnerabilities found.")
		}

		if containers.TotalFindings > len(criticalHigh) {
			b.Section(4, "Additional Vulnerabilities")
			b.Paragraph(fmt.Sprintf("%d medium and low severity vulnerabilities were identified. Review the full devops.json for details.",
				containers.TotalFindings-len(criticalHigh)))
		}
	}
}

func generateGitHubActionsSection(b *report.ReportBuilder, data *ReportData) {
	gha := data.Summary.GitHubActions

	b.Section(2, "GitHub Actions Security")

	if gha.Error != "" {
		b.Paragraph(fmt.Sprintf("Error: %s", gha.Error))
		return
	}

	// Summary stats
	b.Section(3, "Summary")

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Total Findings", fmt.Sprintf("%d", gha.TotalFindings)},
		{"Critical", fmt.Sprintf("%d", gha.Critical)},
		{"High", fmt.Sprintf("%d", gha.High)},
		{"Medium", fmt.Sprintf("%d", gha.Medium)},
		{"Low", fmt.Sprintf("%d", gha.Low)},
		{"Workflows Scanned", fmt.Sprintf("%d", gha.WorkflowsScanned)},
	}
	b.Table(headers, rows)

	// By category breakdown
	if len(gha.ByCategory) > 0 {
		b.Section(3, "Findings by Category")

		catHeaders := []string{"Category", "Count"}
		var catRows [][]string

		// Sort by count descending
		type catCount struct {
			cat   string
			count int
		}
		var catCounts []catCount
		for cat, c := range gha.ByCategory {
			catCounts = append(catCounts, catCount{cat, c})
		}
		sort.Slice(catCounts, func(i, j int) bool {
			return catCounts[i].count > catCounts[j].count
		})

		for _, cc := range catCounts {
			catRows = append(catRows, []string{cc.cat, fmt.Sprintf("%d", cc.count)})
		}
		b.Table(catHeaders, catRows)
	}

	// Detailed findings
	if len(data.Findings.GitHubActions) > 0 {
		b.Section(3, "Critical and High Severity Findings")

		// Filter critical and high
		var criticalHigh []GitHubActionsFinding
		for _, f := range data.Findings.GitHubActions {
			if strings.ToLower(f.Severity) == "critical" || strings.ToLower(f.Severity) == "high" {
				criticalHigh = append(criticalHigh, f)
			}
		}

		// Sort by severity
		sort.Slice(criticalHigh, func(i, j int) bool {
			si := severityToInt(criticalHigh[i].Severity)
			sj := severityToInt(criticalHigh[j].Severity)
			return si < sj
		})

		if len(criticalHigh) > 0 {
			headers := []string{"Severity", "Category", "Title", "File", "Suggestion"}
			var rows [][]string

			for _, f := range criticalHigh {
				line := ""
				if f.Line > 0 {
					line = fmt.Sprintf(":%d", f.Line)
				}
				suggestion := f.Suggestion
				if suggestion == "" {
					suggestion = "-"
				}
				// Truncate long suggestions
				if len(suggestion) > 50 {
					suggestion = suggestion[:47] + "..."
				}
				rows = append(rows, []string{
					strings.ToUpper(f.Severity),
					f.Category,
					f.Title,
					f.File + line,
					suggestion,
				})
			}
			b.Table(headers, rows)

			// Show details for critical findings
			var criticalOnly []GitHubActionsFinding
			for _, f := range criticalHigh {
				if strings.ToLower(f.Severity) == "critical" {
					criticalOnly = append(criticalOnly, f)
				}
			}

			if len(criticalOnly) > 0 {
				b.Section(4, "Critical Findings Details")
				for _, f := range criticalOnly {
					b.KeyValue("Rule", f.RuleID)
					b.KeyValue("Title", f.Title)
					b.KeyValue("Category", f.Category)
					b.KeyValue("File", f.File)
					if f.Line > 0 {
						b.KeyValue("Line", fmt.Sprintf("%d", f.Line))
					}
					if f.Description != "" {
						b.Paragraph(fmt.Sprintf("**Description:** %s", f.Description))
					}
					if f.Suggestion != "" {
						b.Paragraph(fmt.Sprintf("**Suggestion:** %s", f.Suggestion))
					}
					b.Divider()
				}
			}
		} else {
			b.Paragraph("No critical or high severity findings.")
		}

		if gha.TotalFindings > len(criticalHigh) {
			b.Section(4, "Additional Findings")
			b.Paragraph(fmt.Sprintf("%d medium and low severity findings were identified. Review the full devops.json for details.",
				gha.TotalFindings-len(criticalHigh)))
		}
	}
}

func generateDORASection(b *report.ReportBuilder, data *ReportData) {
	dora := data.Summary.DORA

	b.Section(2, "DORA Metrics")

	if dora.Error != "" {
		b.Paragraph(fmt.Sprintf("Error: %s", dora.Error))
		return
	}

	// Overall classification
	b.Section(3, "Performance Classification")
	b.Paragraph(fmt.Sprintf("**Overall DORA Performance:** %s", strings.ToUpper(dora.OverallClass)))
	b.Paragraph(fmt.Sprintf("*Analysis based on %d days of data*", dora.PeriodDays))

	// Metrics breakdown
	b.Section(3, "Key Metrics")

	headers := []string{"Metric", "Value", "Classification"}
	rows := [][]string{
		{"Deployment Frequency", fmt.Sprintf("%.2f per day", dora.DeploymentFrequency), dora.DeploymentFrequencyClass},
		{"Lead Time for Changes", fmt.Sprintf("%.1f hours", dora.LeadTimeHours), dora.LeadTimeClass},
		{"Change Failure Rate", fmt.Sprintf("%.1f%%", dora.ChangeFailureRate*100), dora.ChangeFailureClass},
		{"Time to Restore", fmt.Sprintf("%.1f hours", dora.MTTRHours), dora.MTTRClass},
	}
	b.Table(headers, rows)

	// Detailed deployment data
	if data.Findings.DORA != nil {
		metrics := data.Findings.DORA

		b.Section(3, "Deployment Statistics")
		b.KeyValue("Total Deployments", fmt.Sprintf("%d", metrics.TotalDeployments))
		b.KeyValue("Total Commits", fmt.Sprintf("%d", metrics.TotalCommits))

		if len(metrics.Deployments) > 0 {
			b.Section(4, "Recent Deployments")

			depHeaders := []string{"Tag", "Date", "Commits", "Type"}
			var depRows [][]string

			// Show last 10 deployments
			limit := 10
			if len(metrics.Deployments) < limit {
				limit = len(metrics.Deployments)
			}

			for i := 0; i < limit; i++ {
				d := metrics.Deployments[i]
				depType := "Feature"
				if d.IsFix {
					depType = "Fix"
				}
				depRows = append(depRows, []string{
					d.Tag,
					d.Date.Format("2006-01-02"),
					fmt.Sprintf("%d", d.Commits),
					depType,
				})
			}
			b.Table(depHeaders, depRows)

			if len(metrics.Deployments) > 10 {
				b.Paragraph(fmt.Sprintf("*Showing 10 most recent of %d total deployments*", len(metrics.Deployments)))
			}
		}

		// Performance insights
		b.Section(3, "Performance Insights")

		insights := generateDORAInsights(dora)
		if len(insights) > 0 {
			b.List(insights)
		}
	}
}

func generateGitSection(b *report.ReportBuilder, data *ReportData) {
	git := data.Summary.Git

	b.Section(2, "Team & Repository Health")

	if git.Error != "" {
		b.Paragraph(fmt.Sprintf("Error: %s", git.Error))
		return
	}

	// Summary stats
	b.Section(3, "Team Metrics")

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Total Contributors", fmt.Sprintf("%d", git.TotalContributors)},
		{"Active Contributors (30d)", fmt.Sprintf("%d", git.ActiveContributors30d)},
		{"Active Contributors (90d)", fmt.Sprintf("%d", git.ActiveContributors90d)},
		{"Bus Factor", fmt.Sprintf("%d", git.BusFactor)},
		{"Total Commits", fmt.Sprintf("%d", git.TotalCommits)},
		{"Recent Commits (90d)", fmt.Sprintf("%d", git.Commits90d)},
		{"Activity Level", git.ActivityLevel},
	}
	b.Table(headers, rows)

	// Bus factor analysis
	b.Section(3, "Bus Factor Analysis")
	busFactor := git.BusFactor
	var busSummary string
	switch {
	case busFactor >= 5:
		busSummary = "Healthy knowledge distribution across the team. Low risk of knowledge loss."
	case busFactor >= 3:
		busSummary = "Moderate knowledge distribution. Consider cross-training to reduce risk."
	case busFactor >= 2:
		busSummary = "Limited knowledge distribution. High risk if key contributors leave."
	default:
		busSummary = "Critical risk: Repository knowledge concentrated in very few contributors."
	}
	b.Paragraph(busSummary)

	// Contributor details
	if data.Findings.Git != nil && len(data.Findings.Git.Contributors) > 0 {
		b.Section(3, "Top Contributors (Last 90 Days)")

		// Sort by 90d commits
		contributors := make([]Contributor, len(data.Findings.Git.Contributors))
		copy(contributors, data.Findings.Git.Contributors)
		sort.Slice(contributors, func(i, j int) bool {
			return contributors[i].Commits90d > contributors[j].Commits90d
		})

		contribHeaders := []string{"Name", "Total Commits", "90d Commits", "90d Lines Added", "90d Lines Removed"}
		var contribRows [][]string

		limit := 10
		if len(contributors) < limit {
			limit = len(contributors)
		}

		for i := 0; i < limit; i++ {
			c := contributors[i]
			contribRows = append(contribRows, []string{
				c.Name,
				fmt.Sprintf("%d", c.TotalCommits),
				fmt.Sprintf("%d", c.Commits90d),
				fmt.Sprintf("+%d", c.LinesAdded90d),
				fmt.Sprintf("-%d", c.LinesRemoved90d),
			})
		}
		b.Table(contribHeaders, contribRows)
	}

	// High churn files
	if data.Findings.Git != nil && len(data.Findings.Git.HighChurnFiles) > 0 {
		b.Section(3, "High Churn Files (90 Days)")
		b.Paragraph("Files with frequent changes may indicate technical debt or evolving requirements:")

		churnHeaders := []string{"File", "Changes", "Contributors"}
		var churnRows [][]string

		limit := 10
		if len(data.Findings.Git.HighChurnFiles) < limit {
			limit = len(data.Findings.Git.HighChurnFiles)
		}

		for i := 0; i < limit; i++ {
			f := data.Findings.Git.HighChurnFiles[i]
			churnRows = append(churnRows, []string{
				f.File,
				fmt.Sprintf("%d", f.Changes90d),
				fmt.Sprintf("%d", f.Contributors),
			})
		}
		b.Table(churnHeaders, churnRows)
	}

	// Code age
	if data.Findings.Git != nil && data.Findings.Git.CodeAge != nil {
		age := data.Findings.Git.CodeAge
		b.Section(3, "Code Age Distribution")
		b.Paragraph(fmt.Sprintf("Analysis based on %d sampled files:", age.SampledFiles))

		ageHeaders := []string{"Age Range", "Files", "Percentage"}
		ageRows := [][]string{
			{"0-30 days", fmt.Sprintf("%d", age.Age0to30.Count), fmt.Sprintf("%.1f%%", age.Age0to30.Percentage)},
			{"31-90 days", fmt.Sprintf("%d", age.Age31to90.Count), fmt.Sprintf("%.1f%%", age.Age31to90.Percentage)},
			{"91-365 days", fmt.Sprintf("%d", age.Age91to365.Count), fmt.Sprintf("%.1f%%", age.Age91to365.Percentage)},
			{"365+ days", fmt.Sprintf("%d", age.Age365Plus.Count), fmt.Sprintf("%.1f%%", age.Age365Plus.Percentage)},
		}
		b.Table(ageHeaders, ageRows)
	}

	// Commit patterns
	if data.Findings.Git != nil && data.Findings.Git.Patterns != nil {
		p := data.Findings.Git.Patterns
		b.Section(3, "Commit Patterns")

		patternHeaders := []string{"Pattern", "Value"}
		patternRows := [][]string{
			{"Most Active Day", p.MostActiveDay},
			{"Most Active Hour", fmt.Sprintf("%02d:00", p.MostActiveHour)},
			{"Avg Commit Size", fmt.Sprintf("%d lines", p.AvgCommitSizeLines)},
			{"Avg Commits/Week", fmt.Sprintf("%d", p.AvgCommitsPerWeek)},
			{"First Commit", p.FirstCommit},
			{"Last Commit", p.LastCommit},
		}
		b.Table(patternHeaders, patternRows)
	}
}

// GenerateExecutiveReport creates a high-level summary for engineering leaders
func GenerateExecutiveReport(data *ReportData) string {
	b := report.NewBuilder()

	b.Title("DevOps Maturity Report - Executive Summary")

	meta := report.ReportMeta{
		Repository: data.Repository,
		Timestamp:  data.Timestamp,
	}
	b.Meta(meta)

	// Overall maturity assessment
	b.Section(2, "DevOps Maturity Assessment")

	maturity := assessDevOpsMaturity(data)
	b.Paragraph(fmt.Sprintf("**Overall Maturity:** %s", maturity.overall))
	b.Paragraph(maturity.summary)

	// Key metrics
	b.Section(2, "Key Metrics")

	var metrics []string

	if data.Summary.IaC != nil {
		riskLevel := "Low"
		criticalHigh := data.Summary.IaC.Critical + data.Summary.IaC.High
		if criticalHigh > 10 {
			riskLevel = "High"
		} else if criticalHigh > 5 {
			riskLevel = "Medium"
		}
		metrics = append(metrics, fmt.Sprintf("**Infrastructure Security Risk:** %s (%d critical/high findings)",
			riskLevel, criticalHigh))
	}

	if data.Summary.Containers != nil {
		riskLevel := "Low"
		criticalHigh := data.Summary.Containers.Critical + data.Summary.Containers.High
		if criticalHigh > 20 {
			riskLevel = "High"
		} else if criticalHigh > 10 {
			riskLevel = "Medium"
		}
		metrics = append(metrics, fmt.Sprintf("**Container Security Risk:** %s (%d critical/high vulnerabilities)",
			riskLevel, criticalHigh))
	}

	if data.Summary.GitHubActions != nil {
		riskLevel := "Low"
		criticalHigh := data.Summary.GitHubActions.Critical + data.Summary.GitHubActions.High
		if criticalHigh > 5 {
			riskLevel = "High"
		} else if criticalHigh > 2 {
			riskLevel = "Medium"
		}
		metrics = append(metrics, fmt.Sprintf("**CI/CD Pipeline Risk:** %s (%d critical/high findings)",
			riskLevel, criticalHigh))
	}

	if data.Summary.DORA != nil {
		metrics = append(metrics, fmt.Sprintf("**DORA Performance:** %s class (Deployment frequency: %.2f/day, Lead time: %.1fh)",
			strings.ToUpper(data.Summary.DORA.OverallClass),
			data.Summary.DORA.DeploymentFrequency,
			data.Summary.DORA.LeadTimeHours))
	}

	if data.Summary.Git != nil {
		busFactor := data.Summary.Git.BusFactor
		busRisk := "Low"
		if busFactor < 2 {
			busRisk = "Critical"
		} else if busFactor < 3 {
			busRisk = "High"
		} else if busFactor < 5 {
			busRisk = "Medium"
		}
		metrics = append(metrics, fmt.Sprintf("**Knowledge Risk:** %s (Bus factor: %d)",
			busRisk, busFactor))
	}

	b.List(metrics)

	// Critical findings
	b.Section(2, "Critical Findings")

	criticalItems := collectCriticalFindings(data)
	if len(criticalItems) > 0 {
		b.List(criticalItems)
	} else {
		b.Paragraph("No critical security findings identified.")
	}

	// Recommendations
	b.Section(2, "Recommendations")

	recommendations := generateExecutiveRecommendations(data)
	if len(recommendations.immediate) > 0 {
		b.Section(3, "Immediate Actions")
		b.NumberedList(recommendations.immediate)
	}

	if len(recommendations.strategic) > 0 {
		b.Section(3, "Strategic Improvements")
		b.NumberedList(recommendations.strategic)
	}

	// Business impact
	b.Section(2, "Business Impact")

	impact := assessBusinessImpact(data)
	b.List(impact)

	b.Footer("DevOps")

	return b.String()
}

type maturityAssessment struct {
	overall string
	summary string
}

func assessDevOpsMaturity(data *ReportData) maturityAssessment {
	score := 0
	maxScore := 0

	// Security posture (0-40 points)
	maxScore += 40
	if data.Summary.IaC != nil {
		criticalHigh := data.Summary.IaC.Critical + data.Summary.IaC.High
		if criticalHigh == 0 {
			score += 15
		} else if criticalHigh < 5 {
			score += 10
		} else if criticalHigh < 10 {
			score += 5
		}
	}

	if data.Summary.Containers != nil {
		criticalHigh := data.Summary.Containers.Critical + data.Summary.Containers.High
		if criticalHigh == 0 {
			score += 15
		} else if criticalHigh < 10 {
			score += 10
		} else if criticalHigh < 20 {
			score += 5
		}
	}

	if data.Summary.GitHubActions != nil {
		criticalHigh := data.Summary.GitHubActions.Critical + data.Summary.GitHubActions.High
		if criticalHigh == 0 {
			score += 10
		} else if criticalHigh < 3 {
			score += 5
		}
	}

	// DORA metrics (0-30 points)
	maxScore += 30
	if data.Summary.DORA != nil {
		switch data.Summary.DORA.OverallClass {
		case "elite":
			score += 30
		case "high":
			score += 20
		case "medium":
			score += 10
		}
	}

	// Team health (0-30 points)
	maxScore += 30
	if data.Summary.Git != nil {
		// Bus factor
		if data.Summary.Git.BusFactor >= 5 {
			score += 15
		} else if data.Summary.Git.BusFactor >= 3 {
			score += 10
		} else if data.Summary.Git.BusFactor >= 2 {
			score += 5
		}

		// Activity level
		switch data.Summary.Git.ActivityLevel {
		case "high":
			score += 15
		case "medium":
			score += 10
		case "low":
			score += 5
		}
	}

	percentage := (score * 100) / maxScore
	var level string
	var summary string

	switch {
	case percentage >= 80:
		level = "Excellent"
		summary = "Your DevOps practices demonstrate strong maturity with robust security, high performance, and healthy team dynamics."
	case percentage >= 60:
		level = "Good"
		summary = "Your DevOps practices are solid with some areas for improvement in security, performance, or team structure."
	case percentage >= 40:
		level = "Developing"
		summary = "Your DevOps practices show promise but require focused attention on security, performance, and team health."
	default:
		level = "Needs Improvement"
		summary = "Your DevOps practices require significant improvements across security, performance, and team structure."
	}

	return maturityAssessment{
		overall: fmt.Sprintf("%s (%d%%)", level, percentage),
		summary: summary,
	}
}

func collectCriticalFindings(data *ReportData) []string {
	var findings []string

	// IaC critical findings
	if data.Summary.IaC != nil && data.Summary.IaC.Critical > 0 {
		findings = append(findings, fmt.Sprintf("**Infrastructure:** %d critical IaC misconfigurations detected",
			data.Summary.IaC.Critical))
	}

	// Container critical vulnerabilities
	if data.Summary.Containers != nil && data.Summary.Containers.Critical > 0 {
		findings = append(findings, fmt.Sprintf("**Containers:** %d critical vulnerabilities in container images",
			data.Summary.Containers.Critical))
	}

	// CI/CD critical issues
	if data.Summary.GitHubActions != nil && data.Summary.GitHubActions.Critical > 0 {
		findings = append(findings, fmt.Sprintf("**CI/CD:** %d critical security issues in GitHub Actions workflows",
			data.Summary.GitHubActions.Critical))
	}

	// Bus factor risk
	if data.Summary.Git != nil && data.Summary.Git.BusFactor < 2 {
		findings = append(findings, fmt.Sprintf("**Team Risk:** Critical knowledge concentration (bus factor: %d)",
			data.Summary.Git.BusFactor))
	}

	return findings
}

type executiveRecommendations struct {
	immediate  []string
	strategic  []string
}

func generateExecutiveRecommendations(data *ReportData) executiveRecommendations {
	rec := executiveRecommendations{}

	// Security recommendations
	if data.Summary.IaC != nil {
		criticalHigh := data.Summary.IaC.Critical + data.Summary.IaC.High
		if criticalHigh > 10 {
			rec.immediate = append(rec.immediate, "Address critical infrastructure security misconfigurations immediately")
		} else if criticalHigh > 0 {
			rec.strategic = append(rec.strategic, "Remediate IaC security findings and implement policy-as-code")
		}
	}

	if data.Summary.Containers != nil {
		critical := data.Summary.Containers.Critical
		if critical > 10 {
			rec.immediate = append(rec.immediate, "Urgent: Update container images to address critical vulnerabilities")
		} else if critical > 0 {
			rec.immediate = append(rec.immediate, "Update container base images to address critical vulnerabilities")
		}
		high := data.Summary.Containers.High
		if high > 20 {
			rec.strategic = append(rec.strategic, "Implement automated container scanning in CI/CD pipeline")
		}
	}

	if data.Summary.GitHubActions != nil {
		criticalHigh := data.Summary.GitHubActions.Critical + data.Summary.GitHubActions.High
		if criticalHigh > 5 {
			rec.immediate = append(rec.immediate, "Harden CI/CD pipeline security controls")
		}
	}

	// DORA recommendations
	if data.Summary.DORA != nil {
		if data.Summary.DORA.OverallClass == "low" {
			rec.strategic = append(rec.strategic, "Improve deployment frequency and reduce lead times through automation")
		}
		if data.Summary.DORA.ChangeFailureRate > 0.15 {
			rec.strategic = append(rec.strategic, "Reduce change failure rate by improving testing and deployment practices")
		}
		if data.Summary.DORA.MTTRHours > 24 {
			rec.strategic = append(rec.strategic, "Improve incident response and recovery procedures to reduce MTTR")
		}
	}

	// Team health recommendations
	if data.Summary.Git != nil {
		if data.Summary.Git.BusFactor < 2 {
			rec.immediate = append(rec.immediate, "Critical: Implement knowledge sharing to reduce bus factor risk")
		} else if data.Summary.Git.BusFactor < 3 {
			rec.strategic = append(rec.strategic, "Increase code review participation and cross-training")
		}

		if data.Summary.Git.ActivityLevel == "low" {
			rec.strategic = append(rec.strategic, "Assess team capacity and development velocity")
		}
	}

	return rec
}

func assessBusinessImpact(data *ReportData) []string {
	var impact []string

	// Security risk
	totalCritical := 0
	if data.Summary.IaC != nil {
		totalCritical += data.Summary.IaC.Critical
	}
	if data.Summary.Containers != nil {
		totalCritical += data.Summary.Containers.Critical
	}
	if data.Summary.GitHubActions != nil {
		totalCritical += data.Summary.GitHubActions.Critical
	}

	if totalCritical > 20 {
		impact = append(impact, "**Security Risk:** High - Critical vulnerabilities pose significant risk to production systems")
	} else if totalCritical > 10 {
		impact = append(impact, "**Security Risk:** Medium - Several critical issues require prompt attention")
	} else if totalCritical > 0 {
		impact = append(impact, "**Security Risk:** Low - Limited critical findings identified")
	} else {
		impact = append(impact, "**Security Risk:** Minimal - No critical security findings")
	}

	// Deployment velocity
	if data.Summary.DORA != nil {
		var velocityImpact string
		switch data.Summary.DORA.DeploymentFrequencyClass {
		case "elite":
			velocityImpact = "**Deployment Velocity:** Excellent - Multiple deployments per day enable rapid feature delivery"
		case "high":
			velocityImpact = "**Deployment Velocity:** Good - Daily deployments support competitive time-to-market"
		case "medium":
			velocityImpact = "**Deployment Velocity:** Moderate - Weekly deployments may limit responsiveness to market changes"
		default:
			velocityImpact = "**Deployment Velocity:** Low - Infrequent deployments may impact competitive position"
		}
		impact = append(impact, velocityImpact)

		// Lead time impact
		if data.Summary.DORA.LeadTimeHours < 24 {
			impact = append(impact, "**Time to Market:** Excellent - Changes reach production within one day")
		} else if data.Summary.DORA.LeadTimeHours < 168 {
			impact = append(impact, "**Time to Market:** Good - Changes reach production within one week")
		} else {
			impact = append(impact, "**Time to Market:** Slow - Long lead times may delay customer value delivery")
		}
	}

	// Team sustainability
	if data.Summary.Git != nil {
		if data.Summary.Git.BusFactor < 2 {
			impact = append(impact, "**Team Sustainability:** Critical Risk - Knowledge loss could severely impact operations")
		} else if data.Summary.Git.BusFactor < 3 {
			impact = append(impact, "**Team Sustainability:** Moderate Risk - Limited knowledge distribution increases vulnerability")
		} else if data.Summary.Git.BusFactor >= 5 {
			impact = append(impact, "**Team Sustainability:** Low Risk - Healthy knowledge distribution supports continuity")
		}
	}

	return impact
}

func generateDORAInsights(dora *DORASummary) []string {
	var insights []string

	// Deployment frequency insights
	switch dora.DeploymentFrequencyClass {
	case "elite":
		insights = append(insights, "Deployment frequency is elite - your team deploys multiple times per day, enabling rapid iteration")
	case "high":
		insights = append(insights, "Deployment frequency is high - daily deployments support fast feedback cycles")
	case "medium":
		insights = append(insights, "Deployment frequency is medium - consider increasing automation to enable more frequent deployments")
	default:
		insights = append(insights, "Deployment frequency is low - invest in CI/CD automation to reduce deployment friction")
	}

	// Lead time insights
	switch dora.LeadTimeClass {
	case "elite":
		insights = append(insights, "Lead time is elite - changes reach production in less than one day")
	case "high":
		insights = append(insights, "Lead time is high - changes reach production within one week")
	case "medium":
		insights = append(insights, "Lead time is medium - streamline review and testing processes to reduce time to production")
	default:
		insights = append(insights, "Lead time is low - consider breaking down changes and improving pipeline efficiency")
	}

	// Change failure rate insights
	switch dora.ChangeFailureClass {
	case "elite":
		insights = append(insights, "Change failure rate is elite (0-15%) - excellent quality control")
	case "high":
		insights = append(insights, "Change failure rate is acceptable but could be improved through better testing")
	default:
		insights = append(insights, "Change failure rate is concerning - strengthen testing and review processes")
	}

	// MTTR insights
	switch dora.MTTRClass {
	case "elite":
		insights = append(insights, "Time to restore is elite - incidents are resolved quickly")
	case "high":
		insights = append(insights, "Time to restore is good - incident response is effective")
	default:
		insights = append(insights, "Time to restore needs improvement - enhance monitoring, alerting, and runbooks")
	}

	return insights
}

// WriteReports generates and writes both reports to the analysis directory
func WriteReports(analysisDir string) error {
	data, err := LoadReportData(analysisDir)
	if err != nil {
		return err
	}

	// Write technical report
	techReport := GenerateTechnicalReport(data)
	techPath := filepath.Join(analysisDir, "devops-technical-report.md")
	if err := os.WriteFile(techPath, []byte(techReport), 0644); err != nil {
		return fmt.Errorf("writing technical report: %w", err)
	}

	// Write executive report
	execReport := GenerateExecutiveReport(data)
	execPath := filepath.Join(analysisDir, "devops-executive-report.md")
	if err := os.WriteFile(execPath, []byte(execReport), 0644); err != nil {
		return fmt.Errorf("writing executive report: %w", err)
	}

	return nil
}

// Helper functions

func severityToInt(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}
