package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePatternMarkdown_DockerPatterns(t *testing.T) {
	// Create temp directory with test pattern file
	tmpDir := t.TempDir()

	dockerPatterns := `# Dockerfile Best Practices

**Category**: devops/docker
**Description**: Dockerfile security and engineering best practices
**CWE**: CWE-250 (Execution with Unnecessary Privileges)

---

## Security Patterns

### Using :latest Tag
**Type**: regex
**Severity**: medium
**Pattern**: ` + "`(?i)^FROM\\s+[^:]+:latest\\s*$`" + `
- Using :latest tag makes builds non-reproducible
- Example: ` + "`FROM node:latest`" + `
- Remediation: Use specific version tags

### Running as Root
**Type**: regex
**Severity**: high
**Pattern**: ` + "`(?i)^USER\\s+root\\s*$`" + `
- Running container as root is a security risk
- Example: ` + "`USER root`" + `
- Remediation: Use a non-root user

---

## Best Practice Patterns

### Missing HEALTHCHECK
**Type**: regex
**Severity**: info
**Pattern**: ` + "`^(?!.*HEALTHCHECK).*$`" + `
- HEALTHCHECK enables container health monitoring
- Remediation: Add HEALTHCHECK CMD
`
	patternFile := filepath.Join(tmpDir, "patterns.md")
	if err := os.WriteFile(patternFile, []byte(dockerPatterns), 0644); err != nil {
		t.Fatalf("Failed to write pattern file: %v", err)
	}

	parsed, err := ParsePatternMarkdown(patternFile)
	if err != nil {
		t.Fatalf("ParsePatternMarkdown() error = %v", err)
	}

	// Verify file-level metadata
	if parsed.Category != "devops/docker" {
		t.Errorf("Category = %q, want %q", parsed.Category, "devops/docker")
	}
	if parsed.Description != "Dockerfile security and engineering best practices" {
		t.Errorf("Description = %q, want %q", parsed.Description, "Dockerfile security and engineering best practices")
	}
	if !strings.Contains(parsed.CWE, "CWE-250") {
		t.Errorf("CWE = %q, should contain CWE-250", parsed.CWE)
	}

	// Verify patterns were parsed
	if len(parsed.Patterns) != 3 {
		t.Errorf("Got %d patterns, want 3", len(parsed.Patterns))
	}

	// Verify first pattern
	if len(parsed.Patterns) >= 1 {
		p := parsed.Patterns[0]
		if p.Name != "Using :latest Tag" {
			t.Errorf("Pattern[0].Name = %q, want %q", p.Name, "Using :latest Tag")
		}
		if p.Type != "regex" {
			t.Errorf("Pattern[0].Type = %q, want %q", p.Type, "regex")
		}
		if p.Severity != "medium" {
			t.Errorf("Pattern[0].Severity = %q, want %q", p.Severity, "medium")
		}
		if p.Pattern == "" {
			t.Error("Pattern[0].Pattern should not be empty")
		}
	}

	// Verify second pattern
	if len(parsed.Patterns) >= 2 {
		p := parsed.Patterns[1]
		if p.Name != "Running as Root" {
			t.Errorf("Pattern[1].Name = %q, want %q", p.Name, "Running as Root")
		}
		if p.Severity != "high" {
			t.Errorf("Pattern[1].Severity = %q, want %q", p.Severity, "high")
		}
	}
}

func TestParsePatternMarkdown_APIQualityPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	apiPatterns := `# API Quality Patterns

**Category**: code-quality/api
**Description**: API design and performance patterns
**CWE**: CWE-400 (Uncontrolled Resource Consumption)

---

## API Performance Patterns

### N+1 Query Pattern
**Type**: regex
**Severity**: high
**Pattern**: ` + "`(?:for|forEach|map)\\s*\\([^)]+\\)\\s*(?:=>|{)[^}]*(?:await|\\.then)[^}]*(?:findOne|findById)`" + `
- Potential N+1 query pattern
- Remediation: Use batch queries

### Missing Pagination
**Type**: regex
**Severity**: medium
**Pattern**: ` + "`\\.find\\s*\\(\\s*\\{\\s*\\}\\s*\\)`" + `
- Unbounded query
- Remediation: Add limit and skip
`
	patternFile := filepath.Join(tmpDir, "patterns.md")
	if err := os.WriteFile(patternFile, []byte(apiPatterns), 0644); err != nil {
		t.Fatalf("Failed to write pattern file: %v", err)
	}

	parsed, err := ParsePatternMarkdown(patternFile)
	if err != nil {
		t.Fatalf("ParsePatternMarkdown() error = %v", err)
	}

	if parsed.Category != "code-quality/api" {
		t.Errorf("Category = %q, want %q", parsed.Category, "code-quality/api")
	}

	if len(parsed.Patterns) != 2 {
		t.Errorf("Got %d patterns, want 2", len(parsed.Patterns))
	}
}

func TestParsePatternMarkdown_FileNotFound(t *testing.T) {
	_, err := ParsePatternMarkdown("/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParsePatternMarkdown_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	patternFile := filepath.Join(tmpDir, "empty.md")
	if err := os.WriteFile(patternFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write pattern file: %v", err)
	}

	parsed, err := ParsePatternMarkdown(patternFile)
	if err != nil {
		t.Fatalf("ParsePatternMarkdown() error = %v", err)
	}

	if len(parsed.Patterns) != 0 {
		t.Errorf("Got %d patterns, want 0 for empty file", len(parsed.Patterns))
	}
}

func TestConvertPatternsToSemgrep(t *testing.T) {
	parsed := &ParsedPatternFile{
		Category:    "devops/docker",
		Description: "Dockerfile patterns",
		CWE:         "CWE-250",
		Patterns: []PatternRule{
			{
				Name:        "Using Latest Tag",
				Type:        "regex",
				Severity:    "medium",
				Pattern:     `(?i)^FROM\s+[^:]+:latest\s*$`,
				Description: "Using :latest tag is non-reproducible",
				Remediation: "Use specific version tags",
			},
			{
				Name:        "Running as Root",
				Type:        "regex",
				Severity:    "high",
				Pattern:     `(?i)^USER\s+root\s*$`,
				Description: "Running as root is a security risk",
				Remediation: "Use non-root user",
			},
		},
	}

	rules := ConvertPatternsToSemgrep(parsed, "devops.docker")

	if len(rules.Rules) != 2 {
		t.Fatalf("Got %d rules, want 2", len(rules.Rules))
	}

	// Verify first rule
	r := rules.Rules[0]
	if !strings.Contains(r.ID, "devops.docker") {
		t.Errorf("Rule ID %q should contain 'devops.docker'", r.ID)
	}
	if r.Severity != "WARNING" {
		t.Errorf("Severity = %q, want WARNING (mapped from medium)", r.Severity)
	}
	if r.PatternRegex == "" {
		t.Error("PatternRegex should not be empty for regex type")
	}
	if r.Message != "Using :latest tag is non-reproducible" {
		t.Errorf("Message = %q, want description", r.Message)
	}

	// Verify metadata
	if r.Metadata["category"] != "devops/docker" {
		t.Errorf("Metadata category = %v, want devops/docker", r.Metadata["category"])
	}
	if r.Metadata["cwe"] != "CWE-250" {
		t.Errorf("Metadata cwe = %v, want CWE-250", r.Metadata["cwe"])
	}

	// Verify second rule (high severity maps to ERROR)
	r2 := rules.Rules[1]
	if r2.Severity != "ERROR" {
		t.Errorf("High severity should map to ERROR, got %q", r2.Severity)
	}
}

func TestConvertPatternsToSemgrep_EmptyPatterns(t *testing.T) {
	parsed := &ParsedPatternFile{
		Category: "test",
		Patterns: []PatternRule{},
	}

	rules := ConvertPatternsToSemgrep(parsed, "test")

	if len(rules.Rules) != 0 {
		t.Errorf("Got %d rules, want 0 for empty patterns", len(rules.Rules))
	}
}

func TestConvertPatternsToSemgrep_SkipsEmptyPatternRegex(t *testing.T) {
	parsed := &ParsedPatternFile{
		Category: "test",
		Patterns: []PatternRule{
			{Name: "Valid", Type: "regex", Pattern: ".*", Severity: "low"},
			{Name: "Invalid", Type: "regex", Pattern: "", Severity: "low"}, // Empty pattern
		},
	}

	rules := ConvertPatternsToSemgrep(parsed, "test")

	if len(rules.Rules) != 1 {
		t.Errorf("Got %d rules, want 1 (skip empty pattern)", len(rules.Rules))
	}
}

func TestGenerateRulesFromRAG(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	ragDir := filepath.Join(tmpDir, "rag")
	outputDir := filepath.Join(tmpDir, "output")

	// Create RAG pattern file
	categoryDir := filepath.Join(ragDir, "devops", "docker")
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		t.Fatalf("Failed to create category dir: %v", err)
	}

	patternContent := `# Docker Patterns

**Category**: devops/docker
**Description**: Docker best practices

---

## Patterns

### Test Pattern
**Type**: regex
**Severity**: high
**Pattern**: ` + "`test-pattern`" + `
- Test description
`
	patternFile := filepath.Join(categoryDir, "patterns.md")
	if err := os.WriteFile(patternFile, []byte(patternContent), 0644); err != nil {
		t.Fatalf("Failed to write pattern file: %v", err)
	}

	// Generate rules
	outputPath := filepath.Join(outputDir, "docker.yaml")
	err := GenerateRulesFromRAG(ragDir, "devops/docker", outputPath)
	if err != nil {
		t.Fatalf("GenerateRulesFromRAG() error = %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected output file to be created")
	}

	// Read and verify content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "rules:") {
		t.Error("Output should contain 'rules:' YAML key")
	}
	if !strings.Contains(content, "test-pattern") {
		t.Error("Output should contain the test pattern")
	}
}

func TestGenerateRulesFromRAG_CategoryNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	ragDir := filepath.Join(tmpDir, "rag")
	outputPath := filepath.Join(tmpDir, "output.yaml")

	err := GenerateRulesFromRAG(ragDir, "nonexistent/category", outputPath)
	if err == nil {
		t.Error("Expected error for nonexistent category")
	}
}

func TestSanitizeRuleID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Name", "simple-name"},
		{"Using :latest Tag", "using-latest-tag"},
		{"N+1 Query Pattern", "n-1-query-pattern"},
		{"CamelCase", "camelcase"},
		{"with--multiple---dashes", "with-multiple-dashes"},
	}

	for _, tt := range tests {
		got := sanitizeRuleID(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeRuleID(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMapToSemgrepLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"python", "python"},
		{"Python", "python"},
		{"javascript", "javascript"},
		{"typescript", "typescript"},
		{"go", "go"},
		{"dockerfile", "dockerfile"},
		{"Dockerfile", "dockerfile"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		got := mapToSemgrepLanguage(tt.input)
		if got != tt.expected {
			t.Errorf("mapToSemgrepLanguage(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMapSeverityToSemgrep(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "ERROR"},
		{"high", "ERROR"},
		{"medium", "WARNING"},
		{"low", "INFO"},
		{"info", "INFO"},
		{"unknown", "INFO"},
		{"", "INFO"},
	}

	for _, tt := range tests {
		got := mapSeverityToSemgrep(tt.input)
		if got != tt.expected {
			t.Errorf("mapSeverityToSemgrep(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMapSemgrepSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ERROR", "high"},
		{"WARNING", "medium"},
		{"INFO", "low"},
		{"error", "high"},
		{"warning", "medium"},
		{"info", "low"},
		{"unknown", "info"},
	}

	for _, tt := range tests {
		got := mapSemgrepSeverity(tt.input)
		if got != tt.expected {
			t.Errorf("mapSemgrepSeverity(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestWriteRulesYAML(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "rules.yaml")

	rules := &SemgrepRuleFile{
		Rules: []SemgrepRule{
			{
				ID:           "test.rule.1",
				Message:      "Test message",
				Severity:     "ERROR",
				Languages:    []string{"generic"},
				PatternRegex: "test.*pattern",
				Metadata: map[string]interface{}{
					"category": "test",
				},
			},
		},
	}

	err := WriteRulesYAML(outputPath, rules)
	if err != nil {
		t.Fatalf("WriteRulesYAML() error = %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "test.rule.1") {
		t.Error("Output should contain rule ID")
	}
	if !strings.Contains(content, "Test message") {
		t.Error("Output should contain message")
	}
	if !strings.Contains(content, "pattern-regex") {
		t.Error("Output should contain pattern-regex")
	}
}

func TestHasSemgrep(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The actual result depends on whether semgrep is installed
	result := HasSemgrep()
	t.Logf("HasSemgrep() = %v", result)
}

func TestNewSemgrepRunner(t *testing.T) {
	cfg := SemgrepConfig{
		RulePaths: []string{"/path/to/rules.yaml"},
	}

	runner := NewSemgrepRunner(cfg)
	if runner == nil {
		t.Fatal("Expected non-nil runner")
	}

	// Verify defaults
	if runner.timeout == 0 {
		t.Error("Expected non-zero default timeout")
	}
	if runner.onStatus == nil {
		t.Error("Expected non-nil onStatus callback")
	}
}

func TestSemgrepRunner_NoRules(t *testing.T) {
	runner := NewSemgrepRunner(SemgrepConfig{
		RulePaths: []string{},
	})

	result := runner.Run(nil, "/tmp")
	if result.Error == nil {
		t.Error("Expected error for no rule files")
	}
}

// Integration test - requires actual RAG files
func TestParseRealDockerPatterns(t *testing.T) {
	// Find project root from test file location
	patternFile := findProjectFile(t, "rag/devops/docker/patterns.md")
	if patternFile == "" {
		t.Skip("Skipping: RAG docker patterns file not found")
	}

	parsed, err := ParsePatternMarkdown(patternFile)
	if err != nil {
		t.Fatalf("ParsePatternMarkdown() error = %v", err)
	}

	if parsed.Category != "devops/docker" {
		t.Errorf("Category = %q, want devops/docker", parsed.Category)
	}

	if len(parsed.Patterns) < 10 {
		t.Errorf("Expected at least 10 docker patterns, got %d", len(parsed.Patterns))
	}

	// Verify some known patterns exist
	patternNames := make(map[string]bool)
	for _, p := range parsed.Patterns {
		patternNames[p.Name] = true
	}

	expectedPatterns := []string{
		"Using :latest Tag",
		"Running as Root",
		"Hardcoded Secret in Dockerfile",
	}
	for _, name := range expectedPatterns {
		if !patternNames[name] {
			t.Errorf("Expected to find pattern %q", name)
		}
	}
}

// Integration test - requires actual RAG files
func TestParseRealAPIQualityPatterns(t *testing.T) {
	patternFile := findProjectFile(t, "rag/code-quality/api/patterns.md")
	if patternFile == "" {
		t.Skip("Skipping: RAG API quality patterns file not found")
	}

	parsed, err := ParsePatternMarkdown(patternFile)
	if err != nil {
		t.Fatalf("ParsePatternMarkdown() error = %v", err)
	}

	if parsed.Category != "code-quality/api" {
		t.Errorf("Category = %q, want code-quality/api", parsed.Category)
	}

	if len(parsed.Patterns) < 15 {
		t.Errorf("Expected at least 15 API patterns, got %d", len(parsed.Patterns))
	}
}

// findProjectFile looks for a file relative to the project root
// It walks up from the current directory looking for go.mod to find the project root
func findProjectFile(t *testing.T, relPath string) string {
	t.Helper()

	// Try to find project root by looking for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found project root
			fullPath := filepath.Join(dir, relPath)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
			return ""
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			return ""
		}
		dir = parent
	}
}
