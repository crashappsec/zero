// Package devex provides the consolidated developer experience super scanner
// Features: onboarding, tooling, workflow
package devx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/scanner"
)

const (
	Name    = "devx"
	Version = "1.0.0"
)

func init() {
	scanner.Register(&DevXScanner{})
}

// DevXScanner consolidates developer experience analysis
type DevXScanner struct{}

func (s *DevXScanner) Name() string {
	return Name
}

func (s *DevXScanner) Description() string {
	return "Developer experience analysis: onboarding friction, tooling complexity, workflow efficiency"
}

func (s *DevXScanner) Dependencies() []string {
	return []string{"tech-id"}
}

func (s *DevXScanner) EstimateDuration(fileCount int) time.Duration {
	est := 5 + fileCount/1000
	return time.Duration(est) * time.Second
}

func (s *DevXScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	cfg := getConfig(opts)

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Run features in parallel
	if cfg.Onboarding.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runOnboarding(ctx, opts, cfg.Onboarding)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "onboarding")
			result.Summary.Onboarding = summary
			result.Findings.Onboarding = findings
			mu.Unlock()
		}()
	}

	if cfg.Sprawl.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runSprawl(ctx, opts, cfg.Sprawl)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "sprawl")
			result.Summary.Sprawl = summary
			result.Findings.Sprawl = findings
			mu.Unlock()
		}()
	}

	if cfg.Workflow.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runWorkflow(ctx, opts, cfg.Workflow)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "workflow")
			result.Summary.Workflow = summary
			result.Findings.Workflow = findings
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

		// Generate markdown reports
		if err := WriteReports(opts.OutputDir); err != nil {
			// Non-fatal: log but don't fail the scan
			fmt.Fprintf(os.Stderr, "Warning: failed to generate reports: %v\n", err)
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
// ONBOARDING FEATURE
// ============================================================================

// configFileInfo holds tool and category information for a config file
type configFileInfo struct {
	tool     string
	category string
}

// getConfigFilePatterns returns a mapping of config file names to tool/category info.
// This is a minimal set of patterns for onboarding analysis - the tech-id scanner
// provides comprehensive technology detection via semgrep rules.
var configFilePatterns = map[string]configFileInfo{
	// Package managers (essential for dependency counting)
	"package.json":      {"npm/yarn", "package-manager"},
	"package-lock.json": {"npm", "package-manager"},
	"yarn.lock":         {"yarn", "package-manager"},
	"pnpm-lock.yaml":    {"pnpm", "package-manager"},
	"go.mod":            {"go", "package-manager"},
	"Cargo.toml":        {"cargo", "package-manager"},
	"requirements.txt":  {"pip", "package-manager"},
	"pyproject.toml":    {"python", "package-manager"},
	"Gemfile":           {"bundler", "package-manager"},
	"composer.json":     {"composer", "package-manager"},

	// Build essentials
	"Makefile":       {"make", "build"},
	"Dockerfile":     {"docker", "container"},
	"tsconfig.json":  {"typescript", "compiler"},

	// CI/CD (for build step estimation)
	".github/workflows": {"github-actions", "ci"},
	".gitlab-ci.yml":    {"gitlab-ci", "ci"},
}

func (s *DevXScanner) runOnboarding(ctx context.Context, opts *scanner.ScanOptions, cfg OnboardingConfig) (*OnboardingSummary, *OnboardingFindings) {
	summary := &OnboardingSummary{
		MissingDocs: []string{},
	}
	findings := &OnboardingFindings{
		ConfigFiles:    []ConfigFile{},
		Prerequisites:  []Prerequisite{},
		EnvVariables:   []EnvVariable{},
		SetupBarriers:  []SetupBarrier{},
	}

	// Scan for config files
	configFiles := scanConfigFiles(opts.RepoPath)
	findings.ConfigFiles = configFiles
	summary.ConfigFileCount = len(configFiles)

	// Count dependencies
	summary.DependencyCount = countDependencies(opts.RepoPath)

	// Estimate build steps
	summary.BuildStepCount = estimateBuildSteps(opts.RepoPath, configFiles)

	// Check for contribution docs
	if cfg.CheckContributing {
		if fileExists(filepath.Join(opts.RepoPath, "CONTRIBUTING.md")) ||
			fileExists(filepath.Join(opts.RepoPath, "CONTRIBUTING")) ||
			fileExists(filepath.Join(opts.RepoPath, ".github/CONTRIBUTING.md")) {
			summary.HasContributing = true
		} else {
			summary.MissingDocs = append(summary.MissingDocs, "CONTRIBUTING.md")
		}
	}

	// Check for env example
	if cfg.CheckEnvSetup {
		envFiles := []string{".env.example", ".env.sample", ".env.template", "env.example"}
		for _, ef := range envFiles {
			if fileExists(filepath.Join(opts.RepoPath, ef)) {
				summary.HasEnvExample = true
				// Parse env variables
				findings.EnvVariables = parseEnvFile(filepath.Join(opts.RepoPath, ef))
				summary.EnvVarCount = len(findings.EnvVariables)
				break
			}
		}
		if !summary.HasEnvExample {
			// Check docker-compose for env vars
			dcPath := filepath.Join(opts.RepoPath, "docker-compose.yml")
			if !fileExists(dcPath) {
				dcPath = filepath.Join(opts.RepoPath, "docker-compose.yaml")
			}
			if fileExists(dcPath) {
				envVars := parseDockerComposeEnv(dcPath)
				findings.EnvVariables = envVars
				summary.EnvVarCount = len(envVars)
			}
		}
	}

	// Prerequisites are now derived from tech-id scanner
	// We still detect them locally as a fallback
	prereqs := detectPrerequisites(opts.RepoPath)
	findings.Prerequisites = prereqs
	summary.PrerequisiteCount = len(prereqs)

	// Analyze README quality
	if cfg.CheckReadmeQuality {
		readmeAnalysis := analyzeReadme(opts.RepoPath)
		findings.ReadmeAnalysis = readmeAnalysis
		if readmeAnalysis != nil {
			summary.ReadmeQualityScore = calculateReadmeScore(readmeAnalysis)
		}
	}

	// Identify setup barriers
	findings.SetupBarriers = identifySetupBarriers(summary, findings)

	// Calculate overall score (0-100, higher is easier to onboard)
	summary.Score = calculateOnboardingScore(summary)
	summary.SetupComplexity = getComplexityLevel(summary.Score)

	return summary, findings
}

func scanConfigFiles(repoPath string) []ConfigFile {
	var configs []ConfigFile

	// Check root directory
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return configs
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if info, ok := configFilePatterns[name]; ok {
			cf := ConfigFile{
				Path:     name,
				Type:     info.category,
				Tool:     info.tool,
			}

			// Count lines
			if data, err := os.ReadFile(filepath.Join(repoPath, name)); err == nil {
				cf.LineCount = len(strings.Split(string(data), "\n"))
				cf.Complexity = getConfigComplexity(cf.LineCount)
			}

			configs = append(configs, cf)
		}
	}

	// Check .github directory
	githubDir := filepath.Join(repoPath, ".github")
	if info, err := os.Stat(githubDir); err == nil && info.IsDir() {
		// Check for workflows
		workflowDir := filepath.Join(githubDir, "workflows")
		if wfInfo, err := os.Stat(workflowDir); err == nil && wfInfo.IsDir() {
			files, _ := os.ReadDir(workflowDir)
			for _, f := range files {
				if !f.IsDir() && (strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml")) {
					cf := ConfigFile{
						Path:     filepath.Join(".github/workflows", f.Name()),
						Type:     "ci",
						Tool:     "github-actions",
					}
					if data, err := os.ReadFile(filepath.Join(workflowDir, f.Name())); err == nil {
						cf.LineCount = len(strings.Split(string(data), "\n"))
						cf.Complexity = getConfigComplexity(cf.LineCount)
					}
					configs = append(configs, cf)
				}
			}
		}
	}

	return configs
}

func countDependencies(repoPath string) int {
	count := 0

	// package.json
	pkgPath := filepath.Join(repoPath, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(data, &pkg) == nil {
			if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
				count += len(deps)
			}
			if devDeps, ok := pkg["devDependencies"].(map[string]interface{}); ok {
				count += len(devDeps)
			}
		}
	}

	// go.mod
	goModPath := filepath.Join(repoPath, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		lines := strings.Split(string(data), "\n")
		inRequire := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "require (") {
				inRequire = true
				continue
			}
			if inRequire && line == ")" {
				inRequire = false
				continue
			}
			if inRequire && line != "" && !strings.HasPrefix(line, "//") {
				count++
			}
			if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
				count++
			}
		}
	}

	// requirements.txt
	reqPath := filepath.Join(repoPath, "requirements.txt")
	if data, err := os.ReadFile(reqPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "-") {
				count++
			}
		}
	}

	// pyproject.toml dependencies
	pyprojectPath := filepath.Join(repoPath, "pyproject.toml")
	if data, err := os.ReadFile(pyprojectPath); err == nil {
		content := string(data)
		// Simple heuristic: count lines in dependencies section
		if idx := strings.Index(content, "[project.dependencies]"); idx != -1 {
			section := content[idx:]
			endIdx := strings.Index(section[1:], "[")
			if endIdx != -1 {
				section = section[:endIdx+1]
			}
			for _, line := range strings.Split(section, "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "#") {
					count++
				}
			}
		}
	}

	return count
}

func estimateBuildSteps(repoPath string, configFiles []ConfigFile) int {
	steps := 0

	// Check for package manager - at least 1 install step
	for _, cf := range configFiles {
		if cf.Type == "package-manager" {
			steps++
			break
		}
	}

	// Check for build tools
	hasBuildTool := false
	for _, cf := range configFiles {
		if cf.Type == "bundler" || cf.Type == "build" || cf.Type == "compiler" {
			hasBuildTool = true
			break
		}
	}
	if hasBuildTool {
		steps++
	}

	// Docker adds a step
	if fileExists(filepath.Join(repoPath, "Dockerfile")) {
		steps++
	}

	// Check package.json for scripts
	pkgPath := filepath.Join(repoPath, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(data, &pkg) == nil {
			if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
				// Count build-related scripts
				buildScripts := []string{"build", "compile", "prebuild", "postbuild"}
				for _, bs := range buildScripts {
					if _, exists := scripts[bs]; exists {
						steps++
					}
				}
			}
		}
	}

	// Makefile targets
	makefilePath := filepath.Join(repoPath, "Makefile")
	if data, err := os.ReadFile(makefilePath); err == nil {
		lines := strings.Split(string(data), "\n")
		targetCount := 0
		for _, line := range lines {
			if strings.Contains(line, ":") && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, "#") {
				targetCount++
			}
		}
		if targetCount > 3 {
			steps += 2 // Complex Makefile
		} else if targetCount > 0 {
			steps++
		}
	}

	if steps == 0 {
		steps = 1 // At minimum, there's always some setup
	}

	return steps
}

func parseEnvFile(path string) []EnvVariable {
	var vars []EnvVariable

	data, err := os.ReadFile(path)
	if err != nil {
		return vars
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) >= 1 {
			ev := EnvVariable{
				Name:       strings.TrimSpace(parts[0]),
				Source:     filepath.Base(path),
				HasDefault: len(parts) > 1 && strings.TrimSpace(parts[1]) != "",
			}
			vars = append(vars, ev)
		}
	}

	return vars
}

func parseDockerComposeEnv(path string) []EnvVariable {
	var vars []EnvVariable

	data, err := os.ReadFile(path)
	if err != nil {
		return vars
	}

	// Simple regex to find environment variables
	envPattern := regexp.MustCompile(`\$\{?([A-Z_][A-Z0-9_]*)\}?`)
	matches := envPattern.FindAllStringSubmatch(string(data), -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			seen[match[1]] = true
			vars = append(vars, EnvVariable{
				Name:   match[1],
				Source: "docker-compose",
			})
		}
	}

	return vars
}

// detectPrerequisites infers required tools from config files.
// Note: This could be enhanced to use tech-id scanner output for more comprehensive detection.
func detectPrerequisites(repoPath string) []Prerequisite {
	var prereqs []Prerequisite
	seen := make(map[string]bool)

	// Infer from package manager files
	if fileExists(filepath.Join(repoPath, "package.json")) && !seen["Node.js"] {
		seen["Node.js"] = true
		prereqs = append(prereqs, Prerequisite{Name: "Node.js", Source: "package.json", Required: true})
	}
	if fileExists(filepath.Join(repoPath, "go.mod")) && !seen["Go"] {
		seen["Go"] = true
		prereqs = append(prereqs, Prerequisite{Name: "Go", Source: "go.mod", Required: true})
	}
	if (fileExists(filepath.Join(repoPath, "requirements.txt")) ||
		fileExists(filepath.Join(repoPath, "pyproject.toml"))) && !seen["Python"] {
		seen["Python"] = true
		prereqs = append(prereqs, Prerequisite{Name: "Python", Source: "requirements.txt", Required: true})
	}
	if fileExists(filepath.Join(repoPath, "Cargo.toml")) && !seen["Rust"] {
		seen["Rust"] = true
		prereqs = append(prereqs, Prerequisite{Name: "Rust", Source: "Cargo.toml", Required: true})
	}
	if fileExists(filepath.Join(repoPath, "Gemfile")) && !seen["Ruby"] {
		seen["Ruby"] = true
		prereqs = append(prereqs, Prerequisite{Name: "Ruby", Source: "Gemfile", Required: true})
	}
	if fileExists(filepath.Join(repoPath, "composer.json")) && !seen["PHP"] {
		seen["PHP"] = true
		prereqs = append(prereqs, Prerequisite{Name: "PHP", Source: "composer.json", Required: true})
	}

	// Infer from container files
	if fileExists(filepath.Join(repoPath, "Dockerfile")) && !seen["Docker"] {
		seen["Docker"] = true
		prereqs = append(prereqs, Prerequisite{Name: "Docker", Source: "Dockerfile", Required: true})
	}
	if fileExists(filepath.Join(repoPath, "docker-compose.yml")) ||
		fileExists(filepath.Join(repoPath, "docker-compose.yaml")) {
		if !seen["Docker"] {
			seen["Docker"] = true
			prereqs = append(prereqs, Prerequisite{Name: "Docker", Source: "docker-compose", Required: true})
		}
		if !seen["Docker Compose"] {
			seen["Docker Compose"] = true
			prereqs = append(prereqs, Prerequisite{Name: "Docker Compose", Source: "docker-compose", Required: true})
		}
	}

	// Infer from build files
	if fileExists(filepath.Join(repoPath, "Makefile")) && !seen["Make"] {
		seen["Make"] = true
		prereqs = append(prereqs, Prerequisite{Name: "Make", Source: "Makefile", Required: true})
	}

	return prereqs
}

func analyzeReadme(repoPath string) *ReadmeAnalysis {
	readmeFiles := []string{"README.md", "README.rst", "README.txt", "README"}
	var readmePath string
	for _, rf := range readmeFiles {
		path := filepath.Join(repoPath, rf)
		if fileExists(path) {
			readmePath = path
			break
		}
	}

	if readmePath == "" {
		return nil
	}

	data, err := os.ReadFile(readmePath)
	if err != nil {
		return nil
	}

	content := string(data)
	contentLower := strings.ToLower(content)

	analysis := &ReadmeAnalysis{
		HasInstallSection: strings.Contains(contentLower, "install") ||
			strings.Contains(contentLower, "installation"),
		HasUsageSection: strings.Contains(contentLower, "usage") ||
			strings.Contains(contentLower, "how to use"),
		HasPrerequisites: strings.Contains(contentLower, "prerequisite") ||
			strings.Contains(contentLower, "requirements") ||
			strings.Contains(contentLower, "dependencies"),
		HasQuickStart: strings.Contains(contentLower, "quick start") ||
			strings.Contains(contentLower, "quickstart") ||
			strings.Contains(contentLower, "getting started"),
		HasExamples: strings.Contains(contentLower, "example") ||
			strings.Contains(contentLower, "sample"),
		WordCount: len(strings.Fields(content)),
	}

	// Count headers (Markdown)
	headerPattern := regexp.MustCompile(`(?m)^#{1,6}\s+`)
	analysis.HeaderCount = len(headerPattern.FindAllString(content, -1))

	// Count code blocks
	codeBlockPattern := regexp.MustCompile("```")
	analysis.CodeBlockCount = len(codeBlockPattern.FindAllString(content, -1)) / 2

	return analysis
}

func calculateReadmeScore(analysis *ReadmeAnalysis) int {
	score := 0

	if analysis.HasInstallSection {
		score += 20
	}
	if analysis.HasUsageSection {
		score += 15
	}
	if analysis.HasPrerequisites {
		score += 15
	}
	if analysis.HasQuickStart {
		score += 20
	}
	if analysis.HasExamples {
		score += 10
	}

	// Word count scoring
	if analysis.WordCount > 500 {
		score += 10
	} else if analysis.WordCount > 200 {
		score += 5
	}

	// Header structure
	if analysis.HeaderCount >= 5 {
		score += 5
	}

	// Code examples
	if analysis.CodeBlockCount >= 2 {
		score += 5
	}

	if score > 100 {
		score = 100
	}

	return score
}

func identifySetupBarriers(summary *OnboardingSummary, findings *OnboardingFindings) []SetupBarrier {
	var barriers []SetupBarrier

	// Too many config files
	if summary.ConfigFileCount > 15 {
		barriers = append(barriers, SetupBarrier{
			Category:    "configuration",
			Severity:    "high",
			Description: fmt.Sprintf("High number of config files (%d) may overwhelm new contributors", summary.ConfigFileCount),
			Suggestion:  "Consider consolidating configs or using a unified config approach",
		})
	} else if summary.ConfigFileCount > 10 {
		barriers = append(barriers, SetupBarrier{
			Category:    "configuration",
			Severity:    "medium",
			Description: fmt.Sprintf("Moderate number of config files (%d)", summary.ConfigFileCount),
			Suggestion:  "Document the purpose of each config file",
		})
	}

	// Many dependencies
	if summary.DependencyCount > 100 {
		barriers = append(barriers, SetupBarrier{
			Category:    "dependencies",
			Severity:    "high",
			Description: fmt.Sprintf("Large number of dependencies (%d) increases install time", summary.DependencyCount),
			Suggestion:  "Review dependencies for unused packages, consider code splitting",
		})
	} else if summary.DependencyCount > 50 {
		barriers = append(barriers, SetupBarrier{
			Category:    "dependencies",
			Severity:    "medium",
			Description: fmt.Sprintf("Moderate dependency count (%d)", summary.DependencyCount),
			Suggestion:  "Periodically audit dependencies",
		})
	}

	// Missing contribution docs
	if !summary.HasContributing {
		barriers = append(barriers, SetupBarrier{
			Category:    "documentation",
			Severity:    "medium",
			Description: "No CONTRIBUTING.md file found",
			Suggestion:  "Add a CONTRIBUTING.md with setup instructions and contribution guidelines",
		})
	}

	// Missing env example
	if summary.EnvVarCount > 0 && !summary.HasEnvExample {
		barriers = append(barriers, SetupBarrier{
			Category:    "configuration",
			Severity:    "high",
			Description: "Environment variables detected but no .env.example file",
			Suggestion:  "Add .env.example with all required variables documented",
		})
	}

	// Many prerequisites
	if summary.PrerequisiteCount > 5 {
		barriers = append(barriers, SetupBarrier{
			Category:    "dependencies",
			Severity:    "medium",
			Description: fmt.Sprintf("Many external prerequisites required (%d)", summary.PrerequisiteCount),
			Suggestion:  "Consider Docker or devcontainers to simplify setup",
		})
	}

	// Poor README
	if summary.ReadmeQualityScore < 40 {
		barriers = append(barriers, SetupBarrier{
			Category:    "documentation",
			Severity:    "high",
			Description: "README lacks essential setup information",
			Suggestion:  "Add installation, prerequisites, and quick start sections",
		})
	} else if summary.ReadmeQualityScore < 60 {
		barriers = append(barriers, SetupBarrier{
			Category:    "documentation",
			Severity:    "medium",
			Description: "README could be improved with more setup details",
			Suggestion:  "Add examples and usage documentation",
		})
	}

	return barriers
}

func calculateOnboardingScore(summary *OnboardingSummary) int {
	score := 100

	// Deduct for config complexity
	if summary.ConfigFileCount > 15 {
		score -= 20
	} else if summary.ConfigFileCount > 10 {
		score -= 10
	} else if summary.ConfigFileCount > 5 {
		score -= 5
	}

	// Deduct for dependency count
	if summary.DependencyCount > 100 {
		score -= 15
	} else if summary.DependencyCount > 50 {
		score -= 10
	} else if summary.DependencyCount > 25 {
		score -= 5
	}

	// Deduct for many build steps
	if summary.BuildStepCount > 5 {
		score -= 15
	} else if summary.BuildStepCount > 3 {
		score -= 10
	}

	// Deduct for many env vars without example
	if summary.EnvVarCount > 5 && !summary.HasEnvExample {
		score -= 15
	}

	// Deduct for many prerequisites
	if summary.PrerequisiteCount > 5 {
		score -= 10
	}

	// Bonus for good docs
	if summary.HasContributing {
		score += 10
	}
	if summary.HasEnvExample && summary.EnvVarCount > 0 {
		score += 10
	}

	// README quality bonus/penalty
	readmeBonus := (summary.ReadmeQualityScore - 50) / 5 // -10 to +10
	score += readmeBonus

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func getComplexityLevel(score int) string {
	if score >= 80 {
		return "low"
	} else if score >= 60 {
		return "medium"
	}
	return "high"
}

func getConfigComplexity(lineCount int) string {
	if lineCount > 200 {
		return "high"
	} else if lineCount > 50 {
		return "medium"
	}
	return "low"
}

// ============================================================================
// SPRAWL FEATURE
// Separates Tool Sprawl (dev tools) from Technology Sprawl (tech stack)
// Uses tech-id scanner results for technology detection
// ============================================================================

// TechIDResult represents the structure of tech-id scanner output
type TechIDResult struct {
	Summary struct {
		Technology *struct {
			TotalTechnologies int            `json:"total_technologies"`
			ByCategory        map[string]int `json:"by_category"`
			TopTechnologies   []string       `json:"top_technologies,omitempty"`
			PrimaryLanguages  []string       `json:"primary_languages,omitempty"`
			Frameworks        []string       `json:"frameworks,omitempty"`
			Databases         []string       `json:"databases,omitempty"`
			CloudServices     []string       `json:"cloud_services,omitempty"`
		} `json:"technology,omitempty"`
	} `json:"summary"`
	Findings struct {
		Technology []struct {
			Name       string `json:"name"`
			Category   string `json:"category"`
			Version    string `json:"version,omitempty"`
			Confidence int    `json:"confidence"`
			Source     string `json:"source"`
			File       string `json:"file,omitempty"`
		} `json:"technology,omitempty"`
	} `json:"findings"`
}

// DevOpsResult represents the structure of devops scanner output (optional)
// Used for DORA metrics context
type DevOpsResult struct {
	Summary struct {
		DORA *struct {
			DeploymentFrequency string  `json:"deployment_frequency"` // elite, high, medium, low
			LeadTime            string  `json:"lead_time"`            // elite, high, medium, low
			ChangeFailureRate   float64 `json:"change_failure_rate"`
			MTTR                string  `json:"mttr"`                 // elite, high, medium, low
			OverallPerformance  string  `json:"overall_performance"`
		} `json:"dora,omitempty"`
	} `json:"summary"`
}

// Tool categories represent dev tools (configuration burden)
var toolCategories = map[string]bool{
	"linter":    true,
	"formatter": true,
	"bundler":   true,
	"test":      true,
	"ci-cd":     true,
	"build":     true,
	"coverage":  true,
}

// Technology categories represent technologies (learning curve)
var technologyCategories = map[string]bool{
	"language":       true,
	"framework":      true,
	"database":       true,
	"cloud":          true,
	"container":      true,
	"infrastructure": true,
}

func (s *DevXScanner) runSprawl(ctx context.Context, opts *scanner.ScanOptions, cfg SprawlConfig) (*SprawlSummary, *SprawlFindings) {
	summary := &SprawlSummary{
		ToolSprawl: ToolSprawlMetrics{
			ByCategory: make(map[string]int),
		},
		TechnologySprawl: TechSprawlMetrics{
			ByCategory: make(map[string]int),
		},
	}
	findings := &SprawlFindings{
		Tools:          []DetectedTool{},
		Technologies:   []DetectedTech{},
		ConfigAnalysis: []ConfigAnalysis{},
		SprawlIssues:   []SprawlIssue{},
	}

	// Load tech-id results (from dependency scanner)
	techIDData := loadTechIDResults(opts.OutputDir)

	// Separate tech-id findings into tools vs technologies
	if techIDData != nil && len(techIDData.Findings.Technology) > 0 {
		toolSeen := make(map[string]bool)
		techSeen := make(map[string]bool)

		for _, tech := range techIDData.Findings.Technology {
			category := mapTechIDCategory(tech.Category)

			// Determine if this is a tool or technology
			if isToolCategory(category) {
				if !toolSeen[tech.Name] {
					toolSeen[tech.Name] = true
					findings.Tools = append(findings.Tools, DetectedTool{
						Name:       tech.Name,
						Category:   category,
						ConfigFile: tech.File,
						Version:    tech.Version,
					})
					summary.ToolSprawl.ByCategory[category]++
				}
			} else if isTechnologyCategory(category) {
				if !techSeen[tech.Name] {
					techSeen[tech.Name] = true
					findings.Technologies = append(findings.Technologies, DetectedTech{
						Name:       tech.Name,
						Category:   category,
						Confidence: tech.Confidence,
						Source:     tech.Source,
					})
					summary.TechnologySprawl.ByCategory[category]++
				}
			}
		}

		summary.ToolSprawl.Index = len(toolSeen)
		summary.TechnologySprawl.Index = len(techSeen)
	}

	// Set tool sprawl level
	summary.ToolSprawl.Level = getSprawlLevel(summary.ToolSprawl.Index, cfg.MaxRecommendedTools)

	// Set technology sprawl level
	summary.TechnologySprawl.Level = getSprawlLevel(summary.TechnologySprawl.Index, cfg.MaxRecommendedTechnologies)

	// Calculate learning curve score
	summary.LearningCurveScore = calculateLearningCurveScore(summary, findings)
	summary.LearningCurve = getLearningCurveLevel(summary.LearningCurveScore)

	// Analyze config file complexity
	if cfg.CheckConfigComplexity {
		configs := analyzeConfigComplexity(opts.RepoPath)
		findings.ConfigAnalysis = configs

		totalLines := 0
		maxComplexity := 0
		for _, c := range configs {
			totalLines += c.LineCount
			if c.ComplexityScore > maxComplexity {
				maxComplexity = c.ComplexityScore
			}
		}
		summary.TotalConfigLines = totalLines

		if maxComplexity > 70 || totalLines > 2000 {
			summary.ConfigComplexity = "high"
		} else if maxComplexity > 40 || totalLines > 1000 {
			summary.ConfigComplexity = "medium"
		} else {
			summary.ConfigComplexity = "low"
		}
	}

	// Identify sprawl issues
	if cfg.CheckToolSprawl || cfg.CheckTechnologySprawl {
		sprawlIssues := identifySprawlIssues(findings.Tools, findings.Technologies, cfg)
		findings.SprawlIssues = sprawlIssues
	}

	// Calculate combined score (0-100, higher is simpler)
	summary.CombinedScore = calculateSprawlScore(summary)

	// Optionally load DORA context for insights
	summary.DORAContext = loadDORAContext(opts.OutputDir, summary)

	return summary, findings
}

// isToolCategory returns true if the category represents a dev tool
func isToolCategory(category string) bool {
	return toolCategories[category]
}

// isTechnologyCategory returns true if the category represents a technology
func isTechnologyCategory(category string) bool {
	return technologyCategories[category]
}

// getSprawlLevel returns the sprawl level based on index and threshold
func getSprawlLevel(index, threshold int) string {
	ratio := float64(index) / float64(threshold)
	if ratio > 2.0 {
		return "excessive"
	} else if ratio > 1.4 {
		return "high"
	} else if ratio > 0.8 {
		return "moderate"
	}
	return "low"
}

// calculateLearningCurveScore estimates the learning curve based on technology sprawl
// Lower score = steeper learning curve
func calculateLearningCurveScore(summary *SprawlSummary, findings *SprawlFindings) int {
	score := 100

	// Deduct for number of languages (each additional language adds learning burden)
	langCount := summary.TechnologySprawl.ByCategory["language"]
	if langCount > 4 {
		score -= 25
	} else if langCount > 2 {
		score -= 10
	}

	// Deduct for frameworks (framework complexity varies)
	frameworkCount := summary.TechnologySprawl.ByCategory["framework"]
	if frameworkCount > 5 {
		score -= 20
	} else if frameworkCount > 3 {
		score -= 10
	}

	// Deduct for databases (each DB has different paradigms)
	dbCount := summary.TechnologySprawl.ByCategory["database"]
	if dbCount > 3 {
		score -= 15
	} else if dbCount > 1 {
		score -= 5
	}

	// Deduct for cloud services (complex to understand)
	cloudCount := summary.TechnologySprawl.ByCategory["cloud"]
	if cloudCount > 5 {
		score -= 20
	} else if cloudCount > 2 {
		score -= 10
	}

	// Deduct for infrastructure complexity
	infraCount := summary.TechnologySprawl.ByCategory["infrastructure"] + summary.TechnologySprawl.ByCategory["container"]
	if infraCount > 3 {
		score -= 15
	} else if infraCount > 1 {
		score -= 5
	}

	// Overall technology count penalty
	totalTech := summary.TechnologySprawl.Index
	if totalTech > 20 {
		score -= 20
	} else if totalTech > 15 {
		score -= 10
	} else if totalTech > 10 {
		score -= 5
	}

	if score < 0 {
		score = 0
	}
	return score
}

// getLearningCurveLevel converts score to level
func getLearningCurveLevel(score int) string {
	if score >= 80 {
		return "low"
	} else if score >= 60 {
		return "moderate"
	} else if score >= 40 {
		return "high"
	}
	return "steep"
}

// loadTechIDResults loads the tech-id scanner output from the analysis directory
func loadTechIDResults(outputDir string) *TechIDResult {
	if outputDir == "" {
		return nil
	}

	techIDPath := filepath.Join(outputDir, "tech-id.json")
	data, err := os.ReadFile(techIDPath)
	if err != nil {
		return nil
	}

	var result TechIDResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}

	return &result
}

// loadDORAContext optionally loads devops.json for DORA metrics context
// This provides insights like "high sprawl but elite deployment frequency"
func loadDORAContext(outputDir string, summary *SprawlSummary) *DORAContext {
	if outputDir == "" {
		return nil
	}

	devopsPath := filepath.Join(outputDir, "devops.json")
	data, err := os.ReadFile(devopsPath)
	if err != nil {
		return nil // devops.json not available, skip DORA context
	}

	var devops DevOpsResult
	if err := json.Unmarshal(data, &devops); err != nil {
		return nil
	}

	if devops.Summary.DORA == nil {
		return nil
	}

	ctx := &DORAContext{
		OverallPerformance: devops.Summary.DORA.OverallPerformance,
	}

	// Generate insight based on sprawl + DORA combination
	ctx.Insight = generateDORAInsight(summary, devops.Summary.DORA.OverallPerformance)

	return ctx
}

// generateDORAInsight creates an insight string based on sprawl and DORA performance
func generateDORAInsight(summary *SprawlSummary, doraPerformance string) string {
	sprawlLevel := summary.ToolSprawl.Level
	if summary.TechnologySprawl.Level == "excessive" || summary.TechnologySprawl.Level == "high" {
		sprawlLevel = summary.TechnologySprawl.Level
	}

	switch {
	case sprawlLevel == "excessive" && doraPerformance == "elite":
		return "High sprawl but elite DORA performance - team manages complexity well"
	case sprawlLevel == "excessive" && (doraPerformance == "medium" || doraPerformance == "low"):
		return "High sprawl correlating with lower DORA performance - consider simplification"
	case sprawlLevel == "high" && doraPerformance == "elite":
		return "Moderate sprawl with elite DORA - good balance of tools and velocity"
	case sprawlLevel == "low" && doraPerformance == "low":
		return "Low sprawl but low DORA performance - tooling may not be the bottleneck"
	case sprawlLevel == "low" && doraPerformance == "elite":
		return "Low sprawl with elite DORA - excellent developer experience"
	default:
		return ""
	}
}

// mapTechIDCategory maps tech-id categories to devex-friendly categories
func mapTechIDCategory(techIDCategory string) string {
	categoryMap := map[string]string{
		"language":       "language",
		"framework":      "framework",
		"database":       "database",
		"container":      "container",
		"iac":            "infrastructure",
		"ci-cd":          "ci",
		"cloud":          "cloud",
		"testing":        "test",
		"linting":        "linter",
		"formatting":     "formatter",
		"bundling":       "bundler",
		"build":          "build",
		"package-manager": "package-manager",
		"monitoring":     "monitoring",
		"logging":        "logging",
		"security":       "security",
	}

	if mapped, ok := categoryMap[techIDCategory]; ok {
		return mapped
	}
	return techIDCategory
}

func analyzeConfigComplexity(repoPath string) []ConfigAnalysis {
	var configs []ConfigAnalysis

	// Files to analyze
	filesToAnalyze := []string{
		"package.json",
		"tsconfig.json",
		".eslintrc.json",
		".eslintrc.js",
		"webpack.config.js",
		"vite.config.ts",
		"vite.config.js",
		"turbo.json",
		"nx.json",
		".github/workflows",
	}

	for _, file := range filesToAnalyze {
		fullPath := filepath.Join(repoPath, file)

		// Handle directories (like workflows)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(fullPath)
			for _, entry := range entries {
				if !entry.IsDir() {
					filePath := filepath.Join(fullPath, entry.Name())
					if analysis := analyzeConfigFile(filePath, filepath.Join(file, entry.Name())); analysis != nil {
						configs = append(configs, *analysis)
					}
				}
			}
			continue
		}

		if analysis := analyzeConfigFile(fullPath, file); analysis != nil {
			configs = append(configs, *analysis)
		}
	}

	return configs
}

func analyzeConfigFile(fullPath, relativePath string) *ConfigAnalysis {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	analysis := &ConfigAnalysis{
		Path:      relativePath,
		Tool:      guessToolFromPath(relativePath),
		LineCount: len(lines),
	}

	// Calculate nesting depth for JSON files
	if strings.HasSuffix(relativePath, ".json") {
		analysis.NestingDepth = calculateJSONNesting(content)
	}

	// Count override patterns
	overridePatterns := []string{
		"override", "exclude", "ignore", "disable", "off", "extends",
	}
	for _, pattern := range overridePatterns {
		analysis.OverrideCount += strings.Count(strings.ToLower(content), pattern)
	}

	// Calculate complexity score
	analysis.ComplexityScore = calculateConfigComplexityScore(analysis)

	return analysis
}

func guessToolFromPath(path string) string {
	name := filepath.Base(path)
	if info, ok := configFilePatterns[name]; ok {
		return info.tool
	}
	if strings.Contains(path, "workflows") {
		return "github-actions"
	}
	return "unknown"
}

func calculateJSONNesting(content string) int {
	maxDepth := 0
	currentDepth := 0

	for _, char := range content {
		if char == '{' || char == '[' {
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		} else if char == '}' || char == ']' {
			currentDepth--
		}
	}

	return maxDepth
}

func calculateConfigComplexityScore(analysis *ConfigAnalysis) int {
	score := 0

	// Line count contribution (max 40)
	if analysis.LineCount > 500 {
		score += 40
	} else if analysis.LineCount > 200 {
		score += 30
	} else if analysis.LineCount > 100 {
		score += 20
	} else if analysis.LineCount > 50 {
		score += 10
	}

	// Nesting depth contribution (max 30)
	if analysis.NestingDepth > 10 {
		score += 30
	} else if analysis.NestingDepth > 6 {
		score += 20
	} else if analysis.NestingDepth > 4 {
		score += 10
	}

	// Override count contribution (max 30)
	if analysis.OverrideCount > 50 {
		score += 30
	} else if analysis.OverrideCount > 20 {
		score += 20
	} else if analysis.OverrideCount > 10 {
		score += 10
	}

	return score
}

// detectLanguages is no longer needed - we use tech-id scanner data instead

func identifySprawlIssues(tools []DetectedTool, technologies []DetectedTech, cfg SprawlConfig) []SprawlIssue {
	var issues []SprawlIssue

	// Check for excessive tools
	if len(tools) > cfg.MaxRecommendedTools {
		issues = append(issues, SprawlIssue{
			Category:    "tool_sprawl",
			Severity:    "medium",
			Description: fmt.Sprintf("Tool count (%d) exceeds recommended maximum (%d)", len(tools), cfg.MaxRecommendedTools),
			Suggestion:  "Consider consolidating tools or removing unused ones",
		})
	}

	// Check for excessive technologies
	if len(technologies) > cfg.MaxRecommendedTechnologies {
		issues = append(issues, SprawlIssue{
			Category:    "technology_sprawl",
			Severity:    "medium",
			Description: fmt.Sprintf("Technology count (%d) exceeds recommended maximum (%d)", len(technologies), cfg.MaxRecommendedTechnologies),
			Suggestion:  "High technology sprawl increases onboarding time and cognitive load",
		})
	}

	// Check for overlapping linters
	linterTools := []string{}
	for _, t := range tools {
		if t.Category == "linter" {
			linterTools = append(linterTools, t.Name)
		}
	}
	if len(linterTools) > 3 {
		issues = append(issues, SprawlIssue{
			Category:    "overlap",
			Severity:    "low",
			Description: fmt.Sprintf("Multiple linting tools detected: %s", strings.Join(linterTools, ", ")),
			Tools:       linterTools,
			Suggestion:  "Consider consolidating linters to reduce configuration overhead",
		})
	}

	// Check for multiple bundlers
	bundlerTools := []string{}
	for _, t := range tools {
		if t.Category == "bundler" {
			bundlerTools = append(bundlerTools, t.Name)
		}
	}
	if len(bundlerTools) > 1 {
		issues = append(issues, SprawlIssue{
			Category:    "duplication",
			Severity:    "medium",
			Description: fmt.Sprintf("Multiple bundlers detected: %s", strings.Join(bundlerTools, ", ")),
			Tools:       bundlerTools,
			Suggestion:  "Using multiple bundlers adds complexity; consider standardizing on one",
		})
	}

	// Check for multiple CI systems
	ciTools := []string{}
	for _, t := range tools {
		if t.Category == "ci-cd" {
			ciTools = append(ciTools, t.Name)
		}
	}
	if len(ciTools) > 1 {
		issues = append(issues, SprawlIssue{
			Category:    "duplication",
			Severity:    "low",
			Description: fmt.Sprintf("Multiple CI systems detected: %s", strings.Join(ciTools, ", ")),
			Tools:       ciTools,
			Suggestion:  "Consider standardizing on a single CI system",
		})
	}

	// Check for multiple languages (learning curve)
	langTech := []string{}
	for _, t := range technologies {
		if t.Category == "language" {
			langTech = append(langTech, t.Name)
		}
	}
	if len(langTech) > 4 {
		issues = append(issues, SprawlIssue{
			Category:    "learning_curve",
			Severity:    "medium",
			Description: fmt.Sprintf("Multiple programming languages (%d): %s", len(langTech), strings.Join(langTech, ", ")),
			Tools:       langTech,
			Suggestion:  "Many languages increase onboarding time; consider if all are necessary",
		})
	}

	return issues
}

// calculateSprawlScore returns combined score (0-100, higher is simpler)
func calculateSprawlScore(summary *SprawlSummary) int {
	score := 100

	// Tool sprawl impact (40% weight)
	switch summary.ToolSprawl.Level {
	case "excessive":
		score -= 30
	case "high":
		score -= 20
	case "moderate":
		score -= 10
	}

	// Technology sprawl impact (40% weight)
	switch summary.TechnologySprawl.Level {
	case "excessive":
		score -= 30
	case "high":
		score -= 20
	case "moderate":
		score -= 10
	}

	// Config complexity impact (20% weight)
	switch summary.ConfigComplexity {
	case "high":
		score -= 15
	case "medium":
		score -= 8
	}

	if score < 0 {
		score = 0
	}

	return score
}

// ============================================================================
// WORKFLOW FEATURE
// ============================================================================

func (s *DevXScanner) runWorkflow(ctx context.Context, opts *scanner.ScanOptions, cfg WorkflowConfig) (*WorkflowSummary, *WorkflowFindings) {
	summary := &WorkflowSummary{}
	findings := &WorkflowFindings{
		PRTemplates:    []PRTemplate{},
		IssueTemplates: []IssueTemplate{},
		FeedbackTools:  []FeedbackTool{},
		WorkflowIssues: []WorkflowIssue{},
	}

	// Check for PR templates
	if cfg.CheckPRTemplates {
		prTemplates := findPRTemplates(opts.RepoPath)
		findings.PRTemplates = prTemplates
		summary.HasPRTemplates = len(prTemplates) > 0

		issueTemplates := findIssueTemplates(opts.RepoPath)
		findings.IssueTemplates = issueTemplates
		summary.HasIssueTemplates = len(issueTemplates) > 0
	}

	// Check for local dev setup
	if cfg.CheckLocalDev {
		devSetup := analyzeDevSetup(opts.RepoPath)
		findings.DevSetup = devSetup
		summary.HasDevContainer = devSetup.HasDevContainer
		summary.HasDockerCompose = devSetup.HasDockerCompose
	}

	// Check for feedback loop tools
	if cfg.CheckFeedbackLoop {
		feedbackTools := findFeedbackTools(opts.RepoPath)
		findings.FeedbackTools = feedbackTools
		summary.HasHotReload = false
		summary.HasWatchMode = false
		for _, ft := range feedbackTools {
			if ft.Type == "hot_reload" {
				summary.HasHotReload = true
			}
			if ft.Type == "watch" {
				summary.HasWatchMode = true
			}
		}
	}

	// Identify workflow issues
	findings.WorkflowIssues = identifyWorkflowIssues(summary, findings)

	// Calculate scores
	summary.PRProcessScore = calculatePRProcessScore(summary, findings)
	summary.LocalDevScore = calculateLocalDevScore(summary, findings)
	summary.FeedbackLoopScore = calculateFeedbackLoopScore(summary, findings)

	// Overall score (weighted average)
	summary.Score = (summary.PRProcessScore*30 + summary.LocalDevScore*35 + summary.FeedbackLoopScore*35) / 100

	// Efficiency level
	if summary.Score >= 75 {
		summary.EfficiencyLevel = "high"
	} else if summary.Score >= 50 {
		summary.EfficiencyLevel = "medium"
	} else {
		summary.EfficiencyLevel = "low"
	}

	return summary, findings
}

func findPRTemplates(repoPath string) []PRTemplate {
	var templates []PRTemplate

	// Check common locations
	locations := []string{
		".github/PULL_REQUEST_TEMPLATE.md",
		".github/pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
		"pull_request_template.md",
		"docs/PULL_REQUEST_TEMPLATE.md",
	}

	for _, loc := range locations {
		path := filepath.Join(repoPath, loc)
		if data, err := os.ReadFile(path); err == nil {
			content := string(data)
			template := PRTemplate{
				Path:         loc,
				HasChecklist: strings.Contains(content, "- [ ]"),
				HasSections:  strings.Count(content, "##") > 0,
			}

			// Extract section names
			headerPattern := regexp.MustCompile(`(?m)^##\s+(.+)$`)
			matches := headerPattern.FindAllStringSubmatch(content, -1)
			for _, m := range matches {
				if len(m) > 1 {
					template.Sections = append(template.Sections, m[1])
				}
			}

			templates = append(templates, template)
		}
	}

	// Check .github/PULL_REQUEST_TEMPLATE directory
	templateDir := filepath.Join(repoPath, ".github", "PULL_REQUEST_TEMPLATE")
	if entries, err := os.ReadDir(templateDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				path := filepath.Join(templateDir, entry.Name())
				if data, err := os.ReadFile(path); err == nil {
					content := string(data)
					templates = append(templates, PRTemplate{
						Path:         filepath.Join(".github/PULL_REQUEST_TEMPLATE", entry.Name()),
						HasChecklist: strings.Contains(content, "- [ ]"),
						HasSections:  strings.Count(content, "##") > 0,
					})
				}
			}
		}
	}

	return templates
}

func findIssueTemplates(repoPath string) []IssueTemplate {
	var templates []IssueTemplate

	// Check .github/ISSUE_TEMPLATE directory
	templateDir := filepath.Join(repoPath, ".github", "ISSUE_TEMPLATE")
	if entries, err := os.ReadDir(templateDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			templateType := "general"
			if strings.Contains(strings.ToLower(name), "bug") {
				templateType = "bug"
			} else if strings.Contains(strings.ToLower(name), "feature") {
				templateType = "feature"
			}

			templates = append(templates, IssueTemplate{
				Path: filepath.Join(".github/ISSUE_TEMPLATE", name),
				Name: strings.TrimSuffix(name, filepath.Ext(name)),
				Type: templateType,
			})
		}
	}

	// Check for single issue template
	singleLocs := []string{
		".github/ISSUE_TEMPLATE.md",
		"ISSUE_TEMPLATE.md",
	}
	for _, loc := range singleLocs {
		if fileExists(filepath.Join(repoPath, loc)) {
			templates = append(templates, IssueTemplate{
				Path: loc,
				Name: "default",
				Type: "general",
			})
		}
	}

	return templates
}

func analyzeDevSetup(repoPath string) *DevSetup {
	setup := &DevSetup{
		DevScripts:    []string{},
		SetupCommands: []string{},
	}

	// Check for docker-compose
	dcFiles := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
	for _, dc := range dcFiles {
		if fileExists(filepath.Join(repoPath, dc)) {
			setup.HasDockerCompose = true
			break
		}
	}

	// Check for devcontainer
	if fileExists(filepath.Join(repoPath, ".devcontainer")) ||
		fileExists(filepath.Join(repoPath, ".devcontainer.json")) ||
		fileExists(filepath.Join(repoPath, ".devcontainer", "devcontainer.json")) {
		setup.HasDevContainer = true
	}

	// Check for Makefile
	if fileExists(filepath.Join(repoPath, "Makefile")) {
		setup.HasMakefile = true
	}

	// Check for Taskfile
	taskFiles := []string{"Taskfile.yml", "Taskfile.yaml", "taskfile.yml", "taskfile.yaml"}
	for _, tf := range taskFiles {
		if fileExists(filepath.Join(repoPath, tf)) {
			setup.HasTaskfile = true
			break
		}
	}

	// Check package.json for dev scripts
	pkgPath := filepath.Join(repoPath, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(data, &pkg) == nil {
			if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
				devScripts := []string{"dev", "start", "serve", "watch", "develop"}
				for _, ds := range devScripts {
					if _, exists := scripts[ds]; exists {
						setup.DevScripts = append(setup.DevScripts, fmt.Sprintf("npm run %s", ds))
					}
				}
			}
		}
	}

	// Check Makefile for common targets
	if setup.HasMakefile {
		if data, err := os.ReadFile(filepath.Join(repoPath, "Makefile")); err == nil {
			content := string(data)
			makeTargets := []string{"dev", "run", "serve", "start", "watch"}
			for _, target := range makeTargets {
				if strings.Contains(content, target+":") {
					setup.DevScripts = append(setup.DevScripts, fmt.Sprintf("make %s", target))
				}
			}
		}
	}

	return setup
}

func findFeedbackTools(repoPath string) []FeedbackTool {
	var tools []FeedbackTool

	// Check package.json for hot reload indicators
	pkgPath := filepath.Join(repoPath, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		content := string(data)

		// Hot reload tools
		hotReloadPackages := []string{
			"react-hot-loader",
			"react-refresh",
			"@pmmmwh/react-refresh-webpack-plugin",
			"webpack-hot-middleware",
			"next", // Next.js has built-in hot reload
			"vite", // Vite has built-in HMR
		}

		for _, pkg := range hotReloadPackages {
			if strings.Contains(content, pkg) {
				tools = append(tools, FeedbackTool{
					Name:        pkg,
					Type:        "hot_reload",
					Source:      "package.json",
					Description: "Hot module replacement for fast feedback",
				})
			}
		}

		// Watch mode tools
		watchPackages := []string{
			"nodemon",
			"ts-node-dev",
			"chokidar",
			"concurrently",
		}

		for _, pkg := range watchPackages {
			if strings.Contains(content, pkg) {
				tools = append(tools, FeedbackTool{
					Name:        pkg,
					Type:        "watch",
					Source:      "package.json",
					Description: "File watching for automatic rebuilds",
				})
			}
		}
	}

	// Check for vite.config (Vite has HMR)
	if fileExists(filepath.Join(repoPath, "vite.config.js")) ||
		fileExists(filepath.Join(repoPath, "vite.config.ts")) {
		tools = append(tools, FeedbackTool{
			Name:        "Vite HMR",
			Type:        "hot_reload",
			Source:      "vite.config",
			Description: "Vite's built-in Hot Module Replacement",
		})
	}

	// Check for air (Go hot reload)
	if fileExists(filepath.Join(repoPath, ".air.toml")) ||
		fileExists(filepath.Join(repoPath, "air.toml")) {
		tools = append(tools, FeedbackTool{
			Name:        "air",
			Type:        "hot_reload",
			Source:      ".air.toml",
			Description: "Live reload for Go apps",
		})
	}

	return tools
}

func identifyWorkflowIssues(summary *WorkflowSummary, findings *WorkflowFindings) []WorkflowIssue {
	var issues []WorkflowIssue

	// No PR templates
	if !summary.HasPRTemplates {
		issues = append(issues, WorkflowIssue{
			Category:    "pr_process",
			Severity:    "medium",
			Description: "No PR template found",
			Suggestion:  "Add a PR template with checklist and sections for description, testing, etc.",
		})
	}

	// No issue templates
	if !summary.HasIssueTemplates {
		issues = append(issues, WorkflowIssue{
			Category:    "pr_process",
			Severity:    "low",
			Description: "No issue templates found",
			Suggestion:  "Add issue templates for bugs and feature requests",
		})
	}

	// No local dev setup
	if !summary.HasDockerCompose && !summary.HasDevContainer && findings.DevSetup != nil && len(findings.DevSetup.DevScripts) == 0 {
		issues = append(issues, WorkflowIssue{
			Category:    "local_dev",
			Severity:    "high",
			Description: "No clear local development setup found",
			Suggestion:  "Add docker-compose.yml, devcontainer, or document dev setup in README",
		})
	}

	// No hot reload
	if !summary.HasHotReload && !summary.HasWatchMode {
		issues = append(issues, WorkflowIssue{
			Category:    "feedback_loop",
			Severity:    "medium",
			Description: "No hot reload or watch mode detected",
			Suggestion:  "Consider adding hot reload for faster development feedback",
		})
	}

	return issues
}

func calculatePRProcessScore(summary *WorkflowSummary, findings *WorkflowFindings) int {
	score := 50 // Base score

	if summary.HasPRTemplates {
		score += 25
		// Bonus for quality templates
		for _, t := range findings.PRTemplates {
			if t.HasChecklist {
				score += 10
			}
			if len(t.Sections) >= 3 {
				score += 5
			}
		}
	}

	if summary.HasIssueTemplates {
		score += 10
	}

	if score > 100 {
		score = 100
	}

	return score
}

func calculateLocalDevScore(summary *WorkflowSummary, findings *WorkflowFindings) int {
	score := 30 // Base score

	if summary.HasDockerCompose {
		score += 25
	}

	if summary.HasDevContainer {
		score += 25
	}

	if findings.DevSetup != nil {
		if findings.DevSetup.HasMakefile {
			score += 10
		}
		if findings.DevSetup.HasTaskfile {
			score += 10
		}
		if len(findings.DevSetup.DevScripts) > 0 {
			score += 10
		}
	}

	if score > 100 {
		score = 100
	}

	return score
}

func calculateFeedbackLoopScore(summary *WorkflowSummary, findings *WorkflowFindings) int {
	score := 40 // Base score

	if summary.HasHotReload {
		score += 30
	}

	if summary.HasWatchMode {
		score += 20
	}

	// Bonus for multiple feedback tools
	if len(findings.FeedbackTools) > 1 {
		score += 10
	}

	if score > 100 {
		score = 100
	}

	return score
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
