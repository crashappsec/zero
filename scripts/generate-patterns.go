// +build ignore

// This script parses RAG markdown files and generates patterns.json for native Go detection.
// Run with: go run scripts/generate-patterns.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Output structures
type PatternDatabase struct {
	Version      string                `json:"version"`
	Description  string                `json:"description"`
	Technologies []TechnologyPattern   `json:"technologies"`
}

type TechnologyPattern struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Category    string           `json:"category"`
	Vendor      string           `json:"vendor,omitempty"`
	Homepage    string           `json:"homepage,omitempty"`
	Description string           `json:"description,omitempty"`
	Detections  DetectionConfig  `json:"detections"`
	Security    *SecurityConfig  `json:"security,omitempty"`
}

type DetectionConfig struct {
	Packages    []PackagePattern   `json:"packages,omitempty"`
	Imports     []ImportPattern    `json:"imports,omitempty"`
	ConfigFiles []ConfigPattern    `json:"config_files,omitempty"`
	CodePatterns []CodePattern     `json:"code_patterns,omitempty"`
	Extensions  []string           `json:"extensions,omitempty"`
	Directories []string           `json:"directories,omitempty"`
}

type PackagePattern struct {
	Ecosystem  string `json:"ecosystem"`
	Name       string `json:"name"`
	Confidence int    `json:"confidence"`
}

type ImportPattern struct {
	Language   string `json:"language"`
	Pattern    string `json:"pattern"`
	Confidence int    `json:"confidence"`
}

type ConfigPattern struct {
	Path       string `json:"path"`
	Confidence int    `json:"confidence"`
}

type CodePattern struct {
	Pattern     string `json:"pattern"`
	Description string `json:"description,omitempty"`
	Confidence  int    `json:"confidence"`
}

type SecurityConfig struct {
	Secrets []SecretPattern `json:"secrets,omitempty"`
}

type SecretPattern struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Severity    string `json:"severity"`
	Description string `json:"description,omitempty"`
}

func main() {
	ragPath := findRAGPath()
	if ragPath == "" {
		fmt.Println("Error: Could not find RAG directory")
		os.Exit(1)
	}

	techIDPath := filepath.Join(ragPath, "technology-identification")
	fmt.Printf("Scanning: %s\n", techIDPath)

	db := PatternDatabase{
		Version:     "2.0.0",
		Description: "Auto-generated from RAG markdown files",
	}

	// Walk all markdown files
	err := filepath.Walk(techIDPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		if filepath.Base(path) == "README.md" || filepath.Base(path) == "confidence-config.md" {
			return nil // Skip index files
		}

		tech, err := parseMarkdownFile(path, techIDPath)
		if err != nil {
			fmt.Printf("Warning: parsing %s: %v\n", path, err)
			return nil
		}
		if tech != nil && tech.ID != "" {
			db.Technologies = append(db.Technologies, *tech)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed %d technologies\n", len(db.Technologies))

	// Write output
	outputPath := filepath.Join(filepath.Dir(ragPath), "pkg/scanner/technology-identification/patterns.json")
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Written to: %s\n", outputPath)
}

func parseMarkdownFile(path, basePath string) (*TechnologyPattern, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	tech := &TechnologyPattern{
		Detections: DetectionConfig{},
	}

	// Extract ID from directory structure
	relPath, _ := filepath.Rel(basePath, path)
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) >= 2 {
		tech.Category = parts[0]
		tech.ID = strings.TrimSuffix(parts[len(parts)-1], ".md")
		if len(parts) >= 3 {
			tech.ID = parts[len(parts)-2] // Use directory name as ID
		}
	}

	// Parse header metadata
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			tech.Name = strings.TrimPrefix(line, "# ")
		} else if strings.HasPrefix(line, "**Category**:") {
			tech.Category = strings.TrimSpace(strings.TrimPrefix(line, "**Category**:"))
		} else if strings.HasPrefix(line, "**Description**:") {
			tech.Description = strings.TrimSpace(strings.TrimPrefix(line, "**Description**:"))
		} else if strings.HasPrefix(line, "**Homepage**:") {
			tech.Homepage = strings.TrimSpace(strings.TrimPrefix(line, "**Homepage**:"))
		} else if strings.HasPrefix(line, "**Vendor**:") {
			tech.Vendor = strings.TrimSpace(strings.TrimPrefix(line, "**Vendor**:"))
		} else if strings.HasPrefix(line, "| Maintained By |") {
			// Parse table row: | Maintained By | Vendor |
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				tech.Vendor = strings.TrimSpace(parts[2])
			}
		} else if strings.HasPrefix(line, "| Website |") {
			// Parse table row: | Website | URL |
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				tech.Homepage = strings.TrimSpace(parts[2])
			}
		}
	}

	// Parse sections
	parsePackages(content, tech)
	parseImports(content, tech)
	parseConfigFiles(content, tech)
	parseCodePatterns(content, tech)
	parseSecrets(content, tech)
	parseDirectories(content, tech)

	// Skip if no useful detection info
	if len(tech.Detections.Packages) == 0 &&
		len(tech.Detections.Imports) == 0 &&
		len(tech.Detections.ConfigFiles) == 0 &&
		len(tech.Detections.CodePatterns) == 0 {
		return nil, nil
	}

	return tech, nil
}

func parsePackages(content string, tech *TechnologyPattern) {
	ecosystemMap := map[string]string{
		"NPM":      "npm",
		"PYPI":     "pypi",
		"GO":       "go",
		"MAVEN":    "maven",
		"NUGET":    "nuget",
		"RUBYGEMS": "rubygems",
		"CARGO":    "cargo",
		"PACKAGIST": "packagist",
		"COMPOSER": "packagist",
	}

	// Find package sections - supports multiple formats:
	// 1. "## NPM" or "### NPM" or "#### NPM"
	// 2. "#### NPM (JavaScript/TypeScript)" - with language hint in parens
	// Case-insensitive to match "PyPI", "npm", etc.
	ecosystemRe := regexp.MustCompile(`(?mi)^#{2,4}\s*(NPM|PYPI|GO|MAVEN|NUGET|RUBYGEMS|CARGO|PACKAGIST|COMPOSER)(?:\s*\([^)]*\))?\s*$`)
	packageRe := regexp.MustCompile(`(?m)^-\s+` + "`" + `([^` + "`" + `]+)` + "`")
	packageAltRe := regexp.MustCompile(`(?m)^-\s+([a-zA-Z0-9@/_.-]+)(?:\s|$)`)
	codeBlockRe := regexp.MustCompile("(?s)```\\s*\n(.*?)```")
	// Pattern to detect next section header (any ## or ###)
	nextSectionRe := regexp.MustCompile(`(?m)^#{2,3}\s+[^#]`)

	matches := ecosystemRe.FindAllStringSubmatchIndex(content, -1)
	for i, match := range matches {
		if len(match) < 4 {
			continue
		}
		ecosystem := strings.ToUpper(content[match[2]:match[3]])
		ecosystemKey, ok := ecosystemMap[ecosystem]
		if !ok {
			continue
		}

		// Find section end - either next ecosystem header, or next major section
		sectionEnd := len(content)
		if i+1 < len(matches) {
			sectionEnd = matches[i+1][0]
		}
		// Also check for next major section header (### or ##)
		sectionAfter := content[match[1]:]
		nextSecMatch := nextSectionRe.FindStringIndex(sectionAfter[4:]) // skip the current header
		if nextSecMatch != nil && match[1]+4+nextSecMatch[0] < sectionEnd {
			sectionEnd = match[1] + 4 + nextSecMatch[0]
		}

		section := content[match[1]:sectionEnd]

		// First, try to find packages in code blocks (```...```)
		// Only use the FIRST code block in the section (package list)
		codeBlockMatch := codeBlockRe.FindStringSubmatch(section)
		if codeBlockMatch != nil && len(codeBlockMatch) >= 2 {
			// Each line in the code block is a package
			lines := strings.Split(codeBlockMatch[1], "\n")
			for _, line := range lines {
				pkg := strings.TrimSpace(line)
				// Skip empty lines, comments, and lines with spaces (code examples)
				// Also validate it looks like a package name
				if isValidPackageName(pkg) {
					tech.Detections.Packages = append(tech.Detections.Packages, PackagePattern{
						Ecosystem:  ecosystemKey,
						Name:       pkg,
						Confidence: 95,
					})
				}
			}
		}

		// Also look for bullet point package lists
		pkgMatches := packageRe.FindAllStringSubmatch(section, -1)
		for _, m := range pkgMatches {
			if len(m) >= 2 {
				pkg := strings.TrimSpace(m[1])
				if isValidPackageName(pkg) {
					tech.Detections.Packages = append(tech.Detections.Packages, PackagePattern{
						Ecosystem:  ecosystemKey,
						Name:       pkg,
						Confidence: 95,
					})
				}
			}
		}

		// Try alternate bullet format if no packages found yet
		if codeBlockMatch == nil && len(pkgMatches) == 0 {
			altMatches := packageAltRe.FindAllStringSubmatch(section, -1)
			for _, m := range altMatches {
				if len(m) >= 2 {
					pkg := strings.TrimSpace(m[1])
					if isValidPackageName(pkg) {
						tech.Detections.Packages = append(tech.Detections.Packages, PackagePattern{
							Ecosystem:  ecosystemKey,
							Name:       pkg,
							Confidence: 90,
						})
					}
				}
			}
		}
	}
}

// isValidPackageName checks if a string looks like a valid package name
func isValidPackageName(pkg string) bool {
	if pkg == "" || len(pkg) < 2 {
		return false
	}
	// Skip obvious non-packages
	if strings.HasPrefix(pkg, "#") || strings.HasPrefix(pkg, "//") || strings.HasPrefix(pkg, "-") {
		return false
	}
	if strings.Contains(pkg, " ") || strings.Contains(pkg, "=") || strings.Contains(pkg, "(") || strings.Contains(pkg, ")") {
		return false
	}
	// Skip single punctuation
	if pkg == "{" || pkg == "}" || pkg == "[" || pkg == "]" {
		return false
	}
	// Skip things that look like code, not packages
	if strings.HasPrefix(pkg, "@") && strings.Contains(pkg, ".") && !strings.Contains(pkg, "/") {
		return false // Looks like decorator, not package
	}
	// Must start with letter, @ (scoped package), or valid package char
	first := pkg[0]
	if first != '@' && first != '_' && (first < 'a' || first > 'z') && (first < 'A' || first > 'Z') {
		return false
	}
	return true
}

func parseImports(content string, tech *TechnologyPattern) {
	// Look for Import Patterns section
	importSection := extractSection(content, "Import Patterns", "Import Detection")
	if importSection == "" {
		return
	}

	// Parse language-specific patterns
	langMap := map[string]string{
		"Go":         "go",
		"Python":     "python",
		"JavaScript": "javascript",
		"TypeScript": "javascript",
		"Ruby":       "ruby",
		"Java":       "java",
		"Rust":       "rust",
		"C#":         "csharp",
		"Javascript": "javascript",
	}

	patternRe := regexp.MustCompile(`\*\*Pattern\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	// Match code blocks with language hint: ```javascript or ```python
	codeBlockRe := regexp.MustCompile("(?s)```(javascript|typescript|python|go|java|ruby|rust)\\s*\n(.*?)```")

	for lang, langKey := range langMap {
		langSection := extractLanguageSection(importSection, lang)
		if langSection == "" {
			continue
		}

		// First try explicit **Pattern**: format
		matches := patternRe.FindAllStringSubmatch(langSection, -1)
		for _, m := range matches {
			if len(m) >= 2 {
				pattern := m[1]

				// Skip patterns with Go RE2 unsupported features
				if strings.Contains(pattern, "(?<") || strings.Contains(pattern, "(?!") || strings.Contains(pattern, "(?=") {
					continue
				}

				// Validate regex compiles in Go
				if _, err := regexp.Compile(pattern); err != nil {
					continue
				}

				tech.Detections.Imports = append(tech.Detections.Imports, ImportPattern{
					Language:   langKey,
					Pattern:    pattern,
					Confidence: 90,
				})
			}
		}

		// Also try to extract from code blocks
		codeMatches := codeBlockRe.FindAllStringSubmatch(langSection, -1)
		for _, m := range codeMatches {
			if len(m) >= 3 {
				// Check language matches
				codeLang := strings.ToLower(m[1])
				if codeLang == "typescript" {
					codeLang = "javascript"
				}
				if codeLang != langKey {
					continue
				}

				// Extract import lines from code block
				lines := strings.Split(m[2], "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					// Convert import examples to regex patterns
					pattern := importLineToPattern(line, langKey)
					if pattern != "" {
						tech.Detections.Imports = append(tech.Detections.Imports, ImportPattern{
							Language:   langKey,
							Pattern:    pattern,
							Confidence: 85,
						})
					}
				}
			}
		}
	}
}

// importLineToPattern converts example import lines to regex patterns
func importLineToPattern(line, lang string) string {
	// Skip comments
	if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
		return ""
	}

	switch lang {
	case "javascript":
		// Handle: import { X } from "package"
		// Handle: from "package"
		if strings.Contains(line, "import") || strings.HasPrefix(line, "from") {
			// Extract package name from quotes
			pkgRe := regexp.MustCompile(`["']([^"']+)["']`)
			match := pkgRe.FindStringSubmatch(line)
			if len(match) >= 2 {
				pkg := match[1]
				// Escape special chars for regex
				pkg = regexp.QuoteMeta(pkg)
				return pkg
			}
		}
	case "python":
		// Handle: from package import X
		// Handle: import package
		if strings.HasPrefix(line, "from ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pkg := parts[1]
				return regexp.QuoteMeta(pkg)
			}
		} else if strings.HasPrefix(line, "import ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pkg := strings.TrimSuffix(parts[1], ",")
				return regexp.QuoteMeta(pkg)
			}
		}
	}

	return ""
}

func parseConfigFiles(content string, tech *TechnologyPattern) {
	// Look for Configuration Files section - support multiple section names
	configSection := extractSection(content, "Configuration Files", "Config Files")
	if configSection == "" {
		configSection = extractSection(content, "File Patterns", "")
	}
	if configSection == "" {
		return
	}

	// Extract file patterns from bullet points
	fileRe := regexp.MustCompile(`(?m)^-\s+` + "`" + `([^` + "`" + `]+)` + "`")
	matches := fileRe.FindAllStringSubmatch(configSection, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			file := strings.TrimSpace(m[1])
			if file != "" && !strings.Contains(file, " ") {
				tech.Detections.ConfigFiles = append(tech.Detections.ConfigFiles, ConfigPattern{
					Path:       file,
					Confidence: 95,
				})
			}
		}
	}

	// Also extract from code blocks
	codeBlockRe := regexp.MustCompile("(?s)```\\s*\n(.*?)```")
	codeMatches := codeBlockRe.FindAllStringSubmatch(configSection, -1)
	for _, m := range codeMatches {
		if len(m) >= 2 {
			lines := strings.Split(m[1], "\n")
			for _, line := range lines {
				file := strings.TrimSpace(line)
				// Skip empty lines, comments, and lines with spaces
				if file != "" && !strings.HasPrefix(file, "#") && !strings.Contains(file, " ") && !strings.Contains(file, "=") && !strings.Contains(file, ":") {
					tech.Detections.ConfigFiles = append(tech.Detections.ConfigFiles, ConfigPattern{
						Path:       file,
						Confidence: 90,
					})
				}
			}
		}
	}
}

func parseCodePatterns(content string, tech *TechnologyPattern) {
	// Look for Code Patterns section
	codeSection := extractSection(content, "Code Patterns", "")
	if codeSection == "" {
		return
	}

	patternRe := regexp.MustCompile(`\*\*Pattern\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	descRe := regexp.MustCompile(`(?m)^-\s+(.+)$`)

	matches := patternRe.FindAllStringSubmatchIndex(codeSection, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		pattern := codeSection[match[2]:match[3]]

		// Skip patterns with Go RE2 unsupported features
		if strings.Contains(pattern, "(?<") || strings.Contains(pattern, "(?!") || strings.Contains(pattern, "(?=") {
			continue
		}

		// Validate regex compiles in Go
		if _, err := regexp.Compile(pattern); err != nil {
			continue
		}

		// Find description (next line starting with -)
		afterPattern := codeSection[match[3]:]
		descMatch := descRe.FindStringSubmatch(afterPattern)
		desc := ""
		if len(descMatch) >= 2 {
			desc = descMatch[1]
		}

		tech.Detections.CodePatterns = append(tech.Detections.CodePatterns, CodePattern{
			Pattern:     pattern,
			Description: desc,
			Confidence:  85,
		})
	}
}

func parseSecrets(content string, tech *TechnologyPattern) {
	// Look for Secrets Detection section
	secretsSection := extractSection(content, "Secrets Detection", "Credentials")
	if secretsSection == "" {
		return
	}

	// Parse secret patterns
	nameRe := regexp.MustCompile(`(?m)^#{3,4}\s+(.+)$`)
	patternRe := regexp.MustCompile(`\*\*Pattern\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	severityRe := regexp.MustCompile(`\*\*Severity\*\*:\s*(\w+)`)
	descRe := regexp.MustCompile(`\*\*Description\*\*:\s*(.+)`)

	// Split by headers
	parts := strings.Split(secretsSection, "\n#### ")
	if len(parts) <= 1 {
		parts = strings.Split(secretsSection, "\n### ")
	}

	for _, part := range parts[1:] {
		nameMatch := nameRe.FindStringSubmatch("### " + part)
		patternMatch := patternRe.FindStringSubmatch(part)
		severityMatch := severityRe.FindStringSubmatch(part)
		descMatch := descRe.FindStringSubmatch(part)

		if len(patternMatch) >= 2 {
			pattern := patternMatch[1]

			// Skip patterns with Go RE2 unsupported features (lookahead/lookbehind)
			if strings.Contains(pattern, "(?<") || strings.Contains(pattern, "(?!") || strings.Contains(pattern, "(?=") {
				continue
			}

			// Validate regex compiles in Go
			if _, err := regexp.Compile(pattern); err != nil {
				continue
			}

			secret := SecretPattern{
				Pattern:  pattern,
				Severity: "high",
			}
			if len(nameMatch) >= 2 {
				secret.Name = strings.TrimSpace(nameMatch[1])
			}
			if len(severityMatch) >= 2 {
				secret.Severity = strings.ToLower(severityMatch[1])
			}
			if len(descMatch) >= 2 {
				secret.Description = descMatch[1]
			}

			if tech.Security == nil {
				tech.Security = &SecurityConfig{}
			}
			tech.Security.Secrets = append(tech.Security.Secrets, secret)
		}
	}
}

func parseDirectories(content string, tech *TechnologyPattern) {
	// Look for Configuration Directories section
	dirSection := extractSection(content, "Configuration Directories", "")
	if dirSection == "" {
		return
	}

	dirRe := regexp.MustCompile(`(?m)^-\s+` + "`" + `([^` + "`" + `]+)` + "`")
	matches := dirRe.FindAllStringSubmatch(dirSection, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			dir := strings.TrimSuffix(strings.TrimSpace(m[1]), "/")
			if dir != "" {
				tech.Detections.Directories = append(tech.Detections.Directories, dir)
			}
		}
	}
}

func extractSection(content, header1, header2 string) string {
	// Try first header
	idx := strings.Index(content, "## "+header1)
	if idx == -1 && header2 != "" {
		idx = strings.Index(content, "## "+header2)
	}
	if idx == -1 {
		idx = strings.Index(content, "### "+header1)
	}
	if idx == -1 && header2 != "" {
		idx = strings.Index(content, "### "+header2)
	}
	if idx == -1 {
		return ""
	}

	section := content[idx:]

	// Find end of section (next ## or ---)
	endRe := regexp.MustCompile(`(?m)^(##[^#]|---)`)
	endMatch := endRe.FindStringIndex(section[3:])
	if endMatch != nil {
		section = section[:endMatch[0]+3]
	}

	return section
}

func extractLanguageSection(content, lang string) string {
	patterns := []string{
		"#### " + lang,
		"### " + lang,
		"**" + lang + "**",
	}

	for _, p := range patterns {
		idx := strings.Index(content, p)
		if idx == -1 {
			continue
		}

		section := content[idx:]
		// Find end
		endRe := regexp.MustCompile(`(?m)^(#{2,4}\s|---)`)
		endMatch := endRe.FindStringIndex(section[len(p):])
		if endMatch != nil {
			section = section[:endMatch[0]+len(p)]
		}
		return section
	}

	return ""
}

func findRAGPath() string {
	candidates := []string{
		"rag",
		"../rag",
		"../../rag",
	}

	if zeroHome := os.Getenv("ZERO_HOME"); zeroHome != "" {
		candidates = append([]string{filepath.Join(zeroHome, "rag")}, candidates...)
	}

	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}

	return ""
}
