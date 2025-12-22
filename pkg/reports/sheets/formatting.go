// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package sheets

import (
	"google.golang.org/api/sheets/v4"
)

// Color definitions for conditional formatting
var (
	// Severity colors
	ColorCritical = &sheets.Color{Red: 0.92, Green: 0.26, Blue: 0.21, Alpha: 1} // Red
	ColorHigh     = &sheets.Color{Red: 1.0, Green: 0.6, Blue: 0.0, Alpha: 1}    // Orange
	ColorMedium   = &sheets.Color{Red: 1.0, Green: 0.85, Blue: 0.0, Alpha: 1}   // Yellow
	ColorLow      = &sheets.Color{Red: 0.56, Green: 0.77, Blue: 0.49, Alpha: 1} // Light Green
	ColorInfo     = &sheets.Color{Red: 0.68, Green: 0.85, Blue: 0.90, Alpha: 1} // Light Blue

	// Status colors
	ColorSuccess = &sheets.Color{Red: 0.26, Green: 0.62, Blue: 0.28, Alpha: 1} // Green
	ColorWarning = &sheets.Color{Red: 1.0, Green: 0.76, Blue: 0.03, Alpha: 1}  // Amber
	ColorError   = &sheets.Color{Red: 0.83, Green: 0.18, Blue: 0.18, Alpha: 1} // Dark Red

	// Header colors
	ColorHeaderBg   = &sheets.Color{Red: 0.2, Green: 0.2, Blue: 0.2, Alpha: 1}   // Dark Gray
	ColorHeaderText = &sheets.Color{Red: 1.0, Green: 1.0, Blue: 1.0, Alpha: 1}   // White
	ColorAltRow     = &sheets.Color{Red: 0.95, Green: 0.95, Blue: 0.95, Alpha: 1} // Light Gray

	// Dashboard colors
	ColorDashboardBg    = &sheets.Color{Red: 0.12, Green: 0.16, Blue: 0.22, Alpha: 1} // Dark blue-gray
	ColorDashboardText  = &sheets.Color{Red: 1.0, Green: 1.0, Blue: 1.0, Alpha: 1}    // White
	ColorDashboardAccent = &sheets.Color{Red: 0.26, Green: 0.52, Blue: 0.96, Alpha: 1} // Blue accent
)

// CreateSeverityFormatRule creates a conditional format rule for severity values
func CreateSeverityFormatRule(sheetID int64, colIndex, startRow, endRow int64) []*sheets.Request {
	rules := make([]*sheets.Request, 0)

	severities := []struct {
		value string
		color *sheets.Color
	}{
		{"critical", ColorCritical},
		{"high", ColorHigh},
		{"medium", ColorMedium},
		{"low", ColorLow},
		{"info", ColorInfo},
	}

	for _, sev := range severities {
		rules = append(rules, &sheets.Request{
			AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
				Rule: &sheets.ConditionalFormatRule{
					Ranges: []*sheets.GridRange{
						{
							SheetId:          sheetID,
							StartRowIndex:    startRow,
							EndRowIndex:      endRow,
							StartColumnIndex: colIndex,
							EndColumnIndex:   colIndex + 1,
						},
					},
					BooleanRule: &sheets.BooleanRule{
						Condition: &sheets.BooleanCondition{
							Type: "TEXT_EQ",
							Values: []*sheets.ConditionValue{
								{UserEnteredValue: sev.value},
							},
						},
						Format: &sheets.CellFormat{
							BackgroundColor: sev.color,
							TextFormat: &sheets.TextFormat{
								Bold: true,
								ForegroundColor: &sheets.Color{Red: 1, Green: 1, Blue: 1, Alpha: 1},
							},
						},
					},
				},
				Index: 0,
			},
		})
	}

	return rules
}

// CreateBooleanHighlightRule creates a rule to highlight "Yes" values
func CreateBooleanHighlightRule(sheetID int64, colIndex, startRow, endRow int64, highlightColor *sheets.Color) *sheets.Request {
	return &sheets.Request{
		AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
			Rule: &sheets.ConditionalFormatRule{
				Ranges: []*sheets.GridRange{
					{
						SheetId:          sheetID,
						StartRowIndex:    startRow,
						EndRowIndex:      endRow,
						StartColumnIndex: colIndex,
						EndColumnIndex:   colIndex + 1,
					},
				},
				BooleanRule: &sheets.BooleanRule{
					Condition: &sheets.BooleanCondition{
						Type: "TEXT_EQ",
						Values: []*sheets.ConditionValue{
							{UserEnteredValue: "Yes"},
						},
					},
					Format: &sheets.CellFormat{
						BackgroundColor: highlightColor,
						TextFormat: &sheets.TextFormat{
							Bold: true,
						},
					},
				},
			},
			Index: 0,
		},
	}
}

// CreateHeaderFormatRequest creates formatting for header rows
func CreateHeaderFormatRequest(sheetID int64, endCol int64) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    0,
				EndRowIndex:      1,
				StartColumnIndex: 0,
				EndColumnIndex:   endCol,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: ColorHeaderBg,
					TextFormat: &sheets.TextFormat{
						Bold:            true,
						ForegroundColor: ColorHeaderText,
						FontSize:        11,
					},
					HorizontalAlignment: "CENTER",
					VerticalAlignment:   "MIDDLE",
				},
			},
			Fields: "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	}
}

// CreateAlternatingRowsRequest creates alternating row colors
func CreateAlternatingRowsRequest(sheetID int64, endRow, endCol int64) *sheets.Request {
	return &sheets.Request{
		AddBanding: &sheets.AddBandingRequest{
			BandedRange: &sheets.BandedRange{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    0,
					EndRowIndex:      endRow,
					StartColumnIndex: 0,
					EndColumnIndex:   endCol,
				},
				RowProperties: &sheets.BandingProperties{
					HeaderColor:     ColorHeaderBg,
					FirstBandColor:  &sheets.Color{Red: 1, Green: 1, Blue: 1, Alpha: 1},
					SecondBandColor: ColorAltRow,
				},
			},
		},
	}
}

// CreateRiskLevelFormatRules creates conditional formatting for risk levels
func CreateRiskLevelFormatRules(sheetID int64, colIndex, startRow, endRow int64) []*sheets.Request {
	rules := make([]*sheets.Request, 0)

	riskLevels := []struct {
		value string
		color *sheets.Color
	}{
		{"critical", ColorCritical},
		{"warning", ColorWarning},
		{"healthy", ColorSuccess},
		{"High", ColorCritical},
		{"Medium", ColorWarning},
		{"Low", ColorSuccess},
	}

	for _, risk := range riskLevels {
		rules = append(rules, &sheets.Request{
			AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
				Rule: &sheets.ConditionalFormatRule{
					Ranges: []*sheets.GridRange{
						{
							SheetId:          sheetID,
							StartRowIndex:    startRow,
							EndRowIndex:      endRow,
							StartColumnIndex: colIndex,
							EndColumnIndex:   colIndex + 1,
						},
					},
					BooleanRule: &sheets.BooleanRule{
						Condition: &sheets.BooleanCondition{
							Type: "TEXT_CONTAINS",
							Values: []*sheets.ConditionValue{
								{UserEnteredValue: risk.value},
							},
						},
						Format: &sheets.CellFormat{
							BackgroundColor: risk.color,
							TextFormat: &sheets.TextFormat{
								Bold: true,
							},
						},
					},
				},
				Index: 0,
			},
		})
	}

	return rules
}

// CreateDORAClassFormatRules creates conditional formatting for DORA metric classifications
func CreateDORAClassFormatRules(sheetID int64, colIndex, startRow, endRow int64) []*sheets.Request {
	rules := make([]*sheets.Request, 0)

	classes := []struct {
		value string
		color *sheets.Color
	}{
		{"Elite", &sheets.Color{Red: 0.26, Green: 0.62, Blue: 0.28, Alpha: 1}},   // Green
		{"High", &sheets.Color{Red: 0.56, Green: 0.77, Blue: 0.49, Alpha: 1}},    // Light Green
		{"Medium", &sheets.Color{Red: 1.0, Green: 0.85, Blue: 0.0, Alpha: 1}},    // Yellow
		{"Low", &sheets.Color{Red: 1.0, Green: 0.6, Blue: 0.0, Alpha: 1}},        // Orange
	}

	for _, class := range classes {
		rules = append(rules, &sheets.Request{
			AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
				Rule: &sheets.ConditionalFormatRule{
					Ranges: []*sheets.GridRange{
						{
							SheetId:          sheetID,
							StartRowIndex:    startRow,
							EndRowIndex:      endRow,
							StartColumnIndex: colIndex,
							EndColumnIndex:   colIndex + 1,
						},
					},
					BooleanRule: &sheets.BooleanRule{
						Condition: &sheets.BooleanCondition{
							Type: "TEXT_EQ",
							Values: []*sheets.ConditionValue{
								{UserEnteredValue: class.value},
							},
						},
						Format: &sheets.CellFormat{
							BackgroundColor: class.color,
							TextFormat: &sheets.TextFormat{
								Bold: true,
							},
						},
					},
				},
				Index: 0,
			},
		})
	}

	return rules
}

// CreateMergeCellsRequest creates a request to merge cells
func CreateMergeCellsRequest(sheetID int64, startRow, endRow, startCol, endCol int64) *sheets.Request {
	return &sheets.Request{
		MergeCells: &sheets.MergeCellsRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			MergeType: "MERGE_ALL",
		},
	}
}

// CreateCellFormatRequest creates formatting for a specific cell range
func CreateCellFormatRequest(sheetID int64, startRow, endRow, startCol, endCol int64, format *sheets.CellFormat) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: format,
			},
			Fields: "userEnteredFormat",
		},
	}
}

// CreateAutoResizeRequest creates a request to auto-resize columns
func CreateAutoResizeRequest(sheetID int64, startCol, endCol int64) *sheets.Request {
	return &sheets.Request{
		AutoResizeDimensions: &sheets.AutoResizeDimensionsRequest{
			Dimensions: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "COLUMNS",
				StartIndex: startCol,
				EndIndex:   endCol,
			},
		},
	}
}

// CreateNumberCard creates formatting for a number card in the dashboard
func CreateNumberCard(sheetID int64, row, col int64, value interface{}, label string, bgColor *sheets.Color) []*sheets.Request {
	requests := make([]*sheets.Request, 0)

	// Merge cells for the card (2 rows, 2 columns)
	requests = append(requests, CreateMergeCellsRequest(sheetID, row, row+2, col, col+2))

	// Format the merged cell
	textColor := &sheets.Color{Red: 1, Green: 1, Blue: 1, Alpha: 1}
	if bgColor == nil {
		bgColor = ColorDashboardBg
	}

	requests = append(requests, &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    row,
				EndRowIndex:      row + 2,
				StartColumnIndex: col,
				EndColumnIndex:   col + 2,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: bgColor,
					TextFormat: &sheets.TextFormat{
						Bold:            true,
						FontSize:        24,
						ForegroundColor: textColor,
					},
					HorizontalAlignment: "CENTER",
					VerticalAlignment:   "MIDDLE",
				},
			},
			Fields: "userEnteredFormat",
		},
	})

	return requests
}

// CreateSectionHeader creates a formatted section header
func CreateSectionHeader(sheetID int64, row, startCol, endCol int64) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    row,
				EndRowIndex:      row + 1,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: &sheets.Color{Red: 0.9, Green: 0.9, Blue: 0.9, Alpha: 1},
					TextFormat: &sheets.TextFormat{
						Bold:     true,
						FontSize: 12,
					},
					HorizontalAlignment: "LEFT",
					VerticalAlignment:   "MIDDLE",
				},
			},
			Fields: "userEnteredFormat",
		},
	}
}
