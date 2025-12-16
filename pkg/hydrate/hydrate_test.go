package hydrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/crashappsec/zero/pkg/github"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{100, "100"},
		{999, "999"},
		{1000, "1,000"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{1000000, "1,000,000"},
	}

	for _, tt := range tests {
		got := formatNumber(tt.input)
		if got != tt.want {
			t.Errorf("formatNumber(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0B"},
		{100, "100B"},
		{1023, "1023B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1572864, "1.5MB"},
		{1073741824, "1.0GB"},
		{1610612736, "1.5GB"},
	}

	for _, tt := range tests {
		got := formatBytes(tt.input)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice []string
		item  string
		want  bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"sbom", "package-analysis"}, "sbom", true},
		{[]string{"sbom", "package-analysis"}, "crypto", false},
	}

	for _, tt := range tests {
		got := contains(tt.slice, tt.item)
		if got != tt.want {
			t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.want)
		}
	}
}

func TestContainsScanner(t *testing.T) {
	tests := []struct {
		scanners []string
		name     string
		want     bool
	}{
		{[]string{"sbom", "package-analysis", "crypto"}, "sbom", true},
		{[]string{"sbom", "package-analysis", "crypto"}, "tech-id", false},
		{[]string{}, "sbom", false},
		{nil, "sbom", false},
	}

	for _, tt := range tests {
		got := containsScanner(tt.scanners, tt.name)
		if got != tt.want {
			t.Errorf("containsScanner(%v, %q) = %v, want %v", tt.scanners, tt.name, got, tt.want)
		}
	}
}

func TestFilterOutSkipped(t *testing.T) {
	tests := []struct {
		name         string
		scanners     []string
		skipScanners []string
		want         []string
	}{
		{
			name:         "no skips",
			scanners:     []string{"sbom", "package-analysis", "crypto"},
			skipScanners: []string{},
			want:         []string{"sbom", "package-analysis", "crypto"},
		},
		{
			name:         "skip one",
			scanners:     []string{"sbom", "package-analysis", "crypto"},
			skipScanners: []string{"crypto"},
			want:         []string{"sbom", "package-analysis"},
		},
		{
			name:         "skip multiple",
			scanners:     []string{"sbom", "package-analysis", "crypto", "devops"},
			skipScanners: []string{"crypto", "devops"},
			want:         []string{"sbom", "package-analysis"},
		},
		{
			name:         "skip non-existent",
			scanners:     []string{"sbom", "package-analysis"},
			skipScanners: []string{"crypto"},
			want:         []string{"sbom", "package-analysis"},
		},
		{
			name:         "nil skip list",
			scanners:     []string{"sbom", "package-analysis"},
			skipScanners: nil,
			want:         []string{"sbom", "package-analysis"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterOutSkipped(tt.scanners, tt.skipScanners)
			if len(got) != len(tt.want) {
				t.Errorf("filterOutSkipped() = %v, want %v", got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("filterOutSkipped()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestDeduplicateAndLimit(t *testing.T) {
	tests := []struct {
		name   string
		items  []string
		limit  int
		want   []string
	}{
		{
			name:   "no duplicates within limit",
			items:  []string{"Go", "Python", "JavaScript"},
			limit:  5,
			want:   []string{"Go", "Python", "JavaScript"},
		},
		{
			name:   "duplicates removed",
			items:  []string{"Go", "Python", "Go", "JavaScript", "Python"},
			limit:  5,
			want:   []string{"Go", "Python", "JavaScript"},
		},
		{
			name:   "limited results",
			items:  []string{"Go", "Python", "JavaScript", "Rust", "Ruby"},
			limit:  3,
			want:   []string{"Go", "Python", "JavaScript"},
		},
		{
			name:   "duplicates with limit",
			items:  []string{"Go", "Go", "Python", "JavaScript", "Rust", "Ruby"},
			limit:  3,
			want:   []string{"Go", "Python", "JavaScript"},
		},
		{
			name:   "empty input",
			items:  []string{},
			limit:  5,
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deduplicateAndLimit(tt.items, tt.limit)
			if len(got) != len(tt.want) {
				t.Errorf("deduplicateAndLimit() = %v, want %v", got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("deduplicateAndLimit()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestExtractSBOMStats(t *testing.T) {
	// Create a temporary directory with a mock SBOM file
	tmpDir, err := os.MkdirTemp("", "hydrate-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock sbom.cdx.json
	sbomData := struct {
		Components []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"components"`
	}{
		Components: []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			{Name: "express", Version: "4.18.0"},
			{Name: "lodash", Version: "4.17.21"},
			{Name: "react", Version: "18.0.0"},
		},
	}

	sbomJSON, _ := json.MarshalIndent(sbomData, "", "  ")
	sbomPath := filepath.Join(tmpDir, "sbom.cdx.json")
	if err := os.WriteFile(sbomPath, sbomJSON, 0644); err != nil {
		t.Fatalf("Failed to write sbom.cdx.json: %v", err)
	}

	// Create hydrate instance and test
	h := &Hydrate{zeroHome: tmpDir}
	status := &RepoStatus{}

	h.extractSBOMStats(status, tmpDir)

	if status.SBOMPackages != 3 {
		t.Errorf("SBOMPackages = %d, want 3", status.SBOMPackages)
	}

	if status.SBOMPath != sbomPath {
		t.Errorf("SBOMPath = %q, want %q", status.SBOMPath, sbomPath)
	}

	if status.SBOMSize <= 0 {
		t.Error("SBOMSize should be > 0")
	}
}

func TestExtractTechIDStats(t *testing.T) {
	// Create a temporary directory with a mock tech-id.json file
	tmpDir, err := os.MkdirTemp("", "hydrate-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock tech-id.json
	techIDData := map[string]interface{}{
		"summary": map[string]interface{}{
			"technology": map[string]interface{}{
				"total_technologies": 5,
				"top_technologies":   []string{"Go", "JavaScript", "Python"},
			},
			"models": map[string]interface{}{
				"total_models": 2,
			},
			"security": map[string]interface{}{
				"total_findings": 3,
			},
		},
	}

	techIDJSON, _ := json.MarshalIndent(techIDData, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "tech-id.json"), techIDJSON, 0644); err != nil {
		t.Fatalf("Failed to write tech-id.json: %v", err)
	}

	// Create hydrate instance and test
	h := &Hydrate{zeroHome: tmpDir}
	status := &RepoStatus{}

	h.extractTechIDStats(status, tmpDir)

	if status.TechIDTotalTech != 5 {
		t.Errorf("TechIDTotalTech = %d, want 5", status.TechIDTotalTech)
	}

	if len(status.TechIDTopTechs) != 3 {
		t.Errorf("TechIDTopTechs = %v, want 3 items", status.TechIDTopTechs)
	}

	if status.TechIDTotalModels != 2 {
		t.Errorf("TechIDTotalModels = %d, want 2", status.TechIDTotalModels)
	}

	if status.TechIDSecurityCount != 3 {
		t.Errorf("TechIDSecurityCount = %d, want 3", status.TechIDSecurityCount)
	}
}

func TestExtractOwnershipStats(t *testing.T) {
	// Create a temporary directory with a mock code-ownership.json file
	tmpDir, err := os.MkdirTemp("", "hydrate-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock code-ownership.json
	ownershipData := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_contributors":       10,
			"all_time_contributors":    25,
			"languages_detected":       3,
			"bus_factor":               4,
			"bus_factor_risk":          "low",
			"total_commits":            150,
			"days_since_last_commit":   5,
			"analysis_period_adjusted": true,
			"repo_activity_status":     "active",
			"top_languages": []map[string]interface{}{
				{"name": "Go", "file_count": 100, "percentage": 60.0},
				{"name": "JavaScript", "file_count": 50, "percentage": 30.0},
			},
		},
	}

	ownershipJSON, _ := json.MarshalIndent(ownershipData, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "code-ownership.json"), ownershipJSON, 0644); err != nil {
		t.Fatalf("Failed to write code-ownership.json: %v", err)
	}

	// Create hydrate instance and test
	h := &Hydrate{zeroHome: tmpDir}
	status := &RepoStatus{}

	h.extractOwnershipStats(status, tmpDir)

	if status.OwnershipContributors != 10 {
		t.Errorf("OwnershipContributors = %d, want 10", status.OwnershipContributors)
	}

	if status.OwnershipAllTime != 25 {
		t.Errorf("OwnershipAllTime = %d, want 25", status.OwnershipAllTime)
	}

	if status.OwnershipLanguages != 3 {
		t.Errorf("OwnershipLanguages = %d, want 3", status.OwnershipLanguages)
	}

	if status.OwnershipBusFactor != 4 {
		t.Errorf("OwnershipBusFactor = %d, want 4", status.OwnershipBusFactor)
	}

	if status.OwnershipBusFactorRisk != "low" {
		t.Errorf("OwnershipBusFactorRisk = %q, want %q", status.OwnershipBusFactorRisk, "low")
	}

	if status.OwnershipTotalCommits != 150 {
		t.Errorf("OwnershipTotalCommits = %d, want 150", status.OwnershipTotalCommits)
	}

	if status.OwnershipLastCommitDays != 5 {
		t.Errorf("OwnershipLastCommitDays = %d, want 5", status.OwnershipLastCommitDays)
	}

	if !status.OwnershipPeriodAdjusted {
		t.Error("OwnershipPeriodAdjusted should be true")
	}

	if status.OwnershipActivityStatus != "active" {
		t.Errorf("OwnershipActivityStatus = %q, want %q", status.OwnershipActivityStatus, "active")
	}

	if status.OwnershipTopLanguage != "Go" {
		t.Errorf("OwnershipTopLanguage = %q, want %q", status.OwnershipTopLanguage, "Go")
	}
}

func TestFormatScanCompleteMessage(t *testing.T) {
	tests := []struct {
		name   string
		status *RepoStatus
		want   string
	}{
		{
			name: "ownership only with bus factor",
			status: &RepoStatus{
				Repo:                   github.Repository{Name: "test-repo"},
				Duration:               10 * 1000000000, // 10 seconds in nanoseconds
				ScannersRun:            []string{"code-ownership"},
				OwnershipBusFactorRisk: "low",
				OwnershipContributors:  5,
				OwnershipAllTime:       10,
			},
			want: "test-repo complete (10s) - bus factor: low, 5 contributors (10 all-time)",
		},
		{
			name: "sbom with packages",
			status: &RepoStatus{
				Repo:         github.Repository{Name: "test-repo"},
				Duration:     5 * 1000000000,
				ScannersRun:  []string{"sbom"},
				SBOMPackages: 100,
				SBOMSize:     1048576, // 1MB
			},
			want: "test-repo complete (5s) - 100 packages, 1.0MB",
		},
		{
			name: "tech-id with technologies",
			status: &RepoStatus{
				Repo:              github.Repository{Name: "test-repo"},
				Duration:          3 * 1000000000,
				ScannersRun:       []string{"tech-id"},
				TechIDTotalTech:   5,
				TechIDTopTechs:    []string{"Go", "Python"},
				TechIDTotalModels: 2,
			},
			want: "test-repo complete (3s) - 5 tech: Go, Python, 2 models",
		},
		{
			name: "basic completion",
			status: &RepoStatus{
				Repo:        github.Repository{Name: "test-repo"},
				Duration:    2 * 1000000000,
				ScannersRun: []string{"crypto"},
			},
			want: "test-repo complete (2s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatScanCompleteMessage(tt.status)
			if got != tt.want {
				t.Errorf("formatScanCompleteMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsShallowRepo(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "shallow-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	h := &Hydrate{}

	// Test non-shallow repo (no .git/shallow file)
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	if h.isShallowRepo(tmpDir) {
		t.Error("isShallowRepo should return false for non-shallow repo")
	}

	// Create shallow file
	shallowFile := filepath.Join(gitDir, "shallow")
	if err := os.WriteFile(shallowFile, []byte("abc123"), 0644); err != nil {
		t.Fatalf("Failed to create shallow file: %v", err)
	}

	if !h.isShallowRepo(tmpDir) {
		t.Error("isShallowRepo should return true for shallow repo")
	}
}

func TestNeedsDeepHistory(t *testing.T) {
	tests := []struct {
		profile string
		want    bool
	}{
		{"quick", false},
		{"standard", false},
		{"security", false},
		{"full", true},
		{"health", true},
		{"ownership", true},
		{"code-ownership-only", true},
		{"supply-chain", false},
	}

	for _, tt := range tests {
		h := &Hydrate{opts: &Options{Profile: tt.profile}}
		got := h.needsDeepHistory()
		if got != tt.want {
			t.Errorf("needsDeepHistory() with profile %q = %v, want %v", tt.profile, got, tt.want)
		}
	}
}

func TestCountFiles(t *testing.T) {
	// Create a temporary directory with some files
	tmpDir, err := os.MkdirTemp("", "count-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some files
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create file1.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create file2.go: %v", err)
	}

	// Create a subdirectory with a file
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file3.go"), []byte("package sub"), 0644); err != nil {
		t.Fatalf("Failed to create file3.go: %v", err)
	}

	// Create .git directory with a file (should be ignored)
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main"), 0644); err != nil {
		t.Fatalf("Failed to create HEAD: %v", err)
	}

	h := &Hydrate{}
	count := h.countFiles(tmpDir)

	if count != 3 {
		t.Errorf("countFiles() = %d, want 3", count)
	}
}
