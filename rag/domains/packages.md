# Packages Domain Knowledge

This document consolidates RAG knowledge for the **packages** super scanner.

## Features Covered
- **sbom**: SBOM generation and analysis
- **vulnerabilities**: CVE/vulnerability scanning
- **health**: Package maintenance and health metrics
- **malcontent**: Supply chain compromise detection
- **provenance**: Build provenance and attestations
- **bundle**: Bundle size optimization
- **licenses**: License compliance
- **duplicates**: Duplicate dependency detection
- **recommendations**: Package alternatives
- **typosquats**: Typosquatting detection
- **deprecations**: Deprecated package detection

## Related RAG Directories

### Supply Chain Security
- `rag/supply-chain/` - Core supply chain security knowledge
  - Package malware detection
  - YARA rules for suspicious code patterns
  - Differential analysis techniques

### Legal/Licenses
- `rag/legal-review/` - License compliance knowledge
  - License compatibility matrices
  - Risk assessment for different licenses

## Key Concepts

### SBOM (Software Bill of Materials)
- CycloneDX format (version 1.5 preferred)
- Component identification via purl (Package URL)
- Tools: cdxgen, syft

### Vulnerability Scanning
- OSV database for vulnerability lookup
- NVD/CVE correlation
- EPSS scoring for exploitability

### Package Health Metrics
- Maintenance activity (commit frequency, release cadence)
- Download trends
- Issue response times
- Deprecation status

### Supply Chain Compromise Detection
- Malcontent YARA rules
- Network behavior analysis
- Obfuscation detection
- Post-install script analysis

### Build Provenance
- SLSA levels (1-4)
- Sigstore verification
- NPM provenance attestations

## Agent Expertise

### Cereal Agent
The **Cereal** agent (supply chain specialist) should be consulted for:
- Malcontent findings investigation
- Supply chain risk assessment
- Package health analysis
- Typosquatting detection

### Phreak Agent
The **Phreak** agent (legal counsel) should be consulted for:
- License compatibility analysis
- Legal risk assessment
- Compliance guidance

## Output Schema

The packages scanner produces a single `packages.json` file with:
```json
{
  "features_run": ["sbom", "vulnerabilities", "health", ...],
  "summary": {
    "sbom": { "total_components": N, "by_type": {...} },
    "vulnerabilities": { "total": N, "critical": N, "high": N, ... },
    "health": { "total_packages": N, "unhealthy": N, ... },
    ...
  },
  "findings": {
    "sbom": { "components": [...] },
    "vulnerabilities": [...],
    "health": [...],
    ...
  }
}
```

## Severity Classification

| Finding Type | Critical | High | Medium | Low |
|--------------|----------|------|--------|-----|
| Vulnerability | CVSS 9.0+ | CVSS 7.0-8.9 | CVSS 4.0-6.9 | CVSS < 4.0 |
| Malcontent | Malware, backdoor | Suspicious network, obfuscation | Data collection | Info |
| License | GPL in proprietary | LGPL ambiguity | Attribution required | Permissive |
| Health | Abandoned + vulns | Abandoned | Low activity | - |
