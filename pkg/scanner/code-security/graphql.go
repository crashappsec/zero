// Package codesecurity provides the consolidated code security super scanner
package codesecurity

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/crashappsec/zero/pkg/scanner"
)

// GraphQL security check patterns
var graphqlPatterns = []struct {
	Name        string
	Pattern     *regexp.Regexp
	Severity    string
	Category    string
	OWASPApi    string
	CWE         string
	Description string
	Remediation string
}{
	{
		Name:        "GraphQL Introspection Enabled",
		Pattern:     regexp.MustCompile(`introspection\s*[=:]\s*true|enableIntrospection\s*[=:]\s*true`),
		Severity:    "medium",
		Category:    "api-data-exposure",
		OWASPApi:    "API3:2023",
		CWE:         "CWE-200",
		Description: "GraphQL introspection is enabled, exposing the entire API schema to attackers",
		Remediation: "Disable introspection in production: introspection: false",
	},
	{
		Name:        "GraphQL Introspection Not Disabled",
		Pattern:     regexp.MustCompile(`new\s+ApolloServer\s*\(\s*\{[^}]*\}\s*\)`),
		Severity:    "low",
		Category:    "api-data-exposure",
		OWASPApi:    "API3:2023",
		CWE:         "CWE-200",
		Description: "ApolloServer instantiated without explicit introspection setting (enabled by default in dev)",
		Remediation: "Explicitly set introspection: process.env.NODE_ENV !== 'production'",
	},
	{
		Name:        "GraphQL Server Configuration",
		Pattern:     regexp.MustCompile(`(ApolloServer|graphqlHTTP|GraphQLServer)\s*\(\s*\{`),
		Severity:    "info",
		Category:    "api-rate-limiting",
		OWASPApi:    "API4:2023",
		CWE:         "CWE-770",
		Description: "GraphQL server configuration - verify depth and complexity limits are set",
		Remediation: "Add depth limiting: validationRules: [depthLimit(10)] and complexity analysis",
	},
	{
		Name:        "GraphQL Batching Enabled",
		Pattern:     regexp.MustCompile(`allowBatchedHttpRequests\s*[=:]\s*true|batch\s*[=:]\s*true`),
		Severity:    "medium",
		Category:    "api-rate-limiting",
		OWASPApi:    "API4:2023",
		CWE:         "CWE-770",
		Description: "GraphQL batching is enabled, allowing multiple operations in a single request",
		Remediation: "Disable batching or implement per-operation rate limiting",
	},
	{
		Name:        "GraphQL Playground in Production",
		Pattern:     regexp.MustCompile(`playground\s*[=:]\s*true|graphqlPlayground\s*[=:]\s*true`),
		Severity:    "low",
		Category:    "api-data-exposure",
		OWASPApi:    "API8:2023",
		CWE:         "CWE-200",
		Description: "GraphQL Playground enabled, may expose schema and allow query exploration",
		Remediation: "Disable playground in production: playground: process.env.NODE_ENV !== 'production'",
	},
	{
		Name:        "GraphQL Field Suggestion Enabled",
		Pattern:     regexp.MustCompile(`fieldSuggestion\s*[=:]\s*true|suggestions\s*[=:]\s*true`),
		Severity:    "low",
		Category:    "api-data-exposure",
		OWASPApi:    "API3:2023",
		CWE:         "CWE-200",
		Description: "GraphQL field suggestions enabled, may leak schema information through error messages",
		Remediation: "Disable field suggestions in production",
	},
	{
		Name:        "GraphQL Debug Mode",
		Pattern:     regexp.MustCompile(`debug\s*[=:]\s*true|includeStacktraceInErrorResponses\s*[=:]\s*true`),
		Severity:    "medium",
		Category:    "api-data-exposure",
		OWASPApi:    "API8:2023",
		CWE:         "CWE-209",
		Description: "GraphQL debug mode enabled, may expose sensitive error details",
		Remediation: "Disable debug mode in production",
	},
	{
		Name:        "Unsafe Resolver with SQL",
		Pattern:     regexp.MustCompile(`resolve[rs]?\s*[=:]\s*(?:async\s*)?\(?[^)]*\)?\s*=>\s*\{[^}]*(?:query|execute)\s*\(`),
		Severity:    "critical",
		Category:    "api-injection",
		OWASPApi:    "API8:2023",
		CWE:         "CWE-89",
		Description: "GraphQL resolver uses string interpolation in SQL query, vulnerable to injection",
		Remediation: "Use parameterized queries or an ORM",
	},
	{
		Name:        "Missing Field-Level Authorization",
		Pattern:     regexp.MustCompile(`@auth|@hasRole|@requireAuth|fieldAuthorizationPlugin|AuthorizationDirective`),
		Severity:    "info",
		Category:    "api-auth",
		OWASPApi:    "API1:2023",
		CWE:         "CWE-862",
		Description: "GraphQL schema may lack field-level authorization (no auth directives found)",
		Remediation: "Implement field-level authorization using directives or resolver middleware",
	},
}

// runGraphQLChecks performs GraphQL-specific security analysis
func (s *CodeSecurityScanner) runGraphQLChecks(ctx context.Context, opts *scanner.ScanOptions) []APIFinding {
	var findings []APIFinding

	// Check if GraphQL is used in this repository
	hasGraphQL := s.detectGraphQLUsage(opts.RepoPath)
	if !hasGraphQL {
		return findings
	}

	// Scan relevant files
	fileExtensions := []string{".js", ".ts", ".jsx", ".tsx", ".py", ".rb", ".graphql", ".gql"}

	for _, ext := range fileExtensions {
		err := filepath.Walk(opts.RepoPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			// Skip directories and non-matching files
			if info.IsDir() {
				if shouldSkipGraphQLDir(info.Name()) {
					return filepath.SkipDir
				}
				return nil
			}

			if !strings.HasSuffix(path, ext) {
				return nil
			}

			fileFindings := s.scanGraphQLFile(path, opts.RepoPath)
			findings = append(findings, fileFindings...)
			return nil
		})
		if err != nil {
			continue
		}
	}

	// Check for missing security configurations
	findings = append(findings, s.checkGraphQLMissingConfigs(opts.RepoPath)...)

	return findings
}

// detectGraphQLUsage checks if the repository uses GraphQL
func (s *CodeSecurityScanner) detectGraphQLUsage(repoPath string) bool {
	graphqlIndicators := []string{
		"graphql",
		"apollo",
		"@graphql",
		"type Query",
		"type Mutation",
		"schema {",
		"gql`",
	}

	// Check package.json
	pkgPath := filepath.Join(repoPath, "package.json")
	if content, err := os.ReadFile(pkgPath); err == nil {
		contentStr := string(content)
		for _, indicator := range graphqlIndicators[:2] {
			if strings.Contains(contentStr, indicator) {
				return true
			}
		}
	}

	// Check for .graphql files
	graphqlFiles, _ := filepath.Glob(filepath.Join(repoPath, "**/*.graphql"))
	if len(graphqlFiles) > 0 {
		return true
	}

	// Check for schema files
	schemaPatterns := []string{"schema.graphql", "schema.gql", "*.graphqls"}
	for _, pattern := range schemaPatterns {
		matches, _ := filepath.Glob(filepath.Join(repoPath, pattern))
		if len(matches) > 0 {
			return true
		}
	}

	return false
}

// shouldSkipGraphQLDir checks if a directory should be skipped
func shouldSkipGraphQLDir(name string) bool {
	skipDirs := []string{
		"node_modules", "vendor", ".git", "dist", "build",
		"coverage", "__pycache__", ".venv", "venv",
	}
	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}
	return false
}

// scanGraphQLFile scans a single file for GraphQL security issues
func (s *CodeSecurityScanner) scanGraphQLFile(filePath, repoPath string) []APIFinding {
	var findings []APIFinding

	file, err := os.Open(filePath)
	if err != nil {
		return findings
	}
	defer file.Close()

	relPath := strings.TrimPrefix(filePath, repoPath)
	relPath = strings.TrimPrefix(relPath, "/")

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var fullContent strings.Builder

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		fullContent.WriteString(line)
		fullContent.WriteString("\n")

		// Check each pattern against this line
		for _, p := range graphqlPatterns {
			// Skip the "missing auth" info pattern - check separately
			if p.Severity == "info" {
				continue
			}

			if p.Pattern.MatchString(line) {
				finding := APIFinding{
					RuleID:      "graphql-" + strings.ToLower(strings.ReplaceAll(p.Name, " ", "-")),
					Title:       p.Name,
					Description: p.Description,
					Severity:    p.Severity,
					Confidence:  "high",
					File:        relPath,
					Line:        lineNum,
					Snippet:     truncateSnippet(line, 200),
					Category:    p.Category,
					OWASPApi:    p.OWASPApi,
					CWE:         []string{p.CWE},
					Framework:   "graphql",
					Endpoint:    "/graphql",
					Remediation: p.Remediation,
				}
				findings = append(findings, finding)
			}
		}
	}

	// Check for presence of auth directives in full content
	content := fullContent.String()
	if strings.Contains(content, "type Query") || strings.Contains(content, "type Mutation") {
		hasAuthDirective := false
		authPatterns := []string{
			"@auth", "@hasRole", "@requireAuth", "@authenticated",
			"@isAuthenticated", "@hasPermission", "@can",
		}
		for _, pattern := range authPatterns {
			if strings.Contains(content, pattern) {
				hasAuthDirective = true
				break
			}
		}

		if !hasAuthDirective && strings.Contains(relPath, "schema") {
			findings = append(findings, APIFinding{
				RuleID:      "graphql-no-auth-directives",
				Title:       "GraphQL Schema Without Auth Directives",
				Description: "GraphQL schema file does not contain authorization directives",
				Severity:    "medium",
				Confidence:  "medium",
				File:        relPath,
				Line:        1,
				Category:    "api-auth",
				OWASPApi:    "API1:2023",
				CWE:         []string{"CWE-862"},
				Framework:   "graphql",
				Endpoint:    "/graphql",
				Remediation: "Add @auth or @hasRole directives to protect sensitive fields and operations",
			})
		}
	}

	return findings
}

// checkGraphQLMissingConfigs checks for missing GraphQL security configurations
func (s *CodeSecurityScanner) checkGraphQLMissingConfigs(repoPath string) []APIFinding {
	var findings []APIFinding

	// Check for depth limiting library
	hasDepthLimit := false
	hasComplexityLimit := false

	pkgPath := filepath.Join(repoPath, "package.json")
	if content, err := os.ReadFile(pkgPath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "graphql-depth-limit") {
			hasDepthLimit = true
		}
		if strings.Contains(contentStr, "graphql-query-complexity") ||
			strings.Contains(contentStr, "graphql-cost-analysis") {
			hasComplexityLimit = true
		}
	}

	// Check for Python GraphQL libraries
	reqPath := filepath.Join(repoPath, "requirements.txt")
	if content, err := os.ReadFile(reqPath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "graphene") || strings.Contains(contentStr, "ariadne") {
			// Python GraphQL - check for depth limiting
			if !strings.Contains(contentStr, "graphql-depth-limit") {
				hasDepthLimit = false
			}
		}
	}

	if !hasDepthLimit {
		findings = append(findings, APIFinding{
			RuleID:      "graphql-missing-depth-limit-package",
			Title:       "Missing Query Depth Limiting",
			Description: "No query depth limiting package found (graphql-depth-limit)",
			Severity:    "medium",
			Confidence:  "medium",
			File:        "package.json",
			Line:        1,
			Category:    "api-rate-limiting",
			OWASPApi:    "API4:2023",
			CWE:         []string{"CWE-770"},
			Framework:   "graphql",
			Endpoint:    "/graphql",
			Remediation: "Install graphql-depth-limit: npm install graphql-depth-limit",
		})
	}

	if !hasComplexityLimit {
		findings = append(findings, APIFinding{
			RuleID:      "graphql-missing-complexity-package",
			Title:       "Missing Query Complexity Analysis",
			Description: "No query complexity analysis package found",
			Severity:    "medium",
			Confidence:  "medium",
			File:        "package.json",
			Line:        1,
			Category:    "api-rate-limiting",
			OWASPApi:    "API4:2023",
			CWE:         []string{"CWE-770"},
			Framework:   "graphql",
			Endpoint:    "/graphql",
			Remediation: "Install graphql-query-complexity: npm install graphql-query-complexity",
		})
	}

	return findings
}
