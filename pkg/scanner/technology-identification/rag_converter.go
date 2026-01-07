// Package techid provides the consolidated technology identification super scanner
// This file converts RAG markdown patterns to Semgrep YAML rules
package techid

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SemgrepRule represents a single semgrep rule
type SemgrepRule struct {
	ID            string                 `yaml:"id"`
	Message       string                 `yaml:"message"`
	Severity      string                 `yaml:"severity"`
	Languages     []string               `yaml:"languages"`
	Metadata      map[string]interface{} `yaml:"metadata,omitempty"`
	Pattern       string                 `yaml:"pattern,omitempty"`
	PatternEither []PatternItem          `yaml:"pattern-either,omitempty"`
	PatternRegex  string                 `yaml:"pattern-regex,omitempty"`
}

// PatternItem represents a pattern in pattern-either
type PatternItem struct {
	Pattern string `yaml:"pattern"`
}

// SemgrepRules represents the full rules file
type SemgrepRules struct {
	Rules []SemgrepRule `yaml:"rules"`
}

// RAGPattern holds parsed pattern data from markdown
type RAGPattern struct {
	Name            string
	Category        string
	Description     string
	Homepage        string
	Packages        map[string][]string         // ecosystem -> package names
	Imports         map[string][]ImportPattern  // language -> patterns
	ConfigFiles     []string                    // configuration file names
	EnvVars         []string
	Secrets         []SecretPattern
	SecurityPatterns []SecurityPattern          // regex patterns for security issues
	Confidence      map[string]int
}

// SecurityPattern represents a security detection pattern (regex-based)
type SecurityPattern struct {
	Name        string
	Pattern     string
	Severity    string
	Type        string // "regex", "semgrep", etc.
	Description string
}

// ImportPattern represents an import detection pattern
type ImportPattern struct {
	Pattern     string
	Description string
	Type        string
}

// SecretPattern represents a secret detection pattern
type SecretPattern struct {
	Name        string
	Pattern     string
	Severity    string
	Description string
}

// ConvertRAGToSemgrep converts all RAG patterns to semgrep rules
func ConvertRAGToSemgrep(ragDir, outputDir string) (*ConversionResult, error) {
	result := &ConversionResult{
		TechDiscovery: SemgrepRules{Rules: []SemgrepRule{}},
		Secrets:       SemgrepRules{Rules: []SemgrepRule{}},
		AIML:          SemgrepRules{Rules: []SemgrepRule{}},
		ConfigFiles:   SemgrepRules{Rules: []SemgrepRule{}},
		Cryptography:  SemgrepRules{Rules: []SemgrepRule{}},
		DevOps:        SemgrepRules{Rules: []SemgrepRule{}},
		CodeSecurity:  SemgrepRules{Rules: []SemgrepRule{}},
		SupplyChain:   SemgrepRules{Rules: []SemgrepRule{}},
	}

	// RAG directories to process - dynamically discover all subdirectories
	ragDirs, err := discoverRAGDirectories(ragDir)
	if err != nil {
		// Fall back to known directories if discovery fails
		ragDirs = []string{
			"technology-identification",
			"cryptography",
			"devops",
			"code-security",
			"code-quality",
			"supply-chain",
			"api-security",
			"certificate-analysis",
			"code-ownership",
			"dora-metrics",
			"domains",
			"architecture",
			"legal-review",
			"ai-ml",
			"ai-adoption",
			"tech-debt",
			"backend-engineering",
			"frontend-engineering",
			"brand",
			"cocomo",
			"dora",
			"personas",
			"semgrep",
		}
	}

	for _, dir := range ragDirs {
		dirPath := filepath.Join(ragDir, dir)
		patternFiles, err := findPatternFiles(dirPath)
		if err != nil {
			continue // Skip directories that don't exist
		}

		for _, pf := range patternFiles {
			pattern, err := parsePatternFile(pf)
			if err != nil {
				continue // Skip files that fail to parse
			}

			// Infer category from file path if not set
			if pattern.Category == "" {
				relPath, _ := filepath.Rel(ragDir, pf)
				parts := strings.Split(relPath, string(filepath.Separator))
				if len(parts) > 0 {
					pattern.Category = parts[0]
					if len(parts) > 1 {
						pattern.Category += "/" + parts[1]
					}
				}
			}

			rules := convertPatternToRules(pattern, pf, ragDir)
			for _, rule := range rules {
				// Categorize rules based on source directory and content
				cat := strings.ToLower(pattern.Category)
				switch {
				case strings.Contains(rule.ID, ".secret."):
					result.Secrets.Rules = append(result.Secrets.Rules, rule)
				case strings.HasSuffix(rule.ID, ".config"):
					result.ConfigFiles.Rules = append(result.ConfigFiles.Rules, rule)
				case strings.Contains(cat, "ai-ml") || strings.Contains(cat, "ai/ml"):
					result.AIML.Rules = append(result.AIML.Rules, rule)
				case strings.HasPrefix(cat, "cryptography"):
					result.Cryptography.Rules = append(result.Cryptography.Rules, rule)
				case strings.HasPrefix(cat, "devops") || strings.Contains(cat, "docker") || strings.Contains(cat, "iac"):
					result.DevOps.Rules = append(result.DevOps.Rules, rule)
				case strings.HasPrefix(cat, "code-security") || strings.HasPrefix(cat, "api-security") || strings.Contains(cat, "security"):
					result.CodeSecurity.Rules = append(result.CodeSecurity.Rules, rule)
				case strings.HasPrefix(cat, "supply-chain") || strings.Contains(cat, "package"):
					result.SupplyChain.Rules = append(result.SupplyChain.Rules, rule)
				default:
					result.TechDiscovery.Rules = append(result.TechDiscovery.Rules, rule)
				}
			}
		}
	}

	// Write output files
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	if len(result.TechDiscovery.Rules) > 0 {
		if err := writeYAML(filepath.Join(outputDir, "tech-discovery.yaml"), result.TechDiscovery); err != nil {
			return nil, err
		}
	}

	if len(result.Secrets.Rules) > 0 {
		if err := writeYAML(filepath.Join(outputDir, "secrets.yaml"), result.Secrets); err != nil {
			return nil, err
		}
	}

	if len(result.AIML.Rules) > 0 {
		if err := writeYAML(filepath.Join(outputDir, "ai-ml.yaml"), result.AIML); err != nil {
			return nil, err
		}
	}

	if len(result.ConfigFiles.Rules) > 0 {
		if err := writeYAML(filepath.Join(outputDir, "config-files.yaml"), result.ConfigFiles); err != nil {
			return nil, err
		}
	}

	if len(result.Cryptography.Rules) > 0 {
		if err := writeYAML(filepath.Join(outputDir, "cryptography.yaml"), result.Cryptography); err != nil {
			return nil, err
		}
	}

	if len(result.DevOps.Rules) > 0 {
		if err := writeYAML(filepath.Join(outputDir, "devops.yaml"), result.DevOps); err != nil {
			return nil, err
		}
	}

	if len(result.CodeSecurity.Rules) > 0 {
		if err := writeYAML(filepath.Join(outputDir, "code-security.yaml"), result.CodeSecurity); err != nil {
			return nil, err
		}
	}

	if len(result.SupplyChain.Rules) > 0 {
		if err := writeYAML(filepath.Join(outputDir, "supply-chain.yaml"), result.SupplyChain); err != nil {
			return nil, err
		}
	}

	result.TotalRules = len(result.TechDiscovery.Rules) + len(result.Secrets.Rules) + len(result.AIML.Rules) + len(result.ConfigFiles.Rules) +
		len(result.Cryptography.Rules) + len(result.DevOps.Rules) + len(result.CodeSecurity.Rules) + len(result.SupplyChain.Rules)
	return result, nil
}

// ConversionResult holds the conversion output
type ConversionResult struct {
	TechDiscovery SemgrepRules
	Secrets       SemgrepRules
	AIML          SemgrepRules
	ConfigFiles   SemgrepRules
	Cryptography  SemgrepRules
	DevOps        SemgrepRules
	CodeSecurity  SemgrepRules
	SupplyChain   SemgrepRules
	TotalRules    int
}

// findPatternFiles recursively finds all RAG pattern markdown files
// Finds all .md files except README files
func findPatternFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		// Must be a .md file
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		// Skip README files
		if strings.EqualFold(info.Name(), "readme.md") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// parsePatternFile parses a RAG markdown file and extracts ALL patterns
func parsePatternFile(path string) (*RAGPattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	pattern := &RAGPattern{
		Packages:    make(map[string][]string),
		Imports:     make(map[string][]ImportPattern),
		ConfigFiles: []string{},
		Confidence:  make(map[string]int),
	}

	scanner := bufio.NewScanner(file)
	var currentSection string
	var currentSubsection string
	var currentHeading3 string

	// Regex patterns for parsing
	nameRe := regexp.MustCompile(`^#\s+(.+)$`)
	categoryRe := regexp.MustCompile(`\*\*Category\*\*:\s*(.+)`)
	ecosystemRe := regexp.MustCompile(`\*\*Ecosystem\*\*:\s*(.+)`)
	descRe := regexp.MustCompile(`\*\*Description\*\*:\s*(.+)`)
	homepageRe := regexp.MustCompile(`\*\*Homepage\*\*:\s*(.+)`)
	packageRe := regexp.MustCompile(`^-\s*` + "`" + `([^` + "`" + `]+)` + "`")
	configFileRe := regexp.MustCompile(`^-\s*` + "`" + `([^` + "`" + `]+)` + "`")
	// Match **Pattern**: `...` anywhere in document
	patternRe := regexp.MustCompile(`\*\*Pattern\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	severityRe := regexp.MustCompile(`\*\*Severity\*\*:\s*(\w+)`)
	confidenceRe := regexp.MustCompile(`\*\*Confidence\*\*:\s*(\d+)`)
	typeRe := regexp.MustCompile(`\*\*Type\*\*:\s*(\w+)`)
	languagesRe := regexp.MustCompile(`\*\*Languages\*\*:\s*\[([^\]]+)\]`)

	var inSecretsSection bool
	var inConfigFilesSection bool
	var lastPatternSeverity string
	var lastPatternType string
	var lastPatternLanguages string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Parse name from first heading
		if pattern.Name == "" {
			if m := nameRe.FindStringSubmatch(trimmed); m != nil {
				pattern.Name = m[1]
				continue
			}
		}

		// Parse category (or infer from ecosystem)
		if m := categoryRe.FindStringSubmatch(trimmed); m != nil {
			pattern.Category = m[1]
			continue
		}
		if pattern.Category == "" {
			if m := ecosystemRe.FindStringSubmatch(trimmed); m != nil {
				pattern.Category = "supply-chain/" + strings.ToLower(m[1])
				continue
			}
		}

		// Parse description
		if m := descRe.FindStringSubmatch(trimmed); m != nil {
			pattern.Description = m[1]
			continue
		}

		// Parse homepage
		if m := homepageRe.FindStringSubmatch(trimmed); m != nil {
			pattern.Homepage = strings.TrimSpace(m[1])
			continue
		}

		// Track sections
		if strings.HasPrefix(trimmed, "## ") {
			currentSection = strings.TrimPrefix(trimmed, "## ")
			currentSubsection = ""
			currentHeading3 = ""
			inSecretsSection = strings.Contains(currentSection, "Secrets") && !strings.Contains(currentSection, "Security")
			inConfigFilesSection = strings.Contains(currentSection, "Configuration Files")
			continue
		}

		if strings.HasPrefix(trimmed, "### ") {
			currentSubsection = strings.TrimPrefix(trimmed, "### ")
			currentHeading3 = currentSubsection
			continue
		}

		if strings.HasPrefix(trimmed, "#### ") {
			currentHeading3 = strings.TrimPrefix(trimmed, "#### ")
			if inSecretsSection {
				pattern.Secrets = append(pattern.Secrets, SecretPattern{Name: currentHeading3})
			}
			continue
		}

		// Track metadata that applies to next pattern
		if m := severityRe.FindStringSubmatch(trimmed); m != nil {
			lastPatternSeverity = m[1]
		}
		if m := typeRe.FindStringSubmatch(trimmed); m != nil {
			lastPatternType = m[1]
		}
		if m := languagesRe.FindStringSubmatch(trimmed); m != nil {
			lastPatternLanguages = m[1]
		}
		if m := confidenceRe.FindStringSubmatch(trimmed); m != nil {
			if conf, err := fmt.Sscanf(m[1], "%d", new(int)); err == nil && conf > 0 {
				pattern.Confidence["default"] = conf
			}
		}

		// Parse ALL **Pattern**: markers as SecurityPatterns (most flexible)
		if m := patternRe.FindStringSubmatch(trimmed); m != nil {
			patternValue := m[1]
			patternName := currentHeading3
			if patternName == "" {
				patternName = currentSubsection
			}
			if patternName == "" {
				patternName = currentSection
			}
			if patternName == "" {
				patternName = "Pattern"
			}

			// Determine if this is an import pattern or security pattern
			isImportPattern := strings.Contains(currentSection, "Import Detection") ||
				strings.Contains(currentSection, "Import")

			if isImportPattern {
				lang := strings.ToLower(currentSubsection)
				if lang == "" {
					lang = "generic"
				}
				pattern.Imports[lang] = append(pattern.Imports[lang], ImportPattern{
					Pattern: patternValue,
				})
			} else {
				// Add as SecurityPattern (regex-based)
				secPat := SecurityPattern{
					Name:     patternName,
					Pattern:  patternValue,
					Severity: lastPatternSeverity,
					Type:     lastPatternType,
				}
				if lastPatternLanguages != "" {
					secPat.Description = "Languages: " + lastPatternLanguages
				}
				pattern.SecurityPatterns = append(pattern.SecurityPatterns, secPat)
			}

			// Reset metadata after use
			lastPatternSeverity = ""
			lastPatternType = ""
			lastPatternLanguages = ""
		}

		// Parse configuration files
		if inConfigFilesSection {
			if m := configFileRe.FindStringSubmatch(trimmed); m != nil {
				configFile := strings.TrimSpace(m[1])
				if configFile != "" {
					pattern.ConfigFiles = append(pattern.ConfigFiles, configFile)
				}
			}
		}

		// Parse packages
		if strings.Contains(currentSection, "Package Detection") {
			if m := packageRe.FindStringSubmatch(trimmed); m != nil {
				ecosystem := strings.ToLower(currentSubsection)
				pattern.Packages[ecosystem] = append(pattern.Packages[ecosystem], m[1])
			}
		}

		// Parse environment variables
		if strings.Contains(currentSection, "Environment Variables") {
			if m := packageRe.FindStringSubmatch(trimmed); m != nil {
				pattern.EnvVars = append(pattern.EnvVars, m[1])
			}
		}

		// Parse secrets section patterns
		if inSecretsSection && len(pattern.Secrets) > 0 {
			lastIdx := len(pattern.Secrets) - 1
			if m := patternRe.FindStringSubmatch(trimmed); m != nil {
				pattern.Secrets[lastIdx].Pattern = m[1]
			}
			if m := severityRe.FindStringSubmatch(trimmed); m != nil {
				pattern.Secrets[lastIdx].Severity = m[1]
			}
		}
	}

	return pattern, nil
}

// convertPatternToRules converts a parsed pattern to semgrep rules
func convertPatternToRules(pattern *RAGPattern, filePath, ragDir string) []SemgrepRule {
	var rules []SemgrepRule

	// Generate base ID from path
	relPath, _ := filepath.Rel(ragDir, filePath)
	baseID := pathToID(relPath)

	// Create import detection rules for each language
	for lang, imports := range pattern.Imports {
		if len(imports) == 0 {
			continue
		}

		semgrepLang := mapLanguage(lang)
		if semgrepLang == "" {
			continue
		}

		// Separate patterns into those that convert to semgrep patterns vs regex
		var semgrepPatterns []string
		var regexPatterns []string

		for _, imp := range imports {
			converted := regexToSemgrep(imp.Pattern, lang)
			if converted != "" {
				semgrepPatterns = append(semgrepPatterns, converted)
			} else {
				// Use pattern-regex for patterns that can't be converted
				// Clean up the regex pattern for semgrep
				cleanedRegex := cleanRegexForSemgrep(imp.Pattern)
				if cleanedRegex != "" {
					regexPatterns = append(regexPatterns, cleanedRegex)
				}
			}
		}

		// Create rule with semgrep patterns if any
		if len(semgrepPatterns) > 0 {
			ruleID := fmt.Sprintf("zero.%s.import.%s", baseID, semgrepLang)
			rule := SemgrepRule{
				ID:        ruleID,
				Message:   fmt.Sprintf("%s library import detected", pattern.Name),
				Severity:  "INFO",
				Languages: []string{semgrepLang},
				Metadata: map[string]interface{}{
					"technology":     pattern.Name,
					"category":       pattern.Category,
					"detection_type": "import",
					"confidence":     getConfidence(pattern, "import"),
				},
			}

			if len(semgrepPatterns) == 1 {
				rule.Pattern = semgrepPatterns[0]
			} else {
				for _, p := range semgrepPatterns {
					rule.PatternEither = append(rule.PatternEither, PatternItem{Pattern: p})
				}
			}

			rules = append(rules, rule)
		}

		// Create separate rules for regex patterns (usage patterns)
		for i, regexPat := range regexPatterns {
			ruleID := fmt.Sprintf("zero.%s.usage.%s.%d", baseID, semgrepLang, i+1)
			rule := SemgrepRule{
				ID:        ruleID,
				Message:   fmt.Sprintf("%s usage detected", pattern.Name),
				Severity:  "INFO",
				Languages: []string{semgrepLang},
				Metadata: map[string]interface{}{
					"technology":     pattern.Name,
					"category":       pattern.Category,
					"detection_type": "usage",
					"confidence":     getConfidence(pattern, "usage"),
				},
				PatternRegex: regexPat,
			}
			rules = append(rules, rule)
		}
	}

	// Create secret detection rules
	for _, secret := range pattern.Secrets {
		if secret.Pattern == "" {
			continue
		}

		ruleID := fmt.Sprintf("zero.%s.secret.%s", baseID, sanitizeID(secret.Name))

		// Map severity with metadata preservation
		mappedSev, sevMeta := mapSeverityWithMetadata(secret.Severity)

		// Build metadata
		metadata := map[string]interface{}{
			"technology":  pattern.Name,
			"category":    "secrets",
			"secret_type": secret.Name,
			"confidence":  95,
		}
		// Merge severity metadata
		for k, v := range sevMeta {
			metadata[k] = v
		}

		rule := SemgrepRule{
			ID:           ruleID,
			Message:      fmt.Sprintf("Potential %s %s exposed", pattern.Name, secret.Name),
			Severity:     mappedSev,
			Languages:    []string{"generic"},
			Metadata:     metadata,
			PatternRegex: secret.Pattern,
		}
		rules = append(rules, rule)
	}

	// Create config file detection rules
	if len(pattern.ConfigFiles) > 0 {
		ruleID := fmt.Sprintf("zero.%s.config", baseID)

		// Determine tool type from category
		toolType := categoryToToolType(pattern.Category)

		rule := SemgrepRule{
			ID:        ruleID,
			Message:   fmt.Sprintf("%s configuration file detected", pattern.Name),
			Severity:  "INFO",
			Languages: []string{"generic"},
			Metadata: map[string]interface{}{
				"technology":     pattern.Name,
				"category":       pattern.Category,
				"tool_type":      toolType,
				"detection_type": "config_file",
				"confidence":     getConfidence(pattern, "config"),
				"config_files":   pattern.ConfigFiles,
			},
		}

		// Add homepage if available
		if pattern.Homepage != "" {
			rule.Metadata["homepage"] = pattern.Homepage
		}

		// Build pattern regex to match any of the config filenames
		// This creates a pattern that matches lines containing these filenames
		var filePatterns []string
		for _, cf := range pattern.ConfigFiles {
			// Escape special regex chars in filename
			escaped := regexp.QuoteMeta(cf)
			filePatterns = append(filePatterns, escaped)
		}

		if len(filePatterns) == 1 {
			rule.PatternRegex = filePatterns[0]
		} else {
			rule.PatternRegex = "(" + strings.Join(filePatterns, "|") + ")"
		}

		rules = append(rules, rule)
	}

	// Create security pattern rules (used in devops files like docker.md)
	for i, secPat := range pattern.SecurityPatterns {
		if secPat.Pattern == "" {
			continue
		}

		ruleID := fmt.Sprintf("zero.%s.security.%d", baseID, i+1)

		// Determine language based on category
		language := "generic"
		if strings.Contains(pattern.Category, "docker") {
			language = "dockerfile"
		} else if strings.Contains(pattern.Category, "terraform") {
			language = "hcl"
		} else if strings.Contains(pattern.Category, "yaml") || strings.Contains(pattern.Category, "kubernetes") {
			language = "yaml"
		}

		// Map severity and get additional metadata
		mappedSev, sevMeta := mapSeverityWithMetadata(secPat.Severity)

		// Build metadata with original severity preserved
		metadata := map[string]interface{}{
			"technology":     pattern.Name,
			"category":       pattern.Category,
			"detection_type": "security",
			"pattern_name":   secPat.Name,
			"confidence":     90,
		}
		// Merge severity metadata
		for k, v := range sevMeta {
			metadata[k] = v
		}

		rule := SemgrepRule{
			ID:           ruleID,
			Message:      fmt.Sprintf("%s: %s", pattern.Name, secPat.Name),
			Severity:     mappedSev,
			Languages:    []string{language},
			Metadata:     metadata,
			PatternRegex: secPat.Pattern,
		}
		rules = append(rules, rule)
	}

	return rules
}

// categoryToToolType maps RAG categories to tool/technology types for devex
func categoryToToolType(category string) string {
	// Tool categories (dev tools - configuration burden)
	toolCategories := map[string]string{
		"developer-tools/linting":        "linter",
		"developer-tools/formatting":     "formatter",
		"developer-tools/bundlers":       "bundler",
		"developer-tools/testing":        "test",
		"developer-tools/cicd":           "ci-cd",
		"developer-tools/build":          "build",
		"developer-tools/containers":     "container",
		"developer-tools/infrastructure": "iac",
		"developer-tools/monitoring":     "monitoring",
	}

	// Technology categories (learning curve)
	techCategories := map[string]string{
		"languages":           "language",
		"web-frameworks":      "framework",
		"databases":           "database",
		"cloud":               "cloud",
		"authentication":      "auth",
		"ai-ml":               "ai-ml",
		"cryptographic-libraries": "crypto",
	}

	// Check for tool category match
	for prefix, toolType := range toolCategories {
		if strings.HasPrefix(category, prefix) {
			return toolType
		}
	}

	// Check for technology category match
	for prefix, techType := range techCategories {
		if strings.HasPrefix(category, prefix) {
			return techType
		}
	}

	// Default: try to extract from category path
	parts := strings.Split(category, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "unknown"
}

// pathToID converts a file path to a rule ID component
func pathToID(path string) string {
	// Remove the filename (e.g., weak-ciphers.md)
	dir := filepath.Dir(path)
	// Remove technology-identification prefix if present
	dir = strings.TrimPrefix(dir, "technology-identification/")
	dir = strings.TrimPrefix(dir, "technology-identification\\")
	// Replace path separators with dots
	id := strings.ReplaceAll(dir, "/", ".")
	id = strings.ReplaceAll(id, "\\", ".")
	// Sanitize
	id = regexp.MustCompile(`[^a-z0-9.-]`).ReplaceAllString(strings.ToLower(id), "-")
	return id
}

// sanitizeID sanitizes a string for use in a rule ID
func sanitizeID(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// mapLanguage maps RAG language names to semgrep language names
func mapLanguage(lang string) string {
	mapping := map[string]string{
		"python":     "python",
		"javascript": "javascript",
		"typescript": "typescript",
		"go":         "go",
		"ruby":       "ruby",
		"java":       "java",
		"php":        "php",
		"csharp":     "csharp",
		"c#":         "csharp",
		"rust":       "rust",
		"c":          "c",
		"c++":        "cpp",
		"cpp":        "cpp",
		"c/c++":      "c", // Use C for combined patterns
		"kotlin":     "kotlin",
		"scala":      "scala",
		"swift":      "swift",
		"bash":       "bash",
		"shell":      "bash",
		"sh":         "bash",
		"powershell": "generic", // Semgrep doesn't have native powershell, use generic
		"ps1":        "generic",
		"lua":        "lua",
		"r":          "r",
		"elixir":     "elixir",
		"ocaml":      "ocaml",
		"hcl":        "hcl",         // Terraform
		"terraform":  "hcl",
		"dockerfile": "dockerfile",
		"docker":     "dockerfile",
		"yaml":       "yaml",
		"json":       "json",
		"xml":        "xml",
		"html":       "html",
		"generic":    "generic",
	}
	return mapping[lang]
}

// mapSeverity maps RAG severity to semgrep severity
// Returns both the mapped severity and a flag indicating if it was critical
func mapSeverity(sev string) string {
	mapping := map[string]string{
		"critical": "ERROR",
		"high":     "WARNING",
		"medium":   "WARNING",
		"low":      "INFO",
		"info":     "INFO",
	}
	if mapped, ok := mapping[strings.ToLower(sev)]; ok {
		return mapped
	}
	return "INFO"
}

// mapSeverityWithMetadata maps severity and returns additional metadata
func mapSeverityWithMetadata(sev string) (string, map[string]interface{}) {
	meta := make(map[string]interface{})
	normalizedSev := strings.ToLower(sev)

	// Preserve original severity in metadata
	if normalizedSev != "" {
		meta["original_severity"] = normalizedSev
	}

	// Flag critical severity for easy filtering
	if normalizedSev == "critical" {
		meta["is_critical"] = true
	}

	return mapSeverity(sev), meta
}

// getConfidence returns the confidence score for a detection type
func getConfidence(pattern *RAGPattern, detType string) int {
	key := strings.ReplaceAll(detType, " ", "_") + "_detection"
	if conf, ok := pattern.Confidence[key]; ok {
		return conf
	}
	return 85 // Default
}

// cleanRegexForSemgrep cleans up a regex pattern for use with semgrep pattern-regex
func cleanRegexForSemgrep(regex string) string {
	if regex == "" {
		return ""
	}

	pattern := regex

	// Remove start-of-line anchor (semgrep handles this)
	pattern = strings.TrimPrefix(pattern, "^")

	// Remove end-of-line anchor
	pattern = strings.TrimSuffix(pattern, "$")

	// Skip patterns that are just word boundaries or too simple (overly generic)
	if pattern == `\b` || pattern == `\s+` || pattern == `.*` || pattern == `.+` || pattern == `\w+` {
		return ""
	}

	// Validate regex compiles - this catches truly invalid patterns
	// while allowing valid short patterns like "api", "jwt", etc.
	if _, err := regexp.Compile(pattern); err != nil {
		return ""
	}

	return pattern
}

// regexToSemgrep converts a regex pattern to semgrep pattern syntax
func regexToSemgrep(regex, language string) string {
	pattern := regex

	switch language {
	case "python":
		// `^import openai$` -> `import openai`
		if strings.HasPrefix(pattern, "^import ") {
			result := strings.TrimPrefix(pattern, "^")
			result = strings.TrimSuffix(result, "$")
			return result
		}
		// `^from openai import` -> `from openai import $X`
		if strings.HasPrefix(pattern, "^from ") && strings.Contains(pattern, " import") {
			re := regexp.MustCompile(`\^from\s+(\S+)\s+import`)
			if m := re.FindStringSubmatch(pattern); m != nil {
				return fmt.Sprintf("from %s import $X", m[1])
			}
		}
		// `import\s+psycopg2` -> `import psycopg2`
		if strings.HasPrefix(pattern, "import\\s+") {
			module := strings.TrimPrefix(pattern, "import\\s+")
			// Unescape dots for semgrep
			module = strings.ReplaceAll(module, `\.`, ".")
			return fmt.Sprintf("import %s", module)
		}
		// `from\s+psycopg2` -> `from psycopg2 import $X`
		if strings.HasPrefix(pattern, "from\\s+") {
			module := strings.TrimPrefix(pattern, "from\\s+")
			// Unescape dots for semgrep
			module = strings.ReplaceAll(module, `\.`, ".")
			// Skip patterns that just have "from X" without a complete module
			if strings.TrimSpace(module) == "" || strings.HasSuffix(module, ".") {
				return ""
			}
			return fmt.Sprintf("from %s import $X", module)
		}

	case "javascript", "typescript":
		// `from\s+['"]pg['"]` -> `import $X from "pg"`
		fromRe := regexp.MustCompile(`from\s*\['\"\]([^\[]+)\['\"\]`)
		if m := fromRe.FindStringSubmatch(pattern); m != nil {
			return fmt.Sprintf(`import $X from "%s"`, m[1])
		}
		// `require\(['"]pg['"]\)` -> `require("pg")`
		requireRe := regexp.MustCompile(`require\s*\\\(\s*\['\"\]([^\[]+)\['\"\]\s*\\\)`)
		if m := requireRe.FindStringSubmatch(pattern); m != nil {
			return fmt.Sprintf(`require("%s")`, m[1])
		}
		// Simple require pattern
		if strings.Contains(pattern, "require") {
			simpleReqRe := regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`)
			if m := simpleReqRe.FindStringSubmatch(pattern); m != nil {
				return fmt.Sprintf(`require("%s")`, m[1])
			}
		}

	case "go":
		// `"github\.com/lib/pq"` -> `import "github.com/lib/pq"`
		if strings.HasPrefix(pattern, `"`) && strings.HasSuffix(pattern, `"`) {
			// Unescape dots and return the full quoted string
			unescaped := strings.ReplaceAll(pattern, `\.`, ".")
			return fmt.Sprintf("import %s", unescaped)
		}
		// Patterns without closing quote - skip
		if strings.HasPrefix(pattern, `"`) && !strings.HasSuffix(pattern, `"`) {
			return "" // Invalid pattern
		}

	case "ruby":
		// `require\s+['"]pg['"]` -> `require "pg"`
		if strings.HasPrefix(pattern, "require") {
			rubyRe := regexp.MustCompile(`require\s*\['\"\]([^\[]+)\['\"\]`)
			if m := rubyRe.FindStringSubmatch(pattern); m != nil {
				return fmt.Sprintf(`require "%s"`, m[1])
			}
		}
	}

	return ""
}

// writeYAML writes rules to a YAML file
func writeYAML(path string, rules SemgrepRules) error {
	data, err := yaml.Marshal(rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// discoverRAGDirectories dynamically discovers all RAG subdirectories
func discoverRAGDirectories(ragDir string) ([]string, error) {
	entries, err := os.ReadDir(ragDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read RAG directory: %w", err)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}
