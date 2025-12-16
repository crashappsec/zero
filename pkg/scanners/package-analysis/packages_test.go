package packageanalysis

import (
	"testing"
	"time"
)

func TestPackagesScanner_Name(t *testing.T) {
	s := &PackagesScanner{}
	if s.Name() != "package-analysis" {
		t.Errorf("Name() = %q, want %q", s.Name(), "package-analysis")
	}
}

func TestPackagesScanner_Description(t *testing.T) {
	s := &PackagesScanner{}
	desc := s.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestPackagesScanner_Dependencies(t *testing.T) {
	s := &PackagesScanner{}
	deps := s.Dependencies()
	// Depends on sbom scanner
	if len(deps) != 1 {
		t.Errorf("Dependencies() = %v, want 1 dependency", deps)
	}
	if deps[0] != "sbom" {
		t.Errorf("Dependencies()[0] = %q, want %q", deps[0], "sbom")
	}
}

func TestPackagesScanner_EstimateDuration(t *testing.T) {
	tests := []struct {
		name      string
		config    FeatureConfig
		fileCount int
		wantMin   time.Duration
	}{
		{
			name:      "minimal config",
			config:    FeatureConfig{},
			fileCount: 0,
			wantMin:   5 * time.Second,
		},
		{
			name: "vulns enabled",
			config: FeatureConfig{
				Vulns: VulnsConfig{Enabled: true},
			},
			fileCount: 0,
			wantMin:   10 * time.Second,
		},
		{
			name: "health enabled",
			config: FeatureConfig{
				Health: HealthConfig{Enabled: true},
			},
			fileCount: 0,
			wantMin:   15 * time.Second,
		},
		{
			name: "malcontent enabled",
			config: FeatureConfig{
				Malcontent: MalcontentConfig{Enabled: true},
			},
			fileCount: 4000,
			wantMin:   5 * time.Second,
		},
		{
			name: "reachability enabled",
			config: FeatureConfig{
				Reachability: ReachabilityConfig{Enabled: true},
			},
			fileCount: 0,
			wantMin:   35 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PackagesScanner{config: tt.config}
			got := s.EstimateDuration(tt.fileCount)
			if got < tt.wantMin {
				t.Errorf("EstimateDuration(%d) = %v, want at least %v", tt.fileCount, got, tt.wantMin)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check that key features are enabled by default
	if !cfg.Vulns.Enabled {
		t.Error("Vulns should be enabled by default")
	}
	if !cfg.Health.Enabled {
		t.Error("Health should be enabled by default")
	}
	if !cfg.Malcontent.Enabled {
		t.Error("Malcontent should be enabled by default")
	}
	if !cfg.Licenses.Enabled {
		t.Error("Licenses should be enabled by default")
	}
	if !cfg.Confusion.Enabled {
		t.Error("Confusion should be enabled by default")
	}
	if !cfg.Typosquats.Enabled {
		t.Error("Typosquats should be enabled by default")
	}
	if !cfg.Deprecations.Enabled {
		t.Error("Deprecations should be enabled by default")
	}
	if !cfg.Duplicates.Enabled {
		t.Error("Duplicates should be enabled by default")
	}

	// Check that slow/optional features are disabled by default
	if cfg.Provenance.Enabled {
		t.Error("Provenance should be disabled by default")
	}
	if cfg.Bundle.Enabled {
		t.Error("Bundle should be disabled by default")
	}
	if cfg.Recommendations.Enabled {
		t.Error("Recommendations should be disabled by default")
	}
	if cfg.Reachability.Enabled {
		t.Error("Reachability should be disabled by default")
	}

	// Check default KEV setting
	if !cfg.Vulns.IncludeKEV {
		t.Error("Vulns.IncludeKEV should be true by default")
	}
}

func TestQuickConfig(t *testing.T) {
	cfg := QuickConfig()

	// Only vulns and licenses should be enabled
	if !cfg.Vulns.Enabled {
		t.Error("Vulns should be enabled in quick config")
	}
	if !cfg.Licenses.Enabled {
		t.Error("Licenses should be enabled in quick config")
	}

	// Other features should be disabled
	if cfg.Health.Enabled {
		t.Error("Health should be disabled in quick config")
	}
	if cfg.Malcontent.Enabled {
		t.Error("Malcontent should be disabled in quick config")
	}
	if cfg.Confusion.Enabled {
		t.Error("Confusion should be disabled in quick config")
	}
	if cfg.Typosquats.Enabled {
		t.Error("Typosquats should be disabled in quick config")
	}
	if cfg.Deprecations.Enabled {
		t.Error("Deprecations should be disabled in quick config")
	}
	if cfg.Duplicates.Enabled {
		t.Error("Duplicates should be disabled in quick config")
	}
}

func TestSecurityConfig(t *testing.T) {
	cfg := SecurityConfig()

	// Security features should be enabled
	if !cfg.Vulns.Enabled {
		t.Error("Vulns should be enabled in security config")
	}
	if !cfg.Vulns.IncludeKEV {
		t.Error("Vulns.IncludeKEV should be true in security config")
	}
	if !cfg.Malcontent.Enabled {
		t.Error("Malcontent should be enabled in security config")
	}
	if !cfg.Confusion.Enabled {
		t.Error("Confusion should be enabled in security config")
	}
	if !cfg.Reachability.Enabled {
		t.Error("Reachability should be enabled in security config")
	}
	if !cfg.Typosquats.Enabled {
		t.Error("Typosquats should be enabled in security config")
	}

	// Non-security features should be disabled
	if cfg.Bundle.Enabled {
		t.Error("Bundle should be disabled in security config")
	}
	if cfg.Recommendations.Enabled {
		t.Error("Recommendations should be disabled in security config")
	}
}

func TestFullConfig(t *testing.T) {
	cfg := FullConfig()

	// All features should be enabled
	if !cfg.Vulns.Enabled {
		t.Error("Vulns should be enabled in full config")
	}
	if !cfg.Health.Enabled {
		t.Error("Health should be enabled in full config")
	}
	if !cfg.Malcontent.Enabled {
		t.Error("Malcontent should be enabled in full config")
	}
	if !cfg.Provenance.Enabled {
		t.Error("Provenance should be enabled in full config")
	}
	if !cfg.Bundle.Enabled {
		t.Error("Bundle should be enabled in full config")
	}
	if !cfg.Recommendations.Enabled {
		t.Error("Recommendations should be enabled in full config")
	}
	if !cfg.Confusion.Enabled {
		t.Error("Confusion should be enabled in full config")
	}
	if !cfg.Reachability.Enabled {
		t.Error("Reachability should be enabled in full config")
	}
	if !cfg.Licenses.Enabled {
		t.Error("Licenses should be enabled in full config")
	}
	if !cfg.Typosquats.Enabled {
		t.Error("Typosquats should be enabled in full config")
	}
	if !cfg.Deprecations.Enabled {
		t.Error("Deprecations should be enabled in full config")
	}
	if !cfg.Duplicates.Enabled {
		t.Error("Duplicates should be enabled in full config")
	}

	// Check extended settings
	if cfg.Health.MaxPackages != 100 {
		t.Errorf("Health.MaxPackages = %d, want 100 in full config", cfg.Health.MaxPackages)
	}
	if !cfg.Duplicates.CheckFunctionality {
		t.Error("Duplicates.CheckFunctionality should be enabled in full config")
	}
}

func TestAllowedLicenses(t *testing.T) {
	expected := []string{"MIT", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause", "ISC", "Unlicense", "CC0-1.0", "0BSD"}

	for _, lic := range expected {
		if !allowedLicenses[lic] {
			t.Errorf("allowedLicenses should include %q", lic)
		}
	}
}

func TestDeniedLicenses(t *testing.T) {
	expected := []string{"GPL-2.0", "GPL-2.0-only", "GPL-3.0", "GPL-3.0-only", "AGPL-3.0", "AGPL-3.0-only", "SSPL-1.0"}

	for _, lic := range expected {
		if !deniedLicenses[lic] {
			t.Errorf("deniedLicenses should include %q", lic)
		}
	}
}

func TestLooksInternal(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"internal-utils", true},
		{"private-lib", true},
		{"my-internal-package", true},
		{"utils-private", true},
		{"lodash", false},
		{"express", false},
		{"react", false},
		{"@scope/package", false},
	}

	for _, tt := range tests {
		got := looksInternal(tt.name)
		if got != tt.expected {
			t.Errorf("looksInternal(%q) = %v, want %v", tt.name, got, tt.expected)
		}
	}
}

func TestIsSimilar(t *testing.T) {
	// isSimilar compares char-by-char up to minLen, then adds length difference
	// Returns true if differences > 0 && differences <= 2
	tests := []struct {
		a        string
		b        string
		expected bool
	}{
		{"lodash", "1odash", true},   // 1 char difference at position 0
		{"lodash", "lodaxh", true},   // 1 char difference at position 4
		{"lodash", "lodash", false},  // identical - differences = 0
		{"lodash", "express", false}, // completely different
		{"ab", "cd", false},          // too short (< 3 chars)
		{"express", "lodash", false}, // length diff = 1, but many char differences
		{"lodashe", "lodash", true},  // 1 char added (length diff = 1)
		{"lodas", "lodash", true},    // 1 char removed (length diff = 1)
	}

	for _, tt := range tests {
		got := isSimilar(tt.a, tt.b)
		if got != tt.expected {
			t.Errorf("isSimilar(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-100, 100},
		{100, 100},
	}

	for _, tt := range tests {
		got := abs(tt.input)
		if got != tt.expected {
			t.Errorf("abs(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestExtractEcosystem(t *testing.T) {
	tests := []struct {
		purl     string
		expected string
	}{
		{"pkg:npm/lodash@4.17.21", "npm"},
		{"pkg:pypi/requests@2.28.0", "pypi"},
		{"pkg:golang/github.com/gin-gonic/gin@1.8.1", "golang"},
		{"pkg:cargo/serde@1.0.0", "cargo"},
		{"pkg:maven/org.apache.commons/commons-lang3@3.12.0", "maven"},
		{"invalid-purl", "unknown"},
		{"", "unknown"},
		{"pkg:", "unknown"},
	}

	for _, tt := range tests {
		got := extractEcosystem(tt.purl)
		if got != tt.expected {
			t.Errorf("extractEcosystem(%q) = %q, want %q", tt.purl, got, tt.expected)
		}
	}
}

func TestRunDuplicatesFeature(t *testing.T) {
	s := &PackagesScanner{
		config: FeatureConfig{
			Duplicates: DuplicatesConfig{
				Enabled:            true,
				CheckVersions:      true,
				CheckFunctionality: true,
			},
		},
	}

	components := []ComponentData{
		{Name: "express", Version: "4.17.0", Ecosystem: "npm"},
		{Name: "express", Version: "4.18.0", Ecosystem: "npm"},
		{Name: "moment", Version: "2.29.4", Ecosystem: "npm"},
		{Name: "dayjs", Version: "1.11.0", Ecosystem: "npm"},
	}

	result := s.runDuplicatesFeature(components)

	if result.Summary.TotalPackages != 4 {
		t.Errorf("TotalPackages = %d, want 4", result.Summary.TotalPackages)
	}

	// Should detect express duplicate versions
	if result.Summary.DuplicateVersions != 1 {
		t.Errorf("DuplicateVersions = %d, want 1", result.Summary.DuplicateVersions)
	}

	// Should detect moment + dayjs as duplicate functionality (date)
	if result.Summary.DuplicateFunctionality != 1 {
		t.Errorf("DuplicateFunctionality = %d, want 1", result.Summary.DuplicateFunctionality)
	}
}

func TestRunDuplicatesFeature_VersionsOnly(t *testing.T) {
	s := &PackagesScanner{
		config: FeatureConfig{
			Duplicates: DuplicatesConfig{
				Enabled:            true,
				CheckVersions:      true,
				CheckFunctionality: false,
			},
		},
	}

	components := []ComponentData{
		{Name: "lodash", Version: "4.17.20", Ecosystem: "npm"},
		{Name: "lodash", Version: "4.17.21", Ecosystem: "npm"},
		{Name: "moment", Version: "2.29.4", Ecosystem: "npm"},
		{Name: "dayjs", Version: "1.11.0", Ecosystem: "npm"},
	}

	result := s.runDuplicatesFeature(components)

	// Should detect lodash duplicate versions
	if result.Summary.DuplicateVersions != 1 {
		t.Errorf("DuplicateVersions = %d, want 1", result.Summary.DuplicateVersions)
	}

	// Should NOT detect functionality duplicates (disabled)
	if result.Summary.DuplicateFunctionality != 0 {
		t.Errorf("DuplicateFunctionality = %d, want 0 (disabled)", result.Summary.DuplicateFunctionality)
	}
}

func TestRunProvenanceFeature(t *testing.T) {
	s := &PackagesScanner{}

	components := []ComponentData{
		{Name: "lodash", Version: "4.17.21", Ecosystem: "npm"},
		{Name: "express", Version: "4.18.0", Ecosystem: "npm"},
	}

	result := s.runProvenanceFeature(nil, components)

	// Should count total packages
	if result.Summary.TotalPackages != 2 {
		t.Errorf("TotalPackages = %d, want 2", result.Summary.TotalPackages)
	}

	// All unverified (placeholder implementation)
	if result.Summary.UnverifiedCount != 2 {
		t.Errorf("UnverifiedCount = %d, want 2", result.Summary.UnverifiedCount)
	}
}

func TestRunRecommendationsFeature(t *testing.T) {
	s := &PackagesScanner{}

	// Create a result with vulns and health data
	scanResult := &Result{
		Summary: Summary{
			Vulns: &VulnsSummary{
				Critical: 3,
				High:     5,
			},
			Health: &HealthSummary{
				DeprecatedCount: 2,
			},
		},
	}

	result := s.runRecommendationsFeature(scanResult)

	// Should have recommendations based on critical vulns
	if result.Summary.SecurityRecommendations != 3 {
		t.Errorf("SecurityRecommendations = %d, want 3", result.Summary.SecurityRecommendations)
	}

	// Should have recommendations based on deprecated packages
	if result.Summary.HealthRecommendations != 2 {
		t.Errorf("HealthRecommendations = %d, want 2", result.Summary.HealthRecommendations)
	}

	// Total should be sum
	if result.Summary.TotalRecommendations != 5 {
		t.Errorf("TotalRecommendations = %d, want 5", result.Summary.TotalRecommendations)
	}
}

func TestRunRecommendationsFeature_NoVulns(t *testing.T) {
	s := &PackagesScanner{}

	// Create a result with no vulns or health issues
	scanResult := &Result{
		Summary: Summary{},
	}

	result := s.runRecommendationsFeature(scanResult)

	// Should have no recommendations
	if result.Summary.TotalRecommendations != 0 {
		t.Errorf("TotalRecommendations = %d, want 0", result.Summary.TotalRecommendations)
	}
}
