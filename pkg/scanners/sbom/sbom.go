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

	return summary, findings, nil
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

	// Check package-lock.json
	pkgLockPath := filepath.Join(repoPath, "package-lock.json")
	if _, err := os.Stat(pkgLockPath); err == nil {
		comp := compareLockfile(pkgLockPath, "npm", compMap["npm"])
		comparisons = append(comparisons, comp)
	}

	// Check yarn.lock
	yarnLockPath := filepath.Join(repoPath, "yarn.lock")
	if _, err := os.Stat(yarnLockPath); err == nil {
		comp := compareYarnLock(yarnLockPath, compMap["npm"])
		comparisons = append(comparisons, comp)
	}

	// Check go.sum
	goSumPath := filepath.Join(repoPath, "go.sum")
	if _, err := os.Stat(goSumPath); err == nil {
		comp := compareGoSum(goSumPath, compMap["golang"])
		comparisons = append(comparisons, comp)
	}

	// Check requirements.txt
	reqPath := filepath.Join(repoPath, "requirements.txt")
	if _, err := os.Stat(reqPath); err == nil {
		comp := compareRequirements(reqPath, compMap["pypi"])
		comparisons = append(comparisons, comp)
	}

	return comparisons
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
