# Zero Engineering Roadmap

**Version**: 3.7.0
**Last Updated**: 2025-12-24

This document tracks planned and in-progress features. Completed work is documented in the Appendix.

---

## Implementation Status Legend

| Status | Description |
|--------|-------------|
| IN_PROGRESS | Active development |
| PLANNED | Designed but not started |
| FUTURE | Roadmap item, not designed yet |

---

## 1. Active Work Queue

### 1.1 High Priority

| Feature | Status | Description |
|---------|--------|-------------|
| Reachability Analysis | PLANNED | Vulnerable code path detection |

### 1.2 Medium Priority

| Feature | Status | Description |
|---------|--------|-------------|
| Dependency Graph Visualization | PLANNED | Interactive dependency explorer |
| Circular Dependency Detection | PLANNED | Find problematic cycles |
| Training Data PII Analysis | PLANNED | PII detection in ML datasets |
| Cloud Asset Inventory | PLANNED | AWS/Azure/GCP resource discovery |
| Cloud SBOM Generation | PLANNED | CycloneDX for cloud resources |
| Ocular Integration | PLANNED | Code sync and orchestration |
| Chalk Integration | PLANNED | Build-time attestation |
| GitHub/GitLab Org Analysis | PLANNED | Repository security audit |

### 1.3 Low Priority / Future

| Feature | Status | Description |
|---------|--------|-------------|
| Layer Violation Detection | FUTURE | Architecture rule violations |
| Database Schema Analysis | FUTURE | Migration risks, schema drift |
| Jupyter Notebook Security | FUTURE | Secrets in .ipynb files |
| Runtime vs Build-time SBOM | FUTURE | Compare deployed vs source SBOMs |
| Certificate Monitoring | FUTURE | Live SSL/TLS certificate expiry |
| DNS Security | FUTURE | DNSSEC, SPF, DKIM, DMARC |
| Database Backend | FUTURE | SQLite/DuckDB/PostgreSQL |
| PDF Export | FUTURE | Export reports as PDF |
| Trend Analysis | FUTURE | Historical comparison |

---

## 2. Partial Implementations (Gaps to Address)

### 2.1 Package Analysis Scanner

| Gap | Description |
|-----|-------------|
| Reachability analysis | Not implemented |

### 2.2 Code Quality Scanner

| Gap | Description |
|-----|-------------|
| Test coverage | Coverage report parsing is basic |

### 2.3 Code Security Scanner

| Gap | Description |
|-----|-------------|
| AI false positive reduction | Claude-powered FP detection is partial |

---

## 3. File Reference

### Key Implementation Files

| Area | Files |
|------|-------|
| Scanners | `pkg/scanner/*/` (9 scanner packages) |
| RAG Patterns | `rag/*/` (technology-identification, devops, etc.) |
| Rule Generator | `pkg/core/rules/manager.go` |
| Live APIs | `pkg/core/liveapi/*.go` |
| CycloneDX | `pkg/core/cyclonedx/*.go` |
| Reports | `reports/template/` |
| Config | `config/zero.config.json` |

### Documentation

| Document | Purpose |
|----------|---------|
| `CLAUDE.md` | Agent instructions, scanner architecture |
| `ROADMAP.md` | Public feature roadmap |
| `docs/PATTERN-ARCHITECTURE.md` | RAG pattern architecture |
| `docs/EVIDENCE-INTEGRATION-PLAN.md` | Report system design |
| `docs/devex-implementation-plan.md` | DevEx scanner design |

---

## 4. Contributing

To contribute to any roadmap item:

1. Check the implementation status in this document
2. Review related documentation in `docs/`
3. Create an issue or pick up existing one
4. Follow patterns in `CLAUDE.md` and `PATTERN-ARCHITECTURE.md`
5. Submit PR with tests

---

# Appendix: Completed Features

This section documents completed work for historical reference.

## A. Core Scanners (9 Super Scanners)

### A.1 SBOM Scanner - COMPLETE

| Feature | Description |
|---------|-------------|
| CycloneDX generation | Generates sbom.cdx.json |
| Multi-format SBOM | Supports CycloneDX and SPDX |
| Package detection | npm, pip, go, cargo, maven, etc. |
| Integrity checking | Lock file integrity validation |

### A.2 Package Analysis Scanner - MOSTLY COMPLETE

| Feature | Description |
|---------|-------------|
| Vulnerability scanning | Via OSV.dev live API |
| License detection | SPDX license identification |
| Malcontent analysis | Supply chain threat detection |
| Typosquat detection | Detects typosquatted packages |
| Bundle analysis | Size impact analysis |
| Duplicate detection | Finds duplicate packages |

### A.3 Crypto Scanner - COMPLETE

| Feature | Description |
|---------|-------------|
| Cipher detection | Weak/deprecated cipher detection |
| Key detection | Hardcoded cryptographic keys |
| Random analysis | Insecure random number generation |
| TLS configuration | TLS version and cipher suite analysis |
| Certificate analysis | Certificate validation issues |
| CBOM export | CycloneDX Cryptography BOM |

### A.4 Code Security Scanner - COMPLETE

| Feature | Description |
|---------|-------------|
| SAST scanning | Via Semgrep integration |
| Secrets detection | Semgrep p/secrets + entropy analysis |
| Git history scanning | Secrets in git history |
| API security | Auth, injection, SSRF, CORS checks |
| IaC secrets detection | Secrets in Terraform, K8s, etc. |
| Rotation guidance | Secret rotation recommendations |

### A.5 Code Quality Scanner - MOSTLY COMPLETE

| Feature | Description |
|---------|-------------|
| Tech debt detection | TODO/FIXME counting |
| Complexity analysis | Cyclomatic complexity |
| Documentation analysis | Doc comment coverage |

### A.6 DevOps Scanner - COMPLETE

| Feature | Description |
|---------|-------------|
| IaC scanning | Checkov/Trivy integration |
| Container security | Dockerfile linting, image scanning |
| GitHub Actions | Action pinning, secrets, permissions |
| DORA metrics | Deployment frequency, lead time, MTTR, CFR |
| Git insights | Activity, contributors, patterns |
| IaC secrets scanning | Semgrep-based secrets in IaC |
| IaC organizational policies | RAG patterns for Terraform, K8s, CloudFormation |

### A.7 Tech-ID Scanner - COMPLETE

| Feature | Description |
|---------|-------------|
| Technology detection | Languages, frameworks, tools |
| ML model detection | PyTorch, TensorFlow, ONNX models |
| Dataset detection | CSV, Parquet, HuggingFace datasets |
| AI security analysis | Pickle vulnerabilities, exposed keys |
| AI governance | Model cards, responsible AI checks |
| ML-BOM export | CycloneDX Machine Learning BOM |
| Infrastructure detection | Docker, K8s, Terraform detection |
| Microservice mapping | Service-to-service communication detection |

### A.8 Code Ownership Scanner - COMPLETE

| Feature | Description |
|---------|-------------|
| Contributor analysis | Author statistics |
| Bus factor calculation | Key person risk |
| CODEOWNERS validation | CODEOWNERS file analysis |
| Orphan detection | Files without active maintainers |
| Churn analysis | High-churn file detection |
| Pattern analysis | Commit patterns, timing |

### A.9 Developer Experience Scanner - COMPLETE

| Feature | Description |
|---------|-------------|
| Onboarding analysis | README quality, setup friction |
| Tool sprawl | Development tool complexity |
| Technology sprawl | Learning curve estimation |
| Workflow efficiency | PR templates, local dev, hot reload |

## B. Core Infrastructure

### B.1 RAG Pattern System - COMPLETE

| Feature | Description |
|---------|-------------|
| Tech-ID patterns | Technology detection via RAG |
| Docker patterns | Dockerfile security patterns |
| Secrets patterns | Secret detection patterns |
| IaC policies | Terraform, K8s, CloudFormation policies |
| Secrets-in-IaC | Hardcoded secrets in IaC files |
| Microservice patterns | HTTP, gRPC, message queue detection |
| API Quality patterns | Rate limiting, auth, CORS, injection |
| API Versioning patterns | Deprecated endpoints, sunset APIs |
| Rule generator | Extended for all categories |

### B.2 Live API Clients - COMPLETE

| Feature | Description |
|---------|-------------|
| OSV.dev client | Vulnerability data |
| deps.dev client | Package health, deprecation, SLSA provenance |
| Caching | In-memory with TTL |
| Rate limiting | Token bucket rate limiter |

### B.3 CycloneDX Export - COMPLETE

| Feature | Description |
|---------|-------------|
| Core BOM types | CycloneDX 1.6 ECMA-424 support |
| SBOM export | Standard software BOM |
| ML-BOM export | Machine learning BOM |
| CBOM export | Cryptography BOM |
| Exporter component | Reusable export module |

### B.4 Reports (Evidence.dev) - COMPLETE

| Feature | Description |
|---------|-------------|
| Evidence template | HTML report generation |
| Overview page | Executive summary dashboard |
| Security page | Vulnerabilities, secrets, crypto |
| Dependencies page | SBOM, licenses, packages |
| Supply chain page | Malcontent, health |
| DevOps page | DORA, IaC, containers |
| Quality page | Code quality, ownership |
| AI/ML page | ML models, AI security |

## C. Recently Completed (December 2024)

| Feature | Date | Description |
|---------|------|-------------|
| API Quality Patterns | 2024-12 | Rate limiting, auth, CORS, injection detection via RAG |
| API Versioning Audit | 2024-12 | Deprecated endpoints, sunset APIs, version detection |
| Rule Generator Extension | 2024-12 | Extended for devops-security, code-security, architecture |
| deps.dev Integration | 2024-12 | Wired into package scanner (health, deprecation, SLSA) |
| deps.dev Client | 2024-12 | Package health, deprecation, SLSA provenance |
| Semgrep IaC Enhancement | 2024-12 | Custom organizational policies, secrets-in-IaC |
| CycloneDX Export | 2024-12 | ML-BOM and CBOM export, reusable exporter |
| Microservice Mapping | 2024-12 | Service communication detection, API contracts |

---

*"Hack the planet!"*
