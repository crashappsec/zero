package codequality

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// LoadReportData loads code-quality.json from the analysis directory
func LoadReportData(analysisDir string) (*ReportData, error) {
	qualityPath := filepath.Join(analysisDir, "code-quality.json")
	data, err := os.ReadFile(qualityPath)
	if err != nil {
		return nil, fmt.Errorf("reading code-quality.json: %w", err)
	}

	var result struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing code-quality.json: %w", err)
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

	// Title and metadata
	b.Title("Code Quality Technical Report")
	b.Meta(report.ReportMeta{
		Repository:  data.Repository,
		Timestamp:   data.Timestamp,
		ScannerDesc: "Code quality analysis: technical debt, complexity, test coverage, documentation",
	})

	// Technical Debt Section
	if data.Summary.TechDebt != nil {
		b.Section(2, "1. Technical Debt Analysis")
		td := data.Summary.TechDebt

		if td.Error != "" {
			b.Paragraph(fmt.Sprintf("Error: %s", td.Error))
		} else {
			b.Section(3, "Summary")

			rows := [][]string{
				{"Total Debt Markers", fmt.Sprintf("%d", td.TotalMarkers)},
				{"Total Code Issues", fmt.Sprintf("%d", td.TotalIssues)},
				{"Files Affected", fmt.Sprintf("%d", td.FilesAffected)},
			}
			b.Table([]string{"Metric", "Value"}, rows)

			// By type breakdown
			if len(td.ByType) > 0 {
				b.Section(3, "Markers by Type")
				typeRows := [][]string{}
				for typ, count := range td.ByType {
					typeRows = append(typeRows, []string{typ, fmt.Sprintf("%d", count)})
				}
				b.Table([]string{"Type", "Count"}, typeRows)
			}

			// By priority breakdown
			if len(td.ByPriority) > 0 {
				b.Section(3, "Markers by Priority")
				priorityRows := [][]string{}
				for priority, count := range td.ByPriority {
					priorityRows = append(priorityRows, []string{strings.ToUpper(priority), fmt.Sprintf("%d", count)})
				}
				b.Table([]string{"Priority", "Count"}, priorityRows)
			}

			// Hotspots
			if data.Findings.TechDebt != nil && len(data.Findings.TechDebt.Hotspots) > 0 {
				b.Section(3, "Technical Debt Hotspots")
				b.Paragraph("Files with the most technical debt markers:")

				hotspotRows := [][]string{}
				for i, hs := range data.Findings.TechDebt.Hotspots {
					if i >= 20 {
						break
					}
					types := []string{}
					for t, c := range hs.ByType {
						types = append(types, fmt.Sprintf("%s: %d", t, c))
					}
					hotspotRows = append(hotspotRows, []string{
						hs.File,
						fmt.Sprintf("%d", hs.TotalMarkers),
						strings.Join(types, ", "),
					})
				}
				b.Table([]string{"File", "Total Markers", "Breakdown"}, hotspotRows)
			}

			// High priority markers
			if data.Findings.TechDebt != nil && len(data.Findings.TechDebt.Markers) > 0 {
				highPriorityMarkers := []DebtMarker{}
				for _, m := range data.Findings.TechDebt.Markers {
					if m.Priority == "high" {
						highPriorityMarkers = append(highPriorityMarkers, m)
					}
				}

				if len(highPriorityMarkers) > 0 {
					b.Section(3, "High Priority Markers")
					b.Paragraph(fmt.Sprintf("Found %d high-priority debt markers requiring immediate attention:", len(highPriorityMarkers)))

					markerRows := [][]string{}
					for i, m := range highPriorityMarkers {
						if i >= 50 {
							b.Paragraph(fmt.Sprintf("...and %d more high-priority markers", len(highPriorityMarkers)-50))
							break
						}
						text := m.Text
						if len(text) > 80 {
							text = text[:77] + "..."
						}
						markerRows = append(markerRows, []string{
							m.Type,
							fmt.Sprintf("%s:%d", m.File, m.Line),
							text,
						})
					}
					b.Table([]string{"Type", "Location", "Description"}, markerRows)
				}
			}

			// Code quality issues
			if data.Findings.TechDebt != nil && len(data.Findings.TechDebt.Issues) > 0 {
				b.Section(3, "Code Quality Issues")

				// Group by severity
				highIssues := []DebtIssue{}
				mediumIssues := []DebtIssue{}
				for _, issue := range data.Findings.TechDebt.Issues {
					if issue.Severity == "high" {
						highIssues = append(highIssues, issue)
					} else if issue.Severity == "medium" {
						mediumIssues = append(mediumIssues, issue)
					}
				}

				if len(highIssues) > 0 {
					b.Section(4, "High Severity Issues")
					issueRows := [][]string{}
					for i, issue := range highIssues {
						if i >= 30 {
							b.Paragraph(fmt.Sprintf("...and %d more high-severity issues", len(highIssues)-30))
							break
						}
						issueRows = append(issueRows, []string{
							issue.Type,
							fmt.Sprintf("%s:%d", issue.File, issue.Line),
							issue.Description,
							issue.Suggestion,
						})
					}
					b.Table([]string{"Type", "Location", "Issue", "Suggestion"}, issueRows)
				}

				if len(mediumIssues) > 0 {
					b.Section(4, "Medium Severity Issues")
					b.Paragraph(fmt.Sprintf("Found %d medium-severity code quality issues. Top items:", len(mediumIssues)))

					issueRows := [][]string{}
					for i, issue := range mediumIssues {
						if i >= 20 {
							b.Paragraph(fmt.Sprintf("...and %d more medium-severity issues", len(mediumIssues)-20))
							break
						}
						issueRows = append(issueRows, []string{
							issue.Type,
							fmt.Sprintf("%s:%d", issue.File, issue.Line),
							issue.Description,
						})
					}
					b.Table([]string{"Type", "Location", "Issue"}, issueRows)
				}
			}
		}
	}

	// Complexity Section
	if data.Summary.Complexity != nil {
		b.Section(2, "2. Complexity Analysis")
		c := data.Summary.Complexity

		if c.Error != "" {
			b.Paragraph(fmt.Sprintf("Error: %s", c.Error))
		} else {
			b.Section(3, "Summary")

			rows := [][]string{
				{"Total Complexity Issues", fmt.Sprintf("%d", c.TotalIssues)},
				{"High Severity", fmt.Sprintf("%d", c.High)},
				{"Medium Severity", fmt.Sprintf("%d", c.Medium)},
				{"Low Severity", fmt.Sprintf("%d", c.Low)},
				{"Files Affected", fmt.Sprintf("%d", c.FilesAffected)},
			}
			b.Table([]string{"Metric", "Value"}, rows)

			// Maintainability assessment
			assessment := assessMaintainability(c.TotalIssues, c.High, c.Medium, c.FilesAffected)
			b.Section(3, "Maintainability Assessment")
			b.Paragraph(fmt.Sprintf("Status: %s", b.Bold(assessment.status)))
			b.Paragraph(assessment.description)

			// By type breakdown
			if len(c.ByType) > 0 {
				b.Section(3, "Issues by Type")
				typeRows := [][]string{}
				for typ, count := range c.ByType {
					typeRows = append(typeRows, []string{formatComplexityType(typ), fmt.Sprintf("%d", count)})
				}
				b.Table([]string{"Type", "Count"}, typeRows)
			}

			// Detailed findings
			if data.Findings.Complexity != nil && len(data.Findings.Complexity.Issues) > 0 {
				b.Section(3, "Critical Complexity Issues")

				// Show high severity first
				highComplexity := []ComplexityIssue{}
				for _, issue := range data.Findings.Complexity.Issues {
					if issue.Severity == "high" {
						highComplexity = append(highComplexity, issue)
					}
				}

				if len(highComplexity) > 0 {
					b.Paragraph(fmt.Sprintf("Found %d high-severity complexity issues requiring refactoring:", len(highComplexity)))

					issueRows := [][]string{}
					for i, issue := range highComplexity {
						if i >= 25 {
							b.Paragraph(fmt.Sprintf("...and %d more high-severity complexity issues", len(highComplexity)-25))
							break
						}
						issueRows = append(issueRows, []string{
							formatComplexityType(issue.Type),
							fmt.Sprintf("%s:%d", issue.File, issue.Line),
							issue.Description,
							issue.Suggestion,
						})
					}
					b.Table([]string{"Type", "Location", "Issue", "Recommendation"}, issueRows)
				}
			}
		}
	}

	// Test Coverage Section
	if data.Summary.TestCoverage != nil {
		b.Section(2, "3. Test Coverage Analysis")
		tc := data.Summary.TestCoverage

		if tc.Error != "" {
			b.Paragraph(fmt.Sprintf("Error: %s", tc.Error))
		} else {
			b.Section(3, "Summary")

			rows := [][]string{
				{"Test Files Present", boolToYesNo(tc.HasTestFiles)},
				{"Test Frameworks", formatList(tc.TestFrameworks)},
				{"Coverage Reports", fmt.Sprintf("%d found", len(tc.CoverageReports))},
			}

			if tc.LineCoverage > 0 {
				coverageStatus := "Good"
				if tc.LineCoverage < 60 {
					coverageStatus = "Poor"
				} else if tc.LineCoverage < 80 {
					coverageStatus = "Fair"
				}
				rows = append(rows, []string{"Line Coverage", fmt.Sprintf("%.1f%% (%s)", tc.LineCoverage, coverageStatus)})
				rows = append(rows, []string{"Meets Threshold", boolToYesNo(tc.MeetsThreshold)})
			}

			b.Table([]string{"Metric", "Value"}, rows)

			// Coverage assessment
			if tc.HasTestFiles {
				b.Section(3, "Coverage Assessment")

				if tc.LineCoverage > 0 {
					if tc.LineCoverage >= 80 {
						b.Paragraph("Excellent test coverage! The codebase has strong test coverage above 80%.")
					} else if tc.LineCoverage >= 60 {
						b.Paragraph("Moderate test coverage. Consider improving coverage for critical paths.")
					} else {
						b.Paragraph("Low test coverage detected. Significant gaps in test coverage may impact code quality and reliability.")
					}
				} else {
					b.Paragraph("Test files detected but no coverage reports found. Consider generating coverage reports to track test effectiveness.")
				}

				// List frameworks
				if len(tc.TestFrameworks) > 0 {
					b.Section(4, "Detected Test Frameworks")
					b.List(tc.TestFrameworks)
				}

				// List coverage reports
				if len(tc.CoverageReports) > 0 {
					b.Section(4, "Coverage Reports")
					for _, report := range tc.CoverageReports {
						b.Paragraph(fmt.Sprintf("- %s", b.Code(report)))
					}
					b.Newline()
				}
			} else {
				b.Paragraph("No test files detected in the repository. Consider adding automated tests to improve code quality and catch regressions.")
			}
		}
	}

	// Documentation Section
	if data.Summary.CodeDocs != nil {
		b.Section(2, "4. Documentation Analysis")
		cd := data.Summary.CodeDocs

		if cd.Error != "" {
			b.Paragraph(fmt.Sprintf("Error: %s", cd.Error))
		} else {
			b.Section(3, "Summary")

			rows := [][]string{
				{"Documentation Score", fmt.Sprintf("%d/100 (%s)", cd.Score, scoreToGrade(cd.Score))},
				{"README", boolToYesNo(cd.HasReadme)},
				{"CHANGELOG", boolToYesNo(cd.HasChangelog)},
				{"API Documentation", boolToYesNo(cd.HasApiDocs)},
			}

			if cd.ReadmeFile != "" {
				rows = append(rows, []string{"README File", b.Code(cd.ReadmeFile)})
			}

			b.Table([]string{"Metric", "Value"}, rows)

			// Documentation assessment
			b.Section(3, "Documentation Assessment")

			if cd.Score >= 80 {
				b.Paragraph("Excellent documentation! The project has comprehensive documentation including README, CHANGELOG, and API docs.")
			} else if cd.Score >= 60 {
				b.Paragraph("Good documentation foundation. Consider enhancing with additional documentation types.")
			} else if cd.Score >= 40 {
				b.Paragraph("Basic documentation present. Significant improvements needed for developer onboarding and maintenance.")
			} else {
				b.Paragraph("Minimal documentation detected. Documentation is critical for maintainability and developer productivity.")
			}

			// Recommendations
			recommendations := []string{}
			if !cd.HasReadme {
				recommendations = append(recommendations, "Add a comprehensive README.md with project overview, installation, and usage instructions")
			}
			if !cd.HasChangelog {
				recommendations = append(recommendations, "Create a CHANGELOG.md to track version history and changes")
			}
			if !cd.HasApiDocs {
				recommendations = append(recommendations, "Add API documentation (OpenAPI/Swagger or docs directory)")
			}

			if len(recommendations) > 0 {
				b.Section(4, "Documentation Recommendations")
				b.List(recommendations)
			}
		}
	}

	b.Footer("Code Quality")

	return b.String()
}

// GenerateExecutiveReport creates a high-level summary for engineering leaders
func GenerateExecutiveReport(data *ReportData) string {
	b := report.NewBuilder()

	b.Title("Code Quality Executive Report")
	b.Paragraph(fmt.Sprintf("%s**Repository:** %s", b.Bold("Repository: "), b.Code(data.Repository)))
	b.Paragraph(fmt.Sprintf("%s %s", b.Bold("Date:"), data.Timestamp.Format("January 2, 2006")))
	b.Newline()

	// Executive Summary
	b.Section(2, "Executive Summary")

	overallScore := calculateOverallHealthScore(data.Summary)
	healthGrade := scoreToGrade(overallScore)
	healthStatus := scoreToStatus(overallScore)

	b.Section(3, fmt.Sprintf("Overall Code Health: %s (%d/100)", healthGrade, overallScore))
	b.Paragraph(fmt.Sprintf("Status: %s", b.Bold(healthStatus)))
	b.Newline()

	// Health metrics table
	b.Section(3, "Health Metrics")

	healthRows := [][]string{}

	if data.Summary.TechDebt != nil {
		score := calculateDebtScore(data.Summary.TechDebt)
		healthRows = append(healthRows, []string{
			"Technical Debt",
			fmt.Sprintf("%d/100", score),
			scoreToStatus(score),
		})
	}

	if data.Summary.Complexity != nil {
		score := calculateComplexityScore(data.Summary.Complexity)
		healthRows = append(healthRows, []string{
			"Complexity",
			fmt.Sprintf("%d/100", score),
			scoreToStatus(score),
		})
	}

	if data.Summary.TestCoverage != nil {
		score := calculateCoverageScore(data.Summary.TestCoverage)
		healthRows = append(healthRows, []string{
			"Test Coverage",
			fmt.Sprintf("%d/100", score),
			scoreToStatus(score),
		})
	}

	if data.Summary.CodeDocs != nil {
		healthRows = append(healthRows, []string{
			"Documentation",
			fmt.Sprintf("%d/100", data.Summary.CodeDocs.Score),
			scoreToStatus(data.Summary.CodeDocs.Score),
		})
	}

	b.Table([]string{"Area", "Score", "Status"}, healthRows)

	// Key Metrics
	b.Section(2, "Key Metrics")

	if data.Summary.TechDebt != nil {
		td := data.Summary.TechDebt
		b.Section(3, "Technical Debt")
		items := []string{
			fmt.Sprintf("%s %d debt markers across %d files", b.Bold("Total Markers:"), td.TotalMarkers, td.FilesAffected),
			fmt.Sprintf("%s %d code quality issues", b.Bold("Code Issues:"), td.TotalIssues),
		}
		if td.ByPriority != nil {
			if high, ok := td.ByPriority["high"]; ok && high > 0 {
				items = append(items, fmt.Sprintf("%s %d items requiring immediate attention", b.Bold("High Priority:"), high))
			}
		}
		b.List(items)
	}

	if data.Summary.Complexity != nil {
		c := data.Summary.Complexity
		b.Section(3, "Code Complexity")
		items := []string{
			fmt.Sprintf("%s %d complexity issues across %d files", b.Bold("Total Issues:"), c.TotalIssues, c.FilesAffected),
		}
		if c.High > 0 {
			items = append(items, fmt.Sprintf("%s %d functions require refactoring", b.Bold("High Complexity:"), c.High))
		}
		b.List(items)
	}

	if data.Summary.TestCoverage != nil {
		tc := data.Summary.TestCoverage
		b.Section(3, "Test Coverage")
		items := []string{}

		if tc.HasTestFiles {
			items = append(items, fmt.Sprintf("Test files present using: %s", strings.Join(tc.TestFrameworks, ", ")))
			if tc.LineCoverage > 0 {
				items = append(items, fmt.Sprintf("%s %.1f%%", b.Bold("Line Coverage:"), tc.LineCoverage))
			}
		} else {
			items = append(items, "No test files detected")
		}

		b.List(items)
	}

	if data.Summary.CodeDocs != nil {
		cd := data.Summary.CodeDocs
		b.Section(3, "Documentation")
		items := []string{
			fmt.Sprintf("%s %d/100", b.Bold("Documentation Score:"), cd.Score),
		}
		docTypes := []string{}
		if cd.HasReadme {
			docTypes = append(docTypes, "README")
		}
		if cd.HasChangelog {
			docTypes = append(docTypes, "CHANGELOG")
		}
		if cd.HasApiDocs {
			docTypes = append(docTypes, "API docs")
		}
		if len(docTypes) > 0 {
			items = append(items, fmt.Sprintf("Available: %s", strings.Join(docTypes, ", ")))
		}
		b.List(items)
	}

	// Critical Findings
	findings := collectCriticalFindings(data)

	if len(findings.critical) > 0 {
		b.Section(2, "Critical Issues")
		b.List(findings.critical)
	}

	if len(findings.warnings) > 0 {
		b.Section(2, "Areas for Improvement")
		b.List(findings.warnings)
	}

	if len(findings.strengths) > 0 {
		b.Section(2, "Strengths")
		b.List(findings.strengths)
	}

	// Recommendations
	recommendations := generateExecutiveRecommendations(data)

	b.Section(2, "Recommendations")

	if len(recommendations.immediate) > 0 {
		b.Section(3, "Immediate Actions")
		b.NumberedList(recommendations.immediate)
	}

	if len(recommendations.shortTerm) > 0 {
		b.Section(3, "Short-term Improvements")
		b.NumberedList(recommendations.shortTerm)
	}

	// Business Impact
	b.Section(2, "Business Impact")

	impact := assessBusinessImpact(data)

	b.KeyValue("Maintainability Risk", impact.maintainabilityRisk)
	b.KeyValue("Developer Velocity Impact", impact.velocityImpact)

	if impact.keyRisk != "" {
		b.Newline()
		b.Quote(fmt.Sprintf("%s %s", b.Bold("Key Risk:"), impact.keyRisk))
	}

	b.Footer("Code Quality")

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
	techPath := filepath.Join(analysisDir, "code-quality-technical-report.md")
	if err := os.WriteFile(techPath, []byte(techReport), 0644); err != nil {
		return fmt.Errorf("writing technical report: %w", err)
	}

	// Write executive report
	execReport := GenerateExecutiveReport(data)
	execPath := filepath.Join(analysisDir, "code-quality-executive-report.md")
	if err := os.WriteFile(execPath, []byte(execReport), 0644); err != nil {
		return fmt.Errorf("writing executive report: %w", err)
	}

	return nil
}

// Helper functions

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "None"
	}
	return strings.Join(items, ", ")
}

func formatComplexityType(typ string) string {
	// Convert complexity-type-name to readable format
	typ = strings.TrimPrefix(typ, "complexity-")
	typ = strings.ReplaceAll(typ, "-", " ")
	return strings.Title(typ)
}

func scoreToGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

func scoreToStatus(score int) string {
	switch {
	case score >= 80:
		return "Excellent"
	case score >= 60:
		return "Good"
	case score >= 40:
		return "Needs Improvement"
	default:
		return "Critical"
	}
}

type maintainabilityAssessment struct {
	status      string
	description string
}

func assessMaintainability(totalIssues, high, medium, filesAffected int) maintainabilityAssessment {
	if totalIssues == 0 {
		return maintainabilityAssessment{
			status:      "Excellent",
			description: "No complexity issues detected. Code appears well-structured and maintainable.",
		}
	}

	if high > 10 || (high > 5 && filesAffected > 20) {
		return maintainabilityAssessment{
			status:      "Critical",
			description: "Significant complexity issues detected. High-complexity code may impede development velocity and increase bug risk.",
		}
	}

	if high > 0 || medium > 20 {
		return maintainabilityAssessment{
			status:      "Needs Improvement",
			description: "Moderate complexity issues present. Consider refactoring high-complexity areas to improve maintainability.",
		}
	}

	return maintainabilityAssessment{
		status:      "Good",
		description: "Minor complexity issues detected. Overall code structure is maintainable with room for improvement.",
	}
}

func calculateOverallHealthScore(summary Summary) int {
	scores := []int{}

	if summary.TechDebt != nil {
		scores = append(scores, calculateDebtScore(summary.TechDebt))
	}

	if summary.Complexity != nil {
		scores = append(scores, calculateComplexityScore(summary.Complexity))
	}

	if summary.TestCoverage != nil {
		scores = append(scores, calculateCoverageScore(summary.TestCoverage))
	}

	if summary.CodeDocs != nil {
		scores = append(scores, summary.CodeDocs.Score)
	}

	if len(scores) == 0 {
		return 0
	}

	total := 0
	for _, s := range scores {
		total += s
	}

	return total / len(scores)
}

func calculateDebtScore(td *TechDebtSummary) int {
	if td.Error != "" {
		return 0
	}

	// Start at 100 and deduct points
	score := 100

	// Deduct for total markers (cap at 50 points)
	if td.TotalMarkers > 0 {
		deduction := td.TotalMarkers / 10
		if deduction > 50 {
			deduction = 50
		}
		score -= deduction
	}

	// Extra penalty for high priority items
	if td.ByPriority != nil {
		if high, ok := td.ByPriority["high"]; ok {
			score -= high * 2
		}
	}

	// Deduct for code issues
	if td.TotalIssues > 0 {
		deduction := td.TotalIssues / 5
		if deduction > 30 {
			deduction = 30
		}
		score -= deduction
	}

	if score < 0 {
		score = 0
	}

	return score
}

func calculateComplexityScore(c *ComplexitySummary) int {
	if c.Error != "" {
		return 0
	}

	if c.TotalIssues == 0 {
		return 100
	}

	// Start at 100 and deduct based on severity
	score := 100
	score -= c.High * 5
	score -= c.Medium * 2
	score -= c.Low

	if score < 0 {
		score = 0
	}

	return score
}

func calculateCoverageScore(tc *TestCoverageSummary) int {
	if tc.Error != "" {
		return 0
	}

	if !tc.HasTestFiles {
		return 0
	}

	// If we have coverage data, use it
	if tc.LineCoverage > 0 {
		return int(tc.LineCoverage)
	}

	// Otherwise give partial credit for having tests
	return 50
}

type criticalFindings struct {
	critical  []string
	warnings  []string
	strengths []string
}

func collectCriticalFindings(data *ReportData) criticalFindings {
	cf := criticalFindings{}

	// Technical debt findings
	if data.Summary.TechDebt != nil {
		td := data.Summary.TechDebt
		if td.ByPriority != nil {
			if high, ok := td.ByPriority["high"]; ok && high > 20 {
				cf.critical = append(cf.critical, fmt.Sprintf("Excessive high-priority technical debt: %d markers requiring immediate attention", high))
			} else if high > 10 {
				cf.warnings = append(cf.warnings, fmt.Sprintf("Moderate high-priority technical debt: %d markers should be addressed", high))
			}
		}

		if td.TotalIssues > 50 {
			cf.warnings = append(cf.warnings, fmt.Sprintf("High number of code quality issues: %d issues detected", td.TotalIssues))
		}

		if td.TotalMarkers < 20 && td.TotalIssues < 20 {
			cf.strengths = append(cf.strengths, "Low technical debt - codebase is well-maintained")
		}
	}

	// Complexity findings
	if data.Summary.Complexity != nil {
		c := data.Summary.Complexity
		if c.High > 10 {
			cf.critical = append(cf.critical, fmt.Sprintf("High code complexity: %d functions require refactoring", c.High))
		} else if c.High > 5 {
			cf.warnings = append(cf.warnings, fmt.Sprintf("Moderate code complexity: %d high-complexity functions detected", c.High))
		}

		if c.TotalIssues == 0 {
			cf.strengths = append(cf.strengths, "Excellent code structure - no complexity issues detected")
		} else if c.High == 0 && c.Medium < 10 {
			cf.strengths = append(cf.strengths, "Well-structured code with minimal complexity issues")
		}
	}

	// Test coverage findings
	if data.Summary.TestCoverage != nil {
		tc := data.Summary.TestCoverage
		if !tc.HasTestFiles {
			cf.critical = append(cf.critical, "No automated tests detected - significant quality risk")
		} else if tc.LineCoverage > 0 && tc.LineCoverage < 40 {
			cf.critical = append(cf.critical, fmt.Sprintf("Low test coverage: %.1f%% - critical gaps in testing", tc.LineCoverage))
		} else if tc.LineCoverage > 0 && tc.LineCoverage < 60 {
			cf.warnings = append(cf.warnings, fmt.Sprintf("Moderate test coverage: %.1f%% - room for improvement", tc.LineCoverage))
		} else if tc.LineCoverage >= 80 {
			cf.strengths = append(cf.strengths, fmt.Sprintf("Strong test coverage: %.1f%%", tc.LineCoverage))
		}
	}

	// Documentation findings
	if data.Summary.CodeDocs != nil {
		cd := data.Summary.CodeDocs
		if !cd.HasReadme {
			cf.warnings = append(cf.warnings, "Missing README - impacts developer onboarding")
		}

		if cd.Score >= 80 {
			cf.strengths = append(cf.strengths, "Comprehensive documentation enhances maintainability")
		}
	}

	return cf
}

type executiveRecommendations struct {
	immediate []string
	shortTerm []string
}

func generateExecutiveRecommendations(data *ReportData) executiveRecommendations {
	rec := executiveRecommendations{}

	// Technical debt recommendations
	if data.Summary.TechDebt != nil {
		td := data.Summary.TechDebt
		if td.ByPriority != nil {
			if high, ok := td.ByPriority["high"]; ok && high > 10 {
				rec.immediate = append(rec.immediate, fmt.Sprintf("Address %d high-priority technical debt items (FIXME, BUG, HACK markers)", high))
			}
		}

		if td.TotalMarkers > 100 {
			rec.shortTerm = append(rec.shortTerm, "Implement technical debt reduction program - allocate 15-20% of sprint capacity to debt paydown")
		}
	}

	// Complexity recommendations
	if data.Summary.Complexity != nil {
		c := data.Summary.Complexity
		if c.High > 5 {
			rec.immediate = append(rec.immediate, fmt.Sprintf("Refactor %d high-complexity functions to reduce maintenance burden", c.High))
		}

		if c.Medium > 20 {
			rec.shortTerm = append(rec.shortTerm, "Establish code complexity guidelines and review process")
		}
	}

	// Test coverage recommendations
	if data.Summary.TestCoverage != nil {
		tc := data.Summary.TestCoverage
		if !tc.HasTestFiles {
			rec.immediate = append(rec.immediate, "Implement test framework and begin adding automated tests")
		} else if tc.LineCoverage > 0 && tc.LineCoverage < 60 {
			rec.immediate = append(rec.immediate, "Increase test coverage to minimum 60% threshold")
		} else if tc.LineCoverage > 0 && tc.LineCoverage < 80 {
			rec.shortTerm = append(rec.shortTerm, "Continue improving test coverage toward 80% target")
		}

		if tc.HasTestFiles && len(tc.CoverageReports) == 0 {
			rec.shortTerm = append(rec.shortTerm, "Configure coverage reporting in CI/CD pipeline")
		}
	}

	// Documentation recommendations
	if data.Summary.CodeDocs != nil {
		cd := data.Summary.CodeDocs
		if !cd.HasReadme {
			rec.immediate = append(rec.immediate, "Create comprehensive README with setup and usage instructions")
		}

		if cd.Score < 60 {
			rec.shortTerm = append(rec.shortTerm, "Improve documentation to support developer onboarding and maintenance")
		}

		if !cd.HasApiDocs && cd.HasReadme {
			rec.shortTerm = append(rec.shortTerm, "Add API documentation for better developer experience")
		}
	}

	return rec
}

type businessImpact struct {
	maintainabilityRisk string
	velocityImpact      string
	keyRisk             string
}

func assessBusinessImpact(data *ReportData) businessImpact {
	impact := businessImpact{
		maintainabilityRisk: "Unknown",
		velocityImpact:      "Unknown",
	}

	overallScore := calculateOverallHealthScore(data.Summary)

	// Assess maintainability risk
	switch {
	case overallScore >= 80:
		impact.maintainabilityRisk = "Low"
		impact.velocityImpact = "Minimal"
	case overallScore >= 60:
		impact.maintainabilityRisk = "Moderate"
		impact.velocityImpact = "Minor slowdown"
	case overallScore >= 40:
		impact.maintainabilityRisk = "High"
		impact.velocityImpact = "Moderate slowdown"
		impact.keyRisk = "Code quality issues may slow feature development and increase bug rates"
	default:
		impact.maintainabilityRisk = "Critical"
		impact.velocityImpact = "Significant slowdown"
		impact.keyRisk = "Severe code quality issues pose substantial risk to development velocity, product stability, and team morale"
	}

	// Enhance risk assessment based on specific factors
	if data.Summary.TestCoverage != nil && !data.Summary.TestCoverage.HasTestFiles {
		impact.keyRisk = "Lack of automated tests creates high regression risk and slows feature development"
	}

	if data.Summary.Complexity != nil && data.Summary.Complexity.High > 15 {
		impact.keyRisk = "High code complexity significantly increases bug risk and maintenance costs"
	}

	return impact
}
