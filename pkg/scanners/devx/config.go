// Package devex provides the consolidated developer experience super scanner
// Features: onboarding, tooling, workflow
package devx

// FeatureConfig holds configuration for all developer experience features
type FeatureConfig struct {
	Onboarding OnboardingConfig `json:"onboarding"`
	Sprawl     SprawlConfig     `json:"sprawl"`
	Workflow   WorkflowConfig   `json:"workflow"`
}

// OnboardingConfig configures onboarding friction analysis
type OnboardingConfig struct {
	Enabled            bool `json:"enabled"`
	CheckReadmeQuality bool `json:"check_readme_quality"` // Check for setup instructions in README
	CheckContributing  bool `json:"check_contributing"`   // Check for CONTRIBUTING.md
	CheckEnvSetup      bool `json:"check_env_setup"`      // Check for .env.example, docker-compose
}

// SprawlConfig configures sprawl analysis (tool + technology sprawl)
type SprawlConfig struct {
	Enabled                    bool     `json:"enabled"`
	CheckToolSprawl            bool     `json:"check_tool_sprawl"`             // Count dev tools
	CheckTechnologySprawl      bool     `json:"check_technology_sprawl"`       // Count technologies
	CheckConfigComplexity      bool     `json:"check_config_complexity"`       // Analyze config file complexity
	MaxRecommendedTools        int      `json:"max_recommended_tools"`         // Threshold for tool sprawl warning (default: 10)
	MaxRecommendedTechnologies int      `json:"max_recommended_technologies"`  // Threshold for tech sprawl warning (default: 15)
	ToolCategories             []string `json:"tool_categories"`               // Categories for tools
	TechnologyCategories       []string `json:"technology_categories"`         // Categories for technologies
}

// ToolingConfig is kept for backward compatibility
// Deprecated: use SprawlConfig instead
type ToolingConfig = SprawlConfig

// WorkflowConfig configures workflow efficiency analysis
type WorkflowConfig struct {
	Enabled           bool `json:"enabled"`
	CheckPRTemplates  bool `json:"check_pr_templates"`  // Check for PR templates
	CheckLocalDev     bool `json:"check_local_dev"`     // Check for local dev setup (docker-compose, devcontainer)
	CheckFeedbackLoop bool `json:"check_feedback_loop"` // Check for hot reload, watch mode
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
		Onboarding: OnboardingConfig{
			Enabled:            true,
			CheckReadmeQuality: true,
			CheckContributing:  true,
			CheckEnvSetup:      true,
		},
		Sprawl: SprawlConfig{
			Enabled:                    true,
			CheckToolSprawl:            true,
			CheckTechnologySprawl:      true,
			CheckConfigComplexity:      true,
			MaxRecommendedTools:        10,
			MaxRecommendedTechnologies: 15,
			ToolCategories:             []string{"linter", "formatter", "bundler", "test", "ci-cd", "build"},
			TechnologyCategories:       []string{"language", "framework", "database", "cloud", "container", "infrastructure"},
		},
		Workflow: WorkflowConfig{
			Enabled:           true,
			CheckPRTemplates:  true,
			CheckLocalDev:     true,
			CheckFeedbackLoop: true,
		},
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Sprawl.CheckConfigComplexity = false // Skip detailed config analysis
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return DefaultConfig()
}
