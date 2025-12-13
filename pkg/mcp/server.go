// Package mcp provides an MCP server for Zero analysis data
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server implements the Zero MCP server
type Server struct {
	zeroHome string
	server   *mcp.Server
}

// Project represents a hydrated project
type Project struct {
	ID             string   `json:"id"`
	Owner          string   `json:"owner"`
	Repo           string   `json:"repo"`
	Path           string   `json:"path"`
	AnalysisPath   string   `json:"analysis_path"`
	AvailableScans []string `json:"available_scans"`
}

// NewServer creates a new MCP server
func NewServer(zeroHome string) *Server {
	if zeroHome == "" {
		zeroHome = filepath.Join(os.Getenv("HOME"), ".zero")
	}

	s := &Server{
		zeroHome: zeroHome,
	}

	s.server = mcp.NewServer(&mcp.Implementation{
		Name:    "zero",
		Version: "1.0.0",
	}, nil)

	// Register tools
	s.registerTools()

	return s
}

// Run starts the MCP server on stdio
func (s *Server) Run(ctx context.Context) error {
	return s.server.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) registerTools() {
	// list_projects tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all hydrated projects with their available analyses",
	}, s.handleListProjects)

	// get_project_summary tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_project_summary",
		Description: "Get a summary of a project including available analyses and basic stats",
	}, s.handleGetProjectSummary)

	// get_vulnerabilities tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_vulnerabilities",
		Description: "Get known vulnerabilities (CVEs) for a project's dependencies",
	}, s.handleGetVulnerabilities)

	// get_malcontent tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_malcontent",
		Description: "Get malcontent (malware/suspicious behavior) findings for a project",
	}, s.handleGetMalcontent)

	// get_technologies tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_technologies",
		Description: "Get detected technologies, frameworks, and libraries used in a project",
	}, s.handleGetTechnologies)

	// get_package_health tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_package_health",
		Description: "Get package health scores and dependency analysis for a project",
	}, s.handleGetPackageHealth)

	// get_licenses tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_licenses",
		Description: "Get license information for a project's dependencies",
	}, s.handleGetLicenses)

	// get_secrets tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_secrets",
		Description: "Get detected secrets/credentials in the codebase",
	}, s.handleGetSecrets)

	// get_crypto_issues tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_crypto_issues",
		Description: "Get cryptographic security issues (weak ciphers, hardcoded keys, TLS issues)",
	}, s.handleGetCryptoIssues)

	// get_analysis_raw tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_analysis_raw",
		Description: "Get raw analysis JSON for any analysis type",
	}, s.handleGetAnalysisRaw)

	// search_findings tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "search_findings",
		Description: "Search across all findings for a pattern or keyword",
	}, s.handleSearchFindings)
}

// Input types for tools

// ListProjectsInput parameters for list_projects tool
type ListProjectsInput struct {
	// Filter projects by owner/org name
	Owner string `json:"owner,omitempty"`
}

// ProjectInput common parameter for tools that take a project
type ProjectInput struct {
	// Project ID in owner/repo format
	Project string `json:"project"`
}

// VulnerabilitiesInput parameters for get_vulnerabilities tool
type VulnerabilitiesInput struct {
	// Project ID in owner/repo format
	Project string `json:"project"`
	// Filter by severity (critical/high/medium/low)
	Severity string `json:"severity,omitempty"`
}

// MalcontentInput parameters for get_malcontent tool
type MalcontentInput struct {
	// Project ID in owner/repo format
	Project string `json:"project"`
	// Minimum risk level to include
	MinRisk string `json:"min_risk,omitempty"`
	// Maximum number of findings to return
	Limit int `json:"limit,omitempty"`
}

// AnalysisRawInput parameters for get_analysis_raw tool
type AnalysisRawInput struct {
	// Project ID in owner/repo format
	Project string `json:"project"`
	// Type of analysis (e.g. package-vulns, package-health)
	AnalysisType string `json:"analysis_type"`
}

// SearchInput parameters for search_findings tool
type SearchInput struct {
	// Search query string
	Query string `json:"query"`
	// Limit search to specific project
	Project string `json:"project,omitempty"`
	// Limit search to specific analysis type
	Type string `json:"type,omitempty"`
}

// Output types for tools
type TextOutput struct {
	Text string `json:"text"`
}

// Tool handlers - return (*CallToolResult, Output, error)
// The SDK populates Content from the Output type automatically
func (s *Server) handleListProjects(ctx context.Context, req *mcp.CallToolRequest, input ListProjectsInput) (*mcp.CallToolResult, TextOutput, error) {
	projects, err := s.getProjects()
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("failed to list projects: %w", err)
	}

	// Filter by owner if specified
	if input.Owner != "" {
		var filtered []Project
		for _, p := range projects {
			if p.Owner == input.Owner {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}

	data, _ := json.MarshalIndent(projects, "", "  ")
	return nil, TextOutput{Text: string(data)}, nil
}

func (s *Server) handleGetProjectSummary(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	projects, _ := s.getProjects()
	var found *Project
	for _, p := range projects {
		if p.ID == input.Project {
			found = &p
			break
		}
	}

	if found == nil {
		return nil, TextOutput{}, fmt.Errorf("project '%s' not found", input.Project)
	}

	summary := map[string]interface{}{
		"project":         found.ID,
		"available_scans": found.AvailableScans,
	}

	// Add vuln stats if available
	if contains(found.AvailableScans, "package-vulns") {
		if data, err := s.readAnalysis(input.Project, "package-vulns"); err == nil {
			if summ, ok := data["summary"].(map[string]interface{}); ok {
				summary["vulnerabilities"] = summ
			}
		}
	}

	// Add package stats if available
	if contains(found.AvailableScans, "package-sbom") {
		if data, err := s.readAnalysis(input.Project, "package-sbom"); err == nil {
			if summ, ok := data["summary"].(map[string]interface{}); ok {
				summary["packages"] = summ
			}
		}
	}

	result, _ := json.MarshalIndent(summary, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetVulnerabilities(ctx context.Context, req *mcp.CallToolRequest, input VulnerabilitiesInput) (*mcp.CallToolResult, TextOutput, error) {
	data, err := s.readAnalysis(input.Project, "package-vulns")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no vulnerability data for '%s'", input.Project)
	}

	// Filter by severity if specified
	if input.Severity != "" {
		if findings, ok := data["findings"].([]interface{}); ok {
			var filtered []interface{}
			for _, f := range findings {
				if fm, ok := f.(map[string]interface{}); ok {
					if sev, _ := fm["severity"].(string); strings.EqualFold(sev, input.Severity) {
						filtered = append(filtered, f)
					}
				}
			}
			data["findings"] = filtered
		}
	}

	result, _ := json.MarshalIndent(data, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetMalcontent(ctx context.Context, req *mcp.CallToolRequest, input MalcontentInput) (*mcp.CallToolResult, TextOutput, error) {
	data, err := s.readAnalysis(input.Project, "package-malcontent")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no malcontent data for '%s'", input.Project)
	}

	result, _ := json.MarshalIndent(data, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetTechnologies(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	data, err := s.readAnalysis(input.Project, "tech-discovery")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no technology data for '%s'", input.Project)
	}

	result, _ := json.MarshalIndent(data, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetPackageHealth(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	data, err := s.readAnalysis(input.Project, "package-health")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no package health data for '%s'", input.Project)
	}

	result, _ := json.MarshalIndent(data, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetLicenses(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	data, err := s.readAnalysis(input.Project, "licenses")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no license data for '%s'", input.Project)
	}

	result, _ := json.MarshalIndent(data, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetSecrets(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	data, err := s.readAnalysis(input.Project, "code-secrets")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no secrets data for '%s'", input.Project)
	}

	result, _ := json.MarshalIndent(data, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetCryptoIssues(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	combined := make(map[string]interface{})
	combined["project"] = input.Project

	// Collect all crypto scanner results
	if data, err := s.readAnalysis(input.Project, "crypto-ciphers"); err == nil {
		combined["weak_ciphers"] = data
	}
	if data, err := s.readAnalysis(input.Project, "crypto-keys"); err == nil {
		combined["hardcoded_keys"] = data
	}
	if data, err := s.readAnalysis(input.Project, "crypto-tls"); err == nil {
		combined["tls_issues"] = data
	}
	if data, err := s.readAnalysis(input.Project, "crypto-random"); err == nil {
		combined["weak_random"] = data
	}

	if len(combined) == 1 {
		return nil, TextOutput{}, fmt.Errorf("no crypto data for '%s'", input.Project)
	}

	result, _ := json.MarshalIndent(combined, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetAnalysisRaw(ctx context.Context, req *mcp.CallToolRequest, input AnalysisRawInput) (*mcp.CallToolResult, TextOutput, error) {
	data, err := s.readAnalysis(input.Project, input.AnalysisType)
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no %s data for '%s'", input.AnalysisType, input.Project)
	}

	result, _ := json.MarshalIndent(data, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleSearchFindings(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, TextOutput, error) {
	query := strings.ToLower(input.Query)
	var results []map[string]interface{}

	projects, _ := s.getProjects()
	if input.Project != "" {
		// Filter to specific project
		var filtered []Project
		for _, p := range projects {
			if p.ID == input.Project {
				filtered = append(filtered, p)
				break
			}
		}
		projects = filtered
	}

	searchTypes := []string{"package-vulns", "package-malcontent", "code-secrets"}
	if input.Type != "" && input.Type != "all" {
		searchTypes = []string{input.Type}
	}

	for _, proj := range projects {
		for _, scanType := range searchTypes {
			if !contains(proj.AvailableScans, scanType) {
				continue
			}

			data, err := s.readAnalysis(proj.ID, scanType)
			if err != nil {
				continue
			}

			// Search in findings
			dataStr := strings.ToLower(fmt.Sprintf("%v", data))
			if strings.Contains(dataStr, query) {
				results = append(results, map[string]interface{}{
					"project": proj.ID,
					"type":    scanType,
					"data":    data,
				})
			}
		}
	}

	output := map[string]interface{}{
		"query":   input.Query,
		"results": results,
		"count":   len(results),
	}

	result, _ := json.MarshalIndent(output, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

// Helper functions
func (s *Server) getProjects() ([]Project, error) {
	var projects []Project
	reposDir := filepath.Join(s.zeroHome, "repos")

	orgs, err := os.ReadDir(reposDir)
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
		if !org.IsDir() {
			continue
		}

		orgPath := filepath.Join(reposDir, org.Name())
		repos, err := os.ReadDir(orgPath)
		if err != nil {
			continue
		}

		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}

			repoPath := filepath.Join(orgPath, repo.Name())
			analysisPath := filepath.Join(repoPath, "analysis")

			// Get available analyses
			var availableScans []string
			if files, err := os.ReadDir(analysisPath); err == nil {
				for _, f := range files {
					if strings.HasSuffix(f.Name(), ".json") {
						availableScans = append(availableScans, strings.TrimSuffix(f.Name(), ".json"))
					}
				}
			}

			sort.Strings(availableScans)

			projects = append(projects, Project{
				ID:             fmt.Sprintf("%s/%s", org.Name(), repo.Name()),
				Owner:          org.Name(),
				Repo:           repo.Name(),
				Path:           repoPath,
				AnalysisPath:   analysisPath,
				AvailableScans: availableScans,
			})
		}
	}

	return projects, nil
}

func (s *Server) readAnalysis(projectID, analysisType string) (map[string]interface{}, error) {
	path := filepath.Join(s.zeroHome, "repos", projectID, "analysis", analysisType+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
