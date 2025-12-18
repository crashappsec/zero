package techid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePatternFile(t *testing.T) {
	// Create a temp directory with test pattern file
	tmpDir, err := os.MkdirTemp("", "rag-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	patternContent := `# ESLint

**Category**: developer-tools/linting
**Description**: ESLint - pluggable JavaScript and TypeScript linter
**Homepage**: https://eslint.org

## Package Detection

### NPM
- ` + "`eslint`" + `
- ` + "`@eslint/js`" + `

## Configuration Files

- ` + "`.eslintrc`" + `
- ` + "`.eslintrc.js`" + `
- ` + "`.eslintrc.json`" + `
- ` + "`eslint.config.js`" + ` (flat config)

## Environment Variables

- ` + "`ESLINT_USE_FLAT_CONFIG`" + `

## Detection Confidence

- **Configuration File Detection**: 95% (HIGH)
- **Package Detection**: 95% (HIGH)
`

	patternFile := filepath.Join(tmpDir, "patterns.md")
	if err := os.WriteFile(patternFile, []byte(patternContent), 0644); err != nil {
		t.Fatalf("Failed to write pattern file: %v", err)
	}

	pattern, err := parsePatternFile(patternFile)
	if err != nil {
		t.Fatalf("Failed to parse pattern file: %v", err)
	}

	// Verify name
	if pattern.Name != "ESLint" {
		t.Errorf("Expected name 'ESLint', got '%s'", pattern.Name)
	}

	// Verify category
	if pattern.Category != "developer-tools/linting" {
		t.Errorf("Expected category 'developer-tools/linting', got '%s'", pattern.Category)
	}

	// Verify homepage
	if pattern.Homepage != "https://eslint.org" {
		t.Errorf("Expected homepage 'https://eslint.org', got '%s'", pattern.Homepage)
	}

	// Verify config files
	expectedConfigs := []string{".eslintrc", ".eslintrc.js", ".eslintrc.json", "eslint.config.js"}
	if len(pattern.ConfigFiles) != len(expectedConfigs) {
		t.Errorf("Expected %d config files, got %d: %v", len(expectedConfigs), len(pattern.ConfigFiles), pattern.ConfigFiles)
	} else {
		for i, expected := range expectedConfigs {
			if pattern.ConfigFiles[i] != expected {
				t.Errorf("Config file %d: expected '%s', got '%s'", i, expected, pattern.ConfigFiles[i])
			}
		}
	}

	// Verify packages
	npmPkgs := pattern.Packages["npm"]
	if len(npmPkgs) != 2 {
		t.Errorf("Expected 2 npm packages, got %d: %v", len(npmPkgs), npmPkgs)
	}

	// Verify env vars
	if len(pattern.EnvVars) != 1 || pattern.EnvVars[0] != "ESLINT_USE_FLAT_CONFIG" {
		t.Errorf("Expected env var ESLINT_USE_FLAT_CONFIG, got %v", pattern.EnvVars)
	}

	t.Logf("Parsed pattern: %+v", pattern)
}

func TestConvertPatternToRules_ConfigFiles(t *testing.T) {
	pattern := &RAGPattern{
		Name:        "Webpack",
		Category:    "developer-tools/bundlers",
		Description: "Webpack module bundler",
		Homepage:    "https://webpack.js.org",
		ConfigFiles: []string{"webpack.config.js", "webpack.config.ts", "webpack.dev.js"},
		Packages:    make(map[string][]string),
		Imports:     make(map[string][]ImportPattern),
		Confidence:  make(map[string]int),
	}

	rules := convertPatternToRules(pattern, "/rag/technology-identification/developer-tools/bundlers/webpack/patterns.md", "/rag")

	// Should have at least one config rule
	var configRule *SemgrepRule
	for i, rule := range rules {
		if rule.ID == "zero.developer-tools.bundlers.webpack.config" {
			configRule = &rules[i]
			break
		}
	}

	if configRule == nil {
		t.Fatalf("Expected config rule, got rules: %v", rules)
	}

	// Verify rule properties
	if configRule.Severity != "INFO" {
		t.Errorf("Expected severity INFO, got %s", configRule.Severity)
	}

	if configRule.Languages[0] != "generic" {
		t.Errorf("Expected language 'generic', got %v", configRule.Languages)
	}

	// Verify metadata
	meta := configRule.Metadata
	if meta["technology"] != "Webpack" {
		t.Errorf("Expected technology 'Webpack', got %v", meta["technology"])
	}

	if meta["tool_type"] != "bundler" {
		t.Errorf("Expected tool_type 'bundler', got %v", meta["tool_type"])
	}

	if meta["homepage"] != "https://webpack.js.org" {
		t.Errorf("Expected homepage, got %v", meta["homepage"])
	}

	configFiles, ok := meta["config_files"].([]string)
	if !ok || len(configFiles) != 3 {
		t.Errorf("Expected 3 config files in metadata, got %v", meta["config_files"])
	}

	// Verify pattern regex
	expectedPattern := `(webpack\.config\.js|webpack\.config\.ts|webpack\.dev\.js)`
	if configRule.PatternRegex != expectedPattern {
		t.Errorf("Expected pattern regex '%s', got '%s'", expectedPattern, configRule.PatternRegex)
	}

	t.Logf("Config rule: %+v", configRule)
}

func TestCategoryToToolType(t *testing.T) {
	tests := []struct {
		category string
		expected string
	}{
		{"developer-tools/linting", "linter"},
		{"developer-tools/bundlers", "bundler"},
		{"developer-tools/testing", "test"},
		{"developer-tools/cicd", "ci-cd"},
		{"developer-tools/build", "build"},
		{"developer-tools/containers", "container"},
		{"languages", "language"},
		{"web-frameworks/backend", "framework"},
		{"databases/postgresql", "database"},
		{"ai-ml/apis", "ai-ml"},
		{"unknown/category", "category"},
	}

	for _, tc := range tests {
		result := categoryToToolType(tc.category)
		if result != tc.expected {
			t.Errorf("categoryToToolType(%q) = %q, expected %q", tc.category, result, tc.expected)
		}
	}
}

func TestConvertRAGToSemgrep(t *testing.T) {
	// Test with actual RAG directory if it exists
	ragDir := "/Users/curphey/zero/rag"
	if _, err := os.Stat(ragDir); os.IsNotExist(err) {
		t.Skip("RAG directory not found, skipping integration test")
	}

	// Create temp output directory
	outputDir, err := os.MkdirTemp("", "rag-output")
	if err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}
	defer os.RemoveAll(outputDir)

	result, err := ConvertRAGToSemgrep(ragDir, outputDir)
	if err != nil {
		t.Fatalf("ConvertRAGToSemgrep failed: %v", err)
	}

	t.Logf("Conversion result:")
	t.Logf("  Tech Discovery rules: %d", len(result.TechDiscovery.Rules))
	t.Logf("  Secret rules: %d", len(result.Secrets.Rules))
	t.Logf("  AI/ML rules: %d", len(result.AIML.Rules))
	t.Logf("  Config Files rules: %d", len(result.ConfigFiles.Rules))
	t.Logf("  Total rules: %d", result.TotalRules)

	// Verify config file rules were generated
	if len(result.ConfigFiles.Rules) == 0 {
		t.Error("Expected some config file rules to be generated")
	}

	// Check for specific config rules
	eslintFound := false
	webpackFound := false
	for _, rule := range result.ConfigFiles.Rules {
		if rule.Metadata["technology"] == "ESLint" {
			eslintFound = true
			t.Logf("Found ESLint config rule: %s", rule.ID)
		}
		if rule.Metadata["technology"] == "Webpack" {
			webpackFound = true
			t.Logf("Found Webpack config rule: %s", rule.ID)
		}
	}

	if !eslintFound {
		t.Error("Expected ESLint config rule to be generated")
	}
	if !webpackFound {
		t.Error("Expected Webpack config rule to be generated")
	}

	// Verify output files were created
	expectedFiles := []string{"tech-discovery.yaml", "config-files.yaml"}
	for _, f := range expectedFiles {
		path := filepath.Join(outputDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected output file %s to be created", f)
		} else {
			info, _ := os.Stat(path)
			t.Logf("Output file %s: %d bytes", f, info.Size())
		}
	}
}
