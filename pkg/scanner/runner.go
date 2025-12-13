package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// NativeRunner executes Go-native scanners
type NativeRunner struct {
	ZeroHome    string
	Timeout     time.Duration
	Parallel    int
	OnProgress  func(scanner string, status Status, summary string)
}

// NewNativeRunner creates a new native scanner runner
func NewNativeRunner(zeroHome string) *NativeRunner {
	return &NativeRunner{
		ZeroHome: zeroHome,
		Timeout:  5 * time.Minute,
		Parallel: 4,
	}
}

// RunOptions configures a scanner run
type RunOptions struct {
	RepoPath     string
	OutputDir    string
	Scanners     []Scanner
	SkipScanners []string
	Timeout      time.Duration
	Parallel     int
}

// RunScanners executes all configured scanners for a repository
func (r *NativeRunner) RunScanners(ctx context.Context, opts RunOptions) (*RunResult, error) {
	start := time.Now()

	// Build skip set
	skipSet := make(map[string]bool)
	for _, s := range opts.SkipScanners {
		skipSet[s] = true
	}

	// Filter scanners
	var scanners []Scanner
	for _, s := range opts.Scanners {
		if !skipSet[s.Name()] {
			scanners = append(scanners, s)
		}
	}

	if len(scanners) == 0 {
		return &RunResult{
			Success:  true,
			Duration: time.Since(start),
			Results:  make(map[string]*Result),
		}, nil
	}

	// Sort scanners by dependencies
	sorted, err := TopologicalSort(scanners)
	if err != nil {
		return nil, fmt.Errorf("sorting scanners: %w", err)
	}

	// Group by dependency levels for parallel execution
	levels, err := GroupByDependencies(sorted)
	if err != nil {
		return nil, fmt.Errorf("grouping scanners: %w", err)
	}

	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join(r.ZeroHome, "repos", opts.RepoPath, "analysis")
	}

	// Track results
	results := make(map[string]*Result)
	var resultsMu sync.Mutex

	// Track SBOM path for dependent scanners
	var sbomPath string
	var sbomMu sync.RWMutex

	// Determine timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = r.Timeout
	}

	// Process each level (levels can run in parallel, but must complete before next level)
	for _, level := range levels {
		// Check for cancellation before starting new level
		select {
		case <-ctx.Done():
			return &RunResult{
				Success:  false,
				Results:  results,
				Duration: time.Since(start),
			}, ctx.Err()
		default:
		}

		// Create a wait group for this level
		var wg sync.WaitGroup
		sem := make(chan struct{}, r.Parallel)

		for _, s := range level {
			scanner := s // capture for goroutine

			wg.Add(1)
			go func() {
				defer wg.Done()

				// Check for cancellation before acquiring semaphore
				select {
				case <-ctx.Done():
					return
				case sem <- struct{}{}:
				}
				defer func() { <-sem }()

				// Notify progress
				if r.OnProgress != nil {
					r.OnProgress(scanner.Name(), StatusRunning, "")
				}

				// Build scan options
				sbomMu.RLock()
				scanOpts := &ScanOptions{
					RepoPath:  opts.RepoPath,
					OutputDir: outputDir,
					SBOMPath:  sbomPath,
					Timeout:   timeout,
				}
				sbomMu.RUnlock()

				// Create context with per-scanner timeout
				scanCtx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				// Run scanner
				scanStart := time.Now()
				scanResult, err := scanner.Run(scanCtx, scanOpts)
				duration := time.Since(scanStart)

				// Store result
				result := &Result{
					Scanner:  scanner.Name(),
					Duration: duration,
				}

				if err != nil {
					result.Status = StatusFailed
					result.Error = err
					if r.OnProgress != nil {
						r.OnProgress(scanner.Name(), StatusFailed, err.Error())
					}
				} else if scanResult.Error != "" {
					if scanResult.Error == "timeout" {
						result.Status = StatusTimeout
					} else {
						result.Status = StatusFailed
						result.Error = fmt.Errorf("%s", scanResult.Error)
					}
					if r.OnProgress != nil {
						r.OnProgress(scanner.Name(), result.Status, scanResult.Error)
					}
				} else {
					result.Status = StatusComplete
					result.Summary = extractSummaryString(scanner.Name(), scanResult)
					if r.OnProgress != nil {
						r.OnProgress(scanner.Name(), StatusComplete, result.Summary)
					}
				}

				resultsMu.Lock()
				results[scanner.Name()] = result
				resultsMu.Unlock()

				// If this was SBOM scanner, save the path
				if scanner.Name() == "package-sbom" && result.Status == StatusComplete {
					sbomMu.Lock()
					sbomPath = filepath.Join(outputDir, "sbom.cdx.json")
					sbomMu.Unlock()
				}
			}()
		}

		// Wait for all scanners in this level to complete
		wg.Wait()
	}

	return &RunResult{
		Success:  true,
		Results:  results,
		Duration: time.Since(start),
	}, nil
}

// extractSummaryString creates a human-readable summary from scan result
func extractSummaryString(name string, result *ScanResult) string {
	if result == nil {
		return "complete"
	}

	// Parse summary JSON
	var summary map[string]interface{}
	if err := parseJSON(result.Summary, &summary); err != nil {
		return "complete"
	}

	switch name {
	case "package-sbom":
		if total, ok := summary["total_packages"].(float64); ok {
			// Show ecosystem breakdown if available
			if ecosystems, ok := summary["by_ecosystem"].(map[string]interface{}); ok && len(ecosystems) > 0 {
				parts := make([]string, 0, len(ecosystems))
				for eco, count := range ecosystems {
					if c, ok := count.(float64); ok && c > 0 {
						parts = append(parts, fmt.Sprintf("%s: %.0f", eco, c))
					}
				}
				if len(parts) > 0 {
					return fmt.Sprintf("%.0f packages (%s)", total, strings.Join(parts, ", "))
				}
			}
			return fmt.Sprintf("%.0f packages", total)
		}
	case "package-vulns":
		c := getIntFromMap(summary, "critical")
		h := getIntFromMap(summary, "high")
		m := getIntFromMap(summary, "medium")
		l := getIntFromMap(summary, "low")
		if c+h+m+l == 0 {
			return "no findings"
		}
		return fmt.Sprintf("%d critical, %d high, %d medium, %d low", c, h, m, l)
	case "licenses":
		total := getIntFromMap(summary, "total_packages")
		unique := getIntFromMap(summary, "unique_licenses")
		// Show top license counts
		if counts, ok := summary["license_counts"].(map[string]interface{}); ok && len(counts) > 0 {
			parts := make([]string, 0, len(counts))
			for lic, count := range counts {
				if c, ok := count.(float64); ok && c > 0 {
					parts = append(parts, fmt.Sprintf("%s: %.0f", lic, c))
				}
			}
			if len(parts) > 0 {
				// Limit to top 5 licenses for display
				if len(parts) > 5 {
					parts = parts[:5]
					return fmt.Sprintf("%d packages, %d license types (%s, ...)", total, unique, strings.Join(parts, ", "))
				}
				return fmt.Sprintf("%d packages, %d license types (%s)", total, unique, strings.Join(parts, ", "))
			}
		}
		return fmt.Sprintf("%d packages, %d license types", total, unique)
	case "package-health":
		critical := getIntFromMap(summary, "critical_count")
		warning := getIntFromMap(summary, "warning_count")
		if critical > 0 {
			return fmt.Sprintf("%d critical, %d warnings", critical, warning)
		}
		if warning > 0 {
			return fmt.Sprintf("%d warnings", warning)
		}
		return "healthy"
	case "package-malcontent":
		c := getIntFromMap(summary, "critical")
		h := getIntFromMap(summary, "high")
		if c+h == 0 {
			return "no suspicious behavior"
		}
		return fmt.Sprintf("%d critical, %d high", c, h)
	case "code-secrets":
		c := getIntFromMap(summary, "critical")
		h := getIntFromMap(summary, "high")
		m := getIntFromMap(summary, "medium")
		total := getIntFromMap(summary, "total_findings")
		if total == 0 {
			return "no secrets found"
		}
		return fmt.Sprintf("%d findings (%d critical, %d high, %d medium)", total, c, h, m)
	case "tech-discovery":
		total := getIntFromMap(summary, "total_technologies")
		if total == 0 {
			return "no technologies detected"
		}
		// Get primary languages if available
		if langs, ok := summary["primary_languages"].([]interface{}); ok && len(langs) > 0 {
			langNames := make([]string, 0, len(langs))
			for _, l := range langs {
				if name, ok := l.(string); ok {
					langNames = append(langNames, name)
				}
			}
			if len(langNames) > 0 {
				return fmt.Sprintf("%d technologies (%s)", total, strings.Join(langNames, ", "))
			}
		}
		return fmt.Sprintf("%d technologies detected", total)
	case "iac-security":
		c := getIntFromMap(summary, "critical")
		h := getIntFromMap(summary, "high")
		total := getIntFromMap(summary, "total_findings")
		if total == 0 {
			return "no misconfigurations found"
		}
		return fmt.Sprintf("%d findings (%d critical, %d high)", total, c, h)
	case "container-security":
		c := getIntFromMap(summary, "critical")
		h := getIntFromMap(summary, "high")
		total := getIntFromMap(summary, "total_findings")
		dockerfiles := getIntFromMap(summary, "dockerfiles_scanned")
		images := getIntFromMap(summary, "images_scanned")
		if total == 0 {
			if dockerfiles == 0 {
				return "no Dockerfiles found"
			}
			if images == 0 {
				return fmt.Sprintf("%d Dockerfiles, no base images to scan", dockerfiles)
			}
			return fmt.Sprintf("%d images scanned, no CVEs", images)
		}
		return fmt.Sprintf("%d CVEs (%d critical, %d high) in %d images", total, c, h, images)
	case "crypto-ciphers":
		total := getIntFromMap(summary, "total_findings")
		if total == 0 {
			return "no weak ciphers found"
		}
		// Show by algorithm if available
		if algos, ok := summary["by_algorithm"].(map[string]interface{}); ok && len(algos) > 0 {
			parts := make([]string, 0, len(algos))
			for algo, count := range algos {
				if c, ok := count.(float64); ok && c > 0 {
					parts = append(parts, fmt.Sprintf("%s: %.0f", algo, c))
				}
			}
			if len(parts) > 0 {
				return fmt.Sprintf("%d weak cipher findings (%s)", total, strings.Join(parts, ", "))
			}
		}
		return fmt.Sprintf("%d weak cipher findings", total)
	case "crypto-keys":
		total := getIntFromMap(summary, "total_findings")
		if total == 0 {
			return "no hardcoded keys found"
		}
		// Show by type if available
		if types, ok := summary["by_type"].(map[string]interface{}); ok && len(types) > 0 {
			parts := make([]string, 0, len(types))
			for t, count := range types {
				if c, ok := count.(float64); ok && c > 0 {
					parts = append(parts, fmt.Sprintf("%s: %.0f", t, c))
				}
			}
			if len(parts) > 0 {
				return fmt.Sprintf("%d hardcoded keys (%s)", total, strings.Join(parts, ", "))
			}
		}
		return fmt.Sprintf("%d hardcoded keys found", total)
	case "crypto-tls":
		total := getIntFromMap(summary, "total_findings")
		if total == 0 {
			return "no TLS issues found"
		}
		// Show by severity
		if sev, ok := summary["by_severity"].(map[string]interface{}); ok {
			c := 0
			h := 0
			if v, ok := sev["critical"].(float64); ok {
				c = int(v)
			}
			if v, ok := sev["high"].(float64); ok {
				h = int(v)
			}
			if c > 0 || h > 0 {
				return fmt.Sprintf("%d TLS issues (%d critical, %d high)", total, c, h)
			}
		}
		return fmt.Sprintf("%d TLS issues found", total)
	case "crypto-random":
		total := getIntFromMap(summary, "total_findings")
		if total == 0 {
			return "no weak RNG found"
		}
		// Show by severity
		if sev, ok := summary["by_severity"].(map[string]interface{}); ok {
			h := 0
			m := 0
			if v, ok := sev["high"].(float64); ok {
				h = int(v)
			}
			if v, ok := sev["medium"].(float64); ok {
				m = int(v)
			}
			if h > 0 || m > 0 {
				return fmt.Sprintf("%d weak RNG findings (%d high, %d medium)", total, h, m)
			}
		}
		return fmt.Sprintf("%d weak RNG findings", total)

	case "git":
		totalCommits := getIntFromMap(summary, "total_commits")
		contributors := getIntFromMap(summary, "total_contributors")
		busFactor := getIntFromMap(summary, "bus_factor")
		activity := ""
		if v, ok := summary["activity_level"].(string); ok {
			activity = v
		}
		if totalCommits == 0 {
			return "no commits (shallow clone)"
		}
		return fmt.Sprintf("%d commits, %d contributors, bus factor %d, %s",
			totalCommits, contributors, busFactor, activity)

	case "tech-debt":
		total := getIntFromMap(summary, "total_markers")
		if total == 0 {
			return "no tech debt markers"
		}
		files := getIntFromMap(summary, "files_affected")
		// Show by type
		if types, ok := summary["by_type"].(map[string]interface{}); ok && len(types) > 0 {
			parts := make([]string, 0, len(types))
			for t, count := range types {
				if c, ok := count.(float64); ok && c > 0 {
					parts = append(parts, fmt.Sprintf("%d %s", int(c), t))
				}
			}
			if len(parts) > 0 {
				return fmt.Sprintf("%d markers in %d files (%s)", total, files, strings.Join(parts, ", "))
			}
		}
		return fmt.Sprintf("%d markers in %d files", total, files)

	case "documentation":
		score := getIntFromMap(summary, "overall_score")
		hasReadme := false
		if v, ok := summary["has_readme"].(bool); ok {
			hasReadme = v
		}
		ratio := 0.0
		if v, ok := summary["documentation_ratio"].(float64); ok {
			ratio = v
		}
		if !hasReadme {
			return fmt.Sprintf("score %d%%, no README", score)
		}
		return fmt.Sprintf("score %d%%, %.0f%% files documented", score, ratio)

	case "test-coverage":
		coverage := getIntFromMap(summary, "overall_coverage")
		totalTests := getIntFromMap(summary, "total_tests")
		framework := ""
		if v, ok := summary["test_framework"].(string); ok {
			framework = v
		}
		if framework == "" && totalTests == 0 {
			return "no tests detected"
		}
		if framework != "" {
			return fmt.Sprintf("%d%% coverage, %d tests (%s)", coverage, totalTests, framework)
		}
		return fmt.Sprintf("%d%% coverage, %d tests", coverage, totalTests)

	case "code-ownership":
		contributors := getIntFromMap(summary, "total_contributors")
		hasCodeowners := false
		if v, ok := summary["has_codeowners"].(bool); ok {
			hasCodeowners = v
		}
		rules := getIntFromMap(summary, "codeowners_rules")
		if hasCodeowners {
			return fmt.Sprintf("CODEOWNERS: %d rules, %d contributors", rules, contributors)
		}
		if contributors > 0 {
			return fmt.Sprintf("%d contributors, no CODEOWNERS", contributors)
		}
		return "no ownership data"

	case "dora":
		overallClass := ""
		if v, ok := summary["overall_class"].(string); ok {
			overallClass = v
		}
		deployFreq := getIntFromMap(summary, "deployment_frequency")
		leadTime := getIntFromMap(summary, "lead_time_hours")
		if overallClass == "" {
			return "no DORA data"
		}
		return fmt.Sprintf("%s performer, %d deploys/90d, %dh lead time", overallClass, deployFreq, leadTime)

	case "package-provenance":
		total := getIntFromMap(summary, "total_packages")
		verified := getIntFromMap(summary, "verified_count")
		suspicious := getIntFromMap(summary, "suspicious_count")
		if total == 0 {
			return "no packages to verify"
		}
		if suspicious > 0 {
			return fmt.Sprintf("%d/%d verified, %d suspicious", verified, total, suspicious)
		}
		rate := 0.0
		if v, ok := summary["verification_rate"].(float64); ok {
			rate = v
		}
		return fmt.Sprintf("%d/%d verified (%.0f%%)", verified, total, rate)

	case "api-security":
		total := getIntFromMap(summary, "total_findings")
		if total == 0 {
			return "no API security issues"
		}
		c := getIntFromMap(summary, "critical")
		h := getIntFromMap(summary, "high")
		m := getIntFromMap(summary, "medium")
		return fmt.Sprintf("%d findings (%d critical, %d high, %d medium)", total, c, h, m)

	case "code-vulns":
		total := getIntFromMap(summary, "total_findings")
		if total == 0 {
			return "no code vulnerabilities"
		}
		c := getIntFromMap(summary, "critical")
		h := getIntFromMap(summary, "high")
		m := getIntFromMap(summary, "medium")
		return fmt.Sprintf("%d findings (%d critical, %d high, %d medium)", total, c, h, m)

	case "containers":
		dockerfiles := getIntFromMap(summary, "total_dockerfiles")
		issues := getIntFromMap(summary, "total_issues")
		if dockerfiles == 0 {
			return "no Dockerfiles found"
		}
		if issues == 0 {
			return fmt.Sprintf("%d Dockerfiles analyzed, no issues", dockerfiles)
		}
		// Show by severity if available
		if sev, ok := summary["by_severity"].(map[string]interface{}); ok {
			c := 0
			h := 0
			m := 0
			if v, ok := sev["critical"].(float64); ok {
				c = int(v)
			}
			if v, ok := sev["high"].(float64); ok {
				h = int(v)
			}
			if v, ok := sev["medium"].(float64); ok {
				m = int(v)
			}
			if c > 0 || h > 0 {
				return fmt.Sprintf("%d issues (%d critical, %d high, %d medium)", issues, c, h, m)
			}
		}
		return fmt.Sprintf("%d Dockerfiles, %d issues", dockerfiles, issues)

	case "digital-certificates":
		total := getIntFromMap(summary, "total_certificates")
		expiring := getIntFromMap(summary, "expiring_soon")
		expired := getIntFromMap(summary, "expired")
		if total == 0 {
			return "no certificates found"
		}
		if expired > 0 {
			return fmt.Sprintf("%d certs, %d expired, %d expiring soon", total, expired, expiring)
		}
		if expiring > 0 {
			return fmt.Sprintf("%d certs, %d expiring soon", total, expiring)
		}
		return fmt.Sprintf("%d certificates, all valid", total)

	case "package-bundle-optimization":
		total := getIntFromMap(summary, "total_packages")
		heavy := getIntFromMap(summary, "heavy_packages")
		duplicate := getIntFromMap(summary, "duplicate_packages")
		treeshake := getIntFromMap(summary, "treeshake_candidates")
		if total == 0 {
			return "no packages analyzed"
		}
		issues := heavy + duplicate + treeshake
		if issues == 0 {
			return fmt.Sprintf("%d packages, no optimization issues", total)
		}
		parts := []string{}
		if heavy > 0 {
			parts = append(parts, fmt.Sprintf("%d heavy", heavy))
		}
		if duplicate > 0 {
			parts = append(parts, fmt.Sprintf("%d duplicate", duplicate))
		}
		if treeshake > 0 {
			parts = append(parts, fmt.Sprintf("%d treeshakable", treeshake))
		}
		return fmt.Sprintf("%d packages (%s)", total, strings.Join(parts, ", "))

	case "package-recommendations":
		total := getIntFromMap(summary, "total_recommendations")
		if total == 0 {
			return "no recommendations"
		}
		security := getIntFromMap(summary, "security_recommendations")
		health := getIntFromMap(summary, "health_recommendations")
		if security > 0 {
			return fmt.Sprintf("%d recommendations (%d security, %d health)", total, security, health)
		}
		return fmt.Sprintf("%d package recommendations", total)
	}

	return "complete"
}

func parseJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}
	return json.Unmarshal(data, v)
}

func getIntFromMap(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}
