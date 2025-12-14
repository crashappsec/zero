// Package sbom implements the SBOM (Software Bill of Materials) super scanner
// This scanner is the source of truth for all package/component data.
// Other scanners (packages, etc.) MUST depend on SBOM output.
package sbom

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/scanners/common"
)

const (
	Name    = "sbom"
	Version = "1.0.0"
)

func init() {
	scanner.Register(&SBOMScanner{})
}

// SBOMScanner is the SBOM super scanner
type SBOMScanner struct{}

func (s *SBOMScanner) Name() string {
	return Name
}

func (s *SBOMScanner) Description() string {
	return "Software Bill of Materials generation and integrity verification - source of truth for all package data"
}

func (s *SBOMScanner) Dependencies() []string {
	return nil // SBOM has no dependencies - it's the source of truth
}

func (s *SBOMScanner) EstimateDuration(fileCount int) time.Duration {
	// SBOM generation time depends on project size
	base := 10 * time.Second
	base += time.Duration(fileCount/1000) * time.Second
	return base
}

func (s *SBOMScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	start := time.Now()

	// Get feature config
	cfg := getFeatureConfig(opts)

	// Try to load sbom.config.json for additional configuration
	if configPath := FindSBOMConfig(opts.RepoPath); configPath != "" {
		if sbomCfg, err := LoadSBOMConfig(configPath); err == nil {
			ApplySBOMConfig(&cfg, sbomCfg)
		}
	}

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	// Ensure output directory
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("creating output directory: %w", err)
		}
	}

	var sbomPath string
	var components []Component

	// 1. SBOM Generation (required)
	if cfg.Generation.Enabled {
		genSummary, genFindings, path, err := runGeneration(ctx, opts, cfg.Generation)
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
	if cfg.Integrity.Enabled && sbomPath != "" {
		intSummary, intFindings, err := runIntegrity(opts.RepoPath, sbomPath, components, cfg.Integrity)
		result.FeaturesRun = append(result.FeaturesRun, "integrity")
		if err != nil {
			result.Summary.Errors = append(result.Summary.Errors, fmt.Sprintf("integrity: %v", err))
			intSummary = &IntegritySummary{Error: err.Error()}
		}
		result.Summary.Integrity = intSummary
		result.Findings.Integrity = intFindings
	}

	// Create scan result
	scanResult := scanner.NewScanResult(Name, Version, start)
	scanResult.Repository = opts.RepoPath
	scanResult.SetSummary(result.Summary)
	scanResult.SetFindings(result)

	// Write output
	if opts.OutputDir != "" {
		resultFile := filepath.Join(opts.OutputDir, Name+".json")
		if err := scanResult.WriteJSON(resultFile); err != nil {
			return nil, fmt.Errorf("writing result: %w", err)
		}
	}

	return scanResult, nil
}

func getFeatureConfig(opts *scanner.ScanOptions) FeatureConfig {
	if opts.FeatureConfig == nil {
		return DefaultConfig()
	}

	cfg := DefaultConfig()

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

	return cfg
}

// =============================================================================
// SBOM Generation Feature
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

		ecosystem := extractEcosystem(c.Purl)
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
			BomFormat:   sbomData.BomFormat,
			SpecVersion: sbomData.SpecVersion,
			Version:     sbomData.Version,
			SerialNumber: sbomData.SerialNumber,
			Timestamp:   sbomData.Metadata.Timestamp,
			Tool:        tool,
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

func extractEcosystem(purl string) string {
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

// =============================================================================
// Additional Lockfile Parsers
// =============================================================================

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
			Ecosystem: extractEcosystem(c.Purl),
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
