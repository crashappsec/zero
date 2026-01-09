// Package codeownership provides code ownership and CODEOWNERS analysis
package codeownership

// FeatureConfig holds configuration for code ownership analysis
type FeatureConfig struct {
	Enabled             bool `json:"enabled"`
	EnhancedMode        bool `json:"enhanced_mode"`        // Use enhanced v2.0 analysis with weighted scoring
	AnalyzeContributors bool `json:"analyze_contributors"` // Analyze git contributors
	CheckCodeowners     bool `json:"check_codeowners"`     // Validate CODEOWNERS file
	DetectOrphans       bool `json:"detect_orphans"`       // Find files with no recent commits
	AnalyzeCompetency   bool `json:"analyze_competency"`   // Analyze developer competency by language
	DetectLanguages     bool `json:"detect_languages"`     // Detect programming languages in repo
	PeriodDays          int  `json:"period_days"`          // Analysis period (default 90)
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
		Enabled:             true,
		EnhancedMode:        true, // Use enhanced v2.0 analysis by default
		AnalyzeContributors: true,
		CheckCodeowners:     true,
		DetectOrphans:       true,
		AnalyzeCompetency:   true,
		DetectLanguages:     true,
		PeriodDays:          90,
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.AnalyzeContributors = false // Skip contributor analysis (slow)
	cfg.DetectOrphans = false       // Skip orphan detection (slow)
	cfg.AnalyzeCompetency = false   // Skip competency analysis (slow)
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return FeatureConfig{
		Enabled:             true,
		AnalyzeContributors: true,
		CheckCodeowners:     true,
		DetectOrphans:       true,
		AnalyzeCompetency:   true,
		DetectLanguages:     true,
		PeriodDays:          180, // Longer period for thorough analysis
	}
}

// ============================================================================
// Enhanced Ownership Configuration (v2.0)
// ============================================================================

// EnhancedOwnershipConfig configures the enhanced ownership analysis
type EnhancedOwnershipConfig struct {
	// Scoring weights (must sum to 1.0)
	Weights ScoringWeights `json:"weights"`

	// GitHub integration settings
	GitHub GitHubConfig `json:"github"`

	// CODEOWNERS analysis settings
	CODEOWNERS CODEOWNERSConfig `json:"codeowners"`

	// Monorepo detection settings
	Monorepo MonorepoConfig `json:"monorepo"`

	// Incident contact settings
	Contacts ContactsConfig `json:"contacts"`

	// Specialist domains (match Zero agent skills)
	SpecialistDomains []string `json:"specialist_domains"`
}

// ScoringWeights defines weights for ownership scoring components
type ScoringWeights struct {
	Commits     float64 `json:"commits"`     // Default: 0.30
	Reviews     float64 `json:"reviews"`     // Default: 0.25
	Lines       float64 `json:"lines"`       // Default: 0.20
	Recency     float64 `json:"recency"`     // Default: 0.15
	Consistency float64 `json:"consistency"` // Default: 0.10
}

// GitHubConfig configures GitHub API integration
type GitHubConfig struct {
	Enabled        bool `json:"enabled"`          // Enable GitHub API integration
	FetchPRReviews bool `json:"fetch_pr_reviews"` // Fetch PR review data
	ResolveTeams   bool `json:"resolve_teams"`    // Resolve team memberships
	MaxPRs         int  `json:"max_prs"`          // Max PRs to analyze (default: 500)
}

// CODEOWNERSConfig configures CODEOWNERS validation
type CODEOWNERSConfig struct {
	Validate          bool     `json:"validate"`           // Enable validation
	DetectDrift       bool     `json:"detect_drift"`       // Detect ownership drift
	SensitivePatterns []string `json:"sensitive_patterns"` // Patterns for sensitive files
}

// MonorepoConfig configures monorepo detection
type MonorepoConfig struct {
	Enabled    bool `json:"enabled"`     // Enable monorepo detection
	AutoDetect bool `json:"auto_detect"` // Auto-detect monorepo type
}

// ContactsConfig configures incident contact generation
type ContactsConfig struct {
	Enabled    bool `json:"enabled"`     // Enable incident contacts
	MinPrimary int  `json:"min_primary"` // Minimum primary contacts (default: 1)
	MinBackup  int  `json:"min_backup"`  // Minimum backup contacts (default: 1)
}

// DefaultEnhancedConfig returns the default enhanced ownership configuration
func DefaultEnhancedConfig() EnhancedOwnershipConfig {
	return EnhancedOwnershipConfig{
		Weights: ScoringWeights{
			Commits:     0.30,
			Reviews:     0.25,
			Lines:       0.20,
			Recency:     0.15,
			Consistency: 0.10,
		},
		GitHub: GitHubConfig{
			Enabled:        true,
			FetchPRReviews: true,
			ResolveTeams:   true,
			MaxPRs:         500,
		},
		CODEOWNERS: CODEOWNERSConfig{
			Validate:    true,
			DetectDrift: true,
			SensitivePatterns: []string{
				".github/CODEOWNERS",
				".github/workflows/*",
				"CODEOWNERS",
				"auth/*",
				"security/*",
				"crypto/*",
				"*.key",
				"*.pem",
				".env*",
				"secrets/*",
			},
		},
		Monorepo: MonorepoConfig{
			Enabled:    true,
			AutoDetect: true,
		},
		Contacts: ContactsConfig{
			Enabled:    true,
			MinPrimary: 1,
			MinBackup:  1,
		},
		// Default domains match Zero agent expertise areas
		SpecialistDomains: []string{
			"supply-chain",   // Cereal
			"security",       // Razor
			"compliance",     // Blade
			"legal",          // Phreak
			"frontend",       // Acid
			"backend",        // Dade
			"architecture",   // Nikon
			"cicd",           // Joey
			"infrastructure", // Plague
			"metrics",        // Gibson
			"crypto",         // Gill
			"ai-ml",          // Hal
		},
	}
}

// DomainPatterns maps domains to file path patterns
var DomainPatterns = map[string][]string{
	"supply-chain": {
		"package*.json",
		"go.mod",
		"go.sum",
		"requirements*.txt",
		"Pipfile*",
		"poetry.lock",
		"pyproject.toml",
		"Cargo.toml",
		"Cargo.lock",
		"pom.xml",
		"build.gradle*",
		"Gemfile*",
		"composer.json",
		"yarn.lock",
		"pnpm-lock.yaml",
	},
	"security": {
		"auth/*",
		"authentication/*",
		"authorization/*",
		"security/*",
		"**/auth/**",
		"**/security/**",
		"**/oauth/**",
		"**/jwt/**",
	},
	"compliance": {
		"policy/*",
		"audit/*",
		"compliance/*",
		"**/policies/**",
		"soc2/*",
		"hipaa/*",
	},
	"legal": {
		"LICENSE*",
		"NOTICE*",
		"legal/*",
		"COPYING*",
		"PATENTS*",
	},
	"frontend": {
		"*.tsx",
		"*.jsx",
		"*.vue",
		"*.svelte",
		"src/components/*",
		"**/components/**",
		"**/ui/**",
		"**/pages/**",
		"**/views/**",
		"*.css",
		"*.scss",
		"*.less",
	},
	"backend": {
		"api/*",
		"handlers/*",
		"routes/*",
		"controllers/*",
		"**/api/**",
		"**/handlers/**",
		"**/routes/**",
		"**/services/**",
		"**/middleware/**",
	},
	"architecture": {
		"docs/architecture/*",
		"architecture/*",
		"ADR/*",
		"adr/*",
		"design/*",
		"*.md",
		"docs/*.md",
	},
	"cicd": {
		".github/workflows/*",
		".gitlab-ci.yml",
		"Jenkinsfile",
		".circleci/*",
		".travis.yml",
		"azure-pipelines.yml",
		"bitbucket-pipelines.yml",
		".buildkite/*",
	},
	"infrastructure": {
		"terraform/*",
		"*.tf",
		"k8s/*",
		"kubernetes/*",
		"**/k8s/**",
		"helm/*",
		"docker/*",
		"Dockerfile*",
		"docker-compose*.yml",
		"ansible/*",
		"pulumi/*",
		"cloudformation/*",
	},
	"metrics": {
		"metrics/*",
		"telemetry/*",
		"monitoring/*",
		"**/metrics/**",
		"**/telemetry/**",
		"prometheus/*",
		"grafana/*",
	},
	"crypto": {
		"crypto/*",
		"encryption/*",
		"**/crypto/**",
		"**/encryption/**",
		"tls/*",
		"ssl/*",
		"certs/*",
		"certificates/*",
	},
	"ai-ml": {
		"models/*",
		"ml/*",
		"ai/*",
		"**/models/**",
		"**/ml/**",
		"*.pkl",
		"*.onnx",
		"*.pt",
		"*.pth",
		"*.h5",
		"*.pb",
		"*.tflite",
		"training/*",
		"inference/*",
	},
}

// ActivityThresholds defines days thresholds for activity status
var ActivityThresholds = struct {
	Active   int // Days since last commit to be "active"
	Recent   int // Days since last commit to be "recent"
	Stale    int // Days since last commit to be "stale"
	Inactive int // Days since last commit to be "inactive"
	// Beyond inactive = "abandoned"
}{
	Active:   30,
	Recent:   90,
	Stale:    180,
	Inactive: 365,
}

// BusFactorThresholds defines risk levels
var BusFactorThresholds = struct {
	Critical int // Bus factor at or below this is critical
	Warning  int // Bus factor at or below this is warning
	// Above warning = healthy
}{
	Critical: 1,
	Warning:  2,
}

// GitHubTokenMessage is shown when GITHUB_TOKEN is not set
// #nosec G101 -- This is an instructional message showing token format, not a real credential
const GitHubTokenMessage = `GitHub token not found. For full ownership analysis (PR reviews, team membership), set GITHUB_TOKEN:

1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Select scopes:
   - repo (for private repositories) OR public_repo (for public only)
   - read:org (to resolve team membership)
4. Set: export GITHUB_TOKEN=ghp_xxxxx

Running in degraded mode (git-only analysis)...`
