# CycloneDX SBOM Specification - Technical Reference

## Overview

**CycloneDX** is a lightweight Software Bill of Materials (SBOM) standard designed for use in application security contexts and supply chain component analysis.

**Current Version**: 1.7 (ECMA-424 Standard)
**Website**: https://cyclonedx.org
**Specification**: https://cyclonedx.org/specification/overview/

## Key Features

- **Lightweight**: Designed for real-time analysis and automation
- **Security-focused**: Built for vulnerability management
- **Extensible**: Supports custom properties and extensions
- **Multi-format**: JSON, XML, Protocol Buffers

## Core Structure

### BOM Format

```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.7",
  "serialNumber": "urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79",
  "version": 1,
  "metadata": {
    "timestamp": "2024-11-21T10:00:00Z",
    "tools": [{
      "name": "syft",
      "version": "1.0.0"
    }],
    "component": {
      "type": "application",
      "name": "my-application",
      "version": "1.0.0"
    }
  },
  "components": [
    {
      "type": "library",
      "name": "express",
      "version": "4.17.1",
      "purl": "pkg:npm/express@4.17.1",
      "hashes": [{
        "alg": "SHA-256",
        "content": "abc123..."
      }],
      "licenses": [{
        "license": {
          "id": "MIT"
        }
      }]
    }
  ],
  "dependencies": [
    {
      "ref": "pkg:npm/express@4.17.1",
      "dependsOn": ["pkg:npm/body-parser@1.19.0"]
    }
  ],
  "vulnerabilities": [
    {
      "id": "CVE-2024-1234",
      "source": {
        "name": "NVD",
        "url": "https://nvd.nist.gov/"
      },
      "ratings": [{
        "score": 7.5,
        "severity": "high",
        "method": "CVSSv3"
      }],
      "affects": [{
        "ref": "pkg:npm/express@4.17.1"
      }]
    }
  ]
}
```

## Component Types

- **application**: Root application
- **framework**: Software framework
- **library**: Third-party library
- **container**: Container image
- **operating-system**: OS package
- **device**: Hardware component
- **firmware**: Device firmware
- **file**: Individual file

## Package URL (purl)

CycloneDX uses **purl** for component identification:

```
pkg:npm/express@4.17.1
pkg:pypi/django@3.2
pkg:maven/org.springframework.boot/spring-boot-starter@2.5.0
pkg:docker/library/nginx@sha256:abc123
pkg:golang/github.com/gin-gonic/gin@v1.7.0
```

## Vulnerability Tracking

CycloneDX includes native vulnerability tracking:

```json
{
  "vulnerabilities": [
    {
      "id": "CVE-2024-1234",
      "source": {"name": "OSV"},
      "ratings": [{
        "score": 7.5,
        "severity": "high",
        "method": "CVSSv3",
        "vector": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N"
      }],
      "cwes": [79],
      "description": "Cross-site scripting vulnerability",
      "recommendation": "Upgrade to version 4.18.0 or later",
      "affects": [{
        "ref": "pkg:npm/express@4.17.1",
        "versions": [{
          "version": "4.17.1",
          "status": "affected"
        }]
      }],
      "published": "2024-01-15T00:00:00Z",
      "updated": "2024-01-20T00:00:00Z"
    }
  ]
}
```

## Tools for CycloneDX

### Generation
- **syft**: Multi-language SBOM generator
- **cdxgen**: CycloneDX generator for multiple ecosystems
- **trivy**: Container and filesystem scanner
- **grype**: Vulnerability scanner

### Validation
- **cyclonedx-cli**: Official CLI tool
- **sbom-tool**: Microsoft SBOM tool

### Analysis
- **dependency-track**: SBOM analysis platform
- **bomber**: Vulnerability scanner for SBOMs

## Common Operations

### Generate SBOM
```bash
# Using syft
syft . -o cyclonedx-json=sbom.json

# Using cdxgen
cdxgen -o sbom.json .

# Using Docker
syft docker.io/library/nginx:latest -o cyclonedx-json
```

### Validate SBOM
```bash
cyclonedx-cli validate --input-file sbom.json
```

### Merge SBOMs
```bash
cyclonedx-cli merge \
  --input-files sbom1.json sbom2.json \
  --output-file merged.json
```

## Version History

- **1.7** (Current): Enhanced provenance, attestations
- **1.6**: Improved vulnerability tracking
- **1.5**: Added formulation, licensing enhancements
- **1.4**: Services, compositions
- **1.3**: External references, evidence
- **1.2**: Dependency graph
- **1.1**: Swid, pedigree
- **1.0**: Initial release

## Best Practices

1. **Always include purl**: Enables accurate component matching
2. **Use SHA-256 hashes**: Ensures integrity verification
3. **Track dependencies**: Include dependency graph
4. **Update regularly**: Regenerate SBOM on changes
5. **Include metadata**: Document tools and timestamps
6. **Sign SBOMs**: Cryptographically sign for integrity

## Integration Points

### CI/CD Integration
```yaml
# GitHub Actions
- name: Generate SBOM
  run: syft . -o cyclonedx-json=sbom.json

- name: Upload SBOM
  uses: actions/upload-artifact@v3
  with:
    name: sbom
    path: sbom.json
```

### Container Registries
- Store SBOM as OCI artifact
- Attach to container images
- Sign with cosign

### Vulnerability Databases
- **OSV.dev**: Query by purl
- **NVD**: CVE cross-reference
- **GitHub Advisory**: Security advisories

## References

- **Specification**: https://cyclonedx.org/specification/overview/
- **Schema**: https://cyclonedx.org/docs/1.7/json/
- **ECMA Standard**: https://ecma-international.org/publications-and-standards/standards/ecma-424/
- **Tool Center**: https://cyclonedx.org/tool-center/
