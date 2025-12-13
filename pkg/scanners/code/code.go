// Package code provides a consolidated code security super scanner
// Features: vulns, secrets, api, tech-debt
package code

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/scanners/common"
)

const (
	Name    = "code"
	Version = "3.0.0"
)

func init() {
	scanner.Register(&CodeScanner{})
}

// CodeScanner consolidates all code security analysis
type CodeScanner struct{}

func (s *CodeScanner) Name() string {
	return Name
}

func (s *CodeScanner) Description() string {
	return "Consolidated code security scanner: vulnerabilities, secrets, API security, technical debt"
}

func (s *CodeScanner) Dependencies() []string {
	return nil
}

func (s *CodeScanner) EstimateDuration(fileCount int) time.Duration {
	est := 15 + fileCount/300
	return time.Duration(est) * time.Second
}

func (s *CodeScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	cfg := getConfig(opts)

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Check if semgrep is available (needed for vulns, secrets, api)
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

	if cfg.TechDebt.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, techDebtResult := s.runTechDebt(ctx, opts, cfg.TechDebt, hasSemgrep)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "tech_debt")
			result.Summary.TechDebt = summary
			result.Findings.TechDebt = techDebtResult
			mu.Unlock()
		}()
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

func (s *CodeScanner) runVulns(ctx context.Context, opts *scanner.ScanOptions, cfg VulnsConfig) (*VulnsSummary, []VulnFinding) {
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

func (s *CodeScanner) runSecrets(ctx context.Context, opts *scanner.ScanOptions, cfg SecretsConfig) (*SecretsSummary, []SecretFinding) {
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

func (s *CodeScanner) runAPI(ctx context.Context, opts *scanner.ScanOptions, cfg APIConfig) (*APISummary, []APIFinding) {
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
// TECH DEBT FEATURE
// ============================================================================

func (s *CodeScanner) runTechDebt(ctx context.Context, opts *scanner.ScanOptions, cfg TechDebtConfig, hasSemgrep bool) (*TechDebtSummary, *TechDebtResult) {
	var markers []DebtMarker
	var issues []DebtIssue
	fileStats := make(map[string]*FileDebt)

	// Scan for debt markers
	if cfg.IncludeMarkers || cfg.IncludeIssues {
		scanForDebt(opts.RepoPath, &markers, &issues, fileStats, cfg)
	}

	// Use Semgrep for complexity analysis if available and enabled
	if cfg.IncludeComplexity && hasSemgrep {
		complexityIssues := runSemgrepComplexityAnalysis(ctx, opts.RepoPath, opts.Timeout)
		issues = append(issues, complexityIssues...)
	}

	// Calculate hotspots
	var hotspots []FileDebt
	for _, fs := range fileStats {
		if fs.TotalMarkers > 0 {
			hotspots = append(hotspots, *fs)
		}
	}
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].TotalMarkers > hotspots[j].TotalMarkers
	})
	if len(hotspots) > 20 {
		hotspots = hotspots[:20]
	}

	summary := &TechDebtSummary{
		TotalMarkers:  len(markers),
		TotalIssues:   len(issues),
		ByType:        make(map[string]int),
		ByPriority:    make(map[string]int),
		FilesAffected: len(hotspots),
	}

	for _, m := range markers {
		summary.ByType[m.Type]++
		summary.ByPriority[m.Priority]++
	}

	for _, iss := range issues {
		if strings.HasPrefix(iss.Type, "complexity-") {
			summary.ComplexityIssues++
		}
	}

	return summary, &TechDebtResult{
		Markers:  markers,
		Issues:   issues,
		Hotspots: hotspots,
	}
}

func runSemgrepComplexityAnalysis(ctx context.Context, repoPath string, timeout time.Duration) []DebtIssue {
	var issues []DebtIssue

	if timeout == 0 {
		timeout = 3 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := common.RunCommand(ctx, "semgrep",
		"scan",
		"--config", "p/maintainability",
		"--json",
		"--quiet",
		"--include", "*.go",
		"--include", "*.py",
		"--include", "*.js",
		"--include", "*.ts",
		"--include", "*.java",
		repoPath,
	)

	if err != nil || result == nil {
		return issues
	}

	var semgrepOutput struct {
		Results []struct {
			CheckID string `json:"check_id"`
			Path    string `json:"path"`
			Start   struct {
				Line int `json:"line"`
			} `json:"start"`
			Extra struct {
				Message  string `json:"message"`
				Severity string `json:"severity"`
			} `json:"extra"`
		} `json:"results"`
	}

	if err := json.Unmarshal(result.Stdout, &semgrepOutput); err != nil {
		return issues
	}

	complexityKeywords := []string{
		"complexity", "long", "nested", "deep", "lines", "parameters",
		"function", "method", "class", "cognitive", "cyclomatic",
	}

	for _, r := range semgrepOutput.Results {
		checkLower := strings.ToLower(r.CheckID)
		msgLower := strings.ToLower(r.Extra.Message)

		isComplexity := false
		for _, kw := range complexityKeywords {
			if strings.Contains(checkLower, kw) || strings.Contains(msgLower, kw) {
				isComplexity = true
				break
			}
		}

		if !isComplexity {
			continue
		}

		severity := strings.ToLower(r.Extra.Severity)
		switch severity {
		case "warning":
			severity = "medium"
		case "error":
			severity = "high"
		case "info":
			severity = "low"
		}

		file := r.Path
		if strings.HasPrefix(file, repoPath) {
			file = strings.TrimPrefix(file, repoPath+"/")
		}

		issueType := categorizeComplexityIssue(r.CheckID, r.Extra.Message)

		issues = append(issues, DebtIssue{
			Type:        issueType,
			Severity:    severity,
			File:        file,
			Line:        r.Start.Line,
			Description: r.Extra.Message,
			Suggestion:  getComplexitySuggestion(issueType),
			Source:      "semgrep",
		})
	}

	return issues
}

// Marker patterns
var markerPatterns = []struct {
	pattern  *regexp.Regexp
	typ      string
	priority string
}{
	{regexp.MustCompile(`(?i)\bFIXME\b[:\s]*(.{0,100})`), "FIXME", "high"},
	{regexp.MustCompile(`(?i)\bXXX\b[:\s]*(.{0,100})`), "XXX", "high"},
	{regexp.MustCompile(`(?i)\bBUG\b[:\s]*(.{0,100})`), "BUG", "high"},
	{regexp.MustCompile(`(?i)\bHACK\b[:\s]*(.{0,100})`), "HACK", "high"},
	{regexp.MustCompile(`(?i)\bWORKAROUND\b[:\s]*(.{0,100})`), "WORKAROUND", "high"},
	{regexp.MustCompile(`(?i)\bTODO\b[:\s]*(.{0,100})`), "TODO", "medium"},
	{regexp.MustCompile(`(?i)\bREFACTOR\b[:\s]*(.{0,100})`), "REFACTOR", "medium"},
	{regexp.MustCompile(`(?i)\bOPTIMIZE\b[:\s]*(.{0,100})`), "OPTIMIZE", "medium"},
	{regexp.MustCompile(`(?i)\bCLEANUP\b[:\s]*(.{0,100})`), "CLEANUP", "medium"},
	{regexp.MustCompile(`(?i)\bTECH[_-]?DEBT\b[:\s]*(.{0,100})`), "TECH_DEBT", "medium"},
	{regexp.MustCompile(`(?i)\bNOTE\b[:\s]*(.{0,100})`), "NOTE", "low"},
	{regexp.MustCompile(`(?i)\bIDEA\b[:\s]*(.{0,100})`), "IDEA", "low"},
	{regexp.MustCompile(`(?i)\bREVIEW\b[:\s]*(.{0,100})`), "REVIEW", "low"},
	{regexp.MustCompile(`(?i)\bTEMP\b[:\s]*(.{0,100})`), "TEMP", "medium"},
}

// Issue patterns
var issuePatterns = []struct {
	pattern     *regexp.Regexp
	typ         string
	severity    string
	description string
	suggestion  string
}{
	{
		regexp.MustCompile(`(?i)@deprecated`),
		"deprecated-usage", "medium",
		"Deprecated annotation found",
		"Replace with current alternative",
	},
	{
		regexp.MustCompile(`(?i)(noinspection|@suppress|eslint-disable|noqa|nosec)`),
		"suppressed-warning", "low",
		"Linter/analyzer warning suppressed",
		"Address the underlying issue instead of suppressing",
	},
	{
		regexp.MustCompile(`(?i)console\.(log|debug|info|warn|error)\s*\(`),
		"debug-statement", "low",
		"Console/debug statement in code",
		"Remove debug statements or use proper logging",
	},
	{
		regexp.MustCompile(`(?i)(sleep|wait|delay)\s*\(\s*\d+\s*\)`),
		"hardcoded-delay", "medium",
		"Hardcoded delay/sleep found",
		"Use proper async patterns or event-driven approaches",
	},
	{
		regexp.MustCompile(`(?i)catch\s*\([^)]*\)\s*\{\s*\}`),
		"empty-catch", "high",
		"Empty catch block swallows errors",
		"Handle or log errors appropriately",
	},
	{
		regexp.MustCompile(`(?i)(magic\s*number|hardcoded|hard-coded)`),
		"magic-value", "low",
		"Magic number or hardcoded value mentioned",
		"Extract to named constant",
	},
	{
		regexp.MustCompile(`(?i)DISABLED|SKIP|PENDING`),
		"disabled-test", "medium",
		"Disabled/skipped test detected",
		"Fix or remove disabled tests",
	},
	{
		regexp.MustCompile(`(?i)(process\.exit|os\.exit|sys\.exit|System\.exit)`),
		"hard-exit", "medium",
		"Hard process exit call",
		"Use proper error handling and graceful shutdown",
	},
}

var scanExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
	".jsx": true, ".java": true, ".rb": true, ".php": true, ".cs": true,
	".cpp": true, ".c": true, ".h": true, ".hpp": true, ".rs": true,
	".swift": true, ".kt": true, ".scala": true, ".vue": true, ".svelte": true,
}

func scanForDebt(repoPath string, markers *[]DebtMarker, issues *[]DebtIssue, fileStats map[string]*FileDebt, cfg TechDebtConfig) error {
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" ||
				name == "dist" || name == "build" || name == ".venv" ||
				name == "__pycache__" || name == "target" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !scanExtensions[ext] {
			return nil
		}

		scanFileForDebt(path, repoPath, markers, issues, fileStats, cfg)
		return nil
	})
}

func scanFileForDebt(filePath, repoPath string, markers *[]DebtMarker, issues *[]DebtIssue, fileStats map[string]*FileDebt, cfg TechDebtConfig) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	relPath := filePath
	if strings.HasPrefix(filePath, repoPath) {
		relPath = strings.TrimPrefix(filePath, repoPath+"/")
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Look for debt markers
		if cfg.IncludeMarkers {
			for _, mp := range markerPatterns {
				if matches := mp.pattern.FindStringSubmatch(line); len(matches) > 0 {
					text := strings.TrimSpace(line)
					if len(text) > 150 {
						text = text[:150] + "..."
					}

					marker := DebtMarker{
						Type:     mp.typ,
						Priority: mp.priority,
						File:     relPath,
						Line:     lineNum,
						Text:     text,
					}
					*markers = append(*markers, marker)

					if _, ok := fileStats[relPath]; !ok {
						fileStats[relPath] = &FileDebt{
							File:   relPath,
							ByType: make(map[string]int),
						}
					}
					fileStats[relPath].TotalMarkers++
					fileStats[relPath].ByType[mp.typ]++
				}
			}
		}

		// Look for code issues
		if cfg.IncludeIssues {
			for _, ip := range issuePatterns {
				if ip.pattern.MatchString(line) {
					issue := DebtIssue{
						Type:        ip.typ,
						Severity:    ip.severity,
						File:        relPath,
						Line:        lineNum,
						Description: ip.description,
						Suggestion:  ip.suggestion,
						Source:      "pattern",
					}
					*issues = append(*issues, issue)
				}
			}
		}
	}
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

func categorizeComplexityIssue(checkID, message string) string {
	combined := strings.ToLower(checkID + " " + message)

	if strings.Contains(combined, "cyclomatic") || strings.Contains(combined, "complex") {
		return "complexity-cyclomatic"
	}
	if strings.Contains(combined, "long") && strings.Contains(combined, "function") {
		return "complexity-long-function"
	}
	if strings.Contains(combined, "lines") {
		return "complexity-long-function"
	}
	if strings.Contains(combined, "nested") || strings.Contains(combined, "deep") {
		return "complexity-deep-nesting"
	}
	if strings.Contains(combined, "parameter") || strings.Contains(combined, "argument") {
		return "complexity-too-many-params"
	}
	if strings.Contains(combined, "cognitive") {
		return "complexity-cognitive"
	}

	return "complexity-general"
}

func getComplexitySuggestion(issueType string) string {
	suggestions := map[string]string{
		"complexity-cyclomatic":     "Break down into smaller functions with single responsibilities",
		"complexity-long-function":  "Extract logic into helper functions or separate methods",
		"complexity-deep-nesting":   "Use early returns, guard clauses, or extract nested logic",
		"complexity-too-many-params": "Group related parameters into objects/structs",
		"complexity-cognitive":      "Simplify control flow and reduce cognitive load",
		"complexity-general":        "Consider refactoring to improve maintainability",
	}
	if s, ok := suggestions[issueType]; ok {
		return s
	}
	return "Consider refactoring for better maintainability"
}
