// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package sheets

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/sheets/v4"
)

// ExporterConfig holds configuration for the sheets exporter
type ExporterConfig struct {
	ZeroHome      string
	Title         string
	SpreadsheetID string // If set, update existing spreadsheet
	IncludeSheets []string
	ShareWith     []string
	ChartsEnabled bool
}

// Exporter handles exporting data to Google Sheets
type Exporter struct {
	client      *Client
	transformer *Transformer
	config      ExporterConfig
}

// NewExporter creates a new sheets exporter
func NewExporter(client *Client, config ExporterConfig) *Exporter {
	return &Exporter{
		client:      client,
		transformer: NewTransformer(config.ZeroHome),
		config:      config,
	}
}

// ExportResult contains information about the exported spreadsheet
type ExportResult struct {
	SpreadsheetID  string
	SpreadsheetURL string
	SheetsCreated  []string
	RowsExported   int
}

// ExportOrg exports organization data to Google Sheets
func (e *Exporter) ExportOrg(ctx context.Context, orgName string, progressFn func(status string)) (*ExportResult, error) {
	// Load org data
	progressFn(fmt.Sprintf("Loading scan data for %s...", orgName))
	orgData, err := e.transformer.LoadOrgData(orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to load org data: %w", err)
	}

	if len(orgData.Repos) == 0 {
		return nil, fmt.Errorf("no scan data found for %s. Run 'zero hydrate %s' first", orgName, orgName)
	}

	orgData.GeneratedAt = time.Now().Format("2006-01-02 15:04")

	// Create or get spreadsheet
	var spreadsheet *sheets.Spreadsheet
	if e.config.SpreadsheetID != "" {
		progressFn("Updating existing spreadsheet...")
		spreadsheet, err = e.client.GetSpreadsheet(ctx, e.config.SpreadsheetID)
		if err != nil {
			return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
		}
	} else {
		title := e.config.Title
		if title == "" {
			title = fmt.Sprintf("Zero Security Dashboard - %s", orgName)
		}
		progressFn(fmt.Sprintf("Creating spreadsheet: %s", title))
		spreadsheet, err = e.client.CreateSpreadsheet(ctx, title)
		if err != nil {
			return nil, fmt.Errorf("failed to create spreadsheet: %w", err)
		}
	}

	result := &ExportResult{
		SpreadsheetID:  spreadsheet.SpreadsheetId,
		SpreadsheetURL: GetSpreadsheetURL(spreadsheet.SpreadsheetId),
		SheetsCreated:  make([]string, 0),
	}

	// Track sheets to create
	sheetsToBuild := e.getSheetsToBuild()

	// Create sheets
	for _, sheetName := range sheetsToBuild {
		progressFn(fmt.Sprintf("Creating %s...", sheetName))

		var rows int
		var sheetErr error

		switch sheetName {
		case "Executive Summary":
			rows, sheetErr = e.createSummarySheet(ctx, spreadsheet.SpreadsheetId, orgData)
		case "Vulnerability Tracker":
			rows, sheetErr = e.createVulnSheet(ctx, spreadsheet.SpreadsheetId, orgData)
		case "Secrets & Compliance":
			rows, sheetErr = e.createComplianceSheet(ctx, spreadsheet.SpreadsheetId, orgData)
		case "AI/ML Inventory":
			rows, sheetErr = e.createMLSheet(ctx, spreadsheet.SpreadsheetId, orgData)
		case "All Findings":
			rows, sheetErr = e.createRawDataSheet(ctx, spreadsheet.SpreadsheetId, orgData)
		}

		if sheetErr != nil {
			progressFn(fmt.Sprintf("Warning: Failed to create %s: %v", sheetName, sheetErr))
			continue
		}

		result.SheetsCreated = append(result.SheetsCreated, sheetName)
		result.RowsExported += rows

		progressFn(fmt.Sprintf("  ✓ %s (%d rows)", sheetName, rows))
	}

	// Delete the default "Sheet1" if we created new sheets
	if e.config.SpreadsheetID == "" && len(result.SheetsCreated) > 0 {
		for _, sheet := range spreadsheet.Sheets {
			if sheet.Properties.Title == "Sheet1" {
				e.client.DeleteSheet(ctx, spreadsheet.SpreadsheetId, sheet.Properties.SheetId)
				break
			}
		}
	}

	return result, nil
}

func (e *Exporter) getSheetsToBuild() []string {
	allSheets := []string{
		"Executive Summary",
		"Vulnerability Tracker",
		"Secrets & Compliance",
		"AI/ML Inventory",
		"All Findings",
	}

	if len(e.config.IncludeSheets) == 0 {
		return allSheets
	}

	// Filter to requested sheets
	sheetMap := make(map[string]bool)
	for _, s := range e.config.IncludeSheets {
		sheetMap[s] = true
	}

	result := make([]string, 0)
	for _, s := range allSheets {
		if sheetMap[s] || sheetMap[sheetAbbrev(s)] {
			result = append(result, s)
		}
	}

	return result
}

func sheetAbbrev(name string) string {
	switch name {
	case "Executive Summary":
		return "summary"
	case "Vulnerability Tracker":
		return "vulns"
	case "Secrets & Compliance":
		return "compliance"
	case "AI/ML Inventory":
		return "ml"
	case "All Findings":
		return "raw"
	}
	return name
}

func (e *Exporter) createSummarySheet(ctx context.Context, spreadsheetID string, orgData *OrgData) (int, error) {
	sheetID, err := e.client.AddSheet(ctx, spreadsheetID, "Executive Summary")
	if err != nil {
		return 0, err
	}

	// Build summary data
	rows := [][]interface{}{
		{"ORGANIZATION SECURITY DASHBOARD"},
		{fmt.Sprintf("Generated: %s | Repos: %d", orgData.GeneratedAt, orgData.TotalRepos)},
		{},
		{"RISK OVERVIEW"},
		{"Metric", "Count"},
		{"Critical Vulnerabilities", orgData.CriticalVulns},
		{"High Vulnerabilities", orgData.HighVulns},
		{"Medium Vulnerabilities", orgData.MediumVulns},
		{"Low Vulnerabilities", orgData.LowVulns},
		{"Total Vulnerabilities", orgData.TotalVulns},
		{},
		{"SECURITY STATS"},
		{"Metric", "Count"},
		{"Secrets Found", orgData.TotalSecrets},
		{"License Violations", orgData.TotalLicenses},
		{"ML Models", orgData.TotalMLModels},
		{"Repos with Bus Factor Risk", orgData.BusFactorRisk},
		{},
		{"TOP RISK REPOS"},
		{"Repository", "Critical", "High", "Medium", "Low", "Total"},
	}

	// Add top 10 repos by risk
	for i, repo := range orgData.Repos {
		if i >= 10 {
			break
		}
		rows = append(rows, []interface{}{
			fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
			repo.Summary.CriticalVulns,
			repo.Summary.HighVulns,
			repo.Summary.MediumVulns,
			repo.Summary.LowVulns,
			repo.Summary.TotalVulns,
		})
	}

	// Write data
	rangeStr := fmt.Sprintf("'Executive Summary'!A1")
	if err := e.client.UpdateValues(ctx, spreadsheetID, rangeStr, rows); err != nil {
		return 0, err
	}

	// Apply formatting
	requests := []*sheets.Request{
		// Title formatting
		{
			RepeatCell: &sheets.RepeatCellRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    0,
					EndRowIndex:      1,
					StartColumnIndex: 0,
					EndColumnIndex:   6,
				},
				Cell: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						BackgroundColor: ColorDashboardBg,
						TextFormat: &sheets.TextFormat{
							Bold:            true,
							FontSize:        18,
							ForegroundColor: ColorDashboardText,
						},
					},
				},
				Fields: "userEnteredFormat",
			},
		},
		// Subtitle formatting
		{
			RepeatCell: &sheets.RepeatCellRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    1,
					EndRowIndex:      2,
					StartColumnIndex: 0,
					EndColumnIndex:   6,
				},
				Cell: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						BackgroundColor: ColorDashboardBg,
						TextFormat: &sheets.TextFormat{
							FontSize:        11,
							ForegroundColor: ColorDashboardText,
						},
					},
				},
				Fields: "userEnteredFormat",
			},
		},
	}

	// Section headers
	sectionRows := []int64{3, 10, 17}
	for _, row := range sectionRows {
		requests = append(requests, CreateSectionHeader(sheetID, row, 0, 6))
	}

	// Freeze header row
	requests = append(requests, &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				SheetId: sheetID,
				GridProperties: &sheets.GridProperties{
					FrozenRowCount: 2,
				},
			},
			Fields: "gridProperties.frozenRowCount",
		},
	})

	// Auto-resize columns
	requests = append(requests, CreateAutoResizeRequest(sheetID, 0, 6))

	if _, err := e.client.BatchUpdate(ctx, spreadsheetID, requests); err != nil {
		return 0, err
	}

	// Add chart if enabled
	if e.config.ChartsEnabled {
		chartRequests := []*sheets.Request{
			// Severity pie chart - create data range for it
			CreatePieChartRequest(
				ChartPosition{SheetID: sheetID, Row: 3, Col: 3, Width: 400, Height: 250},
				"Severity Distribution",
				sheetID,
				&sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    4,
					EndRowIndex:      9,
					StartColumnIndex: 0,
					EndColumnIndex:   2,
				},
			),
		}

		if _, err := e.client.BatchUpdate(ctx, spreadsheetID, chartRequests); err != nil {
			// Non-fatal - continue without chart
			fmt.Printf("Warning: Failed to add chart: %v\n", err)
		}
	}

	return len(rows), nil
}

func (e *Exporter) createVulnSheet(ctx context.Context, spreadsheetID string, orgData *OrgData) (int, error) {
	sheetID, err := e.client.AddSheet(ctx, spreadsheetID, "Vulnerability Tracker")
	if err != nil {
		return 0, err
	}

	rows := orgData.ToVulnRows()

	// Write data
	rangeStr := fmt.Sprintf("'Vulnerability Tracker'!A1")
	if err := e.client.UpdateValues(ctx, spreadsheetID, rangeStr, rows); err != nil {
		return 0, err
	}

	endRow := int64(len(rows))
	endCol := int64(10) // Number of columns

	// Apply formatting
	requests := []*sheets.Request{
		CreateHeaderFormatRequest(sheetID, endCol),
	}

	// Severity column formatting (column 4, index 3)
	requests = append(requests, CreateSeverityFormatRule(sheetID, 3, 1, endRow)...)

	// KEV column highlighting (column 8, index 7)
	requests = append(requests, CreateBooleanHighlightRule(sheetID, 7, 1, endRow, ColorCritical))

	// Add filter
	requests = append(requests, &sheets.Request{
		SetBasicFilter: &sheets.SetBasicFilterRequest{
			Filter: &sheets.BasicFilter{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    0,
					EndRowIndex:      endRow,
					StartColumnIndex: 0,
					EndColumnIndex:   endCol,
				},
			},
		},
	})

	// Freeze header row
	requests = append(requests, &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				SheetId: sheetID,
				GridProperties: &sheets.GridProperties{
					FrozenRowCount: 1,
				},
			},
			Fields: "gridProperties.frozenRowCount",
		},
	})

	// Auto-resize columns
	requests = append(requests, CreateAutoResizeRequest(sheetID, 0, endCol))

	if _, err := e.client.BatchUpdate(ctx, spreadsheetID, requests); err != nil {
		return 0, err
	}

	return len(rows) - 1, nil // Subtract header
}

func (e *Exporter) createComplianceSheet(ctx context.Context, spreadsheetID string, orgData *OrgData) (int, error) {
	sheetID, err := e.client.AddSheet(ctx, spreadsheetID, "Secrets & Compliance")
	if err != nil {
		return 0, err
	}

	// Build combined compliance data
	rows := [][]interface{}{
		{"SECRETS"},
	}
	secretRows := orgData.ToSecretRows()
	rows = append(rows, secretRows...)

	rows = append(rows, []interface{}{})
	rows = append(rows, []interface{}{"LICENSE COMPLIANCE"})
	licenseRows := orgData.ToLicenseRows()
	rows = append(rows, licenseRows...)

	rows = append(rows, []interface{}{})
	rows = append(rows, []interface{}{"CODEOWNERS COVERAGE"})
	codeownersRows := orgData.ToCODEOWNERSRows()
	rows = append(rows, codeownersRows...)

	// Write data
	rangeStr := fmt.Sprintf("'Secrets & Compliance'!A1")
	if err := e.client.UpdateValues(ctx, spreadsheetID, rangeStr, rows); err != nil {
		return 0, err
	}

	// Apply formatting
	requests := []*sheets.Request{
		// Section header formatting for "SECRETS"
		CreateSectionHeader(sheetID, 0, 0, 8),
	}

	// Severity formatting for secrets (column 5 - severity)
	secretEndRow := int64(len(secretRows) + 1)
	requests = append(requests, CreateSeverityFormatRule(sheetID, 4, 2, secretEndRow)...)

	// Freeze first row
	requests = append(requests, &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				SheetId: sheetID,
				GridProperties: &sheets.GridProperties{
					FrozenRowCount: 1,
				},
			},
			Fields: "gridProperties.frozenRowCount",
		},
	})

	// Auto-resize columns
	requests = append(requests, CreateAutoResizeRequest(sheetID, 0, 8))

	if _, err := e.client.BatchUpdate(ctx, spreadsheetID, requests); err != nil {
		return 0, err
	}

	return len(rows), nil
}

func (e *Exporter) createMLSheet(ctx context.Context, spreadsheetID string, orgData *OrgData) (int, error) {
	sheetID, err := e.client.AddSheet(ctx, spreadsheetID, "AI/ML Inventory")
	if err != nil {
		return 0, err
	}

	// Build ML data
	rows := [][]interface{}{
		{"ML MODELS"},
	}
	modelRows := orgData.ToMLModelRows()
	rows = append(rows, modelRows...)

	rows = append(rows, []interface{}{})
	rows = append(rows, []interface{}{"ML FRAMEWORKS"})
	frameworkRows := orgData.ToFrameworkRows()
	rows = append(rows, frameworkRows...)

	rows = append(rows, []interface{}{})
	rows = append(rows, []interface{}{"AI SECURITY FINDINGS"})
	findingRows := orgData.ToAIFindingRows()
	rows = append(rows, findingRows...)

	// Write data
	rangeStr := fmt.Sprintf("'AI/ML Inventory'!A1")
	if err := e.client.UpdateValues(ctx, spreadsheetID, rangeStr, rows); err != nil {
		return 0, err
	}

	// Apply formatting
	requests := []*sheets.Request{
		CreateSectionHeader(sheetID, 0, 0, 7),
	}

	// Freeze first row
	requests = append(requests, &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				SheetId: sheetID,
				GridProperties: &sheets.GridProperties{
					FrozenRowCount: 1,
				},
			},
			Fields: "gridProperties.frozenRowCount",
		},
	})

	// Auto-resize columns
	requests = append(requests, CreateAutoResizeRequest(sheetID, 0, 7))

	if _, err := e.client.BatchUpdate(ctx, spreadsheetID, requests); err != nil {
		return 0, err
	}

	return len(rows), nil
}

func (e *Exporter) createRawDataSheet(ctx context.Context, spreadsheetID string, orgData *OrgData) (int, error) {
	sheetID, err := e.client.AddSheet(ctx, spreadsheetID, "All Findings")
	if err != nil {
		return 0, err
	}

	// Build comprehensive raw data
	rows := [][]interface{}{
		{"Repo", "Category", "Type", "Severity", "ID/Rule", "Package/File", "Title/Message", "Details"},
	}

	// Add all vulns
	for _, repo := range orgData.Repos {
		for _, v := range repo.Vulns {
			rows = append(rows, []interface{}{
				v.Repo, "Vulnerability", "Package", v.Severity, v.CVE, v.Package, v.Title, v.FixedVersion,
			})
		}
	}

	// Add all secrets
	for _, repo := range orgData.Repos {
		for _, s := range repo.Secrets {
			rows = append(rows, []interface{}{
				s.Repo, "Secret", s.Type, s.Severity, s.Detection, s.File, fmt.Sprintf("Line %d", s.Line), s.RotationPriority,
			})
		}
	}

	// Add all code vulns
	for _, repo := range orgData.Repos {
		for _, v := range repo.CodeVulns {
			rows = append(rows, []interface{}{
				v.Repo, "Code Security", v.Category, v.Severity, v.RuleID, v.File, v.Title, fmt.Sprintf("Line %d", v.Line),
			})
		}
	}

	// Add all AI findings
	for _, repo := range orgData.Repos {
		for _, f := range repo.AIFindings {
			rows = append(rows, []interface{}{
				f.Repo, "AI Security", f.Category, f.Severity, "", f.File, f.Title, f.Remediation,
			})
		}
	}

	// Write data
	rangeStr := fmt.Sprintf("'All Findings'!A1")
	if err := e.client.UpdateValues(ctx, spreadsheetID, rangeStr, rows); err != nil {
		return 0, err
	}

	endRow := int64(len(rows))
	endCol := int64(8)

	// Apply formatting
	requests := []*sheets.Request{
		CreateHeaderFormatRequest(sheetID, endCol),
	}

	// Severity formatting (column 4, index 3)
	requests = append(requests, CreateSeverityFormatRule(sheetID, 3, 1, endRow)...)

	// Add filter
	requests = append(requests, &sheets.Request{
		SetBasicFilter: &sheets.SetBasicFilterRequest{
			Filter: &sheets.BasicFilter{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    0,
					EndRowIndex:      endRow,
					StartColumnIndex: 0,
					EndColumnIndex:   endCol,
				},
			},
		},
	})

	// Freeze header row
	requests = append(requests, &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				SheetId: sheetID,
				GridProperties: &sheets.GridProperties{
					FrozenRowCount: 1,
				},
			},
			Fields: "gridProperties.frozenRowCount",
		},
	})

	// Auto-resize columns
	requests = append(requests, CreateAutoResizeRequest(sheetID, 0, endCol))

	if _, err := e.client.BatchUpdate(ctx, spreadsheetID, requests); err != nil {
		return 0, err
	}

	return len(rows) - 1, nil // Subtract header
}
