// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package sheets

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Client wraps the Google Sheets API service
type Client struct {
	service *sheets.Service
}

// NewClient creates a new Sheets API client
func NewClient(ctx context.Context, httpClient *http.Client) (*Client, error) {
	service, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &Client{service: service}, nil
}

// CreateSpreadsheet creates a new spreadsheet with the given title
func (c *Client) CreateSpreadsheet(ctx context.Context, title string) (*sheets.Spreadsheet, error) {
	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}

	created, err := c.service.Spreadsheets.Create(spreadsheet).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create spreadsheet: %w", err)
	}

	return created, nil
}

// GetSpreadsheet retrieves an existing spreadsheet by ID
func (c *Client) GetSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	spreadsheet, err := c.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	return spreadsheet, nil
}

// BatchUpdate executes multiple update requests in a single API call
func (c *Client) BatchUpdate(ctx context.Context, spreadsheetID string, requests []*sheets.Request) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	resp, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("batch update failed: %w", err)
	}

	return resp, nil
}

// UpdateValues updates cell values in a range
func (c *Client) UpdateValues(ctx context.Context, spreadsheetID, rangeStr string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.service.Spreadsheets.Values.Update(spreadsheetID, rangeStr, valueRange).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to update values: %w", err)
	}

	return nil
}

// AppendValues appends values to a sheet
func (c *Client) AppendValues(ctx context.Context, spreadsheetID, rangeStr string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.service.Spreadsheets.Values.Append(spreadsheetID, rangeStr, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to append values: %w", err)
	}

	return nil
}

// AddSheet adds a new sheet/tab to the spreadsheet
func (c *Client) AddSheet(ctx context.Context, spreadsheetID, title string) (int64, error) {
	req := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: title,
			},
		},
	}

	resp, err := c.BatchUpdate(ctx, spreadsheetID, []*sheets.Request{req})
	if err != nil {
		return 0, err
	}

	if len(resp.Replies) > 0 && resp.Replies[0].AddSheet != nil {
		return resp.Replies[0].AddSheet.Properties.SheetId, nil
	}

	return 0, nil
}

// DeleteSheet deletes a sheet by ID
func (c *Client) DeleteSheet(ctx context.Context, spreadsheetID string, sheetID int64) error {
	req := &sheets.Request{
		DeleteSheet: &sheets.DeleteSheetRequest{
			SheetId: sheetID,
		},
	}

	_, err := c.BatchUpdate(ctx, spreadsheetID, []*sheets.Request{req})
	return err
}

// ClearSheet clears all data from a sheet
func (c *Client) ClearSheet(ctx context.Context, spreadsheetID, sheetName string) error {
	_, err := c.service.Spreadsheets.Values.Clear(
		spreadsheetID,
		sheetName,
		&sheets.ClearValuesRequest{},
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to clear sheet: %w", err)
	}

	return nil
}

// SetBasicFilter adds a basic filter to a sheet
func (c *Client) SetBasicFilter(ctx context.Context, spreadsheetID string, sheetID int64, endRow, endCol int64) error {
	req := &sheets.Request{
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
	}

	_, err := c.BatchUpdate(ctx, spreadsheetID, []*sheets.Request{req})
	return err
}

// FreezeRows freezes the first N rows
func (c *Client) FreezeRows(ctx context.Context, spreadsheetID string, sheetID int64, numRows int64) error {
	req := &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				SheetId: sheetID,
				GridProperties: &sheets.GridProperties{
					FrozenRowCount: numRows,
				},
			},
			Fields: "gridProperties.frozenRowCount",
		},
	}

	_, err := c.BatchUpdate(ctx, spreadsheetID, []*sheets.Request{req})
	return err
}

// ResizeColumn sets the width of a column
func (c *Client) ResizeColumn(ctx context.Context, spreadsheetID string, sheetID int64, colIndex, width int64) error {
	req := &sheets.Request{
		UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
			Range: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "COLUMNS",
				StartIndex: colIndex,
				EndIndex:   colIndex + 1,
			},
			Properties: &sheets.DimensionProperties{
				PixelSize: width,
			},
			Fields: "pixelSize",
		},
	}

	_, err := c.BatchUpdate(ctx, spreadsheetID, []*sheets.Request{req})
	return err
}

// GetSpreadsheetURL returns the URL to access the spreadsheet
func GetSpreadsheetURL(spreadsheetID string) string {
	return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit", spreadsheetID)
}

// RetryWithBackoff executes a function with exponential backoff on rate limit errors
func RetryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	backoff := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Check if it's a rate limit error (HTTP 429)
		// Google API errors typically contain "429" or "rateLimitExceeded"
		errStr := err.Error()
		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
				if backoff > 32*time.Second {
					backoff = 32 * time.Second
				}
				continue
			}
		}

		// Return error only if we hit a real error after all retries
		if errStr != "" {
			return err
		}
	}

	return nil
}
