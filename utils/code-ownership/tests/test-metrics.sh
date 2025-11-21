#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for Metrics Library
# Tests all metric calculation functions
#############################################################################

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$SCRIPT_DIR/../lib"

# Load library
source "$LIB_DIR/metrics.sh"

# Test framework
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

assert_equals() {
    local expected="$1"
    local actual="$2"
    local test_name="$3"

    ((TESTS_RUN++))

    # Compare with tolerance for floating point
    if [[ "$expected" == "$actual" ]] || \
       [[ $(echo "scale=2; $expected - $actual" | bc -l | sed 's/-//') == "0" ]] || \
       [[ $(echo "scale=2; $expected - $actual" | bc -l | sed 's/-//' | cut -c1-4) == "0.00" ]]; then
        echo "✓ PASS: $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        echo "✗ FAIL: $test_name"
        echo "  Expected: $expected"
        echo "  Actual:   $actual"
        ((TESTS_FAILED++))
        return 1
    fi
}

assert_in_range() {
    local min="$1"
    local max="$2"
    local actual="$3"
    local test_name="$4"

    ((TESTS_RUN++))

    if (( $(echo "$actual >= $min && $actual <= $max" | bc -l) )); then
        echo "✓ PASS: $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        echo "✗ FAIL: $test_name"
        echo "  Expected: $min - $max"
        echo "  Actual:   $actual"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test: calculate_recency_factor
test_recency_factor() {
    echo ""
    echo "Testing calculate_recency_factor..."

    # Test immediate recency (0 days)
    local result=$(calculate_recency_factor 0 90)
    assert_equals "1.0000" "$result" "Recency at day 0 should be 1.0"

    # Test half-life (90 days)
    result=$(calculate_recency_factor 90 90)
    assert_in_range "0.49" "0.51" "$result" "Recency at half-life should be ~0.5"

    # Test double half-life (180 days)
    result=$(calculate_recency_factor 180 90)
    assert_in_range "0.24" "0.26" "$result" "Recency at 2x half-life should be ~0.25"
}

# Test: calculate_gini_coefficient
test_gini_coefficient() {
    echo ""
    echo "Testing calculate_gini_coefficient..."

    # Test perfect equality
    local result=$(calculate_gini_coefficient 10 10 10 10 10)
    assert_equals "0" "$result" "Gini for equal distribution should be 0"

    # Test perfect inequality
    result=$(calculate_gini_coefficient 100 0 0 0 0)
    assert_in_range "0.79" "0.81" "$result" "Gini for complete inequality should be ~0.8"

    # Test moderate inequality
    result=$(calculate_gini_coefficient 50 30 15 5 0)
    assert_in_range "0.45" "0.55" "$result" "Gini for moderate inequality should be ~0.5"
}

# Test: calculate_bus_factor
test_bus_factor() {
    echo ""
    echo "Testing calculate_bus_factor..."

    # Test with single owner (bus factor = 1)
    local result=$(calculate_bus_factor 100 100)
    assert_equals "1" "$result" "Bus factor for single owner should be 1"

    # Test with well-distributed ownership
    result=$(calculate_bus_factor 100 25 25 25 25)
    assert_equals "1" "$result" "Bus factor for equal distribution should be 1"

    # Test with concentrated ownership
    result=$(calculate_bus_factor 100 30 30 20 20)
    assert_in_range "1" "2" "$result" "Bus factor for 30/30/20/20 should be 1-2"
}

# Test: calculate_health_score
test_health_score() {
    echo ""
    echo "Testing calculate_health_score..."

    # Test excellent health
    local result=$(calculate_health_score 95 0.2 90 85)
    assert_in_range "85" "95" "$result" "Health score for excellent metrics should be 85-95"

    # Test poor health
    result=$(calculate_health_score 40 0.8 30 25)
    assert_in_range "25" "40" "$result" "Health score for poor metrics should be 25-40"

    # Test moderate health
    result=$(calculate_health_score 70 0.5 60 55)
    assert_in_range "55" "70" "$result" "Health score for moderate metrics should be 55-70"
}

# Test: get_health_grade
test_health_grade() {
    echo ""
    echo "Testing get_health_grade..."

    local result=$(get_health_grade 90)
    assert_equals "Excellent" "$result" "Grade for 90 should be Excellent"

    result=$(get_health_grade 75)
    assert_equals "Good" "$result" "Grade for 75 should be Good"

    result=$(get_health_grade 55)
    assert_equals "Fair" "$result" "Grade for 55 should be Fair"

    result=$(get_health_grade 35)
    assert_equals "Poor" "$result" "Grade for 35 should be Poor"
}

# Test: calculate_top_n_concentration
test_top_n_concentration() {
    echo ""
    echo "Testing calculate_top_n_concentration..."

    # Test top-1 concentration
    local result=$(calculate_top_n_concentration 100 1 50 30 20)
    assert_equals "50" "$result" "Top-1 concentration should be 50%"

    # Test top-2 concentration
    result=$(calculate_top_n_concentration 100 2 50 30 20)
    assert_equals "80" "$result" "Top-2 concentration should be 80%"

    # Test top-3 concentration
    result=$(calculate_top_n_concentration 100 3 50 30 20)
    assert_equals "100" "$result" "Top-3 concentration should be 100%"
}

# Test: get_staleness_category
test_staleness_category() {
    echo ""
    echo "Testing get_staleness_category..."

    local result=$(get_staleness_category 15)
    assert_equals "Active" "$result" "15 days should be Active"

    result=$(get_staleness_category 45)
    assert_equals "Recent" "$result" "45 days should be Recent"

    result=$(get_staleness_category 75)
    assert_equals "Stale" "$result" "75 days should be Stale"

    result=$(get_staleness_category 120)
    assert_equals "Inactive" "$result" "120 days should be Inactive"

    result=$(get_staleness_category 200)
    assert_equals "Abandoned" "$result" "200 days should be Abandoned"
}

# Test: calculate_commit_frequency_score
test_commit_frequency_score() {
    echo ""
    echo "Testing calculate_commit_frequency_score..."

    # Test full score
    local result=$(calculate_commit_frequency_score 100 100)
    assert_equals "100" "$result" "100/100 commits should be 100"

    # Test half score
    result=$(calculate_commit_frequency_score 50 100)
    assert_equals "50" "$result" "50/100 commits should be 50"

    # Test zero score
    result=$(calculate_commit_frequency_score 0 100)
    assert_equals "0" "$result" "0/100 commits should be 0"
}

# Test: calculate_lines_score
test_lines_score() {
    echo ""
    echo "Testing calculate_lines_score..."

    # Test full score
    local result=$(calculate_lines_score 1000 1000)
    assert_equals "100" "$result" "1000/1000 lines should be 100"

    # Test partial score
    result=$(calculate_lines_score 250 1000)
    assert_equals "25" "$result" "250/1000 lines should be 25"

    # Test capping at 100
    result=$(calculate_lines_score 2000 1000)
    assert_equals "100" "$result" "2000/1000 lines should be capped at 100"
}

# Test: calculate_review_participation_score
test_review_participation_score() {
    echo ""
    echo "Testing calculate_review_participation_score..."

    # Test full score
    local result=$(calculate_review_participation_score 100 100 100 100)
    assert_equals "100" "$result" "Max reviews should be 100"

    # Test half score
    result=$(calculate_review_participation_score 50 50 100 100)
    assert_equals "50" "$result" "Half reviews should be 50"

    # Test zero score
    result=$(calculate_review_participation_score 0 0 100 100)
    assert_equals "0" "$result" "No reviews should be 0"
}

# Test: calculate_ownership_score
test_ownership_score() {
    echo ""
    echo "Testing calculate_ownership_score..."

    # Test maximum score
    local result=$(calculate_ownership_score 100 100 100 1.0 1.0)
    assert_equals "100" "$result" "Maximum inputs should give 100"

    # Test minimum score
    result=$(calculate_ownership_score 0 0 0 0.0 0.0)
    assert_equals "0" "$result" "Minimum inputs should give 0"

    # Test moderate score
    result=$(calculate_ownership_score 50 50 50 0.5 0.5)
    assert_in_range "45" "55" "$result" "Moderate inputs should give ~50"
}

# Run all tests
main() {
    echo "========================================="
    echo "Metrics Library Unit Tests"
    echo "========================================="

    test_recency_factor
    test_gini_coefficient
    test_bus_factor
    test_health_score
    test_health_grade
    test_top_n_concentration
    test_staleness_category
    test_commit_frequency_score
    test_lines_score
    test_review_participation_score
    test_ownership_score

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
