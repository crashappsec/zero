package automation

import (
	"context"
	"sync"
	"time"
)

// ScheduleConfig configures scheduled scanning
type ScheduleConfig struct {
	// Whether scheduling is enabled
	Enabled bool `json:"enabled"`

	// Interval between scheduled runs
	Interval time.Duration `json:"interval"`

	// Specific times to run (optional, overrides interval)
	Times []string `json:"times,omitempty"` // Format: "HH:MM"

	// Days of week to run (0=Sunday, 6=Saturday)
	DaysOfWeek []int `json:"days_of_week,omitempty"`

	// Repositories to scan (empty = all)
	Repositories []string `json:"repositories,omitempty"`

	// Scanners to run
	Scanners []string `json:"scanners"`
}

// DefaultScheduleConfig returns default schedule configuration
func DefaultScheduleConfig() ScheduleConfig {
	return ScheduleConfig{
		Enabled:    false,
		Interval:   24 * time.Hour,
		Scanners:   []string{"sbom", "package-analysis", "code-security"},
		DaysOfWeek: []int{1, 2, 3, 4, 5}, // Monday through Friday
	}
}

// ScheduleCallback is called when a scheduled run triggers
type ScheduleCallback func(repos []string, scanners []string)

// Scheduler handles scheduled scanning
type Scheduler struct {
	config   ScheduleConfig
	callback ScheduleCallback
	running  bool
	stopCh   chan struct{}
	mu       sync.Mutex
}

// NewScheduler creates a new scheduler
func NewScheduler(config ScheduleConfig, callback ScheduleCallback) *Scheduler {
	return &Scheduler{
		config:   config,
		callback: callback,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	if !s.config.Enabled {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	go s.runLoop(ctx)
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopCh)
}

// runLoop runs the scheduler loop
func (s *Scheduler) runLoop(ctx context.Context) {
	// Calculate next run time
	nextRun := s.calculateNextRun()
	timer := time.NewTimer(time.Until(nextRun))

	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-s.stopCh:
			timer.Stop()
			return
		case <-timer.C:
			// Check if today is a valid day
			if s.isValidDay() {
				s.callback(s.config.Repositories, s.config.Scanners)
			}

			// Calculate next run
			nextRun = s.calculateNextRun()
			timer.Reset(time.Until(nextRun))
		}
	}
}

// calculateNextRun calculates the next run time
func (s *Scheduler) calculateNextRun() time.Time {
	now := time.Now()

	if len(s.config.Times) > 0 {
		// Find next scheduled time
		return s.nextScheduledTime(now)
	}

	// Use interval
	return now.Add(s.config.Interval)
}

// nextScheduledTime finds the next time from the configured times
func (s *Scheduler) nextScheduledTime(now time.Time) time.Time {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Check each configured time
	for _, t := range s.config.Times {
		var hour, minute int
		if _, err := parseTime(t, &hour, &minute); err == nil {
			scheduled := today.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
			if scheduled.After(now) {
				return scheduled
			}
		}
	}

	// All times today have passed, schedule for tomorrow
	tomorrow := today.Add(24 * time.Hour)
	if len(s.config.Times) > 0 {
		var hour, minute int
		if _, err := parseTime(s.config.Times[0], &hour, &minute); err == nil {
			return tomorrow.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
		}
	}

	return now.Add(s.config.Interval)
}

// parseTime parses a time string in HH:MM format
func parseTime(s string, hour, minute *int) (bool, error) {
	n, err := parseTimeComponents(s, hour, minute)
	return n == 2, err
}

func parseTimeComponents(s string, hour, minute *int) (int, error) {
	// Simple parsing for HH:MM format
	if len(s) < 3 {
		return 0, nil
	}

	// Find colon position
	colonIdx := -1
	for i, c := range s {
		if c == ':' {
			colonIdx = i
			break
		}
	}

	if colonIdx == -1 {
		return 0, nil
	}

	// Parse hour
	*hour = 0
	for i := 0; i < colonIdx; i++ {
		if s[i] >= '0' && s[i] <= '9' {
			*hour = *hour*10 + int(s[i]-'0')
		}
	}

	// Parse minute
	*minute = 0
	for i := colonIdx + 1; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			*minute = *minute*10 + int(s[i]-'0')
		}
	}

	return 2, nil
}

// isValidDay checks if today is a valid day for running
func (s *Scheduler) isValidDay() bool {
	if len(s.config.DaysOfWeek) == 0 {
		return true // No restriction
	}

	today := int(time.Now().Weekday())
	for _, day := range s.config.DaysOfWeek {
		if day == today {
			return true
		}
	}

	return false
}

// IsRunning returns whether the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetConfig returns the current configuration
func (s *Scheduler) GetConfig() ScheduleConfig {
	return s.config
}

// SetConfig updates the configuration
func (s *Scheduler) SetConfig(config ScheduleConfig) {
	s.config = config
}

// NextRun returns when the next scheduled run will occur
func (s *Scheduler) NextRun() time.Time {
	return s.calculateNextRun()
}
