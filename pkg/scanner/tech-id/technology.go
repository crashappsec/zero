// Package techid provides the consolidated technology identification super scanner
// Includes AI/ML security and ML-BOM generation
package techid

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/core/cyclonedx"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/scanner/common"
)

// ruleLoadMessageOnce ensures we only print the rule loading message once
var ruleLoadMessageOnce sync.Once
var ruleLoadMessagePrinted bool

const (
	Name        = "tech-id"
	Description = "Technology identification, AI/ML security analysis and ML-BOM generation"
)

// TechnologyScanner implements the technology identification super scanner
type TechnologyScanner struct {
	config FeatureConfig
}

// init registers the scanner
func init() {
	scanner.Register(&TechnologyScanner{
		config: DefaultConfig(),
	})
}

// Name returns the scanner name
func (s *TechnologyScanner) Name() string {
	return Name
}

// Description returns the scanner description
func (s *TechnologyScanner) Description() string {
	return Description
}

// Dependencies returns scanner dependencies (none for technology scanner)
func (s *TechnologyScanner) Dependencies() []string {
	return []string{}
}

// EstimateDuration returns estimated scan duration based on file count
func (s *TechnologyScanner) EstimateDuration(fileCount int) time.Duration {
	// Base time + time per file for pattern scanning
	base := 5 * time.Second
	perFile := 10 * time.Millisecond
	return base + time.Duration(fileCount)*perFile
}

// Run executes the AI/ML analysis
func (s *TechnologyScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	startTime := time.Now()

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	repoPath := opts.RepoPath

	// Status callback - use OnStatus from options if provided, otherwise verbose fallback
	onStatus := func(msg string) {
		if opts.OnStatus != nil {
			opts.OnStatus(msg)
		} else if opts.Verbose {
			fmt.Printf("[tech-id] %s\n", msg)
		}
	}

	// Step 1: Check semgrep is installed (required)
	onStatus("Checking semgrep installation...")
	if !HasSemgrep() {
		return nil, fmt.Errorf("semgrep is required but not installed. Install with: pip install semgrep")
	}

	// Step 2: Refresh semgrep rules from RAG patterns
	var ruleManager *RuleManager
	onStatus("Loading RAG technology patterns...")
	ruleManager = NewRuleManager(RuleManagerConfig{
		TTL:      s.config.Semgrep.CacheTTL,
		OnStatus: onStatus,
	})
	refreshResult := ruleManager.RefreshRules(ctx, s.config.Semgrep.ForceRefresh)
	if refreshResult.Error != nil {
		return nil, fmt.Errorf("failed to refresh semgrep rules: %w", refreshResult.Error)
	}
	result.FeaturesRun = append(result.FeaturesRun, "semgrep_rules")
	result.Summary.SemgrepRulesLoaded = refreshResult.TotalRules

	// Log rule loading results only once (for first repo in batch)
	ruleLoadMessageOnce.Do(func() {
		if refreshResult.Refreshed {
			fmt.Printf("          ▸ Converted RAG patterns → %d semgrep rules (%d tech, %d secrets, %d AI/ML)\n",
				refreshResult.TotalRules, refreshResult.TechRules, refreshResult.SecretRules, refreshResult.AIMLRules)
		} else {
			fmt.Printf("          ▸ Using cached semgrep rules (%d patterns)\n", refreshResult.TotalRules)
		}
		ruleLoadMessagePrinted = true
	})

	// Step 3: Run semgrep with generated rules
	rulePaths := ruleManager.GetRulePaths()
	if len(rulePaths) == 0 {
		return nil, fmt.Errorf("no semgrep rules were generated from RAG patterns")
	}
	onStatus(fmt.Sprintf("Running semgrep with %d rule files...", len(rulePaths)))
	semgrepResult := RunSemgrepWithRules(ctx, repoPath, ruleManager, onStatus)
	if semgrepResult.Error != nil {
		return nil, fmt.Errorf("semgrep scan failed: %w", semgrepResult.Error)
	}

	// Report semgrep results - count unique technologies from findings
	uniqueTechs := make(map[string]bool)
	for _, f := range semgrepResult.Findings {
		if f.Technology != "" {
			uniqueTechs[f.Technology] = true
		}
	}
	techCount := len(uniqueTechs)
	secretCount := len(semgrepResult.Secrets)
	if techCount > 0 || secretCount > 0 {
		onStatus(fmt.Sprintf("Semgrep found %d technologies, %d secrets in %.1fs",
			techCount, secretCount, semgrepResult.Duration.Seconds()))
	} else {
		onStatus(fmt.Sprintf("Semgrep scan completed in %.1fs (no findings)", semgrepResult.Duration.Seconds()))
	}

	// Merge semgrep findings into result
	s.mergeSemgrepFindings(semgrepResult, result)
	result.FeaturesRun = append(result.FeaturesRun, "semgrep_scan")

	// Run each enabled feature (Go-native detection)
	if s.config.Technology.Enabled {
		result.FeaturesRun = append(result.FeaturesRun, "technology")
		s.runTechnologyFeature(ctx, repoPath, opts.SBOMPath, result)
	}

	if s.config.Models.Enabled {
		result.FeaturesRun = append(result.FeaturesRun, "models")
		s.runModelsFeature(ctx, repoPath, result)
	}

	if s.config.Frameworks.Enabled {
		result.FeaturesRun = append(result.FeaturesRun, "frameworks")
		s.runFrameworksFeature(ctx, repoPath, result)
	}

	if s.config.Datasets.Enabled {
		result.FeaturesRun = append(result.FeaturesRun, "datasets")
		s.runDatasetsFeature(ctx, repoPath, result)
	}

	if s.config.Security.Enabled {
		result.FeaturesRun = append(result.FeaturesRun, "security")
		s.runSecurityFeature(ctx, repoPath, result)
	}

	if s.config.Governance.Enabled {
		result.FeaturesRun = append(result.FeaturesRun, "governance")
		s.runGovernanceFeature(ctx, repoPath, result)
	}

	if s.config.Infrastructure.Enabled {
		result.FeaturesRun = append(result.FeaturesRun, "infrastructure")
		s.runInfrastructureFeature(ctx, repoPath, onStatus, result)
	}

	// Create scan result using the proper interface
	scanResult := scanner.NewScanResult(Name, "1.0.0", startTime)

	if err := scanResult.SetSummary(result.Summary); err != nil {
		return nil, fmt.Errorf("failed to set summary: %w", err)
	}

	if err := scanResult.SetFindings(result.Findings); err != nil {
		return nil, fmt.Errorf("failed to set findings: %w", err)
	}

	// Add metadata with features run
	metadata := map[string]interface{}{
		"features_run": result.FeaturesRun,
	}
	if err := scanResult.SetMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to set metadata: %w", err)
	}

	// Write output to disk
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("creating output dir: %w", err)
		}
		resultFile := filepath.Join(opts.OutputDir, Name+".json")
		if err := scanResult.WriteJSON(resultFile); err != nil {
			return nil, fmt.Errorf("writing result: %w", err)
		}

		// Export ML-BOM (CycloneDX format)
		if err := s.exportMLBOM(opts.OutputDir, result); err != nil {
			// Log warning but don't fail the scan
			onStatus(fmt.Sprintf("Warning: failed to export ML-BOM: %v", err))
		}
	}

	return scanResult, nil
}

// exportMLBOM exports findings as a CycloneDX ML-BOM
func (s *TechnologyScanner) exportMLBOM(outputDir string, result *Result) error {
	bom := cyclonedx.NewMLBOM()

	// Add ML models as components
	for _, model := range result.Findings.Models {
		c := cyclonedx.MLModelToComponent(
			model.Name,
			model.Version,
			model.Source,
			model.SourceURL,
			model.Format,
			model.Architecture,
			model.Task,
			model.License,
		)

		// Enrich with model card if available
		if model.ModelCard != nil {
			mc := cyclonedx.NewModelCard()
			if model.Task != "" {
				mc.WithTask(model.Task)
			}
			if model.Architecture != "" {
				mc.WithArchitecture(inferArchitectureFamily(model.Architecture), model.Architecture)
			}
			for _, ds := range model.Datasets {
				mc.WithDataset(fmt.Sprintf("dataset/%s", ds), "training")
			}
			if model.ModelCard.Limitations != "" {
				mc.WithLimitation(model.ModelCard.Limitations)
			}
			if model.ModelCard.IntendedUse != "" {
				mc.WithUseCase(model.ModelCard.IntendedUse)
			}
			c.ModelCard = mc
		}

		// Add security risk as property
		if model.SecurityRisk != "" {
			c.AddProperty("zero:security_risk", model.SecurityRisk)
		}

		// Add file path evidence
		if model.FilePath != "" {
			c.Evidence = &cyclonedx.Evidence{
				Occurrences: []cyclonedx.Occurrence{{Location: model.FilePath}},
			}
		}

		bom.WithComponent(c)
	}

	// Add frameworks as components
	for _, fw := range result.Findings.Frameworks {
		c := cyclonedx.FrameworkToComponent(fw.Name, fw.Version, fw.Category, fw.Package)
		bom.WithComponent(c)
	}

	// Add datasets as components
	for _, ds := range result.Findings.Datasets {
		c := cyclonedx.DatasetToComponent(ds.Name, ds.Source, ds.SourceURL, ds.License, ds.Description)
		bom.WithComponent(c)
	}

	// Add security findings as vulnerabilities
	for _, finding := range result.Findings.Security {
		v := cyclonedx.Vulnerability{
			ID: finding.ID,
			Source: &cyclonedx.VulnSource{
				Name: "Zero AI Security Scanner",
			},
			Description:    fmt.Sprintf("%s: %s", finding.Title, finding.Description),
			Recommendation: finding.Remediation,
			Ratings: []cyclonedx.VulnRating{
				{
					Severity: cyclonedx.SeverityToCycloneDX(finding.Severity),
					Method:   "other",
				},
			},
		}

		if finding.ModelName != "" {
			v.Affects = []cyclonedx.VulnAffect{
				{Ref: fmt.Sprintf("model/%s", finding.ModelName)},
			}
		}

		bom.WithVulnerability(v)
	}

	// Add governance findings as vulnerabilities
	for _, finding := range result.Findings.Governance {
		v := cyclonedx.Vulnerability{
			ID: finding.ID,
			Source: &cyclonedx.VulnSource{
				Name: "Zero AI Governance Scanner",
			},
			Description:    fmt.Sprintf("%s: %s", finding.Title, finding.Description),
			Recommendation: finding.Remediation,
			Ratings: []cyclonedx.VulnRating{
				{
					Severity: cyclonedx.SeverityToCycloneDX(finding.Severity),
					Method:   "other",
				},
			},
		}

		if finding.ModelName != "" {
			v.Affects = []cyclonedx.VulnAffect{
				{Ref: fmt.Sprintf("model/%s", finding.ModelName)},
			}
		}

		bom.WithVulnerability(v)
	}

	// Write ML-BOM
	exporter := cyclonedx.NewExporter(outputDir)
	return exporter.WriteMLBOM(bom, "mlbom.cdx.json")
}

// inferArchitectureFamily infers the ML architecture family from model architecture
func inferArchitectureFamily(architecture string) string {
	archLower := strings.ToLower(architecture)
	families := map[string][]string{
		"transformer": {"bert", "gpt", "llama", "mistral", "t5", "roberta", "transformer"},
		"cnn":         {"resnet", "vgg", "inception", "efficientnet", "cnn"},
		"rnn":         {"lstm", "gru", "rnn"},
		"gan":         {"gan", "stylegan", "dcgan"},
		"diffusion":   {"stable-diffusion", "dalle", "diffusion"},
	}
	for family, patterns := range families {
		for _, pattern := range patterns {
			if strings.Contains(archLower, pattern) {
				return family
			}
		}
	}
	return "other"
}

// mergeSemgrepFindings merges semgrep results into the main result
func (s *TechnologyScanner) mergeSemgrepFindings(semgrepResult *SemgrepResult, result *Result) {
	if semgrepResult == nil {
		return
	}

	// Count total findings
	result.Summary.SemgrepFindings = len(semgrepResult.Findings) + len(semgrepResult.Secrets)

	// Merge technology findings with file locations for inline display
	for _, f := range semgrepResult.Findings {
		tech := Technology{
			Name:       f.Technology,
			Category:   f.Category,
			Confidence: f.Confidence,
			Source:     "semgrep",
			File:       f.File,
			Line:       f.Line,
			Match:      f.Match,
		}
		result.Findings.Technology = append(result.Findings.Technology, tech)
	}

	// Merge secret findings into security findings
	for _, sf := range semgrepResult.Secrets {
		finding := SecurityFinding{
			ID:          sf.RuleID,
			Category:    "api_key_exposure",
			Severity:    sf.Severity,
			Title:       fmt.Sprintf("Exposed %s", sf.SecretType),
			Description: sf.Message,
			File:        sf.File,
			Line:        sf.Line,
			Remediation: "Remove the secret and rotate credentials",
		}
		result.Findings.Security = append(result.Findings.Security, finding)
	}

	// Update technology summary with semgrep-detected technologies
	if result.Summary.Technology == nil {
		result.Summary.Technology = &TechnologySummary{
			ByCategory: make(map[string]int),
		}
	}
	// semgrepResult.Technologies is keyed by category -> count
	for category, count := range semgrepResult.Technologies {
		result.Summary.Technology.ByCategory[category] += count
	}

	// Count technologies and track frequency for top technologies
	techCounts := make(map[string]int)
	for _, f := range semgrepResult.Findings {
		if f.Technology != "" {
			techCounts[f.Technology]++
		}
	}
	result.Summary.Technology.TotalTechnologies += len(techCounts)

	// Get top 3 technologies by frequency
	type techFreq struct {
		name  string
		count int
	}
	var techList []techFreq
	for name, count := range techCounts {
		techList = append(techList, techFreq{name, count})
	}
	sort.Slice(techList, func(i, j int) bool {
		return techList[i].count > techList[j].count
	})

	// Take top 3
	topN := 3
	if len(techList) < topN {
		topN = len(techList)
	}
	for i := 0; i < topN; i++ {
		result.Summary.Technology.TopTechnologies = append(
			result.Summary.Technology.TopTechnologies,
			techList[i].name,
		)
	}
}

// runModelsFeature detects ML models in the repository
func (s *TechnologyScanner) runModelsFeature(ctx context.Context, repoPath string, result *Result) {
	summary := &ModelsSummary{
		BySource: make(map[string]int),
		ByFormat: make(map[string]int),
	}

	// 1. Detect local model files
	if s.config.Models.DetectModelFiles {
		modelFiles := s.detectModelFiles(repoPath)
		for _, model := range modelFiles {
			result.Findings.Models = append(result.Findings.Models, model)
			summary.TotalModels++
			summary.LocalModelFiles++
			summary.BySource["local"]++
			if model.Format != "" {
				summary.ByFormat[model.Format]++
			}
		}
	}

	// 2. Scan code for model loading patterns
	if s.config.Models.ScanCodePatterns {
		codeModels := s.scanCodeForModels(repoPath)
		for _, model := range codeModels {
			// Deduplicate with existing models
			if !s.modelExists(result.Findings.Models, model.Name) {
				result.Findings.Models = append(result.Findings.Models, model)
				summary.TotalModels++
				summary.BySource[model.Source]++
			}
		}
	}

	// 3. Scan config files for model references
	if s.config.Models.ScanConfigs {
		configModels := s.scanConfigsForModels(repoPath)
		for _, model := range configModels {
			if !s.modelExists(result.Findings.Models, model.Name) {
				result.Findings.Models = append(result.Findings.Models, model)
				summary.TotalModels++
				summary.BySource[model.Source]++
			}
		}
	}

	// 4. Query model registry APIs for metadata
	if s.config.Models.QueryHuggingFace {
		// Enrich HuggingFace models
		s.enrichWithHuggingFaceMetadata(ctx, result.Findings.Models)

		// Enrich Replicate models
		s.enrichWithReplicateMetadata(ctx, result.Findings.Models)

		// Add source URLs for all models
		s.enrichModelSourceURLs(result.Findings.Models)

		for _, model := range result.Findings.Models {
			if model.ModelCard != nil {
				summary.WithModelCard++
			}
			if model.License != "" {
				summary.WithLicense++
			}
			if len(model.Datasets) > 0 {
				summary.WithDatasetInfo++
			}
		}
	}

	// Count API models and track registry sources
	registryCounts := make(map[string]int)
	for _, model := range result.Findings.Models {
		if model.Source == "api" {
			summary.APIModels++
		}
		registryCounts[model.Source]++
	}

	// Add registry counts to summary
	summary.BySource = registryCounts

	result.Summary.Models = summary
}

// Minimum file size for model detection (10KB) - avoids test data and small files
const minModelFileSize = 10 * 1024

// detectModelFiles scans for model files by extension
func (s *TechnologyScanner) detectModelFiles(repoPath string) []MLModel {
	var models []MLModel

	// Use RAG-loaded formats with fallback to hardcoded
	fileFormats := GetModelFileFormatsFromRAG()

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip common non-model directories
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "__pycache__" || name == ".venv" || name == "venv" || name == "test" || name == "tests" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip files smaller than minimum size (likely test data)
		if info.Size() < minModelFileSize {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		fileName := strings.ToLower(info.Name())

		// Special handling for .pb files - only detect saved_model.pb
		if ext == ".pb" {
			if !isTensorFlowSavedModel(path, fileName) {
				return nil
			}
			// Add TensorFlow SavedModel format info
			relPath, _ := filepath.Rel(repoPath, path)
			models = append(models, MLModel{
				Name:         info.Name(),
				Source:       "local",
				Format:       "tensorflow",
				FilePath:     relPath,
				SecurityRisk: "medium",
			})
			return nil
		}

		if formatInfo, ok := fileFormats[ext]; ok {
			relPath, _ := filepath.Rel(repoPath, path)

			model := MLModel{
				Name:         info.Name(),
				Source:       "local",
				Format:       formatInfo.Format,
				FilePath:     relPath,
				SecurityRisk: formatInfo.Risk,
			}

			if formatInfo.Risk == "high" {
				model.SecurityNotes = append(model.SecurityNotes, formatInfo.RiskReason)
			}

			models = append(models, model)
		}

		return nil
	})

	return models
}

// isTensorFlowSavedModel checks if a .pb file is a TensorFlow SavedModel
func isTensorFlowSavedModel(path, fileName string) bool {
	// Must be named saved_model.pb
	if fileName != "saved_model.pb" {
		return false
	}

	// Check if it's in a SavedModel directory structure
	// SavedModel directories typically contain saved_model.pb and a variables/ subdirectory
	dir := filepath.Dir(path)
	variablesDir := filepath.Join(dir, "variables")
	if info, err := os.Stat(variablesDir); err == nil && info.IsDir() {
		return true
	}

	// Also accept if parent directory looks like a model name
	parentName := filepath.Base(dir)
	modelDirPatterns := []string{"saved_model", "model", "checkpoint", "export"}
	for _, pattern := range modelDirPatterns {
		if strings.Contains(strings.ToLower(parentName), pattern) {
			return true
		}
	}

	return false
}

// Model loading patterns to detect in code
var modelLoadPatterns = []struct {
	Pattern     *regexp.Regexp
	Source      string
	ExtractName func([]string) string
}{
	// HuggingFace Transformers
	{
		Pattern:     regexp.MustCompile(`(?:AutoModel|AutoTokenizer|AutoProcessor|AutoFeatureExtractor|AutoConfig|pipeline)\s*\.?\s*from_pretrained\s*\(\s*["']([^"']+)["']`),
		Source:      "huggingface",
		ExtractName: func(m []string) string { return m[1] },
	},
	// HuggingFace with variable
	{
		Pattern:     regexp.MustCompile(`from_pretrained\s*\(\s*["']([^"']+)["']`),
		Source:      "huggingface",
		ExtractName: func(m []string) string { return m[1] },
	},
	// HuggingFace Hub download
	{
		Pattern:     regexp.MustCompile(`hf_hub_download\s*\([^)]*repo_id\s*=\s*["']([^"']+)["']`),
		Source:      "huggingface",
		ExtractName: func(m []string) string { return m[1] },
	},
	// PyTorch Hub
	{
		Pattern:     regexp.MustCompile(`torch\.hub\.load\s*\(\s*["']([^"']+)["']\s*,\s*["']([^"']+)["']`),
		Source:      "pytorch_hub",
		ExtractName: func(m []string) string { return m[1] + "/" + m[2] },
	},
	// TensorFlow Hub
	{
		Pattern:     regexp.MustCompile(`hub\.(?:KerasLayer|load)\s*\(\s*["']([^"']+)["']`),
		Source:      "tensorflow_hub",
		ExtractName: func(m []string) string { return m[1] },
	},
	// TensorFlow Hub URL pattern
	{
		Pattern:     regexp.MustCompile(`["'](https?://tfhub\.dev/[^"']+)["']`),
		Source:      "tensorflow_hub",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Replicate models - replicate.run("owner/model:version")
	{
		Pattern:     regexp.MustCompile(`replicate\.run\s*\(\s*["']([^"']+)["']`),
		Source:      "replicate",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Replicate Client.run
	{
		Pattern:     regexp.MustCompile(`Replicate\s*\([^)]*\)\s*\.run\s*\(\s*["']([^"']+)["']`),
		Source:      "replicate",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Weights & Biases model artifacts
	{
		Pattern:     regexp.MustCompile(`wandb\.use_artifact\s*\(\s*["']([^"']+)["']`),
		Source:      "wandb",
		ExtractName: func(m []string) string { return m[1] },
	},
	// W&B artifact download
	{
		Pattern:     regexp.MustCompile(`wandb\.Api\s*\([^)]*\)\.artifact\s*\(\s*["']([^"']+)["']`),
		Source:      "wandb",
		ExtractName: func(m []string) string { return m[1] },
	},
	// MLflow model loading
	{
		Pattern:     regexp.MustCompile(`mlflow\.(?:pyfunc|sklearn|pytorch|tensorflow|keras)\.load_model\s*\(\s*["']([^"']+)["']`),
		Source:      "mlflow",
		ExtractName: func(m []string) string { return m[1] },
	},
	// MLflow model registry URI
	{
		Pattern:     regexp.MustCompile(`["'](models:/[^"']+)["']`),
		Source:      "mlflow",
		ExtractName: func(m []string) string { return m[1] },
	},
	// MLflow runs artifacts
	{
		Pattern:     regexp.MustCompile(`["'](runs:/[^"']+)["']`),
		Source:      "mlflow",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Kaggle models download
	{
		Pattern:     regexp.MustCompile(`kaggle\.api\.model_get\s*\([^)]*model\s*=\s*["']([^"']+)["']`),
		Source:      "kaggle",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Kaggle Hub
	{
		Pattern:     regexp.MustCompile(`kagglehub\.model_download\s*\(\s*["']([^"']+)["']`),
		Source:      "kaggle",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Civitai model references (URL pattern)
	{
		Pattern:     regexp.MustCompile(`["'](https?://civitai\.com/(?:api/download/)?models/[^"']+)["']`),
		Source:      "civitai",
		ExtractName: func(m []string) string { return m[1] },
	},
	// NVIDIA NGC catalog
	{
		Pattern:     regexp.MustCompile(`["'](nvcr\.io/[^"']+)["']`),
		Source:      "nvidia_ngc",
		ExtractName: func(m []string) string { return m[1] },
	},
	// NGC CLI pattern
	{
		Pattern:     regexp.MustCompile(`ngc\s+(?:registry\s+)?model\s+download[^"']*["']([^"']+)["']`),
		Source:      "nvidia_ngc",
		ExtractName: func(m []string) string { return m[1] },
	},
	// AWS SageMaker JumpStart
	{
		Pattern:     regexp.MustCompile(`JumpStartModel\s*\(\s*model_id\s*=\s*["']([^"']+)["']`),
		Source:      "aws_jumpstart",
		ExtractName: func(m []string) string { return m[1] },
	},
	// SageMaker model ID
	{
		Pattern:     regexp.MustCompile(`sagemaker\.(?:model_uris|image_uris)\.[^(]+\([^)]*model_id\s*=\s*["']([^"']+)["']`),
		Source:      "aws_jumpstart",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Azure ML model
	{
		Pattern:     regexp.MustCompile(`Model\s*\([^)]*name\s*=\s*["']([^"']+)["'][^)]*workspace`),
		Source:      "azure_ml",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Azure ML model registry
	{
		Pattern:     regexp.MustCompile(`azureml://registries/[^/]+/models/([^/"']+)`),
		Source:      "azure_ml",
		ExtractName: func(m []string) string { return m[1] },
	},
	// OpenAI API models
	{
		Pattern:     regexp.MustCompile(`model\s*[=:]\s*["'](gpt-4[^"']*|gpt-3\.5[^"']*|text-embedding[^"']*|dall-e[^"']*|whisper[^"']*)["']`),
		Source:      "api",
		ExtractName: func(m []string) string { return "openai/" + m[1] },
	},
	// Anthropic API models
	{
		Pattern:     regexp.MustCompile(`model\s*[=:]\s*["'](claude-[^"']+)["']`),
		Source:      "api",
		ExtractName: func(m []string) string { return "anthropic/" + m[1] },
	},
	// Google/Gemini models
	{
		Pattern:     regexp.MustCompile(`model\s*[=:]\s*["'](gemini-[^"']+)["']`),
		Source:      "api",
		ExtractName: func(m []string) string { return "google/" + m[1] },
	},
	// Mistral models
	{
		Pattern:     regexp.MustCompile(`model\s*[=:]\s*["'](mistral-[^"']+|open-mistral[^"']+|open-mixtral[^"']+)["']`),
		Source:      "api",
		ExtractName: func(m []string) string { return "mistral/" + m[1] },
	},
	// Cohere models
	{
		Pattern:     regexp.MustCompile(`model\s*[=:]\s*["'](command[^"']*|embed[^"']*)["']`),
		Source:      "api",
		ExtractName: func(m []string) string { return "cohere/" + m[1] },
	},
	// Together AI models
	{
		Pattern:     regexp.MustCompile(`together\.Complete[^(]*\([^)]*model\s*=\s*["']([^"']+)["']`),
		Source:      "api",
		ExtractName: func(m []string) string { return "together/" + m[1] },
	},
	// Groq models
	{
		Pattern:     regexp.MustCompile(`Groq\s*\([^)]*\)\.chat[^(]*\([^)]*model\s*=\s*["']([^"']+)["']`),
		Source:      "api",
		ExtractName: func(m []string) string { return "groq/" + m[1] },
	},
	// Ollama models
	{
		Pattern:     regexp.MustCompile(`ollama\.(?:chat|generate|embeddings)\s*\([^)]*model\s*=\s*["']([^"']+)["']`),
		Source:      "ollama",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Ollama client pull/show
	{
		Pattern:     regexp.MustCompile(`(?:model|model_name)\s*[=:]\s*["']([a-z0-9]+-?[a-z0-9]*(?::[a-z0-9.]+)?)["']`),
		Source:      "ollama",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Sentence Transformers
	{
		Pattern:     regexp.MustCompile(`SentenceTransformer\s*\(\s*["']([^"']+)["']`),
		Source:      "huggingface",
		ExtractName: func(m []string) string { return m[1] },
	},
	// LangChain ChatOpenAI/ChatAnthropic model param
	{
		Pattern:     regexp.MustCompile(`Chat(?:OpenAI|Anthropic|Google|VertexAI|Cohere|Mistral)\s*\([^)]*model(?:_name)?\s*=\s*["']([^"']+)["']`),
		Source:      "api",
		ExtractName: func(m []string) string { return m[1] },
	},
	// Llama.cpp / llama-cpp-python
	{
		Pattern:     regexp.MustCompile(`Llama\s*\([^)]*model_path\s*=\s*["']([^"']+)["']`),
		Source:      "local",
		ExtractName: func(m []string) string { return filepath.Base(m[1]) },
	},
	// vLLM model loading
	{
		Pattern:     regexp.MustCompile(`LLM\s*\(\s*(?:model\s*=\s*)?["']([^"']+)["']`),
		Source:      "huggingface",
		ExtractName: func(m []string) string { return m[1] },
	},
	// text-generation-inference
	{
		Pattern:     regexp.MustCompile(`--model-id\s+["']?([^\s"']+)`),
		Source:      "huggingface",
		ExtractName: func(m []string) string { return m[1] },
	},
}

// scanCodeForModels scans Python/JS files for model loading patterns
func (s *TechnologyScanner) scanCodeForModels(repoPath string) []MLModel {
	var models []MLModel
	seen := make(map[string]bool)

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip non-code files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".py" && ext != ".js" && ext != ".ts" && ext != ".jsx" && ext != ".tsx" {
			return nil
		}

		// Skip common directories
		if strings.Contains(path, "node_modules") || strings.Contains(path, "__pycache__") ||
			strings.Contains(path, ".git") || strings.Contains(path, "venv") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(repoPath, path)
		lines := strings.Split(string(content), "\n")

		for lineNum, line := range lines {
			for _, pattern := range modelLoadPatterns {
				matches := pattern.Pattern.FindStringSubmatch(line)
				if len(matches) > 0 {
					modelName := pattern.ExtractName(matches)

					// Skip if already seen
					key := modelName + "@" + pattern.Source
					if seen[key] {
						continue
					}
					seen[key] = true

					model := MLModel{
						Name:   modelName,
						Source: pattern.Source,
						CodeLocation: &CodeLocation{
							File:    relPath,
							Line:    lineNum + 1,
							Snippet: strings.TrimSpace(line),
						},
					}

					// Mark API models
					if pattern.Source == "api" {
						model.SecurityNotes = append(model.SecurityNotes, "API-based model - no local weights")
					}

					models = append(models, model)
				}
			}
		}

		return nil
	})

	return models
}

// Config file patterns for model references
var configModelPatterns = []struct {
	Keys    []string
	Pattern *regexp.Regexp
}{
	{Keys: []string{"model", "model_name", "model_id", "base_model", "llm_model"}, Pattern: nil},
	{Keys: []string{"embedding_model", "embeddings_model"}, Pattern: nil},
	{Keys: []string{"hf_model", "huggingface_model"}, Pattern: nil},
}

// scanConfigsForModels scans YAML/JSON config files for model references
func (s *TechnologyScanner) scanConfigsForModels(repoPath string) []MLModel {
	var models []MLModel
	seen := make(map[string]bool)

	configExtensions := map[string]bool{
		".yaml": true, ".yml": true, ".json": true, ".toml": true,
	}

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !configExtensions[ext] {
			return nil
		}

		// Skip lock files and node_modules
		if strings.Contains(path, "node_modules") || strings.Contains(path, "package-lock") ||
			strings.Contains(path, "yarn.lock") || strings.Contains(path, ".git") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(repoPath, path)

		// Simple pattern matching for model-related keys
		// More sophisticated parsing could use proper YAML/JSON parsers
		modelKeyPattern := regexp.MustCompile(`(?i)["']?(model|model_name|model_id|base_model|llm|embedding_model)["']?\s*[:=]\s*["']([^"'\n]+)["']`)

		matches := modelKeyPattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) >= 3 {
				modelName := match[2]

				// Skip obvious non-model values
				if modelName == "true" || modelName == "false" || modelName == "null" ||
					strings.HasPrefix(modelName, "${") || strings.HasPrefix(modelName, "{{") {
					continue
				}

				// Skip if already seen
				if seen[modelName] {
					continue
				}
				seen[modelName] = true

				source := "config"
				// Detect source from model name patterns
				if strings.Contains(modelName, "/") {
					source = "huggingface"
				} else if strings.HasPrefix(modelName, "gpt-") || strings.HasPrefix(modelName, "text-") {
					source = "api"
				} else if strings.HasPrefix(modelName, "claude-") {
					source = "api"
				} else if strings.HasPrefix(modelName, "gemini-") {
					source = "api"
				}

				model := MLModel{
					Name:   modelName,
					Source: source,
					CodeLocation: &CodeLocation{
						File: relPath,
					},
				}

				models = append(models, model)
			}
		}

		return nil
	})

	return models
}

// enrichWithHuggingFaceMetadata queries HuggingFace API for model metadata
func (s *TechnologyScanner) enrichWithHuggingFaceMetadata(ctx context.Context, models []MLModel) {
	client := &http.Client{Timeout: 10 * time.Second}

	for i := range models {
		if models[i].Source != "huggingface" {
			continue
		}

		// Query HuggingFace API
		url := fmt.Sprintf("https://huggingface.co/api/models/%s", models[i].Name)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}

		var hfModel struct {
			ID          string   `json:"id"`
			Author      string   `json:"author"`
			License     string   `json:"license"`
			Tags        []string `json:"tags"`
			PipelineTag string   `json:"pipeline_tag"`
			ModelIndex  []struct {
				Name string `json:"name"`
			} `json:"model-index"`
			CardData struct {
				License   string   `json:"license"`
				Datasets  []string `json:"datasets"`
				BaseModel string   `json:"base_model"`
				Language  []string `json:"language"`
				Tags      []string `json:"tags"`
			} `json:"cardData"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&hfModel); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Enrich model with metadata
		if hfModel.License != "" {
			models[i].License = hfModel.License
		} else if hfModel.CardData.License != "" {
			models[i].License = hfModel.CardData.License
		}

		if hfModel.PipelineTag != "" {
			models[i].Task = hfModel.PipelineTag
		}

		if hfModel.CardData.BaseModel != "" {
			models[i].BaseModel = hfModel.CardData.BaseModel
		}

		if len(hfModel.CardData.Datasets) > 0 {
			models[i].Datasets = hfModel.CardData.Datasets
		}

		models[i].SourceURL = fmt.Sprintf("https://huggingface.co/%s", models[i].Name)

		// Create model card summary
		models[i].ModelCard = &ModelCard{
			Author:   hfModel.Author,
			License:  models[i].License,
			Datasets: models[i].Datasets,
		}

		// Detect architecture from tags
		for _, tag := range hfModel.Tags {
			if tag == "transformers" || tag == "pytorch" || tag == "tensorflow" {
				continue
			}
			if strings.Contains(tag, "bert") || strings.Contains(tag, "gpt") ||
				strings.Contains(tag, "llama") || strings.Contains(tag, "t5") {
				models[i].Architecture = tag
				break
			}
		}
	}
}

// modelExists checks if a model with the given name already exists
func (s *TechnologyScanner) modelExists(models []MLModel, name string) bool {
	for _, m := range models {
		if m.Name == name {
			return true
		}
	}
	return false
}

// enrichWithReplicateMetadata queries Replicate API for model metadata
func (s *TechnologyScanner) enrichWithReplicateMetadata(ctx context.Context, models []MLModel) {
	client := &http.Client{Timeout: 10 * time.Second}

	for i := range models {
		if models[i].Source != "replicate" {
			continue
		}

		// Parse model name (format: owner/model or owner/model:version)
		modelName := models[i].Name
		if idx := strings.Index(modelName, ":"); idx > 0 {
			modelName = modelName[:idx] // Remove version
		}

		// Query Replicate API (no auth required for public info)
		url := fmt.Sprintf("https://api.replicate.com/v1/models/%s", modelName)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}

		var repModel struct {
			Owner       string `json:"owner"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Visibility  string `json:"visibility"`
			URL         string `json:"url"`
			GitHub      struct {
				URL string `json:"url"`
			} `json:"github_url"`
			Paper struct {
				URL string `json:"url"`
			} `json:"paper_url"`
			License struct {
				URL  string `json:"url"`
				Name string `json:"name"`
			} `json:"license_url"`
			LatestVersion struct {
				ID string `json:"id"`
			} `json:"latest_version"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&repModel); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Enrich model
		models[i].SourceURL = fmt.Sprintf("https://replicate.com/%s", modelName)

		if repModel.License.Name != "" {
			models[i].License = repModel.License.Name
		}

		// Create model card
		models[i].ModelCard = &ModelCard{
			Description: repModel.Description,
			Author:      repModel.Owner,
			License:     repModel.License.Name,
		}
	}
}

// enrichModelSourceURLs adds source URLs for models from various registries
func (s *TechnologyScanner) enrichModelSourceURLs(models []MLModel) {
	for i := range models {
		if models[i].SourceURL != "" {
			continue
		}

		switch models[i].Source {
		case "huggingface":
			if !strings.HasPrefix(models[i].Name, "http") {
				models[i].SourceURL = fmt.Sprintf("https://huggingface.co/%s", models[i].Name)
			}
		case "pytorch_hub":
			// Format: owner/repo/model -> https://github.com/owner/repo
			parts := strings.Split(models[i].Name, "/")
			if len(parts) >= 2 {
				models[i].SourceURL = fmt.Sprintf("https://github.com/%s/%s", parts[0], parts[1])
			}
		case "tensorflow_hub":
			if strings.HasPrefix(models[i].Name, "http") {
				models[i].SourceURL = models[i].Name
			} else {
				models[i].SourceURL = fmt.Sprintf("https://tfhub.dev/%s", models[i].Name)
			}
		case "replicate":
			modelName := models[i].Name
			if idx := strings.Index(modelName, ":"); idx > 0 {
				modelName = modelName[:idx]
			}
			models[i].SourceURL = fmt.Sprintf("https://replicate.com/%s", modelName)
		case "wandb":
			// Format: entity/project/artifact:version
			models[i].SourceURL = fmt.Sprintf("https://wandb.ai/artifacts/%s", models[i].Name)
		case "kaggle":
			models[i].SourceURL = fmt.Sprintf("https://kaggle.com/models/%s", models[i].Name)
		case "civitai":
			if strings.HasPrefix(models[i].Name, "http") {
				models[i].SourceURL = models[i].Name
			}
		case "nvidia_ngc":
			if strings.HasPrefix(models[i].Name, "nvcr.io/") {
				models[i].SourceURL = fmt.Sprintf("https://catalog.ngc.nvidia.com/orgs/nvidia/containers/%s",
					strings.TrimPrefix(models[i].Name, "nvcr.io/"))
			}
		case "ollama":
			models[i].SourceURL = fmt.Sprintf("https://ollama.com/library/%s", models[i].Name)
		}

		// Set registry info from ModelRegistries
		if registry, ok := ModelRegistries[models[i].Source]; ok {
			if models[i].Metadata == nil {
				models[i].Metadata = make(map[string]any)
			}
			models[i].Metadata["registry"] = registry.Name
			models[i].Metadata["registry_description"] = registry.Description
		}
	}
}

// runFrameworksFeature detects AI/ML frameworks
func (s *TechnologyScanner) runFrameworksFeature(ctx context.Context, repoPath string, result *Result) {
	summary := &FrameworksSummary{
		ByCategory: make(map[string]int),
	}

	frameworks := s.detectFrameworks(repoPath)
	result.Findings.Frameworks = frameworks

	for _, fw := range frameworks {
		summary.TotalFrameworks++
		summary.Detected = append(summary.Detected, fw.Name)
		summary.ByCategory[fw.Category]++
	}

	result.Summary.Frameworks = summary
}

// Framework detection patterns
var frameworkPatterns = []struct {
	Name     string
	Category string
	Packages []string
	Patterns []*regexp.Regexp
}{
	{
		Name:     "PyTorch",
		Category: "deep_learning",
		Packages: []string{"torch", "torchvision", "torchaudio"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`import\s+torch`),
			regexp.MustCompile(`from\s+torch\s+import`),
		},
	},
	{
		Name:     "TensorFlow",
		Category: "deep_learning",
		Packages: []string{"tensorflow", "tf"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`import\s+tensorflow`),
			regexp.MustCompile(`from\s+tensorflow\s+import`),
		},
	},
	{
		Name:     "JAX",
		Category: "deep_learning",
		Packages: []string{"jax", "flax"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`import\s+jax`),
			regexp.MustCompile(`from\s+jax\s+import`),
		},
	},
	{
		Name:     "HuggingFace Transformers",
		Category: "llm_framework",
		Packages: []string{"transformers", "huggingface_hub"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`from\s+transformers\s+import`),
			regexp.MustCompile(`import\s+transformers`),
		},
	},
	{
		Name:     "LangChain",
		Category: "llm_framework",
		Packages: []string{"langchain", "langchain_core", "langchain_openai"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`from\s+langchain`),
			regexp.MustCompile(`import\s+langchain`),
		},
	},
	{
		Name:     "LlamaIndex",
		Category: "llm_framework",
		Packages: []string{"llama_index", "llama-index"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`from\s+llama_index`),
			regexp.MustCompile(`import\s+llama_index`),
		},
	},
	{
		Name:     "ONNX Runtime",
		Category: "inference",
		Packages: []string{"onnxruntime", "onnx"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`import\s+onnxruntime`),
			regexp.MustCompile(`import\s+onnx`),
		},
	},
	{
		Name:     "MLflow",
		Category: "mlops",
		Packages: []string{"mlflow"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`import\s+mlflow`),
			regexp.MustCompile(`from\s+mlflow`),
		},
	},
	{
		Name:     "Weights & Biases",
		Category: "mlops",
		Packages: []string{"wandb"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`import\s+wandb`),
		},
	},
	{
		Name:     "OpenAI SDK",
		Category: "llm_api",
		Packages: []string{"openai"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`from\s+openai\s+import`),
			regexp.MustCompile(`import\s+openai`),
		},
	},
	{
		Name:     "Anthropic SDK",
		Category: "llm_api",
		Packages: []string{"anthropic"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`from\s+anthropic\s+import`),
			regexp.MustCompile(`import\s+anthropic`),
		},
	},
	{
		Name:     "Scikit-learn",
		Category: "ml_classic",
		Packages: []string{"sklearn", "scikit-learn"},
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`from\s+sklearn`),
			regexp.MustCompile(`import\s+sklearn`),
		},
	},
}

// detectFrameworks scans for AI/ML framework usage
func (s *TechnologyScanner) detectFrameworks(repoPath string) []Framework {
	var frameworks []Framework
	detected := make(map[string]*Framework)

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".py" && ext != ".js" && ext != ".ts" {
			return nil
		}

		if strings.Contains(path, "node_modules") || strings.Contains(path, "__pycache__") ||
			strings.Contains(path, ".git") || strings.Contains(path, "venv") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		relPath, _ := filepath.Rel(repoPath, path)
		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			for _, fp := range frameworkPatterns {
				for _, pattern := range fp.Patterns {
					if pattern.MatchString(line) {
						if existing, ok := detected[fp.Name]; ok {
							existing.CodeLocations = append(existing.CodeLocations, CodeLocation{
								File: relPath,
								Line: lineNum,
							})
						} else {
							fw := &Framework{
								Name:     fp.Name,
								Category: fp.Category,
								Package:  fp.Packages[0],
								CodeLocations: []CodeLocation{{
									File: relPath,
									Line: lineNum,
								}},
							}
							detected[fp.Name] = fw
						}
						break
					}
				}
			}
		}

		return nil
	})

	for _, fw := range detected {
		frameworks = append(frameworks, *fw)
	}

	return frameworks
}

// runDatasetsFeature detects training datasets
func (s *TechnologyScanner) runDatasetsFeature(ctx context.Context, repoPath string, result *Result) {
	summary := &DatasetsSummary{
		BySource: make(map[string]int),
	}

	datasets := s.detectDatasets(repoPath)
	result.Findings.Datasets = datasets

	for _, ds := range datasets {
		summary.TotalDatasets++
		summary.BySource[ds.Source]++
		if ds.License != "" {
			summary.WithLicense++
		}
	}

	result.Summary.Datasets = summary
}

// Dataset loading patterns
var datasetPatterns = []struct {
	Pattern     *regexp.Regexp
	Source      string
	ExtractName func([]string) string
}{
	// HuggingFace datasets
	{
		Pattern:     regexp.MustCompile(`load_dataset\s*\(\s*["']([^"']+)["']`),
		Source:      "huggingface",
		ExtractName: func(m []string) string { return m[1] },
	},
	// TensorFlow datasets
	{
		Pattern:     regexp.MustCompile(`tfds\.load\s*\(\s*["']([^"']+)["']`),
		Source:      "tensorflow",
		ExtractName: func(m []string) string { return m[1] },
	},
}

// detectDatasets scans for dataset loading patterns
func (s *TechnologyScanner) detectDatasets(repoPath string) []Dataset {
	var datasets []Dataset
	seen := make(map[string]bool)

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".py" {
			return nil
		}

		if strings.Contains(path, "__pycache__") || strings.Contains(path, ".git") ||
			strings.Contains(path, "venv") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(repoPath, path)
		lines := strings.Split(string(content), "\n")

		for lineNum, line := range lines {
			for _, pattern := range datasetPatterns {
				matches := pattern.Pattern.FindStringSubmatch(line)
				if len(matches) > 0 {
					dsName := pattern.ExtractName(matches)

					if seen[dsName] {
						continue
					}
					seen[dsName] = true

					ds := Dataset{
						Name:   dsName,
						Source: pattern.Source,
						CodeLocation: &CodeLocation{
							File: relPath,
							Line: lineNum + 1,
						},
					}

					if pattern.Source == "huggingface" {
						ds.SourceURL = fmt.Sprintf("https://huggingface.co/datasets/%s", dsName)
					}

					datasets = append(datasets, ds)
				}
			}
		}

		return nil
	})

	return datasets
}

// runSecurityFeature checks for AI/ML security issues
func (s *TechnologyScanner) runSecurityFeature(ctx context.Context, repoPath string, result *Result) {
	summary := &SecuritySummary{
		ByCategory: make(map[string]int),
	}

	var findings []SecurityFinding

	// Check for unsafe pickle files
	if s.config.Security.CheckPickleFiles {
		pickleFindings := s.checkPickleFiles(result.Findings.Models)
		findings = append(findings, pickleFindings...)
	}

	// Check for unsafe model loading patterns
	if s.config.Security.DetectUnsafeLoading {
		loadingFindings := s.checkUnsafeLoading(repoPath)
		findings = append(findings, loadingFindings...)
	}

	// Check for API key exposure
	if s.config.Security.CheckAPIKeyExposure {
		keyFindings := s.checkAPIKeyExposure(repoPath)
		findings = append(findings, keyFindings...)
	}

	result.Findings.Security = findings

	for _, f := range findings {
		summary.TotalFindings++
		summary.ByCategory[f.Category]++
		switch f.Severity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		}
		if f.Category == "pickle_rce" {
			summary.UnsafePickles++
		}
		if f.Category == "api_key_exposure" {
			summary.ExposedAPIKeys++
		}
	}

	result.Summary.Security = summary
}

// checkPickleFiles identifies unsafe pickle model files
func (s *TechnologyScanner) checkPickleFiles(models []MLModel) []SecurityFinding {
	var findings []SecurityFinding

	for _, model := range models {
		if model.SecurityRisk == "high" && model.Format == "pickle" {
			findings = append(findings, SecurityFinding{
				ID:          fmt.Sprintf("MLSEC-001-%s", model.Name),
				Title:       "Unsafe Pickle Model File",
				Description: fmt.Sprintf("Model file '%s' uses pickle format which allows arbitrary code execution during loading", model.FilePath),
				Severity:    "high",
				Category:    "pickle_rce",
				File:        model.FilePath,
				ModelName:   model.Name,
				Remediation: "Convert model to SafeTensors format using: safetensors.torch.save_file(model.state_dict(), 'model.safetensors')",
				References: []string{
					"https://huggingface.co/docs/safetensors/",
					"https://arxiv.org/abs/2302.08575",
				},
			})
		}
	}

	return findings
}

// checkUnsafeLoading detects unsafe model loading patterns
func (s *TechnologyScanner) checkUnsafeLoading(repoPath string) []SecurityFinding {
	var findings []SecurityFinding

	// Pattern for torch.load without weights_only=True
	unsafeLoadPattern := regexp.MustCompile(`torch\.load\s*\([^)]*\)`)
	safeLoadPattern := regexp.MustCompile(`weights_only\s*=\s*True`)

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".py" {
			return nil
		}

		if strings.Contains(path, "__pycache__") || strings.Contains(path, "venv") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(repoPath, path)
		lines := strings.Split(string(content), "\n")

		for lineNum, line := range lines {
			if unsafeLoadPattern.MatchString(line) && !safeLoadPattern.MatchString(line) {
				findings = append(findings, SecurityFinding{
					ID:          fmt.Sprintf("MLSEC-002-%s-%d", relPath, lineNum),
					Title:       "Unsafe torch.load() Usage",
					Description: "torch.load() called without weights_only=True allows arbitrary code execution",
					Severity:    "high",
					Category:    "unsafe_loading",
					File:        relPath,
					Line:        lineNum + 1,
					Remediation: "Use torch.load(path, weights_only=True) or convert to SafeTensors",
					References: []string{
						"https://pytorch.org/docs/stable/generated/torch.load.html",
					},
				})
			}
		}

		return nil
	})

	return findings
}

// checkAPIKeyExposure detects hardcoded API keys
func (s *TechnologyScanner) checkAPIKeyExposure(repoPath string) []SecurityFinding {
	var findings []SecurityFinding

	// Patterns for API keys
	apiKeyPatterns := []struct {
		Name    string
		Pattern *regexp.Regexp
	}{
		{"OpenAI", regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`)},
		{"Anthropic", regexp.MustCompile(`sk-ant-[a-zA-Z0-9-]{20,}`)},
		{"HuggingFace", regexp.MustCompile(`hf_[a-zA-Z0-9]{20,}`)},
		{"Cohere", regexp.MustCompile(`[a-zA-Z0-9]{40}`)}, // Generic but context-sensitive
	}

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".py" && ext != ".js" && ext != ".ts" && ext != ".env" {
			return nil
		}

		// Skip test files and examples
		if strings.Contains(path, "test") || strings.Contains(path, "example") {
			return nil
		}

		if strings.Contains(path, "node_modules") || strings.Contains(path, "__pycache__") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(repoPath, path)
		lines := strings.Split(string(content), "\n")

		for lineNum, line := range lines {
			// Skip comments
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
				continue
			}

			for _, kp := range apiKeyPatterns {
				// Only check OpenAI and Anthropic patterns (most specific)
				if kp.Name != "OpenAI" && kp.Name != "Anthropic" && kp.Name != "HuggingFace" {
					continue
				}

				if kp.Pattern.MatchString(line) {
					// Skip if it's an environment variable reference
					if strings.Contains(line, "os.environ") || strings.Contains(line, "getenv") ||
						strings.Contains(line, "process.env") {
						continue
					}

					findings = append(findings, SecurityFinding{
						ID:          fmt.Sprintf("MLSEC-003-%s-%d", relPath, lineNum),
						Title:       fmt.Sprintf("Hardcoded %s API Key", kp.Name),
						Description: fmt.Sprintf("Potential %s API key found in source code", kp.Name),
						Severity:    "critical",
						Category:    "api_key_exposure",
						File:        relPath,
						Line:        lineNum + 1,
						Remediation: "Use environment variables or a secrets manager instead of hardcoding API keys",
					})
					break
				}
			}
		}

		return nil
	})

	return findings
}

// runGovernanceFeature checks AI governance requirements
func (s *TechnologyScanner) runGovernanceFeature(ctx context.Context, repoPath string, result *Result) {
	summary := &GovernanceSummary{}
	var findings []GovernanceFinding

	for _, model := range result.Findings.Models {
		// Check for missing model cards
		if s.config.Governance.RequireModelCards && model.ModelCard == nil && model.Source == "huggingface" {
			findings = append(findings, GovernanceFinding{
				ID:          fmt.Sprintf("MLGOV-001-%s", model.Name),
				Title:       "Missing Model Card",
				Description: fmt.Sprintf("Model '%s' does not have associated model card documentation", model.Name),
				Severity:    "medium",
				Category:    "missing_model_card",
				ModelName:   model.Name,
				Remediation: "Add model card documentation to the model repository",
			})
			summary.MissingModelCards++
		}

		// Check for missing licenses
		if s.config.Governance.RequireLicense && model.License == "" && model.Source != "api" {
			findings = append(findings, GovernanceFinding{
				ID:          fmt.Sprintf("MLGOV-002-%s", model.Name),
				Title:       "Missing License Information",
				Description: fmt.Sprintf("Model '%s' does not have license information", model.Name),
				Severity:    "medium",
				Category:    "missing_license",
				ModelName:   model.Name,
				Remediation: "Verify and document the license for this model",
			})
			summary.MissingLicenses++
		}

		// Check for blocked licenses
		if model.License != "" && len(s.config.Governance.BlockedLicenses) > 0 {
			for _, blocked := range s.config.Governance.BlockedLicenses {
				if strings.EqualFold(model.License, blocked) {
					findings = append(findings, GovernanceFinding{
						ID:          fmt.Sprintf("MLGOV-003-%s", model.Name),
						Title:       "Blocked License",
						Description: fmt.Sprintf("Model '%s' uses blocked license: %s", model.Name, model.License),
						Severity:    "high",
						Category:    "blocked_license",
						ModelName:   model.Name,
						Remediation: "Replace with a model using an approved license",
					})
					summary.BlockedLicenses++
					break
				}
			}
		}

		// Check for missing dataset info
		if s.config.Governance.RequireDatasetInfo && len(model.Datasets) == 0 && model.Source == "huggingface" {
			findings = append(findings, GovernanceFinding{
				ID:          fmt.Sprintf("MLGOV-004-%s", model.Name),
				Title:       "Missing Training Dataset Information",
				Description: fmt.Sprintf("Model '%s' does not have training dataset provenance information", model.Name),
				Severity:    "low",
				Category:    "missing_dataset_info",
				ModelName:   model.Name,
				Remediation: "Document the training datasets used for this model",
			})
			summary.MissingDatasetInfo++
		}
	}

	result.Findings.Governance = findings
	summary.TotalIssues = len(findings)
	result.Summary.Governance = summary
}

// =============================================================================
// General Technology Detection Feature (migrated from health scanner)
// =============================================================================

// Config file patterns for technology detection
var configPatterns = map[string]Technology{
	// JavaScript/Node.js
	"package.json":       {Name: "Node.js", Category: "runtime", Confidence: 90, Source: "config"},
	"package-lock.json":  {Name: "npm", Category: "package-manager", Confidence: 90, Source: "config"},
	"yarn.lock":          {Name: "Yarn", Category: "package-manager", Confidence: 90, Source: "config"},
	"pnpm-lock.yaml":     {Name: "pnpm", Category: "package-manager", Confidence: 90, Source: "config"},
	"tsconfig.json":      {Name: "TypeScript", Category: "language", Confidence: 95, Source: "config"},
	"next.config.js":     {Name: "Next.js", Category: "framework", Confidence: 95, Source: "config"},
	"next.config.mjs":    {Name: "Next.js", Category: "framework", Confidence: 95, Source: "config"},
	"nuxt.config.js":     {Name: "Nuxt.js", Category: "framework", Confidence: 95, Source: "config"},
	"nuxt.config.ts":     {Name: "Nuxt.js", Category: "framework", Confidence: 95, Source: "config"},
	"angular.json":       {Name: "Angular", Category: "framework", Confidence: 95, Source: "config"},
	"svelte.config.js":   {Name: "Svelte", Category: "framework", Confidence: 95, Source: "config"},
	"jest.config.js":     {Name: "Jest", Category: "testing", Confidence: 90, Source: "config"},
	"vitest.config.ts":   {Name: "Vitest", Category: "testing", Confidence: 90, Source: "config"},
	"tailwind.config.js": {Name: "Tailwind CSS", Category: "styling", Confidence: 90, Source: "config"},

	// Python
	"requirements.txt": {Name: "Python", Category: "language", Confidence: 90, Source: "config"},
	"pyproject.toml":   {Name: "Python", Category: "language", Confidence: 90, Source: "config"},
	"Pipfile":          {Name: "Pipenv", Category: "package-manager", Confidence: 90, Source: "config"},
	"poetry.lock":      {Name: "Poetry", Category: "package-manager", Confidence: 90, Source: "config"},
	"uv.lock":          {Name: "uv", Category: "package-manager", Confidence: 90, Source: "config"},

	// Go
	"go.mod": {Name: "Go", Category: "language", Confidence: 95, Source: "config"},
	"go.sum": {Name: "Go Modules", Category: "package-manager", Confidence: 90, Source: "config"},

	// Rust
	"Cargo.toml": {Name: "Rust", Category: "language", Confidence: 95, Source: "config"},
	"Cargo.lock": {Name: "Cargo", Category: "package-manager", Confidence: 90, Source: "config"},

	// Java/JVM
	"pom.xml":          {Name: "Maven", Category: "build-tool", Confidence: 90, Source: "config"},
	"build.gradle":     {Name: "Gradle", Category: "build-tool", Confidence: 90, Source: "config"},
	"build.gradle.kts": {Name: "Gradle Kotlin", Category: "build-tool", Confidence: 90, Source: "config"},

	// Ruby
	"Gemfile":      {Name: "Ruby", Category: "language", Confidence: 90, Source: "config"},
	"Gemfile.lock": {Name: "Bundler", Category: "package-manager", Confidence: 90, Source: "config"},

	// PHP
	"composer.json": {Name: "PHP", Category: "language", Confidence: 90, Source: "config"},
	"composer.lock": {Name: "Composer", Category: "package-manager", Confidence: 90, Source: "config"},

	// .NET
	"*.csproj": {Name: "C#/.NET", Category: "language", Confidence: 90, Source: "config"},

	// Infrastructure
	"Dockerfile":          {Name: "Docker", Category: "container", Confidence: 95, Source: "config"},
	"docker-compose.yml":  {Name: "Docker Compose", Category: "container", Confidence: 95, Source: "config"},
	"docker-compose.yaml": {Name: "Docker Compose", Category: "container", Confidence: 95, Source: "config"},
	"serverless.yml":      {Name: "Serverless Framework", Category: "iac", Confidence: 95, Source: "config"},
	"serverless.yaml":     {Name: "Serverless Framework", Category: "iac", Confidence: 95, Source: "config"},
	"Pulumi.yaml":         {Name: "Pulumi", Category: "iac", Confidence: 95, Source: "config"},
	"cdk.json":            {Name: "AWS CDK", Category: "iac", Confidence: 95, Source: "config"},
}

// File extension to technology mapping
var extensionMap = map[string]Technology{
	".py":     {Name: "Python", Category: "language", Confidence: 80, Source: "extension"},
	".js":     {Name: "JavaScript", Category: "language", Confidence: 80, Source: "extension"},
	".ts":     {Name: "TypeScript", Category: "language", Confidence: 85, Source: "extension"},
	".tsx":    {Name: "React/TypeScript", Category: "framework", Confidence: 85, Source: "extension"},
	".jsx":    {Name: "React", Category: "framework", Confidence: 85, Source: "extension"},
	".go":     {Name: "Go", Category: "language", Confidence: 85, Source: "extension"},
	".rs":     {Name: "Rust", Category: "language", Confidence: 85, Source: "extension"},
	".java":   {Name: "Java", Category: "language", Confidence: 85, Source: "extension"},
	".kt":     {Name: "Kotlin", Category: "language", Confidence: 85, Source: "extension"},
	".scala":  {Name: "Scala", Category: "language", Confidence: 85, Source: "extension"},
	".rb":     {Name: "Ruby", Category: "language", Confidence: 80, Source: "extension"},
	".php":    {Name: "PHP", Category: "language", Confidence: 80, Source: "extension"},
	".cs":     {Name: "C#", Category: "language", Confidence: 85, Source: "extension"},
	".swift":  {Name: "Swift", Category: "language", Confidence: 85, Source: "extension"},
	".c":      {Name: "C", Category: "language", Confidence: 80, Source: "extension"},
	".cpp":    {Name: "C++", Category: "language", Confidence: 80, Source: "extension"},
	".cc":     {Name: "C++", Category: "language", Confidence: 80, Source: "extension"},
	".vue":    {Name: "Vue.js", Category: "framework", Confidence: 90, Source: "extension"},
	".svelte": {Name: "Svelte", Category: "framework", Confidence: 90, Source: "extension"},
	".tf":     {Name: "Terraform", Category: "iac", Confidence: 90, Source: "extension"},
	".sol":    {Name: "Solidity", Category: "language", Confidence: 90, Source: "extension"},
	".zig":    {Name: "Zig", Category: "language", Confidence: 90, Source: "extension"},
}

// runTechnologyFeature detects general technologies (languages, frameworks, databases, etc.)
func (s *TechnologyScanner) runTechnologyFeature(ctx context.Context, repoPath, sbomPath string, result *Result) {
	var techs []Technology

	if s.config.Technology.ScanConfig {
		techs = append(techs, s.detectFromConfigFiles(repoPath)...)
	}

	if s.config.Technology.ScanSBOM && sbomPath != "" {
		techs = append(techs, s.detectFromSBOM(sbomPath)...)
	}

	if s.config.Technology.ScanExtensions {
		techs = append(techs, s.detectFromFileExtensions(repoPath)...)
	}

	// Deduplicate and consolidate
	techs = s.consolidateTechnologies(techs)

	// Build summary
	summary := s.buildTechnologySummary(techs)

	result.Summary.Technology = summary
	result.Findings.Technology = techs
}

// detectFromConfigFiles detects technologies from config files
func (s *TechnologyScanner) detectFromConfigFiles(repoPath string) []Technology {
	var techs []Technology

	for pattern, tech := range configPatterns {
		// Skip glob patterns for now (handled separately)
		if strings.Contains(pattern, "*") {
			continue
		}

		// Check for directory patterns
		if strings.Contains(pattern, "/") {
			filePath := filepath.Join(repoPath, pattern)
			if _, err := os.Stat(filePath); err == nil {
				techs = append(techs, tech)
			}
			continue
		}

		// Direct file check
		filePath := filepath.Join(repoPath, pattern)
		if _, err := os.Stat(filePath); err == nil {
			techs = append(techs, tech)
		}
	}

	// Check for Terraform files
	if matches, _ := filepath.Glob(filepath.Join(repoPath, "*.tf")); len(matches) > 0 {
		techs = append(techs, Technology{Name: "Terraform", Category: "iac", Confidence: 90, Source: "config"})
	}

	// Check for .csproj files
	if matches, _ := filepath.Glob(filepath.Join(repoPath, "*.csproj")); len(matches) > 0 {
		techs = append(techs, Technology{Name: "C#/.NET", Category: "language", Confidence: 90, Source: "config"})
	}

	// Check for GitHub Actions
	if _, err := os.Stat(filepath.Join(repoPath, ".github", "workflows")); err == nil {
		techs = append(techs, Technology{Name: "GitHub Actions", Category: "ci-cd", Confidence: 95, Source: "config"})
	}

	// Check for GitLab CI
	if _, err := os.Stat(filepath.Join(repoPath, ".gitlab-ci.yml")); err == nil {
		techs = append(techs, Technology{Name: "GitLab CI", Category: "ci-cd", Confidence: 95, Source: "config"})
	}

	// Check for CircleCI
	if _, err := os.Stat(filepath.Join(repoPath, ".circleci", "config.yml")); err == nil {
		techs = append(techs, Technology{Name: "CircleCI", Category: "ci-cd", Confidence: 95, Source: "config"})
	}

	return techs
}

// detectFromSBOM detects technologies from SBOM components
func (s *TechnologyScanner) detectFromSBOM(sbomPath string) []Technology {
	var techs []Technology

	data, err := os.ReadFile(sbomPath)
	if err != nil {
		return techs
	}

	var sbomData struct {
		Components []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"components"`
	}

	if err := json.Unmarshal(data, &sbomData); err != nil {
		return techs
	}

	sbomPatterns := map[string]Technology{
		"react":         {Name: "React", Category: "framework", Confidence: 95, Source: "sbom"},
		"vue":           {Name: "Vue.js", Category: "framework", Confidence: 95, Source: "sbom"},
		"angular":       {Name: "Angular", Category: "framework", Confidence: 95, Source: "sbom"},
		"express":       {Name: "Express.js", Category: "framework", Confidence: 95, Source: "sbom"},
		"fastify":       {Name: "Fastify", Category: "framework", Confidence: 95, Source: "sbom"},
		"django":        {Name: "Django", Category: "framework", Confidence: 95, Source: "sbom"},
		"flask":         {Name: "Flask", Category: "framework", Confidence: 95, Source: "sbom"},
		"fastapi":       {Name: "FastAPI", Category: "framework", Confidence: 95, Source: "sbom"},
		"spring":        {Name: "Spring", Category: "framework", Confidence: 95, Source: "sbom"},
		"rails":         {Name: "Ruby on Rails", Category: "framework", Confidence: 95, Source: "sbom"},
		"gin-gonic":     {Name: "Gin", Category: "framework", Confidence: 95, Source: "sbom"},
		"fiber":         {Name: "Fiber", Category: "framework", Confidence: 95, Source: "sbom"},
		"postgres":      {Name: "PostgreSQL", Category: "database", Confidence: 85, Source: "sbom"},
		"pg":            {Name: "PostgreSQL", Category: "database", Confidence: 85, Source: "sbom"},
		"mysql":         {Name: "MySQL", Category: "database", Confidence: 85, Source: "sbom"},
		"mongodb":       {Name: "MongoDB", Category: "database", Confidence: 85, Source: "sbom"},
		"mongoose":      {Name: "MongoDB", Category: "database", Confidence: 85, Source: "sbom"},
		"redis":         {Name: "Redis", Category: "database", Confidence: 85, Source: "sbom"},
		"sqlite":        {Name: "SQLite", Category: "database", Confidence: 85, Source: "sbom"},
		"elasticsearch": {Name: "Elasticsearch", Category: "database", Confidence: 85, Source: "sbom"},
		"aws-sdk":       {Name: "AWS SDK", Category: "cloud", Confidence: 90, Source: "sbom"},
		"boto3":         {Name: "AWS SDK (Python)", Category: "cloud", Confidence: 90, Source: "sbom"},
		"@azure":        {Name: "Azure SDK", Category: "cloud", Confidence: 90, Source: "sbom"},
		"@google-cloud": {Name: "Google Cloud SDK", Category: "cloud", Confidence: 90, Source: "sbom"},
		"openai":        {Name: "OpenAI", Category: "ai", Confidence: 95, Source: "sbom"},
		"anthropic":     {Name: "Anthropic", Category: "ai", Confidence: 95, Source: "sbom"},
		"langchain":     {Name: "LangChain", Category: "ai", Confidence: 95, Source: "sbom"},
		"tensorflow":    {Name: "TensorFlow", Category: "ai", Confidence: 95, Source: "sbom"},
		"pytorch":       {Name: "PyTorch", Category: "ai", Confidence: 95, Source: "sbom"},
		"torch":         {Name: "PyTorch", Category: "ai", Confidence: 95, Source: "sbom"},
		"transformers":  {Name: "Hugging Face Transformers", Category: "ai", Confidence: 95, Source: "sbom"},
	}

	seen := make(map[string]bool)
	for _, comp := range sbomData.Components {
		nameLower := strings.ToLower(comp.Name)
		for pattern, tech := range sbomPatterns {
			if strings.Contains(nameLower, pattern) && !seen[tech.Name] {
				t := tech
				t.Version = comp.Version
				techs = append(techs, t)
				seen[tech.Name] = true
			}
		}
	}

	return techs
}

// detectFromFileExtensions detects technologies from file extensions
func (s *TechnologyScanner) detectFromFileExtensions(repoPath string) []Technology {
	extCounts := make(map[string]int)

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == "vendor" || base == ".git" ||
				base == "dist" || base == "build" || base == "__pycache__" ||
				base == ".venv" || base == "venv" || base == "target" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != "" {
			extCounts[ext]++
		}
		return nil
	})

	var techs []Technology
	for ext, count := range extCounts {
		if tech, ok := extensionMap[ext]; ok && count >= 3 {
			techs = append(techs, tech)
		}
	}

	return techs
}

// consolidateTechnologies deduplicates and consolidates technologies
func (s *TechnologyScanner) consolidateTechnologies(techs []Technology) []Technology {
	techMap := make(map[string]Technology)
	for _, t := range techs {
		existing, ok := techMap[t.Name]
		if !ok || t.Confidence > existing.Confidence {
			techMap[t.Name] = t
		}
	}

	var result []Technology
	for _, t := range techMap {
		result = append(result, t)
	}

	// Sort by confidence (descending)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Confidence > result[i].Confidence {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// buildTechnologySummary creates a summary of detected technologies
func (s *TechnologyScanner) buildTechnologySummary(techs []Technology) *TechnologySummary {
	summary := &TechnologySummary{
		TotalTechnologies: len(techs),
		ByCategory:        make(map[string]int),
	}

	// Count by name for top technologies
	techCounts := make(map[string]int)

	for _, t := range techs {
		summary.ByCategory[t.Category]++
		techCounts[t.Name]++

		switch t.Category {
		case "language":
			summary.PrimaryLanguages = append(summary.PrimaryLanguages, t.Name)
		case "framework":
			summary.Frameworks = append(summary.Frameworks, t.Name)
		case "database":
			summary.Databases = append(summary.Databases, t.Name)
		case "cloud":
			summary.CloudServices = append(summary.CloudServices, t.Name)
		}
	}

	// Build top technologies list (top 3 by frequency)
	type techFreq struct {
		name  string
		count int
	}
	var techList []techFreq
	for name, count := range techCounts {
		techList = append(techList, techFreq{name, count})
	}
	sort.Slice(techList, func(i, j int) bool {
		return techList[i].count > techList[j].count
	})

	topN := 3
	if len(techList) < topN {
		topN = len(techList)
	}
	for i := 0; i < topN; i++ {
		summary.TopTechnologies = append(summary.TopTechnologies, techList[i].name)
	}

	return summary
}

// =============================================================================
// Infrastructure/Microservice Detection Feature
// =============================================================================

// runInfrastructureFeature detects microservice communication patterns
func (s *TechnologyScanner) runInfrastructureFeature(ctx context.Context, repoPath string, onStatus func(string), result *Result) {
	onStatus("Detecting microservice communication patterns...")

	// Create microservice scanner
	msScanner := common.NewMicroserviceScanner(common.MicroserviceConfig{
		RAGPath:  filepath.Join(os.Getenv("ZERO_HOME"), "rag", "architecture", "microservices"),
		CacheDir: filepath.Join(os.Getenv("ZERO_HOME"), ".cache", "microservices"),
		Timeout:  120 * time.Second,
		OnStatus: onStatus,
	})

	// Run the scan
	msResult := msScanner.Scan(ctx, repoPath)

	if msResult.Error != nil {
		result.Summary.Infrastructure = &InfrastructureSummary{
			Error: msResult.Error.Error(),
		}
		return
	}

	// Convert to tech-id types
	infraFindings := &InfrastructureFindings{}

	// Convert services
	for _, svc := range msResult.Services {
		endpoints := make([]Endpoint, len(svc.Endpoints))
		for i, ep := range svc.Endpoints {
			endpoints[i] = Endpoint{
				Method:      ep.Method,
				Path:        ep.Path,
				Description: ep.Description,
				File:        ep.File,
				Line:        ep.Line,
			}
		}
		infraFindings.Services = append(infraFindings.Services, ServiceDefinition{
			Name:      svc.Name,
			Type:      svc.Type,
			Endpoints: endpoints,
			Port:      svc.Port,
			File:      svc.File,
			Line:      svc.Line,
			Framework: svc.Framework,
			Metadata:  svc.Metadata,
		})
	}

	// Convert dependencies
	for _, dep := range msResult.Dependencies {
		locations := make([]CodeLocation, len(dep.Locations))
		for i, loc := range dep.Locations {
			locations[i] = CodeLocation{
				File:    loc.File,
				Line:    loc.Line,
				Column:  loc.Column,
				Snippet: loc.Snippet,
			}
		}
		infraFindings.Dependencies = append(infraFindings.Dependencies, ServiceDependency{
			SourceService: dep.SourceService,
			TargetService: dep.TargetService,
			TargetURL:     dep.TargetURL,
			Type:          dep.Type,
			Method:        dep.Method,
			Locations:     locations,
			Confidence:    dep.Confidence,
		})
	}

	// Convert API contracts
	for _, contract := range msResult.APIContracts {
		endpoints := make([]Endpoint, len(contract.Endpoints))
		for i, ep := range contract.Endpoints {
			endpoints[i] = Endpoint{
				Method:      ep.Method,
				Path:        ep.Path,
				Description: ep.Description,
				File:        ep.File,
				Line:        ep.Line,
			}
		}
		infraFindings.APIContracts = append(infraFindings.APIContracts, APIContract{
			Name:      contract.Name,
			Type:      contract.Type,
			Version:   contract.Version,
			File:      contract.File,
			BaseURL:   contract.BaseURL,
			Endpoints: endpoints,
			Services:  contract.Services,
		})
	}

	// Convert message queues
	for _, mq := range msResult.MessageQueues {
		locations := make([]CodeLocation, len(mq.Locations))
		for i, loc := range mq.Locations {
			locations[i] = CodeLocation{
				File:    loc.File,
				Line:    loc.Line,
				Column:  loc.Column,
				Snippet: loc.Snippet,
			}
		}
		infraFindings.MessageQueues = append(infraFindings.MessageQueues, MessageQueueUsage{
			QueueType:     mq.QueueType,
			Role:          mq.Role,
			TopicOrQueue:  mq.TopicOrQueue,
			Brokers:       mq.Brokers,
			ConsumerGroup: mq.ConsumerGroup,
			Locations:     locations,
		})
	}

	result.Findings.Infrastructure = infraFindings

	// Build summary
	summary := &InfrastructureSummary{
		TotalServices:        msResult.Summary.TotalServices,
		TotalDependencies:    msResult.Summary.TotalDependencies,
		TotalAPIContracts:    msResult.Summary.TotalAPIContracts,
		TotalMessageQueues:   msResult.Summary.TotalMessageQueues,
		ByType:               msResult.Summary.CommunicationTypes,
	}

	result.Summary.Infrastructure = summary

	// Report findings
	if summary.TotalServices > 0 || summary.TotalDependencies > 0 || summary.TotalMessageQueues > 0 {
		onStatus(fmt.Sprintf("Found %d services, %d dependencies, %d API contracts, %d message queues",
			summary.TotalServices, summary.TotalDependencies, summary.TotalAPIContracts, summary.TotalMessageQueues))
	} else {
		onStatus("No microservice patterns detected")
	}
}
