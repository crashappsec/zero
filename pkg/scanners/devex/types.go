package devex

// Result holds all feature results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	Onboarding *OnboardingSummary `json:"onboarding,omitempty"`
	Tooling    *ToolingSummary    `json:"tooling,omitempty"`
	Workflow   *WorkflowSummary   `json:"workflow,omitempty"`
	Errors     []string           `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	Onboarding *OnboardingFindings `json:"onboarding,omitempty"`
	Tooling    *ToolingFindings    `json:"tooling,omitempty"`
	Workflow   *WorkflowFindings   `json:"workflow,omitempty"`
}

// ============================================================================
// ONBOARDING TYPES
// ============================================================================

// OnboardingSummary contains onboarding friction summary
type OnboardingSummary struct {
	Score               int      `json:"score"`                 // Overall onboarding score (0-100, higher is easier)
	SetupComplexity     string   `json:"setup_complexity"`      // low, medium, high
	ConfigFileCount     int      `json:"config_file_count"`     // Number of config files in root
	DependencyCount     int      `json:"dependency_count"`      // Total dependencies
	BuildStepCount      int      `json:"build_step_count"`      // Estimated build steps
	EnvVarCount         int      `json:"env_var_count"`         // Required environment variables
	PrerequisiteCount   int      `json:"prerequisite_count"`    // External tool requirements
	HasContributing     bool     `json:"has_contributing"`      // CONTRIBUTING.md exists
	HasEnvExample       bool     `json:"has_env_example"`       // .env.example exists
	ReadmeQualityScore  int      `json:"readme_quality_score"`  // README quality (0-100)
	MissingDocs         []string `json:"missing_docs,omitempty"` // List of missing recommended docs
	Error               string   `json:"error,omitempty"`
}

// OnboardingFindings contains detailed onboarding analysis
type OnboardingFindings struct {
	ConfigFiles       []ConfigFile      `json:"config_files"`
	Prerequisites     []Prerequisite    `json:"prerequisites"`
	EnvVariables      []EnvVariable     `json:"env_variables,omitempty"`
	SetupBarriers     []SetupBarrier    `json:"setup_barriers"`
	ReadmeAnalysis    *ReadmeAnalysis   `json:"readme_analysis,omitempty"`
}

// ConfigFile represents a configuration file found in the repo
type ConfigFile struct {
	Path        string `json:"path"`
	Type        string `json:"type"`        // e.g., "package-manager", "linter", "build"
	Tool        string `json:"tool"`        // e.g., "npm", "eslint", "webpack"
	Complexity  string `json:"complexity"`  // low, medium, high
	LineCount   int    `json:"line_count"`
}

// Prerequisite represents an external tool requirement
type Prerequisite struct {
	Name        string `json:"name"`        // e.g., "Node.js", "Docker", "Python"
	Source      string `json:"source"`      // How detected (README, Dockerfile, etc.)
	Version     string `json:"version,omitempty"` // Required version if specified
	Required    bool   `json:"required"`    // Whether definitely required
}

// EnvVariable represents an environment variable requirement
type EnvVariable struct {
	Name        string `json:"name"`
	Source      string `json:"source"`       // .env.example, docker-compose, etc.
	HasDefault  bool   `json:"has_default"`
	Description string `json:"description,omitempty"`
}

// SetupBarrier represents something that makes setup harder
type SetupBarrier struct {
	Category    string `json:"category"`    // "documentation", "dependencies", "configuration"
	Severity    string `json:"severity"`    // low, medium, high
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
}

// ReadmeAnalysis contains README quality analysis
type ReadmeAnalysis struct {
	HasInstallSection    bool `json:"has_install_section"`
	HasUsageSection      bool `json:"has_usage_section"`
	HasPrerequisites     bool `json:"has_prerequisites"`
	HasQuickStart        bool `json:"has_quick_start"`
	HasExamples          bool `json:"has_examples"`
	WordCount            int  `json:"word_count"`
	HeaderCount          int  `json:"header_count"`
	CodeBlockCount       int  `json:"code_block_count"`
}

// ============================================================================
// TOOLING TYPES
// ============================================================================

// ToolingSummary contains tooling complexity summary
type ToolingSummary struct {
	Score               int            `json:"score"`                 // Overall tooling score (0-100, higher is simpler)
	ToolSprawlIndex     int            `json:"tool_sprawl_index"`     // Number of distinct tools
	SprawlLevel         string         `json:"sprawl_level"`          // low, moderate, high, excessive
	ConfigComplexity    string         `json:"config_complexity"`     // low, medium, high
	TotalConfigLines    int            `json:"total_config_lines"`    // Total lines across all configs
	BuildToolCount      int            `json:"build_tool_count"`      // Build/bundler tools
	LinterCount         int            `json:"linter_count"`          // Linting tools
	TestToolCount       int            `json:"test_tool_count"`       // Testing frameworks
	LanguageCount       int            `json:"language_count"`        // Programming languages detected
	CICDComplexity      int            `json:"cicd_complexity"`       // CI/CD pipeline complexity
	ToolsByCategory     map[string]int `json:"tools_by_category"`     // Tool counts by category
	Error               string         `json:"error,omitempty"`
}

// ToolingFindings contains detailed tooling analysis
type ToolingFindings struct {
	DetectedTools    []DetectedTool  `json:"detected_tools"`
	ConfigAnalysis   []ConfigAnalysis `json:"config_analysis,omitempty"`
	SprawlIssues     []SprawlIssue   `json:"sprawl_issues,omitempty"`
	Languages        []Language      `json:"languages"`
}

// DetectedTool represents a detected development tool
type DetectedTool struct {
	Name        string `json:"name"`         // e.g., "ESLint", "Prettier", "Webpack"
	Category    string `json:"category"`     // "linter", "formatter", "bundler", "test", "ci"
	ConfigFile  string `json:"config_file"`  // Configuration file that indicates this tool
	Version     string `json:"version,omitempty"` // Version if detectable
}

// ConfigAnalysis represents analysis of a config file
type ConfigAnalysis struct {
	Path           string `json:"path"`
	Tool           string `json:"tool"`
	LineCount      int    `json:"line_count"`
	NestingDepth   int    `json:"nesting_depth"`   // Max JSON/YAML nesting depth
	OverrideCount  int    `json:"override_count"`  // Number of rule overrides
	ComplexityScore int   `json:"complexity_score"` // 0-100
}

// SprawlIssue represents a tool sprawl concern
type SprawlIssue struct {
	Category    string   `json:"category"`     // "duplication", "overlap", "excessive"
	Severity    string   `json:"severity"`     // low, medium, high
	Description string   `json:"description"`
	Tools       []string `json:"tools"`        // Tools involved
	Suggestion  string   `json:"suggestion"`
}

// Language represents a detected programming language
type Language struct {
	Name       string `json:"name"`
	FileCount  int    `json:"file_count"`
	Percentage float64 `json:"percentage"` // Percentage of codebase
}

// ============================================================================
// WORKFLOW TYPES
// ============================================================================

// WorkflowSummary contains workflow efficiency summary
type WorkflowSummary struct {
	Score              int    `json:"score"`                // Overall workflow score (0-100)
	EfficiencyLevel    string `json:"efficiency_level"`     // low, medium, high
	FeedbackLoopScore  int    `json:"feedback_loop_score"`  // Local dev feedback speed (0-100)
	PRProcessScore     int    `json:"pr_process_score"`     // PR process quality (0-100)
	LocalDevScore      int    `json:"local_dev_score"`      // Local dev setup ease (0-100)
	HasPRTemplates     bool   `json:"has_pr_templates"`
	HasIssueTemplates  bool   `json:"has_issue_templates"`
	HasDevContainer    bool   `json:"has_devcontainer"`
	HasDockerCompose   bool   `json:"has_docker_compose"`
	HasHotReload       bool   `json:"has_hot_reload"`
	HasWatchMode       bool   `json:"has_watch_mode"`
	Error              string `json:"error,omitempty"`
}

// WorkflowFindings contains detailed workflow analysis
type WorkflowFindings struct {
	PRTemplates      []PRTemplate      `json:"pr_templates,omitempty"`
	IssueTemplates   []IssueTemplate   `json:"issue_templates,omitempty"`
	DevSetup         *DevSetup         `json:"dev_setup,omitempty"`
	FeedbackTools    []FeedbackTool    `json:"feedback_tools,omitempty"`
	WorkflowIssues   []WorkflowIssue   `json:"workflow_issues,omitempty"`
}

// PRTemplate represents a PR template
type PRTemplate struct {
	Path          string   `json:"path"`
	HasChecklist  bool     `json:"has_checklist"`
	HasSections   bool     `json:"has_sections"`
	Sections      []string `json:"sections,omitempty"`
}

// IssueTemplate represents an issue template
type IssueTemplate struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Type     string `json:"type"` // bug, feature, etc.
}

// DevSetup represents local development setup options
type DevSetup struct {
	HasDockerCompose   bool     `json:"has_docker_compose"`
	HasDevContainer    bool     `json:"has_devcontainer"`
	HasMakefile        bool     `json:"has_makefile"`
	HasTaskfile        bool     `json:"has_taskfile"` // task.yaml / Taskfile.yml
	DevScripts         []string `json:"dev_scripts,omitempty"` // npm run dev, make dev, etc.
	SetupCommands      []string `json:"setup_commands,omitempty"` // Detected setup commands
}

// FeedbackTool represents a tool that provides fast feedback
type FeedbackTool struct {
	Name        string `json:"name"`         // e.g., "Hot Reload", "Watch Mode", "Live Server"
	Type        string `json:"type"`         // "hot_reload", "watch", "live_server"
	Source      string `json:"source"`       // Config file where detected
	Description string `json:"description,omitempty"`
}

// WorkflowIssue represents a workflow friction point
type WorkflowIssue struct {
	Category    string `json:"category"`    // "pr_process", "local_dev", "feedback_loop"
	Severity    string `json:"severity"`    // low, medium, high
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
}

// ============================================================================
// OVERALL DX SCORE
// ============================================================================

// DXScoreBreakdown provides a breakdown of the overall DX score
type DXScoreBreakdown struct {
	OverallScore     int `json:"overall_score"`     // 0-100
	OnboardingWeight int `json:"onboarding_weight"` // Weight applied (default 40%)
	ToolingWeight    int `json:"tooling_weight"`    // Weight applied (default 30%)
	WorkflowWeight   int `json:"workflow_weight"`   // Weight applied (default 30%)
}
