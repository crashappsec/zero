#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Master Test Runner for Technology Identification
# Runs all unit and integration tests
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
    echo -e "${BLUE}║  Technology Identification Test Suite   ║${NC}"
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

    if ! command -v syft &> /dev/null; then
        echo -e "${YELLOW}Warning: syft not installed - some integration tests will be skipped${NC}"
        echo "Install with: brew install syft"
    else
        echo -e "${GREEN}✓ syft installed${NC}"
    fi

    if ! command -v osv-scanner &> /dev/null; then
        echo -e "${YELLOW}Warning: osv-scanner not installed - Layer 1b tests will be skipped${NC}"
        echo "Install with: go install github.com/google/osv-scanner/cmd/osv-scanner@latest"
    else
        echo -e "${GREEN}✓ osv-scanner installed${NC}"
    fi

    echo ""

    # Make test scripts executable
    chmod +x "$SCRIPT_DIR"/unit/*.sh 2>/dev/null || true
    chmod +x "$SCRIPT_DIR"/integration/*.sh 2>/dev/null || true

    # Run unit tests
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}  UNIT TESTS${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"

    run_test_suite "SBOM Scanning" "$SCRIPT_DIR/unit/test-sbom-scanning.sh"
    run_test_suite "Confidence Scoring" "$SCRIPT_DIR/unit/test-confidence-scoring.sh"

    # Run integration tests
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}  INTEGRATION TESTS${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"

    run_test_suite "Full Workflow" "$SCRIPT_DIR/integration/test-full-workflow.sh"

    # Print final summary
    print_final_summary
}

# Run main if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main
    exit $?
fi
