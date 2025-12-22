// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/crashappsec/zero/pkg/workflow/vex"
	"github.com/spf13/cobra"
)

var (
	vexOutput       string
	vexIncludeAll   bool
	vexAutoAnalyze  bool
	vexProductName  string
	vexProductVer   string
	vexSupplier     string
	vexFormat       string
)

var vexCmd = &cobra.Command{
	Use:   "vex <owner/repo>",
	Short: "Generate VEX (Vulnerability Exploitability eXchange) document",
	Long: `Generate a VEX document from scan results to communicate vulnerability status.

VEX (Vulnerability Exploitability eXchange) is a companion to SBOM that explains
whether vulnerabilities actually affect your product. It helps teams:

- Triage vulnerabilities by marking them as "not affected", "in triage", etc.
- Communicate exploitability status to downstream consumers
- Satisfy compliance requirements for vulnerability disclosure
- Reduce alert fatigue by documenting false positives

Output follows the CycloneDX VEX specification (v1.5).

Examples:
  zero vex owner/repo                          Generate VEX for a hydrated repo
  zero vex owner/repo -o vex.json              Output to specific file
  zero vex owner/repo --include-all            Include non-affected vulnerabilities
  zero vex owner/repo --product-name MyApp     Set product name in VEX
  zero vex owner/repo --format summary         Show summary only`,
	Args: cobra.ExactArgs(1),
	RunE: runVex,
}

func init() {
	rootCmd.AddCommand(vexCmd)

	vexCmd.Flags().StringVarP(&vexOutput, "output", "o", "", "Output file path (default: stdout or analysis/vex.json)")
	vexCmd.Flags().BoolVar(&vexIncludeAll, "include-all", false, "Include all vulnerabilities (even not_affected)")
	vexCmd.Flags().BoolVar(&vexAutoAnalyze, "auto-analyze", true, "Auto-analyze vulnerabilities using reachability data")
	vexCmd.Flags().StringVar(&vexProductName, "product-name", "", "Product name for VEX metadata")
	vexCmd.Flags().StringVar(&vexProductVer, "product-version", "", "Product version for VEX metadata")
	vexCmd.Flags().StringVar(&vexSupplier, "supplier", "", "Supplier/organization name")
	vexCmd.Flags().StringVar(&vexFormat, "format", "json", "Output format: json, summary")
}

func runVex(cmd *cobra.Command, args []string) error {
	term := terminal.New()
	repo := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	// Check if project exists
	analysisDir := filepath.Join(zeroHome, "repos", repo, "analysis")
	if _, err := os.Stat(analysisDir); os.IsNotExist(err) {
		term.Error("Project not found: %s", repo)
		term.Info("Run: zero hydrate %s", repo)
		return fmt.Errorf("project not found")
	}

	// Check if packages.json exists (required for vulnerability data)
	packagesPath := filepath.Join(analysisDir, "packages.json")
	if _, err := os.Stat(packagesPath); os.IsNotExist(err) {
		term.Error("No vulnerability data found for %s", repo)
		term.Info("Run: zero hydrate %s --profile packages", repo)
		return fmt.Errorf("packages.json not found")
	}

	// Configure generator
	genConfig := vex.DefaultConfig()
	genConfig.AutoAnalyze = vexAutoAnalyze
	genConfig.UseReachability = vexAutoAnalyze
	genConfig.IncludeAll = vexIncludeAll

	if vexProductName != "" {
		genConfig.ProductName = vexProductName
	}
	if vexProductVer != "" {
		genConfig.ProductVersion = vexProductVer
	}
	if vexSupplier != "" {
		genConfig.SupplierName = vexSupplier
	}

	// Generate VEX
	term.Info("Generating VEX document for %s...", repo)

	generator := vex.NewGenerator(genConfig)
	doc, err := generator.GenerateFromScanResults(analysisDir)
	if err != nil {
		return fmt.Errorf("failed to generate VEX: %w", err)
	}

	// Handle output format
	if vexFormat == "summary" {
		return printVexSummary(term, doc)
	}

	// Determine output path
	outputPath := vexOutput
	if outputPath == "" {
		// Default to analysis directory
		outputPath = filepath.Join(analysisDir, "vex.cdx.json")
	}

	// Write VEX document
	if err := doc.WriteJSON(outputPath); err != nil {
		return fmt.Errorf("failed to write VEX: %w", err)
	}

	term.Success("VEX document generated: %s", outputPath)

	// Print summary
	summary := doc.Summary()
	if summary["total"] == 0 {
		term.Info("No vulnerabilities found - VEX document is empty")
	} else {
		term.Info("")
		term.Info("VEX Summary:")
		term.Info("  Total vulnerabilities: %d", summary["total"])
		if summary["exploitable"] > 0 {
			term.Warning("  Exploitable: %d", summary["exploitable"])
		}
		if summary["in_triage"] > 0 {
			term.Info("  In triage: %d", summary["in_triage"])
		}
		if summary["not_affected"] > 0 {
			term.Success("  Not affected: %d", summary["not_affected"])
		}
		if summary["resolved"] > 0 {
			term.Success("  Resolved: %d", summary["resolved"])
		}
	}

	return nil
}

func printVexSummary(term *terminal.Terminal, doc *vex.Document) error {
	summary := doc.Summary()

	term.Info("VEX Summary")
	term.Info("===========")
	term.Info("")
	term.Info("Total vulnerabilities: %d", summary["total"])
	term.Info("")

	if summary["total"] == 0 {
		term.Success("No vulnerabilities found!")
		return nil
	}

	// Print by state
	states := []struct {
		state string
		label string
		color func(string, ...interface{})
	}{
		{"exploitable", "Exploitable", term.Error},
		{"in_triage", "In Triage", term.Warning},
		{"not_affected", "Not Affected", term.Success},
		{"resolved", "Resolved", term.Success},
		{"false_positive", "False Positive", term.Info},
	}

	for _, s := range states {
		if count := summary[s.state]; count > 0 {
			s.color("  %s: %d", s.label, count)
		}
	}

	// Print detailed breakdown
	term.Info("")
	term.Info("Vulnerability Details:")
	term.Info("")

	for _, v := range doc.Vulnerabilities {
		state := "in_triage"
		justification := ""
		if v.Analysis != nil {
			state = string(v.Analysis.State)
			justification = string(v.Analysis.Justification)
		}

		severity := "unknown"
		if len(v.Ratings) > 0 {
			severity = v.Ratings[0].Severity
		}

		stateIcon := "?"
		switch state {
		case "exploitable":
			stateIcon = "!"
		case "not_affected":
			stateIcon = "✓"
		case "resolved":
			stateIcon = "✓"
		case "in_triage":
			stateIcon = "?"
		}

		line := fmt.Sprintf("  [%s] %s (%s)", stateIcon, v.ID, severity)
		if justification != "" {
			line += fmt.Sprintf(" - %s", justification)
		}

		switch state {
		case "exploitable":
			term.Error(line)
		case "not_affected", "resolved":
			term.Success(line)
		default:
			term.Info(line)
		}

		// Print affected packages
		for _, a := range v.Affects {
			term.Info("      Affects: %s", a.Ref)
		}
	}

	return nil
}

// Also output as JSON to stdout if requested
func printVexJSON(doc *vex.Document) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
