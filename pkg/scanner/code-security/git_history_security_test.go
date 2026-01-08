package codesecurity

import (
	"regexp"
	"testing"
)

func TestNewGitHistorySecurityScanner(t *testing.T) {
	config := GitHistorySecurityConfig{
		Enabled:              true,
		MaxCommits:           500,
		MaxAge:               "6m",
		ScanGitignoreHistory: true,
		ScanSensitiveFiles:   true,
		GeneratePurgeReport:  true,
	}

	scanner := NewGitHistorySecurityScanner(config)
	if scanner == nil {
		t.Fatal("NewGitHistorySecurityScanner() returned nil")
	}

	// Patterns are loaded from RAG or fallback
	t.Logf("Loaded %d sensitive patterns", len(scanner.sensitivePatterns))
}

func TestGitHistorySecurityScanner_SensitivePatterns(t *testing.T) {
	// Clear cache to ensure fresh pattern loading
	ClearRAGGitHistoryCache()
	scanner := NewGitHistorySecurityScanner(GitHistorySecurityConfig{})

	if len(scanner.sensitivePatterns) == 0 {
		t.Fatal("No sensitive patterns loaded")
	}

	// Test that we have patterns for common sensitive files
	// Note: Paths should be realistic - the patterns expect file paths like they appear in git
	testCases := []struct {
		path        string
		shouldMatch bool
		category    string
	}{
		{"project/.env", true, "credentials"},
		{"config/.env.local", true, "credentials"},
		{"app/.env.production", true, "credentials"},
		{"config/secrets.json", true, "credentials"},
		{"app/credentials.json", true, "credentials"},
		{".ssh/id_rsa", true, "keys"},
		{"certs/server.pem", true, "keys"},
		{"infra/terraform.tfstate", true, "infrastructure"},
		{"data/app.db", true, "database"},
		{"src/main.go", false, ""},
		{"docs/README.md", false, ""},
	}

	for _, tc := range testCases {
		matched := scanner.matchesSensitivePattern(tc.path)
		if tc.shouldMatch && matched == nil {
			t.Errorf("Expected %q to match a sensitive pattern", tc.path)
		} else if !tc.shouldMatch && matched != nil {
			t.Errorf("Expected %q to NOT match a sensitive pattern, but matched %s", tc.path, matched.Category)
		} else if tc.shouldMatch && matched != nil && tc.category != "" && matched.Category != tc.category {
			t.Errorf("Expected %q to match category %q, but got %q", tc.path, tc.category, matched.Category)
		}
	}
}

func TestGitHistorySecurityScanner_GitignorePatterns(t *testing.T) {
	scanner := NewGitHistorySecurityScanner(GitHistorySecurityConfig{})

	// Test gitignore pattern parsing
	testGitignore := `
# Comments should be ignored
*.log
.env
node_modules/
!important.log
`

	// We can't directly test parseGitignore without a file,
	// but we can test the pattern matching logic
	scanner.gitignoreRules = []gitignoreRule{
		{Pattern: "*.log", Regex: mustCompileGitignore("*.log"), IsNegation: false},
		{Pattern: ".env", Regex: mustCompileGitignore(".env"), IsNegation: false},
		{Pattern: "node_modules/", Regex: mustCompileGitignore("node_modules"), IsDir: true, IsNegation: false},
		{Pattern: "!important.log", Regex: mustCompileGitignore("important.log"), IsNegation: true},
	}

	testCases := []struct {
		path        string
		shouldMatch bool
	}{
		{"debug.log", true},
		{"error.log", true},
		{".env", true},
		{"node_modules/package.json", true},
		{"important.log", false}, // Negated
		{"src/main.go", false},
	}

	for _, tc := range testCases {
		matched, _ := scanner.matchesGitignore(tc.path)
		if matched != tc.shouldMatch {
			t.Errorf("matchesGitignore(%q) = %v, want %v", tc.path, matched, tc.shouldMatch)
		}
	}

	_ = testGitignore // Silence unused variable warning
}

func TestGitHistorySecurityScanner_parseSinceDate(t *testing.T) {
	tests := []struct {
		maxAge      string
		description string
	}{
		{"30d", "30 days"},
		{"90d", "90 days"},
		{"6m", "6 months"},
		{"1y", "1 year"},
		{"2y", "2 years"},
		{"", "default 1 year"},
	}

	for _, tt := range tests {
		t.Run(tt.maxAge, func(t *testing.T) {
			scanner := NewGitHistorySecurityScanner(GitHistorySecurityConfig{MaxAge: tt.maxAge})
			since := scanner.parseSinceDate()

			if since.IsZero() {
				t.Errorf("parseSinceDate() returned zero time for maxAge=%q", tt.maxAge)
			}

			t.Logf("parseSinceDate() with maxAge=%q (%s): %v", tt.maxAge, tt.description, since)
		})
	}
}

func TestGenerateBFGCommand(t *testing.T) {
	tests := []struct {
		file     string
		expected string
	}{
		{".env", "bfg --delete-files '.env'"},
		{"secrets.json", "bfg --delete-files 'secrets.json'"},
		{"path/with spaces/file.txt", "bfg --delete-files 'path/with spaces/file.txt'"},
		{"file'with'quotes.txt", "bfg --delete-files 'file'\\''with'\\''quotes.txt'"},
	}

	for _, tt := range tests {
		got := generateBFGCommand(tt.file)
		if got != tt.expected {
			t.Errorf("generateBFGCommand(%q) = %q, want %q", tt.file, got, tt.expected)
		}
	}
}

func TestGenerateFilterRepoCommand(t *testing.T) {
	tests := []struct {
		file     string
		expected string
	}{
		{".env", "git filter-repo --path '.env' --invert-paths"},
		{"secrets.json", "git filter-repo --path 'secrets.json' --invert-paths"},
	}

	for _, tt := range tests {
		got := generateFilterRepoCommand(tt.file)
		if got != tt.expected {
			t.Errorf("generateFilterRepoCommand(%q) = %q, want %q", tt.file, got, tt.expected)
		}
	}
}

func TestGitHistorySecurityResult_Fields(t *testing.T) {
	result := &GitHistorySecurityResult{
		GitignoreViolations: []GitignoreViolation{
			{File: ".env", GitignoreRule: ".env"},
		},
		SensitiveFiles: []SensitiveFileFinding{
			{File: "credentials.json", Category: "credentials", Severity: "critical"},
		},
		PurgeRecommendations: []PurgeRecommendation{
			{File: ".env", Reason: "Environment file", Severity: "critical", Priority: 1},
		},
		Timeline: []HistoricalEvent{
			{Date: "2024-01-01", EventType: "committed", File: ".env"},
		},
		Summary: GitHistorySecuritySummary{
			TotalViolations:     2,
			GitignoreViolations: 1,
			SensitiveFilesFound: 1,
			FilesToPurge:        1,
			CommitsScanned:      100,
			ByCategory:          map[string]int{"credentials": 1},
			BySeverity:          map[string]int{"critical": 1},
			RiskScore:           75,
			RiskLevel:           "medium",
		},
	}

	if len(result.GitignoreViolations) != 1 {
		t.Errorf("GitignoreViolations count = %d, want 1", len(result.GitignoreViolations))
	}
	if len(result.SensitiveFiles) != 1 {
		t.Errorf("SensitiveFiles count = %d, want 1", len(result.SensitiveFiles))
	}
	if len(result.PurgeRecommendations) != 1 {
		t.Errorf("PurgeRecommendations count = %d, want 1", len(result.PurgeRecommendations))
	}
	if result.Summary.TotalViolations != 2 {
		t.Errorf("TotalViolations = %d, want 2", result.Summary.TotalViolations)
	}
}

// mustCompileGitignore is a helper for tests
func mustCompileGitignore(pattern string) *regexp.Regexp {
	regex := gitignoreToRegex(pattern)
	compiled, err := regexp.Compile("(?i)" + regex)
	if err != nil {
		panic(err)
	}
	return compiled
}
