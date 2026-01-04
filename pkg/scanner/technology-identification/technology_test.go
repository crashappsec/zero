package techid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsTensorFlowSavedModel(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		fileName string
		want     bool
	}{
		{
			name:     "saved_model.pb in my_model dir",
			path:     "/models/my_model/saved_model.pb",
			fileName: "saved_model.pb",
			want:     true, // parent dir contains "model"
		},
		{
			name:     "random .pb file",
			path:     "/protos/message.pb",
			fileName: "message.pb",
			want:     false, // not named saved_model.pb
		},
		{
			name:     "wrong filename",
			path:     "/models/model.pb",
			fileName: "model.pb",
			want:     false, // not named saved_model.pb
		},
		{
			name:     "saved_model.pb in random dir",
			path:     "/foo/bar/saved_model.pb",
			fileName: "saved_model.pb",
			want:     false, // parent dir doesn't match patterns
		},
		{
			name:     "saved_model.pb in export dir",
			path:     "/export/saved_model.pb",
			fileName: "saved_model.pb",
			want:     true, // parent dir contains "export"
		},
		{
			name:     "saved_model.pb in model dir",
			path:     "/trained_model/saved_model.pb",
			fileName: "saved_model.pb",
			want:     true, // parent dir contains "model"
		},
		{
			name:     "saved_model.pb in saved_model dir",
			path:     "/saved_model/saved_model.pb",
			fileName: "saved_model.pb",
			want:     true, // parent dir contains "saved_model"
		},
		{
			name:     "saved_model.pb in checkpoint dir",
			path:     "/checkpoint_v2/saved_model.pb",
			fileName: "saved_model.pb",
			want:     true, // parent dir contains "checkpoint"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTensorFlowSavedModel(tt.path, tt.fileName)
			if got != tt.want {
				t.Errorf("isTensorFlowSavedModel(%q, %q) = %v, want %v", tt.path, tt.fileName, got, tt.want)
			}
		})
	}
}

func TestIsTensorFlowSavedModelWithVariables(t *testing.T) {
	// Create a temporary directory structure that looks like a real SavedModel
	tmpDir, err := os.MkdirTemp("", "savedmodel-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create saved_model.pb
	savedModelPath := filepath.Join(tmpDir, "saved_model.pb")
	if err := os.WriteFile(savedModelPath, []byte("fake model"), 0644); err != nil {
		t.Fatalf("Failed to create saved_model.pb: %v", err)
	}

	// Create variables directory
	variablesDir := filepath.Join(tmpDir, "variables")
	if err := os.MkdirAll(variablesDir, 0755); err != nil {
		t.Fatalf("Failed to create variables dir: %v", err)
	}

	// Test that saved_model.pb with variables directory is detected
	got := isTensorFlowSavedModel(savedModelPath, "saved_model.pb")
	if !got {
		t.Errorf("isTensorFlowSavedModel should return true for SavedModel with variables directory")
	}
}

func TestModelFileFormats(t *testing.T) {
	// Verify .bin is NOT in the model file formats (removed to prevent false positives)
	if _, ok := ModelFileFormats[".bin"]; ok {
		t.Error(".bin should not be in ModelFileFormats (causes false positives)")
	}

	// Verify .pb is NOT in the model file formats (handled specially)
	if _, ok := ModelFileFormats[".pb"]; ok {
		t.Error(".pb should not be in ModelFileFormats (handled specially in detectModelFiles)")
	}

	// Verify legitimate model formats are present
	expectedFormats := []string{
		".pt", ".pth", ".pkl", ".pickle",
		".safetensors", ".onnx", ".gguf", ".ggml",
		".h5", ".keras", ".tflite", ".mlmodel",
	}

	for _, ext := range expectedFormats {
		if _, ok := ModelFileFormats[ext]; !ok {
			t.Errorf("Expected format %s not in ModelFileFormats", ext)
		}
	}
}

func TestMinModelFileSize(t *testing.T) {
	// Verify minimum file size is reasonable (10KB)
	if minModelFileSize != 10*1024 {
		t.Errorf("minModelFileSize should be 10KB (10240), got %d", minModelFileSize)
	}
}

func TestDetectModelFilesSkipsTestDirs(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "model-detect-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test directory with a model file (should be skipped)
	testDir := filepath.Join(tmpDir, "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Create a large .pt file in test directory (should be skipped)
	testModel := filepath.Join(testDir, "model.pt")
	largeData := make([]byte, 20*1024) // 20KB
	if err := os.WriteFile(testModel, largeData, 0644); err != nil {
		t.Fatalf("Failed to create test model: %v", err)
	}

	// Create a model file outside test directory (should be detected)
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}
	srcModel := filepath.Join(srcDir, "real_model.pt")
	if err := os.WriteFile(srcModel, largeData, 0644); err != nil {
		t.Fatalf("Failed to create src model: %v", err)
	}

	// Run detection
	scanner := &TechnologyScanner{}
	models := scanner.detectModelFiles(tmpDir)

	// Should only find the model in src/, not in test/
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

	if len(models) > 0 && models[0].FilePath != "src/real_model.pt" {
		t.Errorf("Expected model at src/real_model.pt, got %s", models[0].FilePath)
	}
}

func TestDetectModelFilesSkipsSmallFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "model-size-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a small .pt file (should be skipped)
	smallModel := filepath.Join(tmpDir, "small.pt")
	smallData := make([]byte, 1024) // 1KB - below minimum
	if err := os.WriteFile(smallModel, smallData, 0644); err != nil {
		t.Fatalf("Failed to create small model: %v", err)
	}

	// Create a large .pt file (should be detected)
	largeModel := filepath.Join(tmpDir, "large.pt")
	largeData := make([]byte, 20*1024) // 20KB - above minimum
	if err := os.WriteFile(largeModel, largeData, 0644); err != nil {
		t.Fatalf("Failed to create large model: %v", err)
	}

	// Run detection
	scanner := &TechnologyScanner{}
	models := scanner.detectModelFiles(tmpDir)

	// Should only find the large model
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

	if len(models) > 0 && models[0].Name != "large.pt" {
		t.Errorf("Expected large.pt, got %s", models[0].Name)
	}
}

func TestBuildTechnologySummary(t *testing.T) {
	techs := []Technology{
		{Name: "Go", Category: "language", Confidence: 95},
		{Name: "Python", Category: "language", Confidence: 80},
		{Name: "React", Category: "framework", Confidence: 90},
		{Name: "GitHub Actions", Category: "ci-cd", Confidence: 95},
	}

	scanner := &TechnologyScanner{}
	summary := scanner.buildTechnologySummary(techs)

	if summary == nil {
		t.Fatal("Technology summary is nil")
	}

	// Check total
	if summary.TotalTechnologies != 4 {
		t.Errorf("Expected 4 technologies, got %d", summary.TotalTechnologies)
	}

	// Check category counts
	if summary.ByCategory["language"] != 2 {
		t.Errorf("Expected 2 languages, got %d", summary.ByCategory["language"])
	}

	// Check top technologies (should be top 3)
	if len(summary.TopTechnologies) > 3 {
		t.Errorf("TopTechnologies should have at most 3 items, got %d", len(summary.TopTechnologies))
	}

	// Check primary languages
	if len(summary.PrimaryLanguages) != 2 {
		t.Errorf("Expected 2 primary languages, got %d", len(summary.PrimaryLanguages))
	}
}

func TestTechnologyScanner_Name(t *testing.T) {
	s := &TechnologyScanner{}
	if s.Name() != "technology-identification" {
		t.Errorf("Name() = %q, want %q", s.Name(), "technology-identification")
	}
}

func TestTechnologyScanner_Description(t *testing.T) {
	s := &TechnologyScanner{}
	desc := s.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestTechnologyScanner_Dependencies(t *testing.T) {
	s := &TechnologyScanner{}
	deps := s.Dependencies()
	// Returns empty slice, not nil
	if len(deps) != 0 {
		t.Errorf("Dependencies() = %v, want empty slice", deps)
	}
}

func TestTechnologyScanner_EstimateDuration(t *testing.T) {
	s := &TechnologyScanner{}

	tests := []struct {
		fileCount int
		wantMin   int
	}{
		{0, 5},     // Base is 5 seconds
		{500, 10},  // 5s + 500*10ms = 10s
		{1000, 15}, // 5s + 1000*10ms = 15s
		{5000, 55}, // 5s + 5000*10ms = 55s
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

	if !cfg.Technology.Enabled {
		t.Error("Technology should be enabled by default")
	}
	if !cfg.Models.Enabled {
		t.Error("Models should be enabled by default")
	}
	if !cfg.Frameworks.Enabled {
		t.Error("Frameworks should be enabled by default")
	}
	if !cfg.Security.Enabled {
		t.Error("Security should be enabled by default")
	}
	if !cfg.Governance.Enabled {
		t.Error("Governance should be enabled by default")
	}
	if !cfg.Semgrep.Enabled {
		t.Error("Semgrep should be enabled by default")
	}
	if !cfg.Security.CheckPickleFiles {
		t.Error("Security.CheckPickleFiles should be enabled by default")
	}
	if !cfg.Security.CheckAPIKeyExposure {
		t.Error("Security.CheckAPIKeyExposure should be enabled by default")
	}
}

func TestQuickConfig(t *testing.T) {
	cfg := QuickConfig()

	if cfg.Semgrep.Enabled {
		t.Error("Semgrep should be disabled in quick config")
	}
	if cfg.Technology.ScanExtensions {
		t.Error("Technology.ScanExtensions should be disabled in quick config")
	}
	if cfg.Models.QueryHuggingFace {
		t.Error("Models.QueryHuggingFace should be disabled in quick config")
	}
	if cfg.Datasets.Enabled {
		t.Error("Datasets should be disabled in quick config")
	}
	if cfg.Governance.Enabled {
		t.Error("Governance should be disabled in quick config")
	}
}

func TestSecurityOnlyConfig(t *testing.T) {
	cfg := SecurityOnlyConfig()

	if !cfg.Security.Enabled {
		t.Error("Security should be enabled in security-only config")
	}
	if !cfg.Security.CheckPickleFiles {
		t.Error("Security.CheckPickleFiles should be enabled in security-only config")
	}
	if !cfg.Security.DetectUnsafeLoading {
		t.Error("Security.DetectUnsafeLoading should be enabled in security-only config")
	}
	if !cfg.Security.CheckAPIKeyExposure {
		t.Error("Security.CheckAPIKeyExposure should be enabled in security-only config")
	}
	if cfg.Governance.Enabled {
		t.Error("Governance should be disabled in security-only config")
	}
	if cfg.Datasets.Enabled {
		t.Error("Datasets should be disabled in security-only config")
	}
}

func TestFullConfig(t *testing.T) {
	cfg := FullConfig()

	// All major features should be enabled
	if !cfg.Technology.Enabled {
		t.Error("Technology should be enabled in full config")
	}
	if !cfg.Models.Enabled {
		t.Error("Models should be enabled in full config")
	}
	if !cfg.Frameworks.Enabled {
		t.Error("Frameworks should be enabled in full config")
	}
	if !cfg.Datasets.Enabled {
		t.Error("Datasets should be enabled in full config")
	}
	if !cfg.Security.Enabled {
		t.Error("Security should be enabled in full config")
	}
	if !cfg.Governance.Enabled {
		t.Error("Governance should be enabled in full config")
	}
	if !cfg.Semgrep.Enabled {
		t.Error("Semgrep should be enabled in full config")
	}

	// Full config specifics
	if !cfg.Models.QueryTFHub {
		t.Error("Models.QueryTFHub should be enabled in full config")
	}
	if !cfg.Datasets.ScanDataFiles {
		t.Error("Datasets.ScanDataFiles should be enabled in full config")
	}
	if !cfg.Governance.RequireModelCards {
		t.Error("Governance.RequireModelCards should be enabled in full config")
	}
	if !cfg.Governance.RequireDatasetInfo {
		t.Error("Governance.RequireDatasetInfo should be enabled in full config")
	}
	if len(cfg.Governance.BlockedLicenses) == 0 {
		t.Error("Governance.BlockedLicenses should have entries in full config")
	}
}

func TestDetectFromConfigFiles(t *testing.T) {
	// Create temp directory with test config files
	tmpDir, err := os.MkdirTemp("", "config-detect-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json for Node/React detection
	packageJSON := `{
		"name": "test-app",
		"dependencies": {
			"react": "^18.2.0",
			"express": "^4.18.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create requirements.txt for Python detection
	requirementsTxt := `flask==2.0.0
pandas>=1.0.0
torch
`
	if err := os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte(requirementsTxt), 0644); err != nil {
		t.Fatalf("Failed to write requirements.txt: %v", err)
	}

	// Create Dockerfile
	dockerfile := `FROM python:3.11
RUN pip install flask
CMD ["python", "app.py"]
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	scanner := &TechnologyScanner{}
	techs := scanner.detectFromConfigFiles(tmpDir)

	// Should detect Node.js, React, Express, Python, Flask, PyTorch, Docker
	if len(techs) == 0 {
		t.Error("detectFromConfigFiles() should detect technologies")
	}

	// Check for expected technologies
	techNames := make(map[string]bool)
	for _, tech := range techs {
		techNames[tech.Name] = true
	}

	expectedTechs := []string{"Node.js", "Docker"}
	for _, expected := range expectedTechs {
		if !techNames[expected] {
			t.Errorf("Expected to detect %s", expected)
		}
	}
}

func TestDetectFromFileExtensions(t *testing.T) {
	// Create temp directory with various source files
	tmpDir, err := os.MkdirTemp("", "ext-detect-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create at least 3 files per extension (the detection requires count >= 3)
	testFiles := map[string]string{
		"main.go":    "package main",
		"utils.go":   "package utils",
		"handler.go": "package handler",
		"app.py":     "print('hello')",
		"utils.py":   "def util(): pass",
		"config.py":  "DEBUG = True",
		"index.js":   "console.log('hi')",
		"app.js":     "const app = {}",
		"utils.js":   "export function util() {}",
	}

	for name, content := range testFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", name, err)
		}
	}

	scanner := &TechnologyScanner{}
	techs := scanner.detectFromFileExtensions(tmpDir)

	// Should detect Go, Python, JavaScript (each has >= 3 files)
	if len(techs) < 3 {
		t.Errorf("detectFromFileExtensions() detected %d technologies, want at least 3", len(techs))
	}

	// Check for expected languages
	techNames := make(map[string]bool)
	for _, tech := range techs {
		techNames[tech.Name] = true
	}

	expectedLangs := []string{"Go", "Python", "JavaScript"}
	for _, expected := range expectedLangs {
		if !techNames[expected] {
			t.Errorf("Expected to detect %s", expected)
		}
	}
}

func TestConsolidateTechnologies(t *testing.T) {
	// Test that duplicate technologies are consolidated
	techs := []Technology{
		{Name: "Python", Category: "language", Confidence: 80, Source: "extension"},
		{Name: "Python", Category: "language", Confidence: 90, Source: "config"},
		{Name: "Go", Category: "language", Confidence: 95, Source: "extension"},
	}

	scanner := &TechnologyScanner{}
	consolidated := scanner.consolidateTechnologies(techs)

	// Should have 2 unique technologies (Python and Go)
	if len(consolidated) != 2 {
		t.Errorf("consolidateTechnologies() returned %d techs, want 2", len(consolidated))
	}

	// Python should have higher confidence (90)
	for _, tech := range consolidated {
		if tech.Name == "Python" && tech.Confidence != 90 {
			t.Errorf("Python confidence = %d, want 90 (highest)", tech.Confidence)
		}
	}
}

func TestModelExists(t *testing.T) {
	models := []MLModel{
		{Name: "gpt-4"},
		{Name: "bert-base-uncased"},
		{Name: "llama-2-7b"},
	}

	scanner := &TechnologyScanner{}

	tests := []struct {
		name     string
		expected bool
	}{
		{"gpt-4", true},
		{"bert-base-uncased", true},
		{"llama-2-7b", true},
		{"nonexistent-model", false},
		{"GPT-4", false}, // case sensitive
	}

	for _, tt := range tests {
		got := scanner.modelExists(models, tt.name)
		if got != tt.expected {
			t.Errorf("modelExists(%q) = %v, want %v", tt.name, got, tt.expected)
		}
	}
}

func TestCheckPickleFiles(t *testing.T) {
	// checkPickleFiles only flags models with SecurityRisk == "high" AND Format == "pickle"
	models := []MLModel{
		{Name: "safe-model", Format: "safetensors", FilePath: "model.safetensors", SecurityRisk: "low"},
		{Name: "unsafe-model", Format: "pickle", FilePath: "model.pkl", SecurityRisk: "high"},
		{Name: "another-unsafe", Format: "pickle", FilePath: "weights.pickle", SecurityRisk: "high"},
		{Name: "onnx-model", Format: "onnx", FilePath: "model.onnx", SecurityRisk: "low"},
		{Name: "pickle-low-risk", Format: "pickle", FilePath: "low.pkl", SecurityRisk: "low"}, // Not flagged
	}

	scanner := &TechnologyScanner{}
	findings := scanner.checkPickleFiles(models)

	// Should find 2 high-risk pickle files (unsafe-model and another-unsafe)
	if len(findings) != 2 {
		t.Errorf("checkPickleFiles() found %d findings, want 2", len(findings))
	}

	// Verify severity is high
	for _, f := range findings {
		if f.Severity != "high" {
			t.Errorf("Pickle file finding severity = %q, want high", f.Severity)
		}
		if f.Category != "pickle_rce" {
			t.Errorf("Pickle file finding category = %q, want pickle_rce", f.Category)
		}
	}
}

func TestCheckUnsafeLoading(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "unsafe-loading-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create Python file with unsafe torch.load
	unsafeCode := `
import torch

model = torch.load("model.pt")  # Unsafe - no weights_only
`
	if err := os.WriteFile(filepath.Join(tmpDir, "unsafe.py"), []byte(unsafeCode), 0644); err != nil {
		t.Fatalf("Failed to write unsafe.py: %v", err)
	}

	// Create Python file with safe torch.load
	safeCode := `
import torch

model = torch.load("model.pt", weights_only=True)  # Safe
`
	if err := os.WriteFile(filepath.Join(tmpDir, "safe.py"), []byte(safeCode), 0644); err != nil {
		t.Fatalf("Failed to write safe.py: %v", err)
	}

	scanner := &TechnologyScanner{}
	findings := scanner.checkUnsafeLoading(tmpDir)

	// Should find the unsafe torch.load
	if len(findings) == 0 {
		t.Error("checkUnsafeLoading() should find unsafe torch.load")
	}

	// Verify finding details
	for _, f := range findings {
		if f.Category != "unsafe_loading" {
			t.Errorf("Unsafe loading finding category = %q, want unsafe_loading", f.Category)
		}
	}
}

func TestCheckAPIKeyExposure(t *testing.T) {
	// Create temp directory with code files
	// Note: checkAPIKeyExposure skips files with "test" or "example" in the path
	tmpDir, err := os.MkdirTemp("", "apikey")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a src subdirectory to avoid "test" in path
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}

	// Create file with exposed API keys - must match patterns:
	// OpenAI: sk-[a-zA-Z0-9]{20,}
	// Anthropic: sk-ant-[a-zA-Z0-9-]{20,}
	exposedCode := `
import openai

openai.api_key = "sk-abcdefghij1234567890abcdefghij1234567890abcd"
`
	if err := os.WriteFile(filepath.Join(srcDir, "config.py"), []byte(exposedCode), 0644); err != nil {
		t.Fatalf("Failed to write config.py: %v", err)
	}

	scanner := &TechnologyScanner{}
	findings := scanner.checkAPIKeyExposure(tmpDir)

	// Should find exposed API keys
	if len(findings) == 0 {
		t.Error("checkAPIKeyExposure() should find exposed API keys")
	}

	// Verify finding details
	for _, f := range findings {
		if f.Category != "api_key_exposure" {
			t.Errorf("API key exposure finding category = %q, want api_key_exposure", f.Category)
		}
		if f.Severity != "high" && f.Severity != "critical" {
			t.Errorf("API key exposure finding severity = %q, want high or critical", f.Severity)
		}
	}
}
