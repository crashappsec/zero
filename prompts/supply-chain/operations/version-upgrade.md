<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Prompt: SBOM Version Upgrade

## Purpose
Upgrade SBOMs from older versions to current specifications while adding new required fields and preserving existing data.

## When to Use
- Modernizing legacy SBOMs
- Leveraging new specification features
- Ensuring compatibility with current tools
- Meeting updated compliance requirements
- Improving SBOM completeness and richness

## Prompt

```
Please upgrade this SBOM to the latest version:

Current Format: [CycloneDX/SPDX]
Current Version: [version]
Target Version: [latest or specific version]

Requirements:
1. Upgrade to target version specification
2. Add all new required fields
3. Preserve all existing data
4. Enhance with new optional fields where possible
5. Validate against target schema
6. Document changes in annotations
7. Provide upgrade report showing:
   - Fields added
   - Structure changes
   - New capabilities enabled
   - Recommendations for further enhancement

[Paste SBOM content here]
```

## Expected Output
- Upgraded SBOM in target version
- Schema validation confirmation
- Upgrade report detailing:
  - Version changes
  - New fields added
  - Structure modifications
  - Suggested enhancements
- Migration notes for future reference

## Variations

### CycloneDX Version Upgrade
```
Upgrade this CycloneDX SBOM from version [X] to 1.7:

Enhancements to include:
- Add serialNumber and version if missing
- Add compositions for completeness tracking
- Include formulation for build processes
- Add annotations with signature support
- Include definitions and declarations
- Add citations for data sources
- Update vulnerability format to latest schema
- Enhance metadata with current tool info

[Paste CycloneDX SBOM]
```

### SPDX Version Upgrade
```
Upgrade this SPDX document from version [X] to 2.3:

Enhancements to include:
- Add new relationship types
- Update external references for security
- Add package verification enhancements
- Update annotation structures
- Add supplier and originator fields where missing
- Include security-related external references

[Paste SPDX document]
```

### Upgrade with Data Enrichment
```
Upgrade this SBOM to latest version and enrich with:
- Current vulnerability data
- Updated dependency information
- Enhanced license details
- Provenance and attestation data
- OpenSSF Scorecard metrics
- Supply chain security metadata

[Paste SBOM]
```

### Batch Upgrade
```
Upgrade these multiple SBOMs to latest versions:
- Standardize all to same target version
- Apply consistent enhancement policies
- Generate migration summary across all SBOMs
- Flag any requiring manual review

[Paste multiple SBOMs]
```

## Examples

### Example Usage: CycloneDX 1.2 → 1.7
```
Please upgrade this SBOM to the latest version:

Current Format: CycloneDX
Current Version: 1.2
Target Version: 1.7

[CycloneDX 1.2 SBOM]
```

### Example Output Structure

**Upgraded SBOM:**
```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.7",
  "serialNumber": "urn:uuid:generated-uuid",
  "version": 1,
  "metadata": {
    "timestamp": "2024-11-20T12:00:00Z",
    "tools": [{
      "vendor": "SBOM Analyser",
      "name": "Version Upgrade Tool",
      "version": "1.0"
    }],
    "component": {...}
  },
  "components": [...],
  "dependencies": [...],
  "compositions": [{
    "aggregate": "complete",
    "assemblies": ["..."],
    "dependencies": ["..."]
  }],
  "annotations": [{
    "text": "Upgraded from CycloneDX 1.2 to 1.7 on 2024-11-20",
    "timestamp": "2024-11-20T12:00:00Z"
  }]
}
```

**Upgrade Report:**
```markdown
# SBOM Upgrade Report

## Summary
- Original Version: CycloneDX 1.2
- Target Version: CycloneDX 1.7
- Validation: PASSED
- Components: 25
- Upgrade Date: 2024-11-20

## Changes Applied

### New Required Fields
✅ Added serialNumber: urn:uuid:generated-uuid
✅ Added version: 1
✅ Enhanced metadata structure

### New Sections Added
✅ Compositions (completeness tracking)
⚠️  Formulation (build info - requires manual population)
⚠️  Declarations (attestations - requires manual population)
✅ Annotations (upgrade documentation)

### Structure Updates
✅ Dependencies format validated
✅ Component types updated to 1.7 taxonomy
✅ Vulnerability schema updated
✅ License expressions validated against SPDX

### Preserved Data
✅ All 25 components preserved
✅ All 48 dependencies maintained
✅ License information intact
✅ Metadata preserved

## New Capabilities Enabled

### Compositions
You can now track SBOM completeness:
- Aggregate type: complete/incomplete/unknown
- Component assemblies
- Dependency relationships

### Formulation
Build and deployment processes can be documented:
- Build workflows
- Manufacturing details
- Deployment configurations

### Declarations
Standards compliance can be attested:
- Conformance claims
- Evidence and attestations
- Regulatory compliance

### Citations
Data source attribution:
- Vulnerability sources
- License information sources
- Component data origins

## Recommendations for Enhancement

### High Priority
1. **Add Formulation Data**: Document build processes
   - CI/CD pipeline information
   - Build environment details
   - Artifact generation steps

2. **Complete Compositions**: Define completeness
   - Mark as complete/incomplete
   - Identify missing components

### Medium Priority
3. **Add Declarations**: Include compliance claims
   - Standards conformance
   - Security attestations
   - Regulatory compliance

4. **Enrich Metadata**: Add missing information
   - Supplier details
   - Manufacturer information
   - Additional tool information

### Low Priority
5. **Add Citations**: Document data sources
   - Vulnerability data sources
   - License information sources
   - Component metadata origins

## Manual Review Required

⚠️  **Formulation Section**: Requires build process details
⚠️  **Declarations**: Requires compliance attestations
⚠️  **Some Component Metadata**: May benefit from enrichment

## Validation Results

✅ Schema validation PASSED
✅ All required fields present
✅ PURL format valid
✅ License expressions valid (SPDX)
✅ Dependency graph complete

## Migration Notes

- Original SBOM archived with reference
- Upgrade annotated in SBOM metadata
- No data loss during upgrade
- All relationships preserved
- Consider implementing formulation in build pipeline
```

## Version-Specific Upgrade Guides

### CycloneDX Upgrades

**1.0-1.3 → 1.7:**
Major changes, extensive enhancements needed

**1.4 → 1.7:**
Moderate changes, add formulation, annotations, definitions, declarations

**1.5 → 1.7:**
Minor changes, add definitions, declarations, citations

**1.6 → 1.7:**
Minimal changes, add citations, enhance declarations

### SPDX Upgrades

**2.0/2.1 → 2.3:**
Moderate changes, relationship types, security references, enhanced metadata

**2.2 → 2.3:**
Minor changes, security references, new annotation categories

## Best Practices

### Before Upgrade
- Back up original SBOM
- Validate current version
- Review target version features
- Plan data enrichment

### During Upgrade
- Preserve all existing data
- Add required fields with defaults
- Document assumptions
- Validate against schema

### After Upgrade
- Test with analysis tools
- Review enhancement opportunities
- Document migration
- Update SBOM generation pipeline

## Handling Missing Data

When required fields are missing:

**SPDX:**
- Use "NOASSERTION" for unknown values
- Use "NONE" when explicitly absent
- Document in comments

**CycloneDX:**
- Use "Unknown" for string fields
- Generate UUIDs for serialNumber
- Use current timestamp if missing
- Document in annotations

## Automated Upgrade Integration

For CI/CD integration:
```
Create an automated upgrade workflow:
1. Detect SBOM version
2. Upgrade to latest
3. Enrich with available data
4. Validate output
5. Generate upgrade report
6. Commit upgraded SBOM

Include this in our build pipeline for continuous SBOM modernization.
```

## Tips
- Always validate upgraded SBOMs
- Keep original SBOMs for reference
- Document upgrade rationale
- Test with target tools
- Consider gradual enrichment over time
- Automate upgrades in build pipeline

## Related Prompts
- [format-conversion.md](format-conversion.md) - Convert between formats
- [sbom-enrichment.md](sbom-enrichment.md) - Add additional data
- [validation.md](validation.md) - Validate SBOM correctness
