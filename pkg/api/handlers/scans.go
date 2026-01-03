package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/crashappsec/zero/pkg/api/jobs"
)

// ScanHandler handles scan-related API requests
type ScanHandler struct {
	queue *jobs.Queue
}

// NewScanHandler creates a new scan handler
func NewScanHandler(queue *jobs.Queue) *ScanHandler {
	return &ScanHandler{
		queue: queue,
	}
}

// StartRequest represents a request to start a new scan
type StartRequest struct {
	Target   string `json:"target"`            // owner/repo or org name
	IsOrg    bool   `json:"is_org,omitempty"`  // true if scanning an org
	Profile  string `json:"profile,omitempty"` // scan profile (default: standard)
	Force    bool   `json:"force,omitempty"`   // re-scan even if exists
	SkipSlow bool   `json:"skip_slow,omitempty"`
	Depth    int    `json:"depth,omitempty"`   // git clone depth
}

// StartResponse is returned when a scan is started
type StartResponse struct {
	JobID      string    `json:"job_id"`
	Target     string    `json:"target"`
	Profile    string    `json:"profile"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	WSEndpoint string    `json:"ws_endpoint"`
}

// Start initiates a new scan job
func (h *ScanHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Validate target
	if req.Target == "" {
		writeError(w, http.StatusBadRequest, "target is required", nil)
		return
	}

	// Determine if it's an org or repo
	if !req.IsOrg && !strings.Contains(req.Target, "/") {
		// No slash, assume it's an org
		req.IsOrg = true
	}

	// Default profile
	if req.Profile == "" {
		req.Profile = "standard"
	}

	// Generate job ID
	jobID := jobs.GenerateJobID()

	// Create job
	job := jobs.NewJob(jobID, req.Target, req.IsOrg, req.Profile)
	job.Force = req.Force
	job.SkipSlow = req.SkipSlow
	job.Depth = req.Depth

	// Enqueue job
	if err := h.queue.Enqueue(job); err != nil {
		writeError(w, http.StatusServiceUnavailable, "failed to enqueue job", err)
		return
	}

	// Return response
	resp := StartResponse{
		JobID:      jobID,
		Target:     req.Target,
		Profile:    req.Profile,
		Status:     string(jobs.JobStatusQueued),
		CreatedAt:  job.StartedAt,
		WSEndpoint: "/ws/scan/" + jobID,
	}

	writeJSON(w, http.StatusAccepted, resp)
}

// Get returns the status of a scan job
func (h *ScanHandler) Get(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")

	job, ok := h.queue.Get(jobID)
	if !ok {
		writeError(w, http.StatusNotFound, "job not found", nil)
		return
	}

	// Build response
	resp := map[string]interface{}{
		"job_id":     job.ID,
		"target":     job.Target,
		"is_org":     job.IsOrg,
		"profile":    job.Profile,
		"status":     job.GetStatus(),
		"started_at": job.StartedAt,
	}

	if job.FinishedAt != nil {
		resp["finished_at"] = job.FinishedAt
		resp["duration_seconds"] = job.FinishedAt.Sub(job.StartedAt).Seconds()
	}

	if job.Error != "" {
		resp["error"] = job.Error
	}

	if job.Progress != nil {
		resp["progress"] = job.Progress.Clone()
	}

	if len(job.ProjectIDs) > 0 {
		resp["project_ids"] = job.ProjectIDs
	}

	writeJSON(w, http.StatusOK, resp)
}

// Cancel cancels a running scan job
func (h *ScanHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")

	if err := h.queue.Cancel(jobID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "job not found", err)
		} else {
			writeError(w, http.StatusBadRequest, "cannot cancel job", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListActive returns all active (running/queued) jobs
func (h *ScanHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	activeJobs := h.queue.ListActive()

	// Build response
	var resp []map[string]interface{}
	for _, job := range activeJobs {
		item := map[string]interface{}{
			"job_id":     job.ID,
			"target":     job.Target,
			"profile":    job.Profile,
			"status":     job.GetStatus(),
			"started_at": job.StartedAt,
		}
		if job.Progress != nil {
			item["progress"] = job.Progress.Clone()
		}
		resp = append(resp, item)
	}

	if resp == nil {
		resp = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  resp,
		"total": len(resp),
	})
}

// ListHistory returns recently completed jobs
func (h *ScanHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	// Get jobs completed in the last 24 hours
	recentJobs := h.queue.ListRecent(24 * time.Hour)

	var resp []map[string]interface{}
	for _, job := range recentJobs {
		item := map[string]interface{}{
			"job_id":      job.ID,
			"target":      job.Target,
			"profile":     job.Profile,
			"status":      job.GetStatus(),
			"started_at":  job.StartedAt,
			"finished_at": job.FinishedAt,
		}
		if job.FinishedAt != nil {
			item["duration_seconds"] = job.FinishedAt.Sub(job.StartedAt).Seconds()
		}
		if job.Error != "" {
			item["error"] = job.Error
		}
		if len(job.ProjectIDs) > 0 {
			item["project_ids"] = job.ProjectIDs
		}
		resp = append(resp, item)
	}

	if resp == nil {
		resp = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  resp,
		"total": len(resp),
	})
}

// Stats returns queue statistics
func (h *ScanHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := h.queue.Stats()
	writeJSON(w, http.StatusOK, stats)
}
