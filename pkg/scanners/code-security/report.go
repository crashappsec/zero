package codesecurity

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

// LoadReportData loads code-security.json from the analysis directory
func LoadReportData(analysisDir string) (*ReportData, error) {
	securityPath := filepath.Join(analysisDir, "code-security.json")
	data, err := os.ReadFile(securityPath)
	if err != nil {
		return nil, fmt.Errorf("reading code-security.json: %w", err)
	}

	var result struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing code-security.json: %w", err)
	}

	return &ReportData{
		Repository: result.Repository,
		Timestamp:  result.Timestamp,
		Summary:    result.Summary,
		Findings:   result.Findings,
	}, nil
}

// GenerateTechnicalReport creates a detailed technical report for engineers
func GenerateTechnicalReport(data *ReportData) string {
	b := report.NewBuilder()

	b.Title("Code Security Technical Report")
	b.Meta(report.ReportMeta{
		Repository:  data.Repository,
		Timestamp:   data.Timestamp,
		ScannerDesc: "SAST, Secret Detection, API Security Analysis",
	})

	// Security Overview
	b.Section(2, "Security Overview")
	b.Paragraph(generateSecurityOverview(data))

	// Vulnerability Analysis Section
	if data.Summary.Vulns != nil {
		b.Section(2, "1. Code Vulnerabilities (SAST)")
		generateVulnSection(b, data)
	}

	// Secret Detection Section
	if data.Summary.Secrets != nil {
		b.Section(2, "2. Secret Detection")
		generateSecretsSection(b, data)
	}

	// API Security Section
	if data.Summary.API != nil {
		b.Section(2, "3. API Security")
		generateAPISection(b, data)
	}

	// Errors Section
	if len(data.Summary.Errors) > 0 {
		b.Section(2, "Scan Errors")
		b.List(data.Summary.Errors)
	}

	b.Footer("Code Security")

	return b.String()
}

// GenerateExecutiveReport creates a high-level summary for security leaders
func GenerateExecutiveReport(data *ReportData) string {
	b := report.NewBuilder()

	b.Title("Code Security Executive Report")
	b.Raw(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	b.Raw(fmt.Sprintf("**Date:** %s\n\n", data.Timestamp.Format("January 2, 2006")))

	// Executive Summary
	b.Section(2, "Executive Summary")

	riskLevel := calculateOverallRiskLevel(data)
	criticalCount := countCriticalFindings(data)
	highCount := countHighFindings(data)

	b.KeyValue("Overall Risk Level", riskLevel)
	b.KeyValue("Critical Findings", fmt.Sprintf("%d", criticalCount))
	b.KeyValue("High Findings", fmt.Sprintf("%d", highCount))
	b.Newline()

	// Security Scorecard
	b.Section(2, "Security Scorecard")
	generateScorecard(b, data)

	// Key Findings
	b.Section(2, "Key Findings")
	generateKeyFindings(b, data)

	// Priority Actions
	b.Section(2, "Priority Actions")
	generatePriorityActions(b, data)

	// Risk Assessment
	b.Section(2, "Risk Assessment")
	generateRiskAssessment(b, data)

	b.Footer("Code Security")

	return b.String()
}

// generateSecurityOverview creates the overview paragraph
func generateSecurityOverview(data *ReportData) string {
	var parts []string

	if data.Summary.Vulns != nil {
		v := data.Summary.Vulns
		parts = append(parts, fmt.Sprintf("**Code Vulnerabilities:** %d findings (%d critical, %d high)",
			v.TotalFindings, v.Critical, v.High))
	}

	if data.Summary.Secrets != nil {
		s := data.Summary.Secrets
		parts = append(parts, fmt.Sprintf("**Secrets Detected:** %d findings (%d critical, %d high) - Risk Level: %s",
			s.TotalFindings, s.Critical, s.High, s.RiskLevel))
	}

	if data.Summary.API != nil {
		a := data.Summary.API
		parts = append(parts, fmt.Sprintf("**API Security:** %d findings (%d critical, %d high)",
			a.TotalFindings, a.Critical, a.High))
	}

	return strings.Join(parts, "\n\n")
}

// generateVulnSection generates the vulnerability analysis section
func generateVulnSection(b *report.ReportBuilder, data *ReportData) {
	v := data.Summary.Vulns

	b.Section(3, "Summary")

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Total Findings", fmt.Sprintf("%d", v.TotalFindings)},
		{"Critical", fmt.Sprintf("%d", v.Critical)},
		{"High", fmt.Sprintf("%d", v.High)},
		{"Medium", fmt.Sprintf("%d", v.Medium)},
		{"Low", fmt.Sprintf("%d", v.Low)},
	}
	b.Table(headers, rows)

	// By Category breakdown
	if len(v.ByCategory) > 0 {
		b.Section(3, "Findings by Category")
		var categories []string
		for cat := range v.ByCategory {
			categories = append(categories, cat)
		}
		sort.Strings(categories)

		catRows := [][]string{}
		for _, cat := range categories {
			catRows = append(catRows, []string{cat, fmt.Sprintf("%d", v.ByCategory[cat])})
		}
		b.Table([]string{"Category", "Count"}, catRows)
	}

	// By CWE breakdown
	if len(v.ByCWE) > 0 {
		b.Section(3, "Findings by CWE")
		var cwes []string
		for cwe := range v.ByCWE {
			cwes = append(cwes, cwe)
		}
		sort.Strings(cwes)

		cweRows := [][]string{}
		for _, cwe := range cwes {
			cweRows = append(cweRows, []string{cwe, fmt.Sprintf("%d", v.ByCWE[cwe])})
		}
		b.Table([]string{"CWE", "Count"}, cweRows)
	}

	// Detailed findings
	if len(data.Findings.Vulns) > 0 {
		b.Section(3, "Detailed Findings")

		// Sort by severity: critical > high > medium > low
		sortedVulns := sortBySeverity(data.Findings.Vulns)

		for i, finding := range sortedVulns {
			b.Section(4, fmt.Sprintf("%d. [%s] %s", i+1, strings.ToUpper(finding.Severity), finding.Title))

			b.KeyValue("Rule ID", finding.RuleID)
			b.KeyValue("Severity", strings.ToUpper(finding.Severity))
			b.KeyValue("File", finding.File)
			b.KeyValue("Line", fmt.Sprintf("%d", finding.Line))

			if finding.Category != "" {
				b.KeyValue("Category", finding.Category)
			}

			if len(finding.CWE) > 0 {
				b.KeyValue("CWE", strings.Join(finding.CWE, ", "))
			}

			if len(finding.OWASP) > 0 {
				b.KeyValue("OWASP", strings.Join(finding.OWASP, ", "))
			}

			b.Newline()
			b.Paragraph(b.Bold("Description:"))
			b.Paragraph(finding.Description)

			if finding.Fix != "" {
				b.Paragraph(b.Bold("Remediation:"))
				b.Paragraph(finding.Fix)
			}

			if i < len(sortedVulns)-1 {
				b.Divider()
			}
		}
	}
}

// generateSecretsSection generates the secret detection section
func generateSecretsSection(b *report.ReportBuilder, data *ReportData) {
	s := data.Summary.Secrets

	b.Section(3, "Summary")

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Total Findings", fmt.Sprintf("%d", s.TotalFindings)},
		{"Critical", fmt.Sprintf("%d", s.Critical)},
		{"High", fmt.Sprintf("%d", s.High)},
		{"Medium", fmt.Sprintf("%d", s.Medium)},
		{"Low", fmt.Sprintf("%d", s.Low)},
		{"Risk Score", fmt.Sprintf("%d", s.RiskScore)},
		{"Risk Level", strings.ToUpper(s.RiskLevel)},
		{"Files Affected", fmt.Sprintf("%d", s.FilesAffected)},
	}

	if s.ConfirmedSecrets > 0 {
		rows = append(rows, []string{"AI-Confirmed Real Secrets", fmt.Sprintf("%d", s.ConfirmedSecrets)})
	}
	if s.FalsePositives > 0 {
		rows = append(rows, []string{"AI-Identified False Positives", fmt.Sprintf("%d", s.FalsePositives)})
	}

	b.Table(headers, rows)

	// Detection sources
	if len(s.BySource) > 0 {
		b.Section(3, "Detection Sources")
		sourceRows := [][]string{}

		sourceOrder := []string{"semgrep", "entropy", "git_history"}
		for _, source := range sourceOrder {
			if count, ok := s.BySource[source]; ok {
				sourceRows = append(sourceRows, []string{strings.Title(strings.ReplaceAll(source, "_", " ")), fmt.Sprintf("%d", count)})
			}
		}

		b.Table([]string{"Source", "Count"}, sourceRows)

		if s.RemovedSecrets > 0 {
			b.Paragraph(fmt.Sprintf("**Note:** %d secrets were found in git history but later removed. These still represent a security risk and should be rotated.", s.RemovedSecrets))
		}
	}

	// By Type breakdown
	if len(s.ByType) > 0 {
		b.Section(3, "Findings by Type")
		var types []string
		for t := range s.ByType {
			types = append(types, t)
		}
		sort.Strings(types)

		typeRows := [][]string{}
		for _, t := range types {
			typeRows = append(typeRows, []string{t, fmt.Sprintf("%d", s.ByType[t])})
		}
		b.Table([]string{"Secret Type", "Count"}, typeRows)
	}

	// Detailed findings
	if len(data.Findings.Secrets) > 0 {
		b.Section(3, "Detailed Findings")

		sortedSecrets := sortSecretsBySeverity(data.Findings.Secrets)

		for i, finding := range sortedSecrets {
			severityBadge := strings.ToUpper(finding.Severity)
			if finding.IsFalsePositive != nil && *finding.IsFalsePositive {
				severityBadge = "FALSE POSITIVE"
			}

			b.Section(4, fmt.Sprintf("%d. [%s] %s", i+1, severityBadge, finding.Type))

			b.KeyValue("Severity", strings.ToUpper(finding.Severity))
			b.KeyValue("File", finding.File)
			b.KeyValue("Line", fmt.Sprintf("%d", finding.Line))

			if finding.DetectionSource != "" {
				b.KeyValue("Detection Method", strings.Title(strings.ReplaceAll(finding.DetectionSource, "_", " ")))
			}

			if finding.EntropyLevel != "" {
				b.KeyValue("Entropy", fmt.Sprintf("%s (%.2f)", finding.EntropyLevel, finding.Entropy))
			}

			b.Newline()
			b.Paragraph(finding.Message)

			// Git history context
			if finding.CommitInfo != nil {
				b.Paragraph(b.Bold("Git History Context:"))
				b.KeyValue("Commit", fmt.Sprintf("%s (%s)", finding.CommitInfo.ShortHash, finding.CommitInfo.Date))
				b.KeyValue("Author", fmt.Sprintf("%s <%s>", finding.CommitInfo.Author, finding.CommitInfo.Email))

				if finding.CommitInfo.IsRemoved {
					b.Quote("WARNING: This secret was later removed from the code but still exists in git history. It must be rotated immediately.")
				}
				b.Newline()
			}

			// AI Analysis
			if finding.AIConfidence > 0 {
				b.Paragraph(b.Bold("AI Analysis:"))
				b.KeyValue("Confidence", fmt.Sprintf("%.0f%%", finding.AIConfidence*100))
				if finding.AIReasoning != "" {
					b.Paragraph(finding.AIReasoning)
				}
				b.Newline()
			}

			// Code snippet
			if finding.Snippet != "" {
				b.Paragraph(b.Bold("Code Snippet:"))
				b.CodeBlock("", finding.Snippet)
			}

			// Rotation guidance
			if finding.Rotation != nil {
				b.Paragraph(b.Bold("Remediation:"))
				b.KeyValue("Priority", strings.ToUpper(finding.Rotation.Priority))

				if len(finding.Rotation.Steps) > 0 {
					b.Paragraph("**Rotation Steps:**")
					b.NumberedList(finding.Rotation.Steps)
				}

				if finding.Rotation.RotationURL != "" {
					b.Paragraph(fmt.Sprintf("**Rotation Guide:** %s", b.Link("Click here", finding.Rotation.RotationURL)))
				}

				if finding.Rotation.CLICommand != "" {
					b.Paragraph("**CLI Command:**")
					b.CodeBlock("bash", finding.Rotation.CLICommand)
				}

				if finding.Rotation.AutomationHint != "" {
					b.Paragraph(fmt.Sprintf("**Automation Tip:** %s", finding.Rotation.AutomationHint))
				}
			}

			if i < len(sortedSecrets)-1 {
				b.Divider()
			}
		}
	}
}

// generateAPISection generates the API security section
func generateAPISection(b *report.ReportBuilder, data *ReportData) {
	a := data.Summary.API

	b.Section(3, "Summary")

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Total Findings", fmt.Sprintf("%d", a.TotalFindings)},
		{"Critical", fmt.Sprintf("%d", a.Critical)},
		{"High", fmt.Sprintf("%d", a.High)},
		{"Medium", fmt.Sprintf("%d", a.Medium)},
		{"Low", fmt.Sprintf("%d", a.Low)},
	}
	b.Table(headers, rows)

	// By Category breakdown
	if len(a.ByCategory) > 0 {
		b.Section(3, "Findings by Category")
		var categories []string
		for cat := range a.ByCategory {
			categories = append(categories, cat)
		}
		sort.Strings(categories)

		catRows := [][]string{}
		for _, cat := range categories {
			catRows = append(catRows, []string{cat, fmt.Sprintf("%d", a.ByCategory[cat])})
		}
		b.Table([]string{"Category", "Count"}, catRows)
	}

	// Detailed findings
	if len(data.Findings.API) > 0 {
		b.Section(3, "Detailed Findings")

		sortedAPI := sortAPIBySeverity(data.Findings.API)

		for i, finding := range sortedAPI {
			b.Section(4, fmt.Sprintf("%d. [%s] %s", i+1, strings.ToUpper(finding.Severity), finding.Title))

			b.KeyValue("Rule ID", finding.RuleID)
			b.KeyValue("Severity", strings.ToUpper(finding.Severity))
			b.KeyValue("File", finding.File)
			b.KeyValue("Line", fmt.Sprintf("%d", finding.Line))
			b.KeyValue("Category", finding.Category)

			if finding.OWASPApi != "" {
				b.KeyValue("OWASP API Security", finding.OWASPApi)
			}

			b.Newline()
			b.Paragraph(finding.Description)

			if i < len(sortedAPI)-1 {
				b.Divider()
			}
		}
	}
}

// generateScorecard creates the security scorecard table
func generateScorecard(b *report.ReportBuilder, data *ReportData) {
	headers := []string{"Category", "Findings", "Critical", "High", "Status"}
	rows := [][]string{}

	if data.Summary.Vulns != nil {
		v := data.Summary.Vulns
		status := getSecurityStatus(v.Critical, v.High, v.TotalFindings)
		rows = append(rows, []string{
			"Code Vulnerabilities",
			fmt.Sprintf("%d", v.TotalFindings),
			fmt.Sprintf("%d", v.Critical),
			fmt.Sprintf("%d", v.High),
			status,
		})
	}

	if data.Summary.Secrets != nil {
		s := data.Summary.Secrets
		status := getSecurityStatus(s.Critical, s.High, s.TotalFindings)
		rows = append(rows, []string{
			"Secret Detection",
			fmt.Sprintf("%d", s.TotalFindings),
			fmt.Sprintf("%d", s.Critical),
			fmt.Sprintf("%d", s.High),
			status,
		})
	}

	if data.Summary.API != nil {
		a := data.Summary.API
		status := getSecurityStatus(a.Critical, a.High, a.TotalFindings)
		rows = append(rows, []string{
			"API Security",
			fmt.Sprintf("%d", a.TotalFindings),
			fmt.Sprintf("%d", a.Critical),
			fmt.Sprintf("%d", a.High),
			status,
		})
	}

	b.Table(headers, rows)
}

// generateKeyFindings generates the key findings section
func generateKeyFindings(b *report.ReportBuilder, data *ReportData) {
	critical := []string{}
	high := []string{}
	notable := []string{}

	// Collect critical findings
	if data.Summary.Vulns != nil && data.Summary.Vulns.Critical > 0 {
		critical = append(critical, fmt.Sprintf("%d critical code vulnerabilities detected", data.Summary.Vulns.Critical))
	}
	if data.Summary.Secrets != nil && data.Summary.Secrets.Critical > 0 {
		critical = append(critical, fmt.Sprintf("%d critical secrets detected (including %d AI-confirmed)",
			data.Summary.Secrets.Critical, data.Summary.Secrets.ConfirmedSecrets))
	}
	if data.Summary.API != nil && data.Summary.API.Critical > 0 {
		critical = append(critical, fmt.Sprintf("%d critical API security issues", data.Summary.API.Critical))
	}

	// Collect high findings
	if data.Summary.Vulns != nil && data.Summary.Vulns.High > 0 {
		high = append(high, fmt.Sprintf("%d high-severity code vulnerabilities", data.Summary.Vulns.High))
	}
	if data.Summary.Secrets != nil && data.Summary.Secrets.High > 0 {
		high = append(high, fmt.Sprintf("%d high-severity secrets", data.Summary.Secrets.High))
	}
	if data.Summary.API != nil && data.Summary.API.High > 0 {
		high = append(high, fmt.Sprintf("%d high-severity API issues", data.Summary.API.High))
	}

	// Notable findings
	if data.Summary.Secrets != nil && data.Summary.Secrets.RemovedSecrets > 0 {
		notable = append(notable, fmt.Sprintf("%d secrets found in git history (already removed from code but still need rotation)",
			data.Summary.Secrets.RemovedSecrets))
	}
	if data.Summary.Secrets != nil && data.Summary.Secrets.FilesAffected > 10 {
		notable = append(notable, fmt.Sprintf("Secrets detected in %d files - possible systemic issue with secret management",
			data.Summary.Secrets.FilesAffected))
	}

	if len(critical) > 0 {
		b.Section(3, "Critical Issues")
		b.List(critical)
	}

	if len(high) > 0 {
		b.Section(3, "High-Severity Issues")
		b.List(high)
	}

	if len(notable) > 0 {
		b.Section(3, "Notable Findings")
		b.List(notable)
	}

	if len(critical) == 0 && len(high) == 0 {
		b.Paragraph("No critical or high-severity security issues detected.")
	}
}

// generatePriorityActions generates priority action items
func generatePriorityActions(b *report.ReportBuilder, data *ReportData) {
	immediate := []string{}
	shortTerm := []string{}

	// Immediate actions (critical)
	if data.Summary.Secrets != nil && (data.Summary.Secrets.Critical > 0 || data.Summary.Secrets.RemovedSecrets > 0) {
		immediate = append(immediate, fmt.Sprintf("Rotate all exposed secrets immediately (%d critical, %d in git history)",
			data.Summary.Secrets.Critical, data.Summary.Secrets.RemovedSecrets))
	}
	if data.Summary.Vulns != nil && data.Summary.Vulns.Critical > 0 {
		immediate = append(immediate, fmt.Sprintf("Remediate %d critical code vulnerabilities", data.Summary.Vulns.Critical))
	}
	if data.Summary.API != nil && data.Summary.API.Critical > 0 {
		immediate = append(immediate, fmt.Sprintf("Fix %d critical API security issues", data.Summary.API.Critical))
	}

	// Short-term actions (high)
	if data.Summary.Vulns != nil && data.Summary.Vulns.High > 0 {
		shortTerm = append(shortTerm, fmt.Sprintf("Address %d high-severity code vulnerabilities", data.Summary.Vulns.High))
	}
	if data.Summary.Secrets != nil && data.Summary.Secrets.High > 0 {
		shortTerm = append(shortTerm, fmt.Sprintf("Rotate %d high-severity secrets", data.Summary.Secrets.High))
	}
	if data.Summary.API != nil && data.Summary.API.High > 0 {
		shortTerm = append(shortTerm, fmt.Sprintf("Remediate %d high-severity API issues", data.Summary.API.High))
	}

	// Systemic improvements
	if data.Summary.Secrets != nil && data.Summary.Secrets.TotalFindings > 5 {
		shortTerm = append(shortTerm, "Implement centralized secret management (e.g., AWS Secrets Manager, HashiCorp Vault)")
	}
	if data.Summary.Vulns != nil && data.Summary.Vulns.TotalFindings > 10 {
		shortTerm = append(shortTerm, "Integrate SAST scanning into CI/CD pipeline to catch issues earlier")
	}

	if len(immediate) > 0 {
		b.Section(3, "Immediate Actions (24-48 hours)")
		b.NumberedList(immediate)
	}

	if len(shortTerm) > 0 {
		b.Section(3, "Short-term Actions (1-2 weeks)")
		b.NumberedList(shortTerm)
	}

	if len(immediate) == 0 && len(shortTerm) == 0 {
		b.Paragraph("No immediate security actions required. Continue monitoring and maintaining secure coding practices.")
	}
}

// generateRiskAssessment generates the risk assessment section
func generateRiskAssessment(b *report.ReportBuilder, data *ReportData) {
	criticalCount := countCriticalFindings(data)
	highCount := countHighFindings(data)

	var riskLevel, impact, recommendation string

	if criticalCount > 0 {
		riskLevel = "CRITICAL"
		impact = "Immediate security risk. Active exploitation could lead to data breach, unauthorized access, or service compromise."
		recommendation = "Initiate incident response procedures. Rotate all exposed credentials. Deploy critical patches immediately."
	} else if highCount > 5 {
		riskLevel = "HIGH"
		impact = "Significant security vulnerabilities present. Exploitation could compromise application security or data integrity."
		recommendation = "Prioritize remediation of high-severity findings within 1-2 weeks. Implement compensating controls where immediate fixes are not possible."
	} else if highCount > 0 {
		riskLevel = "MODERATE"
		impact = "Some security issues identified. While not immediately critical, these should be addressed to maintain security posture."
		recommendation = "Schedule remediation work in upcoming sprints. Review and update security practices."
	} else {
		riskLevel = "LOW"
		impact = "No critical security issues detected. Minor findings may exist but pose minimal immediate risk."
		recommendation = "Continue current security practices. Address medium and low findings during regular maintenance cycles."
	}

	b.KeyValue("Overall Risk Level", riskLevel)
	b.KeyValue("Business Impact", impact)
	b.Newline()
	b.Paragraph(b.Bold("Recommendation:"))
	b.Paragraph(recommendation)

	// Additional context
	if data.Summary.Secrets != nil && data.Summary.Secrets.RemovedSecrets > 0 {
		b.Newline()
		b.Quote(fmt.Sprintf("IMPORTANT: %d secrets found in git history represent ongoing security risk even though they've been removed from the current codebase. All such secrets must be rotated immediately.", data.Summary.Secrets.RemovedSecrets))
	}
}

// Helper functions

func sortBySeverity(vulns []VulnFinding) []VulnFinding {
	sorted := make([]VulnFinding, len(vulns))
	copy(sorted, vulns)

	severityRank := map[string]int{
		"critical": 0,
		"high":     1,
		"medium":   2,
		"low":      3,
	}

	sort.Slice(sorted, func(i, j int) bool {
		rankI := severityRank[strings.ToLower(sorted[i].Severity)]
		rankJ := severityRank[strings.ToLower(sorted[j].Severity)]
		return rankI < rankJ
	})

	return sorted
}

func sortSecretsBySeverity(secrets []SecretFinding) []SecretFinding {
	sorted := make([]SecretFinding, len(secrets))
	copy(sorted, secrets)

	severityRank := map[string]int{
		"critical": 0,
		"high":     1,
		"medium":   2,
		"low":      3,
	}

	sort.Slice(sorted, func(i, j int) bool {
		// False positives go to the end
		if sorted[i].IsFalsePositive != nil && *sorted[i].IsFalsePositive {
			return false
		}
		if sorted[j].IsFalsePositive != nil && *sorted[j].IsFalsePositive {
			return true
		}

		rankI := severityRank[strings.ToLower(sorted[i].Severity)]
		rankJ := severityRank[strings.ToLower(sorted[j].Severity)]
		return rankI < rankJ
	})

	return sorted
}

func sortAPIBySeverity(findings []APIFinding) []APIFinding {
	sorted := make([]APIFinding, len(findings))
	copy(sorted, findings)

	severityRank := map[string]int{
		"critical": 0,
		"high":     1,
		"medium":   2,
		"low":      3,
	}

	sort.Slice(sorted, func(i, j int) bool {
		rankI := severityRank[strings.ToLower(sorted[i].Severity)]
		rankJ := severityRank[strings.ToLower(sorted[j].Severity)]
		return rankI < rankJ
	})

	return sorted
}

func calculateOverallRiskLevel(data *ReportData) string {
	criticalCount := countCriticalFindings(data)
	highCount := countHighFindings(data)

	if criticalCount > 0 {
		return "CRITICAL"
	} else if highCount > 5 {
		return "HIGH"
	} else if highCount > 0 {
		return "MODERATE"
	}
	return "LOW"
}

func countCriticalFindings(data *ReportData) int {
	count := 0
	if data.Summary.Vulns != nil {
		count += data.Summary.Vulns.Critical
	}
	if data.Summary.Secrets != nil {
		count += data.Summary.Secrets.Critical
	}
	if data.Summary.API != nil {
		count += data.Summary.API.Critical
	}
	return count
}

func countHighFindings(data *ReportData) int {
	count := 0
	if data.Summary.Vulns != nil {
		count += data.Summary.Vulns.High
	}
	if data.Summary.Secrets != nil {
		count += data.Summary.Secrets.High
	}
	if data.Summary.API != nil {
		count += data.Summary.API.High
	}
	return count
}

func getSecurityStatus(critical, high, total int) string {
	if critical > 0 {
		return "Critical Risk"
	} else if high > 5 {
		return "High Risk"
	} else if high > 0 {
		return "Moderate Risk"
	} else if total > 0 {
		return "Low Risk"
	}
	return "Good"
}

// WriteReports generates and writes both reports to the analysis directory
func WriteReports(analysisDir string) error {
	data, err := LoadReportData(analysisDir)
	if err != nil {
		return err
	}

	// Write technical report
	techReport := GenerateTechnicalReport(data)
	techPath := filepath.Join(analysisDir, "code-security-technical-report.md")
	if err := os.WriteFile(techPath, []byte(techReport), 0644); err != nil {
		return fmt.Errorf("writing technical report: %w", err)
	}

	// Write executive report
	execReport := GenerateExecutiveReport(data)
	execPath := filepath.Join(analysisDir, "code-security-executive-report.md")
	if err := os.WriteFile(execPath, []byte(execReport), 0644); err != nil {
		return fmt.Errorf("writing executive report: %w", err)
	}

	return nil
}
