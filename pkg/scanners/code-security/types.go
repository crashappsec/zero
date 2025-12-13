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
	Error         string         `json:"error,omitempty"`
}

// APISummary contains API security summary
type APISummary struct {
	TotalFindings int            `json:"total_findings"`
	Critical      int            `json:"critical"`
	High          int            `json:"high"`
	Medium        int            `json:"medium"`
	Low           int            `json:"low"`
	ByCategory    map[string]int `json:"by_category"`
	Error         string         `json:"error,omitempty"`
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
}

// APIFinding represents an API security finding
type APIFinding struct {
	RuleID      string `json:"rule_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Category    string `json:"category"`
	OWASPApi    string `json:"owasp_api,omitempty"`
}
