package developerexperience

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/crashappsec/zero/pkg/scanner"
)

func TestDevXScanner_Run(t *testing.T) {
	// Create a temp directory with test files
	tmpDir, err := os.MkdirTemp("", "devex-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	setupTestRepo(t, tmpDir)

	// Create output directory
	outputDir := filepath.Join(tmpDir, "output")
	os.MkdirAll(outputDir, 0755)

	s := &DevXScanner{}
	opts := &scanner.ScanOptions{
		RepoPath:  tmpDir,
		OutputDir: outputDir,
	}

	result, err := s.Run(context.Background(), opts)
	if err != nil {
		t.Fatalf("Scanner failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Verify output file was created
	outputFile := filepath.Join(outputDir, "developer-experience.json")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Expected developer-experience.json output file")
	}

	// Check summary by unmarshaling from JSON
	var summary Summary
	if err := json.Unmarshal(result.Summary, &summary); err != nil {
		t.Fatalf("Failed to unmarshal summary: %v", err)
	}

	// Verify onboarding feature ran
	if summary.Onboarding == nil {
		t.Error("Expected onboarding summary")
	} else {
		t.Logf("Onboarding score: %d", summary.Onboarding.Score)
		t.Logf("Setup complexity: %s", summary.Onboarding.SetupComplexity)
		t.Logf("Config file count: %d", summary.Onboarding.ConfigFileCount)
		t.Logf("Dependency count: %d", summary.Onboarding.DependencyCount)
		t.Logf("Has CONTRIBUTING: %v", summary.Onboarding.HasContributing)
	}

	// Verify sprawl feature ran
	if summary.Sprawl == nil {
		t.Error("Expected sprawl summary")
	} else {
		t.Logf("Combined score: %d", summary.Sprawl.CombinedScore)
		t.Logf("Tool sprawl index: %d, level: %s", summary.Sprawl.ToolSprawl.Index, summary.Sprawl.ToolSprawl.Level)
		t.Logf("Tech sprawl index: %d, level: %s", summary.Sprawl.TechnologySprawl.Index, summary.Sprawl.TechnologySprawl.Level)
		t.Logf("Learning curve: %s (score: %d)", summary.Sprawl.LearningCurve, summary.Sprawl.LearningCurveScore)
	}

	// Verify workflow feature ran
	if summary.Workflow == nil {
		t.Error("Expected workflow summary")
	} else {
		t.Logf("Workflow score: %d", summary.Workflow.Score)
		t.Logf("Efficiency level: %s", summary.Workflow.EfficiencyLevel)
		t.Logf("Has PR templates: %v", summary.Workflow.HasPRTemplates)
		t.Logf("Has hot reload: %v", summary.Workflow.HasHotReload)
	}
}

func TestDevXScanner_OnboardingFeature(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devex-onboarding-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a well-documented repo
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(`
# Test Project

## Installation

Run npm install to get started.

## Prerequisites

- Node.js 18+
- npm

## Quick Start

1. Clone the repo
2. Run npm install
3. Run npm start

## Usage

Import and use the library.

## Examples

`+"```javascript"+`
const lib = require('test');
lib.doSomething();
`+"```"+`
`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "CONTRIBUTING.md"), []byte(`# Contributing\n\nPR welcome!`), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("DATABASE_URL=\nAPI_KEY=\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{
  "name": "test",
  "dependencies": {"express": "^4.18.0"},
  "devDependencies": {"jest": "^29.0.0"}
}`), 0644)

	s := &DevXScanner{}
	cfg := DefaultConfig()
	summary, findings := s.runOnboarding(context.Background(), &scanner.ScanOptions{RepoPath: tmpDir}, cfg.Onboarding)

	if summary.Score < 70 {
		t.Errorf("Expected high onboarding score for well-documented repo, got %d", summary.Score)
	}

	if !summary.HasContributing {
		t.Error("Expected HasContributing to be true")
	}

	if !summary.HasEnvExample {
		t.Error("Expected HasEnvExample to be true")
	}

	if summary.DependencyCount != 2 {
		t.Errorf("Expected 2 dependencies, got %d", summary.DependencyCount)
	}

	if summary.EnvVarCount != 2 {
		t.Errorf("Expected 2 env vars, got %d", summary.EnvVarCount)
	}

	if summary.ReadmeQualityScore < 60 {
		t.Errorf("Expected good README quality score, got %d", summary.ReadmeQualityScore)
	}

	t.Logf("Onboarding score: %d", summary.Score)
	t.Logf("README quality: %d", summary.ReadmeQualityScore)
	t.Logf("Setup barriers: %d", len(findings.SetupBarriers))
}

func TestDevXScanner_WorkflowFeature(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devex-workflow-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .github directory with templates
	githubDir := filepath.Join(tmpDir, ".github")
	os.MkdirAll(githubDir, 0755)

	os.WriteFile(filepath.Join(githubDir, "PULL_REQUEST_TEMPLATE.md"), []byte(`
## Description
<!-- Describe your changes -->

## Testing
- [ ] Unit tests added
- [ ] Manual testing done

## Checklist
- [ ] Code follows style guide
- [ ] Documentation updated
`), 0644)

	issueDir := filepath.Join(githubDir, "ISSUE_TEMPLATE")
	os.MkdirAll(issueDir, 0755)
	os.WriteFile(filepath.Join(issueDir, "bug_report.md"), []byte("Bug report template"), 0644)
	os.WriteFile(filepath.Join(issueDir, "feature_request.md"), []byte("Feature request template"), 0644)

	// Create devcontainer
	devcontainerDir := filepath.Join(tmpDir, ".devcontainer")
	os.MkdirAll(devcontainerDir, 0755)
	os.WriteFile(filepath.Join(devcontainerDir, "devcontainer.json"), []byte(`{"name": "test"}`), 0644)

	// Create docker-compose
	os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte("version: '3'\nservices:\n  app:\n    build: ."), 0644)

	// Create package.json with hot reload
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{
  "dependencies": {"next": "^14.0.0"},
  "devDependencies": {"nodemon": "^3.0.0"}
}`), 0644)

	s := &DevXScanner{}
	cfg := DefaultConfig()
	summary, findings := s.runWorkflow(context.Background(), &scanner.ScanOptions{RepoPath: tmpDir}, cfg.Workflow)

	if !summary.HasPRTemplates {
		t.Error("Expected HasPRTemplates to be true")
	}

	if !summary.HasIssueTemplates {
		t.Error("Expected HasIssueTemplates to be true")
	}

	if !summary.HasDevContainer {
		t.Error("Expected HasDevContainer to be true")
	}

	if !summary.HasDockerCompose {
		t.Error("Expected HasDockerCompose to be true")
	}

	if !summary.HasHotReload {
		t.Error("Expected HasHotReload to be true (next.js)")
	}

	if !summary.HasWatchMode {
		t.Error("Expected HasWatchMode to be true (nodemon)")
	}

	if len(findings.PRTemplates) == 0 {
		t.Error("Expected PR templates in findings")
	}

	if len(findings.IssueTemplates) != 2 {
		t.Errorf("Expected 2 issue templates, got %d", len(findings.IssueTemplates))
	}

	t.Logf("Workflow score: %d", summary.Score)
	t.Logf("PR process score: %d", summary.PRProcessScore)
	t.Logf("Local dev score: %d", summary.LocalDevScore)
	t.Logf("Feedback loop score: %d", summary.FeedbackLoopScore)
}

func setupTestRepo(t *testing.T, dir string) {
	// Create README
	os.WriteFile(filepath.Join(dir, "README.md"), []byte(`
# Test Project

## Installation

npm install

## Usage

Import the library.
`), 0644)

	// Create package.json
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
  "name": "test-project",
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "^4.17.0"
  },
  "devDependencies": {
    "jest": "^29.0.0",
    "eslint": "^8.0.0"
  },
  "scripts": {
    "dev": "node server.js",
    "build": "tsc"
  }
}`), 0644)

	// Create tsconfig.json
	os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{
  "compilerOptions": {
    "target": "es2020",
    "module": "commonjs"
  }
}`), 0644)

	// Create .eslintrc.json
	os.WriteFile(filepath.Join(dir, ".eslintrc.json"), []byte(`{
  "extends": ["eslint:recommended"],
  "env": {"node": true}
}`), 0644)

	// Create Makefile
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte(`
.PHONY: build test dev

build:
	npm run build

test:
	npm test

dev:
	npm run dev
`), 0644)

	// Create .github workflows
	githubDir := filepath.Join(dir, ".github", "workflows")
	os.MkdirAll(githubDir, 0755)
	os.WriteFile(filepath.Join(githubDir, "ci.yml"), []byte(`
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm test
`), 0644)
}
