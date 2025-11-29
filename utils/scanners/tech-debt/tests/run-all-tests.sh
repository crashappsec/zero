#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Master Test Runner for Tech Debt Scanner
# Runs all unit tests for tech-debt indicator validation
#############################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Test results
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0
FAILED_SUITE_NAMES=()

# Function to run a test suite
run_test_suite() {
    local suite_name="$1"
    local suite_script="$2"

    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Running: $suite_name${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    TOTAL_SUITES=$((TOTAL_SUITES + 1))

    if bash "$suite_script"; then
        PASSED_SUITES=$((PASSED_SUITES + 1))
        echo -e "${GREEN}✓ $suite_name PASSED${NC}"
    else
        FAILED_SUITES=$((FAILED_SUITES + 1))
        FAILED_SUITE_NAMES+=("$suite_name")
        echo -e "${RED}✗ $suite_name FAILED${NC}"
    fi
}

# Print final summary
print_final_summary() {
    echo ""
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  FINAL TEST SUMMARY${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Total Test Suites:  $TOTAL_SUITES"
    echo -e "${GREEN}Passed:             $PASSED_SUITES${NC}"

    if [[ $FAILED_SUITES -gt 0 ]]; then
        echo -e "${RED}Failed:             $FAILED_SUITES${NC}"
        echo ""
        echo "Failed Test Suites:"
        for suite in "${FAILED_SUITE_NAMES[@]}"; do
            echo -e "${RED}  ✗ $suite${NC}"
        done
    else
        echo -e "${GREEN}Failed:             0${NC}"
    fi

    echo ""

    if [[ $FAILED_SUITES -eq 0 ]]; then
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}  ALL TEST SUITES PASSED!${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        return 0
    else
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}  SOME TEST SUITES FAILED${NC}"
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        return 1
    fi
}

# Main execution
main() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  Tech Debt Scanner Test Suite            ║${NC}"
    echo -e "${BLUE}║  Testing RAG-based Indicator Detection   ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════╝${NC}"
    echo ""

    # Check prerequisites
    echo "Checking prerequisites..."

    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is required but not installed${NC}"
        echo "Install with: brew install jq"
        exit 1
    fi
    echo -e "${GREEN}✓ jq installed${NC}"

    if ! command -v bc &> /dev/null; then
        echo -e "${YELLOW}Warning: bc not installed - some calculations may be limited${NC}"
    else
        echo -e "${GREEN}✓ bc installed${NC}"
    fi

    if command -v jscpd &> /dev/null; then
        echo -e "${GREEN}✓ jscpd installed (duplication detection enabled)${NC}"
    else
        echo -e "${YELLOW}○ jscpd not installed (duplication tests will use fallback)${NC}"
        echo "  Install with: npm i -g jscpd"
    fi

    if command -v cloc &> /dev/null; then
        echo -e "${GREEN}✓ cloc installed (accurate line counting)${NC}"
    else
        echo -e "${YELLOW}○ cloc not installed (using basic line counting)${NC}"
        echo "  Install with: brew install cloc"
    fi

    echo ""

    # Make test scripts executable
    chmod +x "$SCRIPT_DIR"/unit/*.sh 2>/dev/null || true

    # Run unit tests for each RAG indicator
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}  UNIT TESTS - RAG Indicators${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"

    # Test code markers (TODO, FIXME, HACK, XXX, etc.)
    # RAG: rag/tech-debt/indicators/code-markers.json
    run_test_suite "Code Markers (TODO/FIXME/HACK/XXX)" "$SCRIPT_DIR/unit/test-code-markers.sh"

    # Test deprecated code detection
    # RAG: rag/tech-debt/indicators/code-markers.json (DEPRECATED marker)
    run_test_suite "Deprecated Code Detection" "$SCRIPT_DIR/unit/test-deprecated.sh"

    # Test file size thresholds
    # RAG: rag/tech-debt/indicators/file-size-thresholds.json
    run_test_suite "File Size Thresholds" "$SCRIPT_DIR/unit/test-file-size.sh"

    # Test test coverage (test-to-code ratio)
    # RAG: rag/tech-debt/indicators/test-coverage-thresholds.json
    run_test_suite "Test Coverage Ratio" "$SCRIPT_DIR/unit/test-test-coverage.sh"

    # Test code duplication
    # RAG: rag/tech-debt/indicators/duplication-thresholds.json
    run_test_suite "Code Duplication" "$SCRIPT_DIR/unit/test-duplication.sh"

    # Test overall scoring calculation
    # RAG: rag/tech-debt/scoring/tech-debt-scoring-guide.md
    run_test_suite "Debt Score Calculation" "$SCRIPT_DIR/unit/test-scoring.sh"

    # Print final summary
    print_final_summary
}

# Run main if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main
    exit $?
fi
