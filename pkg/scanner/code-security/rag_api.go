// Package codesecurity provides the consolidated code security super scanner
package codesecurity

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/crashappsec/zero/pkg/scanner"
)

// APIPattern represents a parsed RAG pattern for API security
type APIPattern struct {
	Name       string   // Pattern name from section header
	Category   string   // api-auth, api-injection, etc.
	Severity   string   // critical, high, medium, low
	Confidence int      // 0-100
	CWE        string   // CWE-89, CWE-78, etc.
	OWASPApi   string   // API1:2023, API2:2023, etc.
	Pattern    string   // regex pattern
	Languages  []string // javascript, typescript, python, etc.
	compiled   *regexp.Regexp
}

// APIPatternLoader loads and manages RAG patterns for API security
type APIPatternLoader struct {
	patterns []APIPattern
	mu       sync.RWMutex
}

// NewAPIPatternLoader creates a new pattern loader
func NewAPIPatternLoader() *APIPatternLoader {
	return &APIPatternLoader{
		patterns: make([]APIPattern, 0),
	}
}

// LoadPatterns loads all API security patterns from the RAG directory
func (l *APIPatternLoader) LoadPatterns(ragDir string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	apiSecurityDir := filepath.Join(ragDir, "api-security")
	entries, err := os.ReadDir(apiSecurityDir)
	if err != nil {
		return fmt.Errorf("reading api-security dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(apiSecurityDir, entry.Name())
		patterns, err := l.parsePatternFile(filePath)
		if err != nil {
			continue // Skip files that fail to parse
		}
		l.patterns = append(l.patterns, patterns...)
	}

	return nil
}

// parsePatternFile parses a single RAG pattern markdown file
func (l *APIPatternLoader) parsePatternFile(filePath string) ([]APIPattern, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []APIPattern
	var current *APIPattern
	var inCodeBlock bool
	var codeBlockContent strings.Builder

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Section header starts a new pattern
		if strings.HasPrefix(line, "### ") {
			// Save previous pattern if exists
			if current != nil && current.Pattern != "" {
				if compiled, err := regexp.Compile(current.Pattern); err == nil {
					current.compiled = compiled
					patterns = append(patterns, *current)
				}
			}
			current = &APIPattern{
				Name: strings.TrimPrefix(line, "### "),
			}
			continue
		}

		if current == nil {
			continue
		}

		// Parse metadata lines
		if strings.HasPrefix(line, "CATEGORY:") {
			current.Category = strings.TrimSpace(strings.TrimPrefix(line, "CATEGORY:"))
		} else if strings.HasPrefix(line, "SEVERITY:") {
			current.Severity = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "SEVERITY:")))
		} else if strings.HasPrefix(line, "CONFIDENCE:") {
			confStr := strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
			if conf, err := strconv.Atoi(confStr); err == nil {
				current.Confidence = conf
			}
		} else if strings.HasPrefix(line, "CWE:") {
			current.CWE = strings.TrimSpace(strings.TrimPrefix(line, "CWE:"))
		} else if strings.HasPrefix(line, "OWASP:") {
			current.OWASPApi = strings.TrimSpace(strings.TrimPrefix(line, "OWASP:"))
		}

		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// End of code block - parse content
				content := codeBlockContent.String()
				parseCodeBlockContent(content, current)
				codeBlockContent.Reset()
				inCodeBlock = false
			} else {
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			codeBlockContent.WriteString(line)
			codeBlockContent.WriteString("\n")
		}
	}

	// Save last pattern
	if current != nil && current.Pattern != "" {
		if compiled, err := regexp.Compile(current.Pattern); err == nil {
			current.compiled = compiled
			patterns = append(patterns, *current)
		}
	}

	return patterns, scanner.Err()
}

// parseCodeBlockContent extracts PATTERN and LANGUAGES from code block
func parseCodeBlockContent(content string, pattern *APIPattern) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PATTERN:") {
			pattern.Pattern = strings.TrimSpace(strings.TrimPrefix(line, "PATTERN:"))
		} else if strings.HasPrefix(line, "LANGUAGES:") {
			langStr := strings.TrimSpace(strings.TrimPrefix(line, "LANGUAGES:"))
			pattern.Languages = strings.Split(langStr, ",")
			for i := range pattern.Languages {
				pattern.Languages[i] = strings.TrimSpace(pattern.Languages[i])
			}
		}
	}
}

// GetPatterns returns all loaded patterns
func (l *APIPatternLoader) GetPatterns() []APIPattern {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.patterns
}

// GetPatternsByCategory returns patterns filtered by category
func (l *APIPatternLoader) GetPatternsByCategory(category string) []APIPattern {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var filtered []APIPattern
	for _, p := range l.patterns {
		if p.Category == category {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// runRAGAPIPatterns applies RAG patterns to the repository
func (s *CodeSecurityScanner) runRAGAPIPatterns(ctx context.Context, opts *scanner.ScanOptions, cfg APIConfig) []APIFinding {
	var findings []APIFinding

	// Load patterns
	loader := NewAPIPatternLoader()
	ragDir := filepath.Join(filepath.Dir(opts.RepoPath), "..", "..", "rag")
	if err := loader.LoadPatterns(ragDir); err != nil {
		// Try alternate location
		ragDir = "rag"
		if err := loader.LoadPatterns(ragDir); err != nil {
			return findings // No patterns available
		}
	}

	patterns := loader.GetPatterns()
	if len(patterns) == 0 {
		return findings
	}

	// Filter patterns based on config
	var activePatterns []APIPattern
	for _, p := range patterns {
		if shouldRunPattern(p, cfg) {
			activePatterns = append(activePatterns, p)
		}
	}

	// Scan files
	filePatterns := []string{"*.js", "*.ts", "*.py", "*.go", "*.java", "*.rb", "*.php"}
	for _, fp := range filePatterns {
		matches, err := filepath.Glob(filepath.Join(opts.RepoPath, "**", fp))
		if err != nil {
			continue
		}

		// Also check root level
		rootMatches, _ := filepath.Glob(filepath.Join(opts.RepoPath, fp))
		matches = append(matches, rootMatches...)

		for _, filePath := range matches {
			if shouldSkipFile(filePath) {
				continue
			}

			fileFindings := s.scanFileWithPatterns(ctx, filePath, opts.RepoPath, activePatterns)
			findings = append(findings, fileFindings...)
		}
	}

	return findings
}

// shouldRunPattern checks if a pattern should be run based on config
func shouldRunPattern(p APIPattern, cfg APIConfig) bool {
	switch p.Category {
	case "api-auth":
		return cfg.CheckAuth
	case "api-injection":
		return cfg.CheckInjection
	case "api-ssrf":
		return cfg.CheckSSRF
	case "api-mass-assignment", "api-data-exposure":
		return cfg.CheckAuth // These relate to auth
	case "api-rate-limiting":
		return true // Always check
	default:
		return true
	}
}

// shouldSkipFile checks if a file should be skipped
func shouldSkipFile(path string) bool {
	skipPatterns := []string{
		"node_modules", "vendor", ".git", "test", "tests",
		"_test.go", ".test.js", ".spec.ts", ".test.ts",
		"__tests__", "__mocks__", "fixtures",
	}
	for _, pattern := range skipPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// scanFileWithPatterns scans a single file against all patterns
func (s *CodeSecurityScanner) scanFileWithPatterns(ctx context.Context, filePath, repoPath string, patterns []APIPattern) []APIFinding {
	var findings []APIFinding

	content, err := os.ReadFile(filePath)
	if err != nil {
		return findings
	}

	lines := strings.Split(string(content), "\n")
	relPath := strings.TrimPrefix(filePath, repoPath)
	relPath = strings.TrimPrefix(relPath, "/")

	lang := detectLanguage(filePath)

	for _, pattern := range patterns {
		// Check if pattern applies to this language
		if !patternMatchesLanguage(pattern, lang) {
			continue
		}

		if pattern.compiled == nil {
			continue
		}

		// Search each line
		for lineNum, line := range lines {
			if pattern.compiled.MatchString(line) {
				finding := APIFinding{
					RuleID:      fmt.Sprintf("rag-%s-%s", pattern.Category, sanitizeRuleID(pattern.Name)),
					Title:       pattern.Name,
					Description: fmt.Sprintf("Potential %s vulnerability detected", pattern.Category),
					Severity:    pattern.Severity,
					Confidence:  confidenceToString(pattern.Confidence),
					File:        relPath,
					Line:        lineNum + 1,
					Snippet:     truncateSnippet(line, 200),
					Category:    pattern.Category,
					OWASPApi:    mapToOWASPAPI2023(pattern.OWASPApi),
					Framework:   detectFramework(string(content)),
				}

				if pattern.CWE != "" {
					finding.CWE = []string{pattern.CWE}
				}

				// Try to extract endpoint info
				if endpoint := extractEndpoint(line); endpoint != "" {
					finding.Endpoint = endpoint
				}
				if method := extractHTTPMethod(line); method != "" {
					finding.HTTPMethod = method
				}

				findings = append(findings, finding)
			}
		}
	}

	return findings
}

// detectLanguage determines the programming language from file extension
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".go":
		return "go"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	default:
		return ""
	}
}

// patternMatchesLanguage checks if a pattern applies to a language
func patternMatchesLanguage(pattern APIPattern, lang string) bool {
	if len(pattern.Languages) == 0 {
		return true // Pattern applies to all languages
	}
	for _, l := range pattern.Languages {
		if l == lang {
			return true
		}
	}
	return false
}

// sanitizeRuleID creates a valid rule ID from a pattern name
func sanitizeRuleID(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return strings.ToLower(re.ReplaceAllString(name, "-"))
}

// confidenceToString converts numeric confidence to string
func confidenceToString(conf int) string {
	if conf >= 80 {
		return "high"
	} else if conf >= 50 {
		return "medium"
	}
	return "low"
}

// truncateSnippet truncates a code snippet to a maximum length
func truncateSnippet(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// mapToOWASPAPI2023 maps various OWASP references to API Top 10 2023
func mapToOWASPAPI2023(owasp string) string {
	// Already in API format
	if strings.HasPrefix(owasp, "API") {
		return owasp
	}

	// Map OWASP Web Top 10 to API Top 10
	mappings := map[string]string{
		"A01:2021": "API5:2023",  // Broken Access Control -> Broken Function Level Auth
		"A02:2021": "API2:2023",  // Cryptographic Failures -> Broken Authentication
		"A03:2021": "API8:2023",  // Injection -> Security Misconfiguration
		"A04:2021": "API8:2023",  // Insecure Design -> Security Misconfiguration
		"A05:2021": "API8:2023",  // Security Misconfiguration
		"A06:2021": "API9:2023",  // Vulnerable Components -> Improper Inventory
		"A07:2021": "API2:2023",  // Auth Failures -> Broken Authentication
		"A08:2021": "API10:2023", // Software/Data Integrity -> Unsafe Consumption
		"A09:2021": "API8:2023",  // Logging Failures -> Security Misconfiguration
		"A10:2021": "API7:2023",  // SSRF
	}

	if mapped, ok := mappings[owasp]; ok {
		return mapped
	}
	return owasp
}

// detectFramework attempts to detect the API framework from file content
func detectFramework(content string) string {
	frameworks := map[string][]string{
		"express":  {"require('express')", "require(\"express\")", "from 'express'", "import express"},
		"fastapi":  {"from fastapi", "FastAPI()", "@app.get", "@app.post"},
		"flask":    {"from flask", "Flask(__name__)", "@app.route"},
		"django":   {"from django", "django.urls", "urlpatterns"},
		"gin":      {"github.com/gin-gonic/gin", "gin.Default()"},
		"chi":      {"github.com/go-chi/chi", "chi.NewRouter()"},
		"spring":   {"@RestController", "@RequestMapping", "@GetMapping", "@PostMapping"},
		"rails":    {"Rails.application", "ActionController", "def index"},
		"laravel":  {"Route::get", "Route::post", "Illuminate\\"},
		"fastify":  {"require('fastify')", "fastify()", "import fastify"},
		"nestjs":   {"@Controller", "@Get(", "@Post(", "@nestjs/"},
		"graphql":  {"graphql", "GraphQLSchema", "type Query", "type Mutation"},
		"apollo":   {"ApolloServer", "apollo-server"},
	}

	for framework, patterns := range frameworks {
		for _, pattern := range patterns {
			if strings.Contains(content, pattern) {
				return framework
			}
		}
	}
	return ""
}

// extractEndpoint attempts to extract an API endpoint from a line of code
func extractEndpoint(line string) string {
	// Match common route patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`['"](/api/[^'"]+)['"]`),
		regexp.MustCompile(`['"](/v\d+/[^'"]+)['"]`),
		regexp.MustCompile(`['"](/[a-z]+/:[^'"]+)['"]`),
		regexp.MustCompile(`path\s*=\s*['"]([^'"]+)['"]`),
	}

	for _, re := range patterns {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// extractHTTPMethod attempts to extract the HTTP method from a line
func extractHTTPMethod(line string) string {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	lineLower := strings.ToLower(line)

	for _, method := range methods {
		if strings.Contains(lineLower, "."+strings.ToLower(method)+"(") ||
			strings.Contains(lineLower, "@"+strings.ToLower(method)) ||
			strings.Contains(lineLower, "method=\""+method+"\"") ||
			strings.Contains(lineLower, "method='"+method+"'") {
			return method
		}
	}
	return ""
}
