#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Integration Tests
# Tests full analyser workflow with a test repository
#############################################################################

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ANALYSER_DIR="$SCRIPT_DIR/.."

# Test framework
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

assert_success() {
    local command="$1"
    local test_name="$2"

    ((TESTS_RUN++))

    if eval "$command" &>/dev/null; then
        echo "✓ PASS: $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        echo "✗ FAIL: $test_name"
        ((TESTS_FAILED++))
        return 1
    fi
}

assert_file_exists() {
    local file="$1"
    local test_name="$2"

    ((TESTS_RUN++))

    if [[ -f "$file" ]]; then
        echo "✓ PASS: $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        echo "✗ FAIL: $test_name"
        echo "  File not found: $file"
        ((TESTS_FAILED++))
        return 1
    fi
}

assert_json_valid() {
    local file="$1"
    local test_name="$2"

    ((TESTS_RUN++))

    if jq empty "$file" 2>/dev/null; then
        echo "✓ PASS: $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        echo "✗ FAIL: $test_name"
        echo "  Invalid JSON in file: $file"
        ((TESTS_FAILED++))
        return 1
    fi
}

assert_json_has_key() {
    local file="$1"
    local key="$2"
    local test_name="$3"

    ((TESTS_RUN++))

    if jq -e ".$key" "$file" &>/dev/null; then
        echo "✓ PASS: $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        echo "✗ FAIL: $test_name"
        echo "  Key '$key' not found in JSON"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Setup test repository
setup_test_repo() {
    local test_repo="$1"

    echo "Setting up test repository..."

    # Create test repo
    mkdir -p "$test_repo"
    cd "$test_repo" || return 1

    # Initialize git
    git init -q

    # Configure git user
    git config user.name "Test User"
    git config user.email "test@example.com"

    # Create some test files
    echo "console.log('Hello');" > main.js
    echo "function test() {}" > utils.js
    echo "# Test" > README.md

    # Make initial commits
    git add main.js
    git commit -q -m "Add main.js"

    git add utils.js
    git commit -q -m "Add utils.js"

    git add README.md
    git commit -q -m "Add README"

    # Add second contributor
    git config user.name "Second User"
    git config user.email "second@example.com"

    echo "console.log('Updated');" >> main.js
    git add main.js
    git commit -q -m "Update main.js"

    # Create CODEOWNERS file
    mkdir -p .github
    cat > .github/CODEOWNERS << EOF
# CODEOWNERS file
*.js @test-user
*.md @second-user
* @test-user @second-user
EOF

    git add .github/CODEOWNERS
    git commit -q -m "Add CODEOWNERS"

    echo "Test repository created at: $test_repo"
}

# Test: Basic analyser execution
test_basic_analysis() {
    echo ""
    echo "Testing basic analysis..."

    local test_repo=$(mktemp -d)
    setup_test_repo "$test_repo"

    local output=$(mktemp)

    # Run analyser
    assert_success "$ANALYSER_DIR/ownership-analyser-v2.sh -f json -o '$output' '$test_repo'" \
        "Analyser should run successfully"

    # Check output file exists
    assert_file_exists "$output" "Output file should be created"

    # Check JSON is valid
    assert_json_valid "$output" "Output should be valid JSON"

    # Check required keys exist
    assert_json_has_key "$output" "metadata" "JSON should have metadata"
    assert_json_has_key "$output" "ownership_health" "JSON should have ownership_health"
    assert_json_has_key "$output" "contributors" "JSON should have contributors"

    # Cleanup
    rm -rf "$test_repo" "$output"
}

# Test: CODEOWNERS validation
test_codeowners_validation() {
    echo ""
    echo "Testing CODEOWNERS validation..."

    local test_repo=$(mktemp -d)
    setup_test_repo "$test_repo"

    local output=$(mktemp)

    # Run analyser with validation
    assert_success "$ANALYSER_DIR/ownership-analyser-v2.sh -f json --validate -o '$output' '$test_repo'" \
        "Analyser with validation should run successfully"

    # Cleanup
    rm -rf "$test_repo" "$output"
}

# Test: Different analysis methods
test_analysis_methods() {
    echo ""
    echo "Testing different analysis methods..."

    local test_repo=$(mktemp -d)
    setup_test_repo "$test_repo"

    # Test commit-based
    local output_commit=$(mktemp)
    assert_success "$ANALYSER_DIR/ownership-analyser-v2.sh -m commit -f json -o '$output_commit' '$test_repo'" \
        "Commit-based analysis should work"
    assert_json_valid "$output_commit" "Commit-based output should be valid JSON"

    # Test line-based
    local output_line=$(mktemp)
    assert_success "$ANALYSER_DIR/ownership-analyser-v2.sh -m line -f json -o '$output_line' '$test_repo'" \
        "Line-based analysis should work"
    assert_json_valid "$output_line" "Line-based output should be valid JSON"

    # Test hybrid
    local output_hybrid=$(mktemp)
    assert_success "$ANALYSER_DIR/ownership-analyser-v2.sh -m hybrid -f json -o '$output_hybrid' '$test_repo'" \
        "Hybrid analysis should work"
    assert_json_valid "$output_hybrid" "Hybrid output should be valid JSON"

    # Cleanup
    rm -rf "$test_repo" "$output_commit" "$output_line" "$output_hybrid"
}

# Test: Text output format
test_text_output() {
    echo ""
    echo "Testing text output format..."

    local test_repo=$(mktemp -d)
    setup_test_repo "$test_repo"

    local output=$(mktemp)

    # Run analyser with text output
    assert_success "$ANALYSER_DIR/ownership-analyser-v2.sh -f text -o '$output' '$test_repo'" \
        "Text format analysis should work"

    # Check output file exists
    assert_file_exists "$output" "Text output file should be created"

    # Check for expected content
    ((TESTS_RUN++))
    if grep -q "Code Ownership Analysis" "$output"; then
        echo "✓ PASS: Text output should contain header"
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL: Text output should contain header"
        ((TESTS_FAILED++))
    fi

    # Cleanup
    rm -rf "$test_repo" "$output"
}

# Test: Configuration system integration
test_config_integration() {
    echo ""
    echo "Testing configuration system integration..."

    local test_repo=$(mktemp -d)
    setup_test_repo "$test_repo"

    # Create local config file
    cat > "$test_repo/.code-ownership.conf" << EOF
analysis_method=commit
analysis_days=60
output_format=json
EOF

    local output=$(mktemp)

    # Run analyser (should use local config)
    assert_success "$ANALYSER_DIR/ownership-analyser-v2.sh -o '$output' '$test_repo'" \
        "Analyser should respect local config"

    # Cleanup
    rm -rf "$test_repo" "$output"
}

# Test: Library loading
test_library_loading() {
    echo ""
    echo "Testing library loading..."

    # Source libraries and check for key functions
    ((TESTS_RUN++))
    if source "$ANALYSER_DIR/lib/metrics.sh" 2>/dev/null && \
       type calculate_gini_coefficient &>/dev/null; then
        echo "✓ PASS: metrics.sh should load correctly"
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL: metrics.sh should load correctly"
        ((TESTS_FAILED++))
    fi

    ((TESTS_RUN++))
    if source "$ANALYSER_DIR/lib/config.sh" 2>/dev/null && \
       type init_config &>/dev/null; then
        echo "✓ PASS: config.sh should load correctly"
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL: config.sh should load correctly"
        ((TESTS_FAILED++))
    fi

    ((TESTS_RUN++))
    if source "$ANALYSER_DIR/lib/github.sh" 2>/dev/null && \
       type get_github_profile &>/dev/null; then
        echo "✓ PASS: github.sh should load correctly"
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL: github.sh should load correctly"
        ((TESTS_FAILED++))
    fi

    ((TESTS_RUN++))
    if source "$ANALYSER_DIR/lib/succession.sh" 2>/dev/null && \
       type identify_successors &>/dev/null; then
        echo "✓ PASS: succession.sh should load correctly"
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL: succession.sh should load correctly"
        ((TESTS_FAILED++))
    fi
}

# Run all tests
main() {
    echo "========================================="
    echo "Integration Tests"
    echo "========================================="

    # Check prerequisites
    if ! command -v git &> /dev/null; then
        echo "Error: git is required for integration tests"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        echo "Error: jq is required for integration tests"
        exit 1
    fi

    test_library_loading
    test_basic_analysis
    test_codeowners_validation
    test_analysis_methods
    test_text_output
    test_config_integration

    echo ""
    echo "========================================="
    echo "Test Results:"
    echo "  Total:  $TESTS_RUN"
    echo "  Passed: $TESTS_PASSED"
    echo "  Failed: $TESTS_FAILED"
    echo "========================================="

    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo "✓ All tests passed!"
        exit 0
    else
        echo "✗ Some tests failed"
        exit 1
    fi
}

main "$@"
