#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Test Runner
# Runs all unit and integration tests
#############################################################################

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Track results
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0

run_test_suite() {
    local test_file="$1"
    local test_name=$(basename "$test_file" .sh)

    ((TOTAL_SUITES++))

    echo -e "${BLUE}Running $test_name...${NC}"
    echo ""

    if "$test_file"; then
        echo -e "${GREEN}✓ $test_name passed${NC}"
        ((PASSED_SUITES++))
        return 0
    else
        echo -e "${RED}✗ $test_name failed${NC}"
        ((FAILED_SUITES++))
        return 1
    fi
}

main() {
    echo "========================================="
    echo "Code Ownership Analyser - Test Suite"
    echo "========================================="
    echo ""

    # Run unit tests
    echo -e "${YELLOW}Unit Tests:${NC}"
    echo ""

    run_test_suite "$SCRIPT_DIR/test-metrics.sh"
    echo ""

    run_test_suite "$SCRIPT_DIR/test-config.sh"
    echo ""

    # Run integration tests
    echo -e "${YELLOW}Integration Tests:${NC}"
    echo ""

    run_test_suite "$SCRIPT_DIR/test-integration.sh"
    echo ""

    # Summary
    echo "========================================="
    echo "Final Results:"
    echo "  Total Suites:  $TOTAL_SUITES"
    echo "  Passed Suites: $PASSED_SUITES"
    echo "  Failed Suites: $FAILED_SUITES"
    echo "========================================="

    if [[ $FAILED_SUITES -eq 0 ]]; then
        echo -e "${GREEN}✓ All test suites passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some test suites failed${NC}"
        exit 1
    fi
}

main "$@"
