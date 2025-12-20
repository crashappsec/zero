// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/export/sheets"
	"github.com/crashappsec/zero/pkg/terminal"
	"github.com/spf13/cobra"
)

var (
	sheetsSpreadsheetID string
	sheetsTitle         string
	sheetsInclude       string
	sheetsSingleRepo    bool
	sheetsShareWith     string
	sheetsNoCharts      bool
)

var sheetsCmd = &cobra.Command{
	Use:   "sheets <org>",
	Short: "Export scan results to Google Sheets dashboard",
	Long: `Export organization scan results to a rich Google Sheets dashboard with:
- Executive Summary with charts and metrics
- Vulnerability Tracker with severity highlighting
- Secrets & Compliance overview
- AI/ML Inventory (ML-BOM)
- Raw data for detailed analysis

First-time setup requires Google Cloud OAuth credentials.
See: https://console.cloud.google.com/apis/credentials

Examples:
  zero sheets crashoverride                           Export org to new spreadsheet
  zero sheets crashoverride --spreadsheet-id ABC123   Update existing spreadsheet
  zero sheets crashoverride/repo --single-repo        Export single repo
  zero sheets crashoverride --title "Q4 Report"       Custom title
  zero sheets crashoverride --sheets summary,vulns    Specific sheets only
  zero sheets crashoverride --share-with user@co.com  Share with email`,
	Args: cobra.ExactArgs(1),
	RunE: runSheets,
}

func init() {
	rootCmd.AddCommand(sheetsCmd)

	sheetsCmd.Flags().StringVar(&sheetsSpreadsheetID, "spreadsheet-id", "", "Update existing spreadsheet instead of creating new")
	sheetsCmd.Flags().StringVar(&sheetsTitle, "title", "", "Custom spreadsheet title (default: 'Zero Security Dashboard - <org>')")
	sheetsCmd.Flags().StringVar(&sheetsInclude, "sheets", "", "Comma-separated list of sheets to include (summary,vulns,compliance,ml,raw)")
	sheetsCmd.Flags().BoolVar(&sheetsSingleRepo, "single-repo", false, "Export single repo instead of org")
	sheetsCmd.Flags().StringVar(&sheetsShareWith, "share-with", "", "Comma-separated emails to share the spreadsheet with")
	sheetsCmd.Flags().BoolVar(&sheetsNoCharts, "no-charts", false, "Disable chart generation")
}

func runSheets(cmd *cobra.Command, args []string) error {
	term := terminal.New()
	ctx := context.Background()

	target := args[0]

	// Parse target (org or owner/repo)
	var orgName string
	if strings.Contains(target, "/") {
		if !sheetsSingleRepo {
			// Assume it's org/repo, extract org
			parts := strings.Split(target, "/")
			orgName = parts[0]
		} else {
			// Single repo mode
			parts := strings.Split(target, "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid repo format: expected owner/repo")
			}
			orgName = parts[0]
		}
	} else {
		orgName = target
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	// Check if sheets config exists
	sheetsCfg := cfg.GetSheetsConfig()
	if sheetsCfg.ClientID == "" || sheetsCfg.ClientSecret == "" {
		term.Warning("Google Sheets OAuth credentials not configured")
		term.Info("")
		term.Info("To set up Google Sheets export:")
		term.Info("1. Go to https://console.cloud.google.com/apis/credentials")
		term.Info("2. Create an OAuth 2.0 Client ID (Desktop application)")
		term.Info("3. Add the credentials to your config:")
		term.Info("")
		term.Info("   Add to config/zero.config.json:")
		term.Info("   \"sheets\": {")
		term.Info("     \"oauth_client_id\": \"<your-client-id>\",")
		term.Info("     \"oauth_client_secret\": \"<your-client-secret>\"")
		term.Info("   }")
		term.Info("")
		term.Info("Or set environment variables:")
		term.Info("   export ZERO_SHEETS_CLIENT_ID=<your-client-id>")
		term.Info("   export ZERO_SHEETS_CLIENT_SECRET=<your-client-secret>")
		return fmt.Errorf("OAuth credentials required")
	}

	// Initialize authenticator
	term.Divider()
	term.Info("Google Sheets Export")
	term.Divider()

	authConfig := sheets.AuthConfig{
		ClientID:     sheetsCfg.ClientID,
		ClientSecret: sheetsCfg.ClientSecret,
		TokenPath:    sheetsCfg.TokenPath,
	}

	auth := sheets.NewAuthenticator(authConfig)

	// Get authenticated client
	term.Info("Authenticating with Google...")
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	term.Success("Authenticated successfully")

	// Create sheets client
	client, err := sheets.NewClient(ctx, httpClient)
	if err != nil {
		return fmt.Errorf("failed to create sheets client: %w", err)
	}

	// Parse include sheets
	var includeSheets []string
	if sheetsInclude != "" {
		includeSheets = strings.Split(sheetsInclude, ",")
		for i, s := range includeSheets {
			includeSheets[i] = strings.TrimSpace(s)
		}
	}

	// Parse share emails
	var shareWith []string
	if sheetsShareWith != "" {
		shareWith = strings.Split(sheetsShareWith, ",")
		for i, s := range shareWith {
			shareWith[i] = strings.TrimSpace(s)
		}
	}

	// Create exporter
	exporterConfig := sheets.ExporterConfig{
		ZeroHome:      zeroHome,
		Title:         sheetsTitle,
		SpreadsheetID: sheetsSpreadsheetID,
		IncludeSheets: includeSheets,
		ShareWith:     shareWith,
		ChartsEnabled: !sheetsNoCharts,
	}

	exporter := sheets.NewExporter(client, exporterConfig)

	// Progress function
	progressFn := func(status string) {
		term.Info(status)
	}

	// Export
	term.Info("")
	result, err := exporter.ExportOrg(ctx, orgName, progressFn)
	if err != nil {
		return err
	}

	// Print result
	term.Info("")
	term.Divider()
	term.Success("Export complete!")
	term.Info("")
	term.Info("Spreadsheet URL:")
	term.Info("  %s", result.SpreadsheetURL)
	term.Info("")
	term.Info("Sheets created: %d", len(result.SheetsCreated))
	for _, sheet := range result.SheetsCreated {
		term.Info("  - %s", sheet)
	}
	term.Info("Total rows exported: %d", result.RowsExported)

	return nil
}
