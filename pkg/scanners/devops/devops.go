// Package devops provides the consolidated DevOps and CI/CD security super scanner
// Features: iac, containers, github-actions, dora, git
// Renamed from infra - absorbed github-actions-security standalone scanner
package devops

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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/scanners/common"
)

const (
	Name    = "devops"
	Version = "3.0.0"
)

func init() {
	scanner.Register(&DevOpsScanner{})
}

// DevOpsScanner consolidates all DevOps and CI/CD security analysis
type DevOpsScanner struct{}

func (s *DevOpsScanner) Name() string {
	return Name
}

func (s *DevOpsScanner) Description() string {
	return "Consolidated DevOps scanner: IaC security, containers, GitHub Actions, DORA metrics, git insights"
}

func (s *DevOpsScanner) Dependencies() []string {
	return nil
}

func (s *DevOpsScanner) EstimateDuration(fileCount int) time.Duration {
	return 30 * time.Second
}

func (s *DevOpsScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	cfg := getConfig(opts)

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Run features in parallel where possible
	if cfg.IaC.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runIaC(ctx, opts, cfg.IaC)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "iac")
			result.Summary.IaC = summary
			result.Findings.IaC = findings
			mu.Unlock()
		}()
	}

	if cfg.Containers.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runContainers(ctx, opts, cfg.Containers)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "containers")
			result.Summary.Containers = summary
			result.Findings.Containers = findings
			mu.Unlock()
		}()
	}

	if cfg.GitHubActions.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runGitHubActions(ctx, opts, cfg.GitHubActions)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "github_actions")
			result.Summary.GitHubActions = summary
			result.Findings.GitHubActions = findings
			mu.Unlock()
		}()
	}

	if cfg.DORA.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, metrics := s.runDORA(ctx, opts, cfg.DORA)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "dora")
			result.Summary.DORA = summary
			result.Findings.DORA = metrics
			mu.Unlock()
		}()
	}

	if cfg.Git.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runGit(ctx, opts, cfg.Git)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "git")
			result.Summary.Git = summary
			result.Findings.Git = findings
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
// IAC FEATURE
// ============================================================================

func (s *DevOpsScanner) runIaC(ctx context.Context, opts *scanner.ScanOptions, cfg IaCConfig) (*IaCSummary, []IaCFinding) {
	var findings []IaCFinding
	summary := &IaCSummary{
		ByType: make(map[string]int),
	}

	useCheckov := cfg.Tool == "checkov" || (cfg.Tool == "auto" && common.ToolExists("checkov"))
	useTrivy := cfg.Tool == "trivy" || (cfg.Tool == "auto" && !useCheckov && common.ToolExists("trivy"))

	if !useCheckov && !useTrivy {
		summary.Error = "neither checkov nor trivy found"
		return summary, findings
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	var result *common.CommandResult
	var err error

	if useCheckov {
		summary.Tool = "checkov"
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		result, err = common.RunCommand(ctx, "checkov",
			"--directory", opts.RepoPath,
			"--output", "json",
			"--quiet",
			"--compact",
			"--skip-path", "node_modules",
			"--skip-path", "vendor",
			"--skip-path", ".git",
		)

		if err != nil && cfg.FallbackTool && common.ToolExists("trivy") {
			useTrivy = true
			useCheckov = false
		}
	}

	if useTrivy && !useCheckov {
		summary.Tool = "trivy"
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		result, err = common.RunCommand(ctx, "trivy",
			"config",
			"--format", "json",
			"--severity", "CRITICAL,HIGH,MEDIUM,LOW",
			"--skip-dirs", "node_modules",
			"--skip-dirs", "vendor",
			"--skip-dirs", ".git",
			opts.RepoPath,
		)
	}

	if err != nil || result == nil {
		return summary, findings
	}

	if summary.Tool == "checkov" {
		findings, summary = parseCheckovOutput(result.Stdout, opts.RepoPath)
	} else {
		findings, summary = parseTrivyIaCOutput(result.Stdout, opts.RepoPath)
	}

	return summary, findings
}

func parseCheckovOutput(data []byte, repoPath string) ([]IaCFinding, *IaCSummary) {
	var findings []IaCFinding
	summary := &IaCSummary{
		ByType: make(map[string]int),
		Tool:   "checkov",
	}

	type checkovOutput struct {
		CheckType string `json:"check_type"`
		Results   struct {
			FailedChecks []struct {
				CheckID       string `json:"check_id"`
				FilePath      string `json:"file_path"`
				FileLineRange []int  `json:"file_line_range"`
				Resource      string `json:"resource"`
				Guideline     string `json:"guideline"`
				Description   string `json:"description"`
			} `json:"failed_checks"`
		} `json:"results"`
	}

	var outputs []checkovOutput
	if err := json.Unmarshal(data, &outputs); err != nil {
		var single checkovOutput
		if err := json.Unmarshal(data, &single); err != nil {
			return findings, summary
		}
		outputs = []checkovOutput{single}
	}

	filesSet := make(map[string]bool)

	for _, output := range outputs {
		checkType := normalizeIaCType(output.CheckType)

		for _, check := range output.Results.FailedChecks {
			file := strings.TrimPrefix(check.FilePath, "/")
			if strings.HasPrefix(file, repoPath) {
				file = strings.TrimPrefix(file, repoPath+"/")
			}

			filesSet[file] = true
			severity := deriveCheckovSeverity(check.CheckID)

			line := 0
			if len(check.FileLineRange) > 0 {
				line = check.FileLineRange[0]
			}

			finding := IaCFinding{
				RuleID:      check.CheckID,
				Title:       check.Description,
				Description: check.Description,
				Severity:    severity,
				File:        file,
				Line:        line,
				Resource:    check.Resource,
				Type:        checkType,
				Resolution:  check.Guideline,
				CheckType:   output.CheckType,
			}
			findings = append(findings, finding)

			summary.TotalFindings++
			summary.ByType[checkType]++

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
	}

	summary.FilesScanned = len(filesSet)
	return findings, summary
}

func parseTrivyIaCOutput(data []byte, repoPath string) ([]IaCFinding, *IaCSummary) {
	var findings []IaCFinding
	summary := &IaCSummary{
		ByType: make(map[string]int),
		Tool:   "trivy",
	}

	type trivyOutput struct {
		Results []struct {
			Target            string `json:"Target"`
			Type              string `json:"Type"`
			Misconfigurations []struct {
				ID          string `json:"ID"`
				Title       string `json:"Title"`
				Description string `json:"Description"`
				Resolution  string `json:"Resolution"`
				Severity    string `json:"Severity"`
				Status      string `json:"Status"`
				CauseMetadata struct {
					Resource  string `json:"Resource"`
					StartLine int    `json:"StartLine"`
				} `json:"CauseMetadata"`
			} `json:"Misconfigurations"`
		} `json:"Results"`
	}

	var output trivyOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return findings, summary
	}

	filesSet := make(map[string]bool)

	for _, result := range output.Results {
		filesSet[result.Target] = true

		for _, misconfig := range result.Misconfigurations {
			if misconfig.Status == "PASS" {
				continue
			}

			file := result.Target
			if strings.HasPrefix(file, repoPath) {
				file = strings.TrimPrefix(file, repoPath+"/")
			}

			severity := strings.ToLower(misconfig.Severity)
			iacType := normalizeIaCType(result.Type)

			finding := IaCFinding{
				RuleID:      misconfig.ID,
				Title:       misconfig.Title,
				Description: misconfig.Description,
				Severity:    severity,
				File:        file,
				Line:        misconfig.CauseMetadata.StartLine,
				Resource:    misconfig.CauseMetadata.Resource,
				Type:        iacType,
				Resolution:  misconfig.Resolution,
			}
			findings = append(findings, finding)

			summary.TotalFindings++
			summary.ByType[iacType]++

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
	}

	summary.FilesScanned = len(filesSet)
	return findings, summary
}

func normalizeIaCType(checkType string) string {
	typeLower := strings.ToLower(checkType)

	switch {
	case strings.Contains(typeLower, "terraform"):
		return "terraform"
	case strings.Contains(typeLower, "kubernetes") || strings.Contains(typeLower, "k8s"):
		return "kubernetes"
	case strings.Contains(typeLower, "dockerfile") || strings.Contains(typeLower, "docker"):
		return "dockerfile"
	case strings.Contains(typeLower, "cloudformation") || strings.Contains(typeLower, "cfn"):
		return "cloudformation"
	case strings.Contains(typeLower, "helm"):
		return "helm"
	case strings.Contains(typeLower, "azure") || strings.Contains(typeLower, "arm"):
		return "azure"
	default:
		return typeLower
	}
}

func deriveCheckovSeverity(checkID string) string {
	idLower := strings.ToLower(checkID)

	criticalPatterns := []string{"public", "encrypt", "privileged", "root", "admin"}
	for _, p := range criticalPatterns {
		if strings.Contains(idLower, p) {
			return "critical"
		}
	}

	highPatterns := []string{"auth", "secret", "password", "credential", "key", "token"}
	for _, p := range highPatterns {
		if strings.Contains(idLower, p) {
			return "high"
		}
	}

	mediumPatterns := []string{"log", "monitor", "backup", "version", "ssl", "tls"}
	for _, p := range mediumPatterns {
		if strings.Contains(idLower, p) {
			return "medium"
		}
	}

	return "low"
}

// ============================================================================
// CONTAINERS FEATURE
// ============================================================================

func (s *DevOpsScanner) runContainers(ctx context.Context, opts *scanner.ScanOptions, cfg ContainersConfig) (*ContainersSummary, []ContainerFinding) {
	var findings []ContainerFinding
	summary := &ContainersSummary{
		ByImage:    make(map[string]int),
		BySeverity: make(map[string]int),
	}

	if !common.ToolExists("trivy") {
		summary.Error = "trivy not found"
		return summary, findings
	}

	dockerfiles := findDockerfiles(opts.RepoPath)
	if len(dockerfiles) == 0 {
		return summary, findings
	}

	summary.DockerfilesScanned = len(dockerfiles)

	if !cfg.ScanBaseImages {
		return summary, findings
	}

	images := extractBaseImages(dockerfiles)
	summary.ImagesScanned = len(images)

	scannedImages := make(map[string]bool)
	for _, img := range images {
		if scannedImages[img.Image] {
			continue
		}
		scannedImages[img.Image] = true

		timeout := opts.Timeout
		if timeout == 0 {
			timeout = 3 * time.Minute
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		result, err := common.RunCommand(ctx, "trivy",
			"image",
			"--format", "json",
			"--severity", "CRITICAL,HIGH,MEDIUM,LOW",
			"--quiet",
			img.Image,
		)
		cancel()

		if err != nil || result == nil {
			continue
		}

		imgFindings := parseTrivyImageOutput(result.Stdout, img)
		findings = append(findings, imgFindings...)

		for _, f := range imgFindings {
			summary.TotalFindings++
			summary.ByImage[img.Image]++
			summary.BySeverity[f.Severity]++

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
	}

	return summary, findings
}

type imageRef struct {
	Image      string
	Dockerfile string
	Line       int
}

func findDockerfiles(repoPath string) []string {
	var dockerfiles []string

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		name := info.Name()
		if name == "Dockerfile" || strings.HasPrefix(name, "Dockerfile.") ||
			strings.HasSuffix(name, ".Dockerfile") {
			dockerfiles = append(dockerfiles, path)
		}

		return nil
	})

	return dockerfiles
}

func extractBaseImages(dockerfiles []string) []imageRef {
	var images []imageRef
	fromRE := regexp.MustCompile(`(?i)^FROM\s+([^\s]+)`)

	for _, df := range dockerfiles {
		file, err := os.Open(df)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())

			matches := fromRE.FindStringSubmatch(line)
			if len(matches) > 1 {
				image := matches[1]

				if strings.Contains(image, "$") || image == "scratch" {
					continue
				}

				images = append(images, imageRef{
					Image:      image,
					Dockerfile: df,
					Line:       lineNum,
				})
			}
		}
		file.Close()
	}

	return images
}

func parseTrivyImageOutput(data []byte, imgRef imageRef) []ContainerFinding {
	var findings []ContainerFinding

	type trivyImageOutput struct {
		Results []struct {
			Target          string `json:"Target"`
			Vulnerabilities []struct {
				VulnerabilityID  string `json:"VulnerabilityID"`
				PkgName          string `json:"PkgName"`
				InstalledVersion string `json:"InstalledVersion"`
				FixedVersion     string `json:"FixedVersion"`
				Title            string `json:"Title"`
				Description      string `json:"Description"`
				Severity         string `json:"Severity"`
				References       []string `json:"References"`
				CVSS             map[string]struct {
					V3Score float64 `json:"V3Score"`
				} `json:"CVSS"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}

	var output trivyImageOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return findings
	}

	for _, result := range output.Results {
		for _, vuln := range result.Vulnerabilities {
			severity := strings.ToLower(vuln.Severity)

			var cvss float64
			for _, v := range vuln.CVSS {
				if v.V3Score > cvss {
					cvss = v.V3Score
				}
			}

			finding := ContainerFinding{
				VulnID:       vuln.VulnerabilityID,
				Title:        vuln.Title,
				Description:  vuln.Description,
				Severity:     severity,
				Image:        imgRef.Image,
				Dockerfile:   imgRef.Dockerfile,
				Package:      vuln.PkgName,
				Version:      vuln.InstalledVersion,
				FixedVersion: vuln.FixedVersion,
				CVSS:         cvss,
				References:   vuln.References,
			}
			findings = append(findings, finding)
		}
	}

	return findings
}

// ============================================================================
// GITHUB ACTIONS FEATURE
// ============================================================================

func (s *DevOpsScanner) runGitHubActions(ctx context.Context, opts *scanner.ScanOptions, cfg GitHubActionsConfig) (*GitHubActionsSummary, []GitHubActionsFinding) {
	var findings []GitHubActionsFinding
	summary := &GitHubActionsSummary{
		ByCategory: make(map[string]int),
	}

	workflowsDir := filepath.Join(opts.RepoPath, ".github", "workflows")
	if _, err := os.Stat(workflowsDir); os.IsNotExist(err) {
		return summary, findings
	}

	var workflows []string
	filepath.Walk(workflowsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
			workflows = append(workflows, path)
		}
		return nil
	})

	summary.WorkflowsScanned = len(workflows)

	for _, wf := range workflows {
		wfFindings := scanWorkflowFile(wf, opts.RepoPath, cfg)
		findings = append(findings, wfFindings...)
	}

	for _, f := range findings {
		summary.TotalFindings++
		summary.ByCategory[f.Category]++

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

	return summary, findings
}

var ghaPatterns = []struct {
	pattern    *regexp.Regexp
	category   string
	severity   string
	title      string
	suggestion string
	enabled    func(GitHubActionsConfig) bool
}{
	{
		regexp.MustCompile(`uses:\s*[^@]+@[a-zA-Z0-9._-]+\s*$`),
		"unpinned-action",
		"high",
		"Action not pinned to SHA",
		"Pin action to a specific commit SHA for security",
		func(cfg GitHubActionsConfig) bool { return cfg.CheckPinning },
	},
	{
		regexp.MustCompile(`\$\{\{\s*secrets\.[^}]+\s*\}\}.*run:`),
		"secret-in-run",
		"high",
		"Secret may be exposed in run command",
		"Pass secrets through environment variables, not directly in run",
		func(cfg GitHubActionsConfig) bool { return cfg.CheckSecrets },
	},
	{
		regexp.MustCompile(`\$\{\{\s*github\.event\.(issue|pull_request|comment)\..*\}\}`),
		"injection-risk",
		"critical",
		"Potential command injection from untrusted input",
		"Sanitize or use intermediate environment variable",
		func(cfg GitHubActionsConfig) bool { return cfg.CheckInjection },
	},
	{
		regexp.MustCompile(`permissions:\s*write-all`),
		"excessive-permissions",
		"high",
		"Write-all permissions granted",
		"Use minimal required permissions",
		func(cfg GitHubActionsConfig) bool { return cfg.CheckPermissions },
	},
	{
		regexp.MustCompile(`contents:\s*write`),
		"write-permissions",
		"medium",
		"Contents write permission granted",
		"Ensure write permission is necessary",
		func(cfg GitHubActionsConfig) bool { return cfg.CheckPermissions },
	},
}

func scanWorkflowFile(path, repoPath string, cfg GitHubActionsConfig) []GitHubActionsFinding {
	var findings []GitHubActionsFinding

	file, err := os.Open(path)
	if err != nil {
		return findings
	}
	defer file.Close()

	relPath := path
	if strings.HasPrefix(path, repoPath) {
		relPath = strings.TrimPrefix(path, repoPath+"/")
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, pat := range ghaPatterns {
			if !pat.enabled(cfg) {
				continue
			}

			if pat.pattern.MatchString(line) {
				findings = append(findings, GitHubActionsFinding{
					RuleID:      pat.category,
					Title:       pat.title,
					Description: pat.title,
					Severity:    pat.severity,
					File:        relPath,
					Line:        lineNum,
					Category:    pat.category,
					Suggestion:  pat.suggestion,
				})
			}
		}
	}

	return findings
}

// ============================================================================
// DORA FEATURE
// ============================================================================

var (
	releaseTagPattern = regexp.MustCompile(`^v?\d+\.\d+(\.\d+)?(-.*)?$`)
	fixPatterns       = regexp.MustCompile(`(?i)(fix|hotfix|patch|bugfix)`)
)

func (s *DevOpsScanner) runDORA(ctx context.Context, opts *scanner.ScanOptions, cfg DORAConfig) (*DORASummary, *DORAMetrics) {
	summary := &DORASummary{
		PeriodDays: cfg.PeriodDays,
	}

	repo, err := git.PlainOpen(opts.RepoPath)
	if err != nil {
		summary.Error = "failed to open repository"
		return summary, nil
	}

	if cfg.PeriodDays == 0 {
		cfg.PeriodDays = 90
	}

	now := time.Now()
	since := now.AddDate(0, 0, -cfg.PeriodDays)

	metrics := calculateDORAMetrics(repo, since, now)

	summary.DeploymentFrequency = metrics.DeploymentFrequency
	summary.DeploymentFrequencyClass = classifyDeploymentFrequency(metrics.DeploymentFrequency)
	summary.LeadTimeHours = metrics.LeadTimeHours
	summary.LeadTimeClass = classifyLeadTime(metrics.LeadTimeHours)
	summary.ChangeFailureRate = metrics.ChangeFailureRate
	summary.ChangeFailureClass = classifyChangeFailureRate(metrics.ChangeFailureRate)
	summary.MTTRHours = metrics.MTTRHours
	summary.MTTRClass = classifyMTTR(metrics.MTTRHours)
	summary.OverallClass = calculateOverallClass(metrics)
	summary.PeriodDays = cfg.PeriodDays

	return summary, metrics
}

func calculateDORAMetrics(repo *git.Repository, since, until time.Time) *DORAMetrics {
	metrics := &DORAMetrics{}

	tags, _ := repo.Tags()
	var deployments []Deployment

	tags.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		if !releaseTagPattern.MatchString(tagName) {
			return nil
		}

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			tagObj, err := repo.TagObject(ref.Hash())
			if err != nil {
				return nil
			}
			commit, err = tagObj.Commit()
			if err != nil {
				return nil
			}
		}

		if commit.Author.When.After(since) && commit.Author.When.Before(until) {
			isFix := fixPatterns.MatchString(tagName) || fixPatterns.MatchString(commit.Message)
			deployments = append(deployments, Deployment{
				Tag:   tagName,
				Date:  commit.Author.When,
				IsFix: isFix,
			})
		}
		return nil
	})

	// If no tags, use commits as proxy
	if len(deployments) == 0 {
		ref, err := repo.Head()
		if err == nil {
			commitIter, _ := repo.Log(&git.LogOptions{From: ref.Hash()})
			var commits []*object.Commit
			commitIter.ForEach(func(c *object.Commit) error {
				if c.Author.When.After(since) && c.Author.When.Before(until) {
					commits = append(commits, c)
				}
				return nil
			})

			weekCommits := make(map[int][]*object.Commit)
			for _, c := range commits {
				_, week := c.Author.When.ISOWeek()
				weekCommits[week] = append(weekCommits[week], c)
			}

			for _, wc := range weekCommits {
				if len(wc) > 0 {
					hasFix := false
					for _, c := range wc {
						if fixPatterns.MatchString(c.Message) {
							hasFix = true
							break
						}
					}
					deployments = append(deployments, Deployment{
						Tag:     "weekly-deployment",
						Date:    wc[0].Author.When,
						Commits: len(wc),
						IsFix:   hasFix,
					})
				}
			}
			metrics.TotalCommits = len(commits)
		}
	}

	metrics.TotalDeployments = len(deployments)
	metrics.Deployments = deployments

	weeks := until.Sub(since).Hours() / (24 * 7)
	if weeks > 0 {
		metrics.DeploymentFrequency = float64(len(deployments)) / weeks
	}

	if len(deployments) > 1 {
		var totalLeadTime float64
		for i := 1; i < len(deployments); i++ {
			leadTime := deployments[i-1].Date.Sub(deployments[i].Date).Hours()
			totalLeadTime += leadTime
		}
		metrics.LeadTimeHours = totalLeadTime / float64(len(deployments)-1)
	}

	var fixes int
	for _, d := range deployments {
		if d.IsFix {
			fixes++
		}
	}
	if len(deployments) > 0 {
		metrics.ChangeFailureRate = float64(fixes) / float64(len(deployments)) * 100
	}

	var fixDeployments []Deployment
	for _, d := range deployments {
		if d.IsFix {
			fixDeployments = append(fixDeployments, d)
		}
	}
	if len(fixDeployments) > 1 {
		var totalMTTR float64
		for i := 1; i < len(fixDeployments); i++ {
			mttr := fixDeployments[i-1].Date.Sub(fixDeployments[i].Date).Hours()
			totalMTTR += mttr
		}
		metrics.MTTRHours = totalMTTR / float64(len(fixDeployments)-1)
	}

	return metrics
}

func classifyDeploymentFrequency(freq float64) string {
	switch {
	case freq >= 7:
		return "elite"
	case freq >= 1:
		return "high"
	case freq >= 0.25:
		return "medium"
	default:
		return "low"
	}
}

func classifyLeadTime(hours float64) string {
	switch {
	case hours < 24:
		return "elite"
	case hours < 168:
		return "high"
	case hours < 720:
		return "medium"
	default:
		return "low"
	}
}

func classifyChangeFailureRate(rate float64) string {
	switch {
	case rate <= 5:
		return "elite"
	case rate <= 10:
		return "high"
	case rate <= 15:
		return "medium"
	default:
		return "low"
	}
}

func classifyMTTR(hours float64) string {
	switch {
	case hours < 1:
		return "elite"
	case hours < 24:
		return "high"
	case hours < 168:
		return "medium"
	default:
		return "low"
	}
}

func calculateOverallClass(m *DORAMetrics) string {
	classes := map[string]int{
		classifyDeploymentFrequency(m.DeploymentFrequency): 0,
		classifyLeadTime(m.LeadTimeHours):                  0,
		classifyChangeFailureRate(m.ChangeFailureRate):     0,
		classifyMTTR(m.MTTRHours):                          0,
	}

	scores := map[string]int{"elite": 4, "high": 3, "medium": 2, "low": 1}
	total := 0
	for class := range classes {
		total += scores[class]
	}

	avg := total / 4
	switch {
	case avg >= 4:
		return "elite"
	case avg >= 3:
		return "high"
	case avg >= 2:
		return "medium"
	default:
		return "low"
	}
}

// ============================================================================
// GIT FEATURE
// ============================================================================

func (s *DevOpsScanner) runGit(ctx context.Context, opts *scanner.ScanOptions, cfg GitConfig) (*GitSummary, *GitFindings) {
	summary := &GitSummary{}
	findings := &GitFindings{}

	repo, err := git.PlainOpen(opts.RepoPath)
	if err != nil {
		summary.Error = "failed to open repository"
		return summary, findings
	}

	now := time.Now()
	days30Ago := now.AddDate(0, 0, -30)
	days90Ago := now.AddDate(0, 0, -90)
	days365Ago := now.AddDate(0, 0, -365)

	commits, err := getAllCommits(repo)
	if err != nil {
		summary.Error = err.Error()
		return summary, findings
	}

	contributors, totalCommits := analyzeContributors(commits, days30Ago, days90Ago, days365Ago)
	findings.Contributors = contributors

	summary.TotalCommits = totalCommits
	summary.TotalContributors = len(contributors)
	summary.ActiveContributors30d = countActiveContributors(contributors, "30d")
	summary.ActiveContributors90d = countActiveContributors(contributors, "90d")
	summary.Commits90d = countCommitsSince(commits, days90Ago)
	summary.BusFactor = calculateBusFactor(contributors)

	if summary.Commits90d > 500 {
		summary.ActivityLevel = "very_high"
	} else if summary.Commits90d > 200 {
		summary.ActivityLevel = "high"
	} else if summary.Commits90d > 50 {
		summary.ActivityLevel = "medium"
	} else {
		summary.ActivityLevel = "low"
	}

	if cfg.IncludeChurn {
		findings.HighChurnFiles = analyzeHighChurnFiles(commits, days90Ago)
	}

	if cfg.IncludeAge {
		codeAge := analyzeCodeAge(repo, commits, now, days30Ago, days90Ago, days365Ago)
		findings.CodeAge = &codeAge
	}

	if cfg.IncludePatterns {
		patterns := analyzeCommitPatterns(commits, days90Ago)
		findings.Patterns = &patterns
	}

	if cfg.IncludeBranches {
		branches := analyzeBranches(repo)
		findings.Branches = &branches
	}

	return summary, findings
}

func getAllCommits(repo *git.Repository) ([]*object.Commit, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime})
	if err != nil {
		return []*object.Commit{headCommit}, nil
	}

	var commits []*object.Commit
	commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})

	if len(commits) == 0 {
		return []*object.Commit{headCommit}, nil
	}

	return commits, nil
}

func analyzeContributors(commits []*object.Commit, days30, days90, days365 time.Time) ([]Contributor, int) {
	type authorStats struct {
		name            string
		totalCommits    int
		commits30d      int
		commits90d      int
		commits365d     int
		linesAdded90d   int
		linesRemoved90d int
	}

	stats := make(map[string]*authorStats)

	for i, c := range commits {
		email := c.Author.Email
		if email == "" {
			email = "unknown"
		}

		if stats[email] == nil {
			stats[email] = &authorStats{name: c.Author.Name}
		}
		s := stats[email]
		s.totalCommits++

		commitTime := c.Author.When
		if commitTime.After(days30) {
			s.commits30d++
		}
		if commitTime.After(days90) {
			s.commits90d++

			if i < len(commits)-1 {
				parent := commits[i+1]
				added, removed := getCommitDiff(c, parent)
				s.linesAdded90d += added
				s.linesRemoved90d += removed
			}
		}
		if commitTime.After(days365) {
			s.commits365d++
		}
	}

	var contributors []Contributor
	for email, s := range stats {
		contributors = append(contributors, Contributor{
			Name:            s.name,
			Email:           email,
			TotalCommits:    s.totalCommits,
			Commits30d:      s.commits30d,
			Commits90d:      s.commits90d,
			Commits365d:     s.commits365d,
			LinesAdded90d:   s.linesAdded90d,
			LinesRemoved90d: s.linesRemoved90d,
		})
	}

	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].TotalCommits > contributors[j].TotalCommits
	})

	return contributors, len(commits)
}

func getCommitDiff(commit, parent *object.Commit) (added, removed int) {
	commitTree, err := commit.Tree()
	if err != nil {
		return 0, 0
	}

	parentTree, err := parent.Tree()
	if err != nil {
		return 0, 0
	}

	changes, err := parentTree.Diff(commitTree)
	if err != nil {
		return 0, 0
	}

	for _, change := range changes {
		patch, err := change.Patch()
		if err != nil {
			continue
		}

		for _, filePatch := range patch.FilePatches() {
			for _, chunk := range filePatch.Chunks() {
				lines := strings.Split(chunk.Content(), "\n")
				switch chunk.Type() {
				case 1:
					added += len(lines)
				case 2:
					removed += len(lines)
				}
			}
		}
	}

	return added, removed
}

func calculateBusFactor(contributors []Contributor) int {
	if len(contributors) == 0 {
		return 0
	}

	var totalCommits int
	for _, c := range contributors {
		totalCommits += c.TotalCommits
	}

	threshold := totalCommits / 2
	cumulative := 0
	busFactor := 0

	for _, c := range contributors {
		cumulative += c.TotalCommits
		busFactor++
		if cumulative >= threshold {
			break
		}
	}

	return busFactor
}

func countActiveContributors(contributors []Contributor, period string) int {
	count := 0
	for _, c := range contributors {
		switch period {
		case "30d":
			if c.Commits30d > 0 {
				count++
			}
		case "90d":
			if c.Commits90d > 0 {
				count++
			}
		}
	}
	return count
}

func countCommitsSince(commits []*object.Commit, since time.Time) int {
	count := 0
	for _, c := range commits {
		if c.Author.When.After(since) {
			count++
		}
	}
	return count
}

func analyzeHighChurnFiles(commits []*object.Commit, since time.Time) []ChurnFile {
	fileChanges := make(map[string]int)
	fileContributors := make(map[string]map[string]bool)

	for i, c := range commits {
		if c.Author.When.Before(since) || i >= len(commits)-1 {
			continue
		}

		parent := commits[i+1]
		commitTree, err := c.Tree()
		if err != nil {
			continue
		}
		parentTree, err := parent.Tree()
		if err != nil {
			continue
		}

		changes, err := parentTree.Diff(commitTree)
		if err != nil {
			continue
		}

		for _, change := range changes {
			name := change.To.Name
			if name == "" {
				name = change.From.Name
			}

			if strings.Contains(name, "node_modules/") ||
				strings.Contains(name, "vendor/") ||
				strings.Contains(name, ".git/") {
				continue
			}

			fileChanges[name]++

			if fileContributors[name] == nil {
				fileContributors[name] = make(map[string]bool)
			}
			fileContributors[name][c.Author.Email] = true
		}
	}

	type fileCount struct {
		file  string
		count int
	}
	var files []fileCount
	for f, c := range fileChanges {
		if c >= 5 {
			files = append(files, fileCount{f, c})
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].count > files[j].count
	})

	if len(files) > 30 {
		files = files[:30]
	}

	var churnFiles []ChurnFile
	for _, f := range files {
		churnFiles = append(churnFiles, ChurnFile{
			File:         f.file,
			Changes90d:   f.count,
			Contributors: len(fileContributors[f.file]),
		})
	}

	return churnFiles
}

func analyzeCodeAge(repo *git.Repository, commits []*object.Commit, now, days30, days90, days365 time.Time) CodeAgeStats {
	stats := CodeAgeStats{}

	fileLastMod := make(map[string]time.Time)

	for i, c := range commits {
		if i >= len(commits)-1 {
			continue
		}

		parent := commits[i+1]
		commitTree, err := c.Tree()
		if err != nil {
			continue
		}
		parentTree, err := parent.Tree()
		if err != nil {
			continue
		}

		changes, err := parentTree.Diff(commitTree)
		if err != nil {
			continue
		}

		for _, change := range changes {
			name := change.To.Name
			if name == "" || !isSourceFile(name) {
				continue
			}

			if strings.Contains(name, "node_modules/") || strings.Contains(name, "vendor/") {
				continue
			}

			if _, exists := fileLastMod[name]; !exists {
				fileLastMod[name] = c.Author.When
			}
		}
	}

	count := 0
	var age0to30, age31to90, age91to365, age365plus int

	for _, lastMod := range fileLastMod {
		if count >= 200 {
			break
		}
		count++
		stats.SampledFiles++

		if lastMod.After(days30) {
			age0to30++
		} else if lastMod.After(days90) {
			age31to90++
		} else if lastMod.After(days365) {
			age91to365++
		} else {
			age365plus++
		}
	}

	if stats.SampledFiles > 0 {
		stats.Age0to30 = AgeBucket{
			Count:      age0to30,
			Percentage: float64(age0to30*100) / float64(stats.SampledFiles),
		}
		stats.Age31to90 = AgeBucket{
			Count:      age31to90,
			Percentage: float64(age31to90*100) / float64(stats.SampledFiles),
		}
		stats.Age91to365 = AgeBucket{
			Count:      age91to365,
			Percentage: float64(age91to365*100) / float64(stats.SampledFiles),
		}
		stats.Age365Plus = AgeBucket{
			Count:      age365plus,
			Percentage: float64(age365plus*100) / float64(stats.SampledFiles),
		}
	}

	return stats
}

func isSourceFile(name string) bool {
	extensions := []string{".py", ".js", ".ts", ".tsx", ".jsx", ".java", ".go", ".rb", ".php", ".c", ".cpp", ".rs", ".swift", ".kt"}
	for _, ext := range extensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

func analyzeCommitPatterns(commits []*object.Commit, since time.Time) CommitPatterns {
	patterns := CommitPatterns{}

	if len(commits) == 0 {
		return patterns
	}

	dayCounts := make(map[string]int)
	hourCounts := make(map[int]int)
	var commitCount90d int

	for _, c := range commits {
		when := c.Author.When
		dayCounts[when.Weekday().String()]++
		hourCounts[when.Hour()]++

		if when.After(since) {
			commitCount90d++
		}
	}

	maxDayCount := 0
	for day, count := range dayCounts {
		if count > maxDayCount {
			maxDayCount = count
			patterns.MostActiveDay = day
		}
	}

	maxHourCount := 0
	for hour, count := range hourCounts {
		if count > maxHourCount {
			maxHourCount = count
			patterns.MostActiveHour = hour
		}
	}

	if len(commits) > 0 {
		patterns.LastCommit = commits[0].Author.When.Format(time.RFC3339)
		patterns.FirstCommit = commits[len(commits)-1].Author.When.Format(time.RFC3339)
	}

	if commitCount90d > 0 {
		patterns.AvgCommitsPerWeek = commitCount90d / 13
	}

	return patterns
}

func analyzeBranches(repo *git.Repository) BranchInfo {
	info := BranchInfo{
		Default: "main",
	}

	head, err := repo.Head()
	if err == nil {
		info.Current = head.Name().Short()
	}

	branches, err := repo.Branches()
	if err == nil {
		branches.ForEach(func(ref *plumbing.Reference) error {
			info.TotalCount++
			return nil
		})
	}

	remotes, err := repo.References()
	if err == nil {
		remotes.ForEach(func(ref *plumbing.Reference) error {
			if ref.Name().IsRemote() {
				info.RemoteCount++
			}
			if ref.Name().String() == "refs/remotes/origin/HEAD" {
				target := ref.Target()
				if target != "" {
					info.Default = strings.TrimPrefix(target.Short(), "origin/")
				}
			}
			return nil
		})
	}

	return info
}
