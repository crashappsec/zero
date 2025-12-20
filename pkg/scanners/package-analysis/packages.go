// Package packageanalysis implements the consolidated package analysis super scanner
// NOTE: This scanner DEPENDS ON the sbom scanner output. It does NOT generate SBOMs.
package packageanalysis

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/scanners/common"
	"github.com/crashappsec/zero/pkg/scanners/sbom"
)

const (
	Name    = "package-analysis"
	Version = "3.0.0"
)

func init() {
	scanner.Register(&PackagesScanner{})
}

// PackagesScanner is the consolidated package analysis super scanner
// It depends on the sbom scanner to provide package data
type PackagesScanner struct {
	config FeatureConfig
}

func (s *PackagesScanner) Name() string {
	return Name
}

func (s *PackagesScanner) Description() string {
	return "Comprehensive package and supply chain security analysis (vulnerabilities, health, malware, licenses, and more) - depends on sbom scanner"
}

func (s *PackagesScanner) Dependencies() []string {
	// This scanner depends on the sbom scanner output
	return []string{"sbom"}
}

func (s *PackagesScanner) EstimateDuration(fileCount int) time.Duration {
	base := 5 * time.Second
	if s.config.Vulns.Enabled {
		base += 5 * time.Second
	}
	if s.config.Health.Enabled {
		base += 10 * time.Second
	}
	if s.config.Malcontent.Enabled {
		base += time.Duration(fileCount/2000+2) * time.Second
	}
	if s.config.Reachability.Enabled {
		base += 30 * time.Second
	}
	return base
}

func (s *PackagesScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	// Load feature config
	if opts.FeatureConfig != nil {
		if cfg, ok := opts.FeatureConfig["packages"].(FeatureConfig); ok {
			s.config = cfg
		} else {
			s.config = DefaultConfig()
		}
	} else {
		s.config = DefaultConfig()
	}

	// Ensure output directory
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("creating output directory: %w", err)
		}
	}

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	// Load SBOM from sbom scanner output (this scanner depends on sbom)
	sbomPath := sbom.GetSBOMPath(opts.OutputDir)
	sbomData, err := loadSBOMData(sbomPath)
	if err != nil {
		result.Summary.Errors = append(result.Summary.Errors, fmt.Sprintf("failed to load SBOM: %v", err))
		// Continue anyway - some features don't need SBOM data
	}

	// Convert SBOM components to our internal format
	var components []ComponentData
	if sbomData != nil {
		for _, c := range sbomData.Components {
			var licenses []string
			for _, l := range c.Licenses {
				licenses = append(licenses, l)
			}
			components = append(components, ComponentData{
				Name:      c.Name,
				Version:   c.Version,
				Purl:      c.Purl,
				Ecosystem: c.Ecosystem,
				Licenses:  licenses,
				Scope:     c.Scope,
			})
		}
	}

	// Parallel features
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 1. Vulnerabilities
	if s.config.Vulns.Enabled && sbomPath != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vulnsResult := s.runVulnsFeature(ctx, opts, sbomPath)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "vulns")
			result.Summary.Vulns = vulnsResult.Summary
			result.Findings.Vulns = vulnsResult.Findings
			mu.Unlock()
		}()
	}

	// 2. Health
	if s.config.Health.Enabled && len(components) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			healthResult := s.runHealthFeature(ctx, components)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "health")
			result.Summary.Health = healthResult.Summary
			result.Findings.Health = healthResult.Findings
			mu.Unlock()
		}()
	}

	// 3. Licenses
	if s.config.Licenses.Enabled && sbomPath != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			licensesResult := s.runLicensesFeature(sbomPath)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "licenses")
			result.Summary.Licenses = licensesResult.Summary
			result.Findings.Licenses = licensesResult.Findings
			mu.Unlock()
		}()
	}

	// 4. Malcontent
	if s.config.Malcontent.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			malcontentResult := s.runMalcontentFeature(ctx, opts)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "malcontent")
			result.Summary.Malcontent = malcontentResult.Summary
			result.Findings.Malcontent = malcontentResult.Findings
			mu.Unlock()
		}()
	}

	// 5. Dependency Confusion
	if s.config.Confusion.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			confusionResult := s.runConfusionFeature(ctx, opts)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "confusion")
			result.Summary.Confusion = confusionResult.Summary
			result.Findings.Confusion = confusionResult.Findings
			mu.Unlock()
		}()
	}

	// 6. Typosquats
	if s.config.Typosquats.Enabled && len(components) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			typosquatsResult := s.runTyposquatsFeature(ctx, components)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "typosquats")
			result.Summary.Typosquats = typosquatsResult.Summary
			result.Findings.Typosquats = typosquatsResult.Findings
			mu.Unlock()
		}()
	}

	// 7. Deprecations
	if s.config.Deprecations.Enabled && len(components) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deprecationsResult := s.runDeprecationsFeature(ctx, components)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "deprecations")
			result.Summary.Deprecations = deprecationsResult.Summary
			result.Findings.Deprecations = deprecationsResult.Findings
			mu.Unlock()
		}()
	}

	// 8. Duplicates
	if s.config.Duplicates.Enabled && len(components) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			duplicatesResult := s.runDuplicatesFeature(components)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "duplicates")
			result.Summary.Duplicates = duplicatesResult.Summary
			result.Findings.Duplicates = duplicatesResult.Findings
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Sequential features

	// 9. Reachability
	if s.config.Reachability.Enabled {
		reachabilityResult := s.runReachabilityFeature(ctx, opts)
		result.FeaturesRun = append(result.FeaturesRun, "reachability")
		result.Summary.Reachability = reachabilityResult.Summary
		result.Findings.Reachability = reachabilityResult.Findings
	}

	// 10. Provenance
	if s.config.Provenance.Enabled && len(components) > 0 {
		provenanceResult := s.runProvenanceFeature(ctx, components)
		result.FeaturesRun = append(result.FeaturesRun, "provenance")
		result.Summary.Provenance = provenanceResult.Summary
		result.Findings.Provenance = provenanceResult.Findings
	}

	// 11. Bundle (npm only)
	if s.config.Bundle.Enabled {
		bundleResult := s.runBundleFeature(ctx, opts)
		if bundleResult != nil {
			result.FeaturesRun = append(result.FeaturesRun, "bundle")
			result.Summary.Bundle = bundleResult.Summary
			result.Findings.Bundle = bundleResult.Findings
		}
	}

	// 12. Recommendations
	if s.config.Recommendations.Enabled {
		recommendationsResult := s.runRecommendationsFeature(result)
		result.FeaturesRun = append(result.FeaturesRun, "recommendations")
		result.Summary.Recommendations = recommendationsResult.Summary
		result.Findings.Recommendations = recommendationsResult.Findings
	}

	// Create scan result
	scanResult := scanner.NewScanResult(Name, Version, start)
	scanResult.Repository = opts.RepoPath
	scanResult.SetSummary(result.Summary)
	scanResult.SetFindings(result.Findings)
	scanResult.SetMetadata(map[string]interface{}{
		"features_run":   result.FeaturesRun,
		"sbom_source":    "sbom scanner",
		"component_count": len(components),
	})

	// Write result
	if opts.OutputDir != "" {
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

// loadSBOMData loads and parses SBOM data from the sbom scanner output
func loadSBOMData(sbomPath string) (*sbom.GenerationFindings, error) {
	return sbom.LoadSBOM(sbomPath)
}

// ==================== Vulns Feature ====================

type vulnsFeatureResult struct {
	Summary  *VulnsSummary
	Findings []VulnFinding
}

func (s *PackagesScanner) runVulnsFeature(ctx context.Context, opts *scanner.ScanOptions, sbomPath string) *vulnsFeatureResult {
	result := &vulnsFeatureResult{
		Summary:  &VulnsSummary{},
		Findings: []VulnFinding{},
	}

	if !common.ToolExists("osv-scanner") {
		result.Summary.Error = "osv-scanner not installed"
		return result
	}

	// Run osv-scanner against the SBOM
	var cmdResult *common.CommandResult
	if _, err := os.Stat(sbomPath); err == nil {
		cmdResult, _ = common.RunCommand(ctx, "osv-scanner", "scan", "source", "--format=json", "-S", sbomPath)
	} else {
		cmdResult, _ = common.RunCommand(ctx, "osv-scanner", "scan", "source", "--format=json", "-r", opts.RepoPath)
	}

	if cmdResult == nil || len(cmdResult.Stdout) == 0 {
		return result
	}

	// Parse output
	var output struct {
		Results []struct {
			Packages []struct {
				Package struct {
					Name      string `json:"name"`
					Version   string `json:"version"`
					Ecosystem string `json:"ecosystem"`
				} `json:"package"`
				Vulnerabilities []struct {
					ID       string `json:"id"`
					Aliases  []string `json:"aliases"`
					Summary  string `json:"summary"`
					Severity []struct {
						Type  string `json:"type"`
						Score string `json:"score"`
					} `json:"severity"`
				} `json:"vulnerabilities"`
			} `json:"packages"`
		} `json:"results"`
	}

	if json.Unmarshal(cmdResult.Stdout, &output) != nil {
		return result
	}

	seen := make(map[string]bool)
	for _, r := range output.Results {
		for _, pkg := range r.Packages {
			for _, vuln := range pkg.Vulnerabilities {
				key := fmt.Sprintf("%s:%s:%s", vuln.ID, pkg.Package.Name, pkg.Package.Version)
				if seen[key] {
					continue
				}
				seen[key] = true

				severity := "medium"
				for _, sev := range vuln.Severity {
					if sev.Type == "CVSS_V3" {
						var score float64
						fmt.Sscanf(sev.Score, "%f", &score)
						if score >= 9.0 {
							severity = "critical"
						} else if score >= 7.0 {
							severity = "high"
						} else if score >= 4.0 {
							severity = "medium"
						} else {
							severity = "low"
						}
						break
					}
				}

				result.Findings = append(result.Findings, VulnFinding{
					ID:        vuln.ID,
					Aliases:   vuln.Aliases,
					Package:   pkg.Package.Name,
					Version:   pkg.Package.Version,
					Ecosystem: pkg.Package.Ecosystem,
					Severity:  severity,
					Title:     vuln.Summary,
				})

				result.Summary.TotalVulnerabilities++
				switch severity {
				case "critical":
					result.Summary.Critical++
				case "high":
					result.Summary.High++
				case "medium":
					result.Summary.Medium++
				case "low":
					result.Summary.Low++
				}
			}
		}
	}

	// KEV enrichment
	if s.config.Vulns.IncludeKEV {
		kevVulns := fetchKEV(ctx)
		for i := range result.Findings {
			if kevVulns[result.Findings[i].ID] {
				result.Findings[i].InKEV = true
				result.Summary.KEVCount++
			} else {
				for _, alias := range result.Findings[i].Aliases {
					if kevVulns[alias] {
						result.Findings[i].InKEV = true
						result.Summary.KEVCount++
						break
					}
				}
			}
		}
	}

	return result
}

func fetchKEV(ctx context.Context) map[string]bool {
	vulns := make(map[string]bool)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json", nil)
	resp, err := client.Do(req)
	if err != nil {
		return vulns
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var catalog struct {
		Vulnerabilities []struct {
			CVEID string `json:"cveID"`
		} `json:"vulnerabilities"`
	}
	if json.Unmarshal(body, &catalog) == nil {
		for _, v := range catalog.Vulnerabilities {
			vulns[v.CVEID] = true
		}
	}
	return vulns
}

// ==================== Health Feature ====================

type healthFeatureResult struct {
	Summary  *HealthSummary
	Findings []HealthFinding
}

func (s *PackagesScanner) runHealthFeature(ctx context.Context, components []ComponentData) *healthFeatureResult {
	result := &healthFeatureResult{
		Summary:  &HealthSummary{},
		Findings: []HealthFinding{},
	}

	client := &http.Client{Timeout: 5 * time.Second}
	maxPackages := s.config.Health.MaxPackages
	if maxPackages == 0 {
		maxPackages = 50
	}

	packages := components
	if len(packages) > maxPackages {
		packages = packages[:maxPackages]
	}

	result.Summary.TotalPackages = len(packages)

	for _, pkg := range packages {
		if pkg.Purl == "" {
			continue
		}

		finding := HealthFinding{
			Package:   pkg.Name,
			Version:   pkg.Version,
			Ecosystem: pkg.Ecosystem,
			Purl:      pkg.Purl,
		}

		apiURL := fmt.Sprintf("https://api.deps.dev/v3alpha/purl/%s", url.PathEscape(pkg.Purl))
		req, _ := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		resp, err := client.Do(req)
		if err != nil {
			finding.Status = "unknown"
			result.Findings = append(result.Findings, finding)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			var data struct {
				Version struct {
					IsDeprecated bool   `json:"isDeprecated"`
					PublishedAt  string `json:"publishedAt"`
				} `json:"version"`
				DefaultVersion string `json:"defaultVersion"`
				Project        struct {
					Scorecard struct {
						OverallScore float64 `json:"overallScore"`
					} `json:"scorecard"`
				} `json:"project"`
			}
			if json.Unmarshal(body, &data) == nil {
				finding.IsDeprecated = data.Version.IsDeprecated
				finding.LatestVersion = data.DefaultVersion
				finding.HealthScore = data.Project.Scorecard.OverallScore
				finding.IsOutdated = data.DefaultVersion != "" && data.DefaultVersion != pkg.Version

				if finding.IsDeprecated {
					finding.Status = "critical"
					result.Summary.CriticalCount++
					result.Summary.DeprecatedCount++
				} else if finding.HealthScore < 5 && finding.HealthScore > 0 {
					finding.Status = "warning"
					result.Summary.WarningCount++
				} else {
					finding.Status = "healthy"
					result.Summary.HealthyCount++
				}
				if finding.IsOutdated {
					result.Summary.OutdatedCount++
				}
				result.Summary.AnalyzedCount++
			}
		}
		resp.Body.Close()
		result.Findings = append(result.Findings, finding)
	}

	return result
}

// ==================== Licenses Feature ====================

type licensesFeatureResult struct {
	Summary  *LicensesSummary
	Findings []LicenseFinding
}

var (
	allowedLicenses = map[string]bool{
		"MIT": true, "Apache-2.0": true, "BSD-2-Clause": true, "BSD-3-Clause": true,
		"ISC": true, "Unlicense": true, "CC0-1.0": true, "0BSD": true,
	}
	deniedLicenses = map[string]bool{
		"GPL-2.0": true, "GPL-2.0-only": true, "GPL-3.0": true, "GPL-3.0-only": true,
		"AGPL-3.0": true, "AGPL-3.0-only": true, "SSPL-1.0": true,
	}
)

func (s *PackagesScanner) runLicensesFeature(sbomPath string) *licensesFeatureResult {
	result := &licensesFeatureResult{
		Summary:  &LicensesSummary{LicenseCounts: make(map[string]int)},
		Findings: []LicenseFinding{},
	}

	data, err := os.ReadFile(sbomPath)
	if err != nil {
		result.Summary.Error = err.Error()
		return result
	}

	var sbomDoc struct {
		Components []struct {
			Name     string `json:"name"`
			Version  string `json:"version"`
			Purl     string `json:"purl"`
			Licenses []struct {
				License struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"license"`
				Expression string `json:"expression"`
			} `json:"licenses"`
		} `json:"components"`
	}

	if json.Unmarshal(data, &sbomDoc) != nil {
		return result
	}

	uniqueLicenses := make(map[string]bool)

	for _, c := range sbomDoc.Components {
		result.Summary.TotalPackages++

		var licenseIDs []string
		for _, lic := range c.Licenses {
			if lic.License.ID != "" {
				licenseIDs = append(licenseIDs, lic.License.ID)
			} else if lic.License.Name != "" {
				licenseIDs = append(licenseIDs, lic.License.Name)
			} else if lic.Expression != "" {
				licenseIDs = append(licenseIDs, lic.Expression)
			}
		}

		status := "unknown"
		if len(licenseIDs) == 0 {
			result.Summary.Unknown++
		} else {
			hasDenied := false
			allAllowed := true

			for _, lic := range licenseIDs {
				uniqueLicenses[lic] = true
				result.Summary.LicenseCounts[lic]++

				if deniedLicenses[lic] {
					hasDenied = true
					allAllowed = false
				} else if !allowedLicenses[lic] {
					allAllowed = false
				}
			}

			if hasDenied {
				status = "denied"
				result.Summary.Denied++
				result.Summary.PolicyViolations++
			} else if allAllowed {
				status = "allowed"
				result.Summary.Allowed++
			} else {
				status = "review"
				result.Summary.NeedsReview++
			}
		}

		result.Findings = append(result.Findings, LicenseFinding{
			Package:   c.Name,
			Version:   c.Version,
			Ecosystem: extractEcosystem(c.Purl),
			Licenses:  licenseIDs,
			Status:    status,
		})
	}

	result.Summary.UniqueLicenses = len(uniqueLicenses)
	return result
}

// ==================== Malcontent Feature ====================

type malcontentFeatureResult struct {
	Summary  *MalcontentSummary
	Findings []MalcontentFinding
}

func (s *PackagesScanner) runMalcontentFeature(ctx context.Context, opts *scanner.ScanOptions) *malcontentFeatureResult {
	result := &malcontentFeatureResult{
		Summary:  &MalcontentSummary{},
		Findings: []MalcontentFinding{},
	}

	if !common.ToolExists("mal") {
		result.Summary.Error = "malcontent not installed (brew install malcontent)"
		return result
	}

	minRisk := s.config.Malcontent.MinRiskLevel
	if minRisk == "" {
		minRisk = "medium"
	}

	cmdResult, err := common.RunCommand(ctx, "mal", "analyze", "--format=json", "--min-file-risk="+minRisk, opts.RepoPath)
	if err != nil || cmdResult == nil {
		result.Summary.Error = "malcontent execution failed"
		return result
	}

	var output struct {
		Files map[string]struct {
			RiskScore int    `json:"risk_score"`
			RiskLevel string `json:"risk_level"`
			Behaviors []struct {
				Description string `json:"description"`
				RiskLevel   string `json:"risk_level"`
			} `json:"behaviors"`
		} `json:"files"`
	}

	if json.Unmarshal(cmdResult.Stdout, &output) != nil {
		return result
	}

	result.Summary.TotalFiles = len(output.Files)

	for path, file := range output.Files {
		if file.RiskScore == 0 {
			continue
		}
		result.Summary.FilesWithRisk++

		var behaviors []string
		for _, b := range file.Behaviors {
			behaviors = append(behaviors, b.Description)
			result.Summary.TotalFindings++
			switch strings.ToLower(b.RiskLevel) {
			case "critical":
				result.Summary.Critical++
			case "high":
				result.Summary.High++
			case "medium":
				result.Summary.Medium++
			case "low":
				result.Summary.Low++
			}
		}

		result.Findings = append(result.Findings, MalcontentFinding{
			File:      path,
			Risk:      file.RiskLevel,
			RiskScore: file.RiskScore,
			Behaviors: behaviors,
		})
	}

	return result
}

// ==================== Confusion Feature ====================

type confusionFeatureResult struct {
	Summary  *ConfusionSummary
	Findings []ConfusionFinding
}

func (s *PackagesScanner) runConfusionFeature(ctx context.Context, opts *scanner.ScanOptions) *confusionFeatureResult {
	result := &confusionFeatureResult{
		Summary:  &ConfusionSummary{ByEcosystem: make(map[string]int)},
		Findings: []ConfusionFinding{},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// Check npm
	if s.config.Confusion.CheckNPM {
		pkgJSONs := findFiles(opts.RepoPath, "package.json")
		for _, pkgJSON := range pkgJSONs {
			data, _ := os.ReadFile(pkgJSON)
			var pkg struct {
				Dependencies    map[string]string `json:"dependencies"`
				DevDependencies map[string]string `json:"devDependencies"`
			}
			if json.Unmarshal(data, &pkg) != nil {
				continue
			}

			allDeps := make(map[string]string)
			for k, v := range pkg.Dependencies {
				allDeps[k] = v
			}
			for k, v := range pkg.DevDependencies {
				allDeps[k] = v
			}

			for name, version := range allDeps {
				if looksInternal(name) {
					exists, pubVersion := checkNPMPackage(ctx, client, name)
					if exists {
						result.Findings = append(result.Findings, ConfusionFinding{
							Package:       name,
							Ecosystem:     "npm",
							RiskLevel:     "critical",
							RiskType:      "dependency-confusion",
							Description:   "Internal-looking package exists on public npm",
							PublicExists:  true,
							PublicVersion: pubVersion,
							LocalVersion:  version,
							File:          strings.TrimPrefix(pkgJSON, opts.RepoPath+"/"),
						})
						result.Summary.Critical++
						result.Summary.TotalFindings++
						result.Summary.ByEcosystem["npm"]++
					}
				}
			}
		}
	}

	// Check PyPI
	if s.config.Confusion.CheckPyPI {
		reqFiles := findFiles(opts.RepoPath, "requirements.txt")
		for _, reqFile := range reqFiles {
			f, _ := os.Open(reqFile)
			if f == nil {
				continue
			}
			scanner := bufio.NewScanner(f)
			pkgRegex := regexp.MustCompile(`^([a-zA-Z0-9][\w\-\.]*[a-zA-Z0-9])`)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				matches := pkgRegex.FindStringSubmatch(line)
				if len(matches) >= 2 && looksInternal(matches[1]) {
					exists, pubVersion := checkPyPIPackage(ctx, client, matches[1])
					if exists {
						result.Findings = append(result.Findings, ConfusionFinding{
							Package:       matches[1],
							Ecosystem:     "pypi",
							RiskLevel:     "critical",
							RiskType:      "dependency-confusion",
							Description:   "Internal-looking package exists on public PyPI",
							PublicExists:  true,
							PublicVersion: pubVersion,
							File:          strings.TrimPrefix(reqFile, opts.RepoPath+"/"),
						})
						result.Summary.Critical++
						result.Summary.TotalFindings++
						result.Summary.ByEcosystem["pypi"]++
					}
				}
			}
			f.Close()
		}
	}

	return result
}

func looksInternal(name string) bool {
	patterns := []string{"internal-", "private-", "-internal", "-private"}
	nameLower := strings.ToLower(name)
	for _, p := range patterns {
		if strings.Contains(nameLower, p) {
			return true
		}
	}
	return false
}

func checkNPMPackage(ctx context.Context, client *http.Client, name string) (bool, string) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://registry.npmjs.org/"+name, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		DistTags struct {
			Latest string `json:"latest"`
		} `json:"dist-tags"`
	}
	json.Unmarshal(body, &pkg)
	return true, pkg.DistTags.Latest
}

func checkPyPIPackage(ctx context.Context, client *http.Client, name string) (bool, string) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://pypi.org/pypi/"+name+"/json", nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	json.Unmarshal(body, &pkg)
	return true, pkg.Info.Version
}

// ==================== Reachability Feature ====================

type reachabilityFeatureResult struct {
	Summary  *ReachabilitySummary
	Findings []ReachabilityFinding
}

func (s *PackagesScanner) runReachabilityFeature(ctx context.Context, opts *scanner.ScanOptions) *reachabilityFeatureResult {
	result := &reachabilityFeatureResult{
		Summary:  &ReachabilitySummary{},
		Findings: []ReachabilityFinding{},
	}

	if !common.ToolExists("osv-scanner") {
		result.Summary.Error = "osv-scanner not installed"
		return result
	}

	// Detect ecosystem
	ecosystem := ""
	if _, err := os.Stat(filepath.Join(opts.RepoPath, "go.mod")); err == nil {
		ecosystem = "Go"
	} else if _, err := os.Stat(filepath.Join(opts.RepoPath, "requirements.txt")); err == nil {
		ecosystem = "Python"
	} else if _, err := os.Stat(filepath.Join(opts.RepoPath, "Cargo.toml")); err == nil {
		ecosystem = "Rust"
	}

	if ecosystem == "" {
		result.Summary.Supported = false
		return result
	}

	result.Summary.Supported = true

	cmdResult, _ := common.RunCommand(ctx, "osv-scanner", "scan", "--format", "json", "--experimental-call-analysis", opts.RepoPath)
	if cmdResult == nil || len(cmdResult.Stdout) == 0 {
		return result
	}

	var output struct {
		Results []struct {
			Packages []struct {
				Package struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				} `json:"package"`
				Vulnerabilities []struct {
					ID       string `json:"id"`
					Summary  string `json:"summary"`
					Analysis *struct {
						Called bool `json:"called"`
					} `json:"analysis,omitempty"`
				} `json:"vulnerabilities"`
			} `json:"packages"`
		} `json:"results"`
	}

	if json.Unmarshal(cmdResult.Stdout, &output) != nil {
		return result
	}

	for _, r := range output.Results {
		for _, pkg := range r.Packages {
			for _, vuln := range pkg.Vulnerabilities {
				status := "unknown"
				reachable := true
				if vuln.Analysis != nil {
					if vuln.Analysis.Called {
						status = "reachable"
					} else {
						status = "unreachable"
						reachable = false
					}
				}

				result.Findings = append(result.Findings, ReachabilityFinding{
					ID:                 vuln.ID,
					Package:            pkg.Package.Name,
					Version:            pkg.Package.Version,
					Summary:            vuln.Summary,
					ReachabilityStatus: status,
					Reachable:          reachable,
				})

				result.Summary.TotalVulns++
				switch status {
				case "reachable":
					result.Summary.ReachableVulns++
				case "unreachable":
					result.Summary.UnreachableVulns++
				default:
					result.Summary.UnknownReachability++
				}
			}
		}
	}

	if result.Summary.TotalVulns > 0 {
		result.Summary.ReductionPercent = float64(result.Summary.UnreachableVulns) / float64(result.Summary.TotalVulns) * 100
	}

	return result
}

// ==================== Provenance Feature ====================

type provenanceFeatureResult struct {
	Summary  *ProvenanceSummary
	Findings []ProvenanceFinding
}

func (s *PackagesScanner) runProvenanceFeature(ctx context.Context, components []ComponentData) *provenanceFeatureResult {
	result := &provenanceFeatureResult{
		Summary:  &ProvenanceSummary{},
		Findings: []ProvenanceFinding{},
	}

	result.Summary.TotalPackages = len(components)
	// Placeholder - full implementation would check npm provenance, sigstore, etc.
	result.Summary.UnverifiedCount = result.Summary.TotalPackages

	return result
}

// ==================== Bundle Feature ====================

type bundleFeatureResult struct {
	Summary  *BundleSummary
	Findings []BundleFinding
}

func (s *PackagesScanner) runBundleFeature(ctx context.Context, opts *scanner.ScanOptions) *bundleFeatureResult {
	// Only for npm projects
	if _, err := os.Stat(filepath.Join(opts.RepoPath, "package.json")); err != nil {
		return nil
	}

	result := &bundleFeatureResult{
		Summary:  &BundleSummary{},
		Findings: []BundleFinding{},
	}

	// Placeholder - full implementation would use bundlephobia API
	return result
}

// ==================== Recommendations Feature ====================

type recommendationsFeatureResult struct {
	Summary  *RecommendationsSummary
	Findings []RecommendationFinding
}

func (s *PackagesScanner) runRecommendationsFeature(scanResult *Result) *recommendationsFeatureResult {
	result := &recommendationsFeatureResult{
		Summary:  &RecommendationsSummary{},
		Findings: []RecommendationFinding{},
	}

	// Generate recommendations based on vulns and health data
	if scanResult.Summary.Vulns != nil && scanResult.Summary.Vulns.Critical > 0 {
		result.Summary.SecurityRecommendations = scanResult.Summary.Vulns.Critical
	}
	if scanResult.Summary.Health != nil && scanResult.Summary.Health.DeprecatedCount > 0 {
		result.Summary.HealthRecommendations = scanResult.Summary.Health.DeprecatedCount
	}

	result.Summary.TotalRecommendations = result.Summary.SecurityRecommendations + result.Summary.HealthRecommendations

	return result
}

// ==================== Typosquats Feature ====================

type typosquatsFeatureResult struct {
	Summary  *TyposquatsSummary
	Findings []TyposquatFinding
}

func (s *PackagesScanner) runTyposquatsFeature(ctx context.Context, components []ComponentData) *typosquatsFeatureResult {
	result := &typosquatsFeatureResult{
		Summary:  &TyposquatsSummary{},
		Findings: []TyposquatFinding{},
	}

	result.Summary.TotalChecked = len(components)

	// Check for packages similar to popular ones
	popularPackages := map[string]bool{
		"lodash": true, "express": true, "react": true, "axios": true,
		"moment": true, "request": true, "async": true, "chalk": true,
		"commander": true, "debug": true, "underscore": true, "bluebird": true,
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, pkg := range components {
		// Check similar names
		if s.config.Typosquats.CheckSimilarNames {
			for popular := range popularPackages {
				if pkg.Name != popular && isSimilar(pkg.Name, popular) {
					result.Findings = append(result.Findings, TyposquatFinding{
						Package:   pkg.Name,
						Ecosystem: pkg.Ecosystem,
						SimilarTo: popular,
						Reason:    fmt.Sprintf("Name similar to popular package '%s'", popular),
						RiskLevel: "medium",
					})
					result.Summary.SuspiciousCount++
				}
			}
		}

		// Check package age
		if s.config.Typosquats.CheckNewPackages && pkg.Ecosystem == "npm" {
			age := getPackageAge(ctx, client, pkg.Name)
			if age >= 0 && age < 30 {
				result.Findings = append(result.Findings, TyposquatFinding{
					Package:   pkg.Name,
					Ecosystem: pkg.Ecosystem,
					AgeInDays: age,
					Reason:    fmt.Sprintf("Package is only %d days old", age),
					RiskLevel: "low",
				})
				result.Summary.NewPackagesCount++
			}
		}
	}

	return result
}

func isSimilar(a, b string) bool {
	if len(a) < 3 || len(b) < 3 {
		return false
	}
	// Simple Levenshtein-like check
	if abs(len(a)-len(b)) > 2 {
		return false
	}
	differences := 0
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			differences++
		}
	}
	differences += abs(len(a) - len(b))
	return differences > 0 && differences <= 2
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getPackageAge(ctx context.Context, client *http.Client, name string) int {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://registry.npmjs.org/"+name, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return -1
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		Time map[string]string `json:"time"`
	}
	if json.Unmarshal(body, &pkg) != nil {
		return -1
	}

	created, ok := pkg.Time["created"]
	if !ok {
		return -1
	}

	t, err := time.Parse(time.RFC3339, created)
	if err != nil {
		return -1
	}

	return int(time.Since(t).Hours() / 24)
}

// ==================== Deprecations Feature ====================

type deprecationsFeatureResult struct {
	Summary  *DeprecationsSummary
	Findings []DeprecationFinding
}

func (s *PackagesScanner) runDeprecationsFeature(ctx context.Context, components []ComponentData) *deprecationsFeatureResult {
	result := &deprecationsFeatureResult{
		Summary:  &DeprecationsSummary{ByEcosystem: make(map[string]int)},
		Findings: []DeprecationFinding{},
	}

	result.Summary.TotalPackages = len(components)

	client := &http.Client{Timeout: 5 * time.Second}

	for _, pkg := range components {
		var deprecated bool
		var message, alternative string

		switch pkg.Ecosystem {
		case "npm":
			if s.config.Deprecations.CheckNPM {
				deprecated, message = checkNPMDeprecation(ctx, client, pkg.Name, pkg.Version)
			}
		case "pypi":
			if s.config.Deprecations.CheckPyPI {
				// PyPI doesn't have a formal deprecation field, but we check classifiers
				deprecated, message = checkPyPIDeprecation(ctx, client, pkg.Name)
			}
		case "golang":
			if s.config.Deprecations.CheckGo {
				// Go modules can have retract directives
				// This is a placeholder - would need to check go.mod
			}
		}

		if deprecated {
			result.Findings = append(result.Findings, DeprecationFinding{
				Package:     pkg.Name,
				Version:     pkg.Version,
				Ecosystem:   pkg.Ecosystem,
				Message:     message,
				Alternative: alternative,
			})
			result.Summary.DeprecatedCount++
			result.Summary.ByEcosystem[pkg.Ecosystem]++
		}
	}

	return result
}

func checkNPMDeprecation(ctx context.Context, client *http.Client, name, version string) (bool, string) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://registry.npmjs.org/%s/%s", name, version), nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		Deprecated string `json:"deprecated"`
	}
	if json.Unmarshal(body, &pkg) != nil {
		return false, ""
	}

	return pkg.Deprecated != "", pkg.Deprecated
}

func checkPyPIDeprecation(ctx context.Context, client *http.Client, name string) (bool, string) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://pypi.org/pypi/"+name+"/json", nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		Info struct {
			Classifiers []string `json:"classifiers"`
		} `json:"info"`
	}
	if json.Unmarshal(body, &pkg) != nil {
		return false, ""
	}

	for _, c := range pkg.Info.Classifiers {
		if strings.Contains(c, "Inactive") || strings.Contains(c, "Deprecated") {
			return true, c
		}
	}

	return false, ""
}

// ==================== Duplicates Feature ====================

type duplicatesFeatureResult struct {
	Summary  *DuplicatesSummary
	Findings []DuplicateFinding
}

func (s *PackagesScanner) runDuplicatesFeature(components []ComponentData) *duplicatesFeatureResult {
	result := &duplicatesFeatureResult{
		Summary:  &DuplicatesSummary{},
		Findings: []DuplicateFinding{},
	}

	result.Summary.TotalPackages = len(components)

	// Check for multiple versions of same package
	if s.config.Duplicates.CheckVersions {
		packageVersions := make(map[string][]string)
		for _, pkg := range components {
			key := pkg.Name
			packageVersions[key] = append(packageVersions[key], pkg.Version)
		}

		for name, versions := range packageVersions {
			if len(versions) > 1 {
				result.Findings = append(result.Findings, DuplicateFinding{
					Package:   name,
					Versions:  versions,
					IssueType: "version",
					Message:   fmt.Sprintf("Package has %d different versions installed", len(versions)),
				})
				result.Summary.DuplicateVersions++
			}
		}
	}

	// Check for packages with same functionality
	if s.config.Duplicates.CheckFunctionality {
		// Known groups of packages with similar functionality
		functionalGroups := map[string][]string{
			"date":    {"moment", "dayjs", "date-fns", "luxon"},
			"http":    {"axios", "node-fetch", "got", "request", "superagent"},
			"lodash":  {"lodash", "underscore", "ramda"},
			"promise": {"bluebird", "q", "when"},
		}

		foundPackages := make(map[string][]string)
		for _, pkg := range components {
			for group, names := range functionalGroups {
				for _, name := range names {
					if pkg.Name == name {
						foundPackages[group] = append(foundPackages[group], pkg.Name)
					}
				}
			}
		}

		for group, found := range foundPackages {
			if len(found) > 1 {
				result.Findings = append(result.Findings, DuplicateFinding{
					Package:   group,
					Versions:  found,
					IssueType: "functionality",
					Message:   fmt.Sprintf("Multiple packages for %s functionality: %s", group, strings.Join(found, ", ")),
				})
				result.Summary.DuplicateFunctionality++
			}
		}
	}

	return result
}

// ==================== Helpers ====================

func findFiles(root, name string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				n := info.Name()
				if n == "node_modules" || n == ".git" || n == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if info.Name() == name {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func extractEcosystem(purl string) string {
	if len(purl) < 5 || purl[:4] != "pkg:" {
		return "unknown"
	}
	rest := purl[4:]
	for i, c := range rest {
		if c == '/' {
			return rest[:i]
		}
	}
	return "unknown"
}
