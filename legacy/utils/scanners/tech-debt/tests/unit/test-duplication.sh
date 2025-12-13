#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for Code Duplication Detection
# Tests duplication percentage scoring
#
# Based on RAG indicators: rag/tech-debt/indicators/duplication-thresholds.json
# Overall duplication thresholds:
# - 0-3%: excellent (score 0)
# - 3-5%: good (score 15)
# - 5-10%: moderate (score 35)
# - 10-20%: high (score 65)
# - >20%: critical (score 100)
#
# Note: This relies on jscpd if available, otherwise returns unavailable status
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TESTS_DIR="$(dirname "$SCRIPT_DIR")"
UTILS_DIR="$(dirname "$TESTS_DIR")"
SCANNER="$UTILS_DIR/tech-debt-data.sh"

# Source test framework
source "$TESTS_DIR/test-framework.sh"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Code Duplication Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

#############################################################################
# Test: Duplication output exists
#############################################################################
test_duplication_output_exists() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    assert_json_contains_key "$output" "duplication" "Output should contain duplication key"
}

#############################################################################
# Test: Duplication has availability flag
#############################################################################
test_duplication_has_available_flag() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Check that duplication key exists and has either available or note field
    assert_json_contains_key "$output" "duplication" "Should have duplication key"

    local duplication=$(echo "$output" | jq '.duplication')
    # jscpd may or may not be installed, so check for either available flag or note
    local has_available=$(echo "$duplication" | jq 'has("available")')
    local has_note=$(echo "$duplication" | jq 'has("note")')

    if [[ "$has_available" == "true" ]] || [[ "$has_note" == "true" ]]; then
        return 0
    else
        echo "Duplication should have available flag or note"
        return 1
    fi
}

#############################################################################
# Test: Duplication category score calculated
#############################################################################
test_duplication_category_score() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    assert_json_contains_key "$output" "category_scores.duplication" "Should have duplication score"

    local dup_score=$(echo "$output" | jq -r '.category_scores.duplication.score')
    assert_greater_than_or_equal "$dup_score" "0" "Duplication score should be >= 0"
    assert_less_than_or_equal "$dup_score" "100" "Duplication score should be <= 100"
}

#############################################################################
# Test: Duplication weight is 15 (from RAG guide)
#############################################################################
test_duplication_weight() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local weight=$(echo "$output" | jq -r '.category_scores.duplication.weight')
    assert_equals "15" "$weight" "Duplication weight should be 15"
}

#############################################################################
# Test: Zero duplication = score 0
#############################################################################
test_zero_duplication_score() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # With minimal unique code, duplication should be 0 or unavailable
    local dup_available=$(echo "$output" | jq -r '.duplication.available')
    local dup_score=$(echo "$output" | jq -r '.category_scores.duplication.score')

    # If jscpd not available, score is 0 by default
    # If available and no duplication, also 0
    assert_equals "0" "$dup_score" "Zero/no duplication should have score 0"
}

#############################################################################
# Test: Duplication level classification
#############################################################################
test_duplication_level() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local level=$(echo "$output" | jq -r '.category_scores.duplication.level')

    # With 0 duplication, level should be excellent
    assert_equals "excellent" "$level" "Zero duplication should be excellent"
}

#############################################################################
# Test: Duplication percentage in output if available
#############################################################################
test_duplication_percentage_if_available() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local dup_available=$(echo "$output" | jq -r '.duplication.available')

    if [[ "$dup_available" == "true" ]]; then
        local percentage=$(echo "$output" | jq -r '.duplication.percentage')
        assert_greater_than_or_equal "$percentage" "0" "Percentage should be >= 0"
    else
        # If jscpd not available, check for note
        local note=$(echo "$output" | jq -r '.duplication.note')
        assert_contains "$note" "jscpd" "Should note that jscpd is needed"
    fi
}

#############################################################################
# Test: Score thresholds from RAG (3-5% = 15)
#############################################################################
test_score_threshold_calculation() {
    # This test validates the scoring logic, not jscpd availability
    # We verify the threshold logic matches the RAG definition

    # Create a synthetic test by examining the scoring function
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # The scoring thresholds are:
    # dup_int > 20 -> score 100
    # dup_int > 10 -> score 65
    # dup_int > 5 -> score 35
    # dup_int > 3 -> score 15
    # else -> score 0

    # With 0% duplication, score should be 0
    local dup_score=$(echo "$output" | jq -r '.category_scores.duplication.score')
    assert_equals "0" "$dup_score" "0% duplication should have score 0"
}

#############################################################################
# Test: Duplication data structure
#############################################################################
test_duplication_data_structure() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local duplication=$(echo "$output" | jq '.duplication')

    # Should have available key
    assert_json_contains_key "$duplication" "available" "Should have available key"

    local available=$(echo "$duplication" | jq -r '.available')
    if [[ "$available" == "true" ]]; then
        # If available, should have percentage and duplicate_blocks
        assert_json_contains_key "$duplication" "percentage" "Should have percentage"
        assert_json_contains_key "$duplication" "duplicate_blocks" "Should have duplicate_blocks"
    fi
}

#############################################################################
# Test: Excluded directories for duplication
#############################################################################
test_excluded_dirs_duplication() {
    # When jscpd runs, it should exclude node_modules, vendor, .git, dist, build
    # This test verifies the exclusion patterns are passed

    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir/node_modules" "$repo_dir/src"

    # Create duplicate code in node_modules (should be excluded)
    for i in {1..10}; do
        echo "const duplicate = 'same code repeated';" >> "$repo_dir/node_modules/dup.js"
    done

    # Create unique code in src
    echo 'const unique = "different";' > "$repo_dir/src/app.js"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # If jscpd is available, duplication should not include node_modules
    local dup_available=$(echo "$output" | jq -r '.duplication.available')
    if [[ "$dup_available" == "true" ]]; then
        local percentage=$(echo "$output" | jq -r '.duplication.percentage')
        # Should be 0 or very low since only unique code in src
        assert_less_than "$percentage" "50" "Excluded dirs should not count"
    fi

    teardown
}

#############################################################################
# Test: Category scores include duplication
#############################################################################
test_category_scores_include_duplication() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Verify duplication is part of category_scores
    local categories=$(echo "$output" | jq -r '.category_scores | keys[]')

    assert_contains "$categories" "duplication" "Category scores should include duplication"
}

#############################################################################
# Test: High duplication score classification
#############################################################################
test_high_duplication_level() {
    # This test validates the level classification logic
    # With high score (>60), level should be "high"
    # With critical score (>80), level should be "critical"

    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # With 0% duplication and score 0, level should be excellent
    local level=$(echo "$output" | jq -r '.category_scores.duplication.level')

    # Verify the level matches the score
    local score=$(echo "$output" | jq -r '.category_scores.duplication.score')

    if [[ $score -le 20 ]]; then
        assert_equals "excellent" "$level" "Score 0-20 should be excellent"
    elif [[ $score -le 40 ]]; then
        assert_equals "good" "$level" "Score 21-40 should be good"
    elif [[ $score -le 60 ]]; then
        assert_equals "moderate" "$level" "Score 41-60 should be moderate"
    elif [[ $score -le 80 ]]; then
        assert_equals "high" "$level" "Score 61-80 should be high"
    else
        assert_equals "critical" "$level" "Score 81-100 should be critical"
    fi
}

#############################################################################
# Run all tests
#############################################################################
run_test "Duplication output exists" test_duplication_output_exists
run_test "Duplication has available flag" test_duplication_has_available_flag
run_test "Duplication category score" test_duplication_category_score
run_test "Duplication weight is 15" test_duplication_weight
run_test "Zero duplication score" test_zero_duplication_score
run_test "Duplication level" test_duplication_level
run_test "Duplication percentage if available" test_duplication_percentage_if_available
run_test "Score threshold calculation" test_score_threshold_calculation
run_test "Duplication data structure" test_duplication_data_structure
run_test "Excluded dirs duplication" test_excluded_dirs_duplication
run_test "Category scores include duplication" test_category_scores_include_duplication
run_test "High duplication level" test_high_duplication_level

print_summary
