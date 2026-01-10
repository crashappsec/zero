// Package techid provides the consolidated technology identification scanner
// This file loads JSON patterns for native Go pattern matching
package techid

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

//go:embed patterns.json
var embeddedPatterns []byte

// PatternDatabase holds all loaded technology patterns
type PatternDatabase struct {
	Version      string                `json:"version"`
	Description  string                `json:"description"`
	Technologies []TechnologyPattern   `json:"technologies"`

	// Compiled patterns for fast matching
	mu               sync.RWMutex
	compiledImports  map[string][]*CompiledImport  // language -> compiled patterns
	compiledCode     map[string][]*CompiledCode    // language -> compiled patterns
	compiledSecrets  []*CompiledSecret
	packageIndex     map[string]map[string][]PackageMatch // ecosystem -> name -> tech matches
	configIndex      map[string][]ConfigMatch             // path -> tech matches
	extensionIndex   map[string][]ExtensionMatch          // extension -> tech matches
}

// TechnologyPattern defines detection patterns for a technology
type TechnologyPattern struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Description string                 `json:"description,omitempty"`
	Homepage    string                 `json:"homepage,omitempty"`
	Vendor      string                 `json:"vendor,omitempty"`
	License     string                 `json:"license,omitempty"`
	RiskLevel   string                 `json:"risk_level,omitempty"`
	Detections  PatternDetectionConfig `json:"detections"`
	Security    PatternSecurity        `json:"security,omitempty"`
}

// PatternDetectionConfig holds all detection methods for a technology
type PatternDetectionConfig struct {
	Packages       []PatternPackage   `json:"packages,omitempty"`
	Imports        []PatternImport    `json:"imports,omitempty"`
	ConfigFiles    []PatternConfig    `json:"config_files,omitempty"`
	FilePatterns   []PatternFile      `json:"file_patterns,omitempty"`
	FileExtensions []PatternExtension `json:"file_extensions,omitempty"`
	EnvVars        []PatternEnvVar    `json:"env_vars,omitempty"`
	CodePatterns   []PatternCode      `json:"code_patterns,omitempty"`
}

// PatternPackage matches packages in SBOM
type PatternPackage struct {
	Ecosystem  string `json:"ecosystem"`           // npm, pypi, go, etc.
	Name       string `json:"name,omitempty"`      // exact name match
	Pattern    string `json:"pattern,omitempty"`   // glob pattern for matching
	Confidence int    `json:"confidence"`
}

// PatternImport matches import statements in code
type PatternImport struct {
	Language   string `json:"language"`
	Pattern    string `json:"pattern"`             // regex pattern
	Confidence int    `json:"confidence"`
}

// PatternConfig matches configuration files
type PatternConfig struct {
	Path           string `json:"path"`
	Confidence     int    `json:"confidence"`
	IsDirectory    bool   `json:"is_directory,omitempty"`
	ContentPattern string `json:"content_pattern,omitempty"` // regex to match in file content
}

// PatternFile matches file name patterns
type PatternFile struct {
	Pattern    string `json:"pattern"`             // glob pattern
	Confidence int    `json:"confidence"`
}

// PatternExtension matches file extensions
type PatternExtension struct {
	Extension   string `json:"extension"`
	Confidence  int    `json:"confidence"`
	Description string `json:"description,omitempty"`
}

// PatternEnvVar matches environment variables
type PatternEnvVar struct {
	Pattern    string `json:"pattern"`             // glob or exact match
	Confidence int    `json:"confidence"`
}

// PatternCode matches code patterns using regex
type PatternCode struct {
	Language   string `json:"language"`
	Pattern    string `json:"pattern"`             // regex pattern
	Confidence int    `json:"confidence"`
}

// PatternSecurity holds security-related patterns
type PatternSecurity struct {
	Secrets              []PatternSecret `json:"secrets,omitempty"`
	AuditPoints          []string        `json:"audit_points,omitempty"`
	CommonVulnerabilities []string       `json:"common_vulnerabilities,omitempty"`
}

// PatternSecret matches secret values
type PatternSecret struct {
	Pattern  string `json:"pattern"`              // regex pattern
	Name     string `json:"name"`
	Severity string `json:"severity"`             // critical, high, medium, low
	ContextRequired bool `json:"context_required,omitempty"` // needs additional context to confirm
}

// Compiled pattern types for fast matching
type CompiledImport struct {
	TechID     string
	TechName   string
	Category   string
	Regex      *regexp.Regexp
	Confidence int
}

type CompiledCode struct {
	TechID     string
	TechName   string
	Category   string
	Regex      *regexp.Regexp
	Confidence int
}

type CompiledSecret struct {
	TechID     string
	TechName   string
	Regex      *regexp.Regexp
	Name       string
	Severity   string
	ContextRequired bool
}

type PackageMatch struct {
	TechID     string
	TechName   string
	Category   string
	Confidence int
	IsGlob     bool
	GlobPattern string
}

type ConfigMatch struct {
	TechID         string
	TechName       string
	Category       string
	Confidence     int
	IsDirectory    bool
	ContentPattern *regexp.Regexp
}

type ExtensionMatch struct {
	TechID      string
	TechName    string
	Category    string
	Confidence  int
	Description string
}

// Global pattern database instance
var (
	globalPatterns *PatternDatabase
	patternsOnce   sync.Once
	patternsErr    error
)

// LoadPatterns loads and compiles the embedded pattern database
func LoadPatterns() (*PatternDatabase, error) {
	patternsOnce.Do(func() {
		globalPatterns, patternsErr = loadPatternsFromBytes(embeddedPatterns)
	})
	return globalPatterns, patternsErr
}

// loadPatternsFromBytes parses JSON and compiles patterns
func loadPatternsFromBytes(data []byte) (*PatternDatabase, error) {
	var db PatternDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("failed to parse patterns: %w", err)
	}

	// Initialize indexes
	db.compiledImports = make(map[string][]*CompiledImport)
	db.compiledCode = make(map[string][]*CompiledCode)
	db.compiledSecrets = make([]*CompiledSecret, 0)
	db.packageIndex = make(map[string]map[string][]PackageMatch)
	db.configIndex = make(map[string][]ConfigMatch)
	db.extensionIndex = make(map[string][]ExtensionMatch)

	// Compile all patterns
	for _, tech := range db.Technologies {
		if err := db.compilePatterns(tech); err != nil {
			return nil, fmt.Errorf("failed to compile patterns for %s: %w", tech.ID, err)
		}
	}

	return &db, nil
}

// compilePatterns compiles regex patterns and builds indexes for a technology
func (db *PatternDatabase) compilePatterns(tech TechnologyPattern) error {
	// Compile import patterns
	for _, imp := range tech.Detections.Imports {
		re, err := regexp.Compile(imp.Pattern)
		if err != nil {
			return fmt.Errorf("invalid import pattern %q: %w", imp.Pattern, err)
		}
		db.compiledImports[imp.Language] = append(db.compiledImports[imp.Language], &CompiledImport{
			TechID:     tech.ID,
			TechName:   tech.Name,
			Category:   tech.Category,
			Regex:      re,
			Confidence: imp.Confidence,
		})
	}

	// Compile code patterns
	for _, code := range tech.Detections.CodePatterns {
		re, err := regexp.Compile(code.Pattern)
		if err != nil {
			return fmt.Errorf("invalid code pattern %q: %w", code.Pattern, err)
		}
		db.compiledCode[code.Language] = append(db.compiledCode[code.Language], &CompiledCode{
			TechID:     tech.ID,
			TechName:   tech.Name,
			Category:   tech.Category,
			Regex:      re,
			Confidence: code.Confidence,
		})
	}

	// Compile secret patterns
	for _, secret := range tech.Security.Secrets {
		re, err := regexp.Compile(secret.Pattern)
		if err != nil {
			return fmt.Errorf("invalid secret pattern %q: %w", secret.Pattern, err)
		}
		db.compiledSecrets = append(db.compiledSecrets, &CompiledSecret{
			TechID:          tech.ID,
			TechName:        tech.Name,
			Regex:           re,
			Name:            secret.Name,
			Severity:        secret.Severity,
			ContextRequired: secret.ContextRequired,
		})
	}

	// Index packages
	for _, pkg := range tech.Detections.Packages {
		if db.packageIndex[pkg.Ecosystem] == nil {
			db.packageIndex[pkg.Ecosystem] = make(map[string][]PackageMatch)
		}

		match := PackageMatch{
			TechID:     tech.ID,
			TechName:   tech.Name,
			Category:   tech.Category,
			Confidence: pkg.Confidence,
		}

		if pkg.Pattern != "" {
			match.IsGlob = true
			match.GlobPattern = pkg.Pattern
			// Store glob patterns under a special key
			db.packageIndex[pkg.Ecosystem]["__globs__"] = append(
				db.packageIndex[pkg.Ecosystem]["__globs__"], match)
		} else {
			db.packageIndex[pkg.Ecosystem][pkg.Name] = append(
				db.packageIndex[pkg.Ecosystem][pkg.Name], match)
		}
	}

	// Index config files
	for _, cfg := range tech.Detections.ConfigFiles {
		var contentRe *regexp.Regexp
		if cfg.ContentPattern != "" {
			var err error
			contentRe, err = regexp.Compile(cfg.ContentPattern)
			if err != nil {
				return fmt.Errorf("invalid content pattern %q: %w", cfg.ContentPattern, err)
			}
		}

		db.configIndex[cfg.Path] = append(db.configIndex[cfg.Path], ConfigMatch{
			TechID:         tech.ID,
			TechName:       tech.Name,
			Category:       tech.Category,
			Confidence:     cfg.Confidence,
			IsDirectory:    cfg.IsDirectory,
			ContentPattern: contentRe,
		})
	}

	// Index file extensions
	for _, ext := range tech.Detections.FileExtensions {
		db.extensionIndex[ext.Extension] = append(db.extensionIndex[ext.Extension], ExtensionMatch{
			TechID:      tech.ID,
			TechName:    tech.Name,
			Category:    tech.Category,
			Confidence:  ext.Confidence,
			Description: ext.Description,
		})
	}

	return nil
}

// MatchPackage checks if a package name matches any technology
func (db *PatternDatabase) MatchPackage(ecosystem, name string) []PackageMatch {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var matches []PackageMatch

	ecosystemIndex := db.packageIndex[ecosystem]
	if ecosystemIndex == nil {
		return matches
	}

	// Check exact matches
	if exactMatches, ok := ecosystemIndex[name]; ok {
		matches = append(matches, exactMatches...)
	}

	// Check glob patterns
	if globs, ok := ecosystemIndex["__globs__"]; ok {
		for _, glob := range globs {
			if matchGlob(glob.GlobPattern, name) {
				matches = append(matches, glob)
			}
		}
	}

	return matches
}

// MatchImport checks if an import statement matches any technology
func (db *PatternDatabase) MatchImport(language, line string) []struct {
	TechID     string
	TechName   string
	Category   string
	Confidence int
} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var matches []struct {
		TechID     string
		TechName   string
		Category   string
		Confidence int
	}

	patterns := db.compiledImports[language]
	for _, p := range patterns {
		if p.Regex.MatchString(line) {
			matches = append(matches, struct {
				TechID     string
				TechName   string
				Category   string
				Confidence int
			}{
				TechID:     p.TechID,
				TechName:   p.TechName,
				Category:   p.Category,
				Confidence: p.Confidence,
			})
		}
	}

	return matches
}

// MatchConfigFile checks if a config file path matches any technology
func (db *PatternDatabase) MatchConfigFile(path string) []ConfigMatch {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var matches []ConfigMatch

	// Check exact path match
	if configMatches, ok := db.configIndex[path]; ok {
		matches = append(matches, configMatches...)
	}

	// Check if path ends with any indexed path
	for indexPath, configMatches := range db.configIndex {
		if strings.HasSuffix(path, "/"+indexPath) || strings.HasSuffix(path, "\\"+indexPath) {
			matches = append(matches, configMatches...)
		}
	}

	return matches
}

// MatchExtension checks if a file extension matches any technology
func (db *PatternDatabase) MatchExtension(ext string) []ExtensionMatch {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.extensionIndex[ext]
}

// MatchSecret checks if content contains any technology-specific secrets
func (db *PatternDatabase) MatchSecret(content string) []struct {
	TechID          string
	TechName        string
	Name            string
	Severity        string
	Match           string
	ContextRequired bool
} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var matches []struct {
		TechID          string
		TechName        string
		Name            string
		Severity        string
		Match           string
		ContextRequired bool
	}

	for _, secret := range db.compiledSecrets {
		if match := secret.Regex.FindString(content); match != "" {
			matches = append(matches, struct {
				TechID          string
				TechName        string
				Name            string
				Severity        string
				Match           string
				ContextRequired bool
			}{
				TechID:          secret.TechID,
				TechName:        secret.TechName,
				Name:            secret.Name,
				Severity:        secret.Severity,
				Match:           match,
				ContextRequired: secret.ContextRequired,
			})
		}
	}

	return matches
}

// GetTechnology returns a technology by ID
func (db *PatternDatabase) GetTechnology(id string) *TechnologyPattern {
	for i := range db.Technologies {
		if db.Technologies[i].ID == id {
			return &db.Technologies[i]
		}
	}
	return nil
}

// GetTechnologiesByCategory returns all technologies in a category
func (db *PatternDatabase) GetTechnologiesByCategory(category string) []TechnologyPattern {
	var techs []TechnologyPattern
	for _, tech := range db.Technologies {
		if strings.HasPrefix(tech.Category, category) {
			techs = append(techs, tech)
		}
	}
	return techs
}

// matchGlob performs simple glob matching (supports * and ?)
func matchGlob(pattern, s string) bool {
	// Convert glob to regex
	regexPattern := "^"
	for _, c := range pattern {
		switch c {
		case '*':
			regexPattern += ".*"
		case '?':
			regexPattern += "."
		case '.', '+', '^', '$', '[', ']', '(', ')', '{', '}', '|', '\\':
			regexPattern += "\\" + string(c)
		default:
			regexPattern += string(c)
		}
	}
	regexPattern += "$"

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// Stats returns statistics about the loaded patterns
func (db *PatternDatabase) Stats() map[string]int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	stats := map[string]int{
		"technologies":    len(db.Technologies),
		"import_patterns": 0,
		"code_patterns":   0,
		"secret_patterns": len(db.compiledSecrets),
		"config_files":    len(db.configIndex),
		"extensions":      len(db.extensionIndex),
	}

	for _, patterns := range db.compiledImports {
		stats["import_patterns"] += len(patterns)
	}
	for _, patterns := range db.compiledCode {
		stats["code_patterns"] += len(patterns)
	}

	return stats
}
