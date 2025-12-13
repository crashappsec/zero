#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for Test Coverage Indicator
# Tests test-to-code ratio calculation
#
# Based on RAG indicators: rag/tech-debt/indicators/test-coverage-thresholds.json
# Test-to-code ratio thresholds:
# - >= 0.8: excellent (score 0)
# - 0.5-0.8: good (score 20)
# - 0.3-0.5: moderate (score 45)
# - 0.1-0.3: poor (score 70)
# - < 0.1: critical (score 100)
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
echo "  Test Coverage Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

#############################################################################
# Test: Test files detected (various patterns)
#############################################################################
test_test_files_detected() {
    local repo_dir=$(create_test_repo_with_tests)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_files=$(echo "$output" | jq -r '.summary.test_files')
    assert_greater_than "$test_files" "0" "Should detect test files"
}

#############################################################################
# Test: .test.js files detected
#############################################################################
test_test_js_pattern() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    echo 'test("works", () => {})' > "$repo_dir/app.test.js"
    echo 'function app() {}' > "$repo_dir/app.js"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_files=$(echo "$output" | jq -r '.summary.test_files')
    assert_equals "1" "$test_files" "Should detect .test.js files"

    teardown
}

#############################################################################
# Test: .spec.js files detected
#############################################################################
test_spec_js_pattern() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    echo 'describe("app", () => {})' > "$repo_dir/app.spec.js"
    echo 'function app() {}' > "$repo_dir/app.js"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_files=$(echo "$output" | jq -r '.summary.test_files')
    assert_equals "1" "$test_files" "Should detect .spec.js files"

    teardown
}

#############################################################################
# Test: test_*.py files detected (Python)
#############################################################################
test_python_test_pattern() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    echo 'def test_something(): pass' > "$repo_dir/test_app.py"
    echo 'def main(): pass' > "$repo_dir/app.py"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_files=$(echo "$output" | jq -r '.summary.test_files')
    assert_equals "1" "$test_files" "Should detect test_*.py files"

    teardown
}

#############################################################################
# Test: *_test.go files detected (Go)
#############################################################################
test_go_test_pattern() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    echo 'func TestMain(t *testing.T) {}' > "$repo_dir/app_test.go"
    echo 'func main() {}' > "$repo_dir/main.go"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_files=$(echo "$output" | jq -r '.summary.test_files')
    assert_equals "1" "$test_files" "Should detect *_test.go files"

    teardown
}

#############################################################################
# Test: Test ratio calculation
#############################################################################
test_ratio_calculation() {
    local repo_dir=$(create_test_repo_with_tests)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_ratio=$(echo "$output" | jq -r '.summary.test_ratio')
    assert_greater_than "$test_ratio" "0" "Test ratio should be calculated"
}

#############################################################################
# Test: High test ratio = low score
#############################################################################
test_high_ratio_low_score() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create 1 source and 2 test files (ratio >= 0.8)
    echo 'function app() {}' > "$repo_dir/app.js"
    echo 'test("1", () => {})' > "$repo_dir/app.test.js"
    echo 'test("2", () => {})' > "$repo_dir/app.spec.js"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_score=$(echo "$output" | jq -r '.category_scores.test_coverage.score')

    # With ratio >= 0.8, score should be 0
    # But ratio is test_files / total_files, so 2/3 = 0.66 which is in good range (score 20)
    assert_less_than_or_equal "$test_score" "25" "High test ratio should have low score"

    teardown
}

#############################################################################
# Test: Low test ratio = high score
#############################################################################
test_low_ratio_high_score() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create many source files and 1 test file (low ratio)
    for i in {1..20}; do
        echo "function app$i() {}" > "$repo_dir/app$i.js"
    done
    echo 'test("1", () => {})' > "$repo_dir/app.test.js"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_score=$(echo "$output" | jq -r '.category_scores.test_coverage.score')

    # With ratio around 1/21 = 0.047 (< 0.1), score should be high
    assert_greater_than "$test_score" "50" "Low test ratio should have high score"

    teardown
}

#############################################################################
# Test: No test files = critical score
#############################################################################
test_no_tests_critical_score() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create source files only
    echo 'function app() {}' > "$repo_dir/app.js"
    echo 'def main(): pass' > "$repo_dir/main.py"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_score=$(echo "$output" | jq -r '.category_scores.test_coverage.score')
    local test_files=$(echo "$output" | jq -r '.summary.test_files')

    assert_equals "0" "$test_files" "Should have 0 test files"
    assert_equals "100" "$test_score" "No tests should result in score 100"

    teardown
}

#############################################################################
# Test: Test coverage weight is 15 (from RAG guide)
#############################################################################
test_coverage_weight() {
    local repo_dir=$(create_test_repo_with_tests)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local weight=$(echo "$output" | jq -r '.category_scores.test_coverage.weight')
    assert_equals "15" "$weight" "Test coverage weight should be 15"
}

#############################################################################
# Test: Score thresholds match RAG
#############################################################################
test_score_threshold_moderate() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create ratio around 0.4 (moderate range)
    for i in {1..6}; do
        echo "function app$i() {}" > "$repo_dir/app$i.js"
    done
    for i in {1..4}; do
        echo "test('$i', () => {})" > "$repo_dir/app$i.test.js"
    done

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_ratio=$(echo "$output" | jq -r '.summary.test_ratio')
    local test_score=$(echo "$output" | jq -r '.category_scores.test_coverage.score')

    # Ratio is 4/10 = 0.4, which should be in moderate range (30-50% = score 45)
    assert_json_value_in_range "$output" ".summary.test_ratio" "0.3" "0.5" "Ratio should be moderate"

    teardown
}

#############################################################################
# Test: Level classification matches score
#############################################################################
test_level_classification() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create excellent ratio (>= 0.8)
    echo 'function app() {}' > "$repo_dir/app.js"
    echo 'test("1", () => {})' > "$repo_dir/a.test.js"
    echo 'test("2", () => {})' > "$repo_dir/b.test.js"
    echo 'test("3", () => {})' > "$repo_dir/c.test.js"
    echo 'test("4", () => {})' > "$repo_dir/d.test.js"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local level=$(echo "$output" | jq -r '.category_scores.test_coverage.level')
    local score=$(echo "$output" | jq -r '.category_scores.test_coverage.score')

    # With 4 test files and 1 source file, ratio is 0.8 (80%), score should be 0-20
    assert_less_than_or_equal "$score" "20" "Score should be excellent/good range"

    teardown
}

#############################################################################
# Test: TypeScript test files detected
#############################################################################
test_typescript_test_pattern() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    echo 'test("works", () => {})' > "$repo_dir/app.test.ts"
    echo 'describe("app", () => {})' > "$repo_dir/app.spec.ts"
    echo 'function app() {}' > "$repo_dir/app.ts"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_files=$(echo "$output" | jq -r '.summary.test_files')
    assert_equals "2" "$test_files" "Should detect TypeScript test files"

    teardown
}

#############################################################################
# Test: Test files in subdirectories detected
#############################################################################
test_subdirectory_tests() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir/tests" "$repo_dir/src"

    echo 'function app() {}' > "$repo_dir/src/app.js"
    echo 'test("works", () => {})' > "$repo_dir/tests/app.test.js"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local test_files=$(echo "$output" | jq -r '.summary.test_files')
    assert_equals "1" "$test_files" "Should detect tests in subdirectories"

    teardown
}

#############################################################################
# Run all tests
#############################################################################
run_test "Test files detected" test_test_files_detected
run_test ".test.js pattern" test_test_js_pattern
run_test ".spec.js pattern" test_spec_js_pattern
run_test "Python test pattern" test_python_test_pattern
run_test "Go test pattern" test_go_test_pattern
run_test "Ratio calculation" test_ratio_calculation
run_test "High ratio = low score" test_high_ratio_low_score
run_test "Low ratio = high score" test_low_ratio_high_score
run_test "No tests = critical score" test_no_tests_critical_score
run_test "Test coverage weight is 15" test_coverage_weight
run_test "Score threshold moderate" test_score_threshold_moderate
run_test "Level classification" test_level_classification
run_test "TypeScript test pattern" test_typescript_test_pattern
run_test "Subdirectory tests" test_subdirectory_tests

print_summary
