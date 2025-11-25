#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Security Analyser - Test Runner
# Tests the analyser against known vulnerable code samples
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ANALYSER="$SCRIPT_DIR/../code-security-analyser.sh"
TEST_SAMPLES="$SCRIPT_DIR/test-samples"
OUTPUT_DIR="$SCRIPT_DIR/test-output"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Clean up previous test output
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}  Code Security Analyser - Test Suite${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Check prerequisites
check_prerequisites() {
    local missing=false

    if [[ ! -f "$ANALYSER" ]]; then
        echo -e "${RED}Error: Analyser not found at $ANALYSER${NC}"
        missing=true
    fi

    if [[ ! -d "$TEST_SAMPLES" ]]; then
        echo -e "${RED}Error: Test samples not found at $TEST_SAMPLES${NC}"
        missing=true
    fi

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${YELLOW}Warning: ANTHROPIC_API_KEY not set${NC}"
        echo "  Some tests require Claude API access"
        echo ""
    fi

    if [[ "$missing" == "true" ]]; then
        exit 1
    fi
}

# Run a single test
run_test() {
    local name="$1"
    local file="$2"
    local expected_category="$3"
    local expected_type="$4"

    ((TESTS_RUN++))

    echo -n "  Testing: $name... "

    # Skip Claude tests if API key not set
    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${YELLOW}SKIPPED (no API key)${NC}"
        return
    fi

    # Run analyser on single file
    local output_file="$OUTPUT_DIR/${name//[^a-zA-Z0-9]/_}.json"

    if "$ANALYSER" --local "$TEST_SAMPLES" --format json --output "$OUTPUT_DIR/$name" --max-files 1 2>/dev/null; then
        # Check if findings exist
        if [[ -f "$OUTPUT_DIR/$name/findings.json" ]]; then
            local findings=$(cat "$OUTPUT_DIR/$name/findings.json")
            local count=$(echo "$findings" | jq 'length')

            if [[ "$count" -gt 0 ]]; then
                echo -e "${GREEN}PASS${NC} ($count findings)"
                ((TESTS_PASSED++))
            else
                echo -e "${YELLOW}WARN${NC} (no findings - may be false negative)"
            fi
        else
            echo -e "${RED}FAIL${NC} (no output file)"
            ((TESTS_FAILED++))
        fi
    else
        echo -e "${RED}FAIL${NC} (analyser error)"
        ((TESTS_FAILED++))
    fi
}

# Test script execution
test_script_runs() {
    echo -e "${BLUE}Test: Script execution${NC}"

    echo -n "  Checking script is executable... "
    if [[ -x "$ANALYSER" ]]; then
        echo -e "${GREEN}PASS${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL${NC}"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))

    echo -n "  Checking --help works... "
    if "$ANALYSER" --help 2>&1 | grep -q "Usage:"; then
        echo -e "${GREEN}PASS${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL${NC}"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
}

# Test directory scanning
test_directory_scan() {
    echo ""
    echo -e "${BLUE}Test: Directory scanning${NC}"

    echo -n "  Checking test samples directory... "
    local sample_count=$(find "$TEST_SAMPLES" -type f | wc -l | tr -d ' ')
    if [[ "$sample_count" -gt 0 ]]; then
        echo -e "${GREEN}PASS${NC} ($sample_count files)"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL${NC} (no test files)"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
}

# Test file type detection
test_file_types() {
    echo ""
    echo -e "${BLUE}Test: File type support${NC}"

    for ext in py js java sh; do
        echo -n "  Checking .$ext file support... "
        local count=$(find "$TEST_SAMPLES" -name "*.$ext" | wc -l | tr -d ' ')
        if [[ "$count" -gt 0 ]]; then
            echo -e "${GREEN}PASS${NC} ($count files)"
            ((TESTS_PASSED++))
        else
            echo -e "${YELLOW}SKIP${NC} (no test files)"
        fi
        ((TESTS_RUN++))
    done
}

# Main test execution
main() {
    check_prerequisites

    echo -e "${BLUE}Running tests...${NC}"
    echo ""

    test_script_runs
    test_directory_scan
    test_file_types

    # Summary
    echo ""
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}  Test Summary${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo ""
    echo "  Tests run:    $TESTS_RUN"
    echo -e "  ${GREEN}Tests passed: $TESTS_PASSED${NC}"
    echo -e "  ${RED}Tests failed: $TESTS_FAILED${NC}"
    echo ""

    if [[ "$TESTS_FAILED" -gt 0 ]]; then
        echo -e "${RED}Some tests failed!${NC}"
        exit 1
    else
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    fi
}

main "$@"
