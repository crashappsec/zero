package techid

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

// LoadReportData loads technology.json from the analysis directory
func LoadReportData(analysisDir string) (*ReportData, error) {
	techPath := filepath.Join(analysisDir, "technology.json")
	data, err := os.ReadFile(techPath)
	if err != nil {
		return nil, fmt.Errorf("reading technology.json: %w", err)
	}

	var result struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing technology.json: %w", err)
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

	sb.WriteString("# Technology Identification Technical Report\n\n")
	sb.WriteString(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", data.Timestamp.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString("---\n\n")

	// Technology Detection Section
	if data.Summary.Technology != nil {
		sb.WriteString("## 1. Technology Stack\n\n")
		t := data.Summary.Technology

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **Total Technologies** | %d |\n", t.TotalTechnologies))
		if len(t.PrimaryLanguages) > 0 {
			sb.WriteString(fmt.Sprintf("| Primary Languages | %s |\n", strings.Join(t.PrimaryLanguages, ", ")))
		}
		if len(t.Frameworks) > 0 {
			sb.WriteString(fmt.Sprintf("| Frameworks | %s |\n", strings.Join(t.Frameworks, ", ")))
		}
		if len(t.Databases) > 0 {
			sb.WriteString(fmt.Sprintf("| Databases | %s |\n", strings.Join(t.Databases, ", ")))
		}
		if len(t.CloudServices) > 0 {
			sb.WriteString(fmt.Sprintf("| Cloud Services | %s |\n", strings.Join(t.CloudServices, ", ")))
		}
		sb.WriteString("\n")

		// Technology breakdown by category
		if len(t.ByCategory) > 0 {
			sb.WriteString("### Technologies by Category\n\n")
			sb.WriteString("| Category | Count |\n")
			sb.WriteString("|----------|-------|\n")

			// Sort categories by count
			type catCount struct {
				category string
				count    int
			}
			var cats []catCount
			for cat, count := range t.ByCategory {
				cats = append(cats, catCount{cat, count})
			}
			sort.Slice(cats, func(i, j int) bool { return cats[i].count > cats[j].count })

			for _, cc := range cats {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", cc.category, cc.count))
			}
			sb.WriteString("\n")
		}

		// Detailed technology list
		if len(data.Findings.Technology) > 0 {
			sb.WriteString("### Detected Technologies\n\n")
			sb.WriteString("| Name | Category | Version | Confidence | Source |\n")
			sb.WriteString("|------|----------|---------|------------|--------|\n")

			// Sort by confidence (highest first)
			technologies := make([]Technology, len(data.Findings.Technology))
			copy(technologies, data.Findings.Technology)
			sort.Slice(technologies, func(i, j int) bool {
				return technologies[i].Confidence > technologies[j].Confidence
			})

			for _, tech := range technologies {
				version := tech.Version
				if version == "" {
					version = "-"
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d%% | %s |\n",
					tech.Name, tech.Category, version, tech.Confidence, tech.Source))
			}
			sb.WriteString("\n")
		}
	}

	// AI/ML Models Section
	if data.Summary.Models != nil {
		sb.WriteString("## 2. Machine Learning Models (ML-BOM)\n\n")
		m := data.Summary.Models

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **Total Models** | %d |\n", m.TotalModels))
		sb.WriteString(fmt.Sprintf("| Local Model Files | %d |\n", m.LocalModelFiles))
		sb.WriteString(fmt.Sprintf("| API Models | %d |\n", m.APIModels))
		sb.WriteString(fmt.Sprintf("| With Model Card | %d |\n", m.WithModelCard))
		sb.WriteString(fmt.Sprintf("| With License Info | %d |\n", m.WithLicense))
		sb.WriteString(fmt.Sprintf("| With Dataset Info | %d |\n", m.WithDatasetInfo))
		sb.WriteString("\n")

		// Models by source
		if len(m.BySource) > 0 {
			sb.WriteString("### Models by Source\n\n")
			sb.WriteString("| Source | Count |\n")
			sb.WriteString("|--------|-------|\n")
			for src, count := range m.BySource {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", src, count))
			}
			sb.WriteString("\n")
		}

		// Models by format
		if len(m.ByFormat) > 0 {
			sb.WriteString("### Models by Format\n\n")
			sb.WriteString("| Format | Count | Security Risk |\n")
			sb.WriteString("|--------|-------|---------------|\n")

			for format, count := range m.ByFormat {
				risk := "Unknown"
				// Find risk level from ModelFileFormats
				for _, info := range ModelFileFormats {
					if info.Format == format {
						risk = strings.Title(info.Risk)
						break
					}
				}
				sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", format, count, risk))
			}
			sb.WriteString("\n")
		}

		// Detailed model list
		if len(data.Findings.Models) > 0 {
			sb.WriteString("### Detected Models\n\n")

			for _, model := range data.Findings.Models {
				sb.WriteString(fmt.Sprintf("#### %s\n\n", model.Name))

				sb.WriteString("| Property | Value |\n")
				sb.WriteString("|----------|-------|\n")
				sb.WriteString(fmt.Sprintf("| Source | %s |\n", model.Source))
				if model.Format != "" {
					sb.WriteString(fmt.Sprintf("| Format | %s |\n", model.Format))
				}
				if model.SecurityRisk != "" {
					sb.WriteString(fmt.Sprintf("| Security Risk | **%s** |\n", strings.ToUpper(model.SecurityRisk)))
				}
				if model.License != "" {
					sb.WriteString(fmt.Sprintf("| License | %s |\n", model.License))
				}
				if model.Architecture != "" {
					sb.WriteString(fmt.Sprintf("| Architecture | %s |\n", model.Architecture))
				}
				if model.Task != "" {
					sb.WriteString(fmt.Sprintf("| Task | %s |\n", model.Task))
				}
				if model.FilePath != "" {
					sb.WriteString(fmt.Sprintf("| File Path | `%s` |\n", model.FilePath))
				}

				if len(model.SecurityNotes) > 0 {
					sb.WriteString("\n**Security Notes:**\n")
					for _, note := range model.SecurityNotes {
						sb.WriteString(fmt.Sprintf("- %s\n", note))
					}
				}

				if len(model.Datasets) > 0 {
					sb.WriteString(fmt.Sprintf("\n**Training Datasets:** %s\n", strings.Join(model.Datasets, ", ")))
				}

				sb.WriteString("\n")
			}
		}
	}

	// AI/ML Frameworks Section
	if data.Summary.Frameworks != nil {
		sb.WriteString("## 3. AI/ML Frameworks\n\n")
		f := data.Summary.Frameworks

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **Total Frameworks** | %d |\n", f.TotalFrameworks))
		if len(f.Detected) > 0 {
			sb.WriteString(fmt.Sprintf("| Detected | %s |\n", strings.Join(f.Detected, ", ")))
		}
		sb.WriteString("\n")

		// Frameworks by category
		if len(f.ByCategory) > 0 {
			sb.WriteString("### Frameworks by Category\n\n")
			sb.WriteString("| Category | Count |\n")
			sb.WriteString("|----------|-------|\n")
			for cat, count := range f.ByCategory {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", cat, count))
			}
			sb.WriteString("\n")
		}

		// Detailed framework list
		if len(data.Findings.Frameworks) > 0 {
			sb.WriteString("### Framework Details\n\n")
			sb.WriteString("| Framework | Category | Version | Package |\n")
			sb.WriteString("|-----------|----------|---------|----------|\n")
			for _, fw := range data.Findings.Frameworks {
				version := fw.Version
				if version == "" {
					version = "-"
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | `%s` |\n",
					fw.Name, fw.Category, version, fw.Package))
			}
			sb.WriteString("\n")
		}
	}

	// Datasets Section
	if data.Summary.Datasets != nil && data.Summary.Datasets.TotalDatasets > 0 {
		sb.WriteString("## 4. Training Datasets\n\n")
		d := data.Summary.Datasets

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **Total Datasets** | %d |\n", d.TotalDatasets))
		sb.WriteString(fmt.Sprintf("| With License | %d |\n", d.WithLicense))
		sb.WriteString(fmt.Sprintf("| With Provenance | %d |\n", d.WithProvenance))
		sb.WriteString("\n")

		// Datasets by source
		if len(d.BySource) > 0 {
			sb.WriteString("### Datasets by Source\n\n")
			sb.WriteString("| Source | Count |\n")
			sb.WriteString("|--------|-------|\n")
			for src, count := range d.BySource {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", src, count))
			}
			sb.WriteString("\n")
		}

		// Detailed dataset list
		if len(data.Findings.Datasets) > 0 {
			sb.WriteString("### Dataset Details\n\n")
			for _, ds := range data.Findings.Datasets {
				sb.WriteString(fmt.Sprintf("#### %s\n\n", ds.Name))
				sb.WriteString(fmt.Sprintf("- **Source:** %s\n", ds.Source))
				if ds.License != "" {
					sb.WriteString(fmt.Sprintf("- **License:** %s\n", ds.License))
				}
				if ds.Size != "" {
					sb.WriteString(fmt.Sprintf("- **Size:** %s\n", ds.Size))
				}
				if len(ds.UsedBy) > 0 {
					sb.WriteString(fmt.Sprintf("- **Used by:** %s\n", strings.Join(ds.UsedBy, ", ")))
				}
				sb.WriteString("\n")
			}
		}
	}

	// Security Findings Section
	if data.Summary.Security != nil && data.Summary.Security.TotalFindings > 0 {
		sb.WriteString("## 5. AI/ML Security Findings\n\n")
		s := data.Summary.Security

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Severity | Count |\n")
		sb.WriteString("|----------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **Critical** | %d |\n", s.Critical))
		sb.WriteString(fmt.Sprintf("| **High** | %d |\n", s.High))
		sb.WriteString(fmt.Sprintf("| **Medium** | %d |\n", s.Medium))
		sb.WriteString(fmt.Sprintf("| **Low** | %d |\n", s.Low))
		sb.WriteString("\n")

		// Key security metrics
		if s.UnsafePickles > 0 || s.ExposedAPIKeys > 0 {
			sb.WriteString("### Key Security Issues\n\n")
			if s.UnsafePickles > 0 {
				sb.WriteString(fmt.Sprintf("- **Unsafe Pickle Models:** %d (risk of arbitrary code execution)\n", s.UnsafePickles))
			}
			if s.ExposedAPIKeys > 0 {
				sb.WriteString(fmt.Sprintf("- **Exposed API Keys:** %d\n", s.ExposedAPIKeys))
			}
			sb.WriteString("\n")
		}

		// Security findings by category
		if len(s.ByCategory) > 0 {
			sb.WriteString("### Findings by Category\n\n")
			sb.WriteString("| Category | Count |\n")
			sb.WriteString("|----------|-------|\n")
			for cat, count := range s.ByCategory {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", cat, count))
			}
			sb.WriteString("\n")
		}

		// Detailed security findings
		if len(data.Findings.Security) > 0 {
			sb.WriteString("### Detailed Findings\n\n")

			// Sort by severity
			findings := make([]SecurityFinding, len(data.Findings.Security))
			copy(findings, data.Findings.Security)
			sort.Slice(findings, func(i, j int) bool {
				severityOrder := map[string]int{"critical": 0, "high": 1, "medium": 2, "low": 3}
				return severityOrder[findings[i].Severity] < severityOrder[findings[j].Severity]
			})

			for _, finding := range findings {
				sb.WriteString(fmt.Sprintf("#### [%s] %s\n\n", strings.ToUpper(finding.Severity), finding.Title))
				sb.WriteString(fmt.Sprintf("**ID:** %s\n\n", finding.ID))
				sb.WriteString(fmt.Sprintf("%s\n\n", finding.Description))

				if finding.File != "" {
					location := finding.File
					if finding.Line > 0 {
						location = fmt.Sprintf("%s:%d", location, finding.Line)
					}
					sb.WriteString(fmt.Sprintf("**Location:** `%s`\n\n", location))
				}

				if finding.ModelName != "" {
					sb.WriteString(fmt.Sprintf("**Model:** %s\n\n", finding.ModelName))
				}

				if finding.Remediation != "" {
					sb.WriteString(fmt.Sprintf("**Remediation:** %s\n\n", finding.Remediation))
				}

				if len(finding.References) > 0 {
					sb.WriteString("**References:**\n")
					for _, ref := range finding.References {
						sb.WriteString(fmt.Sprintf("- %s\n", ref))
					}
					sb.WriteString("\n")
				}
			}
		}
	}

	// Governance Findings Section
	if data.Summary.Governance != nil && data.Summary.Governance.TotalIssues > 0 {
		sb.WriteString("## 6. AI Governance Issues\n\n")
		g := data.Summary.Governance

		sb.WriteString("### Summary\n\n")
		sb.WriteString("| Issue Type | Count |\n")
		sb.WriteString("|------------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Missing Model Cards | %d |\n", g.MissingModelCards))
		sb.WriteString(fmt.Sprintf("| Missing Licenses | %d |\n", g.MissingLicenses))
		sb.WriteString(fmt.Sprintf("| Blocked Licenses | %d |\n", g.BlockedLicenses))
		sb.WriteString(fmt.Sprintf("| Missing Dataset Info | %d |\n", g.MissingDatasetInfo))
		sb.WriteString("\n")

		// Detailed governance findings
		if len(data.Findings.Governance) > 0 {
			sb.WriteString("### Governance Findings\n\n")

			for _, finding := range data.Findings.Governance {
				sb.WriteString(fmt.Sprintf("#### [%s] %s\n\n", strings.ToUpper(finding.Severity), finding.Title))
				sb.WriteString(fmt.Sprintf("%s\n\n", finding.Description))

				if finding.ModelName != "" {
					sb.WriteString(fmt.Sprintf("**Model:** %s\n\n", finding.ModelName))
				}

				if finding.Remediation != "" {
					sb.WriteString(fmt.Sprintf("**Remediation:** %s\n\n", finding.Remediation))
				}
			}
		}
	}

	// Semgrep Analysis Stats
	if data.Summary.SemgrepRulesLoaded > 0 || data.Summary.SemgrepFindings > 0 {
		sb.WriteString("## 7. Static Analysis\n\n")
		sb.WriteString(fmt.Sprintf("- **Semgrep Rules Loaded:** %d\n", data.Summary.SemgrepRulesLoaded))
		sb.WriteString(fmt.Sprintf("- **Semgrep Findings:** %d\n", data.Summary.SemgrepFindings))
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("*Generated by Zero Technology Identification Scanner*\n")

	return sb.String()
}

// GenerateExecutiveReport creates a high-level ML-BOM and technology summary for leadership
func GenerateExecutiveReport(data *ReportData) string {
	var sb strings.Builder

	sb.WriteString("# Technology Stack & ML-BOM Report\n\n")
	sb.WriteString(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	sb.WriteString(fmt.Sprintf("**Date:** %s\n\n", data.Timestamp.Format("January 2, 2006")))

	// Executive Summary
	sb.WriteString("## Executive Summary\n\n")

	// Technology stack overview
	if data.Summary.Technology != nil {
		t := data.Summary.Technology
		sb.WriteString("### Technology Stack\n\n")

		if len(t.PrimaryLanguages) > 0 {
			sb.WriteString(fmt.Sprintf("**Primary Languages:** %s\n\n", strings.Join(t.PrimaryLanguages, ", ")))
		}

		if len(t.Frameworks) > 0 {
			sb.WriteString(fmt.Sprintf("**Key Frameworks:** %s\n\n", strings.Join(t.Frameworks, ", ")))
		}

		sb.WriteString(fmt.Sprintf("**Total Technologies Detected:** %d across %d categories\n\n",
			t.TotalTechnologies, len(t.ByCategory)))
	}

	// ML-BOM Overview
	if data.Summary.Models != nil && data.Summary.Models.TotalModels > 0 {
		m := data.Summary.Models
		sb.WriteString("### Machine Learning Bill of Materials (ML-BOM)\n\n")

		sb.WriteString(fmt.Sprintf("This repository uses **%d machine learning models**.\n\n", m.TotalModels))

		sb.WriteString("| Category | Count |\n")
		sb.WriteString("|----------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Local Model Files | %d |\n", m.LocalModelFiles))
		sb.WriteString(fmt.Sprintf("| API-based Models | %d |\n", m.APIModels))
		sb.WriteString(fmt.Sprintf("| Models with Documentation | %d |\n", m.WithModelCard))
		sb.WriteString(fmt.Sprintf("| Models with License Info | %d |\n", m.WithLicense))
		sb.WriteString("\n")

		// Model format security summary
		hasHighRiskFormats := false
		for format := range m.ByFormat {
			if format == "pickle" {
				hasHighRiskFormats = true
				break
			}
		}

		if hasHighRiskFormats {
			sb.WriteString("#### Security Notice\n\n")
			pickleCount := m.ByFormat["pickle"]
			sb.WriteString(fmt.Sprintf("This repository contains **%d pickle-based models**, which can execute arbitrary code during loading. ", pickleCount))
			sb.WriteString("Consider migrating to safer formats like SafeTensors or ONNX.\n\n")
		}
	}

	// AI/ML Framework Summary
	if data.Summary.Frameworks != nil && data.Summary.Frameworks.TotalFrameworks > 0 {
		f := data.Summary.Frameworks
		sb.WriteString("### AI/ML Frameworks\n\n")

		if len(f.Detected) > 0 {
			sb.WriteString(fmt.Sprintf("**Frameworks in Use:** %s\n\n", strings.Join(f.Detected, ", ")))
		}
	}

	// Key Findings
	sb.WriteString("## Key Findings\n\n")

	findings := collectKeyTechFindings(data)

	if len(findings.critical) > 0 {
		sb.WriteString("### Critical Security Issues\n\n")
		for _, f := range findings.critical {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	if len(findings.warnings) > 0 {
		sb.WriteString("### Recommendations\n\n")
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

	// Model Inventory Table
	if len(data.Findings.Models) > 0 {
		sb.WriteString("## Model Inventory\n\n")
		sb.WriteString("| Model Name | Source | Format | Security Risk | License |\n")
		sb.WriteString("|------------|--------|--------|---------------|----------|\n")

		for _, model := range data.Findings.Models {
			risk := model.SecurityRisk
			if risk == "" {
				risk = "unknown"
			}
			license := model.License
			if license == "" {
				license = "Not specified"
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				model.Name, model.Source, model.Format, strings.Title(risk), license))
		}
		sb.WriteString("\n")
	}

	// Governance Summary
	if data.Summary.Governance != nil && data.Summary.Governance.TotalIssues > 0 {
		g := data.Summary.Governance
		sb.WriteString("## AI Governance\n\n")

		totalModels := 0
		if data.Summary.Models != nil {
			totalModels = data.Summary.Models.TotalModels
		}

		if g.MissingModelCards > 0 {
			sb.WriteString(fmt.Sprintf("- **%d of %d models** lack model cards (documentation)\n",
				g.MissingModelCards, totalModels))
		}
		if g.MissingLicenses > 0 {
			sb.WriteString(fmt.Sprintf("- **%d models** have no license information\n", g.MissingLicenses))
		}
		if g.BlockedLicenses > 0 {
			sb.WriteString(fmt.Sprintf("- **%d models** use blocked or problematic licenses\n", g.BlockedLicenses))
		}
		sb.WriteString("\n")
	}

	// Risk Assessment
	sb.WriteString("## Risk Assessment\n\n")

	impact := assessTechImpact(data)
	sb.WriteString(fmt.Sprintf("**ML Security Risk:** %s\n\n", impact.mlSecurityRisk))
	sb.WriteString(fmt.Sprintf("**Governance Maturity:** %s\n\n", impact.governanceMaturity))

	if impact.keyRisk != "" {
		sb.WriteString(fmt.Sprintf("**Key Risk:** %s\n\n", impact.keyRisk))
	}

	sb.WriteString("---\n\n")
	sb.WriteString("*Generated by Zero Technology Identification Scanner*\n")

	return sb.String()
}

// Helper functions

type techKeyFindings struct {
	critical []string
	warnings []string
	positive []string
}

func collectKeyTechFindings(data *ReportData) techKeyFindings {
	kf := techKeyFindings{}

	// Security findings
	if data.Summary.Security != nil {
		s := data.Summary.Security
		if s.Critical > 0 {
			kf.critical = append(kf.critical,
				fmt.Sprintf("%d critical AI/ML security findings require immediate attention", s.Critical))
		}
		if s.UnsafePickles > 0 {
			kf.critical = append(kf.critical,
				fmt.Sprintf("%d unsafe pickle models detected (arbitrary code execution risk)", s.UnsafePickles))
		}
		if s.ExposedAPIKeys > 0 {
			kf.critical = append(kf.critical,
				fmt.Sprintf("%d exposed API keys found", s.ExposedAPIKeys))
		}
		if s.High > 0 {
			kf.warnings = append(kf.warnings,
				fmt.Sprintf("Address %d high-severity AI/ML security findings", s.High))
		}
	}

	// Governance findings
	if data.Summary.Governance != nil {
		g := data.Summary.Governance
		if g.MissingModelCards > 0 {
			kf.warnings = append(kf.warnings,
				fmt.Sprintf("Add model cards for %d models to improve transparency", g.MissingModelCards))
		}
		if g.BlockedLicenses > 0 {
			kf.critical = append(kf.critical,
				fmt.Sprintf("%d models use blocked licenses", g.BlockedLicenses))
		}
	}

	// Model analysis
	if data.Summary.Models != nil {
		m := data.Summary.Models

		// Check for safe formats
		safeTensorsCount := m.ByFormat["safetensors"]
		if safeTensorsCount > 0 {
			kf.positive = append(kf.positive,
				fmt.Sprintf("%d models use SafeTensors (secure format)", safeTensorsCount))
		}

		// Check documentation
		if m.WithModelCard > 0 {
			kf.positive = append(kf.positive,
				fmt.Sprintf("%d models have model cards (good documentation)", m.WithModelCard))
		}

		// Check if all models are from reputable sources
		if m.BySource["huggingface"] > 0 {
			kf.positive = append(kf.positive,
				fmt.Sprintf("Using HuggingFace Hub for model distribution"))
		}
	}

	// Framework analysis
	if data.Summary.Frameworks != nil && data.Summary.Frameworks.TotalFrameworks > 0 {
		f := data.Summary.Frameworks
		if f.TotalFrameworks > 0 {
			kf.positive = append(kf.positive,
				fmt.Sprintf("Using established AI/ML frameworks (%s)", strings.Join(f.Detected, ", ")))
		}
	}

	return kf
}

type techImpactAssessment struct {
	mlSecurityRisk     string
	governanceMaturity string
	keyRisk            string
}

func assessTechImpact(data *ReportData) techImpactAssessment {
	impact := techImpactAssessment{
		mlSecurityRisk:     "Low",
		governanceMaturity: "Mature",
	}

	// Calculate ML security risk
	if data.Summary.Security != nil {
		s := data.Summary.Security
		criticalAndHigh := s.Critical + s.High

		switch {
		case s.Critical > 0 || s.UnsafePickles > 0:
			impact.mlSecurityRisk = "Critical"
			impact.keyRisk = "Unsafe ML models or critical security vulnerabilities require immediate remediation"
		case criticalAndHigh >= 5:
			impact.mlSecurityRisk = "High"
		case s.High > 0 || s.Medium >= 3:
			impact.mlSecurityRisk = "Medium"
		case s.Medium > 0 || s.Low > 0:
			impact.mlSecurityRisk = "Low"
		}
	}

	// Calculate governance maturity
	if data.Summary.Governance != nil && data.Summary.Models != nil {
		g := data.Summary.Governance
		m := data.Summary.Models

		if m.TotalModels > 0 {
			// Calculate percentage of models with proper documentation
			documentedPct := 0
			if m.WithModelCard > 0 {
				documentedPct = (m.WithModelCard * 100) / m.TotalModels
			}

			licensedPct := 0
			if m.WithLicense > 0 {
				licensedPct = (m.WithLicense * 100) / m.TotalModels
			}

			switch {
			case g.BlockedLicenses > 0:
				impact.governanceMaturity = "Critical Issues"
			case documentedPct < 30 || licensedPct < 30:
				impact.governanceMaturity = "Needs Improvement"
			case documentedPct < 70 || licensedPct < 70:
				impact.governanceMaturity = "Developing"
			case documentedPct >= 90 && licensedPct >= 90:
				impact.governanceMaturity = "Mature"
			default:
				impact.governanceMaturity = "Good"
			}
		}
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
	techPath := filepath.Join(analysisDir, "technology-technical-report.md")
	if err := os.WriteFile(techPath, []byte(techReport), 0644); err != nil {
		return fmt.Errorf("writing technical report: %w", err)
	}

	// Write executive report
	execReport := GenerateExecutiveReport(data)
	execPath := filepath.Join(analysisDir, "technology-executive-report.md")
	if err := os.WriteFile(execPath, []byte(execReport), 0644); err != nil {
		return fmt.Errorf("writing executive report: %w", err)
	}

	return nil
}
