<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Documentation

Comprehensive documentation for the Gibson Powers analysis toolkit.

## Architecture

Understanding how the system works:

| Document | Description |
|----------|-------------|
| [System Overview](architecture/overview.md) | High-level architecture and component relationships |
| [Knowledge Base](architecture/knowledge-base.md) | Knowledge organization within agents |

## Agents

Self-contained, portable AI agents:

| Agent | Description |
|-------|-------------|
| [supply-chain](../agents/supply-chain/) | Supply chain security, vulnerabilities, licenses |
| [code-security](../agents/code-security/) | Static analysis, secrets, code vulnerabilities |
| [frontend-engineer](../agents/frontend-engineer/) | React, TypeScript, web app development |
| [backend-engineer](../agents/backend-engineer/) | APIs, databases, data engineering |
| [architect](../agents/architect/) | System design, patterns, auth frameworks |
| [build-engineer](../agents/build-engineer/) | CI/CD optimization, build performance |
| [devops-engineer](../agents/devops-engineer/) | Deployments, infrastructure, operations |
| [engineering-leader](../agents/engineering-leader/) | Costs, metrics, team effectiveness |

See [Agents README](../agents/README.md) for full documentation.

## CLI Tools

| Tool | Description |
|------|-------------|
| [Supply Chain Scanner](../utils/supply-chain/README.md) | Dependency vulnerability scanning |
| [Code Security Scanner](../utils/code-security/README.md) | Static security analysis |
| [Technology Identification](../utils/technology-identification/README.md) | Detect frameworks and services |
| [Legal Review](../utils/legal-review/README.md) | License compliance analysis |
| [Code Ownership](../utils/code-ownership/README.md) | CODEOWNERS analysis |
| [DORA Metrics](../utils/dora-metrics/README.md) | DevOps performance metrics |

## Implementation Plans

| Document | Description |
|----------|-------------|
| [Specialist Agents Plan](plans/specialist-agents-plan.md) | Agent framework design |
| [Supply Chain Enhancement](supply-chain-enhancement-plan.md) | Scanner improvements |
| [Supply Chain Implementation](supply-chain-implementation-plan.md) | Detailed implementation |
| [Code Security Implementation](code-security-implementation-plan.md) | Security scanner plan |
| [Code Ownership Implementation](code-ownership-implementation-plan.md) | CODEOWNERS analysis |
| [RAG Catalog Plan](rag-catalog-plan.md) | Technology identification |

## Analysis Reports

| Document | Description |
|----------|-------------|
| [Competitive Analysis](competitive-analysis-dpi-tools.md) | DPI tools market analysis |
| [Technology ID Audit](technology-identification-audit.md) | Detection coverage |
| [RAG Catalog (100 Technologies)](rag-catalog-100-technologies.md) | Technology patterns |
