package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/crashappsec/zero/pkg/api/types"
	"github.com/crashappsec/zero/pkg/core/config"
)

func TestSystemHandler_Health(t *testing.T) {
	handler := NewSystemHandler(&config.Config{})

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var health types.HealthResponse
	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if health.Status != "ok" {
		t.Errorf("Health status = %q, want %q", health.Status, "ok")
	}
	if health.Version == "" {
		t.Error("Health version is empty")
	}
	if health.Timestamp.IsZero() {
		t.Error("Health timestamp is zero")
	}
}

func TestSystemHandler_GetConfig(t *testing.T) {
	cfg := &config.Config{
		Settings: config.Settings{
			DefaultProfile:        "all-quick",
			ParallelRepos:        3,
			ParallelScanners:     2,
			ScannerTimeoutSeconds: 600,
		},
		Profiles: map[string]config.Profile{
			"all-quick":    {},
			"all-complete": {},
		},
	}
	handler := NewSystemHandler(cfg)

	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()

	handler.GetConfig(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetConfig() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["default_profile"] != "all-quick" {
		t.Errorf("default_profile = %v, want %v", result["default_profile"], "all-quick")
	}
	if result["parallel_repos"] != float64(3) {
		t.Errorf("parallel_repos = %v, want 3", result["parallel_repos"])
	}
}

func TestSystemHandler_ListProfiles(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"quick": {
				Description:   "Quick scan",
				EstimatedTime: "5m",
				Scanners:      []string{"code-security"},
			},
			"full": {
				Description:   "Full scan",
				EstimatedTime: "30m",
				Scanners:      []string{"code-security", "code-packages"},
			},
		},
	}
	handler := NewSystemHandler(cfg)

	req := httptest.NewRequest("GET", "/api/profiles", nil)
	w := httptest.NewRecorder()

	handler.ListProfiles(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ListProfiles() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result types.ListResponse[types.ProfileInfo]
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
	if len(result.Data) != 2 {
		t.Errorf("Data length = %d, want 2", len(result.Data))
	}

	// Check alphabetical ordering
	if result.Data[0].Name != "full" {
		t.Errorf("First profile = %q, want 'full'", result.Data[0].Name)
	}
}

func TestSystemHandler_ListScanners(t *testing.T) {
	handler := NewSystemHandler(&config.Config{})

	req := httptest.NewRequest("GET", "/api/scanners", nil)
	w := httptest.NewRecorder()

	handler.ListScanners(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ListScanners() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result types.ListResponse[types.ScannerInfo]
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Total != 7 {
		t.Errorf("Total = %d, want 7", result.Total)
	}

	// Verify v4.0 super scanners are present
	scannerNames := make(map[string]bool)
	for _, s := range result.Data {
		scannerNames[s.Name] = true
	}

	expected := []string{"code-packages", "code-security", "code-quality", "devops", "technology-identification", "code-ownership", "developer-experience"}
	for _, name := range expected {
		if !scannerNames[name] {
			t.Errorf("Scanner %q not found", name)
		}
	}
}

func TestSystemHandler_ListAgents(t *testing.T) {
	handler := NewSystemHandler(&config.Config{})

	req := httptest.NewRequest("GET", "/api/agents", nil)
	w := httptest.NewRecorder()

	handler.ListAgents(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ListAgents() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result types.ListResponse[types.AgentInfo]
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Total != 13 {
		t.Errorf("Total = %d, want 13", result.Total)
	}

	// Verify some key agents
	agentIDs := make(map[string]bool)
	for _, a := range result.Data {
		agentIDs[a.ID] = true
	}

	expectedAgents := []string{"zero", "cereal", "razor", "gill", "hal"}
	for _, id := range expectedAgents {
		if !agentIDs[id] {
			t.Errorf("Agent %q not found", id)
		}
	}
}

func TestProjectHandler_List_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewProjectHandler(tmpDir, &config.Config{})

	req := httptest.NewRequest("GET", "/api/repos", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("List() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result types.ListResponse[*types.Project]
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("Total = %d, want 0", result.Total)
	}
}

func TestProjectHandler_List_WithProjects(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a mock project structure
	owner := "test-org"
	repo := "test-repo"
	repoPath := filepath.Join(tmpDir, "repos", owner, repo, "repo")
	analysisPath := filepath.Join(tmpDir, "repos", owner, repo, "analysis")
	freshnessPath := filepath.Join(tmpDir, "repos", owner, repo, "freshness.json")

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo directory: %v", err)
	}
	if err := os.MkdirAll(analysisPath, 0755); err != nil {
		t.Fatalf("Failed to create analysis directory: %v", err)
	}

	// Create a dummy file in repo
	os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# Test"), 0644)

	// Create manifest
	manifest := map[string]interface{}{
		"scan": map[string]interface{}{
			"completed_at": time.Now().Format(time.RFC3339),
		},
	}
	manifestData, _ := json.Marshal(manifest)
	os.WriteFile(filepath.Join(analysisPath, "manifest.json"), manifestData, 0644)

	// Create freshness.json
	freshness := map[string]interface{}{
		"last_scan": time.Now().Format(time.RFC3339),
	}
	freshnessData, _ := json.Marshal(freshness)
	os.WriteFile(freshnessPath, freshnessData, 0644)

	handler := NewProjectHandler(tmpDir, &config.Config{})

	req := httptest.NewRequest("GET", "/api/repos", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("List() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result types.ListResponse[*types.Project]
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
	if len(result.Data) > 0 {
		project := result.Data[0]
		if project.ID != "test-org/test-repo" {
			t.Errorf("Project ID = %q, want %q", project.ID, "test-org/test-repo")
		}
		if project.Owner != "test-org" {
			t.Errorf("Project Owner = %q, want %q", project.Owner, "test-org")
		}
		if project.Name != "test-repo" {
			t.Errorf("Project Name = %q, want %q", project.Name, "test-repo")
		}
	}
}

func TestProjectHandler_List_FilterByOwner(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock projects for different owners
	for _, owner := range []string{"owner1", "owner2"} {
		repoPath := filepath.Join(tmpDir, "repos", owner, "repo", "repo")
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("Failed to create repo directory: %v", err)
		}
		os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# Test"), 0644)
	}

	handler := NewProjectHandler(tmpDir, &config.Config{})

	// Filter by owner1
	req := httptest.NewRequest("GET", "/api/repos?owner=owner1", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	var result types.ListResponse[*types.Project]
	json.Unmarshal(body, &result)

	if result.Total != 1 {
		t.Errorf("Filtered Total = %d, want 1", result.Total)
	}
	if len(result.Data) > 0 && result.Data[0].Owner != "owner1" {
		t.Errorf("Filtered owner = %q, want %q", result.Data[0].Owner, "owner1")
	}
}

func TestProjectHandler_Get(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a mock project
	repoPath := filepath.Join(tmpDir, "repos", "org", "repo", "repo")
	os.MkdirAll(repoPath, 0755)
	os.WriteFile(filepath.Join(repoPath, "main.go"), []byte("package main"), 0644)

	handler := NewProjectHandler(tmpDir, &config.Config{})

	// Setup chi router to extract URL params
	r := chi.NewRouter()
	r.Get("/api/repos/{projectID}", handler.Get)

	req := httptest.NewRequest("GET", "/api/repos/org%2Frepo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Get() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var project types.Project
	if err := json.Unmarshal(body, &project); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if project.ID != "org/repo" {
		t.Errorf("Project ID = %q, want %q", project.ID, "org/repo")
	}
}

func TestProjectHandler_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewProjectHandler(tmpDir, &config.Config{})

	r := chi.NewRouter()
	r.Get("/api/repos/{projectID}", handler.Get)

	req := httptest.NewRequest("GET", "/api/repos/nonexistent%2Frepo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Get() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestProjectHandler_Delete(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a mock project
	projectPath := filepath.Join(tmpDir, "repos", "org", "repo")
	os.MkdirAll(filepath.Join(projectPath, "repo"), 0755)
	os.WriteFile(filepath.Join(projectPath, "repo", "main.go"), []byte("package main"), 0644)

	handler := NewProjectHandler(tmpDir, &config.Config{})

	r := chi.NewRouter()
	r.Delete("/api/repos/{projectID}", handler.Delete)

	req := httptest.NewRequest("DELETE", "/api/repos/org%2Frepo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Delete() status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	// Verify project was deleted
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		t.Error("Project directory should have been deleted")
	}
}

func TestProjectHandler_Delete_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewProjectHandler(tmpDir, &config.Config{})

	r := chi.NewRouter()
	r.Delete("/api/repos/{projectID}", handler.Delete)

	req := httptest.NewRequest("DELETE", "/api/repos/nonexistent%2Frepo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Delete() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(w, http.StatusOK, data)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)
	if result["key"] != "value" {
		t.Errorf("result[key] = %q, want %q", result["key"], "value")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusBadRequest, "invalid request", nil)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	body, _ := io.ReadAll(resp.Body)
	var result types.ErrorResponse
	json.Unmarshal(body, &result)
	if result.Error != "invalid request" {
		t.Errorf("Error = %q, want %q", result.Error, "invalid request")
	}
}

func TestCountFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("2"), 0644)
	subDir := filepath.Join(tmpDir, "sub")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "file3.txt"), []byte("3"), 0644)

	// Create .git directory (should be excluded)
	gitDir := filepath.Join(tmpDir, ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "config"), []byte("git"), 0644)

	count := countFiles(tmpDir)
	if count != 3 {
		t.Errorf("countFiles() = %d, want 3", count)
	}
}

func TestGetDirSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with known sizes
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("12345"), 0644)     // 5 bytes
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("1234567890"), 0644) // 10 bytes

	size := getDirSize(tmpDir)
	if size != 15 {
		t.Errorf("getDirSize() = %d, want 15", size)
	}
}

func TestGetInt(t *testing.T) {
	m := map[string]interface{}{
		"number":  float64(42),
		"string":  "hello",
		"missing": nil,
	}

	if v := getInt(m, "number"); v != 42 {
		t.Errorf("getInt(number) = %d, want 42", v)
	}
	if v := getInt(m, "string"); v != 0 {
		t.Errorf("getInt(string) = %d, want 0", v)
	}
	if v := getInt(m, "notexist"); v != 0 {
		t.Errorf("getInt(notexist) = %d, want 0", v)
	}
}

func TestReadAnalysisFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid analysis file
	data := map[string]interface{}{"key": "value"}
	jsonData, _ := json.Marshal(data)
	os.WriteFile(filepath.Join(tmpDir, "test.json"), jsonData, 0644)

	result := readAnalysisFile(tmpDir, "test")
	if result == nil {
		t.Fatal("readAnalysisFile() returned nil")
	}
	if result["key"] != "value" {
		t.Errorf("result[key] = %v, want 'value'", result["key"])
	}

	// Test non-existent file
	result = readAnalysisFile(tmpDir, "nonexistent")
	if result != nil {
		t.Error("readAnalysisFile() should return nil for non-existent file")
	}

	// Test invalid JSON
	os.WriteFile(filepath.Join(tmpDir, "invalid.json"), []byte("not json"), 0644)
	result = readAnalysisFile(tmpDir, "invalid")
	if result != nil {
		t.Error("readAnalysisFile() should return nil for invalid JSON")
	}
}

// ===== Analysis Handler Tests =====

func TestAnalysisHandler_GetAnalysis(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock analysis file
	analysisPath := filepath.Join(tmpDir, "repos", "org", "repo", "analysis")
	os.MkdirAll(analysisPath, 0755)
	analysisData := map[string]interface{}{
		"scanner": "code-security",
		"findings": map[string]interface{}{
			"vulns": []interface{}{
				map[string]interface{}{"id": "CVE-2024-001", "severity": "high"},
			},
		},
	}
	data, _ := json.Marshal(analysisData)
	os.WriteFile(filepath.Join(analysisPath, "code-security.json"), data, 0644)

	handler := NewAnalysisHandler(tmpDir)

	r := chi.NewRouter()
	r.Get("/api/repos/{projectID}/analysis/{analysisType}", handler.GetAnalysis)

	req := httptest.NewRequest("GET", "/api/repos/org%2Frepo/analysis/code-security", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetAnalysis() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if result["scanner"] != "code-security" {
		t.Errorf("scanner = %v, want 'code-security'", result["scanner"])
	}
}

func TestAnalysisHandler_GetAnalysis_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewAnalysisHandler(tmpDir)

	r := chi.NewRouter()
	r.Get("/api/repos/{projectID}/analysis/{analysisType}", handler.GetAnalysis)

	req := httptest.NewRequest("GET", "/api/repos/org%2Frepo/analysis/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GetAnalysis() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAnalysisHandler_GetSummary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock analysis files
	analysisPath := filepath.Join(tmpDir, "repos", "org", "repo", "analysis")
	os.MkdirAll(analysisPath, 0755)

	// code-security with findings
	securityData := map[string]interface{}{
		"summary": map[string]interface{}{"status": "complete"},
		"findings": map[string]interface{}{
			"vulns": []interface{}{
				map[string]interface{}{"id": "CVE-2024-001", "severity": "high"},
				map[string]interface{}{"id": "CVE-2024-002", "severity": "critical"},
			},
			"secrets": []interface{}{
				map[string]interface{}{"type": "api_key", "severity": "high"},
			},
		},
	}
	data, _ := json.Marshal(securityData)
	os.WriteFile(filepath.Join(analysisPath, "code-security.json"), data, 0644)

	handler := NewAnalysisHandler(tmpDir)

	r := chi.NewRouter()
	r.Get("/api/analysis/{projectID}/summary", handler.GetSummary)

	req := httptest.NewRequest("GET", "/api/analysis/org%2Frepo/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetSummary() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["project"] != "org/repo" {
		t.Errorf("project = %v, want 'org/repo'", result["project"])
	}

	// Check totals
	if totals, ok := result["totals"].(map[string]interface{}); ok {
		if totals["critical"] != float64(1) {
			t.Errorf("totals.critical = %v, want 1", totals["critical"])
		}
		if totals["high"] != float64(2) {
			t.Errorf("totals.high = %v, want 2", totals["high"])
		}
	}
}

func TestAnalysisHandler_GetVulnerabilities(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock analysis files
	analysisPath := filepath.Join(tmpDir, "repos", "org", "repo", "analysis")
	os.MkdirAll(analysisPath, 0755)

	// code-packages with vulns
	pkgData := map[string]interface{}{
		"findings": map[string]interface{}{
			"vulns": []interface{}{
				map[string]interface{}{"id": "CVE-2024-001", "severity": "high"},
			},
		},
	}
	data, _ := json.Marshal(pkgData)
	os.WriteFile(filepath.Join(analysisPath, "code-packages.json"), data, 0644)

	// code-security with vulns
	secData := map[string]interface{}{
		"findings": map[string]interface{}{
			"vulns": []interface{}{
				map[string]interface{}{"id": "CVE-2024-002", "severity": "critical"},
			},
		},
	}
	data, _ = json.Marshal(secData)
	os.WriteFile(filepath.Join(analysisPath, "code-security.json"), data, 0644)

	handler := NewAnalysisHandler(tmpDir)

	r := chi.NewRouter()
	r.Get("/api/analysis/{projectID}/vulnerabilities", handler.GetVulnerabilities)

	// Test without filter
	req := httptest.NewRequest("GET", "/api/analysis/org%2Frepo/vulnerabilities", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetVulnerabilities() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["total"] != float64(2) {
		t.Errorf("total = %v, want 2", result["total"])
	}

	// Test with severity filter
	req = httptest.NewRequest("GET", "/api/analysis/org%2Frepo/vulnerabilities?severity=critical", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body, _ = io.ReadAll(w.Result().Body)
	json.Unmarshal(body, &result)

	if result["total"] != float64(1) {
		t.Errorf("filtered total = %v, want 1", result["total"])
	}
}

func TestAnalysisHandler_GetSecrets(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock analysis files
	analysisPath := filepath.Join(tmpDir, "repos", "org", "repo", "analysis")
	os.MkdirAll(analysisPath, 0755)

	secData := map[string]interface{}{
		"findings": map[string]interface{}{
			"secrets": []interface{}{
				map[string]interface{}{"type": "api_key", "file": "config.json"},
				map[string]interface{}{"type": "password", "file": ".env"},
			},
		},
	}
	data, _ := json.Marshal(secData)
	os.WriteFile(filepath.Join(analysisPath, "code-security.json"), data, 0644)

	handler := NewAnalysisHandler(tmpDir)

	r := chi.NewRouter()
	r.Get("/api/analysis/{projectID}/secrets", handler.GetSecrets)

	req := httptest.NewRequest("GET", "/api/analysis/org%2Frepo/secrets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetSecrets() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["total"] != float64(2) {
		t.Errorf("total = %v, want 2", result["total"])
	}
}

func TestAnalysisHandler_GetSecrets_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewAnalysisHandler(tmpDir)

	r := chi.NewRouter()
	r.Get("/api/analysis/{projectID}/secrets", handler.GetSecrets)

	req := httptest.NewRequest("GET", "/api/analysis/org%2Frepo/secrets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GetSecrets() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAnalysisHandler_GetDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock analysis files
	analysisPath := filepath.Join(tmpDir, "repos", "org", "repo", "analysis")
	os.MkdirAll(analysisPath, 0755)

	pkgData := map[string]interface{}{
		"findings": map[string]interface{}{
			"generation": map[string]interface{}{
				"components": []interface{}{
					map[string]interface{}{"name": "express", "version": "4.18.0"},
					map[string]interface{}{"name": "lodash", "version": "4.17.21"},
				},
			},
			"licenses": []interface{}{
				map[string]interface{}{"license": "MIT", "count": 10},
			},
		},
	}
	data, _ := json.Marshal(pkgData)
	os.WriteFile(filepath.Join(analysisPath, "code-packages.json"), data, 0644)

	handler := NewAnalysisHandler(tmpDir)

	r := chi.NewRouter()
	r.Get("/api/analysis/{projectID}/dependencies", handler.GetDependencies)

	req := httptest.NewRequest("GET", "/api/analysis/org%2Frepo/dependencies", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetDependencies() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["total"] != float64(2) {
		t.Errorf("total = %v, want 2", result["total"])
	}
	if result["licenses"] == nil {
		t.Error("licenses should be present")
	}
}

func TestAnalysisHandler_GetAggregateStats(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two mock projects
	for _, project := range []string{"org1/repo1", "org2/repo2"} {
		analysisPath := filepath.Join(tmpDir, "repos", project, "analysis")
		os.MkdirAll(analysisPath, 0755)

		pkgData := map[string]interface{}{
			"findings": map[string]interface{}{
				"vulns": []interface{}{
					map[string]interface{}{"id": "CVE-2024-001", "severity": "high"},
				},
				"generation": map[string]interface{}{
					"components": []interface{}{
						map[string]interface{}{"name": "pkg1"},
					},
				},
			},
		}
		data, _ := json.Marshal(pkgData)
		os.WriteFile(filepath.Join(analysisPath, "code-packages.json"), data, 0644)

		secData := map[string]interface{}{
			"findings": map[string]interface{}{
				"secrets": []interface{}{
					map[string]interface{}{"type": "api_key"},
				},
			},
		}
		data, _ = json.Marshal(secData)
		os.WriteFile(filepath.Join(analysisPath, "code-security.json"), data, 0644)
	}

	handler := NewAnalysisHandler(tmpDir)

	req := httptest.NewRequest("GET", "/api/analysis/stats", nil)
	w := httptest.NewRecorder()
	handler.GetAggregateStats(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetAggregateStats() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result AggregateStats
	json.Unmarshal(body, &result)

	if result.TotalProjects != 2 {
		t.Errorf("TotalProjects = %d, want 2", result.TotalProjects)
	}
	if result.TotalVulns != 2 {
		t.Errorf("TotalVulns = %d, want 2", result.TotalVulns)
	}
	if result.TotalSecrets != 2 {
		t.Errorf("TotalSecrets = %d, want 2", result.TotalSecrets)
	}
	if result.VulnsBySeverity["high"] != 2 {
		t.Errorf("VulnsBySeverity[high] = %d, want 2", result.VulnsBySeverity["high"])
	}
}

func TestAnalysisHandler_GetAggregateStats_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewAnalysisHandler(tmpDir)

	req := httptest.NewRequest("GET", "/api/analysis/stats", nil)
	w := httptest.NewRecorder()
	handler.GetAggregateStats(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetAggregateStats() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result AggregateStats
	json.Unmarshal(body, &result)

	if result.TotalProjects != 0 {
		t.Errorf("TotalProjects = %d, want 0", result.TotalProjects)
	}
}

func TestMatchSeverity(t *testing.T) {
	tests := []struct {
		vuln     map[string]interface{}
		severity string
		expected bool
	}{
		{map[string]interface{}{"severity": "high"}, "high", true},
		{map[string]interface{}{"severity": "HIGH"}, "high", true},
		{map[string]interface{}{"severity": "high"}, "critical", false},
		{map[string]interface{}{}, "high", false},
		{map[string]interface{}{"severity": 123}, "high", false},
	}

	for _, tt := range tests {
		got := matchSeverity(tt.vuln, tt.severity)
		if got != tt.expected {
			t.Errorf("matchSeverity(%v, %q) = %v, want %v", tt.vuln, tt.severity, got, tt.expected)
		}
	}
}

// ===== Config Handler Tests =====

func TestConfigHandler_GetSettings(t *testing.T) {
	cfg := &config.Config{
		Settings: config.Settings{
			DefaultProfile:        "all-quick",
			ParallelRepos:        3,
			ParallelScanners:     2,
			ScannerTimeoutSeconds: 600,
			CacheTTLHours:        24,
		},
	}
	handler := NewConfigHandler(cfg)

	req := httptest.NewRequest("GET", "/api/settings", nil)
	w := httptest.NewRecorder()
	handler.GetSettings(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetSettings() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["default_profile"] != "all-quick" {
		t.Errorf("default_profile = %v, want 'all-quick'", result["default_profile"])
	}
	if result["parallel_repos"] != float64(3) {
		t.Errorf("parallel_repos = %v, want 3", result["parallel_repos"])
	}
}

func TestConfigHandler_ListProfiles(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"quick": {
				Name:        "quick",
				Description: "Quick scan",
				Scanners:    []string{"code-security"},
			},
			"full": {
				Name:        "full",
				Description: "Full scan",
				Scanners:    []string{"code-security", "code-packages"},
			},
		},
	}
	handler := NewConfigHandler(cfg)

	req := httptest.NewRequest("GET", "/api/profiles", nil)
	w := httptest.NewRecorder()
	handler.ListProfiles(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ListProfiles() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result types.ListResponse[types.ProfileInfo]
	json.Unmarshal(body, &result)

	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
}

func TestConfigHandler_GetProfile(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"quick": {
				Name:        "quick",
				Description: "Quick scan",
				Scanners:    []string{"code-security"},
			},
		},
	}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Get("/api/profiles/{name}", handler.GetProfile)

	req := httptest.NewRequest("GET", "/api/profiles/quick", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetProfile() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result types.ProfileInfo
	json.Unmarshal(body, &result)

	if result.Name != "quick" {
		t.Errorf("Name = %q, want 'quick'", result.Name)
	}
}

func TestConfigHandler_GetProfile_NotFound(t *testing.T) {
	cfg := &config.Config{Profiles: map[string]config.Profile{}}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Get("/api/profiles/{name}", handler.GetProfile)

	req := httptest.NewRequest("GET", "/api/profiles/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GetProfile() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestConfigHandler_CreateProfile(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{},
	}
	handler := NewConfigHandler(cfg)

	body := `{"name": "custom", "description": "Custom profile", "scanners": ["code-security"]}`
	req := httptest.NewRequest("POST", "/api/profiles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.CreateProfile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("CreateProfile() status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	// Verify profile was created
	if _, ok := cfg.Profiles["custom"]; !ok {
		t.Error("Profile 'custom' should have been created")
	}
}

func TestConfigHandler_CreateProfile_MissingName(t *testing.T) {
	cfg := &config.Config{Profiles: map[string]config.Profile{}}
	handler := NewConfigHandler(cfg)

	body := `{"description": "No name", "scanners": ["code-security"]}`
	req := httptest.NewRequest("POST", "/api/profiles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.CreateProfile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateProfile() status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConfigHandler_CreateProfile_MissingScanners(t *testing.T) {
	cfg := &config.Config{Profiles: map[string]config.Profile{}}
	handler := NewConfigHandler(cfg)

	body := `{"name": "test"}`
	req := httptest.NewRequest("POST", "/api/profiles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.CreateProfile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateProfile() status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConfigHandler_DeleteProfile(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"custom": {Name: "custom", Scanners: []string{"code-security"}},
		},
		Settings: config.Settings{DefaultProfile: "all-quick"},
	}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Delete("/api/profiles/{name}", handler.DeleteProfile)

	req := httptest.NewRequest("DELETE", "/api/profiles/custom", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("DeleteProfile() status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestConfigHandler_DeleteProfile_NotFound(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{},
		Settings: config.Settings{DefaultProfile: "all-quick"},
	}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Delete("/api/profiles/{name}", handler.DeleteProfile)

	req := httptest.NewRequest("DELETE", "/api/profiles/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("DeleteProfile() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestConfigHandler_ExportConfig(t *testing.T) {
	cfg := &config.Config{
		Settings: config.Settings{DefaultProfile: "quick"},
		Profiles: map[string]config.Profile{"quick": {Name: "quick"}},
	}
	handler := NewConfigHandler(cfg)

	req := httptest.NewRequest("GET", "/api/config/export", nil)
	w := httptest.NewRecorder()
	handler.ExportConfig(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ExportConfig() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want 'application/json'", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); cd != "attachment; filename=zero.config.json" {
		t.Errorf("Content-Disposition = %q", cd)
	}
}

// ===== Scan Handler Tests =====

func TestIsValidTarget(t *testing.T) {
	tests := []struct {
		target   string
		expected bool
	}{
		// Valid targets
		{"owner/repo", true},
		{"expressjs/express", true},
		{"org-name/my-repo", true},
		{"user_name/repo_name", true},
		{"user.name/repo.name", true},
		{"my-org", true},
		{"123org/456repo", true},

		// Invalid targets
		{"", false},
		{"../etc/passwd", false},
		{"owner//repo", false},
		{"owner;rm -rf /", false},
		{"owner|cat /etc/passwd", false},
		{"owner&background", false},
		{"owner$HOME", false},
		{"owner`id`", false},
		{"owner\nrepo", false},
		{"owner\rrepo", false},
		{"owner\x00repo", false},
		{strings.Repeat("a", 200), false}, // too long
		{"-invalid", false}, // starts with dash
	}

	for _, tt := range tests {
		got := isValidTarget(tt.target)
		if got != tt.expected {
			t.Errorf("isValidTarget(%q) = %v, want %v", tt.target, got, tt.expected)
		}
	}
}

func TestScanHandler_Start_InvalidBody(t *testing.T) {
	handler := NewScanHandler(nil)

	req := httptest.NewRequest("POST", "/api/scans", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()
	handler.Start(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Start() status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestScanHandler_Start_MissingTarget(t *testing.T) {
	handler := NewScanHandler(nil)

	body := `{}`
	req := httptest.NewRequest("POST", "/api/scans", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Start(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Start() status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestScanHandler_Start_InvalidTarget(t *testing.T) {
	handler := NewScanHandler(nil)

	body := `{"target": "../etc/passwd"}`
	req := httptest.NewRequest("POST", "/api/scans", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Start(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Start() status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConfigHandler_UpdateProfile(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"quick": {
				Name:        "quick",
				Description: "Quick scan",
				Scanners:    []string{"code-security"},
			},
		},
	}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Put("/api/profiles/{name}", handler.UpdateProfile)

	body := `{"description": "Updated description", "scanners": ["code-security", "code-packages"]}`
	req := httptest.NewRequest("PUT", "/api/profiles/quick", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("UpdateProfile() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify profile was updated
	updated := cfg.Profiles["quick"]
	if updated.Description != "Updated description" {
		t.Errorf("Description = %q, want 'Updated description'", updated.Description)
	}
	if len(updated.Scanners) != 2 {
		t.Errorf("Scanners count = %d, want 2", len(updated.Scanners))
	}
}

func TestConfigHandler_UpdateProfile_NotFound(t *testing.T) {
	cfg := &config.Config{Profiles: map[string]config.Profile{}}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Put("/api/profiles/{name}", handler.UpdateProfile)

	body := `{"description": "Updated"}`
	req := httptest.NewRequest("PUT", "/api/profiles/nonexistent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("UpdateProfile() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestConfigHandler_GetScanner(t *testing.T) {
	cfg := &config.Config{
		Scanners: map[string]config.Scanner{
			"code-security": {
				Name:        "code-security",
				Description: "Security scanner",
				OutputFile:  "code-security.json",
			},
		},
	}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Get("/api/scanners/{name}", handler.GetScanner)

	req := httptest.NewRequest("GET", "/api/scanners/code-security", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetScanner() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["name"] != "code-security" {
		t.Errorf("name = %v, want 'code-security'", result["name"])
	}
}

func TestConfigHandler_GetScanner_NotFound(t *testing.T) {
	cfg := &config.Config{Scanners: map[string]config.Scanner{}}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Get("/api/scanners/{name}", handler.GetScanner)

	req := httptest.NewRequest("GET", "/api/scanners/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GetScanner() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestConfigHandler_UpdateScanner(t *testing.T) {
	cfg := &config.Config{
		Scanners: map[string]config.Scanner{
			"code-security": {
				Name:        "code-security",
				Description: "Security scanner",
			},
		},
	}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Put("/api/scanners/{name}", handler.UpdateScanner)

	body := `{"description": "Updated security scanner"}`
	req := httptest.NewRequest("PUT", "/api/scanners/code-security", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("UpdateScanner() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestConfigHandler_UpdateScanner_NotFound(t *testing.T) {
	cfg := &config.Config{Scanners: map[string]config.Scanner{}}
	handler := NewConfigHandler(cfg)

	r := chi.NewRouter()
	r.Put("/api/scanners/{name}", handler.UpdateScanner)

	body := `{"description": "Updated"}`
	req := httptest.NewRequest("PUT", "/api/scanners/nonexistent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("UpdateScanner() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestConfigHandler_ImportConfig(t *testing.T) {
	cfg := &config.Config{
		Settings: config.Settings{},
		Profiles: map[string]config.Profile{},
	}
	handler := NewConfigHandler(cfg)

	body := `{"settings": {"default_profile": "imported"}, "profiles": {"imported": {"name": "imported", "scanners": ["code-security"]}}}`
	req := httptest.NewRequest("POST", "/api/config/import", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ImportConfig(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ImportConfig() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestConfigHandler_ImportConfig_InvalidJSON(t *testing.T) {
	cfg := &config.Config{}
	handler := NewConfigHandler(cfg)

	body := `invalid json`
	req := httptest.NewRequest("POST", "/api/config/import", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ImportConfig(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ImportConfig() status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}
