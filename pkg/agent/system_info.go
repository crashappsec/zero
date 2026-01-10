// Package agent provides system information queries for Zero
package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/crashappsec/zero/pkg/core/feeds"
	"github.com/crashappsec/zero/pkg/core/rag"
	"github.com/crashappsec/zero/pkg/core/rules"
)

// SystemInfo provides methods to query Zero system metadata
type SystemInfo struct {
	zeroHome   string
	ragLoader  *rag.RAGLoader
	configPath string
}

// NewSystemInfo creates a new SystemInfo instance
func NewSystemInfo(zeroHome string) *SystemInfo {
	ragPath := rag.FindRAGPath()
	if ragPath == "" {
		// Try relative to zeroHome
		ragPath = filepath.Join(filepath.Dir(zeroHome), "rag")
	}

	return &SystemInfo{
		zeroHome:   zeroHome,
		ragLoader:  rag.NewLoader(ragPath),
		configPath: filepath.Join(filepath.Dir(zeroHome), "config", "zero.config.json"),
	}
}

// GetSystemInfo dispatches to the appropriate category handler
func (s *SystemInfo) GetSystemInfo(category, filter string) (string, error) {
	switch category {
	case "rag-stats":
		return s.getRAGStats()
	case "rag-patterns":
		return s.getRAGPatterns(filter)
	case "rag-search":
		return s.getRAGSearch(filter)
	case "rag-detail":
		return s.getRAGDetail(filter)
	case "rules-status":
		return s.getRulesStatus()
	case "feeds-status":
		return s.getFeedsStatus()
	case "scanners":
		return s.getScannersInfo(filter)
	case "profiles":
		return s.getProfilesInfo()
	case "config":
		return s.getConfigInfo()
	case "agents":
		return s.getAgentsInfo(filter)
	case "versions":
		return s.getVersionsInfo()
	case "help":
		return s.getHelpInfo()
	default:
		return "", fmt.Errorf("unknown category: %s. Valid categories: rag-stats, rag-patterns, rag-search, rag-detail, rules-status, feeds-status, scanners, profiles, config, agents, versions, help", category)
	}
}

// RAGCategoryStats holds stats for a RAG category
type RAGCategoryStats struct {
	Files    int `json:"files"`
	Patterns int `json:"patterns"`
}

// RAGStatsResponse is the response for rag-stats
type RAGStatsResponse struct {
	TotalCategories int                         `json:"total_categories"`
	TotalFiles      int                         `json:"total_files"`
	TotalPatterns   int                         `json:"total_patterns"`
	RAGPath         string                      `json:"rag_path"`
	Categories      map[string]RAGCategoryStats `json:"categories"`
}

func (s *SystemInfo) getRAGStats() (string, error) {
	categories, err := s.ragLoader.ListCategories()
	if err != nil {
		return "", fmt.Errorf("listing RAG categories: %w", err)
	}

	response := RAGStatsResponse{
		RAGPath:    s.ragLoader.RAGPath(),
		Categories: make(map[string]RAGCategoryStats),
	}

	for _, cat := range categories {
		// Count files
		catPath := filepath.Join(s.ragLoader.RAGPath(), cat)
		fileCount := 0
		_ = filepath.Walk(catPath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				fileCount++
			}
			return nil
		})

		// Try to get pattern count
		patternCount, _ := s.ragLoader.GetPatternCount(cat)

		response.Categories[cat] = RAGCategoryStats{
			Files:    fileCount,
			Patterns: patternCount,
		}
		response.TotalFiles += fileCount
		response.TotalPatterns += patternCount
	}

	response.TotalCategories = len(categories)

	return toJSON(response)
}

// RAGPatternExample is an example pattern
type RAGPatternExample struct {
	Name     string `json:"name"`
	Pattern  string `json:"pattern"`
	Severity string `json:"severity"`
	Type     string `json:"type"`
}

// RAGPatternsResponse is the response for rag-patterns
type RAGPatternsResponse struct {
	Category      string              `json:"category"`
	PatternTypes  []string            `json:"pattern_types"`
	TotalPatterns int                 `json:"total_patterns"`
	Examples      []RAGPatternExample `json:"examples"`
}

func (s *SystemInfo) getRAGPatterns(filter string) (string, error) {
	if filter == "" {
		// Return list of available categories
		categories, err := s.ragLoader.ListCategories()
		if err != nil {
			return "", fmt.Errorf("listing categories: %w", err)
		}
		return toJSON(map[string]interface{}{
			"message":    "Specify a category in the 'filter' parameter",
			"categories": categories,
		})
	}

	// Check if category exists
	if !s.ragLoader.HasCategory(filter) {
		categories, _ := s.ragLoader.ListCategories()
		return "", fmt.Errorf("category '%s' not found. Available: %s", filter, strings.Join(categories, ", "))
	}

	// Load category patterns
	result, err := s.ragLoader.LoadCategory(filter)
	if err != nil {
		return "", fmt.Errorf("loading category %s: %w", filter, err)
	}

	response := RAGPatternsResponse{
		Category:      filter,
		TotalPatterns: result.TotalPatterns,
		PatternTypes:  []string{},
		Examples:      []RAGPatternExample{},
	}

	// Extract pattern types and examples
	typesSeen := make(map[string]bool)
	for _, ps := range result.PatternSets {
		if !typesSeen[ps.Technology] {
			response.PatternTypes = append(response.PatternTypes, ps.Technology)
			typesSeen[ps.Technology] = true
		}

		// Add up to 10 example patterns
		for _, p := range ps.Patterns {
			if len(response.Examples) >= 10 {
				break
			}
			response.Examples = append(response.Examples, RAGPatternExample{
				Name:     p.ID,
				Pattern:  truncateString(p.Pattern, 60),
				Severity: p.Severity,
				Type:     p.Type,
			})
		}
	}

	return toJSON(response)
}

// RAGSearchResult is a search result
type RAGSearchResult struct {
	ID         string   `json:"id"`
	Category   string   `json:"category"`
	Technology string   `json:"technology"`
	Pattern    string   `json:"pattern"`
	Type       string   `json:"type"`
	Severity   string   `json:"severity"`
	Confidence int      `json:"confidence,omitempty"`
	Languages  []string `json:"languages,omitempty"`
	Source     string   `json:"source"`
}

// RAGSearchResponse is the response for rag-search
type RAGSearchResponse struct {
	Query        string            `json:"query"`
	TotalResults int               `json:"total_results"`
	Results      []RAGSearchResult `json:"results"`
	Hint         string            `json:"hint,omitempty"`
}

// getRAGSearch searches RAG patterns with flexible filtering
// Filter format: "query:text severity:high category:cryptography language:python limit:20"
func (s *SystemInfo) getRAGSearch(filter string) (string, error) {
	if filter == "" {
		return toJSON(map[string]interface{}{
			"message": "Specify search parameters in the filter",
			"format":  "query:<text> severity:<level> category:<name> language:<lang> limit:<n>",
			"examples": []string{
				"query:password",
				"severity:critical",
				"category:cryptography",
				"query:api severity:high limit:10",
				"language:python category:secrets",
			},
			"available_severities": []string{"critical", "high", "medium", "low", "info"},
			"available_categories": s.getAvailableCategories(),
		})
	}

	// Parse filter parameters
	params := parseSearchParams(filter)
	query := params["query"]
	severityFilter := strings.ToLower(params["severity"])
	categoryFilter := params["category"]
	languageFilter := strings.ToLower(params["language"])
	limit := 20 // Default
	if l, ok := params["limit"]; ok {
		if n, err := fmt.Sscanf(l, "%d", &limit); err == nil && n > 0 {
			if limit > 100 {
				limit = 100 // Cap at 100
			}
		}
	}

	// Get all categories to search
	categories, err := s.ragLoader.ListCategories()
	if err != nil {
		return "", fmt.Errorf("listing categories: %w", err)
	}

	// Filter categories if specified
	if categoryFilter != "" {
		found := false
		for _, c := range categories {
			if strings.EqualFold(c, categoryFilter) || strings.Contains(strings.ToLower(c), strings.ToLower(categoryFilter)) {
				categories = []string{c}
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("category '%s' not found", categoryFilter)
		}
	}

	var results []RAGSearchResult

	// Search through categories
	for _, cat := range categories {
		loadResult, err := s.ragLoader.LoadCategory(cat)
		if err != nil {
			continue // Skip categories that fail to load
		}

		for _, ps := range loadResult.PatternSets {
			for _, p := range ps.Patterns {
				// Apply filters
				if severityFilter != "" && strings.ToLower(p.Severity) != severityFilter {
					continue
				}

				if languageFilter != "" {
					hasLang := false
					for _, lang := range p.Languages {
						if strings.ToLower(lang) == languageFilter {
							hasLang = true
							break
						}
					}
					if !hasLang && len(p.Languages) > 0 {
						continue
					}
				}

				// Text query match
				if query != "" {
					queryLower := strings.ToLower(query)
					matchesQuery := strings.Contains(strings.ToLower(p.ID), queryLower) ||
						strings.Contains(strings.ToLower(p.Pattern), queryLower) ||
						strings.Contains(strings.ToLower(p.Message), queryLower) ||
						strings.Contains(strings.ToLower(ps.Technology), queryLower)
					if !matchesQuery {
						continue
					}
				}

				results = append(results, RAGSearchResult{
					ID:         p.ID,
					Category:   cat,
					Technology: ps.Technology,
					Pattern:    truncateString(p.Pattern, 80),
					Type:       p.Type,
					Severity:   p.Severity,
					Confidence: p.Confidence,
					Languages:  p.Languages,
					Source:     ps.Source,
				})

				if len(results) >= limit {
					break
				}
			}
			if len(results) >= limit {
				break
			}
		}
		if len(results) >= limit {
			break
		}
	}

	response := RAGSearchResponse{
		Query:        filter,
		TotalResults: len(results),
		Results:      results,
	}

	if len(results) == 0 {
		response.Hint = "Try a broader search or different filter combination"
	} else if len(results) == limit {
		response.Hint = fmt.Sprintf("Results limited to %d. Add more filters or increase limit.", limit)
	}

	return toJSON(response)
}

func (s *SystemInfo) getAvailableCategories() []string {
	categories, err := s.ragLoader.ListCategories()
	if err != nil {
		return []string{}
	}
	return categories
}

// RAGDetailResponse is the response for rag-detail
type RAGDetailResponse struct {
	ID          string                 `json:"id"`
	Category    string                 `json:"category"`
	Technology  string                 `json:"technology"`
	Pattern     string                 `json:"pattern"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message,omitempty"`
	Description string                 `json:"description,omitempty"`
	Confidence  int                    `json:"confidence,omitempty"`
	Languages   []string               `json:"languages,omitempty"`
	CWE         string                 `json:"cwe,omitempty"`
	References  []string               `json:"references,omitempty"`
	Examples    []string               `json:"examples,omitempty"`
	Source      string                 `json:"source"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// getRAGDetail retrieves full details for a specific pattern by ID
func (s *SystemInfo) getRAGDetail(filter string) (string, error) {
	if filter == "" {
		return toJSON(map[string]interface{}{
			"message": "Specify a pattern ID to retrieve details",
			"usage":   "filter: <pattern-id>",
			"hint":    "Use rag-search to find pattern IDs first",
			"examples": []string{
				"aws-access-key-id",
				"sql-injection",
				"weak-cipher-des",
			},
		})
	}

	patternID := strings.TrimSpace(filter)

	// Search through all categories for the pattern
	categories, err := s.ragLoader.ListCategories()
	if err != nil {
		return "", fmt.Errorf("listing categories: %w", err)
	}

	for _, cat := range categories {
		loadResult, err := s.ragLoader.LoadCategory(cat)
		if err != nil {
			continue
		}

		for _, ps := range loadResult.PatternSets {
			for _, p := range ps.Patterns {
				// Match by ID (exact or partial)
				if strings.EqualFold(p.ID, patternID) ||
					strings.Contains(strings.ToLower(p.ID), strings.ToLower(patternID)) {

					// Extract CWE from message or description
					cwe := ""
					cwePattern := regexp.MustCompile(`CWE-\d+`)
					if match := cwePattern.FindString(p.Message); match != "" {
						cwe = match
					}

					response := RAGDetailResponse{
						ID:          p.ID,
						Category:    cat,
						Technology:  ps.Technology,
						Pattern:     p.Pattern,
						Type:        p.Type,
						Severity:    p.Severity,
						Message:     p.Message,
						Confidence:  p.Confidence,
						Languages:   p.Languages,
						CWE:         cwe,
						Source:      ps.Source,
						Metadata:    p.Metadata,
					}

					return toJSON(response)
				}
			}
		}
	}

	return "", fmt.Errorf("pattern '%s' not found. Use rag-search to find valid pattern IDs", patternID)
}

// parseSearchParams parses "key:value" pairs from filter string
func parseSearchParams(filter string) map[string]string {
	params := make(map[string]string)

	// Handle simple text query without key
	if !strings.Contains(filter, ":") {
		params["query"] = filter
		return params
	}

	// Parse key:value pairs
	parts := strings.Fields(filter)
	for _, part := range parts {
		if idx := strings.Index(part, ":"); idx > 0 {
			key := part[:idx]
			value := part[idx+1:]
			params[key] = value
		} else {
			// Treat as part of query
			if existing, ok := params["query"]; ok {
				params["query"] = existing + " " + part
			} else {
				params["query"] = part
			}
		}
	}

	return params
}

// RulesStatusResponse is the response for rules-status
type RulesStatusResponse struct {
	Generated *RuleSourceStatus `json:"generated,omitempty"`
	Community *RuleSourceStatus `json:"community,omitempty"`
}

// RuleSourceStatus has status for a rule source
type RuleSourceStatus struct {
	LastGenerate string   `json:"last_generate,omitempty"`
	RuleCount    int      `json:"rule_count"`
	Categories   []string `json:"categories,omitempty"`
	SourceHash   string   `json:"source_hash,omitempty"`
	Error        string   `json:"error,omitempty"`
}

func (s *SystemInfo) getRulesStatus() (string, error) {
	manager := rules.NewManager(s.zeroHome)
	statuses := manager.GetStatus()

	response := RulesStatusResponse{}

	if status, ok := statuses["community"]; ok {
		response.Community = &RuleSourceStatus{
			LastGenerate: status.LastGenerate.Format(time.RFC3339),
			RuleCount:    status.RuleCount,
			Error:        status.Error,
		}
	}

	// If no status exists, return info about how to sync
	if response.Community == nil {
		return toJSON(map[string]interface{}{
			"message":  "No rules synced yet. Run 'zero feeds semgrep' to sync community SAST rules.",
			"commands": []string{"zero feeds semgrep"},
		})
	}

	return toJSON(response)
}

// FeedStatusInfo has status for a feed
type FeedStatusInfo struct {
	LastSync    string `json:"last_sync,omitempty"`
	LastSuccess string `json:"last_success,omitempty"`
	Freshness   string `json:"freshness"`
	ItemCount   int    `json:"item_count,omitempty"`
	Error       string `json:"error,omitempty"`
	Note        string `json:"note,omitempty"`
}

// FeedsStatusResponse is the response for feeds-status
type FeedsStatusResponse struct {
	Feeds map[string]FeedStatusInfo `json:"feeds"`
	Note  string                    `json:"note"`
}

func (s *SystemInfo) getFeedsStatus() (string, error) {
	manager := feeds.NewManager(s.zeroHome)
	if err := manager.LoadStatus(); err != nil {
		// No status file yet
		return toJSON(FeedsStatusResponse{
			Feeds: map[string]FeedStatusInfo{
				"semgrep-rules": {Freshness: "not synced", ItemCount: 0},
			},
			Note: "Run 'zero feeds semgrep' to sync SAST rules. Vulnerability data (OSV) is queried live during scans.",
		})
	}

	response := FeedsStatusResponse{
		Feeds: make(map[string]FeedStatusInfo),
		Note:  "Vulnerability data (OSV.dev) is queried live during scans, not cached.",
	}

	for _, status := range manager.GetAllStatus() {
		freshness := calculateFreshness(status.LastSuccess)
		response.Feeds[string(status.Type)] = FeedStatusInfo{
			LastSync:    status.LastSync.Format(time.RFC3339),
			LastSuccess: status.LastSuccess.Format(time.RFC3339),
			Freshness:   freshness,
			ItemCount:   status.ItemCount,
			Error:       status.LastError,
		}
	}

	// Add note about live feeds
	response.Feeds["osv"] = FeedStatusInfo{
		Freshness: "live",
		Note:      "Queried in real-time during scans via api.osv.dev",
	}

	return toJSON(response)
}

// ScannerInfo describes a scanner
type ScannerInfo struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Features     []string `json:"features"`
	FeatureCount int      `json:"feature_count"`
	OutputFile   string   `json:"output_file"`
}

// ScannersResponse is the response for scanners
type ScannersResponse struct {
	Scanners     []ScannerInfo `json:"scanners"`
	TotalCount   int           `json:"total_count"`
	TotalFeatures int          `json:"total_features"`
}

func (s *SystemInfo) getScannersInfo(filter string) (string, error) {
	// Scanner definitions (static - matches CLAUDE.md documentation)
	scannerDefs := []ScannerInfo{
		{
			Name:        "code-packages",
			Description: "SBOM generation + package/dependency analysis",
			Features:    []string{"generation", "integrity", "vulns", "health", "licenses", "malcontent", "confusion", "typosquats", "deprecations", "duplicates", "reachability", "provenance", "bundle", "recommendations"},
			OutputFile:  "code-packages.json",
		},
		{
			Name:        "code-security",
			Description: "Security-focused code analysis + cryptography",
			Features:    []string{"vulns", "secrets", "api", "ciphers", "keys", "random", "tls", "certificates"},
			OutputFile:  "code-security.json",
		},
		{
			Name:        "code-quality",
			Description: "Code quality metrics and technical debt",
			Features:    []string{"tech_debt", "complexity", "test_coverage", "documentation"},
			OutputFile:  "code-quality.json",
		},
		{
			Name:        "devops",
			Description: "DevOps and CI/CD security analysis",
			Features:    []string{"iac", "containers", "github_actions", "dora", "git"},
			OutputFile:  "devops.json",
		},
		{
			Name:        "technology-identification",
			Description: "Technology detection and ML-BOM generation",
			Features:    []string{"detection", "models", "frameworks", "datasets", "ai_security", "ai_governance", "infrastructure"},
			OutputFile:  "technology-identification.json",
		},
		{
			Name:        "code-ownership",
			Description: "Code ownership and contributor analysis",
			Features:    []string{"contributors", "bus_factor", "codeowners", "orphans", "churn", "patterns"},
			OutputFile:  "code-ownership.json",
		},
		{
			Name:        "developer-experience",
			Description: "Developer experience and onboarding analysis",
			Features:    []string{"onboarding", "sprawl", "workflow"},
			OutputFile:  "developer-experience.json",
		},
	}

	// If filter specified, return only that scanner
	if filter != "" {
		for _, scanner := range scannerDefs {
			if scanner.Name == filter {
				scanner.FeatureCount = len(scanner.Features)
				return toJSON(scanner)
			}
		}
		names := make([]string, len(scannerDefs))
		for i, s := range scannerDefs {
			names[i] = s.Name
		}
		return "", fmt.Errorf("scanner '%s' not found. Available: %s", filter, strings.Join(names, ", "))
	}

	// Return all scanners
	response := ScannersResponse{
		Scanners: make([]ScannerInfo, 0, len(scannerDefs)),
	}
	for _, scanner := range scannerDefs {
		scanner.FeatureCount = len(scanner.Features)
		response.Scanners = append(response.Scanners, scanner)
		response.TotalFeatures += scanner.FeatureCount
	}
	response.TotalCount = len(scannerDefs)

	return toJSON(response)
}

// ProfileInfo describes a scan profile
type ProfileInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Scanners    []string `json:"scanners"`
}

// ProfilesResponse is the response for profiles
type ProfilesResponse struct {
	Profiles map[string]ProfileInfo `json:"profiles"`
	Default  string                 `json:"default"`
}

func (s *SystemInfo) getProfilesInfo() (string, error) {
	// Try to load from config file
	config, err := s.loadConfig()
	if err == nil && config != nil {
		if profiles, ok := config["profiles"].(map[string]interface{}); ok {
			response := ProfilesResponse{
				Profiles: make(map[string]ProfileInfo),
				Default:  "all-quick",
			}

			if settings, ok := config["settings"].(map[string]interface{}); ok {
				if def, ok := settings["default_profile"].(string); ok {
					response.Default = def
				}
			}

			for name, p := range profiles {
				if profile, ok := p.(map[string]interface{}); ok {
					info := ProfileInfo{
						Name:        getString(profile, "name"),
						Description: getString(profile, "description"),
					}
					if scanners, ok := profile["scanners"].([]interface{}); ok {
						for _, s := range scanners {
							if str, ok := s.(string); ok {
								info.Scanners = append(info.Scanners, str)
							}
						}
					}
					response.Profiles[name] = info
				}
			}

			return toJSON(response)
		}
	}

	// Return default profiles if config not found
	return toJSON(ProfilesResponse{
		Default: "all-quick",
		Profiles: map[string]ProfileInfo{
			"all-quick":      {Name: "All Quick", Description: "All scanners with fast defaults", Scanners: []string{"all"}},
			"all-complete":   {Name: "All Complete", Description: "All scanners with all features", Scanners: []string{"all"}},
			"code-packages":  {Name: "Code Packages", Description: "SBOM and package analysis", Scanners: []string{"code-packages"}},
			"code-security":  {Name: "Code Security", Description: "SAST, secrets, crypto", Scanners: []string{"code-security"}},
			"code-quality":   {Name: "Code Quality", Description: "Tech debt, complexity", Scanners: []string{"code-quality"}},
			"devops":         {Name: "DevOps", Description: "IaC, containers, CI/CD", Scanners: []string{"devops"}},
		},
	})
}

// ConfigResponse is the response for config
type ConfigResponse struct {
	ZeroHome         string `json:"zero_home"`
	DefaultProfile   string `json:"default_profile"`
	ParallelRepos    int    `json:"parallel_repos"`
	ParallelScanners int    `json:"parallel_scanners"`
	TimeoutSeconds   int    `json:"timeout_seconds"`
	CacheTTLHours    int    `json:"cache_ttl_hours"`
	ConfigPath       string `json:"config_path"`
}

func (s *SystemInfo) getConfigInfo() (string, error) {
	response := ConfigResponse{
		ZeroHome:         s.zeroHome,
		ConfigPath:       s.configPath,
		DefaultProfile:   "all-quick",
		ParallelRepos:    8,
		ParallelScanners: 4,
		TimeoutSeconds:   300,
		CacheTTLHours:    24,
	}

	// Try to load actual config
	config, err := s.loadConfig()
	if err == nil && config != nil {
		if settings, ok := config["settings"].(map[string]interface{}); ok {
			if v, ok := settings["default_profile"].(string); ok {
				response.DefaultProfile = v
			}
			if v, ok := settings["parallel_repos"].(float64); ok {
				response.ParallelRepos = int(v)
			}
			if v, ok := settings["parallel_scanners"].(float64); ok {
				response.ParallelScanners = int(v)
			}
			if v, ok := settings["scanner_timeout_seconds"].(float64); ok {
				response.TimeoutSeconds = int(v)
			}
			if v, ok := settings["cache_ttl_hours"].(float64); ok {
				response.CacheTTLHours = int(v)
			}
		}
	}

	return toJSON(response)
}

// AgentInfo describes a specialist agent
type AgentInfo struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Domain         string   `json:"domain"`
	PrimaryScanner string   `json:"primary_scanner"`
	CanDelegateTo  []string `json:"can_delegate_to,omitempty"`
}

// AgentsResponse is the response for agents
type AgentsResponse struct {
	Orchestrator string      `json:"orchestrator"`
	TotalAgents  int         `json:"total_agents"`
	Agents       []AgentInfo `json:"agents"`
}

func (s *SystemInfo) getAgentsInfo(filter string) (string, error) {
	// Agent definitions (matches CLAUDE.md documentation)
	agentDefs := []AgentInfo{
		{ID: "zero", Name: "Zero Cool", Domain: "Orchestrator", PrimaryScanner: "all", CanDelegateTo: []string{"all specialists"}},
		{ID: "cereal", Name: "Cereal Killer", Domain: "Supply Chain Security", PrimaryScanner: "code-packages", CanDelegateTo: []string{"phreak", "razor", "plague", "nikon", "gill"}},
		{ID: "razor", Name: "Razor", Domain: "Code Security", PrimaryScanner: "code-security", CanDelegateTo: []string{"cereal", "blade", "nikon", "dade", "gill"}},
		{ID: "blade", Name: "Blade", Domain: "Compliance", PrimaryScanner: "code-packages, code-security", CanDelegateTo: []string{"cereal", "razor", "phreak", "gill"}},
		{ID: "phreak", Name: "Phantom Phreak", Domain: "Legal", PrimaryScanner: "code-packages (licenses)"},
		{ID: "acid", Name: "Acid Burn", Domain: "Frontend", PrimaryScanner: "code-security, code-quality"},
		{ID: "dade", Name: "Dade Murphy", Domain: "Backend", PrimaryScanner: "code-security (api)"},
		{ID: "nikon", Name: "Lord Nikon", Domain: "Architecture", PrimaryScanner: "technology-identification", CanDelegateTo: []string{"all technical domains"}},
		{ID: "joey", Name: "Joey", Domain: "Build/CI", PrimaryScanner: "devops (github_actions)"},
		{ID: "plague", Name: "The Plague", Domain: "DevOps", PrimaryScanner: "devops"},
		{ID: "gibson", Name: "The Gibson", Domain: "Engineering Metrics", PrimaryScanner: "devops, code-ownership"},
		{ID: "gill", Name: "Gill Bates", Domain: "Cryptography", PrimaryScanner: "code-security (crypto)"},
		{ID: "hal", Name: "Hal", Domain: "AI/ML Security", PrimaryScanner: "technology-identification"},
	}

	// If filter specified, return only that agent
	if filter != "" {
		for _, agent := range agentDefs {
			if agent.ID == filter {
				return toJSON(agent)
			}
		}
		ids := make([]string, len(agentDefs))
		for i, a := range agentDefs {
			ids[i] = a.ID
		}
		return "", fmt.Errorf("agent '%s' not found. Available: %s", filter, strings.Join(ids, ", "))
	}

	// Return all agents
	return toJSON(AgentsResponse{
		Orchestrator: "zero",
		TotalAgents:  len(agentDefs),
		Agents:       agentDefs,
	})
}

// VersionsResponse is the response for versions
type VersionsResponse struct {
	Zero     VersionInfo            `json:"zero"`
	Scanners map[string]string      `json:"scanners"`
	API      map[string]string      `json:"api"`
}

// VersionInfo describes version info
type VersionInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date,omitempty"`
}

func (s *SystemInfo) getVersionsInfo() (string, error) {
	return toJSON(VersionsResponse{
		Zero: VersionInfo{
			Version:   "0.4.0",
			BuildDate: time.Now().Format("2006-01-02"),
		},
		Scanners: map[string]string{
			"semgrep":     "1.56.0+",
			"cdxgen":      "10.x",
			"osv-scanner": "1.x",
			"checkov":     "3.x",
			"trivy":       "0.48.x",
		},
		API: map[string]string{
			"claude_model": "claude-sonnet-4-20250514",
			"api_version":  "2023-06-01",
		},
	})
}

// HelpResponse is the response for help
type HelpResponse struct {
	Capabilities     []string `json:"capabilities"`
	ExampleQuestions []string `json:"example_questions"`
	Categories       []string `json:"categories"`
}

func (s *SystemInfo) getHelpInfo() (string, error) {
	return toJSON(HelpResponse{
		Capabilities: []string{
			"Answer questions about Zero's detection rules and patterns",
			"Search RAG patterns by name, severity, category, or language",
			"Explain what scanners and features are available",
			"Show status of security feeds and data freshness",
			"List available specialist agents and their expertise",
			"Describe scan profiles and configuration options",
			"Provide version information for Zero and its dependencies",
		},
		ExampleQuestions: []string{
			"How many rules do we have for secrets detection?",
			"What scanners are available and what do they check?",
			"When were the vulnerability feeds last updated?",
			"Which agent should I use for license compliance?",
			"What detection patterns exist for AWS credentials?",
			"Search for critical severity patterns",
			"Find patterns related to cryptography",
			"What scan profiles are available?",
		},
		Categories: []string{
			"rag-stats - Pattern counts by RAG category",
			"rag-patterns - List patterns in a category (use filter)",
			"rag-search - Search patterns (filter: query:text severity:level category:name language:lang limit:n)",
			"rag-detail - Get full pattern details by ID (filter: pattern-id)",
			"rules-status - Generated and community rule status",
			"feeds-status - Feed synchronization status",
			"scanners - Scanner inventory with features",
			"profiles - Available scan profiles",
			"config - Active configuration summary",
			"agents - Available specialist agents",
			"versions - Zero and scanner versions",
			"help - This help information",
		},
	})
}

// Helper functions

func (s *SystemInfo) loadConfig() (map[string]interface{}, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func calculateFreshness(lastSuccess time.Time) string {
	if lastSuccess.IsZero() {
		return "never synced"
	}

	age := time.Since(lastSuccess)
	switch {
	case age < 24*time.Hour:
		return "fresh"
	case age < 7*24*time.Hour:
		return "stale"
	case age < 30*24*time.Hour:
		return "very stale"
	default:
		return "expired"
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func toJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
