// Package common provides shared utilities for scanners
// This file provides microservice communication detection and dependency mapping
package common

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// MicroserviceScanner detects service-to-service communication patterns
type MicroserviceScanner struct {
	ragPath  string
	cacheDir string
	timeout  time.Duration
	onStatus func(string)
}

// MicroserviceConfig configures the microservice scanner
type MicroserviceConfig struct {
	RAGPath  string
	CacheDir string
	Timeout  time.Duration
	OnStatus func(string)
}

// MicroserviceResult holds the scan results
type MicroserviceResult struct {
	Services      []ServiceDefinition   `json:"services"`
	Dependencies  []ServiceDependency   `json:"dependencies"`
	APIContracts  []APIContract         `json:"api_contracts"`
	MessageQueues []MessageQueueUsage   `json:"message_queues"`
	Summary       MicroserviceSummary   `json:"summary"`
	Error         error                 `json:"-"`
}

// ServiceDefinition represents a service defined in this codebase
type ServiceDefinition struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`        // http, grpc, graphql
	Endpoints   []Endpoint        `json:"endpoints"`
	Port        string            `json:"port,omitempty"`
	File        string            `json:"file,omitempty"`
	Line        int               `json:"line,omitempty"`
	Framework   string            `json:"framework,omitempty"` // express, fastapi, gin, spring
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Method      string `json:"method,omitempty"` // GET, POST, PUT, DELETE
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
}

// ServiceDependency represents a dependency on another service
type ServiceDependency struct {
	SourceService string   `json:"source_service,omitempty"` // Service making the call
	TargetService string   `json:"target_service"`            // Service being called
	TargetURL     string   `json:"target_url,omitempty"`
	Type          string   `json:"type"`                      // http, grpc, message_queue
	Method        string   `json:"method,omitempty"`          // HTTP method or gRPC method
	Locations     []CodeLocation `json:"locations"`
	Confidence    int      `json:"confidence"`                // 0-100
}

// CodeLocation represents where something is found in code
type CodeLocation struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Snippet string `json:"snippet,omitempty"`
}

// APIContract represents an API contract definition
type APIContract struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"`        // openapi, graphql, protobuf
	Version     string     `json:"version,omitempty"`
	File        string     `json:"file"`
	BaseURL     string     `json:"base_url,omitempty"`
	Endpoints   []Endpoint `json:"endpoints,omitempty"`
	Services    []string   `json:"services,omitempty"` // Services defined in this contract
}

// MessageQueueUsage represents message queue producer/consumer
type MessageQueueUsage struct {
	QueueType    string       `json:"queue_type"`   // kafka, rabbitmq, sqs, pubsub, nats
	Role         string       `json:"role"`         // producer, consumer
	TopicOrQueue string       `json:"topic_or_queue"`
	Brokers      []string     `json:"brokers,omitempty"`
	ConsumerGroup string      `json:"consumer_group,omitempty"`
	Locations    []CodeLocation `json:"locations"`
}

// MicroserviceSummary provides summary statistics
type MicroserviceSummary struct {
	TotalServices       int            `json:"total_services"`
	TotalDependencies   int            `json:"total_dependencies"`
	TotalAPIContracts   int            `json:"total_api_contracts"`
	TotalMessageQueues  int            `json:"total_message_queues"`
	CommunicationTypes  map[string]int `json:"communication_types"`  // http, grpc, kafka, etc.
	DependencyGraph     map[string][]string `json:"dependency_graph"` // service -> [dependencies]
}

// NewMicroserviceScanner creates a new microservice scanner
func NewMicroserviceScanner(cfg MicroserviceConfig) *MicroserviceScanner {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}
	if cfg.OnStatus == nil {
		cfg.OnStatus = func(string) {}
	}
	if cfg.RAGPath == "" {
		cfg.RAGPath = findRAGPath()
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = getCacheDir()
	}

	return &MicroserviceScanner{
		ragPath:  cfg.RAGPath,
		cacheDir: cfg.CacheDir,
		timeout:  cfg.Timeout,
		onStatus: cfg.OnStatus,
	}
}

// Scan scans a repository for microservice communication patterns
func (s *MicroserviceScanner) Scan(ctx context.Context, repoPath string) *MicroserviceResult {
	result := &MicroserviceResult{
		Summary: MicroserviceSummary{
			CommunicationTypes: make(map[string]int),
			DependencyGraph:    make(map[string][]string),
		},
	}

	s.onStatus("Scanning for microservice patterns...")

	// 1. Detect API contracts (OpenAPI, GraphQL, Proto files)
	contracts := s.detectAPIContracts(repoPath)
	result.APIContracts = contracts

	// 2. Detect HTTP client calls
	httpDeps := s.detectHTTPClients(ctx, repoPath)
	result.Dependencies = append(result.Dependencies, httpDeps...)

	// 3. Detect gRPC communication
	grpcDeps := s.detectGRPCCommunication(ctx, repoPath)
	result.Dependencies = append(result.Dependencies, grpcDeps...)

	// 4. Detect message queue usage
	mqUsage := s.detectMessageQueues(ctx, repoPath)
	result.MessageQueues = mqUsage

	// 5. Detect service definitions
	services := s.detectServiceDefinitions(repoPath)
	result.Services = services

	// 6. Build summary
	s.buildSummary(result)

	return result
}

// detectAPIContracts finds OpenAPI, GraphQL, and Proto files
func (s *MicroserviceScanner) detectAPIContracts(repoPath string) []APIContract {
	var contracts []APIContract

	// Find OpenAPI specs
	openAPIFiles := s.findFiles(repoPath, []string{
		"**/openapi.yaml", "**/openapi.yml", "**/openapi.json",
		"**/swagger.yaml", "**/swagger.yml", "**/swagger.json",
		"**/api.yaml", "**/api.yml", "**/api.json",
	})

	for _, file := range openAPIFiles {
		if contract := s.parseOpenAPISpec(file); contract != nil {
			contracts = append(contracts, *contract)
		}
	}

	// Find GraphQL schemas
	graphqlFiles := s.findFiles(repoPath, []string{
		"**/*.graphql", "**/schema.graphql", "**/schema.gql",
	})

	for _, file := range graphqlFiles {
		if contract := s.parseGraphQLSchema(file); contract != nil {
			contracts = append(contracts, *contract)
		}
	}

	// Find Proto files
	protoFiles := s.findFiles(repoPath, []string{
		"**/*.proto",
	})

	for _, file := range protoFiles {
		if contract := s.parseProtoFile(file); contract != nil {
			contracts = append(contracts, *contract)
		}
	}

	return contracts
}

// parseOpenAPISpec parses an OpenAPI specification file
func (s *MicroserviceScanner) parseOpenAPISpec(filePath string) *APIContract {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	content := string(data)

	// Check if it's OpenAPI 3.x or Swagger 2.0
	var version, specType string
	if strings.Contains(content, "openapi:") || strings.Contains(content, `"openapi"`) {
		specType = "openapi"
		versionRe := regexp.MustCompile(`["']?openapi["']?\s*:\s*["']?([\d.]+)["']?`)
		if m := versionRe.FindStringSubmatch(content); len(m) > 1 {
			version = m[1]
		}
	} else if strings.Contains(content, "swagger:") || strings.Contains(content, `"swagger"`) {
		specType = "openapi"
		version = "2.0"
	} else {
		return nil
	}

	// Extract title
	var name string
	titleRe := regexp.MustCompile(`(?:info:\s*\n\s*title:\s*["']?([^"'\n]+)["']?|"title"\s*:\s*"([^"]+)")`)
	if m := titleRe.FindStringSubmatch(content); len(m) > 1 {
		if m[1] != "" {
			name = m[1]
		} else {
			name = m[2]
		}
	} else {
		// Use filename as name
		name = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	// Extract base URL
	var baseURL string
	serverRe := regexp.MustCompile(`(?:servers:\s*\n\s*-\s*url:\s*["']?([^"'\n]+)["']?|"basePath"\s*:\s*"([^"]+)")`)
	if m := serverRe.FindStringSubmatch(content); len(m) > 1 {
		if m[1] != "" {
			baseURL = m[1]
		} else {
			baseURL = m[2]
		}
	}

	// Extract paths/endpoints
	var endpoints []Endpoint
	pathRe := regexp.MustCompile(`(?m)^\s*["']?(/[^"'\s:]+)["']?\s*:`)
	for _, m := range pathRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 && !strings.HasPrefix(m[1], "/components") && !strings.HasPrefix(m[1], "/definitions") {
			endpoints = append(endpoints, Endpoint{Path: m[1]})
		}
	}

	relPath, _ := filepath.Rel(s.cacheDir, filePath)
	if relPath == "" {
		relPath = filePath
	}

	return &APIContract{
		Name:      name,
		Type:      specType,
		Version:   version,
		File:      relPath,
		BaseURL:   baseURL,
		Endpoints: endpoints,
	}
}

// parseGraphQLSchema parses a GraphQL schema file
func (s *MicroserviceScanner) parseGraphQLSchema(filePath string) *APIContract {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	content := string(data)

	// Extract type names
	var services []string
	typeRe := regexp.MustCompile(`type\s+(\w+)\s*(?:@\w+[^{]*)*\{`)
	for _, m := range typeRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 && m[1] != "Query" && m[1] != "Mutation" && m[1] != "Subscription" {
			services = append(services, m[1])
		}
	}

	// Extract query/mutation fields as endpoints
	var endpoints []Endpoint
	queryRe := regexp.MustCompile(`(?s)type\s+Query\s*\{([^}]+)\}`)
	if m := queryRe.FindStringSubmatch(content); len(m) > 1 {
		fieldRe := regexp.MustCompile(`(\w+)\s*(?:\([^)]*\))?\s*:\s*(\w+)`)
		for _, f := range fieldRe.FindAllStringSubmatch(m[1], -1) {
			if len(f) > 1 {
				endpoints = append(endpoints, Endpoint{
					Method: "query",
					Path:   f[1],
				})
			}
		}
	}

	mutationRe := regexp.MustCompile(`(?s)type\s+Mutation\s*\{([^}]+)\}`)
	if m := mutationRe.FindStringSubmatch(content); len(m) > 1 {
		fieldRe := regexp.MustCompile(`(\w+)\s*(?:\([^)]*\))?\s*:\s*(\w+)`)
		for _, f := range fieldRe.FindAllStringSubmatch(m[1], -1) {
			if len(f) > 1 {
				endpoints = append(endpoints, Endpoint{
					Method: "mutation",
					Path:   f[1],
				})
			}
		}
	}

	name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	return &APIContract{
		Name:      name,
		Type:      "graphql",
		File:      filePath,
		Endpoints: endpoints,
		Services:  services,
	}
}

// parseProtoFile parses a Protocol Buffers file
func (s *MicroserviceScanner) parseProtoFile(filePath string) *APIContract {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	content := string(data)

	// Extract service definitions
	var services []string
	var endpoints []Endpoint

	serviceRe := regexp.MustCompile(`service\s+(\w+)\s*\{`)
	rpcRe := regexp.MustCompile(`rpc\s+(\w+)\s*\(\s*(\w+)\s*\)\s*returns\s*\(\s*(\w+)\s*\)`)

	for _, m := range serviceRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 {
			services = append(services, m[1])
		}
	}

	for _, m := range rpcRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 {
			endpoints = append(endpoints, Endpoint{
				Method:      "rpc",
				Path:        m[1],
				Description: m[2] + " -> " + m[3],
			})
		}
	}

	// Extract package name
	var name string
	pkgRe := regexp.MustCompile(`package\s+([\w.]+)\s*;`)
	if m := pkgRe.FindStringSubmatch(content); len(m) > 1 {
		name = m[1]
	} else {
		name = strings.TrimSuffix(filepath.Base(filePath), ".proto")
	}

	return &APIContract{
		Name:      name,
		Type:      "protobuf",
		File:      filePath,
		Endpoints: endpoints,
		Services:  services,
	}
}

// detectHTTPClients detects HTTP client usage patterns
func (s *MicroserviceScanner) detectHTTPClients(ctx context.Context, repoPath string) []ServiceDependency {
	var deps []ServiceDependency

	// URL patterns to extract service names
	patterns := []struct {
		pattern *regexp.Regexp
		lang    string
	}{
		// JavaScript/TypeScript patterns
		{regexp.MustCompile(`fetch\s*\(\s*["'\x60]([^"'\x60]+)["'\x60]`), "javascript"},
		{regexp.MustCompile(`axios\.(?:get|post|put|delete|patch)\s*\(\s*["'\x60]([^"'\x60]+)["'\x60]`), "javascript"},
		{regexp.MustCompile(`got\s*\(\s*["'\x60]([^"'\x60]+)["'\x60]`), "javascript"},

		// Python patterns
		{regexp.MustCompile(`requests\.(?:get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`), "python"},
		{regexp.MustCompile(`httpx\.(?:get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`), "python"},
		{regexp.MustCompile(`aiohttp\.\w+\s*\(\s*["']([^"']+)["']`), "python"},

		// Go patterns
		{regexp.MustCompile(`http\.(?:Get|Post|Put|Delete)\s*\(\s*["']([^"']+)["']`), "go"},
		{regexp.MustCompile(`http\.NewRequest\s*\(\s*["']\w+["']\s*,\s*["']([^"']+)["']`), "go"},
	}

	// File extensions to scan
	extensions := map[string][]string{
		"javascript": {".js", ".ts", ".jsx", ".tsx", ".mjs"},
		"python":     {".py"},
		"go":         {".go"},
	}

	seen := make(map[string]bool)

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip common non-source directories
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == "vendor" || name == ".git" ||
				name == "__pycache__" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		for lang, exts := range extensions {
			for _, e := range exts {
				if ext == e {
					data, err := os.ReadFile(path)
					if err != nil {
						continue
					}
					content := string(data)

					for _, p := range patterns {
						if p.lang != lang {
							continue
						}

						for _, m := range p.pattern.FindAllStringSubmatch(content, -1) {
							if len(m) > 1 {
								url := m[1]
								serviceName := extractServiceName(url)

								// Deduplicate
								key := serviceName + "|" + url
								if seen[key] {
									continue
								}
								seen[key] = true

								if serviceName != "" && !isExternalURL(url) {
									relPath, _ := filepath.Rel(repoPath, path)
									deps = append(deps, ServiceDependency{
										TargetService: serviceName,
										TargetURL:     url,
										Type:          "http",
										Confidence:    80,
										Locations: []CodeLocation{
											{File: relPath, Snippet: truncate(m[0], 100)},
										},
									})
								}
							}
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return deps
	}

	return deps
}

// detectGRPCCommunication detects gRPC client/server patterns
func (s *MicroserviceScanner) detectGRPCCommunication(ctx context.Context, repoPath string) []ServiceDependency {
	var deps []ServiceDependency

	patterns := []struct {
		pattern *regexp.Regexp
		lang    string
	}{
		// Python gRPC
		{regexp.MustCompile(`grpc\.(?:insecure_channel|secure_channel)\s*\(\s*["']([^"']+)["']`), "python"},
		{regexp.MustCompile(`(\w+)_pb2_grpc\.\w+Stub\s*\(`), "python"},

		// Go gRPC
		{regexp.MustCompile(`grpc\.(?:Dial|NewClient)\s*\(\s*["']([^"']+)["']`), "go"},
		{regexp.MustCompile(`(\w+)\.New(\w+)Client\s*\(`), "go"},

		// Java gRPC
		{regexp.MustCompile(`ManagedChannelBuilder\.for(?:Address|Target)\s*\(\s*["']([^"']+)["']`), "java"},
	}

	extensions := map[string][]string{
		"python": {".py"},
		"go":     {".go"},
		"java":   {".java"},
	}

	seen := make(map[string]bool)

	_ = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		for lang, exts := range extensions {
			for _, e := range exts {
				if ext == e {
					data, _ := os.ReadFile(path)
					content := string(data)

					for _, p := range patterns {
						if p.lang != lang {
							continue
						}

						for _, m := range p.pattern.FindAllStringSubmatch(content, -1) {
							if len(m) > 1 {
								target := m[1]
								serviceName := extractServiceName(target)

								key := serviceName + "|grpc"
								if seen[key] {
									continue
								}
								seen[key] = true

								if serviceName != "" {
									relPath, _ := filepath.Rel(repoPath, path)
									deps = append(deps, ServiceDependency{
										TargetService: serviceName,
										TargetURL:     target,
										Type:          "grpc",
										Confidence:    85,
										Locations: []CodeLocation{
											{File: relPath, Snippet: truncate(m[0], 100)},
										},
									})
								}
							}
						}
					}
				}
			}
		}
		return nil
	})

	return deps
}

// detectMessageQueues detects message queue producers/consumers
func (s *MicroserviceScanner) detectMessageQueues(ctx context.Context, repoPath string) []MessageQueueUsage {
	var queues []MessageQueueUsage

	patterns := []struct {
		pattern   *regexp.Regexp
		queueType string
		role      string
	}{
		// Kafka
		{regexp.MustCompile(`KafkaProducer\s*\(`), "kafka", "producer"},
		{regexp.MustCompile(`KafkaConsumer\s*\(\s*["']([^"']+)["']`), "kafka", "consumer"},
		{regexp.MustCompile(`\.send\s*\(\s*["']([^"']+)["']`), "kafka", "producer"},
		{regexp.MustCompile(`@KafkaListener\s*\(\s*topics\s*=\s*["']?([^"')\s,]+)`), "kafka", "consumer"},

		// RabbitMQ
		{regexp.MustCompile(`basic_publish\s*\([^)]*routing_key\s*=\s*["']([^"']+)["']`), "rabbitmq", "producer"},
		{regexp.MustCompile(`basic_consume\s*\([^)]*queue\s*=\s*["']([^"']+)["']`), "rabbitmq", "consumer"},
		{regexp.MustCompile(`@RabbitListener\s*\(\s*queues\s*=\s*["']?([^"')\s,]+)`), "rabbitmq", "consumer"},
		{regexp.MustCompile(`sendToQueue\s*\(\s*["']([^"']+)["']`), "rabbitmq", "producer"},

		// AWS SQS
		{regexp.MustCompile(`send_message\s*\([^)]*QueueUrl`), "sqs", "producer"},
		{regexp.MustCompile(`receive_message\s*\([^)]*QueueUrl`), "sqs", "consumer"},

		// Redis Pub/Sub
		{regexp.MustCompile(`\.publish\s*\(\s*["']([^"']+)["']`), "redis", "producer"},
		{regexp.MustCompile(`\.subscribe\s*\(\s*["']([^"']+)["']`), "redis", "consumer"},

		// NATS
		{regexp.MustCompile(`\.Publish\s*\(\s*["']([^"']+)["']`), "nats", "producer"},
		{regexp.MustCompile(`\.Subscribe\s*\(\s*["']([^"']+)["']`), "nats", "consumer"},
	}

	seen := make(map[string]bool)

	_ = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".py" && ext != ".go" && ext != ".java" && ext != ".js" && ext != ".ts" {
			return nil
		}

		data, _ := os.ReadFile(path)
		content := string(data)

		for _, p := range patterns {
			for _, m := range p.pattern.FindAllStringSubmatch(content, -1) {
				topic := ""
				if len(m) > 1 {
					topic = m[1]
				}

				key := p.queueType + "|" + p.role + "|" + topic
				if seen[key] {
					continue
				}
				seen[key] = true

				relPath, _ := filepath.Rel(repoPath, path)
				queues = append(queues, MessageQueueUsage{
					QueueType:    p.queueType,
					Role:         p.role,
					TopicOrQueue: topic,
					Locations: []CodeLocation{
						{File: relPath, Snippet: truncate(m[0], 80)},
					},
				})
			}
		}
		return nil
	})

	return queues
}

// detectServiceDefinitions detects service/server definitions
func (s *MicroserviceScanner) detectServiceDefinitions(repoPath string) []ServiceDefinition {
	var services []ServiceDefinition

	// Detect from docker-compose.yml
	composeFiles := s.findFiles(repoPath, []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"})
	for _, file := range composeFiles {
		svc := s.parseDockerCompose(file)
		services = append(services, svc...)
	}

	// Detect from Kubernetes manifests
	k8sFiles := s.findFiles(repoPath, []string{"**/k8s/*.yaml", "**/kubernetes/*.yaml", "**/deployment.yaml", "**/service.yaml"})
	for _, file := range k8sFiles {
		svc := s.parseK8sService(file)
		if svc != nil {
			services = append(services, *svc)
		}
	}

	return services
}

// parseDockerCompose extracts services from docker-compose.yml
func (s *MicroserviceScanner) parseDockerCompose(filePath string) []ServiceDefinition {
	var services []ServiceDefinition

	data, err := os.ReadFile(filePath)
	if err != nil {
		return services
	}

	content := string(data)

	// Simple regex to extract service names
	// A more robust solution would use a YAML parser
	servicesRe := regexp.MustCompile(`(?m)^services:\s*\n((?:\s+\w+:\s*\n(?:\s+.+\n)*)+)`)
	if m := servicesRe.FindStringSubmatch(content); len(m) > 1 {
		serviceBlock := m[1]
		svcNameRe := regexp.MustCompile(`(?m)^\s{2}(\w+):\s*$`)
		for _, sm := range svcNameRe.FindAllStringSubmatch(serviceBlock, -1) {
			if len(sm) > 1 {
				services = append(services, ServiceDefinition{
					Name: sm[1],
					Type: "container",
					File: filePath,
				})
			}
		}
	}

	return services
}

// parseK8sService extracts service from Kubernetes manifest
func (s *MicroserviceScanner) parseK8sService(filePath string) *ServiceDefinition {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	content := string(data)

	// Check if it's a Service resource
	if !strings.Contains(content, "kind: Service") {
		return nil
	}

	// Extract service name
	nameRe := regexp.MustCompile(`(?m)^\s*name:\s*["']?([^"'\n]+)["']?`)
	var name string
	if m := nameRe.FindStringSubmatch(content); len(m) > 1 {
		name = m[1]
	}

	// Extract port
	portRe := regexp.MustCompile(`(?m)^\s*port:\s*(\d+)`)
	var port string
	if m := portRe.FindStringSubmatch(content); len(m) > 1 {
		port = m[1]
	}

	if name == "" {
		return nil
	}

	return &ServiceDefinition{
		Name: name,
		Type: "kubernetes",
		Port: port,
		File: filePath,
	}
}

// findFiles finds files matching glob patterns
func (s *MicroserviceScanner) findFiles(repoPath string, patterns []string) []string {
	var files []string

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(repoPath, pattern))
		files = append(files, matches...)
	}

	return files
}

// buildSummary builds the summary statistics
func (s *MicroserviceScanner) buildSummary(result *MicroserviceResult) {
	result.Summary.TotalServices = len(result.Services)
	result.Summary.TotalDependencies = len(result.Dependencies)
	result.Summary.TotalAPIContracts = len(result.APIContracts)
	result.Summary.TotalMessageQueues = len(result.MessageQueues)

	// Count communication types
	for _, dep := range result.Dependencies {
		result.Summary.CommunicationTypes[dep.Type]++
	}
	for _, mq := range result.MessageQueues {
		result.Summary.CommunicationTypes[mq.QueueType]++
	}

	// Build dependency graph
	for _, dep := range result.Dependencies {
		source := dep.SourceService
		if source == "" {
			source = "this-service"
		}
		result.Summary.DependencyGraph[source] = append(result.Summary.DependencyGraph[source], dep.TargetService)
	}
}

// Helper functions

// extractServiceName extracts a service name from a URL
func extractServiceName(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "dns:///")

	// Extract host part
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}
	host := parts[0]

	// Remove port
	hostParts := strings.Split(host, ":")
	host = hostParts[0]

	// Skip localhost and IP addresses
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" {
		return ""
	}

	// Check if it looks like a service name (not a domain)
	if strings.Contains(host, ".com") || strings.Contains(host, ".org") ||
		strings.Contains(host, ".io") || strings.Contains(host, ".net") {
		// It's likely an external domain, not internal service
		// But check for kubernetes DNS patterns
		if !strings.Contains(host, ".svc.cluster.local") &&
			!strings.Contains(host, ".default.") &&
			!strings.HasSuffix(host, "-service") &&
			!strings.HasSuffix(host, "-svc") {
			return ""
		}
	}

	return host
}

// isExternalURL checks if URL is external (not internal service)
func isExternalURL(url string) bool {
	externalDomains := []string{
		"googleapis.com", "github.com", "gitlab.com", "bitbucket.org",
		"amazonaws.com", "azure.com", "cloudflare.com",
		"stripe.com", "twilio.com", "sendgrid.com",
		"facebook.com", "twitter.com", "linkedin.com",
	}

	for _, domain := range externalDomains {
		if strings.Contains(url, domain) {
			return true
		}
	}

	return false
}

// truncate truncates a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ToJSON exports result as JSON
func (r *MicroserviceResult) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
