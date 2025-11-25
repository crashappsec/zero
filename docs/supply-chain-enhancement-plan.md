# Supply Chain Scanner Enhancement Plan

## Executive Summary

This plan outlines enhancements to the Gibson-Powers supply chain scanner to address:
1. Version normalization for improved vulnerability matching
2. Deprecated and abandoned package detection
3. Typosquatting and malicious package detection
4. Unused dependency identification
5. Hardened container image recommendations

## Current State Analysis

### Existing Capabilities

The supply chain scanner already includes:
- **Vulnerability Analysis** (`vulnerability-analyser.sh`)
  - OSV.dev integration
  - CISA KEV checking
  - Taint analysis for vulnerable dependency chains

- **Code Security Analysis** (`code-security-analyser.sh`)
  - Secret detection (exposed credentials, API keys)
  - Security pattern analysis

- **Provenance Analysis** (`provenance-analyser.sh`)
  - SLSA verification
  - npm provenance checking
  - Sigstore signature validation

- **Package Health Analysis** (`package-health-analyser.sh`)
  - deps.dev integration
  - OpenSSF Scorecard checks
  - Basic deprecation detection
  - Health scoring system

- **Legal Compliance** (separate module)
  - License analysis

### Gaps Identified

| Gap | Category | Impact | Priority |
|-----|----------|--------|----------|
| Typosquatting detection | Security | Supply chain attacks | High |
| Abandoned package detection | Security | Unpatched vulnerabilities | High |
| Unused dependency detection | Productivity | Larger attack surface, slower builds | Medium |
| Version normalization across ecosystems | Reliability | Inconsistent builds, missed vuln matches | High |
| Container image recommendations | Reliability | OS-level vulnerabilities, inconsistent environments | Medium |
| Behavioral analysis | Security | Malicious packages | Low (complex) |

---

## Implementation Plan

### Phase 1: Version Normalization (Reliability)

#### Objective
Ensure consistent version comparison across ecosystems for reproducible builds and reliable dependency resolution. This foundation supports both vulnerability matching accuracy and standardized engineering environments when combined with hardened container images.

#### New Components

**File: `utils/supply-chain/lib/version-normalizer.sh`**

```bash
#!/bin/bash
# Version normalization library

# Normalize version string based on ecosystem
normalize_version() {
    local version="$1"
    local ecosystem="$2"

    case "$ecosystem" in
        npm|node)
            normalize_npm_version "$version"
            ;;
        pypi|python)
            normalize_pypi_version "$version"
            ;;
        maven|java)
            normalize_maven_version "$version"
            ;;
        nuget|dotnet)
            normalize_nuget_version "$version"
            ;;
        go)
            normalize_go_version "$version"
            ;;
        *)
            echo "$version"
            ;;
    esac
}

# npm: Remove 'v' prefix, pad to semver
normalize_npm_version() {
    echo "$1" | sed 's/^v//' | awk -F. '{
        printf "%d.%d.%d", $1+0, ($2=="" ? 0 : $2+0), ($3=="" ? 0 : $3+0)
    }'
}

# PyPI: PEP 440 normalization
normalize_pypi_version() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | \
        sed 's/[_-]/./g' | \
        sed 's/\.alpha/a/g; s/\.beta/b/g; s/\.rc/rc/g; s/\.preview/rc/g' | \
        sed 's/^0*//'
}
```

#### Integration Points
- Modify `vulnerability-analyser.sh` to normalize versions before OSV lookup
- Update SBOM parsing to store both original and normalized versions
- Add `--normalize-versions` flag to supply-chain-scanner.sh

#### RAG Updates
- [x] Created `rag/supply-chain/version-normalization/version-normalization-guide.md`

---

### Phase 2: Abandoned Package Detection (Week 2)

#### Objective
Identify packages that are no longer maintained and may contain unpatched vulnerabilities.

#### New Components

**File: `utils/supply-chain/package-health-analysis/lib/abandonment-detector.sh`**

```bash
#!/bin/bash
# Abandoned package detection

ABANDONED_THRESHOLD_DAYS=730  # 2 years
WARNING_THRESHOLD_DAYS=365    # 1 year

check_abandonment_status() {
    local pkg="$1"
    local ecosystem="$2"

    local status="healthy"
    local risk_factors=()

    # Check last commit via deps.dev
    local pkg_info=$(get_depsdev_info "$pkg" "$ecosystem")
    local last_update=$(echo "$pkg_info" | jq -r '.lastUpdate // empty')

    # Check if explicitly deprecated
    if check_deprecated "$pkg" "$ecosystem"; then
        status="deprecated"
        risk_factors+=("explicitly_deprecated")
    fi

    # Check if archived
    if check_archived "$pkg" "$ecosystem"; then
        status="archived"
        risk_factors+=("repository_archived")
    fi

    # Check last update age
    local days_since=$(days_since_date "$last_update")
    if [[ $days_since -gt $ABANDONED_THRESHOLD_DAYS ]]; then
        status="abandoned"
        risk_factors+=("no_updates_${days_since}_days")
    elif [[ $days_since -gt $WARNING_THRESHOLD_DAYS ]]; then
        status="stale"
        risk_factors+=("last_update_${days_since}_days_ago")
    fi

    echo "{\"status\": \"$status\", \"risk_factors\": $(printf '%s\n' "${risk_factors[@]}" | jq -R . | jq -s .)}"
}
```

#### Integration Points
- Add `--check-abandonment` flag to package-health-analyser.sh
- Include abandonment status in health scoring
- Generate alert for deprecated/abandoned packages

#### RAG Updates
- [x] Created `rag/supply-chain/malicious-package-detection/abandoned-package-detection.md`

---

### Phase 3: Typosquatting Detection (Week 3)

#### Objective
Detect potentially malicious packages with names similar to popular packages.

#### New Components

**File: `utils/supply-chain/package-health-analysis/lib/typosquat-detector.sh`**

```bash
#!/bin/bash
# Typosquatting detection

# Popular packages by ecosystem (top 100)
declare -A POPULAR_NPM=(
    [lodash]=1 [express]=1 [react]=1 [axios]=1 [moment]=1
    [request]=1 [chalk]=1 [debug]=1 [commander]=1 [async]=1
    # ... expand to top 100
)

# Levenshtein distance calculation
levenshtein() {
    python3 -c "
def lev(s1, s2):
    if len(s1) < len(s2): return lev(s2, s1)
    if len(s2) == 0: return len(s1)
    prev = range(len(s2) + 1)
    for i, c1 in enumerate(s1):
        curr = [i + 1]
        for j, c2 in enumerate(s2):
            curr.append(min(prev[j+1]+1, curr[j]+1, prev[j]+(c1!=c2)))
        prev = curr
    return prev[-1]
print(lev('$1', '$2'))
"
}

check_typosquat() {
    local pkg="$1"
    local ecosystem="$2"

    local popular_list
    case "$ecosystem" in
        npm) popular_list=("${!POPULAR_NPM[@]}") ;;
        pypi) popular_list=("${!POPULAR_PYPI[@]}") ;;
    esac

    for popular in "${popular_list[@]}"; do
        local distance=$(levenshtein "$pkg" "$popular")
        local threshold=$((${#popular} / 3))

        if [[ $distance -gt 0 && $distance -le $threshold ]]; then
            echo "{\"suspicious\": true, \"similar_to\": \"$popular\", \"distance\": $distance}"
            return 0
        fi
    done

    echo "{\"suspicious\": false}"
}
```

#### Integration Points
- Add typosquat check to dependency scanning
- Flag suspicious packages in reports
- Optional: Integrate with Socket.dev or Trusty APIs

#### RAG Updates
- [x] Created `rag/supply-chain/malicious-package-detection/typosquatting-detection.md`

---

### Phase 4: Unused Dependency Detection (Week 4)

#### Objective
Identify dependencies declared but not actually used in the codebase.

#### New Components

**File: `utils/supply-chain/package-health-analysis/lib/unused-detector.sh`**

```bash
#!/bin/bash
# Unused dependency detection

check_unused_npm() {
    local project_dir="$1"

    if ! command -v npx &> /dev/null; then
        echo '{"error": "npx not available"}'
        return 1
    fi

    cd "$project_dir"
    npx depcheck --json 2>/dev/null | jq '{
        unused_dependencies: .dependencies,
        unused_devDependencies: .devDependencies,
        missing_dependencies: .missing
    }'
}

check_unused_python() {
    local project_dir="$1"

    cd "$project_dir"

    # Generate requirements from actual imports
    if command -v pipreqs &> /dev/null; then
        pipreqs . --print --force 2>/dev/null > /tmp/actual_reqs.txt

        # Compare with declared requirements
        if [[ -f requirements.txt ]]; then
            local declared=$(cut -d'=' -f1 requirements.txt | tr '[:upper:]' '[:lower:]' | sort)
            local actual=$(cut -d'=' -f1 /tmp/actual_reqs.txt | tr '[:upper:]' '[:lower:]' | sort)

            local unused=$(comm -23 <(echo "$declared") <(echo "$actual"))
            echo "{\"unused\": $(echo "$unused" | jq -R . | jq -s .)}"
        fi
    fi
}
```

#### Integration Points
- New `--check-unused` flag for package-health-analyser.sh
- Optional tool availability detection (depcheck, pipreqs)
- Include unused count in security scoring

#### RAG Updates
- [x] Created `rag/supply-chain/unused-dependency-detection/unused-dependencies-guide.md`

---

### Phase 5: Container Image Recommendations (Reliability)

#### Objective
Recommend standardized, hardened base images (Chainguard, Minimus, Google Distroless) to create consistent engineering environments. Combined with Phase 1's version normalization, this ensures reproducible builds with both consistent dependencies and runtime environments across development, staging, and production.

#### New Components

**File: `utils/supply-chain/container-analysis/image-recommender.sh`**

```bash
#!/bin/bash
# Container image hardening recommendations

analyze_dockerfile() {
    local dockerfile="$1"
    local recommendations=()

    # Extract base image
    local base_image=$(grep -E "^FROM" "$dockerfile" | head -1 | awk '{print $2}')

    # Check for full OS images
    case "$base_image" in
        ubuntu:*|debian:*|centos:*|fedora:*)
            recommendations+=("Consider distroless or Alpine base image")
            ;;
        *:latest)
            recommendations+=("Pin image version instead of using :latest")
            ;;
    esac

    # Recommend alternatives based on language
    if grep -q "npm\|node" "$dockerfile"; then
        recommendations+=("Consider: gcr.io/distroless/nodejs or cgr.dev/chainguard/node")
    elif grep -q "pip\|python" "$dockerfile"; then
        recommendations+=("Consider: gcr.io/distroless/python3 or cgr.dev/chainguard/python")
    elif grep -q "go build" "$dockerfile"; then
        recommendations+=("Consider: gcr.io/distroless/static or scratch for static Go binaries")
    fi

    # Check for multi-stage builds
    if ! grep -qE "FROM.*AS" "$dockerfile"; then
        recommendations+=("Use multi-stage builds to reduce final image size")
    fi

    printf '%s\n' "${recommendations[@]}" | jq -R . | jq -s '{"recommendations": .}'
}
```

#### Integration Points
- New `--container-analysis` module for supply-chain-scanner.sh
- Dockerfile detection and analysis
- Image signature and SBOM verification

#### RAG Updates
- [x] Created `rag/supply-chain/hardened-images/gold-images-guide.md`

---

## Module Integration Matrix

| Module | Version Norm | Abandonment | Typosquat | Unused | Container |
|--------|-------------|-------------|-----------|--------|-----------|
| vulnerability-analyser | Required | - | - | - | - |
| package-health-analyser | Required | Required | Required | Optional | - |
| provenance-analyser | Optional | - | - | - | - |
| container-analyser (NEW) | - | - | - | - | Required |

---

## Updated Claude Prompts

### Supply Chain Analysis Prompt Enhancement

```markdown
# Enhanced Supply Chain Security Analysis

Analyze the following supply chain scan results with focus on:

## 1. Version Consistency
- Flag packages with non-normalized versions
- Identify version inconsistencies across lockfiles
- Check for version range issues

## 2. Package Health Assessment
- Abandonment risk (no updates > 2 years)
- Deprecation notices
- Maintainer activity
- OpenSSF Scorecard findings

## 3. Malicious Package Indicators
- Typosquatting risk (similar to popular packages)
- Newly published packages with suspicious patterns
- Install script behavior

## 4. Optimization Opportunities
- Unused dependencies
- Duplicate packages (different versions)
- Oversized packages with lighter alternatives

## 5. Container Security (if applicable)
- Base image recommendations
- Multi-stage build opportunities
- Signature verification status

Provide actionable recommendations prioritized by risk level.
```

---

## New CLI Flags

```bash
./supply-chain-scanner.sh \
    --normalize-versions      # Enable version normalization
    --check-abandonment       # Check for abandoned packages
    --check-typosquat         # Enable typosquatting detection
    --check-unused            # Detect unused dependencies
    --container-analysis      # Analyze Dockerfiles
    --all-checks              # Enable all enhanced checks
```

---

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Vulnerability match rate | +15% improvement | Compare before/after normalization |
| Abandoned package detection | 100% of packages > 2 years | Cross-reference with deps.dev |
| Typosquat detection rate | > 90% of known attacks | Test against known malicious packages |
| False positive rate | < 5% | Manual review of flagged packages |
| Scan time impact | < 20% increase | Benchmark full scans |

---

## Testing Strategy

### Unit Tests
- Version normalization across all ecosystems
- Levenshtein distance calculations
- Abandonment threshold logic

### Integration Tests
- Full scan of OWASP Juice Shop (known vulnerabilities)
- Scan of projects with deprecated dependencies
- Container image analysis on real Dockerfiles

### Regression Tests
- Ensure existing functionality unchanged
- Compare reports before/after enhancements

---

## Rollout Plan

1. **Week 1**: Version normalization + unit tests
2. **Week 2**: Abandonment detection + integration
3. **Week 3**: Typosquatting detection + popular package lists
4. **Week 4**: Unused dependency detection + tool integration
5. **Week 5**: Container analysis + recommendations
6. **Week 6**: Full integration testing + documentation

---

## Dependencies

### External Tools (Optional)
- `depcheck` - npm unused dependency detection
- `pipreqs` - Python import-based requirements
- `vulture` - Python dead code detection
- `cosign` - Container image signature verification
- `grype` / `trivy` - Container vulnerability scanning

### APIs
- deps.dev (already integrated)
- OpenSSF Scorecard (already integrated)
- OSV.dev (already integrated)
- npm registry
- PyPI JSON API

---

## RAG Knowledge Base Summary

New documents created:
1. `rag/supply-chain/version-normalization/version-normalization-guide.md`
2. `rag/supply-chain/malicious-package-detection/typosquatting-detection.md`
3. `rag/supply-chain/malicious-package-detection/abandoned-package-detection.md`
4. `rag/supply-chain/unused-dependency-detection/unused-dependencies-guide.md`
5. `rag/supply-chain/hardened-images/gold-images-guide.md`
