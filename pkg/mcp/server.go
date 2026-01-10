// Package mcp provides an MCP server for Zero analysis data
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Default limits - can be overridden via ServerConfig
var (
	// DefaultMaxOutputSize is the default maximum size of JSON output (1MB)
	DefaultMaxOutputSize = 1024 * 1024
	// DefaultMaxFindingsPerCategory is the default limit for findings per category
	DefaultMaxFindingsPerCategory = 500
	// DefaultMaxFileSize is the default maximum analysis file size (50MB)
	DefaultMaxFileSize = int64(50 * 1024 * 1024)
)

// ServerConfig holds configurable limits for the MCP server
type ServerConfig struct {
	// MaxOutputSize limits JSON output size in bytes (default: 1MB)
	MaxOutputSize int
	// MaxFindingsPerCategory limits findings returned per category (default: 500)
	MaxFindingsPerCategory int
	// MaxFileSize limits analysis file size in bytes (default: 50MB)
	MaxFileSize int64
}

// DefaultServerConfig returns the default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		MaxOutputSize:          DefaultMaxOutputSize,
		MaxFindingsPerCategory: DefaultMaxFindingsPerCategory,
		MaxFileSize:            DefaultMaxFileSize,
	}
}

// validProjectPattern matches valid project IDs (owner/repo format)
// Prevents path traversal attacks
var validProjectPattern = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9_.]*[/][a-zA-Z0-9][-a-zA-Z0-9_.]*$`)

// Server implements the Zero MCP server
type Server struct {
	zeroHome string
	server   *mcp.Server
	config   ServerConfig
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

// NewServer creates a new MCP server with default configuration
func NewServer(zeroHome string) *Server {
	return NewServerWithConfig(zeroHome, DefaultServerConfig())
}

// NewServerWithConfig creates a new MCP server with custom configuration
func NewServerWithConfig(zeroHome string, config ServerConfig) *Server {
	if zeroHome == "" {
		zeroHome = filepath.Join(os.Getenv("HOME"), ".zero")
	}

	s := &Server{
		zeroHome: zeroHome,
		config:   config,
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

	// get_devops_findings tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_devops_findings",
		Description: "Get DevOps findings including IaC issues, container security, and GitHub Actions analysis",
	}, s.handleGetDevOpsFindings)

	// get_code_quality tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_code_quality",
		Description: "Get code quality metrics including tech debt, complexity, and test coverage",
	}, s.handleGetCodeQuality)

	// get_ownership_metrics tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_ownership_metrics",
		Description: "Get code ownership metrics including bus factor, top contributors, and orphaned code",
	}, s.handleGetOwnershipMetrics)

	// get_dora_metrics tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_dora_metrics",
		Description: "Get DORA metrics including deployment frequency, lead time, MTTR, and change failure rate",
	}, s.handleGetDoraMetrics)

	// get_devx_analysis tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_devx_analysis",
		Description: "Get developer experience analysis including onboarding friction, tool sprawl, and workflow issues",
	}, s.handleGetDevXAnalysis)
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
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, TextOutput{}, ctx.Err()
	default:
	}

	projects, err := s.getProjects()
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("failed to list projects: %w", err)
	}

	// Filter by owner if specified
	if input.Owner != "" {
		// Validate owner (simple alphanumeric)
		if len(input.Owner) > 100 {
			return nil, TextOutput{}, fmt.Errorf("owner filter too long")
		}
		var filtered []Project
		for _, p := range projects {
			if strings.EqualFold(p.Owner, input.Owner) {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}

	result := map[string]interface{}{
		"projects": projects,
		"count":    len(projects),
	}

	data, err := safeJSONMarshal(result)
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Text: string(data)}, nil
}

func (s *Server) handleGetProjectSummary(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, TextOutput{}, ctx.Err()
	default:
	}

	// Validate project ID
	if err := validateProjectID(input.Project); err != nil {
		return nil, TextOutput{}, err
	}

	projects, err := s.getProjects()
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("failed to list projects: %w", err)
	}

	var found *Project
	for _, p := range projects {
		if p.ID == input.Project {
			found = &p
			break
		}
	}

	if found == nil {
		return nil, TextOutput{}, fmt.Errorf("project '%s' not found. Run 'zero hydrate %s' first", input.Project, input.Project)
	}

	summary := map[string]interface{}{
		"project":           found.ID,
		"available_analyses": found.AvailableScans,
	}

	// Add package/vuln stats from code-packages analyzer (v4.0)
	if contains(found.AvailableScans, "code-packages") {
		if data, err := s.readAnalysis(input.Project, "code-packages"); err == nil {
			if summ, ok := data["summary"].(map[string]interface{}); ok {
				summary["packages"] = summ
			}
		}
	}

	// Add security stats from code-security analyzer (v4.0)
	if contains(found.AvailableScans, "code-security") {
		if data, err := s.readAnalysis(input.Project, "code-security"); err == nil {
			if summ, ok := data["summary"].(map[string]interface{}); ok {
				summary["security"] = summ
			}
		}
	}

	// Add technology stats (v4.0)
	if contains(found.AvailableScans, "technology-identification") {
		if data, err := s.readAnalysis(input.Project, "technology-identification"); err == nil {
			if summ, ok := data["summary"].(map[string]interface{}); ok {
				summary["technologies"] = summ
			}
		}
	}

	// Add quality stats
	if contains(found.AvailableScans, "code-quality") {
		if data, err := s.readAnalysis(input.Project, "code-quality"); err == nil {
			if summ, ok := data["summary"].(map[string]interface{}); ok {
				summary["quality"] = summ
			}
		}
	}

	// Add ownership stats
	if contains(found.AvailableScans, "code-ownership") {
		if data, err := s.readAnalysis(input.Project, "code-ownership"); err == nil {
			if summ, ok := data["summary"].(map[string]interface{}); ok {
				summary["ownership"] = summ
			}
		}
	}

	// Add devops stats
	if contains(found.AvailableScans, "devops") {
		if data, err := s.readAnalysis(input.Project, "devops"); err == nil {
			if summ, ok := data["summary"].(map[string]interface{}); ok {
				summary["devops"] = summ
			}
		}
	}

	result, err := safeJSONMarshal(summary)
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetVulnerabilities(ctx context.Context, req *mcp.CallToolRequest, input VulnerabilitiesInput) (*mcp.CallToolResult, TextOutput, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, TextOutput{}, ctx.Err()
	default:
	}

	// v4.0: Vulnerabilities are in code-packages scanner under findings.vulns
	data, err := s.readAnalysis(input.Project, "code-packages")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("vulnerability data unavailable: %w", err)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	var warnings []string

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if vulns, ok := findings["vulns"].([]interface{}); ok {
			// Filter by severity if specified
			if input.Severity != "" {
				// Validate severity value
				validSeverities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
				if !validSeverities[strings.ToLower(input.Severity)] {
					return nil, TextOutput{}, fmt.Errorf("invalid severity: must be critical, high, medium, or low")
				}

				var filtered []interface{}
				for _, v := range vulns {
					if vm, ok := v.(map[string]interface{}); ok {
						if sev, _ := vm["severity"].(string); strings.EqualFold(sev, input.Severity) {
							filtered = append(filtered, v)
						}
					}
				}
				// Apply limit with warning
				limited, limitWarning := s.limitFindingsWithWarning(filtered, "vulnerabilities")
				if limitWarning != nil {
					result["_findings_warning"] = limitWarning
					warnings = append(warnings, limitWarning["_warning"].(string))
				}
				result["vulnerabilities"] = limited
				result["count"] = len(limited)
				result["total_matching"] = len(filtered)
			} else {
				// Apply limit with warning
				limited, limitWarning := s.limitFindingsWithWarning(vulns, "vulnerabilities")
				if limitWarning != nil {
					result["_findings_warning"] = limitWarning
					warnings = append(warnings, limitWarning["_warning"].(string))
				}
				result["vulnerabilities"] = limited
				result["count"] = len(limited)
				result["total"] = len(vulns)
			}
		} else {
			result["vulnerabilities"] = []interface{}{}
			result["count"] = 0
		}
	} else {
		result["vulnerabilities"] = []interface{}{}
		result["count"] = 0
		result["_note"] = "No findings section in analysis data"
	}

	if len(warnings) > 0 {
		result["_warnings"] = warnings
	}

	output, outputWarning, err := s.safeJSONMarshal(result)
	if err != nil {
		return nil, TextOutput{}, err
	}
	if outputWarning != "" {
		// Warning already in output
	}
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetMalcontent(ctx context.Context, req *mcp.CallToolRequest, input MalcontentInput) (*mcp.CallToolResult, TextOutput, error) {
	// v4.0: Malcontent is in code-packages scanner under findings.malcontent
	data, err := s.readAnalysis(input.Project, "code-packages")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no malcontent data for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if malcontent, ok := findings["malcontent"]; ok {
			result["malcontent"] = malcontent
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetTechnologies(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// v4.0: Technology detection is in technology-identification scanner
	data, err := s.readAnalysis(input.Project, "technology-identification")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no technology data for '%s'", input.Project)
	}

	result, _ := json.MarshalIndent(data, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetPackageHealth(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// v4.0: Package health is in code-packages scanner under findings.health
	data, err := s.readAnalysis(input.Project, "code-packages")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no package health data for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if health, ok := findings["health"]; ok {
			result["health"] = health
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetLicenses(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// v4.0: Licenses are in code-packages scanner under findings.licenses
	data, err := s.readAnalysis(input.Project, "code-packages")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no license data for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if licenses, ok := findings["licenses"]; ok {
			result["licenses"] = licenses
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetSecrets(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// v4.0: Secrets are in code-security scanner under findings.secrets
	data, err := s.readAnalysis(input.Project, "code-security")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no secrets data for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if secrets, ok := findings["secrets"]; ok {
			result["secrets"] = secrets
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetCryptoIssues(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// v4.0: All crypto findings are in code-security scanner
	data, err := s.readAnalysis(input.Project, "code-security")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no crypto data for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		// Collect all crypto-related findings
		if ciphers, ok := findings["ciphers"]; ok {
			result["weak_ciphers"] = ciphers
		}
		if keys, ok := findings["keys"]; ok {
			result["hardcoded_keys"] = keys
		}
		if tls, ok := findings["tls"]; ok {
			result["tls_issues"] = tls
		}
		if random, ok := findings["random"]; ok {
			result["weak_random"] = random
		}
		if certs, ok := findings["certificates"]; ok {
			result["certificates"] = certs
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetAnalysisRaw(ctx context.Context, req *mcp.CallToolRequest, input AnalysisRawInput) (*mcp.CallToolResult, TextOutput, error) {
	data, err := s.readAnalysis(input.Project, input.AnalysisType)
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no %s data for '%s'", input.AnalysisType, input.Project)
	}

	result, _ := json.MarshalIndent(data, "", "  ")
	return nil, TextOutput{Text: string(result)}, nil
}

func (s *Server) handleGetDevOpsFindings(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// DevOps findings are in the devops scanner
	data, err := s.readAnalysis(input.Project, "devops")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no devops data for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if iac, ok := findings["iac"]; ok {
			result["iac_issues"] = iac
		}
		if containers, ok := findings["containers"]; ok {
			result["container_issues"] = containers
		}
		if actions, ok := findings["github_actions"]; ok {
			result["github_actions"] = actions
		}
		if git, ok := findings["git"]; ok {
			result["git_analysis"] = git
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetCodeQuality(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// Code quality findings are in the code-quality scanner
	data, err := s.readAnalysis(input.Project, "code-quality")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no code quality data for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if techDebt, ok := findings["tech_debt"]; ok {
			result["tech_debt"] = techDebt
		}
		if complexity, ok := findings["complexity"]; ok {
			result["complexity"] = complexity
		}
		if coverage, ok := findings["test_coverage"]; ok {
			result["test_coverage"] = coverage
		}
		if docs, ok := findings["documentation"]; ok {
			result["documentation"] = docs
		}
	}

	if summary, ok := data["summary"]; ok {
		result["summary"] = summary
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetOwnershipMetrics(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// Ownership metrics are in the code-ownership scanner
	data, err := s.readAnalysis(input.Project, "code-ownership")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no ownership data for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if contributors, ok := findings["contributors"]; ok {
			result["contributors"] = contributors
		}
		if busFactor, ok := findings["bus_factor"]; ok {
			result["bus_factor"] = busFactor
		}
		if codeowners, ok := findings["codeowners"]; ok {
			result["codeowners"] = codeowners
		}
		if orphans, ok := findings["orphans"]; ok {
			result["orphaned_code"] = orphans
		}
		if churn, ok := findings["churn"]; ok {
			result["code_churn"] = churn
		}
	}

	if summary, ok := data["summary"]; ok {
		result["summary"] = summary
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetDoraMetrics(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// DORA metrics are in the devops scanner under the dora feature
	data, err := s.readAnalysis(input.Project, "devops")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no DORA metrics for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if dora, ok := findings["dora"]; ok {
			result["dora_metrics"] = dora
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
}

func (s *Server) handleGetDevXAnalysis(ctx context.Context, req *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, TextOutput, error) {
	// Developer experience is in the developer-experience scanner
	data, err := s.readAnalysis(input.Project, "developer-experience")
	if err != nil {
		return nil, TextOutput{}, fmt.Errorf("no developer experience data for '%s'", input.Project)
	}

	result := map[string]interface{}{
		"project": input.Project,
	}

	if findings, ok := data["findings"].(map[string]interface{}); ok {
		if onboarding, ok := findings["onboarding"]; ok {
			result["onboarding"] = onboarding
		}
		if sprawl, ok := findings["sprawl"]; ok {
			result["tool_sprawl"] = sprawl
		}
		if workflow, ok := findings["workflow"]; ok {
			result["workflow"] = workflow
		}
	}

	if summary, ok := data["summary"]; ok {
		result["summary"] = summary
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return nil, TextOutput{Text: string(output)}, nil
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

	// v4.0: Super scanner names (all 7 analyzers)
	searchTypes := []string{"code-packages", "code-security", "technology-identification", "devops", "code-ownership", "code-quality", "developer-experience"}
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

// validateProjectID checks if a project ID is valid and safe
func validateProjectID(projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if len(projectID) > 200 {
		return fmt.Errorf("project ID too long (max 200 characters)")
	}
	if !validProjectPattern.MatchString(projectID) {
		return fmt.Errorf("invalid project ID format: must be 'owner/repo'")
	}
	// Additional safety check for path traversal
	if strings.Contains(projectID, "..") {
		return fmt.Errorf("invalid project ID: path traversal not allowed")
	}
	return nil
}

// safeJSONMarshal marshals data to JSON with size limits
// Returns the JSON and a warning message if truncated
func (s *Server) safeJSONMarshal(v interface{}) ([]byte, string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if len(data) > s.config.MaxOutputSize {
		// Add warning to result instead of replacing it
		warning := fmt.Sprintf("WARNING: Output truncated - size %d bytes exceeds limit of %d bytes. Use get_analysis_raw with filters or increase MaxOutputSize config.",
			len(data), s.config.MaxOutputSize)
		truncated := map[string]interface{}{
			"_warning":   warning,
			"_truncated": true,
			"_size":      len(data),
			"_limit":     s.config.MaxOutputSize,
			"_hint":      "Use get_analysis_raw with specific filters, or configure larger MaxOutputSize",
		}
		result, _ := json.MarshalIndent(truncated, "", "  ")
		return result, warning, nil
	}
	return data, "", nil
}

// limitFindingsWithWarning limits the number of findings and returns a warning if truncated
func (s *Server) limitFindingsWithWarning(findings []interface{}, category string) ([]interface{}, map[string]interface{}) {
	if len(findings) <= s.config.MaxFindingsPerCategory {
		return findings, nil
	}

	warning := map[string]interface{}{
		"_warning":        fmt.Sprintf("WARNING: %s findings truncated - showing %d of %d total. Increase MaxFindingsPerCategory config for more.", category, s.config.MaxFindingsPerCategory, len(findings)),
		"_truncated":      true,
		"_shown":          s.config.MaxFindingsPerCategory,
		"_total":          len(findings),
		"_limit":          s.config.MaxFindingsPerCategory,
		"_category":       category,
	}
	return findings[:s.config.MaxFindingsPerCategory], warning
}

// Standalone version for backward compatibility in tests
func safeJSONMarshal(v interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if len(data) > DefaultMaxOutputSize {
		truncated := map[string]interface{}{
			"_warning":   fmt.Sprintf("WARNING: Output truncated - exceeded %d bytes limit", DefaultMaxOutputSize),
			"_truncated": true,
		}
		return json.MarshalIndent(truncated, "", "  ")
	}
	return data, nil
}

// limitFindings limits the number of findings in a slice (backward compat)
func limitFindings(findings []interface{}, limit int) []interface{} {
	if len(findings) <= limit {
		return findings
	}
	return findings[:limit]
}

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
	// Validate project ID to prevent path traversal
	if err := validateProjectID(projectID); err != nil {
		return nil, fmt.Errorf("invalid project: %w", err)
	}

	// Validate analysis type (simple alphanumeric with hyphens)
	if analysisType == "" || len(analysisType) > 50 {
		return nil, fmt.Errorf("invalid analysis type")
	}
	for _, c := range analysisType {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return nil, fmt.Errorf("invalid analysis type: only lowercase letters, numbers, and hyphens allowed")
		}
	}

	path := filepath.Join(s.zeroHome, "repos", projectID, "analysis", analysisType+".json")

	// Check file exists and get info
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("analysis '%s' not found for project '%s'", analysisType, projectID)
		}
		return nil, fmt.Errorf("error accessing analysis file: %w", err)
	}

	// Check file size against configurable limit
	if info.Size() > s.config.MaxFileSize {
		return nil, fmt.Errorf("WARNING: Analysis file too large (%d bytes) - exceeds configured limit of %d bytes. Increase MaxFileSize config to read larger files",
			info.Size(), s.config.MaxFileSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading analysis file: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing analysis JSON: %w", err)
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
