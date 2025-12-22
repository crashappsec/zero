package sarif

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewLog(t *testing.T) {
	log := NewLog()

	if log.Schema != SchemaURI {
		t.Errorf("Schema = %q, want %q", log.Schema, SchemaURI)
	}
	if log.Version != SARIFVersion {
		t.Errorf("Version = %q, want %q", log.Version, SARIFVersion)
	}
	if log.Runs == nil {
		t.Error("Runs should not be nil")
	}
	if len(log.Runs) != 0 {
		t.Errorf("Runs length = %d, want 0", len(log.Runs))
	}
}

func TestNewRun(t *testing.T) {
	run := NewRun("test-tool", "1.0.0", "https://example.com")

	if run.Tool.Driver.Name != "test-tool" {
		t.Errorf("Tool name = %q, want %q", run.Tool.Driver.Name, "test-tool")
	}
	if run.Tool.Driver.Version != "1.0.0" {
		t.Errorf("Tool version = %q, want %q", run.Tool.Driver.Version, "1.0.0")
	}
	if run.Tool.Driver.InformationURI != "https://example.com" {
		t.Errorf("Tool informationUri = %q, want %q", run.Tool.Driver.InformationURI, "https://example.com")
	}
	if len(run.Results) != 0 {
		t.Errorf("Results length = %d, want 0", len(run.Results))
	}
	if len(run.Tool.Driver.Rules) != 0 {
		t.Errorf("Rules length = %d, want 0", len(run.Tool.Driver.Rules))
	}
}

func TestAddRule(t *testing.T) {
	run := NewRun("test-tool", "1.0.0", "")

	idx := run.AddRule("rule-1", "Rule One", "Description of rule one", "https://help.example.com", "error")

	if idx != 0 {
		t.Errorf("First rule index = %d, want 0", idx)
	}
	if len(run.Tool.Driver.Rules) != 1 {
		t.Errorf("Rules length = %d, want 1", len(run.Tool.Driver.Rules))
	}

	rule := run.Tool.Driver.Rules[0]
	if rule.ID != "rule-1" {
		t.Errorf("Rule ID = %q, want %q", rule.ID, "rule-1")
	}
	if rule.Name != "Rule One" {
		t.Errorf("Rule Name = %q, want %q", rule.Name, "Rule One")
	}
	if rule.ShortDescription.Text != "Description of rule one" {
		t.Errorf("Rule description = %q, want %q", rule.ShortDescription.Text, "Description of rule one")
	}
	if rule.HelpURI != "https://help.example.com" {
		t.Errorf("Rule helpUri = %q, want %q", rule.HelpURI, "https://help.example.com")
	}
	if rule.DefaultConfig.Level != "error" {
		t.Errorf("Rule level = %q, want %q", rule.DefaultConfig.Level, "error")
	}

	// Add second rule
	idx2 := run.AddRule("rule-2", "Rule Two", "Description", "", "warning")
	if idx2 != 1 {
		t.Errorf("Second rule index = %d, want 1", idx2)
	}
}

func TestAddResult(t *testing.T) {
	run := NewRun("test-tool", "1.0.0", "")
	ruleIdx := run.AddRule("sql-injection", "SQL Injection", "SQL injection vulnerability", "", "error")

	run.AddResult("sql-injection", ruleIdx, "error", "Found SQL injection", "src/db.go", 42)

	if len(run.Results) != 1 {
		t.Fatalf("Results length = %d, want 1", len(run.Results))
	}

	result := run.Results[0]
	if result.RuleID != "sql-injection" {
		t.Errorf("Result RuleID = %q, want %q", result.RuleID, "sql-injection")
	}
	if result.RuleIndex != ruleIdx {
		t.Errorf("Result RuleIndex = %d, want %d", result.RuleIndex, ruleIdx)
	}
	if result.Level != "error" {
		t.Errorf("Result Level = %q, want %q", result.Level, "error")
	}
	if result.Message.Text != "Found SQL injection" {
		t.Errorf("Result Message = %q, want %q", result.Message.Text, "Found SQL injection")
	}
	if len(result.Locations) != 1 {
		t.Fatalf("Result Locations length = %d, want 1", len(result.Locations))
	}

	loc := result.Locations[0]
	if loc.PhysicalLocation == nil {
		t.Fatal("PhysicalLocation should not be nil")
	}
	if loc.PhysicalLocation.ArtifactLocation.URI != "src/db.go" {
		t.Errorf("Location URI = %q, want %q", loc.PhysicalLocation.ArtifactLocation.URI, "src/db.go")
	}
	if loc.PhysicalLocation.Region.StartLine != 42 {
		t.Errorf("Location StartLine = %d, want 42", loc.PhysicalLocation.Region.StartLine)
	}
}

func TestAddResultNoFile(t *testing.T) {
	run := NewRun("test-tool", "1.0.0", "")
	run.AddRule("no-file-rule", "No File Rule", "Rule without file", "", "warning")

	run.AddResult("no-file-rule", 0, "warning", "Some message", "", 0)

	if len(run.Results) != 1 {
		t.Fatalf("Results length = %d, want 1", len(run.Results))
	}

	// No location should be added when file is empty
	if len(run.Results[0].Locations) != 0 {
		t.Errorf("Locations length = %d, want 0 (no file specified)", len(run.Results[0].Locations))
	}
}

func TestSeverityToLevel(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "error"},
		{"CRITICAL", "error"},
		{"high", "error"},
		{"HIGH", "error"},
		{"medium", "warning"},
		{"MEDIUM", "warning"},
		{"low", "note"},
		{"LOW", "note"},
		{"info", "note"},
		{"INFO", "note"},
		{"unknown", "none"},
		{"", "none"},
	}

	for _, tt := range tests {
		got := SeverityToLevel(tt.severity)
		if got != tt.expected {
			t.Errorf("SeverityToLevel(%q) = %q, want %q", tt.severity, got, tt.expected)
		}
	}
}

func TestWriteJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sarif-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	log := NewLog()
	run := NewRun("test-tool", "1.0.0", "https://example.com")
	run.AddRule("test-rule", "Test Rule", "Test description", "", "warning")
	run.AddResult("test-rule", 0, "warning", "Test message", "test.go", 10)
	log.Runs = append(log.Runs, *run)

	outPath := filepath.Join(tmpDir, "output.sarif.json")
	if err := log.WriteJSON(outPath); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var parsed Log
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if parsed.Version != SARIFVersion {
		t.Errorf("Parsed version = %q, want %q", parsed.Version, SARIFVersion)
	}
	if len(parsed.Runs) != 1 {
		t.Errorf("Parsed runs length = %d, want 1", len(parsed.Runs))
	}
}

func TestExporterWithEmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sarif-export-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	exporter := NewExporter(tmpDir, tmpDir)
	log, err := exporter.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Should return empty log with no runs (no analysis files)
	if log == nil {
		t.Fatal("Export returned nil log")
	}
	if len(log.Runs) != 0 {
		t.Errorf("Export from empty dir should have 0 runs, got %d", len(log.Runs))
	}
}

func TestExporterCodeSecurity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sarif-export-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a code-security.json file
	codeSecurityData := `{
		"findings": {
			"vulns": [
				{
					"rule_id": "sql-injection",
					"file": "src/db.go",
					"line": 42,
					"severity": "critical",
					"message": "SQL injection vulnerability",
					"category": "security"
				}
			],
			"secrets": [
				{
					"type": "aws_access_key",
					"file": "config.go",
					"line": 10,
					"severity": "critical"
				}
			]
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "code-security.json"), []byte(codeSecurityData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	exporter := NewExporter(tmpDir, tmpDir)
	log, err := exporter.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Should have 2 runs: one for code vulns, one for secrets
	if len(log.Runs) != 2 {
		t.Errorf("Expected 2 runs (vulns + secrets), got %d", len(log.Runs))
	}

	// Check code security run
	foundVuln := false
	foundSecret := false
	for _, run := range log.Runs {
		if run.Tool.Driver.Name == "zero-code-security" {
			foundVuln = true
			if len(run.Results) != 1 {
				t.Errorf("Code security run should have 1 result, got %d", len(run.Results))
			}
			if len(run.Results) > 0 && run.Results[0].Level != "error" {
				t.Errorf("Critical severity should map to error level, got %q", run.Results[0].Level)
			}
		}
		if run.Tool.Driver.Name == "zero-secrets" {
			foundSecret = true
			if len(run.Results) != 1 {
				t.Errorf("Secrets run should have 1 result, got %d", len(run.Results))
			}
		}
	}

	if !foundVuln {
		t.Error("Expected zero-code-security run not found")
	}
	if !foundSecret {
		t.Error("Expected zero-secrets run not found")
	}
}

func TestExporterPackageVulns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sarif-export-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a package-analysis.json file
	packageData := `{
		"findings": {
			"vulns": [
				{
					"id": "CVE-2021-12345",
					"aliases": ["GHSA-xxxx-yyyy-zzzz"],
					"package": "lodash",
					"version": "4.17.20",
					"severity": "high",
					"title": "Prototype pollution vulnerability",
					"ecosystem": "npm"
				}
			]
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "package-analysis.json"), []byte(packageData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	exporter := NewExporter(tmpDir, tmpDir)
	log, err := exporter.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(log.Runs) != 1 {
		t.Fatalf("Expected 1 run for package vulns, got %d", len(log.Runs))
	}

	run := log.Runs[0]
	if run.Tool.Driver.Name != "zero-package-vulns" {
		t.Errorf("Expected tool name zero-package-vulns, got %q", run.Tool.Driver.Name)
	}
	if len(run.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(run.Results))
	}

	result := run.Results[0]
	if result.RuleID != "CVE-2021-12345" {
		t.Errorf("Expected rule ID CVE-2021-12345, got %q", result.RuleID)
	}
	if result.Level != "error" {
		t.Errorf("High severity should map to error level, got %q", result.Level)
	}

	// Check logical location
	if len(result.Locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(result.Locations))
	}
	if len(result.Locations[0].LogicalLocations) != 1 {
		t.Fatalf("Expected 1 logical location, got %d", len(result.Locations[0].LogicalLocations))
	}
	logLoc := result.Locations[0].LogicalLocations[0]
	if logLoc.Name != "lodash" {
		t.Errorf("Expected package name lodash, got %q", logLoc.Name)
	}
	if logLoc.Kind != "package" {
		t.Errorf("Expected kind package, got %q", logLoc.Kind)
	}
}

func TestExporterCrypto(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sarif-export-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a crypto.json file
	cryptoData := `{
		"findings": {
			"ciphers": [
				{
					"algorithm": "DES",
					"file": "crypto.go",
					"line": 25,
					"severity": "high",
					"suggestion": "Use AES-256-GCM instead"
				}
			],
			"keys": [
				{
					"type": "RSA-1024",
					"file": "keys.go",
					"line": 15,
					"severity": "critical"
				}
			]
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "crypto.json"), []byte(cryptoData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	exporter := NewExporter(tmpDir, tmpDir)
	log, err := exporter.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(log.Runs) != 1 {
		t.Fatalf("Expected 1 run for crypto, got %d", len(log.Runs))
	}

	run := log.Runs[0]
	if run.Tool.Driver.Name != "zero-crypto" {
		t.Errorf("Expected tool name zero-crypto, got %q", run.Tool.Driver.Name)
	}
	if len(run.Results) != 2 {
		t.Errorf("Expected 2 results (cipher + key), got %d", len(run.Results))
	}
}

func TestExporterCodeQuality(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sarif-export-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a code-quality.json file
	qualityData := `{
		"findings": {
			"tech_debt": {
				"markers": [
					{
						"type": "TODO",
						"file": "main.go",
						"line": 50,
						"message": "TODO: implement error handling",
						"priority": "medium"
					}
				]
			},
			"complexity": [
				{
					"file": "handler.go",
					"function": "ProcessRequest",
					"line": 100,
					"complexity": 25,
					"type": "cyclomatic"
				}
			]
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "code-quality.json"), []byte(qualityData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	exporter := NewExporter(tmpDir, tmpDir)
	log, err := exporter.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(log.Runs) != 1 {
		t.Fatalf("Expected 1 run for code quality, got %d", len(log.Runs))
	}

	run := log.Runs[0]
	if run.Tool.Driver.Name != "zero-code-quality" {
		t.Errorf("Expected tool name zero-code-quality, got %q", run.Tool.Driver.Name)
	}
	if len(run.Results) != 2 {
		t.Errorf("Expected 2 results (marker + complexity), got %d", len(run.Results))
	}

	// Check that tech debt markers are level "note"
	for _, result := range run.Results {
		if result.RuleID == "quality/tech-debt/TODO" {
			if result.Level != "note" {
				t.Errorf("Tech debt markers should be level note, got %q", result.Level)
			}
		}
		if result.RuleID == "quality/complexity/cyclomatic" {
			if result.Level != "warning" {
				t.Errorf("Complexity issues should be level warning, got %q", result.Level)
			}
		}
	}
}

func TestLogJSON(t *testing.T) {
	log := NewLog()
	run := NewRun("test-tool", "1.0.0", "https://example.com")
	run.AddRule("test-rule", "Test", "Description", "", "error")
	run.AddResult("test-rule", 0, "error", "Test message", "test.go", 1)
	log.Runs = append(log.Runs, *run)

	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("Failed to marshal log: %v", err)
	}

	// Verify it's valid JSON and has expected structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed["$schema"] != SchemaURI {
		t.Errorf("$schema = %v, want %v", parsed["$schema"], SchemaURI)
	}
	if parsed["version"] != SARIFVersion {
		t.Errorf("version = %v, want %v", parsed["version"], SARIFVersion)
	}

	runs, ok := parsed["runs"].([]interface{})
	if !ok || len(runs) != 1 {
		t.Error("Expected runs array with 1 element")
	}
}
