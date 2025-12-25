package codecrypto

import "time"

// Result holds all feature results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	Ciphers      *CiphersSummary      `json:"ciphers,omitempty"`
	Keys         *KeysSummary         `json:"keys,omitempty"`
	Random       *RandomSummary       `json:"random,omitempty"`
	TLS          *TLSSummary          `json:"tls,omitempty"`
	Certificates *CertificatesSummary `json:"certificates,omitempty"`
	Errors       []string             `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	Ciphers      []CipherFinding      `json:"ciphers,omitempty"`
	Keys         []KeyFinding         `json:"keys,omitempty"`
	Random       []RandomFinding      `json:"random,omitempty"`
	TLS          []TLSFinding         `json:"tls,omitempty"`
	Certificates *CertificatesResult  `json:"certificates,omitempty"`
}

// Feature summaries

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

// Finding types

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
	Certificates []CertInfo     `json:"certificates"`
	Findings     []CertFinding  `json:"findings,omitempty"`
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
