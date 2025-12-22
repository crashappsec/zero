package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/crashappsec/zero/pkg/core/config"
	evidence "github.com/crashappsec/zero/pkg/reports"
	"github.com/spf13/cobra"
)

var (
	reportOutput string
	reportOpen   bool
	reportServe  bool
	reportRegen  bool
)

var reportCmd = &cobra.Command{
	Use:   "report <target>",
	Short: "Generate or view interactive HTML report",
	Long: `Generate an interactive report using Evidence.

Target can be:
  - owner/repo    Single repository (e.g., expressjs/express)
  - org-name      GitHub organization (e.g., phantom-tests)

The report includes engineering insights:
  - Security findings (vulnerabilities, secrets, crypto)
  - Dependencies and SBOM analysis
  - DevOps metrics and infrastructure
  - Code quality and ownership

By default, this command starts a local HTTP server and opens your browser.
Press Ctrl+C to stop the server when done viewing.

Examples:
  zero report expressjs/express              Single repo report
  zero report phantom-tests                  Org-wide report (all repos)
  zero report expressjs/express --serve      Live dev server (hot reload)
  zero report expressjs/express --regenerate Force regenerate
  zero report expressjs/express --open=false Generate without opening browser`,
	Args: cobra.ExactArgs(1),
	RunE: runReport,
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVarP(&reportOutput, "output", "o", "", "Output directory for report")
	reportCmd.Flags().BoolVar(&reportOpen, "open", true, "Open report in browser")
	reportCmd.Flags().BoolVar(&reportServe, "serve", false, "Start Evidence dev server")
	reportCmd.Flags().BoolVar(&reportRegen, "regenerate", false, "Force regenerate report")
}

func runReport(cmd *cobra.Command, args []string) error {
	target := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	gen := evidence.NewGenerator(cfg.ZeroHome())

	// Detect org vs repo by checking for slash
	isOrg := !strings.Contains(target, "/")

	if isOrg {
		return runOrgReport(gen, target)
	}

	return runRepoReport(gen, target)
}

func runRepoReport(gen *evidence.Generator, repo string) error {
	// Check if report already exists and we don't need to regenerate
	reportPath := gen.ReportPath(repo)
	needsBuild := reportRegen || reportServe || !fileExists(reportPath)

	if !needsBuild {
		if reportOpen {
			term.Info("Serving existing report...")
			gen.OpenBrowser(reportPath)
			return nil
		}
		term.Info("Report exists at: %s", reportPath)
		term.Info("Use --open to view, --regenerate to rebuild")
		return nil
	}

	// Check if analysis data exists
	analysisPath := gen.AnalysisPath(repo)
	if _, err := os.Stat(analysisPath); os.IsNotExist(err) {
		return fmt.Errorf("no analysis data found for %s\nRun: zero hydrate %s", repo, repo)
	}

	opts := evidence.Options{
		Repository:  repo,
		OutputDir:   reportOutput,
		OpenBrowser: false,
		DevServer:   reportServe,
		Force:       reportRegen,
	}

	path, err := gen.Generate(opts)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	if reportServe {
		return nil
	}

	term.Success("Report generated")

	if reportOpen {
		gen.OpenBrowser(path)
	} else {
		term.Info("Report: %s", path)
		term.Info("Use --open to view in browser")
	}

	return nil
}

func runOrgReport(gen *evidence.Generator, orgName string) error {
	term.Info("Generating org-wide report for %s...", orgName)

	opts := evidence.Options{
		OrgMode:     true,
		OrgName:     orgName,
		OutputDir:   reportOutput,
		OpenBrowser: false,
		DevServer:   reportServe,
		Force:       reportRegen,
	}

	path, err := gen.Generate(opts)
	if err != nil {
		return fmt.Errorf("generating org report: %w", err)
	}

	if reportServe {
		return nil
	}

	term.Success("Org report generated")

	if reportOpen {
		gen.OpenBrowser(path)
	} else {
		term.Info("Report: %s", path)
		term.Info("Use --open to view in browser")
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
