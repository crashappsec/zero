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

// API quality patterns (non-security)
var apiQualityPatterns = []struct {
	Name        string
	Pattern     *regexp.Regexp
	Category    string // api-design, api-performance, api-observability, api-documentation
	Severity    string // info, low, medium
	Description string
	Remediation string
}{
	// API Design Patterns
	{
		Name:        "Non-RESTful Route Naming",
		Pattern:     regexp.MustCompile(`\.(get|post)\s*\(\s*['"]/(get|fetch|retrieve|find)[A-Z]`),
		Category:    "api-design",
		Severity:    "info",
		Description: "Route uses verb in path (getUsers) instead of RESTful noun (users)",
		Remediation: "Use nouns for resources: GET /users instead of GET /getUsers",
	},
	{
		Name:        "POST for Read Operation",
		Pattern:     regexp.MustCompile(`\.post\s*\(\s*['"][^'"]*/(get|fetch|retrieve|find|list|search)[^'"]*['"]`),
		Category:    "api-design",
		Severity:    "low",
		Description: "Using POST for a read operation that should be GET",
		Remediation: "Use GET for read operations, POST for creating resources",
	},
	{
		Name:        "GET with Request Body",
		Pattern:     regexp.MustCompile(`\.get\s*\([^)]+\)\s*(?:=>|{)[^}]*req\.body`),
		Category:    "api-design",
		Severity:    "medium",
		Description: "GET request handler accesses request body, which is discouraged",
		Remediation: "GET requests should use query parameters, not body. Use POST if body is needed",
	},
	{
		Name:        "Inconsistent Response Format",
		Pattern:     regexp.MustCompile(`res\.json\s*\(\s*\{[^}]*\}\s*\)[^}]*res\.json\s*\(\s*[^\{]`),
		Category:    "api-design",
		Severity:    "low",
		Description: "Inconsistent response formats in same handler (object vs primitive)",
		Remediation: "Use consistent response wrapper: { data: ..., error: null }",
	},
	{
		Name:        "Magic Status Code",
		Pattern:     regexp.MustCompile(`\.status\s*\(\s*[0-9]+\s*\)`),
		Category:    "api-design",
		Severity:    "info",
		Description: "Using magic number for HTTP status code instead of named constant",
		Remediation: "Use named constants: res.status(HTTP_NOT_FOUND) or res.sendStatus(404)",
	},

	// API Performance Patterns
	{
		Name:        "N+1 Query Pattern",
		Pattern:     regexp.MustCompile(`(?:for|forEach|map)\s*\([^)]+\)\s*(?:=>|{)[^}]*(?:await|\.then)[^}]*(?:findOne|findById|get|fetch)`),
		Category:    "api-performance",
		Severity:    "high",
		Description: "Potential N+1 query pattern: database call inside loop",
		Remediation: "Use batch queries or include/populate to fetch related data in single query",
	},
	{
		Name:        "Missing Pagination",
		Pattern:     regexp.MustCompile(`\.find\s*\(\s*\{\s*\}\s*\)|\.find\s*\(\s*\)|\.(findAll|all)\s*\(`),
		Category:    "api-performance",
		Severity:    "medium",
		Description: "Unbounded query that returns all records without pagination",
		Remediation: "Add .limit() and .skip() or use pagination library",
	},
	{
		Name:        "Response Without Caching Headers",
		Pattern:     regexp.MustCompile(`res\.(json|send)\s*\([^)]+\)`),
		Category:    "api-performance",
		Severity:    "info",
		Description: "Response sent - verify caching headers are set appropriately",
		Remediation: "Consider adding Cache-Control or ETag headers for cacheable responses",
	},
	{
		Name:        "Synchronous File Operation in Handler",
		Pattern:     regexp.MustCompile(`(?:readFileSync|writeFileSync|appendFileSync|existsSync)\s*\(`),
		Category:    "api-performance",
		Severity:    "medium",
		Description: "Synchronous file operation blocks the event loop",
		Remediation: "Use async versions: readFile, writeFile, access with await",
	},
	{
		Name:        "Large Payload Without Streaming",
		Pattern:     regexp.MustCompile(`JSON\.stringify\s*\([^)]*\)|res\.json\s*\(\s*(?:results|data|items|records)\s*\)`),
		Category:    "api-performance",
		Severity:    "info",
		Description: "Large payload may benefit from streaming response",
		Remediation: "Consider streaming for large datasets using res.write() or streams",
	},

	// API Observability Patterns
	{
		Name:        "Error Handler Without Logging",
		Pattern:     regexp.MustCompile(`catch\s*\(\s*(?:err|error|e)\s*\)\s*\{[^}]*res\.(status|json|send)`),
		Category:    "api-observability",
		Severity:    "medium",
		Description: "Error handler found - verify errors are logged before returning",
		Remediation: "Log errors before returning error response: logger.error(err)",
	},
	{
		Name:        "Console.log in Production Code",
		Pattern:     regexp.MustCompile(`console\.(log|info|warn|error)\s*\(`),
		Category:    "api-observability",
		Severity:    "low",
		Description: "Using console.log instead of structured logging",
		Remediation: "Use a structured logger (winston, pino, bunyan) for production",
	},
	{
		Name:        "Swallowed Exception",
		Pattern:     regexp.MustCompile(`catch\s*\(\s*(?:err|error|e|_)\s*\)\s*\{\s*\}`),
		Category:    "api-observability",
		Severity:    "high",
		Description: "Empty catch block swallows exception silently",
		Remediation: "Log the error or re-throw if unhandled",
	},
	{
		Name:        "Generic Error Response",
		Pattern:     regexp.MustCompile(`res\.status\s*\(\s*500\s*\)\.(?:json|send)\s*\(\s*['"](?:error|Error|Something went wrong)['"]`),
		Category:    "api-observability",
		Severity:    "low",
		Description: "Generic error message provides no debugging context",
		Remediation: "Include error code and message for client debugging",
	},
	{
		Name:        "Express App Setup",
		Pattern:     regexp.MustCompile(`(?:express|app)\s*\(\s*\)`),
		Category:    "api-observability",
		Severity:    "info",
		Description: "Express app setup - verify request ID middleware is configured for tracing",
		Remediation: "Add request ID middleware for distributed tracing",
	},

	// API Rate Limiting Patterns
	{
		Name:        "Public Endpoint Without Rate Limiting",
		Pattern:     regexp.MustCompile(`\.(get|post|put|delete)\s*\(\s*['"]/(?:api|v[0-9]+)/(?:auth|login|register|signup|password|reset|verify|token)[^'"]*['"]`),
		Category:    "api-rate-limiting",
		Severity:    "high",
		Description: "Authentication endpoint detected without apparent rate limiting",
		Remediation: "Add rate limiting middleware: app.use('/api/auth', rateLimit({ windowMs: 15*60*1000, max: 5 }))",
	},
	{
		Name:        "File Upload Without Rate Limiting",
		Pattern:     regexp.MustCompile(`(?:multer|upload|formidable|busboy)\s*\(`),
		Category:    "api-rate-limiting",
		Severity:    "medium",
		Description: "File upload detected - verify rate limiting and file size limits are configured",
		Remediation: "Add rate limiting and maxFileSize limits to prevent resource exhaustion",
	},
	{
		Name:        "Search Endpoint",
		Pattern:     regexp.MustCompile(`\.(get|post)\s*\(\s*['"]/(?:api/)?(?:search|query|find|lookup)[^'"]*['"]`),
		Category:    "api-rate-limiting",
		Severity:    "medium",
		Description: "Search endpoint detected - verify rate limiting to prevent abuse",
		Remediation: "Add rate limiting for search endpoints to prevent scraping and DoS",
	},
	{
		Name:        "Webhook Endpoint",
		Pattern:     regexp.MustCompile(`\.(post)\s*\(\s*['"]/(?:api/)?(?:webhook|hook|callback|notify)[^'"]*['"]`),
		Category:    "api-rate-limiting",
		Severity:    "medium",
		Description: "Webhook endpoint detected - verify rate limiting and signature validation",
		Remediation: "Rate limit webhooks and verify request signatures",
	},
	{
		Name:        "Email Sending Endpoint",
		Pattern:     regexp.MustCompile(`(?:sendMail|sendEmail|nodemailer|sendgrid|mailgun|ses\.send)`),
		Category:    "api-rate-limiting",
		Severity:    "high",
		Description: "Email sending detected - verify rate limiting to prevent spam abuse",
		Remediation: "Add strict rate limiting for email sending endpoints",
	},
	{
		Name:        "Express Rate Limit Middleware",
		Pattern:     regexp.MustCompile(`(?:rateLimit|RateLimiter|express-rate-limit|rate-limiter-flexible)\s*\(`),
		Category:    "api-rate-limiting",
		Severity:    "info",
		Description: "Rate limiting middleware detected - good security practice",
		Remediation: "Verify rate limits are appropriate for your use case",
	},
	{
		Name:        "Go Rate Limiter",
		Pattern:     regexp.MustCompile(`(?:rate\.NewLimiter|limiter\.New|tollbooth|throttle)\s*\(`),
		Category:    "api-rate-limiting",
		Severity:    "info",
		Description: "Rate limiting detected in Go code - good security practice",
		Remediation: "Verify rate limits are appropriate for your use case",
	},
	{
		Name:        "Python Rate Limiter",
		Pattern:     regexp.MustCompile(`(?:flask_limiter|Limiter|ratelimit|slowapi|RateLimitMiddleware)`),
		Category:    "api-rate-limiting",
		Severity:    "info",
		Description: "Rate limiting detected in Python code - good security practice",
		Remediation: "Verify rate limits are appropriate for your use case",
	},

	// API Documentation Patterns
	{
		Name:        "Route Definition",
		Pattern:     regexp.MustCompile(`\.(get|post|put|delete|patch)\s*\(\s*['"][^'"]+['"]`),
		Category:    "api-documentation",
		Severity:    "info",
		Description: "API route found - verify documentation exists",
		Remediation: "Add JSDoc or Swagger annotations for API documentation",
	},
	{
		Name:        "Swagger Response Annotation",
		Pattern:     regexp.MustCompile(`@swagger[^@]*@response`),
		Category:    "api-documentation",
		Severity:    "info",
		Description: "Swagger response found - verify schema is included",
		Remediation: "Add response schema to @swagger annotation",
	},
}

// runAPIQualityChecks performs non-security API quality analysis
func (s *CodeSecurityScanner) runAPIQualityChecks(ctx context.Context, opts *scanner.ScanOptions, cfg APIConfig) []APIFinding {
	var findings []APIFinding

	// Only check files with likely API code
	fileExtensions := []string{".js", ".ts", ".py", ".go", ".java", ".rb", ".php"}

	for _, ext := range fileExtensions {
		err := filepath.Walk(opts.RepoPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				if shouldSkipQualityDir(info.Name()) {
					return filepath.SkipDir
				}
				return nil
			}

			if !strings.HasSuffix(path, ext) {
				return nil
			}

			fileFindings := s.scanFileForQualityPatterns(path, opts.RepoPath, cfg)
			findings = append(findings, fileFindings...)
			return nil
		})
		if err != nil {
			continue
		}
	}

	return findings
}

// shouldSkipQualityDir checks if a directory should be skipped for quality checks
func shouldSkipQualityDir(name string) bool {
	skipDirs := []string{
		"node_modules", "vendor", ".git", "dist", "build",
		"coverage", "__pycache__", ".venv", "venv",
		"test", "tests", "__tests__", "spec", "specs",
	}
	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}
	return false
}

// scanFileForQualityPatterns scans a file for API quality patterns
func (s *CodeSecurityScanner) scanFileForQualityPatterns(filePath, repoPath string, cfg APIConfig) []APIFinding {
	var findings []APIFinding

	// Check if file looks like an API route/controller file
	if !isLikelyAPIFile(filePath) {
		return findings
	}

	file, err := os.Open(filePath)
	if err != nil {
		return findings
	}
	defer file.Close()

	relPath := strings.TrimPrefix(filePath, repoPath)
	relPath = strings.TrimPrefix(relPath, "/")

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, p := range apiQualityPatterns {
			// Filter by category based on config
			if !shouldCheckQualityCategory(p.Category, cfg) {
				continue
			}

			if p.Pattern.MatchString(line) {
				finding := APIFinding{
					RuleID:      "api-quality-" + strings.ToLower(strings.ReplaceAll(p.Name, " ", "-")),
					Title:       p.Name,
					Description: p.Description,
					Severity:    p.Severity,
					Confidence:  "medium",
					File:        relPath,
					Line:        lineNum,
					Snippet:     truncateSnippet(line, 200),
					Category:    p.Category,
					Remediation: p.Remediation,
				}

				// Try to extract endpoint and method
				if endpoint := extractEndpoint(line); endpoint != "" {
					finding.Endpoint = endpoint
				}
				if method := extractHTTPMethod(line); method != "" {
					finding.HTTPMethod = method
				}

				findings = append(findings, finding)
			}
		}
	}

	return findings
}

// isLikelyAPIFile checks if a file is likely to contain API routes
func isLikelyAPIFile(path string) bool {
	pathLower := strings.ToLower(path)

	// Positive indicators
	positivePatterns := []string{
		"route", "controller", "handler", "api", "endpoint",
		"server", "app", "router", "rest", "graphql",
	}
	for _, pattern := range positivePatterns {
		if strings.Contains(pathLower, pattern) {
			return true
		}
	}

	// Check common API file patterns
	filename := filepath.Base(pathLower)
	apiFilePatterns := []string{
		"index.js", "index.ts", "app.js", "app.ts",
		"server.js", "server.ts", "main.py", "app.py",
		"main.go", "handlers.go", "routes.go",
	}
	for _, pattern := range apiFilePatterns {
		if filename == pattern {
			return true
		}
	}

	return false
}

// shouldCheckQualityCategory determines if a quality category should be checked
func shouldCheckQualityCategory(category string, cfg APIConfig) bool {
	switch category {
	case "api-design":
		return cfg.CheckDesign
	case "api-performance":
		return cfg.CheckPerformance
	case "api-observability":
		return cfg.CheckObservability
	case "api-documentation":
		return cfg.CheckDocumentation
	default:
		return true
	}
}
