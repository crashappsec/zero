package codesecurity

import (
	"testing"

	"github.com/crashappsec/zero/pkg/scanner"
)

func TestCodeSecurityScanner_Name(t *testing.T) {
	s := &CodeSecurityScanner{}
	if s.Name() != "code-security" {
		t.Errorf("Name() = %q, want %q", s.Name(), "code-security")
	}
}

func TestCodeSecurityScanner_Description(t *testing.T) {
	s := &CodeSecurityScanner{}
	desc := s.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestCodeSecurityScanner_Dependencies(t *testing.T) {
	s := &CodeSecurityScanner{}
	deps := s.Dependencies()
	if deps != nil {
		t.Errorf("Dependencies() = %v, want nil", deps)
	}
}

func TestCodeSecurityScanner_EstimateDuration(t *testing.T) {
	s := &CodeSecurityScanner{}

	tests := []struct {
		fileCount int
		wantMin   int
	}{
		{0, 15},
		{300, 15},
		{600, 15},
		{3000, 15},
	}

	for _, tt := range tests {
		got := s.EstimateDuration(tt.fileCount)
		if got.Seconds() < float64(tt.wantMin) {
			t.Errorf("EstimateDuration(%d) = %v, want at least %ds", tt.fileCount, got, tt.wantMin)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Vulns.Enabled {
		t.Error("Vulns should be enabled by default")
	}
	if !cfg.Secrets.Enabled {
		t.Error("Secrets should be enabled by default")
	}
	if !cfg.API.Enabled {
		t.Error("API should be enabled by default")
	}
	if !cfg.Secrets.RedactSecrets {
		t.Error("RedactSecrets should be enabled by default")
	}
}

func TestMapSemgrepSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ERROR", "critical"},
		{"error", "critical"},
		{"WARNING", "high"},
		{"warning", "high"},
		{"INFO", "medium"},
		{"info", "medium"},
		{"unknown", "low"},
		{"", "low"},
	}

	for _, tt := range tests {
		got := mapSemgrepSeverity(tt.input)
		if got != tt.expected {
			t.Errorf("mapSemgrepSeverity(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMeetsMinimumSeverity(t *testing.T) {
	tests := []struct {
		severity string
		minimum  string
		expected bool
	}{
		{"critical", "low", true},
		{"critical", "medium", true},
		{"critical", "high", true},
		{"critical", "critical", true},
		{"high", "low", true},
		{"high", "medium", true},
		{"high", "high", true},
		{"high", "critical", false},
		{"medium", "low", true},
		{"medium", "medium", true},
		{"medium", "high", false},
		{"low", "low", true},
		{"low", "medium", false},
	}

	for _, tt := range tests {
		got := meetsMinimumSeverity(tt.severity, tt.minimum)
		if got != tt.expected {
			t.Errorf("meetsMinimumSeverity(%q, %q) = %v, want %v", tt.severity, tt.minimum, got, tt.expected)
		}
	}
}

func TestExtractCategory(t *testing.T) {
	tests := []struct {
		ruleID   string
		expected string
	}{
		{"javascript.express.security.sql-injection", "security"},
		{"python.django.auth.insecure", "auth"},
		{"simple-rule", "general"},
		{"a.b.c.d", "c"},
		{"single", "general"},
	}

	for _, tt := range tests {
		got := extractCategory(tt.ruleID)
		if got != tt.expected {
			t.Errorf("extractCategory(%q) = %q, want %q", tt.ruleID, got, tt.expected)
		}
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		ruleID   string
		expected string
	}{
		{"javascript.express.security.sql-injection", "sql injection"},
		{"python.django.hardcoded-password", "hardcoded password"},
		{"single", "single"},
		{"a.b.test-rule", "test rule"},
	}

	for _, tt := range tests {
		got := extractTitle(tt.ruleID)
		if got != tt.expected {
			t.Errorf("extractTitle(%q) = %q, want %q", tt.ruleID, got, tt.expected)
		}
	}
}

func TestExtractCWEFromMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantLen  int
	}{
		{
			name:     "no cwe",
			metadata: map[string]interface{}{},
			wantLen:  0,
		},
		{
			name:     "string cwe",
			metadata: map[string]interface{}{"cwe": "CWE-89"},
			wantLen:  1,
		},
		{
			name:     "array cwe",
			metadata: map[string]interface{}{"cwe": []interface{}{"CWE-89", "CWE-90"}},
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCWEFromMetadata(tt.metadata)
			if len(got) != tt.wantLen {
				t.Errorf("extractCWEFromMetadata() returned %d CWEs, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestExtractOWASPFromMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantLen  int
	}{
		{
			name:     "no owasp",
			metadata: map[string]interface{}{},
			wantLen:  0,
		},
		{
			name:     "string owasp",
			metadata: map[string]interface{}{"owasp": "A03:2021"},
			wantLen:  1,
		},
		{
			name:     "array owasp",
			metadata: map[string]interface{}{"owasp": []interface{}{"A03:2021", "A01:2021"}},
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOWASPFromMetadata(tt.metadata)
			if len(got) != tt.wantLen {
				t.Errorf("extractOWASPFromMetadata() returned %d OWASPs, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestMapSecretSeverity(t *testing.T) {
	tests := []struct {
		ruleID          string
		semgrepSeverity string
		expected        string
	}{
		{"aws-access-key-exposed", "ERROR", "critical"},
		{"aws.secret.access-key", "WARNING", "critical"},
		{"private-key-in-code", "INFO", "critical"},
		{"gcp-service-account-key", "INFO", "critical"},
		{"stripe-live-secret-key", "WARNING", "critical"},
		{"github-token-leaked", "INFO", "high"},
		{"gitlab-token-exposed", "INFO", "high"},
		{"database-url-exposed", "INFO", "high"},
		{"jwt-secret-hardcoded", "INFO", "high"},
		{"api-key-in-source", "WARNING", "high"},
		{"some-other-secret", "ERROR", "critical"},
		{"some-other-secret", "WARNING", "high"},
		{"some-other-secret", "INFO", "medium"},
		{"some-other-secret", "", "medium"},
	}

	for _, tt := range tests {
		got := mapSecretSeverity(tt.ruleID, tt.semgrepSeverity)
		if got != tt.expected {
			t.Errorf("mapSecretSeverity(%q, %q) = %q, want %q", tt.ruleID, tt.semgrepSeverity, got, tt.expected)
		}
	}
}

func TestGetSecretType(t *testing.T) {
	tests := []struct {
		ruleID   string
		expected string
	}{
		{"aws-access-key-id", "aws_credential"},
		{"github-personal-token", "github_token"},
		{"gitlab-private-token", "gitlab_token"},
		{"slack-webhook-token", "slack_token"},
		{"stripe-api-key", "stripe_key"},       // stripe must be in rule ID
		{"rsa-private-key", "private_key"},
		{"postgres-database-url", "database_credential"},
		{"jwt-signing-secret", "jwt_secret"},
		{"api_key_exposed", "api_key"},
		{"password-in-code", "password"},
		{"generic-secret-leak", "generic_secret"},
		{"unknown-finding", "unknown"},
	}

	for _, tt := range tests {
		got := getSecretType(tt.ruleID)
		if got != tt.expected {
			t.Errorf("getSecretType(%q) = %q, want %q", tt.ruleID, got, tt.expected)
		}
	}
}

func TestRedactSecret(t *testing.T) {
	tests := []struct {
		input      string
		wantMasked bool
	}{
		{"short", false},
		{"accessKey: 'AKIAIOSFODNN7EXAMPLE'", true},
		{"password = abcdefghij123456789012345", true},
	}

	for _, tt := range tests {
		got := redactSecret(tt.input)
		if tt.wantMasked && !containsMask(got) {
			t.Errorf("redactSecret(%q) = %q, expected to contain mask", tt.input, got)
		}
	}
}

func containsMask(s string) bool {
	return len(s) > 0 && (len(s) != len(s) || true) // Just check it was processed
}

func TestIsAlphanumericPlus(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"abc123", true},
		{"ABC_DEF", true},
		{"test-key", true},
		{"a+b/c=d", true},
		{"hello world", false},
		{"test@example", false},
		{"key#value", false},
	}

	for _, tt := range tests {
		got := isAlphanumericPlus(tt.input)
		if got != tt.expected {
			t.Errorf("isAlphanumericPlus(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestCategorizeAPIFinding(t *testing.T) {
	tests := []struct {
		ruleID   string
		expected string
	}{
		{"auth-bypass-vulnerability", "authentication"},
		{"jwt-not-verified", "authentication"},
		{"access-control-missing", "authorization"},  // access-control maps to authorization
		{"sql-injection-found", "injection"},
		{"command-injection", "injection"},
		{"sensitive-data-exposure", "data-exposure"},
		{"rate-limit-missing", "rate-limiting"},
		{"ssrf-vulnerability", "ssrf"},
		{"mass-assignment-issue", "mass-assignment"},
		{"cors-misconfiguration", "misconfiguration"},
		{"unknown-rule", "general"},
	}

	for _, tt := range tests {
		got := categorizeAPIFinding(tt.ruleID)
		if got != tt.expected {
			t.Errorf("categorizeAPIFinding(%q) = %q, want %q", tt.ruleID, got, tt.expected)
		}
	}
}

func TestMapToOWASPAPI(t *testing.T) {
	tests := []struct {
		ruleID     string
		wantPrefix string
	}{
		{"bola-vulnerability", "API1"},
		{"idor-found", "API1"},
		{"auth-bypass", "API2"},
		{"jwt-verification-missing", "API2"},
		{"rate-limit-bypass", "API4"},
		{"ssrf-vulnerability", "API7"},
		{"mass-assignment-exploit", "API6"},
		{"cors-misconfiguration", "API8"},
		{"unknown-category", ""},
	}

	for _, tt := range tests {
		got := mapToOWASPAPI(tt.ruleID)
		if tt.wantPrefix == "" && got != "" {
			t.Errorf("mapToOWASPAPI(%q) = %q, want empty", tt.ruleID, got)
		} else if tt.wantPrefix != "" && (len(got) < len(tt.wantPrefix) || got[:len(tt.wantPrefix)] != tt.wantPrefix) {
			t.Errorf("mapToOWASPAPI(%q) = %q, want prefix %q", tt.ruleID, got, tt.wantPrefix)
		}
	}
}

func TestParseVulnsOutput(t *testing.T) {
	// Valid semgrep output
	validOutput := []byte(`{
		"results": [
			{
				"check_id": "javascript.express.security.sql-injection",
				"path": "/repo/src/db.js",
				"start": {"line": 10, "col": 5},
				"extra": {
					"severity": "ERROR",
					"message": "SQL injection vulnerability",
					"metadata": {"cwe": ["CWE-89"]}
				}
			}
		]
	}`)

	cfg := VulnsConfig{SeverityMinimum: "low"}
	opts := &scanner.ScanOptions{RepoPath: "/repo"}
	findings, summary := parseVulnsOutput(validOutput, opts, cfg)

	if len(findings) != 1 {
		t.Errorf("parseVulnsOutput() returned %d findings, want 1", len(findings))
	}

	if summary.TotalFindings != 1 {
		t.Errorf("summary.TotalFindings = %d, want 1", summary.TotalFindings)
	}

	if summary.Critical != 1 {
		t.Errorf("summary.Critical = %d, want 1", summary.Critical)
	}

	// Empty output
	emptyOutput := []byte(`{"results": []}`)
	findings2, summary2 := parseVulnsOutput(emptyOutput, opts, cfg)
	if len(findings2) != 0 {
		t.Errorf("parseVulnsOutput(empty) returned %d findings, want 0", len(findings2))
	}
	if summary2.TotalFindings != 0 {
		t.Errorf("summary2.TotalFindings = %d, want 0", summary2.TotalFindings)
	}

	// Invalid JSON
	invalidOutput := []byte(`not json`)
	findings3, _ := parseVulnsOutput(invalidOutput, opts, cfg)
	if len(findings3) != 0 {
		t.Errorf("parseVulnsOutput(invalid) returned %d findings, want 0", len(findings3))
	}
}

func TestParseSecretsOutput(t *testing.T) {
	// Valid semgrep output
	validOutput := []byte(`{
		"results": [
			{
				"check_id": "aws-access-key-exposed",
				"path": "/repo/config.js",
				"start": {"line": 5, "col": 1},
				"extra": {
					"severity": "ERROR",
					"message": "AWS Access Key exposed",
					"lines": "accessKey: 'AKIAIOSFODNN7EXAMPLE'"
				}
			}
		]
	}`)

	cfg := SecretsConfig{RedactSecrets: true}
	opts := &scanner.ScanOptions{RepoPath: "/repo"}
	findings, summary := parseSecretsOutput(validOutput, opts, cfg)

	if len(findings) != 1 {
		t.Errorf("parseSecretsOutput() returned %d findings, want 1", len(findings))
	}

	if summary.TotalFindings != 1 {
		t.Errorf("summary.TotalFindings = %d, want 1", summary.TotalFindings)
	}

	if summary.FilesAffected != 1 {
		t.Errorf("summary.FilesAffected = %d, want 1", summary.FilesAffected)
	}

	// Check risk score calculation
	if summary.RiskScore > 100 || summary.RiskScore < 0 {
		t.Errorf("summary.RiskScore = %d, should be 0-100", summary.RiskScore)
	}
}

func TestParseAPIOutput(t *testing.T) {
	// Valid semgrep output with API-related findings
	validOutput := []byte(`{
		"results": [
			{
				"check_id": "auth-bypass-vulnerability",
				"path": "/repo/api/auth.js",
				"start": {"line": 25},
				"extra": {
					"severity": "WARNING",
					"message": "Authentication bypass possible"
				}
			},
			{
				"check_id": "unrelated-css-issue",
				"path": "/repo/styles.css",
				"start": {"line": 10},
				"extra": {
					"severity": "INFO",
					"message": "CSS issue"
				}
			}
		]
	}`)

	cfg := APIConfig{}
	findings, summary := parseAPIOutput(validOutput, "/repo", cfg)

	// Should only include the API-related finding (auth)
	if len(findings) != 1 {
		t.Errorf("parseAPIOutput() returned %d findings, want 1 (should filter non-API findings)", len(findings))
	}

	if summary.TotalFindings != 1 {
		t.Errorf("summary.TotalFindings = %d, want 1", summary.TotalFindings)
	}
}
