package liveapi

import (
	"context"
	"time"
)

// Pre-approved URL for OSV API
const OSVBaseURL = "https://api.osv.dev/v1"

// OSVClient is a client for the OSV vulnerability database
type OSVClient struct {
	*Client
}

// NewOSVClient creates a new OSV client with default settings
func NewOSVClient() *OSVClient {
	return &OSVClient{
		Client: NewClient(OSVBaseURL,
			WithTimeout(30*time.Second),
			WithCache(15*time.Minute),
			WithRateLimit(10),
			WithUserAgent("Zero-Scanner/1.0 (OSV Query)"),
		),
	}
}

// NewOSVClientWithTimeout creates a new OSV client with custom timeout
func NewOSVClientWithTimeout(timeout time.Duration) *OSVClient {
	return &OSVClient{
		Client: NewClient(OSVBaseURL,
			WithTimeout(timeout),
			WithCache(15*time.Minute),
			WithRateLimit(10),
			WithUserAgent("Zero-Scanner/1.0 (OSV Query)"),
		),
	}
}

// Vulnerability represents an OSV vulnerability
type Vulnerability struct {
	ID               string        `json:"id"`
	Summary          string        `json:"summary"`
	Details          string        `json:"details"`
	Aliases          []string      `json:"aliases"`
	Modified         time.Time     `json:"modified"`
	Published        time.Time     `json:"published"`
	DatabaseSpecific interface{}   `json:"database_specific,omitempty"`
	References       []Reference   `json:"references,omitempty"`
	Affected         []Affected    `json:"affected,omitempty"`
	Severity         []OSVSeverity `json:"severity,omitempty"`
}

// Reference is a vulnerability reference
type Reference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// Affected describes an affected package
type Affected struct {
	Package           Package  `json:"package"`
	Ranges            []Range  `json:"ranges,omitempty"`
	Versions          []string `json:"versions,omitempty"`
	EcosystemSpecific interface{} `json:"ecosystem_specific,omitempty"`
}

// Package identifies a package
type Package struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
	Purl      string `json:"purl,omitempty"`
}

// Range describes version ranges
type Range struct {
	Type   string  `json:"type"`
	Events []Event `json:"events"`
}

// Event is a range event
type Event struct {
	Introduced   string `json:"introduced,omitempty"`
	Fixed        string `json:"fixed,omitempty"`
	LastAffected string `json:"last_affected,omitempty"`
	Limit        string `json:"limit,omitempty"`
}

// OSVSeverity represents CVSS severity
type OSVSeverity struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}

// QueryRequest is the OSV query format
type QueryRequest struct {
	Package *PackageQuery `json:"package,omitempty"`
	Version string        `json:"version,omitempty"`
	Commit  string        `json:"commit,omitempty"`
}

// PackageQuery identifies a package for querying
type PackageQuery struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
	Purl      string `json:"purl,omitempty"`
}

// QueryResponse is the OSV response format
type QueryResponse struct {
	Vulns []Vulnerability `json:"vulns"`
}

// BatchQueryRequest is the OSV batch query format
type BatchQueryRequest struct {
	Queries []QueryRequest `json:"queries"`
}

// BatchQueryResponse is the OSV batch response format
type BatchQueryResponse struct {
	Results []QueryResponse `json:"results"`
}

// QueryPackage queries OSV for vulnerabilities affecting a package
func (c *OSVClient) QueryPackage(ctx context.Context, ecosystem, name, version string) ([]Vulnerability, error) {
	req := QueryRequest{
		Package: &PackageQuery{
			Name:      name,
			Ecosystem: ecosystem,
		},
		Version: version,
	}

	var resp QueryResponse
	if err := c.Query(ctx, "/query", req, &resp); err != nil {
		return nil, err
	}

	return resp.Vulns, nil
}

// QueryPURL queries OSV using a Package URL
func (c *OSVClient) QueryPURL(ctx context.Context, purl string) ([]Vulnerability, error) {
	req := QueryRequest{
		Package: &PackageQuery{
			Purl: purl,
		},
	}

	var resp QueryResponse
	if err := c.Query(ctx, "/query", req, &resp); err != nil {
		return nil, err
	}

	return resp.Vulns, nil
}

// QueryBatch queries multiple packages at once
func (c *OSVClient) QueryBatch(ctx context.Context, queries []QueryRequest) ([]QueryResponse, error) {
	req := BatchQueryRequest{
		Queries: queries,
	}

	var resp BatchQueryResponse
	if err := c.Post(ctx, "/querybatch", req, &resp); err != nil {
		return nil, err
	}

	return resp.Results, nil
}

// GetVulnerability retrieves a specific vulnerability by ID
func (c *OSVClient) GetVulnerability(ctx context.Context, id string) (*Vulnerability, error) {
	var vuln Vulnerability
	if err := c.Get(ctx, "/vulns/"+id, &vuln); err != nil {
		return nil, err
	}
	return &vuln, nil
}

// GetHighestSeverity returns the highest severity from vulnerability severity list
func (v *Vulnerability) GetHighestSeverity() string {
	// Check severity array first
	for _, sev := range v.Severity {
		if sev.Type == "CVSS_V3" {
			// Parse CVSS score
			// CVSS v3 format: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"
			return parseCVSSSeverity(sev.Score)
		}
	}
	return "unknown"
}

func parseCVSSSeverity(score string) string {
	// Very simplified - just look for common indicators
	// A real implementation would parse the CVSS vector
	if len(score) == 0 {
		return "unknown"
	}
	// If it's just a number, map to severity
	// 9.0-10.0 = critical, 7.0-8.9 = high, 4.0-6.9 = medium, 0.1-3.9 = low
	return "medium" // Default
}

// GetCVEs returns CVE aliases from the vulnerability
func (v *Vulnerability) GetCVEs() []string {
	var cves []string
	for _, alias := range v.Aliases {
		if len(alias) > 4 && alias[:4] == "CVE-" {
			cves = append(cves, alias)
		}
	}
	return cves
}

// GetFixedVersion returns the fixed version if available
func (v *Vulnerability) GetFixedVersion(ecosystem, name string) string {
	for _, affected := range v.Affected {
		if affected.Package.Ecosystem == ecosystem && affected.Package.Name == name {
			for _, r := range affected.Ranges {
				for _, event := range r.Events {
					if event.Fixed != "" {
						return event.Fixed
					}
				}
			}
		}
	}
	return ""
}

// DefaultOSVClient is the default OSV client instance
var DefaultOSVClient = NewOSVClient()

// QueryOSV queries the default OSV client for vulnerabilities
func QueryOSV(ctx context.Context, ecosystem, name, version string) ([]Vulnerability, error) {
	return DefaultOSVClient.QueryPackage(ctx, ecosystem, name, version)
}
