// Package automation provides file watching and automated scan triggering
package automation

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// WatchConfig configures the file watcher
type WatchConfig struct {
	// Paths to watch (directories)
	Paths []string `json:"paths"`

	// File patterns to watch (glob patterns)
	Patterns []string `json:"patterns"`

	// Patterns to ignore
	IgnorePatterns []string `json:"ignore_patterns"`

	// Debounce duration - wait this long after last change before triggering
	DebounceDuration time.Duration `json:"debounce_duration"`

	// Scanners to run when changes detected
	Scanners []string `json:"scanners"`

	// Whether to run on startup
	RunOnStart bool `json:"run_on_start"`
}

// DefaultWatchConfig returns sensible defaults for watching
func DefaultWatchConfig() WatchConfig {
	return WatchConfig{
		Paths: []string{"."},
		Patterns: []string{
			"*.go", "*.py", "*.js", "*.ts", "*.java", "*.rb", "*.rs",
			"package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml",
			"go.mod", "go.sum", "requirements.txt", "Pipfile.lock",
			"Gemfile.lock", "Cargo.lock", "pom.xml", "build.gradle",
			"Dockerfile", "docker-compose.yml", "*.yaml", "*.yml",
			".github/workflows/*.yml",
		},
		IgnorePatterns: []string{
			"node_modules/**", ".git/**", "vendor/**", "__pycache__/**",
			"*.log", "*.tmp", ".zero/**", "dist/**", "build/**",
		},
		DebounceDuration: 2 * time.Second,
		Scanners:         []string{"sbom", "code-security"},
		RunOnStart:       true,
	}
}

// WatchEvent represents a file change event
type WatchEvent struct {
	Path      string    `json:"path"`
	Operation string    `json:"operation"` // create, modify, delete, rename
	Time      time.Time `json:"time"`
}

// WatchCallback is called when changes are detected
type WatchCallback func(events []WatchEvent)

// Watcher monitors files for changes and triggers callbacks
type Watcher struct {
	config   WatchConfig
	callback WatchCallback
	events   []WatchEvent
	eventsMu sync.Mutex
	timer    *time.Timer
	running  bool
	stopCh   chan struct{}
	mu       sync.Mutex
}

// NewWatcher creates a new file watcher
func NewWatcher(config WatchConfig, callback WatchCallback) *Watcher {
	return &Watcher{
		config:   config,
		callback: callback,
		events:   make([]WatchEvent, 0),
		stopCh:   make(chan struct{}),
	}
}

// Start begins watching for file changes
func (w *Watcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = true
	w.mu.Unlock()

	// Run on startup if configured
	if w.config.RunOnStart {
		w.callback([]WatchEvent{{
			Path:      ".",
			Operation: "startup",
			Time:      time.Now(),
		}})
	}

	// Start polling for changes
	go w.pollLoop(ctx)

	return nil
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.running = false
	close(w.stopCh)

	if w.timer != nil {
		w.timer.Stop()
	}
}

// pollLoop periodically checks for file changes
func (w *Watcher) pollLoop(ctx context.Context) {
	// Build initial state
	state := w.buildState()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			newState := w.buildState()
			events := w.compareStates(state, newState)
			if len(events) > 0 {
				w.queueEvents(events)
			}
			state = newState
		}
	}
}

// fileState tracks a file's modification state
type fileState struct {
	ModTime time.Time
	Size    int64
}

// buildState builds the current state of watched files
func (w *Watcher) buildState() map[string]fileState {
	state := make(map[string]fileState)

	for _, basePath := range w.config.Paths {
		_ = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			// Skip directories
			if info.IsDir() {
				// Check if we should skip this directory
				if w.shouldIgnore(path) {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if file matches patterns
			if !w.matchesPatterns(path) {
				return nil
			}

			// Check if file should be ignored
			if w.shouldIgnore(path) {
				return nil
			}

			state[path] = fileState{
				ModTime: info.ModTime(),
				Size:    info.Size(),
			}
			return nil
		})
	}

	return state
}

// compareStates finds differences between two states
func (w *Watcher) compareStates(old, new map[string]fileState) []WatchEvent {
	var events []WatchEvent
	now := time.Now()

	// Check for new or modified files
	for path, newState := range new {
		if oldState, exists := old[path]; exists {
			if newState.ModTime != oldState.ModTime || newState.Size != oldState.Size {
				events = append(events, WatchEvent{
					Path:      path,
					Operation: "modify",
					Time:      now,
				})
			}
		} else {
			events = append(events, WatchEvent{
				Path:      path,
				Operation: "create",
				Time:      now,
			})
		}
	}

	// Check for deleted files
	for path := range old {
		if _, exists := new[path]; !exists {
			events = append(events, WatchEvent{
				Path:      path,
				Operation: "delete",
				Time:      now,
			})
		}
	}

	return events
}

// matchesPatterns checks if a file matches any of the watch patterns
func (w *Watcher) matchesPatterns(path string) bool {
	if len(w.config.Patterns) == 0 {
		return true
	}

	for _, pattern := range w.config.Patterns {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}

		// Also try matching the full path for patterns like ".github/workflows/*.yml"
		matched, _ = filepath.Match(pattern, path)
		if matched {
			return true
		}
	}

	return false
}

// shouldIgnore checks if a path should be ignored
func (w *Watcher) shouldIgnore(path string) bool {
	for _, pattern := range w.config.IgnorePatterns {
		// Handle directory patterns
		if strings.HasSuffix(pattern, "/**") {
			dir := strings.TrimSuffix(pattern, "/**")
			if strings.HasPrefix(path, dir+string(filepath.Separator)) || path == dir {
				return true
			}
		}

		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}

		matched, _ = filepath.Match(pattern, path)
		if matched {
			return true
		}
	}

	return false
}

// queueEvents adds events to the queue and starts/resets the debounce timer
func (w *Watcher) queueEvents(events []WatchEvent) {
	w.eventsMu.Lock()
	defer w.eventsMu.Unlock()

	w.events = append(w.events, events...)

	// Reset or start the debounce timer
	if w.timer != nil {
		w.timer.Stop()
	}

	w.timer = time.AfterFunc(w.config.DebounceDuration, func() {
		w.flushEvents()
	})
}

// flushEvents sends all queued events to the callback
func (w *Watcher) flushEvents() {
	w.eventsMu.Lock()
	events := w.events
	w.events = make([]WatchEvent, 0)
	w.eventsMu.Unlock()

	if len(events) > 0 && w.callback != nil {
		w.callback(events)
	}
}

// IsRunning returns whether the watcher is currently running
func (w *Watcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

// GetConfig returns the current configuration
func (w *Watcher) GetConfig() WatchConfig {
	return w.config
}

// SetConfig updates the configuration (stops and restarts if running)
func (w *Watcher) SetConfig(config WatchConfig) {
	w.config = config
}
