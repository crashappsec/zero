# Scanner Naming Taxonomy

## Principles

1. **Folder name = Scanner ID** (exact match, no exceptions)
2. **Main script = `{scanner-id}.sh`** (no `-data`, `-analyser`, `-scanner` suffixes)
3. **American spelling** throughout (`analyzer` not `analyser`)
4. **All scanners under `utils/scanners/`** (single location)
5. **Consistent `lib/` subfolder** for scanner-specific shared code
6. **Shared libraries in `utils/lib/`** for cross-scanner utilities

---

## Complete Migration Map

### Primary Scanners (Used in Profiles)

| Scanner ID       | OLD Location                                    | NEW Location                          | OLD Script                           | NEW Script           |
|------------------|------------------------------------------------|---------------------------------------|--------------------------------------|----------------------|
| `auth`           | `utils/auth-analysis/`                         | `utils/scanners/auth/`                | `auth-analysis-data.sh`              | `auth.sh`            |
| `documentation`  | `utils/documentation/`                         | `utils/scanners/documentation/`       | `documentation-data.sh`              | `documentation.sh`   |
| `dora`           | `utils/dora-metrics/`                          | `utils/scanners/dora/`                | `dora-analyser-data.sh`              | `dora.sh`            |
| `git`            | `utils/git-insights/`                          | `utils/scanners/git/`                 | `git-insights-data.sh`               | `git.sh`             |
| `iac-security`   | `utils/iac-security/`                          | `utils/scanners/iac-security/`        | `iac-security-data.sh`               | `iac-security.sh`    |
| `licenses`       | `utils/legal-review/`                          | `utils/scanners/licenses/`            | `legal-analyser-data.sh`             | `licenses.sh`        |
| `ownership`      | `utils/code-ownership/`                        | `utils/scanners/ownership/`           | `ownership-analyser-data.sh`         | `ownership.sh`       |
| `package-health` | `utils/supply-chain/package-health-analysis/`  | `utils/scanners/package-health/`      | `package-health-analyser.sh`         | `package-health.sh`  |
| `package-vulns`  | `utils/supply-chain/vulnerability-analysis/`   | `utils/scanners/package-vulns/`       | `vulnerability-analyser-data.sh`     | `package-vulns.sh`   |
| `packages`       | `utils/supply-chain/`                          | `utils/scanners/packages/`            | `supply-chain-scanner.sh`            | `packages.sh`        |
| `package-provenance` | `utils/supply-chain/provenance-analysis/`  | `utils/scanners/package-provenance/`  | `provenance-analyser.sh`             | `package-provenance.sh` |
| `secrets`        | `utils/secrets-scanner/`                       | `utils/scanners/secrets/`             | `secrets-scanner-data.sh`            | `secrets.sh`         |
| `security`       | `utils/code-security/`                         | `utils/scanners/security/`            | `code-security-data.sh`              | `security.sh`        |
| `tech-debt`      | `utils/tech-debt/`                             | `utils/scanners/tech-debt/`           | `tech-debt-data.sh`                  | `tech-debt.sh`       |
| `technology`     | `utils/technology-identification/`             | `utils/scanners/technology/`          | `technology-identification-data.sh`  | `technology.sh`      |
| `tests`          | `utils/test-coverage/`                         | `utils/scanners/tests/`               | `test-coverage-data.sh`              | `tests.sh`           |

### Secondary Scanners (Not in Standard Profiles)

| Scanner ID       | OLD Location                                    | NEW Location                          | OLD Script                           | NEW Script           |
|------------------|------------------------------------------------|---------------------------------------|--------------------------------------|----------------------|
| `bundle`         | `utils/supply-chain/bundle-analysis/`          | `utils/scanners/bundle/`              | `lib/bundle-analyzer.sh`             | `bundle.sh`          |
| `certificates`   | `utils/certificate-analyser/`                  | `utils/scanners/certificates/`        | `cert-analyser.sh`                   | `certificates.sh`    |
| `chalk`          | `utils/chalk-build-analyser/`                  | `utils/scanners/chalk/`               | `chalk-build-analyser.sh`            | `chalk.sh`           |
| `containers`     | `utils/supply-chain/container-analysis/`       | `utils/scanners/containers/`          | `lib/image-recommender.sh`           | `containers.sh`      |
| `package-recommendations` | `utils/supply-chain/library-recommendations/` | `utils/scanners/package-recommendations/` | `lib/recommender.sh`            | `package-recommendations.sh` |

### Utility Folders (Not Scanners)

| Purpose          | OLD Location                 | NEW Location                    | Notes                              |
|------------------|------------------------------|--------------------------------|-------------------------------------|
| Cost estimation  | `utils/cocomo/`              | `utils/tools/cocomo/`          | COCOMO cost estimation              |
| Git utilities    | `utils/git-sync/`            | `utils/tools/git-sync/`        | Git synchronization utilities       |
| Validation       | `utils/validation/`          | `utils/tools/validation/`      | Commit/copyright validation         |
| Phantom CLI      | `utils/phantom/`             | `utils/phantom/`               | Keep as-is (main CLI)               |
| Shared libs      | `utils/lib/`                 | `utils/lib/`                   | Keep as-is (shared libraries)       |

---

## New Directory Structure

```
utils/
├── scanners/                           # All scanners here
│   ├── auth/
│   │   ├── auth.sh                     # Main entry point
│   │   └── lib/                        # Scanner-specific libs
│   │
│   ├── bundle/
│   │   ├── bundle.sh
│   │   └── lib/
│   │       └── debt-scorer.sh
│   │
│   ├── certificates/
│   │   ├── certificates.sh
│   │   └── lib/
│   │       ├── cab-compliance.sh
│   │       ├── cert-compare.sh
│   │       ├── chain-validation.sh
│   │       ├── claude-analysis.sh
│   │       ├── fingerprint.sh
│   │       ├── format-detection.sh
│   │       ├── format-parsers.sh
│   │       ├── ocsp-verification.sh
│   │       └── starttls.sh
│   │
│   ├── chalk/
│   │   └── chalk.sh
│   │
│   ├── containers/
│   │   ├── containers.sh
│   │   └── lib/
│   │
│   ├── documentation/
│   │   └── documentation.sh
│   │
│   ├── dora/
│   │   ├── dora.sh
│   │   └── lib/
│   │
│   ├── git/
│   │   └── git.sh
│   │
│   ├── iac-security/
│   │   └── iac-security.sh
│   │
│   ├── licenses/
│   │   ├── licenses.sh
│   │   └── lib/
│   │
│   ├── ownership/
│   │   ├── ownership.sh
│   │   └── lib/
│   │       ├── scanner-core.sh
│   │       ├── codeowners-generator.sh
│   │       ├── config.sh
│   │       ├── csv.sh
│   │       ├── markdown.sh
│   │       ├── metrics.sh
│   │       ├── succession.sh
│   │       └── trends.sh
│   │
│   ├── package-health/
│   │   ├── package-health.sh
│   │   └── lib/
│   │       ├── abandonment-detector.sh
│   │       ├── deprecation-checker.sh
│   │       ├── health-scoring.sh
│   │       ├── typosquat-detector.sh
│   │       ├── unused-detector.sh
│   │       └── version-analysis.sh
│   │
│   ├── package-vulns/
│   │   ├── package-vulns.sh
│   │   └── lib/
│   │       └── osv-client.sh
│   │
│   ├── packages/
│   │   ├── packages.sh
│   │   └── lib/
│   │       ├── deps-dev-client.sh
│   │       ├── popular-packages.sh
│   │       └── version-normalizer.sh
│   │
│   ├── package-provenance/
│   │   └── package-provenance.sh
│   │
│   ├── package-recommendations/
│   │   ├── package-recommendations.sh
│   │   └── lib/
│   │
│   ├── secrets/
│   │   └── secrets.sh
│   │
│   ├── security/
│   │   ├── security.sh
│   │   └── lib/
│   │       ├── context-builder.sh
│   │       ├── file-scanner.sh
│   │       ├── report-generator.sh
│   │       └── severity-classifier.sh
│   │
│   ├── tech-debt/
│   │   └── tech-debt.sh
│   │
│   ├── technology/
│   │   ├── technology.sh
│   │   └── lib/
│   │       └── pattern-loader.sh
│   │
│   └── tests/
│       └── tests.sh
│
├── lib/                                # Shared libraries
│   ├── claude-cost.sh
│   ├── codeowners-validator.sh
│   ├── config-loader.sh
│   ├── config.sh
│   ├── github.sh
│   ├── org-scanner.sh
│   └── sbom.sh
│
├── phantom/                            # Phantom CLI (unchanged)
│   ├── phantom.sh
│   ├── bootstrap.sh
│   ├── hydrate.sh
│   ├── preflight.sh
│   ├── phantom.config.json
│   └── lib/
│       └── gibson.sh
│
└── tools/                              # Non-scanner utilities
    ├── cocomo/
    ├── git-sync/
    └── validation/
        ├── check-commit-message.sh
        └── check-copyright.sh
```

---

## Updated phantom.config.json Scanners Section

```json
{
  "scanners": {
    "auth": {
      "name": "Auth Analysis",
      "description": "Identify authentication patterns and security configurations",
      "script": "utils/scanners/auth/auth.sh",
      "output_file": "auth.json"
    },
    "bundle": {
      "name": "Bundle Analysis",
      "description": "Analyze JavaScript bundle sizes and optimization opportunities",
      "script": "utils/scanners/bundle/bundle.sh",
      "output_file": "bundle.json"
    },
    "certificates": {
      "name": "Certificate Analysis",
      "description": "Analyze SSL/TLS certificates for security and compliance",
      "script": "utils/scanners/certificates/certificates.sh",
      "output_file": "certificates.json"
    },
    "chalk": {
      "name": "Chalk Build Analysis",
      "description": "Analyze Chalk build artifacts and provenance",
      "script": "utils/scanners/chalk/chalk.sh",
      "output_file": "chalk.json"
    },
    "containers": {
      "name": "Container Analysis",
      "description": "Analyze container images for security and best practices",
      "script": "utils/scanners/containers/containers.sh",
      "output_file": "containers.json"
    },
    "documentation": {
      "name": "Documentation",
      "description": "Analyze README, docs coverage, and documentation quality",
      "script": "utils/scanners/documentation/documentation.sh",
      "output_file": "documentation.json"
    },
    "dora": {
      "name": "DORA Metrics",
      "description": "Calculate deployment frequency, lead time, and other DevOps metrics",
      "script": "utils/scanners/dora/dora.sh",
      "output_file": "dora.json"
    },
    "git": {
      "name": "Git Insights",
      "description": "Analyze commit patterns, contributors, and repository activity",
      "script": "utils/scanners/git/git.sh",
      "output_file": "git.json"
    },
    "iac-security": {
      "name": "IaC Security",
      "description": "Security scanning for Terraform, CloudFormation, Kubernetes configs",
      "script": "utils/scanners/iac-security/iac-security.sh",
      "output_file": "iac-security.json"
    },
    "licenses": {
      "name": "License Compliance",
      "description": "Analyze dependency licenses for compliance risks",
      "script": "utils/scanners/licenses/licenses.sh",
      "output_file": "licenses.json"
    },
    "ownership": {
      "name": "Code Ownership",
      "description": "Analyze CODEOWNERS, contributor patterns, and maintainership",
      "script": "utils/scanners/ownership/ownership.sh",
      "output_file": "ownership.json"
    },
    "package-health": {
      "name": "Package Health",
      "description": "Check for abandoned, deprecated, or unhealthy dependencies",
      "script": "utils/scanners/package-health/package-health.sh",
      "output_file": "package-health.json"
    },
    "package-vulns": {
      "name": "Package Vulnerabilities",
      "description": "Scan packages for known CVEs using OSV database",
      "script": "utils/scanners/package-vulns/package-vulns.sh",
      "output_file": "package-vulns.json"
    },
    "packages": {
      "name": "Packages (SBOM)",
      "description": "Extract package dependencies from manifests (package.json, requirements.txt, etc.)",
      "script": "utils/scanners/packages/packages.sh",
      "output_file": "packages.json"
    },
    "package-provenance": {
      "name": "Package Provenance",
      "description": "Verify build provenance and supply chain attestations (SLSA)",
      "script": "utils/scanners/package-provenance/package-provenance.sh",
      "output_file": "package-provenance.json"
    },
    "package-recommendations": {
      "name": "Package Recommendations",
      "description": "Suggest alternative libraries based on health and security",
      "script": "utils/scanners/package-recommendations/package-recommendations.sh",
      "output_file": "package-recommendations.json"
    },
    "secrets": {
      "name": "Secrets Scanner",
      "description": "Detect exposed API keys, passwords, and credentials",
      "script": "utils/scanners/secrets/secrets.sh",
      "output_file": "secrets.json"
    },
    "security": {
      "name": "Code Security",
      "description": "Static analysis for security vulnerabilities in source code",
      "script": "utils/scanners/security/security.sh",
      "output_file": "security.json"
    },
    "tech-debt": {
      "name": "Technical Debt",
      "description": "Analyze code duplication, complexity, and TODO markers",
      "script": "utils/scanners/tech-debt/tech-debt.sh",
      "output_file": "tech-debt.json"
    },
    "technology": {
      "name": "Technology Stack",
      "description": "Identify frameworks, languages, and tools used in the codebase",
      "script": "utils/scanners/technology/technology.sh",
      "output_file": "technology.json"
    },
    "tests": {
      "name": "Test Coverage",
      "description": "Detect test frameworks and estimate test coverage",
      "script": "utils/scanners/tests/tests.sh",
      "output_file": "tests.json"
    }
  }
}
```

---

## Updated Profiles

```json
{
  "profiles": {
    "quick": {
      "name": "Quick",
      "description": "Fast essential analysis for rapid feedback",
      "estimated_time": "~30 seconds",
      "scanners": ["packages", "technology", "package-vulns", "licenses", "tech-debt"]
    },
    "standard": {
      "name": "Standard",
      "description": "Balanced analysis covering security and code quality",
      "estimated_time": "~2 minutes",
      "scanners": ["packages", "technology", "package-vulns", "licenses", "security", "secrets", "tech-debt", "ownership", "dora"]
    },
    "advanced": {
      "name": "Advanced",
      "description": "Comprehensive static analysis with health checks",
      "estimated_time": "~5 minutes",
      "scanners": ["packages", "technology", "package-vulns", "package-health", "licenses", "security", "iac-security", "secrets", "tech-debt", "documentation", "git", "tests", "ownership", "dora", "package-provenance"]
    },
    "deep": {
      "name": "Deep",
      "description": "Full analysis with Claude AI-assisted insights",
      "estimated_time": "~10 minutes",
      "requires_claude": true,
      "scanners": ["packages", "technology", "package-vulns", "package-health", "licenses", "security", "iac-security", "secrets", "tech-debt", "documentation", "git", "tests", "auth", "ownership", "dora", "package-provenance"]
    },
    "security": {
      "name": "Security",
      "description": "Security-focused analysis for vulnerability assessment",
      "estimated_time": "~3 minutes",
      "scanners": ["packages", "package-vulns", "licenses", "security", "iac-security", "secrets", "auth"]
    },
    "compliance": {
      "name": "Compliance",
      "description": "License and policy compliance checks",
      "estimated_time": "~2 minutes",
      "scanners": ["packages", "licenses", "security", "documentation", "ownership"]
    },
    "devops": {
      "name": "DevOps",
      "description": "CI/CD and operational metrics",
      "estimated_time": "~3 minutes",
      "scanners": ["packages", "technology", "iac-security", "git", "tests", "dora", "package-provenance"]
    }
  }
}
```

---

## Summary of Name Changes

| OLD Name              | NEW Name         | Reason                                      |
|-----------------------|------------------|---------------------------------------------|
| `dependencies`        | `packages`       | More specific - it's about package manifests |
| `vulnerabilities`     | `package-vulns`  | Clarifies it's package CVEs via OSV         |
| `iac-security`        | `iac-security`   | Keep full name for clarity                  |
| `security-findings`   | `security`       | Simpler, the output file clarifies it       |
| `git-insights`        | `git`            | Shorter, clearer                            |
| `test-coverage`       | `tests`          | Shorter, clearer                            |

---

## Migration Checklist

1. [ ] Create `utils/scanners/` directory
2. [ ] Create `utils/tools/` directory
3. [ ] Move and rename each scanner folder
4. [ ] Rename main scripts to `{scanner-id}.sh`
5. [ ] Rename `*-analyser.sh` to `*-analyzer.sh` (American spelling)
6. [ ] Update all internal `source` paths
7. [ ] Update `phantom.config.json` with new paths and "scanners" key
8. [ ] Update `hydrate.sh` scanner dispatcher
9. [ ] Update any RAG folder references
10. [ ] Test all profiles work correctly
