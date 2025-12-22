package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crashappsec/zero/pkg/workflow/automation"
)

func TestWatchCommandFlagDefaults(t *testing.T) {
	// Verify default flag values
	if watchDebounce != 2 {
		t.Errorf("Expected watchDebounce default to be 2, got %d", watchDebounce)
	}
}

func TestWatchConfigDefaults(t *testing.T) {
	cfg := automation.DefaultWatchConfig()

	if len(cfg.Patterns) == 0 {
		t.Error("Expected default patterns to be set")
	}

	if len(cfg.IgnorePatterns) == 0 {
		t.Error("Expected default ignore patterns to be set")
	}

	if cfg.DebounceDuration != 2*time.Second {
		t.Errorf("Expected debounce to be 2s, got %s", cfg.DebounceDuration)
	}

	if !cfg.RunOnStart {
		t.Error("Expected RunOnStart to be true by default")
	}
}

func TestWatcherBasicOperation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(testFile, []byte("package main"), 0644)

	eventsChan := make(chan []automation.WatchEvent, 10)

	cfg := automation.WatchConfig{
		Paths:            []string{tmpDir},
		Patterns:         []string{"*.go"},
		IgnorePatterns:   []string{},
		DebounceDuration: 100 * time.Millisecond,
		RunOnStart:       true,
	}

	watcher := automation.NewWatcher(cfg, func(events []automation.WatchEvent) {
		eventsChan <- events
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	// Should receive startup event
	select {
	case events := <-eventsChan:
		if len(events) == 0 {
			t.Error("Expected startup event")
		}
		if events[0].Operation != "startup" {
			t.Errorf("Expected startup operation, got %s", events[0].Operation)
		}
	case <-ctx.Done():
		t.Fatal("Timed out waiting for startup event")
	}
}

func TestWatcherDetectsFileChange(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(testFile, []byte("package main"), 0644)

	eventsChan := make(chan []automation.WatchEvent, 10)

	cfg := automation.WatchConfig{
		Paths:            []string{tmpDir},
		Patterns:         []string{"*.go"},
		IgnorePatterns:   []string{},
		DebounceDuration: 100 * time.Millisecond,
		RunOnStart:       false, // Skip startup to isolate file change detection
	}

	watcher := automation.NewWatcher(cfg, func(events []automation.WatchEvent) {
		eventsChan <- events
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	// Wait for initial state to be built
	time.Sleep(200 * time.Millisecond)

	// Modify the file
	os.WriteFile(testFile, []byte("package main\nfunc main() {}"), 0644)

	// Should receive modify event
	select {
	case events := <-eventsChan:
		found := false
		for _, e := range events {
			if e.Operation == "modify" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected modify event")
		}
	case <-ctx.Done():
		t.Fatal("Timed out waiting for modify event")
	}
}

func TestWatcherIgnoresPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories
	os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)

	eventReceived := false

	cfg := automation.WatchConfig{
		Paths:            []string{tmpDir},
		Patterns:         []string{"*.js"},
		IgnorePatterns:   []string{"node_modules/**"},
		DebounceDuration: 100 * time.Millisecond,
		RunOnStart:       false,
	}

	watcher := automation.NewWatcher(cfg, func(events []automation.WatchEvent) {
		for _, e := range events {
			if e.Operation != "startup" {
				eventReceived = true
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	// Wait for initial state
	time.Sleep(200 * time.Millisecond)

	// Create file in ignored directory
	os.WriteFile(filepath.Join(tmpDir, "node_modules", "test.js"), []byte(""), 0644)

	// Wait for potential event
	time.Sleep(300 * time.Millisecond)

	if eventReceived {
		t.Error("Should not receive events for ignored paths")
	}
}

func TestWatcherIsRunning(t *testing.T) {
	cfg := automation.DefaultWatchConfig()
	cfg.Paths = []string{"."}
	cfg.RunOnStart = false

	watcher := automation.NewWatcher(cfg, func(events []automation.WatchEvent) {})

	if watcher.IsRunning() {
		t.Error("Expected watcher to not be running initially")
	}

	ctx, cancel := context.WithCancel(context.Background())
	watcher.Start(ctx)

	if !watcher.IsRunning() {
		t.Error("Expected watcher to be running after start")
	}

	cancel()
	watcher.Stop()

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)

	if watcher.IsRunning() {
		t.Error("Expected watcher to not be running after stop")
	}
}
