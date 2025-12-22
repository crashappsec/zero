// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package sheets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// OrgData holds aggregated data for an entire organization
type OrgData struct {
	OrgName     string
	Repos       []RepoData
	TotalRepos  int
	GeneratedAt string

	// Aggregated counts
	TotalVulns       int
	TotalSecrets     int
	TotalLicenses    int
	TotalMLModels    int
	TotalPackages    int
	CriticalVulns    int
	HighVulns        int
	MediumVulns      int
	LowVulns         int
	BusFactorRisk    int // Count of repos with bus factor <= 2
	OrphanedFiles    int
	LicenseViolation int
}

// RepoData holds data for a single repository
type RepoData struct {
	Name  string
	Owner string

	// Package Analysis
	Vulns    []VulnRow
	Licenses []LicenseRow
	Health   HealthSummary

	// Code Security
	Secrets   []SecretRow
	CodeVulns []CodeVulnRow

	// Tech ID
	MLModels    []MLModelRow
	Frameworks  []FrameworkRow
	AIFindings  []AIFindingRow
	TechSummary TechSummary

	// Ownership
	BusFactor       int
	BusFactorRisk   string
	CODEOWNERSCov   float64
	CODEOWNERSIssue int
	Contributors    int

	// DORA
	DORA DORASummary

	// Summary
	Summary RepoSummary
}

// Row types for spreadsheet export

type VulnRow struct {
	Repo         string
	Package      string
	CVE          string
	Severity     string
	CVSS         string
	Ecosystem    string
	FixedVersion string
	KEV          bool
	Reachable    string
	Title        string
}

type SecretRow struct {
	Repo             string
	Type             string
	File             string
	Line             int
	Severity         string
	Detection        string
	Removed          bool
	RotationPriority string
}

type LicenseRow struct {
	Repo      string
	Package   string
	License   string
	Status    string
	Ecosystem string
	Risk      string
}

type MLModelRow struct {
	Repo         string
	Name         string
	Source       string
	Format       string
	License      string
	HasModelCard bool
	SecurityRisk string
}

type FrameworkRow struct {
	Repo         string
	Name         string
	Category     string
	Version      string
	UsagePattern string
}

type AIFindingRow struct {
	Repo        string
	Title       string
	Severity    string
	Category    string
	File        string
	Remediation string
}

type CodeVulnRow struct {
	Repo     string
	RuleID   string
	Title    string
	Severity string
	File     string
	Line     int
	Category string
}

type HealthSummary struct {
	TotalPackages    int
	HealthyCount     int
	WarningCount     int
	CriticalCount    int
	DeprecatedCount  int
	AvgHealthScore   float64
}

type TechSummary struct {
	TotalTechs    int
	TopTechs      []string
	MLModelCount  int
	SecurityCount int
}

type DORASummary struct {
	DeployFreq        string
	DeployFreqClass   string
	LeadTime          string
	LeadTimeClass     string
	ChangeFailureRate string
	ChangeFailClass   string
	MTTR              string
	MTTRClass         string
	OverallClass      string
}

type RepoSummary struct {
	CriticalVulns int
	HighVulns     int
	MediumVulns   int
	LowVulns      int
	TotalVulns    int
	SecretsCount  int
	LicenseIssues int
}

// Transformer handles data loading and transformation
type Transformer struct {
	zeroHome string
}

// NewTransformer creates a new transformer
func NewTransformer(zeroHome string) *Transformer {
	return &Transformer{zeroHome: zeroHome}
}

// LoadOrgData loads and aggregates data for all repos in an organization
func (t *Transformer) LoadOrgData(orgName string) (*OrgData, error) {
	orgDir := filepath.Join(t.zeroHome, "repos", orgName)

	entries, err := os.ReadDir(orgDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read org directory: %w", err)
	}

	orgData := &OrgData{
		OrgName: orgName,
		Repos:   make([]RepoData, 0),
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoName := entry.Name()
		repoData, err := t.LoadRepoData(orgName, repoName)
		if err != nil {
			continue // Skip repos with errors
		}

		orgData.Repos = append(orgData.Repos, *repoData)
		t.aggregateOrgData(orgData, repoData)
	}

	orgData.TotalRepos = len(orgData.Repos)

	// Sort repos by critical vulns (most critical first)
	sort.Slice(orgData.Repos, func(i, j int) bool {
		return orgData.Repos[i].Summary.CriticalVulns > orgData.Repos[j].Summary.CriticalVulns
	})

	return orgData, nil
}

// LoadRepoData loads data for a single repository
func (t *Transformer) LoadRepoData(owner, repo string) (*RepoData, error) {
	analysisDir := filepath.Join(t.zeroHome, "repos", owner, repo, "analysis")

	repoData := &RepoData{
		Name:  repo,
		Owner: owner,
	}

	// Load package analysis
	t.loadPackageAnalysis(analysisDir, repoData)

	// Load code security
	t.loadCodeSecurity(analysisDir, repoData)

	// Load tech-id
	t.loadTechID(analysisDir, repoData)

	// Load ownership
	t.loadOwnership(analysisDir, repoData)

	// Load devops/DORA
	t.loadDevOps(analysisDir, repoData)

	// Calculate summary
	t.calculateSummary(repoData)

	return repoData, nil
}

func (t *Transformer) aggregateOrgData(org *OrgData, repo *RepoData) {
	org.TotalVulns += repo.Summary.TotalVulns
	org.CriticalVulns += repo.Summary.CriticalVulns
	org.HighVulns += repo.Summary.HighVulns
	org.MediumVulns += repo.Summary.MediumVulns
	org.LowVulns += repo.Summary.LowVulns
	org.TotalSecrets += repo.Summary.SecretsCount
	org.TotalLicenses += repo.Summary.LicenseIssues
	org.TotalMLModels += len(repo.MLModels)

	if repo.BusFactor <= 2 {
		org.BusFactorRisk++
	}
}

func (t *Transformer) loadPackageAnalysis(analysisDir string, repo *RepoData) {
	data, err := os.ReadFile(filepath.Join(analysisDir, "package-analysis.json"))
	if err != nil {
		return
	}

	var result struct {
		Findings struct {
			Vulns struct {
				Findings []struct {
					ID           string  `json:"id"`
					Package      string  `json:"package"`
					Version      string  `json:"version"`
					Ecosystem    string  `json:"ecosystem"`
					Severity     string  `json:"severity"`
					Title        string  `json:"title"`
					FixedVersion string  `json:"fixed_version"`
					KEV          bool    `json:"kev"`
					CVSS         float64 `json:"cvss_score"`
				} `json:"findings"`
			} `json:"vulns"`
			Licenses struct {
				Findings []struct {
					Package   string   `json:"package"`
					Licenses  []string `json:"licenses"`
					Status    string   `json:"status"`
					Ecosystem string   `json:"ecosystem"`
				} `json:"findings"`
			} `json:"licenses"`
			Health struct {
				Summary struct {
					TotalPackages int     `json:"total_packages"`
					Healthy       int     `json:"healthy"`
					Warning       int     `json:"warning"`
					Critical      int     `json:"critical"`
					Deprecated    int     `json:"deprecated"`
					AvgScore      float64 `json:"avg_health_score"`
				} `json:"summary"`
			} `json:"health"`
			Reachability struct {
				Findings []struct {
					ID        string `json:"id"`
					Reachable bool   `json:"reachable"`
				} `json:"findings"`
			} `json:"reachability"`
		} `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	// Build reachability map
	reachMap := make(map[string]string)
	for _, r := range result.Findings.Reachability.Findings {
		if r.Reachable {
			reachMap[r.ID] = "Yes"
		} else {
			reachMap[r.ID] = "No"
		}
	}

	// Transform vulns
	repoName := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	for _, v := range result.Findings.Vulns.Findings {
		reachable := reachMap[v.ID]
		if reachable == "" {
			reachable = "Unknown"
		}

		repo.Vulns = append(repo.Vulns, VulnRow{
			Repo:         repoName,
			Package:      v.Package,
			CVE:          v.ID,
			Severity:     v.Severity,
			CVSS:         fmt.Sprintf("%.1f", v.CVSS),
			Ecosystem:    v.Ecosystem,
			FixedVersion: v.FixedVersion,
			KEV:          v.KEV,
			Reachable:    reachable,
			Title:        v.Title,
		})
	}

	// Transform licenses
	for _, l := range result.Findings.Licenses.Findings {
		status := l.Status
		risk := "Low"
		if status == "denied" {
			risk = "High"
		} else if status == "needs_review" {
			risk = "Medium"
		}

		repo.Licenses = append(repo.Licenses, LicenseRow{
			Repo:      repoName,
			Package:   l.Package,
			License:   strings.Join(l.Licenses, ", "),
			Status:    status,
			Ecosystem: l.Ecosystem,
			Risk:      risk,
		})
	}

	// Health summary
	repo.Health = HealthSummary{
		TotalPackages:   result.Findings.Health.Summary.TotalPackages,
		HealthyCount:    result.Findings.Health.Summary.Healthy,
		WarningCount:    result.Findings.Health.Summary.Warning,
		CriticalCount:   result.Findings.Health.Summary.Critical,
		DeprecatedCount: result.Findings.Health.Summary.Deprecated,
		AvgHealthScore:  result.Findings.Health.Summary.AvgScore,
	}
}

func (t *Transformer) loadCodeSecurity(analysisDir string, repo *RepoData) {
	data, err := os.ReadFile(filepath.Join(analysisDir, "code-security.json"))
	if err != nil {
		return
	}

	var result struct {
		Findings struct {
			Secrets struct {
				Findings []struct {
					Type             string `json:"type"`
					File             string `json:"file"`
					Line             int    `json:"line"`
					Severity         string `json:"severity"`
					DetectionSource  string `json:"detection_source"`
					IsRemoved        bool   `json:"is_removed"`
					RotationPriority string `json:"rotation_priority"`
				} `json:"findings"`
			} `json:"secrets"`
			Vulns struct {
				Findings []struct {
					RuleID   string `json:"rule_id"`
					Title    string `json:"title"`
					Severity string `json:"severity"`
					File     string `json:"file"`
					Line     int    `json:"line"`
					Category string `json:"category"`
				} `json:"findings"`
			} `json:"vulns"`
		} `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	repoName := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)

	// Transform secrets
	for _, s := range result.Findings.Secrets.Findings {
		detection := s.DetectionSource
		if detection == "" {
			detection = "semgrep"
		}

		priority := s.RotationPriority
		if priority == "" {
			if s.Severity == "critical" {
				priority = "Immediate"
			} else if s.Severity == "high" {
				priority = "High"
			} else {
				priority = "Medium"
			}
		}

		repo.Secrets = append(repo.Secrets, SecretRow{
			Repo:             repoName,
			Type:             s.Type,
			File:             s.File,
			Line:             s.Line,
			Severity:         s.Severity,
			Detection:        detection,
			Removed:          s.IsRemoved,
			RotationPriority: priority,
		})
	}

	// Transform code vulns
	for _, v := range result.Findings.Vulns.Findings {
		repo.CodeVulns = append(repo.CodeVulns, CodeVulnRow{
			Repo:     repoName,
			RuleID:   v.RuleID,
			Title:    v.Title,
			Severity: v.Severity,
			File:     v.File,
			Line:     v.Line,
			Category: v.Category,
		})
	}
}

func (t *Transformer) loadTechID(analysisDir string, repo *RepoData) {
	data, err := os.ReadFile(filepath.Join(analysisDir, "technology.json"))
	if err != nil {
		return
	}

	var result struct {
		Findings struct {
			Models struct {
				Findings []struct {
					Name         string `json:"name"`
					Source       string `json:"source"`
					Format       string `json:"format"`
					License      string `json:"license"`
					HasModelCard bool   `json:"has_model_card"`
					SecurityRisk string `json:"security_risk"`
				} `json:"findings"`
			} `json:"models"`
			Frameworks struct {
				Findings []struct {
					Name         string `json:"name"`
					Category     string `json:"category"`
					Version      string `json:"version"`
					UsagePattern string `json:"usage_pattern"`
				} `json:"findings"`
			} `json:"frameworks"`
			Security struct {
				Findings []struct {
					Title       string `json:"title"`
					Severity    string `json:"severity"`
					Category    string `json:"category"`
					File        string `json:"file"`
					Remediation string `json:"remediation"`
				} `json:"findings"`
			} `json:"security"`
			Detection struct {
				Summary struct {
					TotalTechs int      `json:"total_technologies"`
					TopTechs   []string `json:"top_technologies"`
				} `json:"summary"`
			} `json:"detection"`
		} `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	repoName := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)

	// Transform models
	for _, m := range result.Findings.Models.Findings {
		risk := m.SecurityRisk
		if risk == "" {
			risk = "Unknown"
		}

		repo.MLModels = append(repo.MLModels, MLModelRow{
			Repo:         repoName,
			Name:         m.Name,
			Source:       m.Source,
			Format:       m.Format,
			License:      m.License,
			HasModelCard: m.HasModelCard,
			SecurityRisk: risk,
		})
	}

	// Transform frameworks
	for _, f := range result.Findings.Frameworks.Findings {
		repo.Frameworks = append(repo.Frameworks, FrameworkRow{
			Repo:         repoName,
			Name:         f.Name,
			Category:     f.Category,
			Version:      f.Version,
			UsagePattern: f.UsagePattern,
		})
	}

	// Transform AI security findings
	for _, s := range result.Findings.Security.Findings {
		repo.AIFindings = append(repo.AIFindings, AIFindingRow{
			Repo:        repoName,
			Title:       s.Title,
			Severity:    s.Severity,
			Category:    s.Category,
			File:        s.File,
			Remediation: s.Remediation,
		})
	}

	// Tech summary
	repo.TechSummary = TechSummary{
		TotalTechs:    result.Findings.Detection.Summary.TotalTechs,
		TopTechs:      result.Findings.Detection.Summary.TopTechs,
		MLModelCount:  len(repo.MLModels),
		SecurityCount: len(repo.AIFindings),
	}
}

func (t *Transformer) loadOwnership(analysisDir string, repo *RepoData) {
	data, err := os.ReadFile(filepath.Join(analysisDir, "code-ownership.json"))
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			BusFactor      int     `json:"bus_factor"`
			BusFactorRisk  string  `json:"bus_factor_risk"`
			Contributors   int     `json:"total_contributors"`
			CODEOWNERSCov  float64 `json:"ownership_coverage"`
			CODEOWNERSErr  int     `json:"codeowners_issues"`
		} `json:"summary"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	repo.BusFactor = result.Summary.BusFactor
	repo.BusFactorRisk = result.Summary.BusFactorRisk
	repo.Contributors = result.Summary.Contributors
	repo.CODEOWNERSCov = result.Summary.CODEOWNERSCov * 100 // Convert to percentage
	repo.CODEOWNERSIssue = result.Summary.CODEOWNERSErr
}

func (t *Transformer) loadDevOps(analysisDir string, repo *RepoData) {
	data, err := os.ReadFile(filepath.Join(analysisDir, "devops.json"))
	if err != nil {
		return
	}

	var result struct {
		Findings struct {
			DORA struct {
				Summary struct {
					DeployFreq      float64 `json:"deployment_frequency"`
					DeployFreqClass string  `json:"deployment_frequency_class"`
					LeadTime        float64 `json:"lead_time_hours"`
					LeadTimeClass   string  `json:"lead_time_class"`
					ChangeFailure   float64 `json:"change_failure_rate"`
					ChangeFailClass string  `json:"change_failure_class"`
					MTTR            float64 `json:"mttr_hours"`
					MTTRClass       string  `json:"mttr_class"`
					OverallClass    string  `json:"overall_class"`
				} `json:"summary"`
			} `json:"dora"`
		} `json:"findings"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	dora := result.Findings.DORA.Summary
	repo.DORA = DORASummary{
		DeployFreq:        fmt.Sprintf("%.2f/day", dora.DeployFreq),
		DeployFreqClass:   dora.DeployFreqClass,
		LeadTime:          fmt.Sprintf("%.1f hrs", dora.LeadTime),
		LeadTimeClass:     dora.LeadTimeClass,
		ChangeFailureRate: fmt.Sprintf("%.1f%%", dora.ChangeFailure*100),
		ChangeFailClass:   dora.ChangeFailClass,
		MTTR:              fmt.Sprintf("%.1f hrs", dora.MTTR),
		MTTRClass:         dora.MTTRClass,
		OverallClass:      dora.OverallClass,
	}
}

func (t *Transformer) calculateSummary(repo *RepoData) {
	for _, v := range repo.Vulns {
		repo.Summary.TotalVulns++
		switch strings.ToLower(v.Severity) {
		case "critical":
			repo.Summary.CriticalVulns++
		case "high":
			repo.Summary.HighVulns++
		case "medium":
			repo.Summary.MediumVulns++
		case "low":
			repo.Summary.LowVulns++
		}
	}

	repo.Summary.SecretsCount = len(repo.Secrets)

	for _, l := range repo.Licenses {
		if l.Status == "denied" || l.Status == "needs_review" {
			repo.Summary.LicenseIssues++
		}
	}
}

// ToVulnRows converts org data to vulnerability rows
func (o *OrgData) ToVulnRows() [][]interface{} {
	rows := [][]interface{}{
		{"Repo", "Package", "CVE", "Severity", "CVSS", "Ecosystem", "Fixed Version", "KEV", "Reachable", "Title"},
	}

	for _, repo := range o.Repos {
		for _, v := range repo.Vulns {
			kevStr := ""
			if v.KEV {
				kevStr = "Yes"
			}
			rows = append(rows, []interface{}{
				v.Repo, v.Package, v.CVE, v.Severity, v.CVSS,
				v.Ecosystem, v.FixedVersion, kevStr, v.Reachable, v.Title,
			})
		}
	}

	return rows
}

// ToSecretRows converts org data to secret rows
func (o *OrgData) ToSecretRows() [][]interface{} {
	rows := [][]interface{}{
		{"Repo", "Type", "File", "Line", "Severity", "Detection", "Removed", "Rotation Priority"},
	}

	for _, repo := range o.Repos {
		for _, s := range repo.Secrets {
			removedStr := ""
			if s.Removed {
				removedStr = "Yes"
			}
			rows = append(rows, []interface{}{
				s.Repo, s.Type, s.File, s.Line, s.Severity,
				s.Detection, removedStr, s.RotationPriority,
			})
		}
	}

	return rows
}

// ToLicenseRows converts org data to license rows
func (o *OrgData) ToLicenseRows() [][]interface{} {
	rows := [][]interface{}{
		{"Repo", "Package", "License", "Status", "Ecosystem", "Risk"},
	}

	for _, repo := range o.Repos {
		for _, l := range repo.Licenses {
			// Only include issues (denied or needs review)
			if l.Status == "denied" || l.Status == "needs_review" {
				rows = append(rows, []interface{}{
					l.Repo, l.Package, l.License, l.Status, l.Ecosystem, l.Risk,
				})
			}
		}
	}

	return rows
}

// ToMLModelRows converts org data to ML model rows
func (o *OrgData) ToMLModelRows() [][]interface{} {
	rows := [][]interface{}{
		{"Repo", "Model Name", "Source", "Format", "License", "Has Model Card", "Security Risk"},
	}

	for _, repo := range o.Repos {
		for _, m := range repo.MLModels {
			hasCardStr := "No"
			if m.HasModelCard {
				hasCardStr = "Yes"
			}
			rows = append(rows, []interface{}{
				m.Repo, m.Name, m.Source, m.Format, m.License, hasCardStr, m.SecurityRisk,
			})
		}
	}

	return rows
}

// ToFrameworkRows converts org data to framework rows
func (o *OrgData) ToFrameworkRows() [][]interface{} {
	rows := [][]interface{}{
		{"Repo", "Framework", "Category", "Version", "Usage Pattern"},
	}

	for _, repo := range o.Repos {
		for _, f := range repo.Frameworks {
			rows = append(rows, []interface{}{
				f.Repo, f.Name, f.Category, f.Version, f.UsagePattern,
			})
		}
	}

	return rows
}

// ToAIFindingRows converts org data to AI security finding rows
func (o *OrgData) ToAIFindingRows() [][]interface{} {
	rows := [][]interface{}{
		{"Repo", "Finding", "Severity", "Category", "File", "Remediation"},
	}

	for _, repo := range o.Repos {
		for _, f := range repo.AIFindings {
			rows = append(rows, []interface{}{
				f.Repo, f.Title, f.Severity, f.Category, f.File, f.Remediation,
			})
		}
	}

	return rows
}

// ToCODEOWNERSRows converts org data to CODEOWNERS coverage rows
func (o *OrgData) ToCODEOWNERSRows() [][]interface{} {
	rows := [][]interface{}{
		{"Repo", "Coverage %", "Issues", "Bus Factor", "Risk Level"},
	}

	for _, repo := range o.Repos {
		rows = append(rows, []interface{}{
			fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
			fmt.Sprintf("%.1f%%", repo.CODEOWNERSCov),
			repo.CODEOWNERSIssue,
			repo.BusFactor,
			repo.BusFactorRisk,
		})
	}

	return rows
}
