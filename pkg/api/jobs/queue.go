package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Queue manages background scan jobs
type Queue struct {
	jobs     map[string]*Job
	pending  chan *Job
	mu       sync.RWMutex
	maxSize  int
}

// NewQueue creates a new job queue
func NewQueue(maxSize int) *Queue {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &Queue{
		jobs:    make(map[string]*Job),
		pending: make(chan *Job, maxSize),
		maxSize: maxSize,
	}
}

// Enqueue adds a job to the queue
func (q *Queue) Enqueue(job *Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if job already exists
	if _, exists := q.jobs[job.ID]; exists {
		return fmt.Errorf("job %s already exists", job.ID)
	}

	// Check queue capacity
	if len(q.pending) >= q.maxSize {
		return fmt.Errorf("queue is full (max %d jobs)", q.maxSize)
	}

	q.jobs[job.ID] = job
	q.pending <- job
	return nil
}

// Dequeue removes and returns the next job from the queue
// Blocks until a job is available or context is canceled
func (q *Queue) Dequeue(ctx context.Context) (*Job, error) {
	select {
	case job := <-q.pending:
		return job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Get returns a job by ID
func (q *Queue) Get(id string) (*Job, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	job, ok := q.jobs[id]
	return job, ok
}

// Cancel cancels a job by ID
func (q *Queue) Cancel(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, ok := q.jobs[id]
	if !ok {
		return fmt.Errorf("job %s not found", id)
	}

	status := job.GetStatus()
	if status == JobStatusComplete || status == JobStatusFailed || status == JobStatusCanceled {
		return fmt.Errorf("job %s already finished", id)
	}

	job.SetStatus(JobStatusCanceled)
	return nil
}

// Remove removes a job from the queue (for cleanup)
func (q *Queue) Remove(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.jobs, id)
}

// ListActive returns all active (non-finished) jobs
func (q *Queue) ListActive() []*Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var active []*Job
	for _, job := range q.jobs {
		status := job.GetStatus()
		if status != JobStatusComplete && status != JobStatusFailed && status != JobStatusCanceled {
			active = append(active, job)
		}
	}
	return active
}

// ListRecent returns recently completed jobs (within duration)
func (q *Queue) ListRecent(duration time.Duration) []*Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	cutoff := time.Now().Add(-duration)
	var recent []*Job
	for _, job := range q.jobs {
		if job.FinishedAt != nil && job.FinishedAt.After(cutoff) {
			recent = append(recent, job)
		}
	}
	return recent
}

// Stats returns queue statistics
func (q *Queue) Stats() QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := QueueStats{
		TotalJobs:  len(q.jobs),
		QueuedJobs: len(q.pending),
	}

	for _, job := range q.jobs {
		switch job.GetStatus() {
		case JobStatusQueued:
			// Already counted in QueuedJobs
		case JobStatusCloning, JobStatusScanning:
			stats.RunningJobs++
		case JobStatusComplete:
			stats.CompletedJobs++
		case JobStatusFailed:
			stats.FailedJobs++
		case JobStatusCanceled:
			stats.CanceledJobs++
		}
	}

	return stats
}

// Cleanup removes old finished jobs (older than duration)
func (q *Queue) Cleanup(maxAge time.Duration) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, job := range q.jobs {
		if job.FinishedAt != nil && job.FinishedAt.Before(cutoff) {
			delete(q.jobs, id)
			removed++
		}
	}

	return removed
}

// QueueStats holds queue statistics
type QueueStats struct {
	TotalJobs     int `json:"total_jobs"`
	QueuedJobs    int `json:"queued_jobs"`
	RunningJobs   int `json:"running_jobs"`
	CompletedJobs int `json:"completed_jobs"`
	FailedJobs    int `json:"failed_jobs"`
	CanceledJobs  int `json:"canceled_jobs"`
}

// GenerateJobID creates a unique job ID
func GenerateJobID() string {
	return fmt.Sprintf("scan-%s", time.Now().Format("20060102-150405"))
}
