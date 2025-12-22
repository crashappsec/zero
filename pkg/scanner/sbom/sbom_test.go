package sbom

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSBOMScanner_Name(t *testing.T) {
	s := &SBOMScanner{}
	if s.Name() != "sbom" {
		t.Errorf("Name() = %q, want %q", s.Name(), "sbom")
	}
}

func TestSBOMScanner_Description(t *testing.T) {
	s := &SBOMScanner{}
	desc := s.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestSBOMScanner_Dependencies(t *testing.T) {
	s := &SBOMScanner{}
	deps := s.Dependencies()
	if deps != nil {
		t.Errorf("Dependencies() = %v, want nil (sbom has no dependencies)", deps)
	}
}

func TestSBOMScanner_EstimateDuration(t *testing.T) {
	s := &SBOMScanner{}

	tests := []struct {
		fileCount int
		wantMin   int // minimum seconds expected
	}{
		{0, 10},
		{1000, 10},
		{5000, 10},
		{10000, 10},
	}

	for _, tt := range tests {
		got := s.EstimateDuration(tt.fileCount)
		if got.Seconds() < float64(tt.wantMin) {
			t.Errorf("EstimateDuration(%d) = %v, want at least %ds", tt.fileCount, got, tt.wantMin)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify generation is enabled by default
	if !cfg.Generation.Enabled {
		t.Error("Generation should be enabled by default")
	}

	// Verify integrity is enabled by default
	if !cfg.Integrity.Enabled {
		t.Error("Integrity should be enabled by default")
	}

	// Verify default tool is auto
	if cfg.Generation.Tool != "auto" {
		t.Errorf("Generation.Tool = %q, want %q", cfg.Generation.Tool, "auto")
	}

	// Verify default spec version
	if cfg.Generation.SpecVersion != "1.5" {
		t.Errorf("Generation.SpecVersion = %q, want %q", cfg.Generation.SpecVersion, "1.5")
	}
}

func TestExtractEcosystem(t *testing.T) {
	tests := []struct {
		purl     string
		expected string
	}{
		{"pkg:npm/lodash@4.17.21", "npm"},
		{"pkg:golang/github.com/gin-gonic/gin@1.9.0", "golang"},
		{"pkg:pypi/requests@2.28.0", "pypi"},
		{"pkg:maven/org.springframework/spring-core@5.3.0", "maven"},
		{"pkg:cargo/serde@1.0.0", "cargo"},
		{"pkg:gem/rails@7.0.0", "gem"},
		{"", ""},
		{"invalid-purl", ""},
		{"not-a-purl", ""},
	}

	for _, tt := range tests {
		got := extractEcosystem(tt.purl)
		if got != tt.expected {
			t.Errorf("extractEcosystem(%q) = %q, want %q", tt.purl, got, tt.expected)
		}
	}
}

func TestGetSBOMPath(t *testing.T) {
	outputDir := "/tmp/analysis"
	expected := "/tmp/analysis/sbom.cdx.json"

	got := GetSBOMPath(outputDir)
	if got != expected {
		t.Errorf("GetSBOMPath(%q) = %q, want %q", outputDir, got, expected)
	}
}

func TestLoadSBOM(t *testing.T) {
	// Create a temporary directory with a test SBOM
	tmpDir, err := os.MkdirTemp("", "sbom-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal valid CycloneDX SBOM as JSON
	sbomJSON := `{
		"bomFormat": "CycloneDX",
		"specVersion": "1.5",
		"version": 1,
		"components": [
			{
				"type": "library",
				"name": "lodash",
				"version": "4.17.21",
				"purl": "pkg:npm/lodash@4.17.21"
			},
			{
				"type": "library",
				"name": "express",
				"version": "4.18.2",
				"purl": "pkg:npm/express@4.18.2"
			}
		],
		"dependencies": [
			{"ref": "pkg:npm/express@4.18.2", "dependsOn": ["pkg:npm/lodash@4.17.21"]}
		]
	}`

	sbomPath := filepath.Join(tmpDir, "sbom.cdx.json")
	if err := os.WriteFile(sbomPath, []byte(sbomJSON), 0644); err != nil {
		t.Fatalf("Failed to write SBOM: %v", err)
	}

	// Load and verify
	findings, err := LoadSBOM(sbomPath)
	if err != nil {
		t.Fatalf("LoadSBOM() error = %v", err)
	}

	if len(findings.Components) != 2 {
		t.Errorf("LoadSBOM() returned %d components, want 2", len(findings.Components))
	}

	if len(findings.Dependencies) != 1 {
		t.Errorf("LoadSBOM() returned %d dependencies, want 1", len(findings.Dependencies))
	}

	// Verify component data
	found := false
	for _, c := range findings.Components {
		if c.Name == "lodash" && c.Version == "4.17.21" && c.Ecosystem == "npm" {
			found = true
			break
		}
	}
	if !found {
		t.Error("LoadSBOM() should include lodash component with correct data")
	}
}

func TestLoadSBOM_NotFound(t *testing.T) {
	_, err := LoadSBOM("/nonexistent/path/sbom.cdx.json")
	if err == nil {
		t.Error("LoadSBOM() should return error for non-existent file")
	}
}

func TestLoadSBOM_InvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tmpDir, err := os.MkdirTemp("", "sbom-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	_, err = LoadSBOM(invalidPath)
	if err == nil {
		t.Error("LoadSBOM() should return error for invalid JSON")
	}
}

func TestFindLockfilesRecursive(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "lockfile-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create lockfiles
	lockfiles := []string{
		"package-lock.json",
		"subdir/package-lock.json",
		"go.sum",
		"requirements.txt",
	}

	for _, lf := range lockfiles {
		path := filepath.Join(tmpDir, lf)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to write lockfile: %v", err)
		}
	}

	// Create node_modules directory (should be skipped)
	nodeModules := filepath.Join(tmpDir, "node_modules")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nodeModules, "package-lock.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write lockfile in node_modules: %v", err)
	}

	found := findLockfilesRecursive(tmpDir)

	// Should find 4 lockfiles, not the one in node_modules
	if len(found) != 4 {
		t.Errorf("findLockfilesRecursive() found %d lockfiles, want 4", len(found))
	}
}

func TestFindSBOMConfig(t *testing.T) {
	// Create a temporary directory with sbom.config.json
	tmpDir, err := os.MkdirTemp("", "sbom-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sbom.config.json
	configPath := filepath.Join(tmpDir, "sbom.config.json")
	if err := os.WriteFile(configPath, []byte(`{"version": "1.0"}`), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Should find it
	found := FindSBOMConfig(tmpDir)
	if found == "" {
		t.Error("FindSBOMConfig() should find sbom.config.json")
	}

	// Should not find in empty directory
	emptyDir, err := os.MkdirTemp("", "empty")
	if err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}
	defer os.RemoveAll(emptyDir)

	notFound := FindSBOMConfig(emptyDir)
	if notFound != "" {
		t.Errorf("FindSBOMConfig() returned %q for empty dir, want empty", notFound)
	}
}

func TestCompareLockfile(t *testing.T) {
	// Create a temporary directory with a package-lock.json
	tmpDir, err := os.MkdirTemp("", "lockfile-compare-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a package-lock.json with dependencies
	lockContent := `{
		"dependencies": {
			"lodash": {"version": "4.17.21"},
			"express": {"version": "4.18.2"}
		}
	}`
	lockPath := filepath.Join(tmpDir, "package-lock.json")
	if err := os.WriteFile(lockPath, []byte(lockContent), 0644); err != nil {
		t.Fatalf("Failed to write lockfile: %v", err)
	}

	// SBOM packages map
	sbomPkgs := map[string]string{
		"lodash":  "4.17.21",
		"express": "4.18.2",
		"react":   "18.0.0", // Extra package in SBOM
	}

	result := compareLockfile(lockPath, "npm", sbomPkgs)

	if result.Ecosystem != "npm" {
		t.Errorf("Ecosystem = %q, want %q", result.Ecosystem, "npm")
	}

	if result.InLockfile != 2 {
		t.Errorf("InLockfile = %d, want 2", result.InLockfile)
	}

	if result.InSBOM != 3 {
		t.Errorf("InSBOM = %d, want 3", result.InSBOM)
	}
}

func TestCompareYarnLock(t *testing.T) {
	// Create a temporary directory with a yarn.lock
	tmpDir, err := os.MkdirTemp("", "yarn-lock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simplified yarn.lock
	yarnContent := `"lodash@^4.17.0":
  version "4.17.21"

"express@^4.18.0":
  version "4.18.2"
`
	yarnPath := filepath.Join(tmpDir, "yarn.lock")
	if err := os.WriteFile(yarnPath, []byte(yarnContent), 0644); err != nil {
		t.Fatalf("Failed to write yarn.lock: %v", err)
	}

	sbomPkgs := map[string]string{
		"lodash":  "4.17.21",
		"express": "4.18.2",
	}

	result := compareYarnLock(yarnPath, sbomPkgs)

	if result.Ecosystem != "npm" {
		t.Errorf("Ecosystem = %q, want %q", result.Ecosystem, "npm")
	}
}

func TestCompareGoSum(t *testing.T) {
	// Create a temporary directory with a go.sum
	tmpDir, err := os.MkdirTemp("", "go-sum-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a go.sum file
	goSumContent := `github.com/gin-gonic/gin v1.9.0 h1:abc123
github.com/gin-gonic/gin v1.9.0/go.mod h1:def456
github.com/stretchr/testify v1.8.0 h1:xyz789
`
	goSumPath := filepath.Join(tmpDir, "go.sum")
	if err := os.WriteFile(goSumPath, []byte(goSumContent), 0644); err != nil {
		t.Fatalf("Failed to write go.sum: %v", err)
	}

	sbomPkgs := map[string]string{
		"github.com/gin-gonic/gin":  "1.9.0",
		"github.com/stretchr/testify": "1.8.0",
	}

	result := compareGoSum(goSumPath, sbomPkgs)

	if result.Ecosystem != "golang" {
		t.Errorf("Ecosystem = %q, want %q", result.Ecosystem, "golang")
	}

	if result.InLockfile < 2 {
		t.Errorf("InLockfile = %d, want at least 2", result.InLockfile)
	}
}

func TestCompareRequirements(t *testing.T) {
	// Create a temporary directory with requirements.txt
	tmpDir, err := os.MkdirTemp("", "requirements-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create requirements.txt
	reqContent := `requests==2.28.0
flask>=2.0.0
django~=4.0
pytest
# This is a comment
`
	reqPath := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(reqPath, []byte(reqContent), 0644); err != nil {
		t.Fatalf("Failed to write requirements.txt: %v", err)
	}

	sbomPkgs := map[string]string{
		"requests": "2.28.0",
		"flask":    "2.0.0",
		"django":   "4.0.0",
		"pytest":   "7.0.0",
	}

	result := compareRequirements(reqPath, sbomPkgs)

	if result.Ecosystem != "pypi" {
		t.Errorf("Ecosystem = %q, want %q", result.Ecosystem, "pypi")
	}

	if result.InLockfile != 4 {
		t.Errorf("InLockfile = %d, want 4", result.InLockfile)
	}
}

func TestComponentStruct(t *testing.T) {
	comp := Component{
		Type:       "library",
		Name:       "lodash",
		Version:    "4.17.21",
		Purl:       "pkg:npm/lodash@4.17.21",
		Ecosystem:  "npm",
		Scope:      "required",
		Licenses:   []string{"MIT"},
		Hashes:     []Hash{{Algorithm: "sha256", Content: "abc123"}},
		Properties: []Property{{Name: "key", Value: "value"}},
	}

	if comp.Name != "lodash" {
		t.Errorf("Name = %q, want %q", comp.Name, "lodash")
	}

	if comp.Ecosystem != "npm" {
		t.Errorf("Ecosystem = %q, want %q", comp.Ecosystem, "npm")
	}

	if len(comp.Licenses) != 1 || comp.Licenses[0] != "MIT" {
		t.Errorf("Licenses = %v, want [MIT]", comp.Licenses)
	}
}

func TestDependencyStruct(t *testing.T) {
	dep := Dependency{
		Ref:       "pkg:npm/express@4.18.2",
		DependsOn: []string{"pkg:npm/lodash@4.17.21", "pkg:npm/body-parser@1.20.0"},
	}

	if dep.Ref != "pkg:npm/express@4.18.2" {
		t.Errorf("Ref = %q, want %q", dep.Ref, "pkg:npm/express@4.18.2")
	}

	if len(dep.DependsOn) != 2 {
		t.Errorf("DependsOn has %d items, want 2", len(dep.DependsOn))
	}
}

func TestSBOMMetadataStruct(t *testing.T) {
	meta := SBOMMetadata{
		BomFormat:    "CycloneDX",
		SpecVersion:  "1.5",
		Version:      1,
		SerialNumber: "urn:uuid:abc123",
		Timestamp:    "2025-01-01T00:00:00Z",
		Tool:         "cdxgen",
	}

	if meta.BomFormat != "CycloneDX" {
		t.Errorf("BomFormat = %q, want %q", meta.BomFormat, "CycloneDX")
	}

	if meta.SpecVersion != "1.5" {
		t.Errorf("SpecVersion = %q, want %q", meta.SpecVersion, "1.5")
	}
}

func TestLockfileComparisonStruct(t *testing.T) {
	comp := LockfileComparison{
		Lockfile:   "package-lock.json",
		Ecosystem:  "npm",
		InLockfile: 150,
		InSBOM:     168,
		Matched:    145,
		Missing:    5,
		Extra:      23,
	}

	if comp.Lockfile != "package-lock.json" {
		t.Errorf("Lockfile = %q, want %q", comp.Lockfile, "package-lock.json")
	}

	// Verify counts are correct
	total := comp.Matched + comp.Missing + comp.Extra
	if total != comp.InLockfile+comp.Extra {
		t.Log("Lockfile comparison counts may vary based on comparison logic")
	}
}
