<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Prompt: SBOM Format Conversion

## Purpose
Convert SBOMs between different formats (CycloneDX ↔ SPDX) while preserving data integrity and maintaining traceability.

## When to Use
- Converting supplier-provided SBOMs to your organization's standard format
- Integrating with tools that require specific SBOM formats
- Standardizing heterogeneous SBOM collections
- Enabling format-specific analysis capabilities
- Meeting compliance requirements for specific formats

## Prompt

```
Please convert this SBOM to [target format]:

Source Format: [CycloneDX/SPDX]
Source Version: [version]
Target Format: [CycloneDX/SPDX]
Target Version: [version]
Output Format: [JSON/XML/YAML/RDF/Tag-Value/Protobuf]

Requirements:
1. Preserve all component and dependency information
2. Maintain license data and SPDX expressions
3. Convert vulnerability information appropriately
4. Preserve metadata and provenance data
5. Handle format-specific features with annotations where necessary
6. Maintain PURLs (Package URLs) for cross-reference
7. Validate output against target schema
8. Provide conversion report showing:
   - Successfully mapped fields
   - Fields requiring manual review
   - Format-specific features handled
   - Any data loss or approximations

[Paste SBOM content here]
```

## Expected Output
- Converted SBOM in target format
- Validation confirmation
- Conversion report detailing:
  - Field mappings performed
  - Format-specific handling
  - Warnings or information loss
  - Recommendations for manual review

## Variations

### CycloneDX to SPDX
```
Convert this CycloneDX SBOM to SPDX 2.3 format in JSON:

- Map components to SPDX packages
- Convert dependencies to SPDX relationships
- Extract vulnerability data to external references
- Preserve all license information
- Handle services and formulation with annotations
- Maintain component pedigree in annotations

[Paste CycloneDX SBOM]
```

### SPDX to CycloneDX
```
Convert this SPDX document to CycloneDX 1.7 format in JSON:

- Map packages to CycloneDX components
- Convert relationships to dependencies structure
- Extract security references to vulnerabilities section
- Handle file-level data appropriately
- Preserve all checksums and verification codes
- Convert SPDX snippets to annotations

[Paste SPDX document]
```

### Bulk Conversion
```
Convert these multiple SBOMs to a unified format:

Target: CycloneDX 1.7 JSON

Requirements:
- Standardize all to same format
- Merge where appropriate
- Maintain traceability to original SBOMs
- Generate combined SBOM if from same application

[Paste multiple SBOMs]
```

### Format Migration with Enhancement
```
Convert this SBOM to [target format] and enhance with:
- Current vulnerability data from OSV.dev
- Updated license information
- Component metadata enrichment
- Supply chain provenance data

[Paste SBOM]
```

## Examples

### Example Usage
```
Please convert this SBOM to SPDX 2.3:

Source Format: CycloneDX
Source Version: 1.4
Target Format: SPDX
Target Version: 2.3
Output Format: JSON

[CycloneDX 1.4 SBOM]
```

### Example Output Structure
```json
{
  "spdxVersion": "SPDX-2.3",
  "dataLicense": "CC0-1.0",
  "SPDXID": "SPDXRef-DOCUMENT",
  "name": "Converted from CycloneDX",
  "documentNamespace": "...",
  "creationInfo": {
    "created": "2024-11-20T12:00:00Z",
    "creators": ["Tool: SBOM Analyser"],
    "comment": "Converted from CycloneDX 1.4"
  },
  "packages": [...],
  "relationships": [...]
}
```

**Conversion Report:**
```markdown
# Conversion Report

## Summary
- Source: CycloneDX 1.4 JSON
- Target: SPDX 2.3 JSON
- Components Converted: 25/25
- Relationships Mapped: 48/48
- Validation: PASSED

## Field Mappings
✅ Components → Packages (25)
✅ Dependencies → Relationships (48)
✅ Licenses → SPDX Expressions (25)
⚠️  Services → Packages with annotations (3)
⚠️  Formulation → Annotations (build info preserved)

## Format-Specific Handling
- CycloneDX services converted to SPDX packages with type annotation
- CycloneDX formulation data preserved in annotations
- Vulnerability data extracted to external references
- Component pedigree maintained in package comments

## Recommendations
- Review service mappings for accuracy
- Consider adding SPDX file-level details
- Validate external references point to correct resources
```

## Tips
- Specify exact target version for validation
- Request validation report to ensure conversion accuracy
- Ask for detailed mapping report for audit trails
- Consider format-specific features that may need manual review
- Test converted SBOMs with target tools before deployment
- Maintain original SBOMs for reference

## Common Conversion Scenarios

### Vendor Integration
Converting supplier SBOMs to internal format:
```
Our vendors provide SPDX 2.2. Convert to CycloneDX 1.7 for our analysis pipeline:
- Standardize format
- Add our metadata
- Validate against our schema requirements

[Vendor SPDX SBOM]
```

### Tool Compatibility
Enabling specific tool usage:
```
This tool requires SPDX Tag-Value format. Convert our CycloneDX SBOM:
- Maintain all compliance data
- Preserve file checksums
- Ensure tool can parse output

[CycloneDX SBOM]
```

### Compliance Requirements
Meeting regulatory format requirements:
```
Compliance requires SPDX format. Convert while ensuring:
- All license info preserved
- File-level data included
- Supplier information complete
- Checksums validated

[CycloneDX SBOM]
```

## Validation

After conversion, always validate:
- Schema compliance (validate against official schemas)
- Data completeness (no critical data lost)
- Relationship integrity (dependencies preserved)
- License expression validity (SPDX identifiers correct)
- PURL consistency (package URLs match)
- Metadata accuracy (creation info, tools, timestamps)

## Related Prompts
- [version-upgrade.md](version-upgrade.md) - Upgrade SBOM versions
- [sbom-merge.md](sbom-merge.md) - Merge multiple SBOMs
- [validation.md](validation.md) - Validate SBOM correctness
