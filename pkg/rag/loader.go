// Package rag provides utilities for loading RAG (Retrieval-Augmented Generation)
// knowledge files that configure scanner behavior dynamically.
package rag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// RAGLoader handles loading and caching of RAG configuration files
type RAGLoader struct {
	ragPath string
	cache   map[string]interface{}
	mu      sync.RWMutex
}

// NewLoader creates a new RAG loader
// If ragPath is empty, it attempts to find the rag directory relative to the binary
func NewLoader(ragPath string) *RAGLoader {
	if ragPath == "" {
		// Try to find rag directory
		candidates := []string{
			"rag",                    // Current directory
			"../rag",                 // Parent directory
			"../../rag",              // Two levels up
			"/Users/curphey/zero/rag", // Absolute fallback
		}
		for _, candidate := range candidates {
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				ragPath = candidate
				break
			}
		}
	}
	return &RAGLoader{
		ragPath: ragPath,
		cache:   make(map[string]interface{}),
	}
}

// LoadJSON loads a JSON file from the RAG directory and unmarshals it into the target
func (l *RAGLoader) LoadJSON(relativePath string, target interface{}) error {
	fullPath := filepath.Join(l.ragPath, relativePath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read RAG file %s: %w", relativePath, err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse RAG file %s: %w", relativePath, err)
	}

	return nil
}

// LoadJSONWithCache loads a JSON file with caching
func (l *RAGLoader) LoadJSONWithCache(relativePath string, target interface{}) error {
	l.mu.RLock()
	if cached, ok := l.cache[relativePath]; ok {
		l.mu.RUnlock()
		// Copy cached value to target using JSON round-trip
		data, _ := json.Marshal(cached)
		return json.Unmarshal(data, target)
	}
	l.mu.RUnlock()

	// Load fresh
	if err := l.LoadJSON(relativePath, target); err != nil {
		return err
	}

	// Cache it
	l.mu.Lock()
	l.cache[relativePath] = target
	l.mu.Unlock()

	return nil
}

// ClearCache clears the loader's cache
func (l *RAGLoader) ClearCache() {
	l.mu.Lock()
	l.cache = make(map[string]interface{})
	l.mu.Unlock()
}

// Model Registry Types

// ModelRegistry represents a model hosting registry
type ModelRegistry struct {
	Name             string `json:"name"`
	BaseURL          string `json:"base_url"`
	APIURL           string `json:"api_url"`
	HasAPI           bool   `json:"has_api"`
	TrustLevel       string `json:"trust_level"`
	Description      string `json:"description"`
	Verification     string `json:"verification"`
	ModelURLTemplate string `json:"model_url_template"`
	DocsURL          string `json:"docs_url,omitempty"`
}

// ModelRegistriesConfig holds all model registry definitions
type ModelRegistriesConfig struct {
	Version     string                    `json:"version"`
	Description string                    `json:"description"`
	Registries  map[string]ModelRegistry `json:"registries"`
}

// LoadModelRegistries loads model registry definitions from RAG
func (l *RAGLoader) LoadModelRegistries() (*ModelRegistriesConfig, error) {
	var config ModelRegistriesConfig
	if err := l.LoadJSON("ai-ml/registries/model-registries.json", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Model Loading Pattern Types

// ModelLoadPattern represents a code pattern for detecting model loading
type ModelLoadPattern struct {
	Name          string   `json:"name"`
	Pattern       string   `json:"pattern"`
	ExtractGroup  int      `json:"extract_group,omitempty"`
	ExtractGroups []int    `json:"extract_groups,omitempty"`
	JoinWith      string   `json:"join_with,omitempty"`
	Description   string   `json:"description"`
	compiled      *regexp.Regexp
}

// Compile compiles the pattern's regex
func (p *ModelLoadPattern) Compile() error {
	var err error
	p.compiled, err = regexp.Compile(p.Pattern)
	return err
}

// Regex returns the compiled regex
func (p *ModelLoadPattern) Regex() *regexp.Regexp {
	return p.compiled
}

// ModelLoadingPatternsConfig holds all model loading patterns by source
type ModelLoadingPatternsConfig struct {
	Version     string                          `json:"version"`
	Description string                          `json:"description"`
	Patterns    map[string][]ModelLoadPattern  `json:"patterns"`
}

// LoadModelLoadingPatterns loads model loading patterns from RAG
func (l *RAGLoader) LoadModelLoadingPatterns() (*ModelLoadingPatternsConfig, error) {
	var config ModelLoadingPatternsConfig
	if err := l.LoadJSON("ai-ml/patterns/model-loading-patterns.json", &config); err != nil {
		return nil, err
	}

	// Compile all patterns
	for source, patterns := range config.Patterns {
		for i := range patterns {
			if err := patterns[i].Compile(); err != nil {
				return nil, fmt.Errorf("failed to compile pattern %s/%s: %w", source, patterns[i].Name, err)
			}
		}
	}

	return &config, nil
}

// API Provider Types

// APIProviderPattern represents a pattern for detecting API provider usage
type APIProviderPattern struct {
	Pattern string `json:"pattern"`
	Prefix  string `json:"prefix"`
}

// APIProvider represents an LLM API provider
type APIProvider struct {
	Name          string               `json:"name"`
	EnvVars       []string             `json:"env_vars"`
	Packages      []string             `json:"packages"`
	ModelPatterns []APIProviderPattern `json:"model_patterns"`
	APIKeyPattern string               `json:"api_key_pattern"`
}

// APIProvidersConfig holds all API provider definitions
type APIProvidersConfig struct {
	Version          string                  `json:"version"`
	Description      string                  `json:"description"`
	Providers        map[string]APIProvider `json:"providers"`
	LangChainPatterns []struct {
		Pattern      string `json:"pattern"`
		ExtractGroup int    `json:"extract_group"`
		Source       string `json:"source"`
	} `json:"langchain_patterns"`
}

// LoadAPIProviders loads API provider definitions from RAG
func (l *RAGLoader) LoadAPIProviders() (*APIProvidersConfig, error) {
	var config APIProvidersConfig
	if err := l.LoadJSON("ai-ml/patterns/api-provider-patterns.json", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Model File Format Types

// ModelFileFormat represents a model file format with security info
type ModelFileFormat struct {
	Name        string  `json:"name"`
	Format      string  `json:"format"`
	Risk        string  `json:"risk"`
	RiskReason  string  `json:"risk_reason"`
	CWE         *string `json:"cwe"`
	Remediation *string `json:"remediation"`
}

// SafeLoadingPattern represents safe vs unsafe loading patterns
type SafeLoadingPattern struct {
	Unsafe string `json:"unsafe"`
	Safe   string `json:"safe"`
	Safest string `json:"safest,omitempty"`
}

// ModelFileFormatsConfig holds all model file format definitions
type ModelFileFormatsConfig struct {
	Version             string                        `json:"version"`
	Description         string                        `json:"description"`
	Formats             map[string]ModelFileFormat   `json:"formats"`
	SafeLoadingPatterns map[string]SafeLoadingPattern `json:"safe_loading_patterns"`
}

// LoadModelFileFormats loads model file format definitions from RAG
func (l *RAGLoader) LoadModelFileFormats() (*ModelFileFormatsConfig, error) {
	var config ModelFileFormatsConfig
	if err := l.LoadJSON("ai-ml/file-formats/model-file-formats.json", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// FindRAGPath attempts to locate the RAG directory
func FindRAGPath() string {
	candidates := []string{
		"rag",                     // Current directory
		"../rag",                  // Parent directory
		"../../rag",               // Two levels up
		"../../../rag",            // Three levels up
	}

	// Also check ZERO_HOME environment variable
	if zeroHome := os.Getenv("ZERO_HOME"); zeroHome != "" {
		candidates = append([]string{filepath.Join(zeroHome, "rag")}, candidates...)
	}

	// Check relative to executable
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates,
			filepath.Join(execDir, "rag"),
			filepath.Join(execDir, "..", "rag"),
			filepath.Join(execDir, "..", "..", "rag"),
		)
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			return absPath
		}
	}

	return ""
}

// DefaultLoader is the default RAG loader instance
var DefaultLoader = NewLoader("")

// Quick access functions using default loader

// GetModelRegistries returns all model registry definitions
func GetModelRegistries() (*ModelRegistriesConfig, error) {
	return DefaultLoader.LoadModelRegistries()
}

// GetModelLoadingPatterns returns all model loading patterns
func GetModelLoadingPatterns() (*ModelLoadingPatternsConfig, error) {
	return DefaultLoader.LoadModelLoadingPatterns()
}

// GetAPIProviders returns all API provider definitions
func GetAPIProviders() (*APIProvidersConfig, error) {
	return DefaultLoader.LoadAPIProviders()
}

// GetModelFileFormats returns all model file format definitions
func GetModelFileFormats() (*ModelFileFormatsConfig, error) {
	return DefaultLoader.LoadModelFileFormats()
}
