// Package codeownership provides code ownership and CODEOWNERS analysis
package codeownership

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// MonorepoDetector detects and analyzes monorepo configurations
type MonorepoDetector struct {
	config MonorepoConfig
}

// NewMonorepoDetector creates a new detector
func NewMonorepoDetector(config MonorepoConfig) *MonorepoDetector {
	return &MonorepoDetector{config: config}
}

// MonorepoType represents the type of monorepo tooling
type MonorepoType string

const (
	MonorepoNone      MonorepoType = ""
	MonorepoTurborepo MonorepoType = "turborepo"
	MonorepoLerna     MonorepoType = "lerna"
	MonorepoNx        MonorepoType = "nx"
	MonorepoPnpm      MonorepoType = "pnpm"
	MonorepoNpm       MonorepoType = "npm"
	MonorepoYarn      MonorepoType = "yarn"
	MonorepoCargo     MonorepoType = "cargo"
	MonorepoGo        MonorepoType = "go"
)

// Detect analyzes a repository for monorepo patterns
func (d *MonorepoDetector) Detect(repoPath string) (*MonorepoAnalysis, error) {
	analysis := &MonorepoAnalysis{
		IsMonorepo: false,
	}

	if !d.config.AutoDetect {
		return analysis, nil
	}

	// Try each detector in order of specificity
	detectors := []struct {
		detector func(string) (*MonorepoAnalysis, error)
		name     MonorepoType
	}{
		{d.detectTurborepo, MonorepoTurborepo},
		{d.detectNx, MonorepoNx},
		{d.detectLerna, MonorepoLerna},
		{d.detectPnpm, MonorepoPnpm},
		{d.detectCargo, MonorepoCargo},
		{d.detectGo, MonorepoGo},
		{d.detectNpmYarn, MonorepoNpm}, // Falls back to npm/yarn workspaces
	}

	for _, dt := range detectors {
		result, err := dt.detector(repoPath)
		if err != nil {
			continue // Try next detector
		}
		if result != nil && result.IsMonorepo {
			return result, nil
		}
	}

	return analysis, nil
}

// detectTurborepo checks for Turborepo
func (d *MonorepoDetector) detectTurborepo(repoPath string) (*MonorepoAnalysis, error) {
	turboPath := filepath.Join(repoPath, "turbo.json")
	if _, err := os.Stat(turboPath); err != nil {
		return nil, err
	}

	// Parse turbo.json
	data, err := os.ReadFile(turboPath)
	if err != nil {
		return nil, err
	}

	var turboConfig struct {
		Pipeline map[string]any `json:"pipeline"`
	}
	if err := json.Unmarshal(data, &turboConfig); err != nil {
		return nil, err
	}

	// Get workspaces from package.json
	workspaces, err := d.getPackageJsonWorkspaces(repoPath)
	if err != nil {
		workspaces = []string{}
	}

	return &MonorepoAnalysis{
		IsMonorepo: true,
		Type:       string(MonorepoTurborepo),
		ConfigFile: turboPath,
		Workspaces: d.expandWorkspaces(repoPath, workspaces),
	}, nil
}

// detectNx checks for Nx
func (d *MonorepoDetector) detectNx(repoPath string) (*MonorepoAnalysis, error) {
	nxPath := filepath.Join(repoPath, "nx.json")
	if _, err := os.Stat(nxPath); err != nil {
		return nil, err
	}

	// Get workspaces from nx.json or project.json files
	workspaces := d.findNxProjects(repoPath)

	return &MonorepoAnalysis{
		IsMonorepo: true,
		Type:       string(MonorepoNx),
		ConfigFile: nxPath,
		Workspaces: d.expandWorkspaces(repoPath, workspaces),
	}, nil
}

// detectLerna checks for Lerna
func (d *MonorepoDetector) detectLerna(repoPath string) (*MonorepoAnalysis, error) {
	lernaPath := filepath.Join(repoPath, "lerna.json")
	if _, err := os.Stat(lernaPath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(lernaPath)
	if err != nil {
		return nil, err
	}

	var lernaConfig struct {
		Packages []string `json:"packages"`
	}
	if err := json.Unmarshal(data, &lernaConfig); err != nil {
		return nil, err
	}

	workspaces := lernaConfig.Packages
	if len(workspaces) == 0 {
		workspaces = []string{"packages/*"}
	}

	return &MonorepoAnalysis{
		IsMonorepo: true,
		Type:       string(MonorepoLerna),
		ConfigFile: lernaPath,
		Workspaces: d.expandWorkspaces(repoPath, workspaces),
	}, nil
}

// detectPnpm checks for pnpm workspaces
func (d *MonorepoDetector) detectPnpm(repoPath string) (*MonorepoAnalysis, error) {
	pnpmPath := filepath.Join(repoPath, "pnpm-workspace.yaml")
	if _, err := os.Stat(pnpmPath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(pnpmPath)
	if err != nil {
		return nil, err
	}

	var pnpmConfig struct {
		Packages []string `yaml:"packages"`
	}
	if err := yaml.Unmarshal(data, &pnpmConfig); err != nil {
		return nil, err
	}

	return &MonorepoAnalysis{
		IsMonorepo: true,
		Type:       string(MonorepoPnpm),
		ConfigFile: pnpmPath,
		Workspaces: d.expandWorkspaces(repoPath, pnpmConfig.Packages),
	}, nil
}

// detectCargo checks for Cargo workspaces
func (d *MonorepoDetector) detectCargo(repoPath string) (*MonorepoAnalysis, error) {
	cargoPath := filepath.Join(repoPath, "Cargo.toml")
	if _, err := os.Stat(cargoPath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cargoPath)
	if err != nil {
		return nil, err
	}

	// Simple TOML parsing for workspace members
	content := string(data)
	if !strings.Contains(content, "[workspace]") {
		return nil, nil
	}

	// Extract workspace members
	workspaces := d.parseCargoWorkspaceMembers(content)
	if len(workspaces) == 0 {
		return nil, nil
	}

	return &MonorepoAnalysis{
		IsMonorepo: true,
		Type:       string(MonorepoCargo),
		ConfigFile: cargoPath,
		Workspaces: d.expandWorkspaces(repoPath, workspaces),
	}, nil
}

// detectGo checks for Go workspaces
func (d *MonorepoDetector) detectGo(repoPath string) (*MonorepoAnalysis, error) {
	goWorkPath := filepath.Join(repoPath, "go.work")
	if _, err := os.Stat(goWorkPath); err != nil {
		// Check for multiple go.mod files
		return d.detectMultipleGoMods(repoPath)
	}

	data, err := os.ReadFile(goWorkPath)
	if err != nil {
		return nil, err
	}

	// Parse go.work for use directives
	workspaces := d.parseGoWorkUse(string(data))

	return &MonorepoAnalysis{
		IsMonorepo: true,
		Type:       string(MonorepoGo),
		ConfigFile: goWorkPath,
		Workspaces: d.expandWorkspaces(repoPath, workspaces),
	}, nil
}

// detectMultipleGoMods checks for repos with multiple go.mod files
func (d *MonorepoDetector) detectMultipleGoMods(repoPath string) (*MonorepoAnalysis, error) {
	var goMods []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "go.mod" && path != filepath.Join(repoPath, "go.mod") {
			rel, _ := filepath.Rel(repoPath, filepath.Dir(path))
			goMods = append(goMods, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Need at least 2 go.mod files (including root) to be a monorepo
	if len(goMods) < 1 {
		return nil, nil
	}

	workspaces := make([]WorkspaceOwnership, 0, len(goMods))
	for _, mod := range goMods {
		workspaces = append(workspaces, WorkspaceOwnership{
			Name: filepath.Base(mod),
			Path: mod,
		})
	}

	return &MonorepoAnalysis{
		IsMonorepo: true,
		Type:       string(MonorepoGo),
		ConfigFile: "",
		Workspaces: workspaces,
	}, nil
}

// detectNpmYarn checks for npm/yarn workspaces in package.json
func (d *MonorepoDetector) detectNpmYarn(repoPath string) (*MonorepoAnalysis, error) {
	workspaces, err := d.getPackageJsonWorkspaces(repoPath)
	if err != nil || len(workspaces) == 0 {
		return nil, err
	}

	return &MonorepoAnalysis{
		IsMonorepo: true,
		Type:       string(MonorepoNpm),
		ConfigFile: filepath.Join(repoPath, "package.json"),
		Workspaces: d.expandWorkspaces(repoPath, workspaces),
	}, nil
}

// getPackageJsonWorkspaces extracts workspaces from package.json
func (d *MonorepoDetector) getPackageJsonWorkspaces(repoPath string) ([]string, error) {
	pkgPath := filepath.Join(repoPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Workspaces any `json:"workspaces"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	if pkg.Workspaces == nil {
		return nil, nil
	}

	// Workspaces can be string[] or {packages: string[]}
	switch ws := pkg.Workspaces.(type) {
	case []any:
		result := make([]string, 0, len(ws))
		for _, w := range ws {
			if s, ok := w.(string); ok {
				result = append(result, s)
			}
		}
		return result, nil
	case map[string]any:
		if packages, ok := ws["packages"].([]any); ok {
			result := make([]string, 0, len(packages))
			for _, p := range packages {
				if s, ok := p.(string); ok {
					result = append(result, s)
				}
			}
			return result, nil
		}
	}

	return nil, nil
}

// findNxProjects finds projects in an Nx monorepo
func (d *MonorepoDetector) findNxProjects(repoPath string) []string {
	var projects []string

	// Look for project.json files
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "project.json" {
			rel, _ := filepath.Rel(repoPath, filepath.Dir(path))
			if rel != "." {
				projects = append(projects, rel)
			}
		}
		return nil
	})
	if err != nil {
		return projects
	}

	return projects
}

// parseCargoWorkspaceMembers extracts members from Cargo.toml workspace section
func (d *MonorepoDetector) parseCargoWorkspaceMembers(content string) []string {
	var members []string

	// Simple parsing - look for members = ["..."]
	lines := strings.Split(content, "\n")
	inWorkspace := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "[workspace]" {
			inWorkspace = true
			continue
		}
		if strings.HasPrefix(line, "[") && line != "[workspace]" {
			inWorkspace = false
		}
		if inWorkspace && strings.HasPrefix(line, "members") {
			// Extract values from members = ["a", "b"]
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			if start != -1 && end != -1 {
				membersStr := line[start+1 : end]
				for _, m := range strings.Split(membersStr, ",") {
					m = strings.TrimSpace(m)
					m = strings.Trim(m, `"'`)
					if m != "" {
						members = append(members, m)
					}
				}
			}
		}
	}

	return members
}

// parseGoWorkUse extracts use directives from go.work
func (d *MonorepoDetector) parseGoWorkUse(content string) []string {
	var uses []string

	lines := strings.Split(content, "\n")
	inUse := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Single line: use ./path
		if strings.HasPrefix(line, "use ") && !strings.HasSuffix(line, "(") {
			path := strings.TrimPrefix(line, "use ")
			path = strings.TrimSpace(path)
			path = strings.Trim(path, "./")
			uses = append(uses, path)
			continue
		}

		// Multi-line block
		if strings.HasPrefix(line, "use (") || line == "use (" {
			inUse = true
			continue
		}
		if line == ")" {
			inUse = false
			continue
		}
		if inUse {
			path := strings.TrimSpace(line)
			path = strings.Trim(path, "./")
			if path != "" {
				uses = append(uses, path)
			}
		}
	}

	return uses
}

// expandWorkspaces converts workspace glob patterns to actual directories
func (d *MonorepoDetector) expandWorkspaces(repoPath string, patterns []string) []WorkspaceOwnership {
	var workspaces []WorkspaceOwnership
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Handle glob patterns
		if strings.Contains(pattern, "*") {
			matches, err := filepath.Glob(filepath.Join(repoPath, pattern))
			if err != nil {
				continue
			}
			for _, match := range matches {
				rel, _ := filepath.Rel(repoPath, match)
				if !seen[rel] {
					seen[rel] = true
					workspaces = append(workspaces, WorkspaceOwnership{
						Name: filepath.Base(match),
						Path: rel,
					})
				}
			}
		} else {
			// Direct path
			fullPath := filepath.Join(repoPath, pattern)
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				if !seen[pattern] {
					seen[pattern] = true
					workspaces = append(workspaces, WorkspaceOwnership{
						Name: filepath.Base(pattern),
						Path: pattern,
					})
				}
			}
		}
	}

	return workspaces
}
