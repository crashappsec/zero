# Scanner Reference

Complete reference for all Zero scanners, organized by category.

## Code Scanners

Code scanners analyze source code for security issues, secrets, and patterns.

### code-vulns

**Purpose:** Static Application Security Testing (SAST)

Detects security vulnerabilities in source code including injection flaws, XSS, authentication issues, and OWASP Top 10 coverage.

```bash
./utils/scanners/code-vulns/code-vulns.sh /path/to/repo
```

**Options:**
| Flag | Description |
|------|-------------|
| `--local-path PATH` | Repository path |
| `--timeout SECONDS` | Timeout per file (default: 60) |
| `--verbose` | Show progress messages |
| `-o, --output FILE` | Write JSON output to file |

**Rules Used:**
- `p/security-audit` - Semgrep security audit rules
- `p/owasp-top-ten` - OWASP Top 10 patterns
- Custom rules from `crypto-security.yaml`

**Output:** `code-vulns.json`

---

### code-secrets

**Purpose:** Secret and credential detection

Detects exposed API keys, credentials, tokens, and private keys using 242+ patterns from RAG-generated rules.

```bash
./utils/scanners/code-secrets/code-secrets.sh /path/to/repo
```

**Options:**
| Flag | Description |
|------|-------------|
| `--local-path PATH` | Repository path |
| `--repo OWNER/REPO` | GitHub repository (uses Zero cache) |
| `--org ORG` | GitHub org (uses first cached repo) |
| `--no-community` | Skip community rules (faster, offline) |
| `--timeout SECONDS` | Timeout per file (default: 60) |
| `--verbose` | Show progress messages |
| `-o, --output FILE` | Write JSON output to file |

**Rules Used:**
- `secrets.yaml` - 242 RAG-generated patterns covering AWS, Azure, GCP, Stripe, OpenAI, and 100+ more
- `p/secrets` - Semgrep registry supplement

**Detected Secret Types:**
- AWS Access Keys and Secret Keys
- GitHub/GitLab Tokens (PAT, OAuth, App)
- Slack Tokens and Webhooks
- Stripe API Keys (live and test)
- Google Cloud Service Account Keys
- Private Keys (RSA, EC, DSA, PGP)
- Database Connection Strings
- JWT Secrets
- API Keys (generic patterns)

**Output:** `code-secrets.json`

---

### tech-discovery

**Purpose:** Technology stack detection

Identifies technologies, frameworks, and libraries used in the codebase.

```bash
./utils/scanners/tech-discovery/tech-discovery.sh /path/to/repo
```

**Rules Used:**
- `tech-discovery.yaml` - Import and package detection patterns

**Output:** `tech-discovery.json`

---

### tech-debt

**Purpose:** Technical debt markers

Finds TODO, FIXME, HACK, and XXX markers in code.

```bash
./utils/scanners/tech-debt/tech-debt.sh /path/to/repo
```

**Rules Used:**
- `tech-debt.yaml` - Marker detection patterns

**Output:** `tech-debt.json`

---

## Cryptography Scanners

Specialized scanners for cryptographic security analysis.

### crypto-ciphers

**Purpose:** Weak cipher detection

Detects deprecated and insecure cryptographic algorithms.

```bash
./utils/scanners/crypto-ciphers/crypto-ciphers.sh /path/to/repo
```

**Detected Issues:**
- DES, 3DES, Blowfish
- RC4 (ARC4)
- MD5, SHA1 for security purposes
- ECB mode encryption
- Weak key derivation functions

**Output:** `crypto-ciphers.json`

---

### crypto-keys

**Purpose:** Hardcoded key detection

Detects hardcoded cryptographic keys, weak key lengths, and exposed private keys.

```bash
./utils/scanners/crypto-keys/crypto-keys.sh /path/to/repo
```

**Detected Issues:**
- Hardcoded symmetric keys
- Embedded private keys (RSA, EC, DSA, PGP)
- Weak key lengths (< 2048 RSA, < 256 AES)
- Keys in configuration files

**Output:** `crypto-keys.json`

---

### crypto-random

**Purpose:** Insecure random detection

Detects insecure random number generation.

```bash
./utils/scanners/crypto-random/crypto-random.sh /path/to/repo
```

**Detected Issues:**
| Language | Insecure | Secure Alternative |
|----------|----------|--------------------|
| JavaScript | `Math.random()` | `crypto.randomBytes()` |
| Python | `random` module | `secrets`, `os.urandom()` |
| Java | `java.util.Random` | `SecureRandom` |
| Go | `math/rand` | `crypto/rand` |
| Ruby | `rand()` | `SecureRandom` |
| PHP | `rand()`, `mt_rand()` | `random_bytes()` |

**Output:** `crypto-random.json`

---

### crypto-tls

**Purpose:** TLS/SSL misconfiguration

Detects insecure TLS/SSL configurations.

```bash
./utils/scanners/crypto-tls/crypto-tls.sh /path/to/repo
```

**Detected Issues:**
- Disabled certificate verification (`verify=False`)
- Deprecated protocols (SSLv3, TLS 1.0, TLS 1.1)
- Disabled hostname verification
- Trust-all certificate managers
- `CERT_NONE` usage
- `rejectUnauthorized: false`

**Output:** `crypto-tls.json`

---

### digital-certificates

**Purpose:** Certificate analysis

Analyzes X.509 certificates in the codebase.

```bash
./utils/scanners/digital-certificates/digital-certificates.sh /path/to/repo
```

**Output:** `digital-certificates.json`

---

## Package Scanners

Scanners for dependency analysis and supply chain security.

### package-sbom

**Purpose:** Software Bill of Materials generation

Generates CycloneDX SBOM from dependency manifests.

```bash
./utils/scanners/package-sbom/package-sbom.sh /path/to/repo
```

**Options:**
| Flag | Description |
|------|-------------|
| `--generator syft\|cdxgen` | SBOM generator to use |
| `--format cyclonedx\|spdx` | Output format |

**Supported Manifests:**
- `package.json`, `package-lock.json`, `yarn.lock`
- `requirements.txt`, `Pipfile.lock`, `poetry.lock`
- `go.mod`, `go.sum`
- `Cargo.toml`, `Cargo.lock`
- `pom.xml`, `build.gradle`
- `Gemfile.lock`

**Output:** `package-sbom.json`, `sbom.cdx.json`

---

### package-vulns

**Purpose:** Vulnerability detection

Scans dependencies for known CVEs using OSV database.

```bash
./utils/scanners/package-vulns/package-vulns.sh /path/to/repo
```

**Options:**
| Flag | Description |
|------|-------------|
| `--sbom FILE` | Use existing SBOM file |

**Data Sources:**
- OSV (Open Source Vulnerabilities)
- GitHub Security Advisories
- NVD (National Vulnerability Database)

**Output:** `package-vulns.json`

---

### package-health

**Purpose:** Dependency health assessment

Analyzes package maintenance, abandonment risk, and community health.

```bash
./utils/scanners/package-health/package-health.sh /path/to/repo
```

**Health Signals:**
- Last publish date
- Download trends
- Maintainer activity
- Open issues/PRs ratio
- Typosquatting risk

**Output:** `package-health.json`

---

### package-malcontent

**Purpose:** Supply chain malware detection

Behavioral analysis for malicious code patterns using malcontent.

```bash
./utils/scanners/package-malcontent/package-malcontent.sh /path/to/repo
```

**Detected Behaviors:**
- Data exfiltration
- Code execution
- Persistence mechanisms
- Network connections
- File system operations
- Post-install scripts

**Output:** `package-malcontent.json`, `package-malcontent/` directory

---

### package-provenance

**Purpose:** Build provenance verification

Verifies SLSA provenance and supply chain integrity.

```bash
./utils/scanners/package-provenance/package-provenance.sh /path/to/repo
```

**Checks:**
- Sigstore signatures
- SLSA provenance attestations
- Build reproducibility

**Output:** `package-provenance.json`

---

### licenses

**Purpose:** License compliance

Analyzes licenses across the dependency tree.

```bash
./utils/scanners/licenses/licenses.sh /path/to/repo
```

**Checks:**
- License identification (SPDX)
- License compatibility
- Copyleft detection
- Unknown/missing licenses

**Output:** `licenses.json`

---

## Infrastructure Scanners

Scanners for infrastructure as code and containers.

### iac-security

**Purpose:** Infrastructure as Code security

Scans Terraform, CloudFormation, Kubernetes manifests using Checkov.

```bash
./utils/scanners/iac-security/iac-security.sh /path/to/repo
```

**Supported Formats:**
- Terraform (`.tf`, `.tfvars`)
- CloudFormation (YAML/JSON)
- Kubernetes manifests
- Helm charts
- Docker Compose
- ARM templates

**Output:** `iac-security.json`

---

### container-security

**Purpose:** Container image security

Analyzes Dockerfiles and container images using Trivy and Hadolint.

```bash
./utils/scanners/container-security/container-security.sh /path/to/repo
```

**Checks:**
- Dockerfile best practices
- Base image vulnerabilities
- Layer security
- Secret exposure in layers

**Output:** `container-security.json`

---

### containers

**Purpose:** Container enumeration

Identifies container technologies and configurations.

```bash
./utils/scanners/containers/containers.sh /path/to/repo
```

**Output:** `containers.json`

---

## Analysis Scanners

Scanners for code quality and project analysis.

### code-ownership

**Purpose:** Code ownership analysis

Analyzes git history for code ownership and bus factor.

```bash
./utils/scanners/code-ownership/code-ownership.sh /path/to/repo
./utils/scanners/code-ownership/bus-factor.sh /path/to/repo
```

**Output:** `code-ownership.json`

---

### dora

**Purpose:** DORA metrics calculation

Calculates DevOps Research and Assessment metrics.

```bash
./utils/scanners/dora/dora.sh /path/to/repo
```

**Metrics:**
- Deployment frequency
- Lead time for changes
- Change failure rate
- Time to restore service

**Output:** `dora.json`

---

### test-coverage

**Purpose:** Test coverage analysis

Analyzes test coverage reports.

```bash
./utils/scanners/test-coverage/test-coverage.sh /path/to/repo
```

**Output:** `test-coverage.json`

---

### documentation

**Purpose:** Documentation analysis

Analyzes project documentation quality.

```bash
./utils/scanners/documentation/documentation.sh /path/to/repo
```

**Output:** `documentation.json`

---

### bundle-analysis

**Purpose:** Frontend bundle analysis

Analyzes JavaScript/TypeScript bundle size and composition.

```bash
./utils/scanners/bundle-analysis/bundle-analysis.sh /path/to/repo
```

**Output:** `bundle-analysis.json`

---

### git

**Purpose:** Git repository analysis

Analyzes git history, contributors, and patterns.

```bash
./utils/scanners/git/git.sh /path/to/repo
```

**Output:** `git.json`

---

## Scan Profiles

Scanners are grouped into profiles for different use cases:

### quick

Fast scan for initial assessment (~2 minutes)

```bash
./zero.sh hydrate owner/repo --quick
```

**Scanners:** `tech-discovery`, `package-sbom`, `package-vulns`, `licenses`

---

### standard

Balanced security analysis (~5 minutes)

```bash
./zero.sh hydrate owner/repo
```

**Scanners:** `tech-discovery`, `package-sbom`, `package-vulns`, `package-health`, `licenses`, `code-secrets`, `code-vulns`

---

### security

Comprehensive security assessment (~10 minutes)

```bash
./zero.sh hydrate owner/repo --security
```

**Scanners:** `tech-discovery`, `package-sbom`, `package-vulns`, `package-health`, `package-malcontent`, `package-provenance`, `licenses`, `code-secrets`, `code-vulns`, `iac-security`, `container-security`

---

### deep

Full analysis including all scanners (~15 minutes)

```bash
./zero.sh hydrate owner/repo --deep
```

**Scanners:** All available scanners

---

### crypto

Comprehensive cryptographic analysis (~5 minutes)

```bash
./zero.sh hydrate owner/repo --crypto
```

**Scanners:** `tech-discovery`, `crypto-ciphers`, `crypto-keys`, `crypto-random`, `crypto-tls`, `code-secrets`, `code-vulns`, `digital-certificates`

---

### packages

Comprehensive dependency analysis (~8 minutes)

```bash
./zero.sh hydrate owner/repo --packages
```

**Scanners:** `package-sbom`, `package-vulns`, `package-health`, `package-malcontent`, `package-provenance`, `licenses`

---

## See Also

- [Scanner Architecture](../architecture/scanners.md) - How scanners work
- [RAG Pipeline](../architecture/rag-pipeline.md) - Pattern generation
- [Output Formats](output-formats.md) - JSON output schemas
