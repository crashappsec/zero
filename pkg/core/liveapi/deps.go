package liveapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// Pre-approved URL for deps.dev API
const DepsDevBaseURL = "https://api.deps.dev/v3alpha"

// DepsDevClient is a client for the deps.dev package intelligence API
type DepsDevClient struct {
	*Client
}

// NewDepsDevClient creates a new deps.dev client with default settings
func NewDepsDevClient() *DepsDevClient {
	return &DepsDevClient{
		Client: NewClient(DepsDevBaseURL,
			WithTimeout(30*time.Second),
			WithCache(24*time.Hour), // Scorecard updates weekly, cache longer than OSV
			WithRateLimit(10),
			WithUserAgent("Zero-Scanner/1.0 (deps.dev Query)"),
		),
	}
}

// NewDepsDevClientWithTimeout creates a new deps.dev client with custom timeout
func NewDepsDevClientWithTimeout(timeout time.Duration) *DepsDevClient {
	return &DepsDevClient{
		Client: NewClient(DepsDevBaseURL,
			WithTimeout(timeout),
			WithCache(24*time.Hour),
			WithRateLimit(10),
			WithUserAgent("Zero-Scanner/1.0 (deps.dev Query)"),
		),
	}
}

// VersionDetails represents version-specific package information
type VersionDetails struct {
	VersionKey      VersionKey        `json:"versionKey"`
	PublishedAt     time.Time         `json:"publishedAt"`
	IsDefault       bool              `json:"isDefault"`
	IsDeprecated    bool              `json:"isDeprecated"`
	Licenses        []string          `json:"licenses"`
	AdvisoryKeys    []AdvisoryKey     `json:"advisoryKeys"`
	Links           []Link            `json:"links"`
	SlsaProvenances []SLSAProvenance  `json:"slsaProvenances"`
	Projects        []ProjectInfo     `json:"projects"`
}

// VersionKey identifies a specific package version
type VersionKey struct {
	System  string `json:"system"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// AdvisoryKey identifies a security advisory
type AdvisoryKey struct {
	ID string `json:"id"`
}

// Link represents a package link
type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// SLSAProvenance represents SLSA provenance attestation
type SLSAProvenance struct {
	SourceRepository string `json:"sourceRepository"`
	Commit           string `json:"commit"`
	URL              string `json:"url"`
	Verified         bool   `json:"verified"`
}

// ProjectInfo represents source project information
type ProjectInfo struct {
	ProjectKey ProjectKey `json:"projectKey"`
	Scorecard  *Scorecard `json:"scorecard,omitempty"`
}

// ProjectKey identifies a source project
type ProjectKey struct {
	ID string `json:"id"`
}

// Scorecard represents OpenSSF Scorecard data
type Scorecard struct {
	Date         string           `json:"date"`
	OverallScore float64          `json:"overallScore"`
	Checks       []ScorecardCheck `json:"checks"`
}

// ScorecardCheck represents an individual scorecard check
type ScorecardCheck struct {
	Name   string   `json:"name"`
	Score  int      `json:"score"`
	Reason string   `json:"reason"`
	Details []string `json:"details,omitempty"`
}

// PackageInfo represents package-level information (all versions)
type PackageInfo struct {
	PackageKey     PackageKey     `json:"packageKey"`
	Versions       []VersionInfo  `json:"versions"`
}

// PackageKey identifies a package
type PackageKey struct {
	System string `json:"system"`
	Name   string `json:"name"`
}

// VersionInfo represents basic version information
type VersionInfo struct {
	VersionKey   VersionKey `json:"versionKey"`
	PublishedAt  time.Time  `json:"publishedAt"`
	IsDefault    bool       `json:"isDefault"`
	IsDeprecated bool       `json:"isDeprecated"`
}

// HealthScore represents a package's health score
type HealthScore struct {
	Score          float64          `json:"score"`
	Checks         []ScorecardCheck `json:"checks"`
	IsDeprecated   bool             `json:"is_deprecated"`
	HasProvenance  bool             `json:"has_provenance"`
	ProvenanceInfo *SLSAProvenance  `json:"provenance_info,omitempty"`
}

// Ecosystem mapping from SBOM ecosystem to deps.dev system
var ecosystemToSystem = map[string]string{
	"npm":       "NPM",
	"pypi":      "PYPI",
	"golang":    "GO",
	"go":        "GO",
	"maven":     "MAVEN",
	"cargo":     "CARGO",
	"nuget":     "NUGET",
	"rubygems":  "RUBYGEMS",
	"packagist": "PACKAGIST",
}

// NormalizeEcosystem converts SBOM ecosystem names to deps.dev system names
func NormalizeEcosystem(ecosystem string) string {
	if system, ok := ecosystemToSystem[ecosystem]; ok {
		return system
	}
	// Return uppercase version as fallback
	return ecosystem
}

// CachedGet performs a cached GET request
func (c *DepsDevClient) CachedGet(ctx context.Context, path string, result any) error {
	// Check cache first
	cacheKey := c.cacheKey(path, nil)
	if c.Cache != nil {
		if cached, ok := c.Cache.Get(cacheKey); ok {
			return json.Unmarshal(cached, result)
		}
	}

	// Wait for rate limiter
	if c.RateLimiter != nil {
		if err := c.RateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit: %w", err)
		}
	}

	// Make request
	respBody, err := c.doRequestRaw(ctx, "GET", path, nil)
	if err != nil {
		return err
	}

	// Cache the response
	if c.Cache != nil {
		c.Cache.Set(cacheKey, respBody)
	}

	return json.Unmarshal(respBody, result)
}

// GetVersionDetails retrieves detailed information about a specific package version
func (c *DepsDevClient) GetVersionDetails(ctx context.Context, ecosystem, name, version string) (*VersionDetails, error) {
	system := NormalizeEcosystem(ecosystem)
	path := fmt.Sprintf("/systems/%s/packages/%s/versions/%s",
		url.PathEscape(system),
		url.PathEscape(name),
		url.PathEscape(version))

	var details VersionDetails
	if err := c.CachedGet(ctx, path, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

// GetVersionDetailsByPURL retrieves version details using a Package URL
func (c *DepsDevClient) GetVersionDetailsByPURL(ctx context.Context, purl string) (*VersionDetails, error) {
	path := fmt.Sprintf("/purl/%s", url.PathEscape(purl))

	var details VersionDetails
	if err := c.CachedGet(ctx, path, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

// GetPackageVersions retrieves all versions of a package
func (c *DepsDevClient) GetPackageVersions(ctx context.Context, ecosystem, name string) (*PackageInfo, error) {
	system := NormalizeEcosystem(ecosystem)
	path := fmt.Sprintf("/systems/%s/packages/%s",
		url.PathEscape(system),
		url.PathEscape(name))

	var info PackageInfo
	if err := c.CachedGet(ctx, path, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// IsDeprecated checks if a specific package version is deprecated
func (c *DepsDevClient) IsDeprecated(ctx context.Context, ecosystem, name, version string) (bool, error) {
	details, err := c.GetVersionDetails(ctx, ecosystem, name, version)
	if err != nil {
		return false, err
	}
	return details.IsDeprecated, nil
}

// GetHealthScore retrieves the OpenSSF Scorecard score for a package
func (c *DepsDevClient) GetHealthScore(ctx context.Context, ecosystem, name, version string) (*HealthScore, error) {
	details, err := c.GetVersionDetails(ctx, ecosystem, name, version)
	if err != nil {
		return nil, err
	}

	health := &HealthScore{
		IsDeprecated:  details.IsDeprecated,
		HasProvenance: len(details.SlsaProvenances) > 0,
	}

	// Extract provenance info if available
	if len(details.SlsaProvenances) > 0 {
		health.ProvenanceInfo = &details.SlsaProvenances[0]
	}

	// Extract scorecard from projects
	for _, proj := range details.Projects {
		if proj.Scorecard != nil {
			health.Score = proj.Scorecard.OverallScore
			health.Checks = proj.Scorecard.Checks
			break
		}
	}

	return health, nil
}

// GetSLSAProvenance retrieves SLSA provenance information for a package version
func (c *DepsDevClient) GetSLSAProvenance(ctx context.Context, ecosystem, name, version string) ([]SLSAProvenance, error) {
	details, err := c.GetVersionDetails(ctx, ecosystem, name, version)
	if err != nil {
		return nil, err
	}
	return details.SlsaProvenances, nil
}

// GetLatestVersion returns the default (latest) version of a package
func (c *DepsDevClient) GetLatestVersion(ctx context.Context, ecosystem, name string) (string, error) {
	info, err := c.GetPackageVersions(ctx, ecosystem, name)
	if err != nil {
		return "", err
	}

	for _, v := range info.Versions {
		if v.IsDefault {
			return v.VersionKey.Version, nil
		}
	}

	// If no default, return the last version
	if len(info.Versions) > 0 {
		return info.Versions[len(info.Versions)-1].VersionKey.Version, nil
	}

	return "", fmt.Errorf("no versions found for %s/%s", ecosystem, name)
}

// IsOutdated checks if a package version is outdated compared to the latest
func (c *DepsDevClient) IsOutdated(ctx context.Context, ecosystem, name, version string) (bool, string, error) {
	latest, err := c.GetLatestVersion(ctx, ecosystem, name)
	if err != nil {
		return false, "", err
	}

	return version != latest, latest, nil
}

// DefaultDepsDevClient is the default deps.dev client instance
var DefaultDepsDevClient = NewDepsDevClient()

// QueryDepsDevHealth queries the default deps.dev client for package health
func QueryDepsDevHealth(ctx context.Context, ecosystem, name, version string) (*HealthScore, error) {
	return DefaultDepsDevClient.GetHealthScore(ctx, ecosystem, name, version)
}
