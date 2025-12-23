// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package vex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	BOMFormat        = "CycloneDX"
	SpecVersion      = "1.5"
	GeneratorVersion = "1.0.0"
)

// Generator creates VEX documents from vulnerability data
type Generator struct {
	config GeneratorConfig
}

// NewGenerator creates a new VEX generator
func NewGenerator(config GeneratorConfig) *Generator {
	return &Generator{config: config}
}

// Generate creates a VEX document from vulnerability inputs
func (g *Generator) Generate(vulns []VulnInput, productName, productVersion string) (*Document, error) {
	doc := &Document{
		BOMFormat:    BOMFormat,
		SpecVersion:  SpecVersion,
		SerialNumber: fmt.Sprintf("urn:uuid:%s", uuid.New().String()),
		Version:      1,
		Metadata: Metadata{
			Timestamp: Timestamp(),
			Tools: []Tool{
				{
					Vendor:  "Crash Override",
					Name:    "Zero",
					Version: GeneratorVersion,
				},
			},
		},
		Vulnerabilities: []Vulnerability{},
	}

	// Set product info if provided
	if productName != "" {
		doc.Metadata.Component = &Component{
			Type:    "application",
			Name:    productName,
			Version: productVersion,
		}
	}

	// Set supplier info
	if g.config.SupplierName != "" {
		doc.Metadata.Supplier = &Supplier{
			Name: g.config.SupplierName,
		}
		if g.config.SupplierURL != "" {
			doc.Metadata.Supplier.URL = []string{g.config.SupplierURL}
		}
	}

	// Process each vulnerability
	for _, v := range vulns {
		vuln := g.processVulnerability(v)

		// Skip not_affected if not including all
		if !g.config.IncludeAll && vuln.Analysis != nil && vuln.Analysis.State == StateNotAffected {
			continue
		}

		doc.Vulnerabilities = append(doc.Vulnerabilities, vuln)
	}

	return doc, nil
}

// processVulnerability converts a VulnInput to a VEX Vulnerability
func (g *Generator) processVulnerability(v VulnInput) Vulnerability {
	vuln := Vulnerability{
		ID: v.ID,
		Source: Source{
			Name: g.getSourceName(v.ID),
			URL:  g.getSourceURL(v.ID),
		},
		Description: v.Description,
		Published:   v.Published,
		CWEs:        v.CWEs,
		Affects: []Affect{
			{
				Ref: v.PURL,
				Versions: []AffectedVersion{
					{
						Version: v.Version,
						Status:  "affected",
					},
				},
			},
		},
	}

	// Add ratings
	if v.CVSS > 0 || v.Severity != "" {
		rating := Rating{
			Severity: strings.ToLower(v.Severity),
		}
		if v.CVSS > 0 {
			rating.Score = v.CVSS
			rating.Method = "CVSSv31"
		}
		if v.CVSSVector != "" {
			rating.Vector = v.CVSSVector
		}
		vuln.Ratings = []Rating{rating}
	}

	// Add references
	for _, ref := range v.References {
		vuln.Advisories = append(vuln.Advisories, Advisory{URL: ref})
	}

	// Add aliases as references
	for _, alias := range v.Aliases {
		if alias != v.ID {
			vuln.References = append(vuln.References, Reference{
				ID: alias,
				Source: Source{
					Name: g.getSourceName(alias),
				},
			})
		}
	}

	// Auto-analyze if enabled
	if g.config.AutoAnalyze {
		vuln.Analysis = g.analyzeVulnerability(v)
	} else {
		// Use default state
		vuln.Analysis = &Analysis{
			State:       g.config.DefaultState,
			FirstIssued: Timestamp(),
			LastUpdated: Timestamp(),
		}
	}

	return vuln
}

// analyzeVulnerability performs automatic VEX analysis
func (g *Generator) analyzeVulnerability(v VulnInput) *Analysis {
	analysis := &Analysis{
		State:       g.config.DefaultState,
		FirstIssued: Timestamp(),
		LastUpdated: Timestamp(),
	}

	// Check reachability if available and enabled
	if g.config.UseReachability && v.IsReachable != nil {
		if !*v.IsReachable {
			analysis.State = StateNotAffected
			analysis.Justification = JustificationCodeNotReachable
			analysis.Detail = "Vulnerable code path is not reachable from application entry points"
			if len(v.UsedFunctions) > 0 {
				analysis.Detail = fmt.Sprintf("Package is imported but vulnerable functions (%s) are not called",
					strings.Join(v.UsedFunctions, ", "))
			}
			return analysis
		}
		// Reachable = affected
		analysis.State = StateExploitable
		if len(v.ReachablePaths) > 0 {
			analysis.Detail = fmt.Sprintf("Vulnerable code is reachable via: %s",
				strings.Join(v.ReachablePaths[:min(3, len(v.ReachablePaths))], " -> "))
		}
	}

	// Check if fix is available
	if v.Fixed != "" {
		analysis.Response = []Response{ResponseUpdate}
		fixDetail := fmt.Sprintf("Update to version %s to fix this vulnerability", v.Fixed)
		if analysis.Detail != "" {
			analysis.Detail = analysis.Detail + ". " + fixDetail
		} else {
			analysis.Detail = fixDetail
		}
	}

	return analysis
}

// getSourceName returns the vulnerability source name from ID
func (g *Generator) getSourceName(id string) string {
	switch {
	case strings.HasPrefix(id, "CVE-"):
		return "NVD"
	case strings.HasPrefix(id, "GHSA-"):
		return "GitHub Advisory"
	case strings.HasPrefix(id, "OSV-"):
		return "OSV"
	case strings.HasPrefix(id, "PYSEC-"):
		return "PyPI Advisory"
	case strings.HasPrefix(id, "RUSTSEC-"):
		return "RustSec Advisory"
	case strings.HasPrefix(id, "GO-"):
		return "Go Vulnerability Database"
	default:
		return "Unknown"
	}
}

// getSourceURL returns the vulnerability source URL from ID
func (g *Generator) getSourceURL(id string) string {
	switch {
	case strings.HasPrefix(id, "CVE-"):
		return fmt.Sprintf("https://nvd.nist.gov/vuln/detail/%s", id)
	case strings.HasPrefix(id, "GHSA-"):
		return fmt.Sprintf("https://github.com/advisories/%s", id)
	case strings.HasPrefix(id, "OSV-"):
		return fmt.Sprintf("https://osv.dev/vulnerability/%s", id)
	case strings.HasPrefix(id, "PYSEC-"):
		return fmt.Sprintf("https://osv.dev/vulnerability/%s", id)
	case strings.HasPrefix(id, "RUSTSEC-"):
		return fmt.Sprintf("https://rustsec.org/advisories/%s", id)
	case strings.HasPrefix(id, "GO-"):
		return fmt.Sprintf("https://pkg.go.dev/vuln/%s", id)
	default:
		return ""
	}
}

// GenerateFromScanResults generates VEX from Zero scan results
func (g *Generator) GenerateFromScanResults(analysisDir string) (*Document, error) {
	// Load packages.json for vulnerability data
	packagesPath := filepath.Join(analysisDir, "packages.json")
	packagesData, err := os.ReadFile(packagesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read packages.json: %w", err)
	}

	var packagesResult struct {
		Summary struct {
			Vulns struct {
				TotalVulnerabilities int `json:"total_vulnerabilities"`
			} `json:"vulns"`
		} `json:"summary"`
		Findings struct {
			Vulns []struct {
				ID          string   `json:"id"`
				Package     string   `json:"package"`
				Version     string   `json:"version"`
				Ecosystem   string   `json:"ecosystem"`
				Severity    string   `json:"severity"`
				CVSS        float64  `json:"cvss"`
				CVSSVector  string   `json:"cvss_vector"`
				Description string   `json:"description"`
				Published   string   `json:"published"`
				Fixed       string   `json:"fixed_version"`
				References  []string `json:"references"`
				Aliases     []string `json:"aliases"`
				CWEs        []int    `json:"cwes"`
			} `json:"vulns"`
		} `json:"findings"`
	}

	if err := json.Unmarshal(packagesData, &packagesResult); err != nil {
		return nil, fmt.Errorf("failed to parse packages.json: %w", err)
	}

	// Convert to VulnInputs
	var vulns []VulnInput
	for _, v := range packagesResult.Findings.Vulns {
		vuln := VulnInput{
			ID:          v.ID,
			Package:     v.Package,
			Version:     v.Version,
			Ecosystem:   v.Ecosystem,
			PURL:        buildPURL(v.Ecosystem, v.Package, v.Version),
			Severity:    v.Severity,
			CVSS:        v.CVSS,
			CVSSVector:  v.CVSSVector,
			CWEs:        v.CWEs,
			Description: v.Description,
			Published:   v.Published,
			Fixed:       v.Fixed,
			References:  v.References,
			Aliases:     v.Aliases,
		}
		vulns = append(vulns, vuln)
	}

	// Try to load reachability data if available
	vulns = g.enrichWithReachability(analysisDir, vulns)

	// Get product info from SBOM if available
	productName, productVersion := g.getProductInfo(analysisDir)

	return g.Generate(vulns, productName, productVersion)
}

// enrichWithReachability adds reachability data to vulnerabilities
func (g *Generator) enrichWithReachability(analysisDir string, vulns []VulnInput) []VulnInput {
	if !g.config.UseReachability {
		return vulns
	}

	// Try to load reachability data from packages.json
	packagesPath := filepath.Join(analysisDir, "packages.json")
	packagesData, err := os.ReadFile(packagesPath)
	if err != nil {
		return vulns
	}

	var reachData struct {
		Findings struct {
			Reachability []struct {
				Package       string   `json:"package"`
				Version       string   `json:"version"`
				IsReachable   bool     `json:"is_reachable"`
				CallPaths     []string `json:"call_paths"`
				UsedFunctions []string `json:"used_functions"`
			} `json:"reachability"`
		} `json:"findings"`
	}

	if err := json.Unmarshal(packagesData, &reachData); err != nil {
		return vulns
	}

	// Build reachability map
	reachMap := make(map[string]struct {
		IsReachable   bool
		CallPaths     []string
		UsedFunctions []string
	})

	for _, r := range reachData.Findings.Reachability {
		key := fmt.Sprintf("%s@%s", r.Package, r.Version)
		reachMap[key] = struct {
			IsReachable   bool
			CallPaths     []string
			UsedFunctions []string
		}{
			IsReachable:   r.IsReachable,
			CallPaths:     r.CallPaths,
			UsedFunctions: r.UsedFunctions,
		}
	}

	// Enrich vulnerabilities
	for i := range vulns {
		key := fmt.Sprintf("%s@%s", vulns[i].Package, vulns[i].Version)
		if reach, ok := reachMap[key]; ok {
			vulns[i].IsReachable = &reach.IsReachable
			vulns[i].ReachablePaths = reach.CallPaths
			vulns[i].UsedFunctions = reach.UsedFunctions
		}
	}

	return vulns
}

// getProductInfo extracts product info from SBOM
func (g *Generator) getProductInfo(analysisDir string) (string, string) {
	if g.config.ProductName != "" {
		return g.config.ProductName, g.config.ProductVersion
	}

	sbomPath := filepath.Join(analysisDir, "sbom.cdx.json")
	sbomData, err := os.ReadFile(sbomPath)
	if err != nil {
		return "", ""
	}

	var sbom struct {
		Metadata struct {
			Component struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"component"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(sbomData, &sbom); err != nil {
		return "", ""
	}

	return sbom.Metadata.Component.Name, sbom.Metadata.Component.Version
}

// buildPURL constructs a Package URL from ecosystem, name, and version
func buildPURL(ecosystem, name, version string) string {
	ecosystemMap := map[string]string{
		"npm":       "npm",
		"pypi":      "pypi",
		"pip":       "pypi",
		"go":        "golang",
		"golang":    "golang",
		"maven":     "maven",
		"cargo":     "cargo",
		"rubygems":  "gem",
		"gem":       "gem",
		"nuget":     "nuget",
		"packagist": "composer",
	}

	purlType := ecosystem
	if mapped, ok := ecosystemMap[strings.ToLower(ecosystem)]; ok {
		purlType = mapped
	}

	return fmt.Sprintf("pkg:%s/%s@%s", purlType, name, version)
}

// WriteJSON writes the VEX document to a JSON file
func (d *Document) WriteJSON(path string) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal VEX document: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write VEX file: %w", err)
	}

	return nil
}

// Summary returns a summary of the VEX document
func (d *Document) Summary() map[string]int {
	summary := map[string]int{
		"total":          len(d.Vulnerabilities),
		"in_triage":      0,
		"exploitable":    0,
		"resolved":       0,
		"not_affected":   0,
		"false_positive": 0,
	}

	for _, v := range d.Vulnerabilities {
		if v.Analysis != nil {
			summary[string(v.Analysis.State)]++
		} else {
			summary["in_triage"]++
		}
	}

	return summary
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
