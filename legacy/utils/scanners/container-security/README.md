# Container Security Scanner

Analyzes Dockerfiles for best practices, detects hardened base images, scans for vulnerabilities, and provides optimization recommendations.

## Features

- **Dockerfile Best Practices**: Linting with hadolint (or regex fallback)
- **Base Image Hardening**: Detects Chainguard, Distroless, Alpine usage
- **Multi-stage Build Analysis**: Identifies optimization opportunities
- **Vulnerability Scanning**: Trivy/Grype integration for CVE detection
- **SBOM Generation**: Syft integration for package inventory

## Usage

```bash
# Analyze local project
./container-security.sh /path/to/project

# Analyze with output file
./container-security.sh -o container-security.json /path/to/project

# Use cached repository from Zero
./container-security.sh --repo expressjs/express

# Include image vulnerability scanning (slower)
./container-security.sh --scan-images /path/to/project
```

## Options

| Option | Description |
|--------|-------------|
| `--local-path PATH` | Use pre-cloned repository |
| `--repo OWNER/REPO` | GitHub repository (from Zero cache) |
| `--org ORG` | GitHub org (first repo in Zero cache) |
| `-o, --output FILE` | Write JSON to file (default: stdout) |
| `--scan-images` | Scan container images with trivy/grype |
| `-k, --keep-clone` | Keep cloned repository |
| `-h, --help` | Show help |

## Output Format

```json
{
  "analyzer": "container-security",
  "version": "1.0.0",
  "timestamp": "2025-12-08T12:00:00Z",
  "target": "/path/to/project",
  "summary": {
    "dockerfiles_found": 2,
    "images_analyzed": 1,
    "total_vulnerabilities": 45,
    "by_severity": {"critical": 2, "high": 8, "medium": 20, "low": 15},
    "dockerfile_issues": 12,
    "uses_multistage": true,
    "uses_hardened_base": false,
    "hardening_score": 65
  },
  "dockerfiles": [...],
  "hardening_analysis": [...],
  "multistage_analysis": [...],
  "images": [...],
  "recommendations": [...]
}
```

## External Tools

The scanner works in degraded mode without external tools, but provides enhanced analysis when available:

| Tool | Purpose | Installation |
|------|---------|--------------|
| **hadolint** | Dockerfile linting | `brew install hadolint` |
| **trivy** | Vulnerability scanning | `brew install trivy` |
| **grype** | Alternative vuln scanner | `brew install grype` |
| **syft** | SBOM generation | `brew install syft` |

## Dockerfile Analysis

### Best Practices Checked

| Check | Severity | Description |
|-------|----------|-------------|
| `NO_USER` | warning | Container runs as root |
| `NO_HEALTHCHECK` | info | Missing HEALTHCHECK instruction |
| `USE_COPY` | warning | Using ADD instead of COPY |
| `PIN_VERSIONS` | warning | Unpinned package versions |
| `SENSITIVE_FILES` | error | Copying .env, credentials, keys |
| `NO_WORKDIR` | info | Missing WORKDIR instruction |

With hadolint installed, 100+ additional rules are checked.

## Hardening Detection

### Image Classification

| Type | Security Rating | Detection |
|------|-----------------|-----------|
| Chainguard | Very High | `cgr.dev/chainguard/*` prefix |
| Distroless | Very High | `gcr.io/distroless/*` prefix |
| Scratch | Very High | `scratch` base |
| Alpine | High | `-alpine` suffix |
| Slim | Medium | `-slim` suffix |
| Standard | Low | Everything else |

### Hardening Score

Calculated as weighted average of base image types:
- Chainguard/Distroless/Scratch: 100 points
- Alpine: 70 points
- Slim: 50 points
- Standard: 20 points

## Multi-stage Analysis

Detects:
- Number of build stages
- Stage purposes (build, test, runtime)
- COPY --from references
- Build tool usage
- Artifact leakage (build deps in final image)

## Library Files

- `lib/dockerfile-analyzer.sh` - Dockerfile parsing and best practices
- `lib/hardened-detector.sh` - Chainguard/Distroless detection
- `lib/multistage-analyzer.sh` - Multi-stage build analysis
- `lib/image-scanner.sh` - Trivy/Grype/Syft integration

## Dependencies

- `bash` 4.0+
- `jq` - JSON processing
- `grep`, `sed`, `awk` - Text processing

Optional:
- `hadolint` - Enhanced Dockerfile linting
- `trivy` - Vulnerability scanning
- `grype` - Alternative vulnerability scanning
- `syft` - SBOM generation
- `docker` - Image inspection

## License

GPL-3.0 - Crash Override Inc.
