// Package ai provides the consolidated AI/ML security super scanner
// Generates ML-BOM (Machine Learning Bill of Materials) in CycloneDX format
package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/crashappsec/zero/pkg/scanner"
)

const (
	Name        = "ai"
	Description = "AI/ML security analysis and ML-BOM generation"
)

// AIScanner implements the AI/ML super scanner
type AIScanner struct {
	config FeatureConfig
}

// init registers the scanner
func init() {
	scanner.Register(&AIScanner{
		config: DefaultConfig(),
	})
}

// Name returns the scanner name
func (s *AIScanner) Name() string {
	return Name
}

// Description returns the scanner description
func (s *AIScanner) Description() string {
	return Description
}

// Dependencies returns scanner dependencies (none for AI scanner)
func (s *AIScanner) Dependencies() []string {
	return []string{}
}

// EstimateDuration returns estimated scan duration based on file count
func (s *AIScanner) EstimateDuration(fileCount int) time.Duration {
	// Base time + time per file for pattern scanning
	base := 5 * time.Second
	perFile := 10 * time.Millisecond
	return base + time.Duration(fileCount)*perFile
}

// Run executes the AI/ML analysis
func (s *AIScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	startTime := time.Now()

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	repoPath := opts.RepoPath

	// Run each enabled feature
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

	return scanResult, nil
}

// runModelsFeature detects ML models in the repository
func (s *AIScanner) runModelsFeature(ctx context.Context, repoPath string, result *Result) {
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

// detectModelFiles scans for model files by extension
func (s *AIScanner) detectModelFiles(repoPath string) []MLModel {
	var models []MLModel

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip common non-model directories
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "__pycache__" || name == ".venv" || name == "venv" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if formatInfo, ok := ModelFileFormats[ext]; ok {
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
func (s *AIScanner) scanCodeForModels(repoPath string) []MLModel {
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
func (s *AIScanner) scanConfigsForModels(repoPath string) []MLModel {
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
func (s *AIScanner) enrichWithHuggingFaceMetadata(ctx context.Context, models []MLModel) {
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
			ID           string   `json:"id"`
			Author       string   `json:"author"`
			License      string   `json:"license"`
			Tags         []string `json:"tags"`
			PipelineTag  string   `json:"pipeline_tag"`
			ModelIndex   []struct {
				Name string `json:"name"`
			} `json:"model-index"`
			CardData struct {
				License      string   `json:"license"`
				Datasets     []string `json:"datasets"`
				BaseModel    string   `json:"base_model"`
				Language     []string `json:"language"`
				Tags         []string `json:"tags"`
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
			Author:  hfModel.Author,
			License: models[i].License,
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
func (s *AIScanner) modelExists(models []MLModel, name string) bool {
	for _, m := range models {
		if m.Name == name {
			return true
		}
	}
	return false
}

// enrichWithReplicateMetadata queries Replicate API for model metadata
func (s *AIScanner) enrichWithReplicateMetadata(ctx context.Context, models []MLModel) {
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
func (s *AIScanner) enrichModelSourceURLs(models []MLModel) {
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
func (s *AIScanner) runFrameworksFeature(ctx context.Context, repoPath string, result *Result) {
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
func (s *AIScanner) detectFrameworks(repoPath string) []Framework {
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
func (s *AIScanner) runDatasetsFeature(ctx context.Context, repoPath string, result *Result) {
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
func (s *AIScanner) detectDatasets(repoPath string) []Dataset {
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
func (s *AIScanner) runSecurityFeature(ctx context.Context, repoPath string, result *Result) {
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
func (s *AIScanner) checkPickleFiles(models []MLModel) []SecurityFinding {
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
func (s *AIScanner) checkUnsafeLoading(repoPath string) []SecurityFinding {
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
func (s *AIScanner) checkAPIKeyExposure(repoPath string) []SecurityFinding {
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
func (s *AIScanner) runGovernanceFeature(ctx context.Context, repoPath string, result *Result) {
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
