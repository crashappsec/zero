package codequality

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReportGeneration(t *testing.T) {
	// Create test data
	data := &ReportData{
		Repository: "test/repo",
		Timestamp:  time.Now(),
		Summary: Summary{
			TechDebt: &TechDebtSummary{
				TotalMarkers:  45,
				TotalIssues:   12,
				FilesAffected: 8,
				ByType: map[string]int{
					"TODO":   20,
					"FIXME":  15,
					"HACK":   10,
				},
				ByPriority: map[string]int{
					"high":   25,
					"medium": 15,
					"low":    5,
				},
			},
			Complexity: &ComplexitySummary{
				TotalIssues:   18,
				High:          5,
				Medium:        8,
				Low:           5,
				FilesAffected: 12,
				ByType: map[string]int{
					"complexity-cyclomatic":    5,
					"complexity-long-function": 8,
					"complexity-deep-nesting":  5,
				},
			},
			TestCoverage: &TestCoverageSummary{
				HasTestFiles:   true,
				TestFrameworks: []string{"jest", "pytest"},
				CoverageReports: []string{"coverage.json", "lcov.info"},
				LineCoverage:    72.5,
				MeetsThreshold:  true,
			},
			CodeDocs: &CodeDocsSummary{
				HasReadme:    true,
				ReadmeFile:   "README.md",
				HasChangelog: true,
				HasApiDocs:   true,
				Score:        85,
			},
		},
		Findings: Findings{
			TechDebt: &TechDebtResult{
				Markers: []DebtMarker{
					{
						Type:     "FIXME",
						Priority: "high",
						File:     "src/app.go",
						Line:     42,
						Text:     "FIXME: Need to handle edge case",
					},
					{
						Type:     "TODO",
						Priority: "medium",
						File:     "src/util.go",
						Line:     100,
						Text:     "TODO: Optimize this function",
					},
				},
				Issues: []DebtIssue{
					{
						Type:        "empty-catch",
						Severity:    "high",
						File:        "src/error.go",
						Line:        55,
						Description: "Empty catch block swallows errors",
						Suggestion:  "Handle or log errors appropriately",
					},
				},
				Hotspots: []FileDebt{
					{
						File:         "src/app.go",
						TotalMarkers: 12,
						ByType: map[string]int{
							"FIXME": 7,
							"TODO":  5,
						},
					},
				},
			},
			Complexity: &ComplexityResult{
				Issues: []ComplexityIssue{
					{
						Type:        "complexity-cyclomatic",
						Severity:    "high",
						File:        "src/processor.go",
						Line:        150,
						Description: "High cyclomatic complexity detected",
						Suggestion:  "Break down into smaller functions with single responsibilities",
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

	// Verify key sections are present
	requiredSections := []string{
		"Code Quality Technical Report",
		"Technical Debt Analysis",
		"Complexity Analysis",
		"Test Coverage Analysis",
		"Documentation Analysis",
	}

	for _, section := range requiredSections {
		if !contains(techReport, section) {
			t.Errorf("Technical report missing section: %s", section)
		}
	}

	// Test executive report generation
	execReport := GenerateExecutiveReport(data)
	if len(execReport) == 0 {
		t.Error("Executive report is empty")
	}

	// Verify key sections are present
	execSections := []string{
		"Code Quality Executive Report",
		"Executive Summary",
		"Overall Code Health",
		"Key Metrics",
		"Recommendations",
		"Business Impact",
	}

	for _, section := range execSections {
		if !contains(execReport, section) {
			t.Errorf("Executive report missing section: %s", section)
		}
	}
}

func TestLoadReportData(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test data
	testResult := struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}{
		Repository: "test/repo",
		Timestamp:  time.Now(),
		Summary: Summary{
			TechDebt: &TechDebtSummary{
				TotalMarkers: 10,
				TotalIssues:  5,
			},
		},
		Findings: Findings{},
	}

	// Write test JSON
	data, err := json.MarshalIndent(testResult, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	qualityFile := filepath.Join(tmpDir, "code-quality.json")
	if err := os.WriteFile(qualityFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test loading
	reportData, err := LoadReportData(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load report data: %v", err)
	}

	if reportData.Repository != "test/repo" {
		t.Errorf("Expected repository 'test/repo', got '%s'", reportData.Repository)
	}

	if reportData.Summary.TechDebt.TotalMarkers != 10 {
		t.Errorf("Expected 10 markers, got %d", reportData.Summary.TechDebt.TotalMarkers)
	}
}

func TestWriteReports(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create minimal test data
	testResult := struct {
		Repository string    `json:"repository"`
		Timestamp  time.Time `json:"timestamp"`
		Summary    Summary   `json:"summary"`
		Findings   Findings  `json:"findings"`
	}{
		Repository: "test/repo",
		Timestamp:  time.Now(),
		Summary: Summary{
			TechDebt: &TechDebtSummary{
				TotalMarkers:  5,
				TotalIssues:   2,
				FilesAffected: 3,
				ByType:        map[string]int{"TODO": 5},
				ByPriority:    map[string]int{"medium": 5},
			},
		},
		Findings: Findings{},
	}

	// Write test JSON
	data, err := json.MarshalIndent(testResult, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	qualityFile := filepath.Join(tmpDir, "code-quality.json")
	if err := os.WriteFile(qualityFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test report writing
	if err := WriteReports(tmpDir); err != nil {
		t.Fatalf("Failed to write reports: %v", err)
	}

	// Verify files were created
	techPath := filepath.Join(tmpDir, "code-quality-technical-report.md")
	if _, err := os.Stat(techPath); err != nil {
		t.Errorf("Technical report not created: %v", err)
	}

	execPath := filepath.Join(tmpDir, "code-quality-executive-report.md")
	if _, err := os.Stat(execPath); err != nil {
		t.Errorf("Executive report not created: %v", err)
	}

	// Verify content is not empty
	techData, err := os.ReadFile(techPath)
	if err != nil {
		t.Errorf("Failed to read technical report: %v", err)
	}
	if len(techData) == 0 {
		t.Error("Technical report is empty")
	}

	execData, err := os.ReadFile(execPath)
	if err != nil {
		t.Errorf("Failed to read executive report: %v", err)
	}
	if len(execData) == 0 {
		t.Error("Executive report is empty")
	}
}

func TestScoreCalculations(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		expected string
	}{
		{"A grade", 95, "A"},
		{"B grade", 85, "B"},
		{"C grade", 75, "C"},
		{"D grade", 65, "D"},
		{"F grade", 45, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grade := scoreToGrade(tt.score)
			if grade != tt.expected {
				t.Errorf("scoreToGrade(%d) = %s, want %s", tt.score, grade, tt.expected)
			}
		})
	}
}

func TestDebtScoreCalculation(t *testing.T) {
	tests := []struct {
		name     string
		summary  *TechDebtSummary
		minScore int
		maxScore int
	}{
		{
			name: "no debt",
			summary: &TechDebtSummary{
				TotalMarkers:  0,
				TotalIssues:   0,
				FilesAffected: 0,
			},
			minScore: 95,
			maxScore: 100,
		},
		{
			name: "high debt",
			summary: &TechDebtSummary{
				TotalMarkers:  200,
				TotalIssues:   50,
				FilesAffected: 30,
				ByPriority: map[string]int{
					"high": 50,
				},
			},
			minScore: 0,
			maxScore: 20,
		},
		{
			name: "moderate debt",
			summary: &TechDebtSummary{
				TotalMarkers:  50,
				TotalIssues:   10,
				FilesAffected: 8,
				ByPriority: map[string]int{
					"high":   5,
					"medium": 30,
					"low":    15,
				},
			},
			minScore: 40,
			maxScore: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateDebtScore(tt.summary)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateDebtScore() = %d, want between %d and %d", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestComplexityScoreCalculation(t *testing.T) {
	tests := []struct {
		name     string
		summary  *ComplexitySummary
		expected int
	}{
		{
			name: "no issues",
			summary: &ComplexitySummary{
				TotalIssues: 0,
			},
			expected: 100,
		},
		{
			name: "high severity issues",
			summary: &ComplexitySummary{
				TotalIssues: 10,
				High:        10,
			},
			expected: 50, // 100 - (10 * 5)
		},
		{
			name: "mixed severity",
			summary: &ComplexitySummary{
				TotalIssues: 15,
				High:        5,
				Medium:      5,
				Low:         5,
			},
			expected: 60, // 100 - (5*5) - (5*2) - 5 = 100 - 25 - 10 - 5 = 60
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateComplexityScore(tt.summary)
			if score != tt.expected {
				t.Errorf("calculateComplexityScore() = %d, want %d", score, tt.expected)
			}
		})
	}
}

func TestCoverageScoreCalculation(t *testing.T) {
	tests := []struct {
		name     string
		summary  *TestCoverageSummary
		expected int
	}{
		{
			name: "no tests",
			summary: &TestCoverageSummary{
				HasTestFiles: false,
			},
			expected: 0,
		},
		{
			name: "tests without coverage",
			summary: &TestCoverageSummary{
				HasTestFiles: true,
				LineCoverage: 0,
			},
			expected: 50,
		},
		{
			name: "high coverage",
			summary: &TestCoverageSummary{
				HasTestFiles: true,
				LineCoverage: 85.5,
			},
			expected: 85,
		},
		{
			name: "full coverage",
			summary: &TestCoverageSummary{
				HasTestFiles: true,
				LineCoverage: 100,
			},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateCoverageScore(tt.summary)
			if score != tt.expected {
				t.Errorf("calculateCoverageScore() = %d, want %d", score, tt.expected)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
