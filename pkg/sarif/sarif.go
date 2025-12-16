// Package sarif provides SARIF (Static Analysis Results Interchange Format) export
// SARIF is a standard format for the output of static analysis tools.
// Specification: https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
package sarif

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// SARIFVersion is the SARIF specification version
	SARIFVersion = "2.1.0"
	// SchemaURI is the JSON schema URI for SARIF 2.1.0
	SchemaURI = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"
)

// Log is the root SARIF object
type Log struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

// Run represents a single run of an analysis tool
type Run struct {
	Tool        Tool         `json:"tool"`
	Results     []Result     `json:"results,omitempty"`
	Invocations []Invocation `json:"invocations,omitempty"`
	Artifacts   []Artifact   `json:"artifacts,omitempty"`
}

// Tool describes the analysis tool
type Tool struct {
	Driver ToolComponent `json:"driver"`
}

// ToolComponent describes a tool component (driver or extension)
type ToolComponent struct {
	Name            string `json:"name"`
	Version         string `json:"version,omitempty"`
	InformationURI  string `json:"informationUri,omitempty"`
	Rules           []Rule `json:"rules,omitempty"`
	SemanticVersion string `json:"semanticVersion,omitempty"`
}

// Rule describes a rule used by the tool
type Rule struct {
	ID               string            `json:"id"`
	Name             string            `json:"name,omitempty"`
	ShortDescription *Message          `json:"shortDescription,omitempty"`
	FullDescription  *Message          `json:"fullDescription,omitempty"`
	HelpURI          string            `json:"helpUri,omitempty"`
	Help             *Message          `json:"help,omitempty"`
	DefaultConfig    *ReportingConfig  `json:"defaultConfiguration,omitempty"`
	Properties       map[string]string `json:"properties,omitempty"`
}

// ReportingConfig contains default configuration for a rule
type ReportingConfig struct {
	Level string `json:"level,omitempty"` // "none", "note", "warning", "error"
}

// Result represents a single finding
type Result struct {
	RuleID    string     `json:"ruleId"`
	RuleIndex int        `json:"ruleIndex,omitempty"`
	Level     string     `json:"level,omitempty"` // "none", "note", "warning", "error"
	Message   Message    `json:"message"`
	Locations []Location `json:"locations,omitempty"`
	Fixes     []Fix      `json:"fixes,omitempty"`
	// Additional properties
	PartialFingerprints map[string]string `json:"partialFingerprints,omitempty"`
	Properties          map[string]any    `json:"properties,omitempty"`
}

// Message contains text content
type Message struct {
	Text     string `json:"text,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

// Location identifies a location in source code
type Location struct {
	PhysicalLocation *PhysicalLocation `json:"physicalLocation,omitempty"`
	LogicalLocations []LogicalLocation `json:"logicalLocations,omitempty"`
}

// PhysicalLocation identifies a physical location in a file
type PhysicalLocation struct {
	ArtifactLocation *ArtifactLocation `json:"artifactLocation,omitempty"`
	Region           *Region           `json:"region,omitempty"`
}

// ArtifactLocation identifies an artifact (file)
type ArtifactLocation struct {
	URI       string `json:"uri,omitempty"`
	URIBaseID string `json:"uriBaseId,omitempty"`
}

// Region identifies a region within a file
type Region struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

// LogicalLocation identifies a logical location (function, class, etc.)
type LogicalLocation struct {
	Name               string `json:"name,omitempty"`
	FullyQualifiedName string `json:"fullyQualifiedName,omitempty"`
	Kind               string `json:"kind,omitempty"` // "function", "member", "module", etc.
}

// Artifact describes an artifact (file) that was analyzed
type Artifact struct {
	Location *ArtifactLocation `json:"location,omitempty"`
	Length   int               `json:"length,omitempty"`
	MimeType string            `json:"mimeType,omitempty"`
}

// Invocation describes a single invocation of the tool
type Invocation struct {
	CommandLine        string    `json:"commandLine,omitempty"`
	StartTimeUtc       time.Time `json:"startTimeUtc,omitempty"`
	EndTimeUtc         time.Time `json:"endTimeUtc,omitempty"`
	ExecutionSuccessful bool     `json:"executionSuccessful"`
	WorkingDirectory   *ArtifactLocation `json:"workingDirectory,omitempty"`
}

// Fix describes a proposed fix for a result
type Fix struct {
	Description     *Message          `json:"description,omitempty"`
	ArtifactChanges []ArtifactChange  `json:"artifactChanges,omitempty"`
}

// ArtifactChange describes a change to an artifact
type ArtifactChange struct {
	ArtifactLocation *ArtifactLocation `json:"artifactLocation,omitempty"`
	Replacements     []Replacement     `json:"replacements,omitempty"`
}

// Replacement describes a text replacement
type Replacement struct {
	DeletedRegion   *Region `json:"deletedRegion,omitempty"`
	InsertedContent *ArtifactContent `json:"insertedContent,omitempty"`
}

// ArtifactContent holds artifact content
type ArtifactContent struct {
	Text string `json:"text,omitempty"`
}

// NewLog creates a new SARIF log
func NewLog() *Log {
	return &Log{
		Schema:  SchemaURI,
		Version: SARIFVersion,
		Runs:    []Run{},
	}
}

// NewRun creates a new run for a tool
func NewRun(toolName, toolVersion, infoURI string) *Run {
	return &Run{
		Tool: Tool{
			Driver: ToolComponent{
				Name:           toolName,
				Version:        toolVersion,
				InformationURI: infoURI,
				Rules:          []Rule{},
			},
		},
		Results: []Result{},
	}
}

// AddRule adds a rule to the run's tool
func (r *Run) AddRule(id, name, description, helpURI string, level string) int {
	rule := Rule{
		ID:   id,
		Name: name,
		ShortDescription: &Message{
			Text: description,
		},
		HelpURI: helpURI,
		DefaultConfig: &ReportingConfig{
			Level: level,
		},
	}
	r.Tool.Driver.Rules = append(r.Tool.Driver.Rules, rule)
	return len(r.Tool.Driver.Rules) - 1
}

// AddResult adds a result to the run
func (r *Run) AddResult(ruleID string, ruleIndex int, level, message, file string, line int) {
	result := Result{
		RuleID:    ruleID,
		RuleIndex: ruleIndex,
		Level:     level,
		Message: Message{
			Text: message,
		},
	}

	if file != "" {
		loc := Location{
			PhysicalLocation: &PhysicalLocation{
				ArtifactLocation: &ArtifactLocation{
					URI: file,
				},
			},
		}
		if line > 0 {
			loc.PhysicalLocation.Region = &Region{
				StartLine: line,
			}
		}
		result.Locations = []Location{loc}
	}

	r.Results = append(r.Results, result)
}

// WriteJSON writes the SARIF log to a JSON file
func (l *Log) WriteJSON(path string) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling SARIF: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// SeverityToLevel converts common severity strings to SARIF levels
func SeverityToLevel(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "high":
		return "error"
	case "medium":
		return "warning"
	case "low", "info":
		return "note"
	default:
		return "none"
	}
}

// Exporter converts Zero scan results to SARIF format
type Exporter struct {
	analysisDir string
	repoPath    string
}

// NewExporter creates a new SARIF exporter
func NewExporter(analysisDir, repoPath string) *Exporter {
	return &Exporter{
		analysisDir: analysisDir,
		repoPath:    repoPath,
	}
}

// Export generates a SARIF log from all scanner results
func (e *Exporter) Export() (*Log, error) {
	log := NewLog()

	// Export each scanner's results
	if err := e.exportCodeSecurity(log); err != nil {
		// Continue even if one scanner fails
	}
	if err := e.exportPackageVulns(log); err != nil {
		// Continue
	}
	if err := e.exportCrypto(log); err != nil {
		// Continue
	}
	if err := e.exportCodeQuality(log); err != nil {
		// Continue
	}

	return log, nil
}

// exportCodeSecurity exports code-security scanner results
func (e *Exporter) exportCodeSecurity(log *Log) error {
	path := filepath.Join(e.analysisDir, "code-security.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // File not found is OK
	}

	var result struct {
		Findings struct {
			Vulns []struct {
				RuleID   string `json:"rule_id"`
				File     string `json:"file"`
				Line     int    `json:"line"`
				Severity string `json:"severity"`
				Message  string `json:"message"`
				Category string `json:"category"`
			} `json:"vulns"`
			Secrets []struct {
				Type     string `json:"type"`
				File     string `json:"file"`
				Line     int    `json:"line"`
				Severity string `json:"severity"`
			} `json:"secrets"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	// Create run for code vulnerabilities
	if len(result.Findings.Vulns) > 0 {
		run := NewRun("zero-code-security", "1.0.0", "https://github.com/crashappsec/zero")
		ruleMap := make(map[string]int)

		for _, v := range result.Findings.Vulns {
			ruleIndex, ok := ruleMap[v.RuleID]
			if !ok {
				ruleIndex = run.AddRule(
					v.RuleID,
					v.RuleID,
					v.Message,
					"",
					SeverityToLevel(v.Severity),
				)
				ruleMap[v.RuleID] = ruleIndex
			}

			run.AddResult(
				v.RuleID,
				ruleIndex,
				SeverityToLevel(v.Severity),
				v.Message,
				v.File,
				v.Line,
			)
		}

		log.Runs = append(log.Runs, *run)
	}

	// Create run for secrets
	if len(result.Findings.Secrets) > 0 {
		run := NewRun("zero-secrets", "1.0.0", "https://github.com/crashappsec/zero")
		ruleMap := make(map[string]int)

		for _, s := range result.Findings.Secrets {
			ruleID := "secret/" + s.Type
			ruleIndex, ok := ruleMap[ruleID]
			if !ok {
				ruleIndex = run.AddRule(
					ruleID,
					s.Type,
					fmt.Sprintf("Detected %s secret", s.Type),
					"",
					"error",
				)
				ruleMap[ruleID] = ruleIndex
			}

			run.AddResult(
				ruleID,
				ruleIndex,
				"error",
				fmt.Sprintf("Detected %s secret", s.Type),
				s.File,
				s.Line,
			)
		}

		log.Runs = append(log.Runs, *run)
	}

	return nil
}

// exportPackageVulns exports package vulnerability results
func (e *Exporter) exportPackageVulns(log *Log) error {
	path := filepath.Join(e.analysisDir, "package-analysis.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var result struct {
		Findings struct {
			Vulns []struct {
				ID        string   `json:"id"`
				Aliases   []string `json:"aliases"`
				Package   string   `json:"package"`
				Version   string   `json:"version"`
				Severity  string   `json:"severity"`
				Title     string   `json:"title"`
				Ecosystem string   `json:"ecosystem"`
			} `json:"vulns"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	if len(result.Findings.Vulns) == 0 {
		return nil
	}

	run := NewRun("zero-package-vulns", "1.0.0", "https://github.com/crashappsec/zero")
	ruleMap := make(map[string]int)

	for _, v := range result.Findings.Vulns {
		ruleIndex, ok := ruleMap[v.ID]
		if !ok {
			ruleIndex = run.AddRule(
				v.ID,
				v.ID,
				v.Title,
				fmt.Sprintf("https://osv.dev/vulnerability/%s", v.ID),
				SeverityToLevel(v.Severity),
			)
			ruleMap[v.ID] = ruleIndex
		}

		// Package vulnerabilities don't have file locations, use logical location
		result := Result{
			RuleID:    v.ID,
			RuleIndex: ruleIndex,
			Level:     SeverityToLevel(v.Severity),
			Message: Message{
				Text: fmt.Sprintf("%s in %s@%s", v.Title, v.Package, v.Version),
			},
			Locations: []Location{
				{
					LogicalLocations: []LogicalLocation{
						{
							Name:               v.Package,
							FullyQualifiedName: fmt.Sprintf("%s:%s@%s", v.Ecosystem, v.Package, v.Version),
							Kind:               "package",
						},
					},
				},
			},
			Properties: map[string]any{
				"package":   v.Package,
				"version":   v.Version,
				"ecosystem": v.Ecosystem,
			},
		}
		run.Results = append(run.Results, result)
	}

	log.Runs = append(log.Runs, *run)
	return nil
}

// exportCrypto exports crypto scanner results
func (e *Exporter) exportCrypto(log *Log) error {
	path := filepath.Join(e.analysisDir, "crypto.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var result struct {
		Findings struct {
			Ciphers []struct {
				Algorithm  string `json:"algorithm"`
				File       string `json:"file"`
				Line       int    `json:"line"`
				Severity   string `json:"severity"`
				Suggestion string `json:"suggestion"`
			} `json:"ciphers"`
			Keys []struct {
				Type     string `json:"type"`
				File     string `json:"file"`
				Line     int    `json:"line"`
				Severity string `json:"severity"`
			} `json:"keys"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	hasFindings := len(result.Findings.Ciphers) > 0 || len(result.Findings.Keys) > 0
	if !hasFindings {
		return nil
	}

	run := NewRun("zero-crypto", "1.0.0", "https://github.com/crashappsec/zero")
	ruleMap := make(map[string]int)

	// Weak ciphers
	for _, c := range result.Findings.Ciphers {
		ruleID := "crypto/weak-cipher/" + c.Algorithm
		ruleIndex, ok := ruleMap[ruleID]
		if !ok {
			ruleIndex = run.AddRule(
				ruleID,
				"Weak Cipher: "+c.Algorithm,
				fmt.Sprintf("Use of weak cipher algorithm %s", c.Algorithm),
				"",
				SeverityToLevel(c.Severity),
			)
			ruleMap[ruleID] = ruleIndex
		}

		msg := fmt.Sprintf("Weak cipher %s detected", c.Algorithm)
		if c.Suggestion != "" {
			msg += ". " + c.Suggestion
		}
		run.AddResult(ruleID, ruleIndex, SeverityToLevel(c.Severity), msg, c.File, c.Line)
	}

	// Hardcoded keys
	for _, k := range result.Findings.Keys {
		ruleID := "crypto/hardcoded-key/" + k.Type
		ruleIndex, ok := ruleMap[ruleID]
		if !ok {
			ruleIndex = run.AddRule(
				ruleID,
				"Hardcoded Key: "+k.Type,
				fmt.Sprintf("Hardcoded %s key detected", k.Type),
				"",
				"error",
			)
			ruleMap[ruleID] = ruleIndex
		}

		run.AddResult(ruleID, ruleIndex, "error", fmt.Sprintf("Hardcoded %s key", k.Type), k.File, k.Line)
	}

	log.Runs = append(log.Runs, *run)
	return nil
}

// exportCodeQuality exports code quality results
func (e *Exporter) exportCodeQuality(log *Log) error {
	path := filepath.Join(e.analysisDir, "code-quality.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var result struct {
		Findings struct {
			TechDebt struct {
				Markers []struct {
					Type     string `json:"type"`
					File     string `json:"file"`
					Line     int    `json:"line"`
					Message  string `json:"message"`
					Priority string `json:"priority"`
				} `json:"markers"`
			} `json:"tech_debt"`
			Complexity []struct {
				File       string `json:"file"`
				Function   string `json:"function"`
				Line       int    `json:"line"`
				Complexity int    `json:"complexity"`
				Type       string `json:"type"`
			} `json:"complexity"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	hasFindings := len(result.Findings.TechDebt.Markers) > 0 || len(result.Findings.Complexity) > 0
	if !hasFindings {
		return nil
	}

	run := NewRun("zero-code-quality", "1.0.0", "https://github.com/crashappsec/zero")
	ruleMap := make(map[string]int)

	// Tech debt markers
	for _, m := range result.Findings.TechDebt.Markers {
		ruleID := "quality/tech-debt/" + m.Type
		ruleIndex, ok := ruleMap[ruleID]
		if !ok {
			ruleIndex = run.AddRule(
				ruleID,
				m.Type+" Comment",
				fmt.Sprintf("Technical debt marker: %s", m.Type),
				"",
				"note",
			)
			ruleMap[ruleID] = ruleIndex
		}

		run.AddResult(ruleID, ruleIndex, "note", m.Message, m.File, m.Line)
	}

	// Complexity
	for _, c := range result.Findings.Complexity {
		ruleID := "quality/complexity/" + c.Type
		ruleIndex, ok := ruleMap[ruleID]
		if !ok {
			ruleIndex = run.AddRule(
				ruleID,
				"High Complexity",
				"Function has high cyclomatic complexity",
				"",
				"warning",
			)
			ruleMap[ruleID] = ruleIndex
		}

		msg := fmt.Sprintf("Function %s has complexity %d", c.Function, c.Complexity)
		run.AddResult(ruleID, ruleIndex, "warning", msg, c.File, c.Line)
	}

	log.Runs = append(log.Runs, *run)
	return nil
}
