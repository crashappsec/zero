#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for Deprecated Code Detection
# Tests detection of @deprecated and DEPRECATED markers
#
# Based on RAG indicators: rag/tech-debt/indicators/code-markers.json
# - @deprecated / @Deprecated / DEPRECATED: weight 1.5 (medium severity)
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
echo "  Deprecated Code Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

#############################################################################
# Test: @deprecated annotation detection (Python style)
#############################################################################
test_deprecated_python_style() {
    local repo_dir=$(create_test_repo_with_deprecated)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local deprecated_count=$(echo "$output" | jq -r '.summary.deprecated_count')
    assert_greater_than "$deprecated_count" "0" "Should detect @deprecated annotations"
}

#############################################################################
# Test: @Deprecated annotation detection (Java style)
#############################################################################
test_deprecated_java_style() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    cat > "$repo_dir/Service.java" << 'EOF'
public class Service {
    @Deprecated
    public void oldMethod() {
        // This method is deprecated
    }

    @Deprecated
    public void anotherOldMethod() {
        // Also deprecated
    }
}
EOF

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local deprecated_count=$(echo "$output" | jq -r '.summary.deprecated_count')
    assert_equals "2" "$deprecated_count" "Should detect @Deprecated Java annotations"

    teardown
}

#############################################################################
# Test: DEPRECATED comment detection
#############################################################################
test_deprecated_comment_style() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    cat > "$repo_dir/legacy.py" << 'EOF'
# DEPRECATED: will be removed in v3
def old_function():
    pass

# DEPRECATED: use new_api instead
def legacy_api():
    pass
EOF

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local deprecated_count=$(echo "$output" | jq -r '.summary.deprecated_count')
    assert_equals "2" "$deprecated_count" "Should detect DEPRECATED comments"

    teardown
}

#############################################################################
# Test: Clean repo has zero deprecated
#############################################################################
test_clean_repo_no_deprecated() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local deprecated_count=$(echo "$output" | jq -r '.summary.deprecated_count')
    assert_equals "0" "$deprecated_count" "Clean repo should have 0 deprecated"
}

#############################################################################
# Test: Deprecated details in output
#############################################################################
test_deprecated_details() {
    local repo_dir=$(create_test_repo_with_deprecated)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local deprecated_array_length=$(echo "$output" | jq '.deprecated | length')
    assert_greater_than "$deprecated_array_length" "0" "Should have deprecated items"

    local first_deprecated=$(echo "$output" | jq '.deprecated[0]')
    assert_json_contains_key "$first_deprecated" "file" "Deprecated should have file"
    assert_json_contains_key "$first_deprecated" "line" "Deprecated should have line"
}

#############################################################################
# Test: Deprecated category score calculation
#############################################################################
test_deprecated_category_score() {
    local repo_dir=$(create_test_repo_with_deprecated)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local deprecated_score=$(echo "$output" | jq -r '.category_scores.deprecated.score')

    # Score = deprecated_count * 1.5 (using integer math: *15/10)
    assert_greater_than "$deprecated_score" "0" "Deprecated score should be > 0"
    assert_less_than_or_equal "$deprecated_score" "100" "Deprecated score should be <= 100"
}

#############################################################################
# Test: Deprecated weight is 5 (from RAG guide)
#############################################################################
test_deprecated_weight() {
    local repo_dir=$(create_test_repo_with_deprecated)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local weight=$(echo "$output" | jq -r '.category_scores.deprecated.weight')
    assert_equals "5" "$weight" "Deprecated category weight should be 5"
}

#############################################################################
# Test: Case variations detected
#############################################################################
test_case_variations() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    cat > "$repo_dir/mixed-case.java" << 'EOF'
public class MixedCase {
    @Deprecated
    public void upperCase() {}

    @deprecated
    public void lowerCase() {}
}
EOF

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local deprecated_count=$(echo "$output" | jq -r '.summary.deprecated_count')
    assert_equals "2" "$deprecated_count" "Should detect both @Deprecated and @deprecated"

    teardown
}

#############################################################################
# Test: Deprecated in various languages
#############################################################################
test_deprecated_various_languages() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Python
    echo '@deprecated' > "$repo_dir/file.py"
    echo 'def old(): pass' >> "$repo_dir/file.py"

    # JavaScript
    echo '/** @deprecated use newFunc */' > "$repo_dir/file.js"
    echo 'function old() {}' >> "$repo_dir/file.js"

    # Java
    echo '@Deprecated' > "$repo_dir/file.java"
    echo 'public void old() {}' >> "$repo_dir/file.java"

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local deprecated_count=$(echo "$output" | jq -r '.summary.deprecated_count')
    assert_equals "3" "$deprecated_count" "Should detect deprecated in all languages"

    teardown
}

#############################################################################
# Test: Deprecated score capped at 100
#############################################################################
test_deprecated_score_capped() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create many deprecated items to exceed cap
    # Score = count * 1.5, so 100 items = 150, should cap at 100
    for i in {1..100}; do
        echo "@Deprecated" >> "$repo_dir/many-deprecated.java"
        echo "public void method$i() {}" >> "$repo_dir/many-deprecated.java"
    done

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local deprecated_score=$(echo "$output" | jq -r '.category_scores.deprecated.score')
    assert_less_than_or_equal "$deprecated_score" "100" "Deprecated score should be capped at 100"

    teardown
}

#############################################################################
# Test: Deprecated level classification
#############################################################################
test_deprecated_level_classification() {
    local repo_dir=$(create_minimal_test_repo)

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local level=$(echo "$output" | jq -r '.category_scores.deprecated.level')
    assert_equals "excellent" "$level" "Zero deprecated should result in excellent level"
}

#############################################################################
# Test: Context captured for deprecated items
#############################################################################
test_deprecated_context_captured() {
    setup
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    cat > "$repo_dir/legacy.py" << 'EOF'
@deprecated
def oldFunction():
    """Old function, use newFunction instead"""
    pass
EOF

    local output=$("$SCANNER" --local-path "$repo_dir" 2>/dev/null)

    local context=$(echo "$output" | jq -r '.deprecated[0].context')
    assert_contains "$context" "deprecated" "Context should contain deprecated marker"

    teardown
}

#############################################################################
# Run all tests
#############################################################################
run_test "Deprecated Python style" test_deprecated_python_style
run_test "Deprecated Java style" test_deprecated_java_style
run_test "Deprecated comment style" test_deprecated_comment_style
run_test "Clean repo no deprecated" test_clean_repo_no_deprecated
run_test "Deprecated details in output" test_deprecated_details
run_test "Deprecated category score" test_deprecated_category_score
run_test "Deprecated weight is 5" test_deprecated_weight
run_test "Case variations detected" test_case_variations
run_test "Deprecated various languages" test_deprecated_various_languages
run_test "Deprecated score capped at 100" test_deprecated_score_capped
run_test "Deprecated level classification" test_deprecated_level_classification
run_test "Deprecated context captured" test_deprecated_context_captured

print_summary
