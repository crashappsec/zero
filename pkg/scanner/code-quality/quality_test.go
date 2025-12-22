package codequality

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/crashappsec/zero/pkg/scanner"
)

func TestQualityScanner_Name(t *testing.T) {
	s := &QualityScanner{}
	if s.Name() != "code-quality" {
		t.Errorf("Name() = %q, want %q", s.Name(), "code-quality")
	}
}

func TestQualityScanner_Description(t *testing.T) {
	s := &QualityScanner{}
	desc := s.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestQualityScanner_Dependencies(t *testing.T) {
	s := &QualityScanner{}
	deps := s.Dependencies()
	if deps != nil {
		t.Errorf("Dependencies() = %v, want nil", deps)
	}
}

func TestQualityScanner_EstimateDuration(t *testing.T) {
	s := &QualityScanner{}

	tests := []struct {
		fileCount int
		wantMin   int
	}{
		{0, 10},
		{500, 10},
		{1000, 10},
		{5000, 10},
	}

	for _, tt := range tests {
		got := s.EstimateDuration(tt.fileCount)
		if got.Seconds() < float64(tt.wantMin) {
			t.Errorf("EstimateDuration(%d) = %v, want at least %ds", tt.fileCount, got, tt.wantMin)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.TechDebt.Enabled {
		t.Error("TechDebt should be enabled by default")
	}
	if !cfg.Complexity.Enabled {
		t.Error("Complexity should be enabled by default")
	}
	if !cfg.TestCoverage.Enabled {
		t.Error("TestCoverage should be enabled by default")
	}
	if !cfg.CodeDocs.Enabled {
		t.Error("CodeDocs should be enabled by default")
	}
	if !cfg.TechDebt.IncludeMarkers {
		t.Error("TechDebt.IncludeMarkers should be enabled by default")
	}
	if !cfg.TechDebt.IncludeIssues {
		t.Error("TechDebt.IncludeIssues should be enabled by default")
	}
}

func TestMarkerPatterns(t *testing.T) {
	// Verify pattern array is not empty
	if len(markerPatterns) == 0 {
		t.Error("markerPatterns should not be empty")
	}

	// Verify each pattern has required fields
	for i, pat := range markerPatterns {
		if pat.pattern == nil {
			t.Errorf("markerPatterns[%d].pattern is nil", i)
		}
		if pat.typ == "" {
			t.Errorf("markerPatterns[%d].typ is empty", i)
		}
		if pat.priority == "" {
			t.Errorf("markerPatterns[%d].priority is empty", i)
		}
	}

	// Test that patterns match expected strings
	tests := []struct {
		input    string
		wantType string
	}{
		{"// TODO: fix this", "TODO"},
		{"# FIXME: broken", "FIXME"},
		{"/* HACK: workaround */", "HACK"},
		{"// XXX: dangerous", "XXX"},
		{"// BUG: known issue", "BUG"},
		{"// WORKAROUND for issue #123", "WORKAROUND"},
		{"// REFACTOR this later", "REFACTOR"},
		{"// NOTE: important detail", "NOTE"},
		{"// IDEA: could use caching", "IDEA"},
	}

	for _, tt := range tests {
		found := false
		for _, mp := range markerPatterns {
			if mp.pattern.MatchString(tt.input) && mp.typ == tt.wantType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Pattern for %q not matched (expected type %q)", tt.input, tt.wantType)
		}
	}
}

func TestIssuePatterns(t *testing.T) {
	// Verify pattern array is not empty
	if len(issuePatterns) == 0 {
		t.Error("issuePatterns should not be empty")
	}

	// Verify each pattern has required fields
	for i, pat := range issuePatterns {
		if pat.pattern == nil {
			t.Errorf("issuePatterns[%d].pattern is nil", i)
		}
		if pat.typ == "" {
			t.Errorf("issuePatterns[%d].typ is empty", i)
		}
		if pat.severity == "" {
			t.Errorf("issuePatterns[%d].severity is empty", i)
		}
	}

	// Test that patterns match expected strings
	tests := []struct {
		input    string
		wantType string
	}{
		{"@deprecated", "deprecated-usage"},
		{"// eslint-disable-next-line", "suppressed-warning"},
		{"# noqa", "suppressed-warning"},
		{"console.log('debug')", "debug-statement"},
		{"console.debug(foo)", "debug-statement"},
		{"sleep(1000)", "hardcoded-delay"},
		{"delay(500)", "hardcoded-delay"},
		{"} catch(e) {}", "empty-catch"},
		{"// DISABLED test", "disabled-test"},
		{"process.exit(1)", "hard-exit"},
		{"os.exit(0)", "hard-exit"},
	}

	for _, tt := range tests {
		found := false
		for _, ip := range issuePatterns {
			if ip.pattern.MatchString(tt.input) && ip.typ == tt.wantType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Pattern for %q not matched (expected type %q)", tt.input, tt.wantType)
		}
	}
}

func TestScanExtensions(t *testing.T) {
	expectedExtensions := []string{".go", ".py", ".js", ".ts", ".tsx", ".java", ".rb", ".php", ".rs"}
	for _, ext := range expectedExtensions {
		if !scanExtensions[ext] {
			t.Errorf("scanExtensions should include %q", ext)
		}
	}
}

func TestCategorizeComplexityIssue(t *testing.T) {
	tests := []struct {
		checkID  string
		message  string
		expected string
	}{
		{"cyclomatic-check", "High cyclomatic complexity", "complexity-cyclomatic"},
		{"complex-function", "Complex function detected", "complexity-cyclomatic"},
		{"long-function", "Function is too long", "complexity-long-function"},
		{"too-many-lines", "Too many lines in function", "complexity-long-function"},
		{"deep-nesting", "Deeply nested code", "complexity-deep-nesting"},
		{"nested-if", "Nested conditions", "complexity-deep-nesting"},
		{"many-params", "Too many parameters", "complexity-too-many-params"},
		{"too-many-arguments", "Too many arguments", "complexity-too-many-params"},
		{"cognitive-check", "High cognitive score", "complexity-cognitive"},
		{"unknown-check", "Some other issue", "complexity-general"},
	}

	for _, tt := range tests {
		got := categorizeComplexityIssue(tt.checkID, tt.message)
		if got != tt.expected {
			t.Errorf("categorizeComplexityIssue(%q, %q) = %q, want %q", tt.checkID, tt.message, got, tt.expected)
		}
	}
}

func TestGetComplexitySuggestion(t *testing.T) {
	tests := []struct {
		issueType string
		wantEmpty bool
	}{
		{"complexity-cyclomatic", false},
		{"complexity-long-function", false},
		{"complexity-deep-nesting", false},
		{"complexity-too-many-params", false},
		{"complexity-cognitive", false},
		{"complexity-general", false},
		{"unknown-type", false}, // Should return default suggestion
	}

	for _, tt := range tests {
		got := getComplexitySuggestion(tt.issueType)
		if (got == "") != tt.wantEmpty {
			t.Errorf("getComplexitySuggestion(%q) = %q, wantEmpty=%v", tt.issueType, got, tt.wantEmpty)
		}
	}
}

func TestParseCoverageReport(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "coverage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test Go coverage file
	goCoverage := `mode: set
github.com/example/pkg/file.go:10.1,15.2 1 1
github.com/example/pkg/file.go:16.1,20.2 1 0
`
	goCovPath := filepath.Join(tmpDir, "coverage.out")
	if err := os.WriteFile(goCovPath, []byte(goCoverage), 0644); err != nil {
		t.Fatalf("Failed to write Go coverage file: %v", err)
	}

	gotGo := parseCoverageReport(goCovPath)
	if gotGo < 0 {
		t.Errorf("parseCoverageReport(go) = %v, should be >= 0", gotGo)
	}

	// Test lcov file
	lcovContent := `SF:src/file.js
LF:100
LH:80
end_of_record
`
	lcovPath := filepath.Join(tmpDir, "lcov.info")
	if err := os.WriteFile(lcovPath, []byte(lcovContent), 0644); err != nil {
		t.Fatalf("Failed to write lcov file: %v", err)
	}

	gotLcov := parseCoverageReport(lcovPath)
	if gotLcov != 80.0 {
		t.Errorf("parseCoverageReport(lcov) = %v, want 80.0", gotLcov)
	}

	// Test non-existent file
	gotNonExistent := parseCoverageReport("/nonexistent/path")
	if gotNonExistent != -1 {
		t.Errorf("parseCoverageReport(nonexistent) = %v, want -1", gotNonExistent)
	}
}

func TestScanForDebt(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "debt-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file with debt markers
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

// TODO: implement this function
func doSomething() {
    // FIXME: this is broken
    fmt.Println("hello")
}

// HACK: workaround for bug
func workaround() {
    console.log("debug")  // debug statement
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var markers []DebtMarker
	var issues []DebtIssue
	fileStats := make(map[string]*FileDebt)

	cfg := TechDebtConfig{
		IncludeMarkers: true,
		IncludeIssues:  true,
	}

	err = scanForDebt(tmpDir, &markers, &issues, fileStats, cfg)
	if err != nil {
		t.Fatalf("scanForDebt() error = %v", err)
	}

	// Should find TODO, FIXME, HACK markers
	if len(markers) < 3 {
		t.Errorf("scanForDebt() found %d markers, want at least 3", len(markers))
	}

	// Should find console.log debug statement
	foundDebug := false
	for _, iss := range issues {
		if iss.Type == "debug-statement" {
			foundDebug = true
			break
		}
	}
	if !foundDebug {
		t.Error("scanForDebt() should detect console.log as debug-statement")
	}
}

func TestRunCodeDocs(t *testing.T) {
	// Create temp directory with doc files
	tmpDir, err := os.MkdirTemp("", "docs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create README
	readmeContent := `# My Project

## Installation

Run npm install

## Usage

Some usage instructions here.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readmeContent), 0644); err != nil {
		t.Fatalf("Failed to write README: %v", err)
	}

	// Create CHANGELOG
	if err := os.WriteFile(filepath.Join(tmpDir, "CHANGELOG.md"), []byte("# Changelog"), 0644); err != nil {
		t.Fatalf("Failed to write CHANGELOG: %v", err)
	}

	// Create docs directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "docs"), 0755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	s := &QualityScanner{}
	cfg := CodeDocsConfig{
		CheckReadme:    true,
		CheckChangelog: true,
		CheckApiDocs:   true,
	}

	opts := &scanner.ScanOptions{
		RepoPath: tmpDir,
	}

	summary := s.runCodeDocs(nil, opts, cfg)

	if !summary.HasReadme {
		t.Error("runCodeDocs() should detect README.md")
	}
	if !summary.HasChangelog {
		t.Error("runCodeDocs() should detect CHANGELOG.md")
	}
	if !summary.HasApiDocs {
		t.Error("runCodeDocs() should detect docs directory")
	}
	if summary.Score == 0 {
		t.Error("runCodeDocs() score should be > 0")
	}
}

func TestRunTestCoverage(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "testcov-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "app.test.js"), []byte("test()"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create coverage report
	lcovContent := `SF:src/file.js
LF:100
LH:75
end_of_record
`
	if err := os.WriteFile(filepath.Join(tmpDir, "lcov.info"), []byte(lcovContent), 0644); err != nil {
		t.Fatalf("Failed to write lcov file: %v", err)
	}

	s := &QualityScanner{}
	cfg := TestCoverageConfig{
		ParseReports:     true,
		MinimumThreshold: 80,
	}

	opts := &scanner.ScanOptions{
		RepoPath: tmpDir,
	}

	summary := s.runTestCoverage(nil, opts, cfg)

	if !summary.HasTestFiles {
		t.Error("runTestCoverage() should detect test files")
	}

	if len(summary.TestFrameworks) == 0 {
		t.Error("runTestCoverage() should detect test frameworks")
	}

	// Check if Go test framework detected
	hasGoTest := false
	hasJest := false
	for _, fw := range summary.TestFrameworks {
		if fw == "go-test" {
			hasGoTest = true
		}
		if fw == "jest" {
			hasJest = true
		}
	}
	if !hasGoTest {
		t.Error("runTestCoverage() should detect go-test framework")
	}
	if !hasJest {
		t.Error("runTestCoverage() should detect jest framework")
	}

	if len(summary.CoverageReports) == 0 {
		t.Error("runTestCoverage() should detect coverage reports")
	}
}
