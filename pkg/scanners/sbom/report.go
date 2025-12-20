package sbom

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
	Version    string
	Summary    Summary
	Findings   Findings
}

// LoadReportData loads sbom.json from the analysis directory
func LoadReportData(analysisDir string) (*ReportData, error) {
	sbomPath := filepath.Join(analysisDir, "sbom.json")
	data, err := os.ReadFile(sbomPath)
	if err != nil {
		return nil, fmt.Errorf("reading sbom.json: %w", err)
	}

	var result struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Version    string    `json:"version"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing sbom.json: %w", err)
	}

	return &ReportData{
		Repository: result.Repository,
		Timestamp:  result.Timestamp,
		Version:    result.Version,
		Summary:    result.Summary,
		Findings:   result.Findings,
	}, nil
}

// GenerateTechnicalReport creates a detailed technical report for engineers
func GenerateTechnicalReport(data *ReportData) string {
	b := report.NewBuilder()

	// Header
	b.Title("SBOM Technical Report")
	b.Meta(report.ReportMeta{
		Repository:  data.Repository,
		Timestamp:   data.Timestamp,
		ScannerDesc: "Software Bill of Materials (SBOM) Scanner",
		Version:     data.Version,
	})

	// Generation Section
	if data.Summary.Generation != nil {
		b.Section(2, "1. SBOM Generation")
		gen := data.Summary.Generation

		if gen.Error != "" {
			b.Paragraph(fmt.Sprintf("**Status:** Failed - %s", gen.Error))
		} else {
			b.Section(3, "Summary")

			rows := [][]string{
				{"SBOM Tool", gen.Tool},
				{"Spec Version", gen.SpecVersion},
				{"Total Components", fmt.Sprintf("%d", gen.TotalComponents)},
				{"Has Dependencies", boolToYesNo(gen.HasDependencies)},
				{"SBOM Path", fmt.Sprintf("`%s`", gen.SBOMPath)},
			}
			b.Table([]string{"Metric", "Value"}, rows)

			// Component breakdown by type
			if len(gen.ByType) > 0 {
				b.Section(3, "Components by Type")
				var typeRows [][]string
				for typ, count := range gen.ByType {
					typeRows = append(typeRows, []string{typ, fmt.Sprintf("%d", count)})
				}
				sort.Slice(typeRows, func(i, j int) bool {
					return typeRows[i][0] < typeRows[j][0]
				})
				b.Table([]string{"Type", "Count"}, typeRows)
			}

			// Ecosystem breakdown
			if len(gen.ByEcosystem) > 0 {
				b.Section(3, "Components by Ecosystem")
				var ecoRows [][]string
				for eco, count := range gen.ByEcosystem {
					ecoRows = append(ecoRows, []string{eco, fmt.Sprintf("%d", count)})
				}
				// Sort by count descending
				sort.Slice(ecoRows, func(i, j int) bool {
					return ecoRows[i][1] > ecoRows[j][1]
				})
				b.Table([]string{"Ecosystem", "Count"}, ecoRows)
			}

			// Component details
			if data.Findings.Generation != nil && len(data.Findings.Generation.Components) > 0 {
				b.Section(3, "Component Details")

				// Group by scope
				scopeGroups := make(map[string][]Component)
				for _, comp := range data.Findings.Generation.Components {
					scope := comp.Scope
					if scope == "" {
						scope = "unspecified"
					}
					scopeGroups[scope] = append(scopeGroups[scope], comp)
				}

				// Display by scope
				for _, scope := range []string{"required", "optional", "dev", "unspecified"} {
					components := scopeGroups[scope]
					if len(components) == 0 {
						continue
					}

					b.Section(4, fmt.Sprintf("%s Dependencies (%d)", strings.Title(scope), len(components)))

					var compRows [][]string
					for _, comp := range components {
						licenses := strings.Join(comp.Licenses, ", ")
						if licenses == "" {
							licenses = "-"
						}
						ecosystem := comp.Ecosystem
						if ecosystem == "" {
							ecosystem = "-"
						}
						compRows = append(compRows, []string{
							comp.Name,
							comp.Version,
							ecosystem,
							comp.Type,
							licenses,
						})
					}
					// Sort by name
					sort.Slice(compRows, func(i, j int) bool {
						return compRows[i][0] < compRows[j][0]
					})
					b.Table([]string{"Name", "Version", "Ecosystem", "Type", "Licenses"}, compRows)
				}
			}

			// Dependency graph stats
			if data.Findings.Generation != nil && len(data.Findings.Generation.Dependencies) > 0 {
				b.Section(3, "Dependency Graph Statistics")

				deps := data.Findings.Generation.Dependencies
				totalEdges := 0
				maxDeps := 0
				minDeps := -1

				for _, dep := range deps {
					numDeps := len(dep.DependsOn)
					totalEdges += numDeps
					if numDeps > maxDeps {
						maxDeps = numDeps
					}
					if minDeps == -1 || numDeps < minDeps {
						minDeps = numDeps
					}
				}

				avgDeps := 0.0
				if len(deps) > 0 {
					avgDeps = float64(totalEdges) / float64(len(deps))
				}

				statsRows := [][]string{
					{"Total Nodes", fmt.Sprintf("%d", len(deps))},
					{"Total Edges", fmt.Sprintf("%d", totalEdges)},
					{"Average Dependencies per Component", fmt.Sprintf("%.1f", avgDeps)},
					{"Max Dependencies", fmt.Sprintf("%d", maxDeps)},
					{"Min Dependencies", fmt.Sprintf("%d", minDeps)},
				}
				b.Table([]string{"Metric", "Value"}, statsRows)
			}

			// SBOM Metadata
			if data.Findings.Generation != nil && data.Findings.Generation.Metadata != nil {
				b.Section(3, "SBOM Metadata")
				meta := data.Findings.Generation.Metadata
				metaRows := [][]string{
					{"BOM Format", meta.BomFormat},
					{"Spec Version", meta.SpecVersion},
					{"SBOM Version", fmt.Sprintf("%d", meta.Version)},
					{"Serial Number", meta.SerialNumber},
					{"Timestamp", meta.Timestamp},
					{"Tool", meta.Tool},
				}
				b.Table([]string{"Property", "Value"}, metaRows)
			}
		}
	}

	// Integrity Section
	if data.Summary.Integrity != nil {
		b.Section(2, "2. SBOM Integrity Verification")
		integrity := data.Summary.Integrity

		if integrity.Error != "" {
			b.Paragraph(fmt.Sprintf("**Status:** Failed - %s", integrity.Error))
		} else {
			b.Section(3, "Summary")

			status := "Complete"
			if !integrity.IsComplete {
				status = "Incomplete"
			}

			driftStatus := "No drift detected"
			if integrity.DriftDetected {
				driftStatus = fmt.Sprintf("Drift detected (%d missing, %d extra)",
					integrity.MissingPackages, integrity.ExtraPackages)
			}

			rows := [][]string{
				{"Completeness Status", status},
				{"Drift Status", driftStatus},
				{"Lockfiles Found", fmt.Sprintf("%d", integrity.LockfilesFound)},
				{"Missing Packages", fmt.Sprintf("%d", integrity.MissingPackages)},
				{"Extra Packages", fmt.Sprintf("%d", integrity.ExtraPackages)},
			}
			b.Table([]string{"Metric", "Value"}, rows)

			// Lockfile comparisons
			if data.Findings.Integrity != nil && len(data.Findings.Integrity.LockfileComparisons) > 0 {
				b.Section(3, "Lockfile Comparisons")
				var compRows [][]string
				for _, comp := range data.Findings.Integrity.LockfileComparisons {
					matchRate := "0%"
					if comp.InLockfile > 0 {
						matchRate = fmt.Sprintf("%.1f%%", float64(comp.Matched)/float64(comp.InLockfile)*100)
					}
					compRows = append(compRows, []string{
						comp.Lockfile,
						comp.Ecosystem,
						fmt.Sprintf("%d", comp.InSBOM),
						fmt.Sprintf("%d", comp.InLockfile),
						fmt.Sprintf("%d", comp.Matched),
						fmt.Sprintf("%d", comp.Missing),
						fmt.Sprintf("%d", comp.Extra),
						matchRate,
					})
				}
				b.Table([]string{"Lockfile", "Ecosystem", "In SBOM", "In Lockfile", "Matched", "Missing", "Extra", "Match Rate"}, compRows)
			}

			// Missing packages
			if data.Findings.Integrity != nil && len(data.Findings.Integrity.MissingPackages) > 0 {
				b.Section(3, "Missing Packages (in lockfile, not in SBOM)")
				var missingRows [][]string
				for _, pkg := range data.Findings.Integrity.MissingPackages {
					missingRows = append(missingRows, []string{
						pkg.Name,
						pkg.Version,
						pkg.Ecosystem,
						pkg.Lockfile,
					})
				}
				// Limit to first 50 to avoid huge reports
				if len(missingRows) > 50 {
					b.Paragraph(fmt.Sprintf("Showing first 50 of %d missing packages:", len(missingRows)))
					missingRows = missingRows[:50]
				}
				b.Table([]string{"Name", "Version", "Ecosystem", "Lockfile"}, missingRows)
			}

			// Extra packages
			if data.Findings.Integrity != nil && len(data.Findings.Integrity.ExtraPackages) > 0 {
				b.Section(3, "Extra Packages (in SBOM, not in lockfile)")
				var extraRows [][]string
				for _, pkg := range data.Findings.Integrity.ExtraPackages {
					extraRows = append(extraRows, []string{
						pkg.Name,
						pkg.Version,
						pkg.Ecosystem,
					})
				}
				// Limit to first 50
				if len(extraRows) > 50 {
					b.Paragraph(fmt.Sprintf("Showing first 50 of %d extra packages:", len(extraRows)))
					extraRows = extraRows[:50]
				}
				b.Table([]string{"Name", "Version", "Ecosystem"}, extraRows)
			}

			// Drift details
			if data.Findings.Integrity != nil && data.Findings.Integrity.DriftDetails != nil {
				drift := data.Findings.Integrity.DriftDetails
				if drift.TotalAdded > 0 || drift.TotalRemoved > 0 || drift.TotalChanged > 0 {
					b.Section(3, "SBOM Drift Analysis")

					driftRows := [][]string{
						{"Components Added", fmt.Sprintf("%d", drift.TotalAdded)},
						{"Components Removed", fmt.Sprintf("%d", drift.TotalRemoved)},
						{"Components Changed", fmt.Sprintf("%d", drift.TotalChanged)},
					}
					b.Table([]string{"Change Type", "Count"}, driftRows)

					// Added components
					if len(drift.Added) > 0 {
						b.Section(4, "Added Components")
						var addedRows [][]string
						for _, comp := range drift.Added {
							ecosystem := comp.Ecosystem
							if ecosystem == "" {
								ecosystem = "-"
							}
							addedRows = append(addedRows, []string{comp.Name, comp.Version, ecosystem})
						}
						b.Table([]string{"Name", "Version", "Ecosystem"}, addedRows)
					}

					// Removed components
					if len(drift.Removed) > 0 {
						b.Section(4, "Removed Components")
						var removedRows [][]string
						for _, comp := range drift.Removed {
							ecosystem := comp.Ecosystem
							if ecosystem == "" {
								ecosystem = "-"
							}
							removedRows = append(removedRows, []string{comp.Name, comp.Version, ecosystem})
						}
						b.Table([]string{"Name", "Version", "Ecosystem"}, removedRows)
					}

					// Version changes
					if len(drift.VersionChanged) > 0 {
						b.Section(4, "Version Changes")
						var changedRows [][]string
						for _, change := range drift.VersionChanged {
							ecosystem := change.Ecosystem
							if ecosystem == "" {
								ecosystem = "-"
							}
							changedRows = append(changedRows, []string{
								change.Name,
								change.OldVersion,
								change.NewVersion,
								ecosystem,
							})
						}
						b.Table([]string{"Name", "Old Version", "New Version", "Ecosystem"}, changedRows)
					}
				}
			}
		}
	}

	// Errors section
	if len(data.Summary.Errors) > 0 {
		b.Section(2, "Errors")
		var errorList []string
		for _, err := range data.Summary.Errors {
			errorList = append(errorList, err)
		}
		b.List(errorList)
	}

	b.Footer("SBOM")
	return b.String()
}

// GenerateExecutiveReport creates a high-level summary for engineering leaders
func GenerateExecutiveReport(data *ReportData) string {
	b := report.NewBuilder()

	// Header
	b.Title("SBOM Executive Report")
	b.Raw(fmt.Sprintf("**Repository:** `%s`\n", data.Repository))
	b.Raw(fmt.Sprintf("**Date:** %s\n\n", data.Timestamp.Format("January 2, 2006")))

	// Executive Summary
	b.Section(2, "Executive Summary")

	overallScore := calculateSBOMScore(data)
	grade := scoreToGrade(overallScore)

	b.Paragraph(fmt.Sprintf("### Overall SBOM Quality: %s (%d/100)", grade, overallScore))

	// Score breakdown
	scores := calculateFeatureScores(data)
	scoreRows := [][]string{}

	if gen := data.Summary.Generation; gen != nil {
		status := "Failed"
		if gen.Error == "" {
			status = scoreToStatus(scores.generation)
		}
		scoreRows = append(scoreRows, []string{
			b.Bold("SBOM Generation"),
			fmt.Sprintf("%d/100", scores.generation),
			status,
		})
	}

	if integrity := data.Summary.Integrity; integrity != nil {
		status := "Failed"
		if integrity.Error == "" {
			status = scoreToStatus(scores.integrity)
		}
		scoreRows = append(scoreRows, []string{
			b.Bold("SBOM Integrity"),
			fmt.Sprintf("%d/100", scores.integrity),
			status,
		})
	}

	b.Table([]string{"Area", "Score", "Status"}, scoreRows)

	// Key Metrics
	b.Section(2, "Key Metrics")

	if gen := data.Summary.Generation; gen != nil && gen.Error == "" {
		b.Section(3, "SBOM Coverage")
		items := []string{
			fmt.Sprintf("**Total Components:** %d", gen.TotalComponents),
			fmt.Sprintf("**SBOM Format:** %s %s", gen.Tool, gen.SpecVersion),
			fmt.Sprintf("**Dependency Graph:** %s", boolToYesNo(gen.HasDependencies)),
		}

		if len(gen.ByEcosystem) > 0 {
			var ecosystems []string
			for eco := range gen.ByEcosystem {
				ecosystems = append(ecosystems, eco)
			}
			items = append(items, fmt.Sprintf("**Ecosystems:** %s", strings.Join(ecosystems, ", ")))
		}

		b.List(items)
	}

	if integrity := data.Summary.Integrity; integrity != nil && integrity.Error == "" {
		b.Section(3, "SBOM Integrity")
		items := []string{
			fmt.Sprintf("**Completeness:** %s", boolToStatus(integrity.IsComplete)),
			fmt.Sprintf("**Drift Detected:** %s", boolToStatus(!integrity.DriftDetected)),
			fmt.Sprintf("**Lockfiles Verified:** %d", integrity.LockfilesFound),
		}

		if integrity.MissingPackages > 0 {
			items = append(items, fmt.Sprintf("**Missing Packages:** %d (in lockfile, not in SBOM)", integrity.MissingPackages))
		}
		if integrity.ExtraPackages > 0 {
			items = append(items, fmt.Sprintf("**Extra Packages:** %d (in SBOM, not in lockfile)", integrity.ExtraPackages))
		}

		b.List(items)
	}

	// Key Findings
	findings := collectKeyFindings(data)

	if len(findings.critical) > 0 {
		b.Section(2, "Critical Issues")
		b.List(findings.critical)
	}

	if len(findings.warnings) > 0 {
		b.Section(2, "Areas for Improvement")
		b.List(findings.warnings)
	}

	if len(findings.positive) > 0 {
		b.Section(2, "Strengths")
		b.List(findings.positive)
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

	// Business Impact
	b.Section(2, "Business Impact")
	impact := assessImpact(data)

	b.Paragraph(fmt.Sprintf("**Supply Chain Risk:** %s", impact.supplyChainRisk))
	b.Paragraph(fmt.Sprintf("**Compliance Readiness:** %s", impact.complianceReadiness))

	if impact.keyRisk != "" {
		b.Paragraph(fmt.Sprintf("**Key Risk:** %s", impact.keyRisk))
	}

	b.Footer("SBOM")
	return b.String()
}

// Helper functions

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func boolToStatus(b bool) string {
	if b {
		return "Good"
	}
	return "Issues Found"
}

type featureScores struct {
	generation int
	integrity  int
}

func calculateFeatureScores(data *ReportData) featureScores {
	scores := featureScores{}

	// Generation score
	if gen := data.Summary.Generation; gen != nil {
		if gen.Error != "" {
			scores.generation = 0
		} else {
			score := 50 // Base score

			// Has components
			if gen.TotalComponents > 0 {
				score += 20
			}

			// Has dependency graph
			if gen.HasDependencies {
				score += 15
			}

			// Multiple ecosystems (better coverage)
			if len(gen.ByEcosystem) > 1 {
				score += 10
			}

			// Good component count (not too sparse)
			if gen.TotalComponents >= 10 {
				score += 5
			}

			scores.generation = min(score, 100)
		}
	}

	// Integrity score
	if integrity := data.Summary.Integrity; integrity != nil {
		if integrity.Error != "" {
			scores.integrity = 0
		} else {
			score := 100

			// Deduct for missing packages
			if integrity.MissingPackages > 0 {
				// More severe penalty for many missing packages
				if integrity.MissingPackages > 50 {
					score -= 40
				} else if integrity.MissingPackages > 10 {
					score -= 30
				} else {
					score -= 20
				}
			}

			// Deduct for extra packages (less severe)
			if integrity.ExtraPackages > 0 {
				if integrity.ExtraPackages > 50 {
					score -= 20
				} else if integrity.ExtraPackages > 10 {
					score -= 15
				} else {
					score -= 10
				}
			}

			// Deduct for drift
			if integrity.DriftDetected {
				score -= 10
			}

			// Bonus for lockfiles found
			if integrity.LockfilesFound > 0 {
				score += 5
			}

			scores.integrity = max(score, 0)
		}
	}

	return scores
}

func calculateSBOMScore(data *ReportData) int {
	scores := calculateFeatureScores(data)
	count := 0
	total := 0

	if data.Summary.Generation != nil {
		total += scores.generation
		count++
	}
	if data.Summary.Integrity != nil {
		total += scores.integrity
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

	// Check generation findings
	if gen := data.Summary.Generation; gen != nil {
		if gen.Error != "" {
			kf.critical = append(kf.critical, fmt.Sprintf("SBOM generation failed: %s", gen.Error))
		} else {
			if gen.TotalComponents == 0 {
				kf.critical = append(kf.critical, "SBOM contains no components")
			}
			if !gen.HasDependencies {
				kf.warnings = append(kf.warnings, "SBOM does not include dependency relationships")
			}
			if gen.TotalComponents > 0 {
				kf.positive = append(kf.positive, fmt.Sprintf("SBOM generated successfully with %d components", gen.TotalComponents))
			}
			if len(gen.ByEcosystem) > 1 {
				kf.positive = append(kf.positive, fmt.Sprintf("Multi-ecosystem support (%d ecosystems)", len(gen.ByEcosystem)))
			}
		}
	}

	// Check integrity findings
	if integrity := data.Summary.Integrity; integrity != nil {
		if integrity.Error != "" {
			kf.critical = append(kf.critical, fmt.Sprintf("SBOM integrity verification failed: %s", integrity.Error))
		} else {
			if integrity.MissingPackages > 50 {
				kf.critical = append(kf.critical, fmt.Sprintf("SBOM is significantly incomplete: %d packages in lockfiles but not in SBOM", integrity.MissingPackages))
			} else if integrity.MissingPackages > 10 {
				kf.warnings = append(kf.warnings, fmt.Sprintf("SBOM missing %d packages that are in lockfiles", integrity.MissingPackages))
			} else if integrity.MissingPackages > 0 {
				kf.warnings = append(kf.warnings, fmt.Sprintf("%d packages in lockfiles but not in SBOM", integrity.MissingPackages))
			}

			if integrity.ExtraPackages > 20 {
				kf.warnings = append(kf.warnings, fmt.Sprintf("%d packages in SBOM but not in lockfiles", integrity.ExtraPackages))
			}

			if integrity.DriftDetected {
				kf.warnings = append(kf.warnings, "SBOM drift detected compared to previous version")
			}

			if integrity.IsComplete && !integrity.DriftDetected {
				kf.positive = append(kf.positive, "SBOM is complete and matches lockfiles")
			}

			if integrity.LockfilesFound > 0 {
				kf.positive = append(kf.positive, fmt.Sprintf("Verified against %d lockfile(s)", integrity.LockfilesFound))
			}
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

	// Generation recommendations
	if gen := data.Summary.Generation; gen != nil {
		if gen.Error != "" {
			rec.immediate = append(rec.immediate, "Fix SBOM generation errors to enable supply chain analysis")
		} else if gen.TotalComponents == 0 {
			rec.immediate = append(rec.immediate, "Investigate why SBOM contains no components")
		}

		if !gen.HasDependencies && gen.TotalComponents > 0 {
			rec.shortTerm = append(rec.shortTerm, "Enable dependency graph generation in SBOM tool for better vulnerability analysis")
		}
	}

	// Integrity recommendations
	if integrity := data.Summary.Integrity; integrity != nil {
		if integrity.Error != "" {
			rec.immediate = append(rec.immediate, "Fix SBOM integrity verification to ensure accuracy")
		} else {
			if integrity.MissingPackages > 50 {
				rec.immediate = append(rec.immediate, "Regenerate SBOM to include all dependencies from lockfiles")
			} else if integrity.MissingPackages > 10 {
				rec.shortTerm = append(rec.shortTerm, "Update SBOM to include missing packages for complete coverage")
			}

			if integrity.ExtraPackages > 20 {
				rec.shortTerm = append(rec.shortTerm, "Review and reconcile extra packages in SBOM with lockfiles")
			}

			if !integrity.IsComplete {
				rec.immediate = append(rec.immediate, "Ensure SBOM is regenerated after dependency changes")
			}
		}
	}

	// General recommendations
	if len(rec.immediate) == 0 && len(rec.shortTerm) == 0 {
		rec.shortTerm = append(rec.shortTerm, "Maintain SBOM freshness by regenerating after dependency updates")
		rec.shortTerm = append(rec.shortTerm, "Consider implementing automated SBOM generation in CI/CD pipeline")
	}

	return rec
}

type impactAssessment struct {
	supplyChainRisk    string
	complianceReadiness string
	keyRisk            string
}

func assessImpact(data *ReportData) impactAssessment {
	impact := impactAssessment{
		supplyChainRisk:    "Unknown",
		complianceReadiness: "Unknown",
	}

	score := calculateSBOMScore(data)

	// Supply chain risk assessment
	switch {
	case score >= 80:
		impact.supplyChainRisk = "Low - SBOM is comprehensive and accurate"
	case score >= 60:
		impact.supplyChainRisk = "Moderate - SBOM has minor gaps"
	case score >= 40:
		impact.supplyChainRisk = "High - SBOM has significant gaps"
	default:
		impact.supplyChainRisk = "Critical - SBOM is incomplete or inaccurate"
		impact.keyRisk = "Incomplete SBOM prevents effective vulnerability management and supply chain risk assessment"
	}

	// Compliance readiness (SBOM required for many frameworks)
	gen := data.Summary.Generation
	integrity := data.Summary.Integrity

	if gen != nil && gen.Error == "" && integrity != nil && integrity.Error == "" {
		if gen.TotalComponents > 0 && integrity.IsComplete {
			impact.complianceReadiness = "Ready - Complete SBOM available for compliance requirements"
		} else if gen.TotalComponents > 0 {
			impact.complianceReadiness = "Partial - SBOM exists but has integrity issues"
		} else {
			impact.complianceReadiness = "Not Ready - SBOM generation issues must be resolved"
		}
	} else {
		impact.complianceReadiness = "Not Ready - SBOM generation or verification failed"
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
	techPath := filepath.Join(analysisDir, "sbom-technical-report.md")
	if err := os.WriteFile(techPath, []byte(techReport), 0644); err != nil {
		return fmt.Errorf("writing technical report: %w", err)
	}

	// Write executive report
	execReport := GenerateExecutiveReport(data)
	execPath := filepath.Join(analysisDir, "sbom-executive-report.md")
	if err := os.WriteFile(execPath, []byte(execReport), 0644); err != nil {
		return fmt.Errorf("writing executive report: %w", err)
	}

	return nil
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
