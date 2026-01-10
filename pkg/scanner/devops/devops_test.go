package devops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDevOpsScanner_Name(t *testing.T) {
	s := &DevOpsScanner{}
	if s.Name() != "devops" {
		t.Errorf("Name() = %q, want %q", s.Name(), "devops")
	}
}

func TestDevOpsScanner_Description(t *testing.T) {
	s := &DevOpsScanner{}
	desc := s.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestDevOpsScanner_Dependencies(t *testing.T) {
	s := &DevOpsScanner{}
	deps := s.Dependencies()
	if deps != nil {
		t.Errorf("Dependencies() = %v, want nil", deps)
	}
}

func TestDevOpsScanner_EstimateDuration(t *testing.T) {
	s := &DevOpsScanner{}

	tests := []struct {
		fileCount int
		wantMin   int
	}{
		{0, 30},
		{500, 30},
		{1000, 30},
		{5000, 30},
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

	if !cfg.IaC.Enabled {
		t.Error("IaC should be enabled by default")
	}
	if !cfg.Containers.Enabled {
		t.Error("Containers should be enabled by default")
	}
	if !cfg.GitHubActions.Enabled {
		t.Error("GitHubActions should be enabled by default")
	}
	if !cfg.DORA.Enabled {
		t.Error("DORA should be enabled by default")
	}
	if !cfg.Git.Enabled {
		t.Error("Git should be enabled by default")
	}
	if cfg.IaC.Tool != "auto" {
		t.Errorf("IaC.Tool = %q, want %q", cfg.IaC.Tool, "auto")
	}
	if cfg.DORA.PeriodDays != 90 {
		t.Errorf("DORA.PeriodDays = %d, want 90", cfg.DORA.PeriodDays)
	}
}

func TestQuickConfig(t *testing.T) {
	cfg := QuickConfig()

	if !cfg.IaC.Enabled {
		t.Error("IaC should be enabled in quick config")
	}
	if cfg.Containers.Enabled {
		t.Error("Containers should be disabled in quick config")
	}
	if !cfg.GitHubActions.Enabled {
		t.Error("GitHubActions should be enabled in quick config")
	}
	if cfg.Git.IncludeChurn {
		t.Error("Git.IncludeChurn should be disabled in quick config")
	}
	if cfg.Git.IncludeAge {
		t.Error("Git.IncludeAge should be disabled in quick config")
	}
}

func TestSecurityConfig(t *testing.T) {
	cfg := SecurityConfig()

	if !cfg.IaC.Enabled {
		t.Error("IaC should be enabled in security config")
	}
	if !cfg.Containers.Enabled {
		t.Error("Containers should be enabled in security config")
	}
	if !cfg.GitHubActions.Enabled {
		t.Error("GitHubActions should be enabled in security config")
	}
	if cfg.DORA.Enabled {
		t.Error("DORA should be disabled in security config")
	}
	if cfg.Git.Enabled {
		t.Error("Git should be disabled in security config")
	}
}

func TestNormalizeIaCType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"terraform", "terraform"},
		{"TERRAFORM", "terraform"},
		{"kubernetes", "kubernetes"},
		{"Kubernetes", "kubernetes"},
		{"k8s", "kubernetes"},
		{"K8S", "kubernetes"},
		{"dockerfile", "dockerfile"},
		{"Dockerfile", "dockerfile"},
		{"docker", "dockerfile"},
		{"cloudformation", "cloudformation"},
		{"CloudFormation", "cloudformation"},
		{"cfn", "cloudformation"},
		{"helm", "helm"},
		{"azure", "azure"},
		{"arm", "azure"},
		{"unknown_type", "unknown_type"},
	}

	for _, tt := range tests {
		got := normalizeIaCType(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeIaCType(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDeriveCheckovSeverity(t *testing.T) {
	tests := []struct {
		checkID  string
		expected string
	}{
		{"CKV_AWS_PUBLIC_ACL", "critical"},
		{"CKV_ENCRYPT_DATA", "critical"},
		{"CKV_PRIVILEGED_CONTAINER", "critical"},
		{"CKV_ROOT_USER", "critical"},
		{"CKV_ADMIN_POLICY", "critical"},
		{"CKV_AUTH_BYPASS", "high"},
		{"CKV_SECRET_EXPOSED", "high"},
		{"CKV_PASSWORD_HARDCODED", "high"},
		{"CKV_CREDENTIAL_LEAK", "high"},
		{"CKV_API_KEY_EXPOSED", "high"},
		{"CKV_TOKEN_IN_CODE", "high"},
		{"CKV_LOG_DISABLED", "medium"},
		{"CKV_MONITOR_OFF", "medium"},
		{"CKV_BACKUP_MISSING", "medium"},
		{"CKV_VERSION_OLD", "medium"},
		{"CKV_SSL_DISABLED", "medium"},
		{"CKV_TLS_V1", "medium"},
		{"CKV_UNKNOWN_CHECK", "low"},
	}

	for _, tt := range tests {
		got := deriveCheckovSeverity(tt.checkID)
		if got != tt.expected {
			t.Errorf("deriveCheckovSeverity(%q) = %q, want %q", tt.checkID, got, tt.expected)
		}
	}
}

func TestClassifyDeploymentFrequency(t *testing.T) {
	tests := []struct {
		freq     float64
		expected string
	}{
		{10.0, "elite"},
		{7.0, "elite"},
		{5.0, "high"},
		{1.0, "high"},
		{0.5, "medium"},
		{0.25, "medium"},
		{0.1, "low"},
		{0.0, "low"},
	}

	for _, tt := range tests {
		got := classifyDeploymentFrequency(tt.freq)
		if got != tt.expected {
			t.Errorf("classifyDeploymentFrequency(%v) = %q, want %q", tt.freq, got, tt.expected)
		}
	}
}

func TestClassifyLeadTime(t *testing.T) {
	tests := []struct {
		hours    float64
		expected string
	}{
		{12.0, "elite"},
		{23.9, "elite"},
		{24.0, "high"},
		{100.0, "high"},
		{167.9, "high"},
		{168.0, "medium"},
		{500.0, "medium"},
		{719.9, "medium"},
		{720.0, "low"},
		{1000.0, "low"},
	}

	for _, tt := range tests {
		got := classifyLeadTime(tt.hours)
		if got != tt.expected {
			t.Errorf("classifyLeadTime(%v) = %q, want %q", tt.hours, got, tt.expected)
		}
	}
}

func TestClassifyChangeFailureRate(t *testing.T) {
	tests := []struct {
		rate     float64
		expected string
	}{
		{0.0, "elite"},
		{5.0, "elite"},
		{6.0, "high"},
		{10.0, "high"},
		{11.0, "medium"},
		{15.0, "medium"},
		{16.0, "low"},
		{50.0, "low"},
	}

	for _, tt := range tests {
		got := classifyChangeFailureRate(tt.rate)
		if got != tt.expected {
			t.Errorf("classifyChangeFailureRate(%v) = %q, want %q", tt.rate, got, tt.expected)
		}
	}
}

func TestClassifyMTTR(t *testing.T) {
	tests := []struct {
		hours    float64
		expected string
	}{
		{0.5, "elite"},
		{0.9, "elite"},
		{1.0, "high"},
		{12.0, "high"},
		{23.9, "high"},
		{24.0, "medium"},
		{100.0, "medium"},
		{167.9, "medium"},
		{168.0, "low"},
		{500.0, "low"},
	}

	for _, tt := range tests {
		got := classifyMTTR(tt.hours)
		if got != tt.expected {
			t.Errorf("classifyMTTR(%v) = %q, want %q", tt.hours, got, tt.expected)
		}
	}
}

func TestIsSourceFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"main.py", true},
		{"app.js", true},
		{"component.ts", true},
		{"component.tsx", true},
		{"component.jsx", true},
		{"App.java", true},
		{"main.go", true},
		{"script.rb", true},
		{"index.php", true},
		{"main.c", true},
		{"lib.cpp", true},
		{"main.rs", true},
		{"app.swift", true},
		{"Main.kt", true},
		{"README.md", false},
		{"package.json", false},
		{"Makefile", false},
		{".gitignore", false},
		{"data.csv", false},
	}

	for _, tt := range tests {
		got := isSourceFile(tt.filename)
		if got != tt.expected {
			t.Errorf("isSourceFile(%q) = %v, want %v", tt.filename, got, tt.expected)
		}
	}
}

func TestCalculateBusFactor(t *testing.T) {
	tests := []struct {
		name         string
		contributors []Contributor
		expected     int
	}{
		{
			name:         "empty",
			contributors: []Contributor{},
			expected:     0,
		},
		{
			name: "single dominant contributor",
			contributors: []Contributor{
				{TotalCommits: 100},
			},
			expected: 1,
		},
		{
			name: "two equal contributors",
			contributors: []Contributor{
				{TotalCommits: 50},
				{TotalCommits: 50},
			},
			expected: 1, // First contributor reaches 50% threshold
		},
		{
			name: "three contributors with majority by one",
			contributors: []Contributor{
				{TotalCommits: 60},
				{TotalCommits: 30},
				{TotalCommits: 10},
			},
			expected: 1,
		},
		{
			name: "three even contributors",
			contributors: []Contributor{
				{TotalCommits: 34},
				{TotalCommits: 33},
				{TotalCommits: 33},
			},
			expected: 2, // Need two contributors to reach 50%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBusFactor(tt.contributors)
			if got != tt.expected {
				t.Errorf("calculateBusFactor() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCountActiveContributors(t *testing.T) {
	contributors := []Contributor{
		{Commits30d: 10, Commits90d: 20},
		{Commits30d: 0, Commits90d: 15},
		{Commits30d: 5, Commits90d: 5},
		{Commits30d: 0, Commits90d: 0},
	}

	tests := []struct {
		period   string
		expected int
	}{
		{"30d", 2},
		{"90d", 3},
	}

	for _, tt := range tests {
		got := countActiveContributors(contributors, tt.period)
		if got != tt.expected {
			t.Errorf("countActiveContributors(period=%q) = %d, want %d", tt.period, got, tt.expected)
		}
	}
}

func TestFindDockerfiles(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "dockerfile-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test Dockerfiles
	testFiles := []string{
		"Dockerfile",
		"Dockerfile.dev",
		"app.Dockerfile",
		"subdir/Dockerfile",
	}

	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		os.MkdirAll(filepath.Dir(path), 0755)
		if err := os.WriteFile(path, []byte("FROM alpine:latest"), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", f, err)
		}
	}

	// Create files that should be ignored
	os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "node_modules/Dockerfile"), []byte("FROM node"), 0644)

	dockerfiles := findDockerfiles(tmpDir)

	// Should find 4 dockerfiles, exclude node_modules
	if len(dockerfiles) != 4 {
		t.Errorf("findDockerfiles() found %d files, want 4", len(dockerfiles))
	}
}

func TestExtractBaseImages(t *testing.T) {
	// Create temp directory with test Dockerfiles
	tmpDir, err := os.MkdirTemp("", "baseimage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create Dockerfile with multiple FROM statements
	dockerfile1 := `FROM alpine:3.18
RUN apk add --no-cache git
FROM node:18-alpine AS builder
COPY . .
FROM python:3.11-slim
CMD ["python", "app.py"]
`
	path1 := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(path1, []byte(dockerfile1), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	// Dockerfile with variable (should be skipped)
	dockerfile2 := `FROM $BASE_IMAGE
RUN echo "hello"
FROM scratch
`
	path2 := filepath.Join(tmpDir, "Dockerfile.var")
	if err := os.WriteFile(path2, []byte(dockerfile2), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile.var: %v", err)
	}

	dockerfiles := []string{path1, path2}
	images := extractBaseImages(dockerfiles)

	// Should find: alpine:3.18, node:18-alpine, python:3.11-slim
	// Should skip: $BASE_IMAGE (variable), scratch
	if len(images) != 3 {
		t.Errorf("extractBaseImages() found %d images, want 3", len(images))
	}

	// Verify specific images
	expectedImages := map[string]bool{
		"alpine:3.18":     false,
		"node:18-alpine":  false,
		"python:3.11-slim": false,
	}

	for _, img := range images {
		expectedImages[img.Image] = true
	}

	for img, found := range expectedImages {
		if !found {
			t.Errorf("Expected image %q not found", img)
		}
	}
}

func TestScanWorkflowFile(t *testing.T) {
	// Create temp directory with test workflow file
	tmpDir, err := os.MkdirTemp("", "workflow-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	workflowDir := filepath.Join(tmpDir, ".github", "workflows")
	os.MkdirAll(workflowDir, 0755)

	// Create workflow with various security issues
	workflow := `name: CI

on: push

permissions: write-all

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@main
      - name: Run with issue
        run: echo "${{ github.event.issue.title }}"
        env:
          SECRET: ${{ secrets.MY_SECRET }}
`
	workflowPath := filepath.Join(workflowDir, "ci.yml")
	if err := os.WriteFile(workflowPath, []byte(workflow), 0644); err != nil {
		t.Fatalf("Failed to write workflow: %v", err)
	}

	cfg := GitHubActionsConfig{
		Enabled:          true,
		CheckPinning:     true,
		CheckSecrets:     true,
		CheckInjection:   true,
		CheckPermissions: true,
	}

	findings := scanWorkflowFile(workflowPath, tmpDir, cfg)

	// Should find: unpinned action (actions/setup-node@main), injection risk, excessive permissions
	if len(findings) < 2 {
		t.Errorf("scanWorkflowFile() found %d findings, want at least 2", len(findings))
	}

	// Verify we found the expected categories
	foundCategories := make(map[string]bool)
	for _, f := range findings {
		foundCategories[f.Category] = true
	}

	if !foundCategories["unpinned-action"] {
		t.Error("Expected to find unpinned-action finding")
	}
	if !foundCategories["excessive-permissions"] {
		t.Error("Expected to find excessive-permissions finding")
	}
	if !foundCategories["injection-risk"] {
		t.Error("Expected to find injection-risk finding")
	}
}

func TestParseCheckovOutput(t *testing.T) {
	// Valid checkov output (array format)
	validOutput := []byte(`[{
		"check_type": "terraform",
		"results": {
			"failed_checks": [
				{
					"check_id": "CKV_AWS_PUBLIC_ACL",
					"file_path": "/terraform/main.tf",
					"file_line_range": [10, 15],
					"resource": "aws_s3_bucket.public",
					"guideline": "Disable public ACL",
					"description": "S3 bucket has public ACL enabled"
				}
			]
		}
	}]`)

	findings, summary := parseCheckovOutput(validOutput, "/repo")

	if len(findings) != 1 {
		t.Errorf("parseCheckovOutput() returned %d findings, want 1", len(findings))
	}

	if summary.TotalFindings != 1 {
		t.Errorf("summary.TotalFindings = %d, want 1", summary.TotalFindings)
	}

	if summary.Tool != "checkov" {
		t.Errorf("summary.Tool = %q, want %q", summary.Tool, "checkov")
	}

	if summary.ByType["terraform"] != 1 {
		t.Errorf("summary.ByType[terraform] = %d, want 1", summary.ByType["terraform"])
	}

	// Invalid JSON
	invalidOutput := []byte(`not json`)
	findings2, summary2 := parseCheckovOutput(invalidOutput, "/repo")
	if len(findings2) != 0 {
		t.Errorf("parseCheckovOutput(invalid) returned %d findings, want 0", len(findings2))
	}
	if summary2.TotalFindings != 0 {
		t.Errorf("summary2.TotalFindings = %d, want 0", summary2.TotalFindings)
	}
}

func TestParseTrivyIaCOutput(t *testing.T) {
	// Valid trivy output
	validOutput := []byte(`{
		"Results": [
			{
				"Target": "terraform/main.tf",
				"Type": "terraform",
				"Misconfigurations": [
					{
						"ID": "AVD-AWS-0001",
						"Title": "S3 bucket encryption",
						"Description": "S3 bucket should have encryption enabled",
						"Resolution": "Enable encryption",
						"Severity": "HIGH",
						"Status": "FAIL",
						"CauseMetadata": {
							"Resource": "aws_s3_bucket.data",
							"StartLine": 25
						}
					},
					{
						"ID": "AVD-AWS-0002",
						"Title": "Passed check",
						"Severity": "LOW",
						"Status": "PASS"
					}
				]
			}
		]
	}`)

	findings, summary := parseTrivyIaCOutput(validOutput, "/repo")

	// Should only include FAIL status, not PASS
	if len(findings) != 1 {
		t.Errorf("parseTrivyIaCOutput() returned %d findings, want 1", len(findings))
	}

	if summary.TotalFindings != 1 {
		t.Errorf("summary.TotalFindings = %d, want 1", summary.TotalFindings)
	}

	if summary.High != 1 {
		t.Errorf("summary.High = %d, want 1", summary.High)
	}

	if summary.Tool != "trivy" {
		t.Errorf("summary.Tool = %q, want %q", summary.Tool, "trivy")
	}
}

func TestParseTrivyImageOutput(t *testing.T) {
	// Valid trivy image output
	validOutput := []byte(`{
		"Results": [
			{
				"Target": "alpine:3.18 (alpine 3.18)",
				"Vulnerabilities": [
					{
						"VulnerabilityID": "CVE-2024-1234",
						"PkgName": "openssl",
						"InstalledVersion": "1.1.1",
						"FixedVersion": "1.1.2",
						"Title": "OpenSSL vulnerability",
						"Description": "A vulnerability in OpenSSL",
						"Severity": "CRITICAL",
						"References": ["https://nvd.nist.gov/vuln/detail/CVE-2024-1234"],
						"CVSS": {
							"nvd": {"V3Score": 9.8}
						}
					}
				]
			}
		]
	}`)

	imgRef := imageRef{
		Image:      "alpine:3.18",
		Dockerfile: "/app/Dockerfile",
		Line:       1,
	}

	findings := parseTrivyImageOutput(validOutput, imgRef)

	if len(findings) != 1 {
		t.Errorf("parseTrivyImageOutput() returned %d findings, want 1", len(findings))
	}

	if findings[0].VulnID != "CVE-2024-1234" {
		t.Errorf("findings[0].VulnID = %q, want %q", findings[0].VulnID, "CVE-2024-1234")
	}

	if findings[0].Severity != "critical" {
		t.Errorf("findings[0].Severity = %q, want %q", findings[0].Severity, "critical")
	}

	if findings[0].CVSS != 9.8 {
		t.Errorf("findings[0].CVSS = %v, want %v", findings[0].CVSS, 9.8)
	}
}

func TestGHAPatterns(t *testing.T) {
	// Verify patterns exist and are properly configured
	if len(ghaPatterns) == 0 {
		t.Error("ghaPatterns should not be empty")
	}

	// Verify each pattern has required fields
	for i, pat := range ghaPatterns {
		if pat.pattern == nil {
			t.Errorf("ghaPatterns[%d].pattern is nil", i)
		}
		if pat.category == "" {
			t.Errorf("ghaPatterns[%d].category is empty", i)
		}
		if pat.severity == "" {
			t.Errorf("ghaPatterns[%d].severity is empty", i)
		}
		if pat.title == "" {
			t.Errorf("ghaPatterns[%d].title is empty", i)
		}
		if pat.enabled == nil {
			t.Errorf("ghaPatterns[%d].enabled is nil", i)
		}
	}

	// Test specific patterns
	tests := []struct {
		input        string
		wantCategory string
	}{
		{"      - uses: actions/checkout@main", "unpinned-action"},
		{"      - uses: actions/setup-node@v3", "unpinned-action"},
		{"permissions: write-all", "excessive-permissions"},
		{"        run: echo ${{ github.event.issue.body }}", "injection-risk"},
	}

	cfg := GitHubActionsConfig{
		CheckPinning:     true,
		CheckSecrets:     true,
		CheckInjection:   true,
		CheckPermissions: true,
	}

	for _, tt := range tests {
		found := false
		for _, pat := range ghaPatterns {
			if !pat.enabled(cfg) {
				continue
			}
			if pat.pattern.MatchString(tt.input) && pat.category == tt.wantCategory {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Pattern for %q not matched (expected category %q)", tt.input, tt.wantCategory)
		}
	}
}

// ============================================================================
// Phase 3: PR-Level Metrics Tests (LinearB alignment)
// ============================================================================

func TestClassifyPickupTime(t *testing.T) {
	tests := []struct {
		hours    float64
		expected string
	}{
		{0.5, "elite"},
		{0.9, "elite"},
		{1.0, "good"},
		{2.0, "good"},
		{4.0, "good"},
		{5.0, "fair"},
		{10.0, "fair"},
		{16.0, "fair"},
		{17.0, "needs_focus"},
		{24.0, "needs_focus"},
		{100.0, "needs_focus"},
	}

	for _, tt := range tests {
		got := classifyPickupTime(tt.hours)
		if got != tt.expected {
			t.Errorf("classifyPickupTime(%v) = %q, want %q", tt.hours, got, tt.expected)
		}
	}
}

func TestClassifyReviewTime(t *testing.T) {
	tests := []struct {
		hours    float64
		expected string
	}{
		{1.0, "elite"},
		{2.9, "elite"},
		{3.0, "good"},
		{10.0, "good"},
		{14.0, "good"},
		{15.0, "fair"},
		{20.0, "fair"},
		{24.0, "fair"},
		{25.0, "needs_focus"},
		{48.0, "needs_focus"},
	}

	for _, tt := range tests {
		got := classifyReviewTime(tt.hours)
		if got != tt.expected {
			t.Errorf("classifyReviewTime(%v) = %q, want %q", tt.hours, got, tt.expected)
		}
	}
}

func TestClassifyMergeTime(t *testing.T) {
	tests := []struct {
		hours    float64
		expected string
	}{
		{0.5, "elite"},
		{0.9, "elite"},
		{1.0, "good"},
		{2.0, "good"},
		{3.0, "good"},
		{4.0, "fair"},
		{10.0, "fair"},
		{16.0, "fair"},
		{17.0, "needs_focus"},
		{24.0, "needs_focus"},
	}

	for _, tt := range tests {
		got := classifyMergeTime(tt.hours)
		if got != tt.expected {
			t.Errorf("classifyMergeTime(%v) = %q, want %q", tt.hours, got, tt.expected)
		}
	}
}

func TestClassifyPRSize(t *testing.T) {
	tests := []struct {
		size     int
		expected string
	}{
		{50, "elite"},
		{99, "elite"},
		{100, "good"},
		{150, "good"},
		{155, "good"},
		{156, "fair"},
		{200, "fair"},
		{228, "fair"},
		{229, "needs_focus"},
		{500, "needs_focus"},
	}

	for _, tt := range tests {
		got := classifyPRSize(tt.size)
		if got != tt.expected {
			t.Errorf("classifyPRSize(%d) = %q, want %q", tt.size, got, tt.expected)
		}
	}
}

// ============================================================================
// Phase 3: Rework Rate Tests (DORA 2025)
// ============================================================================

func TestClassifyReworkRate(t *testing.T) {
	tests := []struct {
		rate     float64
		expected string
	}{
		{0.0, "elite"},
		{2.0, "elite"},
		{2.9, "elite"},
		{3.0, "good"},
		{4.0, "good"},
		{5.0, "good"},
		{6.0, "fair"},
		{7.0, "fair"},
		{8.0, "fair"},
		{9.0, "needs_focus"},
		{15.0, "needs_focus"},
	}

	for _, tt := range tests {
		got := classifyReworkRate(tt.rate)
		if got != tt.expected {
			t.Errorf("classifyReworkRate(%v) = %q, want %q", tt.rate, got, tt.expected)
		}
	}
}

func TestClassifyRefactorRate(t *testing.T) {
	tests := []struct {
		rate     float64
		expected string
	}{
		{5.0, "elite"},
		{10.0, "elite"},
		{10.9, "elite"},
		{11.0, "good"},
		{14.0, "good"},
		{16.0, "good"},
		{17.0, "fair"},
		{20.0, "fair"},
		{22.0, "fair"},
		{23.0, "needs_focus"},
		{30.0, "needs_focus"},
	}

	for _, tt := range tests {
		got := classifyRefactorRate(tt.rate)
		if got != tt.expected {
			t.Errorf("classifyRefactorRate(%v) = %q, want %q", tt.rate, got, tt.expected)
		}
	}
}

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		url           string
		expectedOwner string
		expectedRepo  string
	}{
		// HTTPS format
		{"https://github.com/owner/repo.git", "owner", "repo"},
		{"https://github.com/owner/repo", "owner", "repo"},
		{"https://github.com/expressjs/express.git", "expressjs", "express"},
		// SSH format
		{"git@github.com:owner/repo.git", "owner", "repo"},
		{"git@github.com:owner/repo", "owner", "repo"},
		{"git@github.com:crashappsec/zero.git", "crashappsec", "zero"},
		// Invalid URLs
		{"https://gitlab.com/owner/repo.git", "", ""},
		{"not-a-url", "", ""},
	}

	for _, tt := range tests {
		owner, repo := parseGitHubURL(tt.url)
		if owner != tt.expectedOwner || repo != tt.expectedRepo {
			t.Errorf("parseGitHubURL(%q) = (%q, %q), want (%q, %q)",
				tt.url, owner, repo, tt.expectedOwner, tt.expectedRepo)
		}
	}
}

func TestDORAConfigPRMetrics(t *testing.T) {
	cfg := DefaultConfig()

	// Default should have PR metrics enabled
	if !cfg.DORA.IncludePRMetrics {
		t.Error("DORA.IncludePRMetrics should be enabled by default")
	}
	if cfg.DORA.MaxPRs != 100 {
		t.Errorf("DORA.MaxPRs = %d, want 100", cfg.DORA.MaxPRs)
	}
	if !cfg.DORA.IncludeReworkRate {
		t.Error("DORA.IncludeReworkRate should be enabled by default")
	}

	// Quick config should have PR metrics disabled
	quickCfg := QuickConfig()
	if quickCfg.DORA.IncludePRMetrics {
		t.Error("DORA.IncludePRMetrics should be disabled in quick config")
	}
	if quickCfg.DORA.IncludeReworkRate {
		t.Error("DORA.IncludeReworkRate should be disabled in quick config")
	}

	// Full config should have everything enabled
	fullCfg := FullConfig()
	if !fullCfg.DORA.IncludePRMetrics {
		t.Error("DORA.IncludePRMetrics should be enabled in full config")
	}
	if fullCfg.DORA.MaxPRs != 200 {
		t.Errorf("FullConfig DORA.MaxPRs = %d, want 200", fullCfg.DORA.MaxPRs)
	}
	if !fullCfg.DORA.IncludeReworkRate {
		t.Error("DORA.IncludeReworkRate should be enabled in full config")
	}
}
