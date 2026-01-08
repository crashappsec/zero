package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crashappsec/zero/pkg/storage"
)

// testStore creates a new store with a temp database for testing.
func testStore(t *testing.T) (*Store, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zero-sqlite-test")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := New(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("creating store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestNew(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	if store == nil {
		t.Fatal("store is nil")
	}
	if store.db == nil {
		t.Fatal("db is nil")
	}
}

func TestNew_InvalidPath(t *testing.T) {
	// Attempt to create in a path that shouldn't be writable
	_, err := New("/nonexistent/path/that/doesnt/exist/test.db")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestStore_PingClose(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	// Ping should succeed
	if err := store.Ping(ctx); err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// Close should succeed
	if err := store.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Ping after close should fail
	if err := store.Ping(ctx); err == nil {
		t.Error("Ping should fail after Close")
	}
}

func TestStore_Projects(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("UpsertProject and GetProject", func(t *testing.T) {
		project := &storage.Project{
			ID:             "owner/repo",
			Owner:          "owner",
			Name:           "repo",
			RepoPath:       "/path/to/repo",
			AnalysisPath:   "/path/to/analysis",
			FileCount:      100,
			DiskSize:       1024000,
			LastScan:       time.Now().Truncate(time.Second),
			FreshnessLevel: "fresh",
			FreshnessAge:   2,
		}

		err := store.UpsertProject(ctx, project)
		if err != nil {
			t.Fatalf("UpsertProject failed: %v", err)
		}

		// Retrieve it
		got, err := store.GetProject(ctx, "owner/repo")
		if err != nil {
			t.Fatalf("GetProject failed: %v", err)
		}
		if got == nil {
			t.Fatal("GetProject returned nil")
		}

		if got.ID != project.ID {
			t.Errorf("ID = %q, want %q", got.ID, project.ID)
		}
		if got.Owner != project.Owner {
			t.Errorf("Owner = %q, want %q", got.Owner, project.Owner)
		}
		if got.Name != project.Name {
			t.Errorf("Name = %q, want %q", got.Name, project.Name)
		}
		if got.FileCount != project.FileCount {
			t.Errorf("FileCount = %d, want %d", got.FileCount, project.FileCount)
		}
		if got.FreshnessLevel != project.FreshnessLevel {
			t.Errorf("FreshnessLevel = %q, want %q", got.FreshnessLevel, project.FreshnessLevel)
		}
	})

	t.Run("GetProject non-existent", func(t *testing.T) {
		got, err := store.GetProject(ctx, "nonexistent/repo")
		if err != nil {
			t.Fatalf("GetProject failed: %v", err)
		}
		if got != nil {
			t.Error("expected nil for non-existent project")
		}
	})

	t.Run("UpsertProject update", func(t *testing.T) {
		project := &storage.Project{
			ID:             "owner/repo",
			Owner:          "owner",
			Name:           "repo",
			FileCount:      200, // Updated
			FreshnessLevel: "stale",
		}

		err := store.UpsertProject(ctx, project)
		if err != nil {
			t.Fatalf("UpsertProject failed: %v", err)
		}

		got, _ := store.GetProject(ctx, "owner/repo")
		if got.FileCount != 200 {
			t.Errorf("FileCount = %d, want 200", got.FileCount)
		}
		if got.FreshnessLevel != "stale" {
			t.Errorf("FreshnessLevel = %q, want stale", got.FreshnessLevel)
		}
	})
}

func TestStore_ListProjects(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple projects
	projects := []*storage.Project{
		{ID: "owner1/repo1", Owner: "owner1", Name: "repo1", FreshnessLevel: "fresh"},
		{ID: "owner1/repo2", Owner: "owner1", Name: "repo2", FreshnessLevel: "stale"},
		{ID: "owner2/repo1", Owner: "owner2", Name: "repo1", FreshnessLevel: "fresh"},
	}
	for _, p := range projects {
		if err := store.UpsertProject(ctx, p); err != nil {
			t.Fatalf("UpsertProject failed: %v", err)
		}
	}

	t.Run("list all", func(t *testing.T) {
		got, err := store.ListProjects(ctx, storage.ListOptions{})
		if err != nil {
			t.Fatalf("ListProjects failed: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("got %d projects, want 3", len(got))
		}
	})

	t.Run("filter by owner", func(t *testing.T) {
		got, err := store.ListProjects(ctx, storage.ListOptions{Owner: "owner1"})
		if err != nil {
			t.Fatalf("ListProjects failed: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d projects, want 2", len(got))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		got, err := store.ListProjects(ctx, storage.ListOptions{Limit: 2})
		if err != nil {
			t.Fatalf("ListProjects failed: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d projects, want 2", len(got))
		}

		got2, err := store.ListProjects(ctx, storage.ListOptions{Limit: 2, Offset: 2})
		if err != nil {
			t.Fatalf("ListProjects failed: %v", err)
		}
		if len(got2) != 1 {
			t.Errorf("got %d projects, want 1", len(got2))
		}
	})
}

func TestStore_DeleteProject(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create project with related data
	project := &storage.Project{
		ID:    "to-delete/repo",
		Owner: "to-delete",
		Name:  "repo",
	}
	store.UpsertProject(ctx, project)

	// Add related data
	scan := &storage.Scan{
		ID:        "scan-1",
		ProjectID: "to-delete/repo",
		Status:    "complete",
		StartedAt: time.Now(),
	}
	store.CreateScan(ctx, scan)

	summary := &storage.FindingsSummary{
		ProjectID:     "to-delete/repo",
		VulnsCritical: 5,
	}
	store.UpsertFindingsSummary(ctx, summary)

	// Delete
	err := store.DeleteProject(ctx, "to-delete/repo")
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	// Verify deleted
	got, _ := store.GetProject(ctx, "to-delete/repo")
	if got != nil {
		t.Error("project should have been deleted")
	}

	// Verify related data deleted
	gotScan, _ := store.GetScan(ctx, "scan-1")
	if gotScan != nil {
		t.Error("related scan should have been deleted")
	}

	gotSummary, _ := store.GetFindingsSummary(ctx, "to-delete/repo")
	if gotSummary != nil {
		t.Error("related summary should have been deleted")
	}
}

func TestStore_Scans(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create project first
	project := &storage.Project{ID: "test/repo", Owner: "test", Name: "repo"}
	store.UpsertProject(ctx, project)

	t.Run("CreateScan and GetScan", func(t *testing.T) {
		scan := &storage.Scan{
			ID:              "scan-1",
			ProjectID:       "test/repo",
			Profile:         "all-quick",
			Status:          "scanning",
			CommitSHA:       "abc123",
			StartedAt:       time.Now().Truncate(time.Second),
			DurationSeconds: 0,
		}

		err := store.CreateScan(ctx, scan)
		if err != nil {
			t.Fatalf("CreateScan failed: %v", err)
		}

		got, err := store.GetScan(ctx, "scan-1")
		if err != nil {
			t.Fatalf("GetScan failed: %v", err)
		}
		if got == nil {
			t.Fatal("GetScan returned nil")
		}

		if got.ID != scan.ID {
			t.Errorf("ID = %q, want %q", got.ID, scan.ID)
		}
		if got.Status != scan.Status {
			t.Errorf("Status = %q, want %q", got.Status, scan.Status)
		}
		if got.Profile != scan.Profile {
			t.Errorf("Profile = %q, want %q", got.Profile, scan.Profile)
		}
	})

	t.Run("UpdateScan", func(t *testing.T) {
		scan := &storage.Scan{
			ID:              "scan-1",
			Status:          "complete",
			FinishedAt:      time.Now().Truncate(time.Second),
			DurationSeconds: 120,
		}

		err := store.UpdateScan(ctx, scan)
		if err != nil {
			t.Fatalf("UpdateScan failed: %v", err)
		}

		got, _ := store.GetScan(ctx, "scan-1")
		if got.Status != "complete" {
			t.Errorf("Status = %q, want complete", got.Status)
		}
		if got.DurationSeconds != 120 {
			t.Errorf("DurationSeconds = %d, want 120", got.DurationSeconds)
		}
	})

	t.Run("UpdateScan with error", func(t *testing.T) {
		scan := &storage.Scan{
			ID:     "scan-1",
			Status: "failed",
			Error:  "something went wrong",
		}

		err := store.UpdateScan(ctx, scan)
		if err != nil {
			t.Fatalf("UpdateScan failed: %v", err)
		}

		got, _ := store.GetScan(ctx, "scan-1")
		if got.Error != "something went wrong" {
			t.Errorf("Error = %q, want 'something went wrong'", got.Error)
		}
	})

	t.Run("ListScans", func(t *testing.T) {
		// Create more scans
		for i := 2; i <= 5; i++ {
			scan := &storage.Scan{
				ID:        "scan-" + string(rune('0'+i)),
				ProjectID: "test/repo",
				Status:    "complete",
				StartedAt: time.Now().Add(time.Duration(i) * time.Hour),
			}
			store.CreateScan(ctx, scan)
		}

		scans, err := store.ListScans(ctx, "test/repo", storage.ListOptions{})
		if err != nil {
			t.Fatalf("ListScans failed: %v", err)
		}
		if len(scans) < 4 {
			t.Errorf("got %d scans, want at least 4", len(scans))
		}

		// Test with limit
		limited, _ := store.ListScans(ctx, "test/repo", storage.ListOptions{Limit: 2})
		if len(limited) != 2 {
			t.Errorf("got %d scans with limit, want 2", len(limited))
		}
	})

	t.Run("GetLatestScan", func(t *testing.T) {
		latest, err := store.GetLatestScan(ctx, "test/repo")
		if err != nil {
			t.Fatalf("GetLatestScan failed: %v", err)
		}
		if latest == nil {
			t.Fatal("GetLatestScan returned nil")
		}
	})

	t.Run("GetLatestScan no scans", func(t *testing.T) {
		latest, err := store.GetLatestScan(ctx, "nonexistent/repo")
		if err != nil {
			t.Fatalf("GetLatestScan failed: %v", err)
		}
		if latest != nil {
			t.Error("expected nil for project with no scans")
		}
	})
}

func TestStore_FindingsSummary(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create project first
	project := &storage.Project{ID: "test/repo", Owner: "test", Name: "repo"}
	store.UpsertProject(ctx, project)

	t.Run("UpsertFindingsSummary and GetFindingsSummary", func(t *testing.T) {
		summary := &storage.FindingsSummary{
			ProjectID:         "test/repo",
			VulnsCritical:     5,
			VulnsHigh:         10,
			VulnsMedium:       20,
			VulnsLow:          50,
			VulnsTotal:        85,
			SecretsTotal:      3,
			PackagesTotal:     150,
			TechnologiesTotal: 25,
		}

		err := store.UpsertFindingsSummary(ctx, summary)
		if err != nil {
			t.Fatalf("UpsertFindingsSummary failed: %v", err)
		}

		got, err := store.GetFindingsSummary(ctx, "test/repo")
		if err != nil {
			t.Fatalf("GetFindingsSummary failed: %v", err)
		}
		if got == nil {
			t.Fatal("GetFindingsSummary returned nil")
		}

		if got.VulnsCritical != 5 {
			t.Errorf("VulnsCritical = %d, want 5", got.VulnsCritical)
		}
		if got.VulnsTotal != 85 {
			t.Errorf("VulnsTotal = %d, want 85", got.VulnsTotal)
		}
		if got.SecretsTotal != 3 {
			t.Errorf("SecretsTotal = %d, want 3", got.SecretsTotal)
		}
		if got.PackagesTotal != 150 {
			t.Errorf("PackagesTotal = %d, want 150", got.PackagesTotal)
		}
	})

	t.Run("GetFindingsSummary non-existent", func(t *testing.T) {
		got, err := store.GetFindingsSummary(ctx, "nonexistent/repo")
		if err != nil {
			t.Fatalf("GetFindingsSummary failed: %v", err)
		}
		if got != nil {
			t.Error("expected nil for non-existent project")
		}
	})

	t.Run("UpsertFindingsSummary update", func(t *testing.T) {
		summary := &storage.FindingsSummary{
			ProjectID:     "test/repo",
			VulnsCritical: 10, // Updated
			VulnsTotal:    90,
		}

		err := store.UpsertFindingsSummary(ctx, summary)
		if err != nil {
			t.Fatalf("UpsertFindingsSummary failed: %v", err)
		}

		got, _ := store.GetFindingsSummary(ctx, "test/repo")
		if got.VulnsCritical != 10 {
			t.Errorf("VulnsCritical = %d, want 10", got.VulnsCritical)
		}
	})
}

func TestStore_Vulnerabilities(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create project first
	project := &storage.Project{ID: "test/repo", Owner: "test", Name: "repo"}
	store.UpsertProject(ctx, project)

	t.Run("UpsertVulnerabilities and GetVulnerabilities", func(t *testing.T) {
		vulns := []*storage.Vulnerability{
			{VulnID: "CVE-2023-001", Package: "lodash", Version: "4.17.0", Severity: "critical", Title: "Prototype Pollution", Source: "package", Scanner: "code-packages"},
			{VulnID: "CVE-2023-002", Package: "axios", Version: "0.21.0", Severity: "high", Title: "SSRF", Source: "package", Scanner: "code-packages"},
			{VulnID: "CVE-2023-003", Package: "express", Version: "4.17.0", Severity: "medium", Title: "XSS", Source: "package", Scanner: "code-packages"},
			{VulnID: "rule-001", Package: "src/auth.js", Severity: "low", Title: "Hardcoded secret", Source: "code", Scanner: "code-security"},
		}

		err := store.UpsertVulnerabilities(ctx, "test/repo", vulns)
		if err != nil {
			t.Fatalf("UpsertVulnerabilities failed: %v", err)
		}

		got, total, err := store.GetVulnerabilities(ctx, storage.VulnOptions{ProjectID: "test/repo"})
		if err != nil {
			t.Fatalf("GetVulnerabilities failed: %v", err)
		}
		if len(got) != 4 {
			t.Errorf("got %d vulns, want 4", len(got))
		}
		if total != 4 {
			t.Errorf("total = %d, want 4", total)
		}

		// Should be sorted by severity (critical first)
		if got[0].Severity != "critical" {
			t.Errorf("first vuln severity = %q, want critical", got[0].Severity)
		}
	})

	t.Run("GetVulnerabilities with filters", func(t *testing.T) {
		// Filter by severity
		got, _, err := store.GetVulnerabilities(ctx, storage.VulnOptions{
			ProjectID:  "test/repo",
			Severities: []string{"critical", "high"},
		})
		if err != nil {
			t.Fatalf("GetVulnerabilities failed: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d vulns with severity filter, want 2", len(got))
		}

		// Filter by package
		got2, _, _ := store.GetVulnerabilities(ctx, storage.VulnOptions{
			ProjectID: "test/repo",
			Package:   "lodash",
		})
		if len(got2) != 1 {
			t.Errorf("got %d vulns with package filter, want 1", len(got2))
		}
	})

	t.Run("GetVulnerabilities pagination", func(t *testing.T) {
		got, total, _ := store.GetVulnerabilities(ctx, storage.VulnOptions{
			ProjectID: "test/repo",
			Limit:     2,
		})
		if len(got) != 2 {
			t.Errorf("got %d vulns with limit, want 2", len(got))
		}
		if total != 4 {
			t.Errorf("total = %d, want 4", total)
		}

		got2, _, _ := store.GetVulnerabilities(ctx, storage.VulnOptions{
			ProjectID: "test/repo",
			Limit:     2,
			Offset:    2,
		})
		if len(got2) != 2 {
			t.Errorf("got %d vulns with offset, want 2", len(got2))
		}
	})

	t.Run("DeleteVulnerabilities", func(t *testing.T) {
		err := store.DeleteVulnerabilities(ctx, "test/repo")
		if err != nil {
			t.Fatalf("DeleteVulnerabilities failed: %v", err)
		}

		got, total, _ := store.GetVulnerabilities(ctx, storage.VulnOptions{ProjectID: "test/repo"})
		if len(got) != 0 || total != 0 {
			t.Error("vulnerabilities should have been deleted")
		}
	})
}

func TestStore_Secrets(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create project first
	project := &storage.Project{ID: "test/repo", Owner: "test", Name: "repo"}
	store.UpsertProject(ctx, project)

	t.Run("UpsertSecrets and GetSecrets", func(t *testing.T) {
		secrets := []*storage.Secret{
			{File: "config.js", Line: 10, Type: "api_key", Severity: "critical", Description: "AWS API Key", RedactedMatch: "AKIA***"},
			{File: ".env", Line: 5, Type: "password", Severity: "high", Description: "Database password", RedactedMatch: "pass***"},
			{File: "src/app.js", Line: 100, Type: "token", Severity: "medium", Description: "JWT Token", RedactedMatch: "eyJ***"},
		}

		err := store.UpsertSecrets(ctx, "test/repo", secrets)
		if err != nil {
			t.Fatalf("UpsertSecrets failed: %v", err)
		}

		got, total, err := store.GetSecrets(ctx, storage.SecretOptions{ProjectID: "test/repo"})
		if err != nil {
			t.Fatalf("GetSecrets failed: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("got %d secrets, want 3", len(got))
		}
		if total != 3 {
			t.Errorf("total = %d, want 3", total)
		}

		// Should be sorted by severity
		if got[0].Severity != "critical" {
			t.Errorf("first secret severity = %q, want critical", got[0].Severity)
		}
	})

	t.Run("GetSecrets with filters", func(t *testing.T) {
		// Filter by severity
		got, _, _ := store.GetSecrets(ctx, storage.SecretOptions{
			ProjectID:  "test/repo",
			Severities: []string{"critical"},
		})
		if len(got) != 1 {
			t.Errorf("got %d secrets with severity filter, want 1", len(got))
		}

		// Filter by type
		got2, _, _ := store.GetSecrets(ctx, storage.SecretOptions{
			ProjectID: "test/repo",
			Type:      "api_key",
		})
		if len(got2) != 1 {
			t.Errorf("got %d secrets with type filter, want 1", len(got2))
		}
	})

	t.Run("DeleteSecrets", func(t *testing.T) {
		err := store.DeleteSecrets(ctx, "test/repo")
		if err != nil {
			t.Fatalf("DeleteSecrets failed: %v", err)
		}

		got, total, _ := store.GetSecrets(ctx, storage.SecretOptions{ProjectID: "test/repo"})
		if len(got) != 0 || total != 0 {
			t.Error("secrets should have been deleted")
		}
	})
}

func TestStore_GetAggregateStats(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple projects with different freshness levels
	projects := []*storage.Project{
		{ID: "owner/repo1", Owner: "owner", Name: "repo1", FreshnessLevel: "fresh"},
		{ID: "owner/repo2", Owner: "owner", Name: "repo2", FreshnessLevel: "fresh"},
		{ID: "owner/repo3", Owner: "owner", Name: "repo3", FreshnessLevel: "stale"},
	}
	for _, p := range projects {
		store.UpsertProject(ctx, p)
	}

	// Add findings summaries
	summaries := []*storage.FindingsSummary{
		{ProjectID: "owner/repo1", VulnsCritical: 5, VulnsHigh: 10, VulnsMedium: 20, VulnsLow: 30, VulnsTotal: 65, SecretsTotal: 2, PackagesTotal: 100, TechnologiesTotal: 10},
		{ProjectID: "owner/repo2", VulnsCritical: 2, VulnsHigh: 5, VulnsMedium: 10, VulnsLow: 20, VulnsTotal: 37, SecretsTotal: 1, PackagesTotal: 50, TechnologiesTotal: 5},
		{ProjectID: "owner/repo3", VulnsCritical: 0, VulnsHigh: 3, VulnsMedium: 5, VulnsLow: 10, VulnsTotal: 18, SecretsTotal: 0, PackagesTotal: 30, TechnologiesTotal: 3},
	}
	for _, s := range summaries {
		store.UpsertFindingsSummary(ctx, s)
	}

	stats, err := store.GetAggregateStats(ctx)
	if err != nil {
		t.Fatalf("GetAggregateStats failed: %v", err)
	}

	if stats.TotalProjects != 3 {
		t.Errorf("TotalProjects = %d, want 3", stats.TotalProjects)
	}
	if stats.TotalVulns != 120 { // 65 + 37 + 18
		t.Errorf("TotalVulns = %d, want 120", stats.TotalVulns)
	}
	if stats.VulnsBySeverity["critical"] != 7 { // 5 + 2 + 0
		t.Errorf("critical = %d, want 7", stats.VulnsBySeverity["critical"])
	}
	if stats.TotalSecrets != 3 { // 2 + 1 + 0
		t.Errorf("TotalSecrets = %d, want 3", stats.TotalSecrets)
	}
	if stats.TotalPackages != 180 { // 100 + 50 + 30
		t.Errorf("TotalPackages = %d, want 180", stats.TotalPackages)
	}
	if stats.FreshnessCounts["fresh"] != 2 {
		t.Errorf("fresh count = %d, want 2", stats.FreshnessCounts["fresh"])
	}
	if stats.FreshnessCounts["stale"] != 1 {
		t.Errorf("stale count = %d, want 1", stats.FreshnessCounts["stale"])
	}
}

func TestStore_SyncProjectFromJSON(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	project := &storage.Project{ID: "test/repo", Owner: "test", Name: "repo"}
	store.UpsertProject(ctx, project)

	// Create temp analysis directory with test JSON files
	tmpDir, err := os.MkdirTemp("", "zero-analysis-test")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write code-packages.json
	packagesJSON := `{
		"summary": {
			"sbom": {
				"total_components": 150
			}
		},
		"findings": {
			"vulns": [
				{"id": "CVE-2023-001", "package": "lodash", "version": "4.17.0", "severity": "critical", "title": "Prototype Pollution"},
				{"id": "CVE-2023-002", "package": "axios", "version": "0.21.0", "severity": "high", "title": "SSRF"}
			]
		}
	}`
	os.WriteFile(filepath.Join(tmpDir, "code-packages.json"), []byte(packagesJSON), 0644)

	// Write code-security.json
	securityJSON := `{
		"findings": {
			"vulns": [
				{"rule_id": "hardcoded-secret", "severity": "medium", "message": "Hardcoded secret", "location": {"file": "config.js"}}
			],
			"secrets": [
				{"file": "config.js", "line": 10, "type": "api_key", "severity": "critical", "description": "AWS Key", "redacted_match": "AKIA***"}
			]
		}
	}`
	os.WriteFile(filepath.Join(tmpDir, "code-security.json"), []byte(securityJSON), 0644)

	// Write technology-identification.json
	techJSON := `{
		"summary": {
			"total_technologies": 25
		}
	}`
	os.WriteFile(filepath.Join(tmpDir, "technology-identification.json"), []byte(techJSON), 0644)

	// Sync
	err = store.SyncProjectFromJSON(ctx, "test/repo", tmpDir)
	if err != nil {
		t.Fatalf("SyncProjectFromJSON failed: %v", err)
	}

	// Verify findings summary
	summary, _ := store.GetFindingsSummary(ctx, "test/repo")
	if summary == nil {
		t.Fatal("FindingsSummary is nil")
	}
	if summary.VulnsCritical != 1 {
		t.Errorf("VulnsCritical = %d, want 1", summary.VulnsCritical)
	}
	if summary.VulnsHigh != 1 {
		t.Errorf("VulnsHigh = %d, want 1", summary.VulnsHigh)
	}
	if summary.VulnsMedium != 1 { // From code-security
		t.Errorf("VulnsMedium = %d, want 1", summary.VulnsMedium)
	}
	if summary.VulnsTotal != 3 {
		t.Errorf("VulnsTotal = %d, want 3", summary.VulnsTotal)
	}
	if summary.SecretsTotal != 1 {
		t.Errorf("SecretsTotal = %d, want 1", summary.SecretsTotal)
	}
	if summary.PackagesTotal != 150 {
		t.Errorf("PackagesTotal = %d, want 150", summary.PackagesTotal)
	}
	if summary.TechnologiesTotal != 25 {
		t.Errorf("TechnologiesTotal = %d, want 25", summary.TechnologiesTotal)
	}

	// Verify vulnerabilities
	vulns, total, _ := store.GetVulnerabilities(ctx, storage.VulnOptions{ProjectID: "test/repo"})
	if total != 3 {
		t.Errorf("total vulns = %d, want 3", total)
	}
	if len(vulns) != 3 {
		t.Errorf("got %d vulns, want 3", len(vulns))
	}

	// Verify secrets
	secrets, secretTotal, _ := store.GetSecrets(ctx, storage.SecretOptions{ProjectID: "test/repo"})
	if secretTotal != 1 {
		t.Errorf("total secrets = %d, want 1", secretTotal)
	}
	if len(secrets) != 1 {
		t.Errorf("got %d secrets, want 1", len(secrets))
	}
}

func TestNullHelpers(t *testing.T) {
	t.Run("nullTime", func(t *testing.T) {
		// Zero time
		nt := nullTime(time.Time{})
		if nt.Valid {
			t.Error("zero time should not be valid")
		}

		// Non-zero time
		now := time.Now()
		nt2 := nullTime(now)
		if !nt2.Valid {
			t.Error("non-zero time should be valid")
		}
		if !nt2.Time.Equal(now) {
			t.Error("time should match")
		}
	})

	t.Run("nullString", func(t *testing.T) {
		// Empty string
		ns := nullString("")
		if ns.Valid {
			t.Error("empty string should not be valid")
		}

		// Non-empty string
		ns2 := nullString("hello")
		if !ns2.Valid {
			t.Error("non-empty string should be valid")
		}
		if ns2.String != "hello" {
			t.Error("string should match")
		}
	})
}

// Benchmarks

func BenchmarkStore_UpsertProject(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "zero-bench")
	defer os.RemoveAll(tmpDir)

	store, _ := New(filepath.Join(tmpDir, "bench.db"))
	defer store.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		project := &storage.Project{
			ID:    "owner/repo",
			Owner: "owner",
			Name:  "repo",
		}
		store.UpsertProject(ctx, project)
	}
}

func BenchmarkStore_GetProject(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "zero-bench")
	defer os.RemoveAll(tmpDir)

	store, _ := New(filepath.Join(tmpDir, "bench.db"))
	defer store.Close()

	ctx := context.Background()

	project := &storage.Project{ID: "owner/repo", Owner: "owner", Name: "repo"}
	store.UpsertProject(ctx, project)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetProject(ctx, "owner/repo")
	}
}

func BenchmarkStore_GetVulnerabilities(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "zero-bench")
	defer os.RemoveAll(tmpDir)

	store, _ := New(filepath.Join(tmpDir, "bench.db"))
	defer store.Close()

	ctx := context.Background()

	project := &storage.Project{ID: "owner/repo", Owner: "owner", Name: "repo"}
	store.UpsertProject(ctx, project)

	// Insert 100 vulnerabilities
	var vulns []*storage.Vulnerability
	for i := 0; i < 100; i++ {
		vulns = append(vulns, &storage.Vulnerability{
			VulnID:   "CVE-2023-" + string(rune(i)),
			Package:  "package-" + string(rune(i%10)),
			Severity: []string{"critical", "high", "medium", "low"}[i%4],
		})
	}
	store.UpsertVulnerabilities(ctx, "owner/repo", vulns)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetVulnerabilities(ctx, storage.VulnOptions{ProjectID: "owner/repo", Limit: 20})
	}
}
