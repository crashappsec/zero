// Package codesecurity provides the consolidated code security super scanner
package codesecurity

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/scanner/common"
)

// runAPIQualityWithSemgrep runs Semgrep-based API quality analysis
// Uses RAG-generated rules - no fallback patterns
func (s *CodeSecurityScanner) runAPIQualityWithSemgrep(ctx context.Context, opts *scanner.ScanOptions, cfg APIConfig) []APIFinding {
	var findings []APIFinding

	// Find RAG path for rule generation
	ragPath := findRAGPath()
	if ragPath == "" {
		// No RAG available - return empty
		return findings
	}

	// Generate rules from RAG patterns
	cacheDir := getCacheDir()
	rulesPath := filepath.Join(cacheDir, "rules", "code-quality-api.yaml")

	// Generate rules if needed
	if err := common.GenerateRulesFromRAG(ragPath, "code-quality/api", rulesPath); err != nil {
		// Rule generation failed
		return findings
	}

	// Check if rules were generated
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		return findings
	}

	// Check if Semgrep is available
	if !common.HasSemgrep() {
		return findings
	}

	// Run Semgrep
	runner := common.NewSemgrepRunner(common.SemgrepConfig{
		RulePaths: []string{rulesPath},
		OnStatus:  func(string) {},
	})

	result := runner.Run(ctx, opts.RepoPath)
	if result.Error != nil {
		return findings
	}

	// Convert Semgrep findings to APIFindings
	for _, f := range result.Findings {
		finding := APIFinding{
			RuleID:      f.RuleID,
			Title:       extractTitleFromRuleID(f.RuleID),
			Description: f.Message,
			Severity:    f.Severity,
			Confidence:  "medium",
			File:        f.File,
			Line:        f.Line,
			Snippet:     truncateSnippet(f.Match, 200),
			Category:    f.Category,
			Remediation: f.Remediation,
		}

		// Try to extract endpoint and method
		if endpoint := extractEndpoint(f.Match); endpoint != "" {
			finding.Endpoint = endpoint
		}
		if method := extractHTTPMethod(f.Match); method != "" {
			finding.HTTPMethod = method
		}

		findings = append(findings, finding)
	}

	return findings
}

// extractTitleFromRuleID converts a rule ID to a human-readable title
func extractTitleFromRuleID(ruleID string) string {
	// Convert "zero.code-quality.api.n-1-query-pattern" to "N+1 Query Pattern"
	parts := strings.Split(ruleID, ".")
	if len(parts) > 0 {
		title := parts[len(parts)-1]
		title = strings.ReplaceAll(title, "-", " ")
		return strings.Title(title)
	}
	return ruleID
}

// findRAGPath locates the RAG directory
func findRAGPath() string {
	candidates := []string{
		"rag",
		"../rag",
		"../../rag",
	}

	if zeroHome := os.Getenv("ZERO_HOME"); zeroHome != "" {
		candidates = append([]string{filepath.Join(zeroHome, "rag")}, candidates...)
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			return absPath
		}
	}

	return ""
}

// getCacheDir returns the cache directory for generated rules
func getCacheDir() string {
	if zeroHome := os.Getenv("ZERO_HOME"); zeroHome != "" {
		return filepath.Join(zeroHome, "cache")
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".zero", "cache")
}

// runAPIQualityChecks performs non-security API quality analysis
// Uses Semgrep with RAG-generated rules - requires Semgrep to be installed
func (s *CodeSecurityScanner) runAPIQualityChecks(ctx context.Context, opts *scanner.ScanOptions, cfg APIConfig) []APIFinding {
	return s.runAPIQualityWithSemgrep(ctx, opts, cfg)
}

// shouldSkipQualityDir checks if a directory should be skipped for quality checks
func shouldSkipQualityDir(name string) bool {
	skipDirs := []string{
		"node_modules", "vendor", ".git", "dist", "build",
		"coverage", "__pycache__", ".venv", "venv",
		"test", "tests", "__tests__", "spec", "specs",
	}
	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}
	return false
}

// isLikelyAPIFile checks if a file is likely to contain API routes
func isLikelyAPIFile(path string) bool {
	pathLower := strings.ToLower(path)

	// Positive indicators
	positivePatterns := []string{
		"route", "controller", "handler", "api", "endpoint",
		"server", "app", "router", "rest", "graphql",
	}
	for _, pattern := range positivePatterns {
		if strings.Contains(pathLower, pattern) {
			return true
		}
	}

	// Check common API file patterns
	filename := filepath.Base(pathLower)
	apiFilePatterns := []string{
		"index.js", "index.ts", "app.js", "app.ts",
		"server.js", "server.ts", "main.py", "app.py",
		"main.go", "handlers.go", "routes.go",
	}
	for _, pattern := range apiFilePatterns {
		if filename == pattern {
			return true
		}
	}

	return false
}

// shouldCheckQualityCategory determines if a quality category should be checked
func shouldCheckQualityCategory(category string, cfg APIConfig) bool {
	switch category {
	case "api-design":
		return cfg.CheckDesign
	case "api-performance":
		return cfg.CheckPerformance
	case "api-observability":
		return cfg.CheckObservability
	case "api-documentation":
		return cfg.CheckDocumentation
	default:
		return true
	}
}
