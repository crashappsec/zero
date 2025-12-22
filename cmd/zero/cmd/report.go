package cmd

import (
	"fmt"
	"os"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/reports"
	"github.com/spf13/cobra"
)

var (
	reportOutput string
	reportOpen   bool
	reportServe  bool
	reportRegen  bool
)

var reportCmd = &cobra.Command{
	Use:   "report <owner/repo>",
	Short: "Generate or view interactive HTML report",
	Long: `Generate a beautiful interactive report using Evidence.

The report includes:
  - Executive summary with severity breakdown
  - Security findings (vulnerabilities, secrets, crypto)
  - Dependencies and SBOM analysis
  - DevOps and infrastructure issues
  - Code quality and ownership metrics

By default, this command starts a local HTTP server and opens your browser.
Press Ctrl+C to stop the server when done viewing.

Examples:
  zero report expressjs/express              Generate and open report (Ctrl+C to stop)
  zero report expressjs/express --open=false Generate without opening browser
  zero report expressjs/express --serve      Start live dev server (hot reload)
  zero report expressjs/express --regen      Force regenerate report
  zero report expressjs/express -o ./out     Custom output directory`,
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
	repo := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	gen := evidence.NewGenerator(cfg.ZeroHome())

	// Check if report already exists and we don't need to regenerate
	reportPath := gen.ReportPath(repo)
	needsBuild := reportRegen || reportServe || !fileExists(reportPath)

	if !needsBuild {
		if reportOpen {
			term.Info("Serving existing report...")
			// OpenBrowser now starts HTTP server and blocks
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

	// Generate without opening - we'll open after
	opts := evidence.Options{
		Repository:  repo,
		OutputDir:   reportOutput,
		OpenBrowser: false, // Don't open yet
		DevServer:   reportServe,
		Force:       reportRegen,
	}

	path, err := gen.Generate(opts)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	// For dev server, Generate already handles serving
	if reportServe {
		return nil
	}

	term.Success("Report generated")

	// Now open with HTTP server if requested
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
