#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Test Framework for Technology Identification
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

# Helper function to create test JSON
create_test_json() {
    local key_values=("$@")
    local json="{"
    local first=true

    for kv in "${key_values[@]}"; do
        if [[ "$first" == true ]]; then
            first=false
        else
            json+=","
        fi

        local key="${kv%%=*}"
        local value="${kv#*=}"

        # Check if value is a number
        if [[ "$value" =~ ^[0-9]+$ ]]; then
            json+="\"$key\":$value"
        else
            json+="\"$key\":\"$value\""
        fi
    done

    json+="}"
    echo "$json"
}

# Helper function to create test SBOM
create_test_sbom() {
    local components=("$@")
    local sbom='{
        "bomFormat": "CycloneDX",
        "specVersion": "1.4",
        "version": 1,
        "components": ['

    local first=true
    for component in "${components[@]}"; do
        if [[ "$first" == true ]]; then
            first=false
        else
            sbom+=","
        fi

        # Parse component string: name:version:purl
        IFS=':' read -r name version purl <<< "$component"

        sbom+='{
            "type": "library",
            "name": "'$name'",
            "version": "'$version'",
            "purl": "'$purl'"
        }'
    done

    sbom+=']}'
    echo "$sbom"
}

# Helper to create test repository structure
create_test_repo() {
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create package.json if specified
    if [[ "$1" == "--with-package-json" ]]; then
        shift
        local packages=("$@")
        local package_json='{"dependencies":{'
        local first=true

        for pkg in "${packages[@]}"; do
            if [[ "$first" == true ]]; then
                first=false
            else
                package_json+=","
            fi

            IFS=':' read -r name version <<< "$pkg"
            package_json+="\"$name\":\"$version\""
        done

        package_json+='}}'
        echo "$package_json" > "$repo_dir/package.json"
    fi

    echo "$repo_dir"
}

# Utility function to mock command output
mock_command() {
    local command="$1"
    local output="$2"
    local exit_code="${3:-0}"

    # Create mock script
    local mock_script="$TEST_TEMP_DIR/mock-$command"
    cat > "$mock_script" << EOF
#!/bin/bash
echo '$output'
exit $exit_code
EOF
    chmod +x "$mock_script"

    # Add to PATH
    export PATH="$TEST_TEMP_DIR:$PATH"
}

# Export all functions
export -f setup teardown
export -f assert_equals assert_not_equals assert_contains assert_not_contains
export -f assert_greater_than assert_less_than
export -f assert_file_exists assert_file_not_exists
export -f assert_json_valid assert_json_contains_key assert_json_value
export -f assert_exit_code
export -f run_test print_summary
export -f create_test_json create_test_sbom create_test_repo mock_command
