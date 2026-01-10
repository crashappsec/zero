package cmd

import (
	"fmt"
	"strings"

	"github.com/crashappsec/zero/pkg/reports/markdown"
	"github.com/spf13/cobra"
)

var reportOpts struct {
	Category string
	Analyzer string
	Output   string
}

var reportCmd = &cobra.Command{
	Use:   "report <owner/repo>",
	Short: "Generate markdown reports from analysis data",
	Long: `Generate markdown reports from analysis data.

Reports can be generated for:
  - All pillars (default): Comprehensive engineering intelligence report
  - Specific pillar: --category speed|quality|team|security|supply-chain|technology
  - Specific analyzer: --analyzer code-security|code-packages|devops|etc.

The 6 Pillars of Engineering Intelligence:
  Productivity Pillars:
  - Speed:        DORA metrics, cycle time, delivery performance
  - Quality:      Tech debt, complexity, test coverage
  - Team:         Code ownership, bus factor, onboarding

  Technical Pillars:
  - Security:     Vulnerabilities, secrets, cryptographic issues
  - Supply Chain: Dependencies, licenses, malcontent, package health
  - Technology:   Stack detection, ML-BOM, AI/ML findings

Examples:
  zero report expressjs/express                      # Full report to stdout
  zero report expressjs/express --output report.md  # Save to file
  zero report expressjs/express --category speed    # Speed/DORA report only
  zero report expressjs/express --category security # Security report only
  zero report expressjs/express --analyzer devops   # DevOps analyzer only`,
	Args: cobra.ExactArgs(1),
	RunE: runReport,
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVarP(&reportOpts.Category, "category", "c", "", "Report pillar (speed, quality, team, security, supply-chain, technology)")
	reportCmd.Flags().StringVarP(&reportOpts.Analyzer, "analyzer", "a", "", "Specific analyzer to report on")
	reportCmd.Flags().StringVarP(&reportOpts.Output, "output", "o", "", "Output file path (default: stdout)")
}

func runReport(cmd *cobra.Command, args []string) error {
	project := args[0]

	// Validate category if provided
	var category markdown.Category
	if reportOpts.Category != "" {
		category = markdown.Category(strings.ToLower(reportOpts.Category))
		validCategories := markdown.AllCategories()
		valid := false
		for _, c := range validCategories {
			if c == category {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid category: %s (valid: speed, quality, team, security, supply-chain, technology)", reportOpts.Category)
		}
	}

	opts := &markdown.Options{
		Project:  project,
		Category: category,
		Analyzer: reportOpts.Analyzer,
		Output:   reportOpts.Output,
	}

	gen, err := markdown.New(opts)
	if err != nil {
		return err
	}

	return gen.Generate()
}
