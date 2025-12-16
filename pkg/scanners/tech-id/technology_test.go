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
