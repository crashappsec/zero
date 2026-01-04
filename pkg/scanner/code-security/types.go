package codesecurity

import "time"

// Result holds all feature results
type Result struct {
	FeaturesRun        []string                  `json:"features_run"`
	Summary            Summary                   `json:"summary"`
	Findings           Findings                  `json:"findings"`
	GitHistorySecurity *GitHistorySecurityResult `json:"git_history_security,omitempty"`
}

// Summary holds summaries from all features
type Summary struct {
	Vulns   *VulnsSummary   `json:"vulns,omitempty"`
	Secrets *SecretsSummary `json:"secrets,omitempty"`
	API     *APISummary     `json:"api,omitempty"`
	// Crypto summaries (merged from code-crypto)
	Ciphers      *CiphersSummary      `json:"ciphers,omitempty"`
	Keys         *KeysSummary         `json:"keys,omitempty"`
	Random       *RandomSummary       `json:"random,omitempty"`
	TLS          *TLSSummary          `json:"tls,omitempty"`
	Certificates *CertificatesSummary `json:"certificates,omitempty"`
	Errors       []string             `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	Vulns   []VulnFinding   `json:"vulns,omitempty"`
	Secrets []SecretFinding `json:"secrets,omitempty"`
	API     []APIFinding    `json:"api,omitempty"`
	// Crypto findings (merged from code-crypto)
	Ciphers      []CipherFinding     `json:"ciphers,omitempty"`
	Keys         []KeyFinding        `json:"keys,omitempty"`
	Random       []RandomFinding     `json:"random,omitempty"`
	TLS          []TLSFinding        `json:"tls,omitempty"`
	Certificates *CertificatesResult `json:"certificates,omitempty"`
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

// Crypto feature summaries (merged from code-crypto)

// CiphersSummary contains weak cipher detection summary
type CiphersSummary struct {
	TotalFindings int            `json:"total_findings"`
	BySeverity    map[string]int `json:"by_severity"`
	ByAlgorithm   map[string]int `json:"by_algorithm"`
	UsedSemgrep   bool           `json:"used_semgrep"`
	Error         string         `json:"error,omitempty"`
}

// KeysSummary contains hardcoded key detection summary
type KeysSummary struct {
	TotalFindings int            `json:"total_findings"`
	BySeverity    map[string]int `json:"by_severity"`
	ByType        map[string]int `json:"by_type"`
	Error         string         `json:"error,omitempty"`
}

// RandomSummary contains insecure random detection summary
type RandomSummary struct {
	TotalFindings int            `json:"total_findings"`
	BySeverity    map[string]int `json:"by_severity"`
	ByType        map[string]int `json:"by_type"`
	Error         string         `json:"error,omitempty"`
}

// TLSSummary contains TLS misconfiguration summary
type TLSSummary struct {
	TotalFindings int            `json:"total_findings"`
	BySeverity    map[string]int `json:"by_severity"`
	ByType        map[string]int `json:"by_type"`
	Error         string         `json:"error,omitempty"`
}

// CertificatesSummary contains certificate analysis summary
type CertificatesSummary struct {
	TotalCertificates int            `json:"total_certificates"`
	TotalFindings     int            `json:"total_findings"`
	ExpiringSoon      int            `json:"expiring_soon"`
	Expired           int            `json:"expired"`
	WeakKey           int            `json:"weak_key"`
	BySeverity        map[string]int `json:"by_severity"`
	Error             string         `json:"error,omitempty"`
}

// Crypto finding types (merged from code-crypto)

// CipherFinding represents a weak cipher finding
type CipherFinding struct {
	Algorithm   string `json:"algorithm"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Match       string `json:"match,omitempty"`
	Suggestion  string `json:"suggestion"`
	CWE         string `json:"cwe"`
	Source      string `json:"source"` // "semgrep" or "pattern"
}

// KeyFinding represents a hardcoded key finding
type KeyFinding struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Match       string `json:"match,omitempty"`
	CWE         string `json:"cwe"`
}

// RandomFinding represents an insecure random finding
type RandomFinding struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Match       string `json:"match,omitempty"`
	Suggestion  string `json:"suggestion"`
	CWE         string `json:"cwe"`
}

// TLSFinding represents a TLS misconfiguration finding
type TLSFinding struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Match       string `json:"match,omitempty"`
	Suggestion  string `json:"suggestion"`
	CWE         string `json:"cwe"`
}

// CertificatesResult holds certificate analysis results
type CertificatesResult struct {
	Certificates []CertInfo    `json:"certificates"`
	Findings     []CertFinding `json:"findings,omitempty"`
}

// CertInfo contains information about an X.509 certificate
type CertInfo struct {
	File          string    `json:"file"`
	Subject       string    `json:"subject"`
	Issuer        string    `json:"issuer"`
	NotBefore     time.Time `json:"not_before"`
	NotAfter      time.Time `json:"not_after"`
	DaysUntilExp  int       `json:"days_until_expiry"`
	KeyType       string    `json:"key_type"`
	KeySize       int       `json:"key_size"`
	SignatureAlgo string    `json:"signature_algorithm"`
	IsSelfSigned  bool      `json:"is_self_signed"`
	IsCA          bool      `json:"is_ca"`
	DNSNames      []string  `json:"dns_names,omitempty"`
	Serial        string    `json:"serial"`
}

// CertFinding represents a certificate issue
type CertFinding struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
}
