package crypto

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// LoadReportData loads crypto.json from the analysis directory
func LoadReportData(analysisDir string) (*ReportData, error) {
	cryptoPath := filepath.Join(analysisDir, "crypto.json")
	data, err := os.ReadFile(cryptoPath)
	if err != nil {
		return nil, fmt.Errorf("reading crypto.json: %w", err)
	}

	var result struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing crypto.json: %w", err)
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

	// Title and metadata
	b.Title("Cryptographic Security Analysis - Technical Report")
	b.Meta(report.ReportMeta{
		Repository:  data.Repository,
		Timestamp:   data.Timestamp,
		ScannerDesc: "Crypto Scanner - Analyzes weak ciphers, hardcoded keys, insecure random, TLS config, and certificates",
	})

	// Executive Summary
	b.Section(2, "Executive Summary")
	totalFindings := countTotalFindings(data)
	criticalCount := countBySeverity(data, "critical")
	highCount := countBySeverity(data, "high")

	b.Paragraph(fmt.Sprintf("This scan identified **%d total cryptographic security findings** across the repository. "+
		"Of these, **%d are critical** and **%d are high severity**. Immediate action is recommended for all critical and high severity findings.",
		totalFindings, criticalCount, highCount))

	// Overall Risk Assessment
	riskLevel := assessOverallRisk(criticalCount, highCount, totalFindings)
	b.Section(3, "Overall Cryptographic Security Risk")
	b.KeyValue("Risk Level", riskLevel)
	b.KeyValue("Total Findings", fmt.Sprintf("%d", totalFindings))
	b.KeyValue("Critical Severity", fmt.Sprintf("%d", criticalCount))
	b.KeyValue("High Severity", fmt.Sprintf("%d", highCount))
	b.Newline()

	// Cipher Analysis
	if data.Summary.Ciphers != nil {
		b.Section(2, "1. Weak Cipher Analysis")
		c := data.Summary.Ciphers

		b.Section(3, "Summary")
		rows := [][]string{
			{"Total Findings", fmt.Sprintf("%d", c.TotalFindings)},
			{"Detection Method", detectionMethod(c.UsedSemgrep)},
		}
		b.Table([]string{"Metric", "Value"}, rows)

		if len(c.BySeverity) > 0 {
			b.Section(4, "Findings by Severity")
			var sevRows [][]string
			for sev, count := range c.BySeverity {
				sevRows = append(sevRows, []string{severityBadge(sev), fmt.Sprintf("%d", count)})
			}
			b.Table([]string{"Severity", "Count"}, sevRows)
		}

		if len(c.ByAlgorithm) > 0 {
			b.Section(4, "Findings by Algorithm")
			var algoRows [][]string
			for algo, count := range c.ByAlgorithm {
				algoRows = append(algoRows, []string{algo, fmt.Sprintf("%d", count)})
			}
			b.Table([]string{"Algorithm", "Count"}, algoRows)
		}

		// Detailed cipher findings
		if len(data.Findings.Ciphers) > 0 {
			b.Section(3, "Detailed Cipher Findings")
			criticalCiphers := filterBySeverity(data.Findings.Ciphers, "critical")
			highCiphers := filterBySeverity(data.Findings.Ciphers, "high")

			if len(criticalCiphers) > 0 {
				b.Section(4, "Critical Severity")
				for _, f := range criticalCiphers {
					b.Quote(fmt.Sprintf("**%s** - %s\n\n"+
						"File: `%s:%d`\n\n"+
						"Match: `%s`\n\n"+
						"Recommendation: %s\n\n"+
						"CWE: %s",
						f.Algorithm, f.Description, f.File, f.Line, f.Match, f.Suggestion, f.CWE))
				}
			}

			if len(highCiphers) > 0 {
				b.Section(4, "High Severity")
				for _, f := range highCiphers {
					b.Quote(fmt.Sprintf("**%s** - %s\n\n"+
						"File: `%s:%d`\n\n"+
						"Match: `%s`\n\n"+
						"Recommendation: %s\n\n"+
						"CWE: %s",
						f.Algorithm, f.Description, f.File, f.Line, f.Match, f.Suggestion, f.CWE))
				}
			}
		}
	}

	// Key Management Analysis
	if data.Summary.Keys != nil {
		b.Section(2, "2. Key Management Analysis")
		k := data.Summary.Keys

		b.Section(3, "Summary")
		rows := [][]string{
			{"Total Findings", fmt.Sprintf("%d", k.TotalFindings)},
		}
		b.Table([]string{"Metric", "Value"}, rows)

		if len(k.BySeverity) > 0 {
			b.Section(4, "Findings by Severity")
			var sevRows [][]string
			for sev, count := range k.BySeverity {
				sevRows = append(sevRows, []string{severityBadge(sev), fmt.Sprintf("%d", count)})
			}
			b.Table([]string{"Severity", "Count"}, sevRows)
		}

		if len(k.ByType) > 0 {
			b.Section(4, "Findings by Key Type")
			var typeRows [][]string
			for keyType, count := range k.ByType {
				typeRows = append(typeRows, []string{keyType, fmt.Sprintf("%d", count)})
			}
			b.Table([]string{"Key Type", "Count"}, typeRows)
		}

		// Detailed key findings
		if len(data.Findings.Keys) > 0 {
			b.Section(3, "Detailed Key Findings")
			criticalKeys := filterKeysBySeverity(data.Findings.Keys, "critical")

			if len(criticalKeys) > 0 {
				b.Section(4, "Critical Hardcoded Keys")
				b.Paragraph("The following hardcoded cryptographic keys were detected. These should be moved to secure secret management immediately.")
				for _, f := range criticalKeys {
					b.Quote(fmt.Sprintf("**%s** - %s\n\n"+
						"File: `%s:%d`\n\n"+
						"Match: `%s`\n\n"+
						"CWE: %s",
						f.Type, f.Description, f.File, f.Line, f.Match, f.CWE))
				}
			}
		}
	}

	// Random Number Generation Analysis
	if data.Summary.Random != nil {
		b.Section(2, "3. Random Number Generation Analysis")
		r := data.Summary.Random

		b.Section(3, "Summary")
		rows := [][]string{
			{"Total Findings", fmt.Sprintf("%d", r.TotalFindings)},
		}
		b.Table([]string{"Metric", "Value"}, rows)

		if len(r.BySeverity) > 0 {
			b.Section(4, "Findings by Severity")
			var sevRows [][]string
			for sev, count := range r.BySeverity {
				sevRows = append(sevRows, []string{severityBadge(sev), fmt.Sprintf("%d", count)})
			}
			b.Table([]string{"Severity", "Count"}, sevRows)
		}

		// Detailed random findings
		if len(data.Findings.Random) > 0 {
			b.Section(3, "Detailed Random Generation Findings")
			for _, f := range data.Findings.Random {
				b.Quote(fmt.Sprintf("**%s** - %s\n\n"+
					"File: `%s:%d`\n\n"+
					"Match: `%s`\n\n"+
					"Recommendation: %s\n\n"+
					"CWE: %s",
					f.Type, f.Description, f.File, f.Line, f.Match, f.Suggestion, f.CWE))
			}
		}
	}

	// TLS Configuration Analysis
	if data.Summary.TLS != nil {
		b.Section(2, "4. TLS Configuration Analysis")
		t := data.Summary.TLS

		b.Section(3, "Summary")
		rows := [][]string{
			{"Total Findings", fmt.Sprintf("%d", t.TotalFindings)},
		}
		b.Table([]string{"Metric", "Value"}, rows)

		if len(t.BySeverity) > 0 {
			b.Section(4, "Findings by Severity")
			var sevRows [][]string
			for sev, count := range t.BySeverity {
				sevRows = append(sevRows, []string{severityBadge(sev), fmt.Sprintf("%d", count)})
			}
			b.Table([]string{"Severity", "Count"}, sevRows)
		}

		// Detailed TLS findings
		if len(data.Findings.TLS) > 0 {
			b.Section(3, "Detailed TLS Findings")
			criticalTLS := filterTLSBySeverity(data.Findings.TLS, "critical")

			if len(criticalTLS) > 0 {
				b.Section(4, "Critical TLS Issues")
				b.Paragraph("The following critical TLS misconfigurations were detected. These should be addressed immediately.")
				for _, f := range criticalTLS {
					b.Quote(fmt.Sprintf("**%s** - %s\n\n"+
						"File: `%s:%d`\n\n"+
						"Match: `%s`\n\n"+
						"Recommendation: %s\n\n"+
						"CWE: %s",
						f.Type, f.Description, f.File, f.Line, f.Match, f.Suggestion, f.CWE))
				}
			}
		}
	}

	// Certificate Analysis
	if data.Summary.Certificates != nil {
		b.Section(2, "5. X.509 Certificate Analysis")
		cert := data.Summary.Certificates

		b.Section(3, "Summary")
		rows := [][]string{
			{"Total Certificates", fmt.Sprintf("%d", cert.TotalCertificates)},
			{"Total Findings", fmt.Sprintf("%d", cert.TotalFindings)},
			{"Expired", fmt.Sprintf("%d", cert.Expired)},
			{"Expiring Soon", fmt.Sprintf("%d", cert.ExpiringSoon)},
			{"Weak Keys", fmt.Sprintf("%d", cert.WeakKey)},
		}
		b.Table([]string{"Metric", "Value"}, rows)

		// Certificate findings
		if data.Findings.Certificates != nil {
			if len(data.Findings.Certificates.Certificates) > 0 {
				b.Section(3, "Certificate Inventory")
				var certRows [][]string
				for _, c := range data.Findings.Certificates.Certificates {
					certRows = append(certRows, []string{
						c.File,
						c.Subject,
						c.KeyType,
						fmt.Sprintf("%d", c.KeySize),
						fmt.Sprintf("%d days", c.DaysUntilExp),
						fmt.Sprintf("%t", c.IsSelfSigned),
					})
				}
				b.Table([]string{"File", "Subject", "Key Type", "Key Size", "Expires In", "Self-Signed"}, certRows)
			}

			if len(data.Findings.Certificates.Findings) > 0 {
				b.Section(3, "Certificate Issues")
				for _, f := range data.Findings.Certificates.Findings {
					suggestion := ""
					if f.Suggestion != "" {
						suggestion = fmt.Sprintf("\n\nRecommendation: %s", f.Suggestion)
					}
					b.Quote(fmt.Sprintf("**[%s]** %s - %s\n\n"+
						"File: `%s`%s",
						f.Severity, f.Type, f.Description, f.File, suggestion))
				}
			}
		}
	}

	// Remediation Priorities
	b.Section(2, "Remediation Priorities")
	priorities := generateRemediationPriorities(data)
	if len(priorities.immediate) > 0 {
		b.Section(3, "Immediate Action Required")
		b.NumberedList(priorities.immediate)
	}
	if len(priorities.shortTerm) > 0 {
		b.Section(3, "Short-term Actions")
		b.NumberedList(priorities.shortTerm)
	}
	if len(priorities.longTerm) > 0 {
		b.Section(3, "Long-term Improvements")
		b.NumberedList(priorities.longTerm)
	}

	b.Footer("Crypto")
	return b.String()
}

// GenerateExecutiveReport creates a high-level summary for leadership
func GenerateExecutiveReport(data *ReportData) string {
	b := report.NewBuilder()

	// Title and metadata
	b.Title("Cryptographic Security Assessment - Executive Summary")
	b.Paragraph(fmt.Sprintf("**Repository:** `%s`", data.Repository))
	b.Paragraph(fmt.Sprintf("**Date:** %s", data.Timestamp.Format("January 2, 2006")))
	b.Divider()

	// Overall Risk Assessment
	totalFindings := countTotalFindings(data)
	criticalCount := countBySeverity(data, "critical")
	highCount := countBySeverity(data, "high")
	riskLevel := assessOverallRisk(criticalCount, highCount, totalFindings)

	b.Section(2, "Overall Security Posture")
	b.Paragraph(fmt.Sprintf("**Cryptographic Security Risk Level:** %s", riskLevel))
	b.Paragraph(fmt.Sprintf("This assessment analyzed the cryptographic security of `%s` across five critical areas: "+
		"weak ciphers, key management, random number generation, TLS configuration, and X.509 certificates.",
		data.Repository))

	// Risk Summary Table
	b.Section(3, "Risk Summary")
	rows := [][]string{
		{"Critical Findings", fmt.Sprintf("%d", criticalCount)},
		{"High Severity Findings", fmt.Sprintf("%d", highCount)},
		{"Total Issues Identified", fmt.Sprintf("%d", totalFindings)},
		{"Overall Risk Level", riskLevel},
	}
	b.Table([]string{"Metric", "Value"}, rows)

	// Key Findings by Category
	b.Section(2, "Key Findings by Category")

	if data.Summary.Ciphers != nil && data.Summary.Ciphers.TotalFindings > 0 {
		b.Section(3, "Weak Cryptographic Algorithms")
		criticalCiphers := countCiphersBySeverity(data.Findings.Ciphers, "critical")
		highCiphers := countCiphersBySeverity(data.Findings.Ciphers, "high")

		if criticalCiphers > 0 || highCiphers > 0 {
			b.Paragraph(fmt.Sprintf("**Risk:** High - Detected %d critical and %d high severity weak cipher findings. "+
				"Use of deprecated algorithms like MD5, SHA-1, DES, or RC4 can lead to data breaches.",
				criticalCiphers, highCiphers))
		} else {
			b.Paragraph(fmt.Sprintf("**Risk:** Low - Found %d low/medium severity cipher findings.", data.Summary.Ciphers.TotalFindings))
		}
	}

	if data.Summary.Keys != nil && data.Summary.Keys.TotalFindings > 0 {
		b.Section(3, "Hardcoded Cryptographic Keys")
		criticalKeys := countKeysBySeverity(data.Findings.Keys, "critical")

		b.Paragraph(fmt.Sprintf("**Risk:** Critical - Identified %d hardcoded cryptographic keys in source code. "+
			"These represent an immediate security risk and should be moved to secure secret management.",
			criticalKeys))
	}

	if data.Summary.Random != nil && data.Summary.Random.TotalFindings > 0 {
		b.Section(3, "Insecure Random Number Generation")
		b.Paragraph(fmt.Sprintf("**Risk:** High - Found %d instances of cryptographically insecure random number generation. "+
			"This can lead to predictable tokens, session IDs, or encryption keys.",
			data.Summary.Random.TotalFindings))
	}

	if data.Summary.TLS != nil && data.Summary.TLS.TotalFindings > 0 {
		b.Section(3, "TLS Configuration Issues")
		criticalTLS := countTLSBySeverity(data.Findings.TLS, "critical")

		if criticalTLS > 0 {
			b.Paragraph(fmt.Sprintf("**Risk:** Critical - Detected %d critical TLS misconfigurations including "+
				"disabled certificate verification or deprecated protocols. This exposes data in transit.",
				criticalTLS))
		} else {
			b.Paragraph(fmt.Sprintf("**Risk:** Medium - Found %d TLS configuration issues that should be addressed.",
				data.Summary.TLS.TotalFindings))
		}
	}

	if data.Summary.Certificates != nil && data.Summary.Certificates.TotalCertificates > 0 {
		b.Section(3, "X.509 Certificates")
		if data.Summary.Certificates.Expired > 0 {
			b.Paragraph(fmt.Sprintf("**Risk:** Critical - %d expired certificates detected. "+
				"Expired certificates will cause service disruptions.", data.Summary.Certificates.Expired))
		} else if data.Summary.Certificates.ExpiringSoon > 0 {
			b.Paragraph(fmt.Sprintf("**Risk:** Medium - %d certificates expiring soon. "+
				"Plan for renewal to avoid service disruptions.", data.Summary.Certificates.ExpiringSoon))
		} else {
			b.Paragraph(fmt.Sprintf("**Status:** Good - All %d certificates are valid and properly configured.",
				data.Summary.Certificates.TotalCertificates))
		}
	}

	// Business Impact
	b.Section(2, "Business Impact")
	impact := assessBusinessImpact(criticalCount, highCount, data)
	b.List(impact)

	// Immediate Recommendations
	b.Section(2, "Immediate Recommendations")
	recommendations := generateExecutiveRecommendations(data)
	b.NumberedList(recommendations)

	b.Divider()
	b.Paragraph("*Generated by Zero Crypto Scanner*")

	return b.String()
}

// WriteReports generates and writes both reports to the analysis directory
func WriteReports(analysisDir string) error {
	data, err := LoadReportData(analysisDir)
	if err != nil {
		return err
	}

	// Write technical report
	techReport := GenerateTechnicalReport(data)
	techPath := filepath.Join(analysisDir, "crypto-technical-report.md")
	if err := os.WriteFile(techPath, []byte(techReport), 0644); err != nil {
		return fmt.Errorf("writing technical report: %w", err)
	}

	// Write executive report
	execReport := GenerateExecutiveReport(data)
	execPath := filepath.Join(analysisDir, "crypto-executive-report.md")
	if err := os.WriteFile(execPath, []byte(execReport), 0644); err != nil {
		return fmt.Errorf("writing executive report: %w", err)
	}

	return nil
}

// Helper functions

func countTotalFindings(data *ReportData) int {
	total := 0
	if data.Summary.Ciphers != nil {
		total += data.Summary.Ciphers.TotalFindings
	}
	if data.Summary.Keys != nil {
		total += data.Summary.Keys.TotalFindings
	}
	if data.Summary.Random != nil {
		total += data.Summary.Random.TotalFindings
	}
	if data.Summary.TLS != nil {
		total += data.Summary.TLS.TotalFindings
	}
	if data.Summary.Certificates != nil {
		total += data.Summary.Certificates.TotalFindings
	}
	return total
}

func countBySeverity(data *ReportData, severity string) int {
	count := 0
	if data.Summary.Ciphers != nil {
		count += data.Summary.Ciphers.BySeverity[severity]
	}
	if data.Summary.Keys != nil {
		count += data.Summary.Keys.BySeverity[severity]
	}
	if data.Summary.Random != nil {
		count += data.Summary.Random.BySeverity[severity]
	}
	if data.Summary.TLS != nil {
		count += data.Summary.TLS.BySeverity[severity]
	}
	if data.Summary.Certificates != nil {
		count += data.Summary.Certificates.BySeverity[severity]
	}
	return count
}

func assessOverallRisk(critical, high, total int) string {
	if critical > 5 {
		return "CRITICAL"
	}
	if critical > 0 || high > 10 {
		return "HIGH"
	}
	if high > 0 || total > 20 {
		return "MEDIUM"
	}
	if total > 0 {
		return "LOW"
	}
	return "MINIMAL"
}

func detectionMethod(usedSemgrep bool) string {
	if usedSemgrep {
		return "Semgrep (AST-based) + Pattern Matching"
	}
	return "Pattern Matching"
}

func severityBadge(severity string) string {
	switch severity {
	case "critical":
		return "CRITICAL"
	case "high":
		return "HIGH"
	case "medium":
		return "MEDIUM"
	case "low":
		return "LOW"
	default:
		return severity
	}
}

func filterBySeverity(findings []CipherFinding, severity string) []CipherFinding {
	var filtered []CipherFinding
	for _, f := range findings {
		if f.Severity == severity {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func countCiphersBySeverity(findings []CipherFinding, severity string) int {
	count := 0
	for _, f := range findings {
		if f.Severity == severity {
			count++
		}
	}
	return count
}

func filterKeysBySeverity(findings []KeyFinding, severity string) []KeyFinding {
	var filtered []KeyFinding
	for _, f := range findings {
		if f.Severity == severity {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func countKeysBySeverity(findings []KeyFinding, severity string) int {
	count := 0
	for _, f := range findings {
		if f.Severity == severity {
			count++
		}
	}
	return count
}

func filterTLSBySeverity(findings []TLSFinding, severity string) []TLSFinding {
	var filtered []TLSFinding
	for _, f := range findings {
		if f.Severity == severity {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func countTLSBySeverity(findings []TLSFinding, severity string) int {
	count := 0
	for _, f := range findings {
		if f.Severity == severity {
			count++
		}
	}
	return count
}

type remediationPriorities struct {
	immediate []string
	shortTerm []string
	longTerm  []string
}

func generateRemediationPriorities(data *ReportData) remediationPriorities {
	priorities := remediationPriorities{}

	// Critical and high severity get immediate priority
	criticalCount := countBySeverity(data, "critical")
	highCount := countBySeverity(data, "high")

	if criticalCount > 0 {
		priorities.immediate = append(priorities.immediate,
			fmt.Sprintf("Address %d critical cryptographic security findings immediately", criticalCount))
	}

	// Hardcoded keys are always immediate
	if data.Summary.Keys != nil && data.Summary.Keys.TotalFindings > 0 {
		criticalKeys := countKeysBySeverity(data.Findings.Keys, "critical")
		if criticalKeys > 0 {
			priorities.immediate = append(priorities.immediate,
				fmt.Sprintf("Move %d hardcoded cryptographic keys to secure secret management (HashiCorp Vault, AWS Secrets Manager, etc.)", criticalKeys))
		}
	}

	// Disabled TLS verification is critical
	if data.Summary.TLS != nil {
		for _, f := range data.Findings.TLS {
			if f.Type == "disabled-verification" && f.Severity == "critical" {
				priorities.immediate = append(priorities.immediate,
					"Re-enable TLS certificate verification - disabled verification exposes the application to MITM attacks")
				break
			}
		}
	}

	// Expired certificates are immediate
	if data.Summary.Certificates != nil && data.Summary.Certificates.Expired > 0 {
		priorities.immediate = append(priorities.immediate,
			fmt.Sprintf("Replace %d expired certificates to avoid service disruptions", data.Summary.Certificates.Expired))
	}

	// High severity items are short-term
	if highCount > 0 {
		priorities.shortTerm = append(priorities.shortTerm,
			fmt.Sprintf("Remediate %d high severity cryptographic issues", highCount))
	}

	// Weak ciphers
	if data.Summary.Ciphers != nil && data.Summary.Ciphers.TotalFindings > 0 {
		priorities.shortTerm = append(priorities.shortTerm,
			"Replace deprecated cryptographic algorithms (MD5, SHA-1, DES, RC4) with modern alternatives (SHA-256, AES-256-GCM)")
	}

	// Insecure random
	if data.Summary.Random != nil && data.Summary.Random.TotalFindings > 0 {
		priorities.shortTerm = append(priorities.shortTerm,
			"Replace insecure random number generation with cryptographically secure alternatives (crypto/rand, secrets module, etc.)")
	}

	// Certificate expiry
	if data.Summary.Certificates != nil && data.Summary.Certificates.ExpiringSoon > 0 {
		priorities.shortTerm = append(priorities.shortTerm,
			fmt.Sprintf("Renew %d certificates expiring within 90 days", data.Summary.Certificates.ExpiringSoon))
	}

	// Long-term improvements
	priorities.longTerm = append(priorities.longTerm,
		"Implement automated certificate rotation and monitoring",
		"Establish cryptographic standards and review process",
		"Integrate SAST tools into CI/CD pipeline for continuous cryptographic analysis",
		"Conduct security training on secure cryptographic practices")

	return priorities
}

func assessBusinessImpact(critical, high int, data *ReportData) []string {
	var impacts []string

	if critical > 0 {
		impacts = append(impacts, "**Critical Risk:** Cryptographic vulnerabilities could lead to data breaches, unauthorized access, or compliance violations")
	}

	if data.Summary.Keys != nil && data.Summary.Keys.TotalFindings > 0 {
		impacts = append(impacts, "**Credential Exposure:** Hardcoded keys in source code are accessible to anyone with repository access")
	}

	if data.Summary.TLS != nil {
		for _, f := range data.Findings.TLS {
			if f.Type == "disabled-verification" {
				impacts = append(impacts, "**Man-in-the-Middle Risk:** Disabled certificate verification exposes data in transit to interception")
				break
			}
		}
	}

	if data.Summary.Certificates != nil && data.Summary.Certificates.Expired > 0 {
		impacts = append(impacts, "**Service Disruption:** Expired certificates will cause authentication failures and service outages")
	}

	if data.Summary.Random != nil && data.Summary.Random.TotalFindings > 0 {
		impacts = append(impacts, "**Predictability Risk:** Weak random generation can lead to predictable session tokens, API keys, or encryption")
	}

	if len(impacts) == 0 {
		impacts = append(impacts, "**Low Risk:** No critical cryptographic security issues detected")
	}

	return impacts
}

func generateExecutiveRecommendations(data *ReportData) []string {
	var recommendations []string

	// Prioritize based on actual findings
	if data.Summary.Keys != nil && data.Summary.Keys.TotalFindings > 0 {
		recommendations = append(recommendations,
			"Immediately remove hardcoded cryptographic keys from source code and implement secure secret management")
	}

	if data.Summary.TLS != nil {
		for _, f := range data.Findings.TLS {
			if f.Type == "disabled-verification" && f.Severity == "critical" {
				recommendations = append(recommendations,
					"Re-enable TLS certificate verification across all network connections")
				break
			}
		}
	}

	if data.Summary.Certificates != nil && data.Summary.Certificates.Expired > 0 {
		recommendations = append(recommendations,
			"Replace all expired X.509 certificates immediately to prevent service outages")
	}

	if data.Summary.Ciphers != nil && data.Summary.Ciphers.BySeverity["critical"] > 0 {
		recommendations = append(recommendations,
			"Replace critically weak cryptographic algorithms with NIST-approved alternatives")
	}

	if data.Summary.Random != nil && data.Summary.Random.TotalFindings > 0 {
		recommendations = append(recommendations,
			"Upgrade all random number generation to use cryptographically secure PRNGs")
	}

	// Add automated monitoring if certificates exist
	if data.Summary.Certificates != nil && data.Summary.Certificates.TotalCertificates > 0 {
		recommendations = append(recommendations,
			"Implement automated certificate expiry monitoring and alerting")
	}

	// General recommendation
	recommendations = append(recommendations,
		"Integrate automated cryptographic security scanning into CI/CD pipeline",
		"Establish and enforce cryptographic security standards across all projects")

	return recommendations
}
