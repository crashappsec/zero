// Package codesecurity provides the consolidated code security super scanner
// Features: vulns, secrets, api
package codesecurity

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/scanners/common"
)

const (
	Name    = "code-security"
	Version = "3.2.0"
)

func init() {
	scanner.Register(&CodeSecurityScanner{})
}

// CodeSecurityScanner consolidates security-focused code analysis
type CodeSecurityScanner struct{}

func (s *CodeSecurityScanner) Name() string {
	return Name
}

func (s *CodeSecurityScanner) Description() string {
	return "Security-focused code analysis: vulnerabilities, secrets, API security"
}

func (s *CodeSecurityScanner) Dependencies() []string {
	return nil
}

func (s *CodeSecurityScanner) EstimateDuration(fileCount int) time.Duration {
	est := 15 + fileCount/300
	return time.Duration(est) * time.Second
}

func (s *CodeSecurityScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	cfg := getConfig(opts)

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Check if semgrep is available (needed for all features)
	hasSemgrep := common.ToolExists("semgrep")

	// Run features in parallel where possible
	if cfg.Vulns.Enabled && hasSemgrep {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runVulns(ctx, opts, cfg.Vulns)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "vulns")
			result.Summary.Vulns = summary
			result.Findings.Vulns = findings
			mu.Unlock()
		}()
	} else if cfg.Vulns.Enabled {
		mu.Lock()
		result.Summary.Errors = append(result.Summary.Errors, "vulns: semgrep not installed")
		mu.Unlock()
	}

	if cfg.Secrets.Enabled && hasSemgrep {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runSecrets(ctx, opts, cfg.Secrets)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "secrets")
			result.Summary.Secrets = summary
			result.Findings.Secrets = findings
			mu.Unlock()
		}()
	} else if cfg.Secrets.Enabled {
		mu.Lock()
		result.Summary.Errors = append(result.Summary.Errors, "secrets: semgrep not installed")
		mu.Unlock()
	}

	if cfg.API.Enabled && hasSemgrep {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runAPI(ctx, opts, cfg.API)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "api")
			result.Summary.API = summary
			result.Findings.API = findings
			mu.Unlock()
		}()
	} else if cfg.API.Enabled {
		mu.Lock()
		result.Summary.Errors = append(result.Summary.Errors, "api: semgrep not installed")
		mu.Unlock()
	}

	wg.Wait()

	scanResult := scanner.NewScanResult(Name, Version, start)
	scanResult.Repository = opts.RepoPath
	scanResult.SetSummary(result.Summary)
	scanResult.SetFindings(result.Findings)
	scanResult.SetMetadata(map[string]interface{}{
		"features_run": result.FeaturesRun,
	})

	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("creating output directory: %w", err)
		}
		resultFile := filepath.Join(opts.OutputDir, Name+".json")
		if err := scanResult.WriteJSON(resultFile); err != nil {
			return nil, fmt.Errorf("writing result: %w", err)
		}
	}

	return scanResult, nil
}

func getConfig(opts *scanner.ScanOptions) FeatureConfig {
	if opts.FeatureConfig == nil {
		return DefaultConfig()
	}

	data, err := json.Marshal(opts.FeatureConfig)
	if err != nil {
		return DefaultConfig()
	}

	var cfg FeatureConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}

	return cfg
}

// ============================================================================
// VULNS FEATURE
// ============================================================================

func (s *CodeSecurityScanner) runVulns(ctx context.Context, opts *scanner.ScanOptions, cfg VulnsConfig) (*VulnsSummary, []VulnFinding) {
	var findings []VulnFinding

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build semgrep args
	args := []string{
		"--config", "p/security-audit",
		"--json",
		"--metrics=off",
		"--timeout", "60",
		"--max-memory", "4096",
		"--exclude", "node_modules",
		"--exclude", "vendor",
		"--exclude", ".git",
		"--exclude", "dist",
		"--exclude", "build",
		"--exclude", "*.min.js",
	}

	if cfg.IncludeOWASP {
		args = append(args, "--config", "p/owasp-top-ten")
	}

	args = append(args, opts.RepoPath)

	result, err := common.RunCommand(ctx, "semgrep", args...)
	if err != nil || result == nil {
		return &VulnsSummary{Error: "semgrep execution failed"}, findings
	}

	findings, summary := parseVulnsOutput(result.Stdout, opts.RepoPath, cfg)
	return summary, findings
}

func parseVulnsOutput(data []byte, repoPath string, cfg VulnsConfig) ([]VulnFinding, *VulnsSummary) {
	var findings []VulnFinding
	summary := &VulnsSummary{
		ByCWE:      make(map[string]int),
		ByCategory: make(map[string]int),
	}

	var output struct {
		Results []struct {
			CheckID string `json:"check_id"`
			Path    string `json:"path"`
			Start   struct {
				Line int `json:"line"`
				Col  int `json:"col"`
			} `json:"start"`
			Extra struct {
				Severity string                 `json:"severity"`
				Message  string                 `json:"message"`
				Metadata map[string]interface{} `json:"metadata"`
			} `json:"extra"`
		} `json:"results"`
	}

	if err := json.Unmarshal(data, &output); err != nil {
		return findings, summary
	}

	for _, r := range output.Results {
		file := r.Path
		if strings.HasPrefix(file, repoPath) {
			file = strings.TrimPrefix(file, repoPath+"/")
		}

		severity := mapSemgrepSeverity(r.Extra.Severity)

		// Filter by minimum severity
		if !meetsMinimumSeverity(severity, cfg.SeverityMinimum) {
			continue
		}

		category := extractCategory(r.CheckID)
		cwe := extractCWEFromMetadata(r.Extra.Metadata)
		owasp := extractOWASPFromMetadata(r.Extra.Metadata)

		finding := VulnFinding{
			RuleID:      r.CheckID,
			Title:       extractTitle(r.CheckID),
			Description: r.Extra.Message,
			Severity:    severity,
			File:        file,
			Line:        r.Start.Line,
			Column:      r.Start.Col,
			Category:    category,
			CWE:         cwe,
			OWASP:       owasp,
		}
		findings = append(findings, finding)

		summary.TotalFindings++
		summary.ByCategory[category]++
		for _, c := range cwe {
			summary.ByCWE[c]++
		}

		switch severity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		}
	}

	return findings, summary
}

// ============================================================================
// SECRETS FEATURE
// ============================================================================

func (s *CodeSecurityScanner) runSecrets(ctx context.Context, opts *scanner.ScanOptions, cfg SecretsConfig) (*SecretsSummary, []SecretFinding) {
	var findings []SecretFinding

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := common.RunCommand(ctx, "semgrep",
		"--config", "p/secrets",
		"--json",
		"--metrics=off",
		"--timeout", "60",
		"--max-memory", "4096",
		"--exclude", "node_modules",
		"--exclude", "vendor",
		"--exclude", ".git",
		"--exclude", "dist",
		"--exclude", "build",
		"--exclude", "*.min.js",
		"--exclude", "package-lock.json",
		"--exclude", "yarn.lock",
		"--exclude", "pnpm-lock.yaml",
		"--exclude", "*.env.example",
		"--exclude", "*.env.sample",
		"--exclude", "*.env.template",
		opts.RepoPath,
	)

	if err != nil || result == nil {
		return &SecretsSummary{Error: "semgrep execution failed"}, findings
	}

	findings, summary := parseSecretsOutput(result.Stdout, opts.RepoPath, cfg)
	return summary, findings
}

func parseSecretsOutput(data []byte, repoPath string, cfg SecretsConfig) ([]SecretFinding, *SecretsSummary) {
	var findings []SecretFinding
	summary := &SecretsSummary{
		ByType:    make(map[string]int),
		RiskScore: 100,
		RiskLevel: "excellent",
	}

	var output struct {
		Results []struct {
			CheckID string `json:"check_id"`
			Path    string `json:"path"`
			Start   struct {
				Line int `json:"line"`
				Col  int `json:"col"`
			} `json:"start"`
			Extra struct {
				Severity string `json:"severity"`
				Message  string `json:"message"`
				Lines    string `json:"lines"`
			} `json:"extra"`
		} `json:"results"`
	}

	if err := json.Unmarshal(data, &output); err != nil {
		return findings, summary
	}

	filesSet := make(map[string]bool)

	for _, r := range output.Results {
		severity := mapSecretSeverity(r.CheckID, r.Extra.Severity)
		secretType := getSecretType(r.CheckID)

		file := r.Path
		if strings.HasPrefix(file, repoPath) {
			file = strings.TrimPrefix(file, repoPath+"/")
		}

		snippet := r.Extra.Lines
		if cfg.RedactSecrets {
			snippet = redactSecret(snippet)
		}

		finding := SecretFinding{
			RuleID:   r.CheckID,
			Type:     secretType,
			Severity: severity,
			Message:  r.Extra.Message,
			File:     file,
			Line:     r.Start.Line,
			Column:   r.Start.Col,
			Snippet:  snippet,
		}
		findings = append(findings, finding)

		summary.TotalFindings++
		summary.ByType[secretType]++
		filesSet[file] = true

		switch severity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		}
	}

	summary.FilesAffected = len(filesSet)

	// Calculate risk score
	penalty := summary.Critical*25 + summary.High*15 + summary.Medium*5 + summary.Low*2
	summary.RiskScore = 100 - penalty
	if summary.RiskScore < 0 {
		summary.RiskScore = 0
	}

	switch {
	case summary.RiskScore < 40:
		summary.RiskLevel = "critical"
	case summary.RiskScore < 60:
		summary.RiskLevel = "high"
	case summary.RiskScore < 80:
		summary.RiskLevel = "medium"
	case summary.RiskScore < 95:
		summary.RiskLevel = "low"
	default:
		summary.RiskLevel = "excellent"
	}

	return findings, summary
}

// ============================================================================
// API FEATURE
// ============================================================================

func (s *CodeSecurityScanner) runAPI(ctx context.Context, opts *scanner.ScanOptions, cfg APIConfig) (*APISummary, []APIFinding) {
	var findings []APIFinding

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := common.RunCommand(ctx, "semgrep",
		"--config", "p/owasp-top-ten",
		"--config", "p/security-audit",
		"--json",
		"--metrics=off",
		"--timeout", "120",
		"--max-memory", "4096",
		"--exclude", "node_modules",
		"--exclude", "vendor",
		"--exclude", ".git",
		"--exclude", "test",
		"--exclude", "tests",
		"--exclude", "*_test.go",
		"--exclude", "*.test.js",
		"--exclude", "*.spec.ts",
		opts.RepoPath,
	)

	if err != nil || result == nil {
		return &APISummary{Error: "semgrep execution failed"}, findings
	}

	findings, summary := parseAPIOutput(result.Stdout, opts.RepoPath, cfg)
	return summary, findings
}

func parseAPIOutput(data []byte, repoPath string, cfg APIConfig) ([]APIFinding, *APISummary) {
	var findings []APIFinding
	summary := &APISummary{
		ByCategory: make(map[string]int),
	}

	var output struct {
		Results []struct {
			CheckID string `json:"check_id"`
			Path    string `json:"path"`
			Start   struct {
				Line int `json:"line"`
			} `json:"start"`
			Extra struct {
				Severity string `json:"severity"`
				Message  string `json:"message"`
			} `json:"extra"`
		} `json:"results"`
	}

	if err := json.Unmarshal(data, &output); err != nil {
		return findings, summary
	}

	apiPatterns := []string{
		"auth", "authorization", "authentication",
		"injection", "sql", "nosql", "command",
		"ssrf", "request-forgery",
		"mass-assignment", "data-exposure",
		"rate-limit", "dos",
		"cors", "csrf",
		"jwt", "token", "session",
		"api", "rest", "graphql",
	}

	for _, r := range output.Results {
		ruleIDLower := strings.ToLower(r.CheckID)

		// Filter for API-relevant rules
		isAPIRelated := false
		for _, pattern := range apiPatterns {
			if strings.Contains(ruleIDLower, pattern) {
				isAPIRelated = true
				break
			}
		}

		if !isAPIRelated {
			continue
		}

		file := r.Path
		if strings.HasPrefix(file, repoPath) {
			file = strings.TrimPrefix(file, repoPath+"/")
		}

		severity := mapSemgrepSeverity(r.Extra.Severity)
		category := categorizeAPIFinding(r.CheckID)
		owaspApi := mapToOWASPAPI(r.CheckID)

		finding := APIFinding{
			RuleID:      r.CheckID,
			Title:       extractTitle(r.CheckID),
			Description: r.Extra.Message,
			Severity:    severity,
			File:        file,
			Line:        r.Start.Line,
			Category:    category,
			OWASPApi:    owaspApi,
		}
		findings = append(findings, finding)

		summary.TotalFindings++
		summary.ByCategory[category]++

		switch severity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		}
	}

	return findings, summary
}

// ============================================================================
// UTILITIES
// ============================================================================

func mapSemgrepSeverity(s string) string {
	switch strings.ToUpper(s) {
	case "ERROR":
		return "critical"
	case "WARNING":
		return "high"
	case "INFO":
		return "medium"
	default:
		return "low"
	}
}

func meetsMinimumSeverity(severity, minimum string) bool {
	severityOrder := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}
	return severityOrder[severity] >= severityOrder[minimum]
}

func extractCategory(ruleID string) string {
	parts := strings.Split(ruleID, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return "general"
}

func extractTitle(ruleID string) string {
	parts := strings.Split(ruleID, ".")
	if len(parts) > 0 {
		return strings.ReplaceAll(parts[len(parts)-1], "-", " ")
	}
	return ruleID
}

func extractCWEFromMetadata(metadata map[string]interface{}) []string {
	var cwes []string
	if cwe, ok := metadata["cwe"]; ok {
		switch v := cwe.(type) {
		case []interface{}:
			for _, c := range v {
				if s, ok := c.(string); ok {
					cwes = append(cwes, s)
				}
			}
		case string:
			cwes = append(cwes, v)
		}
	}
	return cwes
}

func extractOWASPFromMetadata(metadata map[string]interface{}) []string {
	var owasp []string
	if o, ok := metadata["owasp"]; ok {
		switch v := o.(type) {
		case []interface{}:
			for _, c := range v {
				if s, ok := c.(string); ok {
					owasp = append(owasp, s)
				}
			}
		case string:
			owasp = append(owasp, v)
		}
	}
	return owasp
}

func mapSecretSeverity(ruleID, semgrepSeverity string) string {
	ruleIDLower := strings.ToLower(ruleID)

	if strings.Contains(ruleIDLower, "aws") && strings.Contains(ruleIDLower, "access") ||
		strings.Contains(ruleIDLower, "private") && strings.Contains(ruleIDLower, "key") ||
		strings.Contains(ruleIDLower, "gcp") && strings.Contains(ruleIDLower, "service") && strings.Contains(ruleIDLower, "account") ||
		strings.Contains(ruleIDLower, "stripe") && strings.Contains(ruleIDLower, "live") {
		return "critical"
	}

	if strings.Contains(ruleIDLower, "github") && strings.Contains(ruleIDLower, "token") ||
		strings.Contains(ruleIDLower, "gitlab") && strings.Contains(ruleIDLower, "token") ||
		strings.Contains(ruleIDLower, "database") && strings.Contains(ruleIDLower, "url") ||
		strings.Contains(ruleIDLower, "jwt") && strings.Contains(ruleIDLower, "secret") ||
		strings.Contains(ruleIDLower, "api") && strings.Contains(ruleIDLower, "key") {
		return "high"
	}

	switch strings.ToUpper(semgrepSeverity) {
	case "ERROR":
		return "critical"
	case "WARNING":
		return "high"
	case "INFO":
		return "medium"
	default:
		return "medium"
	}
}

func getSecretType(ruleID string) string {
	ruleIDLower := strings.ToLower(ruleID)

	types := map[string][]string{
		"aws_credential":      {"aws"},
		"github_token":        {"github"},
		"gitlab_token":        {"gitlab"},
		"slack_token":         {"slack"},
		"stripe_key":          {"stripe"},
		"private_key":         {"private_key", "rsa", "dsa", "ec_private"},
		"database_credential": {"postgres", "mysql", "mongodb", "redis", "database"},
		"jwt_secret":          {"jwt"},
		"api_key":             {"api_key", "apikey"},
		"password":            {"password"},
		"generic_secret":      {"secret"},
	}

	for secretType, patterns := range types {
		for _, pattern := range patterns {
			if strings.Contains(ruleIDLower, pattern) {
				return secretType
			}
		}
	}

	return "unknown"
}

func redactSecret(snippet string) string {
	if len(snippet) > 200 {
		snippet = snippet[:200] + "..."
	}

	result := snippet
	words := strings.Fields(result)
	for i, word := range words {
		clean := strings.Trim(word, "\"'`=:,;")
		if len(clean) > 16 && isAlphanumericPlus(clean) {
			if len(clean) > 8 {
				masked := clean[:8] + "********"
				words[i] = strings.Replace(word, clean, masked, 1)
			}
		}
	}

	return strings.Join(words, " ")
}

func isAlphanumericPlus(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}

func categorizeAPIFinding(ruleID string) string {
	ruleIDLower := strings.ToLower(ruleID)

	categories := map[string][]string{
		"authentication":  {"auth", "jwt", "token", "session", "login", "password"},
		"authorization":   {"authorization", "permission", "access-control", "rbac"},
		"injection":       {"injection", "sql", "nosql", "command", "ldap", "xpath"},
		"data-exposure":   {"data-exposure", "sensitive", "pii", "logging"},
		"rate-limiting":   {"rate-limit", "dos", "throttle", "brute"},
		"ssrf":            {"ssrf", "request-forgery"},
		"mass-assignment": {"mass-assignment", "binding"},
		"misconfiguration": {"cors", "header", "config", "tls", "ssl"},
	}

	for category, patterns := range categories {
		for _, pattern := range patterns {
			if strings.Contains(ruleIDLower, pattern) {
				return category
			}
		}
	}

	return "general"
}

func mapToOWASPAPI(ruleID string) string {
	ruleIDLower := strings.ToLower(ruleID)

	mappings := map[string][]string{
		"API1 - BOLA":                      {"object-level", "bola", "idor"},
		"API2 - Broken Authentication":     {"auth", "authentication", "jwt", "token", "session"},
		"API3 - Broken Property Auth":      {"property", "field"},
		"API4 - Resource Consumption":      {"rate-limit", "dos", "resource"},
		"API5 - Broken Function Auth":      {"function", "admin", "privilege"},
		"API6 - Mass Assignment":           {"mass-assignment", "binding"},
		"API7 - SSRF":                      {"ssrf", "request-forgery"},
		"API8 - Security Misconfiguration": {"cors", "config", "header"},
		"API9 - Improper Inventory":        {"endpoint", "version"},
		"API10 - Unsafe Consumption":       {"external", "third-party"},
	}

	for apiCategory, patterns := range mappings {
		for _, pattern := range patterns {
			if strings.Contains(ruleIDLower, pattern) {
				return apiCategory
			}
		}
	}

	return ""
}
