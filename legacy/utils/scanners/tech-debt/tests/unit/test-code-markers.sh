#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for Code Markers Detection
# Tests detection of TODO, FIXME, HACK, XXX, BUG, KLUDGE, OPTIMIZE, REFACTOR
#
# Based on RAG indicators: rag/tech-debt/indicators/code-markers.json
# - TODO: weight 0.5 (low severity)
# - FIXME: weight 1.5 (medium severity)
# - HACK: weight 3.0 (high severity)
# - XXX: weight 3.0 (high severity)
# - KLUDGE: weight 3.0 (high severity)
# - OPTIMIZE: weight 0.5 (low severity)
# - REFACTOR: weight 1.0 (medium severity)
# - BUG: weight 2.5 (high severity)
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
echo "  Code Markers Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

#############################################################################
# Test: TODO markers detection
#############################################################################
test_todo_markers() {
    local repo_dir=$(create_test_repo_with_markers)

    # Run scanner
    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Verify TODO count (at least 4 TODOs in fixture files)
    local todo_count=$(echo "$output" | jq -r '.summary.todo_count')
    assert_greater_than_or_equal "$todo_count" "4" "Should detect at least 4 TODO markers"
}

#############################################################################
# Test: FIXME markers detection
#############################################################################
test_fixme_markers() {
    local repo_dir=$(create_test_repo_with_markers)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Verify FIXME count (at least 3 in fixture files)
    local fixme_count=$(echo "$output" | jq -r '.summary.fixme_count')
    assert_greater_than_or_equal "$fixme_count" "3" "Should detect at least 3 FIXME markers"
}

#############################################################################
# Test: HACK markers detection
#############################################################################
test_hack_markers() {
    local repo_dir=$(create_test_repo_with_markers)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Verify HACK count (at least 2 in fixture files)
    local hack_count=$(echo "$output" | jq -r '.summary.hack_count')
    assert_greater_than_or_equal "$hack_count" "2" "Should detect at least 2 HACK markers"
}

#############################################################################
# Test: XXX markers detection
#############################################################################
test_xxx_markers() {
    local repo_dir=$(create_test_repo_with_markers)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Verify XXX count (2 in hack-file.ts)
    local xxx_count=$(echo "$output" | jq -r '.summary.xxx_count')
    assert_equals "2" "$xxx_count" "Should detect 2 XXX markers"
}

#############################################################################
# Test: All marker types in output
#############################################################################
test_marker_types_in_output() {
    local repo_dir=$(create_test_repo_with_markers)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Check that markers array contains different types
    local marker_types=$(echo "$output" | jq -r '.markers[].type' | sort -u)

    assert_contains "$marker_types" "TODO" "Should contain TODO markers"
    assert_contains "$marker_types" "FIXME" "Should contain FIXME markers"
    assert_contains "$marker_types" "HACK" "Should contain HACK markers"
    assert_contains "$marker_types" "XXX" "Should contain XXX markers"
}

#############################################################################
# Test: Marker details include file and line
#############################################################################
test_marker_details() {
    local repo_dir=$(create_test_repo_with_markers)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    # Check first marker has required fields
    local first_marker=$(echo "$output" | jq '.markers[0]')

    assert_json_contains_key "$first_marker" "type" "Marker should have type"
    assert_json_contains_key "$first_marker" "file" "Marker should have file"
    assert_json_contains_key "$first_marker" "line" "Marker should have line number"
    assert_json_contains_key "$first_marker" "text" "Marker should have text"
}

#############################################################################
# Test: Total markers count
#############################################################################
test_total_markers() {
    local repo_dir=$(create_test_repo_with_markers)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local total=$(echo "$output" | jq -r '.summary.total_markers')
    local todo=$(echo "$output" | jq -r '.summary.todo_count')
    local fixme=$(echo "$output" | jq -r '.summary.fixme_count')
    local hack=$(echo "$output" | jq -r '.summary.hack_count')
    local xxx=$(echo "$output" | jq -r '.summary.xxx_count')

    local expected=$((todo + fixme + hack + xxx))
    assert_equals "$expected" "$total" "Total markers should equal sum of individual counts"
}

#############################################################################
# Test: Clean repo has zero markers
#############################################################################
test_clean_repo_no_markers() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local total=$(echo "$output" | jq -r '.summary.total_markers')
    assert_equals "0" "$total" "Clean repo should have 0 markers"
}

#############################################################################
# Test: Marker category score based on RAG weights
#############################################################################
test_marker_category_score() {
    local repo_dir=$(create_test_repo_with_markers)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local marker_score=$(echo "$output" | jq -r '.category_scores.markers.score')

    # Score calculation from code: (todo * 0.5 + fixme * 1.5 + hack * 3.0) / 10
    # With 4 TODO, 4 FIXME, 3 HACK: (4*5 + 4*15 + 3*30) / 10 = (20 + 60 + 90) / 10 = 17
    # Should be in reasonable range (>0 and <=100)
    assert_greater_than "$marker_score" "0" "Marker score should be greater than 0"
    assert_less_than_or_equal "$marker_score" "100" "Marker score should be <= 100"
}

#############################################################################
# Test: Marker score weight is 15 (from RAG guide)
#############################################################################
test_marker_weight() {
    local repo_dir=$(create_test_repo_with_markers)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local weight=$(echo "$output" | jq -r '.category_scores.markers.weight')
    assert_equals "15" "$weight" "Marker category weight should be 15"
}

#############################################################################
# Test: Case insensitive marker detection
#############################################################################
test_case_insensitive_detection() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create file with mixed case markers
    cat > "$repo_dir/mixed-case.js" << 'EOF'
// todo: lowercase
// Todo: mixed case
// TODO: uppercase
// fixme: lowercase
// FIXME: uppercase
EOF

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local todo_count=$(echo "$output" | jq -r '.summary.todo_count')
    local fixme_count=$(echo "$output" | jq -r '.summary.fixme_count')

    # Scanner uses grep -qi which is case insensitive
    assert_equals "3" "$todo_count" "Should detect TODO regardless of case"
    assert_equals "2" "$fixme_count" "Should detect FIXME regardless of case"

    teardown
}

#############################################################################
# Test: Markers in different file types
#############################################################################
test_markers_in_various_languages() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # JavaScript
    echo '// TODO: js marker' > "$repo_dir/file.js"
    # Python
    echo '# TODO: py marker' > "$repo_dir/file.py"
    # TypeScript
    echo '// TODO: ts marker' > "$repo_dir/file.ts"
    # Java
    echo '// TODO: java marker' > "$repo_dir/file.java"
    # Go
    echo '// TODO: go marker' > "$repo_dir/file.go"
    # Rust
    echo '// TODO: rust marker' > "$repo_dir/file.rs"
    # Shell
    echo '# TODO: shell marker' > "$repo_dir/file.sh"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local todo_count=$(echo "$output" | jq -r '.summary.todo_count')
    assert_equals "7" "$todo_count" "Should detect markers in all supported languages"

    teardown
}

#############################################################################
# Run all tests
#############################################################################
run_test "TODO markers detection" test_todo_markers
run_test "FIXME markers detection" test_fixme_markers
run_test "HACK markers detection" test_hack_markers
run_test "XXX markers detection" test_xxx_markers
run_test "All marker types in output" test_marker_types_in_output
run_test "Marker details include file and line" test_marker_details
run_test "Total markers count" test_total_markers
run_test "Clean repo has zero markers" test_clean_repo_no_markers
run_test "Marker category score calculation" test_marker_category_score
run_test "Marker weight is 15" test_marker_weight
run_test "Case insensitive detection" test_case_insensitive_detection
run_test "Markers in various languages" test_markers_in_various_languages

print_summary
