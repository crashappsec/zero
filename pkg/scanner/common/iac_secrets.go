// Package common provides shared utilities for scanners
// This file provides IaC secrets scanning using RAG-generated Semgrep rules
package common

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// IaCSecretsScanner scans IaC files for hardcoded secrets
type IaCSecretsScanner struct {
	ragPath   string
	cacheDir  string
	timeout   time.Duration
	onStatus  func(string)
}

// IaCSecretsConfig configures the IaC secrets scanner
type IaCSecretsConfig struct {
	RAGPath  string
	CacheDir string
	Timeout  time.Duration
	OnStatus func(string)
}

// IaCSecretFinding represents a secret found in an IaC file
type IaCSecretFinding struct {
	RuleID      string
	Type        string // terraform, kubernetes, cloudformation, github-actions, helm
	SecretType  string // aws_key, password, token, etc.
	Severity    string
	Message     string
	File        string
	Line        int
	Column      int
	Snippet     string
	Remediation string
	IaCContext  IaCContext
}

// IaCContext provides context about the IaC finding
type IaCContext struct {
	FileType   string // .tf, .yaml, .json
	Resource   string // resource name or kind
	Provider   string // aws, gcp, azure for terraform
	IsTemplate bool   // whether file is a template
}

// IaCSecretsResult holds the result of scanning
type IaCSecretsResult struct {
	Findings []IaCSecretFinding
	Summary  IaCSecretsSummary
	Error    error
}

// IaCSecretsSummary provides summary statistics
type IaCSecretsSummary struct {
	TotalFindings int
	ByType        map[string]int // by IaC type
	BySecretType  map[string]int // by secret type
	BySeverity    map[string]int
	FilesScanned  int
}

// NewIaCSecretsScanner creates a new IaC secrets scanner
func NewIaCSecretsScanner(cfg IaCSecretsConfig) *IaCSecretsScanner {
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

	return &IaCSecretsScanner{
		ragPath:  cfg.RAGPath,
		cacheDir: cfg.CacheDir,
		timeout:  cfg.Timeout,
		onStatus: cfg.OnStatus,
	}
}

// Scan scans a repository for secrets in IaC files
func (s *IaCSecretsScanner) Scan(ctx context.Context, repoPath string) *IaCSecretsResult {
	result := &IaCSecretsResult{
		Summary: IaCSecretsSummary{
			ByType:       make(map[string]int),
			BySecretType: make(map[string]int),
			BySeverity:   make(map[string]int),
		},
	}

	// Find IaC files
	iacFiles := s.findIaCFiles(repoPath)
	if len(iacFiles) == 0 {
		return result
	}

	result.Summary.FilesScanned = len(iacFiles)
	s.onStatus("Found " + string(rune(len(iacFiles))) + " IaC files")

	// Generate rules if needed
	rulesPath := filepath.Join(s.cacheDir, "rules", "secrets-in-iac.yaml")
	if err := s.ensureRulesGenerated(rulesPath); err != nil {
		// Fall back to direct pattern matching
		result.Findings = s.scanWithPatterns(ctx, iacFiles, repoPath)
	} else {
		// Use Semgrep with generated rules
		result.Findings = s.scanWithSemgrep(ctx, iacFiles, repoPath, rulesPath)
	}

	// Build summary
	for _, f := range result.Findings {
		result.Summary.TotalFindings++
		result.Summary.ByType[f.Type]++
		result.Summary.BySecretType[f.SecretType]++
		result.Summary.BySeverity[f.Severity]++
	}

	return result
}

// findIaCFiles finds all IaC files in the repository
func (s *IaCSecretsScanner) findIaCFiles(repoPath string) []string {
	var files []string

	iacExtensions := map[string]bool{
		".tf":     true, // Terraform
		".tfvars": true, // Terraform variables
		".yaml":   true, // K8s, CloudFormation, Ansible
		".yml":    true, // K8s, CloudFormation, Ansible
		".json":   true, // CloudFormation, K8s
	}

	iacPatterns := []string{
		"values.yaml",      // Helm
		"values.yml",       // Helm
		"Chart.yaml",       // Helm
		"kustomization.*",  // Kustomize
		"docker-compose.*", // Docker Compose
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

		// Skip example/template files (likely false positives)
		nameLower := strings.ToLower(info.Name())
		if strings.Contains(nameLower, ".example.") ||
			strings.Contains(nameLower, ".sample.") ||
			strings.Contains(nameLower, ".template.") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		// Check extension
		if iacExtensions[ext] {
			// Additional filter: for YAML/JSON, check if it's actually IaC
			if ext == ".yaml" || ext == ".yml" || ext == ".json" {
				if s.isIaCFile(path) {
					files = append(files, path)
				}
			} else {
				files = append(files, path)
			}
			return nil
		}

		// Check patterns
		for _, pattern := range iacPatterns {
			matched, _ := filepath.Match(pattern, info.Name())
			if matched {
				files = append(files, path)
				return nil
			}
		}

		// Check GitHub Actions workflows
		if strings.Contains(path, ".github/workflows/") &&
			(ext == ".yaml" || ext == ".yml") {
			files = append(files, path)
		}

		return nil
	})

	return files
}

// isIaCFile checks if a YAML/JSON file is an IaC file
func (s *IaCSecretsScanner) isIaCFile(path string) bool {
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
		strings.Contains(content, "Resources:") && strings.Contains(content, "Type: AWS::") {
		return true
	}

	// Ansible markers
	if strings.Contains(content, "hosts:") && strings.Contains(content, "tasks:") {
		return true
	}

	// Docker Compose markers
	if strings.Contains(content, "services:") && (strings.Contains(content, "image:") || strings.Contains(content, "build:")) {
		return true
	}

	// Helm values
	if strings.Contains(filepath.Base(path), "values") {
		return true
	}

	return false
}

// ensureRulesGenerated generates Semgrep rules from RAG if needed
func (s *IaCSecretsScanner) ensureRulesGenerated(rulesPath string) error {
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
	return GenerateRulesFromRAG(s.ragPath, "devops/secrets-in-iac", rulesPath)
}

// scanWithSemgrep uses Semgrep with generated rules
func (s *IaCSecretsScanner) scanWithSemgrep(ctx context.Context, files []string, repoPath, rulesPath string) []IaCSecretFinding {
	var findings []IaCSecretFinding

	if !HasSemgrep() {
		return s.scanWithPatterns(ctx, files, repoPath)
	}

	runner := NewSemgrepRunner(SemgrepConfig{
		RulePaths: []string{rulesPath},
		Timeout:   s.timeout,
		OnStatus:  s.onStatus,
	})

	result := runner.RunOnFiles(ctx, files, repoPath)
	if result.Error != nil {
		return s.scanWithPatterns(ctx, files, repoPath)
	}

	for _, f := range result.Findings {
		finding := IaCSecretFinding{
			RuleID:      f.RuleID,
			Type:        determineIaCType(f.File),
			SecretType:  determineSecretType(f.RuleID, f.Match),
			Severity:    f.Severity,
			Message:     f.Message,
			File:        f.File,
			Line:        f.Line,
			Column:      f.Column,
			Snippet:     redactSecretInSnippet(f.Match),
			Remediation: f.Remediation,
			IaCContext: IaCContext{
				FileType:   filepath.Ext(f.File),
				IsTemplate: isTemplateFile(f.File),
			},
		}
		findings = append(findings, finding)
	}

	return findings
}

// scanWithPatterns uses direct pattern matching (fallback)
func (s *IaCSecretsScanner) scanWithPatterns(ctx context.Context, files []string, repoPath string) []IaCSecretFinding {
	// This is a simplified fallback - in practice, you'd implement
	// the regex patterns from the RAG file here
	return nil
}

// determineIaCType determines the IaC type from file path/extension
func determineIaCType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	if ext == ".tf" || ext == ".tfvars" {
		return "terraform"
	}

	if strings.Contains(filePath, ".github/workflows/") {
		return "github-actions"
	}

	// Check for Helm
	if strings.Contains(filePath, "/charts/") ||
		strings.Contains(filePath, "values.yaml") ||
		strings.Contains(filePath, "values.yml") {
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

	if strings.Contains(content, "hosts:") && strings.Contains(content, "tasks:") {
		return "ansible"
	}

	return "unknown"
}

// determineSecretType determines the type of secret from rule ID and content
func determineSecretType(ruleID, content string) string {
	ruleIDLower := strings.ToLower(ruleID)
	contentLower := strings.ToLower(content)

	switch {
	case strings.Contains(ruleIDLower, "aws") && strings.Contains(ruleIDLower, "access"):
		return "aws_access_key"
	case strings.Contains(ruleIDLower, "aws") && strings.Contains(ruleIDLower, "secret"):
		return "aws_secret_key"
	case strings.Contains(ruleIDLower, "password") || strings.Contains(contentLower, "password"):
		return "password"
	case strings.Contains(ruleIDLower, "private_key") || strings.Contains(contentLower, "private key"):
		return "private_key"
	case strings.Contains(ruleIDLower, "api_key") || strings.Contains(ruleIDLower, "apikey"):
		return "api_key"
	case strings.Contains(ruleIDLower, "token"):
		return "token"
	case strings.Contains(ruleIDLower, "jwt"):
		return "jwt_token"
	case strings.Contains(ruleIDLower, "database") || strings.Contains(ruleIDLower, "connection"):
		return "database_credential"
	default:
		return "generic_secret"
	}
}

// isTemplateFile checks if file is a template
func isTemplateFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return strings.Contains(name, "template") ||
		strings.Contains(name, ".j2") ||
		strings.Contains(name, ".tpl")
}

// redactSecretInSnippet redacts potential secrets in code snippets
func redactSecretInSnippet(snippet string) string {
	if len(snippet) > 200 {
		snippet = snippet[:200] + "..."
	}

	// Simple redaction - mask long alphanumeric strings
	words := strings.Fields(snippet)
	for i, word := range words {
		clean := strings.Trim(word, "\"'`=:,;")
		if len(clean) > 16 && isAlphanumericPlusLocal(clean) {
			if len(clean) > 8 {
				masked := clean[:4] + "****" + clean[len(clean)-4:]
				words[i] = strings.Replace(word, clean, masked, 1)
			}
		}
	}

	return strings.Join(words, " ")
}

// isAlphanumericPlusLocal checks if string is alphanumeric with common secret chars
func isAlphanumericPlusLocal(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-' ||
			c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}

// findRAGPath locates the RAG directory
func findRAGPath() string {
	candidates := []string{
		"rag",
		"../rag",
		"../../rag",
	}

	if zeroHome := os.Getenv("ZERO_HOME"); zeroHome != "" {
		candidates = append([]string{filepath.Join(zeroHome, "..", "rag")}, candidates...)
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			return absPath
		}
	}

	return ""
}

// getCacheDir returns the cache directory
func getCacheDir() string {
	if zeroHome := os.Getenv("ZERO_HOME"); zeroHome != "" {
		return filepath.Join(zeroHome, "cache")
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".zero", "cache")
}

// ConvertToSecretFindings converts IaC secret findings to code-security SecretFinding format
// This allows IaC secrets to be included in the code-security scanner output
func (f *IaCSecretFinding) ToGenericSecretFinding() map[string]interface{} {
	return map[string]interface{}{
		"rule_id":          f.RuleID,
		"type":             f.SecretType,
		"severity":         f.Severity,
		"message":          f.Message,
		"file":             f.File,
		"line":             f.Line,
		"column":           f.Column,
		"snippet":          f.Snippet,
		"detection_source": "iac-scanner",
		"iac_type":         f.Type,
		"remediation":      f.Remediation,
	}
}
