package liveapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDepsDevClient(t *testing.T) {
	client := NewDepsDevClient()

	if client == nil {
		t.Fatal("NewDepsDevClient returned nil")
	}

	if client.Client == nil {
		t.Fatal("Client.Client is nil")
	}

	if client.BaseURL != DepsDevBaseURL {
		t.Errorf("BaseURL = %q, want %q", client.BaseURL, DepsDevBaseURL)
	}

	if client.UserAgent != "Zero-Scanner/1.0 (deps.dev Query)" {
		t.Errorf("UserAgent = %q, want Zero-Scanner/1.0 (deps.dev Query)", client.UserAgent)
	}
}

func TestNewDepsDevClientWithTimeout(t *testing.T) {
	timeout := 5 * time.Second
	client := NewDepsDevClientWithTimeout(timeout)

	if client == nil {
		t.Fatal("NewDepsDevClientWithTimeout returned nil")
	}

	if client.HTTPClient.Timeout != timeout {
		t.Errorf("Timeout = %v, want %v", client.HTTPClient.Timeout, timeout)
	}
}

func TestNormalizeEcosystem(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"npm", "NPM"},
		{"pypi", "PYPI"},
		{"golang", "GO"},
		{"go", "GO"},
		{"maven", "MAVEN"},
		{"cargo", "CARGO"},
		{"nuget", "NUGET"},
		{"rubygems", "RUBYGEMS"},
		{"packagist", "PACKAGIST"},
		{"unknown", "unknown"},           // Unknown returns as-is
		{"CustomEco", "CustomEco"},       // Case preserved for unknown
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeEcosystem(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeEcosystem(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetVersionDetails(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/systems/NPM/packages/express/versions/4.18.2"
		if r.URL.Path != expectedPath {
			t.Errorf("Request path = %q, want %q", r.URL.Path, expectedPath)
		}

		if r.Method != "GET" {
			t.Errorf("Request method = %q, want GET", r.Method)
		}

		response := VersionDetails{
			VersionKey: VersionKey{
				System:  "NPM",
				Name:    "express",
				Version: "4.18.2",
			},
			IsDefault:    true,
			IsDeprecated: false,
			Licenses:     []string{"MIT"},
			SlsaProvenances: []SLSAProvenance{
				{
					SourceRepository: "github.com/expressjs/express",
					Commit:           "abc123",
					Verified:         true,
				},
			},
			Projects: []ProjectInfo{
				{
					ProjectKey: ProjectKey{ID: "github.com/expressjs/express"},
					Scorecard: &Scorecard{
						OverallScore: 7.5,
						Checks: []ScorecardCheck{
							{Name: "Code-Review", Score: 8, Reason: "8 out of 10"},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	client := &DepsDevClient{
		Client: NewClient(server.URL,
			WithTimeout(5*time.Second),
			WithCache(1*time.Minute),
		),
	}

	ctx := context.Background()
	details, err := client.GetVersionDetails(ctx, "npm", "express", "4.18.2")
	if err != nil {
		t.Fatalf("GetVersionDetails() error = %v", err)
	}

	if details.VersionKey.Name != "express" {
		t.Errorf("Name = %q, want express", details.VersionKey.Name)
	}

	if details.VersionKey.Version != "4.18.2" {
		t.Errorf("Version = %q, want 4.18.2", details.VersionKey.Version)
	}

	if !details.IsDefault {
		t.Error("IsDefault should be true")
	}

	if details.IsDeprecated {
		t.Error("IsDeprecated should be false")
	}

	if len(details.Licenses) != 1 || details.Licenses[0] != "MIT" {
		t.Errorf("Licenses = %v, want [MIT]", details.Licenses)
	}

	if len(details.SlsaProvenances) != 1 {
		t.Errorf("SlsaProvenances length = %d, want 1", len(details.SlsaProvenances))
	}

	if len(details.Projects) != 1 || details.Projects[0].Scorecard == nil {
		t.Error("Projects should have scorecard")
	}
}

func TestGetVersionDetails_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	_, err := client.GetVersionDetails(ctx, "npm", "nonexistent-package-xyz", "1.0.0")
	if err == nil {
		t.Fatal("Expected error for non-existent package")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if !apiErr.IsNotFound() {
		t.Errorf("Expected 404 error, got status %d", apiErr.StatusCode)
	}
}

func TestGetPackageVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/systems/NPM/packages/lodash"
		if r.URL.Path != expectedPath {
			t.Errorf("Request path = %q, want %q", r.URL.Path, expectedPath)
		}

		response := PackageInfo{
			PackageKey: PackageKey{
				System: "NPM",
				Name:   "lodash",
			},
			Versions: []VersionInfo{
				{
					VersionKey: VersionKey{System: "NPM", Name: "lodash", Version: "4.17.20"},
					IsDefault:  false,
				},
				{
					VersionKey: VersionKey{System: "NPM", Name: "lodash", Version: "4.17.21"},
					IsDefault:  true,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	info, err := client.GetPackageVersions(ctx, "npm", "lodash")
	if err != nil {
		t.Fatalf("GetPackageVersions() error = %v", err)
	}

	if info.PackageKey.Name != "lodash" {
		t.Errorf("Name = %q, want lodash", info.PackageKey.Name)
	}

	if len(info.Versions) != 2 {
		t.Errorf("Versions count = %d, want 2", len(info.Versions))
	}
}

func TestIsDeprecated(t *testing.T) {
	tests := []struct {
		name         string
		isDeprecated bool
	}{
		{"deprecated-package", true},
		{"active-package", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := VersionDetails{
					VersionKey: VersionKey{
						System:  "NPM",
						Name:    tt.name,
						Version: "1.0.0",
					},
					IsDeprecated: tt.isDeprecated,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client := &DepsDevClient{
				Client: NewClient(server.URL, WithTimeout(5*time.Second)),
			}

			ctx := context.Background()
			deprecated, err := client.IsDeprecated(ctx, "npm", tt.name, "1.0.0")
			if err != nil {
				t.Fatalf("IsDeprecated() error = %v", err)
			}

			if deprecated != tt.isDeprecated {
				t.Errorf("IsDeprecated() = %v, want %v", deprecated, tt.isDeprecated)
			}
		})
	}
}

func TestGetHealthScore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := VersionDetails{
			VersionKey: VersionKey{
				System:  "NPM",
				Name:    "test-pkg",
				Version: "1.0.0",
			},
			IsDeprecated: false,
			SlsaProvenances: []SLSAProvenance{
				{
					SourceRepository: "github.com/test/test-pkg",
					Verified:         true,
				},
			},
			Projects: []ProjectInfo{
				{
					ProjectKey: ProjectKey{ID: "github.com/test/test-pkg"},
					Scorecard: &Scorecard{
						OverallScore: 8.5,
						Checks: []ScorecardCheck{
							{Name: "Code-Review", Score: 9},
							{Name: "Maintained", Score: 8},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	health, err := client.GetHealthScore(ctx, "npm", "test-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("GetHealthScore() error = %v", err)
	}

	if health.Score != 8.5 {
		t.Errorf("Score = %v, want 8.5", health.Score)
	}

	if health.IsDeprecated {
		t.Error("IsDeprecated should be false")
	}

	if !health.HasProvenance {
		t.Error("HasProvenance should be true")
	}

	if health.ProvenanceInfo == nil {
		t.Error("ProvenanceInfo should not be nil")
	} else if !health.ProvenanceInfo.Verified {
		t.Error("ProvenanceInfo.Verified should be true")
	}

	if len(health.Checks) != 2 {
		t.Errorf("Checks count = %d, want 2", len(health.Checks))
	}
}

func TestGetHealthScore_NoScorecard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := VersionDetails{
			VersionKey: VersionKey{
				System:  "NPM",
				Name:    "small-pkg",
				Version: "1.0.0",
			},
			IsDeprecated: false,
			Projects:     []ProjectInfo{}, // No scorecard
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	health, err := client.GetHealthScore(ctx, "npm", "small-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("GetHealthScore() error = %v", err)
	}

	if health.Score != 0 {
		t.Errorf("Score = %v, want 0 (no scorecard)", health.Score)
	}

	if health.HasProvenance {
		t.Error("HasProvenance should be false")
	}
}

func TestGetSLSAProvenance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := VersionDetails{
			VersionKey: VersionKey{
				System:  "NPM",
				Name:    "sigstore-pkg",
				Version: "1.0.0",
			},
			SlsaProvenances: []SLSAProvenance{
				{
					SourceRepository: "github.com/sigstore/sigstore-pkg",
					Commit:           "abcdef123456",
					URL:              "https://provenance.example.com",
					Verified:         true,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	provenances, err := client.GetSLSAProvenance(ctx, "npm", "sigstore-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("GetSLSAProvenance() error = %v", err)
	}

	if len(provenances) != 1 {
		t.Fatalf("Provenances count = %d, want 1", len(provenances))
	}

	if !provenances[0].Verified {
		t.Error("Provenance should be verified")
	}

	if provenances[0].SourceRepository != "github.com/sigstore/sigstore-pkg" {
		t.Errorf("SourceRepository = %q, want github.com/sigstore/sigstore-pkg", provenances[0].SourceRepository)
	}
}

func TestGetLatestVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PackageInfo{
			PackageKey: PackageKey{System: "NPM", Name: "axios"},
			Versions: []VersionInfo{
				{VersionKey: VersionKey{Version: "0.21.0"}, IsDefault: false},
				{VersionKey: VersionKey{Version: "0.21.1"}, IsDefault: false},
				{VersionKey: VersionKey{Version: "1.0.0"}, IsDefault: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	latest, err := client.GetLatestVersion(ctx, "npm", "axios")
	if err != nil {
		t.Fatalf("GetLatestVersion() error = %v", err)
	}

	if latest != "1.0.0" {
		t.Errorf("Latest version = %q, want 1.0.0", latest)
	}
}

func TestGetLatestVersion_NoDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PackageInfo{
			PackageKey: PackageKey{System: "NPM", Name: "old-pkg"},
			Versions: []VersionInfo{
				{VersionKey: VersionKey{Version: "1.0.0"}, IsDefault: false},
				{VersionKey: VersionKey{Version: "2.0.0"}, IsDefault: false},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	latest, err := client.GetLatestVersion(ctx, "npm", "old-pkg")
	if err != nil {
		t.Fatalf("GetLatestVersion() error = %v", err)
	}

	// Should return last version when no default
	if latest != "2.0.0" {
		t.Errorf("Latest version = %q, want 2.0.0 (last version)", latest)
	}
}

func TestGetLatestVersion_NoVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PackageInfo{
			PackageKey: PackageKey{System: "NPM", Name: "empty-pkg"},
			Versions:   []VersionInfo{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	_, err := client.GetLatestVersion(ctx, "npm", "empty-pkg")
	if err == nil {
		t.Fatal("Expected error for package with no versions")
	}
}

func TestIsOutdated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PackageInfo{
			PackageKey: PackageKey{System: "NPM", Name: "react"},
			Versions: []VersionInfo{
				{VersionKey: VersionKey{Version: "17.0.0"}, IsDefault: false},
				{VersionKey: VersionKey{Version: "18.0.0"}, IsDefault: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()

	// Test with outdated version
	outdated, latest, err := client.IsOutdated(ctx, "npm", "react", "17.0.0")
	if err != nil {
		t.Fatalf("IsOutdated() error = %v", err)
	}

	if !outdated {
		t.Error("17.0.0 should be outdated")
	}

	if latest != "18.0.0" {
		t.Errorf("Latest = %q, want 18.0.0", latest)
	}

	// Test with latest version
	outdated, latest, err = client.IsOutdated(ctx, "npm", "react", "18.0.0")
	if err != nil {
		t.Fatalf("IsOutdated() error = %v", err)
	}

	if outdated {
		t.Error("18.0.0 should not be outdated")
	}
}

func TestCaching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := VersionDetails{
			VersionKey: VersionKey{
				System:  "NPM",
				Name:    "cached-pkg",
				Version: "1.0.0",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL,
			WithTimeout(5*time.Second),
			WithCache(1*time.Hour),
		),
	}

	ctx := context.Background()

	// First call should hit server
	_, err := client.GetVersionDetails(ctx, "npm", "cached-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("First call error = %v", err)
	}

	if callCount != 1 {
		t.Errorf("Call count after first request = %d, want 1", callCount)
	}

	// Second call should use cache
	_, err = client.GetVersionDetails(ctx, "npm", "cached-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("Second call error = %v", err)
	}

	if callCount != 1 {
		t.Errorf("Call count after second request = %d, want 1 (cached)", callCount)
	}
}

func TestGetVersionDetailsByPURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check path contains purl - encoding may vary
		expectedPath := "/purl/pkg:npm/express@4.18.2"
		if r.URL.Path != expectedPath {
			t.Errorf("Request path = %q, want %q", r.URL.Path, expectedPath)
		}

		response := VersionDetails{
			VersionKey: VersionKey{
				System:  "NPM",
				Name:    "express",
				Version: "4.18.2",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	details, err := client.GetVersionDetailsByPURL(ctx, "pkg:npm/express@4.18.2")
	if err != nil {
		t.Fatalf("GetVersionDetailsByPURL() error = %v", err)
	}

	if details.VersionKey.Name != "express" {
		t.Errorf("Name = %q, want express", details.VersionKey.Name)
	}
}

// Integration test - requires network access to deps.dev API
func TestIntegration_RealDepsDevAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewDepsDevClient()
	ctx := context.Background()

	// Test with a well-known package
	details, err := client.GetVersionDetails(ctx, "npm", "lodash", "4.17.21")
	if err != nil {
		t.Fatalf("GetVersionDetails(lodash) error = %v", err)
	}

	if details.VersionKey.Name != "lodash" {
		t.Errorf("Name = %q, want lodash", details.VersionKey.Name)
	}

	if len(details.Licenses) == 0 {
		t.Error("Expected at least one license")
	}

	// Test health score
	health, err := client.GetHealthScore(ctx, "npm", "lodash", "4.17.21")
	if err != nil {
		t.Fatalf("GetHealthScore(lodash) error = %v", err)
	}

	t.Logf("Lodash health score: %.1f, deprecated: %v, provenance: %v",
		health.Score, health.IsDeprecated, health.HasProvenance)
}

// Test error handling for rate limiting
func TestRateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limited"}`))
	}))
	defer server.Close()

	client := &DepsDevClient{
		Client: NewClient(server.URL, WithTimeout(5*time.Second)),
	}

	ctx := context.Background()
	_, err := client.GetVersionDetails(ctx, "npm", "test", "1.0.0")
	if err == nil {
		t.Fatal("Expected error for rate limit")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if !apiErr.IsRateLimited() {
		t.Errorf("Expected rate limit error, got status %d", apiErr.StatusCode)
	}
}
