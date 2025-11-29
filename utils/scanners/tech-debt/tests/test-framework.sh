#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Test Framework for Tech Debt Scanner
# Provides assertion functions, test execution, and reporting
#############################################################################

# Colors for test output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
CURRENT_TEST=""
FAILED_TESTS=()

# Setup function called before each test
setup() {
    TEST_TEMP_DIR=$(mktemp -d)
    export TEST_TEMP_DIR
}

# Teardown function called after each test
teardown() {
    if [[ -n "$TEST_TEMP_DIR" ]] && [[ -d "$TEST_TEMP_DIR" ]]; then
        rm -rf "$TEST_TEMP_DIR"
    fi
}

# Assertion functions
assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="${3:-Expected '$expected' but got '$actual'}"

    if [[ "$expected" == "$actual" ]]; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Expected: $expected"
        echo "  Actual:   $actual"
        return 1
    fi
}

assert_not_equals() {
    local not_expected="$1"
    local actual="$2"
    local message="${3:-Expected value different from '$not_expected'}"

    if [[ "$not_expected" != "$actual" ]]; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Should not equal: $not_expected"
        echo "  Actual:          $actual"
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-Expected to find '$needle' in string}"

    if echo "$haystack" | grep -q "$needle"; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Looking for: $needle"
        echo "  In string:   $haystack"
        return 1
    fi
}

assert_not_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-Expected NOT to find '$needle' in string}"

    if ! echo "$haystack" | grep -q "$needle"; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Should not contain: $needle"
        echo "  But found in:       $haystack"
        return 1
    fi
}

assert_greater_than() {
    local value="$1"
    local threshold="$2"
    local message="${3:-Expected $value > $threshold}"

    if awk -v val="$value" -v thresh="$threshold" 'BEGIN { exit !(val > thresh) }'; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Value:     $value"
        echo "  Threshold: $threshold"
        return 1
    fi
}

assert_less_than() {
    local value="$1"
    local threshold="$2"
    local message="${3:-Expected $value < $threshold}"

    if awk -v val="$value" -v thresh="$threshold" 'BEGIN { exit !(val < thresh) }'; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Value:     $value"
        echo "  Threshold: $threshold"
        return 1
    fi
}

assert_greater_than_or_equal() {
    local value="$1"
    local threshold="$2"
    local message="${3:-Expected $value >= $threshold}"

    if awk -v val="$value" -v thresh="$threshold" 'BEGIN { exit !(val >= thresh) }'; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Value:     $value"
        echo "  Threshold: $threshold"
        return 1
    fi
}

assert_less_than_or_equal() {
    local value="$1"
    local threshold="$2"
    local message="${3:-Expected $value <= $threshold}"

    if awk -v val="$value" -v thresh="$threshold" 'BEGIN { exit !(val <= thresh) }'; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Value:     $value"
        echo "  Threshold: $threshold"
        return 1
    fi
}

assert_file_exists() {
    local file="$1"
    local message="${2:-Expected file to exist: $file}"

    if [[ -f "$file" ]]; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        return 1
    fi
}

assert_file_not_exists() {
    local file="$1"
    local message="${2:-Expected file NOT to exist: $file}"

    if [[ ! -f "$file" ]]; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        return 1
    fi
}

assert_json_valid() {
    local json="$1"
    local message="${2:-Expected valid JSON}"

    if echo "$json" | jq empty 2>/dev/null; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  JSON: $json"
        return 1
    fi
}

assert_json_contains_key() {
    local json="$1"
    local key="$2"
    local message="${3:-Expected JSON to contain key: $key}"

    if echo "$json" | jq -e ".$key" >/dev/null 2>&1; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        return 1
    fi
}

assert_json_value() {
    local json="$1"
    local jq_path="$2"
    local expected="$3"
    local message="${4:-Expected JSON value at $jq_path to be $expected}"

    local actual=$(echo "$json" | jq -r "$jq_path" 2>/dev/null)

    if [[ "$actual" == "$expected" ]]; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Path:     $jq_path"
        echo "  Expected: $expected"
        echo "  Actual:   $actual"
        return 1
    fi
}

assert_json_value_in_range() {
    local json="$1"
    local jq_path="$2"
    local min="$3"
    local max="$4"
    local message="${5:-Expected JSON value at $jq_path to be between $min and $max}"

    local actual=$(echo "$json" | jq -r "$jq_path" 2>/dev/null)

    if awk -v val="$actual" -v min="$min" -v max="$max" 'BEGIN { exit !(val >= min && val <= max) }'; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Path:     $jq_path"
        echo "  Expected: $min <= value <= $max"
        echo "  Actual:   $actual"
        return 1
    fi
}

assert_exit_code() {
    local expected_code="$1"
    local command="$2"
    local message="${3:-Expected exit code $expected_code}"

    eval "$command" >/dev/null 2>&1
    local actual_code=$?

    if [[ $actual_code -eq $expected_code ]]; then
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        echo "  Expected: $expected_code"
        echo "  Actual:   $actual_code"
        return 1
    fi
}

# Test execution functions
run_test() {
    local test_name="$1"
    local test_function="$2"

    CURRENT_TEST="$test_name"
    TESTS_RUN=$((TESTS_RUN + 1))

    # Run setup
    setup

    # Run the test and capture result
    if $test_function; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_name")
        echo -e "${RED}✗ FAIL${NC}: $test_name"
    fi

    # Run teardown
    teardown

    echo ""
}

# Print test summary
print_summary() {
    echo ""
    echo "========================================="
    echo "  Test Results"
    echo "========================================="
    echo ""
    echo "Total Tests:  $TESTS_RUN"
    echo -e "${GREEN}Passed:       $TESTS_PASSED${NC}"

    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "${RED}Failed:       $TESTS_FAILED${NC}"
        echo ""
        echo "Failed Tests:"
        for test in "${FAILED_TESTS[@]}"; do
            echo -e "${RED}  ✗ $test${NC}"
        done
    else
        echo -e "${GREEN}Failed:       0${NC}"
    fi

    echo ""

    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo -e "${GREEN}=========================================${NC}"
        echo -e "${GREEN}  ALL TESTS PASSED!${NC}"
        echo -e "${GREEN}=========================================${NC}"
        return 0
    else
        echo -e "${RED}=========================================${NC}"
        echo -e "${RED}  SOME TESTS FAILED${NC}"
        echo -e "${RED}=========================================${NC}"
        return 1
    fi
}

# Helper function to create test repository with debt markers
create_test_repo_with_markers() {
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create files with various debt markers
    cat > "$repo_dir/todo-file.js" << 'EOF'
// Simple file with TODO markers
function doSomething() {
    // TODO: implement proper error handling
    // TODO: add validation
    console.log("doing something");
}

// TODO: refactor this function
function another() {
    return true;
}
EOF

    cat > "$repo_dir/fixme-file.py" << 'EOF'
# File with FIXME markers
def process_data(data):
    # FIXME: race condition here
    # FIXME: null check missing
    return data

# FIXME: add proper exception handling
def risky_operation():
    pass
EOF

    cat > "$repo_dir/hack-file.ts" << 'EOF'
// File with HACK markers
function workaround() {
    // HACK: bypassing validation temporarily
    // HACK: temporary fix for production issue
    return true;
}

// XXX: security concern here
function unsafeOperation() {
    // XXX: not thread safe
    return null;
}
EOF

    cat > "$repo_dir/mixed-file.java" << 'EOF'
public class MixedMarkers {
    // TODO: clean up this class
    // FIXME: memory leak in this method
    // HACK: workaround for library bug
    public void doWork() {
        // BUG: off-by-one error
        // KLUDGE: clumsy solution
        // OPTIMIZE: O(n^2) complexity
        // REFACTOR: extract to separate class
    }
}
EOF

    echo "$repo_dir"
}

# Helper function to create test repository with deprecated code
create_test_repo_with_deprecated() {
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    cat > "$repo_dir/deprecated.py" << 'EOF'
@deprecated
def old_function():
    """This function is deprecated"""
    pass

# DEPRECATED: will be removed in v3
def legacy_function():
    pass
EOF

    cat > "$repo_dir/deprecated.java" << 'EOF'
public class DeprecatedClass {
    @Deprecated
    public void oldMethod() {
        // Legacy code
    }

    @deprecated
    public void anotherOldMethod() {
        // More legacy code
    }
}
EOF

    echo "$repo_dir"
}

# Helper function to create test repository with long files
create_test_repo_with_long_files() {
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    local lines="${1:-600}"  # Default 600 lines
    mkdir -p "$repo_dir"

    # Generate a file with many lines
    cat > "$repo_dir/long-file.js" << 'EOF'
// Start of long file
function startFunction() {
    console.log("start");
}
EOF

    # Add lines to make it long
    for i in $(seq 1 $lines); do
        echo "// Line $i of padding" >> "$repo_dir/long-file.js"
    done

    cat >> "$repo_dir/long-file.js" << 'EOF'
// End of long file
function endFunction() {
    console.log("end");
}
EOF

    echo "$repo_dir"
}

# Helper function to create test repository with test files
create_test_repo_with_tests() {
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir/src" "$repo_dir/tests"

    # Source files
    echo 'function add(a, b) { return a + b; }' > "$repo_dir/src/math.js"
    echo 'function greet(name) { return "Hello " + name; }' > "$repo_dir/src/greet.js"
    echo 'function process(data) { return data; }' > "$repo_dir/src/process.js"

    # Test files
    echo 'test("add works", () => { expect(add(1,2)).toBe(3); });' > "$repo_dir/tests/math.test.js"
    echo 'describe("greet", () => { it("works", () => {}); });' > "$repo_dir/tests/greet.spec.js"

    echo "$repo_dir"
}

# Helper to create minimal repo for quick tests
create_minimal_test_repo() {
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    echo 'console.log("hello");' > "$repo_dir/index.js"
    echo 'print("hello")' > "$repo_dir/main.py"

    echo "$repo_dir"
}

# Export all functions
export -f setup teardown
export -f assert_equals assert_not_equals assert_contains assert_not_contains
export -f assert_greater_than assert_less_than assert_greater_than_or_equal assert_less_than_or_equal
export -f assert_file_exists assert_file_not_exists
export -f assert_json_valid assert_json_contains_key assert_json_value assert_json_value_in_range
export -f assert_exit_code
export -f run_test print_summary
export -f create_test_repo_with_markers create_test_repo_with_deprecated
export -f create_test_repo_with_long_files create_test_repo_with_tests create_minimal_test_repo
