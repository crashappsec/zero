# Gibson Powers Analyzers - Standardization Plan

## Current State Inventory

### Analyzers
1. **certificate-analyzer** (cert-analyzer.sh)
2. **chalk-build-analyzer** (chalk-build-analyzer.sh)
3. **code-ownership** (ownership-analyzer.sh)
4. **dora-metrics** (dora-analyzer.sh)
5. **package-health-analysis** (package-health-analyzer.sh)
6. **provenance-analysis** (provenance-analyzer.sh)
7. **vulnerability-analysis** (vulnerability-analyzer.sh)

### Current Flag Support Matrix

| Analyzer | --org | --repo | --claude | --compare | --format | --output | --keep-clone |
|----------|-------|--------|----------|-----------|----------|----------|--------------|
| certificate-analyzer | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| chalk-build-analyzer | ❌ | ❌ | ✅ | ❌ | ✅ | ✅ | ❌ |
| code-ownership | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| dora-metrics | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ |
| package-health-analysis | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| provenance-analysis | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |
| vulnerability-analysis | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |

### Libraries

#### Global (utils/lib/)
- **github.sh** - Basic org repo listing
- **config-loader.sh** - Configuration loading
- **claude-cost.sh** - API cost tracking (referenced but not found)

#### Analyzer-Specific
- **code-ownership/lib/** - 10 library files including enhanced github.sh
- **package-health-analysis/lib/** - 4 library files

## Standardization Goals

### 1. Common Flag Standards

**Required flags for all analyzers:**
- `--format FORMAT` - Output format (text|json|markdown|csv) - default: markdown
- `--output FILE` - Write to file instead of stdout
- `--claude` - Enable Claude AI analysis
- `-k, --api-key KEY` - Anthropic API key
- `-h, --help` - Show help

**For analyzers that work with repositories:**
- `--org ORGANIZATION` - Analyze all repos in GitHub organization
- `--repo OWNER/REPO` - Analyze single repository
- `--repos REPO1 REPO2...` - Analyze multiple repositories
- `--keep-clone` - Keep cloned repos (don't cleanup)

**For analyzers with Claude support:**
- `--compare` - Run both basic and Claude modes side-by-side

### 2. GitHub Library Consolidation

**Action:** Merge the two github.sh libraries
- Base: `utils/lib/github.sh` (simple org listing)
- Enhanced: `utils/code-ownership/lib/github.sh` (profile caching, mapping)
- Result: Single comprehensive `utils/lib/github.sh` with all features

**Features to include:**
- `list_org_repos()` - List all repos in an organization
- `init_github_cache()` - Initialize profile cache
- `lookup_github_profile()` - Lookup GitHub username from email
- `get_github_profile()` - Get cached profile or lookup
- `cleanup_github_cache()` - Clean up cache file
- Support for GITHUB_TOKEN environment variable

### 3. Cleanup Standardization

**All analyzers that clone repos must:**
- Use `mktemp -d` for temporary directories
- Set `TEMP_DIR` variable
- Define `cleanup()` function
- Use `trap cleanup EXIT` for automatic cleanup
- Support `--keep-clone` flag to disable cleanup
- Handle errors with `return` not `exit` in functions

### 4. Configuration Standards

**Default values:**
```bash
FORMAT="markdown"              # Default output format
OUTPUT_FILE=""                  # Default: stdout
USE_CLAUDE=false                # Default: basic mode
COMPARE_MODE=false              # Default: single mode
CLEANUP=true                    # Default: cleanup temps
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
GITHUB_TOKEN="${GITHUB_TOKEN:-}"
```

### 5. Usage Documentation Standards

**Every analyzer README.md must include:**
1. **Title and Description** - What it does
2. **Prerequisites** - Required tools, tokens
3. **Installation** - How to install/setup
4. **Quick Start** - Simplest usage example
5. **Usage** - Full options table
6. **Examples** - Common use cases:
   - Single repository analysis
   - Organization scanning
   - Multiple repositories
   - Claude AI analysis
   - Compare mode
   - Output formats
7. **Output** - What to expect
8. **Configuration** - Config file options (if applicable)
9. **Troubleshooting** - Common issues

### 6. Code Structure Standards

**All analyzers should follow:**
```bash
#!/bin/bash
# Script header with SPDX license

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source global libraries
source "$REPO_ROOT/utils/lib/github.sh"
source "$REPO_ROOT/utils/lib/claude-cost.sh" 2>/dev/null || true

# Default values
...

# Cleanup function
cleanup() { ... }
trap cleanup EXIT

# Helper functions
...

# Main analysis function
...

# Argument parsing
...

# Main execution
...
```

## Implementation Plan

### Phase 1: Core Infrastructure (Priority 1)
- [x] Fix organization scanning bug (DONE)
- [ ] Create consolidated utils/lib/github.sh
- [ ] Create utils/lib/claude-cost.sh
- [ ] Ensure all analyzers use trap cleanup EXIT

### Phase 2: Flag Standardization (Priority 1)
- [ ] Add --org and --repo support to certificate-analyzer
- [ ] Add --org and --repo support to chalk-build-analyzer
- [ ] Add --org and --repo support to dora-metrics
- [ ] Add --compare mode to provenance-analyzer
- [ ] Add --compare mode to vulnerability-analyzer
- [ ] Standardize --format across all analyzers
- [ ] Add --keep-clone to all repo-based analyzers

### Phase 3: Documentation (Priority 2)
- [ ] Create standardized README template
- [ ] Update certificate-analyzer README
- [ ] Update chalk-build-analyzer README
- [ ] Update code-ownership README
- [ ] Update dora-metrics README
- [ ] Update package-health-analysis README
- [ ] Update provenance-analysis README
- [ ] Update vulnerability-analysis README

### Phase 4: Testing & Validation (Priority 2)
- [ ] Test each analyzer with --org
- [ ] Test each analyzer with --claude
- [ ] Test each analyzer with --compare (where applicable)
- [ ] Test cleanup with interrupts (Ctrl+C)
- [ ] Verify all examples in documentation work

## Benefits

1. **Consistency** - Users know what to expect across all tools
2. **Maintainability** - Single source of truth for common functionality
3. **Reliability** - Proper cleanup prevents disk issues
4. **Discoverability** - Standard options make tools easier to learn
5. **Documentation** - Comprehensive docs help adoption
6. **Quality** - Coding standards improve code quality

## Next Steps

1. Review and approve this plan
2. Begin with Phase 1 (Core Infrastructure)
3. Proceed systematically through each phase
4. Test thoroughly at each step
5. Update main README with standardized tool overview
