# Supply Chain Scanner v2.0 Implementation Plan

## Executive Summary

This document details the implementation plan for transforming the Gibson Powers supply chain scanner from a security-focused tool into a **comprehensive dependency intelligence platform** that addresses security, developer productivity, cost optimization, sustainability, and compliance.

### Core Capabilities



## Architecture Overview

```
utils/supply-chain/
├── supply-chain-scanner.sh          # Main orchestrator (existing)
├── vulnerability-analysis/           # Existing module
├── provenance-analysis/              # Existing module
├── package-health-analysis/          # Enhanced
│   ├── package-health-analyser.sh
│   └── lib/
│       ├── deps-dev-client.sh        # Existing
│       ├── health-scoring.sh         # Existing
│       ├── deprecation-checker.sh    # Existing
│       ├── abandonment-detector.sh   # NEW
│       ├── typosquat-detector.sh     # NEW
│       └── unused-detector.sh        # NEW
├── library-recommendations/          # NEW MODULE
│   ├── library-recommender.sh
│   └── lib/
│       ├── alternative-finder.sh
│       ├── comparison-engine.sh
│       └── migration-estimator.sh
├── container-analysis/               # NEW MODULE
│   ├── image-recommender.sh
│   └── lib/
│       ├── dockerfile-parser.sh
│       ├── base-image-analyzer.sh
│       └── signature-verifier.sh
└── lib/
    ├── version-normalizer.sh         # NEW shared library
    └── popular-packages.sh           # NEW shared data
```

---

## Phase 1: Foundation Libraries (Reliability)

### Objective
Create shared libraries for consistent, reproducible engineering environments. Version normalization ensures reliable dependency resolution across ecosystems, which directly supports standardized container images and reproducible builds.

### Deliverables

#### 1.1 Version Normalizer (`lib/version-normalizer.sh`)

```bash
#!/bin/bash
# Version normalization across ecosystems

# Normalize version for consistent comparison
normalize_version() {
    local version="$1"
    local ecosystem="$2"
    # Implementation per ecosystem rules
}

# Compare two versions
compare_versions() {
    local v1="$1"
    local v2="$2"
    local ecosystem="$3"
    # Returns: -1, 0, or 1
}

# Parse version range (e.g., ">=1.0.0, <2.0.0")
parse_version_range() {
    local range="$1"
    local ecosystem="$2"
    # Returns: min_version, max_version, inclusive flags
}
```

**Functions:**
| Function | Purpose | Ecosystems |
|----------|---------|------------|
| `normalize_npm_version` | Remove 'v' prefix, pad semver | npm |
| `normalize_pypi_version` | PEP 440 normalization | PyPI |
| `normalize_maven_version` | Handle qualifiers | Maven |
| `normalize_nuget_version` | Remove trailing zeros | NuGet |
| `normalize_go_version` | Handle pseudo-versions | Go |
| `compare_semver` | Compare semantic versions | All |

#### 1.2 Popular Packages Database (`lib/popular-packages.sh`)

```bash
#!/bin/bash
# Popular package lists for typosquatting detection

declare -A NPM_POPULAR=(
    [lodash]=1 [express]=1 [react]=1 [axios]=1 [moment]=1
    [request]=1 [chalk]=1 [debug]=1 [commander]=1 [async]=1
    # ... top 500 packages
)

declare -A PYPI_POPULAR=(
    [requests]=1 [boto3]=1 [numpy]=1 [pandas]=1 [flask]=1
    [django]=1 [tensorflow]=1 [pytest]=1 [setuptools]=1
    # ... top 500 packages
)

# Check if package is in popular list
is_popular_package() {
    local pkg="$1"
    local ecosystem="$2"
}

# Get similar popular packages
get_similar_packages() {
    local pkg="$1"
    local ecosystem="$2"
    local threshold="${3:-2}"  # Levenshtein distance
}
```

### Complexity: Medium
### Dependencies: None
### Estimated Files: 2

---

## Phase 2: Abandoned Package Detection

### Objective
Identify packages that are no longer actively maintained.

### Deliverables

#### 2.1 Abandonment Detector (`package-health-analysis/lib/abandonment-detector.sh`)

```bash
#!/bin/bash

ABANDONED_THRESHOLD_DAYS=730   # 2 years
STALE_THRESHOLD_DAYS=365       # 1 year

# Main detection function
check_abandonment_status() {
    local pkg="$1"
    local ecosystem="$2"

    local metrics=$(gather_maintenance_metrics "$pkg" "$ecosystem")
    local score=$(calculate_abandonment_score "$metrics")
    local status=$(classify_status "$score")

    echo "{
        \"package\": \"$pkg\",
        \"status\": \"$status\",
        \"score\": $score,
        \"metrics\": $metrics
    }"
}

# Gather metrics from multiple sources
gather_maintenance_metrics() {
    local pkg="$1"
    local ecosystem="$2"

    local depsdev_info=$(query_depsdev "$pkg" "$ecosystem")
    local github_info=$(query_github_repo "$pkg" "$ecosystem")

    # Extract: last_commit, last_release, open_issues,
    #          maintainer_count, is_archived, is_deprecated
}

# Calculate weighted score
calculate_abandonment_score() {
    local metrics="$1"
    # Returns 0-100, higher = more abandoned
}

# Classify into categories
classify_status() {
    local score="$1"
    # Returns: healthy, stale, abandoned, archived, deprecated
}
```

#### 2.2 Integration with Package Health Analyser

Add to `package-health-analyser.sh`:
```bash
# New flag
--check-abandonment    # Enable abandonment detection

# Integration point
if [[ "$CHECK_ABANDONMENT" == "true" ]]; then
    source "$LIB_DIR/abandonment-detector.sh"
    abandonment_status=$(check_abandonment_status "$pkg" "$ecosystem")
    # Merge with health report
fi
```

### Complexity: Medium
### Dependencies: Phase 1 (version normalizer)
### Estimated Files: 2

---

## Phase 3: Typosquatting Detection

### Objective
Identify packages with names suspiciously similar to popular packages.

### Deliverables

#### 3.1 Typosquat Detector (`package-health-analysis/lib/typosquat-detector.sh`)

```bash
#!/bin/bash

source "$LIB_DIR/popular-packages.sh"

# Calculate Levenshtein distance
levenshtein() {
    local s1="$1"
    local s2="$2"
    # Python implementation for reliability
    python3 -c "
def lev(s1, s2):
    # ... implementation
print(lev('$s1', '$s2'))
"
}

# Check for typosquatting
check_typosquat_risk() {
    local pkg="$1"
    local ecosystem="$2"

    # Skip if package is itself popular
    if is_popular_package "$pkg" "$ecosystem"; then
        echo '{"suspicious": false, "reason": "is_popular"}'
        return
    fi

    # Check against popular packages
    local similar=$(get_similar_packages "$pkg" "$ecosystem" 2)

    if [[ -n "$similar" ]]; then
        local distance=$(levenshtein "$pkg" "$similar")
        echo "{
            \"suspicious\": true,
            \"similar_to\": \"$similar\",
            \"distance\": $distance,
            \"risk_level\": \"$(calculate_risk_level "$distance" "$pkg" "$similar")\"
        }"
    else
        echo '{"suspicious": false}'
    fi
}

# Calculate risk level based on distance and other factors
calculate_risk_level() {
    local distance="$1"
    local pkg="$2"
    local popular="$3"

    # Lower distance + new package + low downloads = higher risk
}
```

#### 3.2 Behavioral Analysis (Optional Enhancement)

```bash
# Analyze package for suspicious behavior patterns
analyze_package_behavior() {
    local pkg_dir="$1"

    local suspicious_patterns=(
        "net.createConnection.*env"
        "child_process.*exec"
        "process.env.NPM_TOKEN"
        # ... more patterns
    )

    for pattern in "${suspicious_patterns[@]}"; do
        if grep -rE "$pattern" "$pkg_dir" 2>/dev/null; then
            echo "SUSPICIOUS: Found pattern '$pattern'"
        fi
    done
}
```

### Complexity: Medium
### Dependencies: Phase 1 (popular packages list)
### Estimated Files: 1

---

## Phase 4: Unused Dependency Detection

### Objective
Identify declared dependencies that are not actually used in the codebase.

### Deliverables

#### 4.1 Unused Detector (`package-health-analysis/lib/unused-detector.sh`)

```bash
#!/bin/bash

# Detect unused npm dependencies
detect_unused_npm() {
    local project_dir="$1"

    if ! command -v npx &> /dev/null; then
        echo '{"error": "npx not available", "tool": "depcheck"}'
        return 1
    fi

    cd "$project_dir"
    npx depcheck --json 2>/dev/null | jq '{
        unused_dependencies: .dependencies,
        unused_devDependencies: .devDependencies,
        missing_dependencies: .missing,
        tool: "depcheck"
    }'
}

# Detect unused Python dependencies
detect_unused_python() {
    local project_dir="$1"

    cd "$project_dir"

    if command -v pipreqs &> /dev/null; then
        # Generate requirements from actual imports
        pipreqs . --print --force 2>/dev/null > /tmp/actual_reqs.txt

        if [[ -f requirements.txt ]]; then
            local declared=$(cut -d'=' -f1 requirements.txt | tr '[:upper:]' '[:lower:]' | sort)
            local actual=$(cut -d'=' -f1 /tmp/actual_reqs.txt | tr '[:upper:]' '[:lower:]' | sort)

            local unused=$(comm -23 <(echo "$declared") <(echo "$actual"))
            echo "{\"unused\": $(echo "$unused" | jq -R . | jq -s .), \"tool\": \"pipreqs\"}"
        fi
    else
        echo '{"error": "pipreqs not available", "tool": "pipreqs"}'
    fi
}

# Main entry point
detect_unused_dependencies() {
    local project_dir="$1"
    local ecosystem="$2"

    case "$ecosystem" in
        npm|node)
            detect_unused_npm "$project_dir"
            ;;
        pypi|python)
            detect_unused_python "$project_dir"
            ;;
        *)
            echo '{"error": "Ecosystem not supported for unused detection"}'
            ;;
    esac
}
```

### Complexity: Low
### Dependencies: External tools (depcheck, pipreqs) - optional
### Estimated Files: 1

---

## Phase 5: Container Image Recommendations (Reliability)

### Objective
Recommend standardized, hardened base images (Chainguard, Minimus, Google Distroless) to create consistent, secure engineering environments. This complements Phase 1's version normalization by ensuring not just dependency consistency but also runtime environment standardization across development, staging, and production.

### Deliverables

#### 5.1 Container Analysis Module (`container-analysis/`)

```bash
# image-recommender.sh

#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$SCRIPT_DIR/lib/dockerfile-parser.sh"
source "$SCRIPT_DIR/lib/base-image-analyzer.sh"

# Main analysis function
analyze_container_security() {
    local dockerfile="$1"
    local recommendations=()

    # Parse Dockerfile
    local base_image=$(extract_base_image "$dockerfile")
    local build_stages=$(count_build_stages "$dockerfile")
    local has_user_directive=$(check_user_directive "$dockerfile")

    # Analyze base image
    local image_analysis=$(analyze_base_image "$base_image")

    # Generate recommendations
    recommendations+=($(recommend_base_image "$base_image" "$image_analysis"))
    recommendations+=($(recommend_build_pattern "$dockerfile" "$build_stages"))
    recommendations+=($(check_security_best_practices "$dockerfile"))

    # Output report
    generate_container_report "$dockerfile" "$image_analysis" "${recommendations[@]}"
}
```

#### 5.2 Base Image Analyzer (`container-analysis/lib/base-image-analyzer.sh`)

```bash
# Analyze base image and recommend alternatives
analyze_base_image() {
    local image="$1"

    # Known image mappings
    declare -A DISTROLESS_ALTERNATIVES=(
        ["python:*"]="gcr.io/distroless/python3-debian12"
        ["node:*"]="gcr.io/distroless/nodejs20-debian12"
        ["golang:*"]="gcr.io/distroless/static-debian12"
        ["openjdk:*"]="gcr.io/distroless/java21-debian12"
        ["ubuntu:*"]="cgr.dev/chainguard/static"
        ["debian:*"]="cgr.dev/chainguard/static"
    )

    # Check for alternatives
    for pattern in "${!DISTROLESS_ALTERNATIVES[@]}"; do
        if [[ "$image" == $pattern ]]; then
            echo "{
                \"current\": \"$image\",
                \"recommended\": \"${DISTROLESS_ALTERNATIVES[$pattern]}\",
                \"type\": \"distroless\",
                \"benefits\": [\"smaller_size\", \"reduced_attack_surface\", \"no_shell\"]
            }"
            return
        fi
    done

    echo '{"current": "'"$image"'", "recommended": null}'
}
```

### Complexity: Medium
### Dependencies: None
### Estimated Files: 4

---

## Phase 6: Library Recommendation Engine

### Objective
AI-powered system that recommends improved library alternatives.

### Deliverables

#### 6.1 Library Recommender Module (`library-recommendations/`)

```bash
# library-recommender.sh

#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RAG_DIR="$SCRIPT_DIR/../../rag/supply-chain/library-recommendations"

source "$SCRIPT_DIR/lib/alternative-finder.sh"
source "$SCRIPT_DIR/lib/comparison-engine.sh"
source "$SCRIPT_DIR/lib/migration-estimator.sh"

# Main recommendation function
recommend_libraries() {
    local manifest="$1"
    local ecosystem="$2"
    local output_format="${3:-markdown}"

    # Extract dependencies
    local deps=$(extract_dependencies "$manifest" "$ecosystem")

    local recommendations=()

    for dep in $deps; do
        # Get current package health
        local health=$(get_package_health "$dep" "$ecosystem")

        # Check for known replacements
        local known_alt=$(check_known_replacements "$dep" "$ecosystem")

        # Find alternatives via deps.dev/snyk
        local api_alts=$(find_api_alternatives "$dep" "$ecosystem")

        # Use Claude for intelligent recommendation
        if should_recommend_alternative "$health" "$known_alt"; then
            local recommendation=$(get_ai_recommendation "$dep" "$health" "$known_alt" "$api_alts")
            recommendations+=("$recommendation")
        fi
    done

    # Generate output
    case "$output_format" in
        markdown)
            generate_markdown_report "${recommendations[@]}"
            ;;
        json)
            generate_json_report "${recommendations[@]}"
            ;;
    esac
}

# AI-powered recommendation
get_ai_recommendation() {
    local pkg="$1"
    local health="$2"
    local known_alt="$3"
    local api_alts="$4"

    # Load RAG context
    local rag_context=$(cat "$RAG_DIR"/*.md)

    # Build prompt
    local prompt="Analyze package '$pkg' and recommend alternatives.

Current health: $health
Known replacement: $known_alt
Potential alternatives: $api_alts

Based on the library recommendation guide:
$rag_context

Provide recommendation in JSON format with:
- recommendation: keep/consider_alternative/replace_urgently
- alternatives: array of {name, reason, migration_effort, improvements}
- rationale: explanation
- priority: critical/high/medium/low"

    call_claude_api "$prompt"
}
```

#### 6.2 Alternative Finder (`lib/alternative-finder.sh`)

```bash
# Find alternatives using multiple sources
find_alternatives() {
    local pkg="$1"
    local ecosystem="$2"

    # Source 1: Known replacements database
    local known=$(check_known_replacements "$pkg" "$ecosystem")

    # Source 2: deps.dev similar packages
    local similar=$(query_depsdev_similar "$pkg" "$ecosystem")

    # Source 3: Snyk advisor (if available)
    local snyk=$(query_snyk_alternatives "$pkg" "$ecosystem")

    # Source 4: Libraries.io (if API key available)
    local libio=$(query_librariesio "$pkg" "$ecosystem")

    # Merge and deduplicate
    merge_alternatives "$known" "$similar" "$snyk" "$libio"
}
```

#### 6.3 Comparison Engine (`lib/comparison-engine.sh`)

```bash
# Compare two packages
compare_packages() {
    local pkg1="$1"
    local pkg2="$2"
    local ecosystem="$3"

    local health1=$(get_package_health "$pkg1" "$ecosystem")
    local health2=$(get_package_health "$pkg2" "$ecosystem")

    # Compare metrics
    local security_delta=$(compare_security "$health1" "$health2")
    local maintenance_delta=$(compare_maintenance "$health1" "$health2")
    local size_delta=$(compare_bundle_size "$pkg1" "$pkg2" "$ecosystem")

    echo "{
        \"package1\": \"$pkg1\",
        \"package2\": \"$pkg2\",
        \"security_improvement\": $security_delta,
        \"maintenance_improvement\": $maintenance_delta,
        \"size_reduction\": \"$size_delta\",
        \"overall_improvement\": $(calculate_overall "$security_delta" "$maintenance_delta" "$size_delta")
    }"
}
```

### Complexity: High
### Dependencies: All previous phases, Claude API
### Estimated Files: 4

---

## Phase 7: Developer Productivity Features

### Objective
Enhance developer productivity through bundle optimization and technical debt analysis.

### Deliverables

#### 7.1 Bundle Size Analyzer (`package-health-analysis/lib/bundle-analyzer.sh`)

```bash
#!/bin/bash
# Analyze npm package bundle sizes and suggest optimizations

# Get bundle size from bundlephobia
get_bundle_size() {
    local pkg="$1"
    curl -s "https://bundlephobia.com/api/size?package=${pkg}" | jq '{
        size: .size,
        gzip: .gzip,
        dependencyCount: .dependencyCount,
        hasJSModule: .hasJSModule,
        hasJSNext: .hasJSNext,
        hasSideEffects: .hasSideEffects
    }'
}

# Analyze package.json for bundle optimization opportunities
analyze_bundle_opportunities() {
    local manifest="$1"
    local recommendations=()

    # Check for large packages with lighter alternatives
    local lodash_size=$(get_bundle_size "lodash" | jq -r '.gzip')
    local lodash_es_size=$(get_bundle_size "lodash-es" | jq -r '.gzip')

    # Check for tree-shaking support
    for pkg in $(jq -r '.dependencies | keys[]' "$manifest"); do
        local info=$(get_bundle_size "$pkg")
        local has_esm=$(echo "$info" | jq -r '.hasJSModule')
        local has_side_effects=$(echo "$info" | jq -r '.hasSideEffects')

        if [[ "$has_esm" == "false" ]]; then
            recommendations+=("$pkg: No ESM support, tree-shaking limited")
        fi
    done

    echo "${recommendations[@]}"
}
```

#### 7.2 Technical Debt Scorer (`lib/technical-debt-scorer.sh`)

```bash
#!/bin/bash
# Calculate technical debt score based on dependency staleness

calculate_dependency_debt() {
    local manifest="$1"
    local ecosystem="$2"

    local total_debt=0
    local packages=0

    for pkg in $(extract_dependencies "$manifest" "$ecosystem"); do
        local current_version=$(get_current_version "$pkg" "$manifest")
        local latest_version=$(get_latest_version "$pkg" "$ecosystem")
        local versions_behind=$(count_versions_behind "$pkg" "$current_version" "$latest_version")
        local days_behind=$(calculate_days_behind "$pkg" "$current_version")

        # Debt formula: versions behind * severity weight * time factor
        local pkg_debt=$((versions_behind * 10 + days_behind / 30))
        total_debt=$((total_debt + pkg_debt))
        packages=$((packages + 1))
    done

    echo "{
        \"total_debt_score\": $total_debt,
        \"average_debt_per_package\": $((total_debt / packages)),
        \"packages_analyzed\": $packages,
        \"debt_level\": \"$(classify_debt_level "$total_debt" "$packages")\"
    }"
}

classify_debt_level() {
    local total="$1"
    local count="$2"
    local avg=$((total / count))

    if [[ $avg -lt 50 ]]; then echo "healthy"
    elif [[ $avg -lt 150 ]]; then echo "moderate"
    elif [[ $avg -lt 300 ]]; then echo "concerning"
    else echo "critical"
    fi
}
```

### Complexity: Medium
### Dependencies: Phase 6
### Estimated Files: 2

---

## CLI Integration

### New Flags for supply-chain-scanner.sh

```bash
# Security & Risk Management
--normalize-versions      # Enable version normalization for vuln matching
--check-abandonment       # Check for abandoned packages
--check-typosquat         # Enable typosquatting detection
--check-confusion         # Detect dependency confusion risks

# Developer Productivity
--check-unused            # Detect unused dependencies
--recommend-libraries     # Generate library recommendations
--bundle-analysis         # Analyze bundle sizes (npm only)
--technical-debt          # Calculate dependency debt score

# Operations
--container-analysis      # Analyze Dockerfiles for hardening recommendations

# Convenience flags
--all-checks              # Enable all enhanced checks
--security-only           # Only security-related checks

# Output options
--format                  # markdown|json|sarif output format
--recommendations-format  # markdown|json for library recommendations
```

### Example Commands

```bash
# Full enhanced scan
./supply-chain-scanner.sh --repo owner/repo --all-checks

# Specific checks
./supply-chain-scanner.sh --repo owner/repo \
    --check-abandonment \
    --check-typosquat \
    --recommend-libraries

# Container security
./supply-chain-scanner.sh --repo owner/repo \
    --container-analysis \
    --recommendations-format json

# Local project scan
./supply-chain-scanner.sh --local /path/to/project \
    --check-unused \
    --recommend-libraries
```

---

## Testing Strategy

### Unit Tests

```bash
# tests/unit/test-version-normalizer.sh
test_npm_normalization() {
    assert_equals "1.0.0" "$(normalize_version 'v1.0.0' 'npm')"
    assert_equals "1.0.0" "$(normalize_version '1' 'npm')"
    assert_equals "1.2.0" "$(normalize_version '1.2' 'npm')"
}

test_pypi_normalization() {
    assert_equals "1.0.0a1" "$(normalize_version '1.0.0-alpha1' 'pypi')"
    assert_equals "1.0.0rc1" "$(normalize_version '1.0.0.RC1' 'pypi')"
}
```

### Integration Tests

```bash
# tests/integration/test-abandonment-detection.sh
test_abandoned_package_detection() {
    # Use known abandoned package
    local result=$(check_abandonment_status "left-pad" "npm")
    assert_contains "$result" '"status": "abandoned"'
}

# tests/integration/test-typosquat-detection.sh
test_typosquat_lodash() {
    local result=$(check_typosquat_risk "loadsh" "npm")
    assert_contains "$result" '"suspicious": true'
    assert_contains "$result" '"similar_to": "lodash"'
}
```

### Test Data

- Sample package.json with known problematic packages
- Sample requirements.txt with deprecated packages
- Sample Dockerfiles with non-optimal base images

---

## Success Metrics

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Version normalization accuracy | 99%+ | Test against known version pairs |
| Abandoned package detection | 95%+ recall | Compare against deps.dev data |
| Typosquat false positive rate | <5% | Manual review of flagged packages |
| Library recommendation acceptance | 70%+ helpful | User feedback |
| Scan time overhead | <30% increase | Benchmark before/after |

---

## Rollout Plan

### Week 1-2: Phase 1 (Foundation)
- [ ] Implement version-normalizer.sh
- [ ] Build popular-packages.sh database
- [ ] Unit tests for both libraries
- [ ] Documentation

### Week 3-4: Phase 2-3 (Detection)
- [ ] Implement abandonment-detector.sh
- [ ] Implement typosquat-detector.sh
- [ ] Integration with package-health-analyser.sh
- [ ] Integration tests

### Week 5-6: Phase 4-5 (Analysis)
- [ ] Implement unused-detector.sh
- [ ] Implement container-analysis module
- [ ] CLI flag integration
- [ ] Documentation

### Week 7-8: Phase 6 (AI Recommendations)
- [ ] Implement library-recommender.sh
- [ ] Claude API integration
- [ ] RAG prompt engineering
- [ ] End-to-end testing

### Week 9-10: Polish & Release
- [ ] Performance optimization
- [ ] Documentation updates
- [ ] ROADMAP.md updates
- [ ] Release notes
- [ ] v0.5.0 release

---

## RAG Knowledge Base Summary

Created documents:
1. `rag/supply-chain/version-normalization/version-normalization-guide.md`
2. `rag/supply-chain/malicious-package-detection/typosquatting-detection.md`
3. `rag/supply-chain/malicious-package-detection/abandoned-package-detection.md`
4. `rag/supply-chain/unused-dependency-detection/unused-dependencies-guide.md`
5. `rag/supply-chain/hardened-images/gold-images-guide.md`
6. `rag/supply-chain/library-recommendations/library-recommendation-guide.md`

Prompt created:
- `prompts/supply-chain/implementation-planning-prompt.md`
