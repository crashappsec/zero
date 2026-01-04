package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// AnalysisHandler handles analysis data requests
type AnalysisHandler struct {
	zeroHome string
}

// NewAnalysisHandler creates a new analysis handler
func NewAnalysisHandler(zeroHome string) *AnalysisHandler {
	return &AnalysisHandler{
		zeroHome: zeroHome,
	}
}

// GetAnalysis returns raw analysis JSON for a specific type
func (h *AnalysisHandler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")
	analysisType := chi.URLParam(r, "analysisType")

	data, err := h.readAnalysis(projectID, analysisType)
	if err != nil {
		writeError(w, http.StatusNotFound, "analysis not found", err)
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// GetSummary returns aggregated summary for a project
func (h *AnalysisHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")

	analysisPath := filepath.Join(h.zeroHome, "repos", projectID, "analysis")

	summary := map[string]interface{}{
		"project": projectID,
	}

	// Aggregate from multiple scanners (v4.0 super scanners)
	scanners := []string{"supply-chain", "code-security", "code-quality", "devops", "tech-id", "code-ownership", "developer-experience"}
	for _, scanner := range scanners {
		if data, err := h.readAnalysis(projectID, scanner); err == nil {
			if summ, ok := data["summary"]; ok {
				summary[scanner] = summ
			}
		}
	}

	// Get available analyses
	var available []string
	if files, err := os.ReadDir(analysisPath); err == nil {
		for _, f := range files {
			name := f.Name()
			if strings.HasSuffix(name, ".json") && name != "manifest.json" {
				available = append(available, strings.TrimSuffix(name, ".json"))
			}
		}
	}
	summary["available_analyses"] = available

	writeJSON(w, http.StatusOK, summary)
}

// GetVulnerabilities returns combined vulnerabilities from all sources
func (h *AnalysisHandler) GetVulnerabilities(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")
	severity := r.URL.Query().Get("severity")

	var allVulns []interface{}

	// Package vulnerabilities (from supply-chain scanner)
	if data, err := h.readAnalysis(projectID, "supply-chain"); err == nil {
		if findings, ok := data["findings"].(map[string]interface{}); ok {
			if vulns, ok := findings["vulns"].([]interface{}); ok {
				for _, v := range vulns {
					if vm, ok := v.(map[string]interface{}); ok {
						vm["source"] = "package"
						if severity == "" || matchSeverity(vm, severity) {
							allVulns = append(allVulns, vm)
						}
					}
				}
			}
		}
	}

	// Code vulnerabilities
	if data, err := h.readAnalysis(projectID, "code-security"); err == nil {
		if findings, ok := data["findings"].(map[string]interface{}); ok {
			if vulns, ok := findings["vulns"].([]interface{}); ok {
				for _, v := range vulns {
					if vm, ok := v.(map[string]interface{}); ok {
						vm["source"] = "code"
						if severity == "" || matchSeverity(vm, severity) {
							allVulns = append(allVulns, vm)
						}
					}
				}
			}
		}
	}

	result := map[string]interface{}{
		"project":         projectID,
		"total":           len(allVulns),
		"vulnerabilities": allVulns,
	}

	writeJSON(w, http.StatusOK, result)
}

// GetSecrets returns detected secrets
func (h *AnalysisHandler) GetSecrets(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")

	data, err := h.readAnalysis(projectID, "code-security")
	if err != nil {
		writeError(w, http.StatusNotFound, "no secrets data found", err)
		return
	}

	var secrets []interface{}
	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if sec, ok := findings["secrets"].([]interface{}); ok {
			secrets = sec
		}
	}

	result := map[string]interface{}{
		"project": projectID,
		"total":   len(secrets),
		"secrets": secrets,
	}

	writeJSON(w, http.StatusOK, result)
}

// GetDependencies returns SBOM packages
func (h *AnalysisHandler) GetDependencies(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")

	// v4.0: supply-chain scanner contains both SBOM generation and package analysis
	data, err := h.readAnalysis(projectID, "supply-chain")
	if err != nil {
		writeError(w, http.StatusNotFound, "no supply-chain data found", err)
		return
	}

	var packages []interface{}
	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if gen, ok := findings["generation"].(map[string]interface{}); ok {
			if components, ok := gen["components"].([]interface{}); ok {
				packages = components
			}
		}
	}

	result := map[string]interface{}{
		"project":  projectID,
		"total":    len(packages),
		"packages": packages,
	}

	// Add license summary if available (now in same supply-chain scanner)
	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if licenses, ok := findings["licenses"].([]interface{}); ok {
			result["licenses"] = licenses
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// readAnalysis reads an analysis file for a project
func (h *AnalysisHandler) readAnalysis(projectID, analysisType string) (map[string]interface{}, error) {
	path := filepath.Join(h.zeroHome, "repos", projectID, "analysis", analysisType+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func matchSeverity(vuln map[string]interface{}, severity string) bool {
	if s, ok := vuln["severity"].(string); ok {
		return strings.EqualFold(s, severity)
	}
	return false
}

// AggregateStats represents overall stats across all projects
type AggregateStats struct {
	TotalProjects   int            `json:"total_projects"`
	TotalVulns      int            `json:"total_vulns"`
	TotalSecrets    int            `json:"total_secrets"`
	TotalDeps       int            `json:"total_deps"`
	VulnsBySeverity map[string]int `json:"vulns_by_severity"`
	ProjectStats    []ProjectStats `json:"project_stats"`
}

// ProjectStats represents stats for a single project
type ProjectStats struct {
	ID       string         `json:"id"`
	Vulns    int            `json:"vulns"`
	Secrets  int            `json:"secrets"`
	Deps     int            `json:"deps"`
	Severity map[string]int `json:"severity"`
}

// GetAggregateStats returns aggregate stats across all projects
func (h *AnalysisHandler) GetAggregateStats(w http.ResponseWriter, r *http.Request) {
	reposPath := filepath.Join(h.zeroHome, "repos")
	entries, err := os.ReadDir(reposPath)
	if err != nil {
		writeJSON(w, http.StatusOK, AggregateStats{
			VulnsBySeverity: map[string]int{},
			ProjectStats:    []ProjectStats{},
		})
		return
	}

	stats := AggregateStats{
		VulnsBySeverity: map[string]int{
			"critical": 0,
			"high":     0,
			"medium":   0,
			"low":      0,
		},
		ProjectStats: []ProjectStats{},
	}

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
			projStats := h.getProjectStats(projectID)

			stats.TotalProjects++
			stats.TotalVulns += projStats.Vulns
			stats.TotalSecrets += projStats.Secrets
			stats.TotalDeps += projStats.Deps

			for sev, count := range projStats.Severity {
				stats.VulnsBySeverity[sev] += count
			}

			stats.ProjectStats = append(stats.ProjectStats, projStats)
		}
	}

	writeJSON(w, http.StatusOK, stats)
}

// getProjectStats calculates stats for a single project
func (h *AnalysisHandler) getProjectStats(projectID string) ProjectStats {
	stats := ProjectStats{
		ID: projectID,
		Severity: map[string]int{
			"critical": 0,
			"high":     0,
			"medium":   0,
			"low":      0,
		},
	}

	// Count vulnerabilities (from supply-chain scanner)
	if data, err := h.readAnalysis(projectID, "supply-chain"); err == nil {
		if findings, ok := data["findings"].(map[string]interface{}); ok {
			if vulns, ok := findings["vulns"].([]interface{}); ok {
				stats.Vulns += len(vulns)
				for _, v := range vulns {
					if vm, ok := v.(map[string]interface{}); ok {
						if sev, ok := vm["severity"].(string); ok {
							stats.Severity[strings.ToLower(sev)]++
						}
					}
				}
			}
		}
	}

	// Count secrets
	if data, err := h.readAnalysis(projectID, "code-security"); err == nil {
		if findings, ok := data["findings"].(map[string]interface{}); ok {
			if secrets, ok := findings["secrets"].([]interface{}); ok {
				stats.Secrets = len(secrets)
			}
		}
	}

	// Count dependencies (from supply-chain scanner)
	if data, err := h.readAnalysis(projectID, "supply-chain"); err == nil {
		if findings, ok := data["findings"].(map[string]interface{}); ok {
			if gen, ok := findings["generation"].(map[string]interface{}); ok {
				if comps, ok := gen["components"].([]interface{}); ok {
					stats.Deps = len(comps)
				}
			}
		}
	}

	return stats
}
