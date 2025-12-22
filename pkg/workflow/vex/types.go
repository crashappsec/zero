// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

// Package vex provides VEX (Vulnerability Exploitability eXchange) document generation.
// VEX is a companion to SBOM that communicates whether vulnerabilities actually affect a product.
// This implementation follows the CycloneDX VEX specification.
package vex

import (
	"time"
)

// Document represents a complete VEX document
type Document struct {
	BOMFormat    string       `json:"bomFormat"`
	SpecVersion  string       `json:"specVersion"`
	SerialNumber string       `json:"serialNumber"`
	Version      int          `json:"version"`
	Metadata     Metadata     `json:"metadata"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

// Metadata contains document metadata
type Metadata struct {
	Timestamp string     `json:"timestamp"`
	Tools     []Tool     `json:"tools,omitempty"`
	Component *Component `json:"component,omitempty"`
	Supplier  *Supplier  `json:"supplier,omitempty"`
}

// Tool identifies the tool that generated the VEX
type Tool struct {
	Vendor  string `json:"vendor"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Supplier identifies the organization supplying the VEX
type Supplier struct {
	Name string   `json:"name"`
	URL  []string `json:"url,omitempty"`
}

// Component identifies the product/component the VEX applies to
type Component struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	PURL    string `json:"purl,omitempty"`
	BOMRef  string `json:"bom-ref,omitempty"`
}

// Vulnerability represents a single vulnerability with VEX analysis
type Vulnerability struct {
	ID          string       `json:"id"`
	Source      Source       `json:"source"`
	References  []Reference  `json:"references,omitempty"`
	Ratings     []Rating     `json:"ratings,omitempty"`
	CWEs        []int        `json:"cwes,omitempty"`
	Description string       `json:"description,omitempty"`
	Detail      string       `json:"detail,omitempty"`
	Advisories  []Advisory   `json:"advisories,omitempty"`
	Published   string       `json:"published,omitempty"`
	Updated     string       `json:"updated,omitempty"`
	Affects     []Affect     `json:"affects"`
	Analysis    *Analysis    `json:"analysis,omitempty"`
}

// Source identifies where the vulnerability was reported
type Source struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Reference provides additional vulnerability references
type Reference struct {
	ID     string `json:"id"`
	Source Source `json:"source"`
}

// Rating provides vulnerability severity rating
type Rating struct {
	Source   *Source `json:"source,omitempty"`
	Score    float64 `json:"score,omitempty"`
	Severity string  `json:"severity,omitempty"`
	Method   string  `json:"method,omitempty"`
	Vector   string  `json:"vector,omitempty"`
}

// Advisory references security advisories
type Advisory struct {
	Title string `json:"title,omitempty"`
	URL   string `json:"url"`
}

// Affect identifies which component is affected
type Affect struct {
	Ref      string           `json:"ref"` // bom-ref or purl
	Versions []AffectedVersion `json:"versions,omitempty"`
}

// AffectedVersion specifies affected version ranges
type AffectedVersion struct {
	Version string `json:"version,omitempty"`
	Range   string `json:"range,omitempty"`
	Status  string `json:"status"` // affected, unaffected, unknown
}

// Analysis contains the VEX exploitability analysis
type Analysis struct {
	State         State         `json:"state"`
	Justification Justification `json:"justification,omitempty"`
	Response      []Response    `json:"response,omitempty"`
	Detail        string        `json:"detail,omitempty"`
	FirstIssued   string        `json:"firstIssued,omitempty"`
	LastUpdated   string        `json:"lastUpdated,omitempty"`
}

// State represents the VEX analysis state
type State string

const (
	StateInTriage    State = "in_triage"
	StateExploitable State = "exploitable"
	StateResolved    State = "resolved"
	StateNotAffected State = "not_affected"
	StateFalsePositive State = "false_positive"
)

// Justification explains why a vulnerability doesn't affect a product
type Justification string

const (
	JustificationCodeNotPresent           Justification = "code_not_present"
	JustificationCodeNotReachable         Justification = "code_not_reachable"
	JustificationRequiresConfiguration    Justification = "requires_configuration"
	JustificationRequiresDependency       Justification = "requires_dependency"
	JustificationRequiresEnvironment      Justification = "requires_environment"
	JustificationProtectedByCompiler      Justification = "protected_by_compiler"
	JustificationProtectedAtRuntime       Justification = "protected_at_runtime"
	JustificationProtectedAtPerimeter     Justification = "protected_at_perimeter"
	JustificationProtectedByMitigatingControl Justification = "protected_by_mitigating_control"
)

// Response indicates what action is being taken
type Response string

const (
	ResponseCanNotFix   Response = "can_not_fix"
	ResponseWillNotFix  Response = "will_not_fix"
	ResponseUpdate      Response = "update"
	ResponseRollback    Response = "rollback"
	ResponseWorkaround  Response = "workaround_available"
)

// GeneratorConfig configures VEX document generation
type GeneratorConfig struct {
	// Auto-analyze vulnerabilities using available data
	AutoAnalyze bool `json:"auto_analyze"`

	// Use reachability data to determine code_not_reachable
	UseReachability bool `json:"use_reachability"`

	// Include all vulnerabilities (even those marked not_affected)
	IncludeAll bool `json:"include_all"`

	// Default state for vulnerabilities without analysis
	DefaultState State `json:"default_state"`

	// Supplier information
	SupplierName string `json:"supplier_name"`
	SupplierURL  string `json:"supplier_url"`

	// Product information (optional, uses SBOM metadata if not set)
	ProductName    string `json:"product_name"`
	ProductVersion string `json:"product_version"`
}

// DefaultConfig returns sensible defaults for VEX generation
func DefaultConfig() GeneratorConfig {
	return GeneratorConfig{
		AutoAnalyze:     true,
		UseReachability: true,
		IncludeAll:      false,
		DefaultState:    StateInTriage,
		SupplierName:    "Zero Security Scanner",
		SupplierURL:     "https://github.com/crashappsec/zero",
	}
}

// VulnInput represents vulnerability data from package-analysis scanner
type VulnInput struct {
	ID          string   `json:"id"`
	Package     string   `json:"package"`
	Version     string   `json:"version"`
	Ecosystem   string   `json:"ecosystem"`
	PURL        string   `json:"purl"`
	Severity    string   `json:"severity"`
	CVSS        float64  `json:"cvss"`
	CVSSVector  string   `json:"cvss_vector"`
	CWEs        []int    `json:"cwes"`
	Description string   `json:"description"`
	Published   string   `json:"published"`
	Fixed       string   `json:"fixed"`
	References  []string `json:"references"`
	Aliases     []string `json:"aliases"`

	// Reachability info (if available)
	IsReachable     *bool    `json:"is_reachable,omitempty"`
	ReachablePaths  []string `json:"reachable_paths,omitempty"`
	UsedFunctions   []string `json:"used_functions,omitempty"`
}

// Timestamp returns current time in RFC3339 format
func Timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
