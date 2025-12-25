// Package codecrypto provides a consolidated cryptographic security super scanner
// Features: ciphers, keys, random, tls, certificates
package codecrypto

import (
	"bufio"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/core/cyclonedx"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/scanner/common"
)

const (
	Name    = "code-crypto"
	Version = "3.0.0"
)

func init() {
	scanner.Register(&CryptoScanner{})
}

// CryptoScanner consolidates all cryptographic security analysis
type CryptoScanner struct{}

func (s *CryptoScanner) Name() string {
	return Name
}

func (s *CryptoScanner) Description() string {
	return "Consolidated cryptographic security scanner: weak ciphers, hardcoded keys, insecure random, TLS config, certificates"
}

func (s *CryptoScanner) Dependencies() []string {
	return nil
}

func (s *CryptoScanner) EstimateDuration(fileCount int) time.Duration {
	// Base estimate: 10 seconds + 1 second per 500 files
	est := 10 + fileCount/500
	return time.Duration(est) * time.Second
}

func (s *CryptoScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	// Get feature config
	cfg := getConfig(opts)

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Run all pattern-based features in parallel (ciphers, keys, random, tls)
	if cfg.Ciphers.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runCiphers(ctx, opts, cfg.Ciphers)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "ciphers")
			result.Summary.Ciphers = summary
			result.Findings.Ciphers = findings
			mu.Unlock()
		}()
	}

	if cfg.Keys.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runKeys(ctx, opts, cfg.Keys)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "keys")
			result.Summary.Keys = summary
			result.Findings.Keys = findings
			mu.Unlock()
		}()
	}

	if cfg.Random.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runRandom(ctx, opts, cfg.Random)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "random")
			result.Summary.Random = summary
			result.Findings.Random = findings
			mu.Unlock()
		}()
	}

	if cfg.TLS.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, findings := s.runTLS(ctx, opts, cfg.TLS)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "tls")
			result.Summary.TLS = summary
			result.Findings.TLS = findings
			mu.Unlock()
		}()
	}

	if cfg.Certificates.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, certResults := s.runCertificates(ctx, opts, cfg.Certificates)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "certificates")
			result.Summary.Certificates = summary
			result.Findings.Certificates = certResults
			mu.Unlock()
		}()
	}

	wg.Wait()

	scanResult := scanner.NewScanResult(Name, Version, start)
	scanResult.Repository = opts.RepoPath
	scanResult.SetSummary(result.Summary)
	scanResult.SetFindings(result.Findings)

	// Add features_run to metadata
	scanResult.SetMetadata(map[string]interface{}{
		"features_run": result.FeaturesRun,
	})

	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("creating output directory: %w", err)
		}
		resultFile := filepath.Join(opts.OutputDir, Name+".json")
		if err := scanResult.WriteJSON(resultFile); err != nil {
			return nil, fmt.Errorf("writing result: %w", err)
		}

		// Export CBOM (CycloneDX Cryptography Bill of Materials)
		if err := s.exportCBOM(opts.OutputDir, result); err != nil {
			// Log warning but don't fail the scan
			if opts.Verbose {
				fmt.Printf("[crypto] Warning: failed to export CBOM: %v\n", err)
			}
		}
	}

	return scanResult, nil
}

// exportCBOM exports findings as a CycloneDX CBOM
func (s *CryptoScanner) exportCBOM(outputDir string, result *Result) error {
	bom := cyclonedx.NewCBOM()

	// Add cipher findings as algorithm components
	for _, cipher := range result.Findings.Ciphers {
		c := cyclonedx.CipherFindingToComponent(
			cipher.Algorithm,
			extractCipherMode(cipher.Match),
			cipher.Severity,
			cipher.File,
			cipher.Line,
			cipher.Description,
		)
		bom.WithComponent(c)

		// Add vulnerability for weak/deprecated ciphers
		if cipher.Severity == "high" || cipher.Severity == "critical" {
			v := cyclonedx.Vulnerability{
				ID: fmt.Sprintf("CRYPTO-WEAK-%s-%s-%d", strings.ToUpper(cipher.Algorithm), filepath.Base(cipher.File), cipher.Line),
				Source: &cyclonedx.VulnSource{
					Name: "Zero Crypto Scanner",
				},
				Description:    cipher.Description,
				Recommendation: cipher.Suggestion,
				Ratings: []cyclonedx.VulnRating{
					{
						Severity: cyclonedx.SeverityToCycloneDX(cipher.Severity),
						Method:   "other",
					},
				},
				Affects: []cyclonedx.VulnAffect{
					{Ref: c.BOMRef},
				},
			}
			if cweInt := cyclonedx.CWEToInt(cipher.CWE); cweInt > 0 {
				v.CWEs = []int{cweInt}
			}
			bom.WithVulnerability(v)
		}
	}

	// Add key findings as vulnerabilities (hardcoded keys are vulnerabilities, not assets)
	for _, key := range result.Findings.Keys {
		v := cyclonedx.Vulnerability{
			ID: fmt.Sprintf("CRYPTO-HARDCODED-KEY-%s-%s-%d", strings.ToUpper(key.Type), filepath.Base(key.File), key.Line),
			Source: &cyclonedx.VulnSource{
				Name: "Zero Crypto Scanner",
			},
			Description:    key.Description,
			Recommendation: "Remove hardcoded keys and use secure key management",
			Detail:         fmt.Sprintf("Found in: %s:%d", key.File, key.Line),
			Ratings: []cyclonedx.VulnRating{
				{
					Severity: cyclonedx.SeverityToCycloneDX(key.Severity),
					Method:   "other",
				},
			},
		}
		if cweInt := cyclonedx.CWEToInt(key.CWE); cweInt > 0 {
			v.CWEs = []int{cweInt}
		}
		bom.WithVulnerability(v)
	}

	// Add TLS findings as protocol components
	for _, tls := range result.Findings.TLS {
		c := cyclonedx.TLSFindingToComponent(
			tls.Type,
			extractTLSVersion(tls.Match, tls.Description),
			tls.Severity,
			tls.File,
			tls.Line,
			tls.Description,
		)
		bom.WithComponent(c)

		// Add vulnerability for TLS issues
		v := cyclonedx.Vulnerability{
			ID: fmt.Sprintf("CRYPTO-TLS-%s-%s-%d", strings.ToUpper(tls.Type), filepath.Base(tls.File), tls.Line),
			Source: &cyclonedx.VulnSource{
				Name: "Zero Crypto Scanner",
			},
			Description:    tls.Description,
			Recommendation: tls.Suggestion,
			Detail:         fmt.Sprintf("Found in: %s:%d", tls.File, tls.Line),
			Ratings: []cyclonedx.VulnRating{
				{
					Severity: cyclonedx.SeverityToCycloneDX(tls.Severity),
					Method:   "other",
				},
			},
			Affects: []cyclonedx.VulnAffect{
				{Ref: c.BOMRef},
			},
		}
		if cweInt := cyclonedx.CWEToInt(tls.CWE); cweInt > 0 {
			v.CWEs = []int{cweInt}
		}
		bom.WithVulnerability(v)
	}

	// Add certificates as components
	if result.Findings.Certificates != nil {
		for _, cert := range result.Findings.Certificates.Certificates {
			c := cyclonedx.CertInfoToComponent(
				cert.Subject,
				cert.Issuer,
				cert.NotBefore.Format(time.RFC3339),
				cert.NotAfter.Format(time.RFC3339),
				cert.KeyType,
				cert.KeySize,
				cert.SignatureAlgo,
				cert.File,
				cert.IsSelfSigned,
			)
			bom.WithComponent(c)
		}

		// Add certificate findings as vulnerabilities
		for _, finding := range result.Findings.Certificates.Findings {
			v := cyclonedx.Vulnerability{
				ID: fmt.Sprintf("CRYPTO-CERT-%s-%s", strings.ToUpper(finding.Type), filepath.Base(finding.File)),
				Source: &cyclonedx.VulnSource{
					Name: "Zero Crypto Scanner",
				},
				Description:    finding.Description,
				Recommendation: finding.Suggestion,
				Ratings: []cyclonedx.VulnRating{
					{
						Severity: cyclonedx.SeverityToCycloneDX(finding.Severity),
						Method:   "other",
					},
				},
			}
			if finding.File != "" {
				v.Detail = fmt.Sprintf("Certificate file: %s", finding.File)
				v.Affects = []cyclonedx.VulnAffect{
					{Ref: fmt.Sprintf("crypto/certificate/%s", finding.File)},
				}
			}
			bom.WithVulnerability(v)
		}
	}

	// Write CBOM
	exporter := cyclonedx.NewExporter(outputDir)
	return exporter.WriteCBOM(bom, "cbom.cdx.json")
}

// extractCipherMode extracts cipher mode from match string
func extractCipherMode(match string) string {
	modes := []string{"gcm", "cbc", "ctr", "ecb", "cfb", "ofb", "ccm"}
	matchLower := strings.ToLower(match)
	for _, mode := range modes {
		if strings.Contains(matchLower, mode) {
			return mode
		}
	}
	return ""
}

// extractTLSVersion extracts TLS version from match/description
func extractTLSVersion(match, description string) string {
	text := strings.ToLower(match + " " + description)
	versions := []string{"1.3", "1.2", "1.1", "1.0"}
	for _, v := range versions {
		if strings.Contains(text, "tls"+v) || strings.Contains(text, "tls "+v) ||
			strings.Contains(text, "tlsv"+v) {
			return v
		}
	}
	if strings.Contains(text, "sslv3") || strings.Contains(text, "ssl3") {
		return "SSLv3"
	}
	return "unknown"
}

func getConfig(opts *scanner.ScanOptions) FeatureConfig {
	if opts.FeatureConfig == nil {
		return DefaultConfig()
	}

	// Try to parse from FeatureConfig
	data, err := json.Marshal(opts.FeatureConfig)
	if err != nil {
		return DefaultConfig()
	}

	var cfg FeatureConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}

	return cfg
}

// ============================================================================
// CIPHERS FEATURE
// ============================================================================

func (s *CryptoScanner) runCiphers(ctx context.Context, opts *scanner.ScanOptions, cfg CiphersConfig) (*CiphersSummary, []CipherFinding) {
	var findings []CipherFinding
	usedSemgrep := false

	// Try Semgrep first for better AST-based detection
	if cfg.UseSemgrep && common.ToolExists("semgrep") {
		semgrepFindings := runSemgrepCryptoAnalysis(ctx, opts.RepoPath, opts.Timeout)
		findings = append(findings, semgrepFindings...)
		usedSemgrep = true
	}

	// Run pattern-based analysis
	if cfg.UsePatterns {
		patternFindings := scanForWeakCiphers(opts.RepoPath)
		findings = append(findings, patternFindings...)
	}

	// Deduplicate
	findings = deduplicateCipherFindings(findings)

	summary := &CiphersSummary{
		TotalFindings: len(findings),
		BySeverity:    make(map[string]int),
		ByAlgorithm:   make(map[string]int),
		UsedSemgrep:   usedSemgrep,
	}

	for _, f := range findings {
		summary.BySeverity[f.Severity]++
		summary.ByAlgorithm[f.Algorithm]++
	}

	return summary, findings
}

func runSemgrepCryptoAnalysis(ctx context.Context, repoPath string, timeout time.Duration) []CipherFinding {
	var findings []CipherFinding

	if timeout == 0 {
		timeout = 3 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := common.RunCommand(ctx, "semgrep",
		"scan",
		"--config", "p/security-audit",
		"--config", "p/secrets",
		"--json",
		"--quiet",
		"--include", "*.go",
		"--include", "*.py",
		"--include", "*.js",
		"--include", "*.ts",
		"--include", "*.java",
		"--include", "*.rb",
		"--include", "*.php",
		"--include", "*.rs",
		repoPath,
	)

	if err != nil || result == nil {
		return findings
	}

	var semgrepOutput struct {
		Results []struct {
			CheckID string `json:"check_id"`
			Path    string `json:"path"`
			Start   struct {
				Line int `json:"line"`
			} `json:"start"`
			Extra struct {
				Message  string `json:"message"`
				Severity string `json:"severity"`
				Lines    string `json:"lines"`
				Metadata struct {
					CWE      interface{} `json:"cwe"`
					Category string      `json:"category"`
				} `json:"metadata"`
			} `json:"extra"`
		} `json:"results"`
	}

	if err := json.Unmarshal(result.Stdout, &semgrepOutput); err != nil {
		return findings
	}

	cryptoKeywords := []string{
		"md5", "sha1", "des", "rc4", "ecb", "weak", "crypto", "cipher",
		"hash", "encrypt", "decrypt", "rsa", "aes", "blowfish",
	}

	for _, r := range semgrepOutput.Results {
		checkLower := strings.ToLower(r.CheckID)
		msgLower := strings.ToLower(r.Extra.Message)

		isCrypto := false
		for _, kw := range cryptoKeywords {
			if strings.Contains(checkLower, kw) || strings.Contains(msgLower, kw) {
				isCrypto = true
				break
			}
		}

		if !isCrypto {
			continue
		}

		severity := strings.ToLower(r.Extra.Severity)
		switch severity {
		case "warning":
			severity = "medium"
		case "error":
			severity = "high"
		case "info":
			severity = "low"
		}

		algorithm := extractAlgorithm(r.CheckID, r.Extra.Message)
		cwe := extractCWE(r.Extra.Metadata.CWE)

		file := r.Path
		if strings.HasPrefix(file, repoPath) {
			file = strings.TrimPrefix(file, repoPath+"/")
		}

		findings = append(findings, CipherFinding{
			Algorithm:   algorithm,
			Severity:    severity,
			File:        file,
			Line:        r.Start.Line,
			Description: r.Extra.Message,
			Match:       strings.TrimSpace(r.Extra.Lines),
			Suggestion:  getCipherSuggestion(algorithm),
			CWE:         cwe,
			Source:      "semgrep",
		})
	}

	return findings
}

func extractAlgorithm(checkID, message string) string {
	combined := strings.ToLower(checkID + " " + message)

	if strings.Contains(combined, "md5") {
		return "MD5"
	}
	if strings.Contains(combined, "sha1") || strings.Contains(combined, "sha-1") {
		return "SHA-1"
	}
	if strings.Contains(combined, "des") {
		return "DES/3DES"
	}
	if strings.Contains(combined, "rc4") || strings.Contains(combined, "rc2") {
		return "RC4/RC2"
	}
	if strings.Contains(combined, "ecb") {
		return "ECB Mode"
	}
	if strings.Contains(combined, "blowfish") || strings.Contains(combined, "cast") {
		return "Legacy Cipher"
	}
	if strings.Contains(combined, "rsa") && strings.Contains(combined, "1024") {
		return "Weak RSA"
	}
	if strings.Contains(combined, "padding") {
		return "Weak Padding"
	}

	return "Weak Crypto"
}

func extractCWE(cwe interface{}) string {
	switch v := cwe.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				return s
			}
		}
	}
	return "CWE-327"
}

func getCipherSuggestion(algorithm string) string {
	suggestions := map[string]string{
		"MD5":          "Use SHA-256 or SHA-3 for hashing, bcrypt/argon2 for passwords",
		"SHA-1":        "Use SHA-256 or SHA-3",
		"DES/3DES":     "Use AES-256-GCM instead",
		"RC4/RC2":      "Use AES-256-GCM instead",
		"ECB Mode":     "Use GCM or CBC with HMAC",
		"Legacy Cipher": "Use AES-256-GCM",
		"Weak RSA":     "Use at least 2048-bit RSA, prefer 4096-bit or ECDSA",
		"Weak Padding": "Use OAEP padding for RSA, PKCS7 for block ciphers",
	}
	if s, ok := suggestions[algorithm]; ok {
		return s
	}
	return "Use modern cryptographic algorithms"
}

func deduplicateCipherFindings(findings []CipherFinding) []CipherFinding {
	seen := make(map[string]bool)
	var result []CipherFinding

	for _, f := range findings {
		key := fmt.Sprintf("%s:%d:%s", f.File, f.Line, f.Algorithm)
		if !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}

	return result
}

var weakCipherPatterns = []struct {
	pattern     *regexp.Regexp
	algorithm   string
	description string
	severity    string
	suggestion  string
	cwe         string
}{
	{
		regexp.MustCompile(`(?i)\b(3DES|DES3|TripleDES|TDES|DES-CBC|DES-ECB|DES_EDE|DES_KEY|DESX|desede)\b|['"]DES['"]|crypto\.DES|Cipher\.DES|DES_MODE`),
		"DES/3DES",
		"DES/3DES is deprecated and should not be used",
		"high",
		"Use AES-256-GCM instead",
		"CWE-327",
	},
	{
		regexp.MustCompile(`(?i)\b(RC4|ARCFOUR|RC2)\b`),
		"RC4/RC2",
		"RC4/RC2 is broken and must not be used",
		"critical",
		"Use AES-256-GCM instead",
		"CWE-327",
	},
	{
		regexp.MustCompile(`(?i)\bMD5\s*\(|\bcrypto\.MD5\b|hashlib\.md5|\.md5\(|createHash\(['"]md5['"]\)|MD5\.Create|new\s+MD5|DigestUtils\.md5`),
		"MD5",
		"MD5 is cryptographically broken for security purposes",
		"high",
		"Use SHA-256 or SHA-3 for hashing, bcrypt/argon2 for passwords",
		"CWE-328",
	},
	{
		regexp.MustCompile(`(?i)\bSHA1\s*\(|\bcrypto\.SHA1\b|hashlib\.sha1|\.sha1\(|createHash\(['"]sha1['"]\)|SHA1\.Create|new\s+SHA1|DigestUtils\.sha1`),
		"SHA-1",
		"SHA-1 is deprecated and vulnerable to collision attacks",
		"medium",
		"Use SHA-256 or SHA-3",
		"CWE-328",
	},
	{
		regexp.MustCompile(`(?i)\b(ECB\s*mode|AES.*ECB|ECB_MODE|MODE_ECB|ECB_PKCS5PADDING|CipherMode\.ECB|AES/ECB)\b`),
		"ECB Mode",
		"ECB mode does not provide semantic security",
		"high",
		"Use GCM or CBC with HMAC",
		"CWE-327",
	},
	{
		regexp.MustCompile(`(?i)\b(Blowfish|CAST5|IDEA)\b`),
		"Legacy Cipher",
		"Legacy cipher algorithm detected",
		"medium",
		"Use AES-256-GCM",
		"CWE-327",
	},
	{
		regexp.MustCompile(`(?i)\bRSA.{0,30}1024\b|1024.{0,30}\bRSA\b|RSAKeyGenParameterSpec\s*\(\s*1024`),
		"Weak RSA",
		"1024-bit RSA is too weak",
		"high",
		"Use at least 2048-bit RSA, prefer 4096-bit or ECDSA",
		"CWE-326",
	},
	{
		regexp.MustCompile(`(?i)padding\s*=\s*['"]?(none|zero|PKCS1v15)['"]?|NoPadding|PKCS1Padding`),
		"Weak Padding",
		"Weak or no padding detected",
		"medium",
		"Use OAEP padding for RSA, PKCS7 for block ciphers",
		"CWE-327",
	},
}

var codeExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".java": true,
	".rb": true, ".php": true, ".cs": true, ".cpp": true, ".c": true,
	".h": true, ".hpp": true, ".rs": true, ".swift": true, ".kt": true,
}

var configExtensions = map[string]bool{
	".yaml": true, ".yml": true, ".json": true, ".xml": true,
	".conf": true, ".config": true, ".ini": true, ".properties": true,
}

func scanForWeakCiphers(repoPath string) []CipherFinding {
	var findings []CipherFinding

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !codeExtensions[ext] {
			return nil
		}

		fileFindings := scanFileForCiphers(path, repoPath)
		findings = append(findings, fileFindings...)
		return nil
	})

	return findings
}

func scanFileForCiphers(filePath, repoPath string) []CipherFinding {
	var findings []CipherFinding

	file, err := os.Open(filePath)
	if err != nil {
		return findings
	}
	defer file.Close()

	relPath := filePath
	if strings.HasPrefix(filePath, repoPath) {
		relPath = strings.TrimPrefix(filePath, repoPath+"/")
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		for _, pat := range weakCipherPatterns {
			if pat.pattern.MatchString(line) {
				findings = append(findings, CipherFinding{
					Algorithm:   pat.algorithm,
					Severity:    pat.severity,
					File:        relPath,
					Line:        lineNum,
					Description: pat.description,
					Match:       truncateMatch(pat.pattern.FindString(line)),
					Suggestion:  pat.suggestion,
					CWE:         pat.cwe,
					Source:      "pattern",
				})
			}
		}
	}

	return findings
}

// ============================================================================
// KEYS FEATURE
// ============================================================================

func (s *CryptoScanner) runKeys(ctx context.Context, opts *scanner.ScanOptions, cfg KeysConfig) (*KeysSummary, []KeyFinding) {
	findings := scanForHardcodedKeys(opts.RepoPath, cfg)

	summary := &KeysSummary{
		TotalFindings: len(findings),
		BySeverity:    make(map[string]int),
		ByType:        make(map[string]int),
	}

	for _, f := range findings {
		summary.BySeverity[f.Severity]++
		summary.ByType[f.Type]++
	}

	return summary, findings
}

var hardcodedKeyPatterns = []struct {
	pattern     *regexp.Regexp
	keyType     string
	description string
	severity    string
	cwe         string
	category    string // api, private, aws, signing
}{
	{
		regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["'][a-zA-Z0-9]{20,}["']`),
		"api-key",
		"Hardcoded API key detected",
		"critical",
		"CWE-798",
		"api",
	},
	{
		regexp.MustCompile(`(?i)(secret[_-]?key|secretkey)\s*[:=]\s*["'][a-zA-Z0-9+/=]{16,}["']`),
		"secret-key",
		"Hardcoded secret key detected",
		"critical",
		"CWE-798",
		"api",
	},
	{
		regexp.MustCompile(`(?i)(encryption[_-]?key|crypto[_-]?key|aes[_-]?key)\s*[:=]\s*["'][a-zA-Z0-9+/=]{16,}["']`),
		"encryption-key",
		"Hardcoded encryption key detected",
		"critical",
		"CWE-798",
		"api",
	},
	{
		regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`),
		"private-key",
		"Private key embedded in code",
		"critical",
		"CWE-798",
		"private",
	},
	{
		regexp.MustCompile(`-----BEGIN\s+EC\s+PRIVATE\s+KEY-----`),
		"ec-private-key",
		"EC private key embedded in code",
		"critical",
		"CWE-798",
		"private",
	},
	{
		regexp.MustCompile(`-----BEGIN\s+ENCRYPTED\s+PRIVATE\s+KEY-----`),
		"encrypted-private-key",
		"Encrypted private key in code (password may be nearby)",
		"high",
		"CWE-798",
		"private",
	},
	{
		regexp.MustCompile(`(?i)(iv|nonce)\s*[:=]\s*["'][a-zA-Z0-9+/=]{16,}["']`),
		"static-iv",
		"Hardcoded IV/nonce detected (may be static)",
		"high",
		"CWE-329",
		"api",
	},
	{
		regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),
		"aws-access-key",
		"AWS Access Key ID detected",
		"critical",
		"CWE-798",
		"aws",
	},
	{
		regexp.MustCompile(`(?i)(signing[_-]?key|hmac[_-]?key)\s*[:=]\s*["'][a-zA-Z0-9+/=]{16,}["']`),
		"signing-key",
		"Hardcoded signing/HMAC key detected",
		"critical",
		"CWE-798",
		"signing",
	},
	{
		regexp.MustCompile(`(?i)(master[_-]?key|root[_-]?key)\s*[:=]\s*["'][a-zA-Z0-9+/=]{16,}["']`),
		"master-key",
		"Hardcoded master/root key detected",
		"critical",
		"CWE-798",
		"signing",
	},
}

func scanForHardcodedKeys(repoPath string, cfg KeysConfig) []KeyFinding {
	var findings []KeyFinding

	// Build list of enabled patterns
	var enabledPatterns []struct {
		pattern     *regexp.Regexp
		keyType     string
		description string
		severity    string
		cwe         string
	}

	for _, pat := range hardcodedKeyPatterns {
		include := false
		switch pat.category {
		case "api":
			include = cfg.CheckAPIKeys
		case "private":
			include = cfg.CheckPrivate
		case "aws":
			include = cfg.CheckAWS
		case "signing":
			include = cfg.CheckSigning
		default:
			include = true
		}
		if include {
			enabledPatterns = append(enabledPatterns, struct {
				pattern     *regexp.Regexp
				keyType     string
				description string
				severity    string
				cwe         string
			}{pat.pattern, pat.keyType, pat.description, pat.severity, pat.cwe})
		}
	}

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !codeExtensions[ext] && !configExtensions[ext] {
			return nil
		}

		fileFindings := scanFileForKeys(path, repoPath, enabledPatterns, cfg.RedactMatches)
		findings = append(findings, fileFindings...)
		return nil
	})

	return findings
}

func scanFileForKeys(filePath, repoPath string, patterns []struct {
	pattern     *regexp.Regexp
	keyType     string
	description string
	severity    string
	cwe         string
}, redact bool) []KeyFinding {
	var findings []KeyFinding

	file, err := os.Open(filePath)
	if err != nil {
		return findings
	}
	defer file.Close()

	relPath := filePath
	if strings.HasPrefix(filePath, repoPath) {
		relPath = strings.TrimPrefix(filePath, repoPath+"/")
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		for _, pat := range patterns {
			if pat.pattern.MatchString(line) {
				match := pat.pattern.FindString(line)
				if redact {
					match = redactSensitive(match)
				}
				findings = append(findings, KeyFinding{
					Type:        pat.keyType,
					Severity:    pat.severity,
					File:        relPath,
					Line:        lineNum,
					Description: pat.description,
					Match:       match,
					CWE:         pat.cwe,
				})
			}
		}
	}

	return findings
}

func redactSensitive(match string) string {
	if len(match) > 30 {
		return match[:10] + "[REDACTED]"
	}
	return "[REDACTED]"
}

// ============================================================================
// RANDOM FEATURE
// ============================================================================

func (s *CryptoScanner) runRandom(ctx context.Context, opts *scanner.ScanOptions, cfg RandomConfig) (*RandomSummary, []RandomFinding) {
	findings := scanForWeakRandom(opts.RepoPath)

	summary := &RandomSummary{
		TotalFindings: len(findings),
		BySeverity:    make(map[string]int),
		ByType:        make(map[string]int),
	}

	for _, f := range findings {
		summary.BySeverity[f.Severity]++
		summary.ByType[f.Type]++
	}

	return summary, findings
}

var weakRandomPatterns = []struct {
	pattern     *regexp.Regexp
	randType    string
	description string
	severity    string
	suggestion  string
	cwe         string
}{
	{
		regexp.MustCompile(`Math\.random\s*\(\)`),
		"js-math-random",
		"Math.random() is not cryptographically secure",
		"high",
		"Use crypto.getRandomValues() or crypto.randomBytes()",
		"CWE-338",
	},
	{
		regexp.MustCompile(`(?i)random\.random\s*\(\)|random\.randint`),
		"python-random",
		"Python random module is not cryptographically secure",
		"high",
		"Use secrets module or os.urandom()",
		"CWE-338",
	},
	{
		regexp.MustCompile(`\brand\s*\(\)|srand\s*\(`),
		"c-rand",
		"C rand() is not cryptographically secure",
		"high",
		"Use arc4random(), getrandom(), or platform CSPRNG",
		"CWE-338",
	},
	{
		regexp.MustCompile(`java\.util\.Random`),
		"java-random",
		"java.util.Random is not cryptographically secure",
		"high",
		"Use java.security.SecureRandom",
		"CWE-338",
	},
	{
		regexp.MustCompile(`new\s+Random\s*\(\s*\)`),
		"random-no-seed",
		"Using predictable Random without seed",
		"medium",
		"Use SecureRandom or crypto-safe alternative",
		"CWE-338",
	},
	{
		regexp.MustCompile(`(?i)uuid\.uuid1|uuid1\s*\(`),
		"uuid1",
		"UUID1 is based on MAC address and time, not random",
		"medium",
		"Use uuid.uuid4() for random UUIDs",
		"CWE-338",
	},
	{
		regexp.MustCompile(`rand\.Seed\s*\(\s*time\.Now`),
		"go-time-seed",
		"Using time as random seed is predictable",
		"high",
		"Use crypto/rand instead of math/rand for security",
		"CWE-338",
	},
	{
		regexp.MustCompile(`math/rand`),
		"go-math-rand",
		"math/rand is not cryptographically secure",
		"medium",
		"Use crypto/rand for security-sensitive operations",
		"CWE-338",
	},
	{
		regexp.MustCompile(`Random\(\s*System\.currentTimeMillis`),
		"java-time-seed",
		"Using system time as random seed is predictable",
		"high",
		"Use SecureRandom instead",
		"CWE-338",
	},
}

func scanForWeakRandom(repoPath string) []RandomFinding {
	var findings []RandomFinding

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !codeExtensions[ext] {
			return nil
		}

		fileFindings := scanFileForRandom(path, repoPath)
		findings = append(findings, fileFindings...)
		return nil
	})

	return findings
}

func scanFileForRandom(filePath, repoPath string) []RandomFinding {
	var findings []RandomFinding

	file, err := os.Open(filePath)
	if err != nil {
		return findings
	}
	defer file.Close()

	relPath := filePath
	if strings.HasPrefix(filePath, repoPath) {
		relPath = strings.TrimPrefix(filePath, repoPath+"/")
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		for _, pat := range weakRandomPatterns {
			if pat.pattern.MatchString(line) {
				findings = append(findings, RandomFinding{
					Type:        pat.randType,
					Severity:    pat.severity,
					File:        relPath,
					Line:        lineNum,
					Description: pat.description,
					Match:       truncateMatch(pat.pattern.FindString(line)),
					Suggestion:  pat.suggestion,
					CWE:         pat.cwe,
				})
			}
		}
	}

	return findings
}

// ============================================================================
// TLS FEATURE
// ============================================================================

func (s *CryptoScanner) runTLS(ctx context.Context, opts *scanner.ScanOptions, cfg TLSConfig) (*TLSSummary, []TLSFinding) {
	findings := scanForTLSIssues(opts.RepoPath, cfg)

	summary := &TLSSummary{
		TotalFindings: len(findings),
		BySeverity:    make(map[string]int),
		ByType:        make(map[string]int),
	}

	for _, f := range findings {
		summary.BySeverity[f.Severity]++
		summary.ByType[f.Type]++
	}

	return summary, findings
}

var tlsPatterns = []struct {
	pattern     *regexp.Regexp
	tlsType     string
	description string
	severity    string
	suggestion  string
	cwe         string
	category    string // protocols, verification, ciphers, urls
}{
	{
		regexp.MustCompile(`(?i)(SSLv2|SSLv3|TLSv1\.0|TLSv1_0|TLS1_0)`),
		"deprecated-protocol",
		"Deprecated TLS/SSL version (SSLv2, SSLv3, TLS 1.0)",
		"critical",
		"Use TLS 1.2 or TLS 1.3 only",
		"CWE-327",
		"protocols",
	},
	{
		regexp.MustCompile(`(?i)TLSv1\.1|TLSv1_1|TLS1_1`),
		"deprecated-protocol",
		"TLS 1.1 is deprecated",
		"high",
		"Use TLS 1.2 or TLS 1.3",
		"CWE-327",
		"protocols",
	},
	{
		regexp.MustCompile(`(?i)InsecureSkipVerify\s*:\s*true`),
		"disabled-verification",
		"TLS certificate verification disabled (Go)",
		"critical",
		"Enable certificate verification",
		"CWE-295",
		"verification",
	},
	{
		regexp.MustCompile(`(?i)verify\s*[:=]\s*false`),
		"disabled-verification",
		"TLS certificate verification disabled",
		"critical",
		"Enable certificate verification",
		"CWE-295",
		"verification",
	},
	{
		regexp.MustCompile(`(?i)CERT_NONE|ssl\.CERT_NONE`),
		"disabled-verification",
		"Python SSL certificate verification disabled",
		"critical",
		"Use ssl.CERT_REQUIRED",
		"CWE-295",
		"verification",
	},
	{
		regexp.MustCompile(`(?i)ssl[._]verify\s*[:=]\s*false`),
		"disabled-verification",
		"SSL verification explicitly disabled",
		"critical",
		"Enable SSL verification",
		"CWE-295",
		"verification",
	},
	{
		regexp.MustCompile(`(?i)rejectUnauthorized\s*:\s*false`),
		"disabled-verification",
		"Node.js TLS verification disabled",
		"critical",
		"Set rejectUnauthorized: true",
		"CWE-295",
		"verification",
	},
	{
		regexp.MustCompile(`(?i)(disable[_-]?ssl|no[_-]?ssl|skip[_-]?ssl)[_-]?verify`),
		"disabled-verification",
		"SSL verification explicitly disabled",
		"critical",
		"Enable SSL verification",
		"CWE-295",
		"verification",
	},
	{
		regexp.MustCompile(`(?i)(MinVersion|min_version)\s*[:=]\s*(tls\.)?VersionTLS10`),
		"weak-min-version",
		"Minimum TLS version set too low",
		"high",
		"Set minimum TLS version to 1.2",
		"CWE-327",
		"protocols",
	},
	{
		regexp.MustCompile(`(?i)http://[^"'\s]+\.(com|org|net|io|dev|app|gov|edu)`),
		"insecure-url",
		"HTTP (non-HTTPS) URL detected",
		"medium",
		"Use HTTPS for all external communications",
		"CWE-319",
		"urls",
	},
	{
		regexp.MustCompile(`(?i)cipher.*NULL|NULL.*cipher`),
		"null-cipher",
		"NULL cipher suite detected",
		"critical",
		"Remove NULL cipher from allowed suites",
		"CWE-327",
		"ciphers",
	},
	{
		regexp.MustCompile(`(?i)(cipher|ssl|tls|openssl).{0,50}(EXPORT|ANON|aNULL|eNULL)\b|(EXPORT|ANON|aNULL|eNULL).{0,50}(cipher|ssl|tls)`),
		"weak-cipher-suite",
		"Weak cipher suite detected in TLS config",
		"high",
		"Use strong cipher suites only",
		"CWE-327",
		"ciphers",
	},
}

func scanForTLSIssues(repoPath string, cfg TLSConfig) []TLSFinding {
	var findings []TLSFinding

	// Build list of enabled patterns
	var enabledPatterns []struct {
		pattern     *regexp.Regexp
		tlsType     string
		description string
		severity    string
		suggestion  string
		cwe         string
	}

	for _, pat := range tlsPatterns {
		include := false
		switch pat.category {
		case "protocols":
			include = cfg.CheckProtocols
		case "verification":
			include = cfg.CheckVerification
		case "ciphers":
			include = cfg.CheckCipherSuites
		case "urls":
			include = cfg.CheckInsecureURLs
		default:
			include = true
		}
		if include {
			enabledPatterns = append(enabledPatterns, struct {
				pattern     *regexp.Regexp
				tlsType     string
				description string
				severity    string
				suggestion  string
				cwe         string
			}{pat.pattern, pat.tlsType, pat.description, pat.severity, pat.suggestion, pat.cwe})
		}
	}

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !codeExtensions[ext] && !configExtensions[ext] {
			return nil
		}

		fileFindings := scanFileForTLS(path, repoPath, enabledPatterns)
		findings = append(findings, fileFindings...)
		return nil
	})

	return findings
}

func scanFileForTLS(filePath, repoPath string, patterns []struct {
	pattern     *regexp.Regexp
	tlsType     string
	description string
	severity    string
	suggestion  string
	cwe         string
}) []TLSFinding {
	var findings []TLSFinding

	file, err := os.Open(filePath)
	if err != nil {
		return findings
	}
	defer file.Close()

	relPath := filePath
	if strings.HasPrefix(filePath, repoPath) {
		relPath = strings.TrimPrefix(filePath, repoPath+"/")
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		for _, pat := range patterns {
			if pat.pattern.MatchString(line) {
				findings = append(findings, TLSFinding{
					Type:        pat.tlsType,
					Severity:    pat.severity,
					File:        relPath,
					Line:        lineNum,
					Description: pat.description,
					Match:       truncateMatch(pat.pattern.FindString(line)),
					Suggestion:  pat.suggestion,
					CWE:         pat.cwe,
				})
			}
		}
	}

	return findings
}

// ============================================================================
// CERTIFICATES FEATURE
// ============================================================================

func (s *CryptoScanner) runCertificates(ctx context.Context, opts *scanner.ScanOptions, cfg CertificatesConfig) (*CertificatesSummary, *CertificatesResult) {
	certs := findCertificates(opts.RepoPath)

	var allCertInfos []CertInfo
	var allFindings []CertFinding

	for _, certPath := range certs {
		certInfos, certFindings := analyzeCertificate(certPath, opts.RepoPath, cfg)
		allCertInfos = append(allCertInfos, certInfos...)
		allFindings = append(allFindings, certFindings...)
	}

	summary := &CertificatesSummary{
		TotalCertificates: len(allCertInfos),
		TotalFindings:     len(allFindings),
		BySeverity:        make(map[string]int),
	}

	for _, f := range allFindings {
		summary.BySeverity[f.Severity]++
		switch f.Type {
		case "expiring-soon":
			summary.ExpiringSoon++
		case "expired":
			summary.Expired++
		case "weak-key":
			summary.WeakKey++
		}
	}

	return summary, &CertificatesResult{
		Certificates: allCertInfos,
		Findings:     allFindings,
	}
}

var certExtensions = map[string]bool{
	".pem": true, ".crt": true, ".cer": true, ".cert": true,
	".der": true, ".p7b": true, ".p7c": true,
}

func findCertificates(repoPath string) []string {
	var certs []string

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if certExtensions[ext] {
			certs = append(certs, path)
			return nil
		}

		base := strings.ToLower(info.Name())
		if strings.Contains(base, "cert") || strings.Contains(base, "ssl") ||
			strings.Contains(base, "tls") || base == "ca-bundle" {
			certs = append(certs, path)
		}

		return nil
	})

	return certs
}

func analyzeCertificate(certPath, repoPath string, cfg CertificatesConfig) ([]CertInfo, []CertFinding) {
	var infos []CertInfo
	var findings []CertFinding

	relPath := certPath
	if strings.HasPrefix(certPath, repoPath) {
		relPath = strings.TrimPrefix(certPath, repoPath+"/")
	}

	data, err := os.ReadFile(certPath)
	if err != nil {
		return infos, findings
	}

	// Try to parse PEM blocks
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			if strings.Contains(block.Type, "PRIVATE KEY") {
				findings = append(findings, CertFinding{
					Type:        "private-key-in-cert-file",
					Severity:    "high",
					File:        relPath,
					Description: "Private key found in certificate file",
					Suggestion:  "Store private keys separately with restricted permissions",
				})
			}
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		info, certFindings := analyzeParsedCert(cert, relPath, cfg)
		infos = append(infos, info)
		findings = append(findings, certFindings...)
	}

	// Try DER format if no PEM blocks found
	if len(infos) == 0 {
		cert, err := x509.ParseCertificate(data)
		if err == nil {
			info, certFindings := analyzeParsedCert(cert, relPath, cfg)
			infos = append(infos, info)
			findings = append(findings, certFindings...)
		}
	}

	return infos, findings
}

func analyzeParsedCert(cert *x509.Certificate, file string, cfg CertificatesConfig) (CertInfo, []CertFinding) {
	var findings []CertFinding
	now := time.Now()

	daysUntilExp := int(cert.NotAfter.Sub(now).Hours() / 24)

	keyType, keySize := getKeyInfo(cert)

	info := CertInfo{
		File:          file,
		Subject:       cert.Subject.String(),
		Issuer:        cert.Issuer.String(),
		NotBefore:     cert.NotBefore,
		NotAfter:      cert.NotAfter,
		DaysUntilExp:  daysUntilExp,
		KeyType:       keyType,
		KeySize:       keySize,
		SignatureAlgo: cert.SignatureAlgorithm.String(),
		IsSelfSigned:  cert.Subject.String() == cert.Issuer.String(),
		IsCA:          cert.IsCA,
		DNSNames:      cert.DNSNames,
		Serial:        cert.SerialNumber.String(),
	}

	// Check expiration
	if daysUntilExp < 0 {
		findings = append(findings, CertFinding{
			Type:        "expired",
			Severity:    "critical",
			File:        file,
			Description: fmt.Sprintf("Certificate expired %d days ago", -daysUntilExp),
			Suggestion:  "Replace with a valid certificate immediately",
		})
	} else if daysUntilExp < 30 {
		findings = append(findings, CertFinding{
			Type:        "expiring-soon",
			Severity:    "high",
			File:        file,
			Description: fmt.Sprintf("Certificate expires in %d days", daysUntilExp),
			Suggestion:  "Renew certificate before expiration",
		})
	} else if daysUntilExp < cfg.ExpiryWarningDays {
		findings = append(findings, CertFinding{
			Type:        "expiring-soon",
			Severity:    "medium",
			File:        file,
			Description: fmt.Sprintf("Certificate expires in %d days", daysUntilExp),
			Suggestion:  "Plan for certificate renewal",
		})
	}

	// Check key strength
	if cfg.CheckKeyStrength {
		switch keyType {
		case "RSA":
			if keySize < 2048 {
				findings = append(findings, CertFinding{
					Type:        "weak-key",
					Severity:    "critical",
					File:        file,
					Description: fmt.Sprintf("RSA key too small: %d bits", keySize),
					Suggestion:  "Use at least 2048-bit RSA, prefer 4096-bit",
				})
			} else if keySize < 4096 {
				findings = append(findings, CertFinding{
					Type:        "weak-key",
					Severity:    "low",
					File:        file,
					Description: fmt.Sprintf("RSA key size: %d bits (recommended: 4096)", keySize),
					Suggestion:  "Consider using 4096-bit RSA or ECDSA P-384",
				})
			}
		case "ECDSA":
			if keySize < 256 {
				findings = append(findings, CertFinding{
					Type:        "weak-key",
					Severity:    "high",
					File:        file,
					Description: fmt.Sprintf("ECDSA key too small: %d bits", keySize),
					Suggestion:  "Use at least P-256, prefer P-384",
				})
			}
		}
	}

	// Check signature algorithm
	if cfg.CheckSignatureAlgo {
		weakSigAlgos := map[string]bool{
			"MD2-RSA":  true,
			"MD5-RSA":  true,
			"SHA1-RSA": true,
		}
		if weakSigAlgos[cert.SignatureAlgorithm.String()] {
			findings = append(findings, CertFinding{
				Type:        "weak-signature",
				Severity:    "high",
				File:        file,
				Description: fmt.Sprintf("Weak signature algorithm: %s", cert.SignatureAlgorithm),
				Suggestion:  "Use SHA-256 or SHA-384 based signatures",
			})
		}
	}

	// Check self-signed
	if cfg.CheckSelfSigned && info.IsSelfSigned && !info.IsCA {
		findings = append(findings, CertFinding{
			Type:        "self-signed",
			Severity:    "medium",
			File:        file,
			Description: "Self-signed certificate detected",
			Suggestion:  "Use certificates from a trusted CA for production",
		})
	}

	// Check for wildcard certificates
	for _, name := range cert.DNSNames {
		if strings.HasPrefix(name, "*") {
			findings = append(findings, CertFinding{
				Type:        "wildcard-cert",
				Severity:    "low",
				File:        file,
				Description: fmt.Sprintf("Wildcard certificate: %s", name),
				Suggestion:  "Consider using specific certificates for better security isolation",
			})
			break
		}
	}

	// Check validity period
	if cfg.CheckValidityPeriod {
		validityDays := int(cert.NotAfter.Sub(cert.NotBefore).Hours() / 24)
		if validityDays > 825 {
			findings = append(findings, CertFinding{
				Type:        "long-validity",
				Severity:    "low",
				File:        file,
				Description: fmt.Sprintf("Certificate validity period too long: %d days", validityDays),
				Suggestion:  "Use certificates with validity <= 398 days per CA/Browser Forum guidelines",
			})
		}
	}

	return info, findings
}

func getKeyInfo(cert *x509.Certificate) (string, int) {
	switch pub := cert.PublicKey.(type) {
	case interface{ Size() int }:
		return "RSA", pub.Size() * 8
	default:
		switch cert.PublicKeyAlgorithm {
		case x509.RSA:
			return "RSA", 2048
		case x509.ECDSA:
			return "ECDSA", 256
		case x509.Ed25519:
			return "Ed25519", 256
		default:
			return "Unknown", 0
		}
	}
}

// ============================================================================
// UTILITIES
// ============================================================================

func truncateMatch(match string) string {
	if len(match) > 50 {
		return match[:50] + "..."
	}
	return match
}
