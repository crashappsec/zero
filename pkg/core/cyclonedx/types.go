// Package cyclonedx provides CycloneDX BOM generation and export capabilities.
// Supports CycloneDX 1.6 specification including ML-BOM and CBOM extensions.
package cyclonedx

import (
	"encoding/json"
	"time"
)

// SpecVersion is the CycloneDX specification version
const SpecVersion = "1.6"

// BOM represents a CycloneDX Bill of Materials
type BOM struct {
	BOMFormat       string           `json:"bomFormat"`
	SpecVersion     string           `json:"specVersion"`
	SerialNumber    string           `json:"serialNumber,omitempty"`
	Version         int              `json:"version"`
	Metadata        *Metadata        `json:"metadata,omitempty"`
	Components      []Component      `json:"components,omitempty"`
	Services        []Service        `json:"services,omitempty"`
	Dependencies    []Dependency     `json:"dependencies,omitempty"`
	Vulnerabilities []Vulnerability  `json:"vulnerabilities,omitempty"`
	Compositions    []Composition    `json:"compositions,omitempty"`
	ExternalRefs    []ExternalRef    `json:"externalReferences,omitempty"`
}

// Metadata contains BOM metadata
type Metadata struct {
	Timestamp  string       `json:"timestamp,omitempty"`
	Lifecycles []Lifecycle  `json:"lifecycles,omitempty"`
	Tools      *Tools       `json:"tools,omitempty"`
	Authors    []Author     `json:"authors,omitempty"`
	Component  *Component   `json:"component,omitempty"`
	Supplier   *OrgEntity   `json:"supplier,omitempty"`
}

// Lifecycle represents a BOM lifecycle phase
type Lifecycle struct {
	Phase       string `json:"phase,omitempty"`       // design, pre-build, build, post-build, operations, discovery, decommission
	Name        string `json:"name,omitempty"`        // Custom lifecycle name
	Description string `json:"description,omitempty"`
}

// Tools contains tool information in CycloneDX 1.5+ format
type Tools struct {
	Components []ToolComponent `json:"components,omitempty"`
}

// ToolComponent represents a tool component
type ToolComponent struct {
	Type         string     `json:"type,omitempty"`
	Name         string     `json:"name,omitempty"`
	Version      string     `json:"version,omitempty"`
	Manufacturer *OrgEntity `json:"manufacturer,omitempty"`
}

// Author represents a BOM author
type Author struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// OrgEntity represents an organization
type OrgEntity struct {
	Name    string   `json:"name,omitempty"`
	URL     []string `json:"url,omitempty"`
	Contact []Author `json:"contact,omitempty"`
}

// Component represents a CycloneDX component
type Component struct {
	Type             string            `json:"type"`
	BOMRef           string            `json:"bom-ref,omitempty"`
	Name             string            `json:"name"`
	Version          string            `json:"version,omitempty"`
	Group            string            `json:"group,omitempty"`
	Description      string            `json:"description,omitempty"`
	Scope            string            `json:"scope,omitempty"` // required, optional, excluded
	Purl             string            `json:"purl,omitempty"`
	CPE              string            `json:"cpe,omitempty"`
	Hashes           []Hash            `json:"hashes,omitempty"`
	Licenses         []LicenseChoice   `json:"licenses,omitempty"`
	Copyright        string            `json:"copyright,omitempty"`
	Supplier         *OrgEntity        `json:"supplier,omitempty"`
	Author           string            `json:"author,omitempty"`
	Properties       []Property        `json:"properties,omitempty"`
	ExternalRefs     []ExternalRef     `json:"externalReferences,omitempty"`
	Components       []Component       `json:"components,omitempty"`
	Evidence         *Evidence         `json:"evidence,omitempty"`
	Pedigree         *Pedigree         `json:"pedigree,omitempty"`
	ModelCard        *ModelCard        `json:"modelCard,omitempty"`        // ML-BOM
	Data             *DataComponent    `json:"data,omitempty"`             // Data component
	CryptoProperties *CryptoProperties `json:"cryptoProperties,omitempty"` // CBOM
}

// ComponentType constants
const (
	ComponentTypeApplication      = "application"
	ComponentTypeFramework        = "framework"
	ComponentTypeLibrary          = "library"
	ComponentTypeContainer        = "container"
	ComponentTypePlatform         = "platform"
	ComponentTypeOS               = "operating-system"
	ComponentTypeDevice           = "device"
	ComponentTypeDeviceDriver     = "device-driver"
	ComponentTypeFirmware         = "firmware"
	ComponentTypeFile             = "file"
	ComponentTypeMLModel          = "machine-learning-model"
	ComponentTypeData             = "data"
	ComponentTypeCryptographicAsset = "cryptographic-asset"
)

// Hash represents a component hash
type Hash struct {
	Algorithm string `json:"alg"`
	Content   string `json:"content"`
}

// HashAlgorithm constants
const (
	HashAlgSHA256    = "SHA-256"
	HashAlgSHA384    = "SHA-384"
	HashAlgSHA512    = "SHA-512"
	HashAlgSHA3_256  = "SHA3-256"
	HashAlgSHA3_512  = "SHA3-512"
	HashAlgBLAKE2b   = "BLAKE2b-256"
	HashAlgBLAKE3    = "BLAKE3"
)

// LicenseChoice represents a license or expression
type LicenseChoice struct {
	License    *License `json:"license,omitempty"`
	Expression string   `json:"expression,omitempty"`
}

// License represents a SPDX license
type License struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Property represents a key-value property
type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ExternalRef represents an external reference
type ExternalRef struct {
	Type    string   `json:"type"`
	URL     string   `json:"url"`
	Comment string   `json:"comment,omitempty"`
	Hashes  []Hash   `json:"hashes,omitempty"`
}

// ExternalRefType constants
const (
	ExternalRefVCS           = "vcs"
	ExternalRefIssueTracker  = "issue-tracker"
	ExternalRefWebsite       = "website"
	ExternalRefAdvisories    = "advisories"
	ExternalRefBOM           = "bom"
	ExternalRefDocumentation = "documentation"
	ExternalRefSupport       = "support"
	ExternalRefLicense       = "license"
	ExternalRefBuildMeta     = "build-meta"
	ExternalRefReleaseNotes  = "release-notes"
	ExternalRefModelCard     = "model-card"
	ExternalRefEvidence      = "evidence"
	ExternalRefAttestation   = "attestation"
)

// Evidence represents component evidence
type Evidence struct {
	Identity   *EvidenceIdentity `json:"identity,omitempty"`
	Occurrences []Occurrence     `json:"occurrences,omitempty"`
}

// EvidenceIdentity represents identity evidence
type EvidenceIdentity struct {
	Field      string  `json:"field,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	Methods    []struct {
		Technique  string  `json:"technique,omitempty"`
		Confidence float64 `json:"confidence,omitempty"`
		Value      string  `json:"value,omitempty"`
	} `json:"methods,omitempty"`
}

// Occurrence represents where a component occurs
type Occurrence struct {
	Location string `json:"location,omitempty"`
	Line     int    `json:"line,omitempty"`
}

// Pedigree represents component lineage
type Pedigree struct {
	Ancestors   []Component `json:"ancestors,omitempty"`
	Descendants []Component `json:"descendants,omitempty"`
	Variants    []Component `json:"variants,omitempty"`
	Notes       string      `json:"notes,omitempty"`
}

// Service represents a CycloneDX service
type Service struct {
	BOMRef         string        `json:"bom-ref,omitempty"`
	Name           string        `json:"name"`
	Version        string        `json:"version,omitempty"`
	Description    string        `json:"description,omitempty"`
	Endpoints      []string      `json:"endpoints,omitempty"`
	Authenticated  bool          `json:"authenticated,omitempty"`
	XTrustBoundary bool          `json:"x-trust-boundary,omitempty"`
	TrustZone      string        `json:"trustZone,omitempty"`
	Data           []ServiceData `json:"data,omitempty"`
	Licenses       []LicenseChoice `json:"licenses,omitempty"`
	ExternalRefs   []ExternalRef   `json:"externalReferences,omitempty"`
	Services       []Service     `json:"services,omitempty"`
}

// ServiceData represents data flowing through a service
type ServiceData struct {
	Classification string `json:"classification,omitempty"`
	Flow           string `json:"flow,omitempty"` // inbound, outbound, bi-directional
}

// Dependency represents a dependency relationship
type Dependency struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn,omitempty"`
	Provides  []string `json:"provides,omitempty"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID             string              `json:"id"`
	Source         *VulnSource         `json:"source,omitempty"`
	References     []VulnReference     `json:"references,omitempty"`
	Ratings        []VulnRating        `json:"ratings,omitempty"`
	CWEs           []int               `json:"cwes,omitempty"`
	Description    string              `json:"description,omitempty"`
	Detail         string              `json:"detail,omitempty"`
	Recommendation string              `json:"recommendation,omitempty"`
	Workaround     string              `json:"workaround,omitempty"`
	Advisories     []string            `json:"advisories,omitempty"`
	Created        string              `json:"created,omitempty"`
	Published      string              `json:"published,omitempty"`
	Updated        string              `json:"updated,omitempty"`
	Credits        *VulnCredits        `json:"credits,omitempty"`
	Analysis       *VulnAnalysis       `json:"analysis,omitempty"`
	Affects        []VulnAffect        `json:"affects,omitempty"`
}

// VulnSource represents a vulnerability source
type VulnSource struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// VulnReference represents a vulnerability reference
type VulnReference struct {
	ID     string      `json:"id,omitempty"`
	Source *VulnSource `json:"source,omitempty"`
}

// VulnRating represents a vulnerability rating
type VulnRating struct {
	Source   *VulnSource `json:"source,omitempty"`
	Score    float64     `json:"score,omitempty"`
	Severity string      `json:"severity,omitempty"`
	Method   string      `json:"method,omitempty"` // CVSSv2, CVSSv3, CVSSv31, CVSSv4, OWASP, other
	Vector   string      `json:"vector,omitempty"`
}

// VulnCredits represents vulnerability credits
type VulnCredits struct {
	Individuals   []Author    `json:"individuals,omitempty"`
	Organizations []OrgEntity `json:"organizations,omitempty"`
}

// VulnAnalysis represents vulnerability analysis (VEX)
type VulnAnalysis struct {
	State         string   `json:"state,omitempty"`         // resolved, resolved_with_pedigree, exploitable, in_triage, false_positive, not_affected
	Justification string   `json:"justification,omitempty"` // code_not_present, code_not_reachable, requires_configuration, etc.
	Response      []string `json:"response,omitempty"`      // can_not_fix, will_not_fix, update, rollback, workaround_available
	Detail        string   `json:"detail,omitempty"`
	FirstIssued   string   `json:"firstIssued,omitempty"`
	LastUpdated   string   `json:"lastUpdated,omitempty"`
}

// VEX analysis states
const (
	VEXStateResolved              = "resolved"
	VEXStateResolvedWithPedigree  = "resolved_with_pedigree"
	VEXStateExploitable           = "exploitable"
	VEXStateInTriage              = "in_triage"
	VEXStateFalsePositive         = "false_positive"
	VEXStateNotAffected           = "not_affected"
)

// VEX justifications
const (
	VEXJustCodeNotPresent          = "code_not_present"
	VEXJustCodeNotReachable        = "code_not_reachable"
	VEXJustRequiresConfiguration   = "requires_configuration"
	VEXJustRequiresDependency      = "requires_dependency"
	VEXJustRequiresEnvironment     = "requires_environment"
	VEXJustProtectedByCompiler     = "protected_by_compiler"
	VEXJustProtectedAtRuntime      = "protected_at_runtime"
	VEXJustProtectedAtPerimeter    = "protected_at_perimeter"
	VEXJustProtectedByMitigating   = "protected_by_mitigating_control"
)

// VulnAffect represents an affected component
type VulnAffect struct {
	Ref      string               `json:"ref"`
	Versions []VulnAffectVersion  `json:"versions,omitempty"`
}

// VulnAffectVersion represents an affected version
type VulnAffectVersion struct {
	Version string `json:"version,omitempty"`
	Range   string `json:"range,omitempty"`
	Status  string `json:"status,omitempty"` // affected, unaffected, unknown
}

// Composition represents completeness information
type Composition struct {
	BOMRef       string   `json:"bom-ref,omitempty"`
	Aggregate    string   `json:"aggregate"`  // complete, incomplete, incomplete_first_party_only, etc.
	Assemblies   []string `json:"assemblies,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// CompositionAggregate constants
const (
	CompositionComplete           = "complete"
	CompositionIncomplete         = "incomplete"
	CompositionFirstPartyOnly     = "incomplete_first_party_only"
	CompositionThirdPartyOnly     = "incomplete_third_party_only"
	CompositionUnknown            = "unknown"
	CompositionNotSpecified       = "not_specified"
)

// DataComponent represents a data component
type DataComponent struct {
	Type         string            `json:"type,omitempty"` // configuration, dataset, source-code, other
	Name         string            `json:"name,omitempty"`
	Contents     *DataContents     `json:"contents,omitempty"`
	Classification string          `json:"classification,omitempty"`
	SensitiveData []string         `json:"sensitiveData,omitempty"`
	Governance   *DataGovernance   `json:"governance,omitempty"`
}

// DataContents represents data contents
type DataContents struct {
	Attachment *Attachment `json:"attachment,omitempty"`
	URL        string      `json:"url,omitempty"`
	Properties []Property  `json:"properties,omitempty"`
}

// Attachment represents an attachment
type Attachment struct {
	ContentType string `json:"contentType,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	Content     string `json:"content,omitempty"`
}

// DataGovernance represents data governance
type DataGovernance struct {
	Custodians    []OrgEntity     `json:"custodians,omitempty"`
	Stewards      []OrgEntity     `json:"stewards,omitempty"`
	Owners        []OrgEntity     `json:"owners,omitempty"`
	Licenses      []LicenseChoice `json:"licenses,omitempty"`
}

// ToJSON serializes the BOM to JSON
func (b *BOM) ToJSON() ([]byte, error) {
	return json.MarshalIndent(b, "", "  ")
}

// ToJSONCompact serializes the BOM to compact JSON
func (b *BOM) ToJSONCompact() ([]byte, error) {
	return json.Marshal(b)
}

// FromJSON deserializes a BOM from JSON
func FromJSON(data []byte) (*BOM, error) {
	var bom BOM
	if err := json.Unmarshal(data, &bom); err != nil {
		return nil, err
	}
	return &bom, nil
}

// TimestampNow returns the current time in ISO 8601 format
func TimestampNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}
