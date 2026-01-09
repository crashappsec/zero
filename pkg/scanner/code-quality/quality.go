// Package codequality provides the consolidated code quality super scanner
// Features: tech_debt, complexity, test_coverage, code_docs
package codequality

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
	"github.com/crashappsec/zero/pkg/scanner/common"
)

const (
	Name    = "code-quality"
	Version = "3.2.0"
)

func init() {
	scanner.Register(&QualityScanner{})
}

// QualityScanner consolidates code quality analysis
type QualityScanner struct{}

func (s *QualityScanner) Name() string {
	return Name
}

func (s *QualityScanner) Description() string {
	return "Code quality analysis: technical debt, complexity, test coverage, documentation"
}

func (s *QualityScanner) Dependencies() []string {
	return nil
}

func (s *QualityScanner) EstimateDuration(fileCount int) time.Duration {
	est := 10 + fileCount/500
	return time.Duration(est) * time.Second
}

func (s *QualityScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	cfg := getConfig(opts)

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Check if semgrep is available (needed for complexity analysis)
	hasSemgrep := common.ToolExists("semgrep")

	// Run features in parallel where possible
	if cfg.TechDebt.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, techDebtResult := s.runTechDebt(ctx, opts, cfg.TechDebt)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "tech_debt")
			result.Summary.TechDebt = summary
			result.Findings.TechDebt = techDebtResult
			mu.Unlock()
		}()
	}

	if cfg.Complexity.Enabled && hasSemgrep {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, complexityResult := s.runComplexity(ctx, opts, cfg.Complexity)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "complexity")
			result.Summary.Complexity = summary
			result.Findings.Complexity = complexityResult
			mu.Unlock()
		}()
	} else if cfg.Complexity.Enabled {
		mu.Lock()
		result.Summary.Errors = append(result.Summary.Errors, "complexity: semgrep not installed")
		mu.Unlock()
	}

	if cfg.TestCoverage.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary := s.runTestCoverage(ctx, opts, cfg.TestCoverage)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "test_coverage")
			result.Summary.TestCoverage = summary
			mu.Unlock()
		}()
	}

	if cfg.CodeDocs.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary := s.runCodeDocs(ctx, opts, cfg.CodeDocs)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "code_docs")
			result.Summary.CodeDocs = summary
			mu.Unlock()
		}()
	}

	wg.Wait()

	scanResult := scanner.NewScanResult(Name, Version, start)
	scanResult.Repository = opts.RepoPath
	_ = scanResult.SetSummary(result.Summary)
	_ = scanResult.SetFindings(result.Findings)
	_ = scanResult.SetMetadata(map[string]interface{}{
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
// TECH DEBT FEATURE
// ============================================================================

func (s *QualityScanner) runTechDebt(ctx context.Context, opts *scanner.ScanOptions, cfg TechDebtConfig) (*TechDebtSummary, *TechDebtResult) {
	var markers []DebtMarker
	var issues []DebtIssue
	fileStats := make(map[string]*FileDebt)

	// Scan for debt markers and issues
	if cfg.IncludeMarkers || cfg.IncludeIssues {
		_ = scanForDebt(opts.RepoPath, &markers, &issues, fileStats, cfg)
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

	return summary, &TechDebtResult{
		Markers:  markers,
		Issues:   issues,
		Hotspots: hotspots,
	}
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

	fileScanner := bufio.NewScanner(file)
	lineNum := 0

	for fileScanner.Scan() {
		lineNum++
		line := fileScanner.Text()

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
// COMPLEXITY FEATURE
// ============================================================================

func (s *QualityScanner) runComplexity(ctx context.Context, opts *scanner.ScanOptions, cfg ComplexityConfig) (*ComplexitySummary, *ComplexityResult) {
	var issues []ComplexityIssue

	timeout := opts.Timeout
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
		opts.RepoPath,
	)

	if err != nil || result == nil {
		return &ComplexitySummary{Error: "semgrep execution failed"}, &ComplexityResult{Issues: issues}
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
		return &ComplexitySummary{Error: "failed to parse semgrep output"}, &ComplexityResult{Issues: issues}
	}

	complexityKeywords := []string{
		"complexity", "long", "nested", "deep", "lines", "parameters",
		"function", "method", "class", "cognitive", "cyclomatic",
	}

	filesAffected := make(map[string]bool)

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
		if strings.HasPrefix(file, opts.RepoPath) {
			file = strings.TrimPrefix(file, opts.RepoPath+"/")
		}

		issueType := categorizeComplexityIssue(r.CheckID, r.Extra.Message)

		issues = append(issues, ComplexityIssue{
			Type:        issueType,
			Severity:    severity,
			File:        file,
			Line:        r.Start.Line,
			Description: r.Extra.Message,
			Suggestion:  getComplexitySuggestion(issueType),
			Source:      "semgrep",
		})

		filesAffected[file] = true
	}

	summary := &ComplexitySummary{
		TotalIssues:   len(issues),
		FilesAffected: len(filesAffected),
		ByType:        make(map[string]int),
	}

	for _, iss := range issues {
		summary.ByType[iss.Type]++
		switch iss.Severity {
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		}
	}

	return summary, &ComplexityResult{Issues: issues}
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
		"complexity-cyclomatic":      "Break down into smaller functions with single responsibilities",
		"complexity-long-function":   "Extract logic into helper functions or separate methods",
		"complexity-deep-nesting":    "Use early returns, guard clauses, or extract nested logic",
		"complexity-too-many-params": "Group related parameters into objects/structs",
		"complexity-cognitive":       "Simplify control flow and reduce cognitive load",
		"complexity-general":         "Consider refactoring to improve maintainability",
	}
	if s, ok := suggestions[issueType]; ok {
		return s
	}
	return "Consider refactoring for better maintainability"
}

// ============================================================================
// TEST COVERAGE FEATURE
// ============================================================================

func (s *QualityScanner) runTestCoverage(ctx context.Context, opts *scanner.ScanOptions, cfg TestCoverageConfig) *TestCoverageSummary {
	summary := &TestCoverageSummary{
		HasTestFiles:    false,
		TestFrameworks:  []string{},
		CoverageReports: []string{},
	}

	// Look for test files and frameworks
	testPatterns := map[string]string{
		"*_test.go":      "go-test",
		"*.test.js":      "jest",
		"*.spec.js":      "jest",
		"*.test.ts":      "jest",
		"*.spec.ts":      "jest",
		"test_*.py":      "pytest",
		"*_test.py":      "pytest",
		"*Test.java":     "junit",
		"*_spec.rb":      "rspec",
		"*.test.tsx":     "jest",
		"*.spec.tsx":     "jest",
	}

	frameworks := make(map[string]bool)

	_ = filepath.Walk(opts.RepoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			// Skip common non-test directories
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == "node_modules" || name == "vendor" || name == ".git" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		fileName := info.Name()
		for pattern, framework := range testPatterns {
			matched, _ := filepath.Match(pattern, fileName)
			if matched {
				summary.HasTestFiles = true
				frameworks[framework] = true
				break
			}
		}

		// Look for coverage reports
		coverageFiles := []string{
			"coverage.xml", "coverage.json", "lcov.info",
			"coverage.out", "cobertura.xml", "jacoco.xml",
		}
		for _, cf := range coverageFiles {
			if fileName == cf {
				relPath := path
				if strings.HasPrefix(path, opts.RepoPath) {
					relPath = strings.TrimPrefix(path, opts.RepoPath+"/")
				}
				summary.CoverageReports = append(summary.CoverageReports, relPath)
			}
		}

		return nil
	})

	for fw := range frameworks {
		summary.TestFrameworks = append(summary.TestFrameworks, fw)
	}

	// Parse coverage if reports exist
	if cfg.ParseReports && len(summary.CoverageReports) > 0 {
		// Parse first available coverage report
		for _, report := range summary.CoverageReports {
			fullPath := filepath.Join(opts.RepoPath, report)
			coverage := parseCoverageReport(fullPath)
			if coverage >= 0 {
				summary.LineCoverage = coverage
				summary.MeetsThreshold = coverage >= float64(cfg.MinimumThreshold)
				break
			}
		}
	}

	return summary
}

func parseCoverageReport(path string) float64 {
	// Simple coverage parsing - returns -1 if not parseable
	data, err := os.ReadFile(path)
	if err != nil {
		return -1
	}

	content := string(data)

	// Try to parse Go coverage
	if strings.HasSuffix(path, ".out") {
		lines := strings.Split(content, "\n")
		var covered, total int
		for _, line := range lines {
			if strings.HasPrefix(line, "mode:") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				total++
				if parts[2] != "0" {
					covered++
				}
			}
		}
		if total > 0 {
			return float64(covered) / float64(total) * 100
		}
	}

	// Try to parse lcov
	if strings.HasSuffix(path, "lcov.info") {
		var lh, lf int
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "LH:") {
				_, _ = fmt.Sscanf(line, "LH:%d", &lh)
			}
			if strings.HasPrefix(line, "LF:") {
				_, _ = fmt.Sscanf(line, "LF:%d", &lf)
			}
		}
		if lf > 0 {
			return float64(lh) / float64(lf) * 100
		}
	}

	return -1
}

// ============================================================================
// CODE DOCS FEATURE
// ============================================================================

func (s *QualityScanner) runCodeDocs(ctx context.Context, opts *scanner.ScanOptions, cfg CodeDocsConfig) *CodeDocsSummary {
	summary := &CodeDocsSummary{
		HasReadme:    false,
		HasChangelog: false,
		HasApiDocs:   false,
		Score:        0,
	}

	// Check for README
	readmeFiles := []string{"README.md", "README.rst", "README.txt", "README", "readme.md"}
	for _, rf := range readmeFiles {
		if _, err := os.Stat(filepath.Join(opts.RepoPath, rf)); err == nil {
			summary.HasReadme = true
			summary.ReadmeFile = rf
			break
		}
	}

	// Check for CHANGELOG
	changelogFiles := []string{"CHANGELOG.md", "CHANGELOG", "HISTORY.md", "CHANGES.md"}
	for _, cf := range changelogFiles {
		if _, err := os.Stat(filepath.Join(opts.RepoPath, cf)); err == nil {
			summary.HasChangelog = true
			break
		}
	}

	// Check for API docs
	apiDocDirs := []string{"docs", "doc", "api", "api-docs"}
	for _, dir := range apiDocDirs {
		if info, err := os.Stat(filepath.Join(opts.RepoPath, dir)); err == nil && info.IsDir() {
			summary.HasApiDocs = true
			break
		}
	}

	// Check for OpenAPI/Swagger
	apiSpecFiles := []string{"openapi.yaml", "openapi.json", "swagger.yaml", "swagger.json"}
	for _, sf := range apiSpecFiles {
		if _, err := os.Stat(filepath.Join(opts.RepoPath, sf)); err == nil {
			summary.HasApiDocs = true
			break
		}
	}

	// Calculate documentation score
	if summary.HasReadme {
		summary.Score += 40
	}
	if summary.HasChangelog {
		summary.Score += 20
	}
	if summary.HasApiDocs {
		summary.Score += 20
	}

	// Check README quality if exists and CheckReadme is enabled
	if cfg.CheckReadme && summary.HasReadme {
		readmePath := filepath.Join(opts.RepoPath, summary.ReadmeFile)
		data, err := os.ReadFile(readmePath)
		if err == nil {
			content := string(data)
			wordCount := len(strings.Fields(content))

			// Basic quality checks
			if wordCount > 100 {
				summary.Score += 10 // Reasonable length
			}
			if strings.Contains(content, "## ") || strings.Contains(content, "# ") {
				summary.Score += 5 // Has headers
			}
			if strings.Contains(strings.ToLower(content), "install") {
				summary.Score += 5 // Has installation info
			}
		}
	}

	if summary.Score > 100 {
		summary.Score = 100
	}

	return summary
}
