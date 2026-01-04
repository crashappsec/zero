// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package diff

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// FingerprintGenerator generates fingerprints for findings
type FingerprintGenerator struct{}

// NewFingerprintGenerator creates a new fingerprint generator
func NewFingerprintGenerator() *FingerprintGenerator {
	return &FingerprintGenerator{}
}

// FingerprintFindings generates fingerprints for all findings in a scanner result
func (g *FingerprintGenerator) FingerprintFindings(scanner string, data json.RawMessage) ([]FingerprintedFinding, error) {
	var result struct {
		Findings json.RawMessage `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	switch scanner {
	case "code-security":
		return g.fingerprintCodeSecurity(result.Findings)
	case "package-analysis":
		return g.fingerprintPackageAnalysis(result.Findings)
	case "crypto":
		return g.fingerprintCrypto(result.Findings)
	case "devops":
		return g.fingerprintDevops(result.Findings)
	case "technology-identification":
		return g.fingerprintTechID(result.Findings)
	default:
		// Unknown scanner - return empty
		return nil, nil
	}
}

// FingerprintedFinding combines a fingerprint with the original finding data
type FingerprintedFinding struct {
	Fingerprint FindingFingerprint `json:"fingerprint"`
	Finding     json.RawMessage    `json:"finding"`
	Severity    string             `json:"severity"`
	Scanner     string             `json:"scanner"`
	Feature     string             `json:"feature,omitempty"`
	File        string             `json:"file,omitempty"`
	Line        int                `json:"line,omitempty"`
	Message     string             `json:"message,omitempty"`
}

// fingerprintCodeSecurity handles code-security scanner findings
func (g *FingerprintGenerator) fingerprintCodeSecurity(data json.RawMessage) ([]FingerprintedFinding, error) {
	var findings struct {
		Vulns   []json.RawMessage `json:"vulns"`
		Secrets []json.RawMessage `json:"secrets"`
		API     []json.RawMessage `json:"api"`
	}
	if err := json.Unmarshal(data, &findings); err != nil {
		return nil, err
	}

	var result []FingerprintedFinding

	// Process vulns
	for _, raw := range findings.Vulns {
		var f struct {
			RuleID   string `json:"rule_id"`
			Title    string `json:"title"`
			Severity string `json:"severity"`
			File     string `json:"file"`
			Line     int    `json:"line"`
			Column   int    `json:"column"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "code-security/vulns",
			PrimaryKey:  fmt.Sprintf("%s:%s", f.RuleID, normalizePath(f.File)),
			LocationKey: fmt.Sprintf("%s:%d:%d", f.File, f.Line, f.Column),
			ContentHash: hashContent(f.RuleID, f.File, f.Title),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "code-security",
			Feature:     "vulns",
			File:        f.File,
			Line:        f.Line,
			Message:     f.Title,
		})
	}

	// Process secrets
	for _, raw := range findings.Secrets {
		var f struct {
			RuleID          string `json:"rule_id"`
			Type            string `json:"type"`
			Severity        string `json:"severity"`
			Message         string `json:"message"`
			File            string `json:"file"`
			Line            int    `json:"line"`
			Column          int    `json:"column"`
			DetectionSource string `json:"detection_source"`
			Snippet         string `json:"snippet"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		// Use detection source if available, otherwise default to semgrep
		source := f.DetectionSource
		if source == "" {
			source = "semgrep"
		}

		fp := FindingFingerprint{
			Scanner:     "code-security/secrets",
			PrimaryKey:  fmt.Sprintf("%s:%s:%s", f.Type, normalizePath(f.File), source),
			LocationKey: fmt.Sprintf("%s:%d:%d", f.File, f.Line, f.Column),
			ContentHash: hashContent(f.Type, f.File, maskSecret(f.Snippet)),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "code-security",
			Feature:     "secrets",
			File:        f.File,
			Line:        f.Line,
			Message:     fmt.Sprintf("%s: %s", f.Type, f.Message),
		})
	}

	// Process API findings
	for _, raw := range findings.API {
		var f struct {
			RuleID   string `json:"rule_id"`
			Title    string `json:"title"`
			Severity string `json:"severity"`
			File     string `json:"file"`
			Line     int    `json:"line"`
			Category string `json:"category"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "code-security/api",
			PrimaryKey:  fmt.Sprintf("%s:%s", f.RuleID, normalizePath(f.File)),
			LocationKey: fmt.Sprintf("%s:%d", f.File, f.Line),
			ContentHash: hashContent(f.RuleID, f.File, f.Category),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "code-security",
			Feature:     "api",
			File:        f.File,
			Line:        f.Line,
			Message:     f.Title,
		})
	}

	return result, nil
}

// fingerprintPackageAnalysis handles package-analysis scanner findings
func (g *FingerprintGenerator) fingerprintPackageAnalysis(data json.RawMessage) ([]FingerprintedFinding, error) {
	var findings struct {
		Vulns      json.RawMessage `json:"vulns"`
		Malcontent json.RawMessage `json:"malcontent"`
		Confusion  json.RawMessage `json:"confusion"`
		Typosquats json.RawMessage `json:"typosquats"`
	}
	if err := json.Unmarshal(data, &findings); err != nil {
		return nil, err
	}

	var result []FingerprintedFinding

	// Process vulns
	if len(findings.Vulns) > 0 {
		var vulns []json.RawMessage
		if err := json.Unmarshal(findings.Vulns, &vulns); err == nil {
			for _, raw := range vulns {
				var f struct {
					ID        string `json:"id"`
					Package   string `json:"package"`
					Version   string `json:"version"`
					Ecosystem string `json:"ecosystem"`
					Severity  string `json:"severity"`
					Title     string `json:"title"`
				}
				if err := json.Unmarshal(raw, &f); err != nil {
					continue
				}

				fp := FindingFingerprint{
					Scanner:     "package-analysis/vulns",
					PrimaryKey:  fmt.Sprintf("%s:%s:%s", f.ID, f.Package, f.Ecosystem),
					LocationKey: "", // Package findings don't have file locations
					ContentHash: hashContent(f.ID, f.Package, f.Version),
				}

				result = append(result, FingerprintedFinding{
					Fingerprint: fp,
					Finding:     raw,
					Severity:    f.Severity,
					Scanner:     "package-analysis",
					Feature:     "vulns",
					Message:     fmt.Sprintf("%s: %s@%s - %s", f.ID, f.Package, f.Version, f.Title),
				})
			}
		}
	}

	// Process malcontent
	if len(findings.Malcontent) > 0 {
		var malcontent []json.RawMessage
		if err := json.Unmarshal(findings.Malcontent, &malcontent); err == nil {
			for _, raw := range malcontent {
				var f struct {
					File      string   `json:"file"`
					Risk      string   `json:"risk"`
					RiskScore int      `json:"risk_score"`
					Behaviors []string `json:"behaviors"`
				}
				if err := json.Unmarshal(raw, &f); err != nil {
					continue
				}

				fp := FindingFingerprint{
					Scanner:     "package-analysis/malcontent",
					PrimaryKey:  fmt.Sprintf("%s:%s", normalizePath(f.File), f.Risk),
					LocationKey: f.File,
					ContentHash: hashContent(f.File, f.Risk, strings.Join(f.Behaviors, ",")),
				}

				result = append(result, FingerprintedFinding{
					Fingerprint: fp,
					Finding:     raw,
					Severity:    mapRiskToSeverity(f.Risk),
					Scanner:     "package-analysis",
					Feature:     "malcontent",
					File:        f.File,
					Message:     fmt.Sprintf("Malicious behavior: %s", f.Risk),
				})
			}
		}
	}

	// Process confusion findings
	if len(findings.Confusion) > 0 {
		var confusion []json.RawMessage
		if err := json.Unmarshal(findings.Confusion, &confusion); err == nil {
			for _, raw := range confusion {
				var f struct {
					Package   string `json:"package"`
					Ecosystem string `json:"ecosystem"`
					RiskLevel string `json:"risk_level"`
					RiskType  string `json:"risk_type"`
					File      string `json:"file"`
				}
				if err := json.Unmarshal(raw, &f); err != nil {
					continue
				}

				fp := FindingFingerprint{
					Scanner:     "package-analysis/confusion",
					PrimaryKey:  fmt.Sprintf("%s:%s:%s", f.Package, f.Ecosystem, f.RiskType),
					LocationKey: f.File,
					ContentHash: hashContent(f.Package, f.Ecosystem, f.RiskType),
				}

				result = append(result, FingerprintedFinding{
					Fingerprint: fp,
					Finding:     raw,
					Severity:    f.RiskLevel,
					Scanner:     "package-analysis",
					Feature:     "confusion",
					File:        f.File,
					Message:     fmt.Sprintf("Dependency confusion: %s (%s)", f.Package, f.RiskType),
				})
			}
		}
	}

	// Process typosquats
	if len(findings.Typosquats) > 0 {
		var typosquats []json.RawMessage
		if err := json.Unmarshal(findings.Typosquats, &typosquats); err == nil {
			for _, raw := range typosquats {
				var f struct {
					Package   string `json:"package"`
					Ecosystem string `json:"ecosystem"`
					SimilarTo string `json:"similar_to"`
					RiskLevel string `json:"risk_level"`
					Reason    string `json:"reason"`
				}
				if err := json.Unmarshal(raw, &f); err != nil {
					continue
				}

				fp := FindingFingerprint{
					Scanner:     "package-analysis/typosquats",
					PrimaryKey:  fmt.Sprintf("%s:%s", f.Package, f.Ecosystem),
					LocationKey: "",
					ContentHash: hashContent(f.Package, f.Ecosystem, f.SimilarTo),
				}

				result = append(result, FingerprintedFinding{
					Fingerprint: fp,
					Finding:     raw,
					Severity:    f.RiskLevel,
					Scanner:     "package-analysis",
					Feature:     "typosquats",
					Message:     fmt.Sprintf("Typosquat: %s (similar to %s)", f.Package, f.SimilarTo),
				})
			}
		}
	}

	return result, nil
}

// fingerprintCrypto handles crypto scanner findings
func (g *FingerprintGenerator) fingerprintCrypto(data json.RawMessage) ([]FingerprintedFinding, error) {
	var findings struct {
		Ciphers []json.RawMessage `json:"ciphers"`
		Keys    []json.RawMessage `json:"keys"`
		Random  []json.RawMessage `json:"random"`
		TLS     []json.RawMessage `json:"tls"`
	}
	if err := json.Unmarshal(data, &findings); err != nil {
		return nil, err
	}

	var result []FingerprintedFinding

	// Process ciphers
	for _, raw := range findings.Ciphers {
		var f struct {
			Algorithm string `json:"algorithm"`
			Severity  string `json:"severity"`
			File      string `json:"file"`
			Line      int    `json:"line"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "crypto/ciphers",
			PrimaryKey:  fmt.Sprintf("%s:%s", f.Algorithm, normalizePath(f.File)),
			LocationKey: fmt.Sprintf("%s:%d", f.File, f.Line),
			ContentHash: hashContent(f.Algorithm, f.File),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "crypto",
			Feature:     "ciphers",
			File:        f.File,
			Line:        f.Line,
			Message:     fmt.Sprintf("Weak cipher: %s", f.Algorithm),
		})
	}

	// Process keys
	for _, raw := range findings.Keys {
		var f struct {
			Type     string `json:"type"`
			Severity string `json:"severity"`
			File     string `json:"file"`
			Line     int    `json:"line"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "crypto/keys",
			PrimaryKey:  fmt.Sprintf("%s:%s", f.Type, normalizePath(f.File)),
			LocationKey: fmt.Sprintf("%s:%d", f.File, f.Line),
			ContentHash: hashContent(f.Type, f.File),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "crypto",
			Feature:     "keys",
			File:        f.File,
			Line:        f.Line,
			Message:     fmt.Sprintf("Hardcoded key: %s", f.Type),
		})
	}

	// Process random
	for _, raw := range findings.Random {
		var f struct {
			Type     string `json:"type"`
			Severity string `json:"severity"`
			File     string `json:"file"`
			Line     int    `json:"line"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "crypto/random",
			PrimaryKey:  fmt.Sprintf("%s:%s", f.Type, normalizePath(f.File)),
			LocationKey: fmt.Sprintf("%s:%d", f.File, f.Line),
			ContentHash: hashContent(f.Type, f.File),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "crypto",
			Feature:     "random",
			File:        f.File,
			Line:        f.Line,
			Message:     fmt.Sprintf("Insecure random: %s", f.Type),
		})
	}

	// Process TLS
	for _, raw := range findings.TLS {
		var f struct {
			Type     string `json:"type"`
			Severity string `json:"severity"`
			File     string `json:"file"`
			Line     int    `json:"line"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "crypto/tls",
			PrimaryKey:  fmt.Sprintf("%s:%s", f.Type, normalizePath(f.File)),
			LocationKey: fmt.Sprintf("%s:%d", f.File, f.Line),
			ContentHash: hashContent(f.Type, f.File),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "crypto",
			Feature:     "tls",
			File:        f.File,
			Line:        f.Line,
			Message:     fmt.Sprintf("TLS issue: %s", f.Type),
		})
	}

	return result, nil
}

// fingerprintDevops handles devops scanner findings
func (g *FingerprintGenerator) fingerprintDevops(data json.RawMessage) ([]FingerprintedFinding, error) {
	var findings struct {
		IaC           []json.RawMessage `json:"iac"`
		Containers    []json.RawMessage `json:"containers"`
		GitHubActions []json.RawMessage `json:"github_actions"`
	}
	if err := json.Unmarshal(data, &findings); err != nil {
		return nil, err
	}

	var result []FingerprintedFinding

	// Process IaC
	for _, raw := range findings.IaC {
		var f struct {
			RuleID   string `json:"rule_id"`
			Title    string `json:"title"`
			Severity string `json:"severity"`
			File     string `json:"file"`
			Line     int    `json:"line"`
			Resource string `json:"resource"`
			Type     string `json:"type"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "devops/iac",
			PrimaryKey:  fmt.Sprintf("%s:%s:%s", f.RuleID, f.Resource, normalizePath(f.File)),
			LocationKey: fmt.Sprintf("%s:%d", f.File, f.Line),
			ContentHash: hashContent(f.RuleID, f.Resource, f.Type),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "devops",
			Feature:     "iac",
			File:        f.File,
			Line:        f.Line,
			Message:     f.Title,
		})
	}

	// Process Containers
	for _, raw := range findings.Containers {
		var f struct {
			VulnID     string `json:"vuln_id"`
			Title      string `json:"title"`
			Severity   string `json:"severity"`
			Image      string `json:"image"`
			Dockerfile string `json:"dockerfile"`
			Package    string `json:"package"`
			Version    string `json:"version"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "devops/containers",
			PrimaryKey:  fmt.Sprintf("%s:%s:%s", f.VulnID, f.Image, f.Package),
			LocationKey: f.Dockerfile,
			ContentHash: hashContent(f.VulnID, f.Package, f.Version),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "devops",
			Feature:     "containers",
			File:        f.Dockerfile,
			Message:     fmt.Sprintf("%s: %s in %s", f.VulnID, f.Package, f.Image),
		})
	}

	// Process GitHub Actions
	for _, raw := range findings.GitHubActions {
		var f struct {
			RuleID   string `json:"rule_id"`
			Title    string `json:"title"`
			Severity string `json:"severity"`
			File     string `json:"file"`
			Line     int    `json:"line"`
			Category string `json:"category"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "devops/github_actions",
			PrimaryKey:  fmt.Sprintf("%s:%s", f.RuleID, normalizePath(f.File)),
			LocationKey: fmt.Sprintf("%s:%d", f.File, f.Line),
			ContentHash: hashContent(f.RuleID, f.File, f.Category),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "devops",
			Feature:     "github_actions",
			File:        f.File,
			Line:        f.Line,
			Message:     f.Title,
		})
	}

	return result, nil
}

// fingerprintTechID handles technology-identification scanner findings (security and governance)
func (g *FingerprintGenerator) fingerprintTechID(data json.RawMessage) ([]FingerprintedFinding, error) {
	var findings struct {
		Security   []json.RawMessage `json:"security"`
		Governance []json.RawMessage `json:"governance"`
	}
	if err := json.Unmarshal(data, &findings); err != nil {
		return nil, err
	}

	var result []FingerprintedFinding

	// Process security findings
	for _, raw := range findings.Security {
		var f struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Severity  string `json:"severity"`
			Category  string `json:"category"`
			File      string `json:"file"`
			Line      int    `json:"line"`
			ModelName string `json:"model_name"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "technology-identification/security",
			PrimaryKey:  fmt.Sprintf("%s:%s", f.ID, f.Category),
			LocationKey: fmt.Sprintf("%s:%d", f.File, f.Line),
			ContentHash: hashContent(f.ID, f.Category, f.ModelName),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "technology-identification",
			Feature:     "security",
			File:        f.File,
			Line:        f.Line,
			Message:     f.Title,
		})
	}

	// Process governance findings
	for _, raw := range findings.Governance {
		var f struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Severity  string `json:"severity"`
			Category  string `json:"category"`
			ModelName string `json:"model_name"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}

		fp := FindingFingerprint{
			Scanner:     "technology-identification/governance",
			PrimaryKey:  fmt.Sprintf("%s:%s:%s", f.ID, f.Category, f.ModelName),
			LocationKey: "",
			ContentHash: hashContent(f.ID, f.Category, f.ModelName),
		}

		result = append(result, FingerprintedFinding{
			Fingerprint: fp,
			Finding:     raw,
			Severity:    f.Severity,
			Scanner:     "technology-identification",
			Feature:     "governance",
			Message:     f.Title,
		})
	}

	return result, nil
}

// Helper functions

// normalizePath normalizes a file path for consistent comparison
func normalizePath(path string) string {
	// Convert to forward slashes
	path = filepath.ToSlash(path)
	// Remove leading ./
	path = strings.TrimPrefix(path, "./")
	// Remove leading /
	path = strings.TrimPrefix(path, "/")
	return path
}

// hashContent creates a SHA256 hash of multiple strings
func hashContent(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0}) // Separator
	}
	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars
}

// maskSecret masks the middle portion of a secret for hashing
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// mapRiskToSeverity maps risk levels to standard severities
func mapRiskToSeverity(risk string) string {
	switch strings.ToLower(risk) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium", "moderate":
		return "medium"
	case "low":
		return "low"
	default:
		return "info"
	}
}
