package cyclonedx

import (
	"crypto/rand"
	"fmt"
)

// ZeroVersion is the version of Zero generating the BOM
const ZeroVersion = "3.7.0"

// NewBOM creates a new CycloneDX BOM with default metadata
func NewBOM() *BOM {
	return &BOM{
		BOMFormat:   "CycloneDX",
		SpecVersion: SpecVersion,
		Version:     1,
		Metadata: &Metadata{
			Timestamp: TimestampNow(),
			Tools: &Tools{
				Components: []ToolComponent{
					{
						Type:    "application",
						Name:    "zero",
						Version: ZeroVersion,
						Manufacturer: &OrgEntity{
							Name: "Crash Override",
						},
					},
				},
			},
		},
		Components: []Component{},
	}
}

// NewMLBOM creates a new ML-BOM (Machine Learning Bill of Materials)
func NewMLBOM() *BOM {
	bom := NewBOM()
	bom.SerialNumber = generateUUID()
	bom.Metadata.Lifecycles = []Lifecycle{
		{Phase: "discovery"},
	}
	return bom
}

// NewCBOM creates a new CBOM (Cryptography Bill of Materials)
func NewCBOM() *BOM {
	bom := NewBOM()
	bom.SerialNumber = generateUUID()
	bom.Metadata.Lifecycles = []Lifecycle{
		{Phase: "discovery"},
	}
	return bom
}

// WithSerialNumber sets the BOM serial number
func (b *BOM) WithSerialNumber(serial string) *BOM {
	b.SerialNumber = serial
	return b
}

// WithLifecycle adds a lifecycle phase
func (b *BOM) WithLifecycle(phase string) *BOM {
	if b.Metadata == nil {
		b.Metadata = &Metadata{}
	}
	b.Metadata.Lifecycles = append(b.Metadata.Lifecycles, Lifecycle{Phase: phase})
	return b
}

// WithComponent adds a component to the BOM
func (b *BOM) WithComponent(c Component) *BOM {
	b.Components = append(b.Components, c)
	return b
}

// WithComponents adds multiple components to the BOM
func (b *BOM) WithComponents(components []Component) *BOM {
	b.Components = append(b.Components, components...)
	return b
}

// WithVulnerability adds a vulnerability to the BOM
func (b *BOM) WithVulnerability(v Vulnerability) *BOM {
	b.Vulnerabilities = append(b.Vulnerabilities, v)
	return b
}

// WithVulnerabilities adds multiple vulnerabilities to the BOM
func (b *BOM) WithVulnerabilities(vulns []Vulnerability) *BOM {
	b.Vulnerabilities = append(b.Vulnerabilities, vulns...)
	return b
}

// WithDependency adds a dependency relationship
func (b *BOM) WithDependency(d Dependency) *BOM {
	b.Dependencies = append(b.Dependencies, d)
	return b
}

// WithMetadataComponent sets the metadata component (the main project)
func (b *BOM) WithMetadataComponent(c *Component) *BOM {
	if b.Metadata == nil {
		b.Metadata = &Metadata{}
	}
	b.Metadata.Component = c
	return b
}

// AddProperty adds a property to a component
func (c *Component) AddProperty(name, value string) {
	c.Properties = append(c.Properties, Property{Name: name, Value: value})
}

// AddExternalRef adds an external reference to a component
func (c *Component) AddExternalRef(refType, url string) {
	c.ExternalRefs = append(c.ExternalRefs, ExternalRef{Type: refType, URL: url})
}

// AddLicense adds a license to a component
func (c *Component) AddLicense(id string) {
	c.Licenses = append(c.Licenses, LicenseChoice{
		License: &License{ID: id},
	})
}

// NewComponent creates a new component with the given type and name
func NewComponent(componentType, name string) Component {
	return Component{
		Type:   componentType,
		Name:   name,
		BOMRef: fmt.Sprintf("%s/%s", componentType, name),
	}
}

// NewMLModelComponent creates a new ML model component
func NewMLModelComponent(name, version string) Component {
	return Component{
		Type:    ComponentTypeMLModel,
		Name:    name,
		Version: version,
		BOMRef:  fmt.Sprintf("model/%s@%s", name, version),
	}
}

// NewCryptoComponent creates a new cryptographic asset component
func NewCryptoComponent(name string) Component {
	return Component{
		Type:   ComponentTypeCryptographicAsset,
		Name:   name,
		BOMRef: fmt.Sprintf("crypto/%s", name),
	}
}

// NewDataComponent creates a new data component
func NewDataComponent(name string) Component {
	return Component{
		Type:   ComponentTypeData,
		Name:   name,
		BOMRef: fmt.Sprintf("data/%s", name),
	}
}

// generateUUID generates a URN UUID for serial numbers
func generateUUID() string {
	uuid := make([]byte, 16)
	_, _ = rand.Read(uuid)
	// Set version (4) and variant bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("urn:uuid:%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// SeverityToCycloneDX converts Zero severity to CycloneDX severity
func SeverityToCycloneDX(severity string) string {
	switch severity {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	case "info", "informational":
		return "info"
	default:
		return "unknown"
	}
}

// CWEToInt converts a CWE string (e.g., "CWE-79") to an integer
func CWEToInt(cwe string) int {
	var id int
	_, _ = fmt.Sscanf(cwe, "CWE-%d", &id)
	return id
}
