// Package jobs provides background job queue functionality for async operations
package jobs

import (
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/scanner"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusCloning   JobStatus = "cloning"
	JobStatusScanning  JobStatus = "scanning"
	JobStatusComplete  JobStatus = "complete"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCanceled  JobStatus = "canceled"
)

// Job represents a background scan job
type Job struct {
	ID          string        `json:"id"`
	Target      string        `json:"target"`      // owner/repo or org name
	IsOrg       bool          `json:"is_org"`      // true if scanning an org
	Profile     string        `json:"profile"`     // scan profile
	Status      JobStatus     `json:"status"`
	Progress    *JobProgress  `json:"progress,omitempty"`
	StartedAt   time.Time     `json:"started_at"`
	FinishedAt  *time.Time    `json:"finished_at,omitempty"`
	Error       string        `json:"error,omitempty"`
	ProjectIDs  []string      `json:"project_ids,omitempty"` // resulting project IDs

	// Options
	Force       bool          `json:"force"`
	SkipSlow    bool          `json:"skip_slow"`
	Depth       int           `json:"depth"`

	mu          sync.RWMutex
}

// JobProgress tracks detailed progress of a scan job
type JobProgress struct {
	Phase            string                  `json:"phase"`             // cloning, scanning
	ReposTotal       int                     `json:"repos_total"`
	ReposComplete    int                     `json:"repos_complete"`
	CurrentRepo      string                  `json:"current_repo,omitempty"`
	ScannersTotal    int                     `json:"scanners_total"`
	ScannersComplete int                     `json:"scanners_complete"`
	CurrentScanner   string                  `json:"current_scanner,omitempty"`
	ScannerStatuses  map[string]ScannerState `json:"scanner_statuses,omitempty"`

	mu               sync.RWMutex
}

// ScannerState represents the state of a single scanner
type ScannerState struct {
	Status   scanner.Status `json:"status"`
	Summary  string         `json:"summary,omitempty"`
	Duration float64        `json:"duration_seconds,omitempty"`
}

// NewJob creates a new scan job
func NewJob(id, target string, isOrg bool, profile string) *Job {
	return &Job{
		ID:        id,
		Target:    target,
		IsOrg:     isOrg,
		Profile:   profile,
		Status:    JobStatusQueued,
		StartedAt: time.Now(),
		Progress:  NewJobProgress(),
	}
}

// NewJobProgress creates a new progress tracker
func NewJobProgress() *JobProgress {
	return &JobProgress{
		ScannerStatuses: make(map[string]ScannerState),
	}
}

// SetStatus updates the job status (thread-safe)
func (j *Job) SetStatus(status JobStatus) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = status
	if status == JobStatusComplete || status == JobStatusFailed || status == JobStatusCanceled {
		now := time.Now()
		j.FinishedAt = &now
	}
}

// SetError sets the error message and marks job as failed
func (j *Job) SetError(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Error = err.Error()
	j.Status = JobStatusFailed
	now := time.Now()
	j.FinishedAt = &now
}

// GetStatus returns the current status (thread-safe)
func (j *Job) GetStatus() JobStatus {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Status
}

// SetPhase updates the progress phase
func (p *JobProgress) SetPhase(phase string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Phase = phase
}

// SetRepoProgress updates repo progress
func (p *JobProgress) SetRepoProgress(current string, complete, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CurrentRepo = current
	p.ReposComplete = complete
	p.ReposTotal = total
}

// SetScannerProgress updates scanner progress
func (p *JobProgress) SetScannerProgress(current string, complete, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CurrentScanner = current
	p.ScannersComplete = complete
	p.ScannersTotal = total
}

// UpdateScanner updates a single scanner's state
func (p *JobProgress) UpdateScanner(name string, status scanner.Status, summary string, duration float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ScannerStatuses[name] = ScannerState{
		Status:   status,
		Summary:  summary,
		Duration: duration,
	}

	// Count completed scanners
	complete := 0
	for _, s := range p.ScannerStatuses {
		if s.Status == scanner.StatusComplete || s.Status == scanner.StatusFailed || s.Status == scanner.StatusSkipped {
			complete++
		}
	}
	p.ScannersComplete = complete
}

// Clone returns a snapshot of the progress (thread-safe)
func (p *JobProgress) Clone() *JobProgress {
	p.mu.RLock()
	defer p.mu.RUnlock()

	statuses := make(map[string]ScannerState, len(p.ScannerStatuses))
	for k, v := range p.ScannerStatuses {
		statuses[k] = v
	}

	return &JobProgress{
		Phase:            p.Phase,
		ReposTotal:       p.ReposTotal,
		ReposComplete:    p.ReposComplete,
		CurrentRepo:      p.CurrentRepo,
		ScannersTotal:    p.ScannersTotal,
		ScannersComplete: p.ScannersComplete,
		CurrentScanner:   p.CurrentScanner,
		ScannerStatuses:  statuses,
	}
}

// WebSocket message types for broadcasting progress

// JobStatusMessage is sent when job status changes
type JobStatusMessage struct {
	Type   string    `json:"type"`
	JobID  string    `json:"job_id"`
	Status JobStatus `json:"status"`
	Error  string    `json:"error,omitempty"`
}

// CloneProgressMessage is sent during cloning
type CloneProgressMessage struct {
	Type      string `json:"type"`
	JobID     string `json:"job_id"`
	Repo      string `json:"repo"`
	Status    string `json:"status"` // cloning, complete, failed
	FileCount int    `json:"file_count,omitempty"`
	Size      string `json:"size,omitempty"`
}

// ScannerProgressMessage is sent during scanning
type ScannerProgressMessage struct {
	Type     string         `json:"type"`
	JobID    string         `json:"job_id"`
	Scanner  string         `json:"scanner"`
	Status   scanner.Status `json:"status"`
	Summary  string         `json:"summary,omitempty"`
	Duration float64        `json:"duration_seconds,omitempty"`
}

// ScanCompleteMessage is sent when scan finishes
type ScanCompleteMessage struct {
	Type       string   `json:"type"`
	JobID      string   `json:"job_id"`
	Success    bool     `json:"success"`
	ProjectIDs []string `json:"project_ids"`
	Duration   float64  `json:"duration_seconds"`
	Error      string   `json:"error,omitempty"`
}
