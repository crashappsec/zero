package cyclonedx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Exporter handles CycloneDX BOM generation and export
type Exporter struct {
	outputDir string
}

// NewExporter creates a new CycloneDX exporter
func NewExporter(outputDir string) *Exporter {
	return &Exporter{outputDir: outputDir}
}

// ExportMLBOM exports an ML-BOM from tech-id scanner results
func (e *Exporter) ExportMLBOM(techIDResult interface{}) (*BOM, error) {
	bom := NewMLBOM()

	// Parse the tech-id result (could be raw JSON or typed struct)
	data, err := toMap(techIDResult)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tech-id result: %w", err)
	}

	// Extract findings
	findings, _ := data["findings"].(map[string]interface{})

	// Process models
	if models, ok := findings["models"].([]interface{}); ok {
		for _, m := range models {
			model, _ := m.(map[string]interface{})
			c := e.modelToComponent(model)
			bom.WithComponent(c)
		}
	}

	// Process frameworks
	if frameworks, ok := findings["frameworks"].([]interface{}); ok {
		for _, f := range frameworks {
			framework, _ := f.(map[string]interface{})
			c := e.frameworkToComponent(framework)
			bom.WithComponent(c)
		}
	}

	// Process datasets
	if datasets, ok := findings["datasets"].([]interface{}); ok {
		for _, d := range datasets {
			dataset, _ := d.(map[string]interface{})
			c := e.datasetToComponent(dataset)
			bom.WithComponent(c)
		}
	}

	// Process security findings as vulnerabilities
	if security, ok := findings["security"].([]interface{}); ok {
		for _, s := range security {
			finding, _ := s.(map[string]interface{})
			v := e.securityFindingToVulnerability(finding)
			bom.WithVulnerability(v)
		}
	}

	// Process governance findings as vulnerabilities
	if governance, ok := findings["governance"].([]interface{}); ok {
		for _, g := range governance {
			finding, _ := g.(map[string]interface{})
			v := e.governanceFindingToVulnerability(finding)
			bom.WithVulnerability(v)
		}
	}

	return bom, nil
}

// ExportCBOM exports a CBOM from crypto scanner results
func (e *Exporter) ExportCBOM(cryptoResult interface{}) (*BOM, error) {
	bom := NewCBOM()

	// Parse the crypto result
	data, err := toMap(cryptoResult)
	if err != nil {
		return nil, fmt.Errorf("failed to parse crypto result: %w", err)
	}

	// Extract findings
	findings, _ := data["findings"].(map[string]interface{})

	// Process cipher findings
	if ciphers, ok := findings["ciphers"].([]interface{}); ok {
		for _, c := range ciphers {
			cipher, _ := c.(map[string]interface{})
			comp := e.cipherToComponent(cipher)
			bom.WithComponent(comp)

			// Add vulnerability for weak ciphers
			if severity, _ := cipher["severity"].(string); severity == "high" || severity == "critical" {
				v := e.cipherFindingToVulnerability(cipher)
				bom.WithVulnerability(v)
			}
		}
	}

	// Process key findings
	if keys, ok := findings["keys"].([]interface{}); ok {
		for _, k := range keys {
			key, _ := k.(map[string]interface{})
			// Keys are vulnerabilities (hardcoded keys)
			v := e.keyFindingToVulnerability(key)
			bom.WithVulnerability(v)
		}
	}

	// Process TLS findings
	if tls, ok := findings["tls"].([]interface{}); ok {
		for _, t := range tls {
			tlsFinding, _ := t.(map[string]interface{})
			comp := e.tlsToComponent(tlsFinding)
			bom.WithComponent(comp)

			// Add vulnerability for TLS issues
			v := e.tlsFindingToVulnerability(tlsFinding)
			bom.WithVulnerability(v)
		}
	}

	// Process certificate findings
	if certs, ok := findings["certificates"].(map[string]interface{}); ok {
		if certList, ok := certs["certificates"].([]interface{}); ok {
			for _, c := range certList {
				cert, _ := c.(map[string]interface{})
				comp := e.certToComponent(cert)
				bom.WithComponent(comp)
			}
		}
		// Certificate findings as vulnerabilities
		if certFindings, ok := certs["findings"].([]interface{}); ok {
			for _, f := range certFindings {
				finding, _ := f.(map[string]interface{})
				v := e.certFindingToVulnerability(finding)
				bom.WithVulnerability(v)
			}
		}
	}

	return bom, nil
}

// WriteMLBOM writes the ML-BOM to a file
func (e *Exporter) WriteMLBOM(bom *BOM, filename string) error {
	return e.writeBOM(bom, filename)
}

// WriteCBOM writes the CBOM to a file
func (e *Exporter) WriteCBOM(bom *BOM, filename string) error {
	return e.writeBOM(bom, filename)
}

// writeBOM writes any BOM to a file
func (e *Exporter) writeBOM(bom *BOM, filename string) error {
	data, err := bom.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize BOM: %w", err)
	}

	path := filepath.Join(e.outputDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write BOM: %w", err)
	}

	return nil
}

// Helper conversion methods

func (e *Exporter) modelToComponent(model map[string]interface{}) Component {
	name, _ := model["name"].(string)
	version, _ := model["version"].(string)
	source, _ := model["source"].(string)
	sourceURL, _ := model["source_url"].(string)
	format, _ := model["format"].(string)
	architecture, _ := model["architecture"].(string)
	task, _ := model["task"].(string)
	license, _ := model["license"].(string)

	if version == "" {
		version = "unknown"
	}

	c := MLModelToComponent(name, version, source, sourceURL, format, architecture, task, license)

	// Add model card from metadata if available
	if mcData, ok := model["model_card"].(map[string]interface{}); ok {
		mc := NewModelCard()
		if desc, ok := mcData["description"].(string); ok {
			c.Description = desc
		}
		if limitations, ok := mcData["limitations"].(string); ok {
			mc.WithLimitation(limitations)
		}
		if intendedUse, ok := mcData["intended_use"].(string); ok {
			mc.WithUseCase(intendedUse)
		}
		if datasets, ok := mcData["datasets"].([]interface{}); ok {
			for _, d := range datasets {
				if ds, ok := d.(string); ok {
					mc.WithDataset(fmt.Sprintf("dataset/%s", ds), "training")
				}
			}
		}
		c.ModelCard = mc
	}

	// Add security risk as property
	if risk, ok := model["security_risk"].(string); ok && risk != "" {
		c.AddProperty("zero:security_risk", risk)
	}

	// Add file path if present
	if filePath, ok := model["file_path"].(string); ok && filePath != "" {
		c.Evidence = &Evidence{
			Occurrences: []Occurrence{{Location: filePath}},
		}
	}

	return c
}

func (e *Exporter) frameworkToComponent(framework map[string]interface{}) Component {
	name, _ := framework["name"].(string)
	version, _ := framework["version"].(string)
	category, _ := framework["category"].(string)
	pkg, _ := framework["package"].(string)

	if version == "" {
		version = "unknown"
	}

	return FrameworkToComponent(name, version, category, pkg)
}

func (e *Exporter) datasetToComponent(dataset map[string]interface{}) Component {
	name, _ := dataset["name"].(string)
	source, _ := dataset["source"].(string)
	sourceURL, _ := dataset["source_url"].(string)
	license, _ := dataset["license"].(string)
	description, _ := dataset["description"].(string)

	return DatasetToComponent(name, source, sourceURL, license, description)
}

func (e *Exporter) securityFindingToVulnerability(finding map[string]interface{}) Vulnerability {
	id, _ := finding["id"].(string)
	title, _ := finding["title"].(string)
	description, _ := finding["description"].(string)
	severity, _ := finding["severity"].(string)
	category, _ := finding["category"].(string)
	remediation, _ := finding["remediation"].(string)
	modelName, _ := finding["model_name"].(string)

	v := Vulnerability{
		ID:             id,
		Source:         &VulnSource{Name: "Zero AI Security Scanner"},
		Description:    fmt.Sprintf("%s: %s", title, description),
		Recommendation: remediation,
		Ratings: []VulnRating{
			{
				Severity: SeverityToCycloneDX(severity),
				Method:   "other",
			},
		},
	}

	// Add category as CWE if applicable
	if cwe := categoryToCWE(category); cwe > 0 {
		v.CWEs = []int{cwe}
	}

	// Link to affected model if present
	if modelName != "" {
		v.Affects = []VulnAffect{
			{Ref: fmt.Sprintf("model/%s", modelName)},
		}
	}

	return v
}

func (e *Exporter) governanceFindingToVulnerability(finding map[string]interface{}) Vulnerability {
	id, _ := finding["id"].(string)
	title, _ := finding["title"].(string)
	description, _ := finding["description"].(string)
	severity, _ := finding["severity"].(string)
	remediation, _ := finding["remediation"].(string)
	modelName, _ := finding["model_name"].(string)

	v := Vulnerability{
		ID:             id,
		Source:         &VulnSource{Name: "Zero AI Governance Scanner"},
		Description:    fmt.Sprintf("%s: %s", title, description),
		Recommendation: remediation,
		Ratings: []VulnRating{
			{
				Severity: SeverityToCycloneDX(severity),
				Method:   "other",
			},
		},
	}

	if modelName != "" {
		v.Affects = []VulnAffect{
			{Ref: fmt.Sprintf("model/%s", modelName)},
		}
	}

	return v
}

func (e *Exporter) cipherToComponent(cipher map[string]interface{}) Component {
	algorithm, _ := cipher["algorithm"].(string)
	severity, _ := cipher["severity"].(string)
	file, _ := cipher["file"].(string)
	line := int(getFloat64(cipher, "line"))
	description, _ := cipher["description"].(string)

	// Extract mode from match or description
	mode := extractMode(algorithm, description)

	return CipherFindingToComponent(algorithm, mode, severity, file, line, description)
}

func (e *Exporter) cipherFindingToVulnerability(cipher map[string]interface{}) Vulnerability {
	algorithm, _ := cipher["algorithm"].(string)
	severity, _ := cipher["severity"].(string)
	description, _ := cipher["description"].(string)
	suggestion, _ := cipher["suggestion"].(string)
	cwe, _ := cipher["cwe"].(string)
	file, _ := cipher["file"].(string)

	v := Vulnerability{
		ID:             fmt.Sprintf("CRYPTO-WEAK-%s", strings.ToUpper(algorithm)),
		Source:         &VulnSource{Name: "Zero Crypto Scanner"},
		Description:    description,
		Recommendation: suggestion,
		Ratings: []VulnRating{
			{
				Severity: SeverityToCycloneDX(severity),
				Method:   "other",
			},
		},
		Affects: []VulnAffect{
			{Ref: fmt.Sprintf("crypto/algorithm/%s", strings.ToLower(algorithm))},
		},
	}

	if cweInt := CWEToInt(cwe); cweInt > 0 {
		v.CWEs = []int{cweInt}
	}

	// Add file location as detail
	if file != "" {
		v.Detail = fmt.Sprintf("Found in: %s", file)
	}

	return v
}

func (e *Exporter) keyFindingToVulnerability(key map[string]interface{}) Vulnerability {
	keyType, _ := key["type"].(string)
	severity, _ := key["severity"].(string)
	description, _ := key["description"].(string)
	file, _ := key["file"].(string)
	line := int(getFloat64(key, "line"))
	cwe, _ := key["cwe"].(string)

	v := Vulnerability{
		ID:             fmt.Sprintf("CRYPTO-HARDCODED-KEY-%s", strings.ToUpper(keyType)),
		Source:         &VulnSource{Name: "Zero Crypto Scanner"},
		Description:    description,
		Recommendation: "Remove hardcoded keys and use secure key management",
		Detail:         fmt.Sprintf("Found in: %s:%d", file, line),
		Ratings: []VulnRating{
			{
				Severity: SeverityToCycloneDX(severity),
				Method:   "other",
			},
		},
	}

	if cweInt := CWEToInt(cwe); cweInt > 0 {
		v.CWEs = []int{cweInt}
	}

	return v
}

func (e *Exporter) tlsToComponent(tls map[string]interface{}) Component {
	tlsType, _ := tls["type"].(string)
	severity, _ := tls["severity"].(string)
	file, _ := tls["file"].(string)
	line := int(getFloat64(tls, "line"))
	description, _ := tls["description"].(string)

	// Extract version from type or description
	version := extractTLSVersion(tlsType, description)

	return TLSFindingToComponent(tlsType, version, severity, file, line, description)
}

func (e *Exporter) tlsFindingToVulnerability(tls map[string]interface{}) Vulnerability {
	tlsType, _ := tls["type"].(string)
	severity, _ := tls["severity"].(string)
	description, _ := tls["description"].(string)
	suggestion, _ := tls["suggestion"].(string)
	cwe, _ := tls["cwe"].(string)
	file, _ := tls["file"].(string)
	line := int(getFloat64(tls, "line"))

	v := Vulnerability{
		ID:             fmt.Sprintf("CRYPTO-TLS-%s", strings.ToUpper(tlsType)),
		Source:         &VulnSource{Name: "Zero Crypto Scanner"},
		Description:    description,
		Recommendation: suggestion,
		Detail:         fmt.Sprintf("Found in: %s:%d", file, line),
		Ratings: []VulnRating{
			{
				Severity: SeverityToCycloneDX(severity),
				Method:   "other",
			},
		},
	}

	if cweInt := CWEToInt(cwe); cweInt > 0 {
		v.CWEs = []int{cweInt}
	}

	return v
}

func (e *Exporter) certToComponent(cert map[string]interface{}) Component {
	subject, _ := cert["subject"].(string)
	issuer, _ := cert["issuer"].(string)
	notBefore, _ := cert["not_before"].(string)
	notAfter, _ := cert["not_after"].(string)
	keyType, _ := cert["key_type"].(string)
	keySize := int(getFloat64(cert, "key_size"))
	sigAlgo, _ := cert["signature_algorithm"].(string)
	file, _ := cert["file"].(string)
	isSelfSigned, _ := cert["is_self_signed"].(bool)

	return CertInfoToComponent(subject, issuer, notBefore, notAfter, keyType, keySize, sigAlgo, file, isSelfSigned)
}

func (e *Exporter) certFindingToVulnerability(finding map[string]interface{}) Vulnerability {
	certType, _ := finding["type"].(string)
	severity, _ := finding["severity"].(string)
	description, _ := finding["description"].(string)
	suggestion, _ := finding["suggestion"].(string)
	file, _ := finding["file"].(string)

	v := Vulnerability{
		ID:             fmt.Sprintf("CRYPTO-CERT-%s", strings.ToUpper(certType)),
		Source:         &VulnSource{Name: "Zero Crypto Scanner"},
		Description:    description,
		Recommendation: suggestion,
		Ratings: []VulnRating{
			{
				Severity: SeverityToCycloneDX(severity),
				Method:   "other",
			},
		},
	}

	if file != "" {
		v.Detail = fmt.Sprintf("Certificate file: %s", file)
		v.Affects = []VulnAffect{
			{Ref: fmt.Sprintf("crypto/certificate/%s", file)},
		}
	}

	return v
}

// Helper functions

func toMap(v interface{}) (map[string]interface{}, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		return val, nil
	case []byte:
		var m map[string]interface{}
		if err := json.Unmarshal(val, &m); err != nil {
			return nil, err
		}
		return m, nil
	default:
		// Try to marshal and unmarshal to get a map
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return m, nil
	}
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return 0
}

func extractMode(algorithm, description string) string {
	modes := []string{"gcm", "cbc", "ctr", "ecb", "cfb", "ofb", "ccm"}
	text := strings.ToLower(algorithm + " " + description)
	for _, mode := range modes {
		if strings.Contains(text, mode) {
			return mode
		}
	}
	return ""
}

func extractTLSVersion(tlsType, description string) string {
	text := strings.ToLower(tlsType + " " + description)
	versions := []string{"1.3", "1.2", "1.1", "1.0"}
	for _, v := range versions {
		if strings.Contains(text, "tls"+v) || strings.Contains(text, "tls "+v) ||
		   strings.Contains(text, "tlsv"+v) {
			return v
		}
	}
	if strings.Contains(text, "sslv3") || strings.Contains(text, "ssl3") {
		return "SSLv3"
	}
	return "unknown"
}

func categoryToCWE(category string) int {
	cwes := map[string]int{
		"pickle_rce":       502, // Deserialization of Untrusted Data
		"unsafe_loading":   502,
		"api_key_exposure": 798, // Use of Hard-coded Credentials
		"prompt_injection": 94,  // Improper Control of Generation of Code
		"model_poisoning":  502,
	}
	if cwe, ok := cwes[category]; ok {
		return cwe
	}
	return 0
}
