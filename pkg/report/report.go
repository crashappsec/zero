// Package report implements the report command for generating analysis reports
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/terminal"
)

// Options configures the report command
type Options struct {
	Org      string
	Repo     string
	Type     string // "summary", "security", "licenses", "sbom", "supply-chain", "full"
	Format   string // "text", "json", "markdown", "html"
	Output   string // Output file path
	Scanners []string
}

// Report handles the report command
type Report struct {
	cfg      *config.Config
	term     *terminal.Terminal
	opts     *Options
	zeroHome string
}

// New creates a new Report instance
func New(opts *Options) (*Report, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if opts.Format == "" {
		opts.Format = "text"
	}

	return &Report{
		cfg:      cfg,
		term:     terminal.New(),
		opts:     opts,
		zeroHome: cfg.ZeroHome(),
	}, nil
}

// Run executes the report command
func (r *Report) Run() error {
	if r.opts.Repo != "" {
		return r.reportSingleRepo(r.opts.Repo)
	}

	if r.opts.Org != "" {
		return r.reportOrg(r.opts.Org)
	}

	return fmt.Errorf("specify --org or --repo")
}

// reportOrg generates a report for an entire organization
func (r *Report) reportOrg(org string) error {
	orgPath := filepath.Join(r.zeroHome, "repos", org)

	if _, err := os.Stat(orgPath); os.IsNotExist(err) {
		return fmt.Errorf("organization not found: %s (run hydrate first)", org)
	}

	// Get all repos in org
	repos, err := os.ReadDir(orgPath)
	if err != nil {
		return fmt.Errorf("reading org directory: %w", err)
	}

	// Aggregate findings
	findings := &OrgFindings{
		Org:   org,
		Repos: make(map[string]*RepoFindings),
	}

	for _, repo := range repos {
		if !repo.IsDir() {
			continue
		}

		repoFindings := r.loadRepoFindings(org, repo.Name())
		if repoFindings != nil {
			findings.Repos[repo.Name()] = repoFindings
			findings.TotalVulns += repoFindings.VulnCount
			findings.TotalSecrets += repoFindings.SecretCount
			findings.TotalPackages += repoFindings.PackageCount
		}
	}

	switch r.opts.Format {
	case "json":
		return r.outputJSON(findings)
	case "markdown":
		return r.outputMarkdown(findings)
	default:
		return r.outputText(findings)
	}
}

// reportSingleRepo generates a report for a single repository
func (r *Report) reportSingleRepo(repo string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: use owner/repo")
	}

	repoFindings := r.loadRepoFindings(parts[0], parts[1])
	if repoFindings == nil {
		return fmt.Errorf("repo not found: %s (run hydrate first)", repo)
	}

	switch r.opts.Format {
	case "json":
		return r.outputJSON(repoFindings)
	case "markdown":
		return r.outputRepoMarkdown(repoFindings)
	default:
		return r.outputRepoText(repoFindings)
	}
}

// OrgFindings aggregates findings across an organization
type OrgFindings struct {
	Org           string                   `json:"org"`
	Repos         map[string]*RepoFindings `json:"repos"`
	TotalVulns    int                      `json:"total_vulns"`
	TotalSecrets  int                      `json:"total_secrets"`
	TotalPackages int                      `json:"total_packages"`
}

// RepoFindings contains findings for a single repository
type RepoFindings struct {
	Name         string         `json:"name"`
	Owner        string         `json:"owner"`
	PackageCount int            `json:"package_count"`
	VulnCount    int            `json:"vuln_count"`
	SecretCount  int            `json:"secret_count"`
	CodeIssues   int            `json:"code_issues"`
	Vulns        []Vulnerability `json:"vulnerabilities,omitempty"`
	Secrets      []Secret        `json:"secrets,omitempty"`
}

// Vulnerability represents a package vulnerability
type Vulnerability struct {
	ID       string `json:"id"`
	Package  string `json:"package"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	FixedIn  string `json:"fixed_in,omitempty"`
}

// Secret represents a detected secret
type Secret struct {
	Type     string `json:"type"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Severity string `json:"severity"`
}

// loadRepoFindings loads findings for a repository from analysis files
func (r *Report) loadRepoFindings(owner, name string) *RepoFindings {
	analysisPath := filepath.Join(r.zeroHome, "repos", owner, name, "analysis")

	if _, err := os.Stat(analysisPath); os.IsNotExist(err) {
		return nil
	}

	findings := &RepoFindings{
		Name:  name,
		Owner: owner,
	}

	// Load package vulnerabilities
	vulnsPath := filepath.Join(analysisPath, "package-vulns.json")
	if data, err := os.ReadFile(vulnsPath); err == nil {
		var vulnsData struct {
			Summary struct {
				Total    int `json:"total"`
				Critical int `json:"critical"`
				High     int `json:"high"`
				Medium   int `json:"medium"`
				Low      int `json:"low"`
			} `json:"summary"`
			Vulnerabilities []struct {
				ID       string `json:"id"`
				Package  string `json:"package"`
				Version  string `json:"version"`
				Severity string `json:"severity"`
				Summary  string `json:"summary"`
				Fixed    string `json:"fixed"`
			} `json:"vulnerabilities"`
		}
		if err := json.Unmarshal(data, &vulnsData); err == nil {
			findings.VulnCount = vulnsData.Summary.Total

			for _, v := range vulnsData.Vulnerabilities {
				findings.Vulns = append(findings.Vulns, Vulnerability{
					ID:       v.ID,
					Package:  fmt.Sprintf("%s@%s", v.Package, v.Version),
					Severity: strings.ToUpper(v.Severity),
					Title:    v.Summary,
					FixedIn:  v.Fixed,
				})
			}
		}
	}

	// Load SBOM for package count
	sbomPath := filepath.Join(analysisPath, "package-sbom.json")
	if data, err := os.ReadFile(sbomPath); err == nil {
		var sbomData struct {
			TotalDependencies int `json:"total_dependencies"`
			Summary           struct {
				Total int `json:"total"`
			} `json:"summary"`
		}
		if err := json.Unmarshal(data, &sbomData); err == nil {
			if sbomData.Summary.Total > 0 {
				findings.PackageCount = sbomData.Summary.Total
			} else if sbomData.TotalDependencies > 0 {
				findings.PackageCount = sbomData.TotalDependencies
			}
		}
	}

	// Load secrets
	secretsPath := filepath.Join(analysisPath, "code-secrets.json")
	if data, err := os.ReadFile(secretsPath); err == nil {
		var secretsData struct {
			Summary struct {
				Total int `json:"total"`
			} `json:"summary"`
			Findings []struct {
				Type     string `json:"check_id"`
				Path     string `json:"path"`
				Line     int    `json:"start_line"`
				Severity string `json:"severity"`
			} `json:"findings"`
		}
		if err := json.Unmarshal(data, &secretsData); err == nil {
			findings.SecretCount = secretsData.Summary.Total
			for _, s := range secretsData.Findings {
				findings.Secrets = append(findings.Secrets, Secret{
					Type:     s.Type,
					File:     s.Path,
					Line:     s.Line,
					Severity: s.Severity,
				})
			}
		}
	}

	// Load code vulns
	codeVulnsPath := filepath.Join(analysisPath, "code-vulns.json")
	if data, err := os.ReadFile(codeVulnsPath); err == nil {
		var codeData struct {
			Summary struct {
				Total int `json:"total"`
			} `json:"summary"`
		}
		if err := json.Unmarshal(data, &codeData); err == nil {
			findings.CodeIssues = codeData.Summary.Total
		}
	}

	return findings
}

// outputText prints the org report in text format
func (r *Report) outputText(findings *OrgFindings) error {
	r.term.Divider()
	r.term.Info("%s %s",
		r.term.Color(terminal.Bold, "Security Report:"),
		r.term.Color(terminal.Cyan, findings.Org),
	)
	r.term.Divider()
	fmt.Println()

	// Summary
	r.term.Info("%s", r.term.Color(terminal.Bold, "Summary"))
	r.term.Info("  Repositories:    %d", len(findings.Repos))
	r.term.Info("  Total packages:  %s", formatNumber(findings.TotalPackages))

	if findings.TotalVulns > 0 {
		r.term.Info("  Vulnerabilities: %s", r.term.Color(terminal.Red, fmt.Sprintf("%d", findings.TotalVulns)))
	} else {
		r.term.Info("  Vulnerabilities: %s", r.term.Color(terminal.Green, "0"))
	}

	if findings.TotalSecrets > 0 {
		r.term.Info("  Secrets:         %s", r.term.Color(terminal.Red, fmt.Sprintf("%d", findings.TotalSecrets)))
	} else {
		r.term.Info("  Secrets:         %s", r.term.Color(terminal.Green, "0"))
	}
	fmt.Println()

	// Repos with issues
	var reposWithIssues []*RepoFindings
	for _, repo := range findings.Repos {
		if repo.VulnCount > 0 || repo.SecretCount > 0 {
			reposWithIssues = append(reposWithIssues, repo)
		}
	}

	if len(reposWithIssues) > 0 {
		// Sort by vuln count
		sort.Slice(reposWithIssues, func(i, j int) bool {
			return reposWithIssues[i].VulnCount > reposWithIssues[j].VulnCount
		})

		r.term.Info("%s", r.term.Color(terminal.Bold, "Repositories with Issues"))
		for _, repo := range reposWithIssues {
			issues := []string{}
			if repo.VulnCount > 0 {
				issues = append(issues, fmt.Sprintf("%d vulns", repo.VulnCount))
			}
			if repo.SecretCount > 0 {
				issues = append(issues, fmt.Sprintf("%d secrets", repo.SecretCount))
			}
			r.term.Info("  %s %s: %s",
				r.term.Color(terminal.Yellow, "⚠"),
				repo.Name,
				strings.Join(issues, ", "),
			)
		}
		fmt.Println()
	}

	// Top vulnerabilities
	allVulns := r.collectAllVulns(findings)
	if len(allVulns) > 0 {
		r.term.Info("%s", r.term.Color(terminal.Bold, "Top Vulnerabilities"))

		// Show top 10
		count := min(10, len(allVulns))
		for i := 0; i < count; i++ {
			v := allVulns[i]
			severityColor := terminal.Dim
			switch v.Severity {
			case "CRITICAL":
				severityColor = terminal.BoldRed
			case "HIGH":
				severityColor = terminal.Red
			case "MEDIUM":
				severityColor = terminal.Yellow
			}

			r.term.Info("  %s [%s] %s in %s",
				r.term.Color(severityColor, fmt.Sprintf("%-8s", v.Severity)),
				v.ID,
				v.Package,
				v.repo,
			)
		}

		if len(allVulns) > 10 {
			r.term.Info("  %s", r.term.Color(terminal.Dim, fmt.Sprintf("... and %d more", len(allVulns)-10)))
		}
	}

	return nil
}

type vulnWithRepo struct {
	Vulnerability
	repo string
}

func (r *Report) collectAllVulns(findings *OrgFindings) []vulnWithRepo {
	var all []vulnWithRepo

	for repoName, repo := range findings.Repos {
		for _, v := range repo.Vulns {
			all = append(all, vulnWithRepo{
				Vulnerability: v,
				repo:          repoName,
			})
		}
	}

	// Sort by severity
	severityOrder := map[string]int{
		"CRITICAL": 0,
		"HIGH":     1,
		"MEDIUM":   2,
		"LOW":      3,
	}

	sort.Slice(all, func(i, j int) bool {
		return severityOrder[all[i].Severity] < severityOrder[all[j].Severity]
	})

	return all
}

// outputRepoText prints a single repo report in text format
func (r *Report) outputRepoText(findings *RepoFindings) error {
	r.term.Divider()
	r.term.Info("%s %s/%s",
		r.term.Color(terminal.Bold, "Security Report:"),
		r.term.Color(terminal.Cyan, findings.Owner),
		r.term.Color(terminal.Cyan, findings.Name),
	)
	r.term.Divider()
	fmt.Println()

	r.term.Info("%s", r.term.Color(terminal.Bold, "Summary"))
	r.term.Info("  Packages:        %d", findings.PackageCount)

	if findings.VulnCount > 0 {
		r.term.Info("  Vulnerabilities: %s", r.term.Color(terminal.Red, fmt.Sprintf("%d", findings.VulnCount)))
	} else {
		r.term.Info("  Vulnerabilities: %s", r.term.Color(terminal.Green, "0"))
	}

	if findings.SecretCount > 0 {
		r.term.Info("  Secrets:         %s", r.term.Color(terminal.Red, fmt.Sprintf("%d", findings.SecretCount)))
	} else {
		r.term.Info("  Secrets:         %s", r.term.Color(terminal.Green, "0"))
	}

	if findings.CodeIssues > 0 {
		r.term.Info("  Code issues:     %s", r.term.Color(terminal.Yellow, fmt.Sprintf("%d", findings.CodeIssues)))
	}
	fmt.Println()

	// Vulnerabilities
	if len(findings.Vulns) > 0 {
		r.term.Info("%s", r.term.Color(terminal.Bold, "Vulnerabilities"))
		for _, v := range findings.Vulns {
			severityColor := terminal.Dim
			switch v.Severity {
			case "CRITICAL":
				severityColor = terminal.BoldRed
			case "HIGH":
				severityColor = terminal.Red
			case "MEDIUM":
				severityColor = terminal.Yellow
			}

			r.term.Info("  %s [%s] %s",
				r.term.Color(severityColor, fmt.Sprintf("%-8s", v.Severity)),
				v.ID,
				v.Package,
			)
			if v.Title != "" {
				r.term.Info("           %s", r.term.Color(terminal.Dim, v.Title))
			}
			if v.FixedIn != "" {
				r.term.Info("           %s", r.term.Color(terminal.Green, fmt.Sprintf("Fixed in: %s", v.FixedIn)))
			}
		}
		fmt.Println()
	}

	// Secrets
	if len(findings.Secrets) > 0 {
		r.term.Info("%s", r.term.Color(terminal.Bold, "Secrets Detected"))
		for _, s := range findings.Secrets {
			r.term.Info("  %s %s:%d - %s",
				r.term.Color(terminal.Red, "⚠"),
				s.File,
				s.Line,
				s.Type,
			)
		}
	}

	return nil
}

// outputJSON prints findings as JSON
func (r *Report) outputJSON(findings interface{}) error {
	data, err := json.MarshalIndent(findings, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// outputMarkdown prints the org report in markdown format
func (r *Report) outputMarkdown(findings *OrgFindings) error {
	fmt.Printf("# Security Report: %s\n\n", findings.Org)

	fmt.Println("## Summary")
	fmt.Println()
	fmt.Printf("| Metric | Value |\n")
	fmt.Printf("|--------|-------|\n")
	fmt.Printf("| Repositories | %d |\n", len(findings.Repos))
	fmt.Printf("| Total Packages | %d |\n", findings.TotalPackages)
	fmt.Printf("| Vulnerabilities | %d |\n", findings.TotalVulns)
	fmt.Printf("| Secrets | %d |\n", findings.TotalSecrets)
	fmt.Println()

	// Vulnerabilities table
	allVulns := r.collectAllVulns(findings)
	if len(allVulns) > 0 {
		fmt.Println("## Vulnerabilities")
		fmt.Println()
		fmt.Println("| Severity | ID | Package | Repository |")
		fmt.Println("|----------|-----|---------|------------|")
		for _, v := range allVulns {
			fmt.Printf("| %s | %s | %s | %s |\n", v.Severity, v.ID, v.Package, v.repo)
		}
		fmt.Println()
	}

	return nil
}

// outputRepoMarkdown prints a single repo report in markdown format
func (r *Report) outputRepoMarkdown(findings *RepoFindings) error {
	fmt.Printf("# Security Report: %s/%s\n\n", findings.Owner, findings.Name)

	fmt.Println("## Summary")
	fmt.Println()
	fmt.Printf("- **Packages:** %d\n", findings.PackageCount)
	fmt.Printf("- **Vulnerabilities:** %d\n", findings.VulnCount)
	fmt.Printf("- **Secrets:** %d\n", findings.SecretCount)
	fmt.Printf("- **Code Issues:** %d\n", findings.CodeIssues)
	fmt.Println()

	if len(findings.Vulns) > 0 {
		fmt.Println("## Vulnerabilities")
		fmt.Println()
		fmt.Println("| Severity | ID | Package | Fix |")
		fmt.Println("|----------|-----|---------|-----|")
		for _, v := range findings.Vulns {
			fmt.Printf("| %s | %s | %s | %s |\n", v.Severity, v.ID, v.Package, v.FixedIn)
		}
		fmt.Println()
	}

	return nil
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d,%03d", n/1000, n%1000)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Reporter generates detailed terminal reports for hydrate output
type Reporter struct {
	term     *terminal.Terminal
	zeroHome string
}

// NewReporter creates a new Reporter for detailed output
func NewReporter(zeroHome string) *Reporter {
	return &Reporter{
		term:     terminal.New(),
		zeroHome: zeroHome,
	}
}

// GenerateReport prints detailed findings for all scanned projects
func (r *Reporter) GenerateReport(projectIDs []string) {
	fmt.Println()
	fmt.Printf("%s\n", r.term.Color(terminal.Bold+terminal.Cyan, "══════════════════════════════════════════════════════════════════════════════"))
	fmt.Printf("%s\n", r.term.Color(terminal.Bold, "                         DETAILED FINDINGS REPORT"))
	fmt.Printf("%s\n\n", r.term.Color(terminal.Bold+terminal.Cyan, "══════════════════════════════════════════════════════════════════════════════"))

	for _, projectID := range projectIDs {
		r.reportProject(projectID)
	}
}

func (r *Reporter) reportProject(projectID string) {
	analysisDir := filepath.Join(r.zeroHome, "repos", projectID, "analysis")

	// Check if analysis exists
	if _, err := os.Stat(analysisDir); os.IsNotExist(err) {
		return
	}

	fmt.Printf("\n%s %s\n", r.term.Color(terminal.Bold, "Repository:"), r.term.Color(terminal.Cyan, projectID))
	fmt.Printf("%s\n", r.term.Color(terminal.Dim, strings.Repeat("─", 78)))

	// Report each scanner's findings
	r.reportVulnerabilities(analysisDir)
	r.reportSecrets(analysisDir)
	r.reportLicenses(analysisDir)
	r.reportPackageHealth(analysisDir)
	r.reportMalcontent(analysisDir)
	r.reportCryptoCiphers(analysisDir)
	r.reportCryptoKeys(analysisDir)
	r.reportCryptoTLS(analysisDir)
	r.reportCryptoRandom(analysisDir)
}

func (r *Reporter) reportVulnerabilities(dir string) {
	path := filepath.Join(dir, "package-vulns.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			Critical int `json:"critical"`
			High     int `json:"high"`
			Medium   int `json:"medium"`
			Low      int `json:"low"`
		} `json:"summary"`
		Findings []struct {
			ID       string   `json:"id"`
			Aliases  []string `json:"aliases"`
			Package  string   `json:"package"`
			Version  string   `json:"version"`
			Severity string   `json:"severity"`
			Title    string   `json:"title"`
			FixedIn  string   `json:"fixed_in"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	total := result.Summary.Critical + result.Summary.High + result.Summary.Medium + result.Summary.Low
	if total == 0 {
		return
	}

	fmt.Printf("\n  %s %s\n", r.term.Color(terminal.Bold, "VULNERABILITIES"),
		r.term.Color(terminal.Dim, fmt.Sprintf("(%d critical, %d high, %d medium, %d low)",
			result.Summary.Critical, result.Summary.High, result.Summary.Medium, result.Summary.Low)))

	// Show critical and high only
	for _, f := range result.Findings {
		if f.Severity != "critical" && f.Severity != "high" {
			continue
		}
		sevColor := terminal.Red
		if f.Severity == "critical" {
			sevColor = terminal.BoldRed
		}

		// Find CVE ID
		cveID := f.ID
		for _, a := range f.Aliases {
			if strings.HasPrefix(a, "CVE-") {
				cveID = a
				break
			}
		}

		fix := ""
		if f.FixedIn != "" {
			fix = r.term.Color(terminal.Green, " -> "+f.FixedIn)
		}

		fmt.Printf("    %s %-15s %s@%s%s\n",
			r.term.Color(sevColor, strings.ToUpper(f.Severity[:1])),
			r.term.Color(terminal.Dim, cveID),
			f.Package, f.Version, fix)
	}
}

func (r *Reporter) reportSecrets(dir string) {
	path := filepath.Join(dir, "code-secrets.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			TotalFindings int `json:"total_findings"`
		} `json:"summary"`
		Findings []struct {
			Type     string `json:"type"`
			Severity string `json:"severity"`
			File     string `json:"file"`
			Line     int    `json:"line"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.TotalFindings == 0 {
		return
	}

	fmt.Printf("\n  %s %s\n", r.term.Color(terminal.Bold, "SECRETS"),
		r.term.Color(terminal.Dim, fmt.Sprintf("(%d findings)", result.Summary.TotalFindings)))

	for _, f := range result.Findings {
		fmt.Printf("    %s %s:%d %s\n",
			r.term.Color(terminal.Red, "!"),
			f.File, f.Line,
			r.term.Color(terminal.Dim, f.Type))
	}
}

func (r *Reporter) reportLicenses(dir string) {
	path := filepath.Join(dir, "licenses.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			UniqueLicenses int            `json:"unique_licenses"`
			LicenseCounts  map[string]int `json:"license_counts"`
			Denied         int            `json:"denied"`
			NeedsReview    int            `json:"needs_review"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.UniqueLicenses == 0 {
		return
	}

	fmt.Printf("\n  %s %s\n", r.term.Color(terminal.Bold, "LICENSES"),
		r.term.Color(terminal.Dim, fmt.Sprintf("(%d types)", result.Summary.UniqueLicenses)))

	// Sort by count
	type licCount struct {
		name  string
		count int
	}
	var counts []licCount
	for name, count := range result.Summary.LicenseCounts {
		counts = append(counts, licCount{name, count})
	}
	sort.Slice(counts, func(i, j int) bool { return counts[i].count > counts[j].count })

	// Show top 8
	shown := 0
	for _, lc := range counts {
		if shown >= 8 {
			break
		}
		fmt.Printf("    %4d  %s\n", lc.count, lc.name)
		shown++
	}
	if len(counts) > 8 {
		fmt.Printf("    %s\n", r.term.Color(terminal.Dim, fmt.Sprintf("... and %d more license types", len(counts)-8)))
	}
}

func (r *Reporter) reportPackageHealth(dir string) {
	path := filepath.Join(dir, "package-health.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			CriticalCount int `json:"critical_count"`
			WarningCount  int `json:"warning_count"`
		} `json:"summary"`
		Findings []struct {
			Package     string `json:"package"`
			Severity    string `json:"severity"`
			Issue       string `json:"issue"`
			Description string `json:"description"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.CriticalCount == 0 && result.Summary.WarningCount == 0 {
		return
	}

	fmt.Printf("\n  %s %s\n", r.term.Color(terminal.Bold, "PACKAGE HEALTH"),
		r.term.Color(terminal.Dim, fmt.Sprintf("(%d critical, %d warnings)",
			result.Summary.CriticalCount, result.Summary.WarningCount)))

	for _, f := range result.Findings {
		sevColor := terminal.Yellow
		if f.Severity == "critical" {
			sevColor = terminal.Red
		}
		fmt.Printf("    %s %s - %s\n",
			r.term.Color(sevColor, f.Issue),
			f.Package,
			r.term.Color(terminal.Dim, truncate(f.Description, 50)))
	}
}

func (r *Reporter) reportMalcontent(dir string) {
	path := filepath.Join(dir, "package-malcontent.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			Critical int `json:"critical"`
			High     int `json:"high"`
		} `json:"summary"`
		Findings []struct {
			Path      string `json:"path"`
			RiskLevel string `json:"risk_level"`
			RiskScore int    `json:"risk_score"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.Critical == 0 && result.Summary.High == 0 {
		return
	}

	fmt.Printf("\n  %s %s\n", r.term.Color(terminal.Bold, "MALCONTENT"),
		r.term.Color(terminal.Dim, fmt.Sprintf("(%d critical, %d high)",
			result.Summary.Critical, result.Summary.High)))

	for _, f := range result.Findings {
		if f.RiskLevel != "CRITICAL" && f.RiskLevel != "HIGH" {
			continue
		}
		sevColor := terminal.Red
		if f.RiskLevel == "CRITICAL" {
			sevColor = terminal.BoldRed
		}
		fmt.Printf("    %s %s (score: %d)\n",
			r.term.Color(sevColor, f.RiskLevel[:1]),
			f.Path, f.RiskScore)
	}
}

func (r *Reporter) reportCryptoCiphers(dir string) {
	path := filepath.Join(dir, "crypto-ciphers.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			TotalFindings int            `json:"total_findings"`
			ByAlgorithm   map[string]int `json:"by_algorithm"`
		} `json:"summary"`
		Findings []struct {
			Algorithm  string `json:"algorithm"`
			File       string `json:"file"`
			Line       int    `json:"line"`
			Suggestion string `json:"suggestion"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.TotalFindings == 0 {
		return
	}

	fmt.Printf("\n  %s %s\n", r.term.Color(terminal.Bold, "WEAK CIPHERS"),
		r.term.Color(terminal.Dim, fmt.Sprintf("(%d findings)", result.Summary.TotalFindings)))

	// Group by algorithm
	shown := 0
	for algo, count := range result.Summary.ByAlgorithm {
		if shown >= 5 {
			break
		}
		fmt.Printf("    %s: %d occurrences\n", r.term.Color(terminal.Yellow, algo), count)
		shown++
	}
}

func (r *Reporter) reportCryptoKeys(dir string) {
	path := filepath.Join(dir, "crypto-keys.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			TotalFindings int `json:"total_findings"`
		} `json:"summary"`
		Findings []struct {
			Type string `json:"type"`
			File string `json:"file"`
			Line int    `json:"line"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.TotalFindings == 0 {
		return
	}

	fmt.Printf("\n  %s %s\n", r.term.Color(terminal.Bold, "HARDCODED KEYS"),
		r.term.Color(terminal.Dim, fmt.Sprintf("(%d findings)", result.Summary.TotalFindings)))

	for _, f := range result.Findings {
		fmt.Printf("    %s %s:%d\n",
			r.term.Color(terminal.Red, f.Type),
			f.File, f.Line)
	}
}

func (r *Reporter) reportCryptoTLS(dir string) {
	path := filepath.Join(dir, "crypto-tls.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			TotalFindings int            `json:"total_findings"`
			BySeverity    map[string]int `json:"by_severity"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	critical := result.Summary.BySeverity["critical"]
	high := result.Summary.BySeverity["high"]

	if critical == 0 && high == 0 {
		return
	}

	fmt.Printf("\n  %s %s\n", r.term.Color(terminal.Bold, "TLS ISSUES"),
		r.term.Color(terminal.Dim, fmt.Sprintf("(%d critical, %d high)", critical, high)))
}

func (r *Reporter) reportCryptoRandom(dir string) {
	path := filepath.Join(dir, "crypto-random.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			TotalFindings int            `json:"total_findings"`
			ByType        map[string]int `json:"by_type"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.TotalFindings == 0 {
		return
	}

	fmt.Printf("\n  %s %s\n", r.term.Color(terminal.Bold, "WEAK RANDOM"),
		r.term.Color(terminal.Dim, fmt.Sprintf("(%d findings)", result.Summary.TotalFindings)))

	for typ, count := range result.Summary.ByType {
		fmt.Printf("    %s: %d occurrences\n", r.term.Color(terminal.Yellow, typ), count)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
