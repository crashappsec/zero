package codesecurity

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// EntropyAnalyzer detects high-entropy strings that may be secrets
type EntropyAnalyzer struct {
	config EntropyConfig
}

// NewEntropyAnalyzer creates a new entropy analyzer
func NewEntropyAnalyzer(config EntropyConfig) *EntropyAnalyzer {
	return &EntropyAnalyzer{config: config}
}

// EntropyResult holds results from entropy analysis
type EntropyResult struct {
	Findings []SecretFinding
	Summary  struct {
		FilesScanned  int
		HighEntropy   int
		MediumEntropy int
	}
}

// CalculateEntropy computes Shannon entropy of a string
// Returns a value between 0 and 8 (for base-256 character set)
func CalculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	// Count character frequencies
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}

	// Calculate entropy
	length := float64(len(s))
	entropy := 0.0
	for _, count := range freq {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}

	return entropy
}

// GetEntropyLevel returns "high", "medium", or "low" based on entropy value
func (a *EntropyAnalyzer) GetEntropyLevel(entropy float64) string {
	if entropy >= a.config.HighThreshold {
		return "high"
	}
	if entropy >= a.config.MedThreshold {
		return "medium"
	}
	return "low"
}

// ScanDirectory scans all code files in a directory for high-entropy strings
func (a *EntropyAnalyzer) ScanDirectory(dir string) (*EntropyResult, error) {
	result := &EntropyResult{}

	// File extensions to scan
	codeExtensions := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".py": true, ".rb": true, ".java": true, ".c": true, ".cpp": true,
		".cs": true, ".php": true, ".rs": true, ".swift": true, ".kt": true,
		".scala": true, ".sh": true, ".bash": true, ".zsh": true,
		".yaml": true, ".yml": true, ".json": true, ".xml": true,
		".env": true, ".config": true, ".ini": true, ".toml": true,
		".properties": true, ".conf": true,
	}

	// Skip directories
	skipDirs := map[string]bool{
		"node_modules": true, "vendor": true, ".git": true,
		"dist": true, "build": true, "__pycache__": true,
		".venv": true, "venv": true, ".tox": true,
		"target": true, "bin": true, "obj": true,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check extension
		ext := strings.ToLower(filepath.Ext(path))
		if !codeExtensions[ext] {
			return nil
		}

		// Skip large files
		if info.Size() > 1024*1024 { // 1MB limit
			return nil
		}

		findings := a.ScanFile(path)
		result.Findings = append(result.Findings, findings...)
		result.Summary.FilesScanned++

		return nil
	})

	// Count by level
	for _, f := range result.Findings {
		switch f.EntropyLevel {
		case "high":
			result.Summary.HighEntropy++
		case "medium":
			result.Summary.MediumEntropy++
		}
	}

	return result, err
}

// ScanFile scans a single file for high-entropy strings
func (a *EntropyAnalyzer) ScanFile(path string) []SecretFinding {
	var findings []SecretFinding

	file, err := os.Open(path)
	if err != nil {
		return findings
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Find potential secrets in the line
		candidates := a.extractCandidates(line)
		for _, candidate := range candidates {
			entropy := CalculateEntropy(candidate.value)
			level := a.GetEntropyLevel(entropy)

			if level == "low" {
				continue
			}

			// Skip if it's a known false positive
			if a.isFalsePositive(candidate.value, line) {
				continue
			}

			// Determine secret type based on patterns
			secretType := a.inferSecretType(candidate.value, line)

			findings = append(findings, SecretFinding{
				RuleID:          "entropy-detection",
				Type:            secretType,
				Severity:        a.getSeverity(level),
				Message:         "High-entropy string detected - potential secret",
				File:            path,
				Line:            lineNum,
				Column:          candidate.column,
				Snippet:         a.redactSnippet(candidate.value),
				Entropy:         entropy,
				EntropyLevel:    level,
				DetectionSource: "entropy",
			})
		}
	}

	return findings
}

// candidate represents a potential secret string
type candidate struct {
	value  string
	column int
}

// extractCandidates finds potential secret strings in a line
func (a *EntropyAnalyzer) extractCandidates(line string) []candidate {
	var candidates []candidate

	// Pattern 1: Quoted strings (single or double quotes)
	quotedPattern := regexp.MustCompile(`["']([^"']{16,})["']`)
	matches := quotedPattern.FindAllStringSubmatchIndex(line, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			value := line[match[2]:match[3]]
			if len(value) >= a.config.MinLength {
				candidates = append(candidates, candidate{value: value, column: match[2] + 1})
			}
		}
	}

	// Pattern 2: Assignment values (key = value, key: value)
	assignPattern := regexp.MustCompile(`[=:]\s*["']?([A-Za-z0-9+/=_\-]{16,})["']?`)
	matches = assignPattern.FindAllStringSubmatchIndex(line, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			value := line[match[2]:match[3]]
			if len(value) >= a.config.MinLength && !a.isDuplicate(value, candidates) {
				candidates = append(candidates, candidate{value: value, column: match[2] + 1})
			}
		}
	}

	// Pattern 3: Base64-like strings
	base64Pattern := regexp.MustCompile(`[A-Za-z0-9+/=]{32,}`)
	matches = base64Pattern.FindAllStringSubmatchIndex(line, -1)
	for _, match := range matches {
		value := line[match[0]:match[1]]
		if len(value) >= a.config.MinLength && !a.isDuplicate(value, candidates) {
			candidates = append(candidates, candidate{value: value, column: match[0] + 1})
		}
	}

	// Pattern 4: Hex strings
	hexPattern := regexp.MustCompile(`[0-9a-fA-F]{32,}`)
	matches = hexPattern.FindAllStringSubmatchIndex(line, -1)
	for _, match := range matches {
		value := line[match[0]:match[1]]
		if len(value) >= a.config.MinLength && !a.isDuplicate(value, candidates) {
			candidates = append(candidates, candidate{value: value, column: match[0] + 1})
		}
	}

	return candidates
}

// isDuplicate checks if value already exists in candidates
func (a *EntropyAnalyzer) isDuplicate(value string, candidates []candidate) bool {
	for _, c := range candidates {
		if c.value == value || strings.Contains(c.value, value) || strings.Contains(value, c.value) {
			return true
		}
	}
	return false
}

// isFalsePositive checks if a string is a known false positive
func (a *EntropyAnalyzer) isFalsePositive(value, context string) bool {
	valueLower := strings.ToLower(value)
	contextLower := strings.ToLower(context)

	// Placeholder patterns
	placeholders := []string{
		"example", "test", "sample", "demo", "placeholder",
		"your_", "your-", "xxx", "abc123", "changeme",
		"password123", "secretkey", "apikey123",
		"replace", "insert", "enter_", "put_your",
	}
	for _, p := range placeholders {
		if strings.Contains(valueLower, p) {
			return true
		}
	}

	// Known example keys
	exampleKeys := []string{
		"akiaiosfodnn7example",   // AWS example key
		"wjalrxutnfemi/k7mdeng", // AWS example secret
		"ghp_examplexxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", // GitHub example
		"sk_test_", "pk_test_", // Stripe test keys
		"xoxb-example", "xoxp-example", // Slack example
	}
	for _, key := range exampleKeys {
		if strings.Contains(valueLower, key) {
			return true
		}
	}

	// UUID pattern (v4)
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if uuidPattern.MatchString(valueLower) {
		return true
	}

	// Git SHA pattern
	gitShaPattern := regexp.MustCompile(`^[0-9a-f]{40}$`)
	if gitShaPattern.MatchString(valueLower) {
		// Check context for git-related keywords
		if strings.Contains(contextLower, "commit") || strings.Contains(contextLower, "sha") ||
			strings.Contains(contextLower, "ref") || strings.Contains(contextLower, "hash") {
			return true
		}
	}

	// Hash output indicators in context
	hashIndicators := []string{
		"sha256", "sha512", "sha1", "md5", "hash",
		"checksum", "digest", "fingerprint",
	}
	for _, indicator := range hashIndicators {
		if strings.Contains(contextLower, indicator) {
			return true
		}
	}

	// All same character (unlikely to be a real secret)
	if isAllSameChar(value) {
		return true
	}

	// Only alphanumeric with no special chars and too uniform
	if isUniformAlphanumeric(value) {
		return true
	}

	// Test file indicators
	testIndicators := []string{"_test.", "test_", "/test/", "/tests/", "mock", "fixture", "spec."}
	for _, indicator := range testIndicators {
		if strings.Contains(contextLower, indicator) {
			return true
		}
	}

	return false
}

// isAllSameChar checks if string is all the same character
func isAllSameChar(s string) bool {
	if len(s) == 0 {
		return true
	}
	first := rune(s[0])
	for _, c := range s {
		if c != first {
			return false
		}
	}
	return true
}

// isUniformAlphanumeric checks if string is too uniform to be a secret
func isUniformAlphanumeric(s string) bool {
	if len(s) < 16 {
		return false
	}

	digits := 0
	letters := 0
	for _, c := range s {
		if unicode.IsDigit(c) {
			digits++
		} else if unicode.IsLetter(c) {
			letters++
		}
	}

	total := len(s)
	// If it's all letters or all digits with no mixing, likely not a secret
	if digits == total || letters == total {
		return true
	}

	return false
}

// inferSecretType tries to determine the type of secret based on patterns
func (a *EntropyAnalyzer) inferSecretType(value, context string) string {
	valueLower := strings.ToLower(value)
	contextLower := strings.ToLower(context)

	// AWS patterns
	if strings.HasPrefix(value, "AKIA") || strings.HasPrefix(value, "ABIA") ||
		strings.HasPrefix(value, "ACCA") || strings.HasPrefix(value, "ASIA") {
		return "aws_access_key"
	}
	if strings.Contains(contextLower, "aws") && len(value) == 40 {
		return "aws_secret_key"
	}

	// GitHub patterns
	if strings.HasPrefix(value, "ghp_") || strings.HasPrefix(value, "gho_") ||
		strings.HasPrefix(value, "ghu_") || strings.HasPrefix(value, "ghs_") ||
		strings.HasPrefix(value, "ghr_") {
		return "github_token"
	}

	// Stripe patterns
	if strings.HasPrefix(valueLower, "sk_live_") || strings.HasPrefix(valueLower, "rk_live_") {
		return "stripe_secret_key"
	}

	// Slack patterns
	if strings.HasPrefix(valueLower, "xoxb-") || strings.HasPrefix(valueLower, "xoxp-") ||
		strings.HasPrefix(valueLower, "xoxa-") || strings.HasPrefix(valueLower, "xoxr-") {
		return "slack_token"
	}

	// OpenAI/Anthropic patterns
	if strings.HasPrefix(valueLower, "sk-") && len(value) > 40 {
		if strings.Contains(contextLower, "anthropic") {
			return "anthropic_api_key"
		}
		return "openai_api_key"
	}

	// Private key patterns
	if strings.Contains(contextLower, "private") && strings.Contains(contextLower, "key") {
		return "private_key"
	}

	// Database patterns
	if strings.Contains(contextLower, "database") || strings.Contains(contextLower, "db_") ||
		strings.Contains(contextLower, "mysql") || strings.Contains(contextLower, "postgres") ||
		strings.Contains(contextLower, "mongo") {
		return "database_credential"
	}

	// JWT patterns
	if strings.Contains(contextLower, "jwt") || strings.HasPrefix(value, "eyJ") {
		return "jwt_token"
	}

	// Generic API key
	if strings.Contains(contextLower, "api_key") || strings.Contains(contextLower, "apikey") ||
		strings.Contains(contextLower, "api-key") {
		return "api_key"
	}

	// Generic secret
	if strings.Contains(contextLower, "secret") || strings.Contains(contextLower, "password") ||
		strings.Contains(contextLower, "token") || strings.Contains(contextLower, "credential") {
		return "generic_secret"
	}

	return "high_entropy_string"
}

// getSeverity returns severity based on entropy level
func (a *EntropyAnalyzer) getSeverity(level string) string {
	switch level {
	case "high":
		return "medium" // High entropy but not confirmed secret
	case "medium":
		return "low"
	default:
		return "info"
	}
}

// redactSnippet redacts the middle portion of a secret
func (a *EntropyAnalyzer) redactSnippet(value string) string {
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}
