// Package codeownership provides report generation for code ownership analysis.
//
// This file implements two types of reports:
//
// 1. Technical Report (code-ownership-technical-report.md):
//    - Detailed analysis for engineers
//    - Complete contributor metrics and scoring
//    - CODEOWNERS validation and drift analysis
//    - Orphaned files listing
//    - Developer competencies by language
//    - Monorepo workspace ownership
//    - Incident response contacts
//
// 2. Executive Report (code-ownership-executive-report.md):
//    - High-level summary for engineering leaders
//    - Team health status and bus factor assessment
//    - Risk analysis (knowledge loss, maintenance, incident response)
//    - Business impact metrics
//    - Prioritized recommendations
//
// Both reports focus on ownership health, highlighting:
//   - Bus factor risks (knowledge concentration)
//   - Files without clear owners
//   - Areas with high churn but low ownership
//   - Contributor distribution and team health
package codeownership

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

// LoadReportData loads code-ownership.json from the analysis directory
func LoadReportData(analysisDir string) (*ReportData, error) {
	ownershipPath := filepath.Join(analysisDir, "code-ownership.json")
	data, err := os.ReadFile(ownershipPath)
	if err != nil {
		return nil, fmt.Errorf("reading code-ownership.json: %w", err)
	}

	var result struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing code-ownership.json: %w", err)
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

	b.Title("Code Ownership Technical Report")
	b.Raw(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	b.Raw(fmt.Sprintf("**Generated:** %s\n", data.Timestamp.Format("2006-01-02 15:04:05 UTC")))
	b.Raw(fmt.Sprintf("**Scanner:** %s\n\n", Description))
	b.Divider()

	// Overview Section
	b.Section(2, "Overview")
	overviewRows := [][]string{
		{"Total Contributors", fmt.Sprintf("%d", data.Summary.TotalContributors)},
		{"Files Analyzed", fmt.Sprintf("%d", data.Summary.FilesAnalyzed)},
		{"Analysis Period", fmt.Sprintf("%d days", data.Summary.PeriodDays)},
		{"Bus Factor", formatBusFactor(data.Summary.BusFactor, data.Summary.BusFactorRisk)},
		{"Ownership Coverage", fmt.Sprintf("%.1f%%", data.Summary.OwnershipCoverage*100)},
		{"Has CODEOWNERS", formatBool(data.Summary.HasCodeowners)},
	}

	if data.Summary.HasCodeowners {
		overviewRows = append(overviewRows, []string{"CODEOWNERS Rules", fmt.Sprintf("%d", data.Summary.CodeownersRules)})
		if data.Summary.CodeownersIssues > 0 {
			overviewRows = append(overviewRows, []string{"CODEOWNERS Issues", fmt.Sprintf("%d", data.Summary.CodeownersIssues)})
		}
	}

	overviewRows = append(overviewRows, []string{"Orphaned Files", fmt.Sprintf("%d", data.Summary.OrphanedFiles)})

	if data.Summary.IsMonorepo {
		overviewRows = append(overviewRows, []string{"Monorepo", fmt.Sprintf("Yes (%d workspaces)", data.Summary.WorkspaceCount)})
	}

	b.Table([]string{"Metric", "Value"}, overviewRows)

	// Repository Activity
	if data.Summary.LastCommitDate != "" {
		b.Section(2, "Repository Activity")
		activityRows := [][]string{
			{"Last Commit", data.Summary.LastCommitDate},
			{"Days Since Last Commit", fmt.Sprintf("%d", data.Summary.DaysSinceLastCommit)},
			{"Activity Status", formatActivityStatus(data.Summary.RepoActivityStatus)},
			{"Total Commits (All Time)", fmt.Sprintf("%d", data.Summary.TotalCommits)},
			{"All-Time Contributors", fmt.Sprintf("%d", data.Summary.AllTimeContributors)},
		}
		b.Table([]string{"Metric", "Value"}, activityRows)
	}

	// Language Distribution
	if len(data.Summary.TopLanguages) > 0 {
		b.Section(2, "Language Distribution")
		langRows := [][]string{}
		for _, lang := range data.Summary.TopLanguages {
			langRows = append(langRows, []string{
				lang.Name,
				fmt.Sprintf("%d", lang.FileCount),
				fmt.Sprintf("%.1f%%", lang.Percentage),
			})
		}
		b.Table([]string{"Language", "Files", "Percentage"}, langRows)
	}

	// Top Contributors
	if len(data.Findings.Contributors) > 0 {
		b.Section(2, "Top Contributors")
		b.Paragraph(fmt.Sprintf("Based on activity in the last %d days.", data.Summary.PeriodDays))

		contribRows := [][]string{}
		sortedContribs := sortContributorsByCommits(data.Findings.Contributors)
		limit := min(15, len(sortedContribs))

		for i := 0; i < limit; i++ {
			c := sortedContribs[i]
			contribRows = append(contribRows, []string{
				c.Name,
				c.Email,
				fmt.Sprintf("%d", c.Commits),
				fmt.Sprintf("%d", c.FilesTouched),
				fmt.Sprintf("+%d/-%d", c.LinesAdded, c.LinesRemoved),
			})
		}

		b.Table([]string{"Name", "Email", "Commits", "Files", "Lines"}, contribRows)

		if len(sortedContribs) > limit {
			b.Paragraph(fmt.Sprintf("*...and %d more contributors*", len(sortedContribs)-limit))
		}
	}

	// Enhanced Ownership Analysis
	if len(data.Findings.EnhancedOwnership) > 0 {
		b.Section(2, "Enhanced Ownership Scores")
		b.Paragraph("Multi-factor ownership analysis combining commits, reviews, recency, and consistency.")

		ownershipRows := [][]string{}
		sortedOwners := sortEnhancedOwnershipByScore(data.Findings.EnhancedOwnership)
		limit := min(15, len(sortedOwners))

		for i := 0; i < limit; i++ {
			o := sortedOwners[i]
			ownershipRows = append(ownershipRows, []string{
				o.Name,
				fmt.Sprintf("%.1f", o.OwnershipScore),
				fmt.Sprintf("%.1f", o.ScoreBreakdown.CommitScore),
				fmt.Sprintf("%.1f", o.ScoreBreakdown.ReviewScore),
				fmt.Sprintf("%.1f", o.ScoreBreakdown.RecencyScore),
				o.ActivityStatus,
			})
		}

		b.Table([]string{"Name", "Overall", "Commits", "Reviews", "Recency", "Status"}, ownershipRows)

		if len(sortedOwners) > limit {
			b.Paragraph(fmt.Sprintf("*...and %d more contributors*", len(sortedOwners)-limit))
		}
	}

	// Bus Factor Analysis
	b.Section(2, "Bus Factor Analysis")

	riskLevel := data.Summary.BusFactorRisk
	if riskLevel == "" {
		riskLevel = calculateBusFactorRisk(data.Summary.BusFactor)
	}

	b.Paragraph(fmt.Sprintf("**Bus Factor:** %d", data.Summary.BusFactor))
	b.Paragraph(fmt.Sprintf("**Risk Level:** %s", formatBusFactorRisk(riskLevel)))

	switch riskLevel {
	case "critical":
		b.Quote("CRITICAL: Knowledge is concentrated in 1-2 people. Loss of key contributor(s) would severely impact the project.")
	case "warning":
		b.Quote("WARNING: Knowledge is concentrated in a small group. Consider cross-training and documentation.")
	case "healthy":
		b.Quote("HEALTHY: Knowledge is well-distributed across the team.")
	}

	// CODEOWNERS Analysis
	if data.Summary.HasCodeowners && data.Findings.CodeownersAnalysis != nil {
		b.Section(2, "CODEOWNERS Analysis")

		ca := data.Findings.CodeownersAnalysis
		b.Paragraph(fmt.Sprintf("**File:** `%s`", ca.FilePath))
		b.Paragraph(fmt.Sprintf("**Rules:** %d", ca.RulesCount))
		b.Paragraph(fmt.Sprintf("**Coverage:** %.1f%%", ca.Coverage*100))

		// Validation Issues
		if len(ca.ValidationIssues) > 0 {
			b.Section(3, "Validation Issues")

			issueRows := [][]string{}
			for _, issue := range ca.ValidationIssues {
				issueRows = append(issueRows, []string{
					issue.ID,
					issue.Severity,
					issue.Category,
					fmt.Sprintf("Line %d", issue.Line),
					issue.Message,
				})
			}
			b.Table([]string{"ID", "Severity", "Category", "Location", "Message"}, issueRows)
		}

		// Recommendations
		if len(ca.Recommendations) > 0 {
			b.Section(3, "Recommendations")

			recList := []string{}
			for _, rec := range ca.Recommendations {
				recList = append(recList, fmt.Sprintf("[%s] %s - %s",
					strings.ToUpper(rec.Priority), rec.Type, rec.Message))
			}
			b.List(recList)
		}

		// Drift Analysis
		if ca.DriftAnalysis != nil && ca.DriftAnalysis.HasDrift {
			b.Section(3, "Ownership Drift")
			b.Paragraph(fmt.Sprintf("**Drift Score:** %.1f/100 (higher = more drift)", ca.DriftAnalysis.DriftScore))
			b.Paragraph("Files where declared owners (CODEOWNERS) differ from actual contributors:")

			if len(ca.DriftAnalysis.DriftDetails) > 0 {
				driftRows := [][]string{}
				limit := min(10, len(ca.DriftAnalysis.DriftDetails))
				for i := 0; i < limit; i++ {
					d := ca.DriftAnalysis.DriftDetails[i]
					driftRows = append(driftRows, []string{
						fmt.Sprintf("`%s`", d.Path),
						strings.Join(d.DeclaredOwners, ", "),
						strings.Join(d.ActualTopOwners, ", "),
						fmt.Sprintf("%.0f%%", d.OverlapScore*100),
					})
				}
				b.Table([]string{"File", "Declared Owners", "Actual Owners", "Match"}, driftRows)

				if len(ca.DriftAnalysis.DriftDetails) > limit {
					b.Paragraph(fmt.Sprintf("*...and %d more drifted files*", len(ca.DriftAnalysis.DriftDetails)-limit))
				}
			}
		}
	}

	// Orphaned Files
	if len(data.Findings.OrphanedFiles) > 0 {
		b.Section(2, "Orphaned Files")
		b.Paragraph(fmt.Sprintf("**Total:** %d files with no clear owner in the last %d days",
			len(data.Findings.OrphanedFiles), data.Summary.PeriodDays))

		if len(data.Findings.OrphanedFiles) <= 20 {
			orphanList := []string{}
			for _, f := range data.Findings.OrphanedFiles {
				orphanList = append(orphanList, fmt.Sprintf("`%s`", f))
			}
			b.List(orphanList)
		} else {
			// Show first 20
			orphanList := []string{}
			for i := 0; i < 20; i++ {
				orphanList = append(orphanList, fmt.Sprintf("`%s`", data.Findings.OrphanedFiles[i]))
			}
			b.List(orphanList)
			b.Paragraph(fmt.Sprintf("*...and %d more orphaned files*", len(data.Findings.OrphanedFiles)-20))
		}
	}

	// Monorepo Analysis
	if data.Findings.Monorepo != nil && data.Findings.Monorepo.IsMonorepo {
		b.Section(2, "Monorepo Analysis")

		mono := data.Findings.Monorepo
		b.Paragraph(fmt.Sprintf("**Type:** %s", mono.Type))
		b.Paragraph(fmt.Sprintf("**Config:** `%s`", mono.ConfigFile))
		b.Paragraph(fmt.Sprintf("**Workspaces:** %d", len(mono.Workspaces)))

		if len(mono.Workspaces) > 0 {
			b.Section(3, "Workspace Ownership")

			wsRows := [][]string{}
			for _, ws := range mono.Workspaces {
				topOwners := ""
				if len(ws.TopContributors) > 0 {
					ownerNames := []string{}
					limit := min(3, len(ws.TopContributors))
					for i := 0; i < limit; i++ {
						ownerNames = append(ownerNames, ws.TopContributors[i].Name)
					}
					topOwners = strings.Join(ownerNames, ", ")
				}

				wsRows = append(wsRows, []string{
					ws.Name,
					fmt.Sprintf("`%s`", ws.Path),
					fmt.Sprintf("%d (%s)", ws.BusFactor, ws.BusFactorRisk),
					topOwners,
				})
			}

			b.Table([]string{"Workspace", "Path", "Bus Factor", "Top Contributors"}, wsRows)
		}

		if len(mono.CrossWorkspaceOwners) > 0 {
			b.Section(3, "Cross-Workspace Contributors")
			b.Paragraph("Contributors working across multiple workspaces:")
			b.List(mono.CrossWorkspaceOwners)
		}
	}

	// Developer Competencies
	if len(data.Findings.Competencies) > 0 {
		b.Section(2, "Developer Competencies")
		b.Paragraph("Language expertise and contribution patterns.")

		compRows := [][]string{}
		sortedComps := sortDeveloperProfilesByScore(data.Findings.Competencies)
		limit := min(10, len(sortedComps))

		for i := 0; i < limit; i++ {
			p := sortedComps[i]
			langs := []string{}
			langLimit := min(3, len(p.Languages))
			for j := 0; j < langLimit; j++ {
				langs = append(langs, p.Languages[j].Language)
			}
			langsStr := strings.Join(langs, ", ")

			compRows = append(compRows, []string{
				p.Name,
				p.TopLanguage,
				langsStr,
				fmt.Sprintf("%d", p.TotalCommits),
				fmt.Sprintf("%.1f", p.CompetencyScore),
			})
		}

		b.Table([]string{"Developer", "Top Language", "Languages", "Commits", "Score"}, compRows)

		if len(sortedComps) > limit {
			b.Paragraph(fmt.Sprintf("*...and %d more developers*", len(sortedComps)-limit))
		}
	}

	// Incident Contacts
	if len(data.Findings.IncidentContacts) > 0 {
		b.Section(2, "Incident Response Contacts")
		b.Paragraph("Recommended contacts for specific paths or components.")

		contactRows := [][]string{}
		limit := min(10, len(data.Findings.IncidentContacts))

		for i := 0; i < limit; i++ {
			ic := data.Findings.IncidentContacts[i]
			primary := ""
			if len(ic.Primary) > 0 {
				primary = fmt.Sprintf("%s (%.1f)", ic.Primary[0].Name, ic.Primary[0].ExpertiseScore)
			}
			backup := ""
			if len(ic.Backup) > 0 {
				backup = fmt.Sprintf("%s (%.1f)", ic.Backup[0].Name, ic.Backup[0].ExpertiseScore)
			}

			contactRows = append(contactRows, []string{
				fmt.Sprintf("`%s`", ic.Path),
				primary,
				backup,
			})
		}

		b.Table([]string{"Path", "Primary Contact", "Backup Contact"}, contactRows)

		if len(data.Findings.IncidentContacts) > limit {
			b.Paragraph(fmt.Sprintf("*...and %d more paths*", len(data.Findings.IncidentContacts)-limit))
		}
	}

	// PR Analysis
	if data.Findings.PRAnalysis != nil && !data.Findings.PRAnalysis.Skipped {
		b.Section(2, "Pull Request Review Activity")

		pra := data.Findings.PRAnalysis
		b.Paragraph(fmt.Sprintf("**PRs Analyzed:** %d", pra.PRsAnalyzed))

		if len(pra.Reviewers) > 0 {
			reviewerRows := [][]string{}
			sortedReviewers := sortPRReviewersByReviews(pra.Reviewers)
			limit := min(10, len(sortedReviewers))

			for i := 0; i < limit; i++ {
				r := sortedReviewers[i]
				reviewerRows = append(reviewerRows, []string{
					r.Name,
					fmt.Sprintf("%d", r.ReviewsGiven),
					fmt.Sprintf("%d", r.ApprovalsGiven),
					fmt.Sprintf("%d", r.CommentsGiven),
					fmt.Sprintf("%d", r.FilesReviewed),
				})
			}

			b.Table([]string{"Reviewer", "Reviews", "Approvals", "Comments", "Files"}, reviewerRows)

			if len(sortedReviewers) > limit {
				b.Paragraph(fmt.Sprintf("*...and %d more reviewers*", len(sortedReviewers)-limit))
			}
		}
	}

	// Warnings and Errors
	if len(data.Summary.Warnings) > 0 || len(data.Summary.Errors) > 0 {
		b.Section(2, "Analysis Notes")

		if len(data.Summary.Warnings) > 0 {
			b.Section(3, "Warnings")
			b.List(data.Summary.Warnings)
		}

		if len(data.Summary.Errors) > 0 {
			b.Section(3, "Errors")
			b.List(data.Summary.Errors)
		}
	}

	b.Footer("Code Ownership")

	return b.String()
}

// GenerateExecutiveReport creates a high-level summary for engineering leaders
func GenerateExecutiveReport(data *ReportData) string {
	b := report.NewBuilder()

	b.Title("Code Ownership Executive Report")
	b.Raw(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	b.Raw(fmt.Sprintf("**Date:** %s\n\n", data.Timestamp.Format("January 2, 2006")))

	// Executive Summary
	b.Section(2, "Executive Summary")

	busFactor := data.Summary.BusFactor
	riskLevel := data.Summary.BusFactorRisk
	if riskLevel == "" {
		riskLevel = calculateBusFactorRisk(busFactor)
	}

	// Overall health indicator
	healthStatus := assessOverallHealth(data)
	b.Paragraph(fmt.Sprintf("**Team Health Status:** %s", healthStatus.status))
	b.Newline()

	// Key metrics table
	summaryRows := [][]string{
		{"Active Contributors", fmt.Sprintf("%d", data.Summary.TotalContributors)},
		{"Bus Factor", formatBusFactor(busFactor, riskLevel)},
		{"Ownership Coverage", fmt.Sprintf("%.0f%%", data.Summary.OwnershipCoverage*100)},
		{"Files Without Owners", fmt.Sprintf("%d", data.Summary.OrphanedFiles)},
	}

	if data.Summary.HasCodeowners {
		summaryRows = append(summaryRows, []string{"CODEOWNERS Status", formatCodeownersStatus(data)})
	} else {
		summaryRows = append(summaryRows, []string{"CODEOWNERS Status", "Not configured"})
	}

	b.Table([]string{"Metric", "Value"}, summaryRows)

	// Team Health Analysis
	b.Section(2, "Team Health Analysis")

	findings := assessTeamHealth(data)

	if len(findings.critical) > 0 {
		b.Section(3, "Critical Issues")
		b.List(findings.critical)
	}

	if len(findings.warnings) > 0 {
		b.Section(3, "Areas of Concern")
		b.List(findings.warnings)
	}

	if len(findings.strengths) > 0 {
		b.Section(3, "Strengths")
		b.List(findings.strengths)
	}

	// Risk Assessment
	b.Section(2, "Risk Assessment")

	risks := assessRisks(data)
	b.Paragraph(fmt.Sprintf("**Knowledge Loss Risk:** %s", risks.knowledgeLoss))
	b.Paragraph(fmt.Sprintf("**Maintenance Risk:** %s", risks.maintenance))
	b.Paragraph(fmt.Sprintf("**Incident Response Readiness:** %s", risks.incidentResponse))
	b.Newline()

	if risks.keyRisk != "" {
		b.Quote(fmt.Sprintf("**Key Risk:** %s", risks.keyRisk))
	}

	// Contributor Distribution
	if len(data.Findings.Contributors) > 0 {
		b.Section(2, "Contributor Distribution")

		distro := analyzeContributorDistribution(data.Findings.Contributors)
		b.Paragraph(fmt.Sprintf("**Total Contributors:** %d (last %d days)", len(data.Findings.Contributors), data.Summary.PeriodDays))
		b.Paragraph(fmt.Sprintf("**Active Contributors:** %d", distro.active))
		b.Paragraph(fmt.Sprintf("**Top 20%% Contributors Account For:** %.0f%% of commits", distro.top20Percent*100))

		if distro.concentration > 0.7 {
			b.Quote("HIGH CONCENTRATION: The top contributors handle the majority of changes. Consider knowledge sharing and cross-training.")
		}
	}

	// Code Ownership Coverage
	b.Section(2, "Ownership Coverage")

	coverage := data.Summary.OwnershipCoverage
	b.Paragraph(fmt.Sprintf("**Files with Clear Owners:** %.0f%%", coverage*100))

	switch {
	case coverage >= 0.9:
		b.Paragraph("Excellent ownership coverage. Most files have clear maintainers.")
	case coverage >= 0.7:
		b.Paragraph("Good ownership coverage, but some files lack clear owners.")
	case coverage >= 0.5:
		b.Paragraph("Moderate ownership coverage. Many files need owner assignment.")
	default:
		b.Paragraph("Low ownership coverage. Significant portion of codebase lacks clear owners.")
	}

	if data.Summary.OrphanedFiles > 0 {
		b.Paragraph(fmt.Sprintf("\n**Orphaned Files:** %d files have no recent contributions or clear owners.", data.Summary.OrphanedFiles))
	}

	// Monorepo Insights
	if data.Findings.Monorepo != nil && data.Findings.Monorepo.IsMonorepo {
		b.Section(2, "Monorepo Structure")

		mono := data.Findings.Monorepo
		b.Paragraph(fmt.Sprintf("**Workspaces:** %d", len(mono.Workspaces)))

		// Identify high-risk workspaces
		highRiskWs := []string{}
		for _, ws := range mono.Workspaces {
			if ws.BusFactorRisk == "critical" {
				highRiskWs = append(highRiskWs, fmt.Sprintf("%s (bus factor: %d)", ws.Name, ws.BusFactor))
			}
		}

		if len(highRiskWs) > 0 {
			b.Section(3, "High-Risk Workspaces")
			b.List(highRiskWs)
		}

		if len(mono.CrossWorkspaceOwners) > 0 {
			b.Paragraph(fmt.Sprintf("\n**Cross-Workspace Contributors:** %d developers work across multiple workspaces, providing valuable integration knowledge.",
				len(mono.CrossWorkspaceOwners)))
		}
	}

	// Recommendations
	b.Section(2, "Recommendations")

	recommendations := generateRecommendations(data)

	if len(recommendations.immediate) > 0 {
		b.Section(3, "Immediate Actions")
		b.NumberedList(recommendations.immediate)
	}

	if len(recommendations.shortTerm) > 0 {
		b.Section(3, "Short-term Improvements")
		b.NumberedList(recommendations.shortTerm)
	}

	if len(recommendations.longTerm) > 0 {
		b.Section(3, "Long-term Strategy")
		b.NumberedList(recommendations.longTerm)
	}

	// Business Impact
	b.Section(2, "Business Impact")

	impact := assessBusinessImpact(data)
	b.Paragraph(fmt.Sprintf("**Estimated Impact of Key Person Loss:** %s", impact.keyPersonLoss))
	b.Paragraph(fmt.Sprintf("**Team Scalability:** %s", impact.scalability))
	b.Paragraph(fmt.Sprintf("**Knowledge Transfer Needs:** %s", impact.knowledgeTransfer))

	b.Footer("Code Ownership")

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
	techPath := filepath.Join(analysisDir, "code-ownership-technical-report.md")
	if err := os.WriteFile(techPath, []byte(techReport), 0644); err != nil {
		return fmt.Errorf("writing technical report: %w", err)
	}

	// Write executive report
	execReport := GenerateExecutiveReport(data)
	execPath := filepath.Join(analysisDir, "code-ownership-executive-report.md")
	if err := os.WriteFile(execPath, []byte(execReport), 0644); err != nil {
		return fmt.Errorf("writing executive report: %w", err)
	}

	return nil
}

// Helper functions

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func formatBusFactor(bf int, risk string) string {
	if risk == "" {
		return fmt.Sprintf("%d", bf)
	}
	return fmt.Sprintf("%d (%s)", bf, risk)
}

func formatActivityStatus(status string) string {
	statusMap := map[string]string{
		"active":    "Active (recent commits)",
		"recent":    "Recent (commits within last month)",
		"stale":     "Stale (no commits in 1-3 months)",
		"inactive":  "Inactive (no commits in 3-6 months)",
		"abandoned": "Abandoned (no commits in 6+ months)",
	}
	if mapped, ok := statusMap[status]; ok {
		return mapped
	}
	return strings.Title(status)
}

func formatBusFactorRisk(risk string) string {
	riskMap := map[string]string{
		"critical": "CRITICAL - High concentration of knowledge",
		"warning":  "WARNING - Moderate knowledge concentration",
		"healthy":  "HEALTHY - Good knowledge distribution",
	}
	if mapped, ok := riskMap[risk]; ok {
		return mapped
	}
	return strings.ToUpper(risk)
}

func calculateBusFactorRisk(busFactor int) string {
	switch {
	case busFactor <= 2:
		return "critical"
	case busFactor <= 5:
		return "warning"
	default:
		return "healthy"
	}
}

func formatCodeownersStatus(data *ReportData) string {
	if !data.Summary.HasCodeowners {
		return "Not configured"
	}

	issues := data.Summary.CodeownersIssues
	coverage := data.Summary.OwnershipCoverage

	if issues == 0 && coverage >= 0.8 {
		return fmt.Sprintf("Configured (%.0f%% coverage)", coverage*100)
	} else if issues > 0 {
		return fmt.Sprintf("Configured (%d issues, %.0f%% coverage)", issues, coverage*100)
	}
	return fmt.Sprintf("Configured (%.0f%% coverage)", coverage*100)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Sorting functions

func sortContributorsByCommits(contributors []Contributor) []Contributor {
	sorted := make([]Contributor, len(contributors))
	copy(sorted, contributors)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Commits > sorted[j].Commits
	})
	return sorted
}

func sortEnhancedOwnershipByScore(owners []EnhancedOwnership) []EnhancedOwnership {
	sorted := make([]EnhancedOwnership, len(owners))
	copy(sorted, owners)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].OwnershipScore > sorted[j].OwnershipScore
	})
	return sorted
}

func sortDeveloperProfilesByScore(profiles []DeveloperProfile) []DeveloperProfile {
	sorted := make([]DeveloperProfile, len(profiles))
	copy(sorted, profiles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CompetencyScore > sorted[j].CompetencyScore
	})
	return sorted
}

func sortPRReviewersByReviews(reviewers []PRReviewer) []PRReviewer {
	sorted := make([]PRReviewer, len(reviewers))
	copy(sorted, reviewers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ReviewsGiven > sorted[j].ReviewsGiven
	})
	return sorted
}

// Assessment functions

type healthStatus struct {
	status string
	score  int
}

func assessOverallHealth(data *ReportData) healthStatus {
	score := 100

	// Bus factor penalty
	bf := data.Summary.BusFactor
	switch {
	case bf <= 2:
		score -= 30
	case bf <= 5:
		score -= 15
	}

	// Ownership coverage penalty
	coverage := data.Summary.OwnershipCoverage
	if coverage < 0.5 {
		score -= 20
	} else if coverage < 0.7 {
		score -= 10
	}

	// Orphaned files penalty
	orphanPercent := float64(data.Summary.OrphanedFiles) / float64(data.Summary.FilesAnalyzed)
	if orphanPercent > 0.2 {
		score -= 15
	} else if orphanPercent > 0.1 {
		score -= 8
	}

	// CODEOWNERS issues penalty
	if data.Summary.CodeownersIssues > 10 {
		score -= 10
	} else if data.Summary.CodeownersIssues > 5 {
		score -= 5
	}

	status := ""
	switch {
	case score >= 85:
		status = "Excellent"
	case score >= 70:
		status = "Good"
	case score >= 55:
		status = "Fair"
	case score >= 40:
		status = "Needs Improvement"
	default:
		status = "Critical"
	}

	return healthStatus{status: status, score: score}
}

type teamHealthFindings struct {
	critical  []string
	warnings  []string
	strengths []string
}

func assessTeamHealth(data *ReportData) teamHealthFindings {
	findings := teamHealthFindings{}

	// Bus factor issues
	bf := data.Summary.BusFactor
	if bf <= 1 {
		findings.critical = append(findings.critical,
			"Single point of failure: Only 1 person has significant knowledge of the codebase")
	} else if bf == 2 {
		findings.critical = append(findings.critical,
			"Very high knowledge concentration: Only 2 people have deep codebase knowledge")
	} else if bf <= 5 {
		findings.warnings = append(findings.warnings,
			fmt.Sprintf("Moderate knowledge concentration: Bus factor of %d indicates knowledge is concentrated in a small group", bf))
	} else {
		findings.strengths = append(findings.strengths,
			fmt.Sprintf("Healthy knowledge distribution: Bus factor of %d indicates good team coverage", bf))
	}

	// Ownership coverage issues
	coverage := data.Summary.OwnershipCoverage
	if coverage < 0.5 {
		findings.critical = append(findings.critical,
			fmt.Sprintf("Low ownership clarity: Only %.0f%% of files have clear owners", coverage*100))
	} else if coverage < 0.7 {
		findings.warnings = append(findings.warnings,
			fmt.Sprintf("Moderate ownership gaps: %.0f%% of files have clear owners", coverage*100))
	} else {
		findings.strengths = append(findings.strengths,
			fmt.Sprintf("Good ownership clarity: %.0f%% of files have clear owners", coverage*100))
	}

	// Orphaned files
	orphanPercent := float64(data.Summary.OrphanedFiles) / float64(data.Summary.FilesAnalyzed)
	if orphanPercent > 0.2 {
		findings.critical = append(findings.critical,
			fmt.Sprintf("High number of orphaned files: %d files (%.0f%%) lack recent contributions",
				data.Summary.OrphanedFiles, orphanPercent*100))
	} else if orphanPercent > 0.1 {
		findings.warnings = append(findings.warnings,
			fmt.Sprintf("Some orphaned files: %d files (%.0f%%) lack recent contributions",
				data.Summary.OrphanedFiles, orphanPercent*100))
	}

	// CODEOWNERS status
	if !data.Summary.HasCodeowners {
		findings.warnings = append(findings.warnings,
			"No CODEOWNERS file: Missing formal ownership documentation and PR routing")
	} else if data.Summary.CodeownersIssues > 0 {
		findings.warnings = append(findings.warnings,
			fmt.Sprintf("CODEOWNERS has %d validation issues requiring attention", data.Summary.CodeownersIssues))
	} else {
		findings.strengths = append(findings.strengths,
			"CODEOWNERS file is configured and validated")
	}

	// Activity status
	if data.Summary.RepoActivityStatus == "stale" || data.Summary.RepoActivityStatus == "inactive" {
		findings.warnings = append(findings.warnings,
			fmt.Sprintf("Repository activity is %s: %d days since last commit",
				data.Summary.RepoActivityStatus, data.Summary.DaysSinceLastCommit))
	} else if data.Summary.RepoActivityStatus == "abandoned" {
		findings.critical = append(findings.critical,
			fmt.Sprintf("Repository appears abandoned: %d days since last commit", data.Summary.DaysSinceLastCommit))
	}

	// Drift analysis
	if data.Findings.CodeownersAnalysis != nil && data.Findings.CodeownersAnalysis.DriftAnalysis != nil {
		drift := data.Findings.CodeownersAnalysis.DriftAnalysis
		if drift.HasDrift && drift.DriftScore > 50 {
			findings.warnings = append(findings.warnings,
				fmt.Sprintf("Significant ownership drift: CODEOWNERS differs from actual contributors (drift score: %.0f)", drift.DriftScore))
		}
	}

	return findings
}

type riskAssessment struct {
	knowledgeLoss    string
	maintenance      string
	incidentResponse string
	keyRisk          string
}

func assessRisks(data *ReportData) riskAssessment {
	risks := riskAssessment{}

	// Knowledge loss risk
	bf := data.Summary.BusFactor
	switch {
	case bf <= 2:
		risks.knowledgeLoss = "CRITICAL - Loss of 1-2 people would severely impact the project"
	case bf <= 5:
		risks.knowledgeLoss = "HIGH - Loss of key contributors would impact development velocity"
	case bf <= 10:
		risks.knowledgeLoss = "MODERATE - Team has some redundancy but knowledge could be better distributed"
	default:
		risks.knowledgeLoss = "LOW - Knowledge is well-distributed across the team"
	}

	// Maintenance risk
	orphanPercent := float64(data.Summary.OrphanedFiles) / float64(data.Summary.FilesAnalyzed)
	coverage := data.Summary.OwnershipCoverage
	switch {
	case orphanPercent > 0.2 || coverage < 0.5:
		risks.maintenance = "HIGH - Many files lack clear owners or recent maintenance"
	case orphanPercent > 0.1 || coverage < 0.7:
		risks.maintenance = "MODERATE - Some files lack clear ownership"
	default:
		risks.maintenance = "LOW - Most files have clear owners and active maintenance"
	}

	// Incident response readiness
	hasIncidentContacts := len(data.Findings.IncidentContacts) > 0
	hasCodeowners := data.Summary.HasCodeowners
	switch {
	case !hasCodeowners && !hasIncidentContacts:
		risks.incidentResponse = "LOW - No formal ownership or contact information available"
	case hasCodeowners && data.Summary.CodeownersIssues > 5:
		risks.incidentResponse = "MODERATE - CODEOWNERS exists but has validation issues"
	case hasCodeowners || hasIncidentContacts:
		risks.incidentResponse = "GOOD - Ownership information available for incident routing"
	}

	// Key risk identification
	if bf <= 2 {
		risks.keyRisk = "Critical knowledge concentration: Loss of 1-2 key people would severely impact development capacity and project continuity"
	} else if orphanPercent > 0.3 {
		risks.keyRisk = "High technical debt: Large number of orphaned files indicates maintenance challenges and unclear ownership"
	} else if !hasCodeowners && coverage < 0.6 {
		risks.keyRisk = "Unclear ownership: Lack of formal ownership documentation may slow incident response and maintenance"
	}

	return risks
}

type contributorDistribution struct {
	active        int
	top20Percent  float64
	concentration float64
}

func analyzeContributorDistribution(contributors []Contributor) contributorDistribution {
	if len(contributors) == 0 {
		return contributorDistribution{}
	}

	sorted := sortContributorsByCommits(contributors)

	totalCommits := 0
	for _, c := range sorted {
		totalCommits += c.Commits
	}

	// Calculate top 20% contribution
	top20Count := max(1, len(sorted)/5)
	top20Commits := 0
	for i := 0; i < top20Count; i++ {
		top20Commits += sorted[i].Commits
	}

	concentration := float64(top20Commits) / float64(totalCommits)

	return contributorDistribution{
		active:        len(contributors),
		top20Percent:  concentration,
		concentration: concentration,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type recommendationSet struct {
	immediate []string
	shortTerm []string
	longTerm  []string
}

func generateRecommendations(data *ReportData) recommendationSet {
	rec := recommendationSet{}

	// Bus factor recommendations
	bf := data.Summary.BusFactor
	if bf <= 2 {
		rec.immediate = append(rec.immediate,
			"Implement immediate knowledge sharing sessions for critical components")
		rec.immediate = append(rec.immediate,
			"Document key architectural decisions and system knowledge")
		rec.shortTerm = append(rec.shortTerm,
			"Establish pair programming or code review practices to distribute knowledge")
	} else if bf <= 5 {
		rec.shortTerm = append(rec.shortTerm,
			"Rotate responsibilities to increase knowledge distribution across the team")
	}

	// CODEOWNERS recommendations
	if !data.Summary.HasCodeowners {
		rec.immediate = append(rec.immediate,
			"Create CODEOWNERS file to formalize ownership and enable automated PR routing")
	} else if data.Summary.CodeownersIssues > 0 {
		rec.immediate = append(rec.immediate,
			fmt.Sprintf("Fix %d CODEOWNERS validation issues", data.Summary.CodeownersIssues))
	}

	// Ownership drift recommendations
	if data.Findings.CodeownersAnalysis != nil && data.Findings.CodeownersAnalysis.DriftAnalysis != nil {
		drift := data.Findings.CodeownersAnalysis.DriftAnalysis
		if drift.HasDrift && drift.DriftScore > 50 {
			rec.shortTerm = append(rec.shortTerm,
				"Update CODEOWNERS to reflect actual contributors and current team structure")
		}
	}

	// Orphaned files recommendations
	orphanPercent := float64(data.Summary.OrphanedFiles) / float64(data.Summary.FilesAnalyzed)
	if orphanPercent > 0.2 {
		rec.immediate = append(rec.immediate,
			fmt.Sprintf("Assign owners to %d orphaned files or determine if they can be removed", data.Summary.OrphanedFiles))
	} else if orphanPercent > 0.1 {
		rec.shortTerm = append(rec.shortTerm,
			"Review and assign owners to orphaned files")
	}

	// Coverage recommendations
	if data.Summary.OwnershipCoverage < 0.7 {
		rec.shortTerm = append(rec.shortTerm,
			"Improve ownership coverage by assigning clear owners to key components")
	}

	// Monorepo recommendations
	if data.Findings.Monorepo != nil && data.Findings.Monorepo.IsMonorepo {
		highRiskCount := 0
		for _, ws := range data.Findings.Monorepo.Workspaces {
			if ws.BusFactorRisk == "critical" {
				highRiskCount++
			}
		}
		if highRiskCount > 0 {
			rec.immediate = append(rec.immediate,
				fmt.Sprintf("Address critical bus factor in %d workspace(s)", highRiskCount))
		}
	}

	// Long-term strategic recommendations
	rec.longTerm = append(rec.longTerm,
		"Establish regular rotation of on-call and maintenance responsibilities")
	rec.longTerm = append(rec.longTerm,
		"Create mentorship programs to build depth in critical areas")
	rec.longTerm = append(rec.longTerm,
		"Monitor ownership metrics quarterly to track improvements")

	if len(data.Findings.IncidentContacts) == 0 {
		rec.longTerm = append(rec.longTerm,
			"Develop incident response contacts and escalation paths")
	}

	return rec
}

type businessImpact struct {
	keyPersonLoss     string
	scalability       string
	knowledgeTransfer string
}

func assessBusinessImpact(data *ReportData) businessImpact {
	impact := businessImpact{}

	// Key person loss impact
	bf := data.Summary.BusFactor
	switch {
	case bf <= 1:
		impact.keyPersonLoss = "SEVERE - Project would face significant delays or potential failure"
	case bf <= 2:
		impact.keyPersonLoss = "HIGH - Development velocity would drop significantly (50-70%)"
	case bf <= 5:
		impact.keyPersonLoss = "MODERATE - Team would experience slowdowns but could recover (20-40% impact)"
	default:
		impact.keyPersonLoss = "LOW - Team has sufficient redundancy (< 20% impact)"
	}

	// Team scalability
	coverage := data.Summary.OwnershipCoverage
	hasCodeowners := data.Summary.HasCodeowners
	switch {
	case coverage < 0.5 && !hasCodeowners:
		impact.scalability = "POOR - Unclear ownership will hinder onboarding and scaling"
	case coverage < 0.7:
		impact.scalability = "FAIR - Some ownership clarity but gaps exist"
	case coverage >= 0.8 && hasCodeowners:
		impact.scalability = "GOOD - Clear ownership supports efficient scaling"
	default:
		impact.scalability = "MODERATE - Reasonable ownership structure"
	}

	// Knowledge transfer needs
	distro := analyzeContributorDistribution(data.Findings.Contributors)
	switch {
	case distro.concentration > 0.8:
		impact.knowledgeTransfer = "URGENT - Knowledge highly concentrated, immediate documentation and training needed"
	case distro.concentration > 0.6:
		impact.knowledgeTransfer = "HIGH - Should prioritize knowledge sharing and documentation"
	case distro.concentration > 0.4:
		impact.knowledgeTransfer = "MODERATE - Some knowledge sharing needed"
	default:
		impact.knowledgeTransfer = "LOW - Knowledge is reasonably distributed"
	}

	return impact
}
