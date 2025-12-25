package codesecurity

// Result holds all feature results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	Vulns   *VulnsSummary   `json:"vulns,omitempty"`
	Secrets *SecretsSummary `json:"secrets,omitempty"`
	API     *APISummary     `json:"api,omitempty"`
	Errors  []string        `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	Vulns   []VulnFinding   `json:"vulns,omitempty"`
	Secrets []SecretFinding `json:"secrets,omitempty"`
	API     []APIFinding    `json:"api,omitempty"`
}

// Feature summaries

// VulnsSummary contains code vulnerability summary
type VulnsSummary struct {
	TotalFindings int            `json:"total_findings"`
	Critical      int            `json:"critical"`
	High          int            `json:"high"`
	Medium        int            `json:"medium"`
	Low           int            `json:"low"`
	ByCWE         map[string]int `json:"by_cwe,omitempty"`
	ByCategory    map[string]int `json:"by_category,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// SecretsSummary contains secret detection summary
type SecretsSummary struct {
	TotalFindings int            `json:"total_findings"`
	Critical      int            `json:"critical"`
	High          int            `json:"high"`
	Medium        int            `json:"medium"`
	Low           int            `json:"low"`
	RiskScore     int            `json:"risk_score"`
	RiskLevel     string         `json:"risk_level"`
	ByType        map[string]int `json:"by_type"`
	FilesAffected int            `json:"files_affected"`

	// Enhanced detection sources
	BySource        map[string]int `json:"by_source,omitempty"`         // semgrep, entropy, git_history
	EntropyFindings int            `json:"entropy_findings,omitempty"`  // Findings from entropy analysis
	HistoryFindings int            `json:"history_findings,omitempty"`  // Findings from git history
	RemovedSecrets  int            `json:"removed_secrets,omitempty"`   // Secrets later removed from history

	// AI analysis results
	FalsePositives   int `json:"false_positives,omitempty"`   // AI-identified false positives
	ConfirmedSecrets int `json:"confirmed_secrets,omitempty"` // AI-confirmed real secrets

	Error string `json:"error,omitempty"`
}

// APISummary contains API security summary
type APISummary struct {
	TotalFindings  int            `json:"total_findings"`
	Critical       int            `json:"critical"`
	High           int            `json:"high"`
	Medium         int            `json:"medium"`
	Low            int            `json:"low"`
	ByCategory     map[string]int `json:"by_category"`
	ByOWASPApi     map[string]int `json:"by_owasp_api,omitempty"`
	ByFramework    map[string]int `json:"by_framework,omitempty"`
	EndpointsFound int            `json:"endpoints_found,omitempty"`
	Error          string         `json:"error,omitempty"`
}

// Finding types

// VulnFinding represents a code vulnerability finding
type VulnFinding struct {
	RuleID      string   `json:"rule_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	Category    string   `json:"category,omitempty"`
	CWE         []string `json:"cwe,omitempty"`
	OWASP       []string `json:"owasp,omitempty"`
	Fix         string   `json:"fix,omitempty"`
}

// SecretFinding represents a detected secret
type SecretFinding struct {
	RuleID   string `json:"rule_id"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Snippet  string `json:"snippet"`

	// Detection source tracking
	Entropy         float64 `json:"entropy,omitempty"`          // Shannon entropy score (0-8)
	EntropyLevel    string  `json:"entropy_level,omitempty"`    // "low", "medium", "high"
	DetectionSource string  `json:"detection_source,omitempty"` // "semgrep", "entropy", "git_history", "iac-scanner"
	IaCType         string  `json:"iac_type,omitempty"`         // terraform, kubernetes, cloudformation, github-actions, helm

	// Git history context
	CommitInfo *CommitInfo `json:"commit_info,omitempty"` // For git history findings

	// AI analysis results
	AIConfidence    float64 `json:"ai_confidence,omitempty"`     // 0.0-1.0
	AIReasoning     string  `json:"ai_reasoning,omitempty"`      // Why it's FP or real
	IsFalsePositive *bool   `json:"is_false_positive,omitempty"` // AI determination

	// Remediation guidance
	Rotation        *RotationGuide `json:"rotation,omitempty"`         // Rotation steps, URLs, commands
	ServiceProvider string         `json:"service_provider,omitempty"` // "aws", "github", "stripe", etc.
}

// CommitInfo contains git commit context for history findings
type CommitInfo struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"short_hash"`
	Author    string `json:"author"`
	Email     string `json:"email"`
	Date      string `json:"date"`
	Message   string `json:"message"`
	IsRemoved bool   `json:"is_removed"` // Was the secret later removed?
}

// RotationGuide contains remediation guidance for rotating a secret
type RotationGuide struct {
	Priority       string   `json:"priority"`                  // "immediate", "high", "medium", "low"
	Steps          []string `json:"steps"`                     // Step-by-step rotation instructions
	RotationURL    string   `json:"rotation_url,omitempty"`    // Direct link to rotation page
	CLICommand     string   `json:"cli_command,omitempty"`     // CLI command to rotate
	AutomationHint string   `json:"automation_hint,omitempty"` // Vault, Secrets Manager, etc.
	ExpiresIn      string   `json:"expires_in,omitempty"`      // When the secret expires
}

// APIFinding represents an API security finding
type APIFinding struct {
	RuleID      string   `json:"rule_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Confidence  string   `json:"confidence,omitempty"` // high, medium, low
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Column      int      `json:"column,omitempty"`
	Snippet     string   `json:"snippet,omitempty"`
	Category    string   `json:"category"`
	OWASPApi    string   `json:"owasp_api,omitempty"`
	CWE         []string `json:"cwe,omitempty"`
	HTTPMethod  string   `json:"http_method,omitempty"` // GET, POST, PUT, DELETE, etc.
	Endpoint    string   `json:"endpoint,omitempty"`    // /api/users, /graphql, etc.
	Framework   string   `json:"framework,omitempty"`   // express, fastapi, django, etc.
	Remediation string   `json:"remediation,omitempty"`
}
