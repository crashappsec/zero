// Package health implements the consolidated project health super scanner
// Features: technology, documentation, tests
// Note: Ownership analysis has been moved to the dedicated ownership scanner
package health

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/scanner"
)

const (
	Name    = "health"
	Version = "1.0.0"
)

func init() {
	scanner.Register(&HealthScanner{})
}

// HealthScanner is the consolidated project health scanner
type HealthScanner struct{}

func (s *HealthScanner) Name() string {
	return Name
}

func (s *HealthScanner) Description() string {
	return "Consolidated project health scanner: technology discovery, documentation, test coverage, code ownership"
}

func (s *HealthScanner) Dependencies() []string {
	return nil
}

func (s *HealthScanner) EstimateDuration(fileCount int) time.Duration {
	// Base time plus per-file overhead
	return 30*time.Second + time.Duration(fileCount/100)*time.Second
}

func (s *HealthScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	// Get feature config
	cfg := getFeatureConfig(opts)

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Run technology detection
	if cfg.Technology.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings, err := runTechnologyAnalysis(ctx, opts.RepoPath, cfg.Technology, opts.SBOMPath)
			mu.Lock()
			defer mu.Unlock()
			result.FeaturesRun = append(result.FeaturesRun, "technology")
			if err != nil {
				result.Summary.Errors = append(result.Summary.Errors, fmt.Sprintf("technology: %v", err))
				return
			}
			result.Summary.Technology = summary
			result.Findings.Technology = findings
		}()
	}

	// Run documentation analysis
	if cfg.Documentation.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings, err := runDocumentationAnalysis(ctx, opts.RepoPath, cfg.Documentation)
			mu.Lock()
			defer mu.Unlock()
			result.FeaturesRun = append(result.FeaturesRun, "documentation")
			if err != nil {
				result.Summary.Errors = append(result.Summary.Errors, fmt.Sprintf("documentation: %v", err))
				return
			}
			result.Summary.Documentation = summary
			result.Findings.Documentation = findings
		}()
	}

	// Run test coverage analysis
	if cfg.Tests.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings, err := runTestsAnalysis(ctx, opts.RepoPath, cfg.Tests)
			mu.Lock()
			defer mu.Unlock()
			result.FeaturesRun = append(result.FeaturesRun, "tests")
			if err != nil {
				result.Summary.Errors = append(result.Summary.Errors, fmt.Sprintf("tests: %v", err))
				return
			}
			result.Summary.Tests = summary
			result.Findings.Tests = findings
		}()
	}

	// Note: Ownership analysis has been moved to the dedicated ownership scanner

	wg.Wait()

	// Create scan result
	scanResult := scanner.NewScanResult(Name, Version, start)
	scanResult.Repository = opts.RepoPath
	scanResult.SetSummary(result.Summary)
	scanResult.SetFindings(result)

	// Write output
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

func getFeatureConfig(opts *scanner.ScanOptions) FeatureConfig {
	if opts.FeatureConfig == nil {
		return DefaultConfig()
	}

	cfg := DefaultConfig()

	if techCfg, ok := opts.FeatureConfig["technology"].(map[string]interface{}); ok {
		if enabled, ok := techCfg["enabled"].(bool); ok {
			cfg.Technology.Enabled = enabled
		}
		if v, ok := techCfg["scan_extensions"].(bool); ok {
			cfg.Technology.ScanExtensions = v
		}
		if v, ok := techCfg["scan_config"].(bool); ok {
			cfg.Technology.ScanConfig = v
		}
		if v, ok := techCfg["scan_sbom"].(bool); ok {
			cfg.Technology.ScanSBOM = v
		}
	}

	if docCfg, ok := opts.FeatureConfig["documentation"].(map[string]interface{}); ok {
		if enabled, ok := docCfg["enabled"].(bool); ok {
			cfg.Documentation.Enabled = enabled
		}
		if v, ok := docCfg["check_readme"].(bool); ok {
			cfg.Documentation.CheckReadme = v
		}
		if v, ok := docCfg["check_code_docs"].(bool); ok {
			cfg.Documentation.CheckCodeDocs = v
		}
		if v, ok := docCfg["check_changelog"].(bool); ok {
			cfg.Documentation.CheckChangelog = v
		}
		if v, ok := docCfg["check_api_docs"].(bool); ok {
			cfg.Documentation.CheckAPIDocs = v
		}
	}

	if testsCfg, ok := opts.FeatureConfig["tests"].(map[string]interface{}); ok {
		if enabled, ok := testsCfg["enabled"].(bool); ok {
			cfg.Tests.Enabled = enabled
		}
		if v, ok := testsCfg["parse_reports"].(bool); ok {
			cfg.Tests.ParseReports = v
		}
		if v, ok := testsCfg["analyze_infra"].(bool); ok {
			cfg.Tests.AnalyzeInfra = v
		}
		if v, ok := testsCfg["coverage_threshold"].(float64); ok {
			cfg.Tests.CoverageThreshold = v
		}
	}

	// Note: Ownership config parsing has been moved to pkg/scanners/ownership

	return cfg
}

// =============================================================================
// Technology Detection Feature
// =============================================================================

var configPatterns = map[string]Technology{
	// JavaScript/Node.js
	"package.json":      {Name: "Node.js", Category: "runtime", Confidence: 90, Source: "config"},
	"package-lock.json": {Name: "npm", Category: "package-manager", Confidence: 90, Source: "config"},
	"yarn.lock":         {Name: "Yarn", Category: "package-manager", Confidence: 90, Source: "config"},
	"pnpm-lock.yaml":    {Name: "pnpm", Category: "package-manager", Confidence: 90, Source: "config"},
	"tsconfig.json":     {Name: "TypeScript", Category: "language", Confidence: 95, Source: "config"},
	"next.config.js":    {Name: "Next.js", Category: "framework", Confidence: 95, Source: "config"},
	"nuxt.config.js":    {Name: "Nuxt.js", Category: "framework", Confidence: 95, Source: "config"},
	"angular.json":      {Name: "Angular", Category: "framework", Confidence: 95, Source: "config"},
	"svelte.config.js":  {Name: "Svelte", Category: "framework", Confidence: 95, Source: "config"},
	"jest.config.js":    {Name: "Jest", Category: "testing", Confidence: 90, Source: "config"},

	// Python
	"requirements.txt": {Name: "Python", Category: "language", Confidence: 90, Source: "config"},
	"pyproject.toml":   {Name: "Python", Category: "language", Confidence: 90, Source: "config"},
	"Pipfile":          {Name: "Pipenv", Category: "package-manager", Confidence: 90, Source: "config"},
	"poetry.lock":      {Name: "Poetry", Category: "package-manager", Confidence: 90, Source: "config"},

	// Go
	"go.mod": {Name: "Go", Category: "language", Confidence: 95, Source: "config"},
	"go.sum": {Name: "Go Modules", Category: "package-manager", Confidence: 90, Source: "config"},

	// Rust
	"Cargo.toml": {Name: "Rust", Category: "language", Confidence: 95, Source: "config"},
	"Cargo.lock": {Name: "Cargo", Category: "package-manager", Confidence: 90, Source: "config"},

	// Java/JVM
	"pom.xml":         {Name: "Maven", Category: "build-tool", Confidence: 90, Source: "config"},
	"build.gradle":    {Name: "Gradle", Category: "build-tool", Confidence: 90, Source: "config"},
	"build.gradle.kts": {Name: "Gradle Kotlin", Category: "build-tool", Confidence: 90, Source: "config"},

	// Ruby
	"Gemfile":      {Name: "Ruby", Category: "language", Confidence: 90, Source: "config"},
	"Gemfile.lock": {Name: "Bundler", Category: "package-manager", Confidence: 90, Source: "config"},

	// PHP
	"composer.json": {Name: "PHP", Category: "language", Confidence: 90, Source: "config"},

	// Infrastructure
	"Dockerfile":          {Name: "Docker", Category: "container", Confidence: 95, Source: "config"},
	"docker-compose.yml":  {Name: "Docker Compose", Category: "container", Confidence: 95, Source: "config"},
	"docker-compose.yaml": {Name: "Docker Compose", Category: "container", Confidence: 95, Source: "config"},
	"serverless.yml":      {Name: "Serverless Framework", Category: "iac", Confidence: 95, Source: "config"},
	"Pulumi.yaml":         {Name: "Pulumi", Category: "iac", Confidence: 95, Source: "config"},
}

var extensionMap = map[string]Technology{
	".py":     {Name: "Python", Category: "language", Confidence: 80, Source: "extension"},
	".js":     {Name: "JavaScript", Category: "language", Confidence: 80, Source: "extension"},
	".ts":     {Name: "TypeScript", Category: "language", Confidence: 85, Source: "extension"},
	".tsx":    {Name: "React/TypeScript", Category: "framework", Confidence: 85, Source: "extension"},
	".jsx":    {Name: "React", Category: "framework", Confidence: 85, Source: "extension"},
	".go":     {Name: "Go", Category: "language", Confidence: 85, Source: "extension"},
	".rs":     {Name: "Rust", Category: "language", Confidence: 85, Source: "extension"},
	".java":   {Name: "Java", Category: "language", Confidence: 85, Source: "extension"},
	".kt":     {Name: "Kotlin", Category: "language", Confidence: 85, Source: "extension"},
	".scala":  {Name: "Scala", Category: "language", Confidence: 85, Source: "extension"},
	".rb":     {Name: "Ruby", Category: "language", Confidence: 80, Source: "extension"},
	".php":    {Name: "PHP", Category: "language", Confidence: 80, Source: "extension"},
	".cs":     {Name: "C#", Category: "language", Confidence: 85, Source: "extension"},
	".swift":  {Name: "Swift", Category: "language", Confidence: 85, Source: "extension"},
	".c":      {Name: "C", Category: "language", Confidence: 80, Source: "extension"},
	".cpp":    {Name: "C++", Category: "language", Confidence: 80, Source: "extension"},
	".vue":    {Name: "Vue.js", Category: "framework", Confidence: 90, Source: "extension"},
	".svelte": {Name: "Svelte", Category: "framework", Confidence: 90, Source: "extension"},
	".tf":     {Name: "Terraform", Category: "iac", Confidence: 90, Source: "extension"},
}

func runTechnologyAnalysis(ctx context.Context, repoPath string, cfg TechnologyConfig, sbomPath string) (*TechnologySummary, *TechnologyFindings, error) {
	var techs []Technology

	if cfg.ScanConfig {
		techs = append(techs, detectFromConfigFiles(repoPath)...)
	}

	if cfg.ScanSBOM && sbomPath != "" {
		techs = append(techs, detectFromSBOM(sbomPath)...)
	}

	if cfg.ScanExtensions {
		techs = append(techs, detectFromFileExtensions(repoPath)...)
	}

	// Deduplicate
	techs = consolidateTechnologies(techs)

	summary := buildTechnologySummary(techs)
	findings := &TechnologyFindings{Technologies: techs}

	return summary, findings, nil
}

func detectFromConfigFiles(repoPath string) []Technology {
	var techs []Technology

	for pattern, tech := range configPatterns {
		// Check for directory patterns
		if strings.Contains(pattern, "/") {
			filePath := filepath.Join(repoPath, pattern)
			if _, err := os.Stat(filePath); err == nil {
				techs = append(techs, tech)
			}
			continue
		}

		// Direct file check
		filePath := filepath.Join(repoPath, pattern)
		if _, err := os.Stat(filePath); err == nil {
			techs = append(techs, tech)
		}
	}

	// Check for Terraform files
	if matches, _ := filepath.Glob(filepath.Join(repoPath, "*.tf")); len(matches) > 0 {
		techs = append(techs, Technology{Name: "Terraform", Category: "iac", Confidence: 90, Source: "config"})
	}

	// Check for GitHub Actions
	if _, err := os.Stat(filepath.Join(repoPath, ".github", "workflows")); err == nil {
		techs = append(techs, Technology{Name: "GitHub Actions", Category: "ci-cd", Confidence: 95, Source: "config"})
	}

	// Check for GitLab CI
	if _, err := os.Stat(filepath.Join(repoPath, ".gitlab-ci.yml")); err == nil {
		techs = append(techs, Technology{Name: "GitLab CI", Category: "ci-cd", Confidence: 95, Source: "config"})
	}

	return techs
}

func detectFromSBOM(sbomPath string) []Technology {
	var techs []Technology

	data, err := os.ReadFile(sbomPath)
	if err != nil {
		return techs
	}

	var sbomData struct {
		Components []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"components"`
	}

	if err := json.Unmarshal(data, &sbomData); err != nil {
		return techs
	}

	frameworkPatterns := map[string]Technology{
		"react":      {Name: "React", Category: "framework", Confidence: 95, Source: "sbom"},
		"vue":        {Name: "Vue.js", Category: "framework", Confidence: 95, Source: "sbom"},
		"angular":    {Name: "Angular", Category: "framework", Confidence: 95, Source: "sbom"},
		"express":    {Name: "Express.js", Category: "framework", Confidence: 95, Source: "sbom"},
		"django":     {Name: "Django", Category: "framework", Confidence: 95, Source: "sbom"},
		"flask":      {Name: "Flask", Category: "framework", Confidence: 95, Source: "sbom"},
		"fastapi":    {Name: "FastAPI", Category: "framework", Confidence: 95, Source: "sbom"},
		"postgres":   {Name: "PostgreSQL", Category: "database", Confidence: 85, Source: "sbom"},
		"mysql":      {Name: "MySQL", Category: "database", Confidence: 85, Source: "sbom"},
		"mongodb":    {Name: "MongoDB", Category: "database", Confidence: 85, Source: "sbom"},
		"redis":      {Name: "Redis", Category: "database", Confidence: 85, Source: "sbom"},
		"aws-sdk":    {Name: "AWS SDK", Category: "cloud", Confidence: 90, Source: "sbom"},
		"openai":     {Name: "OpenAI", Category: "ai", Confidence: 95, Source: "sbom"},
		"anthropic":  {Name: "Anthropic", Category: "ai", Confidence: 95, Source: "sbom"},
		"langchain":  {Name: "LangChain", Category: "ai", Confidence: 95, Source: "sbom"},
		"tensorflow": {Name: "TensorFlow", Category: "ai", Confidence: 95, Source: "sbom"},
		"pytorch":    {Name: "PyTorch", Category: "ai", Confidence: 95, Source: "sbom"},
	}

	seen := make(map[string]bool)
	for _, comp := range sbomData.Components {
		nameLower := strings.ToLower(comp.Name)
		for pattern, tech := range frameworkPatterns {
			if strings.Contains(nameLower, pattern) && !seen[tech.Name] {
				t := tech
				t.Version = comp.Version
				techs = append(techs, t)
				seen[tech.Name] = true
			}
		}
	}

	return techs
}

func detectFromFileExtensions(repoPath string) []Technology {
	extCounts := make(map[string]int)

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == "vendor" || base == ".git" ||
				base == "dist" || base == "build" || base == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != "" {
			extCounts[ext]++
		}
		return nil
	})

	var techs []Technology
	for ext, count := range extCounts {
		if tech, ok := extensionMap[ext]; ok && count >= 3 {
			techs = append(techs, tech)
		}
	}

	return techs
}

func consolidateTechnologies(techs []Technology) []Technology {
	techMap := make(map[string]Technology)
	for _, t := range techs {
		existing, ok := techMap[t.Name]
		if !ok || t.Confidence > existing.Confidence {
			techMap[t.Name] = t
		}
	}

	var result []Technology
	for _, t := range techMap {
		result = append(result, t)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Confidence > result[j].Confidence
	})

	return result
}

func buildTechnologySummary(techs []Technology) *TechnologySummary {
	summary := &TechnologySummary{
		TotalTechnologies: len(techs),
		ByCategory:        make(map[string]int),
	}

	for _, t := range techs {
		summary.ByCategory[t.Category]++

		switch t.Category {
		case "language":
			summary.PrimaryLanguages = append(summary.PrimaryLanguages, t.Name)
		case "framework":
			summary.Frameworks = append(summary.Frameworks, t.Name)
		case "database":
			summary.Databases = append(summary.Databases, t.Name)
		case "cloud":
			summary.CloudServices = append(summary.CloudServices, t.Name)
		}
	}

	return summary
}

// =============================================================================
// Documentation Analysis Feature
// =============================================================================

func runDocumentationAnalysis(ctx context.Context, repoPath string, cfg DocumentationConfig) (*DocumentationSummary, *DocumentationFindings, error) {
	projectDocs := analyzeProjectDocs(repoPath, cfg)
	codeDocs := CodeDocumentation{ByLanguage: make(map[string]LangDocs)}

	if cfg.CheckCodeDocs {
		codeDocs = analyzeCodeDocs(repoPath)
	}

	overallScore := calculateDocScore(projectDocs, codeDocs)

	summary := &DocumentationSummary{
		OverallScore:       overallScore,
		HasReadme:          projectDocs.HasReadme,
		HasChangelog:       projectDocs.HasChangelog,
		HasContributing:    projectDocs.HasContributing,
		HasLicense:         projectDocs.HasLicense,
		HasAPIDocs:         projectDocs.HasAPIDocs,
		DocumentedFiles:    codeDocs.DocumentedFiles,
		TotalSourceFiles:   codeDocs.TotalFiles,
		DocumentationRatio: codeDocs.DocumentationRatio,
	}

	findings := &DocumentationFindings{
		ProjectDocs: projectDocs,
		CodeDocs:    codeDocs,
		Issues:      identifyDocIssues(projectDocs, codeDocs),
	}

	return summary, findings, nil
}

func analyzeProjectDocs(repoPath string, cfg DocumentationConfig) ProjectDocumentation {
	docs := ProjectDocumentation{}

	if cfg.CheckReadme {
		readmePatterns := []string{"README.md", "README.rst", "README.txt", "README", "readme.md"}
		for _, pattern := range readmePatterns {
			path := filepath.Join(repoPath, pattern)
			if data, err := os.ReadFile(path); err == nil {
				docs.HasReadme = true
				docs.ReadmeQuality = analyzeReadme(string(data))
				docs.DocumentationFiles = append(docs.DocumentationFiles, DocFile{
					Path:      pattern,
					Type:      "readme",
					WordCount: countWords(string(data)),
				})
				break
			}
		}
	}

	if cfg.CheckChangelog {
		changelogPatterns := []string{"CHANGELOG.md", "CHANGELOG", "HISTORY.md", "CHANGES.md", "NEWS.md"}
		for _, pattern := range changelogPatterns {
			if _, err := os.Stat(filepath.Join(repoPath, pattern)); err == nil {
				docs.HasChangelog = true
				break
			}
		}
	}

	// Check for CONTRIBUTING
	contributingPatterns := []string{"CONTRIBUTING.md", "CONTRIBUTING", ".github/CONTRIBUTING.md"}
	for _, pattern := range contributingPatterns {
		if _, err := os.Stat(filepath.Join(repoPath, pattern)); err == nil {
			docs.HasContributing = true
			break
		}
	}

	// Check for LICENSE
	licensePatterns := []string{"LICENSE", "LICENSE.md", "LICENSE.txt", "COPYING"}
	for _, pattern := range licensePatterns {
		if _, err := os.Stat(filepath.Join(repoPath, pattern)); err == nil {
			docs.HasLicense = true
			break
		}
	}

	if cfg.CheckAPIDocs {
		apiDocPatterns := []string{
			"docs/api", "api-docs", "apidoc",
			"swagger.json", "swagger.yaml", "openapi.json", "openapi.yaml",
		}
		for _, pattern := range apiDocPatterns {
			if _, err := os.Stat(filepath.Join(repoPath, pattern)); err == nil {
				docs.HasAPIDocs = true
				break
			}
		}
	}

	// Check for architecture documentation
	archDocPatterns := []string{"docs/architecture", "ARCHITECTURE.md", "docs/design", "docs/adr"}
	for _, pattern := range archDocPatterns {
		if _, err := os.Stat(filepath.Join(repoPath, pattern)); err == nil {
			docs.HasArchitectureDocs = true
			break
		}
	}

	return docs
}

func analyzeReadme(content string) ReadmeAnalysis {
	analysis := ReadmeAnalysis{
		WordCount: countWords(content),
	}

	contentLower := strings.ToLower(content)

	analysis.HasInstallation = strings.Contains(contentLower, "install") ||
		strings.Contains(contentLower, "getting started")
	analysis.HasUsage = strings.Contains(contentLower, "usage") ||
		strings.Contains(contentLower, "how to use")
	analysis.HasExamples = strings.Contains(contentLower, "example") ||
		strings.Contains(content, "```")
	analysis.HasBadges = strings.Contains(content, "![") ||
		strings.Contains(content, "[![")
	analysis.HasTableOfContents = strings.Contains(contentLower, "table of contents") ||
		strings.Contains(contentLower, "## contents")

	expectedSections := map[string]bool{
		"installation": analysis.HasInstallation,
		"usage":        analysis.HasUsage,
		"examples":     analysis.HasExamples,
	}

	for section, present := range expectedSections {
		if !present {
			analysis.MissingSections = append(analysis.MissingSections, section)
		}
	}

	return analysis
}

func analyzeCodeDocs(repoPath string) CodeDocumentation {
	codeDocs := CodeDocumentation{
		ByLanguage: make(map[string]LangDocs),
	}

	langExtensions := map[string]string{
		".go": "go", ".py": "python", ".js": "javascript", ".ts": "typescript",
		".java": "java", ".rb": "ruby", ".rs": "rust", ".cpp": "cpp", ".c": "c",
	}

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" ||
					name == "dist" || name == "build" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		lang, ok := langExtensions[ext]
		if !ok {
			return nil
		}

		// Skip test files
		if strings.Contains(path, "_test.") || strings.Contains(path, ".test.") ||
			strings.Contains(path, ".spec.") || strings.Contains(path, "test_") {
			return nil
		}

		codeDocs.TotalFiles++
		langDocs := codeDocs.ByLanguage[lang]
		langDocs.TotalFiles++

		hasDoc := analyzeFileDocumentation(path, lang)
		if hasDoc {
			codeDocs.DocumentedFiles++
			langDocs.DocumentedFiles++
		}

		codeDocs.ByLanguage[lang] = langDocs

		return nil
	})

	if codeDocs.TotalFiles > 0 {
		codeDocs.DocumentationRatio = float64(codeDocs.DocumentedFiles) / float64(codeDocs.TotalFiles) * 100
	}

	for lang, langDocs := range codeDocs.ByLanguage {
		if langDocs.TotalFiles > 0 {
			langDocs.Ratio = float64(langDocs.DocumentedFiles) / float64(langDocs.TotalFiles) * 100
			codeDocs.ByLanguage[lang] = langDocs
		}
	}

	return codeDocs
}

func analyzeFileDocumentation(filePath, lang string) bool {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	content := string(data)

	switch lang {
	case "go":
		return strings.HasPrefix(strings.TrimSpace(content), "// Package") ||
			strings.HasPrefix(strings.TrimSpace(content), "/*")
	case "python":
		return strings.HasPrefix(strings.TrimSpace(content), "\"\"\"") ||
			strings.HasPrefix(strings.TrimSpace(content), "'''")
	case "javascript", "typescript":
		minLen := 500
		if len(content) < minLen {
			minLen = len(content)
		}
		return strings.Contains(content[:minLen], "/**") ||
			strings.Contains(content[:minLen], "/*")
	case "java":
		minLen := 500
		if len(content) < minLen {
			minLen = len(content)
		}
		return strings.Contains(content[:minLen], "/**")
	default:
		return strings.HasPrefix(strings.TrimSpace(content), "//") ||
			strings.HasPrefix(strings.TrimSpace(content), "/*") ||
			strings.HasPrefix(strings.TrimSpace(content), "#")
	}
}

func countWords(text string) int {
	return len(strings.Fields(text))
}

func calculateDocScore(projectDocs ProjectDocumentation, codeDocs CodeDocumentation) float64 {
	score := 0.0

	// Project documentation (40 points)
	if projectDocs.HasReadme {
		score += 15
		if projectDocs.ReadmeQuality.HasInstallation {
			score += 3
		}
		if projectDocs.ReadmeQuality.HasUsage {
			score += 3
		}
		if projectDocs.ReadmeQuality.HasExamples {
			score += 3
		}
		if projectDocs.ReadmeQuality.WordCount > 500 {
			score += 3
		}
	}
	if projectDocs.HasChangelog {
		score += 5
	}
	if projectDocs.HasContributing {
		score += 5
	}
	if projectDocs.HasLicense {
		score += 3
	}

	// Code documentation (40 points)
	if codeDocs.DocumentationRatio > 80 {
		score += 40
	} else if codeDocs.DocumentationRatio > 60 {
		score += 30
	} else if codeDocs.DocumentationRatio > 40 {
		score += 20
	} else if codeDocs.DocumentationRatio > 20 {
		score += 10
	}

	// API docs (10 points)
	if projectDocs.HasAPIDocs {
		score += 10
	}

	// Architecture docs (10 points)
	if projectDocs.HasArchitectureDocs {
		score += 10
	}

	return score
}

func identifyDocIssues(projectDocs ProjectDocumentation, codeDocs CodeDocumentation) []DocIssue {
	var issues []DocIssue

	if !projectDocs.HasReadme {
		issues = append(issues, DocIssue{
			Type:        "missing-readme",
			Severity:    "critical",
			Description: "No README file found",
			Suggestion:  "Add a README.md with project description, installation, and usage instructions",
		})
	} else {
		if !projectDocs.ReadmeQuality.HasInstallation {
			issues = append(issues, DocIssue{
				Type:        "incomplete-readme",
				Severity:    "medium",
				Description: "README missing installation instructions",
				Suggestion:  "Add installation/setup instructions to README",
			})
		}
		if !projectDocs.ReadmeQuality.HasUsage {
			issues = append(issues, DocIssue{
				Type:        "incomplete-readme",
				Severity:    "medium",
				Description: "README missing usage documentation",
				Suggestion:  "Add usage examples to README",
			})
		}
	}

	if !projectDocs.HasLicense {
		issues = append(issues, DocIssue{
			Type:        "missing-license",
			Severity:    "high",
			Description: "No LICENSE file found",
			Suggestion:  "Add a LICENSE file to specify usage terms",
		})
	}

	if codeDocs.DocumentationRatio < 50 {
		issues = append(issues, DocIssue{
			Type:        "low-code-docs",
			Severity:    "medium",
			Description: fmt.Sprintf("Only %.1f%% of source files have documentation", codeDocs.DocumentationRatio),
			Suggestion:  "Add documentation comments to public functions and types",
		})
	}

	return issues
}

// =============================================================================
// Test Coverage Analysis Feature
// =============================================================================

func runTestsAnalysis(ctx context.Context, repoPath string, cfg TestsConfig) (*TestsSummary, *TestsFindings, error) {
	coverage := CoverageData{ByDirectory: make(map[string]float64)}

	if cfg.ParseReports {
		coverageFiles := findCoverageReports(repoPath)
		for _, cf := range coverageFiles {
			parseCoverageReport(cf, &coverage)
		}
	}

	// Estimate from tests if no reports found
	if coverage.OverallCoverage == 0 {
		estimateCoverageFromTests(repoPath, &coverage)
	}

	infra := TestInfrastructure{}
	if cfg.AnalyzeInfra {
		infra = analyzeTestInfrastructure(repoPath)
	}

	summary := &TestsSummary{
		OverallCoverage:   coverage.OverallCoverage,
		LineCoverage:      coverage.LineCoverage,
		BranchCoverage:    coverage.BranchCoverage,
		TotalFiles:        coverage.TotalFiles,
		CoveredFiles:      coverage.CoveredFiles,
		UncoveredFiles:    len(coverage.UncoveredFiles),
		TestFramework:     infra.Framework,
		TotalTests:        infra.TotalTests,
		HasCoverageConfig: infra.HasCoverageConfig,
		CoverageThreshold: cfg.CoverageThreshold,
	}

	findings := &TestsFindings{
		Coverage:       coverage,
		Infrastructure: infra,
		Issues:         identifyCoverageIssues(coverage, infra, cfg.CoverageThreshold),
	}

	return summary, findings, nil
}

func findCoverageReports(repoPath string) []string {
	var reports []string

	patterns := []string{
		"coverage.json", "coverage.xml", "coverage.lcov", "lcov.info",
		"cover.out", "coverage.out", ".coverage",
		"coverage/lcov.info", "coverage/coverage-final.json",
	}

	for _, pattern := range patterns {
		path := filepath.Join(repoPath, pattern)
		if _, err := os.Stat(path); err == nil {
			reports = append(reports, path)
		}
	}

	return reports
}

func parseCoverageReport(reportPath string, coverage *CoverageData) {
	ext := strings.ToLower(filepath.Ext(reportPath))
	base := filepath.Base(reportPath)

	switch {
	case ext == ".json":
		parseCoverageJSON(reportPath, coverage)
	case ext == ".lcov" || base == "lcov.info":
		parseLCOV(reportPath, coverage)
	case ext == ".out" || strings.HasSuffix(base, "cover.out"):
		parseGoCoverage(reportPath, coverage)
	}
}

func parseCoverageJSON(reportPath string, coverage *CoverageData) {
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return
	}

	var istanbulCov map[string]struct {
		S map[string]int `json:"s"`
	}

	if err := json.Unmarshal(data, &istanbulCov); err == nil && len(istanbulCov) > 0 {
		var totalStatements, coveredStatements int
		for _, fileCov := range istanbulCov {
			for _, count := range fileCov.S {
				totalStatements++
				if count > 0 {
					coveredStatements++
				}
			}
		}
		if totalStatements > 0 {
			coverage.TotalFiles = len(istanbulCov)
			coverage.LineCoverage = float64(coveredStatements) / float64(totalStatements) * 100
			coverage.OverallCoverage = coverage.LineCoverage
		}
	}
}

func parseLCOV(reportPath string, coverage *CoverageData) {
	file, err := os.Open(reportPath)
	if err != nil {
		return
	}
	defer file.Close()

	var totalLines, coveredLines int
	var totalBranches, coveredBranches int
	currentFile := ""
	filesWithCoverage := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "SF:") {
			currentFile = strings.TrimPrefix(line, "SF:")
		} else if strings.HasPrefix(line, "DA:") {
			parts := strings.Split(strings.TrimPrefix(line, "DA:"), ",")
			if len(parts) >= 2 {
				totalLines++
				if count, _ := strconv.Atoi(parts[1]); count > 0 {
					coveredLines++
					filesWithCoverage[currentFile] = true
				}
			}
		} else if strings.HasPrefix(line, "BRDA:") {
			totalBranches++
			parts := strings.Split(strings.TrimPrefix(line, "BRDA:"), ",")
			if len(parts) >= 4 && parts[3] != "-" && parts[3] != "0" {
				coveredBranches++
			}
		}
	}

	if totalLines > 0 {
		coverage.LineCoverage = float64(coveredLines) / float64(totalLines) * 100
		coverage.OverallCoverage = coverage.LineCoverage
		coverage.CoveredFiles = len(filesWithCoverage)
	}
	if totalBranches > 0 {
		coverage.BranchCoverage = float64(coveredBranches) / float64(totalBranches) * 100
	}
}

func parseGoCoverage(reportPath string, coverage *CoverageData) {
	file, err := os.Open(reportPath)
	if err != nil {
		return
	}
	defer file.Close()

	var totalStatements, coveredStatements int
	filesWithCoverage := make(map[string]bool)

	coveragePattern := regexp.MustCompile(`^(.+):(\d+)\.(\d+),(\d+)\.(\d+)\s+(\d+)\s+(\d+)$`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}

		matches := coveragePattern.FindStringSubmatch(line)
		if len(matches) == 8 {
			statements, _ := strconv.Atoi(matches[6])
			count, _ := strconv.Atoi(matches[7])
			totalStatements += statements
			if count > 0 {
				coveredStatements += statements
				filesWithCoverage[matches[1]] = true
			}
		}
	}

	if totalStatements > 0 {
		coverage.LineCoverage = float64(coveredStatements) / float64(totalStatements) * 100
		coverage.OverallCoverage = coverage.LineCoverage
		coverage.CoveredFiles = len(filesWithCoverage)
	}
}

func estimateCoverageFromTests(repoPath string, coverage *CoverageData) {
	var sourceFiles, testFiles int

	testPatterns := []string{"_test.go", ".test.js", ".test.ts", ".spec.js", ".spec.ts", "_test.py", "test_"}

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py" ||
			ext == ".java" || ext == ".rb" || ext == ".rs" {

			isTest := false
			for _, pattern := range testPatterns {
				if strings.Contains(path, pattern) {
					isTest = true
					break
				}
			}

			if isTest {
				testFiles++
			} else {
				sourceFiles++
			}
		}
		return nil
	})

	coverage.TotalFiles = sourceFiles
	if sourceFiles > 0 && testFiles > 0 {
		estimatedCoverage := float64(testFiles) / float64(sourceFiles) * 50
		if estimatedCoverage > 50 {
			estimatedCoverage = 50
		}
		coverage.OverallCoverage = estimatedCoverage
		coverage.LineCoverage = estimatedCoverage
	}
}

func analyzeTestInfrastructure(repoPath string) TestInfrastructure {
	infra := TestInfrastructure{}

	// Detect framework
	if _, err := os.Stat(filepath.Join(repoPath, "jest.config.js")); err == nil {
		infra.Framework = "jest"
	} else if _, err := os.Stat(filepath.Join(repoPath, "pytest.ini")); err == nil {
		infra.Framework = "pytest"
	} else if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err == nil {
		infra.Framework = "go test"
	} else if _, err := os.Stat(filepath.Join(repoPath, "Cargo.toml")); err == nil {
		infra.Framework = "cargo test"
	}

	// Check for coverage configuration
	coverageConfigs := []string{
		".nycrc", ".nycrc.json", "nyc.config.js",
		"jest.config.js", "jest.config.ts",
		".coveragerc", "setup.cfg", "pyproject.toml",
		"codecov.yml", ".codecov.yml",
	}

	for _, config := range coverageConfigs {
		if _, err := os.Stat(filepath.Join(repoPath, config)); err == nil {
			infra.HasCoverageConfig = true
			break
		}
	}

	// Check for CI integration
	ciFiles := map[string]string{
		".github/workflows":    "github-actions",
		".gitlab-ci.yml":       "gitlab-ci",
		"Jenkinsfile":          "jenkins",
		".circleci/config.yml": "circleci",
		".travis.yml":          "travis",
	}

	for path, ci := range ciFiles {
		fullPath := filepath.Join(repoPath, path)
		if _, err := os.Stat(fullPath); err == nil {
			infra.CIIntegration = append(infra.CIIntegration, ci)
		}
	}

	// Count test files and estimate tests
	testPatterns := []string{"_test.go", ".test.js", ".test.ts", ".spec.js", ".spec.ts", "_test.py", "test_"}

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		for _, pattern := range testPatterns {
			if strings.Contains(path, pattern) {
				infra.TestFiles++
				infra.TotalTests += estimateTestsInFile(path)
				break
			}
		}
		return nil
	})

	return infra
}

func estimateTestsInFile(testFile string) int {
	data, err := os.ReadFile(testFile)
	if err != nil {
		return 0
	}

	content := string(data)
	count := 0

	testPatterns := []*regexp.Regexp{
		regexp.MustCompile(`func\s+Test\w+\s*\(`),
		regexp.MustCompile(`(?:it|test|describe)\s*\(\s*['"]`),
		regexp.MustCompile(`def\s+test_\w+\s*\(`),
		regexp.MustCompile(`@Test`),
	}

	for _, pattern := range testPatterns {
		count += len(pattern.FindAllString(content, -1))
	}

	return count
}

func identifyCoverageIssues(coverage CoverageData, infra TestInfrastructure, threshold float64) []TestIssue {
	var issues []TestIssue

	if coverage.OverallCoverage < 50 {
		issues = append(issues, TestIssue{
			Type:        "low-coverage",
			Severity:    "high",
			Description: fmt.Sprintf("Overall test coverage is %.1f%% (below 50%%)", coverage.OverallCoverage),
			Suggestion:  "Add tests for critical paths and untested files",
		})
	} else if coverage.OverallCoverage < threshold {
		issues = append(issues, TestIssue{
			Type:        "moderate-coverage",
			Severity:    "medium",
			Description: fmt.Sprintf("Overall test coverage is %.1f%% (below %.0f%%)", coverage.OverallCoverage, threshold),
			Suggestion:  fmt.Sprintf("Aim for %.0f%%+ coverage for production code", threshold),
		})
	}

	if !infra.HasCoverageConfig {
		issues = append(issues, TestIssue{
			Type:        "no-coverage-config",
			Severity:    "low",
			Description: "No coverage configuration found",
			Suggestion:  "Add coverage configuration to enforce coverage thresholds",
		})
	}

	if infra.TotalTests == 0 {
		issues = append(issues, TestIssue{
			Type:        "no-tests",
			Severity:    "critical",
			Description: "No test files found in the repository",
			Suggestion:  "Add unit tests to ensure code quality",
		})
	}

	if len(infra.CIIntegration) == 0 {
		issues = append(issues, TestIssue{
			Type:        "no-ci",
			Severity:    "medium",
			Description: "No CI/CD configuration found",
			Suggestion:  "Set up CI to run tests automatically on each commit",
		})
	}

	return issues
}

// =============================================================================
// Code Ownership Analysis Feature (moved to ownership scanner)
// =============================================================================
// Note: Ownership analysis has been extracted to pkg/scanners/ownership
// The health scanner now focuses on technology, documentation, and tests only.
