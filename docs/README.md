<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Documentation

Comprehensive documentation for the Zero analysis toolkit.

## Architecture

Understanding how the system works:

| Document | Description |
|----------|-------------|
| [System Overview](architecture/overview.md) | High-level architecture and component relationships |
| [Knowledge Base](architecture/knowledge-base.md) | Knowledge organization within agents |
| [Zero Architecture](plans/zero-architecture.md) | Master orchestrator design |

## Agents (Hackers-Themed)

Self-contained, portable AI agents named after characters from Hackers (1995):

| Agent | Character | Directory | Expertise |
|-------|-----------|-----------|-----------|
| **Zero** | Zero Cool | (orchestrator) | Master coordinator |
| **Cereal** | Cereal Killer | [cereal/](../agents/cereal/) | Supply chain security, CVEs, malcontent |
| **Razor** | Razor | [razor/](../agents/razor/) | Static analysis, secrets, SAST |
| **Blade** | Blade | [blade/](../agents/blade/) | Compliance, SOC 2, ISO 27001 |
| **Phreak** | Phantom Phreak | [phreak/](../agents/phreak/) | Legal, licenses, data privacy |
| **Acid** | Acid Burn | [acid/](../agents/acid/) | React, TypeScript, accessibility |
| **Dade** | Dade Murphy | [dade/](../agents/dade/) | APIs, databases, data engineering |
| **Nikon** | Lord Nikon | [nikon/](../agents/nikon/) | System design, architecture patterns |
| **Joey** | Joey | [joey/](../agents/joey/) | CI/CD optimization, build performance |
| **Plague** | The Plague | [plague/](../agents/plague/) | Infrastructure, Kubernetes, operations |
| **Gibson** | The Gibson | [gibson/](../agents/gibson/) | DORA metrics, team effectiveness |

See [Agents README](../agents/README.md) for full documentation.

## Scanners

| Scanner | Directory | Description |
|---------|-----------|-------------|
| Tech Discovery | [utils/scanners/tech-discovery](../utils/scanners/tech-discovery/) | Technology stack identification (100+ patterns) |
| Vulnerabilities | [utils/scanners/vulnerabilities](../utils/scanners/vulnerabilities/) | CVE scanning via OSV |
| Package Malcontent | [utils/scanners/package-malcontent](../utils/scanners/package-malcontent/) | Supply chain compromise detection |
| Package Health | [utils/scanners/package-health](../utils/scanners/package-health/) | Dependency health and abandonment |
| Licenses | [utils/scanners/licenses](../utils/scanners/licenses/) | SPDX license analysis |
| Code Security | [utils/scanners/code-security](../utils/scanners/code-security/) | Static analysis findings |
| Secrets Scanner | [utils/scanners/secrets-scanner](../utils/scanners/secrets-scanner/) | Secret detection |
| Package SBOM | [utils/scanners/package-sbom](../utils/scanners/package-sbom/) | CycloneDX SBOM via Syft |
| DORA | [utils/scanners/dora](../utils/scanners/dora/) | DORA metrics calculation |
| Code Ownership | [utils/scanners/code-ownership](../utils/scanners/code-ownership/) | Contributor analysis |

## Reference Documents

| Document | Description |
|----------|-------------|
| [Competitive Analysis](competitive-analysis-dpi-tools.md) | DPI tools market analysis |
