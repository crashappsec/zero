package cyclonedx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewBOM(t *testing.T) {
	bom := NewBOM()

	if bom.BOMFormat != "CycloneDX" {
		t.Errorf("expected BOMFormat 'CycloneDX', got '%s'", bom.BOMFormat)
	}
	if bom.SpecVersion != SpecVersion {
		t.Errorf("expected SpecVersion '%s', got '%s'", SpecVersion, bom.SpecVersion)
	}
	if bom.Version != 1 {
		t.Errorf("expected Version 1, got %d", bom.Version)
	}
	if bom.Metadata == nil {
		t.Error("expected Metadata to be set")
	}
	if bom.Metadata.Timestamp == "" {
		t.Error("expected Timestamp to be set")
	}
	if bom.Metadata.Tools == nil || len(bom.Metadata.Tools.Components) == 0 {
		t.Error("expected Tools to be set with Zero component")
	}
}

func TestNewMLBOM(t *testing.T) {
	bom := NewMLBOM()

	if bom.SerialNumber == "" {
		t.Error("expected SerialNumber to be set")
	}
	if !strings.HasPrefix(bom.SerialNumber, "urn:uuid:") {
		t.Errorf("expected SerialNumber to start with 'urn:uuid:', got '%s'", bom.SerialNumber)
	}
	if len(bom.Metadata.Lifecycles) == 0 {
		t.Error("expected Lifecycles to be set")
	}
	if bom.Metadata.Lifecycles[0].Phase != "discovery" {
		t.Errorf("expected phase 'discovery', got '%s'", bom.Metadata.Lifecycles[0].Phase)
	}
}

func TestNewCBOM(t *testing.T) {
	bom := NewCBOM()

	if bom.SerialNumber == "" {
		t.Error("expected SerialNumber to be set")
	}
	if len(bom.Metadata.Lifecycles) == 0 {
		t.Error("expected Lifecycles to be set")
	}
}

func TestWithComponent(t *testing.T) {
	bom := NewBOM()
	component := NewComponent(ComponentTypeLibrary, "test-lib")
	component.Version = "1.0.0"

	bom.WithComponent(component)

	if len(bom.Components) != 1 {
		t.Errorf("expected 1 component, got %d", len(bom.Components))
	}
	if bom.Components[0].Name != "test-lib" {
		t.Errorf("expected name 'test-lib', got '%s'", bom.Components[0].Name)
	}
}

func TestWithVulnerability(t *testing.T) {
	bom := NewBOM()
	vuln := Vulnerability{
		ID:          "CVE-2024-1234",
		Description: "Test vulnerability",
		Ratings: []VulnRating{
			{Severity: "high", Method: "CVSSv31"},
		},
	}

	bom.WithVulnerability(vuln)

	if len(bom.Vulnerabilities) != 1 {
		t.Errorf("expected 1 vulnerability, got %d", len(bom.Vulnerabilities))
	}
	if bom.Vulnerabilities[0].ID != "CVE-2024-1234" {
		t.Errorf("expected ID 'CVE-2024-1234', got '%s'", bom.Vulnerabilities[0].ID)
	}
}

func TestNewMLModelComponent(t *testing.T) {
	c := NewMLModelComponent("bert-base", "1.0.0")

	if c.Type != ComponentTypeMLModel {
		t.Errorf("expected type '%s', got '%s'", ComponentTypeMLModel, c.Type)
	}
	if c.Name != "bert-base" {
		t.Errorf("expected name 'bert-base', got '%s'", c.Name)
	}
	if c.BOMRef != "model/bert-base@1.0.0" {
		t.Errorf("expected BOMRef 'model/bert-base@1.0.0', got '%s'", c.BOMRef)
	}
}

func TestNewCryptoComponent(t *testing.T) {
	c := NewCryptoComponent("AES-256-GCM")

	if c.Type != ComponentTypeCryptographicAsset {
		t.Errorf("expected type '%s', got '%s'", ComponentTypeCryptographicAsset, c.Type)
	}
	if c.Name != "AES-256-GCM" {
		t.Errorf("expected name 'AES-256-GCM', got '%s'", c.Name)
	}
}

func TestAddProperty(t *testing.T) {
	c := NewComponent(ComponentTypeLibrary, "test")
	c.AddProperty("key1", "value1")
	c.AddProperty("key2", "value2")

	if len(c.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(c.Properties))
	}
	if c.Properties[0].Name != "key1" || c.Properties[0].Value != "value1" {
		t.Error("first property not set correctly")
	}
}

func TestAddExternalRef(t *testing.T) {
	c := NewComponent(ComponentTypeLibrary, "test")
	c.AddExternalRef(ExternalRefWebsite, "https://example.com")

	if len(c.ExternalRefs) != 1 {
		t.Errorf("expected 1 external ref, got %d", len(c.ExternalRefs))
	}
	if c.ExternalRefs[0].Type != ExternalRefWebsite {
		t.Errorf("expected type '%s', got '%s'", ExternalRefWebsite, c.ExternalRefs[0].Type)
	}
}

func TestAddLicense(t *testing.T) {
	c := NewComponent(ComponentTypeLibrary, "test")
	c.AddLicense("MIT")

	if len(c.Licenses) != 1 {
		t.Errorf("expected 1 license, got %d", len(c.Licenses))
	}
	if c.Licenses[0].License == nil || c.Licenses[0].License.ID != "MIT" {
		t.Error("license not set correctly")
	}
}

func TestSeverityToCycloneDX(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "critical"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"info", "info"},
		{"informational", "info"},
		{"unknown_severity", "unknown"},
	}

	for _, tt := range tests {
		result := SeverityToCycloneDX(tt.input)
		if result != tt.expected {
			t.Errorf("SeverityToCycloneDX(%s): expected '%s', got '%s'", tt.input, tt.expected, result)
		}
	}
}

func TestCWEToInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"CWE-79", 79},
		{"CWE-327", 327},
		{"CWE-798", 798},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		result := CWEToInt(tt.input)
		if result != tt.expected {
			t.Errorf("CWEToInt(%s): expected %d, got %d", tt.input, tt.expected, result)
		}
	}
}

func TestBOMToJSON(t *testing.T) {
	bom := NewBOM()
	bom.WithComponent(NewComponent(ComponentTypeLibrary, "test-lib"))

	data, err := bom.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("ToJSON produced invalid JSON: %v", err)
	}

	// Verify key fields
	if parsed["bomFormat"] != "CycloneDX" {
		t.Error("bomFormat not found in JSON")
	}
	if parsed["specVersion"] != SpecVersion {
		t.Error("specVersion not found in JSON")
	}
}

func TestFromJSON(t *testing.T) {
	original := NewMLBOM()
	original.WithComponent(NewMLModelComponent("test-model", "1.0.0"))

	data, _ := original.ToJSON()

	parsed, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if parsed.BOMFormat != original.BOMFormat {
		t.Errorf("BOMFormat mismatch: expected '%s', got '%s'", original.BOMFormat, parsed.BOMFormat)
	}
	if len(parsed.Components) != 1 {
		t.Errorf("expected 1 component, got %d", len(parsed.Components))
	}
}

func TestModelCard(t *testing.T) {
	mc := NewModelCard()
	mc.WithApproach(ApproachSupervised)
	mc.WithTask("text-classification")
	mc.WithArchitecture("transformer", "BERT")
	mc.WithDataset("dataset/imdb", "training")
	mc.WithMetric("accuracy", "0.95")
	mc.WithUseCase("sentiment analysis")
	mc.WithLimitation("English only")
	mc.WithEthicalConsideration("bias in training data", "diverse dataset curation")

	if mc.ModelParameters == nil {
		t.Fatal("ModelParameters should not be nil")
	}
	if mc.ModelParameters.Approach == nil || mc.ModelParameters.Approach.Type != ApproachSupervised {
		t.Error("Approach not set correctly")
	}
	if mc.ModelParameters.Task != "text-classification" {
		t.Error("Task not set correctly")
	}
	if mc.ModelParameters.ArchitectureFamily != "transformer" {
		t.Error("ArchitectureFamily not set correctly")
	}
	if len(mc.ModelParameters.Datasets) != 1 {
		t.Error("Datasets not set correctly")
	}
	if mc.QuantitativeAnalysis == nil || len(mc.QuantitativeAnalysis.PerformanceMetrics) != 1 {
		t.Error("PerformanceMetrics not set correctly")
	}
	if mc.Considerations == nil || len(mc.Considerations.UseCases) != 1 {
		t.Error("UseCases not set correctly")
	}
}

func TestNewAlgorithmComponent(t *testing.T) {
	c := NewAlgorithmComponent("AES", "gcm", 256)

	if c.Type != ComponentTypeCryptographicAsset {
		t.Errorf("expected type '%s', got '%s'", ComponentTypeCryptographicAsset, c.Type)
	}
	if c.CryptoProperties == nil {
		t.Fatal("CryptoProperties should not be nil")
	}
	if c.CryptoProperties.AssetType != CryptoAssetAlgorithm {
		t.Errorf("expected asset type '%s', got '%s'", CryptoAssetAlgorithm, c.CryptoProperties.AssetType)
	}
	if c.CryptoProperties.AlgorithmProperties == nil {
		t.Fatal("AlgorithmProperties should not be nil")
	}
	if c.CryptoProperties.AlgorithmProperties.Mode != "gcm" {
		t.Errorf("expected mode 'gcm', got '%s'", c.CryptoProperties.AlgorithmProperties.Mode)
	}
}

func TestNewProtocolComponent(t *testing.T) {
	c := NewProtocolComponent(ProtocolTLS, "1.3")

	if c.CryptoProperties == nil {
		t.Fatal("CryptoProperties should not be nil")
	}
	if c.CryptoProperties.AssetType != CryptoAssetProtocol {
		t.Errorf("expected asset type '%s', got '%s'", CryptoAssetProtocol, c.CryptoProperties.AssetType)
	}
	if c.CryptoProperties.ProtocolProperties.Type != ProtocolTLS {
		t.Errorf("expected protocol type '%s', got '%s'", ProtocolTLS, c.CryptoProperties.ProtocolProperties.Type)
	}
	if c.CryptoProperties.ProtocolProperties.Version != "1.3" {
		t.Errorf("expected version '1.3', got '%s'", c.CryptoProperties.ProtocolProperties.Version)
	}
}

func TestInferPrimitive(t *testing.T) {
	tests := []struct {
		algorithm string
		expected  string
	}{
		{"AES-GCM", PrimitiveAE},
		{"ChaCha20-Poly1305", PrimitiveAE},
		{"SHA-256", PrimitiveHash},
		{"MD5", PrimitiveHash},
		{"HMAC-SHA256", PrimitiveMAC},
		{"RSA-2048", PrimitiveDSA},
		{"ECDSA-P256", PrimitiveDSA},
		{"ML-KEM", PrimitiveKEM},
		{"HKDF", PrimitiveKDF},
		{"unknown-algo", PrimitiveOther},
	}

	for _, tt := range tests {
		result := inferPrimitive(tt.algorithm)
		if result != tt.expected {
			t.Errorf("inferPrimitive(%s): expected '%s', got '%s'", tt.algorithm, tt.expected, result)
		}
	}
}

func TestInferSecurityLevel(t *testing.T) {
	tests := []struct {
		algorithm string
		keySize   int
		expected  int
	}{
		{"MD5", 0, 0},        // Broken
		{"SHA-1", 0, 80},     // Deprecated
		{"SHA-256", 0, 128},
		{"AES", 128, 128},
		{"AES", 256, 256},
		{"RSA", 2048, 112},
		{"RSA", 4096, 140},
		{"Ed25519", 0, 128},
	}

	for _, tt := range tests {
		result := inferSecurityLevel(tt.algorithm, tt.keySize)
		if result != tt.expected {
			t.Errorf("inferSecurityLevel(%s, %d): expected %d, got %d", tt.algorithm, tt.keySize, tt.expected, result)
		}
	}
}

func TestExporterWriteMLBOM(t *testing.T) {
	tmpDir := t.TempDir()
	exporter := NewExporter(tmpDir)

	bom := NewMLBOM()
	bom.WithComponent(NewMLModelComponent("test-model", "1.0.0"))

	err := exporter.WriteMLBOM(bom, "test-mlbom.cdx.json")
	if err != nil {
		t.Fatalf("WriteMLBOM failed: %v", err)
	}

	// Verify file was created
	path := filepath.Join(tmpDir, "test-mlbom.cdx.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("ML-BOM file was not created")
	}

	// Verify content
	data, _ := os.ReadFile(path)
	var parsed BOM
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Written file is not valid JSON: %v", err)
	}
	if parsed.BOMFormat != "CycloneDX" {
		t.Error("Written BOM has wrong format")
	}
}

func TestExporterWriteCBOM(t *testing.T) {
	tmpDir := t.TempDir()
	exporter := NewExporter(tmpDir)

	bom := NewCBOM()
	bom.WithComponent(NewAlgorithmComponent("AES", "gcm", 256))

	err := exporter.WriteCBOM(bom, "test-cbom.cdx.json")
	if err != nil {
		t.Fatalf("WriteCBOM failed: %v", err)
	}

	// Verify file was created
	path := filepath.Join(tmpDir, "test-cbom.cdx.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("CBOM file was not created")
	}
}

func TestMLModelToComponent(t *testing.T) {
	c := MLModelToComponent(
		"gpt-2",
		"1.0.0",
		"huggingface",
		"https://huggingface.co/gpt2",
		"safetensors",
		"transformer",
		"text-generation",
		"MIT",
	)

	if c.Type != ComponentTypeMLModel {
		t.Errorf("expected type '%s', got '%s'", ComponentTypeMLModel, c.Type)
	}
	if c.Name != "gpt-2" {
		t.Errorf("expected name 'gpt-2', got '%s'", c.Name)
	}
	if len(c.Licenses) != 1 || c.Licenses[0].License.ID != "MIT" {
		t.Error("license not set correctly")
	}
	if c.ModelCard == nil {
		t.Error("ModelCard should be set")
	}
	if c.ModelCard.ModelParameters.Task != "text-generation" {
		t.Errorf("expected task 'text-generation', got '%s'", c.ModelCard.ModelParameters.Task)
	}
}

func TestCipherFindingToComponent(t *testing.T) {
	c := CipherFindingToComponent(
		"MD5",
		"",
		"high",
		"crypto.go",
		42,
		"MD5 is broken",
	)

	if c.Type != ComponentTypeCryptographicAsset {
		t.Errorf("expected type '%s', got '%s'", ComponentTypeCryptographicAsset, c.Type)
	}
	if c.CryptoProperties == nil {
		t.Fatal("CryptoProperties should be set")
	}
	if c.CryptoProperties.AssetType != CryptoAssetAlgorithm {
		t.Error("AssetType should be algorithm")
	}
	if c.Evidence == nil || len(c.Evidence.Occurrences) != 1 {
		t.Error("Evidence should include occurrence")
	}
	if c.Evidence.Occurrences[0].Line != 42 {
		t.Errorf("expected line 42, got %d", c.Evidence.Occurrences[0].Line)
	}
}
