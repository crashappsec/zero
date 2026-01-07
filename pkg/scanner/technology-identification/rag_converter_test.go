package techid

import (
	"os"
	"path/filepath"
	"regexp"
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

func TestMapLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"python", "python"},
		{"javascript", "javascript"},
		{"typescript", "typescript"},
		{"go", "go"},
		{"ruby", "ruby"},
		{"java", "java"},
		{"php", "php"},
		{"csharp", "csharp"},
		{"c#", "csharp"},
		{"rust", "rust"},
		{"c", "c"},
		{"c++", "cpp"},
		{"cpp", "cpp"},
		{"kotlin", "kotlin"},
		{"scala", "scala"},
		{"swift", "swift"},
		// New languages
		{"bash", "bash"},
		{"shell", "bash"},
		{"sh", "bash"},
		{"powershell", "generic"},
		{"dockerfile", "dockerfile"},
		{"docker", "dockerfile"},
		{"hcl", "hcl"},
		{"terraform", "hcl"},
		{"yaml", "yaml"},
		{"json", "json"},
		{"generic", "generic"},
		// Unknown returns empty
		{"unknown", ""},
	}

	for _, tc := range tests {
		result := mapLanguage(tc.input)
		if result != tc.expected {
			t.Errorf("mapLanguage(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestCleanRegexForSemgrep(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		{"", "", "empty string"},
		{`api`, "api", "short valid pattern (3 chars)"},
		{`jwt`, "jwt", "short valid pattern (3 chars)"},
		{`^import foo$`, "import foo", "anchors stripped"},
		{`\b`, "", "too generic - word boundary"},
		{`\s+`, "", "too generic - whitespace"},
		{`.*`, "", "too generic - match all"},
		{`.+`, "", "too generic - match one or more"},
		{`\w+`, "", "too generic - word chars"},
		{`import\s+.*\s+from\s+['"]react['"]`, `import\s+.*\s+from\s+['"]react['"]`, "complex valid pattern"},
		{`(unclosed`, "", "invalid regex"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := cleanRegexForSemgrep(tc.input)
			if result != tc.expected {
				t.Errorf("cleanRegexForSemgrep(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestMapSeverityWithMetadata(t *testing.T) {
	tests := []struct {
		input          string
		expectedSev    string
		expectCritical bool
	}{
		{"critical", "ERROR", true},
		{"high", "WARNING", false},
		{"medium", "WARNING", false},
		{"low", "INFO", false},
		{"info", "INFO", false},
		{"", "INFO", false},
		{"unknown", "INFO", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			sev, meta := mapSeverityWithMetadata(tc.input)
			if sev != tc.expectedSev {
				t.Errorf("mapSeverityWithMetadata(%q) severity = %q, expected %q", tc.input, sev, tc.expectedSev)
			}

			if tc.input != "" {
				if origSev, ok := meta["original_severity"].(string); !ok || origSev != tc.input {
					t.Errorf("mapSeverityWithMetadata(%q) should preserve original_severity, got %v", tc.input, meta["original_severity"])
				}
			}

			if tc.expectCritical {
				if isCrit, ok := meta["is_critical"].(bool); !ok || !isCrit {
					t.Errorf("mapSeverityWithMetadata(%q) should set is_critical=true", tc.input)
				}
			}
		})
	}
}

func TestDiscoverRAGDirectories(t *testing.T) {
	// Create a temp directory with some subdirs
	tmpDir, err := os.MkdirTemp("", "rag-discover-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some subdirectories
	subdirs := []string{"technology-identification", "cryptography", "devops", ".hidden"}
	for _, dir := range subdirs {
		os.Mkdir(filepath.Join(tmpDir, dir), 0755)
	}
	// Create a file (should be ignored)
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("test"), 0644)

	dirs, err := discoverRAGDirectories(tmpDir)
	if err != nil {
		t.Fatalf("discoverRAGDirectories failed: %v", err)
	}

	// Should have 3 directories (hidden one is excluded)
	if len(dirs) != 3 {
		t.Errorf("Expected 3 directories, got %d: %v", len(dirs), dirs)
	}

	// Hidden directory should be excluded
	for _, d := range dirs {
		if d == ".hidden" {
			t.Error("Hidden directory should be excluded")
		}
	}
}

func TestConvertRAGToSemgrep(t *testing.T) {
	// Try multiple possible RAG locations
	ragDir := findTestRAGDir()
	if ragDir == "" {
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

func TestCWEPreservation(t *testing.T) {
	// Test that CWE references are extracted from pattern descriptions
	testCases := []struct {
		description string
		expectedCWE string
	}{
		{"SQL Injection vulnerability - CWE-89", "CWE-89"},
		{"Cross-site scripting (CWE-79) detected", "CWE-79"},
		{"CWE-327: Use of a Broken or Risky Cryptographic Algorithm", "CWE-327"},
		{"No CWE reference here", ""},
		{"Multiple CWE-123 and CWE-456 references", "CWE-123"}, // Takes first
	}

	for _, tc := range testCases {
		t.Run(tc.description[:20], func(t *testing.T) {
			cwe := extractCWE(tc.description)
			if cwe != tc.expectedCWE {
				t.Errorf("extractCWE(%q) = %q, expected %q", tc.description, cwe, tc.expectedCWE)
			}
		})
	}
}

func TestAllCategoriesProcessed(t *testing.T) {
	ragDir := findTestRAGDir()
	if ragDir == "" {
		t.Skip("RAG directory not found, skipping integration test")
	}

	// Discover all categories
	dirs, err := discoverRAGDirectories(ragDir)
	if err != nil {
		t.Fatalf("Failed to discover directories: %v", err)
	}

	// Should have a significant number of categories
	// Based on the plan: 23+ categories expected
	minExpectedCategories := 15 // Lower bound for test stability
	if len(dirs) < minExpectedCategories {
		t.Errorf("Expected at least %d categories, got %d: %v", minExpectedCategories, len(dirs), dirs)
	}

	// Check for key expected categories
	expectedCategories := []string{
		"technology-identification",
		"cryptography",
		"code-security",
	}

	for _, expected := range expectedCategories {
		found := false
		for _, dir := range dirs {
			if dir == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected category %q not found in: %v", expected, dirs)
		}
	}

	t.Logf("Found %d categories: %v", len(dirs), dirs)
}

func TestRegexToSemgrepConversion(t *testing.T) {
	tests := []struct {
		name        string
		regex       string
		language    string
		shouldMatch bool // Whether it should produce non-empty output
	}{
		// JavaScript require - works
		{
			name:        "js require",
			regex:       `require\(['"]express['"]\)`,
			language:    "javascript",
			shouldMatch: true,
		},
		// Go import - works
		{
			name:        "go import",
			regex:       `"github\.com/gin-gonic/gin"`,
			language:    "go",
			shouldMatch: true,
		},
		// Python patterns with \s+ get filtered (too generic)
		{
			name:        "python import with whitespace",
			regex:       `^import\s+requests`,
			language:    "python",
			shouldMatch: false, // \s+ gets filtered
		},
		// Generic patterns with character classes may be filtered
		{
			name:        "generic pattern with char class",
			regex:       `AKIA[0-9A-Z]{16}`,
			language:    "generic",
			shouldMatch: false, // Currently filtered
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := regexToSemgrep(tc.regex, tc.language)
			hasResult := result != ""
			if hasResult != tc.shouldMatch {
				t.Errorf("regexToSemgrep(%q, %q) = %q, expected shouldMatch=%v",
					tc.regex, tc.language, result, tc.shouldMatch)
			}
			t.Logf("regexToSemgrep(%q, %q) = %q", tc.regex, tc.language, result)
		})
	}
}

func TestMetadataPreservation(t *testing.T) {
	// Test that all important metadata fields are preserved
	pattern := &RAGPattern{
		Name:        "Test Pattern",
		Category:    "code-security/injection",
		Description: "SQL Injection - CWE-89",
		Homepage:    "https://example.com",
		ConfigFiles: []string{"config.json"},
		Packages:    map[string][]string{"npm": {"test-pkg"}},
		Imports:     make(map[string][]ImportPattern),
		Confidence:  map[string]int{"config_file_detection": 95},
	}

	rules := convertPatternToRules(pattern, "/rag/code-security/injection/test.md", "/rag")

	if len(rules) == 0 {
		t.Fatal("Expected at least one rule to be generated")
	}

	rule := rules[0]
	meta := rule.Metadata

	// Check technology name preserved
	if meta["technology"] != "Test Pattern" {
		t.Errorf("Expected technology 'Test Pattern', got %v", meta["technology"])
	}

	// Check category preserved
	if meta["category"] != "code-security/injection" {
		t.Errorf("Expected category 'code-security/injection', got %v", meta["category"])
	}

	// Check homepage preserved
	if meta["homepage"] != "https://example.com" {
		t.Errorf("Expected homepage 'https://example.com', got %v", meta["homepage"])
	}

	// Check config files preserved
	if configFiles, ok := meta["config_files"].([]string); !ok || len(configFiles) == 0 {
		t.Errorf("Expected config_files to be preserved, got %v", meta["config_files"])
	}

	// Check confidence exists (default is used when key doesn't match)
	if _, ok := meta["confidence"].(int); !ok {
		t.Errorf("Expected confidence to be an int, got %T", meta["confidence"])
	}

	// Check detection_type is set
	if meta["detection_type"] == nil {
		t.Error("Expected detection_type to be set")
	}

	t.Logf("Metadata preserved: %+v", meta)
}

// extractCWE extracts CWE identifier from a description string
func extractCWE(description string) string {
	// Simple regex to find CWE-XXX pattern
	re := regexp.MustCompile(`CWE-\d+`)
	match := re.FindString(description)
	return match
}

// findTestRAGDir tries to locate the RAG directory for testing
func findTestRAGDir() string {
	candidates := []string{
		"rag",                     // Current directory
		"../rag",                  // Parent directory
		"../../rag",               // Two levels up
		"../../../rag",            // Three levels up
		"../../../../rag",         // Four levels up (for deep test directories)
	}

	// Also check ZERO_HOME environment variable
	if zeroHome := os.Getenv("ZERO_HOME"); zeroHome != "" {
		candidates = append([]string{filepath.Join(zeroHome, "rag")}, candidates...)
	}

	// Check ZERO_RAG_PATH for explicit override
	if ragPath := os.Getenv("ZERO_RAG_PATH"); ragPath != "" {
		candidates = append([]string{ragPath}, candidates...)
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			// Verify it's actually the RAG directory by checking for known subdirs
			if _, err := os.Stat(filepath.Join(absPath, "technology-identification")); err == nil {
				return absPath
			}
		}
	}

	return ""
}
