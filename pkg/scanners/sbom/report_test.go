package sbom

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReportGeneration(t *testing.T) {
	// Create sample data
	data := &ReportData{
		Repository: "test/repo",
		Timestamp:  time.Now(),
		Version:    "1.0.0",
		Summary: Summary{
			Generation: &GenerationSummary{
				Tool:            "cdxgen",
				SpecVersion:     "1.5",
				TotalComponents: 150,
				ByType: map[string]int{
					"library": 145,
					"application": 5,
				},
				ByEcosystem: map[string]int{
					"npm": 100,
					"pypi": 50,
				},
				HasDependencies: true,
				SBOMPath:        "/path/to/sbom.cdx.json",
			},
			Integrity: &IntegritySummary{
				IsComplete:      true,
				DriftDetected:   false,
				MissingPackages: 0,
				ExtraPackages:   2,
				LockfilesFound:  2,
			},
		},
		Findings: Findings{
			Generation: &GenerationFindings{
				Components: []Component{
					{
						Type:      "library",
						Name:      "express",
						Version:   "4.18.2",
						Purl:      "pkg:npm/express@4.18.2",
						Ecosystem: "npm",
						Licenses:  []string{"MIT"},
						Scope:     "required",
					},
					{
						Type:      "library",
						Name:      "lodash",
						Version:   "4.17.21",
						Purl:      "pkg:npm/lodash@4.17.21",
						Ecosystem: "npm",
						Licenses:  []string{"MIT"},
						Scope:     "required",
					},
				},
				Dependencies: []Dependency{
					{
						Ref:       "pkg:npm/express@4.18.2",
						DependsOn: []string{"pkg:npm/body-parser@1.20.1"},
					},
				},
				Metadata: &SBOMMetadata{
					BomFormat:   "CycloneDX",
					SpecVersion: "1.5",
					Version:     1,
					SerialNumber: "urn:uuid:test-123",
					Timestamp:   "2025-12-19T00:00:00Z",
					Tool:        "cdxgen",
				},
			},
			Integrity: &IntegrityFindings{
				LockfileComparisons: []LockfileComparison{
					{
						Lockfile:   "package-lock.json",
						Ecosystem:  "npm",
						InSBOM:     100,
						InLockfile: 100,
						Matched:    98,
						Missing:    0,
						Extra:      2,
					},
				},
			},
		},
	}

	// Test technical report generation
	techReport := GenerateTechnicalReport(data)
	if len(techReport) == 0 {
		t.Error("Technical report is empty")
	}
	if !contains(techReport, "SBOM Technical Report") {
		t.Error("Technical report missing title")
	}
	if !contains(techReport, "express") {
		t.Error("Technical report missing component details")
	}

	// Test executive report generation
	execReport := GenerateExecutiveReport(data)
	if len(execReport) == 0 {
		t.Error("Executive report is empty")
	}
	if !contains(execReport, "SBOM Executive Report") {
		t.Error("Executive report missing title")
	}
	if !contains(execReport, "Executive Summary") {
		t.Error("Executive report missing executive summary")
	}

	// Test score calculation
	score := calculateSBOMScore(data)
	if score < 0 || score > 100 {
		t.Errorf("Invalid score: %d (should be 0-100)", score)
	}
}

func TestReportGenerationWithErrors(t *testing.T) {
	// Create data with errors
	data := &ReportData{
		Repository: "test/repo",
		Timestamp:  time.Now(),
		Version:    "1.0.0",
		Summary: Summary{
			Generation: &GenerationSummary{
				Error: "Failed to parse SBOM",
			},
			Integrity: &IntegritySummary{
				IsComplete:      false,
				DriftDetected:   true,
				MissingPackages: 100,
				ExtraPackages:   10,
				LockfilesFound:  1,
			},
			Errors: []string{
				"generation: Failed to parse SBOM",
			},
		},
	}

	// Test technical report handles errors
	techReport := GenerateTechnicalReport(data)
	if !contains(techReport, "Failed to parse SBOM") {
		t.Error("Technical report should show error message")
	}

	// Test executive report handles errors
	execReport := GenerateExecutiveReport(data)
	if !contains(execReport, "Critical") {
		t.Error("Executive report should show critical status for errors")
	}

	// Test score is low with errors
	score := calculateSBOMScore(data)
	if score > 50 {
		t.Errorf("Score too high for error case: %d (expected <= 50)", score)
	}
}

func TestWriteReports(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "sbom-report-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sample sbom.json
	sbomData := `{
		"repository": "test/repo",
		"timestamp": "2025-12-19T00:00:00Z",
		"version": "1.0.0",
		"summary": {
			"generation": {
				"tool": "cdxgen",
				"spec_version": "1.5",
				"total_components": 10,
				"by_type": {"library": 10},
				"by_ecosystem": {"npm": 10},
				"has_dependencies": true,
				"sbom_path": "/path/to/sbom.cdx.json"
			},
			"integrity": {
				"is_complete": true,
				"drift_detected": false,
				"missing_packages": 0,
				"extra_packages": 0,
				"lockfiles_found": 1
			}
		},
		"findings": {
			"generation": {
				"components": [],
				"dependencies": []
			},
			"integrity": {}
		}
	}`

	sbomPath := filepath.Join(tmpDir, "sbom.json")
	if err := os.WriteFile(sbomPath, []byte(sbomData), 0644); err != nil {
		t.Fatalf("Failed to write test sbom.json: %v", err)
	}

	// Test WriteReports
	if err := WriteReports(tmpDir); err != nil {
		t.Fatalf("WriteReports failed: %v", err)
	}

	// Check technical report exists
	techPath := filepath.Join(tmpDir, "sbom-technical-report.md")
	if _, err := os.Stat(techPath); os.IsNotExist(err) {
		t.Error("Technical report file not created")
	}

	// Check executive report exists
	execPath := filepath.Join(tmpDir, "sbom-executive-report.md")
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		t.Error("Executive report file not created")
	}

	// Read and verify content
	techContent, err := os.ReadFile(techPath)
	if err != nil {
		t.Fatalf("Failed to read technical report: %v", err)
	}
	if !contains(string(techContent), "SBOM Technical Report") {
		t.Error("Technical report has wrong content")
	}

	execContent, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("Failed to read executive report: %v", err)
	}
	if !contains(string(execContent), "SBOM Executive Report") {
		t.Error("Executive report has wrong content")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
