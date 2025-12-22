package codesecurity

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/crashappsec/zero/pkg/analysis/rag"
)

// RAGSecretPattern represents a secret pattern loaded from RAG files
type RAGSecretPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	RawPattern  string
	Severity    string
	Description string
	Technology  string
	Category    string
}

// ragSecretsCache holds cached RAG secret patterns
var ragSecretsCache struct {
	sync.RWMutex
	patterns []*RAGSecretPattern
	loaded   bool
}

// LoadRAGSecretPatterns loads all secret patterns from RAG patterns.md files
// Returns cached patterns if already loaded
func LoadRAGSecretPatterns() ([]*RAGSecretPattern, error) {
	ragSecretsCache.RLock()
	if ragSecretsCache.loaded {
		patterns := ragSecretsCache.patterns
		ragSecretsCache.RUnlock()
		return patterns, nil
	}
	ragSecretsCache.RUnlock()

	ragSecretsCache.Lock()
	defer ragSecretsCache.Unlock()

	// Double-check after acquiring write lock
	if ragSecretsCache.loaded {
		return ragSecretsCache.patterns, nil
	}

	// Find RAG directory
	ragPath := rag.FindRAGPath()
	if ragPath == "" {
		// Return empty slice if RAG not found - we'll fall back to hardcoded patterns
		return nil, nil
	}

	techIDDir := filepath.Join(ragPath, "technology-identification")
	if _, err := os.Stat(techIDDir); os.IsNotExist(err) {
		return nil, nil
	}

	// Find all patterns.md files
	patternFiles, err := findPatternFiles(techIDDir)
	if err != nil {
		return nil, err
	}

	var allPatterns []*RAGSecretPattern
	for _, pf := range patternFiles {
		patterns, err := parseSecretsFromFile(pf, ragPath)
		if err != nil {
			continue // Skip files that fail to parse
		}
		allPatterns = append(allPatterns, patterns...)
	}

	ragSecretsCache.patterns = allPatterns
	ragSecretsCache.loaded = true

	return allPatterns, nil
}

// ClearRAGSecretsCache clears the cached patterns (useful for testing)
func ClearRAGSecretsCache() {
	ragSecretsCache.Lock()
	defer ragSecretsCache.Unlock()
	ragSecretsCache.patterns = nil
	ragSecretsCache.loaded = false
}

// findPatternFiles recursively finds all patterns.md files in directory
func findPatternFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && info.Name() == "patterns.md" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// parseSecretsFromFile parses secret patterns from a single patterns.md file
func parseSecretsFromFile(path, ragPath string) ([]*RAGSecretPattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []*RAGSecretPattern

	scanner := bufio.NewScanner(file)

	// Extract technology name and category from path
	relPath, _ := filepath.Rel(ragPath, path)
	relPath = strings.TrimPrefix(relPath, "technology-identification/")
	relPath = strings.TrimSuffix(relPath, "/patterns.md")
	parts := strings.Split(relPath, "/")

	category := ""
	technology := ""
	if len(parts) >= 1 {
		category = parts[0]
	}
	if len(parts) >= 2 {
		technology = parts[len(parts)-1]
	}

	// Parse state tracking
	var techName string
	var inSecretsSection bool
	var currentSecretName string
	var currentPattern string
	var currentSeverity string
	var currentDescription string

	// Regex for parsing
	nameRe := regexp.MustCompile(`^#\s+(.+)$`)
	secretHeaderRe := regexp.MustCompile(`^####\s+(.+)$`)
	patternRe := regexp.MustCompile(`\*\*Pattern\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	severityRe := regexp.MustCompile(`\*\*Severity\*\*:\s*(\w+)`)
	descRe := regexp.MustCompile(`\*\*Description\*\*:\s*(.+)`)

	// Helper to save current pattern
	saveCurrentPattern := func() {
		if currentSecretName != "" && currentPattern != "" {
			compiled, err := regexp.Compile(currentPattern)
			if err == nil {
				patterns = append(patterns, &RAGSecretPattern{
					Name:        normalizeSecretName(currentSecretName, technology),
					Pattern:     compiled,
					RawPattern:  currentPattern,
					Severity:    normalizeSeverity(currentSeverity),
					Description: currentDescription,
					Technology:  technology,
					Category:    category,
				})
			}
		}
		currentSecretName = ""
		currentPattern = ""
		currentSeverity = ""
		currentDescription = ""
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Extract technology name from first heading
		if techName == "" {
			if m := nameRe.FindStringSubmatch(trimmed); m != nil {
				techName = m[1]
				if technology == "" {
					technology = strings.ToLower(strings.ReplaceAll(techName, " ", "_"))
				}
				continue
			}
		}

		// Track section changes
		if strings.HasPrefix(trimmed, "## ") {
			// Save any pending pattern before changing sections
			saveCurrentPattern()
			section := strings.TrimPrefix(trimmed, "## ")
			inSecretsSection = strings.Contains(section, "Secrets Detection")
			continue
		}

		// Only process lines in Secrets Detection section
		if !inSecretsSection {
			continue
		}

		// New secret pattern header (#### level)
		if m := secretHeaderRe.FindStringSubmatch(trimmed); m != nil {
			// Save previous pattern before starting new one
			saveCurrentPattern()
			currentSecretName = m[1]
			continue
		}

		// Pattern line
		if m := patternRe.FindStringSubmatch(trimmed); m != nil {
			currentPattern = m[1]
			continue
		}

		// Severity line
		if m := severityRe.FindStringSubmatch(trimmed); m != nil {
			currentSeverity = m[1]
			continue
		}

		// Description line
		if m := descRe.FindStringSubmatch(trimmed); m != nil {
			currentDescription = m[1]
			continue
		}
	}

	// Save last pattern if any
	saveCurrentPattern()

	return patterns, nil
}

// normalizeSecretName converts a secret name to a standardized format
func normalizeSecretName(name, technology string) string {
	// Convert to lowercase and replace spaces with underscores
	normalized := strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	// Build a consistent name format: technology_type
	tech := strings.ToLower(strings.ReplaceAll(technology, "-", "_"))

	// Handle common naming patterns - check most specific first
	switch {
	// Check AWS secret key before access key (secret access key contains both "secret" and "access_key")
	case strings.Contains(normalized, "secret") && strings.Contains(tech, "aws"):
		return "aws_secret_key"
	case strings.Contains(normalized, "access_key") && strings.Contains(tech, "aws"):
		return "aws_access_key"
	case strings.Contains(normalized, "session") && strings.Contains(tech, "aws"):
		return "aws_session_token"
	case strings.Contains(normalized, "api") && strings.Contains(tech, "openai"):
		return "openai_api_key"
	case strings.Contains(normalized, "api") && strings.Contains(tech, "anthropic"):
		return "anthropic_api_key"
	case strings.Contains(normalized, "api") && strings.Contains(tech, "google"):
		return "google_api_key"
	case strings.Contains(tech, "github") && (strings.Contains(normalized, "token") || strings.Contains(normalized, "access")):
		return "github_token"
	case strings.Contains(normalized, "secret") && strings.Contains(tech, "stripe"):
		return "stripe_secret_key"
	case strings.Contains(normalized, "publishable") && strings.Contains(tech, "stripe"):
		return "stripe_publishable_key"
	case strings.Contains(tech, "slack") && (strings.Contains(normalized, "token") || strings.Contains(normalized, "bot")):
		return "slack_token"
	case strings.Contains(normalized, "webhook") && strings.Contains(tech, "slack"):
		return "slack_webhook"
	case strings.Contains(normalized, "private") && strings.Contains(normalized, "key"):
		return "private_key"
	case strings.Contains(normalized, "jwt"):
		return "jwt_token"
	case strings.Contains(normalized, "database") || strings.Contains(normalized, "connection"):
		return "database_credential"
	}

	// Remove common suffixes before returning
	normalized = strings.TrimSuffix(normalized, "_key")
	normalized = strings.TrimSuffix(normalized, "_token")

	// Default: use technology_name format
	if tech != "" && !strings.HasPrefix(normalized, tech) {
		return tech + "_" + normalized
	}
	return normalized
}

// normalizeSeverity normalizes severity to standard values
func normalizeSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	default:
		return "medium" // Default to medium if not specified
	}
}

// GetRAGPatternCount returns the number of loaded RAG patterns (for logging)
func GetRAGPatternCount() int {
	ragSecretsCache.RLock()
	defer ragSecretsCache.RUnlock()
	return len(ragSecretsCache.patterns)
}

// GetRAGPatternSummary returns a summary of loaded patterns by category
func GetRAGPatternSummary() map[string]int {
	patterns, _ := LoadRAGSecretPatterns()
	summary := make(map[string]int)
	for _, p := range patterns {
		summary[p.Category]++
	}
	return summary
}
