# Legal Analyzer Architecture Alignment Plan

**Date:** 2025-11-23
**Status:** Planning Phase
**Goal:** Align legal analyzer with supply chain architecture patterns

---

## Executive Summary

This document outlines the plan to modernize the legal analyzer to match the architectural patterns, argument structure, and integration approach used in the supply chain analyzers (package health, provenance, vulnerability).

## Current State Analysis

### Legal Analyzer Current Features

**File:** `/utils/legal-review/legal-analyser.sh` (959 lines)

#### Existing Capabilities
- ✅ License scanning
- ✅ Secret detection
- ✅ Content policy scanning (profanity, non-inclusive language)
- ✅ Claude AI integration (`--claude` flag)
- ✅ RAG context loading
- ✅ Output formats: markdown, json
- ✅ Verbose logging
- ✅ Multiple scan modes (`--licenses-only`, `--secrets-only`, `--content-only`)

#### Current Arguments
```bash
--repo OWNER/REPO          # Analyze GitHub repository
--path PATH                # Analyze local path
--licenses-only            # Scan licenses only
--secrets-only             # Scan secrets only
--content-only             # Scan content policy only
--format FORMAT            # Output format: markdown, json
--output FILE              # Write output to file
--claude                   # Use Claude AI for enhanced analysis
--verbose                  # Enable verbose output
-h, --help                 # Show help
```

#### Missing Compared to Supply Chain Analyzers
- ❌ `--parallel` flag for batch/parallel processing
- ❌ Batch API support (if applicable)
- ❌ Integration with main supply-chain-scanner.sh
- ❌ Multi-repository support (`--org`, multiple `--repo`)
- ❌ Shared SBOM/repository usage
- ❌ `--compare` mode (basic vs Claude side-by-side)
- ❌ Cost tracking for Claude API usage
- ❌ `--local-path` for pre-cloned repos
- ❌ Consistent error handling patterns
- ❌ Enhanced Claude prompts with best practices from RAG

---

## Supply Chain Architecture Patterns

### Common Patterns Across Analyzers

#### 1. **Argument Structure**

**Package Health Analyzer:**
```bash
--repo OWNER/REPO          # Single repository
--org ORGANIZATION         # All repos in org
--sbom FILE                # Analyze existing SBOM
--local-path PATH          # Pre-cloned repository
--format FORMAT            # json, markdown, table
--output FILE              # Output to file
--no-version-analysis      # Skip version analysis
--no-deprecation-check     # Skip deprecation
--claude                   # Claude AI analysis
--parallel                 # Batch API mode
--compare                  # Basic vs Claude comparison
-k, --api-key KEY          # Anthropic API key
--verbose                  # Verbose logging
```

**Provenance Analyzer:**
```bash
--org ORG_NAME             # Scan all repos in org
--repo OWNER/REPO          # Scan specific repository
--verify-signatures        # Cryptographic verification
--min-level LEVEL          # Require minimum SLSA level
--strict                   # Fail on missing provenance
--claude                   # Claude AI analysis
--parallel                 # Parallel processing
--jobs N                   # Number of parallel jobs
--local-path PATH          # Pre-cloned repository
```

**Vulnerability Analyzer:**
```bash
--org ORG_NAME             # Scan all repos in org
--repo OWNER/REPO          # Scan specific repository
-t, --taint-analysis       # Enable call graph analysis
-p, --prioritize           # Intelligent prioritization
--claude                   # Claude AI analysis
--local-path PATH          # Pre-cloned repository
```

#### 2. **Integration with supply-chain-scanner.sh**

All analyzers integrate via standardized functions:
```bash
run_package_health_analysis() {
    # Build command with optional flags
    cmd="$analyser"
    [[ "$USE_CLAUDE" == "true" ]] && cmd="$cmd --claude"
    [[ "$PARALLEL" == "true" ]] && cmd="$cmd --parallel"
    # Use shared SBOM if available
    [[ -n "$SHARED_SBOM_FILE" ]] && cmd="$cmd --sbom-file $SHARED_SBOM_FILE"
    # Use shared repository if available
    [[ -n "$SHARED_REPO_DIR" ]] && cmd="$cmd --local-path $SHARED_REPO_DIR"
    eval "$cmd"
}
```

#### 3. **Claude AI Integration Pattern**

**Enhanced Prompts with RAG:**
```bash
analyze_with_claude() {
    local data="$1"
    local model="claude-sonnet-4-20250514"

    # Comprehensive prompt with best practices from RAG
    local prompt="Analyze this [analyzer type] data and provide actionable insights...

    ## Analysis Focus Areas:
    ### 1. [Critical Area] (from RAG best practices)
    ### 2. [Important Area] (from RAG best practices)
    ...

    ## Output Format:
    ### Executive Summary
    ### Detailed Findings
    ### Prioritized Remediation Plan

    Data:
    $data"

    # API call with error handling
    response=$(curl -s https://api.anthropic.com/v1/messages ...)

    # Check for API errors
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        error_type=$(echo "$response" | jq -r '.error.type')
        error_message=$(echo "$response" | jq -r '.error.message')
        echo "Error: Claude API request failed - $error_type: $error_message" >&2
        return 1
    fi

    # Extract and validate analysis
    analysis=$(echo "$response" | jq -r '.content[0].text // empty')
    [[ -z "$analysis" ]] && { echo "Error: No analysis returned" >&2; return 1; }

    echo "$analysis"
}
```

**Cost Tracking:**
```bash
# Load cost tracking if using Claude
if [[ "$USE_CLAUDE" == "true" ]] || [[ "$COMPARE_MODE" == "true" ]]; then
    if [ -f "$REPO_ROOT/lib/claude-cost.sh" ]; then
        source "$REPO_ROOT/lib/claude-cost.sh"
        init_cost_tracking
    fi
fi

# Record API usage
if command -v record_api_usage &> /dev/null; then
    record_api_usage "$response" "$model" > /dev/null
fi

# Display cost summary
if command -v display_api_cost_summary &> /dev/null; then
    display_api_cost_summary
fi
```

#### 4. **Batch/Parallel Processing Patterns**

**Pattern A: Batch API (Package Health)**
```bash
if [[ "$PARALLEL" == "true" ]]; then
    # Prepare packages for batch request
    batch_packages=$(echo "$packages" | jq -s 'map({...})')

    # Single batch API call
    batch_response=$(get_versions_batch "$batch_packages")

    # Build lookup map and process
    version_lookup=$(echo "$batch_response" | jq '...')

    # Aggregate results
    ...
fi
```

**Pattern B: Parallel Processing (Provenance)**
```bash
if [[ "$PARALLEL" == "true" ]]; then
    # Export functions for subshells
    export -f check_npm_provenance
    export -f parse_purl

    # Process in parallel using xargs
    echo "$components" | nl | xargs -P "$PARALLEL_JOBS" -I {} bash -c '...'

    # Aggregate results from temp files
    ...
fi
```

---

## Alignment Plan

### Phase 1: Argument Structure Harmonization

#### Add Missing Arguments

**Multi-Repository Support:**
```bash
--org ORG_NAME             # Scan all repos in GitHub organization
# Allow multiple --repo flags
```

**Shared Resource Support:**
```bash
--local-path PATH          # Use pre-cloned repository (skips cloning)
--sbom FILE                # Use existing SBOM file
```

**Comparison Mode:**
```bash
--compare                  # Run both basic and Claude modes side-by-side
```

**API Key Management:**
```bash
-k, --api-key KEY          # Anthropic API key (or use ANTHROPIC_API_KEY env var)
```

#### Update Existing Arguments

**Remove redundant flags:**
- Keep: `--licenses-only`, `--secrets-only`, `--content-only`
- Consider: Consolidate into `--scan-types` with comma-separated values

**Standardize format:**
```bash
--format FORMAT            # json, markdown, table (add table format)
```

### Phase 2: Parallel/Batch Processing

#### Determine Batch API Availability

**For License Scanning:**
- Check if license APIs (npm, PyPI, etc.) have batch endpoints
- If yes: Implement batch API pattern
- If no: Implement parallel processing pattern

**For Secret Detection:**
- Sequential is likely fine (file-based, local scanning)
- Consider parallel file processing for large repositories

**For Content Policy:**
- Parallel file processing pattern (similar to provenance)

#### Implementation Approach

```bash
# Add flags
PARALLEL=false
PARALLEL_JOBS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "4")

# Add to usage
--parallel                 # Enable parallel analysis (faster)
--jobs N                   # Number of parallel jobs (default: CPU count)

# Implement parallel scanning
if [[ "$PARALLEL" == "true" ]]; then
    # For license scanning
    scan_licenses_parallel "$path" "$PARALLEL_JOBS"

    # For content policy
    scan_content_parallel "$path" "$PARALLEL_JOBS"
else
    # Sequential mode (default)
    scan_licenses "$path"
    scan_content "$path"
fi
```

### Phase 3: Supply Chain Scanner Integration

#### Create Integration Function

Add to `supply-chain-scanner.sh`:

```bash
# Function to run legal analysis
run_legal_analysis() {
    local target=$(normalize_target "$1")
    local analyser="$SCRIPT_DIR/../legal-review/legal-analyser.sh"

    if [[ ! -f "$analyser" ]]; then
        echo -e "${RED}✗ Legal analyser not found${NC}"
        return 1
    fi

    # Build command with optional flags
    local cmd="$analyser"

    # Add Claude flag if enabled
    if [[ "$USE_CLAUDE" == "true" ]]; then
        cmd="$cmd --claude"
    fi

    # Add parallel flag if enabled
    if [[ "$PARALLEL" == "true" ]]; then
        cmd="$cmd --parallel"
    fi

    # Use shared SBOM if available (for license extraction)
    if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
        cmd="$cmd --sbom $SHARED_SBOM_FILE"
    fi

    # Use shared repository if available
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        cmd="$cmd --local-path $SHARED_REPO_DIR $target"
    else
        cmd="$cmd --repo $target"
    fi

    eval "$cmd"
}
```

#### Add Module Flag

```bash
# In supply-chain-scanner.sh
MODULES:
    --vulnerability, -v     Run vulnerability analysis
    --provenance, -p        Run provenance analysis (SLSA)
    --package-health        Run package health analysis
    --legal                 Run legal compliance analysis
    --all, -a               Run all analysis modules
```

#### Add to analyze_target Function

```bash
analyze_target() {
    local target="$1"

    # ... existing code ...

    if should_run_module "legal"; then
        echo -e "\n${CYAN}Running Legal Analysis...${NC}"
        run_legal_analysis "$target"
    fi
}
```

### Phase 4: Enhanced Claude Integration

#### Update Claude Prompts

**License Analysis Prompt Enhancement:**
```bash
claude_analyze_licenses() {
    local scan_results="$1"

    local prompt="Analyze this license compliance data following best practices from our RAG documentation.

## Analysis Focus Areas:

### 1. License Compatibility & Conflicts
- Identify incompatible license combinations (e.g., GPL + proprietary)
- Explain copyleft implications for commercial use
- Flag viral license risks (AGPL, GPL v3, etc.)
- Check weak copyleft compatibility (LGPL, MPL)

### 2. Risk Assessment by License Category
- **Permissive** (MIT, Apache, BSD): Generally safe, attribution required
- **Weak Copyleft** (LGPL, MPL): Safe if dynamic linking, risky if static
- **Strong Copyleft** (GPL v2/v3): Code must be open-sourced
- **Network Copyleft** (AGPL): Triggers on network use
- **Proprietary/Custom**: Requires legal review
- **Unknown/Missing**: Critical compliance risk

### 3. Compliance Requirements
- Attribution requirements (copyright notices, license files)
- Source code disclosure obligations
- Patent grant implications
- Trademark restrictions
- Redistribution conditions

### 4. Business Impact Assessment
- Commercial use restrictions
- SaaS deployment implications
- Distribution requirements
- Derivative work definitions

### 5. Remediation Recommendations
For each violation:
- Specific action required (remove, replace, add attribution, etc.)
- Alternative libraries with compatible licenses
- Migration complexity estimate
- Timeline for remediation

## Output Format:

### Executive Summary
- Overall compliance status (Pass/Warning/Fail)
- Critical issues requiring immediate attention
- Total licenses detected and breakdown by category

### License Inventory
Table format:
| Package | Version | License | Category | Risk Level | Status |

### Compatibility Analysis
- License conflict matrix
- Copyleft implications
- Commercial use assessment

### Prioritized Action Items
1. **Critical** (0-24h): Violations blocking release
2. **High** (1-7d): Significant compliance risks
3. **Medium** (1-30d): Attribution fixes, documentation
4. **Low** (30-90d): Best practice improvements

Data:
$scan_results"

    call_claude_api "$prompt"
}
```

**Content Policy Prompt Enhancement:**
```bash
claude_analyze_content() {
    local scan_results="$1"

    local prompt="Analyze this content policy compliance data with focus on professional standards and inclusive language.

## Analysis Focus Areas:

### 1. Profanity & Offensive Language
- Direct profanity in code/comments
- Offensive variable names
- Inappropriate commit messages
- Context evaluation (technical vs offensive)

### 2. Non-Inclusive Language
- Master/slave → primary/replica, leader/follower
- Whitelist/blacklist → allowlist/denylist
- Blackhat/whitehat → malicious/ethical
- Gendered terms → gender-neutral alternatives
- Cultural sensitivity issues

### 3. Business Risk Assessment
- PR and brand reputation impact
- Team morale and inclusivity
- Customer perception risks
- Legal/HR compliance issues

### 4. Context-Aware Recommendations
- Distinguish technical terms from violations (e.g., "git master" OK)
- Preserve necessary technical terminology
- Provide context-appropriate alternatives
- Suggest gradual migration paths

## Best Practices from RAG:
- Use professional, inclusive language in all code and documentation
- Update terminology proactively during refactoring
- Add linter rules to prevent future violations
- Document exceptions with clear justification

## Output Format:

### Executive Summary
- Content policy compliance status
- Total violations by category
- Immediate actions required

### Findings by Category
For each issue:
- Location (file:line)
- Current term
- Recommended replacement
- Context analysis
- Priority level

### Remediation Plan
1. **High Priority**: Offensive language, inappropriate content
2. **Medium Priority**: Non-inclusive terms in public APIs/docs
3. **Low Priority**: Internal code, comments (gradual migration)

Data:
$scan_results"

    call_claude_api "$prompt"
}
```

#### Add Error Handling

```bash
call_claude_api() {
    local prompt="$1"
    local model="${2:-claude-sonnet-4-20250514}"

    # Check for API key
    if [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
        echo "Error: ANTHROPIC_API_KEY environment variable not set" >&2
        echo "Set it with: export ANTHROPIC_API_KEY=your-api-key" >&2
        echo "Or pass via: -k your-api-key" >&2
        return 1
    fi

    # Make API call
    response=$(curl -s -X POST https://api.anthropic.com/v1/messages ...)

    # Check for API errors
    if echo "$response" | jq -e '.error' >/dev/null 2>&1; then
        error_type=$(echo "$response" | jq -r '.error.type')
        error_message=$(echo "$response" | jq -r '.error.message')
        echo "Error: Claude API request failed - $error_type: $error_message" >&2
        return 1
    fi

    # Extract and validate content
    analysis=$(echo "$response" | jq -r '.content[0].text // empty')

    if [[ -z "$analysis" ]]; then
        echo "Error: No analysis returned from Claude API" >&2
        echo "Response: $response" >&2
        return 1
    fi

    echo "$analysis"
}
```

#### Add Cost Tracking

```bash
# In main function
main() {
    parse_args "$@"

    # Load cost tracking if using Claude
    if [[ "$USE_CLAUDE" == "true" ]] || [[ "$COMPARE_MODE" == "true" ]]; then
        if [ -f "$REPO_ROOT/lib/claude-cost.sh" ]; then
            source "$REPO_ROOT/lib/claude-cost.sh"
            init_cost_tracking
        fi
    fi

    # ... existing analysis code ...

    # Display cost summary
    if [[ "$USE_CLAUDE" == "true" ]] && command -v display_api_cost_summary &> /dev/null; then
        echo ""
        echo "API Cost Summary:"
        display_api_cost_summary
    fi
}
```

### Phase 5: Output Format Standardization

#### Add Table Format

```bash
format_table() {
    local json_data=$1

    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║            Legal Compliance Analysis Summary               ║"
    echo "╠════════════════════════════════════════════════════════════╣"
    echo "║ License Violations:       $(jq -r '.summary.license_violations' <<< "$json_data") ║"
    echo "║ Content Issues:           $(jq -r '.summary.content_issues' <<< "$json_data") ║"
    echo "║ Secret Exposures:         $(jq -r '.summary.secret_exposures' <<< "$json_data") ║"
    echo "╚════════════════════════════════════════════════════════════╝"

    # Detailed license table
    echo ""
    echo "Licenses:"
    jq -r '.licenses[] | [.package, .license, .risk_level] | @tsv' <<< "$json_data" | \
        column -t -s $'\t' -N "Package,License,Risk"
}
```

#### Standardize JSON Output

```json
{
  "scan_metadata": {
    "timestamp": "2025-11-23T...",
    "target_repository": "owner/repo",
    "scan_types": ["licenses", "secrets", "content"],
    "analyser_version": "1.0.0",
    "analyser_type": "basic|claude"
  },
  "summary": {
    "license_violations": 5,
    "content_issues": 12,
    "secret_exposures": 0,
    "overall_status": "warning"
  },
  "licenses": [...],
  "content_policy": {...},
  "secrets": {...},
  "claude_analysis": "..." // If --claude enabled
}
```

### Phase 6: Testing & Validation

#### Test Cases

**1. Basic Mode:**
```bash
./legal-analyser.sh --repo crashappsec/chalk
```

**2. Claude Mode:**
```bash
export ANTHROPIC_API_KEY=...
./legal-analyser.sh --repo crashappsec/chalk --claude
```

**3. Parallel Mode:**
```bash
./legal-analyser.sh --repo crashappsec/chalk --parallel --jobs 8
```

**4. Comparison Mode:**
```bash
./legal-analyser.sh --repo crashappsec/chalk --compare
```

**5. Multi-Repo via Supply Chain Scanner:**
```bash
./supply-chain-scanner.sh --legal --parallel --org crashappsec
```

**6. Integration with Other Modules:**
```bash
./supply-chain-scanner.sh --all --parallel --claude --repo crashappsec/chalk
```

---

## Implementation Roadmap

### Week 1: Argument Structure
- [ ] Add `--org`, `--local-path`, `--sbom`, `--compare`, `-k` flags
- [ ] Update argument parsing
- [ ] Update usage documentation
- [ ] Test backward compatibility

### Week 2: Parallel Processing
- [ ] Research batch API availability for license scanning
- [ ] Implement parallel file processing for content policy
- [ ] Add `--parallel` and `--jobs` flags
- [ ] Test performance improvements

### Week 3: Supply Chain Integration
- [ ] Create `run_legal_analysis` function in supply-chain-scanner.sh
- [ ] Add `--legal` module flag
- [ ] Implement shared SBOM/repo support
- [ ] Test multi-module execution

### Week 4: Enhanced Claude Integration
- [ ] Update Claude prompts with best practices from RAG
- [ ] Add comprehensive error handling
- [ ] Implement cost tracking
- [ ] Add `--compare` mode

### Week 5: Output Standardization
- [ ] Add table format support
- [ ] Standardize JSON structure
- [ ] Update markdown formatting
- [ ] Ensure consistency across formats

### Week 6: Testing & Documentation
- [ ] Comprehensive testing (all modes, all flags)
- [ ] Update README
- [ ] Add examples to documentation
- [ ] Performance benchmarking

---

## Success Criteria

### Functional Requirements
- ✅ All supply chain analyzer flags supported
- ✅ Integration with supply-chain-scanner.sh working
- ✅ Parallel processing functional
- ✅ Claude AI integration enhanced
- ✅ Cost tracking implemented
- ✅ All output formats consistent

### Performance Requirements
- ✅ Parallel mode 3-5x faster than sequential
- ✅ Batch API mode (if applicable) 5-10x faster
- ✅ Large repo scanning completes in reasonable time (<5min for typical repo)

### Quality Requirements
- ✅ Backward compatible with existing scripts
- ✅ Comprehensive error handling
- ✅ Clear, actionable error messages
- ✅ Consistent with supply chain architecture

---

## References

### Related Files
- `/utils/supply-chain/package-health-analysis/package-health-analyser.sh`
- `/utils/supply-chain/provenance-analysis/provenance-analyser.sh`
- `/utils/supply-chain/vulnerability-analysis/vulnerability-analyser.sh`
- `/utils/supply-chain/supply-chain-scanner.sh`
- `/utils/legal-review/legal-analyser.sh`

### Best Practices Documentation
- `/rag/legal-review/license-compliance-guide.md`
- `/rag/legal-review/content-policy-guide.md`
- `/rag/supply-chain/package-health/package-management-best-practices.md`

### Architecture Documents
- `/ROADMAP.md`
- `/TEST_RESULTS.md` (supply chain)

---

**Last Updated:** 2025-11-23
**Next Review:** After Phase 1 completion
