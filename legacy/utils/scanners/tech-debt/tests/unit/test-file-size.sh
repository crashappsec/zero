#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for File Size Thresholds
# Tests detection of oversized files based on line counts
#
# Based on RAG indicators: rag/tech-debt/indicators/file-size-thresholds.json
# - 0-200 lines: excellent (score 0)
# - 201-500 lines: acceptable (score 10)
# - 501-1000 lines: warning (score 30)
# - 1001-2000 lines: high (score 60)
# - >2000 lines: critical (score 100) "God file"
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
echo "  File Size Thresholds Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

#############################################################################
# Test: Small files not flagged as long
#############################################################################
test_small_files_not_flagged() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local long_files_count=$(echo "$output" | jq -r '.summary.long_files_count')
    assert_equals "0" "$long_files_count" "Small files should not be flagged as long"
}

#############################################################################
# Test: Files over 500 lines flagged (default threshold)
#############################################################################
test_long_file_detection() {
    local repo_dir=$(create_test_repo_with_long_files 600)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local long_files_count=$(echo "$output" | jq -r '.summary.long_files_count')
    assert_equals "1" "$long_files_count" "Files over 500 lines should be flagged"
}

#############################################################################
# Test: Long file details in output
#############################################################################
test_long_file_details() {
    local repo_dir=$(create_test_repo_with_long_files 700)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local first_long_file=$(echo "$output" | jq '.long_files[0]')

    assert_json_contains_key "$first_long_file" "file" "Long file should have file path"
    assert_json_contains_key "$first_long_file" "lines" "Long file should have line count"
    assert_json_contains_key "$first_long_file" "threshold" "Long file should have threshold"
    assert_json_contains_key "$first_long_file" "excess" "Long file should have excess count"
}

#############################################################################
# Test: Excess calculation is correct
#############################################################################
test_excess_calculation() {
    local repo_dir=$(create_test_repo_with_long_files 600)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local lines=$(echo "$output" | jq -r '.long_files[0].lines')
    local threshold=$(echo "$output" | jq -r '.long_files[0].threshold')
    local excess=$(echo "$output" | jq -r '.long_files[0].excess')

    local expected_excess=$((lines - threshold))
    assert_equals "$expected_excess" "$excess" "Excess should equal lines minus threshold"
}

#############################################################################
# Test: File size category score calculation
#############################################################################
test_file_size_category_score() {
    # Create repo with multiple long files
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create 5 files over 500 lines (2 points each = 10 points)
    for i in {1..5}; do
        for j in $(seq 1 600); do
            echo "// Line $j" >> "$repo_dir/long-file-$i.js"
        done
    done

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local file_score=$(echo "$output" | jq -r '.category_scores.file_size.score')

    # Score = long_files_count * 2, so 5 * 2 = 10
    assert_equals "10" "$file_score" "File size score should be long_files * 2"

    teardown
}

#############################################################################
# Test: File size score capped at 100
#############################################################################
test_file_size_score_capped() {
    # Create repo with many long files to exceed cap
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create 60 files over 500 lines (2 points each = 120, but capped at 100)
    for i in {1..60}; do
        for j in $(seq 1 600); do
            echo "// Line $j" >> "$repo_dir/long-file-$i.js"
        done
    done

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local file_score=$(echo "$output" | jq -r '.category_scores.file_size.score')

    assert_less_than_or_equal "$file_score" "100" "File size score should be capped at 100"

    teardown
}

#############################################################################
# Test: File size weight is 10 (from RAG guide)
#############################################################################
test_file_size_weight() {
    local repo_dir=$(create_test_repo_with_long_files 600)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local weight=$(echo "$output" | jq -r '.category_scores.file_size.weight')
    assert_equals "10" "$weight" "File size category weight should be 10"
}

#############################################################################
# Test: Different file types are checked
#############################################################################
test_various_file_types_checked() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create long files in different languages
    for ext in js ts py java go rs; do
        for j in $(seq 1 600); do
            echo "// Line $j" >> "$repo_dir/long.$ext"
        done
    done

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local long_files_count=$(echo "$output" | jq -r '.summary.long_files_count')
    assert_equals "6" "$long_files_count" "Should detect long files in all supported languages"

    teardown
}

#############################################################################
# Test: Excluded directories not scanned
#############################################################################
test_excluded_directories() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir/node_modules" "$repo_dir/vendor" "$repo_dir/dist"

    # Create long files in excluded directories
    for j in $(seq 1 600); do
        echo "// Line $j" >> "$repo_dir/node_modules/long.js"
        echo "// Line $j" >> "$repo_dir/vendor/long.php"
        echo "// Line $j" >> "$repo_dir/dist/long.js"
    done

    # Create short file in main directory
    echo 'console.log("test")' > "$repo_dir/index.js"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local long_files_count=$(echo "$output" | jq -r '.summary.long_files_count')
    assert_equals "0" "$long_files_count" "Should not scan excluded directories"

    teardown
}

#############################################################################
# Test: Boundary at 500 lines (499 should not flag, 501 should)
#############################################################################
test_threshold_boundary() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create file with exactly 500 lines (should not be flagged)
    for j in $(seq 1 500); do
        echo "// Line $j" >> "$repo_dir/exactly-500.js"
    done

    # Create file with 501 lines (should be flagged)
    for j in $(seq 1 501); do
        echo "// Line $j" >> "$repo_dir/over-500.ts"
    done

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local long_files_count=$(echo "$output" | jq -r '.summary.long_files_count')
    assert_equals "1" "$long_files_count" "Only files over 500 lines should be flagged"

    # Verify the correct file was flagged
    local flagged_file=$(echo "$output" | jq -r '.long_files[0].file')
    assert_contains "$flagged_file" "over-500" "File over 500 should be flagged"

    teardown
}

#############################################################################
# Test: File size level classification
#############################################################################
test_file_size_level_classification() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local level=$(echo "$output" | jq -r '.category_scores.file_size.level')

    # With no long files, score is 0, level should be excellent
    assert_equals "excellent" "$level" "Zero long files should result in excellent level"
}

#############################################################################
# Test: Multiple long files all reported
#############################################################################
test_multiple_long_files_reported() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create 3 long files
    for i in {1..3}; do
        for j in $(seq 1 600); do
            echo "// Line $j" >> "$repo_dir/long-$i.js"
        done
    done

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local long_files_array_length=$(echo "$output" | jq '.long_files | length')
    assert_equals "3" "$long_files_array_length" "All long files should be in the array"

    teardown
}

#############################################################################
# Run all tests
#############################################################################
run_test "Small files not flagged" test_small_files_not_flagged
run_test "Long file detection" test_long_file_detection
run_test "Long file details in output" test_long_file_details
run_test "Excess calculation" test_excess_calculation
run_test "File size category score" test_file_size_category_score
run_test "File size score capped at 100" test_file_size_score_capped
run_test "File size weight is 10" test_file_size_weight
run_test "Various file types checked" test_various_file_types_checked
run_test "Excluded directories not scanned" test_excluded_directories
run_test "Threshold boundary (500 lines)" test_threshold_boundary
run_test "File size level classification" test_file_size_level_classification
run_test "Multiple long files reported" test_multiple_long_files_reported

print_summary
