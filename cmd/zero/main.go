// Package main is the entry point for the zero CLI
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/crashappsec/zero/pkg/hydrate"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Set up context with cancellation on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nInterrupted, cleaning up...")
		cancel()
	}()

	// Parse command and flags
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: zero <command> [options]\n\nCommands:\n  hydrate    Clone and scan repositories")
	}

	switch os.Args[1] {
	case "hydrate":
		return runHydrate(ctx, os.Args[2:])
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func runHydrate(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("hydrate", flag.ExitOnError)

	var opts hydrate.Options
	var profile string

	fs.StringVar(&opts.Org, "org", "", "GitHub organization to scan")
	fs.IntVar(&opts.Limit, "limit", 100, "Maximum number of repos to process")
	fs.StringVar(&profile, "profile", "", "Scan profile (packages, security, full)")
	fs.BoolVar(&opts.Force, "force", false, "Force re-scan even if results exist")
	fs.BoolVar(&opts.SkipSlow, "skip-slow", false, "Skip slow scanners automatically")
	fs.BoolVar(&opts.Yes, "yes", false, "Auto-accept prompts")
	fs.BoolVar(&opts.Yes, "y", false, "Auto-accept prompts (shorthand)")
	fs.IntVar(&opts.Parallel, "parallel", 4, "Number of parallel jobs")

	// Parse named profile flags
	packages := fs.Bool("packages", false, "Use packages profile")
	security := fs.Bool("security", false, "Use security profile")
	full := fs.Bool("full", false, "Use full profile")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Determine profile
	if *packages {
		opts.Profile = "packages"
	} else if *security {
		opts.Profile = "security"
	} else if *full {
		opts.Profile = "full"
	} else if profile != "" {
		opts.Profile = profile
	} else {
		opts.Profile = "packages" // default
	}

	if opts.Org == "" {
		return fmt.Errorf("--org is required")
	}

	h, err := hydrate.New(&opts)
	if err != nil {
		return err
	}

	return h.Run(ctx)
}
