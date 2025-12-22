// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package sheets

import (
	"google.golang.org/api/sheets/v4"
)

// ChartPosition defines where to place a chart
type ChartPosition struct {
	SheetID  int64
	Row      int64
	Col      int64
	OffsetX  int64
	OffsetY  int64
	Width    int64
	Height   int64
}

// CreatePieChartRequest creates a pie chart for severity distribution
func CreatePieChartRequest(pos ChartPosition, title string, dataSheetID int64, dataRange *sheets.GridRange) *sheets.Request {
	return &sheets.Request{
		AddChart: &sheets.AddChartRequest{
			Chart: &sheets.EmbeddedChart{
				Position: &sheets.EmbeddedObjectPosition{
					OverlayPosition: &sheets.OverlayPosition{
						AnchorCell: &sheets.GridCoordinate{
							SheetId:     pos.SheetID,
							RowIndex:    pos.Row,
							ColumnIndex: pos.Col,
						},
						OffsetXPixels: pos.OffsetX,
						OffsetYPixels: pos.OffsetY,
						WidthPixels:   pos.Width,
						HeightPixels:  pos.Height,
					},
				},
				Spec: &sheets.ChartSpec{
					Title: title,
					PieChart: &sheets.PieChartSpec{
						LegendPosition: "RIGHT_LEGEND",
						Domain: &sheets.ChartData{
							SourceRange: &sheets.ChartSourceRange{
								Sources: []*sheets.GridRange{
									{
										SheetId:          dataSheetID,
										StartRowIndex:    dataRange.StartRowIndex,
										EndRowIndex:      dataRange.EndRowIndex,
										StartColumnIndex: dataRange.StartColumnIndex,
										EndColumnIndex:   dataRange.StartColumnIndex + 1,
									},
								},
							},
						},
						Series: &sheets.ChartData{
							SourceRange: &sheets.ChartSourceRange{
								Sources: []*sheets.GridRange{
									{
										SheetId:          dataSheetID,
										StartRowIndex:    dataRange.StartRowIndex,
										EndRowIndex:      dataRange.EndRowIndex,
										StartColumnIndex: dataRange.StartColumnIndex + 1,
										EndColumnIndex:   dataRange.StartColumnIndex + 2,
									},
								},
							},
						},
						PieHole: 0.4, // Donut chart
					},
				},
			},
		},
	}
}

// CreateBarChartRequest creates a horizontal bar chart (e.g., for top repos)
func CreateBarChartRequest(pos ChartPosition, title string, dataSheetID int64, dataRange *sheets.GridRange) *sheets.Request {
	return &sheets.Request{
		AddChart: &sheets.AddChartRequest{
			Chart: &sheets.EmbeddedChart{
				Position: &sheets.EmbeddedObjectPosition{
					OverlayPosition: &sheets.OverlayPosition{
						AnchorCell: &sheets.GridCoordinate{
							SheetId:     pos.SheetID,
							RowIndex:    pos.Row,
							ColumnIndex: pos.Col,
						},
						OffsetXPixels: pos.OffsetX,
						OffsetYPixels: pos.OffsetY,
						WidthPixels:   pos.Width,
						HeightPixels:  pos.Height,
					},
				},
				Spec: &sheets.ChartSpec{
					Title: title,
					BasicChart: &sheets.BasicChartSpec{
						ChartType:      "BAR",
						LegendPosition: "NO_LEGEND",
						Axis: []*sheets.BasicChartAxis{
							{
								Position: "BOTTOM_AXIS",
								Title:    "Count",
							},
							{
								Position: "LEFT_AXIS",
								Title:    "",
							},
						},
						Domains: []*sheets.BasicChartDomain{
							{
								Domain: &sheets.ChartData{
									SourceRange: &sheets.ChartSourceRange{
										Sources: []*sheets.GridRange{
											{
												SheetId:          dataSheetID,
												StartRowIndex:    dataRange.StartRowIndex,
												EndRowIndex:      dataRange.EndRowIndex,
												StartColumnIndex: dataRange.StartColumnIndex,
												EndColumnIndex:   dataRange.StartColumnIndex + 1,
											},
										},
									},
								},
							},
						},
						Series: []*sheets.BasicChartSeries{
							{
								Series: &sheets.ChartData{
									SourceRange: &sheets.ChartSourceRange{
										Sources: []*sheets.GridRange{
											{
												SheetId:          dataSheetID,
												StartRowIndex:    dataRange.StartRowIndex,
												EndRowIndex:      dataRange.EndRowIndex,
												StartColumnIndex: dataRange.StartColumnIndex + 1,
												EndColumnIndex:   dataRange.StartColumnIndex + 2,
											},
										},
									},
								},
								Color: &sheets.Color{Red: 0.26, Green: 0.52, Blue: 0.96, Alpha: 1}, // Blue
							},
						},
						HeaderCount: 1,
					},
				},
			},
		},
	}
}

// CreateColumnChartRequest creates a vertical column chart
func CreateColumnChartRequest(pos ChartPosition, title string, dataSheetID int64, dataRange *sheets.GridRange, seriesCount int) *sheets.Request {
	series := make([]*sheets.BasicChartSeries, 0)
	colors := []*sheets.Color{
		{Red: 0.92, Green: 0.26, Blue: 0.21, Alpha: 1}, // Red (Critical)
		{Red: 1.0, Green: 0.6, Blue: 0.0, Alpha: 1},    // Orange (High)
		{Red: 1.0, Green: 0.85, Blue: 0.0, Alpha: 1},   // Yellow (Medium)
		{Red: 0.56, Green: 0.77, Blue: 0.49, Alpha: 1}, // Green (Low)
	}

	for i := 0; i < seriesCount && i < len(colors); i++ {
		series = append(series, &sheets.BasicChartSeries{
			Series: &sheets.ChartData{
				SourceRange: &sheets.ChartSourceRange{
					Sources: []*sheets.GridRange{
						{
							SheetId:          dataSheetID,
							StartRowIndex:    dataRange.StartRowIndex,
							EndRowIndex:      dataRange.EndRowIndex,
							StartColumnIndex: dataRange.StartColumnIndex + int64(i) + 1,
							EndColumnIndex:   dataRange.StartColumnIndex + int64(i) + 2,
						},
					},
				},
			},
			Color: colors[i],
		})
	}

	return &sheets.Request{
		AddChart: &sheets.AddChartRequest{
			Chart: &sheets.EmbeddedChart{
				Position: &sheets.EmbeddedObjectPosition{
					OverlayPosition: &sheets.OverlayPosition{
						AnchorCell: &sheets.GridCoordinate{
							SheetId:     pos.SheetID,
							RowIndex:    pos.Row,
							ColumnIndex: pos.Col,
						},
						OffsetXPixels: pos.OffsetX,
						OffsetYPixels: pos.OffsetY,
						WidthPixels:   pos.Width,
						HeightPixels:  pos.Height,
					},
				},
				Spec: &sheets.ChartSpec{
					Title: title,
					BasicChart: &sheets.BasicChartSpec{
						ChartType:      "COLUMN",
						LegendPosition: "BOTTOM_LEGEND",
						StackedType:    "STACKED",
						Axis: []*sheets.BasicChartAxis{
							{
								Position: "BOTTOM_AXIS",
								Title:    "",
							},
							{
								Position: "LEFT_AXIS",
								Title:    "Count",
							},
						},
						Domains: []*sheets.BasicChartDomain{
							{
								Domain: &sheets.ChartData{
									SourceRange: &sheets.ChartSourceRange{
										Sources: []*sheets.GridRange{
											{
												SheetId:          dataSheetID,
												StartRowIndex:    dataRange.StartRowIndex,
												EndRowIndex:      dataRange.EndRowIndex,
												StartColumnIndex: dataRange.StartColumnIndex,
												EndColumnIndex:   dataRange.StartColumnIndex + 1,
											},
										},
									},
								},
							},
						},
						Series:      series,
						HeaderCount: 1,
					},
				},
			},
		},
	}
}

// CreateLineChartRequest creates a line chart (e.g., for trends)
func CreateLineChartRequest(pos ChartPosition, title string, dataSheetID int64, dataRange *sheets.GridRange) *sheets.Request {
	return &sheets.Request{
		AddChart: &sheets.AddChartRequest{
			Chart: &sheets.EmbeddedChart{
				Position: &sheets.EmbeddedObjectPosition{
					OverlayPosition: &sheets.OverlayPosition{
						AnchorCell: &sheets.GridCoordinate{
							SheetId:     pos.SheetID,
							RowIndex:    pos.Row,
							ColumnIndex: pos.Col,
						},
						OffsetXPixels: pos.OffsetX,
						OffsetYPixels: pos.OffsetY,
						WidthPixels:   pos.Width,
						HeightPixels:  pos.Height,
					},
				},
				Spec: &sheets.ChartSpec{
					Title: title,
					BasicChart: &sheets.BasicChartSpec{
						ChartType:      "LINE",
						LegendPosition: "BOTTOM_LEGEND",
						Axis: []*sheets.BasicChartAxis{
							{
								Position: "BOTTOM_AXIS",
								Title:    "Date",
							},
							{
								Position: "LEFT_AXIS",
								Title:    "Findings",
							},
						},
						Domains: []*sheets.BasicChartDomain{
							{
								Domain: &sheets.ChartData{
									SourceRange: &sheets.ChartSourceRange{
										Sources: []*sheets.GridRange{
											{
												SheetId:          dataSheetID,
												StartRowIndex:    dataRange.StartRowIndex,
												EndRowIndex:      dataRange.EndRowIndex,
												StartColumnIndex: dataRange.StartColumnIndex,
												EndColumnIndex:   dataRange.StartColumnIndex + 1,
											},
										},
									},
								},
							},
						},
						Series: []*sheets.BasicChartSeries{
							{
								Series: &sheets.ChartData{
									SourceRange: &sheets.ChartSourceRange{
										Sources: []*sheets.GridRange{
											{
												SheetId:          dataSheetID,
												StartRowIndex:    dataRange.StartRowIndex,
												EndRowIndex:      dataRange.EndRowIndex,
												StartColumnIndex: dataRange.StartColumnIndex + 1,
												EndColumnIndex:   dataRange.StartColumnIndex + 2,
											},
										},
									},
								},
								Color: &sheets.Color{Red: 0.26, Green: 0.52, Blue: 0.96, Alpha: 1}, // Blue
							},
						},
						HeaderCount: 1,
					},
				},
			},
		},
	}
}

// CreateScorecard creates a scorecard chart for a single metric
func CreateScorecard(pos ChartPosition, keyValueSheetID int64, keyValueRange *sheets.GridRange) *sheets.Request {
	return &sheets.Request{
		AddChart: &sheets.AddChartRequest{
			Chart: &sheets.EmbeddedChart{
				Position: &sheets.EmbeddedObjectPosition{
					OverlayPosition: &sheets.OverlayPosition{
						AnchorCell: &sheets.GridCoordinate{
							SheetId:     pos.SheetID,
							RowIndex:    pos.Row,
							ColumnIndex: pos.Col,
						},
						OffsetXPixels: pos.OffsetX,
						OffsetYPixels: pos.OffsetY,
						WidthPixels:   pos.Width,
						HeightPixels:  pos.Height,
					},
				},
				Spec: &sheets.ChartSpec{
					ScorecardChart: &sheets.ScorecardChartSpec{
						KeyValueData: &sheets.ChartData{
							SourceRange: &sheets.ChartSourceRange{
								Sources: []*sheets.GridRange{keyValueRange},
							},
						},
						NumberFormatSource: "FROM_DATA",
					},
				},
			},
		},
	}
}

// SeverityChartData prepares data for severity pie chart
type SeverityChartData struct {
	Critical int
	High     int
	Medium   int
	Low      int
	Info     int
}

// ToRows converts severity data to spreadsheet rows
func (s *SeverityChartData) ToRows() [][]interface{} {
	return [][]interface{}{
		{"Severity", "Count"},
		{"Critical", s.Critical},
		{"High", s.High},
		{"Medium", s.Medium},
		{"Low", s.Low},
	}
}

// TopReposChartData prepares data for top repos bar chart
type TopReposChartData struct {
	Repos []struct {
		Name     string
		Critical int
	}
}

// ToRows converts top repos data to spreadsheet rows
func (t *TopReposChartData) ToRows() [][]interface{} {
	rows := [][]interface{}{
		{"Repository", "Critical Findings"},
	}
	for _, r := range t.Repos {
		rows = append(rows, []interface{}{r.Name, r.Critical})
	}
	return rows
}
