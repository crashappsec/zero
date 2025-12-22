// Package codesecurity provides the consolidated code security super scanner
package codesecurity

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/crashappsec/zero/pkg/scanner"
	"gopkg.in/yaml.v3"
)

// OpenAPISpec represents a parsed OpenAPI/Swagger specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi" yaml:"openapi"`
	Swagger    string                 `json:"swagger" yaml:"swagger"`
	Info       OpenAPIInfo            `json:"info" yaml:"info"`
	Paths      map[string]PathItem    `json:"paths" yaml:"paths"`
	Components OpenAPIComponents      `json:"components" yaml:"components"`
	Security   []map[string][]string  `json:"security" yaml:"security"`
}

// OpenAPIInfo contains API metadata
type OpenAPIInfo struct {
	Title   string `json:"title" yaml:"title"`
	Version string `json:"version" yaml:"version"`
}

// PathItem represents an API path with its operations
type PathItem struct {
	Get     *Operation `json:"get" yaml:"get"`
	Post    *Operation `json:"post" yaml:"post"`
	Put     *Operation `json:"put" yaml:"put"`
	Delete  *Operation `json:"delete" yaml:"delete"`
	Patch   *Operation `json:"patch" yaml:"patch"`
	Options *Operation `json:"options" yaml:"options"`
}

// Operation represents an API operation
type Operation struct {
	OperationID string                `json:"operationId" yaml:"operationId"`
	Summary     string                `json:"summary" yaml:"summary"`
	Description string                `json:"description" yaml:"description"`
	Security    []map[string][]string `json:"security" yaml:"security"`
	Deprecated  bool                  `json:"deprecated" yaml:"deprecated"`
	Tags        []string              `json:"tags" yaml:"tags"`
	Parameters  []Parameter           `json:"parameters" yaml:"parameters"`
	RequestBody *RequestBody          `json:"requestBody" yaml:"requestBody"`
	Responses   map[string]Response   `json:"responses" yaml:"responses"`
}

// Parameter represents an API parameter
type Parameter struct {
	Name     string `json:"name" yaml:"name"`
	In       string `json:"in" yaml:"in"` // query, path, header, cookie
	Required bool   `json:"required" yaml:"required"`
	Schema   Schema `json:"schema" yaml:"schema"`
}

// RequestBody represents a request body
type RequestBody struct {
	Required bool `json:"required" yaml:"required"`
}

// Response represents an API response
type Response struct {
	Description string `json:"description" yaml:"description"`
}

// Schema represents a JSON schema
type Schema struct {
	Type string `json:"type" yaml:"type"`
}

// OpenAPIComponents contains reusable components
type OpenAPIComponents struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes" yaml:"securitySchemes"`
}

// SecurityScheme represents an authentication scheme
type SecurityScheme struct {
	Type         string `json:"type" yaml:"type"`
	Scheme       string `json:"scheme" yaml:"scheme"`
	BearerFormat string `json:"bearerFormat" yaml:"bearerFormat"`
	In           string `json:"in" yaml:"in"`
	Name         string `json:"name" yaml:"name"`
}

// runOpenAPIValidation scans for OpenAPI specs and validates security
func (s *CodeSecurityScanner) runOpenAPIValidation(ctx context.Context, opts *scanner.ScanOptions) []APIFinding {
	var findings []APIFinding

	// Find OpenAPI/Swagger spec files
	specPatterns := []string{
		"openapi.yaml", "openapi.yml", "openapi.json",
		"swagger.yaml", "swagger.yml", "swagger.json",
		"api.yaml", "api.yml", "api.json",
		"api-spec.yaml", "api-spec.yml", "api-spec.json",
	}

	for _, pattern := range specPatterns {
		// Check root
		specPath := filepath.Join(opts.RepoPath, pattern)
		if specFindings := s.validateOpenAPISpec(specPath, opts.RepoPath); specFindings != nil {
			findings = append(findings, specFindings...)
		}

		// Check common subdirectories
		for _, subdir := range []string{"docs", "api", "spec", "specs", "openapi"} {
			specPath = filepath.Join(opts.RepoPath, subdir, pattern)
			if specFindings := s.validateOpenAPISpec(specPath, opts.RepoPath); specFindings != nil {
				findings = append(findings, specFindings...)
			}
		}
	}

	return findings
}

// validateOpenAPISpec validates a single OpenAPI spec file
func (s *CodeSecurityScanner) validateOpenAPISpec(specPath, repoPath string) []APIFinding {
	content, err := os.ReadFile(specPath)
	if err != nil {
		return nil // File doesn't exist
	}

	var spec OpenAPISpec
	relPath := strings.TrimPrefix(specPath, repoPath)
	relPath = strings.TrimPrefix(relPath, "/")

	// Parse based on extension
	if strings.HasSuffix(specPath, ".json") {
		if err := json.Unmarshal(content, &spec); err != nil {
			return []APIFinding{{
				RuleID:      "openapi-parse-error",
				Title:       "Invalid OpenAPI Specification",
				Description: "Failed to parse OpenAPI/Swagger specification file",
				Severity:    "medium",
				Confidence:  "high",
				File:        relPath,
				Line:        1,
				Category:    "api-misconfiguration",
				OWASPApi:    "API9:2023",
			}}
		}
	} else {
		if err := yaml.Unmarshal(content, &spec); err != nil {
			return []APIFinding{{
				RuleID:      "openapi-parse-error",
				Title:       "Invalid OpenAPI Specification",
				Description: "Failed to parse OpenAPI/Swagger specification file",
				Severity:    "medium",
				Confidence:  "high",
				File:        relPath,
				Line:        1,
				Category:    "api-misconfiguration",
				OWASPApi:    "API9:2023",
			}}
		}
	}

	var findings []APIFinding

	// Check 1: Missing global security schemes
	if len(spec.Components.SecuritySchemes) == 0 && len(spec.Security) == 0 {
		findings = append(findings, APIFinding{
			RuleID:      "openapi-no-security-schemes",
			Title:       "No Security Schemes Defined",
			Description: "The OpenAPI specification has no security schemes defined. All endpoints may be publicly accessible.",
			Severity:    "high",
			Confidence:  "high",
			File:        relPath,
			Line:        1,
			Category:    "api-auth",
			OWASPApi:    "API2:2023",
			Remediation: "Add securitySchemes to components and apply security requirements to endpoints",
		})
	}

	// Check 2: Endpoints without security requirements
	hasGlobalSecurity := len(spec.Security) > 0

	for path, pathItem := range spec.Paths {
		operations := map[string]*Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"DELETE":  pathItem.Delete,
			"PATCH":   pathItem.Patch,
			"OPTIONS": pathItem.Options,
		}

		for method, op := range operations {
			if op == nil {
				continue
			}

			// Check for missing security
			if !hasGlobalSecurity && len(op.Security) == 0 {
				// Some endpoints like health checks are OK without auth
				if !isExemptEndpoint(path) {
					findings = append(findings, APIFinding{
						RuleID:      "openapi-endpoint-no-auth",
						Title:       "Endpoint Without Authentication",
						Description: "API endpoint has no security requirements defined",
						Severity:    "high",
						Confidence:  "medium",
						File:        relPath,
						Line:        1,
						Category:    "api-auth",
						OWASPApi:    "API2:2023",
						HTTPMethod:  method,
						Endpoint:    path,
						Remediation: "Add security requirements to this endpoint or define global security",
					})
				}
			}

			// Check for deprecated endpoints
			if op.Deprecated {
				findings = append(findings, APIFinding{
					RuleID:      "openapi-deprecated-endpoint",
					Title:       "Deprecated API Endpoint",
					Description: "This endpoint is marked as deprecated and should be reviewed for removal",
					Severity:    "low",
					Confidence:  "high",
					File:        relPath,
					Line:        1,
					Category:    "api-versioning",
					OWASPApi:    "API9:2023",
					HTTPMethod:  method,
					Endpoint:    path,
					Remediation: "Consider removing deprecated endpoints from the API",
				})
			}

			// Check for sensitive operations without rate limiting headers
			if isSensitiveOperation(method, path) {
				// Check responses for rate limit headers
				hasRateLimitHeaders := false
				for _, resp := range op.Responses {
					if strings.Contains(strings.ToLower(resp.Description), "rate limit") {
						hasRateLimitHeaders = true
						break
					}
				}

				if !hasRateLimitHeaders {
					findings = append(findings, APIFinding{
						RuleID:      "openapi-no-rate-limit",
						Title:       "Missing Rate Limit Documentation",
						Description: "Sensitive endpoint does not document rate limiting in responses",
						Severity:    "medium",
						Confidence:  "low",
						File:        relPath,
						Line:        1,
						Category:    "api-rate-limiting",
						OWASPApi:    "API4:2023",
						HTTPMethod:  method,
						Endpoint:    path,
						Remediation: "Document rate limiting behavior in response headers (X-RateLimit-*)",
					})
				}
			}
		}
	}

	// Check 3: Using basic auth without HTTPS
	for name, scheme := range spec.Components.SecuritySchemes {
		if scheme.Type == "http" && scheme.Scheme == "basic" {
			findings = append(findings, APIFinding{
				RuleID:      "openapi-basic-auth",
				Title:       "Basic Authentication Detected",
				Description: "Basic authentication transmits credentials in every request. Consider using token-based auth.",
				Severity:    "medium",
				Confidence:  "high",
				File:        relPath,
				Line:        1,
				Category:    "api-auth",
				OWASPApi:    "API2:2023",
				Remediation: "Consider using Bearer token or OAuth2 instead of Basic auth. Security scheme: " + name,
			})
		}

		// Check for API key in query string
		if scheme.Type == "apiKey" && scheme.In == "query" {
			findings = append(findings, APIFinding{
				RuleID:      "openapi-apikey-in-query",
				Title:       "API Key in Query String",
				Description: "API keys in query strings can be logged in server logs and browser history",
				Severity:    "medium",
				Confidence:  "high",
				File:        relPath,
				Line:        1,
				Category:    "api-auth",
				OWASPApi:    "API2:2023",
				Remediation: "Move API key to header (X-API-Key) instead of query parameter. Security scheme: " + name,
			})
		}
	}

	return findings
}

// isExemptEndpoint checks if an endpoint is exempt from auth requirements
func isExemptEndpoint(path string) bool {
	exemptPatterns := []string{
		"/health", "/healthz", "/ready", "/readiness", "/live", "/liveness",
		"/ping", "/status", "/version", "/info",
		"/docs", "/swagger", "/openapi", "/redoc",
		"/metrics", "/prometheus",
	}

	pathLower := strings.ToLower(path)
	for _, pattern := range exemptPatterns {
		if strings.Contains(pathLower, pattern) {
			return true
		}
	}
	return false
}

// isSensitiveOperation checks if an operation is security-sensitive
func isSensitiveOperation(method, path string) bool {
	// All write operations are sensitive
	if method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH" {
		return true
	}

	// Sensitive paths
	sensitivePatterns := []string{
		"/auth", "/login", "/register", "/password", "/token",
		"/admin", "/user", "/account", "/profile",
		"/payment", "/billing", "/subscription",
		"/key", "/secret", "/credential",
	}

	pathLower := strings.ToLower(path)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(pathLower, pattern) {
			return true
		}
	}
	return false
}
