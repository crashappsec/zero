// Package storage provides database abstraction for Zero's data layer.
// This enables fast queries via SQLite while keeping JSON files as source of truth.
package storage

import (
	"context"
	"time"
)

// Store defines the storage interface for Zero's data layer.
// Implementations include SQLite (local) and optionally Supabase (cloud).
type Store interface {
	// Projects
	ListProjects(ctx context.Context, opts ListOptions) ([]*Project, error)
	GetProject(ctx context.Context, id string) (*Project, error)
	UpsertProject(ctx context.Context, project *Project) error
	DeleteProject(ctx context.Context, id string) error

	// Scans
	CreateScan(ctx context.Context, scan *Scan) error
	UpdateScan(ctx context.Context, scan *Scan) error
	GetScan(ctx context.Context, id string) (*Scan, error)
	ListScans(ctx context.Context, projectID string, opts ListOptions) ([]*Scan, error)
	GetLatestScan(ctx context.Context, projectID string) (*Scan, error)

	// Findings Summary (aggregated stats per project)
	UpsertFindingsSummary(ctx context.Context, summary *FindingsSummary) error
	GetFindingsSummary(ctx context.Context, projectID string) (*FindingsSummary, error)

	// Aggregations (fast indexed queries)
	GetAggregateStats(ctx context.Context) (*AggregateStats, error)

	// Vulnerabilities (cross-project queries)
	UpsertVulnerabilities(ctx context.Context, projectID string, vulns []*Vulnerability) error
	GetVulnerabilities(ctx context.Context, opts VulnOptions) ([]*Vulnerability, int, error)
	DeleteVulnerabilities(ctx context.Context, projectID string) error

	// Secrets (cross-project queries)
	UpsertSecrets(ctx context.Context, projectID string, secrets []*Secret) error
	GetSecrets(ctx context.Context, opts SecretOptions) ([]*Secret, int, error)
	DeleteSecrets(ctx context.Context, projectID string) error

	// Sync from JSON files
	SyncProjectFromJSON(ctx context.Context, projectID string, analysisDir string) error

	// Lifecycle
	Ping(ctx context.Context) error
	Close() error
	Migrate(ctx context.Context) error
}

// ListOptions provides pagination and filtering for list queries.
type ListOptions struct {
	Owner    string
	Limit    int
	Offset   int
	SortBy   string
	SortDesc bool
}

// VulnOptions provides filtering for vulnerability queries.
type VulnOptions struct {
	ProjectID  string
	Severities []string // critical, high, medium, low
	Package    string
	Limit      int
	Offset     int
}

// SecretOptions provides filtering for secret queries.
type SecretOptions struct {
	ProjectID  string
	Severities []string
	Type       string
	Limit      int
	Offset     int
}

// Project represents a hydrated repository.
type Project struct {
	ID             string    `json:"id"`              // "owner/repo"
	Owner          string    `json:"owner"`           // GitHub owner
	Name           string    `json:"name"`            // Repository name
	RepoPath       string    `json:"repo_path"`       // Path to cloned repo
	AnalysisPath   string    `json:"analysis_path"`   // Path to analysis JSON files
	FileCount      int       `json:"file_count"`      // Number of files in repo
	DiskSize       int64     `json:"disk_size"`       // Total disk usage in bytes
	LastScan       time.Time `json:"last_scan"`       // Last scan timestamp
	FreshnessLevel string    `json:"freshness_level"` // fresh, stale, very-stale, expired
	FreshnessAge   int       `json:"freshness_age"`   // Age in hours
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Scan represents a scan execution record.
type Scan struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	Profile         string    `json:"profile"`
	Status          string    `json:"status"` // queued, cloning, scanning, complete, failed
	CommitSHA       string    `json:"commit_sha"`
	StartedAt       time.Time `json:"started_at"`
	FinishedAt      time.Time `json:"finished_at,omitempty"`
	DurationSeconds int       `json:"duration_seconds"`
	Error           string    `json:"error,omitempty"`
}

// FindingsSummary contains aggregated finding counts for a project.
type FindingsSummary struct {
	ProjectID         string    `json:"project_id"`
	VulnsCritical     int       `json:"vulns_critical"`
	VulnsHigh         int       `json:"vulns_high"`
	VulnsMedium       int       `json:"vulns_medium"`
	VulnsLow          int       `json:"vulns_low"`
	VulnsTotal        int       `json:"vulns_total"`
	SecretsTotal      int       `json:"secrets_total"`
	PackagesTotal     int       `json:"packages_total"`
	TechnologiesTotal int       `json:"technologies_total"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// AggregateStats contains global statistics across all projects.
type AggregateStats struct {
	TotalProjects     int                        `json:"total_projects"`
	TotalVulns        int                        `json:"total_vulns"`
	VulnsBySeverity   map[string]int             `json:"vulns_by_severity"`
	TotalSecrets      int                        `json:"total_secrets"`
	TotalPackages     int                        `json:"total_packages"`
	TotalTechnologies int                        `json:"total_technologies"`
	ProjectStats      []*FindingsSummary         `json:"project_stats,omitempty"`
	FreshnessCounts   map[string]int             `json:"freshness_counts"` // fresh, stale, etc.
}

// Vulnerability represents a detected vulnerability.
type Vulnerability struct {
	ID          int64  `json:"id,omitempty"`
	ProjectID   string `json:"project_id"`
	VulnID      string `json:"vuln_id"` // CVE/GHSA ID
	Package     string `json:"package"`
	Version     string `json:"version"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	FixVersion  string `json:"fix_version,omitempty"`
	Source      string `json:"source"` // package, code
	Scanner     string `json:"scanner"`
}

// Secret represents a detected secret.
type Secret struct {
	ID           int64  `json:"id,omitempty"`
	ProjectID    string `json:"project_id"`
	File         string `json:"file"`
	Line         int    `json:"line"`
	Type         string `json:"type"`
	Severity     string `json:"severity"`
	Description  string `json:"description"`
	RedactedMatch string `json:"redacted_match"`
}
