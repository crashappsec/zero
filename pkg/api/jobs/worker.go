package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/api/ws"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/workflow/hydrate"
)

// Worker processes jobs from the queue
type Worker struct {
	id       int
	queue    *Queue
	hub      *ws.Hub
	wg       *sync.WaitGroup
}

// WorkerPool manages a pool of workers
type WorkerPool struct {
	queue   *Queue
	hub     *ws.Hub
	workers []*Worker
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(queue *Queue, hub *ws.Hub, numWorkers int) *WorkerPool {
	if numWorkers <= 0 {
		numWorkers = 1
	}

	return &WorkerPool{
		queue:   queue,
		hub:     hub,
		workers: make([]*Worker, numWorkers),
	}
}

// Start begins processing jobs
func (p *WorkerPool) Start(ctx context.Context) {
	ctx, p.cancel = context.WithCancel(ctx)

	for i := 0; i < len(p.workers); i++ {
		p.workers[i] = &Worker{
			id:    i,
			queue: p.queue,
			hub:   p.hub,
			wg:    &p.wg,
		}
		p.wg.Add(1)
		go p.workers[i].run(ctx)
	}

	log.Printf("Started %d scan workers", len(p.workers))
}

// Stop gracefully stops all workers
func (p *WorkerPool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	log.Println("All scan workers stopped")
}

// run is the main worker loop
func (w *Worker) run(ctx context.Context) {
	defer w.wg.Done()

	for {
		job, err := w.queue.Dequeue(ctx)
		if err != nil {
			// Context canceled, exit
			return
		}

		w.executeJob(ctx, job)
	}
}

// executeJob runs a single scan job
func (w *Worker) executeJob(ctx context.Context, job *Job) {
	log.Printf("[Worker %d] Starting job %s: %s", w.id, job.ID, job.Target)
	startTime := time.Now()

	// Broadcast job started
	w.broadcast(job.ID, JobStatusMessage{
		Type:   "job_status",
		JobID:  job.ID,
		Status: JobStatusCloning,
	})
	job.SetStatus(JobStatusCloning)
	job.Progress.SetPhase("cloning")

	// Create hydrate options
	opts := &hydrate.Options{
		Profile:  job.Profile,
		Force:    job.Force,
		SkipSlow: job.SkipSlow,
		Depth:    job.Depth,
	}

	if job.IsOrg {
		opts.Org = job.Target
	} else {
		opts.Repo = job.Target
	}

	// Create hydrate instance
	h, err := hydrate.New(opts)
	if err != nil {
		job.SetError(fmt.Errorf("failed to create hydrate: %w", err))
		w.broadcastError(job, err)
		return
	}

	// Hook into scanner progress for real-time updates
	w.setupProgressHook(job, h)

	// Update status to scanning
	job.SetStatus(JobStatusScanning)
	job.Progress.SetPhase("scanning")
	w.broadcast(job.ID, JobStatusMessage{
		Type:   "job_status",
		JobID:  job.ID,
		Status: JobStatusScanning,
	})

	// Run the hydrate workflow
	projectIDs, err := h.Run(ctx)
	if err != nil {
		// Check if it was canceled
		if ctx.Err() != nil {
			job.SetStatus(JobStatusCanceled)
			w.broadcast(job.ID, JobStatusMessage{
				Type:   "job_status",
				JobID:  job.ID,
				Status: JobStatusCanceled,
			})
			return
		}

		job.SetError(err)
		w.broadcastError(job, err)
		return
	}

	// Success!
	job.mu.Lock()
	job.ProjectIDs = projectIDs
	job.mu.Unlock()
	job.SetStatus(JobStatusComplete)

	duration := time.Since(startTime).Seconds()
	log.Printf("[Worker %d] Completed job %s in %.1fs: %d projects", w.id, job.ID, duration, len(projectIDs))

	w.broadcast(job.ID, ScanCompleteMessage{
		Type:       "scan_complete",
		JobID:      job.ID,
		Success:    true,
		ProjectIDs: projectIDs,
		Duration:   duration,
	})
}

// setupProgressHook hooks into the hydrate runner for real-time progress
func (w *Worker) setupProgressHook(job *Job, h *hydrate.Hydrate) {
	// Note: We need to access the runner from hydrate
	// This requires adding a method to hydrate.Hydrate to expose the runner
	// or accepting a callback directly

	// For now, we'll set up a basic progress tracker that polls
	// In a full implementation, we'd modify hydrate.Hydrate to accept
	// an OnProgress callback that we can hook into here

	// The hydrate package's runner.OnProgress is set internally
	// We would need to either:
	// 1. Add a method to Hydrate to set a custom OnProgress callback
	// 2. Or wrap the runner with our own progress tracking

	// For Phase 2, we'll use a simplified approach where we track
	// status changes via the job's progress field
}

// broadcast sends a message to all clients watching this job
func (w *Worker) broadcast(jobID string, msg interface{}) {
	if w.hub == nil {
		return
	}
	if err := w.hub.BroadcastToJob(jobID, msg); err != nil {
		log.Printf("[Worker %d] Failed to broadcast: %v", w.id, err)
	}
}

// broadcastError sends an error message
func (w *Worker) broadcastError(job *Job, err error) {
	w.broadcast(job.ID, JobStatusMessage{
		Type:   "job_status",
		JobID:  job.ID,
		Status: JobStatusFailed,
		Error:  err.Error(),
	})
}

// broadcastScannerProgress sends scanner progress update
func (w *Worker) broadcastScannerProgress(jobID string, name string, status scanner.Status, summary string, duration float64) {
	w.broadcast(jobID, ScannerProgressMessage{
		Type:     "scanner_progress",
		JobID:    jobID,
		Scanner:  name,
		Status:   status,
		Summary:  summary,
		Duration: duration,
	})
}
