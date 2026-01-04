// Package techid provides the consolidated technology identification super scanner
// Includes AI/ML security and ML-BOM generation
package techid

import "time"

// FeatureConfig holds configuration for all technology identification features
type FeatureConfig struct {
	Technology     TechnologyDetectionConfig `json:"technology"`     // General technology detection
	Models         ModelsConfig              `json:"models"`         // ML model detection
	Frameworks     FrameworksConfig          `json:"frameworks"`     // AI/ML framework detection
	Datasets       DatasetsConfig            `json:"datasets"`
	Security       SecurityConfig            `json:"security"`
	Governance     GovernanceConfig          `json:"governance"`
	Infrastructure InfrastructureConfig      `json:"infrastructure"` // Microservice mapping
	Semgrep        SemgrepScanConfig         `json:"semgrep"`        // Semgrep integration
}

// SemgrepScanConfig configures semgrep integration for technology detection
type SemgrepScanConfig struct {
	Enabled      bool          `json:"enabled"`       // Use semgrep for detection (required)
	RefreshRules bool          `json:"refresh_rules"` // Auto-refresh rules on each scan
	ForceRefresh bool          `json:"force_refresh"` // Force rule refresh even if cached
	CacheTTL     time.Duration `json:"cache_ttl"`     // How long to cache generated rules
}

// TechnologyDetectionConfig configures general technology discovery
type TechnologyDetectionConfig struct {
	Enabled        bool `json:"enabled"`
	ScanExtensions bool `json:"scan_extensions"` // Detect from file extensions
	ScanConfig     bool `json:"scan_config"`     // Detect from config files
	ScanSBOM       bool `json:"scan_sbom"`       // Detect from SBOM
}

// ModelsConfig configures ML model detection and inventory
type ModelsConfig struct {
	Enabled            bool `json:"enabled"`
	DetectModelFiles   bool `json:"detect_model_files"`   // Scan for .pt, .onnx, .safetensors, .gguf
	ScanCodePatterns   bool `json:"scan_code_patterns"`   // Detect model loading in code
	ScanConfigs        bool `json:"scan_configs"`         // Check YAML/JSON for model refs
	QueryHuggingFace   bool `json:"query_huggingface"`    // Fetch metadata from HF API
	QueryTFHub         bool `json:"query_tf_hub"`         // Fetch metadata from TensorFlow Hub
	ExtractModelCards  bool `json:"extract_model_cards"`  // Parse model card metadata
	TrackBaseModels    bool `json:"track_base_models"`    // Track fine-tuning lineage
}

// FrameworksConfig configures AI/ML framework detection
type FrameworksConfig struct {
	Enabled          bool `json:"enabled"`
	DetectPyTorch    bool `json:"detect_pytorch"`
	DetectTensorFlow bool `json:"detect_tensorflow"`
	DetectJAX        bool `json:"detect_jax"`
	DetectHuggingFace bool `json:"detect_huggingface"`
	DetectLangChain  bool `json:"detect_langchain"`
	DetectLlamaIndex bool `json:"detect_llamaindex"`
	DetectONNX       bool `json:"detect_onnx"`
	DetectMLFlow     bool `json:"detect_mlflow"`
}

// DatasetsConfig configures training dataset detection
type DatasetsConfig struct {
	Enabled              bool `json:"enabled"`
	DetectHFDatasets     bool `json:"detect_hf_datasets"`     // HuggingFace datasets library
	ExtractFromModelCard bool `json:"extract_from_model_card"` // Parse model cards for dataset info
	ScanDataFiles        bool `json:"scan_data_files"`         // Detect .parquet, .csv, .jsonl
}

// SecurityConfig configures AI/ML security analysis
type SecurityConfig struct {
	Enabled              bool `json:"enabled"`
	CheckPickleFiles     bool `json:"check_pickle_files"`     // Flag unsafe pickle models
	CheckModelProvenance bool `json:"check_model_provenance"` // Verify model sources
	DetectPromptInjection bool `json:"detect_prompt_injection"` // Scan for prompt injection vulns
	CheckAPIKeyExposure  bool `json:"check_api_key_exposure"`  // LLM API keys in code
	DetectUnsafeLoading  bool `json:"detect_unsafe_loading"`   // torch.load without weights_only
}

// GovernanceConfig configures AI governance checks
type GovernanceConfig struct {
	Enabled             bool     `json:"enabled"`
	RequireModelCards   bool     `json:"require_model_cards"`   // Flag models without documentation
	RequireLicense      bool     `json:"require_license"`       // Flag models without license info
	BlockedLicenses     []string `json:"blocked_licenses"`      // Licenses to flag
	RequireDatasetInfo  bool     `json:"require_dataset_info"`  // Flag models without dataset provenance
}

// InfrastructureConfig configures microservice mapping
type InfrastructureConfig struct {
	Enabled           bool `json:"enabled"`
	DetectAPIContracts bool `json:"detect_api_contracts"` // OpenAPI, GraphQL, Proto files
	DetectHTTPClients bool `json:"detect_http_clients"`  // HTTP client usage patterns
	DetectGRPC        bool `json:"detect_grpc"`          // gRPC client/server patterns
	DetectMessageQueues bool `json:"detect_message_queues"` // Kafka, RabbitMQ, SQS, etc.
	DetectServices    bool `json:"detect_services"`       // Docker Compose, K8s services
	BuildDependencyGraph bool `json:"build_dependency_graph"` // Build service dependency graph
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
		Semgrep: SemgrepScanConfig{
			Enabled:      true,           // Use semgrep (required)
			RefreshRules: true,           // Auto-refresh rules
			ForceRefresh: false,          // Don't force, use cache if valid
			CacheTTL:     24 * time.Hour, // Cache for 24 hours
		},
		Technology: TechnologyDetectionConfig{
			Enabled:        true,
			ScanExtensions: true,
			ScanConfig:     true,
			ScanSBOM:       true,
		},
		Models: ModelsConfig{
			Enabled:            true,
			DetectModelFiles:   true,
			ScanCodePatterns:   true,
			ScanConfigs:        true,
			QueryHuggingFace:   true,
			QueryTFHub:         false, // Off by default - slower
			ExtractModelCards:  true,
			TrackBaseModels:    true,
		},
		Frameworks: FrameworksConfig{
			Enabled:          true,
			DetectPyTorch:    true,
			DetectTensorFlow: true,
			DetectJAX:        true,
			DetectHuggingFace: true,
			DetectLangChain:  true,
			DetectLlamaIndex: true,
			DetectONNX:       true,
			DetectMLFlow:     true,
		},
		Datasets: DatasetsConfig{
			Enabled:              true,
			DetectHFDatasets:     true,
			ExtractFromModelCard: true,
			ScanDataFiles:        false, // Off by default - can be slow
		},
		Security: SecurityConfig{
			Enabled:              true,
			CheckPickleFiles:     true,
			CheckModelProvenance: true,
			DetectPromptInjection: true,
			CheckAPIKeyExposure:  true,
			DetectUnsafeLoading:  true,
		},
		Governance: GovernanceConfig{
			Enabled:             true,
			RequireModelCards:   false, // Off by default
			RequireLicense:      true,
			BlockedLicenses:     []string{},
			RequireDatasetInfo:  false, // Off by default
		},
		Infrastructure: InfrastructureConfig{
			Enabled:             true,
			DetectAPIContracts:  true,
			DetectHTTPClients:   true,
			DetectGRPC:          true,
			DetectMessageQueues: true,
			DetectServices:      true,
			BuildDependencyGraph: true,
		},
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Semgrep.Enabled = false          // Skip semgrep for quick scans
	cfg.Technology.ScanExtensions = false // Skip file extension scan (slow)
	cfg.Models.QueryHuggingFace = false
	cfg.Models.ExtractModelCards = false
	cfg.Datasets.Enabled = false
	cfg.Governance.Enabled = false
	cfg.Infrastructure.Enabled = false   // Skip infrastructure for quick scans
	return cfg
}

// SecurityConfig returns security-focused config
func SecurityOnlyConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Security.Enabled = true
	cfg.Security.CheckPickleFiles = true
	cfg.Security.DetectUnsafeLoading = true
	cfg.Security.CheckAPIKeyExposure = true
	cfg.Governance.Enabled = false
	cfg.Datasets.Enabled = false
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return FeatureConfig{
		Semgrep: SemgrepScanConfig{
			Enabled:      true,
			RefreshRules: true,
			ForceRefresh: false,
			CacheTTL:     24 * time.Hour,
		},
		Technology: TechnologyDetectionConfig{
			Enabled:        true,
			ScanExtensions: true,
			ScanConfig:     true,
			ScanSBOM:       true,
		},
		Models: ModelsConfig{
			Enabled:            true,
			DetectModelFiles:   true,
			ScanCodePatterns:   true,
			ScanConfigs:        true,
			QueryHuggingFace:   true,
			QueryTFHub:         true,
			ExtractModelCards:  true,
			TrackBaseModels:    true,
		},
		Frameworks: FrameworksConfig{
			Enabled:          true,
			DetectPyTorch:    true,
			DetectTensorFlow: true,
			DetectJAX:        true,
			DetectHuggingFace: true,
			DetectLangChain:  true,
			DetectLlamaIndex: true,
			DetectONNX:       true,
			DetectMLFlow:     true,
		},
		Datasets: DatasetsConfig{
			Enabled:              true,
			DetectHFDatasets:     true,
			ExtractFromModelCard: true,
			ScanDataFiles:        true,
		},
		Security: SecurityConfig{
			Enabled:              true,
			CheckPickleFiles:     true,
			CheckModelProvenance: true,
			DetectPromptInjection: true,
			CheckAPIKeyExposure:  true,
			DetectUnsafeLoading:  true,
		},
		Governance: GovernanceConfig{
			Enabled:             true,
			RequireModelCards:   true,
			RequireLicense:      true,
			BlockedLicenses:     []string{"CC-BY-NC-4.0", "CC-BY-NC-SA-4.0"}, // Non-commercial
			RequireDatasetInfo:  true,
		},
		Infrastructure: InfrastructureConfig{
			Enabled:             true,
			DetectAPIContracts:  true,
			DetectHTTPClients:   true,
			DetectGRPC:          true,
			DetectMessageQueues: true,
			DetectServices:      true,
			BuildDependencyGraph: true,
		},
	}
}
