// Package github provides GitHub API interactions
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

// TokenInfo contains information about the GitHub token
type TokenInfo struct {
	// Token type: "classic" (PAT), "fine-grained", "github-app", "oauth", or "unknown"
	Type string `json:"type"`
	// Scopes for classic PATs (e.g., "repo", "read:org", "admin:org")
	Scopes []string `json:"scopes,omitempty"`
	// Permissions for fine-grained PATs
	Permissions map[string]string `json:"permissions,omitempty"`
	// Username associated with the token
	Username string `json:"username,omitempty"`
	// Whether the token is valid
	Valid bool `json:"valid"`
	// Error message if invalid
	Error string `json:"error,omitempty"`
	// Rate limit info
	RateLimit     int `json:"rate_limit,omitempty"`
	RateRemaining int `json:"rate_remaining,omitempty"`
}

// ScannerRequirement defines what a scanner needs to function
type ScannerRequirement struct {
	Scanner     string   `json:"scanner"`
	Description string   `json:"description"`
	// Required scopes for classic PATs
	RequiredScopes []string `json:"required_scopes,omitempty"`
	// Required permissions for fine-grained PATs (permission -> access level)
	RequiredPermissions map[string]string `json:"required_permissions,omitempty"`
	// Whether the scanner needs GitHub API access at all
	NeedsGitHubAPI bool `json:"needs_github_api"`
	// Whether the scanner needs org-level access
	NeedsOrgAccess bool `json:"needs_org_access"`
	// External tools required
	RequiredTools []string `json:"required_tools,omitempty"`
}

// ScannerCompatibility shows whether a scanner will work
type ScannerCompatibility struct {
	Scanner       string   `json:"scanner"`
	Status        string   `json:"status"` // "ready", "limited", "unavailable"
	Reason        string   `json:"reason,omitempty"`
	MissingScopes []string `json:"missing_scopes,omitempty"`
	MissingTools  []string `json:"missing_tools,omitempty"`
}

// RoadmapResult contains the full compatibility analysis
type RoadmapResult struct {
	TokenInfo    TokenInfo              `json:"token_info"`
	Scanners     []ScannerCompatibility `json:"scanners"`
	Summary      RoadmapSummary         `json:"summary"`
	ToolsStatus  map[string]bool        `json:"tools_status"`
}

// RoadmapSummary provides a quick overview
type RoadmapSummary struct {
	Ready       int `json:"ready"`
	Limited     int `json:"limited"`
	Unavailable int `json:"unavailable"`
	Total       int `json:"total"`
}

// ScannerRequirements defines what each scanner needs
var ScannerRequirements = map[string]ScannerRequirement{
	// Package scanners - mostly local, some need API for health checks
	"package-sbom": {
		Scanner:        "package-sbom",
		Description:    "Generate Software Bill of Materials",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"cdxgen", "syft"},
	},
	"package-vulns": {
		Scanner:        "package-vulns",
		Description:    "Scan for known vulnerabilities",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"osv-scanner", "grype"},
	},
	"package-health": {
		Scanner:        "package-health",
		Description:    "Check package health scores",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"public_repo"},
		RequiredPermissions: map[string]string{
			"metadata": "read",
		},
	},
	"package-provenance": {
		Scanner:        "package-provenance",
		Description:    "Verify package provenance and signatures",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"public_repo"},
		RequiredPermissions: map[string]string{
			"metadata": "read",
		},
	},
	"package-malcontent": {
		Scanner:        "package-malcontent",
		Description:    "Detect malicious package behavior",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"malcontent"},
	},
	"licenses": {
		Scanner:        "licenses",
		Description:    "Analyze license compliance",
		NeedsGitHubAPI: false,
	},

	// Code security scanners - mostly local
	"code-vulns": {
		Scanner:        "code-vulns",
		Description:    "Static code analysis for vulnerabilities",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"semgrep"},
	},
	"code-secrets": {
		Scanner:        "code-secrets",
		Description:    "Detect hardcoded secrets",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"gitleaks", "trufflehog"},
	},
	"api-security": {
		Scanner:        "api-security",
		Description:    "Analyze API security",
		NeedsGitHubAPI: false,
	},

	// Crypto scanners - all local
	"crypto-ciphers": {
		Scanner:        "crypto-ciphers",
		Description:    "Detect weak cryptographic algorithms",
		NeedsGitHubAPI: false,
	},
	"crypto-keys": {
		Scanner:        "crypto-keys",
		Description:    "Find hardcoded keys",
		NeedsGitHubAPI: false,
	},
	"crypto-random": {
		Scanner:        "crypto-random",
		Description:    "Detect insecure random generation",
		NeedsGitHubAPI: false,
	},
	"crypto-tls": {
		Scanner:        "crypto-tls",
		Description:    "Analyze TLS configuration",
		NeedsGitHubAPI: false,
	},

	// Git/repo analysis - needs repo access
	"git": {
		Scanner:        "git",
		Description:    "Git history and contributor analysis",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo"},
		RequiredPermissions: map[string]string{
			"contents": "read",
		},
	},
	"code-ownership": {
		Scanner:        "code-ownership",
		Description:    "Analyze code ownership patterns",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo"},
		RequiredPermissions: map[string]string{
			"contents": "read",
		},
	},
	"dora-metrics": {
		Scanner:        "dora-metrics",
		Description:    "Calculate DORA metrics",
		NeedsGitHubAPI: true,
		NeedsOrgAccess: true,
		RequiredScopes: []string{"repo", "read:org"},
		RequiredPermissions: map[string]string{
			"contents":     "read",
			"pull_requests": "read",
			"organization_administration": "read",
		},
	},

	// Infrastructure scanners
	"iac-security": {
		Scanner:        "iac-security",
		Description:    "Infrastructure as Code security",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"trivy", "checkov"},
	},
	"container-security": {
		Scanner:        "container-security",
		Description:    "Container image analysis",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"trivy"},
	},
	"containers": {
		Scanner:        "containers",
		Description:    "Dockerfile best practices",
		NeedsGitHubAPI: false,
	},

	// Tech analysis - local
	"tech-discovery": {
		Scanner:        "tech-discovery",
		Description:    "Identify technologies used",
		NeedsGitHubAPI: false,
	},
	"tech-debt": {
		Scanner:        "tech-debt",
		Description:    "Analyze technical debt",
		NeedsGitHubAPI: false,
	},

	// GitHub-specific scanners that need API access
	"github-actions": {
		Scanner:        "github-actions",
		Description:    "Analyze GitHub Actions security",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "workflow"},
		RequiredPermissions: map[string]string{
			"actions":  "read",
			"contents": "read",
		},
	},
	"dependabot": {
		Scanner:        "dependabot",
		Description:    "Check Dependabot alerts",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "security_events"},
		RequiredPermissions: map[string]string{
			"vulnerability_alerts": "read",
		},
	},
	"code-scanning": {
		Scanner:        "code-scanning",
		Description:    "GitHub Code Scanning alerts",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "security_events"},
		RequiredPermissions: map[string]string{
			"security_events": "read",
		},
	},
	"secret-scanning": {
		Scanner:        "secret-scanning",
		Description:    "GitHub Secret Scanning alerts",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "security_events"},
		RequiredPermissions: map[string]string{
			"secret_scanning_alerts": "read",
		},
	},
}

// CheckTokenPermissions analyzes a GitHub token and returns its capabilities
func (c *Client) CheckTokenPermissions() (*TokenInfo, error) {
	info := &TokenInfo{
		Type:  "unknown",
		Valid: false,
	}

	if c.token == "" {
		info.Error = "No GitHub token found (set GITHUB_TOKEN or use 'gh auth login')"
		return info, nil
	}

	// Make a request to check the token
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to create request: %v", err)
		return info, nil
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to connect to GitHub: %v", err)
		return info, nil
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == 401 {
		info.Error = "Token is invalid or expired"
		return info, nil
	}

	if resp.StatusCode != 200 {
		info.Error = fmt.Sprintf("GitHub API returned status %d", resp.StatusCode)
		return info, nil
	}

	info.Valid = true

	// Parse user info
	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err == nil {
		info.Username = user.Login
	}

	// Check scopes from header (classic PATs)
	scopeHeader := resp.Header.Get("X-OAuth-Scopes")
	if scopeHeader != "" {
		info.Type = "classic"
		scopes := strings.Split(scopeHeader, ", ")
		for _, s := range scopes {
			s = strings.TrimSpace(s)
			if s != "" {
				info.Scopes = append(info.Scopes, s)
			}
		}
	}

	// Check for fine-grained token indicators
	// Fine-grained tokens don't return X-OAuth-Scopes
	if scopeHeader == "" && info.Valid {
		// Could be fine-grained or GitHub App token
		// Try to detect by checking a permission-specific endpoint
		info.Type = "fine-grained"
		info.Permissions = c.detectFineGrainedPermissions()
	}

	// Get rate limit info
	if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
		fmt.Sscanf(limit, "%d", &info.RateLimit)
	}
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		fmt.Sscanf(remaining, "%d", &info.RateRemaining)
	}

	return info, nil
}

// detectFineGrainedPermissions tries to determine permissions for fine-grained tokens
func (c *Client) detectFineGrainedPermissions() map[string]string {
	perms := make(map[string]string)

	// Test various endpoints to detect permissions
	endpoints := map[string]struct {
		url        string
		permission string
		level      string
	}{
		"repos": {
			url:        "https://api.github.com/user/repos?per_page=1",
			permission: "contents",
			level:      "read",
		},
		"orgs": {
			url:        "https://api.github.com/user/orgs?per_page=1",
			permission: "organization",
			level:      "read",
		},
	}

	for _, ep := range endpoints {
		req, _ := http.NewRequest("GET", ep.url, nil)
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			perms[ep.permission] = ep.level
		}
	}

	return perms
}

// CheckToolAvailability checks if required external tools are installed
func CheckToolAvailability(tools []string) map[string]bool {
	status := make(map[string]bool)
	for _, tool := range tools {
		_, err := exec.LookPath(tool)
		status[tool] = err == nil
	}
	return status
}

// GenerateRoadmap analyzes token permissions and tool availability
func (c *Client) GenerateRoadmap(scanners []string) (*RoadmapResult, error) {
	result := &RoadmapResult{
		Scanners:    make([]ScannerCompatibility, 0),
		ToolsStatus: make(map[string]bool),
	}

	// Check token permissions
	tokenInfo, err := c.CheckTokenPermissions()
	if err != nil {
		return nil, err
	}
	result.TokenInfo = *tokenInfo

	// Collect all required tools
	allTools := make(map[string]bool)
	for _, name := range scanners {
		if req, ok := ScannerRequirements[name]; ok {
			for _, tool := range req.RequiredTools {
				allTools[tool] = true
			}
		}
	}

	// Check tool availability
	toolList := make([]string, 0, len(allTools))
	for tool := range allTools {
		toolList = append(toolList, tool)
	}
	result.ToolsStatus = CheckToolAvailability(toolList)

	// Analyze each scanner
	for _, name := range scanners {
		compat := ScannerCompatibility{
			Scanner: name,
			Status:  "ready",
		}

		req, ok := ScannerRequirements[name]
		if !ok {
			// Unknown scanner, assume it works
			compat.Reason = "No specific requirements defined"
			result.Scanners = append(result.Scanners, compat)
			continue
		}

		// Check required tools
		for _, tool := range req.RequiredTools {
			if !result.ToolsStatus[tool] {
				compat.MissingTools = append(compat.MissingTools, tool)
			}
		}

		// Check GitHub API requirements
		if req.NeedsGitHubAPI {
			if !tokenInfo.Valid {
				compat.Status = "unavailable"
				compat.Reason = "Requires valid GitHub token"
			} else if tokenInfo.Type == "classic" {
				// Check scopes for classic PAT
				for _, scope := range req.RequiredScopes {
					if !hasScope(tokenInfo.Scopes, scope) {
						compat.MissingScopes = append(compat.MissingScopes, scope)
					}
				}
			} else if tokenInfo.Type == "fine-grained" {
				// Check permissions for fine-grained PAT
				for perm, level := range req.RequiredPermissions {
					if tokenLevel, ok := tokenInfo.Permissions[perm]; !ok || !hasAccessLevel(tokenLevel, level) {
						compat.MissingScopes = append(compat.MissingScopes, fmt.Sprintf("%s:%s", perm, level))
					}
				}
			}
		}

		// Determine final status
		if len(compat.MissingTools) > 0 && len(compat.MissingScopes) > 0 {
			compat.Status = "unavailable"
			compat.Reason = fmt.Sprintf("Missing tools (%s) and permissions (%s)",
				strings.Join(compat.MissingTools, ", "),
				strings.Join(compat.MissingScopes, ", "))
		} else if len(compat.MissingTools) > 0 {
			// Check if any alternative tool is available
			hasAlternative := false
			for _, tool := range req.RequiredTools {
				if result.ToolsStatus[tool] {
					hasAlternative = true
					break
				}
			}
			if hasAlternative {
				compat.Status = "ready"
				compat.Reason = fmt.Sprintf("Using available tool (missing: %s)",
					strings.Join(compat.MissingTools, ", "))
			} else {
				compat.Status = "unavailable"
				compat.Reason = fmt.Sprintf("Missing required tools: %s",
					strings.Join(compat.MissingTools, ", "))
			}
		} else if len(compat.MissingScopes) > 0 {
			compat.Status = "limited"
			compat.Reason = fmt.Sprintf("Missing permissions: %s",
				strings.Join(compat.MissingScopes, ", "))
		}

		result.Scanners = append(result.Scanners, compat)
	}

	// Calculate summary
	for _, s := range result.Scanners {
		result.Summary.Total++
		switch s.Status {
		case "ready":
			result.Summary.Ready++
		case "limited":
			result.Summary.Limited++
		case "unavailable":
			result.Summary.Unavailable++
		}
	}

	return result, nil
}

// hasScope checks if a scope is present in the list
func hasScope(scopes []string, target string) bool {
	for _, s := range scopes {
		// Handle scope hierarchies (e.g., "repo" includes "public_repo")
		if s == target {
			return true
		}
		// "repo" scope includes many sub-scopes
		if s == "repo" && (target == "public_repo" || strings.HasPrefix(target, "repo:")) {
			return true
		}
		// "admin:org" includes "read:org" and "write:org"
		if s == "admin:org" && (target == "read:org" || target == "write:org") {
			return true
		}
		if s == "write:org" && target == "read:org" {
			return true
		}
	}
	return false
}

// hasAccessLevel checks if the token has sufficient access level
func hasAccessLevel(have, need string) bool {
	levels := map[string]int{
		"none":  0,
		"read":  1,
		"write": 2,
		"admin": 3,
	}
	return levels[have] >= levels[need]
}

// GetToken returns the current token (for external use)
func (c *Client) GetToken() string {
	return c.token
}

// HasToken returns true if a token is configured
func (c *Client) HasToken() bool {
	return c.token != ""
}
