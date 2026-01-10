package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/crashappsec/zero/pkg/workflow/automation"
	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/workflow/freshness"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/spf13/cobra"
)

var (
	watchDebounce int
	watchProfile  string
	watchScanners []string
	watchIgnore   []string
)

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Watch for file changes and trigger scans",
	Long: `Watch a directory for file changes and automatically run scans.

This command monitors the specified directory (or current directory) for
changes to source files, dependencies, and configuration. When changes
are detected, it waits for activity to settle (debounce), then runs
the configured scanners.

Examples:
  zero watch                        Watch current directory
  zero watch /path/to/repo          Watch specific path
  zero watch --debounce 5           Wait 5 seconds after last change
  zero watch --scanners sbom,code-security   Only run specific scanners
  zero watch --profile quick        Use quick profile`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().IntVar(&watchDebounce, "debounce", 2, "Seconds to wait after last change before scanning")
	watchCmd.Flags().StringVar(&watchProfile, "profile", "", "Scan profile to use")
	watchCmd.Flags().StringSliceVar(&watchScanners, "scanners", nil, "Specific scanners to run (comma-separated)")
	watchCmd.Flags().StringSliceVar(&watchIgnore, "ignore", nil, "Additional patterns to ignore")
}

func runWatch(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Determine watch path
	watchPath := "."
	if len(args) > 0 {
		watchPath = args[0]
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(watchPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Check path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path not found: %s", absPath)
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	// Determine profile and scanners
	profile := watchProfile
	if profile == "" {
		profile = cfg.Settings.DefaultProfile
		if profile == "" {
			profile = "all-quick"
		}
	}

	var scannersToRun []string
	if len(watchScanners) > 0 {
		scannersToRun = watchScanners
	} else {
		scannersToRun, _ = cfg.GetProfileScanners(profile)
	}

	term := terminal.New()

	// Configure watcher
	watchConfig := automation.DefaultWatchConfig()
	watchConfig.Paths = []string{absPath}
	watchConfig.DebounceDuration = time.Duration(watchDebounce) * time.Second
	watchConfig.Scanners = scannersToRun
	watchConfig.RunOnStart = true

	// Add custom ignore patterns
	if len(watchIgnore) > 0 {
		watchConfig.IgnorePatterns = append(watchConfig.IgnorePatterns, watchIgnore...)
	}

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	term.Divider()
	term.Info("%s %s",
		term.Color(terminal.Bold, "Watching"),
		absPath,
	)
	term.Info("  Profile: %s", term.Color(terminal.Cyan, profile))
	term.Info("  Scanners: %s", term.Color(terminal.Cyan, strings.Join(scannersToRun, ", ")))
	term.Info("  Debounce: %ds", watchDebounce)
	term.Divider()
	term.Info("Press Ctrl+C to stop")
	term.Divider()

	// Track scan state
	scanCount := 0
	lastScanTime := time.Time{}

	// Create watcher with callback
	watcher := automation.NewWatcher(watchConfig, func(events []automation.WatchEvent) {
		scanCount++

		// Show what changed
		if len(events) > 0 && events[0].Operation != "startup" {
			term.Info("\n%s Changes detected:", term.Color(terminal.Yellow, "▸"))
			shown := 0
			for _, e := range events {
				if shown < 5 {
					term.Info("  %s %s", e.Operation, filepath.Base(e.Path))
					shown++
				}
			}
			if len(events) > 5 {
				term.Info("  ... and %d more", len(events)-5)
			}
		} else {
			term.Info("\n%s Initial scan", term.Color(terminal.Cyan, "▸"))
		}

		// Run scan
		runWatchScan(ctx, term, cfg, absPath, profile, scannersToRun, zeroHome)
		lastScanTime = time.Now()

		term.Divider()
		term.Info("Watching for changes... (scan #%d at %s)",
			scanCount,
			lastScanTime.Format("15:04:05"),
		)
	})

	// Start watcher
	if err := watcher.Start(ctx); err != nil {
		return fmt.Errorf("starting watcher: %w", err)
	}
	defer watcher.Stop()

	// Wait for signal
	<-sigChan
	term.Info("\nStopping watcher...")

	return nil
}

func runWatchScan(ctx context.Context, term *terminal.Terminal, cfg *config.Config, repoPath, profile string, scanners []string, zeroHome string) {
	// Determine repo name from path
	repoName := filepath.Base(repoPath)
	parentDir := filepath.Base(filepath.Dir(repoPath))
	if parentDir != "." && parentDir != "/" {
		repoName = parentDir + "/" + repoName
	}

	runner := scanner.NewRunner(zeroHome)
	freshMgr := freshness.NewManager(zeroHome)

	term.Info("  Running %d scanners...", len(scanners))
	start := time.Now()

	progress := scanner.NewProgress(scanners)
	result, err := runner.Run(ctx, repoName, profile, progress, nil)
	duration := time.Since(start)

	if err != nil {
		term.Error("  Scan failed: %v", err)
		return
	}

	if result.Success {
		// Record successful scan
		scanResults := make([]freshness.ScanResult, 0)
		for name, sr := range result.Results {
			errStr := ""
			if sr.Error != nil {
				errStr = sr.Error.Error()
			}
			scanResults = append(scanResults, freshness.ScanResult{
				Name:     name,
				Success:  sr.Status == scanner.StatusComplete,
				Duration: sr.Duration,
				Error:    errStr,
			})
		}
		_ = freshMgr.RecordScan(repoName, scanResults)

		term.Success("  Complete (%ds)", int(duration.Seconds()))

		// Show scanner breakdown
		for name, sr := range result.Results {
			status := term.Color(terminal.Green, "✓")
			if sr.Status != scanner.StatusComplete {
				status = term.Color(terminal.Red, "✗")
			}
			term.Info("    %s %s", status, name)
		}
	} else {
		term.Error("  Scan failed")
		for name, sr := range result.Results {
			if sr.Status != scanner.StatusComplete {
				errStr := "unknown error"
				if sr.Error != nil {
					errStr = sr.Error.Error()
				}
				term.Error("    ✗ %s: %s", name, errStr)
			}
		}
	}
}
