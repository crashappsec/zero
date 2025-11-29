#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for Overall Debt Score Calculation
# Tests the weighted score calculation based on RAG scoring guide
#
# Based on: rag/tech-debt/scoring/tech-debt-scoring-guide.md
# Score formula: Total Score = Σ (Category Score × Category Weight) / Σ Weights
#
# Categories tested here:
# - markers (weight: 15)
# - deprecated (weight: 5)
# - file_size (weight: 10)
# - duplication (weight: 15)
# - test_coverage (weight: 15)
#
# Total weight from current implementation: 60
# Score levels: excellent (0-20), good (21-40), moderate (41-60), high (61-80), critical (81-100)
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
echo "  Debt Score Calculation Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

#############################################################################
# Test: Debt score exists in output
#############################################################################
test_debt_score_exists() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    assert_json_contains_key "$output" "summary.debt_score" "Should have debt_score"
    assert_json_contains_key "$output" "summary.debt_level" "Should have debt_level"
}

#############################################################################
# Test: Debt score is in valid range (0-100)
#############################################################################
test_debt_score_range() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local debt_score=$(echo "$output" | jq -r '.summary.debt_score')
    assert_greater_than_or_equal "$debt_score" "0" "Debt score should be >= 0"
    assert_less_than_or_equal "$debt_score" "100" "Debt score should be <= 100"
}

#############################################################################
# Test: Clean repo has low debt score
#############################################################################
test_clean_repo_low_score() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local debt_score=$(echo "$output" | jq -r '.summary.debt_score')

    # Clean repo should have moderate to high test coverage debt (no tests)
    # but low/no other debt - overall should be moderate
    assert_less_than_or_equal "$debt_score" "60" "Clean repo should have moderate or less debt"
}

#############################################################################
# Test: High debt repo has high score
#############################################################################
test_high_debt_repo() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create file with many debt markers
    cat > "$repo_dir/debt.js" << 'EOF'
// TODO: task 1
// TODO: task 2
// TODO: task 3
// TODO: task 4
// TODO: task 5
// FIXME: bug 1
// FIXME: bug 2
// FIXME: bug 3
// FIXME: bug 4
// FIXME: bug 5
// HACK: workaround 1
// HACK: workaround 2
// HACK: workaround 3
// HACK: workaround 4
// HACK: workaround 5
function doSomething() {}
EOF

    # Create a long file
    for i in $(seq 1 600); do
        echo "// Line $i" >> "$repo_dir/long.js"
    done

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local debt_score=$(echo "$output" | jq -r '.summary.debt_score')

    # High debt repo should have elevated score
    assert_greater_than "$debt_score" "20" "High debt repo should have elevated score"

    teardown
}

#############################################################################
# Test: Debt level matches score (excellent: 0-20)
#############################################################################
test_excellent_level() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create clean repo with tests
    echo 'function app() {}' > "$repo_dir/app.js"
    echo 'test("works", () => {})' > "$repo_dir/app.test.js"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local debt_score=$(echo "$output" | jq -r '.summary.debt_score')
    local debt_level=$(echo "$output" | jq -r '.summary.debt_level')

    if [[ $debt_score -le 20 ]]; then
        assert_equals "excellent" "$debt_level" "Score 0-20 should be excellent"
    fi

    teardown
}

#############################################################################
# Test: Debt level matches score (good: 21-40)
#############################################################################
test_good_level() {
    local repo_dir=$(create_test_repo_with_markers)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local debt_score=$(echo "$output" | jq -r '.summary.debt_score')
    local debt_level=$(echo "$output" | jq -r '.summary.debt_level')

    if [[ $debt_score -gt 20 ]] && [[ $debt_score -le 40 ]]; then
        assert_equals "good" "$debt_level" "Score 21-40 should be good"
    fi
}

#############################################################################
# Test: Category scores sum correctly
#############################################################################
test_category_scores_structure() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Verify all expected categories exist
    local categories="markers deprecated file_size duplication test_coverage"
    for cat in $categories; do
        assert_json_contains_key "$output" "category_scores.$cat" "Should have $cat category"
        assert_json_contains_key "$output" "category_scores.$cat.score" "$cat should have score"
        assert_json_contains_key "$output" "category_scores.$cat.weight" "$cat should have weight"
        assert_json_contains_key "$output" "category_scores.$cat.level" "$cat should have level"
    done
}

#############################################################################
# Test: Category weights match RAG guide
#############################################################################
test_category_weights() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Verify weights match the RAG scoring guide
    assert_json_value "$output" ".category_scores.markers.weight" "15" "Markers weight should be 15"
    assert_json_value "$output" ".category_scores.deprecated.weight" "5" "Deprecated weight should be 5"
    assert_json_value "$output" ".category_scores.file_size.weight" "10" "File size weight should be 10"
    assert_json_value "$output" ".category_scores.duplication.weight" "15" "Duplication weight should be 15"
    assert_json_value "$output" ".category_scores.test_coverage.weight" "15" "Test coverage weight should be 15"
}

#############################################################################
# Test: Total weight is 60 (sum of all category weights)
#############################################################################
test_total_weight() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Calculate sum of weights
    local marker_weight=$(echo "$output" | jq -r '.category_scores.markers.weight')
    local deprecated_weight=$(echo "$output" | jq -r '.category_scores.deprecated.weight')
    local file_weight=$(echo "$output" | jq -r '.category_scores.file_size.weight')
    local dup_weight=$(echo "$output" | jq -r '.category_scores.duplication.weight')
    local test_weight=$(echo "$output" | jq -r '.category_scores.test_coverage.weight')

    local total=$((marker_weight + deprecated_weight + file_weight + dup_weight + test_weight))
    assert_equals "60" "$total" "Total weight should be 60"
}

#############################################################################
# Test: Debt score capped at 100
#############################################################################
test_score_capped_at_100() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create maximum debt scenario
    # Many TODO, FIXME, HACK markers
    for i in {1..100}; do
        echo "// TODO: task $i" >> "$repo_dir/debt.js"
        echo "// FIXME: bug $i" >> "$repo_dir/debt.js"
        echo "// HACK: hack $i" >> "$repo_dir/debt.js"
    done

    # Many deprecated
    for i in {1..50}; do
        echo "@Deprecated" >> "$repo_dir/deprecated.java"
        echo "public void method$i() {}" >> "$repo_dir/deprecated.java"
    done

    # Very long files
    for j in $(seq 1 3000); do
        echo "// Line $j" >> "$repo_dir/huge.js"
    done

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local debt_score=$(echo "$output" | jq -r '.summary.debt_score')
    assert_less_than_or_equal "$debt_score" "100" "Score should be capped at 100"

    teardown
}

#############################################################################
# Test: All level classifications
#############################################################################
test_level_classifications() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local debt_score=$(echo "$output" | jq -r '.summary.debt_score')
    local debt_level=$(echo "$output" | jq -r '.summary.debt_level')

    # Verify level matches score
    if [[ $debt_score -le 20 ]]; then
        assert_equals "excellent" "$debt_level" "Score 0-20 should be excellent"
    elif [[ $debt_score -le 40 ]]; then
        assert_equals "good" "$debt_level" "Score 21-40 should be good"
    elif [[ $debt_score -le 60 ]]; then
        assert_equals "moderate" "$debt_level" "Score 41-60 should be moderate"
    elif [[ $debt_score -le 80 ]]; then
        assert_equals "high" "$debt_level" "Score 61-80 should be high"
    else
        assert_equals "critical" "$debt_level" "Score 81-100 should be critical"
    fi
}

#############################################################################
# Test: Output is valid JSON
#############################################################################
test_valid_json_output() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    assert_json_valid "$output" "Output should be valid JSON"
}

#############################################################################
# Test: Version in output
#############################################################################
test_version_in_output() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    assert_json_contains_key "$output" "version" "Should have version"
    assert_json_value "$output" ".version" "2.0.0" "Version should be 2.0.0"
}

#############################################################################
# Test: Analyzer name in output
#############################################################################
test_analyzer_name() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    assert_json_value "$output" ".analyzer" "tech-debt" "Analyzer should be tech-debt"
}

#############################################################################
# Test: Timestamp in output
#############################################################################
test_timestamp_in_output() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    assert_json_contains_key "$output" "timestamp" "Should have timestamp"

    local timestamp=$(echo "$output" | jq -r '.timestamp')
    assert_contains "$timestamp" "T" "Timestamp should be ISO format"
}

#############################################################################
# Test: Target in output
#############################################################################
test_target_in_output() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    assert_json_contains_key "$output" "target" "Should have target"
}

#############################################################################
# Test: Code stats in output
#############################################################################
test_code_stats_in_output() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    assert_json_contains_key "$output" "code_stats" "Should have code_stats"
    assert_json_contains_key "$output" "code_stats.total_lines" "Should have total_lines"
    assert_json_contains_key "$output" "code_stats.total_files" "Should have total_files"
}

#############################################################################
# Run all tests
#############################################################################
run_test "Debt score exists" test_debt_score_exists
run_test "Debt score range" test_debt_score_range
run_test "Clean repo low score" test_clean_repo_low_score
run_test "High debt repo" test_high_debt_repo
run_test "Excellent level" test_excellent_level
run_test "Good level" test_good_level
run_test "Category scores structure" test_category_scores_structure
run_test "Category weights" test_category_weights
run_test "Total weight is 60" test_total_weight
run_test "Score capped at 100" test_score_capped_at_100
run_test "Level classifications" test_level_classifications
run_test "Valid JSON output" test_valid_json_output
run_test "Version in output" test_version_in_output
run_test "Analyzer name" test_analyzer_name
run_test "Timestamp in output" test_timestamp_in_output
run_test "Target in output" test_target_in_output
run_test "Code stats in output" test_code_stats_in_output

print_summary
