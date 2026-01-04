// Package supplychain implements the consolidated supply chain security scanner
// This scanner generates SBOMs and performs comprehensive package analysis.
package supplychain

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/core/liveapi"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/scanner/common"
)

const (
	Name    = "supply-chain"
	Version = "4.0.0"
)

func init() {
	scanner.Register(&SupplyChainScanner{})
}

// SupplyChainScanner is the consolidated package analysis super scanner
// It depends on the sbom scanner to provide package data
type SupplyChainScanner struct {
	config FeatureConfig
}

func (s *SupplyChainScanner) Name() string {
	return Name
}

func (s *SupplyChainScanner) Description() string {
	return "Comprehensive package and supply chain security analysis (vulnerabilities, health, malware, licenses, and more) - includes SBOM generation"
}

func (s *SupplyChainScanner) Dependencies() []string {
	// This scanner depends on the sbom scanner output
	return nil
}

func (s *SupplyChainScanner) EstimateDuration(fileCount int) time.Duration {
	base := 5 * time.Second
	if s.config.Vulns.Enabled {
		base += 5 * time.Second
	}
	if s.config.Health.Enabled {
		base += 10 * time.Second
	}
	if s.config.Malcontent.Enabled {
		base += time.Duration(fileCount/2000+2) * time.Second
	}
	if s.config.Reachability.Enabled {
		base += 30 * time.Second
	}
	return base
}

func (s *SupplyChainScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	// Load feature config
	s.config = getFeatureConfig(opts)

	// Ensure output directory
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("creating output directory: %w", err)
		}
	}

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	var sbomPath string
	var components []Component

	// 1. SBOM Generation (internal - this scanner generates its own SBOM)
	if s.config.Generation.Enabled {
		genSummary, genFindings, path, err := runGeneration(ctx, opts, s.config.Generation)
		result.FeaturesRun = append(result.FeaturesRun, "generation")
		if err != nil {
			result.Summary.Errors = append(result.Summary.Errors, fmt.Sprintf("generation: %v", err))
			genSummary = &GenerationSummary{Error: err.Error()}
		}
		result.Summary.Generation = genSummary
		result.Findings.Generation = genFindings
		sbomPath = path
		if genFindings != nil {
			components = genFindings.Components
		}
	}

	// 2. Integrity Verification
	if s.config.Integrity.Enabled && sbomPath != "" {
		intSummary, intFindings, err := runIntegrity(opts.RepoPath, sbomPath, components, s.config.Integrity)
		result.FeaturesRun = append(result.FeaturesRun, "integrity")
		if err != nil {
			result.Summary.Errors = append(result.Summary.Errors, fmt.Sprintf("integrity: %v", err))
			intSummary = &IntegritySummary{Error: err.Error()}
		}
		result.Summary.Integrity = intSummary
		result.Findings.Integrity = intFindings
	}

	// Convert to ComponentData for package analysis features
	var componentData []ComponentData
	for _, c := range components {
		componentData = append(componentData, ComponentData{
			Name:      c.Name,
			Version:   c.Version,
			Purl:      c.Purl,
			Ecosystem: c.Ecosystem,
			Licenses:  c.Licenses,
			Scope:     c.Scope,
		})
	}

	// Parallel features
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 3. Vulnerabilities
	if s.config.Vulns.Enabled && sbomPath != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vulnsResult := s.runVulnsFeature(ctx, opts, sbomPath)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "vulns")
			result.Summary.Vulns = vulnsResult.Summary
			result.Findings.Vulns = vulnsResult.Findings
			mu.Unlock()
		}()
	}

	// 4. Health
	if s.config.Health.Enabled && len(componentData) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			healthResult := s.runHealthFeature(ctx, componentData)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "health")
			result.Summary.Health = healthResult.Summary
			result.Findings.Health = healthResult.Findings
			mu.Unlock()
		}()
	}

	// 5. Licenses
	if s.config.Licenses.Enabled && sbomPath != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			licensesResult := s.runLicensesFeature(sbomPath)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "licenses")
			result.Summary.Licenses = licensesResult.Summary
			result.Findings.Licenses = licensesResult.Findings
			mu.Unlock()
		}()
	}

	// 6. Malcontent
	if s.config.Malcontent.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			malcontentResult := s.runMalcontentFeature(ctx, opts)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "malcontent")
			result.Summary.Malcontent = malcontentResult.Summary
			result.Findings.Malcontent = malcontentResult.Findings
			mu.Unlock()
		}()
	}

	// 7. Dependency Confusion
	if s.config.Confusion.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			confusionResult := s.runConfusionFeature(ctx, opts)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "confusion")
			result.Summary.Confusion = confusionResult.Summary
			result.Findings.Confusion = confusionResult.Findings
			mu.Unlock()
		}()
	}

	// 8. Typosquats
	if s.config.Typosquats.Enabled && len(componentData) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			typosquatsResult := s.runTyposquatsFeature(ctx, componentData)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "typosquats")
			result.Summary.Typosquats = typosquatsResult.Summary
			result.Findings.Typosquats = typosquatsResult.Findings
			mu.Unlock()
		}()
	}

	// 9. Deprecations
	if s.config.Deprecations.Enabled && len(componentData) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deprecationsResult := s.runDeprecationsFeature(ctx, componentData)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "deprecations")
			result.Summary.Deprecations = deprecationsResult.Summary
			result.Findings.Deprecations = deprecationsResult.Findings
			mu.Unlock()
		}()
	}

	// 10. Duplicates
	if s.config.Duplicates.Enabled && len(componentData) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			duplicatesResult := s.runDuplicatesFeature(componentData)
			mu.Lock()
			result.FeaturesRun = append(result.FeaturesRun, "duplicates")
			result.Summary.Duplicates = duplicatesResult.Summary
			result.Findings.Duplicates = duplicatesResult.Findings
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Sequential features

	// 11. Reachability
	if s.config.Reachability.Enabled {
		reachabilityResult := s.runReachabilityFeature(ctx, opts)
		result.FeaturesRun = append(result.FeaturesRun, "reachability")
		result.Summary.Reachability = reachabilityResult.Summary
		result.Findings.Reachability = reachabilityResult.Findings
	}

	// 12. Provenance
	if s.config.Provenance.Enabled && len(componentData) > 0 {
		provenanceResult := s.runProvenanceFeature(ctx, componentData)
		result.FeaturesRun = append(result.FeaturesRun, "provenance")
		result.Summary.Provenance = provenanceResult.Summary
		result.Findings.Provenance = provenanceResult.Findings
	}

	// 13. Bundle (npm only)
	if s.config.Bundle.Enabled {
		bundleResult := s.runBundleFeature(ctx, opts)
		if bundleResult != nil {
			result.FeaturesRun = append(result.FeaturesRun, "bundle")
			result.Summary.Bundle = bundleResult.Summary
			result.Findings.Bundle = bundleResult.Findings
		}
	}

	// 14. Recommendations
	if s.config.Recommendations.Enabled {
		recommendationsResult := s.runRecommendationsFeature(result)
		result.FeaturesRun = append(result.FeaturesRun, "recommendations")
		result.Summary.Recommendations = recommendationsResult.Summary
		result.Findings.Recommendations = recommendationsResult.Findings
	}

	// Create scan result
	scanResult := scanner.NewScanResult(Name, Version, start)
	scanResult.Repository = opts.RepoPath
	scanResult.SetSummary(result.Summary)
	scanResult.SetFindings(result)
	scanResult.SetMetadata(map[string]interface{}{
		"features_run":    result.FeaturesRun,
		"sbom_source":     "internal",
		"component_count": len(componentData),
	})

	// Write result
	if opts.OutputDir != "" {
		resultFile := filepath.Join(opts.OutputDir, Name+".json")
		if err := scanResult.WriteJSON(resultFile); err != nil {
			return nil, fmt.Errorf("writing result: %w", err)
		}

	}

	return scanResult, nil
}

// getFeatureConfig loads feature configuration from scan options
func getFeatureConfig(opts *scanner.ScanOptions) FeatureConfig {
	if opts.FeatureConfig == nil {
		return DefaultConfig()
	}

	cfg := DefaultConfig()

	// Parse generation config
	if genCfg, ok := opts.FeatureConfig["generation"].(map[string]interface{}); ok {
		if v, ok := genCfg["enabled"].(bool); ok {
			cfg.Generation.Enabled = v
		}
		if v, ok := genCfg["tool"].(string); ok {
			cfg.Generation.Tool = v
		}
		if v, ok := genCfg["spec_version"].(string); ok {
			cfg.Generation.SpecVersion = v
		}
		if v, ok := genCfg["fallback_to_syft"].(bool); ok {
			cfg.Generation.FallbackToSyft = v
		}
		if v, ok := genCfg["include_dev"].(bool); ok {
			cfg.Generation.IncludeDev = v
		}
		if v, ok := genCfg["deep"].(bool); ok {
			cfg.Generation.Deep = v
		}
	}

	// Parse integrity config
	if intCfg, ok := opts.FeatureConfig["integrity"].(map[string]interface{}); ok {
		if v, ok := intCfg["enabled"].(bool); ok {
			cfg.Integrity.Enabled = v
		}
		if v, ok := intCfg["verify_lockfiles"].(bool); ok {
			cfg.Integrity.VerifyLockfiles = v
		}
		if v, ok := intCfg["detect_drift"].(bool); ok {
			cfg.Integrity.DetectDrift = v
		}
		if v, ok := intCfg["check_completeness"].(bool); ok {
			cfg.Integrity.CheckCompleteness = v
		}
	}

	// Parse vulns config
	if vulnsCfg, ok := opts.FeatureConfig["vulns"].(map[string]interface{}); ok {
		if v, ok := vulnsCfg["enabled"].(bool); ok {
			cfg.Vulns.Enabled = v
		}
		if v, ok := vulnsCfg["include_kev"].(bool); ok {
			cfg.Vulns.IncludeKEV = v
		}
	}

	// Parse health config
	if healthCfg, ok := opts.FeatureConfig["health"].(map[string]interface{}); ok {
		if v, ok := healthCfg["enabled"].(bool); ok {
			cfg.Health.Enabled = v
		}
		if v, ok := healthCfg["max_packages"].(float64); ok {
			cfg.Health.MaxPackages = int(v)
		}
	}

	// Parse licenses config
	if licCfg, ok := opts.FeatureConfig["licenses"].(map[string]interface{}); ok {
		if v, ok := licCfg["enabled"].(bool); ok {
			cfg.Licenses.Enabled = v
		}
	}

	// Parse malcontent config
	if malCfg, ok := opts.FeatureConfig["malcontent"].(map[string]interface{}); ok {
		if v, ok := malCfg["enabled"].(bool); ok {
			cfg.Malcontent.Enabled = v
		}
		if v, ok := malCfg["min_risk_level"].(string); ok {
			cfg.Malcontent.MinRiskLevel = v
		}
	}

	// Parse confusion config
	if confCfg, ok := opts.FeatureConfig["confusion"].(map[string]interface{}); ok {
		if v, ok := confCfg["enabled"].(bool); ok {
			cfg.Confusion.Enabled = v
		}
	}

	// Parse reachability config
	if reachCfg, ok := opts.FeatureConfig["reachability"].(map[string]interface{}); ok {
		if v, ok := reachCfg["enabled"].(bool); ok {
			cfg.Reachability.Enabled = v
		}
	}

	// Parse provenance config
	if provCfg, ok := opts.FeatureConfig["provenance"].(map[string]interface{}); ok {
		if v, ok := provCfg["enabled"].(bool); ok {
			cfg.Provenance.Enabled = v
		}
	}

	// Parse bundle config
	if bundleCfg, ok := opts.FeatureConfig["bundle"].(map[string]interface{}); ok {
		if v, ok := bundleCfg["enabled"].(bool); ok {
			cfg.Bundle.Enabled = v
		}
	}

	// Parse recommendations config
	if recCfg, ok := opts.FeatureConfig["recommendations"].(map[string]interface{}); ok {
		if v, ok := recCfg["enabled"].(bool); ok {
			cfg.Recommendations.Enabled = v
		}
	}

	// Parse typosquats config
	if typoCfg, ok := opts.FeatureConfig["typosquats"].(map[string]interface{}); ok {
		if v, ok := typoCfg["enabled"].(bool); ok {
			cfg.Typosquats.Enabled = v
		}
	}

	// Parse deprecations config
	if depCfg, ok := opts.FeatureConfig["deprecations"].(map[string]interface{}); ok {
		if v, ok := depCfg["enabled"].(bool); ok {
			cfg.Deprecations.Enabled = v
		}
	}

	// Parse duplicates config
	if dupCfg, ok := opts.FeatureConfig["duplicates"].(map[string]interface{}); ok {
		if v, ok := dupCfg["enabled"].(bool); ok {
			cfg.Duplicates.Enabled = v
		}
	}

	return cfg
}

// =============================================================================
// SBOM Generation Feature (absorbed from sbom scanner)
// =============================================================================

func runGeneration(ctx context.Context, opts *scanner.ScanOptions, cfg GenerationConfig) (*GenerationSummary, *GenerationFindings, string, error) {
	// Determine tool to use
	tool := cfg.Tool
	switch tool {
	case "cdxgen":
		if !common.ToolExists("cdxgen") {
			if cfg.FallbackToSyft && common.ToolExists("syft") {
				tool = "syft"
			} else {
				return nil, nil, "", fmt.Errorf("cdxgen not found and no fallback available")
			}
		}
	case "syft":
		if !common.ToolExists("syft") {
			return nil, nil, "", fmt.Errorf("syft not found")
		}
	default: // auto
		tool, _ = common.PreferTool("cdxgen", "syft")
		if tool == "" {
			return nil, nil, "", fmt.Errorf("no SBOM tool found (install cdxgen or syft)")
		}
	}

	// Generate SBOM file path
	sbomFile := filepath.Join(opts.OutputDir, "sbom.cdx.json")

	// Run SBOM generation
	var err error
	switch tool {
	case "cdxgen":
		args := []string{
			"-o", sbomFile,
			"--spec-version", cfg.SpecVersion,
		}
		if cfg.Deep {
			args = append(args, "--deep")
		}
		args = append(args, opts.RepoPath)

		cmdResult, cmdErr := common.RunCommand(ctx, "cdxgen", args...)
		if cmdErr != nil || cmdResult.ExitCode != 0 {
			// Try fallback
			if cfg.FallbackToSyft && common.ToolExists("syft") {
				tool = "syft"
				common.RunCommand(ctx, "syft", "scan", "dir:"+opts.RepoPath, "-o", "cyclonedx-json="+sbomFile)
			} else {
				err = fmt.Errorf("cdxgen failed: %v", cmdErr)
			}
		}
	case "syft":
		_, cmdErr := common.RunCommand(ctx, "syft", "scan", "dir:"+opts.RepoPath, "-o", "cyclonedx-json="+sbomFile)
		if cmdErr != nil {
			err = cmdErr
		}
	}

	if err != nil {
		return nil, nil, "", err
	}

	// Parse the generated SBOM
	sbomData, parseErr := parseSBOM(sbomFile)
	if parseErr != nil {
		return nil, nil, sbomFile, fmt.Errorf("parsing SBOM: %w", parseErr)
	}

	// Build summary
	summary := &GenerationSummary{
		Tool:            tool,
		SpecVersion:     sbomData.SpecVersion,
		TotalComponents: len(sbomData.Components),
		ByType:          make(map[string]int),
		ByEcosystem:     make(map[string]int),
		HasDependencies: len(sbomData.Dependencies) > 0,
		SBOMPath:        sbomFile,
	}

	// Convert to our component format
	components := make([]Component, 0, len(sbomData.Components))
	for _, c := range sbomData.Components {
		summary.ByType[c.Type]++

		ecosystem := extractEcosystemFromPurl(c.Purl)
		if ecosystem != "" {
			summary.ByEcosystem[ecosystem]++
		}

		comp := Component{
			Type:      c.Type,
			Name:      c.Name,
			Version:   c.Version,
			Purl:      c.Purl,
			Ecosystem: ecosystem,
			Scope:     c.Scope,
		}

		// Extract licenses
		for _, lic := range c.Licenses {
			if lic.License.ID != "" {
				comp.Licenses = append(comp.Licenses, lic.License.ID)
			} else if lic.License.Name != "" {
				comp.Licenses = append(comp.Licenses, lic.License.Name)
			} else if lic.Expression != "" {
				comp.Licenses = append(comp.Licenses, lic.Expression)
			}
		}

		// Extract hashes
		for _, h := range c.Hashes {
			comp.Hashes = append(comp.Hashes, Hash{
				Algorithm: h.Alg,
				Content:   h.Content,
			})
		}

		// Extract properties
		for _, p := range c.Properties {
			comp.Properties = append(comp.Properties, Property{
				Name:  p.Name,
				Value: p.Value,
			})
		}

		components = append(components, comp)
	}

	// Convert dependencies
	dependencies := make([]Dependency, 0, len(sbomData.Dependencies))
	for _, d := range sbomData.Dependencies {
		dependencies = append(dependencies, Dependency{
			Ref:       d.Ref,
			DependsOn: d.DependsOn,
		})
	}

	findings := &GenerationFindings{
		Components:   components,
		Dependencies: dependencies,
		Metadata: &SBOMMetadata{
			BomFormat:    sbomData.BomFormat,
			SpecVersion:  sbomData.SpecVersion,
			Version:      sbomData.Version,
			SerialNumber: sbomData.SerialNumber,
			Timestamp:    sbomData.Metadata.Timestamp,
			Tool:         tool,
		},
	}

	return summary, findings, sbomFile, nil
}

func parseSBOM(sbomPath string) (*CycloneDXBOM, error) {
	data, err := os.ReadFile(sbomPath)
	if err != nil {
		return nil, err
	}

	var bom CycloneDXBOM
	if err := json.Unmarshal(data, &bom); err != nil {
		return nil, err
	}

	return &bom, nil
}

func extractEcosystemFromPurl(purl string) string {
	if purl == "" {
		return ""
	}
	// purl format: pkg:type/namespace/name@version
	if !strings.HasPrefix(purl, "pkg:") {
		return ""
	}
	purl = strings.TrimPrefix(purl, "pkg:")
	parts := strings.SplitN(purl, "/", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// =============================================================================
// SBOM Integrity Feature
// =============================================================================

func runIntegrity(repoPath, sbomPath string, components []Component, cfg IntegrityConfig) (*IntegritySummary, *IntegrityFindings, error) {
	summary := &IntegritySummary{
		IsComplete: true,
	}
	findings := &IntegrityFindings{}

	if cfg.VerifyLockfiles {
		comparisons := verifyAgainstLockfiles(repoPath, components)
		findings.LockfileComparisons = comparisons
		summary.LockfilesFound = len(comparisons)

		for _, comp := range comparisons {
			summary.MissingPackages += comp.Missing
			summary.ExtraPackages += comp.Extra
		}

		if summary.MissingPackages > 0 {
			summary.IsComplete = false
		}
	}

	// Detect drift from previous SBOM
	if cfg.DetectDrift {
		previousPath := cfg.PreviousSBOMPath
		if previousPath == "" {
			// Try to find previous SBOM in standard location
			previousPath = filepath.Join(filepath.Dir(sbomPath), "sbom.cdx.json.previous")
		}

		if _, err := os.Stat(previousPath); err == nil {
			drift, err := compareSBOMs(previousPath, sbomPath, components)
			if err == nil {
				findings.DriftDetails = drift
				summary.DriftDetected = drift.TotalAdded > 0 || drift.TotalRemoved > 0 || drift.TotalChanged > 0
			}
		}
	}

	return summary, findings, nil
}

// compareSBOMs compares two SBOMs and returns the differences
func compareSBOMs(previousPath, currentPath string, currentComponents []Component) (*DriftDetails, error) {
	// Load previous SBOM
	previousFindings, err := LoadSBOM(previousPath)
	if err != nil {
		return nil, fmt.Errorf("loading previous SBOM: %w", err)
	}

	drift := &DriftDetails{
		PreviousSBOMPath: previousPath,
	}

	// Build maps for comparison
	previousMap := make(map[string]Component) // name -> component
	currentMap := make(map[string]Component)

	for _, c := range previousFindings.Components {
		key := c.Ecosystem + "/" + c.Name
		previousMap[key] = c
	}

	for _, c := range currentComponents {
		key := c.Ecosystem + "/" + c.Name
		currentMap[key] = c
	}

	// Find added components (in current but not in previous)
	for key, current := range currentMap {
		if _, exists := previousMap[key]; !exists {
			drift.Added = append(drift.Added, ComponentDiff{
				Name:      current.Name,
				Version:   current.Version,
				Ecosystem: current.Ecosystem,
			})
		}
	}

	// Find removed components (in previous but not in current)
	for key, previous := range previousMap {
		if _, exists := currentMap[key]; !exists {
			drift.Removed = append(drift.Removed, ComponentDiff{
				Name:      previous.Name,
				Version:   previous.Version,
				Ecosystem: previous.Ecosystem,
			})
		}
	}

	// Find version changes
	for key, current := range currentMap {
		if previous, exists := previousMap[key]; exists {
			if previous.Version != current.Version {
				drift.VersionChanged = append(drift.VersionChanged, VersionChange{
					Name:       current.Name,
					Ecosystem:  current.Ecosystem,
					OldVersion: previous.Version,
					NewVersion: current.Version,
				})
			}
		}
	}

	drift.TotalAdded = len(drift.Added)
	drift.TotalRemoved = len(drift.Removed)
	drift.TotalChanged = len(drift.VersionChanged)

	return drift, nil
}

func verifyAgainstLockfiles(repoPath string, components []Component) []LockfileComparison {
	var comparisons []LockfileComparison

	// Build component map by ecosystem
	compMap := make(map[string]map[string]string) // ecosystem -> name -> version
	for _, c := range components {
		eco := c.Ecosystem
		if eco == "" {
			continue
		}
		if compMap[eco] == nil {
			compMap[eco] = make(map[string]string)
		}
		compMap[eco][c.Name] = c.Version
	}

	// Find all lockfiles recursively
	lockfiles := findLockfilesRecursive(repoPath)

	// Process each lockfile
	for _, lockPath := range lockfiles {
		var comp LockfileComparison
		base := filepath.Base(lockPath)

		switch base {
		case "package-lock.json":
			comp = compareLockfile(lockPath, "npm", compMap["npm"])
		case "yarn.lock":
			comp = compareYarnLock(lockPath, compMap["npm"])
		case "pnpm-lock.yaml":
			comp = comparePnpmLock(lockPath, compMap["npm"])
		case "go.sum":
			comp = compareGoSum(lockPath, compMap["golang"])
		case "requirements.txt":
			comp = compareRequirements(lockPath, compMap["pypi"])
		case "poetry.lock":
			comp = comparePoetryLock(lockPath, compMap["pypi"])
		case "uv.lock":
			comp = compareUvLock(lockPath, compMap["pypi"])
		case "Cargo.lock":
			comp = compareCargoLock(lockPath, compMap["cargo"])
		case "Gemfile.lock":
			comp = compareGemfileLock(lockPath, compMap["gem"])
		case "composer.lock":
			comp = compareComposerLock(lockPath, compMap["composer"])
		case "packages.lock.json":
			comp = compareNugetLock(lockPath, compMap["nuget"])
		case "gradle.lockfile":
			comp = compareGradleLock(lockPath, compMap["maven"])
		default:
			continue
		}

		// Include relative path for subdirectory lockfiles
		relPath, _ := filepath.Rel(repoPath, lockPath)
		if relPath != base {
			comp.Lockfile = relPath
		}

		comparisons = append(comparisons, comp)
	}

	return comparisons
}

// findLockfilesRecursive finds all supported lockfiles in a directory tree
func findLockfilesRecursive(root string) []string {
	var lockfiles []string

	// Supported lockfile names
	lockfileNames := map[string]bool{
		"package-lock.json":  true,
		"yarn.lock":          true,
		"pnpm-lock.yaml":     true,
		"go.sum":             true,
		"requirements.txt":   true,
		"poetry.lock":        true,
		"uv.lock":            true,
		"Cargo.lock":         true,
		"Gemfile.lock":       true,
		"composer.lock":      true,
		"packages.lock.json": true,
		"gradle.lockfile":    true,
	}

	// Directories to skip
	skipDirs := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".git":         true,
		".svn":         true,
		"target":       true,
		"build":        true,
		"dist":         true,
		".venv":        true,
		"venv":         true,
		"__pycache__":  true,
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip excluded directories
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a lockfile
		if lockfileNames[info.Name()] {
			lockfiles = append(lockfiles, path)
		}

		return nil
	})

	return lockfiles
}

func compareLockfile(lockPath, ecosystem string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: ecosystem,
		InSBOM:    len(sbomPkgs),
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		return comp
	}

	var lockData struct {
		Packages map[string]struct {
			Version string `json:"version"`
		} `json:"packages"`
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal(data, &lockData); err != nil {
		return comp
	}

	lockPkgs := make(map[string]string)

	// npm v2+ format (packages)
	for name, pkg := range lockData.Packages {
		if name == "" || strings.HasPrefix(name, "node_modules/") {
			// Extract package name from path
			parts := strings.Split(name, "node_modules/")
			if len(parts) > 1 {
				name = parts[len(parts)-1]
			}
		}
		if name != "" && pkg.Version != "" {
			lockPkgs[name] = pkg.Version
		}
	}

	// npm v1 format (dependencies)
	for name, pkg := range lockData.Dependencies {
		if pkg.Version != "" {
			lockPkgs[name] = pkg.Version
		}
	}

	comp.InLockfile = len(lockPkgs)

	// Compare
	for name := range sbomPkgs {
		if _, ok := lockPkgs[name]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		if _, ok := sbomPkgs[name]; !ok {
			comp.Missing++
		}
	}

	return comp
}

func compareYarnLock(lockPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: "npm",
		InSBOM:    len(sbomPkgs),
	}

	file, err := os.Open(lockPath)
	if err != nil {
		return comp
	}
	defer file.Close()

	lockPkgs := make(map[string]string)
	scanner := bufio.NewScanner(file)
	var currentPkg string

	for scanner.Scan() {
		line := scanner.Text()

		// Package line: "package-name@version":
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, " ") {
			// Extract package name
			name := strings.TrimSuffix(line, ":")
			name = strings.Trim(name, "\"")
			// Remove version constraint
			if idx := strings.LastIndex(name, "@"); idx > 0 {
				name = name[:idx]
			}
			currentPkg = name
		}

		// Version line: "  version "x.x.x""
		if strings.HasPrefix(line, "  version ") && currentPkg != "" {
			version := strings.TrimPrefix(line, "  version ")
			version = strings.Trim(version, "\"")
			lockPkgs[currentPkg] = version
		}
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		if _, ok := lockPkgs[name]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		if _, ok := sbomPkgs[name]; !ok {
			comp.Missing++
		}
	}

	return comp
}

func compareGoSum(sumPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(sumPath),
		Ecosystem: "golang",
		InSBOM:    len(sbomPkgs),
	}

	file, err := os.Open(sumPath)
	if err != nil {
		return comp
	}
	defer file.Close()

	lockPkgs := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			version := strings.TrimSuffix(parts[1], "/go.mod")
			version = strings.TrimPrefix(version, "v")
			lockPkgs[name] = version
		}
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		if _, ok := lockPkgs[name]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		if _, ok := sbomPkgs[name]; !ok {
			comp.Missing++
		}
	}

	return comp
}

func compareRequirements(reqPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(reqPath),
		Ecosystem: "pypi",
		InSBOM:    len(sbomPkgs),
	}

	file, err := os.Open(reqPath)
	if err != nil {
		return comp
	}
	defer file.Close()

	lockPkgs := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle various formats: pkg==1.0, pkg>=1.0, pkg
		var name, version string
		for _, sep := range []string{"==", ">=", "<=", "~=", "!=", ">", "<"} {
			if idx := strings.Index(line, sep); idx > 0 {
				name = strings.TrimSpace(line[:idx])
				version = strings.TrimSpace(line[idx+len(sep):])
				break
			}
		}
		if name == "" {
			name = line
		}

		// Normalize name (PEP 503)
		name = strings.ToLower(strings.ReplaceAll(name, "_", "-"))
		lockPkgs[name] = version
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		normalizedName := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
		if _, ok := lockPkgs[normalizedName]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		found := false
		for sbomName := range sbomPkgs {
			if strings.ToLower(strings.ReplaceAll(sbomName, "_", "-")) == name {
				found = true
				break
			}
		}
		if !found {
			comp.Missing++
		}
	}

	return comp
}

// comparePnpmLock parses pnpm-lock.yaml and compares with SBOM
func comparePnpmLock(lockPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: "npm",
		InSBOM:    len(sbomPkgs),
	}

	file, err := os.Open(lockPath)
	if err != nil {
		return comp
	}
	defer file.Close()

	lockPkgs := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// pnpm-lock.yaml v9 format: package@version: under packages:
	// pnpm-lock.yaml v6 format: /package/version: under packages:
	inPackages := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "packages:") {
			inPackages = true
			continue
		}

		if inPackages && !strings.HasPrefix(line, " ") && line != "" {
			inPackages = false
		}

		if inPackages {
			line = strings.TrimSpace(line)
			// v9 format: 'package@version':
			if strings.HasSuffix(line, ":") && strings.Contains(line, "@") {
				entry := strings.TrimSuffix(line, ":")
				entry = strings.Trim(entry, "'\"")
				// Handle scoped packages: @scope/name@version
				atIdx := strings.LastIndex(entry, "@")
				if atIdx > 0 {
					name := entry[:atIdx]
					version := entry[atIdx+1:]
					lockPkgs[name] = version
				}
			}
			// v6 format: /package/version:
			if strings.HasPrefix(line, "/") && strings.HasSuffix(line, ":") {
				entry := strings.TrimPrefix(line, "/")
				entry = strings.TrimSuffix(entry, ":")
				parts := strings.Split(entry, "/")
				if len(parts) >= 2 {
					// Handle scoped packages
					if strings.HasPrefix(parts[0], "@") && len(parts) >= 3 {
						name := parts[0] + "/" + parts[1]
						version := parts[2]
						lockPkgs[name] = version
					} else {
						name := parts[0]
						version := parts[1]
						lockPkgs[name] = version
					}
				}
			}
		}
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		if _, ok := lockPkgs[name]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		if _, ok := sbomPkgs[name]; !ok {
			comp.Missing++
		}
	}

	return comp
}

// comparePoetryLock parses poetry.lock (TOML) and compares with SBOM
func comparePoetryLock(lockPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: "pypi",
		InSBOM:    len(sbomPkgs),
	}

	file, err := os.Open(lockPath)
	if err != nil {
		return comp
	}
	defer file.Close()

	lockPkgs := make(map[string]string)
	scanner := bufio.NewScanner(file)

	var currentName, currentVersion string
	inPackage := false

	for scanner.Scan() {
		line := scanner.Text()

		// Start of package block
		if line == "[[package]]" {
			if currentName != "" && currentVersion != "" {
				lockPkgs[strings.ToLower(strings.ReplaceAll(currentName, "_", "-"))] = currentVersion
			}
			currentName = ""
			currentVersion = ""
			inPackage = true
			continue
		}

		if inPackage {
			if strings.HasPrefix(line, "name = ") {
				currentName = strings.Trim(strings.TrimPrefix(line, "name = "), "\"")
			}
			if strings.HasPrefix(line, "version = ") {
				currentVersion = strings.Trim(strings.TrimPrefix(line, "version = "), "\"")
			}
		}
	}

	// Don't forget the last package
	if currentName != "" && currentVersion != "" {
		lockPkgs[strings.ToLower(strings.ReplaceAll(currentName, "_", "-"))] = currentVersion
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		normalizedName := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
		if _, ok := lockPkgs[normalizedName]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		found := false
		for sbomName := range sbomPkgs {
			if strings.ToLower(strings.ReplaceAll(sbomName, "_", "-")) == name {
				found = true
				break
			}
		}
		if !found {
			comp.Missing++
		}
	}

	return comp
}

// compareUvLock parses uv.lock (TOML) and compares with SBOM
func compareUvLock(lockPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: "pypi",
		InSBOM:    len(sbomPkgs),
	}

	file, err := os.Open(lockPath)
	if err != nil {
		return comp
	}
	defer file.Close()

	lockPkgs := make(map[string]string)
	scanner := bufio.NewScanner(file)

	var currentName, currentVersion string
	inPackage := false

	for scanner.Scan() {
		line := scanner.Text()

		// Start of package block
		if line == "[[package]]" {
			if currentName != "" && currentVersion != "" {
				lockPkgs[strings.ToLower(strings.ReplaceAll(currentName, "_", "-"))] = currentVersion
			}
			currentName = ""
			currentVersion = ""
			inPackage = true
			continue
		}

		if inPackage {
			if strings.HasPrefix(line, "name = ") {
				currentName = strings.Trim(strings.TrimPrefix(line, "name = "), "\"")
			}
			if strings.HasPrefix(line, "version = ") {
				currentVersion = strings.Trim(strings.TrimPrefix(line, "version = "), "\"")
			}
		}
	}

	// Don't forget the last package
	if currentName != "" && currentVersion != "" {
		lockPkgs[strings.ToLower(strings.ReplaceAll(currentName, "_", "-"))] = currentVersion
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		normalizedName := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
		if _, ok := lockPkgs[normalizedName]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		found := false
		for sbomName := range sbomPkgs {
			if strings.ToLower(strings.ReplaceAll(sbomName, "_", "-")) == name {
				found = true
				break
			}
		}
		if !found {
			comp.Missing++
		}
	}

	return comp
}

// compareCargoLock parses Cargo.lock (TOML) and compares with SBOM
func compareCargoLock(lockPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: "cargo",
		InSBOM:    len(sbomPkgs),
	}

	file, err := os.Open(lockPath)
	if err != nil {
		return comp
	}
	defer file.Close()

	lockPkgs := make(map[string]string)
	scanner := bufio.NewScanner(file)

	var currentName, currentVersion string
	inPackage := false

	for scanner.Scan() {
		line := scanner.Text()

		// Start of package block
		if line == "[[package]]" {
			if currentName != "" && currentVersion != "" {
				lockPkgs[currentName] = currentVersion
			}
			currentName = ""
			currentVersion = ""
			inPackage = true
			continue
		}

		if inPackage {
			if strings.HasPrefix(line, "name = ") {
				currentName = strings.Trim(strings.TrimPrefix(line, "name = "), "\"")
			}
			if strings.HasPrefix(line, "version = ") {
				currentVersion = strings.Trim(strings.TrimPrefix(line, "version = "), "\"")
			}
		}
	}

	// Don't forget the last package
	if currentName != "" && currentVersion != "" {
		lockPkgs[currentName] = currentVersion
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		if _, ok := lockPkgs[name]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		if _, ok := sbomPkgs[name]; !ok {
			comp.Missing++
		}
	}

	return comp
}

// compareGemfileLock parses Gemfile.lock and compares with SBOM
func compareGemfileLock(lockPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: "gem",
		InSBOM:    len(sbomPkgs),
	}

	file, err := os.Open(lockPath)
	if err != nil {
		return comp
	}
	defer file.Close()

	lockPkgs := make(map[string]string)
	scanner := bufio.NewScanner(file)

	inSpecs := false
	for scanner.Scan() {
		line := scanner.Text()

		// GEM section with specs
		if strings.TrimSpace(line) == "specs:" {
			inSpecs = true
			continue
		}

		// End of specs section (blank line or new section)
		if inSpecs && (line == "" || !strings.HasPrefix(line, "    ")) {
			if !strings.HasPrefix(line, " ") {
				inSpecs = false
			}
			continue
		}

		// Parse gem entries: "    gem-name (version)"
		if inSpecs && strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "      ") {
			line = strings.TrimSpace(line)
			if idx := strings.Index(line, " ("); idx > 0 {
				name := line[:idx]
				version := strings.TrimSuffix(line[idx+2:], ")")
				lockPkgs[name] = version
			}
		}
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		if _, ok := lockPkgs[name]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		if _, ok := sbomPkgs[name]; !ok {
			comp.Missing++
		}
	}

	return comp
}

// compareComposerLock parses composer.lock (JSON) and compares with SBOM
func compareComposerLock(lockPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: "composer",
		InSBOM:    len(sbomPkgs),
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		return comp
	}

	var lockData struct {
		Packages []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packages"`
		PackagesDev []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packages-dev"`
	}

	if err := json.Unmarshal(data, &lockData); err != nil {
		return comp
	}

	lockPkgs := make(map[string]string)

	for _, pkg := range lockData.Packages {
		// Composer versions often have "v" prefix
		version := strings.TrimPrefix(pkg.Version, "v")
		lockPkgs[pkg.Name] = version
	}

	for _, pkg := range lockData.PackagesDev {
		version := strings.TrimPrefix(pkg.Version, "v")
		lockPkgs[pkg.Name] = version
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		if _, ok := lockPkgs[name]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		if _, ok := sbomPkgs[name]; !ok {
			comp.Missing++
		}
	}

	return comp
}

// compareNugetLock parses packages.lock.json (NuGet) and compares with SBOM
func compareNugetLock(lockPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: "nuget",
		InSBOM:    len(sbomPkgs),
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		return comp
	}

	var lockData struct {
		Version      int `json:"version"`
		Dependencies map[string]map[string]struct {
			Type     string `json:"type"`
			Resolved string `json:"resolved"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal(data, &lockData); err != nil {
		return comp
	}

	lockPkgs := make(map[string]string)

	// Dependencies are keyed by target framework, then by package name
	for _, frameworkDeps := range lockData.Dependencies {
		for name, pkg := range frameworkDeps {
			lockPkgs[name] = pkg.Resolved
		}
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		if _, ok := lockPkgs[name]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		if _, ok := sbomPkgs[name]; !ok {
			comp.Missing++
		}
	}

	return comp
}

// compareGradleLock parses gradle.lockfile and compares with SBOM
func compareGradleLock(lockPath string, sbomPkgs map[string]string) LockfileComparison {
	comp := LockfileComparison{
		Lockfile:  filepath.Base(lockPath),
		Ecosystem: "maven",
		InSBOM:    len(sbomPkgs),
	}

	file, err := os.Open(lockPath)
	if err != nil {
		return comp
	}
	defer file.Close()

	lockPkgs := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip the "empty=" line
		if strings.HasPrefix(line, "empty=") {
			continue
		}

		// Format: group:artifact:version=configuration1,configuration2
		if idx := strings.Index(line, "="); idx > 0 {
			coords := line[:idx]
			parts := strings.Split(coords, ":")
			if len(parts) >= 3 {
				// group:artifact:version
				name := parts[0] + ":" + parts[1]
				version := parts[2]
				lockPkgs[name] = version
			}
		}
	}

	comp.InLockfile = len(lockPkgs)

	for name := range sbomPkgs {
		if _, ok := lockPkgs[name]; ok {
			comp.Matched++
		} else {
			comp.Extra++
		}
	}

	for name := range lockPkgs {
		if _, ok := sbomPkgs[name]; !ok {
			comp.Missing++
		}
	}

	return comp
}

// GetSBOMPath returns the path to the SBOM file in the output directory
// This is used by other scanners to locate the SBOM
func GetSBOMPath(outputDir string) string {
	return filepath.Join(outputDir, "sbom.cdx.json")
}

// LoadSBOM loads and parses an existing SBOM file
// This is used by other scanners to consume SBOM data
func LoadSBOM(sbomPath string) (*GenerationFindings, error) {
	bom, err := parseSBOM(sbomPath)
	if err != nil {
		return nil, err
	}

	components := make([]Component, 0, len(bom.Components))
	for _, c := range bom.Components {
		comp := Component{
			Type:      c.Type,
			Name:      c.Name,
			Version:   c.Version,
			Purl:      c.Purl,
			Ecosystem: extractEcosystemFromPurl(c.Purl),
			Scope:     c.Scope,
		}

		for _, lic := range c.Licenses {
			if lic.License.ID != "" {
				comp.Licenses = append(comp.Licenses, lic.License.ID)
			} else if lic.License.Name != "" {
				comp.Licenses = append(comp.Licenses, lic.License.Name)
			}
		}

		components = append(components, comp)
	}

	dependencies := make([]Dependency, 0, len(bom.Dependencies))
	for _, d := range bom.Dependencies {
		dependencies = append(dependencies, Dependency{
			Ref:       d.Ref,
			DependsOn: d.DependsOn,
		})
	}

	return &GenerationFindings{
		Components:   components,
		Dependencies: dependencies,
		Metadata: &SBOMMetadata{
			BomFormat:   bom.BomFormat,
			SpecVersion: bom.SpecVersion,
			Version:     bom.Version,
		},
	}, nil
}

// ==================== Vulns Feature ====================

type vulnsFeatureResult struct {
	Summary  *VulnsSummary
	Findings []VulnFinding
}

func (s *SupplyChainScanner) runVulnsFeature(ctx context.Context, opts *scanner.ScanOptions, sbomPath string) *vulnsFeatureResult {
	result := &vulnsFeatureResult{
		Summary:  &VulnsSummary{},
		Findings: []VulnFinding{},
	}

	if !common.ToolExists("osv-scanner") {
		result.Summary.Error = "osv-scanner not installed"
		return result
	}

	// Run osv-scanner against the SBOM
	var cmdResult *common.CommandResult
	if _, err := os.Stat(sbomPath); err == nil {
		cmdResult, _ = common.RunCommand(ctx, "osv-scanner", "scan", "source", "--format=json", "-S", sbomPath)
	} else {
		cmdResult, _ = common.RunCommand(ctx, "osv-scanner", "scan", "source", "--format=json", "-r", opts.RepoPath)
	}

	if cmdResult == nil || len(cmdResult.Stdout) == 0 {
		return result
	}

	// Parse output
	var output struct {
		Results []struct {
			Packages []struct {
				Package struct {
					Name      string `json:"name"`
					Version   string `json:"version"`
					Ecosystem string `json:"ecosystem"`
				} `json:"package"`
				Vulnerabilities []struct {
					ID       string   `json:"id"`
					Aliases  []string `json:"aliases"`
					Summary  string   `json:"summary"`
					Severity []struct {
						Type  string `json:"type"`
						Score string `json:"score"`
					} `json:"severity"`
				} `json:"vulnerabilities"`
			} `json:"packages"`
		} `json:"results"`
	}

	if json.Unmarshal(cmdResult.Stdout, &output) != nil {
		return result
	}

	seen := make(map[string]bool)
	for _, r := range output.Results {
		for _, pkg := range r.Packages {
			for _, vuln := range pkg.Vulnerabilities {
				key := fmt.Sprintf("%s:%s:%s", vuln.ID, pkg.Package.Name, pkg.Package.Version)
				if seen[key] {
					continue
				}
				seen[key] = true

				severity := "medium"
				for _, sev := range vuln.Severity {
					if sev.Type == "CVSS_V3" {
						var score float64
						fmt.Sscanf(sev.Score, "%f", &score)
						if score >= 9.0 {
							severity = "critical"
						} else if score >= 7.0 {
							severity = "high"
						} else if score >= 4.0 {
							severity = "medium"
						} else {
							severity = "low"
						}
						break
					}
				}

				result.Findings = append(result.Findings, VulnFinding{
					ID:        vuln.ID,
					Aliases:   vuln.Aliases,
					Package:   pkg.Package.Name,
					Version:   pkg.Package.Version,
					Ecosystem: pkg.Package.Ecosystem,
					Severity:  severity,
					Title:     vuln.Summary,
				})

				result.Summary.TotalVulnerabilities++
				switch severity {
				case "critical":
					result.Summary.Critical++
				case "high":
					result.Summary.High++
				case "medium":
					result.Summary.Medium++
				case "low":
					result.Summary.Low++
				}
			}
		}
	}

	// KEV enrichment
	if s.config.Vulns.IncludeKEV {
		kevVulns := fetchKEV(ctx)
		for i := range result.Findings {
			if kevVulns[result.Findings[i].ID] {
				result.Findings[i].InKEV = true
				result.Summary.KEVCount++
			} else {
				for _, alias := range result.Findings[i].Aliases {
					if kevVulns[alias] {
						result.Findings[i].InKEV = true
						result.Summary.KEVCount++
						break
					}
				}
			}
		}
	}

	return result
}

func fetchKEV(ctx context.Context) map[string]bool {
	vulns := make(map[string]bool)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json", nil)
	resp, err := client.Do(req)
	if err != nil {
		return vulns
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var catalog struct {
		Vulnerabilities []struct {
			CVEID string `json:"cveID"`
		} `json:"vulnerabilities"`
	}
	if json.Unmarshal(body, &catalog) == nil {
		for _, v := range catalog.Vulnerabilities {
			vulns[v.CVEID] = true
		}
	}
	return vulns
}

// ==================== Health Feature ====================

type healthFeatureResult struct {
	Summary  *HealthSummary
	Findings []HealthFinding
}

func (s *SupplyChainScanner) runHealthFeature(ctx context.Context, components []ComponentData) *healthFeatureResult {
	result := &healthFeatureResult{
		Summary:  &HealthSummary{},
		Findings: []HealthFinding{},
	}

	// Use deps.dev client for package health enrichment
	depsClient := liveapi.NewDepsDevClient()

	maxPackages := s.config.Health.MaxPackages
	if maxPackages == 0 {
		maxPackages = 50
	}

	packages := components
	if len(packages) > maxPackages {
		packages = packages[:maxPackages]
	}

	result.Summary.TotalPackages = len(packages)

	for _, pkg := range packages {
		if pkg.Purl == "" {
			continue
		}

		finding := HealthFinding{
			Package:   pkg.Name,
			Version:   pkg.Version,
			Ecosystem: pkg.Ecosystem,
			Purl:      pkg.Purl,
		}

		// Query deps.dev for package health data
		details, err := depsClient.GetVersionDetailsByPURL(ctx, pkg.Purl)
		if err != nil {
			finding.Status = "unknown"
			result.Findings = append(result.Findings, finding)
			continue
		}

		// Extract health data from deps.dev response
		finding.IsDeprecated = details.IsDeprecated

		// Get latest version
		latestVersion, err := depsClient.GetLatestVersion(ctx, pkg.Ecosystem, pkg.Name)
		if err == nil {
			finding.LatestVersion = latestVersion
			finding.IsOutdated = latestVersion != "" && latestVersion != pkg.Version
		}

		// Get health score from scorecard
		for _, proj := range details.Projects {
			if proj.Scorecard != nil {
				finding.HealthScore = proj.Scorecard.OverallScore
				break
			}
		}

		// Determine status based on health metrics
		if finding.IsDeprecated {
			finding.Status = "critical"
			result.Summary.CriticalCount++
			result.Summary.DeprecatedCount++
		} else if finding.HealthScore < 5 && finding.HealthScore > 0 {
			finding.Status = "warning"
			result.Summary.WarningCount++
		} else {
			finding.Status = "healthy"
			result.Summary.HealthyCount++
		}
		if finding.IsOutdated {
			result.Summary.OutdatedCount++
		}
		result.Summary.AnalyzedCount++

		result.Findings = append(result.Findings, finding)
	}

	return result
}

// ==================== Licenses Feature ====================

type licensesFeatureResult struct {
	Summary  *LicensesSummary
	Findings []LicenseFinding
}

var (
	allowedLicenses = map[string]bool{
		"MIT": true, "Apache-2.0": true, "BSD-2-Clause": true, "BSD-3-Clause": true,
		"ISC": true, "Unlicense": true, "CC0-1.0": true, "0BSD": true,
	}
	deniedLicenses = map[string]bool{
		"GPL-2.0": true, "GPL-2.0-only": true, "GPL-3.0": true, "GPL-3.0-only": true,
		"AGPL-3.0": true, "AGPL-3.0-only": true, "SSPL-1.0": true,
	}
)

func (s *SupplyChainScanner) runLicensesFeature(sbomPath string) *licensesFeatureResult {
	result := &licensesFeatureResult{
		Summary:  &LicensesSummary{LicenseCounts: make(map[string]int)},
		Findings: []LicenseFinding{},
	}

	data, err := os.ReadFile(sbomPath)
	if err != nil {
		result.Summary.Error = err.Error()
		return result
	}

	var sbomDoc struct {
		Components []struct {
			Name     string `json:"name"`
			Version  string `json:"version"`
			Purl     string `json:"purl"`
			Licenses []struct {
				License struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"license"`
				Expression string `json:"expression"`
			} `json:"licenses"`
		} `json:"components"`
	}

	if json.Unmarshal(data, &sbomDoc) != nil {
		return result
	}

	uniqueLicenses := make(map[string]bool)

	for _, c := range sbomDoc.Components {
		result.Summary.TotalPackages++

		var licenseIDs []string
		for _, lic := range c.Licenses {
			if lic.License.ID != "" {
				licenseIDs = append(licenseIDs, lic.License.ID)
			} else if lic.License.Name != "" {
				licenseIDs = append(licenseIDs, lic.License.Name)
			} else if lic.Expression != "" {
				licenseIDs = append(licenseIDs, lic.Expression)
			}
		}

		status := "unknown"
		if len(licenseIDs) == 0 {
			result.Summary.Unknown++
		} else {
			hasDenied := false
			allAllowed := true

			for _, lic := range licenseIDs {
				uniqueLicenses[lic] = true
				result.Summary.LicenseCounts[lic]++

				if deniedLicenses[lic] {
					hasDenied = true
					allAllowed = false
				} else if !allowedLicenses[lic] {
					allAllowed = false
				}
			}

			if hasDenied {
				status = "denied"
				result.Summary.Denied++
				result.Summary.PolicyViolations++
			} else if allAllowed {
				status = "allowed"
				result.Summary.Allowed++
			} else {
				status = "review"
				result.Summary.NeedsReview++
			}
		}

		result.Findings = append(result.Findings, LicenseFinding{
			Package:   c.Name,
			Version:   c.Version,
			Ecosystem: extractEcosystem(c.Purl),
			Licenses:  licenseIDs,
			Status:    status,
		})
	}

	result.Summary.UniqueLicenses = len(uniqueLicenses)
	return result
}

// ==================== Malcontent Feature ====================

type malcontentFeatureResult struct {
	Summary  *MalcontentSummary
	Findings []MalcontentFinding
}

func (s *SupplyChainScanner) runMalcontentFeature(ctx context.Context, opts *scanner.ScanOptions) *malcontentFeatureResult {
	result := &malcontentFeatureResult{
		Summary:  &MalcontentSummary{},
		Findings: []MalcontentFinding{},
	}

	if !common.ToolExists("mal") {
		result.Summary.Error = "malcontent not installed (brew install malcontent)"
		return result
	}

	minRisk := s.config.Malcontent.MinRiskLevel
	if minRisk == "" {
		minRisk = "medium"
	}

	cmdResult, err := common.RunCommand(ctx, "mal", "analyze", "--format=json", "--min-file-risk="+minRisk, opts.RepoPath)
	if err != nil || cmdResult == nil {
		result.Summary.Error = "malcontent execution failed"
		return result
	}

	var output struct {
		Files map[string]struct {
			RiskScore int    `json:"risk_score"`
			RiskLevel string `json:"risk_level"`
			Behaviors []struct {
				Description string `json:"description"`
				RiskLevel   string `json:"risk_level"`
			} `json:"behaviors"`
		} `json:"files"`
	}

	if json.Unmarshal(cmdResult.Stdout, &output) != nil {
		return result
	}

	result.Summary.TotalFiles = len(output.Files)

	for path, file := range output.Files {
		if file.RiskScore == 0 {
			continue
		}
		result.Summary.FilesWithRisk++

		var behaviors []string
		for _, b := range file.Behaviors {
			behaviors = append(behaviors, b.Description)
			result.Summary.TotalFindings++
			switch strings.ToLower(b.RiskLevel) {
			case "critical":
				result.Summary.Critical++
			case "high":
				result.Summary.High++
			case "medium":
				result.Summary.Medium++
			case "low":
				result.Summary.Low++
			}
		}

		result.Findings = append(result.Findings, MalcontentFinding{
			File:      path,
			Risk:      file.RiskLevel,
			RiskScore: file.RiskScore,
			Behaviors: behaviors,
		})
	}

	return result
}

// ==================== Confusion Feature ====================

type confusionFeatureResult struct {
	Summary  *ConfusionSummary
	Findings []ConfusionFinding
}

func (s *SupplyChainScanner) runConfusionFeature(ctx context.Context, opts *scanner.ScanOptions) *confusionFeatureResult {
	result := &confusionFeatureResult{
		Summary:  &ConfusionSummary{ByEcosystem: make(map[string]int)},
		Findings: []ConfusionFinding{},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// Check npm
	if s.config.Confusion.CheckNPM {
		pkgJSONs := findFiles(opts.RepoPath, "package.json")
		for _, pkgJSON := range pkgJSONs {
			data, _ := os.ReadFile(pkgJSON)
			var pkg struct {
				Dependencies    map[string]string `json:"dependencies"`
				DevDependencies map[string]string `json:"devDependencies"`
			}
			if json.Unmarshal(data, &pkg) != nil {
				continue
			}

			allDeps := make(map[string]string)
			for k, v := range pkg.Dependencies {
				allDeps[k] = v
			}
			for k, v := range pkg.DevDependencies {
				allDeps[k] = v
			}

			for name, version := range allDeps {
				if looksInternal(name) {
					exists, pubVersion := checkNPMPackage(ctx, client, name)
					if exists {
						result.Findings = append(result.Findings, ConfusionFinding{
							Package:       name,
							Ecosystem:     "npm",
							RiskLevel:     "critical",
							RiskType:      "dependency-confusion",
							Description:   "Internal-looking package exists on public npm",
							PublicExists:  true,
							PublicVersion: pubVersion,
							LocalVersion:  version,
							File:          strings.TrimPrefix(pkgJSON, opts.RepoPath+"/"),
						})
						result.Summary.Critical++
						result.Summary.TotalFindings++
						result.Summary.ByEcosystem["npm"]++
					}
				}
			}
		}
	}

	// Check PyPI
	if s.config.Confusion.CheckPyPI {
		reqFiles := findFiles(opts.RepoPath, "requirements.txt")
		for _, reqFile := range reqFiles {
			f, _ := os.Open(reqFile)
			if f == nil {
				continue
			}
			scanner := bufio.NewScanner(f)
			pkgRegex := regexp.MustCompile(`^([a-zA-Z0-9][\w\-\.]*[a-zA-Z0-9])`)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				matches := pkgRegex.FindStringSubmatch(line)
				if len(matches) >= 2 && looksInternal(matches[1]) {
					exists, pubVersion := checkPyPIPackage(ctx, client, matches[1])
					if exists {
						result.Findings = append(result.Findings, ConfusionFinding{
							Package:       matches[1],
							Ecosystem:     "pypi",
							RiskLevel:     "critical",
							RiskType:      "dependency-confusion",
							Description:   "Internal-looking package exists on public PyPI",
							PublicExists:  true,
							PublicVersion: pubVersion,
							File:          strings.TrimPrefix(reqFile, opts.RepoPath+"/"),
						})
						result.Summary.Critical++
						result.Summary.TotalFindings++
						result.Summary.ByEcosystem["pypi"]++
					}
				}
			}
			f.Close()
		}
	}

	return result
}

func looksInternal(name string) bool {
	patterns := []string{"internal-", "private-", "-internal", "-private"}
	nameLower := strings.ToLower(name)
	for _, p := range patterns {
		if strings.Contains(nameLower, p) {
			return true
		}
	}
	return false
}

func checkNPMPackage(ctx context.Context, client *http.Client, name string) (bool, string) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://registry.npmjs.org/"+name, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		DistTags struct {
			Latest string `json:"latest"`
		} `json:"dist-tags"`
	}
	json.Unmarshal(body, &pkg)
	return true, pkg.DistTags.Latest
}

func checkPyPIPackage(ctx context.Context, client *http.Client, name string) (bool, string) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://pypi.org/pypi/"+name+"/json", nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	json.Unmarshal(body, &pkg)
	return true, pkg.Info.Version
}

// ==================== Reachability Feature ====================

type reachabilityFeatureResult struct {
	Summary  *ReachabilitySummary
	Findings []ReachabilityFinding
}

func (s *SupplyChainScanner) runReachabilityFeature(ctx context.Context, opts *scanner.ScanOptions) *reachabilityFeatureResult {
	result := &reachabilityFeatureResult{
		Summary:  &ReachabilitySummary{},
		Findings: []ReachabilityFinding{},
	}

	if !common.ToolExists("osv-scanner") {
		result.Summary.Error = "osv-scanner not installed"
		return result
	}

	// Detect ecosystem
	ecosystem := ""
	if _, err := os.Stat(filepath.Join(opts.RepoPath, "go.mod")); err == nil {
		ecosystem = "Go"
	} else if _, err := os.Stat(filepath.Join(opts.RepoPath, "requirements.txt")); err == nil {
		ecosystem = "Python"
	} else if _, err := os.Stat(filepath.Join(opts.RepoPath, "Cargo.toml")); err == nil {
		ecosystem = "Rust"
	}

	if ecosystem == "" {
		result.Summary.Supported = false
		return result
	}

	result.Summary.Supported = true

	cmdResult, _ := common.RunCommand(ctx, "osv-scanner", "scan", "--format", "json", "--experimental-call-analysis", opts.RepoPath)
	if cmdResult == nil || len(cmdResult.Stdout) == 0 {
		return result
	}

	var output struct {
		Results []struct {
			Packages []struct {
				Package struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				} `json:"package"`
				Vulnerabilities []struct {
					ID       string `json:"id"`
					Summary  string `json:"summary"`
					Analysis *struct {
						Called bool `json:"called"`
					} `json:"analysis,omitempty"`
				} `json:"vulnerabilities"`
			} `json:"packages"`
		} `json:"results"`
	}

	if json.Unmarshal(cmdResult.Stdout, &output) != nil {
		return result
	}

	for _, r := range output.Results {
		for _, pkg := range r.Packages {
			for _, vuln := range pkg.Vulnerabilities {
				status := "unknown"
				reachable := true
				if vuln.Analysis != nil {
					if vuln.Analysis.Called {
						status = "reachable"
					} else {
						status = "unreachable"
						reachable = false
					}
				}

				result.Findings = append(result.Findings, ReachabilityFinding{
					ID:                 vuln.ID,
					Package:            pkg.Package.Name,
					Version:            pkg.Package.Version,
					Summary:            vuln.Summary,
					ReachabilityStatus: status,
					Reachable:          reachable,
				})

				result.Summary.TotalVulns++
				switch status {
				case "reachable":
					result.Summary.ReachableVulns++
				case "unreachable":
					result.Summary.UnreachableVulns++
				default:
					result.Summary.UnknownReachability++
				}
			}
		}
	}

	if result.Summary.TotalVulns > 0 {
		result.Summary.ReductionPercent = float64(result.Summary.UnreachableVulns) / float64(result.Summary.TotalVulns) * 100
	}

	return result
}

// ==================== Provenance Feature ====================

type provenanceFeatureResult struct {
	Summary  *ProvenanceSummary
	Findings []ProvenanceFinding
}

func (s *SupplyChainScanner) runProvenanceFeature(ctx context.Context, components []ComponentData) *provenanceFeatureResult {
	result := &provenanceFeatureResult{
		Summary:  &ProvenanceSummary{},
		Findings: []ProvenanceFinding{},
	}

	result.Summary.TotalPackages = len(components)

	// Use deps.dev client for SLSA provenance data
	depsClient := liveapi.NewDepsDevClient()

	for _, pkg := range components {
		if pkg.Purl == "" {
			result.Summary.UnverifiedCount++
			continue
		}

		// Query deps.dev for SLSA provenance
		provenances, err := depsClient.GetSLSAProvenance(ctx, pkg.Ecosystem, pkg.Name, pkg.Version)
		if err != nil {
			result.Summary.UnverifiedCount++
			result.Findings = append(result.Findings, ProvenanceFinding{
				Package:  pkg.Name,
				Version:  pkg.Version,
				Verified: false,
			})
			continue
		}

		if len(provenances) > 0 {
			// Has provenance attestation
			prov := provenances[0]
			verified := prov.Verified

			finding := ProvenanceFinding{
				Package:  pkg.Name,
				Version:  pkg.Version,
				Verified: verified,
			}

			if prov.SourceRepository != "" {
				finding.Source = prov.SourceRepository
			}
			if prov.URL != "" {
				finding.Attestation = prov.URL
			}

			result.Findings = append(result.Findings, finding)

			if verified {
				result.Summary.VerifiedCount++
			} else {
				// Has provenance but not verified - suspicious
				result.Summary.SuspiciousCount++
			}
		} else {
			// No provenance data
			result.Summary.UnverifiedCount++
			result.Findings = append(result.Findings, ProvenanceFinding{
				Package:  pkg.Name,
				Version:  pkg.Version,
				Verified: false,
			})
		}
	}

	// Calculate verification rate
	if result.Summary.TotalPackages > 0 {
		result.Summary.VerificationRate = float64(result.Summary.VerifiedCount) / float64(result.Summary.TotalPackages) * 100
	}

	return result
}

// ==================== Bundle Feature ====================

type bundleFeatureResult struct {
	Summary  *BundleSummary
	Findings []BundleFinding
}

func (s *SupplyChainScanner) runBundleFeature(ctx context.Context, opts *scanner.ScanOptions) *bundleFeatureResult {
	// Only for npm projects
	if _, err := os.Stat(filepath.Join(opts.RepoPath, "package.json")); err != nil {
		return nil
	}

	result := &bundleFeatureResult{
		Summary:  &BundleSummary{},
		Findings: []BundleFinding{},
	}

	// Placeholder - full implementation would use bundlephobia API
	return result
}

// ==================== Recommendations Feature ====================

type recommendationsFeatureResult struct {
	Summary  *RecommendationsSummary
	Findings []RecommendationFinding
}

func (s *SupplyChainScanner) runRecommendationsFeature(scanResult *Result) *recommendationsFeatureResult {
	result := &recommendationsFeatureResult{
		Summary:  &RecommendationsSummary{},
		Findings: []RecommendationFinding{},
	}

	// Generate recommendations based on vulns and health data
	if scanResult.Summary.Vulns != nil && scanResult.Summary.Vulns.Critical > 0 {
		result.Summary.SecurityRecommendations = scanResult.Summary.Vulns.Critical
	}
	if scanResult.Summary.Health != nil && scanResult.Summary.Health.DeprecatedCount > 0 {
		result.Summary.HealthRecommendations = scanResult.Summary.Health.DeprecatedCount
	}

	result.Summary.TotalRecommendations = result.Summary.SecurityRecommendations + result.Summary.HealthRecommendations

	return result
}

// ==================== Typosquats Feature ====================

type typosquatsFeatureResult struct {
	Summary  *TyposquatsSummary
	Findings []TyposquatFinding
}

func (s *SupplyChainScanner) runTyposquatsFeature(ctx context.Context, components []ComponentData) *typosquatsFeatureResult {
	result := &typosquatsFeatureResult{
		Summary:  &TyposquatsSummary{},
		Findings: []TyposquatFinding{},
	}

	result.Summary.TotalChecked = len(components)

	// Check for packages similar to popular ones
	popularPackages := map[string]bool{
		"lodash": true, "express": true, "react": true, "axios": true,
		"moment": true, "request": true, "async": true, "chalk": true,
		"commander": true, "debug": true, "underscore": true, "bluebird": true,
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, pkg := range components {
		// Check similar names
		if s.config.Typosquats.CheckSimilarNames {
			for popular := range popularPackages {
				if pkg.Name != popular && isSimilar(pkg.Name, popular) {
					result.Findings = append(result.Findings, TyposquatFinding{
						Package:   pkg.Name,
						Ecosystem: pkg.Ecosystem,
						SimilarTo: popular,
						Reason:    fmt.Sprintf("Name similar to popular package '%s'", popular),
						RiskLevel: "medium",
					})
					result.Summary.SuspiciousCount++
				}
			}
		}

		// Check package age
		if s.config.Typosquats.CheckNewPackages && pkg.Ecosystem == "npm" {
			age := getPackageAge(ctx, client, pkg.Name)
			if age >= 0 && age < 30 {
				result.Findings = append(result.Findings, TyposquatFinding{
					Package:   pkg.Name,
					Ecosystem: pkg.Ecosystem,
					AgeInDays: age,
					Reason:    fmt.Sprintf("Package is only %d days old", age),
					RiskLevel: "low",
				})
				result.Summary.NewPackagesCount++
			}
		}
	}

	return result
}

func isSimilar(a, b string) bool {
	if len(a) < 3 || len(b) < 3 {
		return false
	}
	// Simple Levenshtein-like check
	if abs(len(a)-len(b)) > 2 {
		return false
	}
	differences := 0
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			differences++
		}
	}
	differences += abs(len(a) - len(b))
	return differences > 0 && differences <= 2
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getPackageAge(ctx context.Context, client *http.Client, name string) int {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://registry.npmjs.org/"+name, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return -1
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		Time map[string]string `json:"time"`
	}
	if json.Unmarshal(body, &pkg) != nil {
		return -1
	}

	created, ok := pkg.Time["created"]
	if !ok {
		return -1
	}

	t, err := time.Parse(time.RFC3339, created)
	if err != nil {
		return -1
	}

	return int(time.Since(t).Hours() / 24)
}

// ==================== Deprecations Feature ====================

type deprecationsFeatureResult struct {
	Summary  *DeprecationsSummary
	Findings []DeprecationFinding
}

func (s *SupplyChainScanner) runDeprecationsFeature(ctx context.Context, components []ComponentData) *deprecationsFeatureResult {
	result := &deprecationsFeatureResult{
		Summary:  &DeprecationsSummary{ByEcosystem: make(map[string]int)},
		Findings: []DeprecationFinding{},
	}

	result.Summary.TotalPackages = len(components)

	// Use deps.dev client as primary source (cross-ecosystem)
	depsClient := liveapi.NewDepsDevClient()
	httpClient := &http.Client{Timeout: 5 * time.Second}

	for _, pkg := range components {
		var deprecated bool
		var message, alternative string

		// First, try deps.dev (works for all supported ecosystems)
		if pkg.Purl != "" {
			details, err := depsClient.GetVersionDetailsByPURL(ctx, pkg.Purl)
			if err == nil && details.IsDeprecated {
				deprecated = true
				message = "Package version is deprecated"
			}
		}

		// Fallback to ecosystem-specific APIs for richer deprecation messages
		if !deprecated {
			switch pkg.Ecosystem {
			case "npm":
				if s.config.Deprecations.CheckNPM {
					deprecated, message = checkNPMDeprecation(ctx, httpClient, pkg.Name, pkg.Version)
				}
			case "pypi":
				if s.config.Deprecations.CheckPyPI {
					deprecated, message = checkPyPIDeprecation(ctx, httpClient, pkg.Name)
				}
			case "golang", "go":
				if s.config.Deprecations.CheckGo {
					deprecated, message = checkGoDeprecation(ctx, httpClient, pkg.Name, pkg.Version)
				}
			}
		}

		if deprecated {
			result.Findings = append(result.Findings, DeprecationFinding{
				Package:     pkg.Name,
				Version:     pkg.Version,
				Ecosystem:   pkg.Ecosystem,
				Message:     message,
				Alternative: alternative,
			})
			result.Summary.DeprecatedCount++
			result.Summary.ByEcosystem[pkg.Ecosystem]++
		}
	}

	return result
}

func checkNPMDeprecation(ctx context.Context, client *http.Client, name, version string) (bool, string) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://registry.npmjs.org/%s/%s", name, version), nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		Deprecated string `json:"deprecated"`
	}
	if json.Unmarshal(body, &pkg) != nil {
		return false, ""
	}

	return pkg.Deprecated != "", pkg.Deprecated
}

func checkPyPIDeprecation(ctx context.Context, client *http.Client, name string) (bool, string) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://pypi.org/pypi/"+name+"/json", nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var pkg struct {
		Info struct {
			Classifiers []string `json:"classifiers"`
		} `json:"info"`
	}
	if json.Unmarshal(body, &pkg) != nil {
		return false, ""
	}

	for _, c := range pkg.Info.Classifiers {
		// Check for various deprecation indicators in classifiers
		// Development Status :: 7 - Inactive
		// Development Status :: 1 - Planning (sometimes indicates abandoned)
		if strings.Contains(c, "Development Status :: 7") ||
			strings.Contains(c, "Inactive") ||
			strings.Contains(c, "Deprecated") {
			return true, c
		}
	}

	return false, ""
}

// checkGoDeprecation checks if a Go module version is deprecated/retracted
func checkGoDeprecation(ctx context.Context, client *http.Client, modulePath, version string) (bool, string) {
	// Query the Go module proxy for version info
	// The proxy returns retracted status in the .info endpoint
	// Module paths must be escaped: uppercase letters become !lowercase
	escapedPath := escapeModulePath(modulePath)
	proxyURL := fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.info",
		escapedPath,
		version)

	req, err := http.NewRequestWithContext(ctx, "GET", proxyURL, nil)
	if err != nil {
		return false, ""
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info struct {
		Version string `json:"Version"`
		Time    string `json:"Time"`
		Retract string `json:"Retract,omitempty"` // Retraction message if retracted
	}
	if json.Unmarshal(body, &info) != nil {
		return false, ""
	}

	if info.Retract != "" {
		return true, fmt.Sprintf("Version retracted: %s", info.Retract)
	}

	// Also check for deprecated modules via pkg.go.dev API
	// Some modules are marked deprecated at the module level
	pkgURL := fmt.Sprintf("https://pkg.go.dev/%s?tab=versions", modulePath)
	req, err = http.NewRequestWithContext(ctx, "GET", pkgURL, nil)
	if err != nil {
		return false, ""
	}
	req.Header.Set("Accept", "text/html")

	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return false, ""
	}
	defer resp.Body.Close()

	// Quick check for deprecation notice in HTML
	htmlBody, _ := io.ReadAll(resp.Body)
	htmlStr := string(htmlBody)
	if strings.Contains(htmlStr, "Deprecated") && strings.Contains(htmlStr, "deprecated") {
		return true, "Module marked as deprecated on pkg.go.dev"
	}

	return false, ""
}

// escapeModulePath escapes a module path for use in Go proxy URLs.
// Uppercase letters are encoded as !lowercase per the Go module proxy protocol.
func escapeModulePath(path string) string {
	var escaped strings.Builder
	for _, r := range path {
		if r >= 'A' && r <= 'Z' {
			escaped.WriteByte('!')
			escaped.WriteRune(r + ('a' - 'A'))
		} else {
			escaped.WriteRune(r)
		}
	}
	return escaped.String()
}

// ==================== Duplicates Feature ====================

type duplicatesFeatureResult struct {
	Summary  *DuplicatesSummary
	Findings []DuplicateFinding
}

func (s *SupplyChainScanner) runDuplicatesFeature(components []ComponentData) *duplicatesFeatureResult {
	result := &duplicatesFeatureResult{
		Summary:  &DuplicatesSummary{},
		Findings: []DuplicateFinding{},
	}

	result.Summary.TotalPackages = len(components)

	// Check for multiple versions of same package
	if s.config.Duplicates.CheckVersions {
		packageVersions := make(map[string][]string)
		for _, pkg := range components {
			key := pkg.Name
			packageVersions[key] = append(packageVersions[key], pkg.Version)
		}

		for name, versions := range packageVersions {
			if len(versions) > 1 {
				result.Findings = append(result.Findings, DuplicateFinding{
					Package:   name,
					Versions:  versions,
					IssueType: "version",
					Message:   fmt.Sprintf("Package has %d different versions installed", len(versions)),
				})
				result.Summary.DuplicateVersions++
			}
		}
	}

	// Check for packages with same functionality
	if s.config.Duplicates.CheckFunctionality {
		// Known groups of packages with similar functionality
		functionalGroups := map[string][]string{
			"date":    {"moment", "dayjs", "date-fns", "luxon"},
			"http":    {"axios", "node-fetch", "got", "request", "superagent"},
			"lodash":  {"lodash", "underscore", "ramda"},
			"promise": {"bluebird", "q", "when"},
		}

		foundPackages := make(map[string][]string)
		for _, pkg := range components {
			for group, names := range functionalGroups {
				for _, name := range names {
					if pkg.Name == name {
						foundPackages[group] = append(foundPackages[group], pkg.Name)
					}
				}
			}
		}

		for group, found := range foundPackages {
			if len(found) > 1 {
				result.Findings = append(result.Findings, DuplicateFinding{
					Package:   group,
					Versions:  found,
					IssueType: "functionality",
					Message:   fmt.Sprintf("Multiple packages for %s functionality: %s", group, strings.Join(found, ", ")),
				})
				result.Summary.DuplicateFunctionality++
			}
		}
	}

	return result
}

// ==================== Helpers ====================

func findFiles(root, name string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				n := info.Name()
				if n == "node_modules" || n == ".git" || n == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if info.Name() == name {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func extractEcosystem(purl string) string {
	if len(purl) < 5 || purl[:4] != "pkg:" {
		return "unknown"
	}
	rest := purl[4:]
	for i, c := range rest {
		if c == '/' {
			return rest[:i]
		}
	}
	return "unknown"
}
