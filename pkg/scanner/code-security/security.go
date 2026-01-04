// Package codesecurity provides the consolidated code security super scanner
// Features: vulns, secrets, api, ciphers, keys, random, tls, certificates
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
	"github.com/crashappsec/zero/pkg/scanner/common"
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
	return "Security-focused code analysis: vulnerabilities, secrets, API security, cryptography"
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

	// Run git history security scanning if enabled (no semgrep required)
	if cfg.Secrets.GitHistorySecurity.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			securityScanner := NewGitHistorySecurityScanner(cfg.Secrets.GitHistorySecurity)
			securityResult, err := securityScanner.ScanRepository(opts.RepoPath)
			// Include results even with errors (may have partial results)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "git_history_security")
			if securityResult != nil {
				result.GitHistorySecurity = securityResult
			}
			if err != nil {
				// Add error to result summary
				result.Summary.Errors = append(result.Summary.Errors, "git_history_security: "+err.Error())
			}
			mu.Unlock()
		}()
	}

	// Crypto features (merged from code-crypto scanner)
	if cfg.Ciphers.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runCiphers(ctx, opts, cfg.Ciphers)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "ciphers")
			result.Summary.Ciphers = summary
			result.Findings.Ciphers = findings
			mu.Unlock()
		}()
	}

	if cfg.Keys.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runKeys(ctx, opts, cfg.Keys)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "keys")
			result.Summary.Keys = summary
			result.Findings.Keys = findings
			mu.Unlock()
		}()
	}

	if cfg.Random.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runRandom(ctx, opts, cfg.Random)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "random")
			result.Summary.Random = summary
			result.Findings.Random = findings
			mu.Unlock()
		}()
	}

	if cfg.TLS.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runTLS(ctx, opts, cfg.TLS)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "tls")
			result.Summary.TLS = summary
			result.Findings.TLS = findings
			mu.Unlock()
		}()
	}

	if cfg.Certificates.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, certResults := s.runCertificates(ctx, opts, cfg.Certificates)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "certificates")
			result.Summary.Certificates = summary
			result.Findings.Certificates = certResults
			mu.Unlock()
		}()
	}

	wg.Wait()

	scanResult := scanner.NewScanResult(Name, Version, start)
	scanResult.Repository = opts.RepoPath
	scanResult.SetSummary(result.Summary)
	scanResult.SetFindings(result.Findings)

	// Build metadata with features_run and git_history_security results
	metadata := map[string]interface{}{
		"features_run": result.FeaturesRun,
	}
	if result.GitHistorySecurity != nil {
		metadata["git_history_security"] = result.GitHistorySecurity
	}
	scanResult.SetMetadata(metadata)

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
// SECRETS FEATURE (Enhanced with entropy, git history, rotation)
// ============================================================================

func (s *CodeSecurityScanner) runSecrets(ctx context.Context, opts *scanner.ScanOptions, cfg SecretsConfig) (*SecretsSummary, []SecretFinding) {
	var allFindings []SecretFinding
	var semgrepFindings []SecretFinding
	var entropyFindings []SecretFinding
	var historyFindings []SecretFinding
	var iacSecretsFindings []SecretFinding

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var semgrepSummary *SecretsSummary

	// Run semgrep secrets detection
	wg.Add(1)
	go func() {
		defer wg.Done()
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

		if err == nil && result != nil {
			mu.Lock()
			semgrepFindings, semgrepSummary = parseSecretsOutput(result.Stdout, opts.RepoPath, cfg)
			// Mark detection source
			for i := range semgrepFindings {
				semgrepFindings[i].DetectionSource = "semgrep"
			}
			mu.Unlock()
		}
	}()

	// Run IaC secrets scanning if enabled
	if cfg.IaCSecrets.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			iacScanner := common.NewIaCSecretsScanner(common.IaCSecretsConfig{
				Timeout: timeout,
			})
			result := iacScanner.Scan(ctx, opts.RepoPath)
			if result.Error == nil {
				mu.Lock()
				// Convert IaC secrets findings to SecretFinding format
				for _, f := range result.Findings {
					finding := SecretFinding{
						RuleID:          f.RuleID,
						Type:            f.SecretType,
						Severity:        f.Severity,
						Message:         f.Message,
						File:            f.File,
						Line:            f.Line,
						Column:          f.Column,
						Snippet:         f.Snippet,
						DetectionSource: "iac-scanner",
						IaCType:         f.Type,
					}
					iacSecretsFindings = append(iacSecretsFindings, finding)
				}
				mu.Unlock()
			}
		}()
	}

	// Run entropy analysis if enabled
	if cfg.EntropyAnalysis.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			analyzer := NewEntropyAnalyzer(cfg.EntropyAnalysis)
			result, err := analyzer.ScanDirectory(opts.RepoPath)
			if err == nil && result != nil {
				mu.Lock()
				entropyFindings = result.Findings
				mu.Unlock()
			}
		}()
	}

	// Run git history scanning if enabled
	if cfg.GitHistoryScan.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := NewGitHistoryScanner(cfg.GitHistoryScan)
			result, err := scanner.ScanRepository(opts.RepoPath)
			if err == nil && result != nil {
				mu.Lock()
				historyFindings = result.Findings
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Initialize summary if semgrep didn't run or failed
	if semgrepSummary == nil {
		semgrepSummary = &SecretsSummary{
			ByType:    make(map[string]int),
			BySource:  make(map[string]int),
			RiskScore: 100,
			RiskLevel: "excellent",
		}
	}
	if semgrepSummary.BySource == nil {
		semgrepSummary.BySource = make(map[string]int)
	}

	// Merge all findings and deduplicate
	allFindings = append(allFindings, semgrepFindings...)
	semgrepSummary.BySource["semgrep"] = len(semgrepFindings)

	// Add entropy findings (dedupe by file:line)
	seen := make(map[string]bool)
	for _, f := range allFindings {
		key := fmt.Sprintf("%s:%d", f.File, f.Line)
		seen[key] = true
	}

	for _, f := range entropyFindings {
		key := fmt.Sprintf("%s:%d", f.File, f.Line)
		if !seen[key] {
			allFindings = append(allFindings, f)
			seen[key] = true
			semgrepSummary.EntropyFindings++
		}
	}
	semgrepSummary.BySource["entropy"] = semgrepSummary.EntropyFindings

	// Add git history findings (dedupe)
	for _, f := range historyFindings {
		key := fmt.Sprintf("%s:%d", f.File, f.Line)
		if !seen[key] {
			allFindings = append(allFindings, f)
			seen[key] = true
			semgrepSummary.HistoryFindings++
			if f.CommitInfo != nil && f.CommitInfo.IsRemoved {
				semgrepSummary.RemovedSecrets++
			}
		}
	}
	semgrepSummary.BySource["git_history"] = semgrepSummary.HistoryFindings

	// Add IaC secrets findings (dedupe)
	iacSecretsCount := 0
	for _, f := range iacSecretsFindings {
		key := fmt.Sprintf("%s:%d", f.File, f.Line)
		if !seen[key] {
			allFindings = append(allFindings, f)
			seen[key] = true
			iacSecretsCount++
		}
	}
	semgrepSummary.BySource["iac_secrets"] = iacSecretsCount

	// Add rotation guidance if enabled
	if cfg.RotationGuidance {
		rotationDB := NewRotationDatabase()
		allFindings = EnrichWithRotation(allFindings, rotationDB)
	}

	// Run AI analysis for false positive detection if enabled
	if cfg.AIAnalysis.Enabled {
		aiAnalyzer := NewAIAnalyzer(cfg.AIAnalysis, opts.RepoPath)
		if aiAnalyzer.IsAvailable() {
			allFindings, _ = aiAnalyzer.AnalyzeFindings(ctx, allFindings)
			// Update summary with AI analysis results
			fp, confirmed := CountFalsePositives(allFindings)
			semgrepSummary.FalsePositives = fp
			semgrepSummary.ConfirmedSecrets = confirmed
		}
	}

	// Recalculate summary stats
	semgrepSummary.TotalFindings = len(allFindings)
	semgrepSummary.Critical = 0
	semgrepSummary.High = 0
	semgrepSummary.Medium = 0
	semgrepSummary.Low = 0
	filesSet := make(map[string]bool)
	semgrepSummary.ByType = make(map[string]int)

	for _, f := range allFindings {
		filesSet[f.File] = true
		semgrepSummary.ByType[f.Type]++
		switch f.Severity {
		case "critical":
			semgrepSummary.Critical++
		case "high":
			semgrepSummary.High++
		case "medium":
			semgrepSummary.Medium++
		case "low":
			semgrepSummary.Low++
		}
	}
	semgrepSummary.FilesAffected = len(filesSet)

	// Recalculate risk score
	penalty := semgrepSummary.Critical*25 + semgrepSummary.High*15 + semgrepSummary.Medium*5 + semgrepSummary.Low*2
	semgrepSummary.RiskScore = 100 - penalty
	if semgrepSummary.RiskScore < 0 {
		semgrepSummary.RiskScore = 0
	}

	switch {
	case semgrepSummary.RiskScore < 40:
		semgrepSummary.RiskLevel = "critical"
	case semgrepSummary.RiskScore < 60:
		semgrepSummary.RiskLevel = "high"
	case semgrepSummary.RiskScore < 80:
		semgrepSummary.RiskLevel = "medium"
	case semgrepSummary.RiskScore < 95:
		semgrepSummary.RiskLevel = "low"
	default:
		semgrepSummary.RiskLevel = "excellent"
	}

	return semgrepSummary, allFindings
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
	var allFindings []APIFinding

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. Run Semgrep-based security checks
	if cfg.CheckAuth || cfg.CheckInjection || cfg.CheckSSRF || cfg.CheckCORS {
		semgrepFindings := s.runSemgrepAPIChecks(ctx, opts, cfg)
		allFindings = append(allFindings, semgrepFindings...)
	}

	// 2. Run RAG pattern-based checks
	ragFindings := s.runRAGAPIPatterns(ctx, opts, cfg)
	allFindings = append(allFindings, ragFindings...)

	// 3. Run OpenAPI validation
	if cfg.CheckOpenAPI {
		openAPIFindings := s.runOpenAPIValidation(ctx, opts)
		allFindings = append(allFindings, openAPIFindings...)
	}

	// 4. Run GraphQL checks
	if cfg.CheckGraphQL {
		graphQLFindings := s.runGraphQLChecks(ctx, opts)
		allFindings = append(allFindings, graphQLFindings...)
	}

	// 5. Run API quality checks (non-security)
	if cfg.CheckDesign || cfg.CheckPerformance || cfg.CheckObservability || cfg.CheckDocumentation {
		qualityFindings := s.runAPIQualityChecks(ctx, opts, cfg)
		allFindings = append(allFindings, qualityFindings...)
	}

	// Build summary from all findings
	summary := buildAPISummary(allFindings)
	return summary, allFindings
}

// runSemgrepAPIChecks runs Semgrep-based API security scanning
func (s *CodeSecurityScanner) runSemgrepAPIChecks(ctx context.Context, opts *scanner.ScanOptions, cfg APIConfig) []APIFinding {
	var findings []APIFinding

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
		return findings
	}

	findings, _ = parseAPIOutput(result.Stdout, opts.RepoPath, cfg)
	return findings
}

// buildAPISummary creates a summary from all API findings
func buildAPISummary(findings []APIFinding) *APISummary {
	summary := &APISummary{
		ByCategory:  make(map[string]int),
		ByOWASPApi:  make(map[string]int),
		ByFramework: make(map[string]int),
	}

	endpoints := make(map[string]bool)

	for _, f := range findings {
		summary.TotalFindings++
		summary.ByCategory[f.Category]++

		if f.OWASPApi != "" {
			summary.ByOWASPApi[f.OWASPApi]++
		}
		if f.Framework != "" {
			summary.ByFramework[f.Framework]++
		}
		if f.Endpoint != "" {
			endpoints[f.Endpoint] = true
		}

		switch f.Severity {
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

	summary.EndpointsFound = len(endpoints)
	return summary
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

	// Order matters - check more specific patterns first
	// Using ordered slice instead of map to ensure consistent matching
	types := []struct {
		secretType string
		patterns   []string
	}{
		{"aws_credential", []string{"aws"}},
		{"github_token", []string{"github"}},
		{"gitlab_token", []string{"gitlab"}},
		{"slack_token", []string{"slack"}},
		{"stripe_key", []string{"stripe"}},
		{"private_key", []string{"private_key", "rsa", "dsa", "ec_private"}},
		{"database_credential", []string{"postgres", "mysql", "mongodb", "redis", "database"}},
		{"jwt_secret", []string{"jwt"}},          // Check jwt before generic secret
		{"api_key", []string{"api_key", "apikey"}},
		{"password", []string{"password"}},
		{"generic_secret", []string{"secret"}},   // Check this last as it's generic
	}

	for _, t := range types {
		for _, pattern := range t.patterns {
			if strings.Contains(ruleIDLower, pattern) {
				return t.secretType
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
