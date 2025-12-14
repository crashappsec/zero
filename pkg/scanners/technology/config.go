// Package technology provides the consolidated technology identification super scanner
// Includes AI/ML security and ML-BOM generation
package technology

// FeatureConfig holds configuration for all AI/ML analysis features
type FeatureConfig struct {
	Models      ModelsConfig      `json:"models"`
	Frameworks  FrameworksConfig  `json:"frameworks"`
	Datasets    DatasetsConfig    `json:"datasets"`
	Security    SecurityConfig    `json:"security"`
	Governance  GovernanceConfig  `json:"governance"`
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

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
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
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Models.QueryHuggingFace = false
	cfg.Models.ExtractModelCards = false
	cfg.Datasets.Enabled = false
	cfg.Governance.Enabled = false
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
	}
}
