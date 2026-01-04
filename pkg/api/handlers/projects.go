// Package handlers provides HTTP request handlers for the Zero API
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/crashappsec/zero/pkg/api/types"
	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/workflow/freshness"
)

// ProjectHandler handles project-related API requests
type ProjectHandler struct {
	zeroHome     string
	cfg          *config.Config
	freshnessMgr *freshness.Manager
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(zeroHome string, cfg *config.Config) *ProjectHandler {
	return &ProjectHandler{
		zeroHome:     zeroHome,
		cfg:          cfg,
		freshnessMgr: freshness.NewManager(zeroHome),
	}
}

// List returns all hydrated projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	owner := r.URL.Query().Get("owner")

	projects, err := h.listProjects()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list projects", err)
		return
	}

	// Filter by owner if specified
	if owner != "" {
		filtered := make([]*types.Project, 0)
		for _, p := range projects {
			if p.Owner == owner {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}

	writeJSON(w, http.StatusOK, types.ListResponse[*types.Project]{
		Data:  projects,
		Total: len(projects),
	})
}

// Get returns a single project by ID
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	// Handle URL-encoded slashes
	projectID = strings.ReplaceAll(projectID, "%2F", "/")

	project := h.loadProject(projectID)
	if project == nil {
		writeError(w, http.StatusNotFound, "project not found", nil)
		return
	}

	writeJSON(w, http.StatusOK, project)
}

// Delete removes a project
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")

	projectPath := filepath.Join(h.zeroHome, "repos", projectID)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		writeError(w, http.StatusNotFound, "project not found", nil)
		return
	}

	if err := os.RemoveAll(projectPath); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete project", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetFreshness returns freshness info for a project
func (h *ProjectHandler) GetFreshness(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	projectID = strings.ReplaceAll(projectID, "%2F", "/")

	result, err := h.freshnessMgr.Check(projectID)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found or no freshness data", err)
		return
	}

	info := &types.FreshnessInfo{
		Level:        result.Level,
		LevelString:  string(result.Level),
		AgeString:    result.AgeString,
		NeedsRefresh: result.NeedsRefresh,
	}

	writeJSON(w, http.StatusOK, info)
}

// listProjects finds all hydrated projects
func (h *ProjectHandler) listProjects() ([]*types.Project, error) {
	reposPath := filepath.Join(h.zeroHome, "repos")

	if _, err := os.Stat(reposPath); os.IsNotExist(err) {
		return []*types.Project{}, nil
	}

	var projects []*types.Project

	orgs, err := os.ReadDir(reposPath)
	if err != nil {
		return nil, fmt.Errorf("reading repos directory: %w", err)
	}

	for _, org := range orgs {
		if !org.IsDir() {
			continue
		}

		orgPath := filepath.Join(reposPath, org.Name())
		repos, err := os.ReadDir(orgPath)
		if err != nil {
			continue
		}

		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}

			projectID := fmt.Sprintf("%s/%s", org.Name(), repo.Name())
			project := h.loadProject(projectID)
			if project != nil {
				projects = append(projects, project)
			}
		}
	}

	// Sort by last scan time (most recent first)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastScan.After(projects[j].LastScan)
	})

	return projects, nil
}

// loadProject loads project information from disk
func (h *ProjectHandler) loadProject(projectID string) *types.Project {
	parts := strings.SplitN(projectID, "/", 2)
	if len(parts) != 2 {
		return nil
	}
	owner, name := parts[0], parts[1]

	basePath := filepath.Join(h.zeroHome, "repos", owner, name)
	repoPath := filepath.Join(basePath, "repo")
	analysisPath := filepath.Join(basePath, "analysis")

	// Check if repo exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil
	}

	project := &types.Project{
		ID:           projectID,
		Owner:        owner,
		Name:         name,
		RepoPath:     repoPath,
		AnalysisPath: analysisPath,
	}

	// Get file count
	project.FileCount = countFiles(repoPath)

	// Get disk size
	project.DiskSize = getDirSize(basePath)

	// Get available scans
	project.AvailableScans = h.getAvailableScans(analysisPath)

	// Load manifest for scan info
	manifestPath := filepath.Join(analysisPath, "manifest.json")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var manifest struct {
			Scan struct {
				CompletedAt string `json:"completed_at"`
			} `json:"scan"`
		}
		if err := json.Unmarshal(data, &manifest); err == nil {
			if t, err := time.Parse(time.RFC3339, manifest.Scan.CompletedAt); err == nil {
				project.LastScan = t
			}
		}
	}

	// Load freshness info
	if result, err := h.freshnessMgr.Check(projectID); err == nil {
		project.Freshness = &types.FreshnessInfo{
			Level:        result.Level,
			LevelString:  string(result.Level),
			AgeString:    result.AgeString,
			NeedsRefresh: result.NeedsRefresh,
		}
	}

	// Load summary stats
	project.Summary = h.loadSummary(analysisPath)

	return project
}

// getAvailableScans returns list of available analysis types
func (h *ProjectHandler) getAvailableScans(analysisPath string) []string {
	var scans []string
	if files, err := os.ReadDir(analysisPath); err == nil {
		for _, f := range files {
			name := f.Name()
			if strings.HasSuffix(name, ".json") && name != "manifest.json" && name != "languages.json" {
				scans = append(scans, strings.TrimSuffix(name, ".json"))
			}
		}
	}
	sort.Strings(scans)
	return scans
}

// loadSummary loads summary statistics from analysis files
func (h *ProjectHandler) loadSummary(analysisPath string) *types.ProjectSummary {
	summary := &types.ProjectSummary{}

	// v4.0: Load from code-packages scanner (contains both SBOM and package analysis)
	if data := readAnalysisFile(analysisPath, "code-packages"); data != nil {
		if summ, ok := data["summary"].(map[string]interface{}); ok {
			// Load vulnerability counts
			if vulnSummary, ok := summ["vulns"].(map[string]interface{}); ok {
				summary.Vulnerabilities = &types.SeverityCounts{
					Critical: getInt(vulnSummary, "critical"),
					High:     getInt(vulnSummary, "high"),
					Medium:   getInt(vulnSummary, "medium"),
					Low:      getInt(vulnSummary, "low"),
				}
			}
			// Load package count from SBOM generation
			if gen, ok := summ["generation"].(map[string]interface{}); ok {
				summary.Packages = getInt(gen, "total_components")
			}
		}
	}

	// Load secrets count from code-security scanner
	if data := readAnalysisFile(analysisPath, "code-security"); data != nil {
		if summ, ok := data["summary"].(map[string]interface{}); ok {
			if secrets, ok := summ["secrets"].(map[string]interface{}); ok {
				summary.Secrets = getInt(secrets, "total_findings")
			}
		}
	}

	// Load technology count from technology-identification scanner
	if data := readAnalysisFile(analysisPath, "technology-identification"); data != nil {
		if summ, ok := data["summary"].(map[string]interface{}); ok {
			if det, ok := summ["detection"].(map[string]interface{}); ok {
				summary.Technologies = getInt(det, "total_technologies")
			}
		}
	}

	return summary
}

// Helper functions

func countFiles(path string) int {
	count := 0
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && !strings.Contains(p, ".git") {
			count++
		}
		return nil
	})
	return count
}

func getDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func readAnalysisFile(analysisPath, name string) map[string]interface{} {
	path := filepath.Join(analysisPath, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string, err error) {
	resp := types.ErrorResponse{
		Error: message,
	}
	if err != nil {
		resp.Details = err.Error()
	}
	writeJSON(w, status, resp)
}
