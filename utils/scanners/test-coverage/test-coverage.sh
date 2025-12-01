#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Test Coverage - Data Collector
# Analyzes test files, frameworks, and coverage indicators
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./test-coverage-data.sh [options] <target>
# Output: JSON with test inventory, frameworks, and coverage estimation
#############################################################################

set -e

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
REPO=""
ORG=""
TEMP_DIR=""
CLEANUP=true
TARGET=""

usage() {
    cat << EOF
Test Coverage - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --repo OWNER/REPO       GitHub repository (looks in phantom cache)
    --org ORG               GitHub org (uses first repo found in phantom cache)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - summary: test file count, test-to-code ratio, coverage estimate
    - test_frameworks: detected testing frameworks
    - test_types: unit, integration, e2e test breakdown
    - mocking_libraries: detected mocking tools
    - uncovered_directories: directories without tests

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.phantom/projects/foo/repo
    $0 -o test-coverage.json /path/to/project

EOF
    exit 0
}

# Clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}" >&2
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Cloned${NC}" >&2
        return 0
    else
        echo '{"error": "Failed to clone repository"}'
        exit 1
    fi
}

# Cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Count source files by language
count_source_files() {
    local repo_dir="$1"

    # JavaScript/TypeScript - only exclude directories named exactly "test" or "tests" or "__tests__"
    local js_files=$(find "$repo_dir" -type f \( -name "*.js" -o -name "*.jsx" \) \
        ! -path "*node_modules*" ! -path "*vendor*" ! -path "*.git*" ! -path "*dist*" ! -path "*build*" \
        ! -name "*.test.js" ! -name "*.spec.js" ! -name "*.test.jsx" ! -name "*.spec.jsx" \
        ! -path "*/__tests__/*" ! -path "*/tests/*" ! -path "*/test/*" 2>/dev/null | wc -l | tr -d ' ')

    local ts_files=$(find "$repo_dir" -type f \( -name "*.ts" -o -name "*.tsx" \) \
        ! -path "*node_modules*" ! -path "*vendor*" ! -path "*.git*" ! -path "*dist*" ! -path "*build*" \
        ! -name "*.test.ts" ! -name "*.spec.ts" ! -name "*.test.tsx" ! -name "*.spec.tsx" \
        ! -path "*/__tests__/*" ! -path "*/tests/*" ! -path "*/test/*" 2>/dev/null | wc -l | tr -d ' ')

    # Python
    local py_files=$(find "$repo_dir" -type f -name "*.py" \
        ! -path "*vendor*" ! -path "*.git*" ! -path "*venv*" ! -path "*__pycache__*" \
        ! -name "test_*.py" ! -name "*_test.py" ! -path "*/tests/*" ! -path "*/test/*" 2>/dev/null | wc -l | tr -d ' ')

    # Java
    local java_files=$(find "$repo_dir" -type f -name "*.java" \
        ! -path "*vendor*" ! -path "*.git*" ! -path "*/test/*" ! -path "*/tests/*" ! -name "*Test.java" 2>/dev/null | wc -l | tr -d ' ')

    # Go
    local go_files=$(find "$repo_dir" -type f -name "*.go" \
        ! -path "*vendor*" ! -path "*.git*" ! -name "*_test.go" 2>/dev/null | wc -l | tr -d ' ')

    # Ruby
    local rb_files=$(find "$repo_dir" -type f -name "*.rb" \
        ! -path "*vendor*" ! -path "*.git*" ! -path "*/spec/*" ! -path "*/test/*" ! -path "*/tests/*" 2>/dev/null | wc -l | tr -d ' ')

    # Rust
    local rs_files=$(find "$repo_dir" -type f -name "*.rs" \
        ! -path "*vendor*" ! -path "*.git*" ! -path "*/test/*" ! -path "*/tests/*" 2>/dev/null | wc -l | tr -d ' ')

    local total=$((js_files + ts_files + py_files + java_files + go_files + rb_files + rs_files))

    jq -n \
        --argjson js "$js_files" \
        --argjson ts "$ts_files" \
        --argjson py "$py_files" \
        --argjson java "$java_files" \
        --argjson go "$go_files" \
        --argjson rb "$rb_files" \
        --argjson rs "$rs_files" \
        --argjson total "$total" \
        '{
            "javascript": $js,
            "typescript": $ts,
            "python": $py,
            "java": $java,
            "go": $go,
            "ruby": $rb,
            "rust": $rs,
            "total": $total
        }'
}

# Count test files
count_test_files() {
    local repo_dir="$1"
    local test_files='{}'

    # JavaScript/TypeScript tests
    local js_unit=$(find "$repo_dir" -type f \( -name "*.test.js" -o -name "*.spec.js" -o -name "*.test.jsx" -o -name "*.spec.jsx" \) \
        ! -path "*node_modules*" ! -path "*dist*" ! -path "*build*" 2>/dev/null | wc -l | tr -d ' ')
    local ts_unit=$(find "$repo_dir" -type f \( -name "*.test.ts" -o -name "*.spec.ts" -o -name "*.test.tsx" -o -name "*.spec.tsx" \) \
        ! -path "*node_modules*" ! -path "*dist*" ! -path "*build*" 2>/dev/null | wc -l | tr -d ' ')

    # Python tests
    local py_tests=$(find "$repo_dir" -type f \( -name "test_*.py" -o -name "*_test.py" \) \
        ! -path "*venv*" ! -path "*__pycache__*" 2>/dev/null | wc -l | tr -d ' ')

    # Java tests
    local java_tests=$(find "$repo_dir" -type f -name "*Test.java" \
        ! -path "*vendor*" 2>/dev/null | wc -l | tr -d ' ')

    # Go tests
    local go_tests=$(find "$repo_dir" -type f -name "*_test.go" \
        ! -path "*vendor*" 2>/dev/null | wc -l | tr -d ' ')

    # Ruby tests (rspec)
    local rb_tests=$(find "$repo_dir" -type f -name "*_spec.rb" \
        ! -path "*vendor*" 2>/dev/null | wc -l | tr -d ' ')

    # E2E tests
    local e2e_tests=$(find "$repo_dir" -type f \( -name "*.e2e.ts" -o -name "*.e2e.js" -o -name "*.e2e-spec.ts" \) \
        ! -path "*node_modules*" 2>/dev/null | wc -l | tr -d ' ')
    local cypress_tests=$(find "$repo_dir" -type f -name "*.cy.js" -o -name "*.cy.ts" \
        ! -path "*node_modules*" 2>/dev/null | wc -l | tr -d ' ')
    local playwright_tests=$(find "$repo_dir" -type f -name "*.spec.ts" -path "*e2e*" -o -name "*.spec.ts" -path "*playwright*" \
        ! -path "*node_modules*" 2>/dev/null | wc -l | tr -d ' ')

    # Integration tests
    local integration=$(find "$repo_dir" -type f \( -name "*.integration.test.*" -o -name "*.int.test.*" \) \
        ! -path "*node_modules*" 2>/dev/null | wc -l | tr -d ' ')

    local total=$((js_unit + ts_unit + py_tests + java_tests + go_tests + rb_tests + e2e_tests + cypress_tests + playwright_tests + integration))

    jq -n \
        --argjson js "$js_unit" \
        --argjson ts "$ts_unit" \
        --argjson py "$py_tests" \
        --argjson java "$java_tests" \
        --argjson go "$go_tests" \
        --argjson rb "$rb_tests" \
        --argjson e2e "$e2e_tests" \
        --argjson cypress "$cypress_tests" \
        --argjson playwright "$playwright_tests" \
        --argjson integration "$integration" \
        --argjson total "$total" \
        '{
            "javascript_unit": $js,
            "typescript_unit": $ts,
            "python": $py,
            "java": $java,
            "go": $go,
            "ruby_rspec": $rb,
            "e2e": $e2e,
            "cypress": $cypress,
            "playwright": $playwright,
            "integration": $integration,
            "total": $total
        }'
}

# Detect test frameworks from dependencies
detect_test_frameworks() {
    local repo_dir="$1"
    local frameworks="[]"

    # Check package.json
    if [[ -f "$repo_dir/package.json" ]]; then
        local pkg=$(cat "$repo_dir/package.json")

        # Jest
        if echo "$pkg" | jq -e '.devDependencies.jest or .dependencies.jest' > /dev/null 2>&1; then
            frameworks=$(echo "$frameworks" | jq '. + ["jest"]')
        fi

        # Mocha
        if echo "$pkg" | jq -e '.devDependencies.mocha or .dependencies.mocha' > /dev/null 2>&1; then
            frameworks=$(echo "$frameworks" | jq '. + ["mocha"]')
        fi

        # Vitest
        if echo "$pkg" | jq -e '.devDependencies.vitest or .dependencies.vitest' > /dev/null 2>&1; then
            frameworks=$(echo "$frameworks" | jq '. + ["vitest"]')
        fi

        # Cypress
        if echo "$pkg" | jq -e '.devDependencies.cypress or .dependencies.cypress' > /dev/null 2>&1; then
            frameworks=$(echo "$frameworks" | jq '. + ["cypress"]')
        fi

        # Playwright
        if echo "$pkg" | jq -e '.devDependencies["@playwright/test"] or .dependencies["@playwright/test"]' > /dev/null 2>&1; then
            frameworks=$(echo "$frameworks" | jq '. + ["playwright"]')
        fi

        # Testing Library
        if echo "$pkg" | jq -e '.devDependencies["@testing-library/react"] or .devDependencies["@testing-library/vue"]' > /dev/null 2>&1; then
            frameworks=$(echo "$frameworks" | jq '. + ["testing-library"]')
        fi

        # Jasmine
        if echo "$pkg" | jq -e '.devDependencies.jasmine or .dependencies.jasmine' > /dev/null 2>&1; then
            frameworks=$(echo "$frameworks" | jq '. + ["jasmine"]')
        fi

        # AVA
        if echo "$pkg" | jq -e '.devDependencies.ava or .dependencies.ava' > /dev/null 2>&1; then
            frameworks=$(echo "$frameworks" | jq '. + ["ava"]')
        fi
    fi

    # Check requirements.txt / pyproject.toml for Python
    if [[ -f "$repo_dir/requirements.txt" ]] || [[ -f "$repo_dir/requirements-dev.txt" ]]; then
        local reqs=$(cat "$repo_dir/requirements.txt" "$repo_dir/requirements-dev.txt" 2>/dev/null || echo "")

        if echo "$reqs" | grep -qi "pytest"; then
            frameworks=$(echo "$frameworks" | jq '. + ["pytest"]')
        fi
        if echo "$reqs" | grep -qi "unittest"; then
            frameworks=$(echo "$frameworks" | jq '. + ["unittest"]')
        fi
        if echo "$reqs" | grep -qi "nose"; then
            frameworks=$(echo "$frameworks" | jq '. + ["nose"]')
        fi
    fi

    if [[ -f "$repo_dir/pyproject.toml" ]]; then
        local pyproject=$(cat "$repo_dir/pyproject.toml")
        if echo "$pyproject" | grep -qi "pytest"; then
            frameworks=$(echo "$frameworks" | jq '. + ["pytest"]')
        fi
    fi

    # Check for Go testing (built-in)
    if find "$repo_dir" -name "*_test.go" ! -path "*vendor*" 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["go-testing"]')
    fi

    # Check for Ruby RSpec
    if [[ -f "$repo_dir/Gemfile" ]]; then
        if grep -qi "rspec" "$repo_dir/Gemfile"; then
            frameworks=$(echo "$frameworks" | jq '. + ["rspec"]')
        fi
        if grep -qi "minitest" "$repo_dir/Gemfile"; then
            frameworks=$(echo "$frameworks" | jq '. + ["minitest"]')
        fi
    fi

    # Check for Java JUnit
    if [[ -f "$repo_dir/pom.xml" ]]; then
        if grep -qi "junit" "$repo_dir/pom.xml"; then
            frameworks=$(echo "$frameworks" | jq '. + ["junit"]')
        fi
    fi
    if [[ -f "$repo_dir/build.gradle" ]]; then
        if grep -qi "junit" "$repo_dir/build.gradle"; then
            frameworks=$(echo "$frameworks" | jq '. + ["junit"]')
        fi
    fi

    # Deduplicate
    echo "$frameworks" | jq 'unique'
}

# Detect mocking libraries
detect_mocking_libraries() {
    local repo_dir="$1"
    local mocking="[]"

    if [[ -f "$repo_dir/package.json" ]]; then
        local pkg=$(cat "$repo_dir/package.json")

        # MSW (Mock Service Worker)
        if echo "$pkg" | jq -e '.devDependencies.msw or .dependencies.msw' > /dev/null 2>&1; then
            mocking=$(echo "$mocking" | jq '. + ["msw"]')
        fi

        # Nock
        if echo "$pkg" | jq -e '.devDependencies.nock or .dependencies.nock' > /dev/null 2>&1; then
            mocking=$(echo "$mocking" | jq '. + ["nock"]')
        fi

        # Sinon
        if echo "$pkg" | jq -e '.devDependencies.sinon or .dependencies.sinon' > /dev/null 2>&1; then
            mocking=$(echo "$mocking" | jq '. + ["sinon"]')
        fi

        # Jest mocking (built-in with jest)
        if echo "$pkg" | jq -e '.devDependencies.jest' > /dev/null 2>&1; then
            mocking=$(echo "$mocking" | jq '. + ["jest-mock"]')
        fi

        # Testdouble
        if echo "$pkg" | jq -e '.devDependencies.testdouble' > /dev/null 2>&1; then
            mocking=$(echo "$mocking" | jq '. + ["testdouble"]')
        fi
    fi

    # Python mocking
    if [[ -f "$repo_dir/requirements.txt" ]] || [[ -f "$repo_dir/requirements-dev.txt" ]]; then
        local reqs=$(cat "$repo_dir/requirements.txt" "$repo_dir/requirements-dev.txt" 2>/dev/null || echo "")

        if echo "$reqs" | grep -qi "pytest-mock"; then
            mocking=$(echo "$mocking" | jq '. + ["pytest-mock"]')
        fi
        if echo "$reqs" | grep -qi "responses"; then
            mocking=$(echo "$mocking" | jq '. + ["responses"]')
        fi
        if echo "$reqs" | grep -qi "httpretty"; then
            mocking=$(echo "$mocking" | jq '. + ["httpretty"]')
        fi
        if echo "$reqs" | grep -qi "mock"; then
            mocking=$(echo "$mocking" | jq '. + ["unittest.mock"]')
        fi
    fi

    echo "$mocking" | jq 'unique'
}

# Find test configuration files
find_test_configs() {
    local repo_dir="$1"
    local configs="[]"

    local config_files=(
        "jest.config.js" "jest.config.ts" "jest.config.json"
        "vitest.config.js" "vitest.config.ts"
        "cypress.config.js" "cypress.config.ts"
        "playwright.config.js" "playwright.config.ts"
        "karma.conf.js"
        "pytest.ini" "setup.cfg" "pyproject.toml"
        ".mocharc.js" ".mocharc.json" ".mocharc.yml"
        "ava.config.js"
    )

    for config in "${config_files[@]}"; do
        if [[ -f "$repo_dir/$config" ]]; then
            configs=$(echo "$configs" | jq --arg file "$config" '. + [$file]')
        fi
    done

    # Check for test directories
    local test_dirs=("test" "tests" "__tests__" "spec" "specs" "e2e" "cypress" "playwright")
    for dir in "${test_dirs[@]}"; do
        if [[ -d "$repo_dir/$dir" ]]; then
            configs=$(echo "$configs" | jq --arg dir "$dir/" '. + [$dir]')
        fi
    done

    echo "$configs"
}

# Find directories without tests
find_uncovered_directories() {
    local repo_dir="$1"
    local uncovered="[]"

    # Get source directories (src, lib, app, etc.)
    local source_dirs=$(find "$repo_dir" -maxdepth 2 -type d \( \
        -name "src" -o -name "lib" -o -name "app" -o -name "packages" -o -name "modules" \
    \) ! -path "*node_modules*" ! -path "*vendor*" 2>/dev/null)

    while IFS= read -r src_dir; do
        [[ -z "$src_dir" ]] && continue

        # Get immediate subdirectories
        local subdirs=$(find "$src_dir" -maxdepth 1 -type d ! -path "$src_dir" 2>/dev/null)

        while IFS= read -r subdir; do
            [[ -z "$subdir" ]] && continue

            local dirname=$(basename "$subdir")
            local rel_path="${subdir#$repo_dir/}"

            # Skip common non-source directories
            [[ "$dirname" =~ ^(node_modules|vendor|dist|build|__pycache__|\.git|test|tests|__tests__|spec)$ ]] && continue

            # Check if there are any test files for this directory
            local has_tests=false

            # Check for test files matching directory name
            if find "$repo_dir" -type f \( \
                -name "*${dirname}*.test.*" -o -name "*${dirname}*.spec.*" -o \
                -name "test_${dirname}*" -o -name "*${dirname}_test*" \
            \) ! -path "*node_modules*" 2>/dev/null | head -1 | grep -q .; then
                has_tests=true
            fi

            # Check for tests within the directory
            if find "$subdir" -type f \( \
                -name "*.test.*" -o -name "*.spec.*" -o -name "test_*" -o -name "*_test.*" \
            \) 2>/dev/null | head -1 | grep -q .; then
                has_tests=true
            fi

            if [[ "$has_tests" == "false" ]]; then
                # Count source files in directory
                local file_count=$(find "$subdir" -type f \( \
                    -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.java" -o -name "*.go" \
                \) ! -name "*.test.*" ! -name "*.spec.*" 2>/dev/null | wc -l | tr -d ' ')

                if [[ "$file_count" -gt 0 ]]; then
                    uncovered=$(echo "$uncovered" | jq \
                        --arg dir "$rel_path" \
                        --argjson files "$file_count" \
                        '. + [{"directory": $dir, "source_files": $files}]')
                fi
            fi
        done <<< "$subdirs"
    done <<< "$source_dirs"

    echo "$uncovered"
}

# Estimate coverage level
estimate_coverage() {
    local test_count="$1"
    local source_count="$2"

    if [[ "$source_count" -eq 0 ]]; then
        echo "none"
        return
    fi

    local ratio=$(echo "scale=2; $test_count / $source_count" | bc 2>/dev/null || echo "0")

    if (( $(echo "$ratio >= 0.8" | bc -l 2>/dev/null || echo 0) )); then
        echo "high"
    elif (( $(echo "$ratio >= 0.5" | bc -l 2>/dev/null || echo 0) )); then
        echo "medium"
    elif (( $(echo "$ratio >= 0.2" | bc -l 2>/dev/null || echo 0) )); then
        echo "low"
    else
        echo "minimal"
    fi
}

# Main analysis
analyze_target() {
    local repo_dir="$1"

    echo -e "${BLUE}Counting source files...${NC}" >&2
    local source_files=$(count_source_files "$repo_dir")
    local source_total=$(echo "$source_files" | jq -r '.total')
    echo -e "${GREEN}✓ Found $source_total source files${NC}" >&2

    echo -e "${BLUE}Counting test files...${NC}" >&2
    local test_files=$(count_test_files "$repo_dir")
    local test_total=$(echo "$test_files" | jq -r '.total')
    echo -e "${GREEN}✓ Found $test_total test files${NC}" >&2

    echo -e "${BLUE}Detecting test frameworks...${NC}" >&2
    local frameworks=$(detect_test_frameworks "$repo_dir")
    local framework_count=$(echo "$frameworks" | jq 'length')
    echo -e "${GREEN}✓ Detected $framework_count frameworks${NC}" >&2

    echo -e "${BLUE}Detecting mocking libraries...${NC}" >&2
    local mocking=$(detect_mocking_libraries "$repo_dir")
    echo -e "${GREEN}✓ Mocking libraries detected${NC}" >&2

    echo -e "${BLUE}Finding test configurations...${NC}" >&2
    local configs=$(find_test_configs "$repo_dir")
    echo -e "${GREEN}✓ Configuration files found${NC}" >&2

    echo -e "${BLUE}Finding directories without tests...${NC}" >&2
    local uncovered=$(find_uncovered_directories "$repo_dir")
    local uncovered_count=$(echo "$uncovered" | jq 'length')
    echo -e "${GREEN}✓ Found $uncovered_count directories without tests${NC}" >&2

    # Calculate test-to-code ratio
    local ratio=0
    if [[ "$source_total" -gt 0 ]]; then
        ratio=$(echo "scale=2; $test_total / $source_total" | bc 2>/dev/null || echo "0")
    fi

    # Estimate coverage
    local coverage=$(estimate_coverage "$test_total" "$source_total")

    echo -e "${CYAN}Test-to-code ratio: $ratio ($coverage coverage)${NC}" >&2

    # Determine test types breakdown
    local unit_tests=$(($(echo "$test_files" | jq -r '.javascript_unit + .typescript_unit + .python + .java + .go + .ruby_rspec')))
    local e2e_tests=$(($(echo "$test_files" | jq -r '.e2e + .cypress + .playwright')))
    local integration_tests=$(echo "$test_files" | jq -r '.integration')

    # Build final output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.0.0" \
        --argjson test_total "$test_total" \
        --argjson source_total "$source_total" \
        --arg ratio "$ratio" \
        --arg coverage "$coverage" \
        --argjson unit "$unit_tests" \
        --argjson e2e "$e2e_tests" \
        --argjson integration "$integration_tests" \
        --argjson source_files "$source_files" \
        --argjson test_files "$test_files" \
        --argjson frameworks "$frameworks" \
        --argjson mocking "$mocking" \
        --argjson configs "$configs" \
        --argjson uncovered "$uncovered" \
        '{
            analyzer: "test-coverage",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            summary: {
                test_files: $test_total,
                source_files: $source_total,
                test_to_code_ratio: ($ratio | tonumber),
                estimated_coverage: $coverage,
                unit_tests: $unit,
                e2e_tests: $e2e,
                integration_tests: $integration,
                frameworks_count: ($frameworks | length),
                uncovered_directories: ($uncovered | length)
            },
            source_file_breakdown: $source_files,
            test_file_breakdown: $test_files,
            test_frameworks: $frameworks,
            mocking_libraries: $mocking,
            test_configurations: $configs,
            uncovered_directories: $uncovered
        }'
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage ;;
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        --repo)
            REPO="$2"
            shift 2
            ;;
        --org)
            ORG="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -k|--keep-clone)
            CLEANUP=false
            shift
            ;;
        -*)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Main execution
scan_path=""

if [[ -n "$LOCAL_PATH" ]]; then
    [[ ! -d "$LOCAL_PATH" ]] && { echo '{"error": "Local path does not exist"}'; exit 1; }
    scan_path="$LOCAL_PATH"
    TARGET="$LOCAL_PATH"
elif [[ -n "$REPO" ]]; then
    # Look in phantom/gibson cache
    REPO_ORG=$(echo "$REPO" | cut -d'/' -f1)
    REPO_NAME=$(echo "$REPO" | cut -d'/' -f2)
    PHANTOM_PATH="$HOME/.phantom/projects/$REPO_ORG/$REPO_NAME/repo"
    GIBSON_PATH="$HOME/.gibson/projects/${REPO_ORG}-${REPO_NAME}/repo"

    if [[ -d "$PHANTOM_PATH" ]]; then
        scan_path="$PHANTOM_PATH"
        TARGET="$REPO"
    elif [[ -d "$GIBSON_PATH" ]]; then
        scan_path="$GIBSON_PATH"
        TARGET="$REPO"
    else
        echo '{"error": "Repository not found in cache. Clone it first or use --local-path"}'
        exit 1
    fi
elif [[ -n "$ORG" ]]; then
    # Scan ALL repos in the org
    ORG_PATH="$HOME/.phantom/projects/$ORG"
    if [[ -d "$ORG_PATH" ]]; then
        # Collect repos with and without cloned code
        REPOS_TO_SCAN=()
        REPOS_NOT_CLONED=()
        for repo_dir in "$ORG_PATH"/*/; do
            repo_name=$(basename "$repo_dir")
            if [[ -d "$repo_dir/repo" ]]; then
                REPOS_TO_SCAN+=("$repo_name")
            else
                REPOS_NOT_CLONED+=("$repo_name")
            fi
        done

        # Check if there are uncloned repos and prompt user
        if [[ ${#REPOS_NOT_CLONED[@]} -gt 0 ]]; then
            echo -e "${YELLOW}Found ${#REPOS_NOT_CLONED[@]} repositories without cloned code:${NC}" >&2
            for repo in "${REPOS_NOT_CLONED[@]}"; do
                echo -e "  - $repo" >&2
            done
            echo "" >&2

            read -p "Would you like to hydrate these repos for analysis? [y/N] " -n 1 -r >&2
            echo "" >&2

            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${BLUE}Hydrating ${#REPOS_NOT_CLONED[@]} repositories...${NC}" >&2
                for repo in "${REPOS_NOT_CLONED[@]}"; do
                    echo -e "${CYAN}Cloning $ORG/$repo...${NC}" >&2
                    "$REPO_ROOT/utils/phantom/hydrate.sh" --repo "$ORG/$repo" --quick >&2 2>&1 || true
                    if [[ -d "$ORG_PATH/$repo/repo" ]]; then
                        REPOS_TO_SCAN+=("$repo")
                        echo -e "${GREEN}✓ $repo ready${NC}" >&2
                    else
                        echo -e "${RED}✗ Failed to clone $repo${NC}" >&2
                    fi
                done
                echo "" >&2
            else
                echo -e "${CYAN}Continuing with ${#REPOS_TO_SCAN[@]} already-cloned repositories...${NC}" >&2
            fi
        fi

        if [[ ${#REPOS_TO_SCAN[@]} -eq 0 ]]; then
            echo '{"error": "No repositories with cloned code found in org cache. Hydrate repos first."}'
            exit 1
        fi

        # Analyze each repo and aggregate results
        echo -e "${BLUE}Scanning ${#REPOS_TO_SCAN[@]} repositories in $ORG...${NC}" >&2

        all_results="[]"
        repo_count=0
        total_repos=${#REPOS_TO_SCAN[@]}

        for repo_name in "${REPOS_TO_SCAN[@]}"; do
            ((repo_count++))
            scan_path="$ORG_PATH/$repo_name/repo"
            TARGET="$ORG/$repo_name"

            echo -e "\n${CYAN}[$repo_count/$total_repos] Analyzing: $TARGET${NC}" >&2

            repo_json=$(analyze_target "$scan_path")
            repo_json=$(echo "$repo_json" | jq --arg repo "$TARGET" '. + {repository: $repo}')

            all_results=$(echo "$all_results" | jq --argjson repo "$repo_json" '. + [$repo]')
        done

        # Build aggregated output
        timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

        final_json=$(jq -n \
            --arg ts "$timestamp" \
            --arg org "$ORG" \
            --arg ver "1.0.0" \
            --argjson repo_count "$total_repos" \
            --argjson repositories "$all_results" \
            '{
                analyzer: "test-coverage",
                version: $ver,
                timestamp: $ts,
                organization: $org,
                summary: {
                    repositories_scanned: $repo_count
                },
                repositories: $repositories
            }')

        echo -e "\n${CYAN}=== Organization Summary ===${NC}" >&2
        echo -e "${CYAN}Repos analyzed: $total_repos${NC}" >&2

        # Output
        if [[ -n "$OUTPUT_FILE" ]]; then
            echo "$final_json" > "$OUTPUT_FILE"
            echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
        else
            echo "$final_json"
        fi
        exit 0
    else
        echo '{"error": "Org not found in cache. Hydrate repos first."}'
        exit 1
    fi
elif [[ -n "$TARGET" ]]; then
    if is_git_url "$TARGET"; then
        clone_repository "$TARGET"
        scan_path="$TEMP_DIR"
    elif [[ -d "$TARGET" ]]; then
        scan_path="$TARGET"
    else
        echo '{"error": "Invalid target - must be URL or directory"}'
        exit 1
    fi
else
    echo '{"error": "No target specified"}'
    exit 1
fi

echo -e "${BLUE}Analyzing test coverage: $TARGET${NC}" >&2

final_json=$(analyze_target "$scan_path")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
