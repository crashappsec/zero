package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/spf13/cobra"
)

var scanForce bool
var scanSkipSlow bool
var scanYes bool

var scanCmd = &cobra.Command{
	Use:   "scan <target> [profile]",
	Short: "Run scanners on already-cloned repositories",
	Long: `Run security scanners on repositories that have already been cloned.

Target can be:
  - owner/repo    Single repository (e.g., expressjs/express)
  - org-name      GitHub organization (e.g., phantom-tests)

The profile argument specifies which scanners to run. Profiles are defined
in the config file (config/zero.config.json).

Examples:
  zero scan expressjs/express          Scan single repo with default profile
  zero scan expressjs/express security Scan with security profile
  zero scan phantom-tests              Scan all repos in org
  zero scan phantom-tests quick        Scan org with quick profile
  zero scan owner/repo --force         Re-scan even if results exist`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVar(&scanForce, "force", false, "Re-scan even if results exist")
	scanCmd.Flags().BoolVar(&scanSkipSlow, "skip-slow", false, "Skip slow scanners")
	scanCmd.Flags().BoolVarP(&scanYes, "yes", "y", false, "Auto-accept prompts")
}

func runScan(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Parse target: owner/repo (single repo) or org-name (organization)
	target := args[0]
	var repo, org string
	profile := cfg.Settings.DefaultProfile
	if profile == "" {
		profile = "standard"
	}

	// Check if target is org or repo based on presence of "/"
	if strings.Contains(target, "/") {
		repo = target
	} else {
		org = target
	}

	// Second arg is profile if provided
	if len(args) > 1 {
		profile = args[1]
	}

	// Validate profile exists in config
	if _, ok := cfg.Profiles[profile]; !ok {
		availableProfiles := cfg.GetProfileNames()
		sort.Strings(availableProfiles)
		return fmt.Errorf("unknown profile: %s\n\nAvailable profiles:\n  %s\n\nProfiles are defined in config/zero.config.json",
			profile, strings.Join(availableProfiles, "\n  "))
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	term := terminal.New()
	runner := scanner.NewRunner(zeroHome)

	// Get repos to scan
	var repos []string
	if repo != "" {
		repos = []string{repo}
	} else {
		// Find all repos in org
		orgPath := filepath.Join(zeroHome, "repos", org)
		entries, err := os.ReadDir(orgPath)
		if err != nil {
			return fmt.Errorf("org not found: %s (run hydrate first)", org)
		}
		for _, e := range entries {
			if e.IsDir() {
				repos = append(repos, fmt.Sprintf("%s/%s", org, e.Name()))
			}
		}
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repos to scan")
	}

	// Get scanners for profile
	scanners, err := cfg.GetProfileScanners(profile)
	if err != nil {
		return err
	}

	term.Divider()
	term.Info("%s %d repos with profile %s",
		term.Color(terminal.Bold, "Scanning"),
		len(repos),
		term.Color(terminal.Cyan, profile),
	)
	term.Divider()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		term.Info("\nInterrupted...")
		cancel()
	}()

	success := 0
	failed := 0

	for _, r := range repos {
		term.Info("\n%s %s", term.Color(terminal.Cyan, "▸"), r)

		progress := scanner.NewProgress(scanners)
		result, err := runner.Run(ctx, r, profile, progress, nil)

		if err != nil {
			term.Error("  Failed: %v", err)
			failed++
			continue
		}

		if result.Success {
			term.Success("  Complete (%ds)", int(result.Duration.Seconds()))
			success++
		} else {
			term.Error("  Failed")
			failed++
		}
	}

	term.Divider()
	if failed > 0 {
		term.Info("Complete: %d success, %d failed", success, failed)
	} else {
		term.Success("Complete: %d repos scanned", success)
	}

	return nil
}
