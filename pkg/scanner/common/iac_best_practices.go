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

	// Find IaC files
	iacFiles := s.findIaCFiles(repoPath)
	if len(iacFiles) == 0 {
		return result
	}

	result.Summary.FilesScanned = len(iacFiles)
	s.onStatus("Scanning IaC files for best practices")

	// Generate rules if needed
	rulesPath := filepath.Join(s.cacheDir, "rules", "iac-best-practices.yaml")
	if err := s.ensureRulesGenerated(rulesPath); err != nil {
		result.Error = err
		return result
	}

	// Run Semgrep with generated rules
	result.Findings = s.scanWithSemgrep(ctx, iacFiles, repoPath, rulesPath)

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

	_ = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
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

	runner := NewSemgrepRunner(SemgrepConfig{
		RulePaths: []string{rulesPath},
		Timeout:   s.timeout,
		OnStatus:  s.onStatus,
	})

	result := runner.RunOnFiles(ctx, files, repoPath)
	if result.Error != nil {
		return findings
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
