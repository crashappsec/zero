package cmd

import (
	"fmt"
	"os"

	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/evidence"
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

Examples:
  zero report expressjs/express           Generate and open report
  zero report expressjs/express --serve   Start live dev server
  zero report expressjs/express --regen   Force regenerate report
  zero report expressjs/express -o ./out  Custom output directory`,
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
	if !reportRegen && !reportServe && fileExists(reportPath) {
		if reportOpen {
			term.Info("Opening existing report...")
			gen.OpenBrowser(reportPath)
			term.Box(fmt.Sprintf("View Report: file://%s", reportPath))
			return nil
		}
		term.Info("Report: file://%s", reportPath)
		term.Info("Use --regenerate to rebuild the report")
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
		OpenBrowser: reportOpen,
		DevServer:   reportServe,
		Force:       reportRegen,
	}

	path, err := gen.Generate(opts)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	if !reportServe {
		term.Success("Report generated")
		term.Box(fmt.Sprintf("View Report: file://%s", path))
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
