package cmd

import (
	"github.com/crashappsec/zero/pkg/report"
	"github.com/spf13/cobra"
)

var reportOpts report.Options

var reportCmd = &cobra.Command{
	Use:   "report [owner/repo]",
	Short: "Generate security reports",
	Long: `Generate analysis reports in various formats.

Report Types:
  summary       High-level overview (default)
  security      Vulnerabilities, secrets, code issues
  licenses      License compliance
  sbom          Software Bill of Materials
  supply-chain  Dependencies, provenance, health
  full          Comprehensive report

Output Formats:
  text       Colored terminal output (default)
  json       Structured JSON
  markdown   GitHub-flavored markdown
  html       Self-contained HTML

Examples:
  zero report owner/repo                    Summary report
  zero report --org myorg --type security   Org security report
  zero report owner/repo --format markdown  Markdown output
  zero report owner/repo -o report.md       Write to file`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReport,
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVar(&reportOpts.Org, "org", "", "Organization to report on")
	reportCmd.Flags().StringVarP(&reportOpts.Type, "type", "t", "summary", "Report type")
	reportCmd.Flags().StringVarP(&reportOpts.Format, "format", "f", "text", "Output format")
	reportCmd.Flags().StringVarP(&reportOpts.Output, "output", "o", "", "Output file")
}

func runReport(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		reportOpts.Repo = args[0]
	}

	if reportOpts.Org == "" && reportOpts.Repo == "" {
		return cmd.Help()
	}

	r, err := report.New(&reportOpts)
	if err != nil {
		return err
	}
	return r.Run()
}
