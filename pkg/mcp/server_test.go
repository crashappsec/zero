package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// setupTestZeroHome creates a temporary .zero directory structure for testing
func setupTestZeroHome(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zero-mcp-test")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}

	// Create repos directory structure
	reposDir := filepath.Join(tmpDir, "repos")
	os.MkdirAll(reposDir, 0755)

	// Create test project: owner1/repo1
	owner1Repo1 := filepath.Join(reposDir, "owner1", "repo1", "analysis")
	os.MkdirAll(owner1Repo1, 0755)

	// Write test analysis files
	codePackagesJSON := `{
		"summary": {
			"sbom": {"total_components": 150},
			"vulns": {"critical": 2, "high": 5}
		},
		"findings": {
			"vulns": [
				{"id": "CVE-2023-001", "package": "lodash", "version": "4.17.0", "severity": "critical", "title": "Prototype Pollution"},
				{"id": "CVE-2023-002", "package": "axios", "version": "0.21.0", "severity": "high", "title": "SSRF"}
			],
			"health": [
				{"package": "lodash", "score": 85, "status": "healthy"}
			],
			"licenses": [
				{"package": "lodash", "license": "MIT"}
			],
			"malcontent": [
				{"package": "suspicious-pkg", "risk": "high", "behavior": "network_call"}
			]
		}
	}`
	os.WriteFile(filepath.Join(owner1Repo1, "code-packages.json"), []byte(codePackagesJSON), 0644)

	codeSecurityJSON := `{
		"summary": {
			"secrets": {"total": 3},
			"vulns": {"total": 5}
		},
		"findings": {
			"secrets": [
				{"file": "config.js", "line": 10, "type": "api_key", "severity": "critical"}
			],
			"ciphers": [
				{"file": "crypto.js", "cipher": "DES", "severity": "high"}
			],
			"keys": [
				{"file": "keys.pem", "type": "RSA", "bits": 1024, "severity": "medium"}
			],
			"tls": [
				{"file": "server.js", "issue": "TLS 1.0", "severity": "high"}
			],
			"random": [
				{"file": "auth.js", "function": "Math.random", "severity": "medium"}
			],
			"certificates": []
		}
	}`
	os.WriteFile(filepath.Join(owner1Repo1, "code-security.json"), []byte(codeSecurityJSON), 0644)

	techJSON := `{
		"summary": {
			"total_technologies": 25
		},
		"findings": {
			"detection": [
				{"name": "React", "category": "frontend"},
				{"name": "Node.js", "category": "runtime"}
			]
		}
	}`
	os.WriteFile(filepath.Join(owner1Repo1, "technology-identification.json"), []byte(techJSON), 0644)

	// Create test project: owner1/repo2
	owner1Repo2 := filepath.Join(reposDir, "owner1", "repo2", "analysis")
	os.MkdirAll(owner1Repo2, 0755)
	os.WriteFile(filepath.Join(owner1Repo2, "code-packages.json"), []byte(`{"summary": {}}`), 0644)

	// Create test project: owner2/repo1
	owner2Repo1 := filepath.Join(reposDir, "owner2", "repo1", "analysis")
	os.MkdirAll(owner2Repo1, 0755)
	os.WriteFile(filepath.Join(owner2Repo1, "devops.json"), []byte(`{"findings": {}}`), 0644)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestNewServer(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.zeroHome != zeroHome {
		t.Errorf("zeroHome = %q, want %q", s.zeroHome, zeroHome)
	}
	if s.server == nil {
		t.Error("server is nil")
	}
}

func TestNewServer_DefaultPath(t *testing.T) {
	s := NewServer("")
	if s == nil {
		t.Fatal("NewServer returned nil")
	}

	expectedHome := filepath.Join(os.Getenv("HOME"), ".zero")
	if s.zeroHome != expectedHome {
		t.Errorf("zeroHome = %q, want %q", s.zeroHome, expectedHome)
	}
}

func TestServer_GetProjects(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)

	projects, err := s.getProjects()
	if err != nil {
		t.Fatalf("getProjects failed: %v", err)
	}

	if len(projects) != 3 {
		t.Errorf("got %d projects, want 3", len(projects))
	}

	// Check for expected projects
	projectIDs := make(map[string]bool)
	for _, p := range projects {
		projectIDs[p.ID] = true
	}

	expectedIDs := []string{"owner1/repo1", "owner1/repo2", "owner2/repo1"}
	for _, id := range expectedIDs {
		if !projectIDs[id] {
			t.Errorf("project %q not found", id)
		}
	}
}

func TestServer_GetProjects_CheckFields(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	projects, _ := s.getProjects()

	// Find owner1/repo1
	var repo1 *Project
	for _, p := range projects {
		if p.ID == "owner1/repo1" {
			repo1 = &p
			break
		}
	}

	if repo1 == nil {
		t.Fatal("owner1/repo1 not found")
	}

	if repo1.Owner != "owner1" {
		t.Errorf("Owner = %q, want owner1", repo1.Owner)
	}
	if repo1.Repo != "repo1" {
		t.Errorf("Repo = %q, want repo1", repo1.Repo)
	}

	// Check available scans (should include all 3 JSON files we created)
	expectedScans := []string{"code-packages", "code-security", "technology-identification"}
	for _, scan := range expectedScans {
		if !contains(repo1.AvailableScans, scan) {
			t.Errorf("missing expected scan %q", scan)
		}
	}
}

func TestServer_GetProjects_EmptyRepos(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zero-mcp-empty-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create empty repos directory
	os.MkdirAll(filepath.Join(tmpDir, "repos"), 0755)

	s := NewServer(tmpDir)
	projects, err := s.getProjects()
	if err != nil {
		t.Fatalf("getProjects failed: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("got %d projects, want 0", len(projects))
	}
}

func TestServer_GetProjects_NoReposDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zero-mcp-norepos-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewServer(tmpDir)
	_, err = s.getProjects()
	if err == nil {
		t.Error("expected error for missing repos directory")
	}
}

func TestServer_ReadAnalysis(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)

	t.Run("valid analysis", func(t *testing.T) {
		data, err := s.readAnalysis("owner1/repo1", "code-packages")
		if err != nil {
			t.Fatalf("readAnalysis failed: %v", err)
		}

		if data == nil {
			t.Fatal("data is nil")
		}

		// Check structure
		if _, ok := data["summary"]; !ok {
			t.Error("missing summary field")
		}
		if _, ok := data["findings"]; !ok {
			t.Error("missing findings field")
		}
	})

	t.Run("missing project", func(t *testing.T) {
		_, err := s.readAnalysis("nonexistent/repo", "code-packages")
		if err == nil {
			t.Error("expected error for missing project")
		}
	})

	t.Run("missing analysis type", func(t *testing.T) {
		_, err := s.readAnalysis("owner1/repo1", "nonexistent-analysis")
		if err == nil {
			t.Error("expected error for missing analysis type")
		}
	})
}

func TestServer_ReadAnalysis_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zero-mcp-invalid-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid JSON file
	analysisDir := filepath.Join(tmpDir, "repos", "owner", "repo", "analysis")
	os.MkdirAll(analysisDir, 0755)
	os.WriteFile(filepath.Join(analysisDir, "bad.json"), []byte("invalid json {"), 0644)

	s := NewServer(tmpDir)
	_, err = s.readAnalysis("owner/repo", "bad")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{nil, "a", false},
		{[]string{"code-packages", "code-security"}, "code-packages", true},
		{[]string{"code-packages", "code-security"}, "devops", false},
	}

	for _, tt := range tests {
		result := contains(tt.slice, tt.item)
		if result != tt.expected {
			t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, result, tt.expected)
		}
	}
}

func TestServer_HandleListProjects(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	t.Run("list all", func(t *testing.T) {
		_, output, err := s.handleListProjects(ctx, nil, ListProjectsInput{})
		if err != nil {
			t.Fatalf("handleListProjects failed: %v", err)
		}

		if output.Text == "" {
			t.Error("output is empty")
		}

		// Should contain all projects
		if !containsString(output.Text, "owner1/repo1") {
			t.Error("output missing owner1/repo1")
		}
		if !containsString(output.Text, "owner2/repo1") {
			t.Error("output missing owner2/repo1")
		}
	})

	t.Run("filter by owner", func(t *testing.T) {
		_, output, err := s.handleListProjects(ctx, nil, ListProjectsInput{Owner: "owner1"})
		if err != nil {
			t.Fatalf("handleListProjects failed: %v", err)
		}

		if !containsString(output.Text, "owner1/repo1") {
			t.Error("output missing owner1/repo1")
		}
		if containsString(output.Text, "owner2/repo1") {
			t.Error("output should not contain owner2/repo1")
		}
	})
}

func TestServer_HandleGetProjectSummary(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	t.Run("existing project", func(t *testing.T) {
		_, output, err := s.handleGetProjectSummary(ctx, nil, ProjectInput{Project: "owner1/repo1"})
		if err != nil {
			t.Fatalf("handleGetProjectSummary failed: %v", err)
		}

		if !containsString(output.Text, "owner1/repo1") {
			t.Error("output missing project ID")
		}
		// Should have packages summary since code-packages.json exists
		if !containsString(output.Text, "packages") {
			t.Error("output missing packages summary")
		}
	})

	t.Run("non-existent project", func(t *testing.T) {
		_, _, err := s.handleGetProjectSummary(ctx, nil, ProjectInput{Project: "nonexistent/repo"})
		if err == nil {
			t.Error("expected error for non-existent project")
		}
	})
}

func TestServer_HandleGetVulnerabilities(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	t.Run("all vulnerabilities", func(t *testing.T) {
		_, output, err := s.handleGetVulnerabilities(ctx, nil, VulnerabilitiesInput{Project: "owner1/repo1"})
		if err != nil {
			t.Fatalf("handleGetVulnerabilities failed: %v", err)
		}

		if !containsString(output.Text, "CVE-2023-001") {
			t.Error("output missing CVE-2023-001")
		}
		if !containsString(output.Text, "CVE-2023-002") {
			t.Error("output missing CVE-2023-002")
		}
	})

	t.Run("filter by severity", func(t *testing.T) {
		_, output, err := s.handleGetVulnerabilities(ctx, nil, VulnerabilitiesInput{
			Project:  "owner1/repo1",
			Severity: "critical",
		})
		if err != nil {
			t.Fatalf("handleGetVulnerabilities failed: %v", err)
		}

		if !containsString(output.Text, "CVE-2023-001") {
			t.Error("output missing critical CVE-2023-001")
		}
		// count should be 1 for critical only
		if !containsString(output.Text, `"count": 1`) {
			t.Error("count should be 1 for critical severity")
		}
	})

	t.Run("non-existent project", func(t *testing.T) {
		_, _, err := s.handleGetVulnerabilities(ctx, nil, VulnerabilitiesInput{Project: "nonexistent/repo"})
		if err == nil {
			t.Error("expected error for non-existent project")
		}
	})
}

func TestServer_HandleGetMalcontent(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	_, output, err := s.handleGetMalcontent(ctx, nil, MalcontentInput{Project: "owner1/repo1"})
	if err != nil {
		t.Fatalf("handleGetMalcontent failed: %v", err)
	}

	if !containsString(output.Text, "suspicious-pkg") {
		t.Error("output missing malcontent finding")
	}
}

func TestServer_HandleGetTechnologies(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	_, output, err := s.handleGetTechnologies(ctx, nil, ProjectInput{Project: "owner1/repo1"})
	if err != nil {
		t.Fatalf("handleGetTechnologies failed: %v", err)
	}

	if !containsString(output.Text, "React") {
		t.Error("output missing React technology")
	}
	if !containsString(output.Text, "Node.js") {
		t.Error("output missing Node.js technology")
	}
}

func TestServer_HandleGetPackageHealth(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	_, output, err := s.handleGetPackageHealth(ctx, nil, ProjectInput{Project: "owner1/repo1"})
	if err != nil {
		t.Fatalf("handleGetPackageHealth failed: %v", err)
	}

	if !containsString(output.Text, "lodash") {
		t.Error("output missing lodash health data")
	}
	if !containsString(output.Text, "healthy") {
		t.Error("output missing health status")
	}
}

func TestServer_HandleGetLicenses(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	_, output, err := s.handleGetLicenses(ctx, nil, ProjectInput{Project: "owner1/repo1"})
	if err != nil {
		t.Fatalf("handleGetLicenses failed: %v", err)
	}

	if !containsString(output.Text, "MIT") {
		t.Error("output missing MIT license")
	}
}

func TestServer_HandleGetSecrets(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	_, output, err := s.handleGetSecrets(ctx, nil, ProjectInput{Project: "owner1/repo1"})
	if err != nil {
		t.Fatalf("handleGetSecrets failed: %v", err)
	}

	if !containsString(output.Text, "api_key") {
		t.Error("output missing secret type")
	}
	if !containsString(output.Text, "config.js") {
		t.Error("output missing secret file")
	}
}

func TestServer_HandleGetCryptoIssues(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	_, output, err := s.handleGetCryptoIssues(ctx, nil, ProjectInput{Project: "owner1/repo1"})
	if err != nil {
		t.Fatalf("handleGetCryptoIssues failed: %v", err)
	}

	// Check for various crypto findings
	if !containsString(output.Text, "weak_ciphers") || !containsString(output.Text, "DES") {
		t.Error("output missing weak cipher finding")
	}
	if !containsString(output.Text, "hardcoded_keys") {
		t.Error("output missing hardcoded keys finding")
	}
	if !containsString(output.Text, "tls_issues") || !containsString(output.Text, "TLS 1.0") {
		t.Error("output missing TLS issue")
	}
	if !containsString(output.Text, "weak_random") || !containsString(output.Text, "Math.random") {
		t.Error("output missing weak random finding")
	}
}

func TestServer_HandleGetAnalysisRaw(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	t.Run("valid analysis", func(t *testing.T) {
		_, output, err := s.handleGetAnalysisRaw(ctx, nil, AnalysisRawInput{
			Project:      "owner1/repo1",
			AnalysisType: "code-packages",
		})
		if err != nil {
			t.Fatalf("handleGetAnalysisRaw failed: %v", err)
		}

		if !containsString(output.Text, "summary") {
			t.Error("output missing summary")
		}
		if !containsString(output.Text, "findings") {
			t.Error("output missing findings")
		}
	})

	t.Run("invalid analysis type", func(t *testing.T) {
		_, _, err := s.handleGetAnalysisRaw(ctx, nil, AnalysisRawInput{
			Project:      "owner1/repo1",
			AnalysisType: "nonexistent",
		})
		if err == nil {
			t.Error("expected error for non-existent analysis type")
		}
	})
}

func TestServer_HandleSearchFindings(t *testing.T) {
	zeroHome, cleanup := setupTestZeroHome(t)
	defer cleanup()

	s := NewServer(zeroHome)
	ctx := context.Background()

	t.Run("search all projects", func(t *testing.T) {
		_, output, err := s.handleSearchFindings(ctx, nil, SearchInput{Query: "lodash"})
		if err != nil {
			t.Fatalf("handleSearchFindings failed: %v", err)
		}

		if !containsString(output.Text, "lodash") {
			t.Error("output missing lodash in results")
		}
		if !containsString(output.Text, "owner1/repo1") {
			t.Error("output missing project")
		}
	})

	t.Run("search specific project", func(t *testing.T) {
		_, output, err := s.handleSearchFindings(ctx, nil, SearchInput{
			Query:   "React",
			Project: "owner1/repo1",
		})
		if err != nil {
			t.Fatalf("handleSearchFindings failed: %v", err)
		}

		if !containsString(output.Text, "React") {
			t.Error("output missing React")
		}
	})

	t.Run("search specific type", func(t *testing.T) {
		_, output, err := s.handleSearchFindings(ctx, nil, SearchInput{
			Query: "api_key",
			Type:  "code-security",
		})
		if err != nil {
			t.Fatalf("handleSearchFindings failed: %v", err)
		}

		if !containsString(output.Text, "api_key") {
			t.Error("output missing api_key")
		}
	})

	t.Run("no results", func(t *testing.T) {
		_, output, err := s.handleSearchFindings(ctx, nil, SearchInput{Query: "nonexistent-query-xyz"})
		if err != nil {
			t.Fatalf("handleSearchFindings failed: %v", err)
		}

		if !containsString(output.Text, `"count": 0`) {
			t.Error("should have 0 results")
		}
	})
}

// Helper function for test assertions
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmarks

func BenchmarkServer_GetProjects(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "zero-mcp-bench")
	defer os.RemoveAll(tmpDir)

	// Create 100 projects
	reposDir := filepath.Join(tmpDir, "repos")
	for i := 0; i < 100; i++ {
		analysisDir := filepath.Join(reposDir, "owner", "repo"+string(rune('0'+i%10))+string(rune('0'+i/10)), "analysis")
		os.MkdirAll(analysisDir, 0755)
		os.WriteFile(filepath.Join(analysisDir, "code-packages.json"), []byte(`{}`), 0644)
	}

	s := NewServer(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.getProjects()
	}
}

func BenchmarkServer_ReadAnalysis(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "zero-mcp-bench")
	defer os.RemoveAll(tmpDir)

	analysisDir := filepath.Join(tmpDir, "repos", "owner", "repo", "analysis")
	os.MkdirAll(analysisDir, 0755)
	os.WriteFile(filepath.Join(analysisDir, "code-packages.json"), []byte(`{"summary": {}, "findings": {}}`), 0644)

	s := NewServer(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.readAnalysis("owner/repo", "code-packages")
	}
}
