package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// ReportStatus represents the status of a report generation
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusReady      ReportStatus = "ready"
	ReportStatusError      ReportStatus = "error"
)

// ReportInfo contains information about a generated report
type ReportInfo struct {
	ProjectID   string       `json:"project_id"`
	Status      ReportStatus `json:"status"`
	GeneratedAt *time.Time   `json:"generated_at,omitempty"`
	URL         string       `json:"url,omitempty"`
	Error       string       `json:"error,omitempty"`
}

// ReportHandler handles report generation requests
type ReportHandler struct {
	zeroHome string
	mu       sync.RWMutex
	jobs     map[string]*ReportInfo
}

// NewReportHandler creates a new report handler
func NewReportHandler(zeroHome string) *ReportHandler {
	return &ReportHandler{
		zeroHome: zeroHome,
		jobs:     make(map[string]*ReportInfo),
	}
}

// List returns all reports
func (h *ReportHandler) List(w http.ResponseWriter, r *http.Request) {
	reposPath := filepath.Join(h.zeroHome, "repos")
	entries, err := os.ReadDir(reposPath)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data":  []ReportInfo{},
			"total": 0,
		})
		return
	}

	var reports []ReportInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check for nested owner/repo structure
		ownerPath := filepath.Join(reposPath, entry.Name())
		subEntries, err := os.ReadDir(ownerPath)
		if err != nil {
			continue
		}

		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				continue
			}

			projectID := entry.Name() + "/" + subEntry.Name()
			report := h.getReportInfo(projectID)
			reports = append(reports, report)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  reports,
		"total": len(reports),
	})
}

// Get returns report info for a specific project
func (h *ReportHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")

	report := h.getReportInfo(projectID)
	writeJSON(w, http.StatusOK, report)
}

// Generate triggers report generation for a project
func (h *ReportHandler) Generate(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")

	// Check if project exists
	projectPath := filepath.Join(h.zeroHome, "repos", projectID)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		writeError(w, http.StatusNotFound, "project not found", err)
		return
	}

	// Check for force regenerate
	var body struct {
		Force bool `json:"force"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body)
	}

	// Check if already generating
	h.mu.RLock()
	existing := h.jobs[projectID]
	h.mu.RUnlock()

	if existing != nil && existing.Status == ReportStatusGenerating {
		writeJSON(w, http.StatusAccepted, existing)
		return
	}

	// Check if report exists and not forcing
	if !body.Force {
		reportPath := filepath.Join(projectPath, "report", "index.html")
		if info, err := os.Stat(reportPath); err == nil {
			generatedAt := info.ModTime()
			report := &ReportInfo{
				ProjectID:   projectID,
				Status:      ReportStatusReady,
				GeneratedAt: &generatedAt,
				URL:         fmt.Sprintf("/api/reports/%s/view", projectID),
			}
			writeJSON(w, http.StatusOK, report)
			return
		}
	}

	// Start generation
	report := &ReportInfo{
		ProjectID: projectID,
		Status:    ReportStatusGenerating,
	}

	h.mu.Lock()
	h.jobs[projectID] = report
	h.mu.Unlock()

	// Generate in background
	go h.generateReport(projectID)

	writeJSON(w, http.StatusAccepted, report)
}

// View serves the generated report
func (h *ReportHandler) View(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")

	// Serve static files from report directory
	reportPath := filepath.Join(h.zeroHome, "repos", projectID, "report")
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		writeError(w, http.StatusNotFound, "report not found", err)
		return
	}

	// Get the file path from the request
	filePath := chi.URLParam(r, "*")
	if filePath == "" {
		filePath = "index.html"
	}

	fullPath := filepath.Join(reportPath, filePath)

	// Security: ensure we're not escaping the report directory
	if !strings.HasPrefix(fullPath, reportPath) {
		writeError(w, http.StatusForbidden, "invalid path", nil)
		return
	}

	http.ServeFile(w, r, fullPath)
}

// getReportInfo returns report info for a project
func (h *ReportHandler) getReportInfo(projectID string) ReportInfo {
	// Check in-memory jobs first
	h.mu.RLock()
	if job, ok := h.jobs[projectID]; ok && job.Status == ReportStatusGenerating {
		h.mu.RUnlock()
		return *job
	}
	h.mu.RUnlock()

	// Check if report exists on disk
	reportPath := filepath.Join(h.zeroHome, "repos", projectID, "report", "index.html")
	if info, err := os.Stat(reportPath); err == nil {
		generatedAt := info.ModTime()
		return ReportInfo{
			ProjectID:   projectID,
			Status:      ReportStatusReady,
			GeneratedAt: &generatedAt,
			URL:         fmt.Sprintf("/api/reports/%s/view", projectID),
		}
	}

	return ReportInfo{
		ProjectID: projectID,
		Status:    ReportStatusPending,
	}
}

// generateReport runs the report generation process
func (h *ReportHandler) generateReport(projectID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	projectPath := filepath.Join(h.zeroHome, "repos", projectID)
	reportPath := filepath.Join(projectPath, "report")

	// For now, we'll create a simple placeholder report
	// In production, this would call the Evidence.dev report generator
	err := h.createPlaceholderReport(ctx, projectID, reportPath)

	h.mu.Lock()
	defer h.mu.Unlock()

	if err != nil {
		h.jobs[projectID] = &ReportInfo{
			ProjectID: projectID,
			Status:    ReportStatusError,
			Error:     err.Error(),
		}
		return
	}

	now := time.Now()
	h.jobs[projectID] = &ReportInfo{
		ProjectID:   projectID,
		Status:      ReportStatusReady,
		GeneratedAt: &now,
		URL:         fmt.Sprintf("/api/reports/%s/view", projectID),
	}
}

// createPlaceholderReport creates a simple HTML report
func (h *ReportHandler) createPlaceholderReport(ctx context.Context, projectID, reportPath string) error {
	if err := os.MkdirAll(reportPath, 0755); err != nil {
		return fmt.Errorf("creating report directory: %w", err)
	}

	// Load analysis data
	analysisPath := filepath.Join(h.zeroHome, "repos", projectID, "analysis")

	// Count findings from analysis files
	var vulnCount, secretCount, depCount int

	// Read packages for vulns
	if data, err := os.ReadFile(filepath.Join(analysisPath, "packages.json")); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(data, &pkg) == nil {
			if findings, ok := pkg["findings"].(map[string]interface{}); ok {
				if vulns, ok := findings["vulns"].([]interface{}); ok {
					vulnCount = len(vulns)
				}
			}
		}
	}

	// Read code-security for secrets
	if data, err := os.ReadFile(filepath.Join(analysisPath, "code-security.json")); err == nil {
		var sec map[string]interface{}
		if json.Unmarshal(data, &sec) == nil {
			if findings, ok := sec["findings"].(map[string]interface{}); ok {
				if secrets, ok := findings["secrets"].([]interface{}); ok {
					secretCount = len(secrets)
				}
			}
		}
	}

	// Read SBOM for dependencies
	if data, err := os.ReadFile(filepath.Join(analysisPath, "sbom.json")); err == nil {
		var sbom map[string]interface{}
		if json.Unmarshal(data, &sbom) == nil {
			if findings, ok := sbom["findings"].(map[string]interface{}); ok {
				if gen, ok := findings["generation"].(map[string]interface{}); ok {
					if comps, ok := gen["components"].([]interface{}); ok {
						depCount = len(comps)
					}
				}
			}
		}
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Zero Report - %s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #111827;
            color: #f3f4f6;
            min-height: 100vh;
            padding: 2rem;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        header {
            display: flex;
            align-items: center;
            gap: 1rem;
            margin-bottom: 2rem;
            padding-bottom: 1rem;
            border-bottom: 1px solid #374151;
        }
        .logo {
            width: 48px;
            height: 48px;
            background: #059669;
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: bold;
            font-size: 1.5rem;
        }
        h1 { font-size: 1.5rem; }
        .subtitle { color: #9ca3af; font-size: 0.875rem; }
        .cards {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        .card {
            background: #1f2937;
            border: 1px solid #374151;
            border-radius: 8px;
            padding: 1.5rem;
            text-align: center;
        }
        .card-value {
            font-size: 2.5rem;
            font-weight: bold;
            margin-bottom: 0.5rem;
        }
        .card-label { color: #9ca3af; font-size: 0.875rem; }
        .vuln { color: #ef4444; }
        .secrets { color: #f59e0b; }
        .deps { color: #3b82f6; }
        .section {
            background: #1f2937;
            border: 1px solid #374151;
            border-radius: 8px;
            padding: 1.5rem;
            margin-bottom: 1rem;
        }
        .section-title {
            font-size: 1.125rem;
            margin-bottom: 1rem;
            color: #10b981;
        }
        .info { color: #9ca3af; }
        .timestamp { margin-top: 2rem; text-align: center; color: #6b7280; font-size: 0.75rem; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="logo">Z</div>
            <div>
                <h1>%s</h1>
                <p class="subtitle">Security Analysis Report</p>
            </div>
        </header>

        <div class="cards">
            <div class="card">
                <div class="card-value vuln">%d</div>
                <div class="card-label">Vulnerabilities</div>
            </div>
            <div class="card">
                <div class="card-value secrets">%d</div>
                <div class="card-label">Secrets</div>
            </div>
            <div class="card">
                <div class="card-value deps">%d</div>
                <div class="card-label">Dependencies</div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">About This Report</h2>
            <p class="info">
                This is a placeholder report generated by Zero. For full interactive reports
                with charts and detailed analysis, run <code>zero report %s</code> from the CLI
                to generate an Evidence.dev report.
            </p>
        </div>

        <p class="timestamp">Generated: %s</p>
    </div>
</body>
</html>`, projectID, projectID, vulnCount, secretCount, depCount, projectID, time.Now().Format(time.RFC1123))

	return os.WriteFile(filepath.Join(reportPath, "index.html"), []byte(html), 0644)
}
