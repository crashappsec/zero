// Package types provides API type definitions for Zero
package types

import (
	"time"

	"github.com/crashappsec/zero/pkg/workflow/freshness"
)

// Project represents a hydrated project in API responses
type Project struct {
	ID             string          `json:"id"`
	Owner          string          `json:"owner"`
	Name           string          `json:"name"`
	RepoPath       string          `json:"repo_path"`
	AnalysisPath   string          `json:"analysis_path"`
	LastScan       time.Time       `json:"last_scan,omitempty"`
	FileCount      int             `json:"file_count"`
	DiskSize       int64           `json:"disk_size"`
	Freshness      *FreshnessInfo  `json:"freshness,omitempty"`
	AvailableScans []string        `json:"available_scans"`
	Summary        *ProjectSummary `json:"summary,omitempty"`
}

// FreshnessInfo holds freshness status
type FreshnessInfo struct {
	Level        freshness.Level `json:"level"`
	LevelString  string          `json:"level_string"`
	AgeString    string          `json:"age_string"`
	NeedsRefresh bool            `json:"needs_refresh"`
}

// ProjectSummary contains key metrics for a project
type ProjectSummary struct {
	Vulnerabilities *SeverityCounts `json:"vulnerabilities,omitempty"`
	Secrets         int             `json:"secrets"`
	Packages        int             `json:"packages"`
	Technologies    int             `json:"technologies"`
}

// SeverityCounts holds counts by severity level
type SeverityCounts struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// ScanRequest represents a request to start a new scan
type ScanRequest struct {
	Target  string       `json:"target"`
	Profile string       `json:"profile,omitempty"`
	Options *ScanOptions `json:"options,omitempty"`
}

// ScanOptions configures scan behavior
type ScanOptions struct {
	Force    bool `json:"force"`
	SkipSlow bool `json:"skip_slow"`
	Depth    int  `json:"depth"`
}

// ScanJob represents an active or completed scan job
type ScanJob struct {
	ID         string       `json:"id"`
	Target     string       `json:"target"`
	Profile    string       `json:"profile"`
	Status     JobStatus    `json:"status"`
	Progress   *JobProgress `json:"progress,omitempty"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt *time.Time   `json:"finished_at,omitempty"`
	Error      string       `json:"error,omitempty"`
	WSEndpoint string       `json:"ws_endpoint,omitempty"`
}

// JobStatus represents the status of a scan job
type JobStatus string

const (
	JobStatusQueued   JobStatus = "queued"
	JobStatusCloning  JobStatus = "cloning"
	JobStatusScanning JobStatus = "scanning"
	JobStatusComplete JobStatus = "complete"
	JobStatusFailed   JobStatus = "failed"
	JobStatusCanceled JobStatus = "canceled"
)

// JobProgress tracks scan progress
type JobProgress struct {
	Phase           string            `json:"phase"`
	ScannersTotal   int               `json:"scanners_total"`
	ScannersComplete int              `json:"scanners_complete"`
	CurrentScanner  string            `json:"current_scanner,omitempty"`
	ScannerStatuses map[string]string `json:"scanner_statuses,omitempty"`
}

// ProfileInfo describes a scan profile
type ProfileInfo struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	EstimatedTime string   `json:"estimated_time,omitempty"`
	Scanners      []string `json:"scanners"`
}

// ScannerInfo describes an available scanner
type ScannerInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Features    []string `json:"features,omitempty"`
}

// AgentInfo describes an available agent
type AgentInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Persona     string `json:"persona"`
	Description string `json:"description"`
	Scanner     string `json:"primary_scanner"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ListResponse wraps paginated list responses
type ListResponse[T any] struct {
	Data   []T `json:"data"`
	Total  int `json:"total"`
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}
