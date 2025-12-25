package cyclonedx

import "fmt"

// ModelCard represents CycloneDX ML-BOM modelCard object
type ModelCard struct {
	ModelParameters      *ModelParameters      `json:"modelParameters,omitempty"`
	QuantitativeAnalysis *QuantitativeAnalysis `json:"quantitativeAnalysis,omitempty"`
	Considerations       *Considerations       `json:"considerations,omitempty"`
}

// ModelParameters contains ML model parameters
type ModelParameters struct {
	Approach           *ModelApproach `json:"approach,omitempty"`
	Task               string         `json:"task,omitempty"`
	ArchitectureFamily string         `json:"architectureFamily,omitempty"`
	ModelArchitecture  string         `json:"modelArchitecture,omitempty"`
	Datasets           []DatasetRef   `json:"datasets,omitempty"`
	Inputs             []ModelIO      `json:"inputs,omitempty"`
	Outputs            []ModelIO      `json:"outputs,omitempty"`
}

// ModelApproach describes the learning approach
type ModelApproach struct {
	Type string `json:"type,omitempty"` // supervised, unsupervised, reinforcement-learning, semi-supervised, self-supervised
}

// ModelApproach type constants
const (
	ApproachSupervised             = "supervised"
	ApproachUnsupervised           = "unsupervised"
	ApproachReinforcementLearning  = "reinforcement-learning"
	ApproachSemiSupervised         = "semi-supervised"
	ApproachSelfSupervised         = "self-supervised"
)

// DatasetRef references a dataset
type DatasetRef struct {
	Type           string         `json:"type,omitempty"` // dataset
	Ref            string         `json:"ref,omitempty"`  // bom-ref to dataset component
	Contents       *DataContents  `json:"contents,omitempty"`
	Classification string         `json:"classification,omitempty"` // training, validation, testing
	Governance     *DataGovernance `json:"governance,omitempty"`
}

// ModelIO represents model input/output format
type ModelIO struct {
	Format string `json:"format,omitempty"`
}

// QuantitativeAnalysis contains performance metrics
type QuantitativeAnalysis struct {
	PerformanceMetrics []PerformanceMetric `json:"performanceMetrics,omitempty"`
	Graphics           *Graphics           `json:"graphics,omitempty"`
}

// PerformanceMetric represents a model metric
type PerformanceMetric struct {
	Type               string              `json:"type,omitempty"`
	Value              string              `json:"value,omitempty"`
	Slice              string              `json:"slice,omitempty"`
	ConfidenceInterval *ConfidenceInterval `json:"confidenceInterval,omitempty"`
}

// ConfidenceInterval represents a statistical confidence interval
type ConfidenceInterval struct {
	LowerBound string `json:"lowerBound,omitempty"`
	UpperBound string `json:"upperBound,omitempty"`
}

// Graphics contains visual performance representations
type Graphics struct {
	Description string    `json:"description,omitempty"`
	Collection  []Graphic `json:"collection,omitempty"`
}

// Graphic represents a graphic
type Graphic struct {
	Name  string      `json:"name,omitempty"`
	Image *Attachment `json:"image,omitempty"`
}

// Considerations contains ethical and usage considerations
type Considerations struct {
	Users                 []string               `json:"users,omitempty"`
	UseCases              []string               `json:"useCases,omitempty"`
	TechnicalLimitations  []string               `json:"technicalLimitations,omitempty"`
	PerformanceTradeoffs  []string               `json:"performanceTradeoffs,omitempty"`
	EthicalConsiderations []EthicalConsideration `json:"ethicalConsiderations,omitempty"`
	FairnessAssessments   []FairnessAssessment   `json:"fairnessAssessments,omitempty"`
}

// EthicalConsideration represents an ethical consideration
type EthicalConsideration struct {
	Name               string `json:"name,omitempty"`
	MitigationStrategy string `json:"mitigationStrategy,omitempty"`
}

// FairnessAssessment represents a fairness assessment
type FairnessAssessment struct {
	GroupAtRisk        string `json:"groupAtRisk,omitempty"`
	Benefits           string `json:"benefits,omitempty"`
	Harms              string `json:"harms,omitempty"`
	MitigationStrategy string `json:"mitigationStrategy,omitempty"`
}

// NewModelCard creates a new ModelCard
func NewModelCard() *ModelCard {
	return &ModelCard{
		ModelParameters:      &ModelParameters{},
		QuantitativeAnalysis: &QuantitativeAnalysis{},
		Considerations:       &Considerations{},
	}
}

// WithApproach sets the learning approach
func (mc *ModelCard) WithApproach(approachType string) *ModelCard {
	if mc.ModelParameters == nil {
		mc.ModelParameters = &ModelParameters{}
	}
	mc.ModelParameters.Approach = &ModelApproach{Type: approachType}
	return mc
}

// WithTask sets the model task
func (mc *ModelCard) WithTask(task string) *ModelCard {
	if mc.ModelParameters == nil {
		mc.ModelParameters = &ModelParameters{}
	}
	mc.ModelParameters.Task = task
	return mc
}

// WithArchitecture sets the model architecture
func (mc *ModelCard) WithArchitecture(family, architecture string) *ModelCard {
	if mc.ModelParameters == nil {
		mc.ModelParameters = &ModelParameters{}
	}
	mc.ModelParameters.ArchitectureFamily = family
	mc.ModelParameters.ModelArchitecture = architecture
	return mc
}

// WithDataset adds a dataset reference
func (mc *ModelCard) WithDataset(ref, classification string) *ModelCard {
	if mc.ModelParameters == nil {
		mc.ModelParameters = &ModelParameters{}
	}
	mc.ModelParameters.Datasets = append(mc.ModelParameters.Datasets, DatasetRef{
		Type:           "dataset",
		Ref:            ref,
		Classification: classification,
	})
	return mc
}

// WithInput adds an input format
func (mc *ModelCard) WithInput(format string) *ModelCard {
	if mc.ModelParameters == nil {
		mc.ModelParameters = &ModelParameters{}
	}
	mc.ModelParameters.Inputs = append(mc.ModelParameters.Inputs, ModelIO{Format: format})
	return mc
}

// WithOutput adds an output format
func (mc *ModelCard) WithOutput(format string) *ModelCard {
	if mc.ModelParameters == nil {
		mc.ModelParameters = &ModelParameters{}
	}
	mc.ModelParameters.Outputs = append(mc.ModelParameters.Outputs, ModelIO{Format: format})
	return mc
}

// WithMetric adds a performance metric
func (mc *ModelCard) WithMetric(metricType, value string) *ModelCard {
	if mc.QuantitativeAnalysis == nil {
		mc.QuantitativeAnalysis = &QuantitativeAnalysis{}
	}
	mc.QuantitativeAnalysis.PerformanceMetrics = append(
		mc.QuantitativeAnalysis.PerformanceMetrics,
		PerformanceMetric{Type: metricType, Value: value},
	)
	return mc
}

// WithUseCase adds a use case
func (mc *ModelCard) WithUseCase(useCase string) *ModelCard {
	if mc.Considerations == nil {
		mc.Considerations = &Considerations{}
	}
	mc.Considerations.UseCases = append(mc.Considerations.UseCases, useCase)
	return mc
}

// WithLimitation adds a technical limitation
func (mc *ModelCard) WithLimitation(limitation string) *ModelCard {
	if mc.Considerations == nil {
		mc.Considerations = &Considerations{}
	}
	mc.Considerations.TechnicalLimitations = append(mc.Considerations.TechnicalLimitations, limitation)
	return mc
}

// WithEthicalConsideration adds an ethical consideration
func (mc *ModelCard) WithEthicalConsideration(name, mitigation string) *ModelCard {
	if mc.Considerations == nil {
		mc.Considerations = &Considerations{}
	}
	mc.Considerations.EthicalConsiderations = append(
		mc.Considerations.EthicalConsiderations,
		EthicalConsideration{Name: name, MitigationStrategy: mitigation},
	)
	return mc
}

// MLModelToComponent converts ML model data to a CycloneDX component
func MLModelToComponent(name, version, source, sourceURL, format, architecture, task, license string) Component {
	c := NewMLModelComponent(name, version)
	c.Description = fmt.Sprintf("%s model from %s", architecture, source)

	// Add license if present
	if license != "" {
		c.AddLicense(license)
	}

	// Add source URL as external reference
	if sourceURL != "" {
		c.AddExternalRef(ExternalRefWebsite, sourceURL)
	}

	// Add properties for source and format
	if source != "" {
		c.AddProperty("zero:source", source)
	}
	if format != "" {
		c.AddProperty("zero:format", format)
	}

	// Create model card
	modelCard := NewModelCard()
	if architecture != "" {
		modelCard.WithArchitecture(inferArchitectureFamily(architecture), architecture)
	}
	if task != "" {
		modelCard.WithTask(task)
	}
	c.ModelCard = modelCard

	return c
}

// DatasetToComponent converts dataset data to a CycloneDX component
func DatasetToComponent(name, source, sourceURL, license, description string) Component {
	c := NewDataComponent(name)
	c.Description = description

	if license != "" {
		c.AddLicense(license)
	}

	if sourceURL != "" {
		c.AddExternalRef(ExternalRefWebsite, sourceURL)
	}

	if source != "" {
		c.AddProperty("zero:source", source)
	}

	return c
}

// FrameworkToComponent converts framework data to a CycloneDX component
func FrameworkToComponent(name, version, category, packageName string) Component {
	c := Component{
		Type:    ComponentTypeFramework,
		Name:    name,
		Version: version,
		BOMRef:  fmt.Sprintf("framework/%s@%s", name, version),
	}

	if category != "" {
		c.AddProperty("zero:category", category)
	}
	if packageName != "" {
		c.AddProperty("zero:package", packageName)
	}

	return c
}

// inferArchitectureFamily infers the architecture family from architecture name
func inferArchitectureFamily(architecture string) string {
	// Map common architectures to families
	families := map[string][]string{
		"transformer": {"bert", "gpt", "llama", "mistral", "t5", "roberta", "albert", "electra", "transformer"},
		"cnn":         {"resnet", "vgg", "inception", "efficientnet", "mobilenet", "convnext", "cnn"},
		"rnn":         {"lstm", "gru", "rnn"},
		"gan":         {"gan", "stylegan", "dcgan", "wgan"},
		"diffusion":   {"stable-diffusion", "dalle", "midjourney", "diffusion"},
		"autoencoder": {"vae", "autoencoder"},
	}

	archLower := architecture
	for family, patterns := range families {
		for _, pattern := range patterns {
			if containsIgnoreCase(archLower, pattern) {
				return family
			}
		}
	}
	return "other"
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsIgnoreCase(s[1:], substr))
}
