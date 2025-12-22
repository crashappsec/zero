package techid

import (
	"regexp"

	"github.com/crashappsec/zero/pkg/analysis/rag"
)

// RAGPatterns holds compiled patterns loaded from RAG files
type RAGPatterns struct {
	ModelLoadPatterns map[string][]CompiledPattern
	APIProviders      map[string]APIProviderInfo
	ModelFileFormats  map[string]ModelFormatInfo
	ModelRegistries   map[string]ModelRegistry
	Loaded            bool
}

// CompiledPattern is a regex pattern with metadata
type CompiledPattern struct {
	Name          string
	Pattern       *regexp.Regexp
	ExtractGroup  int
	ExtractGroups []int
	JoinWith      string
	Description   string
}

// ragPatterns holds the singleton RAG patterns instance
var ragPatterns *RAGPatterns

// LoadRAGPatterns loads patterns from RAG files
// Returns cached patterns if already loaded
func LoadRAGPatterns() (*RAGPatterns, error) {
	if ragPatterns != nil && ragPatterns.Loaded {
		return ragPatterns, nil
	}

	ragPatterns = &RAGPatterns{
		ModelLoadPatterns: make(map[string][]CompiledPattern),
		APIProviders:      make(map[string]APIProviderInfo),
		ModelFileFormats:  make(map[string]ModelFormatInfo),
		ModelRegistries:   make(map[string]ModelRegistry),
	}

	loader := rag.NewLoader("")

	// Load model loading patterns
	loadingPatterns, err := loader.LoadModelLoadingPatterns()
	if err == nil {
		for source, patterns := range loadingPatterns.Patterns {
			var compiled []CompiledPattern
			for _, p := range patterns {
				if p.Regex() != nil {
					compiled = append(compiled, CompiledPattern{
						Name:          p.Name,
						Pattern:       p.Regex(),
						ExtractGroup:  p.ExtractGroup,
						ExtractGroups: p.ExtractGroups,
						JoinWith:      p.JoinWith,
						Description:   p.Description,
					})
				}
			}
			ragPatterns.ModelLoadPatterns[source] = compiled
		}
	}

	// Load API providers
	apiProviders, err := loader.LoadAPIProviders()
	if err == nil {
		for name, provider := range apiProviders.Providers {
			ragPatterns.APIProviders[name] = APIProviderInfo{
				Name:     provider.Name,
				EnvVars:  provider.EnvVars,
				Packages: provider.Packages,
			}
		}
	}

	// Load model file formats
	fileFormats, err := loader.LoadModelFileFormats()
	if err == nil {
		for ext, format := range fileFormats.Formats {
			ragPatterns.ModelFileFormats[ext] = ModelFormatInfo{
				Name:       format.Name,
				Format:     format.Format,
				Risk:       format.Risk,
				RiskReason: format.RiskReason,
			}
		}
	}

	// Load model registries
	registries, err := loader.LoadModelRegistries()
	if err == nil {
		for name, registry := range registries.Registries {
			ragPatterns.ModelRegistries[name] = ModelRegistry{
				Name:        registry.Name,
				BaseURL:     registry.BaseURL,
				APIURL:      registry.APIURL,
				HasAPI:      registry.HasAPI,
				Description: registry.Description,
			}
		}
	}

	ragPatterns.Loaded = true
	return ragPatterns, nil
}

// GetModelFileFormatsFromRAG returns file formats from RAG or falls back to hardcoded
func GetModelFileFormatsFromRAG() map[string]ModelFormatInfo {
	patterns, err := LoadRAGPatterns()
	if err != nil || len(patterns.ModelFileFormats) == 0 {
		return ModelFileFormats // Fall back to hardcoded
	}
	return patterns.ModelFileFormats
}

// GetModelRegistriesFromRAG returns registries from RAG or falls back to hardcoded
func GetModelRegistriesFromRAG() map[string]ModelRegistry {
	patterns, err := LoadRAGPatterns()
	if err != nil || len(patterns.ModelRegistries) == 0 {
		return ModelRegistries // Fall back to hardcoded
	}
	return patterns.ModelRegistries
}
