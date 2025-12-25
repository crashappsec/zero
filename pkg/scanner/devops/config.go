// Package devops provides the consolidated DevOps and CI/CD security super scanner
// Renamed from infra - now includes all infrastructure, CI/CD, and GitHub Actions security
package devops

// FeatureConfig holds configuration for all DevOps analysis features
type FeatureConfig struct {
	IaC           IaCConfig           `json:"iac"`
	Containers    ContainersConfig    `json:"containers"`
	GitHubActions GitHubActionsConfig `json:"github_actions"`
	DORA          DORAConfig          `json:"dora"`
	Git           GitConfig           `json:"git"`
}

// IaCConfig configures Infrastructure as Code scanning
type IaCConfig struct {
	Enabled      bool   `json:"enabled"`
	Tool         string `json:"tool"`          // checkov, trivy, auto
	FallbackTool bool   `json:"fallback_tool"` // Use trivy if checkov fails
	ScanSecrets  bool   `json:"scan_secrets"`  // Scan for hardcoded secrets in IaC files
}

// ContainersConfig configures container image scanning
type ContainersConfig struct {
	Enabled        bool `json:"enabled"`
	ScanBaseImages bool `json:"scan_base_images"` // Scan images from Dockerfiles
}

// GitHubActionsConfig configures GitHub Actions security scanning
type GitHubActionsConfig struct {
	Enabled          bool `json:"enabled"`
	CheckPinning     bool `json:"check_pinning"`     // Check if actions are pinned to SHA
	CheckSecrets     bool `json:"check_secrets"`     // Check for secret exposure
	CheckInjection   bool `json:"check_injection"`   // Check for injection vulnerabilities
	CheckPermissions bool `json:"check_permissions"` // Check for excessive permissions
}

// DORAConfig configures DORA metrics calculation
type DORAConfig struct {
	Enabled    bool `json:"enabled"`
	PeriodDays int  `json:"period_days"` // Analysis period (default 90)
}

// GitConfig configures git insights analysis
type GitConfig struct {
	Enabled         bool `json:"enabled"`
	IncludeChurn    bool `json:"include_churn"`    // Include high-churn file analysis
	IncludeAge      bool `json:"include_age"`      // Include code age analysis
	IncludePatterns bool `json:"include_patterns"` // Include commit pattern analysis
	IncludeBranches bool `json:"include_branches"` // Include branch analysis
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
		IaC: IaCConfig{
			Enabled:      true,
			Tool:         "auto",
			FallbackTool: true,
			ScanSecrets:  true,
		},
		Containers: ContainersConfig{
			Enabled:        true,
			ScanBaseImages: true,
		},
		GitHubActions: GitHubActionsConfig{
			Enabled:          true,
			CheckPinning:     true,
			CheckSecrets:     true,
			CheckInjection:   true,
			CheckPermissions: true,
		},
		DORA: DORAConfig{
			Enabled:    true,
			PeriodDays: 90,
		},
		Git: GitConfig{
			Enabled:         true,
			IncludeChurn:    true,
			IncludeAge:      true,
			IncludePatterns: true,
			IncludeBranches: true,
		},
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Containers.Enabled = false // Skip image scanning (slow)
	cfg.Git.IncludeChurn = false   // Skip churn analysis (slow)
	cfg.Git.IncludeAge = false     // Skip age analysis (slow)
	return cfg
}

// SecurityConfig returns security-focused config
func SecurityConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.DORA.Enabled = false // Skip DORA (not security-focused)
	cfg.Git.Enabled = false  // Skip git insights (not security-focused)
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return FeatureConfig{
		IaC: IaCConfig{
			Enabled:      true,
			Tool:         "auto",
			FallbackTool: true,
			ScanSecrets:  true,
		},
		Containers: ContainersConfig{
			Enabled:        true,
			ScanBaseImages: true,
		},
		GitHubActions: GitHubActionsConfig{
			Enabled:          true,
			CheckPinning:     true,
			CheckSecrets:     true,
			CheckInjection:   true,
			CheckPermissions: true,
		},
		DORA: DORAConfig{
			Enabled:    true,
			PeriodDays: 90,
		},
		Git: GitConfig{
			Enabled:         true,
			IncludeChurn:    true,
			IncludeAge:      true,
			IncludePatterns: true,
			IncludeBranches: true,
		},
	}
}
