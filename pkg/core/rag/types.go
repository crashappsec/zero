// Package rag provides RAG (Retrieval-Augmented Generation) pattern loading utilities
package rag

// PatternSet represents a collection of patterns from RAG
type PatternSet struct {
	Category   string    `json:"category"`
	Technology string    `json:"technology"`
	Source     string    `json:"source"`
	Patterns   []Pattern `json:"patterns"`
}

// Pattern represents a single detection pattern
type Pattern struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"` // semgrep, regex, glob
	Pattern    string         `json:"pattern"`
	Message    string         `json:"message"`
	Severity   string         `json:"severity"`
	Confidence int            `json:"confidence"`
	Languages  []string       `json:"languages"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// RuleSet represents generated Semgrep rules
type RuleSet struct {
	Name       string `json:"name"`
	OutputFile string `json:"output_file"`
	RuleCount  int    `json:"rule_count"`
	Version    string `json:"version"`
}

// SemgrepRule represents a Semgrep rule structure
type SemgrepRule struct {
	ID        string            `yaml:"id" json:"id"`
	Message   string            `yaml:"message" json:"message"`
	Severity  string            `yaml:"severity" json:"severity"`
	Languages []string          `yaml:"languages" json:"languages"`
	Pattern   string            `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Patterns  []SemgrepPattern  `yaml:"patterns,omitempty" json:"patterns,omitempty"`
	Metadata  map[string]any    `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// SemgrepPattern represents a pattern within a rule
type SemgrepPattern struct {
	Pattern        string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	PatternEither  []any  `yaml:"pattern-either,omitempty" json:"pattern-either,omitempty"`
	PatternNot     string `yaml:"pattern-not,omitempty" json:"pattern-not,omitempty"`
	PatternInside  string `yaml:"pattern-inside,omitempty" json:"pattern-inside,omitempty"`
	PatternRegex   string `yaml:"pattern-regex,omitempty" json:"pattern-regex,omitempty"`
	MetavarPattern any    `yaml:"metavariable-pattern,omitempty" json:"metavariable-pattern,omitempty"`
}

// SemgrepConfig represents a complete Semgrep config file
type SemgrepConfig struct {
	Rules []SemgrepRule `yaml:"rules" json:"rules"`
}

// RAGCategory represents known RAG categories
type RAGCategory string

const (
	CategoryTechID       RAGCategory = "technology-identification"
	CategorySecrets      RAGCategory = "secrets"
	CategoryVulns        RAGCategory = "vulnerabilities"
	CategoryCrypto       RAGCategory = "crypto"
	CategoryAISecurity   RAGCategory = "ai-security"
	CategoryCompliance   RAGCategory = "compliance"
	CategoryDevOps       RAGCategory = "devops"        // docker, iac, cicd
	CategoryDevOpsSec    RAGCategory = "devops-security" // IaC policies, secrets-in-iac
	CategoryCodeSecurity RAGCategory = "code-security" // api-quality, api-versioning
	CategoryCodeQuality  RAGCategory = "code-quality"  // patterns, style
	CategoryArchitecture RAGCategory = "architecture"  // microservices
)

// String returns the category as a string
func (c RAGCategory) String() string {
	return string(c)
}

// PatternType represents types of patterns
type PatternType string

const (
	PatternTypeSemgrep PatternType = "semgrep"
	PatternTypeRegex   PatternType = "regex"
	PatternTypeGlob    PatternType = "glob"
)

// LoadResult represents the result of loading RAG patterns
type LoadResult struct {
	Category     string       `json:"category"`
	PatternSets  []PatternSet `json:"pattern_sets"`
	TotalPatterns int         `json:"total_patterns"`
	Errors       []string     `json:"errors,omitempty"`
}

// ConversionResult represents the result of converting RAG to Semgrep
type ConversionResult struct {
	RuleSets     []RuleSet `json:"rule_sets"`
	TotalRules   int       `json:"total_rules"`
	OutputDir    string    `json:"output_dir"`
	Errors       []string  `json:"errors,omitempty"`
}
