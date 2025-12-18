// Package devex provides the consolidated developer experience super scanner
// Features: onboarding, tooling, workflow
package devex

// FeatureConfig holds configuration for all developer experience features
type FeatureConfig struct {
	Onboarding OnboardingConfig `json:"onboarding"`
	Tooling    ToolingConfig    `json:"tooling"`
	Workflow   WorkflowConfig   `json:"workflow"`
}

// OnboardingConfig configures onboarding friction analysis
type OnboardingConfig struct {
	Enabled            bool `json:"enabled"`
	CheckReadmeQuality bool `json:"check_readme_quality"` // Check for setup instructions in README
	CheckContributing  bool `json:"check_contributing"`   // Check for CONTRIBUTING.md
	CheckEnvSetup      bool `json:"check_env_setup"`      // Check for .env.example, docker-compose
	CheckPrerequisites bool `json:"check_prerequisites"`  // Check for required tool documentation
}

// ToolingConfig configures tooling complexity analysis
type ToolingConfig struct {
	Enabled              bool `json:"enabled"`
	CheckToolSprawl      bool `json:"check_tool_sprawl"`       // Count distinct tools
	CheckConfigComplexity bool `json:"check_config_complexity"` // Analyze config file complexity
	MaxRecommendedTools  int  `json:"max_recommended_tools"`   // Threshold for tool sprawl warning
}

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
			CheckPrerequisites: true,
		},
		Tooling: ToolingConfig{
			Enabled:              true,
			CheckToolSprawl:      true,
			CheckConfigComplexity: true,
			MaxRecommendedTools:  10,
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
	cfg.Tooling.CheckConfigComplexity = false // Skip detailed config analysis
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return DefaultConfig()
}
