package packageanalysis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ReportData holds the data needed to generate reports
type ReportData struct {
	Repository string
	Timestamp  time.Time
	Summary    Summary
	Findings   Findings
}

// LoadReportData loads package-analysis.json from the analysis directory
func LoadReportData(analysisDir string) (*ReportData, error) {
	packagePath := filepath.Join(analysisDir, "package-analysis.json")
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return nil, fmt.Errorf("reading package-analysis.json: %w", err)
	}

	var result struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing package-analysis.json: %w", err)
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
	var sb strings.Builder

	sb.WriteString("# Package Analysis Technical Report\n\n")
	sb.WriteString(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", data.Timestamp.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString("---\n\n")

	// Vulnerabilities Section
	if data.Summary.Vulns != nil {
		v := data.Summary.Vulns
		sb.WriteString("## 1. Vulnerabilities\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Severity | Count |\n")
		sb.WriteString("|----------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **CRITICAL** | %d |\n", v.Critical))
		sb.WriteString(fmt.Sprintf("| **HIGH** | %d |\n", v.High))
		sb.WriteString(fmt.Sprintf("| **MEDIUM** | %d |\n", v.Medium))
		sb.WriteString(fmt.Sprintf("| **LOW** | %d |\n", v.Low))
		sb.WriteString(fmt.Sprintf("| **TOTAL** | %d |\n", v.TotalVulnerabilities))
		sb.WriteString("\n")

		if v.KEVCount > 0 {
			sb.WriteString(fmt.Sprintf("**WARNING:** %d vulnerabilities are in CISA's Known Exploited Vulnerabilities (KEV) catalog.\n\n", v.KEVCount))
		}

		// Vulnerability findings details
		if vulns, ok := data.Findings.Vulns.([]interface{}); ok && len(vulns) > 0 {
			sb.WriteString("### Vulnerability Details\n\n")

			// Group by severity
			criticalVulns := []map[string]interface{}{}
			highVulns := []map[string]interface{}{}
			mediumVulns := []map[string]interface{}{}
			lowVulns := []map[string]interface{}{}

			for _, vData := range vulns {
				if vMap, ok := vData.(map[string]interface{}); ok {
					severity := strings.ToLower(fmt.Sprintf("%v", vMap["severity"]))
					switch severity {
					case "critical":
						criticalVulns = append(criticalVulns, vMap)
					case "high":
						highVulns = append(highVulns, vMap)
					case "medium":
						mediumVulns = append(mediumVulns, vMap)
					case "low":
						lowVulns = append(lowVulns, vMap)
					}
				}
			}

			// Display critical and high vulnerabilities
			if len(criticalVulns) > 0 {
				sb.WriteString("#### Critical Vulnerabilities\n\n")
				for _, v := range criticalVulns {
					writeVulnDetail(&sb, v)
				}
			}

			if len(highVulns) > 0 {
				sb.WriteString("#### High Vulnerabilities\n\n")
				for _, v := range highVulns {
					writeVulnDetail(&sb, v)
				}
			}

			// Summarize medium and low
			if len(mediumVulns) > 0 {
				sb.WriteString(fmt.Sprintf("#### Medium Vulnerabilities (%d)\n\n", len(mediumVulns)))
				sb.WriteString("<details>\n<summary>Show medium severity vulnerabilities</summary>\n\n")
				for _, v := range mediumVulns {
					writeVulnDetail(&sb, v)
				}
				sb.WriteString("</details>\n\n")
			}

			if len(lowVulns) > 0 {
				sb.WriteString(fmt.Sprintf("#### Low Vulnerabilities (%d)\n\n", len(lowVulns)))
				sb.WriteString("<details>\n<summary>Show low severity vulnerabilities</summary>\n\n")
				for _, v := range lowVulns {
					writeVulnDetail(&sb, v)
				}
				sb.WriteString("</details>\n\n")
			}
		}
	}

	// Malcontent Section
	if data.Summary.Malcontent != nil && data.Summary.Malcontent.TotalFindings > 0 {
		m := data.Summary.Malcontent
		sb.WriteString("## 2. Malware & Supply Chain Compromise\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Count |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Files Scanned | %d |\n", m.TotalFiles))
		sb.WriteString(fmt.Sprintf("| Files with Risk | %d |\n", m.FilesWithRisk))
		sb.WriteString(fmt.Sprintf("| Total Findings | %d |\n", m.TotalFindings))
		sb.WriteString(fmt.Sprintf("| Critical | %d |\n", m.Critical))
		sb.WriteString(fmt.Sprintf("| High | %d |\n", m.High))
		sb.WriteString(fmt.Sprintf("| Medium | %d |\n", m.Medium))
		sb.WriteString(fmt.Sprintf("| Low | %d |\n", m.Low))
		sb.WriteString("\n")

		// Malcontent findings details
		if malcontent, ok := data.Findings.Malcontent.([]interface{}); ok && len(malcontent) > 0 {
			sb.WriteString("### Malcontent Findings\n\n")

			// Filter critical and high only
			for _, mData := range malcontent {
				if mMap, ok := mData.(map[string]interface{}); ok {
					risk := strings.ToUpper(fmt.Sprintf("%v", mMap["risk"]))
					if risk == "CRITICAL" || risk == "HIGH" {
						writeMalcontentDetail(&sb, mMap)
					}
				}
			}
		}
	}

	// Package Health Section
	if data.Summary.Health != nil {
		h := data.Summary.Health
		sb.WriteString("## 3. Package Health\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Total Packages | %d |\n", h.TotalPackages))
		sb.WriteString(fmt.Sprintf("| Analyzed | %d |\n", h.AnalyzedCount))
		sb.WriteString(fmt.Sprintf("| Average Score | %.1f/100 |\n", h.AverageScore))
		sb.WriteString(fmt.Sprintf("| Healthy | %d |\n", h.HealthyCount))
		sb.WriteString(fmt.Sprintf("| Warnings | %d |\n", h.WarningCount))
		sb.WriteString(fmt.Sprintf("| Critical | %d |\n", h.CriticalCount))
		sb.WriteString(fmt.Sprintf("| Deprecated | %d |\n", h.DeprecatedCount))
		sb.WriteString(fmt.Sprintf("| Outdated | %d |\n", h.OutdatedCount))
		sb.WriteString("\n")

		// Health findings - show critical and warning packages
		if health, ok := data.Findings.Health.([]interface{}); ok && len(health) > 0 {
			sb.WriteString("### Packages Requiring Attention\n\n")

			criticalPkgs := []map[string]interface{}{}
			warningPkgs := []map[string]interface{}{}

			for _, hData := range health {
				if hMap, ok := hData.(map[string]interface{}); ok {
					status := strings.ToLower(fmt.Sprintf("%v", hMap["status"]))
					if status == "critical" {
						criticalPkgs = append(criticalPkgs, hMap)
					} else if status == "warning" {
						warningPkgs = append(warningPkgs, hMap)
					}
				}
			}

			if len(criticalPkgs) > 0 {
				sb.WriteString("#### Critical Health Issues\n\n")
				for _, pkg := range criticalPkgs {
					writeHealthDetail(&sb, pkg)
				}
			}

			if len(warningPkgs) > 0 {
				sb.WriteString(fmt.Sprintf("#### Warnings (%d packages)\n\n", len(warningPkgs)))
				sb.WriteString("<details>\n<summary>Show warning packages</summary>\n\n")
				for _, pkg := range warningPkgs {
					writeHealthDetail(&sb, pkg)
				}
				sb.WriteString("</details>\n\n")
			}
		}
	}

	// Licenses Section
	if data.Summary.Licenses != nil {
		l := data.Summary.Licenses
		sb.WriteString("## 4. License Analysis\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Total Packages | %d |\n", l.TotalPackages))
		sb.WriteString(fmt.Sprintf("| Unique Licenses | %d |\n", l.UniqueLicenses))
		sb.WriteString(fmt.Sprintf("| Allowed | %d |\n", l.Allowed))
		sb.WriteString(fmt.Sprintf("| Denied | %d |\n", l.Denied))
		sb.WriteString(fmt.Sprintf("| Needs Review | %d |\n", l.NeedsReview))
		sb.WriteString(fmt.Sprintf("| Unknown | %d |\n", l.Unknown))
		sb.WriteString(fmt.Sprintf("| Policy Violations | %d |\n", l.PolicyViolations))
		sb.WriteString("\n")

		// License distribution
		if len(l.LicenseCounts) > 0 {
			sb.WriteString("### License Distribution\n\n")
			sb.WriteString("| License | Package Count |\n")
			sb.WriteString("|---------|---------------|\n")

			// Sort licenses by count
			type licCount struct {
				name  string
				count int
			}
			var counts []licCount
			for name, count := range l.LicenseCounts {
				counts = append(counts, licCount{name, count})
			}
			sort.Slice(counts, func(i, j int) bool {
				return counts[i].count > counts[j].count
			})

			for _, lc := range counts {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", lc.name, lc.count))
			}
			sb.WriteString("\n")
		}

		// License findings - show denied and needs review
		if licenses, ok := data.Findings.Licenses.([]interface{}); ok && len(licenses) > 0 {
			deniedLics := []map[string]interface{}{}
			reviewLics := []map[string]interface{}{}

			for _, licData := range licenses {
				if licMap, ok := licData.(map[string]interface{}); ok {
					status := strings.ToLower(fmt.Sprintf("%v", licMap["status"]))
					if status == "denied" {
						deniedLics = append(deniedLics, licMap)
					} else if status == "needs_review" {
						reviewLics = append(reviewLics, licMap)
					}
				}
			}

			if len(deniedLics) > 0 {
				sb.WriteString("### Denied Licenses\n\n")
				for _, lic := range deniedLics {
					writeLicenseDetail(&sb, lic)
				}
			}

			if len(reviewLics) > 0 {
				sb.WriteString("### Licenses Requiring Review\n\n")
				for _, lic := range reviewLics {
					writeLicenseDetail(&sb, lic)
				}
			}
		}
	}

	// Confusion Section
	if data.Summary.Confusion != nil && data.Summary.Confusion.TotalFindings > 0 {
		c := data.Summary.Confusion
		sb.WriteString("## 5. Dependency Confusion\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Severity | Count |\n")
		sb.WriteString("|----------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Critical | %d |\n", c.Critical))
		sb.WriteString(fmt.Sprintf("| High | %d |\n", c.High))
		sb.WriteString(fmt.Sprintf("| Medium | %d |\n", c.Medium))
		sb.WriteString(fmt.Sprintf("| Low | %d |\n", c.Low))
		sb.WriteString(fmt.Sprintf("| **Total** | %d |\n", c.TotalFindings))
		sb.WriteString("\n")

		if len(c.ByEcosystem) > 0 {
			sb.WriteString("### By Ecosystem\n\n")
			for eco, count := range c.ByEcosystem {
				sb.WriteString(fmt.Sprintf("- **%s:** %d findings\n", eco, count))
			}
			sb.WriteString("\n")
		}
	}

	// Typosquats Section
	if data.Summary.Typosquats != nil && data.Summary.Typosquats.SuspiciousCount > 0 {
		t := data.Summary.Typosquats
		sb.WriteString("## 6. Typosquatting Detection\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Packages Checked:** %d\n", t.TotalChecked))
		sb.WriteString(fmt.Sprintf("- **Suspicious Packages:** %d\n", t.SuspiciousCount))
		sb.WriteString(fmt.Sprintf("- **New Packages (<30 days):** %d\n\n", t.NewPackagesCount))

		if typosquats, ok := data.Findings.Typosquats.([]interface{}); ok && len(typosquats) > 0 {
			sb.WriteString("### Suspicious Packages\n\n")
			for _, tsData := range typosquats {
				if tsMap, ok := tsData.(map[string]interface{}); ok {
					writeTyposquatDetail(&sb, tsMap)
				}
			}
		}
	}

	// Deprecations Section
	if data.Summary.Deprecations != nil && data.Summary.Deprecations.DeprecatedCount > 0 {
		d := data.Summary.Deprecations
		sb.WriteString("## 7. Deprecated Packages\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Total Packages:** %d\n", d.TotalPackages))
		sb.WriteString(fmt.Sprintf("- **Deprecated:** %d\n\n", d.DeprecatedCount))

		if len(d.ByEcosystem) > 0 {
			sb.WriteString("### By Ecosystem\n\n")
			for eco, count := range d.ByEcosystem {
				sb.WriteString(fmt.Sprintf("- **%s:** %d deprecated\n", eco, count))
			}
			sb.WriteString("\n")
		}

		if deps, ok := data.Findings.Deprecations.([]interface{}); ok && len(deps) > 0 {
			sb.WriteString("### Deprecated Package Details\n\n")
			for _, depData := range deps {
				if depMap, ok := depData.(map[string]interface{}); ok {
					writeDeprecationDetail(&sb, depMap)
				}
			}
		}
	}

	// Duplicates Section
	if data.Summary.Duplicates != nil && (data.Summary.Duplicates.DuplicateVersions > 0 || data.Summary.Duplicates.DuplicateFunctionality > 0) {
		dup := data.Summary.Duplicates
		sb.WriteString("## 8. Duplicate Dependencies\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Total Packages:** %d\n", dup.TotalPackages))
		sb.WriteString(fmt.Sprintf("- **Duplicate Versions:** %d (same package, different versions)\n", dup.DuplicateVersions))
		sb.WriteString(fmt.Sprintf("- **Duplicate Functionality:** %d (different packages, same purpose)\n\n", dup.DuplicateFunctionality))

		if dups, ok := data.Findings.Duplicates.([]interface{}); ok && len(dups) > 0 {
			sb.WriteString("### Duplicate Details\n\n")
			for _, dupData := range dups {
				if dupMap, ok := dupData.(map[string]interface{}); ok {
					writeDuplicateDetail(&sb, dupMap)
				}
			}
		}
	}

	// Reachability Section
	if data.Summary.Reachability != nil && data.Summary.Reachability.Supported {
		r := data.Summary.Reachability
		sb.WriteString("## 9. Vulnerability Reachability Analysis\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Total Vulnerabilities | %d |\n", r.TotalVulns))
		sb.WriteString(fmt.Sprintf("| Reachable | %d |\n", r.ReachableVulns))
		sb.WriteString(fmt.Sprintf("| Unreachable | %d |\n", r.UnreachableVulns))
		sb.WriteString(fmt.Sprintf("| Unknown | %d |\n", r.UnknownReachability))
		sb.WriteString(fmt.Sprintf("| **Noise Reduction** | %.1f%% |\n", r.ReductionPercent))
		sb.WriteString("\n")

		if r.ReductionPercent > 0 {
			sb.WriteString(fmt.Sprintf("**Analysis Impact:** Reachability analysis reduced the effective vulnerability count by %.1f%%, allowing teams to focus on %d truly reachable vulnerabilities.\n\n", r.ReductionPercent, r.ReachableVulns))
		}
	}

	// Provenance Section
	if data.Summary.Provenance != nil && data.Summary.Provenance.TotalPackages > 0 {
		p := data.Summary.Provenance
		sb.WriteString("## 10. Provenance Verification\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Total Packages | %d |\n", p.TotalPackages))
		sb.WriteString(fmt.Sprintf("| Verified | %d |\n", p.VerifiedCount))
		sb.WriteString(fmt.Sprintf("| Unverified | %d |\n", p.UnverifiedCount))
		sb.WriteString(fmt.Sprintf("| Suspicious | %d |\n", p.SuspiciousCount))
		sb.WriteString(fmt.Sprintf("| Verification Rate | %.1f%% |\n", p.VerificationRate))
		sb.WriteString("\n")

		if p.SuspiciousCount > 0 {
			sb.WriteString(fmt.Sprintf("**WARNING:** %d packages have suspicious provenance and should be investigated.\n\n", p.SuspiciousCount))
		}
	}

	// Bundle Analysis Section
	if data.Summary.Bundle != nil && data.Summary.Bundle.TotalPackages > 0 {
		b := data.Summary.Bundle
		sb.WriteString("## 11. Bundle Analysis\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Total Packages | %d |\n", b.TotalPackages))
		sb.WriteString(fmt.Sprintf("| Heavy Packages | %d |\n", b.HeavyPackages))
		sb.WriteString(fmt.Sprintf("| Duplicates | %d |\n", b.DuplicatePackages))
		sb.WriteString(fmt.Sprintf("| Treeshake Candidates | %d |\n", b.TreeshakeCandidates))
		sb.WriteString(fmt.Sprintf("| Total Size | %d KB |\n", b.TotalSizeKB))
		sb.WriteString("\n")
	}

	// Recommendations Section
	if data.Summary.Recommendations != nil && data.Summary.Recommendations.TotalRecommendations > 0 {
		rec := data.Summary.Recommendations
		sb.WriteString("## 12. Package Recommendations\n\n")

		sb.WriteString("### Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Total Recommendations:** %d\n", rec.TotalRecommendations))
		sb.WriteString(fmt.Sprintf("- **Security:** %d\n", rec.SecurityRecommendations))
		sb.WriteString(fmt.Sprintf("- **Health:** %d\n\n", rec.HealthRecommendations))

		if recs, ok := data.Findings.Recommendations.([]interface{}); ok && len(recs) > 0 {
			sb.WriteString("### Recommendations\n\n")

			// Group by priority
			highPriority := []map[string]interface{}{}
			mediumPriority := []map[string]interface{}{}
			lowPriority := []map[string]interface{}{}

			for _, recData := range recs {
				if recMap, ok := recData.(map[string]interface{}); ok {
					priority := strings.ToLower(fmt.Sprintf("%v", recMap["priority"]))
					switch priority {
					case "high":
						highPriority = append(highPriority, recMap)
					case "medium":
						mediumPriority = append(mediumPriority, recMap)
					case "low":
						lowPriority = append(lowPriority, recMap)
					}
				}
			}

			if len(highPriority) > 0 {
				sb.WriteString("#### High Priority\n\n")
				for _, r := range highPriority {
					writeRecommendationDetail(&sb, r)
				}
			}

			if len(mediumPriority) > 0 {
				sb.WriteString("#### Medium Priority\n\n")
				for _, r := range mediumPriority {
					writeRecommendationDetail(&sb, r)
				}
			}

			if len(lowPriority) > 0 {
				sb.WriteString("#### Low Priority\n\n")
				for _, r := range lowPriority {
					writeRecommendationDetail(&sb, r)
				}
			}
		}
	}

	// Errors Section
	if len(data.Summary.Errors) > 0 {
		sb.WriteString("## Errors\n\n")
		for _, err := range data.Summary.Errors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("*Generated by Zero Package Analysis Scanner*\n")

	return sb.String()
}

// GenerateExecutiveReport creates a high-level security summary for leadership
func GenerateExecutiveReport(data *ReportData) string {
	var sb strings.Builder

	sb.WriteString("# Package Security Report\n\n")
	sb.WriteString(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	sb.WriteString(fmt.Sprintf("**Date:** %s\n\n", data.Timestamp.Format("January 2, 2006")))

	// Executive Summary
	sb.WriteString("## Executive Summary\n\n")

	// Calculate risk level
	riskLevel := calculateRiskLevel(data.Summary)
	sb.WriteString(fmt.Sprintf("### Overall Risk Level: %s\n\n", strings.ToUpper(riskLevel)))

	// Key Metrics
	sb.WriteString("## Key Metrics\n\n")
	sb.WriteString("| Category | Status | Details |\n")
	sb.WriteString("|----------|--------|----------|\n")

	// Vulnerabilities
	if data.Summary.Vulns != nil {
		v := data.Summary.Vulns
		vulnStatus := getVulnStatus(v.Critical, v.High)
		sb.WriteString(fmt.Sprintf("| **Vulnerabilities** | %s | %d total (%d critical, %d high) |\n", vulnStatus, v.TotalVulnerabilities, v.Critical, v.High))
		if v.KEVCount > 0 {
			sb.WriteString(fmt.Sprintf("| | **WARNING** | %d in CISA KEV catalog |\n", v.KEVCount))
		}
	}

	// Malcontent
	if data.Summary.Malcontent != nil && data.Summary.Malcontent.TotalFindings > 0 {
		m := data.Summary.Malcontent
		malcontentStatus := getMalcontentStatus(m.Critical, m.High)
		sb.WriteString(fmt.Sprintf("| **Malware/Supply Chain** | %s | %d findings (%d critical, %d high) |\n", malcontentStatus, m.TotalFindings, m.Critical, m.High))
	}

	// Package Health
	if data.Summary.Health != nil {
		h := data.Summary.Health
		healthStatus := getHealthStatus(h.AverageScore)
		sb.WriteString(fmt.Sprintf("| **Package Health** | %s | Avg score: %.1f/100 (%d outdated, %d deprecated) |\n", healthStatus, h.AverageScore, h.OutdatedCount, h.DeprecatedCount))
	}

	// Licenses
	if data.Summary.Licenses != nil {
		l := data.Summary.Licenses
		licenseStatus := getLicenseStatus(l.Denied, l.NeedsReview)
		sb.WriteString(fmt.Sprintf("| **Licenses** | %s | %d unique (%d denied, %d need review) |\n", licenseStatus, l.UniqueLicenses, l.Denied, l.NeedsReview))
	}

	// Typosquats
	if data.Summary.Typosquats != nil && data.Summary.Typosquats.SuspiciousCount > 0 {
		t := data.Summary.Typosquats
		sb.WriteString(fmt.Sprintf("| **Typosquatting** | WARNING | %d suspicious packages detected |\n", t.SuspiciousCount))
	}

	// Confusion
	if data.Summary.Confusion != nil && data.Summary.Confusion.TotalFindings > 0 {
		c := data.Summary.Confusion
		confusionStatus := getConfusionStatus(c.Critical, c.High)
		sb.WriteString(fmt.Sprintf("| **Dependency Confusion** | %s | %d findings (%d critical, %d high) |\n", confusionStatus, c.TotalFindings, c.Critical, c.High))
	}

	sb.WriteString("\n")

	// Critical Findings
	criticalFindings := collectCriticalFindings(data)
	if len(criticalFindings) > 0 {
		sb.WriteString("## Critical Findings\n\n")
		for _, finding := range criticalFindings {
			sb.WriteString(fmt.Sprintf("- %s\n", finding))
		}
		sb.WriteString("\n")
	}

	// Risk Assessment
	sb.WriteString("## Risk Assessment\n\n")

	risks := assessRisks(data)
	if len(risks.immediate) > 0 {
		sb.WriteString("### Immediate Attention Required\n\n")
		for _, risk := range risks.immediate {
			sb.WriteString(fmt.Sprintf("- %s\n", risk))
		}
		sb.WriteString("\n")
	}

	if len(risks.shortTerm) > 0 {
		sb.WriteString("### Short-term Concerns\n\n")
		for _, risk := range risks.shortTerm {
			sb.WriteString(fmt.Sprintf("- %s\n", risk))
		}
		sb.WriteString("\n")
	}

	// Recommendations
	sb.WriteString("## Recommended Actions\n\n")

	actions := generateActions(data)
	if len(actions.critical) > 0 {
		sb.WriteString("### Critical Actions\n\n")
		for i, action := range actions.critical {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, action))
		}
		sb.WriteString("\n")
	}

	if len(actions.important) > 0 {
		sb.WriteString("### Important Actions\n\n")
		for i, action := range actions.important {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, action))
		}
		sb.WriteString("\n")
	}

	// Positive Findings
	positiveFindings := collectPositiveFindings(data)
	if len(positiveFindings) > 0 {
		sb.WriteString("## Positive Findings\n\n")
		for _, finding := range positiveFindings {
			sb.WriteString(fmt.Sprintf("- %s\n", finding))
		}
		sb.WriteString("\n")
	}

	// Business Impact
	sb.WriteString("## Business Impact\n\n")
	impact := assessBusinessImpact(data)
	sb.WriteString(fmt.Sprintf("**Security Risk:** %s\n\n", impact.securityRisk))
	sb.WriteString(fmt.Sprintf("**Compliance Risk:** %s\n\n", impact.complianceRisk))
	if impact.keyRisk != "" {
		sb.WriteString(fmt.Sprintf("**Key Risk:** %s\n\n", impact.keyRisk))
	}

	sb.WriteString("---\n\n")
	sb.WriteString("*Generated by Zero Package Analysis Scanner*\n")

	return sb.String()
}

// WriteReports generates and writes both reports to the analysis directory
func WriteReports(analysisDir string) error {
	data, err := LoadReportData(analysisDir)
	if err != nil {
		return err
	}

	// Write technical report
	techReport := GenerateTechnicalReport(data)
	techPath := filepath.Join(analysisDir, "package-analysis-technical-report.md")
	if err := os.WriteFile(techPath, []byte(techReport), 0644); err != nil {
		return fmt.Errorf("writing technical report: %w", err)
	}

	// Write executive report
	execReport := GenerateExecutiveReport(data)
	execPath := filepath.Join(analysisDir, "package-analysis-executive-report.md")
	if err := os.WriteFile(execPath, []byte(execReport), 0644); err != nil {
		return fmt.Errorf("writing executive report: %w", err)
	}

	return nil
}

// Helper functions for writing details

func writeVulnDetail(sb *strings.Builder, v map[string]interface{}) {
	pkg := fmt.Sprintf("%v", v["package"])
	version := fmt.Sprintf("%v", v["version"])
	id := fmt.Sprintf("%v", v["id"])
	title := fmt.Sprintf("%v", v["title"])
	fixedIn := fmt.Sprintf("%v", v["fixed_in"])
	inKEV := fmt.Sprintf("%v", v["in_kev"]) == "true"

	sb.WriteString(fmt.Sprintf("**%s** - `%s@%s`\n", id, pkg, version))
	if title != "" && title != "<nil>" {
		sb.WriteString(fmt.Sprintf("- %s\n", title))
	}
	if fixedIn != "" && fixedIn != "<nil>" {
		sb.WriteString(fmt.Sprintf("- **Fix:** Upgrade to version %s\n", fixedIn))
	}
	if inKEV {
		sb.WriteString("- **CISA KEV:** This vulnerability is in the Known Exploited Vulnerabilities catalog\n")
	}
	sb.WriteString("\n")
}

func writeMalcontentDetail(sb *strings.Builder, m map[string]interface{}) {
	file := fmt.Sprintf("%v", m["file"])
	risk := fmt.Sprintf("%v", m["risk"])
	riskScore := fmt.Sprintf("%v", m["risk_score"])
	behaviors := m["behaviors"]

	sb.WriteString(fmt.Sprintf("**[%s]** `%s` (Risk Score: %s)\n", risk, file, riskScore))
	if behaviorList, ok := behaviors.([]interface{}); ok && len(behaviorList) > 0 {
		sb.WriteString("Suspicious behaviors:\n")
		for _, b := range behaviorList {
			sb.WriteString(fmt.Sprintf("  - %v\n", b))
		}
	}
	sb.WriteString("\n")
}

func writeHealthDetail(sb *strings.Builder, h map[string]interface{}) {
	pkg := fmt.Sprintf("%v", h["package"])
	version := fmt.Sprintf("%v", h["version"])
	healthScore := fmt.Sprintf("%v", h["health_score"])
	isDeprecated := fmt.Sprintf("%v", h["is_deprecated"]) == "true"
	isOutdated := fmt.Sprintf("%v", h["is_outdated"]) == "true"
	latestVersion := fmt.Sprintf("%v", h["latest_version"])

	sb.WriteString(fmt.Sprintf("**%s@%s** - Health Score: %s/100\n", pkg, version, healthScore))
	if isDeprecated {
		sb.WriteString("- **DEPRECATED** - This package is no longer maintained\n")
	}
	if isOutdated && latestVersion != "" && latestVersion != "<nil>" {
		sb.WriteString(fmt.Sprintf("- **OUTDATED** - Latest version: %s\n", latestVersion))
	}
	sb.WriteString("\n")
}

func writeLicenseDetail(sb *strings.Builder, l map[string]interface{}) {
	pkg := fmt.Sprintf("%v", l["package"])
	version := fmt.Sprintf("%v", l["version"])
	licenses := l["licenses"]
	status := fmt.Sprintf("%v", l["status"])

	sb.WriteString(fmt.Sprintf("**%s@%s** - Status: %s\n", pkg, version, strings.ToUpper(status)))
	if licList, ok := licenses.([]interface{}); ok && len(licList) > 0 {
		sb.WriteString("Licenses: ")
		licStrings := []string{}
		for _, lic := range licList {
			licStrings = append(licStrings, fmt.Sprintf("%v", lic))
		}
		sb.WriteString(strings.Join(licStrings, ", "))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func writeTyposquatDetail(sb *strings.Builder, t map[string]interface{}) {
	pkg := fmt.Sprintf("%v", t["package"])
	similarTo := fmt.Sprintf("%v", t["similar_to"])
	reason := fmt.Sprintf("%v", t["reason"])
	riskLevel := fmt.Sprintf("%v", t["risk_level"])

	sb.WriteString(fmt.Sprintf("**[%s]** `%s`\n", strings.ToUpper(riskLevel), pkg))
	if similarTo != "" && similarTo != "<nil>" {
		sb.WriteString(fmt.Sprintf("- Similar to: %s\n", similarTo))
	}
	if reason != "" && reason != "<nil>" {
		sb.WriteString(fmt.Sprintf("- Reason: %s\n", reason))
	}
	sb.WriteString("\n")
}

func writeDeprecationDetail(sb *strings.Builder, d map[string]interface{}) {
	pkg := fmt.Sprintf("%v", d["package"])
	version := fmt.Sprintf("%v", d["version"])
	message := fmt.Sprintf("%v", d["message"])
	alternative := fmt.Sprintf("%v", d["alternative"])

	sb.WriteString(fmt.Sprintf("**%s@%s**\n", pkg, version))
	if message != "" && message != "<nil>" {
		sb.WriteString(fmt.Sprintf("- %s\n", message))
	}
	if alternative != "" && alternative != "<nil>" {
		sb.WriteString(fmt.Sprintf("- Suggested alternative: %s\n", alternative))
	}
	sb.WriteString("\n")
}

func writeDuplicateDetail(sb *strings.Builder, d map[string]interface{}) {
	pkg := fmt.Sprintf("%v", d["package"])
	issueType := fmt.Sprintf("%v", d["issue_type"])
	message := fmt.Sprintf("%v", d["message"])
	versions := d["versions"]

	sb.WriteString(fmt.Sprintf("**%s** - Type: %s\n", pkg, issueType))
	if message != "" && message != "<nil>" {
		sb.WriteString(fmt.Sprintf("- %s\n", message))
	}
	if verList, ok := versions.([]interface{}); ok && len(verList) > 0 {
		sb.WriteString("- Versions: ")
		verStrings := []string{}
		for _, ver := range verList {
			verStrings = append(verStrings, fmt.Sprintf("%v", ver))
		}
		sb.WriteString(strings.Join(verStrings, ", "))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func writeRecommendationDetail(sb *strings.Builder, r map[string]interface{}) {
	pkg := fmt.Sprintf("%v", r["package"])
	currentVersion := fmt.Sprintf("%v", r["current_version"])
	alternative := fmt.Sprintf("%v", r["alternative"])
	reason := fmt.Sprintf("%v", r["reason"])

	sb.WriteString(fmt.Sprintf("**%s@%s**\n", pkg, currentVersion))
	if reason != "" && reason != "<nil>" {
		sb.WriteString(fmt.Sprintf("- Reason: %s\n", reason))
	}
	if alternative != "" && alternative != "<nil>" {
		sb.WriteString(fmt.Sprintf("- Recommended: %s\n", alternative))
	}
	sb.WriteString("\n")
}

// Helper functions for executive report

func calculateRiskLevel(summary Summary) string {
	score := 0

	// Vulnerabilities (highest weight)
	if summary.Vulns != nil {
		if summary.Vulns.Critical > 0 {
			score += 40
		}
		if summary.Vulns.High > 0 {
			score += 20
		}
		if summary.Vulns.KEVCount > 0 {
			score += 20
		}
	}

	// Malcontent (high weight)
	if summary.Malcontent != nil {
		if summary.Malcontent.Critical > 0 {
			score += 30
		}
		if summary.Malcontent.High > 0 {
			score += 15
		}
	}

	// Confusion (medium weight)
	if summary.Confusion != nil {
		if summary.Confusion.Critical > 0 {
			score += 20
		}
		if summary.Confusion.High > 0 {
			score += 10
		}
	}

	// Typosquats (medium weight)
	if summary.Typosquats != nil && summary.Typosquats.SuspiciousCount > 0 {
		score += 10
	}

	// Licenses (low weight but important)
	if summary.Licenses != nil && summary.Licenses.Denied > 0 {
		score += 10
	}

	// Determine risk level
	switch {
	case score >= 50:
		return "critical"
	case score >= 30:
		return "high"
	case score >= 15:
		return "medium"
	default:
		return "low"
	}
}

func getRiskColor(level string) string {
	switch level {
	case "critical":
		return "red"
	case "high":
		return "orange"
	case "medium":
		return "yellow"
	default:
		return "green"
	}
}

func getVulnStatus(critical, high int) string {
	if critical > 0 {
		return "CRITICAL"
	}
	if high > 0 {
		return "HIGH"
	}
	return "OK"
}

func getMalcontentStatus(critical, high int) string {
	if critical > 0 {
		return "CRITICAL"
	}
	if high > 0 {
		return "HIGH"
	}
	return "WARNING"
}

func getHealthStatus(avgScore float64) string {
	switch {
	case avgScore >= 80:
		return "Excellent"
	case avgScore >= 60:
		return "Good"
	case avgScore >= 40:
		return "Fair"
	default:
		return "Poor"
	}
}

func getLicenseStatus(denied, needsReview int) string {
	if denied > 0 {
		return "CRITICAL"
	}
	if needsReview > 0 {
		return "Needs Review"
	}
	return "OK"
}

func getConfusionStatus(critical, high int) string {
	if critical > 0 {
		return "CRITICAL"
	}
	if high > 0 {
		return "HIGH"
	}
	return "WARNING"
}

func collectCriticalFindings(data *ReportData) []string {
	findings := []string{}

	// Critical vulnerabilities
	if data.Summary.Vulns != nil && data.Summary.Vulns.Critical > 0 {
		findings = append(findings, fmt.Sprintf("%d critical vulnerabilities detected", data.Summary.Vulns.Critical))
	}

	// KEV vulnerabilities
	if data.Summary.Vulns != nil && data.Summary.Vulns.KEVCount > 0 {
		findings = append(findings, fmt.Sprintf("%d vulnerabilities are actively exploited (CISA KEV)", data.Summary.Vulns.KEVCount))
	}

	// Critical malcontent
	if data.Summary.Malcontent != nil && data.Summary.Malcontent.Critical > 0 {
		findings = append(findings, fmt.Sprintf("%d packages with critical malware/supply chain indicators", data.Summary.Malcontent.Critical))
	}

	// Denied licenses
	if data.Summary.Licenses != nil && data.Summary.Licenses.Denied > 0 {
		findings = append(findings, fmt.Sprintf("%d packages using denied licenses", data.Summary.Licenses.Denied))
	}

	// Critical dependency confusion
	if data.Summary.Confusion != nil && data.Summary.Confusion.Critical > 0 {
		findings = append(findings, fmt.Sprintf("%d critical dependency confusion risks", data.Summary.Confusion.Critical))
	}

	return findings
}

type riskAssessment struct {
	immediate []string
	shortTerm []string
}

func assessRisks(data *ReportData) riskAssessment {
	risks := riskAssessment{}

	// Immediate risks
	if data.Summary.Vulns != nil && (data.Summary.Vulns.Critical > 0 || data.Summary.Vulns.KEVCount > 0) {
		risks.immediate = append(risks.immediate, "Critical vulnerabilities or actively exploited vulnerabilities present")
	}

	if data.Summary.Malcontent != nil && data.Summary.Malcontent.Critical > 0 {
		risks.immediate = append(risks.immediate, "Potential supply chain compromise detected in dependencies")
	}

	if data.Summary.Licenses != nil && data.Summary.Licenses.Denied > 0 {
		risks.immediate = append(risks.immediate, "License compliance violations that may create legal exposure")
	}

	// Short-term risks
	if data.Summary.Vulns != nil && data.Summary.Vulns.High > 5 {
		risks.shortTerm = append(risks.shortTerm, fmt.Sprintf("%d high severity vulnerabilities should be addressed", data.Summary.Vulns.High))
	}

	if data.Summary.Health != nil && data.Summary.Health.DeprecatedCount > 0 {
		risks.shortTerm = append(risks.shortTerm, fmt.Sprintf("%d deprecated packages may lose support and security updates", data.Summary.Health.DeprecatedCount))
	}

	if data.Summary.Health != nil && data.Summary.Health.OutdatedCount > 10 {
		risks.shortTerm = append(risks.shortTerm, fmt.Sprintf("%d outdated packages may miss important security patches", data.Summary.Health.OutdatedCount))
	}

	if data.Summary.Typosquats != nil && data.Summary.Typosquats.SuspiciousCount > 0 {
		risks.shortTerm = append(risks.shortTerm, fmt.Sprintf("%d suspicious packages that may be typosquats", data.Summary.Typosquats.SuspiciousCount))
	}

	return risks
}

type actionItems struct {
	critical  []string
	important []string
}

func generateActions(data *ReportData) actionItems {
	actions := actionItems{}

	// Critical actions
	if data.Summary.Vulns != nil && data.Summary.Vulns.Critical > 0 {
		actions.critical = append(actions.critical, "Patch or mitigate all critical vulnerabilities immediately")
	}

	if data.Summary.Vulns != nil && data.Summary.Vulns.KEVCount > 0 {
		actions.critical = append(actions.critical, "Address all CISA KEV vulnerabilities - these are actively exploited")
	}

	if data.Summary.Malcontent != nil && data.Summary.Malcontent.Critical > 0 {
		actions.critical = append(actions.critical, "Investigate and remove packages with critical malware indicators")
	}

	if data.Summary.Licenses != nil && data.Summary.Licenses.Denied > 0 {
		actions.critical = append(actions.critical, "Remove or replace packages with denied licenses")
	}

	// Important actions
	if data.Summary.Vulns != nil && data.Summary.Vulns.High > 0 {
		actions.important = append(actions.important, fmt.Sprintf("Address %d high severity vulnerabilities", data.Summary.Vulns.High))
	}

	if data.Summary.Health != nil && data.Summary.Health.DeprecatedCount > 0 {
		actions.important = append(actions.important, "Plan migration away from deprecated packages")
	}

	if data.Summary.Health != nil && data.Summary.Health.OutdatedCount > 5 {
		actions.important = append(actions.important, "Update outdated packages to latest secure versions")
	}

	if data.Summary.Licenses != nil && data.Summary.Licenses.NeedsReview > 0 {
		actions.important = append(actions.important, "Review license compatibility for flagged packages")
	}

	if data.Summary.Duplicates != nil && data.Summary.Duplicates.DuplicateVersions > 0 {
		actions.important = append(actions.important, "Consolidate duplicate package versions to reduce attack surface")
	}

	return actions
}

func collectPositiveFindings(data *ReportData) []string {
	findings := []string{}

	// No critical vulnerabilities
	if data.Summary.Vulns != nil && data.Summary.Vulns.Critical == 0 && data.Summary.Vulns.TotalVulnerabilities > 0 {
		findings = append(findings, "No critical vulnerabilities detected")
	}

	// Good health score
	if data.Summary.Health != nil && data.Summary.Health.AverageScore >= 75 {
		findings = append(findings, fmt.Sprintf("Good average package health score: %.1f/100", data.Summary.Health.AverageScore))
	}

	// No malcontent
	if data.Summary.Malcontent != nil && data.Summary.Malcontent.Critical == 0 && data.Summary.Malcontent.High == 0 {
		findings = append(findings, "No critical supply chain compromise indicators detected")
	}

	// License compliance
	if data.Summary.Licenses != nil && data.Summary.Licenses.Denied == 0 {
		findings = append(findings, "No license compliance violations")
	}

	// Reachability analysis
	if data.Summary.Reachability != nil && data.Summary.Reachability.Supported && data.Summary.Reachability.ReductionPercent > 50 {
		findings = append(findings, fmt.Sprintf("Reachability analysis reduced effective vulnerability count by %.1f%%", data.Summary.Reachability.ReductionPercent))
	}

	// Provenance verification
	if data.Summary.Provenance != nil && data.Summary.Provenance.VerificationRate > 80 {
		findings = append(findings, fmt.Sprintf("Strong provenance verification: %.1f%% of packages verified", data.Summary.Provenance.VerificationRate))
	}

	return findings
}

type businessImpact struct {
	securityRisk   string
	complianceRisk string
	keyRisk        string
}

func assessBusinessImpact(data *ReportData) businessImpact {
	impact := businessImpact{
		securityRisk:   "Unknown",
		complianceRisk: "Unknown",
	}

	// Security risk assessment
	riskLevel := calculateRiskLevel(data.Summary)
	switch riskLevel {
	case "critical":
		impact.securityRisk = "Critical - Immediate action required"
		impact.keyRisk = "Active security vulnerabilities or supply chain compromise pose immediate risk to application and data security"
	case "high":
		impact.securityRisk = "High - Address within 7 days"
	case "medium":
		impact.securityRisk = "Medium - Address within 30 days"
	case "low":
		impact.securityRisk = "Low - Monitor and maintain"
	}

	// Compliance risk assessment
	if data.Summary.Licenses != nil {
		if data.Summary.Licenses.Denied > 0 {
			impact.complianceRisk = "High - License violations detected"
			if impact.keyRisk == "" {
				impact.keyRisk = "License compliance violations may result in legal liability and intellectual property risks"
			}
		} else if data.Summary.Licenses.NeedsReview > 0 {
			impact.complianceRisk = "Medium - Licenses require review"
		} else {
			impact.complianceRisk = "Low - No compliance issues"
		}
	}

	return impact
}
