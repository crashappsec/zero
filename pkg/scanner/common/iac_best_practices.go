// Package common provides shared utilities for scanners
// This file provides IaC best practices scanning using RAG-generated Semgrep rules
package common

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// IaCBestPracticesScanner scans IaC files for best practices violations
type IaCBestPracticesScanner struct {
	ragPath   string
	cacheDir  string
	timeout   time.Duration
	onStatus  func(string)
}

// IaCBestPracticesConfig configures the IaC best practices scanner
type IaCBestPracticesConfig struct {
	RAGPath  string
	CacheDir string
	Timeout  time.Duration
	OnStatus func(string)
}

// IaCBestPracticeFinding represents a best practice violation in an IaC file
type IaCBestPracticeFinding struct {
	RuleID      string
	Type        string // terraform, kubernetes, cloudformation, helm, dockerfile
	Category    string // always "best-practice"
	Severity    string // medium, low, info
	Title       string
	Message     string
	File        string
	Line        int
	Column      int
	Remediation string
}

// IaCBestPracticesResult holds the result of scanning
type IaCBestPracticesResult struct {
	Findings []IaCBestPracticeFinding
	Summary  IaCBestPracticesSummary
	Error    error
}

// IaCBestPracticesSummary provides summary statistics
type IaCBestPracticesSummary struct {
	TotalFindings int
	ByType        map[string]int // by IaC type
	BySeverity    map[string]int
	FilesScanned  int
}

// NewIaCBestPracticesScanner creates a new IaC best practices scanner
func NewIaCBestPracticesScanner(cfg IaCBestPracticesConfig) *IaCBestPracticesScanner {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}
	if cfg.OnStatus == nil {
		cfg.OnStatus = func(string) {}
	}
	if cfg.RAGPath == "" {
		cfg.RAGPath = findRAGPath()
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = getCacheDir()
	}

	return &IaCBestPracticesScanner{
		ragPath:  cfg.RAGPath,
		cacheDir: cfg.CacheDir,
		timeout:  cfg.Timeout,
		onStatus: cfg.OnStatus,
	}
}

// Scan scans a repository for best practices violations in IaC files
func (s *IaCBestPracticesScanner) Scan(ctx context.Context, repoPath string) *IaCBestPracticesResult {
	result := &IaCBestPracticesResult{
		Summary: IaCBestPracticesSummary{
			ByType:     make(map[string]int),
			BySeverity: make(map[string]int),
		},
	}

	// Find IaC files (reuse logic from secrets scanner)
	iacFiles := s.findIaCFiles(repoPath)
	if len(iacFiles) == 0 {
		return result
	}

	result.Summary.FilesScanned = len(iacFiles)
	s.onStatus("Scanning IaC files for best practices")

	// Generate rules if needed
	rulesPath := filepath.Join(s.cacheDir, "rules", "iac-best-practices.yaml")
	if err := s.ensureRulesGenerated(rulesPath); err != nil {
		// Fall back to structural checks
		result.Findings = s.scanWithStructuralChecks(ctx, iacFiles, repoPath)
	} else {
		// Use Semgrep with generated rules
		result.Findings = s.scanWithSemgrep(ctx, iacFiles, repoPath, rulesPath)
	}

	// Build summary
	for _, f := range result.Findings {
		result.Summary.TotalFindings++
		result.Summary.ByType[f.Type]++
		result.Summary.BySeverity[f.Severity]++
	}

	return result
}

// findIaCFiles finds all IaC files in the repository
func (s *IaCBestPracticesScanner) findIaCFiles(repoPath string) []string {
	var files []string

	iacExtensions := map[string]bool{
		".tf":         true, // Terraform
		".tfvars":     true, // Terraform variables
		".yaml":       true, // K8s, CloudFormation, Helm
		".yml":        true, // K8s, CloudFormation, Helm
		".json":       true, // CloudFormation, K8s
		"Dockerfile":  true, // Docker
	}

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip common non-IaC directories
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == "vendor" || name == ".git" ||
				name == "__pycache__" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for Dockerfile
		if info.Name() == "Dockerfile" || strings.HasPrefix(info.Name(), "Dockerfile.") {
			files = append(files, path)
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		// Check extension
		if iacExtensions[ext] {
			if ext == ".yaml" || ext == ".yml" || ext == ".json" {
				if s.isIaCFile(path) {
					files = append(files, path)
				}
			} else {
				files = append(files, path)
			}
		}

		return nil
	})

	return files
}

// isIaCFile checks if a YAML/JSON file is an IaC file
func (s *IaCBestPracticesScanner) isIaCFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	content := string(data)

	// Kubernetes markers
	if strings.Contains(content, "apiVersion:") && strings.Contains(content, "kind:") {
		return true
	}

	// CloudFormation markers
	if strings.Contains(content, "AWSTemplateFormatVersion") ||
		(strings.Contains(content, "Resources:") && strings.Contains(content, "Type: AWS::")) {
		return true
	}

	// Helm Chart.yaml
	if strings.Contains(filepath.Base(path), "Chart") {
		return true
	}

	// Helm values
	if strings.Contains(filepath.Base(path), "values") {
		return true
	}

	return false
}

// ensureRulesGenerated generates Semgrep rules from RAG if needed
func (s *IaCBestPracticesScanner) ensureRulesGenerated(rulesPath string) error {
	// Check if rules exist and are recent
	if info, err := os.Stat(rulesPath); err == nil {
		if time.Since(info.ModTime()) < 24*time.Hour {
			return nil // Rules are fresh
		}
	}

	if s.ragPath == "" {
		return os.ErrNotExist
	}

	// Generate rules from RAG patterns
	return GenerateRulesFromRAG(s.ragPath, "devops/iac-best-practices", rulesPath)
}

// scanWithSemgrep uses Semgrep with generated rules
func (s *IaCBestPracticesScanner) scanWithSemgrep(ctx context.Context, files []string, repoPath, rulesPath string) []IaCBestPracticeFinding {
	var findings []IaCBestPracticeFinding

	if !HasSemgrep() {
		return s.scanWithStructuralChecks(ctx, files, repoPath)
	}

	runner := NewSemgrepRunner(SemgrepConfig{
		RulePaths: []string{rulesPath},
		Timeout:   s.timeout,
		OnStatus:  s.onStatus,
	})

	result := runner.RunOnFiles(ctx, files, repoPath)
	if result.Error != nil {
		return s.scanWithStructuralChecks(ctx, files, repoPath)
	}

	for _, f := range result.Findings {
		finding := IaCBestPracticeFinding{
			RuleID:      f.RuleID,
			Type:        s.determineIaCType(f.File),
			Category:    "best-practice",
			Severity:    f.Severity,
			Title:       f.Message,
			Message:     f.Message,
			File:        f.File,
			Line:        f.Line,
			Column:      f.Column,
			Remediation: f.Remediation,
		}
		findings = append(findings, finding)
	}

	return findings
}

// scanWithStructuralChecks performs basic structural checks as fallback
func (s *IaCBestPracticesScanner) scanWithStructuralChecks(ctx context.Context, files []string, repoPath string) []IaCBestPracticeFinding {
	var findings []IaCBestPracticeFinding

	for _, file := range files {
		select {
		case <-ctx.Done():
			return findings
		default:
		}

		fileFindings := s.checkFileStructure(file)
		findings = append(findings, fileFindings...)
	}

	return findings
}

// checkFileStructure performs structural checks on a single file
func (s *IaCBestPracticesScanner) checkFileStructure(filePath string) []IaCBestPracticeFinding {
	var findings []IaCBestPracticeFinding

	data, err := os.ReadFile(filePath)
	if err != nil {
		return findings
	}

	content := string(data)
	fileType := s.determineIaCType(filePath)

	switch fileType {
	case "dockerfile":
		findings = append(findings, s.checkDockerfile(filePath, content)...)
	case "kubernetes":
		findings = append(findings, s.checkKubernetes(filePath, content)...)
	case "terraform":
		findings = append(findings, s.checkTerraform(filePath, content)...)
	}

	return findings
}

// checkDockerfile checks Dockerfile for best practices
func (s *IaCBestPracticesScanner) checkDockerfile(filePath, content string) []IaCBestPracticeFinding {
	var findings []IaCBestPracticeFinding

	// Check for missing HEALTHCHECK
	if !strings.Contains(content, "HEALTHCHECK") {
		findings = append(findings, IaCBestPracticeFinding{
			RuleID:      "dockerfile-missing-healthcheck",
			Type:        "dockerfile",
			Category:    "best-practice",
			Severity:    "low",
			Title:       "Missing HEALTHCHECK instruction",
			Message:     "Dockerfile should include HEALTHCHECK for container health monitoring",
			File:        filePath,
			Remediation: "Add HEALTHCHECK instruction: HEALTHCHECK CMD curl -f http://localhost/ || exit 1",
		})
	}

	// Check for using latest tag
	if strings.Contains(content, ":latest") {
		findings = append(findings, IaCBestPracticeFinding{
			RuleID:      "dockerfile-latest-tag",
			Type:        "dockerfile",
			Category:    "best-practice",
			Severity:    "medium",
			Title:       "Using :latest tag",
			Message:     "Avoid using :latest tag for reproducible builds",
			File:        filePath,
			Remediation: "Pin to specific version tag, e.g., FROM node:18.19.0-alpine",
		})
	}

	return findings
}

// checkKubernetes checks Kubernetes manifests for best practices
func (s *IaCBestPracticesScanner) checkKubernetes(filePath, content string) []IaCBestPracticeFinding {
	var findings []IaCBestPracticeFinding

	// Check for missing resource limits
	if strings.Contains(content, "containers:") && !strings.Contains(content, "resources:") {
		findings = append(findings, IaCBestPracticeFinding{
			RuleID:      "k8s-missing-resources",
			Type:        "kubernetes",
			Category:    "best-practice",
			Severity:    "medium",
			Title:       "Missing resource requests/limits",
			Message:     "Containers should define resource requests and limits",
			File:        filePath,
			Remediation: "Add resources: { requests: { cpu: '100m', memory: '128Mi' }, limits: { ... } }",
		})
	}

	// Check for missing liveness probe
	if strings.Contains(content, "containers:") && !strings.Contains(content, "livenessProbe:") {
		findings = append(findings, IaCBestPracticeFinding{
			RuleID:      "k8s-missing-liveness-probe",
			Type:        "kubernetes",
			Category:    "best-practice",
			Severity:    "low",
			Title:       "Missing liveness probe",
			Message:     "Containers should define liveness probes for health checking",
			File:        filePath,
			Remediation: "Add livenessProbe: { httpGet: { path: /health, port: 8080 } }",
		})
	}

	return findings
}

// checkTerraform checks Terraform files for best practices
func (s *IaCBestPracticesScanner) checkTerraform(filePath, content string) []IaCBestPracticeFinding {
	var findings []IaCBestPracticeFinding

	// Check for missing tags on AWS resources
	if strings.Contains(content, `resource "aws_`) && !strings.Contains(content, "tags") {
		findings = append(findings, IaCBestPracticeFinding{
			RuleID:      "terraform-missing-tags",
			Type:        "terraform",
			Category:    "best-practice",
			Severity:    "medium",
			Title:       "Missing resource tags",
			Message:     "AWS resources should have tags for organization and cost tracking",
			File:        filePath,
			Remediation: "Add tags = { Environment = var.environment, Owner = var.owner }",
		})
	}

	// Check for missing variable descriptions
	if strings.Contains(content, "variable ") && !strings.Contains(content, "description") {
		findings = append(findings, IaCBestPracticeFinding{
			RuleID:      "terraform-variable-no-description",
			Type:        "terraform",
			Category:    "best-practice",
			Severity:    "low",
			Title:       "Variable without description",
			Message:     "Variables should have descriptions for documentation",
			File:        filePath,
			Remediation: "Add description = \"Purpose of this variable\"",
		})
	}

	return findings
}

// determineIaCType determines the IaC type from file path/extension
func (s *IaCBestPracticesScanner) determineIaCType(filePath string) string {
	name := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))

	// Dockerfile
	if name == "Dockerfile" || strings.HasPrefix(name, "Dockerfile.") {
		return "dockerfile"
	}

	// Terraform
	if ext == ".tf" || ext == ".tfvars" {
		return "terraform"
	}

	// Helm
	if strings.Contains(filePath, "/charts/") ||
		strings.Contains(name, "values.yaml") ||
		strings.Contains(name, "values.yml") ||
		name == "Chart.yaml" {
		return "helm"
	}

	// Read file to determine type
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	content := string(data)

	if strings.Contains(content, "apiVersion:") && strings.Contains(content, "kind:") {
		return "kubernetes"
	}

	if strings.Contains(content, "AWSTemplateFormatVersion") ||
		strings.Contains(content, "Type: AWS::") {
		return "cloudformation"
	}

	return "unknown"
}
