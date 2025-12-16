package techid

// Result holds all AI/ML analysis results (ML-BOM)
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	Technology         *TechnologySummary `json:"technology,omitempty"` // General technology detection
	Models             *ModelsSummary     `json:"models,omitempty"`
	Frameworks         *FrameworksSummary `json:"frameworks,omitempty"`
	Datasets           *DatasetsSummary   `json:"datasets,omitempty"`
	Security           *SecuritySummary   `json:"security,omitempty"`
	Governance         *GovernanceSummary `json:"governance,omitempty"`
	SemgrepRulesLoaded int                `json:"semgrep_rules_loaded,omitempty"` // Number of semgrep rules loaded
	SemgrepFindings    int                `json:"semgrep_findings,omitempty"`     // Findings from semgrep
	Errors             []string           `json:"errors,omitempty"`
}

// Findings holds detailed findings from all features
type Findings struct {
	Technology []Technology        `json:"technology,omitempty"` // General technology detection
	Models     []MLModel           `json:"models,omitempty"`
	Frameworks []Framework         `json:"frameworks,omitempty"`
	Datasets   []Dataset           `json:"datasets,omitempty"`
	Security   []SecurityFinding   `json:"security,omitempty"`
	Governance []GovernanceFinding `json:"governance,omitempty"`
}

// Feature summaries

// TechnologySummary contains general technology detection summary
type TechnologySummary struct {
	TotalTechnologies int            `json:"total_technologies"`
	ByCategory        map[string]int `json:"by_category"`
	TopTechnologies   []string       `json:"top_technologies,omitempty"` // Top 3 most detected technologies
	PrimaryLanguages  []string       `json:"primary_languages,omitempty"`
	Frameworks        []string       `json:"frameworks,omitempty"`
	Databases         []string       `json:"databases,omitempty"`
	CloudServices     []string       `json:"cloud_services,omitempty"`
	Error             string         `json:"error,omitempty"`
}

// Technology represents a detected technology (language, framework, database, etc.)
type Technology struct {
	Name       string `json:"name"`
	Category   string `json:"category"`   // language, framework, database, container, iac, ci-cd, etc.
	Version    string `json:"version,omitempty"`
	Confidence int    `json:"confidence"` // 0-100
	Source     string `json:"source"`     // config, extension, sbom, semgrep
	File       string `json:"file,omitempty"`       // File where detected (for semgrep findings)
	Line       int    `json:"line,omitempty"`       // Line number (for semgrep findings)
	Match      string `json:"match,omitempty"`      // Matched code snippet
}

// ModelsSummary contains ML model inventory summary
type ModelsSummary struct {
	TotalModels      int            `json:"total_models"`
	BySource         map[string]int `json:"by_source"`          // huggingface, local, api, etc.
	ByFormat         map[string]int `json:"by_format"`          // pickle, safetensors, onnx, etc.
	WithModelCard    int            `json:"with_model_card"`
	WithLicense      int            `json:"with_license"`
	WithDatasetInfo  int            `json:"with_dataset_info"`
	LocalModelFiles  int            `json:"local_model_files"`
	APIModels        int            `json:"api_models"`
	Error            string         `json:"error,omitempty"`
}

// FrameworksSummary contains AI/ML framework detection summary
type FrameworksSummary struct {
	TotalFrameworks int            `json:"total_frameworks"`
	Detected        []string       `json:"detected"`
	ByCategory      map[string]int `json:"by_category"` // deep_learning, llm, mlops, etc.
	Error           string         `json:"error,omitempty"`
}

// DatasetsSummary contains training dataset detection summary
type DatasetsSummary struct {
	TotalDatasets   int            `json:"total_datasets"`
	BySource        map[string]int `json:"by_source"` // huggingface, local, url
	WithLicense     int            `json:"with_license"`
	WithProvenance  int            `json:"with_provenance"`
	Error           string         `json:"error,omitempty"`
}

// SecuritySummary contains AI security findings summary
type SecuritySummary struct {
	TotalFindings    int            `json:"total_findings"`
	Critical         int            `json:"critical"`
	High             int            `json:"high"`
	Medium           int            `json:"medium"`
	Low              int            `json:"low"`
	ByCategory       map[string]int `json:"by_category"`
	UnsafePickles    int            `json:"unsafe_pickles"`
	ExposedAPIKeys   int            `json:"exposed_api_keys"`
	Error            string         `json:"error,omitempty"`
}

// GovernanceSummary contains AI governance check summary
type GovernanceSummary struct {
	TotalIssues         int `json:"total_issues"`
	MissingModelCards   int `json:"missing_model_cards"`
	MissingLicenses     int `json:"missing_licenses"`
	BlockedLicenses     int `json:"blocked_licenses"`
	MissingDatasetInfo  int `json:"missing_dataset_info"`
	Error               string `json:"error,omitempty"`
}

// Finding types

// MLModel represents a detected ML model
type MLModel struct {
	Name            string          `json:"name"`
	Version         string          `json:"version,omitempty"`
	Source          string          `json:"source"`                    // huggingface, tensorflow_hub, local, api
	SourceURL       string          `json:"source_url,omitempty"`
	Format          string          `json:"format,omitempty"`          // pickle, safetensors, onnx, gguf, keras
	FilePath        string          `json:"file_path,omitempty"`       // For local models
	CodeLocation    *CodeLocation   `json:"code_location,omitempty"`   // Where model is loaded in code
	License         string          `json:"license,omitempty"`
	BaseModel       string          `json:"base_model,omitempty"`      // For fine-tuned models
	Architecture    string          `json:"architecture,omitempty"`    // transformer, cnn, rnn, etc.
	Task            string          `json:"task,omitempty"`            // text-generation, classification, etc.
	Datasets        []string        `json:"datasets,omitempty"`        // Training datasets
	ModelCard       *ModelCard      `json:"model_card,omitempty"`
	SecurityRisk    string          `json:"security_risk,omitempty"`   // high, medium, low
	SecurityNotes   []string        `json:"security_notes,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
}

// CodeLocation represents where something is referenced in code
type CodeLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`
	Snippet string `json:"snippet,omitempty"`
}

// ModelCard contains model card metadata (subset of CycloneDX modelCard)
type ModelCard struct {
	Description         string            `json:"description,omitempty"`
	Author              string            `json:"author,omitempty"`
	License             string            `json:"license,omitempty"`
	Datasets            []string          `json:"datasets,omitempty"`
	Metrics             map[string]any    `json:"metrics,omitempty"`
	Limitations         string            `json:"limitations,omitempty"`
	EthicalConsiderations string          `json:"ethical_considerations,omitempty"`
	EnvironmentalImpact string            `json:"environmental_impact,omitempty"`
	IntendedUse         string            `json:"intended_use,omitempty"`
	OutOfScopeUse       string            `json:"out_of_scope_use,omitempty"`
}

// Framework represents a detected AI/ML framework
type Framework struct {
	Name          string        `json:"name"`
	Version       string        `json:"version,omitempty"`
	Category      string        `json:"category"`      // deep_learning, llm_framework, mlops, vector_db
	Package       string        `json:"package"`       // pip/npm package name
	CodeLocations []CodeLocation `json:"code_locations,omitempty"`
	UsagePatterns []string      `json:"usage_patterns,omitempty"` // training, inference, fine-tuning
}

// Dataset represents a detected training/evaluation dataset
type Dataset struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`      // huggingface, local, url
	SourceURL   string   `json:"source_url,omitempty"`
	License     string   `json:"license,omitempty"`
	Split       string   `json:"split,omitempty"`       // train, test, validation
	Size        string   `json:"size,omitempty"`
	Description string   `json:"description,omitempty"`
	UsedBy      []string `json:"used_by,omitempty"`     // Model names that use this dataset
	CodeLocation *CodeLocation `json:"code_location,omitempty"`
}

// SecurityFinding represents an AI/ML security issue
type SecurityFinding struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Severity    string        `json:"severity"` // critical, high, medium, low
	Category    string        `json:"category"` // pickle_rce, api_key_exposure, unsafe_loading, prompt_injection
	File        string        `json:"file,omitempty"`
	Line        int           `json:"line,omitempty"`
	ModelName   string        `json:"model_name,omitempty"`
	Remediation string        `json:"remediation,omitempty"`
	References  []string      `json:"references,omitempty"`
}

// GovernanceFinding represents an AI governance issue
type GovernanceFinding struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Category    string `json:"category"` // missing_model_card, missing_license, blocked_license, missing_dataset_info
	ModelName   string `json:"model_name,omitempty"`
	Remediation string `json:"remediation,omitempty"`
}

// Model file formats and their security characteristics
// NOTE: .bin removed (too generic, causes false positives)
// NOTE: .pb requires context check (must be saved_model.pb or in SavedModel directory)
var ModelFileFormats = map[string]ModelFormatInfo{
	".pt":          {Name: "PyTorch Pickle", Format: "pickle", Risk: "high", RiskReason: "Arbitrary code execution during deserialization"},
	".pth":         {Name: "PyTorch Pickle", Format: "pickle", Risk: "high", RiskReason: "Arbitrary code execution during deserialization"},
	".pkl":         {Name: "Python Pickle", Format: "pickle", Risk: "high", RiskReason: "Arbitrary code execution during deserialization"},
	".pickle":      {Name: "Python Pickle", Format: "pickle", Risk: "high", RiskReason: "Arbitrary code execution during deserialization"},
	// .bin removed - too generic, matches test data and other binary files
	".safetensors": {Name: "SafeTensors", Format: "safetensors", Risk: "low", RiskReason: "Secure format, no code execution"},
	".onnx":        {Name: "ONNX", Format: "onnx", Risk: "medium", RiskReason: "Custom operators may execute code"},
	".gguf":        {Name: "GGUF", Format: "gguf", Risk: "low", RiskReason: "Inference-only format"},
	".ggml":        {Name: "GGML", Format: "ggml", Risk: "low", RiskReason: "Inference-only format"},
	".h5":          {Name: "HDF5/Keras", Format: "keras", Risk: "medium", RiskReason: "May contain Lambda layers with code"},
	".keras":       {Name: "Keras", Format: "keras", Risk: "medium", RiskReason: "May contain Lambda layers with code"},
	// .pb handled specially in detectModelFiles - only saved_model.pb files
	".tflite":      {Name: "TensorFlow Lite", Format: "tflite", Risk: "low", RiskReason: "Mobile inference format"},
	".mlmodel":     {Name: "Core ML", Format: "coreml", Risk: "low", RiskReason: "Apple inference format"},
}

// ModelFormatInfo describes a model file format
type ModelFormatInfo struct {
	Name       string `json:"name"`
	Format     string `json:"format"`
	Risk       string `json:"risk"`        // high, medium, low
	RiskReason string `json:"risk_reason"`
}

// API model providers
var APIModelProviders = map[string]APIProviderInfo{
	"openai":    {Name: "OpenAI", EnvVars: []string{"OPENAI_API_KEY"}, Packages: []string{"openai"}},
	"anthropic": {Name: "Anthropic", EnvVars: []string{"ANTHROPIC_API_KEY"}, Packages: []string{"anthropic"}},
	"google":    {Name: "Google AI", EnvVars: []string{"GOOGLE_API_KEY", "GEMINI_API_KEY"}, Packages: []string{"google-generativeai", "vertexai"}},
	"cohere":    {Name: "Cohere", EnvVars: []string{"COHERE_API_KEY"}, Packages: []string{"cohere"}},
	"mistral":   {Name: "Mistral", EnvVars: []string{"MISTRAL_API_KEY"}, Packages: []string{"mistralai"}},
	"replicate": {Name: "Replicate", EnvVars: []string{"REPLICATE_API_TOKEN"}, Packages: []string{"replicate"}},
	"together":  {Name: "Together AI", EnvVars: []string{"TOGETHER_API_KEY"}, Packages: []string{"together"}},
	"groq":      {Name: "Groq", EnvVars: []string{"GROQ_API_KEY"}, Packages: []string{"groq"}},
}

// APIProviderInfo describes an API model provider
type APIProviderInfo struct {
	Name     string   `json:"name"`
	EnvVars  []string `json:"env_vars"`
	Packages []string `json:"packages"`
}

// ModelRegistry represents a model hosting registry
type ModelRegistry struct {
	Name        string `json:"name"`
	BaseURL     string `json:"base_url"`
	APIURL      string `json:"api_url,omitempty"`
	HasAPI      bool   `json:"has_api"`
	Description string `json:"description"`
}

// ModelRegistries defines supported model registries
var ModelRegistries = map[string]ModelRegistry{
	"huggingface": {
		Name:        "HuggingFace Hub",
		BaseURL:     "https://huggingface.co",
		APIURL:      "https://huggingface.co/api/models",
		HasAPI:      true,
		Description: "Largest open ML model repository with 400k+ models",
	},
	"tensorflow_hub": {
		Name:        "TensorFlow Hub",
		BaseURL:     "https://tfhub.dev",
		APIURL:      "",
		HasAPI:      false,
		Description: "Google's repository for reusable TensorFlow models",
	},
	"pytorch_hub": {
		Name:        "PyTorch Hub",
		BaseURL:     "https://pytorch.org/hub",
		APIURL:      "",
		HasAPI:      false,
		Description: "Official PyTorch model repository",
	},
	"replicate": {
		Name:        "Replicate",
		BaseURL:     "https://replicate.com",
		APIURL:      "https://api.replicate.com/v1/models",
		HasAPI:      true,
		Description: "Cloud ML platform with versioned models",
	},
	"wandb": {
		Name:        "Weights & Biases",
		BaseURL:     "https://wandb.ai",
		APIURL:      "https://api.wandb.ai/artifacts",
		HasAPI:      true,
		Description: "MLOps platform with model artifacts",
	},
	"mlflow": {
		Name:        "MLflow Model Registry",
		BaseURL:     "",
		APIURL:      "",
		HasAPI:      false,
		Description: "Self-hosted MLOps model registry",
	},
	"civitai": {
		Name:        "Civitai",
		BaseURL:     "https://civitai.com",
		APIURL:      "https://civitai.com/api/v1/models",
		HasAPI:      true,
		Description: "Community platform for Stable Diffusion models",
	},
	"kaggle": {
		Name:        "Kaggle Models",
		BaseURL:     "https://kaggle.com/models",
		APIURL:      "",
		HasAPI:      false,
		Description: "Kaggle's ML model repository",
	},
	"ollama": {
		Name:        "Ollama Library",
		BaseURL:     "https://ollama.com/library",
		APIURL:      "",
		HasAPI:      false,
		Description: "Local LLM model library for ollama",
	},
	"nvidia_ngc": {
		Name:        "NVIDIA NGC",
		BaseURL:     "https://catalog.ngc.nvidia.com",
		APIURL:      "",
		HasAPI:      true,
		Description: "NVIDIA's GPU-optimized model catalog",
	},
	"aws_jumpstart": {
		Name:        "AWS SageMaker JumpStart",
		BaseURL:     "",
		APIURL:      "",
		HasAPI:      false,
		Description: "AWS ML model catalog for SageMaker",
	},
	"azure_ml": {
		Name:        "Azure ML Model Catalog",
		BaseURL:     "",
		APIURL:      "",
		HasAPI:      false,
		Description: "Microsoft Azure ML model repository",
	},
}
