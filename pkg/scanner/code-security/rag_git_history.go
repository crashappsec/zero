package codesecurity

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/crashappsec/zero/pkg/core/rag"
)

// RAGGitHistoryPattern represents a pattern loaded from RAG files for git history scanning
type RAGGitHistoryPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	RawPattern  string
	Severity    string
	Category    string
	Description string
	Type        string // "filepath" for file path patterns
}

// ragGitHistoryCache holds cached RAG git history patterns
var ragGitHistoryCache struct {
	sync.RWMutex
	gitignorePatterns  []*RAGGitHistoryPattern
	sensitivePatterns  []*RAGGitHistoryPattern
	loaded             bool
}

// LoadRAGGitHistoryPatterns loads all git history security patterns from RAG files
// Returns cached patterns if already loaded
func LoadRAGGitHistoryPatterns() (gitignore []*RAGGitHistoryPattern, sensitive []*RAGGitHistoryPattern, err error) {
	ragGitHistoryCache.RLock()
	if ragGitHistoryCache.loaded {
		gitignore = ragGitHistoryCache.gitignorePatterns
		sensitive = ragGitHistoryCache.sensitivePatterns
		ragGitHistoryCache.RUnlock()
		return gitignore, sensitive, nil
	}
	ragGitHistoryCache.RUnlock()

	ragGitHistoryCache.Lock()
	defer ragGitHistoryCache.Unlock()

	// Double-check after acquiring write lock
	if ragGitHistoryCache.loaded {
		return ragGitHistoryCache.gitignorePatterns, ragGitHistoryCache.sensitivePatterns, nil
	}

	// Find RAG directory
	ragPath := rag.FindRAGPath()
	if ragPath == "" {
		return nil, nil, nil
	}

	gitHistoryDir := filepath.Join(ragPath, "devops", "git-history-security")
	if _, err := os.Stat(gitHistoryDir); os.IsNotExist(err) {
		return nil, nil, nil
	}

	// Load gitignore patterns
	gitignorePath := filepath.Join(gitHistoryDir, "gitignore-best-practices.md")
	if _, err := os.Stat(gitignorePath); err == nil {
		patterns, _ := parseGitHistoryPatterns(gitignorePath)
		ragGitHistoryCache.gitignorePatterns = patterns
	}

	// Load sensitive file patterns
	sensitivePath := filepath.Join(gitHistoryDir, "sensitive-files.md")
	if _, err := os.Stat(sensitivePath); err == nil {
		patterns, _ := parseGitHistoryPatterns(sensitivePath)
		ragGitHistoryCache.sensitivePatterns = patterns
	}

	ragGitHistoryCache.loaded = true

	return ragGitHistoryCache.gitignorePatterns, ragGitHistoryCache.sensitivePatterns, nil
}

// ClearRAGGitHistoryCache clears the cached patterns (useful for testing)
func ClearRAGGitHistoryCache() {
	ragGitHistoryCache.Lock()
	defer ragGitHistoryCache.Unlock()
	ragGitHistoryCache.gitignorePatterns = nil
	ragGitHistoryCache.sensitivePatterns = nil
	ragGitHistoryCache.loaded = false
}

// parseGitHistoryPatterns parses git history security patterns from a markdown file
func parseGitHistoryPatterns(path string) ([]*RAGGitHistoryPattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []*RAGGitHistoryPattern
	scanner := bufio.NewScanner(file)

	// Parse state tracking
	var currentName string
	var currentPattern string
	var currentSeverity string
	var currentCategory string
	var currentType string
	var currentDescription string
	var currentSectionCategory string // Category from ## header
	var inPatternBlock bool

	// Regex for parsing
	sectionHeaderRe := regexp.MustCompile(`^###\s+(.+)$`)
	majorSectionRe := regexp.MustCompile(`^##\s+(.+)$`)
	typeRe := regexp.MustCompile(`\*\*Type\*\*:\s*(\w+)`)
	severityRe := regexp.MustCompile(`\*\*Severity\*\*:\s*(\w+)`)
	categoryRe := regexp.MustCompile(`\*\*Category\*\*:\s*(\w+)`)
	patternRe := regexp.MustCompile(`\*\*Pattern\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")

	// Helper to save current pattern
	saveCurrentPattern := func() {
		if currentName != "" && currentPattern != "" {
			// Convert filepath pattern to regex if needed
			regexPattern := currentPattern
			if currentType == "filepath" {
				regexPattern = filepathToRegex(currentPattern)
			}

			// Use explicit category if set, otherwise derive from section header
			category := currentCategory
			if category == "" && currentSectionCategory != "" {
				category = sectionToCategory(currentSectionCategory)
			}

			compiled, err := regexp.Compile("(?i)" + regexPattern)
			if err == nil {
				patterns = append(patterns, &RAGGitHistoryPattern{
					Name:        normalizePatternName(currentName),
					Pattern:     compiled,
					RawPattern:  currentPattern,
					Severity:    normalizeGitHistorySeverity(currentSeverity),
					Category:    category,
					Description: currentDescription,
					Type:        currentType,
				})
			}
		}
		currentName = ""
		currentPattern = ""
		currentSeverity = ""
		currentCategory = ""
		currentType = ""
		currentDescription = ""
		inPatternBlock = false
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and non-pattern content
		if trimmed == "" || strings.HasPrefix(trimmed, "---") {
			continue
		}

		// New pattern section (### header)
		if m := sectionHeaderRe.FindStringSubmatch(trimmed); m != nil {
			saveCurrentPattern()
			currentName = m[1]
			inPatternBlock = true
			continue
		}

		// ## header sets section category and ends current pattern block
		if m := majorSectionRe.FindStringSubmatch(trimmed); m != nil {
			saveCurrentPattern()
			currentSectionCategory = m[1]
			inPatternBlock = false
			continue
		}

		if !inPatternBlock {
			continue
		}

		// Type line
		if m := typeRe.FindStringSubmatch(trimmed); m != nil {
			currentType = strings.ToLower(m[1])
			continue
		}

		// Severity line
		if m := severityRe.FindStringSubmatch(trimmed); m != nil {
			currentSeverity = m[1]
			continue
		}

		// Category line
		if m := categoryRe.FindStringSubmatch(trimmed); m != nil {
			currentCategory = m[1]
			continue
		}

		// Pattern line
		if m := patternRe.FindStringSubmatch(trimmed); m != nil {
			currentPattern = m[1]
			continue
		}

		// Description line (first line starting with -)
		if strings.HasPrefix(trimmed, "- ") && currentDescription == "" {
			currentDescription = strings.TrimPrefix(trimmed, "- ")
			continue
		}
	}

	// Save last pattern if any
	saveCurrentPattern()

	return patterns, nil
}

// filepathToRegex converts a filepath glob pattern to a regex
func filepathToRegex(pattern string) string {
	// The patterns in our RAG files are already regex-like
	// Just ensure proper anchoring
	if !strings.HasPrefix(pattern, "^") && !strings.HasPrefix(pattern, "(") {
		// Allow matching anywhere in path
		pattern = "(^|/)" + pattern
	}
	return pattern
}

// normalizePatternName converts a pattern name to a standardized format
func normalizePatternName(name string) string {
	// Convert to lowercase and replace spaces with underscores
	normalized := strings.ToLower(name)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	// Remove any non-alphanumeric characters except underscores
	var result strings.Builder
	for _, c := range normalized {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
			result.WriteRune(c)
		}
	}
	return result.String()
}

// normalizeGitHistorySeverity normalizes severity to standard values
func normalizeGitHistorySeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	case "info":
		return "info"
	default:
		return "medium"
	}
}

// sectionToCategory converts a section header to a category
func sectionToCategory(section string) string {
	lower := strings.ToLower(section)
	switch {
	case strings.Contains(lower, "environment") || strings.Contains(lower, "credential") ||
		strings.Contains(lower, "secret") || strings.Contains(lower, "password"):
		return "credentials"
	case strings.Contains(lower, "ssh") || strings.Contains(lower, "ssl") ||
		strings.Contains(lower, "key") || strings.Contains(lower, "certificate"):
		return "keys"
	case strings.Contains(lower, "cloud") || strings.Contains(lower, "aws") ||
		strings.Contains(lower, "gcp") || strings.Contains(lower, "azure"):
		return "credentials"
	case strings.Contains(lower, "database") || strings.Contains(lower, "sql"):
		return "database"
	case strings.Contains(lower, "terraform") || strings.Contains(lower, "ansible") ||
		strings.Contains(lower, "infrastructure") || strings.Contains(lower, "iac"):
		return "infrastructure"
	case strings.Contains(lower, "build") || strings.Contains(lower, "artifact") ||
		strings.Contains(lower, "vendor") || strings.Contains(lower, "node_modules"):
		return "build_artifact"
	case strings.Contains(lower, "ide") || strings.Contains(lower, "editor"):
		return "ide"
	case strings.Contains(lower, "log") || strings.Contains(lower, "temp"):
		return "logs"
	case strings.Contains(lower, "docker") || strings.Contains(lower, "kubernetes") ||
		strings.Contains(lower, "container"):
		return "configuration"
	case strings.Contains(lower, "backup"):
		return "backup"
	default:
		return "configuration"
	}
}

// GetRAGGitHistoryPatternCounts returns the count of loaded patterns
func GetRAGGitHistoryPatternCounts() (gitignore int, sensitive int) {
	ragGitHistoryCache.RLock()
	defer ragGitHistoryCache.RUnlock()
	return len(ragGitHistoryCache.gitignorePatterns), len(ragGitHistoryCache.sensitivePatterns)
}

// ConvertToSensitiveFilePatterns converts RAG patterns to sensitiveFilePattern structs
// for use with the git history security scanner
func ConvertToSensitiveFilePatterns(ragPatterns []*RAGGitHistoryPattern) []sensitiveFilePattern {
	result := make([]sensitiveFilePattern, 0, len(ragPatterns))
	for _, rp := range ragPatterns {
		result = append(result, sensitiveFilePattern{
			Pattern:     rp.RawPattern,
			Regex:       rp.Pattern,
			Category:    rp.Category,
			Severity:    rp.Severity,
			Description: rp.Description,
		})
	}
	return result
}
