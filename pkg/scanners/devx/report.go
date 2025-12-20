package devx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// LoadReportData loads devx.json from the analysis directory
func LoadReportData(analysisDir string) (*ReportData, error) {
	devxPath := filepath.Join(analysisDir, "devx.json")
	data, err := os.ReadFile(devxPath)
	if err != nil {
		return nil, fmt.Errorf("reading devx.json: %w", err)
	}

	var result struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing devx.json: %w", err)
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
	var sb strings.Builder

	sb.WriteString("# DevX Technical Report\n\n")
	sb.WriteString(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", data.Timestamp.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString("---\n\n")

	// Onboarding Section
	if data.Summary.Onboarding != nil {
		sb.WriteString("## 1. Onboarding Analysis\n\n")
		o := data.Summary.Onboarding

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **Score** | %d/100 |\n", o.Score))
		sb.WriteString(fmt.Sprintf("| Setup Complexity | %s |\n", o.SetupComplexity))
		sb.WriteString(fmt.Sprintf("| Config Files | %d |\n", o.ConfigFileCount))
		sb.WriteString(fmt.Sprintf("| Dependencies | %d |\n", o.DependencyCount))
		sb.WriteString(fmt.Sprintf("| Build Steps | %d |\n", o.BuildStepCount))
		sb.WriteString(fmt.Sprintf("| Env Variables | %d |\n", o.EnvVarCount))
		sb.WriteString(fmt.Sprintf("| Prerequisites | %d |\n", o.PrerequisiteCount))
		sb.WriteString(fmt.Sprintf("| README Quality | %d/100 |\n", o.ReadmeQualityScore))
		sb.WriteString(fmt.Sprintf("| CONTRIBUTING.md | %s |\n", boolToCheckmark(o.HasContributing)))
		sb.WriteString(fmt.Sprintf("| .env.example | %s |\n", boolToCheckmark(o.HasEnvExample)))
		sb.WriteString("\n")

		// Config files detail
		if data.Findings.Onboarding != nil && len(data.Findings.Onboarding.ConfigFiles) > 0 {
			sb.WriteString("### Configuration Files\n\n")
			sb.WriteString("| File | Type | Tool | Lines | Complexity |\n")
			sb.WriteString("|------|------|------|-------|------------|\n")
			for _, cf := range data.Findings.Onboarding.ConfigFiles {
				sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %d | %s |\n",
					cf.Path, cf.Type, cf.Tool, cf.LineCount, cf.Complexity))
			}
			sb.WriteString("\n")
		}

		// Prerequisites
		if data.Findings.Onboarding != nil && len(data.Findings.Onboarding.Prerequisites) > 0 {
			sb.WriteString("### Prerequisites\n\n")
			for _, p := range data.Findings.Onboarding.Prerequisites {
				required := ""
				if p.Required {
					required = " (required)"
				}
				sb.WriteString(fmt.Sprintf("- **%s**%s - detected from `%s`\n", p.Name, required, p.Source))
			}
			sb.WriteString("\n")
		}

		// Setup barriers
		if data.Findings.Onboarding != nil && len(data.Findings.Onboarding.SetupBarriers) > 0 {
			sb.WriteString("### Setup Barriers\n\n")
			for _, b := range data.Findings.Onboarding.SetupBarriers {
				sb.WriteString(fmt.Sprintf("**[%s]** %s\n", strings.ToUpper(b.Severity), b.Description))
				sb.WriteString(fmt.Sprintf("   - Suggestion: %s\n\n", b.Suggestion))
			}
		}

		// README analysis
		if data.Findings.Onboarding != nil && data.Findings.Onboarding.ReadmeAnalysis != nil {
			ra := data.Findings.Onboarding.ReadmeAnalysis
			sb.WriteString("### README Quality Breakdown\n\n")
			sb.WriteString("| Section | Present |\n")
			sb.WriteString("|---------|----------|\n")
			sb.WriteString(fmt.Sprintf("| Installation | %s |\n", boolToCheckmark(ra.HasInstallSection)))
			sb.WriteString(fmt.Sprintf("| Usage | %s |\n", boolToCheckmark(ra.HasUsageSection)))
			sb.WriteString(fmt.Sprintf("| Prerequisites | %s |\n", boolToCheckmark(ra.HasPrerequisites)))
			sb.WriteString(fmt.Sprintf("| Quick Start | %s |\n", boolToCheckmark(ra.HasQuickStart)))
			sb.WriteString(fmt.Sprintf("| Examples | %s |\n", boolToCheckmark(ra.HasExamples)))
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("- Word count: %d\n", ra.WordCount))
			sb.WriteString(fmt.Sprintf("- Headers: %d\n", ra.HeaderCount))
			sb.WriteString(fmt.Sprintf("- Code blocks: %d\n", ra.CodeBlockCount))
			sb.WriteString("\n")
		}
	}

	// Sprawl Section
	if data.Summary.Sprawl != nil {
		sb.WriteString("## 2. Sprawl Analysis\n\n")
		s := data.Summary.Sprawl

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **Combined Score** | %d/100 |\n", s.CombinedScore))
		sb.WriteString(fmt.Sprintf("| Tool Sprawl Index | %d (%s) |\n", s.ToolSprawl.Index, s.ToolSprawl.Level))
		sb.WriteString(fmt.Sprintf("| Technology Sprawl Index | %d (%s) |\n", s.TechnologySprawl.Index, s.TechnologySprawl.Level))
		sb.WriteString(fmt.Sprintf("| Config Complexity | %s |\n", s.ConfigComplexity))
		sb.WriteString(fmt.Sprintf("| Total Config Lines | %d |\n", s.TotalConfigLines))
		sb.WriteString(fmt.Sprintf("| Learning Curve | %s (%d/100) |\n", s.LearningCurve, s.LearningCurveScore))
		sb.WriteString("\n")

		// Tool sprawl by category
		if len(s.ToolSprawl.ByCategory) > 0 {
			sb.WriteString("### Tool Sprawl by Category\n\n")
			sb.WriteString("| Category | Count |\n")
			sb.WriteString("|----------|-------|\n")
			for cat, count := range s.ToolSprawl.ByCategory {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", cat, count))
			}
			sb.WriteString("\n")
		}

		// Technology sprawl by category
		if len(s.TechnologySprawl.ByCategory) > 0 {
			sb.WriteString("### Technology Sprawl by Category\n\n")
			sb.WriteString("| Category | Count |\n")
			sb.WriteString("|----------|-------|\n")
			for cat, count := range s.TechnologySprawl.ByCategory {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", cat, count))
			}
			sb.WriteString("\n")
		}

		// Detected tools
		if data.Findings.Sprawl != nil && len(data.Findings.Sprawl.Tools) > 0 {
			sb.WriteString("### Detected Developer Tools\n\n")
			sb.WriteString("| Tool | Category | Config File |\n")
			sb.WriteString("|------|----------|-------------|\n")
			for _, t := range data.Findings.Sprawl.Tools {
				sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", t.Name, t.Category, t.ConfigFile))
			}
			sb.WriteString("\n")
		}

		// Detected technologies
		if data.Findings.Sprawl != nil && len(data.Findings.Sprawl.Technologies) > 0 {
			sb.WriteString("### Detected Technologies\n\n")
			sb.WriteString("| Technology | Category | Confidence | Source |\n")
			sb.WriteString("|------------|----------|------------|--------|\n")
			for _, t := range data.Findings.Sprawl.Technologies {
				sb.WriteString(fmt.Sprintf("| %s | %s | %d%% | %s |\n", t.Name, t.Category, t.Confidence, t.Source))
			}
			sb.WriteString("\n")
		}

		// Config complexity analysis
		if data.Findings.Sprawl != nil && len(data.Findings.Sprawl.ConfigAnalysis) > 0 {
			sb.WriteString("### Configuration Complexity Analysis\n\n")
			sb.WriteString("| File | Tool | Lines | Nesting | Overrides | Score |\n")
			sb.WriteString("|------|------|-------|---------|-----------|-------|\n")
			for _, c := range data.Findings.Sprawl.ConfigAnalysis {
				sb.WriteString(fmt.Sprintf("| `%s` | %s | %d | %d | %d | %d |\n",
					c.Path, c.Tool, c.LineCount, c.NestingDepth, c.OverrideCount, c.ComplexityScore))
			}
			sb.WriteString("\n")
		}

		// Sprawl issues
		if data.Findings.Sprawl != nil && len(data.Findings.Sprawl.SprawlIssues) > 0 {
			sb.WriteString("### Sprawl Issues\n\n")
			for _, issue := range data.Findings.Sprawl.SprawlIssues {
				sb.WriteString(fmt.Sprintf("**[%s]** %s\n", strings.ToUpper(issue.Severity), issue.Description))
				if len(issue.Tools) > 0 {
					sb.WriteString(fmt.Sprintf("   - Tools: %s\n", strings.Join(issue.Tools, ", ")))
				}
				sb.WriteString(fmt.Sprintf("   - Suggestion: %s\n\n", issue.Suggestion))
			}
		}

		// DORA context
		if s.DORAContext != nil && s.DORAContext.Insight != "" {
			sb.WriteString("### DORA Context\n\n")
			sb.WriteString(fmt.Sprintf("**Overall Performance:** %s\n\n", s.DORAContext.OverallPerformance))
			sb.WriteString(fmt.Sprintf("**Insight:** %s\n\n", s.DORAContext.Insight))
		}
	}

	// Workflow Section
	if data.Summary.Workflow != nil {
		sb.WriteString("## 3. Workflow Analysis\n\n")
		w := data.Summary.Workflow

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **Score** | %d/100 |\n", w.Score))
		sb.WriteString(fmt.Sprintf("| Efficiency Level | %s |\n", w.EfficiencyLevel))
		sb.WriteString(fmt.Sprintf("| PR Process Score | %d/100 |\n", w.PRProcessScore))
		sb.WriteString(fmt.Sprintf("| Local Dev Score | %d/100 |\n", w.LocalDevScore))
		sb.WriteString(fmt.Sprintf("| Feedback Loop Score | %d/100 |\n", w.FeedbackLoopScore))
		sb.WriteString("\n")

		sb.WriteString("### Capabilities\n\n")
		sb.WriteString("| Capability | Status |\n")
		sb.WriteString("|------------|--------|\n")
		sb.WriteString(fmt.Sprintf("| PR Templates | %s |\n", boolToCheckmark(w.HasPRTemplates)))
		sb.WriteString(fmt.Sprintf("| Issue Templates | %s |\n", boolToCheckmark(w.HasIssueTemplates)))
		sb.WriteString(fmt.Sprintf("| DevContainer | %s |\n", boolToCheckmark(w.HasDevContainer)))
		sb.WriteString(fmt.Sprintf("| Docker Compose | %s |\n", boolToCheckmark(w.HasDockerCompose)))
		sb.WriteString(fmt.Sprintf("| Hot Reload | %s |\n", boolToCheckmark(w.HasHotReload)))
		sb.WriteString(fmt.Sprintf("| Watch Mode | %s |\n", boolToCheckmark(w.HasWatchMode)))
		sb.WriteString("\n")

		// Dev setup details
		if data.Findings.Workflow != nil && data.Findings.Workflow.DevSetup != nil {
			ds := data.Findings.Workflow.DevSetup
			if len(ds.DevScripts) > 0 {
				sb.WriteString("### Available Dev Scripts\n\n")
				for _, script := range ds.DevScripts {
					sb.WriteString(fmt.Sprintf("- `%s`\n", script))
				}
				sb.WriteString("\n")
			}
		}

		// Feedback tools
		if data.Findings.Workflow != nil && len(data.Findings.Workflow.FeedbackTools) > 0 {
			sb.WriteString("### Feedback Tools\n\n")
			sb.WriteString("| Tool | Type | Source |\n")
			sb.WriteString("|------|------|--------|\n")
			for _, ft := range data.Findings.Workflow.FeedbackTools {
				sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", ft.Name, ft.Type, ft.Source))
			}
			sb.WriteString("\n")
		}

		// Workflow issues
		if data.Findings.Workflow != nil && len(data.Findings.Workflow.WorkflowIssues) > 0 {
			sb.WriteString("### Workflow Issues\n\n")
			for _, issue := range data.Findings.Workflow.WorkflowIssues {
				sb.WriteString(fmt.Sprintf("**[%s]** %s\n", strings.ToUpper(issue.Severity), issue.Description))
				sb.WriteString(fmt.Sprintf("   - Suggestion: %s\n\n", issue.Suggestion))
			}
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString("*Generated by Zero DevX Scanner*\n")

	return sb.String()
}

// GenerateExecutiveReport creates a high-level summary for engineering leaders
func GenerateExecutiveReport(data *ReportData) string {
	var sb strings.Builder

	sb.WriteString("# Developer Experience Report\n\n")
	sb.WriteString(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	sb.WriteString(fmt.Sprintf("**Date:** %s\n\n", data.Timestamp.Format("January 2, 2006")))

	// Overall Score Card
	sb.WriteString("## Executive Summary\n\n")

	overallScore := calculateOverallScore(data.Summary)
	overallGrade := scoreToGrade(overallScore)

	sb.WriteString(fmt.Sprintf("### Overall Developer Experience: %s (%d/100)\n\n", overallGrade, overallScore))

	// Score breakdown
	sb.WriteString("| Area | Score | Status |\n")
	sb.WriteString("|------|-------|--------|\n")

	if data.Summary.Onboarding != nil {
		status := scoreToStatus(data.Summary.Onboarding.Score)
		sb.WriteString(fmt.Sprintf("| **Onboarding** | %d/100 | %s |\n", data.Summary.Onboarding.Score, status))
	}
	if data.Summary.Sprawl != nil {
		status := scoreToStatus(data.Summary.Sprawl.CombinedScore)
		sb.WriteString(fmt.Sprintf("| **Complexity** | %d/100 | %s |\n", data.Summary.Sprawl.CombinedScore, status))
	}
	if data.Summary.Workflow != nil {
		status := scoreToStatus(data.Summary.Workflow.Score)
		sb.WriteString(fmt.Sprintf("| **Workflow** | %d/100 | %s |\n", data.Summary.Workflow.Score, status))
	}
	sb.WriteString("\n")

	// Key Metrics
	sb.WriteString("## Key Metrics\n\n")

	if data.Summary.Onboarding != nil {
		o := data.Summary.Onboarding
		sb.WriteString("### Onboarding\n\n")
		sb.WriteString(fmt.Sprintf("- **Setup Complexity:** %s\n", strings.Title(o.SetupComplexity)))
		sb.WriteString(fmt.Sprintf("- **Dependencies:** %d packages\n", o.DependencyCount))
		sb.WriteString(fmt.Sprintf("- **README Quality:** %d/100\n", o.ReadmeQualityScore))
		if !o.HasContributing {
			sb.WriteString("- Missing CONTRIBUTING.md\n")
		}
		sb.WriteString("\n")
	}

	if data.Summary.Sprawl != nil {
		s := data.Summary.Sprawl
		sb.WriteString("### Technology Stack\n\n")
		sb.WriteString(fmt.Sprintf("- **Tool Sprawl:** %s (%d tools)\n", strings.Title(s.ToolSprawl.Level), s.ToolSprawl.Index))
		sb.WriteString(fmt.Sprintf("- **Technology Sprawl:** %s (%d technologies)\n", strings.Title(s.TechnologySprawl.Level), s.TechnologySprawl.Index))
		sb.WriteString(fmt.Sprintf("- **Learning Curve:** %s\n", strings.Title(s.LearningCurve)))
		sb.WriteString("\n")
	}

	if data.Summary.Workflow != nil {
		w := data.Summary.Workflow
		sb.WriteString("### Development Workflow\n\n")
		sb.WriteString(fmt.Sprintf("- **Efficiency:** %s\n", strings.Title(w.EfficiencyLevel)))
		if w.HasHotReload {
			sb.WriteString("- Hot reload enabled\n")
		}
		if w.HasDevContainer {
			sb.WriteString("- DevContainer available\n")
		}
		if !w.HasPRTemplates {
			sb.WriteString("- Missing PR templates\n")
		}
		sb.WriteString("\n")
	}

	// Key Findings
	sb.WriteString("## Key Findings\n\n")

	findings := collectKeyFindings(data)
	if len(findings.critical) > 0 {
		sb.WriteString("### Critical Issues\n\n")
		for _, f := range findings.critical {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	if len(findings.warnings) > 0 {
		sb.WriteString("### Areas for Improvement\n\n")
		for _, f := range findings.warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	if len(findings.positive) > 0 {
		sb.WriteString("### Strengths\n\n")
		for _, f := range findings.positive {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	// Recommendations
	sb.WriteString("## Recommendations\n\n")

	recommendations := generateRecommendations(data)
	if len(recommendations.immediate) > 0 {
		sb.WriteString("### Immediate Actions\n\n")
		for i, r := range recommendations.immediate {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, r))
		}
		sb.WriteString("\n")
	}

	if len(recommendations.shortTerm) > 0 {
		sb.WriteString("### Short-term Improvements\n\n")
		for i, r := range recommendations.shortTerm {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, r))
		}
		sb.WriteString("\n")
	}

	// Impact Assessment
	sb.WriteString("## Business Impact\n\n")

	impact := assessImpact(data)
	sb.WriteString(fmt.Sprintf("**Estimated Onboarding Time:** %s\n\n", impact.onboardingTime))
	sb.WriteString(fmt.Sprintf("**Developer Productivity Risk:** %s\n\n", impact.productivityRisk))

	if impact.keyRisk != "" {
		sb.WriteString(fmt.Sprintf("**Key Risk:** %s\n\n", impact.keyRisk))
	}

	sb.WriteString("---\n\n")
	sb.WriteString("*Generated by Zero DevX Scanner*\n")

	return sb.String()
}

// Helper functions

func boolToCheckmark(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func calculateOverallScore(summary Summary) int {
	count := 0
	total := 0

	if summary.Onboarding != nil {
		total += summary.Onboarding.Score
		count++
	}
	if summary.Sprawl != nil {
		total += summary.Sprawl.CombinedScore
		count++
	}
	if summary.Workflow != nil {
		total += summary.Workflow.Score
		count++
	}

	if count == 0 {
		return 0
	}
	return total / count
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

type keyFindings struct {
	critical []string
	warnings []string
	positive []string
}

func collectKeyFindings(data *ReportData) keyFindings {
	kf := keyFindings{}

	// Check onboarding findings
	if data.Findings.Onboarding != nil {
		for _, b := range data.Findings.Onboarding.SetupBarriers {
			if b.Severity == "high" {
				kf.critical = append(kf.critical, b.Description)
			} else if b.Severity == "medium" {
				kf.warnings = append(kf.warnings, b.Description)
			}
		}
	}

	// Check sprawl findings
	if data.Findings.Sprawl != nil {
		for _, issue := range data.Findings.Sprawl.SprawlIssues {
			if issue.Severity == "high" {
				kf.critical = append(kf.critical, issue.Description)
			} else if issue.Severity == "medium" {
				kf.warnings = append(kf.warnings, issue.Description)
			}
		}
	}

	// Check workflow findings
	if data.Findings.Workflow != nil {
		for _, issue := range data.Findings.Workflow.WorkflowIssues {
			if issue.Severity == "high" {
				kf.critical = append(kf.critical, issue.Description)
			} else if issue.Severity == "medium" {
				kf.warnings = append(kf.warnings, issue.Description)
			}
		}
	}

	// Collect positives
	if data.Summary.Onboarding != nil {
		if data.Summary.Onboarding.Score >= 80 {
			kf.positive = append(kf.positive, "Strong onboarding documentation and setup")
		}
		if data.Summary.Onboarding.ReadmeQualityScore >= 80 {
			kf.positive = append(kf.positive, "High-quality README with comprehensive documentation")
		}
	}

	if data.Summary.Sprawl != nil {
		if data.Summary.Sprawl.ToolSprawl.Level == "low" {
			kf.positive = append(kf.positive, "Well-managed tooling with minimal sprawl")
		}
		if data.Summary.Sprawl.LearningCurve == "low" {
			kf.positive = append(kf.positive, "Low learning curve for new developers")
		}
	}

	if data.Summary.Workflow != nil {
		if data.Summary.Workflow.HasHotReload {
			kf.positive = append(kf.positive, "Hot reload enabled for fast feedback")
		}
		if data.Summary.Workflow.HasDevContainer {
			kf.positive = append(kf.positive, "DevContainer available for consistent environments")
		}
	}

	return kf
}

type recommendations struct {
	immediate []string
	shortTerm []string
}

func generateRecommendations(data *ReportData) recommendations {
	rec := recommendations{}

	// Onboarding recommendations
	if data.Summary.Onboarding != nil {
		o := data.Summary.Onboarding
		if !o.HasContributing {
			rec.immediate = append(rec.immediate, "Add CONTRIBUTING.md with setup instructions")
		}
		if o.ReadmeQualityScore < 60 {
			rec.immediate = append(rec.immediate, "Improve README with installation and quick start sections")
		}
		if o.EnvVarCount > 0 && !o.HasEnvExample {
			rec.immediate = append(rec.immediate, "Create .env.example with all required environment variables")
		}
	}

	// Workflow recommendations
	if data.Summary.Workflow != nil {
		w := data.Summary.Workflow
		if !w.HasPRTemplates {
			rec.shortTerm = append(rec.shortTerm, "Add PR templates to standardize code reviews")
		}
		if !w.HasHotReload && !w.HasWatchMode {
			rec.shortTerm = append(rec.shortTerm, "Implement hot reload or watch mode for faster development feedback")
		}
		if !w.HasDevContainer && !w.HasDockerCompose {
			rec.shortTerm = append(rec.shortTerm, "Add DevContainer or Docker Compose for consistent development environments")
		}
	}

	// Sprawl recommendations
	if data.Summary.Sprawl != nil {
		s := data.Summary.Sprawl
		if s.ToolSprawl.Level == "excessive" || s.ToolSprawl.Level == "high" {
			rec.shortTerm = append(rec.shortTerm, "Audit and consolidate development tools to reduce configuration overhead")
		}
		if s.TechnologySprawl.Level == "excessive" {
			rec.shortTerm = append(rec.shortTerm, "Review technology stack for opportunities to reduce complexity")
		}
	}

	return rec
}

type impactAssessment struct {
	onboardingTime   string
	productivityRisk string
	keyRisk          string
}

func assessImpact(data *ReportData) impactAssessment {
	impact := impactAssessment{
		onboardingTime:   "Unknown",
		productivityRisk: "Unknown",
	}

	if data.Summary.Onboarding != nil {
		o := data.Summary.Onboarding
		switch {
		case o.Score >= 80:
			impact.onboardingTime = "1-2 days (fast)"
		case o.Score >= 60:
			impact.onboardingTime = "3-5 days (moderate)"
		case o.Score >= 40:
			impact.onboardingTime = "1-2 weeks (slow)"
		default:
			impact.onboardingTime = "2+ weeks (very slow)"
		}
	}

	overallScore := calculateOverallScore(data.Summary)
	switch {
	case overallScore >= 80:
		impact.productivityRisk = "Low"
	case overallScore >= 60:
		impact.productivityRisk = "Moderate"
	case overallScore >= 40:
		impact.productivityRisk = "High"
	default:
		impact.productivityRisk = "Critical"
		impact.keyRisk = "Developer experience issues may significantly impact team velocity and retention"
	}

	return impact
}

// WriteReports generates and writes both reports to the analysis directory
func WriteReports(analysisDir string) error {
	data, err := LoadReportData(analysisDir)
	if err != nil {
		return err
	}

	// Write technical report
	techReport := GenerateTechnicalReport(data)
	techPath := filepath.Join(analysisDir, "devx-technical-report.md")
	if err := os.WriteFile(techPath, []byte(techReport), 0644); err != nil {
		return fmt.Errorf("writing technical report: %w", err)
	}

	// Write executive report
	execReport := GenerateExecutiveReport(data)
	execPath := filepath.Join(analysisDir, "devx-executive-report.md")
	if err := os.WriteFile(execPath, []byte(execReport), 0644); err != nil {
		return fmt.Errorf("writing executive report: %w", err)
	}

	return nil
}
